package utils

// P returns a pointer to a copied value.
func P[T any](v T) *T {
	return &v
}

// PMap returns a pointer to the result of the given function.
func PMap[T any, U any](pt *T, fn func(*T) *U) *U {
	if pt == nil {
		return nil
	}
	return fn(pt)
}

func Ptr[T any](v T) *T {
	return &v
}

func PtrMap[T any, U any](pt *T, fn func(*T) *U) *U {
	if pt == nil {
		return nil
	}
	return fn(pt)
}
