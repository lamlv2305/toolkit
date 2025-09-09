package toolkit

import (
	"sync/atomic"
	"time"
)

type WithHolder[T any] func(*Holder[T])

func WithCloser[T any](closer CloserFunc[T]) WithHolder[T] {
	return func(h *Holder[T]) {
		h.closer = closer
	}
}

func WithGrace[T any](grace time.Duration) WithHolder[T] {
	return func(h *Holder[T]) {
		h.grace = grace
	}
}

// CloserFunc is optional cleanup logic (e.g., close old DB/Redis connections).
type CloserFunc[T any] func(old *T)

// Holder provides lock-free read + atomic swap on reload.
type Holder[T any] struct {
	value  atomic.Value // holds *T
	closer CloserFunc[T]
	grace  time.Duration
}

// NewHolder creates a new hot-reload holder with optional cleanup.
func NewHolder[T any](opts ...WithHolder[T]) *Holder[T] {
	holder := &Holder[T]{
		closer: nil,
		grace:  time.Second * 5,
	}

	for _, opt := range opts {
		opt(holder)
	}

	return holder
}

func (h *Holder[T]) Get() *T {
	val := h.value.Load()
	if val == nil {
		return nil
	}
	return val.(*T)
}

func (h *Holder[T]) Set(newVal *T) {
	oldVal := h.Get()
	h.value.Store(newVal)

	if oldVal != nil && h.closer != nil {
		// Close old resource in background after grace period
		go func(old *T) {
			if h.grace > 0 {
				time.Sleep(h.grace)
			}

			if h.closer != nil {
				h.closer(old)
			}
		}(oldVal)
	}
}
