package lib

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mattn/go-shellwords"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	. "github.com/knaka/go-utils"
)

type GoListOutput struct {
	Version string `json:"Version"`
}

var logger *zap.SugaredLogger

func init() {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.InfoLevel)
	logger = V(config.Build()).Sugar()
}

func SetVerbose() {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.DebugLevel)
	logger = V(config.Build()).Sugar()
}

var rePackageBuildOptsSeparator = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`\s+`)
})

func getGobinList(dirPath string) (
	gobinList GobinList, // Parsed `Gobinfile`
	gobinLock GobinList, // Parsed `Gobinfile-lock.json`
	gobinBinPath string, // ~/go/bin or local .gobin directory
	err error,
) {
	defer Catch(&err)
	gobinListPath, gobinLockPath, gobinBinPath := V3(findGobinFile(dirPath))
	gobinList.Path = gobinListPath
	gobinList.Map = make(GobinMap)
	if stat, err := os.Stat(gobinListPath); err == nil && !stat.IsDir() {
		reader := V(os.Open(gobinListPath))
		defer (func() { V0(reader.Close()) })()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			divs := strings.SplitN(line, "#", 2)
			pkgVerBuildOpts := strings.TrimSpace(divs[0])
			if pkgVerBuildOpts == "" {
				continue
			}
			//comment := ternaryF(len(divs) >= 2, func() string { return strings.TrimSpace(divs[1]) }, nil)
			divs = rePackageBuildOptsSeparator().Split(pkgVerBuildOpts, 2)
			pkgVer := divs[0]
			optsStr := ternaryF(len(divs) >= 2, func() string { return divs[1] }, nil)
			opts := V(shellwords.Parse(optsStr))
			divs = strings.SplitN(pkgVer, "@", 2)
			pkgWoVer := divs[0]
			ver := ternaryF(len(divs) >= 2, func() string { return divs[1] }, func() string { return "latest" })
			gobinList.Map[pkgWoVer] = Gobin{
				Version:   ver,
				BuildOpts: opts,
			}
		}
	}
	// Add gobin itself@latest to the list if it is not listed.
	if key, _ := gobinList.Map.Find("gobin"); key == "" {
		gobinList.Map["github.com/knaka/gobin"] = Gobin{
			Version:   "latest",
			BuildOpts: []string{},
		}
	}
	gobinLock.Path = gobinLockPath
	gobinLock.Map = make(GobinMap)
	if stat, err := os.Stat(gobinLockPath); err == nil && !stat.IsDir() {
		reader := V(os.Open(gobinLockPath))
		defer (func() { V0(reader.Close()) })()
		body := string(V(io.ReadAll(reader)))
		if string(body) == "" {
			body = "{}"
		}
		V0(json.Unmarshal([]byte(body), &gobinLock.Map))
	}
	return
}

var reVerDiv = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`v[0-9]+`)
})

// resolveVersion resolves the version of a package. This function needs network connection.
func resolveVersion(pkg string, ver string) (resolvedVer string, err error) {
	// Early return if the version is already resolved (i.e., fully qualified).
	divs := strings.Split(pkg, ".")
	if ver != "latest" || len(divs) == 3 {
		resolvedVer = ver
		return
	}
	divs = strings.Split(pkg, "/")
	if len(divs) < 3 {
		return "", fmt.Errorf("invalid package name: %s", pkg)
	}
	var modules []string
	// Contains a major version suffix.
	if len(divs) >= 4 && reVerDiv().MatchString(divs[3]) {
		modules = append(modules, fmt.Sprintf("%s/%s/%s/%s", divs[0], divs[1], divs[2], divs[3]))
	}
	// Without it.
	modules = append(modules, fmt.Sprintf("%s/%s/%s", divs[0], divs[1], divs[2]))
	for _, module := range modules {
		cmd := exec.Command(
			V(getGoCmdPath()),
			"list", "-m", "--json", fmt.Sprintf("%s@%s", module, ver),
		)
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmd.Stderr = os.Stderr
		goListOutput := GoListOutput{}
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		V0(json.Unmarshal(output, &goListOutput))
		if strings.HasSuffix(goListOutput.Version, "+incompatible") {
			continue
		}
		return goListOutput.Version, nil
	}
	return "", fmt.Errorf("failed to resolve the version of %s", pkg)
}

func ensurePackageInstalled(gobinBinPath, pkgWoVer, resolvedVer string, buildOpts []string) (err error) {
	defer Catch(&err)
	cmd := path.Base(pkgWoVer)
	cmdPath := filepath.Join(gobinBinPath, cmd)
	cmdVer := fmt.Sprintf("%s@%s", cmd, resolvedVer)
	cmdVerPath := filepath.Join(gobinBinPath, fmt.Sprintf("%s@%s", cmd, resolvedVer))
	if stat, err := os.Stat(cmdVerPath); err == nil && !stat.IsDir() {
		V0(fmt.Fprintf(os.Stderr, "Skipping: %s@%s\n", pkgWoVer, resolvedVer))
	} else {
		args := []string{"install"}
		args = append(args, buildOpts...)
		args = append(args, fmt.Sprintf("%s@%s", pkgWoVer, resolvedVer))
		logger.Debugf("Installing: go %+v", args)
		installCmd := exec.Command(V(getGoCmdPath()), args...)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		installCmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinBinPath))
		V0(installCmd.Run())
		V0(os.Rename(cmdPath, cmdVerPath))
		V0(fmt.Fprintf(os.Stderr, "Installed: %s@%s\n", pkgWoVer, resolvedVer))
	}
	Ignore(os.Remove(cmdPath))
	V0(os.Symlink(cmdVer, cmdPath))
	return nil
}

// rePackage is a regular expression to check if a string is a Go package name.
var rePackage = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(
		// Are there a module with `{1,}` name?
		`[-a-zA-Z0-9@:%._+~#=]{2,256}\.[a-z]{2,6}(/[-a-zA-Z0-9:%_+.~#?&=]+){2,}@[-a-zA-Z0-9.]+`,
	)
})

// isPackage checks if a string is a package name.
func isPackage(s string) bool {
	return rePackage().MatchString(strings.TrimSpace(s))
}

type InstallExParams struct {
	Dir           string
	Env           []string
	WithGobinPath bool
}

type Opt func(params *InstallExParams) error

func WithDir(dir string) Opt {
	return func(params *InstallExParams) (err error) {
		params.Dir = dir
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

// RunWith installs a binary and runs it.
func RunWith(args []string, opts ...Opt) (err error) {
	return installEx(args, true, opts...)
}

func Run(args ...string) (err error) {
	return installEx(args, true)
}

// Install installs binaries.
func Install(cmds ...string) (err error) {
	defer Catch(&err)
	if len(cmds) == 0 {
		return Apply(cmds)
	}
	for _, cmd := range cmds {
		V0(installEx([]string{cmd}, false))
	}
	return
}

func installEx(args []string, shouldRun bool, opts ...Opt) (err error) {
	defer Catch(&err)

	funcParams := InstallExParams{
		WithGobinPath: true,
	}
	for _, opt := range opts {
		V0(opt(&funcParams))
	}

	// Install the binary which is listed in the gobin list file.

	gobinList, gobinLock, gobinBinPath := V3(getGobinList(V(os.Getwd())))
	for pkgWoVer, gobin := range gobinList.Map {
		pkgBase := path.Base(pkgWoVer)
		pkg := fmt.Sprintf("%s@%s", pkgWoVer, gobin.Version)
		logger.Debugf("pkgWoVer: %s, pkgBase: %s, pkg: %s, args[0]: %s", pkgWoVer, pkgBase, pkg, args[0])
		if !slices.Contains([]string{pkgWoVer, pkgBase, pkg}, args[0]) {
			continue
		}
		lockInfo, ok := gobinLock.Map[pkgWoVer]
		resolvedVer := TernaryF(ok,
			func() string { return lockInfo.Version },
			func() string { return V(resolveVersion(pkgWoVer, gobin.Version)) },
		)
		V0(ensurePackageInstalled(gobinBinPath, pkgWoVer, resolvedVer, gobin.BuildOpts))
		gobinLock.Map[pkgWoVer] = Gobin{
			Version:   resolvedVer,
			BuildOpts: gobin.BuildOpts,
		}
		V0(gobinLock.SaveJson())
		if !shouldRun {
			return nil
		}
		cmd := exec.Command(filepath.Join(gobinBinPath, pkgBase), args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = TernaryF(funcParams.Dir != "",
			func() string { return funcParams.Dir },
			func() string { return V(os.Getwd()) },
		)
		cmd.Env = append(os.Environ(), funcParams.Env...)
		if funcParams.WithGobinPath {
			// Prepend gobinBinPath to $PATH
			// Does this work on Windows?
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s%s%s", gobinBinPath, string(os.PathListSeparator), os.Getenv("PATH")))
		}
		V0(cmd.Run())
		return nil
	}

	// Install the binary which is not listed in the gobin list file.

	for i, arg := range args {
		if !isPackage(arg) {
			continue
		}
		pkg := arg
		buildOpts := args[:i]
		cmdOpts := args[i+1:]
		divs := strings.Split(pkg, "@")
		pkgWoVer := divs[0]
		ver := divs[1]
		resolvedVer := V(resolveVersion(pkgWoVer, ver))
		V0(ensurePackageInstalled(gobinBinPath, pkgWoVer, resolvedVer, buildOpts))
		if !shouldRun {
			return nil
		}
		cmd := exec.Command(filepath.Join(gobinBinPath, path.Base(pkgWoVer)), cmdOpts...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		V0(cmd.Run())
		return nil
	}

	// No matching command found.

	return fmt.Errorf("no matching command found")
}

// Apply installs all the binaries listed in a gobin list file.
func Apply(_ []string) (err error) {
	defer Catch(&err)
	gobinList, gobinLock, gobinBinPath := V3(getGobinList(V(os.Getwd())))
	for pkgWoVer, gobin := range gobinList.Map {
		var resolvedVer string
		var buildOpts []string
		if lockInfo, ok := gobinLock.Map[pkgWoVer]; ok {
			resolvedVer = lockInfo.Version
			buildOpts = lockInfo.BuildOpts
		} else {
			resolvedVer = V(resolveVersion(pkgWoVer, gobin.Version))
			buildOpts = gobin.BuildOpts
		}
		V0(ensurePackageInstalled(gobinBinPath, pkgWoVer, resolvedVer, buildOpts))
		gobinLock.Map[pkgWoVer] = Gobin{
			Version:   resolvedVer,
			BuildOpts: buildOpts,
		}
	}
	V0(gobinLock.SaveJson())
	return nil
}
