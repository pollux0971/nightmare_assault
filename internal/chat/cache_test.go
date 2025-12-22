package chat

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResponseCache tests cache creation.
func TestNewResponseCache(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)

	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Len())
}

// TestNewResponseCache_InvalidParams tests cache creation with invalid parameters.
func TestNewResponseCache_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		maxSize int
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "negative maxSize",
			maxSize: -1,
			ttl:     10 * time.Minute,
			wantErr: true,
		},
		{
			name:    "zero maxSize",
			maxSize: 0,
			ttl:     10 * time.Minute,
			wantErr: true,
		},
		{
			name:    "negative ttl",
			maxSize: 100,
			ttl:     -1 * time.Minute,
			wantErr: true,
		},
		{
			name:    "zero ttl",
			maxSize: 100,
			ttl:     0,
			wantErr: true,
		},
		{
			name:    "valid params",
			maxSize: 100,
			ttl:     10 * time.Minute,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewResponseCache(tt.maxSize, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
			}
		})
	}
}

// TestResponseCache_GetSet tests basic cache get/set operations.
func TestResponseCache_GetSet(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)
	require.NoError(t, err)

	key := CacheKey{
		NPCID:        "npc_001",
		EmotionState: "T50F30S40",
		MessageType:  "greeting",
	}

	// Cache miss on first get
	_, found := cache.Get(key)
	assert.False(t, found)

	// Set value
	cache.Set(key, "Hello there!")

	// Cache hit on second get
	value, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, "Hello there!", value)
}

// TestResponseCache_TTLExpiration tests TTL expiration.
func TestResponseCache_TTLExpiration(t *testing.T) {
	cache, err := NewResponseCache(100, 100*time.Millisecond)
	require.NoError(t, err)

	key := CacheKey{
		NPCID:        "npc_001",
		EmotionState: "T50F30S40",
		MessageType:  "greeting",
	}

	// Set value
	cache.Set(key, "Hello!")

	// Should be available immediately
	value, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, "Hello!", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get(key)
	assert.False(t, found)
}

// TestResponseCache_LRUEviction tests LRU eviction policy.
func TestResponseCache_LRUEviction(t *testing.T) {
	cache, err := NewResponseCache(3, 10*time.Minute) // Only 3 items
	require.NoError(t, err)

	// Add 3 items
	cache.Set(CacheKey{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 1")
	cache.Set(CacheKey{NPCID: "npc_002", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 2")
	cache.Set(CacheKey{NPCID: "npc_003", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 3")

	assert.Equal(t, 3, cache.Len())

	// Add 4th item - should evict oldest (npc_001)
	cache.Set(CacheKey{NPCID: "npc_004", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 4")

	assert.Equal(t, 3, cache.Len())

	// npc_001 should be evicted
	_, found := cache.Get(CacheKey{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"})
	assert.False(t, found)

	// npc_004 should be present
	value, found := cache.Get(CacheKey{NPCID: "npc_004", EmotionState: "T50F30S40", MessageType: "greeting"})
	assert.True(t, found)
	assert.Equal(t, "Response 4", value)
}

// TestResponseCache_Clear tests cache clearing.
func TestResponseCache_Clear(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)
	require.NoError(t, err)

	// Add some entries
	cache.Set(CacheKey{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 1")
	cache.Set(CacheKey{NPCID: "npc_002", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 2")

	assert.Equal(t, 2, cache.Len())

	// Clear cache
	cache.Clear()

	assert.Equal(t, 0, cache.Len())

	// Entries should be gone
	_, found := cache.Get(CacheKey{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"})
	assert.False(t, found)
}

// TestResponseCache_InvalidateNPC tests NPC-specific invalidation.
func TestResponseCache_InvalidateNPC(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)
	require.NoError(t, err)

	// Add entries for multiple NPCs
	cache.Set(CacheKey{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 1")
	cache.Set(CacheKey{NPCID: "npc_001", EmotionState: "T60F20S30", MessageType: "question"}, "Response 2")
	cache.Set(CacheKey{NPCID: "npc_002", EmotionState: "T50F30S40", MessageType: "greeting"}, "Response 3")

	assert.Equal(t, 3, cache.Len())

	// Invalidate npc_001 (currently clears entire cache due to LRU limitation)
	cache.InvalidateNPC("npc_001")

	// After invalidation, cache should be cleared
	assert.Equal(t, 0, cache.Len())
}

// TestCacheKey_String tests cache key serialization.
func TestCacheKey_String(t *testing.T) {
	key := CacheKey{
		NPCID:        "npc_001",
		EmotionState: "T50F30S40",
		MessageType:  "greeting",
	}

	keyStr := key.String()
	assert.Equal(t, "npc_001:T50F30S40:greeting", keyStr)
}

// TestMakeCacheKey tests cache key creation from NPC state.
func TestMakeCacheKey(t *testing.T) {
	emotion := manager.EmotionState{
		Trust:  55,
		Fear:   32,
		Stress: 48,
	}

	key := MakeCacheKey("npc_001", emotion, "greeting")

	assert.Equal(t, "npc_001", key.NPCID)
	assert.Equal(t, "T50F30S40", key.EmotionState) // Rounded to buckets of 10
	assert.Equal(t, "greeting", key.MessageType)
}

// TestMakeCacheKey_EmotionBucketing tests emotion state bucketing.
func TestMakeCacheKey_EmotionBucketing(t *testing.T) {
	tests := []struct {
		name          string
		emotion       manager.EmotionState
		expectedState string
	}{
		{
			name:          "exact multiples of 10",
			emotion:       manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			expectedState: "T50F30S40",
		},
		{
			name:          "round down to 50",
			emotion:       manager.EmotionState{Trust: 55, Fear: 35, Stress: 45},
			expectedState: "T50F30S40",
		},
		{
			name:          "round down to 90",
			emotion:       manager.EmotionState{Trust: 99, Fear: 95, Stress: 91},
			expectedState: "T90F90S90",
		},
		{
			name:          "zeros",
			emotion:       manager.EmotionState{Trust: 0, Fear: 0, Stress: 0},
			expectedState: "T0F0S0",
		},
		{
			name:          "single digits",
			emotion:       manager.EmotionState{Trust: 5, Fear: 3, Stress: 8},
			expectedState: "T0F0S0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := MakeCacheKey("npc_001", tt.emotion, "greeting")
			assert.Equal(t, tt.expectedState, key.EmotionState)
		})
	}
}

// TestDetermineMessageType tests message type detection.
func TestDetermineMessageType(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "English greeting",
			message:  "hello",
			expected: "greeting",
		},
		{
			name:     "Chinese greeting",
			message:  "你好",
			expected: "greeting",
		},
		{
			name:     "question with ?",
			message:  "What is your name?",
			expected: "question",
		},
		{
			name:     "Chinese question",
			message:  "你叫什麼名字？",
			expected: "question",
		},
		{
			name:     "command",
			message:  "tell me about yourself",
			expected: "command",
		},
		{
			name:     "statement",
			message:  "I am lost.",
			expected: "statement",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineMessageType(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldCache tests cache eligibility determination.
func TestShouldCache(t *testing.T) {
	tests := []struct {
		name         string
		messageType  string
		response     string
		usedFallback bool
		expected     bool
	}{
		{
			name:         "greeting should cache",
			messageType:  "greeting",
			response:     "Hello!",
			usedFallback: false,
			expected:     true,
		},
		{
			name:         "opening should cache",
			messageType:  "opening",
			response:     "Welcome!",
			usedFallback: false,
			expected:     true,
		},
		{
			name:         "question should cache",
			messageType:  "question",
			response:     "I don't know.",
			usedFallback: false,
			expected:     true,
		},
		{
			name:         "statement should cache",
			messageType:  "statement",
			response:     "I see.",
			usedFallback: false,
			expected:     true,
		},
		{
			name:         "command should not cache",
			messageType:  "command",
			response:     "Here's what I know...",
			usedFallback: false,
			expected:     false,
		},
		{
			name:         "fallback should not cache",
			messageType:  "greeting",
			response:     "Hello!",
			usedFallback: true,
			expected:     false,
		},
		{
			name:         "empty response should not cache",
			messageType:  "greeting",
			response:     "",
			usedFallback: false,
			expected:     false,
		},
		{
			name:         "long response should not cache",
			messageType:  "greeting",
			response:     string(make([]byte, 600)), // 600 chars
			usedFallback: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldCache(tt.messageType, tt.response, tt.usedFallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResponseCache_ConcurrentAccess tests thread safety.
func TestResponseCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)
	require.NoError(t, err)

	const goroutines = 10
	const operations = 100

	done := make(chan bool, goroutines*2)

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				key := CacheKey{
					NPCID:        "npc_001",
					EmotionState: "T50F30S40",
					MessageType:  "greeting",
				}
				cache.Set(key, "Response")
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				key := CacheKey{
					NPCID:        "npc_001",
					EmotionState: "T50F30S40",
					MessageType:  "greeting",
				}
				_, _ = cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines*2; i++ {
		<-done
	}

	// Should have at least one entry
	assert.GreaterOrEqual(t, cache.Len(), 1)
}

// TestResponseCache_MultipleKeys tests different cache keys.
func TestResponseCache_MultipleKeys(t *testing.T) {
	cache, err := NewResponseCache(100, 10*time.Minute)
	require.NoError(t, err)

	// Set different responses for different keys
	keys := []CacheKey{
		{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "greeting"},
		{NPCID: "npc_001", EmotionState: "T60F20S30", MessageType: "greeting"}, // Different emotion
		{NPCID: "npc_001", EmotionState: "T50F30S40", MessageType: "question"}, // Different type
		{NPCID: "npc_002", EmotionState: "T50F30S40", MessageType: "greeting"}, // Different NPC
	}

	responses := []string{
		"Response 1",
		"Response 2",
		"Response 3",
		"Response 4",
	}

	for i, key := range keys {
		cache.Set(key, responses[i])
	}

	// Verify all keys are distinct and retrievable
	for i, key := range keys {
		value, found := cache.Get(key)
		assert.True(t, found, "Key %d should be found", i)
		assert.Equal(t, responses[i], value, "Key %d should have correct response", i)
	}
}

// BenchmarkResponseCache_Set benchmarks cache set operations.
func BenchmarkResponseCache_Set(b *testing.B) {
	cache, _ := NewResponseCache(1000, 10*time.Minute)

	key := CacheKey{
		NPCID:        "npc_001",
		EmotionState: "T50F30S40",
		MessageType:  "greeting",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(key, "Response")
	}
}

// BenchmarkResponseCache_Get benchmarks cache get operations.
func BenchmarkResponseCache_Get(b *testing.B) {
	cache, _ := NewResponseCache(1000, 10*time.Minute)

	key := CacheKey{
		NPCID:        "npc_001",
		EmotionState: "T50F30S40",
		MessageType:  "greeting",
	}

	cache.Set(key, "Response")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(key)
	}
}

// BenchmarkMakeCacheKey benchmarks cache key creation.
func BenchmarkMakeCacheKey(b *testing.B) {
	emotion := manager.EmotionState{
		Trust:  55,
		Fear:   32,
		Stress: 48,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MakeCacheKey("npc_001", emotion, "greeting")
	}
}
