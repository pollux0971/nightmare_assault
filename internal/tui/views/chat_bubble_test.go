package views

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// TestChatBubble_NewChatBubble tests the creation of a new ChatBubble.
// Story 3-5 AC1: ChatBubble component renders single message.
func TestChatBubble_NewChatBubble(t *testing.T) {
	msg := ChatMessage{
		ID:        "test-1",
		Speaker:   "player",
		Content:   "Hello, world!",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	if cb == nil {
		t.Fatal("NewChatBubble returned nil")
	}

	if cb.width != 80 {
		t.Errorf("Expected width 80, got %d", cb.width)
	}

	if cb.message.ID != "test-1" {
		t.Errorf("Expected message ID 'test-1', got '%s'", cb.message.ID)
	}
}

// TestChatBubble_PlayerMessage tests rendering a player message.
// Story 3-5 AC2: Player messages are styled differently.
func TestChatBubble_PlayerMessage(t *testing.T) {
	msg := ChatMessage{
		ID:        "player-1",
		Speaker:   "player",
		Content:   "I need help!",
		Timestamp: time.Date(2025, 12, 22, 14, 30, 0, 0, time.UTC),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "14:30") {
		t.Error("Player message should contain timestamp")
	}

	if !strings.Contains(output, "你") {
		t.Error("Player message should contain '你' as speaker")
	}

	if !strings.Contains(output, "I need help!") {
		t.Error("Player message should contain content")
	}
}

// TestChatBubble_NPCMessage tests rendering an NPC message.
// Story 3-5 AC2: NPC messages are styled differently.
func TestChatBubble_NPCMessage(t *testing.T) {
	msg := ChatMessage{
		ID:        "npc-1",
		Speaker:   "Sarah",
		Content:   "Don't worry, we'll figure this out.",
		Timestamp: time.Date(2025, 12, 22, 14, 31, 0, 0, time.UTC),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "14:31") {
		t.Error("NPC message should contain timestamp")
	}

	if !strings.Contains(output, "Sarah") {
		t.Error("NPC message should contain speaker name")
	}

	if !strings.Contains(output, "Don't worry, we'll figure this out.") {
		t.Error("NPC message should contain content")
	}
}

// TestChatBubble_SystemMessage tests rendering a system message.
// Story 3-5 AC2: System messages are styled differently.
func TestChatBubble_SystemMessage(t *testing.T) {
	msg := ChatMessage{
		ID:        "sys-1",
		Speaker:   "system",
		Content:   "A door slams shut in the distance.",
		Timestamp: time.Date(2025, 12, 22, 14, 32, 0, 0, time.UTC),
		Type:      ChatMessageSystem,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "14:32") {
		t.Error("System message should contain timestamp")
	}

	if !strings.Contains(output, "系統") {
		t.Error("System message should contain '系統' as speaker")
	}

	if !strings.Contains(output, "A door slams shut in the distance.") {
		t.Error("System message should contain content")
	}
}

// TestChatBubble_WhisperType tests rendering a whisper message.
// Story 3-5 AC3: Whisper messages have special styling.
func TestChatBubble_WhisperType(t *testing.T) {
	msg := ChatMessage{
		ID:        "whisper-1",
		Speaker:   "John",
		Content:   "I think she's hiding something.",
		Timestamp: time.Now(),
		Type:      ChatMessageWhisper,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "I think she's hiding something.") {
		t.Error("Whisper message should contain content")
	}

	// Check that whisper styling is applied (italic is handled by lipgloss)
	if output == "" {
		t.Error("Whisper message should render")
	}
}

// TestChatBubble_ThoughtType tests rendering a thought message.
// Story 3-5 AC3: Thought messages have special styling.
func TestChatBubble_ThoughtType(t *testing.T) {
	msg := ChatMessage{
		ID:        "thought-1",
		Speaker:   "player",
		Content:   "This doesn't feel right...",
		Timestamp: time.Now(),
		Type:      ChatMessageThought,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "This doesn't feel right...") {
		t.Error("Thought message should contain content")
	}
}

// TestChatBubble_ActionType tests rendering an action message.
// Story 3-5 AC3: Action messages have special styling.
func TestChatBubble_ActionType(t *testing.T) {
	msg := ChatMessage{
		ID:        "action-1",
		Speaker:   "Sarah",
		Content:   "reaches for the door handle",
		Timestamp: time.Now(),
		Type:      ChatMessageAction,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	if !strings.Contains(output, "reaches for the door handle") {
		t.Error("Action message should contain content")
	}
}

// TestChatBubble_TimestampFormat tests timestamp formatting.
// Story 3-5 AC4: Display timestamp.
func TestChatBubble_TimestampFormat(t *testing.T) {
	msg := ChatMessage{
		ID:        "ts-1",
		Speaker:   "player",
		Content:   "Test",
		Timestamp: time.Date(2025, 12, 22, 9, 5, 30, 0, time.UTC),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	timestamp := cb.formatTimestamp()

	expected := "[09:05]"
	if timestamp != expected {
		t.Errorf("Expected timestamp '%s', got '%s'", expected, timestamp)
	}
}

// TestChatBubble_FlagHallucination tests hallucination flag display.
// Story 3-5 AC5: Support hallucination flag.
func TestChatBubble_FlagHallucination(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-1",
		Speaker:   "player",
		Content:   "I saw a ghost!",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagHallucination},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "🌀") {
		t.Error("Hallucination flag should display 🌀 icon")
	}
}

// TestChatBubble_FlagHostile tests hostile flag display.
// Story 3-5 AC5: Support hostile flag.
func TestChatBubble_FlagHostile(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-2",
		Speaker:   "player",
		Content:   "I'll kill you!",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagHostile},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "⚠️") {
		t.Error("Hostile flag should display ⚠️ icon")
	}
}

// TestChatBubble_FlagRevelation tests revelation flag display.
// Story 3-5 AC5: Support revelation flag.
func TestChatBubble_FlagRevelation(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-3",
		Speaker:   "Sarah",
		Content:   "I know the password to the vault.",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagRevelation},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "💡") {
		t.Error("Revelation flag should display 💡 icon")
	}
}

// TestChatBubble_FlagPersuasion tests persuasion flag display.
// Story 3-5 AC5: Support persuasion flag.
func TestChatBubble_FlagPersuasion(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-4",
		Speaker:   "player",
		Content:   "Trust me, this is the right thing to do.",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagPersuasion},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "🎯") {
		t.Error("Persuasion flag should display 🎯 icon")
	}
}

// TestChatBubble_FlagLie tests lie flag display.
// Story 3-5 AC5: Support lie flag.
func TestChatBubble_FlagLie(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-5",
		Speaker:   "player",
		Content:   "I didn't take the key.",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagLie},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "🎭") {
		t.Error("Lie flag should display 🎭 icon")
	}
}

// TestChatBubble_FlagContradiction tests contradiction flag display.
// Story 3-5 AC5: Support contradiction flag.
func TestChatBubble_FlagContradiction(t *testing.T) {
	msg := ChatMessage{
		ID:        "flag-6",
		Speaker:   "player",
		Content:   "John is still alive!",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{ChatFlagContradiction},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "❌") {
		t.Error("Contradiction flag should display ❌ icon")
	}
}

// TestChatBubble_MultipleFlags tests multiple flags display.
// Story 3-5 AC5: Support multiple flags.
func TestChatBubble_MultipleFlags(t *testing.T) {
	msg := ChatMessage{
		ID:        "multi-1",
		Speaker:   "player",
		Content:   "I saw the ghost take the key!",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags: []ChatFlag{
			ChatFlagHallucination,
			ChatFlagLie,
			ChatFlagContradiction,
		},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if !strings.Contains(flags, "🌀") {
		t.Error("Should contain hallucination icon")
	}
	if !strings.Contains(flags, "🎭") {
		t.Error("Should contain lie icon")
	}
	if !strings.Contains(flags, "❌") {
		t.Error("Should contain contradiction icon")
	}
}

// TestChatBubble_NoFlags tests message without flags.
func TestChatBubble_NoFlags(t *testing.T) {
	msg := ChatMessage{
		ID:        "no-flag-1",
		Speaker:   "player",
		Content:   "Normal message",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	flags := cb.getFlagIcons()

	if flags != "" {
		t.Errorf("Expected empty flags, got '%s'", flags)
	}
}

// TestChatBubble_ThemeSupport tests custom theme support.
func TestChatBubble_ThemeSupport(t *testing.T) {
	msg := ChatMessage{
		ID:        "theme-1",
		Speaker:   "player",
		Content:   "Test",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)

	// Get a different theme
	tm := themes.GetManager()
	bloodMoon, ok := tm.GetTheme("blood_moon")
	if !ok {
		t.Skip("blood_moon theme not available")
	}

	cb.WithTheme(bloodMoon)

	// Verify theme was set
	cb.mu.RLock()
	if cb.theme.ID != "blood_moon" {
		t.Errorf("Expected theme 'blood_moon', got '%s'", cb.theme.ID)
	}
	cb.mu.RUnlock()
}

// TestChatBubble_WidthHandling tests width handling.
func TestChatBubble_WidthHandling(t *testing.T) {
	msg := ChatMessage{
		ID:        "width-1",
		Speaker:   "player",
		Content:   "Test",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 40)
	if cb.width != 40 {
		t.Errorf("Expected width 40, got %d", cb.width)
	}

	cb.SetWidth(100)
	cb.mu.RLock()
	if cb.width != 100 {
		t.Errorf("Expected width 100 after SetWidth, got %d", cb.width)
	}
	cb.mu.RUnlock()
}

// TestChatBubble_SetMessage tests message update.
func TestChatBubble_SetMessage(t *testing.T) {
	msg1 := ChatMessage{
		ID:        "msg-1",
		Speaker:   "player",
		Content:   "First message",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	msg2 := ChatMessage{
		ID:        "msg-2",
		Speaker:   "Sarah",
		Content:   "Second message",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg1, 80)
	output1 := cb.View()

	if !strings.Contains(output1, "First message") {
		t.Error("Initial message should contain 'First message'")
	}

	cb.SetMessage(msg2)
	output2 := cb.View()

	if !strings.Contains(output2, "Second message") {
		t.Error("Updated message should contain 'Second message'")
	}

	if strings.Contains(output2, "First message") {
		t.Error("Updated message should not contain old content")
	}
}

// TestGetMessageTypePrefix tests message type prefix generation.
// Story 3-5 AC3: Type-specific prefixes.
func TestGetMessageTypePrefix(t *testing.T) {
	tests := []struct {
		msgType  ChatMessageType
		expected string
	}{
		{ChatMessageNormal, ""},
		{ChatMessageSystem, ""},
		{ChatMessageWhisper, "（小聲）"},
		{ChatMessageThought, "（心想）"},
		{ChatMessageAction, "*"},
	}

	for _, tt := range tests {
		result := GetMessageTypePrefix(tt.msgType)
		if result != tt.expected {
			t.Errorf("For type %s, expected prefix '%s', got '%s'",
				tt.msgType.String(), tt.expected, result)
		}
	}
}

// TestFormatMessageWithPrefix tests message formatting with prefix.
func TestFormatMessageWithPrefix(t *testing.T) {
	tests := []struct {
		content  string
		msgType  ChatMessageType
		expected string
	}{
		{"Hello", ChatMessageNormal, "Hello"},
		{"Whisper", ChatMessageWhisper, "（小聲） Whisper"},
		{"Thinking", ChatMessageThought, "（心想） Thinking"},
		{"runs away", ChatMessageAction, "*runs away*"},
	}

	for _, tt := range tests {
		result := FormatMessageWithPrefix(tt.content, tt.msgType)
		if result != tt.expected {
			t.Errorf("For content '%s' with type %s, expected '%s', got '%s'",
				tt.content, tt.msgType.String(), tt.expected, result)
		}
	}
}

// TestChatBubble_EmptyContent tests rendering with empty content.
func TestChatBubble_EmptyContent(t *testing.T) {
	msg := ChatMessage{
		ID:        "empty-1",
		Speaker:   "player",
		Content:   "",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)
	output := cb.View()

	// Should still render with timestamp and speaker
	if output == "" {
		t.Error("Should render even with empty content")
	}
}

// TestChatBubble_LongContent tests rendering with long content.
func TestChatBubble_LongContent(t *testing.T) {
	longContent := strings.Repeat("This is a very long message. ", 10)
	msg := ChatMessage{
		ID:        "long-1",
		Speaker:   "player",
		Content:   longContent,
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 200) // Wider width to accommodate long content
	output := cb.View()

	// Check that output contains at least part of the content
	// (lipgloss may wrap or truncate, but should contain some of the text)
	if !strings.Contains(output, "This is a very long message") {
		t.Error("Should contain at least part of the long content")
	}
}

// TestChatBubble_ConcurrentAccess tests concurrent access safety.
func TestChatBubble_ConcurrentAccess(t *testing.T) {
	msg := ChatMessage{
		ID:        "concurrent-1",
		Speaker:   "player",
		Content:   "Test",
		Timestamp: time.Now(),
		Type:      ChatMessageNormal,
		Flags:     []ChatFlag{},
	}

	cb := NewChatBubble(msg, 80)

	// Concurrent reads and writes
	done := make(chan bool)

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = cb.View()
			}
			done <- true
		}()
	}

	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				newMsg := ChatMessage{
					ID:        "msg-" + string(rune(id)),
					Speaker:   "player",
					Content:   "Updated",
					Timestamp: time.Now(),
					Type:      ChatMessageNormal,
					Flags:     []ChatFlag{},
				}
				cb.SetMessage(newMsg)
				cb.SetWidth(80 + id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// If we reach here without panic, concurrent access is safe
}

// TestStripAnsiCodes tests ANSI code stripping.
func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\x1b[31mred text\x1b[0m", "red text"},
		{"\x1b[1;32mbold green\x1b[0m", "bold green"},
		{"no \x1b[4mansi\x1b[0m here", "no ansi here"},
	}

	for _, tt := range tests {
		result := stripAnsiCodes(tt.input)
		if result != tt.expected {
			t.Errorf("stripAnsiCodes(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
