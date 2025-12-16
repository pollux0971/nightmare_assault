package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// HTTPError 表示 HTTP 錯誤
//
// 用於封裝 HTTP 請求失敗時的狀態碼和錯誤訊息。
// 根據 HTTP 狀態碼的不同，錯誤的可重試性也不同：
// - 5xx 錯誤（服務器錯誤）通常是可重試的
// - 4xx 錯誤（客戶端錯誤）通常不可重試，除了 429 (Too Many Requests)
type HTTPError struct {
	// StatusCode 是 HTTP 狀態碼
	StatusCode int

	// Message 是錯誤訊息
	Message string
}

// Error 實作標準的 error 接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// ClassifyError 將錯誤分類為 AgentError
//
// 此函數分析錯誤的類型，並決定錯誤是否可重試。
// 分類邏輯如下：
//
// 1. 超時錯誤（context.DeadlineExceeded）→ 不可重試
//    超時通常表示請求已經等待足夠長的時間，重試不太可能成功
//
// 2. 取消錯誤（context.Canceled）→ 不可重試
//    取消是主動操作，重試沒有意義
//
// 3. 網路錯誤（net.Error）→ 可重試
//    網路問題可能是暫時性的，重試有機會成功
//
// 4. HTTP 錯誤：
//    - 5xx（服務器錯誤）→ 可重試
//      服務器問題可能是暫時性的
//    - 429（Too Many Requests）→ 可重試
//      速率限制通常在等待後可以重試
//    - 4xx（客戶端錯誤）→ 不可重試
//      請求本身有問題，重試不會改變結果
//
// 5. JSON 解析錯誤（json.SyntaxError）→ 不可重試
//    解析錯誤通常是數據格式問題，重試不會改變結果
//
// 6. 未知錯誤 → 可重試（默認）
//    對於未知錯誤類型，採取保守策略允許重試
//
// 參數：
//   - agentName: 發生錯誤的 Agent 名稱
//   - operation: 發生錯誤的操作名稱（如 "Invoke", "BuildPrompt" 等）
//   - err: 原始錯誤
//
// 返回：
//   - *AgentError: 包含錯誤上下文和可重試標記的 AgentError
func ClassifyError(agentName, operation string, err error) *AgentError {
	agentErr := &AgentError{
		AgentName: agentName,
		Operation: operation,
		Cause:     err,
	}

	// 1. 超時錯誤（不可重試）
	if errors.Is(err, context.DeadlineExceeded) {
		agentErr.Retryable = false
		return agentErr
	}

	// 2. 取消錯誤（不可重試）
	if errors.Is(err, context.Canceled) {
		agentErr.Retryable = false
		return agentErr
	}

	// 3. 網路錯誤（可重試）
	var netErr net.Error
	if errors.As(err, &netErr) {
		agentErr.Retryable = true
		return agentErr
	}

	// 4. HTTP 錯誤（5xx 和 429 可重試，其他 4xx 不可重試）
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		// 5xx 錯誤或 429 錯誤可重試
		if (httpErr.StatusCode >= 500 && httpErr.StatusCode < 600) || httpErr.StatusCode == 429 {
			agentErr.Retryable = true
		} else {
			agentErr.Retryable = false
		}
		return agentErr
	}

	// 5. JSON 解析錯誤（不可重試）
	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		agentErr.Retryable = false
		return agentErr
	}

	// 6. 默認：未知錯誤可重試
	agentErr.Retryable = true
	return agentErr
}
