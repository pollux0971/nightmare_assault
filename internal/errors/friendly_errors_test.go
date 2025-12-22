package errors

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
)

// mockTranslator provides a simple mock translator for testing
type mockTranslator struct {
	locale string
}

func newMockTranslator(locale string) *mockTranslator {
	return &mockTranslator{locale: locale}
}

func (m *mockTranslator) T(key string, params ...interface{}) string {
	// Mock translations for testing
	// Note: Uses "errors." prefix to match the actual i18n key structure
	translations := map[string]string{
		"errors.network_failure":            "Network failure",
		"errors.network_suggestion":         "Check your network",
		"errors.api_timeout":                "API timeout",
		"errors.api_key_invalid":            "Invalid API key",
		"errors.api_key_suggestion":         "Reconfigure API key",
		"errors.api_connection_failed":      "Cannot connect to {0}",
		"errors.api_rate_limited":           "Rate limited",
		"errors.api_rate_limit_suggestion":  "Wait before retrying",
		"errors.api_server_error":           "{0} server error",
		"errors.api_server_suggestion":      "Try again later",
		"errors.api_timeout_suggestion":     "Check connection speed",
		"errors.retry_prompt":               "Retry available",
		"errors.save_corrupted":             "Save slot {0} corrupted",
		"errors.save_corrupted_suggestion":  "Start new game or try different slot",
		"errors.save_not_found":             "Save slot {0} not found",
		"errors.save_not_found_suggestion":  "Select different slot",
		"errors.save_permission_denied":     "Permission denied",
		"errors.save_permission_suggestion": "Check file permissions",
		"errors.save_failed":                "Save to slot {0} failed",
		"errors.save_failed_suggestion":     "Check disk space",
		"errors.config_invalid":             "Config error: {0}",
		"errors.config_suggestion":          "Reconfigure settings",
	}

	if val, ok := translations[key]; ok {
		// Simple parameter replacement
		result := val
		for i, param := range params {
			placeholder := "{" + string(rune('0'+i)) + "}"
			result = strings.ReplaceAll(result, placeholder, toString(param))
		}
		return result
	}
	return "[" + key + "]"
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	if i, ok := v.(int); ok {
		return string(rune('0' + i))
	}
	return ""
}

func TestNetworkError(t *testing.T) {
	translator := newMockTranslator("en-US")
	baseErr := errors.New("connection refused")
	err := NewNetworkErrorFriendly("API call", baseErr)

	t.Run("implements FriendlyError", func(t *testing.T) {
		var friendlyErr FriendlyError
		if !errors.As(err, &friendlyErr) {
			t.Error("NetworkError should implement FriendlyError")
		}
	})

	t.Run("Error method", func(t *testing.T) {
		if !strings.Contains(err.Error(), "network error") {
			t.Errorf("Error() should contain 'network error', got: %s", err.Error())
		}
		if !strings.Contains(err.Error(), "API call") {
			t.Errorf("Error() should contain operation, got: %s", err.Error())
		}
	})

	t.Run("UserMessage", func(t *testing.T) {
		msg := err.UserMessage(translator)
		if msg != "Network failure" {
			t.Errorf("UserMessage() = %q, want 'Network failure'", msg)
		}
	})

	t.Run("ShouldRetry", func(t *testing.T) {
		if !err.ShouldRetry() {
			t.Error("NetworkError should be retryable")
		}
	})

	t.Run("ErrorCode", func(t *testing.T) {
		if err.ErrorCode() != "NET_001" {
			t.Errorf("ErrorCode() = %q, want 'NET_001'", err.ErrorCode())
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		suggestion := err.Suggestion(translator)
		if suggestion != "Check your network" {
			t.Errorf("Suggestion() = %q, want 'Check your network'", suggestion)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		if !errors.Is(err, baseErr) {
			t.Error("Should unwrap to base error")
		}
	})
}

func TestAPIErrorFriendly(t *testing.T) {
	translator := newMockTranslator("en-US")

	t.Run("401 Unauthorized", func(t *testing.T) {
		err := NewAPIErrorFriendly("anthropic", 401, "test connection", errors.New("unauthorized"))

		if err.ShouldRetry() {
			t.Error("401 errors should not be retryable")
		}

		if err.ErrorCode() != "API_401" {
			t.Errorf("ErrorCode() = %q, want 'API_401'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if msg != "Invalid API key" {
			t.Errorf("UserMessage() = %q, want 'Invalid API key'", msg)
		}

		suggestion := err.Suggestion(translator)
		if suggestion != "Reconfigure API key" {
			t.Errorf("Suggestion() = %q, want 'Reconfigure API key'", suggestion)
		}
	})

	t.Run("429 Rate Limited", func(t *testing.T) {
		err := NewAPIErrorFriendly("openai", 429, "send message", errors.New("rate limited"))

		if !err.ShouldRetry() {
			t.Error("429 errors should be retryable")
		}

		if err.ErrorCode() != "API_429" {
			t.Errorf("ErrorCode() = %q, want 'API_429'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if msg != "Rate limited" {
			t.Errorf("UserMessage() = %q, want 'Rate limited'", msg)
		}
	})

	t.Run("500 Server Error", func(t *testing.T) {
		err := NewAPIErrorFriendly("openrouter", 500, "send message", errors.New("server error"))

		if !err.ShouldRetry() {
			t.Error("500 errors should be retryable")
		}

		if err.ErrorCode() != "API_5XX" {
			t.Errorf("ErrorCode() = %q, want 'API_5XX'", err.ErrorCode())
		}
	})

	t.Run("Timeout Error", func(t *testing.T) {
		err := NewAPIErrorFriendly("anthropic", 0, "send message", context.DeadlineExceeded)

		if !err.ShouldRetry() {
			t.Error("Timeout errors should be retryable")
		}

		if err.ErrorCode() != "API_TIMEOUT" {
			t.Errorf("ErrorCode() = %q, want 'API_TIMEOUT'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if msg != "API timeout" {
			t.Errorf("UserMessage() = %q, want 'API timeout'", msg)
		}
	})

	t.Run("Network Error", func(t *testing.T) {
		netErr := &net.DNSError{Err: "no such host"}
		err := NewAPIErrorFriendly("anthropic", 0, "send message", netErr)

		if !err.ShouldRetry() {
			t.Error("Network errors should be retryable")
		}
	})
}

func TestSaveFileError(t *testing.T) {
	translator := newMockTranslator("en-US")

	t.Run("Not Found Error", func(t *testing.T) {
		err := NewSaveFileNotFoundError(2)

		if err.ShouldRetry() {
			t.Error("Not found errors should not be retryable")
		}

		if err.ErrorCode() != "SAVE_404" {
			t.Errorf("ErrorCode() = %q, want 'SAVE_404'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if !strings.Contains(msg, "not found") {
			t.Errorf("UserMessage() should contain 'not found', got: %q", msg)
		}

		if !errors.Is(err, os.ErrNotExist) {
			t.Error("Should wrap os.ErrNotExist")
		}
	})

	t.Run("Corrupted Error", func(t *testing.T) {
		err := NewSaveFileCorruptedError(3, errors.New("invalid JSON"))

		if err.ShouldRetry() {
			t.Error("Corrupted errors should not be retryable")
		}

		if err.ErrorCode() != "SAVE_CORRUPT" {
			t.Errorf("ErrorCode() = %q, want 'SAVE_CORRUPT'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if !strings.Contains(msg, "corrupted") {
			t.Errorf("UserMessage() should contain 'corrupted', got: %q", msg)
		}
	})

	t.Run("Permission Error", func(t *testing.T) {
		err := NewSaveFileError("write", 1, os.ErrPermission)

		if err.ShouldRetry() {
			t.Error("Permission errors should not be retryable")
		}

		if err.ErrorCode() != "SAVE_PERM" {
			t.Errorf("ErrorCode() = %q, want 'SAVE_PERM'", err.ErrorCode())
		}

		msg := err.UserMessage(translator)
		if msg != "Permission denied" {
			t.Errorf("UserMessage() = %q, want 'Permission denied'", msg)
		}
	})
}

func TestConfigError(t *testing.T) {
	baseErr := errors.New("invalid value")
	err := NewConfigErrorFriendly("api_key", baseErr)

	t.Run("implements FriendlyError", func(t *testing.T) {
		var friendlyErr FriendlyError
		if !errors.As(err, &friendlyErr) {
			t.Error("ConfigError should implement FriendlyError")
		}
	})

	t.Run("ShouldRetry", func(t *testing.T) {
		if err.ShouldRetry() {
			t.Error("Config errors should not be retryable")
		}
	})

	t.Run("ErrorCode", func(t *testing.T) {
		if err.ErrorCode() != "CFG_001" {
			t.Errorf("ErrorCode() = %q, want 'CFG_001'", err.ErrorCode())
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		if !errors.Is(err, baseErr) {
			t.Error("Should unwrap to base error")
		}
	})
}

func TestWrapWithFriendlyError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := WrapWithFriendlyError(nil, "test")
		if err != nil {
			t.Errorf("WrapWithFriendlyError(nil) should return nil, got: %v", err)
		}
	})

	t.Run("already friendly error", func(t *testing.T) {
		original := NewNetworkErrorFriendly("test", errors.New("base"))
		wrapped := WrapWithFriendlyError(original, "test")
		if wrapped != original {
			t.Error("Should return same error if already friendly")
		}
	})

	t.Run("network error", func(t *testing.T) {
		netErr := &net.DNSError{Err: "no such host"}
		wrapped := WrapWithFriendlyError(netErr, "API call")

		var networkErr *NetworkError
		if !errors.As(wrapped, &networkErr) {
			t.Error("Should wrap network error as NetworkError")
		}
	})

	t.Run("timeout error", func(t *testing.T) {
		wrapped := WrapWithFriendlyError(context.DeadlineExceeded, "API call")

		// Should wrap timeout error
		if wrapped == nil {
			t.Fatal("Should wrap timeout error")
		}
	})
}

func TestFormatUserError(t *testing.T) {
	translator := newMockTranslator("en-US")

	t.Run("nil error", func(t *testing.T) {
		msg := FormatUserError(nil, translator)
		if msg != "" {
			t.Errorf("FormatUserError(nil) should return empty string, got: %q", msg)
		}
	})

	t.Run("friendly error with suggestion", func(t *testing.T) {
		err := NewNetworkErrorFriendly("test", errors.New("base"))
		msg := FormatUserError(err, translator)

		if !strings.Contains(msg, "Network failure") {
			t.Errorf("Should contain user message, got: %q", msg)
		}
		if !strings.Contains(msg, "Check your network") {
			t.Errorf("Should contain suggestion, got: %q", msg)
		}
	})

	t.Run("non-friendly error", func(t *testing.T) {
		err := errors.New("generic error")
		msg := FormatUserError(err, translator)

		if msg != "generic error" {
			t.Errorf("Should return error message for non-friendly error, got: %q", msg)
		}
	})
}

func TestIsFriendlyError(t *testing.T) {
	t.Run("is friendly", func(t *testing.T) {
		err := NewNetworkErrorFriendly("test", errors.New("base"))
		if !IsFriendlyError(err) {
			t.Error("Should return true for FriendlyError")
		}
	})

	t.Run("not friendly", func(t *testing.T) {
		err := errors.New("generic error")
		if IsFriendlyError(err) {
			t.Error("Should return false for non-FriendlyError")
		}
	})

	t.Run("wrapped friendly error", func(t *testing.T) {
		baseErr := NewNetworkErrorFriendly("test", errors.New("base"))
		wrappedErr := errors.Join(baseErr, errors.New("other"))
		if !IsFriendlyError(wrappedErr) {
			t.Error("Should detect wrapped FriendlyError")
		}
	})
}

func TestShouldRetryError(t *testing.T) {
	t.Run("retryable error", func(t *testing.T) {
		err := NewNetworkErrorFriendly("test", errors.New("base"))
		if !ShouldRetryError(err) {
			t.Error("NetworkError should be retryable")
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		err := NewConfigErrorFriendly("field", errors.New("base"))
		if ShouldRetryError(err) {
			t.Error("ConfigError should not be retryable")
		}
	})

	t.Run("non-friendly error", func(t *testing.T) {
		err := errors.New("generic error")
		if ShouldRetryError(err) {
			t.Error("Non-FriendlyError should not be retryable by default")
		}
	})
}

func TestGetErrorCode(t *testing.T) {
	t.Run("friendly error", func(t *testing.T) {
		err := NewNetworkErrorFriendly("test", errors.New("base"))
		code := GetErrorCode(err)
		if code != "NET_001" {
			t.Errorf("GetErrorCode() = %q, want 'NET_001'", code)
		}
	})

	t.Run("non-friendly error", func(t *testing.T) {
		err := errors.New("generic error")
		code := GetErrorCode(err)
		if code != "UNKNOWN" {
			t.Errorf("GetErrorCode() = %q, want 'UNKNOWN'", code)
		}
	})
}

func TestRealI18nIntegration(t *testing.T) {
	// This test requires actual i18n translator initialization
	// Skip if translator cannot be initialized
	translator, err := i18n.New("en-US")
	if err != nil {
		t.Skipf("Skipping i18n integration test: %v", err)
	}

	t.Run("NetworkError with real translator", func(t *testing.T) {
		netErr := NewNetworkErrorFriendly("API call", errors.New("connection refused"))
		msg := netErr.UserMessage(translator)

		if msg == "" || msg == "[error.network_failure]" {
			t.Errorf("Should return translated message, got: %q", msg)
		}
	})

	t.Run("APIError with real translator", func(t *testing.T) {
		apiErr := NewAPIErrorFriendly("anthropic", 401, "test", errors.New("unauthorized"))
		msg := apiErr.UserMessage(translator)

		if msg == "" || msg == "[error.api_key_invalid]" {
			t.Errorf("Should return translated message, got: %q", msg)
		}
	})

	t.Run("SaveFileError with real translator", func(t *testing.T) {
		saveErr := NewSaveFileNotFoundError(2)
		msg := saveErr.UserMessage(translator)

		if msg == "" || strings.HasPrefix(msg, "[error.") {
			t.Errorf("Should return translated message, got: %q", msg)
		}
	})
}

func TestConvertAPIError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := ConvertAPIError(nil, "test", 0, "test")
		if err != nil {
			t.Errorf("ConvertAPIError(nil) should return nil, got: %v", err)
		}
	})

	t.Run("already friendly error", func(t *testing.T) {
		original := NewNetworkErrorFriendly("test", errors.New("base"))
		converted := ConvertAPIError(original, "test", 0, "test")
		if converted != original {
			t.Error("Should return same error if already friendly")
		}
	})

	t.Run("network error", func(t *testing.T) {
		netErr := &net.DNSError{Err: "no such host"}
		converted := ConvertAPIError(netErr, "anthropic", 0, "send message")

		var networkErr *NetworkError
		if !errors.As(converted, &networkErr) {
			t.Error("Should convert to NetworkError")
		}
	})

	t.Run("timeout error", func(t *testing.T) {
		converted := ConvertAPIError(context.DeadlineExceeded, "anthropic", 0, "send message")

		// Context errors are treated as API errors
		if converted == nil {
			t.Fatal("Should convert to error")
		}
	})

	t.Run("generic error with status code", func(t *testing.T) {
		converted := ConvertAPIError(errors.New("test"), "openai", 500, "send message")

		var apiErr *APIErrorFriendly
		if !errors.As(converted, &apiErr) {
			t.Error("Should convert to APIErrorFriendly")
		}
	})
}
