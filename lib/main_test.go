package lib

import (
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
