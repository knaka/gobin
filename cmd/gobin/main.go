package main

import (
	"flag"
	"fmt"
	lib "github.com/knaka/gobin/lib"
	"github.com/knaka/gobin/log"
	stdlog "log"
	"os"
	"os/exec"

	. "github.com/knaka/go-utils"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, `gobin is a tool for managing Go binaries.

Usage: gobin [options] <apply|install|run|help> [args]
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
		cmd := V(lib.CommandEx(subArgs, lib.WithGlobal(*global)))
		err = cmd.Run()
		if err == nil {
			os.Exit(0)
		}
		errExit := ErrorAs[*exec.ExitError](err)
		if errExit != nil {
			os.Exit(errExit.ExitCode())
		}
		return err
	//case "install":
	//	return lib.Install(subArgs...)
	//case "apply":
	//	return lib.Apply(subArgs)
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
	Debugger()
	if err := Main(); err != nil {
		stdlog.Fatalf("Error: %+v", err)
	}
}
