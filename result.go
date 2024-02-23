package utils

import "errors"

// Result returns a pointer + error result context to ignore specific errors.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func Result[T any](ptr *T, err error) *ptrResult[T] {
	return &ptrResult[T]{
		Ptr: ptr,
		Err: err,
	}
}

// ValueResult returns a value + error result context to ignore specific errors.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func ValueResult[T any](value T, err error) *ptrResult[T] {
	return &ptrResult[T]{
		Ptr: &value,
		Err: err,
	}
}

type ptrResult[T any] struct {
	Ptr *T
	Err error
}

func (e *ptrResult[T]) NilIf(errs ...error) *T {
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

//func (e *ptrResult[T]) NilIfF(fn ...func(error) bool) (*T, error) {
//	if e.Err == nil {
//		return e.Ptr, nil
//	}
//	for _, f := range fn {
//		if f(e.Err) {
//			return nil, nil
//		}
//	}
//	return e.Ptr, e.Err
//}

func (e *ptrResult[T]) TrueIf(errs ...error) bool {
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

func (e *ptrResult[T]) FalseIf(errs ...error) bool {
	return !e.TrueIf(errs...)
}
