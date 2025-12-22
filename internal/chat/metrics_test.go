package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChatMetrics_NewChatMetrics tests the creation of a new ChatMetrics instance.
func TestChatMetrics_NewChatMetrics(t *testing.T) {
	metrics := NewChatMetrics()

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.ResponseLatencies)
	assert.NotNil(t, metrics.TokensByNPC)
	assert.Equal(t, 0, len(metrics.ResponseLatencies))
	assert.Equal(t, time.Duration(0), metrics.P50Latency)
	assert.Equal(t, time.Duration(0), metrics.P90Latency)
	assert.Equal(t, time.Duration(0), metrics.P99Latency)
}

// TestChatMetrics_RecordLatency tests recording latencies.
func TestChatMetrics_RecordLatency(t *testing.T) {
	metrics := NewChatMetrics()

	// Record some latencies
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordLatency(200 * time.Millisecond)
	metrics.RecordLatency(300 * time.Millisecond)

	assert.Equal(t, 3, len(metrics.ResponseLatencies))
	assert.Equal(t, 100*time.Millisecond, metrics.ResponseLatencies[0])
	assert.Equal(t, 200*time.Millisecond, metrics.ResponseLatencies[1])
	assert.Equal(t, 300*time.Millisecond, metrics.ResponseLatencies[2])
}

// TestChatMetrics_UpdatePercentiles tests percentile calculation.
func TestChatMetrics_UpdatePercentiles(t *testing.T) {
	metrics := NewChatMetrics()

	// Record 100 latencies
	for i := 1; i <= 100; i++ {
		metrics.RecordLatency(time.Duration(i) * time.Millisecond)
	}

	// Get metrics (which triggers percentile update)
	snapshot := metrics.GetMetrics()

	// Verify percentiles
	assert.Equal(t, 50*time.Millisecond, snapshot.P50Latency)
	assert.Equal(t, 90*time.Millisecond, snapshot.P90Latency)
	assert.Equal(t, 99*time.Millisecond, snapshot.P99Latency)
	assert.Equal(t, 100, snapshot.SampleSize)
}

// TestChatMetrics_MeasureLatency tests the measureLatency helper.
func TestChatMetrics_MeasureLatency(t *testing.T) {
	metrics := NewChatMetrics()

	// Measure a function that sleeps for 100ms
	err := metrics.measureLatency("test_operation", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(metrics.ResponseLatencies))
	assert.GreaterOrEqual(t, metrics.ResponseLatencies[0], 100*time.Millisecond)
}

// TestChatMetrics_MeasureLatency_WithError tests measureLatency with error.
func TestChatMetrics_MeasureLatency_WithError(t *testing.T) {
	metrics := NewChatMetrics()

	// Measure a function that returns error
	err := metrics.measureLatency("test_operation", func() error {
		return assert.AnError
	})

	assert.Error(t, err)
	assert.Equal(t, 1, len(metrics.ResponseLatencies))
}

// TestChatMetrics_RecordTokenUsage tests token usage recording.
func TestChatMetrics_RecordTokenUsage(t *testing.T) {
	metrics := NewChatMetrics()

	// Record token usage for different NPCs
	metrics.RecordTokenUsage("npc_001", 100, 50)   // Total: 150
	metrics.RecordTokenUsage("npc_002", 200, 100)  // Total: 300
	metrics.RecordTokenUsage("npc_001", 50, 25)    // Total: 75

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 350, snapshot.TotalInputTokens)    // 100+200+50
	assert.Equal(t, 175, snapshot.TotalOutputTokens)   // 50+100+25
	assert.Equal(t, 525, snapshot.TotalTokens())       // 350+175
	assert.Equal(t, 225, snapshot.TokensByNPC["npc_001"]) // 150+75
	assert.Equal(t, 300, snapshot.TokensByNPC["npc_002"]) // 300
}

// TestChatMetrics_RecordTokenUsage_EmptyNPCID tests token recording without NPC ID.
func TestChatMetrics_RecordTokenUsage_EmptyNPCID(t *testing.T) {
	metrics := NewChatMetrics()

	// Record token usage without NPC ID (system calls)
	metrics.RecordTokenUsage("", 100, 50)

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 100, snapshot.TotalInputTokens)
	assert.Equal(t, 50, snapshot.TotalOutputTokens)
	assert.Equal(t, 0, len(snapshot.TokensByNPC)) // Should not be tracked by NPC
}

// TestChatMetrics_CacheStatistics tests cache hit/miss recording.
func TestChatMetrics_CacheStatistics(t *testing.T) {
	metrics := NewChatMetrics()

	// Record cache hits and misses
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()
	metrics.RecordCacheHit()

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 3, snapshot.CacheHits)
	assert.Equal(t, 1, snapshot.CacheMisses)
	assert.Equal(t, 0.75, snapshot.CacheHitRate) // 3/4 = 0.75
}

// TestChatMetrics_CacheHitRate_NoCalls tests cache hit rate with no calls.
func TestChatMetrics_CacheHitRate_NoCalls(t *testing.T) {
	metrics := NewChatMetrics()

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 0, snapshot.CacheHits)
	assert.Equal(t, 0, snapshot.CacheMisses)
	assert.Equal(t, 0.0, snapshot.CacheHitRate)
}

// TestChatMetrics_Reset tests resetting metrics.
func TestChatMetrics_Reset(t *testing.T) {
	metrics := NewChatMetrics()

	// Record some data
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordTokenUsage("npc_001", 100, 50)
	metrics.RecordCacheHit()

	// Reset
	metrics.Reset()

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 0, snapshot.SampleSize)
	assert.Equal(t, time.Duration(0), snapshot.P50Latency)
	assert.Equal(t, time.Duration(0), snapshot.P90Latency)
	assert.Equal(t, time.Duration(0), snapshot.P99Latency)
	assert.Equal(t, 0, snapshot.TotalInputTokens)
	assert.Equal(t, 0, snapshot.TotalOutputTokens)
	assert.Equal(t, 0, snapshot.CacheHits)
	assert.Equal(t, 0, snapshot.CacheMisses)
	assert.Equal(t, 0.0, snapshot.CacheHitRate)
}

// TestChatMetrics_GetMetrics_ThreadSafety tests concurrent access.
func TestChatMetrics_GetMetrics_ThreadSafety(t *testing.T) {
	metrics := NewChatMetrics()

	// Concurrently record latencies and get metrics
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			metrics.RecordLatency(time.Duration(i) * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = metrics.GetMetrics()
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done

	snapshot := metrics.GetMetrics()
	assert.Equal(t, 100, snapshot.SampleSize)
}

// TestMetricsSnapshot_TotalTokens tests the TotalTokens helper.
func TestMetricsSnapshot_TotalTokens(t *testing.T) {
	snapshot := MetricsSnapshot{
		TotalInputTokens:  100,
		TotalOutputTokens: 50,
	}

	assert.Equal(t, 150, snapshot.TotalTokens())
}

// TestMetricsSnapshot_P90MeetsTarget tests the P90 target checker.
func TestMetricsSnapshot_P90MeetsTarget(t *testing.T) {
	tests := []struct {
		name       string
		p90Latency time.Duration
		expected   bool
	}{
		{
			name:       "meets target (1s)",
			p90Latency: 1 * time.Second,
			expected:   true,
		},
		{
			name:       "meets target (1.9s)",
			p90Latency: 1900 * time.Millisecond,
			expected:   true,
		},
		{
			name:       "exactly 2s (does not meet)",
			p90Latency: 2 * time.Second,
			expected:   false,
		},
		{
			name:       "exceeds target (3s)",
			p90Latency: 3 * time.Second,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := MetricsSnapshot{
				P90Latency: tt.p90Latency,
			}

			assert.Equal(t, tt.expected, snapshot.P90MeetsTarget())
		})
	}
}

// TestChatMetrics_PercentileCalculation_EdgeCases tests edge cases in percentile calculation.
func TestChatMetrics_PercentileCalculation_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		latencies  []time.Duration
		expectedP50 time.Duration
		expectedP90 time.Duration
		expectedP99 time.Duration
	}{
		{
			name:        "single value",
			latencies:   []time.Duration{100 * time.Millisecond},
			expectedP50: 100 * time.Millisecond,
			expectedP90: 100 * time.Millisecond,
			expectedP99: 100 * time.Millisecond,
		},
		{
			name:        "two values",
			latencies:   []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
			expectedP50: 100 * time.Millisecond,  // (2-1)*50/100 = 0 -> index 0
			expectedP90: 100 * time.Millisecond,  // (2-1)*90/100 = 0 -> index 0
			expectedP99: 100 * time.Millisecond,  // (2-1)*99/100 = 0 -> index 0
		},
		{
			name:        "ten values",
			latencies:   []time.Duration{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectedP50: 5,  // (10-1)*50/100 = 4 -> index 4 -> value 5
			expectedP90: 9,  // (10-1)*90/100 = 8 -> index 8 -> value 9
			expectedP99: 9,  // (10-1)*99/100 = 8 -> index 8 -> value 9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewChatMetrics()

			for _, latency := range tt.latencies {
				metrics.RecordLatency(latency)
			}

			snapshot := metrics.GetMetrics()

			assert.Equal(t, tt.expectedP50, snapshot.P50Latency)
			assert.Equal(t, tt.expectedP90, snapshot.P90Latency)
			assert.Equal(t, tt.expectedP99, snapshot.P99Latency)
		})
	}
}

// TestChatMetrics_LargeDataset tests metrics with a large dataset.
func TestChatMetrics_LargeDataset(t *testing.T) {
	metrics := NewChatMetrics()

	// Record 10,000 latencies
	for i := 1; i <= 10000; i++ {
		metrics.RecordLatency(time.Duration(i) * time.Millisecond)
	}

	snapshot := metrics.GetMetrics()

	assert.Equal(t, 10000, snapshot.SampleSize)
	assert.Equal(t, 5000*time.Millisecond, snapshot.P50Latency)
	assert.Equal(t, 9000*time.Millisecond, snapshot.P90Latency)
	assert.Equal(t, 9900*time.Millisecond, snapshot.P99Latency)
}

// BenchmarkChatMetrics_RecordLatency benchmarks latency recording.
func BenchmarkChatMetrics_RecordLatency(b *testing.B) {
	metrics := NewChatMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordLatency(time.Duration(i) * time.Millisecond)
	}
}

// BenchmarkChatMetrics_RecordTokenUsage benchmarks token usage recording.
func BenchmarkChatMetrics_RecordTokenUsage(b *testing.B) {
	metrics := NewChatMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordTokenUsage("npc_001", 100, 50)
	}
}

// BenchmarkChatMetrics_GetMetrics benchmarks metrics retrieval.
func BenchmarkChatMetrics_GetMetrics(b *testing.B) {
	metrics := NewChatMetrics()

	// Populate with some data
	for i := 0; i < 1000; i++ {
		metrics.RecordLatency(time.Duration(i) * time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.GetMetrics()
	}
}

// TestChatMetrics_ConcurrentAccess tests concurrent read/write access.
func TestChatMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewChatMetrics()

	const goroutines = 10
	const operations = 100

	done := make(chan bool, goroutines*3)

	// Concurrent latency recording
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				metrics.RecordLatency(time.Duration(j) * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Concurrent token recording
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				metrics.RecordTokenUsage("npc_001", 10, 5)
			}
			done <- true
		}(i)
	}

	// Concurrent metrics reading
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				_ = metrics.GetMetrics()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines*3; i++ {
		<-done
	}

	snapshot := metrics.GetMetrics()

	// Verify expected totals
	require.Equal(t, goroutines*operations, snapshot.SampleSize)
	require.Equal(t, goroutines*operations*10, snapshot.TotalInputTokens)
	require.Equal(t, goroutines*operations*5, snapshot.TotalOutputTokens)
}
