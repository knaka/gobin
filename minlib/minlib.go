package minlib

import (
	"bufio"
	"crypto/sha1"
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
	initialDirPath   string
	global           bool
	installationOnly bool
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

func installationOnly(f bool) ConfDirPathOption {
	return func(params *paramsT) error {
		params.installationOnly = f
		return nil
	}
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

const ManifestLockFileBase = "Gobinfile-lock"
const goModFileBase = "go.mod"
const gobinBase = ".gobin"

// ConfDirPath returns the configuration directory path (and the $GOBIN path).
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
func EnsureInstalled(gobinPath string, pkgPath string, ver string, tags string) (cmdPkgVerPath string, err error) {
	pkgBase := path.Base(pkgPath)
	pkgBaseVer := pkgBase + "@" + ver
	cmdPath := filepath.Join(gobinPath, pkgBase)
	cmdPkgVerPath = filepath.Join(gobinPath, pkgBaseVer)
	if tags != "" {
		hash := sha1.New()
		hash.Write([]byte(tags))
		sevenDigits := fmt.Sprintf("%x", hash.Sum(nil))[:7]
		cmdPkgVerPath += "-" + sevenDigits
	}
	if _, err_ := os.Stat(cmdPkgVerPath); err_ != nil {
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
		v0(os.Symlink(cmdPkgVerPath, cmdPath))
	}
	return
}

//goland:noinspection GoUnusedFunction
func run() {
	confDirPath, gobinPath := v2(ConfDirPath())
	pkgVerLockMap := v(PkgVerLockMap(confDirPath))
	modPath := "github.com/knaka/gobin"
	pkgPath := "github.com/knaka/gobin/cmd/gobin"
	tags := ""
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
	cmdPkgVerPath := v(EnsureInstalled(gobinPath, pkgPath, ver, tags))
	cmd := exec.Command(cmdPkgVerPath, append([]string{"run"}, os.Args[1:]...)...)
	cmd.Stdin = os.Stdin
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
