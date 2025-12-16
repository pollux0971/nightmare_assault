package engine

import (
	"testing"
)

// Story 3.2 AC1: Calculate correct delta for each event type
func TestTensionManager_CalculateDelta(t *testing.T) {
	tm := NewTensionManager(nil)

	testCases := []struct {
		eventType     TensionEventType
		expectedDelta int
	}{
		{EventRuleViolation, 30},
		{EventNPCDeath, 25},
		{EventMajorReveal, 20},
		{EventSceneChange, 10},
		{EventSafeAction, -5},
		{EventCooldown, -20},
	}

	for _, tc := range testCases {
		delta := tm.CalculateDelta(tc.eventType)
		if delta != tc.expectedDelta {
			t.Errorf("Event %s: expected delta %d, got %d",
				tc.eventType, tc.expectedDelta, delta)
		}
	}
}

// Story 3.2 AC1: Apply event and update state
func TestTensionManager_ApplyEvent(t *testing.T) {
	state := NewTensionState()
	state.SetValue(50) // Start at 50
	tm := NewTensionManager(state)

	event := TensionEvent{
		Type:   EventRuleViolation,
		Reason: "玩家違反規則",
	}

	newValue := tm.ApplyEvent(1, event)

	// Should be 50 + 30 = 80
	if newValue != 80 {
		t.Errorf("Expected new value 80, got %d", newValue)
	}

	if tm.GetValue() != 80 {
		t.Errorf("State not updated: expected 80, got %d", tm.GetValue())
	}

	// Check history was recorded
	history := state.GetHistory()
	if len(history) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(history))
	}

	entry := history[0]
	if entry.OldValue != 50 {
		t.Errorf("History: expected OldValue 50, got %d", entry.OldValue)
	}
	if entry.NewValue != 80 {
		t.Errorf("History: expected NewValue 80, got %d", entry.NewValue)
	}
	if entry.Delta != 30 {
		t.Errorf("History: expected Delta 30, got %d", entry.Delta)
	}
}

// Story 3.2 AC2: Test climax cooldown mechanism
func TestTensionManager_ProcessBeat_Cooldown(t *testing.T) {
	state := NewTensionState()
	state.SetValue(85) // Start in climax zone
	tm := NewTensionManager(state)

	// Process beat should apply cooldown
	cooledDown := tm.ProcessBeat(10)

	if !cooledDown {
		t.Error("ProcessBeat should return true when cooldown is applied")
	}

	// Tension should be reduced by 20: 85 - 20 = 65
	if tm.GetValue() != 65 {
		t.Errorf("After cooldown: expected 65, got %d", tm.GetValue())
	}

	// BeatsSinceClimax should have been incremented then maybe reset
	// Actually, it gets incremented, then cooldown is applied (which doesn't reset it again)
	// Let me check the logic... ProcessBeat increments, then applies cooldown if >= 80
	// ApplyEvent resets BeatsSinceClimax if newValue >= 80
	// After cooldown: 85 -> 65, which is < 80, so BeatsSinceClimax won't be reset by ApplyEvent

	if tm.state.GetBeatsSinceClimax() != 1 {
		t.Errorf("BeatsSinceClimax should be 1, got %d", tm.state.GetBeatsSinceClimax())
	}

	// Check history includes cooldown event
	history := state.GetHistory()
	if len(history) == 0 {
		t.Fatal("History should have cooldown entry")
	}

	lastEntry := history[len(history)-1]
	if lastEntry.EventType != string(EventCooldown) {
		t.Errorf("Last history entry should be cooldown, got %s", lastEntry.EventType)
	}
}

// Story 3.2 AC2: No cooldown when below climax threshold
func TestTensionManager_ProcessBeat_NoCooldown(t *testing.T) {
	state := NewTensionState()
	state.SetValue(50) // Below climax zone
	tm := NewTensionManager(state)

	cooledDown := tm.ProcessBeat(10)

	if cooledDown {
		t.Error("ProcessBeat should return false when no cooldown is applied")
	}

	// Value should remain unchanged
	if tm.GetValue() != 50 {
		t.Errorf("Value should remain 50, got %d", tm.GetValue())
	}

	// BeatsSinceClimax should be incremented
	if tm.state.GetBeatsSinceClimax() != 1 {
		t.Errorf("BeatsSinceClimax should be 1, got %d", tm.state.GetBeatsSinceClimax())
	}
}

// Story 3.2 AC2: BeatsSinceClimax resets when reaching climax
func TestTensionManager_BeatsSinceClimaxReset(t *testing.T) {
	state := NewTensionState()
	state.SetValue(60)
	tm := NewTensionManager(state)

	// Increment beats
	tm.state.IncrementBeatsSinceClimax()
	tm.state.IncrementBeatsSinceClimax()
	tm.state.IncrementBeatsSinceClimax()

	if tm.state.GetBeatsSinceClimax() != 3 {
		t.Fatalf("Setup: BeatsSinceClimax should be 3, got %d", tm.state.GetBeatsSinceClimax())
	}

	// Apply event that brings tension to climax zone
	event := TensionEvent{
		Type:   EventRuleViolation,
		Reason: "觸發高潮",
	}
	newValue := tm.ApplyEvent(10, event)

	// 60 + 30 = 90, which is >= 80
	if newValue < 80 {
		t.Fatalf("Tension should be in climax zone, got %d", newValue)
	}

	// BeatsSinceClimax should be reset to 0
	if tm.state.GetBeatsSinceClimax() != 0 {
		t.Errorf("BeatsSinceClimax should be reset to 0, got %d", tm.state.GetBeatsSinceClimax())
	}
}

// Test applying multiple events
func TestTensionManager_ApplyMultipleEvents(t *testing.T) {
	state := NewTensionState()
	state.SetValue(30)
	tm := NewTensionManager(state)

	events := []TensionEvent{
		{Type: EventSceneChange, Reason: "進入新場景"},
		{Type: EventMajorReveal, Reason: "發現重要線索"},
		{Type: EventSafeAction, Reason: "找到安全區"},
	}

	finalValue := tm.ApplyMultipleEvents(5, events)

	// 30 + 10 + 20 - 5 = 55
	if finalValue != 55 {
		t.Errorf("Expected final value 55, got %d", finalValue)
	}

	// History should have 3 entries
	history := state.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(history))
	}
}

// Test NewTensionManager with nil state
func TestNewTensionManager_NilState(t *testing.T) {
	tm := NewTensionManager(nil)

	if tm == nil {
		t.Fatal("NewTensionManager(nil) should create manager with new state")
	}

	if tm.GetState() == nil {
		t.Error("Manager should have a state")
	}

	// Should have default initial value
	if tm.GetValue() != 20 {
		t.Errorf("Expected initial value 20, got %d", tm.GetValue())
	}
}

// Test FormatEventReason helper
func TestFormatEventReason(t *testing.T) {
	testCases := []struct {
		eventType TensionEventType
		details   string
		expected  string
	}{
		{EventRuleViolation, "不能直視怪物", "違反規則：不能直視怪物"},
		{EventNPCDeath, "小明被殺", "NPC 死亡：小明被殺"},
		{EventMajorReveal, "發現真相", "重大揭示：發現真相"},
		{EventSceneChange, "從醫院到森林", "場景變化：從醫院到森林"},
		{EventSafeAction, "躲在櫃子裡", "安全行動：躲在櫃子裡"},
		{EventCooldown, "", "高潮冷卻機制"},
	}

	for _, tc := range testCases {
		result := FormatEventReason(tc.eventType, tc.details)
		if result != tc.expected {
			t.Errorf("Event %s: expected '%s', got '%s'",
				tc.eventType, tc.expected, result)
		}
	}
}

// Test level changes through events
func TestTensionManager_LevelChanges(t *testing.T) {
	state := NewTensionState()
	state.SetValue(25) // LOW
	tm := NewTensionManager(state)

	// Should start as LOW
	if tm.GetLevel() != TensionLevelLow {
		t.Errorf("Expected initial level LOW, got %s", tm.GetLevel())
	}

	// Apply event to reach MEDIUM
	event1 := TensionEvent{Type: EventSceneChange, Reason: "test"}
	tm.ApplyEvent(1, event1)
	// 25 + 10 = 35 (MEDIUM)

	if tm.GetLevel() != TensionLevelMedium {
		t.Errorf("Expected level MEDIUM, got %s", tm.GetLevel())
	}

	// Apply event to reach HIGH
	event2 := TensionEvent{Type: EventRuleViolation, Reason: "test"}
	event3 := TensionEvent{Type: EventMajorReveal, Reason: "test"}
	tm.ApplyEvent(2, event2)
	tm.ApplyEvent(3, event3)
	// 35 + 30 + 20 = 85 (HIGH)

	if tm.GetLevel() != TensionLevelHigh {
		t.Errorf("Expected level HIGH, got %s", tm.GetLevel())
	}
}

// Test cooldown brings level back down
func TestTensionManager_CooldownLevelChange(t *testing.T) {
	state := NewTensionState()
	state.SetValue(85) // HIGH and >= 80 (climax zone)
	tm := NewTensionManager(state)

	if tm.GetLevel() != TensionLevelHigh {
		t.Fatalf("Setup: expected HIGH, got %s", tm.GetLevel())
	}

	// Process beat triggers cooldown: 85 - 20 = 65 (MEDIUM)
	tm.ProcessBeat(10)

	if tm.GetValue() != 65 {
		t.Errorf("Expected value 65, got %d", tm.GetValue())
	}

	if tm.GetLevel() != TensionLevelMedium {
		t.Errorf("After cooldown: expected MEDIUM, got %s", tm.GetLevel())
	}
}
