package lib

// Do not use 3rd party packages because this file is used to generate standalone go program files.
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const gobinDirBase = ".gobin"

var gobinListBases = []string{
	"Gobinfile",
	".Gobinfile",
}

// GoEnv is a struct to hold the output of `go env -json`.
type GoEnv struct {
	Version string `json:"GOVERSION"`
	Gobin   string `json:"GOBIN"`
	Gopath  string `json:"GOPATH"`
}

var getGoCmdPath = sync.OnceValues(func() (goCmdPath string, err error) {
	goCmdPath = filepath.Join(runtime.GOROOT(), "bin", "go")
	if stat, err := os.Stat(goCmdPath); err == nil && !stat.IsDir() {
		return goCmdPath, nil
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

func findGobinListFile(dirPath string) (gobinListPath string, gobinPath string, err error) {
	for {
		for _, gobinListBase := range gobinListBases {
			pkgListPath := filepath.Join(dirPath, gobinListBase)
			_, err = os.Stat(pkgListPath)
			if err == nil {
				return pkgListPath, filepath.Join(dirPath, gobinDirBase), nil
			}
			if !os.IsNotExist(err) {
				return "", "", err
			}
		}
		parent := filepath.Dir(dirPath)
		if parent == dirPath {
			break
		}
		dirPath = parent
	}
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	globalGobinPath, err := getGlobalGobinPath()
	if err != nil {
		return "", "", err
	}
	for _, gobinListBase := range gobinListBases {
		pkgListPath := filepath.Join(homeDirPath, gobinListBase)
		_, err = os.Stat(pkgListPath)
		if err == nil {
			return pkgListPath, globalGobinPath, nil
		}
		if !os.IsNotExist(err) {
			return "", "", err
		}
	}
	return "", "", fmt.Errorf("no gobin list file found")
}

// Bootstrap functions to be called from standalone main function.

func ensureGobinCmdInstalled() (cmdPath string, err error) {
	module := "github.com/knaka/gobin"
	ver := "latest"
	pkg := module + ""
	name := filepath.Base(pkg)
	moduleVer := fmt.Sprintf("%s@%s", module, ver)
	pkgVer := fmt.Sprintf("%s@%s", pkg, ver)
	gobinPath, err := getGlobalGobinPath()
	if err != nil {
		return "", err
	}
	// Check if the binary of any version is already installed.
	if _, err = os.Stat(filepath.Join(gobinPath, name)); err == nil {
		return filepath.Join(gobinPath, name), err
	}
	cmd := exec.Command("go", "list", "-m", moduleVer)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}
	divs := strings.SplitN(string(cmdOutput), " ", 2)
	resolvedVer := divs[1]
	if _, err = os.Stat(filepath.Join(gobinPath, fmt.Sprintf("%s@%s", name, resolvedVer))); err == nil {
		return "", err
	}
	cmd = exec.Command("go", "install", pkgVer)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	err = os.Rename(filepath.Join(gobinPath, name), filepath.Join(gobinPath, fmt.Sprintf("%s@%s", name, resolvedVer)))
	if err != nil {
		return "", err
	}
	err = os.Symlink(fmt.Sprintf("%s@%s", name, resolvedVer), filepath.Join(gobinPath, name))
	if err != nil {
		return "", err

	}
	return filepath.Join(gobinPath, name), nil
}

func gobin(args []string) (err error) {
	cmdPath, err := ensureGobinCmdInstalled()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
	return nil
}

func gobinApply(args []string) (err error) {
	return gobin(append([]string{"apply"}, args...))
}

func gobinInstall(args []string) (err error) {
	return gobin(append([]string{"install"}, args...))
}

func gobinRun(args []string) (err error) {
	return gobin(append([]string{"run"}, args...))
}
