package lib

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/mod/modfile"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

// Average number of launches between cleanups
const cleanupCycle = 100

// Number of days after which a binary is considered old
const cleanupThresholdDays = 90

// Number of days after which a “@latest” version binary must be rebuilt
const latestVersionRebuildThresholdDays = 30

func findGoMod(modDir string) (string, error) {
	var err error
	for {
		_, err = os.Stat(filepath.Join(modDir, "go.mod"))
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		err = nil
		parent := filepath.Dir(modDir)
		if parent == modDir {
			return "", errors.New("go.mod not found")
		}
		modDir = parent
	}
	return filepath.Join(modDir, "go.mod"), nil
}

type goEnv struct {
	Version string `json:"GOVERSION"`
}

func getGoCmd() (string, error) {
	p := filepath.Join(os.Getenv("GOROOT"), "bin", "go")
	if stat, err := os.Stat(p); err == nil && !stat.IsDir() {
		return p, nil
	}
	goPath, err := exec.LookPath("go")
	if err == nil {
		return goPath, nil
	}
	p = filepath.Join(runtime.GOROOT(), "bin", "go")
	if stat, err := os.Stat(p); err == nil && !stat.IsDir() {
		return p, nil
	}
	return "", errors.New("go command not found")
}

// splitArgs splits the arguments into the arguments for `go run` and the arguments for the command.
func splitArgs(goCmd string, args []string) (runArgs []string, cmdArgs []string, err error) {
	var buf bytes.Buffer
	cmd := exec.Command(goCmd, append([]string{"run", "-n"}, args...)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		_, _ = os.Stderr.Write(buf.Bytes())
		return
	}
	lines := strings.Split(buf.String(), "\n")
	lastLine := lines[len(lines)-2]
	fields := strings.SplitN(lastLine, " ", 2)
	runArgs = args
	if len(fields) > 1 {
		expandedCmdArgs := ""
		delim := ""
		for {
			elem := runArgs[len(runArgs)-1]
			runArgs = runArgs[:len(runArgs)-1]
			cmdArgs = append(cmdArgs, elem)
			expandedCmdArgs = elem + delim + expandedCmdArgs
			delim = " "
			if expandedCmdArgs == fields[1] {
				break
			}
		}
	}
	return
}

func getGlobalCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userCacheDir, "go-run-cache"), nil
}

func getModuleCacheDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	goMod, err := findGoMod(wd)
	if err != nil {
		return "", err
	}
	modDir := filepath.Dir(goMod)
	return filepath.Join(modDir, "cache", ".go-run-cache"), nil
}

type file struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"-"`
}

type buildInfo struct {
	GoVersion string   `json:"go_version"`
	BuildArgs []string `json:"build_args"`
	Package   *string  `json:"package,omitempty"`
	Files     []*file  `json:"files,omitempty"`
	Hash      string   `json:"hash"`
}

func getFileInfo(name string) (*file, error) {
	stat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, errors.New("not a file")
	}
	hashStr, err := (func() (string, error) {
		hash_ := sha1.New()
		reader, err := os.Open(name)
		if err != nil {
			return "", err
		}
		defer (func() { _ = reader.Close() })()
		_, err = io.Copy(hash_, reader)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(hash_.Sum(nil)), nil
	})()
	if err != nil {
		return nil, err
	}
	return &file{
		Name: name,
		Hash: hashStr,
		Size: stat.Size(),
	}, nil
}

func getFiles(args []string) ([]*file, []string, error) {
	if len(args) == 0 {
		return nil, args, nil
	}
	var files []*file
	buildArgsWoTgt := args
	name := buildArgsWoTgt[len(buildArgsWoTgt)-1]
	buildArgsWoTgt = buildArgsWoTgt[:len(buildArgsWoTgt)-1]
	if stat, err := os.Stat(name); err == nil && stat.IsDir() {
		_, err := findGoMod(filepath.Dir(name))
		if err == nil {
			return nil, args, nil
		}
		goFiles, err := os.ReadDir(name)
		if err != nil {
			return nil, args, nil
		}
		for _, goFile := range goFiles {
			if strings.HasSuffix(goFile.Name(), ".go") {
				p := filepath.Join(name, goFile.Name())
				fileInfo, err := getFileInfo(p)
				if err != nil {
					return nil, args, err
				}
				files = append(files, fileInfo)
			}
		}
	} else {
	outer:
		for {
			if !strings.HasSuffix(name, ".go") {
				break outer
			}
			fileInfo, err := getFileInfo(name)
			if err != nil {
				return nil, args, err
			}
			files = append(files, fileInfo)
			if len(buildArgsWoTgt) == 0 {
				break outer
			}
			name = buildArgsWoTgt[len(buildArgsWoTgt)-1]
			buildArgsWoTgt = buildArgsWoTgt[:len(buildArgsWoTgt)-1]
		}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Size == files[j].Size {
			return files[i].Hash < files[j].Hash
		}
		return files[i].Size < files[j].Size
	})
	return files, buildArgsWoTgt, nil
}

func getBuildInfo(goVersion string, buildArgsWoTgt []string, pkg *string, files []*file) buildInfo {
	hashAccu := sha1.New()
	hashAccu.Write([]byte(goVersion))
	for _, arg := range buildArgsWoTgt {
		hashAccu.Write([]byte(arg))
	}
	if pkg != nil {
		hashAccu.Write([]byte(*pkg))
	}
	for _, f := range files {
		hashAccu.Write([]byte(f.Hash))
	}
	return buildInfo{
		GoVersion: goVersion,
		BuildArgs: buildArgsWoTgt,
		Package:   pkg,
		Files:     files,
		Hash:      hex.EncodeToString(hashAccu.Sum(nil)),
	}
}

func putBuildInfo(exeCacheDir string, buildInfo_ *buildInfo) error {
	var err error
	infoPath := filepath.Join(exeCacheDir, "build_info.json")
	buildInfoJson, err := json.Marshal(buildInfo_)
	if err != nil {
		return err
	}
	err = os.WriteFile(infoPath, buildInfoJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func cleanupOldBinaries(dir string) error {
	subdirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}
		if stat, err := os.Stat(filepath.Join(dir, subdir.Name(), "main")); err == nil && !stat.IsDir() {
			if stat.ModTime().Before(stat.ModTime().Add(-cleanupThresholdDays * 24 * time.Hour)) {
				err = os.RemoveAll(filepath.Join(dir, subdir.Name()))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func Run(args []string) (err error) {
	goCmd, err := getGoCmd()
	if err != nil {
		return
	}
	goEnvOutput, err := exec.Command(goCmd, "env", "-json").Output()
	if err != nil {
		return
	}
	var goEnv goEnv
	err = json.Unmarshal(goEnvOutput, &goEnv)
	if err != nil {
		return
	}
	buildArgs, cmdArgs, err := splitArgs(goCmd, args)
	if err != nil {
		return
	}
	globalCacheDir, err := getGlobalCacheDir()
	if err != nil {
		return
	}

	// Clean up old binaries.
	if rand.Intn(cleanupCycle) == 0 {
		err = cleanupOldBinaries(globalCacheDir)
		if err != nil {
			return
		}
	}
	files, buildArgsWoTgt, err := getFiles(buildArgs)
	if err != nil {
		return
	}
	var pkg *string
	if len(files) == 0 {
		tmp := buildArgs[len(buildArgs)-1]
		if tmp[0] != '.' && tmp[0] != os.PathSeparator {
			pkg = &buildArgs[len(buildArgs)-1]
			buildArgsWoTgt = buildArgs[:len(buildArgs)-1]
		}
	}

	buildInfo_ := getBuildInfo(
		goEnv.Version,
		buildArgsWoTgt,
		pkg,
		files,
	)

	exeCacheDir := filepath.Join(globalCacheDir, buildInfo_.Hash)
	var mainPath string
	if len(files) > 0 {
		err = os.MkdirAll(exeCacheDir, 0755)
		if err != nil {
			return
		}
		mainPath = filepath.Join(exeCacheDir, "main")
		if stat, err := os.Stat(mainPath); err != nil || stat.IsDir() {
			cmd := exec.Command(goCmd, append([]string{"build", "-o", mainPath}, buildArgs...)...)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
			err = putBuildInfo(exeCacheDir, &buildInfo_)
			if err != nil {
				return err
			}
		}
	} else if pkg != nil && *pkg != "" {
		err = os.MkdirAll(exeCacheDir, 0755)
		if err != nil {
			return
		}
		pkgFields := strings.Split(*pkg, "@")
		if len(pkgFields) < 2 {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			goMod, err := findGoMod(wd)
			if err != nil {
				return errors.New("package name must be in the form of pkg@ver if not in module-aware mode")
			}
			reader, err := os.Open(goMod)
			if err != nil {
				return err
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			f, err := modfile.Parse(goMod, body, nil)
			if err != nil {
				return err
			}
			mainPkgName := *pkg
			modName := *pkg
		outer:
			for {
				for _, r := range f.Require {
					if r.Mod.Path == modName {
						pkgFields = []string{mainPkgName, r.Mod.Version}
						break outer
					}
				}
				parent := path.Dir(modName)
				if parent == modName {
					return errors.New("package not found in go.mod")
				}
				modName = parent
			}
			moduleCacheDir, err := getModuleCacheDir()
			if err != nil {
				return err
			}
			// Clean old binaries sometimes
			if rand.Intn(cleanupCycle) == 0 {
				err = cleanupOldBinaries(moduleCacheDir)
				if err != nil {
					return err
				}
			}
			exeCacheDir = filepath.Join(moduleCacheDir, buildInfo_.Hash)
		}
		pkgVer := pkgFields[1]
		err = os.MkdirAll(exeCacheDir, 0755)
		if err != nil {
			return
		}
		mainPath = filepath.Join(exeCacheDir, "main")
		stat, err := os.Stat(mainPath)
		needBuild := err != nil || stat.IsDir()
		if !needBuild &&
			pkgVer == "latest" &&
			stat.ModTime().Before(stat.ModTime().Add(-latestVersionRebuildThresholdDays*24*time.Hour)) {
			needBuild = true
		}
		if needBuild {
			cmd := exec.Command("go", append([]string{"install"}, buildArgs...)...)
			cmd.Env = append(os.Environ(), "GOBIN="+exeCacheDir)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
			pkgName := pkgFields[0]
			pkgBase := path.Base(pkgName)
			err = os.Rename(filepath.Join(exeCacheDir, pkgBase), mainPath)
			if err != nil {
				return err
			}
			err = putBuildInfo(exeCacheDir, &buildInfo_)
			if err != nil {
				return err
			}
		}
	} else {
		log.Println("Not caching.")
		tmpDir, err := os.MkdirTemp("", "go-run-cache")
		if err != nil {
			return err
		}
		defer (func() { _ = os.RemoveAll(tmpDir) })()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigCh
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				_ = os.RemoveAll(tmpDir)
			}
		}()
		mainPath = filepath.Join(tmpDir, "main")
		cmd := exec.Command(goCmd, append([]string{"build", "-o", mainPath}, buildArgs...)...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command(mainPath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		return err
	}
	cmd := exec.Command(mainPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}
