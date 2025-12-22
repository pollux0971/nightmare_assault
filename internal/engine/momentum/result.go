// Package momentum 提供敘事動量控制系統
// Story 7-1: AutoResolveResult 資料結構
package momentum

// AutoResolveResult 表示自動演繹的結果
// Story 7-1 AC2: 返回 AutoResolveResult (Narratives/BeatsResolved/HPDelta/SANDelta)
type AutoResolveResult struct {
	// Narratives 是自動生成的敘事段落列表
	Narratives []string `json:"narratives"`

	// BeatsResolved 是已演繹的回合數
	BeatsResolved int `json:"beats_resolved"`

	// HPDelta 是累計的 HP 變化（可能為負）
	HPDelta int `json:"hp_delta"`

	// SANDelta 是累計的 SAN 變化（可能為負）
	SANDelta int `json:"san_delta"`

	// StopReason 是停止自動演繹的原因
	StopReason StopReason `json:"stop_reason"`

	// StopContext 包含停止時的上下文信息
	StopContext *NarrativeContext `json:"stop_context,omitempty"`

	// PlantedSeeds 是新種植的 Seeds
	PlantedSeeds []string `json:"planted_seeds,omitempty"`

	// RevealedClues 是揭露的線索
	RevealedClues []string `json:"revealed_clues,omitempty"`
}

// AutoContentRequest 表示自動內容生成請求
// 這是傳給 NarrationAgent.InvokeAutoContent() 的請求結構
type AutoContentRequest struct {
	// CurrentBeat 是當前回合數
	CurrentBeat int

	// AutoAction 是自動行動描述（可選）
	AutoAction string

	// Context 是當前敘事上下文
	Context *NarrativeContext

	// 其他必要字段可從 GameState 獲取
	// 這裡只定義 momentum 包需要的最小接口
}

// AutoContentResponse 表示自動內容生成響應
// 這是 NarrationAgent.InvokeAutoContent() 返回的結構
type AutoContentResponse struct {
	// Narrative 是生成的敘事文本
	Narrative string

	// HPDelta 是 HP 變化
	HPDelta int

	// SANDelta 是 SAN 變化
	SANDelta int

	// PlantedSeeds 是新種植的 Seeds
	PlantedSeeds []string

	// RevealedClues 是揭露的線索
	RevealedClues []string
}
