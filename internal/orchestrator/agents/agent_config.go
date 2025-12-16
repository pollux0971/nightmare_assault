package agents

import (
	"context"
	"errors"
	"time"
)

// AgentConfig 是 Agent 的配置結構
//
// 包含 Agent 運行所需的所有配置參數，包括超時設定、重試策略、
// 以及依賴注入的 LLM 客戶端和 Prompt 加載器。
type AgentConfig struct {
	// Name 是 Agent 的名稱，用於日誌和錯誤追蹤
	Name string

	// Timeout 是單次 LLM 調用的超時時間
	// 默認值：30 秒
	Timeout time.Duration

	// MaxRetries 是失敗後的最大重試次數
	// 默認值：3 次
	MaxRetries int

	// LLMClient 是 LLM 客戶端接口，用於與不同的 LLM Provider 通信
	LLMClient LLMClient

	// PromptLoader 是 Prompt 模板加載器，用於加載和渲染 prompt 模板
	PromptLoader PromptLoader
}

// DefaultAgentConfig 返回默認的 Agent 配置
//
// 默認配置：
//   - Timeout: 30 秒
//   - MaxRetries: 3 次
//   - Name: 空字串（需要由具體 Agent 設定）
//   - LLMClient: nil（需要由具體 Agent 注入）
//   - PromptLoader: nil（需要由具體 Agent 注入）
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// ValidateConfig 驗證 Agent 配置的有效性
//
// 檢查項目：
//   - Name 不能為空
//   - Timeout 必須為正數
//   - MaxRetries 不能為負數
//
// 返回：
//   - error: 驗證失敗時返回錯誤，成功時返回 nil
func ValidateConfig(config AgentConfig) error {
	if config.Name == "" {
		return errors.New("agent name cannot be empty")
	}

	if config.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}

	return nil
}

// LLMClient 是 LLM 客戶端的接口
//
// 支援多個 LLM Provider（OpenAI, Anthropic, Gemini, Cohere 等）
// 的統一調用接口。
type LLMClient interface {
	// Generate 發送 prompt 到 LLM 並獲取響應
	//
	// 參數:
	//   - ctx: 上下文，用於超時控制
	//   - prompt: 輸入的 prompt 字串
	//   - options: LLM 調用選項（如 temperature, max_tokens 等）
	//
	// 返回:
	//   - string: LLM 的原始響應
	//   - error: 錯誤信息
	Generate(ctx context.Context, prompt string, options map[string]any) (string, error)
}

// PromptLoader 是 Prompt 模板加載器的接口
//
// 用於從文件系統或其他來源加載 prompt 模板，並使用數據渲染模板。
type PromptLoader interface {
	// LoadTemplate 從指定路徑加載 prompt 模板
	//
	// 參數:
	//   - name: 模板名稱或路徑
	//
	// 返回:
	//   - string: 模板內容
	//   - error: 錯誤信息
	LoadTemplate(name string) (string, error)

	// RenderTemplate 使用數據渲染 prompt 模板
	//
	// 參數:
	//   - template: 模板字串
	//   - data: 用於填充模板的數據
	//
	// 返回:
	//   - string: 渲染後的 prompt
	//   - error: 錯誤信息
	RenderTemplate(template string, data any) (string, error)
}
