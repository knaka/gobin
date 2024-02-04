package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type BuildInfo struct {
	GoVersion string    `json:"go_version"`
	BuildArgs []string  `json:"build_args"`
	Package   *string   `json:"package,omitempty"`
	Files     []*GoFile `json:"files,omitempty"`
	Hash      string    `json:"hash"`
}

func newBuildInfoWithPkg(goVersion string, buildArgsWoTgt []string, pkg *string) BuildInfo {
	return newBuildInfo(goVersion, buildArgsWoTgt, pkg, nil)
}

func newBuildInfoWithFiles(goVersion string, buildArgsWoTgt []string, files []*GoFile) BuildInfo {
	return newBuildInfo(goVersion, buildArgsWoTgt, nil, files)
}

func newBuildInfo(goVersion string, buildArgsWoTgt []string, pkg *string, files []*GoFile) BuildInfo {
	hashAccu := sha1.New()
	hashAccu.Write([]byte(goVersion))
	for _, arg := range buildArgsWoTgt {
		hashAccu.Write([]byte(arg))
	}
	if pkg != nil {
		hashAccu.Write([]byte(*pkg))
	}
	for _, f := range files {
		hashAccu.Write([]byte(f.Hash))
	}
	return BuildInfo{
		GoVersion: goVersion,
		BuildArgs: buildArgsWoTgt,
		Package:   pkg,
		Files:     files,
		Hash:      hex.EncodeToString(hashAccu.Sum(nil)),
	}
}

func putBuildInfo(exeCacheDir string, buildInfo_ *BuildInfo) error {
	var err error
	infoPath := filepath.Join(exeCacheDir, "build_info.json")
	buildInfoJson, err := json.Marshal(buildInfo_)
	if err != nil {
		return err
	}
	err = os.WriteFile(infoPath, buildInfoJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func cleanupOldBinaries(dir string) error {
	subdirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}
		if stat, err := os.Stat(filepath.Join(dir, subdir.Name(), "main")); err == nil && !stat.IsDir() {
			if stat.ModTime().Before(stat.ModTime().Add(-cleanupThresholdDays * 24 * time.Hour)) {
				err = os.RemoveAll(filepath.Join(dir, subdir.Name()))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
