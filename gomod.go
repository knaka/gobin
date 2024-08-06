package gobin

import (
	. "github.com/knaka/go-utils"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"os"
	"path/filepath"
	"strings"
)

const goModBase = "go.mod"

// goModDefT represents the go.mod module definition file.
type goModDefT struct {
	name            string
	requiredModules []*module.Version
}

func (mod *goModDefT) requiredModule(moduleName string) (x *module.Version) {
	for _, req := range mod.requiredModules {
		if req.Path == moduleName {
			x = req
			return
		}
	}
	return
}

func (mod *goModDefT) requiredModuleByPkg(pkgName string) (x *module.Version) {
	for _, reqMod := range mod.requiredModules {
		if reqMod.Path == pkgName || strings.HasPrefix(pkgName, reqMod.Path+"/") {
			x = reqMod
			return
		}
	}
	return
}

func parseGoMod(dirPath string) (goModDef *goModDefT, err error) {
	defer Catch(&err)
	filePath := TernaryF(
		filepath.Base(dirPath) == goModBase,
		func() string { return dirPath },
		func() string { return filepath.Join(dirPath, goModBase) },
	)
	goModFile := V(modfile.Parse(filePath, V(os.ReadFile(filePath)), nil))
	goModDef = &goModDefT{
		name: goModFile.Module.Mod.Path,
		requiredModules: lo.Map(goModFile.Require, func(reqMod *modfile.Require, _ int) *module.Version {
			return &reqMod.Mod
		}),
	}
	return
}
