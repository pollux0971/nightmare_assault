package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// CluesCommand shows discovered clues
type CluesCommand struct {
	gameState *engine.GameStateV2
}

// NewCluesCommand creates a new clues command.
func NewCluesCommand(gameState *engine.GameStateV2) *CluesCommand {
	return &CluesCommand{
		gameState: gameState,
	}
}

// Name returns the command name
func (c *CluesCommand) Name() string {
	return "clues"
}

// Execute executes the clues command
func (c *CluesCommand) Execute(args []string) (string, error) {
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("              CLUES / 線索\n")
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	if c.gameState == nil {
		output.WriteString("🔍 已發現線索:\n")
		output.WriteString("  (遊戲狀態未初始化)\n")
		return output.String(), nil
	}

	// Get Global Seeds (Dream/Story clues)
	globalSeeds := c.gameState.GetGlobalSeeds()

	// Get Local Seeds (Environmental/NPC clues)
	localSeeds := c.gameState.GetLocalSeeds()

	// Count revealed clues
	totalClues := 0

	// Display Dream/Story Clues (Global Seeds)
	dreamClues := make([]string, 0)
	for _, seed := range globalSeeds {
		// Show all revealed clues (from tier 1 up to current tier)
		// CurrentTier starts at 1, and after AdvanceTier() it becomes 2, etc.
		// We want to show all clues from tier 1 to (CurrentTier - 1)
		for tier := 1; tier < seed.CurrentTier; tier++ {
			if tier <= len(seed.ClueChain) {
				clue := &seed.ClueChain[tier-1]
				dreamClues = append(dreamClues, fmt.Sprintf("「%s」", clue.Content))
				totalClues++
			}
		}
	}

	// Display Environmental/Scene Clues (Local Seeds)
	envClues := make([]string, 0)
	for _, seed := range localSeeds {
		// Only show harvested clues
		if seed.IsHarvested {
			envClues = append(envClues, fmt.Sprintf("「%s」", seed.Content))
			totalClues++
		}
	}

	// NPC Clues - For now, we'll use a subset of local seeds
	// In the future, this could be a separate category in GameStateV2
	npcClues := make([]string, 0)
	// TODO: Add dedicated NPC clue tracking in Epic 6

	if totalClues == 0 {
		output.WriteString("🔍 已發現線索:\n")
		output.WriteString("  (目前無線索 / No clues discovered)\n\n")
		output.WriteString("提示: 仔細觀察環境，與 NPC 對話，注意夢境細節\n")
		output.WriteString("Hint: Observe the environment, talk to NPCs, pay attention to dream details\n")
	} else {
		output.WriteString(fmt.Sprintf("🔍 已發現線索 (共 %d 條):\n\n", totalClues))

		// Dream/Story Clues
		if len(dreamClues) > 0 {
			output.WriteString("【夢境線索 / Dream Clues】\n")
			for i, clue := range dreamClues {
				output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, clue))
			}
			output.WriteString("\n")
		}

		// Environmental Clues
		if len(envClues) > 0 {
			output.WriteString("【環境線索 / Environmental Clues】\n")
			for i, clue := range envClues {
				output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, clue))
			}
			output.WriteString("\n")
		}

		// NPC Clues
		if len(npcClues) > 0 {
			output.WriteString("【NPC 線索 / NPC Clues】\n")
			for i, clue := range npcClues {
				output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, clue))
			}
			output.WriteString("\n")
		}

		output.WriteString("提示: 線索之間可能存在關聯，仔細思考它們的意義\n")
		output.WriteString("Hint: Clues may be related - think carefully about their meaning\n")
	}

	output.WriteString("\n═══════════════════════════════════════════════════\n")
	output.WriteString("Press any key to return to game...\n")

	return output.String(), nil
}

// Help returns the command help text
func (c *CluesCommand) Help() string {
	return "Show discovered clues categorized by type / 顯示已發現的線索（分類顯示）"
}

// Usage returns the command usage
func (c *CluesCommand) Usage() string {
	return "/clues"
}

// Description returns the command description
func (c *CluesCommand) Description() string {
	return "查看所有已經發現的線索（夢境、環境、NPC）"
}
