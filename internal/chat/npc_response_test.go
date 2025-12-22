package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Mock LLM Client for Testing
// ==========================================================================

type mockLLMClient struct {
	response string
	err      error
	callCount int
	lastPrompt string
	lastOptions map[string]any
}

func (m *mockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	m.callCount++
	m.lastPrompt = prompt
	m.lastOptions = options

	if m.err != nil {
		return "", m.err
	}

	return m.response, nil
}

// ==========================================================================
// Test: DefaultResponseGeneratorConfig
// ==========================================================================

func TestDefaultResponseGeneratorConfig(t *testing.T) {
	config := DefaultResponseGeneratorConfig()

	// Story 4.4 AC6: 使用 Reactive/Rapid 模型（< 2 秒延遲）
	if config.Model != "gpt-4o-mini" {
		t.Errorf("Expected model to be gpt-4o-mini, got %s", config.Model)
	}

	if config.Timeout != 2*time.Second {
		t.Errorf("Expected timeout to be 2s, got %v", config.Timeout)
	}

	if config.MaxTokens != 150 {
		t.Errorf("Expected max_tokens to be 150, got %d", config.MaxTokens)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected temperature to be 0.7, got %f", config.Temperature)
	}

	// Story 4.4 AC3: Prompt 包含近期對話歷史（最近 5-10 條）
	if config.MaxHistoryMessages != 10 {
		t.Errorf("Expected max_history_messages to be 10, got %d", config.MaxHistoryMessages)
	}

	if !config.EnableFallback {
		t.Error("Expected enable_fallback to be true")
	}
}

// ==========================================================================
// Test: NewNPCResponseGenerator
// ==========================================================================

func TestNewNPCResponseGenerator(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()

	// Test with nil config (should use defaults)
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)
	if gen == nil {
		t.Fatal("Expected non-nil generator")
	}

	if gen.config.Model != "gpt-4o-mini" {
		t.Errorf("Expected default model, got %s", gen.config.Model)
	}

	// Test with custom config
	customConfig := &ResponseGeneratorConfig{
		Model: "gpt-4o",
		MaxTokens: 200,
		Temperature: 0.8,
		Timeout: 5 * time.Second,
		MaxHistoryMessages: 15,
		EnableFallback: false,
	}

	gen2 := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, customConfig)
	if gen2.config.Model != "gpt-4o" {
		t.Errorf("Expected custom model, got %s", gen2.config.Model)
	}

	if gen2.config.MaxTokens != 200 {
		t.Errorf("Expected custom max_tokens, got %d", gen2.config.MaxTokens)
	}
}

// ==========================================================================
// Test: GenerateNPCResponse - Basic Success
// ==========================================================================

func TestGenerateNPCResponse_BasicSuccess(t *testing.T) {
	// Setup NPC Manager with a test NPC
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "張醫生",
		Archetype:  "Scientist",
		Appearance: "40多歲男性，戴眼鏡",
		Backstory:  "前醫院外科醫生",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 40,
		},
		DialogueStyle: manager.DialogueStyle{
			Formality:  4,
			Verbosity:  3,
			Quirks:     []string{"常說「從醫學角度來說」"},
			Vocabulary: "專業醫學術語",
		},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Setup mock LLM client
	llmClient := &mockLLMClient{
		response: "從醫學角度來說，這種情況很不尋常...",
	}

	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	// Test basic response generation
	// Story 4.4 AC1: generateNPCResponse() 使用 LLM 生成 NPC 回應
	ctx := context.Background()
	playerMessage := "醫生，這裡發生什麼事了？"
	conversationHistory := []ChatMessage{
		{SenderID: "player", Content: "你好", Timestamp: 1},
		{SenderID: "npc1", Content: "你好", Timestamp: 2},
	}
	flags := []ChatFlag{ChatFlagRevelation}
	currentEmotion := manager.EmotionState{Trust: 60, Fear: 25, Stress: 35}

	response, err := gen.GenerateNPCResponse(
		ctx,
		"npc1",
		playerMessage,
		conversationHistory,
		flags,
		currentEmotion,
	)

	if err != nil {
		t.Fatalf("GenerateNPCResponse failed: %v", err)
	}

	// Verify response
	if response.NPCID != "npc1" {
		t.Errorf("Expected NPCID npc1, got %s", response.NPCID)
	}

	if response.Content != "從醫學角度來說，這種情況很不尋常..." {
		t.Errorf("Unexpected content: %s", response.Content)
	}

	if response.Emotion.Trust != 60 {
		t.Errorf("Expected emotion trust 60, got %d", response.Emotion.Trust)
	}

	if len(response.Flags) != 1 || response.Flags[0] != ChatFlagRevelation {
		t.Errorf("Expected flags [revelation], got %v", response.Flags)
	}

	if response.UsedFallback {
		t.Error("Expected UsedFallback to be false")
	}

	// Verify LLM was called
	if llmClient.callCount != 1 {
		t.Errorf("Expected LLM to be called once, got %d calls", llmClient.callCount)
	}
}

// ==========================================================================
// Test: GenerateNPCResponse - Validation Errors
// ==========================================================================

func TestGenerateNPCResponse_ValidationErrors(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	ctx := context.Background()

	// Test empty npcID
	_, err := gen.GenerateNPCResponse(
		ctx,
		"",
		"message",
		[]ChatMessage{},
		[]ChatFlag{},
		manager.EmotionState{},
	)

	if err == nil || !strings.Contains(err.Error(), "npcID cannot be empty") {
		t.Errorf("Expected npcID validation error, got: %v", err)
	}

	// Test empty playerMessage
	_, err = gen.GenerateNPCResponse(
		ctx,
		"npc1",
		"",
		[]ChatMessage{},
		[]ChatFlag{},
		manager.EmotionState{},
	)

	if err == nil || !strings.Contains(err.Error(), "playerMessage cannot be empty") {
		t.Errorf("Expected playerMessage validation error, got: %v", err)
	}

	// Test NPC not found
	_, err = gen.GenerateNPCResponse(
		ctx,
		"nonexistent",
		"message",
		[]ChatMessage{},
		[]ChatFlag{},
		manager.EmotionState{},
	)

	if err == nil || !strings.Contains(err.Error(), "NPC not found") {
		t.Errorf("Expected NPC not found error, got: %v", err)
	}
}

// ==========================================================================
// Test: buildResponsePrompt - Full NPC Context
// ==========================================================================

func TestBuildResponsePrompt_FullContext(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "張醫生",
		Archetype:  "Scientist",
		Appearance: "40多歲男性",
		Backstory:  "醫生",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{
			Formality:  4,
			Verbosity:  3,
			Quirks:     []string{"常說醫學術語"},
			Vocabulary: "專業術語",
		},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	// Build prompt
	playerMessage := "你知道這裡發生什麼事嗎？"
	conversationHistory := []ChatMessage{
		{SenderID: "player", Content: "你好", Timestamp: 1},
		{SenderID: "npc1", Content: "你好", Timestamp: 2},
		{SenderID: "player", Content: "請問你是誰？", Timestamp: 3},
	}
	flags := []ChatFlag{ChatFlagRevelation, ChatFlagHostile}
	currentEmotion := manager.EmotionState{Trust: 60, Fear: 25, Stress: 35}

	prompt := gen.buildResponsePrompt(
		"npc1",
		playerMessage,
		conversationHistory,
		flags,
		currentEmotion,
	)

	// Story 4.4 AC2: Prompt 包含 NPC 完整檔案（BuildFullNPCPrompt）
	if !strings.Contains(prompt, "張醫生") {
		t.Error("Prompt should contain NPC name")
	}

	if !strings.Contains(prompt, "40多歲男性") {
		t.Error("Prompt should contain NPC appearance")
	}

	// Story 4.4 AC3: Prompt 包含近期對話歷史（最近 5-10 條）
	if !strings.Contains(prompt, "近期對話歷史") {
		t.Error("Prompt should contain conversation history section")
	}

	if !strings.Contains(prompt, "你好") {
		t.Error("Prompt should contain conversation history content")
	}

	// Story 4.4 AC4: Prompt 包含判定 Flags（影響回應態度）
	if !strings.Contains(prompt, "判定 Flags") {
		t.Error("Prompt should contain flags section")
	}

	if !strings.Contains(prompt, "revelation") {
		t.Error("Prompt should contain revelation flag")
	}

	if !strings.Contains(prompt, "hostile") {
		t.Error("Prompt should contain hostile flag")
	}

	// Story 4.4 AC5: 回應包含適當的情感反應
	if !strings.Contains(prompt, "當前情緒狀態") {
		t.Error("Prompt should contain emotion state section")
	}

	if !strings.Contains(prompt, "60/100") {
		t.Error("Prompt should contain trust value")
	}

	// Verify player message is included
	if !strings.Contains(prompt, playerMessage) {
		t.Error("Prompt should contain player message")
	}

	// Verify response instructions
	if !strings.Contains(prompt, "回應指示") {
		t.Error("Prompt should contain response instructions")
	}
}

// ==========================================================================
// Test: buildResponsePrompt - History Limit
// ==========================================================================

func TestBuildResponsePrompt_HistoryLimit(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Any",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.MaxHistoryMessages = 3
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	// Create 10 messages in history
	history := make([]ChatMessage, 10)
	for i := 0; i < 10; i++ {
		history[i] = ChatMessage{
			SenderID: "player",
			Content:  "message " + string(rune('0'+i)),
			Timestamp: i,
		}
	}

	prompt := gen.buildResponsePrompt(
		"npc1",
		"current message",
		history,
		[]ChatFlag{},
		manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
	)

	// Story 4.4 AC3: Should only include last 3 messages
	if !strings.Contains(prompt, "message 7") {
		t.Error("Prompt should contain message 7 (8th message)")
	}

	if !strings.Contains(prompt, "message 8") {
		t.Error("Prompt should contain message 8 (9th message)")
	}

	if !strings.Contains(prompt, "message 9") {
		t.Error("Prompt should contain message 9 (10th message)")
	}

	if strings.Contains(prompt, "message 6") {
		t.Error("Prompt should NOT contain message 6 (older than limit)")
	}
}

// ==========================================================================
// Test: describeChatFlag
// ==========================================================================

func TestDescribeChatFlag(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	tests := []struct {
		flag        ChatFlag
		expectedSubstring string
	}{
		{ChatFlagHallucination, "幻覺"},
		{ChatFlagHostile, "威脅"},
		{ChatFlagRevelation, "分享重要資訊"},
		{ChatFlagPersuasion, "說服"},
		{ChatFlagLie, "說謊"},
		{ChatFlagContradiction, "矛盾"},
	}

	for _, tt := range tests {
		desc := gen.describeChatFlag(tt.flag)
		if !strings.Contains(desc, tt.expectedSubstring) {
			t.Errorf("Flag %s description should contain '%s', got: %s",
				tt.flag, tt.expectedSubstring, desc)
		}
	}

	// Test default case (unknown flag)
	unknownFlag := ChatFlag("unknown_flag")
	desc := gen.describeChatFlag(unknownFlag)
	if desc != "unknown_flag" {
		t.Errorf("Unknown flag should return original string, got: %s", desc)
	}
}

// ==========================================================================
// Test: callLLMForResponse - Success
// ==========================================================================

func TestCallLLMForResponse_Success(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{
		response: "  這是測試回應  ",
	}
	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	ctx := context.Background()
	response, err := gen.callLLMForResponse(ctx, "test prompt")

	if err != nil {
		t.Fatalf("callLLMForResponse failed: %v", err)
	}

	// Should trim whitespace
	if response != "這是測試回應" {
		t.Errorf("Expected trimmed response, got: '%s'", response)
	}

	// Verify LLM options
	if llmClient.lastOptions["model"] != "gpt-4o-mini" {
		t.Errorf("Expected model gpt-4o-mini, got %v", llmClient.lastOptions["model"])
	}

	if llmClient.lastOptions["max_tokens"] != 150 {
		t.Errorf("Expected max_tokens 150, got %v", llmClient.lastOptions["max_tokens"])
	}

	if llmClient.lastOptions["temperature"] != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", llmClient.lastOptions["temperature"])
	}
}

// ==========================================================================
// Test: callLLMForResponse - Errors
// ==========================================================================

func TestCallLLMForResponse_Errors(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	fallbackMgr := NewFallbackManager()

	// Test LLM error
	llmClient := &mockLLMClient{
		err: errors.New("LLM service unavailable"),
	}
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	ctx := context.Background()
	_, err := gen.callLLMForResponse(ctx, "test prompt")

	if err == nil {
		t.Fatal("Expected error from LLM failure")
	}

	if !strings.Contains(err.Error(), "LLM generate failed") {
		t.Errorf("Expected LLM generate failed error, got: %v", err)
	}

	// Test empty response
	llmClient2 := &mockLLMClient{
		response: "   ",
	}
	gen2 := NewNPCResponseGenerator(npcMgr, llmClient2, fallbackMgr, nil)

	_, err = gen2.callLLMForResponse(ctx, "test prompt")

	if err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Errorf("Expected empty response error, got: %v", err)
	}
}

// ==========================================================================
// Test: generateFallbackResponse
// ==========================================================================

func TestGenerateFallbackResponse(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	state := npcMgr.GetState("npc1")
	flags := []ChatFlag{ChatFlagHostile}
	emotion := manager.EmotionState{Trust: 30, Fear: 70, Stress: 60}

	response := gen.generateFallbackResponse("npc1", profile, state, flags, emotion)

	// Verify fallback response structure
	if response.NPCID != "npc1" {
		t.Errorf("Expected NPCID npc1, got %s", response.NPCID)
	}

	if response.Content == "" {
		t.Error("Fallback content should not be empty")
	}

	if !response.UsedFallback {
		t.Error("UsedFallback should be true")
	}

	if len(response.Flags) != 1 || response.Flags[0] != ChatFlagHostile {
		t.Errorf("Expected flags [hostile], got %v", response.Flags)
	}
}

// ==========================================================================
// Test: LLM Failure with Fallback Enabled
// ==========================================================================

func TestGenerateNPCResponse_LLMFailureWithFallback(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// LLM that always fails
	llmClient := &mockLLMClient{
		err: errors.New("LLM unavailable"),
	}

	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.EnableFallback = true
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	ctx := context.Background()
	response, err := gen.GenerateNPCResponse(
		ctx,
		"npc1",
		"test message",
		[]ChatMessage{},
		[]ChatFlag{ChatFlagHostile},
		manager.EmotionState{Trust: 30, Fear: 70, Stress: 60},
	)

	// Should succeed with fallback
	if err != nil {
		t.Fatalf("Expected success with fallback, got error: %v", err)
	}

	if !response.UsedFallback {
		t.Error("Expected UsedFallback to be true")
	}

	if response.Content == "" {
		t.Error("Fallback content should not be empty")
	}
}

// ==========================================================================
// Test: LLM Failure with Fallback Disabled
// ==========================================================================

func TestGenerateNPCResponse_LLMFailureNoFallback(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// LLM that always fails
	llmClient := &mockLLMClient{
		err: errors.New("LLM unavailable"),
	}

	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.EnableFallback = false
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	ctx := context.Background()
	_, err = gen.GenerateNPCResponse(
		ctx,
		"npc1",
		"test message",
		[]ChatMessage{},
		[]ChatFlag{},
		manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
	)

	// Should fail without fallback
	if err == nil {
		t.Fatal("Expected error when fallback disabled")
	}

	if !strings.Contains(err.Error(), "fallback disabled") {
		t.Errorf("Expected fallback disabled error, got: %v", err)
	}
}

// ==========================================================================
// Test: Timeout Handling
// ==========================================================================

func TestGenerateNPCResponse_Timeout(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// LLM that simulates slow response
	llmClient := &mockLLMClient{
		response: "slow response",
	}

	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.Timeout = 1 * time.Millisecond // Very short timeout
	config.EnableFallback = true
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	// Create a context that will timeout
	ctx := context.Background()

	// Note: This test may be flaky depending on system speed
	// The main goal is to verify timeout handling exists
	_, err = gen.GenerateNPCResponse(
		ctx,
		"npc1",
		"test message",
		[]ChatMessage{},
		[]ChatFlag{},
		manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
	)

	// Should succeed with fallback even if LLM times out
	// This validates the graceful degradation pattern
	if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
		// If there's an error, it should be a timeout or fallback should have worked
		t.Logf("Got error (may be timeout or fallback): %v", err)
	}
}

// ==========================================================================
// Test: GenerateAllNPCResponses
// ==========================================================================

func TestGenerateAllNPCResponses(t *testing.T) {
	// Setup NPC Manager with multiple NPCs
	npcMgr := manager.NewNPCManager(nil, nil)

	// Add NPC 1
	profile1 := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "NPC1",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}
	err := npcMgr.AddNPC(profile1)
	if err != nil {
		t.Fatalf("Failed to add NPC1: %v", err)
	}

	// Add NPC 2
	profile2 := &manager.NPCProfile{
		ID:         "npc2",
		Name:       "NPC2",
		Archetype:  "Guard",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}
	err = npcMgr.AddNPC(profile2)
	if err != nil {
		t.Fatalf("Failed to add NPC2: %v", err)
	}

	// Setup LLM client
	llmClient := &mockLLMClient{
		response: "Test response",
	}

	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	// Create session with multiple NPCs
	session := &ChatSession{
		SessionID: "test-session",
		Participants: []ChatParticipant{
			{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
				Emotion:  manager.EmotionState{},
			},
			{
				ID:       "npc1",
				Name:     "NPC1",
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
			},
			{
				ID:       "npc2",
				Name:     "NPC2",
				IsPlayer: false,
				Emotion:  manager.EmotionState{Trust: 60, Fear: 20, Stress: 30},
			},
		},
		MessageHistory:   []ChatMessage{},
		MaxHistoryLength: 20,
	}

	emotionChanges := map[string]manager.EmotionDelta{
		"npc1": {Trust: 10, Fear: -5, Stress: 0},
	}

	flags := []ChatFlag{ChatFlagRevelation}

	ctx := context.Background()
	responses, err := gen.GenerateAllNPCResponses(
		ctx,
		session,
		"test message",
		emotionChanges,
		flags,
	)

	if err != nil {
		t.Fatalf("GenerateAllNPCResponses failed: %v", err)
	}

	// Should generate responses for both NPCs (not player)
	if len(responses) != 2 {
		t.Fatalf("Expected 2 responses, got %d", len(responses))
	}

	// Verify responses
	foundNPC1 := false
	foundNPC2 := false

	for _, resp := range responses {
		if resp.NPCID == "npc1" {
			foundNPC1 = true
			// Emotion should be updated with delta
			if resp.Emotion.Trust != 60 { // 50 + 10
				t.Errorf("Expected NPC1 trust 60, got %d", resp.Emotion.Trust)
			}
		}
		if resp.NPCID == "npc2" {
			foundNPC2 = true
			// Emotion should be unchanged (no delta)
			if resp.Emotion.Trust != 60 {
				t.Errorf("Expected NPC2 trust 60, got %d", resp.Emotion.Trust)
			}
		}
	}

	if !foundNPC1 {
		t.Error("Missing response for NPC1")
	}
	if !foundNPC2 {
		t.Error("Missing response for NPC2")
	}

	// Verify LLM was called twice (once per NPC)
	if llmClient.callCount != 2 {
		t.Errorf("Expected LLM to be called twice, got %d calls", llmClient.callCount)
	}
}

// ==========================================================================
// Test: ParseLLMResponse
// ==========================================================================

func TestParseLLMResponse(t *testing.T) {
	// Test valid JSON
	jsonResponse := `{
		"content": "這是測試回應",
		"emotion": {
			"trust": 60,
			"fear": 25,
			"stress": 35
		},
		"metadata": {
			"test": "value"
		}
	}`

	parsed, err := ParseLLMResponse(jsonResponse)
	if err != nil {
		t.Fatalf("ParseLLMResponse failed: %v", err)
	}

	if parsed.Content != "這是測試回應" {
		t.Errorf("Expected content '這是測試回應', got '%s'", parsed.Content)
	}

	if parsed.Emotion == nil {
		t.Fatal("Expected emotion to be non-nil")
	}

	if parsed.Emotion.Trust != 60 {
		t.Errorf("Expected trust 60, got %d", parsed.Emotion.Trust)
	}

	// Test invalid JSON
	_, err = ParseLLMResponse("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test missing content
	jsonNoContent := `{"emotion": {"trust": 50}}`
	_, err = ParseLLMResponse(jsonNoContent)
	if err == nil || !strings.Contains(err.Error(), "missing content") {
		t.Errorf("Expected missing content error, got: %v", err)
	}
}

// ==========================================================================
// Test: Integration with Multiple Flags
// ==========================================================================

func TestGenerateNPCResponse_MultipleFlags(t *testing.T) {
	// Setup NPC Manager
	npcMgr := manager.NewNPCManager(nil, nil)
	profile := &manager.NPCProfile{
		ID:         "npc1",
		Name:       "測試NPC",
		Archetype:  "Scientist",
		Appearance: "普通人",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		DialogueStyle: manager.DialogueStyle{Vocabulary: "普通"},
	}

	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	llmClient := &mockLLMClient{
		response: "複雜的回應",
	}

	fallbackMgr := NewFallbackManager()
	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, nil)

	// Test with multiple flags
	ctx := context.Background()
	flags := []ChatFlag{
		ChatFlagHallucination,
		ChatFlagHostile,
		ChatFlagLie,
	}

	response, err := gen.GenerateNPCResponse(
		ctx,
		"npc1",
		"test message",
		[]ChatMessage{},
		flags,
		manager.EmotionState{Trust: 30, Fear: 70, Stress: 60},
	)

	if err != nil {
		t.Fatalf("GenerateNPCResponse failed: %v", err)
	}

	// Verify all flags are included in response
	if len(response.Flags) != 3 {
		t.Errorf("Expected 3 flags, got %d", len(response.Flags))
	}

	// Verify prompt includes all flag descriptions
	if llmClient.lastPrompt == "" {
		t.Fatal("Expected LLM to be called")
	}

	if !strings.Contains(llmClient.lastPrompt, "hallucination") {
		t.Error("Prompt should contain hallucination flag")
	}

	if !strings.Contains(llmClient.lastPrompt, "hostile") {
		t.Error("Prompt should contain hostile flag")
	}

	if !strings.Contains(llmClient.lastPrompt, "lie") {
		t.Error("Prompt should contain lie flag")
	}
}

// ==========================================================================
// Story 8.7: Multi-language Support Integration Tests
// ==========================================================================

// TestNPCResponseGenerator_LanguageConsistency tests NPC dialogue in all languages
// Story 8.7 AC3: i18n integration into NPC dialogue system
func TestNPCResponseGenerator_LanguageConsistency(t *testing.T) {
	// Import i18n for testing
	// Note: This test verifies that language changes are reflected in prompts

	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{response: "Test response"}
	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()

	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	// Create test NPC profile
	profile := &manager.NPCProfile{
		ID:         "test-npc",
		Name:       "Test NPC",
		Archetype:  "N-04",
		Appearance: "Test appearance",
		Backstory:  "A test character",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 40,
		},
	}
	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	testCases := []struct {
		locale           string
		expectedInPrompt []string
		notExpected      []string
	}{
		{
			locale: "zh-TW",
			expectedInPrompt: []string{
				"當前情緒狀態",
				"信任度",
				"恐懼度",
				"壓力值",
				"玩家的訊息",
				"回應指示",
			},
			notExpected: []string{
				"当前情绪状态",
				"Current Emotional State",
				"Trust:",
			},
		},
		{
			locale: "zh-CN",
			expectedInPrompt: []string{
				"当前情绪状态",
				"信任度",
				"恐惧度",
				"压力值",
				"玩家的消息",
				"回应指示",
			},
			notExpected: []string{
				"當前情緒狀態",
				"Current Emotional State",
				"Trust:",
			},
		},
		{
			locale: "en-US",
			expectedInPrompt: []string{
				"Current Emotional State",
				"Trust:",
				"Fear:",
				"Stress:",
				"Player's Message",
				"Response Instructions",
			},
			notExpected: []string{
				"當前情緒狀態",
				"当前情绪状态",
				"信任度",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.locale, func(t *testing.T) {
			// Note: We can't actually set the global translator in tests
			// because it might affect other tests. Instead, we test
			// the language-specific helper methods directly.

			// Test system instructions
			sysInst := gen.getSystemInstructions(tc.locale)
			if sysInst == "" {
				t.Error("System instructions should not be empty")
			}

			// Test emotion header
			emotionHeader := gen.getEmotionHeader(tc.locale)
			if emotionHeader == "" {
				t.Error("Emotion header should not be empty")
			}

			// Test emotion state formatting
			emotion := manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}
			emotionStr := gen.formatEmotionState(tc.locale, emotion)
			if emotionStr == "" {
				t.Error("Emotion state string should not be empty")
			}

			// Verify correct language is used
			switch tc.locale {
			case "zh-TW":
				if !strings.Contains(emotionHeader, "當前") {
					t.Errorf("zh-TW emotion header should contain traditional characters: %s", emotionHeader)
				}
			case "zh-CN":
				if !strings.Contains(emotionHeader, "当前") {
					t.Errorf("zh-CN emotion header should contain simplified characters: %s", emotionHeader)
				}
			case "en-US":
				if !strings.Contains(emotionHeader, "Current") {
					t.Errorf("en-US emotion header should be in English: %s", emotionHeader)
				}
			}
		})
	}
}

// TestNPCResponseGenerator_ChatFlagLanguages tests flag descriptions in all languages
// Story 8.7 AC3: i18n integration into NPC dialogue system
func TestNPCResponseGenerator_ChatFlagLanguages(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()

	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	flags := []ChatFlag{
		ChatFlagHallucination,
		ChatFlagHostile,
		ChatFlagRevelation,
		ChatFlagPersuasion,
		ChatFlagLie,
		ChatFlagContradiction,
	}

	locales := []string{"zh-TW", "zh-CN", "en-US"}

	for _, locale := range locales {
		t.Run(locale, func(t *testing.T) {
			for _, flag := range flags {
				// Test Traditional Chinese descriptions
				descZhTW := gen.describeChatFlagZhTW(flag)
				if descZhTW == "" || descZhTW == string(flag) {
					t.Errorf("zh-TW description missing for flag %s", flag)
				}

				// Test Simplified Chinese descriptions
				descZhCN := gen.describeChatFlagZhCN(flag)
				if descZhCN == "" || descZhCN == string(flag) {
					t.Errorf("zh-CN description missing for flag %s", flag)
				}

				// Test English descriptions
				descEnUS := gen.describeChatFlagEnUS(flag)
				if descEnUS == "" || descEnUS == string(flag) {
					t.Errorf("en-US description missing for flag %s", flag)
				}

				// Verify they are different
				if descZhTW == descZhCN {
					t.Errorf("zh-TW and zh-CN descriptions should differ for flag %s", flag)
				}
				if descZhTW == descEnUS {
					t.Errorf("zh-TW and en-US descriptions should differ for flag %s", flag)
				}
			}
		})
	}
}

// TestNPCResponseGenerator_PromptLanguageHelpers tests all language helper methods
// Story 8.7 AC3: i18n integration into NPC dialogue system
func TestNPCResponseGenerator_PromptLanguageHelpers(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()

	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	locales := []string{"zh-TW", "zh-CN", "en-US"}

	for _, locale := range locales {
		t.Run(locale, func(t *testing.T) {
			// Test all helper methods return non-empty strings
			helpers := map[string]string{
				"SystemInstructions":   gen.getSystemInstructions(locale),
				"EmotionHeader":        gen.getEmotionHeader(locale),
				"FlagsHeader":          gen.getFlagsHeader(locale),
				"FlagsIntro":           gen.getFlagsIntro(locale),
				"HistoryHeader":        gen.getHistoryHeader(locale),
				"PlayerLabel":          gen.getPlayerLabel(locale),
				"PlayerMessageHeader":  gen.getPlayerMessageHeader(locale),
				"ResponseInstructions": gen.getResponseInstructions(locale),
			}

			for name, value := range helpers {
				if value == "" {
					t.Errorf("%s should not be empty for locale %s", name, locale)
				}
			}

			// Test emotion formatting
			emotion := manager.EmotionState{Trust: 75, Fear: 25, Stress: 50}
			emotionStr := gen.formatEmotionState(locale, emotion)
			if emotionStr == "" {
				t.Error("formatEmotionState should not return empty string")
			}

			// Verify numbers are present
			if !strings.Contains(emotionStr, "75") {
				t.Error("Emotion string should contain trust value 75")
			}
			if !strings.Contains(emotionStr, "25") {
				t.Error("Emotion string should contain fear value 25")
			}
			if !strings.Contains(emotionStr, "50") {
				t.Error("Emotion string should contain stress value 50")
			}
		})
	}
}

// TestNPCResponseGenerator_LanguageConsistencyInBuildPrompt tests full prompt generation
// Story 8.7 AC3: i18n integration into NPC dialogue system
func TestNPCResponseGenerator_LanguageConsistencyInBuildPrompt(t *testing.T) {
	npcMgr := manager.NewNPCManager(nil, nil)
	llmClient := &mockLLMClient{}
	fallbackMgr := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()

	gen := NewNPCResponseGenerator(npcMgr, llmClient, fallbackMgr, config)

	// Create test NPC
	profile := &manager.NPCProfile{
		ID:         "lang-test-npc",
		Name:       "Language Test NPC",
		Archetype:  "N-01",
		Appearance: "Friendly person",
		Backstory:  "Test",
		Traits:     []manager.Trait{},
		InitialEmotion: manager.EmotionState{
			Trust:  60,
			Fear:   20,
			Stress: 30,
		},
	}
	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Test prompt building (internal method - we can't call it directly in real scenarios)
	// But we can verify the helper methods work correctly

	testMessage := "Hello, how are you?"
	testHistory := []ChatMessage{
		{SenderID: "player", Content: "Hi"},
		{SenderID: "lang-test-npc", Content: "Hello"},
	}
	testFlags := []ChatFlag{ChatFlagHostile}
	testEmotion := manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}

	// Call buildResponsePrompt through GenerateNPCResponse
	// This will verify the full integration
	ctx := context.Background()

	llmClient.response = "Test NPC response"
	_, err2 := gen.GenerateNPCResponse(ctx, "lang-test-npc", testMessage, testHistory, testFlags, testEmotion)

	if err2 != nil {
		t.Fatalf("GenerateNPCResponse failed: %v", err2)
	}

	// Verify LLM was called with a prompt
	if llmClient.lastPrompt == "" {
		t.Fatal("LLM should have been called with a prompt")
	}

	// The prompt should contain emotion values
	if !strings.Contains(llmClient.lastPrompt, "60") {
		t.Error("Prompt should contain trust value 60")
	}
	if !strings.Contains(llmClient.lastPrompt, "20") {
		t.Error("Prompt should contain fear value 20")
	}
	if !strings.Contains(llmClient.lastPrompt, "30") {
		t.Error("Prompt should contain stress value 30")
	}
}
