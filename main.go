package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	. "github.com/knaka/go-utils"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, "Usage: %s <subcommand> [options]\n", os.Args[0]))
}

var fileBaseList = []string{
	"Gofile",
	"Gobinfile",
	".Gofile",
	".Gobinfile",
}

type Gobin struct {
	pkg     string
	ver     string
	tags    string
	comment string
}

type GoListOutput struct {
	Version string `json:"Version"`
}

func getGobinList() (gobinList []Gobin, err error) {
	defer Catch(&err)
	filePathList := lo.FilterMap(fileBaseList, func(fileBase string, _ int) (string, bool) {
		filePath := filepath.Join(V(os.UserHomeDir()), fileBase)
		stat, err := os.Stat(filePath)
		if err != nil {
			return "", false
		}
		if stat.IsDir() {
			return "", false
		}
		return filePath, true
	})
	sugar.Infof("filePathList: %+v", filePathList)
	if len(filePathList) == 0 {
		err = fmt.Errorf("no gobin list file found")
		return
	}
	filePath := filePathList[0]
	scanner := bufio.NewScanner(V(os.Open(filePath)))
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
		comment := TernaryF(len(divs) >= 2, func() string { return strings.TrimSpace(divs[1]) }, nil)
		divs = strings.SplitN(pkgVerTags, ",", 2)
		pkgVer := divs[0]
		tags := TernaryF(len(divs) >= 2, func() string { return divs[1] }, nil)
		divs = strings.SplitN(pkgVer, "@", 2)
		pkg := divs[0]
		ver := TernaryF(len(divs) >= 2, func() string { return divs[1] }, func() string { return "latest" })
		if ver == "latest" {
			divs = strings.Split(pkg, "/")
			module := fmt.Sprintf("%s/%s/%s", divs[0], divs[1], divs[2])
			divs = divs[3:]
			for {
				cmd := exec.Command("go", "list", "-m", "--json", fmt.Sprintf("%s@%s", module, ver))
				cmd.Env = append(os.Environ(), "GO111MODULE=on")
				goListOutput := GoListOutput{}
				V0(json.Unmarshal(V(cmd.Output()), &goListOutput))
				if !strings.HasSuffix(goListOutput.Version, "+incompatible") {
					ver = goListOutput.Version
					break
				}
				if len(divs) == 0 {
					return nil, fmt.Errorf("no version found for %s", pkg)
				}
				module = fmt.Sprintf("%s/%s", module, divs[0])
				divs = divs[1:]
			}
		}
		gobinList = append(gobinList, Gobin{
			pkg:     pkg,
			ver:     ver,
			tags:    tags,
			comment: comment})
	}
	sugar.Infof("gobinList: %+v", gobinList)
	return gobinList, nil
}

// GoEnv is a struct to hold the output of `go env -json`.
type GoEnv struct {
	Version string `json:"GOVERSION"`
	Gobin   string `json:"GOBIN"`
	Gopath  string `json:"GOPATH"`
}

var goVersion = sync.OnceValues(func() (goVersion string, err error) {
	defer Catch(&err)
	return V(goEnv()).Version, nil
})

var goBin = sync.OnceValues(func() (goBin string, err error) {
	defer Catch(&err)
	goBin = V(goEnv()).Gobin
	if goBin != "" {
		return goBin, nil
	}
	return filepath.Join(V(goEnv()).Gopath, "bin"), nil
})

var goCmd = sync.OnceValues(func() (goPath string, err error) {
	defer Catch(&err)
	goPath = filepath.Join(runtime.GOROOT(), "bin", "go")
	if stat, err := os.Stat(goPath); err == nil && !stat.IsDir() {
		return goPath, nil
	}
	goPath, err = exec.LookPath("go")
	if err == nil {
		return goPath, nil
	}
	return "", fmt.Errorf("go command not found")
})

var goEnv = sync.OnceValues(func() (goEnv_ GoEnv, err error) {
	defer Catch(&err)
	outStr := V(exec.Command(V(goCmd()), "env", "-json").Output())
	V0(json.Unmarshal(outStr, &goEnv_))
	return
})

func gobinApply(_ []string) (err error) {
	defer Catch(&err)
	gobinList := V(getGobinList())

	for _, gobin := range gobinList {
		if stat, err := os.Stat(filepath.Join(V(goBin()), fmt.Sprintf("%s@%s", path.Base(gobin.pkg), gobin.ver))); err == nil && !stat.IsDir() {
			sugar.Infof("Skipping %s@%s", gobin.pkg, gobin.ver)
			continue
		}
		args := []string{"install"}
		if gobin.tags != "" {
			args = append(args, "-tags", gobin.tags)
		}
		args = append(args, fmt.Sprintf("%s@%s", gobin.pkg, gobin.ver))
		//goCmd := V(goCmd())
		goCmd := "echo"
		cmd := exec.Command(goCmd, args...)
		cmd.Stdout = os.Stdout
		V0(cmd.Run())
	}
	return nil
}

var logger = zap.NewNop()
var sugar = logger.Sugar()

func main() {
	WaitForDebugger()
	if os.Getenv("VERBOSE") != "" {
		logger = V(zap.NewDevelopment())
		sugar = logger.Sugar()
	}
	var err error
	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}
	subCmd := os.Args[1]
	args := os.Args[2:]
	switch subCmd {
	case "apply":
		err = gobinApply(args)
	}
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
}
