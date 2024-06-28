package utils

import (
	"log"
	"testing"
)

func TestAssert(t *testing.T) {
	fnS := []func(){
		func() { A(true) },
		func() { A(true, "x") },
		func() { A(false) },
		func() { A(false, "a") },
		func() { A(false, "a", "b") },
		func() { A(false, "a", "b", "c") },
	}
	for _, fn := range fnS {
		err := func() (err error) {
			defer Catch(&err)
			fn()
			return nil
		}()
		log.Printf("%+v", err)
	}
}
