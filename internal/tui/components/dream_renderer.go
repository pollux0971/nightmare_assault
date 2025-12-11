package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DreamModel handles dream rendering with special visual style
type DreamModel struct {
	viewport      viewport.Model
	content       string
	dreamType     string
	isTransition  bool
	transitionMsg string
	width         int
	height        int
	style         lipgloss.Style
}

// DreamStyle returns the visual style for dreams
var DreamStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7f8c9d")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#4a5568")).
	BorderStyle(lipgloss.Border{
		Top:         "- ",
		Bottom:      "- ",
		Left:        "│ ",
		Right:       " │",
		TopLeft:     "┌ ",
		TopRight:    " ┐",
		BottomLeft:  "└ ",
		BottomRight: " ┘",
	}).
	Padding(1, 2)

// TransitionStyle for dream transitions
var TransitionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#9ca3af")).
	Italic(true).
	Align(lipgloss.Center)

// NewDreamModel creates a new dream renderer
func NewDreamModel(content, dreamType string, width, height int) DreamModel {
	vp := viewport.New(width-6, height-8)
	vp.SetContent(content)

	return DreamModel{
		viewport:      vp,
		content:       content,
		dreamType:     dreamType,
		isTransition:  false,
		transitionMsg: "",
		width:         width,
		height:        height,
		style:         DreamStyle,
	}
}

// Init initializes the dream model
func (m DreamModel) Init() tea.Cmd {
	return nil
}

// Update handles dream model updates
func (m DreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter", " ":
			// Exit dream on enter/space
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 6
		m.viewport.Height = msg.Height - 8
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the dream
func (m DreamModel) View() string {
	var b strings.Builder

	// Header
	header := "【夢境】"
	if m.dreamType == "opening" {
		header = "【開場夢境】"
	} else if m.dreamType == "chapter" {
		header = "【章節夢境】"
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7f8c9d")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width)

	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Dream content with style
	contentView := m.style.Width(m.width - 4).Render(m.viewport.View())
	b.WriteString(contentView)

	b.WriteString("\n\n")

	// Footer hint
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.width)

	footer := "按 Enter 繼續..."
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

// ShowTransition renders a transition message
func ShowTransition(msg string, entering bool) string {
	style := TransitionStyle.Width(80)

	if entering {
		return style.Render("意識逐漸模糊...")
	}
	return style.Render("你驚醒了")
}

// FormatDreamContent formats dream content for display
func FormatDreamContent(content string, accessible bool) string {
	if accessible {
		return "【夢境開始】\n\n" + content + "\n\n【夢境結束】"
	}
	return content
}
