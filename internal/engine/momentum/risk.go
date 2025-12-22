// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.3 - RiskEvaluator 風險評估器
package momentum

import (
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// RiskAssessment 表示風險評估結果
// AC2: 包含整體風險等級與貢獻因素列表
type RiskAssessment struct {
	Level   RiskLevel    // 整體風險等級 (所有因素的最高等級)
	Factors []RiskFactor // 所有風險因素列表
}

// RiskFactor 表示單一風險因素
// AC2: 包含因素名稱與風險等級
type RiskFactor struct {
	Name  string    // 因素名稱 (如 "low_hp", "high_tension")
	Level RiskLevel // 該因素的風險等級
}

// addFactor 添加風險因素並更新整體風險等級
// 採用「最高風險優先」策略
func (a *RiskAssessment) addFactor(name string, level RiskLevel) {
	a.Factors = append(a.Factors, RiskFactor{
		Name:  name,
		Level: level,
	})
	// 更新整體風險為最高等級
	if level > a.Level {
		a.Level = level
	}
}

// RiskEvaluator 評估當前情境的風險等級
// Story 6.3: 完整實作風險評估邏輯
type RiskEvaluator struct {
	// 未來可注入 rulesEngine, sceneConfig 等依賴
}

// NewRiskEvaluator 創建新的風險評估器
func NewRiskEvaluator() *RiskEvaluator {
	return &RiskEvaluator{}
}

// EvaluateContext 評估當前場景的風險等級
// AC1: 綜合考慮 HP/SAN, 活躍規則, 張力, 場景危險度
// AC2: 返回包含風險等級與貢獻因素的 RiskAssessment
func (e *RiskEvaluator) EvaluateContext(state *engine.GameStateV2, scene string) (*RiskAssessment, error) {
	assessment := &RiskAssessment{
		Level:   RiskNone,
		Factors: make([]RiskFactor, 0),
	}

	// 如果 state 為 nil，返回空評估
	if state == nil {
		return assessment, nil
	}

	// AC3: 評估 HP 狀態
	e.evaluateHP(state, assessment)

	// AC3: 評估 SAN 狀態
	e.evaluateSAN(state, assessment)

	// AC4: 評估活躍規則數量
	e.evaluateActiveRules(state, assessment)

	// AC5: 評估張力等級
	e.evaluateTension(state, assessment)

	// AC6: 評估場景危險度
	e.evaluateScene(state, scene, assessment)

	return assessment, nil
}

// evaluateHP 評估 HP 狀態的風險
// AC3: HP ≤ 20 → High, HP ≤ 40 → Medium
func (e *RiskEvaluator) evaluateHP(state *engine.GameStateV2, assessment *RiskAssessment) {
	hp := state.GetHP()

	if hp <= 20 {
		assessment.addFactor("low_hp", RiskHigh)
	} else if hp <= 40 {
		assessment.addFactor("medium_hp", RiskMedium)
	}
	// HP > 40: 不添加因素
}

// evaluateSAN 評估 SAN 狀態的風險
// AC3: SAN ≤ 20 → High, SAN ≤ 40 → Medium
func (e *RiskEvaluator) evaluateSAN(state *engine.GameStateV2, assessment *RiskAssessment) {
	san := state.GetSAN()

	if san <= 20 {
		assessment.addFactor("low_san", RiskHigh)
	} else if san <= 40 {
		assessment.addFactor("medium_san", RiskMedium)
	}
	// SAN > 40: 不添加因素
}

// evaluateActiveRules 評估活躍規則數量的風險
// AC4: 規則數 ≥ 5 → High, 規則數 ≥ 3 → Medium
func (e *RiskEvaluator) evaluateActiveRules(state *engine.GameStateV2, assessment *RiskAssessment) {
	ruleCount := len(state.ActiveRules)

	if ruleCount >= 5 {
		assessment.addFactor("many_active_rules", RiskHigh)
	} else if ruleCount >= 3 {
		assessment.addFactor("some_active_rules", RiskMedium)
	}
	// 規則數 < 3: 不添加因素
}

// evaluateTension 評估張力等級的風險
// AC5: 張力 ≥ 80 → High, 張力 ≥ 50 → Medium
func (e *RiskEvaluator) evaluateTension(state *engine.GameStateV2, assessment *RiskAssessment) {
	// 安全檢查：Tension 可能為 nil
	if state.Tension == nil {
		return
	}

	tension := state.Tension.GetValue()

	if tension >= 80 {
		assessment.addFactor("high_tension", RiskHigh)
	} else if tension >= 50 {
		assessment.addFactor("medium_tension", RiskMedium)
	}
	// 張力 < 50: 不添加因素
}

// evaluateScene 評估場景危險度
// AC6: 檢查場景名稱關鍵字，返回場景風險等級
func (e *RiskEvaluator) evaluateScene(state *engine.GameStateV2, scene string, assessment *RiskAssessment) {
	sceneRisk := e.evaluateSceneDanger(scene, state)

	// 如果場景危險度 ≥ RiskMedium, 添加因素
	if sceneRisk >= RiskMedium {
		assessment.addFactor("dangerous_scene", sceneRisk)
	}
}

// evaluateSceneDanger 評估場景本身的危險度
// AC6: 基於場景名稱關鍵字判斷危險度
// 此為簡單實作，未來可擴展為 LLM 判斷或配置檔案
func (e *RiskEvaluator) evaluateSceneDanger(scene string, state *engine.GameStateV2) RiskLevel {
	if scene == "" {
		return RiskNone
	}

	// 轉為小寫方便比對
	sceneLower := strings.ToLower(scene)

	// 高風險場景關鍵字
	highRiskKeywords := []string{
		"地下室", "basement", "cellar",
		"墓地", "cemetery", "graveyard",
		"太平間", "morgue",
		"手術室", "operating room", "surgery",
		"密室", "secret room", "hidden room",
		"祭壇", "altar", "ritual",
		"深淵", "abyss", "chasm",
	}

	for _, keyword := range highRiskKeywords {
		if strings.Contains(sceneLower, keyword) {
			return RiskHigh
		}
	}

	// 中風險場景關鍵字
	mediumRiskKeywords := []string{
		"走廊", "corridor", "hallway",
		"樓梯", "stairs", "stairway",
		"陰暗", "dark", "shadowy",
		"廢棄", "abandoned", "derelict",
		"病房", "ward", "hospital room",
		"實驗室", "laboratory", "lab",
		"閣樓", "attic",
	}

	for _, keyword := range mediumRiskKeywords {
		if strings.Contains(sceneLower, keyword) {
			return RiskMedium
		}
	}

	// 預設為無風險 (如: "安全房", "休息室", "大廳" 等)
	return RiskNone
}

// Evaluate 評估風險等級 (向後兼容舊介面)
// 此方法保留以維持與 controller.go 的兼容性
// 建議未來遷移到 EvaluateContext
func (e *RiskEvaluator) Evaluate(ctx *NarrativeContext) RiskLevel {
	// 如果 context 為 nil，返回無風險
	if ctx == nil {
		return RiskNone
	}

	// 返回 context 中已評估的風險等級
	return ctx.RiskLevel
}
