package engine

import (
	"fmt"
)

// TensionEventType represents different types of events that affect tension
type TensionEventType string

const (
	EventRuleViolation TensionEventType = "rule_violation" // 違反規則
	EventNPCDeath      TensionEventType = "npc_death"      // NPC 死亡
	EventMajorReveal   TensionEventType = "major_reveal"   // 重大揭示
	EventSceneChange   TensionEventType = "scene_change"   // 場景變化
	EventSafeAction    TensionEventType = "safe_action"    // 安全行動
	EventClimax        TensionEventType = "climax"         // 高潮
	EventCooldown      TensionEventType = "cooldown"       // 冷卻
)

// TensionEvent represents an event that triggers tension changes
type TensionEvent struct {
	Type   TensionEventType
	Reason string // 詳細原因描述
}

// TensionManager manages tension state and calculates deltas
type TensionManager struct {
	state *TensionState
}

// NewTensionManager creates a new TensionManager with the given state
func NewTensionManager(state *TensionState) *TensionManager {
	if state == nil {
		state = NewTensionState()
	}
	return &TensionManager{
		state: state,
	}
}

// CalculateDelta calculates the tension delta for a given event type
func (tm *TensionManager) CalculateDelta(eventType TensionEventType) int {
	switch eventType {
	case EventRuleViolation:
		return 30
	case EventNPCDeath:
		return 25
	case EventMajorReveal:
		return 20
	case EventSceneChange:
		return 10
	case EventSafeAction:
		return -5
	case EventCooldown:
		return -20
	default:
		return 0
	}
}

// ApplyEvent applies a tension event and updates the state
// Returns the new tension value
func (tm *TensionManager) ApplyEvent(beat int, event TensionEvent) int {
	oldValue := tm.state.GetValue()
	delta := tm.CalculateDelta(event.Type)
	newValue := oldValue + delta

	// Update state
	tm.state.SetValue(newValue)

	// Record history
	entry := &TensionHistoryEntry{
		Beat:      beat,
		OldValue:  oldValue,
		NewValue:  newValue,
		Delta:     delta,
		Reason:    event.Reason,
		EventType: string(event.Type),
	}
	tm.state.AddToHistory(entry)

	// Check for climax
	if newValue >= 80 {
		tm.state.ResetBeatsSinceClimax()
	}

	return newValue
}

// ApplyMultipleEvents applies multiple events in sequence
func (tm *TensionManager) ApplyMultipleEvents(beat int, events []TensionEvent) int {
	finalValue := tm.state.GetValue()

	for _, event := range events {
		finalValue = tm.ApplyEvent(beat, event)
	}

	return finalValue
}

// ProcessBeat processes a beat, applying cooldown if in climax zone
// Returns true if cooldown was applied
func (tm *TensionManager) ProcessBeat(beat int) bool {
	currentValue := tm.state.GetValue()

	// Increment beats since climax
	tm.state.IncrementBeatsSinceClimax()

	// Apply cooldown if in climax zone (>= 80)
	if currentValue >= 80 {
		cooldownEvent := TensionEvent{
			Type:   EventCooldown,
			Reason: "高潮冷卻機制：避免張力疲勞",
		}
		tm.ApplyEvent(beat, cooldownEvent)
		return true
	}

	return false
}

// GetState returns the current tension state
func (tm *TensionManager) GetState() *TensionState {
	return tm.state
}

// GetValue returns the current tension value
func (tm *TensionManager) GetValue() int {
	return tm.state.GetValue()
}

// GetLevel returns the current tension level
func (tm *TensionManager) GetLevel() TensionLevel {
	return tm.state.GetLevel()
}

// FormatEventReason formats a reason string for common events
func FormatEventReason(eventType TensionEventType, details string) string {
	switch eventType {
	case EventRuleViolation:
		return fmt.Sprintf("違反規則：%s", details)
	case EventNPCDeath:
		return fmt.Sprintf("NPC 死亡：%s", details)
	case EventMajorReveal:
		return fmt.Sprintf("重大揭示：%s", details)
	case EventSceneChange:
		return fmt.Sprintf("場景變化：%s", details)
	case EventSafeAction:
		return fmt.Sprintf("安全行動：%s", details)
	case EventCooldown:
		return "高潮冷卻機制"
	default:
		return details
	}
}
