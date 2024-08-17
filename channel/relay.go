package channel

import "sync"

func WithRelay[T any](relay int) func(*options[T]) {
	return func(o *options[T]) {
		o.relay = relay
	}
}

type RelaySlice struct {
	mu   *sync.RWMutex
	data []any
	size int
}

func (r *RelaySlice) Add(data any) {
	r.mu.Lock()

	r.data = append(r.data, data)

	// Remove first item to keep relay size
	if len(r.data) > r.size {
		r.data = r.data[len(r.data)-r.size:]
	}

	r.mu.Unlock()
}

func (r *RelaySlice) Get() []any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}
