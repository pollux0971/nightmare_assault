package agents

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestBaseAgentImplCreation 測試 BaseAgentImpl 創建
func TestBaseAgentImplCreation(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	if impl.Config.Name != "TestAgent" {
		t.Errorf("Expected name 'TestAgent', got %s", impl.Config.Name)
	}

	if impl.Config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", impl.Config.Timeout)
	}

	if impl.Config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", impl.Config.MaxRetries)
	}
}

// TestInvokeWithRetry_Success 測試成功的調用（不需重試）
func TestInvokeWithRetry_Success(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	invokeFn := func(ctx context.Context) (any, error) {
		return "success", nil
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}
}

// TestInvokeWithRetry_RetryableError 測試可重試錯誤（應該重試）
func TestInvokeWithRetry_RetryableError(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	attemptCount := 0
	invokeFn := func(ctx context.Context) (any, error) {
		attemptCount++
		if attemptCount < 3 {
			// 前兩次返回可重試錯誤
			return nil, &HTTPError{StatusCode: 500, Message: "Internal Server Error"}
		}
		// 第三次成功
		return "success after retry", nil
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}

	if result != "success after retry" {
		t.Errorf("Expected result 'success after retry', got %v", result)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

// TestInvokeWithRetry_NonRetryableError 測試不可重試錯誤（不應重試）
func TestInvokeWithRetry_NonRetryableError(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	attemptCount := 0
	invokeFn := func(ctx context.Context) (any, error) {
		attemptCount++
		// 返回不可重試錯誤（HTTP 400）
		return nil, &HTTPError{StatusCode: 400, Message: "Bad Request"}
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	// 不可重試錯誤應該只嘗試一次
	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt (no retry), got %d", attemptCount)
	}

	// 驗證錯誤是 AgentError
	var agentErr *AgentError
	if !errors.As(err, &agentErr) {
		t.Error("Expected AgentError type")
	}

	if agentErr.Retryable {
		t.Error("Expected non-retryable error")
	}
}

// TestInvokeWithRetry_MaxRetriesExceeded 測試超過最大重試次數
func TestInvokeWithRetry_MaxRetriesExceeded(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	attemptCount := 0
	invokeFn := func(ctx context.Context) (any, error) {
		attemptCount++
		// 始終返回可重試錯誤
		return nil, &HTTPError{StatusCode: 503, Message: "Service Unavailable"}
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	// 應該嘗試 MaxRetries 次
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

// TestInvokeWithRetry_Timeout 測試超時控制
func TestInvokeWithRetry_Timeout(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    100 * time.Millisecond, // 短超時時間
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	invokeFn := func(ctx context.Context) (any, error) {
		// 模擬長時間操作
		select {
		case <-time.After(200 * time.Millisecond):
			return "should not reach here", nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	// 超時錯誤應該是不可重試的
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		if agentErr.Retryable {
			t.Error("Expected timeout error to be non-retryable")
		}
	}

	// 驗證底層錯誤是 DeadlineExceeded
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("Expected context.DeadlineExceeded error")
	}
}

// TestInvokeWithRetry_ContextCanceled 測試上下文取消
func TestInvokeWithRetry_ContextCanceled(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	// 創建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	invokeFn := func(ctx context.Context) (any, error) {
		return nil, ctx.Err()
	}

	result, err := impl.InvokeWithRetry(ctx, invokeFn)

	if err == nil {
		t.Error("Expected canceled error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	// 取消錯誤應該是不可重試的
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		if agentErr.Retryable {
			t.Error("Expected canceled error to be non-retryable")
		}
	}

	// 驗證底層錯誤是 Canceled
	if !errors.Is(err, context.Canceled) {
		t.Error("Expected context.Canceled error")
	}
}

// TestInvokeWithRetry_ExponentialBackoff 測試指數退避
func TestInvokeWithRetry_ExponentialBackoff(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	impl := NewBaseAgentImpl(config)

	attemptTimes := []time.Time{}
	invokeFn := func(ctx context.Context) (any, error) {
		attemptTimes = append(attemptTimes, time.Now())
		if len(attemptTimes) < 3 {
			// 前兩次返回可重試錯誤
			return nil, &HTTPError{StatusCode: 500, Message: "Internal Server Error"}
		}
		// 第三次成功
		return "success", nil
	}

	_, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}

	// 驗證至少有 3 次嘗試
	if len(attemptTimes) != 3 {
		t.Fatalf("Expected 3 attempts, got %d", len(attemptTimes))
	}

	// 驗證第二次嘗試距離第一次至少 1 秒（第一次退避）
	backoff1 := attemptTimes[1].Sub(attemptTimes[0])
	if backoff1 < 1*time.Second {
		t.Errorf("Expected first backoff >= 1s, got %v", backoff1)
	}

	// 驗證第三次嘗試距離第二次至少 2 秒（第二次退避）
	backoff2 := attemptTimes[2].Sub(attemptTimes[1])
	if backoff2 < 2*time.Second {
		t.Errorf("Expected second backoff >= 2s, got %v", backoff2)
	}
}
