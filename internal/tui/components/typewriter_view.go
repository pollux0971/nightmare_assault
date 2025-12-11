package components

import (
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// TypewriterTickMsg signals it's time to display the next character.
type TypewriterTickMsg struct {
	Time time.Time
}

// TypewriterCompleteMsg signals typewriter animation is complete.
type TypewriterCompleteMsg struct{}

// TypewriterView renders text with typewriter effect.
type TypewriterView struct {
	config       config.TypewriterConfig
	buffer       *engine.StreamBuffer
	content      string
	cursorBlink  bool
	lastBlink    time.Time
	width        int
	height       int
	accessible   bool // Accessibility mode flag
	san          int  // Current SAN value
}

// NewTypewriterView creates a new typewriter view.
func NewTypewriterView(cfg config.TypewriterConfig, san int, accessible bool) TypewriterView {
	// Create streaming config from UI config
	streamConfig := engine.TypewriterConfig{
		MinCharsPerSecond: cfg.Speed,
		MaxCharsPerSecond: cfg.Speed,
		PunctuationDelay:  50 * time.Millisecond,
		ParagraphDelay:    100 * time.Millisecond,
		Enabled:           cfg.Enabled,
		ShowCursor:        cfg.ShowCursor,
		SAN:               san,
	}

	buffer := engine.NewStreamBuffer(streamConfig)

	return TypewriterView{
		config:      cfg,
		buffer:      buffer,
		cursorBlink: true,
		lastBlink:   time.Now(),
		accessible:  accessible,
		san:         san,
	}
}

// SetContent sets the content to display with typewriter effect.
func (m *TypewriterView) SetContent(content string) tea.Cmd {
	m.buffer.Append(content)

	// Set callback to update displayed content
	m.buffer.SetCallbacks(
		func(r rune) {
			m.content += string(r)
		},
		func() {
			// Animation complete
		},
	)

	if m.config.Enabled {
		m.buffer.Start()
		return m.tick()
	}

	// If disabled, show all content immediately
	m.content = content
	return nil
}

// Update handles messages.
func (m TypewriterView) Update(msg tea.Msg) (TypewriterView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle skip (Space or Enter)
		if msg.String() == " " || msg.String() == "enter" {
			if m.buffer.State() == engine.TypewriterPlaying {
				m.buffer.Skip()
				m.content = m.buffer.GetFull()
				return m, func() tea.Msg { return TypewriterCompleteMsg{} }
			}
		}

	case TypewriterTickMsg:
		// Update cursor blink
		if time.Since(m.lastBlink) > m.cursorBlinkInterval() {
			m.cursorBlink = !m.cursorBlink
			m.lastBlink = msg.Time
		}

		// Check if animation is still playing
		state := m.buffer.State()
		if state == engine.TypewriterPlaying {
			// Get current displayed content
			m.content = m.buffer.GetDisplayed()
			return m, m.tick()
		} else if state == engine.TypewriterDone {
			return m, func() tea.Msg { return TypewriterCompleteMsg{} }
		}
	}

	return m, nil
}

// View renders the typewriter view.
func (m TypewriterView) View() string {
	var b strings.Builder

	// Process content for special characters (AC9)
	rendered := m.renderContent(m.content)

	b.WriteString(rendered)

	// Add cursor if enabled and animation is playing
	if m.config.ShowCursor && m.buffer.State() == engine.TypewriterPlaying {
		if m.cursorBlink {
			b.WriteString(m.cursorChar())
		}
	}

	// Wrap to width
	return m.wrapText(b.String())
}

// SetSize sets the view dimensions.
func (m *TypewriterView) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetSAN updates the SAN value for horror effects.
func (m *TypewriterView) SetSAN(san int) {
	m.san = san
	// Update buffer config
	// Note: This would require exposing buffer config setter
}

// IsComplete returns true if typewriter animation is complete.
func (m TypewriterView) IsComplete() bool {
	state := m.buffer.State()
	return state == engine.TypewriterDone || state == engine.TypewriterSkipped
}

// Progress returns animation progress (0-100).
func (m TypewriterView) Progress() int {
	return m.buffer.Progress()
}

// tick creates a tick command for animation.
func (m TypewriterView) tick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return TypewriterTickMsg{Time: t}
	})
}

// cursorBlinkInterval returns cursor blink interval based on SAN.
func (m TypewriterView) cursorBlinkInterval() time.Duration {
	if m.san >= 40 {
		return 500 * time.Millisecond
	}

	// Low SAN: unstable cursor (AC8)
	// Faster blink when anxious
	base := 500 - (40-m.san)*10 // 500ms down to 100ms
	if base < 100 {
		base = 100
	}
	return time.Duration(base) * time.Millisecond
}

// cursorChar returns the cursor character.
func (m TypewriterView) cursorChar() string {
	return "▌"
}

// renderContent processes content for markdown, colors, and emojis (AC9).
func (m TypewriterView) renderContent(content string) string {
	// For now, pass through content as-is
	// Full markdown/ANSI processing would be done here
	// This is a placeholder for AC9 implementation

	// Handle basic markdown bold/italic
	content = m.processBasicMarkdown(content)

	return content
}

// processBasicMarkdown handles basic markdown formatting.
func (m TypewriterView) processBasicMarkdown(content string) string {
	// Bold: **text** or __text__
	// Italic: *text* or _text_
	// This is a simplified version

	// For proper implementation, we would use lipgloss styles
	// For now, preserve the markdown as-is
	return content
}

// wrapText wraps text to fit width.
func (m TypewriterView) wrapText(text string) string {
	if m.width <= 0 {
		return text
	}

	var lines []string
	currentLine := strings.Builder{}
	currentWidth := 0

	for _, r := range text {
		charWidth := runeWidth(r)

		if r == '\n' {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
			continue
		}

		if currentWidth+charWidth > m.width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}

		currentLine.WriteRune(r)
		currentWidth += charWidth
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Limit to height if specified
	if m.height > 0 && len(lines) > m.height {
		lines = lines[:m.height]
	}

	return strings.Join(lines, "\n")
}

// runeWidth returns the display width of a rune.
func runeWidth(r rune) int {
	// ASCII characters are 1 wide
	if r < 128 {
		return 1
	}

	// CJK characters are typically 2 wide
	// This is a simplified check
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
		return 2
	}
	if r >= 0x3400 && r <= 0x4DBF { // CJK Extension A
		return 2
	}
	if r >= 0x3040 && r <= 0x309F { // Hiragana
		return 2
	}
	if r >= 0x30A0 && r <= 0x30FF { // Katakana
		return 2
	}

	// Emoji are typically 2 wide (AC9)
	if r >= 0x1F300 && r <= 0x1F9FF {
		return 2
	}
	// Additional emoji ranges
	if r >= 0x2600 && r <= 0x27BF { // Miscellaneous Symbols
		return 2
	}

	// Default to 1
	return 1
}

// CountChars counts displayable characters (for speed calculation).
func CountChars(s string) int {
	return utf8.RuneCountInString(s)
}

// TypewriterStyle creates a style for typewriter text.
func TypewriterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1, 2)
}

// SkipHintStyle creates a style for skip hint.
func SkipHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
}

// SkipHint returns the skip hint text (AC10 - accessibility).
func SkipHint(accessible bool) string {
	if accessible {
		return "按空格鍵跳過動畫"
	}
	return "Space 跳過"
}
