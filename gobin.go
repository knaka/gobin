package gobin

import (
	"github.com/samber/lo"
	"io"
)

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

type installParams struct {
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

type Option func(params *installParams) error

//goland:noinspection GoUnusedExportedFunction
func Global(f bool) Option {
	return func(params *installParams) (err error) {
		params.global = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithDir(dir string) Option {
	return func(params *installParams) (err error) {
		params.Dir = dir
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func Verbose(f bool) Option {
	return func(params *installParams) (err error) {
		params.verbose = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithEnv(env []string) Option {
	return func(params *installParams) (err error) {
		params.Env = env
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithGobinPath(f bool) Option {
	return func(params *installParams) (err error) {
		params.WithGobinPath = f
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStdin(stdin io.Reader) Option {
	return func(params *installParams) (err error) {
		params.stdin = stdin
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStdout(stdout io.Writer) Option {
	return func(params *installParams) (err error) {
		params.stdout = stdout
		return
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStderr(stderr io.Writer) Option {
	return func(params *installParams) (err error) {
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
	log.Printf("Querying version for %s\n", pkg)
	for _, candidate := range V(candidateModules(pkg)) {
		cmd := exec.Command("go", "list", "-m",
			"--json", fmt.Sprintf("%s@%s", candidate, latestVer))
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

func newInstallParams() *installParams {
	return &installParams{
		WithGobinPath: true,
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}
}

func install(targets []string, global bool, confDirPath string, gobinPath string) (cmdPath string, err error) {
	defer Catch(&err)
	for {
		if len(targets) == 0 {
			break
		}
		target := targets[0]
		targets = targets[1:]
		if !global {
			goModDef := V(parseGoMod(confDirPath))
			reqMod := goModDef.requiredModuleByPkg(target)
			if reqMod != nil {
				cmdPath = Elvis(cmdPath, V(minlib.EnsureInstalled(gobinPath, target, reqMod.Version, "", log.Logger)))
				continue
			}
		}
		manifest := V(parseManifest(confDirPath))
		shouldSave := false
		entry := manifest.lookup(target)
		if entry != nil {
			targets = append(targets, entry.Requires...)
			if entry.LockedVersion == latestVer {
				entry.LockedVersion = V(queryVersion(entry.Pkg))
				shouldSave = true
			}
			cmdPath = Elvis(cmdPath, V(minlib.EnsureInstalled(gobinPath, entry.Pkg, entry.LockedVersion, entry.Tags, log.Logger)))
			if shouldSave {
				V0(manifest.saveLockfile())
			}
			continue
		}
		err = errors.New(fmt.Sprintf("command “%s” is not defined", target))
		return
	}
	return
}

func InstallEx(patterns []string, opts ...Option) (err error) {
	defer Catch(&err)
	params := newInstallParams()
	for _, opt := range opts {
		V0(opt(params))
	}
	if params.verbose {
		log.SetOutput(params.stderr)
	}
	goModOptions := []minlib.ConfDirPathOption{
		minlib.WithGlobal(params.global),
	}
	confDirPath, gobinPath := V2(minlib.ConfDirPath(goModOptions...))
	_ = V(install(patterns, params.global, confDirPath, gobinPath))
	return
}

//goland:noinspection GoUnusedExportedFunction
func Install(patterns ...string) (err error) {
	return InstallEx(patterns)
}

func CommandEx(args []string, opts ...Option) (cmd *exec.Cmd, err error) {
	defer Catch(&err)
	if len(args) == 0 {
		err = errors.New("no command specified")
		return
	}
	params := newInstallParams()
	for _, opt := range opts {
		V0(opt(params))
	}
	if params.verbose {
		log.SetOutput(params.stderr)
	}
	goModOptions := []minlib.ConfDirPathOption{
		minlib.WithGlobal(params.global),
	}
	confDirPath, gobinPath := V2(minlib.ConfDirPath(goModOptions...))
	cmdPath := V(install([]string{args[0]}, params.global, confDirPath, gobinPath))
	cmd = exec.Command(cmdPath, args[1:]...)
	cmd.Stdin = params.stdin
	cmd.Stdout = params.stdout
	cmd.Stderr = params.stderr
	cmd.Env = os.Environ()
	if params.Env != nil {
		cmd.Env = append(cmd.Env, params.Env...)
	}
	if params.WithGobinPath {
		cmd.Env = append(cmd.Env, "PATH="+gobinPath+string(filepath.ListSeparator)+os.Getenv("PATH"))
	}
	return
}

//goland:noinspection GoUnusedExportedFunction
func Command(args ...string) (cmd *exec.Cmd, err error) {
	return CommandEx(args)
}

//goland:noinspection GoUnusedExportedFunction
func RunEx(args []string, opts ...Option) (err error) {
	defer Catch(&err)
	cmd := V(CommandEx(args, opts...))
	err = cmd.Run()
	if err == nil {
		os.Exit(0)
	}
	errExit := ErrorAs[*exec.ExitError](err)
	if errExit != nil {
		os.Exit(errExit.ExitCode())
	}
	return
}

//goland:noinspection GoUnusedExportedFunction
func Run(args ...string) (err error) {
	return RunEx(args)
}

func UpdateEx(patterns []string, opts ...Option) (err error) {
	defer Catch(&err)
	params := newInstallParams()
	for _, opt := range opts {
		V0(opt(params))
	}
	if params.verbose {
		log.SetOutput(params.stderr)
	}
	goModOptions := []minlib.ConfDirPathOption{
		minlib.WithGlobal(params.global),
	}
	confDirPath, _ := V2(minlib.ConfDirPath(goModOptions...))
	manifest := V(parseManifest(confDirPath))
	var latestEntries []*maniEntry
	if len(patterns) == 0 {
		latestEntries = lo.Filter(manifest.Entries(), func(entry *maniEntry, _ int) (f bool) {
			if entry.Version == latestVer {
				f = true
			}
			return
		})
	} else {
		latestEntries = lo.FilterMap(patterns, func(pattern string, _ int) (entry *maniEntry, f bool) {
			entry = manifest.lookup(pattern)
			if entry == nil {
				Throw(errors.New(fmt.Sprintf("command “%s” is not defined", pattern)))
			}
			if entry.Version == latestVer {
				f = true
			}
			return
		})
	}
	for _, entry := range latestEntries {
		entry.LockedVersion = V(queryVersion(entry.Pkg))
	}
	V0(manifest.saveLockfile())
	return
}
