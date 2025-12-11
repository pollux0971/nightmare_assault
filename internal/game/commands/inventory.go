package commands

import (
	"strings"
)

// InventoryCommand shows player's inventory
type InventoryCommand struct{}

// Name returns the command name
func (c *InventoryCommand) Name() string {
	return "inventory"
}

// Execute executes the inventory command
func (c *InventoryCommand) Execute(args []string) (string, error) {
	// TODO: Implement actual inventory system (Epic 2 integration)
	// For now, return placeholder
	var output strings.Builder

	output.WriteString("=== 背包 / Inventory ===\n\n")
	output.WriteString("📦 物品清單:\n")
	output.WriteString("  (目前無物品)\n\n")
	output.WriteString("💡 提示: 背包系統將在遊戲核心功能完成後啟用\n")
	output.WriteString("💡 Hint: Inventory system will be enabled after core game features are completed\n")

	return output.String(), nil
}

// Help returns the command help text
func (c *InventoryCommand) Help() string {
	return "顯示背包物品 / Show inventory items"
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
