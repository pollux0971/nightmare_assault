package guardian

import (
	"math"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDifficultyTuner tests the constructor
func TestNewDifficultyTuner(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	tuner := NewDifficultyTuner(guardian)

	assert.NotNil(t, tuner)
	assert.Equal(t, guardian, tuner.guardian)
	assert.False(t, tuner.modifiers.IsActive, "Modifiers should start inactive")
	assert.Equal(t, 0.0, tuner.modifiers.DamageReduction)
	assert.Equal(t, 0, tuner.modifiers.CheckBonus)
	assert.Equal(t, 0.0, tuner.modifiers.EncounterFrequencyReduction)
}

// TestNewDifficultyTunerWithConfig tests constructor with custom config
func TestNewDifficultyTunerWithConfig(t *testing.T) {
	guardianCfg := DefaultGuardianConfig()
	guardianCfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(guardianCfg)

	tuningConfig := DifficultyTuningConfig{
		MaxDamageReduction:          0.5,
		MaxCheckBonus:               20,
		MaxEncounterReduction:       0.7,
		MinSurvivalRateForReduction: 0.6,
	}

	tuner := NewDifficultyTunerWithConfig(guardian, tuningConfig)

	assert.NotNil(t, tuner)
	assert.Equal(t, tuningConfig, tuner.config)
}

// TestAdjustDifficulty_NilGameState tests nil game state handling
func TestAdjustDifficulty_NilGameState(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	modifiers := tuner.AdjustDifficulty(nil)

	assert.False(t, modifiers.IsActive)
}

// TestAdjustDifficulty_TuningDisabled tests behavior when tuning is disabled
func TestAdjustDifficulty_TuningDisabled(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = false
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(10)
	gameState.SetSAN(10)

	modifiers := tuner.AdjustDifficulty(gameState)

	assert.False(t, modifiers.IsActive)
	assert.Equal(t, 0.0, modifiers.DamageReduction)
	assert.Equal(t, 0, modifiers.CheckBonus)
	assert.Equal(t, 0.0, modifiers.EncounterFrequencyReduction)
}

// TestAdjustDifficulty_HealthyPlayer tests no adjustments for healthy players
func TestAdjustDifficulty_HealthyPlayer(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	tests := []struct {
		name string
		hp   int
		san  int
	}{
		{"Full Health", 100, 100},
		{"High Health", 80, 90},
		{"Threshold", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)

			modifiers := tuner.AdjustDifficulty(gameState)

			assert.False(t, modifiers.IsActive, "Healthy players should not get adjustments")
			assert.Equal(t, 0.0, modifiers.DamageReduction)
			assert.Equal(t, 0, modifiers.CheckBonus)
			assert.Equal(t, 0.0, modifiers.EncounterFrequencyReduction)
		})
	}
}

// TestAdjustDifficulty_StrugglingPlayer tests adjustments for struggling players
func TestAdjustDifficulty_StrugglingPlayer(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	tests := []struct {
		name                        string
		hp                          int
		san                         int
		expectedMinDamageReduction  float64
		expectedMaxDamageReduction  float64
		expectedMinCheckBonus       int
		expectedMaxCheckBonus       int
		expectedMinEncounterReduction float64
		expectedMaxEncounterReduction float64
	}{
		{
			name:                        "Critical State",
			hp:                          0,
			san:                         0,
			expectedMinDamageReduction:  0.29,
			expectedMaxDamageReduction:  0.31,
			expectedMinCheckBonus:       14,
			expectedMaxCheckBonus:       15,
			expectedMinEncounterReduction: 0.49,
			expectedMaxEncounterReduction: 0.51,
		},
		{
			name:                        "Low Health",
			hp:                          20,
			san:                         30,
			expectedMinDamageReduction:  0.14,
			expectedMaxDamageReduction:  0.16,
			expectedMinCheckBonus:       7,
			expectedMaxCheckBonus:       8,
			expectedMinEncounterReduction: 0.24,
			expectedMaxEncounterReduction: 0.26,
		},
		{
			name:                        "Slightly Below Threshold",
			hp:                          40,
			san:                         50,
			expectedMinDamageReduction:  0.02,
			expectedMaxDamageReduction:  0.04,
			expectedMinCheckBonus:       1,
			expectedMaxCheckBonus:       2,
			expectedMinEncounterReduction: 0.04,
			expectedMaxEncounterReduction: 0.06,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)

			modifiers := tuner.AdjustDifficulty(gameState)

			assert.True(t, modifiers.IsActive, "Struggling players should get adjustments")

			assert.GreaterOrEqual(t, modifiers.DamageReduction, tt.expectedMinDamageReduction)
			assert.LessOrEqual(t, modifiers.DamageReduction, tt.expectedMaxDamageReduction)

			assert.GreaterOrEqual(t, modifiers.CheckBonus, tt.expectedMinCheckBonus)
			assert.LessOrEqual(t, modifiers.CheckBonus, tt.expectedMaxCheckBonus)

			assert.GreaterOrEqual(t, modifiers.EncounterFrequencyReduction, tt.expectedMinEncounterReduction)
			assert.LessOrEqual(t, modifiers.EncounterFrequencyReduction, tt.expectedMaxEncounterReduction)
		})
	}
}

// TestAdjustDifficulty_MaxReductions tests maximum reduction values
func TestAdjustDifficulty_MaxReductions(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)

	modifiers := tuner.AdjustDifficulty(gameState)

	assert.True(t, modifiers.IsActive)
	assert.LessOrEqual(t, modifiers.DamageReduction, 0.3, "Damage reduction should not exceed 30%")
	assert.LessOrEqual(t, modifiers.CheckBonus, 15, "Check bonus should not exceed +15")
	assert.LessOrEqual(t, modifiers.EncounterFrequencyReduction, 0.5, "Encounter reduction should not exceed 50%")
}

// TestGetModifiers tests getting current modifiers
func TestGetModifiers(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	// Initially no modifiers
	modifiers := tuner.GetModifiers()
	assert.False(t, modifiers.IsActive)

	// After adjustment
	gameState := engine.NewGameStateV2()
	gameState.SetHP(20)
	gameState.SetSAN(20)

	tuner.AdjustDifficulty(gameState)
	modifiers = tuner.GetModifiers()

	assert.True(t, modifiers.IsActive)
	assert.Greater(t, modifiers.DamageReduction, 0.0)
	assert.Greater(t, modifiers.CheckBonus, 0)
	assert.Greater(t, modifiers.EncounterFrequencyReduction, 0.0)
}

// TestResetModifiers tests resetting modifiers
func TestResetModifiers(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	// Set up some modifiers
	gameState := engine.NewGameStateV2()
	gameState.SetHP(10)
	gameState.SetSAN(10)
	tuner.AdjustDifficulty(gameState)

	modifiers := tuner.GetModifiers()
	require.True(t, modifiers.IsActive, "Modifiers should be active before reset")

	// Reset
	tuner.ResetModifiers()

	modifiers = tuner.GetModifiers()
	assert.False(t, modifiers.IsActive)
	assert.Equal(t, 0.0, modifiers.DamageReduction)
	assert.Equal(t, 0, modifiers.CheckBonus)
	assert.Equal(t, 0.0, modifiers.EncounterFrequencyReduction)
}

// TestApplyDamageModifier tests damage reduction application
func TestApplyDamageModifier(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	tests := []struct {
		name           string
		hp             int
		san            int
		originalDamage int
		expectReduced  bool
	}{
		{
			name:           "No Reduction - Healthy",
			hp:             100,
			san:            100,
			originalDamage: 50,
			expectReduced:  false,
		},
		{
			name:           "Reduction - Critical",
			hp:             0,
			san:            0,
			originalDamage: 50,
			expectReduced:  true,
		},
		{
			name:           "Reduction - Low Health",
			hp:             20,
			san:            30,
			originalDamage: 40,
			expectReduced:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)
			tuner.AdjustDifficulty(gameState)

			adjustedDamage := tuner.ApplyDamageModifier(tt.originalDamage)

			if tt.expectReduced {
				assert.Less(t, adjustedDamage, tt.originalDamage, "Damage should be reduced")
				assert.GreaterOrEqual(t, adjustedDamage, 0, "Damage should not be negative")
			} else {
				assert.Equal(t, tt.originalDamage, adjustedDamage, "Damage should not be modified")
			}
		})
	}
}

// TestApplyDamageModifier_NegativeProtection tests that damage doesn't go negative
func TestApplyDamageModifier_NegativeProtection(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)

	// Custom config with extreme reduction
	tuningConfig := DifficultyTuningConfig{
		MaxDamageReduction:          0.99, // 99% reduction
		MaxCheckBonus:               50,
		MaxEncounterReduction:       0.99,
		MinSurvivalRateForReduction: 0.5,
	}
	tuner := NewDifficultyTunerWithConfig(guardian, tuningConfig)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)
	tuner.AdjustDifficulty(gameState)

	adjustedDamage := tuner.ApplyDamageModifier(10)
	assert.GreaterOrEqual(t, adjustedDamage, 0, "Damage should never be negative")
}

// TestApplyCheckModifier tests check bonus application
func TestApplyCheckModifier(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	tests := []struct {
		name                 string
		hp                   int
		san                  int
		originalDifficulty   int
		expectEasier         bool
	}{
		{
			name:                 "No Change - Healthy",
			hp:                   100,
			san:                  100,
			originalDifficulty:   50,
			expectEasier:         false,
		},
		{
			name:                 "Easier - Critical",
			hp:                   0,
			san:                  0,
			originalDifficulty:   50,
			expectEasier:         true,
		},
		{
			name:                 "Easier - Low Health",
			hp:                   20,
			san:                  30,
			originalDifficulty:   30,
			expectEasier:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tt.hp)
			gameState.SetSAN(tt.san)
			tuner.AdjustDifficulty(gameState)

			adjustedDifficulty := tuner.ApplyCheckModifier(tt.originalDifficulty)

			if tt.expectEasier {
				assert.Less(t, adjustedDifficulty, tt.originalDifficulty, "Difficulty should be reduced")
				assert.GreaterOrEqual(t, adjustedDifficulty, 1, "Difficulty should be at least 1")
			} else {
				assert.Equal(t, tt.originalDifficulty, adjustedDifficulty, "Difficulty should not be modified")
			}
		})
	}
}

// TestApplyCheckModifier_MinimumDifficulty tests that difficulty stays at minimum 1
func TestApplyCheckModifier_MinimumDifficulty(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)
	tuner.AdjustDifficulty(gameState)

	// Very low difficulty that should be clamped to 1
	adjustedDifficulty := tuner.ApplyCheckModifier(5)
	assert.GreaterOrEqual(t, adjustedDifficulty, 1, "Difficulty should never go below 1")
}

// TestShouldReduceEncounter tests encounter reduction logic
func TestShouldReduceEncounter(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	// Set up modifiers for critical state
	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)
	tuner.AdjustDifficulty(gameState)

	modifiers := tuner.GetModifiers()
	require.True(t, modifiers.IsActive)
	require.Greater(t, modifiers.EncounterFrequencyReduction, 0.0)

	// Test various encounter chances
	tests := []struct {
		encounterChance  float64
		expectReduction  bool
	}{
		{0.0, true},   // Very low chance - should reduce
		{0.1, true},   // Low chance - should reduce
		{0.3, true},   // Below reduction threshold - should reduce
		{0.6, false},  // Above reduction threshold - should not reduce
		{0.9, false},  // High chance - should not reduce
		{1.0, false},  // Maximum chance - should not reduce
	}

	for _, tt := range tests {
		shouldReduce := tuner.ShouldReduceEncounter(tt.encounterChance)
		if tt.expectReduction {
			assert.True(t, shouldReduce, "Encounter with chance %.2f should be reduced", tt.encounterChance)
		} else {
			assert.False(t, shouldReduce, "Encounter with chance %.2f should not be reduced", tt.encounterChance)
		}
	}
}

// TestShouldReduceEncounter_NoModifiers tests no reduction when inactive
func TestShouldReduceEncounter_NoModifiers(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	// No adjustments - player is healthy
	gameState := engine.NewGameStateV2()
	gameState.SetHP(100)
	gameState.SetSAN(100)
	tuner.AdjustDifficulty(gameState)

	shouldReduce := tuner.ShouldReduceEncounter(0.1)
	assert.False(t, shouldReduce, "Should not reduce encounters when modifiers are inactive")
}

// TestDifficultyProgression tests difficulty adjustment through game progression
func TestDifficultyProgression(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()

	progression := []struct {
		turn        int
		hp          int
		san         int
		expectActive bool
	}{
		{1, 100, 100, false}, // Healthy - no adjustments
		{2, 80, 90, false},   // Still healthy
		{3, 60, 70, false},   // At threshold
		{4, 40, 50, true},    // Below threshold - adjustments start
		{5, 20, 30, true},    // Struggling - stronger adjustments
		{6, 10, 10, true},    // Critical - maximum adjustments
	}

	var previousDamageReduction float64

	for _, p := range progression {
		gameState.SetHP(p.hp)
		gameState.SetSAN(p.san)

		modifiers := tuner.AdjustDifficulty(gameState)

		if p.expectActive {
			assert.True(t, modifiers.IsActive, "Turn %d: Modifiers should be active", p.turn)
			assert.Greater(t, modifiers.DamageReduction, 0.0, "Turn %d: Should have damage reduction", p.turn)
			assert.Greater(t, modifiers.CheckBonus, 0, "Turn %d: Should have check bonus", p.turn)

			// Reductions should increase as health decreases
			if previousDamageReduction > 0 {
				assert.GreaterOrEqual(t, modifiers.DamageReduction, previousDamageReduction,
					"Turn %d: Damage reduction should increase or stay same", p.turn)
			}
			previousDamageReduction = modifiers.DamageReduction
		} else {
			assert.False(t, modifiers.IsActive, "Turn %d: Modifiers should not be active", p.turn)
		}
	}
}

// TestCustomConfig tests tuner with custom configuration
func TestCustomConfig(t *testing.T) {
	guardianCfg := DefaultGuardianConfig()
	guardianCfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(guardianCfg)

	customConfig := DifficultyTuningConfig{
		MaxDamageReduction:          0.5,  // 50% max reduction
		MaxCheckBonus:               20,   // +20 max bonus
		MaxEncounterReduction:       0.7,  // 70% max reduction
		MinSurvivalRateForReduction: 0.6,  // Start helping at 60% survival
	}

	tuner := NewDifficultyTunerWithConfig(guardian, customConfig)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)

	modifiers := tuner.AdjustDifficulty(gameState)

	assert.True(t, modifiers.IsActive)

	// Should use custom max values
	epsilon := 0.01
	assert.True(t, math.Abs(modifiers.DamageReduction - 0.5) < epsilon,
		"Should use custom max damage reduction: got %.3f", modifiers.DamageReduction)
	assert.Equal(t, 20, modifiers.CheckBonus, "Should use custom max check bonus")
	assert.True(t, math.Abs(modifiers.EncounterFrequencyReduction - 0.7) < epsilon,
		"Should use custom max encounter reduction: got %.3f", modifiers.EncounterFrequencyReduction)
}

// TestIntegrationScenario_GuardianWithTuner tests integration between Guardian and Tuner
func TestIntegrationScenario_GuardianWithTuner(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	cfg.MaxConsecutiveDeaths = 2
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()

	// Scenario: Player dies twice, then Guardian activates
	gameState.SetHP(0)
	gameState.SetSAN(0)
	guardian.OnTurnEnd(gameState)

	gameState.SetHP(0)
	gameState.SetSAN(0)
	guardian.OnTurnEnd(gameState)

	shouldProtect := guardian.ShouldActivateProtection(gameState)
	require.True(t, shouldProtect, "Guardian should activate protection after 2 deaths")

	// Apply difficulty adjustments
	modifiers := tuner.AdjustDifficulty(gameState)
	assert.True(t, modifiers.IsActive, "Difficulty adjustments should be active")
	assert.Greater(t, modifiers.DamageReduction, 0.0)

	// Test applying modifiers
	damage := 50
	adjustedDamage := tuner.ApplyDamageModifier(damage)
	assert.Less(t, adjustedDamage, damage, "Damage should be reduced")

	difficulty := 30
	adjustedDifficulty := tuner.ApplyCheckModifier(difficulty)
	assert.Less(t, adjustedDifficulty, difficulty, "Difficulty should be reduced")
}

// TestDefaultDifficultyTuningConfig tests default configuration values
func TestDefaultDifficultyTuningConfig(t *testing.T) {
	config := DefaultDifficultyTuningConfig()

	assert.Equal(t, 0.3, config.MaxDamageReduction)
	assert.Equal(t, 15, config.MaxCheckBonus)
	assert.Equal(t, 0.5, config.MaxEncounterReduction)
	assert.Equal(t, 0.5, config.MinSurvivalRateForReduction)
}
