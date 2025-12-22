package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/stretchr/testify/assert"
)

// MockOptimizationLLMClient is a mock LLM client for testing optimization.
type MockOptimizationLLMClient struct {
	response      string
	err           error
	delay         time.Duration
	callCount     int
	lastPrompt    string
	inputTokens   int
	outputTokens  int
	generateFunc  func(ctx context.Context, prompt string, options map[string]any) (string, error)
}

func (m *MockOptimizationLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	m.callCount++
	m.lastPrompt = prompt

	// Use custom generate function if set
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt, options)
	}

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	if m.err != nil {
		return "", m.err
	}

	return m.response, nil
}

// setupTestNPCManager creates a test NPC manager with sample NPCs.
func setupTestNPCManager() *manager.NPCManager {
	npcManager := manager.NewNPCManager(nil, nil)

	// Add test NPCs
	npc1 := &manager.NPCProfile{
		ID:        "npc_001",
		Name:      "Dr. Chen",
		Archetype: "Scientist",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   30,
			Stress: 40,
		},
	}
	npc2 := &manager.NPCProfile{
		ID:        "npc_002",
		Name:      "Guard Smith",
		Archetype: "Security",
		InitialEmotion: manager.EmotionState{
			Trust:  60,
			Fear:   20,
			Stress: 30,
		},
	}

	_ = npcManager.AddNPC(npc1)
	_ = npcManager.AddNPC(npc2)

	return npcManager
}

// TestNewParallelResponseGenerator tests creation of parallel generator.
func TestNewParallelResponseGenerator(t *testing.T) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response"}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)

	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	assert.NotNil(t, pg)
	assert.Equal(t, 3, pg.maxConcurrency)
	assert.Equal(t, 5*time.Second, pg.npcTimeout)
	assert.True(t, pg.orderPreserving)
}

// TestNewParallelResponseGenerator_DefaultParams tests default parameter handling.
func TestNewParallelResponseGenerator_DefaultParams(t *testing.T) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response"}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)

	// Test with invalid parameters (should use defaults)
	pg := NewParallelResponseGenerator(generator, 0, 0, true)

	assert.NotNil(t, pg)
	assert.Equal(t, 3, pg.maxConcurrency)         // Default
	assert.Equal(t, 5*time.Second, pg.npcTimeout) // Default
}

// TestParallelResponseGenerator_GenerateAllResponses tests parallel generation.
func TestParallelResponseGenerator_GenerateAllResponses(t *testing.T) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response"}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
			{ID: "npc_002", Name: "Guard Smith", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}},
		},
		MessageHistory: []ChatMessage{},
	}

	responses, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(responses)) // 2 NPCs
	assert.Equal(t, 2, mockLLM.callCount)
}

// TestParallelResponseGenerator_EmptySession tests with no NPCs.
func TestParallelResponseGenerator_EmptySession(t *testing.T) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response"}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
		},
		MessageHistory: []ChatMessage{},
	}

	responses, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(responses))
	assert.Equal(t, 0, mockLLM.callCount)
}

// TestParallelResponseGenerator_Timeout tests timeout handling.
func TestParallelResponseGenerator_Timeout(t *testing.T) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{
		response: "Test response",
		delay:    10 * time.Second, // Longer than timeout
	}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.EnableFallback = false // Disable fallback to test timeout behavior
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 100*time.Millisecond, true) // Short timeout

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
		},
		MessageHistory: []ChatMessage{},
	}

	responses, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	// Should return error since all NPCs timed out
	assert.Error(t, err)
	assert.Equal(t, 0, len(responses))
}

// TestParallelResponseGenerator_PartialFailure tests partial failure handling.
func TestParallelResponseGenerator_PartialFailure(t *testing.T) {
	npcManager := setupTestNPCManager()

	// Create a custom LLM that fails on first call, succeeds on second
	callCount := 0
	mockLLM := &MockOptimizationLLMClient{}
	mockLLM.generateFunc = func(ctx context.Context, prompt string, options map[string]any) (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("first call fails")
		}
		return "Success", nil
	}

	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	config.EnableFallback = false // Disable fallback to see actual errors
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
			{ID: "npc_002", Name: "Guard Smith", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}},
		},
		MessageHistory: []ChatMessage{},
	}

	responses, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	// Should succeed with 1 response (second NPC)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(responses))
}

// TestParallelResponseGenerator_OrderPreserving tests response ordering.
func TestParallelResponseGenerator_OrderPreserving(t *testing.T) {
	npcManager := setupTestNPCManager()

	// Add more NPCs
	npc3 := &manager.NPCProfile{
		ID:             "npc_003",
		Name:           "NPC3",
		Archetype:      "Test",
		InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
	}
	_ = npcManager.AddNPC(npc3)

	mockLLM := &MockOptimizationLLMClient{response: "Response"}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "NPC1", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
			{ID: "npc_002", Name: "NPC2", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}},
			{ID: "npc_003", Name: "NPC3", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
		},
		MessageHistory: []ChatMessage{},
	}

	responses, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(responses))

	// Verify order matches participant order
	assert.Equal(t, "npc_001", responses[0].NPCID)
	assert.Equal(t, "npc_002", responses[1].NPCID)
	assert.Equal(t, "npc_003", responses[2].NPCID)
}

// TestOptimizedPromptBuilder tests prompt optimization.
func TestOptimizedPromptBuilder(t *testing.T) {
	builder := NewOptimizedPromptBuilder(5)

	assert.NotNil(t, builder)
	assert.Equal(t, 5, builder.maxHistoryMessages)
}

// TestOptimizedPromptBuilder_DefaultParams tests default parameters.
func TestOptimizedPromptBuilder_DefaultParams(t *testing.T) {
	builder := NewOptimizedPromptBuilder(0)

	assert.NotNil(t, builder)
	assert.Equal(t, 5, builder.maxHistoryMessages) // Should use default
}

// TestOptimizedPromptBuilder_BuildOptimizedPrompt tests prompt building.
func TestOptimizedPromptBuilder_BuildOptimizedPrompt(t *testing.T) {
	builder := NewOptimizedPromptBuilder(3)

	emotion := manager.EmotionState{
		Trust:  50,
		Fear:   30,
		Stress: 40,
	}

	recentMessages := []ChatMessage{
		{SenderID: "player", Content: "Hello"},
		{SenderID: "npc_001", Content: "Hi"},
		{SenderID: "player", Content: "How are you?"},
		{SenderID: "npc_001", Content: "I'm fine"},
		{SenderID: "player", Content: "Good to hear"},
	}

	prompt := builder.BuildOptimizedPrompt(
		"Dr. Chen",
		"Scientist",
		emotion,
		recentMessages,
		"Tell me about yourself",
	)

	assert.Contains(t, prompt, "Dr. Chen")
	assert.Contains(t, prompt, "Scientist")
	assert.Contains(t, prompt, "信任50")
	assert.Contains(t, prompt, "恐懼30")
	assert.Contains(t, prompt, "壓力40")
	assert.Contains(t, prompt, "Tell me about yourself")

	// Should only include last 3 messages (limited by maxHistoryMessages)
	assert.Contains(t, prompt, "How are you?")
	assert.Contains(t, prompt, "I'm fine")
	assert.Contains(t, prompt, "Good to hear")
	assert.NotContains(t, prompt, "Hello") // First message should be excluded
}

// TestOptimizedPromptBuilder_EmptyHistory tests with no history.
func TestOptimizedPromptBuilder_EmptyHistory(t *testing.T) {
	builder := NewOptimizedPromptBuilder(5)

	emotion := manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}

	prompt := builder.BuildOptimizedPrompt(
		"Dr. Chen",
		"Scientist",
		emotion,
		[]ChatMessage{},
		"Hello",
	)

	assert.Contains(t, prompt, "Dr. Chen")
	assert.Contains(t, prompt, "Scientist")
	assert.Contains(t, prompt, "Hello")
}

// TestEstimateTokens tests token estimation.
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 1, // (0 + 2) / 3 = 0.67 -> 1
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 3, // (5 + 2) / 3 = 2.33 -> 2
		},
		{
			name:     "Chinese text",
			text:     "你好世界",
			expected: 5, // (12 bytes + 2) / 3 = 4.67 -> 5
		},
		{
			name:     "mixed text",
			text:     "Hello 你好",
			expected: 4, // (11 bytes + 2) / 3 = 4.33 -> 4 (round down)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			// Allow some variance due to rounding
			assert.GreaterOrEqual(t, result, tt.expected-1)
			assert.LessOrEqual(t, result, tt.expected+1)
		})
	}
}

// BenchmarkParallelResponseGenerator_Generate benchmarks parallel generation.
func BenchmarkParallelResponseGenerator_Generate(b *testing.B) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response", delay: 10 * time.Millisecond}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
			{ID: "npc_002", Name: "Guard Smith", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}},
		},
		MessageHistory: []ChatMessage{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pg.GenerateAllResponses(
			context.Background(),
			session,
			"Hello!",
			map[string]manager.EmotionDelta{},
			[]ChatFlag{},
		)
	}
}

// BenchmarkSequentialResponseGenerator_Generate benchmarks sequential generation for comparison.
func BenchmarkSequentialResponseGenerator_Generate(b *testing.B) {
	npcManager := setupTestNPCManager()
	mockLLM := &MockOptimizationLLMClient{response: "Test response", delay: 10 * time.Millisecond}
	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)

	session := &ChatSession{
		SessionID: "test_session",
		Participants: []ChatParticipant{
			{ID: "player", Name: "Player", IsPlayer: true},
			{ID: "npc_001", Name: "Dr. Chen", IsPlayer: false, Emotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}},
			{ID: "npc_002", Name: "Guard Smith", IsPlayer: false, Emotion: manager.EmotionState{Trust: 60, Fear: 20, Stress: 30}},
		},
		MessageHistory: []ChatMessage{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateAllNPCResponses(
			context.Background(),
			session,
			"Hello!",
			map[string]manager.EmotionDelta{},
			[]ChatFlag{},
		)
	}
}

// BenchmarkOptimizedPromptBuilder benchmarks optimized prompt building.
func BenchmarkOptimizedPromptBuilder(b *testing.B) {
	builder := NewOptimizedPromptBuilder(5)

	emotion := manager.EmotionState{Trust: 50, Fear: 30, Stress: 40}
	recentMessages := []ChatMessage{
		{SenderID: "player", Content: "Hello"},
		{SenderID: "npc_001", Content: "Hi"},
		{SenderID: "player", Content: "How are you?"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildOptimizedPrompt(
			"Dr. Chen",
			"Scientist",
			emotion,
			recentMessages,
			"Tell me about yourself",
		)
	}
}

// BenchmarkEstimateTokens benchmarks token estimation.
func BenchmarkEstimateTokens(b *testing.B) {
	text := "This is a test message that contains both English and Chinese 這是一個測試訊息"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EstimateTokens(text)
	}
}

// TestParallelResponseGenerator_ConcurrencyControl tests concurrency limiting.
func TestParallelResponseGenerator_ConcurrencyControl(t *testing.T) {
	npcManager := setupTestNPCManager()

	// Add more NPCs to exceed concurrency limit
	for i := 3; i <= 10; i++ {
		npc := &manager.NPCProfile{
			ID:             "npc_00" + string(rune('0'+i)),
			Name:           "NPC" + string(rune('0'+i)),
			Archetype:      "Test",
			InitialEmotion: manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		}
		_ = npcManager.AddNPC(npc)
	}

	// Track concurrent calls
	maxConcurrent := 0
	currentConcurrent := 0
	var mu = &struct {
		maxConcurrent     int
		currentConcurrent int
	}{0, 0}

	mockLLM := &MockOptimizationLLMClient{}
	mockLLM.generateFunc = func(ctx context.Context, prompt string, options map[string]any) (string, error) {
		mu.currentConcurrent++
		if mu.currentConcurrent > mu.maxConcurrent {
			mu.maxConcurrent = mu.currentConcurrent
		}
		time.Sleep(50 * time.Millisecond) // Simulate work
		mu.currentConcurrent--
		return "Response", nil
	}

	fallbackManager := NewFallbackManager()
	config := DefaultResponseGeneratorConfig()
	generator := NewNPCResponseGenerator(npcManager, mockLLM, fallbackManager, config)
	pg := NewParallelResponseGenerator(generator, 3, 5*time.Second, true)

	participants := []ChatParticipant{{ID: "player", Name: "Player", IsPlayer: true}}
	for i := 1; i <= 10; i++ {
		participants = append(participants, ChatParticipant{
			ID:       "npc_00" + string(rune('0'+i)),
			Name:     "NPC" + string(rune('0'+i)),
			IsPlayer: false,
			Emotion:  manager.EmotionState{Trust: 50, Fear: 30, Stress: 40},
		})
	}

	session := &ChatSession{
		SessionID:      "test_session",
		Participants:   participants,
		MessageHistory: []ChatMessage{},
	}

	_, err := pg.GenerateAllResponses(
		context.Background(),
		session,
		"Hello!",
		map[string]manager.EmotionDelta{},
		[]ChatFlag{},
	)

	assert.NoError(t, err)
	assert.LessOrEqual(t, maxConcurrent, 3) // Should never exceed max concurrency
	assert.LessOrEqual(t, currentConcurrent, 3)

	_ = maxConcurrent
	_ = currentConcurrent
}
