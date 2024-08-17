package channel

func WithSideEffect[T any](fn func(T)) WithOptions[T] {
	return func(o *options[T]) {
		o.sideEffect = append(o.sideEffect, fn)
	}
}
