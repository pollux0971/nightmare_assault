package guardian

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGuardianConfig(t *testing.T) {
	cfg := DefaultGuardianConfig()

	assert.Equal(t, 2, cfg.MaxConsecutiveDeaths, "Default MaxConsecutiveDeaths should be 2")
	assert.Equal(t, 3, cfg.LowStatStreakLimit, "Default LowStatStreakLimit should be 3")
	assert.Equal(t, 20, cfg.LowHPThreshold, "Default LowHPThreshold should be 20")
	assert.Equal(t, 30, cfg.LowSANThreshold, "Default LowSANThreshold should be 30")
}

func TestNewExperienceGuardian(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths: 3,
		LowStatStreakLimit:   4,
		LowHPThreshold:       15,
		LowSANThreshold:      25,
	}

	guardian := NewExperienceGuardian(cfg)

	require.NotNil(t, guardian, "NewExperienceGuardian should not return nil")
	assert.Equal(t, cfg, guardian.config, "Config should match")
	assert.Equal(t, 0, guardian.consecutiveDeaths, "Initial consecutiveDeaths should be 0")
	assert.Equal(t, 0, guardian.lowHPSANStreak, "Initial lowHPSANStreak should be 0")
	assert.False(t, guardian.lastDead, "Initial lastDead should be false")
	assert.False(t, guardian.protectionActive, "Initial protectionActive should be false")
}

func TestOnTurnEnd_NilGameState(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())

	// Should not panic with nil gameState
	assert.NotPanics(t, func() {
		guardian.OnTurnEnd(nil)
	}, "OnTurnEnd should handle nil gameState gracefully")
}

func TestOnTurnEnd_FirstDeath(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()
	gameState.SetHP(0) // Player is dead

	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 1, guardian.consecutiveDeaths, "First death should set consecutiveDeaths to 1")
	assert.True(t, guardian.lastDead, "lastDead should be true")
}

func TestOnTurnEnd_ConsecutiveDeaths(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// First death
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.consecutiveDeaths, "First death")

	// Second consecutive death (player still dead)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.consecutiveDeaths, "Second consecutive death")

	// Third consecutive death
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 3, guardian.consecutiveDeaths, "Third consecutive death")
}

func TestOnTurnEnd_RecoveryFromDeath(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Player dies
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.consecutiveDeaths, "First death")

	// Player recovers
	gameState.SetHP(50)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.consecutiveDeaths, "Death counter should reset after recovery")
	assert.False(t, guardian.lastDead, "lastDead should be false")
}

func TestOnTurnEnd_LowHPDetection(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Set HP to low value (≤20)
	gameState.SetHP(15)
	gameState.SetSAN(100)

	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 1, guardian.lowHPSANStreak, "Low HP should increment streak")
}

func TestOnTurnEnd_LowSANDetection(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Set SAN to low value (≤30)
	gameState.SetHP(100)
	gameState.SetSAN(25)

	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 1, guardian.lowHPSANStreak, "Low SAN should increment streak")
}

func TestOnTurnEnd_LowHPAndSAN(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Both HP and SAN are low
	gameState.SetHP(10)
	gameState.SetSAN(20)

	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 1, guardian.lowHPSANStreak, "Low HP and SAN should increment streak")
}

func TestOnTurnEnd_LowHPSANStreak(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Turn 1: Low HP
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.lowHPSANStreak, "Streak = 1")

	// Turn 2: Still low HP
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.lowHPSANStreak, "Streak = 2")

	// Turn 3: Still low HP
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 3, guardian.lowHPSANStreak, "Streak = 3")
}

func TestOnTurnEnd_RecoveryFromLowStats(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Build up streak
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.lowHPSANStreak, "Streak = 2")

	// Recover HP
	gameState.SetHP(50)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.lowHPSANStreak, "Streak should reset after recovery")
}

func TestOnTurnEnd_EdgeCaseLowHPThreshold(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Exactly at threshold (20)
	gameState.SetHP(20)
	gameState.SetSAN(100)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.lowHPSANStreak, "HP=20 should trigger (≤20)")

	// Just above threshold
	guardian.lowHPSANStreak = 0
	gameState.SetHP(21)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.lowHPSANStreak, "HP=21 should not trigger")
}

func TestOnTurnEnd_EdgeCaseLowSANThreshold(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Exactly at threshold (30)
	gameState.SetHP(100)
	gameState.SetSAN(30)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.lowHPSANStreak, "SAN=30 should trigger (≤30)")

	// Just above threshold
	guardian.lowHPSANStreak = 0
	gameState.SetSAN(31)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.lowHPSANStreak, "SAN=31 should not trigger")
}

func TestShouldActivateProtection_NilGameState(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())

	shouldActivate := guardian.ShouldActivateProtection(nil)

	assert.False(t, shouldActivate, "Should not activate protection with nil gameState")
	assert.False(t, guardian.protectionActive, "Protection should not be active")
}

func TestShouldActivateProtection_NoConditions(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	shouldActivate := guardian.ShouldActivateProtection(gameState)

	assert.False(t, shouldActivate, "Should not activate protection when no conditions are met")
	assert.False(t, guardian.protectionActive, "Protection should not be active")
}

func TestShouldActivateProtection_ConsecutiveDeaths(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Simulate consecutive deaths
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 2, guardian.consecutiveDeaths, "Should have 2 consecutive deaths")

	shouldActivate := guardian.ShouldActivateProtection(gameState)

	assert.True(t, shouldActivate, "Should activate protection after 2 consecutive deaths")
	assert.True(t, guardian.protectionActive, "Protection should be active")
}

func TestShouldActivateProtection_ConsecutiveDeathsCustomThreshold(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths: 3,
		LowStatStreakLimit:   5,
		LowHPThreshold:       15,
		LowSANThreshold:      25,
	}
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// 2 deaths - should not trigger with threshold of 3
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 2, guardian.consecutiveDeaths, "Should have 2 deaths")
	assert.False(t, guardian.ShouldActivateProtection(gameState), "Should not activate with only 2 deaths")

	// 3rd death - should trigger
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 3, guardian.consecutiveDeaths, "Should have 3 deaths")
	assert.True(t, guardian.ShouldActivateProtection(gameState), "Should activate after 3 deaths")
}

func TestShouldActivateProtection_LowHPSANStreak(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Simulate 3 turns of low HP
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 3, guardian.lowHPSANStreak, "Should have 3 turn streak")

	shouldActivate := guardian.ShouldActivateProtection(gameState)

	assert.True(t, shouldActivate, "Should activate protection after 3 turns of low HP")
	assert.True(t, guardian.protectionActive, "Protection should be active")
}

func TestShouldActivateProtection_LowHPSANStreakCustomThreshold(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths: 2,
		LowStatStreakLimit:   4,
		LowHPThreshold:       20,
		LowSANThreshold:      30,
	}
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// 3 turns - should not trigger with threshold of 4
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 3, guardian.lowHPSANStreak, "Should have 3 turn streak")
	assert.False(t, guardian.ShouldActivateProtection(gameState), "Should not activate with only 3 turns")

	// 4th turn - should trigger
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 4, guardian.lowHPSANStreak, "Should have 4 turn streak")
	assert.True(t, guardian.ShouldActivateProtection(gameState), "Should activate after 4 turns")
}

func TestShouldActivateProtection_BothConditions(t *testing.T) {
	cfg := DefaultGuardianConfig()
	guardian := NewExperienceGuardian(cfg)
	gameState := engine.NewGameStateV2()

	// Simulate both consecutive deaths and low HP streak
	gameState.SetHP(0) // Dead with low HP
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	// Should trigger on consecutive deaths (first condition checked)
	shouldActivate := guardian.ShouldActivateProtection(gameState)

	assert.True(t, shouldActivate, "Should activate when multiple conditions are met")
	assert.True(t, guardian.protectionActive, "Protection should be active")
}

func TestResetProtectionState(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Build up state
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.ShouldActivateProtection(gameState)

	assert.Equal(t, 2, guardian.consecutiveDeaths, "Should have deaths")
	assert.True(t, guardian.protectionActive, "Protection should be active")

	// Reset
	guardian.ResetProtectionState()

	assert.Equal(t, 0, guardian.consecutiveDeaths, "Deaths should be reset")
	assert.Equal(t, 0, guardian.lowHPSANStreak, "Streak should be reset")
	assert.False(t, guardian.lastDead, "lastDead should be reset")
	assert.False(t, guardian.protectionActive, "Protection should be deactivated")
}

func TestResetProtectionState_WithLowHPStreak(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Build up low HP streak
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.ShouldActivateProtection(gameState)

	assert.Equal(t, 3, guardian.lowHPSANStreak, "Should have streak")
	assert.True(t, guardian.protectionActive, "Protection should be active")

	// Reset
	guardian.ResetProtectionState()

	assert.Equal(t, 0, guardian.lowHPSANStreak, "Streak should be reset")
	assert.False(t, guardian.protectionActive, "Protection should be deactivated")
}

func TestIsProtectionActive(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Initially not active
	assert.False(t, guardian.IsProtectionActive(), "Initially protection should not be active")

	// Trigger protection
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.ShouldActivateProtection(gameState)

	assert.True(t, guardian.IsProtectionActive(), "Protection should be active after trigger")

	// Reset
	guardian.ResetProtectionState()
	assert.False(t, guardian.IsProtectionActive(), "Protection should not be active after reset")
}

func TestGetConsecutiveDeaths(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	assert.Equal(t, 0, guardian.GetConsecutiveDeaths(), "Initially 0 deaths")

	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetConsecutiveDeaths(), "Should return 1 death")

	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.GetConsecutiveDeaths(), "Should return 2 deaths")
}

func TestGetLowHPSANStreak(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	assert.Equal(t, 0, guardian.GetLowHPSANStreak(), "Initially 0 streak")

	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetLowHPSANStreak(), "Should return 1 turn streak")

	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.GetLowHPSANStreak(), "Should return 2 turn streak")
}

func TestGetConfig(t *testing.T) {
	cfg := GuardianConfig{
		MaxConsecutiveDeaths: 5,
		LowStatStreakLimit:   6,
		LowHPThreshold:       10,
		LowSANThreshold:      15,
	}
	guardian := NewExperienceGuardian(cfg)

	returnedCfg := guardian.GetConfig()

	assert.Equal(t, cfg, returnedCfg, "Should return the same config")
}

func TestIntegrationScenario_ConsecutiveDeaths(t *testing.T) {
	// Integration test: Simulate a scenario with consecutive deaths
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Player is healthy
	gameState.SetHP(100)
	gameState.SetSAN(100)
	guardian.OnTurnEnd(gameState)
	assert.False(t, guardian.ShouldActivateProtection(gameState), "No protection needed")

	// Player dies (first time)
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetConsecutiveDeaths())
	assert.False(t, guardian.ShouldActivateProtection(gameState), "One death not enough")

	// Player still dead (second consecutive death)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.GetConsecutiveDeaths())
	assert.True(t, guardian.ShouldActivateProtection(gameState), "Protection should trigger")
	assert.True(t, guardian.IsProtectionActive())

	// Apply protection and reset
	guardian.ResetProtectionState()
	assert.Equal(t, 0, guardian.GetConsecutiveDeaths())
	assert.False(t, guardian.IsProtectionActive())
}

func TestIntegrationScenario_LowHPStreak(t *testing.T) {
	// Integration test: Simulate a scenario with low HP streak
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Turn 1: Low HP
	gameState.SetHP(18)
	gameState.SetSAN(100)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetLowHPSANStreak())
	assert.False(t, guardian.ShouldActivateProtection(gameState))

	// Turn 2: Still low HP
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.GetLowHPSANStreak())
	assert.False(t, guardian.ShouldActivateProtection(gameState))

	// Turn 3: Still low HP - should trigger
	gameState.SetHP(10)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 3, guardian.GetLowHPSANStreak())
	assert.True(t, guardian.ShouldActivateProtection(gameState), "Protection should trigger after 3 turns")

	// Apply protection and reset
	guardian.ResetProtectionState()
	assert.Equal(t, 0, guardian.GetLowHPSANStreak())
}

func TestIntegrationScenario_LowSANStreak(t *testing.T) {
	// Integration test: Simulate a scenario with low SAN streak
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Turn 1-3: Low SAN
	gameState.SetHP(100)
	gameState.SetSAN(25)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)

	assert.Equal(t, 3, guardian.GetLowHPSANStreak())
	assert.True(t, guardian.ShouldActivateProtection(gameState), "Protection should trigger")
}

func TestIntegrationScenario_Recovery(t *testing.T) {
	// Integration test: Player builds streak then recovers
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Build up low HP streak
	gameState.SetHP(15)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 2, guardian.GetLowHPSANStreak())

	// Player recovers
	gameState.SetHP(80)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.GetLowHPSANStreak(), "Streak should reset after recovery")
	assert.False(t, guardian.ShouldActivateProtection(gameState), "No protection needed")

	// Build streak again
	gameState.SetHP(10)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 3, guardian.GetLowHPSANStreak())
	assert.True(t, guardian.ShouldActivateProtection(gameState))
}

func TestIntegrationScenario_MixedConditions(t *testing.T) {
	// Integration test: Mix of death, recovery, and low stats
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Player alive with low HP
	gameState.SetHP(18)
	gameState.SetSAN(100)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetLowHPSANStreak())
	assert.Equal(t, 0, guardian.GetConsecutiveDeaths())

	// Player dies
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 1, guardian.GetConsecutiveDeaths())
	// Low HP streak continues since dead player has HP=0 which is ≤20
	assert.Equal(t, 2, guardian.GetLowHPSANStreak())

	// Player recovers
	gameState.SetHP(100)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, 0, guardian.GetConsecutiveDeaths(), "Death counter resets")
	assert.Equal(t, 0, guardian.GetLowHPSANStreak(), "HP streak resets")
	assert.False(t, guardian.ShouldActivateProtection(gameState))
}
