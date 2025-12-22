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

// ==========================================================================
// Story 7.3: Intent Classification Types
// ==========================================================================

// IntentClassification represents the parsed intent from free text input
//
// Used when player enters free text instead of selecting predefined options.
// Judge Agent analyzes the text to understand what the player wants to do.
type IntentClassification struct {
	// Action is the primary action verb (e.g., "檢查", "攻擊", "逃跑")
	Action string

	// Target is the object of the action (e.g., "鏡子", "門", "角落")
	Target string

	// IsAmbiguous indicates if the intent is unclear and needs clarification
	IsAmbiguous bool

	// Confidence is the confidence level (0.0-1.0) of the classification
	Confidence float64

	// Keywords are the key terms extracted from the input
	Keywords []string

	// NormalizedIntent is the standardized form of the intent (for rule matching)
	NormalizedIntent string
}

// ClarificationNeeded represents a clarification request for ambiguous input
//
// When the player's intent is unclear, Judge Agent requests clarification
// before proceeding with rule checking.
type ClarificationNeeded struct {
	// Reason explains why clarification is needed
	Reason string

	// SuggestedInterpretations offers possible interpretations
	SuggestedInterpretations []string

	// Question is the clarification question to ask the player
	Question string
}

// JudgeResponseV2 extends JudgeResponse with clarification support
//
// Story 7.3 AC2: Support free text interpretation with clarification flow
type JudgeResponseV2 struct {
	// Embed original JudgeResponse
	*JudgeResponse

	// IntentClassification is the parsed intent (if input was free text)
	IntentClassification *IntentClassification

	// ClarificationNeeded indicates if player input needs clarification
	ClarificationNeeded *ClarificationNeeded
}

// LLMIntentResponse is the LLM's raw response for intent classification
// Used for parsing LLM output (Story 7.3 AC2)
type LLMIntentResponse struct {
	Action               string   `json:"action"`
	Target               string   `json:"target"`
	IsAmbiguous          bool     `json:"is_ambiguous"`
	Confidence           float64  `json:"confidence"`
	Keywords             []string `json:"keywords"`
	NormalizedIntent     string   `json:"normalized_intent"`
	ClarificationReason  string   `json:"clarification_reason"`
	SuggestedInterpret   []string `json:"suggested_interpretations"`
	ClarificationQuestion string  `json:"clarification_question"`
}

// ==========================================================================
// Story 4-1: Chat Judgment Types
// ==========================================================================

// JudgeChatRequest is the request for judging a player's chat message
//
// Story 4-1 AC2: Contains all necessary context for chat message judgment
type JudgeChatRequest struct {
	// PlayerMessage is the player's chat message to analyze
	PlayerMessage string

	// Participants is the list of all chat participants (NPCs and player)
	Participants []ChatParticipant

	// ConversationHistory is the recent chat message history (5-10 messages)
	ConversationHistory []ChatMessage

	// GameState is the current game state context
	GameState *GameStateSnapshot

	// RelevantFacts are the known facts relevant to this conversation
	RelevantFacts []string
}

// JudgeChatResult is the result of chat message judgment
//
// Story 4-1 AC3: Contains flags, confidence, and reasoning
type JudgeChatResult struct {
	// Flags are the detected chat flags (hallucination, hostile, revelation, etc.)
	Flags []ChatFlag

	// Confidence is the confidence level of the judgment (0.0-1.0)
	Confidence float64

	// Reasoning is the LLM's explanation for the judgment
	Reasoning string
}

// ChatParticipant represents a participant in the chat session
//
// Story 4-1 AC2: Contains participant information including emotion state
type ChatParticipant struct {
	ID           string       // Unique identifier (npc_id or "player")
	Name         string       // Display name
	IsPlayer     bool         // Whether this is the player
	Emotion      EmotionState // Current emotion state (Trust/Fear/Stress)
	Relationship string       // Relationship status: "friendly"/"neutral"/"hostile"
}

// ChatMessage represents a single chat message
//
// Story 4-1 AC2: Contains message information for conversation context
type ChatMessage struct {
	Speaker   string // Player ID or NPC ID
	Content   string // Message text content
	Timestamp string // When the message was sent (can be simplified for judgment)
}

// EmotionState represents an NPC's emotional state
//
// Story 4-1 AC5: Used in prompt to inform judgment
type EmotionState struct {
	Trust  int // Trust level (0-100)
	Fear   int // Fear level (0-100)
	Stress int // Stress level (0-100)
}

// ChatFlag represents semantic properties of chat messages
//
// Story 4-1 AC4: All possible chat flags for message classification
type ChatFlag string

const (
	// FlagHallucination indicates player states something contradicting known facts
	FlagHallucination ChatFlag = "hallucination"

	// FlagHostile indicates player shows threat or aggression
	FlagHostile ChatFlag = "hostile"

	// FlagRevelation indicates player shares important new information
	FlagRevelation ChatFlag = "revelation"

	// FlagContradiction indicates player's statement conflicts with NPCs' knowledge
	FlagContradiction ChatFlag = "contradiction"

	// FlagPersuasion indicates player attempts to convince NPCs
	FlagPersuasion ChatFlag = "persuasion"

	// FlagLie indicates player is lying (may be exposed later)
	FlagLie ChatFlag = "lie"
)

// String returns the string representation of ChatFlag
func (f ChatFlag) String() string {
	return string(f)
}

// LLMChatJudgeResponse is the LLM's raw response for chat judgment
// Used for parsing LLM JSON output
type LLMChatJudgeResponse struct {
	Flags      []string `json:"flags"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
}
