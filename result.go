package utils

import "errors"

// R returns a pointer + error result context to ignore specific errors.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func R[T any](ptr *T, err error) *ptrResult[T] {
	return &ptrResult[T]{
		Ptr: ptr,
		Err: err,
	}
}

// RV returns a value + error result context to ignore specific errors.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func RV[T any](value T, err error) *ptrResult[T] {
	return &ptrResult[T]{
		Ptr: &value,
		Err: err,
	}
}

type ptrResult[T any] struct {
	Ptr *T
	Err error
}

func (e *ptrResult[T]) NilIf(errs ...error) (*T, error) {
	if e.Err == nil {
		return e.Ptr, nil
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return nil, nil
		}
	}
	return e.Ptr, e.Err
}

func (e *ptrResult[T]) NilIfF(fn ...func(error) bool) (*T, error) {
	if e.Err == nil {
		return e.Ptr, nil
	}
	for _, f := range fn {
		if f(e.Err) {
			return nil, nil
		}
	}
	return e.Ptr, e.Err
}

func (e *ptrResult[T]) TrueIf(errs ...error) (bool, error) {
	if e.Err == nil {
		return false, nil
	}
	for _, err := range errs {
		if errors.Is(e.Err, err) {
			return true, nil
		}
	}
	return false, e.Err
}

func (e *ptrResult[T]) FalseIf(errs ...error) (bool, error) {
	b, err := e.TrueIf(errs...)
	return !b, err
}
