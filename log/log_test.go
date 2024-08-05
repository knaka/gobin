package log

import (
	"os"
	"testing"
)

func TestPrintln(t *testing.T) {
	SetOutput(os.Stderr)
	type args struct {
		v []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test", args{[]interface{}{"Test"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Println(tt.args.v...)
		})
	}
}
