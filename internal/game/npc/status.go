package npc

import (
	"time"
)

// Location represents a teammate's location in the game world
type Location struct {
	Scene            string `json:"scene"`              // Scene/room name
	DistanceToPlayer int    `json:"distance_to_player"` // 0=same scene, 1=adjacent, 2+=far away
}

// EmotionalState represents a teammate's current emotional state
type EmotionalState string

const (
	EmotionCalm     EmotionalState = "calm"
	EmotionAnxious  EmotionalState = "anxious"
	EmotionPanicked EmotionalState = "panicked"
	EmotionRelieved EmotionalState = "relieved"
	EmotionGrieving EmotionalState = "grieving"
)

// InjuryLevel represents the severity of injuries
type InjuryLevel int

const (
	InjuryNone     InjuryLevel = 0 // HP 100-70
	InjuryMinor    InjuryLevel = 1 // HP 69-30
	InjurySerious  InjuryLevel = 2 // HP 29-15
	InjuryCritical InjuryLevel = 3 // HP 14-1
)

// TeammateMessage represents a message from a separated teammate
type TeammateMessage struct {
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	ClueID    *string    `json:"clue_id,omitempty"` // If message contains a clue
}

// BehaviorModifier represents how HP affects teammate behavior
type BehaviorModifier struct {
	MoveSpeed   float64 // Movement speed multiplier (0.0-1.0)
	Reaction    float64 // Reaction speed multiplier (0.0-1.0)
	Description string  // Text description of behavior
}

// CalculateInjuryLevel determines injury level based on HP
func CalculateInjuryLevel(hp int) InjuryLevel {
	switch {
	case hp >= 70:
		return InjuryNone
	case hp >= 30:
		return InjuryMinor
	case hp >= 15:
		return InjurySerious
	default:
		return InjuryCritical
	}
}

// GetBehaviorModifier returns behavior modifiers based on HP
func GetBehaviorModifier(hp int) BehaviorModifier {
	switch {
	case hp >= 70:
		return BehaviorModifier{
			MoveSpeed:   1.0,
			Reaction:    1.0,
			Description: "",
		}
	case hp >= 30:
		return BehaviorModifier{
			MoveSpeed:   0.8,
			Reaction:    0.9,
			Description: "略顯疲憊",
		}
	case hp >= 15:
		return BehaviorModifier{
			MoveSpeed:   0.5,
			Reaction:    0.6,
			Description: "步履蹣跚，表情痛苦",
		}
	default:
		return BehaviorModifier{
			MoveSpeed:   0.0,
			Reaction:    0.3,
			Description: "無法自行移動，需要攙扶",
		}
	}
}

// Extended Teammate methods for status tracking

// UpdateLocation updates the teammate's location
func (t *Teammate) UpdateLocation(location Location) {
	t.Location = location.Scene
	t.IsSeparated = location.DistanceToPlayer > 0
	t.LastSeen = time.Now()
}

// UpdateHP updates HP and related status fields
func (t *Teammate) UpdateHP(newHP int) {
	// Clamp HP to [0, 100]
	if newHP > 100 {
		newHP = 100
	}
	if newHP < 0 {
		newHP = 0
	}

	t.HP = newHP
	t.InjuryLevel = CalculateInjuryLevel(newHP)
	t.LastSeen = time.Now() // Update last seen timestamp

	// Update status based on HP - aligned with InjuryLevel thresholds
	switch {
	case newHP == 0:
		t.Status.Alive = false
		t.Status.Conscious = false
		t.Status.Condition = "dead"
	case newHP < 15:
		t.Status.Condition = "critical"
		t.Status.Conscious = true
	case newHP < 30:
		t.Status.Condition = "injured"
		t.Status.Conscious = true
	case newHP < 70:
		t.Status.Condition = "injured" // Fixed: was "healthy", should be "injured"
		t.Status.Conscious = true
	default:
		t.Status.Condition = "healthy"
		t.Status.Conscious = true
	}
}

// UpdateEmotionalState updates the teammate's emotional state
func (t *Teammate) UpdateEmotionalState(emotion EmotionalState) {
	t.EmotionalState = emotion
}

// GetBehavior returns the current behavior modifier
func (t *Teammate) GetBehavior() BehaviorModifier {
	return GetBehaviorModifier(t.HP)
}
