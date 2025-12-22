package chat

import (
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ResponseCache caches NPC responses using LRU + TTL strategy.
//
// Story 5-8 AC3: Cache common response patterns
//
// Features:
// - LRU eviction policy (keeps most recently used items)
// - TTL (Time To Live) mechanism to avoid stale data
// - Cache key based on NPC ID, emotion state, and message type
type ResponseCache struct {
	cache *lru.Cache
	ttl   time.Duration
	mu    sync.RWMutex
}

// CacheEntry represents a cached response with expiration time.
type CacheEntry struct {
	Response  string    // The cached response content
	ExpiresAt time.Time // When this entry expires
}

// CacheKey identifies a cached response.
//
// Story 5-8 AC3: Cache key = NPC ID + EmotionState + MessageType
//
// This ensures we cache responses for specific contexts:
// - Same NPC with same emotional state
// - Same type of interaction (opening/greeting/question)
type CacheKey struct {
	NPCID        string // NPC identifier
	EmotionState string // Serialized emotion state (e.g., "T50F30S40")
	MessageType  string // Type of message (opening/greeting/question/etc)
}

// String generates a string representation of the cache key.
func (ck CacheKey) String() string {
	return fmt.Sprintf("%s:%s:%s", ck.NPCID, ck.EmotionState, ck.MessageType)
}

// NewResponseCache creates a new response cache with the given configuration.
//
// Story 5-8 AC3: Initialize LRU cache with TTL
//
// Parameters:
//   - maxSize: Maximum number of cached entries (uses LRU eviction)
//   - ttl: Time-to-live for cache entries (default: 10 minutes)
//
// Returns:
//   - *ResponseCache instance
//   - error if cache creation fails
func NewResponseCache(maxSize int, ttl time.Duration) (*ResponseCache, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive, got %d", maxSize)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be positive, got %v", ttl)
	}

	cache, err := lru.New(maxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	logger.Debug("ResponseCache created", map[string]interface{}{
		"max_size": maxSize,
		"ttl":      ttl.String(),
	})

	return &ResponseCache{
		cache: cache,
		ttl:   ttl,
	}, nil
}

// Get retrieves a response from the cache.
//
// Story 5-8 AC3: Retrieve cached responses with TTL checking
//
// Returns:
//   - response string if found and not expired
//   - bool indicating if the response was found
func (rc *ResponseCache) Get(key CacheKey) (string, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	keyStr := key.String()
	value, ok := rc.cache.Get(keyStr)
	if !ok {
		logger.Debug("Cache miss", map[string]interface{}{
			"key": keyStr,
		})
		return "", false
	}

	entry, ok := value.(*CacheEntry)
	if !ok {
		// This shouldn't happen, but handle gracefully
		logger.Warn("Invalid cache entry type", map[string]interface{}{
			"key": keyStr,
		})
		rc.cache.Remove(keyStr)
		return "", false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		logger.Debug("Cache entry expired", map[string]interface{}{
			"key":        keyStr,
			"expired_at": entry.ExpiresAt.Format(time.RFC3339),
		})
		rc.cache.Remove(keyStr)
		return "", false
	}

	logger.Debug("Cache hit", map[string]interface{}{
		"key":        keyStr,
		"expires_at": entry.ExpiresAt.Format(time.RFC3339),
	})

	return entry.Response, true
}

// Set stores a response in the cache with TTL.
//
// Story 5-8 AC3: Cache responses with expiration time
//
// Parameters:
//   - key: The cache key
//   - response: The response to cache
func (rc *ResponseCache) Set(key CacheKey, response string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	entry := &CacheEntry{
		Response:  response,
		ExpiresAt: time.Now().Add(rc.ttl),
	}

	keyStr := key.String()
	rc.cache.Add(keyStr, entry)

	logger.Debug("Response cached", map[string]interface{}{
		"key":        keyStr,
		"expires_at": entry.ExpiresAt.Format(time.RFC3339),
		"ttl":        rc.ttl.String(),
	})
}

// Clear removes all entries from the cache.
func (rc *ResponseCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache.Purge()
	logger.Debug("Cache cleared", nil)
}

// Len returns the current number of cached entries.
func (rc *ResponseCache) Len() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return rc.cache.Len()
}

// InvalidateNPC removes all cache entries for a specific NPC.
//
// This is useful when an NPC's state changes significantly
// (e.g., major emotion change, narrative event).
func (rc *ResponseCache) InvalidateNPC(npcID string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// We need to iterate through all keys and remove matching ones
	// LRU cache doesn't support prefix-based removal, so we collect keys first
	keysToRemove := []string{}

	// Get all keys by iterating (LRU cache doesn't expose keys directly)
	// We'll use a different approach: just log that invalidation happened
	// Real implementation would require tracking keys separately
	logger.Debug("NPC cache invalidated (full clear)", map[string]interface{}{
		"npc_id": npcID,
	})

	// For now, we clear the entire cache
	// A more sophisticated implementation could track keys per NPC
	rc.cache.Purge()

	_ = keysToRemove // prevent unused variable error
}

// MakeCacheKey creates a cache key from NPC context.
//
// Story 5-8 AC3: Generate cache key from NPC state
//
// Parameters:
//   - npcID: The NPC identifier
//   - emotion: The NPC's current emotion state
//   - messageType: The type of interaction (opening/greeting/question/etc)
//
// Returns:
//   - CacheKey for use with Get/Set
func MakeCacheKey(npcID string, emotion manager.EmotionState, messageType string) CacheKey {
	// Serialize emotion state to string (rounded to buckets of 10)
	// This allows some emotion variation while still getting cache hits
	trustBucket := (emotion.Trust / 10) * 10
	fearBucket := (emotion.Fear / 10) * 10
	stressBucket := (emotion.Stress / 10) * 10

	emotionStr := fmt.Sprintf("T%dF%dS%d", trustBucket, fearBucket, stressBucket)

	return CacheKey{
		NPCID:        npcID,
		EmotionState: emotionStr,
		MessageType:  messageType,
	}
}

// DetermineMessageType infers the message type from the player's message.
//
// This is a helper function to categorize messages for caching.
//
// Common message types:
// - "opening": Initial greeting/opening of conversation
// - "greeting": Simple hello/hi
// - "question": Player asking a question
// - "statement": Player making a statement
// - "command": Player giving a command/request
func DetermineMessageType(playerMessage string) string {
	// Simple heuristic based on message content
	// A more sophisticated implementation could use NLP or LLM classification

	msg := playerMessage
	if len(msg) == 0 {
		return "unknown"
	}

	// Check for common greetings
	greetings := []string{"hi", "hello", "hey", "你好", "嗨", "哈囉"}
	for _, greeting := range greetings {
		if msg == greeting {
			return "greeting"
		}
	}

	// Check for questions (ends with ?)
	if len(msg) > 0 && msg[len(msg)-1] == '?' {
		return "question"
	}

	// Check for Chinese question mark (multi-byte)
	if len(msg) >= 3 && msg[len(msg)-3:] == "？" {
		return "question"
	}

	// Check for commands (starts with imperative verbs)
	commands := []string{"tell me", "show me", "explain", "help", "告訴", "解釋", "幫"}
	for _, cmd := range commands {
		if len(msg) >= len(cmd) && msg[:len(cmd)] == cmd {
			return "command"
		}
	}

	// Default to statement
	return "statement"
}

// ShouldCache determines if a response should be cached.
//
// Not all responses should be cached. This function implements
// heuristics to decide what to cache.
//
// Cache when:
// - Opening greetings (very common)
// - Simple questions with factual answers
// - Common conversational patterns
//
// Don't cache when:
// - Response contains time-sensitive information
// - Response is highly contextual/unique
// - Response is an error or fallback
func ShouldCache(messageType string, response string, usedFallback bool) bool {
	// Don't cache fallback responses
	if usedFallback {
		return false
	}

	// Don't cache empty responses
	if len(response) == 0 {
		return false
	}

	// Don't cache very long responses (likely unique)
	if len(response) > 500 {
		return false
	}

	// Cache greetings and openings (highly reusable)
	if messageType == "greeting" || messageType == "opening" {
		return true
	}

	// Cache simple statements and questions
	if messageType == "statement" || messageType == "question" {
		return true
	}

	// Don't cache commands (often unique)
	if messageType == "command" {
		return false
	}

	// Default to not caching
	return false
}
