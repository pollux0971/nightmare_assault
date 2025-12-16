package narration

import (
	"fmt"
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/templates"
)

// SelectedTemplates 存儲選中的模板
type SelectedTemplates struct {
	Rules []*templates.RuleTemplate
	Scene *templates.SceneTemplate
	NPCs  []*templates.NPCTemplate
}

// selectTemplates 從 Template Library 選擇規則、場景和 NPC
//
// AC #2 要求：
//  - 選擇 3-5 條規則（簡單難度 ≤6 條）
//  - 選擇 1 個場景模板
//  - 選擇 2-3 個 NPC 原型（必含 N-01 犧牲者）
//
// 參數：
//   - request: Skeleton 請求
//
// 返回：
//   - *SelectedTemplates: 選中的模板
//   - error: 錯誤信息
func (a *NarrationAgent) selectTemplates(request *SkeletonRequest) (*SelectedTemplates, error) {
	lib := request.TemplateLibrary
	if lib == nil {
		return nil, fmt.Errorf("TemplateLibrary is nil")
	}

	selected := &SelectedTemplates{}

	// 1. 選擇規則 (3-5 條，簡單難度 ≤6 條)
	ruleCount := 4 // 默認 4 條
	if request.Difficulty == "easy" {
		ruleCount = 3 // 簡單難度少一點
	} else if request.Difficulty == "hard" || request.Difficulty == "hell" {
		ruleCount = 5 // 困難難度多一點
	}

	// 根據難度限制最大規則數
	maxRules := 10
	if request.Difficulty == "easy" {
		maxRules = 6
	}
	if ruleCount > maxRules {
		ruleCount = maxRules
	}

	selectedRules := make([]*templates.RuleTemplate, 0, ruleCount)
	for i := 0; i < ruleCount; i++ {
		// 隨機選擇規則，優先選擇符合難度的
		var difficulty *templates.RuleDifficulty
		if request.Difficulty != "" {
			diff := mapDifficulty(request.Difficulty)
			difficulty = &diff
		}

		rule := lib.SelectRandomRule(nil, difficulty, nil)
		if rule != nil {
			selectedRules = append(selectedRules, rule)
			log.Printf("[%s] Selected rule: %s (%s)", a.config.Name, rule.Name, rule.ID)
		}
	}
	selected.Rules = selectedRules

	// 2. 選擇場景模板 (1 個)
	scene := lib.SelectRandomScene(nil, nil)
	if scene != nil {
		selected.Scene = scene
		log.Printf("[%s] Selected scene: %s (%s)", a.config.Name, scene.Name, scene.ID)
	}

	// 3. 選擇 NPC 原型 (2-3 個，必含犧牲者)
	npcCount := 3
	selectedNPCs := make([]*templates.NPCTemplate, 0, npcCount)

	// 首先選擇犧牲者 (N-01)
	victimArchetype := templates.NPCArchetypeVictim
	victim := lib.SelectRandomNPC(&victimArchetype, nil, nil)
	if victim != nil {
		selectedNPCs = append(selectedNPCs, victim)
		log.Printf("[%s] Selected NPC (victim): %s (%s)", a.config.Name, victim.Name, victim.ID)
	}

	// 再選擇其他 NPC
	for i := 1; i < npcCount; i++ {
		npc := lib.SelectRandomNPC(nil, nil, nil)
		if npc != nil {
			selectedNPCs = append(selectedNPCs, npc)
			log.Printf("[%s] Selected NPC: %s (%s)", a.config.Name, npc.Name, npc.ID)
		}
	}
	selected.NPCs = selectedNPCs

	log.Printf("[%s] Template selection completed: %d rules, %d NPCs",
		a.config.Name, len(selectedRules), len(selectedNPCs))

	return selected, nil
}

// mapDifficulty 將字符串難度映射到 RuleDifficulty
func mapDifficulty(diff string) templates.RuleDifficulty {
	switch diff {
	case "easy":
		return templates.RuleDifficultyEasy
	case "normal":
		return templates.RuleDifficultyMedium
	case "hard":
		return templates.RuleDifficultyHard
	case "hell":
		return templates.RuleDifficultyHell
	default:
		return templates.RuleDifficultyMedium
	}
}
