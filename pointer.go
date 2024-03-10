package rok

func Pointer[T any](data T) *T {
	return &data
}

func Value[T any](data *T, defaultValue T) T {
	if data == nil {
		return defaultValue
	}

	return *data
}
