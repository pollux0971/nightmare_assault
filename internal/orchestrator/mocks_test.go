package orchestrator

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// Test clamp function
func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		expected int
	}{
		{"below min", -10, 0, 100, 0},
		{"at min", 0, 0, 100, 0},
		{"within range", 50, 0, 100, 50},
		{"at max", 100, 0, 100, 100},
		{"above max", 150, 0, 100, 100},
		{"far below", -1000, 0, 100, 0},
		{"far above", 1000, 0, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%d, %d, %d) = %d, want %d",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

// Test MockStateManager clamping behavior (Issue #5 - User suggestion)
func TestMockStateManager_ClampingBehavior(t *testing.T) {
	gameState := engine.NewGameStateV2()
	sm := NewMockStateManager(gameState)

	// Test HP clamping - lower bound
	sm.ApplyChanges(StateChanges{HPDelta: -200})
	if gameState.HP != 0 {
		t.Errorf("HP should clamp to 0, got %d", gameState.HP)
	}

	// Test HP clamping - upper bound
	sm.ApplyChanges(StateChanges{HPDelta: 200})
	if gameState.HP != 100 {
		t.Errorf("HP should clamp to 100, got %d", gameState.HP)
	}

	// Test SAN clamping - lower bound
	sm.ApplyChanges(StateChanges{SANDelta: -200})
	if gameState.SAN != 0 {
		t.Errorf("SAN should clamp to 0, got %d", gameState.SAN)
	}

	// Test SAN clamping - upper bound
	gameState.SAN = 50
	sm.ApplyChanges(StateChanges{SANDelta: 100})
	if gameState.SAN != 100 {
		t.Errorf("SAN should clamp to 100, got %d", gameState.SAN)
	}

	// Test Tension clamping - lower bound
	gameState.Tension.Value = 10
	sm.ApplyChanges(StateChanges{TensionDelta: -50})
	if gameState.Tension.Value != 0 {
		t.Errorf("Tension should clamp to 0, got %d", gameState.Tension.Value)
	}

	// Test Tension clamping - upper bound
	gameState.Tension.Value = 90
	sm.ApplyChanges(StateChanges{TensionDelta: 50})
	if gameState.Tension.Value != 100 {
		t.Errorf("Tension should clamp to 100, got %d", gameState.Tension.Value)
	}
}

// Test multiple changes applied together
func TestMockStateManager_MultipleChanges(t *testing.T) {
	gameState := engine.NewGameStateV2()
	sm := NewMockStateManager(gameState)

	// Apply multiple changes at once
	sm.ApplyChanges(StateChanges{
		HPDelta:      -20,
		SANDelta:     -30,
		TensionDelta: 15,
	})

	if gameState.HP != 80 {
		t.Errorf("Expected HP=80, got %d", gameState.HP)
	}

	if gameState.SAN != 70 {
		t.Errorf("Expected SAN=70, got %d", gameState.SAN)
	}

	// Note: TensionState initializes with Value=20 (see internal/engine/tension.go:42)
	// So starting value (20) + delta (15) = 35
	if gameState.Tension.Value != 35 {
		t.Errorf("Expected Tension=35 (initial 20 + delta 15), got %d", gameState.Tension.Value)
	}
}

// Test exact boundary conditions
func TestMockStateManager_BoundaryConditions(t *testing.T) {
	gameState := engine.NewGameStateV2()
	sm := NewMockStateManager(gameState)

	// Test exact 0 boundary
	gameState.HP = 5
	sm.ApplyChanges(StateChanges{HPDelta: -5})
	if gameState.HP != 0 {
		t.Errorf("HP should be exactly 0, got %d", gameState.HP)
	}

	// Test exact 100 boundary
	gameState.SAN = 95
	sm.ApplyChanges(StateChanges{SANDelta: 5})
	if gameState.SAN != 100 {
		t.Errorf("SAN should be exactly 100, got %d", gameState.SAN)
	}

	// Test no change when delta is 0
	initialHP := gameState.HP
	sm.ApplyChanges(StateChanges{HPDelta: 0})
	if gameState.HP != initialHP {
		t.Errorf("HP should not change when delta=0, got %d", gameState.HP)
	}
}

// Test nil gameState safety
func TestMockStateManager_NilGameState(t *testing.T) {
	sm := NewMockStateManager(nil)

	// Should not panic
	sm.ApplyChanges(StateChanges{
		HPDelta:      -50,
		SANDelta:     -50,
		TensionDelta: 50,
	})

	// Test passes if no panic occurs
}
