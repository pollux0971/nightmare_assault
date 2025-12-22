// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.1 - MomentumController 基礎結構
package momentum

// FrequencyLevel 定義玩家互動頻率等級
type FrequencyLevel int

const (
	// FrequencyHigh 高頻率 - 每回合都暫停詢問玩家
	FrequencyHigh FrequencyLevel = iota
	// FrequencyMedium 中等頻率 - 在中風險以上暫停
	FrequencyMedium
	// FrequencyLow 低頻率 - 只在高風險或關鍵點暫停
	FrequencyLow
)

// String 返回頻率等級的字符串表示
func (f FrequencyLevel) String() string {
	switch f {
	case FrequencyHigh:
		return "High"
	case FrequencyMedium:
		return "Medium"
	case FrequencyLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// RiskLevel 定義當前情境的風險等級
type RiskLevel int

const (
	// RiskNone 無風險 - 安全的日常互動
	RiskNone RiskLevel = iota
	// RiskLow 低風險 - 輕微的挑戰或探索
	RiskLow
	// RiskMedium 中風險 - 可能造成傷害或損失
	RiskMedium
	// RiskHigh 高風險 - 嚴重的危險或重要決策
	RiskHigh
	// RiskLethal 致命風險 - 可能導致角色死亡
	RiskLethal
)

// String 返回風險等級的字符串表示
func (r RiskLevel) String() string {
	switch r {
	case RiskNone:
		return "None"
	case RiskLow:
		return "Low"
	case RiskMedium:
		return "Medium"
	case RiskHigh:
		return "High"
	case RiskLethal:
		return "Lethal"
	default:
		return "Unknown"
	}
}

// MomentumConfig 定義動量系統的配置
type MomentumConfig struct {
	// 頻率等級
	Frequency FrequencyLevel `json:"frequency"`

	// 自動演繹設定
	AutoResolve  bool `json:"auto_resolve"`   // 啟用自動演繹
	MaxAutoBeats int  `json:"max_auto_beats"` // 最大連續自動演繹回合數

	// 暫停條件
	PauseOnRisk  RiskLevel `json:"pause_on_risk"`  // 達到此風險等級時暫停
	PauseOnPlot  bool      `json:"pause_on_plot"`  // 劇情點時暫停
	PauseOnNPC   bool      `json:"pause_on_npc"`   // NPC 主動對話時暫停
	PauseOnEvent bool      `json:"pause_on_event"` // 重大事件時暫停

	// 玩家覆寫
	PlayerOverride bool `json:"player_override"` // 玩家可以覆寫自動演繹設定
}

// GameEvent 表示遊戲中發生的事件
type GameEvent struct {
	ID          string `json:"id"`           // 事件唯一標識
	Type        string `json:"type"`         // 事件類型 (npc_death, discovery, etc.)
	Description string `json:"description"`  // 事件描述
	IsMajor     bool   `json:"is_major"`     // 是否為重大事件
	Beat        int    `json:"beat"`         // 事件發生的回合數
	Location    string `json:"location"`     // 事件發生地點
	Initiator   string `json:"initiator"`    // 事件發起者
}

// NarrativeContext 表示當前敘事的上下文資訊
type NarrativeContext struct {
	// 當前狀態
	CurrentBeat  int    `json:"current_beat"`  // 當前回合數
	CurrentScene string `json:"current_scene"` // 當前場景

	// 風險評估
	RiskLevel   RiskLevel `json:"risk_level"`   // 當前風險等級
	RiskFactors []string  `json:"risk_factors"` // 風險因素列表

	// 劇情標記
	IsPlotPoint   bool   `json:"is_plot_point"`   // 是否為劇情點
	PlotPointType string `json:"plot_point_type"` // 劇情點類型

	// NPC 互動
	NPCInitiatesConversation bool   `json:"npc_initiates_conversation"` // NPC 是否主動發起對話
	InitiatingNPC            string `json:"initiating_npc"`              // 發起對話的 NPC ID

	// 事件觸發
	PendingEvents []*GameEvent `json:"pending_events"` // 待處理事件列表

	// 歷史
	RecentChoices     []string `json:"recent_choices"`      // 最近的選擇列表
	AutoResolvedBeats int      `json:"auto_resolved_beats"` // 已自動演繹的回合數
}

// StopReason 定義暫停原因
type StopReason int

const (
	// StopReasonNone 不暫停
	StopReasonNone StopReason = iota
	// StopReasonNPCConversation NPC 主動對話
	StopReasonNPCConversation
	// StopReasonMajorEvent 重大事件
	StopReasonMajorEvent
	// StopReasonPlotPoint 劇情點
	StopReasonPlotPoint
	// StopReasonRiskLevel 風險等級過高
	StopReasonRiskLevel
	// StopReasonMaxAutoBeats 達到最大自動回合數
	StopReasonMaxAutoBeats
	// StopReasonFrequency 頻率設定要求暫停
	StopReasonFrequency
)

// String 返回暫停原因的字符串表示
func (s StopReason) String() string {
	switch s {
	case StopReasonNone:
		return "None"
	case StopReasonNPCConversation:
		return "NPCConversation"
	case StopReasonMajorEvent:
		return "MajorEvent"
	case StopReasonPlotPoint:
		return "PlotPoint"
	case StopReasonRiskLevel:
		return "RiskLevel"
	case StopReasonMaxAutoBeats:
		return "MaxAutoBeats"
	case StopReasonFrequency:
		return "Frequency"
	default:
		return "Unknown"
	}
}
