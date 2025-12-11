package effects

import (
	"github.com/charmbracelet/lipgloss"
)

// ApplyColorShift applies a hue shift to a LipGloss style.
// shift is the degree of hue rotation (0-360).
//
// This creates disturbing color changes as SAN decreases:
//   - Small shifts (5-10°): Subtle unease
//   - Medium shifts (15-30°): Noticeable wrongness
//   - Large shifts (45-90°): Severe distortion
//   - Extreme shifts (120-180°): Complete inversion
func ApplyColorShift(style lipgloss.Style, shift int) lipgloss.Style {
	if shift == 0 {
		return style
	}

	// For now, we'll use predefined color shifts
	// In a full implementation, this would do HSL transformations
	// Since LipGloss doesn't have built-in HSL support, we use color replacement

	// Apply shift by changing to a different color based on shift amount
	if shift >= 120 {
		// Extreme: invert to red/magenta
		style = style.Foreground(lipgloss.Color("9"))  // Red
	} else if shift >= 45 {
		// Severe: shift to purple/cyan
		style = style.Foreground(lipgloss.Color("13")) // Magenta
	} else if shift >= 15 {
		// Medium: shift to yellow/orange
		style = style.Foreground(lipgloss.Color("11")) // Yellow
	} else if shift >= 5 {
		// Slight: subtle shift
		style = style.Foreground(lipgloss.Color("14")) // Cyan
	}

	// If original had a background, shift it too
	bgColor := style.GetBackground()
	if bgColor != lipgloss.TerminalColor(nil) && bgColor != lipgloss.Color("") {
		style = style.Background(lipgloss.Color("52")) // Dark red
	}

	return style
}

// ShiftThemeColors applies color shift to an entire theme based on HorrorStyle.
// This is used to distort the entire UI color scheme as SAN drops.
func ShiftThemeColors(baseColor lipgloss.TerminalColor, shift int) lipgloss.TerminalColor {
	if shift == 0 {
		return baseColor
	}

	// Map shift ranges to disturbed colors
	switch {
	case shift >= 120:
		return lipgloss.Color("9")  // Extreme: Red
	case shift >= 45:
		return lipgloss.Color("13") // Severe: Magenta
	case shift >= 15:
		return lipgloss.Color("11") // Medium: Yellow
	case shift >= 5:
		return lipgloss.Color("14") // Slight: Cyan
	default:
		return baseColor
	}
}
