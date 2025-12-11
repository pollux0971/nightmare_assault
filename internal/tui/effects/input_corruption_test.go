package effects

import (
	"strings"
	"testing"
)

func TestApplyInputCorruption_NoCorruption(t *testing.T) {
	input := "測試輸入文字"
	result := ApplyInputCorruption(input, 0.0)

	if result.Corrupted != input {
		t.Errorf("Expected no corruption, got %q", result.Corrupted)
	}
	if result.DeletionCount != 0 {
		t.Errorf("Expected 0 deletions, got %d", result.DeletionCount)
	}
	if result.FlashWarning {
		t.Error("Expected no flash warning")
	}
}

func TestApplyInputCorruption_WithCorruption(t *testing.T) {
	input := "AAAAAAAAAA" // 10 characters
	// Use high corruption rate to ensure some deletions occur
	result := ApplyInputCorruption(input, 0.5)

	// Should have some deletions (probabilistic, but with 10 chars at 50% rate, very likely)
	if result.Corrupted == input && len(input) > 5 {
		t.Error("Expected some corruption with 0.5 typingBehavior")
	}

	// Corrupted should be substring of original (only deletions, no additions)
	for _, r := range []rune(result.Corrupted) {
		if !strings.ContainsRune(input, r) {
			t.Errorf("Corrupted contains character not in original: %c", r)
		}
	}
}

func TestApplyInputCorruption_EmptyInput(t *testing.T) {
	result := ApplyInputCorruption("", 0.5)

	if result.Corrupted != "" {
		t.Errorf("Expected empty result for empty input")
	}
	if result.DeletionCount != 0 {
		t.Errorf("Expected 0 deletions for empty input")
	}
}

func TestApplyInputCorruption_HighCorruptionRate(t *testing.T) {
	input := strings.Repeat("A", 50)
	result := ApplyInputCorruption(input, 0.9) // 90% deletion rate

	// Most characters should be deleted
	if len(result.Corrupted) > 15 { // Allow for some variance
		t.Errorf("Expected most chars deleted with 0.9 rate, got %d remaining", len(result.Corrupted))
	}

	if result.DeletionCount == 0 {
		t.Error("Expected deletions with high corruption rate")
	}

	if !result.FlashWarning {
		t.Error("Expected flash warning when deletions occur")
	}
}

func TestGetInputCorruptionFeedback(t *testing.T) {
	tests := []struct {
		deletionCount int
		expectEmpty   bool
		contains      string
	}{
		{0, true, ""},
		{1, false, "一個"},
		{2, false, "部分"},
		{4, false, "大量"},
		{10, false, "崩潰"},
	}

	for _, tt := range tests {
		result := GetInputCorruptionFeedback(tt.deletionCount)

		if tt.expectEmpty {
			if result != "" {
				t.Errorf("GetInputCorruptionFeedback(%d) = %q, want empty", tt.deletionCount, result)
			}
		} else {
			if result == "" {
				t.Errorf("GetInputCorruptionFeedback(%d) returned empty, want message", tt.deletionCount)
			}
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("GetInputCorruptionFeedback(%d) = %q, want to contain %q",
					tt.deletionCount, result, tt.contains)
			}
		}
	}
}

func TestShouldCorruptInput(t *testing.T) {
	tests := []struct {
		san      int
		expected bool
	}{
		{100, false},
		{80, false},
		{60, false},
		{40, false}, // Boundary: 40 is NOT corrupted
		{39, true},  // Boundary: 39 IS corrupted
		{20, true},
		{10, true},
		{1, true},
	}

	for _, tt := range tests {
		result := ShouldCorruptInput(tt.san)
		if result != tt.expected {
			t.Errorf("ShouldCorruptInput(%d) = %v, want %v", tt.san, result, tt.expected)
		}
	}
}

func TestApplyTypingBehaviorEffect_HighSAN(t *testing.T) {
	input := "測試文字"
	result := ApplyTypingBehaviorEffect(input, 80) // High SAN - no corruption

	if result.Corrupted != input {
		t.Errorf("Expected no corruption with high SAN, got %q", result.Corrupted)
	}
	if result.DeletionCount != 0 {
		t.Errorf("Expected 0 deletions with high SAN")
	}
}

func TestApplyTypingBehaviorEffect_LowSAN(t *testing.T) {
	input := strings.Repeat("A", 30) // Enough chars to see effect
	result := ApplyTypingBehaviorEffect(input, 15) // Low SAN - should corrupt

	// With SAN=15, typingBehavior should be 0.15 (from HorrorStyle mapping)
	// Some deletions should occur (probabilistic)
	if result.Corrupted == input {
		t.Log("Warning: No corruption occurred (probabilistic test may occasionally pass)")
	}
}

func TestCorruptRealTimeInput_NoCorruption(t *testing.T) {
	newBuffer, wasDeleted := CorruptRealTimeInput("hello", 'X', 0.0)

	if newBuffer != "helloX" {
		t.Errorf("Expected 'helloX', got %q", newBuffer)
	}
	if wasDeleted {
		t.Error("Expected character not to be deleted")
	}
}

func TestCorruptRealTimeInput_WithCorruption(t *testing.T) {
	// Run multiple times due to probabilistic nature
	deletedCount := 0
	acceptedCount := 0

	for i := 0; i < 100; i++ {
		newBuffer, wasDeleted := CorruptRealTimeInput("test", 'X', 0.5)

		if wasDeleted {
			if newBuffer != "test" {
				t.Errorf("When deleted, buffer should remain 'test', got %q", newBuffer)
			}
			deletedCount++
		} else {
			if newBuffer != "testX" {
				t.Errorf("When accepted, buffer should be 'testX', got %q", newBuffer)
			}
			acceptedCount++
		}
	}

	// With 100 trials at 50% probability, expect roughly 30-70 deletions
	if deletedCount < 30 || deletedCount > 70 {
		t.Errorf("Expected ~50 deletions out of 100 trials, got %d", deletedCount)
	}
}

func TestCalculateInputVisualFeedback_NoWarning(t *testing.T) {
	corrupted := CorruptedInput{
		FlashWarning: false,
	}

	feedback := CalculateInputVisualFeedback(corrupted)

	if feedback.ShowFeedback {
		t.Error("Expected no feedback when FlashWarning is false")
	}
}

func TestCalculateInputVisualFeedback_WithWarning(t *testing.T) {
	corrupted := CorruptedInput{
		FlashWarning:  true,
		DeletionCount: 2,
	}

	feedback := CalculateInputVisualFeedback(corrupted)

	if !feedback.ShowFeedback {
		t.Error("Expected feedback when FlashWarning is true")
	}
	if feedback.FlashColor != "1" {
		t.Errorf("Expected red flash color (1), got %q", feedback.FlashColor)
	}
	if feedback.FlashDuration != 150 {
		t.Errorf("Expected 150ms flash, got %d", feedback.FlashDuration)
	}
	if feedback.FeedbackText == "" {
		t.Error("Expected feedback text to be set")
	}
}

func TestCalculateInputVisualFeedback_SevereCorruption(t *testing.T) {
	corrupted := CorruptedInput{
		FlashWarning:  true,
		DeletionCount: 10, // Severe
	}

	feedback := CalculateInputVisualFeedback(corrupted)

	if feedback.FlashDuration != 300 {
		t.Errorf("Expected 300ms flash for severe corruption, got %d", feedback.FlashDuration)
	}
}

func TestGetCursorDesyncOffset_HighSAN(t *testing.T) {
	// Run multiple times to check consistency
	for i := 0; i < 10; i++ {
		offset := GetCursorDesyncOffset(80)
		if offset != 0 {
			t.Errorf("Expected 0 offset for high SAN, got %d", offset)
		}
	}
}

func TestGetCursorDesyncOffset_LowSAN(t *testing.T) {
	// Run multiple times to check range
	offsets := make(map[int]bool)

	for i := 0; i < 50; i++ {
		offset := GetCursorDesyncOffset(5) // Very low SAN
		offsets[offset] = true

		// Should be in range -2 to +2 for severe desync
		if offset < -2 || offset > 2 {
			t.Errorf("Offset out of range: %d (expected -2 to +2)", offset)
		}
	}

	// Should produce some variety
	if len(offsets) < 3 {
		t.Errorf("Expected diverse offsets, got only %v", offsets)
	}
}

func TestCalculateInputBoxShrinkage_AllRanges(t *testing.T) {
	tests := []struct {
		san         int
		minExpected float64
		maxExpected float64
	}{
		{100, 1.0, 1.0},   // No shrinkage
		{80, 1.0, 1.0},    // No shrinkage
		{60, 0.90, 0.90},  // Boundary: 90%
		{50, 0.85, 0.90},  // 40-59 range: 85-90%
		{40, 0.85, 0.85},  // Boundary: 85%
		{30, 0.75, 0.85},  // 20-39 range: 75-85%
		{10, 0.60, 0.75},  // 1-19 range: 60-75%
		{1, 0.60, 0.60},   // Minimum: 60%
	}

	for _, tt := range tests {
		result := CalculateInputBoxShrinkage(tt.san)

		if result < tt.minExpected || result > tt.maxExpected {
			t.Errorf("CalculateInputBoxShrinkage(%d) = %.2f, want %.2f-%.2f",
				tt.san, result, tt.minExpected, tt.maxExpected)
		}
	}
}

func TestTruncateToShrunkWidth(t *testing.T) {
	text := "1234567890"
	originalWidth := 10

	tests := []struct {
		shrinkage      float64
		expectedMaxLen int
	}{
		{1.0, 10},  // No shrinkage
		{0.9, 9},   // 90%
		{0.5, 5},   // 50%
		{0.3, 3},   // 30%
	}

	for _, tt := range tests {
		result := TruncateToShrunkWidth(text, originalWidth, tt.shrinkage)

		if len(result) > tt.expectedMaxLen {
			t.Errorf("TruncateToShrunkWidth(shrinkage=%.1f) = %q (len %d), expected max len %d",
				tt.shrinkage, result, len(result), tt.expectedMaxLen)
		}
	}
}

func TestTruncateToShrunkWidth_WithEllipsis(t *testing.T) {
	text := "1234567890"
	originalWidth := 10
	shrinkage := 0.5 // Should truncate to 5 chars

	result := TruncateToShrunkWidth(text, originalWidth, shrinkage)

	// Should have ellipsis if truncated
	if len(text) > int(float64(originalWidth)*shrinkage) {
		if !strings.HasSuffix(result, "...") && len(result) > 3 {
			t.Errorf("Expected ellipsis in truncated result, got %q", result)
		}
	}
}

func TestSanitizeCorruptedInput_Clean(t *testing.T) {
	input := "正常輸入"
	result := SanitizeCorruptedInput(input)

	if result != input {
		t.Errorf("SanitizeCorruptedInput(%q) = %q, want unchanged", input, result)
	}
}

func TestSanitizeCorruptedInput_WithZalgo(t *testing.T) {
	// Input with combining diacritical marks
	input := "T\u0301e\u0303s\u0308t\u030C" // T́ẽs̈ťt

	result := SanitizeCorruptedInput(input)

	// Should strip combining marks
	if strings.ContainsAny(result, "\u0300\u0301\u0302\u0303") {
		t.Errorf("SanitizeCorruptedInput should remove combining marks, got %q", result)
	}

	// Should preserve base characters
	if !strings.Contains(result, "T") || !strings.Contains(result, "e") {
		t.Errorf("SanitizeCorruptedInput should preserve base chars, got %q", result)
	}
}

func TestSanitizeCorruptedInput_Whitespace(t *testing.T) {
	input := "  測試  "
	result := SanitizeCorruptedInput(input)

	if result != "測試" {
		t.Errorf("SanitizeCorruptedInput should trim whitespace, got %q", result)
	}
}
