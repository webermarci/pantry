# Pantry

[![Go Reference](https://pkg.go.dev/badge/github.com/webermarci/pantry.svg)](https://pkg.go.dev/github.com/webermarci/pantry)
[![Test](https://github.com/webermarci/pantry/actions/workflows/test.yml/badge.svg)](https://github.com/webermarci/pantry/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Pantry is a **high-performance, sharded, thread-safe**, in-memory key-value store for Go. It is designed for low-latency environments where minimizing lock contention and garbage collection (GC) pressure is critical.

## Features

- **Sharded Architecture:** Uses multiple internal shards to minimize lock contention under heavy parallel load.
- **Persistence & Snapshotting:** Interfaces for granular key-saving or bulk state-saving.
- **Zero-Allocation Reads:** Optimized `Get` path with 0 B/op on hits.
- **Observability:** Built-in `Observer` interface for tracking hits, misses, evictions, and storage errors.
- **Atomic Updates:** Built-in `Update` method for thread-safe read-modify-write operations.
- **Go 1.23+ Iterators:** Native support for `for range` sequences.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"[github.com/webermarci/pantry](https://github.com/webermarci/pantry)"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new pantry with 64 shards and a 5-minute default TTL
	p := pantry.New[string, int](ctx, 
		pantry.WithShards[string, int](64),
		pantry.WithDefaultTTL[string, int](5 * time.Minute),
	)

	// Set a value
	p.Set("sensor_1", 22)

	// Atomic update: increment a value safely
	p.Update("sensor_1", func(val int, exists bool) int {
		return val + 1
	})

	// Get with zero allocations
	if val, found := p.Get("sensor_1"); found {
		fmt.Printf("Value: %d\n", val)
	}

	// Iterate using Go 1.23 iterators
	for k, v := range p.All() {
		fmt.Printf("%s: %d\n", k, v)
	}
}
```

## Benchmarks

```
goos: darwin
goarch: arm64
pkg: github.com/webermarci/pantry
cpu: Apple M5
BenchmarkPantry_Get_NoContention-10             146078527    7.79 ns/op   0 B/op   0 allocs/op
BenchmarkPantry_Set_Overwrite-10                 59000676   19.81 ns/op   8 B/op   1 allocs/op
BenchmarkPantry_Set_NewKeys-10                   30562476   35.67 ns/op   8 B/op   1 allocs/op
BenchmarkPantry_Sharding_Scaling/Shards-1-10     28397157   37.93 ns/op   0 B/op   0 allocs/op
BenchmarkPantry_Sharding_Scaling/Shards-16-10    37383273   32.06 ns/op   0 B/op   0 allocs/op
BenchmarkPantry_Sharding_Scaling/Shards-64-10    56208392   20.95 ns/op   0 B/op   0 allocs/op
BenchmarkPantry_Sharding_Scaling/Shards-256-10   78686805   16.03 ns/op   0 B/op   0 allocs/op
BenchmarkPantry_JanitorContention-10             33666648   35.28 ns/op   8 B/op   1 allocs/op
```
