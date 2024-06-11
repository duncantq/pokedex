package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entries map[string]cacheEntry
	mu      sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	value     []byte
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		entries: make(map[string]cacheEntry),
	}

	go cache.reapLoop(interval)

	return &cache
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return entry.value, true
}

func (c *Cache) Add(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		value:     value,
	}
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.createdAt).Seconds() > interval.Seconds() {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}
