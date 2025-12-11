package commands

import (
	"strings"
)

// DreamsCommand shows experienced dream sequences
type DreamsCommand struct{}

// Name returns the command name
func (c *DreamsCommand) Name() string {
	return "dreams"
}

// Execute executes the dreams command
func (c *DreamsCommand) Execute(args []string) (string, error) {
	// TODO: Integrate with Story 6.6 dream system
	// For now, return placeholder
	var output strings.Builder

	output.WriteString("=== 夢境記錄 / Dream Journal ===\n\n")
	output.WriteString("🌙 已經歷夢境:\n")
	output.WriteString("  (目前無夢境記錄)\n\n")
	output.WriteString("💡 提示: 夢境系統已實作，待整合至遊戲主循環\n")
	output.WriteString("💡 Hint: Dream system is implemented, pending integration with main game loop\n")

	return output.String(), nil
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
	return "回顧所有經歷過的夢境片段"
}
