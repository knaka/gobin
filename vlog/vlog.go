package vlog

import (
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
)

var std *stdlog.Logger

func Default() *stdlog.Logger {
	return std
}

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	prefix := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath))) + ": "
	std = stdlog.New(io.Discard, prefix, stdlog.LstdFlags)
}

func Println(v ...interface{}) {
	std.Println(v...)
}

func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

func SetVerbose(f bool) {
	if f {
		std.SetOutput(os.Stderr)
	} else {
		std.SetOutput(io.Discard)
	}
}

func Verbose() bool {
	return std.Writer() == os.Stderr
}
