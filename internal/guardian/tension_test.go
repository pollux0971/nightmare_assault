package guardian

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTensionManager tests TensionManager creation
func TestNewTensionManager(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)
	tensionConfig := DefaultTensionManagerConfig()

	tm := NewTensionManager(guardianConfig, momentumController, tensionConfig)

	require.NotNil(t, tm, "NewTensionManager should not return nil")
	assert.Equal(t, guardianConfig, tm.config, "GuardianConfig should be set")
	assert.Equal(t, momentumController, tm.momentumController, "MomentumController should be set")
	assert.Equal(t, tensionConfig.MinTensionThreshold, tm.minTensionThreshold, "MinTensionThreshold should be set")
	assert.Equal(t, 0, tm.consecutiveDeaths, "Initial consecutive deaths should be 0")
	assert.Equal(t, 0, tm.lowHPSANStreak, "Initial low HP/SAN streak should be 0")
	assert.NotNil(t, tm.adjustmentHistory, "Adjustment history should be initialized")
	assert.Len(t, tm.adjustmentHistory, 0, "Adjustment history should be empty initially")
}

// TestDefaultTensionManagerConfig tests default configuration
func TestDefaultTensionManagerConfig(t *testing.T) {
	config := DefaultTensionManagerConfig()

	assert.Equal(t, 10, config.MinTensionThreshold, "Default MinTensionThreshold should be 10")
}

// TestTensionManager_GetCurrentPhase tests getting current tension phase
func TestTensionManager_GetCurrentPhase(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	tests := []struct {
		name     string
		tension  int
		expected TensionPhase
	}{
		{"Rest phase", 15, PhaseRest},
		{"Buildup phase", 40, PhaseBuildup},
		{"Peak phase", 75, PhasePeak},
		{"Release phase", 95, PhaseRelease},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState.Tension.SetValue(tt.tension)
			phase := tm.GetCurrentPhase(gameState)
			assert.Equal(t, tt.expected, phase)
		})
	}
}

// TestTensionManager_GetCurrentPhase_NilGameState tests handling nil gameState
func TestTensionManager_GetCurrentPhase_NilGameState(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	// Should not panic and return Rest as default
	assert.NotPanics(t, func() {
		phase := tm.GetCurrentPhase(nil)
		assert.Equal(t, PhaseRest, phase, "Should return Rest phase for nil gameState")
	})
}

// TestShouldReduceTension_ConsecutiveDeaths tests reduction trigger on consecutive deaths
func TestShouldReduceTension_ConsecutiveDeaths(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.MaxConsecutiveDeaths = 2
	tm := NewTensionManager(guardianConfig, nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// No consecutive deaths - should not reduce
	shouldReduce, reason := tm.ShouldReduceTension(gameState)
	assert.False(t, shouldReduce)
	assert.Empty(t, reason)

	// 1 consecutive death - should not reduce yet
	tm.consecutiveDeaths = 1
	shouldReduce, reason = tm.ShouldReduceTension(gameState)
	assert.False(t, shouldReduce)

	// 2 consecutive deaths - should reduce
	tm.consecutiveDeaths = 2
	shouldReduce, reason = tm.ShouldReduceTension(gameState)
	assert.True(t, shouldReduce)
	assert.Equal(t, "consecutive_deaths", reason)

	// 3 consecutive deaths - should still reduce
	tm.consecutiveDeaths = 3
	shouldReduce, reason = tm.ShouldReduceTension(gameState)
	assert.True(t, shouldReduce)
	assert.Equal(t, "consecutive_deaths", reason)
}

// TestShouldReduceTension_LowHPSANStreak tests reduction trigger on low HP/SAN
func TestShouldReduceTension_LowHPSANStreak(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.LowStatStreakLimit = 3
	tm := NewTensionManager(guardianConfig, nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// No low HP/SAN streak - should not reduce
	shouldReduce, reason := tm.ShouldReduceTension(gameState)
	assert.False(t, shouldReduce)
	assert.Empty(t, reason)

	// 2 turns of low HP/SAN - should not reduce yet
	tm.lowHPSANStreak = 2
	shouldReduce, reason = tm.ShouldReduceTension(gameState)
	assert.False(t, shouldReduce)

	// 3 turns of low HP/SAN - should reduce
	tm.lowHPSANStreak = 3
	shouldReduce, reason = tm.ShouldReduceTension(gameState)
	assert.True(t, shouldReduce)
	assert.Equal(t, "low_hp_san_streak", reason)
}

// TestShouldReduceTension_NilGameState tests handling nil gameState
func TestShouldReduceTension_NilGameState(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	// Should not panic and return false
	assert.NotPanics(t, func() {
		shouldReduce, reason := tm.ShouldReduceTension(nil)
		assert.False(t, shouldReduce)
		assert.Empty(t, reason)
	})
}

// TestAdjustTension_Success tests successful tension adjustment
func TestAdjustTension_Success(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Reduce tension by 10
	adjusted := tm.AdjustTension(gameState, -10, "test_reduction")

	assert.True(t, adjusted, "Should successfully adjust tension")
	assert.Equal(t, 40, gameState.Tension.GetValue(), "Tension should be reduced")

	// Check adjustment history
	history := tm.GetAdjustmentHistory()
	require.Len(t, history, 1, "Should have 1 adjustment in history")
	assert.Equal(t, -10, history[0].Delta)
	assert.Equal(t, "test_reduction", history[0].Reason)
	assert.Equal(t, 50, history[0].OldValue)
	assert.Equal(t, 40, history[0].NewValue)
}

// TestAdjustTension_MinimumThreshold tests minimum threshold enforcement
func TestAdjustTension_MinimumThreshold(t *testing.T) {
	config := DefaultTensionManagerConfig()
	config.MinTensionThreshold = 10
	tm := NewTensionManager(DefaultGuardianConfig(), nil, config)
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(15)

	// Try to reduce tension below minimum
	adjusted := tm.AdjustTension(gameState, -10, "test_reduction")

	assert.True(t, adjusted, "Should adjust tension")
	assert.Equal(t, 10, gameState.Tension.GetValue(), "Tension should be clamped to minimum threshold")

	// Try to reduce when already at minimum
	adjusted = tm.AdjustTension(gameState, -5, "test_reduction_2")
	assert.False(t, adjusted, "Should not adjust when already at minimum")
	assert.Equal(t, 10, gameState.Tension.GetValue(), "Tension should stay at minimum")
}

// TestAdjustTension_IncreaseTension tests increasing tension
func TestAdjustTension_IncreaseTension(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(30)

	// Increase tension
	adjusted := tm.AdjustTension(gameState, 20, "test_increase")

	assert.True(t, adjusted, "Should successfully adjust tension")
	assert.Equal(t, 50, gameState.Tension.GetValue(), "Tension should be increased")
}

// TestAdjustTension_NilGameState tests handling nil gameState
func TestAdjustTension_NilGameState(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	// Should not panic and return false
	assert.NotPanics(t, func() {
		adjusted := tm.AdjustTension(nil, -10, "test")
		assert.False(t, adjusted, "Should return false for nil gameState")
	})
}

// TestAdjustTension_HistoryLimit tests that history is limited to 50 entries
func TestAdjustTension_HistoryLimit(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Make 60 adjustments
	for i := 0; i < 60; i++ {
		// Alternate between small increases and decreases
		delta := 1
		if i%2 == 0 {
			delta = -1
		}
		tm.AdjustTension(gameState, delta, "test")
	}

	history := tm.GetAdjustmentHistory()
	assert.Len(t, history, 50, "History should be limited to 50 entries")
}

// TestOnTurnEnd_DeathTracking tests death tracking
func TestOnTurnEnd_DeathTracking(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// Initially alive
	assert.Equal(t, 0, tm.GetConsecutiveDeaths())

	// Player dies
	gameState.SetHP(0)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 1, tm.GetConsecutiveDeaths(), "Should track first death")

	// Player still dead
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 2, tm.GetConsecutiveDeaths(), "Should increment consecutive deaths")

	// Player recovers
	gameState.SetHP(50)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 0, tm.GetConsecutiveDeaths(), "Should reset on recovery")
}

// TestOnTurnEnd_LowHPSANTracking tests low HP/SAN tracking
func TestOnTurnEnd_LowHPSANTracking(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.LowHPThreshold = 20
	guardianConfig.LowSANThreshold = 30
	tm := NewTensionManager(guardianConfig, nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// Initially healthy
	gameState.SetHP(50)
	gameState.SetSAN(50)
	assert.Equal(t, 0, tm.GetLowHPSANStreak())

	// Low HP
	gameState.SetHP(15)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 1, tm.GetLowHPSANStreak(), "Should track low HP")

	// Still low HP
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 2, tm.GetLowHPSANStreak(), "Should increment streak")

	// Recover HP but low SAN
	gameState.SetHP(50)
	gameState.SetSAN(25)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 3, tm.GetLowHPSANStreak(), "Should continue streak with low SAN")

	// Full recovery
	gameState.SetSAN(50)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 0, tm.GetLowHPSANStreak(), "Should reset on full recovery")
}

// TestOnTurnEnd_AutomaticTensionReduction tests automatic tension reduction
func TestOnTurnEnd_AutomaticTensionReduction(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.MaxConsecutiveDeaths = 2
	tm := NewTensionManager(guardianConfig, nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Simulate 2 consecutive deaths
	gameState.SetHP(0)
	tm.OnTurnEnd(gameState) // 1st death
	tm.OnTurnEnd(gameState) // 2nd death

	// Tension should be reduced automatically
	assert.Less(t, gameState.Tension.GetValue(), 50, "Tension should be reduced after consecutive deaths")
	assert.Equal(t, 2, tm.GetConsecutiveDeaths())

	// Check adjustment history
	history := tm.GetAdjustmentHistory()
	require.NotEmpty(t, history, "Should have adjustment in history")
	assert.Equal(t, "consecutive_deaths", history[len(history)-1].Reason)
}

// TestTensionManager_OnTurnEnd_NilGameState tests handling nil gameState
func TestTensionManager_OnTurnEnd_NilGameState(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	// Should not panic
	assert.NotPanics(t, func() {
		tm.OnTurnEnd(nil)
	})
}

// TestCalculateTensionReduction tests reduction calculation
func TestCalculateTensionReduction(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	tests := []struct {
		name     string
		phase    TensionPhase
		reason   string
		expected int
	}{
		// Consecutive deaths
		{"Deaths in Rest", PhaseRest, "consecutive_deaths", -7},    // -15 * 0.5
		{"Deaths in Buildup", PhaseBuildup, "consecutive_deaths", -11}, // -15 * 0.75
		{"Deaths in Peak", PhasePeak, "consecutive_deaths", -15},   // -15 * 1.0
		{"Deaths in Release", PhaseRelease, "consecutive_deaths", -18}, // -15 * 1.2

		// Low HP/SAN
		{"Low HP/SAN in Rest", PhaseRest, "low_hp_san_streak", -5},    // -10 * 0.5
		{"Low HP/SAN in Buildup", PhaseBuildup, "low_hp_san_streak", -7}, // -10 * 0.75
		{"Low HP/SAN in Peak", PhasePeak, "low_hp_san_streak", -10},   // -10 * 1.0
		{"Low HP/SAN in Release", PhaseRelease, "low_hp_san_streak", -12}, // -10 * 1.2

		// Unknown reason
		{"Unknown in Peak", PhasePeak, "unknown", -5}, // -5 * 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reduction := tm.calculateTensionReduction(tt.phase, tt.reason)
			assert.Equal(t, tt.expected, reduction)
		})
	}
}

// TestGetAdjustmentHistory tests getting adjustment history
func TestGetAdjustmentHistory(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Make some adjustments
	tm.AdjustTension(gameState, -5, "test1")
	tm.AdjustTension(gameState, -10, "test2")
	tm.AdjustTension(gameState, 15, "test3")

	history := tm.GetAdjustmentHistory()
	require.Len(t, history, 3, "Should have 3 adjustments")
	assert.Equal(t, "test1", history[0].Reason)
	assert.Equal(t, "test2", history[1].Reason)
	assert.Equal(t, "test3", history[2].Reason)

	// Verify it's a copy (modifying returned slice shouldn't affect internal state)
	history[0].Reason = "modified"
	newHistory := tm.GetAdjustmentHistory()
	assert.Equal(t, "test1", newHistory[0].Reason, "Internal history should not be modified")
}

// TestTensionManager_GetConsecutiveDeaths tests getting consecutive deaths
func TestTensionManager_GetConsecutiveDeaths(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	assert.Equal(t, 0, tm.GetConsecutiveDeaths(), "Initial value should be 0")

	tm.consecutiveDeaths = 3
	assert.Equal(t, 3, tm.GetConsecutiveDeaths(), "Should return current value")
}

// TestTensionManager_GetLowHPSANStreak tests getting low HP/SAN streak
func TestTensionManager_GetLowHPSANStreak(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	assert.Equal(t, 0, tm.GetLowHPSANStreak(), "Initial value should be 0")

	tm.lowHPSANStreak = 5
	assert.Equal(t, 5, tm.GetLowHPSANStreak(), "Should return current value")
}

// TestTensionManager_ResetProtectionState tests resetting state
func TestTensionManager_ResetProtectionState(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())

	// Set some values
	tm.consecutiveDeaths = 3
	tm.lowHPSANStreak = 5

	tm.ResetProtectionState()

	assert.Equal(t, 0, tm.GetConsecutiveDeaths(), "Consecutive deaths should be reset")
	assert.Equal(t, 0, tm.GetLowHPSANStreak(), "Low HP/SAN streak should be reset")
}

// TestIntegrateWithMomentumController tests MomentumController integration
func TestIntegrateWithMomentumController(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)
	tm := NewTensionManager(guardianConfig, momentumController, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	tests := []struct {
		name                 string
		tension              int
		expectedFrequency    momentum.FrequencyLevel
		expectedPauseOnRisk  momentum.RiskLevel
	}{
		{
			name:                "Rest phase",
			tension:             15,
			expectedFrequency:   momentum.FrequencyLow,
			expectedPauseOnRisk: momentum.RiskHigh,
		},
		{
			name:                "Buildup phase",
			tension:             40,
			expectedFrequency:   momentum.FrequencyMedium,
			expectedPauseOnRisk: momentum.RiskMedium,
		},
		{
			name:                "Peak phase",
			tension:             75,
			expectedFrequency:   momentum.FrequencyHigh,
			expectedPauseOnRisk: momentum.RiskLow,
		},
		{
			name:                "Release phase",
			tension:             95,
			expectedFrequency:   momentum.FrequencyHigh,
			expectedPauseOnRisk: momentum.RiskLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState.Tension.SetValue(tt.tension)
			tm.IntegrateWithMomentumController(gameState)

			config := momentumController.GetConfig()
			assert.Equal(t, tt.expectedFrequency, config.Frequency, "Frequency should match phase")
			assert.Equal(t, tt.expectedPauseOnRisk, config.PauseOnRisk, "PauseOnRisk should match phase")
		})
	}
}

// TestIntegrateWithMomentumController_NilController tests handling nil controller
func TestIntegrateWithMomentumController_NilController(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// Should not panic
	assert.NotPanics(t, func() {
		tm.IntegrateWithMomentumController(gameState)
	})
}

// TestIntegrateWithMomentumController_NilGameState tests handling nil gameState
func TestIntegrateWithMomentumController_NilGameState(t *testing.T) {
	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tm := NewTensionManager(DefaultGuardianConfig(), momentumController, DefaultTensionManagerConfig())

	// Should not panic
	assert.NotPanics(t, func() {
		tm.IntegrateWithMomentumController(nil)
	})
}

// TestIntegrationScenario_FullCycle tests a complete gameplay cycle
func TestIntegrationScenario_FullCycle(t *testing.T) {
	guardianConfig := DefaultGuardianConfig()
	guardianConfig.MaxConsecutiveDeaths = 2
	guardianConfig.LowHPThreshold = 20
	guardianConfig.LowSANThreshold = 30

	momentumConfig := momentum.DefaultMomentumConfig()
	momentumController := momentum.NewMomentumController(momentumConfig, nil)

	tensionConfig := DefaultTensionManagerConfig()
	tm := NewTensionManager(guardianConfig, momentumController, tensionConfig)

	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(60) // Start at Peak phase

	// Turn 1: Player dies (HP=0 also counts as low HP)
	gameState.SetHP(0)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 1, tm.GetConsecutiveDeaths())
	// Note: HP=0 also triggers low HP streak
	lowHPStreakAfterDeath := tm.GetLowHPSANStreak()
	// Tension should not change yet (need 2 deaths)
	assert.Equal(t, 60, gameState.Tension.GetValue())

	// Turn 2: Player still dead (2nd consecutive death)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 2, tm.GetConsecutiveDeaths())
	// Tension should be reduced
	assert.Less(t, gameState.Tension.GetValue(), 60)

	// Turn 3: Player recovers but with low HP
	recoveredTension := gameState.Tension.GetValue()
	gameState.SetHP(15)
	gameState.SetSAN(50)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 0, tm.GetConsecutiveDeaths(), "Deaths reset on recovery")
	// Low HP streak continues from death phase
	lowHPStreakAfterRecovery := tm.GetLowHPSANStreak()
	assert.Greater(t, lowHPStreakAfterRecovery, lowHPStreakAfterDeath, "Low HP streak should continue")
	// Tension adjustment may have happened if streak >= 3
	if lowHPStreakAfterRecovery < 3 {
		assert.Equal(t, recoveredTension, gameState.Tension.GetValue(), "Tension shouldn't change yet (need 3 turns)")
	}

	// Turn 4-5: Continue with low HP
	tm.OnTurnEnd(gameState)
	tm.OnTurnEnd(gameState)
	finalLowHPStreak := tm.GetLowHPSANStreak()
	assert.GreaterOrEqual(t, finalLowHPStreak, 3, "Should have at least 3 turns of low HP")
	// Tension should be reduced again
	assert.Less(t, gameState.Tension.GetValue(), recoveredTension)

	// Turn 6: Full recovery
	tensionBeforeRecovery := gameState.Tension.GetValue()
	gameState.SetHP(80)
	tm.OnTurnEnd(gameState)
	assert.Equal(t, 0, tm.GetLowHPSANStreak(), "Streak reset on recovery")
	// Tension should remain stable (no further adjustments)
	assert.Equal(t, tensionBeforeRecovery, gameState.Tension.GetValue())

	// Verify adjustment history
	history := tm.GetAdjustmentHistory()
	assert.Greater(t, len(history), 0, "Should have adjustments recorded")

	// Verify at least one adjustment for each trigger type
	hasDeathAdjustment := false
	hasLowStatAdjustment := false
	for _, adj := range history {
		if adj.Reason == "consecutive_deaths" {
			hasDeathAdjustment = true
		}
		if adj.Reason == "low_hp_san_streak" {
			hasLowStatAdjustment = true
		}
	}
	assert.True(t, hasDeathAdjustment, "Should have death-triggered adjustment")
	assert.True(t, hasLowStatAdjustment, "Should have low stat-triggered adjustment")
}

// TestIntegrationScenario_MomentumSync tests momentum controller synchronization
func TestIntegrationScenario_MomentumSync(t *testing.T) {
	momentumController := momentum.NewMomentumController(momentum.DefaultMomentumConfig(), nil)
	tm := NewTensionManager(DefaultGuardianConfig(), momentumController, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()

	// Start at high tension (Peak)
	gameState.Tension.SetValue(70)
	tm.IntegrateWithMomentumController(gameState)
	config := momentumController.GetConfig()
	freq1 := config.Frequency
	risk1 := config.PauseOnRisk
	assert.Equal(t, momentum.FrequencyHigh, freq1)
	assert.Equal(t, momentum.RiskLow, risk1)

	// Reduce to medium tension (Buildup)
	gameState.Tension.SetValue(40)
	tm.IntegrateWithMomentumController(gameState)
	config = momentumController.GetConfig()
	freq2 := config.Frequency
	risk2 := config.PauseOnRisk
	assert.Equal(t, momentum.FrequencyMedium, freq2)
	assert.Equal(t, momentum.RiskMedium, risk2)

	// Reduce to low tension (Rest)
	gameState.Tension.SetValue(15)
	tm.IntegrateWithMomentumController(gameState)
	config = momentumController.GetConfig()
	freq3 := config.Frequency
	risk3 := config.PauseOnRisk
	assert.Equal(t, momentum.FrequencyLow, freq3)
	assert.Equal(t, momentum.RiskHigh, risk3)

	// Verify momentum controller is actually updated across phases
	assert.NotEqual(t, freq1, freq3, "Frequency should change from Peak to Rest")
	assert.NotEqual(t, risk1, risk3, "PauseOnRisk should change from Peak to Rest")
}

// TestTensionAdjustment_Timestamp tests that timestamp is recorded
func TestTensionAdjustment_Timestamp(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	before := time.Now()
	tm.AdjustTension(gameState, -10, "test")
	after := time.Now()

	history := tm.GetAdjustmentHistory()
	require.Len(t, history, 1)

	timestamp := history[0].Timestamp
	assert.True(t, timestamp.After(before) || timestamp.Equal(before), "Timestamp should be after or equal to before time")
	assert.True(t, timestamp.Before(after) || timestamp.Equal(after), "Timestamp should be before or equal to after time")
}

// TestEdgeCases_ZeroThreshold tests edge case with zero minimum threshold
func TestEdgeCases_ZeroThreshold(t *testing.T) {
	config := DefaultTensionManagerConfig()
	config.MinTensionThreshold = 0
	tm := NewTensionManager(DefaultGuardianConfig(), nil, config)
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(10)

	// Should be able to reduce to 0
	tm.AdjustTension(gameState, -10, "test")
	assert.Equal(t, 0, gameState.Tension.GetValue())

	// Should not go below 0
	tm.AdjustTension(gameState, -5, "test")
	assert.Equal(t, 0, gameState.Tension.GetValue())
}

// TestEdgeCases_NegativeDelta tests negative delta adjustments
func TestEdgeCases_NegativeDelta(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Large negative delta
	tm.AdjustTension(gameState, -100, "test")
	assert.GreaterOrEqual(t, gameState.Tension.GetValue(), 10, "Should not go below minimum threshold")
}

// TestEdgeCases_PositiveDelta tests positive delta adjustments
func TestEdgeCases_PositiveDelta(t *testing.T) {
	tm := NewTensionManager(DefaultGuardianConfig(), nil, DefaultTensionManagerConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension.SetValue(50)

	// Large positive delta
	tm.AdjustTension(gameState, 100, "test")
	assert.Equal(t, 150, gameState.Tension.GetValue(), "Should allow increasing above 100")
}
