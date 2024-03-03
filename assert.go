package utils

import "errors"

func Assert(b bool) {
	if !b {
		panic(errors.New("assertion failed"))
	}
}
