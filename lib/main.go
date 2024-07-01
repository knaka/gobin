package lib

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
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

type Gobin struct {
	pkgWoVer  string
	ver       string
	buildOpts []string
	comment   string
}

type GoListOutput struct {
	Version string `json:"Version"`
}

var logger = V(zap.NewDevelopment())
var sugar = logger.Sugar()

var goVersion = sync.OnceValues(func() (goVersion string, err error) {
	defer Catch(&err)
	return V(getGoEnv()).Version, nil
})

func resolveLatestVersion(pkg string, ver string) (resolvedVer string, err error) {
	if ver != "latest" {
		return ver, nil
	}
	divs := strings.Split(pkg, "/")
	module := fmt.Sprintf("%s/%s/%s", divs[0], divs[1], divs[2])
	divs = divs[3:]
	for {
		cmd := exec.Command(V(getGoCmdPath()), "list", "-m", "--json", fmt.Sprintf("%s@%s", module, ver))
		cmd.Env = append(os.Environ(), "GO111MODULE=on")
		goListOutput := GoListOutput{}
		V0(json.Unmarshal(V(cmd.Output()), &goListOutput))
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

func ensurePackageInstalled(gobinPath, pkgWoVer, ver string, buildOpts []string) (err error) {
	defer Catch(&err)
	name := path.Base(pkgWoVer)
	resolvedVer := V(resolveLatestVersion(pkgWoVer, ver))
	namePath := filepath.Join(gobinPath, name)
	nameVer := fmt.Sprintf("%s@%s", name, resolvedVer)
	baseVerPath := filepath.Join(gobinPath, fmt.Sprintf("%s@%s", name, resolvedVer))
	if stat, err := os.Stat(baseVerPath); err == nil && !stat.IsDir() {
		sugar.Infof("Skipping %s@%s", pkgWoVer, resolvedVer)
	} else {
		args := []string{"install"}
		args = append(args, buildOpts...)
		args = append(args, fmt.Sprintf("%s@%s", pkgWoVer, resolvedVer))
		sugar.Infof("go %+v", args)
		cmd := exec.Command(V(getGoCmdPath()), args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
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
func Run(args []string) (err error) {
	return installEx(args, true)
}

// Install installs a binary.
func Install(args []string) (err error) {
	return installEx(args, false)
}

func installEx(args []string, shouldRun bool) (err error) {
	defer Catch(&err)
	gobinList, gobinPath := V2(getGobinList("."))
	for _, gobin := range gobinList {
		pkgWoVer := gobin.pkgWoVer
		pkgBase := path.Base(pkgWoVer)
		pkg := fmt.Sprintf("%s@%s", pkgWoVer, gobin.ver)
		if !slices.Contains([]string{pkgWoVer, pkgBase, pkg}, args[0]) {
			continue
		}
		resolvedVer := V(resolveLatestVersion(pkgWoVer, gobin.ver))
		V0(ensurePackageInstalled(gobinPath, pkgWoVer, resolvedVer, gobin.buildOpts))
		if !shouldRun {
			return nil
		}
		cmd := exec.Command(filepath.Join(gobinPath, pkgBase), args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		V0(cmd.Run())
		return nil
	}
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
		resolvedVer := V(resolveLatestVersion(pkgWoVer, ver))
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
	return fmt.Errorf("no matching command found")
}

// Apply installs all the binaries listed in a gobin list file.
func Apply(_ []string) (err error) {
	defer Catch(&err)
	gobinList, gobinPath := V2(getGobinList("."))
	for _, gobin := range gobinList {
		resolvedVer := V(resolveLatestVersion(gobin.pkgWoVer, gobin.ver))
		V0(ensurePackageInstalled(gobinPath, gobin.pkgWoVer, resolvedVer, gobin.buildOpts))
	}
	return nil
}
