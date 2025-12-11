// Package themes provides color theme definitions and management for Nightmare Assault.
package themes

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ThemeColors defines the color palette for a theme.
type ThemeColors struct {
	Primary    lipgloss.Color // Main text color
	Secondary  lipgloss.Color // Secondary text color
	Accent     lipgloss.Color // Highlight/selected items
	Background lipgloss.Color // Background color
	Border     lipgloss.Color // Border color
	Error      lipgloss.Color // Error messages
	Success    lipgloss.Color // Success messages
	Warning    lipgloss.Color // Warning messages
}

// Theme represents a complete color theme.
type Theme struct {
	ID          string      // Theme identifier (e.g., "midnight")
	Name        string      // Display name (e.g., "Midnight")
	Description string      // Short description
	Colors      ThemeColors // Color palette
}

// ThemeManager manages the current theme and available themes.
type ThemeManager struct {
	mu           sync.RWMutex
	currentTheme *Theme
	themes       map[string]*Theme
}

var (
	defaultManager     *ThemeManager
	defaultManagerOnce sync.Once
)

// GetManager returns the global ThemeManager instance.
func GetManager() *ThemeManager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewThemeManager()
	})
	return defaultManager
}

// NewThemeManager creates a new ThemeManager with all built-in themes.
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		themes: make(map[string]*Theme),
	}

	// Register all built-in themes
	tm.registerTheme(MidnightTheme())
	tm.registerTheme(BloodMoonTheme())
	tm.registerTheme(TerminalGreenTheme())
	tm.registerTheme(SilentHillFogTheme())
	tm.registerTheme(HighContrastTheme())

	// Set default theme
	tm.currentTheme = tm.themes["midnight"]

	return tm
}

func (tm *ThemeManager) registerTheme(t *Theme) {
	tm.themes[t.ID] = t
}

// GetCurrentTheme returns the currently active theme.
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTheme
}

// SetTheme changes the current theme by ID.
func (tm *ThemeManager) SetTheme(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	theme, ok := tm.themes[id]
	if !ok {
		return ErrThemeNotFound
	}

	tm.currentTheme = theme
	return nil
}

// GetTheme returns a theme by ID.
func (tm *ThemeManager) GetTheme(id string) (*Theme, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	theme, ok := tm.themes[id]
	return theme, ok
}

// ListThemes returns all available theme IDs.
func (tm *ThemeManager) ListThemes() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0, len(tm.themes))
	for id := range tm.themes {
		ids = append(ids, id)
	}
	return ids
}

// GetAllThemes returns all available themes.
func (tm *ThemeManager) GetAllThemes() []*Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]*Theme, 0, len(tm.themes))
	// Return in consistent order
	order := []string{"midnight", "blood_moon", "terminal_green", "silent_hill_fog", "high_contrast"}
	for _, id := range order {
		if t, ok := tm.themes[id]; ok {
			themes = append(themes, t)
		}
	}
	return themes
}

// IsCurrentTheme checks if the given theme ID is the current theme.
func (tm *ThemeManager) IsCurrentTheme(id string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentTheme != nil && tm.currentTheme.ID == id
}

// Errors
var (
	ErrThemeNotFound = &ThemeError{Message: "theme not found"}
)

// ThemeError represents a theme-related error.
type ThemeError struct {
	Message string
}

func (e *ThemeError) Error() string {
	return e.Message
}
