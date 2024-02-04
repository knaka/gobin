package lib

import (
	"math/rand"
	"os"
	"path/filepath"
)

// Average number of launches between cleanups
const cleanupCycle = 100

// Number of days after which a binary is considered old
const cleanupThresholdDays = 90

// Number of days after which a “@latest” version binary must be rebuilt
const latestVersionRebuildThresholdDays = 30

func Run(args []string) (err error) {
	goCmd, err := findGoCmd()
	if err != nil {
		return
	}
	goEnv, err := getGoEnv(goCmd)
	if err != nil {
		return
	}
	buildArgs, cmdArgs, err := splitArgs(goCmd, args)
	if err != nil {
		return
	}
	globalCacheDir, err := ensureGlobalCacheDir()
	if err != nil {
		return
	}

	// Clean up old binaries.
	if rand.Intn(cleanupCycle) == 0 {
		err = cleanupOldBinaries(globalCacheDir)
		if err != nil {
			return
		}
	}

	// Go file targets
	goFileInfoList, buildArgsWoTgt, err := getGoFileInfoList(buildArgs)
	if err != nil {
		return
	}

	// Target package
	pkg := (func() *string {
		if len(goFileInfoList) > 0 {
			return nil
		}
		elem := buildArgsWoTgt[len(buildArgsWoTgt)-1]
		if elem != "" && elem[0] != '.' && elem[0] != os.PathSeparator {
			return &buildArgsWoTgt[len(buildArgsWoTgt)-1]
		}
		return nil
	})()

	// Run
	if len(goFileInfoList) > 0 {
		buildInfo := newBuildInfoWithFiles(
			goEnv.Version,
			buildArgsWoTgt,
			goFileInfoList,
		)
		exeCacheDir := filepath.Join(globalCacheDir, buildInfo.Hash)
		err = goRunFiles(buildInfo, exeCacheDir, goCmd, buildArgs, cmdArgs)
	} else if pkg != nil && *pkg != "" {
		buildInfo := newBuildInfoWithPkg(
			goEnv.Version,
			buildArgsWoTgt,
			pkg,
		)
		exeCacheDir := filepath.Join(globalCacheDir, buildInfo.Hash)
		err = goRunPackage(buildInfo, exeCacheDir, goCmd, buildArgs, pkg, cmdArgs)
	} else {
		err = goRunNoCache(goCmd, buildArgs, cmdArgs)
	}
	if err != nil {
		return
	}

	return
}
