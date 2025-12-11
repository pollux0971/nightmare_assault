package engine

import (
	"testing"
)

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int // Minimum expected token count
		maxCount int // Maximum expected token count
	}{
		{
			name:     "Empty string",
			text:     "",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "Simple sentence",
			text:     "Hello, world!",
			minCount: 2,
			maxCount: 5,
		},
		{
			name:     "Chinese text",
			text:     "你好，世界！這是一段中文測試。",
			minCount: 10,
			maxCount: 30,
		},
		{
			name:     "Long text",
			text:     "This is a longer piece of text that should have more tokens. It contains multiple sentences and various punctuation marks.",
			minCount: 20,
			maxCount: 40,
		},
		{
			name:     "Mixed language",
			text:     "Hello 你好 world 世界",
			minCount: 4,
			maxCount: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountTokens(tt.text)
			if count < tt.minCount || count > tt.maxCount {
				t.Errorf("CountTokens(%q) = %d, want between %d and %d",
					tt.text, count, tt.minCount, tt.maxCount)
			}
		})
	}
}

func TestCountTokensDeterministic(t *testing.T) {
	text := "This is a test sentence for deterministic token counting."

	count1 := CountTokens(text)
	count2 := CountTokens(text)

	if count1 != count2 {
		t.Errorf("CountTokens should be deterministic: got %d and %d", count1, count2)
	}
}

func TestTokenMonitor(t *testing.T) {
	monitor := NewTokenMonitor(1000) // 1000 token limit

	if monitor.Limit != 1000 {
		t.Errorf("Expected limit 1000, got %d", monitor.Limit)
	}

	if monitor.CurrentUsage != 0 {
		t.Errorf("Expected initial usage 0, got %d", monitor.CurrentUsage)
	}
}

func TestTokenMonitorAdd(t *testing.T) {
	monitor := NewTokenMonitor(1000)

	text := "This is some text to add."
	tokens := monitor.Add(text)

	if tokens <= 0 {
		t.Error("Add should return positive token count")
	}

	if monitor.CurrentUsage != tokens {
		t.Errorf("CurrentUsage should be %d, got %d", tokens, monitor.CurrentUsage)
	}
}

func TestTokenMonitorReset(t *testing.T) {
	monitor := NewTokenMonitor(1000)
	monitor.Add("Some text")

	if monitor.CurrentUsage == 0 {
		t.Fatal("Usage should not be 0 after Add")
	}

	monitor.Reset()

	if monitor.CurrentUsage != 0 {
		t.Errorf("Usage should be 0 after Reset, got %d", monitor.CurrentUsage)
	}
}

func TestTokenMonitorPercentage(t *testing.T) {
	monitor := NewTokenMonitor(1000)
	monitor.CurrentUsage = 800

	pct := monitor.Percentage()
	expected := 80.0

	if pct != expected {
		t.Errorf("Expected percentage %.1f, got %.1f", expected, pct)
	}
}

func TestTokenMonitorShouldCompress(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		usage     int
		threshold float64
		expected  bool
	}{
		{"Under threshold", 1000, 700, 0.8, false},
		{"At threshold", 1000, 800, 0.8, true},
		{"Over threshold", 1000, 850, 0.8, true},
		{"Different threshold", 1000, 650, 0.7, false},
		{"At different threshold", 1000, 700, 0.7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewTokenMonitor(tt.limit)
			monitor.CurrentUsage = tt.usage

			result := monitor.ShouldCompress(tt.threshold)
			if result != tt.expected {
				t.Errorf("ShouldCompress() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTokenMonitorGetUsageInfo(t *testing.T) {
	monitor := NewTokenMonitor(1000)
	monitor.CurrentUsage = 750

	info := monitor.GetUsageInfo()

	if info.Current != 750 {
		t.Errorf("Expected current 750, got %d", info.Current)
	}

	if info.Limit != 1000 {
		t.Errorf("Expected limit 1000, got %d", info.Limit)
	}

	if info.Percentage != 75.0 {
		t.Errorf("Expected percentage 75.0, got %.1f", info.Percentage)
	}

	if info.Remaining != 250 {
		t.Errorf("Expected remaining 250, got %d", info.Remaining)
	}
}

func TestEstimateTokensSimple(t *testing.T) {
	text := "Hello world"
	estimated := EstimateTokens(text)

	if estimated <= 0 {
		t.Error("EstimateTokens should return positive value")
	}

	// Simple heuristic: roughly 1 token per 4 characters for English
	// This is very approximate
	if estimated < 1 || estimated > 10 {
		t.Logf("Warning: Estimate seems off: %d tokens for %q", estimated, text)
	}
}

func TestEstimateTokensChinese(t *testing.T) {
	text := "你好世界"
	estimated := EstimateTokens(text)

	if estimated <= 0 {
		t.Error("EstimateTokens should return positive value")
	}
}

func BenchmarkCountTokens(b *testing.B) {
	text := "This is a benchmark test for token counting. It should be reasonably fast."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountTokens(text)
	}
}

func BenchmarkEstimateTokens(b *testing.B) {
	text := "This is a benchmark test for token estimation. It should be very fast."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EstimateTokens(text)
	}
}
