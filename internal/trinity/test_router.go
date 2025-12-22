package trinity

import (
	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// NewTrinityRouterWithProviders creates a TrinityRouter with pre-configured providers
// This is primarily used for testing to inject mock providers
//
// Parameters:
//   - thinkingProvider: Provider for Thinking tier
//   - reactiveProvider: Provider for Reactive tier
//   - rapidProvider: Provider for Rapid tier
//   - fallbackEnabled: Whether to enable fallback mechanism
//   - agentTierOverrides: Optional tier overrides for specific agents
//
// Returns:
//   - *TrinityRouter: Router instance with injected providers
func NewTrinityRouterWithProviders(
	thinkingProvider, reactiveProvider, rapidProvider client.Provider,
	fallbackEnabled bool,
	agentTierOverrides map[string]TierLevel,
) *TrinityRouter {
	// Merge default mapping with user overrides
	agentTierMap := make(map[string]TierLevel)
	for agent, tier := range DefaultAgentTierMapping {
		agentTierMap[agent] = tier
	}
	for agent, tier := range agentTierOverrides {
		agentTierMap[agent] = tier
	}

	// Create fallback handler with retry configuration
	var fallbackHandler *FallbackHandler
	if fallbackEnabled {
		retryConfig := client.DefaultRetryConfig()
		fallbackHandler = NewFallbackHandler(retryConfig)
	}

	// Initialize ThinkingMiddleware for automatic thinking tag processing
	thinkingMiddleware := NewThinkingMiddleware()

	// Initialize TrinityMetrics for metrics collection
	metrics := NewTrinityMetrics(1000)

	logger.Debug("TrinityRouter created with injected providers", map[string]interface{}{
		"fallback_enabled": fallbackEnabled,
		"agent_overrides":  len(agentTierOverrides),
	})

	return &TrinityRouter{
		thinkingProvider:   thinkingProvider,
		reactiveProvider:   reactiveProvider,
		rapidProvider:      rapidProvider,
		agentTierMap:       agentTierMap,
		fallbackEnabled:    fallbackEnabled,
		fallbackHandler:    fallbackHandler,
		thinkingMiddleware: thinkingMiddleware,
		metrics:            metrics,
	}
}
