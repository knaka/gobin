package minlib

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	. "github.com/knaka/go-utils"
)

func TestGetConfPath(t *testing.T) {
	defaultTempDir := os.TempDir()
	Ignore(os.Remove(filepath.Join(defaultTempDir, "go.mod")))
	tempDir := V(os.MkdirTemp(defaultTempDir, "gobin-test"))
	t.Cleanup(func() { Ignore(os.RemoveAll(tempDir)) })
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
			V(filepath.Abs(filepath.Join(".."))),
			V(filepath.Abs(filepath.Join("..", ".gobin"))),
			false,
		},
		{
			"Test go.mod",
			args{[]ConfDirPathOption{WithInitialDir(filepath.Join(
				".",
				"testdata",
				"foo",
				"bar",
				"baz",
			))}},
			V(filepath.Abs(filepath.Join(
				"testdata",
				"foo",
				"bar",
			))),
			V(filepath.Abs(filepath.Join(
				"testdata",
				"foo",
				"bar",
				".gobin",
			))),
			false,
		},
		{
			"No go.mod",
			args{
				[]ConfDirPathOption{WithInitialDir(filepath.Join(tempDir))},
			},
			"",
			"",
			true,
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
	//confDirPath, _ := V2(ConfDirPath())
	confDirPath := filepath.Join("testdata", "foo", "bar")
	lockList := V(PkgVerLockMap(confDirPath))
	assert.Equal(t, "v0.23.0", lockList["golang.org/x/tools/cmd/stringer"])
	assert.Equal(t, "v4.1.0", lockList["github.com/hairyhenderson/gomplate/v4/cmd/gomplate"])
}
