package game

import (
	"errors"
	"sync"
	"time"
)

// SanityState represents the psychological state based on SAN value.
type SanityState int

const (
	SanityClearHeaded SanityState = iota // 80-100
	SanityAnxious                        // 50-79
	SanityPanicked                       // 20-49
	SanityInsanity                       // 0-19
)

// String returns the string representation of sanity state.
func (s SanityState) String() string {
	switch s {
	case SanityClearHeaded:
		return "Clear-headed"
	case SanityAnxious:
		return "Anxious"
	case SanityPanicked:
		return "Panicked"
	case SanityInsanity:
		return "Insanity"
	default:
		return "Unknown"
	}
}

// PlayerStats tracks the player's HP and SAN values.
type PlayerStats struct {
	HP      int
	SAN     int
	MaxHP   int
	MaxSAN  int
	State   SanityState
	History []StatChange
	mu      sync.RWMutex
}

// NewPlayerStats creates a new player stats with initial values.
func NewPlayerStats() *PlayerStats {
	return &PlayerStats{
		HP:      100,
		SAN:     100,
		MaxHP:   100,
		MaxSAN:  100,
		State:   SanityClearHeaded,
		History: make([]StatChange, 0),
	}
}

// GetSanityState returns the current sanity state based on SAN value.
func (p *PlayerStats) GetSanityState() SanityState {
	if p.SAN >= 80 {
		return SanityClearHeaded
	} else if p.SAN >= 50 {
		return SanityAnxious
	} else if p.SAN >= 20 {
		return SanityPanicked
	}
	return SanityInsanity
}

// IsDead returns true if HP is 0.
func (p *PlayerStats) IsDead() bool {
	return p.HP <= 0
}

// IsInsane returns true if in insanity state.
func (p *PlayerStats) IsInsane() bool {
	return p.SAN < 20
}

// StatChange represents a single stat modification.
type StatChange struct {
	Timestamp time.Time
	StatType  string // "HP" or "SAN"
	Delta     int
	NewValue  int
	Reason    string
}

// NewStatChange creates a new stat change record.
func NewStatChange(statType string, delta, newValue int, reason string) StatChange {
	return StatChange{
		Timestamp: time.Now(),
		StatType:  statType,
		Delta:     delta,
		NewValue:  newValue,
		Reason:    reason,
	}
}

// IsSignificant returns true if the change is > 10 points.
func (s *StatChange) IsSignificant() bool {
	if s.Delta < 0 {
		return s.Delta <= -11
	}
	return s.Delta >= 11
}

// StatsManager manages player stats and applies difficulty modifiers.
type StatsManager struct {
	stats      *PlayerStats
	difficulty DifficultyLevel
	mu         sync.Mutex
}

// Error definitions
var (
	ErrInvalidStatType = errors.New("無效的數值類型")
)

// NewStatsManager creates a new stats manager.
func NewStatsManager(config *GameConfig) *StatsManager {
	return &StatsManager{
		stats:      NewPlayerStats(),
		difficulty: config.Difficulty,
	}
}

// GetStats returns a copy of current stats.
func (sm *StatsManager) GetStats() *PlayerStats {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Return a copy to prevent race conditions
	statsCopy := *sm.stats
	statsCopy.History = make([]StatChange, len(sm.stats.History))
	copy(statsCopy.History, sm.stats.History)

	return &statsCopy
}

// ApplyDelta applies a stat change with difficulty multiplier.
func (sm *StatsManager) ApplyDelta(statType string, delta int, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Apply difficulty multiplier
	multipliedDelta := sm.applyMultiplier(statType, delta)

	// Apply the change
	var newValue int
	switch statType {
	case "HP":
		sm.stats.HP += multipliedDelta
		// Clamp to bounds
		if sm.stats.HP < 0 {
			sm.stats.HP = 0
		}
		if sm.stats.HP > sm.stats.MaxHP {
			sm.stats.HP = sm.stats.MaxHP
		}
		newValue = sm.stats.HP

	case "SAN":
		sm.stats.SAN += multipliedDelta
		// Clamp to bounds
		if sm.stats.SAN < 0 {
			sm.stats.SAN = 0
		}
		if sm.stats.SAN > sm.stats.MaxSAN {
			sm.stats.SAN = sm.stats.MaxSAN
		}
		newValue = sm.stats.SAN
		// Update sanity state
		sm.stats.State = sm.stats.GetSanityState()

	default:
		return ErrInvalidStatType
	}

	// Record change
	change := NewStatChange(statType, multipliedDelta, newValue, reason)
	sm.stats.History = append(sm.stats.History, change)

	return nil
}

// applyMultiplier applies difficulty multiplier to delta.
func (sm *StatsManager) applyMultiplier(statType string, delta int) int {
	// Only apply multipliers to negative changes (damage/drain)
	if delta >= 0 {
		return delta
	}

	var multiplier float64
	switch statType {
	case "HP":
		switch sm.difficulty {
		case DifficultyEasy:
			multiplier = 0.5
		case DifficultyHard:
			multiplier = 1.0
		case DifficultyHell:
			multiplier = 1.5
		default:
			multiplier = 1.0
		}
	case "SAN":
		switch sm.difficulty {
		case DifficultyEasy:
			multiplier = 0.7
		case DifficultyHard:
			multiplier = 1.0
		case DifficultyHell:
			multiplier = 1.3
		default:
			multiplier = 1.0
		}
	default:
		multiplier = 1.0
	}

	return int(float64(delta) * multiplier)
}

// Reset resets stats to initial values.
func (sm *StatsManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.stats = NewPlayerStats()
}

// GetHistory returns a copy of the stat change history.
func (sm *StatsManager) GetHistory() []StatChange {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	history := make([]StatChange, len(sm.stats.History))
	copy(history, sm.stats.History)
	return history
}
