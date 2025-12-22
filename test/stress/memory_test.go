package stress

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestMemoryStability tests memory usage stability over extended operation.
// Story 8.8 AC2: Memory usage remains stable (no linear growth trend)
func TestMemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		duration       = 2 * time.Minute // Shortened from 4 hours for testing
		sampleInterval = 5 * time.Second
	)

	t.Logf("Testing memory stability over %v", duration)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Force GC to establish baseline
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	metrics.SampleMemory()

	// Start periodic sampling
	sampler := NewPeriodicSampler(sampleInterval, func() {
		metrics.SampleMemory()
		metrics.SampleGoroutines()
	})
	sampler.Start()
	defer sampler.Stop()

	// Create NPC manager and simulate activity
	npcMgr := manager.NewNPCManager(nil, nil)
	npcs := createTestNPCs(t, npcMgr, 10)

	// Run continuous operations until duration expires
	deadline := time.Now().Add(duration)
	operationCount := 0

	for time.Now().Before(deadline) {
		// Simulate various operations
		for _, npcID := range npcs {
			// Build prompt
			_ = npcMgr.BuildNPCPrompt(npcID)

			// Adjust emotion
			delta := manager.EmotionDelta{Trust: 1, Fear: -1, Stress: 0}
			_ = npcMgr.AdjustEmotion(npcID, delta)

			// Record interaction
			_ = npcMgr.RecordInteraction(npcID, manager.NPCInteraction{
				InteractionType: "memory_test",
				Description:     fmt.Sprintf("Operation %d", operationCount),
			})

			operationCount++
		}

		// Periodically force GC to detect real memory growth
		if operationCount%1000 == 0 {
			runtime.GC()
			t.Logf("Progress: %d operations completed", operationCount)
		}

		// Small sleep to prevent CPU thrashing
		time.Sleep(10 * time.Millisecond)
	}

	// Final GC and measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	metrics.SampleMemory()

	t.Logf("Completed %d operations over %v", operationCount, duration)

	// Analyze memory trend
	if len(metrics.MemorySamples) < 3 {
		t.Fatalf("Insufficient memory samples: %d", len(metrics.MemorySamples))
	}

	// Check for linear growth (simple heuristic: compare first third vs last third)
	thirdMark := len(metrics.MemorySamples) / 3
	var earlyAvg, lateAvg uint64

	for i := 0; i < thirdMark; i++ {
		earlyAvg += metrics.MemorySamples[i].HeapAlloc
	}
	earlyAvg /= uint64(thirdMark)

	for i := 2 * thirdMark; i < len(metrics.MemorySamples); i++ {
		lateAvg += metrics.MemorySamples[i].HeapAlloc
	}
	lateAvg /= uint64(len(metrics.MemorySamples) - 2*thirdMark)

	growthPercent := float64(lateAvg-earlyAvg) / float64(earlyAvg) * 100

	t.Logf("Memory growth: early avg=%s, late avg=%s, growth=%.2f%%",
		formatBytes(earlyAvg), formatBytes(lateAvg), growthPercent)

	// Verify growth is within acceptable bounds (10% for this shorter test)
	if growthPercent > 10.0 {
		t.Errorf("Memory growth exceeds threshold: %.2f%% > 10%%", growthPercent)
	}

	// Check for goroutine leaks
	if metrics.DetectGoroutineLeak(5) {
		t.Errorf("Goroutine leak detected")
	}
}

// TestGoroutineStability tests goroutine count stability.
// Story 8.8 AC2: No goroutine leaks over extended operation
func TestGoroutineStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		iterations = 1000
	)

	t.Logf("Testing goroutine stability over %d iterations", iterations)

	metrics := NewMetricsCollector()
	defer func() {
		metrics.Stop()
		t.Log(metrics.Report())
	}()

	// Establish baseline
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	metrics.SampleGoroutines()
	baselineCount := runtime.NumGoroutine()

	// Create NPC manager
	npcMgr := manager.NewNPCManager(nil, nil)
	npcs := createTestNPCs(t, npcMgr, 5)

	// Run iterations with goroutine-creating operations
	for i := 0; i < iterations; i++ {
		// Simulate operations that might spawn goroutines
		for _, npcID := range npcs {
			_ = npcMgr.BuildNPCPrompt(npcID)
			_ = npcMgr.GetState(npcID)
		}

		// Sample periodically
		if i%100 == 0 {
			metrics.SampleGoroutines()
			currentCount := runtime.NumGoroutine()
			t.Logf("Iteration %d: goroutines=%d (baseline=%d)", i, currentCount, baselineCount)
		}

		// Brief pause to allow goroutines to complete
		time.Sleep(5 * time.Millisecond)
	}

	// Final measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	metrics.SampleGoroutines()
	finalCount := runtime.NumGoroutine()

	increase := finalCount - baselineCount
	t.Logf("Goroutine count: baseline=%d, final=%d, increase=%d", baselineCount, finalCount, increase)

	// Verify goroutine count is stable (allow small increase for background tasks)
	if increase > 10 {
		t.Errorf("Too many goroutines leaked: increase=%d > threshold=10", increase)
	}
}

// TestMemoryProfilerUtility tests the metrics collector functionality.
// Story 8.8 AC2: Memory profiling and leak detection utilities work correctly
func TestMemoryProfilerUtility(t *testing.T) {
	metrics := NewMetricsCollector()

	// Take baseline sample
	metrics.SampleMemory()
	baseline := metrics.MemoryBaseline

	// Allocate some memory
	data := make([]byte, 10*1024*1024) // 10 MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Take another sample
	time.Sleep(100 * time.Millisecond)
	metrics.SampleMemory()

	// Verify samples were recorded
	if len(metrics.MemorySamples) < 2 {
		t.Errorf("Expected at least 2 memory samples, got %d", len(metrics.MemorySamples))
	}

	// Verify peak is tracked
	if metrics.MemoryPeak <= baseline {
		t.Errorf("Peak memory should be higher than baseline: peak=%d, baseline=%d",
			metrics.MemoryPeak, baseline)
	}

	// Test leak detection
	metrics.Stop()
	if !metrics.DetectMemoryLeak(5.0) {
		t.Log("No leak detected (expected for this test)")
	}

	// Test report generation
	report := metrics.Report()
	if len(report) == 0 {
		t.Error("Report should not be empty")
	}
	t.Logf("Report:\n%s", report)

	// Keep reference to data to prevent early GC
	_ = data
}

// TestCustomMetricsCollection tests custom metrics recording.
// Story 8.8 AC7: Metrics collection framework is extensible
func TestCustomMetricsCollection(t *testing.T) {
	metrics := NewMetricsCollector()

	// Record various custom metrics
	for i := 0; i < 100; i++ {
		metrics.RecordMetric("response_time_ms", float64(i*10))
		metrics.RecordMetric("token_count", float64(i*50))
	}

	metrics.Stop()

	// Verify metrics were recorded
	if len(metrics.CustomMetrics["response_time_ms"]) != 100 {
		t.Errorf("Expected 100 response_time samples, got %d",
			len(metrics.CustomMetrics["response_time_ms"]))
	}

	// Test metric calculations
	responseTimes := metrics.CustomMetrics["response_time_ms"]
	avg := calculateAverage(responseTimes)
	min := calculateMin(responseTimes)
	max := calculateMax(responseTimes)
	p50 := calculatePercentile(responseTimes, 50)
	p90 := calculatePercentile(responseTimes, 90)
	p99 := calculatePercentile(responseTimes, 99)

	t.Logf("Response times: avg=%.2f, min=%.2f, max=%.2f", avg, min, max)
	t.Logf("Percentiles: p50=%.2f, p90=%.2f, p99=%.2f", p50, p90, p99)

	// Verify calculations are reasonable
	if min > avg || avg > max {
		t.Errorf("Invalid metric calculations: min=%.2f, avg=%.2f, max=%.2f", min, avg, max)
	}
	if p50 > p90 || p90 > p99 {
		t.Errorf("Invalid percentiles: p50=%.2f, p90=%.2f, p99=%.2f", p50, p90, p99)
	}

	// Test report includes custom metrics
	report := metrics.Report()
	if len(report) == 0 {
		t.Error("Report should not be empty")
	}
	t.Logf("Full report:\n%s", report)
}
