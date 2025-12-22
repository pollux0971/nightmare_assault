// Package components provides reusable TUI components
// Story 7-6: MomentumIndicator UI component
package components

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// MomentumState represents the current state of narrative momentum
type MomentumState int

const (
	// MomentumIdle indicates waiting for player input
	MomentumIdle MomentumState = iota
	// MomentumAutoResolving indicates automatic narrative resolution in progress
	MomentumAutoResolving
	// MomentumPaused indicates the system has paused for player decision
	MomentumPaused
)

// String returns the string representation of MomentumState
func (s MomentumState) String() string {
	switch s {
	case MomentumIdle:
		return "等待輸入"
	case MomentumAutoResolving:
		return "自動演繹中"
	case MomentumPaused:
		return "已暫停"
	default:
		return "未知"
	}
}

// MomentumIndicator displays the current narrative momentum status
// Story 7-6 AC:
//   - AC1: Display current momentum state
//   - AC2: Show "[自動演繹中: 3/5 回合]" when auto-resolving
//   - AC3: Show "[等待輸入]" when waiting
//   - AC4: Integrate with status bar
type MomentumIndicator struct {
	state           MomentumState
	currentBeat     int // Current beat in auto-resolve sequence
	maxAutoBeats    int // Maximum auto beats allowed
	mu              sync.RWMutex
	enabled         bool // Whether momentum system is enabled
	pauseReason     string // Reason for pausing (optional)
}

// NewMomentumIndicator creates a new momentum indicator
func NewMomentumIndicator() *MomentumIndicator {
	return &MomentumIndicator{
		state:        MomentumIdle,
		currentBeat:  0,
		maxAutoBeats: 5,
		enabled:      true,
	}
}

// SetState updates the momentum state
func (m *MomentumIndicator) SetState(state MomentumState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
}

// SetProgress updates the auto-resolve progress
func (m *MomentumIndicator) SetProgress(current, max int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBeat = current
	m.maxAutoBeats = max
}

// SetPauseReason sets the reason for pausing
func (m *MomentumIndicator) SetPauseReason(reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pauseReason = reason
}

// SetEnabled enables or disables the momentum indicator
func (m *MomentumIndicator) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = enabled
}

// View renders the momentum indicator
func (m *MomentumIndicator) View() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return ""
	}

	// Style configuration
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("237")).
		Padding(0, 1)

	activeStyle := baseStyle.Copy().
		Foreground(lipgloss.Color("10")). // Green for active
		Bold(true)

	pausedStyle := baseStyle.Copy().
		Foreground(lipgloss.Color("11")) // Yellow for paused

	var content string
	switch m.state {
	case MomentumIdle:
		// AC3: Show "[等待輸入]" when idle
		content = "[等待輸入]"
		return baseStyle.Render(content)

	case MomentumAutoResolving:
		// AC2: Show "[自動演繹中: 3/5 回合]" when auto-resolving
		content = fmt.Sprintf("[自動演繹中: %d/%d 回合]", m.currentBeat, m.maxAutoBeats)
		return activeStyle.Render(content)

	case MomentumPaused:
		// Show paused state with optional reason
		if m.pauseReason != "" {
			content = fmt.Sprintf("[已暫停: %s]", m.pauseReason)
		} else {
			content = "[已暫停]"
		}
		return pausedStyle.Render(content)

	default:
		content = "[未知狀態]"
		return baseStyle.Render(content)
	}
}

// GetHeight returns the height of the indicator (always 1 line)
func (m *MomentumIndicator) GetHeight() int {
	return 1
}

// Reset resets the indicator to idle state
func (m *MomentumIndicator) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = MomentumIdle
	m.currentBeat = 0
	m.pauseReason = ""
}
