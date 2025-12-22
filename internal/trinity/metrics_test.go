package trinity

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// Story 9-5: Trinity Metrics Collection Tests
// AC4: Test coverage >80%

func TestNewTrinityMetrics(t *testing.T) {
	tests := []struct {
		name       string
		maxSamples int
		wantSamples int
	}{
		{
			name:       "default max samples",
			maxSamples: 0,
			wantSamples: 1000,
		},
		{
			name:       "custom max samples",
			maxSamples: 500,
			wantSamples: 500,
		},
		{
			name:       "negative max samples uses default",
			maxSamples: -100,
			wantSamples: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTrinityMetrics(tt.maxSamples)

			if tm == nil {
				t.Fatal("NewTrinityMetrics returned nil")
			}

			if tm.maxSamples != tt.wantSamples {
				t.Errorf("maxSamples = %d, want %d", tm.maxSamples, tt.wantSamples)
			}

			// Verify all tiers are initialized
			for _, tier := range []TierLevel{TierThinking, TierReactive, TierRapid} {
				if _, ok := tm.tierMetrics[tier]; !ok {
					t.Errorf("tier %s not initialized", tier.String())
				}
			}

			if tm.upgradeCount == nil {
				t.Error("upgradeCount map not initialized")
			}

			if tm.downgradeCount == nil {
				t.Error("downgradeCount map not initialized")
			}

			if tm.startTime.IsZero() {
				t.Error("startTime not set")
			}
		})
	}
}

func TestRecordRequest_Success(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record a successful request
	duration := 250 * time.Millisecond
	tm.RecordRequest(TierThinking, duration, nil)

	stats := tm.GetTierStats(TierThinking)

	if stats.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", stats.TotalRequests)
	}

	if stats.SuccessRequests != 1 {
		t.Errorf("SuccessRequests = %d, want 1", stats.SuccessRequests)
	}

	if stats.FailedRequests != 0 {
		t.Errorf("FailedRequests = %d, want 0", stats.FailedRequests)
	}

	if stats.MinDuration != duration {
		t.Errorf("MinDuration = %v, want %v", stats.MinDuration, duration)
	}

	if stats.MaxDuration != duration {
		t.Errorf("MaxDuration = %v, want %v", stats.MaxDuration, duration)
	}

	if stats.AverageDuration != duration {
		t.Errorf("AverageDuration = %v, want %v", stats.AverageDuration, duration)
	}

	if stats.SuccessRate != 1.0 {
		t.Errorf("SuccessRate = %f, want 1.0", stats.SuccessRate)
	}
}

func TestRecordRequest_Failure(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record a failed request
	duration := 100 * time.Millisecond
	testErr := errors.New("test error")
	tm.RecordRequest(TierReactive, duration, testErr)

	stats := tm.GetTierStats(TierReactive)

	if stats.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", stats.TotalRequests)
	}

	if stats.SuccessRequests != 0 {
		t.Errorf("SuccessRequests = %d, want 0", stats.SuccessRequests)
	}

	if stats.FailedRequests != 1 {
		t.Errorf("FailedRequests = %d, want 1", stats.FailedRequests)
	}

	if stats.ErrorRate != 1.0 {
		t.Errorf("ErrorRate = %f, want 1.0", stats.ErrorRate)
	}

	if stats.LastError == nil || stats.LastError.Error() != "test error" {
		t.Errorf("LastError = %v, want 'test error'", stats.LastError)
	}

	if stats.LastErrorTime.IsZero() {
		t.Error("LastErrorTime not set")
	}
}

func TestRecordRequest_MultipleRequests(t *testing.T) {
	tm := NewTrinityMetrics(100)

	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		150 * time.Millisecond,
		250 * time.Millisecond,
	}

	for _, d := range durations {
		tm.RecordRequest(TierRapid, d, nil)
	}

	stats := tm.GetTierStats(TierRapid)

	if stats.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", stats.TotalRequests)
	}

	if stats.SuccessRequests != 5 {
		t.Errorf("SuccessRequests = %d, want 5", stats.SuccessRequests)
	}

	expectedMin := 100 * time.Millisecond
	if stats.MinDuration != expectedMin {
		t.Errorf("MinDuration = %v, want %v", stats.MinDuration, expectedMin)
	}

	expectedMax := 300 * time.Millisecond
	if stats.MaxDuration != expectedMax {
		t.Errorf("MaxDuration = %v, want %v", stats.MaxDuration, expectedMax)
	}

	// Average should be 200ms
	expectedAvg := 200 * time.Millisecond
	if stats.AverageDuration != expectedAvg {
		t.Errorf("AverageDuration = %v, want %v", stats.AverageDuration, expectedAvg)
	}
}

func TestRecordRequest_MixedSuccessAndFailure(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// 3 successful, 2 failed
	tm.RecordRequest(TierThinking, 100*time.Millisecond, nil)
	tm.RecordRequest(TierThinking, 200*time.Millisecond, nil)
	tm.RecordRequest(TierThinking, 150*time.Millisecond, errors.New("error 1"))
	tm.RecordRequest(TierThinking, 300*time.Millisecond, nil)
	tm.RecordRequest(TierThinking, 250*time.Millisecond, errors.New("error 2"))

	stats := tm.GetTierStats(TierThinking)

	if stats.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", stats.TotalRequests)
	}

	if stats.SuccessRequests != 3 {
		t.Errorf("SuccessRequests = %d, want 3", stats.SuccessRequests)
	}

	if stats.FailedRequests != 2 {
		t.Errorf("FailedRequests = %d, want 2", stats.FailedRequests)
	}

	expectedSuccessRate := 0.6
	if stats.SuccessRate != expectedSuccessRate {
		t.Errorf("SuccessRate = %f, want %f", stats.SuccessRate, expectedSuccessRate)
	}

	expectedErrorRate := 0.4
	if stats.ErrorRate != expectedErrorRate {
		t.Errorf("ErrorRate = %f, want %f", stats.ErrorRate, expectedErrorRate)
	}

	// Average should be (100 + 200 + 300) / 3 = 200ms (only successful requests)
	expectedAvg := 200 * time.Millisecond
	if stats.AverageDuration != expectedAvg {
		t.Errorf("AverageDuration = %v, want %v", stats.AverageDuration, expectedAvg)
	}
}

func TestRecordRequest_SampleTrimming(t *testing.T) {
	maxSamples := 10
	tm := NewTrinityMetrics(maxSamples)

	// Record more requests than maxSamples
	for i := 0; i < 15; i++ {
		tm.RecordRequest(TierReactive, time.Duration(i+1)*time.Millisecond, nil)
	}

	tm.mu.RLock()
	samples := tm.tierMetrics[TierReactive].DurationSamples
	tm.mu.RUnlock()

	if len(samples) != maxSamples {
		t.Errorf("len(DurationSamples) = %d, want %d", len(samples), maxSamples)
	}

	// Verify oldest samples were removed (first 5 should be gone)
	// Latest samples should be 6ms through 15ms
	if samples[0] != 6*time.Millisecond {
		t.Errorf("First sample = %v, want 6ms", samples[0])
	}
}

func TestRecordUpgrade(t *testing.T) {
	tm := NewTrinityMetrics(100)

	tm.RecordUpgrade(TierRapid, TierReactive)
	tm.RecordUpgrade(TierReactive, TierThinking)
	tm.RecordUpgrade(TierRapid, TierThinking)

	summary := tm.GetMetrics()

	if summary.TotalUpgrades != 3 {
		t.Errorf("TotalUpgrades = %d, want 3", summary.TotalUpgrades)
	}

	expectedKey := "Rapid_to_Reactive"
	if count, ok := summary.UpgradeDetails[expectedKey]; !ok || count != 1 {
		t.Errorf("UpgradeDetails[%s] = %d, want 1", expectedKey, count)
	}

	expectedKey2 := "Reactive_to_Thinking"
	if count, ok := summary.UpgradeDetails[expectedKey2]; !ok || count != 1 {
		t.Errorf("UpgradeDetails[%s] = %d, want 1", expectedKey2, count)
	}

	expectedKey3 := "Rapid_to_Thinking"
	if count, ok := summary.UpgradeDetails[expectedKey3]; !ok || count != 1 {
		t.Errorf("UpgradeDetails[%s] = %d, want 1", expectedKey3, count)
	}
}

func TestRecordDowngrade(t *testing.T) {
	tm := NewTrinityMetrics(100)

	tm.RecordDowngrade(TierThinking, TierReactive)
	tm.RecordDowngrade(TierReactive, TierRapid)
	tm.RecordDowngrade(TierThinking, TierReactive)

	summary := tm.GetMetrics()

	if summary.TotalDowngrades != 3 {
		t.Errorf("TotalDowngrades = %d, want 3", summary.TotalDowngrades)
	}

	expectedKey := "Thinking_to_Reactive"
	if count, ok := summary.DowngradeDetails[expectedKey]; !ok || count != 2 {
		t.Errorf("DowngradeDetails[%s] = %d, want 2", expectedKey, count)
	}

	expectedKey2 := "Reactive_to_Rapid"
	if count, ok := summary.DowngradeDetails[expectedKey2]; !ok || count != 1 {
		t.Errorf("DowngradeDetails[%s] = %d, want 1", expectedKey2, count)
	}
}

func TestGetMetricsSummary(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record some data
	tm.RecordRequest(TierThinking, 100*time.Millisecond, nil)
	tm.RecordRequest(TierReactive, 50*time.Millisecond, nil)
	tm.RecordRequest(TierRapid, 25*time.Millisecond, nil)
	tm.RecordUpgrade(TierRapid, TierReactive)
	tm.RecordDowngrade(TierThinking, TierReactive)

	summary := tm.GetMetrics()

	if summary.TotalRequests != 3 {
		t.Errorf("TotalRequests = %d, want 3", summary.TotalRequests)
	}

	if summary.TotalUpgrades != 1 {
		t.Errorf("TotalUpgrades = %d, want 1", summary.TotalUpgrades)
	}

	if summary.TotalDowngrades != 1 {
		t.Errorf("TotalDowngrades = %d, want 1", summary.TotalDowngrades)
	}

	if summary.ThinkingStats.TotalRequests != 1 {
		t.Errorf("ThinkingStats.TotalRequests = %d, want 1", summary.ThinkingStats.TotalRequests)
	}

	if summary.ReactiveStats.TotalRequests != 1 {
		t.Errorf("ReactiveStats.TotalRequests = %d, want 1", summary.ReactiveStats.TotalRequests)
	}

	if summary.RapidStats.TotalRequests != 1 {
		t.Errorf("RapidStats.TotalRequests = %d, want 1", summary.RapidStats.TotalRequests)
	}

	if summary.Uptime <= 0 {
		t.Error("Uptime should be > 0")
	}
}

func TestReset(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record some data
	tm.RecordRequest(TierThinking, 100*time.Millisecond, nil)
	tm.RecordRequest(TierReactive, 50*time.Millisecond, errors.New("test error"))
	tm.RecordUpgrade(TierRapid, TierReactive)
	tm.RecordDowngrade(TierThinking, TierReactive)

	// Reset
	tm.Reset()

	summary := tm.GetMetrics()

	if summary.TotalRequests != 0 {
		t.Errorf("TotalRequests after reset = %d, want 0", summary.TotalRequests)
	}

	if summary.TotalUpgrades != 0 {
		t.Errorf("TotalUpgrades after reset = %d, want 0", summary.TotalUpgrades)
	}

	if summary.TotalDowngrades != 0 {
		t.Errorf("TotalDowngrades after reset = %d, want 0", summary.TotalDowngrades)
	}

	// Verify all tier stats are reset
	for _, tier := range []TierLevel{TierThinking, TierReactive, TierRapid} {
		stats := tm.GetTierStats(tier)
		if stats.TotalRequests != 0 {
			t.Errorf("Tier %s TotalRequests after reset = %d, want 0", tier.String(), stats.TotalRequests)
		}
	}
}

func TestCalculatePercentiles(t *testing.T) {
	tests := []struct {
		name    string
		samples []time.Duration
		want    PercentileStats
	}{
		{
			name:    "empty samples",
			samples: []time.Duration{},
			want:    PercentileStats{},
		},
		{
			name: "single sample",
			samples: []time.Duration{
				100 * time.Millisecond,
			},
			want: PercentileStats{
				P50: 100 * time.Millisecond,
				P90: 100 * time.Millisecond,
				P99: 100 * time.Millisecond,
			},
		},
		{
			name: "multiple samples",
			samples: []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				300 * time.Millisecond,
				400 * time.Millisecond,
				500 * time.Millisecond,
			},
			want: PercentileStats{
				P50: 300 * time.Millisecond, // 50th percentile (index 2)
				P90: 500 * time.Millisecond, // 90th percentile (index 4)
				P99: 500 * time.Millisecond, // 99th percentile (index 4)
			},
		},
		{
			name: "unsorted samples",
			samples: []time.Duration{
				500 * time.Millisecond,
				100 * time.Millisecond,
				300 * time.Millisecond,
				200 * time.Millisecond,
				400 * time.Millisecond,
			},
			want: PercentileStats{
				P50: 300 * time.Millisecond,
				P90: 500 * time.Millisecond,
				P99: 500 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePercentiles(tt.samples)

			if got.P50 != tt.want.P50 {
				t.Errorf("P50 = %v, want %v", got.P50, tt.want.P50)
			}
			if got.P90 != tt.want.P90 {
				t.Errorf("P90 = %v, want %v", got.P90, tt.want.P90)
			}
			if got.P99 != tt.want.P99 {
				t.Errorf("P99 = %v, want %v", got.P99, tt.want.P99)
			}
		})
	}
}

func TestGetTierStats_Percentiles(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record samples with known distribution
	durations := []time.Duration{
		100 * time.Millisecond,
		150 * time.Millisecond,
		200 * time.Millisecond,
		250 * time.Millisecond,
		300 * time.Millisecond,
		350 * time.Millisecond,
		400 * time.Millisecond,
		450 * time.Millisecond,
		500 * time.Millisecond,
		550 * time.Millisecond,
	}

	for _, d := range durations {
		tm.RecordRequest(TierThinking, d, nil)
	}

	stats := tm.GetTierStats(TierThinking)

	// With 10 samples:
	// P50 (50%) = index 5 = 350ms
	// P90 (90%) = index 9 = 550ms
	// P99 (99%) = index 9 = 550ms

	if stats.P50Duration != 350*time.Millisecond {
		t.Errorf("P50Duration = %v, want 350ms", stats.P50Duration)
	}

	if stats.P90Duration != 550*time.Millisecond {
		t.Errorf("P90Duration = %v, want 550ms", stats.P90Duration)
	}

	if stats.P99Duration != 550*time.Millisecond {
		t.Errorf("P99Duration = %v, want 550ms", stats.P99Duration)
	}
}

// TestConcurrentRecordRequest tests thread safety
// AC4: Concurrent safety test
func TestConcurrentRecordRequest(t *testing.T) {
	tm := NewTrinityMetrics(1000)

	var wg sync.WaitGroup
	numGoroutines := 10
	requestsPerGoroutine := 100

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			tier := TierLevel(goroutineID % 3) // Distribute across tiers
			for j := 0; j < requestsPerGoroutine; j++ {
				duration := time.Duration(j+1) * time.Millisecond
				var err error
				if j%10 == 0 {
					err = errors.New("test error")
				}
				tm.RecordRequest(tier, duration, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify total requests
	summary := tm.GetMetrics()
	expectedTotal := int64(numGoroutines * requestsPerGoroutine)
	if summary.TotalRequests != expectedTotal {
		t.Errorf("TotalRequests = %d, want %d", summary.TotalRequests, expectedTotal)
	}
}

// TestConcurrentUpgradeDowngrade tests concurrent tier transitions
func TestConcurrentUpgradeDowngrade(t *testing.T) {
	tm := NewTrinityMetrics(100)

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				if j%2 == 0 {
					tm.RecordUpgrade(TierRapid, TierReactive)
				} else {
					tm.RecordDowngrade(TierThinking, TierReactive)
				}
			}
		}()
	}

	wg.Wait()

	summary := tm.GetMetrics()

	expectedUpgrades := int64(numGoroutines * operationsPerGoroutine / 2)
	if summary.TotalUpgrades != expectedUpgrades {
		t.Errorf("TotalUpgrades = %d, want %d", summary.TotalUpgrades, expectedUpgrades)
	}

	expectedDowngrades := int64(numGoroutines * operationsPerGoroutine / 2)
	if summary.TotalDowngrades != expectedDowngrades {
		t.Errorf("TotalDowngrades = %d, want %d", summary.TotalDowngrades, expectedDowngrades)
	}
}

// TestGetGlobalMetrics tests singleton pattern
func TestGetGlobalMetrics(t *testing.T) {
	m1 := GetGlobalMetrics()
	m2 := GetGlobalMetrics()

	if m1 != m2 {
		t.Error("GetGlobalMetrics should return the same instance")
	}

	// Record a request and verify it's reflected in both references
	m1.RecordRequest(TierThinking, 100*time.Millisecond, nil)

	stats1 := m1.GetTierStats(TierThinking)
	stats2 := m2.GetTierStats(TierThinking)

	if stats1.TotalRequests != stats2.TotalRequests {
		t.Error("Global metrics instance not shared properly")
	}
}

// TestLogSummary verifies LogSummary doesn't panic
func TestLogSummary(t *testing.T) {
	tm := NewTrinityMetrics(100)

	// Record some data
	tm.RecordRequest(TierThinking, 100*time.Millisecond, nil)
	tm.RecordRequest(TierReactive, 50*time.Millisecond, errors.New("test error"))
	tm.RecordUpgrade(TierRapid, TierReactive)

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogSummary panicked: %v", r)
		}
	}()

	tm.LogSummary()
}

// TestGetTierStats_NoData verifies handling of empty tier
func TestGetTierStats_NoData(t *testing.T) {
	tm := NewTrinityMetrics(100)

	stats := tm.GetTierStats(TierThinking)

	if stats.TotalRequests != 0 {
		t.Errorf("TotalRequests = %d, want 0", stats.TotalRequests)
	}

	if stats.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0", stats.SuccessRate)
	}

	if stats.ErrorRate != 0 {
		t.Errorf("ErrorRate = %f, want 0", stats.ErrorRate)
	}

	if stats.AverageDuration != 0 {
		t.Errorf("AverageDuration = %v, want 0", stats.AverageDuration)
	}
}
