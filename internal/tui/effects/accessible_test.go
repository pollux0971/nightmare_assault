package effects

import (
	"strings"
	"testing"
)

func TestApplyAccessibleEffects_FullEffects(t *testing.T) {
	AccessibleMode = false
	defer func() { AccessibleMode = false }()

	style := HorrorStyle{TextCorruption: 0.6}
	text := "測試文字"

	result := ApplyAccessibleEffects(text, style)

	// In full effects mode, should apply Zalgo
	if !strings.Contains(result, "測") || !strings.Contains(result, "試") {
		t.Errorf("Expected Zalgo to preserve base characters, got %q", result)
	}

	// Should have combining characters (length > original due to combining marks)
	if len(result) <= len(text) {
		t.Errorf("Expected Zalgo to add combining characters")
	}
}

func TestApplyAccessibleEffects_AccessibleMode(t *testing.T) {
	AccessibleMode = true
	defer func() { AccessibleMode = false }()

	tests := []struct {
		corruption float64
		expected   string
		text       string
	}{
		{0.95, "[文字嚴重混亂]", "測試"},
		{0.7, "[文字混亂]", "測試"},
		{0.4, "[文字微亂]", "測試"},
		{0.2, "測試", "測試"}, // Minimal, no description
		{0.05, "測試", "測試"}, // Below threshold
	}

	for _, tt := range tests {
		style := HorrorStyle{TextCorruption: tt.corruption}
		result := ApplyAccessibleEffects(tt.text, style)

		if !strings.Contains(result, tt.expected) {
			t.Errorf("ApplyAccessibleEffects(corruption=%.2f) = %q, want to contain %q",
				tt.corruption, result, tt.expected)
		}
	}
}

func TestGetAccessibleSANStateDescription_AllRanges(t *testing.T) {
	tests := []struct {
		san         int
		shouldEmpty bool
		contains    string
	}{
		{100, true, ""},
		{85, true, ""},
		{80, true, ""},
		{75, false, "模糊"},
		{60, false, "模糊"},
		{55, false, "不太真實"},
		{40, false, "不太真實"},
		{35, false, "混亂"},
		{20, false, "混亂"},
		{15, false, "幻覺"},
		{1, false, "幻覺"},
		{0, false, "崩塌"},
	}

	for _, tt := range tests {
		result := GetAccessibleSANStateDescription(tt.san)

		if tt.shouldEmpty {
			if result != "" {
				t.Errorf("GetAccessibleSANStateDescription(%d) = %q, want empty string", tt.san, result)
			}
		} else {
			if result == "" {
				t.Errorf("GetAccessibleSANStateDescription(%d) returned empty, want description", tt.san)
			}
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("GetAccessibleSANStateDescription(%d) = %q, want to contain %q",
					tt.san, result, tt.contains)
			}
		}
	}
}

func TestScaleEffectIntensity_FullMode(t *testing.T) {
	AccessibleMode = false
	defer func() { AccessibleMode = false }()

	style := HorrorStyle{
		TextCorruption:    0.8,
		TypingBehavior:    0.1,
		ColorShift:        100,
		UIStability:       4,
		OptionReliability: 0.5,
	}

	result := ScaleEffectIntensity(style)

	// Should return unchanged
	if result.TextCorruption != 0.8 {
		t.Errorf("Expected TextCorruption unchanged, got %.2f", result.TextCorruption)
	}
	if result.ColorShift != 100 {
		t.Errorf("Expected ColorShift unchanged, got %d", result.ColorShift)
	}
}

func TestScaleEffectIntensity_AccessibleMode(t *testing.T) {
	AccessibleMode = true
	defer func() { AccessibleMode = false }()

	style := HorrorStyle{
		TextCorruption:    0.8,
		TypingBehavior:    0.1,
		ColorShift:        100,
		UIStability:       4,
		OptionReliability: 0.5,
	}

	result := ScaleEffectIntensity(style)

	// All effects should be reduced by 50%
	expectedTextCorruption := 0.8 * 0.5
	if result.TextCorruption != expectedTextCorruption {
		t.Errorf("TextCorruption = %.2f, want %.2f", result.TextCorruption, expectedTextCorruption)
	}

	expectedTypingBehavior := 0.1 * 0.5
	if result.TypingBehavior != expectedTypingBehavior {
		t.Errorf("TypingBehavior = %.2f, want %.2f", result.TypingBehavior, expectedTypingBehavior)
	}

	expectedColorShift := 100 / 2
	if result.ColorShift != expectedColorShift {
		t.Errorf("ColorShift = %d, want %d", result.ColorShift, expectedColorShift)
	}

	expectedUIStability := 4 / 2
	if result.UIStability != expectedUIStability {
		t.Errorf("UIStability = %d, want %d", result.UIStability, expectedUIStability)
	}

	// OptionReliability should remain unchanged
	if result.OptionReliability != 0.5 {
		t.Errorf("OptionReliability = %.2f, want %.2f", result.OptionReliability, 0.5)
	}
}

func TestScaleEffectIntensity_BoundaryValues(t *testing.T) {
	AccessibleMode = true
	defer func() { AccessibleMode = false }()

	// Test with maximum values
	maxStyle := HorrorStyle{
		TextCorruption:    1.0,
		TypingBehavior:    0.2,
		ColorShift:        360,
		UIStability:       5,
		OptionReliability: 1.0,
	}

	result := ScaleEffectIntensity(maxStyle)

	if result.TextCorruption != 0.5 {
		t.Errorf("Max TextCorruption scaled = %.2f, want 0.50", result.TextCorruption)
	}
	if result.ColorShift != 180 {
		t.Errorf("Max ColorShift scaled = %d, want 180", result.ColorShift)
	}

	// Test with zero values
	zeroStyle := HorrorStyle{
		TextCorruption:    0.0,
		TypingBehavior:    0.0,
		ColorShift:        0,
		UIStability:       0,
		OptionReliability: 0.0,
	}

	result = ScaleEffectIntensity(zeroStyle)

	if result.TextCorruption != 0.0 {
		t.Errorf("Zero TextCorruption scaled = %.2f, want 0.00", result.TextCorruption)
	}
	if result.ColorShift != 0 {
		t.Errorf("Zero ColorShift scaled = %d, want 0", result.ColorShift)
	}
}
