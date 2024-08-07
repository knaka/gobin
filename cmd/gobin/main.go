package main

import (
	"flag"
	"fmt"
	. "github.com/knaka/go-utils"
	"github.com/knaka/gobin"
	"github.com/knaka/gobin/log"
	stdlog "log"
	"os"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, `gobin is a tool for managing Go binaries.

Usage: gobin [options] <install|run|help> [args]
`))
}

func Main() (err error) {
	defer Catch(&err)
	verbose := flag.Bool("v", false, "Verbose output")
	shouldHelp := flag.Bool("h", false, "Show help")
	global := flag.Bool("g", false, "Install globally")
	flag.Parse()
	if *verbose {
		log.SetOutput(os.Stderr)
	}
	if *shouldHelp {
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
		Debugger()
		return gobin.RunEx(subArgs,
			gobin.WithStdin(os.Stdin),
			gobin.WithStdout(os.Stdout),
			gobin.WithStderr(os.Stderr),
			gobin.Global(*global),
			gobin.Verbose(*verbose),
		)
	case "install":
		return gobin.InstallEx(subArgs,
			gobin.Global(*global),
			gobin.Verbose(*verbose),
		)
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
	if err := Main(); err != nil {
		stdlog.Fatalf("Error: %+v", err)
	}
}
