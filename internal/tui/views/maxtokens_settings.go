// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// MaxTokensSettingsModel represents the MaxTokens settings state.
type MaxTokensSettingsModel struct {
	textInput textinput.Model
	config    *config.Config
	errorMsg  string
	width     int
	height    int
	done      bool
}

// NewMaxTokensSettingsModel creates a new MaxTokens settings model.
func NewMaxTokensSettingsModel(cfg *config.Config) MaxTokensSettingsModel {
	ti := textinput.New()
	ti.Placeholder = "輸入 MaxTokens 值 (例如: 100000)"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 50

	// Set current value
	currentValue := cfg.API.Provider.MaxTokens
	if currentValue > 0 {
		ti.SetValue(fmt.Sprintf("%d", currentValue))
	} else {
		ti.SetValue("100000")
	}

	return MaxTokensSettingsModel{
		textInput: ti,
		config:    cfg,
	}
}

// Init initializes the model.
func (m MaxTokensSettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

// MaxTokensSavedMsg is sent when MaxTokens is saved.
type MaxTokensSavedMsg struct{}

// Update handles messages.
func (m MaxTokensSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Cancel and go back
			m.done = true
			return m, nil

		case "enter":
			return m.handleSave()
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m MaxTokensSettingsModel) handleSave() (tea.Model, tea.Cmd) {
	valueStr := strings.TrimSpace(m.textInput.Value())
	if valueStr == "" {
		m.errorMsg = "請輸入 MaxTokens 值"
		return m, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		m.errorMsg = "請輸入有效的數字"
		return m, nil
	}

	if value < 1000 {
		m.errorMsg = "MaxTokens 不能小於 1000"
		return m, nil
	}

	if value > 1000000 {
		m.errorMsg = "MaxTokens 不能大於 1,000,000"
		return m, nil
	}

	// Save to config
	m.config.API.Provider.MaxTokens = value
	if err := m.config.Save(); err != nil {
		m.errorMsg = fmt.Sprintf("儲存失敗: %v", err)
		return m, nil
	}

	m.done = true
	return m, func() tea.Msg {
		return MaxTokensSavedMsg{}
	}
}

// View renders the MaxTokens settings.
func (m MaxTokensSettingsModel) View() string {
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
	b.WriteString(titleStyle.Render("⚙️  MaxTokens 設定"))
	b.WriteString("\n\n")

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	b.WriteString(descStyle.Render("MaxTokens 控制 AI 生成的最大長度。"))
	b.WriteString("\n")
	b.WriteString(descStyle.Render("建議值：100000（可根據需求調整）"))
	b.WriteString("\n\n")

	// Current value
	infoStyle := lipgloss.NewStyle().
		Foreground(colors.Primary)
	currentValue := m.config.API.Provider.MaxTokens
	b.WriteString(infoStyle.Render(fmt.Sprintf("當前值：%d", currentValue)))
	b.WriteString("\n\n")

	// Input
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true)
		b.WriteString(errorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
	}

	// Hints
	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	hints := "Enter: 儲存  |  ESC: 返回"
	b.WriteString(hintStyle.Render(hints))

	return b.String()
}

// IsDone returns true if settings is complete.
func (m MaxTokensSettingsModel) IsDone() bool {
	return m.done
}

// GetConfig returns the updated config.
func (m MaxTokensSettingsModel) GetConfig() *config.Config {
	return m.config
}
