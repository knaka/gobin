package minlib

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type goModParams struct {
	initialDirPath string
	global         bool
}

type GoModOption func(*goModParams) error

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

func WithInitialDir(initialDir string) GoModOption {
	return func(params *goModParams) (err error) {
		params.initialDirPath, err = canonAbs(initialDir)
		return
	}
}

func WithGlobal(f bool) GoModOption {
	return func(params *goModParams) error {
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

type PkgVerMapT map[string]string

const manifestFileBase = "Gobinfile"
const ManifestLockFileBase = "Gobinfile-lock.tsv"
const goModFileBase = "go.mod"
const gobinBase = ".gobin"

func ConfAndGobinPaths(opts ...GoModOption) (
	confDirPath string,
	gobinPath string,
	err error,
) {
	params := &goModParams{}
	for _, opt := range opts {
		err = (opt(params))
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
	for {
		if stat, errSub := os.Stat(filepath.Join(confDirPath, goModFileBase)); errSub == nil && !stat.IsDir() {
			break
		}
		confDirPath = filepath.Dir(confDirPath)
		if isRootDir(confDirPath) {
			confDirPath, gobinPath, err = "", "", errors.New("go.mod not found")
			return
		}
	}
	gobinPath = filepath.Join(confDirPath, gobinBase)
	return
}

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

type GoListOutput struct {
	Version string `json:"Version"`
}

func PkgVerMap(dirPath string) (lockList *PkgVerMapT, err error) {
	manifestLockPath := filepath.Join(dirPath, ManifestLockFileBase)
	if _, err_ := os.Stat(manifestLockPath); err_ != nil {
		return
	}
	reader := v(os.Open(manifestLockPath))
	scanner := bufio.NewScanner(reader)
	m := make(PkgVerMapT)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		divs := strings.SplitN(line, " ", 2)
		m[divs[0]] = divs[1]
	}
	return &m, nil
}

func Run() {
	confDirPath, gobinPath := v2(ConfAndGobinPaths())
	lockList := v(PkgVerMap(confDirPath))
	module := "github.com/knaka/gobin"
	pkgName := "github.com/knaka/gobin/cmd/gobin"
	ver, ok := (*lockList)[pkgName]
	if !ok {
		cmd := exec.Command("go", "list", "-m",
			"--json", fmt.Sprintf("%s@%s", module, "latest"))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmd.Stderr = os.Stderr
		goListOutput := GoListOutput{}
		output := v(cmd.Output())
		v0(json.Unmarshal(output, &goListOutput))
		ver = goListOutput.Version
	}
	base := path.Base(pkgName)
	pkgBaseVer := base + "@" + ver
	cmdPkgVerPath := filepath.Join(gobinPath, pkgBaseVer)
	if _, err := os.Stat(cmdPkgVerPath); err != nil {
		cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", module, ver))
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinPath))
		cmd.Stderr = os.Stderr
		v0(cmd.Run())
		_ = os.Remove(cmdPkgVerPath)
		v0(os.Rename(filepath.Join(gobinPath, base), cmdPkgVerPath))
		v0(os.Link(cmdPkgVerPath, filepath.Join(gobinPath, base)))
	}
	cmd := exec.Command(cmdPkgVerPath, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		os.Exit(0)
	}
	var execErr *exec.ExitError
	if errors.As(err, &execErr) {
		os.Exit(execErr.ExitCode())
	}
	log.Fatalf("Error: %+v", err)
}
