package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	var err error
	var buildArgs []string
	var cmdArgs []string
	isBuildArg := true
	for _, arg := range os.Args[1:] {
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
	pkgNameVer := buildArgs[len(buildArgs)-1]
	fields := strings.Split(pkgNameVer, "@")
	pkgName := fields[0]
	pkgBase := path.Base(pkgName)
	pkgVer := fields[1]
	pkgBaseVer := pkgBase + "@" + pkgVer
	outDir := ".gobin"
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		panic(err)
	}
	binary := filepath.Join(outDir, pkgBase+"@"+pkgVer)
	link := filepath.Join(outDir, pkgBase)
	_, errOutPath := os.Stat(binary)
	_, errLinkPath := os.Stat(link)
	if errOutPath != nil || errLinkPath != nil {
		cmd := exec.Command("go", append([]string{"install"}, buildArgs...)...)
		absOutDir, err := filepath.Abs(outDir)
		if err != nil {
			panic(err)
		}
		cmd.Env = append(os.Environ(), "GOBIN="+absOutDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			panic(err)
		}
		err = os.Rename(link, binary)
		if err != nil {
			panic(err)
		}
		err = os.Symlink(pkgBaseVer, link)
		if err != nil {
			panic(err)
		}
	}
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
