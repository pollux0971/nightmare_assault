package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

// TestClassifyError_Timeout 測試超時錯誤分類
func TestClassifyError_Timeout(t *testing.T) {
	err := context.DeadlineExceeded
	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if agentErr.AgentName != "TestAgent" {
		t.Errorf("Expected agent name 'TestAgent', got %s", agentErr.AgentName)
	}

	if agentErr.Operation != "Invoke" {
		t.Errorf("Expected operation 'Invoke', got %s", agentErr.Operation)
	}

	if agentErr.Retryable {
		t.Error("Timeout errors should NOT be retryable")
	}

	if !errors.Is(agentErr, context.DeadlineExceeded) {
		t.Error("Should preserve error chain")
	}
}

// TestClassifyError_Canceled 測試取消錯誤分類
func TestClassifyError_Canceled(t *testing.T) {
	err := context.Canceled
	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if agentErr.Retryable {
		t.Error("Canceled errors should NOT be retryable")
	}
}

// TestClassifyError_NetworkError 測試網路錯誤分類
func TestClassifyError_NetworkError(t *testing.T) {
	// 模擬網路錯誤
	_, err := net.DialTimeout("tcp", "invalid-host:99999", 1*time.Millisecond)
	if err == nil {
		t.Fatal("Expected network error for invalid host")
	}

	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if !agentErr.Retryable {
		t.Error("Network errors should be retryable")
	}
}

// TestClassifyError_HTTPError5xx 測試 HTTP 5xx 錯誤
func TestClassifyError_HTTPError5xx(t *testing.T) {
	err := &HTTPError{
		StatusCode: 500,
		Message:    "Internal Server Error",
	}

	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if !agentErr.Retryable {
		t.Error("HTTP 5xx errors should be retryable")
	}
}

// TestClassifyError_HTTPError4xx 測試 HTTP 4xx 錯誤
func TestClassifyError_HTTPError4xx(t *testing.T) {
	err := &HTTPError{
		StatusCode: 400,
		Message:    "Bad Request",
	}

	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if agentErr.Retryable {
		t.Error("HTTP 4xx errors should NOT be retryable")
	}
}

// TestClassifyError_JSONParseError 測試 JSON 解析錯誤
func TestClassifyError_JSONParseError(t *testing.T) {
	var data map[string]any
	err := json.Unmarshal([]byte("invalid json"), &data)
	if err == nil {
		t.Fatal("Expected JSON parse error")
	}

	agentErr := ClassifyError("TestAgent", "ParseResponse", err)

	if agentErr.Retryable {
		t.Error("JSON parse errors should NOT be retryable")
	}
}

// TestClassifyError_UnknownError 測試未知錯誤（默認行為）
func TestClassifyError_UnknownError(t *testing.T) {
	err := errors.New("unknown error type")

	agentErr := ClassifyError("TestAgent", "Invoke", err)

	if !agentErr.Retryable {
		t.Error("Unknown errors should be retryable by default")
	}
}

// TestClassifyError_AllHTTPStatusCodes 測試所有 HTTP 狀態碼
func TestClassifyError_AllHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		statusCode int
		retryable  bool
	}{
		{400, false}, // Bad Request
		{401, false}, // Unauthorized
		{403, false}, // Forbidden
		{404, false}, // Not Found
		{429, true},  // Too Many Requests (可重試)
		{500, true},  // Internal Server Error
		{502, true},  // Bad Gateway
		{503, true},  // Service Unavailable
		{504, true},  // Gateway Timeout
	}

	for _, tt := range tests {
		t.Run(formatStatusCode(tt.statusCode), func(t *testing.T) {
			err := &HTTPError{
				StatusCode: tt.statusCode,
				Message:    "test error",
			}

			agentErr := ClassifyError("TestAgent", "Invoke", err)

			if agentErr.Retryable != tt.retryable {
				t.Errorf("HTTP %d: expected retryable=%v, got %v",
					tt.statusCode, tt.retryable, agentErr.Retryable)
			}
		})
	}
}

// formatStatusCode 格式化狀態碼為測試名稱
func formatStatusCode(code int) string {
	return fmt.Sprintf("HTTP_%d", code)
}
