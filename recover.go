package utils

func RecoverError(errRef *error) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			if errRef != nil {
				*errRef = err
			}
		} else {
			panic(r)
		}
	}
}

func RecoverPanicObject[T comparable](errRef *error, target T, fn func() error) {
	if r := recover(); r != nil {
		if err, ok := r.(T); ok && err == target {
			var newError error
			if fn != nil {
				newError = fn()
			}
			if errRef != nil {
				*errRef = newError
			}
		} else {
			panic(r)
		}
	}
}

func RecoverPanicType[T any](errRef *error, fn func(T) error) {
	if r := recover(); r != nil {
		if err, ok := r.(T); ok {
			newError := fn(err)
			if errRef != nil {
				*errRef = newError
			}
		} else {
			panic(r)
		}
	}
}
