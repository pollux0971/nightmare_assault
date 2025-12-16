package agents

import (
	"errors"
	"testing"
)

// TestAgentErrorCreation 測試 AgentError 創建
func TestAgentErrorCreation(t *testing.T) {
	causeErr := errors.New("network timeout")
	agentErr := &AgentError{
		AgentName: "TestAgent",
		Operation: "Invoke",
		Cause:     causeErr,
		Retryable: true,
	}

	if agentErr.AgentName != "TestAgent" {
		t.Errorf("Expected agent name 'TestAgent', got %s", agentErr.AgentName)
	}

	if agentErr.Operation != "Invoke" {
		t.Errorf("Expected operation 'Invoke', got %s", agentErr.Operation)
	}

	if agentErr.Cause != causeErr {
		t.Errorf("Expected cause to be original error")
	}

	if !agentErr.Retryable {
		t.Error("Expected error to be retryable")
	}
}

// TestAgentErrorError 測試 Error() 方法
func TestAgentErrorError(t *testing.T) {
	causeErr := errors.New("connection refused")
	agentErr := &AgentError{
		AgentName: "NarrationAgent",
		Operation: "BuildPrompt",
		Cause:     causeErr,
		Retryable: false,
	}

	expected := "[NarrationAgent] BuildPrompt failed: connection refused (retryable=false)"
	if agentErr.Error() != expected {
		t.Errorf("Expected error message:\n%s\ngot:\n%s", expected, agentErr.Error())
	}
}

// TestAgentErrorUnwrap 測試 Unwrap() 方法
func TestAgentErrorUnwrap(t *testing.T) {
	causeErr := errors.New("original error")
	agentErr := &AgentError{
		AgentName: "TestAgent",
		Operation: "Invoke",
		Cause:     causeErr,
		Retryable: true,
	}

	unwrapped := agentErr.Unwrap()
	if unwrapped != causeErr {
		t.Error("Unwrap() should return the original cause error")
	}

	// 測試 errors.Is() 可以使用
	if !errors.Is(agentErr, causeErr) {
		t.Error("errors.Is() should work with AgentError")
	}
}

// TestAgentErrorIsRetryable 測試 IsRetryable() 方法
func TestAgentErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		retryable bool
	}{
		{"retryable error", true},
		{"non-retryable error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentErr := &AgentError{
				AgentName: "TestAgent",
				Operation: "Test",
				Cause:     errors.New("test error"),
				Retryable: tt.retryable,
			}

			if agentErr.IsRetryable() != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", agentErr.IsRetryable(), tt.retryable)
			}
		})
	}
}

// TestAgentErrorChaining 測試錯誤鏈
func TestAgentErrorChaining(t *testing.T) {
	// 創建錯誤鏈
	originalErr := errors.New("database connection failed")
	wrappedErr := &AgentError{
		AgentName: "SeedAgent",
		Operation: "LoadSeeds",
		Cause:     originalErr,
		Retryable: true,
	}

	// 測試可以透過 errors.Is 找到原始錯誤
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Should be able to find original error in chain")
	}

	// 測試可以透過 errors.As 找到 AgentError
	var agentErr *AgentError
	if !errors.As(wrappedErr, &agentErr) {
		t.Error("Should be able to extract AgentError from chain")
	}

	if agentErr.AgentName != "SeedAgent" {
		t.Errorf("Expected agent name 'SeedAgent', got %s", agentErr.AgentName)
	}
}
