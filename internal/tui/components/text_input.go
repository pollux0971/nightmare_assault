package components

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FreeTextInput provides a multi-line text input component.
type FreeTextInput struct {
	textarea    textarea.Model
	maxLength   int
	focused     bool
	placeholder string
	mu          sync.RWMutex
}

// NewFreeTextInput creates a new free text input component.
func NewFreeTextInput() *FreeTextInput {
	ti := textarea.New()
	ti.Placeholder = "輸入你的行動... (按 Enter 送出, ESC 取消)"
	ti.CharLimit = 200
	ti.ShowLineNumbers = false
	ti.SetHeight(3)

	return &FreeTextInput{
		textarea:    ti,
		maxLength:   200,
		focused:     false,
		placeholder: "輸入你的行動...",
	}
}

// GetValue returns the current input value.
func (f *FreeTextInput) GetValue() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.textarea.Value()
}

// SetValue sets the input value.
func (f *FreeTextInput) SetValue(value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Truncate if too long
	runes := []rune(value)
	if len(runes) > f.maxLength {
		value = string(runes[:f.maxLength])
	}
	f.textarea.SetValue(value)
}

// Clear clears the input.
func (f *FreeTextInput) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.textarea.Reset()
}

// CharCount returns the number of characters.
func (f *FreeTextInput) CharCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len([]rune(f.textarea.Value()))
}

// IsEmpty returns true if the input is empty.
func (f *FreeTextInput) IsEmpty() bool {
	return f.CharCount() == 0
}

// IsFull returns true if the input is at max length.
func (f *FreeTextInput) IsFull() bool {
	return f.CharCount() >= f.maxLength
}

// RemainingChars returns the number of remaining characters.
func (f *FreeTextInput) RemainingChars() int {
	remaining := f.maxLength - f.CharCount()
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SetFocused sets the focused state.
func (f *FreeTextInput) SetFocused(focused bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.focused = focused
	if focused {
		f.textarea.Focus()
	} else {
		f.textarea.Blur()
	}
}

// IsFocused returns whether the input is focused.
func (f *FreeTextInput) IsFocused() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.focused
}

// SetPlaceholder sets the placeholder text.
func (f *FreeTextInput) SetPlaceholder(placeholder string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.placeholder = placeholder
	f.textarea.Placeholder = placeholder
}

// Reset resets the input to initial state.
func (f *FreeTextInput) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.textarea.Reset()
	f.focused = false
	f.textarea.Blur()
}

// Update handles BubbleTea messages.
func (f *FreeTextInput) Update(msg tea.Msg) tea.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()

	var cmd tea.Cmd
	f.textarea, cmd = f.textarea.Update(msg)
	return cmd
}

// View renders the input component.
func (f *FreeTextInput) View() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	focusedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	counterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	var style lipgloss.Style
	if f.focused {
		style = focusedStyle
	} else {
		style = borderStyle
	}

	counter := fmt.Sprintf("%d/%d", f.CharCount(), f.maxLength)
	counterText := counterStyle.Render(counter)

	textArea := f.textarea.View()
	content := style.Render(textArea)

	return content + "\n" + counterText
}
