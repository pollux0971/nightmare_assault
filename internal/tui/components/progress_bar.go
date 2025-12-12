// Package components provides reusable TUI components for Nightmare Assault.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBarMode represents the display mode of the progress bar.
type ProgressBarMode int

const (
	// ModeIndeterminate shows infinite loop animation (no percentage)
	ModeIndeterminate ProgressBarMode = iota
	// ModeDeterminate shows actual progress with percentage
	ModeDeterminate
)

// ProgressBarStyle defines the visual style of the progress bar.
type ProgressBarStyle struct {
	FilledColor   lipgloss.Color
	EmptyColor    lipgloss.Color
	ShowPercent   bool
	Width         int
	CharFilled    string
	CharEmpty     string
}

// DefaultProgressBarStyle returns the default style for progress bars.
func DefaultProgressBarStyle() ProgressBarStyle {
	return ProgressBarStyle{
		FilledColor: lipgloss.Color("#FF0000"),   // Horror red
		EmptyColor:  lipgloss.Color("#3A3A3A"),   // Dark gray
		ShowPercent: true,
		Width:       40,
		CharFilled:  "█",
		CharEmpty:   "░",
	}
}

// HorrorProgressBarStyle returns a horror-themed style.
func HorrorProgressBarStyle() ProgressBarStyle {
	return ProgressBarStyle{
		FilledColor: lipgloss.Color("#8B0000"),   // Dark red
		EmptyColor:  lipgloss.Color("#1C1C1C"),   // Almost black
		ShowPercent: true,
		Width:       45,
		CharFilled:  "█",
		CharEmpty:   "░",
	}
}

// DreamProgressBarStyle returns a dream-themed style.
func DreamProgressBarStyle() ProgressBarStyle {
	return ProgressBarStyle{
		FilledColor: lipgloss.Color("#9370DB"),   // Medium purple
		EmptyColor:  lipgloss.Color("#2C2C3C"),   // Dark blue-gray
		ShowPercent: true,
		Width:       40,
		CharFilled:  "▓",
		CharEmpty:   "░",
	}
}

// ProgressBar wraps the bubbles progress component with custom styling.
type ProgressBar struct {
	progress     progress.Model
	mode         ProgressBarMode
	percent      float64  // 0.0 - 1.0
	style        ProgressBarStyle
	indeterminatePos int  // For animation
}

// NewProgressBar creates a new progress bar with the given style.
func NewProgressBar(style ProgressBarStyle) ProgressBar {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = style.Width

	return ProgressBar{
		progress: prog,
		mode:     ModeDeterminate,
		percent:  0.0,
		style:    style,
	}
}

// NewIndeterminateProgressBar creates an indeterminate progress bar.
func NewIndeterminateProgressBar(style ProgressBarStyle) ProgressBar {
	pb := NewProgressBar(style)
	pb.mode = ModeIndeterminate
	return pb
}

// SetPercent sets the progress percentage (0-100).
func (pb *ProgressBar) SetPercent(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	pb.percent = float64(percent) / 100.0
}

// SetPercentFloat sets the progress percentage (0.0 - 1.0).
func (pb *ProgressBar) SetPercentFloat(percent float64) {
	if percent < 0.0 {
		percent = 0.0
	}
	if percent > 1.0 {
		percent = 1.0
	}
	pb.percent = percent
}

// Increment increments the progress by the given amount (0-100).
func (pb *ProgressBar) Increment(amount int) {
	pb.SetPercent(int(pb.percent*100) + amount)
}

// GetPercent returns the current percentage (0-100).
func (pb *ProgressBar) GetPercent() int {
	return int(pb.percent * 100)
}

// SetMode sets the progress bar mode.
func (pb *ProgressBar) SetMode(mode ProgressBarMode) {
	pb.mode = mode
}

// Reset resets the progress to 0.
func (pb *ProgressBar) Reset() {
	pb.percent = 0.0
	pb.indeterminatePos = 0
}

// Init initializes the progress bar.
func (pb ProgressBar) Init() tea.Cmd {
	return nil
}

// Update handles messages for the progress bar.
func (pb ProgressBar) Update(msg tea.Msg) (ProgressBar, tea.Cmd) {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		pb.progress.Width = pb.style.Width
	}

	// Update indeterminate animation
	if pb.mode == ModeIndeterminate {
		pb.indeterminatePos = (pb.indeterminatePos + 1) % (pb.style.Width * 2)
	}

	return pb, nil
}

// View renders the progress bar.
func (pb ProgressBar) View() string {
	if pb.mode == ModeIndeterminate {
		return pb.viewIndeterminate()
	}
	return pb.viewDeterminate()
}

// ViewWithLabel renders the progress bar with a label.
func (pb ProgressBar) ViewWithLabel(label string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Bold(true)

	return labelStyle.Render(label) + "\n" + pb.View()
}

// viewDeterminate renders a determinate progress bar.
func (pb ProgressBar) viewDeterminate() string {
	filledWidth := int(float64(pb.style.Width) * pb.percent)
	emptyWidth := pb.style.Width - filledWidth

	filled := strings.Repeat(pb.style.CharFilled, filledWidth)
	empty := strings.Repeat(pb.style.CharEmpty, emptyWidth)

	filledStyle := lipgloss.NewStyle().Foreground(pb.style.FilledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(pb.style.EmptyColor)

	bar := filledStyle.Render(filled) + emptyStyle.Render(empty)

	if pb.style.ShowPercent {
		percentText := fmt.Sprintf("  %d%%", pb.GetPercent())
		percentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
		bar += percentStyle.Render(percentText)
	}

	return bar
}

// viewIndeterminate renders an indeterminate progress bar with animation.
func (pb ProgressBar) viewIndeterminate() string {
	// Create a moving block effect
	blockSize := pb.style.Width / 5
	pos := pb.indeterminatePos % pb.style.Width

	var result strings.Builder
	emptyStyle := lipgloss.NewStyle().Foreground(pb.style.EmptyColor)
	filledStyle := lipgloss.NewStyle().Foreground(pb.style.FilledColor)

	for i := 0; i < pb.style.Width; i++ {
		// Create a moving "wave" effect
		if i >= pos && i < pos+blockSize {
			result.WriteString(filledStyle.Render(pb.style.CharFilled))
		} else {
			result.WriteString(emptyStyle.Render(pb.style.CharEmpty))
		}
	}

	return result.String()
}

// ViewCompact renders a compact version without percentage text.
func (pb ProgressBar) ViewCompact() string {
	originalShow := pb.style.ShowPercent
	pb.style.ShowPercent = false
	view := pb.View()
	pb.style.ShowPercent = originalShow
	return view
}

// ViewWithStats renders the progress bar with detailed statistics.
func (pb ProgressBar) ViewWithStats(label string, elapsed string, remaining string) string {
	var b strings.Builder

	// Label
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Bold(true)
	b.WriteString(labelStyle.Render(label))
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(pb.View())
	b.WriteString("\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	stats := fmt.Sprintf("⏱  已經過: %s", elapsed)
	if remaining != "" && pb.mode == ModeDeterminate {
		stats += fmt.Sprintf("  |  預計剩餘: %s", remaining)
	}

	b.WriteString(statsStyle.Render(stats))

	return b.String()
}
