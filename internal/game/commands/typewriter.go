// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TypewriterCommand implements the /typewriter command (Story 9-8).
// Allows control over typewriter effect settings during gameplay.
type TypewriterCommand struct {
	config *config.Config
}

// NewTypewriterCommand creates a new typewriter command.
func NewTypewriterCommand(cfg *config.Config) *TypewriterCommand {
	return &TypewriterCommand{
		config: cfg,
	}
}

// Name returns the command name.
func (c *TypewriterCommand) Name() string {
	return "typewriter"
}

// Aliases returns command aliases.
func (c *TypewriterCommand) Aliases() []string {
	return []string{"tw", "type"}
}

// Help returns help text.
func (c *TypewriterCommand) Help() string {
	return "控制打字機效果 - 用法: /typewriter [on|off|speed N]"
}

// Execute runs the /typewriter command (Story 9-8 AC4).
// Usage:
//   /typewriter - Show current status
//   /typewriter on - Enable typewriter effect
//   /typewriter off - Disable typewriter effect
//   /typewriter speed N - Set speed (10-200 chars/sec)
func (c *TypewriterCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		// Show current status
		return c.showStatus(), nil
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "on", "enable":
		return c.enable()
	case "off", "disable":
		return c.disable()
	case "speed":
		if len(args) < 2 {
			return "", fmt.Errorf("用法: /typewriter speed <10-200>")
		}
		speed, err := strconv.Atoi(args[1])
		if err != nil {
			return "", fmt.Errorf("無效的速度值: %s", args[1])
		}
		return c.setSpeed(speed)
	case "status":
		return c.showStatus(), nil
	default:
		// Try to parse as direct speed value
		if speed, err := strconv.Atoi(subcommand); err == nil {
			return c.setSpeed(speed)
		}
		return "", fmt.Errorf("無效的參數: %s\n用法: /typewriter [on|off|speed N]", subcommand)
	}
}

// showStatus shows current typewriter settings (Story 9-8 AC4).
func (c *TypewriterCommand) showStatus() string {
	status := "已停用"
	if c.config.Typewriter.Enabled {
		status = "已啟用"
	}

	// Calculate delay per character in milliseconds (Story 9-8 AC2)
	msPerChar := 1000 / c.config.Typewriter.Speed

	var b strings.Builder
	b.WriteString("⌨️ 打字機效果狀態\n\n")
	b.WriteString(fmt.Sprintf("狀態: %s\n", status))
	b.WriteString(fmt.Sprintf("速度: %d 字/秒 (%dms/字)\n", c.config.Typewriter.Speed, msPerChar))
	b.WriteString(fmt.Sprintf("游標: %s\n", map[bool]string{true: "顯示", false: "隱藏"}[c.config.Typewriter.ShowCursor]))
	b.WriteString("\n用法:\n")
	b.WriteString("  /typewriter on      - 啟用效果\n")
	b.WriteString("  /typewriter off     - 停用效果\n")
	b.WriteString("  /typewriter speed N - 設定速度 (10-200)\n")
	b.WriteString("\n提示: 按任意鍵可跳過打字效果")

	return b.String()
}

// enable enables the typewriter effect (Story 9-8 AC4).
func (c *TypewriterCommand) enable() (string, error) {
	c.config.Typewriter.Enabled = true

	// Save to config (Story 9-8 AC5)
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	return fmt.Sprintf("✓ 打字機效果已啟用\n速度: %d 字/秒", c.config.Typewriter.Speed), nil
}

// disable disables the typewriter effect (Story 9-8 AC4).
func (c *TypewriterCommand) disable() (string, error) {
	c.config.Typewriter.Enabled = false

	// Save to config (Story 9-8 AC5)
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	return "✓ 打字機效果已停用\n文字將立即顯示", nil
}

// setSpeed sets the typewriter speed (Story 9-8 AC2, AC4, AC5).
func (c *TypewriterCommand) setSpeed(speed int) (string, error) {
	// Validate range: 10-200 chars/sec (Story 9-8 AC2: 預設30-50ms = 20-33 chars/sec)
	// Extended range for flexibility
	if speed < 10 || speed > 200 {
		return "", fmt.Errorf("速度必須在 10-200 字/秒之間\n推薦值: 30-50 (每字符 20-33ms)")
	}

	c.config.Typewriter.Speed = speed
	c.config.Typewriter.Enabled = true // Enable when setting speed

	// Save to config (Story 9-8 AC5)
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	msPerChar := 1000 / speed
	return fmt.Sprintf("✓ 打字機速度已設定為 %d 字/秒 (%dms/字)\n打字機效果已啟用", speed, msPerChar), nil
}

// CalculateDelay calculates the delay in milliseconds per character.
// Story 9-8 AC2: Default speed should be 30-50ms per character.
func (c *TypewriterCommand) CalculateDelay() int {
	if !c.config.Typewriter.Enabled {
		return 0
	}
	return 1000 / c.config.Typewriter.Speed
}

// ShouldDelayAfterPunctuation checks if there should be extra delay after punctuation (Story 9-8 AC6).
func ShouldDelayAfterPunctuation(char rune) bool {
	// Chinese and English punctuation that should have extra delay
	punctuation := "。！？；：，、.!?;:,"
	return strings.ContainsRune(punctuation, char)
}

// GetPunctuationDelay returns the extra delay in milliseconds after punctuation (Story 9-8 AC6).
func GetPunctuationDelay() int {
	return 300 // 300ms extra delay after punctuation
}
