package npc

import (
	"sync"
)

// DeathState represents the current state in the death progression
type DeathState string

const (
	StateSafe        DeathState = "safe"
	StateForeshadowed DeathState = "foreshadowed" // Foreshadow phase
	StateWarned      DeathState = "warned"        // Warning phase
	StateEndangered  DeathState = "endangered"    // Danger phase
	StateDead        DeathState = "dead"
	StateSaved       DeathState = "saved"
)

// Foreshadow represents a clue hinting at upcoming danger
type Foreshadow struct {
	Type    string `json:"type"`    // "obvious" or "subtle"
	Content string `json:"content"` // The foreshadow text
	Turn    int    `json:"turn"`    // Turn number when it appeared
}

// PlayerIntervention represents player's attempt to prevent death
type PlayerIntervention struct {
	Turn      int    `json:"turn"`
	Action    string `json:"action"`
	Success   bool   `json:"success"`
	Rationale string `json:"rationale"` // Why it succeeded/failed
}

// DeathEvent tracks the progression of a potential teammate death
type DeathEvent struct {
	TeammateID     string              `json:"teammate_id"`
	CurrentState   DeathState          `json:"current_state"`
	ForeshadowTurn int                 `json:"foreshadow_turn"` // Turn when foreshadow appears
	WarningTurn    int                 `json:"warning_turn"`    // Turn when warning appears
	DeadlineTurn   int                 `json:"deadline_turn"`   // Turn when death occurs (if no intervention)
	Foreshadows    []Foreshadow        `json:"foreshadows"`
	Intervention   *PlayerIntervention `json:"intervention,omitempty"`
}

// DeathManager manages death events for all teammates
type DeathManager struct {
	mu     sync.RWMutex
	events map[string]DeathEvent // TeammateID -> DeathEvent
}

// NewDeathManager creates a new death manager
func NewDeathManager() *DeathManager {
	return &DeathManager{
		events: make(map[string]DeathEvent),
	}
}

// AddEvent adds a new death event
func (dm *DeathManager) AddEvent(event DeathEvent) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.events[event.TeammateID] = event
}

// GetEvent retrieves a death event by teammate ID
func (dm *DeathManager) GetEvent(teammateID string) (DeathEvent, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	event, exists := dm.events[teammateID]
	return event, exists
}

// UpdateState updates the state of a death event
func (dm *DeathManager) UpdateState(teammateID string, newState DeathState) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if event, exists := dm.events[teammateID]; exists {
		event.CurrentState = newState
		dm.events[teammateID] = event
	}
}

// AddForeshadow adds a foreshadow to a death event
func (dm *DeathManager) AddForeshadow(teammateID string, foreshadow Foreshadow) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if event, exists := dm.events[teammateID]; exists {
		event.Foreshadows = append(event.Foreshadows, foreshadow)
		dm.events[teammateID] = event
	}
}

// CheckStateTransition checks and updates state based on current turn
// Handles multiple state transitions if turns are skipped
func (dm *DeathManager) CheckStateTransition(teammateID string, currentTurn int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	event, exists := dm.events[teammateID]
	if !exists {
		return
	}

	// Don't transition if already dead or saved
	if event.CurrentState == StateDead || event.CurrentState == StateSaved {
		return
	}

	// State machine transitions - check all states in order to handle skipped turns
	if event.CurrentState == StateSafe && currentTurn >= event.ForeshadowTurn {
		event.CurrentState = StateForeshadowed
	}
	if event.CurrentState == StateForeshadowed && currentTurn >= event.WarningTurn {
		event.CurrentState = StateWarned
	}
	if event.CurrentState == StateWarned && currentTurn >= event.DeadlineTurn {
		event.CurrentState = StateEndangered
	}

	dm.events[teammateID] = event
}

// RecordIntervention records a player intervention attempt
func (dm *DeathManager) RecordIntervention(teammateID string, intervention PlayerIntervention) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if event, exists := dm.events[teammateID]; exists {
		event.Intervention = &intervention
		if intervention.Success {
			event.CurrentState = StateSaved
		}
		dm.events[teammateID] = event
	}
}

// GetAllEvents returns all active death events
func (dm *DeathManager) GetAllEvents() []DeathEvent {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	events := make([]DeathEvent, 0, len(dm.events))
	for _, event := range dm.events {
		events = append(events, event)
	}
	return events
}

// RemoveEvent removes a death event (e.g., after death or save)
func (dm *DeathManager) RemoveEvent(teammateID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.events, teammateID)
}

// CalculateDeathSANLoss calculates SAN loss from teammate death
// intimacy: 0-100, representing relationship strength
// Returns: SAN loss in range [15, 25]
func CalculateDeathSANLoss(intimacy int) int {
	// Clamp intimacy to valid range [0, 100]
	if intimacy < 0 {
		intimacy = 0
	}
	if intimacy > 100 {
		intimacy = 100
	}

	baseLoss := 20
	intimacyModifier := intimacy / 10 // 0-10 additional loss
	totalLoss := baseLoss + intimacyModifier

	// Clamp to [15, 25]
	if totalLoss < 15 {
		return 15
	}
	if totalLoss > 25 {
		return 25
	}
	return totalLoss
}
