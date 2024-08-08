package log

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
	std = stdlog.New(os.Stderr, prefix, stdlog.LstdFlags)
}

func Println(v ...interface{}) {
	std.Println(v...)
}

func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

func SetSilent(f bool) {
	if f {
		std.SetOutput(io.Discard)
	} else {
		std.SetOutput(os.Stderr)
	}
}

func Silent() bool {
	return std.Writer() == io.Discard
}
