package pantry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Item[T any] struct {
	Value   T
	Expires int64
}

type Pantry[T any] struct {
	store   map[string]Item[T]
	mutex   sync.RWMutex
	close   chan struct{}
	options Options
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

func (pantry *Pantry[T]) Set(key string, value T, expiration time.Duration) *Result[T] {
	item := Item[T]{
		Value:   value,
		Expires: time.Now().Add(expiration).UnixNano(),
	}
	pantry.mutex.Lock()
	pantry.store[key] = item
	pantry.mutex.Unlock()

	return &Result[T]{
		action: "set",
		key:    key,
		item:   item,
		pantry: pantry,
	}
}

func (pantry *Pantry[T]) Remove(key string) *Result[T] {
	pantry.mutex.Lock()
	delete(pantry.store, key)
	pantry.mutex.Unlock()

	return &Result[T]{
		action: "remove",
		key:    key,
		pantry: pantry,
	}
}

func (pantry *Pantry[T]) IsEmpty() bool {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()
	return len(pantry.store) == 0
}

func (pantry *Pantry[T]) Close() {
	pantry.close <- struct{}{}
	pantry.mutex.Lock()
	pantry.store = make(map[string]Item[T])
	pantry.mutex.Unlock()
}

func (pantry *Pantry[T]) Load() error {
	directory := pantry.options.PersistenceDirectory

	if directory == "" {
		return nil
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, 0755); err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, f := range files {
		fullPath := fmt.Sprintf("%s/%s", directory, f.Name())
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return err
		}

		buffer := bytes.NewBuffer(content)
		decoder := gob.NewDecoder(buffer)

		var item Item[T]
		if err := decoder.Decode(&item); err != nil {
			return err
		}

		pantry.mutex.Lock()
		pantry.store[f.Name()] = item
		pantry.mutex.Unlock()
	}

	return nil
}

func New[T any](options *Options) *Pantry[T] {
	finalOptions := Options{
		CleaningInterval:     options.CleaningInterval,
		PersistenceDirectory: options.PersistenceDirectory,
	}

	if options.CleaningInterval == 0 {
		finalOptions.CleaningInterval = time.Minute
	}

	pantry := &Pantry[T]{
		store:   make(map[string]Item[T]),
		mutex:   sync.RWMutex{},
		options: finalOptions,
		close:   make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(pantry.options.CleaningInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pantry.mutex.Lock()
				for key, item := range pantry.store {
					if time.Now().UnixNano() > item.Expires {
						delete(pantry.store, key)

						if options.PersistenceDirectory != "" {
							fileName := fmt.Sprintf("%s/%s",
								pantry.options.PersistenceDirectory, key)
							os.Remove(fileName)
						}
					}
				}
				pantry.mutex.Unlock()

			case <-pantry.close:
				return
			}
		}
	}()

	return pantry
}
