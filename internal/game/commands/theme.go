// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// ThemeCommand handles the /theme command.
type ThemeCommand struct {
	config       *config.Config
	themeManager *themes.ThemeManager
}

// NewThemeCommand creates a new theme command handler.
func NewThemeCommand(cfg *config.Config) *ThemeCommand {
	return &ThemeCommand{
		config:       cfg,
		themeManager: themes.GetManager(),
	}
}

// Execute runs the /theme command and returns the result.
func (c *ThemeCommand) Execute(args string) CommandResult {
	args = strings.TrimSpace(args)

	switch {
	case args == "" || args == "list":
		return c.listThemes()
	case args == "current":
		return c.showCurrent()
	default:
		return c.switchTheme(args)
	}
}

func (c *ThemeCommand) listThemes() CommandResult {
	var b strings.Builder

	b.WriteString("🎨 **可用的主題**\n\n")

	allThemes := c.themeManager.GetAllThemes()
	for i, theme := range allThemes {
		marker := "  "
		if c.themeManager.IsCurrentTheme(theme.ID) {
			marker = "✓ "
		}
		b.WriteString(fmt.Sprintf("%s%d. **%s**\n", marker, i+1, theme.Name))
		b.WriteString(fmt.Sprintf("      %s\n", theme.Description))
	}

	b.WriteString("\n使用 `/theme <名稱>` 切換主題")
	b.WriteString("\n例如: `/theme blood_moon`")

	return CommandResult{Success: true, Message: b.String()}
}

func (c *ThemeCommand) showCurrent() CommandResult {
	current := c.themeManager.GetCurrentTheme()
	if current == nil {
		return CommandResult{
			Success: false,
			Message: "❌ 無法取得當前主題",
		}
	}

	return CommandResult{
		Success: true,
		Message: fmt.Sprintf("🎨 當前主題: **%s**\n%s", current.Name, current.Description),
	}
}

func (c *ThemeCommand) switchTheme(themeID string) CommandResult {
	// Normalize theme ID
	themeID = strings.ToLower(strings.TrimSpace(themeID))
	themeID = strings.ReplaceAll(themeID, " ", "_")

	// Check if theme exists
	if _, ok := c.themeManager.GetTheme(themeID); !ok {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("❌ 未知的主題: %s\n使用 `/theme list` 查看可用主題", themeID),
		}
	}

	// Apply theme
	if err := c.themeManager.SetTheme(themeID); err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("❌ 切換主題失敗: %v", err),
		}
	}

	// Save to config
	c.config.Theme = themeID
	if err := c.config.Save(); err != nil {
		// Theme applied but not saved
		return CommandResult{
			Success: true,
			Message: fmt.Sprintf("✓ 已切換至主題: %s\n⚠️ 儲存配置失敗，重啟後將恢復原設定", themeID),
			NeedsRedraw: true,
		}
	}

	theme, _ := c.themeManager.GetTheme(themeID)
	return CommandResult{
		Success:     true,
		Message:     fmt.Sprintf("✓ 已切換至主題: **%s**\n%s", theme.Name, theme.Description),
		NeedsRedraw: true,
	}
}

// Name returns the command name.
func (c *ThemeCommand) Name() string {
	return "theme"
}

// Help returns the help text.
func (c *ThemeCommand) Help() string {
	return "切換顏色主題"
}
