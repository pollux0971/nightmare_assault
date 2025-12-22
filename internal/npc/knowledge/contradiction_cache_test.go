package knowledge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewContradictionCache tests cache creation
func TestNewContradictionCache(t *testing.T) {
	t.Run("creates cache with default config", func(t *testing.T) {
		cache := NewContradictionCache(nil)
		require.NotNil(t, cache)
		assert.Equal(t, 1000, cache.maxSize)
		assert.Equal(t, 30*time.Minute, cache.ttl)
		assert.Equal(t, 0, cache.Size())
	})

	t.Run("creates cache with custom config", func(t *testing.T) {
		config := &CacheConfig{
			MaxSize: 100,
			TTL:     5 * time.Minute,
		}
		cache := NewContradictionCache(config)
		require.NotNil(t, cache)
		assert.Equal(t, 100, cache.maxSize)
		assert.Equal(t, 5*time.Minute, cache.ttl)
	})
}

// TestCachePutAndGet tests basic cache operations
func TestCachePutAndGet(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        8,
		Type:            "direct",
		Explanation:     "Test contradiction",
	}

	t.Run("put and get returns same result", func(t *testing.T) {
		cache.Put("statement A", "statement B", result)
		retrieved := cache.Get("statement A", "statement B")

		require.NotNil(t, retrieved)
		assert.Equal(t, result.IsContradictory, retrieved.IsContradictory)
		assert.Equal(t, result.Severity, retrieved.Severity)
		assert.Equal(t, result.Type, retrieved.Type)
		assert.Equal(t, result.Explanation, retrieved.Explanation)
	})

	t.Run("get with reversed order returns same result", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		retrieved := cache.Get("B", "A")

		require.NotNil(t, retrieved)
		assert.Equal(t, result.IsContradictory, retrieved.IsContradictory)
	})

	t.Run("get non-existent returns nil", func(t *testing.T) {
		retrieved := cache.Get("nonexistent A", "nonexistent B")
		assert.Nil(t, retrieved)
	})

	t.Run("cache size increases with put", func(t *testing.T) {
		cache.Clear()
		assert.Equal(t, 0, cache.Size())

		cache.Put("A", "B", result)
		assert.Equal(t, 1, cache.Size())

		cache.Put("C", "D", result)
		assert.Equal(t, 2, cache.Size())
	})

	t.Run("updating existing entry doesn't increase size", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		assert.Equal(t, 1, cache.Size())

		newResult := &ContradictionAnalysisResult{
			IsContradictory: false,
			Severity:        2,
			Type:            "none",
			Explanation:     "Updated",
		}
		cache.Put("A", "B", newResult)
		assert.Equal(t, 1, cache.Size())

		retrieved := cache.Get("A", "B")
		assert.Equal(t, newResult.Explanation, retrieved.Explanation)
	})
}

// TestCacheLRUEviction tests LRU eviction when cache is full
func TestCacheLRUEviction(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 3,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        5,
		Type:            "direct",
		Explanation:     "Test",
	}

	t.Run("evicts LRU entry when cache is full", func(t *testing.T) {
		// Add 3 entries (fills cache)
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)
		cache.Put("E", "F", result)
		assert.Equal(t, 3, cache.Size())

		// Add 4th entry - should evict "A", "B" (oldest)
		cache.Put("G", "H", result)
		assert.Equal(t, 3, cache.Size())

		// First entry should be evicted
		assert.Nil(t, cache.Get("A", "B"))

		// Others should still be present
		assert.NotNil(t, cache.Get("C", "D"))
		assert.NotNil(t, cache.Get("E", "F"))
		assert.NotNil(t, cache.Get("G", "H"))
	})

	t.Run("accessing entry makes it most recent", func(t *testing.T) {
		cache.Clear()

		cache.Put("A", "B", result)
		cache.Put("C", "D", result)
		cache.Put("E", "F", result)

		// Access "A", "B" to make it most recent
		cache.Get("A", "B")

		// Add new entry - should evict "C", "D" (now oldest)
		cache.Put("G", "H", result)

		assert.NotNil(t, cache.Get("A", "B")) // Should still be present
		assert.Nil(t, cache.Get("C", "D"))     // Should be evicted
		assert.NotNil(t, cache.Get("E", "F"))
		assert.NotNil(t, cache.Get("G", "H"))
	})

	t.Run("tracks eviction count", func(t *testing.T) {
		cache.Clear()

		// Fill cache
		for i := 0; i < 3; i++ {
			cache.Put(string(rune('A'+i)), string(rune('B'+i)), result)
		}

		stats := cache.GetStats()
		assert.Equal(t, 0, stats.Evictions)

		// Trigger evictions
		cache.Put("D", "E", result)
		cache.Put("F", "G", result)

		stats = cache.GetStats()
		assert.Equal(t, 2, stats.Evictions)
	})
}

// TestCacheTTLExpiration tests TTL-based expiration
func TestCacheTTLExpiration(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     50 * time.Millisecond, // Very short TTL for testing
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        7,
		Type:            "temporal",
		Explanation:     "Test",
	}

	t.Run("expired entry returns nil", func(t *testing.T) {
		cache.Put("A", "B", result)

		// Immediately should be available
		assert.NotNil(t, cache.Get("A", "B"))

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		// Should be expired now
		assert.Nil(t, cache.Get("A", "B"))
	})

	t.Run("expired entry is removed from cache", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		assert.Equal(t, 1, cache.Size())

		time.Sleep(60 * time.Millisecond)
		cache.Get("A", "B") // Trigger expiration check

		assert.Equal(t, 0, cache.Size())
	})

	t.Run("tracks expiration count", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)

		stats := cache.GetStats()
		assert.Equal(t, 0, stats.Expirations)

		time.Sleep(60 * time.Millisecond)
		cache.Get("A", "B")
		cache.Get("C", "D")

		stats = cache.GetStats()
		assert.Equal(t, 2, stats.Expirations)
	})
}

// TestCacheStats tests cache statistics tracking
func TestCacheStats(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        6,
		Type:            "indirect",
		Explanation:     "Test",
	}

	t.Run("tracks cache hits and misses", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)

		// Cache hit
		cache.Get("A", "B")

		// Cache miss
		cache.Get("X", "Y")

		stats := cache.GetStats()
		assert.Equal(t, 2, stats.TotalRequests)
		assert.Equal(t, 1, stats.CacheHits)
		assert.Equal(t, 1, stats.CacheMisses)
		assert.Equal(t, 50.0, stats.HitRate)
	})

	t.Run("calculates hit rate correctly", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)

		// 3 hits, 1 miss = 75% hit rate
		cache.Get("A", "B")
		cache.Get("A", "B")
		cache.Get("A", "B")
		cache.Get("X", "Y")

		stats := cache.GetStats()
		assert.Equal(t, 4, stats.TotalRequests)
		assert.Equal(t, 3, stats.CacheHits)
		assert.Equal(t, 1, stats.CacheMisses)
		assert.Equal(t, 75.0, stats.HitRate)
	})

	t.Run("reports current size and max size", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)

		stats := cache.GetStats()
		assert.Equal(t, 2, stats.CurrentSize)
		assert.Equal(t, 10, stats.MaxSize)
	})

	t.Run("handles zero requests gracefully", func(t *testing.T) {
		cache.Clear()
		stats := cache.GetStats()

		assert.Equal(t, 0, stats.TotalRequests)
		assert.Equal(t, 0.0, stats.HitRate)
	})
}

// TestCacheCleanExpired tests manual expired entry cleanup
func TestCacheCleanExpired(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     50 * time.Millisecond,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        5,
		Type:            "direct",
		Explanation:     "Test",
	}

	t.Run("removes all expired entries", func(t *testing.T) {
		cache.Clear()

		// Add entries
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)
		cache.Put("E", "F", result)
		assert.Equal(t, 3, cache.Size())

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		// Clean expired
		removed := cache.CleanExpired()
		assert.Equal(t, 3, removed)
		assert.Equal(t, 0, cache.Size())
	})

	t.Run("doesn't remove non-expired entries", func(t *testing.T) {
		cache.Clear()

		// Add entry
		cache.Put("A", "B", result)

		// Clean immediately (nothing should be expired)
		removed := cache.CleanExpired()
		assert.Equal(t, 0, removed)
		assert.Equal(t, 1, cache.Size())
	})

	t.Run("tracks expirations in stats", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)

		time.Sleep(60 * time.Millisecond)
		cache.CleanExpired()

		stats := cache.GetStats()
		assert.Equal(t, 2, stats.Expirations)
	})
}

// TestCacheComputeHash tests hash computation
func TestCacheComputeHash(t *testing.T) {
	cache := NewContradictionCache(nil)

	t.Run("same statements produce same hash", func(t *testing.T) {
		hash1 := cache.computeHash("A", "B")
		hash2 := cache.computeHash("A", "B")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("reversed statements produce same hash", func(t *testing.T) {
		hash1 := cache.computeHash("A", "B")
		hash2 := cache.computeHash("B", "A")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different statements produce different hash", func(t *testing.T) {
		hash1 := cache.computeHash("A", "B")
		hash2 := cache.computeHash("C", "D")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("normalizes whitespace and case", func(t *testing.T) {
		hash1 := cache.computeHash("  A  ", "  B  ")
		hash2 := cache.computeHash("a", "b")
		assert.Equal(t, hash1, hash2)
	})
}

// TestCacheClear tests cache clearing
func TestCacheClear(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        5,
		Type:            "direct",
		Explanation:     "Test",
	}

	t.Run("clears all entries", func(t *testing.T) {
		cache.Put("A", "B", result)
		cache.Put("C", "D", result)
		assert.Equal(t, 2, cache.Size())

		cache.Clear()
		assert.Equal(t, 0, cache.Size())
		assert.Nil(t, cache.Get("A", "B"))
		assert.Nil(t, cache.Get("C", "D"))
	})

	t.Run("resets all statistics", func(t *testing.T) {
		cache.Put("A", "B", result)
		cache.Get("A", "B")
		cache.Get("X", "Y")

		stats := cache.GetStats()
		assert.Greater(t, stats.TotalRequests, 0)

		cache.Clear()
		stats = cache.GetStats()

		assert.Equal(t, 0, stats.TotalRequests)
		assert.Equal(t, 0, stats.CacheHits)
		assert.Equal(t, 0, stats.CacheMisses)
		assert.Equal(t, 0, stats.Evictions)
		assert.Equal(t, 0, stats.Expirations)
	})
}

// TestCacheThreadSafety tests concurrent access
func TestCacheThreadSafety(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 100,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        5,
		Type:            "direct",
		Explanation:     "Test",
	}

	t.Run("concurrent put and get", func(t *testing.T) {
		cache.Clear()
		done := make(chan bool)

		// Multiple goroutines putting
		for i := 0; i < 10; i++ {
			go func(idx int) {
				cache.Put(string(rune('A'+idx)), string(rune('B'+idx)), result)
				done <- true
			}(i)
		}

		// Multiple goroutines getting
		for i := 0; i < 10; i++ {
			go func(idx int) {
				cache.Get(string(rune('A'+idx)), string(rune('B'+idx)))
				done <- true
			}(i)
		}

		// Wait for all
		for i := 0; i < 20; i++ {
			<-done
		}

		// Should not panic and cache should be in valid state
		stats := cache.GetStats()
		assert.GreaterOrEqual(t, stats.CurrentSize, 0)
		assert.LessOrEqual(t, stats.CurrentSize, 100)
	})
}

// TestCacheHitTracking tests individual entry hit counting
func TestCacheHitTracking(t *testing.T) {
	cache := NewContradictionCache(&CacheConfig{
		MaxSize: 10,
		TTL:     1 * time.Hour,
	})

	result := &ContradictionAnalysisResult{
		IsContradictory: true,
		Severity:        5,
		Type:            "direct",
		Explanation:     "Test",
	}

	t.Run("entry hits are tracked", func(t *testing.T) {
		cache.Clear()
		cache.Put("A", "B", result)

		// Multiple accesses
		for i := 0; i < 5; i++ {
			cache.Get("A", "B")
		}

		// We can't directly access entry.Hits, but we can verify through stats
		stats := cache.GetStats()
		assert.Equal(t, 5, stats.CacheHits)
	})
}
