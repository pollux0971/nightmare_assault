package manager

import (
	"encoding/json"
	"time"
)

// MentalState represents the mental/psychological state of an NPC.
type MentalState int

const (
	// Normal represents a stable, rational mental state
	Normal MentalState = iota
	// Anxious represents a heightened state of worry and unease
	Anxious
	// Corrupted represents a state of mental breakdown or supernatural influence
	Corrupted
)

// String returns a string representation of the MentalState.
func (m MentalState) String() string {
	switch m {
	case Normal:
		return "normal"
	case Anxious:
		return "anxious"
	case Corrupted:
		return "corrupted"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for MentalState.
func (m MentalState) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON implements json.Unmarshaler for MentalState.
func (m *MentalState) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "normal":
		*m = Normal
	case "anxious":
		*m = Anxious
	case "corrupted":
		*m = Corrupted
	default:
		*m = Normal
	}
	return nil
}

// RelationshipType represents the type of relationship between player and NPC.
type RelationshipType int

const (
	// Neutral represents a balanced, uncommitted relationship
	Neutral RelationshipType = iota
	// Friendly represents a positive, trusting relationship
	Friendly
	// Hostile represents an antagonistic, distrustful relationship
	Hostile
	// Fearful represents a relationship dominated by fear
	Fearful
)

// String returns a string representation of the RelationshipType.
func (r RelationshipType) String() string {
	switch r {
	case Friendly:
		return "friendly"
	case Neutral:
		return "neutral"
	case Hostile:
		return "hostile"
	case Fearful:
		return "fearful"
	default:
		return "neutral"
	}
}

// MarshalJSON implements json.Marshaler for RelationshipType.
func (r RelationshipType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler for RelationshipType.
func (r *RelationshipType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "friendly":
		*r = Friendly
	case "neutral":
		*r = Neutral
	case "hostile":
		*r = Hostile
	case "fearful":
		*r = Fearful
	default:
		*r = Neutral
	}
	return nil
}

// CalculateRelationship determines the relationship type based on emotional state.
// Logic (per Story 1.4 requirements):
// - Trust >= 60 && Fear < 40 -> Friendly (highest priority)
// - Fear >= 60 -> Fearful (fear overrides other states)
// - Trust < 30 && Fear < 40 -> Hostile (low trust + low fear = hostile)
// - Otherwise -> Neutral (default)
func CalculateRelationship(emotion EmotionState) RelationshipType {
	if emotion.Trust >= 60 && emotion.Fear < 40 {
		return Friendly
	}
	if emotion.Fear >= 60 {
		return Fearful
	}
	if emotion.Trust < 30 && emotion.Fear < 40 {
		return Hostile
	}
	return Neutral
}

// CalculateRelationshipScore calculates a numerical relationship score.
// Formula: (Trust - 50) - (Fear / 2)
// Range: -100 (extremely hostile) to +100 (extremely friendly)
func CalculateRelationshipScore(emotion EmotionState) int {
	score := (emotion.Trust - 50) - (emotion.Fear / 2)
	if score < -100 {
		return -100
	}
	if score > 100 {
		return 100
	}
	return score
}

// Note: TraitStatus is now defined in trait.go as part of Story 1.6 implementation.
// It is an enum type (int-based) with values: Hidden, Hinting, Revealed

// NPCInteraction records a single interaction between player and NPC.
type NPCInteraction struct {
	Timestamp       time.Time     `json:"timestamp"`        // When the interaction occurred
	InteractionType string        `json:"interaction_type"` // Type of interaction (e.g., "dialogue", "help", "threaten")
	EmotionDelta    EmotionDelta  `json:"emotion_delta"`    // How this interaction affected emotions
	Description     string        `json:"description"`      // Human-readable description of the interaction
}

// NPCRuntimeState represents the mutable runtime state of an NPC during gameplay.
// This is separate from the immutable NPCProfile.
type NPCRuntimeState struct {
	Emotion          EmotionState              `json:"emotion"`           // Current emotional state
	MentalState      MentalState               `json:"mental_state"`      // Current mental/psychological state
	IsAlive          bool                      `json:"is_alive"`          // Whether the NPC is still alive
	Relationship     RelationshipType          `json:"relationship"`      // Current relationship with player
	RelationshipScore int                      `json:"relationship_score"` // Numerical relationship score for fine-grained tracking
	TraitStates      map[string]string         `json:"trait_states"`      // Current status of personality traits (stored as string: "hidden", "hinting", "revealed")
	Interactions     []NPCInteraction          `json:"interactions"`      // History of all interactions
	LastInteraction  time.Time                 `json:"last_interaction"`  // Timestamp of most recent interaction
}

// NewNPCRuntimeState creates a new runtime state with default values.
func NewNPCRuntimeState() *NPCRuntimeState {
	return &NPCRuntimeState{
		Emotion:          DefaultEmotionState(),
		MentalState:      Normal,
		IsAlive:          true,
		Relationship:     Neutral,
		RelationshipScore: 0,
		TraitStates:      make(map[string]string),
		Interactions:     []NPCInteraction{},
		LastInteraction:  time.Time{}, // Zero time
	}
}

// AddInteraction adds a new interaction to the NPC's history and updates emotional state.
// It automatically updates the relationship type based on the new emotional state.
func (s *NPCRuntimeState) AddInteraction(interactionType, description string, delta EmotionDelta) {
	// Create interaction record
	interaction := NPCInteraction{
		Timestamp:       time.Now(),
		InteractionType: interactionType,
		EmotionDelta:    delta,
		Description:     description,
	}

	// Add to history
	s.Interactions = append(s.Interactions, interaction)
	s.LastInteraction = interaction.Timestamp

	// Update emotional state
	s.Emotion = s.Emotion.Apply(delta)

	// Recalculate relationship type and score
	s.Relationship = CalculateRelationship(s.Emotion)
	s.RelationshipScore = CalculateRelationshipScore(s.Emotion)
}

// GetRecentInteractions returns the n most recent interactions.
// If n is greater than the total number of interactions, returns all interactions.
// If n <= 0, returns an empty slice.
func (s *NPCRuntimeState) GetRecentInteractions(n int) []NPCInteraction {
	if n <= 0 {
		return []NPCInteraction{}
	}

	total := len(s.Interactions)
	if n >= total {
		// Return all interactions
		result := make([]NPCInteraction, total)
		copy(result, s.Interactions)
		return result
	}

	// Return the last n interactions
	result := make([]NPCInteraction, n)
	copy(result, s.Interactions[total-n:])
	return result
}

// UpdateMentalState updates the mental state based on current emotional conditions.
// This method delegates to checkMentalStateTransition() for the actual transition logic.
// Per Story 1.5 requirements.
func (s *NPCRuntimeState) UpdateMentalState() {
	s.MentalState = checkMentalStateTransition(s.MentalState, s.Emotion)
}

// Copy creates a deep copy of the NPCRuntimeState.
func (s *NPCRuntimeState) Copy() *NPCRuntimeState {
	// Copy trait states
	traitsCopy := make(map[string]string, len(s.TraitStates))
	for k, v := range s.TraitStates {
		traitsCopy[k] = v
	}

	// Copy interactions
	interactionsCopy := make([]NPCInteraction, len(s.Interactions))
	copy(interactionsCopy, s.Interactions)

	return &NPCRuntimeState{
		Emotion:           s.Emotion.Copy(),
		MentalState:       s.MentalState,
		IsAlive:           s.IsAlive,
		Relationship:      s.Relationship,
		RelationshipScore: s.RelationshipScore,
		TraitStates:       traitsCopy,
		Interactions:      interactionsCopy,
		LastInteraction:   s.LastInteraction,
	}
}

// String returns a JSON string representation of the runtime state.
func (s *NPCRuntimeState) String() string {
	data, _ := json.Marshal(s)
	return string(data)
}
