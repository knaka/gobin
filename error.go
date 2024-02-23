package utils

import "github.com/friendsofgo/errors"

func wrapWithStack(err error) error {
	if _, ok := err.(interface{ Cause() error }); ok {
		return err
	}
	return errors.Wrap(err, "Error wrapped with stack")
}

// V returns the value. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func V[T any](value T, err error) T {
	if err != nil {
		panic(wrapWithStack(err))
	}
	return value
}

// V0 returns no value. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func V0[T any](first T, rest ...any) {
	if len(rest) > 0 {
		if err, ok := (rest[len(rest)-1]).(error); ok && err != nil {
			panic(wrapWithStack(err))
		}
	}
	if err, ok := any(first).(error); ok && err != nil {
		panic(wrapWithStack(err))
	}
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
func Ignore[T any](T, ...any) {}
