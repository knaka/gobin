package utils

import (
	"fmt"
	"github.com/friendsofgo/errors"
)

type withStack struct {
	errWithStack error
}

var _ error = (*withStack)(nil)
var _ interface{ Unwrap() error } = (*withStack)(nil)

func (w *withStack) Error() string { return fmt.Sprintf("%+v", w.errWithStack) }
func (w *withStack) Unwrap() error { return w.errWithStack }

func WithStack(err error) error {
	var _withStack *withStack
	if errors.As(err, &_withStack) {
		return err
	}
	return &withStack{errWithStack: errors.WithStack(err)}
}

// V returns the value. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func V[T any](value T, err error) T {
	if err == nil {
		return value
	}
	panic(WithStack(err))
}

// V2 returns two values. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func V2[T any, U any](value1 T, value2 U, err error) (T, U) {
	if err == nil {
		return value1, value2
	}
	panic(WithStack(err))
}

// V3 returns three values. If err is not nil, it panics.
func V3[T any, U any, V any](value1 T, value2 U, value3 V, err error) (T, U, V) {
	if err == nil {
		return value1, value2, value3
	}
	panic(WithStack(err))
}

// V0 returns no value. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func V0[T any](first T, rest ...any) {
	if len(rest) > 0 {
		if err, ok := (rest[len(rest)-1]).(error); ok && err != nil {
			panic(WithStack(err))
		}
	}
	if err, ok := any(first).(error); ok && err != nil {
		panic(WithStack(err))
	}
	panic("no error argument passed")
}

// E returns the error.
func E(rest ...any) error {
	if len(rest) > 0 {
		last := rest[len(rest)-1]
		if last == nil {
			return nil
		}
		if err, ok := last.(error); ok {
			return err
		}
	}
	panic("no error argument passed")
}

// Ignore ignores an error explicitly.
//
//goland:noinspection GoUnusedParameter
func Ignore[T any](T, ...any) {}
