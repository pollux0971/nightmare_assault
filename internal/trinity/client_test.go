package trinity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// TestNewTrinityLLMClient tests client creation
// AC1: NewTrinityLLMClient() can create client instance
func TestNewTrinityLLMClient(t *testing.T) {
	tests := []struct {
		name          string
		routerConfig  RouterConfig
		clientConfig  TrinityClientConfig
		expectError   bool
		errorContains string
	}{
		{
			name:         "Valid configuration",
			routerConfig: createTestRouterConfig(),
			clientConfig: TrinityClientConfig{
				EnableThinkingExtraction: true,
				EnableFallback:           true,
				EnableMetrics:            true,
				DefaultTimeout:           30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "Invalid provider",
			routerConfig: RouterConfig{
				ThinkingProvider: ProviderTierConfig{
					ProviderID: "invalid_provider",
					APIKey:     "test-key",
				},
				ReactiveProvider: createTestProviderConfig("anthropic"),
				RapidProvider:    createTestProviderConfig("anthropic"),
			},
			clientConfig: DefaultClientConfig(),
			expectError:  true,
			errorContains: "unknown provider",
		},
		{
			name:         "Metrics disabled",
			routerConfig: createTestRouterConfig(),
			clientConfig: TrinityClientConfig{
				EnableThinkingExtraction: false,
				EnableFallback:           false,
				EnableMetrics:            false,
				DefaultTimeout:           0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trinityClient, err := NewTrinityLLMClient(tt.routerConfig, tt.clientConfig)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if trinityClient == nil {
				t.Fatal("Expected client to be non-nil")
			}

			if trinityClient.router == nil {
				t.Error("Expected router to be non-nil")
			}

			if trinityClient.config.EnableFallback != tt.clientConfig.EnableFallback {
				t.Errorf("Expected EnableFallback=%v, got %v", tt.clientConfig.EnableFallback, trinityClient.config.EnableFallback)
			}
		})
	}
}

// TestSendMessage tests basic message sending
// AC1: SendMessage() can route based on agent name automatically
func TestSendMessage(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		expectedTier TierLevel
	}{
		{
			name:        "JudgeAgent uses Thinking tier",
			agentName:   "JudgeAgent",
			expectedTier: TierThinking,
		},
		{
			name:        "NarrationAgent uses Reactive tier",
			agentName:   "NarrationAgent",
			expectedTier: TierReactive,
		},
		{
			name:        "DreamAgent uses Rapid tier",
			agentName:   "DreamAgent",
			expectedTier: TierRapid,
		},
		{
			name:        "Unknown agent defaults to Reactive",
			agentName:   "UnknownAgent",
			expectedTier: TierReactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock providers
			mockProviders := createMockProviders()

			trinityClient := createMockClient(mockProviders, TrinityClientConfig{
				EnableThinkingExtraction: true,
				EnableFallback:           false, // Disable for predictable testing
				EnableMetrics:            true,
			})

			messages := []client.Message{
				{Role: "user", Content: "Test message"},
			}

			ctx := context.Background()
			resp, err := trinityClient.SendMessage(ctx, tt.agentName, messages)

			if err != nil {
				t.Fatalf("SendMessage failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Expected response to be non-nil")
			}

			// Verify the correct provider was called based on tier
			expectedProvider := mockProviders[tt.expectedTier]
			if resp.Content != expectedProvider.Name() {
				t.Errorf("Expected response from %s, got %s", expectedProvider.Name(), resp.Content)
			}
		})
	}
}

// TestSendMessageWithOptions tests advanced options
// AC5: SendMessageWithOptions() allows tier override
func TestSendMessageWithOptions_TierOverride(t *testing.T) {
	mockProviders := createMockProviders()

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           false,
		EnableMetrics:            true,
	})

	messages := []client.Message{
		{Role: "user", Content: "Test message"},
	}

	// Test tier override
	overrideTier := TierRapid
	opts := SendOptions{
		TierOverride: &overrideTier,
	}

	ctx := context.Background()
	resp, err := trinityClient.SendMessageWithOptions(ctx, "JudgeAgent", messages, opts)

	if err != nil {
		t.Fatalf("SendMessageWithOptions failed: %v", err)
	}

	// Should use Rapid tier instead of default Thinking tier
	expectedProvider := mockProviders[TierRapid]
	if resp.Content != expectedProvider.Name() {
		t.Errorf("Expected response from %s, got %s", expectedProvider.Name(), resp.Content)
	}
}

// TestSendMessageWithOptions_Timeout tests timeout handling
// AC5: Support context timeout
func TestSendMessageWithOptions_Timeout(t *testing.T) {
	mockProviders := createMockProvidersWithDelay(100 * time.Millisecond)

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: false,
		EnableMetrics:  true,
	})

	messages := []client.Message{
		{Role: "user", Content: "Test message"},
	}

	// Test with very short timeout
	opts := SendOptions{
		Timeout: 1 * time.Millisecond, // Shorter than provider delay
	}

	ctx := context.Background()
	_, err := trinityClient.SendMessageWithOptions(ctx, "NarrationAgent", messages, opts)

	if err == nil {
		t.Fatal("Expected timeout error but got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		// The error might be wrapped
		if !contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	}
}

// TestSendMessageWithOptions_DisableFallback tests fallback disabling
// AC3: SendMessageWithOptions() allows disabling fallback per request
func TestSendMessageWithOptions_DisableFallback(t *testing.T) {
	// Create provider that always fails
	failingProvider := &mockProvider{
		name: "failing-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, errors.New("provider failure")
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: failingProvider,
		TierReactive: createSuccessProvider("reactive-provider"),
		TierRapid:    createSuccessProvider("rapid-provider"),
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: true, // Fallback enabled globally
		EnableMetrics:  true,
	})

	messages := []client.Message{
		{Role: "user", Content: "Test message"},
	}

	// Test with fallback disabled for this request
	opts := SendOptions{
		DisableFallback: true,
	}

	ctx := context.Background()
	_, err := trinityClient.SendMessageWithOptions(ctx, "JudgeAgent", messages, opts)

	// Should fail immediately without fallback
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if !contains(err.Error(), "provider failure") {
		t.Errorf("Expected provider failure error, got: %v", err)
	}
}

// TestExtractThinking tests thinking tag extraction
// AC2: ExtractThinking() correctly extracts and cleans content
func TestExtractThinking(t *testing.T) {
	tests := []struct {
		name            string
		response        *client.Response
		expectedThinking string
		expectedContent  string
	}{
		{
			name: "Response with thinking tags in content",
			response: &client.Response{
				Content: "<thinking>This is my reasoning</thinking>This is the answer",
			},
			expectedThinking: "This is my reasoning",
			expectedContent:  "This is the answer",
		},
		{
			name: "Response with thinking in metadata",
			response: &client.Response{
				Content: "This is the answer",
				Metadata: map[string]interface{}{
					"thinking_chain": "This is my reasoning",
				},
			},
			expectedThinking: "This is my reasoning",
			expectedContent:  "This is the answer",
		},
		{
			name: "Response without thinking tags",
			response: &client.Response{
				Content: "Just a plain answer",
			},
			expectedThinking: "",
			expectedContent:  "Just a plain answer",
		},
		{
			name:            "Nil response",
			response:        nil,
			expectedThinking: "",
			expectedContent:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trinityClient, err := NewTrinityLLMClient(createTestRouterConfig(), DefaultClientConfig())
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			thinking, content := trinityClient.ExtractThinking(tt.response)

			if thinking != tt.expectedThinking {
				t.Errorf("Expected thinking=%q, got %q", tt.expectedThinking, thinking)
			}

			if content != tt.expectedContent {
				t.Errorf("Expected content=%q, got %q", tt.expectedContent, content)
			}
		})
	}
}

// TestHasThinking tests thinking tag detection
// AC2: HasThinking() correctly detects thinking tags
func TestHasThinking(t *testing.T) {
	tests := []struct {
		name     string
		response *client.Response
		expected bool
	}{
		{
			name: "Response with thinking tags in content",
			response: &client.Response{
				Content: "<thinking>reasoning</thinking>answer",
			},
			expected: true,
		},
		{
			name: "Response with thinking in metadata",
			response: &client.Response{
				Content: "answer",
				Metadata: map[string]interface{}{
					"thinking_chain": "reasoning",
				},
			},
			expected: true,
		},
		{
			name: "Response without thinking",
			response: &client.Response{
				Content: "just an answer",
			},
			expected: false,
		},
		{
			name:     "Nil response",
			response: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trinityClient, err := NewTrinityLLMClient(createTestRouterConfig(), DefaultClientConfig())
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := trinityClient.HasThinking(tt.response)

			if result != tt.expected {
				t.Errorf("Expected HasThinking=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetMetrics tests metrics retrieval
// AC4: GetMetrics() returns complete metrics summary
func TestGetMetrics(t *testing.T) {
	mockProviders := createMockProviders()

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableMetrics: true,
	})

	// Send some test requests
	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	_, _ = trinityClient.SendMessage(ctx, "JudgeAgent", messages)   // Thinking tier
	_, _ = trinityClient.SendMessage(ctx, "NarrationAgent", messages) // Reactive tier

	metrics := trinityClient.GetMetrics()

	if metrics.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.ThinkingStats.TotalRequests != 1 {
		t.Errorf("Expected 1 Thinking request, got %d", metrics.ThinkingStats.TotalRequests)
	}

	if metrics.ReactiveStats.TotalRequests != 1 {
		t.Errorf("Expected 1 Reactive request, got %d", metrics.ReactiveStats.TotalRequests)
	}
}

// TestGetMetrics_Disabled tests metrics when disabled
func TestGetMetrics_Disabled(t *testing.T) {
	mockProviders := createMockProviders()

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableMetrics: false,
	})

	metrics := trinityClient.GetMetrics()

	// Should return empty metrics
	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests when metrics disabled, got %d", metrics.TotalRequests)
	}
}

// TestGetFallbackMetrics tests fallback metrics retrieval
// AC3: Fallback metrics available when enabled
func TestGetFallbackMetrics(t *testing.T) {
	failingProvider := &mockProvider{
		name: "failing-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, client.NewAPIError("test", 500, "server error", nil)
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: failingProvider,
		TierReactive: createSuccessProvider("reactive-provider"),
		TierRapid:    createSuccessProvider("rapid-provider"),
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: true,
		EnableMetrics:  true,
	})

	// Send request that will trigger fallback
	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}
	_, _ = trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	fallbackMetrics := trinityClient.GetFallbackMetrics()

	if fallbackMetrics == nil {
		t.Fatal("Expected fallback metrics to be non-nil")
	}

	if fallbackMetrics.TotalFallbacks == 0 {
		t.Error("Expected at least one fallback event")
	}
}

// TestUpdateAgentTier tests dynamic tier mapping updates
// AC5: UpdateAgentTier() can dynamically modify agent mapping
func TestUpdateAgentTier(t *testing.T) {
	mockProviders := createMockProviders()

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: false,
	})

	// Update agent tier
	trinityClient.UpdateAgentTier("CustomAgent", TierRapid)

	// Verify the mapping was updated
	tier := GetTierForAgent("CustomAgent", trinityClient.router.agentTierMap)
	if tier != TierRapid {
		t.Errorf("Expected tier to be Rapid, got %s", tier.String())
	}

	// Send message using the updated mapping
	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}
	resp, err := trinityClient.SendMessage(ctx, "CustomAgent", messages)

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Should use Rapid tier
	expectedProvider := mockProviders[TierRapid]
	if resp.Content != expectedProvider.Name() {
		t.Errorf("Expected response from %s, got %s", expectedProvider.Name(), resp.Content)
	}
}

// TestResetMetrics_Client tests metrics reset
// AC5: ResetMetrics() works correctly
func TestResetMetrics_Client(t *testing.T) {
	mockProviders := createMockProviders()

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableMetrics: true,
	})

	// Send some requests
	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}
	_, _ = trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	// Verify metrics were recorded
	metrics := trinityClient.GetMetrics()
	if metrics.TotalRequests == 0 {
		t.Fatal("Expected requests to be recorded")
	}

	// Reset metrics
	trinityClient.ResetMetrics()

	// Verify metrics were reset
	metricsAfterReset := trinityClient.GetMetrics()
	if metricsAfterReset.TotalRequests != 0 {
		t.Errorf("Expected 0 requests after reset, got %d", metricsAfterReset.TotalRequests)
	}
}

// TestDefaultClientConfig tests default configuration
func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	if !config.EnableThinkingExtraction {
		t.Error("Expected EnableThinkingExtraction to be true by default")
	}

	if !config.EnableFallback {
		t.Error("Expected EnableFallback to be true by default")
	}

	if !config.EnableMetrics {
		t.Error("Expected EnableMetrics to be true by default")
	}

	if config.DefaultTimeout != 60*time.Second {
		t.Errorf("Expected DefaultTimeout to be 60s, got %v", config.DefaultTimeout)
	}
}

// Helper functions

func createTestRouterConfig() RouterConfig {
	return RouterConfig{
		ThinkingProvider: createTestProviderConfig("anthropic"),
		ReactiveProvider: createTestProviderConfig("anthropic"),
		RapidProvider:    createTestProviderConfig("anthropic"),
		FallbackEnabled:  false,
	}
}

func createTestProviderConfig(providerID string) ProviderTierConfig {
	return ProviderTierConfig{
		ProviderID: providerID,
		APIKey:     "test-key",
		Model:      "test-model",
		MaxTokens:  4096,
	}
}

func createMockProviders() map[TierLevel]*mockProvider {
	return map[TierLevel]*mockProvider{
		TierThinking: createSuccessProvider("thinking-provider"),
		TierReactive: createSuccessProvider("reactive-provider"),
		TierRapid:    createSuccessProvider("rapid-provider"),
	}
}

func createSuccessProvider(name string) *mockProvider {
	return &mockProvider{
		name: name,
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: name, // Return provider name as content for easy verification
			}, nil
		},
	}
}

func createMockProvidersWithDelay(delay time.Duration) map[TierLevel]*mockProvider {
	return map[TierLevel]*mockProvider{
		TierThinking: createDelayedProvider("thinking-provider", delay),
		TierReactive: createDelayedProvider("reactive-provider", delay),
		TierRapid:    createDelayedProvider("rapid-provider", delay),
	}
}

func createDelayedProvider(name string, delay time.Duration) *mockProvider {
	return &mockProvider{
		name: name,
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			select {
			case <-time.After(delay):
				return &client.Response{Content: name}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
}

// createMockClient creates a TrinityLLMClient with mock providers for testing
func createMockClient(providers map[TierLevel]*mockProvider, config TrinityClientConfig) *TrinityLLMClient {
	// Create router directly with mock providers
	router := &TrinityRouter{
		thinkingProvider:   providers[TierThinking],
		reactiveProvider:   providers[TierReactive],
		rapidProvider:      providers[TierRapid],
		agentTierMap:       make(map[string]TierLevel),
		fallbackEnabled:    config.EnableFallback,
		thinkingMiddleware: NewThinkingMiddleware(),
		metrics:            NewTrinityMetrics(1000),
	}

	// Copy default mapping
	for agent, tier := range DefaultAgentTierMapping {
		router.agentTierMap[agent] = tier
	}

	// Add fallback handler if enabled
	if config.EnableFallback {
		router.fallbackHandler = NewDefaultFallbackHandler()
	}

	return &TrinityLLMClient{
		router: router,
		config: config,
	}
}

func createMockRouterConfig(providers map[TierLevel]*mockProvider) RouterConfig {
	return RouterConfig{
		ThinkingProvider: ProviderTierConfig{ProviderID: "mock"},
		ReactiveProvider: ProviderTierConfig{ProviderID: "mock"},
		RapidProvider:    ProviderTierConfig{ProviderID: "mock"},
		FallbackEnabled:  false,
	}
}

// Helper removed - using existing contains() from router_test.go
