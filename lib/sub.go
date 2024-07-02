package lib

// Do not use 3rd party packages because this file is used to generate standalone go program files.
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const gobinListBase = "Gobinfile"
const gobinLockBase = "Gobinfile-lock.json"
const gobinLocalDirBase = ".gobin"

// GoEnv is a struct to hold the output of `go env --json`.
type GoEnv struct {
	Version string `json:"GOVERSION"`
	Gobin   string `json:"GOBIN"`
	Gopath  string `json:"GOPATH"`
}

var getGoCmdPath = sync.OnceValues(func() (goCmdPath string, err error) {
	goCmdPath = filepath.Join(runtime.GOROOT(), "bin", "go")
	if stat, errX := os.Stat(goCmdPath); errX != nil && !stat.IsDir() {
		return
	}
	goCmdPath, err = exec.LookPath("go")
	if err == nil {
		return goCmdPath, nil
	}
	return "", fmt.Errorf("go command not found")
})

var getGoEnv = sync.OnceValues(func() (goEnv GoEnv, err error) {
	goCmdPath, err := getGoCmdPath()
	if err != nil {
		return
	}
	outStr, err := exec.Command(goCmdPath, "env", "-json").Output()
	if err != nil {
		return
	}
	err = json.Unmarshal(outStr, &goEnv)
	if err != nil {
		return
	}
	return
})

var getGlobalGobinPath = sync.OnceValues(func() (ret string, err error) {
	goEnv, err := getGoEnv()
	if err != nil {
		return "", err

	}
	if goEnv.Gobin != "" {
		return goEnv.Gobin, nil
	}
	return filepath.Join(goEnv.Gopath, "bin"), nil
})

// ternaryF returns the result of the first function if cond is true, otherwise the result of the second function.
func ternaryF[T any](
	cond bool,
	t func() T,
	f func() T,
) (ret T) {
	if cond {
		if t != nil {
			ret = t()
		}
	} else {
		if f != nil {
			ret = f()
		}
	}
	return
}

type Gobin struct {
	Version   string   `json:"version"`
	BuildOpts []string `json:"build_opts"`
}

type GobinMap map[string]Gobin

type GobinList struct {
	Path string
	Map  GobinMap
}

func (l *GobinList) SaveJson() (err error) {
	b, err := json.Marshal(l.Map)
	if err != nil {
		return
	}
	err = os.WriteFile(l.Path, b, 0644)
	if err != nil {
		return
	}
	return
}

func (m *GobinMap) Find(name string) (string, *Gobin) {
	for pkgWoVer, gobin := range *m {
		if pkgWoVer == name || path.Base(pkgWoVer) == name {
			return pkgWoVer, &gobin
		}
	}
	return "", nil
}

func findGobinFile(workingDirPath string) (
	gobinListPath string,
	gobinLockPath string,
	gobinBinPath string,
	err error,
) {
	homeDirPath := ""
	homeDirPath, err = os.UserHomeDir()
	if err != nil {
		return
	}
	gobinDirPath := ""
	gobinListPath, gobinDirPath, err = findGobinFileSub(workingDirPath, homeDirPath)
	gobinLockPath = filepath.Join(gobinDirPath, gobinLockBase)
	if gobinDirPath == homeDirPath {
		gobinBinPath, err = getGlobalGobinPath()
		if err != nil {
			return
		}
	} else {
		gobinBinPath = filepath.Join(gobinDirPath, gobinLocalDirBase)
	}
	return
}

func findGobinFileSub(workingDirPath string, homeDirPath string) (
	gobinListPath string,
	gobinDirPath string,
	err error,
) {
	if err != nil {
		return
	}
	// Search for a project local gobin list file.
	gobinDirPath = workingDirPath
	for {
		if gobinDirPath == homeDirPath {
			break
		}
		gobinListPath = filepath.Join(gobinDirPath, gobinListBase)
		// Found a package local gobin list file.
		if _, err = os.Stat(gobinListPath); err == nil {
			return
		}
		if !os.IsNotExist(err) {
			return
		}
		parentDirPath := filepath.Dir(gobinDirPath)
		// Reached the root directory.
		if parentDirPath == gobinDirPath {
			break
		}
		gobinDirPath = parentDirPath
	}
	// Search for a gobin list file in the home directory.
	gobinDirPath = homeDirPath
	gobinListPath = filepath.Join(homeDirPath, gobinListBase)
	// Found a gobin list file in the home directory.
	if _, err = os.Stat(gobinListPath); err == nil {
		return
	}
	if !os.IsNotExist(err) {
		return
	}
	// When no gobin list file is found, create a gobin list file in the home directory.
	gobinDirPath = homeDirPath
	gobinListPath = filepath.Join(gobinDirPath, gobinListBase)
	return
}

// Bootstrap functions to be called from standalone main function.

func ensureGobinCmdInstalled() (cmdPath string, err error) {
	module := "github.com/knaka/gobin"
	pkg := module + ""
	name := filepath.Base(pkg)
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	_, _, gobinPath, err := findGobinFile(wd)
	ver := "latest"
	cmdPath = filepath.Join(gobinPath, name)
	if _, err = os.Stat(cmdPath); err == nil {
		return
	}
	divs := strings.Split(ver, ".")
	if len(divs) < 3 {
		cmd := exec.Command("go", "list", "-m", fmt.Sprintf("%s@%s", pkg, ver))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmdOutput, errX := cmd.Output()
		if errX != nil {
			return "", errX
		}
		cmdOutput2 := strings.TrimSpace(string(cmdOutput))
		divs = strings.SplitN(cmdOutput2, " ", 2)
		ver = divs[1]
	}
	pkgVer := fmt.Sprintf("%s@%s", pkg, ver)
	nameVer := fmt.Sprintf("%s@%s", name, ver)
	if _, err = os.Stat(filepath.Join(gobinPath, fmt.Sprintf("%s@%s", name, ver))); err == nil {
		return "", err
	}
	cmd := exec.Command("go", "install", pkgVer)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinPath))
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	err = os.Rename(filepath.Join(gobinPath, name), filepath.Join(gobinPath, fmt.Sprintf("%s@%s", name, ver)))
	if err != nil {
		return "", err
	}
	err = os.Symlink(fmt.Sprintf("%s@%s", name, ver), filepath.Join(gobinPath, name))
	if err != nil {
		return "", err

	}
	if err != nil {
		return
	}
	return filepath.Join(gobinPath, nameVer), nil
}

func gobinBoot(args []string) (err error) {
	cmdPath, err := ensureGobinCmdInstalled()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdPath, "install", "github.com/knaka/gobin@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return
	}
	cmd = exec.Command(cmdPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
	return nil
}

//goland:noinspection GoUnusedFunction
func apply(args []string) (err error) {
	return gobinBoot(append([]string{"apply"}, args...))
}

//goland:noinspection GoUnusedFunction
func install(args []string) (err error) {
	return gobinBoot(append([]string{"install"}, args...))
}

//goland:noinspection GoUnusedFunction
func run(args []string) (err error) {
	return gobinBoot(append([]string{"run"}, args...))
}
