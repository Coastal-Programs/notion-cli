package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	if c.Size() != 0 {
		t.Errorf("expected empty cache, got size %d", c.Size())
	}
	if c.maxSize != 100 {
		t.Errorf("expected maxSize 100, got %d", c.maxSize)
	}
}

func TestNewCacheInvalidSize(t *testing.T) {
	c := NewCache(0)
	defer c.Stop()

	if c.maxSize != 1000 {
		t.Errorf("expected default maxSize 1000 for invalid input, got %d", c.maxSize)
	}

	c2 := NewCache(-5)
	defer c2.Stop()

	if c2.maxSize != 1000 {
		t.Errorf("expected default maxSize 1000 for negative input, got %d", c2.maxSize)
	}
}

func TestSetAndGet(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "value1", 1*time.Minute)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestGetMiss(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	val, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected miss for nonexistent key")
	}
	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}

func TestGetExpired(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "value1", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	val, ok := c.Get("key1")
	if ok {
		t.Error("expected expired key to return false")
	}
	if val != nil {
		t.Errorf("expected nil for expired key, got %v", val)
	}
}

func TestSetOverwrite(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Minute)
	c.Set("key1", "v2", 1*time.Minute)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "v2" {
		t.Errorf("expected v2, got %v", val)
	}
	if c.Size() != 1 {
		t.Errorf("expected size 1 after overwrite, got %d", c.Size())
	}
}

func TestDelete(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "value1", 1*time.Minute)
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	// Should not panic
	c.Delete("nonexistent")
}

func TestClear(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Minute)
	c.Set("key2", "v2", 1*time.Minute)
	c.Set("key3", "v3", 1*time.Minute)

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected empty cache after clear, got size %d", c.Size())
	}

	stats := c.Stats()
	if stats.Hits != 0 || stats.Misses != 0 || stats.Evictions != 0 {
		t.Error("expected stats reset after clear")
	}
}

func TestSize(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Minute)
	c.Set("key2", "v2", 1*time.Minute)

	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	c.Delete("key1")
	if c.Size() != 1 {
		t.Errorf("expected size 1 after delete, got %d", c.Size())
	}
}

func TestStats(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Minute)

	// 1 hit
	c.Get("key1")
	// 2 misses
	c.Get("nonexistent1")
	c.Get("nonexistent2")

	stats := c.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("expected 2 misses, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}
	// HitRate = 1/3 = 0.3333
	if stats.HitRate < 0.33 || stats.HitRate > 0.34 {
		t.Errorf("expected hit rate ~0.3333, got %f", stats.HitRate)
	}
}

func TestStatsZeroDivision(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	stats := c.Stats()
	if stats.HitRate != 0 {
		t.Errorf("expected 0 hit rate with no operations, got %f", stats.HitRate)
	}
}

func TestEviction(t *testing.T) {
	c := NewCache(3)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Minute)
	time.Sleep(1 * time.Millisecond)
	c.Set("key2", "v2", 1*time.Minute)
	time.Sleep(1 * time.Millisecond)
	c.Set("key3", "v3", 1*time.Minute)

	// Cache is full, adding key4 should evict key1 (oldest)
	c.Set("key4", "v4", 1*time.Minute)

	if c.Size() != 3 {
		t.Errorf("expected size 3, got %d", c.Size())
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be evicted")
	}

	val, ok := c.Get("key4")
	if !ok {
		t.Error("expected key4 to exist")
	}
	if val != "v4" {
		t.Errorf("expected v4, got %v", val)
	}

	stats := c.Stats()
	if stats.Evictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.Evictions)
	}
}

func TestDifferentValueTypes(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("string", "hello", 1*time.Minute)
	c.Set("int", 42, 1*time.Minute)
	c.Set("float", 3.14, 1*time.Minute)
	c.Set("bool", true, 1*time.Minute)
	c.Set("struct", struct{ Name string }{"test"}, 1*time.Minute)
	c.Set("nil", nil, 1*time.Minute)

	if v, ok := c.Get("string"); !ok || v != "hello" {
		t.Error("string value mismatch")
	}
	if v, ok := c.Get("int"); !ok || v != 42 {
		t.Error("int value mismatch")
	}
	if v, ok := c.Get("float"); !ok || v != 3.14 {
		t.Error("float value mismatch")
	}
	if v, ok := c.Get("bool"); !ok || v != true {
		t.Error("bool value mismatch")
	}
	if v, ok := c.Get("nil"); !ok || v != nil {
		t.Errorf("nil value mismatch, ok=%v, v=%v", ok, v)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := NewCache(1000)
	defer c.Stop()

	var wg sync.WaitGroup
	numGoroutines := 50
	numOps := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				c.Set(key, j, 1*time.Minute)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				c.Get(key)
			}
		}(i)
	}

	wg.Wait()
	// No data races or panics means pass
}

func TestCacheKeyForResource(t *testing.T) {
	tests := []struct {
		resourceType string
		id           string
		expected     string
	}{
		{"database", "abc-123", "database:abc-123"},
		{"page", "xyz", "page:xyz"},
		{"user", "user-1", "user:user-1"},
		{"block", "blk-42", "block:blk-42"},
	}

	for _, tt := range tests {
		result := CacheKeyForResource(tt.resourceType, tt.id)
		if result != tt.expected {
			t.Errorf("CacheKeyForResource(%q, %q) = %q, want %q",
				tt.resourceType, tt.id, result, tt.expected)
		}
	}
}

func TestTTLConstants(t *testing.T) {
	if DatabaseTTL != 10*time.Minute {
		t.Errorf("DatabaseTTL = %v, want 10m", DatabaseTTL)
	}
	if UserTTL != 1*time.Hour {
		t.Errorf("UserTTL = %v, want 1h", UserTTL)
	}
	if PageTTL != 1*time.Minute {
		t.Errorf("PageTTL = %v, want 1m", PageTTL)
	}
	if BlockTTL != 30*time.Second {
		t.Errorf("BlockTTL = %v, want 30s", BlockTTL)
	}
}

func TestRemoveExpired(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("short", "v1", 1*time.Millisecond)
	c.Set("long", "v2", 1*time.Hour)

	time.Sleep(5 * time.Millisecond)
	c.removeExpired()

	if c.Size() != 1 {
		t.Errorf("expected 1 entry after cleanup, got %d", c.Size())
	}

	_, ok := c.Get("long")
	if !ok {
		t.Error("expected 'long' to survive cleanup")
	}
}

func TestExpiredEntryDoesNotCountAsHit(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "v1", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	c.Get("key1") // should be a miss since expired

	stats := c.Stats()
	if stats.Hits != 0 {
		t.Errorf("expected 0 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestEvictionOrder(t *testing.T) {
	c := NewCache(2)
	defer c.Stop()

	c.Set("first", "v1", 1*time.Minute)
	time.Sleep(1 * time.Millisecond)
	c.Set("second", "v2", 1*time.Minute)

	// This should evict "first" (oldest created)
	c.Set("third", "v3", 1*time.Minute)

	if _, ok := c.Get("first"); ok {
		t.Error("expected 'first' to be evicted")
	}
	if _, ok := c.Get("second"); !ok {
		t.Error("expected 'second' to still exist")
	}
	if _, ok := c.Get("third"); !ok {
		t.Error("expected 'third' to still exist")
	}
}
