package utils

func Assign[T any](dst *T, src T) T {
	if dst != nil {
		*dst = src
	}
	return *dst
}
