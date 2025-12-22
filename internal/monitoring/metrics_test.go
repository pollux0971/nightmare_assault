package monitoring

import (
	"errors"
	"testing"
	"time"
)

// ==========================================================================
// Story 8.4: Performance Monitoring System - Tests
// ==========================================================================

func TestNewMetricsCollector(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		maxRecords int
		wantMax    int
	}{
		{
			name:       "enabled with default max",
			enabled:    true,
			maxRecords: 0,
			wantMax:    10000,
		},
		{
			name:       "enabled with custom max",
			enabled:    true,
			maxRecords: 5000,
			wantMax:    5000,
		},
		{
			name:       "disabled",
			enabled:    false,
			maxRecords: 1000,
			wantMax:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.enabled, tt.maxRecords)
			if mc == nil {
				t.Fatal("NewMetricsCollector returned nil")
			}
			if mc.enabled != tt.enabled {
				t.Errorf("enabled = %v, want %v", mc.enabled, tt.enabled)
			}
			if mc.maxRecords != tt.wantMax {
				t.Errorf("maxRecords = %v, want %v", mc.maxRecords, tt.wantMax)
			}
			if len(mc.agentMetrics) == 0 {
				t.Error("agentMetrics not initialized")
			}
		})
	}
}

func TestRecordTokenUsage(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC1: Token consumption tracking
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)

	metrics := mc.GetAgentMetrics(AgentTypeNarration)
	if metrics.TotalCalls != 1 {
		t.Errorf("TotalCalls = %v, want 1", metrics.TotalCalls)
	}
	if metrics.SuccessfulCalls != 1 {
		t.Errorf("SuccessfulCalls = %v, want 1", metrics.SuccessfulCalls)
	}
	if metrics.PromptTokens != 500 {
		t.Errorf("PromptTokens = %v, want 500", metrics.PromptTokens)
	}
	if metrics.ResponseTokens != 300 {
		t.Errorf("ResponseTokens = %v, want 300", metrics.ResponseTokens)
	}
	if metrics.TotalTokens != 800 {
		t.Errorf("TotalTokens = %v, want 800", metrics.TotalTokens)
	}

	// Test failed call
	mc.RecordTokenUsage(AgentTypeChoice, 100, 0, "gpt-4", false, errors.New("API error"))

	metrics = mc.GetAgentMetrics(AgentTypeChoice)
	if metrics.TotalCalls != 1 {
		t.Errorf("TotalCalls = %v, want 1", metrics.TotalCalls)
	}
	if metrics.FailedCalls != 1 {
		t.Errorf("FailedCalls = %v, want 1", metrics.FailedCalls)
	}
	if metrics.TotalTokens != 0 {
		t.Errorf("TotalTokens = %v, want 0 (failed calls should not count)", metrics.TotalTokens)
	}
}

func TestRecordTokenUsageDisabled(t *testing.T) {
	mc := NewMetricsCollector(false, 10000)

	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)

	// Should not record when disabled
	if len(mc.tokenUsages) != 0 {
		t.Errorf("tokenUsages length = %v, want 0 (disabled)", len(mc.tokenUsages))
	}
}

func TestRecordLLMLatency(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC2: LLM latency monitoring
	duration := 1500 * time.Millisecond
	mc.RecordLLMLatency(AgentTypeJudge, duration, true, nil)

	metrics := mc.GetAgentMetrics(AgentTypeJudge)
	if len(metrics.LatencyRecords) != 1 {
		t.Errorf("LatencyRecords length = %v, want 1", len(metrics.LatencyRecords))
	}
	if metrics.LatencyRecords[0] != duration {
		t.Errorf("LatencyRecords[0] = %v, want %v", metrics.LatencyRecords[0], duration)
	}
	if metrics.AverageLatency != duration {
		t.Errorf("AverageLatency = %v, want %v", metrics.AverageLatency, duration)
	}

	// Add more records to test average calculation
	mc.RecordLLMLatency(AgentTypeJudge, 2500*time.Millisecond, true, nil)
	metrics = mc.GetAgentMetrics(AgentTypeJudge)

	expectedAvg := (1500 + 2500) / 2
	actualAvg := metrics.AverageLatency.Milliseconds()
	if actualAvg != int64(expectedAvg) {
		t.Errorf("AverageLatency = %v ms, want %v ms", actualAvg, expectedAvg)
	}
}

func TestRecordChatLatency(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC3: Chat response latency monitoring
	duration := 1800 * time.Millisecond
	mc.RecordChatLatency(AgentTypeChat, duration, true, nil)

	if len(mc.chatLatencies) != 1 {
		t.Errorf("chatLatencies length = %v, want 1", len(mc.chatLatencies))
	}

	record := mc.chatLatencies[0]
	if record.Duration != duration {
		t.Errorf("Duration = %v, want %v", record.Duration, duration)
	}
	if record.AgentType != AgentTypeChat {
		t.Errorf("AgentType = %v, want %v", record.AgentType, AgentTypeChat)
	}
	if !record.Success {
		t.Error("Success should be true")
	}
}

func TestRecordCacheHitMiss(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Record cache hits and misses
	mc.RecordCacheHit(AgentTypeNPC)
	mc.RecordCacheHit(AgentTypeNPC)
	mc.RecordCacheMiss(AgentTypeNPC)

	metrics := mc.GetAgentMetrics(AgentTypeNPC)
	if metrics.CacheHits != 2 {
		t.Errorf("CacheHits = %v, want 2", metrics.CacheHits)
	}
	if metrics.CacheMisses != 1 {
		t.Errorf("CacheMisses = %v, want 1", metrics.CacheMisses)
	}

	hitRate := mc.GetCacheHitRate(AgentTypeNPC)
	expectedRate := 2.0 / 3.0
	if hitRate < expectedRate-0.01 || hitRate > expectedRate+0.01 {
		t.Errorf("CacheHitRate = %v, want %v", hitRate, expectedRate)
	}
}

func TestGetLLMLatencyStats(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC2: Calculate p50/p90/p99
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
		600 * time.Millisecond,
		700 * time.Millisecond,
		800 * time.Millisecond,
		900 * time.Millisecond,
		1000 * time.Millisecond,
	}

	for _, d := range durations {
		mc.RecordLLMLatency(AgentTypeDream, d, true, nil)
	}

	stats := mc.GetLLMLatencyStats(AgentTypeDream)

	if stats.Count != 10 {
		t.Errorf("Count = %v, want 10", stats.Count)
	}
	if stats.Min != 100*time.Millisecond {
		t.Errorf("Min = %v, want 100ms", stats.Min)
	}
	if stats.Max != 1000*time.Millisecond {
		t.Errorf("Max = %v, want 1000ms", stats.Max)
	}

	// P50 should be around 500ms (50th percentile)
	if stats.P50 < 400*time.Millisecond || stats.P50 > 600*time.Millisecond {
		t.Errorf("P50 = %v, want around 500ms", stats.P50)
	}

	// P90 should be around 900ms (90th percentile)
	if stats.P90 < 800*time.Millisecond || stats.P90 > 1000*time.Millisecond {
		t.Errorf("P90 = %v, want around 900ms", stats.P90)
	}

	// P99 should be around 990ms (99th percentile)
	if stats.P99 < 900*time.Millisecond || stats.P99 > 1000*time.Millisecond {
		t.Errorf("P99 = %v, want around 990ms", stats.P99)
	}
}

func TestGetChatLatencyStats(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC3: Chat latency stats
	mc.RecordChatLatency(AgentTypeChat, 1000*time.Millisecond, true, nil)
	mc.RecordChatLatency(AgentTypeChat, 2000*time.Millisecond, true, nil)
	mc.RecordChatLatency(AgentTypeChat, 3000*time.Millisecond, true, nil)

	stats := mc.GetChatLatencyStats(AgentTypeChat)

	if stats.Count != 3 {
		t.Errorf("Count = %v, want 3", stats.Count)
	}
	if stats.Min != 1000*time.Millisecond {
		t.Errorf("Min = %v, want 1000ms", stats.Min)
	}
	if stats.Max != 3000*time.Millisecond {
		t.Errorf("Max = %v, want 3000ms", stats.Max)
	}
}

func TestGetTotalTokenUsage(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// AC1: Total token tracking
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	mc.RecordTokenUsage(AgentTypeChoice, 200, 100, "gpt-4", true, nil)
	mc.RecordTokenUsage(AgentTypeJudge, 300, 200, "gpt-4", true, nil)

	promptTokens, responseTokens, totalTokens := mc.GetTotalTokenUsage()

	if promptTokens != 1000 {
		t.Errorf("PromptTokens = %v, want 1000", promptTokens)
	}
	if responseTokens != 600 {
		t.Errorf("ResponseTokens = %v, want 600", responseTokens)
	}
	if totalTokens != 1600 {
		t.Errorf("TotalTokens = %v, want 1600", totalTokens)
	}
}

func TestGetTokenUsageByAgent(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	mc.RecordTokenUsage(AgentTypeNarration, 400, 200, "gpt-4", true, nil)
	mc.RecordTokenUsage(AgentTypeChoice, 200, 100, "gpt-4", true, nil)

	usage := mc.GetTokenUsageByAgent()

	if usage[AgentTypeNarration] != 1400 {
		t.Errorf("AgentTypeNarration = %v, want 1400", usage[AgentTypeNarration])
	}
	if usage[AgentTypeChoice] != 300 {
		t.Errorf("AgentTypeChoice = %v, want 300", usage[AgentTypeChoice])
	}
}

func TestTrimRecords(t *testing.T) {
	mc := NewMetricsCollector(true, 100) // Small max for testing

	// Add more than max records
	for i := 0; i < 150; i++ {
		mc.RecordTokenUsage(AgentTypeNarration, 100, 100, "gpt-4", true, nil)
	}

	// Should be trimmed to maxRecords
	if len(mc.tokenUsages) > mc.maxRecords {
		t.Errorf("tokenUsages length = %v, want <= %v", len(mc.tokenUsages), mc.maxRecords)
	}
}

func TestEnableDisable(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	if !mc.IsEnabled() {
		t.Error("Should be enabled initially")
	}

	mc.Disable()
	if mc.IsEnabled() {
		t.Error("Should be disabled after Disable()")
	}

	// Recording should not work when disabled
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	if len(mc.tokenUsages) != 0 {
		t.Error("Should not record when disabled")
	}

	mc.Enable()
	if !mc.IsEnabled() {
		t.Error("Should be enabled after Enable()")
	}

	// Recording should work when re-enabled
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	if len(mc.tokenUsages) != 1 {
		t.Error("Should record when re-enabled")
	}
}

func TestReset(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Add some data
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	mc.RecordLLMLatency(AgentTypeChoice, 1500*time.Millisecond, true, nil)
	mc.RecordChatLatency(AgentTypeChat, 2000*time.Millisecond, true, nil)

	// Reset
	mc.Reset()

	// Check all data is cleared
	if len(mc.tokenUsages) != 0 {
		t.Errorf("tokenUsages length = %v, want 0", len(mc.tokenUsages))
	}
	if len(mc.latencyRecords) != 0 {
		t.Errorf("latencyRecords length = %v, want 0", len(mc.latencyRecords))
	}
	if len(mc.chatLatencies) != 0 {
		t.Errorf("chatLatencies length = %v, want 0", len(mc.chatLatencies))
	}

	metrics := mc.GetAgentMetrics(AgentTypeNarration)
	if metrics.TotalCalls != 0 {
		t.Errorf("TotalCalls = %v, want 0", metrics.TotalCalls)
	}
}

func TestCalculatePercentiles(t *testing.T) {
	// Empty case
	stats := calculatePercentiles([]time.Duration{})
	if stats.Count != 0 {
		t.Errorf("Empty percentiles Count = %v, want 0", stats.Count)
	}

	// Single value
	durations := []time.Duration{500 * time.Millisecond}
	stats = calculatePercentiles(durations)
	if stats.Count != 1 {
		t.Errorf("Count = %v, want 1", stats.Count)
	}
	if stats.P50 != 500*time.Millisecond {
		t.Errorf("P50 = %v, want 500ms", stats.P50)
	}
	if stats.Min != 500*time.Millisecond {
		t.Errorf("Min = %v, want 500ms", stats.Min)
	}
	if stats.Max != 500*time.Millisecond {
		t.Errorf("Max = %v, want 500ms", stats.Max)
	}
}

func TestGetGlobalCollector(t *testing.T) {
	collector1 := GetGlobalCollector()
	collector2 := GetGlobalCollector()

	// Should return same instance (singleton)
	if collector1 != collector2 {
		t.Error("GetGlobalCollector should return same instance")
	}

	if !collector1.IsEnabled() {
		t.Error("Global collector should be enabled by default")
	}
}

func TestLogSummary(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Add some test data
	mc.RecordTokenUsage(AgentTypeNarration, 500, 300, "gpt-4", true, nil)
	mc.RecordLLMLatency(AgentTypeNarration, 1500*time.Millisecond, true, nil)
	mc.RecordCacheHit(AgentTypeNarration)
	mc.RecordCacheMiss(AgentTypeNarration)

	// AC4: Should not panic when logging
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogSummary panicked: %v", r)
		}
	}()

	mc.LogSummary()
}

func TestHighTokenUsageWarning(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Should trigger warning for high token usage (> 5000)
	mc.RecordTokenUsage(AgentTypeNarration, 6000, 2000, "gpt-4", true, nil)

	metrics := mc.GetAgentMetrics(AgentTypeNarration)
	if metrics.TotalTokens != 8000 {
		t.Errorf("TotalTokens = %v, want 8000", metrics.TotalTokens)
	}
}

func TestSlowLLMCallWarning(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Should trigger warning for slow LLM call (> 5s)
	mc.RecordLLMLatency(AgentTypeNarration, 6*time.Second, true, nil)

	metrics := mc.GetAgentMetrics(AgentTypeNarration)
	if len(metrics.LatencyRecords) != 1 {
		t.Errorf("LatencyRecords length = %v, want 1", len(metrics.LatencyRecords))
	}
}

func TestSlowChatResponseWarning(t *testing.T) {
	mc := NewMetricsCollector(true, 10000)

	// Should trigger warning for slow chat response (> 2s)
	mc.RecordChatLatency(AgentTypeChat, 3*time.Second, true, nil)

	if len(mc.chatLatencies) != 1 {
		t.Errorf("chatLatencies length = %v, want 1", len(mc.chatLatencies))
	}
}
