package channel

type WithOptions[T any] func(*options[T])

type options[T any] struct {
	middlewares []Middleware[T]
	sideEffect  []func(T)
	relay       int
}
