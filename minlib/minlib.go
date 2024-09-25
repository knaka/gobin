package minlib

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	stdlog "log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var verbose = false

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

// realpath returns the canonical absolute path of the given value.
func realpath(s string) (ret string, err error) {
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
		params.initialDirPath, err = realpath(initialDir)
		return
	}
}

func WithGlobal(f bool) ConfDirPathOption {
	return func(params *paramsT) error {
		params.global = f
		return nil
	}
}

var mockRootDirPath = ""

// setMockRootDir sets the root directory path for testing.
// Introduced because the temporary directory of Windows is under the user home directory.
func setMockRootDir(dir string) {
	mockRootDirPath = v(realpath(dir))
}

func isRootDir(dir string) bool {
	dirPath, err := realpath(dir)
	if err != nil {
		return false
	}
	if mockRootDirPath != "" {
		return dirPath == mockRootDirPath
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

func GlobalConfDirPath() (confDirPath string, gobinPath string, err error) {
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

func parentDirOf(dir string) string {
	dirPath := v(realpath(dir))
	if mockRootDirPath != "" && mockRootDirPath == dirPath {
		return dir
	}
	return filepath.Dir(dirPath)
}

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
		return GlobalConfDirPath()
	}
	confDirPath = params.initialDirPath
	if confDirPath == "" {
		confDirPath, err = realpath(".")
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
		confDirPath = parentDirOf(confDirPath)
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
func EnsureInstalled(gobinPath string, pkgPath string, ver string, tags string, log *stdlog.Logger, _ *stdlog.Logger) (cmdPkgVerPath string, err error) {
	pkgBase := path.Base(pkgPath)
	pkgBaseVer := pkgBase + "@" + ver
	cmdPath := filepath.Join(gobinPath, pkgBase+ExeExt())
	if tags != "" {
		hash := sha1.New()
		hash.Write([]byte(tags))
		sevenDigits := fmt.Sprintf("%x", hash.Sum(nil))[:7]
		pkgBaseVer += "-" + sevenDigits
	}
	cmdPkgVerPath = filepath.Join(gobinPath, pkgBaseVer+ExeExt())
	if _, err_ := os.Stat(cmdPkgVerPath); err_ != nil {
		log.Printf("Installing %s@%s\n", pkgPath, ver)
		cmd := exec.Command(getGoCmd(), "install", fmt.Sprintf("%s@%s", pkgPath, ver))
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinPath))
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		_ = os.Remove(cmdPath)
		err = cmd.Run()
		if err != nil {
			return
		}
		_ = os.Remove(cmdPkgVerPath)
		err = os.Rename(cmdPath, cmdPkgVerPath)
		if err != nil {
			return
		}
		if pkgBase == GobinCmdBase {
			v0(os.Symlink(pkgBaseVer+ExeExt(), cmdPath))
		} else {
			v0(os.Symlink(GobinCmdBase+ExeExt(), cmdPath))
		}
	}
	return
}

func getGoroot() (gobinPath string, err error) {
	if goPath, err := exec.LookPath("go"); err == nil {
		return goPath, nil
	}
	var dirPaths []string
	dirPaths = append(dirPaths, "/usr/local/go")
	dirPaths = append(dirPaths, v(filepath.Glob("/usr/local/Cellar/go/*"))...)
	dirPaths = append(dirPaths, "/Program Files/Go")
	for _, dirPath := range dirPaths {
		if _, err := exec.LookPath(filepath.Join(dirPath, "bin", "go")); err == nil {
			return filepath.Join(dirPath, "bin"), nil
		}
	}
	var latestDirPath string
	dirPaths = v(filepath.Glob(filepath.Join(os.Getenv("HOME"), "sdk", "go1*")))
	dirPaths = append(dirPaths, v(filepath.Glob(filepath.Join(os.Getenv("HOME"), "go", "go1*")))...)
	for _, dirPath := range dirPaths {
		if latestDirPath == "" {
			latestDirPath = dirPath
			continue
		}
		if filepath.Base(dirPath) > filepath.Base(latestDirPath) {
			latestDirPath = dirPath
			continue
		}
	}
	if latestDirPath != "" {
		return filepath.Join(latestDirPath, "bin"), nil
	}
	ver := "1.23.1"
	homeDir := v(os.UserHomeDir())
	sdkDirPath := filepath.Join(homeDir, "sdk")
	goRoot := filepath.Join(sdkDirPath, "go"+ver)
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		tempDir := v(os.MkdirTemp("", ""))
		defer (func() { v0(os.RemoveAll(tempDir)) })()
		zipPath := filepath.Join(tempDir, "temp.zip")
		url := fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.zip", ver, runtime.GOOS, runtime.GOARCH)
		cmd := exec.Command("curl.exe", "--location", "-o", zipPath,
			url)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		v0(cmd.Run())
		cmd = exec.Command("tar.exe", "-C", sdkDirPath, "-xzf", zipPath)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		v0(cmd.Run())
	} else {
		cmd := exec.Command("curl", "--location", "-o", "-",
			fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.tar.gz", ver, runtime.GOOS, runtime.GOARCH))
		cmd.Stderr = os.Stderr
		cmd.Dir = sdkDirPath
		v0(cmd.Run())
	}
	//	Then rename.
	v0(os.Rename(filepath.Join(sdkDirPath, "go"), goRoot))
	return filepath.Join(goRoot, "bin"), nil
}

func getGoCmd() string {
	cmdPath := filepath.Join(v(getGoroot()), "go"+ExeExt())
	if verbose {
		log.Printf("The path to the go command is %s\n", cmdPath)
	}
	return cmdPath
}

func EnsureGobinCmdInstalled(global bool) (cmdPath string, err error) {
	var opts []ConfDirPathOption
	if global {
		opts = append(opts, WithGlobal(true))
	}
	confDirPath, gobinPath := v2(ConfDirPath(opts...))
	pkgVerLockMap := v(PkgVerLockMap(confDirPath))
	modPath := "github.com/knaka/gobin"
	pkgPath := "github.com/knaka/gobin/cmd/gobin"
	ver, ok := pkgVerLockMap[pkgPath]
	if ok {
		if verbose {
			log.Printf("The locked version of %s is %s\n", pkgPath, ver)
		}
	} else {
		if verbose {
			log.Printf("Querying the latest version of %s\n", pkgPath)
		}
		cmd := exec.Command(getGoCmd(), "list", "-m",
			"--json", fmt.Sprintf("%s@%s", modPath, "latest"))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmd.Stderr = os.Stderr
		output := v(cmd.Output())
		goListOutput := GoListOutput{}
		v0(json.Unmarshal(output, &goListOutput))
		ver = goListOutput.Version
		if verbose {
			log.Printf("The latest version of %s is %s\n", pkgPath, ver)
		}
		manifestLockPath := filepath.Join(confDirPath, ManifestLockFileBase)
		writer := v(os.OpenFile(manifestLockPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600))
		defer (func() { v0(writer.Close()) })()
		_ = v(writer.WriteString(fmt.Sprintf("%s@%s\n", pkgPath, ver)))
	}
	return EnsureInstalled(gobinPath, pkgPath, ver, "", stdlog.Default(), stdlog.Default())
}

func Command(name string, arg ...string) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return
}

func RunCommand(name string, arg ...string) (execErr *exec.ExitError, err error) {
	cmd, err := Command(name, arg...)
	if err != nil {
		return
	}
	err = cmd.Run()
	errors.As(err, &execErr)
	return
}

// bootstrapMain is the main function of the bootstrap file.
func bootstrapMain() {
outer:
	for {
		if len(os.Args) <= 1 {
			break
		}
		switch os.Args[1] {
		case "--verbose":
			verbose = true
		default:
			break outer
		}
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
	gobinCmdPath := v(EnsureGobinCmdInstalled(false))
	errExec, err := RunCommand(gobinCmdPath, os.Args[1:]...)
	if err == nil {
		os.Exit(0)
	}
	if errExec != nil {
		os.Exit(errExec.ExitCode())
	}
	stdlog.Fatalf("Error 560d8bf: %+v", err)
}

var ExeExt = sync.OnceValue(func() (exeExt string) {
	switch runtime.GOOS {
	case "windows":
		exeExt = ".exe"
	}
	return
})
