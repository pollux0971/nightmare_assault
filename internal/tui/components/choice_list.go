// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ChoiceList manages a list of choices for player selection.
type ChoiceList struct {
	choices     []string
	selected    int
	highlighted bool
	mu          sync.RWMutex
}

// NewChoiceList creates a new choice list.
func NewChoiceList(choices []string) *ChoiceList {
	return &ChoiceList{
		choices:     choices,
		selected:    0,
		highlighted: false,
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
		return c.choices[index], true
	}
	return "", false
}

// GetSelectedChoice returns the currently selected choice.
func (c *ChoiceList) GetSelectedChoice() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.selected >= 0 && c.selected < len(c.choices) {
		return c.choices[c.selected], true
	}
	return "", false
}

// Count returns the number of choices.
func (c *ChoiceList) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.choices)
}

// SetChoices updates the choice list and resets selection.
func (c *ChoiceList) SetChoices(choices []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.choices = choices
	c.selected = 0
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

	for i, choice := range c.choices {
		number := fmt.Sprintf("%d. ", i+1)

		// Add default indicator for first choice
		if i == 0 {
			choice = choice + " [預設]"
		}

		var line string
		if c.highlighted && i == c.selected {
			line = highlightStyle.Render(number + choice)
		} else if i == c.selected {
			line = selectedStyle.Render(number + choice)
		} else {
			line = normalStyle.Render(number + choice)
		}

		output += line
		if i < len(c.choices)-1 {
			output += "\n"
		}
	}

	return output
}
