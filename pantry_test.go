package pantry

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	New[string](ctx, time.Hour)
	cancel()
}

func TestCleaning(t *testing.T) {
	p := New[string](context.Background(), 100*time.Millisecond)

	p.Set("test", "hello")

	if _, found := p.Get("test"); !found {
		t.Fatal("not found")
	}

	time.Sleep(5 * time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}

	time.Sleep(time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}
}

func TestIsEmtpy(t *testing.T) {
	p := New[string](context.Background(), time.Hour)

	if !p.IsEmpty() {
		t.Log(p.store)
		t.Fatal("not empty")
	}
}

func TestSet(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string](context.Background(), time.Hour)

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	p.Set(key, value)

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}
}

func TestRemove(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string](context.Background(), time.Hour)

	p.Set(key, value)

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	p.Remove(key)

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}
}

func TestGetIgnoreExpired(t *testing.T) {
	p := New[int](context.Background(), 10*time.Millisecond)

	p.Set(t.Name(), 1)

	_, found := p.Get(t.Name())
	if !found {
		t.Fatal("not found")
	}

	time.Sleep(20 * time.Millisecond)

	_, found = p.Get(t.Name())
	if found {
		t.Fatal("not ignored")
	}
}

func TestValues(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for value := range p.Values() {
		t.Log(value)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}
}

func TestValuesBreak(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	for value := range p.Values() {
		t.Log(value)
		break
	}
}

func TestValuesIgnoreExpired(t *testing.T) {
	p := New[int](context.Background(), 10*time.Millisecond)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for value := range p.Values() {
		t.Log(value)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}

	time.Sleep(20 * time.Millisecond)

	counter = 0

	for value := range p.Values() {
		t.Log(value)
		counter++
	}

	if counter != 0 {
		t.Fatal("not ignored")
	}
}

func TestKeys(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for key := range p.Keys() {
		t.Log(key)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}
}

func TestKeysBreak(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	for key := range p.Keys() {
		t.Log(key)
		break
	}
}

func TestKeysIgnoreExpired(t *testing.T) {
	p := New[int](context.Background(), 10*time.Millisecond)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for key := range p.Keys() {
		t.Log(key)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}

	time.Sleep(20 * time.Millisecond)

	counter = 0

	for key := range p.Keys() {
		t.Log(key)
		counter++
	}

	if counter != 0 {
		t.Fatal("not ignored")
	}
}

func TestAll(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for key, value := range p.All() {
		t.Log(key, value)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}
}

func TestAllBreak(t *testing.T) {
	p := New[int](context.Background(), time.Hour)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	for key, value := range p.All() {
		t.Log(key, value)
		break
	}
}

func TestAllIgnoreExpired(t *testing.T) {
	p := New[int](context.Background(), 10*time.Millisecond)

	p.Set("first", 1)
	p.Set("second", 2)
	p.Set("third", 3)

	counter := 0

	for key, value := range p.All() {
		t.Log(key, value)
		counter++
	}

	if counter != 3 {
		t.Fatal("not 3 items")
	}

	time.Sleep(20 * time.Millisecond)

	counter = 0

	for key, value := range p.All() {
		t.Log(key, value)
		counter++
	}

	if counter != 0 {
		t.Fatal("not ignored")
	}
}

func BenchmarkGet(b *testing.B) {
	p := New[int](context.Background(), time.Hour)

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		p.Set(key, i)

		b.Run(key, func(b *testing.B) {
			p.Get(key)
		})
	}
}

func BenchmarkSet(b *testing.B) {
	p := New[int](context.Background(), time.Hour)

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)

		b.Run(key, func(b *testing.B) {
			p.Set(key, i)
		})
	}
}
