// Package game provides game-related types and logic for Nightmare Assault.
package game

import (
	"encoding/json"
	"sync"
	"time"
)

// ChoiceRecord represents a single choice made by the player.
// Story 7.3 AC5: Records choice content, timestamp, beat, and state snapshot.
type ChoiceRecord struct {
	// Choice metadata
	ChoiceText string    `json:"choice_text"` // The text of the choice made
	Timestamp  time.Time `json:"timestamp"`   // When the choice was made
	BeatNumber int       `json:"beat_number"` // Beat number when choice was made
	IsFreeText bool      `json:"is_free_text"` // Whether it was a free text input

	// State snapshot after choice
	HPAfter      int `json:"hp_after"`      // HP after applying choice consequences
	SANAfter     int `json:"san_after"`     // SAN after applying choice consequences
	TensionAfter int `json:"tension_after"` // Tension after applying choice consequences

	// Rule violations (Code Review Fix 7-3-3)
	RulesViolated int `json:"rules_violated"` // Number of rules violated by this choice

	// Additional context
	Scene       string `json:"scene"`        // Scene where choice was made
	Narration   string `json:"narration"`    // The narration that led to this choice (optional)
	Consequence string `json:"consequence"`  // Brief consequence description (optional)
}

// ChoiceHistory manages the history of player choices.
// Story 7.3 AC5: Records and retrieves choice history for /log command.
//
// Thread-safe implementation for concurrent access.
type ChoiceHistory struct {
	records []ChoiceRecord
	mu      sync.RWMutex
}

// NewChoiceHistory creates a new choice history manager.
func NewChoiceHistory() *ChoiceHistory {
	return &ChoiceHistory{
		records: make([]ChoiceRecord, 0),
	}
}

// RecordChoice records a new choice to history.
// Story 7.3 AC5: Save choice content, timestamp, beat, and state snapshot.
func (ch *ChoiceHistory) RecordChoice(record ChoiceRecord) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// Set timestamp if not already set
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	ch.records = append(ch.records, record)
}

// GetAllRecords returns all choice records in chronological order.
// Returns a copy to prevent external modification.
func (ch *ChoiceHistory) GetAllRecords() []ChoiceRecord {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	// Return a copy
	recordsCopy := make([]ChoiceRecord, len(ch.records))
	copy(recordsCopy, ch.records)
	return recordsCopy
}

// GetRecordsByBeatRange returns choices made within a beat range [start, end] inclusive.
func (ch *ChoiceHistory) GetRecordsByBeatRange(startBeat, endBeat int) []ChoiceRecord {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	filtered := make([]ChoiceRecord, 0)
	for _, record := range ch.records {
		if record.BeatNumber >= startBeat && record.BeatNumber <= endBeat {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetRecordsByScene returns all choices made in a specific scene.
func (ch *ChoiceHistory) GetRecordsByScene(scene string) []ChoiceRecord {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	filtered := make([]ChoiceRecord, 0)
	for _, record := range ch.records {
		if record.Scene == scene {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetRecent returns the N most recent choices.
func (ch *ChoiceHistory) GetRecent(n int) []ChoiceRecord {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if n <= 0 {
		return []ChoiceRecord{}
	}

	start := len(ch.records) - n
	if start < 0 {
		start = 0
	}

	// Return a copy of the recent records
	recent := make([]ChoiceRecord, len(ch.records)-start)
	copy(recent, ch.records[start:])
	return recent
}

// GetLastRecord returns the most recent choice record.
// Returns nil if no records exist.
func (ch *ChoiceHistory) GetLastRecord() *ChoiceRecord {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.records) == 0 {
		return nil
	}

	// Return a copy
	lastRecord := ch.records[len(ch.records)-1]
	return &lastRecord
}

// GetRecordCount returns the total number of recorded choices.
func (ch *ChoiceHistory) GetRecordCount() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.records)
}

// Clear clears all choice records.
// Useful for starting a new game.
func (ch *ChoiceHistory) Clear() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.records = make([]ChoiceRecord, 0)
}

// GetStatsSummary returns a summary of choices and their impact.
// Story 7.3 AC5: Support for death debrief analysis.
func (ch *ChoiceHistory) GetStatsSummary() *ChoiceStatsSummary {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	summary := &ChoiceStatsSummary{
		TotalChoices:     len(ch.records),
		FreeTextChoices:  0,
		PredefinedChoices: 0,
		TotalHPLost:      0,
		TotalSANLost:     0,
		TensionIncrease:  0,
	}

	if len(ch.records) == 0 {
		return summary
	}

	// Calculate initial state from first record (reverse calculate)
	initialHP := ch.records[0].HPAfter
	initialSAN := ch.records[0].SANAfter
	initialTension := ch.records[0].TensionAfter

	// Count choice types and rules violated
	for _, record := range ch.records {
		if record.IsFreeText {
			summary.FreeTextChoices++
		} else {
			summary.PredefinedChoices++
		}
		summary.TotalRulesViolated += record.RulesViolated
	}

	// Calculate total changes from first to last
	lastRecord := ch.records[len(ch.records)-1]
	summary.TotalHPLost = initialHP - lastRecord.HPAfter
	summary.TotalSANLost = initialSAN - lastRecord.SANAfter
	summary.TensionIncrease = lastRecord.TensionAfter - initialTension

	return summary
}

// ChoiceStatsSummary provides statistics about choices made.
type ChoiceStatsSummary struct {
	TotalChoices       int `json:"total_choices"`
	FreeTextChoices    int `json:"free_text_choices"`
	PredefinedChoices  int `json:"predefined_choices"`
	TotalHPLost        int `json:"total_hp_lost"`
	TotalSANLost       int `json:"total_san_lost"`
	TensionIncrease    int `json:"tension_increase"`
	TotalRulesViolated int `json:"total_rules_violated"` // Code Review Fix 7-3-3
}

// ToJSON serializes the choice history to JSON.
func (ch *ChoiceHistory) ToJSON() ([]byte, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	return json.MarshalIndent(ch.records, "", "  ")
}

// FromJSON deserializes the choice history from JSON.
func (ch *ChoiceHistory) FromJSON(data []byte) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	var records []ChoiceRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	ch.records = records
	return nil
}

// Clone creates a deep copy of the choice history.
func (ch *ChoiceHistory) Clone() *ChoiceHistory {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	clone := NewChoiceHistory()
	clone.records = make([]ChoiceRecord, len(ch.records))
	copy(clone.records, ch.records)

	return clone
}

// FormatForDisplay formats the choice history for display in /log command.
// Story 7.3 AC5: Support history review via /log command.
//
// Returns a formatted string with:
//   - Beat number
//   - Choice text
//   - State changes (HP/SAN/Tension)
func (ch *ChoiceHistory) FormatForDisplay(maxRecords int) string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.records) == 0 {
		return "尚未做出任何選擇。"
	}

	// Determine which records to display
	start := 0
	if maxRecords > 0 && len(ch.records) > maxRecords {
		start = len(ch.records) - maxRecords
	}

	result := "=== 選擇歷史 ===\n\n"

	for i := start; i < len(ch.records); i++ {
		record := ch.records[i]
		result += formatSingleRecord(&record, i+1)
		result += "\n"
	}

	return result
}

// formatSingleRecord formats a single choice record for display.
func formatSingleRecord(record *ChoiceRecord, index int) string {
	result := ""

	// Header with beat and timestamp
	result += "【Beat " + string(rune('0'+record.BeatNumber/10)) + string(rune('0'+record.BeatNumber%10)) + "】"
	if record.IsFreeText {
		result += " ✎ 自由輸入\n"
	} else {
		result += " 選項\n"
	}

	// Choice text
	result += "選擇：" + record.ChoiceText + "\n"

	// State after choice
	result += "結果：HP=" + string(rune('0'+record.HPAfter/100)) + string(rune('0'+(record.HPAfter%100)/10)) + string(rune('0'+record.HPAfter%10))
	result += " SAN=" + string(rune('0'+record.SANAfter/100)) + string(rune('0'+(record.SANAfter%100)/10)) + string(rune('0'+record.SANAfter%10))
	result += " 張力=" + string(rune('0'+record.TensionAfter/10)) + string(rune('0'+record.TensionAfter%10)) + "\n"

	// Optional consequence
	if record.Consequence != "" {
		result += "後果：" + record.Consequence + "\n"
	}

	return result
}
