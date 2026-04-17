package pantry

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"
)

type Observer interface {
	OnHit(key any)
	OnMiss(key any)
	OnSet(key any, ttl time.Duration)
	OnRemove(key any)
	OnEvict(key any)
	OnStorageError(err error, op string)
}

type Persister[K comparable, T any] interface {
	Save(key K, value T) error
	Delete(key K) error
	Load() (map[K]T, error)
}

type Snapshotter[K comparable, T any] interface {
	Save(data map[K]T) error
	Load() (map[K]T, error)
}

type item[T any] struct {
	value   T
	expires int64
}

type shard[K comparable, T any] struct {
	sync.RWMutex
	store map[K]item[T]
}

type setConfig struct {
	ttl time.Duration
}

type Pantry[K comparable, T any] struct {
	shards           []shard[K, T]
	mask             uint32
	defaultTTL       time.Duration
	janitorInterval  time.Duration
	snapshotInterval time.Duration
	observer         Observer
	persister        Persister[K, T]
	snapshotter      Snapshotter[K, T]
	wg               sync.WaitGroup
}

type PantryOption[K comparable, T any] func(*Pantry[K, T])
type SetOption func(*setConfig)

func WithDefaultTTL[K comparable, T any](ttl time.Duration) PantryOption[K, T] {
	return func(p *Pantry[K, T]) { p.defaultTTL = ttl }
}

func WithJanitorInterval[K comparable, T any](d time.Duration) PantryOption[K, T] {
	return func(p *Pantry[K, T]) { p.janitorInterval = d }
}

func WithShards[K comparable, T any](count int) PantryOption[K, T] {
	return func(p *Pantry[K, T]) {
		actualCount := nextPowerOfTwo(count)
		p.shards = make([]shard[K, T], actualCount)
		p.mask = uint32(actualCount - 1)
		for i := range actualCount {
			p.shards[i].store = make(map[K]item[T])
		}
	}
}

func WithPersistence[K comparable, T any](pers Persister[K, T]) PantryOption[K, T] {
	return func(p *Pantry[K, T]) { p.persister = pers }
}

func WithSnapshotting[K comparable, T any](s Snapshotter[K, T], interval time.Duration) PantryOption[K, T] {
	return func(p *Pantry[K, T]) {
		p.snapshotter = s
		p.snapshotInterval = interval
	}
}

func WithObserver[K comparable, T any](o Observer) PantryOption[K, T] {
	return func(p *Pantry[K, T]) { p.observer = o }
}

func WithTTL(d time.Duration) SetOption {
	return func(c *setConfig) { c.ttl = d }
}

func New[K comparable, T any](ctx context.Context, opts ...PantryOption[K, T]) *Pantry[K, T] {
	p := &Pantry[K, T]{
		defaultTTL:       0,
		janitorInterval:  5 * time.Second,
		snapshotInterval: 5 * time.Minute,
	}

	WithShards[K, T](16)(p)
	for _, opt := range opts {
		opt(p)
	}

	hydrate := func(data map[K]T) {
		for k, v := range data {
			s := p.getShard(k)
			s.Lock()
			s.store[k] = item[T]{value: v, expires: 0}
			s.Unlock()
		}
	}

	if p.snapshotter != nil {
		if data, err := p.snapshotter.Load(); err == nil {
			hydrate(data)
		}
		p.wg.Add(1)
		go p.snapshotLoop(ctx)
	}

	if p.persister != nil {
		if data, err := p.persister.Load(); err == nil {
			hydrate(data)
		}
	}

	p.wg.Add(1)
	go p.janitor(ctx)

	return p
}

func (p *Pantry[K, T]) Wait() { p.wg.Wait() }

func (p *Pantry[K, T]) Get(key K) (T, bool) {
	s := p.getShard(key)
	s.RLock()
	it, found := s.store[key]
	s.RUnlock()

	if found && it.expires > 0 && time.Now().UnixNano() > it.expires {
		if p.observer != nil {
			p.observer.OnMiss(key)
		}
		return *new(T), false
	}

	if p.observer != nil {
		if found {
			p.observer.OnHit(key)
		} else {
			p.observer.OnMiss(key)
		}
	}
	return it.value, found
}

func (p *Pantry[K, T]) Set(key K, value T, opts ...SetOption) {
	cfg := setConfig{ttl: p.defaultTTL}
	for _, opt := range opts {
		opt(&cfg)
	}

	s := p.getShard(key)
	s.Lock()
	defer s.Unlock()

	var exp int64
	if cfg.ttl > 0 {
		exp = time.Now().Add(cfg.ttl).UnixNano()
	}

	s.store[key] = item[T]{value: value, expires: exp}

	if p.observer != nil {
		p.observer.OnSet(key, cfg.ttl)
	}
	if p.persister != nil {
		if err := p.persister.Save(key, value); err != nil && p.observer != nil {
			p.observer.OnStorageError(err, "persister_save")
		}
	}
}

func (p *Pantry[K, T]) Update(key K, fn func(val T, exists bool) T) {
	s := p.getShard(key)
	s.Lock()
	defer s.Unlock()

	it, found := s.store[key]
	now := time.Now().UnixNano()
	if found && it.expires > 0 && now > it.expires {
		found = false
	}

	newVal := fn(it.value, found)
	var exp int64
	if found {
		exp = it.expires
	} else if p.defaultTTL > 0 {
		exp = now + p.defaultTTL.Nanoseconds()
	}

	s.store[key] = item[T]{value: newVal, expires: exp}

	if p.observer != nil {
		p.observer.OnSet(key, time.Duration(exp-now))
	}
	if p.persister != nil {
		if err := p.persister.Save(key, newVal); err != nil && p.observer != nil {
			p.observer.OnStorageError(err, "persister_update")
		}
	}
}

func (p *Pantry[K, T]) Remove(key K) {
	s := p.getShard(key)
	s.Lock()
	delete(s.store, key)
	s.Unlock()

	if p.observer != nil {
		p.observer.OnRemove(key)
	}
	if p.persister != nil {
		if err := p.persister.Delete(key); err != nil && p.observer != nil {
			p.observer.OnStorageError(err, "persister_delete")
		}
	}
}

func (p *Pantry[K, T]) All() iter.Seq2[K, T] {
	return func(yield func(K, T) bool) {
		now := time.Now().UnixNano()
		for i := range p.shards {
			s := &p.shards[i]
			s.RLock()
			for k, it := range s.store {
				if it.expires > 0 && now > it.expires {
					continue
				}
				if !yield(k, it.value) {
					s.RUnlock()
					return
				}
			}
			s.RUnlock()
		}
	}
}

func (p *Pantry[K, T]) getShard(key K) *shard[K, T] {
	var hash uint32 = 2166136261
	switch v := any(key).(type) {
	case string:
		for i := 0; i < len(v); i++ {
			hash ^= uint32(v[i])
			hash *= 16777619
		}
	case int:
		hash ^= uint32(v)
		hash *= 16777619
	case uint32:
		hash ^= v
		hash *= 16777619
	default:
		s := fmt.Sprintf("%v", key)
		for i := 0; i < len(s); i++ {
			hash ^= uint32(s[i])
			hash *= 16777619
		}
	}
	return &p.shards[hash&p.mask]
}

func (p *Pantry[K, T]) janitor(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(p.janitorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().UnixNano()
			for i := range p.shards {
				p.shards[i].Lock()
				for k, it := range p.shards[i].store {
					if it.expires > 0 && now > it.expires {
						delete(p.shards[i].store, k)
						if p.observer != nil {
							p.observer.OnEvict(k)
						}
						if p.persister != nil {
							_ = p.persister.Delete(k)
						}
					}
				}
				p.shards[i].Unlock()
			}
		}
	}
}

func (p *Pantry[K, T]) snapshotLoop(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(p.snapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			p.saveSnapshot()
			return
		case <-ticker.C:
			p.saveSnapshot()
		}
	}
}

func (p *Pantry[K, T]) saveSnapshot() {
	if p.snapshotter == nil {
		return
	}
	data := make(map[K]T)
	now := time.Now().UnixNano()
	for i := range p.shards {
		p.shards[i].RLock()
		for k, it := range p.shards[i].store {
			if it.expires == 0 || now < it.expires {
				data[k] = it.value
			}
		}
		p.shards[i].RUnlock()
	}
	if err := p.snapshotter.Save(data); err != nil && p.observer != nil {
		p.observer.OnStorageError(err, "snapshot_save")
	}
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
