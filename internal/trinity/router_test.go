package trinity

import (
	"context"
	"errors"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// mockProvider implements the client.Provider interface for testing
type mockProvider struct {
	name            string
	sendMessageFunc func(ctx context.Context, messages []client.Message) (*client.Response, error)
	testConnFunc    func(ctx context.Context) error
	modelInfo       client.ModelInfo
	callCount       int
	shouldFail      bool
	failureError    error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) TestConnection(ctx context.Context) error {
	if m.testConnFunc != nil {
		return m.testConnFunc(ctx)
	}
	return nil
}

func (m *mockProvider) SendMessage(ctx context.Context, messages []client.Message) (*client.Response, error) {
	m.callCount++
	// Check shouldFail first (used by fallback integration tests)
	if m.shouldFail && m.failureError != nil {
		return nil, m.failureError
	}
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, messages)
	}
	return &client.Response{
		Content: "mock response from " + m.name,
		Metadata: map[string]interface{}{
			"provider": m.name,
		},
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, messages []client.Message, callback func(chunk string)) error {
	return errors.New("streaming not implemented in mock")
}

func (m *mockProvider) ModelInfo() client.ModelInfo {
	return m.modelInfo
}

// createMockRouter creates a router with mock providers for testing
func createMockRouter() (*TrinityRouter, *mockProvider, *mockProvider, *mockProvider) {
	thinkingMock := &mockProvider{
		name: "thinking-provider",
		modelInfo: client.ModelInfo{
			Provider:  "anthropic",
			Model:     "claude-opus-4-20250514",
			MaxTokens: 16000,
		},
	}

	reactiveMock := &mockProvider{
		name: "reactive-provider",
		modelInfo: client.ModelInfo{
			Provider:  "anthropic",
			Model:     "claude-3-5-sonnet-20241022",
			MaxTokens: 8000,
		},
	}

	rapidMock := &mockProvider{
		name: "rapid-provider",
		modelInfo: client.ModelInfo{
			Provider:  "anthropic",
			Model:     "claude-3-haiku-20240307",
			MaxTokens: 4000,
		},
	}

	router := &TrinityRouter{
		thinkingProvider: thinkingMock,
		reactiveProvider: reactiveMock,
		rapidProvider:    rapidMock,
		agentTierMap:     make(map[string]TierLevel),
		fallbackEnabled:  false,
	}

	// Copy default mapping
	for agent, tier := range DefaultAgentTierMapping {
		router.agentTierMap[agent] = tier
	}

	return router, thinkingMock, reactiveMock, rapidMock
}

func TestNewTrinityRouter_Success(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-opus-4-20250514",
			MaxTokens:  16000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-5-sonnet-20241022",
			MaxTokens:  8000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-haiku-20240307",
			MaxTokens:  4000,
		},
		AgentTierOverrides: map[string]TierLevel{
			"CustomAgent": TierThinking,
		},
		FallbackEnabled: true,
	}

	router, err := NewTrinityRouter(cfg)
	if err != nil {
		t.Fatalf("NewTrinityRouter() failed: %v", err)
	}

	if router == nil {
		t.Fatal("NewTrinityRouter() returned nil router")
	}

	if router.thinkingProvider == nil {
		t.Error("thinkingProvider is nil")
	}
	if router.reactiveProvider == nil {
		t.Error("reactiveProvider is nil")
	}
	if router.rapidProvider == nil {
		t.Error("rapidProvider is nil")
	}

	// Verify default mapping was loaded
	if router.agentTierMap["JudgeAgent"] != TierThinking {
		t.Error("Default mapping for JudgeAgent not loaded")
	}

	// Verify override was applied
	if router.agentTierMap["CustomAgent"] != TierThinking {
		t.Error("Override for CustomAgent not applied")
	}

	if !router.fallbackEnabled {
		t.Error("fallbackEnabled should be true")
	}
}

func TestNewTrinityRouter_InvalidProvider(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "invalid-provider",
			APIKey:     "test-key",
			Model:      "some-model",
			MaxTokens:  8000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-5-sonnet-20241022",
			MaxTokens:  8000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-haiku-20240307",
			MaxTokens:  4000,
		},
	}

	_, err := NewTrinityRouter(cfg)
	if err == nil {
		t.Fatal("NewTrinityRouter() should fail with invalid provider")
	}

	expectedErr := "failed to create Thinking provider"
	if len(err.Error()) < len(expectedErr) || err.Error()[:len(expectedErr)] != expectedErr {
		t.Errorf("Expected error containing %q, got: %v", expectedErr, err)
	}
}

func TestRoute_ThinkingTier(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()
	ctx := context.Background()

	// Test agents that should route to Thinking tier
	thinkingAgents := []string{"JudgeAgent", "SeedAgent", "NPCAgent", "ContradictionAgent"}

	for _, agentName := range thinkingAgents {
		t.Run(agentName, func(t *testing.T) {
			// Reset call counts
			thinkingMock.callCount = 0
			reactiveMock.callCount = 0
			rapidMock.callCount = 0

			messages := []client.Message{
				{Role: "user", Content: "test message"},
			}

			resp, err := router.Route(ctx, agentName, messages)
			if err != nil {
				t.Fatalf("Route() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Route() returned nil response")
			}

			// Verify only thinking provider was called
			if thinkingMock.callCount != 1 {
				t.Errorf("Expected thinkingProvider to be called once, got %d", thinkingMock.callCount)
			}
			if reactiveMock.callCount != 0 {
				t.Errorf("reactiveProvider should not be called, got %d calls", reactiveMock.callCount)
			}
			if rapidMock.callCount != 0 {
				t.Errorf("rapidProvider should not be called, got %d calls", rapidMock.callCount)
			}

			if resp.Metadata["provider"] != "thinking-provider" {
				t.Errorf("Expected response from thinking-provider, got %v", resp.Metadata["provider"])
			}
		})
	}
}

func TestRoute_ReactiveTier(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()
	ctx := context.Background()

	// Test agents that should route to Reactive tier
	reactiveAgents := []string{"NarrationAgent", "ChoiceAgent", "ChatProcessor"}

	for _, agentName := range reactiveAgents {
		t.Run(agentName, func(t *testing.T) {
			// Reset call counts
			thinkingMock.callCount = 0
			reactiveMock.callCount = 0
			rapidMock.callCount = 0

			messages := []client.Message{
				{Role: "user", Content: "test message"},
			}

			resp, err := router.Route(ctx, agentName, messages)
			if err != nil {
				t.Fatalf("Route() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Route() returned nil response")
			}

			// Verify only reactive provider was called
			if reactiveMock.callCount != 1 {
				t.Errorf("Expected reactiveProvider to be called once, got %d", reactiveMock.callCount)
			}
			if thinkingMock.callCount != 0 {
				t.Errorf("thinkingProvider should not be called, got %d calls", thinkingMock.callCount)
			}
			if rapidMock.callCount != 0 {
				t.Errorf("rapidProvider should not be called, got %d calls", rapidMock.callCount)
			}

			if resp.Metadata["provider"] != "reactive-provider" {
				t.Errorf("Expected response from reactive-provider, got %v", resp.Metadata["provider"])
			}
		})
	}
}

func TestRoute_RapidTier(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()
	ctx := context.Background()

	// Test agents that should route to Rapid tier
	rapidAgents := []string{"DreamAgent", "EnvironmentAgent", "SummaryAgent"}

	for _, agentName := range rapidAgents {
		t.Run(agentName, func(t *testing.T) {
			// Reset call counts
			thinkingMock.callCount = 0
			reactiveMock.callCount = 0
			rapidMock.callCount = 0

			messages := []client.Message{
				{Role: "user", Content: "test message"},
			}

			resp, err := router.Route(ctx, agentName, messages)
			if err != nil {
				t.Fatalf("Route() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Route() returned nil response")
			}

			// Verify only rapid provider was called
			if rapidMock.callCount != 1 {
				t.Errorf("Expected rapidProvider to be called once, got %d", rapidMock.callCount)
			}
			if thinkingMock.callCount != 0 {
				t.Errorf("thinkingProvider should not be called, got %d calls", thinkingMock.callCount)
			}
			if reactiveMock.callCount != 0 {
				t.Errorf("reactiveProvider should not be called, got %d calls", reactiveMock.callCount)
			}

			if resp.Metadata["provider"] != "rapid-provider" {
				t.Errorf("Expected response from rapid-provider, got %v", resp.Metadata["provider"])
			}
		})
	}
}

func TestRoute_UnknownAgent_DefaultsToReactive(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()
	ctx := context.Background()

	// Test unknown agent
	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	resp, err := router.Route(ctx, "UnknownAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Route() returned nil response")
	}

	// Should default to Reactive tier
	if reactiveMock.callCount != 1 {
		t.Errorf("Expected reactiveProvider to be called once (default), got %d", reactiveMock.callCount)
	}
	if thinkingMock.callCount != 0 {
		t.Errorf("thinkingProvider should not be called, got %d calls", thinkingMock.callCount)
	}
	if rapidMock.callCount != 0 {
		t.Errorf("rapidProvider should not be called, got %d calls", rapidMock.callCount)
	}
}

func TestRoute_UserOverride(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()
	ctx := context.Background()

	// Override DreamAgent (normally Rapid) to Thinking
	router.agentTierMap["DreamAgent"] = TierThinking

	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	resp, err := router.Route(ctx, "DreamAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Route() returned nil response")
	}

	// Should use Thinking tier (override)
	if thinkingMock.callCount != 1 {
		t.Errorf("Expected thinkingProvider to be called (override), got %d", thinkingMock.callCount)
	}
	if reactiveMock.callCount != 0 {
		t.Errorf("reactiveProvider should not be called, got %d calls", reactiveMock.callCount)
	}
	if rapidMock.callCount != 0 {
		t.Errorf("rapidProvider should not be called, got %d calls", rapidMock.callCount)
	}
}

func TestRoute_ProviderError(t *testing.T) {
	router, _, _, _ := createMockRouter()
	ctx := context.Background()

	// Create a provider that returns an error
	errorProvider := &mockProvider{
		name: "error-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, errors.New("provider failure")
		},
	}

	// Override thinking provider with error provider
	router.thinkingProvider = errorProvider

	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	_, err := router.Route(ctx, "JudgeAgent", messages)
	if err == nil {
		t.Fatal("Route() should fail when provider returns error")
	}

	expectedErr := "tier Thinking failed for agent JudgeAgent"
	if len(err.Error()) < len(expectedErr) || err.Error()[:len(expectedErr)] != expectedErr {
		t.Errorf("Expected error containing %q, got: %v", expectedErr, err)
	}
}

func TestRoute_MultipleMessages(t *testing.T) {
	router, thinkingMock, _, _ := createMockRouter()
	ctx := context.Background()

	messages := []client.Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	resp, err := router.Route(ctx, "JudgeAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Route() returned nil response")
	}

	if thinkingMock.callCount != 1 {
		t.Errorf("Expected thinkingProvider to be called once, got %d", thinkingMock.callCount)
	}
}

func TestGetProviderForTier(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()

	tests := []struct {
		tier     TierLevel
		expected *mockProvider
	}{
		{TierThinking, thinkingMock},
		{TierReactive, reactiveMock},
		{TierRapid, rapidMock},
		{TierLevel(999), reactiveMock}, // Unknown tier defaults to Reactive
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			provider := router.getProviderForTier(tt.tier)
			if provider != tt.expected {
				t.Errorf("getProviderForTier(%v) returned wrong provider", tt.tier)
			}
		})
	}
}

func TestRoute_EmptyMessages(t *testing.T) {
	router, _, reactiveMock, _ := createMockRouter()
	ctx := context.Background()

	// Test with empty messages slice
	messages := []client.Message{}

	resp, err := router.Route(ctx, "UnknownAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed with empty messages: %v", err)
	}

	if resp == nil {
		t.Fatal("Route() returned nil response")
	}

	if reactiveMock.callCount != 1 {
		t.Errorf("Expected reactiveProvider to be called once, got %d", reactiveMock.callCount)
	}
}

func TestRoute_AllAgentsHaveMapping(t *testing.T) {
	router, _, _, _ := createMockRouter()
	ctx := context.Background()

	// Test all default agents have correct mappings
	for agentName, expectedTier := range DefaultAgentTierMapping {
		t.Run(agentName, func(t *testing.T) {
			actualTier := GetTierForAgent(agentName, router.agentTierMap)
			if actualTier != expectedTier {
				t.Errorf("Agent %q: expected tier %v, got %v", agentName, expectedTier, actualTier)
			}

			// Verify routing works for this agent
			messages := []client.Message{{Role: "user", Content: "test"}}
			_, err := router.Route(ctx, agentName, messages)
			if err != nil {
				t.Errorf("Route() failed for agent %q: %v", agentName, err)
			}
		})
	}
}

func TestRoute_ContextCancellation(t *testing.T) {
	router, _, _, _ := createMockRouter()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Override provider to check context
	errorProvider := &mockProvider{
		name: "context-check-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return &client.Response{Content: "should not reach here"}, nil
		},
	}
	router.reactiveProvider = errorProvider

	messages := []client.Message{{Role: "user", Content: "test"}}
	_, err := router.Route(ctx, "UnknownAgent", messages)

	if err == nil {
		t.Fatal("Route() should fail with cancelled context")
	}

	if !errors.Is(err, context.Canceled) && err.Error() != "tier Reactive failed for agent UnknownAgent: context canceled" {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestDefaultRouterConfig(t *testing.T) {
	cfg := DefaultRouterConfig()

	// Verify Thinking tier config
	if cfg.ThinkingProvider.ProviderID != "anthropic" {
		t.Errorf("Expected ThinkingProvider.ProviderID = anthropic, got %s", cfg.ThinkingProvider.ProviderID)
	}
	if cfg.ThinkingProvider.Model != "claude-opus-4-20250514" {
		t.Errorf("Expected ThinkingProvider.Model = claude-opus-4-20250514, got %s", cfg.ThinkingProvider.Model)
	}

	// Verify Reactive tier config
	if cfg.ReactiveProvider.ProviderID != "anthropic" {
		t.Errorf("Expected ReactiveProvider.ProviderID = anthropic, got %s", cfg.ReactiveProvider.ProviderID)
	}
	if cfg.ReactiveProvider.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected ReactiveProvider.Model = claude-3-5-sonnet-20241022, got %s", cfg.ReactiveProvider.Model)
	}

	// Verify Rapid tier config
	if cfg.RapidProvider.ProviderID != "anthropic" {
		t.Errorf("Expected RapidProvider.ProviderID = anthropic, got %s", cfg.RapidProvider.ProviderID)
	}
	if cfg.RapidProvider.Model != "claude-3-haiku-20240307" {
		t.Errorf("Expected RapidProvider.Model = claude-3-haiku-20240307, got %s", cfg.RapidProvider.Model)
	}

	// Verify fallback is enabled by default
	if !cfg.FallbackEnabled {
		t.Error("Expected FallbackEnabled to be true by default")
	}

	// Verify overrides map is initialized
	if cfg.AgentTierOverrides == nil {
		t.Error("Expected AgentTierOverrides to be initialized")
	}
}

func TestProviderTierConfig_CreateProvider_Anthropic(t *testing.T) {
	cfg := ProviderTierConfig{
		ProviderID: "anthropic",
		APIKey:     "test-key",
		Model:      "claude-opus-4-20250514",
		MaxTokens:  16000,
	}

	provider, err := cfg.CreateProvider()
	if err != nil {
		t.Fatalf("CreateProvider() failed: %v", err)
	}

	if provider == nil {
		t.Fatal("CreateProvider() returned nil provider")
	}

	info := provider.ModelInfo()
	if info.Provider != "anthropic" {
		t.Errorf("Expected provider = anthropic, got %s", info.Provider)
	}
}

func TestProviderTierConfig_CreateProvider_OpenAI(t *testing.T) {
	cfg := ProviderTierConfig{
		ProviderID: "openai",
		APIKey:     "test-key",
		Model:      "gpt-4",
		MaxTokens:  8000,
	}

	provider, err := cfg.CreateProvider()
	if err != nil {
		t.Fatalf("CreateProvider() failed: %v", err)
	}

	if provider == nil {
		t.Fatal("CreateProvider() returned nil provider")
	}

	info := provider.ModelInfo()
	if info.Provider != "openai" {
		t.Errorf("Expected provider = openai, got %s", info.Provider)
	}
}

func TestProviderTierConfig_CreateProvider_OpenRouter(t *testing.T) {
	cfg := ProviderTierConfig{
		ProviderID: "openrouter",
		APIKey:     "test-key",
		Model:      "anthropic/claude-opus-4",
		MaxTokens:  8000,
	}

	provider, err := cfg.CreateProvider()
	if err != nil {
		t.Fatalf("CreateProvider() failed: %v", err)
	}

	if provider == nil {
		t.Fatal("CreateProvider() returned nil provider")
	}

	info := provider.ModelInfo()
	if info.Provider != "openrouter" {
		t.Errorf("Expected provider = openrouter, got %s", info.Provider)
	}
}

func TestProviderTierConfig_CreateProvider_InvalidProvider(t *testing.T) {
	cfg := ProviderTierConfig{
		ProviderID: "invalid-provider",
		APIKey:     "test-key",
		Model:      "some-model",
		MaxTokens:  8000,
	}

	_, err := cfg.CreateProvider()
	if err == nil {
		t.Fatal("CreateProvider() should fail with invalid provider")
	}

	expectedErr := "unknown provider: invalid-provider"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got: %v", expectedErr, err)
	}
}

func TestNewTrinityRouter_MergesDefaultAndOverrides(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-opus-4-20250514",
			MaxTokens:  16000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-5-sonnet-20241022",
			MaxTokens:  8000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "anthropic",
			APIKey:     "test-key",
			Model:      "claude-3-haiku-20240307",
			MaxTokens:  4000,
		},
		AgentTierOverrides: map[string]TierLevel{
			"JudgeAgent":  TierRapid,    // Override default
			"CustomAgent": TierThinking, // New agent
		},
	}

	router, err := NewTrinityRouter(cfg)
	if err != nil {
		t.Fatalf("NewTrinityRouter() failed: %v", err)
	}

	// Verify override applied
	if tier := router.agentTierMap["JudgeAgent"]; tier != TierRapid {
		t.Errorf("JudgeAgent tier = %v, want %v (override)", tier, TierRapid)
	}

	// Verify new agent added
	if tier := router.agentTierMap["CustomAgent"]; tier != TierThinking {
		t.Errorf("CustomAgent tier = %v, want %v", tier, TierThinking)
	}

	// Verify default agents still work
	if tier := router.agentTierMap["NPCAgent"]; tier != TierThinking {
		t.Errorf("NPCAgent tier = %v, want %v (default)", tier, TierThinking)
	}
}

// Story 9-2: Integration Tests for ThinkingMiddleware in TrinityRouter

// TestRoute_ThinkingMiddleware_ProcessesTags tests that thinking tags are automatically processed
func TestRoute_ThinkingMiddleware_ProcessesTags(t *testing.T) {
	router, thinkingMock, _, _ := createMockRouter()

	// Initialize middleware (normally done in NewTrinityRouter)
	router.thinkingMiddleware = NewThinkingMiddleware()

	// Mock provider response with thinking tags
	thinkingMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return &client.Response{
			Content: "Here is the answer. <thinking>I analyzed this carefully and considered multiple factors</thinking> The final conclusion.",
			Metadata: map[string]interface{}{
				"model": "claude-opus",
			},
		}, nil
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test question"}}

	// Route to Thinking tier (JudgeAgent uses Thinking tier)
	resp, err := router.Route(ctx, "JudgeAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	// Verify thinking tags were removed from content
	expectedContent := "Here is the answer.  The final conclusion."
	if resp.Content != expectedContent {
		t.Errorf("Content = %q, want %q", resp.Content, expectedContent)
	}

	// Verify thinking chain was extracted to metadata
	thinkingChain, ok := resp.Metadata["thinking_chain"].(string)
	if !ok {
		t.Fatal("Expected thinking_chain in metadata")
	}

	expectedChain := "I analyzed this carefully and considered multiple factors"
	if thinkingChain != expectedChain {
		t.Errorf("thinking_chain = %q, want %q", thinkingChain, expectedChain)
	}

	// Verify original metadata preserved
	if resp.Metadata["model"] != "claude-opus" {
		t.Error("Original metadata should be preserved")
	}
}

// TestRoute_ThinkingMiddleware_NoTagsPassThrough tests that responses without thinking tags pass through unchanged
func TestRoute_ThinkingMiddleware_NoTagsPassThrough(t *testing.T) {
	router, thinkingMock, _, _ := createMockRouter()
	router.thinkingMiddleware = NewThinkingMiddleware()

	originalContent := "This is a normal response without any thinking tags."
	thinkingMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return &client.Response{
			Content: originalContent,
		}, nil
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := router.Route(ctx, "NPCAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	// Content should be unchanged
	if resp.Content != originalContent {
		t.Errorf("Content = %q, want %q", resp.Content, originalContent)
	}

	// No thinking_chain should be in metadata
	if resp.Metadata != nil {
		if _, ok := resp.Metadata["thinking_chain"]; ok {
			t.Error("thinking_chain should not be present when no tags")
		}
	}
}

// TestRoute_ThinkingMiddleware_OnlyForThinkingTier tests that middleware only processes Thinking tier responses
func TestRoute_ThinkingMiddleware_OnlyForThinkingTier(t *testing.T) {
	router, _, reactiveMock, rapidMock := createMockRouter()
	router.thinkingMiddleware = NewThinkingMiddleware()

	contentWithTags := "Answer <thinking>internal reasoning</thinking> result"

	tests := []struct {
		name      string
		agentName string
		tier      TierLevel
		provider  *mockProvider
	}{
		{
			name:      "Reactive tier should NOT process thinking tags",
			agentName: "NarrationAgent",
			tier:      TierReactive,
			provider:  reactiveMock,
		},
		{
			name:      "Rapid tier should NOT process thinking tags",
			agentName: "DreamAgent",
			tier:      TierRapid,
			provider:  rapidMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.provider.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
				return &client.Response{
					Content: contentWithTags,
				}, nil
			}

			ctx := context.Background()
			messages := []client.Message{{Role: "user", Content: "Test"}}

			resp, err := router.Route(ctx, tt.agentName, messages)
			if err != nil {
				t.Fatalf("Route() failed: %v", err)
			}

			// For non-Thinking tiers, thinking tags should NOT be processed
			// (middleware should only run for Thinking tier)
			if resp.Content != contentWithTags {
				t.Errorf("Content should be unchanged for %s tier, got %q", tt.tier.String(), resp.Content)
			}

			if resp.Metadata != nil {
				if _, ok := resp.Metadata["thinking_chain"]; ok {
					t.Errorf("thinking_chain should not be present for %s tier", tt.tier.String())
				}
			}
		})
	}
}

// TestRoute_ThinkingMiddleware_MultilineThinking tests multiline thinking tag processing
func TestRoute_ThinkingMiddleware_MultilineThinking(t *testing.T) {
	router, thinkingMock, _, _ := createMockRouter()
	router.thinkingMiddleware = NewThinkingMiddleware()

	multilineContent := `Initial analysis.
<thinking>
Line 1: Consider user intent
Line 2: Evaluate options
Line 3: Make decision
</thinking>
Final response.`

	thinkingMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return &client.Response{
			Content: multilineContent,
		}, nil
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Complex question"}}

	resp, err := router.Route(ctx, "ContradictionAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	// Verify thinking content extracted
	thinkingChain, ok := resp.Metadata["thinking_chain"].(string)
	if !ok {
		t.Fatal("Expected thinking_chain in metadata")
	}

	// Check that all lines are present
	expectedLines := []string{"Line 1: Consider user intent", "Line 2: Evaluate options", "Line 3: Make decision"}
	for _, line := range expectedLines {
		if !contains(thinkingChain, line) {
			t.Errorf("thinking_chain missing line: %q", line)
		}
	}

	// Verify tags removed from content
	if contains(resp.Content, "thinking") || contains(resp.Content, "Line 1") {
		t.Error("Thinking tags and content should be removed from response")
	}
}

// TestRoute_ThinkingMiddleware_WithFallback tests thinking middleware with fallback enabled
func TestRoute_ThinkingMiddleware_WithFallback(t *testing.T) {
	router, thinkingMock, _, _ := createMockRouter()
	router.thinkingMiddleware = NewThinkingMiddleware()
	router.fallbackEnabled = true
	router.fallbackHandler = NewDefaultFallbackHandler()

	thinkingMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return &client.Response{
			Content: "Response <thinking>reasoning process</thinking> conclusion",
		}, nil
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := router.Route(ctx, "SeedAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	// Even with fallback enabled, thinking tags should be processed
	if contains(resp.Content, "thinking") {
		t.Error("Thinking tags should be removed even with fallback enabled")
	}

	if _, ok := resp.Metadata["thinking_chain"]; !ok {
		t.Error("thinking_chain should be in metadata")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAtAnyPosition(s, substr))
}

func containsAtAnyPosition(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
// Story 9-5: Metrics Integration Tests

func TestRouter_MetricsIntegration_SuccessfulRequest(t *testing.T) {
	router, _, reactiveMock, _ := createMockRouter()

	// Initialize metrics
	router.metrics = NewTrinityMetrics(100)

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	// Route a request to NarrationAgent (Reactive tier)
	_, err := router.Route(ctx, "NarrationAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	// Verify metrics were recorded
	stats := router.metrics.GetTierStats(TierReactive)

	if stats.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", stats.TotalRequests)
	}

	if stats.SuccessRequests != 1 {
		t.Errorf("SuccessRequests = %d, want 1", stats.SuccessRequests)
	}

	if stats.FailedRequests != 0 {
		t.Errorf("FailedRequests = %d, want 0", stats.FailedRequests)
	}

	if reactiveMock.callCount != 1 {
		t.Errorf("Provider call count = %d, want 1", reactiveMock.callCount)
	}

	// Verify duration was recorded
	if stats.MinDuration == 0 {
		t.Error("MinDuration should be > 0")
	}
}

func TestRouter_MetricsIntegration_FailedRequest(t *testing.T) {
	router, _, reactiveMock, _ := createMockRouter()

	// Initialize metrics
	router.metrics = NewTrinityMetrics(100)

	// Set up mock to fail
	testErr := errors.New("provider error")
	reactiveMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return nil, testErr
	}

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	// Route a request that will fail
	_, err := router.Route(ctx, "NarrationAgent", messages)
	if err == nil {
		t.Fatal("Route() should have failed but didn't")
	}

	// Verify metrics were recorded
	stats := router.metrics.GetTierStats(TierReactive)

	if stats.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", stats.TotalRequests)
	}

	if stats.SuccessRequests != 0 {
		t.Errorf("SuccessRequests = %d, want 0", stats.SuccessRequests)
	}

	if stats.FailedRequests != 1 {
		t.Errorf("FailedRequests = %d, want 1", stats.FailedRequests)
	}

	if stats.ErrorRate != 1.0 {
		t.Errorf("ErrorRate = %f, want 1.0", stats.ErrorRate)
	}

	if stats.LastError == nil {
		t.Error("LastError should be set")
	}
}

func TestRouter_MetricsIntegration_MultipleRequests(t *testing.T) {
	router, thinkingMock, reactiveMock, rapidMock := createMockRouter()

	// Initialize metrics
	router.metrics = NewTrinityMetrics(100)

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	// Route requests to different tiers
	router.Route(ctx, "JudgeAgent", messages)      // Thinking tier
	router.Route(ctx, "NarrationAgent", messages)  // Reactive tier
	router.Route(ctx, "DreamAgent", messages)      // Rapid tier
	router.Route(ctx, "NPCAgent", messages)        // Thinking tier
	router.Route(ctx, "ChoiceAgent", messages)     // Reactive tier

	// Verify total requests
	summary := router.GetMetricsSummary()

	if summary.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", summary.TotalRequests)
	}

	// Verify per-tier counts
	if summary.ThinkingStats.TotalRequests != 2 {
		t.Errorf("ThinkingStats.TotalRequests = %d, want 2", summary.ThinkingStats.TotalRequests)
	}

	if summary.ReactiveStats.TotalRequests != 2 {
		t.Errorf("ReactiveStats.TotalRequests = %d, want 2", summary.ReactiveStats.TotalRequests)
	}

	if summary.RapidStats.TotalRequests != 1 {
		t.Errorf("RapidStats.TotalRequests = %d, want 1", summary.RapidStats.TotalRequests)
	}

	// Verify provider call counts
	if thinkingMock.callCount != 2 {
		t.Errorf("Thinking provider call count = %d, want 2", thinkingMock.callCount)
	}

	if reactiveMock.callCount != 2 {
		t.Errorf("Reactive provider call count = %d, want 2", reactiveMock.callCount)
	}

	if rapidMock.callCount != 1 {
		t.Errorf("Rapid provider call count = %d, want 1", rapidMock.callCount)
	}
}

func TestRouter_MetricsIntegration_WithFallback(t *testing.T) {
	router, thinkingMock, reactiveMock, _ := createMockRouter()

	// Enable fallback
	router.fallbackEnabled = true
	router.fallbackHandler = NewDefaultFallbackHandler()
	router.metrics = NewTrinityMetrics(100)

	// Make Thinking tier fail
	thinkingMock.sendMessageFunc = func(ctx context.Context, messages []client.Message) (*client.Response, error) {
		return nil, &client.APIError{
			StatusCode: 503,
			Message:    "Service unavailable",
			Err:        client.ErrServiceUnavailable,
		}
	}

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test message"},
	}

	// Route request to JudgeAgent (Thinking tier, should fallback to Reactive)
	resp, err := router.Route(ctx, "JudgeAgent", messages)
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	// Verify downgrade was recorded
	summary := router.GetMetricsSummary()

	if summary.TotalDowngrades != 1 {
		t.Errorf("TotalDowngrades = %d, want 1", summary.TotalDowngrades)
	}

	// Verify the reactive tier handled the request
	if summary.ReactiveStats.TotalRequests != 1 {
		t.Errorf("ReactiveStats.TotalRequests = %d, want 1", summary.ReactiveStats.TotalRequests)
	}

	if summary.ReactiveStats.SuccessRequests != 1 {
		t.Errorf("ReactiveStats.SuccessRequests = %d, want 1", summary.ReactiveStats.SuccessRequests)
	}

	// Verify reactive provider was called
	if reactiveMock.callCount == 0 {
		t.Error("Reactive provider should have been called")
	}
}

func TestRouter_GetMetrics(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = NewTrinityMetrics(100)

	metrics := router.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics() should not return nil")
	}

	if metrics != router.metrics {
		t.Error("GetMetrics() should return the router's metrics instance")
	}
}

func TestRouter_GetMetricsSummary(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = NewTrinityMetrics(100)

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test"},
	}

	// Generate some metrics
	router.Route(ctx, "NPCAgent", messages)

	summary := router.GetMetricsSummary()

	if summary.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", summary.TotalRequests)
	}

	if summary.Uptime == 0 {
		t.Error("Uptime should be > 0")
	}
}

func TestRouter_GetMetricsSummary_NoMetrics(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = nil // No metrics initialized

	summary := router.GetMetricsSummary()

	// Should return empty summary without panicking
	if summary.TotalRequests != 0 {
		t.Errorf("TotalRequests = %d, want 0 (no metrics)", summary.TotalRequests)
	}
}

func TestRouter_ResetMetrics(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = NewTrinityMetrics(100)

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test"},
	}

	// Generate some metrics
	router.Route(ctx, "NPCAgent", messages)
	router.Route(ctx, "ChoiceAgent", messages)

	summary := router.GetMetricsSummary()
	if summary.TotalRequests != 2 {
		t.Fatalf("TotalRequests = %d, want 2", summary.TotalRequests)
	}

	// Reset metrics
	router.ResetMetrics()

	// Verify metrics are reset
	summary = router.GetMetricsSummary()
	if summary.TotalRequests != 0 {
		t.Errorf("TotalRequests after reset = %d, want 0", summary.TotalRequests)
	}
}

func TestRouter_LogMetricsSummary(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = NewTrinityMetrics(100)

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "test"},
	}

	// Generate some metrics
	router.Route(ctx, "NPCAgent", messages)

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogMetricsSummary panicked: %v", r)
		}
	}()

	router.LogMetricsSummary()
}

func TestRouter_LogMetricsSummary_NoMetrics(t *testing.T) {
	router, _, _, _ := createMockRouter()
	router.metrics = nil // No metrics

	// This should not panic even without metrics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogMetricsSummary panicked: %v", r)
		}
	}()

	router.LogMetricsSummary()
}
