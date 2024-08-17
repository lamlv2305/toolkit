package channel

import (
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/zerolog/log"
	"sync"
)

func NewFanout[T any](params ...WithOptions[T]) Fanout[T] {
	opts := options[T]{}
	for idx := range params {
		params[idx](&opts)
	}

	ins := Fanout[T]{
		data:    cmap.New[*SingleData[T]](),
		options: opts,
	}

	if opts.relay > 0 {
		ins.relay = &RelaySlice{
			mu:   &sync.RWMutex{},
			data: nil,
			size: opts.relay,
		}
	}

	return ins
}

type Fanout[T any] struct {
	data    cmap.ConcurrentMap[string, *SingleData[T]]
	relay   *RelaySlice
	options options[T]
}

type SingleData[T any] struct {
	ch     chan T
	closed bool
	ignore func(T) bool
}

func (f *Fanout[T]) Send(sendingData T) {
	defer func() {
		// Quick hack
		// Do ko đc phép close khi có nhiều subscriber nhưng mà đang quick hack nên tạm thời để như này đã
		if r := recover(); r != nil {
			log.Error().Any("err", r).Msg("Fail to handle channel")
		}
	}()

	for idx := range f.options.sideEffect {
		// show something like log here
		go f.options.sideEffect[idx](sendingData)
	}

	for idx := range f.options.middlewares {
		if !f.options.middlewares[idx](sendingData) {
			// filter data
			return
		}
	}

	if f.relay != nil {
		go f.relay.Add(sendingData)
	}

	for m := range f.data.IterBuffered() {
		if m.Val.closed {
			continue
		}

		if m.Val.ignore != nil && m.Val.ignore(sendingData) {
			continue
		}

		m.Val.ch <- sendingData
	}
}

// Wait
// buffer -> channel size
// ignore -> tương tự filter bên js
func (f *Fanout[T]) Wait(buffer int, ignore func(T) bool) (chan T, func()) {
	ch := make(chan T, buffer)

	id := uuid.New().String()
	f.data.Set(id, &SingleData[T]{
		ch:     ch,
		closed: false,
		ignore: ignore,
	})

	go func() {
		// sending relay data
		if f.relay == nil {
			return
		}

		relayedData := f.relay.Get()
		for idx := range relayedData {
			if ignore != nil && ignore(relayedData[idx].(T)) {
				continue
			}

			ch <- relayedData[idx].(T)
		}
	}()

	return ch, func() {
		f.data.RemoveCb(id, func(key string, v *SingleData[T], exists bool) bool {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Any("request", r).Msg("[CRITICAL] Fail to handling close fanout")
				}
			}()

			if exists && v != nil && !v.closed {
				v.closed = true
				close(v.ch)
			}

			return true
		})
	}
}
