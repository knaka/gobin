package gobin

import (
	"fmt"
	"github.com/knaka/go-testutils/fs"
	"github.com/knaka/gobin/minlib"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	. "github.com/knaka/go-utils"
)

func TestCommandEx(t *testing.T) {
	cmd := V(CommandEx([]string{"stringer", "-help"}))
	println(cmd)
}

func Test_goModModules(t *testing.T) {
	confDirPath, _ := V2(minlib.ConfDirPath())
	goMod := V(parseGoMod(confDirPath))
	assert.NotNil(t, goMod.requiredModule("github.com/knaka/go-utils"))
	assert.Nil(t, goMod.requiredModule("github.com/knaka/go-utils/cmd/foo"))
	assert.NotNil(t, goMod.requiredModuleByPkg("github.com/knaka/go-utils/cmd/foo"))
	assert.Nil(t, goMod.requiredModuleByPkg("github.com/knaka/go-util"))
}

// canonAbs returns the canonical absolute path of the given value.
func canonAbs(s string) (ret string, err error) {
	ret, err = filepath.Abs(s)
	if err != nil {
		return
	}
	ret, err = filepath.EvalSymlinks(ret)
	if err != nil {
		return
	}
	ret = filepath.Clean(ret)
	return
}

func Test_parseManifest(t *testing.T) {
	tempDir := V(canonAbs(V(os.MkdirTemp("", "gobin-test"))))
	t.Cleanup(func() { Ignore(os.RemoveAll(tempDir)) })

	testdataDirPath := filepath.Join(tempDir, "minlib", "testdata")
	V0(fs.CopyDir(testdataDirPath, filepath.Join("minlib", "testdata")))
	V0(fs.CopyFile(
		filepath.Join(testdataDirPath, "foo", "bar", "go.mod"),
		filepath.Join(testdataDirPath, "foo", "bar", "go.mod.orig"),
	))

	confDirPath := filepath.Join(testdataDirPath, "foo", "bar")
	manifest := V(parseManifest(confDirPath))
	assert.NotNil(t, manifest)

	var entry *maniEntry

	entry = manifest.lookup("stringer")
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

	newLockPath := filepath.Join(tempDir, maniLockBase)
	V0(manifest.saveLockfileAs(newLockPath))
}

func Test_candidateModules(t *testing.T) {
	type args struct {
		pkg string
	}
	tests := []struct {
		name string
		args args
		ret  []string
		err  error
	}{
		{
			"Test 1",
			args{"golang.org/x/tools/cmd/stringer"},
			[]string{
				"golang.org/x/tools/cmd/stringer",
				"golang.org/x/tools/cmd",
				"golang.org/x/tools",
				"golang.org/x",
				"golang.org",
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, errGotten := candidateModules(tt.args.pkg)
			assert.Equal(t, tt.err, errGotten)
			if errGotten == nil {
				assert.Equalf(t, tt.ret, ret, "candidateModules(%v)", tt.args.pkg)
			}
		})
	}
}

func Test_queryVersion(t *testing.T) {
	type args struct {
		pkg string
	}
	tests := []struct {
		name        string
		args        args
		wantVersion *regexp.Regexp
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Test",
			args{"golang.org/x/tools/cmd/stringer"},
			regexp.MustCompile(`^v\d+\.\d+\.\d+$`),
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, err := queryVersion(tt.args.pkg)
			if !tt.wantErr(t, err, fmt.Sprintf("queryVersion(%v)", tt.args.pkg)) {
				return
			}
			assert.Truef(t, tt.wantVersion.MatchString(gotVersion), "queryVersion(%v) = %v, want match %v", tt.args.pkg, gotVersion, tt.wantVersion)
		})
	}
}
