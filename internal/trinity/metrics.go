package trinity

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// Story 9-5: Trinity Metrics Collection
// AC1: Create TrinityMetrics structure
// AC2: Implement metrics collection functions
// AC3: Integrate with TrinityRouter
// AC4: Achieve >80% test coverage

// TierMetrics tracks metrics for a single tier
type TierMetrics struct {
	// Request tracking
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64

	// Response time tracking
	TotalDuration    time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	DurationSamples  []time.Duration

	// Error tracking
	LastError     error
	LastErrorTime time.Time
}

// TrinityMetrics collects and manages Trinity system metrics
// AC1: TrinityMetrics structure
type TrinityMetrics struct {
	mu sync.RWMutex

	// Metrics for each tier
	tierMetrics map[TierLevel]*TierMetrics

	// Tier transition tracking
	upgradeCount   map[string]int64 // "from_to" -> count
	downgradeCount map[string]int64 // "from_to" -> count

	// System-wide metrics
	startTime      time.Time
	totalRequests  int64
	maxSamples     int // Maximum number of duration samples to keep per tier
}

// NewTrinityMetrics creates a new TrinityMetrics instance
func NewTrinityMetrics(maxSamples int) *TrinityMetrics {
	if maxSamples <= 0 {
		maxSamples = 1000 // Default: keep last 1000 samples per tier
	}

	tm := &TrinityMetrics{
		tierMetrics:    make(map[TierLevel]*TierMetrics),
		upgradeCount:   make(map[string]int64),
		downgradeCount: make(map[string]int64),
		startTime:      time.Now(),
		maxSamples:     maxSamples,
	}

	// Initialize metrics for all tiers
	for _, tier := range []TierLevel{TierThinking, TierReactive, TierRapid} {
		tm.tierMetrics[tier] = &TierMetrics{
			DurationSamples: make([]time.Duration, 0, maxSamples),
		}
	}

	logger.Debug("TrinityMetrics initialized", map[string]interface{}{
		"max_samples": maxSamples,
	})

	return tm
}

// RecordRequest records a request to a specific tier
// AC2: RecordRequest(tier, duration, error)
func (tm *TrinityMetrics) RecordRequest(tier TierLevel, duration time.Duration, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	metrics, ok := tm.tierMetrics[tier]
	if !ok {
		// This should never happen if properly initialized
		metrics = &TierMetrics{
			DurationSamples: make([]time.Duration, 0, tm.maxSamples),
		}
		tm.tierMetrics[tier] = metrics
	}

	// Update request counts
	metrics.TotalRequests++
	tm.totalRequests++

	if err != nil {
		metrics.FailedRequests++
		metrics.LastError = err
		metrics.LastErrorTime = time.Now()

		logger.Debug("Trinity request failed", map[string]interface{}{
			"tier":        tier.String(),
			"duration_ms": duration.Milliseconds(),
			"error":       err.Error(),
		})
	} else {
		metrics.SuccessRequests++

		// Update duration statistics only for successful requests
		metrics.TotalDuration += duration

		if metrics.MinDuration == 0 || duration < metrics.MinDuration {
			metrics.MinDuration = duration
		}
		if duration > metrics.MaxDuration {
			metrics.MaxDuration = duration
		}

		// Add to samples and trim if needed
		metrics.DurationSamples = append(metrics.DurationSamples, duration)
		if len(metrics.DurationSamples) > tm.maxSamples {
			metrics.DurationSamples = metrics.DurationSamples[1:]
		}

		logger.Debug("Trinity request completed", map[string]interface{}{
			"tier":        tier.String(),
			"duration_ms": duration.Milliseconds(),
		})
	}
}

// RecordUpgrade records a tier upgrade event
// AC2: RecordUpgrade(from, to)
func (tm *TrinityMetrics) RecordUpgrade(from, to TierLevel) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	key := fmt.Sprintf("%s_to_%s", from.String(), to.String())
	tm.upgradeCount[key]++

	logger.Info("Trinity tier upgrade", map[string]interface{}{
		"from": from.String(),
		"to":   to.String(),
	})
}

// RecordDowngrade records a tier downgrade event
// AC2: RecordDowngrade(from, to)
func (tm *TrinityMetrics) RecordDowngrade(from, to TierLevel) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	key := fmt.Sprintf("%s_to_%s", from.String(), to.String())
	tm.downgradeCount[key]++

	logger.Info("Trinity tier downgrade", map[string]interface{}{
		"from": from.String(),
		"to":   to.String(),
	})
}

// TierStats contains statistical information for a tier
type TierStats struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	SuccessRate     float64
	ErrorRate       float64

	// Duration statistics
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	P50Duration     time.Duration
	P90Duration     time.Duration
	P99Duration     time.Duration

	// Error information
	LastError     error
	LastErrorTime time.Time
}

// GetTierStats returns statistics for a specific tier
func (tm *TrinityMetrics) GetTierStats(tier TierLevel) TierStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	metrics, ok := tm.tierMetrics[tier]
	if !ok {
		return TierStats{}
	}

	stats := TierStats{
		TotalRequests:   metrics.TotalRequests,
		SuccessRequests: metrics.SuccessRequests,
		FailedRequests:  metrics.FailedRequests,
		MinDuration:     metrics.MinDuration,
		MaxDuration:     metrics.MaxDuration,
		LastError:       metrics.LastError,
		LastErrorTime:   metrics.LastErrorTime,
	}

	// Calculate rates
	if metrics.TotalRequests > 0 {
		stats.SuccessRate = float64(metrics.SuccessRequests) / float64(metrics.TotalRequests)
		stats.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// Calculate average duration
	if metrics.SuccessRequests > 0 {
		stats.AverageDuration = metrics.TotalDuration / time.Duration(metrics.SuccessRequests)
	}

	// Calculate percentiles from samples
	if len(metrics.DurationSamples) > 0 {
		percentiles := calculatePercentiles(metrics.DurationSamples)
		stats.P50Duration = percentiles.P50
		stats.P90Duration = percentiles.P90
		stats.P99Duration = percentiles.P99
	}

	return stats
}

// PercentileStats represents percentile statistics
type PercentileStats struct {
	P50 time.Duration
	P90 time.Duration
	P99 time.Duration
}

// calculatePercentiles calculates percentile statistics from duration samples
func calculatePercentiles(samples []time.Duration) PercentileStats {
	if len(samples) == 0 {
		return PercentileStats{}
	}

	// Create a copy and sort
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate percentile indices
	p50Idx := len(sorted) * 50 / 100
	p90Idx := len(sorted) * 90 / 100
	p99Idx := len(sorted) * 99 / 100

	// Ensure indices are within bounds
	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	if p90Idx >= len(sorted) {
		p90Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return PercentileStats{
		P50: sorted[p50Idx],
		P90: sorted[p90Idx],
		P99: sorted[p99Idx],
	}
}

// MetricsSummary contains overall metrics summary
// AC2: GetMetrics() returns statistical data
type MetricsSummary struct {
	// System-wide metrics
	Uptime        time.Duration
	TotalRequests int64

	// Per-tier statistics
	ThinkingStats TierStats
	ReactiveStats TierStats
	RapidStats    TierStats

	// Tier transition counts
	TotalUpgrades   int64
	TotalDowngrades int64
	UpgradeDetails   map[string]int64
	DowngradeDetails map[string]int64
}

// GetMetrics returns a summary of all collected metrics
// AC2: GetMetrics() returns statistical data
func (tm *TrinityMetrics) GetMetrics() MetricsSummary {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	summary := MetricsSummary{
		Uptime:           time.Since(tm.startTime),
		TotalRequests:    tm.totalRequests,
		UpgradeDetails:   make(map[string]int64),
		DowngradeDetails: make(map[string]int64),
	}

	// Get per-tier stats (we need to release the lock temporarily for each GetTierStats call)
	tm.mu.RUnlock()
	summary.ThinkingStats = tm.GetTierStats(TierThinking)
	summary.ReactiveStats = tm.GetTierStats(TierReactive)
	summary.RapidStats = tm.GetTierStats(TierRapid)
	tm.mu.RLock()

	// Copy upgrade/downgrade counts
	for k, v := range tm.upgradeCount {
		summary.UpgradeDetails[k] = v
		summary.TotalUpgrades += v
	}

	for k, v := range tm.downgradeCount {
		summary.DowngradeDetails[k] = v
		summary.TotalDowngrades += v
	}

	return summary
}

// Reset resets all metrics
func (tm *TrinityMetrics) Reset() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.startTime = time.Now()
	tm.totalRequests = 0

	// Reset tier metrics
	for tier := range tm.tierMetrics {
		tm.tierMetrics[tier] = &TierMetrics{
			DurationSamples: make([]time.Duration, 0, tm.maxSamples),
		}
	}

	// Reset transition counts
	tm.upgradeCount = make(map[string]int64)
	tm.downgradeCount = make(map[string]int64)

	logger.Info("Trinity metrics reset", nil)
}

// LogSummary logs a summary of all metrics
func (tm *TrinityMetrics) LogSummary() {
	summary := tm.GetMetrics()

	logger.Info("=== Trinity Metrics Summary ===", map[string]interface{}{
		"uptime":         summary.Uptime.String(),
		"total_requests": summary.TotalRequests,
		"upgrades":       summary.TotalUpgrades,
		"downgrades":     summary.TotalDowngrades,
	})

	// Log Thinking tier stats
	tm.logTierStats("Thinking", summary.ThinkingStats)
	tm.logTierStats("Reactive", summary.ReactiveStats)
	tm.logTierStats("Rapid", summary.RapidStats)
}

// logTierStats logs statistics for a specific tier
func (tm *TrinityMetrics) logTierStats(tierName string, stats TierStats) {
	if stats.TotalRequests == 0 {
		return
	}

	logger.Info(fmt.Sprintf("Tier: %s", tierName), map[string]interface{}{
		"total_requests":   stats.TotalRequests,
		"success_requests": stats.SuccessRequests,
		"failed_requests":  stats.FailedRequests,
		"success_rate":     fmt.Sprintf("%.2f%%", stats.SuccessRate*100),
		"error_rate":       fmt.Sprintf("%.2f%%", stats.ErrorRate*100),
		"avg_duration_ms":  stats.AverageDuration.Milliseconds(),
		"min_duration_ms":  stats.MinDuration.Milliseconds(),
		"max_duration_ms":  stats.MaxDuration.Milliseconds(),
		"p50_duration_ms":  stats.P50Duration.Milliseconds(),
		"p90_duration_ms":  stats.P90Duration.Milliseconds(),
		"p99_duration_ms":  stats.P99Duration.Milliseconds(),
	})

	if stats.LastError != nil {
		logger.Warn(fmt.Sprintf("Last error for %s tier", tierName), map[string]interface{}{
			"error": stats.LastError.Error(),
			"time":  stats.LastErrorTime.Format(time.RFC3339),
		})
	}
}

// Global metrics instance
var globalMetrics *TrinityMetrics
var metricsOnce sync.Once

// GetGlobalMetrics returns the global Trinity metrics instance
func GetGlobalMetrics() *TrinityMetrics {
	metricsOnce.Do(func() {
		globalMetrics = NewTrinityMetrics(1000)
	})
	return globalMetrics
}
