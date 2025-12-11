package api

import (
	"errors"
	"fmt"
)

// Common API errors.
var (
	ErrInvalidAPIKey      = errors.New("API Key 無效，請檢查格式")
	ErrNetworkError       = errors.New("網路連線失敗，請檢查網路")
	ErrServiceUnavailable = errors.New("API 服務暫時無法使用")
	ErrRateLimited        = errors.New("請求過於頻繁，請稍後再試")
	ErrContextCanceled    = errors.New("操作已取消")
	ErrInvalidResponse    = errors.New("API 回應格式錯誤")
	ErrEmptyResponse      = errors.New("API 回應為空")
	ErrModelNotFound      = errors.New("指定的模型不存在")
	ErrQuotaExceeded      = errors.New("API 配額已用盡")
)

// APIError wraps an error with additional context.
type APIError struct {
	Provider   ProviderType
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
func NewAPIError(provider ProviderType, statusCode int, message string, err error) *APIError {
	return &APIError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403 ||
			errors.Is(apiErr.Err, ErrInvalidAPIKey)
	}
	return errors.Is(err, ErrInvalidAPIKey)
}

// IsNetworkError checks if the error is a network error.
func IsNetworkError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Err, ErrNetworkError)
	}
	return errors.Is(err, ErrNetworkError)
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429 || errors.Is(apiErr.Err, ErrRateLimited)
	}
	return errors.Is(err, ErrRateLimited)
}

// GetFriendlyMessage returns a user-friendly error message.
// Implements narrative error handling per ARCHITECTURE.md#14.3
func GetFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for specific error types
	switch {
	case IsAuthError(err):
		return "🔑 API Key 無效或已過期。請檢查您的設定..."
	case IsNetworkError(err):
		return "📡 通訊中斷...訊號在虛空中消散。請檢查網路連線。"
	case IsRateLimitError(err):
		return "⏳ 思緒變得遲鈍...請求過於頻繁，稍後再試。"
	case errors.Is(err, ErrServiceUnavailable):
		return "🌑 遠方的聲音暫時沉默...API 服務暫時無法使用。"
	case errors.Is(err, ErrEmptyResponse):
		return "👁️ 虛空凝視著你...但沒有回應。"
	case errors.Is(err, ErrQuotaExceeded):
		return "💀 力量的代價已付清...API 配額已用盡。"
	default:
		return fmt.Sprintf("❓ 未知的異常: %s", err.Error())
	}
}
