// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// SettingsAction represents a settings action.
type SettingsAction int

const (
	SettingsActionTheme SettingsAction = iota
	SettingsActionAPI
	SettingsActionAudio
	SettingsActionBack
)

// SettingsItem represents a settings menu item.
type SettingsItem struct {
	title       string
	description string
	action      SettingsAction
}

// SettingsMenuModel represents the settings menu state.
type SettingsMenuModel struct {
	items         []SettingsItem
	selectedIndex int
	width         int
	height        int
}

// NewSettingsMenuModel creates a new settings menu model.
func NewSettingsMenuModel() SettingsMenuModel {
	items := []SettingsItem{
		{title: "主題", description: "切換界面主題", action: SettingsActionTheme},
		{title: "API 設定", description: "管理 API 供應商", action: SettingsActionAPI},
		{title: "音效設定", description: "調整音量與音效", action: SettingsActionAudio},
		{title: "返回", description: "返回主選單", action: SettingsActionBack},
	}

	return SettingsMenuModel{
		items:         items,
		selectedIndex: 0,
	}
}

// Init initializes the model.
func (m SettingsMenuModel) Init() tea.Cmd {
	return nil
}

// SettingsSelectMsg is sent when a settings item is selected.
type SettingsSelectMsg struct {
	Action SettingsAction
}

// Update handles messages.
func (m SettingsMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg {
				return SettingsSelectMsg{Action: SettingsActionBack}
			}

		case "q", "ctrl+c":
			return m, tea.Quit

		case "1":
			m.selectedIndex = 0
			return m.handleSelect()
		case "2":
			m.selectedIndex = 1
			return m.handleSelect()
		case "3":
			m.selectedIndex = 2
			return m.handleSelect()
		case "4":
			m.selectedIndex = 3
			return m.handleSelect()

		case "enter", " ":
			return m.handleSelect()

		case "up", "k":
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.items) - 1
			}
			return m, nil

		case "down", "j":
			m.selectedIndex++
			if m.selectedIndex >= len(m.items) {
				m.selectedIndex = 0
			}
			return m, nil
		}
	}

	return m, nil
}

func (m SettingsMenuModel) handleSelect() (tea.Model, tea.Cmd) {
	item := m.items[m.selectedIndex]
	return m, func() tea.Msg {
		return SettingsSelectMsg{Action: item.action}
	}
}

// View renders the settings menu.
func (m SettingsMenuModel) View() string {
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
	b.WriteString(titleStyle.Render("⚙️  設定"))
	b.WriteString("\n\n")

	// Menu items
	selectedStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(colors.Primary)
	descStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)

	for i, item := range m.items {
		prefix := "  "
		style := normalStyle

		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}

		b.WriteString(fmt.Sprintf("%s%d. %s", prefix, i+1, style.Render(item.title)))
		if item.description != "" {
			b.WriteString(fmt.Sprintf("  %s", descStyle.Render(item.description)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Hints
	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)
	hints := "↑/↓ 或 j/k: 移動  |  Enter: 確認  |  ESC: 返回"
	b.WriteString(hintStyle.Render(hints))

	return b.String()
}
