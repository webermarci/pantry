package pantry

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// pregenKeys ensures we don't benchmark string formatting
func pregenKeys(n int) []string {
	k := make([]string, n)
	for i := range n {
		k[i] = fmt.Sprintf("key-%d", i)
	}
	return k
}

// --- Read Benchmarks ---

func BenchmarkPantry_Get_NoContention(b *testing.B) {
	ctx := context.Background()
	p := New[string, int](ctx)
	p.Set("bench", 42)

	for b.Loop() {
		p.Get("bench")
	}
}

// --- Allocation & Set Benchmarks ---

func BenchmarkPantry_Set_Overwrite(b *testing.B) {
	ctx := context.Background()
	p := New[string, int](ctx)
	p.Set("key", 1)

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		p.Set("key", i)
	}
}

func BenchmarkPantry_Set_NewKeys(b *testing.B) {
	ctx := context.Background()
	p := New[string, int](ctx)
	// Pregenerating more keys than b.N to be safe
	keys := pregenKeys(100000)

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		// Use modulo to stay within pregenerated slice
		p.Set(keys[i%100000], i)
	}
}

// --- Scaling Benchmarks ---

func BenchmarkPantry_Sharding_Scaling(b *testing.B) {
	ctx := context.Background()
	keys := pregenKeys(1000)

	shardCounts := []int{1, 16, 64, 256}

	for _, sc := range shardCounts {
		b.Run(fmt.Sprintf("Shards-%d", sc), func(b *testing.B) {
			p := New(ctx, WithShards[string, int](sc))

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := keys[i%1000]
					// 9:1 Read/Write ratio (standard cache simulation)
					if i%10 == 0 {
						p.Set(key, i)
					} else {
						p.Get(key)
					}
					i++
				}
			})
		})
	}
}

func BenchmarkPantry_JanitorContention(b *testing.B) {
	p := New(b.Context(),
		WithShards[int, int](256),
		WithJanitorInterval[int, int](1*time.Millisecond),
	)

	for i := range 10000 {
		p.Set(i, i, WithTTL(1*time.Millisecond))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Try to Set/Get while janitor is constantly locking shards to delete
			p.Set(i%100, i)
			p.Get(i % 100)
			i++
		}
	})
}
