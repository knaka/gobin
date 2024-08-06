package gobin

import (
	"bufio"
	. "github.com/knaka/go-utils"
	"github.com/knaka/gobin/minlib"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type maniEntry struct {
	Pkg      string
	Version  string `json:"version"`
	Tags     string
	Requires []string
}

// manifestT is the internal representation of the manifest and the manifest lock file.
type manifestT struct {
	filePath  string
	entries   []*maniEntry
	lockPath  string
	pkgMapVer minlib.PkgVerLockMapT
}

const maniBase = "Gobinfile"
const maniLockBase = "Gobinfile-lock"

var reSpaces = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`\s+`) })

func parseManifest(dirPath string) (gobinManifest *manifestT, err error) {
	defer Catch(&err)
	gobinManifest = &manifestT{
		filePath: filepath.Join(dirPath, maniBase),
		lockPath: filepath.Join(dirPath, maniLockBase),
	}
	if _, err_ := os.Stat(gobinManifest.filePath); err_ == nil {
		reader := V(os.Open(gobinManifest.filePath))
		defer (func() { V0(reader.Close()) })()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			divs := strings.SplitN(line, "#", 2)
			line = strings.TrimSpace(divs[0])
			divs = reSpaces().Split(line, 2)
			pkgVer := divs[0]
			optsStr := TernaryF(len(divs) >= 2,
				func() string { return divs[1] },
				func() string { return "" },
			)
			var requires []string
			var tags string
			if optsStr != "" {
				divs = reSpaces().Split(optsStr, -1)
				for _, opt := range divs {
					x := strings.SplitN(opt, "=", 2)
					if len(x) < 2 {
						continue
					}
					key := x[0]
					val := x[1]
					switch key {
					case "requires":
						reqs := strings.Split(val, ",")
						for _, req := range reqs {
							requires = append(requires, req)
						}
					case "tags":
						tags = val
					}
				}
			}
			divs = strings.SplitN(pkgVer, "@", 2)
			pkg := divs[0]
			ver := TernaryF(len(divs) >= 2,
				func() string { return Ternary(divs[1] == "latest", "", divs[1]) },
				func() string { return "" },
			)
			gobinManifest.entries = append(gobinManifest.entries, &maniEntry{
				Pkg:      pkg,
				Version:  ver,
				Tags:     tags,
				Requires: requires,
			})
		}
	}
	if _, err_ := os.Stat(gobinManifest.lockPath); err_ == nil {
		gobinManifest.pkgMapVer = V(minlib.PkgVerLockMap(dirPath))
	}
	for _, entry := range gobinManifest.entries {
		if lockedVer, ok := gobinManifest.pkgMapVer[entry.Pkg]; ok {
			entry.Version = lockedVer
		}
	}
	return
}

func (mani *manifestT) saveLockfile() (err error) {
	return mani.saveLockfileAs(mani.lockPath)
}

func (mani *manifestT) saveLockfileAs(filePath string) (err error) {
	defer Catch(&err)
	writer := V(os.Create(filePath))
	defer (func() { V0(writer.Close()) })()
	sort.Slice(mani.entries, func(i, j int) bool {
		return mani.entries[i].Pkg < mani.entries[j].Pkg
	})
	for _, entry := range mani.entries {
		if entry.Version != "" {
			_, err = writer.WriteString(entry.Pkg + "@" + entry.Version + "\n")
		}
	}
	return
}

func (mani *manifestT) lookup(pattern string) (entry *maniEntry) {
	divs := strings.SplitN(pattern, "@", 2)
	pkg := ""
	base := ""
	if len(divs) == 2 {
		pkg = divs[0]
	} else if strings.Contains(pattern, "/") {
		pkg = pattern
	} else {
		base = pattern
	}
	if pkg == "" && base != "" {
		for _, entry_ := range mani.entries {
			pkgBase := path.Base(entry_.Pkg)
			if pkgBase == base {
				pkg = entry_.Pkg
			}
		}
	}
	if pkg == "" {
		return
	}
	for _, entry_ := range mani.entries {
		if entry_.Pkg == pkg {
			entry = entry_
		}
	}
	return
}
