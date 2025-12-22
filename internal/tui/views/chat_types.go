package views

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/chat"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ChatFlag is an alias to chat.ChatFlag to avoid import cycles.
// All ChatFlag functionality is defined in the chat package.
type ChatFlag = chat.ChatFlag

// Re-export ChatFlag constants for backward compatibility
const (
	ChatFlagHallucination = chat.ChatFlagHallucination
	ChatFlagHostile       = chat.ChatFlagHostile
	ChatFlagRevelation    = chat.ChatFlagRevelation
	ChatFlagPersuasion    = chat.ChatFlagPersuasion
	ChatFlagLie           = chat.ChatFlagLie
	ChatFlagContradiction = chat.ChatFlagContradiction
)

// ParseChatFlag converts a string to ChatFlag.
// Re-exported from chat package for backward compatibility.
// Returns (flag, true) if valid, or (ChatFlagHallucination, false) if invalid.
func ParseChatFlag(s string) (ChatFlag, error) {
	flag, ok := chat.ParseChatFlag(s)
	if !ok {
		return ChatFlagHallucination, fmt.Errorf("invalid chat flag: %s", s)
	}
	return flag, nil
}

// ChatMessage represents a single message in a chat session.
// It contains all necessary information for display, processing, and semantic analysis.
type ChatMessage struct {
	ID             string                           `json:"id"`              // Unique message identifier
	Speaker        string                           `json:"speaker"`         // Player ID or NPC ID
	Content        string                           `json:"content"`         // Message text content
	Timestamp      time.Time                        `json:"timestamp"`       // When the message was sent
	Type           ChatMessageType                  `json:"type"`            // Message type (normal, system, etc.)
	Flags          []ChatFlag                       `json:"flags"`           // Semantic flags detected by JudgeAgent
	EmotionEffects map[string]*manager.EmotionDelta `json:"emotion_effects"` // NPC ID -> Emotion change
}

// ChatMessageType defines different types of chat messages.
// Each type may be rendered differently in the UI.
type ChatMessageType int

const (
	// ChatMessageNormal represents a standard chat message from player or NPC
	ChatMessageNormal ChatMessageType = iota

	// ChatMessageSystem represents a system-generated message (announcements, events, etc.)
	ChatMessageSystem

	// ChatMessageWhisper represents a private message to/from specific participant
	ChatMessageWhisper

	// ChatMessageThought represents an internal thought or observation
	ChatMessageThought

	// ChatMessageAction represents an action performed by a participant
	ChatMessageAction
)

// String returns the string representation of ChatMessageType.
func (t ChatMessageType) String() string {
	switch t {
	case ChatMessageNormal:
		return "normal"
	case ChatMessageSystem:
		return "system"
	case ChatMessageWhisper:
		return "whisper"
	case ChatMessageThought:
		return "thought"
	case ChatMessageAction:
		return "action"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ParseChatMessageType converts a string to ChatMessageType.
// It performs case-insensitive matching.
func ParseChatMessageType(s string) (ChatMessageType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "normal":
		return ChatMessageNormal, nil
	case "system":
		return ChatMessageSystem, nil
	case "whisper":
		return ChatMessageWhisper, nil
	case "thought":
		return ChatMessageThought, nil
	case "action":
		return ChatMessageAction, nil
	default:
		return ChatMessageNormal, fmt.Errorf("invalid chat message type: %s", s)
	}
}

// MarshalJSON implements json.Marshaler for ChatMessageType.
func (t ChatMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler for ChatMessageType.
func (t *ChatMessageType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseChatMessageType(s)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// HasFlag checks if a ChatMessage has a specific flag.
func (m *ChatMessage) HasFlag(flag ChatFlag) bool {
	for _, f := range m.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// AddFlag adds a flag to the message if it doesn't already exist.
func (m *ChatMessage) AddFlag(flag ChatFlag) {
	if !m.HasFlag(flag) {
		m.Flags = append(m.Flags, flag)
	}
}

// RemoveFlag removes a flag from the message if it exists.
func (m *ChatMessage) RemoveFlag(flag ChatFlag) {
	newFlags := make([]ChatFlag, 0, len(m.Flags))
	for _, f := range m.Flags {
		if f != flag {
			newFlags = append(newFlags, f)
		}
	}
	m.Flags = newFlags
}

// GetEmotionEffect returns the emotion effect for a specific NPC, or nil if not found.
func (m *ChatMessage) GetEmotionEffect(npcID string) *manager.EmotionDelta {
	if m.EmotionEffects == nil {
		return nil
	}
	return m.EmotionEffects[npcID]
}

// SetEmotionEffect sets the emotion effect for a specific NPC.
func (m *ChatMessage) SetEmotionEffect(npcID string, delta *manager.EmotionDelta) {
	if m.EmotionEffects == nil {
		m.EmotionEffects = make(map[string]*manager.EmotionDelta)
	}
	m.EmotionEffects[npcID] = delta
}

// NewChatMessage creates a new ChatMessage with the given parameters.
// It automatically sets the timestamp to the current time and initializes empty slices/maps.
func NewChatMessage(id, speaker, content string, msgType ChatMessageType) *ChatMessage {
	return &ChatMessage{
		ID:             id,
		Speaker:        speaker,
		Content:        content,
		Timestamp:      time.Now(),
		Type:           msgType,
		Flags:          []ChatFlag{},
		EmotionEffects: make(map[string]*manager.EmotionDelta),
	}
}

// Copy creates a deep copy of the ChatMessage.
func (m *ChatMessage) Copy() *ChatMessage {
	// Copy flags
	flags := make([]ChatFlag, len(m.Flags))
	copy(flags, m.Flags)

	// Copy emotion effects
	emotionEffects := make(map[string]*manager.EmotionDelta)
	for k, v := range m.EmotionEffects {
		if v != nil {
			deltaCopy := *v
			emotionEffects[k] = &deltaCopy
		}
	}

	return &ChatMessage{
		ID:             m.ID,
		Speaker:        m.Speaker,
		Content:        m.Content,
		Timestamp:      m.Timestamp,
		Type:           m.Type,
		Flags:          flags,
		EmotionEffects: emotionEffects,
	}
}

// ==========================================================================
// ChatContext: Context information for chat initialization
// ==========================================================================

// ChatContext provides all necessary context for initializing a chat session.
// Used by EnterWithContext() for full NPC/Knowledge integration.
type ChatContext struct {
	CurrentRoom   string              // Current room/location name
	PlayerID      string              // Player identifier
	NPCManager    interface{}         // *npc.Manager (avoid import cycle)
	UpdateManager interface{}         // *knowledge.UpdateManager (avoid import cycle)
	RoomOccupants map[string][]string // Room -> Entity IDs mapping
}

// ==========================================================================
// ChatSummary: Summary generated after chat session
// ==========================================================================

// ChatSummary represents a summary of a chat session.
// Story 5.4 AC2, AC3: ChatSummary contains all structured summary information
type ChatSummary struct {
	// AC2: Main conversation elements
	MainTopics      []string          `json:"main_topics"`      // Main topics discussed (2-3 key topics)
	KeyDecisions    []string          `json:"key_decisions"`    // Important decisions made by player or NPCs
	RelationChanges map[string]string `json:"relation_changes"` // NPC ID -> relationship change description

	// AC3: Detailed information and narrative impact
	FactsShared     []string `json:"facts_shared"`     // Facts/information shared during conversation
	Flags           []string `json:"flags"`            // Special markers (lies, contradictions, revelations, etc.)
	NarrativeImpact string   `json:"narrative_impact"` // Overall impact on main narrative (200-400 chars)

	// AC4: Emotional and thematic elements
	EmotionChanges   map[string]string `json:"emotion_changes,omitempty"`   // NPC ID -> emotion change summary
	UnresolvedIssues []string          `json:"unresolved_issues,omitempty"` // Unresolved tensions or questions
}

// ==========================================================================
// Story 5.1: Chat Time Flow Control - ChatConfig
// ==========================================================================

// ChatConfig provides configuration for chat overlay time flow and behavior.
// Story 5.1 AC2, AC5: Configuration for time scale and chat turns per beat.
type ChatConfig struct {
	TimeScale        float64 `json:"time_scale"`          // Time scaling factor (default 0.1 = 10% of main timeline)
	ChatTurnsPerBeat int     `json:"chat_turns_per_beat"` // Number of chat turns per main beat (default 10)
	AllowInterrupts  bool    `json:"allow_interrupts"`    // Whether events can interrupt chat (Story 5.3)
}

// DefaultChatConfig returns a ChatConfig with default values.
// AC2: chatTurnsPerBeat = 10 (default)
// AC5: timeScale = 0.1 (default)
func DefaultChatConfig() ChatConfig {
	return ChatConfig{
		TimeScale:        0.1,  // 10 chat turns = 1 main beat
		ChatTurnsPerBeat: 10,   // Default cadence
		AllowInterrupts:  true, // Story 5.3 AC3: Enable interrupts by default
	}
}

// ==========================================================================
// Story 5.2: Pending Events Mechanism - Event Data Structures
// ==========================================================================

// EventType defines the type of a pending chat event.
// AC1: PendingChatEvent 包含 Type 欄位
type EventType string

const (
	// EventTypeCombat represents combat-related events (attacks, enemy encounters)
	EventTypeCombat EventType = "combat"

	// EventTypeDiscovery represents discovery events (clues, items, secrets)
	EventTypeDiscovery EventType = "discovery"

	// EventTypeEnvironmental represents environmental changes (weather, time, location)
	EventTypeEnvironmental EventType = "environmental"

	// EventTypeSocial represents social events (NPC interactions, relationships)
	EventTypeSocial EventType = "social"

	// EventTypeDanger represents immediate danger events (traps, hazards)
	EventTypeDanger EventType = "danger"

	// EventTypeNPCAction represents NPC-initiated actions
	EventTypeNPCAction EventType = "npc_action"

	// EventTypeRevelation represents narrative revelations (plot twists, revelations)
	EventTypeRevelation EventType = "revelation"
)

// String returns the string representation of EventType.
func (e EventType) String() string {
	return string(e)
}

// EventSeverity defines the severity/importance level of an event.
// AC1: PendingChatEvent 包含 Severity 欄位
type EventSeverity string

const (
	// SeverityLow represents low-priority background events
	SeverityLow EventSeverity = "low"

	// SeverityMedium represents noteworthy events
	SeverityMedium EventSeverity = "medium"

	// SeverityHigh represents important events
	SeverityHigh EventSeverity = "high"

	// SeverityCritical represents critical events requiring immediate attention
	SeverityCritical EventSeverity = "critical"
)

// String returns the string representation of EventSeverity.
func (e EventSeverity) String() string {
	return string(e)
}

// EventEffects represents the effects of a pending chat event on game state.
// AC1: PendingChatEvent 包含 Effects 欄位
type EventEffects struct {
	HPDelta       int               `json:"hp_delta,omitempty"`        // HP change (positive or negative)
	SANDelta      int               `json:"san_delta,omitempty"`       // Sanity change (positive or negative)
	StatusChanges map[string]bool   `json:"status_changes,omitempty"`  // Status effects to add/remove
	ItemsGained   []string          `json:"items_gained,omitempty"`    // Items gained from event
	ItemsLost     []string          `json:"items_lost,omitempty"`      // Items lost in event
	CluesRevealed []string          `json:"clues_revealed,omitempty"`  // Clues revealed by event
}

// PendingChatEvent represents an event scheduled to trigger during a chat session.
// AC1: PendingChatEvent 資料結構定義
// Story 5.2: Complete event data structure with all required fields
type PendingChatEvent struct {
	ID             string        `json:"id"`                         // Unique event identifier
	Type           EventType     `json:"type"`                       // Event type (combat, discovery, etc.)
	Description    string        `json:"description"`                // Human-readable event description
	TriggerBeat    int           `json:"trigger_beat"`               // Main timeline beat when event triggers
	IsInterrupting bool          `json:"is_interrupting"`            // Whether event interrupts chat
	Severity       EventSeverity `json:"severity"`                   // Event severity/importance
	RequiredAction string        `json:"required_action,omitempty"`  // Action required from player (if any)
	Effects        *EventEffects `json:"effects,omitempty"`          // Game state effects
}
