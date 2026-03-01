// Package cache provides an in-memory TTL cache with background eviction
// and workspace database caching for the Notion CLI.
package cache

import (
	"math"
	"sync"
	"time"
)

// Default TTLs by resource type.
const (
	DatabaseTTL = 10 * time.Minute
	UserTTL     = 1 * time.Hour
	PageTTL     = 1 * time.Minute
	BlockTTL    = 30 * time.Second
)

// CacheStats holds cache performance metrics.
type CacheStats struct {
	Hits      int     `json:"hits"`
	Misses    int     `json:"misses"`
	Size      int     `json:"size"`
	Evictions int     `json:"evictions"`
	HitRate   float64 `json:"hit_rate"`
}

// entry represents a cached value with an expiration time.
type entry struct {
	value     any
	expiresAt time.Time
	createdAt time.Time
}

// Cache is a thread-safe in-memory cache with TTL-based expiration.
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]*entry
	maxSize  int
	stats    CacheStats
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewCache creates a new Cache with the given maximum number of entries.
// It starts a background goroutine that cleans up expired entries every 30 seconds.
func NewCache(maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	c := &Cache{
		entries: make(map[string]*entry),
		maxSize: maxSize,
		stopCh:  make(chan struct{}),
	}
	go c.backgroundCleanup()
	return c
}

// Set adds or updates a key in the cache with the given TTL.
// If the cache is full, the oldest entry is evicted.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, just update it
	if _, exists := c.entries[key]; exists {
		c.entries[key] = &entry{
			value:     value,
			expiresAt: time.Now().Add(ttl),
			createdAt: time.Now(),
		}
		return
	}

	// Evict oldest if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		createdAt: time.Now(),
	}
}

// Get retrieves a value from the cache. Returns the value and true if found
// and not expired, or nil and false otherwise.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, exists := c.entries[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	if time.Now().After(e.expiresAt) {
		delete(c.entries, key)
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return e.value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache and resets stats.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*entry)
	c.stats = CacheStats{}
}

// Size returns the number of entries currently in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns a snapshot of cache performance metrics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.stats
	s.Size = len(c.entries)
	total := s.Hits + s.Misses
	if total > 0 {
		s.HitRate = math.Round(float64(s.Hits)/float64(total)*10000) / 10000
	}
	return s
}

// Stop terminates the background cleanup goroutine.
func (c *Cache) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// CacheKeyForResource generates a cache key for a given resource type and ID.
func CacheKeyForResource(resourceType, id string) string {
	return resourceType + ":" + id
}

// evictOldest removes the entry with the earliest createdAt time.
// Must be called with c.mu held.
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for k, e := range c.entries {
		if first || e.createdAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = e.createdAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.Evictions++
	}
}

// backgroundCleanup runs periodically to remove expired entries.
func (c *Cache) backgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.removeExpired()
		}
	}
}

// removeExpired deletes all expired entries from the cache.
func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
}
