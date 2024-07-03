package main

import (
	"flag"
	"fmt"
	"github.com/knaka/gobin/lib"
	"log"
	"os"

	. "github.com/knaka/go-utils"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, "gobin is a tool for managing Go binaries.\n"))
}

func Main() (err error) {
	defer Catch(&err)
	verbose := flag.Bool("verbose", false, "Verbose output")
	showHelp := flag.Bool("help", false, "Show help")
	flag.Parse()
	if *verbose {
		lib.SetVerbose()
	}
	if *showHelp {
		help()
		return nil
	}
	switch flag.Arg(0) {
	case "run":
		return lib.Run(flag.Args()[1:])
	case "install":
		return lib.Install(flag.Args()[1:])
	case "apply":
		return lib.Apply(flag.Args()[1:])
	case "help":
		help()
		return nil
	default:
		V0(fmt.Errorf("unknown subcommand: %s", flag.Arg(0)))
		help()
		os.Exit(1)
	}
	return
}

func main() {
	WaitForDebugger()
	if err := Main(); err != nil {
		log.Fatalf("Error: %+v", err)
	}
}
