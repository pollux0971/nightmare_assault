package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 8.6: Accessibility Optimization
// ==========================================================================
// AC1: Screen reader support (ARIA labels)
// AC2: Keyboard navigation improvements
// AC3: High contrast theme
// AC4: Text size adjustment
// ==========================================================================

// AccessibilityConfig holds accessibility configuration options.
type AccessibilityConfig struct {
	// AC1: Screen reader support
	ScreenReaderMode bool
	ARIALabels       bool
	AnnounceChanges  bool

	// AC2: Keyboard navigation
	KeyboardNavigation bool
	FocusIndicators    bool
	ShortcutsEnabled   bool

	// AC3: High contrast mode
	HighContrastMode bool
	ColorBlindMode   string // "none", "protanopia", "deuteranopia", "tritanopia"

	// AC4: Text size adjustment
	TextSizeMultiplier float64 // 0.8 to 2.0
	LineHeightAdjust   int     // Extra line spacing

	// Other options
	ReduceMotion     bool
	SimplifiedLayout bool
}

// DefaultAccessibilityConfig returns the default accessibility configuration.
func DefaultAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		ScreenReaderMode:   false,
		ARIALabels:         true,
		AnnounceChanges:    true,
		KeyboardNavigation: true,
		FocusIndicators:    true,
		ShortcutsEnabled:   true,
		HighContrastMode:   false,
		ColorBlindMode:     "none",
		TextSizeMultiplier: 1.0,
		LineHeightAdjust:   0,
		ReduceMotion:       false,
		SimplifiedLayout:   false,
	}
}

// AccessibilityManager manages accessibility features.
type AccessibilityManager struct {
	config *AccessibilityConfig
}

// NewAccessibilityManager creates a new accessibility manager.
func NewAccessibilityManager(config *AccessibilityConfig) *AccessibilityManager {
	if config == nil {
		config = DefaultAccessibilityConfig()
	}

	logger.Info("AccessibilityManager initialized", map[string]interface{}{
		"screen_reader": config.ScreenReaderMode,
		"high_contrast": config.HighContrastMode,
		"text_size":     config.TextSizeMultiplier,
	})

	return &AccessibilityManager{
		config: config,
	}
}

// GetConfig returns the current accessibility configuration.
func (am *AccessibilityManager) GetConfig() *AccessibilityConfig {
	return am.config
}

// SetTextSizeMultiplier sets the text size multiplier.
// AC4: Text size adjustment
func (am *AccessibilityManager) SetTextSizeMultiplier(multiplier float64) {
	if multiplier < 0.8 {
		multiplier = 0.8
	}
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	am.config.TextSizeMultiplier = multiplier
	logger.Info("Text size multiplier changed", map[string]interface{}{
		"multiplier": multiplier,
	})
}

// SetHighContrastMode enables or disables high contrast mode.
// AC3: High contrast theme
func (am *AccessibilityManager) SetHighContrastMode(enabled bool) {
	am.config.HighContrastMode = enabled
	logger.Info("High contrast mode changed", map[string]interface{}{
		"enabled": enabled,
	})
}

// SetColorBlindMode sets the color blind mode.
// AC3: Color blind support
func (am *AccessibilityManager) SetColorBlindMode(mode string) {
	validModes := map[string]bool{
		"none":        true,
		"protanopia":  true, // Red-blind
		"deuteranopia": true, // Green-blind
		"tritanopia":  true, // Blue-blind
	}

	if !validModes[mode] {
		logger.Warn("Invalid color blind mode", map[string]interface{}{
			"mode": mode,
		})
		return
	}

	am.config.ColorBlindMode = mode
	logger.Info("Color blind mode changed", map[string]interface{}{
		"mode": mode,
	})
}

// SetScreenReaderMode enables or disables screen reader mode.
// AC1: Screen reader support
func (am *AccessibilityManager) SetScreenReaderMode(enabled bool) {
	am.config.ScreenReaderMode = enabled
	logger.Info("Screen reader mode changed", map[string]interface{}{
		"enabled": enabled,
	})
}

// SetReduceMotion enables or disables motion reduction.
func (am *AccessibilityManager) SetReduceMotion(enabled bool) {
	am.config.ReduceMotion = enabled
	logger.Info("Reduce motion changed", map[string]interface{}{
		"enabled": enabled,
	})
}

// AdjustStyle applies accessibility adjustments to a lipgloss style.
// AC3: High contrast theme
// AC4: Text size adjustment
func (am *AccessibilityManager) AdjustStyle(style lipgloss.Style) lipgloss.Style {
	// AC4: Text size adjustment (simulated with padding for terminal)
	if am.config.TextSizeMultiplier > 1.0 {
		// Add extra padding to simulate larger text
		extraPadding := int((am.config.TextSizeMultiplier - 1.0) * 2)
		style = style.PaddingLeft(extraPadding).PaddingRight(extraPadding)
	}

	// AC3: High contrast mode adjustments
	if am.config.HighContrastMode {
		// Use high contrast colors
		style = style.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#000000")).
			Bold(true)
	}

	// Line height adjustment
	if am.config.LineHeightAdjust > 0 {
		// Add extra line spacing (simulated with margin)
		style = style.MarginTop(am.config.LineHeightAdjust).MarginBottom(am.config.LineHeightAdjust)
	}

	return style
}

// GetHighContrastColors returns high contrast color scheme.
// AC3: High contrast theme
func (am *AccessibilityManager) GetHighContrastColors() map[string]lipgloss.Color {
	if am.config.HighContrastMode {
		return map[string]lipgloss.Color{
			"foreground":    lipgloss.Color("#FFFFFF"),
			"background":    lipgloss.Color("#000000"),
			"primary":       lipgloss.Color("#FFFF00"), // Yellow
			"secondary":     lipgloss.Color("#00FFFF"), // Cyan
			"error":         lipgloss.Color("#FF0000"), // Red
			"success":       lipgloss.Color("#00FF00"), // Green
			"warning":       lipgloss.Color("#FFA500"), // Orange
			"border":        lipgloss.Color("#FFFFFF"),
			"highlight":     lipgloss.Color("#FFFF00"),
		}
	}

	// Apply color blind adjustments
	switch am.config.ColorBlindMode {
	case "protanopia":
		return map[string]lipgloss.Color{
			"foreground":    lipgloss.Color("#E0E0E0"),
			"background":    lipgloss.Color("#1A1A1A"),
			"primary":       lipgloss.Color("#0080FF"), // Blue instead of red
			"secondary":     lipgloss.Color("#FFD700"), // Gold
			"error":         lipgloss.Color("#FF8800"), // Orange instead of red
			"success":       lipgloss.Color("#00BFFF"), // Blue instead of green
			"warning":       lipgloss.Color("#FFD700"), // Gold
			"border":        lipgloss.Color("#808080"),
			"highlight":     lipgloss.Color("#FFD700"),
		}
	case "deuteranopia":
		return map[string]lipgloss.Color{
			"foreground":    lipgloss.Color("#E0E0E0"),
			"background":    lipgloss.Color("#1A1A1A"),
			"primary":       lipgloss.Color("#0080FF"), // Blue instead of green
			"secondary":     lipgloss.Color("#FFD700"), // Gold
			"error":         lipgloss.Color("#FF0080"), // Magenta instead of red
			"success":       lipgloss.Color("#00BFFF"), // Blue instead of green
			"warning":       lipgloss.Color("#FFD700"), // Gold
			"border":        lipgloss.Color("#808080"),
			"highlight":     lipgloss.Color("#FFD700"),
		}
	case "tritanopia":
		return map[string]lipgloss.Color{
			"foreground":    lipgloss.Color("#E0E0E0"),
			"background":    lipgloss.Color("#1A1A1A"),
			"primary":       lipgloss.Color("#FF0080"), // Magenta instead of blue
			"secondary":     lipgloss.Color("#00FF80"), // Cyan-green
			"error":         lipgloss.Color("#FF0000"), // Red
			"success":       lipgloss.Color("#00FF80"), // Cyan-green instead of green
			"warning":       lipgloss.Color("#FF8800"), // Orange
			"border":        lipgloss.Color("#808080"),
			"highlight":     lipgloss.Color("#FF0080"),
		}
	}

	// Default colors (no accessibility mode)
	return map[string]lipgloss.Color{
		"foreground":    lipgloss.Color("#D0D0D0"),
		"background":    lipgloss.Color("#1A1A1A"),
		"primary":       lipgloss.Color("#FF6B6B"),
		"secondary":     lipgloss.Color("#4ECDC4"),
		"error":         lipgloss.Color("#FF4444"),
		"success":       lipgloss.Color("#44FF44"),
		"warning":       lipgloss.Color("#FFAA44"),
		"border":        lipgloss.Color("#666666"),
		"highlight":     lipgloss.Color("#FFFF88"),
	}
}

// AddARIALabel adds ARIA label to content for screen readers.
// AC1: Screen reader support (ARIA labels)
func (am *AccessibilityManager) AddARIALabel(content, label string) string {
	if !am.config.ARIALabels || !am.config.ScreenReaderMode {
		return content
	}

	// Prepend ARIA label (invisible marker for screen readers)
	// In terminal context, we use special markers that screen readers can detect
	return fmt.Sprintf("[%s] %s", label, content)
}

// AnnounceChange announces a change for screen readers.
// AC1: Screen reader announcements
func (am *AccessibilityManager) AnnounceChange(message string) string {
	if !am.config.AnnounceChanges || !am.config.ScreenReaderMode {
		return ""
	}

	// Return announcement in a format screen readers can detect
	return fmt.Sprintf("\n[ANNOUNCEMENT] %s\n", message)
}

// GetFocusIndicator returns a focus indicator style.
// AC2: Keyboard navigation - focus indicators
func (am *AccessibilityManager) GetFocusIndicator() lipgloss.Style {
	if !am.config.FocusIndicators {
		return lipgloss.NewStyle()
	}

	if am.config.HighContrastMode {
		return lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#FFFF00")).
			Bold(true)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Bold(true)
}

// FormatKeyboardShortcut formats a keyboard shortcut description.
// AC2: Keyboard navigation improvements
func (am *AccessibilityManager) FormatKeyboardShortcut(key, description string) string {
	if !am.config.ShortcutsEnabled {
		return ""
	}

	if am.config.ScreenReaderMode {
		return fmt.Sprintf("%s: %s", key, description)
	}

	if am.config.HighContrastMode {
		keyStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#000000")).
			Padding(0, 1)

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		return fmt.Sprintf("%s %s", keyStyle.Render(key), descStyle.Render(description))
	}

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ECDC4")).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D0D0D0"))

	return fmt.Sprintf("%s %s", keyStyle.Render(key), descStyle.Render(description))
}

// GetKeyboardShortcutsHelp returns a formatted list of keyboard shortcuts.
// AC2: Keyboard navigation improvements
func (am *AccessibilityManager) GetKeyboardShortcutsHelp(shortcuts map[string]string) string {
	if !am.config.ShortcutsEnabled {
		return ""
	}

	var lines []string

	// Add ARIA label for screen readers
	if am.config.ScreenReaderMode {
		lines = append(lines, "[KEYBOARD SHORTCUTS HELP]")
	}

	// Format each shortcut
	for key, desc := range shortcuts {
		lines = append(lines, am.FormatKeyboardShortcut(key, desc))
	}

	return strings.Join(lines, "\n")
}

// ShouldReduceMotion returns whether motion should be reduced.
func (am *AccessibilityManager) ShouldReduceMotion() bool {
	return am.config.ReduceMotion
}

// ShouldSimplifyLayout returns whether layout should be simplified.
func (am *AccessibilityManager) ShouldSimplifyLayout() bool {
	return am.config.SimplifiedLayout
}

// GetAccessibleText formats text for accessibility.
// AC1: Screen reader support
// AC4: Text size adjustment
func (am *AccessibilityManager) GetAccessibleText(text string, context string) string {
	// Add ARIA label if in screen reader mode
	if am.config.ScreenReaderMode && context != "" {
		text = am.AddARIALabel(text, context)
	}

	// Adjust line breaks for better readability
	if am.config.LineHeightAdjust > 0 {
		lines := strings.Split(text, "\n")
		extraLines := strings.Repeat("\n", am.config.LineHeightAdjust)
		text = strings.Join(lines, extraLines)
	}

	return text
}

// LogAccessibilityEvent logs an accessibility-related event.
func (am *AccessibilityManager) LogAccessibilityEvent(event, details string) {
	logger.Debug("Accessibility event", map[string]interface{}{
		"event":         event,
		"details":       details,
		"screen_reader": am.config.ScreenReaderMode,
		"high_contrast": am.config.HighContrastMode,
	})
}

// ExportConfig exports the accessibility configuration as a map.
func (am *AccessibilityManager) ExportConfig() map[string]interface{} {
	return map[string]interface{}{
		"screen_reader_mode":   am.config.ScreenReaderMode,
		"aria_labels":          am.config.ARIALabels,
		"announce_changes":     am.config.AnnounceChanges,
		"keyboard_navigation":  am.config.KeyboardNavigation,
		"focus_indicators":     am.config.FocusIndicators,
		"shortcuts_enabled":    am.config.ShortcutsEnabled,
		"high_contrast_mode":   am.config.HighContrastMode,
		"color_blind_mode":     am.config.ColorBlindMode,
		"text_size_multiplier": am.config.TextSizeMultiplier,
		"line_height_adjust":   am.config.LineHeightAdjust,
		"reduce_motion":        am.config.ReduceMotion,
		"simplified_layout":    am.config.SimplifiedLayout,
	}
}

// ImportConfig imports accessibility configuration from a map.
func (am *AccessibilityManager) ImportConfig(config map[string]interface{}) {
	if v, ok := config["screen_reader_mode"].(bool); ok {
		am.config.ScreenReaderMode = v
	}
	if v, ok := config["aria_labels"].(bool); ok {
		am.config.ARIALabels = v
	}
	if v, ok := config["announce_changes"].(bool); ok {
		am.config.AnnounceChanges = v
	}
	if v, ok := config["keyboard_navigation"].(bool); ok {
		am.config.KeyboardNavigation = v
	}
	if v, ok := config["focus_indicators"].(bool); ok {
		am.config.FocusIndicators = v
	}
	if v, ok := config["shortcuts_enabled"].(bool); ok {
		am.config.ShortcutsEnabled = v
	}
	if v, ok := config["high_contrast_mode"].(bool); ok {
		am.config.HighContrastMode = v
	}
	if v, ok := config["color_blind_mode"].(string); ok {
		am.SetColorBlindMode(v)
	}
	if v, ok := config["text_size_multiplier"].(float64); ok {
		am.SetTextSizeMultiplier(v)
	}
	if v, ok := config["line_height_adjust"].(int); ok {
		am.config.LineHeightAdjust = v
	}
	if v, ok := config["reduce_motion"].(bool); ok {
		am.config.ReduceMotion = v
	}
	if v, ok := config["simplified_layout"].(bool); ok {
		am.config.SimplifiedLayout = v
	}

	logger.Info("Accessibility configuration imported", map[string]interface{}{
		"screen_reader": am.config.ScreenReaderMode,
		"high_contrast": am.config.HighContrastMode,
	})
}
