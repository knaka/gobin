package main

import (
	"github.com/knaka/gobin/lib"
	"os"
)

func main() {
	err := lib.RunCmd(os.Args[1:])
	if err != nil {
		panic(err)
	}
}
