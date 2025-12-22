package client

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRetry_Success tests successful execution on first attempt.
func TestRetry_Success(t *testing.T) {
	cfg := DefaultRetryConfig()
	attempts := 0

	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return nil // Success on first try
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got: %d", attempts)
	}
}

// TestRetry_SuccessOnSecondAttempt tests retry mechanism with eventual success.
func TestRetry_SuccessOnSecondAttempt(t *testing.T) {
	cfg := DefaultRetryConfig()
	attempts := 0

	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		if attempts == 1 {
			return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
		}
		return nil // Success on second try
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got: %d", attempts)
	}
}

// TestRetry_MaxAttempts tests that retry stops after max attempts.
func TestRetry_MaxAttempts(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	attempts := 0

	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
	})

	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}

	// Check that error is wrapped as RetryError
	var retryErr *RetryError
	if !errors.As(err, &retryErr) {
		t.Error("Expected RetryError")
	}

	if retryErr != nil && retryErr.Attempts != 3 {
		t.Errorf("Expected RetryError with 3 attempts, got: %d", retryErr.Attempts)
	}
}

// TestRetry_NoRetryOnAuthError tests that auth errors are not retried.
func TestRetry_NoRetryOnAuthError(t *testing.T) {
	cfg := DefaultRetryConfig()
	attempts := 0

	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return NewAPIError("test", 401, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	})

	if err == nil {
		t.Error("Expected error")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry on auth error), got: %d", attempts)
	}

	// Should not be wrapped as RetryError since it wasn't retried
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		t.Error("Expected no RetryError for non-retried auth error")
	}
}

// TestRetry_RateLimitRetry tests that rate limit errors are retried.
func TestRetry_RateLimitRetry(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 2
	cfg.InitialBackoff = 10 * time.Millisecond // Fast for testing
	attempts := 0

	startTime := time.Now()
	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		if attempts == 1 {
			return NewAPIError("test", 429, ErrRateLimited.Error(), ErrRateLimited)
		}
		return nil // Success on second try
	})
	elapsed := time.Since(startTime)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got: %d", attempts)
	}

	// Should have waited at least the initial backoff
	if elapsed < cfg.InitialBackoff {
		t.Errorf("Expected at least %v delay, got: %v", cfg.InitialBackoff, elapsed)
	}
}

// TestRetry_ExponentialBackoff tests exponential backoff timing.
func TestRetry_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  50 * time.Millisecond,
		MaxBackoff:      500 * time.Millisecond,
		BackoffMultiple: 2.0,
	}
	attempts := 0

	startTime := time.Now()
	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
	})
	elapsed := time.Since(startTime)

	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}

	// Expected delays: 50ms (1st retry) + 100ms (2nd retry) = 150ms total minimum
	minExpected := 150 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("Expected at least %v delay with exponential backoff, got: %v", minExpected, elapsed)
	}
}

// TestRetry_ContextCancellation tests that retry respects context cancellation.
func TestRetry_ContextCancellation(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.InitialBackoff = 1 * time.Second // Long backoff to ensure cancellation happens
	attempts := 0

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := WithRetry(ctx, cfg, func(ctx context.Context) error {
		attempts++
		return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
	})

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}

	// Should have attempted once and then cancelled during backoff
	if attempts < 1 {
		t.Errorf("Expected at least 1 attempt, got: %d", attempts)
	}
}

// TestShouldRetry tests the shouldRetry decision logic.
func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantRetry  bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
		{
			name:      "network error",
			err:       NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError),
			wantRetry: true,
		},
		{
			name:      "rate limit error",
			err:       NewAPIError("test", 429, ErrRateLimited.Error(), ErrRateLimited),
			wantRetry: true,
		},
		{
			name:      "auth error 401",
			err:       NewAPIError("test", 401, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey),
			wantRetry: false,
		},
		{
			name:      "auth error 403",
			err:       NewAPIError("test", 403, "Forbidden", nil),
			wantRetry: false,
		},
		{
			name:      "service unavailable 503",
			err:       NewAPIError("test", 503, "Service Unavailable", nil),
			wantRetry: true,
		},
		{
			name:      "gateway timeout 504",
			err:       NewAPIError("test", 504, "Gateway Timeout", nil),
			wantRetry: true,
		},
		{
			name:      "bad request 400",
			err:       NewAPIError("test", 400, "Bad Request", nil),
			wantRetry: false,
		},
		{
			name:      "context deadline exceeded",
			err:       context.DeadlineExceeded,
			wantRetry: true,
		},
		{
			name:      "generic error",
			err:       errors.New("some random error"),
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRetry(tt.err)
			if got != tt.wantRetry {
				t.Errorf("shouldRetry() = %v, want %v for error: %v", got, tt.wantRetry, tt.err)
			}
		})
	}
}

// TestRetry_MaxBackoff tests that backoff doesn't exceed maximum.
func TestRetry_MaxBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:     5,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      50 * time.Millisecond, // Low max to test capping
		BackoffMultiple: 2.0,
	}
	attempts := 0

	startTime := time.Now()
	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
	})
	elapsed := time.Since(startTime)

	if err == nil {
		t.Error("Expected error after max attempts")
	}

	// With exponential backoff: 10ms, 20ms, 40ms, 80ms (capped to 50ms)
	// Total: 10 + 20 + 40 + 50 = 120ms minimum
	minExpected := 120 * time.Millisecond
	maxExpected := 200 * time.Millisecond // Allow some overhead

	if elapsed < minExpected {
		t.Errorf("Expected at least %v delay, got: %v", minExpected, elapsed)
	}

	if elapsed > maxExpected {
		t.Errorf("Expected at most %v delay, got: %v", maxExpected, elapsed)
	}
}

// TestRetryError_Unwrap tests that RetryError can be unwrapped.
func TestRetryError_Unwrap(t *testing.T) {
	originalErr := NewAPIError("test", 0, "original error", ErrNetworkError)
	retryErr := &RetryError{
		Attempts: 3,
		LastErr:  originalErr,
	}

	unwrapped := errors.Unwrap(retryErr)
	if unwrapped != originalErr {
		t.Error("RetryError.Unwrap() did not return original error")
	}

	// Test with errors.Is
	if !errors.Is(retryErr, ErrNetworkError) {
		t.Error("errors.Is() should find ErrNetworkError through RetryError")
	}
}

// TestDefaultRetryConfig tests the default retry configuration.
func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts=3, got: %d", cfg.MaxAttempts)
	}

	if cfg.InitialBackoff != 500*time.Millisecond {
		t.Errorf("Expected InitialBackoff=500ms, got: %v", cfg.InitialBackoff)
	}

	if cfg.MaxBackoff != 10*time.Second {
		t.Errorf("Expected MaxBackoff=10s, got: %v", cfg.MaxBackoff)
	}

	if cfg.BackoffMultiple != 2.0 {
		t.Errorf("Expected BackoffMultiple=2.0, got: %f", cfg.BackoffMultiple)
	}
}

// TestRetry_NetworkErrorRetry tests network error retry behavior.
func TestRetry_NetworkErrorRetry(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.InitialBackoff = 10 * time.Millisecond

	attempts := 0

	err := WithRetry(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return NewAPIError("test", 0, ErrNetworkError.Error(), ErrNetworkError)
		}
		return nil // Success on third attempt
	})

	if err != nil {
		t.Errorf("Expected success on third attempt, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
}
