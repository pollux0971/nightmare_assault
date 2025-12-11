// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// Death view animation constants
const (
	// deathTransitionFPS is the frame rate for the death transition animation.
	deathTransitionFPS = 30
	// deathTransitionFrames is the total number of frames in the transition (~2 seconds).
	deathTransitionFrames = 60
	// deathFrameDuration is the time between animation frames.
	deathFrameDuration = time.Second / deathTransitionFPS
	// glitchTickDuration is the interval for insanity glitch effects.
	glitchTickDuration = 100 * time.Millisecond
)

// DeathSelectMsg is sent when a death screen option is selected.
type DeathSelectMsg struct {
	Action string // "debrief" or "menu"
}

// TransitionTickMsg is sent for transition animation frames.
type TransitionTickMsg struct{}

// DeathModel represents the death screen view.
type DeathModel struct {
	deathInfo      *game.DeathInfo
	width          int
	height         int
	narrative      string
	selected       int  // 0 = debrief, 1 = menu
	showTransition bool
	transitionTick int
	glitchOffset   int
}

// NewDeathModel creates a new death view model.
func NewDeathModel(deathInfo *game.DeathInfo) DeathModel {
	return DeathModel{
		deathInfo:      deathInfo,
		narrative:      deathInfo.Narrative,
		selected:       0,
		showTransition: true,
		transitionTick: 0,
	}
}

// Init initializes the death view and starts the transition animation.
func (m DeathModel) Init() tea.Cmd {
	return tea.Tick(deathFrameDuration, func(t time.Time) tea.Msg {
		return TransitionTickMsg{}
	})
}

// Update handles messages for the death view.
func (m DeathModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TransitionTickMsg:
		if m.showTransition {
			m.transitionTick++
			if m.transitionTick >= deathTransitionFrames {
				m.showTransition = false
				return m, nil
			}
			return m, tea.Tick(deathFrameDuration, func(t time.Time) tea.Msg {
				return TransitionTickMsg{}
			})
		}
		// Glitch effect for insanity
		if m.deathInfo != nil && m.deathInfo.IsInsanity() {
			m.glitchOffset = rand.Intn(3) - 1
			return m, tea.Tick(glitchTickDuration, func(t time.Time) tea.Msg {
				return TransitionTickMsg{}
			})
		}

	case tea.KeyMsg:
		if m.showTransition {
			// Block input during transition
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < 1 {
				m.selected++
			}
		case "enter":
			action := "debrief"
			if m.selected == 1 {
				action = "menu"
			}
			return m, func() tea.Msg {
				return DeathSelectMsg{Action: action}
			}
		case "q", "esc":
			return m, func() tea.Msg {
				return DeathSelectMsg{Action: "menu"}
			}
		}
	}

	return m, nil
}

// View renders the death screen.
func (m DeathModel) View() string {
	if m.showTransition {
		return m.renderTransition()
	}

	if m.deathInfo != nil && m.deathInfo.IsInsanity() {
		return m.renderInsanityDeath()
	}

	return m.renderNormalDeath()
}

// renderTransition renders the transition animation.
func (m DeathModel) renderTransition() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Calculate fade progress (0-100)
	progress := float64(m.transitionTick) / float64(deathTransitionFrames)

	// Red color intensity increases with progress (using theme error color as base)
	redIntensity := int(139 * progress) // 139 = 0x8B (dark red)
	bgColor := lipgloss.Color(lipgloss.CompleteColor{
		TrueColor: rgbToHex(redIntensity, 0, 0),
	}.TrueColor)

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(bgColor).
		Foreground(theme.Colors.Primary).
		Align(lipgloss.Center, lipgloss.Center)

	// Pulsing text during transition
	var text string
	phase1End := deathTransitionFrames / 3
	phase2End := (deathTransitionFrames * 2) / 3
	if m.transitionTick < phase1End {
		text = "..."
	} else if m.transitionTick < phase2End {
		text = "你感覺到..."
	} else {
		text = "一切都結束了..."
	}

	return style.Render(text)
}

// renderNormalDeath renders the normal death screen (HP=0 or rule violation).
func (m DeathModel) renderNormalDeath() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Dark red background (death-specific)
	bgColor := theme.Colors.Error

	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(bgColor)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Primary).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(2)

	// Death type
	deathType := "你死了"
	if m.deathInfo != nil {
		switch m.deathInfo.Type {
		case game.DeathTypeHP:
			deathType = "【體力耗盡】"
		case game.DeathTypeRule:
			deathType = "【違反潛規則】"
		}
	}

	// Narrative style
	narrativeStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Primary).
		Align(lipgloss.Center).
		Width(min(m.width-10, 80)).
		MarginTop(2)

	// Options
	optionStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary)

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Bold(true)

	options := []string{"查看覆盤", "返回主選單"}
	var optionLines []string
	for i, opt := range options {
		if i == m.selected {
			optionLines = append(optionLines, selectedStyle.Render("> "+opt))
		} else {
			optionLines = append(optionLines, optionStyle.Render("  "+opt))
		}
	}

	optionsBlock := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(3).
		Render(strings.Join(optionLines, "\n"))

	// Build content
	content := titleStyle.Render(deathType) + "\n\n"
	content += narrativeStyle.Render(m.narrative) + "\n"
	content += optionsBlock

	// Center vertically
	contentHeight := lipgloss.Height(content)
	topPadding := (m.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	paddedContent := lipgloss.NewStyle().
		PaddingTop(topPadding).
		Render(content)

	return containerStyle.Render(paddedContent)
}

// renderInsanityDeath renders the insanity death screen (SAN=0) with glitch effects.
func (m DeathModel) renderInsanityDeath() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Dark background with occasional green flashes
	bgColor := theme.Colors.Background

	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(bgColor)

	// Glitchy title - use Success color (green) for matrix effect
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Success).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(2)

	// Glitch effect on title
	title := "【理智崩潰】"
	if m.glitchOffset != 0 {
		title = glitchText(title)
	}

	// Narrative with occasional corruption
	narrativeStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary).
		Align(lipgloss.Center).
		Width(min(m.width-10, 80)).
		MarginTop(2)

	narrative := m.narrative
	if m.glitchOffset != 0 {
		narrative = partialGlitch(narrative, 0.1) // 10% character corruption
	}

	// Options with glitch
	optionStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary)

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Error).
		Bold(true)

	options := []string{"查看覆盤", "返回主選單"}
	var optionLines []string
	for i, opt := range options {
		if i == m.selected {
			optionLines = append(optionLines, selectedStyle.Render("> "+opt))
		} else {
			optionLines = append(optionLines, optionStyle.Render("  "+opt))
		}
	}

	optionsBlock := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(3).
		Render(strings.Join(optionLines, "\n"))

	// Build content
	content := titleStyle.Render(title) + "\n\n"
	content += narrativeStyle.Render(narrative) + "\n"
	content += optionsBlock

	// Center vertically with potential offset
	contentHeight := lipgloss.Height(content)
	topPadding := (m.height-contentHeight)/2 + m.glitchOffset
	if topPadding < 0 {
		topPadding = 0
	}

	paddedContent := lipgloss.NewStyle().
		PaddingTop(topPadding).
		Render(content)

	return containerStyle.Render(paddedContent)
}

// Helper functions

func rgbToHex(r, g, b int) string {
	return lipgloss.CompleteColor{
		TrueColor: "#" + hexByte(r) + hexByte(g) + hexByte(b),
	}.TrueColor
}

func hexByte(n int) string {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	hex := "0123456789ABCDEF"
	return string(hex[n/16]) + string(hex[n%16])
}

// Note: min() is a builtin in Go 1.21+, no need for custom implementation

// glitchText corrupts a string with random characters.
func glitchText(s string) string {
	runes := []rune(s)
	glitchChars := []rune{'█', '▓', '░', '▒', '▀', '▄', '■'}

	result := make([]rune, len(runes))
	for i, r := range runes {
		if rand.Float64() < 0.3 { // 30% chance of glitch
			result[i] = glitchChars[rand.Intn(len(glitchChars))]
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// partialGlitch corrupts a percentage of characters in a string.
func partialGlitch(s string, percentage float64) string {
	runes := []rune(s)
	glitchChars := []rune{'█', '▓', '░', '?', '!', '#', '@'}

	result := make([]rune, len(runes))
	for i, r := range runes {
		if rand.Float64() < percentage && r != ' ' && r != '\n' {
			result[i] = glitchChars[rand.Intn(len(glitchChars))]
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// SetNarrative updates the death narrative.
func (m *DeathModel) SetNarrative(narrative string) {
	m.narrative = narrative
}

// SetSize sets the view dimensions.
func (m *DeathModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsTransitioning returns true if transition is still playing.
func (m DeathModel) IsTransitioning() bool {
	return m.showTransition
}
