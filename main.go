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
	V0(fmt.Fprintf(os.Stderr, `gobin is a tool for managing Go binaries.

Usage: gobin [options] <apply|install|run|help> [args]
`))
}

func Main() (err error) {
	defer Catch(&err)
	verbose := flag.Bool("verbose", false, "Verbose output")
	showsHelp := flag.Bool("help", false, "Show help")
	flag.Parse()
	if *verbose {
		lib.SetVerbose()
	}
	if *showsHelp {
		help()
		return nil
	}
	if flag.NArg() == 0 {
		help()
		os.Exit(1)
	}
	subCmd := flag.Arg(0)
	subArgs := flag.Args()[1:]
	switch subCmd {
	case "run":
		return lib.Run(subArgs)
	case "install":
		return lib.Install(subArgs...)
	case "apply":
		return lib.Apply(subArgs)
	case "help":
		help()
		return nil
	default:
		V0(fmt.Errorf("unknown subcommand: %s", subCmd))
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
