// Package logger provides structured logging for Nightmare Assault.
package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSetDebugMode tests enabling and disabling debug mode.
// Story 10-8 AC1, AC2: Debug mode controls log level and file routing
func TestSetDebugMode(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := New(tmpDir, WARN, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Verify initial state (debug mode disabled by default)
	if logger.IsDebugMode() {
		t.Error("Expected debug mode to be disabled by default")
	}
	if logger.GetLevel() != WARN {
		t.Errorf("Expected log level WARN, got %v", logger.GetLevel())
	}

	// Enable debug mode
	if err := logger.SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode: %v", err)
	}

	// Verify debug mode is enabled
	if !logger.IsDebugMode() {
		t.Error("Expected debug mode to be enabled")
	}
	if logger.GetLevel() != DEBUG {
		t.Errorf("Expected log level DEBUG, got %v", logger.GetLevel())
	}

	// Verify debug.log file was created
	debugPath := filepath.Join(tmpDir, "debug.log")
	if _, err := os.Stat(debugPath); os.IsNotExist(err) {
		t.Error("Expected debug.log to be created when debug mode is enabled")
	}

	// Disable debug mode
	if err := logger.SetDebugMode(false); err != nil {
		t.Fatalf("Failed to disable debug mode: %v", err)
	}

	// Verify debug mode is disabled
	if logger.IsDebugMode() {
		t.Error("Expected debug mode to be disabled")
	}
	if logger.GetLevel() != WARN {
		t.Errorf("Expected log level WARN, got %v", logger.GetLevel())
	}
}

// TestDebugModeLogging tests that debug logs are written to debug.log.
// Story 10-8 AC1: Debug mode logs all levels to debug.log
func TestDebugModeLogging(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger with debug mode enabled
	logger, err := New(tmpDir, DEBUG, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Enable debug mode
	if err := logger.SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode: %v", err)
	}

	// Log messages at various levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Give file I/O time to complete
	time.Sleep(100 * time.Millisecond)

	// Read debug.log
	debugPath := filepath.Join(tmpDir, "debug.log")
	debugData, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("Failed to read debug.log: %v", err)
	}

	// Verify all messages are in debug.log
	debugContent := string(debugData)
	if debugContent == "" {
		t.Error("Expected debug.log to contain log entries")
	}

	// Parse JSON lines and verify messages
	expectedMessages := []string{"Debug message", "Info message", "Warning message", "Error message"}
	for _, expectedMsg := range expectedMessages {
		if !contains(debugContent, expectedMsg) {
			t.Errorf("Expected debug.log to contain %q", expectedMsg)
		}
	}
}

// TestNormalModeLogging tests that only errors/warnings are logged in normal mode.
// Story 10-8 AC2: Normal mode only logs errors and warnings to error.log
func TestNormalModeLogging(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger with WARN level (normal mode)
	logger, err := New(tmpDir, WARN, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Verify debug mode is disabled
	if logger.IsDebugMode() {
		t.Error("Expected debug mode to be disabled")
	}

	// Log messages at various levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Give file I/O time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify debug.log was NOT created
	debugPath := filepath.Join(tmpDir, "debug.log")
	if _, err := os.Stat(debugPath); !os.IsNotExist(err) {
		t.Error("Expected debug.log to NOT be created in normal mode")
	}

	// Read error.log
	errorPath := filepath.Join(tmpDir, "error.log")
	errorData, err := os.ReadFile(errorPath)
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	errorContent := string(errorData)

	// Verify only warnings and errors are in error.log
	if !contains(errorContent, "Warning message") {
		t.Error("Expected error.log to contain warning message")
	}
	if !contains(errorContent, "Error message") {
		t.Error("Expected error.log to contain error message")
	}

	// Verify debug and info messages are NOT in error.log
	if contains(errorContent, "Debug message") {
		t.Error("Expected error.log to NOT contain debug message in normal mode")
	}
	if contains(errorContent, "Info message") {
		t.Error("Expected error.log to NOT contain info message in normal mode")
	}
}

// TestErrorLoggingInDebugMode tests that errors are logged to both debug.log and error.log.
// Story 10-8 AC1, AC2: Errors should be in both files when debug mode is enabled
func TestErrorLoggingInDebugMode(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger with debug mode enabled
	logger, err := New(tmpDir, DEBUG, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Enable debug mode
	if err := logger.SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode: %v", err)
	}

	// Log an error
	logger.Error("Test error message")

	// Give file I/O time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify error is in debug.log
	debugPath := filepath.Join(tmpDir, "debug.log")
	debugData, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("Failed to read debug.log: %v", err)
	}
	if !contains(string(debugData), "Test error message") {
		t.Error("Expected error to be in debug.log")
	}

	// Verify error is also in error.log
	errorPath := filepath.Join(tmpDir, "error.log")
	errorData, err := os.ReadFile(errorPath)
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}
	if !contains(string(errorData), "Test error message") {
		t.Error("Expected error to be in error.log")
	}
}

// TestLogContext tests logging with context data.
func TestLogContext(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger with debug mode enabled
	logger, err := New(tmpDir, DEBUG, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if err := logger.SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode: %v", err)
	}

	// Log with context
	context := map[string]interface{}{
		"user_id":  123,
		"action":   "login",
		"duration": 1.5,
	}
	logger.Debug("User action", context)

	// Give file I/O time to complete
	time.Sleep(100 * time.Millisecond)

	// Read debug.log
	debugPath := filepath.Join(tmpDir, "debug.log")
	debugData, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("Failed to read debug.log: %v", err)
	}

	// Parse JSON entry
	var entry LogEntry
	if err := json.Unmarshal(debugData, &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify context data
	if entry.Context["user_id"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected user_id 123, got %v", entry.Context["user_id"])
	}
	if entry.Context["action"] != "login" {
		t.Errorf("Expected action 'login', got %v", entry.Context["action"])
	}
}

// TestGlobalDebugFunctions tests the global debug mode functions.
func TestGlobalDebugFunctions(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := New(tmpDir, WARN, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Set as global logger
	SetGlobal(logger)

	// Test global SetDebugMode
	if err := SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode via global function: %v", err)
	}

	// Test global IsDebugMode
	if !IsDebugMode() {
		t.Error("Expected global debug mode to be enabled")
	}

	// Disable debug mode
	if err := SetDebugMode(false); err != nil {
		t.Fatalf("Failed to disable debug mode via global function: %v", err)
	}

	if IsDebugMode() {
		t.Error("Expected global debug mode to be disabled")
	}
}

// TestFileRotationWithDebugMode tests that debug and error files are properly managed.
func TestFileRotationWithDebugMode(t *testing.T) {
	// Create temporary log directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := New(tmpDir, WARN, 7)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Enable debug mode
	if err := logger.SetDebugMode(true); err != nil {
		t.Fatalf("Failed to enable debug mode: %v", err)
	}

	// Log some messages
	logger.Debug("Debug 1")
	logger.Error("Error 1")

	// Close logger
	if err := logger.Close(); err != nil {
		t.Fatalf("Failed to close logger: %v", err)
	}

	// Give file I/O time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify both files exist and have content
	debugPath := filepath.Join(tmpDir, "debug.log")
	debugInfo, err := os.Stat(debugPath)
	if err != nil {
		t.Fatalf("debug.log not found: %v", err)
	}
	if debugInfo.Size() == 0 {
		t.Error("Expected debug.log to have content")
	}

	errorPath := filepath.Join(tmpDir, "error.log")
	errorInfo, err := os.Stat(errorPath)
	if err != nil {
		t.Fatalf("error.log not found: %v", err)
	}
	if errorInfo.Size() == 0 {
		t.Error("Expected error.log to have content")
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
		anySubstring(s, substr)))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
