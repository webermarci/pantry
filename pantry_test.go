package pantry

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestCleaning(t *testing.T) {
	p := New[string]()
	defer p.Close()

	p.Set("test", "hello", 100*time.Millisecond)

	if _, found := p.Get("test"); !found {
		t.Fatal("not found")
	}

	time.Sleep(3 * time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}

	time.Sleep(3 * time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}
}

func TestCleaningPersisted(t *testing.T) {
	p := New[string]().WithPersistence(t.Name())
	defer p.Close()

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set("test", "hello", 100*time.Millisecond).Persist(); err != nil {
		t.Fatal(err)
	}

	if _, found := p.Get("test"); !found {
		t.Fatal("not found")
	}

	time.Sleep(3 * time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}

	time.Sleep(3 * time.Second)

	if _, found := p.Get("test"); found {
		t.Fatal("found")
	}
}

func TestIsEmtpy(t *testing.T) {
	p := New[string]()
	defer p.Close()

	if !p.IsEmpty() {
		t.Log(p.store)
		t.Fatal("not empty")
	}
}

func TestLoadWithoutSetDirectory(t *testing.T) {
	p := New[string]()
	defer p.Close()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadWithMissingDirectory(t *testing.T) {
	p := New[string]().WithPersistence(t.Name())
	defer p.Close()

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestSet(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string]()
	defer p.Close()

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	p.Set(key, value, time.Hour)

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}
}

func TestSetPersisted(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string]().WithPersistence(t.Name())

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	if err := p.Set(key, value, time.Hour).Persist(); err != nil {
		t.Log(p.store)
		t.Fatal(err)
	}

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	p.Close()

	p = New[string]().WithPersistence(t.Name())

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	p.Close()
}

func TestSetPersistedStruct(t *testing.T) {
	type TestData struct {
		Number int
		Text   string
	}

	key := "test"
	value := TestData{Number: 42, Text: "hello"}

	p := New[TestData]().WithPersistence(t.Name())

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	if err := p.Set(key, value, time.Hour).Persist(); err != nil {
		t.Log(p.store)
		t.Fatal(err)
	}

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	p.Close()

	p = New[TestData]().WithPersistence(t.Name())

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	v, found := p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if v.Number != 42 {
		t.Fatal("not 42")
	}

	p.Close()
}

func TestRemove(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string]()
	defer p.Close()

	p.Set(key, value, time.Hour)

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

func TestRemovePersisted(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string]().WithPersistence(t.Name())

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set(key, value, time.Hour).Persist(); err != nil {
		t.Fatal(err)
	}

	if _, found := p.Get(key); !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if err := p.Remove(key).Persist(); err != nil {
		t.Fatal(err)
	}

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	p.Close()

	p = New[string]().WithPersistence(t.Name())

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	p.Close()
}

func TestGetAll(t *testing.T) {
	p := New[int]()
	defer p.Close()

	p.Set("first", 1, time.Hour)
	p.Set("second", 2, time.Hour)
	p.Set("third", 3, time.Hour)

	values := p.GetAll()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}
}

func TestGetAllIgnoreExpired(t *testing.T) {
	p := New[int]()
	defer p.Close()

	p.Set("first", 1, 10*time.Millisecond)
	p.Set("second", 2, 10*time.Millisecond)
	p.Set("third", 3, 10*time.Millisecond)

	values := p.GetAll()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}

	time.Sleep(20 * time.Millisecond)

	values = p.GetAll()
	if len(values) != 0 {
		t.Log(values)
		t.Fatal("not ignored")
	}
}

func TestGetAllPersisted(t *testing.T) {
	p := New[int]().WithPersistence(t.Name())

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set("first", 1, time.Hour).Persist(); err != nil {
		t.Fatal(err)
	}

	if err := p.Set("second", 2, time.Hour).Persist(); err != nil {
		t.Fatal(err)
	}

	if err := p.Set("third", 3, time.Hour).Persist(); err != nil {
		t.Fatal(err)
	}

	values := p.GetAll()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}

	p.Close()

	p = New[int]().WithPersistence(t.Name())

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	values = p.GetAll()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}
}

func TestGetAllFlat(t *testing.T) {
	p := New[int]()
	defer p.Close()

	p.Set("first", 1, time.Hour)
	p.Set("second", 2, time.Hour)
	p.Set("third", 3, time.Hour)

	values := p.GetAllFlat()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}
}

func TestInvalidResultAction(t *testing.T) {
	p := New[string]().WithPersistence(t.Name())
	defer p.Close()

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	result := Result[string]{
		action: "invalid",
		pantry: p,
	}

	if err := result.Persist(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPersistingWithoutPersistencePath(t *testing.T) {
	p := New[string]()
	defer p.Close()

	if err := p.Set("test", "hello", time.Second).Persist(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCleanExpiredPersisted(t *testing.T) {
	key := "test"
	value := "hello"

	p := New[string]().WithPersistence(t.Name())

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}

	if err := p.Set(key, value, 2*time.Millisecond).Persist(); err != nil {
		t.Log(p.store)
		t.Fatal(err)
	}

	time.Sleep(3 * time.Millisecond)

	if _, found := p.Get(key); found {
		t.Log(p.store)
		t.Fatal("found")
	}
}

func BenchmarkCache(b *testing.B) {
	p := New[int]()
	defer p.Close()

	for i := 0; i < b.N; i++ {
		v := strconv.Itoa(i)
		p.Set(v, i, time.Hour)
	}
}

func BenchmarkPersisted(b *testing.B) {
	p := New[int]().WithPersistence(b.Name())
	defer p.Close()

	defer func() {
		err := os.RemoveAll(p.persistencePath)
		if err != nil {
			b.Fatal(err)
		}
	}()

	for i := 0; i < b.N; i++ {
		v := strconv.Itoa(i)
		if err := p.Set(v, i, time.Hour).Persist(); err != nil {
			b.Fatal(err)
		}
	}
}
