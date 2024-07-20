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

// TernaryF returns the result of the first function if cond is true, otherwise the result of the second function.
func TernaryF[T any](
	cond bool,
	t func() T,
	f func() T,
) (ret T) {
	if cond {
		if t != nil {
			ret = t()
		}
	} else {
		if f != nil {
			ret = f()
		}
	}
	return
}

// TernaryF2 returns the result of the first function if cond is true, otherwise the result of the second function.
func TernaryF2[T any, U any](
	cond bool,
	t func() (T, U),
	f func() (T, U),
) (ret1 T, ret2 U) {
	if cond {
		if t != nil {
			ret1, ret2 = t()
		}
	} else {
		if f != nil {
			ret1, ret2 = f()
		}
	}
	return
}
