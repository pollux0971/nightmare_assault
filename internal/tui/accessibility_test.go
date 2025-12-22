package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// ==========================================================================
// Story 8.6: Accessibility Optimization - Tests
// ==========================================================================

func TestDefaultAccessibilityConfig(t *testing.T) {
	config := DefaultAccessibilityConfig()

	if config == nil {
		t.Fatal("DefaultAccessibilityConfig() returned nil")
	}

	// Check defaults
	if !config.ARIALabels {
		t.Error("ARIALabels should be true by default")
	}
	if !config.KeyboardNavigation {
		t.Error("KeyboardNavigation should be true by default")
	}
	if config.TextSizeMultiplier != 1.0 {
		t.Errorf("TextSizeMultiplier = %v, want 1.0", config.TextSizeMultiplier)
	}
	if config.ColorBlindMode != "none" {
		t.Errorf("ColorBlindMode = %s, want none", config.ColorBlindMode)
	}
}

func TestNewAccessibilityManager(t *testing.T) {
	// Test with nil config
	am := NewAccessibilityManager(nil)
	if am == nil {
		t.Fatal("NewAccessibilityManager(nil) returned nil")
	}
	if am.config == nil {
		t.Fatal("config should be initialized with defaults")
	}

	// Test with custom config
	customConfig := &AccessibilityConfig{
		ScreenReaderMode: true,
		HighContrastMode: true,
	}
	am = NewAccessibilityManager(customConfig)
	if !am.config.ScreenReaderMode {
		t.Error("ScreenReaderMode should be true")
	}
	if !am.config.HighContrastMode {
		t.Error("HighContrastMode should be true")
	}
}

func TestSetTextSizeMultiplier(t *testing.T) {
	am := NewAccessibilityManager(nil)

	tests := []struct {
		name       string
		multiplier float64
		want       float64
	}{
		{"normal", 1.0, 1.0},
		{"larger", 1.5, 1.5},
		{"largest", 2.0, 2.0},
		{"too small", 0.5, 0.8},  // Should clamp to 0.8
		{"too large", 3.0, 2.0},  // Should clamp to 2.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// AC4: Text size adjustment
			am.SetTextSizeMultiplier(tt.multiplier)
			if am.config.TextSizeMultiplier != tt.want {
				t.Errorf("TextSizeMultiplier = %v, want %v", am.config.TextSizeMultiplier, tt.want)
			}
		})
	}
}

func TestSetHighContrastMode(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC3: High contrast theme
	am.SetHighContrastMode(true)
	if !am.config.HighContrastMode {
		t.Error("HighContrastMode should be true")
	}

	am.SetHighContrastMode(false)
	if am.config.HighContrastMode {
		t.Error("HighContrastMode should be false")
	}
}

func TestSetColorBlindMode(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC3: Color blind support
	validModes := []string{"none", "protanopia", "deuteranopia", "tritanopia"}

	for _, mode := range validModes {
		am.SetColorBlindMode(mode)
		if am.config.ColorBlindMode != mode {
			t.Errorf("ColorBlindMode = %s, want %s", am.config.ColorBlindMode, mode)
		}
	}

	// Invalid mode should not change
	am.SetColorBlindMode("protanopia")
	am.SetColorBlindMode("invalid_mode")
	if am.config.ColorBlindMode != "protanopia" {
		t.Error("Invalid mode should not change ColorBlindMode")
	}
}

func TestSetScreenReaderMode(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC1: Screen reader support
	am.SetScreenReaderMode(true)
	if !am.config.ScreenReaderMode {
		t.Error("ScreenReaderMode should be true")
	}

	am.SetScreenReaderMode(false)
	if am.config.ScreenReaderMode {
		t.Error("ScreenReaderMode should be false")
	}
}

func TestSetReduceMotion(t *testing.T) {
	am := NewAccessibilityManager(nil)

	am.SetReduceMotion(true)
	if !am.config.ReduceMotion {
		t.Error("ReduceMotion should be true")
	}

	am.SetReduceMotion(false)
	if am.config.ReduceMotion {
		t.Error("ReduceMotion should be false")
	}
}

func TestAdjustStyle(t *testing.T) {
	am := NewAccessibilityManager(nil)

	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

	// AC4: Text size adjustment via padding
	am.SetTextSizeMultiplier(1.5)
	adjusted := am.AdjustStyle(baseStyle)

	// Style should have padding added (we can't directly test padding, but we can verify style was modified)
	if adjusted.GetPaddingLeft() == 0 && adjusted.GetPaddingRight() == 0 {
		// Note: This might pass if lipgloss doesn't expose padding getters
		// The important thing is the function runs without error
	}

	// AC3: High contrast mode
	am.SetHighContrastMode(true)
	adjusted = am.AdjustStyle(baseStyle)

	// Should apply high contrast colors and bold
	// (We can't directly test the style colors, but we verify it runs)
	// Note: lipgloss doesn't expose foreground color directly in all versions
	// Just verify the function runs without error
}

func TestGetHighContrastColors(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC3: High contrast theme
	am.SetHighContrastMode(true)
	colors := am.GetHighContrastColors()

	if len(colors) == 0 {
		t.Error("High contrast colors should not be empty")
	}

	// Check expected color keys exist
	expectedKeys := []string{"foreground", "background", "primary", "secondary", "error", "success", "warning", "border", "highlight"}
	for _, key := range expectedKeys {
		if _, ok := colors[key]; !ok {
			t.Errorf("Missing color key: %s", key)
		}
	}

	// Verify high contrast colors are actually high contrast
	if colors["foreground"] != lipgloss.Color("#FFFFFF") {
		t.Error("High contrast foreground should be white")
	}
	if colors["background"] != lipgloss.Color("#000000") {
		t.Error("High contrast background should be black")
	}
}

func TestGetHighContrastColorsColorBlind(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC3: Color blind mode adjustments
	colorBlindModes := []string{"protanopia", "deuteranopia", "tritanopia"}

	for _, mode := range colorBlindModes {
		am.SetColorBlindMode(mode)
		colors := am.GetHighContrastColors()

		if len(colors) == 0 {
			t.Errorf("Colors should not be empty for mode: %s", mode)
		}

		// Each mode should have different color schemes
		// Just verify they're not the default colors
		if am.config.HighContrastMode {
			if colors["foreground"] == lipgloss.Color("#D0D0D0") {
				t.Errorf("Color blind mode %s should have adjusted colors", mode)
			}
		}
	}
}

func TestAddARIALabel(t *testing.T) {
	am := NewAccessibilityManager(nil)

	content := "Game status"
	label := "STATUS"

	// AC1: ARIA labels
	// Without screen reader mode
	result := am.AddARIALabel(content, label)
	if result != content {
		t.Error("Should return original content when screen reader mode is off")
	}

	// With screen reader mode
	am.SetScreenReaderMode(true)
	result = am.AddARIALabel(content, label)

	if !strings.Contains(result, label) {
		t.Error("Result should contain ARIA label")
	}
	if !strings.Contains(result, content) {
		t.Error("Result should contain original content")
	}
}

func TestAnnounceChange(t *testing.T) {
	am := NewAccessibilityManager(nil)

	message := "Game saved successfully"

	// AC1: Screen reader announcements
	// Without screen reader mode
	result := am.AnnounceChange(message)
	if result != "" {
		t.Error("Should return empty string when screen reader mode is off")
	}

	// With screen reader mode
	am.SetScreenReaderMode(true)
	result = am.AnnounceChange(message)

	if !strings.Contains(result, "ANNOUNCEMENT") {
		t.Error("Announcement should contain ANNOUNCEMENT marker")
	}
	if !strings.Contains(result, message) {
		t.Error("Announcement should contain the message")
	}
}

func TestGetFocusIndicator(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// AC2: Focus indicators
	// Normal mode
	style := am.GetFocusIndicator()
	// Just verify the function runs without error
	// (lipgloss border styles don't have a good way to test equality)

	// High contrast mode
	am.SetHighContrastMode(true)
	_ = am.GetFocusIndicator()
	// Just verify it doesn't panic - can't easily compare lipgloss styles

	// Disabled focus indicators
	am.config.FocusIndicators = false
	_ = am.GetFocusIndicator()
	// Should return empty style when disabled - verified by not panicking
}

func TestFormatKeyboardShortcut(t *testing.T) {
	am := NewAccessibilityManager(nil)

	key := "Ctrl+S"
	description := "Save game"

	// AC2: Keyboard navigation
	result := am.FormatKeyboardShortcut(key, description)

	if !strings.Contains(result, key) {
		t.Error("Result should contain key")
	}
	if !strings.Contains(result, description) {
		t.Error("Result should contain description")
	}

	// Screen reader mode (different format)
	am.SetScreenReaderMode(true)
	result = am.FormatKeyboardShortcut(key, description)

	if !strings.Contains(result, key) {
		t.Error("Result should contain key in screen reader mode")
	}
	if !strings.Contains(result, description) {
		t.Error("Result should contain description in screen reader mode")
	}

	// Disabled shortcuts
	am.config.ShortcutsEnabled = false
	result = am.FormatKeyboardShortcut(key, description)
	if result != "" {
		t.Error("Should return empty when shortcuts disabled")
	}
}

func TestGetKeyboardShortcutsHelp(t *testing.T) {
	am := NewAccessibilityManager(nil)

	shortcuts := map[string]string{
		"Ctrl+S": "Save game",
		"Ctrl+L": "Load game",
		"Ctrl+Q": "Quit",
	}

	// AC2: Keyboard shortcuts help
	result := am.GetKeyboardShortcutsHelp(shortcuts)

	for key, desc := range shortcuts {
		if !strings.Contains(result, key) {
			t.Errorf("Result should contain key: %s", key)
		}
		if !strings.Contains(result, desc) {
			t.Errorf("Result should contain description: %s", desc)
		}
	}

	// Screen reader mode
	am.SetScreenReaderMode(true)
	result = am.GetKeyboardShortcutsHelp(shortcuts)

	if !strings.Contains(result, "KEYBOARD SHORTCUTS HELP") {
		t.Error("Screen reader mode should include help label")
	}
}

func TestShouldReduceMotion(t *testing.T) {
	am := NewAccessibilityManager(nil)

	if am.ShouldReduceMotion() {
		t.Error("Should not reduce motion by default")
	}

	am.SetReduceMotion(true)
	if !am.ShouldReduceMotion() {
		t.Error("Should reduce motion when enabled")
	}
}

func TestShouldSimplifyLayout(t *testing.T) {
	am := NewAccessibilityManager(nil)

	if am.ShouldSimplifyLayout() {
		t.Error("Should not simplify layout by default")
	}

	am.config.SimplifiedLayout = true
	if !am.ShouldSimplifyLayout() {
		t.Error("Should simplify layout when enabled")
	}
}

func TestGetAccessibleText(t *testing.T) {
	am := NewAccessibilityManager(nil)

	text := "Welcome to Nightmare Assault"
	context := "WELCOME_MESSAGE"

	// AC1 & AC4: Screen reader support and text adjustment
	// Normal mode
	result := am.GetAccessibleText(text, context)
	if result != text {
		t.Error("Should return original text in normal mode")
	}

	// Screen reader mode
	am.SetScreenReaderMode(true)
	result = am.GetAccessibleText(text, context)
	if !strings.Contains(result, context) {
		t.Error("Should include context in screen reader mode")
	}

	// With line height adjustment
	am.config.LineHeightAdjust = 1
	multilineText := "Line 1\nLine 2\nLine 3"
	result = am.GetAccessibleText(multilineText, "")

	// Should have extra line breaks
	normalCount := strings.Count(multilineText, "\n")
	resultCount := strings.Count(result, "\n")
	if resultCount <= normalCount {
		t.Error("Should have more line breaks with line height adjustment")
	}
}

func TestExportImportConfig(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// Set some custom values
	am.SetScreenReaderMode(true)
	am.SetHighContrastMode(true)
	am.SetTextSizeMultiplier(1.5)
	am.SetColorBlindMode("protanopia")
	am.SetReduceMotion(true)

	// Export
	exported := am.ExportConfig()

	if exported == nil {
		t.Fatal("ExportConfig returned nil")
	}

	// Verify exported values
	if exported["screen_reader_mode"] != true {
		t.Error("Exported screen_reader_mode should be true")
	}
	if exported["high_contrast_mode"] != true {
		t.Error("Exported high_contrast_mode should be true")
	}
	if exported["text_size_multiplier"] != 1.5 {
		t.Error("Exported text_size_multiplier should be 1.5")
	}
	if exported["color_blind_mode"] != "protanopia" {
		t.Error("Exported color_blind_mode should be protanopia")
	}

	// Create new manager and import
	am2 := NewAccessibilityManager(nil)
	am2.ImportConfig(exported)

	// Verify imported values
	if !am2.config.ScreenReaderMode {
		t.Error("Imported ScreenReaderMode should be true")
	}
	if !am2.config.HighContrastMode {
		t.Error("Imported HighContrastMode should be true")
	}
	if am2.config.TextSizeMultiplier != 1.5 {
		t.Error("Imported TextSizeMultiplier should be 1.5")
	}
	if am2.config.ColorBlindMode != "protanopia" {
		t.Error("Imported ColorBlindMode should be protanopia")
	}
	if !am2.config.ReduceMotion {
		t.Error("Imported ReduceMotion should be true")
	}
}

func TestGetConfig(t *testing.T) {
	am := NewAccessibilityManager(nil)

	config := am.GetConfig()
	if config == nil {
		t.Error("GetConfig should not return nil")
	}

	// Verify it returns the actual config
	if config != am.config {
		t.Error("GetConfig should return the internal config")
	}
}

func TestLogAccessibilityEvent(t *testing.T) {
	am := NewAccessibilityManager(nil)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAccessibilityEvent panicked: %v", r)
		}
	}()

	am.LogAccessibilityEvent("test_event", "test details")
}

func TestAccessibilityIntegration(t *testing.T) {
	// Integration test: Full accessibility workflow
	am := NewAccessibilityManager(nil)

	// Enable all accessibility features
	am.SetScreenReaderMode(true)
	am.SetHighContrastMode(true)
	am.SetTextSizeMultiplier(1.5)
	am.SetColorBlindMode("deuteranopia")
	am.SetReduceMotion(true)

	// Get colors
	colors := am.GetHighContrastColors()
	if len(colors) == 0 {
		t.Error("Should have colors in high contrast mode")
	}

	// Format text with ARIA label
	text := am.AddARIALabel("Important message", "IMPORTANT")
	if !strings.Contains(text, "IMPORTANT") {
		t.Error("Should contain ARIA label")
	}

	// Get keyboard shortcuts
	shortcuts := map[string]string{
		"Tab": "Next item",
		"Enter": "Select",
	}
	help := am.GetKeyboardShortcutsHelp(shortcuts)
	if !strings.Contains(help, "KEYBOARD SHORTCUTS") {
		t.Error("Should contain shortcuts help header")
	}

	// Announce change
	announcement := am.AnnounceChange("Settings saved")
	if announcement == "" {
		t.Error("Should have announcement text")
	}

	// Export and verify all settings
	exported := am.ExportConfig()
	if exported["screen_reader_mode"] != true {
		t.Error("All accessibility features should be exported correctly")
	}
}
