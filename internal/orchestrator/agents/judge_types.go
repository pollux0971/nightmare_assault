package agents

// ==========================================================================
// Story 6-7: Judge Agent Types
// ==========================================================================

// ImpactLevel represents the impact level of a player choice
type ImpactLevel int

const (
	// ImpactNone indicates no impact
	ImpactNone ImpactLevel = iota
	// ImpactMinor indicates minor impact (-5 HP/SAN)
	ImpactMinor
	// ImpactModerate indicates moderate impact (-20~40 HP/SAN)
	ImpactModerate
	// ImpactMajor indicates major impact (-50~70 HP/SAN)
	ImpactMajor
	// ImpactLethal indicates lethal impact (-100 HP or SAN)
	ImpactLethal
)

// String returns the string representation of ImpactLevel
func (il ImpactLevel) String() string {
	switch il {
	case ImpactNone:
		return "None"
	case ImpactMinor:
		return "Minor"
	case ImpactModerate:
		return "Moderate"
	case ImpactMajor:
		return "Major"
	case ImpactLethal:
		return "Lethal"
	default:
		return "Unknown"
	}
}

// RuleType represents the type of hidden rule
type RuleType int

const (
	// RuleTypeScene indicates scene-based rule (highest priority)
	RuleTypeScene RuleType = iota
	// RuleTypeTime indicates time-based rule
	RuleTypeTime
	// RuleTypeBehavior indicates behavior-based rule
	RuleTypeBehavior
	// RuleTypeState indicates state-based rule (lowest priority)
	RuleTypeState
)

// String returns the string representation of RuleType
func (rt RuleType) String() string {
	switch rt {
	case RuleTypeScene:
		return "Scene"
	case RuleTypeTime:
		return "Time"
	case RuleTypeBehavior:
		return "Behavior"
	case RuleTypeState:
		return "State"
	default:
		return "Unknown"
	}
}

// GetRulePriority returns the priority value for a rule type
// Higher value = higher priority
func GetRulePriority(ruleType RuleType) int {
	switch ruleType {
	case RuleTypeScene:
		return 4 // Highest priority
	case RuleTypeTime:
		return 3
	case RuleTypeBehavior:
		return 2
	case RuleTypeState:
		return 1 // Lowest priority
	default:
		return 0
	}
}

// NextActionType represents the next action to take after judgment
type NextActionType int

const (
	// ActionContinueStory continues the story via Narration Agent
	ActionContinueStory NextActionType = iota
	// ActionApplyDamage applies damage and continues
	ActionApplyDamage
	// ActionTriggerDeath triggers the death flow
	ActionTriggerDeath
)

// String returns the string representation of NextActionType
func (nat NextActionType) String() string {
	switch nat {
	case ActionContinueStory:
		return "ContinueStory"
	case ActionApplyDamage:
		return "ApplyDamage"
	case ActionTriggerDeath:
		return "TriggerDeath"
	default:
		return "Unknown"
	}
}

// JudgeRequest is the request for judging a player's choice
//
// Parameters:
//   - PlayerChoice: The text of the player's choice
//   - GameState: Current game state (HP, SAN, warnings, etc.)
//   - ActiveRules: List of currently active hidden rules
type JudgeRequest struct {
	PlayerChoice string
	GameState    *GameStateSnapshot
	ActiveRules  []JudgeHiddenRule
}

// GameStateSnapshot is a snapshot of the current game state
type GameStateSnapshot struct {
	HP            int
	SAN           int
	CurrentScene  string
	Difficulty    string
	RuleWarnings  map[string]int // RuleID → Remaining warnings
	PlayerItems   []string       // Current items
	TurnNumber    int
}

// JudgeHiddenRule represents a hidden rule for judgment
//
// Simplified version for JudgeAgent use.
// Full implementation is in Epic 3 (Rule System).
type JudgeHiddenRule struct {
	ID                string
	Name              string
	Type              RuleType
	TriggerKeywords   []string
	TriggerRegex      string // Optional regex for complex patterns
	TriggerCondition  string // Human-readable description
	Punishment        RulePunishment
	MaxWarnings       int    // Maximum warnings before full damage
	DirectHint        string // Direct hint for easy difficulty
	MetaphorHint      string // Metaphor hint for normal difficulty
}

// JudgeResponse is the response containing judgment result
//
// Contains:
//   - RulesViolated: List of rules that were violated
//   - ImpactLevel: Overall impact level (None to Lethal)
//   - SuggestedStateChanges: Suggested HP/SAN changes
//   - DeathReason: Reason for death (if ImpactLevel = Lethal)
//   - NextAction: Suggested next action for orchestrator
type JudgeResponse struct {
	RulesViolated         []RuleViolation
	ImpactLevel           ImpactLevel
	SuggestedStateChanges StateChanges
	DeathReason           string
	NextAction            NextActionType
	Reasoning             string // Why this judgment was made
}

// RuleViolation represents a single rule violation
type RuleViolation struct {
	RuleID    string
	RuleName  string
	RuleType  RuleType
	Severity  ImpactLevel
	HPDamage  int
	SANDamage int
	IsFatal   bool
	Reason    string // Why this rule was triggered
}

// StateChanges represents suggested state changes
type StateChanges struct {
	HP                int
	SAN               int
	WarningsRemaining map[string]int // RuleID → Remaining warnings after this judgment
	Items             []string       // Item changes (future use)
}

// LLMJudgeResponse is the LLM's raw response for judgment
// Used for parsing LLM output
type LLMJudgeResponse struct {
	RulesViolated []string `json:"rules_violated"`
	ImpactLevel   string   `json:"impact_level"`
	Reasoning     string   `json:"reasoning"`
}
