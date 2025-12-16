package agents

import (
	"fmt"
)

// AgentError 是 Agent 層的統一錯誤類型
//
// 封裝了錯誤的上下文信息（Agent 名稱、操作類型）以及是否可重試的標記。
// 實作了標準的 error 接口以及 Unwrap() 方法以支援錯誤鏈。
type AgentError struct {
	// AgentName 是發生錯誤的 Agent 名稱
	AgentName string

	// Operation 是發生錯誤的操作名稱（如 "Invoke", "BuildPrompt" 等）
	Operation string

	// Cause 是底層的原始錯誤
	Cause error

	// Retryable 標記此錯誤是否可以重試
	// true: 可以重試（如網路錯誤、5xx HTTP 錯誤）
	// false: 不可重試（如參數錯誤、4xx HTTP 錯誤、超時）
	Retryable bool
}

// Error 實作標準的 error 接口
//
// 返回格式化的錯誤訊息：
// [AgentName] Operation failed: Cause (retryable=true/false)
func (e *AgentError) Error() string {
	return fmt.Sprintf("[%s] %s failed: %v (retryable=%t)",
		e.AgentName, e.Operation, e.Cause, e.Retryable)
}

// Unwrap 返回底層的原始錯誤
//
// 實作此方法使 AgentError 支援 Go 1.13+ 的錯誤鏈機制，
// 允許使用 errors.Is() 和 errors.As() 來檢查和提取錯誤。
func (e *AgentError) Unwrap() error {
	return e.Cause
}

// IsRetryable 返回錯誤是否可以重試
//
// 這是一個便利方法，直接返回 Retryable 字段的值。
// 可用於在重試邏輯中快速判斷是否應該重試。
func (e *AgentError) IsRetryable() bool {
	return e.Retryable
}
