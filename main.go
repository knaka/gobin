package main

import (
	"fmt"
	. "github.com/knaka/go-utils"
	"github.com/knaka/gobin/lib"
	"log"
	"os"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, "Usage: %s <subcommand> [options]\n", os.Args[0]))
}

func main() {
	WaitForDebugger()
	var err error
	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}
	subCmd := os.Args[1]
	args := os.Args[2:]
	switch subCmd {
	case "apply":
		err = lib.Apply(args)
	case "install":
		err = lib.Install(args)
	case "run":
		err = lib.Run(args)
	default:
		help()
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
}
