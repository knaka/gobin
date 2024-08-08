package log

import (
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
)

var logger *stdlog.Logger

//goland:noinspection GoUnusedExportedFunction
func Logger() *stdlog.Logger {
	return logger
}

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	prefix := filepath.Base(filepath.Dir(filepath.Dir(thisFilePath))) + ": "
	logger = stdlog.New(os.Stderr, prefix, stdlog.LstdFlags)
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
func SetSilent(f bool) {
	if f {
		logger.SetOutput(io.Discard)
	} else {
		logger.SetOutput(os.Stderr)
	}
}

//goland:noinspection GoUnusedExportedFunction
func Silent() bool {
	return logger.Writer() == io.Discard
}
