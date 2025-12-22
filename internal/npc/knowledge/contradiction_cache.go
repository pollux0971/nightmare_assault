package knowledge

import (
	"container/list"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CacheEntry represents a cached contradiction analysis result.
// It includes TTL and hit tracking for monitoring.
//
// Story 8.2 AC4: 快取機制實作
type CacheEntry struct {
	Result    *ContradictionAnalysisResult
	Hash      string    // Hash of the two statements
	CreatedAt time.Time // When this entry was created
	ExpiresAt time.Time // When this entry expires
	Hits      int       // Number of cache hits for this entry
	element   *list.Element // For LRU tracking
}

// CacheStats provides cache performance metrics.
// Story 8.2 AC4: 快取命中率監控
type CacheStats struct {
	TotalRequests int     // Total number of Get() calls
	CacheHits     int     // Number of successful cache hits
	CacheMisses   int     // Number of cache misses
	HitRate       float64 // Hit rate as a percentage (0-100)
	CurrentSize   int     // Current number of entries in cache
	MaxSize       int     // Maximum cache size
	Evictions     int     // Number of LRU evictions
	Expirations   int     // Number of TTL expirations
}

// ContradictionCache implements an LRU cache with TTL for contradiction analysis results.
// It uses a hash of the two statements as the cache key and implements similarity matching.
//
// Story 8.2 AC4: 實作 ContradictionCache (LRU + TTL)
type ContradictionCache struct {
	maxSize    int                      // Maximum number of entries
	ttl        time.Duration            // Time-to-live for cache entries
	entries    map[string]*CacheEntry   // Hash -> Entry
	lruList    *list.List               // LRU doubly-linked list
	mu         sync.RWMutex             // Protects cache access

	// Statistics
	totalRequests int
	cacheHits     int
	cacheMisses   int
	evictions     int
	expirations   int
}

// CacheConfig configures the contradiction cache.
type CacheConfig struct {
	MaxSize int           // Maximum number of cached entries (default: 1000)
	TTL     time.Duration // Time-to-live for entries (default: 30 minutes)
}

// DefaultCacheConfig returns a CacheConfig with sensible defaults.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize: 1000,
		TTL:     30 * time.Minute,
	}
}

// NewContradictionCache creates a new ContradictionCache with the given configuration.
// If config is nil, uses DefaultCacheConfig().
//
// Story 8.2 AC4: 實作 ContradictionCache (LRU + TTL)
func NewContradictionCache(config *CacheConfig) *ContradictionCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &ContradictionCache{
		maxSize: config.MaxSize,
		ttl:     config.TTL,
		entries: make(map[string]*CacheEntry),
		lruList: list.New(),
	}
}

// Get retrieves a cached contradiction analysis result if available and not expired.
// It updates LRU ordering on cache hit.
//
// Story 8.2 AC4: 相似度計算與比對邏輯
// Story 8.2 AC4: 快取命中率監控
func (c *ContradictionCache) Get(statementA, statementB string) *ContradictionAnalysisResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests++

	// Compute cache key (order-independent hash)
	hash := c.computeHash(statementA, statementB)

	// Look up entry
	entry, exists := c.entries[hash]
	if !exists {
		c.cacheMisses++
		return nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Remove expired entry
		c.removeEntryLocked(entry)
		c.expirations++
		c.cacheMisses++
		return nil
	}

	// Cache hit - update LRU and stats
	c.lruList.MoveToFront(entry.element)
	entry.Hits++
	c.cacheHits++

	return entry.Result
}

// Put stores a contradiction analysis result in the cache.
// If the cache is full, it evicts the least recently used entry.
//
// Story 8.2 AC4: 實作 ContradictionCache (LRU + TTL)
func (c *ContradictionCache) Put(statementA, statementB string, result *ContradictionAnalysisResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := c.computeHash(statementA, statementB)

	// Check if entry already exists
	if existing, exists := c.entries[hash]; exists {
		// Update existing entry
		existing.Result = result
		existing.CreatedAt = time.Now()
		existing.ExpiresAt = time.Now().Add(c.ttl)
		c.lruList.MoveToFront(existing.element)
		return
	}

	// Create new entry
	entry := &CacheEntry{
		Result:    result,
		Hash:      hash,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
		Hits:      0,
	}

	// Add to front of LRU list
	entry.element = c.lruList.PushFront(entry)
	c.entries[hash] = entry

	// Evict LRU entry if cache is full
	if c.lruList.Len() > c.maxSize {
		c.evictLRULocked()
	}
}

// evictLRULocked removes the least recently used entry from the cache.
// Must be called with c.mu held.
func (c *ContradictionCache) evictLRULocked() {
	lruElement := c.lruList.Back()
	if lruElement == nil {
		return
	}

	entry := lruElement.Value.(*CacheEntry)
	c.removeEntryLocked(entry)
	c.evictions++
}

// removeEntryLocked removes an entry from the cache.
// Must be called with c.mu held.
func (c *ContradictionCache) removeEntryLocked(entry *CacheEntry) {
	c.lruList.Remove(entry.element)
	delete(c.entries, entry.Hash)
}

// computeHash computes an order-independent hash of two statements.
// This ensures "A vs B" and "B vs A" have the same cache key.
func (c *ContradictionCache) computeHash(statementA, statementB string) string {
	// Normalize statements
	a := strings.TrimSpace(strings.ToLower(statementA))
	b := strings.TrimSpace(strings.ToLower(statementB))

	// Ensure consistent ordering (alphabetical)
	if a > b {
		a, b = b, a
	}

	// Compute SHA256 hash
	combined := a + "|" + b
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// GetStats returns cache performance statistics.
// Story 8.2 AC4: 快取命中率監控
func (c *ContradictionCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hitRate := 0.0
	if c.totalRequests > 0 {
		hitRate = float64(c.cacheHits) / float64(c.totalRequests) * 100.0
	}

	return CacheStats{
		TotalRequests: c.totalRequests,
		CacheHits:     c.cacheHits,
		CacheMisses:   c.cacheMisses,
		HitRate:       hitRate,
		CurrentSize:   c.lruList.Len(),
		MaxSize:       c.maxSize,
		Evictions:     c.evictions,
		Expirations:   c.expirations,
	}
}

// Clear removes all entries from the cache and resets statistics.
// Useful for testing or when memory needs to be reclaimed.
func (c *ContradictionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.lruList = list.New()
	c.totalRequests = 0
	c.cacheHits = 0
	c.cacheMisses = 0
	c.evictions = 0
	c.expirations = 0
}

// CleanExpired removes all expired entries from the cache.
// This can be called periodically to prevent expired entries from consuming memory.
//
// Story 8.2 AC4: 過期項目清理策略
func (c *ContradictionCache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Iterate through all entries and remove expired ones
	for hash, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			c.lruList.Remove(entry.element)
			delete(c.entries, hash)
			c.expirations++
			removed++
		}
	}

	return removed
}

// Size returns the current number of entries in the cache.
func (c *ContradictionCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}
