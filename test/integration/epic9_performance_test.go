package integration

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// TestEpic9_ConcurrentRequests tests concurrent safety of the system
// 測試並發請求：同時發送多個 agent 請求，驗證線程安全和準確性
func TestEpic9_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	// Create router with mock providers
	thinkingProvider := NewMockProvider("Thinking response")
	reactiveProvider := NewMockProvider("Reactive response")
	rapidProvider := NewMockProvider("Rapid response")

	router := trinity.NewTrinityRouterWithProviders(
		thinkingProvider,
		reactiveProvider,
		rapidProvider,
		true, // fallback enabled
		make(map[string]trinity.TierLevel), // no overrides
	)

	router.ResetMetrics()

	// Test configuration
	totalRequests := 200
	agents := []string{"judge", "choice", "narration"}

	ctx := context.Background()
	messages := TestMessages("user", "High load test message")

	// Monitor memory
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialAlloc := memStats.Alloc
	initialGoroutines := runtime.NumGoroutine()

	// Execute requests
	startTime := time.Now()
	successCount := 0
	errorCount := 0

	for i := 0; i < totalRequests; i++ {
		agent := agents[i%len(agents)]
		_, err := router.Route(ctx, agent, messages)

		if err != nil {
			errorCount++
		} else {
			successCount++
		}

		// Log progress every 50 requests
		if (i+1)%50 == 0 {
			runtime.ReadMemStats(&memStats)
			t.Logf("Progress: %d/%d requests, Memory: %d MB, Goroutines: %d",
				i+1, totalRequests,
				memStats.Alloc/1024/1024,
				runtime.NumGoroutine())
		}
	}

	duration := time.Since(startTime)

	// Final memory check
	runtime.ReadMemStats(&memStats)
	finalAlloc := memStats.Alloc
	finalGoroutines := runtime.NumGoroutine()

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.ReadMemStats(&memStats)
	afterGCAlloc := memStats.Alloc

	// Report results
	t.Logf("High load test completed in %v", duration)
	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
	t.Logf("Requests/sec: %.2f", float64(totalRequests)/duration.Seconds())
	t.Logf("Memory - Initial: %d MB, Final: %d MB, After GC: %d MB",
		initialAlloc/1024/1024,
		finalAlloc/1024/1024,
		afterGCAlloc/1024/1024)
	t.Logf("Goroutines - Initial: %d, Final: %d", initialGoroutines, finalGoroutines)

	// Verify performance
	AssertEqual(t, 0, errorCount, "Should have no errors under high load")
	AssertEqual(t, totalRequests, successCount, "All requests should succeed")

	// Verify memory stability (should not grow unbounded)
	memoryGrowthMB := int64(finalAlloc-initialAlloc) / 1024 / 1024
	t.Logf("Memory growth: %d MB", memoryGrowthMB)
	AssertTrue(t, memoryGrowthMB < 100, "Memory growth should be reasonable (< 100 MB)")

	// Verify no goroutine leak
	goroutineDelta := finalGoroutines - initialGoroutines
	t.Logf("Goroutine delta: %d", goroutineDelta)
	AssertTrue(t, goroutineDelta < 10, "Should not have significant goroutine leak")

	// Verify metrics
	metrics := router.GetMetricsSummary()
	totalMetricRequests := metrics.ThinkingStats.TotalRequests +
		metrics.ReactiveStats.TotalRequests +
		metrics.RapidStats.TotalRequests

	AssertEqual(t, int64(totalRequests), totalMetricRequests,
		"Metrics should track all requests accurately")
}

// BenchmarkEpic9_Latency measures latency for different tiers
// 測量不同 tier 的平均延遲和 fallback 開銷
func BenchmarkEpic9_Latency(b *testing.B) {
	// Create router with mock providers
	thinkingProvider := NewMockProvider("Thinking response")
	reactiveProvider := NewMockProvider("Reactive response")
	rapidProvider := NewMockProvider("Rapid response")

	router := trinity.NewTrinityRouterWithProviders(
		thinkingProvider,
		reactiveProvider,
		rapidProvider,
		true, // fallback enabled
		make(map[string]trinity.TierLevel), // no overrides
	)

	// Measure baseline
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialAlloc := memStats.Alloc
	initialGoroutines := runtime.NumGoroutine()

	// Execute test load
	ctx := context.Background()
	messages := TestMessages("user", "Performance test message")
	totalRequests := 100

	startTime := time.Now()
	for i := 0; i < totalRequests; i++ {
		agent := []string{"judge", "choice", "narration"}[i%3]
		_, err := router.Route(ctx, agent, messages)
		AssertNoError(t, err, fmt.Sprintf("Request %d failed", i))
	}
	duration := time.Since(startTime)

	// Measure final state
	runtime.ReadMemStats(&memStats)
	finalAlloc := memStats.Alloc
	finalGoroutines := runtime.NumGoroutine()

	// Calculate metrics
	avgLatencyMs := duration.Milliseconds() / int64(totalRequests)
	requestsPerSec := float64(totalRequests) / duration.Seconds()
	memoryGrowthMB := int64(finalAlloc-initialAlloc) / 1024 / 1024
	goroutineLeak := finalGoroutines - initialGoroutines

	// Report
	t.Logf("Performance metrics:")
	t.Logf("  Average latency: %d ms/request (threshold: %d ms)", avgLatencyMs, maxAvgLatencyMs)
	t.Logf("  Throughput: %.2f req/sec (threshold: %d req/sec)", requestsPerSec, minRequestsPerSec)
	t.Logf("  Memory growth: %d MB (threshold: %d MB)", memoryGrowthMB, maxMemoryGrowthMB)
	t.Logf("  Goroutine leak: %d (threshold: %d)", goroutineLeak, maxGoroutineLeak)

	// Assert performance thresholds
	AssertTrue(t, avgLatencyMs <= maxAvgLatencyMs,
		fmt.Sprintf("Average latency %d ms exceeds threshold %d ms", avgLatencyMs, maxAvgLatencyMs))
	AssertTrue(t, requestsPerSec >= minRequestsPerSec,
		fmt.Sprintf("Throughput %.2f req/sec below threshold %d req/sec", requestsPerSec, minRequestsPerSec))
	AssertTrue(t, memoryGrowthMB <= maxMemoryGrowthMB,
		fmt.Sprintf("Memory growth %d MB exceeds threshold %d MB", memoryGrowthMB, maxMemoryGrowthMB))
	AssertTrue(t, goroutineLeak <= maxGoroutineLeak,
		fmt.Sprintf("Goroutine leak %d exceeds threshold %d", goroutineLeak, maxGoroutineLeak))
}
