package guardian

import (
	"fmt"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTensionPhase_String tests the String() method of TensionPhase
func TestTensionPhase_String(t *testing.T) {
	tests := []struct {
		name     string
		phase    TensionPhase
		expected string
	}{
		{"Rest", PhaseRest, "Rest"},
		{"Buildup", PhaseBuildup, "Buildup"},
		{"Peak", PhasePeak, "Peak"},
		{"Release", PhaseRelease, "Release"},
		{"Unknown", TensionPhase(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.phase.String()
			assert.Equal(t, tt.expected, result, "String() should return correct value")
		})
	}
}

// TestCalculatePhaseFromTension tests tension value to phase mapping
func TestCalculatePhaseFromTension(t *testing.T) {
	tests := []struct {
		name     string
		tension  int
		expected TensionPhase
	}{
		// Rest phase (0-24)
		{"Tension 0 -> Rest", 0, PhaseRest},
		{"Tension 10 -> Rest", 10, PhaseRest},
		{"Tension 24 -> Rest", 24, PhaseRest},

		// Buildup phase (25-59)
		{"Tension 25 -> Buildup", 25, PhaseBuildup},
		{"Tension 40 -> Buildup", 40, PhaseBuildup},
		{"Tension 59 -> Buildup", 59, PhaseBuildup},

		// Peak phase (60-89)
		{"Tension 60 -> Peak", 60, PhasePeak},
		{"Tension 75 -> Peak", 75, PhasePeak},
		{"Tension 89 -> Peak", 89, PhasePeak},

		// Release phase (90-100)
		{"Tension 90 -> Release", 90, PhaseRelease},
		{"Tension 95 -> Release", 95, PhaseRelease},
		{"Tension 100 -> Release", 100, PhaseRelease},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePhaseFromTension(tt.tension)
			assert.Equal(t, tt.expected, result, "calculatePhaseFromTension should return correct phase")
		})
	}
}

// TestCalculatePhaseFromTension_BoundaryValues tests boundary values
func TestCalculatePhaseFromTension_BoundaryValues(t *testing.T) {
	// Test boundary values specifically
	assert.Equal(t, PhaseRest, calculatePhaseFromTension(24), "24 should be Rest")
	assert.Equal(t, PhaseBuildup, calculatePhaseFromTension(25), "25 should be Buildup")
	assert.Equal(t, PhaseBuildup, calculatePhaseFromTension(59), "59 should be Buildup")
	assert.Equal(t, PhasePeak, calculatePhaseFromTension(60), "60 should be Peak")
	assert.Equal(t, PhasePeak, calculatePhaseFromTension(89), "89 should be Peak")
	assert.Equal(t, PhaseRelease, calculatePhaseFromTension(90), "90 should be Release")
}

// TestUpdateTensionPhase_NilGameState tests handling of nil gameState
func TestUpdateTensionPhase_NilGameState(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())

	// Should not panic with nil gameState
	assert.NotPanics(t, func() {
		changed := guardian.updateTensionPhase(nil)
		assert.False(t, changed, "Should return false for nil gameState")
	}, "updateTensionPhase should handle nil gameState gracefully")
}

// TestUpdateTensionPhase_NilTensionState tests handling of nil Tension
func TestUpdateTensionPhase_NilTensionState(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()
	gameState.Tension = nil

	// Should not panic with nil Tension
	assert.NotPanics(t, func() {
		changed := guardian.updateTensionPhase(gameState)
		assert.False(t, changed, "Should return false for nil Tension")
	}, "updateTensionPhase should handle nil Tension gracefully")
}

// TestUpdateTensionPhase_PhaseChange tests phase change detection
func TestUpdateTensionPhase_PhaseChange(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Initial phase should be Rest
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase(), "Initial phase should be Rest")

	// Set tension to Buildup range
	gameState.Tension.SetValue(30)
	changed := guardian.updateTensionPhase(gameState)

	assert.True(t, changed, "Phase should have changed")
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase(), "Phase should be Buildup")
}

// TestUpdateTensionPhase_NoPhaseChange tests when phase doesn't change
func TestUpdateTensionPhase_NoPhaseChange(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Set to Rest range
	gameState.Tension.SetValue(10)
	guardian.updateTensionPhase(gameState)
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase(), "Should be Rest")

	// Change tension but stay in Rest range
	gameState.Tension.SetValue(15)
	changed := guardian.updateTensionPhase(gameState)

	assert.False(t, changed, "Phase should not have changed")
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase(), "Should still be Rest")
}

// TestUpdateTensionPhase_AllPhaseTransitions tests all phase transitions
func TestUpdateTensionPhase_AllPhaseTransitions(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Start in Rest
	gameState.Tension.SetValue(10)
	guardian.updateTensionPhase(gameState)
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase())

	// Rest -> Buildup
	gameState.Tension.SetValue(30)
	changed := guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Rest -> Buildup should change")
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase())

	// Buildup -> Peak
	gameState.Tension.SetValue(70)
	changed = guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Buildup -> Peak should change")
	assert.Equal(t, PhasePeak, guardian.GetCurrentPhase())

	// Peak -> Release
	gameState.Tension.SetValue(95)
	changed = guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Peak -> Release should change")
	assert.Equal(t, PhaseRelease, guardian.GetCurrentPhase())

	// Release -> Peak (going back down)
	gameState.Tension.SetValue(80)
	changed = guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Release -> Peak should change")
	assert.Equal(t, PhasePeak, guardian.GetCurrentPhase())

	// Peak -> Buildup
	gameState.Tension.SetValue(50)
	changed = guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Peak -> Buildup should change")
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase())

	// Buildup -> Rest
	gameState.Tension.SetValue(20)
	changed = guardian.updateTensionPhase(gameState)
	assert.True(t, changed, "Buildup -> Rest should change")
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase())
}

// TestOnTurnEnd_UpdatesPhase tests that OnTurnEnd calls updateTensionPhase
func TestOnTurnEnd_UpdatesPhase(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Initial phase is Rest
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase())

	// Set tension to Buildup range and call OnTurnEnd
	gameState.Tension.SetValue(40)
	guardian.OnTurnEnd(gameState)

	// Phase should now be Buildup
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase(), "OnTurnEnd should update phase")
}

// TestOnTurnEnd_PhaseProgression tests phase progression through multiple turns
func TestOnTurnEnd_PhaseProgression(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Turn 1: Low tension
	gameState.Tension.SetValue(15)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase(), "Turn 1: Rest")

	// Turn 2: Medium tension
	gameState.Tension.SetValue(45)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase(), "Turn 2: Buildup")

	// Turn 3: High tension
	gameState.Tension.SetValue(75)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhasePeak, guardian.GetCurrentPhase(), "Turn 3: Peak")

	// Turn 4: Very high tension
	gameState.Tension.SetValue(95)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhaseRelease, guardian.GetCurrentPhase(), "Turn 4: Release")
}

// TestGetCurrentPhase tests the GetCurrentPhase getter method
func TestGetCurrentPhase(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Initial phase
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase())

	// Change to Buildup
	gameState.Tension.SetValue(35)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhaseBuildup, guardian.GetCurrentPhase())

	// Change to Peak
	gameState.Tension.SetValue(65)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhasePeak, guardian.GetCurrentPhase())

	// Change to Release
	gameState.Tension.SetValue(92)
	guardian.OnTurnEnd(gameState)
	assert.Equal(t, PhaseRelease, guardian.GetCurrentPhase())
}

// TestIntegrationScenario_TensionCurve tests a complete tension curve scenario
func TestIntegrationScenario_TensionCurve(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Scenario: Game starts calm, builds tension, peaks, then releases
	tensionProgression := []struct {
		turn     int
		tension  int
		expected TensionPhase
	}{
		{1, 10, PhaseRest},
		{2, 15, PhaseRest},
		{3, 30, PhaseBuildup},
		{4, 45, PhaseBuildup},
		{5, 55, PhaseBuildup},
		{6, 65, PhasePeak},
		{7, 75, PhasePeak},
		{8, 85, PhasePeak},
		{9, 92, PhaseRelease},
		{10, 95, PhaseRelease},
	}

	for _, tp := range tensionProgression {
		t.Run(fmt.Sprintf("Turn %d", tp.turn), func(t *testing.T) {
			gameState.Tension.SetValue(tp.tension)
			guardian.OnTurnEnd(gameState)
			assert.Equal(t, tp.expected, guardian.GetCurrentPhase(),
				"Turn %d: tension=%d should be %s", tp.turn, tp.tension, tp.expected.String())
		})
	}
}

// TestIntegrationScenario_TensionFluctuation tests fluctuating tension
func TestIntegrationScenario_TensionFluctuation(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())
	gameState := engine.NewGameStateV2()

	// Scenario: Tension goes up and down
	tensionProgression := []struct {
		turn     int
		tension  int
		expected TensionPhase
	}{
		{1, 20, PhaseRest},
		{2, 40, PhaseBuildup},
		{3, 25, PhaseBuildup},
		{4, 70, PhasePeak},
		{5, 50, PhaseBuildup},
		{6, 90, PhaseRelease},
		{7, 60, PhasePeak},
		{8, 20, PhaseRest},
	}

	for _, tp := range tensionProgression {
		t.Run(fmt.Sprintf("Turn %d", tp.turn), func(t *testing.T) {
			gameState.Tension.SetValue(tp.tension)
			guardian.OnTurnEnd(gameState)
			assert.Equal(t, tp.expected, guardian.GetCurrentPhase(),
				"Turn %d: tension=%d should be %s", tp.turn, tp.tension, tp.expected.String())
		})
	}
}

// TestNewExperienceGuardian_InitialPhase tests that new guardian starts in Rest phase
func TestNewExperienceGuardian_InitialPhase(t *testing.T) {
	guardian := NewExperienceGuardian(DefaultGuardianConfig())

	require.NotNil(t, guardian, "NewExperienceGuardian should not return nil")
	assert.Equal(t, PhaseRest, guardian.GetCurrentPhase(), "New guardian should start in Rest phase")
}

// TestTensionPhase_EdgeCases tests edge cases for tension phase calculation
func TestTensionPhase_EdgeCases(t *testing.T) {
	// Test negative tension (shouldn't happen but good to test)
	assert.Equal(t, PhaseRest, calculatePhaseFromTension(-10), "Negative tension should be Rest")

	// Test very high tension
	assert.Equal(t, PhaseRelease, calculatePhaseFromTension(150), "Very high tension should be Release")

	// Test exact boundaries again
	assert.Equal(t, PhaseRest, calculatePhaseFromTension(0), "0 should be Rest")
	assert.Equal(t, PhaseRest, calculatePhaseFromTension(24), "24 should be Rest")
	assert.Equal(t, PhaseBuildup, calculatePhaseFromTension(25), "25 should be Buildup")
	assert.Equal(t, PhaseBuildup, calculatePhaseFromTension(59), "59 should be Buildup")
	assert.Equal(t, PhasePeak, calculatePhaseFromTension(60), "60 should be Peak")
	assert.Equal(t, PhasePeak, calculatePhaseFromTension(89), "89 should be Peak")
	assert.Equal(t, PhaseRelease, calculatePhaseFromTension(90), "90 should be Release")
	assert.Equal(t, PhaseRelease, calculatePhaseFromTension(100), "100 should be Release")
}
