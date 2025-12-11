package api

import (
	"context"
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// NewProvider creates a new provider instance based on the configuration.
func NewProvider(config ProviderConfig) (Provider, error) {
	// Get provider info to determine format
	info := GetProviderInfo(config.ProviderID)
	if info == nil && config.ProviderID != "custom" {
		return nil, fmt.Errorf("unknown provider: %s", config.ProviderID)
	}

	// Determine the API format
	format := config.Format
	if format == "" && info != nil {
		format = info.Format
	}
	if format == "" {
		format = FormatOpenAI // Default to OpenAI format
	}

	// Determine base URL
	baseURL := config.BaseURL
	if baseURL == "" && info != nil {
		baseURL = info.BaseURL
	}

	switch format {
	case FormatOpenAI:
		return &providerWrapper{client.NewOpenAIClient(client.OpenAIConfig{
			ProviderID: config.ProviderID,
			APIKey:     config.APIKey,
			BaseURL:    baseURL,
			Model:      config.Model,
			MaxTokens:  config.MaxTokens,
		})}, nil

	case FormatAnthropic:
		return &providerWrapper{client.NewAnthropicClient(client.AnthropicConfig{
			ProviderID: config.ProviderID,
			APIKey:     config.APIKey,
			BaseURL:    baseURL,
			Model:      config.Model,
			MaxTokens:  config.MaxTokens,
		})}, nil

	case FormatGoogle:
		return &providerWrapper{client.NewGoogleClient(client.GoogleConfig{
			ProviderID: config.ProviderID,
			APIKey:     config.APIKey,
			BaseURL:    baseURL,
			Model:      config.Model,
			MaxTokens:  config.MaxTokens,
		})}, nil

	case FormatCohere:
		return &providerWrapper{client.NewCohereClient(client.CohereConfig{
			ProviderID: config.ProviderID,
			APIKey:     config.APIKey,
			BaseURL:    baseURL,
			Model:      config.Model,
			MaxTokens:  config.MaxTokens,
		})}, nil

	default:
		return nil, fmt.Errorf("unsupported API format: %s", format)
	}
}

// providerWrapper adapts client.Provider to api.Provider
type providerWrapper struct {
	inner client.Provider
}

func (w *providerWrapper) Name() string {
	return w.inner.Name()
}

func (w *providerWrapper) TestConnection(ctx context.Context) error {
	return w.inner.TestConnection(ctx)
}

func (w *providerWrapper) SendMessage(ctx context.Context, messages []Message) (*Response, error) {
	// Convert api.Message to client.Message
	clientMessages := make([]client.Message, len(messages))
	for i, m := range messages {
		clientMessages[i] = client.Message{Role: m.Role, Content: m.Content}
	}

	resp, err := w.inner.SendMessage(ctx, clientMessages)
	if err != nil {
		return nil, err
	}

	return &Response{
		Content:  resp.Content,
		Metadata: resp.Metadata,
	}, nil
}

func (w *providerWrapper) Stream(ctx context.Context, messages []Message, callback func(chunk string)) error {
	// Convert api.Message to client.Message
	clientMessages := make([]client.Message, len(messages))
	for i, m := range messages {
		clientMessages[i] = client.Message{Role: m.Role, Content: m.Content}
	}

	return w.inner.Stream(ctx, clientMessages, callback)
}

func (w *providerWrapper) ModelInfo() ModelInfo {
	info := w.inner.ModelInfo()
	return ModelInfo{
		Provider:  info.Provider,
		Model:     info.Model,
		MaxTokens: info.MaxTokens,
	}
}

// MustNewProvider creates a new provider and panics on error.
func MustNewProvider(config ProviderConfig) Provider {
	p, err := NewProvider(config)
	if err != nil {
		panic(err)
	}
	return p
}

// NewProviderFromID creates a provider using just the provider ID and API key.
func NewProviderFromID(providerID string, apiKey string) (Provider, error) {
	return NewProvider(ProviderConfig{
		ProviderID: providerID,
		APIKey:     apiKey,
	})
}
