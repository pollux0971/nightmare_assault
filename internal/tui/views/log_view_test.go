package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewLogView(t *testing.T) {
	entries := []game.LogEntry{
		{
			Timestamp: time.Now(),
			Type:      game.LogNarrative,
			Content:   "Test narrative",
		},
	}

	lv := NewLogView(entries, 80, 24)

	if lv.width != 80 {
		t.Errorf("Expected width 80, got %d", lv.width)
	}

	if lv.height != 24 {
		t.Errorf("Expected height 24, got %d", lv.height)
	}

	if len(lv.entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(lv.entries))
	}
}

func TestLogViewInit(t *testing.T) {
	entries := []game.LogEntry{}
	lv := NewLogView(entries, 80, 24)

	cmd := lv.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestLogViewUpdateQuit(t *testing.T) {
	entries := []game.LogEntry{}
	lv := NewLogView(entries, 80, 24)

	// Simulate 'q' key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := lv.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command")
	}
}

func TestLogViewUpdateEsc(t *testing.T) {
	entries := []game.LogEntry{}
	lv := NewLogView(entries, 80, 24)

	// Simulate ESC key press
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := lv.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command on ESC")
	}
}

func TestLogViewFormatEntry(t *testing.T) {
	entries := []game.LogEntry{}
	lv := NewLogView(entries, 80, 24)

	now := time.Date(2025, 12, 11, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		entry    game.LogEntry
		contains []string
	}{
		{
			name: "Narrative entry",
			entry: game.LogEntry{
				Timestamp: now,
				Type:      game.LogNarrative,
				Content:   "Test narrative",
			},
			contains: []string{"15:30:45", "[敘事]", "Test narrative"},
		},
		{
			name: "Player input entry",
			entry: game.LogEntry{
				Timestamp: now,
				Type:      game.LogPlayerInput,
				Content:   "Player command",
			},
			contains: []string{"15:30:45", "[玩家]", "Player command"},
		},
		{
			name: "Option choice entry",
			entry: game.LogEntry{
				Timestamp: now,
				Type:      game.LogOptionChoice,
				Content:   "Choice 1",
			},
			contains: []string{"15:30:45", "[選項]", "Choice 1"},
		},
		{
			name: "System message entry",
			entry: game.LogEntry{
				Timestamp: now,
				Type:      game.LogSystem,
				Content:   "System message",
			},
			contains: []string{"15:30:45", "[系統]", "System message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := lv.formatLogEntry(tt.entry)

			for _, expected := range tt.contains {
				// Remove ANSI codes for testing
				plainText := stripAnsi(formatted)
				if !strings.Contains(plainText, expected) {
					t.Errorf("Expected formatted entry to contain '%s', got: %s", expected, plainText)
				}
			}
		})
	}
}

func TestLogViewSetEntries(t *testing.T) {
	entries := []game.LogEntry{
		{
			Timestamp: time.Now(),
			Type:      game.LogNarrative,
			Content:   "Entry 1",
		},
	}

	lv := NewLogView(entries, 80, 24)

	if len(lv.entries) != 1 {
		t.Fatalf("Expected 1 initial entry, got %d", len(lv.entries))
	}

	newEntries := []game.LogEntry{
		{
			Timestamp: time.Now(),
			Type:      game.LogNarrative,
			Content:   "Entry 2",
		},
		{
			Timestamp: time.Now(),
			Type:      game.LogPlayerInput,
			Content:   "Entry 3",
		},
	}

	lv.SetEntries(newEntries)

	if len(lv.entries) != 2 {
		t.Errorf("Expected 2 entries after SetEntries, got %d", len(lv.entries))
	}

	if lv.entries[0].Content != "Entry 2" {
		t.Errorf("Expected first entry 'Entry 2', got '%s'", lv.entries[0].Content)
	}
}

func TestLogViewView(t *testing.T) {
	entries := []game.LogEntry{
		{
			Timestamp: time.Now(),
			Type:      game.LogNarrative,
			Content:   "Test content",
		},
	}

	lv := NewLogView(entries, 80, 24)

	// Before ready, should show loading
	view := lv.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Expected 'Loading...' before ready")
	}

	// Simulate window size message to make it ready
	lv.ready = true
	view = lv.View()

	plainView := stripAnsi(view)

	// Check for header
	if !strings.Contains(plainView, "遊戲日誌") {
		t.Error("Expected header to contain '遊戲日誌'")
	}

	// Check for footer
	if !strings.Contains(plainView, "[↑↓] 捲動") {
		t.Error("Expected footer to contain '[↑↓] 捲動'")
	}

	if !strings.Contains(plainView, "[ESC/Q] 關閉") {
		t.Error("Expected footer to contain '[ESC/Q] 關閉'")
	}
}

// stripAnsi removes ANSI escape codes from a string for testing.
func stripAnsi(str string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range str {
		if r == '\x1b' {
			inEscape = true
			continue
		}

		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		result.WriteRune(r)
	}

	return result.String()
}
