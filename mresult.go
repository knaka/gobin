package utils

// Monad-style error handling functions.

// Ok returns the given value and nil.
func Ok[T any](value T) (T, error) {
	return value, nil
}

// Nil returns a zero value.
func Nil[T any]() (t T) { return }

// Err returns a zero value and the given error.
func Err[T any](err error) (T, error) {
	return Nil[T](), err
}

// Bind returns the result of the given function that can fail if err is nil, otherwise the error.
func Bind[T any](err error, fn func() (T, error)) (T, error) {
	if err != nil {
		return Nil[T](), err
	}
	t, err := fn()
	return t, TernaryF(err == nil,
		func() error { return nil },
		func() error { return wrapWithStack(err) },
	)
}

// Bind0 is an alias of Then.
func Bind0(err error, fn func() error) error {
	if err != nil {
		return err
	}
	err = fn()
	return TernaryF(err == nil,
		func() error { return nil },
		func() error { return wrapWithStack(err) },
	)
}

// Bind1 is an alias of Bind.
func Bind1[T any](err error, fn func() (T, error)) (T, error) {
	if err != nil {
		return Nil[T](), err
	}
	t, err := fn()
	return t, TernaryF(err == nil,
		func() error { return nil },
		func() error { return wrapWithStack(err) },
	)
}

func Bind2[T any, U any](err error, fn func() (T, U, error)) (T, U, error) {
	if err != nil {
		return Nil[T](), Nil[U](), err
	}
	t, u, err := fn()
	return t, u, TernaryF(err == nil,
		func() error { return nil },
		func() error { return wrapWithStack(err) },
	)
}

// Then calls the given function if err is nil.
func Then(err error, fn func() error) error {
	if err != nil {
		return err
	}
	err = fn()
	return TernaryF(err == nil,
		func() error { return nil },
		func() error { return wrapWithStack(err) },
	)
}

// Let returns the result of the given function if err is nil.
func Let[T any](err error, fn func() T) T {
	if err != nil {
		return Nil[T]()
	}
	return fn()
}

// Let0 calls the given function if err is nil and returns nothing.
func Let0(err error, fn func()) {
	if err != nil {
		return
	}
	fn()
}

// Let1 is an alias of Let.
func Let1[T any](err error, fn func() T) T {
	if err != nil {
		return Nil[T]()
	}
	return fn()
}

func Let2[T any, U any](err error, fn func() (T, U)) (T, U) {
	if err != nil {
		return Nil[T](), Nil[U]()
	}
	return fn()
}
