package trinity

import (
	"fmt"
	"time"
)

// Example_metricsCollection demonstrates how to use Trinity metrics collection
func Example_metricsCollection() {
	// Create a router with metrics enabled
	cfg := DefaultRouterConfig()
	cfg.ThinkingProvider.APIKey = "test-key"
	cfg.ReactiveProvider.APIKey = "test-key"
	cfg.RapidProvider.APIKey = "test-key"

	// Note: In real usage, this would create actual providers
	// For this example, we'll demonstrate the metrics API

	// Create metrics instance
	metrics := NewTrinityMetrics(1000)

	// Simulate some requests
	metrics.RecordRequest(TierThinking, 150*time.Millisecond, nil)
	metrics.RecordRequest(TierReactive, 75*time.Millisecond, nil)
	metrics.RecordRequest(TierRapid, 25*time.Millisecond, nil)

	// Simulate a tier downgrade
	metrics.RecordDowngrade(TierThinking, TierReactive)

	// Get statistics for a specific tier
	stats := metrics.GetTierStats(TierThinking)
	fmt.Printf("Thinking Tier - Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Thinking Tier - Success Rate: %.0f%%\n", stats.SuccessRate*100)

	// Get overall metrics summary
	summary := metrics.GetMetrics()
	fmt.Printf("Total Requests: %d\n", summary.TotalRequests)
	fmt.Printf("Total Downgrades: %d\n", summary.TotalDowngrades)

	// Output:
	// Thinking Tier - Total Requests: 1
	// Thinking Tier - Success Rate: 100%
	// Total Requests: 3
	// Total Downgrades: 1
}

// Example_routerMetrics demonstrates using metrics through TrinityRouter
func Example_routerMetrics() {
	// Note: This is a conceptual example showing the API
	// In real usage, you would have actual API keys and providers configured

	// After routing some requests...
	// router.Route(ctx, "NPCAgent", messages)
	// router.Route(ctx, "NarrationAgent", messages)

	// Get metrics summary from router
	// summary := router.GetMetricsSummary()
	// fmt.Printf("Total Requests: %d\n", summary.TotalRequests)

	// Get specific tier statistics
	// metrics := router.GetMetrics()
	// thinkingStats := metrics.GetTierStats(TierThinking)
	// fmt.Printf("Thinking tier average latency: %v\n", thinkingStats.AverageDuration)

	// Log metrics summary
	// router.LogMetricsSummary()

	// Reset metrics if needed
	// router.ResetMetrics()

	fmt.Println("See router integration tests for full working examples")

	// Output:
	// See router integration tests for full working examples
}

// Example_globalMetrics demonstrates using the global metrics instance
func Example_globalMetrics() {
	// Get the global metrics instance
	metrics := GetGlobalMetrics()

	// Record some requests
	metrics.RecordRequest(TierReactive, 100*time.Millisecond, nil)

	// Get tier stats
	stats := metrics.GetTierStats(TierReactive)
	fmt.Printf("Has requests: %v\n", stats.TotalRequests > 0)

	// Reset for clean state
	metrics.Reset()

	// Output:
	// Has requests: true
}

// Example_metricsPercentiles demonstrates percentile calculations
func Example_metricsPercentiles() {
	metrics := NewTrinityMetrics(100)

	// Simulate multiple requests with varying latencies
	for i := 1; i <= 10; i++ {
		duration := time.Duration(i*100) * time.Millisecond
		metrics.RecordRequest(TierReactive, duration, nil)
	}

	// Get tier statistics with percentiles
	stats := metrics.GetTierStats(TierReactive)

	fmt.Printf("Min Duration: %v\n", stats.MinDuration)
	fmt.Printf("Max Duration: %v\n", stats.MaxDuration)
	fmt.Printf("Average Duration: %v\n", stats.AverageDuration)
	fmt.Printf("P50 (Median): %v\n", stats.P50Duration)
	fmt.Printf("P90: %v\n", stats.P90Duration)
	fmt.Printf("P99: %v\n", stats.P99Duration)

	// Output:
	// Min Duration: 100ms
	// Max Duration: 1s
	// Average Duration: 550ms
	// P50 (Median): 600ms
	// P90: 1s
	// P99: 1s
}
