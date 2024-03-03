package utils

import "errors"

func Assert(b bool, msgs ...string) {
	panic(errors.New("assertion failed"))
}
