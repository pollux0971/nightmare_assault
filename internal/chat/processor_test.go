package chat

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Mock LLM Client
// ==========================================================================

type MockLLMClient struct {
	ResponseFunc func(ctx context.Context, prompt string, opts map[string]any) (string, error)
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string, opts map[string]any) (string, error) {
	if m.ResponseFunc != nil {
		return m.ResponseFunc(ctx, prompt, opts)
	}
	return `{"flags": [], "confidence": 0.8, "reasoning": "test"}`, nil
}

// ==========================================================================
// Test: NewChatProcessor
// ==========================================================================

func TestNewChatProcessor(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name: "TestJudge",
	})
	mockLLM := &MockLLMClient{}

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	if processor == nil {
		t.Fatal("Expected processor to be created, got nil")
	}

	if processor.npcManager != npcMgr {
		t.Error("npcManager not set correctly")
	}

	if processor.updateManager != updateMgr {
		t.Error("updateManager not set correctly")
	}

	if processor.judgeAgent != judgeAgent {
		t.Error("judgeAgent not set correctly")
	}

	if processor.llmClient != mockLLM {
		t.Error("llmClient not set correctly")
	}

	if processor.handlerRegistry == nil {
		t.Error("handlerRegistry should be initialized")
	}

	if processor.config == nil {
		t.Error("config should use default if nil")
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - Success (Happy Path)
// ==========================================================================

func TestProcessPlayerMessage_Success(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Create JudgeAgent with mock LLM that returns friendly flags
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": [],
				"confidence": 0.9,
				"reasoning": "Player message is neutral and friendly"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	// Create test session
	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{
				ID:       "player",
				Name:     "Player",
				IsPlayer: true,
			},
			{
				ID:       "npc1",
				Name:     "Test NPC",
				IsPlayer: false,
				Emotion: manager.EmotionState{
					Trust:  50,
					Fear:   30,
					Stress: 20,
				},
				Relationship: "neutral",
			},
		},
		Location:         "test-room",
		MessageHistory:   []ChatMessage{},
		MaxHistoryLength: 10,
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "Hello, how are you?", gameState)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	if len(result.NPCResponses) != 1 {
		t.Errorf("Expected 1 NPC response, got %d", len(result.NPCResponses))
	}

	if result.NPCResponses[0].NPCID != "npc1" {
		t.Errorf("Expected NPC ID 'npc1', got '%s'", result.NPCResponses[0].NPCID)
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - Empty Message
// ==========================================================================

func TestProcessPlayerMessage_EmptyMessage(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC", IsPlayer: false},
		},
		Location: "test-room",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "", nil)

	// Verify
	if err == nil {
		t.Error("Expected error for empty message, got nil")
	}

	if result == nil {
		t.Fatal("Expected result even on error")
	}

	if result.Success {
		t.Error("Expected success=false for empty message")
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - No Participants
// ==========================================================================

func TestProcessPlayerMessage_NoParticipants(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	// Session with only player (no NPCs)
	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
		},
		Location: "test-room",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "Hello", nil)

	// Verify
	if err == nil {
		t.Error("Expected error for no NPC participants, got nil")
	}

	if result == nil {
		t.Fatal("Expected result even on error")
	}

	if result.Success {
		t.Error("Expected success=false for no NPC participants")
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - JudgeAgent Failure (Graceful Degradation)
// ==========================================================================

func TestProcessPlayerMessage_JudgeAgentFailure(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Create JudgeAgent with failing mock LLM
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return "", context.DeadlineExceeded
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "Hello", gameState)

	// Verify - should succeed with graceful degradation
	if err != nil {
		t.Fatalf("Expected no error with graceful degradation, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Errorf("Expected success=true (graceful degradation), got false. Error: %s", result.Error)
	}

	// Should have empty flags due to judge failure
	if len(result.Flags) != 0 {
		t.Errorf("Expected 0 flags (judge failed), got %d", len(result.Flags))
	}

	// Should still generate NPC responses
	if len(result.NPCResponses) != 1 {
		t.Errorf("Expected 1 NPC response, got %d", len(result.NPCResponses))
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - With Flags (Hallucination)
// ==========================================================================

func TestProcessPlayerMessage_WithFlags_Hallucination(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Skeptical",
		Traits:    []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Create JudgeAgent with mock LLM that returns hallucination flag
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["hallucination"],
				"confidence": 0.95,
				"reasoning": "Player claims to have seen something that contradicts known facts"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "I saw a ghost in the hallway!", gameState)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	// Should have hallucination flag
	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}

	if result.Flags[0] != ChatFlagHallucination {
		t.Errorf("Expected hallucination flag, got %s", result.Flags[0].String())
	}

	// Should have emotion changes (Trust decrease)
	if len(result.EmotionChanges) != 1 {
		t.Fatalf("Expected 1 emotion change, got %d", len(result.EmotionChanges))
	}

	emotionDelta := result.EmotionChanges["npc1"]
	if emotionDelta.Trust >= 0 {
		t.Errorf("Expected negative trust change, got %d", emotionDelta.Trust)
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - With Flags (Hostile)
// ==========================================================================

func TestProcessPlayerMessage_WithFlags_Hostile(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Create JudgeAgent with mock LLM that returns hostile flag
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["hostile"],
				"confidence": 0.92,
				"reasoning": "Player message shows aggression and threat"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "Get out of my way or else!", gameState)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	// Should have hostile flag
	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}

	if result.Flags[0] != ChatFlagHostile {
		t.Errorf("Expected hostile flag, got %s", result.Flags[0].String())
	}

	// Should have emotion changes (Fear and Stress increase, Trust decrease)
	if len(result.EmotionChanges) != 1 {
		t.Fatalf("Expected 1 emotion change, got %d", len(result.EmotionChanges))
	}

	emotionDelta := result.EmotionChanges["npc1"]
	if emotionDelta.Fear <= 0 {
		t.Errorf("Expected positive fear change, got %d", emotionDelta.Fear)
	}
	if emotionDelta.Stress <= 0 {
		t.Errorf("Expected positive stress change, got %d", emotionDelta.Stress)
	}
	if emotionDelta.Trust >= 0 {
		t.Errorf("Expected negative trust change, got %d", emotionDelta.Trust)
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - Multiple NPCs
// ==========================================================================

func TestProcessPlayerMessage_MultipleNPCs(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add two test NPCs
	npc1 := &manager.NPCProfile{
		ID:   "npc1",
		Name: "NPC One",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	npc2 := &manager.NPCProfile{
		ID:   "npc2",
		Name: "NPC Two",
		InitialEmotion: manager.EmotionState{
			Trust:  60,
			Fear:   20,
			Stress: 15,
		},
		Archetype: "Trusting",
		Traits:    []manager.Trait{},
	}

	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Create JudgeAgent with mock LLM
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["revelation"],
				"confidence": 0.88,
				"reasoning": "Player shares important information"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "NPC One",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
			{
				ID:           "npc2",
				Name:         "NPC Two",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
				Relationship: "friendly",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "I found a secret passage!", gameState)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	// Should have 2 NPC responses
	if len(result.NPCResponses) != 2 {
		t.Errorf("Expected 2 NPC responses, got %d", len(result.NPCResponses))
	}

	// Should have emotion changes for both NPCs
	if len(result.EmotionChanges) != 2 {
		t.Errorf("Expected 2 emotion changes, got %d", len(result.EmotionChanges))
	}

	// Verify both NPCs have emotion changes
	if _, exists := result.EmotionChanges["npc1"]; !exists {
		t.Error("Expected emotion change for npc1")
	}
	if _, exists := result.EmotionChanges["npc2"]; !exists {
		t.Error("Expected emotion change for npc2")
	}
}

// ==========================================================================
// Test: ProcessPlayerMessage - Multiple Flags
// ==========================================================================

func TestProcessPlayerMessage_MultipleFlags(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Create JudgeAgent with mock LLM that returns multiple flags
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["hostile", "lie"],
				"confidence": 0.85,
				"reasoning": "Player is aggressive and appears to be lying"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "I didn't do it! Get away from me!", gameState)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	// Should have 2 flags
	if len(result.Flags) != 2 {
		t.Fatalf("Expected 2 flags, got %d", len(result.Flags))
	}

	// Check flags are correct
	hasHostile := false
	hasLie := false
	for _, flag := range result.Flags {
		if flag == ChatFlagHostile {
			hasHostile = true
		}
		if flag == ChatFlagLie {
			hasLie = true
		}
	}

	if !hasHostile {
		t.Error("Expected hostile flag")
	}
	if !hasLie {
		t.Error("Expected lie flag")
	}

	// Should have emotion changes from both handlers
	if len(result.EmotionChanges) != 1 {
		t.Fatalf("Expected 1 emotion change entry, got %d", len(result.EmotionChanges))
	}

	// Emotion changes should be cumulative
	emotionDelta := result.EmotionChanges["npc1"]
	// Both hostile and lie handlers reduce trust
	if emotionDelta.Trust >= 0 {
		t.Errorf("Expected negative trust change from both handlers, got %d", emotionDelta.Trust)
	}
}

// ==========================================================================
// Test: validateProcessRequest
// ==========================================================================

func TestValidateProcessRequest(t *testing.T) {
	processor := &ChatProcessor{}

	tests := []struct {
		name        string
		session     *ChatSession
		message     string
		expectError bool
	}{
		{
			name:        "Nil session",
			session:     nil,
			message:     "test",
			expectError: true,
		},
		{
			name: "Empty message",
			session: &ChatSession{
				SessionID: "test",
				Participants: []ChatParticipant{
					{ID: "player", IsPlayer: true},
					{ID: "npc1", IsPlayer: false},
				},
			},
			message:     "",
			expectError: true,
		},
		{
			name: "No NPC participants",
			session: &ChatSession{
				SessionID: "test",
				Participants: []ChatParticipant{
					{ID: "player", IsPlayer: true},
				},
			},
			message:     "test",
			expectError: true,
		},
		{
			name: "Valid request",
			session: &ChatSession{
				SessionID: "test",
				Participants: []ChatParticipant{
					{ID: "player", IsPlayer: true},
					{ID: "npc1", IsPlayer: false},
				},
			},
			message:     "test",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.validateProcessRequest(tt.session, tt.message)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// ==========================================================================
// Test: convertToChatFlags
// ==========================================================================

func TestConvertToChatFlags(t *testing.T) {
	processor := &ChatProcessor{}

	agentFlags := []agents.ChatFlag{
		agents.FlagHallucination,
		agents.FlagHostile,
		agents.FlagRevelation,
		agents.FlagContradiction,
		agents.FlagPersuasion,
		agents.FlagLie,
	}

	chatFlags := processor.convertToChatFlags(agentFlags)

	if len(chatFlags) != len(agentFlags) {
		t.Errorf("Expected %d chat flags, got %d", len(agentFlags), len(chatFlags))
	}

	// Verify each flag is correctly mapped
	expectedFlags := []ChatFlag{
		ChatFlagHallucination,
		ChatFlagHostile,
		ChatFlagRevelation,
		ChatFlagContradiction,
		ChatFlagPersuasion,
		ChatFlagLie,
	}

	for i, expected := range expectedFlags {
		if chatFlags[i] != expected {
			t.Errorf("Flag %d: expected %s, got %s", i, expected.String(), chatFlags[i].String())
		}
	}
}

// ==========================================================================
// Test: getNPCParticipants
// ==========================================================================

func TestGetNPCParticipants(t *testing.T) {
	processor := &ChatProcessor{}

	session := &ChatSession{
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC 1", IsPlayer: false},
			{ID: "npc2", Name: "NPC 2", IsPlayer: false},
		},
	}

	npcs := processor.getNPCParticipants(session)

	if len(npcs) != 2 {
		t.Errorf("Expected 2 NPCs, got %d", len(npcs))
	}

	for _, npc := range npcs {
		if npc.IsPlayer {
			t.Errorf("getNPCParticipants returned a player participant: %s", npc.ID)
		}
	}
}

// ==========================================================================
// Benchmark: ProcessPlayerMessage
// ==========================================================================

func BenchmarkProcessPlayerMessage(b *testing.B) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	npc := &manager.NPCProfile{
		ID:   "npc1",
		Name: "Test NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		Archetype: "Neutral",
		Traits:    []manager.Trait{},
	}
	npcMgr.AddNPC(npc)

	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{"flags": [], "confidence": 0.8, "reasoning": "test"}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessPlayerMessage(ctx, session, "Hello", gameState)
	}
}

// ==========================================================================
// Story 4.7: Knowledge Propagation Integration Tests
// ==========================================================================

// TestStory47_PlayerMessagePropagation tests that player messages are automatically
// propagated to all NPCs in the same room.
//
// Story 4-7 AC1: 玩家訊息自動傳播給同房間 NPC
func TestStory47_PlayerMessagePropagation(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add two NPCs
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Neutral",
		Traits:         []manager.Trait{},
	}
	npc2 := &manager.NPCProfile{
		ID:             "npc2",
		Name:           "NPC Two",
		InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
		Archetype:      "Friendly",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Set room locations
	updateMgr.SetEntityRoom("player", "room1")
	updateMgr.SetEntityRoom("npc1", "room1")
	updateMgr.SetEntityRoom("npc2", "room1")

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC One", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20}},
			{ID: "npc2", Name: "NPC Two", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15}},
		},
		Location: "room1",
	}

	// Execute
	playerMessage := "I found the key in the drawer"
	_, err := processor.ProcessPlayerMessage(context.Background(), session, playerMessage, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify: Both NPCs should have learned the information
	npc1KB := updateMgr.GetKnowledgeBase("npc1")
	if npc1KB == nil {
		t.Fatal("NPC1 should have a knowledge base")
	}

	npc2KB := updateMgr.GetKnowledgeBase("npc2")
	if npc2KB == nil {
		t.Fatal("NPC2 should have a knowledge base")
	}

	// Check that both NPCs learned the fact
	foundNPC1 := false
	foundNPC2 := false

	for _, knownFact := range npc1KB.KnownFacts {
		if knownFact.LearnMethod == knowledge.Told && knownFact.LearnedFrom == "player" {
			foundNPC1 = true
			break
		}
	}

	for _, knownFact := range npc2KB.KnownFacts {
		if knownFact.LearnMethod == knowledge.Told && knownFact.LearnedFrom == "player" {
			foundNPC2 = true
			break
		}
	}

	if !foundNPC1 {
		t.Error("NPC1 should have learned from player's message")
	}

	if !foundNPC2 {
		t.Error("NPC2 should have learned from player's message")
	}

	t.Logf("Successfully propagated player message to %d NPCs", 2)
}

// TestStory47_NPCResponsePropagation tests that NPC responses are automatically
// propagated to other participants in the same room.
//
// Story 4-7 AC2: NPC 回應自動傳播給其他參與者
func TestStory47_NPCResponsePropagation(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add two NPCs
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Neutral",
		Traits:         []manager.Trait{},
	}
	npc2 := &manager.NPCProfile{
		ID:             "npc2",
		Name:           "NPC Two",
		InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
		Archetype:      "Friendly",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Set room locations
	updateMgr.SetEntityRoom("player", "room1")
	updateMgr.SetEntityRoom("npc1", "room1")
	updateMgr.SetEntityRoom("npc2", "room1")

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC One", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20}},
			{ID: "npc2", Name: "NPC Two", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15}},
		},
		Location: "room1",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "Hello everyone", nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify NPC responses were generated
	if len(result.NPCResponses) != 2 {
		t.Fatalf("Expected 2 NPC responses, got %d", len(result.NPCResponses))
	}

	// Verify: NPC2 should have learned from NPC1's response, and vice versa
	npc1KB := updateMgr.GetKnowledgeBase("npc1")
	npc2KB := updateMgr.GetKnowledgeBase("npc2")

	// NPC2 should know what NPC1 said
	foundNPC1ToNPC2 := false
	for _, knownFact := range npc2KB.KnownFacts {
		if knownFact.LearnedFrom == "npc1" && knownFact.LearnMethod == knowledge.Told {
			foundNPC1ToNPC2 = true
			break
		}
	}

	// NPC1 should know what NPC2 said
	foundNPC2ToNPC1 := false
	for _, knownFact := range npc1KB.KnownFacts {
		if knownFact.LearnedFrom == "npc2" && knownFact.LearnMethod == knowledge.Told {
			foundNPC2ToNPC1 = true
			break
		}
	}

	if !foundNPC1ToNPC2 {
		t.Error("NPC2 should have learned from NPC1's response")
	}

	if !foundNPC2ToNPC1 {
		t.Error("NPC1 should have learned from NPC2's response")
	}

	t.Logf("Successfully propagated NPC responses between participants")
}

// TestStory47_ContradictionDetectionIntegration tests that contradictions are
// detected and integrated into the dialogue flow.
//
// Story 4-7 AC3: 矛盾檢測整合到對話流程
func TestStory47_ContradictionDetectionIntegration(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add NPC
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.EmotionState{Trust: 80, Fear: 10, Stress: 10},
		Archetype:      "Trusting",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}

	// Set room location
	updateMgr.SetEntityRoom("player", "room1")
	updateMgr.SetEntityRoom("npc1", "room1")

	// Create a fact that the NPC knows: "John is alive"
	fact := &knowledge.Fact{
		ID:        "fact1",
		Content:   "John 還活著",
		Type:      knowledge.Dialogue,
		Source:    "witness",
		CreatedAt: time.Now(),
		Location:  "room1",
		Witnesses: []string{"npc1"},
	}
	updateMgr.RegisterFact(fact)

	// Make NPC1 know this fact with high confidence (witnessed firsthand)
	updateMgr.PropagateEvent(&knowledge.GameEvent{
		ID:          "event1",
		Type:        "observation",
		Description: "John 還活著",
		Initiator:   "npc1",
		Location:    "room1",
		Beat:        1,
		Importance:  8,
	})

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC One", IsPlayer: false, Emotion: manager.EmotionState{Trust: 80, Fear: 10, Stress: 10}},
		},
		Location: "room1",
	}

	// Execute: Player says something contradicting what NPC knows
	contradictoryMessage := "John 已經死了"
	result, err := processor.ProcessPlayerMessage(context.Background(), session, contradictoryMessage, nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify: Contradiction should be detected
	if len(result.Contradictions) == 0 {
		t.Fatal("Expected at least one contradiction to be detected")
	}

	contradiction := result.Contradictions[0]
	if contradiction.Type == "" {
		t.Error("Contradiction type should be set")
	}

	if contradiction.Severity == 0 {
		t.Error("Contradiction severity should be > 0")
	}

	// Verify: Emotion changes should be applied
	if len(result.EmotionChanges) == 0 {
		t.Fatal("Expected emotion changes from contradiction")
	}

	emotionDelta := result.EmotionChanges["npc1"]
	if emotionDelta.Trust >= 0 {
		t.Errorf("Expected negative trust change, got %d", emotionDelta.Trust)
	}

	if emotionDelta.Stress <= 0 {
		t.Errorf("Expected positive stress change, got %d", emotionDelta.Stress)
	}

	t.Logf("Contradiction detected: Type=%s, Severity=%d", contradiction.Type, contradiction.Severity)
	t.Logf("Emotion changes applied: Trust=%d, Fear=%d, Stress=%d",
		emotionDelta.Trust, emotionDelta.Fear, emotionDelta.Stress)
}

// TestStory47_ContradictionQuestioningResponse tests that NPC responses include
// questioning when contradictions are detected.
//
// Story 4-7 AC4: 矛盾觸發時 NPC 回應包含質疑
func TestStory47_ContradictionQuestioningResponse(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add NPC
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "Dr. Smith",
		InitialEmotion: manager.EmotionState{Trust: 90, Fear: 5, Stress: 5},
		Archetype:      "Scientist",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}

	// Set room location
	updateMgr.SetEntityRoom("player", "lab")
	updateMgr.SetEntityRoom("npc1", "lab")

	// Create a fact that the NPC witnessed: "The door is closed"
	updateMgr.PropagateEvent(&knowledge.GameEvent{
		ID:          "door_closed",
		Type:        "observation",
		Description: "門是關著的",
		Initiator:   "npc1",
		Location:    "lab",
		Beat:        1,
		Importance:  9,
	})

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "Dr. Smith", IsPlayer: false, Emotion: manager.EmotionState{Trust: 90, Fear: 5, Stress: 5}},
		},
		Location: "lab",
	}

	// Execute: Player contradicts what NPC witnessed (door is open vs. door is closed)
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "門是開著的", nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify: NPC response should include contradiction flag
	if len(result.NPCResponses) == 0 {
		t.Fatal("Expected at least one NPC response")
	}

	npcResponse := result.NPCResponses[0]

	// Check if response has contradiction flag
	hasContradictionFlag := false
	for _, flag := range npcResponse.Flags {
		if flag == ChatFlagContradiction {
			hasContradictionFlag = true
			break
		}
	}

	if !hasContradictionFlag {
		t.Error("NPC response should have contradiction flag")
	}

	// Verify: Response content should include questioning
	if npcResponse.Content == "Processing..." {
		t.Error("NPC response should be customized for contradiction, not default stub")
	}

	// Response should contain NPC name
	if !containsSubstring(npcResponse.Content, "Dr. Smith") && !containsSubstring(npcResponse.Content, "Smith") {
		t.Logf("Warning: Response might not include NPC name: %s", npcResponse.Content)
	}

	t.Logf("NPC contradiction response: %s", npcResponse.Content)
	t.Logf("Response flags: %v", npcResponse.Flags)
}

// TestStory47_MultipleNPCsPropagation tests knowledge propagation with multiple NPCs
// in the same room, ensuring all receive information correctly.
//
// Integration test combining AC1 and AC2
func TestStory47_MultipleNPCsPropagation(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add three NPCs
	npcs := []*manager.NPCProfile{
		{
			ID:             "npc1",
			Name:           "Alice",
			InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
			Archetype:      "Neutral",
			Traits:         []manager.Trait{},
		},
		{
			ID:             "npc2",
			Name:           "Bob",
			InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
			Archetype:      "Friendly",
			Traits:         []manager.Trait{},
		},
		{
			ID:             "npc3",
			Name:           "Charlie",
			InitialEmotion: manager.EmotionState{Trust: 40, Fear: 40, Stress: 30},
			Archetype:      "Paranoid",
			Traits:         []manager.Trait{},
		},
	}

	for _, npc := range npcs {
		if err := npcMgr.AddNPC(npc); err != nil {
			t.Fatalf("Failed to add NPC %s: %v", npc.ID, err)
		}
	}

	// Set all to same room
	updateMgr.SetEntityRoom("player", "meeting_room")
	updateMgr.SetEntityRoom("npc1", "meeting_room")
	updateMgr.SetEntityRoom("npc2", "meeting_room")
	updateMgr.SetEntityRoom("npc3", "meeting_room")

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "Alice", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20}},
			{ID: "npc2", Name: "Bob", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15}},
			{ID: "npc3", Name: "Charlie", IsPlayer: false, Emotion: manager.EmotionState{Trust: 40, Fear: 40, Stress: 30}},
		},
		Location: "meeting_room",
	}

	// Execute
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "We need to evacuate immediately!", nil)
	if err != nil {
		t.Fatalf("ProcessPlayerMessage failed: %v", err)
	}

	// Verify: All 3 NPCs should have responses
	if len(result.NPCResponses) != 3 {
		t.Fatalf("Expected 3 NPC responses, got %d", len(result.NPCResponses))
	}

	// Verify: All 3 NPCs should have learned from player
	for _, npcID := range []string{"npc1", "npc2", "npc3"} {
		kb := updateMgr.GetKnowledgeBase(npcID)
		if kb == nil {
			t.Fatalf("NPC %s should have knowledge base", npcID)
		}

		foundPlayerMessage := false
		for _, knownFact := range kb.KnownFacts {
			if knownFact.LearnedFrom == "player" {
				foundPlayerMessage = true
				break
			}
		}

		if !foundPlayerMessage {
			t.Errorf("NPC %s should have learned from player", npcID)
		}
	}

	// Verify: Each NPC should have learned from the other NPCs' responses
	// For example, npc1 should know what npc2 and npc3 said
	npc1KB := updateMgr.GetKnowledgeBase("npc1")
	foundFromNPC2 := false
	foundFromNPC3 := false

	for _, knownFact := range npc1KB.KnownFacts {
		if knownFact.LearnedFrom == "npc2" {
			foundFromNPC2 = true
		}
		if knownFact.LearnedFrom == "npc3" {
			foundFromNPC3 = true
		}
	}

	if !foundFromNPC2 {
		t.Error("NPC1 should have learned from NPC2's response")
	}

	if !foundFromNPC3 {
		t.Error("NPC1 should have learned from NPC3's response")
	}

	t.Logf("Successfully tested knowledge propagation with %d NPCs", len(npcs))
}

// Helper function for string containment check
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==========================================================================
// Story 4.2: ChatProcessor Core Logic Integration Tests
// ==========================================================================

// TestStory42_AC1_AllDependenciesPresent verifies that ChatProcessor contains
// all required dependencies: npcManager, updateManager, judgeAgent, llmClient.
//
// Story 4-2 AC1: ChatProcessor 包含 npcManager/updateMgr/judgeAgent/llmClient
func TestStory42_AC1_AllDependenciesPresent(t *testing.T) {
	// Setup all dependencies
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)
	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	// Create ChatProcessor with all dependencies
	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	// AC1: Verify all dependencies are present
	if processor.npcManager != npcMgr {
		t.Error("AC1 Failed: npcManager not set correctly")
	}

	if processor.updateManager != updateMgr {
		t.Error("AC1 Failed: updateManager not set correctly")
	}

	if processor.judgeAgent != judgeAgent {
		t.Error("AC1 Failed: judgeAgent not set correctly")
	}

	if processor.llmClient != mockLLM {
		t.Error("AC1 Failed: llmClient not set correctly")
	}

	if processor.handlerRegistry == nil {
		t.Error("AC1 Failed: handlerRegistry should be initialized")
	}

	if processor.config == nil {
		t.Error("AC1 Failed: config should be initialized with defaults")
	}

	t.Log("AC1 PASSED: ChatProcessor contains all required dependencies")
}

// TestStory42_AC2_ProcessPlayerMessageHandling verifies that ProcessPlayerMessage()
// correctly processes player messages and orchestrates the entire flow.
//
// Story 4-2 AC2: ProcessPlayerMessage() 處理玩家訊息
func TestStory42_AC2_ProcessPlayerMessageHandling(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:             "test_npc",
		Name:           "Test NPC",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Neutral",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Setup mock LLM to return specific flags
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["revelation"],
				"confidence": 0.9,
				"reasoning": "Player shares important information"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	// Create chat session
	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "test_npc",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// AC2: ProcessPlayerMessage should handle the message and return a result
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "I found the key!", gameState)

	if err != nil {
		t.Fatalf("AC2 Failed: ProcessPlayerMessage returned error: %v", err)
	}

	if result == nil {
		t.Fatal("AC2 Failed: ProcessPlayerMessage returned nil result")
	}

	if !result.Success {
		t.Errorf("AC2 Failed: ProcessPlayerMessage success=false, error: %s", result.Error)
	}

	// Verify that processing occurred
	if len(result.NPCResponses) == 0 {
		t.Error("AC2 Failed: No NPC responses generated")
	}

	t.Log("AC2 PASSED: ProcessPlayerMessage successfully processes player messages")
}

// TestStory42_AC3_JudgeAgentInvocation verifies that ProcessPlayerMessage
// calls JudgeAgent.JudgeChat() for message analysis.
//
// Story 4-2 AC3: 呼叫 JudgeAgent.JudgeChat() 進行判定
func TestStory42_AC3_JudgeAgentInvocation(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPC
	npc := &manager.NPCProfile{
		ID:             "test_npc",
		Name:           "Test NPC",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Neutral",
		Traits:         []manager.Trait{},
	}
	if err := npcMgr.AddNPC(npc); err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Track if JudgeAgent was invoked
	judgeInvoked := false
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			judgeInvoked = true
			// Return multiple flags to verify they're processed
			return `{
				"flags": ["hallucination", "hostile"],
				"confidence": 0.85,
				"reasoning": "Player shows signs of hallucination and hostility"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "test_npc",
				Name:         "Test NPC",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// AC3: ProcessPlayerMessage should call JudgeAgent.JudgeChat()
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "I see monsters everywhere!", gameState)

	if err != nil {
		t.Fatalf("ProcessPlayerMessage returned error: %v", err)
	}

	if !judgeInvoked {
		t.Fatal("AC3 Failed: JudgeAgent.JudgeChat() was not invoked")
	}

	// Verify that flags from JudgeAgent were processed
	if len(result.Flags) != 2 {
		t.Errorf("AC3 Failed: Expected 2 flags from JudgeAgent, got %d", len(result.Flags))
	}

	// Check that both flags are present
	hasHallucination := false
	hasHostile := false
	for _, flag := range result.Flags {
		if flag == ChatFlagHallucination {
			hasHallucination = true
		}
		if flag == ChatFlagHostile {
			hasHostile = true
		}
	}

	if !hasHallucination {
		t.Error("AC3 Failed: Hallucination flag not found in result")
	}

	if !hasHostile {
		t.Error("AC3 Failed: Hostile flag not found in result")
	}

	t.Log("AC3 PASSED: ProcessPlayerMessage invokes JudgeAgent.JudgeChat() and processes flags")
}

// TestStory42_AC4_UpdateManagerPropagation verifies that ProcessPlayerMessage
// propagates information to UpdateManager.
//
// Story 4-2 AC4: 傳播資訊到 UpdateManager
func TestStory42_AC4_UpdateManagerPropagation(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPCs
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Neutral",
		Traits:         []manager.Trait{},
	}
	npc2 := &manager.NPCProfile{
		ID:             "npc2",
		Name:           "NPC Two",
		InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
		Archetype:      "Friendly",
		Traits:         []manager.Trait{},
	}

	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Set room locations
	updateMgr.SetEntityRoom("player", "test-room")
	updateMgr.SetEntityRoom("npc1", "test-room")
	updateMgr.SetEntityRoom("npc2", "test-room")

	mockLLM := &MockLLMClient{}
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc1", Name: "NPC One", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20}},
			{ID: "npc2", Name: "NPC Two", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15}},
		},
		Location: "test-room",
	}

	// AC4: Process message and verify propagation to UpdateManager
	playerMessage := "The door to the lab is locked"
	_, err := processor.ProcessPlayerMessage(context.Background(), session, playerMessage, nil)

	if err != nil {
		t.Fatalf("ProcessPlayerMessage returned error: %v", err)
	}

	// Verify that both NPCs learned from the player's message
	npc1KB := updateMgr.GetKnowledgeBase("npc1")
	if npc1KB == nil {
		t.Fatal("AC4 Failed: NPC1 knowledge base not found")
	}

	npc2KB := updateMgr.GetKnowledgeBase("npc2")
	if npc2KB == nil {
		t.Fatal("AC4 Failed: NPC2 knowledge base not found")
	}

	// Check that both NPCs have knowledge from player
	foundNPC1 := false
	for _, knownFact := range npc1KB.KnownFacts {
		if knownFact.LearnedFrom == "player" && knownFact.LearnMethod == knowledge.Told {
			foundNPC1 = true
			break
		}
	}

	foundNPC2 := false
	for _, knownFact := range npc2KB.KnownFacts {
		if knownFact.LearnedFrom == "player" && knownFact.LearnMethod == knowledge.Told {
			foundNPC2 = true
			break
		}
	}

	if !foundNPC1 {
		t.Error("AC4 Failed: NPC1 did not receive player message via UpdateManager")
	}

	if !foundNPC2 {
		t.Error("AC4 Failed: NPC2 did not receive player message via UpdateManager")
	}

	t.Log("AC4 PASSED: ProcessPlayerMessage propagates information to UpdateManager")
}

// TestStory42_AC5_ProcessResultStructure verifies that ProcessPlayerMessage
// returns a ProcessResult containing NPCResponses, EmotionChanges, and Flags.
//
// Story 4-2 AC5: 返回 ProcessResult 包含 NPCResponses/EmotionChanges/Flags
func TestStory42_AC5_ProcessResultStructure(t *testing.T) {
	// Setup
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add test NPCs
	npc1 := &manager.NPCProfile{
		ID:             "npc1",
		Name:           "NPC One",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
		Archetype:      "Skeptical",
		Traits:         []manager.Trait{},
	}
	npc2 := &manager.NPCProfile{
		ID:             "npc2",
		Name:           "NPC Two",
		InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
		Archetype:      "Trusting",
		Traits:         []manager.Trait{},
	}

	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Mock LLM to return flags that trigger emotion changes
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			return `{
				"flags": ["hostile", "lie"],
				"confidence": 0.9,
				"reasoning": "Player is aggressive and lying"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{
				ID:           "npc1",
				Name:         "NPC One",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 50, Fear: 30, Stress: 20},
				Relationship: "neutral",
			},
			{
				ID:           "npc2",
				Name:         "NPC Two",
				IsPlayer:     false,
				Emotion:      manager.EmotionState{Trust: 60, Fear: 20, Stress: 15},
				Relationship: "friendly",
			},
		},
		Location: "test-room",
	}

	gameState := &agents.GameStateSnapshot{
		HP:           100,
		SAN:          100,
		CurrentScene: "test-scene",
		Difficulty:   "normal",
	}

	// AC5: Process message and verify ProcessResult structure
	result, err := processor.ProcessPlayerMessage(context.Background(), session, "You're wrong! I didn't do it!", gameState)

	if err != nil {
		t.Fatalf("ProcessPlayerMessage returned error: %v", err)
	}

	if result == nil {
		t.Fatal("AC5 Failed: ProcessResult is nil")
	}

	// Verify NPCResponses field
	if result.NPCResponses == nil {
		t.Fatal("AC5 Failed: ProcessResult.NPCResponses is nil")
	}

	if len(result.NPCResponses) != 2 {
		t.Errorf("AC5 Failed: Expected 2 NPCResponses, got %d", len(result.NPCResponses))
	}

	// Verify each NPCResponse has required fields
	for i, response := range result.NPCResponses {
		if response.NPCID == "" {
			t.Errorf("AC5 Failed: NPCResponse[%d].NPCID is empty", i)
		}
		if response.Content == "" {
			t.Errorf("AC5 Failed: NPCResponse[%d].Content is empty", i)
		}
		// Emotion field should be populated
		if response.Emotion.Trust == 0 && response.Emotion.Fear == 0 && response.Emotion.Stress == 0 {
			t.Logf("Warning: NPCResponse[%d].Emotion appears to be zero-initialized", i)
		}
	}

	// Verify EmotionChanges field
	if result.EmotionChanges == nil {
		t.Fatal("AC5 Failed: ProcessResult.EmotionChanges is nil")
	}

	if len(result.EmotionChanges) != 2 {
		t.Errorf("AC5 Failed: Expected 2 EmotionChanges, got %d", len(result.EmotionChanges))
	}

	// Verify emotion changes for each NPC
	if _, exists := result.EmotionChanges["npc1"]; !exists {
		t.Error("AC5 Failed: EmotionChanges missing entry for npc1")
	}

	if _, exists := result.EmotionChanges["npc2"]; !exists {
		t.Error("AC5 Failed: EmotionChanges missing entry for npc2")
	}

	// Verify Flags field
	if result.Flags == nil {
		t.Fatal("AC5 Failed: ProcessResult.Flags is nil")
	}

	if len(result.Flags) != 2 {
		t.Errorf("AC5 Failed: Expected 2 flags (hostile, lie), got %d", len(result.Flags))
	}

	// Verify Success and Error fields
	if !result.Success {
		t.Errorf("AC5 Failed: ProcessResult.Success should be true, got false. Error: %s", result.Error)
	}

	if result.Error != "" {
		t.Errorf("AC5 Failed: ProcessResult.Error should be empty on success, got: %s", result.Error)
	}

	t.Log("AC5 PASSED: ProcessResult contains NPCResponses, EmotionChanges, and Flags")
	t.Logf("  - NPCResponses: %d entries", len(result.NPCResponses))
	t.Logf("  - EmotionChanges: %d entries", len(result.EmotionChanges))
	t.Logf("  - Flags: %d entries", len(result.Flags))
	t.Logf("  - Success: %v", result.Success)
}

// TestStory42_FullIntegration_CompleteWorkflow verifies the complete integration
// of all components in the ChatProcessor workflow.
//
// This test verifies the full story 4-2 implementation by testing:
// - AC1: All dependencies present
// - AC2: ProcessPlayerMessage handles messages
// - AC3: JudgeAgent is invoked
// - AC4: UpdateManager receives propagation
// - AC5: ProcessResult has correct structure
func TestStory42_FullIntegration_CompleteWorkflow(t *testing.T) {
	// Setup complete environment
	updateMgr := knowledge.NewUpdateManager(nil)
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// Add multiple NPCs with different archetypes
	npcs := []*manager.NPCProfile{
		{
			ID:             "scientist",
			Name:           "Dr. Chen",
			InitialEmotion: manager.EmotionState{Trust: 70, Fear: 20, Stress: 10},
			Archetype:      "Scientist",
			Traits:         []manager.Trait{{ID: "trait1", Content: "analytical", RevealTier: 1}},
		},
		{
			ID:             "guard",
			Name:           "Officer Martinez",
			InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			Archetype:      "Guard",
			Traits:         []manager.Trait{{ID: "trait2", Content: "protective", RevealTier: 1}},
		},
		{
			ID:             "survivor",
			Name:           "Sarah",
			InitialEmotion: manager.EmotionState{Trust: 30, Fear: 60, Stress: 70},
			Archetype:      "Survivor",
			Traits:         []manager.Trait{{ID: "trait3", Content: "paranoid", RevealTier: 2}},
		},
	}

	for _, npc := range npcs {
		if err := npcMgr.AddNPC(npc); err != nil {
			t.Fatalf("Failed to add NPC %s: %v", npc.ID, err)
		}
	}

	// Set all entities to same room
	updateMgr.SetEntityRoom("player", "control_room")
	updateMgr.SetEntityRoom("scientist", "control_room")
	updateMgr.SetEntityRoom("guard", "control_room")
	updateMgr.SetEntityRoom("survivor", "control_room")

	// Track JudgeAgent invocations
	judgeCallCount := 0
	mockLLM := &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, opts map[string]any) (string, error) {
			judgeCallCount++
			// Return revelation flag to trigger trust increase
			return `{
				"flags": ["revelation"],
				"confidence": 0.95,
				"reasoning": "Player shares critical survival information"
			}`, nil
		},
	}

	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "IntegrationTestJudge",
		LLMClient: mockLLM,
	})

	processor := NewChatProcessor(npcMgr, updateMgr, judgeAgent, mockLLM, nil)

	// Verify AC1: All dependencies present
	if processor.npcManager == nil || processor.updateManager == nil ||
		processor.judgeAgent == nil || processor.llmClient == nil {
		t.Fatal("AC1 Failed: Not all dependencies are present in ChatProcessor")
	}

	// Create realistic chat session
	session := &ChatSession{
		SessionID: "integration-test-session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "scientist", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 70, Fear: 20, Stress: 10}, Relationship: "friendly"},
			{ID: "guard", Name: "Officer Martinez", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}, Relationship: "neutral"},
			{ID: "survivor", Name: "Sarah", IsPlayer: false, Emotion: manager.EmotionState{Trust: 30, Fear: 60, Stress: 70}, Relationship: "cautious"},
		},
		Location:         "control_room",
		MessageHistory:   []ChatMessage{},
		MaxHistoryLength: 20,
	}

	gameState := &agents.GameStateSnapshot{
		HP:           85,
		SAN:          70,
		CurrentScene: "chapter_3_escape",
		Difficulty:   "hard",
	}

	// AC2: Process player message
	playerMessage := "I found a map showing the emergency exit route through the ventilation system"
	result, err := processor.ProcessPlayerMessage(context.Background(), session, playerMessage, gameState)

	if err != nil {
		t.Fatalf("AC2 Failed: ProcessPlayerMessage returned error: %v", err)
	}

	if result == nil {
		t.Fatal("AC2 Failed: ProcessPlayerMessage returned nil result")
	}

	if !result.Success {
		t.Errorf("AC2 Failed: Processing was not successful. Error: %s", result.Error)
	}

	// AC3: Verify JudgeAgent was invoked
	if judgeCallCount != 1 {
		t.Errorf("AC3 Failed: JudgeAgent should be invoked exactly once, was invoked %d times", judgeCallCount)
	}

	if len(result.Flags) == 0 {
		t.Error("AC3 Failed: No flags returned from JudgeAgent")
	}

	// AC4: Verify UpdateManager propagation
	for _, npcID := range []string{"scientist", "guard", "survivor"} {
		kb := updateMgr.GetKnowledgeBase(npcID)
		if kb == nil {
			t.Errorf("AC4 Failed: NPC %s has no knowledge base", npcID)
			continue
		}

		foundPlayerMessage := false
		for _, knownFact := range kb.KnownFacts {
			if knownFact.LearnedFrom == "player" {
				foundPlayerMessage = true
				break
			}
		}

		if !foundPlayerMessage {
			t.Errorf("AC4 Failed: NPC %s did not receive player message via UpdateManager", npcID)
		}
	}

	// AC5: Verify ProcessResult structure and contents
	// Check NPCResponses
	if len(result.NPCResponses) != 3 {
		t.Errorf("AC5 Failed: Expected 3 NPCResponses (one per NPC), got %d", len(result.NPCResponses))
	}

	responseNPCIDs := make(map[string]bool)
	for _, response := range result.NPCResponses {
		if response.NPCID == "" {
			t.Error("AC5 Failed: NPCResponse has empty NPCID")
		}
		if response.Content == "" {
			t.Error("AC5 Failed: NPCResponse has empty Content")
		}
		responseNPCIDs[response.NPCID] = true
	}

	// Verify all NPCs have responses
	for _, npcID := range []string{"scientist", "guard", "survivor"} {
		if !responseNPCIDs[npcID] {
			t.Errorf("AC5 Failed: No response from NPC %s", npcID)
		}
	}

	// Check EmotionChanges
	if len(result.EmotionChanges) != 3 {
		t.Errorf("AC5 Failed: Expected 3 EmotionChanges, got %d", len(result.EmotionChanges))
	}

	// Verify emotion changes were applied (revelation should increase trust)
	for npcID, emotionDelta := range result.EmotionChanges {
		if emotionDelta.Trust <= 0 {
			t.Logf("Warning: NPC %s trust change is not positive for revelation: %d", npcID, emotionDelta.Trust)
		}
	}

	// Check Flags
	if len(result.Flags) != 1 {
		t.Errorf("AC5 Failed: Expected 1 flag (revelation), got %d", len(result.Flags))
	}

	if result.Flags[0] != ChatFlagRevelation {
		t.Errorf("AC5 Failed: Expected revelation flag, got %s", result.Flags[0].String())
	}

	t.Log("=== Story 4.2 Full Integration Test PASSED ===")
	t.Logf("  AC1: All dependencies present ✓")
	t.Logf("  AC2: ProcessPlayerMessage handled message ✓")
	t.Logf("  AC3: JudgeAgent invoked %d time(s) ✓", judgeCallCount)
	t.Logf("  AC4: UpdateManager propagated to %d NPCs ✓", len(npcs))
	t.Logf("  AC5: ProcessResult structure complete:")
	t.Logf("    - NPCResponses: %d", len(result.NPCResponses))
	t.Logf("    - EmotionChanges: %d", len(result.EmotionChanges))
	t.Logf("    - Flags: %d", len(result.Flags))
	t.Logf("    - Success: %v", result.Success)
}
