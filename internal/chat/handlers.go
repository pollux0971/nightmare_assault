package chat

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// AC1: handleHallucination() - Reduces Trust
// ==========================================================================

// handleHallucination processes the hallucination flag.
// When a player states something that seems like a hallucination (false information),
// it reduces the NPC's trust in the player.
//
// Trust reduction:
// - Base: -10
// - Skeptical NPCs: -15 (1.5x multiplier)
// - Trusting NPCs: -7 (0.7x multiplier)
//
// AC1: handleHallucination() 處理幻覺標記（降低信任）
func handleHallucination(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	// Base trust loss
	baseTrustLoss := -10

	// Adjust based on NPC archetype
	multiplier := 1.0
	if npc.Archetype == "Skeptical" {
		multiplier = 1.5 // More sensitive to hallucinations
	} else if npc.Archetype == "Trusting" {
		multiplier = 0.7 // More forgiving
	}

	// Check for skeptical/trusting traits
	if hasTraitContains(npc, "Skeptical") || hasTraitContains(npc, "Paranoid") {
		multiplier *= 1.2
	} else if hasTraitContains(npc, "Trusting") || hasTraitContains(npc, "Naive") {
		multiplier *= 0.8
	}

	trustLoss := int(float64(baseTrustLoss) * multiplier)

	delta := manager.EmotionDelta{
		Trust: trustLoss,
	}

	// Apply emotion change
	err := ctx.NPCManager.AdjustEmotion(ctx.NPCID, delta)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to adjust emotion for NPC %s: %v", ctx.NPCID, err))
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to adjust emotion: %v", err),
		}
	}

	// Log the hallucination event
	logger.Info(fmt.Sprintf("Hallucination detected from player to NPC %s (%s). Trust: %d",
		ctx.NPCID, npc.Name, trustLoss))

	return FlagHandlerResult{
		EmotionDelta: delta,
		Success:      true,
		Metadata: map[string]interface{}{
			"event_type": "hallucination",
			"npc_id":     ctx.NPCID,
			"npc_name":   npc.Name,
			"archetype":  npc.Archetype,
			"multiplier": multiplier,
		},
	}
}

// ==========================================================================
// AC2: handleHostility() - Increases Fear/Stress
// ==========================================================================

// handleHostility processes the hostility flag.
// When a player shows hostility or aggression, it increases the NPC's
// fear and stress while reducing trust.
//
// Default values:
// - Fear: +15
// - Stress: +10
// - Trust: -15
//
// Adjustments:
// - Cowardly NPCs: Fear +50%, Stress +30%
// - Brave NPCs: Fear -40%, Stress -20%
//
// AC2: handleHostility() 處理敵意標記（增加恐懼/壓力）
func handleHostility(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	// Base values
	baseFear := 15
	baseStress := 10
	baseTrustLoss := -15

	// Adjust based on personality traits
	fearMultiplier := 1.0
	stressMultiplier := 1.0

	if hasTraitContains(npc, "Cowardly") || hasTraitContains(npc, "Timid") {
		fearMultiplier = 1.5
		stressMultiplier = 1.3
	} else if hasTraitContains(npc, "Brave") || hasTraitContains(npc, "Fearless") {
		fearMultiplier = 0.6
		stressMultiplier = 0.8
	}

	// Archetype adjustments
	if npc.Archetype == "Cowardly" {
		fearMultiplier *= 1.3
		stressMultiplier *= 1.2
	} else if npc.Archetype == "Brave" || npc.Archetype == "Soldier" {
		fearMultiplier *= 0.7
		stressMultiplier *= 0.9
	}

	fear := int(float64(baseFear) * fearMultiplier)
	stress := int(float64(baseStress) * stressMultiplier)

	delta := manager.EmotionDelta{
		Trust:  baseTrustLoss,
		Fear:   fear,
		Stress: stress,
	}

	// Apply emotion change
	err := ctx.NPCManager.AdjustEmotion(ctx.NPCID, delta)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to adjust emotion for NPC %s: %v", ctx.NPCID, err))
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to adjust emotion: %v", err),
		}
	}

	// Log the hostility event
	logger.Info(fmt.Sprintf("Hostility detected from player to NPC %s (%s). Trust: %d, Fear: +%d, Stress: +%d",
		ctx.NPCID, npc.Name, baseTrustLoss, fear, stress))

	return FlagHandlerResult{
		EmotionDelta: delta,
		Success:      true,
		Metadata: map[string]interface{}{
			"event_type":        "hostility",
			"npc_id":            ctx.NPCID,
			"npc_name":          npc.Name,
			"fear_multiplier":   fearMultiplier,
			"stress_multiplier": stressMultiplier,
		},
	}
}

// ==========================================================================
// AC3: handleRevelation() - Propagates New Information
// ==========================================================================

// handleRevelation processes the revelation flag.
// When a player shares important new information, it:
// 1. Registers the information as a new fact
// 2. Propagates it to NPCs in the same room via UpdateManager
// 3. Slightly increases trust (player is sharing valuable info)
//
// Trust increase: +3 to +8 based on information value
//
// AC3: handleRevelation() 處理揭露標記（傳播新資訊）
func handleRevelation(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	// Generate unique fact ID
	factID := generateFactID()

	// Extract location from game state
	location := "unknown"
	if ctx.GameState != nil {
		location = ctx.GameState.CurrentScene
	}

	// Create the fact
	fact := knowledge.Fact{
		ID:        factID,
		Content:   ctx.PlayerMessage,
		Type:      knowledge.Discovery, // Revelation is a type of discovery
		Source:    "player",
		CreatedAt: time.Now(),
		Location:  location,
		Witnesses: []string{ctx.NPCID, "player"},
	}

	// Register the fact globally
	if ctx.UpdateManager != nil {
		ctx.UpdateManager.RegisterFact(&fact)

		// Propagate to NPCs in the same room using GameEvent
		beat := 0
		if ctx.GameState != nil {
			beat = ctx.GameState.CurrentBeat
		}
		event := &knowledge.GameEvent{
			ID:          factID,
			Type:        "revelation",
			Description: ctx.PlayerMessage,
			Initiator:   "player",
			Location:    location,
			Beat:        beat,
			Importance:  5,
		}
		ctx.UpdateManager.PropagateEvent(event)
	}

	// Calculate trust increase based on information value
	// Base: +5, can vary from +3 to +8
	trustIncrease := 5
	// Add some randomness to simulate varying information value
	trustIncrease += rand.Intn(4) - 1 // -1 to +2

	delta := manager.EmotionDelta{
		Trust: trustIncrease,
	}

	// Apply emotion change
	err := ctx.NPCManager.AdjustEmotion(ctx.NPCID, delta)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to adjust emotion for NPC %s: %v", ctx.NPCID, err))
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to adjust emotion: %v", err),
		}
	}

	// Log the revelation event
	logger.Info(fmt.Sprintf("Revelation from player to NPC %s (%s). Trust: +%d, Fact: %s",
		ctx.NPCID, npc.Name, trustIncrease, factID))

	return FlagHandlerResult{
		EmotionDelta: delta,
		NewFacts:     []knowledge.Fact{fact},
		Success:      true,
		Metadata: map[string]interface{}{
			"event_type": "revelation",
			"npc_id":     ctx.NPCID,
			"npc_name":   npc.Name,
			"fact_id":    factID,
			"location":   location,
		},
	}
}

// ==========================================================================
// AC4: handleContradiction() - Calls UpdateManager
// ==========================================================================

// handleContradiction processes the contradiction flag.
// When a player's statement contradicts what the NPC knows, it:
// 1. Calls UpdateManager.CheckContradiction()
// 2. Applies the suggested emotion delta
// 3. Records the contradiction
//
// Emotion impact depends on contradiction severity (determined by UpdateManager).
//
// AC4: handleContradiction() 處理矛盾標記（呼叫 UpdateManager）
func handleContradiction(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	// Check for contradiction using UpdateManager
	var contradiction *knowledge.ContradictionResult
	if ctx.UpdateManager != nil {
		contradiction = ctx.UpdateManager.CheckContradiction(ctx.NPCID, ctx.PlayerMessage)
	}

	// If no contradiction found, return empty result
	if contradiction == nil {
		logger.Info(fmt.Sprintf("Contradiction flag set but no contradiction detected for NPC %s", ctx.NPCID))
		return FlagHandlerResult{
			Success: true,
			Metadata: map[string]interface{}{
				"event_type": "contradiction",
				"npc_id":     ctx.NPCID,
				"found":      false,
			},
		}
	}

	// Convert knowledge.EmotionDelta to manager.EmotionDelta
	delta := manager.EmotionDelta{
		Trust:  contradiction.SuggestedDelta.Trust,
		Fear:   contradiction.SuggestedDelta.Fear,
		Stress: contradiction.SuggestedDelta.Stress,
	}

	// Apply the suggested emotion delta
	err := ctx.NPCManager.AdjustEmotion(ctx.NPCID, delta)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to adjust emotion for NPC %s: %v", ctx.NPCID, err))
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to adjust emotion: %v", err),
		}
	}

	// Log the contradiction event
	logger.Info(fmt.Sprintf("Contradiction detected for NPC %s (%s). Severity: %d (%s), Trust: %d, Fear: %d, Stress: %d",
		ctx.NPCID, npc.Name, contradiction.Severity, contradiction.Type, delta.Trust, delta.Fear, delta.Stress))

	return FlagHandlerResult{
		EmotionDelta:   delta,
		Contradictions: []knowledge.ContradictionResult{*contradiction},
		Success:        true,
		Metadata: map[string]interface{}{
			"event_type":        "contradiction",
			"npc_id":            ctx.NPCID,
			"npc_name":          npc.Name,
			"severity":          contradiction.Severity,
			"type":              string(contradiction.Type),
			"suggested_reaction": contradiction.SuggestedReaction,
			"found":             true,
		},
	}
}

// ==========================================================================
// AC5: handlePersuasion() - Emotion Changes Based on Success
// ==========================================================================

// handlePersuasion processes the persuasion flag.
// When a player attempts to persuade an NPC, the success depends on:
// - Current trust level (higher trust = higher success chance)
// - NPC personality traits
//
// Success:
// - Trust: +5 to +12
// - Fear: -3 to -8
//
// Failure:
// - Trust: -2 to -5
//
// AC5: handlePersuasion() 處理說服標記（情感變化）
func handlePersuasion(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	state := ctx.NPCManager.GetState(ctx.NPCID)
	if state == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC state %s not found", ctx.NPCID),
		}
	}

	// Calculate success probability based on trust level
	// Trust 0-30: 20% chance
	// Trust 30-60: 50% chance
	// Trust 60-100: 80% chance
	baseSuccessChance := 0.2
	if state.Emotion.Trust >= 60 {
		baseSuccessChance = 0.8
	} else if state.Emotion.Trust >= 30 {
		baseSuccessChance = 0.5
	}

	// Adjust based on NPC traits
	successMultiplier := 1.0
	if hasTraitContains(npc, "Stubborn") || hasTraitContains(npc, "Independent") {
		successMultiplier = 0.7 // Harder to persuade
	} else if hasTraitContains(npc, "Impressionable") || hasTraitContains(npc, "Naive") {
		successMultiplier = 1.3 // Easier to persuade
	}

	successChance := baseSuccessChance * successMultiplier

	// Roll for success
	success := rand.Float64() < successChance

	var delta manager.EmotionDelta
	if success {
		// Successful persuasion
		trustGain := 5 + rand.Intn(8)    // +5 to +12
		fearReduction := -(3 + rand.Intn(6)) // -3 to -8
		delta = manager.EmotionDelta{
			Trust: trustGain,
			Fear:  fearReduction,
		}
	} else {
		// Failed persuasion
		trustLoss := -(2 + rand.Intn(4)) // -2 to -5
		delta = manager.EmotionDelta{
			Trust: trustLoss,
		}
	}

	// Apply emotion change
	err := ctx.NPCManager.AdjustEmotion(ctx.NPCID, delta)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to adjust emotion for NPC %s: %v", ctx.NPCID, err))
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to adjust emotion: %v", err),
		}
	}

	// Log the persuasion attempt
	logger.Info(fmt.Sprintf("Persuasion attempt on NPC %s (%s). Success: %t, Trust change: %d",
		ctx.NPCID, npc.Name, success, delta.Trust))

	return FlagHandlerResult{
		EmotionDelta: delta,
		Success:      true,
		Metadata: map[string]interface{}{
			"event_type":      "persuasion",
			"npc_id":          ctx.NPCID,
			"npc_name":        npc.Name,
			"persuasion_success": success,
			"success_chance":  successChance,
			"base_trust":      state.Emotion.Trust,
		},
	}
}

// ==========================================================================
// AC6: handleLie() - Records for Later Verification
// ==========================================================================

// handleLie processes the lie flag.
// When a player lies, the lie is recorded for potential future exposure.
// Currently, no immediate emotion change occurs (NPC doesn't know it's a lie yet).
//
// The lie is stored in a pending list and can be exposed later, which would
// cause severe trust damage.
//
// AC6: handleLie() 處理謊言標記（後續揭穿時懲罰）
func handleLie(ctx FlagHandlerContext) FlagHandlerResult {
	npc := ctx.NPCManager.GetProfile(ctx.NPCID)
	if npc == nil {
		return FlagHandlerResult{
			Success: false,
			Error:   fmt.Sprintf("NPC %s not found", ctx.NPCID),
		}
	}

	// Record the lie for potential future verification
	// In a full implementation, this would be stored in a global lie tracker
	// For now, we just log it and return metadata

	timestamp := 0
	if ctx.GameState != nil {
		timestamp = ctx.GameState.CurrentBeat
	}

	pendingLie := PendingLie{
		NPCID:         ctx.NPCID,
		Content:       ctx.PlayerMessage,
		Timestamp:     timestamp,
		PlayerMessage: ctx.PlayerMessage,
		Verified:      false,
		ExposedAt:     0,
	}

	// Log the lie
	logger.Info(fmt.Sprintf("Lie detected from player to NPC %s (%s). Recorded for future verification. Beat: %d",
		ctx.NPCID, npc.Name, timestamp))

	// Return empty emotion delta (no immediate impact)
	return FlagHandlerResult{
		EmotionDelta: manager.EmotionDelta{}, // No immediate emotion change
		Success:      true,
		Metadata: map[string]interface{}{
			"event_type":    "lie",
			"npc_id":        ctx.NPCID,
			"npc_name":      npc.Name,
			"pending_lie":   pendingLie,
			"timestamp":     timestamp,
			"immediate_impact": false,
		},
	}
}

// ==========================================================================
// Helper Functions
// ==========================================================================

// hasTraitContains checks if an NPC has a trait containing the given keyword.
// Case-insensitive partial match.
func hasTraitContains(npc *manager.NPCProfile, keyword string) bool {
	if npc == nil {
		return false
	}

	lowerKeyword := strings.ToLower(keyword)
	for _, trait := range npc.Traits {
		if strings.Contains(strings.ToLower(trait.Content), lowerKeyword) {
			return true
		}
	}
	return false
}

// generateFactID generates a unique fact ID using UUID.
func generateFactID() string {
	return fmt.Sprintf("fact_%s", uuid.New().String())
}
