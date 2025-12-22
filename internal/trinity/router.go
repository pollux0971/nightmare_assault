package trinity

import (
	"context"
	"fmt"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// TrinityRouter routes agent requests to appropriate model tiers
type TrinityRouter struct {
	thinkingProvider client.Provider
	reactiveProvider client.Provider
	rapidProvider    client.Provider

	agentTierMap       map[string]TierLevel
	fallbackEnabled    bool
	fallbackHandler    *FallbackHandler
	thinkingMiddleware *ThinkingMiddleware
	metrics            *TrinityMetrics // Story 9-5: Metrics collection
}

// NewTrinityRouter creates a new TrinityRouter instance
func NewTrinityRouter(cfg RouterConfig) (*TrinityRouter, error) {
	// Create providers for each tier using CreateProvider() method
	thinkingProvider, err := cfg.ThinkingProvider.CreateProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create Thinking provider: %w", err)
	}

	reactiveProvider, err := cfg.ReactiveProvider.CreateProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create Reactive provider: %w", err)
	}

	rapidProvider, err := cfg.RapidProvider.CreateProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create Rapid provider: %w", err)
	}

	// Merge default mapping with user overrides
	agentTierMap := make(map[string]TierLevel)
	for agent, tier := range DefaultAgentTierMapping {
		agentTierMap[agent] = tier
	}
	for agent, tier := range cfg.AgentTierOverrides {
		agentTierMap[agent] = tier
	}

	// Create fallback handler with retry configuration
	var fallbackHandler *FallbackHandler
	if cfg.FallbackEnabled {
		retryConfig := cfg.RetryConfig
		if retryConfig.MaxAttempts == 0 {
			retryConfig = client.DefaultRetryConfig()
		}
		fallbackHandler = NewFallbackHandler(retryConfig)
	}

	// Story 9-2: Initialize ThinkingMiddleware for automatic thinking tag processing
	thinkingMiddleware := NewThinkingMiddleware()

	// Story 9-5: Initialize TrinityMetrics for metrics collection
	metrics := NewTrinityMetrics(1000)

	return &TrinityRouter{
		thinkingProvider:   thinkingProvider,
		reactiveProvider:   reactiveProvider,
		rapidProvider:      rapidProvider,
		agentTierMap:       agentTierMap,
		fallbackEnabled:    cfg.FallbackEnabled,
		fallbackHandler:    fallbackHandler,
		thinkingMiddleware: thinkingMiddleware,
		metrics:            metrics,
	}, nil
}

// Route routes an agent request to the appropriate tier
// Story 9-5 AC3: Integrate with TrinityRouter - record metrics for each request
func (r *TrinityRouter) Route(ctx context.Context, agentName string, messages []client.Message) (*client.Response, error) {
	// Determine tier for this agent
	tier := GetTierForAgent(agentName, r.agentTierMap)

	logger.Debug("Trinity routing", map[string]interface{}{
		"agent": agentName,
		"tier":  tier.String(),
	})

	// Story 9-5: Record start time for metrics
	startTime := time.Now()

	// If fallback is enabled, use fallback handler
	if r.fallbackEnabled && r.fallbackHandler != nil {
		resp, usedTier, err := r.fallbackHandler.ExecuteWithFallback(
			ctx,
			tier,
			r.getProviderForTier,
			messages,
		)

		// Story 9-5: Record metrics for the used tier
		duration := time.Since(startTime)
		if r.metrics != nil {
			r.metrics.RecordRequest(usedTier, duration, err)
		}

		if err != nil {
			logger.Debug("All fallback tiers failed", map[string]interface{}{
				"agent":        agentName,
				"initial_tier": tier.String(),
				"final_tier":   usedTier.String(),
				"error":        err.Error(),
			})
			return nil, fmt.Errorf("all tiers failed for agent %s: %w", agentName, err)
		}

		// Story 9-5: Record tier transition if degraded
		if usedTier != tier && r.metrics != nil {
			r.metrics.RecordDowngrade(tier, usedTier)
		}

		// Log if tier was degraded
		if usedTier != tier {
			logger.Debug("Request completed with degraded tier", map[string]interface{}{
				"agent":        agentName,
				"initial_tier": tier.String(),
				"used_tier":    usedTier.String(),
			})
		}

		// Story 9-2: Process thinking tags if using Thinking tier
		if usedTier == TierThinking && r.thinkingMiddleware != nil {
			resp = r.processThinkingTags(resp)
		}

		return resp, nil
	}

	// Fallback disabled - use simple single-tier attempt
	provider := r.getProviderForTier(tier)
	resp, err := provider.SendMessage(ctx, messages)

	// Story 9-5: Record metrics for the request
	duration := time.Since(startTime)
	if r.metrics != nil {
		r.metrics.RecordRequest(tier, duration, err)
	}

	if err != nil {
		logger.Debug("Provider failed", map[string]interface{}{
			"agent": agentName,
			"tier":  tier.String(),
			"error": err.Error(),
		})
		return nil, fmt.Errorf("tier %s failed for agent %s: %w", tier.String(), agentName, err)
	}

	// Story 9-2: Process thinking tags if using Thinking tier
	if tier == TierThinking && r.thinkingMiddleware != nil {
		resp = r.processThinkingTags(resp)
	}

	return resp, nil
}

// getProviderForTier returns the provider for a given tier
func (r *TrinityRouter) getProviderForTier(tier TierLevel) client.Provider {
	switch tier {
	case TierThinking:
		return r.thinkingProvider
	case TierReactive:
		return r.reactiveProvider
	case TierRapid:
		return r.rapidProvider
	default:
		// Default to Reactive if unknown
		return r.reactiveProvider
	}
}

// GetFallbackMetrics returns current fallback metrics if fallback is enabled
func (r *TrinityRouter) GetFallbackMetrics() *FallbackMetrics {
	if r.fallbackHandler != nil {
		metrics := r.fallbackHandler.GetMetrics()
		return &metrics
	}
	return nil
}

// ResetFallbackMetrics resets fallback metrics if fallback is enabled
func (r *TrinityRouter) ResetFallbackMetrics() {
	if r.fallbackHandler != nil {
		r.fallbackHandler.ResetMetrics()
	}
}

// processThinkingTags applies ThinkingMiddleware processing to a response
// Story 9-2: Extract and remove thinking tags from Thinking-tier responses
func (r *TrinityRouter) processThinkingTags(resp *client.Response) *client.Response {
	if resp == nil {
		return resp
	}

	// Extract thinking chain
	thinkingChain := r.thinkingMiddleware.extractThinking(resp.Content)

	// If thinking tags were found
	if thinkingChain != "" {
		// Store thinking chain in metadata
		if resp.Metadata == nil {
			resp.Metadata = make(map[string]interface{})
		}
		resp.Metadata["thinking_chain"] = thinkingChain

		// Remove thinking tags from content
		resp.Content = r.thinkingMiddleware.removeThinkingTags(resp.Content)

		logger.Debug("ThinkingMiddleware: Processed thinking tags", map[string]interface{}{
			"chain_length": len(thinkingChain),
		})
	}

	return resp
}

// GetMetrics returns the Trinity metrics instance
// Story 9-5: Provide access to metrics
func (r *TrinityRouter) GetMetrics() *TrinityMetrics {
	return r.metrics
}

// GetMetricsSummary returns a summary of all collected metrics
// Story 9-5: Provide convenient access to metrics summary
func (r *TrinityRouter) GetMetricsSummary() MetricsSummary {
	if r.metrics != nil {
		return r.metrics.GetMetrics()
	}
	return MetricsSummary{}
}

// ResetMetrics resets all Trinity metrics
// Story 9-5: Allow metrics reset
func (r *TrinityRouter) ResetMetrics() {
	if r.metrics != nil {
		r.metrics.Reset()
	}
}

// LogMetricsSummary logs a summary of Trinity metrics
// Story 9-5: Provide convenient logging of metrics
func (r *TrinityRouter) LogMetricsSummary() {
	if r.metrics != nil {
		r.metrics.LogSummary()
	}
}
