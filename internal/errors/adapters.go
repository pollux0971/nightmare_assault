package errors

import (
	"context"
	"errors"
	"net/http"
	"strings"

	apiclient "github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// AdaptAPIError converts an API client error to a GameError.
func AdaptAPIError(err error, provider string) *GameError {
	if err == nil {
		return nil
	}

	// If already a GameError, return as is
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr
	}

	// Check for context errors
	if errors.Is(err, context.Canceled) {
		return &GameError{
			Type:        ErrorTypeAPI,
			Message:     "請求已取消",
			UserMessage: "操作已取消。",
			Retryable:   false,
			Err:         err,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return NewTimeoutError("API 請求超時", err)
	}

	// Check for API client errors
	var apiErr *apiclient.APIError
	if errors.As(err, &apiErr) {
		return adaptAPIClientError(apiErr, provider)
	}

	// Check for standard errors
	switch {
	case errors.Is(err, apiclient.ErrInvalidAPIKey):
		return NewAuthError("API Key 無效", err)
	case errors.Is(err, apiclient.ErrNetworkError):
		return NewNetworkError("網路連線失敗", err)
	case errors.Is(err, apiclient.ErrServiceUnavailable):
		return NewServiceUnavailableError("API 服務不可用", err)
	case errors.Is(err, apiclient.ErrRateLimited):
		return NewRateLimitError("請求頻率超過限制", err)
	case errors.Is(err, apiclient.ErrEmptyResponse):
		return NewAPIError("API 回應為空", err, provider)
	default:
		// Generic API error
		return NewAPIError(err.Error(), err, provider)
	}
}

// adaptAPIClientError converts an API client error to a GameError.
func adaptAPIClientError(apiErr *apiclient.APIError, provider string) *GameError {
	// Determine error type based on status code
	switch apiErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return NewAuthError(apiErr.Message, apiErr.Err)

	case http.StatusTooManyRequests:
		return NewRateLimitError(apiErr.Message, apiErr.Err)

	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return NewTimeoutError(apiErr.Message, apiErr.Err)

	case http.StatusServiceUnavailable, http.StatusBadGateway:
		return NewServiceUnavailableError(apiErr.Message, apiErr.Err)

	case http.StatusBadRequest:
		return NewConfigError(apiErr.Message, apiErr.Err)

	default:
		// Check underlying error
		if errors.Is(apiErr.Err, apiclient.ErrNetworkError) {
			return NewNetworkError(apiErr.Message, apiErr.Err)
		}
		if errors.Is(apiErr.Err, apiclient.ErrRateLimited) {
			return NewRateLimitError(apiErr.Message, apiErr.Err)
		}

		// Generic API error
		return NewAPIError(apiErr.Message, apiErr.Err, provider)
	}
}

// AdaptSaveError converts a save file error to a GameError.
func AdaptSaveError(err error, slotID int) *GameError {
	if err == nil {
		return nil
	}

	// If already a GameError, return as is
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr
	}

	errMsg := err.Error()

	// Check for specific save error patterns
	switch {
	case strings.Contains(errMsg, "不存在") || strings.Contains(errMsg, "空的"):
		return NewSaveNotFoundError(errMsg, slotID)

	case strings.Contains(errMsg, "損壞") || strings.Contains(errMsg, "扭曲") ||
		strings.Contains(errMsg, "校驗") || strings.Contains(errMsg, "無效"):
		return NewSaveCorruptError(errMsg, err, slotID)

	case strings.Contains(errMsg, "無法讀取") || strings.Contains(errMsg, "無法創建"):
		return &GameError{
			Type:        ErrorTypeSaveCorrupt,
			Message:     errMsg,
			UserMessage: "無法存取存檔檔案。",
			Suggestion:  "建議操作：\n• 檢查檔案權限\n• 確認儲存空間充足\n• 嘗試其他存檔槽",
			Retryable:   false,
			Err:         err,
			Context: map[string]interface{}{
				"slot_id": slotID,
			},
		}

	default:
		// Generic save error
		return &GameError{
			Type:        ErrorTypeSaveCorrupt,
			Message:     errMsg,
			UserMessage: "存檔操作失敗。",
			Suggestion:  "建議操作：\n• 檢查儲存空間\n• 重新啟動遊戲\n• 如問題持續，請報告",
			Retryable:   false,
			Err:         err,
		}
	}
}

// FormatRetryPrompt formats a retry prompt for the user.
func FormatRetryPrompt(err error) string {
	if !IsRetryable(err) {
		return ""
	}

	msg := GetFullMessage(err)
	msg += "\n\n是否要重試？(y/n)"
	return msg
}

// ShouldShowRetryOption checks if retry option should be shown.
func ShouldShowRetryOption(err error) bool {
	return IsRetryable(err)
}
