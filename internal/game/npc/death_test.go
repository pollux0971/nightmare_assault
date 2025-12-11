package npc

import (
	"testing"
)

func TestDeathState(t *testing.T) {
	states := []DeathState{
		StateSafe,
		StateForeshadowed,
		StateWarned,
		StateEndangered,
		StateDead,
		StateSaved,
	}

	if len(states) != 6 {
		t.Errorf("Expected 6 death states, got %d", len(states))
	}
}

func TestForeshadow(t *testing.T) {
	foreshadow := Foreshadow{
		Type:    "obvious",
		Content: "小李突然咳嗽不止",
		Turn:    5,
	}

	if foreshadow.Type != "obvious" && foreshadow.Type != "subtle" {
		t.Errorf("Foreshadow type should be 'obvious' or 'subtle', got '%s'", foreshadow.Type)
	}

	if foreshadow.Content == "" {
		t.Error("Foreshadow content should not be empty")
	}

	if foreshadow.Turn < 0 {
		t.Error("Foreshadow turn should not be negative")
	}
}

func TestDeathEvent(t *testing.T) {
	event := DeathEvent{
		TeammateID:     "tm-001",
		CurrentState:   StateSafe,
		ForeshadowTurn: 5,
		WarningTurn:    7,
		DeadlineTurn:   8,
		Foreshadows:    []Foreshadow{},
	}

	if event.TeammateID != "tm-001" {
		t.Errorf("Expected TeammateID 'tm-001', got '%s'", event.TeammateID)
	}

	if event.CurrentState != StateSafe {
		t.Errorf("Expected CurrentState 'safe', got '%s'", event.CurrentState)
	}

	if event.ForeshadowTurn >= event.WarningTurn {
		t.Error("ForeshadowTurn should be before WarningTurn")
	}

	if event.WarningTurn >= event.DeadlineTurn {
		t.Error("WarningTurn should be before DeadlineTurn")
	}
}

func TestDeathManager_AddEvent(t *testing.T) {
	manager := NewDeathManager()

	event := DeathEvent{
		TeammateID:     "tm-001",
		CurrentState:   StateSafe,
		ForeshadowTurn: 5,
		WarningTurn:    7,
		DeadlineTurn:   8,
	}

	manager.AddEvent(event)

	retrieved, exists := manager.GetEvent("tm-001")
	if !exists {
		t.Error("Event should exist after adding")
	}

	if retrieved.TeammateID != "tm-001" {
		t.Errorf("Retrieved event TeammateID = %s, want tm-001", retrieved.TeammateID)
	}
}

func TestDeathManager_UpdateState(t *testing.T) {
	manager := NewDeathManager()

	event := DeathEvent{
		TeammateID:   "tm-001",
		CurrentState: StateSafe,
	}
	manager.AddEvent(event)

	manager.UpdateState("tm-001", StateForeshadowed)

	retrieved, _ := manager.GetEvent("tm-001")
	if retrieved.CurrentState != StateForeshadowed {
		t.Errorf("State = %s, want %s", retrieved.CurrentState, StateForeshadowed)
	}
}

func TestDeathManager_AddForeshadow(t *testing.T) {
	manager := NewDeathManager()

	event := DeathEvent{
		TeammateID:  "tm-001",
		Foreshadows: []Foreshadow{},
	}
	manager.AddEvent(event)

	foreshadow := Foreshadow{
		Type:    "obvious",
		Content: "Test foreshadow",
		Turn:    5,
	}

	manager.AddForeshadow("tm-001", foreshadow)

	retrieved, _ := manager.GetEvent("tm-001")
	if len(retrieved.Foreshadows) != 1 {
		t.Errorf("Expected 1 foreshadow, got %d", len(retrieved.Foreshadows))
	}

	if retrieved.Foreshadows[0].Content != "Test foreshadow" {
		t.Error("Foreshadow content mismatch")
	}
}

func TestDeathManager_CheckStateTransition(t *testing.T) {
	manager := NewDeathManager()

	event := DeathEvent{
		TeammateID:     "tm-001",
		CurrentState:   StateSafe,
		ForeshadowTurn: 5,
		WarningTurn:    7,
		DeadlineTurn:   8,
	}
	manager.AddEvent(event)

	// Turn 5: Should transition to Foreshadowed
	manager.CheckStateTransition("tm-001", 5)
	retrieved, _ := manager.GetEvent("tm-001")
	if retrieved.CurrentState != StateForeshadowed {
		t.Errorf("At turn 5, state should be Foreshadowed, got %s", retrieved.CurrentState)
	}

	// Turn 7: Should transition to Warned
	manager.CheckStateTransition("tm-001", 7)
	retrieved, _ = manager.GetEvent("tm-001")
	if retrieved.CurrentState != StateWarned {
		t.Errorf("At turn 7, state should be Warned, got %s", retrieved.CurrentState)
	}

	// Turn 8: Should transition to Endangered
	manager.CheckStateTransition("tm-001", 8)
	retrieved, _ = manager.GetEvent("tm-001")
	if retrieved.CurrentState != StateEndangered {
		t.Errorf("At turn 8, state should be Endangered, got %s", retrieved.CurrentState)
	}
}

func TestCalculateDeathSANLoss(t *testing.T) {
	tests := []struct {
		intimacy int
		want     int // Should be in range [15, 25]
	}{
		{0, 20},   // Base loss
		{50, 25},  // 20 + 50/10 = 25
		{100, 25}, // 20 + 100/10 = 30, clamped to 25
		{10, 21},  // 20 + 10/10 = 21
	}

	for _, tt := range tests {
		got := CalculateDeathSANLoss(tt.intimacy)
		if got < 15 || got > 25 {
			t.Errorf("CalculateDeathSANLoss(%d) = %d, should be in [15, 25]", tt.intimacy, got)
		}
		if tt.intimacy <= 50 && got != tt.want {
			t.Errorf("CalculateDeathSANLoss(%d) = %d, want %d", tt.intimacy, got, tt.want)
		}
	}
}

func TestPlayerIntervention(t *testing.T) {
	intervention := PlayerIntervention{
		Turn:     7,
		Action:   "拉住小李的手",
		Success:  true,
		Rationale: "及時干預",
	}

	if intervention.Action == "" {
		t.Error("Action should not be empty")
	}

	if intervention.Turn < 0 {
		t.Error("Turn should not be negative")
	}
}

// TestDeathManager_CheckStateTransition_SkippedTurns tests that the state machine
// handles turn skips correctly (e.g., jumping from turn 1 directly to turn 8)
func TestDeathManager_CheckStateTransition_SkippedTurns(t *testing.T) {
	manager := NewDeathManager()

	event := DeathEvent{
		TeammateID:     "tm-001",
		CurrentState:   StateSafe,
		ForeshadowTurn: 3,
		WarningTurn:    5,
		DeadlineTurn:   7,
	}
	manager.AddEvent(event)

	// Skip directly from turn 1 to turn 8 (past all transitions)
	// Should cascade through all state transitions
	manager.CheckStateTransition("tm-001", 8)

	retrieved, _ := manager.GetEvent("tm-001")
	if retrieved.CurrentState != StateEndangered {
		t.Errorf("After skipping to turn 8, state should be Endangered, got %s", retrieved.CurrentState)
	}
}

// TestDeathManager_CheckStateTransition_MultipleSkips tests various turn skip scenarios
func TestDeathManager_CheckStateTransition_MultipleSkips(t *testing.T) {
	tests := []struct {
		name           string
		startState     DeathState
		foreshadowTurn int
		warningTurn    int
		deadlineTurn   int
		currentTurn    int
		expectedState  DeathState
	}{
		{
			name:           "Skip from Safe to past Foreshadow",
			startState:     StateSafe,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    4,
			expectedState:  StateForeshadowed,
		},
		{
			name:           "Skip from Safe to past Warning",
			startState:     StateSafe,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    6,
			expectedState:  StateWarned,
		},
		{
			name:           "Skip from Safe to past Deadline",
			startState:     StateSafe,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    10,
			expectedState:  StateEndangered,
		},
		{
			name:           "Skip from Foreshadowed to past Deadline",
			startState:     StateForeshadowed,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    8,
			expectedState:  StateEndangered,
		},
		{
			name:           "Already Dead - no transition",
			startState:     StateDead,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    10,
			expectedState:  StateDead,
		},
		{
			name:           "Already Saved - no transition",
			startState:     StateSaved,
			foreshadowTurn: 3,
			warningTurn:    5,
			deadlineTurn:   7,
			currentTurn:    10,
			expectedState:  StateSaved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewDeathManager()

			event := DeathEvent{
				TeammateID:     "tm-test",
				CurrentState:   tt.startState,
				ForeshadowTurn: tt.foreshadowTurn,
				WarningTurn:    tt.warningTurn,
				DeadlineTurn:   tt.deadlineTurn,
			}
			manager.AddEvent(event)

			manager.CheckStateTransition("tm-test", tt.currentTurn)

			retrieved, _ := manager.GetEvent("tm-test")
			if retrieved.CurrentState != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, retrieved.CurrentState)
			}
		})
	}
}

// TestCalculateDeathSANLoss_EdgeCases tests edge cases for intimacy calculation
func TestCalculateDeathSANLoss_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		intimacy int
		wantMin  int
		wantMax  int
	}{
		{
			name:     "Negative intimacy",
			intimacy: -50,
			wantMin:  15,
			wantMax:  25,
		},
		{
			name:     "Above max intimacy",
			intimacy: 150,
			wantMin:  15,
			wantMax:  25,
		},
		{
			name:     "Zero intimacy",
			intimacy: 0,
			wantMin:  15,
			wantMax:  25,
		},
		{
			name:     "Max valid intimacy",
			intimacy: 100,
			wantMin:  15,
			wantMax:  25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDeathSANLoss(tt.intimacy)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateDeathSANLoss(%d) = %d, want in range [%d, %d]",
					tt.intimacy, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
