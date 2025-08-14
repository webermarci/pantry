// Package pantry provides a thread-safe, in-memory key-value store for Go with expiring items.
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

// Pantry is a thread-safe, in-memory key-value store with expiring items
type Pantry[T any] struct {
	expiration time.Duration
	store      map[string]item[T]
	mutex      sync.RWMutex
}

// Get retrieves a value from the pantry. If the item has expired, it will be removed and `false` will be returned.
func (pantry *Pantry[T]) Get(key string) (T, bool) {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	item, found := pantry.store[key]
	if found && time.Now().UnixNano() > item.expires {
		return *new(T), false
	}
	return item.value, found
}

// Set adds a value to the pantry. The item will expire after the default expiration time.
func (pantry *Pantry[T]) Set(key string, value T) {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	pantry.store[key] = item[T]{
		value:   value,
		expires: time.Now().Add(pantry.expiration).UnixNano(),
	}
}

// Remove removes a value from the pantry.
func (pantry *Pantry[T]) Remove(key string) {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	delete(pantry.store, key)
}

// IsEmpty returns `true` if the pantry is empty.
func (pantry *Pantry[T]) IsEmpty() bool {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	return len(pantry.store) == 0
}

// Contains returns `true` if the key exists in the pantry.
func (pantry *Pantry[T]) Contains(key string) bool {
	_, found := pantry.Get(key)
	return found
}

// Count returns the number of items in the pantry.
func (pantry *Pantry[T]) Count() int {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	return len(pantry.store)
}

// Clear removes all items from the pantry.
func (pantry *Pantry[T]) Clear() {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	pantry.store = make(map[string]item[T])
}

// Keys returns an iterator over the keys in the pantry.
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

// Values returns an iterator over the values in the pantry.
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

// All returns an iterator over the keys and values in the pantry.
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

// New creates a new pantry. The expiration duration is the time-to-live for items.
// The context can be used to gracefully shutdown the pantry and free up resources.
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
