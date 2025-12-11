// Package api provides LLM API provider interfaces and implementations.
package api

import (
	"context"
)

// Provider defines the interface for LLM API providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// TestConnection tests the API connection.
	TestConnection(ctx context.Context) error

	// SendMessage sends a message and returns the response.
	SendMessage(ctx context.Context, messages []Message) (*Response, error)

	// Stream sends a message and streams the response via callback.
	Stream(ctx context.Context, messages []Message, callback func(chunk string)) error

	// ModelInfo returns information about the model.
	ModelInfo() ModelInfo
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// Response represents an API response.
type Response struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ModelInfo contains information about the model.
type ModelInfo struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

// APIFormat represents the API format type.
type APIFormat string

const (
	FormatOpenAI    APIFormat = "openai"
	FormatAnthropic APIFormat = "anthropic"
	FormatGoogle    APIFormat = "google"
	FormatCohere    APIFormat = "cohere"
)

// ProviderType is an alias for provider ID strings.
type ProviderType string
