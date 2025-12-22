package guardian

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationScenario_GuardianWithRuleEngine tests the complete integration
// between Guardian, DifficultyTuner, and RuleEngine.
// Story 9-8: This demonstrates the end-to-end difficulty adjustment workflow.
func TestIntegrationScenario_GuardianWithRuleEngine(t *testing.T) {
	// Setup Guardian with difficulty tuning enabled
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	cfg.MaxConsecutiveDeaths = 2
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)
	ruleEngine := rules.NewRuleEngine()

	gameState := engine.NewGameStateV2()

	// Scenario: Player is struggling, dies twice
	// Guardian activates, difficulty adjustments are applied

	// Turn 1: Player dies
	gameState.SetHP(0)
	gameState.SetSAN(0)
	guardian.OnTurnEnd(gameState)

	// Turn 2: Player dies again
	gameState.SetHP(0)
	gameState.SetSAN(0)
	guardian.OnTurnEnd(gameState)

	// Guardian should activate protection
	shouldProtect := guardian.ShouldActivateProtection(gameState)
	require.True(t, shouldProtect, "Guardian should activate after 2 deaths")

	// Apply difficulty adjustments
	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive, "Difficulty modifiers should be active")
	require.Greater(t, modifiers.DamageReduction, 0.0, "Damage reduction should be applied")

	// Create a rule violation
	violation := rules.RuleViolation{
		RuleID:             "mirror-rule",
		RuleName:           "倒影殺手",
		HPDamage:           50,
		SANDamage:          30,
		IsFatal:            false,
		ViolationNarrative: "你直視鏡中的倒影...",
	}

	// Apply difficulty adjustment to the violation
	adjustedViolation := ruleEngine.ApplyDifficultyAdjustment(violation, modifiers.DamageReduction)

	// Verify damage was reduced
	assert.Less(t, adjustedViolation.HPDamage, violation.HPDamage,
		"HP damage should be reduced by Guardian protection")
	assert.Less(t, adjustedViolation.SANDamage, violation.SANDamage,
		"SAN damage should be reduced by Guardian protection")
	assert.False(t, adjustedViolation.IsFatal)

	// Verify the reduction is approximately 30% (maximum at 0 HP/SAN)
	expectedHPDamage := int(float64(violation.HPDamage) * (1.0 - modifiers.DamageReduction))
	expectedSANDamage := int(float64(violation.SANDamage) * (1.0 - modifiers.DamageReduction))

	assert.Equal(t, expectedHPDamage, adjustedViolation.HPDamage)
	assert.Equal(t, expectedSANDamage, adjustedViolation.SANDamage)

	// Log the results for verification
	t.Logf("Guardian Protection Activated:")
	t.Logf("  Consecutive Deaths: %d", guardian.GetConsecutiveDeaths())
	t.Logf("  Damage Reduction: %.2f%%", modifiers.DamageReduction*100)
	t.Logf("  Check Bonus: +%d", modifiers.CheckBonus)
	t.Logf("  Encounter Reduction: %.2f%%", modifiers.EncounterFrequencyReduction*100)
	t.Logf("Original Violation: HP=%d, SAN=%d", violation.HPDamage, violation.SANDamage)
	t.Logf("Adjusted Violation: HP=%d, SAN=%d", adjustedViolation.HPDamage, adjustedViolation.SANDamage)
}

// TestIntegrationScenario_ProgressiveProtection tests how protection increases
// as player health decreases.
func TestIntegrationScenario_ProgressiveProtection(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)
	ruleEngine := rules.NewRuleEngine()

	baseViolation := rules.RuleViolation{
		RuleID:    "test-rule",
		RuleName:  "測試規則",
		HPDamage:  100,
		SANDamage: 50,
		IsFatal:   false,
	}

	testCases := []struct {
		name               string
		hp                 int
		san                int
		minExpectedHP      int
		maxExpectedHP      int
		minExpectedSAN     int
		maxExpectedSAN     int
	}{
		{
			name:           "Healthy - No Protection",
			hp:             100,
			san:            100,
			minExpectedHP:  100, // No reduction
			maxExpectedHP:  100,
			minExpectedSAN: 50,
			maxExpectedSAN: 50,
		},
		{
			name:           "Moderate Health - Some Protection",
			hp:             40,
			san:            50,
			minExpectedHP:  95, // Small reduction
			maxExpectedHP:  100,
			minExpectedSAN: 47,
			maxExpectedSAN: 50,
		},
		{
			name:           "Low Health - Strong Protection",
			hp:             20,
			san:            30,
			minExpectedHP:  80, // Moderate reduction
			maxExpectedHP:  90,
			minExpectedSAN: 40,
			maxExpectedSAN: 45,
		},
		{
			name:           "Critical - Maximum Protection",
			hp:             0,
			san:            0,
			minExpectedHP:  65, // Maximum reduction (~30%)
			maxExpectedHP:  75,
			minExpectedSAN: 32,
			maxExpectedSAN: 37,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetHP(tc.hp)
			gameState.SetSAN(tc.san)

			modifiers := tuner.AdjustDifficulty(gameState)
			adjustedViolation := ruleEngine.ApplyDifficultyAdjustment(baseViolation, modifiers.DamageReduction)

			assert.GreaterOrEqual(t, adjustedViolation.HPDamage, tc.minExpectedHP,
				"HP damage should be at least %d, got %d", tc.minExpectedHP, adjustedViolation.HPDamage)
			assert.LessOrEqual(t, adjustedViolation.HPDamage, tc.maxExpectedHP,
				"HP damage should be at most %d, got %d", tc.maxExpectedHP, adjustedViolation.HPDamage)

			assert.GreaterOrEqual(t, adjustedViolation.SANDamage, tc.minExpectedSAN,
				"SAN damage should be at least %d, got %d", tc.minExpectedSAN, adjustedViolation.SANDamage)
			assert.LessOrEqual(t, adjustedViolation.SANDamage, tc.maxExpectedSAN,
				"SAN damage should be at most %d, got %d", tc.maxExpectedSAN, adjustedViolation.SANDamage)

			t.Logf("HP: %d, SAN: %d -> Damage Reduction: %.2f%% -> HP: %d->%d, SAN: %d->%d",
				tc.hp, tc.san,
				modifiers.DamageReduction*100,
				baseViolation.HPDamage, adjustedViolation.HPDamage,
				baseViolation.SANDamage, adjustedViolation.SANDamage)
		})
	}
}

// TestIntegrationScenario_CheckModifierWithRules tests applying check bonuses
// to difficulty checks in rule generation.
func TestIntegrationScenario_CheckModifierWithRules(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(10)
	gameState.SetSAN(10)

	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive)
	require.Greater(t, modifiers.CheckBonus, 0)

	// Simulate rule difficulty checks
	testChecks := []struct {
		name               string
		originalDifficulty int
		expectedMax        int
	}{
		{"Easy Check", 10, 10},
		{"Medium Check", 20, 20},
		{"Hard Check", 30, 30},
		{"Very Hard Check", 50, 50},
	}

	for _, tc := range testChecks {
		t.Run(tc.name, func(t *testing.T) {
			adjustedDifficulty := tuner.ApplyCheckModifier(tc.originalDifficulty)

			assert.Less(t, adjustedDifficulty, tc.originalDifficulty,
				"Check difficulty should be reduced")
			assert.GreaterOrEqual(t, adjustedDifficulty, 1,
				"Check difficulty should never go below 1")

			reductionAmount := tc.originalDifficulty - adjustedDifficulty
			assert.LessOrEqual(t, reductionAmount, modifiers.CheckBonus,
				"Reduction should not exceed check bonus")

			t.Logf("Difficulty: %d -> %d (bonus: +%d)",
				tc.originalDifficulty, adjustedDifficulty, modifiers.CheckBonus)
		})
	}
}

// TestIntegrationScenario_EncounterReductionWithRules tests reducing hostile
// encounter frequency based on Guardian protection.
func TestIntegrationScenario_EncounterReductionWithRules(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(5)
	gameState.SetSAN(5)

	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive)
	require.Greater(t, modifiers.EncounterFrequencyReduction, 0.0)

	// Simulate multiple encounter rolls
	totalEncounters := 100
	reducedCount := 0

	for i := 0; i < totalEncounters; i++ {
		encounterChance := float64(i) / float64(totalEncounters)
		if tuner.ShouldReduceEncounter(encounterChance) {
			reducedCount++
		}
	}

	// With ~50% reduction at critical health, approximately half should be reduced
	expectedReduction := int(float64(totalEncounters) * modifiers.EncounterFrequencyReduction)
	tolerance := int(float64(totalEncounters) * 0.1) // 10% tolerance

	assert.InDelta(t, expectedReduction, reducedCount, float64(tolerance),
		"Encounter reduction should be approximately %.0f%% (expected ~%d, got %d)",
		modifiers.EncounterFrequencyReduction*100, expectedReduction, reducedCount)

	t.Logf("Encounter Reduction: %.2f%% -> Reduced %d/%d encounters",
		modifiers.EncounterFrequencyReduction*100, reducedCount, totalEncounters)
}

// TestIntegrationScenario_PunishmentAdjustment tests preemptive punishment adjustment
// when Guardian activates.
func TestIntegrationScenario_PunishmentAdjustment(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)
	ruleEngine := rules.NewRuleEngine()

	gameState := engine.NewGameStateV2()
	gameState.SetHP(15)
	gameState.SetSAN(15)

	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive)

	// Simulate adjusting rule punishments when Guardian activates
	originalPunishment := rules.HiddenRulePunishment{
		HPDamage:     60,
		SANDamage:    40,
		IsFatal:      false,
		CustomEffect: "規則懲罰效果",
	}

	adjustedPunishment := ruleEngine.AdjustPunishmentDamage(originalPunishment, modifiers.DamageReduction)

	assert.Less(t, adjustedPunishment.HPDamage, originalPunishment.HPDamage)
	assert.Less(t, adjustedPunishment.SANDamage, originalPunishment.SANDamage)
	assert.Equal(t, originalPunishment.CustomEffect, adjustedPunishment.CustomEffect,
		"Custom effect should be preserved")

	t.Logf("Punishment Adjustment: HP %d->%d, SAN %d->%d",
		originalPunishment.HPDamage, adjustedPunishment.HPDamage,
		originalPunishment.SANDamage, adjustedPunishment.SANDamage)
}

// TestIntegrationScenario_ResetProtection tests resetting protection and modifiers
// when player recovers.
func TestIntegrationScenario_ResetProtection(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)
	ruleEngine := rules.NewRuleEngine()

	gameState := engine.NewGameStateV2()

	// Player is in critical state
	gameState.SetHP(0)
	gameState.SetSAN(0)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	require.True(t, guardian.ShouldActivateProtection(gameState))

	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive)

	violation := rules.RuleViolation{
		RuleID:    "test",
		RuleName:  "測試",
		HPDamage:  100,
		SANDamage: 50,
		IsFatal:   false,
	}

	adjustedBefore := ruleEngine.ApplyDifficultyAdjustment(violation, modifiers.DamageReduction)
	assert.Less(t, adjustedBefore.HPDamage, violation.HPDamage, "Should have protection")

	// Player recovers
	guardian.ResetProtectionState()
	tuner.ResetModifiers()

	gameState.SetHP(100)
	gameState.SetSAN(100)
	guardian.OnTurnEnd(gameState)

	require.False(t, guardian.ShouldActivateProtection(gameState))

	modifiersAfter := tuner.AdjustDifficulty(gameState)
	require.False(t, modifiersAfter.IsActive)

	adjustedAfter := ruleEngine.ApplyDifficultyAdjustment(violation, modifiersAfter.DamageReduction)
	assert.Equal(t, violation.HPDamage, adjustedAfter.HPDamage, "Should have no protection")
	assert.Equal(t, violation.SANDamage, adjustedAfter.SANDamage, "Should have no protection")

	t.Log("Protection successfully reset after player recovery")
}

// TestIntegrationScenario_FatalRuleWithGuardian tests that fatal rules remain
// fatal even with Guardian protection.
func TestIntegrationScenario_FatalRuleWithGuardian(t *testing.T) {
	cfg := DefaultGuardianConfig()
	cfg.EnableDifficultyTuning = true
	guardian := NewExperienceGuardian(cfg)
	tuner := NewDifficultyTuner(guardian)
	ruleEngine := rules.NewRuleEngine()

	gameState := engine.NewGameStateV2()
	gameState.SetHP(0)
	gameState.SetSAN(0)

	modifiers := tuner.AdjustDifficulty(gameState)
	require.True(t, modifiers.IsActive)

	fatalViolation := rules.RuleViolation{
		RuleID:    "fatal-rule",
		RuleName:  "致命規則",
		HPDamage:  100,
		SANDamage: 100,
		IsFatal:   true,
	}

	adjustedViolation := ruleEngine.ApplyDifficultyAdjustment(fatalViolation, modifiers.DamageReduction)

	// Damage is reduced (gives player a fighting chance)
	assert.Less(t, adjustedViolation.HPDamage, fatalViolation.HPDamage,
		"Damage should be reduced even for fatal rules")

	// But fatal flag is preserved (maintains stakes)
	assert.True(t, adjustedViolation.IsFatal,
		"Fatal rules should remain fatal even with Guardian protection")

	t.Logf("Fatal Rule: HP %d->%d, SAN %d->%d, Fatal: %v->%v",
		fatalViolation.HPDamage, adjustedViolation.HPDamage,
		fatalViolation.SANDamage, adjustedViolation.SANDamage,
		fatalViolation.IsFatal, adjustedViolation.IsFatal)
}
