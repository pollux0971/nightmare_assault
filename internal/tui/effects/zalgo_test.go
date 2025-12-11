package effects

import (
	"strings"
	"testing"
	"unicode"
)

func TestApplyZalgo_ZeroIntensity(t *testing.T) {
	text := "Hello World"
	result := ApplyZalgo(text, 0.0)

	if result != text {
		t.Errorf("ApplyZalgo with 0.0 intensity should return original text, got %q", result)
	}
}

func TestApplyZalgo_FullIntensity(t *testing.T) {
	text := "Test"
	result := ApplyZalgo(text, 1.0)

	// 應該包含組合字符
	hasCombining := false
	for _, r := range result {
		if unicode.In(r, unicode.Mn) { // Mn = Nonspacing Mark (combining characters)
			hasCombining = true
			break
		}
	}

	if !hasCombining {
		t.Error("ApplyZalgo with 1.0 intensity should add combining characters")
	}

	// 應該比原文字長（因為加了組合字符）
	if len([]rune(result)) <= len([]rune(text)) {
		t.Error("ApplyZalgo should add characters at full intensity")
	}
}

func TestApplyZalgo_PartialIntensity(t *testing.T) {
	text := "AAAAAAAA" // 8 個字符
	result := ApplyZalgo(text, 0.5)

	// 50% intensity 應該影響約一半的字符
	originalRunes := []rune(text)
	resultRunes := []rune(result)

	if len(resultRunes) == len(originalRunes) {
		t.Error("ApplyZalgo with 0.5 intensity should add some combining characters")
	}
}

func TestApplyZalgo_EmptyString(t *testing.T) {
	result := ApplyZalgo("", 0.5)
	if result != "" {
		t.Errorf("ApplyZalgo on empty string should return empty string, got %q", result)
	}
}

func TestApplyZalgo_ChineseCharacters(t *testing.T) {
	text := "測試文字"
	result := ApplyZalgo(text, 0.3)

	// 應該仍然包含原始中文字符（組合字符加在後面）
	if !strings.Contains(result, "測") {
		t.Error("ApplyZalgo should preserve original Chinese characters")
	}
}

func TestApplyZalgo_IntensityRange(t *testing.T) {
	tests := []struct {
		intensity float64
		shouldAdd bool
	}{
		{0.0, false},
		{0.5, true},  // 使用較高的機率以確保穩定測試
		{0.9, true},
		{1.0, true},
	}

	text := "AAAAAAAAAAAAAAAA" // 16 個字符增加機率

	for _, tt := range tests {
		result := ApplyZalgo(text, tt.intensity)
		hasExtra := len([]rune(result)) > len([]rune(text))

		if tt.shouldAdd && !hasExtra {
			t.Errorf("Intensity %f: should add combining characters", tt.intensity)
		}
		if !tt.shouldAdd && hasExtra {
			t.Errorf("Intensity %f: should NOT add combining characters", tt.intensity)
		}
	}
}

func TestApplyZalgo_MaxCombiningLimit(t *testing.T) {
	// 根據 Dev Notes，每字符最多 3 個組合標記
	text := "A"
	result := ApplyZalgo(text, 1.0)

	resultRunes := []rune(result)
	combiningCount := 0

	for _, r := range resultRunes {
		if unicode.In(r, unicode.Mn) {
			combiningCount++
		}
	}

	if combiningCount > 3 {
		t.Errorf("ApplyZalgo added %d combining marks, should be max 3 per character", combiningCount)
	}
}
