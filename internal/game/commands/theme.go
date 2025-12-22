// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// ThemeCommand handles the /theme command (Story 9-7).
// Allows players to switch between visual themes during gameplay.
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

// Name returns the command name.
func (c *ThemeCommand) Name() string {
	return "theme"
}

// Help returns the help text.
func (c *ThemeCommand) Help() string {
	return "切換顏色主題 - 用法: /theme [list|current|<名稱>]"
}

// Execute runs the /theme command and returns output or error.
// Story 9-7 AC2: Implements /theme command with list, current, and switch functionality.
func (c *ThemeCommand) Execute(args []string) (string, error) {
	// Convert args array to single string for easier parsing
	argsStr := strings.TrimSpace(strings.Join(args, " "))

	switch {
	case argsStr == "" || argsStr == "list":
		return c.listThemes(), nil
	case argsStr == "current":
		return c.showCurrent(), nil
	default:
		return c.switchTheme(argsStr)
	}
}

// listThemes lists all available themes (Story 9-7 AC2).
func (c *ThemeCommand) listThemes() string {
	var b strings.Builder

	b.WriteString("🎨 可用的主題\n\n")

	allThemes := c.themeManager.GetAllThemes()
	for i, theme := range allThemes {
		marker := "  "
		if c.themeManager.IsCurrentTheme(theme.ID) {
			marker = "✓ "
		}
		b.WriteString(fmt.Sprintf("%s%d. %s\n", marker, i+1, theme.Name))
		b.WriteString(fmt.Sprintf("      %s\n", theme.Description))
	}

	b.WriteString("\n使用 /theme <名稱> 切換主題")
	b.WriteString("\n例如: /theme blood_moon 或 /theme abyss_blue")

	return b.String()
}

// showCurrent shows the current active theme (Story 9-7 AC2).
func (c *ThemeCommand) showCurrent() string {
	current := c.themeManager.GetCurrentTheme()
	if current == nil {
		return "❌ 無法取得當前主題"
	}

	return fmt.Sprintf("🎨 當前主題: %s\n%s", current.Name, current.Description)
}

// switchTheme switches to a different theme (Story 9-7 AC3, AC4, AC5).
func (c *ThemeCommand) switchTheme(themeID string) (string, error) {
	// Normalize theme ID (AC3)
	themeID = strings.ToLower(strings.TrimSpace(themeID))
	themeID = strings.ReplaceAll(themeID, " ", "_")

	// Check if theme exists
	theme, ok := c.themeManager.GetTheme(themeID)
	if !ok {
		return "", fmt.Errorf("未知的主題: %s\n使用 /theme list 查看可用主題", themeID)
	}

	// Apply theme (AC4: 主題切換應即時生效)
	if err := c.themeManager.SetTheme(themeID); err != nil {
		return "", fmt.Errorf("切換主題失敗: %v", err)
	}

	// Save to config (AC5: 主題偏好應保存到配置檔案)
	c.config.Theme = themeID
	if err := c.config.Save(); err != nil {
		// Theme applied but not saved - warn user
		return fmt.Sprintf("✓ 已切換至主題: %s\n%s\n\n⚠️ 儲存配置失敗，重啟後將恢復原設定",
			theme.Name, theme.Description), nil
	}

	return fmt.Sprintf("✓ 已切換至主題: %s\n%s\n\n主題已生效並保存",
		theme.Name, theme.Description), nil
}
