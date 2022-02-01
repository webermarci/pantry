package pantry

import (
	"os"
	"testing"
	"time"
)

func TestEmptyOptions(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	if p.options.DatabasePath != "" {
		t.Error("database name is not empty")
		return
	}

	if p.options.CleaningInterval != time.Minute {
		t.Errorf("cleaning interval is not 1 minute: %s", p.options.CleaningInterval)
		return
	}
}

func TestDatabasePathOption(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	if p.options.DatabasePath != "test.db" {
		t.Error("database name is not set")
		return
	}
}

func TestGetSetRemove(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	value, found := p.Get("test")
	if found {
		t.Errorf("value found: %s", value)
		return
	}

	err := p.Set("test", "hello", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	value, found = p.Get("test")
	if !found {
		t.Error("value not found")
		return
	}

	if value != "hello" {
		t.Errorf("not the correct value found: %s", value)
		return
	}

	err = p.Remove("test")
	if err != nil {
		t.Error(err)
		return
	}

	value, found = p.Get("test")
	if found {
		t.Errorf("value found: %s", value)
		return
	}
}

func TestSetGetAll(t *testing.T) {
	p := New(&Options{})
	defer p.Close()

	err := p.Set("test1", "hello1", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	err = p.Set("test2", "hello2", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	err = p.Set("test3", "hello3", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	items := p.GetAll()

	if len(items) != 3 {
		t.Errorf("not 3 items found: %d", len(items))
		return
	}
}

func TestCleaning(t *testing.T) {
	p := New(&Options{
		CleaningInterval: 20 * time.Millisecond,
	})
	defer p.Close()

	if p.options.CleaningInterval != 20*time.Millisecond {
		t.Errorf("cleaning interval is not set correctly: %s", p.options.CleaningInterval)
		return
	}

	err := p.Set("test", "hello", 10*time.Millisecond)
	if err != nil {
		t.Error(err)
		return
	}

	_, found := p.Get("test")
	if !found {
		t.Error("value not found")
		return
	}

	time.Sleep(15 * time.Millisecond)

	_, found = p.Get("test")
	if found {
		t.Error("value found")
		return
	}

	if len(p.GetAll()) != 0 {
		t.Error("values found")
		return
	}

	time.Sleep(15 * time.Millisecond)

	_, found = p.Get("test")
	if found {
		t.Error("value found")
		return
	}
}

func TestPersistence(t *testing.T) {
	p1 := New(&Options{
		DatabasePath: "test.db",
	})
	defer p1.Close()

	defer func() {
		err := os.Remove(p1.options.DatabasePath)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}()

	err := p1.Load()
	if err != nil {
		t.Error(err)
		return
	}

	value, found := p1.Get("test")
	if found {
		t.Errorf("value found: %s", value)
		return
	}

	err = p1.Set("test", "hello", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	err = p1.Set("remove", "this", time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	err = p1.Remove("remove")
	if err != nil {
		t.Error(err)
		return
	}

	p2 := New(&Options{
		DatabasePath: "test.db",
	})
	defer p2.Close()

	err = p2.Load()
	if err != nil {
		t.Error(err)
		return
	}

	_, found = p2.Get("test")
	if !found {
		t.Errorf("value not found")
		return
	}
}

func TestCorruptedDatabase(t *testing.T) {
	p := New(&Options{
		DatabasePath: "test.db",
	})
	defer p.Close()

	err := os.WriteFile(p.options.DatabasePath, []byte("hello"), 0644)
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		err := os.Remove(p.options.DatabasePath)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}()

	err = p.Load()
	if err != nil {
		t.Error(err)
		return
	}
}
