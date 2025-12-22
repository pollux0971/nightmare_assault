package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DreamsCommand shows experienced dream sequences.
// Story 9-2 AC: Implement /dreams command to review all dreams.
type DreamsCommand struct {
	gameState *engine.GameStateV2
}

// NewDreamsCommand creates a new dreams command.
func NewDreamsCommand(gameState *engine.GameStateV2) *DreamsCommand {
	return &DreamsCommand{
		gameState: gameState,
	}
}

// Name returns the command name
func (c *DreamsCommand) Name() string {
	return "dreams"
}

// Execute executes the dreams command.
// Story 9-2 AC: Display all experienced dreams with details.
func (c *DreamsCommand) Execute(args []string) (string, error) {
	if c.gameState == nil {
		return "遊戲狀態未初始化", nil
	}

	dreamLog := c.gameState.GetDreamLog()
	if dreamLog == nil || dreamLog.DreamCount() == 0 {
		return c.formatEmptyDreams(), nil
	}

	return c.formatDreamList(dreamLog), nil
}

// formatEmptyDreams formats the output when no dreams have been experienced.
func (c *DreamsCommand) formatEmptyDreams() string {
	var output strings.Builder

	output.WriteString("=== 夢境記錄 / Dream Journal ===\n\n")
	output.WriteString("🌙 你還沒有經歷過任何夢境。\n\n")
	output.WriteString("💡 提示: 夢境會在特定條件下觸發，它們可能包含重要線索。\n")

	return output.String()
}

// formatDreamList formats the list of dreams for display.
// Story 9-2 AC: Show dream number, type, beat, and preview.
func (c *DreamsCommand) formatDreamList(dreamLog *game.DreamLog) string {
	var output strings.Builder

	output.WriteString("=== 夢境記錄 / Dream Journal ===\n\n")
	output.WriteString(fmt.Sprintf("🌙 已經歷 %d 個夢境:\n\n", dreamLog.DreamCount()))

	dreams := dreamLog.Dreams
	for i, dream := range dreams {
		dreamNum := i + 1
		dreamType := formatDreamType(dream.Type)
		beatNum := dream.Context.ChapterNum

		// Dream header
		output.WriteString(fmt.Sprintf("#%d - %s (第 %d 章)\n", dreamNum, dreamType, beatNum))

		// Dream preview (first 100 chars)
		preview := dream.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		output.WriteString(fmt.Sprintf("   %s\n", preview))

		// Dream metadata
		if dream.RelatedRuleID != "" {
			output.WriteString(fmt.Sprintf("   🔍 關聯規則: %s\n", dream.RelatedRuleID))
		}

		output.WriteString(fmt.Sprintf("   📊 當時狀態: HP=%d SAN=%d\n",
			dream.Context.PlayerHP, dream.Context.PlayerSAN))

		output.WriteString("\n")
	}

	output.WriteString("💡 提示: 這些夢境可能包含隱藏規則的線索。\n")

	return output.String()
}

// formatDreamType converts dream type to display string.
func formatDreamType(dreamType game.DreamType) string {
	switch dreamType {
	case game.DreamTypeOpening:
		return "開場夢境"
	case game.DreamTypeChapter:
		return "章節夢境"
	default:
		return "未知夢境"
	}
}

// Help returns the command help text
func (c *DreamsCommand) Help() string {
	return "查看經歷過的夢境片段 / Show experienced dreams"
}

// Usage returns the command usage
func (c *DreamsCommand) Usage() string {
	return "/dreams"
}

// Description returns the command description
func (c *DreamsCommand) Description() string {
	return "回顧所有經歷過的夢境片段，夢境可能包含重要線索"
}
