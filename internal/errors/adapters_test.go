package errors

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	apiclient "github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

func TestAdaptAPIError_ContextCanceled(t *testing.T) {
	err := context.Canceled
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeAPI {
		t.Errorf("Type = %v, want ErrorTypeAPI", adapted.Type)
	}
	if adapted.Retryable {
		t.Error("Canceled errors should not be retryable")
	}
	if !strings.Contains(adapted.UserMessage, "取消") {
		t.Error("UserMessage should mention cancellation")
	}
}

func TestAdaptAPIError_ContextDeadlineExceeded(t *testing.T) {
	err := context.DeadlineExceeded
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want ErrorTypeTimeout", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Timeout errors should be retryable")
	}
	if !strings.Contains(adapted.UserMessage, "逾時") {
		t.Error("UserMessage should mention timeout")
	}
}

func TestAdaptAPIError_InvalidAPIKey(t *testing.T) {
	err := apiclient.ErrInvalidAPIKey
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want ErrorTypeAuth", adapted.Type)
	}
	if adapted.Retryable {
		t.Error("Auth errors should not be retryable")
	}
	if !strings.Contains(adapted.UserMessage, "API Key") {
		t.Error("UserMessage should mention API Key")
	}
}

func TestAdaptAPIError_NetworkError(t *testing.T) {
	err := apiclient.ErrNetworkError
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeNetwork {
		t.Errorf("Type = %v, want ErrorTypeNetwork", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Network errors should be retryable")
	}
	if !strings.Contains(adapted.UserMessage, "網路") {
		t.Error("UserMessage should mention network")
	}
}

func TestAdaptAPIError_RateLimited(t *testing.T) {
	err := apiclient.ErrRateLimited
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want ErrorTypeRateLimit", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Rate limit errors should be retryable")
	}
	if !strings.Contains(adapted.UserMessage, "頻繁") {
		t.Error("UserMessage should mention frequency")
	}
}

func TestAdaptAPIError_ServiceUnavailable(t *testing.T) {
	err := apiclient.ErrServiceUnavailable
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeServiceUnavailable {
		t.Errorf("Type = %v, want ErrorTypeServiceUnavailable", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Service unavailable errors should be retryable")
	}
}

func TestAdaptAPIError_EmptyResponse(t *testing.T) {
	err := apiclient.ErrEmptyResponse
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeAPI {
		t.Errorf("Type = %v, want ErrorTypeAPI", adapted.Type)
	}
	if !strings.Contains(adapted.UserMessage, "LLM 服務") {
		t.Error("UserMessage should mention LLM service")
	}
}

func TestAdaptAPIError_AlreadyGameError(t *testing.T) {
	original := NewNetworkError("network error", nil)
	adapted := AdaptAPIError(original, "openai")

	if adapted != original {
		t.Error("Should return original GameError")
	}
}

func TestAdaptAPIError_GenericError(t *testing.T) {
	err := errors.New("some generic error")
	adapted := AdaptAPIError(err, "openai")

	if adapted.Type != ErrorTypeAPI {
		t.Errorf("Type = %v, want ErrorTypeAPI", adapted.Type)
	}
	if adapted.Retryable != true {
		t.Error("Generic API errors should be retryable")
	}
}

func TestAdaptAPIClientError_Unauthorized(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusUnauthorized,
		Message:    "unauthorized",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want ErrorTypeAuth", adapted.Type)
	}
	if adapted.Retryable {
		t.Error("Auth errors should not be retryable")
	}
}

func TestAdaptAPIClientError_TooManyRequests(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusTooManyRequests,
		Message:    "rate limited",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want ErrorTypeRateLimit", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Rate limit errors should be retryable")
	}
}

func TestAdaptAPIClientError_RequestTimeout(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusRequestTimeout,
		Message:    "timeout",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want ErrorTypeTimeout", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Timeout errors should be retryable")
	}
}

func TestAdaptAPIClientError_ServiceUnavailable(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusServiceUnavailable,
		Message:    "service down",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeServiceUnavailable {
		t.Errorf("Type = %v, want ErrorTypeServiceUnavailable", adapted.Type)
	}
	if !adapted.Retryable {
		t.Error("Service unavailable errors should be retryable")
	}
}

func TestAdaptAPIClientError_BadRequest(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusBadRequest,
		Message:    "bad request",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeConfig {
		t.Errorf("Type = %v, want ErrorTypeConfig", adapted.Type)
	}
	if adapted.Retryable {
		t.Error("Config errors should not be retryable")
	}
}

func TestAdaptAPIClientError_WithNetworkError(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: 0,
		Message:    "network failed",
		Err:        apiclient.ErrNetworkError,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeNetwork {
		t.Errorf("Type = %v, want ErrorTypeNetwork", adapted.Type)
	}
}

func TestAdaptSaveError_NotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"不存在", errors.New("檔案不存在")},
		{"空的", errors.New("槽位是空的")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapted := AdaptSaveError(tt.err, 2)

			if adapted.Type != ErrorTypeSaveNotFound {
				t.Errorf("Type = %v, want ErrorTypeSaveNotFound", adapted.Type)
			}
			if adapted.Retryable {
				t.Error("Save not found errors should not be retryable")
			}
			if adapted.Context["slot_id"] != 2 {
				t.Error("Context should contain slot_id")
			}
		})
	}
}

func TestAdaptSaveError_Corrupted(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"損壞", errors.New("存檔損壞")},
		{"扭曲", errors.New("記憶扭曲")},
		{"校驗", errors.New("校驗失敗")},
		{"無效", errors.New("格式無效")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapted := AdaptSaveError(tt.err, 3)

			if adapted.Type != ErrorTypeSaveCorrupt {
				t.Errorf("Type = %v, want ErrorTypeSaveCorrupt", adapted.Type)
			}
			if adapted.Retryable {
				t.Error("Save corrupt errors should not be retryable")
			}
			if !strings.Contains(adapted.Suggestion, "新遊戲") {
				t.Error("Suggestion should mention new game")
			}
		})
	}
}

func TestAdaptSaveError_FileAccess(t *testing.T) {
	err := errors.New("無法讀取檔案")
	adapted := AdaptSaveError(err, 1)

	if adapted.Type != ErrorTypeSaveCorrupt {
		t.Errorf("Type = %v, want ErrorTypeSaveCorrupt", adapted.Type)
	}
	if !strings.Contains(adapted.Suggestion, "權限") {
		t.Error("Suggestion should mention permissions")
	}
}

func TestAdaptSaveError_AlreadyGameError(t *testing.T) {
	original := NewSaveCorruptError("corrupted", nil, 1)
	adapted := AdaptSaveError(original, 1)

	if adapted != original {
		t.Error("Should return original GameError")
	}
}

func TestAdaptSaveError_GenericError(t *testing.T) {
	err := errors.New("unknown save error")
	adapted := AdaptSaveError(err, 2)

	if adapted.Type != ErrorTypeSaveCorrupt {
		t.Errorf("Type = %v, want ErrorTypeSaveCorrupt", adapted.Type)
	}
	if adapted.Retryable {
		t.Error("Generic save errors should not be retryable")
	}
}

func TestAdaptSaveError_Nil(t *testing.T) {
	adapted := AdaptSaveError(nil, 1)

	if adapted != nil {
		t.Error("AdaptSaveError(nil) should return nil")
	}
}

func TestFormatRetryPrompt(t *testing.T) {
	retryableErr := NewNetworkError("network error", nil)
	nonRetryableErr := NewAuthError("auth error", nil)

	// Test retryable error
	prompt := FormatRetryPrompt(retryableErr)
	if !strings.Contains(prompt, "重試") {
		t.Error("Prompt should contain retry question")
	}
	if !strings.Contains(prompt, "y/n") {
		t.Error("Prompt should contain y/n options")
	}
	if !strings.Contains(prompt, retryableErr.UserMessage) {
		t.Error("Prompt should contain user message")
	}

	// Test non-retryable error
	prompt = FormatRetryPrompt(nonRetryableErr)
	if prompt != "" {
		t.Error("Non-retryable error should return empty prompt")
	}
}

func TestShouldShowRetryOption(t *testing.T) {
	retryableErr := NewNetworkError("network error", nil)
	nonRetryableErr := NewAuthError("auth error", nil)

	if !ShouldShowRetryOption(retryableErr) {
		t.Error("Should show retry option for retryable error")
	}
	if ShouldShowRetryOption(nonRetryableErr) {
		t.Error("Should not show retry option for non-retryable error")
	}
}

func TestAdaptAPIError_Nil(t *testing.T) {
	adapted := AdaptAPIError(nil, "openai")

	if adapted != nil {
		t.Error("AdaptAPIError(nil) should return nil")
	}
}

func TestAdaptAPIClientError_GatewayTimeout(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusGatewayTimeout,
		Message:    "gateway timeout",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want ErrorTypeTimeout", adapted.Type)
	}
}

func TestAdaptAPIClientError_BadGateway(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusBadGateway,
		Message:    "bad gateway",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeServiceUnavailable {
		t.Errorf("Type = %v, want ErrorTypeServiceUnavailable", adapted.Type)
	}
}

func TestAdaptAPIClientError_Forbidden(t *testing.T) {
	apiErr := &apiclient.APIError{
		Provider:   "openai",
		StatusCode: http.StatusForbidden,
		Message:    "forbidden",
		Err:        nil,
	}

	adapted := AdaptAPIError(apiErr, "openai")

	if adapted.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want ErrorTypeAuth", adapted.Type)
	}
}
