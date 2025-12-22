package trinity

// TierLevel represents the three-tier model hierarchy
type TierLevel int

const (
	// TierThinking - Highest tier, most capable and expensive (Opus 4.5)
	// Used for: Complex reasoning, strategic decisions, critical judgments
	TierThinking TierLevel = iota

	// TierReactive - Middle tier, balanced performance and cost (Sonnet 3.5)
	// Used for: Dynamic content generation, moderate complexity tasks
	TierReactive

	// TierRapid - Lowest tier, fastest and cheapest (Haiku 3.5)
	// Used for: Simple responses, high-frequency operations
	TierRapid
)

// String returns the string representation of TierLevel
func (t TierLevel) String() string {
	switch t {
	case TierThinking:
		return "Thinking"
	case TierReactive:
		return "Reactive"
	case TierRapid:
		return "Rapid"
	default:
		return "Unknown"
	}
}

// DefaultAgentTierMapping defines the default Agent → Tier mapping
// This can be overridden by user configuration
var DefaultAgentTierMapping = map[string]TierLevel{
	// Thinking Tier - Complex reasoning and critical decisions
	"JudgeAgent":         TierThinking,
	"SeedAgent":          TierThinking,
	"NPCAgent":           TierThinking,
	"ContradictionAgent": TierThinking,

	// Reactive Tier - Dynamic content generation
	"NarrationAgent": TierReactive,
	"ChoiceAgent":    TierReactive,
	"ChatProcessor":  TierReactive,

	// Rapid Tier - Simple and frequent operations
	"DreamAgent":       TierRapid,
	"EnvironmentAgent": TierRapid,
	"SummaryAgent":     TierRapid,
}

// GetTierForAgent returns the tier level for a given agent name
// Falls back to TierReactive if agent is not in the mapping
func GetTierForAgent(agentName string, overrides map[string]TierLevel) TierLevel {
	// Check user overrides first
	if tier, ok := overrides[agentName]; ok {
		return tier
	}

	// Check default mapping
	if tier, ok := DefaultAgentTierMapping[agentName]; ok {
		return tier
	}

	// Default to Reactive tier for unknown agents
	return TierReactive
}

// ParseTierLevel converts a string to TierLevel
func ParseTierLevel(s string) (TierLevel, bool) {
	switch s {
	case "Thinking", "thinking":
		return TierThinking, true
	case "Reactive", "reactive":
		return TierReactive, true
	case "Rapid", "rapid":
		return TierRapid, true
	default:
		return TierReactive, false
	}
}
