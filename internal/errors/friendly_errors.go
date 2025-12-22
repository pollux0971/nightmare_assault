// Package errors provides user-friendly error handling for Nightmare Assault.
package errors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

// Translator interface for i18n support
type Translator interface {
	T(key string, params ...interface{}) string
}

// FriendlyError provides user-friendly error messages with i18n support.
type FriendlyError interface {
	error
	// UserMessage returns a localized, user-friendly error message
	UserMessage(translator Translator) string
	// ShouldRetry indicates if this error is retryable
	ShouldRetry() bool
	// ErrorCode returns a unique error code for debugging
	ErrorCode() string
	// Suggestion returns actionable advice for the user
	Suggestion(translator Translator) string
}

// NetworkError represents network connectivity errors.
type NetworkError struct {
	operation string
	err       error
}

// NewNetworkErrorFriendly creates a new NetworkError.
func NewNetworkErrorFriendly(operation string, err error) *NetworkError {
	return &NetworkError{
		operation: operation,
		err:       err,
	}
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %v", e.operation, e.err)
}

func (e *NetworkError) Unwrap() error {
	return e.err
}

func (e *NetworkError) UserMessage(t Translator) string {
	return t.T("errors.network_failure")
}

func (e *NetworkError) ShouldRetry() bool {
	return true
}

func (e *NetworkError) ErrorCode() string {
	return "NET_001"
}

func (e *NetworkError) Suggestion(t Translator) string {
	return t.T("errors.network_suggestion")
}

// APIError represents API-related errors with provider context.
type APIErrorFriendly struct {
	provider   string
	statusCode int
	operation  string
	err        error
}

// NewAPIErrorFriendly creates a new API error.
func NewAPIErrorFriendly(provider string, statusCode int, operation string, err error) *APIErrorFriendly {
	return &APIErrorFriendly{
		provider:   provider,
		statusCode: statusCode,
		operation:  operation,
		err:        err,
	}
}

func (e *APIErrorFriendly) Error() string {
	return fmt.Sprintf("API error [%s] during %s (status: %d): %v", e.provider, e.operation, e.statusCode, e.err)
}

func (e *APIErrorFriendly) Unwrap() error {
	return e.err
}

func (e *APIErrorFriendly) UserMessage(t Translator) string {
	// Check for specific error types
	if e.statusCode == 401 || e.statusCode == 403 {
		return t.T("errors.api_key_invalid")
	}
	if e.statusCode == 429 {
		return t.T("errors.api_rate_limited")
	}
	if e.statusCode >= 500 {
		return t.T("errors.api_server_error", e.provider)
	}
	if errors.Is(e.err, context.DeadlineExceeded) {
		return t.T("errors.api_timeout")
	}
	return t.T("errors.api_connection_failed", e.provider)
}

func (e *APIErrorFriendly) ShouldRetry() bool {
	// Don't retry authentication errors
	if e.statusCode == 401 || e.statusCode == 403 {
		return false
	}
	// Retry rate limits, server errors, timeouts
	if e.statusCode == 429 || e.statusCode >= 500 || errors.Is(e.err, context.DeadlineExceeded) {
		return true
	}
	// Retry network errors
	var netErr net.Error
	if errors.As(e.err, &netErr) {
		return true
	}
	return false
}

func (e *APIErrorFriendly) ErrorCode() string {
	switch {
	case e.statusCode == 401 || e.statusCode == 403:
		return "API_401"
	case e.statusCode == 429:
		return "API_429"
	case e.statusCode >= 500:
		return "API_5XX"
	case errors.Is(e.err, context.DeadlineExceeded):
		return "API_TIMEOUT"
	default:
		return "API_000"
	}
}

func (e *APIErrorFriendly) Suggestion(t Translator) string {
	if e.statusCode == 401 || e.statusCode == 403 {
		return t.T("errors.api_key_suggestion")
	}
	if e.statusCode == 429 {
		return t.T("errors.api_rate_limit_suggestion")
	}
	if e.statusCode >= 500 {
		return t.T("errors.api_server_suggestion")
	}
	if errors.Is(e.err, context.DeadlineExceeded) {
		return t.T("errors.api_timeout_suggestion")
	}
	return t.T("errors.retry_prompt")
}

// SaveFileError represents save file errors.
type SaveFileError struct {
	operation string
	slotID    int
	err       error
	corrupted bool
}

// NewSaveFileError creates a new save file error.
func NewSaveFileError(operation string, slotID int, err error) *SaveFileError {
	// Detect corruption vs not found vs permission
	corrupted := false
	if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrPermission) && !errors.Is(err, syscall.EACCES) {
		corrupted = true
	}
	return &SaveFileError{
		operation: operation,
		slotID:    slotID,
		err:       err,
		corrupted: corrupted,
	}
}

// NewSaveFileCorruptedError creates a corrupted save file error.
func NewSaveFileCorruptedError(slotID int, err error) *SaveFileError {
	return &SaveFileError{
		operation: "load",
		slotID:    slotID,
		err:       err,
		corrupted: true,
	}
}

// NewSaveFileNotFoundError creates a not found save file error.
func NewSaveFileNotFoundError(slotID int) *SaveFileError {
	return &SaveFileError{
		operation: "load",
		slotID:    slotID,
		err:       os.ErrNotExist,
		corrupted: false,
	}
}

// NewMigrationError creates a migration error.
func NewMigrationError(operation string, err error) error {
	return fmt.Errorf("migration error during %s: %w", operation, err)
}

func (e *SaveFileError) Error() string {
	return fmt.Sprintf("save file error during %s (slot %d): %v", e.operation, e.slotID, e.err)
}

func (e *SaveFileError) Unwrap() error {
	return e.err
}

func (e *SaveFileError) UserMessage(t Translator) string {
	if errors.Is(e.err, os.ErrNotExist) {
		return t.T("errors.save_not_found", e.slotID)
	}
	if e.corrupted {
		return t.T("errors.save_corrupted", e.slotID)
	}
	if errors.Is(e.err, os.ErrPermission) || errors.Is(e.err, syscall.EACCES) {
		return t.T("errors.save_permission_denied")
	}
	return t.T("errors.save_failed", e.slotID)
}

func (e *SaveFileError) ShouldRetry() bool {
	// Don't retry corruption or permission errors
	if e.corrupted || errors.Is(e.err, os.ErrPermission) {
		return false
	}
	// Don't retry not found errors
	if errors.Is(e.err, os.ErrNotExist) {
		return false
	}
	return false
}

func (e *SaveFileError) ErrorCode() string {
	if errors.Is(e.err, os.ErrNotExist) {
		return "SAVE_404"
	}
	if e.corrupted {
		return "SAVE_CORRUPT"
	}
	if errors.Is(e.err, os.ErrPermission) {
		return "SAVE_PERM"
	}
	return "SAVE_000"
}

func (e *SaveFileError) Suggestion(t Translator) string {
	if errors.Is(e.err, os.ErrNotExist) {
		return t.T("errors.save_not_found_suggestion")
	}
	if e.corrupted {
		return t.T("errors.save_corrupted_suggestion")
	}
	if errors.Is(e.err, os.ErrPermission) {
		return t.T("errors.save_permission_suggestion")
	}
	return t.T("errors.save_failed_suggestion")
}

// ConfigError represents configuration errors.
type ConfigError struct {
	field string
	err   error
}

// NewConfigErrorFriendly creates a new configuration error.
func NewConfigErrorFriendly(field string, err error) *ConfigError {
	return &ConfigError{
		field: field,
		err:   err,
	}
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in field '%s': %v", e.field, e.err)
}

func (e *ConfigError) Unwrap() error {
	return e.err
}

func (e *ConfigError) UserMessage(t Translator) string {
	return t.T("errors.config_invalid", e.field)
}

func (e *ConfigError) ShouldRetry() bool {
	return false
}

func (e *ConfigError) ErrorCode() string {
	return "CFG_001"
}

func (e *ConfigError) Suggestion(t Translator) string {
	return t.T("errors.config_suggestion")
}

// WrapWithFriendlyError wraps a standard error with a friendly error type.
// It detects the error type and returns the appropriate FriendlyError.
func WrapWithFriendlyError(err error, ctxStr string) error {
	if err == nil {
		return nil
	}

	// Already a friendly error
	var friendlyErr FriendlyError
	if errors.As(err, &friendlyErr) {
		return err
	}

	// Network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return NewNetworkErrorFriendly(ctxStr, err)
	}

	// Timeout errors
	if errors.Is(err, context.DeadlineExceeded) {
		return NewAPIErrorFriendly("unknown", 0, ctxStr, err)
	}

	// File errors
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return err // Let caller wrap with SaveFileError
	}

	// Return original error if can't classify
	return err
}

// FormatUserError formats a FriendlyError for display to the user.
func FormatUserError(err error, t Translator) string {
	if err == nil {
		return ""
	}

	var friendlyErr FriendlyError
	if errors.As(err, &friendlyErr) {
		msg := friendlyErr.UserMessage(t)
		suggestion := friendlyErr.Suggestion(t)
		if suggestion != "" {
			msg = msg + "\n\n" + suggestion
		}
		return msg
	}

	// Fallback to error message
	return err.Error()
}

// IsFriendlyError checks if an error is a FriendlyError.
func IsFriendlyError(err error) bool {
	var friendlyErr FriendlyError
	return errors.As(err, &friendlyErr)
}

// ShouldRetryError checks if an error should be retried.
func ShouldRetryError(err error) bool {
	var friendlyErr FriendlyError
	if errors.As(err, &friendlyErr) {
		return friendlyErr.ShouldRetry()
	}
	return false
}

// GetErrorCode extracts the error code from a FriendlyError.
func GetErrorCode(err error) string {
	var friendlyErr FriendlyError
	if errors.As(err, &friendlyErr) {
		return friendlyErr.ErrorCode()
	}
	return "UNKNOWN"
}
