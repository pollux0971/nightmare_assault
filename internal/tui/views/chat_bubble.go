// Package views provides TUI view implementations.
package views

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// ChatBubble renders a single chat message with appropriate styling.
// Story 3-5: ChatBubble message rendering component.
type ChatBubble struct {
	message ChatMessage
	width   int
	theme   *themes.Theme
	mu      sync.RWMutex
}

// NewChatBubble creates a new ChatBubble component.
// Story 3-5 AC1: ChatBubble component renders single message.
func NewChatBubble(msg ChatMessage, width int) *ChatBubble {
	return &ChatBubble{
		message: msg,
		width:   width,
		theme:   themes.GetManager().GetCurrentTheme(),
	}
}

// WithTheme sets a custom theme for the chat bubble (chainable).
func (cb *ChatBubble) WithTheme(theme *themes.Theme) *ChatBubble {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.theme = theme
	return cb
}

// SetMessage updates the message to display.
func (cb *ChatBubble) SetMessage(msg ChatMessage) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.message = msg
}

// SetWidth sets the maximum width for the chat bubble.
func (cb *ChatBubble) SetWidth(width int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.width = width
}

// View renders the chat bubble.
// Story 3-5 AC1-AC5: Complete message rendering with all features.
func (cb *ChatBubble) View() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Get all styled components
	timestamp := cb.formatTimestamp()
	speaker := cb.formatSpeaker()
	content := cb.message.Content
	flags := cb.getFlagIcons()

	// Determine if this is a player or system message
	isPlayer := cb.isPlayerMessage()
	isSystem := cb.isSystemMessage()

	// Get appropriate styles
	bubbleStyle := cb.getBubbleStyle(isPlayer, isSystem)
	speakerStyle := cb.getSpeakerStyle(isPlayer, isSystem)
	contentStyle := cb.getContentStyle()
	timestampStyle := cb.getTimestampStyle()
	flagStyle := cb.getFlagStyle()

	// Build the message line
	var messageLine string
	if isSystem {
		// System messages: centered with special formatting
		// Format: [12:34] [系統] Message content 🌀⚠️
		messageLine = fmt.Sprintf("%s %s %s %s",
			timestampStyle.Render(timestamp),
			speakerStyle.Render(speaker),
			contentStyle.Render(content),
			flagStyle.Render(flags),
		)
		messageLine = cb.centerText(messageLine)
	} else if isPlayer {
		// Player messages: right-aligned
		// Format: [12:34] 你: Message content 🌀⚠️
		messageLine = fmt.Sprintf("%s %s %s %s",
			timestampStyle.Render(timestamp),
			speakerStyle.Render(speaker+":"),
			contentStyle.Render(content),
			flagStyle.Render(flags),
		)
		messageLine = cb.rightAlign(messageLine)
	} else {
		// NPC messages: left-aligned
		// Format: [12:34] NPC名稱: Message content 🌀⚠️
		messageLine = fmt.Sprintf("%s %s %s %s",
			timestampStyle.Render(timestamp),
			speakerStyle.Render(speaker+":"),
			contentStyle.Render(content),
			flagStyle.Render(flags),
		)
	}

	// Apply bubble style and return
	return bubbleStyle.Render(messageLine)
}

// formatTimestamp formats the message timestamp.
// Story 3-5 AC4: Display timestamp.
func (cb *ChatBubble) formatTimestamp() string {
	return fmt.Sprintf("[%02d:%02d]", cb.message.Timestamp.Hour(), cb.message.Timestamp.Minute())
}

// formatSpeaker formats the speaker name.
// Story 3-5 AC2: Differentiate player/NPC/system messages.
func (cb *ChatBubble) formatSpeaker() string {
	if cb.isSystemMessage() {
		return "[系統]"
	}
	if cb.isPlayerMessage() {
		return "你"
	}
	return cb.message.Speaker
}

// getFlagIcons returns flag icons for the message.
// Story 3-5 AC5: Support flag display (hallucination, hostile, etc.).
func (cb *ChatBubble) getFlagIcons() string {
	if len(cb.message.Flags) == 0 {
		return ""
	}

	icons := make([]string, 0, len(cb.message.Flags))
	for _, flag := range cb.message.Flags {
		icon := cb.getFlagIcon(flag)
		if icon != "" {
			icons = append(icons, icon)
		}
	}

	if len(icons) == 0 {
		return ""
	}

	return strings.Join(icons, "")
}

// getFlagIcon returns the icon for a specific flag.
// Story 3-5 AC5: Flag icon mapping.
func (cb *ChatBubble) getFlagIcon(flag ChatFlag) string {
	switch flag {
	case ChatFlagHallucination:
		return "🌀" // Hallucination
	case ChatFlagHostile:
		return "⚠️" // Hostile
	case ChatFlagRevelation:
		return "💡" // Revelation
	case ChatFlagPersuasion:
		return "🎯" // Persuasion
	case ChatFlagLie:
		return "🎭" // Lie
	case ChatFlagContradiction:
		return "❌" // Contradiction
	default:
		return ""
	}
}

// isPlayerMessage returns true if this is a player message.
func (cb *ChatBubble) isPlayerMessage() bool {
	return cb.message.Speaker != "system" &&
		cb.message.Type != ChatMessageSystem &&
		cb.message.Speaker != "" &&
		// Simple heuristic: if speaker is "player" or "你" or contains player ID
		(cb.message.Speaker == "player" || cb.message.Speaker == "你")
}

// isSystemMessage returns true if this is a system message.
func (cb *ChatBubble) isSystemMessage() bool {
	return cb.message.Speaker == "system" || cb.message.Type == ChatMessageSystem
}

// getBubbleStyle returns the bubble style based on message type.
// Story 3-5 AC2: Different styles for player/NPC/system.
func (cb *ChatBubble) getBubbleStyle(isPlayer, isSystem bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Width(cb.width).
		Padding(0, 1)

	if isSystem {
		// System messages: subtle background
		return baseStyle.
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("250"))
	}

	if isPlayer {
		// Player messages: blue-ish background
		return baseStyle.
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("117")) // Light blue
	}

	// NPC messages: green-ish background
	return baseStyle.
		Background(lipgloss.Color("234")).
		Foreground(lipgloss.Color("121")) // Light green
}

// getSpeakerStyle returns the speaker name style.
// Story 3-5 AC2: Different speaker styles.
func (cb *ChatBubble) getSpeakerStyle(isPlayer, isSystem bool) lipgloss.Style {
	if isSystem {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")). // Yellow
			Bold(true)
	}

	if isPlayer {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")). // Light blue
			Bold(true)
	}

	// NPC
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("121")). // Light green
		Bold(true)
}

// getContentStyle returns the content style based on message type.
// Story 3-5 AC3: Different visual styles for ChatMessageType.
func (cb *ChatBubble) getContentStyle() lipgloss.Style {
	baseStyle := lipgloss.NewStyle()

	switch cb.message.Type {
	case ChatMessageSystem:
		// System: bold
		return baseStyle.
			Foreground(lipgloss.Color("250")).
			Bold(true)

	case ChatMessageWhisper:
		// Whisper: italic, lighter color
		return baseStyle.
			Foreground(lipgloss.Color("244")).
			Italic(true)

	case ChatMessageThought:
		// Thought: italic, gray
		return baseStyle.
			Foreground(lipgloss.Color("245")).
			Italic(true)

	case ChatMessageAction:
		// Action: italic, orange
		return baseStyle.
			Foreground(lipgloss.Color("214")).
			Italic(true)

	case ChatMessageNormal:
		fallthrough
	default:
		// Normal: standard style
		return baseStyle.Foreground(lipgloss.Color("252"))
	}
}

// getTimestampStyle returns the timestamp style.
// Story 3-5 AC4: Timestamp styling.
func (cb *ChatBubble) getTimestampStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dark gray
		Faint(true)
}

// getFlagStyle returns the flag icon style.
// Story 3-5 AC5: Flag styling.
func (cb *ChatBubble) getFlagStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Bold(true)
}

// centerText centers the text within the bubble width.
func (cb *ChatBubble) centerText(text string) string {
	// Remove ANSI codes for length calculation
	plainText := stripAnsiCodes(text)
	textLen := len(plainText)

	if textLen >= cb.width {
		return text
	}

	leftPadding := (cb.width - textLen) / 2
	return strings.Repeat(" ", leftPadding) + text
}

// rightAlign right-aligns the text within the bubble width.
func (cb *ChatBubble) rightAlign(text string) string {
	// Remove ANSI codes for length calculation
	plainText := stripAnsiCodes(text)
	textLen := len(plainText)

	if textLen >= cb.width {
		return text
	}

	leftPadding := cb.width - textLen - 2 // -2 for right margin
	if leftPadding < 0 {
		leftPadding = 0
	}
	return strings.Repeat(" ", leftPadding) + text
}

// stripAnsiCodes removes ANSI escape codes for length calculation.
// This is a simple implementation - lipgloss.Width() would be better but this is simpler.
func stripAnsiCodes(s string) string {
	// Simple heuristic: remove common ANSI sequences
	// This is not perfect but sufficient for width calculation
	result := s
	// Remove color codes and other ANSI sequences
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// Note: Type-specific visual styling (AC3) is implemented through getContentStyle()
// which applies italic, color changes, etc. Text prefixes were considered but
// removed in favor of pure visual styling for cleaner UX.
