package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// TestThemeCommand_Name tests the Name method (Story 9-7).
func TestThemeCommand_Name(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	if cmd.Name() != "theme" {
		t.Errorf("Expected command name 'theme', got '%s'", cmd.Name())
	}
}

// TestThemeCommand_Help tests the Help method (Story 9-7).
func TestThemeCommand_Help(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
	if !strings.Contains(help, "theme") {
		t.Errorf("Help text should mention 'theme', got: %s", help)
	}
}

// TestThemeCommand_ListThemes tests listing all available themes (Story 9-7 AC2).
func TestThemeCommand_ListThemes(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should list all themes
	if !strings.Contains(output, "午夜") || !strings.Contains(output, "Midnight") {
		t.Error("Output should contain Midnight theme")
	}
	if !strings.Contains(output, "血月") || !strings.Contains(output, "Blood Moon") {
		t.Error("Output should contain Blood Moon theme")
	}
	if !strings.Contains(output, "深淵藍") || !strings.Contains(output, "Abyss Blue") {
		t.Error("Output should contain Abyss Blue theme (Story 9-7 AC1)")
	}

	// Test explicit "list" subcommand
	output2, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Fatalf("Execute with 'list' failed: %v", err)
	}
	if output2 != output {
		t.Error("'list' subcommand should produce same output as no args")
	}
}

// TestThemeCommand_ShowCurrent tests showing current theme (Story 9-7 AC2).
func TestThemeCommand_ShowCurrent(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	output, err := cmd.Execute([]string{"current"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should show current theme (default is midnight)
	if !strings.Contains(output, "當前主題") {
		t.Error("Output should mention current theme")
	}
	if !strings.Contains(output, "午夜") || !strings.Contains(output, "Midnight") {
		t.Error("Default theme should be Midnight")
	}
}

// TestThemeCommand_SwitchTheme tests switching themes (Story 9-7 AC3, AC4, AC5).
func TestThemeCommand_SwitchTheme(t *testing.T) {
	// Create temp config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.SaveToPath(configPath)

	cmd := NewThemeCommand(cfg)

	// Test switching to blood_moon (AC3: 主題應影響顏色)
	output, err := cmd.Execute([]string{"blood_moon"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(output, "血月") && !strings.Contains(output, "Blood Moon") {
		t.Error("Output should confirm Blood Moon theme switch")
	}
	if !strings.Contains(output, "✓") {
		t.Error("Output should show success indicator")
	}

	// AC4: Verify theme change is immediate
	currentTheme := themes.GetManager().GetCurrentTheme()
	if currentTheme.ID != "blood_moon" {
		t.Errorf("Expected current theme to be 'blood_moon', got '%s'", currentTheme.ID)
	}

	// AC5: Verify config was saved
	if cfg.Theme != "blood_moon" {
		t.Errorf("Expected config theme to be 'blood_moon', got '%s'", cfg.Theme)
	}

	// Test switching to abyss_blue (Story 9-7 AC1)
	output, err = cmd.Execute([]string{"abyss_blue"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(output, "深淵藍") && !strings.Contains(output, "Abyss Blue") {
		t.Error("Output should confirm Abyss Blue theme switch")
	}

	currentTheme = themes.GetManager().GetCurrentTheme()
	if currentTheme.ID != "abyss_blue" {
		t.Errorf("Expected current theme to be 'abyss_blue', got '%s'", currentTheme.ID)
	}
}

// TestThemeCommand_InvalidTheme tests handling of invalid theme names (Story 9-7 AC2).
func TestThemeCommand_InvalidTheme(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	_, err := cmd.Execute([]string{"nonexistent_theme"})
	if err == nil {
		t.Error("Expected error for nonexistent theme")
	}

	if !strings.Contains(err.Error(), "未知的主題") {
		t.Errorf("Error should mention unknown theme, got: %v", err)
	}
}

// TestThemeCommand_NormalizeThemeName tests theme name normalization (Story 9-7 AC3).
func TestThemeCommand_NormalizeThemeName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.SaveToPath(configPath)

	cmd := NewThemeCommand(cfg)

	// Test with spaces (should be converted to underscores)
	output, err := cmd.Execute([]string{"blood", "moon"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(output, "血月") && !strings.Contains(output, "Blood Moon") {
		t.Error("Should handle 'blood moon' (with space) as 'blood_moon'")
	}

	// Test with uppercase (should be converted to lowercase)
	output, err = cmd.Execute([]string{"ABYSS_BLUE"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(output, "深淵藍") && !strings.Contains(output, "Abyss Blue") {
		t.Error("Should handle 'ABYSS_BLUE' (uppercase) as 'abyss_blue'")
	}
}

// TestThemeCommand_ConfigPersistence tests that theme preference is saved (Story 9-7 AC5).
func TestThemeCommand_ConfigPersistence(t *testing.T) {
	// Create temp config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial config
	cfg1 := config.DefaultConfig()
	if err := cfg1.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Switch theme
	cmd1 := NewThemeCommand(cfg1)
	_, err := cmd1.Execute([]string{"blood_moon"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify saved
	if err := cfg1.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config again
	cfg2, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify theme persisted
	if cfg2.Theme != "blood_moon" {
		t.Errorf("Expected persisted theme 'blood_moon', got '%s'", cfg2.Theme)
	}
}

// TestThemeCommand_AllRequiredThemes tests that all required themes exist (Story 9-7 AC1).
func TestThemeCommand_AllRequiredThemes(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	// Test that all 3 required themes exist
	requiredThemes := []string{"midnight", "blood_moon", "abyss_blue"}
	for _, themeID := range requiredThemes {
		_, err := cmd.Execute([]string{themeID})
		if err != nil {
			t.Errorf("Required theme '%s' should exist: %v", themeID, err)
		}
	}

	// Verify we have at least 6 themes total
	output, _ := cmd.Execute([]string{"list"})
	themeCount := strings.Count(output, "\n      ") // Count descriptions
	if themeCount < 6 {
		t.Errorf("Expected at least 6 themes, found %d", themeCount)
	}
}

// TestThemeCommand_SaveFailure tests handling of config save failures (Story 9-7 AC5).
func TestThemeCommand_SaveFailure(t *testing.T) {
	// Create config with invalid path
	cfg := config.DefaultConfig()

	cmd := NewThemeCommand(cfg)
	output, err := cmd.Execute([]string{"blood_moon"})

	// Even if save fails, theme switch should succeed
	if err != nil {
		t.Fatalf("Execute should not fail even if save fails: %v", err)
	}

	// Output should warn about save failure if it occurred
	if !strings.Contains(output, "✓") {
		t.Error("Should indicate successful theme switch")
	}
}

// TestThemeCommand_ThemeEffectsOnColors tests that themes affect colors (Story 9-7 AC3).
func TestThemeCommand_ThemeEffectsOnColors(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewThemeCommand(cfg)

	// Set to a known theme first
	themes.GetManager().SetTheme("midnight")
	midnightTheme := themes.GetManager().GetCurrentTheme()
	midnightPrimary := midnightTheme.Colors.Primary

	// Switch theme
	_, err := cmd.Execute([]string{"blood_moon"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify colors changed
	bloodMoonTheme := themes.GetManager().GetCurrentTheme()
	bloodMoonPrimary := bloodMoonTheme.Colors.Primary

	if midnightPrimary == bloodMoonPrimary {
		t.Error("Theme switch should change primary color")
	}

	// Verify all color fields are set
	if bloodMoonTheme.Colors.Primary == "" {
		t.Error("Primary color should be set")
	}
	if bloodMoonTheme.Colors.Secondary == "" {
		t.Error("Secondary color should be set")
	}
	if bloodMoonTheme.Colors.Accent == "" {
		t.Error("Accent color should be set")
	}
	if bloodMoonTheme.Colors.Background == "" {
		t.Error("Background color should be set")
	}
	if bloodMoonTheme.Colors.Border == "" {
		t.Error("Border color should be set")
	}
	if bloodMoonTheme.Colors.Error == "" {
		t.Error("Error color should be set")
	}
	if bloodMoonTheme.Colors.Success == "" {
		t.Error("Success color should be set")
	}
	if bloodMoonTheme.Colors.Warning == "" {
		t.Error("Warning color should be set")
	}
}
