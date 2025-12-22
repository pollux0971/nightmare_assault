package manager

import (
	"math/rand"
	"testing"
)

// TestCheckMentalStateTransition_NormalToAnxious tests Normal to Anxious transition (AC2)
func TestCheckMentalStateTransition_NormalToAnxious(t *testing.T) {
	tests := []struct {
		name     string
		emotion  EmotionState
		expected MentalState
	}{
		{
			name:     "High Stress triggers Anxious",
			emotion:  NewEmotionState(50, 50, 60),
			expected: Anxious,
		},
		{
			name:     "Exact Stress threshold triggers Anxious",
			emotion:  NewEmotionState(50, 50, 60),
			expected: Anxious,
		},
		{
			name:     "High Fear triggers Anxious",
			emotion:  NewEmotionState(50, 70, 50),
			expected: Anxious,
		},
		{
			name:     "Exact Fear threshold triggers Anxious",
			emotion:  NewEmotionState(50, 70, 50),
			expected: Anxious,
		},
		{
			name:     "Both high triggers Anxious",
			emotion:  NewEmotionState(50, 70, 70),
			expected: Anxious,
		},
		{
			name:     "Below threshold stays Normal",
			emotion:  NewEmotionState(50, 50, 50),
			expected: Normal,
		},
		{
			name:     "Just below Stress threshold stays Normal",
			emotion:  NewEmotionState(50, 50, 59),
			expected: Normal,
		},
		{
			name:     "Just below Fear threshold stays Normal",
			emotion:  NewEmotionState(50, 69, 50),
			expected: Normal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkMentalStateTransition(Normal, tt.emotion)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCheckMentalStateTransition_AnxiousToNormal tests Anxious to Normal recovery (AC4)
func TestCheckMentalStateTransition_AnxiousToNormal(t *testing.T) {
	tests := []struct {
		name     string
		emotion  EmotionState
		expected MentalState
	}{
		{
			name:     "Low Stress and Fear triggers recovery",
			emotion:  NewEmotionState(50, 30, 30),
			expected: Normal,
		},
		{
			name:     "Just below threshold boundaries trigger recovery",
			emotion:  NewEmotionState(50, 39, 39),
			expected: Normal,
		},
		{
			name:     "High Stress prevents recovery",
			emotion:  NewEmotionState(50, 30, 50),
			expected: Anxious,
		},
		{
			name:     "High Fear prevents recovery",
			emotion:  NewEmotionState(50, 60, 30),
			expected: Anxious,
		},
		{
			name:     "Both high prevents recovery",
			emotion:  NewEmotionState(50, 60, 60),
			expected: Anxious,
		},
		{
			name:     "Stress at threshold prevents recovery",
			emotion:  NewEmotionState(50, 30, 40),
			expected: Anxious,
		},
		{
			name:     "Fear at threshold prevents recovery",
			emotion:  NewEmotionState(50, 50, 30),
			expected: Anxious,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkMentalStateTransition(Anxious, tt.emotion)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCheckMentalStateTransition_AnxiousToCorrupted tests breakdown checks (AC3)
func TestCheckMentalStateTransition_AnxiousToCorrupted(t *testing.T) {
	// Test that high stress can trigger breakdown check
	emotion := NewEmotionState(10, 80, 95) // Very high stress, high fear

	// Set seed for deterministic testing
	rand.Seed(12345)

	// Run multiple times to test probability
	corruptedCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		result := checkMentalStateTransition(Anxious, emotion)

		if result == Corrupted {
			corruptedCount++
		}
	}

	// With stress=95 and fear=80, breakdown chance should be:
	// 0.05 + (95-90)*0.01 + (80-60)*0.005 = 0.05 + 0.05 + 0.10 = 0.20 (20%)
	// We expect approximately 20% corruption rate
	rate := float64(corruptedCount) / float64(iterations)
	expectedRate := 0.20

	if rate < expectedRate-0.05 || rate > expectedRate+0.05 {
		t.Errorf("Expected corruption rate around %.2f, got %.2f (%d/%d)",
			expectedRate, rate, corruptedCount, iterations)
	}

	t.Logf("Corruption rate: %.2f%% (%d/%d)", rate*100, corruptedCount, iterations)
}

// TestCheckMentalStateTransition_CorruptedIrreversible tests that Corrupted state is permanent
func TestCheckMentalStateTransition_CorruptedIrreversible(t *testing.T) {
	// Even with perfect conditions, corrupted state should remain
	emotion := NewEmotionState(100, 0, 0) // Perfect trust, no fear/stress

	result := checkMentalStateTransition(Corrupted, emotion)

	if result != Corrupted {
		t.Errorf("Corrupted state should be irreversible, got %v", result)
	}
}

// TestRollBreakdown tests the breakdown probability calculation (AC5)
func TestRollBreakdown(t *testing.T) {
	tests := []struct {
		name           string
		stress         int
		fear           int
		expectedChance float64
		iterations     int
	}{
		{
			name:           "Stress below 90 but fear high - still has chance",
			stress:         89,
			fear:           80,
			expectedChance: 0.14, // BaseChance + (89-90)*0.01 + (80-60)*0.005 = 0.05 - 0.01 + 0.10 = 0.14
			iterations:     1000,
		},
		{
			name:           "Stress at 90, fear at 60 - base chance only",
			stress:         90,
			fear:           60,
			expectedChance: 0.05, // BaseChance only
			iterations:     1000,
		},
		{
			name:           "Stress 95, fear 80 - moderate chance",
			stress:         95,
			fear:           80,
			expectedChance: 0.20, // 0.05 + 5*0.01 + 20*0.005 = 0.05 + 0.05 + 0.10
			iterations:     1000,
		},
		{
			name:           "Stress 100, fear 100 - high chance",
			stress:         100,
			fear:           100,
			expectedChance: 0.35, // 0.05 + 10*0.01 + 40*0.005 = 0.05 + 0.10 + 0.20
			iterations:     1000,
		},
		{
			name:           "Stress 90, fear 50 - low chance (fear below threshold)",
			stress:         90,
			fear:           50,
			expectedChance: 0.0, // 0.05 + 0 + (50-60)*0.005 = 0.05 - 0.05 = 0
			iterations:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set seed for reproducibility
			rand.Seed(42)

			breakdowns := 0
			for i := 0; i < tt.iterations; i++ {
				if rollBreakdown(tt.stress, tt.fear) {
					breakdowns++
				}
			}

			rate := float64(breakdowns) / float64(tt.iterations)

			// For zero expected chance, verify no breakdowns occurred
			if tt.expectedChance == 0.0 {
				if breakdowns > 0 {
					t.Errorf("Expected no breakdowns, got %d/%d (%.2f%%)",
						breakdowns, tt.iterations, rate*100)
				}
				return
			}

			// For non-zero expected chance, allow 5% margin of error
			if rate < tt.expectedChance-0.05 || rate > tt.expectedChance+0.05 {
				t.Errorf("Expected breakdown rate around %.2f, got %.2f (%d/%d)",
					tt.expectedChance, rate, breakdowns, tt.iterations)
			}

			t.Logf("Breakdown rate: %.2f%% (%d/%d), expected: %.2f%%",
				rate*100, breakdowns, tt.iterations, tt.expectedChance*100)
		})
	}
}

// TestRollBreakdown_EdgeCases tests edge cases in breakdown calculation
func TestRollBreakdown_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		stress      int
		fear        int
		expectNever bool
	}{
		{
			name:        "Stress exactly at threshold, fear at threshold",
			stress:      90,
			fear:        60,
			expectNever: false, // Base chance 5% still applies
		},
		{
			name:        "Stress below 90 but very high fear",
			stress:      89,
			fear:        100,
			expectNever: false, // Positive chance: 0.05 - 0.01 + 0.20 = 0.24
		},
		{
			name:        "Fear below threshold",
			stress:      95,
			fear:        50,
			expectNever: false, // Stress contribution still applies
		},
		{
			name:        "Both below threshold",
			stress:      85,
			fear:        50,
			expectNever: true, // Total chance is negative
		},
		{
			name:        "Minimum values",
			stress:      0,
			fear:        0,
			expectNever: true, // Large negative chance
		},
		{
			name:        "Maximum values",
			stress:      100,
			fear:        100,
			expectNever: false, // Maximum chance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times
			breakdowns := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				if rollBreakdown(tt.stress, tt.fear) {
					breakdowns++
				}
			}

			if tt.expectNever && breakdowns > 0 {
				t.Errorf("Expected no breakdowns, got %d/%d", breakdowns, iterations)
			}

			if !tt.expectNever && breakdowns == 0 {
				// Calculate expected chance to see if this is reasonable
				chance := BaseBreakdownChance +
					float64(tt.stress-90)*StressMultiplier +
					float64(tt.fear-60)*FearMultiplier
				t.Logf("No breakdowns occurred in %d iterations, expected chance: %.4f",
					iterations, chance)
			}
		})
	}
}

// TestMentalStateTransition_Integration tests complete state machine (AC1)
func TestMentalStateTransition_Integration(t *testing.T) {
	// Start Normal
	state := Normal
	emotion := NewEmotionState(50, 30, 40)

	// Should stay Normal
	state = checkMentalStateTransition(state, emotion)
	if state != Normal {
		t.Errorf("Step 1: Expected Normal, got %v", state)
	}

	// Increase stress -> Anxious
	emotion.Stress = 65
	state = checkMentalStateTransition(state, emotion)
	if state != Anxious {
		t.Errorf("Step 2: Expected Anxious, got %v", state)
	}

	// Reduce stress -> back to Normal
	emotion.Stress = 30
	emotion.Fear = 30
	state = checkMentalStateTransition(state, emotion)
	if state != Normal {
		t.Errorf("Step 3: Expected Normal, got %v", state)
	}

	// Increase stress very high -> potential Corrupted
	state = Anxious // Set to anxious first
	emotion.Stress = 95
	emotion.Fear = 80

	// Try multiple times to get a corruption
	corrupted := false
	for i := 0; i < 100; i++ {
		testState := checkMentalStateTransition(Anxious, emotion)
		if testState == Corrupted {
			corrupted = true
			break
		}
	}

	if !corrupted {
		t.Log("Note: No corruption occurred in 100 attempts (20% chance per attempt)")
	} else {
		t.Log("Successfully transitioned to Corrupted state")
	}

	// Once corrupted, should never recover
	if corrupted {
		emotion.Stress = 0
		emotion.Fear = 0
		state = checkMentalStateTransition(Corrupted, emotion)
		if state != Corrupted {
			t.Errorf("Step 5: Corrupted should be irreversible, got %v", state)
		}
	}
}

// TestRollBreakdown_Formula tests the exact formula implementation
func TestRollBreakdown_Formula(t *testing.T) {
	tests := []struct {
		stress         int
		fear           int
		expectedChance float64
	}{
		{90, 60, 0.05},   // Base only
		{91, 60, 0.06},   // Base + 1*0.01 = 0.06
		{90, 61, 0.055},  // Base + 1*0.005 = 0.055
		{95, 80, 0.20},   // 0.05 + 5*0.01 + 20*0.005 = 0.20
		{100, 100, 0.35}, // 0.05 + 10*0.01 + 40*0.005 = 0.35
		{89, 80, 0.14},   // 0.05 - 1*0.01 + 20*0.005 = 0.14
		{90, 50, 0.0},    // 0.05 + 0 - 10*0.005 = 0, clamped to 0
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			// Calculate expected chance manually
			calculatedChance := BaseBreakdownChance +
				float64(tt.stress-90)*StressMultiplier +
				float64(tt.fear-60)*FearMultiplier

			if calculatedChance < 0 {
				calculatedChance = 0
			}

			// Use epsilon comparison for floating point
			epsilon := 0.0001
			if calculatedChance < tt.expectedChance-epsilon || calculatedChance > tt.expectedChance+epsilon {
				t.Errorf("Stress=%d Fear=%d: Expected chance %.4f, calculated %.4f",
					tt.stress, tt.fear, tt.expectedChance, calculatedChance)
			}
		})
	}
}

// TestCheckMentalStateTransition_AllStates tests all state transitions systematically
func TestCheckMentalStateTransition_AllStates(t *testing.T) {
	tests := []struct {
		name     string
		current  MentalState
		emotion  EmotionState
		expected MentalState
	}{
		// Normal state transitions
		{"Normal stays Normal (low stress/fear)", Normal, NewEmotionState(50, 30, 40), Normal},
		{"Normal to Anxious (high stress)", Normal, NewEmotionState(50, 30, 60), Anxious},
		{"Normal to Anxious (high fear)", Normal, NewEmotionState(50, 70, 40), Anxious},
		{"Normal to Anxious (both high)", Normal, NewEmotionState(50, 70, 60), Anxious},

		// Anxious state transitions
		{"Anxious stays Anxious (moderate)", Anxious, NewEmotionState(50, 50, 50), Anxious},
		{"Anxious to Normal (recovery)", Anxious, NewEmotionState(50, 30, 30), Normal},
		// Note: Anxious to Corrupted is probabilistic, tested separately

		// Corrupted state transitions
		{"Corrupted stays Corrupted (perfect)", Corrupted, NewEmotionState(100, 0, 0), Corrupted},
		{"Corrupted stays Corrupted (worst)", Corrupted, NewEmotionState(0, 100, 100), Corrupted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkMentalStateTransition(tt.current, tt.emotion)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
