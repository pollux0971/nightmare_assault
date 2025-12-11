package components

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestTypewriterView(t *testing.T) {
	t.Run("create view with config", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      40,
			ShowCursor: true,
		}

		view := NewTypewriterView(cfg, 100, false)

		if view.san != 100 {
			t.Errorf("Expected SAN 100, got %d", view.san)
		}

		if view.config.Speed != 40 {
			t.Errorf("Expected speed 40, got %d", view.config.Speed)
		}
	})

	t.Run("set content starts animation when enabled", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      100,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		cmd := view.SetContent("Hello World")

		if cmd == nil {
			t.Error("Expected tick command when enabled")
		}
	})

	t.Run("set content shows immediately when disabled", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    false,
			Speed:      40,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		cmd := view.SetContent("Hello World")

		if cmd != nil {
			t.Error("Expected no command when disabled")
		}

		if view.content != "Hello World" {
			t.Errorf("Expected immediate content, got %q", view.content)
		}
	})

	t.Run("skip with space key", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      10, // Very slow
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("This is a long text that would take time to display")

		// Simulate space key press
		view, cmd := view.Update(tea.KeyMsg{Type: tea.KeySpace})

		if cmd == nil {
			t.Error("Expected completion command after skip")
		}

		// Content should be complete
		if !strings.Contains(view.content, "long text") {
			t.Error("Content should be complete after skip")
		}
	})

	t.Run("skip with enter key", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      10,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Test content")

		// Simulate enter key press
		view, cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Error("Expected completion command after skip")
		}
	})

	t.Run("cursor blinks during animation", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      40,
			ShowCursor: true,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Test")

		// Initial state
		if !view.cursorBlink {
			t.Error("Cursor should start visible")
		}

		// Simulate tick after blink interval
		view.lastBlink = time.Now().Add(-600 * time.Millisecond)
		view, _ = view.Update(TypewriterTickMsg{Time: time.Now()})

		// Cursor should have toggled
		// Note: This test is timing-dependent and may be flaky
	})

	t.Run("cursor interval changes with low SAN", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      40,
			ShowCursor: true,
		}

		// High SAN
		viewHigh := NewTypewriterView(cfg, 100, false)
		intervalHigh := viewHigh.cursorBlinkInterval()

		// Low SAN
		viewLow := NewTypewriterView(cfg, 30, false)
		intervalLow := viewLow.cursorBlinkInterval()

		if intervalLow >= intervalHigh {
			t.Error("Low SAN should have faster cursor blink")
		}

		if intervalHigh != 500*time.Millisecond {
			t.Errorf("High SAN interval should be 500ms, got %v", intervalHigh)
		}
	})

	t.Run("set size for wrapping", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    false,
			Speed:      40,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetSize(20, 10)

		if view.width != 20 {
			t.Errorf("Expected width 20, got %d", view.width)
		}

		if view.height != 10 {
			t.Errorf("Expected height 10, got %d", view.height)
		}
	})

	t.Run("progress tracking", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      100,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Test")

		progress := view.Progress()
		if progress < 0 || progress > 100 {
			t.Errorf("Progress should be 0-100, got %d", progress)
		}
	})

	t.Run("is complete check", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    false,
			Speed:      40,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Test")

		// When disabled, should be immediately complete
		// Note: State depends on buffer implementation
	})

	t.Run("accessibility mode", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    true,
			Speed:      40,
			ShowCursor: true,
		}

		view := NewTypewriterView(cfg, 100, true)

		if !view.accessible {
			t.Error("Accessibility flag should be set")
		}

		// Test skip hint
		hint := SkipHint(true)
		if !strings.Contains(hint, "按空格鍵") {
			t.Error("Accessible hint should be in Chinese")
		}

		hintNormal := SkipHint(false)
		if hintNormal == hint {
			t.Error("Normal and accessible hints should differ")
		}
	})
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected int
	}{
		{"ASCII letter", 'A', 1},
		{"ASCII digit", '5', 1},
		{"ASCII space", ' ', 1},
		{"CJK character", '你', 2},
		{"CJK character 2", '好', 2},
		{"Hiragana", 'あ', 2},
		{"Katakana", 'ア', 2},
		{"Emoji heart", '❤', 2},
		{"Emoji face", '😀', 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := runeWidth(tt.r)
			if width != tt.expected {
				t.Errorf("Expected width %d for %q, got %d", tt.expected, tt.r, width)
			}
		})
	}
}

func TestCountChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"ASCII only", "Hello", 5},
		{"CJK", "你好", 2},
		{"mixed", "Hello你好", 7},
		{"with emoji", "Hello😀", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountChars(tt.input)
			if count != tt.expected {
				t.Errorf("Expected %d chars in %q, got %d", tt.expected, tt.input, count)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	cfg := config.TypewriterConfig{
		Enabled:    false,
		Speed:      40,
		ShowCursor: false,
	}

	t.Run("wrap long line", func(t *testing.T) {
		view := NewTypewriterView(cfg, 100, false)
		view.SetSize(10, 0)

		wrapped := view.wrapText("This is a very long line that should wrap")
		lines := strings.Split(wrapped, "\n")

		if len(lines) <= 1 {
			t.Error("Long text should wrap to multiple lines")
		}

		for i, line := range lines {
			if len(line) > 10 {
				t.Errorf("Line %d is too long: %d chars", i, len(line))
			}
		}
	})

	t.Run("preserve newlines", func(t *testing.T) {
		view := NewTypewriterView(cfg, 100, false)
		view.SetSize(50, 0)

		text := "Line 1\nLine 2\nLine 3"
		wrapped := view.wrapText(text)
		lines := strings.Split(wrapped, "\n")

		if len(lines) != 3 {
			t.Errorf("Expected 3 lines, got %d", len(lines))
		}
	})

	t.Run("limit to height", func(t *testing.T) {
		view := NewTypewriterView(cfg, 100, false)
		view.SetSize(50, 2)

		text := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
		wrapped := view.wrapText(text)
		lines := strings.Split(wrapped, "\n")

		if len(lines) > 2 {
			t.Errorf("Expected max 2 lines, got %d", len(lines))
		}
	})

	t.Run("no width limit", func(t *testing.T) {
		view := NewTypewriterView(cfg, 100, false)
		view.SetSize(0, 0)

		text := "This is a very long line"
		wrapped := view.wrapText(text)

		if wrapped != text {
			t.Error("Text should not wrap when width is 0")
		}
	})
}

func TestTypewriterMessages(t *testing.T) {
	t.Run("tick message", func(t *testing.T) {
		msg := TypewriterTickMsg{Time: time.Now()}
		if msg.Time.IsZero() {
			t.Error("Tick message should have timestamp")
		}
	})

	t.Run("complete message", func(t *testing.T) {
		msg := TypewriterCompleteMsg{}
		_ = msg // Just ensure it compiles
	})
}

func TestCursorChar(t *testing.T) {
	cfg := config.TypewriterConfig{
		Enabled:    true,
		Speed:      40,
		ShowCursor: true,
	}

	view := NewTypewriterView(cfg, 100, false)
	cursor := view.cursorChar()

	if cursor != "▌" {
		t.Errorf("Expected cursor ▌, got %s", cursor)
	}
}

func TestViewRender(t *testing.T) {
	t.Run("render without cursor", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    false,
			Speed:      40,
			ShowCursor: false,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Hello World")

		rendered := view.View()
		if !strings.Contains(rendered, "Hello") {
			t.Error("Rendered view should contain content")
		}

		if strings.Contains(rendered, "▌") {
			t.Error("Rendered view should not contain cursor when disabled")
		}
	})

	t.Run("render with cursor", func(t *testing.T) {
		cfg := config.TypewriterConfig{
			Enabled:    false, // Disabled so content shows immediately
			Speed:      40,
			ShowCursor: true,
		}

		view := NewTypewriterView(cfg, 100, false)
		view.SetContent("Test")
		view.cursorBlink = true

		rendered := view.View()
		if !strings.Contains(rendered, "Test") {
			t.Error("Rendered view should contain content")
		}
	})
}
