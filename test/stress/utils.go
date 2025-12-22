package stress

import (
	"fmt"
	"runtime"
	"time"
)

// ===========================================================================
// Metrics Collection
// ===========================================================================

// MetricsCollector tracks performance metrics during stress tests.
type MetricsCollector struct {
	StartTime      time.Time
	EndTime        time.Time
	MemoryBaseline uint64
	MemoryPeak     uint64
	MemorySamples  []MemorySample
	GoroutineCount []GoroutineSample
	CustomMetrics  map[string][]float64
}

// MemorySample represents a memory measurement at a specific time.
type MemorySample struct {
	Timestamp time.Time
	HeapAlloc uint64
	HeapSys   uint64
	NumGC     uint32
}

// GoroutineSample represents a goroutine count measurement.
type GoroutineSample struct {
	Timestamp time.Time
	Count     int
}

// NewMetricsCollector creates a new metrics collector and starts tracking.
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		StartTime:     time.Now(),
		CustomMetrics: make(map[string][]float64),
	}

	// Capture baseline memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mc.MemoryBaseline = m.HeapAlloc
	mc.MemoryPeak = m.HeapAlloc

	return mc
}

// SampleMemory captures current memory usage.
func (mc *MetricsCollector) SampleMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sample := MemorySample{
		Timestamp: time.Now(),
		HeapAlloc: m.HeapAlloc,
		HeapSys:   m.HeapSys,
		NumGC:     m.NumGC,
	}

	mc.MemorySamples = append(mc.MemorySamples, sample)

	// Update peak if necessary
	if m.HeapAlloc > mc.MemoryPeak {
		mc.MemoryPeak = m.HeapAlloc
	}
}

// SampleGoroutines captures current goroutine count.
func (mc *MetricsCollector) SampleGoroutines() {
	sample := GoroutineSample{
		Timestamp: time.Now(),
		Count:     runtime.NumGoroutine(),
	}
	mc.GoroutineCount = append(mc.GoroutineCount, sample)
}

// RecordMetric records a custom metric value.
func (mc *MetricsCollector) RecordMetric(name string, value float64) {
	mc.CustomMetrics[name] = append(mc.CustomMetrics[name], value)
}

// Stop finalizes metric collection.
func (mc *MetricsCollector) Stop() {
	mc.EndTime = time.Now()
	mc.SampleMemory()
	mc.SampleGoroutines()
}

// Report generates a summary report of collected metrics.
func (mc *MetricsCollector) Report() string {
	duration := mc.EndTime.Sub(mc.StartTime)
	memoryGrowth := float64(mc.MemoryPeak-mc.MemoryBaseline) / float64(mc.MemoryBaseline) * 100

	report := fmt.Sprintf(`
=== Stress Test Metrics Report ===
Duration: %v
Memory Baseline: %s
Memory Peak: %s
Memory Growth: %.2f%%
Memory Samples: %d
Goroutine Samples: %d
`,
		duration,
		formatBytes(mc.MemoryBaseline),
		formatBytes(mc.MemoryPeak),
		memoryGrowth,
		len(mc.MemorySamples),
		len(mc.GoroutineCount),
	)

	// Add custom metrics
	if len(mc.CustomMetrics) > 0 {
		report += "\nCustom Metrics:\n"
		for name, values := range mc.CustomMetrics {
			if len(values) == 0 {
				continue
			}
			avg := calculateAverage(values)
			min := calculateMin(values)
			max := calculateMax(values)
			report += fmt.Sprintf("  %s: avg=%.2f, min=%.2f, max=%.2f, samples=%d\n",
				name, avg, min, max, len(values))
		}
	}

	return report
}

// DetectMemoryLeak checks if there's a significant memory leak.
// Returns true if memory growth exceeds the threshold (default 5%).
func (mc *MetricsCollector) DetectMemoryLeak(thresholdPercent float64) bool {
	if mc.MemoryBaseline == 0 {
		return false
	}
	growth := float64(mc.MemoryPeak-mc.MemoryBaseline) / float64(mc.MemoryBaseline) * 100
	return growth > thresholdPercent
}

// DetectGoroutineLeak checks if goroutine count is continuously growing.
// Returns true if final goroutine count is significantly higher than initial.
func (mc *MetricsCollector) DetectGoroutineLeak(thresholdIncrease int) bool {
	if len(mc.GoroutineCount) < 2 {
		return false
	}
	initial := mc.GoroutineCount[0].Count
	final := mc.GoroutineCount[len(mc.GoroutineCount)-1].Count
	return (final - initial) > thresholdIncrease
}

// ===========================================================================
// Helper Functions
// ===========================================================================

// formatBytes formats byte count as human-readable string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// calculateAverage calculates the average of a slice of float64 values.
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateMin finds the minimum value in a slice.
func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

// calculateMax finds the maximum value in a slice.
func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

// calculatePercentile calculates the given percentile (0-100) from values.
func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values (simple bubble sort for small datasets)
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile index
	index := int(float64(len(sorted)-1) * percentile / 100.0)
	return sorted[index]
}

// ===========================================================================
// Test Utilities
// ===========================================================================

// RunWithTimeout runs a function with a timeout. Returns error if timeout exceeded.
func RunWithTimeout(fn func() error, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// PeriodicSampler runs a sampling function at regular intervals until stopped.
type PeriodicSampler struct {
	interval time.Duration
	fn       func()
	stopChan chan struct{}
	stopped  bool
}

// NewPeriodicSampler creates a new periodic sampler.
func NewPeriodicSampler(interval time.Duration, fn func()) *PeriodicSampler {
	return &PeriodicSampler{
		interval: interval,
		fn:       fn,
		stopChan: make(chan struct{}),
	}
}

// Start begins periodic sampling in a background goroutine.
func (ps *PeriodicSampler) Start() {
	go func() {
		ticker := time.NewTicker(ps.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ps.fn()
			case <-ps.stopChan:
				return
			}
		}
	}()
}

// Stop halts periodic sampling.
func (ps *PeriodicSampler) Stop() {
	if !ps.stopped {
		close(ps.stopChan)
		ps.stopped = true
	}
}
