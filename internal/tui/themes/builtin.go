// Package themes provides color theme definitions and management for Nightmare Assault.
package themes

import "github.com/charmbracelet/lipgloss"

// MidnightTheme returns the Midnight theme (default).
// Deep blue background with cool, mysterious tones.
func MidnightTheme() *Theme {
	return &Theme{
		ID:          "midnight",
		Name:        "午夜 (Midnight)",
		Description: "深邃午夜，神秘冷峻",
		Colors: ThemeColors{
			Primary:    lipgloss.Color("#E0E7FF"), // Light blue-white
			Secondary:  lipgloss.Color("#9CA3AF"), // Gray-blue
			Accent:     lipgloss.Color("#60A5FA"), // Bright blue
			Background: lipgloss.Color("#1E293B"), // Deep blue-gray
			Border:     lipgloss.Color("#475569"), // Medium blue-gray
			Error:      lipgloss.Color("#F87171"), // Light red
			Success:    lipgloss.Color("#34D399"), // Light green
			Warning:    lipgloss.Color("#FBBF24"), // Light yellow
		},
	}
}

// BloodMoonTheme returns the Blood Moon theme.
// Deep red tones evoking horror atmosphere.
func BloodMoonTheme() *Theme {
	return &Theme{
		ID:          "blood_moon",
		Name:        "血月 (Blood Moon)",
		Description: "血色迷霧，恐怖氛圍",
		Colors: ThemeColors{
			Primary:    lipgloss.Color("#FCA5A5"), // Light blood red
			Secondary:  lipgloss.Color("#881337"), // Dark red
			Accent:     lipgloss.Color("#DC2626"), // Vivid red
			Background: lipgloss.Color("#450a0a"), // Very dark red
			Border:     lipgloss.Color("#7F1D1D"), // Deep red
			Error:      lipgloss.Color("#FEE2E2"), // Very light red
			Success:    lipgloss.Color("#86EFAC"), // Light green (contrast)
			Warning:    lipgloss.Color("#FDE047"), // Bright yellow
		},
	}
}

// TerminalGreenTheme returns the Terminal Green theme.
// Classic 80s retro terminal aesthetic.
func TerminalGreenTheme() *Theme {
	return &Theme{
		ID:          "terminal_green",
		Name:        "終端綠 (Terminal Green)",
		Description: "經典終端，復古綠光",
		Colors: ThemeColors{
			Primary:    lipgloss.Color("#00FF00"), // Bright green
			Secondary:  lipgloss.Color("#008000"), // Deep green
			Accent:     lipgloss.Color("#00FF00"), // Bright green
			Background: lipgloss.Color("#000000"), // Pure black
			Border:     lipgloss.Color("#00AA00"), // Medium green
			Error:      lipgloss.Color("#FF0000"), // Red
			Success:    lipgloss.Color("#00FF00"), // Green
			Warning:    lipgloss.Color("#FFFF00"), // Yellow
		},
	}
}

// SilentHillFogTheme returns the Silent Hill Fog theme.
// Misty gray tones with eerie atmosphere.
func SilentHillFogTheme() *Theme {
	return &Theme{
		ID:          "silent_hill_fog",
		Name:        "寂靜嶺迷霧 (Silent Hill Fog)",
		Description: "迷霧寂靜，詭譎不安",
		Colors: ThemeColors{
			Primary:    lipgloss.Color("#D1D5DB"), // Light gray
			Secondary:  lipgloss.Color("#6B7280"), // Medium gray
			Accent:     lipgloss.Color("#FCD34D"), // Pale yellow (fog light)
			Background: lipgloss.Color("#374151"), // Gray-blue
			Border:     lipgloss.Color("#9CA3AF"), // Light gray
			Error:      lipgloss.Color("#EF4444"), // Red
			Success:    lipgloss.Color("#10B981"), // Green
			Warning:    lipgloss.Color("#F59E0B"), // Orange
		},
	}
}

// HighContrastTheme returns the High Contrast theme.
// Accessibility-focused with maximum readability.
func HighContrastTheme() *Theme {
	return &Theme{
		ID:          "high_contrast",
		Name:        "高對比 (High Contrast)",
		Description: "高對比，清晰易讀",
		Colors: ThemeColors{
			Primary:    lipgloss.Color("#FFFFFF"), // Pure white
			Secondary:  lipgloss.Color("#CCCCCC"), // Light gray
			Accent:     lipgloss.Color("#FFFF00"), // Pure yellow
			Background: lipgloss.Color("#000000"), // Pure black
			Border:     lipgloss.Color("#FFFFFF"), // Pure white
			Error:      lipgloss.Color("#FF0000"), // Pure red
			Success:    lipgloss.Color("#00FF00"), // Pure green
			Warning:    lipgloss.Color("#FFFF00"), // Pure yellow
		},
	}
}
