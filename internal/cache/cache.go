package cache

import (
	"time"
)

type Entry struct {
	Value      interface{}
	FetchedAt  time.Time
	ExpiresAt  time.Time
	IsStale    bool
}

type Stats struct {
	Hits   int `json:"hits"`
	Misses int `json:"misses"`
	Size   int `json:"size"`
}

type Cache interface {
	Get(key string) (*Entry, bool)
	Set(key string, value interface{}, ttl time.Duration, staleWindow time.Duration)
	Delete(key string)
	Clear()
	Stats() Stats
	StartCleanup(interval time.Duration, stopChan <-chan struct{})
}
