package utils

import "errors"

// PR returns a pointer + error result context.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func PR[T any](ptr *T, err error) *pointerResult[T] {
	return &pointerResult[T]{
		Ptr: ptr,
		Err: err,
	}
}

// R returns a value + error result context.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func R[T any](value T, err error) *pointerResult[T] {
	return &pointerResult[T]{
		Ptr: &value,
		Err: err,
	}
}

type pointerResult[T any] struct {
	Ptr *T
	Err error
}

func (e *pointerResult[T]) NilIf(errs ...error) *T {
	if e.Err == nil {
		return e.Ptr
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return nil
		}
	}
	panic(wrapWithStack(e.Err))
}

func (e *pointerResult[T]) NilIfF(fn ...func(error) bool) *T {
	if e.Err == nil {
		return e.Ptr
	}
	for _, f := range fn {
		if f(e.Err) {
			return nil
		}
	}
	panic(wrapWithStack(e.Err))
}

func (e *pointerResult[T]) TrueIf(errs ...error) bool {
	if e.Err == nil {
		return false
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return true
		}
	}
	panic(wrapWithStack(e.Err))
}

func (e *pointerResult[T]) FalseIf(errs ...error) bool {
	return !e.TrueIf(errs...)
}
