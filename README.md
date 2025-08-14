# Pantry

[![Go Reference](https://pkg.go.dev/badge/github.com/webermarci/pantry.svg)](https://pkg.go.dev/github.com/webermarci/pantry)
[![Test](https://github.com/webermarci/pantry/actions/workflows/testing.yml/badge.svg)](https://github.com/webermarci/pantry/actions/workflows/testing.yml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Pantry is a generic, thread-safe, in-memory key-value store for Go with expiring
items.

## Features

- Thread-safe for concurrent use
- Generic, works with any type
- Items expire automatically
- Graceful shutdown with context cancellation
- Iterators for keys, values, and all items

## Installation

```bash
go get github.com/webermarci/pantry
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/webermarci/pantry"
)

func main() {
	// Create a new pantry with a default expiration of 1 hour.
	p := pantry.New[string](context.Background(), time.Hour)

	// Set a value with the default expiration.
	p.Set("hello", "world")

	// Get a value.
	if value, found := p.Get("hello"); found {
		fmt.Println(value)
	}

	// Check if a key exists.
	if p.Contains("hello") {
		fmt.Println("hello exists")
	}

	// Get the number of items.
	fmt.Printf("pantry contains %d items\n", p.Count())

	// Iterate over keys.
	for key := range p.Keys() {
		fmt.Println(key)
	}

	// Iterate over values.
	for value := range p.Values() {
		fmt.Println(value)
	}

	// Iterate over all items.
	for key, value := range p.All() {
		fmt.Println(key, value)
	}

	// Remove a value.
	p.Remove("hello")

	// Clear all items.
	p.Clear()

	// Check if the pantry is empty.
	if p.IsEmpty() {
		fmt.Println("pantry is empty")
	}
}
```

## API

### `New[T any](ctx context.Context, expiration time.Duration) *Pantry[T]`

Creates a new pantry. The `expiration` duration is the time-to-live for items in
the pantry. The context can be used to gracefully shutdown the pantry and free
up resources.

### `(p *Pantry[T]) Get(key string) (T, bool)`

Gets a value from the pantry. If the item has expired, it will be removed from
the pantry and `false` will be returned.

### `(p *Pantry[T]) Set(key string, value T)`

Sets a value in the pantry. The item will expire after the expiration time.

### `(p *Pantry[T]) Remove(key string)`

Removes a value from the pantry.

### `(p *Pantry[T]) IsEmpty() bool`

Returns `true` if the pantry is empty.

### `(p *Pantry[T]) Clear()`

Removes all items from the pantry.

### `(p *Pantry[T]) Contains(key string) bool`

Returns `true` if the key exists in the pantry.

### `(p *Pantry[T]) Count() int`

Returns the number of items in the pantry.

### `(p *Pantry[T]) Keys() iter.Seq[string]`

Returns an iterator over the keys in the pantry.

### `(p *Pantry[T]) Values() iter.Seq[T]`

Returns an iterator over the values in the pantry.

### `(p *Pantry[T]) All() iter.Seq2[string, T]`

Returns an iterator over the keys and values in the pantry.

## Expiration

Items in the pantry will be automatically removed after they expire. The
expiration time is set when the pantry is created.

## Context Cancellation

The pantry can be gracefully shut down by canceling the context that was passed
to the `New` function. This will stop the background cleanup goroutine and
remove all items from the pantry.

## Benchmarks

```
goos: darwin
goarch: arm64
pkg: github.com/webermarci/pantry
cpu: Apple M1
BenchmarkGet/0-8    	1000000000	         0.0000002 ns/op	       0 B/op	       0 allocs/op
BenchmarkSet/0-8    	1000000000	         0.0000010 ns/op	       0 B/op	       0 allocs/op
BenchmarkRemove/0-8 	1000000000	         0.0000002 ns/op	       0 B/op	       0 allocs/op
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an
issue.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.
