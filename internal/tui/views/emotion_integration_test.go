package views

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 4.6: Emotion Update Integration - Integration Tests
// Target: ≥3 integration tests
// ==========================================================================

// TestFullChatFlow_EmotionUpdate tests the complete chat → emotion update flow.
// Story 4.6: Full integration from chat processing to emotion update
func TestFullChatFlow_EmotionUpdate(t *testing.T) {
	// This is a simulation of the full flow:
	// 1. Player sends message
	// 2. ChatProcessor processes it (Story 4-2) - simulated here
	// 3. ProcessResult contains emotion changes
	// 4. ChatOverlay applies changes via NPCManager
	// 5. UI updates and history is recorded

	// Setup NPCManager
	npcMgr := manager.NewNPCManager(nil, nil)
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Alice",
		InitialEmotion: manager.NewEmotionState(50, 30, 40),
		Traits: []manager.Trait{
			{
				ID: "paranoid",
			},
		},
	})

	// Setup ChatOverlay
	overlay := NewChatOverlayModel()
	overlay.SetNPCManager(npcMgr)
	overlay.Enter("player", []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.EmotionState{}),
		NewChatParticipant("npc1", "Alice", false, manager.NewEmotionState(50, 30, 40)),
	}, "living_room")

	// Player sends message
	playerMessage := "I saw something strange in the basement..."
	overlay.SetLastPlayerMessage(playerMessage)

	// Simulate ChatProcessor result (Story 4-2 would generate this)
	// JudgeAgent detected the message as potentially frightening
	// Handler applies fear increase to Alice
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: -5, Fear: 15, Stress: 10}, // Paranoid trait amplifies fear
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{
		{
			NPCID:   "npc1",
			Content: "W-what did you see? I knew there was something wrong...",
			Flags:   []ChatFlag{ChatFlagHostile},
		},
	}

	// Apply the result
	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify emotion was updated
	state := npcMgr.GetState("npc1")
	if state == nil {
		t.Fatal("NPC state is nil")
	}

	expectedTrust := 45  // 50 - 5
	expectedFear := 45   // 30 + 15
	expectedStress := 50 // 40 + 10

	if state.Emotion.Trust != expectedTrust {
		t.Errorf("Trust: expected %d, got %d", expectedTrust, state.Emotion.Trust)
	}
	if state.Emotion.Fear != expectedFear {
		t.Errorf("Fear: expected %d, got %d", expectedFear, state.Emotion.Fear)
	}
	if state.Emotion.Stress != expectedStress {
		t.Errorf("Stress: expected %d, got %d", expectedStress, state.Emotion.Stress)
	}

	// Verify relationship changed
	newRelationship := manager.CalculateRelationship(state.Emotion)
	if newRelationship == manager.Friendly {
		t.Error("Relationship should not be Friendly after negative interaction")
	}

	// Verify interaction was recorded
	if len(state.Interactions) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(state.Interactions))
	}
	interaction := state.Interactions[0]
	if interaction.InteractionType != "chat_message" {
		t.Errorf("Wrong interaction type: %s", interaction.InteractionType)
	}
	if interaction.EmotionDelta.Fear != 15 {
		t.Errorf("Wrong fear delta: %d", interaction.EmotionDelta.Fear)
	}

	// Verify UI participant was updated
	participants := overlay.GetActiveParticipants()
	var aliceParticipant *ChatParticipant
	for _, p := range participants {
		if p.ID == "npc1" {
			aliceParticipant = &p
			break
		}
	}
	if aliceParticipant == nil {
		t.Fatal("Alice not found in participants")
	}
	if aliceParticipant.Emotion.Fear != expectedFear {
		t.Errorf("UI Fear not updated: expected %d, got %d", expectedFear, aliceParticipant.Emotion.Fear)
	}

	// Verify message was added
	messages := overlay.GetMessages()
	if len(messages) != 2 { // System message + NPC response
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
	npcMessage := messages[1]
	if npcMessage.Speaker != "npc1" {
		t.Errorf("Wrong speaker: %s", npcMessage.Speaker)
	}
	if npcMessage.Content != "W-what did you see? I knew there was something wrong..." {
		t.Errorf("Wrong message content: %s", npcMessage.Content)
	}
}

// TestEmotionTriggersRelationshipChange tests that emotion changes trigger relationship updates.
// Story 4.6: Emotion updates cascade to relationship changes (via NPCManager)
func TestEmotionTriggersRelationshipChange(t *testing.T) {
	// Setup
	npcMgr := manager.NewNPCManager(nil, nil)
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Bob",
		InitialEmotion: manager.NewEmotionState(55, 35, 40), // Neutral relationship
	})

	overlay := NewChatOverlayModel()
	overlay.SetNPCManager(npcMgr)
	overlay.Enter("player", []ChatParticipant{
		NewChatParticipant("npc1", "Bob", false, manager.NewEmotionState(55, 35, 40)),
	}, "kitchen")
	overlay.SetLastPlayerMessage("I'll help you escape this place")

	// Initial relationship
	initialState := npcMgr.GetState("npc1")
	initialRelationship := initialState.Relationship
	if initialRelationship != manager.Neutral {
		t.Errorf("Initial relationship should be Neutral, got %s", initialRelationship.String())
	}

	// Apply large trust increase to push into Friendly territory
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 20, Fear: -10, Stress: -15}, // Big positive change
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify relationship changed to Friendly
	finalState := npcMgr.GetState("npc1")
	finalRelationship := finalState.Relationship
	if finalRelationship != manager.Friendly {
		t.Errorf("Final relationship should be Friendly (Trust=%d, Fear=%d), got %s",
			finalState.Emotion.Trust, finalState.Emotion.Fear, finalRelationship.String())
	}

	// Verify relationship score increased
	if finalState.RelationshipScore <= initialState.RelationshipScore {
		t.Error("Relationship score should have increased")
	}
}

// TestEmotionTriggersMentalTransition tests that extreme emotions trigger mental state changes.
// Story 4.6: Emotion updates cascade to mental state transitions (via NPCManager)
func TestEmotionTriggersMentalTransition(t *testing.T) {
	// Setup
	npcMgr := manager.NewNPCManager(nil, nil)
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Charlie",
		InitialEmotion: manager.NewEmotionState(40, 50, 55), // Starting anxious territory
	})

	overlay := NewChatOverlayModel()
	overlay.SetNPCManager(npcMgr)
	overlay.Enter("player", []ChatParticipant{
		NewChatParticipant("npc1", "Charlie", false, manager.NewEmotionState(40, 50, 55)),
	}, "dark_room")
	overlay.SetLastPlayerMessage("The thing is coming for us all...")

	// Initial mental state should be Anxious
	initialState := npcMgr.GetState("npc1")
	if initialState.MentalState != manager.Anxious {
		t.Logf("Note: Initial mental state is %s, expected Anxious", initialState.MentalState.String())
	}

	// Apply extreme stress and fear to trigger Corrupted state
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: -30, Fear: 40, Stress: 35}, // Extreme negative change
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify mental state changed to Corrupted
	finalState := npcMgr.GetState("npc1")
	// Note: Mental state transition thresholds are defined in manager.checkMentalStateTransition()
	// If Fear >= 80 and Stress >= 80, should transition to Corrupted
	expectedFear := 90   // 50 + 40
	expectedStress := 90 // 55 + 35

	if finalState.Emotion.Fear != expectedFear {
		t.Errorf("Fear: expected %d, got %d", expectedFear, finalState.Emotion.Fear)
	}
	if finalState.Emotion.Stress != expectedStress {
		t.Errorf("Stress: expected %d, got %d", expectedStress, finalState.Emotion.Stress)
	}

	// Mental state should be Corrupted if thresholds met
	if finalState.Emotion.Fear >= 80 && finalState.Emotion.Stress >= 80 {
		if finalState.MentalState != manager.Corrupted {
			t.Errorf("Mental state should be Corrupted with Fear=%d Stress=%d, got %s",
				finalState.Emotion.Fear, finalState.Emotion.Stress, finalState.MentalState.String())
		}
	}

	// Verify interaction recorded the extreme change
	if len(finalState.Interactions) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(finalState.Interactions))
	}
	interaction := finalState.Interactions[0]
	if interaction.EmotionDelta.Stress != 35 {
		t.Errorf("Wrong stress delta in interaction: %d", interaction.EmotionDelta.Stress)
	}
}

// TestMultipleEmotionUpdates tests sequential emotion updates accumulate correctly.
// Story 4.6: Multiple emotion updates should accumulate
func TestMultipleEmotionUpdates(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Dave",
		InitialEmotion: manager.NewEmotionState(50, 50, 50),
	})

	overlay := NewChatOverlayModel()
	overlay.SetNPCManager(npcMgr)
	overlay.Enter("player", []ChatParticipant{
		NewChatParticipant("npc1", "Dave", false, manager.NewEmotionState(50, 50, 50)),
	}, "hallway")

	emptyResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	// First interaction: slight trust increase
	overlay.SetLastPlayerMessage("Let me help you")
	emotionChanges1 := map[string]manager.EmotionDelta{
		"npc1": {Trust: 10, Fear: -5, Stress: 0},
	}
	overlay.HandleProcessResultDirect(emotionChanges1, emptyResponses, true, "")

	// Second interaction: more trust increase
	overlay.SetLastPlayerMessage("I found some supplies")
	emotionChanges2 := map[string]manager.EmotionDelta{
		"npc1": {Trust: 15, Fear: -5, Stress: -10},
	}
	overlay.HandleProcessResultDirect(emotionChanges2, emptyResponses, true, "")

	// Third interaction: stress event
	overlay.SetLastPlayerMessage("We need to move now!")
	emotionChanges3 := map[string]manager.EmotionDelta{
		"npc1": {Trust: 5, Fear: 10, Stress: 20},
	}
	overlay.HandleProcessResultDirect(emotionChanges3, emptyResponses, true, "")

	// Verify cumulative changes
	state := npcMgr.GetState("npc1")
	expectedTrust := 50 + 10 + 15 + 5  // = 80
	expectedFear := 50 - 5 - 5 + 10    // = 50
	expectedStress := 50 + 0 - 10 + 20 // = 60

	if state.Emotion.Trust != expectedTrust {
		t.Errorf("Trust: expected %d, got %d", expectedTrust, state.Emotion.Trust)
	}
	if state.Emotion.Fear != expectedFear {
		t.Errorf("Fear: expected %d, got %d", expectedFear, state.Emotion.Fear)
	}
	if state.Emotion.Stress != expectedStress {
		t.Errorf("Stress: expected %d, got %d", expectedStress, state.Emotion.Stress)
	}

	// Verify all 3 interactions were recorded
	if len(state.Interactions) != 3 {
		t.Fatalf("Expected 3 interactions, got %d", len(state.Interactions))
	}

	// Verify interaction order
	if state.Interactions[0].EmotionDelta.Trust != 10 {
		t.Error("First interaction not recorded correctly")
	}
	if state.Interactions[2].EmotionDelta.Stress != 20 {
		t.Error("Third interaction not recorded correctly")
	}
}

// TestEmotionClampingEdgeCases tests that emotions are properly clamped to 0-100 range.
// Story 4.6: Edge case - emotions should never exceed bounds
func TestEmotionClampingEdgeCases(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:             "npc1",
		Name:           "Eve",
		InitialEmotion: manager.NewEmotionState(95, 5, 90),
	})

	overlay := NewChatOverlayModel()
	overlay.SetNPCManager(npcMgr)
	overlay.Enter("player", []ChatParticipant{
		NewChatParticipant("npc1", "Eve", false, manager.NewEmotionState(95, 5, 90)),
	}, "room")
	overlay.SetLastPlayerMessage("Extreme message")

	// Apply changes that would push values out of bounds
	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 50, Fear: -50, Stress: 50}, // Trust would be 145, Fear -45, Stress 140
	}
	npcResponses := []struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}{}

	overlay.HandleProcessResultDirect(emotionChanges, npcResponses, true, "")

	// Verify values are clamped
	state := npcMgr.GetState("npc1")
	if state.Emotion.Trust != 100 {
		t.Errorf("Trust should be clamped to 100, got %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 0 {
		t.Errorf("Fear should be clamped to 0, got %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress != 100 {
		t.Errorf("Stress should be clamped to 100, got %d", state.Emotion.Stress)
	}
}
