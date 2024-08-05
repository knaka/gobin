package log

import (
	"io"
	stdlog "log"
	"path/filepath"
	"runtime"
)

var logger *stdlog.Logger

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	prefix := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath))) + " "
	logger = stdlog.New(io.Discard, prefix, stdlog.LstdFlags)
}

func Println(v ...interface{}) {
	logger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}
