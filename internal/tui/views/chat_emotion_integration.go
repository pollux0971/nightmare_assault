package views

import (
	"fmt"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 4.6: Emotion Update Integration
// This file contains the full implementation of emotion updates from chat
// Note: To avoid import cycles, we use interface{} for ProcessResult and
// access fields via reflection or type assertion at the integration point.
// ==========================================================================

// ProcessResultInterface is a duck-typed interface for chat.ProcessResult
// to avoid import cycle. The actual type is defined in internal/chat/types.go
type ProcessResultInterface interface {
	GetEmotionChanges() map[string]manager.EmotionDelta
	GetNPCResponses() []NPCResponseInterface
	IsSuccess() bool
	GetError() string
}

// NPCResponseInterface is a duck-typed interface for chat.NPCResponse
type NPCResponseInterface interface {
	GetNPCID() string
	GetContent() string
	GetFlags() []ChatFlag
}

// HandleProcessResultDirect processes emotion changes and NPC responses directly.
// Story 4.6 AC1: ProcessResult contains emotion changes for each NPC (map[string]EmotionDelta).
// Story 4.6 AC2: ChatOverlay receives ProcessResult and applies emotion changes via NPCManager.
// Story 4.6 AC3: Participant emotions update in real-time in UI.
// Story 4.6 AC4: Emotion changes recorded to NPCInteraction history.
//
// This method takes the individual components to avoid import cycle with internal/chat.
// The actual HandleProcessResult wrapper will be in the integration layer.
func (m *ChatOverlayModel) HandleProcessResultDirect(
	emotionChanges map[string]manager.EmotionDelta,
	npcResponses []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	},
	success bool,
	errorMsg string,
) {
	if !success {
		// Log error but don't crash - chat should continue
		logger.Warn("ProcessResult failed", map[string]interface{}{
			"error": errorMsg,
		})
		return
	}

	// AC2: Apply emotion changes to each NPC
	for npcID, delta := range emotionChanges {
		if m.npcManager == nil {
			logger.Warn("NPCManager not set, skipping emotion update", map[string]interface{}{
				"npcID": npcID,
			})
			continue
		}

		// Apply emotion delta via NPCManager
		err := m.npcManager.AdjustEmotion(npcID, delta)
		if err != nil {
			logger.Warn("Failed to adjust emotion", map[string]interface{}{
				"npcID": npcID,
				"error": err.Error(),
			})
			continue
		}

		// AC3: Update UI participant state
		m.updateParticipantEmotion(npcID)

		// AC4: Record interaction history
		m.recordInteraction(npcID, delta, "chat_message")
	}

	// Add NPC responses to message list
	for _, npcResp := range npcResponses {
		m.AddNPCMessage(npcResp.NPCID, npcResp.Content, npcResp.Flags)
	}
}

// ApplyEmotionChanges is a lower-level method that applies emotion changes without
// adding NPC response messages. This is useful for testing or special scenarios.
// Story 4.6 AC2: Direct emotion application via NPCManager.
func (m *ChatOverlayModel) ApplyEmotionChanges(emotionChanges map[string]manager.EmotionDelta) error {
	if m.npcManager == nil {
		return fmt.Errorf("NPCManager not set")
	}

	for npcID, delta := range emotionChanges {
		// Apply emotion delta
		if err := m.npcManager.AdjustEmotion(npcID, delta); err != nil {
			return fmt.Errorf("failed to adjust emotion for NPC %s: %w", npcID, err)
		}

		// Update UI
		m.updateParticipantEmotion(npcID)

		// Record interaction
		m.recordInteraction(npcID, delta, "direct_emotion_change")
	}

	return nil
}

// GetNPCEmotion returns the current emotion state of an NPC participant.
// Returns nil if NPC is not in chat or NPCManager is not set.
func (m *ChatOverlayModel) GetNPCEmotion(npcID string) *manager.EmotionState {
	if m.npcManager == nil {
		return nil
	}

	state := m.npcManager.GetState(npcID)
	if state == nil {
		return nil
	}

	emotion := state.Emotion
	return &emotion
}

// GetNPCInteractionHistory returns the recent interaction history for an NPC.
// Returns nil if NPC is not found or NPCManager is not set.
// The count parameter limits how many recent interactions to return (0 = all).
func (m *ChatOverlayModel) GetNPCInteractionHistory(npcID string, count int) []manager.NPCInteraction {
	if m.npcManager == nil {
		return nil
	}

	state := m.npcManager.GetState(npcID)
	if state == nil {
		return nil
	}

	if count <= 0 || count > len(state.Interactions) {
		// Return all interactions
		result := make([]manager.NPCInteraction, len(state.Interactions))
		copy(result, state.Interactions)
		return result
	}

	// Return last N interactions
	return state.GetRecentInteractions(count)
}

// SetLastPlayerMessage sets the last player message for interaction recording.
// This should be called before processing chat results.
func (m *ChatOverlayModel) SetLastPlayerMessage(message string) {
	m.lastPlayerMessage = message
}

// updateParticipantEmotion is the internal implementation that updates UI state.
// Story 4.6 AC3: Participant emotions update in real-time in UI.
func (m *ChatOverlayModel) updateParticipantEmotion(npcID string) {
	if m.npcManager == nil {
		return
	}

	state := m.npcManager.GetState(npcID)
	if state == nil {
		return
	}

	for i, p := range m.participants {
		if p.ID == npcID {
			m.participants[i].Emotion = state.Emotion
			// Update relationship display as well
			m.participants[i].IsActive = state.IsAlive
			break
		}
	}
}

// recordInteraction is the internal implementation that records to history.
// Story 4.6 AC4: Emotion changes recorded to NPCInteraction history.
func (m *ChatOverlayModel) recordInteraction(npcID string, delta manager.EmotionDelta, reason string) {
	if m.npcManager == nil {
		return
	}

	interaction := manager.NPCInteraction{
		Timestamp:       time.Now(),
		InteractionType: reason,
		EmotionDelta:    delta,
		Description:     fmt.Sprintf("%s: %s (Trust:%+d, Fear:%+d, Stress:%+d)", reason, m.lastPlayerMessage, delta.Trust, delta.Fear, delta.Stress),
	}

	err := m.npcManager.RecordInteraction(npcID, interaction)
	if err != nil {
		// Log warning but don't fail - this is non-critical
		logger.Warn("Failed to record interaction", map[string]interface{}{
			"npcID": npcID,
			"error": err.Error(),
		})
	}
}
