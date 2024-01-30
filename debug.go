package utils

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
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func WaitForDebugger() {
	if os.Getenv("DEBUG") == "" &&
		os.Getenv("WAIT") == "" {
		return
	}
	pid := os.Getpid()
	_, _ = fmt.Fprintf(os.Stderr, "Process %d is waiting\n", pid)
outer:
	for {
		time.Sleep(1 * time.Second)
		lines := (func() string {
			var err error
			command := exec.Command("ps", "w")
			stdout, err := command.StdoutPipe()
			defer Let0(err, func() { _ = stdout.Close() })
			err = Then(err, func() error { return command.Start() })
			defer Let0(err, func() { _ = command.Wait() })
			data, err := Bind(err, func() ([]byte, error) { return io.ReadAll(stdout) })
			if err != nil {
				panic(err)
			}
			return string(data)
		})()
		for _, line := range strings.Split(lines, "\n") {
			if strings.Contains(line, "dlv") &&
				strings.Contains(line, fmt.Sprintf("attach %d", pid)) {
				break outer
			}
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "Debugger connected")
	time.Sleep(1 * time.Second)
}
