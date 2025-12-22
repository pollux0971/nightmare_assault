package guardian

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTensionManager_GuardianIntegration tests integration between Guardian and TensionManager
func TestTensionManager_GuardianIntegration(t *testing.T) {
	// Setup
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.MaxConsecutiveDeaths = 2
	guardianConfig.LowHPThreshold = 20
	guardianConfig.LowSANThreshold = 30

	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tensionManager := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())
	guardian := NewExperienceGuardian(guardianConfig)

	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(70) // Start at high tension

	// Scenario: Player struggles and dies multiple times
	// Turn 1: Player in danger
	gameState.SetHP(10)
	gameState.SetSAN(25)
	guardian.OnTurnEnd(gameState)
	tensionManager.OnTurnEnd(gameState)

	// Both should track low HP/SAN
	assert.Equal(t, guardian.GetLowHPSANStreak(), tensionManager.GetLowHPSANStreak())

	// Turn 2: Player dies
	gameState.SetHP(0)
	guardian.OnTurnEnd(gameState)
	tensionManager.OnTurnEnd(gameState)

	// Both should track death
	assert.Equal(t, guardian.GetConsecutiveDeaths(), tensionManager.GetConsecutiveDeaths())

	// Turn 3: Player still dead
	guardian.OnTurnEnd(gameState)
	tensionManager.OnTurnEnd(gameState)

	// Both should have 2 consecutive deaths
	assert.Equal(t, 2, guardian.GetConsecutiveDeaths())
	assert.Equal(t, 2, tensionManager.GetConsecutiveDeaths())

	// Guardian should activate protection
	assert.True(t, guardian.ShouldActivateProtection(gameState))

	// TensionManager should have reduced tension
	assert.Less(t, gameState.Tension.GetValue(), 70)

	// Verify tension reduction was recorded
	history := tensionManager.GetAdjustmentHistory()
	require.NotEmpty(t, history)
	assert.Equal(t, "consecutive_deaths", history[len(history)-1].Reason)
}

// TestTensionManager_MomentumControllerWorkflow tests complete workflow with momentum
func TestTensionManager_MomentumControllerWorkflow(t *testing.T) {
	// Setup
	guardianConfig := DefaultGuardianConfig()
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)
	tensionManager := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())

	gameState := engine.NewGameStateV2()

	// Scenario 1: High tension -> High frequency gameplay
	gameState.Tension.SetValue(80)
	tensionManager.IntegrateWithMomentumController(gameState)

	config := momentumController.GetConfig()
	assert.Equal(t, momentum.FrequencyHigh, config.Frequency)
	assert.Equal(t, momentum.RiskLow, config.PauseOnRisk)

	// Create narrative context for momentum check
	ctx := &momentum.NarrativeContext{
		CurrentBeat:  1,
		CurrentScene: "test_scene",
		RiskLevel:    momentum.RiskMedium,
	}

	// With high frequency, should pause even on medium risk
	shouldPause := momentumController.ShouldPauseForChoice(ctx)
	assert.True(t, shouldPause, "Should pause in high frequency mode")

	// Scenario 2: Player struggles, tension reduces significantly, frequency changes
	// Manually reduce tension to force phase change
	gameState.Tension.SetValue(40) // Move to Buildup phase

	// Update momentum controller based on new tension
	tensionManager.IntegrateWithMomentumController(gameState)

	// Should now have medium frequency
	config = momentumController.GetConfig()
	assert.Equal(t, momentum.FrequencyMedium, config.Frequency, "Should change to medium frequency in Buildup phase")
	assert.Equal(t, momentum.RiskMedium, config.PauseOnRisk, "Should change to medium risk in Buildup phase")

	// Scenario 3: Tension drops further
	gameState.Tension.SetValue(15) // Move to Rest phase
	tensionManager.IntegrateWithMomentumController(gameState)

	// Should now have low frequency (most auto-play)
	config = momentumController.GetConfig()
	assert.Equal(t, momentum.FrequencyLow, config.Frequency, "Should change to low frequency in Rest phase")
}

// TestTensionManager_RealisticGameplay tests realistic gameplay scenario
func TestTensionManager_RealisticGameplay(t *testing.T) {
	// Setup with realistic config
	guardianConfig := GuardianConfig{
		MaxConsecutiveDeaths: 2,
		LowStatStreakLimit:   3,
		LowHPThreshold:       25,
		LowSANThreshold:      30,
		EnableDifficultyTuning: true,
	}

	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tensionManager := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())

	gameState := engine.NewGameStateV2()
	gameState.SetHP(100)
	gameState.SetSAN(100)
	gameState.Tension.SetValue(50) // Start at medium tension

	// Simulate 10 turns of gameplay
	scenarios := []struct {
		turn        int
		hp          int
		san         int
		description string
	}{
		{1, 100, 100, "Start - healthy"},
		{2, 85, 90, "Minor damage"},
		{3, 70, 80, "Taking more damage"},
		{4, 55, 65, "Struggling"},
		{5, 20, 40, "Critical - low HP and SAN"},
		{6, 15, 35, "Still critical"},
		{7, 10, 30, "Barely surviving"},
		{8, 50, 60, "Recovered somewhat"},
		{9, 80, 85, "Much better"},
		{10, 100, 100, "Full recovery"},
	}

	initialTension := gameState.Tension.GetValue()

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			gameState.SetHP(scenario.hp)
			gameState.SetSAN(scenario.san)

			tensionManager.OnTurnEnd(gameState)
			tensionManager.IntegrateWithMomentumController(gameState)

			t.Logf("Turn %d: HP=%d, SAN=%d, Tension=%d, Phase=%s",
				scenario.turn,
				scenario.hp,
				scenario.san,
				gameState.Tension.GetValue(),
				tensionManager.GetCurrentPhase(gameState).String())
		})
	}

	// By the end, tension should be lower due to critical turns 5-7
	finalTension := gameState.Tension.GetValue()
	assert.Less(t, finalTension, initialTension, "Tension should decrease after critical phase")

	// Verify adjustments were made
	history := tensionManager.GetAdjustmentHistory()
	assert.NotEmpty(t, history, "Should have tension adjustments")

	// Verify low HP/SAN was tracked
	assert.Equal(t, 0, tensionManager.GetLowHPSANStreak(), "Should be reset after recovery")
}

// TestTensionManager_EdgeCaseRecovery tests recovery from extreme states
func TestTensionManager_EdgeCaseRecovery(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tensionManager := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())

	gameState := engine.NewGameStateV2()

	// Extreme case: Multiple deaths in a row
	gameState.Tension.SetValue(90) // Very high tension
	initialTension := gameState.Tension.GetValue()

	// Die 5 times
	gameState.SetHP(0)
	for i := 0; i < 5; i++ {
		tensionManager.OnTurnEnd(gameState)
	}

	// Tension should have been reduced significantly
	tensionAfterDeaths := gameState.Tension.GetValue()
	assert.Less(t, tensionAfterDeaths, initialTension, "Tension should decrease after deaths")

	// Verify minimum threshold was respected
	assert.GreaterOrEqual(t, tensionAfterDeaths, 10, "Should not go below minimum threshold")

	// Recover
	gameState.SetHP(100)
	tensionManager.OnTurnEnd(gameState)

	// Death counter should reset
	assert.Equal(t, 0, tensionManager.GetConsecutiveDeaths())

	// Tension should stabilize
	tensionAfterRecovery := gameState.Tension.GetValue()
	assert.Equal(t, tensionAfterDeaths, tensionAfterRecovery, "Tension should stabilize after recovery")
}

// TestTensionManager_ConcurrentPhaseUpdates tests phase-based tension management
func TestTensionManager_ConcurrentPhaseUpdates(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.MaxConsecutiveDeaths = 1 // Trigger quickly for testing
	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tensionManager := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())

	gameState := engine.NewGameStateV2()

	// Test reduction amounts vary by phase
	testCases := []struct {
		name            string
		initialTension  int
		expectedPhase   TensionPhase
		triggerDeath    bool
		minReduction    int // Minimum expected reduction
	}{
		{"Rest phase death", 20, PhaseRest, true, 5},
		{"Buildup phase death", 40, PhaseBuildup, true, 10},
		{"Peak phase death", 70, PhasePeak, true, 12},
		{"Release phase death", 95, PhaseRelease, true, 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameState.Tension.SetValue(tc.initialTension)
			assert.Equal(t, tc.expectedPhase, tensionManager.GetCurrentPhase(gameState))

			if tc.triggerDeath {
				gameState.SetHP(0)
				tensionManager.OnTurnEnd(gameState) // First death
				tensionManager.OnTurnEnd(gameState) // Second death to trigger reduction

				newTension := gameState.Tension.GetValue()
				reduction := tc.initialTension - newTension

				assert.GreaterOrEqual(t, reduction, tc.minReduction,
					"Should reduce tension by at least %d in %s phase", tc.minReduction, tc.expectedPhase.String())

				// Reset for next test
				tensionManager.ResetProtectionState()
				gameState.SetHP(100)
			}
		})
	}
}
