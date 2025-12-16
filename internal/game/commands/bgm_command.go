package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// BGMCommand implements BGM control commands.
type BGMCommand struct {
	audioManager *audio.AudioManager
	config       *config.Config
}

// NewBGMCommand creates a new BGM command.
func NewBGMCommand(audioManager *audio.AudioManager, cfg *config.Config) *BGMCommand {
	return &BGMCommand{
		audioManager: audioManager,
		config:       cfg,
	}
}

// Name returns the command name.
func (c *BGMCommand) Name() string {
	return "bgm"
}

// Help returns help text for the command.
func (c *BGMCommand) Help() string {
	return `BGM control commands:
  /bgm on             - Enable BGM playback
  /bgm off            - Disable BGM playback (fade out 1s)
  /bgm volume <0-100> - Set BGM volume (0-100%)
  /bgm list           - List all available BGM tracks`
}

// Execute executes the BGM command.
func (c *BGMCommand) Execute(args []string) (string, error) {
	if len(args) == 0 {
		return c.Help(), nil
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "on":
		return c.enableBGM()
	case "off":
		return c.disableBGM()
	case "volume":
		if len(args) < 2 {
			return "", fmt.Errorf("usage: /bgm volume <0-100>")
		}
		return c.setVolume(args[1])
	case "list":
		return c.listBGM(), nil
	default:
		return "", fmt.Errorf("unknown subcommand: %s\nUse /bgm for help", subcommand)
	}
}

// enableBGM enables BGM playback.
func (c *BGMCommand) enableBGM() (string, error) {
	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.BGMPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	if player.IsEnabled() {
		return "ℹ️  BGM 已經啟用", nil
	}

	// Enable BGM
	player.Enable()

	// Update config
	c.config.Audio.BGMEnabled = true
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	// TODO: Play BGM for current scene with fade in (requires scene context)
	// For now, just enable it
	return "✅ BGM 已啟用", nil
}

// disableBGM disables BGM playback with fade out.
func (c *BGMCommand) disableBGM() (string, error) {
	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.BGMPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	if !player.IsEnabled() {
		return "ℹ️  BGM 已經停用", nil
	}

	// Fade out current BGM
	if player.IsPlaying() {
		if err := player.FadeOut(1 * time.Second); err != nil {
			return "", fmt.Errorf("failed to fade out BGM: %w", err)
		}
		player.Stop()
	}

	// Disable BGM
	player.Disable()

	// Update config
	c.config.Audio.BGMEnabled = false
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return "✅ BGM 已停用", nil
}

// setVolume sets the BGM volume.
func (c *BGMCommand) setVolume(volumeStr string) (string, error) {
	if c.audioManager == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	player := c.audioManager.BGMPlayer()
	if player == nil {
		return "⚠️  音訊系統未初始化", nil
	}

	// Parse volume (0-100)
	volumeInt, err := strconv.Atoi(volumeStr)
	if err != nil {
		return "", fmt.Errorf("invalid volume: %s (must be 0-100)", volumeStr)
	}

	if volumeInt < 0 || volumeInt > 100 {
		return "", fmt.Errorf("volume out of range: %d (must be 0-100)", volumeInt)
	}

	// Convert to 0.0-1.0 range
	volume := float64(volumeInt) / 100.0

	// Set volume
	player.SetVolume(volume)

	// Update config
	c.config.Audio.BGMVolume = volumeInt
	if err := c.config.Save(); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return fmt.Sprintf("🔊 BGM 音量設定為 %d%%", volumeInt), nil
}

// listBGM lists all available BGM tracks.
func (c *BGMCommand) listBGM() string {
	player := c.audioManager.BGMPlayer()

	var sb strings.Builder
	sb.WriteString("🎵 可用的 BGM 清單：\n\n")

	scenes := []audio.BGMScene{
		audio.BGMSceneExploration,
		audio.BGMSceneChase,
		audio.BGMSceneSafe,
		audio.BGMSceneHorror,
		audio.BGMSceneMystery,
		audio.BGMSceneDeath,
	}

	sceneNames := map[audio.BGMScene]string{
		audio.BGMSceneExploration: "探索場景",
		audio.BGMSceneChase:       "緊張/追逐場景",
		audio.BGMSceneSafe:        "安全區/休息場景",
		audio.BGMSceneHorror:      "恐怖揭露時刻",
		audio.BGMSceneMystery:     "解謎場景",
		audio.BGMSceneDeath:       "死亡/結局場景",
	}

	currentBGM := ""
	if player != nil {
		currentBGM = player.CurrentBGM()
	}

	for _, scene := range scenes {
		filename := audio.GetBGMFilename(scene)
		name := sceneNames[scene]

		// Check if this is currently playing
		marker := "  "
		if player != nil && strings.Contains(currentBGM, filename) {
			marker = "▶ "
		}

		sb.WriteString(fmt.Sprintf("%s%s - %s\n", marker, filename, name))
	}

	// Show status
	sb.WriteString("\n狀態：\n")
	if player == nil {
		sb.WriteString("⚠️  音訊系統未初始化\n")
	} else {
		if player.IsEnabled() {
			sb.WriteString(fmt.Sprintf("🔊 BGM 已啟用 (音量: %.0f%%)\n", player.Volume()*100))
		} else {
			sb.WriteString("🔇 BGM 已停用\n")
		}
		if player.IsPlaying() {
			sb.WriteString("▶️  正在播放\n")
		} else {
			sb.WriteString("⏸️  未播放\n")
		}
	}

	return sb.String()
}
