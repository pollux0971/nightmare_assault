package themes

import (
	"testing"
)

func TestNewThemeManager(t *testing.T) {
	tm := NewThemeManager()

	if tm == nil {
		t.Fatal("Expected non-nil ThemeManager")
	}

	// Should have 5 themes
	themes := tm.ListThemes()
	if len(themes) != 5 {
		t.Errorf("Expected 5 themes, got %d", len(themes))
	}

	// Default theme should be midnight
	current := tm.GetCurrentTheme()
	if current == nil {
		t.Fatal("Expected default theme")
	}
	if current.ID != "midnight" {
		t.Errorf("Expected default theme 'midnight', got '%s'", current.ID)
	}
}

func TestSetTheme(t *testing.T) {
	tm := NewThemeManager()

	// Switch to blood_moon
	err := tm.SetTheme("blood_moon")
	if err != nil {
		t.Errorf("SetTheme failed: %v", err)
	}

	current := tm.GetCurrentTheme()
	if current.ID != "blood_moon" {
		t.Errorf("Expected 'blood_moon', got '%s'", current.ID)
	}

	// Try invalid theme
	err = tm.SetTheme("invalid_theme")
	if err == nil {
		t.Error("Expected error for invalid theme")
	}
}

func TestGetTheme(t *testing.T) {
	tm := NewThemeManager()

	// Get existing theme
	theme, ok := tm.GetTheme("terminal_green")
	if !ok {
		t.Error("Expected to find terminal_green theme")
	}
	if theme.ID != "terminal_green" {
		t.Errorf("Expected 'terminal_green', got '%s'", theme.ID)
	}

	// Get non-existing theme
	_, ok = tm.GetTheme("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent theme")
	}
}

func TestGetAllThemes(t *testing.T) {
	tm := NewThemeManager()
	themes := tm.GetAllThemes()

	if len(themes) != 5 {
		t.Errorf("Expected 5 themes, got %d", len(themes))
	}

	// Check order
	expectedOrder := []string{"midnight", "blood_moon", "terminal_green", "silent_hill_fog", "high_contrast"}
	for i, theme := range themes {
		if theme.ID != expectedOrder[i] {
			t.Errorf("Expected theme %d to be '%s', got '%s'", i, expectedOrder[i], theme.ID)
		}
	}
}

func TestIsCurrentTheme(t *testing.T) {
	tm := NewThemeManager()

	if !tm.IsCurrentTheme("midnight") {
		t.Error("Expected midnight to be current")
	}

	if tm.IsCurrentTheme("blood_moon") {
		t.Error("Expected blood_moon not to be current")
	}

	tm.SetTheme("blood_moon")

	if tm.IsCurrentTheme("midnight") {
		t.Error("Expected midnight not to be current after switch")
	}

	if !tm.IsCurrentTheme("blood_moon") {
		t.Error("Expected blood_moon to be current after switch")
	}
}

func TestThemeColors(t *testing.T) {
	tm := NewThemeManager()

	// All themes should have valid colors
	for _, theme := range tm.GetAllThemes() {
		if theme.Colors.Primary == "" {
			t.Errorf("Theme %s missing Primary color", theme.ID)
		}
		if theme.Colors.Secondary == "" {
			t.Errorf("Theme %s missing Secondary color", theme.ID)
		}
		if theme.Colors.Accent == "" {
			t.Errorf("Theme %s missing Accent color", theme.ID)
		}
		if theme.Colors.Error == "" {
			t.Errorf("Theme %s missing Error color", theme.ID)
		}
		if theme.Colors.Success == "" {
			t.Errorf("Theme %s missing Success color", theme.ID)
		}
		if theme.Colors.Warning == "" {
			t.Errorf("Theme %s missing Warning color", theme.ID)
		}
	}
}

func TestGlobalManager(t *testing.T) {
	// GetManager should return same instance
	tm1 := GetManager()
	tm2 := GetManager()

	if tm1 != tm2 {
		t.Error("Expected GetManager to return same instance")
	}
}
