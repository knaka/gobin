package lib

import (
	"github.com/knaka/gobin/minlib"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
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
	confDirPath := filepath.Join("minlib", "testdata", "foo", "bar")
	manifest := V(parseManifest(confDirPath))
	assert.NotNil(t, manifest)

	entry := manifest.lookup("stringer")
	assert.Equal(t, "golang.org/x/tools/cmd/stringer", entry.Pkg)
	assert.Equal(t, "v0.23.0", entry.Version)
	assert.Equal(t, "", entry.Tags)

	entry = manifest.lookup("github.com/hairyhenderson/gomplate/v4/cmd/gomplate")
	assert.Equal(t, "github.com/hairyhenderson/gomplate/v4/cmd/gomplate", entry.Pkg)
	assert.Equal(t, "v4.1.0", entry.Version)
	assert.Equal(t, "foo,bar", entry.Tags)

	entry.Version = "v4.2.0"
	entry = manifest.lookup("github.com/hairyhenderson/gomplate/v4/cmd/gomplate")
	assert.Equal(t, "v4.2.0", entry.Version)

	tempDir := V(os.MkdirTemp("", "gobin"))
	t.Cleanup(func() { V0(os.RemoveAll(tempDir)) })

	newLockPath := filepath.Join(tempDir, maniLockBase)
	V0(manifest.saveLockfileAs(newLockPath))
	manifest.saveLockfile()
}
