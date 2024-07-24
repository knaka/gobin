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
	V0(fmt.Fprintf(os.Stderr, "Process %d is waiting\n", pid))
	for {
		time.Sleep(duration)
		if debuggerProcessExists(pid) {
			break
		}
	}
	V0(fmt.Fprintf(os.Stderr, "Debugger connected"))
	time.Sleep(duration)
}

var WaitForDebugger = Debugger

// This function can be platform specific.
func debuggerProcessExists(pid int) (exists bool) {
	cmd := exec.Command("ps", "w")
	cmdOut := V(cmd.StdoutPipe())
	defer (func() { Ignore(cmdOut.Close()) })()
	scanner := bufio.NewScanner(cmdOut)
	V0(cmd.Start()) // Start() does not wait while Run() does
	defer (func() { V0(cmd.Wait()) })()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "dlv") &&
			strings.Contains(line, fmt.Sprintf("attach %d", pid)) {
			return true
		}
	}
	return false
}
