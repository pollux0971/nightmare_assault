package trinity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFallbackHandler(t *testing.T) {
	retryConfig := client.RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      1 * time.Second,
		BackoffMultiple: 2.0,
	}

	handler := NewFallbackHandler(retryConfig)
	require.NotNil(t, handler)
	assert.NotNil(t, handler.metrics)
	assert.Equal(t, 3, handler.retryConfig.MaxAttempts)
}

func TestNewDefaultFallbackHandler(t *testing.T) {
	handler := NewDefaultFallbackHandler()
	require.NotNil(t, handler)
	assert.NotNil(t, handler.metrics)
	assert.Equal(t, 3, handler.retryConfig.MaxAttempts) // Default is 3
}

func TestHandleFailure_NetworkError(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	// Thinking → Reactive
	nextTier, err := handler.HandleFailure(TierThinking, client.ErrNetworkError)
	require.NoError(t, err)
	assert.Equal(t, TierReactive, nextTier)
	assert.Equal(t, 1, handler.metrics.TotalFallbacks)
	assert.Equal(t, 1, handler.metrics.FallbacksByTier[TierThinking])

	// Reactive → Rapid
	nextTier, err = handler.HandleFailure(TierReactive, client.ErrNetworkError)
	require.NoError(t, err)
	assert.Equal(t, TierRapid, nextTier)
	assert.Equal(t, 2, handler.metrics.TotalFallbacks)
	assert.Equal(t, 1, handler.metrics.FallbacksByTier[TierReactive])

	// Rapid → Exhausted
	_, err = handler.HandleFailure(TierRapid, client.ErrNetworkError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all fallback tiers exhausted")
	assert.Equal(t, 1, handler.metrics.FullDegradationCount)
}

func TestHandleFailure_AuthError(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	authErr := client.NewAPIError("test", 401, "unauthorized", client.ErrInvalidAPIKey)

	// Auth errors should not trigger fallback
	nextTier, err := handler.HandleFailure(TierThinking, authErr)
	require.Error(t, err)
	assert.Equal(t, TierThinking, nextTier)
	assert.Equal(t, 0, handler.metrics.TotalFallbacks)
}

func TestHandleFailure_RateLimitError(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	rateLimitErr := client.NewAPIError("test", 429, "too many requests", client.ErrRateLimited)

	// Rate limit errors should trigger fallback
	nextTier, err := handler.HandleFailure(TierThinking, rateLimitErr)
	require.NoError(t, err)
	assert.Equal(t, TierReactive, nextTier)
	assert.Equal(t, 1, handler.metrics.TotalFallbacks)
}

func TestHandleFailure_ServiceUnavailable(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	serviceErr := client.NewAPIError("test", 503, "service unavailable", client.ErrServiceUnavailable)

	// Service unavailable should trigger fallback
	nextTier, err := handler.HandleFailure(TierThinking, serviceErr)
	require.NoError(t, err)
	assert.Equal(t, TierReactive, nextTier)
	assert.Equal(t, 1, handler.metrics.TotalFallbacks)
}

func TestHandleFailure_NoError(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	nextTier, err := handler.HandleFailure(TierThinking, nil)
	require.NoError(t, err)
	assert.Equal(t, TierThinking, nextTier)
	assert.Equal(t, 0, handler.metrics.TotalFallbacks)
}

func TestExecuteWithFallback_Success(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     1, // No retries for this test
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	thinkingProvider := &mockProvider{
		name: "thinking",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{Content: "test response from thinking"}, nil
		},
	}
	reactiveProvider := &mockProvider{
		name: "reactive",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{Content: "test response from reactive"}, nil
		},
	}
	rapidProvider := &mockProvider{
		name: "rapid",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{Content: "test response from rapid"}, nil
		},
	}

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		case TierRapid:
			return rapidProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierThinking,
		getProvider,
		messages,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, TierThinking, usedTier)
	assert.Equal(t, "test response from thinking", resp.Content)
	assert.Equal(t, 1, thinkingProvider.callCount)
	assert.Equal(t, 0, reactiveProvider.callCount)
	assert.Equal(t, 0, rapidProvider.callCount)
}

func TestExecuteWithFallback_DegradationChain(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     1, // No retries for simplicity
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	// Thinking fails, Reactive succeeds
	thinkingProvider := &mockProvider{
		name: "thinking",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, client.ErrNetworkError
		},
	}
	reactiveProvider := &mockProvider{
		name: "reactive",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{Content: "test response from reactive"}, nil
		},
	}
	rapidProvider := &mockProvider{
		name: "rapid",
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{Content: "test response from rapid"}, nil
		},
	}

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		case TierRapid:
			return rapidProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierThinking,
		getProvider,
		messages,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, TierReactive, usedTier) // Degraded to Reactive
	assert.Equal(t, "test response from reactive", resp.Content)
	assert.Equal(t, 1, thinkingProvider.callCount)
	assert.Equal(t, 1, reactiveProvider.callCount)
	assert.Equal(t, 0, rapidProvider.callCount)
	assert.Equal(t, 1, handler.metrics.TotalFallbacks)
}

func TestExecuteWithFallback_AllTiersFail(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     1, // No retries for simplicity
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	// All tiers fail
	thinkingProvider := createFailingMockProvider("thinking", client.ErrNetworkError)
	reactiveProvider := createFailingMockProvider("reactive", client.ErrNetworkError)
	rapidProvider := createFailingMockProvider("rapid", client.ErrNetworkError)

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		case TierRapid:
			return rapidProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierThinking,
		getProvider,
		messages,
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, TierRapid, usedTier) // Last tier tried
	assert.Contains(t, err.Error(), "all fallback tiers exhausted")
	assert.Equal(t, 1, thinkingProvider.callCount)
	assert.Equal(t, 1, reactiveProvider.callCount)
	assert.Equal(t, 1, rapidProvider.callCount)
	assert.Equal(t, 2, handler.metrics.TotalFallbacks) // Thinking→Reactive, Reactive→Rapid
	assert.Equal(t, 1, handler.metrics.FullDegradationCount)
}

func TestExecuteWithFallback_StartFromReactive(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     1,
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	// Start from Reactive, should degrade to Rapid if fails
	thinkingProvider := createSuccessMockProvider("thinking")
	reactiveProvider := createFailingMockProvider("reactive", client.ErrNetworkError)
	rapidProvider := createSuccessMockProvider("rapid")

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		case TierRapid:
			return rapidProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierReactive, // Start from Reactive
		getProvider,
		messages,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, TierRapid, usedTier)
	assert.Equal(t, "test response from rapid", resp.Content)
	assert.Equal(t, 0, thinkingProvider.callCount) // Never tried Thinking
	assert.Equal(t, 1, reactiveProvider.callCount)
	assert.Equal(t, 1, rapidProvider.callCount)
}

func TestExecuteWithFallback_NonRetryableError(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     1,
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	// Auth error should not trigger fallback
	thinkingProvider := createFailingMockProvider("thinking", client.NewAPIError("test", 401, "unauthorized", client.ErrInvalidAPIKey))
	reactiveProvider := createSuccessMockProvider("reactive")
	rapidProvider := createSuccessMockProvider("rapid")

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		case TierRapid:
			return rapidProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierThinking,
		getProvider,
		messages,
	)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, TierThinking, usedTier) // Did not degrade
	assert.Equal(t, 1, thinkingProvider.callCount)
	assert.Equal(t, 0, reactiveProvider.callCount) // Never tried
	assert.Equal(t, 0, rapidProvider.callCount)
	assert.Equal(t, 0, handler.metrics.TotalFallbacks) // No fallback attempted
}

func TestExecuteWithFallback_ContextCancellation(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      1 * time.Second,
		BackoffMultiple: 2.0,
	})

	thinkingProvider := createFailingMockProvider("thinking", client.ErrNetworkError)

	getProvider := func(tier TierLevel) client.Provider {
		return thinkingProvider
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messages := []client.Message{{Role: "user", Content: "test"}}

	_, _, err := handler.ExecuteWithFallback(ctx, TierThinking, getProvider, messages)

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestGetNextTier(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	tests := []struct {
		name        string
		tier        TierLevel
		wantNext    TierLevel
		wantCanDegrade bool
	}{
		{
			name:        "Thinking to Reactive",
			tier:        TierThinking,
			wantNext:    TierReactive,
			wantCanDegrade: true,
		},
		{
			name:        "Reactive to Rapid",
			tier:        TierReactive,
			wantNext:    TierRapid,
			wantCanDegrade: true,
		},
		{
			name:        "Rapid exhausted",
			tier:        TierRapid,
			wantNext:    TierRapid,
			wantCanDegrade: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTier, canDegrade := handler.getNextTier(tt.tier)
			assert.Equal(t, tt.wantNext, nextTier)
			assert.Equal(t, tt.wantCanDegrade, canDegrade)
		})
	}
}

func TestShouldFallback(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	tests := []struct {
		name           string
		err            error
		shouldFallback bool
	}{
		{
			name:           "nil error",
			err:            nil,
			shouldFallback: false,
		},
		{
			name:           "network error",
			err:            client.ErrNetworkError,
			shouldFallback: true,
		},
		{
			name:           "rate limit error",
			err:            client.ErrRateLimited,
			shouldFallback: true,
		},
		{
			name:           "service unavailable",
			err:            client.ErrServiceUnavailable,
			shouldFallback: true,
		},
		{
			name:           "context deadline exceeded",
			err:            context.DeadlineExceeded,
			shouldFallback: true,
		},
		{
			name:           "auth error 401",
			err:            client.NewAPIError("test", 401, "unauthorized", client.ErrInvalidAPIKey),
			shouldFallback: false,
		},
		{
			name:           "auth error 403",
			err:            client.NewAPIError("test", 403, "forbidden", client.ErrInvalidAPIKey),
			shouldFallback: false,
		},
		{
			name:           "bad request 400",
			err:            client.NewAPIError("test", 400, "bad request", errors.New("invalid")),
			shouldFallback: false,
		},
		{
			name:           "rate limit 429",
			err:            client.NewAPIError("test", 429, "too many requests", client.ErrRateLimited),
			shouldFallback: true,
		},
		{
			name:           "server error 500",
			err:            client.NewAPIError("test", 500, "internal error", errors.New("server")),
			shouldFallback: true,
		},
		{
			name:           "server error 503",
			err:            client.NewAPIError("test", 503, "unavailable", client.ErrServiceUnavailable),
			shouldFallback: true,
		},
		{
			name:           "retry error",
			err:            &client.RetryError{Attempts: 3, LastErr: client.ErrNetworkError},
			shouldFallback: true,
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown"),
			shouldFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.shouldFallback(tt.err)
			assert.Equal(t, tt.shouldFallback, result)
		})
	}
}

func TestFallbackHandler_GetMetrics(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	// Initial metrics should be zero
	metrics := handler.GetMetrics()
	assert.Equal(t, 0, metrics.TotalFallbacks)
	assert.Equal(t, 0, metrics.TotalRetries)
	assert.Equal(t, 0, metrics.FullDegradationCount)

	// Trigger some fallbacks
	handler.HandleFailure(TierThinking, client.ErrNetworkError)
	handler.HandleFailure(TierReactive, client.ErrNetworkError)

	metrics = handler.GetMetrics()
	assert.Equal(t, 2, metrics.TotalFallbacks)
	assert.Equal(t, 1, metrics.FallbacksByTier[TierThinking])
	assert.Equal(t, 1, metrics.FallbacksByTier[TierReactive])
}

func TestResetMetrics(t *testing.T) {
	handler := NewDefaultFallbackHandler()

	// Trigger some fallbacks
	handler.HandleFailure(TierThinking, client.ErrNetworkError)
	handler.HandleFailure(TierReactive, client.ErrNetworkError)

	metrics := handler.GetMetrics()
	assert.Equal(t, 2, metrics.TotalFallbacks)

	// Reset metrics
	handler.ResetMetrics()

	metrics = handler.GetMetrics()
	assert.Equal(t, 0, metrics.TotalFallbacks)
	assert.Equal(t, 0, metrics.TotalRetries)
	assert.Equal(t, 0, metrics.FullDegradationCount)
	assert.Equal(t, 0, len(metrics.FallbacksByTier))
}

func TestExecuteWithFallback_Retries(t *testing.T) {
	handler := NewFallbackHandler(client.RetryConfig{
		MaxAttempts:     3, // Allow retries
		InitialBackoff:  1 * time.Millisecond,
		MaxBackoff:      10 * time.Millisecond,
		BackoffMultiple: 2.0,
	})

	thinkingProvider := createFailingMockProvider("thinking", client.ErrNetworkError)
	reactiveProvider := createSuccessMockProvider("reactive")

	getProvider := func(tier TierLevel) client.Provider {
		switch tier {
		case TierThinking:
			return thinkingProvider
		case TierReactive:
			return reactiveProvider
		default:
			return reactiveProvider
		}
	}

	messages := []client.Message{{Role: "user", Content: "test"}}

	resp, usedTier, err := handler.ExecuteWithFallback(
		context.Background(),
		TierThinking,
		getProvider,
		messages,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, TierReactive, usedTier)
	// Thinking should have been retried 3 times before degrading
	assert.Equal(t, 3, thinkingProvider.callCount)
	assert.Equal(t, 1, reactiveProvider.callCount)
	assert.True(t, handler.metrics.TotalRetries >= 3)
}
