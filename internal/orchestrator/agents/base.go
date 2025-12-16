package agents

import (
	"context"
	"time"
)

// BaseAgent 是所有 Agent 的基礎接口
//
// 所有 Agent 必須實作此接口以確保一致的行為和錯誤處理。
// 這個接口提供了統一的調用方式、超時控制和生命週期管理。
type BaseAgent interface {
	// Invoke 是 Agent 的核心方法，處理輸入請求並返回結果
	//
	// 參數:
	//   - ctx: 上下文，用於超時控制和取消操作
	//   - request: 輸入請求（類型由具體 Agent 定義）
	//
	// 返回:
	//   - any: 處理結果（類型由具體 Agent 定義）
	//   - error: 錯誤信息（如果發生）
	Invoke(ctx context.Context, request any) (any, error)

	// GetName 返回 Agent 的名稱
	//
	// 用於日誌記錄和錯誤追蹤，每個 Agent 應該有唯一的名稱。
	GetName() string

	// GetTimeout 返回 Agent 的超時時間
	//
	// 用於設定 LLM 調用的最大等待時間，默認建議 30 秒。
	GetTimeout() time.Duration

	// BuildPrompt 構建 LLM prompt（可選方法）
	//
	// 由具體 Agent 實作以生成特定的 prompt 格式。
	// 如果 Agent 不需要此功能，可以返回空字串或 nil。
	//
	// 參數:
	//   - request: 輸入請求
	//
	// 返回:
	//   - string: 生成的 prompt
	//   - error: 錯誤信息
	BuildPrompt(request any) (string, error)

	// ParseResponse 解析 LLM 響應（可選方法）
	//
	// 由具體 Agent 實作以將 LLM 的原始響應轉換為結構化數據。
	// 如果 Agent 不需要此功能，可以直接返回原始字串。
	//
	// 參數:
	//   - raw: LLM 的原始響應字串
	//
	// 返回:
	//   - any: 解析後的結構化數據
	//   - error: 錯誤信息
	ParseResponse(raw string) (any, error)
}
