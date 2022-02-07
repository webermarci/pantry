package pantry

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Options struct {
	CleaningInterval time.Duration
	DatabasePath     string
}

type Pantry struct {
	store      map[string]Item
	mutex      sync.RWMutex
	options    Options
	registered map[string]bool
	close      chan struct{}
}

type Item struct {
	Value   interface{}
	Expires int64
}

type Result struct {
	pantry *Pantry
}

func (result *Result) Save() error {
	return result.pantry.Save()
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
	pantry.mutex.Lock()
	pantry.store[key] = Item{
		Value:   value,
		Expires: time.Now().Add(expiration).UnixNano(),
	}
	pantry.mutex.Unlock()

	return &Result{
		pantry: pantry,
	}
}

func (pantry *Pantry) Remove(key string) *Result {
	pantry.mutex.Lock()
	delete(pantry.store, key)
	pantry.mutex.Unlock()

	return &Result{
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
	if fileExists(pantry.options.DatabasePath) {
		content, err := ioutil.ReadFile(pantry.options.DatabasePath)
		if err != nil {
			return err
		}

		buffer := bytes.NewBuffer(content)
		decoder := gob.NewDecoder(buffer)

		pantry.mutex.Lock()
		err = decoder.Decode(&pantry.store)
		pantry.mutex.Unlock()

		if err != nil {
			return err
		}
		return nil
	} else {
		return pantry.Save()
	}
}

func (pantry *Pantry) Save() error {
	pantry.mutex.RLock()
	defer pantry.mutex.RUnlock()

	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(pantry.store); err != nil {
		return err
	}
	return os.WriteFile(pantry.options.DatabasePath, buffer.Bytes(), 0644)
}

func (pantry *Pantry) Type(v interface{}) *Pantry {
	if pantry.options.DatabasePath != "" {
		gob.Register(v)
	}
	return pantry
}

func New(options *Options) *Pantry {
	finalOptions := Options{
		CleaningInterval: options.CleaningInterval,
		DatabasePath:     options.DatabasePath,
	}

	if options.CleaningInterval == 0 {
		finalOptions.CleaningInterval = time.Minute
	}

	pantry := &Pantry{
		store:      make(map[string]Item),
		mutex:      sync.RWMutex{},
		options:    finalOptions,
		registered: make(map[string]bool),
		close:      make(chan struct{}),
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
