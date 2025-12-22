package chat

import (
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TemplateCategory defines the type of dialogue template.
// Each category represents a different emotional/behavioral response pattern.
type TemplateCategory string

const (
	// CategoryAgree represents agreement or cooperation
	CategoryAgree TemplateCategory = "agree"

	// CategoryDisagree represents disagreement or resistance
	CategoryDisagree TemplateCategory = "disagree"

	// CategoryConfused represents confusion or disorientation
	CategoryConfused TemplateCategory = "confused"

	// CategoryFearful represents fear or anxiety
	CategoryFearful TemplateCategory = "fearful"

	// CategoryCurious represents curiosity or interest
	CategoryCurious TemplateCategory = "curious"

	// CategoryDefensive represents defensiveness or guardedness
	CategoryDefensive TemplateCategory = "defensive"

	// CategoryNeutral represents neutral or non-committal responses
	CategoryNeutral TemplateCategory = "neutral"
)

// TemplateConditions defines the conditions under which a template can be used.
// All conditions must be satisfied for the template to match.
type TemplateConditions struct {
	// MinTrust specifies the minimum trust level required (0-100)
	MinTrust *int

	// MaxTrust specifies the maximum trust level allowed (0-100)
	MaxTrust *int

	// MinFear specifies the minimum fear level required (0-100)
	MinFear *int

	// MaxFear specifies the maximum fear level allowed (0-100)
	MaxFear *int

	// MinStress specifies the minimum stress level required (0-100)
	MinStress *int

	// MaxStress specifies the maximum stress level allowed (0-100)
	MaxStress *int

	// RequiredMentalState specifies a required mental state (if set)
	RequiredMentalState *manager.MentalState
}

// DialogueTemplate represents a single dialogue template.
type DialogueTemplate struct {
	// ID is a unique identifier for this template
	ID string

	// Category is the type of response this template represents
	Category TemplateCategory

	// Archetype is the NPC archetype this template is designed for.
	// Use "Any" for universal templates that work with all archetypes.
	Archetype string

	// Content is the actual dialogue text.
	// Supports variable substitution: {npc.name}, {player.name}
	Content string

	// Conditions specify when this template can be used
	Conditions TemplateConditions
}

// FallbackContext provides context for selecting a fallback template.
type FallbackContext struct {
	// NPCID is the ID of the NPC who needs to respond
	NPCID string

	// NPCName is the display name of the NPC
	NPCName string

	// PlayerName is the display name of the player
	PlayerName string

	// Emotion is the NPC's current emotional state
	Emotion manager.EmotionState

	// MentalState is the NPC's current mental state
	MentalState manager.MentalState

	// Archetype is the NPC's archetype (e.g., "Scientist", "Guard", "Survivor")
	Archetype string

	// Flags are the chat flags that triggered this context
	Flags []string

	// HasHallucination indicates if the hallucination flag is present
	HasHallucination bool

	// HasHostile indicates if the hostile flag is present
	HasHostile bool

	// HasRevelation indicates if the revelation flag is present
	HasRevelation bool
}

// FallbackManager manages the template library and selection logic.
type FallbackManager struct {
	// templates is the complete library of all available templates
	templates []DialogueTemplate
}
