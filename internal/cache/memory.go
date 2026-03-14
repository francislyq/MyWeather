package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type item struct {
	value     interface{}
	fetchedAt time.Time
	expiresAt time.Time
}

type MemoryCache struct {
	mu     sync.RWMutex
	items  map[string]*item
	hits   atomic.Int64
	misses atomic.Int64
}

func NewMemoryCache() Cache {
	return &MemoryCache{
		items: make(map[string]*item),
	}
}

func (c *MemoryCache) Get(key string) (*Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, found := c.items[key]

	if !found {
		c.misses.Add(1)
		return nil, false
	}

	if time.Now().After(it.expiresAt) {
		c.misses.Add(1)
		c.Delete(key)
		return nil, false
	}

	c.hits.Add(1)
	return &Entry{
		Value:     it.value,
		FetchedAt: it.fetchedAt.Unix(),
		ExpiresAt: it.expiresAt.Unix(),
	}, true
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &item{
		value:     value,
		fetchedAt: time.Now(),
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*item)
	c.hits.Store(0)
	c.misses.Store(0)
}

func (c *MemoryCache) Stats() Stats {
	return Stats{
		Hits:   int(c.hits.Load()),
		Misses: int(c.misses.Load()),
		Size:   len(c.items),
	}
}

func (c *MemoryCache) StartCleanup(interval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-stopChan:
			return
		}
	}
}

func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, it := range c.items {
		if now.After(it.expiresAt) {
			delete(c.items, key)
		}
	}
}
