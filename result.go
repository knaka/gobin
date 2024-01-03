package utils

// Ok returns the given value and nil.
func Ok[T any](value T) (T, error) {
	return value, nil
}

func Empty[T any]() (t T) { return }

// Err returns a zero value and the given error.
func Err[T any](err error) (T, error) {
	return Empty[T](), err
}

// SafePipeErr returns the result of the given function that can fail if err is nil, otherwise the error.
func SafePipeErr[T any, U any](r T, err error, f func(T) (U, error)) (U, error) {
	if err != nil {
		return Empty[U](), err
	}
	return f(r)
}

// SafePipe returns the result of the given function if err is nil, otherwise the error.
func SafePipe[T any, U any](r T, err error, f func(T) U) (U, error) {
	if err != nil {
		return Empty[U](), err
	}
	return f(r), nil
}

// SafeCallErr returns the result of the given function that can fail if err is nil, otherwise the error.
func SafeCallErr[T any](err error, fn func() (T, error)) (T, error) {
	if err != nil {
		return Empty[T](), err
	}
	return fn()
}

// SafeCall returns the result of the given function if err is nil, otherwise the error.
func SafeCall[T any](err error, fn func() T) (T, error) {
	if err != nil {
		return Empty[T](), err
	}
	return fn(), nil
}

// SafeCall0 calls the given procedure if err is nil.
func SafeCall0(err error, pr func()) {
	if err == nil {
		pr()
	}
}
