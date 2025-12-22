package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// InventoryCommand shows player's inventory
type InventoryCommand struct {
	gameState *engine.GameStateV2
}

// NewInventoryCommand creates a new inventory command.
func NewInventoryCommand(gameState *engine.GameStateV2) *InventoryCommand {
	return &InventoryCommand{
		gameState: gameState,
	}
}

// Name returns the command name
func (c *InventoryCommand) Name() string {
	return "inventory"
}

// Execute executes the inventory command
func (c *InventoryCommand) Execute(args []string) (string, error) {
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("              INVENTORY / 背包\n")
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	if c.gameState == nil {
		output.WriteString("📦 物品清單:\n")
		output.WriteString("  (遊戲狀態未初始化)\n")
		return output.String(), nil
	}

	inventory := c.gameState.Inventory
	if len(inventory) == 0 {
		output.WriteString("📦 物品清單:\n")
		output.WriteString("  (目前無物品 / Empty)\n\n")
		output.WriteString("提示: 在遊戲中探索並收集物品\n")
		output.WriteString("Hint: Explore and collect items during gameplay\n")
	} else {
		output.WriteString(fmt.Sprintf("📦 物品清單 (共 %d 件):\n\n", len(inventory)))

		// Categorize items by type (for now, we just list them)
		// In future, we could add item metadata for categorization
		for i, item := range inventory {
			output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, item))
		}

		output.WriteString("\n提示: 使用物品名稱與 NPC 互動或解決謎題\n")
		output.WriteString("Hint: Use item names to interact with NPCs or solve puzzles\n")
	}

	output.WriteString("\n═══════════════════════════════════════════════════\n")
	output.WriteString("Press any key to return to game...\n")

	return output.String(), nil
}

// Help returns the command help text
func (c *InventoryCommand) Help() string {
	return "Show inventory items / 顯示背包物品"
}

// Aliases returns command aliases
func (c *InventoryCommand) Aliases() []string {
	return []string{"inv", "i"}
}

// Usage returns the command usage
func (c *InventoryCommand) Usage() string {
	return "/inventory"
}

// Description returns the command description
func (c *InventoryCommand) Description() string {
	return "查看背包中的所有物品"
}
