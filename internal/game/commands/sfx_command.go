package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// SFXCommand handles /sfx commands
type SFXCommand struct {
	audioManager *audio.AudioManager
	config       *config.Config
}

// NewSFXCommand creates a new SFX command
func NewSFXCommand(audioManager *audio.AudioManager, cfg *config.Config) *SFXCommand {
	return &SFXCommand{
		audioManager: audioManager,
		config:       cfg,
	}
}

// Name returns the command name
func (c *SFXCommand) Name() string {
	return "sfx"
}

// Help returns help text
func (c *SFXCommand) Help() string {
	return `SFX 音效控制指令：
  /sfx on      - 啟用音效
  /sfx off     - 停用音效
  /sfx volume <0-100> - 設定音效音量 (0-100)
  /sfx list    - 顯示可用音效列表`
}

// Execute executes the command
func (c *SFXCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		return c.Help(), nil
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "on":
		return c.enableSFX()
	case "off":
		return c.disableSFX()
	case "volume":
		if len(args) < 2 {
			return "", fmt.Errorf("usage: /sfx volume <0-100>")
		}
		return c.setVolume(args[1])
	case "list":
		return c.listSFX(), nil
	default:
		return "", fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *SFXCommand) enableSFX() (string, error) {
	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.SFXPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player.Enable()
	c.config.Audio.SFXEnabled = true
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return "✅ SFX 已啟用", nil
}

func (c *SFXCommand) disableSFX() (string, error) {
	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.SFXPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player.Disable()
	c.config.Audio.SFXEnabled = false
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return "✅ SFX 已停用", nil
}

func (c *SFXCommand) setVolume(volumeStr string) (string, error) {
	volume, err := strconv.Atoi(volumeStr)
	if err != nil {
		return "", fmt.Errorf("invalid volume: %s", volumeStr)
	}

	if volume < 0 || volume > 100 {
		return "", fmt.Errorf("volume must be between 0 and 100")
	}

	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.SFXPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	volumeFloat := float64(volume) / 100.0
	player.SetVolume(volumeFloat)
	c.config.Audio.SFXVolume = volumeFloat
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return fmt.Sprintf("✅ SFX 音量已設定為 %d%%", volume), nil
}

func (c *SFXCommand) listSFX() string {
	output := `📋 可用音效列表：

**環境音效** (低優先級 20):
  • door_open.wav - 門開啟
  • door_close.wav - 門關閉
  • footsteps.wav - 腳步聲
  • glass_break.wav - 玻璃碎裂
  • thunder.wav - 雷聲
  • whisper.wav - 耳語

**警告音效** (高優先級 80):
  • warning.wav - 規則觸發警告

**死亡音效** (最高優先級 100):
  • death.wav - 死亡音效

**心跳音效** (中優先級 50):
  • heartbeat.wav - 心跳聲 (循環)

---
狀態：`

	if c.audioManager == nil {
		output += "❌ 音訊系統未初始化"
		return output
	}

	player := c.audioManager.SFXPlayer()
	if player == nil {
		output += "❌ 音訊系統未初始化"
		return output
	}

	if player.IsEnabled() {
		output += fmt.Sprintf("✅ 啟用中 (音量: %.0f%%)", player.Volume()*100)
	} else {
		output += "❌ 已停用"
	}

	return output
}
