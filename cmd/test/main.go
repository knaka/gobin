package main

import (
	. "github.com/knaka/go-utils"
	"log"
	"os"
)

func Foo() (file *os.File, err error) {
	defer Catch(&err)
	file = V(os.Open("not_exists"))
	return
}

func Bar() (file *os.File, err error) {
	defer Catch(&err)
	file = V(Foo())
	return
}

func main() {
	//file := V(os.Open("not_exists"))
	file, err := Bar()
	log.Printf("%x, %+v", file, err)
}
