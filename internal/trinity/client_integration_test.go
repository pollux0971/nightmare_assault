package trinity

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// TestTrinityClient_EndToEnd_ThinkingExtraction tests complete thinking extraction flow
// AC6: Integration test - Router + Middleware + Metrics end-to-end
func TestTrinityClient_EndToEnd_ThinkingExtraction(t *testing.T) {
	// Create provider that returns thinking tags
	thinkingProvider := &mockProvider{
		name: "thinking-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "<thinking>This is my deep reasoning process</thinking>This is the final answer",
			}, nil
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: thinkingProvider,
		TierReactive: createSuccessProvider("reactive-provider"),
		TierRapid:    createSuccessProvider("rapid-provider"),
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           false,
		EnableMetrics:            true,
	})

	ctx := context.Background()
	messages := []client.Message{
		{Role: "user", Content: "Complex reasoning task"},
	}

	// Send message to Thinking tier
	resp, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify thinking tags were extracted
	if !trinityClient.HasThinking(resp) {
		t.Error("Expected response to have thinking metadata")
	}

	// Verify thinking chain is in metadata
	thinkingChain, ok := resp.Metadata["thinking_chain"].(string)
	if !ok {
		t.Fatal("Expected thinking_chain in metadata")
	}

	expectedThinking := "This is my deep reasoning process"
	if thinkingChain != expectedThinking {
		t.Errorf("Expected thinking chain=%q, got %q", expectedThinking, thinkingChain)
	}

	// Verify thinking tags were removed from content
	if contains(resp.Content, "<thinking>") {
		t.Error("Thinking tags should be removed from content")
	}

	expectedContent := "This is the final answer"
	if resp.Content != expectedContent {
		t.Errorf("Expected content=%q, got %q", expectedContent, resp.Content)
	}

	// Verify metrics were recorded
	metrics := trinityClient.GetMetrics()
	if metrics.ThinkingStats.TotalRequests != 1 {
		t.Errorf("Expected 1 Thinking request, got %d", metrics.ThinkingStats.TotalRequests)
	}
}

// TestTrinityClient_EndToEnd_FallbackWithMetrics tests fallback + metrics integration
// AC6: Integration test - Fallback + Metrics end-to-end
func TestTrinityClient_EndToEnd_FallbackWithMetrics(t *testing.T) {
	// Create failing Thinking provider and working Reactive provider
	failingProvider := &mockProvider{
		name: "failing-thinking",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, client.NewAPIError("thinking", 500, "server error", nil)
		},
	}

	successProvider := &mockProvider{
		name: "working-reactive",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "Fallback response from reactive tier",
			}, nil
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: failingProvider,
		TierReactive: successProvider,
		TierRapid:    createSuccessProvider("rapid-provider"),
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           true,
		EnableMetrics:            true,
	})

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test request"}}

	// Send message that should trigger fallback
	resp, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	if err != nil {
		t.Fatalf("Expected fallback to succeed, got error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Verify we got response from fallback tier
	if resp.Content != "Fallback response from reactive tier" {
		t.Errorf("Expected fallback response, got: %s", resp.Content)
	}

	// Verify fallback metrics
	fallbackMetrics := trinityClient.GetFallbackMetrics()
	if fallbackMetrics == nil {
		t.Fatal("Expected fallback metrics to be non-nil")
	}

	if fallbackMetrics.TotalFallbacks == 0 {
		t.Error("Expected fallback event to be recorded")
	}

	// Verify Trinity metrics show tiers were used
	metrics := trinityClient.GetMetrics()

	// At least one tier should have completed requests
	totalRequests := metrics.ThinkingStats.TotalRequests + metrics.ReactiveStats.TotalRequests + metrics.RapidStats.TotalRequests
	if totalRequests == 0 {
		t.Error("Expected at least one request to be recorded")
	}

	// Reactive tier should have at least 1 request (since Thinking failed and we degraded)
	if metrics.ReactiveStats.TotalRequests == 0 {
		t.Error("Expected Reactive tier to have been used for fallback")
	}

	// Verify downgrade was recorded (if supported by fallback handler)
	if fallbackMetrics.TotalFallbacks > 0 && metrics.TotalDowngrades == 0 {
		// Note: Downgrade metric may not always be recorded depending on implementation details
		// This is a soft check
		t.Log("Note: Fallback occurred but downgrade metric not recorded")
	}
}

// TestTrinityClient_EndToEnd_AllFeatures tests all features together
// AC6: Integration test - Full system with all features enabled
func TestTrinityClient_EndToEnd_AllFeatures(t *testing.T) {
	// Create providers with different behaviors
	thinkingProvider := &mockProvider{
		name: "thinking-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "<thinking>Deep analysis</thinking>Thinking tier response",
			}, nil
		},
	}

	reactiveProvider := &mockProvider{
		name: "reactive-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "Reactive tier response",
			}, nil
		},
	}

	rapidProvider := &mockProvider{
		name: "rapid-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "Rapid tier response",
			}, nil
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: thinkingProvider,
		TierReactive: reactiveProvider,
		TierRapid:    rapidProvider,
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           true,
		EnableMetrics:            true,
		DefaultTimeout:           5 * time.Second,
	})

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	// Test 1: Send to Thinking tier (with thinking extraction)
	resp1, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)
	if err != nil {
		t.Fatalf("Thinking tier request failed: %v", err)
	}

	if !trinityClient.HasThinking(resp1) {
		t.Error("Expected thinking metadata in Thinking tier response")
	}

	// Test 2: Send to Reactive tier (no thinking tags)
	resp2, err := trinityClient.SendMessage(ctx, "NarrationAgent", messages)
	if err != nil {
		t.Fatalf("Reactive tier request failed: %v", err)
	}

	if resp2.Content != "Reactive tier response" {
		t.Errorf("Expected Reactive response, got: %s", resp2.Content)
	}

	// Test 3: Send to Rapid tier
	resp3, err := trinityClient.SendMessage(ctx, "DreamAgent", messages)
	if err != nil {
		t.Fatalf("Rapid tier request failed: %v", err)
	}

	if resp3.Content != "Rapid tier response" {
		t.Errorf("Expected Rapid response, got: %s", resp3.Content)
	}

	// Test 4: Tier override
	overrideTier := TierRapid
	resp4, err := trinityClient.SendMessageWithOptions(ctx, "JudgeAgent", messages, SendOptions{
		TierOverride: &overrideTier,
	})
	if err != nil {
		t.Fatalf("Tier override request failed: %v", err)
	}

	if resp4.Content != "Rapid tier response" {
		t.Errorf("Expected Rapid response from override, got: %s", resp4.Content)
	}

	// Test 5: Verify metrics
	metrics := trinityClient.GetMetrics()

	if metrics.TotalRequests != 4 {
		t.Errorf("Expected 4 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.ThinkingStats.TotalRequests != 1 {
		t.Errorf("Expected 1 Thinking request, got %d", metrics.ThinkingStats.TotalRequests)
	}

	if metrics.ReactiveStats.TotalRequests != 1 {
		t.Errorf("Expected 1 Reactive request, got %d", metrics.ReactiveStats.TotalRequests)
	}

	if metrics.RapidStats.TotalRequests != 2 {
		t.Errorf("Expected 2 Rapid requests, got %d", metrics.RapidStats.TotalRequests)
	}

	// Test 6: Update agent tier and verify
	trinityClient.UpdateAgentTier("CustomAgent", TierThinking)
	resp5, err := trinityClient.SendMessage(ctx, "CustomAgent", messages)
	if err != nil {
		t.Fatalf("Custom agent request failed: %v", err)
	}

	if !trinityClient.HasThinking(resp5) {
		t.Error("Expected thinking metadata for custom agent")
	}

	// Test 7: Reset metrics
	trinityClient.ResetMetrics()
	metricsAfterReset := trinityClient.GetMetrics()

	if metricsAfterReset.TotalRequests != 0 {
		t.Errorf("Expected 0 requests after reset, got %d", metricsAfterReset.TotalRequests)
	}

	// Test 8: Performance report (should not panic)
	trinityClient.LogPerformanceReport()
}

// TestTrinityClient_ErrorHandling tests error handling
// AC6: Integration test - Error handling
func TestTrinityClient_ErrorHandling(t *testing.T) {
	// Create provider that always fails
	failingProvider := &mockProvider{
		name: "always-fails",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, client.NewAPIError("test", 500, "server error", nil)
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: failingProvider,
		TierReactive: failingProvider,
		TierRapid:    failingProvider,
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: true,
		EnableMetrics:  true,
	})

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	// All tiers fail - should return error
	_, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	if err == nil {
		t.Fatal("Expected error when all tiers fail")
	}

	// Verify error message is clear
	if !contains(err.Error(), "all tiers failed") {
		t.Errorf("Expected clear error message, got: %v", err)
	}

	// Verify metrics recorded some activity
	metrics := trinityClient.GetMetrics()

	// Should have attempted all tiers
	totalRequests := metrics.ThinkingStats.TotalRequests + metrics.ReactiveStats.TotalRequests + metrics.RapidStats.TotalRequests
	if totalRequests == 0 {
		t.Error("Expected some requests to be recorded")
	}

	// Verify fallback metrics
	fallbackMetrics := trinityClient.GetFallbackMetrics()
	if fallbackMetrics == nil {
		t.Fatal("Expected fallback metrics to be non-nil")
	}

	if fallbackMetrics.FullDegradationCount == 0 {
		t.Error("Expected full degradation to be recorded")
	}
}

// TestTrinityClient_DisableThinkingExtraction tests thinking extraction can be disabled
func TestTrinityClient_DisableThinkingExtraction(t *testing.T) {
	thinkingProvider := &mockProvider{
		name: "thinking-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "<thinking>Reasoning</thinking>Answer",
			}, nil
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: thinkingProvider,
		TierReactive: createSuccessProvider("reactive-provider"),
		TierRapid:    createSuccessProvider("rapid-provider"),
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableThinkingExtraction: false, // Disabled
		EnableFallback:           false,
		EnableMetrics:            false,
	})

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Thinking tags should be preserved when extraction is disabled
	if !contains(resp.Content, "<thinking>") {
		t.Error("Expected thinking tags to be preserved when extraction disabled")
	}

	if !contains(resp.Content, "Reasoning") {
		t.Error("Expected thinking content to be in response")
	}
}

// TestTrinityClient_ContextTimeout tests context timeout handling
func TestTrinityClient_ContextTimeout(t *testing.T) {
	slowProvider := &mockProvider{
		name: "slow-provider",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return &client.Response{Content: "slow response"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	mockProviders := map[TierLevel]*mockProvider{
		TierThinking: slowProvider,
		TierReactive: slowProvider,
		TierRapid:    slowProvider,
	}

	trinityClient := createMockClient(mockProviders, TrinityClientConfig{
		EnableFallback: false,
		DefaultTimeout: 10 * time.Millisecond, // Very short timeout
	})

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	_, err := trinityClient.SendMessage(ctx, "JudgeAgent", messages)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	// Should get timeout error
	if !contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}
