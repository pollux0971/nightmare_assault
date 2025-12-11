package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// LogView is a Bubble Tea component for displaying game logs.
type LogView struct {
	viewport viewport.Model
	entries  []game.LogEntry
	ready    bool
	width    int
	height   int
}

// Styles for different log types
var (
	narrativeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	playerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))  // Cyan
	optionStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // Yellow
	systemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray
	timestampStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	footerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// NewLogView creates a new log view with the given entries.
func NewLogView(entries []game.LogEntry, width, height int) LogView {
	vp := viewport.New(width, height-4) // -4 for header and footer
	vp.Style = lipgloss.NewStyle()

	lv := LogView{
		viewport: vp,
		entries:  entries,
		ready:    false,
		width:    width,
		height:   height,
	}

	lv.updateContent()
	return lv
}

// Init initializes the log view.
func (m LogView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the log view.
func (m LogView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return a custom message to close the view
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.viewport.Style = lipgloss.NewStyle()
			m.updateContent()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
			m.updateContent()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the log view.
func (m LogView) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := headerStyle.Render(fmt.Sprintf("┌─ 遊戲日誌 (最近 %d 筆) ", len(m.entries))) +
		strings.Repeat("─", m.width-len("遊戲日誌 (最近    筆) ")-len(fmt.Sprintf("%d", len(m.entries)))-3) +
		"┐\n"

	footer := "\n" + footerStyle.Render("│ [↑↓] 捲動 | [ESC/Q] 關閉 "+
		strings.Repeat(" ", m.width-len("[↑↓] 捲動 | [ESC/Q] 關閉 ")-3)+"│\n"+
		"└"+strings.Repeat("─", m.width-2)+"┘")

	return header + m.viewport.View() + footer
}

// updateContent updates the viewport content with formatted log entries.
func (m *LogView) updateContent() {
	var content strings.Builder

	for _, entry := range m.entries {
		line := m.formatLogEntry(entry)
		content.WriteString(line)
		content.WriteString("\n")
	}

	m.viewport.SetContent(content.String())
	// Scroll to bottom to show latest entries
	m.viewport.GotoBottom()
}

// formatLogEntry formats a single log entry with timestamp and color coding.
func (m LogView) formatLogEntry(entry game.LogEntry) string {
	timestamp := timestampStyle.Render(fmt.Sprintf("[%s]", entry.Timestamp.Format("15:04:05")))

	var typeLabel string
	var styledContent string

	switch entry.Type {
	case game.LogNarrative:
		typeLabel = "[敘事]"
		styledContent = narrativeStyle.Render(entry.Content)
	case game.LogPlayerInput:
		typeLabel = "[玩家]"
		styledContent = playerStyle.Render(entry.Content)
	case game.LogOptionChoice:
		typeLabel = "[選項]"
		styledContent = optionStyle.Render(entry.Content)
	case game.LogSystem:
		typeLabel = "[系統]"
		styledContent = systemStyle.Render(entry.Content)
	default:
		typeLabel = "[?]"
		styledContent = entry.Content
	}

	return fmt.Sprintf("%s %s %s", timestamp, typeLabel, styledContent)
}

// SetEntries updates the log entries and refreshes the view.
func (m *LogView) SetEntries(entries []game.LogEntry) {
	m.entries = entries
	m.updateContent()
}
