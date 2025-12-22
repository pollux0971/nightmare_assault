// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// AudioSettingType represents the type of audio setting being adjusted.
type AudioSettingType int

const (
	AudioSettingMasterVolume AudioSettingType = iota
	AudioSettingBGMVolume
	AudioSettingSFXVolume
	AudioSettingBGMEnabled
	AudioSettingSFXEnabled
	AudioSettingBack
)

// AudioSettingsModel represents the audio settings state.
type AudioSettingsModel struct {
	config        *config.Config
	audioManager  *audio.AudioManager
	selectedIndex int
	width         int
	height        int
	done          bool
	saved         bool
}

// NewAudioSettingsModel creates a new audio settings model.
func NewAudioSettingsModel(cfg *config.Config, audioMgr *audio.AudioManager) AudioSettingsModel {
	return AudioSettingsModel{
		config:        cfg,
		audioManager:  audioMgr,
		selectedIndex: 0,
	}
}

// Init initializes the model.
func (m AudioSettingsModel) Init() tea.Cmd {
	return nil
}

// AudioSettingsSavedMsg is sent when audio settings are saved.
type AudioSettingsSavedMsg struct{}

// Update handles messages.
func (m AudioSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Save and go back
			m.saveConfig()
			m.done = true
			return m, func() tea.Msg {
				return AudioSettingsSavedMsg{}
			}

		case "up", "k":
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = 5 // 6 items total (0-5)
			}
			return m, nil

		case "down", "j":
			m.selectedIndex++
			if m.selectedIndex > 5 {
				m.selectedIndex = 0
			}
			return m, nil

		case "left", "h":
			m.adjustValue(-5)
			return m, nil

		case "right", "l":
			m.adjustValue(5)
			return m, nil

		case "enter", " ":
			if m.selectedIndex == int(AudioSettingBack) {
				m.saveConfig()
				m.done = true
				return m, func() tea.Msg {
					return AudioSettingsSavedMsg{}
				}
			}
			// Toggle for enable/disable options
			if m.selectedIndex == int(AudioSettingBGMEnabled) {
				m.config.Audio.BGMEnabled = !m.config.Audio.BGMEnabled
				m.applyBGMEnabled()
			} else if m.selectedIndex == int(AudioSettingSFXEnabled) {
				m.config.Audio.SFXEnabled = !m.config.Audio.SFXEnabled
			}
			return m, nil

		case "1":
			m.selectedIndex = 0
			return m, nil
		case "2":
			m.selectedIndex = 1
			return m, nil
		case "3":
			m.selectedIndex = 2
			return m, nil
		case "4":
			m.selectedIndex = 3
			return m, nil
		case "5":
			m.selectedIndex = 4
			return m, nil
		case "6":
			m.selectedIndex = 5
			return m, nil
		}
	}

	return m, nil
}

func (m *AudioSettingsModel) adjustValue(delta int) {
	switch AudioSettingType(m.selectedIndex) {
	case AudioSettingMasterVolume:
		m.config.Audio.MasterVolume = clampVolume(m.config.Audio.MasterVolume + delta)
		m.applyVolume()
	case AudioSettingBGMVolume:
		m.config.Audio.BGMVolume = clampVolume(m.config.Audio.BGMVolume + delta)
		m.applyVolume()
	case AudioSettingSFXVolume:
		m.config.Audio.SFXVolume = clampVolume(m.config.Audio.SFXVolume + delta)
	case AudioSettingBGMEnabled:
		m.config.Audio.BGMEnabled = !m.config.Audio.BGMEnabled
		m.applyBGMEnabled()
	case AudioSettingSFXEnabled:
		m.config.Audio.SFXEnabled = !m.config.Audio.SFXEnabled
	}
}

func clampVolume(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func (m *AudioSettingsModel) applyVolume() {
	if m.audioManager == nil || !m.audioManager.IsInitialized() {
		return
	}

	// Apply BGM volume
	if bgmPlayer := m.audioManager.BGMPlayer(); bgmPlayer != nil {
		// Calculate effective volume: master * bgm / 100
		effectiveVolume := float64(m.config.Audio.MasterVolume) * float64(m.config.Audio.BGMVolume) / 10000.0
		bgmPlayer.SetVolume(effectiveVolume)
	}
}

func (m *AudioSettingsModel) applyBGMEnabled() {
	if m.audioManager == nil || !m.audioManager.IsInitialized() {
		return
	}

	if bgmPlayer := m.audioManager.BGMPlayer(); bgmPlayer != nil {
		if m.config.Audio.BGMEnabled {
			bgmPlayer.Enable()
			// Resume playing if not already playing
			if !bgmPlayer.IsPlaying() {
				go bgmPlayer.Play(audio.GetBGMFilename(audio.BGMSceneMystery))
			}
		} else {
			bgmPlayer.Disable()
			bgmPlayer.Stop()
		}
	}
}

func (m *AudioSettingsModel) saveConfig() {
	if m.config != nil {
		m.config.Save()
		m.saved = true
	}
}

// View renders the audio settings.
func (m AudioSettingsModel) View() string {
	var b strings.Builder

	// Get theme colors
	tm := themes.GetManager()
	theme := tm.GetCurrentTheme()
	colors := theme.Colors

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("🔊 音效設定"))
	b.WriteString("\n\n")

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	b.WriteString(descStyle.Render("使用 ←/→ 或 h/l 調整音量，Enter 或 Space 切換開關"))
	b.WriteString("\n\n")

	// Settings items
	selectedStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(colors.Primary)
	valueStyle := lipgloss.NewStyle().
		Foreground(colors.Success)
	disabledStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)

	items := []struct {
		label   string
		value   string
		enabled bool
	}{
		{"主音量", m.renderVolumeBar(m.config.Audio.MasterVolume), true},
		{"BGM 音量", m.renderVolumeBar(m.config.Audio.BGMVolume), m.config.Audio.BGMEnabled},
		{"音效音量", m.renderVolumeBar(m.config.Audio.SFXVolume), m.config.Audio.SFXEnabled},
		{"BGM 開關", m.renderToggle(m.config.Audio.BGMEnabled), true},
		{"音效開關", m.renderToggle(m.config.Audio.SFXEnabled), true},
		{"返回", "", true},
	}

	for i, item := range items {
		prefix := "  "
		style := normalStyle

		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}

		if !item.enabled && i < 3 { // Volume sliders
			style = disabledStyle
		}

		label := fmt.Sprintf("%s%d. %s", prefix, i+1, style.Render(item.label))
		if item.value != "" {
			if i < 3 { // Volume items
				label += "  " + valueStyle.Render(item.value)
			} else { // Toggle items
				label += "  " + item.value
			}
		}
		b.WriteString(label)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Current effective volume
	infoStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	effectiveBGM := m.config.Audio.MasterVolume * m.config.Audio.BGMVolume / 100
	effectiveSFX := m.config.Audio.MasterVolume * m.config.Audio.SFXVolume / 100
	b.WriteString(infoStyle.Render(fmt.Sprintf("實際 BGM 音量: %d%%  |  實際音效音量: %d%%", effectiveBGM, effectiveSFX)))
	b.WriteString("\n\n")

	// Hints
	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	hints := "↑/↓: 選擇  |  ←/→: 調整  |  Enter: 確認/切換  |  ESC: 儲存並返回"
	b.WriteString(hintStyle.Render(hints))

	return b.String()
}

func (m AudioSettingsModel) renderVolumeBar(volume int) string {
	const barWidth = 20
	filled := volume * barWidth / 100
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %3d%%", bar, volume)
}

func (m AudioSettingsModel) renderToggle(enabled bool) string {
	tm := themes.GetManager()
	theme := tm.GetCurrentTheme()

	if enabled {
		return lipgloss.NewStyle().
			Foreground(theme.Colors.Success).
			Render("● 開啟")
	}
	return lipgloss.NewStyle().
		Foreground(theme.Colors.Error).
		Render("○ 關閉")
}

// IsDone returns true if settings is complete.
func (m AudioSettingsModel) IsDone() bool {
	return m.done
}

// GetConfig returns the updated config.
func (m AudioSettingsModel) GetConfig() *config.Config {
	return m.config
}
