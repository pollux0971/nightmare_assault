package components

import (
	"github.com/charmbracelet/lipgloss"
)

// ShortcutBar displays quick help shortcuts at the bottom.
type ShortcutBar struct {
	width int
}

// NewShortcutBar creates a new shortcut bar.
func NewShortcutBar() *ShortcutBar {
	return &ShortcutBar{
		width: 80,
	}
}

// SetWidth sets the shortcut bar width.
func (s *ShortcutBar) SetWidth(width int) {
	s.width = width
}

// Height returns the fixed height.
func (s *ShortcutBar) Height() int {
	return 1
}

// View renders the shortcut bar.
func (s *ShortcutBar) View() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(s.width).
		Align(lipgloss.Center)

	text := "ESC: Menu  │  /help: Help  │  /quit: Quit  │  ↑↓: Scroll"
	return style.Render(text)
}
