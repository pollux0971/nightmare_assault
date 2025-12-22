package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
	"github.com/nightmare-assault/nightmare-assault/internal/game/savefile"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
)

// StateAccessor provides access to game state for save/load operations.
type StateAccessor interface {
	GetGameState() *engine.GameStateV2
	GetStoryBible() *orchestrator.StoryBible
	GetGameSettings() *savefile.GameSettings
	GetGameStartTime() time.Time
}

// SaveCommandV2 implements the /save command for v2.0 architecture.
// Story 8.1 AC #1: Save command integration
type SaveCommandV2 struct {
	stateAccessor StateAccessor
	saveDir       string
}

// NewSaveCommandV2 creates a new save command for v2.0.
func NewSaveCommandV2(stateAccessor StateAccessor) *SaveCommandV2 {
	return &SaveCommandV2{
		stateAccessor: stateAccessor,
		saveDir:       save.GetSaveDir(),
	}
}

// Name returns the command name.
func (c *SaveCommandV2) Name() string {
	return "save"
}

// Execute executes the save command.
// Story 8.1 AC #1: /save <slot> saves current game state
func (c *SaveCommandV2) Execute(args []string) (string, error) {
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

	// Get game state from accessor
	gameState := c.stateAccessor.GetGameState()
	if gameState == nil {
		return "", fmt.Errorf("無法獲取遊戲狀態")
	}

	// Get story bible from accessor
	storyBible := c.stateAccessor.GetStoryBible()
	if storyBible == nil {
		return "", fmt.Errorf("無法獲取故事聖經")
	}

	// Get game settings from accessor
	settings := c.stateAccessor.GetGameSettings()
	if settings == nil {
		return "", fmt.Errorf("無法獲取遊戲設定")
	}

	// Calculate playtime (High Priority Issue #6)
	gameStartTime := c.stateAccessor.GetGameStartTime()
	playtime := int(time.Since(gameStartTime).Seconds())

	// Convert orchestrator.StoryBible to savefile.StoryBibleData
	bibleData := convertStoryBible(storyBible)

	// Create save file
	saveFile := savefile.NewSaveFileV2(slotID, *settings, gameState, bibleData)
	saveFile.Meta.PlayTime = playtime

	// Save to disk
	if err := savefile.SaveV2(c.saveDir, slotID, saveFile); err != nil {
		return "", fmt.Errorf("存檔失敗：%w", err)
	}

	return fmt.Sprintf(`
記憶已固化...

槽位 %d 的存檔已保存
遊戲時間：%d 秒
當前章節：第 %d 幕

你可以隨時使用 /load %d 來恢復這段記憶。
`, slotID, playtime, gameState.GetCurrentBeat(), slotID), nil
}

// Help returns brief command description.
func (c *SaveCommandV2) Help() string {
	return "保存遊戲進度到存檔槽位"
}

// showSlotList shows available save slots.
func (c *SaveCommandV2) showSlotList() (string, error) {
	var result string
	result += "可用的存檔槽位：\n\n"

	for slotID := save.MinSlotID; slotID <= save.MaxSlotID; slotID++ {
		slotInfo, err := savefile.GetSlotInfo(c.saveDir, slotID)
		if err != nil {
			continue
		}

		if slotInfo.IsEmpty {
			result += fmt.Sprintf("  槽位 %d: [空]\n", slotID)
		} else {
			result += fmt.Sprintf("  槽位 %d: 第 %d 幕 | %d 秒 | %s\n",
				slotID,
				slotInfo.Chapter,
				slotInfo.PlayTime,
				slotInfo.SavedAt.Format("2006-01-02 15:04"))
		}
	}

	result += "\n使用 /save <槽位> 來保存遊戲進度\n"
	result += "例如: /save 1\n"

	return result, nil
}

// convertStoryBible converts orchestrator.StoryBible to savefile.StoryBibleData
func convertStoryBible(src *orchestrator.StoryBible) *savefile.StoryBibleData {
	if src == nil {
		return nil
	}

	dest := &savefile.StoryBibleData{
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
		dest.WorldSetting = &savefile.WorldSetting{
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
		dest.CoreMystery = &savefile.CoreMystery{
			Question:   src.CoreMystery.Question,
			CoreTruth:  src.CoreMystery.CoreTruth,
			Revelation: src.CoreMystery.Revelation,
			HiddenFrom: src.CoreMystery.HiddenFrom,
		}
	}

	// Convert StoryArc
	if src.StoryArc != nil {
		dest.StoryArc = &savefile.StoryArc{
			Act1End:  src.StoryArc.Act1End,
			Midpoint: src.StoryArc.Midpoint,
			Act2End:  src.StoryArc.Act2End,
		}
		if src.StoryArc.TurningPoints != nil {
			dest.StoryArc.TurningPoints = make([]*savefile.TurningPoint, len(src.StoryArc.TurningPoints))
			for i, tp := range src.StoryArc.TurningPoints {
				dest.StoryArc.TurningPoints[i] = &savefile.TurningPoint{
					Name:        tp.Name,
					Beat:        tp.Beat,
					Description: tp.Description,
				}
			}
		}
	}

	// Convert HiddenRules
	if src.HiddenRules != nil {
		dest.HiddenRules = make([]*savefile.HiddenRule, len(src.HiddenRules))
		for i, hr := range src.HiddenRules {
			dest.HiddenRules[i] = &savefile.HiddenRule{
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
		dest.NPCProfiles = make([]*savefile.NPCProfile, len(src.NPCProfiles))
		for i, np := range src.NPCProfiles {
			dest.NPCProfiles[i] = &savefile.NPCProfile{
				ID:           np.ID,
				Name:         np.Name,
				Archetype:    string(np.Archetype),
				Personality:  np.Personality,
				Appearance:   np.Appearance,
				Backstory:    np.Backstory,
				Skills:       np.Skills,
				Inventory:    np.Inventory,
				Secret:       np.Secret,
				Introduction: np.Introduction,
				LinkedSeeds:  np.LinkedSeeds,
				DeathTiming:  np.DeathTiming,
				Status:       string(np.Status),
				DeathBeat:    np.DeathBeat,
				DeathReason:  np.DeathReason,
				Description:  np.Description,
			}
		}
	}

	// Convert PossibleEndings
	if src.PossibleEndings != nil {
		dest.PossibleEndings = make([]*savefile.Ending, len(src.PossibleEndings))
		for i, e := range src.PossibleEndings {
			ending := &savefile.Ending{
				ID:                     e.ID,
				Name:                   e.Name,
				Type:                   e.Type,
				Description:            e.Description,
				RequiredSeedPercentage: e.RequiredSeedPercentage,
			}
			if e.Condition != nil {
				ending.Condition = &savefile.EndingCondition{
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
		dest.UsedTemplates = &savefile.UsedTemplates{
			Rules:  src.UsedTemplates.Rules,
			Scenes: src.UsedTemplates.Scenes,
		}
	}

	return dest
}

// SetSaveDir sets the save directory (for testing).
func (c *SaveCommandV2) SetSaveDir(dir string) {
	c.saveDir = dir
}
