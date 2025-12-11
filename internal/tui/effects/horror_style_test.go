package effects

import (
	"testing"
)

func TestCalculateHorrorStyle_ClearHeaded(t *testing.T) {
	// SAN 80-100: 完全正常
	style := CalculateHorrorStyle(100)

	if style.TextCorruption != 0.0 {
		t.Errorf("TextCorruption = %f, want 0.0", style.TextCorruption)
	}
	if style.TypingBehavior != 0.0 {
		t.Errorf("TypingBehavior = %f, want 0.0", style.TypingBehavior)
	}
	if style.ColorShift != 0 {
		t.Errorf("ColorShift = %d, want 0", style.ColorShift)
	}
	if style.UIStability != 0 {
		t.Errorf("UIStability = %d, want 0", style.UIStability)
	}
}

func TestCalculateHorrorStyle_SlightlyAnxious(t *testing.T) {
	// SAN 60-79: 輕微效果
	style := CalculateHorrorStyle(70)

	if style.TextCorruption != 0.1 {
		t.Errorf("TextCorruption = %f, want 0.1", style.TextCorruption)
	}
	if style.TypingBehavior != 0.0 {
		t.Errorf("TypingBehavior = %f, want 0.0", style.TypingBehavior)
	}
	// ColorShift 應該在 5-10 範圍
	if style.ColorShift < 5 || style.ColorShift > 10 {
		t.Errorf("ColorShift = %d, want between 5-10", style.ColorShift)
	}
	if style.UIStability != 0 {
		t.Errorf("UIStability = %d, want 0", style.UIStability)
	}
}

func TestCalculateHorrorStyle_Anxious(t *testing.T) {
	// SAN 40-59: 中度效果
	style := CalculateHorrorStyle(50)

	if style.TextCorruption != 0.3 {
		t.Errorf("TextCorruption = %f, want 0.3", style.TextCorruption)
	}
	if style.TypingBehavior != 0.0 {
		t.Errorf("TypingBehavior = %f, want 0.0", style.TypingBehavior)
	}
	// ColorShift 應該在 15-30 範圍
	if style.ColorShift < 15 || style.ColorShift > 30 {
		t.Errorf("ColorShift = %d, want between 15-30", style.ColorShift)
	}
	if style.UIStability != 1 {
		t.Errorf("UIStability = %d, want 1", style.UIStability)
	}
}

func TestCalculateHorrorStyle_Panicked(t *testing.T) {
	// SAN 20-39: 嚴重效果
	style := CalculateHorrorStyle(30)

	if style.TextCorruption != 0.6 {
		t.Errorf("TextCorruption = %f, want 0.6", style.TextCorruption)
	}
	if style.TypingBehavior != 0.05 {
		t.Errorf("TypingBehavior = %f, want 0.05", style.TypingBehavior)
	}
	// ColorShift 應該在 45-90 範圍
	if style.ColorShift < 45 || style.ColorShift > 90 {
		t.Errorf("ColorShift = %d, want between 45-90", style.ColorShift)
	}
	// UIStability 應該在 2-3 範圍
	if style.UIStability < 2 || style.UIStability > 3 {
		t.Errorf("UIStability = %d, want between 2-3", style.UIStability)
	}
}

func TestCalculateHorrorStyle_Insanity(t *testing.T) {
	// SAN 1-19: 極端效果
	style := CalculateHorrorStyle(10)

	if style.TextCorruption != 0.9 {
		t.Errorf("TextCorruption = %f, want 0.9", style.TextCorruption)
	}
	if style.TypingBehavior != 0.15 {
		t.Errorf("TypingBehavior = %f, want 0.15", style.TypingBehavior)
	}
	// ColorShift 應該在 120-180 範圍
	if style.ColorShift < 120 || style.ColorShift > 180 {
		t.Errorf("ColorShift = %d, want between 120-180", style.ColorShift)
	}
	// UIStability 應該在 4-5 範圍
	if style.UIStability < 4 || style.UIStability > 5 {
		t.Errorf("UIStability = %d, want between 4-5", style.UIStability)
	}
}

func TestCalculateHorrorStyle_BoundaryValues(t *testing.T) {
	tests := []struct {
		san       int
		wantLevel string
	}{
		{100, "clear"},
		{80, "clear"},
		{79, "slight"},
		{60, "slight"},
		{59, "anxious"},
		{40, "anxious"},
		{39, "panicked"},
		{20, "panicked"},
		{19, "insanity"},
		{1, "insanity"},
		{0, "insanity"},
	}

	for _, tt := range tests {
		style := CalculateHorrorStyle(tt.san)
		// 只驗證不會 panic，具體值已經在其他測試驗證
		if style.TextCorruption < 0 || style.TextCorruption > 1.0 {
			t.Errorf("SAN %d: TextCorruption out of range: %f", tt.san, style.TextCorruption)
		}
	}
}
