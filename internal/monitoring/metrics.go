package monitoring

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 8.4: Performance Monitoring System
// ==========================================================================
// AC1: Token consumption tracking (每個 Agent 調用)
// AC2: LLM延遲監控（p50/p90/p99）
// AC3: 聊天回應延遲監控
// AC4: 記錄到日誌與指標系統
// ==========================================================================

// AgentType represents different types of agents in the system.
type AgentType string

const (
	AgentTypeNarration AgentType = "narration"
	AgentTypeChoice    AgentType = "choice"
	AgentTypeJudge     AgentType = "judge"
	AgentTypeNPC       AgentType = "npc"
	AgentTypeDream     AgentType = "dream"
	AgentTypeOpening   AgentType = "opening"
	AgentTypeEnding    AgentType = "ending"
	AgentTypeChat      AgentType = "chat"
)

// MetricType represents different types of metrics.
type MetricType string

const (
	MetricTypeTokenUsage    MetricType = "token_usage"
	MetricTypeLLMLatency    MetricType = "llm_latency"
	MetricTypeChatLatency   MetricType = "chat_latency"
	MetricTypeAgentCall     MetricType = "agent_call"
	MetricTypeError         MetricType = "error"
	MetricTypeCacheHit      MetricType = "cache_hit"
	MetricTypeCacheMiss     MetricType = "cache_miss"
)

// TokenUsage represents token consumption for a single API call.
type TokenUsage struct {
	Timestamp      time.Time
	AgentType      AgentType
	PromptTokens   int
	ResponseTokens int
	TotalTokens    int
	Model          string
	Success        bool
	Error          string
}

// LatencyRecord represents a latency measurement.
type LatencyRecord struct {
	Timestamp time.Time
	AgentType AgentType
	Duration  time.Duration
	Success   bool
	Error     string
}

// PercentileStats represents percentile statistics.
type PercentileStats struct {
	P50   time.Duration
	P90   time.Duration
	P99   time.Duration
	Min   time.Duration
	Max   time.Duration
	Mean  time.Duration
	Count int
}

// AgentMetrics represents metrics for a specific agent type.
type AgentMetrics struct {
	TotalCalls       int
	SuccessfulCalls  int
	FailedCalls      int
	TotalTokens      int64
	PromptTokens     int64
	ResponseTokens   int64
	LatencyRecords   []time.Duration
	LastCallTime     time.Time
	AverageLatency   time.Duration
	CacheHits        int
	CacheMisses      int
}

// MetricsCollector collects and manages performance metrics.
type MetricsCollector struct {
	mu              sync.RWMutex
	startTime       time.Time
	tokenUsages     []TokenUsage
	latencyRecords  []LatencyRecord
	chatLatencies   []LatencyRecord
	agentMetrics    map[AgentType]*AgentMetrics
	enabled         bool
	maxRecords      int // Maximum number of records to keep in memory
}

// NewMetricsCollector creates a new MetricsCollector.
// AC4: Record to logging and metrics system
func NewMetricsCollector(enabled bool, maxRecords int) *MetricsCollector {
	if maxRecords <= 0 {
		maxRecords = 10000 // Default: keep last 10k records
	}

	mc := &MetricsCollector{
		startTime:      time.Now(),
		tokenUsages:    make([]TokenUsage, 0, 1000),
		latencyRecords: make([]LatencyRecord, 0, 1000),
		chatLatencies:  make([]LatencyRecord, 0, 1000),
		agentMetrics:   make(map[AgentType]*AgentMetrics),
		enabled:        enabled,
		maxRecords:     maxRecords,
	}

	// Initialize metrics for all agent types
	for _, agentType := range []AgentType{
		AgentTypeNarration, AgentTypeChoice, AgentTypeJudge,
		AgentTypeNPC, AgentTypeDream, AgentTypeOpening,
		AgentTypeEnding, AgentTypeChat,
	} {
		mc.agentMetrics[agentType] = &AgentMetrics{
			LatencyRecords: make([]time.Duration, 0, 100),
		}
	}

	logger.Info("MetricsCollector initialized", map[string]interface{}{
		"enabled":    enabled,
		"maxRecords": maxRecords,
	})
	return mc
}

// RecordTokenUsage records token consumption for an API call.
// AC1: Token consumption tracking
func (mc *MetricsCollector) RecordTokenUsage(agentType AgentType, promptTokens, responseTokens int, model string, success bool, err error) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	usage := TokenUsage{
		Timestamp:      time.Now(),
		AgentType:      agentType,
		PromptTokens:   promptTokens,
		ResponseTokens: responseTokens,
		TotalTokens:    promptTokens + responseTokens,
		Model:          model,
		Success:        success,
		Error:          errMsg,
	}

	mc.tokenUsages = append(mc.tokenUsages, usage)
	mc.trimRecords()

	// Update agent metrics
	metrics := mc.agentMetrics[agentType]
	if success {
		metrics.SuccessfulCalls++
		metrics.TotalTokens += int64(usage.TotalTokens)
		metrics.PromptTokens += int64(promptTokens)
		metrics.ResponseTokens += int64(responseTokens)
	} else {
		metrics.FailedCalls++
	}
	metrics.TotalCalls++

	// Log significant token usage
	if usage.TotalTokens > 5000 {
		logger.Warn("High token usage detected", map[string]interface{}{
			"agent":       agentType,
			"total_tokens": usage.TotalTokens,
			"model":       model,
		})
	}

	logger.Debug("Token usage recorded", map[string]interface{}{
		"agent":           agentType,
		"prompt_tokens":   promptTokens,
		"response_tokens": responseTokens,
		"total_tokens":    usage.TotalTokens,
		"model":           model,
		"success":         success,
	})
}

// RecordLLMLatency records LLM API call latency.
// AC2: LLM latency monitoring (p50/p90/p99)
func (mc *MetricsCollector) RecordLLMLatency(agentType AgentType, duration time.Duration, success bool, err error) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	record := LatencyRecord{
		Timestamp: time.Now(),
		AgentType: agentType,
		Duration:  duration,
		Success:   success,
		Error:     errMsg,
	}

	mc.latencyRecords = append(mc.latencyRecords, record)
	mc.trimRecords()

	// Update agent metrics
	metrics := mc.agentMetrics[agentType]
	metrics.LatencyRecords = append(metrics.LatencyRecords, duration)
	metrics.LastCallTime = time.Now()

	// Recalculate average latency
	if len(metrics.LatencyRecords) > 0 {
		var total time.Duration
		for _, d := range metrics.LatencyRecords {
			total += d
		}
		metrics.AverageLatency = total / time.Duration(len(metrics.LatencyRecords))
	}

	// Log slow LLM calls
	if duration > 5*time.Second {
		logger.Warn("Slow LLM call detected", map[string]interface{}{
			"agent":       agentType,
			"duration_ms": duration.Milliseconds(),
			"success":     success,
		})
	}

	logger.Debug("LLM latency recorded", map[string]interface{}{
		"agent":       agentType,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	})
}

// RecordChatLatency records chat response latency.
// AC3: Chat response latency monitoring
func (mc *MetricsCollector) RecordChatLatency(agentType AgentType, duration time.Duration, success bool, err error) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	record := LatencyRecord{
		Timestamp: time.Now(),
		AgentType: agentType,
		Duration:  duration,
		Success:   success,
		Error:     errMsg,
	}

	mc.chatLatencies = append(mc.chatLatencies, record)
	mc.trimRecords()

	// Log slow chat responses
	if duration > 2*time.Second {
		logger.Warn("Slow chat response detected", map[string]interface{}{
			"agent":       agentType,
			"duration_ms": duration.Milliseconds(),
			"success":     success,
		})
	}

	logger.Debug("Chat latency recorded", map[string]interface{}{
		"agent":       agentType,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	})
}

// RecordCacheHit records a cache hit event.
func (mc *MetricsCollector) RecordCacheHit(agentType AgentType) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics := mc.agentMetrics[agentType]
	metrics.CacheHits++

	logger.Debug("Cache hit recorded", map[string]interface{}{
		"agent": agentType,
	})
}

// RecordCacheMiss records a cache miss event.
func (mc *MetricsCollector) RecordCacheMiss(agentType AgentType) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics := mc.agentMetrics[agentType]
	metrics.CacheMisses++

	logger.Debug("Cache miss recorded", map[string]interface{}{
		"agent": agentType,
	})
}

// GetLLMLatencyStats calculates percentile statistics for LLM latency.
// AC2: Calculate p50/p90/p99
func (mc *MetricsCollector) GetLLMLatencyStats(agentType AgentType) *PercentileStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, ok := mc.agentMetrics[agentType]
	if !ok || len(metrics.LatencyRecords) == 0 {
		return &PercentileStats{}
	}

	return calculatePercentiles(metrics.LatencyRecords)
}

// GetChatLatencyStats calculates percentile statistics for chat latency.
// AC3: Chat response latency stats
func (mc *MetricsCollector) GetChatLatencyStats(agentType AgentType) *PercentileStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	durations := make([]time.Duration, 0)
	for _, record := range mc.chatLatencies {
		if record.AgentType == agentType && record.Success {
			durations = append(durations, record.Duration)
		}
	}

	if len(durations) == 0 {
		return &PercentileStats{}
	}

	return calculatePercentiles(durations)
}

// GetAgentMetrics returns metrics for a specific agent type.
func (mc *MetricsCollector) GetAgentMetrics(agentType AgentType) *AgentMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, ok := mc.agentMetrics[agentType]
	if !ok {
		return &AgentMetrics{}
	}

	// Return a copy to avoid race conditions
	copy := *metrics
	copy.LatencyRecords = append([]time.Duration{}, metrics.LatencyRecords...)
	return &copy
}

// GetTotalTokenUsage returns total token usage across all agents.
// AC1: Total token tracking
func (mc *MetricsCollector) GetTotalTokenUsage() (promptTokens, responseTokens, totalTokens int64) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, metrics := range mc.agentMetrics {
		promptTokens += metrics.PromptTokens
		responseTokens += metrics.ResponseTokens
		totalTokens += metrics.TotalTokens
	}

	return
}

// GetTokenUsageByAgent returns token usage grouped by agent type.
func (mc *MetricsCollector) GetTokenUsageByAgent() map[AgentType]int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[AgentType]int64)
	for agentType, metrics := range mc.agentMetrics {
		result[agentType] = metrics.TotalTokens
	}

	return result
}

// GetCacheHitRate returns cache hit rate for a specific agent type.
func (mc *MetricsCollector) GetCacheHitRate(agentType AgentType) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, ok := mc.agentMetrics[agentType]
	if !ok {
		return 0.0
	}

	total := metrics.CacheHits + metrics.CacheMisses
	if total == 0 {
		return 0.0
	}

	return float64(metrics.CacheHits) / float64(total)
}

// LogSummary logs a summary of all collected metrics.
// AC4: Logging to metrics system
func (mc *MetricsCollector) LogSummary() {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	uptime := time.Since(mc.startTime)
	promptTokens, responseTokens, totalTokens := mc.getTotalTokenUsageUnsafe()

	logger.Info("=== Performance Metrics Summary ===", map[string]interface{}{
		"uptime":          uptime.String(),
		"total_tokens":    totalTokens,
		"prompt_tokens":   promptTokens,
		"response_tokens": responseTokens,
	})

	for agentType, metrics := range mc.agentMetrics {
		if metrics.TotalCalls == 0 {
			continue
		}

		stats := calculatePercentiles(metrics.LatencyRecords)
		cacheHitRate := 0.0
		if metrics.CacheHits+metrics.CacheMisses > 0 {
			cacheHitRate = float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses)
		}

		logger.Info(fmt.Sprintf("Agent: %s", agentType), map[string]interface{}{
			"total_calls":      metrics.TotalCalls,
			"successful_calls": metrics.SuccessfulCalls,
			"failed_calls":     metrics.FailedCalls,
			"total_tokens":     metrics.TotalTokens,
			"avg_latency_ms":   metrics.AverageLatency.Milliseconds(),
			"p50_ms":           stats.P50.Milliseconds(),
			"p90_ms":           stats.P90.Milliseconds(),
			"p99_ms":           stats.P99.Milliseconds(),
			"cache_hit_rate":   fmt.Sprintf("%.2f%%", cacheHitRate*100),
		})
	}
}

// Reset resets all metrics.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.startTime = time.Now()
	mc.tokenUsages = make([]TokenUsage, 0, 1000)
	mc.latencyRecords = make([]LatencyRecord, 0, 1000)
	mc.chatLatencies = make([]LatencyRecord, 0, 1000)

	for agentType := range mc.agentMetrics {
		mc.agentMetrics[agentType] = &AgentMetrics{
			LatencyRecords: make([]time.Duration, 0, 100),
		}
	}

	logger.Info("Metrics reset", nil)
}

// Enable enables metrics collection.
func (mc *MetricsCollector) Enable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = true
	logger.Info("Metrics collection enabled", nil)
}

// Disable disables metrics collection.
func (mc *MetricsCollector) Disable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = false
	logger.Info("Metrics collection disabled", nil)
}

// IsEnabled returns whether metrics collection is enabled.
func (mc *MetricsCollector) IsEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.enabled
}

// Helper functions

// trimRecords trims records to keep memory usage bounded.
func (mc *MetricsCollector) trimRecords() {
	if len(mc.tokenUsages) > mc.maxRecords {
		mc.tokenUsages = mc.tokenUsages[len(mc.tokenUsages)-mc.maxRecords:]
	}
	if len(mc.latencyRecords) > mc.maxRecords {
		mc.latencyRecords = mc.latencyRecords[len(mc.latencyRecords)-mc.maxRecords:]
	}
	if len(mc.chatLatencies) > mc.maxRecords {
		mc.chatLatencies = mc.chatLatencies[len(mc.chatLatencies)-mc.maxRecords:]
	}

	// Trim agent metrics latency records
	for _, metrics := range mc.agentMetrics {
		if len(metrics.LatencyRecords) > 1000 {
			metrics.LatencyRecords = metrics.LatencyRecords[len(metrics.LatencyRecords)-1000:]
		}
	}
}

// getTotalTokenUsageUnsafe returns total token usage without locking (caller must hold lock).
func (mc *MetricsCollector) getTotalTokenUsageUnsafe() (promptTokens, responseTokens, totalTokens int64) {
	for _, metrics := range mc.agentMetrics {
		promptTokens += metrics.PromptTokens
		responseTokens += metrics.ResponseTokens
		totalTokens += metrics.TotalTokens
	}
	return
}

// calculatePercentiles calculates percentile statistics from durations.
// AC2: Calculate p50/p90/p99
func calculatePercentiles(durations []time.Duration) *PercentileStats {
	if len(durations) == 0 {
		return &PercentileStats{}
	}

	// Create a copy and sort
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate percentiles
	p50 := sorted[len(sorted)*50/100]
	p90 := sorted[len(sorted)*90/100]
	p99 := sorted[len(sorted)*99/100]
	min := sorted[0]
	max := sorted[len(sorted)-1]

	// Calculate mean
	var total time.Duration
	for _, d := range sorted {
		total += d
	}
	mean := total / time.Duration(len(sorted))

	return &PercentileStats{
		P50:   p50,
		P90:   p90,
		P99:   p99,
		Min:   min,
		Max:   max,
		Mean:  mean,
		Count: len(sorted),
	}
}

// Global metrics collector instance
var globalCollector *MetricsCollector
var once sync.Once

// GetGlobalCollector returns the global metrics collector instance.
func GetGlobalCollector() *MetricsCollector {
	once.Do(func() {
		globalCollector = NewMetricsCollector(true, 10000)
	})
	return globalCollector
}
