package engine

import (
	"encoding/json"
	"sync"
	"testing"
)

// Story 3.1 AC1: TensionState should include required fields
func TestNewTensionState(t *testing.T) {
	ts := NewTensionState()

	if ts == nil {
		t.Fatal("NewTensionState() returned nil")
	}

	// AC: Initial value should be 20
	if ts.GetValue() != 20 {
		t.Errorf("Expected initial value 20, got %d", ts.GetValue())
	}

	// AC: Level should be LOW for value 20
	if ts.GetLevel() != TensionLevelLow {
		t.Errorf("Expected level LOW, got %s", ts.GetLevel())
	}

	// AC: BeatsSinceClimax should start at 0
	if ts.GetBeatsSinceClimax() != 0 {
		t.Errorf("Expected BeatsSinceClimax 0, got %d", ts.GetBeatsSinceClimax())
	}

	// AC: History should be initialized
	if ts.GetHistory() == nil {
		t.Error("History should be initialized")
	}

	if len(ts.GetHistory()) != 0 {
		t.Errorf("History should start empty, got %d entries", len(ts.GetHistory()))
	}
}

// Story 3.1 AC2: Level should auto-update based on value
func TestTensionState_LevelCalculation(t *testing.T) {
	testCases := []struct {
		value         int
		expectedLevel TensionLevel
	}{
		{0, TensionLevelLow},
		{10, TensionLevelLow},
		{29, TensionLevelLow},
		{30, TensionLevelMedium},
		{50, TensionLevelMedium},
		{69, TensionLevelMedium},
		{70, TensionLevelHigh},
		{85, TensionLevelHigh},
		{100, TensionLevelHigh},
	}

	for _, tc := range testCases {
		ts := NewTensionState()
		ts.SetValue(tc.value)

		if ts.GetLevel() != tc.expectedLevel {
			t.Errorf("Value %d: expected level %s, got %s",
				tc.value, tc.expectedLevel, ts.GetLevel())
		}
	}
}

// Test CalculateLevel utility function
func TestCalculateLevel(t *testing.T) {
	testCases := []struct {
		value         int
		expectedLevel TensionLevel
	}{
		{0, TensionLevelLow},
		{29, TensionLevelLow},
		{30, TensionLevelMedium},
		{69, TensionLevelMedium},
		{70, TensionLevelHigh},
		{100, TensionLevelHigh},
	}

	for _, tc := range testCases {
		level := CalculateLevel(tc.value)
		if level != tc.expectedLevel {
			t.Errorf("CalculateLevel(%d): expected %s, got %s",
				tc.value, tc.expectedLevel, level)
		}
	}
}

// Test thread safety
func TestTensionState_ThreadSafety(t *testing.T) {
	ts := NewTensionState()
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes to Value
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			ts.SetValue(val)
		}(i % 100)
	}

	// Concurrent reads
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = ts.GetValue()
			_ = ts.GetLevel()
		}()
	}

	wg.Wait()
	// Test passes if no race condition occurs
}

// Test SetValue and GetValue
func TestTensionState_SetGetValue(t *testing.T) {
	ts := NewTensionState()

	testValues := []int{0, 20, 50, 75, 100}
	for _, val := range testValues {
		ts.SetValue(val)
		if got := ts.GetValue(); got != val {
			t.Errorf("SetValue(%d): got %d", val, got)
		}
	}
}

// Test BeatsSinceClimax operations
func TestTensionState_BeatsSinceClimax(t *testing.T) {
	ts := NewTensionState()

	// Test increment
	for i := 1; i <= 5; i++ {
		ts.IncrementBeatsSinceClimax()
		if got := ts.GetBeatsSinceClimax(); got != i {
			t.Errorf("After %d increments, got %d", i, got)
		}
	}

	// Test reset
	ts.ResetBeatsSinceClimax()
	if got := ts.GetBeatsSinceClimax(); got != 0 {
		t.Errorf("After reset, expected 0, got %d", got)
	}
}

// Test AddToHistory
func TestTensionState_AddToHistory(t *testing.T) {
	ts := NewTensionState()

	entry1 := &TensionHistoryEntry{
		Beat:      1,
		OldValue:  20,
		NewValue:  50,
		Delta:     30,
		Reason:    "規則違反",
		EventType: "rule_violation",
	}

	entry2 := &TensionHistoryEntry{
		Beat:      2,
		OldValue:  50,
		NewValue:  75,
		Delta:     25,
		Reason:    "NPC 死亡",
		EventType: "npc_death",
	}

	ts.AddToHistory(entry1)
	ts.AddToHistory(entry2)

	history := ts.GetHistory()
	if len(history) != 2 {
		t.Fatalf("Expected 2 history entries, got %d", len(history))
	}

	if history[0].Beat != 1 {
		t.Errorf("Entry 0: expected beat 1, got %d", history[0].Beat)
	}

	if history[1].EventType != "npc_death" {
		t.Errorf("Entry 1: expected event_type 'npc_death', got '%s'", history[1].EventType)
	}
}

// Test history limit (max 50 entries)
func TestTensionState_HistoryLimit(t *testing.T) {
	ts := NewTensionState()

	// Add 60 entries
	for i := 1; i <= 60; i++ {
		entry := &TensionHistoryEntry{
			Beat:      i,
			OldValue:  20,
			NewValue:  30,
			Delta:     10,
			Reason:    "測試",
			EventType: "test",
		}
		ts.AddToHistory(entry)
	}

	history := ts.GetHistory()

	// Should only keep last 50
	if len(history) != 50 {
		t.Errorf("Expected history to be limited to 50 entries, got %d", len(history))
	}

	// First entry should be beat 11 (entries 1-10 were removed)
	if history[0].Beat != 11 {
		t.Errorf("Expected first entry to be beat 11, got %d", history[0].Beat)
	}

	// Last entry should be beat 60
	if history[49].Beat != 60 {
		t.Errorf("Expected last entry to be beat 60, got %d", history[49].Beat)
	}
}

// Test JSON serialization
func TestTensionState_JSONSerialization(t *testing.T) {
	ts := NewTensionState()
	ts.SetValue(75)
	ts.IncrementBeatsSinceClimax()
	ts.IncrementBeatsSinceClimax()

	entry := &TensionHistoryEntry{
		Beat:      1,
		OldValue:  20,
		NewValue:  75,
		Delta:     55,
		Reason:    "測試序列化",
		EventType: "test",
	}
	ts.AddToHistory(entry)

	// Marshal
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var restored TensionState
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify data integrity
	if restored.Value != 75 {
		t.Errorf("Value mismatch: got %d, want 75", restored.Value)
	}

	if restored.Level != TensionLevelHigh {
		t.Errorf("Level mismatch: got %s, want %s", restored.Level, TensionLevelHigh)
	}

	if restored.BeatsSinceClimax != 2 {
		t.Errorf("BeatsSinceClimax mismatch: got %d, want 2", restored.BeatsSinceClimax)
	}

	if len(restored.History) != 1 {
		t.Errorf("History length mismatch: got %d, want 1", len(restored.History))
	}

	if restored.History[0].Reason != "測試序列化" {
		t.Errorf("History entry reason mismatch: got '%s'", restored.History[0].Reason)
	}
}

// Test JSON doesn't include mutex
func TestTensionState_JSONDoesNotIncludeMutex(t *testing.T) {
	ts := NewTensionState()

	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if _, exists := raw["mu"]; exists {
		t.Error("JSON should not include 'mu' field")
	}
}

// Test boundary conditions
func TestTensionState_BoundaryConditions(t *testing.T) {
	ts := NewTensionState()

	// Test negative value
	ts.SetValue(-10)
	if ts.GetValue() != -10 {
		t.Errorf("Should allow negative values, got %d", ts.GetValue())
	}
	if ts.GetLevel() != TensionLevelLow {
		t.Errorf("Negative value should map to LOW, got %s", ts.GetLevel())
	}

	// Test value > 100
	ts.SetValue(150)
	if ts.GetValue() != 150 {
		t.Errorf("Should allow values > 100, got %d", ts.GetValue())
	}
	if ts.GetLevel() != TensionLevelHigh {
		t.Errorf("Value > 100 should map to HIGH, got %s", ts.GetLevel())
	}

	// Test exact boundaries
	ts.SetValue(30)
	if ts.GetLevel() != TensionLevelMedium {
		t.Errorf("Value 30 should be MEDIUM, got %s", ts.GetLevel())
	}

	ts.SetValue(70)
	if ts.GetLevel() != TensionLevelHigh {
		t.Errorf("Value 70 should be HIGH, got %s", ts.GetLevel())
	}
}
