package log

import (
	"io"
	stdlog "log"
	"path/filepath"
	"runtime"
	"strings"
)

var logger *stdlog.Logger

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	prefix := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath))) + " "
	logger = stdlog.New(io.Discard, prefix, stdlog.LstdFlags)
}

func Println(v ...interface{}) {
	pc, _, _, _ := runtime.Caller(2)
	x := runtime.FuncForPC(pc).Name()
	println(x)
	parts := strings.Split(x, ".")
	println(parts)

	logger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}
