package internal

// Cond returns the result of the first function if cond is true, otherwise the result of the second function.
func Cond[T any](
	cond bool,
	t func() T,
	f func() T,
) T {
	if cond {
		return t()
	}
	return f()
}

// Ptr returns a pointer to a copied value.
func Ptr[T any](v T) *T {
	return &v
}

// PtrMap returns a pointer to the result of the given function.
func PtrMap[T any, U any](pt *T, fn func(*T) *U) *U {
	if pt == nil {
		return nil
	}
	return fn(pt)
}

// Ok returns the given value and nil.
func Ok[T any](value T) (T, error) {
	return value, nil
}

func empty[T any]() (t T) { return }

// Err returns a zero value and the given error.
func Err[T any](err error) (T, error) {
	return empty[T](), err
}

// CondPipe returns the result of the given function if err is nil, otherwise the error.
func CondPipe[T any, U any](r T, err error, f func(T) (U, error)) (U, error) {
	if err != nil {
		return empty[U](), err
	}
	return f(r)
}

// CondFunc returns the result of the given function if err is nil, otherwise the error.
func CondFunc[T any](err error, fn func() (T, error)) (T, error) {
	if err != nil {
		return empty[T](), err
	}
	return fn()
}

// CondProc calls the given procedure] if err is nil.
func CondProc(err error, pr func()) {
	if err == nil {
		pr()
	}
}
