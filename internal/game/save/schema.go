// Package save provides save data structures and management for Nightmare Assault.
package save

import (
	"time"
)

// CurrentVersion is the current save data format version.
const CurrentVersion = 1

// LogType represents the type of log entry.
type LogType int

const (
	// LogNarrative represents game narrative text.
	LogNarrative LogType = iota
	// LogPlayerInput represents player input.
	LogPlayerInput
	// LogOptionChoice represents option selections.
	LogOptionChoice
	// LogSystem represents system messages.
	LogSystem
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      LogType   `json:"type"`
	Content   string    `json:"content"`
}

// SaveData represents the complete save game structure.
type SaveData struct {
	Version    int             `json:"version"`
	Metadata   Metadata        `json:"metadata"`
	Player     PlayerState     `json:"player"`
	Game       GameState       `json:"game"`
	Teammates  []TeammateState `json:"teammates"`
	Context    StoryContext    `json:"context"`
	LogEntries []LogEntry      `json:"log_entries"`
	Checksum   string          `json:"checksum"`
}

// Metadata contains save file metadata.
type Metadata struct {
	SavedAt     time.Time `json:"saved_at"`
	PlayTime    int       `json:"play_time_seconds"`
	Difficulty  string    `json:"difficulty"`
	StoryLength string    `json:"story_length"`
}

// PlayerState contains the player's current state.
type PlayerState struct {
	HP         int      `json:"hp"`
	SAN        int      `json:"san"`
	Location   string   `json:"location"`
	Inventory  []Item   `json:"inventory"`
	KnownClues []string `json:"known_clues"`
}

// GameState contains the current game state.
type GameState struct {
	CurrentChapter  int          `json:"current_chapter"`
	ChapterProgress float32      `json:"chapter_progress"`
	TriggeredRules  []string     `json:"triggered_rules"`
	DiscoveredRules []string     `json:"discovered_rules"`
	HiddenRules     *RuleStorage `json:"hidden_rules,omitempty"` // AC3: Rules stored in game state
}

// RuleStorage contains rule data for save/load (hidden from player).
type RuleStorage struct {
	Rules []SavedRule `json:"rules"`
}

// SavedRule represents a rule in save data format.
type SavedRule struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	TriggerType   string `json:"trigger_type"`
	TriggerValue  string `json:"trigger_value"`
	Consequence   string `json:"consequence"`
	HPDamage      int    `json:"hp_damage,omitempty"`
	SANDamage     int    `json:"san_damage,omitempty"`
	Clues         []string `json:"clues"`
	Priority      int    `json:"priority"`
	Violations    int    `json:"violations"`
	MaxViolations int    `json:"max_violations"`
	Discovered    bool   `json:"discovered"`
	Active        bool   `json:"active"`
}

// TeammateState contains a teammate's current state.
type TeammateState struct {
	Name         string `json:"name"`
	Alive        bool   `json:"alive"`
	HP           int    `json:"hp"`
	Location     string `json:"location"`
	Items        []Item `json:"items"`
	Relationship int    `json:"relationship"`
}

// StoryContext contains the story context for AI continuity.
type StoryContext struct {
	RecentSummary string `json:"recent_summary"`
	CurrentScene  string `json:"current_scene"`
	GameBible     string `json:"game_bible_snapshot"`
}

// Item represents an inventory item.
type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewSaveData creates a new SaveData with default values.
func NewSaveData() *SaveData {
	return &SaveData{
		Version: CurrentVersion,
		Metadata: Metadata{
			SavedAt:     time.Now(),
			PlayTime:    0,
			Difficulty:  "normal",
			StoryLength: "medium",
		},
		Player: PlayerState{
			HP:         100,
			SAN:        100,
			Location:   "",
			Inventory:  []Item{},
			KnownClues: []string{},
		},
		Game: GameState{
			CurrentChapter:  1,
			ChapterProgress: 0.0,
			TriggeredRules:  []string{},
			DiscoveredRules: []string{},
		},
		Teammates:  []TeammateState{},
		LogEntries: []LogEntry{},
		Context: StoryContext{
			RecentSummary: "",
			CurrentScene:  "",
			GameBible:     "",
		},
	}
}
