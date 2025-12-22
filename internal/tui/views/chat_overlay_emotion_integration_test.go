package views

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 4.6: Emotion Update Integration - Unit Tests
// ==========================================================================

// TestApplyProcessResult_EmotionUpdates tests AC1 & AC2:
// ProcessResult contains emotion changes for each NPC,
// and ChatOverlay applies emotion changes upon receiving ProcessResult.
func TestApplyProcessResult_EmotionUpdates(t *testing.T) {
	// Create chat overlay with NPC manager
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add test NPCs
	npc1 := createTestNPCProfile("npc_001", "Alice", 60, 30, 40)
	npc2 := createTestNPCProfile("npc_002", "Bob", 50, 50, 50)
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add npc1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add npc2: %v", err)
	}

	m.SetNPCManager(npcMgr)

	// Set up participants in chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
		NewChatParticipant("npc_002", "Bob", false, manager.NewEmotionState(50, 50, 50)),
	}
	m.Enter("player", participants, "TestRoom")

	// Create ProcessResult with emotion changes
	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: 10, Fear: -5, Stress: -5},
			"npc_002": {Trust: -15, Fear: 10, Stress: 5},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "Thanks for being honest!",
				Emotion: manager.NewEmotionState(70, 25, 35),
			},
			{
				NPCID:   "npc_002",
				Content: "I don't trust you...",
				Emotion: manager.NewEmotionState(35, 60, 55),
			},
		},
		Success: true,
	}

	// AC1 & AC2: Apply ProcessResult
	m.ApplyProcessResult(result)

	// Verify emotion changes were applied to NPC manager
	state1 := npcMgr.GetState("npc_001")
	if state1 == nil {
		t.Fatal("NPC 001 state should exist")
	}
	// Original: Trust=60, new should be 70 (60 + 10)
	if state1.Emotion.Trust != 70 {
		t.Errorf("NPC 001 Trust = %d, want 70", state1.Emotion.Trust)
	}
	// Original: Fear=30, new should be 25 (30 - 5)
	if state1.Emotion.Fear != 25 {
		t.Errorf("NPC 001 Fear = %d, want 25", state1.Emotion.Fear)
	}

	state2 := npcMgr.GetState("npc_002")
	if state2 == nil {
		t.Fatal("NPC 002 state should exist")
	}
	// Original: Trust=50, new should be 35 (50 - 15)
	if state2.Emotion.Trust != 35 {
		t.Errorf("NPC 002 Trust = %d, want 35", state2.Emotion.Trust)
	}
	// Original: Fear=50, new should be 60 (50 + 10)
	if state2.Emotion.Fear != 60 {
		t.Errorf("NPC 002 Fear = %d, want 60", state2.Emotion.Fear)
	}

	// AC3: Verify participant emotions were updated in UI
	participant1 := m.GetParticipant("npc_001")
	if participant1 == nil {
		t.Fatal("Participant npc_001 should exist")
	}
	if participant1.Emotion.Trust != 70 {
		t.Errorf("Participant npc_001 UI Trust = %d, want 70", participant1.Emotion.Trust)
	}

	participant2 := m.GetParticipant("npc_002")
	if participant2 == nil {
		t.Fatal("Participant npc_002 should exist")
	}
	if participant2.Emotion.Trust != 35 {
		t.Errorf("Participant npc_002 UI Trust = %d, want 35", participant2.Emotion.Trust)
	}
}

// TestApplyProcessResult_InteractionRecording tests AC4:
// Emotion changes are recorded to NPCInteraction history.
func TestApplyProcessResult_InteractionRecording(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add test NPC
	npc := createTestNPCProfile("npc_001", "Alice", 60, 30, 40)
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add npc: %v", err)
	}

	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
	}
	m.Enter("player", participants, "TestRoom")

	// Get initial interaction count
	initialState := npcMgr.GetState("npc_001")
	initialInteractionCount := len(initialState.Interactions)

	// Create ProcessResult with emotion change
	emotionDelta := manager.EmotionDelta{Trust: 15, Fear: -10, Stress: -5}
	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": emotionDelta,
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "I really appreciate your help!",
				Emotion: manager.NewEmotionState(75, 20, 35),
			},
		},
		Success: true,
	}

	// AC4: Apply ProcessResult and record interaction
	m.ApplyProcessResult(result)

	// Verify interaction was recorded
	updatedState := npcMgr.GetState("npc_001")
	if len(updatedState.Interactions) != initialInteractionCount+1 {
		t.Errorf("Interaction count = %d, want %d", len(updatedState.Interactions), initialInteractionCount+1)
	}

	// Verify interaction details
	if len(updatedState.Interactions) > 0 {
		lastInteraction := updatedState.Interactions[len(updatedState.Interactions)-1]
		// The existing implementation uses "chat_message" as the interaction type
		if lastInteraction.InteractionType != "chat_message" {
			t.Errorf("Interaction type = %s, want 'chat_message'", lastInteraction.InteractionType)
		}
		if lastInteraction.EmotionDelta.Trust != emotionDelta.Trust {
			t.Errorf("Recorded delta Trust = %d, want %d", lastInteraction.EmotionDelta.Trust, emotionDelta.Trust)
		}
		if lastInteraction.EmotionDelta.Fear != emotionDelta.Fear {
			t.Errorf("Recorded delta Fear = %d, want %d", lastInteraction.EmotionDelta.Fear, emotionDelta.Fear)
		}
		if lastInteraction.EmotionDelta.Stress != emotionDelta.Stress {
			t.Errorf("Recorded delta Stress = %d, want %d", lastInteraction.EmotionDelta.Stress, emotionDelta.Stress)
		}
	}
}

// TestApplyProcessResult_NPCResponses tests that NPC responses are added to chat.
func TestApplyProcessResult_NPCResponses(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npc1 := createTestNPCProfile("npc_001", "Alice", 60, 30, 40)
	npc2 := createTestNPCProfile("npc_002", "Bob", 50, 50, 50)
	npcMgr.AddNPC(npc1)
	npcMgr.AddNPC(npc2)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
		NewChatParticipant("npc_002", "Bob", false, manager.NewEmotionState(50, 50, 50)),
	}
	m.Enter("player", participants, "TestRoom")

	initialMessageCount := len(m.messages)

	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: 5, Fear: 0, Stress: -5},
			"npc_002": {Trust: -10, Fear: 5, Stress: 10},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:        "npc_001",
				Content:      "That's a good point.",
				Emotion:      manager.NewEmotionState(65, 30, 35),
				Flags:        []ChatFlag{ChatFlagRevelation},
				UsedFallback: false,
			},
			{
				NPCID:        "npc_002",
				Content:      "I disagree completely!",
				Emotion:      manager.NewEmotionState(40, 55, 60),
				Flags:        []ChatFlag{ChatFlagHostile},
				UsedFallback: false,
			},
		},
		Success: true,
	}

	m.ApplyProcessResult(result)

	// Verify NPC messages were added
	if len(m.messages) != initialMessageCount+2 {
		t.Errorf("Message count = %d, want %d", len(m.messages), initialMessageCount+2)
	}

	// Verify message contents and flags
	foundAlice := false
	foundBob := false
	for _, msg := range m.messages {
		if msg.Speaker == "npc_001" && msg.Content == "That's a good point." {
			foundAlice = true
			if !msg.HasFlag(ChatFlagRevelation) {
				t.Error("Alice's message should have ChatFlagRevelation")
			}
		}
		if msg.Speaker == "npc_002" && msg.Content == "I disagree completely!" {
			foundBob = true
			if !msg.HasFlag(ChatFlagHostile) {
				t.Error("Bob's message should have ChatFlagHostile")
			}
		}
	}

	if !foundAlice {
		t.Error("Alice's response message not found")
	}
	if !foundBob {
		t.Error("Bob's response message not found")
	}
}

// TestApplyProcessResult_ContradictionHandling tests contradiction detection integration.
func TestApplyProcessResult_ContradictionHandling(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npc := createTestNPCProfile("npc_001", "Alice", 70, 20, 30)
	npcMgr.AddNPC(npc)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(70, 20, 30)),
	}
	m.Enter("player", participants, "TestRoom")

	// Create ProcessResult with contradictions
	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: -20, Fear: 10, Stress: 15},
		},
		Contradictions: []ContradictionStruct{
			{
				NPCID:          "npc_001",
				PlayerClaim:    "The door was red",
				NPCBelief:      "The door was blue",
				Severity:       "high",
				SuggestedDelta: manager.EmotionDelta{Trust: -20, Fear: 10, Stress: 15},
			},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "That's not what I remember at all!",
				Emotion: manager.NewEmotionState(50, 30, 45),
				Flags:   []ChatFlag{ChatFlagContradiction},
			},
		},
		Success: true,
	}

	m.ApplyProcessResult(result)

	// Verify contradiction was handled and emotions updated
	state := npcMgr.GetState("npc_001")
	if state.Emotion.Trust != 50 { // 70 - 20
		t.Errorf("Trust after contradiction = %d, want 50", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 30 { // 20 + 10
		t.Errorf("Fear after contradiction = %d, want 30", state.Emotion.Fear)
	}

	// Verify UI was updated
	participant := m.GetParticipant("npc_001")
	if participant.Emotion.Trust != 50 {
		t.Errorf("UI Trust after contradiction = %d, want 50", participant.Emotion.Trust)
	}
}

// TestApplyProcessResult_NoNPCManager tests graceful handling when NPCManager is nil.
func TestApplyProcessResult_NoNPCManager(t *testing.T) {
	m := NewChatOverlayModel()
	// Don't set NPC manager

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
	}
	m.Enter("player", participants, "TestRoom")

	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: 10, Fear: -5, Stress: -5},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "Thanks!",
				Emotion: manager.NewEmotionState(70, 25, 35),
			},
		},
		Success: true,
	}

	// Should not panic
	m.ApplyProcessResult(result)

	// At least NPC responses should be added
	found := false
	for _, msg := range m.messages {
		if msg.Speaker == "npc_001" && msg.Content == "Thanks!" {
			found = true
			break
		}
	}
	if !found {
		t.Error("NPC response should still be added even without NPC manager")
	}
}

// TestApplyProcessResult_EmptyResult tests handling of empty ProcessResult.
func TestApplyProcessResult_EmptyResult(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMessageCount := len(m.messages)

	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{},
		NPCResponses:   []NPCResponseStruct{},
		Success:        true,
	}

	// Should not panic
	m.ApplyProcessResult(result)

	// No messages should be added
	if len(m.messages) != initialMessageCount {
		t.Errorf("Empty result should not add messages, got %d messages", len(m.messages))
	}
}

// TestApplyProcessResult_FailedProcessing tests handling of failed ProcessResult.
func TestApplyProcessResult_FailedProcessing(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npc := createTestNPCProfile("npc_001", "Alice", 60, 30, 40)
	npcMgr.AddNPC(npc)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
	}
	m.Enter("player", participants, "TestRoom")

	initialMessageCount := len(m.messages)

	result := &ProcessResultStruct{
		Success: false,
		Error:   "LLM API failed",
	}

	m.ApplyProcessResult(result)

	// Should add error system message
	if len(m.messages) != initialMessageCount+1 {
		t.Errorf("Failed result should add error message, got %d messages", len(m.messages))
	}

	// Verify error message was added
	lastMsg := m.messages[len(m.messages)-1]
	if lastMsg.Type != ChatMessageSystem {
		t.Error("Error message should be system type")
	}
	if lastMsg.Content != "處理失敗: LLM API failed" {
		t.Errorf("Error message content = %q, want error notification", lastMsg.Content)
	}
}

// TestApplyProcessResult_MultipleEmotionUpdates tests multiple emotion updates in sequence.
func TestApplyProcessResult_MultipleEmotionUpdates(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	npc := createTestNPCProfile("npc_001", "Alice", 50, 50, 50)
	npcMgr.AddNPC(npc)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(50, 50, 50)),
	}
	m.Enter("player", participants, "TestRoom")

	// First update: positive interaction
	result1 := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: 20, Fear: -10, Stress: -10},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "Thank you so much!",
				Emotion: manager.NewEmotionState(70, 40, 40),
			},
		},
		Success: true,
	}
	m.ApplyProcessResult(result1)

	// Verify first update
	state := npcMgr.GetState("npc_001")
	if state.Emotion.Trust != 70 {
		t.Errorf("After first update, Trust = %d, want 70", state.Emotion.Trust)
	}

	// Second update: negative interaction
	result2 := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: -30, Fear: 20, Stress: 25},
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "I can't believe you did that!",
				Emotion: manager.NewEmotionState(40, 60, 65),
			},
		},
		Success: true,
	}
	m.ApplyProcessResult(result2)

	// Verify cumulative updates
	state = npcMgr.GetState("npc_001")
	if state.Emotion.Trust != 40 { // 70 - 30
		t.Errorf("After second update, Trust = %d, want 40", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 60 { // 40 + 20
		t.Errorf("After second update, Fear = %d, want 60", state.Emotion.Fear)
	}

	// Verify UI reflects final state
	participant := m.GetParticipant("npc_001")
	if participant.Emotion.Trust != 40 {
		t.Errorf("UI Trust = %d, want 40", participant.Emotion.Trust)
	}

	// Verify both interactions were recorded
	if len(state.Interactions) < 2 {
		t.Errorf("Should have at least 2 interactions recorded, got %d", len(state.Interactions))
	}
}

// TestApplyProcessResult_ClampingBehavior tests emotion value clamping (0-100).
func TestApplyProcessResult_ClampingBehavior(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Start with extreme values
	npc := createTestNPCProfile("npc_001", "Alice", 95, 5, 10)
	npcMgr.AddNPC(npc)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(95, 5, 10)),
	}
	m.Enter("player", participants, "TestRoom")

	// Apply extreme positive delta
	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001": {Trust: 50, Fear: -50, Stress: -50}, // Would exceed bounds
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "You're amazing!",
				Emotion: manager.NewEmotionState(100, 0, 0),
			},
		},
		Success: true,
	}
	m.ApplyProcessResult(result)

	// Verify clamping
	state := npcMgr.GetState("npc_001")
	if state.Emotion.Trust != 100 {
		t.Errorf("Trust should be clamped to 100, got %d", state.Emotion.Trust)
	}
	if state.Emotion.Fear != 0 {
		t.Errorf("Fear should be clamped to 0, got %d", state.Emotion.Fear)
	}
	if state.Emotion.Stress != 0 {
		t.Errorf("Stress should be clamped to 0, got %d", state.Emotion.Stress)
	}
}

// TestApplyProcessResult_NonExistentNPC tests handling of emotion updates for non-existent NPCs.
func TestApplyProcessResult_NonExistentNPC(t *testing.T) {
	m := NewChatOverlayModel()
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add only one NPC
	npc := createTestNPCProfile("npc_001", "Alice", 60, 30, 40)
	npcMgr.AddNPC(npc)
	m.SetNPCManager(npcMgr)

	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(60, 30, 40)),
	}
	m.Enter("player", participants, "TestRoom")

	// Try to update non-existent NPC
	result := &ProcessResultStruct{
		EmotionChanges: map[string]manager.EmotionDelta{
			"npc_001":     {Trust: 10, Fear: -5, Stress: -5},
			"npc_999":     {Trust: 20, Fear: -10, Stress: -10}, // Non-existent
			"npc_invalid": {Trust: 5, Fear: 0, Stress: 0},      // Non-existent
		},
		NPCResponses: []NPCResponseStruct{
			{
				NPCID:   "npc_001",
				Content: "Thanks!",
				Emotion: manager.NewEmotionState(70, 25, 35),
			},
		},
		Success: true,
	}

	// Should not panic
	m.ApplyProcessResult(result)

	// Verify valid NPC was updated
	state := npcMgr.GetState("npc_001")
	if state.Emotion.Trust != 70 {
		t.Errorf("Valid NPC Trust = %d, want 70", state.Emotion.Trust)
	}

	// Verify non-existent NPCs were gracefully ignored
	state999 := npcMgr.GetState("npc_999")
	if state999 != nil {
		t.Error("Non-existent NPC should not be created")
	}
}

// ==========================================================================
// Helper Functions
// ==========================================================================

// createTestNPCProfile creates a test NPC profile for unit tests.
func createTestNPCProfile(id, name string, trust, fear, stress int) *manager.NPCProfile {
	return &manager.NPCProfile{
		ID:             id,
		Name:           name,
		Archetype:      "Test",
		Appearance:     "Test appearance",
		Backstory:      "Test backstory",
		Skills:         []string{},
		Inventory:      []string{},
		Secret:         "",
		SecretTier:     0,
		Traits:         []manager.Trait{},
		LinkedSeeds:    []string{},
		DeathBeat:      0,
		InitialEmotion: manager.NewEmotionState(trust, fear, stress),
	}
}
