package narration

import (
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
)

// ConvertToStoryBible 將 SkeletonResponse 轉換為 StoryBible
//
// Epic 2 整合：將 Skeleton 骨架轉換為可供後續 Agents 使用的 StoryBible 結構
//
// 參數：
//   - response: Skeleton 響應
//   - theme: 遊戲主題
//
// 返回：
//   - *orchestrator.StoryBible: 故事聖經
func ConvertToStoryBible(response *SkeletonResponse, theme string) *orchestrator.StoryBible {
	if response == nil {
		return nil
	}

	// 轉換 Global Seeds - 使用 Epic 2 的 NewGlobalSeed 構造函數
	globalSeeds := make([]*seed.GlobalSeed, 0, len(response.PlotStructure.GlobalSeeds))
	for _, seedBlueprint := range response.PlotStructure.GlobalSeeds {
		// 轉換 ClueBlueprint 到 ClueTier
		clueChain := make([]seed.ClueTier, len(seedBlueprint.ClueChain))
		for i, clue := range seedBlueprint.ClueChain {
			clueChain[i] = seed.ClueTier{
				Tier:      clue.Tier,
				Content:   clue.ClueContent,
				Keywords:  []string{}, // TODO: 從 ClueContent 提取關鍵字
				BeatStart: clue.BeatRange[0],
				BeatEnd:   clue.BeatRange[1],
			}
		}

		// 使用 Epic 2 提供的構造函數
		globalSeed, err := seed.NewGlobalSeed(
			seedBlueprint.ID,
			seedBlueprint.Content,
			seedBlueprint.LinkedTruth,
			seedBlueprint.LinkedEnding,
			clueChain,
		)
		if err != nil {
			log.Printf("[StoryBible] Warning: Failed to create GlobalSeed %s: %v", seedBlueprint.ID, err)
			continue
		}
		globalSeeds = append(globalSeeds, globalSeed)
	}

	// 轉換 NPC Profiles
	npcProfiles := make([]*orchestrator.NPCProfile, 0, len(response.SelectedNPCs))
	for _, npcTemplate := range response.SelectedNPCs {
		if npcTemplate != nil {
			profile := &orchestrator.NPCProfile{
				ID:          npcTemplate.ID,
				Name:        npcTemplate.Name,
				Description: npcTemplate.FunctionalRole,
			}
			npcProfiles = append(npcProfiles, profile)
		}
	}

	bible := &orchestrator.StoryBible{
		WorldView:   response.WorldView.Background,
		MainTheme:   theme,
		Setting:     response.WorldView.Setting,
		GlobalSeeds: globalSeeds,
		NPCProfiles: npcProfiles,
		// UsedTemplates 將在後續填充
	}

	log.Printf("[StoryBible] Converted skeleton to StoryBible: %d global seeds, %d NPCs",
		len(globalSeeds), len(npcProfiles))

	return bible
}
