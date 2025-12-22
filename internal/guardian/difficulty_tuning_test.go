package guardian

import (
	"math"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/stretchr/testify/assert"
)

// floatEquals compares two float64 values with a small epsilon for floating point precision
func floatEquals(a, b float64) bool {
	epsilon := 0.0001
	return math.Abs(a-b) < epsilon
}

// TestGetDifficultyAdjustment_Disabled tests that adjustment returns 1.0 when disabled
func TestGetDifficultyAdjustment_Disabled(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = false
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(50)
	gameState.SetSAN(50)

	adjustment := guardian.GetDifficultyAdjustment(gameState)

	assert.Equal(t, 1.0, adjustment, "Adjustment should be 1.0 when tuning is disabled")
}

// TestGetDifficultyAdjustment_DisabledWithLowHP tests disabled tuning with low HP
func TestGetDifficultyAdjustment_DisabledWithLowHP(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = false
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)

	adjustment := guardian.GetDifficultyAdjustment(gameState)

	assert.Equal(t, 1.0, adjustment, "Adjustment should be 1.0 even with low stats when disabled")
}

// TestGetDifficultyAdjustment_NilGameState tests nil gameState handling
func TestGetDifficultyAdjustment_NilGameState(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	adjustment := guardian.GetDifficultyAdjustment(nil)

	assert.Equal(t, 1.0, adjustment, "Should return 1.0 for nil gameState")
}

// TestGetDifficultyAdjustment_FullHealth tests maximum difficulty increase
func TestGetDifficultyAdjustment_FullHealth(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(100)
	gameState.SetSAN(100)

	adjustment := guardian.GetDifficultyAdjustment(gameState)

	// survivalRate = (100/100 + 100/100) / 2 = 1.0
	// adjustment = 0.3 + 0.9 * 1.0 = 1.2
	assert.True(t, floatEquals(1.2, adjustment), "Full health should give 1.2 adjustment, got %f", adjustment)
}

// TestGetDifficultyAdjustment_ZeroHealth tests maximum difficulty reduction
func TestGetDifficultyAdjustment_ZeroHealth(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)

	adjustment := guardian.GetDifficultyAdjustment(gameState)

	// survivalRate = (0/100 + 0/100) / 2 = 0.0
	// adjustment = 0.3 + 0.9 * 0.0 = 0.3
	assert.True(t, floatEquals(0.3, adjustment), "Zero health should give 0.3 adjustment, got %f", adjustment)
}

// TestGetDifficultyAdjustment_MidRange tests mid-range adjustment
func TestGetDifficultyAdjustment_MidRange(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(50)
	gameState.SetSAN(50)

	adjustment := guardian.GetDifficultyAdjustment(gameState)

	// survivalRate = (50/100 + 50/100) / 2 = 0.5
	// adjustment = 0.3 + 0.9 * 0.5 = 0.75
	assert.True(t, floatEquals(0.75, adjustment), "HP=50, SAN=50 should give 0.75 adjustment, got %f", adjustment)
}

// TestGetDifficultyAdjustment_AsymmetricStats tests different HP and SAN values
func TestGetDifficultyAdjustment_AsymmetricStats(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	tests := []struct {
		name       string
		hp         int
		san        int
		expectedAdj float64
	}{
		{
			name:       "High HP, Low SAN",
			hp:         100,
			san:        0,
			expectedAdj: 0.75, // survivalRate = (1.0 + 0.0) / 2 = 0.5 -> 0.3 + 0.45 = 0.75
		},
		{
			name:       "Low HP, High SAN",
			hp:         0,
			san:        100,
			expectedAdj: 0.75, // survivalRate = (0.0 + 1.0) / 2 = 0.5 -> 0.3 + 0.45 = 0.75
		},
		{
			name:       "High HP, Mid SAN",
			hp:         100,
			san:        50,
			expectedAdj: 0.975, // survivalRate = (1.0 + 0.5) / 2 = 0.75 -> 0.3 + 0.675 = 0.975
		},
		{
			name:       "Mid HP, High SAN",
			hp:         50,
			san:        100,
			expectedAdj: 0.975, // survivalRate = (0.5 + 1.0) / 2 = 0.75 -> 0.3 + 0.675 = 0.975
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)

			adjustment := guardian.GetDifficultyAdjustment(gameState)

			assert.True(t, floatEquals(tt.expectedAdj, adjustment),
				"HP=%d, SAN=%d should give %.3f adjustment, got %.3f",
				tt.hp, tt.san, tt.expectedAdj, adjustment)
		})
	}
}

// TestGetDifficultyAdjustment_BoundaryValues tests boundary HP/SAN values
func TestGetDifficultyAdjustment_BoundaryValues(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	tests := []struct {
		name        string
		hp          int
		san         int
		expectedAdj float64
	}{
		{"0/0", 0, 0, 0.3},
		{"0/100", 0, 100, 0.75},
		{"100/0", 100, 0, 0.75},
		{"100/100", 100, 100, 1.2},
		{"25/25", 25, 25, 0.525},   // survivalRate = 0.25 -> 0.3 + 0.225 = 0.525
		{"75/75", 75, 75, 0.975},   // survivalRate = 0.75 -> 0.3 + 0.675 = 0.975
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)

			adjustment := guardian.GetDifficultyAdjustment(gameState)

			assert.True(t, floatEquals(tt.expectedAdj, adjustment),
				"HP=%d, SAN=%d should give %.3f adjustment, got %.3f",
				tt.hp, tt.san, tt.expectedAdj, adjustment)
		})
	}
}

// TestGetDifficultyAdjustment_Range tests that adjustment is always in valid range
func TestGetDifficultyAdjustment_Range(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	// Test many combinations to ensure range is always [0.3, 1.2]
	for hp := 0; hp <= 100; hp += 10 {
		for san := 0; san <= 100; san += 10 {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(hp)
			gameState.SetSAN(san)

			adjustment := guardian.GetDifficultyAdjustment(gameState)

			assert.GreaterOrEqual(t, adjustment, 0.3, "Adjustment should be >= 0.3 for HP=%d, SAN=%d", hp, san)
			assert.LessOrEqual(t, adjustment, 1.2, "Adjustment should be <= 1.2 for HP=%d, SAN=%d", hp, san)
		}
	}
}

// TestGetDifficultyAdjustment_Formula tests the exact formula
func TestGetDifficultyAdjustment_Formula(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	tests := []struct {
		hp  int
		san int
	}{
		{20, 30},
		{40, 60},
		{80, 90},
		{15, 85},
		{33, 67},
	}

	for _, tt := range tests {
		gameState := engine.NewGameStateV2()
		gameState.SetHP(tt.hp)
		gameState.SetSAN(tt.san)

		adjustment := guardian.GetDifficultyAdjustment(gameState)

		// Calculate expected value
		hpRate := float64(tt.hp) / 100.0
		sanRate := float64(tt.san) / 100.0
		survivalRate := (hpRate + sanRate) / 2.0
		expected := 0.3 + 0.9*survivalRate

		assert.True(t, floatEquals(expected, adjustment),
			"HP=%d, SAN=%d: expected %.3f, got %.3f", tt.hp, tt.san, expected, adjustment)
	}
}

// TestGetDifficultyAdjustment_EnabledVsDisabled tests the config flag behavior
func TestGetDifficultyAdjustment_EnabledVsDisabled(t *testing.T) {
	gameState := engine.NewGameStateV2()
	gameState.SetHP(30)
	gameState.SetSAN(40)

	// Test with tuning disabled
	cfgDisabled := DefaultGuardianConfig()
	cfgDisabled.EnableDifficultyTuning = false
	guardianDisabled := NewExperienceGuardian(cfgDisabled)
	adjustmentDisabled := guardianDisabled.GetDifficultyAdjustment(gameState)

	assert.Equal(t, 1.0, adjustmentDisabled, "Disabled tuning should return 1.0")

	// Test with tuning enabled
	cfgEnabled := DefaultGuardianConfig()
	cfgEnabled.EnableDifficultyTuning = true
	guardianEnabled := NewExperienceGuardian(cfgEnabled)
	adjustmentEnabled := guardianEnabled.GetDifficultyAdjustment(gameState)

	assert.NotEqual(t, 1.0, adjustmentEnabled, "Enabled tuning should not return 1.0 for non-50/50 stats")
	// HP=30, SAN=40: survivalRate = 0.35 -> adjustment = 0.3 + 0.315 = 0.615
	assert.True(t, floatEquals(0.615, adjustmentEnabled), "Expected 0.615, got %.3f", adjustmentEnabled)
}

// TestGetDifficultyAdjustment_Neutral tests the neutral point
func TestGetDifficultyAdjustment_Neutral(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()

	// Find the point where adjustment = 1.0
	// 1.0 = 0.3 + 0.9 * survivalRate
	// 0.7 = 0.9 * survivalRate
	// survivalRate = 0.7 / 0.9 = 0.7777...
	// For HP=SAN: rate = HP/100 = 0.7777... -> HP = 77.77...

	// Test with HP=78, SAN=78 (close to neutral)
	gameState.SetHP(78)
	gameState.SetSAN(78)
	adjustment := guardian.GetDifficultyAdjustment(gameState)

	// survivalRate = 0.78 -> adjustment = 0.3 + 0.702 = 1.002
	assert.True(t, floatEquals(1.002, adjustment), "HP=78, SAN=78 should give ~1.0 adjustment, got %.3f", adjustment)
}

// TestIntegrationScenario_DifficultyProgression tests difficulty adjustment through game progression
func TestIntegrationScenario_DifficultyProgression(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()

	// Scenario: Player starts healthy, takes damage, gets weaker
	progression := []struct {
		turn        int
		hp          int
		san         int
		expectedMin float64
		expectedMax float64
	}{
		{1, 100, 100, 1.19, 1.21}, // Full health -> 1.2
		{2, 80, 90, 1.05, 1.07},   // survivalRate 0.85 -> 1.065
		{3, 60, 70, 0.88, 0.90},   // survivalRate 0.65 -> 0.885
		{4, 40, 50, 0.70, 0.72},   // survivalRate 0.45 -> 0.705
		{5, 20, 30, 0.52, 0.54},   // survivalRate 0.25 -> 0.525
		{6, 10, 10, 0.38, 0.40},   // survivalRate 0.1 -> 0.39
	}

	for _, p := range progression {
		gameState.SetHP(p.hp)
		gameState.SetSAN(p.san)

		adjustment := guardian.GetDifficultyAdjustment(gameState)

		assert.GreaterOrEqual(t, adjustment, p.expectedMin,
			"Turn %d: adjustment should be >= %.2f for HP=%d, SAN=%d", p.turn, p.expectedMin, p.hp, p.san)
		assert.LessOrEqual(t, adjustment, p.expectedMax,
			"Turn %d: adjustment should be <= %.2f for HP=%d, SAN=%d", p.turn, p.expectedMax, p.hp, p.san)
	}
}

// TestIntegrationScenario_DifficultyRecovery tests difficulty adjustment during recovery
func TestIntegrationScenario_DifficultyRecovery(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	gameState := engine.NewGameStateV2()

	// Scenario: Player recovers from low health
	gameState.SetHP(10)
	gameState.SetSAN(10)
	adj1 := guardian.GetDifficultyAdjustment(gameState)
	assert.True(t, floatEquals(0.39, adj1), "Low health should give low adjustment")

	gameState.SetHP(50)
	gameState.SetSAN(50)
	adj2 := guardian.GetDifficultyAdjustment(gameState)
	assert.True(t, floatEquals(0.75, adj2), "Mid health should give mid adjustment")

	gameState.SetHP(90)
	gameState.SetSAN(90)
	adj3 := guardian.GetDifficultyAdjustment(gameState)
	assert.True(t, floatEquals(1.11, adj3), "High health should give high adjustment")

	// Adjustments should be increasing
	assert.Less(t, adj1, adj2, "Adjustment should increase as health recovers")
	assert.Less(t, adj2, adj3, "Adjustment should continue increasing")
}

// TestDefaultGuardianConfig_DifficultyTuningDisabled tests default config has tuning disabled
func TestDefaultGuardianConfig_DifficultyTuningDisabled(t *testing.T) {
	cfg := DefaultGuardianConfig()

	assert.False(t, cfg.EnableDifficultyTuning, "Default config should have difficulty tuning disabled")
}
