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

func main() {
	var err error
	verbose := flag.Bool("v", false, "Verbose output.")
	silent := flag.Bool("s", false, "Silent output.")
	shouldHelp := flag.Bool("h", false, "Show help.")
	global := flag.Bool("g", false, "Install globally.")
	flag.Usage = func() {
		V0(fmt.Fprintln(os.Stderr, `Usage: gobin [options] <command> [<args>...]

Options:`))
		flag.PrintDefaults()
		V0(fmt.Fprintln(os.Stderr))
		V0(fmt.Fprintln(os.Stderr, `Commands:
  list                    List packages listed in the manifest file “Gobinfile”.
  run <name> [<args>...]  Run the specified program package.
  install [<name>...]     Install the specified package(s).
  update [<name>...]      Update the specified “@latest” program package(s). If no package is specified, update all packages.

Environment variables:
  NOSWITCH                If set, not switch to the locally installed (in “.gobin” directory) gobin command.`))
	}
	flag.Parse()
	vlog.SetVerbose(*verbose)
	log.SetSilent(*silent)
	if *shouldHelp {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		flag.Usage()
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
			fmt.Printf("%s@%s -> %s\n",
				entry.Pkg,
				entry.Version,
				Ternary(entry.LockedVersion == "" || entry.LockedVersion == "latest",
					"?",
					entry.LockedVersion,
				),
			)
		}
	case "help":
		flag.Usage()
		os.Exit(0)
	default:
		V0(fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subCmd))
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		stdlog.Fatalf("Error: %+v", err)
	}
	return
}
