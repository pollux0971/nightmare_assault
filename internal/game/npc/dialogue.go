package npc

import (
	"sync"
	"time"
)

// DialogueEntry represents a single dialogue exchange
type DialogueEntry struct {
	Timestamp    time.Time       `json:"timestamp"`
	Speaker      string          `json:"speaker"` // "Player" or Teammate.Name
	Content      string          `json:"content"`
	ClueRevealed *ClueRevelation `json:"clue_revealed,omitempty"`
}

// ClueRevelation represents a clue revealed during dialogue
type ClueRevelation struct {
	ClueID         string `json:"clue_id"`
	RevelationType string `json:"revelation_type"` // "hint", "worry", "observation", "memory"
	Subtlety       int    `json:"subtlety"`        // 1-10, higher = more subtle
}

// DialogueSystem manages dialogue between player and teammates
type DialogueSystem struct {
	mu          sync.RWMutex
	dialogueLog []DialogueEntry
	teammates   []*Teammate
}

// NewDialogueSystem creates a new dialogue system
func NewDialogueSystem() *DialogueSystem {
	return &DialogueSystem{
		dialogueLog: []DialogueEntry{},
		teammates:   []*Teammate{},
	}
}

// AddEntry adds a new dialogue entry to the log
func (ds *DialogueSystem) AddEntry(entry DialogueEntry) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dialogueLog = append(ds.dialogueLog, entry)
}

// GetHistory returns the full dialogue history
func (ds *DialogueSystem) GetHistory() []DialogueEntry {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Return a deep copy to prevent external modification and race conditions
	history := make([]DialogueEntry, len(ds.dialogueLog))
	for i, entry := range ds.dialogueLog {
		history[i] = entry
		// Deep copy ClueRevealed pointer if exists
		if entry.ClueRevealed != nil {
			clueCopy := *entry.ClueRevealed
			history[i].ClueRevealed = &clueCopy
		}
	}
	return history
}

// GetRecentHistory returns the last N dialogue entries
func (ds *DialogueSystem) GetRecentHistory(count int) []DialogueEntry {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Validate count
	if count <= 0 {
		return []DialogueEntry{}
	}

	if count >= len(ds.dialogueLog) {
		count = len(ds.dialogueLog)
	}

	start := len(ds.dialogueLog) - count
	history := make([]DialogueEntry, count)

	// Deep copy entries
	for i := 0; i < count; i++ {
		entry := ds.dialogueLog[start+i]
		history[i] = entry
		// Deep copy ClueRevealed pointer if exists
		if entry.ClueRevealed != nil {
			clueCopy := *entry.ClueRevealed
			history[i].ClueRevealed = &clueCopy
		}
	}
	return history
}

// ClearHistory clears all dialogue history
func (ds *DialogueSystem) ClearHistory() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dialogueLog = []DialogueEntry{}
}

// SetTeammates sets the teammates for this dialogue system
func (ds *DialogueSystem) SetTeammates(teammates []*Teammate) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.teammates = teammates
}

// GetTeammates returns the current teammates
func (ds *DialogueSystem) GetTeammates() []*Teammate {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	teammates := make([]*Teammate, len(ds.teammates))
	copy(teammates, ds.teammates)
	return teammates
}

// BuildDialoguePrompt creates a prompt for teammate dialogue generation
func BuildDialoguePrompt(teammate *Teammate, playerMessage string, recentHistory []DialogueEntry) string {
	// This will be implemented when integrating with LLM
	return ""
}

// GenerateTeammateResponse generates a response from a teammate (placeholder)
func (ds *DialogueSystem) GenerateTeammateResponse(teammate *Teammate, playerMessage string) (string, *ClueRevelation, error) {
	// This will be implemented when integrating with LLM provider
	return "", nil, nil
}
