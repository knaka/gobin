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
