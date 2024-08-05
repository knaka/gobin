package minlib

import (
	"errors"
	"os"
	"path/filepath"
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

func withGlobal(f bool) GoModOption {
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

const manifestFileBase = "Gobinfile"
const manifestLockFileBase = "Gobinfile-lock"
const goModFileBase = "go.mod"
const gobinBase = ".gobin"

func ConfAndGobinPaths(opts ...GoModOption) (
	confDirPath string,
	gobinDirPath string,
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
		gobinDirPath = os.Getenv("GOBIN")
		if gobinDirPath == "" {
			gobinDirPath = filepath.Join(confDirPath, "go", "bin")
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
			confDirPath, gobinDirPath, err = "", "", errors.New("go.mod not found")
			return
		}
	}
	gobinDirPath = filepath.Join(confDirPath, gobinBase)
	return
}
