package fat

import (
	"context"
	"time"

	"github.com/panjf2000/ants/v2"
)

type AsyncOption struct {
	ctx     context.Context
	pool    *ants.Pool
	timeout time.Duration
}

type AsyncOptionFunc func(*AsyncOption)

func WithContext(ctx context.Context) AsyncOptionFunc {
	return func(opt *AsyncOption) {
		opt.ctx = ctx
	}
}

func WithPool(pool *ants.Pool) AsyncOptionFunc {
	return func(opt *AsyncOption) {
		opt.pool = pool
	}
}

func WithTimeout(timeout time.Duration) AsyncOptionFunc {
	return func(opt *AsyncOption) {
		opt.timeout = timeout
	}
}

// AsyncExec executes a function asynchronously using ants pool
func AsyncExec(fn func(ctx context.Context) error, opts ...AsyncOptionFunc) chan error {
	options := &AsyncOption{
		ctx:  context.Background(),
		pool: getDefaultPool(),
	}

	for _, opt := range opts {
		opt(options)
	}

	// Apply timeout if specified
	if options.timeout > 0 {
		ctx, cancel := context.WithTimeout(options.ctx, options.timeout)
		defer cancel()
		options.ctx = ctx
	}

	errCh := make(chan error, 1)

	// Submit task to ants pool
	err := options.pool.Submit(func() {
		defer close(errCh)

		// Create a done channel for the function execution
		done := make(chan error, 1)

		// Execute function in a separate goroutine to handle context cancellation
		go func() {
			done <- fn(options.ctx)
		}()

		// Wait for either completion or context cancellation
		select {
		case err := <-done:
			if err != nil {
				errCh <- err
			}
		case <-options.ctx.Done():
			errCh <- options.ctx.Err()
		}
	})
	// If pool submission fails, handle synchronously
	if err != nil {
		go func() {
			defer close(errCh)
			errCh <- err
		}()
	}

	return errCh
}

// Cleanup releases the default pool (call on shutdown)
func Cleanup() {
	if defaultPool != nil {
		defaultPool.Release()
	}
}
