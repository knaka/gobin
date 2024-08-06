package gobin

import "io"

//go:generate_input gen-bootstrap/* minlib/minlib.go
//go:generate_output gobin-run.go
//go:generate go run ./gen-bootstrap

// //go:generate go run gobin-run.go gomplate --help
// //go:generate go run gobin-run.go golang.org/x/tools/cmd/stringer -h

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/knaka/go-utils"
	"github.com/knaka/gobin/log"
	"github.com/knaka/gobin/minlib"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type installExParams struct {
	Dir             string
	Env             []string
	WithGobinPath   bool
	shouldReturnCmd bool
	stdin           io.Reader
	stdout          io.Writer
	stderr          io.Writer
	verbose         bool
	global          bool
}

type Option func(params *installExParams) error

//goland:noinspection GoUnusedExportedFunction
func WithGlobal(f bool) Option {
	return func(params *installExParams) (err error) {
		params.global = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithDir(dir string) Option {
	return func(params *installExParams) (err error) {
		params.Dir = dir
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func Verbose(f bool) Option {
	return func(params *installExParams) (err error) {
		params.verbose = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithEnv(env []string) Option {
	return func(params *installExParams) (err error) {
		params.Env = env
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithGobinPath(f bool) Option {
	return func(params *installExParams) (err error) {
		params.WithGobinPath = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStdin(stdin io.Reader) Option {
	return func(params *installExParams) (err error) {
		params.stdin = stdin
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStdout(stdout io.Writer) Option {
	return func(params *installExParams) (err error) {
		params.stdout = stdout
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStderr(stderr io.Writer) Option {
	return func(params *installExParams) (err error) {
		params.stderr = stderr
		return
	}
}

func candidateModules(pkg string) (ret []string, err error) {
	divs := strings.Split(pkg, "/")
	for {
		if len(divs) == 0 {
			break
		}
		ret = append(ret, strings.Join(divs, "/"))
		divs = divs[:len(divs)-1]
	}
	return
}

func queryVersion(pkg string) (version string, err error) {
	for _, candidate := range V(candidateModules(pkg)) {
		cmd := exec.Command("go", "list", "-m",
			"--json", fmt.Sprintf("%s@%s", candidate, "latest"))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		goListOutput := minlib.GoListOutput{}
		output, err_ := cmd.Output()
		if err_ != nil {
			continue
		}
		V0(json.Unmarshal(output, &goListOutput))
		version = goListOutput.Version
		break
	}
	return
}

func CommandEx(args []string, opts ...Option) (cmd *exec.Cmd, err error) {
	defer Catch(&err)
	params := installExParams{
		WithGobinPath: true,
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}
	for _, opt := range opts {
		V0(opt(&params))
	}
	if params.verbose {
		log.SetOutput(params.stderr)
	}
	goModOptions := []minlib.ConfDirPathOption{
		minlib.WithGlobal(params.global),
	}
	if len(args) == 0 {
		err = errors.New("no command specified")
		return
	}
	confDirPath, gobinPath := V2(minlib.ConfDirPath(goModOptions...))
	var cmdPath string
	for {
		if !params.global {
			goModDef := V(parseGoMod(confDirPath))
			reqMod := goModDef.requiredModuleByPkg(args[0])
			if reqMod != nil {
				cmdPath = V(minlib.EnsureInstalled(gobinPath, args[0], reqMod.Version))
				break
			}
		}
		manifest := V(parseManifest(confDirPath))
		shouldSave := false
		entry := manifest.lookup(args[0])
		if entry != nil {
			if entry.Version == "" {
				entry.Version = V(queryVersion(entry.Pkg))
				shouldSave = true
			}
			cmdPath = V(minlib.EnsureInstalled(gobinPath, entry.Pkg, entry.Version))
			if shouldSave {
				V0(manifest.saveLockfile())
			}
			break
		}
		err = errors.New("no module provides the command")
		return
	}
	cmd = exec.Command(cmdPath, args[1:]...)
	cmd.Stdin = params.stdin
	cmd.Stdout = params.stdout
	cmd.Stderr = params.stderr
	if params.WithGobinPath {
		cmd.Env = append(os.Environ(), "PATH="+gobinPath+string(filepath.ListSeparator)+os.Getenv("PATH"))
	}
	return
}
