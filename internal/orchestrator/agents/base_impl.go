package agents

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// BaseAgentImpl 是 BaseAgent 的通用實現
//
// 提供了所有 Agents 共用的核心功能：
// - 超時控制：使用 context.WithTimeout 限制每次調用的時間
// - 重試機制：自動重試可重試的錯誤，最多重試 MaxRetries 次
// - 指數退避：每次重試之間的等待時間呈指數增長（1s, 2s, 4s, ...）
// - 錯誤分類：自動分類錯誤並決定是否重試
//
// 使用方式：
//
//	config := AgentConfig{
//	    Name: "MyAgent",
//	    Timeout: 30 * time.Second,
//	    MaxRetries: 3,
//	}
//	impl := NewBaseAgentImpl(config)
//	result, err := impl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
//	    // 執行實際的 Agent 邏輯
//	    return doSomething(ctx)
//	})
type BaseAgentImpl struct {
	// Config 是 Agent 的配置
	Config AgentConfig
}

// NewBaseAgentImpl 創建一個新的 BaseAgentImpl 實例
//
// 參數：
//   - config: Agent 配置，包含名稱、超時時間、最大重試次數等
//
// 返回：
//   - *BaseAgentImpl: 新創建的 BaseAgentImpl 實例
func NewBaseAgentImpl(config AgentConfig) *BaseAgentImpl {
	return &BaseAgentImpl{
		Config: config,
	}
}

// InvokeWithRetry 執行 Agent 調用並自動處理重試邏輯
//
// 此方法實作了完整的重試機制，包括：
// 1. 超時控制：為每次調用設置超時時間
// 2. 錯誤分類：使用 ClassifyError 判斷錯誤是否可重試
// 3. 重試循環：對可重試錯誤自動重試，最多 MaxRetries 次
// 4. 指數退避：每次重試之間等待時間呈指數增長（2^attempt 秒）
// 5. 日誌記錄：記錄每次重試的詳細信息
//
// 重試策略：
// - 不可重試錯誤（如超時、取消、4xx HTTP 錯誤）：立即返回，不重試
// - 可重試錯誤（如網路錯誤、5xx HTTP 錯誤）：重試直到成功或達到 MaxRetries
// - 上下文取消：立即停止重試並返回
//
// 參數：
//   - ctx: 上下文，用於控制整個調用鏈的生命週期
//   - invokeFn: 實際執行的函數，接收 context 並返回結果和錯誤
//
// 返回：
//   - any: 調用成功時的結果
//   - error: 調用失敗時的錯誤（包裝為 AgentError）
func (impl *BaseAgentImpl) InvokeWithRetry(
	ctx context.Context,
	invokeFn func(context.Context) (any, error),
) (any, error) {
	var lastErr error

	for attempt := 0; attempt < impl.Config.MaxRetries; attempt++ {
		// 為每次嘗試設定超時
		ctxWithTimeout, cancel := context.WithTimeout(ctx, impl.Config.Timeout)
		defer cancel()

		// 執行調用
		result, err := invokeFn(ctxWithTimeout)
		if err == nil {
			// 調用成功，返回結果
			return result, nil
		}

		// 錯誤分類
		agentErr := ClassifyError(impl.Config.Name, "Invoke", err)
		lastErr = agentErr

		// 非重試錯誤，直接返回
		if !agentErr.IsRetryable() {
			log.Printf("[%s] Non-retryable error on attempt %d/%d: %v",
				impl.Config.Name, attempt+1, impl.Config.MaxRetries, err)
			return nil, agentErr
		}

		// 記錄重試
		log.Printf("[%s] Attempt %d/%d failed: %v, retrying...",
			impl.Config.Name, attempt+1, impl.Config.MaxRetries, err)

		// 如果不是最後一次嘗試，執行指數退避
		if attempt < impl.Config.MaxRetries-1 {
			// 計算退避時間：2^attempt 秒
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second

			// 等待退避時間或上下文取消
			select {
			case <-time.After(backoff):
				// 繼續下一次重試
				continue
			case <-ctx.Done():
				// 上下文已取消，停止重試
				cancelErr := ClassifyError(impl.Config.Name, "Invoke", ctx.Err())
				return nil, cancelErr
			}
		}
	}

	// 達到最大重試次數
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
