package narration

import "fmt"

// BuildFallbackSkeleton 構建降級故事骨架
//
// 當 LLM 調用失敗時，使用此方法生成一個通用的故事骨架。
// 降級骨架基於請求參數（主題、難度、長度）生成基本結構。
//
// 參數：
//   - request: Skeleton 請求
//
// 返回：
//   - *SkeletonResponse: 降級骨架
func BuildFallbackSkeleton(request *SkeletonRequest) *SkeletonResponse {
	// 規劃三幕結構
	threeAct := PlanThreeActStructure(request.StoryLength)

	// 規劃結局
	endings := PlanEndings(5) // 假設 5 個 Seeds

	// 規劃 Global Seeds
	// 根據難度決定 Seeds 數量，而不是長度
	numSeeds := getSeedsCountByDifficulty(request.Difficulty)

	// 生成足夠數量的 Seeds
	seeds := make([]GlobalSeedBlueprint, numSeeds)
	for i := 0; i < numSeeds; i++ {
		seedID := fmt.Sprintf("gs-%d", i+1)
		linkedEnding := ""
		if len(endings) > 0 {
			linkedEnding = endings[i%len(endings)].ID
		}

		clueChain := []ClueBlueprint{
			{
				Tier:        1,
				BeatRange:   threeAct.Act1.BeatRange,
				ClueContent: fmt.Sprintf("表面線索 #%d：暗示異常", i+1),
			},
			{
				Tier:        2,
				BeatRange:   threeAct.Act2.BeatRange,
				ClueContent: fmt.Sprintf("深層線索 #%d：揭示部分真相", i+1),
			},
			{
				Tier:        3,
				BeatRange:   threeAct.Act3.BeatRange,
				ClueContent: fmt.Sprintf("真相線索 #%d：完整揭露", i+1),
			},
		}

		plantBeatRange := [2]int{
			threeAct.Act1.BeatRange[0],
			threeAct.Act2.BeatRange[0] + 1,
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

	// 規劃關鍵劇情點
	plotPoints := PlanKeyPlotPoints(threeAct)

	// 構建世界觀
	worldView := WorldView{
		Setting:    fmt.Sprintf("神秘的%s", request.Theme),
		Atmosphere: "詭異、壓抑、未知的恐怖",
		TimeFrame:  "現代",
		Background: fmt.Sprintf(`這是一個關於%s的恐怖故事。

在這個看似平常的地方，隱藏著不為人知的秘密。隨著故事的推進，玩家將逐步發現表象之下的真相。

這個世界遵循著某些潛規則，違反這些規則將帶來可怕的後果。只有細心觀察、謹慎行動，才能在這個危險的環境中生存下來。

恐怖不是來自直接的展示，而是來自未知、暗示和逐步揭開的真相。`, request.Theme),
	}

	// 構建核心真相
	coreTruth := CoreTruth{
		Truth:      fmt.Sprintf("%s背後隱藏著被掩蓋的真相", request.Theme),
		HiddenFrom: "表象之下的規則與秘密",
		Revelation: "第三幕 Climax 時完整揭露",
	}

	// 構建劇情結構
	plotStructure := PlotStructure{
		ThreeAct:       threeAct,
		KeyPlotPoints:  plotPoints,
		GlobalSeeds:    seeds,
		EstimatedBeats: threeAct.Act3.BeatRange[1],
	}

	return &SkeletonResponse{
		WorldView:       worldView,
		CoreTruth:       coreTruth,
		PlotStructure:   plotStructure,
		PossibleEndings: endings,
		SelectedRules:   nil,
		SelectedScene:   nil,
		SelectedNPCs:    nil,
	}
}

// getSeedsCountByDifficulty 根據難度決定 Seeds 數量
func getSeedsCountByDifficulty(difficulty string) int {
	switch difficulty {
	case "easy":
		return 3
	case "normal":
		return 4
	case "hard", "hell":
		return 5
	default:
		return 4
	}
}
