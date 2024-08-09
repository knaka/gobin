package main

import (
	"errors"
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
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var err error
	if os.Getenv("GOBIN_SILENT") != "" {
		log.SetSilent(true)
	}
	if os.Getenv("GOBIN_VERBOSE") != "" {
		vlog.SetVerbose(true)
	}
	if !filepath.IsAbs(os.Args[0]) {
		// If the command is called without an absolute path, search for the command in the $PATH.
		os.Args[0] = V(exec.LookPath(os.Args[0]))
	}
	cmdPath := filepath.Clean(V(filepath.Abs(os.Args[0])))
	_, globalGoBinPath := V2(minlib.GlobalConfDirPath())

	// Switch to the installed gobin command of the appropriate version.

	if os.Getenv("NOSWITCH") == "" {
		cmdGobinPath, err_ := minlib.EnsureGobinCmdInstalled(V(fsutils.IsSubDir(cmdPath, globalGoBinPath)))
		if err_ != nil {
			stdlog.Fatalf("Error 2c4804d: %+v", err)
		}
		if V(fsutils.CanonPath(V(os.Executable()))) != V(fsutils.CanonPath(cmdGobinPath)) {
			vlog.Printf("Switching to the installed gobin command: %s\n", cmdGobinPath)
			cmd := exec.Command(cmdGobinPath)
			// Save the original command path.
			cmd.Args = os.Args
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err_ = cmd.Run()
			if err_ == nil {
				os.Exit(0)
			}
			var errExec *exec.ExitError
			if errors.As(err_, &errExec) && errExec != nil {
				os.Exit(errExec.ExitCode())
			}
			stdlog.Fatalf("Error 7d70a88: %+v", err_)
		}
	}

	// If called as a symlink to the locally installed program, run the program of the appropriate version.

	cmdBase := filepath.Base(os.Args[0])
	if !(cmdBase == minlib.GobinCmdBase || strings.HasPrefix(cmdBase, minlib.GobinCmdBase+"@")) {
		if filepath.Base(filepath.Dir(cmdPath)) == minlib.GobinDirBase ||
			V(fsutils.IsSubDir(filepath.Dir(cmdPath), globalGoBinPath)) {
			opts := []gobin.Option{}
			if V(fsutils.IsSubDir(cmdPath, globalGoBinPath)) {
				opts = append(opts, gobin.Global(true))
			}
			targetCmdPath, err_ := gobin.InstallEx([]string{cmdBase}, opts...)
			vlog.Printf("Switching to the installed command: %s\n", targetCmdPath)
			if err_ != nil {
				stdlog.Fatalf("Error 078a110: %+v", err_)
			}
			cmd, err_ := minlib.Command(targetCmdPath)
			cmd.Args = os.Args
			err_ = cmd.Run()
			if err_ == nil {
				os.Exit(0)
			}
			var execErr *exec.ExitError
			if errors.As(err, &execErr) && execErr != nil {
				os.Exit(execErr.ExitCode())
			}
			stdlog.Fatalf("Error 608a109: %+v", err_)
		}
	}

	// Continue as gobin command.

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
	if !filepath.IsAbs(os.Args[0]) {
		os.Args[0] = V(exec.LookPath(os.Args[0]))
	}
	vlog.SetVerbose(*verbose)
	log.SetSilent(*silent)
	if os.Getenv("GOBIN_SILENT") != "" {
		log.SetSilent(true)
	}
	if os.Getenv("GOBIN_VERBOSE") != "" {
		vlog.SetVerbose(true)
	}
	if *shouldHelp {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	subCmd := flag.Arg(0)
	subArgs := flag.Args()[1:]
	switch subCmd {
	case "run":
		_, err = gobin.RunEx(subArgs,
			gobin.WithStdin(os.Stdin),
			gobin.WithStdout(os.Stdout),
			gobin.WithStderr(os.Stderr),
			gobin.Global(*global),
		)
	case "install":
		_, err = gobin.InstallEx(subArgs,
			gobin.Global(*global),
		)
	case "update":
		err = gobin.UpdateEx(subArgs,
			gobin.Global(*global),
		)
	case "list":
		l, err_ := gobin.List(*global)
		if err_ != nil {
			stdlog.Fatalf("Error 6eaee66: %+v", err_)
		}
		for _, entry := range l {
			fmt.Printf("%s@%s -> %s\n",
				entry.Pkg,
				entry.Version,
				Ternary(entry.LockedVersion == "latest",
					"undefined",
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
		stdlog.Fatalf("Error beba31d: %+v", err)
	}
	return
}
