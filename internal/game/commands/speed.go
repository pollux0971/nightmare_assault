package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// SpeedCommand implements the /speed command for typewriter effect control.
type SpeedCommand struct {
	config *config.Config
}

// NewSpeedCommand creates a new speed command.
func NewSpeedCommand(cfg *config.Config) *SpeedCommand {
	return &SpeedCommand{
		config: cfg,
	}
}

// Name returns the command name.
func (c *SpeedCommand) Name() string {
	return "speed"
}

// Execute runs the /speed command.
// Usage:
//   /speed - Show current speed
//   /speed on - Enable typewriter effect
//   /speed off - Disable typewriter effect
//   /speed <number> - Set speed (10-200 chars/sec)
func (c *SpeedCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		// Show current status
		return c.showStatus(), nil
	}

	arg := strings.ToLower(args[0])

	switch arg {
	case "on":
		return c.enable()
	case "off":
		return c.disable()
	default:
		// Try to parse as number
		speed, err := strconv.Atoi(arg)
		if err != nil {
			return "", fmt.Errorf("無效的參數。用法: /speed [on|off|<10-200>]")
		}
		return c.setSpeed(speed)
	}
}

// Help returns help text.
func (c *SpeedCommand) Help() string {
	return "控制打字機效果 - 用法: /speed [on|off|<10-200>]"
}

func (c *SpeedCommand) showStatus() string {
	status := "已停用"
	if c.config.Typewriter.Enabled {
		status = "已啟用"
	}

	cursor := "隱藏"
	if c.config.Typewriter.ShowCursor {
		cursor = "顯示"
	}

	return fmt.Sprintf(`打字機效果狀態:
  狀態: %s
  速度: %d 字/秒
  游標: %s

用法:
  /speed on      - 啟用效果
  /speed off     - 停用效果
  /speed <數字>  - 設定速度 (10-200)`, status, c.config.Typewriter.Speed, cursor)
}

func (c *SpeedCommand) enable() (string, error) {
	c.config.Typewriter.Enabled = true

	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	return fmt.Sprintf("✓ 打字機效果已啟用 (速度: %d 字/秒)", c.config.Typewriter.Speed), nil
}

func (c *SpeedCommand) disable() (string, error) {
	c.config.Typewriter.Enabled = false

	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	return "✓ 打字機效果已停用", nil
}

func (c *SpeedCommand) setSpeed(speed int) (string, error) {
	// Validate range (AC6)
	if speed < 10 || speed > 200 {
		// Use default value and show error
		c.config.Typewriter.Speed = 40
		if err := c.config.Save(); err != nil {
			return "", fmt.Errorf("儲存設定失敗: %v", err)
		}
		return "⚠ 速度超出範圍 (10-200),已使用預設值 40", nil
	}

	c.config.Typewriter.Speed = speed
	c.config.Typewriter.Enabled = true // Enable when setting speed

	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("儲存設定失敗: %v", err)
	}

	return fmt.Sprintf("✓ 打字機速度已設定為 %d 字/秒", speed), nil
}
