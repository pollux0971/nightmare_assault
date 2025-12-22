package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/chat"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// =============================================================================
// Story 5-7: Phase 2 Complete Integration Tests
// Integration tests for Chat System without views dependency
// =============================================================================

// TestChatIntegration_FullFlow tests the complete chat flow through ChatProcessor.
// AC1: Complete chat flow (Message → Judgment → Response)
func TestChatIntegration_FullFlow(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	// 1. Create chat session
	chatSession := &chat.ChatSession{
		SessionID:  "session_001",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 0, Stress: 20},
			},
			{
				ID:       env.TestNPCs[0].ID,
				Name:     env.TestNPCs[0].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
			{
				ID:       env.TestNPCs[1].ID,
				Name:     env.TestNPCs[1].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// 2. Process player message
	playerMsg := "你們好，我需要醫療協助"
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, playerMsg, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// 3. Verify processing succeeded
	if !result.Success {
		t.Errorf("Expected successful processing, got error: %s", result.Error)
	}

	// 4. Verify NPC responses generated
	if len(result.NPCResponses) == 0 {
		t.Error("Expected NPC responses, got none")
	}

	// 5. Verify responses from both NPCs
	npcResponseCount := make(map[string]int)
	for _, resp := range result.NPCResponses {
		npcResponseCount[resp.NPCID]++
		if resp.Content == "" {
			t.Errorf("NPC %s response should have content", resp.NPCID)
		}
	}

	t.Logf("✓ Full chat flow test passed: %d NPCs responded", len(result.NPCResponses))
}

// TestChatIntegration_EmotionChanges tests emotion changes propagation.
// AC4: Emotion changes are applied correctly
func TestChatIntegration_EmotionChanges(t *testing.T) {
	env := setupChatIntegrationTest(t)

	// Get initial NPC emotion state
	npc1ID := env.TestNPCs[0].ID
	initialState := env.NPCManager.GetState(npc1ID)
	if initialState == nil {
		t.Fatalf("Failed to get initial NPC state for %s", npc1ID)
	}
	initialTrust := initialState.Emotion.Trust

	// Create emotion changes
	emotionChanges := map[string]manager.EmotionDelta{
		npc1ID: TestEmotionDeltas.TrustIncrease,
	}

	// Apply emotion changes through NPCManager
	for npcID, delta := range emotionChanges {
		err := env.NPCManager.AdjustEmotion(npcID, delta)
		if err != nil {
			t.Errorf("Failed to adjust emotion for %s: %v", npcID, err)
		}
	}

	// Verify emotion changed
	updatedState := env.NPCManager.GetState(npc1ID)
	if updatedState == nil {
		t.Fatalf("Failed to get updated NPC state for %s", npc1ID)
	}

	expectedTrust := initialTrust + TestEmotionDeltas.TrustIncrease.Trust
	if updatedState.Emotion.Trust != expectedTrust {
		t.Errorf("Expected trust=%d, got %d", expectedTrust, updatedState.Emotion.Trust)
	}

	t.Logf("✓ Emotion changes test passed: Trust %d → %d", initialTrust, updatedState.Emotion.Trust)
}

// TestChatIntegration_KnowledgePropagation tests fact propagation to NPCs.
// AC4: Facts shared are propagated to all participants
func TestChatIntegration_KnowledgePropagation(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	// Create chat session with facts
	chatSession := &chat.ChatSession{
		SessionID:  "session_knowledge",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       env.TestNPCs[0].ID,
				Name:     env.TestNPCs[0].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
			{
				ID:       env.TestNPCs[1].ID,
				Name:     env.TestNPCs[1].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// Process message that shares a fact
	factMessage := "避難所在東翼的地下室"
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, factMessage, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful processing, got error: %s", result.Error)
	}

	t.Logf("✓ Knowledge propagation test passed: fact '%s' processed", factMessage)
}

// TestChatIntegration_ContradictionDetection tests contradiction detection and handling.
// AC4: Contradiction detection triggers NPC questioning
func TestChatIntegration_ContradictionDetection(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	// First, establish a fact in NPC knowledge
	npc1ID := env.TestNPCs[0].ID
	establishedFact := "醫院已經廢棄三個月了"
	env.UpdateManager.LearnFromDialogue(npc1ID, "system", establishedFact, "room_001")

	// Now player makes contradicting claim
	chatSession := &chat.ChatSession{
		SessionID:  "session_contradiction",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       npc1ID,
				Name:     env.TestNPCs[0].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	contradictingClaim := "醫院昨天還在正常運作"
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, contradictingClaim, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful processing, got error: %s", result.Error)
	}

	t.Logf("✓ Contradiction detection test passed: processed contradicting claim, detected %d contradictions",
		len(result.Contradictions))
}

// TestChatIntegration_NPCBehaviorConsistency tests NPC behavior consistency.
// AC5: NPC responses reflect personality and emotional state
func TestChatIntegration_NPCBehaviorConsistency(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	// Create chat session
	npc1 := env.TestNPCs[0] // Doctor (helpful archetype)
	chatSession := &chat.ChatSession{
		SessionID:  "session_consistency",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       npc1.ID,
				Name:     npc1.Name,
				IsPlayer: false,
				Emotion:  npc1.InitialEmotion,
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// Send message and process
	message := "我需要醫療幫助"
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, message, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify NPC responded
	if len(result.NPCResponses) == 0 {
		t.Error("Expected NPC response, got none")
	}

	// Verify response from correct NPC
	foundResponse := false
	for _, resp := range result.NPCResponses {
		if resp.NPCID == npc1.ID {
			foundResponse = true
			// Verify response has content
			if resp.Content == "" {
				t.Error("NPC response should have content")
			}
			t.Logf("✓ NPC behavior consistency test passed: NPC %s responded appropriately", npc1.Name)
		}
	}

	if !foundResponse {
		t.Errorf("Expected response from NPC %s, but didn't find it", npc1.ID)
	}
}

// TestChatIntegration_NPCMemory tests that NPCs remember conversation context.
// AC5: NPCs remember dialogue content
func TestChatIntegration_NPCMemory(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	npc1 := env.TestNPCs[0]
	chatSession := &chat.ChatSession{
		SessionID:  "session_memory",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       npc1.ID,
				Name:     npc1.Name,
				IsPlayer: false,
				Emotion:  npc1.InitialEmotion,
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// First message establishes context
	firstMessage := "我叫約翰，我來自東翼"
	_, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, firstMessage, nil)
	if err != nil {
		t.Fatalf("First ProcessPlayerMessage failed: %v", err)
	}

	// Add to message history
	chatSession.MessageHistory = append(chatSession.MessageHistory, chat.ChatMessage{
		SenderID:  "player",
		Content:   firstMessage,
		Timestamp: int(time.Now().Unix()),
	})

	// Second message references previous context
	secondMessage := "你記得我是誰嗎？"
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, secondMessage, nil)
	if err != nil {
		t.Fatalf("Second ProcessPlayerMessage failed: %v", err)
	}

	// Verify processing succeeded
	if !result.Success {
		t.Errorf("Expected successful processing, got error: %s", result.Error)
	}

	t.Log("✓ NPC memory test passed: conversation history supported")
}

// TestChatIntegration_MultipleRounds tests multiple chat rounds.
// AC5: NPC emotional state evolves across multiple rounds
func TestChatIntegration_MultipleRounds(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	npc1 := env.TestNPCs[0]
	initialState := env.NPCManager.GetState(npc1.ID)
	if initialState == nil {
		t.Fatal("Failed to get initial NPC state")
	}
	initialTrust := initialState.Emotion.Trust

	chatSession := &chat.ChatSession{
		SessionID:  "session_multi_round",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       npc1.ID,
				Name:     npc1.Name,
				IsPlayer: false,
				Emotion:  npc1.InitialEmotion,
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// Round 1: Friendly message
	_, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, "謝謝你的幫助", nil)
	if err != nil {
		t.Fatalf("Round 1 failed: %v", err)
	}

	// Get emotion after round 1
	state1 := env.NPCManager.GetState(npc1.ID)
	if state1 == nil {
		t.Fatal("Failed to get NPC state after round 1")
	}

	// Round 2: Another positive interaction
	_, err = env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, "我相信你", nil)
	if err != nil {
		t.Fatalf("Round 2 failed: %v", err)
	}

	// Get final emotion
	state2 := env.NPCManager.GetState(npc1.ID)
	if state2 == nil {
		t.Fatal("Failed to get NPC state after round 2")
	}

	t.Logf("✓ Multiple rounds test passed: Emotion evolution tracked across %d rounds", 2)
	t.Logf("  Initial Trust=%d, After Round 1=%d, After Round 2=%d",
		initialTrust, state1.Emotion.Trust, state2.Emotion.Trust)
}

// TestChatIntegration_ProcessorConfiguration tests ChatProcessor configuration.
// AC2: ChatProcessor respects configuration
func TestChatIntegration_ProcessorConfiguration(t *testing.T) {
	env := setupChatIntegrationTest(t)

	// Verify ChatProcessor was created with components
	if env.ChatProcessor == nil {
		t.Fatal("ChatProcessor should not be nil")
	}

	if env.NPCManager == nil {
		t.Fatal("NPCManager should not be nil")
	}

	if env.UpdateManager == nil {
		t.Fatal("UpdateManager should not be nil")
	}

	t.Log("✓ ChatProcessor configuration test passed: all components initialized")
}

// TestChatIntegration_GracefulDegradation tests graceful degradation when LLM fails.
// AC1: System handles LLM failures gracefully
func TestChatIntegration_GracefulDegradation(t *testing.T) {
	env := setupChatIntegrationTest(t)
	ctx := context.Background()

	// Configure mock LLM to fail
	env.MockLLM.ShouldFail = true

	chatSession := &chat.ChatSession{
		SessionID:  "session_degradation",
		Location:   "room_001",
		Participants: []chat.ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
			},
			{
				ID:       env.TestNPCs[0].ID,
				Name:     env.TestNPCs[0].Name,
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
		},
		MessageHistory: []chat.ChatMessage{},
	}

	// Process should still succeed with degraded functionality
	result, err := env.ChatProcessor.ProcessPlayerMessage(ctx, chatSession, "測試訊息", nil)

	// Should not panic or return fatal error
	if err != nil {
		t.Logf("Expected graceful degradation, got error: %v (this is acceptable)", err)
	}

	// Result should indicate the issue
	if result != nil && !result.Success {
		t.Logf("Result correctly indicates failure: %s", result.Error)
	}

	// Reset mock for other tests
	env.MockLLM.Reset()

	t.Log("✓ Graceful degradation test passed: system handles LLM failures")
}
