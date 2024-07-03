package main

import (
	. "github.com/knaka/go-utils"
	"github.com/knaka/gobin/lib"
	"github.com/spf13/cobra"
	"log"
)

var cmdRoot = &cobra.Command{
	Use:   "gobin",
	Short: "Installs the binaries to `~/go/bin` or project-local directories according to the versions specified in the package list file.",
	Long:  "Installs the binaries to `~/go/bin` or project-local directories according to the versions specified in the package list file. It is useful for managing Go tools that are not installed globally in `$GOBIN` or `$GOPATH/bin`",
}

func Main() (err error) {
	defer Catch(&err)
	verbose := cmdRoot.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	subCmdRun := cobra.Command{
		Use:   "run",
		Short: "Run the specified binary",
		Long: `Run the specified binary.

The name of the binary to run is specified as an argument.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lib.Run(args, *verbose)
		},
	}
	cmdRoot.AddCommand(&subCmdRun)

	subCmdInstall := cobra.Command{
		Use:   "install",
		Short: "Install the specified binary, or all binaries if no argument is given",
		Long: `Install the specified binary.

If no argument is given, it installs all binaries specified in the package list file.`,
		RunE: func(_ *cobra.Command, args []string) error { return lib.Install(args, *verbose) },
	}
	cmdRoot.AddCommand(&subCmdInstall)

	subCmdApply := cobra.Command{
		Use:   "apply",
		Short: "Apply the package list file to the environment",
		Long: `Apply the package list file to the environment.

It installs the binaries specified in the package list file.`,
		RunE: func(_ *cobra.Command, args []string) error { return lib.Apply(args, *verbose) },
	}
	cmdRoot.AddCommand(&subCmdApply)

	V0(cmdRoot.Execute())
	return nil
}

func main() {
	WaitForDebugger()
	if err := Main(); err != nil {
		log.Fatalf("Error: %+v", err)
	}
}
