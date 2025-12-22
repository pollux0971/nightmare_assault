// Package errors provides API client error adapters.
package errors

import (
	"context"
	"errors"
	"net"
)

// APIClientError represents the standard APIError from api/client package.
// This is used to avoid circular imports.
type APIClientError interface {
	error
	Unwrap() error
}

// ConvertAPIError converts an api/client error to a FriendlyError.
// This adapter function is used to convert existing APIError types to FriendlyError.
func ConvertAPIError(err error, provider string, statusCode int, operation string) error {
	if err == nil {
		return nil
	}

	// Already a friendly error
	var friendlyErr FriendlyError
	if errors.As(err, &friendlyErr) {
		return err
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return NewNetworkErrorFriendly(operation, err)
	}

	// Check for timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return NewAPIErrorFriendly(provider, 0, operation, err)
	}

	// Create API error with status code
	return NewAPIErrorFriendly(provider, statusCode, operation, err)
}

// WrapAPITestConnection wraps API test connection errors with friendly errors.
func WrapAPITestConnection(err error, provider string) error {
	if err == nil {
		return nil
	}

	// Extract status code if available
	var apiErr interface {
		StatusCode() int
	}
	statusCode := 0
	if errors.As(err, &apiErr) {
		statusCode = apiErr.StatusCode()
	}

	return ConvertAPIError(err, provider, statusCode, "test connection")
}

// WrapAPISendMessage wraps API send message errors with friendly errors.
func WrapAPISendMessage(err error, provider string) error {
	if err == nil {
		return nil
	}

	// Extract status code if available
	var apiErr interface {
		StatusCode() int
	}
	statusCode := 0
	if errors.As(err, &apiErr) {
		statusCode = apiErr.StatusCode()
	}

	return ConvertAPIError(err, provider, statusCode, "send message")
}

// WrapAPIStream wraps API stream errors with friendly errors.
func WrapAPIStream(err error, provider string) error {
	if err == nil {
		return nil
	}

	// Extract status code if available
	var apiErr interface {
		StatusCode() int
	}
	statusCode := 0
	if errors.As(err, &apiErr) {
		statusCode = apiErr.StatusCode()
	}

	return ConvertAPIError(err, provider, statusCode, "stream")
}
