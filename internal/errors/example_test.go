package errors_test

import (
	"context"
	"fmt"
	"testing"

	apiclient "github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/errors"
)

// Example demonstrates how to handle API errors.
func ExampleAdaptAPIError() {
	// Simulate an API timeout
	err := context.DeadlineExceeded

	// Convert to GameError
	gameErr := errors.AdaptAPIError(err, "openai")

	// Display user-friendly message
	fmt.Println(gameErr.GetUserMessage())
	fmt.Println("\n" + gameErr.GetSuggestion())

	// Check if retryable
	if gameErr.IsRetryable() {
		fmt.Println("\n可以重試")
	}
}

// Example demonstrates how to handle save file errors.
func ExampleAdaptSaveError() {
	// Simulate a corrupted save file
	err := fmt.Errorf("存檔檔案損壞")

	// Convert to GameError
	gameErr := errors.AdaptSaveError(err, 2)

	// Get full message with suggestion
	fullMsg := gameErr.GetFullMessage()
	fmt.Println(fullMsg)
}

// Example demonstrates error wrapping.
func ExampleWrapError() {
	// Original error
	originalErr := fmt.Errorf("connection refused")

	// Wrap with context
	gameErr := errors.WrapError(originalErr, errors.ErrorTypeNetwork, "無法連接到伺服器")

	// Access error information
	fmt.Printf("Type: %s\n", gameErr.Type)
	fmt.Printf("User Message: %s\n", gameErr.GetUserMessage())
	fmt.Printf("Retryable: %v\n", gameErr.IsRetryable())
}

// Example demonstrates creating API errors directly.
func ExampleNewAPIError() {
	baseErr := fmt.Errorf("timeout after 30s")
	gameErr := errors.NewAPIError("API 請求超時", baseErr, "anthropic")

	// Get formatted retry prompt
	if errors.IsRetryable(gameErr) {
		retryPrompt := errors.FormatRetryPrompt(gameErr)
		fmt.Println(retryPrompt)
	}
}

// Example demonstrates adapting different API client errors.
func ExampleAdaptAPIError_statusCodes() {
	testCases := []struct {
		name       string
		statusCode int
		errType    errors.ErrorType
	}{
		{"Unauthorized", 401, errors.ErrorTypeAuth},
		{"Rate Limited", 429, errors.ErrorTypeRateLimit},
		{"Timeout", 504, errors.ErrorTypeTimeout},
		{"Service Unavailable", 503, errors.ErrorTypeServiceUnavailable},
	}

	for _, tc := range testCases {
		apiErr := &apiclient.APIError{
			Provider:   "test",
			StatusCode: tc.statusCode,
			Message:    tc.name,
		}

		gameErr := errors.AdaptAPIError(apiErr, "test")
		fmt.Printf("%s -> %s\n", tc.name, gameErr.Type)
	}
}

// TestUsageExample shows typical usage in application code.
func TestUsageExample(t *testing.T) {
	// Simulating API call failure
	err := simulateAPICall()

	// Adapt error
	gameErr := errors.AdaptAPIError(err, "openai")

	// Display to user
	userMsg := gameErr.GetUserMessage()
	if userMsg == "" {
		t.Error("User message should not be empty")
	}

	// Check if we should offer retry
	if errors.ShouldShowRetryOption(err) {
		retryPrompt := errors.FormatRetryPrompt(err)
		if retryPrompt == "" {
			t.Error("Retry prompt should not be empty for retryable errors")
		}
	}

	// Log technical details (for debugging)
	technicalMsg := gameErr.Error()
	if technicalMsg == "" {
		t.Error("Technical message should not be empty")
	}
}

// TestSaveErrorHandling shows typical save error handling.
func TestSaveErrorHandling(t *testing.T) {
	// Simulating save file corruption
	err := fmt.Errorf("存檔文件損壞或版本不兼容")
	slotID := 1

	// Adapt error
	gameErr := errors.AdaptSaveError(err, slotID)

	// Check error type
	if gameErr.Type != errors.ErrorTypeSaveCorrupt {
		t.Errorf("Expected save corrupt error, got %v", gameErr.Type)
	}

	// Display message to user
	fullMsg := gameErr.GetFullMessage()
	if fullMsg == "" {
		t.Error("Full message should not be empty")
	}

	// Verify suggestion includes alternative actions
	suggestion := gameErr.GetSuggestion()
	if suggestion == "" {
		t.Error("Suggestion should provide alternative actions")
	}
}

// TestErrorChaining demonstrates error chaining.
func TestErrorChaining(t *testing.T) {
	// Original low-level error
	baseErr := fmt.Errorf("socket closed")

	// Wrap with network context
	networkErr := errors.NewNetworkError("連線中斷", baseErr)

	// Check unwrapping
	if networkErr.Unwrap() != baseErr {
		t.Error("Should preserve original error")
	}

	// Verify error chain
	if networkErr.Err != baseErr {
		t.Error("Error chain should be preserved")
	}
}

// simulateAPICall simulates an API call that might fail.
func simulateAPICall() error {
	// Simulate timeout
	return context.DeadlineExceeded
}

// TestErrorTypeChecking demonstrates type checking patterns.
func TestErrorTypeChecking(t *testing.T) {
	testErrors := []error{
		errors.NewNetworkError("network", nil),
		errors.NewAuthError("auth", nil),
		errors.NewRateLimitError("rate", nil),
	}

	for _, err := range testErrors {
		// Pattern 1: Check retryability
		if errors.IsRetryable(err) {
			// Can retry
			t.Logf("Error is retryable: %v", err)
		}

		// Pattern 2: Get user message
		userMsg := errors.GetUserMessage(err)
		if userMsg == "" {
			t.Error("User message should not be empty")
		}

		// Pattern 3: Get suggestion
		suggestion := errors.GetSuggestion(err)
		// Suggestion may be empty for some error types
		_ = suggestion
	}
}

// TestContextPreservation demonstrates context preservation.
func TestContextPreservation(t *testing.T) {
	slotID := 3
	saveErr := errors.NewSaveCorruptError("corrupted", nil, slotID)

	// Access context
	if saveErr.Context == nil {
		t.Fatal("Context should not be nil")
	}

	storedSlotID, ok := saveErr.Context["slot_id"].(int)
	if !ok || storedSlotID != slotID {
		t.Errorf("Context should preserve slot_id, got %v", saveErr.Context["slot_id"])
	}
}
