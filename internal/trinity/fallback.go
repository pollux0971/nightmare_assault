package trinity

import (
	"context"
	"errors"
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// FallbackHandler handles provider failures and tier degradation
type FallbackHandler struct {
	retryConfig client.RetryConfig
	metrics     *FallbackMetrics
}

// FallbackMetrics tracks fallback statistics
type FallbackMetrics struct {
	TotalFallbacks      int
	FallbacksByTier     map[TierLevel]int
	TotalRetries        int
	FullDegradationCount int
}

// NewFallbackHandler creates a new FallbackHandler
func NewFallbackHandler(retryConfig client.RetryConfig) *FallbackHandler {
	return &FallbackHandler{
		retryConfig: retryConfig,
		metrics: &FallbackMetrics{
			FallbacksByTier: make(map[TierLevel]int),
		},
	}
}

// NewDefaultFallbackHandler creates a FallbackHandler with default retry configuration
func NewDefaultFallbackHandler() *FallbackHandler {
	return NewFallbackHandler(client.DefaultRetryConfig())
}

// HandleFailure attempts to handle a provider failure by degrading to a lower tier
// Returns the next tier to try, or an error if all tiers have been exhausted
func (h *FallbackHandler) HandleFailure(tier TierLevel, err error) (TierLevel, error) {
	if err == nil {
		return tier, nil
	}

	logger.Debug("Handling provider failure", map[string]interface{}{
		"tier":  tier.String(),
		"error": err.Error(),
	})

	// Check if error is non-retryable (e.g., auth errors)
	if !h.shouldFallback(err) {
		logger.Debug("Error is not retryable, returning immediately", map[string]interface{}{
			"tier":  tier.String(),
			"error": err.Error(),
		})
		return tier, err
	}

	// Determine next tier in degradation chain
	nextTier, canDegrade := h.getNextTier(tier)
	if !canDegrade {
		// All tiers exhausted
		h.metrics.FullDegradationCount++
		logger.Debug("All tiers exhausted", map[string]interface{}{
			"original_tier": tier.String(),
			"error":         err.Error(),
		})
		return tier, fmt.Errorf("all fallback tiers exhausted: %w", err)
	}

	// Record fallback metrics
	h.metrics.TotalFallbacks++
	h.metrics.FallbacksByTier[tier]++

	logger.Debug("Degrading to lower tier", map[string]interface{}{
		"from_tier": tier.String(),
		"to_tier":   nextTier.String(),
	})

	return nextTier, nil
}

// ExecuteWithFallback executes a provider call with automatic retry and fallback
func (h *FallbackHandler) ExecuteWithFallback(
	ctx context.Context,
	initialTier TierLevel,
	getProvider func(TierLevel) client.Provider,
	messages []client.Message,
) (*client.Response, TierLevel, error) {
	currentTier := initialTier

	// Try each tier in degradation chain
	for {
		provider := getProvider(currentTier)

		logger.Debug("Attempting tier", map[string]interface{}{
			"tier": currentTier.String(),
		})

		// Execute with retry for this tier
		var resp *client.Response
		err := client.WithRetry(ctx, h.retryConfig, func(ctx context.Context) error {
			h.metrics.TotalRetries++
			var err error
			resp, err = provider.SendMessage(ctx, messages)
			return err
		})

		if err == nil {
			// Success!
			logger.Debug("Tier succeeded", map[string]interface{}{
				"tier":          currentTier.String(),
				"initial_tier":  initialTier.String(),
				"degraded":      currentTier != initialTier,
			})
			return resp, currentTier, nil
		}

		logger.Debug("Tier failed after retries", map[string]interface{}{
			"tier":  currentTier.String(),
			"error": err.Error(),
		})

		// Try to degrade to next tier
		nextTier, fallbackErr := h.HandleFailure(currentTier, err)
		if fallbackErr != nil {
			// Cannot degrade further
			return nil, currentTier, fallbackErr
		}

		currentTier = nextTier
	}
}

// getNextTier returns the next lower tier in the degradation chain
// Degradation order: Thinking → Reactive → Rapid → (exhausted)
func (h *FallbackHandler) getNextTier(tier TierLevel) (TierLevel, bool) {
	switch tier {
	case TierThinking:
		return TierReactive, true
	case TierReactive:
		return TierRapid, true
	case TierRapid:
		return TierRapid, false // No more tiers
	default:
		return TierReactive, false
	}
}

// shouldFallback determines if an error should trigger fallback
func (h *FallbackHandler) shouldFallback(err error) bool {
	if err == nil {
		return false
	}

	// Check for APIError
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		// Don't fallback on auth errors (401, 403) - these won't be fixed by trying another tier
		if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
			return false
		}

		// Don't fallback on client errors (4xx except rate limit)
		if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
			return false
		}

		// Fallback on network errors, rate limits, server errors
		if errors.Is(apiErr.Err, client.ErrNetworkError) ||
			errors.Is(apiErr.Err, client.ErrRateLimited) ||
			errors.Is(apiErr.Err, client.ErrServiceUnavailable) ||
			apiErr.StatusCode >= 500 {
			return true
		}
	}

	// Check for RetryError (already retried and failed)
	var retryErr *client.RetryError
	if errors.As(err, &retryErr) {
		// If retries are exhausted, should fallback
		return true
	}

	// Fallback on common retryable errors
	if errors.Is(err, client.ErrNetworkError) ||
		errors.Is(err, client.ErrRateLimited) ||
		errors.Is(err, client.ErrServiceUnavailable) ||
		errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Default: don't fallback on unknown errors
	return false
}

// GetMetrics returns current fallback metrics
func (h *FallbackHandler) GetMetrics() FallbackMetrics {
	return *h.metrics
}

// ResetMetrics resets all fallback metrics
func (h *FallbackHandler) ResetMetrics() {
	h.metrics = &FallbackMetrics{
		FallbacksByTier: make(map[TierLevel]int),
	}
}
