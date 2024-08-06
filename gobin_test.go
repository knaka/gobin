package lib

import (
	"github.com/knaka/gobin/minlib"
	"github.com/stretchr/testify/assert"
	"testing"

	. "github.com/knaka/go-utils"
)

func TestCommandEx(t *testing.T) {
	cmd := V(CommandEx([]string{"stringer", "-help"}))
	println(cmd)
}

func Test_goModModules(t *testing.T) {
	confDirPath, _ := V2(minlib.ConfAndGobinPaths())
	goMod := V(parseGoMod(confDirPath))
	assert.NotNil(t, goMod.requiredModule("github.com/knaka/go-utils"))
	assert.Nil(t, goMod.requiredModule("github.com/knaka/go-utils/cmd/foo"))
	assert.NotNil(t, goMod.requiredModuleByPkg("github.com/knaka/go-utils/cmd/foo"))
	assert.Nil(t, goMod.requiredModuleByPkg("github.com/knaka/go-util"))
}

func Test_parseManifest(t *testing.T) {
	confDirPath, _ := V2(minlib.ConfAndGobinPaths())
	manifest := V(parseManifest(confDirPath))
	assert.NotNil(t, manifest)
}
