package narration

import (
	"fmt"
)

// PlanGlobalSeeds 規劃 Global Seeds 分佈
//
// 根據劇情結構和結局規劃 3-5 個 Global Seeds，每個 Seed 包含：
//   - 3 層線索鏈（Tier 1: 表面 / Tier 2: 深層 / Tier 3: 真相）
//   - 關聯到特定結局
//   - 埋設時機範圍
//
// 參數：
//   - plotStructure: 三幕結構
//   - endings: 可能的結局列表
//   - length: 故事長度（用於調整 Seeds 數量）
//
// 返回：
//   - []GlobalSeedBlueprint: Global Seeds 列表
func PlanGlobalSeeds(plotStructure ThreeActStructure, endings []Ending, length string) []GlobalSeedBlueprint {
	// 根據故事長度決定 Seeds 數量
	numSeeds := 3
	if length == "medium" {
		numSeeds = 4
	} else if length == "long" {
		numSeeds = 5
	}

	// 確保至少有 3 個 Seeds
	if numSeeds < 3 {
		numSeeds = 3
	}

	seeds := make([]GlobalSeedBlueprint, numSeeds)

	for i := 0; i < numSeeds; i++ {
		// 生成 Seed ID
		seedID := fmt.Sprintf("gs-%d", i+1)

		// 分配關聯的結局（循環分配）
		linkedEnding := ""
		if len(endings) > 0 {
			linkedEnding = endings[i%len(endings)].ID
		}

		// 生成 3 層線索鏈
		clueChain := []ClueBlueprint{
			{
				Tier:        1,
				BeatRange:   plotStructure.Act1.BeatRange,
				ClueContent: fmt.Sprintf("表面線索 #%d：暗示異常", i+1),
			},
			{
				Tier:        2,
				BeatRange:   plotStructure.Act2.BeatRange,
				ClueContent: fmt.Sprintf("深層線索 #%d：揭示部分真相", i+1),
			},
			{
				Tier:        3,
				BeatRange:   plotStructure.Act3.BeatRange,
				ClueContent: fmt.Sprintf("真相線索 #%d：完整揭露", i+1),
			},
		}

		// 埋設時機範圍（在 Act1-Act2 之間）
		plantBeatRange := [2]int{
			plotStructure.Act1.BeatRange[0],
			plotStructure.Act2.BeatRange[0] + 1,
		}

		seeds[i] = GlobalSeedBlueprint{
			ID:             seedID,
			Content:        fmt.Sprintf("Global Seed #%d：核心伏筆", i+1),
			LinkedTruth:    "核心真相的一部分",
			LinkedEnding:   linkedEnding,
			ClueChain:      clueChain,
			PlantBeatRange: plantBeatRange,
		}
	}

	return seeds
}

// PlanThreeActStructure 規劃三幕結構
//
// 根據故事長度規劃三幕結構：
//   - Act1: Setup（25% beats）
//   - Act2: Confrontation（50% beats）
//   - Act3: Resolution（25% beats）
//
// 參數：
//   - length: 故事長度（short/medium/long）
//
// 返回：
//   - ThreeActStructure: 三幕結構
func PlanThreeActStructure(length string) ThreeActStructure {
	// 根據長度決定總 Beats 數
	totalBeats := estimateBeats(length)

	// 計算三幕 Beat 範圍
	act1End := int(float64(totalBeats) * 0.25)
	if act1End < 1 {
		act1End = 1
	}

	act2End := int(float64(totalBeats) * 0.75)
	if act2End <= act1End {
		act2End = act1End + 1
	}

	return ThreeActStructure{
		Act1: Act{
			Name:      "Setup（設定）",
			BeatRange: [2]int{1, act1End},
			Goals: []string{
				"介紹世界觀與主角",
				"暗示潛規則的存在",
				"建立詭異氛圍",
			},
			KeyEvents: []string{
				"序章：進入恐怖場景",
				"Inciting Incident：第一次異常事件",
			},
		},
		Act2: Act{
			Name:      "Confrontation（對抗）",
			BeatRange: [2]int{act1End + 1, act2End},
			Goals: []string{
				"張力累積與高潮",
				"Global Seeds 線索逐步揭露",
				"規則觸發與學習",
			},
			KeyEvents: []string{
				"First Plot Point：衝突升級",
				"Midpoint：重大發現或轉折",
				"NPC 死亡（教學規則）",
			},
		},
		Act3: Act{
			Name:      "Resolution（解決）",
			BeatRange: [2]int{act2End + 1, totalBeats},
			Goals: []string{
				"真相揭露",
				"結局收束",
			},
			KeyEvents: []string{
				"Climax：最終對抗或逃脫",
				"結局：根據 Seeds 揭示度分支",
			},
		},
	}
}

// PlanKeyPlotPoints 規劃關鍵劇情點
//
// 根據三幕結構規劃關鍵劇情點：
//   - Inciting Incident: Act1 末尾
//   - First Plot Point: Act1-Act2 轉折
//   - Midpoint: Act2 中點
//   - Climax: Act3 開始
//
// 參數：
//   - structure: 三幕結構
//
// 返回：
//   - []PlotPoint: 關鍵劇情點列表
func PlanKeyPlotPoints(structure ThreeActStructure) []PlotPoint {
	plotPoints := []PlotPoint{
		{
			Name:        "Inciting Incident",
			Beat:        structure.Act1.BeatRange[1],
			Description: "第一次異常事件，打破日常",
		},
		{
			Name:        "First Plot Point",
			Beat:        structure.Act2.BeatRange[0],
			Description: "衝突升級，主角被迫應對",
		},
		{
			Name:        "Midpoint",
			Beat:        (structure.Act2.BeatRange[0] + structure.Act2.BeatRange[1]) / 2,
			Description: "重大發現或轉折，改變故事走向",
		},
		{
			Name:        "Climax",
			Beat:        structure.Act3.BeatRange[0],
			Description: "最終對抗或逃脫，真相揭露",
		},
	}

	return plotPoints
}

// PlanEndings 規劃多重結局
//
// 根據 Global Seeds 數量規劃 2-3 個結局：
//   - True Ending: ≥80% Seeds 揭示
//   - Good Ending: 40-79% Seeds 揭示
//   - Bad Ending: <40% Seeds 揭示 或 Death/Madness
//
// 參數：
//   - numSeeds: Global Seeds 數量
//
// 返回：
//   - []Ending: 結局列表
func PlanEndings(numSeeds int) []Ending {
	endings := []Ending{
		{
			ID:   "true-ending",
			Name: "True Ending（真實結局）",
			Condition: EndingCondition{
				MinSeedPercentage: 0.8,
				MaxRuleViolations: 2,
				MinHP:             30,
				MinSAN:            20,
			},
			Description:            "揭露完整真相，達成最佳結局。玩家成功發現並理解了所有核心伏筆，找到了事件背後的完整真相。",
			RequiredSeedPercentage: 0.8,
		},
		{
			ID:   "good-ending",
			Name: "Good Ending（良好結局）",
			Condition: EndingCondition{
				MinSeedPercentage: 0.4,
				MaxRuleViolations: 5,
				MinHP:             10,
				MinSAN:            10,
			},
			Description:            "部分揭露真相，成功逃脫。玩家理解了部分真相，雖然未能完全解開謎團，但成功存活並逃離。",
			RequiredSeedPercentage: 0.4,
		},
		{
			ID:   "bad-ending",
			Name: "Bad Ending（悲劇結局）",
			Condition: EndingCondition{
				MinSeedPercentage: 0.0,
				MaxRuleViolations: 999,
				MinHP:             0,
				MinSAN:            0,
			},
			Description:            "真相未明，陷入絕境。玩家未能理解足夠的真相，或是違反了過多規則，導致悲劇結局（死亡或瘋狂）。",
			RequiredSeedPercentage: 0.0,
		},
	}

	return endings
}

// estimateBeats 估算故事總 Beats 數
func estimateBeats(length string) int {
	switch length {
	case "short":
		return 4 // 3-5 章（取中間值）
	case "medium":
		return 7 // 5-8 章（取中間值）
	case "long":
		return 12 // 8-15 章（取中間值）
	default:
		return 7
	}
}
