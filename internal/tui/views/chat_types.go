package views

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ChatMessage represents a single message in a chat session.
// It contains all necessary information for display, processing, and semantic analysis.
type ChatMessage struct {
	ID             string                       `json:"id"`              // Unique message identifier
	Speaker        string                       `json:"speaker"`         // Player ID or NPC ID
	Content        string                       `json:"content"`         // Message text content
	Timestamp      time.Time                    `json:"timestamp"`       // When the message was sent
	Type           ChatMessageType              `json:"type"`            // Message type (normal, system, etc.)
	Flags          []ChatFlag                   `json:"flags"`           // Semantic flags detected by JudgeAgent
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

// ChatFlag marks semantic properties detected by JudgeAgent during chat analysis.
// These flags indicate potentially important player behaviors that affect NPC responses.
type ChatFlag int

const (
	// ChatFlagHallucination indicates the player is claiming something that didn't happen
	ChatFlagHallucination ChatFlag = iota

	// ChatFlagHostile indicates aggressive or threatening language
	ChatFlagHostile

	// ChatFlagRevelation indicates sharing important information or secrets
	ChatFlagRevelation

	// ChatFlagPersuasion indicates attempting to convince or manipulate
	ChatFlagPersuasion

	// ChatFlagLie indicates the player is deliberately lying
	ChatFlagLie

	// ChatFlagContradiction indicates information contradicting what NPC knows
	ChatFlagContradiction
)

// String returns the string representation of ChatFlag.
func (f ChatFlag) String() string {
	switch f {
	case ChatFlagHallucination:
		return "hallucination"
	case ChatFlagHostile:
		return "hostile"
	case ChatFlagRevelation:
		return "revelation"
	case ChatFlagPersuasion:
		return "persuasion"
	case ChatFlagLie:
		return "lie"
	case ChatFlagContradiction:
		return "contradiction"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// ParseChatFlag converts a string to ChatFlag.
// It performs case-insensitive matching.
func ParseChatFlag(s string) (ChatFlag, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hallucination":
		return ChatFlagHallucination, nil
	case "hostile":
		return ChatFlagHostile, nil
	case "revelation":
		return ChatFlagRevelation, nil
	case "persuasion":
		return ChatFlagPersuasion, nil
	case "lie":
		return ChatFlagLie, nil
	case "contradiction":
		return ChatFlagContradiction, nil
	default:
		return ChatFlagHallucination, fmt.Errorf("invalid chat flag: %s", s)
	}
}

// MarshalJSON implements json.Marshaler for ChatFlag.
func (f ChatFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// UnmarshalJSON implements json.Unmarshaler for ChatFlag.
func (f *ChatFlag) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseChatFlag(s)
	if err != nil {
		return err
	}
	*f = parsed
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
	CurrentRoom    string                 // Current room/location name
	PlayerID       string                 // Player identifier
	NPCManager     interface{}            // *npc.Manager (avoid import cycle)
	UpdateManager  interface{}            // *knowledge.UpdateManager (avoid import cycle)
	RoomOccupants  map[string][]string    // Room -> Entity IDs mapping
}

// ==========================================================================
// ChatSummary: Summary generated after chat session
// ==========================================================================

// ChatSummary represents a summary of a chat session.
// Story 5.4 AC2: ChatSummary 包含 MainTopics/KeyDecisions/RelationChanges
type ChatSummary struct {
	MainTopics       []string                       `json:"main_topics"`       // Main topics discussed
	KeyDecisions     []string                       `json:"key_decisions"`     // Important decisions made
	RelationChanges  map[string]*manager.EmotionDelta `json:"relation_changes"`  // Emotion changes per NPC
	FactsShared      []string                       `json:"facts_shared"`      // Facts that were shared
	Flags            []ChatFlag                     `json:"flags"`             // Significant flags encountered
	NarrativeImpact  string                         `json:"narrative_impact"`  // How this chat impacts the story
}
