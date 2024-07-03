package lib

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mattn/go-shellwords"
	"go.uber.org/zap"
	"io"
	"log"
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

var logger *zap.Logger
var sugar *zap.SugaredLogger

var reBuildOptsSeparator = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`\s+`)
})

func getGobinList(dirPath string) (
	gobinList GobinList,
	gobinLock GobinList,
	gobinBinPath string,
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
			pkgVerTags := strings.TrimSpace(divs[0])
			if pkgVerTags == "" {
				continue
			}
			//comment := ternaryF(len(divs) >= 2, func() string { return strings.TrimSpace(divs[1]) }, nil)
			//divs = strings.SplitN(pkgVerTags, " ", 2)
			divs = reBuildOptsSeparator().Split(pkgVerTags, 2)
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
			body = "{}\n"
		}
		V0(json.Unmarshal([]byte(body), &gobinLock.Map))
	}
	return
}

// resolveVersion resolves the version of a package. This function needs network connection.
func resolveVersion(pkg string, ver string) (resolvedVer string, err error) {
	divs := strings.Split(pkg, ".")
	if ver != "latest" || len(divs) == 3 {
		resolvedVer = ver
		return
	}
	divs = strings.Split(pkg, "/")
	module := fmt.Sprintf("%s/%s/%s", divs[0], divs[1], divs[2])
	divs = divs[3:]
	for {
		cmd := exec.Command(V(getGoCmdPath()), "list", "-m", "--json", fmt.Sprintf("%s@%s", module, ver))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		cmd.Stderr = os.Stderr
		goListOutput := GoListOutput{}
		V0(json.Unmarshal(V(cmd.Output()), &goListOutput))
		sugar.Debugf("6a45931: %s", goListOutput.Version)
		if !strings.HasSuffix(goListOutput.Version, "+incompatible") {
			ver = goListOutput.Version
			break
		}
		if len(divs) == 0 {
			return "", fmt.Errorf("no version found for %s", pkg)
		}
		module = fmt.Sprintf("%s/%s", module, divs[0])
		divs = divs[1:]
	}
	return ver, nil
}

func ensurePackageInstalled(gobinBinPath, pkgWoVer, resolvedVer string, buildOpts []string) (err error) {
	defer Catch(&err)
	name := path.Base(pkgWoVer)
	namePath := filepath.Join(gobinBinPath, name)
	nameVer := fmt.Sprintf("%s@%s", name, resolvedVer)
	baseVerPath := filepath.Join(gobinBinPath, fmt.Sprintf("%s@%s", name, resolvedVer))
	if stat, err := os.Stat(baseVerPath); err == nil && !stat.IsDir() {
		sugar.Debugf("Skipping %s@%s", pkgWoVer, resolvedVer)
	} else {
		args := []string{"install"}
		args = append(args, buildOpts...)
		args = append(args, fmt.Sprintf("%s@%s", pkgWoVer, resolvedVer))
		sugar.Infof("go %+v", args)
		cmd := exec.Command(V(getGoCmdPath()), args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobinBinPath))
		V0(cmd.Run())
		V0(os.Rename(namePath, baseVerPath))
	}
	Ignore(os.Remove(namePath))
	V0(os.Symlink(nameVer, namePath))
	return nil
}

var rePackage = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(
		// Are there a module with `{1,}` name?
		`[-a-zA-Z0-9@:%._+~#=]{2,256}\.[a-z]{2,6}(/[-a-zA-Z0-9:%_+.~#?&=]+){2,}@[-a-zA-Z0-9.]+`,
	)
})

// isPackage checks if a string is a package name.
func isPackage(s string) bool {
	s = strings.TrimSpace(s)
	return rePackage().MatchString(s)
}

// Run installs a binary and runs it
func Run(args []string, verbose bool) (err error) {
	return installEx(args, true, verbose)
}

// Install installs a binary.
func Install(args []string, verbose bool) (err error) {
	if len(args) == 0 {
		return Apply(args, verbose)
	}
	return installEx(args, false, verbose)
}

func installEx(args []string, shouldRun bool, verbose bool) (err error) {
	defer Catch(&err)

	if verbose {
		logger = V(zap.NewDevelopment())
		sugar = logger.Sugar()
		sugar.Debugf("Verbose mode")
	} else {
		logger = V(zap.NewProduction())
		sugar = logger.Sugar()
		sugar.Debugf("Production mode")
	}

	// Install the binary which is listed in the gobin list file.

	gobinList, gobinLock, gobinPath := V3(getGobinList(V(os.Getwd())))
	for pkgWoVer, gobin := range gobinList.Map {
		pkgBase := path.Base(pkgWoVer)
		pkg := fmt.Sprintf("%s@%s", pkgWoVer, gobin.Version)
		if !slices.Contains([]string{pkgWoVer, pkgBase, pkg}, args[0]) {
			continue
		}
		lockInfo, ok := gobinLock.Map[pkgWoVer]
		resolvedVer := TernaryF(ok,
			func() string { return lockInfo.Version },
			func() string { return V(resolveVersion(pkgWoVer, gobin.Version)) },
		)
		V0(ensurePackageInstalled(gobinPath, pkgWoVer, resolvedVer, gobin.BuildOpts))
		gobinLock.Map[pkgWoVer] = Gobin{
			Version:   resolvedVer,
			BuildOpts: gobin.BuildOpts,
		}
		V0(gobinLock.SaveJson())
		if !shouldRun {
			return nil
		}
		cmd := exec.Command(filepath.Join(gobinPath, pkgBase), args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		V0(cmd.Run())
		return nil
	}

	log.Println("Refer %v", gobinLock)

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
		V0(ensurePackageInstalled(gobinPath, pkgWoVer, resolvedVer, buildOpts))
		if !shouldRun {
			return nil
		}
		cmd := exec.Command(filepath.Join(gobinPath, path.Base(pkgWoVer)), cmdOpts...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		V0(cmd.Run())
		return nil
	}

	// No matching command found.

	return fmt.Errorf("no matching command found")
}

// Apply installs all the binaries listed in a gobin list file.
func Apply(_ []string, verbose bool) (err error) {
	defer Catch(&err)

	if verbose {
		logger = V(zap.NewDevelopment())
		sugar = logger.Sugar()
	} else {
		logger = V(zap.NewProduction())
		sugar = logger.Sugar()
	}

	gobinList, gobinLock, gobinPath := V3(getGobinList(V(os.Getwd())))
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
		V0(ensurePackageInstalled(gobinPath, pkgWoVer, resolvedVer, buildOpts))
		gobinLock.Map[pkgWoVer] = Gobin{
			Version:   resolvedVer,
			BuildOpts: buildOpts,
		}
	}
	V0(gobinLock.SaveJson())
	return nil
}
