package utils

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
