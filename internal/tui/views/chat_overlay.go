package views

import (
	"context"
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
// Story 4.6: Emotion update integration
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
	// Story 5.1: Chat Time Flow Control
	timeScale        float64 // Time flow rate (0.1 = 10 chat turns per main turn)
	tickAccumulator  float64 // Accumulated time ticks
	chatTurns        int     // Number of chat turns taken
	chatTurnsPerBeat int     // Number of chat turns per main beat (default 10)
	startBeat        int     // The main timeline beat when chat started

	// Internal state
	focused           bool      // Whether input field is focused
	location          string    // Current location where chat is happening
	sessionID         string    // Unique session identifier
	sessionStart      time.Time // When the session started
	lastPlayerMessage string    // Last player message sent (for interaction recording)

	// Story 4.6: NPC integration for emotion updates
	npcManager *manager.NPCManager // NPC manager for emotion updates

	// Story 5.4: Summary generation
	summaryGenerator SummaryGenerator // LLM-based summary generator (optional)

	// Story 5.2: Pending events mechanism
	pendingEvents    []*PendingChatEvent // Events that may interrupt chat (AC2)
	backgroundEvents []*PendingChatEvent // Background events that don't interrupt (AC3)

	// Story 5.3: Event Interruption Logic
	interruptPending bool                // Whether there is a pending interruption (AC1)
	interruptReason  string              // Reason for the interruption (AC1)
	interruptEvent   *PendingChatEvent   // The event that triggered the interruption (AC1)
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
		// Story 5.1: Chat Time Flow Control
		timeScale:        0.1, // 10 chat turns = 1 main turn
		tickAccumulator:  0.0,
		chatTurns:        0,
		chatTurnsPerBeat: 10, // Default: 10 chat turns per main beat
		startBeat:        0,  // Will be set when entering chat

		// Internal state
		focused:      false,
		location:     "",
		sessionID:    "",
		sessionStart: time.Time{},

		// Story 5.2: Initialize pending events
		pendingEvents:    []*PendingChatEvent{},
		backgroundEvents: []*PendingChatEvent{},
	}
}

// ==========================================================================
// Story 3.3: Chat Enter/Exit Logic
// ==========================================================================

// Enter activates the chat overlay with the given participants.
// Story 3.3 AC1, AC2: Initialize chat state and set participants.
// Story 5.2 AC2, AC3: Initialize pending events queues
// Note: startBeat should be set separately via SetStartBeat() after calling Enter()
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
	m.startBeat = 0 // Default to 0, should be set via SetStartBeat()
	m.focused = true
	m.inputField.Focus()
	m.inputField.SetValue("")

	// Story 5.2 AC2, AC3: Reset pending events
	m.pendingEvents = []*PendingChatEvent{}
	m.backgroundEvents = []*PendingChatEvent{}

	// Story 5.3 AC1: Reset interrupt state
	m.interruptPending = false
	m.interruptReason = ""
	m.interruptEvent = nil

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

// SetStartBeat sets the beat when the chat started.
// This should be called right after Enter() to track when the chat began in game time.
func (m *ChatOverlayModel) SetStartBeat(beat int) {
	m.startBeat = beat
}

// Exit deactivates the chat overlay and returns a session summary.
// Story 3.3 AC4, AC5: Generate summary and save session.
// Story 5.4 AC1: Generate chat summary using LLM.
// Story 5.6 AC1: Populate ChatSession with all required fields.
// endBeat parameter is the current beat when chat ends.
func (m *ChatOverlayModel) Exit() *ChatSession {
	if !m.active {
		return nil
	}

	// Story 5.4 AC1: Generate summary before exiting
	var summary *ChatSummary
	ctx := context.Background()
	summaryResult, err := m.generateSummary(ctx, m.summaryGenerator)
	if err != nil {
		// Error already logged in generateSummary, just use nil summary
		summary = nil
	} else {
		summary = summaryResult
	}

	// Calculate end beat based on chat turns
	// Each chatTurnsPerBeat chat turns equals 1 main beat
	beatsSpent := m.chatTurns / m.chatTurnsPerBeat
	endBeat := m.startBeat + beatsSpent

	// Extract participant IDs
	participantIDs := make([]string, len(m.participants))
	for i, p := range m.participants {
		participantIDs[i] = p.ID
	}

	session := &ChatSession{
		// Story 5.6 AC1: Core fields
		ID:              m.sessionID,
		Participants:    participantIDs,
		Messages:        m.messages, // Already []*ChatMessage
		StartBeat:       m.startBeat,
		EndBeat:         endBeat,
		Summary:         summary, // Story 5.4: Generated summary
		Interrupted:     false,
		InterruptReason: "",
		CreatedAt:       m.sessionStart,

		// Legacy fields for backward compatibility
		SessionID:          m.sessionID,
		Initiator:          m.initiator,
		ParticipantDetails: m.participants,
		StartTime:          m.sessionStart,
		EndTime:            time.Now(),
		Location:           m.location,
		TurnsSpent:         m.chatTurns,
	}

	// Reset state
	m.active = false
	m.focused = false
	m.inputField.Blur()
	m.inputField.SetValue("")

	return session
}

// ExitWithInterruption exits chat with interruption information.
// Story 5.6 AC1: Support interrupted sessions with reason.
func (m *ChatOverlayModel) ExitWithInterruption(reason string) *ChatSession {
	session := m.Exit()
	if session != nil {
		session.Interrupted = true
		session.InterruptReason = reason
	}
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
			// Story 5.1: Integrated with time flow control
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

					// Story 5.1: Use incrementChatTurns for time flow control
					m.incrementChatTurns()

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
// Includes ID, Participants, Messages, StartBeat, EndBeat, Summary, Interrupted, InterruptReason
type ChatSession struct {
	ID              string         `json:"id"`                         // Unique session ID
	Participants    []string       `json:"participants"`               // Participant IDs (player, npc_001, etc.)
	Messages        []*ChatMessage `json:"messages"`                   // Complete message list
	StartBeat       int            `json:"start_beat"`                 // Start beat number
	EndBeat         int            `json:"end_beat"`                   // End beat number
	Summary         *ChatSummary   `json:"summary,omitempty"`          // Summary (if generated)
	Interrupted     bool           `json:"interrupted"`                // Whether interrupted
	InterruptReason string         `json:"interrupt_reason,omitempty"` // Interrupt reason (if interrupted)
	CreatedAt       time.Time      `json:"created_at"`                 // Creation timestamp

	// Legacy fields for backward compatibility
	SessionID          string            `json:"session_id,omitempty"`          // Deprecated: use ID
	Initiator          string            `json:"initiator,omitempty"`           // Who initiated ("player" or NPC ID)
	ParticipantDetails []ChatParticipant `json:"participant_details,omitempty"` // Full participant info
	StartTime          time.Time         `json:"start_time,omitempty"`          // Deprecated: use CreatedAt
	EndTime            time.Time         `json:"end_time,omitempty"`            // Deprecated: calculated from EndBeat
	Location           string            `json:"location,omitempty"`            // Where the chat took place
	TurnsSpent         int               `json:"turns_spent,omitempty"`         // How many chat turns were spent
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

// ==========================================================================
// Story 5.1: Chat Time Flow Control - Time Management Methods
// ==========================================================================

// incrementChatTurns increments the chat turn counter and checks for main timeline triggers.
// AC1: chatTurns 追蹤聊天回合數
// AC3: tickAccumulator 累積時間流
// AC4: 每 10 聊天回合觸發 1 主時間回合事件
// Returns true if a main timeline event was triggered.
func (m *ChatOverlayModel) incrementChatTurns() bool {
	// AC1: Increment chat turns counter
	m.chatTurns++

	// AC3: Accumulate time flow with time scale
	m.tickAccumulator += m.timeScale

	// AC4: Check if we should trigger a main timeline event
	shouldTrigger := false

	// Condition 1: chatTurns reached chatTurnsPerBeat
	if m.chatTurns%m.chatTurnsPerBeat == 0 {
		shouldTrigger = true
		// When chatTurns aligns with beat, reset accumulator to stay synchronized
		m.tickAccumulator = 0.0
	}

	// Condition 2: tickAccumulator reached 1.0 (if not already triggered by Condition 1)
	if !shouldTrigger && m.tickAccumulator >= 1.0 {
		shouldTrigger = true
		m.tickAccumulator -= 1.0 // Reset accumulator after triggering
	}

	// Trigger main timeline event if conditions met
	if shouldTrigger {
		m.triggerMainTimelineEvent()
	}

	return shouldTrigger
}

// checkMainTimelineTick checks if a main timeline tick should be triggered.
// AC4: 檢查是否應觸發主時間回合
// Returns true if conditions for main timeline tick are met.
func (m *ChatOverlayModel) checkMainTimelineTick() bool {
	// Condition 1: chatTurns reached a multiple of chatTurnsPerBeat
	if m.chatTurns%m.chatTurnsPerBeat == 0 && m.chatTurns > 0 {
		return true
	}

	// Condition 2: tickAccumulator accumulated enough
	if m.tickAccumulator >= 1.0 {
		return true
	}

	return false
}

// triggerMainTimelineEvent handles triggering a main timeline beat event.
// AC4: 觸發主時間回合事件
// Story 5.2 AC4: Check and trigger pending events when time advances
func (m *ChatOverlayModel) triggerMainTimelineEvent() {
	// Calculate current main beat
	mainBeat := m.startBeat + (m.chatTurns / m.chatTurnsPerBeat)

	// Add system message to notify player
	// AC5: 向玩家明確傳達時間流較慢
	systemMsg := fmt.Sprintf("時間流逝... (主時間回合 %d)", mainBeat)
	m.AddSystemMessage(systemMsg)

	// Story 5.2 AC4: Check for pending events at current beat
	m.checkPendingEvents(mainBeat)

	// Story 5.3: Handle any interruptions triggered by events
	m.handleInterruption()
}

// ApplyConfig applies a ChatConfig to the chat overlay model.
// AC2, AC5: 可根據難度或場景調整時間流配置
func (m *ChatOverlayModel) ApplyConfig(config ChatConfig) {
	m.timeScale = config.TimeScale
	m.chatTurnsPerBeat = config.ChatTurnsPerBeat
	// AllowInterrupts will be used in Story 5.3
}

// GetTimeFlowInfo returns current time flow information for display.
// AC5: timeScale = 0.1 視覺化指示
func (m *ChatOverlayModel) GetTimeFlowInfo() (chatTurns int, timeScale float64, accumulator float64, turnsPerBeat int) {
	return m.chatTurns, m.timeScale, m.tickAccumulator, m.chatTurnsPerBeat
}

// ==========================================================================
// Story 5.4: Summary Generator Configuration
// ==========================================================================

// SetSummaryGenerator sets the summary generator for this chat overlay.
// This should be called during initialization to enable LLM-based summary generation.
func (m *ChatOverlayModel) SetSummaryGenerator(generator SummaryGenerator) {
	m.summaryGenerator = generator
}

// ==========================================================================
// Story 4.6: Emotion Update Integration
// ==========================================================================

// SetNPCManager sets the NPC manager for emotion updates.
// This should be called during initialization to enable emotion integration.
func (m *ChatOverlayModel) SetNPCManager(npcManager *manager.NPCManager) {
	m.npcManager = npcManager
}

// ==========================================================================
// Story 4.6: Emotion Update Integration - ProcessResult Application
// ==========================================================================

// ApplyProcessResult processes the result from ChatProcessor and applies all updates.
// This is a high-level wrapper that uses the existing HandleProcessResultDirect implementation.
// This method integrates all Epic 4 components.
//
// AC1: ProcessResult contains emotion changes for each NPC
// AC2: ChatOverlay receives ProcessResult and applies emotion changes
// AC3: Participant emotions are updated in real-time (UI display)
// AC4: Emotion changes are recorded to NPCInteraction history
func (m *ChatOverlayModel) ApplyProcessResult(result *ProcessResultStruct) {
	if result == nil {
		return
	}

	// Handle failed processing
	if !result.Success {
		errorMsg := "處理失敗"
		if result.Error != "" {
			errorMsg = fmt.Sprintf("處理失敗: %s", result.Error)
		}
		m.AddSystemMessage(errorMsg)
		return
	}

	// Convert NPCResponseInterface to the struct format expected by HandleProcessResultDirect
	npcResponses := make([]struct {
		NPCID   string
		Content string
		Flags   []ChatFlag
	}, len(result.NPCResponses))

	for i, resp := range result.NPCResponses {
		npcResponses[i] = struct {
			NPCID   string
			Content string
			Flags   []ChatFlag
		}{
			NPCID:   resp.NPCID,
			Content: resp.Content,
			Flags:   resp.Flags,
		}
	}

	// Use existing HandleProcessResultDirect implementation
	m.HandleProcessResultDirect(result.EmotionChanges, npcResponses, result.Success, result.Error)

	// Handle contradictions if detected
	if len(result.Contradictions) > 0 {
		m.handleContradictions(result.Contradictions)
	}
}

// handleContradictions processes contradiction results and adds system notifications.
// This method provides user feedback when contradictions are detected during chat processing.
func (m *ChatOverlayModel) handleContradictions(contradictions []ContradictionStruct) {
	for _, contradiction := range contradictions {
		// Add system message to notify about contradiction
		msg := fmt.Sprintf("矛盾偵測: %s 對於「%s」的說法與已知資訊「%s」不符",
			m.getParticipantName(contradiction.NPCID),
			contradiction.PlayerClaim,
			contradiction.NPCBelief,
		)
		m.AddSystemMessage(msg)
	}
}

// ProcessResultStruct is a struct version for ProcessResult from chat package.
// We use a struct instead of the interface to avoid import cycle with internal/chat.
type ProcessResultStruct struct {
	// NPCResponses contains all NPC responses generated (Story 4.4)
	NPCResponses []NPCResponseStruct

	// EmotionChanges maps NPCID to emotion delta applied (Story 4.6)
	EmotionChanges map[string]manager.EmotionDelta

	// Flags contains all chat flags detected by JudgeAgent (Story 4.1)
	Flags []ChatFlag

	// Contradictions contains all contradictions detected (Story 4.7)
	Contradictions []ContradictionStruct

	// Success indicates if processing completed successfully
	Success bool

	// Error contains error message if processing failed
	Error string
}

// NPCResponseStruct represents a single NPC's response to the player's message.
type NPCResponseStruct struct {
	// NPCID is the unique identifier of the responding NPC
	NPCID string

	// Content is the generated response text
	Content string

	// Emotion is the NPC's emotional state when responding
	Emotion manager.EmotionState

	// Flags are any special flags associated with this response
	Flags []ChatFlag

	// UsedFallback indicates if this response used a fallback template (Story 4.5)
	UsedFallback bool
}

// ContradictionStruct represents a detected contradiction.
type ContradictionStruct struct {
	NPCID          string
	PlayerClaim    string
	NPCBelief      string
	Severity       string
	SuggestedDelta manager.EmotionDelta
}

// ==========================================================================
// Story 5.2: Pending Events Mechanism - Event Queue Management
// ==========================================================================

// AddPendingEvent adds a pending event to the appropriate queue.
// AC2, AC3: Events are sorted by TriggerBeat and added to the correct queue
// based on IsInterrupting flag.
func (m *ChatOverlayModel) AddPendingEvent(event *PendingChatEvent) {
	if event == nil {
		return
	}

	// Add to appropriate queue based on IsInterrupting flag
	if event.IsInterrupting {
		// AC2: Add to pendingEvents queue
		m.pendingEvents = append(m.pendingEvents, event)
		// Sort by TriggerBeat (ascending order - earliest first)
		sort.Slice(m.pendingEvents, func(i, j int) bool {
			return m.pendingEvents[i].TriggerBeat < m.pendingEvents[j].TriggerBeat
		})
	} else {
		// AC3: Add to backgroundEvents queue
		m.backgroundEvents = append(m.backgroundEvents, event)
		// Sort by TriggerBeat (ascending order - earliest first)
		sort.Slice(m.backgroundEvents, func(i, j int) bool {
			return m.backgroundEvents[i].TriggerBeat < m.backgroundEvents[j].TriggerBeat
		})
	}
}

// checkPendingEvents checks for and triggers events at the current beat.
// AC4: 時間到達時觸發事件
// Returns the list of triggered events.
func (m *ChatOverlayModel) checkPendingEvents(currentBeat int) []*PendingChatEvent {
	triggeredEvents := []*PendingChatEvent{}

	// AC4: Check interrupting events (pendingEvents)
	for len(m.pendingEvents) > 0 {
		event := m.pendingEvents[0]

		// Check if event should trigger
		if event.TriggerBeat <= currentBeat {
			// Process the event
			m.handleEvent(event)
			triggeredEvents = append(triggeredEvents, event)

			// Remove from queue (shift array left)
			m.pendingEvents = m.pendingEvents[1:]
		} else {
			// Events are sorted, so if first event hasn't triggered, none will
			break
		}
	}

	// AC4: Check background events (backgroundEvents)
	for len(m.backgroundEvents) > 0 {
		event := m.backgroundEvents[0]

		// Check if event should trigger
		if event.TriggerBeat <= currentBeat {
			// Process background event (silently)
			m.handleBackgroundEvent(event)
			triggeredEvents = append(triggeredEvents, event)

			// Remove from queue (shift array left)
			m.backgroundEvents = m.backgroundEvents[1:]
		} else {
			// Events are sorted, so if first event hasn't triggered, none will
			break
		}
	}

	return triggeredEvents
}

// handleEvent processes a triggered interrupting event.
// AC4: 處理事件並從佇列移除
func (m *ChatOverlayModel) handleEvent(event *PendingChatEvent) {
	// 1. Apply event effects
	if event.Effects != nil {
		m.applyEventEffects(event.Effects)
	}

	// 2. Add system message to notify player
	systemMsg := fmt.Sprintf("[事件] %s", event.Description)
	m.AddSystemMessage(systemMsg)

	// 3. For critical/high severity events, add additional context
	if event.Severity == SeverityCritical || event.Severity == SeverityHigh {
		if event.RequiredAction != "" {
			actionMsg := fmt.Sprintf("需要行動: %s", event.RequiredAction)
			m.AddSystemMessage(actionMsg)
		}
	}

	// 4. Story 5.3 AC1: Set interrupt flags if this is an interrupting event
	if event.IsInterrupting {
		m.interruptPending = true
		m.interruptReason = event.Description
		m.interruptEvent = event
	}
}

// handleBackgroundEvent processes a triggered background event.
// AC3: 背景事件靜默發生，只記錄不顯示
func (m *ChatOverlayModel) handleBackgroundEvent(event *PendingChatEvent) {
	// Background events apply effects silently
	if event.Effects != nil {
		m.applyEventEffects(event.Effects)
	}

	// Background events may be mentioned in chat summary (Story 5.4)
	// but don't generate immediate system messages
}

// applyEventEffects applies the effects of an event to the game state.
// AC4: 應用事件效果
func (m *ChatOverlayModel) applyEventEffects(effects *EventEffects) {
	// Note: Actual game state modification requires integration with
	// the main game state manager. For now, we track what effects
	// would be applied.

	// HP/SAN changes would be applied through ChatContext.GameState
	// Items/Clues would be added to player inventory
	// Status changes would update player status effects

	// This is a placeholder for the actual implementation which
	// will require access to the game state through ChatContext

	// TODO: In future stories, integrate with:
	// - m.context.GameState.ModifyHP(effects.HPDelta)
	// - m.context.GameState.ModifySAN(effects.SANDelta)
	// - m.context.GameState.AddItems(effects.ItemsGained)
	// - m.context.GameState.RemoveItems(effects.ItemsLost)
	// - m.context.GameState.RevealClues(effects.CluesRevealed)
}

// GetPendingEvents returns a copy of the current pending events queue.
// Used for inspection and testing.
func (m *ChatOverlayModel) GetPendingEvents() []*PendingChatEvent {
	// Return a copy to prevent external modification
	events := make([]*PendingChatEvent, len(m.pendingEvents))
	copy(events, m.pendingEvents)
	return events
}

// GetBackgroundEvents returns a copy of the current background events queue.
// Used for inspection and testing.
func (m *ChatOverlayModel) GetBackgroundEvents() []*PendingChatEvent {
	// Return a copy to prevent external modification
	events := make([]*PendingChatEvent, len(m.backgroundEvents))
	copy(events, m.backgroundEvents)
	return events
}

// ClearPendingEvents removes all pending events.
// Used for testing or when events should be cancelled.
func (m *ChatOverlayModel) ClearPendingEvents() {
	m.pendingEvents = []*PendingChatEvent{}
	m.backgroundEvents = []*PendingChatEvent{}
}

// ==========================================================================
// Story 5.3: Event Interruption Logic
// ==========================================================================

// handleInterruption processes a pending chat interruption based on severity.
// Story 5.3 AC3, AC4: Handle interruptions according to config and severity
//
// Behavior:
//   - If AllowInterrupts is false, downgrades to notification only
//   - Critical: Forces immediate chat exit
//   - High: Displays urgent notification (suggests exit but doesn't force)
//   - Medium: Displays normal notification
//   - Low: Should not be interrupting
func (m *ChatOverlayModel) handleInterruption() {
	if !m.interruptPending || m.interruptEvent == nil {
		return
	}

	event := m.interruptEvent

	// AC3: Check if interrupts are allowed
	// Note: Config not yet integrated, using default behavior
	// TODO: Add config field to ChatOverlayModel when ChatConfig is integrated
	allowInterrupts := true // Default to true per AC3

	if !allowInterrupts {
		// Downgrade to notification only
		m.displayInterruptNotification(event, false)
		m.interruptPending = false
		return
	}

	// AC4: Handle based on severity
	switch event.Severity {
	case SeverityCritical:
		// Critical: Force immediate exit
		m.forceExit(event.Description)

	case SeverityHigh:
		// High: Display urgent notification
		// Note: Confirmation dialog not implemented in this version
		// Display notification and let player decide to continue or exit manually
		m.displayInterruptNotification(event, true)
		m.interruptPending = false

	case SeverityMedium:
		// Medium: Display normal notification, continue chat
		m.displayInterruptNotification(event, false)
		m.interruptPending = false

	default:
		// Low or other: Should not be marked as interrupting
		m.interruptPending = false
	}
}

// forceExit forces the chat to exit due to a Critical event.
// Story 5.3 AC4: Critical events immediately force chat exit
//
// This method:
// 1. Displays a forced exit message
// 2. Calls Exit() to generate summary
// 3. Marks the summary as interrupted
// 4. Deactivates the chat overlay
func (m *ChatOverlayModel) forceExit(reason string) {
	// 1. Display forced exit message
	exitMsg := fmt.Sprintf("⚠️ [緊急] %s - 聊天被迫中斷！", reason)
	m.AddSystemMessage(exitMsg)

	// 2. Generate summary (even if incomplete)
	session := m.Exit()

	// 3. Mark summary as interrupted
	if session != nil && session.Summary != nil {
		if session.Summary.NarrativeImpact != "" {
			session.Summary.NarrativeImpact = fmt.Sprintf(
				"對話被中斷：%s。%s",
				reason,
				session.Summary.NarrativeImpact,
			)
		} else {
			session.Summary.NarrativeImpact = fmt.Sprintf("對話被中斷：%s", reason)
		}
	}

	// 4. Deactivate chat overlay
	m.active = false
	m.interruptPending = false
	m.interruptReason = ""
	m.interruptEvent = nil

	// Note: In a full TUI implementation, this would return a tea.Cmd
	// to notify the main game loop of the forced exit
}

// displayInterruptNotification displays a highlighted notification for an interrupting event.
// Story 5.3 AC2: Display system messages with visual distinction
//
// Parameters:
//   - event: The event to display
//   - urgent: Whether to use urgent styling (for High severity)
func (m *ChatOverlayModel) displayInterruptNotification(event *PendingChatEvent, urgent bool) {
	// Determine icon and style based on severity
	var icon string
	var stylePrefix string

	if urgent {
		// High severity: Use urgent styling
		icon = "⚡"
		stylePrefix = "[緊急] "
	} else {
		// Medium severity or downgraded: Use normal styling
		switch event.Severity {
		case SeverityHigh:
			icon = "⚠️"
		case SeverityMedium:
			icon = "ℹ️"
		default:
			icon = "·"
		}
		stylePrefix = ""
	}

	// Construct highlighted message
	message := fmt.Sprintf("%s %s%s", icon, stylePrefix, event.Description)

	// Add required action if present
	if event.RequiredAction != "" && (event.Severity == SeverityCritical || event.Severity == SeverityHigh) {
		message += fmt.Sprintf(" - %s", event.RequiredAction)
	}

	// Add to system messages
	m.AddSystemMessage(message)

	// Note: In a full TUI implementation, this would use lipgloss styling
	// for colors, borders, and visual distinction
	// For now, we rely on the icon and prefix for distinction
}

