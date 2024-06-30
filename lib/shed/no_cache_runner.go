package lib

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// goRunNoCache builds and runs the binary in a temporary directory to enable debug information.
func goRunNoCache(
	goCmd string,
	buildArgs []string,
	cmdArgs []string,
) (err error) {
	log.Println("Not caching.")
	tmpDir, err := os.MkdirTemp("", "go-run-cache")
	if err != nil {
		return err
	}
	defer (func() { _ = os.RemoveAll(tmpDir) })()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			_ = os.RemoveAll(tmpDir)
		}
	}()
	mainPath := filepath.Join(tmpDir, "main")
	cmd := exec.Command(goCmd, append([]string{"build", "-o", mainPath}, buildArgs...)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	cmd = exec.Command(mainPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}
