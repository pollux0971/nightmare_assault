package views

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 4.6: Emotion Update Integration - Unit Tests
// Target: ≥6 unit tests with >80% coverage
// ==========================================================================

// TestHandleProcessResult_SingleNPC tests emotion update for a single NPC.
// Story 4.6 AC1, AC2, AC3, AC4: Full emotion update flow
func TestHandleProcessResult_SingleNPC(t *testing.T) {
	// Setup
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add test NPC
	profile := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	}
	npcMgr.AddNPC(profile)

	// Add participant to overlay
	overlay.participants = []ChatParticipant{
		NewChatParticipant("npc1", "Test NPC", false, manager.NewEmotionState(50, 50, 50)),
	}
	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Hello, friend!")

	// Create emotion changes and NPC responses
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 10, Fear: -5, Stress: -3},
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{
		{
			NPCID:   "npc1",
			Content: "Hello to you too!",
			Flags:   []ChatFlag{},
		},
	}

	// Execute
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify emotion was updated in NPCManager
	state := npcMgr.GetState("npc1")
	if state == nil {
		t.Fatal("NPC state is nil")
	}
	if state.Emotion.Trust != 60 {
		t.Errorf("Expected Trust=60, got %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 45 {
		t.Errorf("Expected Fear=45, got %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress != 47 {
		t.Errorf("Expected Stress=47, got %d", state.Emotion.Stress)
	}

	// Verify UI participant was updated
	if overlay.participants[0].Emotion.Trust != 60 {
		t.Errorf("UI participant Trust not updated: got %d", overlay.participants[0].Emotion.Trust)
	}

	// Verify interaction was recorded
	if len(state.Interactions) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(state.Interactions))
	}
	interaction := state.Interactions[0]
	if interaction.InteractionType != "chat_message" {
		t.Errorf("Expected type 'chat_message', got '%s'", interaction.InteractionType)
	}
	if interaction.EmotionDelta.Trust != 10 {
		t.Errorf("Expected delta Trust=10, got %d", interaction.EmotionDelta.Trust)
	}

	// Verify NPC response was added to messages
	if len(overlay.messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(overlay.messages))
	}
	if overlay.messages[0].Content != "Hello to you too!" {
		t.Errorf("Message content mismatch: %s", overlay.messages[0].Content)
	}
}

// TestHandleProcessResult_MultipleNPCs tests emotion updates for multiple NPCs.
// Story 4.6 AC2: Multiple NPCs can be updated in one call
func TestHandleProcessResult_MultipleNPCs(t *testing.T) {
	// Setup
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add two NPCs
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc2",
		Name:           "NPC Two",
		InitialEmotion: manager.NewEmotionState(40, 60, 55),
	})

	overlay.participants = []ChatParticipant{
		NewChatParticipant("npc1", "NPC One", false, manager.NewEmotionState(50, 50, 50)),
		NewChatParticipant("npc2", "NPC Two", false, manager.NewEmotionState(40, 60, 55)),
	}
	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Group message")

	// Create emotion changes and NPC responses
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 15, Fear: -10, Stress: 0},
		"npc2": {Trust: -5, Fear: 20, Stress: 10},
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{
		{NPCID: "npc1", Content: "I trust you more!"},
		{NPCID: "npc2", Content: "I'm scared..."},
	}

	// Execute
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify both NPCs were updated
	state1 := npcMgr.GetState("npc1")
	if state1.Emotion.Trust != 65 {
		t.Errorf("NPC1 Trust: expected 65, got %d", state1.Emotion.Trust)
	}
	if state1.Emotion.Fear != 40 {
		t.Errorf("NPC1 Fear: expected 40, got %d", state1.Emotion.Fear)
	}

	state2 := npcMgr.GetState("npc2")
	if state2.Emotion.Trust != 35 {
		t.Errorf("NPC2 Trust: expected 35, got %d", state2.Emotion.Trust)
	}
	if state2.Emotion.Fear != 80 {
		t.Errorf("NPC2 Fear: expected 80, got %d", state2.Emotion.Fear)
	}

	// Verify both UI participants updated
	if overlay.participants[0].Emotion.Trust != 65 {
		t.Errorf("UI NPC1 not updated")
	}
	if overlay.participants[1].Emotion.Fear != 80 {
		t.Errorf("UI NPC2 not updated")
	}

	// Verify both interactions recorded
	if len(state1.Interactions) != 1 {
		t.Errorf("NPC1 should have 1 interaction")
	}
	if len(state2.Interactions) != 1 {
		t.Errorf("NPC2 should have 1 interaction")
	}
}

// TestUpdateParticipantEmotion_EmotionIntegration tests UI participant emotion update.
// Story 4.6 AC3: Participant emotions update in real-time in UI
func TestUpdateParticipantEmotion_EmotionIntegration(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay.participants = []ChatParticipant{
		NewChatParticipant("npc1", "Test NPC", false, manager.NewEmotionState(50, 50, 50)),
	}
	overlay.SetNPCManager(npcMgr)

	// Update emotion in NPCManager
	npcMgr.AdjustEmotion("npc1", manager.EmotionDelta{Trust: 20, Fear: -10, Stress: 5})

	// Call updateParticipantEmotion
	overlay.updateParticipantEmotion("npc1")

	// Verify UI was updated
	if overlay.participants[0].Emotion.Trust != 70 {
		t.Errorf("Expected Trust=70, got %d", overlay.participants[0].Emotion.Trust)
	}
	if overlay.participants[0].Emotion.Fear != 40 {
		t.Errorf("Expected Fear=40, got %d", overlay.participants[0].Emotion.Fear)
	}
	if overlay.participants[0].Emotion.Stress != 55 {
		t.Errorf("Expected Stress=55, got %d", overlay.participants[0].Emotion.Stress)
	}
}

// TestRecordInteraction tests interaction recording.
// Story 4.6 AC4: Emotion changes recorded to NPCInteraction history
func TestRecordInteraction(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Test message")

	// Record interaction
	delta := manager.EmotionDelta{Trust: 5, Fear: -2, Stress: 1}
	overlay.recordInteraction("npc1", delta, "test_reason")

	// Verify interaction was recorded
	state := npcMgr.GetState("npc1")
	if len(state.Interactions) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(state.Interactions))
	}

	interaction := state.Interactions[0]
	if interaction.InteractionType != "test_reason" {
		t.Errorf("Wrong interaction type: %s", interaction.InteractionType)
	}
	if interaction.EmotionDelta.Trust != 5 {
		t.Errorf("Wrong delta Trust: %d", interaction.EmotionDelta.Trust)
	}
	if interaction.Description == "" {
		t.Error("Description should not be empty")
	}
}

// TestEmotionChangeZero tests handling of zero emotion changes.
// Story 4.6: Edge case - zero delta should not fail
func TestEmotionChangeZero(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay.participants = []ChatParticipant{
		NewChatParticipant("npc1", "Test NPC", false, manager.NewEmotionState(50, 50, 50)),
	}
	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Neutral message")

	// Process with zero emotion change
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 0, Fear: 0, Stress: 0},
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify emotion didn't change
	state := npcMgr.GetState("npc1")
	if state.Emotion.Trust != 50 || state.Emotion.Fear != 50 || state.Emotion.Stress != 50 {
		t.Error("Emotion should remain unchanged with zero delta")
	}

	// Verify interaction was still recorded
	if len(state.Interactions) != 1 {
		t.Error("Interaction should be recorded even with zero delta")
	}
}

// TestInvalidNPCID tests handling of invalid NPC IDs.
// Story 4.6: Error handling - invalid NPCID should be gracefully handled
func TestInvalidNPCID(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Message to non-existent NPC")

	// Process with non-existent NPC
	emotionChanges := map[string]manager.EmotionDelta{
		"nonexistent": {Trust: 10, Fear: 5, Stress: 0},
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	// Should not panic
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// No messages should be added (since NPC doesn't exist)
	// This is graceful degradation - we log a warning but continue
}

// TestApplyEmotionChanges tests the lower-level emotion change application.
// Story 4.6 AC2: Direct emotion application
func TestApplyEmotionChanges(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay.participants = []ChatParticipant{
		NewChatParticipant("npc1", "Test NPC", false, manager.NewEmotionState(50, 50, 50)),
	}
	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Direct change")

	// Apply emotion changes directly
	changes := map[string]manager.EmotionDelta{
		"npc1": {Trust: 15, Fear: -5, Stress: 10},
	}

	err := overlay.ApplyEmotionChanges(changes)
	if err != nil {
		t.Fatalf("ApplyEmotionChanges failed: %v", err)
	}

	// Verify changes applied
	state := npcMgr.GetState("npc1")
	if state.Emotion.Trust != 65 {
		t.Errorf("Expected Trust=65, got %d", state.Emotion.Trust)
	}

	// Verify UI updated
	if overlay.participants[0].Emotion.Trust != 65 {
		t.Error("UI not updated")
	}

	// Verify interaction recorded
	if len(state.Interactions) != 1 {
		t.Error("Interaction not recorded")
	}
}

// TestGetNPCEmotion tests retrieving current NPC emotion state.
// Story 4.6: Helper method for querying emotion state
func TestGetNPCEmotion(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(60, 40, 55),
	})

	overlay.SetNPCManager(npcMgr)

	// Get emotion
	emotion := overlay.GetNPCEmotion("npc1")
	if emotion == nil {
		t.Fatal("Emotion should not be nil")
	}
	if emotion.Trust != 60 {
		t.Errorf("Expected Trust=60, got %d", emotion.Trust)
	}

	// Test non-existent NPC
	emotion2 := overlay.GetNPCEmotion("nonexistent")
	if emotion2 != nil {
		t.Error("Non-existent NPC should return nil")
	}
}

// TestGetNPCInteractionHistory tests retrieving interaction history.
// Story 4.6: Helper method for querying interaction history
func TestGetNPCInteractionHistory(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Test NPC",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay.SetNPCManager(npcMgr)
	overlay.SetLastPlayerMessage("Message 1")

	// Record multiple interactions
	overlay.recordInteraction("npc1", manager.EmotionDelta{Trust: 5}, "reason1")
	time.Sleep(time.Millisecond) // Ensure different timestamps
	overlay.recordInteraction("npc1", manager.EmotionDelta{Fear: 10}, "reason2")
	time.Sleep(time.Millisecond)
	overlay.recordInteraction("npc1", manager.EmotionDelta{Stress: -5}, "reason3")

	// Get all interactions
	history := overlay.GetNPCInteractionHistory("npc1", 0)
	if len(history) != 3 {
		t.Fatalf("Expected 3 interactions, got %d", len(history))
	}

	// Get last 2 interactions
	recent := overlay.GetNPCInteractionHistory("npc1", 2)
	if len(recent) != 2 {
		t.Fatalf("Expected 2 recent interactions, got %d", len(recent))
	}
	if recent[0].InteractionType != "reason2" {
		t.Errorf("Wrong order: expected reason2, got %s", recent[0].InteractionType)
	}

	// Test non-existent NPC
	none := overlay.GetNPCInteractionHistory("nonexistent", 0)
	if none != nil {
		t.Error("Non-existent NPC should return nil")
	}
}

// TestHandleProcessResult_NoNPCManager tests graceful handling when NPCManager is not set.
// Story 4.6: Error handling - missing NPCManager
func TestHandleProcessResult_NoNPCManager(t *testing.T) {
	overlay := NewChatOverlayModel()
	// Don't set NPCManager

	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 10},
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	// Should not panic
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")
}

// TestHandleProcessResult_FailedResult tests handling of failed ProcessResult.
// Story 4.6: Error handling - ProcessResult.Success = false
func TestHandleProcessResult_FailedResult(t *testing.T) {
	overlay := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)
	overlay.SetNPCManager(npcMgr)

	emotionChanges := map[string]manager.EmotionDelta{}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	// Should not panic and should not apply any changes
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, false, "Processing failed")
}
