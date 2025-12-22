package game

import (
	"testing"
)

// TestGetControlLevel tests the control level calculation for different SAN values.
// Story 7.5 AC6: Gradual control deprivation mapping.
// Code Review Fix 7-4-1: Updated to match san_effects.go boundaries
func TestGetControlLevel(t *testing.T) {
	tests := []struct {
		name                  string
		san                   int
		expectedControl       float64
		expectedVisual        float64
		expectedDistortion    float64
		expectedHallucination float64
		expectedDeletion      float64
	}{
		{
			name:                  "SAN 100 - Full control (Clear)",
			san:                   100,
			expectedControl:       1.0,
			expectedVisual:        0.0,
			expectedDistortion:    0.0,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 80 - Clear state boundary",
			san:                   80,
			expectedControl:       1.0,
			expectedVisual:        0.0,
			expectedDistortion:    0.0,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 70 - Anxious state",
			san:                   70,
			expectedControl:       0.95,
			expectedVisual:        0.15,
			expectedDistortion:    0.0,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 50 - Anxious state boundary",
			san:                   50,
			expectedControl:       0.95,
			expectedVisual:        0.15,
			expectedDistortion:    0.0,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 49 - Panic state",
			san:                   49,
			expectedControl:       0.85,
			expectedVisual:        0.4,
			expectedDistortion:    0.3,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 30 - Panic state mid",
			san:                   30,
			expectedControl:       0.85,
			expectedVisual:        0.4,
			expectedDistortion:    0.3,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 20 - Panic boundary",
			san:                   20,
			expectedControl:       0.85,
			expectedVisual:        0.4,
			expectedDistortion:    0.3,
			expectedHallucination: 0.0,
			expectedDeletion:      0.0,
		},
		{
			name:                  "SAN 19 - Breakdown state",
			san:                   19,
			expectedControl:       0.60,
			expectedVisual:        0.8,
			expectedDistortion:    0.6,
			expectedHallucination: 0.5,
			expectedDeletion:      0.15,
		},
		{
			name:                  "SAN 5 - Severe breakdown",
			san:                   5,
			expectedControl:       0.60,
			expectedVisual:        0.8,
			expectedDistortion:    0.6,
			expectedHallucination: 0.5,
			expectedDeletion:      0.15,
		},
		{
			name:                  "SAN 1 - Minimum",
			san:                   1,
			expectedControl:       0.60,
			expectedVisual:        0.8,
			expectedDistortion:    0.6,
			expectedHallucination: 0.5,
			expectedDeletion:      0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)

			if level.SAN != tt.san {
				t.Errorf("SAN = %d, want %d", level.SAN, tt.san)
			}

			if level.ControlPercentage != tt.expectedControl {
				t.Errorf("ControlPercentage = %.2f, want %.2f", level.ControlPercentage, tt.expectedControl)
			}

			if level.VisualInterference != tt.expectedVisual {
				t.Errorf("VisualInterference = %.2f, want %.2f", level.VisualInterference, tt.expectedVisual)
			}

			if level.TextDistortion != tt.expectedDistortion {
				t.Errorf("TextDistortion = %.2f, want %.2f", level.TextDistortion, tt.expectedDistortion)
			}

			if level.HallucinationRate != tt.expectedHallucination {
				t.Errorf("HallucinationRate = %.2f, want %.2f", level.HallucinationRate, tt.expectedHallucination)
			}

			if level.CharDeletionRate != tt.expectedDeletion {
				t.Errorf("CharDeletionRate = %.2f, want %.2f", level.CharDeletionRate, tt.expectedDeletion)
			}
		})
	}
}

// TestGetControlLevelBoundaries tests boundary values.
func TestGetControlLevelBoundaries(t *testing.T) {
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
			level := GetControlLevel(tt.san)

			// Should not panic and should clamp values
			if level.ControlPercentage < 0 || level.ControlPercentage > 1.0 {
				t.Errorf("ControlPercentage out of range: %.2f", level.ControlPercentage)
			}
		})
	}
}

// TestGetCharDeletionCount tests character deletion count calculation.
// Story 7.5 AC4: Delete 10-20% of characters when SAN 1-19.
func TestGetCharDeletionCount(t *testing.T) {
	tests := []struct {
		name       string
		san        int
		textLength int
		wantMin    int
		wantMax    int
	}{
		{
			name:       "SAN 100 - No deletion",
			san:        100,
			textLength: 100,
			wantMin:    0,
			wantMax:    0,
		},
		{
			name:       "SAN 20 - No deletion boundary",
			san:        20,
			textLength: 100,
			wantMin:    0,
			wantMax:    0,
		},
		{
			name:       "SAN 15 - Deletion occurs",
			san:        15,
			textLength: 100,
			wantMin:    5,  // At least 5% due to randomness
			wantMax:    25, // At most 25% due to randomness
		},
		{
			name:       "SAN 5 - High deletion",
			san:        5,
			textLength: 100,
			wantMin:    5,
			wantMax:    25,
		},
		{
			name:       "Empty text",
			san:        5,
			textLength: 0,
			wantMin:    0,
			wantMax:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)

			// Run multiple times to test randomness
			for i := 0; i < 10; i++ {
				count := level.GetCharDeletionCount(tt.textLength)

				if count < tt.wantMin {
					t.Errorf("GetCharDeletionCount() = %d, want >= %d", count, tt.wantMin)
				}

				if count > tt.wantMax {
					t.Errorf("GetCharDeletionCount() = %d, want <= %d", count, tt.wantMax)
				}
			}
		})
	}
}

// TestShouldShowHallucinationOption tests hallucination option probability.
// Story 7.5 AC5: Hallucination options when SAN < 20.
func TestShouldShowHallucinationOption(t *testing.T) {
	tests := []struct {
		name            string
		san             int
		expectNever     bool
		expectSometimes bool
	}{
		{
			name:            "SAN 100 - Never",
			san:             100,
			expectNever:     true,
			expectSometimes: false,
		},
		{
			name:            "SAN 20 - Never boundary",
			san:             20,
			expectNever:     true,
			expectSometimes: false,
		},
		{
			name:            "SAN 15 - Sometimes",
			san:             15,
			expectNever:     false,
			expectSometimes: true,
		},
		{
			name:            "SAN 5 - Sometimes",
			san:             5,
			expectNever:     false,
			expectSometimes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)

			trueCount := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				if level.ShouldShowHallucinationOption() {
					trueCount++
				}
			}

			if tt.expectNever && trueCount > 0 {
				t.Errorf("Expected never to show hallucination, but got %d/%d", trueCount, iterations)
			}

			if tt.expectSometimes && trueCount == 0 {
				t.Errorf("Expected sometimes to show hallucination, but got 0/%d", iterations)
			}

			if tt.expectSometimes && trueCount == iterations {
				t.Errorf("Expected sometimes to show hallucination, but got all %d/%d", trueCount, iterations)
			}
		})
	}
}

// TestGetHallucinationCount tests hallucination count generation.
// Story 7.5 AC5: Insert 1-2 hallucination options when SAN < 20.
func TestGetHallucinationCount(t *testing.T) {
	tests := []struct {
		name    string
		san     int
		wantMin int
		wantMax int
	}{
		{
			name:    "SAN 100 - No hallucinations",
			san:     100,
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "SAN 20 - No hallucinations boundary",
			san:     20,
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "SAN 15 - 0-2 hallucinations",
			san:     15,
			wantMin: 0,
			wantMax: 2,
		},
		{
			name:    "SAN 5 - 0-2 hallucinations",
			san:     5,
			wantMin: 0,
			wantMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)

			// Test multiple times for randomness
			counts := make(map[int]int)
			iterations := 100

			for i := 0; i < iterations; i++ {
				count := level.GetHallucinationCount()
				counts[count]++

				if count < tt.wantMin {
					t.Errorf("GetHallucinationCount() = %d, want >= %d", count, tt.wantMin)
				}

				if count > tt.wantMax {
					t.Errorf("GetHallucinationCount() = %d, want <= %d", count, tt.wantMax)
				}
			}

			// For SAN < 20, we should see a mix of 0, 1, and 2
			if tt.san < 20 {
				// At least some variation expected
				if len(counts) == 1 && iterations > 10 {
					t.Errorf("Expected variation in hallucination count, but got only %v", counts)
				}
			}
		})
	}
}

// TestDescribeControlState tests control state descriptions.
// Code Review Fix 7-4-1: Updated for aligned boundaries
func TestDescribeControlState(t *testing.T) {
	tests := []struct {
		name string
		san  int
		want string
	}{
		{"SAN 100", 100, "完全控制"},
		{"SAN 50", 50, "幾乎完全控制"},  // 95% control
		{"SAN 30", 30, "部分控制受限"},  // 85% control
		{"SAN 15", 15, "控制嚴重受限"},  // 60% control
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)
			desc := level.DescribeControlState()

			if desc == "" {
				t.Errorf("DescribeControlState() returned empty string")
			}

			if desc != tt.want {
				t.Errorf("DescribeControlState() = %q, want %q", desc, tt.want)
			}
		})
	}
}

// TestDescribePlayerFeeling tests player feeling descriptions.
func TestDescribePlayerFeeling(t *testing.T) {
	tests := []struct {
		name string
		san  int
	}{
		{"SAN 100", 100},
		{"SAN 70", 70},
		{"SAN 50", 50},
		{"SAN 30", 30},
		{"SAN 15", 15},
		{"SAN 5", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetControlLevel(tt.san)
			feeling := level.DescribePlayerFeeling()

			if feeling == "" {
				t.Errorf("DescribePlayerFeeling() returned empty string")
			}
		})
	}
}

// TestControlLevelProgression tests that control degrades smoothly as SAN decreases.
// Story 7.5 AC6: Gradual control deprivation.
func TestControlLevelProgression(t *testing.T) {
	var prevControl float64 = 1.0

	for san := 100; san >= 1; san -= 10 {
		level := GetControlLevel(san)

		// Control should never increase as SAN decreases
		if level.ControlPercentage > prevControl {
			t.Errorf("Control increased from %.2f to %.2f when SAN dropped to %d", prevControl, level.ControlPercentage, san)
		}

		prevControl = level.ControlPercentage
	}
}

// TestControlNeverFullyRemoved tests that player always retains some control.
// Story 7.5 AC6: Never completely remove control.
func TestControlNeverFullyRemoved(t *testing.T) {
	for san := 100; san >= 0; san -= 5 {
		level := GetControlLevel(san)

		if level.ControlPercentage <= 0.0 {
			t.Errorf("Control completely removed at SAN %d, want > 0.0", san)
		}

		// Story requirement: Minimum 60% control at lowest SAN
		if san < 20 && level.ControlPercentage < 0.60 {
			t.Errorf("Control %.2f below minimum 0.60 at SAN %d", level.ControlPercentage, san)
		}
	}
}
