package minlib

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func v0(err error) {
	if err != nil {
		panic(err)
	}
}

func v[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func v2[T any, U any](t T, u U, err error) (T, U) {
	if err != nil {
		panic(err)
	}
	return t, u
}

type paramsT struct {
	initialDirPath string
	global         bool
}

type ConfDirPathOption func(*paramsT) error

// canonAbs returns the canonical absolute path of the given value.
func canonAbs(s string) (ret string, err error) {
	ret, err = filepath.Abs(s)
	if err != nil {
		return
	}
	ret, err = filepath.EvalSymlinks(ret)
	if err != nil {
		return
	}
	ret = filepath.Clean(ret)
	return
}

func WithInitialDir(initialDir string) ConfDirPathOption {
	return func(params *paramsT) (err error) {
		params.initialDirPath, err = canonAbs(initialDir)
		return
	}
}

func WithGlobal(f bool) ConfDirPathOption {
	return func(params *paramsT) error {
		params.global = f
		return nil
	}
}

func isRootDir(dir string) bool {
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	dirPath = filepath.Clean(dirPath)
	return dirPath == filepath.Dir(dirPath)
}

type PkgVerLockMapT map[string]string

const GobinCmdBase = "gobin"
const ManifestFileBase = "Gobinfile"
const ManifestLockFileBase = "Gobinfile-lock"
const goModFileBase = "go.mod"
const GobinDirBase = ".gobin"

// ConfDirPath returns the configuration directory path. If the global option is true, it returns the global (home)  configuration directory path. This returns the directory which contains the manifest file. If no manifest file is found in any parent directory, it returns the directory which contains the go.mod file.
func ConfDirPath(opts ...ConfDirPathOption) (
	confDirPath string,
	gobinPath string,
	err error,
) {
	params := &paramsT{}
	for _, opt := range opts {
		err = opt(params)
		if err != nil {
			return
		}
	}
	if params.global {
		confDirPath, err = os.UserHomeDir()
		if err != nil {
			return
		}
		gobinPath = os.Getenv("GOBIN")
		if gobinPath == "" {
			gobinPath = filepath.Join(confDirPath, "go", "bin")
		}
		return
	}
	confDirPath = params.initialDirPath
	if confDirPath == "" {
		confDirPath, err = canonAbs(".")
		if err != nil {
			return
		}
	}
	goModDirPath := ""
	for {
		// Record the directory which contains the go.mod file.
		if stat, errSub := os.Stat(filepath.Join(confDirPath, goModFileBase)); errSub == nil && !stat.IsDir() {
			if goModDirPath == "" {
				goModDirPath = confDirPath
			}
		}
		// When the manifest file is found, return the directory.
		if stat, err_ := os.Stat(filepath.Join(confDirPath, ManifestFileBase)); err_ == nil && !stat.IsDir() {
			break
		}
		if stat, err_ := os.Stat(filepath.Join(confDirPath, ManifestLockFileBase)); err_ == nil && stat.IsDir() {
			break
		}
		confDirPath = filepath.Dir(confDirPath)
		if isRootDir(confDirPath) {
			// If no manifest file is found in any parent directory, return the directory which contains the go.mod file.
			if goModDirPath != "" {
				confDirPath = goModDirPath
				break
			}
			confDirPath, gobinPath, err = "", "", errors.New("no go.mod or manifest file found")
			return
		}
	}
	gobinPath = filepath.Join(confDirPath, GobinDirBase)
	return
}

// GoListOutput represents the output of the `go list` command.
type GoListOutput struct {
	Version string `json:"Version"`
}

// PkgVerLockMap returns the package version lock map.
func PkgVerLockMap(dirPath string) (lockList PkgVerLockMapT, err error) {
	manifestLockPath := filepath.Join(dirPath, ManifestLockFileBase)
	if _, err_ := os.Stat(manifestLockPath); err_ != nil {
		return
	}
	reader := v(os.Open(manifestLockPath))
	scanner := bufio.NewScanner(reader)
	lockList = make(PkgVerLockMapT)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		divs := strings.SplitN(line, "@", 2)
		lockList[divs[0]] = divs[1]
	}
	return
}

// EnsureInstalled ensures that the program package is installed.
func EnsureInstalled(gobinPath string, pkgPath string, ver string, tags string, log *stdlog.Logger, vlog *stdlog.Logger) (cmdPkgVerPath string, err error) {
	pkgBase := path.Base(pkgPath)
	pkgBaseVer := pkgBase + "@" + ver
	cmdPath := filepath.Join(gobinPath, pkgBase)
	if tags != "" {
		hash := sha1.New()
		hash.Write([]byte(tags))
		sevenDigits := fmt.Sprintf("%x", hash.Sum(nil))[:7]
		pkgBaseVer += "-" + sevenDigits
	}
	cmdPkgVerPath = filepath.Join(gobinPath, pkgBaseVer)
	if _, err_ := os.Stat(cmdPkgVerPath); err_ != nil {
		log.Printf("Installing %s@%s\n", pkgPath, ver)
		cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", pkgPath, ver))
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinPath))
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		_ = os.Remove(cmdPath)
		err = cmd.Run()
		if err != nil {
			return
		}
		_ = os.Remove(cmdPkgVerPath)
		err = os.Rename(filepath.Join(gobinPath, pkgBase), cmdPkgVerPath)
		if err != nil {
			return
		}
		if pkgBase == GobinCmdBase {
			v0(os.Symlink(pkgBaseVer, cmdPath))
		} else {
			v0(os.Symlink(GobinCmdBase, cmdPath))
		}
	}
	return
}

func EnsureGobinCmdInstalled() (cmdPath string, err error) {
	confDirPath, gobinPath := v2(ConfDirPath())
	pkgVerLockMap := v(PkgVerLockMap(confDirPath))
	modPath := "github.com/knaka/gobin"
	pkgPath := "github.com/knaka/gobin/cmd/gobin"
	ver, ok := pkgVerLockMap[pkgPath]
	if !ok {
		cmd := exec.Command("go", "list", "-m",
			"--json", fmt.Sprintf("%s@%s", modPath, "latest"))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmd.Stderr = os.Stderr
		output := v(cmd.Output())
		goListOutput := GoListOutput{}
		v0(json.Unmarshal(output, &goListOutput))
		ver = goListOutput.Version
		manifestLockPath := filepath.Join(confDirPath, ManifestLockFileBase)
		writer := v(os.OpenFile(manifestLockPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600))
		defer (func() { v0(writer.Close()) })()
		_ = v(writer.WriteString(fmt.Sprintf("%s@%s\n", pkgPath, ver)))
	}
	return EnsureInstalled(gobinPath, pkgPath, ver, "", stdlog.Default(), stdlog.Default())
}

func RunCommand(name string, arg ...string) (execErr *exec.ExitError, err error) {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	errors.As(err, &execErr)
	return
}

//goland:noinspection GoUnusedFunction
func run() {
	gobinCmdPath := v(EnsureGobinCmdInstalled())
	errExec, err := RunCommand(gobinCmdPath, append([]string{"run"}, os.Args[1:]...)...)
	if err == nil {
		os.Exit(0)
	}
	if errExec != nil {
		os.Exit(errExec.ExitCode())
	}
	stdlog.Fatalf("Error b025d51: %+v", err)
}

//goland:noinspection GoUnusedFunction
func install() {
	gobinCmdPath := v(EnsureGobinCmdInstalled())
	errExec, err := RunCommand(gobinCmdPath, append([]string{"install"}, os.Args[1:]...)...)
	if err == nil {
		os.Exit(0)
	}
	if errExec != nil {
		os.Exit(errExec.ExitCode())
	}
	stdlog.Fatalf("Error 560d8bf: %+v", err)
}
