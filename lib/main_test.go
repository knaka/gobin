package lib

import (
	"regexp"
	"testing"
)

func Test_isPackage(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test 1", args: args{"github.com/knaka/gobin/cmd/foo@latest"}, want: true},
		{name: "Test 2", args: args{"foo@example.com"}, want: false},
		{name: "Test 3", args: args{"golang.org/x/tools/cmd/goyacc@latest"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPackage(tt.args.s); got != tt.want {
				t.Errorf("isPackage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveVersion(t *testing.T) {
	type args struct {
		pkg string
		ver string
	}
	tests := []struct {
		name            string
		args            args
		wantResolvedVer *regexp.Regexp
		wantErr         bool
	}{
		{
			"Test 1",
			args{"github.com/wailsapp/wails/v2", "latest"},
			regexp.MustCompile(`v2\..*`),
			false,
		},
		{
			"Test 2",
			args{"golang.org/x/tools/cmd/goyacc@latest", "latest"},
			regexp.MustCompile(`v0\..*`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResolvedVer, err := resolveVersion(tt.args.pkg, tt.args.ver)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantResolvedVer.MatchString(gotResolvedVer) {
				t.Errorf("resolveVersion() gotResolvedVer = %v, want %v", gotResolvedVer, tt.wantResolvedVer)
			}
		})
	}
}
