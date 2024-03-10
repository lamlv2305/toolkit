package rok

import "errors"

func Fallback[T any](args ...func() (*T, error)) (*T, error) {
	if len(args) == 0 {
		return nil, errors.New("empty conditions")
	}

	rest := args[:len(args)-1]
	last := args[len(args)-1]

	for idx := range rest {
		result, err := rest[idx]()

		if err == nil && result != nil {
			return result, nil
		}
	}

	return last()
}
