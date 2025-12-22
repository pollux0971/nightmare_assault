package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/chat"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ChatTestEnv contains all components needed for chat integration testing.
// This provides a complete environment for testing Phase 2 chat system.
type ChatTestEnv struct {
	// Core components
	Orchestrator  *Orchestrator
	NPCManager    *manager.NPCManager
	UpdateManager *knowledge.UpdateManager
	ChatProcessor *chat.ChatProcessor
	MockLLM       *MockLLMClient

	// Test data
	TestNPCs []*manager.NPCProfile
}

// setupChatIntegrationTest creates a complete test environment for chat integration tests.
// This is the main setup function used by all AC tests.
//
// Components initialized:
//   - NPCManager with test NPCs
//   - UpdateManager with room/entity mappings
//   - Mock LLM client with predefined responses
//   - ChatProcessor with all dependencies
//   - Orchestrator with mock components
func setupChatIntegrationTest(t *testing.T) *ChatTestEnv {
	t.Helper()

	// 1. Create UpdateManager first (needed by NPCManager)
	updateMgr := knowledge.NewUpdateManager(nil)

	// 2. Create NPCManager
	npcMgr := manager.NewNPCManager(updateMgr, nil)

	// 3. Create test NPCs (2 NPCs for most tests)
	npc1 := createTestNPCProfile("npc_001", "張醫生")
	npc2 := createTestNPCProfile("npc_002", "李護士")

	// Add NPCs to manager
	if err := npcMgr.AddNPC(npc1); err != nil {
		t.Fatalf("Failed to add NPC 1: %v", err)
	}
	if err := npcMgr.AddNPC(npc2); err != nil {
		t.Fatalf("Failed to add NPC 2: %v", err)
	}

	// 4. Set initial room configuration
	updateMgr.SetEntityRoom("player", "room_001")
	updateMgr.SetEntityRoom("npc_001", "room_001")
	updateMgr.SetEntityRoom("npc_002", "room_001")

	// 5. Create Mock LLM client
	mockLLM := NewMockLLMClient()

	// 6. Create JudgeAgent with mock LLM
	judgeAgent := agents.NewJudgeAgent(agents.AgentConfig{
		Name:      "TestJudgeAgent",
		LLMClient: mockLLM,
	})

	// 7. Create ChatProcessor
	chatProcessor := chat.NewChatProcessor(
		npcMgr,
		updateMgr,
		judgeAgent,
		mockLLM,
		nil, // Use default config
	)

	// 8. Create Orchestrator with mock components
	orch := NewOrchestrator()

	return &ChatTestEnv{
		Orchestrator:  orch,
		NPCManager:    npcMgr,
		UpdateManager: updateMgr,
		ChatProcessor: chatProcessor,
		MockLLM:       mockLLM,
		TestNPCs:      []*manager.NPCProfile{npc1, npc2},
	}
}

// createTestNPCProfile creates a test NPC profile with the given ID and name.
func createTestNPCProfile(id, name string) *manager.NPCProfile {
	return &manager.NPCProfile{
		ID:         id,
		Name:       name,
		Archetype:  "醫療人員",
		Appearance: "穿著白色工作服的醫療人員",
		Backstory:  "在醫院工作多年的專業人員",
		Skills:     []string{"醫療知識", "急救"},
		Inventory:  []string{"醫療包", "筆記本"},
		Secret:     "知道醫院的一些不為人知的秘密",
		SecretTier: 2,
		Traits: []manager.Trait{
			{ID: "trait_helpful", Content: "helpful", RevealTier: 1},
			{ID: "trait_cautious", Content: "cautious", RevealTier: 2},
		},
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 20,
		},
		DialogueStyle: manager.DialogueStyle{
			Formality: 3,
			Verbosity: 3,
			Quirks:    []string{"常說'嗯'", "說話謹慎"},
		},
	}
}

// MockLLMClient is a mock LLM client for testing.
// It provides predefined responses for different request types.
type MockLLMClient struct {
	// Predefined responses for different request types
	Responses map[string]string

	// Track calls for verification
	CallCount    int
	LastRequest  string
	LastPrompt   string
	ShouldFail   bool
	FailureError error
}

// NewMockLLMClient creates a new mock LLM client with default responses.
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		Responses: map[string]string{
			"judge":      `{"flags": [], "confidence": 0.8, "reasoning": "No issues detected"}`,
			"npc_reply":  "我理解你的意思。",
			"summary":    `{"main_topics": ["test"], "narrative_impact": "test summary"}`,
			"default":    "Mock response",
		},
		CallCount:    0,
		ShouldFail:   false,
		FailureError: nil,
	}
}

// Generate implements the agents.LLMClient interface.
// It returns predefined responses based on the prompt content.
func (m *MockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	m.CallCount++
	m.LastPrompt = prompt

	// Check if should fail
	if m.ShouldFail {
		if m.FailureError != nil {
			return "", m.FailureError
		}
		return "", errors.New("mock LLM failed")
	}

	// Determine response type based on prompt content
	responseType := "default"

	// Check for judge-related prompts
	if strings.Contains(prompt, "narrative judge") || strings.Contains(prompt, "analyzing player dialogue") {
		responseType = "judge"
	} else if strings.Contains(prompt, "summarize") || strings.Contains(prompt, "conversation summary") {
		responseType = "summary"
	} else if strings.Contains(prompt, "NPC response") || strings.Contains(prompt, "reply as") {
		responseType = "npc_reply"
	}

	// Return predefined response
	if response, ok := m.Responses[responseType]; ok {
		return response, nil
	}

	return m.Responses["default"], nil
}

// Reset resets the mock client state for the next test.
func (m *MockLLMClient) Reset() {
	m.CallCount = 0
	m.LastRequest = ""
	m.LastPrompt = ""
	m.ShouldFail = false
	m.FailureError = nil
}

// SetResponse sets a custom response for a specific request type.
func (m *MockLLMClient) SetResponse(requestType, response string) {
	m.Responses[requestType] = response
}
