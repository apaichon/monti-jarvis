package rag

import (
	"sync"
	"time"
)

const voicePreloadTTL = 5 * time.Minute

type cacheEntry struct {
	result    Result
	expiresAt time.Time
}

type preloadCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
}

func newPreloadCache() *preloadCache {
	return &preloadCache{items: make(map[string]cacheEntry)}
}

func (c *preloadCache) get(key string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return Result{}, false
	}
	return entry.result, true
}

func (c *preloadCache) set(key string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{result: result, expiresAt: time.Now().Add(voicePreloadTTL)}
}