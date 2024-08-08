package utils

import (
	"fmt"
	"io"
)

var logWriters []io.Writer

func AddLogWriter(writer io.Writer) {
	logWriters = append(logWriters, writer)
}

func LogPrintf(format string, v ...interface{}) {
	for _, writer := range logWriters {
		_, _ = writer.Write([]byte(fmt.Sprintf(format, v...)))
	}
}

func LogPrint(v ...interface{}) {
	for _, writer := range logWriters {
		_, _ = writer.Write([]byte(fmt.Sprint(v...)))
	}
}

func LogPrintln(v ...interface{}) {
	for _, writer := range logWriters {
		_, _ = writer.Write([]byte(fmt.Sprintln(v...)))
	}
}
