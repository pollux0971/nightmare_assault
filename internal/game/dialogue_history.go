// Package game provides game-related types and logic for Nightmare Assault.
package game

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ==========================================================================
// Story 7.7: NPC Dialogue History System
// ==========================================================================

// DialogueRecord represents a single NPC dialogue entry.
//
// Story 7.7 AC #6: NPC Dialogue Recording
//   - Store NPC name, dialogue content, current beat
//   - Support history review via /log command
type DialogueRecord struct {
	// Dialogue metadata
	NPCName    string    `json:"npc_name"`    // Name of the NPC
	NPCID      string    `json:"npc_id"`      // ID of the NPC
	Dialogue   string    `json:"dialogue"`    // The dialogue content
	Timestamp  time.Time `json:"timestamp"`   // When dialogue occurred
	BeatNumber int       `json:"beat_number"` // Beat number when dialogue occurred

	// Context
	Scene      string `json:"scene"`       // Scene where dialogue occurred
	Tension    int    `json:"tension"`     // Tension level at time of dialogue
	SAN        int    `json:"san"`         // Player SAN at time of dialogue
	IsQuestion bool   `json:"is_question"` // Whether this was a response to player question

	// Additional metadata
	ClueRevealed   bool   `json:"clue_revealed"`    // Whether a clue was revealed
	SeedID         string `json:"seed_id,omitempty"` // Global seed ID if clue revealed
	IsDeathDialogue bool  `json:"is_death_dialogue"` // Whether this is death dialogue
}

// DialogueHistory manages the history of NPC dialogues.
//
// Story 7.7 AC #6: Records and retrieves dialogue history for /log command.
// Thread-safe implementation for concurrent access.
type DialogueHistory struct {
	records []DialogueRecord
	mu      sync.RWMutex
}

// NewDialogueHistory creates a new dialogue history manager.
func NewDialogueHistory() *DialogueHistory {
	return &DialogueHistory{
		records: make([]DialogueRecord, 0),
	}
}

// RecordDialogue records a new NPC dialogue to history.
//
// Story 7.7 AC #6: Save NPC name, dialogue content, and current beat.
func (dh *DialogueHistory) RecordDialogue(record DialogueRecord) {
	dh.mu.Lock()
	defer dh.mu.Unlock()

	// Set timestamp if not already set
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	dh.records = append(dh.records, record)
}

// GetAllRecords returns all dialogue records in chronological order.
// Returns a copy to prevent external modification.
func (dh *DialogueHistory) GetAllRecords() []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	recordsCopy := make([]DialogueRecord, len(dh.records))
	copy(recordsCopy, dh.records)
	return recordsCopy
}

// GetRecordsByBeatRange returns dialogues within a beat range [start, end] inclusive.
func (dh *DialogueHistory) GetRecordsByBeatRange(startBeat, endBeat int) []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	filtered := make([]DialogueRecord, 0)
	for _, record := range dh.records {
		if record.BeatNumber >= startBeat && record.BeatNumber <= endBeat {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetRecordsByNPC returns all dialogues by a specific NPC.
func (dh *DialogueHistory) GetRecordsByNPC(npcID string) []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	filtered := make([]DialogueRecord, 0)
	for _, record := range dh.records {
		if record.NPCID == npcID {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetRecordsByScene returns all dialogues in a specific scene.
func (dh *DialogueHistory) GetRecordsByScene(scene string) []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	filtered := make([]DialogueRecord, 0)
	for _, record := range dh.records {
		if record.Scene == scene {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetRecent returns the N most recent dialogues.
func (dh *DialogueHistory) GetRecent(n int) []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	if n <= 0 {
		return []DialogueRecord{}
	}

	start := len(dh.records) - n
	if start < 0 {
		start = 0
	}

	recent := make([]DialogueRecord, len(dh.records)-start)
	copy(recent, dh.records[start:])
	return recent
}

// GetLastRecord returns the most recent dialogue record.
// Returns nil if no records exist.
func (dh *DialogueHistory) GetLastRecord() *DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	if len(dh.records) == 0 {
		return nil
	}

	lastRecord := dh.records[len(dh.records)-1]
	return &lastRecord
}

// GetRecordCount returns the total number of recorded dialogues.
func (dh *DialogueHistory) GetRecordCount() int {
	dh.mu.RLock()
	defer dh.mu.RUnlock()
	return len(dh.records)
}

// GetClueRecords returns all dialogues that revealed clues.
func (dh *DialogueHistory) GetClueRecords() []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	filtered := make([]DialogueRecord, 0)
	for _, record := range dh.records {
		if record.ClueRevealed {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// GetDeathDialogues returns all death dialogues.
func (dh *DialogueHistory) GetDeathDialogues() []DialogueRecord {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	filtered := make([]DialogueRecord, 0)
	for _, record := range dh.records {
		if record.IsDeathDialogue {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// Clear clears all dialogue records.
// Useful for starting a new game.
func (dh *DialogueHistory) Clear() {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	dh.records = make([]DialogueRecord, 0)
}

// GetStatsSummary returns a summary of dialogues and clue revelations.
func (dh *DialogueHistory) GetStatsSummary() *DialogueStatsSummary {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	summary := &DialogueStatsSummary{
		TotalDialogues:    len(dh.records),
		ClueRevelations:   0,
		DeathDialogues:    0,
		QuestionResponses: 0,
	}

	// Count by NPC
	summary.DialoguesByNPC = make(map[string]int)

	for _, record := range dh.records {
		if record.ClueRevealed {
			summary.ClueRevelations++
		}
		if record.IsDeathDialogue {
			summary.DeathDialogues++
		}
		if record.IsQuestion {
			summary.QuestionResponses++
		}

		// Count by NPC name
		summary.DialoguesByNPC[record.NPCName]++
	}

	return summary
}

// DialogueStatsSummary provides statistics about NPC dialogues.
type DialogueStatsSummary struct {
	TotalDialogues    int            `json:"total_dialogues"`
	ClueRevelations   int            `json:"clue_revelations"`
	DeathDialogues    int            `json:"death_dialogues"`
	QuestionResponses int            `json:"question_responses"`
	DialoguesByNPC    map[string]int `json:"dialogues_by_npc"`
}

// ToJSON serializes the dialogue history to JSON.
func (dh *DialogueHistory) ToJSON() ([]byte, error) {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	return json.MarshalIndent(dh.records, "", "  ")
}

// FromJSON deserializes the dialogue history from JSON.
func (dh *DialogueHistory) FromJSON(data []byte) error {
	dh.mu.Lock()
	defer dh.mu.Unlock()

	var records []DialogueRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	dh.records = records
	return nil
}

// Clone creates a deep copy of the dialogue history.
func (dh *DialogueHistory) Clone() *DialogueHistory {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	clone := NewDialogueHistory()
	clone.records = make([]DialogueRecord, len(dh.records))
	copy(clone.records, dh.records)

	return clone
}

// FormatForDisplay formats the dialogue history for display in /log command.
//
// Story 7.7 AC #6: Support history review via /log command.
// Returns a formatted string with NPC name, dialogue, and beat number.
func (dh *DialogueHistory) FormatForDisplay(maxRecords int) string {
	return dh.FormatForDisplayPaged(maxRecords, 1)
}

// FormatForDisplayPaged formats the dialogue history with pagination support.
// Story 10-4 AC: 支援翻頁查看更早記錄
// page 1 = most recent records, page 2 = older records, etc.
func (dh *DialogueHistory) FormatForDisplayPaged(recordsPerPage, pageNumber int) string {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	if len(dh.records) == 0 {
		return "尚未記錄任何 NPC 對話。"
	}

	totalRecords := len(dh.records)

	// Handle zero or negative recordsPerPage by showing all records
	if recordsPerPage <= 0 {
		recordsPerPage = totalRecords
	}

	totalPages := (totalRecords + recordsPerPage - 1) / recordsPerPage

	// Validate page number
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageNumber > totalPages {
		pageNumber = totalPages
	}

	// Calculate record range for this page
	// Page 1 shows most recent, so we calculate from the end
	endIndex := totalRecords - (pageNumber-1)*recordsPerPage
	startIndex := endIndex - recordsPerPage
	if startIndex < 0 {
		startIndex = 0
	}

	result := fmt.Sprintf("=== NPC 對話歷史 (第 %d/%d 頁) ===\n\n", pageNumber, totalPages)

	for i := startIndex; i < endIndex; i++ {
		record := dh.records[i]
		result += formatSingleDialogueRecord(&record)
		result += "\n"
	}

	// Navigation hints
	if totalPages > 1 {
		result += fmt.Sprintf("───────────────────────────────────────────────────\n")
		result += fmt.Sprintf("📄 共 %d 條記錄，%d 頁\n", totalRecords, totalPages)
		if pageNumber < totalPages {
			result += fmt.Sprintf("   /log page %d 查看更早記錄\n", pageNumber+1)
		}
		if pageNumber > 1 {
			result += fmt.Sprintf("   /log page %d 查看較新記錄\n", pageNumber-1)
		}
	}

	return result
}

// formatSingleDialogueRecord formats a single dialogue record for display.
// Story 10-4 AC: 格式: [Beat N] 玩家選擇: XX → 結果: YY
func formatSingleDialogueRecord(record *DialogueRecord) string {
	var result string

	// Header with beat number (AC format)
	result = fmt.Sprintf("[Beat %d] ", record.BeatNumber)

	// Show interaction type and NPC name
	if record.IsQuestion {
		result += fmt.Sprintf("玩家提問: %s → ", record.NPCName)
	} else {
		result += fmt.Sprintf("與 %s 對話 → ", record.NPCName)
	}

	// Show result/outcome
	if record.IsDeathDialogue {
		result += "☠ 臨終遺言\n"
	} else if record.ClueRevealed {
		result += "💡 獲得線索\n"
	} else {
		result += "對話完成\n"
	}

	// Dialogue content (indented)
	result += "   " + record.Dialogue + "\n"

	// Additional clue info
	if record.ClueRevealed && record.SeedID != "" {
		result += fmt.Sprintf("   📌 線索ID: %s\n", record.SeedID)
	}

	return result
}
