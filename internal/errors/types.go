// Package errors provides unified error handling for Nightmare Assault.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error.
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeNetwork represents a network connectivity error.
	ErrorTypeNetwork
	// ErrorTypeAPI represents an API-related error.
	ErrorTypeAPI
	// ErrorTypeAuth represents an authentication error.
	ErrorTypeAuth
	// ErrorTypeRateLimit represents a rate limiting error.
	ErrorTypeRateLimit
	// ErrorTypeSaveCorrupt represents a corrupted save file.
	ErrorTypeSaveCorrupt
	// ErrorTypeSaveNotFound represents a missing save file.
	ErrorTypeSaveNotFound
	// ErrorTypeConfig represents a configuration error.
	ErrorTypeConfig
	// ErrorTypeTimeout represents a timeout error.
	ErrorTypeTimeout
	// ErrorTypeServiceUnavailable represents a service unavailability error.
	ErrorTypeServiceUnavailable
)

// String returns the display name of the error type.
func (t ErrorType) String() string {
	switch t {
	case ErrorTypeNetwork:
		return "網路錯誤"
	case ErrorTypeAPI:
		return "API錯誤"
	case ErrorTypeAuth:
		return "認證錯誤"
	case ErrorTypeRateLimit:
		return "請求限制"
	case ErrorTypeSaveCorrupt:
		return "存檔損壞"
	case ErrorTypeSaveNotFound:
		return "存檔不存在"
	case ErrorTypeConfig:
		return "配置錯誤"
	case ErrorTypeTimeout:
		return "連線逾時"
	case ErrorTypeServiceUnavailable:
		return "服務不可用"
	default:
		return "未知錯誤"
	}
}

// GameError represents a unified game error with context.
type GameError struct {
	// Type is the error category
	Type ErrorType
	// Message is the technical error message
	Message string
	// UserMessage is the user-friendly message
	UserMessage string
	// Suggestion is what the user should do
	Suggestion string
	// Retryable indicates if the error can be retried
	Retryable bool
	// Err is the underlying error
	Err error
	// Context provides additional error context
	Context map[string]interface{}
}

// Error implements the error interface.
func (e *GameError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type.String(), e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying error.
func (e *GameError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the error can be retried.
func (e *GameError) IsRetryable() bool {
	return e.Retryable
}

// GetUserMessage returns the user-friendly error message.
func (e *GameError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return e.Message
}

// GetSuggestion returns the suggested action for the user.
func (e *GameError) GetSuggestion() string {
	return e.Suggestion
}

// GetFullMessage returns the complete user-facing message with suggestion.
func (e *GameError) GetFullMessage() string {
	msg := e.GetUserMessage()
	if e.Suggestion != "" {
		msg += "\n\n" + e.Suggestion
	}
	return msg
}

// Common predefined errors.
var (
	ErrNetworkConnection  = errors.New("網路連線失敗")
	ErrAPICallFailed      = errors.New("API 呼叫失敗")
	ErrInvalidAPIKey      = errors.New("API Key 無效")
	ErrRateLimitExceeded  = errors.New("請求頻率超過限制")
	ErrSaveFileCorrupted  = errors.New("存檔檔案損壞")
	ErrSaveFileNotFound   = errors.New("存檔檔案不存在")
	ErrConfigInvalid      = errors.New("配置無效")
	ErrTimeout            = errors.New("操作逾時")
	ErrServiceUnavailable = errors.New("服務暫時不可用")
)

// NewNetworkError creates a network error.
func NewNetworkError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeNetwork,
		Message:     message,
		UserMessage: "無法連接到網路，請檢查您的網路連線。",
		Suggestion:  "請確認：\n• 網路連線是否正常\n• 防火牆是否允許連線\n• 代理設定是否正確",
		Retryable:   true,
		Err:         err,
	}
}

// NewAPIError creates an API error.
func NewAPIError(message string, err error, provider string) *GameError {
	return &GameError{
		Type:        ErrorTypeAPI,
		Message:     message,
		UserMessage: fmt.Sprintf("無法連接到 LLM 服務，請檢查網路連接與 API Key 設定。"),
		Suggestion:  "建議操作：\n• 輸入 /api 重新配置 API Key\n• 確認 API Key 是否有效\n• 檢查 API 服務狀態",
		Retryable:   true,
		Err:         err,
		Context: map[string]interface{}{
			"provider": provider,
		},
	}
}

// NewAuthError creates an authentication error.
func NewAuthError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeAuth,
		Message:     message,
		UserMessage: "API Key 驗證失敗，請檢查您的 API Key 設定。",
		Suggestion:  "建議操作：\n• 輸入 /api 重新設定 API Key\n• 確認 API Key 格式正確\n• 檢查 API Key 是否已過期",
		Retryable:   false,
		Err:         err,
	}
}

// NewRateLimitError creates a rate limit error.
func NewRateLimitError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeRateLimit,
		Message:     message,
		UserMessage: "請求過於頻繁，API 服務已限制請求速率。",
		Suggestion:  "建議操作：\n• 稍等片刻後再試\n• 減少操作頻率\n• 考慮升級 API 配額",
		Retryable:   true,
		Err:         err,
	}
}

// NewSaveCorruptError creates a save corruption error.
func NewSaveCorruptError(message string, err error, slotID int) *GameError {
	return &GameError{
		Type:        ErrorTypeSaveCorrupt,
		Message:     message,
		UserMessage: "存檔文件損壞或版本不兼容。",
		Suggestion:  "建議操作：\n• 開始新遊戲\n• 嘗試其他存檔槽\n• 如果問題持續，請報告此問題",
		Retryable:   false,
		Err:         err,
		Context: map[string]interface{}{
			"slot_id": slotID,
		},
	}
}

// NewSaveNotFoundError creates a save not found error.
func NewSaveNotFoundError(message string, slotID int) *GameError {
	return &GameError{
		Type:        ErrorTypeSaveNotFound,
		Message:     message,
		UserMessage: fmt.Sprintf("存檔槽位 %d 是空的。", slotID),
		Suggestion:  "建議操作：\n• 選擇其他存檔槽位\n• 開始新遊戲",
		Retryable:   false,
		Context: map[string]interface{}{
			"slot_id": slotID,
		},
	}
}

// NewConfigError creates a configuration error.
func NewConfigError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeConfig,
		Message:     message,
		UserMessage: "遊戲配置有誤，請檢查設定。",
		Suggestion:  "建議操作：\n• 使用 /api 重新配置\n• 檢查配置檔案格式\n• 刪除配置檔案重新設定",
		Retryable:   false,
		Err:         err,
	}
}

// NewTimeoutError creates a timeout error.
func NewTimeoutError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeTimeout,
		Message:     message,
		UserMessage: "連線逾時，伺服器回應時間過長。",
		Suggestion:  "建議操作：\n• 稍後重試\n• 檢查網路連線速度\n• 選擇較快的 API 服務",
		Retryable:   true,
		Err:         err,
	}
}

// NewServiceUnavailableError creates a service unavailable error.
func NewServiceUnavailableError(message string, err error) *GameError {
	return &GameError{
		Type:        ErrorTypeServiceUnavailable,
		Message:     message,
		UserMessage: "API 服務暫時無法使用。",
		Suggestion:  "建議操作：\n• 稍後重試\n• 檢查 API 服務狀態頁面\n• 考慮切換其他 API 提供者",
		Retryable:   true,
		Err:         err,
	}
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr.IsRetryable()
	}
	return false
}

// GetUserMessage extracts the user-friendly message from an error.
func GetUserMessage(err error) string {
	if err == nil {
		return ""
	}

	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr.GetUserMessage()
	}

	return err.Error()
}

// GetSuggestion extracts the suggestion from an error.
func GetSuggestion(err error) string {
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr.GetSuggestion()
	}
	return ""
}

// GetFullMessage extracts the full user-facing message from an error.
func GetFullMessage(err error) string {
	if err == nil {
		return ""
	}

	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr.GetFullMessage()
	}

	return err.Error()
}

// WrapError wraps an existing error with GameError context.
func WrapError(err error, errType ErrorType, message string) *GameError {
	if err == nil {
		return nil
	}

	// If already a GameError, return as is
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr
	}

	// Create appropriate GameError based on type
	switch errType {
	case ErrorTypeNetwork:
		return NewNetworkError(message, err)
	case ErrorTypeAPI:
		return NewAPIError(message, err, "unknown")
	case ErrorTypeAuth:
		return NewAuthError(message, err)
	case ErrorTypeRateLimit:
		return NewRateLimitError(message, err)
	case ErrorTypeTimeout:
		return NewTimeoutError(message, err)
	case ErrorTypeServiceUnavailable:
		return NewServiceUnavailableError(message, err)
	case ErrorTypeConfig:
		return NewConfigError(message, err)
	default:
		return &GameError{
			Type:        errType,
			Message:     message,
			UserMessage: message,
			Err:         err,
			Retryable:   false,
		}
	}
}
