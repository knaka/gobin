package lib

import (
	"os"
	"os/exec"
	"path/filepath"
)

func goRunFiles(
	buildInfo_ BuildInfo,
	exeCacheDir string,
	goCmd string,
	buildArgs []string,
	cmdArgs []string,
) (err error) {
	err = os.MkdirAll(exeCacheDir, 0755)
	if err != nil {
		return
	}
	mainPath := filepath.Join(exeCacheDir, "main")
	if stat, err := os.Stat(mainPath); err != nil || stat.IsDir() {
		cmd := exec.Command(goCmd, append([]string{"build", "-o", mainPath}, buildArgs...)...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		err = putBuildInfo(exeCacheDir, &buildInfo_)
		if err != nil {
			return err
		}
	}
	cmd := exec.Command(mainPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}
