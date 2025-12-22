package trinity

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_FallbackIntegration_Success(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)
	require.NotNil(t, router.fallbackHandler)

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, err := router.Route(context.Background(), "JudgeAgent", messages)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "response") // Mock returns "mock response from thinking"
}

func TestRouter_FallbackIntegration_DegradationChain(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)

	// Make Thinking tier fail
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	thinkingProvider.shouldFail = true
	thinkingProvider.failureError = client.ErrNetworkError

	messages := []client.Message{{Role: "user", Content: "test"}}

	// Should degrade to Reactive and succeed
	resp, err := router.Route(context.Background(), "JudgeAgent", messages)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "reactive")

	// Check metrics
	metrics := router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.TotalFallbacks)
	assert.Equal(t, 1, metrics.FallbacksByTier[TierThinking])
}

func TestRouter_FallbackIntegration_AllTiersFail(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)

	// Make all tiers fail
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	thinkingProvider.shouldFail = true
	thinkingProvider.failureError = client.ErrNetworkError

	reactiveProvider := router.reactiveProvider.(*mockProvider)
	reactiveProvider.shouldFail = true
	reactiveProvider.failureError = client.ErrNetworkError

	rapidProvider := router.rapidProvider.(*mockProvider)
	rapidProvider.shouldFail = true
	rapidProvider.failureError = client.ErrNetworkError

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, err := router.Route(context.Background(), "JudgeAgent", messages)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "all tiers failed")

	// Check metrics
	metrics := router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 2, metrics.TotalFallbacks) // Thinking→Reactive, Reactive→Rapid
	assert.Equal(t, 1, metrics.FullDegradationCount)
}

func TestRouter_FallbackIntegration_DisabledFallback(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: false, // Fallback disabled
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)
	assert.Nil(t, router.fallbackHandler)

	// Make Thinking tier fail
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	thinkingProvider.shouldFail = true
	thinkingProvider.failureError = client.ErrNetworkError

	messages := []client.Message{{Role: "user", Content: "test"}}

	// Should fail without fallback
	resp, err := router.Route(context.Background(), "JudgeAgent", messages)
	require.Error(t, err)
	assert.Nil(t, resp)

	// Reactive should not have been called
	reactiveProvider := router.reactiveProvider.(*mockProvider)
	assert.Equal(t, 0, reactiveProvider.callCount)

	// No metrics should be available
	metrics := router.GetFallbackMetrics()
	assert.Nil(t, metrics)
}

func TestRouter_FallbackIntegration_StartFromReactive(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)

	// Make Reactive tier fail
	reactiveProvider := router.reactiveProvider.(*mockProvider)
	reactiveProvider.shouldFail = true
	reactiveProvider.failureError = client.ErrNetworkError

	messages := []client.Message{{Role: "user", Content: "test"}}

	// NarrationAgent starts at Reactive tier
	resp, err := router.Route(context.Background(), "NarrationAgent", messages)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "rapid")

	// Thinking should not have been tried
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	assert.Equal(t, 0, thinkingProvider.callCount)

	// Check metrics
	metrics := router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.TotalFallbacks)
	assert.Equal(t, 1, metrics.FallbacksByTier[TierReactive])
}

func TestRouter_FallbackIntegration_NonRetryableError(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)

	// Make Thinking tier fail with auth error (non-retryable)
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	thinkingProvider.shouldFail = true
	thinkingProvider.failureError = client.NewAPIError("test", 401, "unauthorized", client.ErrInvalidAPIKey)

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, err := router.Route(context.Background(), "JudgeAgent", messages)
	require.Error(t, err)
	assert.Nil(t, resp)

	// Should not have degraded to Reactive
	reactiveProvider := router.reactiveProvider.(*mockProvider)
	assert.Equal(t, 0, reactiveProvider.callCount)

	// No fallbacks should have occurred
	metrics := router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalFallbacks)
}

func TestRouter_FallbackIntegration_MetricsReset(t *testing.T) {
	cfg := RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID: "test-thinking",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID: "test-reactive",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID: "test-rapid",
			Model:      "test-model",
			MaxTokens:  1000,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:     1,
			InitialBackoff:  1 * time.Millisecond,
			MaxBackoff:      10 * time.Millisecond,
			BackoffMultiple: 2.0,
		},
	}

	router, err := createTestRouter(cfg)
	require.NoError(t, err)

	// Make Thinking tier fail
	thinkingProvider := router.thinkingProvider.(*mockProvider)
	thinkingProvider.shouldFail = true
	thinkingProvider.failureError = client.ErrNetworkError

	messages := []client.Message{{Role: "user", Content: "test"}}

	// Trigger fallback
	_, err = router.Route(context.Background(), "JudgeAgent", messages)
	require.NoError(t, err)

	metrics := router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.TotalFallbacks)

	// Reset metrics
	router.ResetFallbackMetrics()

	metrics = router.GetFallbackMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalFallbacks)
}

// Helper function to create a test router with mock providers
func createTestRouter(cfg RouterConfig) (*TrinityRouter, error) {
	router := &TrinityRouter{
		thinkingProvider: &mockProvider{name: "thinking"},
		reactiveProvider: &mockProvider{name: "reactive"},
		rapidProvider:    &mockProvider{name: "rapid"},
		agentTierMap:     make(map[string]TierLevel),
		fallbackEnabled:  cfg.FallbackEnabled,
	}

	// Copy default agent tier mapping
	for agent, tier := range DefaultAgentTierMapping {
		router.agentTierMap[agent] = tier
	}

	// Apply overrides
	for agent, tier := range cfg.AgentTierOverrides {
		router.agentTierMap[agent] = tier
	}

	// Create fallback handler if enabled
	if cfg.FallbackEnabled {
		retryConfig := cfg.RetryConfig
		if retryConfig.MaxAttempts == 0 {
			retryConfig = client.RetryConfig{
				MaxAttempts:     3,
				InitialBackoff:  100 * time.Millisecond,
				MaxBackoff:      1 * time.Second,
				BackoffMultiple: 2.0,
			}
		}
		router.fallbackHandler = NewFallbackHandler(retryConfig)
	}

	return router, nil
}
