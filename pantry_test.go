package pantry

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestEmptyOptions(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	if p.options.DatabasePath != "" {
		t.Fatal("database name is not empty")
	}

	if p.options.CleaningInterval != time.Minute {
		t.Fatalf("cleaning interval is not 1 minute: %s", p.options.CleaningInterval)
	}
}

func TestDatabasePathOption(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if p.options.DatabasePath != "test.db" {
		t.Fatal("database name is not set")
	}
}

func TestCleaning(t *testing.T) {
	p := New(&Options{
		CleaningInterval: 3 * time.Millisecond,
	})
	defer p.Close()

	if p.options.CleaningInterval != 3*time.Millisecond {
		t.Fatalf("cleaning interval is not set correctly: %s", p.options.CleaningInterval)
	}

	err := p.Set("test", "hello", time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	_, found := p.Get("test")
	if !found {
		t.Fatal("not found")
	}

	time.Sleep(2 * time.Millisecond)

	_, found = p.Get("test")
	if found {
		t.Fatal("found")
	}

	time.Sleep(2 * time.Millisecond)

	_, found = p.Get("test")
	if found {
		t.Fatal("found")
	}
}

func TestIsEmtpy(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	if !p.IsEmpty() {
		t.Log(p.store)
		t.Fatal("not empty")
	}
}

func TestCorruptedDatabase(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	err := os.WriteFile(p.options.DatabasePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = p.Load()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseDoesntExists(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	if err := p.Set("test", "hello", time.Hour); err != nil {
		t.Fatal(err)
	}

	_, found := p.Get("test")
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if err := p.Remove("test"); err != nil {
		t.Fatal(err)
	}

	_, found = p.Get("test")
	if found {
		t.Log(p.store)
		t.Fatal("found")
	}
}

func TestRemovePersisted(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set("test", "hello", time.Hour); err != nil {
		t.Fatal(err)
	}

	_, found := p.Get("test")
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if err := p.Remove("test"); err != nil {
		t.Fatal(err)
	}

	_, found = p.Get("test")
	if found {
		t.Log(p.store)
		t.Fatal("found")
	}
}

func TestGetAll(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	if err := p.Set("first", "1", time.Hour); err != nil {
		t.Fatal(err)
	}

	if err := p.Set("second", "2", time.Hour); err != nil {
		t.Fatal(err)
	}

	if err := p.Set("third", "3", time.Hour); err != nil {
		t.Fatal(err)
	}

	values := p.GetAll()
	if len(values) != 3 {
		t.Log(values)
		t.Fatal("not 3 items")
	}
}

func TestString(t *testing.T) {
	key := "string"
	value := "hello"

	p := New(&Options{
		DatabasePath: "test.db",
	})

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set(key, value, time.Hour); err != nil {
		t.Fatal(err)
	}

	v, found := p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.String {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted := v.(string)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}

	p.Close()

	p = New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	v, found = p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.String {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted = v.(string)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}
}

func TestInt(t *testing.T) {
	key := "int"
	value := 42

	p := New(&Options{
		DatabasePath: "test.db",
	})

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set(key, value, time.Hour); err != nil {
		t.Fatal(err)
	}

	v, found := p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Int {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted := v.(int)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}

	p.Close()

	p = New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	v, found = p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Int {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted = v.(int)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}
}

func TestFloat(t *testing.T) {
	key := "int"
	value := 3.14

	p := New(&Options{
		DatabasePath: "test.db",
	})

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set(key, value, time.Hour); err != nil {
		t.Fatal(err)
	}

	v, found := p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Float64 {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted := v.(float64)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}

	p.Close()

	p = New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	v, found = p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Float64 {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted = v.(float64)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}
}

func TestStruct(t *testing.T) {
	type TestData struct {
		Text   string
		Number int
	}

	key := "int"
	value := TestData{Text: "test", Number: 42}

	p := New(&Options{
		DatabasePath: "test.db",
		CustomType:   TestData{},
	})

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := p.Set(key, value, time.Hour); err != nil {
		t.Fatal(err)
	}

	v, found := p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Struct {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted := v.(TestData)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}

	p.Close()

	p = New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	v, found = p.Get(key)
	if !found {
		t.Log(p.store)
		t.Fatal("not found")
	}

	if reflect.TypeOf(v).Kind() != reflect.Struct {
		t.Log(v)
		t.Fatalf("invalid type: %s", reflect.TypeOf(v))
	}

	casted = v.(TestData)

	if casted != value {
		t.Log(p.store)
		t.Fatal("invalid value")
	}
}
