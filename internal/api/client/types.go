// Package client provides API client implementations for different formats.
package client

import (
	"context"
	"errors"
	"fmt"
)

// Provider defines the interface for LLM API providers.
type Provider interface {
	Name() string
	TestConnection(ctx context.Context) error
	SendMessage(ctx context.Context, messages []Message) (*Response, error)
	Stream(ctx context.Context, messages []Message, callback func(chunk string)) error
	ModelInfo() ModelInfo
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
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

// Common errors.
var (
	ErrInvalidAPIKey      = errors.New("API Key 無效，請檢查格式")
	ErrNetworkError       = errors.New("網路連線失敗，請檢查網路")
	ErrServiceUnavailable = errors.New("API 服務暫時無法使用")
	ErrRateLimited        = errors.New("請求過於頻繁，請稍後再試")
	ErrEmptyResponse      = errors.New("API 回應為空")
)

// APIError wraps an error with additional context.
type APIError struct {
	Provider   string
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%s] %s (HTTP %d)", e.Provider, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("[%s] %s", e.Provider, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError.
func NewAPIError(provider string, statusCode int, message string, err error) *APIError {
	return &APIError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
