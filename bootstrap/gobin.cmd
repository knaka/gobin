@echo off
setlocal enabledelayedexpansion

@REM All releases - The Go Programming Language https://go.dev/dl/
set "ver=1.23.1"

set "exit_code=1"

:unique_temp_loop
set "temp_dir_path=%TEMP%\%~n0-%RANDOM%"
if exist "!temp_dir_path!" goto unique_temp_loop
mkdir "!temp_dir_path!" || goto :exit

@REM Command in %PATH%
where go >nul 2>&1
if !ERRORLEVEL! == 0 (
    set "go_cmd_path=go"
    goto found_go_cmd
)
@REM Trivial installation paths
set "dirs=%USERPROFILE%\go\go!ver!\bin;%USERPROFILE%\sdk\go!ver!\bin;\Program Files\Go\bin"
for %%d in ("!dirs:;=" "!") do (
    if exist "%%d\go.exe" (
        set "go_cmd_path=%%d\go.exe"
        goto found_go_cmd
    )
)
@REM Download if not found
set "goos=windows"
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "goarch=amd64"
) else if "%PROCESSOR_ARCHITEW6432%"=="AMD64" (
    set "goarch=amd64"
) else (
    goto :exit
)
set "sdk_dir_path=%USERPROFILE%\sdk"
if not exist "!sdk_dir_path!" (
    mkdir "!sdk_dir_path!" || goto :exit
)
set "zip_path=!temp_dir_path!\go.zip"
curl.exe --fail --location -o "!zip_path!" "https://go.dev/dl/go!ver!.%goos%-%goarch%.zip" || goto :exit
cd "!sdk_dir_path!" || goto :exit
tar.exe -xf "!zip_path!" || goto :exit
move /y "!sdk_dir_path!\go" "!sdk_dir_path!\go!ver!" || goto :exit
set "go_cmd_path=!sdk_dir_path!\go!ver!\bin\go.exe"

:found_go_cmd
if not defined GOPATH (
    set "GOPATH=%USERPROFILE%\go"
)
if not exist "!GOPATH!\bin" (
    mkdir !%GOPATH!\bin"
)

set "name=embedded-%~f0"
set "name=!name:\=_!"
set "name=!name::=_!"
set "name=!name:/=_!"
if exist "!GOPATH!\bin\!name!.exe" (
    xcopy /l /d /y "!GOPATH!\bin\!name!.exe" "%~f0" | findstr /b /c:"1 " >nul 2>&1
    if !ERRORLEVEL! == 0 (
        goto :execute
    )
)

:build
for /f "usebackq tokens=1 delims=:" %%i in (`findstr /n /b :embed_53c8fd5 "%~f0"`) do set n=%%i
set "tempfile=!temp_dir_path!\!name!.go"
more +%n% "%~f0" > "!tempfile!"

!go_cmd_path! build -o !GOPATH!\bin\!name!.exe "!tempfile!" || goto :exit
del /q "!temp_dir_path!"
goto :execute

:execute
!GOPATH!\bin\!name!.exe %* || goto :exit
set "exit_code=0"

:exit
if exist "!temp_dir_path!" (
    del /q "!temp_dir_path!"
)
exit /b !exit_code!

endlocal

:embed_53c8fd5
// Code generated by gen-bootstrap; DO NOT EDIT.

// Latest version is available by running:
//
//   curl --remote-name https://raw.githubusercontent.com/knaka/gobin/main/bootstrap/gobin.go

//go:build ignore

package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func getGoroot() (gobinPath string, err error) {
	envGoRoot := os.Getenv("GOROOT")
	if envGoRoot != "" {
		return filepath.Join(envGoRoot, "bin"), nil
	}
	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "env", "GOROOT")
		cmd.Stderr = os.Stderr
		output := v(cmd.Output())
		goRoot := strings.TrimSpace(string(output))
		return filepath.Join(goRoot, "bin"), nil
	}
	var dirPaths []string
	for _, drPath := range v(filepath.Glob(filepath.Join(os.Getenv("HOME"), "sdk", "go*"))) {
		dirPaths = append(dirPaths, drPath)
	}
	dirPaths = append(dirPaths, "/usr/local/go")
	dirPaths = append(dirPaths, v(filepath.Glob("/usr/local/Cellar/go/*"))...)
	dirPaths = append(dirPaths, "/Program Files/Go")
	dirPaths = append(dirPaths, filepath.Join(os.Getenv("HOME"), "go"))
	for _, dirPath := range dirPaths {
		if _, err := exec.LookPath(filepath.Join(dirPath, "bin", "go")); err == nil {
			return filepath.Join(dirPath, "bin"), nil
		}
	}
	ver := "1.23.1"
	homeDir := v(os.UserHomeDir())
	sdkDirPath := filepath.Join(homeDir, "sdk")
	goRoot := filepath.Join(sdkDirPath, "go"+ver)
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		tempDir := v(os.MkdirTemp("", ""))
		zipPath := filepath.Join(tempDir, "temp.zip")
		url := fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.zip", ver, runtime.GOOS, runtime.GOARCH)
		cmd := exec.Command("curl.exe", "--location", "-o", zipPath,
			url)
		cmd.Stderr = os.Stderr
		v0(cmd.Run())
		unzip(zipPath, sdkDirPath)
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
	return filepath.Join(v(getGoroot()), "go"+ExeExt())
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
	if !ok {
		cmd := exec.Command(getGoCmd(), "list", "-m",
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

func main() {
	bootstrapMain()
}

