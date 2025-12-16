package agents

import (
	"context"
	"errors"
	"time"
)

// MockLLMClient 是用於測試的可配置 LLM 客戶端模擬
//
// 支援以下測試場景：
// - 模擬成功回應（設定 Response）
// - 模擬延遲（設定 Delay，可用於測試超時）
// - 模擬錯誤（設定 Error）
// - 記錄調用次數（CallCount）
// - 記錄調用參數（LastPrompt, LastOptions）
//
// 使用範例：
//
//	// 成功場景
//	mock := &MockLLMClient{
//	    Response: "test response",
//	}
//
//	// 超時場景
//	mock := &MockLLMClient{
//	    Delay: 200 * time.Millisecond, // 超過超時時間
//	}
//
//	// 錯誤場景
//	mock := &MockLLMClient{
//	    Error: &HTTPError{StatusCode: 500, Message: "Internal Server Error"},
//	}
type MockLLMClient struct {
	// Response 是模擬的成功回應內容
	Response string

	// Error 是模擬的錯誤（如果設置，Generate 將返回此錯誤）
	Error error

	// Delay 是模擬的處理延遲（用於測試超時場景）
	Delay time.Duration

	// CallCount 記錄 Generate 被調用的次數
	CallCount int

	// LastPrompt 記錄最後一次調用的 prompt 參數
	LastPrompt string

	// LastOptions 記錄最後一次調用的 options 參數
	LastOptions map[string]any

	// ResponseFunc 是自定義回應函數（如果設置，優先於 Response 和 Error）
	// 可用於實現複雜的測試場景（如前幾次失敗，最後一次成功）
	ResponseFunc func(ctx context.Context, prompt string, options map[string]any) (string, error)
}

// Generate 模擬 LLM 客戶端的 Generate 方法
//
// 執行順序：
// 1. 記錄調用次數和參數
// 2. 如果設置了 Delay，等待指定時間（或直到上下文取消）
// 3. 如果設置了 ResponseFunc，調用自定義函數
// 4. 否則，如果設置了 Error，返回錯誤
// 5. 否則，返回 Response
//
// 參數：
//   - ctx: 上下文（用於超時和取消）
//   - prompt: 提示詞
//   - options: 可選參數
//
// 返回：
//   - string: 模擬的回應內容
//   - error: 模擬的錯誤
func (m *MockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	// 記錄調用
	m.CallCount++
	m.LastPrompt = prompt
	m.LastOptions = options

	// 模擬延遲
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
			// 延遲結束，繼續執行
		case <-ctx.Done():
			// 上下文已取消（可能是超時）
			return "", ctx.Err()
		}
	}

	// 使用自定義回應函數
	if m.ResponseFunc != nil {
		return m.ResponseFunc(ctx, prompt, options)
	}

	// 返回設定的錯誤
	if m.Error != nil {
		return "", m.Error
	}

	// 返回設定的回應
	return m.Response, nil
}

// Reset 重置 Mock 的狀態
//
// 清除 CallCount、LastPrompt 和 LastOptions，
// 但保留 Response、Error、Delay 和 ResponseFunc 的設定。
func (m *MockLLMClient) Reset() {
	m.CallCount = 0
	m.LastPrompt = ""
	m.LastOptions = nil
}

// NewMockLLMClient 創建一個新的 MockLLMClient
//
// 參數：
//   - response: 預設回應內容
//
// 返回：
//   - *MockLLMClient: 新創建的 MockLLMClient
func NewMockLLMClient(response string) *MockLLMClient {
	return &MockLLMClient{
		Response: response,
	}
}

// NewMockLLMClientWithError 創建一個返回錯誤的 MockLLMClient
//
// 參數：
//   - err: 要返回的錯誤
//
// 返回：
//   - *MockLLMClient: 新創建的 MockLLMClient
func NewMockLLMClientWithError(err error) *MockLLMClient {
	return &MockLLMClient{
		Error: err,
	}
}

// NewMockLLMClientWithDelay 創建一個帶延遲的 MockLLMClient
//
// 參數：
//   - response: 預設回應內容
//   - delay: 延遲時間
//
// 返回：
//   - *MockLLMClient: 新創建的 MockLLMClient
func NewMockLLMClientWithDelay(response string, delay time.Duration) *MockLLMClient {
	return &MockLLMClient{
		Response: response,
		Delay:    delay,
	}
}

// NewMockLLMClientWithFunc 創建一個使用自定義函數的 MockLLMClient
//
// 參數：
//   - fn: 自定義回應函數
//
// 返回：
//   - *MockLLMClient: 新創建的 MockLLMClient
func NewMockLLMClientWithFunc(fn func(ctx context.Context, prompt string, options map[string]any) (string, error)) *MockLLMClient {
	return &MockLLMClient{
		ResponseFunc: fn,
	}
}

// NewRetryMockLLMClient 創建一個用於測試重試的 MockLLMClient
//
// 前 failCount 次調用返回 err，之後返回 successResponse。
//
// 參數：
//   - failCount: 失敗次數
//   - err: 失敗時返回的錯誤
//   - successResponse: 成功時返回的回應
//
// 返回：
//   - *MockLLMClient: 新創建的 MockLLMClient
func NewRetryMockLLMClient(failCount int, err error, successResponse string) *MockLLMClient {
	callCount := 0
	return &MockLLMClient{
		ResponseFunc: func(ctx context.Context, prompt string, options map[string]any) (string, error) {
			callCount++
			if callCount <= failCount {
				return "", err
			}
			return successResponse, nil
		},
	}
}

// CommonErrors 提供常見的測試錯誤
var CommonErrors = struct {
	Timeout         error
	Canceled        error
	NetworkError    error
	HTTP500         error
	HTTP503         error
	HTTP400         error
	HTTP429         error
	JSONParseError  error
}{
	Timeout:         errors.New("timeout error"),
	Canceled:        errors.New("canceled error"),
	NetworkError:    errors.New("network connection failed"),
	HTTP500:         &HTTPError{StatusCode: 500, Message: "Internal Server Error"},
	HTTP503:         &HTTPError{StatusCode: 503, Message: "Service Unavailable"},
	HTTP400:         &HTTPError{StatusCode: 400, Message: "Bad Request"},
	HTTP429:         &HTTPError{StatusCode: 429, Message: "Too Many Requests"},
	JSONParseError:  errors.New("json parse error"),
}
