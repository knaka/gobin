package main

import (
	"github.com/knaka/gobin/lib"
	"os"
)

func main() {
	err := lib.Run(os.Args[1:]...)
	if err != nil {
		panic(err)
	}
}
