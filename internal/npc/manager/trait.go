package manager

import "encoding/json"

// ==========================================================================
// Story 1.6: Trait Structure and Revelation Logic
// ==========================================================================

// TraitStatus represents the current revelation status of a trait.
// Story 8.1: Traits progress through four stages: Hidden -> HintPhase1 -> HintPhase2 -> Revealed
type TraitStatus int

const (
	// Hidden means the trait is completely hidden from the player
	Hidden TraitStatus = iota
	// HintPhase1 means the trait is being subtly revealed through initial hints
	HintPhase1
	// HintPhase2 means the trait is being revealed through more explicit hints
	HintPhase2
	// Revealed means the trait is fully revealed to the player
	Revealed

	// Legacy compatibility: Hinting maps to HintPhase1 for backward compatibility
	Hinting = HintPhase1
)

// String returns the string representation of TraitStatus.
func (ts TraitStatus) String() string {
	switch ts {
	case Hidden:
		return "hidden"
	case HintPhase1:
		return "hint_phase_1"
	case HintPhase2:
		return "hint_phase_2"
	case Revealed:
		return "revealed"
	default:
		return "hidden"
	}
}

// MarshalJSON implements json.Marshaler for TraitStatus.
func (ts TraitStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.String())
}

// UnmarshalJSON implements json.Unmarshaler for TraitStatus.
func (ts *TraitStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "hidden":
		*ts = Hidden
	case "hinting", "hint_phase_1": // Legacy "hinting" maps to HintPhase1
		*ts = HintPhase1
	case "hint_phase_2":
		*ts = HintPhase2
	case "revealed":
		*ts = Revealed
	default:
		*ts = Hidden
	}
	return nil
}

// TriggerType represents the type of condition that can trigger trait revelation.
type TriggerType int

const (
	// TrustLevel triggers based on trust emotional state
	TrustLevel TriggerType = iota
	// FearLevel triggers based on fear emotional state
	FearLevel
	// StressLevel triggers based on stress emotional state
	StressLevel
	// Event triggers based on specific game events
	Event
	// InteractionCount triggers based on number of interactions
	InteractionCount
	// TimeBased triggers based on game time/beats
	TimeBased
)

// String returns the string representation of TriggerType.
func (tt TriggerType) String() string {
	switch tt {
	case TrustLevel:
		return "trust_level"
	case FearLevel:
		return "fear_level"
	case StressLevel:
		return "stress_level"
	case Event:
		return "event"
	case InteractionCount:
		return "interaction_count"
	case TimeBased:
		return "time_based"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for TriggerType.
func (tt TriggerType) MarshalJSON() ([]byte, error) {
	return json.Marshal(tt.String())
}

// UnmarshalJSON implements json.Unmarshaler for TriggerType.
func (tt *TriggerType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "trust_level":
		*tt = TrustLevel
	case "fear_level":
		*tt = FearLevel
	case "stress_level":
		*tt = StressLevel
	case "event":
		*tt = Event
	case "interaction_count":
		*tt = InteractionCount
	case "time_based":
		*tt = TimeBased
	default:
		*tt = TrustLevel
	}
	return nil
}

// TraitTrigger defines a condition that must be met for trait revelation.
// Multiple triggers in a trait use AND logic (all must be satisfied).
type TraitTrigger struct {
	Type       TriggerType `json:"type"`       // Type of trigger
	Threshold  int         `json:"threshold"`  // Threshold value for numeric triggers
	EventName  string      `json:"event_name"` // Event name for event-based triggers
	Comparator string      `json:"comparator"` // Comparison operator: ">=", "<=", "==", ">", "<"
}

// TraitFull represents the complete trait structure with revelation logic.
// This extends the basic Trait structure from profile.go.
// Story 8.1: Enhanced with multi-phase hints for progressive revelation
type TraitFull struct {
	ID         string         `json:"id"`          // Unique identifier
	Content    string         `json:"content"`     // Full trait description
	RevealTier int            `json:"reveal_tier"` // Reveal difficulty (1=easy, 2=medium, 3=hard)
	Triggers   []TraitTrigger `json:"triggers"`    // Conditions for revelation
	Status     TraitStatus    `json:"status"`      // Current revelation status

	// Story 8.1: Phase-specific hints for progressive revelation
	// Hints shown during legacy Hinting phase (maps to HintPhase1 for compatibility)
	Hints       []string `json:"hints"`
	// HintsPhase1: Subtle, indirect hints (e.g., "似乎對某事感到不安")
	HintsPhase1 []string `json:"hints_phase_1"`
	// HintsPhase2: More explicit hints (e.g., "經常提及過去的失敗經驗")
	HintsPhase2 []string `json:"hints_phase_2"`

	// Story 8.1: Interaction tracking for phase transition
	InteractionCount int `json:"interaction_count"` // Number of relevant interactions
	LastRevealCheck  int `json:"last_reveal_check"` // Last game beat when reveal was checked
}

// ToBasicTrait converts TraitFull to the basic Trait structure.
// This allows compatibility with the existing profile.go Trait type.
func (tf TraitFull) ToBasicTrait() Trait {
	return Trait{
		ID:         tf.ID,
		Content:    tf.Content,
		RevealTier: tf.RevealTier,
	}
}

// FromBasicTrait creates a TraitFull from a basic Trait.
// Additional fields will need to be populated separately.
func FromBasicTrait(t Trait) TraitFull {
	return TraitFull{
		ID:         t.ID,
		Content:    t.Content,
		RevealTier: t.RevealTier,
		Triggers:   []TraitTrigger{},
		Status:     Hidden,
		Hints:      []string{},
	}
}

// RevealContext provides the context needed to evaluate trait revelation triggers.
type RevealContext struct {
	CurrentBeat      int      `json:"current_beat"`      // Current game beat/time
	RecentEvents     []string `json:"recent_events"`     // Recent game events
	InteractionCount int      `json:"interaction_count"` // Total number of interactions with this NPC
}
