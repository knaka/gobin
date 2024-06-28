package lib

import (
	"os"
	"path/filepath"
)

// ensureGlobalCacheDir returns the global cache directory, creating it if it doesn't exist.
func ensureGlobalCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	globalCacheDir := filepath.Join(userCacheDir, "go-run-cache")
	err = os.MkdirAll(globalCacheDir, 0700)
	if err != nil {
		return "", err
	}
	return globalCacheDir, nil
}

// getModuleCacheDir returns the cache directory for the current module.
func getModuleCacheDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	goMod, err := findGoModFile(wd)
	if err != nil {
		return "", err
	}
	modDir := filepath.Dir(goMod)
	moduleCacheDir := filepath.Join(modDir, ".go-run-cache")
	err = os.MkdirAll(moduleCacheDir, 0700)
	if err != nil {
		return "", err
	}
	return moduleCacheDir, nil
}
