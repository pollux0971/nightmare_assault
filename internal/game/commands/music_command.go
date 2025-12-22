package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// MusicCommand 實作自定義 BGM 控制指令
type MusicCommand struct {
	audioManager *audio.AudioManager
	config       *config.Config
	customBGMMgr *audio.CustomBGMManager
}

// NewMusicCommand 創建新的音樂指令
func NewMusicCommand(audioManager *audio.AudioManager, cfg *config.Config) *MusicCommand {
	var customMgr *audio.CustomBGMManager

	// 初始化自定義 BGM 管理器
	if audioManager != nil {
		// 使用 ~/.nightmare/bgm/custom/ 作為自定義音樂目錄 (AC 10.2 規格)
		customDir := filepath.Join(audioManager.AudioDir(), "bgm", "custom")
		customMgr = audio.NewCustomBGMManager(customDir)

		// 掃描自定義音樂檔案
		if err := customMgr.ScanCustomDirectory(); err != nil {
			// 錯誤不阻塞，只記錄
			fmt.Printf("⚠️  掃描自定義音樂失敗: %v\n", err)
		}
	}

	return &MusicCommand{
		audioManager: audioManager,
		config:       cfg,
		customBGMMgr: customMgr,
	}
}

// Name 返回指令名稱
func (c *MusicCommand) Name() string {
	return "music"
}

// Help 返回指令幫助訊息
func (c *MusicCommand) Help() string {
	return `自定義 BGM 控制指令：
  /music list          - 列出所有可用音樂
  /music play <名稱>   - 播放指定音樂
  /music stop          - 停止播放
  /music loop on       - 啟用循環播放
  /music loop off      - 停用循環播放
  /music set <場景> <檔名> - 設定場景專用 BGM

自定義音樂設置：
  1. 將 MP3/WAV/OGG 檔案放到 ~/.nightmare/bgm/custom/
  2. 檔案大小限制 20MB
  3. 命名規則: custom_<場景>.mp3 可自動覆蓋場景 BGM
     (如 custom_explore.mp3, custom_tension.mp3)
  4. 使用 /music list 查看可用音樂

場景名稱：
  explore (探索), tension (緊張), safe (安全)
  horror (恐怖), mystery (解謎), ending (結局)

注意：
  - 自定義音樂會覆蓋預設 BGM
  - 停止播放後會恢復預設 BGM
  - 循環播放預設啟用`
}

// Aliases 返回指令別名
func (c *MusicCommand) Aliases() []string {
	return []string{"音樂", "歌曲"}
}

// Execute 執行音樂指令
func (c *MusicCommand) Execute(args []string) (string, error) {
	// 預設行為：顯示幫助
	if len(args) == 0 {
		return c.Help(), nil
	}

	subcommand := strings.ToLower(args[0])

	// 檢查音訊系統是否可用（除了 help 和 loop 之外的指令需要）
	if c.audioManager == nil && subcommand != "loop" {
		return "⚠️  音訊系統未初始化", nil
	}

	if c.customBGMMgr == nil && subcommand != "loop" {
		return "⚠️  自定義 BGM 管理器未初始化", nil
	}

	switch subcommand {
	case "list", "列表":
		return c.listMusic()
	case "play", "播放":
		if len(args) < 2 {
			return "", fmt.Errorf("用法: /music play <音樂名稱>")
		}
		return c.playMusic(args[1])
	case "stop", "停止":
		return c.stopMusic()
	case "loop", "循環":
		if len(args) < 2 {
			return c.showLoopStatus()
		}
		return c.setLoop(args[1])
	case "set", "設定":
		if len(args) < 3 {
			return "", fmt.Errorf("用法: /music set <場景> <檔名>\n場景: explore, tension, safe, horror, mystery, ending")
		}
		return c.setMoodBGM(args[1], args[2])
	default:
		return "", fmt.Errorf("未知子指令: %s\n使用 /music 查看幫助", subcommand)
	}
}

// listMusic 列出所有可用的自定義音樂
func (c *MusicCommand) listMusic() (string, error) {
	var output strings.Builder

	output.WriteString("🎵 可用的自定義音樂：\n\n")

	// 獲取可用檔案
	files := c.customBGMMgr.GetAvailableFiles()

	if len(files) == 0 {
		output.WriteString("⚠️  沒有找到自定義音樂檔案\n\n")
		output.WriteString("請將音樂檔案 (MP3/WAV/OGG) 放到以下目錄：\n")
		output.WriteString("  ~/.nightmare/bgm/custom/\n\n")
		output.WriteString("檔案限制：\n")
		output.WriteString("  - 支援格式: MP3, WAV, OGG\n")
		output.WriteString("  - 最大大小: 20MB\n")
		output.WriteString("\n命名規則：\n")
		output.WriteString("  - custom_<場景>.mp3 可自動覆蓋場景 BGM\n")
		output.WriteString("  - 如 custom_explore.mp3, custom_tension.mp3\n")
		return output.String(), nil
	}

	// 顯示檔案列表
	output.WriteString(fmt.Sprintf("找到 %d 個音樂檔案：\n\n", len(files)))
	for i, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		name := strings.TrimSuffix(file, ext)
		output.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, name, strings.TrimPrefix(ext, ".")))
	}

	output.WriteString("\n使用方法：\n")
	output.WriteString("  /music play <檔名>   - 播放指定音樂\n")
	output.WriteString("  /music stop          - 停止並恢復預設 BGM\n")

	return output.String(), nil
}

// playMusic 播放指定的自定義音樂
func (c *MusicCommand) playMusic(filename string) (string, error) {
	player := c.audioManager.BGMPlayer()
	if player == nil {
		return "⚠️  BGM 播放器未初始化", nil
	}

	// 檢查檔案是否存在
	files := c.customBGMMgr.GetAvailableFiles()
	var fullFilename string
	found := false

	// 允許使用不帶副檔名的檔名
	for _, file := range files {
		nameWithoutExt := strings.TrimSuffix(file, filepath.Ext(file))
		if file == filename || nameWithoutExt == filename {
			fullFilename = file
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("找不到音樂檔案: %s\n使用 /music list 查看可用音樂", filename)
	}

	// 構建完整路徑 (使用 AC 規格的 ~/.nightmare/bgm/custom/)
	customDir := filepath.Join(c.audioManager.AudioDir(), "bgm", "custom")
	fullPath := filepath.Join(customDir, fullFilename)

	// 驗證檔案
	if err := audio.ValidateCustomAudioFile(fullPath); err != nil {
		// AC2: 檔案格式錯誤時回退到默認 BGM 並記錄警告
		c.logWarning(fmt.Sprintf("自定義音樂檔案無效: %v, 使用預設 BGM", err))
		return "⚠️  音樂檔案無效，已回退到預設 BGM", nil
	}

	// 播放自定義音樂 (使用絕對路徑)
	if err := player.Play(fullPath); err != nil {
		// AC2: 播放失敗時回退到默認 BGM
		c.logWarning(fmt.Sprintf("播放自定義音樂失敗: %v, 使用預設 BGM", err))
		return "⚠️  播放失敗，已回退到預設 BGM", nil
	}

	return fmt.Sprintf("🎵 正在播放: %s", fullFilename), nil
}

// logWarning 記錄警告日誌
func (c *MusicCommand) logWarning(msg string) {
	// 記錄到 debug.log
	if c.audioManager != nil {
		logPath := filepath.Join(filepath.Dir(c.audioManager.AudioDir()), "debug.log")
		if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer f.Close()
			fmt.Fprintf(f, "[WARN] [CUSTOM_BGM] %s\n", msg)
		}
	}
}

// stopMusic 停止播放自定義音樂並恢復預設 BGM
func (c *MusicCommand) stopMusic() (string, error) {
	player := c.audioManager.BGMPlayer()
	if player == nil {
		return "⚠️  BGM 播放器未初始化", nil
	}

	// 停止當前播放
	if player.IsPlaying() {
		player.Stop()
	}

	// 恢復預設 BGM (根據當前 mood)
	currentMood := player.GetCurrentMood()
	defaultBGM := audio.GetBGMForMood(currentMood)

	if err := player.Play(defaultBGM); err != nil {
		c.logWarning(fmt.Sprintf("恢復預設 BGM 失敗: %v", err))
		return "🔇 已停止播放自定義音樂\n⚠️  恢復預設 BGM 失敗", nil
	}

	return fmt.Sprintf("🔇 已停止自定義音樂，恢復預設 BGM (%s)", defaultBGM), nil
}

// showLoopStatus 顯示當前循環播放狀態
func (c *MusicCommand) showLoopStatus() (string, error) {
	if c.config == nil {
		return "⚠️  配置未初始化", nil
	}

	var output strings.Builder

	output.WriteString("🔁 循環播放狀態：\n\n")

	if c.config.Audio.BGMLoop {
		output.WriteString("  ✅ 已啟用\n\n")
	} else {
		output.WriteString("  ❌ 已停用\n\n")
	}

	output.WriteString("使用方法：\n")
	output.WriteString("  /music loop on   - 啟用循環播放\n")
	output.WriteString("  /music loop off  - 停用循環播放\n")

	return output.String(), nil
}

// setLoop 設定循環播放
func (c *MusicCommand) setLoop(value string) (string, error) {
	if c.config == nil {
		return "⚠️  配置未初始化", nil
	}

	value = strings.ToLower(value)

	switch value {
	case "on", "true", "1", "啟用", "開":
		c.config.Audio.BGMLoop = true
		if err := c.config.Save(); err != nil {
			return "", fmt.Errorf("儲存配置失敗: %w", err)
		}
		return "🔁 已啟用循環播放\n\n" +
			"所有 BGM 將會循環播放", nil

	case "off", "false", "0", "停用", "關":
		c.config.Audio.BGMLoop = false
		if err := c.config.Save(); err != nil {
			return "", fmt.Errorf("儲存配置失敗: %w", err)
		}
		return "⏸️  已停用循環播放\n\n" +
			"BGM 播放完畢後將會停止", nil

	default:
		return "", fmt.Errorf("無效的值: %s\n請使用 on 或 off", value)
	}
}

// setMoodBGM 設定特定場景的自定義 BGM
func (c *MusicCommand) setMoodBGM(sceneName, filename string) (string, error) {
	// 解析場景名稱到 MoodType
	mood, err := parseMoodName(sceneName)
	if err != nil {
		return "", err
	}

	// 如果 filename 是 "default"，則重置為預設
	if strings.ToLower(filename) == "default" {
		c.customBGMMgr.ResetToDefault(mood)
		return fmt.Sprintf("✅ 已將 %s 場景重置為預設 BGM", sceneName), nil
	}

	// 設定自定義 BGM
	if err := c.customBGMMgr.SetMoodBGM(mood, filename); err != nil {
		return "", fmt.Errorf("設定失敗: %w", err)
	}

	return fmt.Sprintf("✅ 已將 %s 場景設定為: %s", sceneName, filename), nil
}

// parseMoodName 解析場景名稱到 MoodType
func parseMoodName(name string) (engine.MoodType, error) {
	name = strings.ToLower(name)
	switch name {
	case "explore", "exploration", "探索":
		return engine.MoodExploration, nil
	case "tension", "緊張":
		return engine.MoodTension, nil
	case "safe", "安全":
		return engine.MoodSafe, nil
	case "horror", "恐怖":
		return engine.MoodHorror, nil
	case "mystery", "puzzle", "解謎":
		return engine.MoodMystery, nil
	case "ending", "結局":
		return engine.MoodEnding, nil
	default:
		return 0, fmt.Errorf("未知場景: %s\n支援: explore, tension, safe, horror, mystery, ending", name)
	}
}
