package main

import (
	"errors"
	"io"
	"log"
	"strings"
)

import . "github.com/knaka/go-utils"

func Bar() (err error) {
	defer Catch(&err)
	reader := strings.NewReader("Hello, Reader!")
	bytes := make([]byte, 8)
	for {
		_ = V(reader.Read(bytes))
	}
}

func Foo() (err error) {
	defer Catch(&err)
	V0(Bar())
	return
}

func Main() (err error) {
	defer Catch(&err)
	V0(Foo())
	return
}

func main() {
	V0((func() error { return nil })())
	//err := Bar()
	err := Main()
	Assertf(errors.Is(err, io.EOF), "should be io.EOF: %v", err)
	if err != nil {
		log.Fatalf("%v", err)
	}
	//if err != nil {
	//	panic(err)
	//}
}
