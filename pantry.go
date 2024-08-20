package pantry

import (
	"context"
	"iter"
	"sync"
	"time"
)

type item[T any] struct {
	value   T
	expires int64
}

type Pantry[T any] struct {
	expiration time.Duration
	store      map[string]item[T]
	mutex      sync.RWMutex
}

func (pantry *Pantry[T]) Get(key string) (T, bool) {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	item, found := pantry.store[key]
	if found && time.Now().UnixNano() > item.expires {
		return *new(T), false
	}
	return item.value, found
}

func (pantry *Pantry[T]) Set(key string, value T) {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	pantry.store[key] = item[T]{
		value:   value,
		expires: time.Now().Add(pantry.expiration).UnixNano(),
	}
}

func (pantry *Pantry[T]) Remove(key string) {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	delete(pantry.store, key)
}

func (pantry *Pantry[T]) IsEmpty() bool {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	return len(pantry.store) == 0
}

func (pantry *Pantry[T]) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		pantry.mutex.RLock()
		defer pantry.mutex.RUnlock()

		for key, item := range pantry.store {
			if time.Now().UnixNano() > item.expires {
				continue
			}

			if !yield(key) {
				return
			}
		}
	}
}

func (pantry *Pantry[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		pantry.mutex.RLock()
		defer pantry.mutex.RUnlock()

		for _, item := range pantry.store {
			if time.Now().UnixNano() > item.expires {
				continue
			}

			if !yield(item.value) {
				return
			}
		}
	}
}

func (pantry *Pantry[T]) All() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		pantry.mutex.RLock()
		defer pantry.mutex.RUnlock()

		for key, item := range pantry.store {
			if time.Now().UnixNano() > item.expires {
				continue
			}

			if !yield(key, item.value) {
				return
			}
		}
	}
}

func New[T any](ctx context.Context, expiration time.Duration) *Pantry[T] {
	pantry := &Pantry[T]{
		expiration: expiration,
		store:      make(map[string]item[T]),
		mutex:      sync.RWMutex{},
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pantry.mutex.Lock()
				for key, item := range pantry.store {
					if time.Now().UnixNano() > item.expires {
						delete(pantry.store, key)
					}
				}
				pantry.mutex.Unlock()

			case <-ctx.Done():
				pantry.mutex.Lock()
				pantry.store = make(map[string]item[T])
				pantry.mutex.Unlock()
				return
			}
		}
	}()

	return pantry
}
