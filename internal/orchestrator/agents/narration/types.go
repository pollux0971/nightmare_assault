package narration

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/templates"
)

// SkeletonRequest 是骨架規劃請求
//
// 包含生成故事骨架所需的所有信息：
//   - Theme: 主題（如「廢棄醫院」「詭異學校」）
//   - Difficulty: 難度等級（easy/normal/hard/hell）
//   - StoryLength: 故事長度（short/medium/long）
//   - Adult18Plus: 是否啟用成人內容
//   - TemplateLibrary: 模板庫（規則/場景/NPC）
type SkeletonRequest struct {
	Theme       string
	Difficulty  string // easy/normal/hard/hell
	StoryLength string // short/medium/long
	Adult18Plus bool

	// Epic 4 整合：Template Library
	TemplateLibrary *templates.TemplateLibrary
}

// SkeletonResponse 是骨架規劃響應
//
// 包含完整的故事骨架：
//   - WorldView: 世界觀設定
//   - CoreTruth: 核心真相
//   - PlotStructure: 劇情結構（三幕結構、Global Seeds、關鍵劇情點）
//   - PossibleEndings: 多重結局
//   - SelectedRules/Scene/NPCs: 選中的模板（Epic 4 整合）
type SkeletonResponse struct {
	WorldView       WorldView       `json:"world_view"`
	CoreTruth       CoreTruth       `json:"core_truth"`
	PlotStructure   PlotStructure   `json:"plot_structure"`
	PossibleEndings []Ending        `json:"possible_endings"`

	// 選中的模板（Epic 4 整合）
	SelectedRules []*templates.RuleTemplate  `json:"selected_rules,omitempty"`
	SelectedScene *templates.SceneTemplate   `json:"selected_scene,omitempty"`
	SelectedNPCs  []*templates.NPCTemplate   `json:"selected_npcs,omitempty"`
}

// WorldView 是世界觀設定
//
// 定義遊戲世界的基本設定：
//   - Setting: 場景設定（如「廢棄醫院」）
//   - Atmosphere: 氛圍描述（如「壓抑、詭異、血腥」）
//   - TimeFrame: 時間範圍（如「1990年代」）
//   - Background: 背景故事（500-800 字）
type WorldView struct {
	Setting    string `json:"setting"`
	Atmosphere string `json:"atmosphere"`
	TimeFrame  string `json:"time_frame"`
	Background string `json:"background"`
}

// CoreTruth 是核心真相
//
// 隱藏在故事表象之下的核心真相：
//   - Truth: 真相描述（如「醫院進行非法人體實驗」）
//   - HiddenFrom: 隱藏方式（如「偽裝成精神病院」）
//   - Revelation: 揭露時機（如「第三幕 Climax」）
type CoreTruth struct {
	Truth      string `json:"truth"`
	HiddenFrom string `json:"hidden_from"`
	Revelation string `json:"revelation"`
}

// PlotStructure 是劇情結構
//
// 包含故事的完整結構：
//   - ThreeAct: 三幕結構（Setup/Confrontation/Resolution）
//   - KeyPlotPoints: 關鍵劇情點（Inciting Incident, Midpoint, Climax）
//   - GlobalSeeds: Global Seeds 藍圖（3-5 個，每個含 3 層線索鏈）
//   - EstimatedBeats: 預估總章節數
type PlotStructure struct {
	ThreeAct      ThreeActStructure   `json:"three_act"`
	KeyPlotPoints []PlotPoint         `json:"key_plot_points"`
	GlobalSeeds   []GlobalSeedBlueprint `json:"global_seeds"`
	EstimatedBeats int                 `json:"estimated_beats"`
}

// ThreeActStructure 是三幕結構
//
// 經典的三幕劇結構：
//   - Act1: Setup（20-30% beats）：介紹世界觀、規則暗示
//   - Act2: Confrontation（40-60% beats）：張力累積、伏筆回收
//   - Act3: Resolution（10-20% beats）：真相揭露、結局收束
type ThreeActStructure struct {
	Act1 Act `json:"act1"`
	Act2 Act `json:"act2"`
	Act3 Act `json:"act3"`
}

// Act 是單幕結構
//
// 定義單幕的基本信息：
//   - Name: 幕名稱（如「Setup」）
//   - BeatRange: Beat 範圍（[開始 Beat, 結束 Beat]）
//   - Goals: 本幕目標
//   - KeyEvents: 關鍵事件
type Act struct {
	Name      string   `json:"name"`
	BeatRange [2]int   `json:"beat_range"`
	Goals     []string `json:"goals,omitempty"`
	KeyEvents []string `json:"key_events,omitempty"`
}

// PlotPoint 是關鍵劇情點
//
// 故事中的重要節點：
//   - Name: 劇情點名稱（如「Inciting Incident」）
//   - Beat: 發生在第幾章
//   - Description: 描述
type PlotPoint struct {
	Name        string `json:"name"`
	Beat        int    `json:"beat"`
	Description string `json:"description"`
}

// GlobalSeedBlueprint 是 Global Seed 藍圖
//
// 定義一個 Global Seed 的完整信息：
//   - ID: Seed 唯一標識
//   - Content: 伏筆內容（如「醫院地下室傳來奇怪聲音」）
//   - LinkedTruth: 關聯的真相
//   - LinkedEnding: 關聯的結局
//   - ClueChain: 3 層線索鏈（Tier 1-3）
//   - PlantBeatRange: 埋設時機範圍
type GlobalSeedBlueprint struct {
	ID             string           `json:"id"`
	Content        string           `json:"content"`
	LinkedTruth    string           `json:"linked_truth"`
	LinkedEnding   string           `json:"linked_ending"`
	ClueChain      []ClueBlueprint  `json:"clue_chain"`
	PlantBeatRange [2]int           `json:"plant_beat_range"`
}

// ClueBlueprint 是線索藍圖
//
// 定義 Global Seed 的單層線索：
//   - Tier: 層級（1: 表面 / 2: 深層 / 3: 真相）
//   - BeatRange: 出現時機範圍
//   - ClueContent: 線索內容
type ClueBlueprint struct {
	Tier        int    `json:"tier"`
	BeatRange   [2]int `json:"beat_range"`
	ClueContent string `json:"clue_content"`
}

// Ending 是結局
//
// 定義一個可能的結局：
//   - ID: 結局唯一標識
//   - Name: 結局名稱（如「True Ending」）
//   - Condition: 觸發條件
//   - Description: 結局描述（1000-1500 字）
//   - RequiredSeedPercentage: 需要的 Seed 揭示度（如 0.8 表示 80%）
type Ending struct {
	ID                     string          `json:"id"`
	Name                   string          `json:"name"`
	Condition              EndingCondition `json:"condition"`
	Description            string          `json:"description"`
	RequiredSeedPercentage float64         `json:"required_seed_percentage"`
}

// EndingCondition 是結局觸發條件
//
// 定義觸發結局所需的條件：
//   - MinSeedPercentage: 最低 Seed 揭示度（0.0-1.0）
//   - MaxRuleViolations: 最多規則違反次數
//   - MinHP: 最低 HP
//   - MinSAN: 最低 SAN
type EndingCondition struct {
	MinSeedPercentage float64 `json:"min_seed_percentage"`
	MaxRuleViolations int     `json:"max_rule_violations,omitempty"`
	MinHP             int     `json:"min_hp,omitempty"`
	MinSAN            int     `json:"min_san,omitempty"`
}

// ==========================================================================
// Opening & Ending Mode Types (Story 6.4)
// ==========================================================================

// StoryBible 是故事骨架的簡化視圖
//
// 用於 Opening & Ending 模式的輸入數據，避免直接依賴 orchestrator.StoryBible
type StoryBible struct {
	WorldView       WorldView
	CoreTruth       CoreTruth
	PlotStructure   PlotStructure
	PossibleEndings []Ending
}

// OpeningRequest 是序章生成請求
//
// 用於 Genesis Phase 生成引人入勝的序章（800-1200 字）：
//   - StoryBible: 故事骨架（來自 Skeleton Mode）
//   - Difficulty: 難度等級（影響規則暗示的隱晦程度）
//   - NPCs: NPC instances with introductions (Story 7.6)
type OpeningRequest struct {
	StoryBible *StoryBible
	Difficulty string // easy/normal/hard/hell
	NPCs       []NPCInfo // Story 7.6: NPC instances with introductions
}

// NPCInfo contains NPC information for opening narrative (Story 7.6)
type NPCInfo struct {
	Name         string
	Introduction string // Show-Don't-Tell introduction
}

// OpeningResponse 是序章生成響應
//
// 包含序章敘事與初始遊戲狀態：
//   - OpeningNarrative: 序章文本（800-1200 字）
//   - InitialTension: 初始張力值（10-20）
//   - FirstChoice: 引導第一個選擇的文本（50 字內）
type OpeningResponse struct {
	OpeningNarrative string `json:"opening_narrative"`
	InitialTension   int    `json:"initial_tension"`   // 10-20
	FirstChoice      string `json:"first_choice_prompt"` // 引導第一個選擇
}

// EndingRequest 是結局生成請求
//
// 用於 Convergence Phase 生成多結局敘事（1000-1500 字）：
//   - GameState: 當前遊戲狀態（含 Global Seeds 揭露狀態）
//   - EndingType: 結局類型（True/Good/Bad，可選，未指定則自動判定）
type EndingRequest struct {
	GameState  *engine.GameStateV2
	EndingType string // "true"/"good"/"bad" (optional, will auto-determine if empty)
}

// EndingResponse 是結局生成響應
//
// 包含結局敘事與情感解析：
//   - EndingNarrative: 結局文本（1000-1500 字，根據類型）
//   - FinalEmotion: 最終情感（shock/relief/despair）
//   - ClosingLine: 結局金句（最後一句，留下深刻印象）
type EndingResponse struct {
	EndingNarrative string `json:"ending_narrative"`
	FinalEmotion    string `json:"final_emotion"`    // shock/relief/despair
	ClosingLine     string `json:"closing_line"`     // 結局金句
}

// EndingType 枚舉定義
//
// 結局類型基於 Global Seeds 揭露程度：
//   - EndingTrue: ≥80% Global Seeds 完全揭露（Tier 1, 2, 3）
//   - EndingGood: 40-79% Global Seeds 完全揭露
//   - EndingBad: <40% Global Seeds 揭露
const (
	EndingTrue = "true"
	EndingGood = "good"
	EndingBad  = "bad"
)
