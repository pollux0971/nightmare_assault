package components

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// StatusBar displays game status information.
type StatusBar struct {
	stats     *game.PlayerStats
	turnCount int
	gameMode  string
	width     int
	mu        sync.RWMutex
}

// NewStatusBar creates a new status bar.
func NewStatusBar(stats *game.PlayerStats, turnCount int) *StatusBar {
	return &StatusBar{
		stats:     stats,
		turnCount: turnCount,
		gameMode:  "Playing",
		width:     80,
	}
}

// SetTurnCount sets the current turn count.
func (s *StatusBar) SetTurnCount(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.turnCount = count
}

// SetGameMode sets the current game mode text.
func (s *StatusBar) SetGameMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gameMode = mode
}

// UpdateStats updates the player stats reference.
func (s *StatusBar) UpdateStats(stats *game.PlayerStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.width = width
}

// Height returns the fixed height of the status bar.
func (s *StatusBar) Height() int {
	return 3
}

// View renders the status bar.
func (s *StatusBar) View() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Styles
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("255")).
		Width(s.width).
		Padding(0, 1)

	hpColor := s.getHPColor(s.stats.HP)
	sanColor := s.getSanityStateColor(s.stats.State)

	// Line 1: HP and SAN bars
	hpBar := s.getHPBar(20)
	sanBar := s.getSANBar(20)

	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hpColor))
	sanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(sanColor))

	line1 := fmt.Sprintf("HP:  %s %3d/100  │  SAN: %s %3d/100",
		hpStyle.Render(hpBar),
		s.stats.HP,
		sanStyle.Render(sanBar),
		s.stats.SAN,
	)

	// Line 2: State and turn info
	stateText := s.stats.State.String()
	line2 := fmt.Sprintf("狀態: %s  │  回合: %d  │  模式: %s",
		sanStyle.Render(stateText),
		s.turnCount,
		s.gameMode,
	)

	// Line 3: Empty spacer
	line3 := ""

	// Render all lines with bar style
	content := line1 + "\n" + line2 + "\n" + line3
	return barStyle.Render(content)
}

// getHPColor returns the color code for HP based on value.
func (s *StatusBar) getHPColor(hp int) string {
	if hp >= 70 {
		return "10" // green
	} else if hp >= 40 {
		return "11" // yellow
	} else if hp >= 20 {
		return "9" // orange
	}
	return "1" // red
}

// getSanityStateColor returns the color code for sanity state.
func (s *StatusBar) getSanityStateColor(state game.SanityState) string {
	switch state {
	case game.SanityClearHeaded:
		return "10" // green
	case game.SanityAnxious:
		return "11" // yellow
	case game.SanityPanicked:
		return "9" // orange
	case game.SanityInsanity:
		return "1" // red
	default:
		return "7" // white
	}
}

// getHPBar creates a visual HP bar.
func (s *StatusBar) getHPBar(width int) string {
	filled := int(float64(s.stats.HP) / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)
	return bar + empty
}

// getSANBar creates a visual SAN bar.
func (s *StatusBar) getSANBar(width int) string {
	filled := int(float64(s.stats.SAN) / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)
	return bar + empty
}
