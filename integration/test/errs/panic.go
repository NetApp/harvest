package errs

// PanicIfNotNil panics if the provided error is not nil. It is intended for use in contexts where an error is considered fatal.
func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}
