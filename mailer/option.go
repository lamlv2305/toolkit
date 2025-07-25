package mailer

import "time"

type Options struct {
	retries int
	delay   time.Duration
}

type WithOptions func(*Options)

func WithRetry(count uint) WithOptions {
	return func(opts *Options) {
		opts.retries = int(count)
	}
}

func WithDelay(delay time.Duration) WithOptions {
	return func(opts *Options) {
		opts.delay = delay
	}
}
