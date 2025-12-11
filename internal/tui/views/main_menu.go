// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// MenuAction represents an action that can be taken from the menu.
type MenuAction int

const (
	ActionNewGame MenuAction = iota
	ActionContinue
	ActionSettings
	ActionExit
)

// MenuItem represents a menu item.
type MenuItem struct {
	title       string
	description string
	enabled     bool
	action      MenuAction
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.description }
func (i MenuItem) FilterValue() string { return i.title }

// MainMenuModel represents the main menu state.
type MainMenuModel struct {
	list          list.Model
	width         int
	height        int
	hasSaveFiles  bool
	exitConfirm   bool
	selectedIndex int
	version       string
	updateAvailable bool
	newVersion    string
}

// getMenuStyles returns styles based on current theme
func getMenuStyles() (titleStyle, selectedStyle, normalStyle, disabledStyle, hintStyle, borderStyle lipgloss.Style) {
	tm := themes.GetManager()
	theme := tm.GetCurrentTheme()
	colors := theme.Colors

	titleStyle = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
		Foreground(colors.Primary)

	disabledStyle = lipgloss.NewStyle().
		Foreground(colors.Secondary)

	hintStyle = lipgloss.NewStyle().
		Foreground(colors.Secondary).
		MarginTop(1)

	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2)

	return
}

// NewMainMenuModel creates a new main menu model.
func NewMainMenuModel(version string, hasSaveFiles bool) MainMenuModel {
	items := []list.Item{
		MenuItem{title: "新遊戲", description: "開始一場新的惡夢冒險", enabled: true, action: ActionNewGame},
		MenuItem{title: "繼續遊戲", description: "載入上次的存檔", enabled: hasSaveFiles, action: ActionContinue},
		MenuItem{title: "設定", description: "調整遊戲設定", enabled: true, action: ActionSettings},
		MenuItem{title: "離開", description: "退出遊戲", enabled: true, action: ActionExit},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#9D4EDD")).
		BorderForeground(lipgloss.Color("#9D4EDD"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#7B2CBF"))
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#ECF0F1"))
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("#7F8C8D"))

	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)

	return MainMenuModel{
		list:            l,
		hasSaveFiles:    hasSaveFiles,
		version:         version,
		updateAvailable: false,
		newVersion:      "",
	}
}

// Init initializes the model.
func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

// MenuSelectMsg is sent when a menu item is selected.
type MenuSelectMsg struct {
	Action MenuAction
}

// ExitConfirmMsg is sent when exit is confirmed.
type ExitConfirmMsg struct {
	Confirmed bool
}

// UpdateAvailableMsg is sent when an update is available.
type UpdateAvailableMsg struct {
	NewVersion string
}

// Update handles messages.
func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateAvailableMsg:
		m.updateAvailable = true
		m.newVersion = msg.NewVersion
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-8, msg.Height-16)
		return m, nil

	case tea.KeyMsg:
		if m.exitConfirm {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			case "n", "N", "esc":
				m.exitConfirm = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "1":
			return m.selectItem(0)
		case "2":
			return m.selectItem(1)
		case "3":
			return m.selectItem(2)
		case "4":
			return m.selectItem(3)

		case "enter", " ":
			return m.handleSelect()

		case "up", "k":
			m.selectedIndex = m.moveSelection(-1)
			m.list.Select(m.selectedIndex)
			return m, nil

		case "down", "j":
			m.selectedIndex = m.moveSelection(1)
			m.list.Select(m.selectedIndex)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainMenuModel) selectItem(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(m.list.Items()) {
		return m, nil
	}

	item := m.list.Items()[index].(MenuItem)
	if !item.enabled {
		return m, nil
	}

	m.list.Select(index)
	m.selectedIndex = index
	return m.handleSelect()
}

func (m MainMenuModel) moveSelection(delta int) int {
	items := m.list.Items()
	newIndex := m.selectedIndex + delta

	// Wrap around
	if newIndex < 0 {
		newIndex = len(items) - 1
	} else if newIndex >= len(items) {
		newIndex = 0
	}

	// Skip disabled items
	item := items[newIndex].(MenuItem)
	if !item.enabled {
		// Try next item in same direction
		return m.skipDisabled(newIndex, delta)
	}

	return newIndex
}

func (m MainMenuModel) skipDisabled(start, delta int) int {
	items := m.list.Items()
	index := start

	for i := 0; i < len(items); i++ {
		index = (index + delta + len(items)) % len(items)
		item := items[index].(MenuItem)
		if item.enabled {
			return index
		}
	}

	return m.selectedIndex // No enabled item found, stay put
}

func (m MainMenuModel) handleSelect() (tea.Model, tea.Cmd) {
	item := m.list.SelectedItem().(MenuItem)
	if !item.enabled {
		return m, nil
	}

	switch item.action {
	case ActionExit:
		m.exitConfirm = true
		return m, nil
	default:
		return m, func() tea.Msg {
			return MenuSelectMsg{Action: item.action}
		}
	}
}

// View renders the menu.
func (m MainMenuModel) View() string {
	var b strings.Builder

	// Get theme-aware styles
	menuTitleStyle, menuSelectedStyle, menuNormalStyle, menuDisabledStyle, menuHintStyle, menuBorderStyle := getMenuStyles()

	// Title - ASCII art style
	title := `
 ███╗   ██╗██╗ ██████╗ ██╗  ██╗████████╗███╗   ███╗ █████╗ ██████╗ ███████╗
 ████╗  ██║██║██╔════╝ ██║  ██║╚══██╔══╝████╗ ████║██╔══██╗██╔══██╗██╔════╝
 ██╔██╗ ██║██║██║  ███╗███████║   ██║   ██╔████╔██║███████║██████╔╝█████╗
 ██║╚██╗██║██║██║   ██║██╔══██║   ██║   ██║╚██╔╝██║██╔══██║██╔══██╗██╔══╝
 ██║ ╚████║██║╚██████╔╝██║  ██║   ██║   ██║ ╚═╝ ██║██║  ██║██║  ██║███████╗
 ╚═╝  ╚═══╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
                     🎃 NIGHTMARE ASSAULT 🎃
`
	b.WriteString(menuTitleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(menuHintStyle.Render(fmt.Sprintf("                            v%s", m.version)))
	b.WriteString("\n\n")

	// Update notification banner
	if m.updateAvailable {
		tm := themes.GetManager()
		theme := tm.GetCurrentTheme()
		updateBannerStyle := lipgloss.NewStyle().
			Background(theme.Colors.Accent).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1).
			Align(lipgloss.Center)

		updateText := fmt.Sprintf("🚀 新版本可用: %s → %s | 使用 --update 更新", m.version, m.newVersion)
		b.WriteString(updateBannerStyle.Render(updateText))
		b.WriteString("\n\n")
	}

	// Exit confirmation dialog
	if m.exitConfirm {
		tm := themes.GetManager()
		theme := tm.GetCurrentTheme()
		confirmStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Colors.Error).
			Padding(1, 3).
			Align(lipgloss.Center)

		confirmContent := "確定要離開嗎？\n\n(y) 是  (n) 否"
		b.WriteString(confirmStyle.Render(confirmContent))
		return b.String()
	}

	// Menu items
	var menuContent strings.Builder
	items := m.list.Items()
	for i, item := range items {
		mi := item.(MenuItem)
		prefix := "  "
		style := menuNormalStyle

		if i == m.selectedIndex {
			prefix = "❯ "
			style = menuSelectedStyle
		}
		if !mi.enabled {
			style = menuDisabledStyle
		}

		menuContent.WriteString(fmt.Sprintf("%s%d. %s", prefix, i+1, style.Render(mi.title)))
		if mi.description != "" {
			menuContent.WriteString(fmt.Sprintf("  %s", menuDisabledStyle.Render(mi.description)))
		}
		menuContent.WriteString("\n")
	}

	b.WriteString(menuBorderStyle.Render(menuContent.String()))
	b.WriteString("\n")

	// Hints
	hints := "↑/↓ 或 j/k: 移動  |  Enter: 確認  |  1-4: 直接選擇  |  q: 離開"
	b.WriteString(menuHintStyle.Render(hints))

	return b.String()
}

// IsExitConfirming returns true if exit confirmation is shown.
func (m MainMenuModel) IsExitConfirming() bool {
	return m.exitConfirm
}
