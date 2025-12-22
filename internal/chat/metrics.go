package chat

import (
	"sort"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ChatMetrics tracks chat system performance metrics.
//
// Story 5-8 AC1: NPC response latency < 2 seconds (90th percentile)
// Story 5-8 AC2: Token consumption monitoring and optimization
//
// This structure tracks:
// - Response latencies (p50, p90, p99)
// - Token consumption (input/output)
// - Cache statistics (hits/misses)
type ChatMetrics struct {
	// Latency statistics
	ResponseLatencies []time.Duration // All recorded response latencies
	P50Latency        time.Duration   // 50th percentile
	P90Latency        time.Duration   // 90th percentile
	P99Latency        time.Duration   // 99th percentile

	// Token consumption
	TotalInputTokens  int            // Total input tokens consumed
	TotalOutputTokens int            // Total output tokens consumed
	TokensByNPC       map[string]int // Token consumption by NPC ID

	// Cache statistics
	CacheHits    int     // Number of cache hits
	CacheMisses  int     // Number of cache misses
	CacheHitRate float64 // Hit rate (hits / total)

	mu sync.RWMutex
}

// NewChatMetrics creates a new ChatMetrics instance.
func NewChatMetrics() *ChatMetrics {
	return &ChatMetrics{
		ResponseLatencies: make([]time.Duration, 0, 1000),
		TokensByNPC:       make(map[string]int),
	}
}

// measureLatency measures the latency of an operation and records it.
//
// Story 5-8 AC1: Measure and track response latencies
//
// Parameters:
//   - operation: Description of the operation being measured
//   - fn: The function to measure
//
// Returns:
//   - error from the operation function
func (m *ChatMetrics) measureLatency(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Record latency
	m.ResponseLatencies = append(m.ResponseLatencies, duration)

	// Update percentiles every 10 measurements to reduce overhead
	if len(m.ResponseLatencies)%10 == 0 {
		m.updatePercentiles()
	}

	// Log warning if latency exceeds 2 seconds
	if duration > 2*time.Second {
		logger.Warn("Operation exceeded 2s latency threshold", map[string]interface{}{
			"operation": operation,
			"duration":  duration.String(),
			"threshold": "2s",
		})
	}

	logger.Debug("Operation latency measured", map[string]interface{}{
		"operation": operation,
		"duration":  duration.String(),
	})

	return err
}

// RecordLatency manually records a latency measurement.
//
// This is useful when you can't use measureLatency() with a closure.
func (m *ChatMetrics) RecordLatency(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ResponseLatencies = append(m.ResponseLatencies, duration)

	// Update percentiles every 10 measurements
	if len(m.ResponseLatencies)%10 == 0 {
		m.updatePercentiles()
	}
}

// updatePercentiles calculates and updates p50, p90, p99 latency values.
//
// Story 5-8 AC1: Track 90th percentile latency (must be < 2s)
//
// This method must be called with m.mu held (lock already acquired).
func (m *ChatMetrics) updatePercentiles() {
	if len(m.ResponseLatencies) == 0 {
		return
	}

	// Create a sorted copy
	sorted := make([]time.Duration, len(m.ResponseLatencies))
	copy(sorted, m.ResponseLatencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate percentiles using proper indexing
	// For p50: index = (n-1) * 0.50
	// For p90: index = (n-1) * 0.90
	// For p99: index = (n-1) * 0.99
	n := len(sorted)
	m.P50Latency = sorted[(n-1)*50/100]
	m.P90Latency = sorted[(n-1)*90/100]
	m.P99Latency = sorted[(n-1)*99/100]

	logger.Debug("Latency percentiles updated", map[string]interface{}{
		"p50":         m.P50Latency.String(),
		"p90":         m.P90Latency.String(),
		"p99":         m.P99Latency.String(),
		"sample_size": len(m.ResponseLatencies),
	})
}

// RecordTokenUsage records token consumption for an LLM call.
//
// Story 5-8 AC2: Track input/output token consumption
//
// Parameters:
//   - npcID: The NPC ID (or "system" for non-NPC calls)
//   - inputTokens: Number of input tokens used
//   - outputTokens: Number of output tokens used
func (m *ChatMetrics) RecordTokenUsage(npcID string, inputTokens, outputTokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalInputTokens += inputTokens
	m.TotalOutputTokens += outputTokens

	if npcID != "" {
		m.TokensByNPC[npcID] += inputTokens + outputTokens
	}

	logger.Debug("Token usage recorded", map[string]interface{}{
		"npc_id":        npcID,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"total_tokens":  inputTokens + outputTokens,
	})
}

// RecordCacheHit records a cache hit.
//
// Story 5-8 AC3: Track cache hit rate
func (m *ChatMetrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++
	m.updateCacheHitRate()
}

// RecordCacheMiss records a cache miss.
//
// Story 5-8 AC3: Track cache hit rate
func (m *ChatMetrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheMisses++
	m.updateCacheHitRate()
}

// updateCacheHitRate recalculates the cache hit rate.
//
// This method must be called with m.mu held (lock already acquired).
func (m *ChatMetrics) updateCacheHitRate() {
	total := m.CacheHits + m.CacheMisses
	if total > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(total)
	} else {
		m.CacheHitRate = 0.0
	}
}

// GetMetrics returns a snapshot of the current metrics.
//
// Story 5-8 AC1 & AC2: Provide metrics API for monitoring
//
// Returns a MetricsSnapshot with all current statistics.
func (m *ChatMetrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure percentiles are up to date
	if len(m.ResponseLatencies) > 0 {
		m.mu.RUnlock()
		m.mu.Lock()
		m.updatePercentiles()
		m.mu.Unlock()
		m.mu.RLock()
	}

	// Create a copy of TokensByNPC map
	tokensByNPC := make(map[string]int, len(m.TokensByNPC))
	for k, v := range m.TokensByNPC {
		tokensByNPC[k] = v
	}

	return MetricsSnapshot{
		P50Latency:        m.P50Latency,
		P90Latency:        m.P90Latency,
		P99Latency:        m.P99Latency,
		TotalInputTokens:  m.TotalInputTokens,
		TotalOutputTokens: m.TotalOutputTokens,
		TokensByNPC:       tokensByNPC,
		CacheHits:         m.CacheHits,
		CacheMisses:       m.CacheMisses,
		CacheHitRate:      m.CacheHitRate,
		SampleSize:        len(m.ResponseLatencies),
	}
}

// Reset resets all metrics to zero.
//
// Useful for testing and benchmarking.
func (m *ChatMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ResponseLatencies = make([]time.Duration, 0, 1000)
	m.P50Latency = 0
	m.P90Latency = 0
	m.P99Latency = 0
	m.TotalInputTokens = 0
	m.TotalOutputTokens = 0
	m.TokensByNPC = make(map[string]int)
	m.CacheHits = 0
	m.CacheMisses = 0
	m.CacheHitRate = 0.0

	logger.Debug("Chat metrics reset", nil)
}

// MetricsSnapshot is a point-in-time snapshot of chat metrics.
//
// This is returned by GetMetrics() to provide a consistent view
// without holding locks.
type MetricsSnapshot struct {
	// Latency statistics
	P50Latency time.Duration
	P90Latency time.Duration
	P99Latency time.Duration

	// Token consumption
	TotalInputTokens  int
	TotalOutputTokens int
	TokensByNPC       map[string]int

	// Cache statistics
	CacheHits    int
	CacheMisses  int
	CacheHitRate float64

	// Additional info
	SampleSize int // Number of latency samples
}

// TotalTokens returns the total token consumption (input + output).
func (s MetricsSnapshot) TotalTokens() int {
	return s.TotalInputTokens + s.TotalOutputTokens
}

// P90MeetsTarget returns true if P90 latency is under 2 seconds.
//
// Story 5-8 AC1: Verify P90 latency < 2s
func (s MetricsSnapshot) P90MeetsTarget() bool {
	return s.P90Latency < 2*time.Second
}
