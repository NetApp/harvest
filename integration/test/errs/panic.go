package errs

func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}
