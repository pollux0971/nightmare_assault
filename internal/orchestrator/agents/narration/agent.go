package narration

import (
	"context"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// NarrationMode 定義 Narration Agent 的工作模式
//
// Narration Agent 支援四種模式：
//   - Skeleton: 規劃故事骨架（世界觀、劇情結構、伏筆分佈、多重結局）
//   - Content: 生成章節內容（敘事文本、選項、NPC 對話）
//   - Opening: 生成開場（根據骨架生成遊戲開場）
//   - Ending: 生成結局（根據 Seeds 揭示度生成對應結局）
type NarrationMode int

const (
	// ModeSkeleton 骨架模式 - 規劃故事結構
	ModeSkeleton NarrationMode = iota

	// ModeContent 內容模式 - 生成章節敘事
	ModeContent

	// ModeOpening 開場模式 - 生成遊戲開場
	ModeOpening

	// ModeEnding 結局模式 - 生成結局
	ModeEnding
)

// String 返回 NarrationMode 的字串表示
func (m NarrationMode) String() string {
	switch m {
	case ModeSkeleton:
		return "Skeleton"
	case ModeContent:
		return "Content"
	case ModeOpening:
		return "Opening"
	case ModeEnding:
		return "Ending"
	default:
		return "Unknown"
	}
}

// NarrationAgent 是故事生成 Agent
//
// 負責遊戲中所有故事相關的生成任務：
//   - Phase 1 (Genesis): 使用 Skeleton 模式規劃故事骨架
//   - Phase 2 (Game Loop): 使用 Content 模式生成章節內容
//   - Phase 3 (Convergence): 使用 Opening/Ending 模式生成開場與結局
//
// NarrationAgent 繼承 BaseAgentImpl，復用超時控制、重試機制和錯誤處理。
type NarrationAgent struct {
	// Config 是 Agent 配置
	config agents.AgentConfig

	// BaseImpl 提供通用的 Agent 功能
	baseImpl *agents.BaseAgentImpl
}

// NewNarrationAgent 創建一個新的 NarrationAgent
//
// 參數：
//   - config: Agent 配置，包含 Name, Timeout, LLMClient 等
//
// 返回：
//   - *NarrationAgent: 新創建的 NarrationAgent 實例
func NewNarrationAgent(config agents.AgentConfig) *NarrationAgent {
	// 設置默認名稱
	if config.Name == "" {
		config.Name = "NarrationAgent"
	}

	// 設置默認超時
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// 設置默認重試次數
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &NarrationAgent{
		config:   config,
		baseImpl: agents.NewBaseAgentImpl(config),
	}
}

// Invoke 實現 BaseAgent 接口的主調用方法
//
// 根據 request 的類型分發到不同的模式：
//   - *SkeletonRequest → InvokeSkeleton
//   - *ContentRequest → InvokeContent
//   - *AutoContentRequest → InvokeAutoContent (Story 7-2)
//   - *OpeningRequest → InvokeOpening (未實現)
//   - *EndingRequest → InvokeEnding (未實現)
//
// 參數：
//   - ctx: 上下文，用於超時和取消控制
//   - request: 請求對象，可以是 Skeleton/Content/Auto/Opening/Ending 請求
//
// 返回：
//   - any: 對應模式的響應對象
//   - error: 錯誤信息
func (a *NarrationAgent) Invoke(ctx context.Context, request any) (any, error) {
	switch req := request.(type) {
	case *SkeletonRequest:
		return a.InvokeSkeleton(ctx, req)
	case *ContentRequest:
		return a.InvokeContent(ctx, req)
	case *AutoContentRequest:
		return a.InvokeAutoContent(ctx, req)
	default:
		return nil, &agents.AgentError{
			AgentName: a.config.Name,
			Operation: "Invoke",
			Cause:     ErrUnsupportedRequestType,
			Retryable: false,
		}
	}
}

// GetName 返回 Agent 名稱
func (a *NarrationAgent) GetName() string {
	return a.config.Name
}

// GetTimeout 返回 Agent 超時時間
func (a *NarrationAgent) GetTimeout() time.Duration {
	return a.config.Timeout
}

// BuildPrompt 構建 Prompt（由具體模式實現）
//
// 此方法會根據 request 的類型分發到對應的 Prompt 構建邏輯。
// 目前僅實現 Skeleton 模式。
func (a *NarrationAgent) BuildPrompt(request any) (string, error) {
	switch req := request.(type) {
	case *SkeletonRequest:
		return a.buildSkeletonPrompt(req)
	default:
		return "", ErrUnsupportedRequestType
	}
}

// ParseResponse 解析 LLM 響應（由具體模式實現）
//
// 此方法會根據當前模式解析 LLM 的原始回應為結構化數據。
// 目前僅實現 Skeleton 模式的解析。
func (a *NarrationAgent) ParseResponse(raw string) (any, error) {
	// 默認使用 Skeleton 解析（後續可根據上下文判斷模式）
	return parseSkeletonResponse(raw)
}
