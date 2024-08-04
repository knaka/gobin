package utils

func Throw(err error) {
	panic(WithStack(err))
}

func Catch(errRef *error, fns ...func(error)) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			if errRef != nil {
				*errRef = err
			}
			// Call the callback functions to do something with the error.
			for _, fn := range fns {
				fn(err)
			}
		} else {
			// “Re-throwing” panic does not lose the stack trace but stacks this call over the original. So you can trace to the original panic.
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
