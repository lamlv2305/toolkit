package slice

func MapTo[T any, R any](input []T, transform func(T) R) []R {
	var result []R
	for idx := range input {
		result = append(result, transform(input[idx]))
	}

	return result
}
