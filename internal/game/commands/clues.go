package commands

import (
	"strings"
)

// CluesCommand shows discovered clues
type CluesCommand struct{}

// Name returns the command name
func (c *CluesCommand) Name() string {
	return "clues"
}

// Execute executes the clues command
func (c *CluesCommand) Execute(args []string) (string, error) {
	// TODO: Implement actual clues system (Epic 2/3 integration)
	// For now, return placeholder
	var output strings.Builder

	output.WriteString("=== 線索 / Clues ===\n\n")
	output.WriteString("🔍 已發現線索:\n")
	output.WriteString("  (目前無線索)\n\n")
	output.WriteString("💡 提示: 線索系統將在遊戲核心功能完成後啟用\n")
	output.WriteString("💡 Hint: Clues system will be enabled after core game features are completed\n")

	return output.String(), nil
}

// Help returns the command help text
func (c *CluesCommand) Help() string {
	return "顯示已發現的線索 / Show discovered clues"
}

// Usage returns the command usage
func (c *CluesCommand) Usage() string {
	return "/clues"
}

// Description returns the command description
func (c *CluesCommand) Description() string {
	return "查看所有已經發現的線索"
}
