// Package game provides death-related types and logic for Nightmare Assault.
package game

import (
	"time"
)

// DeathType categorizes how the player died.
type DeathType int

const (
	// DeathTypeHP is death from HP reaching 0.
	DeathTypeHP DeathType = iota
	// DeathTypeSAN is death from SAN reaching 0 (insanity).
	DeathTypeSAN
	// DeathTypeRule is instant death from violating a hidden rule.
	DeathTypeRule
)

// String returns the display name of the death type.
func (d DeathType) String() string {
	switch d {
	case DeathTypeHP:
		return "體力耗盡"
	case DeathTypeSAN:
		return "理智崩潰"
	case DeathTypeRule:
		return "規則懲罰"
	default:
		return "未知死因"
	}
}

// EnglishName returns the English name for the death type.
func (d DeathType) EnglishName() string {
	switch d {
	case DeathTypeHP:
		return "HP Depleted"
	case DeathTypeSAN:
		return "Sanity Collapse"
	case DeathTypeRule:
		return "Rule Violation"
	default:
		return "Unknown"
	}
}

// DeathInfo contains all information about a player death.
type DeathInfo struct {
	// Type of death
	Type DeathType `json:"type"`
	// Chapter where death occurred
	Chapter int `json:"chapter"`
	// Timestamp of death
	Timestamp time.Time `json:"timestamp"`
	// FinalHP at time of death
	FinalHP int `json:"final_hp"`
	// FinalSAN at time of death
	FinalSAN int `json:"final_san"`
	// TriggeringRuleID if death was caused by a rule (empty otherwise)
	TriggeringRuleID string `json:"triggering_rule_id,omitempty"`
	// LastAction is the player action that triggered death
	LastAction string `json:"last_action"`
	// Location where death occurred
	Location string `json:"location"`
	// Narrative generated for this death
	Narrative string `json:"narrative,omitempty"`
}

// NewDeathInfo creates a new DeathInfo with timestamp.
func NewDeathInfo(deathType DeathType) *DeathInfo {
	return &DeathInfo{
		Type:      deathType,
		Timestamp: time.Now(),
	}
}

// IsInsanity returns true if this was a sanity-based death.
func (d *DeathInfo) IsInsanity() bool {
	return d.Type == DeathTypeSAN
}

// IsRuleViolation returns true if this was a rule-based death.
func (d *DeathInfo) IsRuleViolation() bool {
	return d.Type == DeathTypeRule
}

// DeathState tracks death-related state for a game session.
type DeathState struct {
	// Deaths is a list of all deaths in this session (for multiple attempts)
	Deaths []*DeathInfo `json:"deaths"`
	// CurrentDeath is the most recent death (nil if alive)
	CurrentDeath *DeathInfo `json:"current_death,omitempty"`
}

// NewDeathState creates an empty death state.
func NewDeathState() *DeathState {
	return &DeathState{
		Deaths: make([]*DeathInfo, 0),
	}
}

// RecordDeath records a new death and sets it as current.
func (s *DeathState) RecordDeath(death *DeathInfo) {
	s.Deaths = append(s.Deaths, death)
	s.CurrentDeath = death
}

// GetDeathCount returns the total number of deaths.
func (s *DeathState) GetDeathCount() int {
	return len(s.Deaths)
}

// GetLastDeath returns the most recent death (or nil).
func (s *DeathState) GetLastDeath() *DeathInfo {
	if len(s.Deaths) == 0 {
		return nil
	}
	return s.Deaths[len(s.Deaths)-1]
}

// ClearCurrentDeath clears the current death (after restart/continue).
func (s *DeathState) ClearCurrentDeath() {
	s.CurrentDeath = nil
}
