package main

import (
	"github.com/knaka/gobin/lib"
	"os"
)

func main() {
	var buildArgs []string // Arguments for `go install ...`.
	var cmdArgs []string   // Arguments for the binary.
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
	err := lib.Run(buildArgs, cmdArgs)
	if err != nil {
		panic(err)
	}
}
