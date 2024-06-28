package main

import (
	"bufio"
	"fmt"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
	"strings"

	. "github.com/knaka/go-utils"
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
	tags    string
	comment string
}

func gobinApply(args []string) (err error) {
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
	sugar := logger.Sugar()
	sugar.Infof("filePathList: %+v", filePathList)
	if len(filePathList) == 0 {
		err = fmt.Errorf("no gobin list file found")
		return
	}
	filePath := filePathList[0]
	var gobinList []Gobin
	// Reead file line by line
	scanner := bufio.NewScanner(V(os.Open(filePath)))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		divs := strings.SplitN(line, "#", 2)
		pkg := strings.TrimSpace(divs[0])
		if pkg == "" {
			continue
		}
		comment := TernaryF(len(divs) >= 2, func() string { return strings.TrimSpace(divs[1]) }, nil)
		gobinList = append(gobinList, Gobin{pkg: pkg, comment: comment})
	}
	sugar.Infof("gobinList: %+v", gobinList)
	return
}

var logger = zap.NewNop()

func main() {
	WaitForDebugger()
	if os.Getenv("VERBOSE") != "" {
		logger = V(zap.NewDevelopment())
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
