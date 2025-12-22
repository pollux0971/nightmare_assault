package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{ErrorTypeNetwork, "網路錯誤"},
		{ErrorTypeAPI, "API錯誤"},
		{ErrorTypeAuth, "認證錯誤"},
		{ErrorTypeRateLimit, "請求限制"},
		{ErrorTypeSaveCorrupt, "存檔損壞"},
		{ErrorTypeSaveNotFound, "存檔不存在"},
		{ErrorTypeConfig, "配置錯誤"},
		{ErrorTypeTimeout, "連線逾時"},
		{ErrorTypeServiceUnavailable, "服務不可用"},
		{ErrorTypeUnknown, "未知錯誤"},
	}

	for _, tt := range tests {
		if got := tt.errType.String(); got != tt.expected {
			t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestGameErrorError(t *testing.T) {
	baseErr := errors.New("base error")
	gameErr := &GameError{
		Type:    ErrorTypeNetwork,
		Message: "connection failed",
		Err:     baseErr,
	}

	result := gameErr.Error()
	if !strings.Contains(result, "網路錯誤") {
		t.Errorf("Error() should contain error type, got: %s", result)
	}
	if !strings.Contains(result, "connection failed") {
		t.Errorf("Error() should contain message, got: %s", result)
	}
}

func TestGameErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	gameErr := &GameError{
		Type: ErrorTypeNetwork,
		Err:  baseErr,
	}

	if gameErr.Unwrap() != baseErr {
		t.Error("Unwrap() should return underlying error")
	}
}

func TestGameErrorIsRetryable(t *testing.T) {
	retryableErr := &GameError{Retryable: true}
	nonRetryableErr := &GameError{Retryable: false}

	if !retryableErr.IsRetryable() {
		t.Error("IsRetryable() should return true for retryable error")
	}
	if nonRetryableErr.IsRetryable() {
		t.Error("IsRetryable() should return false for non-retryable error")
	}
}

func TestGameErrorGetUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		gameErr  *GameError
		expected string
	}{
		{
			name: "with user message",
			gameErr: &GameError{
				Message:     "technical message",
				UserMessage: "user-friendly message",
			},
			expected: "user-friendly message",
		},
		{
			name: "without user message",
			gameErr: &GameError{
				Message: "technical message",
			},
			expected: "technical message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.gameErr.GetUserMessage(); got != tt.expected {
				t.Errorf("GetUserMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGameErrorGetSuggestion(t *testing.T) {
	gameErr := &GameError{
		Suggestion: "try this",
	}

	if got := gameErr.GetSuggestion(); got != "try this" {
		t.Errorf("GetSuggestion() = %v, want 'try this'", got)
	}
}

func TestGameErrorGetFullMessage(t *testing.T) {
	gameErr := &GameError{
		UserMessage: "error occurred",
		Suggestion:  "try again",
	}

	result := gameErr.GetFullMessage()
	if !strings.Contains(result, "error occurred") {
		t.Error("GetFullMessage() should contain user message")
	}
	if !strings.Contains(result, "try again") {
		t.Error("GetFullMessage() should contain suggestion")
	}
}

func TestNewNetworkError(t *testing.T) {
	baseErr := errors.New("connection refused")
	err := NewNetworkError("network down", baseErr)

	if err.Type != ErrorTypeNetwork {
		t.Errorf("Type = %v, want ErrorTypeNetwork", err.Type)
	}
	if err.Message != "network down" {
		t.Errorf("Message = %v, want 'network down'", err.Message)
	}
	if !err.Retryable {
		t.Error("Network errors should be retryable")
	}
	if err.Err != baseErr {
		t.Error("Underlying error not preserved")
	}
	if !strings.Contains(err.UserMessage, "網路") {
		t.Error("UserMessage should mention network in Chinese")
	}
	if !strings.Contains(err.Suggestion, "網路連線") {
		t.Error("Suggestion should mention network connection")
	}
}

func TestNewAPIError(t *testing.T) {
	baseErr := errors.New("API call failed")
	err := NewAPIError("API timeout", baseErr, "openai")

	if err.Type != ErrorTypeAPI {
		t.Errorf("Type = %v, want ErrorTypeAPI", err.Type)
	}
	if !err.Retryable {
		t.Error("API errors should be retryable")
	}
	if !strings.Contains(err.UserMessage, "LLM 服務") {
		t.Error("UserMessage should mention LLM service")
	}
	if !strings.Contains(err.Suggestion, "/api") {
		t.Error("Suggestion should mention /api command")
	}
	if err.Context["provider"] != "openai" {
		t.Error("Context should contain provider")
	}
}

func TestNewAuthError(t *testing.T) {
	baseErr := errors.New("invalid key")
	err := NewAuthError("auth failed", baseErr)

	if err.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want ErrorTypeAuth", err.Type)
	}
	if err.Retryable {
		t.Error("Auth errors should not be retryable")
	}
	if !strings.Contains(err.UserMessage, "API Key") {
		t.Error("UserMessage should mention API Key")
	}
	if !strings.Contains(err.Suggestion, "/api") {
		t.Error("Suggestion should mention /api command")
	}
}

func TestNewRateLimitError(t *testing.T) {
	baseErr := errors.New("rate limit")
	err := NewRateLimitError("too many requests", baseErr)

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want ErrorTypeRateLimit", err.Type)
	}
	if !err.Retryable {
		t.Error("Rate limit errors should be retryable")
	}
	if !strings.Contains(err.UserMessage, "頻繁") {
		t.Error("UserMessage should mention frequency")
	}
	if !strings.Contains(err.Suggestion, "稍等") {
		t.Error("Suggestion should suggest waiting")
	}
}

func TestNewSaveCorruptError(t *testing.T) {
	baseErr := errors.New("checksum mismatch")
	err := NewSaveCorruptError("corrupted save", baseErr, 2)

	if err.Type != ErrorTypeSaveCorrupt {
		t.Errorf("Type = %v, want ErrorTypeSaveCorrupt", err.Type)
	}
	if err.Retryable {
		t.Error("Save corrupt errors should not be retryable")
	}
	if !strings.Contains(err.UserMessage, "損壞") {
		t.Error("UserMessage should mention corruption")
	}
	if !strings.Contains(err.Suggestion, "新遊戲") {
		t.Error("Suggestion should suggest new game")
	}
	if err.Context["slot_id"] != 2 {
		t.Error("Context should contain slot_id")
	}
}

func TestNewSaveNotFoundError(t *testing.T) {
	err := NewSaveNotFoundError("save not found", 3)

	if err.Type != ErrorTypeSaveNotFound {
		t.Errorf("Type = %v, want ErrorTypeSaveNotFound", err.Type)
	}
	if err.Retryable {
		t.Error("Save not found errors should not be retryable")
	}
	if !strings.Contains(err.UserMessage, "3") {
		t.Error("UserMessage should contain slot number")
	}
	if err.Context["slot_id"] != 3 {
		t.Error("Context should contain slot_id")
	}
}

func TestNewConfigError(t *testing.T) {
	baseErr := errors.New("invalid config")
	err := NewConfigError("config error", baseErr)

	if err.Type != ErrorTypeConfig {
		t.Errorf("Type = %v, want ErrorTypeConfig", err.Type)
	}
	if err.Retryable {
		t.Error("Config errors should not be retryable")
	}
	if !strings.Contains(err.UserMessage, "配置") {
		t.Error("UserMessage should mention configuration")
	}
}

func TestNewTimeoutError(t *testing.T) {
	baseErr := errors.New("timeout")
	err := NewTimeoutError("request timeout", baseErr)

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want ErrorTypeTimeout", err.Type)
	}
	if !err.Retryable {
		t.Error("Timeout errors should be retryable")
	}
	if !strings.Contains(err.UserMessage, "逾時") {
		t.Error("UserMessage should mention timeout")
	}
}

func TestNewServiceUnavailableError(t *testing.T) {
	baseErr := errors.New("service down")
	err := NewServiceUnavailableError("service unavailable", baseErr)

	if err.Type != ErrorTypeServiceUnavailable {
		t.Errorf("Type = %v, want ErrorTypeServiceUnavailable", err.Type)
	}
	if !err.Retryable {
		t.Error("Service unavailable errors should be retryable")
	}
	if !strings.Contains(err.UserMessage, "服務") {
		t.Error("UserMessage should mention service")
	}
}

func TestIsRetryable(t *testing.T) {
	retryableErr := NewNetworkError("network error", nil)
	nonRetryableErr := NewAuthError("auth error", nil)
	regularErr := errors.New("regular error")

	if !IsRetryable(retryableErr) {
		t.Error("IsRetryable() should return true for network error")
	}
	if IsRetryable(nonRetryableErr) {
		t.Error("IsRetryable() should return false for auth error")
	}
	if IsRetryable(regularErr) {
		t.Error("IsRetryable() should return false for regular error")
	}
}

func TestGetUserMessage(t *testing.T) {
	gameErr := NewNetworkError("network error", nil)
	regularErr := errors.New("regular error")

	if msg := GetUserMessage(gameErr); !strings.Contains(msg, "網路") {
		t.Errorf("GetUserMessage() should return Chinese message, got: %s", msg)
	}
	if msg := GetUserMessage(regularErr); msg != "regular error" {
		t.Errorf("GetUserMessage() should return error string for regular error, got: %s", msg)
	}
	if msg := GetUserMessage(nil); msg != "" {
		t.Errorf("GetUserMessage(nil) should return empty string, got: %s", msg)
	}
}

func TestGetSuggestion(t *testing.T) {
	gameErr := NewAPIError("api error", nil, "openai")
	regularErr := errors.New("regular error")

	if sug := GetSuggestion(gameErr); !strings.Contains(sug, "/api") {
		t.Errorf("GetSuggestion() should contain /api, got: %s", sug)
	}
	if sug := GetSuggestion(regularErr); sug != "" {
		t.Errorf("GetSuggestion() should return empty for regular error, got: %s", sug)
	}
}

func TestGetFullMessage(t *testing.T) {
	gameErr := NewAPIError("api error", nil, "openai")
	regularErr := errors.New("regular error")

	msg := GetFullMessage(gameErr)
	if !strings.Contains(msg, "LLM 服務") {
		t.Error("GetFullMessage() should contain user message")
	}
	if !strings.Contains(msg, "/api") {
		t.Error("GetFullMessage() should contain suggestion")
	}

	if msg := GetFullMessage(regularErr); msg != "regular error" {
		t.Errorf("GetFullMessage() should return error string for regular error, got: %s", msg)
	}
	if msg := GetFullMessage(nil); msg != "" {
		t.Errorf("GetFullMessage(nil) should return empty string, got: %s", msg)
	}
}

func TestWrapError(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		errType  ErrorType
		expected ErrorType
	}{
		{"network", ErrorTypeNetwork, ErrorTypeNetwork},
		{"api", ErrorTypeAPI, ErrorTypeAPI},
		{"auth", ErrorTypeAuth, ErrorTypeAuth},
		{"rate_limit", ErrorTypeRateLimit, ErrorTypeRateLimit},
		{"timeout", ErrorTypeTimeout, ErrorTypeTimeout},
		{"service_unavailable", ErrorTypeServiceUnavailable, ErrorTypeServiceUnavailable},
		{"config", ErrorTypeConfig, ErrorTypeConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(baseErr, tt.errType, "wrapped message")
			if wrapped.Type != tt.expected {
				t.Errorf("WrapError() type = %v, want %v", wrapped.Type, tt.expected)
			}
			if wrapped.Err != baseErr {
				t.Error("WrapError() should preserve underlying error")
			}
		})
	}

	// Test wrapping GameError
	gameErr := NewNetworkError("network error", baseErr)
	wrapped := WrapError(gameErr, ErrorTypeAPI, "api message")
	if wrapped.Type != ErrorTypeNetwork {
		t.Error("WrapError() should preserve GameError type")
	}

	// Test wrapping nil
	if wrapped := WrapError(nil, ErrorTypeNetwork, "message"); wrapped != nil {
		t.Error("WrapError(nil) should return nil")
	}
}

func TestGameErrorContext(t *testing.T) {
	err := NewSaveCorruptError("corrupted", nil, 5)

	if err.Context == nil {
		t.Fatal("Context should not be nil")
	}
	if slotID, ok := err.Context["slot_id"].(int); !ok || slotID != 5 {
		t.Error("Context should contain correct slot_id")
	}
}

func TestErrorMessagesInChinese(t *testing.T) {
	// Verify all error messages are in Chinese
	errors := []*GameError{
		NewNetworkError("test", nil),
		NewAPIError("test", nil, "test"),
		NewAuthError("test", nil),
		NewRateLimitError("test", nil),
		NewSaveCorruptError("test", nil, 1),
		NewSaveNotFoundError("test", 1),
		NewConfigError("test", nil),
		NewTimeoutError("test", nil),
		NewServiceUnavailableError("test", nil),
	}

	for _, err := range errors {
		// Check UserMessage contains Chinese characters
		if !containsChinese(err.UserMessage) {
			t.Errorf("UserMessage should be in Chinese: %s", err.UserMessage)
		}
		// Check Suggestion contains Chinese characters
		if err.Suggestion != "" && !containsChinese(err.Suggestion) {
			t.Errorf("Suggestion should be in Chinese: %s", err.Suggestion)
		}
	}
}

// containsChinese checks if a string contains Chinese characters.
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
