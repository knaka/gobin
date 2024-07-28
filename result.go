package utils

import "errors"

type result[T any] struct {
	Value T
	Err   error
}

// R returns a value + error result context.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func R[T any](value T, err error) *result[T] {
	return &result[T]{
		Value: value,
		Err:   err,
	}
}

// PR returns a pointer + error result context.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func PR[T any](value T, err error) *result[*T] {
	return &result[*T]{
		Value: &value,
		Err:   err,
	}
}

// NilIf returns nil if the error is one of the given errors and returns the value otherwise.
func (e *result[T]) NilIf(errs ...error) T {
	if e.Err == nil {
		return e.Value
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return Nil[T]()
		}
	}
	panic(WithStack(e.Err))
}

func (e *result[T]) NilIfF(fn ...func(error) bool) T {
	if e.Err == nil {
		return e.Value
	}
	for _, f := range fn {
		if f(e.Err) {
			return Nil[T]()
		}
	}
	panic(WithStack(e.Err))
}

func (e *result[T]) TrueIf(errs ...error) bool {
	if e.Err == nil {
		return false
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return true
		}
	}
	panic(WithStack(e.Err))
}

func (e *result[T]) FalseIf(errs ...error) bool {
	return !e.TrueIf(errs...)
}
