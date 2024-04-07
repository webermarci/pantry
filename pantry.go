package pantry

import (
	"context"
	"sync"
	"time"
)

type Item[T any] struct {
	Value   T
	Expires int64
}

type Pantry[T any] struct {
	expiration time.Duration
	store      map[string]Item[T]
	mutex      sync.RWMutex
}

func (pantry *Pantry[T]) Get(key string) (T, bool) {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	item, found := pantry.store[key]
	if found && time.Now().UnixNano() > item.Expires {
		return *new(T), false
	}
	return item.Value, found
}

func (pantry *Pantry[T]) GetAll() map[string]T {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	items := make(map[string]T)
	for key, item := range pantry.store {
		if time.Now().UnixNano() > item.Expires {
			continue
		}
		items[key] = item.Value
	}
	return items
}

func (pantry *Pantry[T]) GetAllFlat() []T {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	items := []T{}
	for _, item := range pantry.store {
		if time.Now().UnixNano() > item.Expires {
			continue
		}
		items = append(items, item.Value)
	}
	return items
}

func (pantry *Pantry[T]) Set(key string, value T) {
	pantry.mutex.Lock()
	defer pantry.mutex.Unlock()

	pantry.store[key] = Item[T]{
		Value:   value,
		Expires: time.Now().Add(pantry.expiration).UnixNano(),
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

func New[T any](ctx context.Context, expiration time.Duration) *Pantry[T] {
	pantry := &Pantry[T]{
		expiration: expiration,
		store:      make(map[string]Item[T]),
		mutex:      sync.RWMutex{},
	}

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for key, item := range pantry.store {
					if time.Now().UnixNano() > item.Expires {
						pantry.Remove(key)
					}
				}

			case <-ctx.Done():
				pantry.mutex.Lock()
				pantry.store = make(map[string]Item[T])
				pantry.mutex.Unlock()
				return
			}
		}
	}()

	return pantry
}
