package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestApplyDifficultyAdjustment tests applying difficulty adjustments to rule violations
func TestApplyDifficultyAdjustment(t *testing.T) {
	engine := NewRuleEngine()

	tests := []struct {
		name            string
		violation       RuleViolation
		damageReduction float64
		expectedHP      int
		expectedSAN     int
		expectedFatal   bool
	}{
		{
			name: "No Reduction",
			violation: RuleViolation{
				RuleID:    "test-1",
				RuleName:  "Test Rule",
				HPDamage:  50,
				SANDamage: 30,
				IsFatal:   false,
			},
			damageReduction: 0.0,
			expectedHP:      50,
			expectedSAN:     30,
			expectedFatal:   false,
		},
		{
			name: "30% Reduction",
			violation: RuleViolation{
				RuleID:    "test-2",
				RuleName:  "Test Rule",
				HPDamage:  50,
				SANDamage: 30,
				IsFatal:   false,
			},
			damageReduction: 0.3,
			expectedHP:      35, // 50 - (50 * 0.3) = 35
			expectedSAN:     21, // 30 - (30 * 0.3) = 21
			expectedFatal:   false,
		},
		{
			name: "10% Reduction",
			violation: RuleViolation{
				RuleID:    "test-3",
				RuleName:  "Test Rule",
				HPDamage:  100,
				SANDamage: 50,
				IsFatal:   false,
			},
			damageReduction: 0.1,
			expectedHP:      90, // 100 - (100 * 0.1) = 90
			expectedSAN:     45, // 50 - (50 * 0.1) = 45
			expectedFatal:   false,
		},
		{
			name: "Fatal Remains Fatal",
			violation: RuleViolation{
				RuleID:    "test-4",
				RuleName:  "Test Rule",
				HPDamage:  100,
				SANDamage: 100,
				IsFatal:   true,
			},
			damageReduction: 0.3,
			expectedHP:      70,  // Damage is reduced
			expectedSAN:     70,  // Damage is reduced
			expectedFatal:   true, // But fatal flag remains
		},
		{
			name: "Only HP Damage",
			violation: RuleViolation{
				RuleID:    "test-5",
				RuleName:  "Test Rule",
				HPDamage:  60,
				SANDamage: 0,
				IsFatal:   false,
			},
			damageReduction: 0.2,
			expectedHP:      48, // 60 - (60 * 0.2) = 48
			expectedSAN:     0,
			expectedFatal:   false,
		},
		{
			name: "Only SAN Damage",
			violation: RuleViolation{
				RuleID:    "test-6",
				RuleName:  "Test Rule",
				HPDamage:  0,
				SANDamage: 40,
				IsFatal:   false,
			},
			damageReduction: 0.25,
			expectedHP:      0,
			expectedSAN:     30, // 40 - (40 * 0.25) = 30
			expectedFatal:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := engine.ApplyDifficultyAdjustment(tt.violation, tt.damageReduction)

			assert.Equal(t, tt.expectedHP, adjusted.HPDamage,
				"HP damage should be %d, got %d", tt.expectedHP, adjusted.HPDamage)
			assert.Equal(t, tt.expectedSAN, adjusted.SANDamage,
				"SAN damage should be %d, got %d", tt.expectedSAN, adjusted.SANDamage)
			assert.Equal(t, tt.expectedFatal, adjusted.IsFatal,
				"Fatal flag should be %v", tt.expectedFatal)

			// Original violation should be unchanged
			assert.Equal(t, tt.violation.HPDamage, tt.violation.HPDamage,
				"Original violation should not be modified")
		})
	}
}

// TestApplyDifficultyAdjustment_NegativeProtection tests damage doesn't go negative
func TestApplyDifficultyAdjustment_NegativeProtection(t *testing.T) {
	engine := NewRuleEngine()

	violation := RuleViolation{
		RuleID:    "test",
		RuleName:  "Test Rule",
		HPDamage:  10,
		SANDamage: 5,
		IsFatal:   false,
	}

	// Extreme reduction that would cause negative damage
	adjusted := engine.ApplyDifficultyAdjustment(violation, 0.99)

	assert.GreaterOrEqual(t, adjusted.HPDamage, 0, "HP damage should not be negative")
	assert.GreaterOrEqual(t, adjusted.SANDamage, 0, "SAN damage should not be negative")
}

// TestAdjustPunishmentDamage tests adjusting punishment damage
func TestAdjustPunishmentDamage(t *testing.T) {
	engine := NewRuleEngine()

	tests := []struct {
		name            string
		punishment      HiddenRulePunishment
		damageReduction float64
		expectedHP      int
		expectedSAN     int
		expectedFatal   bool
	}{
		{
			name: "No Reduction",
			punishment: HiddenRulePunishment{
				HPDamage:  40,
				SANDamage: 20,
				IsFatal:   false,
			},
			damageReduction: 0.0,
			expectedHP:      40,
			expectedSAN:     20,
			expectedFatal:   false,
		},
		{
			name: "30% Reduction",
			punishment: HiddenRulePunishment{
				HPDamage:  40,
				SANDamage: 20,
				IsFatal:   false,
			},
			damageReduction: 0.3,
			expectedHP:      28, // 40 - (40 * 0.3) = 28
			expectedSAN:     14, // 20 - (20 * 0.3) = 14
			expectedFatal:   false,
		},
		{
			name: "Fatal Remains Fatal",
			punishment: HiddenRulePunishment{
				HPDamage:  100,
				SANDamage: 100,
				IsFatal:   true,
			},
			damageReduction: 0.3,
			expectedHP:      70,
			expectedSAN:     70,
			expectedFatal:   true, // Fatal flag preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := engine.AdjustPunishmentDamage(tt.punishment, tt.damageReduction)

			assert.Equal(t, tt.expectedHP, adjusted.HPDamage)
			assert.Equal(t, tt.expectedSAN, adjusted.SANDamage)
			assert.Equal(t, tt.expectedFatal, adjusted.IsFatal)
		})
	}
}

// TestAdjustPunishmentDamage_NegativeProtection tests punishment damage doesn't go negative
func TestAdjustPunishmentDamage_NegativeProtection(t *testing.T) {
	engine := NewRuleEngine()

	punishment := HiddenRulePunishment{
		HPDamage:  5,
		SANDamage: 3,
		IsFatal:   false,
	}

	adjusted := engine.AdjustPunishmentDamage(punishment, 0.99)

	assert.GreaterOrEqual(t, adjusted.HPDamage, 0)
	assert.GreaterOrEqual(t, adjusted.SANDamage, 0)
}

// TestIntegrationScenario_DifficultyAdjustmentWorkflow tests complete workflow
func TestIntegrationScenario_DifficultyAdjustmentWorkflow(t *testing.T) {
	engine := NewRuleEngine()

	// Scenario: Player violates a rule while Guardian protection is active

	// Step 1: Create a rule violation
	violation := RuleViolation{
		RuleID:             "mirror-rule",
		RuleName:           "倒影殺手",
		HPDamage:           50,
		SANDamage:          30,
		IsFatal:            false,
		ViolationNarrative: "你直視鏡中的倒影...",
	}

	// Step 2: Guardian determines player is struggling (survival rate < 0.5)
	// Difficulty adjustment = 30% damage reduction
	damageReduction := 0.3

	// Step 3: Apply difficulty adjustment to violation
	adjustedViolation := engine.ApplyDifficultyAdjustment(violation, damageReduction)

	// Verify damage was reduced
	assert.Equal(t, 35, adjustedViolation.HPDamage, "HP damage should be reduced from 50 to 35")
	assert.Equal(t, 21, adjustedViolation.SANDamage, "SAN damage should be reduced from 30 to 21")
	assert.False(t, adjustedViolation.IsFatal)

	// Step 4: Verify original violation unchanged
	assert.Equal(t, 50, violation.HPDamage, "Original violation should remain unchanged")

	// The adjusted violation would then be applied to the game state
	// resulting in less damage and giving the struggling player a better chance
}

// TestIntegrationScenario_PreventiveAdjustment tests adjusting rules before violations
func TestIntegrationScenario_PreventiveAdjustment(t *testing.T) {
	engine := NewRuleEngine()

	// Scenario: Adjust rule punishments when Guardian activates
	// to prevent future violations from being too harsh

	originalPunishment := HiddenRulePunishment{
		HPDamage:     60,
		SANDamage:    40,
		IsFatal:      false,
		CustomEffect: "你的視線開始扭曲...",
	}

	// Guardian is active with 25% damage reduction
	damageReduction := 0.25

	adjustedPunishment := engine.AdjustPunishmentDamage(originalPunishment, damageReduction)

	assert.Equal(t, 45, adjustedPunishment.HPDamage, "HP damage reduced from 60 to 45")
	assert.Equal(t, 30, adjustedPunishment.SANDamage, "SAN damage reduced from 40 to 30")
	assert.Equal(t, originalPunishment.CustomEffect, adjustedPunishment.CustomEffect,
		"Custom effect should be preserved")
}

// TestIntegrationScenario_ProgressiveReduction tests increasing reduction as player struggles more
func TestIntegrationScenario_ProgressiveReduction(t *testing.T) {
	engine := NewRuleEngine()

	baseViolation := RuleViolation{
		RuleID:    "test",
		RuleName:  "Test",
		HPDamage:  100,
		SANDamage: 50,
		IsFatal:   false,
	}

	// Simulate progressive difficulty reduction as player continues to struggle

	// Stage 1: Slight struggle (10% reduction)
	stage1 := engine.ApplyDifficultyAdjustment(baseViolation, 0.1)
	assert.Equal(t, 90, stage1.HPDamage)
	assert.Equal(t, 45, stage1.SANDamage)

	// Stage 2: Moderate struggle (20% reduction)
	stage2 := engine.ApplyDifficultyAdjustment(baseViolation, 0.2)
	assert.Equal(t, 80, stage2.HPDamage)
	assert.Equal(t, 40, stage2.SANDamage)

	// Stage 3: Severe struggle (30% reduction)
	stage3 := engine.ApplyDifficultyAdjustment(baseViolation, 0.3)
	assert.Equal(t, 70, stage3.HPDamage)
	assert.Equal(t, 35, stage3.SANDamage)

	// Verify damage is progressively reduced
	assert.Greater(t, stage1.HPDamage, stage2.HPDamage)
	assert.Greater(t, stage2.HPDamage, stage3.HPDamage)
}

// TestIntegrationScenario_FatalRulesPreservation tests that fatal rules remain dangerous
func TestIntegrationScenario_FatalRulesPreservation(t *testing.T) {
	engine := NewRuleEngine()

	// Even with Guardian active, fatal violations should remain fatal
	// This maintains game stakes and prevents trivialization

	fatalViolation := RuleViolation{
		RuleID:    "instant-death",
		RuleName:  "致命規則",
		HPDamage:  100,
		SANDamage: 100,
		IsFatal:   true,
	}

	// Apply maximum reduction
	adjusted := engine.ApplyDifficultyAdjustment(fatalViolation, 0.3)

	// Damage is reduced (gives player a chance if they're not at full health)
	assert.Less(t, adjusted.HPDamage, fatalViolation.HPDamage)

	// But the fatal flag is preserved
	assert.True(t, adjusted.IsFatal,
		"Fatal violations should remain fatal even with Guardian active")
}

// TestRoundingBehavior tests integer rounding in damage calculations
func TestRoundingBehavior(t *testing.T) {
	engine := NewRuleEngine()

	tests := []struct {
		name       string
		damage     int
		reduction  float64
		expectedHP int
	}{
		{"Round Down 1", 33, 0.3, 23},  // 33 - 9.9 = 23.1 -> 23
		{"Round Down 2", 47, 0.15, 39}, // 47 - 7.05 = 39.95 -> 39
		{"Exact", 50, 0.2, 40},         // 50 - 10 = 40
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := RuleViolation{
				RuleID:   "test",
				HPDamage: tt.damage,
			}

			adjusted := engine.ApplyDifficultyAdjustment(violation, tt.reduction)
			assert.Equal(t, tt.expectedHP, adjusted.HPDamage)
		})
	}
}
