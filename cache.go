package recached

import (
	"context"
	"sync"
	"time"
)

// ReCached is a cache that can be refreshed
type ReCached[T any] interface {
	Get() T
	Update()
}

type reCached[T any] struct {
	mu         sync.RWMutex
	value      T
	period     time.Duration
	updateFunc func() (T, error)
}

// Global registry to keep track of all cache instances
var (
	globalCachesMutex sync.RWMutex
	globalCaches      []interface{ Update() }
)

func New[T any](ctx context.Context, period time.Duration, updateFunc func() (T, error)) ReCached[T] {
	cache := &reCached[T]{
		period:     period,
		updateFunc: updateFunc,
	}

	cache.Update()
	go cache.updateLoop(ctx)

	// Register the cache in the global registry
	globalCachesMutex.Lock()
	globalCaches = append(globalCaches, cache)
	globalCachesMutex.Unlock()

	return cache
}

func (r *reCached[T]) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(r.period):
			r.Update()
		}
	}
}

func (r *reCached[T]) Get() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

func (r *reCached[T]) Update() {
	newValue, err := r.updateFunc()
	if err != nil {
		return
	}

	r.mu.Lock()
	r.value = newValue
	r.mu.Unlock()
}

// GlobalCacheUpdate updates all cache instances created via New
func GlobalCacheUpdate() {
	globalCachesMutex.RLock()
	defer globalCachesMutex.RUnlock()

	// Create a wait group to update all caches concurrently
	var wg sync.WaitGroup
	wg.Add(len(globalCaches))

	// Update all caches concurrently
	for _, cache := range globalCaches {
		go func(c interface{ Update() }) {
			defer wg.Done()
			c.Update()
		}(cache)
	}

	// Wait for all updates to complete
	wg.Wait()
}
