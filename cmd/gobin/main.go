package main

import (
	"flag"
	"fmt"
	. "github.com/knaka/go-utils"
	fsutils "github.com/knaka/go-utils/fs"
	"github.com/knaka/gobin"
	"github.com/knaka/gobin/log"
	"github.com/knaka/gobin/minlib"
	"github.com/knaka/gobin/vlog"
	stdlog "log"
	"os"
)

func help() {
	V0(fmt.Fprintf(os.Stderr, `gobin is a tool for managing Go binaries.

Usage: gobin [options] <install|run|update|help> [args]
`))
}

func main() {
	var err error
	verbose := flag.Bool("v", false, "Verbose output")
	silent := flag.Bool("s", false, "Silent output")
	shouldHelp := flag.Bool("h", false, "Show help")
	global := flag.Bool("g", false, "Install globally")
	flag.Parse()
	vlog.SetVerbose(*verbose)
	log.SetSilent(*silent)
	if *shouldHelp {
		help()
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		help()
		os.Exit(1)
	}
	if os.Getenv("NOSWITCH") == "" {
		// Switch to the locally installed gobin command.
		cmdPath, err_ := minlib.EnsureGobinCmdInstalled()
		if err_ != nil {
			stdlog.Fatalf("Error: %+v", err)
		}
		if V(fsutils.CanonPath(V(os.Executable()))) != V(fsutils.CanonPath(cmdPath)) {
			vlog.Printf("Switching to the locally installed gobin command: %s\n", cmdPath)
			errExec, err_ := minlib.RunCommand(cmdPath, os.Args[1:]...)
			if err_ == nil {
				os.Exit(0)
			}
			if errExec != nil {
				os.Exit(errExec.ExitCode())
			}
			stdlog.Fatalf("Error: %+v", err_)
		}
	}
	subCmd := flag.Arg(0)
	subArgs := flag.Args()[1:]
	switch subCmd {
	case "run":
		err = gobin.RunEx(subArgs,
			gobin.WithStdin(os.Stdin),
			gobin.WithStdout(os.Stdout),
			gobin.WithStderr(os.Stderr),
			gobin.Global(*global),
		)
	case "install":
		err = gobin.InstallEx(subArgs,
			gobin.Global(*global),
		)
	case "update":
		err = gobin.UpdateEx(subArgs,
			gobin.Global(*global),
		)
	case "list":
		l, err_ := gobin.List()
		if err_ != nil {
			stdlog.Fatalf("Error: %+v", err_)
		}
		for _, entry := range l {
			fmt.Printf("%s@%s -> %s\n", entry.Pkg, entry.Version, entry.LockedVersion)
		}
	case "help":
		help()
		os.Exit(0)
	default:
		V0(fmt.Errorf("unknown subcommand: %s", subCmd))
		help()
		os.Exit(1)
	}
	if err != nil {
		stdlog.Fatalf("Error: %+v", err)
	}
	return
}
