package agents

// ==========================================================================
// Story 6-6: Choice Agent Types
// ==========================================================================

// SceneType represents the type of scene for choice generation
type SceneType int

const (
	// SceneExplore indicates exploration scenario
	SceneExplore SceneType = iota
	// SceneDialogue indicates dialogue scenario
	SceneDialogue
	// SceneCombat indicates combat scenario
	SceneCombat
	// SceneEscape indicates escape scenario
	SceneEscape
)

// String returns the string representation of SceneType
func (st SceneType) String() string {
	switch st {
	case SceneExplore:
		return "explore"
	case SceneDialogue:
		return "dialogue"
	case SceneCombat:
		return "combat"
	case SceneEscape:
		return "escape"
	default:
		return "unknown"
	}
}

// RiskLevel represents the risk level of a choice option
type RiskLevel int

const (
	// RiskSafe indicates no risk
	RiskSafe RiskLevel = iota
	// RiskWarning indicates potential minor rule trigger
	RiskWarning
	// RiskDanger indicates potential moderate rule trigger
	RiskDanger
	// RiskLethal indicates potential fatal rule trigger
	RiskLethal
)

// String returns the string representation of RiskLevel
func (rl RiskLevel) String() string {
	switch rl {
	case RiskSafe:
		return "Safe"
	case RiskWarning:
		return "Warning"
	case RiskDanger:
		return "Danger"
	case RiskLethal:
		return "Lethal"
	default:
		return "Unknown"
	}
}

// ChoiceRequest is the request for generating player choice options
//
// Parameters:
//   - StoryContext: Current story narrative context
//   - SceneType: Type of scene (explore/dialogue/combat/escape)
//   - TensionLevel: Current tension level (affects choice tone)
//   - ActiveRules: Currently active hidden rules for risk assessment
//   - PlayerSAN: Player's current SAN value (affects hallucination)
//   - Difficulty: Game difficulty (easy/normal/hard/hell)
type ChoiceRequest struct {
	StoryContext string
	SceneType    SceneType
	TensionLevel int // 0-100
	ActiveRules  []HiddenRule
	PlayerSAN    int
	Difficulty   string
}

// ChoiceResponse is the response containing generated choice options
//
// Contains:
//   - Options: List of choice options (2-4 options)
//   - RiskLevels: Risk level for each option (internal use only)
//   - HallucinationFlags: Marks hallucination options (internal use only)
type ChoiceResponse struct {
	Options            []Option
	RiskLevels         map[int]RiskLevel // OptionIndex → RiskLevel
	HallucinationFlags map[int]bool      // OptionIndex → IsHallucination
}

// Option represents a single choice option for the player
//
// Design:
//   - Text must be ≤ 15 characters (Traditional Chinese)
//   - Index starts from 1 (player-facing numbering)
//   - Description is for internal use (not shown to player)
type Option struct {
	Index       int    // Option number (1-based)
	Text        string // Display text (≤ 15 chars)
	Description string // Internal description
}

// HiddenRule represents a hidden rule for risk assessment
//
// Simplified version for ChoiceAgent use.
// Full implementation is in Epic 3 (Rule System).
type HiddenRule struct {
	ID              string
	Name            string
	TriggerKeywords []string // Keywords that trigger this rule
	Punishment      RulePunishment
	DirectHint      string // Direct hint for easy difficulty
	MetaphorHint    string // Metaphor hint for normal difficulty
}

// RulePunishment represents the punishment for rule violation
type RulePunishment struct {
	IsFatal   bool
	HPDamage  int
	SANDamage int
}

// OptionTemplate represents a template for generating options
//
// Used for template-driven generation (fast path).
type OptionTemplate struct {
	ID          string
	Text        string   // Template text with placeholders (e.g., "檢查{object}")
	Variants    []string // Variants to fill placeholders
	Description string
	BaseRisk    RiskLevel
}
