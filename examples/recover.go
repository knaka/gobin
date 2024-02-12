package main

import (
	. "github.com/knaka/go-utils"
	"log"
	"os"
)

func foo() (err error) {
	defer RecoverError(&err)
	_ = Ensure(os.ReadDir("hoge"))
	//return io.EOF
	return nil
}

func main() {
	if err := foo(); err != nil {
		log.Printf("error: %+v", err)
	}
}
