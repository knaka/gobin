package utils

// Ternary returns the first value if cond is true, otherwise the second value.
func Ternary[T any](
	cond bool,
	t T,
	f T,
) T {
	if cond {
		return t
	}
	return f
}

// Ternary returns the result of the first function if cond is true, otherwise the result of the second function.
func TernaryF[T any](
	cond bool,
	t func() T,
	f func() T,
) T {
	if cond {
		if t == nil {
			return empty[T]()
		}
		return t()
	}
	if f == nil {
		return empty[T]()
	}
	return f()
}
