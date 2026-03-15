package cache

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheSetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("test_key", "test_value", 5*time.Minute)

	entry, found := cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "test_value", entry.Value)
	assert.NotZero(t, entry.FetchedAt)
	assert.NotZero(t, entry.ExpiresAt)
	assert.True(t, entry.ExpiresAt.After(entry.FetchedAt))
}

func TestMemoryCacheGetMissingKey(t *testing.T) {
	cache := NewMemoryCache()

	entry, found := cache.Get("missing_key")
	assert.False(t, found)
	assert.Nil(t, entry)
}

func TestMemoryCacheTTLExpiry(t *testing.T) {
	cache := NewMemoryCache().(*MemoryCache)

	stopChan := make(chan struct{})
	defer close(stopChan)
	go cache.StartCleanup(200*time.Millisecond, stopChan)

	cache.Set("temp_key", "temp_value", 1*time.Second)
	entry, found := cache.Get("temp_key")
	assert.True(t, found)
	assert.Equal(t, "temp_value", entry.Value)

	// Wait for TTL to expire and cleanup to run
	time.Sleep(1500 * time.Millisecond)

	entry, found = cache.Get("temp_key")
	assert.False(t, found)
	assert.Nil(t, entry)
}

func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key_to_delete", "value", 5*time.Minute)
	entry, found := cache.Get("key_to_delete")
	assert.True(t, found)
	assert.Equal(t, "value", entry.Value)

	cache.Delete("key_to_delete")
	entry, found = cache.Get("key_to_delete")
	assert.False(t, found)
	assert.Nil(t, entry)
}

func TestMemoryCacheStats(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", "value1", 5*time.Minute)
	cache.Set("key2", "value2", 5*time.Minute)

	// Cache hit
	cache.Get("key1")
	// Cache miss
	cache.Get("missing_key")

	stats := cache.Stats()
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
	assert.Equal(t, 2, stats.Size)
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", "value1", 5*time.Minute)
	cache.Set("key2", "value2", 5*time.Minute)

	// Cache hit
	cache.Get("key1")
	// Cache miss
	cache.Get("missing_key")

	stats := cache.Stats()
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
	assert.Equal(t, 2, stats.Size)

	cache.Clear()

	stats = cache.Stats()
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 0, stats.Misses)
	assert.Equal(t, 0, stats.Size)

	entry, found := cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, entry)

	entry, found = cache.Get("key2")
	assert.False(t, found)
	assert.Nil(t, entry)

	stats = cache.Stats()
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 2, stats.Misses)
	assert.Equal(t, 0, stats.Size)
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	numGoroutines := 100
	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()

			key := "key" + strconv.Itoa(i)
			cache.Set(key, "value"+strconv.Itoa(i), 5*time.Minute)

			entry, found := cache.Get(key)
			assert.True(t, found)
			assert.Equal(t, "value"+strconv.Itoa(i), entry.Value)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	stats := cache.Stats()
	assert.Equal(t, numGoroutines, stats.Hits)
	assert.Equal(t, 0, stats.Misses)
	assert.Equal(t, numGoroutines, stats.Size)
}

func TestMemoryCacheOverrideEntry(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("test_key", "initial_value", 5*time.Minute)
	entry, found := cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "initial_value", entry.Value)

	cache.Set("test_key", "new_value", 5*time.Minute)
	entry, found = cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "new_value", entry.Value)
}

func TestMemoryCacheEvictExpiredEntry(t *testing.T) {
	cache := NewMemoryCache().(*MemoryCache)

	stopChan := make(chan struct{})
	defer close(stopChan)
	go cache.StartCleanup(2000*time.Millisecond, stopChan)

	cache.Set("temp_key", "temp_value", 2000*time.Millisecond)
	entry, found := cache.Get("temp_key")
	assert.True(t, found)
	assert.Equal(t, "temp_value", entry.Value)

	// Wait for TTL to expire
	time.Sleep(2500 * time.Millisecond)

	entry, found = cache.Get("temp_key")
	assert.False(t, found)
	assert.Nil(t, entry)
}
