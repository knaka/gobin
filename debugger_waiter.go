package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// WaitForDebugger waits for a debugger to connect if the environment variable $WAIT or $DEBUG is set
//
// noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func WaitForDebugger() {
	if os.Getenv("DEBUG") == "" &&
		os.Getenv("WAIT") == "" {
		return
	}
	pid := os.Getpid()
	Assert(fmt.Fprintf(os.Stderr, "Process %d is waiting\n", pid))
outer:
	for {
		time.Sleep(1 * time.Second)
		lines := (func() string {
			command := exec.Command("ps", "w")
			stdout := Ensure(command.StdoutPipe())
			defer (func() { Ignore(stdout.Close()) })()
			Assert(command.Start())
			defer (func() { Assert(command.Wait()) })()
			return string(Ensure(io.ReadAll(stdout)))
		})()
		for _, line := range strings.Split(lines, "\n") {
			if strings.Contains(line, "dlv") &&
				strings.Contains(line, fmt.Sprintf("attach %d", pid)) {
				break outer
			}
		}
	}
	Assert(fmt.Fprintf(os.Stderr, "Debugger connected"))
	time.Sleep(1 * time.Second)
}
