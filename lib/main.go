package lib

import (
	"errors"
	"github.com/knaka/go-utils"
	"golang.org/x/mod/modfile"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func findGoMod() (string, error) {
	modDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		_, err = os.Stat(filepath.Join(modDir, "go.mod"))
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		err = nil
		parent := filepath.Dir(modDir)
		if parent == modDir {
			return "", errors.New("go.mod not found")
		}
		modDir = parent
	}
	return filepath.Join(modDir, "go.mod"), nil
}

func Run(args ...string) (err error) {
	utils.WaitForDebugger()
	var buildArgs []string // Arguments for `go install ...`.
	var cmdArgs []string   // Arguments for the binary.
	isBuildArg := true
	for _, arg := range args {
		if isBuildArg && arg == "--" {
			isBuildArg = false
			continue
		}
		if isBuildArg {
			buildArgs = append(buildArgs, arg)
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}
	if len(buildArgs) == 0 {
		return errors.New("package name must be specified")
	}
	pkg := buildArgs[len(buildArgs)-1]

	outDir := ".gobin"
	goMod, err := findGoMod()
	if err == nil {
		outDir = filepath.Join(filepath.Dir(goMod), ".gobin")
	} else {
		goMod = ""
		outDir = ".gobin"
	}

	pkgFields := strings.Split(pkg, "@")
	if len(pkgFields) < 2 {
		if goMod == "" {
			return errors.New("package name must be in the form of pkg@ver if not in module-aware mode")
		}
		mainPkgName := pkg
		modName := pkg
		reader, errSub := os.Open(goMod)
		if errSub != nil {
			return errSub
		}
		body, errSub := io.ReadAll(reader)
		if errSub != nil {
			return errSub
		}
		f, errSub := modfile.Parse(goMod, body, nil)
		if errSub != nil {
			return errSub
		}
	outer:
		for {
			for _, r := range f.Require {
				if r.Mod.Path == modName {
					pkgFields = []string{mainPkgName, r.Mod.Version}
					break outer
				}
			}
			parent := path.Dir(modName)
			if parent == modName {
				return errors.New("package not found in go.mod")
			}
			modName = parent
		}
	}
	mainPkgName := pkgFields[0]
	pkgBaseName := path.Base(mainPkgName)
	pkgVer := pkgFields[1]
	pkgBaseVer := pkgBaseName + "@" + pkgVer
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return
	}
	binary := filepath.Join(outDir, pkgBaseName+"@"+pkgVer)
	link := filepath.Join(outDir, pkgBaseName)
	_, errOutPath := os.Stat(binary)
	_, errLinkPath := os.Stat(link)
	if errOutPath != nil || errLinkPath != nil {
		cmd := exec.Command("go", append([]string{"install"}, buildArgs...)...)
		absOutDir, errSub := filepath.Abs(outDir)
		if errSub != nil {
			return errSub
		}
		cmd.Env = append(os.Environ(), "GOBIN="+absOutDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		errSub = cmd.Run()
		if errSub != nil {
			return errSub
		}
		errSub = os.Rename(link, binary)
		if errSub != nil {
			return errSub
		}
		errSub = os.Symlink(pkgBaseVer, link)
		if errSub != nil {
			return errSub
		}
	}
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}
