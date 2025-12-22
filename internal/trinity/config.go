package trinity

import (
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// RouterConfig contains configuration for TrinityRouter
type RouterConfig struct {
	// Three provider configurations for each tier
	ThinkingProvider ProviderTierConfig
	ReactiveProvider ProviderTierConfig
	RapidProvider    ProviderTierConfig

	// Agent tier overrides (optional)
	AgentTierOverrides map[string]TierLevel

	// Enable fallback mechanism
	FallbackEnabled bool

	// Retry configuration for fallback mechanism
	// If not set, will use client.DefaultRetryConfig()
	RetryConfig client.RetryConfig
}

// ProviderTierConfig contains settings for a single tier's provider
type ProviderTierConfig struct {
	ProviderID  string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
}

// CreateProvider creates a Provider instance from ProviderTierConfig
func (c ProviderTierConfig) CreateProvider() (client.Provider, error) {
	switch c.ProviderID {
	case "anthropic":
		return client.NewAnthropicClient(client.AnthropicConfig{
			ProviderID: c.ProviderID,
			APIKey:     c.APIKey,
			Model:      c.Model,
			MaxTokens:  c.MaxTokens,
		}), nil
	case "openai", "openrouter":
		return client.NewOpenAIClient(client.OpenAIConfig{
			ProviderID: c.ProviderID,
			APIKey:     c.APIKey,
			Model:      c.Model,
			MaxTokens:  c.MaxTokens,
		}), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", c.ProviderID)
	}
}


// DefaultRouterConfig returns a reasonable default configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		ThinkingProvider: ProviderTierConfig{
			ProviderID:  "anthropic",
			Model:       "claude-opus-4-20250514",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: ProviderTierConfig{
			ProviderID:  "anthropic",
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: ProviderTierConfig{
			ProviderID:  "anthropic",
			Model:       "claude-3-haiku-20240307",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled:    true,
		AgentTierOverrides: make(map[string]TierLevel),
	}
}
