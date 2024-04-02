package logger

type LogOption struct {
	Pretty bool
}

func WithPretty(pretty bool) func(option *LogOption) {
	return func(option *LogOption) {
		option.Pretty = pretty
	}
}
