package commands

import (
	"fmt"
	"strconv"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
	"github.com/nightmare-assault/nightmare-assault/internal/game/savefile"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// StateRestorer provides methods to restore game state from save files.
type StateRestorer interface {
	RestoreGameState(*engine.GameStateV2, *orchestrator.StoryBible, *savefile.GameSettings) error
}

// LoadCommandV2 implements the /load command for v2.0 architecture.
// Story 8.2 AC: Load command integration
type LoadCommandV2 struct {
	stateRestorer StateRestorer
	saveDir       string
}

// NewLoadCommandV2 creates a new load command for v2.0.
func NewLoadCommandV2(stateRestorer StateRestorer) *LoadCommandV2 {
	return &LoadCommandV2{
		stateRestorer: stateRestorer,
		saveDir:       save.GetSaveDir(),
	}
}

// Name returns the command name.
func (c *LoadCommandV2) Name() string {
	return "load"
}

// Execute executes the load command.
// Story 8.2 AC: /load <slot> restores saved game state
func (c *LoadCommandV2) Execute(args []string) (string, error) {
	// Parse slot ID
	if len(args) == 0 {
		return c.showSlotList()
	}

	slotID, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("無效的槽位編號：%s", args[0])
	}

	// Validate slot ID
	if err := save.ValidateSlotID(slotID); err != nil {
		return "", err
	}

	// Load save file
	saveFile, err := savefile.LoadV2(c.saveDir, slotID)
	if err != nil {
		return "", fmt.Errorf("讀檔失敗：%w", err)
	}

	// Convert savefile.StoryBibleData back to orchestrator.StoryBible
	storyBible := convertToOrchestratorStoryBible(saveFile.StoryBible)

	// Restore game state
	if err := c.stateRestorer.RestoreGameState(saveFile.GameState, storyBible, &saveFile.Settings); err != nil {
		return "", fmt.Errorf("恢復遊戲狀態失敗：%w", err)
	}

	return fmt.Sprintf(`
記憶正在甦醒...

槽位 %d 的存檔已載入
遊戲時間：%d 秒
當前章節：第 %d 幕

歡迎回到噩夢...
`, slotID, saveFile.Meta.PlayTime, saveFile.GameState.GetCurrentBeat()), nil
}

// Help returns brief command description.
func (c *LoadCommandV2) Help() string {
	return "從存檔槽位載入遊戲進度"
}

// showSlotList shows available save slots.
func (c *LoadCommandV2) showSlotList() (string, error) {
	var result string
	result += "可用的存檔：\n\n"

	hasAnySave := false
	for slotID := save.MinSlotID; slotID <= save.MaxSlotID; slotID++ {
		slotInfo, err := savefile.GetSlotInfo(c.saveDir, slotID)
		if err != nil {
			continue
		}

		if !slotInfo.IsEmpty {
			hasAnySave = true
			result += fmt.Sprintf("  槽位 %d: 第 %d 幕 | %d 秒 | %s\n",
				slotID,
				slotInfo.Chapter,
				slotInfo.PlayTime,
				slotInfo.SavedAt.Format("2006-01-02 15:04"))
		}
	}

	if !hasAnySave {
		result += "  [沒有存檔]\n"
	}

	result += "\n使用 /load <槽位> 來載入存檔\n"
	result += "例如: /load 1\n"

	return result, nil
}

// convertToOrchestratorStoryBible converts savefile.StoryBibleData back to orchestrator.StoryBible
func convertToOrchestratorStoryBible(src *savefile.StoryBibleData) *orchestrator.StoryBible {
	if src == nil {
		return nil
	}

	dest := &orchestrator.StoryBible{
		GameID:     src.GameID,
		CreatedAt:  src.CreatedAt,
		Difficulty: src.Difficulty,
		TotalBeats: src.TotalBeats,
		WorldView:  src.WorldView,
		MainTheme:  src.MainTheme,
		Setting:    src.Setting,
	}

	// Convert WorldSetting
	if src.WorldSetting != nil {
		dest.WorldSetting = &orchestrator.WorldSetting{
			Location:      src.WorldSetting.Location,
			History:       src.WorldSetting.History,
			WeirdElements: src.WorldSetting.WeirdElements,
			Atmosphere:    src.WorldSetting.Atmosphere,
			TimeFrame:     src.WorldSetting.TimeFrame,
			Background:    src.WorldSetting.Background,
		}
	}

	// Convert CoreMystery
	if src.CoreMystery != nil {
		dest.CoreMystery = &orchestrator.CoreMystery{
			Question:   src.CoreMystery.Question,
			CoreTruth:  src.CoreMystery.CoreTruth,
			Revelation: src.CoreMystery.Revelation,
			HiddenFrom: src.CoreMystery.HiddenFrom,
		}
	}

	// Convert StoryArc
	if src.StoryArc != nil {
		dest.StoryArc = &orchestrator.StoryArc{
			Act1End:  src.StoryArc.Act1End,
			Midpoint: src.StoryArc.Midpoint,
			Act2End:  src.StoryArc.Act2End,
		}
		if src.StoryArc.TurningPoints != nil {
			dest.StoryArc.TurningPoints = make([]*orchestrator.TurningPoint, len(src.StoryArc.TurningPoints))
			for i, tp := range src.StoryArc.TurningPoints {
				dest.StoryArc.TurningPoints[i] = &orchestrator.TurningPoint{
					Name:        tp.Name,
					Beat:        tp.Beat,
					Description: tp.Description,
				}
			}
		}
	}

	// Convert HiddenRules
	if src.HiddenRules != nil {
		dest.HiddenRules = make([]*orchestrator.HiddenRule, len(src.HiddenRules))
		for i, hr := range src.HiddenRules {
			dest.HiddenRules[i] = &orchestrator.HiddenRule{
				ID:          hr.ID,
				Name:        hr.Name,
				Description: hr.Description,
				Hints:       hr.Hints,
				Penalty:     hr.Penalty,
			}
		}
	}

	// GlobalSeeds can be copied directly (same type)
	dest.GlobalSeeds = src.GlobalSeeds

	// Convert NPCProfiles
	if src.NPCProfiles != nil {
		dest.NPCProfiles = make([]*orchestrator.NPCProfile, len(src.NPCProfiles))
		for i, np := range src.NPCProfiles {
			dest.NPCProfiles[i] = &orchestrator.NPCProfile{
				ID:           np.ID,
				Name:         np.Name,
				Archetype:    agents.NPCArchetype(np.Archetype),
				Personality:  np.Personality,
				Appearance:   np.Appearance,
				Backstory:    np.Backstory,
				Skills:       np.Skills,
				Inventory:    np.Inventory,
				Secret:       np.Secret,
				Introduction: np.Introduction,
				LinkedSeeds:  np.LinkedSeeds,
				DeathTiming:  np.DeathTiming,
				Status:       agents.NPCStatus(np.Status),
				DeathBeat:    np.DeathBeat,
				DeathReason:  np.DeathReason,
				Description:  np.Description,
			}
		}
	}

	// Convert PossibleEndings
	if src.PossibleEndings != nil {
		dest.PossibleEndings = make([]*orchestrator.Ending, len(src.PossibleEndings))
		for i, e := range src.PossibleEndings {
			ending := &orchestrator.Ending{
				ID:                     e.ID,
				Name:                   e.Name,
				Type:                   e.Type,
				Description:            e.Description,
				RequiredSeedPercentage: e.RequiredSeedPercentage,
			}
			if e.Condition != nil {
				ending.Condition = &orchestrator.EndingCondition{
					MinSeedPercentage: e.Condition.MinSeedPercentage,
					MaxRuleViolations: e.Condition.MaxRuleViolations,
					MinHP:             e.Condition.MinHP,
					MinSAN:            e.Condition.MinSAN,
				}
			}
			dest.PossibleEndings[i] = ending
		}
	}

	// Convert UsedTemplates
	if src.UsedTemplates != nil {
		dest.UsedTemplates = &engine.UsedTemplates{
			Rules:  src.UsedTemplates.Rules,
			Scenes: src.UsedTemplates.Scenes,
		}
	}

	return dest
}
