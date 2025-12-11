package effects

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestApplyColorShift_NoShift(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := ApplyColorShift(style, 0)

	if result.GetForeground() != lipgloss.Color("10") {
		t.Errorf("Expected no color shift with shift=0")
	}
}

func TestApplyColorShift_SlightShift(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := ApplyColorShift(style, 7) // 5-10 range

	// Should shift to Cyan (14)
	if result.GetForeground() != lipgloss.Color("14") {
		t.Errorf("Expected slight shift to Cyan, got %v", result.GetForeground())
	}
}

func TestApplyColorShift_MediumShift(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := ApplyColorShift(style, 20) // 15-30 range

	// Should shift to Yellow (11)
	if result.GetForeground() != lipgloss.Color("11") {
		t.Errorf("Expected medium shift to Yellow, got %v", result.GetForeground())
	}
}

func TestApplyColorShift_SevereShift(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := ApplyColorShift(style, 60) // 45-90 range

	// Should shift to Magenta (13)
	if result.GetForeground() != lipgloss.Color("13") {
		t.Errorf("Expected severe shift to Magenta, got %v", result.GetForeground())
	}
}

func TestApplyColorShift_ExtremeShift(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := ApplyColorShift(style, 150) // 120-180 range

	// Should shift to Red (9)
	if result.GetForeground() != lipgloss.Color("9") {
		t.Errorf("Expected extreme shift to Red, got %v", result.GetForeground())
	}
}

func TestApplyColorShift_WithBackground(t *testing.T) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("0"))

	result := ApplyColorShift(style, 150)

	// Background should also shift to dark red (52)
	if result.GetBackground() != lipgloss.Color("52") {
		t.Errorf("Expected background shift to dark red, got %v", result.GetBackground())
	}
}

func TestShiftThemeColors_AllRanges(t *testing.T) {
	tests := []struct {
		shift    int
		expected lipgloss.TerminalColor
		name     string
	}{
		{0, lipgloss.Color("10"), "no shift"},
		{7, lipgloss.Color("14"), "slight shift"},
		{20, lipgloss.Color("11"), "medium shift"},
		{60, lipgloss.Color("13"), "severe shift"},
		{150, lipgloss.Color("9"), "extreme shift"},
	}

	baseColor := lipgloss.Color("10")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShiftThemeColors(baseColor, tt.shift)
			if result != tt.expected {
				t.Errorf("ShiftThemeColors(%d) = %v, want %v", tt.shift, result, tt.expected)
			}
		})
	}
}

func TestShiftThemeColors_BoundaryValues(t *testing.T) {
	baseColor := lipgloss.Color("10")

	// Test exact boundaries
	tests := []struct {
		shift    int
		expected lipgloss.TerminalColor
	}{
		{4, baseColor},       // Just below slight (5)
		{5, lipgloss.Color("14")},  // Exact start of slight
		{14, lipgloss.Color("14")}, // Just below medium (15)
		{15, lipgloss.Color("11")}, // Exact start of medium
		{44, lipgloss.Color("11")}, // Just below severe (45)
		{45, lipgloss.Color("13")}, // Exact start of severe
		{119, lipgloss.Color("13")}, // Just below extreme (120)
		{120, lipgloss.Color("9")}, // Exact start of extreme
	}

	for _, tt := range tests {
		result := ShiftThemeColors(baseColor, tt.shift)
		if result != tt.expected {
			t.Errorf("ShiftThemeColors(%d) = %v, want %v", tt.shift, result, tt.expected)
		}
	}
}
