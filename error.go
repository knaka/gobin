package utils

// Ensure checks if the value is available. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func Ensure[T any](value T, err error) T {
	if err != nil {
		panic(wrapWithStack(err))
	}
	return value
}

func _1[T any](value T, err error) T {
	if err != nil {
		panic(wrapWithStack(err))
	}
	return value
}

// Ensure0 checks that err is nil. If err is not nil, it panics.
//
//goland:noinspection GoUnusedExportedFunction, GoUnnecessarilyExportedIdentifiers
func Ensure0[T any](first T, rest ...any) {
	if len(rest) > 0 {
		if err, ok := (rest[len(rest)-1]).(error); ok && err != nil {
			panic(wrapWithStack)
		}
	}
	if err, ok := any(first).(error); ok && err != nil {
		panic(wrapWithStack)
	}
}

func _0[T any](first T, rest ...any) {
	if len(rest) > 0 {
		if err, ok := (rest[len(rest)-1]).(error); ok && err != nil {
			panic(wrapWithStack)
		}
	}
	if err, ok := any(first).(error); ok && err != nil {
		panic(wrapWithStack)
	}
}

func Ensure1[T any](value T, err error) T {
	if err != nil {
		panic(wrapWithStack(err))
	}
	return value
}

func Ensure2[T any, U any](value T, value2 U, err error) (T, U) {
	if err != nil {
		panic(wrapWithStack(err))
	}
	return value, value2
}

// Ignore ignores errors explicitly.
func Ignore[T any](T, ...any) {}
