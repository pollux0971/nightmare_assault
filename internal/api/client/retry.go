package client

import (
	"context"
	"errors"
	"math"
	"time"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of attempts (including first try)
	InitialBackoff  time.Duration // Initial backoff duration
	MaxBackoff      time.Duration // Maximum backoff duration
	BackoffMultiple float64       // Backoff multiplier for exponential backoff
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  500 * time.Millisecond,
		MaxBackoff:      10 * time.Second,
		BackoffMultiple: 2.0,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func(ctx context.Context) error

// WithRetry executes a function with exponential backoff retry.
// It retries on network errors and rate limit errors (429), but not on auth errors (401, 403).
func WithRetry(ctx context.Context, cfg RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on certain errors
		if !shouldRetry(err) {
			return err
		}

		// Don't retry if this was the last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate backoff duration with exponential increase
		backoff := time.Duration(float64(cfg.InitialBackoff) * math.Pow(cfg.BackoffMultiple, float64(attempt-1)))
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}

		// Wait before retrying, respecting context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	// Return last error with retry context
	return &RetryError{
		Attempts: cfg.MaxAttempts,
		LastErr:  lastErr,
	}
}

// shouldRetry determines if an error is retryable.
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Unwrap APIError if present
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Don't retry authentication errors
		if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
			return false
		}

		// Retry network errors
		if errors.Is(apiErr.Err, ErrNetworkError) {
			return true
		}

		// Retry rate limit errors
		if errors.Is(apiErr.Err, ErrRateLimited) {
			return true
		}

		// Retry service unavailable
		if apiErr.StatusCode == 503 {
			return true
		}

		// Retry timeout errors
		if apiErr.StatusCode == 504 {
			return true
		}

		// Don't retry other HTTP errors
		if apiErr.StatusCode >= 400 {
			return false
		}

		return true
	}

	// Retry context deadline exceeded (timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Retry generic network errors
	if errors.Is(err, ErrNetworkError) {
		return true
	}

	// Retry rate limit errors
	if errors.Is(err, ErrRateLimited) {
		return true
	}

	// Retry service unavailable
	if errors.Is(err, ErrServiceUnavailable) {
		return true
	}

	// Default: don't retry
	return false
}

// RetryError wraps an error that failed after multiple retry attempts.
type RetryError struct {
	Attempts int
	LastErr  error
}

func (e *RetryError) Error() string {
	return e.LastErr.Error() + " (已重試 " + string(rune(e.Attempts)) + " 次)"
}

func (e *RetryError) Unwrap() error {
	return e.LastErr
}
