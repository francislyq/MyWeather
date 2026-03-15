package cache

import (
	"strconv"
	"testing"
	"time"
)

// func BenchmarkMemoryCache(b *testing.B) {
// 	cache := NewMemoryCache().(*MemoryCache)
// 	cfg := &config.Config{
// 		Cities: []model.City{
// 			{ID: 1, Name: "City1"},
// 			{ID: 2, Name: "City2"},
// 			{ID: 3, Name: "City3"},
// 		},
// 	}
// 	log := logrus.New()

// 	var wg sync.WaitGroup
// 	for i := 0; i < b.N; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			key := strconv.Itoa((i % 3) + 1) // Cycle through city IDs
// 			cache.Set(key, "weather_data", 5*time.Second)
// 			entry, found := cache.Get(key)
// 			assert.True(b, found)
// 			assert.Equal(b, "weather_data", entry.Value)
// 		}(i)
// 	}
// 	wg.Wait()

// 	stats := cache.Stats()
// 	assert.Equal(b, b.N, stats.Hits+stats.Misses)
// }

func BenchmarkMemoryCacheSet(b *testing.B) {
	cache := NewMemoryCache().(*MemoryCache)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cache.Set(key, "weather_data", 5*time.Second)
	}
}

func BenchmarkMemoryCacheGet(b *testing.B) {
	cache := NewMemoryCache().(*MemoryCache)
	// Pre-populate cache with 1000 entries
	for i := 0; i < 1000; i++ {
		key := strconv.Itoa(i)
		cache.Set(key, "weather_data", 5*time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i % 1000) // Cycle through existing keys
		cache.Get(key)
	}
}

func BenchmarkMemoryCacheConcurrentAccess(b *testing.B) {
	cache := NewMemoryCache().(*MemoryCache)
	numGoroutines := 100
	b.ResetTimer()
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			key := strconv.Itoa(i)
			cache.Set(key, "weather_data", 5*time.Second)
			cache.Get(key)
		}(i)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Get("key")
			} else {
				cache.Set("key", "value", 5*time.Minute)
			}
			i++
		}
	})
}
