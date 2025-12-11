// Package styles provides LipGloss styles for the Nightmare Assault TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette - Cosmic Horror theme
var (
	ColorPrimary   = lipgloss.Color("#9B59B6") // Purple - mysterious
	ColorSecondary = lipgloss.Color("#2ECC71") // Green - eldritch
	ColorWarning   = lipgloss.Color("#E74C3C") // Red - danger
	ColorText      = lipgloss.Color("#ECF0F1") // Light gray - readable
	ColorMuted     = lipgloss.Color("#7F8C8D") // Muted gray - hints
	ColorBorder    = lipgloss.Color("#34495E") // Dark blue-gray - borders
)

// Base styles
var (
	// Title style - bold and prominent
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		MarginBottom(1)

	// Subtitle style - version info etc.
	Subtitle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// Text style - normal text
	Text = lipgloss.NewStyle().
		Foreground(ColorText)

	// Hint style - help text
	Hint = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// Warning style - error/warning messages
	Warning = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning)

	// Success style - success messages
	Success = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	// Container style - main content container
	Container = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)
)
