package utils

// Ternary returns the result of the first function if cond is true, otherwise the result of the second function.
func Ternary[T any](
	cond bool,
	t func() T,
	f func() T,
) T {
	if cond {
		return t()
	}
	return f()
}
