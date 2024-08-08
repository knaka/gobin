package log

import (
	"io"
	stdlog "log"
	"path/filepath"
	"runtime"
)

var Logger *stdlog.Logger

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	prefix := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath))) + ": "
	Logger = stdlog.New(io.Discard, prefix, stdlog.LstdFlags)
}

func Println(v ...interface{}) {
	Logger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	Logger.Printf(format, v...)
}

func SetOutput(w io.Writer) {
	Logger.SetOutput(w)
}
