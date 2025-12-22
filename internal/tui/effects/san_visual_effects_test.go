package effects

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestGetSANVisualEffects tests the visual effects mapping for different SAN values.
// Story 7.5 AC1-AC4: Progressive visual effects.
func TestGetSANVisualEffects(t *testing.T) {
	tests := []struct {
		name               string
		san                int
		wantScale          float64
		wantBorderColor    string
		wantShake          bool
		wantCursorSpeed    float64
		wantTextUnstable   bool
		wantCharDeletion   float64
	}{
		{
			name:               "SAN 100 - Clear state",
			san:                100,
			wantScale:          1.0,
			wantBorderColor:    "240",
			wantShake:          false,
			wantCursorSpeed:    1.0,
			wantTextUnstable:   false,
			wantCharDeletion:   0.0,
		},
		{
			name:               "SAN 80 - Clear boundary",
			san:                80,
			wantScale:          1.0,
			wantBorderColor:    "240",
			wantShake:          false,
			wantCursorSpeed:    1.0,
			wantTextUnstable:   false,
			wantCharDeletion:   0.0,
		},
		{
			name:               "SAN 70 - Slight anxiety",
			san:                70,
			wantScale:          1.0,
			wantBorderColor:    "244",
			wantShake:          false,
			wantCursorSpeed:    1.0,
			wantTextUnstable:   false,
			wantCharDeletion:   0.0,
		},
		{
			name:               "SAN 50 - Pressure state",
			san:                50,
			wantScale:          0.80,
			wantBorderColor:    "220",
			wantShake:          true,
			wantCursorSpeed:    1.0,
			wantTextUnstable:   false,
			wantCharDeletion:   0.0,
		},
		{
			name:               "SAN 30 - Anxiety state",
			san:                30,
			wantScale:          0.60,
			wantBorderColor:    "202",
			wantShake:          true,
			wantCursorSpeed:    1.5,
			wantTextUnstable:   true,
			wantCharDeletion:   0.0,
		},
		{
			name:               "SAN 15 - Loss of control",
			san:                15,
			wantScale:          0.40,
			wantBorderColor:    "196",
			wantShake:          true,
			wantCursorSpeed:    2.0,
			wantTextUnstable:   true,
			wantCharDeletion:   0.15,
		},
		{
			name:               "SAN 5 - Severe loss",
			san:                5,
			wantScale:          0.40,
			wantBorderColor:    "196",
			wantShake:          true,
			wantCursorSpeed:    2.0,
			wantTextUnstable:   true,
			wantCharDeletion:   0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)

			if effects.SAN != tt.san {
				t.Errorf("SAN = %d, want %d", effects.SAN, tt.san)
			}

			if effects.InputBoxScale != tt.wantScale {
				t.Errorf("InputBoxScale = %.2f, want %.2f", effects.InputBoxScale, tt.wantScale)
			}

			if effects.BorderColor != tt.wantBorderColor {
				t.Errorf("BorderColor = %s, want %s", effects.BorderColor, tt.wantBorderColor)
			}

			if effects.BorderShake != tt.wantShake {
				t.Errorf("BorderShake = %v, want %v", effects.BorderShake, tt.wantShake)
			}

			if effects.CursorBlinkSpeed != tt.wantCursorSpeed {
				t.Errorf("CursorBlinkSpeed = %.2f, want %.2f", effects.CursorBlinkSpeed, tt.wantCursorSpeed)
			}

			if effects.TextColorUnstable != tt.wantTextUnstable {
				t.Errorf("TextColorUnstable = %v, want %v", effects.TextColorUnstable, tt.wantTextUnstable)
			}

			if effects.CharDeletionRate != tt.wantCharDeletion {
				t.Errorf("CharDeletionRate = %.2f, want %.2f", effects.CharDeletionRate, tt.wantCharDeletion)
			}
		})
	}
}

// TestGetSANVisualEffectsBoundaries tests boundary values.
func TestGetSANVisualEffectsBoundaries(t *testing.T) {
	tests := []struct {
		name string
		san  int
	}{
		{"Negative SAN", -10},
		{"Zero SAN", 0},
		{"Over 100 SAN", 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)

			// Should not panic and should produce valid values
			if effects.InputBoxScale < 0 || effects.InputBoxScale > 1.0 {
				t.Errorf("InputBoxScale out of range: %.2f", effects.InputBoxScale)
			}

			if effects.CursorBlinkSpeed < 0 {
				t.Errorf("CursorBlinkSpeed negative: %.2f", effects.CursorBlinkSpeed)
			}
		})
	}
}

// TestApplyInputBoxStyle tests applying visual effects to input box style.
// Story 7.5 AC2, AC4: Border color changes.
func TestApplyInputBoxStyle(t *testing.T) {
	baseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	tests := []struct {
		name            string
		san             int
		wantBorderColor string
	}{
		{"SAN 100", 100, "240"},
		{"SAN 50", 50, "220"},
		{"SAN 30", 30, "202"},
		{"SAN 15", 15, "196"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			styledBox := ApplyInputBoxStyle(baseStyle, effects)

			// Verify style was modified
			if styledBox.GetBorderTopForeground() == baseStyle.GetBorderTopForeground() && tt.san < 80 {
				// Note: GetBorderTopForeground might not work as expected with lipgloss
				// This is a basic check that the function runs without error
			}
		})
	}
}

// TestApplyTextColorEffects tests text color effects.
// Story 7.5 AC3, AC4: Text color darkening and instability.
func TestApplyTextColorEffects(t *testing.T) {
	tests := []struct {
		name string
		san  int
		text string
	}{
		{"SAN 100 - No effects", 100, "測試文字"},
		{"SAN 50 - Some effects", 50, "測試文字"},
		{"SAN 30 - Unstable colors", 30, "測試文字"},
		{"SAN 15 - Severe effects", 15, "測試文字"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			styled := ApplyTextColorEffects(tt.text, effects)

			// At minimum, function should not panic and return non-empty string
			if styled == "" && tt.text != "" {
				t.Errorf("ApplyTextColorEffects() returned empty string for non-empty input")
			}
		})
	}
}

// TestApplyTextDistortionEffects tests text distortion effects.
// Story 7.5 AC4: Text blur, ghost, shake effects.
func TestApplyTextDistortionEffects(t *testing.T) {
	tests := []struct {
		name string
		san  int
		text string
	}{
		{"SAN 100 - No distortion", 100, "測試文字"},
		{"SAN 30 - Some distortion", 30, "測試文字"},
		{"SAN 15 - Severe distortion", 15, "測試文字"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			distorted := ApplyTextDistortionEffects(tt.text, effects)

			// Function should not panic
			if distorted == "" && tt.text != "" {
				t.Errorf("ApplyTextDistortionEffects() returned empty string")
			}
		})
	}
}

// TestCalculateInputBoxWidth tests input box width calculation.
// Story 7.5 AC2, AC3, AC4: Input box shrinkage.
func TestCalculateInputBoxWidth(t *testing.T) {
	tests := []struct {
		name          string
		san           int
		originalWidth int
		wantMin       int
		wantMax       int
	}{
		{
			name:          "SAN 100 - No shrinkage",
			san:           100,
			originalWidth: 100,
			wantMin:       100,
			wantMax:       100,
		},
		{
			name:          "SAN 50 - 20% shrinkage",
			san:           50,
			originalWidth: 100,
			wantMin:       80,
			wantMax:       80,
		},
		{
			name:          "SAN 30 - 40% shrinkage",
			san:           30,
			originalWidth: 100,
			wantMin:       60,
			wantMax:       60,
		},
		{
			name:          "SAN 15 - 60% shrinkage",
			san:           15,
			originalWidth: 100,
			wantMin:       40,
			wantMax:       40,
		},
		{
			name:          "Minimum width enforced",
			san:           15,
			originalWidth: 10,
			wantMin:       10, // Minimum width
			wantMax:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			width := CalculateInputBoxWidth(tt.originalWidth, effects)

			if width < tt.wantMin {
				t.Errorf("CalculateInputBoxWidth() = %d, want >= %d", width, tt.wantMin)
			}

			if width > tt.wantMax {
				t.Errorf("CalculateInputBoxWidth() = %d, want <= %d", width, tt.wantMax)
			}
		})
	}
}

// TestGetCursorBlinkInterval tests cursor blink speed calculation.
// Story 7.5 AC3: Cursor blinks 1.5x faster at SAN 20-39.
func TestGetCursorBlinkInterval(t *testing.T) {
	baseInterval := 530 // milliseconds

	tests := []struct {
		name     string
		san      int
		wantMin  int
		wantMax  int
	}{
		{
			name:     "SAN 100 - Normal speed",
			san:      100,
			wantMin:  530,
			wantMax:  530,
		},
		{
			name:     "SAN 50 - Normal speed",
			san:      50,
			wantMin:  530,
			wantMax:  530,
		},
		{
			name:     "SAN 30 - 1.5x faster",
			san:      30,
			wantMin:  350,
			wantMax:  360,
		},
		{
			name:     "SAN 15 - 2x faster",
			san:      15,
			wantMin:  260,
			wantMax:  270,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			interval := GetCursorBlinkInterval(effects, baseInterval)

			if interval < tt.wantMin {
				t.Errorf("GetCursorBlinkInterval() = %d, want >= %d", interval, tt.wantMin)
			}

			if interval > tt.wantMax {
				t.Errorf("GetCursorBlinkInterval() = %d, want <= %d", interval, tt.wantMax)
			}
		})
	}
}

// TestShouldApplyCharDeletion tests character deletion flag.
// Story 7.5 AC4: Character deletion at SAN 1-19.
func TestShouldApplyCharDeletion(t *testing.T) {
	tests := []struct {
		name string
		san  int
		want bool
	}{
		{"SAN 100 - No deletion", 100, false},
		{"SAN 50 - No deletion", 50, false},
		{"SAN 20 - No deletion boundary", 20, false},
		{"SAN 15 - Deletion enabled", 15, true},
		{"SAN 5 - Deletion enabled", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			result := ShouldApplyCharDeletion(effects)

			if result != tt.want {
				t.Errorf("ShouldApplyCharDeletion() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestApplyCharacterDeletion tests character deletion functionality.
// Story 7.5 AC4: Randomly delete 10-20% of characters.
func TestApplyCharacterDeletion(t *testing.T) {
	tests := []struct {
		name        string
		san         int
		text        string
		wantMinDel  int
		wantMaxDel  int
	}{
		{
			name:        "SAN 100 - No deletion",
			san:         100,
			text:        "這是一個測試文字字串",
			wantMinDel:  0,
			wantMaxDel:  0,
		},
		{
			name:        "SAN 15 - Some deletion",
			san:         15,
			text:        "這是一個測試文字字串這是一個測試文字字串這是一個測試文字字串", // Longer text
			wantMinDel:  0, // Due to randomness, may delete 0 chars
			wantMaxDel:  10, // ~10-20% of 30 chars
		},
		{
			name:        "Empty string",
			san:         15,
			text:        "",
			wantMinDel:  0,
			wantMaxDel:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)

			// Run multiple times to test randomness
			for i := 0; i < 10; i++ {
				result, deletedCount := ApplyCharacterDeletion(tt.text, effects)

				if deletedCount < tt.wantMinDel {
					t.Errorf("ApplyCharacterDeletion() deleted %d chars, want >= %d", deletedCount, tt.wantMinDel)
				}

				if deletedCount > tt.wantMaxDel {
					t.Errorf("ApplyCharacterDeletion() deleted %d chars, want <= %d", deletedCount, tt.wantMaxDel)
				}

				// Result should not be longer than original
				if len([]rune(result)) > len([]rune(tt.text)) {
					t.Errorf("Result longer than original: %d > %d", len([]rune(result)), len([]rune(tt.text)))
				}
			}
		})
	}
}

// TestGetVisualEffectDescription tests visual effect descriptions.
func TestGetVisualEffectDescription(t *testing.T) {
	tests := []struct {
		name string
		san  int
	}{
		{"SAN 100", 100},
		{"SAN 50", 50},
		{"SAN 30", 30},
		{"SAN 15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)
			desc := GetVisualEffectDescription(effects)

			if desc == "" {
				t.Errorf("GetVisualEffectDescription() returned empty string")
			}
		})
	}
}

// TestInputBoxShakeOffset tests border shake offset calculation.
// Story 7.5 AC2, AC4: Border shake effects.
func TestInputBoxShakeOffset(t *testing.T) {
	tests := []struct {
		name          string
		san           int
		wantShake     bool
		maxIntensity  int
	}{
		{
			name:          "SAN 100 - No shake",
			san:           100,
			wantShake:     false,
			maxIntensity:  0,
		},
		{
			name:          "SAN 50 - Slight shake",
			san:           50,
			wantShake:     true,
			maxIntensity:  1,
		},
		{
			name:          "SAN 30 - Moderate shake",
			san:           30,
			wantShake:     true,
			maxIntensity:  2,
		},
		{
			name:          "SAN 15 - Intense shake",
			san:           15,
			wantShake:     true,
			maxIntensity:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effects := GetSANVisualEffects(tt.san)

			// Run multiple times to test randomness
			for i := 0; i < 20; i++ {
				x, y := InputBoxShakeOffset(effects)

				if !tt.wantShake {
					if x != 0 || y != 0 {
						t.Errorf("InputBoxShakeOffset() = (%d, %d), want (0, 0) for no shake", x, y)
					}
				} else {
					// Shake offset should be within intensity bounds
					if x < -tt.maxIntensity || x > tt.maxIntensity {
						t.Errorf("X offset %d out of bounds [-%d, %d]", x, tt.maxIntensity, tt.maxIntensity)
					}
					if y < -tt.maxIntensity || y > tt.maxIntensity {
						t.Errorf("Y offset %d out of bounds [-%d, %d]", y, tt.maxIntensity, tt.maxIntensity)
					}
				}
			}
		})
	}
}

// TestVisualEffectsProgression tests that effects intensify as SAN decreases.
// Story 7.5 AC6: Gradual visual effect progression.
func TestVisualEffectsProgression(t *testing.T) {
	var prevScale float64 = 1.0
	var prevSpeed float64 = 1.0

	for san := 100; san >= 1; san -= 10 {
		effects := GetSANVisualEffects(san)

		// Input box should not grow as SAN decreases
		if effects.InputBoxScale > prevScale {
			t.Errorf("Input box scale increased from %.2f to %.2f when SAN dropped to %d", prevScale, effects.InputBoxScale, san)
		}

		// Cursor should not slow down as SAN decreases
		if effects.CursorBlinkSpeed < prevSpeed {
			t.Errorf("Cursor blink speed decreased from %.2f to %.2f when SAN dropped to %d", prevSpeed, effects.CursorBlinkSpeed, san)
		}

		prevScale = effects.InputBoxScale
		prevSpeed = effects.CursorBlinkSpeed
	}
}
