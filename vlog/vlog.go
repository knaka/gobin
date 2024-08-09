package vlog

import (
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var logger *stdlog.Logger

//goland:noinspection GoUnusedExportedFunction
func Logger() *stdlog.Logger {
	return logger
}

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	name := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath)))
	divs := strings.Split(name, "@")
	name = divs[0]
	prefix := name + ": "
	logger = stdlog.New(io.Discard, prefix, stdlog.LstdFlags)
}

//goland:noinspection GoUnusedExportedFunction
func Println(v ...interface{}) {
	logger.Println(v...)
}

//goland:noinspection GoUnusedExportedFunction
func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

//goland:noinspection GoUnusedExportedFunction
func SetVerbose(f bool) {
	if f {
		logger.SetOutput(os.Stderr)
	} else {
		logger.SetOutput(io.Discard)
	}
}

//goland:noinspection GoUnusedExportedFunction
func Verbose() bool {
	return logger.Writer() == os.Stderr
}
