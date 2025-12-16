package context

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SummaryGenerator defines the interface for generating summaries from game history.
type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, entries []HistoryEntry) (string, error)
}

// GenerateRequest contains parameters for LLM generation.
type GenerateRequest struct {
	Prompt    string
	Model     string // e.g., "gpt-4o-mini"
	MaxTokens int    // e.g., 300
}

// LLMClient defines the interface for LLM API calls needed by summary generation.
// This interface supports model selection and token limits.
type LLMClient interface {
	Generate(ctx context.Context, req *GenerateRequest) (string, error)
}

// SummaryGeneratorConfig contains configuration for summary generation.
type SummaryGeneratorConfig struct {
	Model     string // Default: "gpt-4o-mini"
	MaxTokens int    // Default: 300
}

// DefaultSummaryConfig returns the default configuration for summary generation.
func DefaultSummaryConfig() SummaryGeneratorConfig {
	return SummaryGeneratorConfig{
		Model:     "gpt-4o-mini",
		MaxTokens: 300,
	}
}

// LLMSummaryGenerator generates summaries using an LLM client.
type LLMSummaryGenerator struct {
	client        LLMClient
	promptBuilder *SummaryPromptBuilder
	config        SummaryGeneratorConfig
}

// NewLLMSummaryGenerator creates a new LLMSummaryGenerator with default config.
func NewLLMSummaryGenerator(client LLMClient) *LLMSummaryGenerator {
	return NewLLMSummaryGeneratorWithConfig(client, DefaultSummaryConfig())
}

// NewLLMSummaryGeneratorWithConfig creates a new LLMSummaryGenerator with custom config.
func NewLLMSummaryGeneratorWithConfig(client LLMClient, config SummaryGeneratorConfig) *LLMSummaryGenerator {
	// Apply defaults if not specified
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 300
	}

	return &LLMSummaryGenerator{
		client:        client,
		promptBuilder: NewSummaryPromptBuilder(),
		config:        config,
	}
}

const (
	// MaxRetries is the maximum number of retry attempts for LLM API calls
	MaxRetries = 3
	// InitialBackoff is the initial backoff duration for retries
	InitialBackoff = 1 * time.Second
	// MaxBackoff is the maximum backoff duration for retries
	MaxBackoff = 10 * time.Second
)

// GenerateSummary generates a summary from the provided history entries.
// It builds a prompt, calls the LLM with retry logic, and validates the response format.
func (g *LLMSummaryGenerator) GenerateSummary(ctx context.Context, entries []HistoryEntry) (string, error) {
	// Build the prompt
	prompt := g.promptBuilder.BuildSummaryPrompt(entries)

	var lastErr error
	backoff := InitialBackoff

	// Retry loop with exponential backoff
	for attempt := 0; attempt < MaxRetries; attempt++ {
		// Check if context is cancelled
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("context cancelled: %w", err)
		}

		// Apply backoff delay (skip on first attempt)
		if attempt > 0 {
			select {
			case <-time.After(backoff):
				// Double the backoff, capped at MaxBackoff
				if backoff < MaxBackoff {
					backoff = backoff * 2
					if backoff > MaxBackoff {
						backoff = MaxBackoff
					}
				}
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		// Call LLM with configured model and max tokens
		summary, err := g.client.Generate(ctx, &GenerateRequest{
			Prompt:    prompt,
			Model:     g.config.Model,
			MaxTokens: g.config.MaxTokens,
		})

		if err == nil {
			// Validate response format
			if err := g.validateSummaryFormat(summary); err != nil {
				lastErr = fmt.Errorf("invalid format on attempt %d: %w", attempt+1, err)
				continue // Retry on format validation failure
			}
			return summary, nil
		}

		lastErr = err
		// Log retry attempt (in real implementation, use proper logger)
	}

	return "", fmt.Errorf("all %d retry attempts failed: %w", MaxRetries, lastErr)
}

// validateSummaryFormat checks if the summary follows the required format.
// Required format: "[Chapter X Summary: 簡要劇情. 角色狀態: XX. 已知線索: YY. 當前目標: ZZ.]"
func (g *LLMSummaryGenerator) validateSummaryFormat(summary string) error {
	summary = strings.TrimSpace(summary)

	if summary == "" {
		return errors.New("summary is empty")
	}

	// Check basic structure: must be wrapped in []
	if !strings.HasPrefix(summary, "[") || !strings.HasSuffix(summary, "]") {
		return errors.New("summary must be wrapped in []")
	}

	// Check required fields
	requiredFields := []string{
		"Chapter",
		"Summary:",
		"角色狀態:",
		"已知線索:",
		"當前目標:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(summary, field) {
			return errors.New("summary must contain '" + field + "'")
		}
	}

	return nil
}
