package game

import (
	"testing"
	"time"
)

func TestSanityState_Constants(t *testing.T) {
	states := []SanityState{
		SanityClearHeaded,
		SanityAnxious,
		SanityPanicked,
		SanityInsanity,
	}
	if len(states) != 4 {
		t.Error("Should have 4 sanity states")
	}
}

func TestNewPlayerStats(t *testing.T) {
	stats := NewPlayerStats()

	if stats.HP != 100 {
		t.Errorf("Initial HP = %d, want 100", stats.HP)
	}
	if stats.SAN != 100 {
		t.Errorf("Initial SAN = %d, want 100", stats.SAN)
	}
	if stats.MaxHP != 100 {
		t.Errorf("MaxHP = %d, want 100", stats.MaxHP)
	}
	if stats.MaxSAN != 100 {
		t.Errorf("MaxSAN = %d, want 100", stats.MaxSAN)
	}
	if stats.State != SanityClearHeaded {
		t.Errorf("Initial state = %v, want SanityClearHeaded", stats.State)
	}
}

func TestPlayerStats_GetSanityState(t *testing.T) {
	tests := []struct {
		san      int
		expected SanityState
	}{
		{100, SanityClearHeaded},
		{80, SanityClearHeaded},
		{79, SanityAnxious},
		{50, SanityAnxious},
		{49, SanityPanicked},
		{20, SanityPanicked},
		{19, SanityInsanity},
		{0, SanityInsanity},
	}

	for _, tt := range tests {
		stats := &PlayerStats{SAN: tt.san}
		state := stats.GetSanityState()
		if state != tt.expected {
			t.Errorf("SAN=%d: GetSanityState = %v, want %v", tt.san, state, tt.expected)
		}
	}
}

func TestPlayerStats_IsDead(t *testing.T) {
	stats := &PlayerStats{HP: 0}
	if !stats.IsDead() {
		t.Error("HP=0 should return true for IsDead")
	}

	stats.HP = 1
	if stats.IsDead() {
		t.Error("HP=1 should return false for IsDead")
	}
}

func TestPlayerStats_IsInsane(t *testing.T) {
	stats := &PlayerStats{SAN: 19}
	if !stats.IsInsane() {
		t.Error("SAN=19 should return true for IsInsane")
	}

	stats.SAN = 20
	if stats.IsInsane() {
		t.Error("SAN=20 should return false for IsInsane")
	}
}

func TestStatChange_Timestamp(t *testing.T) {
	before := time.Now()
	change := NewStatChange("HP", -10, 90, "Test event")
	after := time.Now()

	if change.Timestamp.Before(before) || change.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
	if change.StatType != "HP" {
		t.Errorf("StatType = %q, want 'HP'", change.StatType)
	}
	if change.Delta != -10 {
		t.Errorf("Delta = %d, want -10", change.Delta)
	}
	if change.NewValue != 90 {
		t.Errorf("NewValue = %d, want 90", change.NewValue)
	}
}

func TestNewStatsManager(t *testing.T) {
	config := &GameConfig{
		Difficulty: DifficultyEasy,
	}
	mgr := NewStatsManager(config)

	if mgr == nil {
		t.Fatal("NewStatsManager should not return nil")
	}
	if mgr.stats == nil {
		t.Error("StatsManager should have stats")
	}
	if mgr.difficulty != DifficultyEasy {
		t.Errorf("Difficulty = %v, want DifficultyEasy", mgr.difficulty)
	}
}

func TestStatsManager_GetStats(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	stats := mgr.GetStats()
	if stats.HP != 100 {
		t.Errorf("HP = %d, want 100", stats.HP)
	}
}

func TestStatsManager_ApplyDelta_HP(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	err := mgr.ApplyDelta("HP", -20, "Test damage")
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	// Easy difficulty: 0.5x damage
	expectedHP := 100 - int(float64(20)*0.5)
	if mgr.stats.HP != expectedHP {
		t.Errorf("After damage, HP = %d, want %d", mgr.stats.HP, expectedHP)
	}

	// Check history
	if len(mgr.stats.History) != 1 {
		t.Errorf("History length = %d, want 1", len(mgr.stats.History))
	}
}

func TestStatsManager_ApplyDelta_SAN(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	err := mgr.ApplyDelta("SAN", -30, "Terror")
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	// Easy difficulty: 0.7x SAN drain
	expectedSAN := 100 - int(float64(30)*0.7)
	if mgr.stats.SAN != expectedSAN {
		t.Errorf("After SAN drain, SAN = %d, want %d", mgr.stats.SAN, expectedSAN)
	}

	// State should change to Anxious
	if mgr.stats.State != SanityAnxious {
		t.Errorf("State = %v, want SanityAnxious", mgr.stats.State)
	}
}

func TestStatsManager_ApplyDelta_Bounds(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	// Test lower bound
	err := mgr.ApplyDelta("HP", -200, "Massive damage")
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}
	if mgr.stats.HP != 0 {
		t.Errorf("HP should be bounded to 0, got %d", mgr.stats.HP)
	}

	// Test upper bound
	err = mgr.ApplyDelta("HP", 200, "Overheal")
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}
	if mgr.stats.HP != 100 {
		t.Errorf("HP should be bounded to 100, got %d", mgr.stats.HP)
	}
}

func TestStatsManager_ApplyDelta_InvalidStat(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	err := mgr.ApplyDelta("INVALID", -10, "Test")
	if err == nil {
		t.Error("Expected error for invalid stat type")
	}
}

func TestStatsManager_Reset(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	// Make some changes
	mgr.ApplyDelta("HP", -50, "Test")
	mgr.ApplyDelta("SAN", -50, "Test")

	// Reset
	mgr.Reset()

	if mgr.stats.HP != 100 {
		t.Errorf("After reset, HP = %d, want 100", mgr.stats.HP)
	}
	if mgr.stats.SAN != 100 {
		t.Errorf("After reset, SAN = %d, want 100", mgr.stats.SAN)
	}
	if len(mgr.stats.History) != 0 {
		t.Error("After reset, history should be empty")
	}
}

func TestStatsManager_GetHistory(t *testing.T) {
	mgr := NewStatsManager(&GameConfig{Difficulty: DifficultyEasy})

	mgr.ApplyDelta("HP", -10, "First")
	mgr.ApplyDelta("SAN", -5, "Second")

	history := mgr.GetHistory()
	if len(history) != 2 {
		t.Errorf("History length = %d, want 2", len(history))
	}
}

func TestDifficultyMultipliers(t *testing.T) {
	tests := []struct {
		difficulty   DifficultyLevel
		hpMultiplier float64
		sanMultiplier float64
	}{
		{DifficultyEasy, 0.5, 0.7},
		{DifficultyHard, 1.0, 1.0},
		{DifficultyHell, 1.5, 1.3},
	}

	for _, tt := range tests {
		mgr := NewStatsManager(&GameConfig{Difficulty: tt.difficulty})

		// Test HP damage
		mgr.ApplyDelta("HP", -20, "Test")
		expectedHP := 100 - int(float64(20)*tt.hpMultiplier)
		if mgr.stats.HP != expectedHP {
			t.Errorf("%v: HP = %d, want %d", tt.difficulty, mgr.stats.HP, expectedHP)
		}

		// Reset and test SAN drain
		mgr.Reset()
		mgr.ApplyDelta("SAN", -20, "Test")
		expectedSAN := 100 - int(float64(20)*tt.sanMultiplier)
		if mgr.stats.SAN != expectedSAN {
			t.Errorf("%v: SAN = %d, want %d", tt.difficulty, mgr.stats.SAN, expectedSAN)
		}
	}
}

func TestStatChange_IsSignificant(t *testing.T) {
	tests := []struct {
		delta    int
		expected bool
	}{
		{-10, false},
		{-11, true},
		{11, true},
		{10, false},
		{-15, true},
		{0, false},
	}

	for _, tt := range tests {
		change := StatChange{Delta: tt.delta}
		if change.IsSignificant() != tt.expected {
			t.Errorf("Delta=%d: IsSignificant = %v, want %v", tt.delta, change.IsSignificant(), tt.expected)
		}
	}
}
