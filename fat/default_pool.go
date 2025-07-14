package fat

import (
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

var (
	defaultPool *ants.Pool
	poolOnce    sync.Once
)

func getDefaultPool() *ants.Pool {
	poolOnce.Do(func() {
		var err error
		defaultPool, err = ants.NewPool(100,
			ants.WithExpiryDuration(30*time.Second),
			ants.WithPreAlloc(false),
			ants.WithNonblocking(true),
		)
		if err != nil {
			panic(err)
		}
	})
	return defaultPool
}
