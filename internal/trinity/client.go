package trinity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// TrinityLLMClient provides a unified LLM client interface with Trinity routing
type TrinityLLMClient struct {
	router *TrinityRouter
	config TrinityClientConfig
}

// TrinityClientConfig contains configuration for TrinityLLMClient
type TrinityClientConfig struct {
	EnableThinkingExtraction bool          // Auto-extract thinking tags from responses
	EnableFallback           bool          // Enable automatic tier degradation
	EnableMetrics            bool          // Collect performance metrics
	DefaultTimeout           time.Duration // Default timeout for requests
}

// SendOptions contains optional parameters for SendMessageWithOptions
type SendOptions struct {
	TierOverride    *TierLevel    // Override the default tier for this agent
	DisableFallback bool          // Disable fallback for this specific request
	Timeout         time.Duration // Request-specific timeout
}

// NewTrinityLLMClient creates a new TrinityLLMClient instance
// AC1: NewTrinityLLMClient() can create client instance
func NewTrinityLLMClient(routerConfig RouterConfig, clientConfig TrinityClientConfig) (*TrinityLLMClient, error) {
	// Set fallback based on clientConfig
	routerConfig.FallbackEnabled = clientConfig.EnableFallback

	// Create the underlying router
	router, err := NewTrinityRouter(routerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Trinity router: %w", err)
	}

	logger.Debug("TrinityLLMClient initialized", map[string]interface{}{
		"thinking_extraction": clientConfig.EnableThinkingExtraction,
		"fallback_enabled":    clientConfig.EnableFallback,
		"metrics_enabled":     clientConfig.EnableMetrics,
		"default_timeout":     clientConfig.DefaultTimeout.String(),
	})

	return &TrinityLLMClient{
		router: router,
		config: clientConfig,
	}, nil
}

// SendMessage routes a message to the appropriate tier based on agent name
// AC1: SendMessage() can route based on agent name automatically
// AC3: Unknown agents default to Reactive tier
func (c *TrinityLLMClient) SendMessage(ctx context.Context, agentName string, messages []client.Message) (*client.Response, error) {
	// Use SendMessageWithOptions with no options (default behavior)
	return c.SendMessageWithOptions(ctx, agentName, messages, SendOptions{})
}

// SendMessageWithOptions sends a message with advanced options
// AC5: SendMessageWithOptions() allows tier override
// AC3: SendMessageWithOptions() allows disabling fallback per request
// AC5: Support context timeout
func (c *TrinityLLMClient) SendMessageWithOptions(ctx context.Context, agentName string, messages []client.Message, opts SendOptions) (*client.Response, error) {
	// Apply timeout if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	} else if c.config.DefaultTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.DefaultTimeout)
		defer cancel()
	}

	// Determine which tier to use
	var tier TierLevel
	if opts.TierOverride != nil {
		// Use override tier if specified
		tier = *opts.TierOverride
		logger.Debug("Using tier override", map[string]interface{}{
			"agent":         agentName,
			"override_tier": tier.String(),
		})
	} else {
		// Use router's tier mapping
		tier = GetTierForAgent(agentName, c.router.agentTierMap)
	}

	// If tier override is used OR fallback is disabled for this request, call provider directly
	if opts.TierOverride != nil || (opts.DisableFallback && c.router.fallbackEnabled) || !c.router.fallbackEnabled {
		// Call the provider directly without routing
		provider := c.router.getProviderForTier(tier)

		startTime := time.Now()
		resp, err := provider.SendMessage(ctx, messages)
		duration := time.Since(startTime)

		// Record metrics if enabled
		if c.config.EnableMetrics && c.router.metrics != nil {
			c.router.metrics.RecordRequest(tier, duration, err)
		}

		if err != nil {
			logger.Debug("Request failed (direct call)", map[string]interface{}{
				"agent": agentName,
				"tier":  tier.String(),
				"error": err.Error(),
			})
			return nil, fmt.Errorf("tier %s failed for agent %s: %w", tier.String(), agentName, err)
		}

		// Process thinking tags if enabled and using Thinking tier
		if c.config.EnableThinkingExtraction && tier == TierThinking && c.router.thinkingMiddleware != nil {
			resp = c.router.processThinkingTags(resp)
		}

		return resp, nil
	}

	// Use router's normal routing (with fallback enabled)
	resp, err := c.router.Route(ctx, agentName, messages)
	if err != nil {
		return nil, err
	}

	// Thinking tag processing is already handled by router.Route()
	// But if thinking extraction is disabled in client config, we need to undo it
	if !c.config.EnableThinkingExtraction && resp.Metadata != nil {
		// Remove thinking chain from metadata if client doesn't want it
		if thinkingChain, ok := resp.Metadata["thinking_chain"].(string); ok {
			// Restore thinking tags to content
			resp.Content = fmt.Sprintf("<thinking>%s</thinking>%s", thinkingChain, resp.Content)
			delete(resp.Metadata, "thinking_chain")
		}
	}

	return resp, nil
}

// ExtractThinking extracts thinking chain from response and returns cleaned content
// AC2: ExtractThinking() correctly extracts and cleans content
func (c *TrinityLLMClient) ExtractThinking(resp *client.Response) (thinkingChain string, cleanedContent string) {
	if resp == nil {
		return "", ""
	}

	// Check if thinking chain is already in metadata (already processed)
	if resp.Metadata != nil {
		if chain, ok := resp.Metadata["thinking_chain"].(string); ok {
			return chain, resp.Content
		}
	}

	// If not in metadata, extract from content
	if c.router.thinkingMiddleware != nil {
		thinkingChain = c.router.thinkingMiddleware.extractThinking(resp.Content)
		cleanedContent = c.router.thinkingMiddleware.removeThinkingTags(resp.Content)
		return thinkingChain, cleanedContent
	}

	// No middleware available
	return "", resp.Content
}

// HasThinking checks if a response contains thinking tags
// AC2: HasThinking() correctly detects thinking tags
func (c *TrinityLLMClient) HasThinking(resp *client.Response) bool {
	if resp == nil {
		return false
	}

	// Check metadata first
	if resp.Metadata != nil {
		if _, ok := resp.Metadata["thinking_chain"]; ok {
			return true
		}
	}

	// Check content for thinking tags
	return strings.Contains(resp.Content, "<thinking>") && strings.Contains(resp.Content, "</thinking>")
}

// GetMetrics returns the current Trinity metrics summary
// AC4: GetMetrics() returns complete metrics summary
func (c *TrinityLLMClient) GetMetrics() MetricsSummary {
	if !c.config.EnableMetrics {
		logger.Debug("Metrics not enabled", nil)
		return MetricsSummary{}
	}

	if c.router.metrics != nil {
		return c.router.GetMetricsSummary()
	}

	return MetricsSummary{}
}

// GetFallbackMetrics returns current fallback metrics
// AC3: Fallback metrics available when enabled
func (c *TrinityLLMClient) GetFallbackMetrics() *FallbackMetrics {
	if !c.config.EnableFallback {
		logger.Debug("Fallback not enabled", nil)
		return nil
	}

	return c.router.GetFallbackMetrics()
}

// LogPerformanceReport outputs a readable performance report
// AC4: LogPerformanceReport() outputs readable performance report
func (c *TrinityLLMClient) LogPerformanceReport() {
	if !c.config.EnableMetrics {
		logger.Info("Performance metrics not enabled", nil)
		return
	}

	logger.Info("=== Trinity LLM Client Performance Report ===", nil)

	// Log Trinity metrics summary
	if c.router.metrics != nil {
		c.router.LogMetricsSummary()
	}

	// Log fallback metrics if enabled
	if c.config.EnableFallback {
		fallbackMetrics := c.GetFallbackMetrics()
		if fallbackMetrics != nil {
			logger.Info("=== Fallback Metrics ===", map[string]interface{}{
				"total_fallbacks":        fallbackMetrics.TotalFallbacks,
				"full_degradation_count": fallbackMetrics.FullDegradationCount,
				"total_retries":          fallbackMetrics.TotalRetries,
			})

			if len(fallbackMetrics.FallbacksByTier) > 0 {
				logger.Info("Fallbacks by tier:", nil)
				for tier, count := range fallbackMetrics.FallbacksByTier {
					logger.Info(fmt.Sprintf("  %s: %d", tier.String(), count), nil)
				}
			}
		}
	}

	logger.Info("=== End of Performance Report ===", nil)
}

// UpdateAgentTier dynamically updates an agent's tier mapping
// AC5: UpdateAgentTier() can dynamically modify agent mapping
func (c *TrinityLLMClient) UpdateAgentTier(agentName string, tier TierLevel) {
	if c.router.agentTierMap == nil {
		c.router.agentTierMap = make(map[string]TierLevel)
	}

	c.router.agentTierMap[agentName] = tier

	logger.Info("Updated agent tier mapping", map[string]interface{}{
		"agent": agentName,
		"tier":  tier.String(),
	})
}

// ResetMetrics resets all Trinity metrics
// AC5: ResetMetrics() works correctly
func (c *TrinityLLMClient) ResetMetrics() {
	// Reset Trinity metrics
	if c.router.metrics != nil {
		c.router.ResetMetrics()
	}

	// Reset fallback metrics
	if c.router.fallbackHandler != nil {
		c.router.ResetFallbackMetrics()
	}

	logger.Info("Trinity metrics reset", nil)
}

// GetRouter returns the underlying TrinityRouter (for advanced usage)
func (c *TrinityLLMClient) GetRouter() *TrinityRouter {
	return c.router
}

// DefaultClientConfig returns a reasonable default client configuration
func DefaultClientConfig() TrinityClientConfig {
	return TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           true,
		EnableMetrics:            true,
		DefaultTimeout:           60 * time.Second,
	}
}
