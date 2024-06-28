package utils

import (
	"fmt"
	"github.com/friendsofgo/errors"
)

// A panics with the given message if the condition is false.
func A(b bool, msgS ...string) {
	if b {
		return
	}
	err := TernaryF(len(msgS) == 0,
		func() error {
			return errors.New("assertion failed")
		},
		func() (err error) {
			err = errors.New(msgS[0])
			for _, msg := range msgS[1:] {
				err = fmt.Errorf("%s: %w", msg, err)
			}
			return err
		},
	)
	panic(errors.Wrap(err, "wrapped with stack"))
}

var Assert = A
