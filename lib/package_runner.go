package lib

import (
	"errors"
	"golang.org/x/mod/modfile"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func goRunPackage(
	buildInfo_ BuildInfo,
	exeCacheDir string,
	goCmd string,
	buildArgs []string,
	pkg *string,
	cmdArgs []string,
) (err error) {
	err = os.MkdirAll(exeCacheDir, 0755)
	if err != nil {
		return
	}
	pkgFields := strings.Split(*pkg, "@")
	if len(pkgFields) < 2 {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		goMod, err := findGoModFile(wd)
		if err != nil {
			return errors.New("package name must be in the form of pkg@ver if not in module-aware mode")
		}
		reader, err := os.Open(goMod)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		f, err := modfile.Parse(goMod, body, nil)
		if err != nil {
			return err
		}
		mainPkgName := *pkg
		modName := *pkg
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
		moduleCacheDir, err := getModuleCacheDir()
		if err != nil {
			return err
		}
		// Clean old binaries sometimes
		if rand.Intn(cleanupCycle) == 0 {
			err = cleanupOldBinaries(moduleCacheDir)
			if err != nil {
				return err
			}
		}
		exeCacheDir = filepath.Join(moduleCacheDir, buildInfo_.Hash)
	}
	pkgVer := pkgFields[1]
	err = os.MkdirAll(exeCacheDir, 0755)
	if err != nil {
		return
	}
	mainPath := filepath.Join(exeCacheDir, "main")
	stat, err := os.Stat(mainPath)
	needBuild := err != nil || stat.IsDir()
	if !needBuild &&
		pkgVer == "latest" &&
		stat.ModTime().Before(stat.ModTime().Add(-latestVersionRebuildThresholdDays*24*time.Hour)) {
		needBuild = true
	}
	if needBuild {
		cmd := exec.Command(goCmd, append([]string{"install"}, buildArgs...)...)
		cmd.Env = append(os.Environ(), "GOBIN="+exeCacheDir)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		pkgName := pkgFields[0]
		pkgBase := path.Base(pkgName)
		err = os.Rename(filepath.Join(exeCacheDir, pkgBase), mainPath)
		if err != nil {
			return err
		}
		err = putBuildInfo(exeCacheDir, &buildInfo_)
		if err != nil {
			return err
		}
	}
	cmd := exec.Command(mainPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}
