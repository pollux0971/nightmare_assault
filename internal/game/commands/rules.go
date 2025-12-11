package commands

import (
	"strings"
)

// RulesCommand shows known game rules
type RulesCommand struct{}

// Name returns the command name
func (c *RulesCommand) Name() string {
	return "rules"
}

// Execute executes the rules command
func (c *RulesCommand) Execute(args []string) (string, error) {
	// TODO: Integrate with Story 3.1-3.2 hidden rules system
	// For now, return placeholder
	var output strings.Builder

	output.WriteString("=== 已知規則 / Known Rules ===\n\n")
	output.WriteString("📜 已揭露規則:\n")
	output.WriteString("  (目前無已知規則)\n\n")
	output.WriteString("💡 提示: 隱藏規則系統已實作，待整合至遊戲主循環\n")
	output.WriteString("💡 Hint: Hidden rules system is implemented, pending integration with main game loop\n")

	return output.String(), nil
}

// Help returns the command help text
func (c *RulesCommand) Help() string {
	return "顯示已知的遊戲規則 / Show known game rules"
}

// Usage returns the command usage
func (c *RulesCommand) Usage() string {
	return "/rules"
}

// Description returns the command description
func (c *RulesCommand) Description() string {
	return "查看所有已經發現的隱藏規則"
}
