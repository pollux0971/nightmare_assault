package components

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/effects"
)

// FreeTextInput provides a multi-line text input component.
// Story 7.3 AC2: Free text input with validation and character limit (200 chars).
// Story 7.5: Integrates SAN-based visual effects and control deprivation.
type FreeTextInput struct {
	textarea    textarea.Model
	maxLength   int
	focused     bool
	placeholder string
	cancelled   bool // Story 7.3 AC2: Support cancel operation
	currentSAN  int  // Story 7.5: Current SAN value for visual effects
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
		currentSAN:  100, // Story 7.5: Default to full SAN
	}
}

// GetValue returns the current input value.
func (f *FreeTextInput) GetValue() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.textarea.Value()
}

// GetValueWithSANEffects returns the input value with SAN-based character deletion applied.
// Story 7.5 AC4: Randomly delete 10-20% of characters when SAN 1-19.
func (f *FreeTextInput) GetValueWithSANEffects() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value := f.textarea.Value()

	// Story 7.5 AC4: Apply character deletion if SAN is very low
	visualEffects := effects.GetSANVisualEffects(f.currentSAN)
	if effects.ShouldApplyCharDeletion(visualEffects) {
		corruptedValue, _ := effects.ApplyCharacterDeletion(value, visualEffects)
		return corruptedValue
	}

	return value
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
	f.cancelled = false
	f.textarea.Blur()
}

// Cancel marks the input as cancelled.
// Story 7.3 AC2: Support cancel operation returning to choice list.
func (f *FreeTextInput) Cancel() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cancelled = true
}

// IsCancelled returns whether the input was cancelled.
func (f *FreeTextInput) IsCancelled() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cancelled
}

// SetSAN sets the current SAN value for visual effects.
// Story 7.5: SAN affects input box appearance and behavior.
func (f *FreeTextInput) SetSAN(san int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.currentSAN = san
}

// GetSAN returns the current SAN value.
func (f *FreeTextInput) GetSAN() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.currentSAN
}

// Validate validates the current input.
// Story 7.3 AC2: Input validation - no empty input, filter special characters.
//
// Returns error if:
//   - Input is empty or whitespace only
//   - Input contains dangerous patterns (injection attempts)
func (f *FreeTextInput) Validate() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value := strings.TrimSpace(f.textarea.Value())

	// AC2: No empty input
	if value == "" {
		return &ValidationError{Message: "輸入不能為空"}
	}

	// AC2: Filter special characters to prevent injection
	if containsDangerousPattern(value) {
		return &ValidationError{Message: "輸入包含不允許的字符或模式"}
	}

	return nil
}

// Sanitize returns a sanitized version of the input.
// Story 7.3 AC2: Filter special characters to prevent injection attacks.
func (f *FreeTextInput) Sanitize() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	value := f.textarea.Value()
	return sanitizeInput(value)
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
// Story 7.5: Applies SAN-based visual effects to input box.
func (f *FreeTextInput) View() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Story 7.5: Get visual effects based on current SAN
	visualEffects := effects.GetSANVisualEffects(f.currentSAN)

	// Base border styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	focusedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	// Story 7.5 AC2, AC4: Apply border color based on SAN
	if f.focused {
		focusedStyle = effects.ApplyInputBoxStyle(focusedStyle, visualEffects)
	} else {
		borderStyle = effects.ApplyInputBoxStyle(borderStyle, visualEffects)
	}

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

	// Story 7.5 AC2, AC3, AC4: Apply input box width scaling
	// Note: In a real implementation, this would adjust the textarea width
	// For now, we just apply the visual border effects
	content := style.Render(textArea)

	// Story 7.5 AC4: Add visual feedback if SAN is very low
	feedback := ""
	if f.currentSAN < 20 {
		feedback = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Italic(true).
			Render("[你的思緒正在崩潰...]")
	}

	return content + "\n" + counterText + feedback
}

// ==========================================================================
// Story 7.3 AC2: Input Validation and Sanitization
// ==========================================================================

// ValidationError represents a validation error.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// containsDangerousPattern checks for dangerous injection patterns.
// Story 7.3 AC2: Prevent prompt injection and other attacks.
func containsDangerousPattern(input string) bool {
	lowerInput := strings.ToLower(input)

	// Check for LLM prompt injection patterns
	llmPatterns := []string{
		"ignore previous",
		"ignore all previous",
		"disregard",
		"forget",
		"system:",
		"assistant:",
		"user:",
		"[inst]",
		"[/inst]",
		"<|",
		"|>",
	}

	for _, pattern := range llmPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for script injection patterns
	scriptPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
	}

	for _, pattern := range scriptPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for SQL injection patterns (just in case)
	sqlPatterns := []string{
		"drop table",
		"delete from",
		"; --",
		"' or '",
		"\" or \"",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for regex-based patterns
	dangerousRegexes := []string{
		`<[^>]+>`,         // HTML tags
		`\{\{[^}]+\}\}`,   // Template injection
		`\$\{[^}]+\}`,     // Variable substitution
	}

	for _, pattern := range dangerousRegexes {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}

	return false
}

// sanitizeInput filters and sanitizes input text.
// Story 7.3 AC2: Remove dangerous characters while preserving normal text.
func sanitizeInput(input string) string {
	// Remove control characters except space, tab, and newline
	var result strings.Builder
	for _, r := range input {
		// Allow printable characters and basic whitespace
		if r >= 32 || r == '\t' || r == '\n' {
			// Filter out specific dangerous characters
			if !isDangerousRune(r) {
				result.WriteRune(r)
			}
		}
	}

	sanitized := result.String()

	// Remove any remaining dangerous patterns
	// Replace markdown code blocks
	sanitized = regexp.MustCompile("```[^`]*```").ReplaceAllString(sanitized, "")

	// Remove HTML-like tags
	sanitized = regexp.MustCompile("<[^>]+>").ReplaceAllString(sanitized, "")

	return strings.TrimSpace(sanitized)
}

// isDangerousRune checks if a rune is potentially dangerous.
func isDangerousRune(r rune) bool {
	// Dangerous characters that could be used for injection
	dangerousRunes := []rune{
		'<', '>', '`', '|', '{', '}',
	}

	for _, dangerous := range dangerousRunes {
		if r == dangerous {
			return true
		}
	}

	return false
}
