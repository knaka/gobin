package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const duration = 1 * time.Second

// Debugger waits for a debugger to connect if the environment variable $WAIT or $DEBUG is set
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func Debugger() {
	if os.Getenv("DEBUG") == "" &&
		os.Getenv("WAIT") == "" {
		return
	}
	pid := os.Getpid()
	_ = V(fmt.Fprintf(os.Stderr, "Process %d is waiting\n", pid))
outer:
	for {
		time.Sleep(duration)
		if debuggerProcessExists(pid) {
			break outer
		}
	}
	_ = V(fmt.Fprintf(os.Stderr, "Debugger connected"))
	time.Sleep(duration)
}

var WaitForDebugger = Debugger

// This function can be platform specific.
func debuggerProcessExists(pid int) (b bool) {
	cmd := exec.Command("ps", "w")
	stdout := V(cmd.StdoutPipe())
	defer (func() { Ignore(stdout.Close()) })()
	scanner := bufio.NewScanner(stdout)
	V0(cmd.Start()) // Start() does not wait while Run() does
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "dlv") &&
			strings.Contains(line, fmt.Sprintf("attach %d", pid)) {
			b = true
			break
		}
	}
	V0(cmd.Wait())
	return
}
