package utils

// Ensure checks if the value is available. If err is not nil, it panics.
//
// noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func Ensure[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Assert checks that err is nil. If err is not nil, it panics.
//
// noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func Assert[T any](first T, rest ...any) {
	if len(rest) > 0 {
		if err, ok := (rest[len(rest)-1]).(error); ok && err != nil {
			panic(err)
		}
	}
	if err, ok := any(first).(error); ok && err != nil {
		panic(err)
	}
}

// Ignore ignores errors explicitly.
func Ignore[T any](T, ...any) {}

// WithErrorContext returns an error context to ignore specific errors.
//
// noinspection GoExportedFuncWithUnexportedType
func WithErrorContext[T any](first T, rest ...any) *ptrErrorContext[any] {
	var err error
	if len(rest) > 0 {
		if errNew, ok := (rest[len(rest)-1]).(error); ok {
			err = errNew
		}
	} else if errNew, ok := any(first).(error); ok {
		err = errNew
	}
	return &ptrErrorContext[any]{
		Err: err,
	}
}

// WithValueErrorContext returns a value + error context to ignore specific errors.
//
// noinspection GoExportedFuncWithUnexportedType
func WithValueErrorContext[T any](value T, err error) *ptrErrorContext[T] {
	return &ptrErrorContext[T]{
		Ptr: &value,
		Err: err,
	}
}

// WithPtrErrorContext returns a pointer + error context to ignore specific errors.
//
// noinspection GoExportedFuncWithUnexportedType
func WithPtrErrorContext[T any](ptr *T, err error) *ptrErrorContext[T] {
	return &ptrErrorContext[T]{
		Ptr: ptr,
		Err: err,
	}
}

type ptrErrorContext[T any] struct {
	Ptr *T
	Err error
}

func (e *ptrErrorContext[T]) NilIf(errs ...error) *T {
	if e.Err == nil {
		return e.Ptr
	}
	for _, err := range errs {
		if e.Err == err {
			return nil
		}
	}
	panic(e.Err)
}

func (e *ptrErrorContext[T]) TrueIf(errs ...error) bool {
	if e.Err == nil {
		return false
	}
	for _, err := range errs {
		if e.Err == err {
			return true
		}
	}
	panic(e.Err)
}

func (e *ptrErrorContext[T]) FalseIf(errs ...error) bool {
	return !e.TrueIf(errs...)
}
