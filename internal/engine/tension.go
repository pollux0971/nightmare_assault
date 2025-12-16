package engine

import (
	"encoding/json"
	"sync"
)

// TensionLevel represents the tension intensity level
type TensionLevel string

const (
	TensionLevelLow    TensionLevel = "LOW"    // 0-29: 鋪墊階段
	TensionLevelMedium TensionLevel = "MEDIUM" // 30-69: 懸疑階段
	TensionLevelHigh   TensionLevel = "HIGH"   // 70-100: 高潮階段
)

// TensionHistoryEntry records a single tension change event
type TensionHistoryEntry struct {
	Beat      int    `json:"beat"`       // 回合數
	OldValue  int    `json:"old_value"`  // 變化前張力值
	NewValue  int    `json:"new_value"`  // 變化後張力值
	Delta     int    `json:"delta"`      // 變化量
	Reason    string `json:"reason"`     // 變化原因
	EventType string `json:"event_type"` // 事件類型 (rule_violation, npc_death, etc.)
}

// TensionState represents the tension/suspense state of the game.
// Manages 0-100 tension value with automatic level calculation and history tracking.
type TensionState struct {
	Value            int                    `json:"value"`              // 當前張力值 (0-100)
	Level            TensionLevel           `json:"level"`              // 當前張力等級
	BeatsSinceClimax int                    `json:"beats_since_climax"` // 距離上次高潮的回合數
	History          []*TensionHistoryEntry `json:"history"`            // 張力歷史記錄 (最多50筆)

	mu sync.RWMutex `json:"-"` // 線程安全鎖
}

// NewTensionState creates a new TensionState with default initial values.
// Initial tension is set to 20 (LOW level) for a calm opening.
func NewTensionState() *TensionState {
	return &TensionState{
		Value:            20,
		Level:            TensionLevelLow,
		BeatsSinceClimax: 0,
		History:          make([]*TensionHistoryEntry, 0, 50),
	}
}

// GetValue returns the current tension value (thread-safe)
func (t *TensionState) GetValue() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Value
}

// GetLevel returns the current tension level (thread-safe)
func (t *TensionState) GetLevel() TensionLevel {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Level
}

// GetBeatsSinceClimax returns beats since last climax (thread-safe)
func (t *TensionState) GetBeatsSinceClimax() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.BeatsSinceClimax
}

// GetHistory returns a copy of tension history (thread-safe)
func (t *TensionState) GetHistory() []*TensionHistoryEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*TensionHistoryEntry, len(t.History))
	copy(history, t.History)
	return history
}

// SetValue sets the tension value and auto-updates the level (thread-safe)
func (t *TensionState) SetValue(value int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Value = value
	t.updateLevel()
}

// updateLevel calculates and updates the tension level based on current value.
// Must be called with lock held.
func (t *TensionState) updateLevel() {
	if t.Value >= 70 {
		t.Level = TensionLevelHigh
	} else if t.Value >= 30 {
		t.Level = TensionLevelMedium
	} else {
		t.Level = TensionLevelLow
	}
}

// CalculateLevel returns the level for a given tension value without modifying state
func CalculateLevel(value int) TensionLevel {
	if value >= 70 {
		return TensionLevelHigh
	} else if value >= 30 {
		return TensionLevelMedium
	}
	return TensionLevelLow
}

// AddToHistory adds a new entry to tension history (thread-safe).
// Maintains a maximum of 50 entries (FIFO).
func (t *TensionState) AddToHistory(entry *TensionHistoryEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.History = append(t.History, entry)

	// Keep only the last 50 entries
	if len(t.History) > 50 {
		t.History = t.History[len(t.History)-50:]
	}
}

// IncrementBeatsSinceClimax increments the beats counter (thread-safe)
func (t *TensionState) IncrementBeatsSinceClimax() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.BeatsSinceClimax++
}

// ResetBeatsSinceClimax resets the beats counter to 0 (thread-safe)
func (t *TensionState) ResetBeatsSinceClimax() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.BeatsSinceClimax = 0
}

// MarshalJSON implements custom JSON marshaling
func (t *TensionState) MarshalJSON() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	type Alias TensionState
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling
func (t *TensionState) UnmarshalJSON(data []byte) error {
	type Alias TensionState
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure history is initialized
	if t.History == nil {
		t.History = make([]*TensionHistoryEntry, 0, 50)
	}

	return nil
}
