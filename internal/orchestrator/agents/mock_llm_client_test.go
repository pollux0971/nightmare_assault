package agents

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestMockLLMClientSuccess 測試成功場景
func TestMockLLMClientSuccess(t *testing.T) {
	mock := NewMockLLMClient("test response")

	result, err := mock.Generate(context.Background(), "test prompt", nil)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "test response" {
		t.Errorf("Expected 'test response', got '%s'", result)
	}

	if mock.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.CallCount)
	}

	if mock.LastPrompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", mock.LastPrompt)
	}
}

// TestMockLLMClientError 測試錯誤場景
func TestMockLLMClientError(t *testing.T) {
	testErr := errors.New("test error")
	mock := NewMockLLMClientWithError(testErr)

	result, err := mock.Generate(context.Background(), "test prompt", nil)

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}

	if mock.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.CallCount)
	}
}

// TestMockLLMClientDelay 測試延遲場景
func TestMockLLMClientDelay(t *testing.T) {
	mock := NewMockLLMClientWithDelay("delayed response", 50*time.Millisecond)

	start := time.Now()
	result, err := mock.Generate(context.Background(), "test prompt", nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "delayed response" {
		t.Errorf("Expected 'delayed response', got '%s'", result)
	}

	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected delay >= 50ms, got %v", elapsed)
	}
}

// TestMockLLMClientTimeout 測試超時場景
func TestMockLLMClientTimeout(t *testing.T) {
	mock := NewMockLLMClientWithDelay("should not return", 200*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := mock.Generate(ctx, "test prompt", nil)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result on timeout, got '%s'", result)
	}
}

// TestMockLLMClientCustomFunc 測試自定義函數
func TestMockLLMClientCustomFunc(t *testing.T) {
	callCount := 0
	mock := NewMockLLMClientWithFunc(func(ctx context.Context, prompt string, options map[string]any) (string, error) {
		callCount++
		if callCount == 1 {
			return "", errors.New("first call fails")
		}
		return "second call succeeds", nil
	})

	// 第一次調用失敗
	_, err := mock.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Error("Expected error on first call")
	}

	// 第二次調用成功
	result, err := mock.Generate(context.Background(), "prompt", nil)
	if err != nil {
		t.Errorf("Expected no error on second call, got %v", err)
	}

	if result != "second call succeeds" {
		t.Errorf("Expected 'second call succeeds', got '%s'", result)
	}

	if mock.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", mock.CallCount)
	}
}

// TestNewRetryMockLLMClient 測試重試場景
func TestNewRetryMockLLMClient(t *testing.T) {
	mock := NewRetryMockLLMClient(
		2,                  // 前兩次失敗
		CommonErrors.HTTP500, // 返回 500 錯誤
		"success after retry",
	)

	// 第一次調用失敗
	_, err := mock.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Error("Expected error on first call")
	}

	// 第二次調用失敗
	_, err = mock.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Error("Expected error on second call")
	}

	// 第三次調用成功
	result, err := mock.Generate(context.Background(), "prompt", nil)
	if err != nil {
		t.Errorf("Expected no error on third call, got %v", err)
	}

	if result != "success after retry" {
		t.Errorf("Expected 'success after retry', got '%s'", result)
	}
}

// TestMockLLMClientReset 測試 Reset 方法
func TestMockLLMClientReset(t *testing.T) {
	mock := NewMockLLMClient("test response")

	// 調用一次
	mock.Generate(context.Background(), "first prompt", map[string]any{"key": "value"})

	if mock.CallCount != 1 {
		t.Errorf("Expected 1 call before reset, got %d", mock.CallCount)
	}

	// 重置
	mock.Reset()

	if mock.CallCount != 0 {
		t.Errorf("Expected 0 calls after reset, got %d", mock.CallCount)
	}

	if mock.LastPrompt != "" {
		t.Errorf("Expected empty prompt after reset, got '%s'", mock.LastPrompt)
	}

	if mock.LastOptions != nil {
		t.Errorf("Expected nil options after reset, got %v", mock.LastOptions)
	}

	// Response 應該保留
	if mock.Response != "test response" {
		t.Errorf("Expected Response to be preserved after reset, got '%s'", mock.Response)
	}
}

// TestMockLLMClientOptions 測試記錄 options
func TestMockLLMClientOptions(t *testing.T) {
	mock := NewMockLLMClient("test response")

	options := map[string]any{
		"temperature": 0.7,
		"max_tokens":  100,
	}

	mock.Generate(context.Background(), "prompt", options)

	if mock.LastOptions == nil {
		t.Fatal("Expected LastOptions to be recorded")
	}

	if mock.LastOptions["temperature"] != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", mock.LastOptions["temperature"])
	}

	if mock.LastOptions["max_tokens"] != 100 {
		t.Errorf("Expected max_tokens 100, got %v", mock.LastOptions["max_tokens"])
	}
}

// TestCommonErrors 測試 CommonErrors 常量
func TestCommonErrors(t *testing.T) {
	if CommonErrors.HTTP500 == nil {
		t.Error("Expected HTTP500 to be defined")
	}

	if CommonErrors.HTTP429 == nil {
		t.Error("Expected HTTP429 to be defined")
	}

	// 驗證 HTTP500 是 HTTPError 類型
	var httpErr *HTTPError
	if !errors.As(CommonErrors.HTTP500, &httpErr) {
		t.Error("Expected HTTP500 to be HTTPError type")
	}

	if httpErr.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", httpErr.StatusCode)
	}
}
