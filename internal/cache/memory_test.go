package cache

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheSetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("test_key", "test_value", 5*time.Minute, 0)

	entry, found := cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "test_value", entry.Value)
	assert.NotZero(t, entry.FetchedAt)
	assert.NotZero(t, entry.ExpiresAt)
	assert.True(t, entry.ExpiresAt.After(entry.FetchedAt))
	assert.False(t, entry.IsStale)
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

	cache.Set("temp_key", "temp_value", 1*time.Second, 0)
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

	cache.Set("key_to_delete", "value", 5*time.Minute, 0)
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

	cache.Set("key1", "value1", 5*time.Minute, 0)
	cache.Set("key2", "value2", 5*time.Minute, 0)

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

	cache.Set("key1", "value1", 5*time.Minute, 0)
	cache.Set("key2", "value2", 5*time.Minute, 0)

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
			cache.Set(key, "value"+strconv.Itoa(i), 5*time.Minute, 0)

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

	cache.Set("test_key", "initial_value", 5*time.Minute, 0)
	entry, found := cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "initial_value", entry.Value)

	cache.Set("test_key", "new_value", 5*time.Minute, 0)
	entry, found = cache.Get("test_key")
	assert.True(t, found)
	assert.Equal(t, "new_value", entry.Value)
}

func TestMemoryCacheEvictExpiredEntry(t *testing.T) {
	cache := NewMemoryCache().(*MemoryCache)

	stopChan := make(chan struct{})
	defer close(stopChan)
	go cache.StartCleanup(2000*time.Millisecond, stopChan)

	cache.Set("temp_key", "temp_value", 2000*time.Millisecond, 0)
	entry, found := cache.Get("temp_key")
	assert.True(t, found)
	assert.Equal(t, "temp_value", entry.Value)

	// Wait for TTL to expire
	time.Sleep(2500 * time.Millisecond)

	entry, found = cache.Get("temp_key")
	assert.False(t, found)
	assert.Nil(t, entry)
}

// --- Stale-while-revalidate tests ---

// TestSWRFreshEntryIsNotStale verifies that an entry within its TTL is returned
// as fresh (IsStale=false).
func TestSWRFreshEntryIsNotStale(t *testing.T) {
	c := NewMemoryCache()

	c.Set("city", "toronto", 5*time.Minute, 1*time.Minute)

	entry, found := c.Get("city")
	assert.True(t, found)
	assert.Equal(t, "toronto", entry.Value)
	assert.False(t, entry.IsStale, "entry should be fresh when fetched within TTL")
}

// TestSWRStaleEntryServedWithinWindow verifies that once the TTL passes but the
// stale window has not, Get returns the entry with IsStale=true (stale hit).
func TestSWRStaleEntryServedWithinWindow(t *testing.T) {
	c := NewMemoryCache()

	// TTL=50ms, stale window=500ms → staleUntil=550ms
	c.Set("city", "toronto", 50*time.Millisecond, 500*time.Millisecond)

	// Let the TTL expire, but stay inside the stale window
	time.Sleep(100 * time.Millisecond)

	entry, found := c.Get("city")
	assert.True(t, found, "stale entry should still be found within the stale window")
	assert.Equal(t, "toronto", entry.Value)
	assert.True(t, entry.IsStale, "entry past its TTL but within stale window should be marked stale")
}

// TestSWRExpiredBeyondStaleWindowIsMiss verifies that once both the TTL and the
// stale window have elapsed, Get returns a cache miss.
func TestSWRExpiredBeyondStaleWindowIsMiss(t *testing.T) {
	c := NewMemoryCache()

	// TTL=50ms, stale window=50ms → staleUntil=100ms
	c.Set("city", "toronto", 50*time.Millisecond, 50*time.Millisecond)

	// Wait past both the TTL and the stale window
	time.Sleep(200 * time.Millisecond)

	entry, found := c.Get("city")
	assert.False(t, found, "entry past the stale window should be a cache miss")
	assert.Nil(t, entry)
}

// TestSWRZeroStaleWindowEnforcesTTL verifies that setting staleWindow=0 reverts
// to strict TTL behaviour: an expired entry is immediately a miss on Get().
func TestSWRZeroStaleWindowEnforcesTTL(t *testing.T) {
	c := NewMemoryCache()

	// staleWindow=0 → staleUntil=expiresAt (strict TTL)
	c.Set("city", "toronto", 50*time.Millisecond, 0)

	// Wait past the TTL
	time.Sleep(100 * time.Millisecond)

	entry, found := c.Get("city")
	assert.False(t, found, "with staleWindow=0, an expired entry should be an immediate miss")
	assert.Nil(t, entry)
}

// TestSWRStaleHitCountedAsHit verifies that serving a stale entry increments
// the hit counter (not miss), consistent with data having been returned.
func TestSWRStaleHitCountedAsHit(t *testing.T) {
	c := NewMemoryCache()

	c.Set("city", "toronto", 50*time.Millisecond, 500*time.Millisecond)

	// Let the TTL expire so next Get is a stale hit
	time.Sleep(100 * time.Millisecond)

	entry, found := c.Get("city")
	assert.True(t, found)
	assert.True(t, entry.IsStale)

	stats := c.Stats()
	assert.Equal(t, 1, stats.Hits, "stale hit should increment the hit counter")
	assert.Equal(t, 0, stats.Misses)
}

// TestSWRExpiredBeyondStaleWindowCountedAsMiss verifies that a Get() past the
// stale window increments the miss counter.
func TestSWRExpiredBeyondStaleWindowCountedAsMiss(t *testing.T) {
	c := NewMemoryCache()

	c.Set("city", "toronto", 50*time.Millisecond, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	_, found := c.Get("city")
	assert.False(t, found)

	stats := c.Stats()
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 1, stats.Misses, "get past the stale window should increment the miss counter")
}

// TestSWRCleanupRespectsStaleWindow verifies that the background cleanup only
// removes entries whose stale window has fully elapsed, leaving entries that are
// stale-but-still-serveable intact.
func TestSWRCleanupRespectsStaleWindow(t *testing.T) {
	c := NewMemoryCache().(*MemoryCache)

	// key1: TTL=50ms, stale window=50ms → staleUntil≈100ms (fully expired after 200ms)
	c.Set("key1", "val1", 50*time.Millisecond, 50*time.Millisecond)
	// key2: TTL=50ms, stale window=5s  → staleUntil≈5050ms (still serveable after 200ms)
	c.Set("key2", "val2", 50*time.Millisecond, 5*time.Second)

	// Wait past both TTLs and key1's stale window, but inside key2's stale window
	time.Sleep(200 * time.Millisecond)

	c.cleanup()

	// key1 should have been swept — it is past its staleUntil
	entry, found := c.Get("key1")
	assert.False(t, found, "entry past its stale window should be removed by cleanup")
	assert.Nil(t, entry)

	// key2 should still be present and marked stale
	entry, found = c.Get("key2")
	assert.True(t, found, "entry inside its stale window should survive cleanup")
	assert.True(t, entry.IsStale, "surviving entry should be marked stale")
	assert.Equal(t, "val2", entry.Value)
}

// TestSWRRevalidationRefreshesEntry verifies that after a stale hit, storing a
// fresh value (simulating the background revalidation result) produces a
// non-stale entry on the subsequent Get.
func TestSWRRevalidationRefreshesEntry(t *testing.T) {
	c := NewMemoryCache()

	c.Set("city", "toronto-v1", 50*time.Millisecond, 500*time.Millisecond)

	// Let TTL expire so the entry becomes stale
	time.Sleep(100 * time.Millisecond)

	entry, found := c.Get("city")
	assert.True(t, found)
	assert.True(t, entry.IsStale)
	assert.Equal(t, "toronto-v1", entry.Value)

	// Simulate background revalidation writing fresh data
	c.Set("city", "toronto-v2", 5*time.Minute, 1*time.Minute)

	entry, found = c.Get("city")
	assert.True(t, found)
	assert.False(t, entry.IsStale, "entry should be fresh after revalidation")
	assert.Equal(t, "toronto-v2", entry.Value)
}
