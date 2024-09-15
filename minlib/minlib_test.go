package minlib

import (
	fsutils "github.com/knaka/go-utils/fs"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/knaka/go-utils"
)

func TestGetConfPath(t *testing.T) {
	defaultTempDir := os.TempDir()
	setMockRootDir(defaultTempDir)
	Ignore(os.Remove(filepath.Join(defaultTempDir, "go.mod")))
	tempDir := V(realpath(V(os.MkdirTemp(defaultTempDir, "gobin-test"))))
	noGoMod := filepath.Join(tempDir, "no-go-mod")
	V0(os.MkdirAll(noGoMod, 0755))
	hasGoMod := filepath.Join(tempDir, "has-go-mod", "foo", "bar")
	V0(os.MkdirAll(hasGoMod, 0755))
	goModDir := filepath.Join(tempDir, "has-go-mod", "foo")
	V0(fsutils.Touch(filepath.Join(goModDir, "go.mod")))
	t.Cleanup(func() { Ignore(os.RemoveAll(tempDir)) })
	testdataDirPath := filepath.Join(tempDir, "testdata")
	V0(fsutils.Copy(filepath.Join("testdata"), testdataDirPath))
	V0(fsutils.Move(
		filepath.Join(testdataDirPath, "foo", "bar", "go.mod.orig"),
		filepath.Join(testdataDirPath, "foo", "bar", "go.mod"),
	))
	V0(os.MkdirAll(filepath.Join("..", ".gobin"), 0755))
	type args struct {
		opts []ConfDirPathOption
	}
	tests := []struct {
		name            string
		args            args
		wantConfDirPath string
		wantGobinPath   string
		wantErr         bool
	}{
		{
			"This project's go.mod file",
			args{},
			V(fsutils.CanonPath(filepath.Join(".."))),
			V(fsutils.CanonPath(filepath.Join("..", ".gobin"))),
			false,
		},
		{
			"Test go.mod",
			args{[]ConfDirPathOption{WithInitialDir(filepath.Join(
				testdataDirPath,
				"foo",
				"bar",
				"baz",
			))}},
			V(filepath.Abs(filepath.Join(
				testdataDirPath,
				"foo",
			))),
			V(filepath.Abs(filepath.Join(
				testdataDirPath,
				"foo",
				".gobin",
			))),
			false,
		},
		{
			"No go.mod",
			args{
				[]ConfDirPathOption{WithInitialDir(noGoMod)},
			},
			"",
			"",
			true,
		},
		{
			"Has go.mod",
			args{
				[]ConfDirPathOption{WithInitialDir(hasGoMod)},
			},
			goModDir,
			filepath.Join(goModDir, ".gobin"),
			false,
		},
		{
			"Global",
			args{[]ConfDirPathOption{WithGlobal(true)}},
			V(os.UserHomeDir()),
			Elvis(os.Getenv("GOBIN"), filepath.Join(V(os.UserHomeDir()), "go", "bin")),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfDirPath, gotGobinDirPath, err := ConfDirPath(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfDirPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotConfDirPath != tt.wantConfDirPath {
				t.Errorf("ConfDirPath() gotConfDirPath = %v, want %v", gotConfDirPath, tt.wantConfDirPath)
			}
			if gotGobinDirPath != tt.wantGobinPath {
				t.Errorf("ConfDirPath() gotGobinDirPath = %v, want %v", gotGobinDirPath, tt.wantGobinPath)
			}
		})
	}
}

func Test_manifestLockModules(t *testing.T) {
	confDirPath := filepath.Join("testdata", "foo")
	lockList := V(PkgVerLockMap(confDirPath))
	assert.Equal(t, "v0.23.0", lockList["golang.org/x/tools/cmd/stringer"])
	assert.Equal(t, "v4.1.0", lockList["github.com/hairyhenderson/gomplate/v4/cmd/gomplate"])
}

func TestRunCommand(t *testing.T) {
	type args struct {
		name string
		arg  []string
	}
	tests := []struct {
		name        string
		args        args
		wantExecErr *exec.ExitError
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Test",
			args{"true", []string{}},
			nil,
			assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExecErr, err := RunCommand(tt.args.name, tt.args.arg...)
			_ = gotExecErr
			_ = err
			//if !tt.wantErr(t, err, fmt.Sprintf("RunCommand(%v, %v)", tt.args.name, tt.args.arg...)) {
			//	return
			//}
			//assert.Equalf(t, tt.wantExecErr, gotExecErr, "RunCommand(%v, %v)", tt.args.name, tt.args.arg...)
		})
	}
}

func Test_getGobin(t *testing.T) {
	gobin, err := getGoroot()
	assert.Nil(t, err)
	assert.Equal(t, "", gobin)
}
