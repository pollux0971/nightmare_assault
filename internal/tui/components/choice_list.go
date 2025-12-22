// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ChoiceOption represents a single choice option with metadata
type ChoiceOption struct {
	Text      string  // Choice text
	IsRisky   bool    // Whether this is a high-risk option (AC1)
	IsFreeText bool   // Whether this is the free text input option
}

// ChoiceList manages a list of choices for player selection.
// Story 7.3 AC1: Supports 2-3 predefined options + 1 free text option
// Story 7.3 AC1: Adjusts option count based on tension (low=3, high=2)
type ChoiceList struct {
	choices     []ChoiceOption
	selected    int
	highlighted bool
	tension     int  // Current tension level for dynamic option count (AC1)
	mu          sync.RWMutex
}

// NewChoiceList creates a new choice list from string slices.
// Deprecated: Use NewChoiceListWithOptions for Story 7.3 features.
func NewChoiceList(choices []string) *ChoiceList {
	options := make([]ChoiceOption, len(choices))
	for i, text := range choices {
		options[i] = ChoiceOption{
			Text:      text,
			IsRisky:   false,
			IsFreeText: false,
		}
	}
	return &ChoiceList{
		choices:     options,
		selected:    0,
		highlighted: false,
		tension:     0,
	}
}

// NewChoiceListWithOptions creates a new choice list with full options.
// Story 7.3 AC1: Supports risk marking and free text option.
func NewChoiceListWithOptions(choices []ChoiceOption, tension int) *ChoiceList {
	return &ChoiceList{
		choices:     choices,
		selected:    0,
		highlighted: false,
		tension:     tension,
	}
}

// GetSelected returns the currently selected choice index.
func (c *ChoiceList) GetSelected() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selected
}

// SetSelected sets the selected choice index.
func (c *ChoiceList) SetSelected(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index >= 0 && index < len(c.choices) {
		c.selected = index
	}
}

// GetChoice returns the choice at the given index.
func (c *ChoiceList) GetChoice(index int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if index >= 0 && index < len(c.choices) {
		return c.choices[index].Text, true
	}
	return "", false
}

// GetChoiceOption returns the full choice option at the given index.
// Story 7.3: Returns ChoiceOption with metadata (risk, free text flag).
func (c *ChoiceList) GetChoiceOption(index int) (ChoiceOption, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if index >= 0 && index < len(c.choices) {
		return c.choices[index], true
	}
	return ChoiceOption{}, false
}

// GetSelectedChoice returns the currently selected choice.
func (c *ChoiceList) GetSelectedChoice() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.selected >= 0 && c.selected < len(c.choices) {
		return c.choices[c.selected].Text, true
	}
	return "", false
}

// GetSelectedChoiceOption returns the currently selected choice option.
// Story 7.3: Returns full ChoiceOption with metadata.
func (c *ChoiceList) GetSelectedChoiceOption() (ChoiceOption, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.selected >= 0 && c.selected < len(c.choices) {
		return c.choices[c.selected], true
	}
	return ChoiceOption{}, false
}

// Count returns the number of choices.
func (c *ChoiceList) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.choices)
}

// SetChoices updates the choice list and resets selection.
// Deprecated: Use SetChoicesWithOptions for Story 7.3 features.
func (c *ChoiceList) SetChoices(choices []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	options := make([]ChoiceOption, len(choices))
	for i, text := range choices {
		options[i] = ChoiceOption{
			Text:      text,
			IsRisky:   false,
			IsFreeText: false,
		}
	}
	c.choices = options
	c.selected = 0
}

// SetChoicesWithOptions updates the choice list with full options.
// Story 7.3 AC1: Supports risk marking and tension-based filtering.
func (c *ChoiceList) SetChoicesWithOptions(choices []ChoiceOption, tension int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.choices = choices
	c.tension = tension
	c.selected = 0
}

// SetTension updates the tension level.
// Story 7.3 AC1: Used for dynamic option count adjustment.
func (c *ChoiceList) SetTension(tension int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tension = tension
}

// SetHighlighted sets the highlighted state.
func (c *ChoiceList) SetHighlighted(highlighted bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.highlighted = highlighted
}

// IsHighlighted returns whether the list is highlighted.
func (c *ChoiceList) IsHighlighted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.highlighted
}

// Reset resets the choice list to initial state.
func (c *ChoiceList) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selected = 0
	c.highlighted = false
}

// View renders the choice list.
// Story 7.3 AC1: Displays risk markers and free text indicators.
func (c *ChoiceList) View() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.choices) == 0 {
		return ""
	}

	var output string

	// Style definitions
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("211")).
		Background(lipgloss.Color("235")).
		Bold(true)

	// Story 7.3 AC1: High-risk option style (red/orange)
	riskyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")). // Orange for risky options
		Bold(true)

	riskyHighlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red when selected
		Background(lipgloss.Color("235")).
		Bold(true)

	// Free text option style
	freeTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")). // Light blue
		Italic(true)

	for i, choice := range c.choices {
		number := fmt.Sprintf("%d. ", i+1)
		text := choice.Text

		// Story 7.3 AC1: Add risk marker for high-risk options
		if choice.IsRisky {
			text = "⚠ " + text
		}

		// Story 7.3 AC2: Mark free text input option
		if choice.IsFreeText {
			text = "✎ " + text
		}

		// Add default indicator for first choice (if not free text)
		if i == 0 && !choice.IsFreeText {
			text = text + " [預設]"
		}

		var line string
		// Apply styles based on state and risk level
		if choice.IsFreeText {
			// Free text option gets its own style
			if c.highlighted && i == c.selected {
				line = highlightStyle.Render(number + text)
			} else if i == c.selected {
				line = selectedStyle.Render(number + text)
			} else {
				line = freeTextStyle.Render(number + text)
			}
		} else if choice.IsRisky {
			// Risky options get warning colors
			if c.highlighted && i == c.selected {
				line = riskyHighlightStyle.Render(number + text)
			} else if i == c.selected {
				line = riskyStyle.Render(number + text)
			} else {
				line = riskyStyle.Render(number + text)
			}
		} else {
			// Normal options
			if c.highlighted && i == c.selected {
				line = highlightStyle.Render(number + text)
			} else if i == c.selected {
				line = selectedStyle.Render(number + text)
			} else {
				line = normalStyle.Render(number + text)
			}
		}

		output += line
		if i < len(c.choices)-1 {
			output += "\n"
		}
	}

	return output
}
