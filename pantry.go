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

type Item struct {
	Value   interface{}
	Expires int64
}

type Pantry struct {
	store   map[string]Item
	mutex   sync.RWMutex
	close   chan struct{}
	options Options
}

func (pantry *Pantry) Type(v interface{}) *Pantry {
	if pantry.options.PersistenceDirectory != "" {
		gob.Register(v)
	}
	return pantry
}

func (pantry *Pantry) Get(key string) (interface{}, bool) {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	item, found := pantry.store[key]
	if found && time.Now().UnixNano() > item.Expires {
		return "", false
	}
	return item.Value, found
}

func (pantry *Pantry) GetAll() map[string]interface{} {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	items := make(map[string]interface{})
	for key, item := range pantry.store {
		if time.Now().UnixNano() > item.Expires {
			continue
		}
		items[key] = item.Value
	}
	return items
}

func (pantry *Pantry) Set(key string, value interface{}, expiration time.Duration) *Result {
	item := Item{
		Value:   value,
		Expires: time.Now().Add(expiration).UnixNano(),
	}
	pantry.mutex.Lock()
	pantry.store[key] = item
	pantry.mutex.Unlock()

	return &Result{
		action: "set",
		key:    key,
		item:   item,
		pantry: pantry,
	}
}

func (pantry *Pantry) Remove(key string) *Result {
	pantry.mutex.Lock()
	delete(pantry.store, key)
	pantry.mutex.Unlock()

	return &Result{
		action: "remove",
		key:    key,
		pantry: pantry,
	}
}

func (pantry *Pantry) IsEmpty() bool {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()
	return len(pantry.store) == 0
}

func (pantry *Pantry) Close() {
	pantry.close <- struct{}{}
	pantry.mutex.Lock()
	pantry.store = make(map[string]Item)
	pantry.mutex.Unlock()
}

func (pantry *Pantry) Load() error {
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

		var item Item
		if err := decoder.Decode(&item); err != nil {
			return err
		}

		pantry.mutex.Lock()
		pantry.store[f.Name()] = item
		pantry.mutex.Unlock()
	}

	return nil
}

func New(options *Options) *Pantry {
	finalOptions := Options{
		CleaningInterval:     options.CleaningInterval,
		PersistenceDirectory: options.PersistenceDirectory,
	}

	if options.CleaningInterval == 0 {
		finalOptions.CleaningInterval = time.Minute
	}

	pantry := &Pantry{
		store:   make(map[string]Item),
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
