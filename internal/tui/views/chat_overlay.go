package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 3-4: Participant Management System
// ==========================================================================

// ChatParticipant represents a participant in the chat session.
// AC1: Contains ID, Name, IsPlayer, Emotion, IsActive
type ChatParticipant struct {
	ID       string               `json:"id"`        // Unique identifier (npc_id or "player")
	Name     string               `json:"name"`      // Display name
	IsPlayer bool                 `json:"is_player"` // Whether this is the player
	Emotion  manager.EmotionState `json:"emotion"`   // Current emotion state (Trust/Fear/Stress)
	IsActive bool                 `json:"is_active"` // Whether currently active in chat
}

// NewChatParticipant creates a new ChatParticipant instance.
func NewChatParticipant(id, name string, isPlayer bool, emotion manager.EmotionState) ChatParticipant {
	return ChatParticipant{
		ID:       id,
		Name:     name,
		IsPlayer: isPlayer,
		Emotion:  emotion,
		IsActive: true, // Default to active when created
	}
}

// ChatOverlayModel represents the chat overlay view.
// Story 3.1: ChatOverlayModel TUI 基礎 - Complete implementation
// Story 3.4: Participant management
// Story 3.6: Basic input handling
type ChatOverlayModel struct {
	// Story 3.1 AC1: Core state fields
	active       bool              // Whether chat overlay is active
	initiator    string            // ID of who initiated the chat (player or NPC ID)
	participants []ChatParticipant // All participants in the chat
	messages     []*ChatMessage    // Chat message history

	// Story 3.1 AC2: Bubble Tea TUI components
	inputField textinput.Model // Text input for player messages
	viewport   viewport.Model  // Scrollable message display area
	width      int             // Current terminal width
	height     int             // Current terminal height

	// Story 3.1 AC3: Time control fields
	timeScale       float64 // Time flow rate (0.1 = 10 chat turns per main turn)
	tickAccumulator float64 // Accumulated time ticks
	chatTurns       int     // Number of chat turns taken

	// Internal state
	focused      bool      // Whether input field is focused
	location     string    // Current location where chat is happening
	sessionID    string    // Unique session identifier
	sessionStart time.Time // When the session started
}

// NewChatOverlayModel creates a new ChatOverlayModel instance.
// Story 3.1 AC2: Properly initializes all TUI components with default values.
func NewChatOverlayModel() ChatOverlayModel {
	// Story 3.1 AC2: Initialize text input
	ti := textinput.New()
	ti.Placeholder = "輸入訊息... (Enter 發送 | ESC 清空 | Tab 退出)"
	ti.CharLimit = 200
	ti.Width = 50

	// Story 3.1 AC2: Initialize viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A3A3A"))

	return ChatOverlayModel{
		// Story 3.1 AC1: Initialize core state
		active:       false,
		initiator:    "",
		participants: []ChatParticipant{},
		messages:     []*ChatMessage{},

		// Story 3.1 AC2: Initialize TUI components
		inputField: ti,
		viewport:   vp,
		width:      80,
		height:     24,

		// Story 3.1 AC3: Initialize time control
		timeScale:       0.1, // 10 chat turns = 1 main turn
		tickAccumulator: 0.0,
		chatTurns:       0,

		// Internal state
		focused:      false,
		location:     "",
		sessionID:    "",
		sessionStart: time.Time{},
	}
}

// ==========================================================================
// Story 3.3: Chat Enter/Exit Logic
// ==========================================================================

// Enter activates the chat overlay with the given participants.
// Story 3.3 AC1, AC2: Initialize chat state and set participants.
func (m *ChatOverlayModel) Enter(initiator string, participants []ChatParticipant, location string) {
	m.active = true
	m.initiator = initiator
	m.participants = participants
	m.messages = []*ChatMessage{}
	m.chatTurns = 0
	m.tickAccumulator = 0.0
	m.location = location
	m.sessionID = fmt.Sprintf("chat_%d", time.Now().Unix())
	m.sessionStart = time.Now()
	m.focused = true
	m.inputField.Focus()
	m.inputField.SetValue("")

	// Add system message announcing chat start
	systemMsg := NewChatMessage(
		fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		"system",
		fmt.Sprintf("聊天室已開始 - 地點: %s", location),
		ChatMessageSystem,
	)
	m.messages = append(m.messages, systemMsg)
	m.updateViewport()
}

// Exit deactivates the chat overlay and returns a session summary.
// Story 3.3 AC4, AC5: Generate summary and save session.
func (m *ChatOverlayModel) Exit() *ChatSession {
	if !m.active {
		return nil
	}

	// Convert []*ChatMessage to []ChatMessage for session
	messages := make([]ChatMessage, len(m.messages))
	for i, msg := range m.messages {
		if msg != nil {
			messages[i] = *msg
		}
	}

	session := &ChatSession{
		SessionID:    m.sessionID,
		Initiator:    m.initiator,
		Participants: m.participants,
		Messages:     messages,
		StartTime:    m.sessionStart,
		EndTime:      time.Now(),
		Summary:      nil, // Summary generation will be implemented in Story 5.4
		Location:     m.location,
		TurnsSpent:   m.chatTurns,
	}

	// Reset state
	m.active = false
	m.focused = false
	m.inputField.Blur()
	m.inputField.SetValue("")

	return session
}

// AddMessage adds a new message to the chat history.
// Automatically updates viewport to show the latest message.
func (m *ChatOverlayModel) AddMessage(msg *ChatMessage) {
	m.messages = append(m.messages, msg)
	m.updateViewport()
}

// GetChatTurns returns the number of chat turns taken.
func (m *ChatOverlayModel) GetChatTurns() int {
	return m.chatTurns
}

// GetTickAccumulator returns the current tick accumulator value.
func (m *ChatOverlayModel) GetTickAccumulator() float64 {
	return m.tickAccumulator
}

// ==========================================================================
// Story 3-4 AC3: Participant Management Logic
// ==========================================================================

// AddParticipant adds a new participant to the chat.
// Prevents duplicate additions and triggers UI update.
func (m *ChatOverlayModel) AddParticipant(participant ChatParticipant) {
	// Check for duplicates
	for i, p := range m.participants {
		if p.ID == participant.ID {
			// If already exists, reactivate if inactive
			if !p.IsActive {
				m.participants[i].IsActive = true
			}
			return
		}
	}

	// Add new participant
	participant.IsActive = true
	m.participants = append(m.participants, participant)
}

// RemoveParticipant marks a participant as inactive.
// Preserves history by setting IsActive=false instead of deleting.
func (m *ChatOverlayModel) RemoveParticipant(participantID string) {
	for i, p := range m.participants {
		if p.ID == participantID {
			m.participants[i].IsActive = false
			return
		}
	}
}

// UpdateParticipantEmotion updates the emotion state of a participant.
// Triggers UI re-render for real-time emotion display.
func (m *ChatOverlayModel) UpdateParticipantEmotion(participantID string, emotion manager.EmotionState) {
	for i, p := range m.participants {
		if p.ID == participantID {
			m.participants[i].Emotion = emotion
			return
		}
	}
}

// GetParticipant retrieves a participant by ID.
func (m *ChatOverlayModel) GetParticipant(participantID string) *ChatParticipant {
	for i, p := range m.participants {
		if p.ID == participantID {
			return &m.participants[i]
		}
	}
	return nil
}

// GetActiveParticipants returns a list of currently active participants.
func (m *ChatOverlayModel) GetActiveParticipants() []ChatParticipant {
	active := []ChatParticipant{}
	for _, p := range m.participants {
		if p.IsActive {
			active = append(active, p)
		}
	}
	return active
}

// GetParticipantCount returns the count of active and total participants.
func (m *ChatOverlayModel) GetParticipantCount() (active, total int) {
	total = len(m.participants)
	for _, p := range m.participants {
		if p.IsActive {
			active++
		}
	}
	return active, total
}

// ==========================================================================
// Story 3-4 AC2 & AC4: Participant UI Rendering
// ==========================================================================

// renderParticipantsList renders the participants panel with emotion states.
func (m ChatOverlayModel) renderParticipantsList() string {
	var builder strings.Builder

	active, total := m.GetParticipantCount()
	header := fmt.Sprintf("參與者 (%d/%d)", active, total)
	builder.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 25))
	builder.WriteString("\n\n")

	// Sort participants: active first, then player first within each group
	sorted := make([]ChatParticipant, len(m.participants))
	copy(sorted, m.participants)
	sort.Slice(sorted, func(i, j int) bool {
		// Active participants come first
		if sorted[i].IsActive != sorted[j].IsActive {
			return sorted[i].IsActive
		}
		// Within same active status, player comes first
		if sorted[i].IsPlayer != sorted[j].IsPlayer {
			return sorted[i].IsPlayer
		}
		// Otherwise maintain order
		return i < j
	})

	// Render each participant
	for _, p := range sorted {
		builder.WriteString(m.renderParticipant(p))
		builder.WriteString("\n\n")
	}

	return builder.String()
}

// renderParticipant renders a single participant with emotion state.
// AC2: Display emotion state with color coding
// AC4: Display active/inactive status
func (m ChatOverlayModel) renderParticipant(p ChatParticipant) string {
	var builder strings.Builder

	// Name with player indicator
	nameStyle := lipgloss.NewStyle()
	if !p.IsActive {
		nameStyle = nameStyle.Foreground(lipgloss.Color("8")) // Gray for inactive
	}

	name := p.Name
	if p.IsPlayer {
		name = "[你] " + name
	}
	if !p.IsActive {
		name = "[已離開] " + name
	}

	builder.WriteString(nameStyle.Bold(true).Render(name))
	builder.WriteString("\n")

	// Emotion state (only for active participants)
	if p.IsActive {
		// Trust
		trustColor := getEmotionColor(p.Emotion.Trust)
		trustStyle := lipgloss.NewStyle().Foreground(trustColor)
		builder.WriteString(trustStyle.Render(fmt.Sprintf("  信任:%d", p.Emotion.Trust)))

		// Fear
		fearColor := getEmotionColor(p.Emotion.Fear)
		fearStyle := lipgloss.NewStyle().Foreground(fearColor)
		builder.WriteString(fearStyle.Render(fmt.Sprintf(" 恐懼:%d", p.Emotion.Fear)))
		builder.WriteString("\n")

		// Stress
		stressColor := getEmotionColor(p.Emotion.Stress)
		stressStyle := lipgloss.NewStyle().Foreground(stressColor)
		builder.WriteString(stressStyle.Render(fmt.Sprintf("  壓力:%d", p.Emotion.Stress)))

		// Relationship type with icon
		rel := manager.CalculateRelationship(p.Emotion)
		relIcon := getRelationshipIcon(rel)
		relStyle := lipgloss.NewStyle()
		builder.WriteString(relStyle.Render(fmt.Sprintf(" [%s] %s", rel.String(), relIcon)))
	}

	return builder.String()
}

// getEmotionColor returns color based on emotion value.
// AC2: Color coding for emotion levels
// High (>=70): Green (positive)
// Medium (30-69): Yellow (neutral)
// Low (<30): Red (negative)
func getEmotionColor(value int) lipgloss.Color {
	if value >= 70 {
		return lipgloss.Color("10") // Green
	} else if value >= 30 {
		return lipgloss.Color("11") // Yellow
	}
	return lipgloss.Color("9") // Red
}

// getRelationshipIcon returns icon for relationship type.
// AC2: Visual indicator for relationship
func getRelationshipIcon(rel manager.RelationshipType) string {
	switch rel {
	case manager.Friendly:
		return "●" // Full circle (friendly)
	case manager.Hostile:
		return "▲" // Triangle (hostile)
	case manager.Fearful:
		return "◆" // Diamond (fearful)
	default:
		return "○" // Empty circle (neutral)
	}
}

// ==========================================================================
// Story 3.1: Bubble Tea Interface Implementation
// ==========================================================================

// Init initializes the chat overlay model.
func (m ChatOverlayModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles Bubble Tea messages and updates the model state.
// Story 3.1 AC4: Handles Tab (exit), Enter (send), ESC (clear).
func (m ChatOverlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			// Story 3.1 AC4: Tab key exits chat
			if m.active {
				m.Exit()
				// In a real implementation, this would send a message to parent
				// to switch back to main game view
			}
			return m, nil

		case tea.KeyEnter:
			// Story 3.1 AC4: Enter sends message
			if m.active && m.focused {
				value := strings.TrimSpace(m.inputField.Value())
				if value != "" {
					// Create player message
					playerMsg := NewChatMessage(
						fmt.Sprintf("msg_%d", time.Now().UnixNano()),
						"player",
						value,
						ChatMessageNormal,
					)
					m.AddMessage(playerMsg)
					m.chatTurns++
					m.tickAccumulator += m.timeScale

					// Clear input
					m.inputField.SetValue("")

					// In a real implementation, this would trigger NPC response
					// generation via ChatProcessor (Story 4.2)
				}
			}
			return m, nil

		case tea.KeyEsc:
			// Story 3.1 AC4: ESC clears input or exits
			if m.focused && m.inputField.Value() != "" {
				m.inputField.SetValue("")
			} else if m.active {
				m.Exit()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		// Story 3.1 AC4: Handle window resize
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSize()
		return m, nil
	}

	// Update child components
	if m.focused {
		m.inputField, tiCmd = m.inputField.Update(msg)
	}
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

// View renders the chat overlay UI.
// Story 3.1 AC5: Renders participants list, messages, and input field.
func (m ChatOverlayModel) View() string {
	if !m.active {
		return ""
	}

	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1)

	inputBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	// Story 3.1 AC5: Render title
	title := titleStyle.Render(fmt.Sprintf("聊天室 - %s", m.location))

	// Story 3.1 AC5: Render participants list
	participantsView := m.renderParticipantsList()

	// Story 3.1 AC5: Render messages viewport
	messagesView := m.viewport.View()

	// Story 3.1 AC5: Render input field
	inputView := inputBorderStyle.Render(m.inputField.View())

	// Render status bar with time control info
	statusText := fmt.Sprintf(
		"回合: %d | 時間流速: %.1fx | 累積刻度: %.2f | 參與者: %d",
		m.chatTurns,
		m.timeScale,
		m.tickAccumulator,
		len(m.participants),
	)
	statusView := statusStyle.Render(statusText)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		participantsView,
		"", // Spacer
		messagesView,
		"", // Spacer
		inputView,
		statusView,
	)

	// Add outer border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		MaxWidth(m.width).
		MaxHeight(m.height)

	return containerStyle.Render(content)
}

// IsActive returns whether the chat overlay is active.
func (m ChatOverlayModel) IsActive() bool {
	return m.active
}

// SetActive sets the active state of the chat overlay.
func (m *ChatOverlayModel) SetActive(active bool) {
	m.active = active
}

// ==========================================================================
// Story 3.1: Helper Methods for Viewport and Message Rendering
// ==========================================================================

// updateViewport updates the viewport content with all messages.
// Story 3.1 AC4: Viewport auto-scrolls to latest message.
func (m *ChatOverlayModel) updateViewport() {
	if len(m.messages) == 0 {
		m.viewport.SetContent("")
		return
	}

	// Render all messages
	var lines []string
	for _, msg := range m.messages {
		if msg != nil {
			lines = append(lines, m.renderMessage(msg))
		}
	}

	content := strings.Join(lines, "\n\n")
	m.viewport.SetContent(content)

	// Auto-scroll to bottom
	m.viewport.GotoBottom()
}

// renderMessage renders a single chat message with appropriate styling.
// Story 3.1 AC5: Different styles for player/NPC/system messages.
func (m *ChatOverlayModel) renderMessage(msg *ChatMessage) string {
	var style lipgloss.Style
	var prefix string

	switch msg.Type {
	case ChatMessageSystem:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		prefix = "[系統]"

	case ChatMessageNormal:
		if msg.Speaker == "player" {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")). // Blue for player
				Bold(true)
			prefix = "[你]"
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")) // Pink for NPCs
			prefix = fmt.Sprintf("[%s]", msg.Speaker)
		}

	case ChatMessageWhisper:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("177")). // Purple for whispers
			Italic(true)
		prefix = fmt.Sprintf("[私訊:%s]", msg.Speaker)

	case ChatMessageThought:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")). // Gray for thoughts
			Italic(true)
		prefix = "(內心)"

	case ChatMessageAction:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange for actions
			Italic(true)
		prefix = "*"

	default:
		style = lipgloss.NewStyle()
		prefix = ""
	}

	// Add timestamp
	timestamp := msg.Timestamp.Format("15:04")

	// Format: [HH:MM] [Speaker] Content
	formatted := fmt.Sprintf("[%s] %s %s", timestamp, prefix, msg.Content)

	// Add flag indicators if present
	if len(msg.Flags) > 0 {
		var flagStrs []string
		for _, flag := range msg.Flags {
			flagStrs = append(flagStrs, getFlagIndicator(flag))
		}
		formatted += " " + strings.Join(flagStrs, " ")
	}

	return style.Render(formatted)
}

// getFlagIndicator returns a visual indicator for a chat flag.
func getFlagIndicator(flag ChatFlag) string {
	switch flag {
	case ChatFlagHallucination:
		return "[幻覺]"
	case ChatFlagHostile:
		return "[敵意]"
	case ChatFlagRevelation:
		return "[揭露]"
	case ChatFlagPersuasion:
		return "[說服]"
	case ChatFlagLie:
		return "[謊言]"
	case ChatFlagContradiction:
		return "[矛盾]"
	default:
		return "[?]"
	}
}

// updateViewportSize adjusts viewport dimensions based on window size.
func (m *ChatOverlayModel) updateViewportSize() {
	// Reserve space for title, participants, input, status, borders
	// Approximate calculation:
	// - Title: 2 lines
	// - Participants: variable (estimate 5 lines for 2-3 participants)
	// - Input: 3 lines
	// - Status: 2 lines
	// - Borders/padding: 6 lines
	// - Spacers: 2 lines
	// Total reserved: ~20 lines
	reservedHeight := 20
	viewportHeight := m.height - reservedHeight
	if viewportHeight < 5 {
		viewportHeight = 5 // Minimum height
	}

	// Width with padding
	viewportWidth := m.width - 10 // Account for borders and padding
	if viewportWidth < 40 {
		viewportWidth = 40 // Minimum width
	}

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
	m.inputField.Width = viewportWidth - 4
}

// ==========================================================================
// ChatSession Type Definition
// ==========================================================================

// ChatSession represents a complete chat session.
// Story 5.6 AC1: ChatSession 包含完整對話記錄
type ChatSession struct {
	SessionID    string            `json:"session_id"`    // Unique session ID
	Initiator    string            `json:"initiator"`     // Who initiated ("player" or NPC ID)
	Participants []ChatParticipant `json:"participants"`  // Participants in this session
	Messages     []ChatMessage     `json:"messages"`      // All messages in this session
	StartTime    time.Time         `json:"start_time"`    // When the session started
	EndTime      time.Time         `json:"end_time"`      // When the session ended
	Summary      interface{}       `json:"summary"`       // Session summary (ChatSummary - will be defined in Story 5.4)
	Location     string            `json:"location"`      // Where the chat took place
	TurnsSpent   int               `json:"turns_spent"`   // How many chat turns were spent
}

// ==========================================================================
// Story 3-6: Additional Helper Methods for Input Handling
// ==========================================================================

// sendMessage handles sending a player message via input field.
// AC1: 玩家輸入訊息後按 Enter 發送
// This is a private method called from Update()
func (m *ChatOverlayModel) sendMessage() {
	content := strings.TrimSpace(m.inputField.Value())
	if content == "" {
		return
	}

	// Create player message
	msg := NewChatMessage(
		fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		"player",
		content,
		ChatMessageNormal,
	)

	// Add message to list
	m.AddMessage(msg)
	m.chatTurns++

	// Clear input field
	m.inputField.SetValue("")
}

// GetMessages returns the current message list.
func (m *ChatOverlayModel) GetMessages() []*ChatMessage {
	return m.messages
}

// AddSystemMessage adds a system message to the chat.
func (m *ChatOverlayModel) AddSystemMessage(content string) {
	msg := NewChatMessage(
		fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		"system",
		content,
		ChatMessageSystem,
	)
	m.AddMessage(msg)
}

// AddNPCMessage adds an NPC message to the chat with optional flags.
func (m *ChatOverlayModel) AddNPCMessage(npcID, content string, flags []ChatFlag) {
	msg := NewChatMessage(
		fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		npcID,
		content,
		ChatMessageNormal,
	)
	msg.Flags = flags
	m.AddMessage(msg)
}

// updateViewportContent updates the viewport with the current message list.
// AC5: viewport 自動捲動到最新訊息 (called by updateViewport)
func (m *ChatOverlayModel) updateViewportContent() {
	// This functionality is already handled by updateViewport()
	m.updateViewport()
}

// getParticipantName retrieves the display name of a participant by ID.
// Returns the ID itself if participant not found.
func (m *ChatOverlayModel) getParticipantName(participantID string) string {
	for _, p := range m.participants {
		if p.ID == participantID {
			return p.Name
		}
	}
	return participantID
}
