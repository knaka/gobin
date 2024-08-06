package lib

import (
	"bufio"
	"errors"
	"github.com/knaka/gobin/log"
	"github.com/knaka/gobin/minlib"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	gomodule "golang.org/x/mod/module"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	. "github.com/knaka/go-utils"
)

type InstallExParams struct {
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

type Opt func(params *InstallExParams) error

func WithGlobal(f bool) Opt {
	return func(params *InstallExParams) (err error) {
		params.global = f
		return
	}
}

func WithDir(dir string) Opt {
	return func(params *InstallExParams) (err error) {
		params.Dir = dir
		return
	}
}

func Verbose(f bool) Opt {
	return func(params *InstallExParams) (err error) {
		params.verbose = f
		return
	}
}

func WithEnv(env []string) Opt {
	return func(params *InstallExParams) (err error) {
		params.Env = env
		return
	}
}

func WithGobinPath(f bool) Opt {
	return func(params *InstallExParams) (err error) {
		params.WithGobinPath = f
		return
	}
}

func WithStdin(stdin io.Reader) Opt {
	return func(params *InstallExParams) (err error) {
		params.stdin = stdin
		return
	}
}

func WithStdout(stdout io.Writer) Opt {
	return func(params *InstallExParams) (err error) {
		params.stdout = stdout
		return
	}
}

func WithStderr(stderr io.Writer) Opt {
	return func(params *InstallExParams) (err error) {
		params.stderr = stderr
		return
	}
}

type thePkg struct {
	module string
	pkg    string
	ver    string
}

func search(command string, confDirPath string) (toinstall []*thePkg) {
	goMod_ := V(parseGoMod(confDirPath))
	reqMod := goMod_.requiredModuleByPkg(command)
	if reqMod != nil {
		toinstall = append(toinstall,
			&thePkg{
				module: reqMod.Path,
				pkg:    command,
				ver:    reqMod.Version,
			},
		)
	}
	pkgVerMap := V(minlib.PkgVerMap(confDirPath))
	for pkg_, lockedVer := range pkgVerMap {
		if pkg_ == command || path.Base(pkg_) == command {
			toinstall = append(toinstall, &thePkg{
				module: "",
				pkg:    pkg_,
				ver:    lockedVer,
			})
			return
		}
	}

	return
}

func CommandEx(args []string, opts ...Opt) (cmd *exec.Cmd, err error) {
	defer Catch(&err)
	params := InstallExParams{
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
	goModOptions := []minlib.GoModOption{
		minlib.WithGlobal(params.global),
	}

	if len(args) == 0 {
		err = errors.New("no command specified")
		return
	}

	command := args[0]

	confDirPath /* gobinPath */, _ := V2(minlib.ConfAndGobinPaths(goModOptions...))

	search(command, confDirPath)
	return

}

const goModBase = "go.mod"

// goModDefT represents the go.mod module definition file.
type goModDefT struct {
	name            string
	requiredModules []*gomodule.Version
}

func (mod *goModDefT) requiredModule(moduleName string) (x *gomodule.Version) {
	for _, req := range mod.requiredModules {
		if req.Path == moduleName {
			x = req
			return
		}
	}
	return
}

func (mod *goModDefT) requiredModuleByPkg(pkgName string) (x *gomodule.Version) {
	for _, reqMod := range mod.requiredModules {
		if reqMod.Path == pkgName || strings.HasPrefix(pkgName, reqMod.Path+"/") {
			x = reqMod
			return
		}
	}
	return
}

func parseGoMod(dirPath string) (goModDef *goModDefT, err error) {
	defer Catch(&err)
	filePath := TernaryF(
		filepath.Base(dirPath) == goModBase,
		func() string { return dirPath },
		func() string { return filepath.Join(dirPath, goModBase) },
	)
	goModFile := V(modfile.Parse(filePath, V(os.ReadFile(filePath)), nil))
	goModDef = &goModDefT{
		name: goModFile.Module.Mod.Path,
		requiredModules: lo.Map(goModFile.Require, func(reqMod *modfile.Require, _ int) *gomodule.Version {
			return &reqMod.Mod
		}),
	}
	return
}

type maniEntry struct {
	Pkg      string
	Version  string `json:"version"`
	Tags     string
	Requires []string
}

// maniT is the internal representation of the manifest and the manifest lock file.
type maniT struct {
	filePath  string
	entries   []*maniEntry
	lockPath  string
	pkgMapVer minlib.PkgVerMapT
}

const maniBase = "Gobinfile"
const maniLockBase = "Gobinfile-lock"

var reSpaces = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`\s+`) })

func parseManifest(dirPath string) (gobinManifest *maniT, err error) {
	defer Catch(&err)
	gobinManifest = &maniT{
		filePath: filepath.Join(dirPath, maniBase),
		lockPath: filepath.Join(dirPath, maniLockBase),
	}
	if _, err_ := os.Stat(gobinManifest.filePath); err_ == nil {
		reader := V(os.Open(gobinManifest.filePath))
		defer (func() { V0(reader.Close()) })()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			divs := strings.SplitN(line, "#", 2)
			line = strings.TrimSpace(divs[0])
			divs = reSpaces().Split(line, 2)
			pkgVer := divs[0]
			optsStr := TernaryF(len(divs) >= 2,
				func() string { return divs[1] },
				func() string { return "" },
			)
			var requires []string
			var tags string
			if optsStr != "" {
				divs = reSpaces().Split(optsStr, -1)
				for _, opt := range divs {
					x := strings.SplitN(opt, "=", 2)
					if len(x) < 2 {
						continue
					}
					key := x[0]
					val := x[1]
					switch key {
					case "requires":
						reqs := strings.Split(val, ",")
						for _, req := range reqs {
							requires = append(requires, req)
						}
					case "tags":
						tags = val
					}
				}
			}
			divs = strings.SplitN(pkgVer, "@", 2)
			pkg := divs[0]
			ver := TernaryF(len(divs) >= 2,
				func() string { return Ternary(divs[1] == "latest", "", divs[1]) },
				func() string { return "" },
			)
			gobinManifest.entries = append(gobinManifest.entries, &maniEntry{
				Pkg:      pkg,
				Version:  ver,
				Tags:     tags,
				Requires: requires,
			})
		}
	}
	if _, err_ := os.Stat(gobinManifest.lockPath); err_ == nil {
		gobinManifest.pkgMapVer = V(minlib.PkgVerMap(dirPath))
	}
	for _, entry := range gobinManifest.entries {
		if lockedVer, ok := gobinManifest.pkgMapVer[entry.Pkg]; ok {
			entry.Version = lockedVer
		}
	}
	return
}

func (mani *maniT) save() (err error) {
	return mani.saveAs(mani.lockPath)
}

func (mani *maniT) saveAs(filePath string) (err error) {
	defer Catch(&err)
	writer := V(os.Create(filePath))
	defer (func() { V0(writer.Close()) })()
	for _, entry := range mani.entries {
		_, err = writer.WriteString(entry.Pkg + " " + entry.Version + "\n")
	}
	return
}

func (mani *maniT) lookup(pattern string) (entry *maniEntry) {
	divs := strings.SplitN(pattern, "@", 2)
	pkg := ""
	base := ""
	if len(divs) == 2 {
		pkg = divs[0]
	} else if strings.Contains(pattern, "/") {
		pkg = pattern
	} else {
		base = pattern
	}
	if pkg == "" && base != "" {
		for _, entry_ := range mani.entries {
			pkgBase := path.Base(entry_.Pkg)
			if pkgBase == base {
				pkg = entry_.Pkg
			}
		}
	}
	if pkg == "" {
		return
	}
	for _, entry_ := range mani.entries {
		if entry_.Pkg == pkg {
			entry = entry_
		}
	}
	return
}
