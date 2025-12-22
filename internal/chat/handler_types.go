package chat

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// FlagHandlerContext contains all context needed by flag handlers.
// This provides handlers with access to game state, managers, and the judgment result.
type FlagHandlerContext struct {
	// NPCID is the NPC that the handler is processing for
	NPCID string

	// PlayerMessage is the player's original message
	PlayerMessage string

	// JudgeResult contains the JudgeAgent's analysis of the message
	JudgeResult *agents.JudgeResponseV2

	// GameState provides access to current game state
	GameState *engine.GameStateV2

	// UpdateManager handles knowledge propagation and contradictions
	UpdateManager *knowledge.UpdateManager

	// NPCManager handles NPC profiles and emotional state
	NPCManager *manager.NPCManager
}

// FlagHandlerResult is the result returned by a flag handler.
// It contains emotion changes, new facts, contradictions, and metadata.
type FlagHandlerResult struct {
	// EmotionDelta is the emotional change applied to the NPC
	EmotionDelta manager.EmotionDelta

	// NewFacts are any new facts that were registered
	NewFacts []knowledge.Fact

	// Contradictions are any contradictions that were detected
	Contradictions []knowledge.ContradictionResult

	// Metadata contains additional information about the handler execution
	Metadata map[string]interface{}

	// Success indicates if the handler executed successfully
	Success bool

	// Error contains an error message if Success is false
	Error string
}

// FlagHandler is the function signature for all flag handlers.
// Each handler receives context and returns a result.
type FlagHandler func(ctx FlagHandlerContext) FlagHandlerResult

// PendingLie represents a lie that has been told but not yet verified or exposed.
// Used by handleLie to track lies for potential future consequences.
type PendingLie struct {
	// NPCID is the NPC who heard the lie
	NPCID string

	// Content is the content of the lie
	Content string

	// Timestamp is when the lie was told (beat number)
	Timestamp int

	// PlayerMessage is the original player message
	PlayerMessage string

	// Verified indicates if the lie has been verified or exposed
	Verified bool

	// ExposedAt is when the lie was exposed (0 if not exposed)
	ExposedAt int
}
