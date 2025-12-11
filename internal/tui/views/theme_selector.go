// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// ThemeSelectorModel represents the theme selection view.
type ThemeSelectorModel struct {
	themes        []*themes.Theme
	selectedIndex int
	themeManager  *themes.ThemeManager
	width         int
	height        int
}

// NewThemeSelectorModel creates a new theme selector.
func NewThemeSelectorModel() ThemeSelectorModel {
	tm := themes.GetManager()
	return ThemeSelectorModel{
		themes:        tm.GetAllThemes(),
		selectedIndex: 0,
		themeManager:  tm,
	}
}

// Init initializes the model.
func (m ThemeSelectorModel) Init() tea.Cmd {
	return nil
}

// ThemeSelectedMsg is sent when a theme is selected.
type ThemeSelectedMsg struct {
	ThemeID string
}

// ThemeBackMsg is sent when user wants to go back.
type ThemeBackMsg struct{}

// Update handles messages.
func (m ThemeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg { return ThemeBackMsg{} }

		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.themes) - 1
			}
			return m, nil

		case "down", "j":
			m.selectedIndex++
			if m.selectedIndex >= len(m.themes) {
				m.selectedIndex = 0
			}
			return m, nil

		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.themes) {
				m.selectedIndex = idx
				return m.handleSelect()
			}
			return m, nil

		case "enter", " ":
			return m.handleSelect()
		}
	}

	return m, nil
}

func (m ThemeSelectorModel) handleSelect() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.themes) {
		theme := m.themes[m.selectedIndex]
		// Apply theme immediately
		m.themeManager.SetTheme(theme.ID)
		return m, func() tea.Msg {
			return ThemeSelectedMsg{ThemeID: theme.ID}
		}
	}
	return m, nil
}

// View renders the theme selector.
func (m ThemeSelectorModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D4EDD")).
		Bold(true).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("🎨  選擇主題"))
	b.WriteString("\n\n")

	// Theme list
	for i, theme := range m.themes {
		isSelected := i == m.selectedIndex
		isCurrent := m.themeManager.IsCurrentTheme(theme.ID)

		// Theme item style
		var itemStyle lipgloss.Style
		if isSelected {
			itemStyle = lipgloss.NewStyle().
				Foreground(theme.Colors.Accent).
				Bold(true)
		} else {
			itemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ECF0F1"))
		}

		// Prefix
		prefix := "  "
		if isSelected {
			prefix = "❯ "
		}

		// Current marker
		currentMark := ""
		if isCurrent {
			currentMark = " ✓ 使用中"
		}

		// Theme line
		b.WriteString(fmt.Sprintf("%s%d. %s%s\n",
			prefix,
			i+1,
			itemStyle.Render(theme.Name),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(currentMark),
		))

		// Description
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		b.WriteString(fmt.Sprintf("      %s\n", descStyle.Render(theme.Description)))
	}

	b.WriteString("\n")

	// Preview section
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.themes) {
		b.WriteString(m.renderPreview(m.themes[m.selectedIndex]))
	}

	b.WriteString("\n")

	// Hints
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8C8D"))
	hints := "↑/↓ 或 j/k: 移動  |  Enter: 套用  |  ESC: 返回"
	b.WriteString(hintStyle.Render(hints))

	return b.String()
}

func (m ThemeSelectorModel) renderPreview(theme *themes.Theme) string {
	var b strings.Builder

	// Preview box using the theme's colors
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Border).
		Padding(1, 2).
		Width(40)

	var previewContent strings.Builder
	previewContent.WriteString("預覽\n\n")

	// Primary text
	primaryStyle := lipgloss.NewStyle().Foreground(theme.Colors.Primary)
	previewContent.WriteString(primaryStyle.Render("主要文字"))
	previewContent.WriteString("\n")

	// Secondary text
	secondaryStyle := lipgloss.NewStyle().Foreground(theme.Colors.Secondary)
	previewContent.WriteString(secondaryStyle.Render("次要文字"))
	previewContent.WriteString("\n")

	// Accent text
	accentStyle := lipgloss.NewStyle().Foreground(theme.Colors.Accent).Bold(true)
	previewContent.WriteString(accentStyle.Render("強調文字"))
	previewContent.WriteString("\n\n")

	// Status messages
	successStyle := lipgloss.NewStyle().Foreground(theme.Colors.Success)
	previewContent.WriteString(successStyle.Render("✓ 成功訊息"))
	previewContent.WriteString("\n")

	errorStyle := lipgloss.NewStyle().Foreground(theme.Colors.Error)
	previewContent.WriteString(errorStyle.Render("✗ 錯誤訊息"))
	previewContent.WriteString("\n")

	warningStyle := lipgloss.NewStyle().Foreground(theme.Colors.Warning)
	previewContent.WriteString(warningStyle.Render("⚠ 警告訊息"))

	b.WriteString(borderStyle.Render(previewContent.String()))

	return b.String()
}
