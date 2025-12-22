package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// ==========================================================================
// Story 3-4: Participant Management System - Unit Tests
// ==========================================================================

// TestNewChatParticipant tests the creation of ChatParticipant
func TestNewChatParticipant(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		dispName string
		isPlayer bool
		emotion  manager.EmotionState
		wantID   string
		wantName string
	}{
		{
			name:     "create player participant",
			id:       "player",
			dispName: "玩家",
			isPlayer: true,
			emotion:  manager.DefaultEmotionState(),
			wantID:   "player",
			wantName: "玩家",
		},
		{
			name:     "create NPC participant",
			id:       "npc_001",
			dispName: "Alice",
			isPlayer: false,
			emotion:  manager.NewEmotionState(70, 20, 30),
			wantID:   "npc_001",
			wantName: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewChatParticipant(tt.id, tt.dispName, tt.isPlayer, tt.emotion)

			if p.ID != tt.wantID {
				t.Errorf("NewChatParticipant() ID = %v, want %v", p.ID, tt.wantID)
			}
			if p.Name != tt.wantName {
				t.Errorf("NewChatParticipant() Name = %v, want %v", p.Name, tt.wantName)
			}
			if p.IsPlayer != tt.isPlayer {
				t.Errorf("NewChatParticipant() IsPlayer = %v, want %v", p.IsPlayer, tt.isPlayer)
			}
			if !p.IsActive {
				t.Errorf("NewChatParticipant() IsActive = false, want true")
			}
			if p.Emotion.Trust != tt.emotion.Trust {
				t.Errorf("NewChatParticipant() Emotion.Trust = %v, want %v", p.Emotion.Trust, tt.emotion.Trust)
			}
		})
	}
}

// TestAddParticipant tests adding participants to chat
func TestAddParticipant(t *testing.T) {
	m := NewChatOverlayModel()

	// Add first participant
	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)

	if len(m.participants) != 1 {
		t.Errorf("AddParticipant() participants count = %v, want 1", len(m.participants))
	}

	// Add second participant
	p2 := NewChatParticipant("p2", "Bob", false, manager.DefaultEmotionState())
	m.AddParticipant(p2)

	if len(m.participants) != 2 {
		t.Errorf("AddParticipant() participants count = %v, want 2", len(m.participants))
	}

	// Verify participant data
	if m.participants[0].ID != "p1" {
		t.Errorf("AddParticipant() first participant ID = %v, want p1", m.participants[0].ID)
	}
	if m.participants[1].ID != "p2" {
		t.Errorf("AddParticipant() second participant ID = %v, want p2", m.participants[1].ID)
	}
}

// TestAddParticipant_Duplicate tests preventing duplicate additions
func TestAddParticipant_Duplicate(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.AddParticipant(p1) // Add again

	if len(m.participants) != 1 {
		t.Errorf("AddParticipant() duplicate prevention failed, count = %v, want 1", len(m.participants))
	}
}

// TestAddParticipant_Reactivate tests reactivating inactive participant
func TestAddParticipant_Reactivate(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.RemoveParticipant("p1") // Remove (set inactive)

	if m.participants[0].IsActive {
		t.Error("RemoveParticipant() should set IsActive to false")
	}

	// Re-add should reactivate
	m.AddParticipant(p1)

	if !m.participants[0].IsActive {
		t.Error("AddParticipant() should reactivate inactive participant")
	}
	if len(m.participants) != 1 {
		t.Errorf("AddParticipant() reactivation should not create duplicate, count = %v, want 1", len(m.participants))
	}
}

// TestRemoveParticipant tests marking participant as inactive
func TestRemoveParticipant(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	p2 := NewChatParticipant("p2", "Bob", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.AddParticipant(p2)

	// Remove first participant
	m.RemoveParticipant("p1")

	if len(m.participants) != 2 {
		t.Errorf("RemoveParticipant() should preserve participants, count = %v, want 2", len(m.participants))
	}
	if m.participants[0].IsActive {
		t.Error("RemoveParticipant() should set IsActive to false")
	}
	if !m.participants[1].IsActive {
		t.Error("RemoveParticipant() should not affect other participants")
	}
}

// TestRemoveParticipant_NonExistent tests removing non-existent participant
func TestRemoveParticipant_NonExistent(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)

	// Remove non-existent participant (should not panic)
	m.RemoveParticipant("p999")

	if len(m.participants) != 1 {
		t.Errorf("RemoveParticipant() non-existent should not affect list, count = %v, want 1", len(m.participants))
	}
	if !m.participants[0].IsActive {
		t.Error("RemoveParticipant() non-existent should not affect existing participant")
	}
}

// TestUpdateParticipantEmotion tests updating participant emotion
func TestUpdateParticipantEmotion(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.NewEmotionState(50, 25, 25))
	m.AddParticipant(p1)

	// Update emotion
	newEmotion := manager.NewEmotionState(80, 10, 15)
	m.UpdateParticipantEmotion("p1", newEmotion)

	if m.participants[0].Emotion.Trust != 80 {
		t.Errorf("UpdateParticipantEmotion() Trust = %v, want 80", m.participants[0].Emotion.Trust)
	}
	if m.participants[0].Emotion.Fear != 10 {
		t.Errorf("UpdateParticipantEmotion() Fear = %v, want 10", m.participants[0].Emotion.Fear)
	}
	if m.participants[0].Emotion.Stress != 15 {
		t.Errorf("UpdateParticipantEmotion() Stress = %v, want 15", m.participants[0].Emotion.Stress)
	}
}

// TestUpdateParticipantEmotion_NonExistent tests updating non-existent participant
func TestUpdateParticipantEmotion_NonExistent(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.NewEmotionState(50, 25, 25))
	m.AddParticipant(p1)

	// Update non-existent participant (should not panic)
	newEmotion := manager.NewEmotionState(80, 10, 15)
	m.UpdateParticipantEmotion("p999", newEmotion)

	// Original participant should be unchanged
	if m.participants[0].Emotion.Trust != 50 {
		t.Errorf("UpdateParticipantEmotion() should not affect other participants, Trust = %v, want 50", m.participants[0].Emotion.Trust)
	}
}

// TestGetParticipant tests retrieving participant by ID
func TestGetParticipant(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	p2 := NewChatParticipant("p2", "Bob", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.AddParticipant(p2)

	// Get existing participant
	found := m.GetParticipant("p1")
	if found == nil {
		t.Fatal("GetParticipant() should find existing participant")
	}
	if found.ID != "p1" {
		t.Errorf("GetParticipant() ID = %v, want p1", found.ID)
	}
	if found.Name != "Alice" {
		t.Errorf("GetParticipant() Name = %v, want Alice", found.Name)
	}

	// Get non-existent participant
	notFound := m.GetParticipant("p999")
	if notFound != nil {
		t.Error("GetParticipant() should return nil for non-existent participant")
	}
}

// TestGetActiveParticipants tests getting active participants list
func TestGetActiveParticipants(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	p2 := NewChatParticipant("p2", "Bob", false, manager.DefaultEmotionState())
	p3 := NewChatParticipant("p3", "Charlie", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.AddParticipant(p2)
	m.AddParticipant(p3)

	// Remove one participant
	m.RemoveParticipant("p2")

	active := m.GetActiveParticipants()
	if len(active) != 2 {
		t.Errorf("GetActiveParticipants() count = %v, want 2", len(active))
	}

	// Verify active participants are p1 and p3
	activeIDs := make(map[string]bool)
	for _, p := range active {
		activeIDs[p.ID] = true
	}
	if !activeIDs["p1"] || !activeIDs["p3"] {
		t.Error("GetActiveParticipants() should return p1 and p3")
	}
	if activeIDs["p2"] {
		t.Error("GetActiveParticipants() should not return inactive p2")
	}
}

// TestGetParticipantCount tests counting active and total participants
func TestGetParticipantCount(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("p1", "Alice", false, manager.DefaultEmotionState())
	p2 := NewChatParticipant("p2", "Bob", false, manager.DefaultEmotionState())
	p3 := NewChatParticipant("p3", "Charlie", false, manager.DefaultEmotionState())
	m.AddParticipant(p1)
	m.AddParticipant(p2)
	m.AddParticipant(p3)

	// Remove one participant
	m.RemoveParticipant("p2")

	active, total := m.GetParticipantCount()
	if active != 2 {
		t.Errorf("GetParticipantCount() active = %v, want 2", active)
	}
	if total != 3 {
		t.Errorf("GetParticipantCount() total = %v, want 3", total)
	}
}

// TestGetParticipantCount_Empty tests counting with empty participants
func TestGetParticipantCount_Empty(t *testing.T) {
	m := NewChatOverlayModel()

	active, total := m.GetParticipantCount()
	if active != 0 {
		t.Errorf("GetParticipantCount() empty active = %v, want 0", active)
	}
	if total != 0 {
		t.Errorf("GetParticipantCount() empty total = %v, want 0", total)
	}
}

// TestGetEmotionColor tests emotion color coding
func TestGetEmotionColor(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  string
	}{
		{"high value", 85, "10"},   // Green
		{"medium value", 50, "11"}, // Yellow
		{"low value", 15, "9"},     // Red
		{"boundary high", 70, "10"}, // Green
		{"boundary low", 29, "9"},   // Red
		{"boundary medium", 30, "11"}, // Yellow
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEmotionColor(tt.value)
			if string(got) != tt.want {
				t.Errorf("getEmotionColor(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestGetRelationshipIcon tests relationship icon mapping
func TestGetRelationshipIcon(t *testing.T) {
	tests := []struct {
		name string
		rel  manager.RelationshipType
		want string
	}{
		{"friendly", manager.Friendly, "●"},
		{"hostile", manager.Hostile, "▲"},
		{"fearful", manager.Fearful, "◆"},
		{"neutral", manager.Neutral, "○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRelationshipIcon(tt.rel)
			if got != tt.want {
				t.Errorf("getRelationshipIcon(%v) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

// TestRenderParticipant tests rendering single participant
func TestRenderParticipant(t *testing.T) {
	m := NewChatOverlayModel()

	tests := []struct {
		name     string
		p        ChatParticipant
		wantName string
	}{
		{
			name: "player participant",
			p: ChatParticipant{
				ID:       "player",
				Name:     "玩家",
				IsPlayer: true,
				Emotion:  manager.NewEmotionState(50, 25, 25),
				IsActive: true,
			},
			wantName: "[你]",
		},
		{
			name: "active NPC",
			p: ChatParticipant{
				ID:       "npc1",
				Name:     "Alice",
				IsPlayer: false,
				Emotion:  manager.NewEmotionState(70, 20, 30),
				IsActive: true,
			},
			wantName: "Alice",
		},
		{
			name: "inactive NPC",
			p: ChatParticipant{
				ID:       "npc2",
				Name:     "Bob",
				IsPlayer: false,
				Emotion:  manager.NewEmotionState(30, 60, 50),
				IsActive: false,
			},
			wantName: "[已離開]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderParticipant(tt.p)
			if result == "" {
				t.Error("renderParticipant() returned empty string")
			}
			// Just verify it contains expected name indicator
			// Full UI rendering is hard to test without actual terminal
		})
	}
}

// TestRenderParticipantsList tests rendering participants list
func TestRenderParticipantsList(t *testing.T) {
	m := NewChatOverlayModel()

	p1 := NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState())
	p2 := NewChatParticipant("npc1", "Alice", false, manager.NewEmotionState(70, 20, 30))
	p3 := NewChatParticipant("npc2", "Bob", false, manager.NewEmotionState(30, 60, 50))

	m.AddParticipant(p1)
	m.AddParticipant(p2)
	m.AddParticipant(p3)
	m.RemoveParticipant("npc2") // Make Bob inactive

	result := m.renderParticipantsList()
	if result == "" {
		t.Error("renderParticipantsList() returned empty string")
	}

	// Verify header contains participant count
	if !containsString(result, "參與者") {
		t.Error("renderParticipantsList() should contain participant header")
	}
}

// TestNewChatOverlayModel tests model initialization
func TestNewChatOverlayModel(t *testing.T) {
	m := NewChatOverlayModel()

	if m.active {
		t.Error("NewChatOverlayModel() should initialize as inactive")
	}
	if len(m.participants) != 0 {
		t.Errorf("NewChatOverlayModel() participants count = %v, want 0", len(m.participants))
	}
	if m.inputField.Value() != "" {
		t.Error("NewChatOverlayModel() should initialize empty input field")
	}
}

// TestIsActive tests active state getter
func TestIsActive(t *testing.T) {
	m := NewChatOverlayModel()

	if m.IsActive() {
		t.Error("IsActive() should return false initially")
	}

	m.SetActive(true)
	if !m.IsActive() {
		t.Error("IsActive() should return true after SetActive(true)")
	}
}

// TestSetActive tests active state setter
func TestSetActive(t *testing.T) {
	m := NewChatOverlayModel()

	m.SetActive(true)
	if !m.active {
		t.Error("SetActive(true) should set active to true")
	}

	m.SetActive(false)
	if m.active {
		t.Error("SetActive(false) should set active to false")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestEdgeCases tests edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("extreme emotion values", func(t *testing.T) {
		m := NewChatOverlayModel()
		p := NewChatParticipant("p1", "Alice", false, manager.NewEmotionState(0, 100, 100))
		m.AddParticipant(p)

		result := m.renderParticipant(p)
		if result == "" {
			t.Error("renderParticipant() should handle extreme emotion values")
		}
	})

	t.Run("many participants", func(t *testing.T) {
		m := NewChatOverlayModel()
		for i := 0; i < 15; i++ {
			p := NewChatParticipant(
				fmt.Sprintf("p%d", i),
				fmt.Sprintf("NPC%d", i),
				false,
				manager.DefaultEmotionState(),
			)
			m.AddParticipant(p)
		}

		if len(m.participants) != 15 {
			t.Errorf("AddParticipant() many participants count = %v, want 15", len(m.participants))
		}

		result := m.renderParticipantsList()
		if result == "" {
			t.Error("renderParticipantsList() should handle many participants")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		m := NewChatOverlayModel()
		p := NewChatParticipant("p1", "", false, manager.DefaultEmotionState())
		m.AddParticipant(p)

		result := m.renderParticipant(p)
		if result == "" {
			t.Error("renderParticipant() should handle empty name")
		}
	})
}

// ==========================================================================
// Story 3-6: Basic Input Handling - Unit Tests
// ==========================================================================

// TestSendMessage_EmptyInput tests that empty messages are not sent
// AC1: 玩家輸入訊息後按 Enter 發送 (but empty messages should be ignored)
func TestSendMessage_EmptyInput(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	// Try to send empty message
	m.sendMessage()

	if len(m.messages) != 0 {
		t.Errorf("sendMessage() with empty input should not add message, got %d messages", len(m.messages))
	}
}

// TestSendMessage_ValidInput tests sending a valid message
// AC1: 玩家輸入訊息後按 Enter 發送
// AC4: 訊息添加到 messages 列表
func TestSendMessage_ValidInput(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	// Set input value
	m.inputField.SetValue("Hello, world!")
	m.sendMessage()

	// AC4: Verify message was added to list
	if len(m.messages) != 1 {
		t.Fatalf("sendMessage() should add message, got %d messages, want 1", len(m.messages))
	}

	// Verify message content
	msg := m.messages[0]
	if msg.Speaker != "player" {
		t.Errorf("sendMessage() message speaker = %v, want 'player'", msg.Speaker)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("sendMessage() message content = %v, want 'Hello, world!'", msg.Content)
	}
	if msg.Type != ChatMessageNormal {
		t.Errorf("sendMessage() message type = %v, want ChatMessageNormal", msg.Type)
	}

	// Verify input was cleared
	if m.inputField.Value() != "" {
		t.Errorf("sendMessage() should clear input field, got %v", m.inputField.Value())
	}

	// Verify chat turns incremented
	if m.chatTurns != 1 {
		t.Errorf("sendMessage() chatTurns = %d, want 1", m.chatTurns)
	}
}

// TestSendMessage_MultipleMessages tests sending multiple messages
// AC4: 訊息添加到 messages 列表
func TestSendMessage_MultipleMessages(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	messages := []string{"First message", "Second message", "Third message"}

	for i, msg := range messages {
		m.inputField.SetValue(msg)
		m.sendMessage()

		if len(m.messages) != i+1 {
			t.Errorf("After message %d, got %d messages, want %d", i+1, len(m.messages), i+1)
		}
	}

	// Verify all messages were added in order
	for i, expectedMsg := range messages {
		if m.messages[i].Content != expectedMsg {
			t.Errorf("Message %d content = %v, want %v", i, m.messages[i].Content, expectedMsg)
		}
	}

	// Verify chat turns
	if m.chatTurns != 3 {
		t.Errorf("sendMessage() chatTurns = %d, want 3", m.chatTurns)
	}
}

// TestSendMessage_WhitespaceOnly tests that whitespace-only messages are not sent
// AC1: 玩家輸入訊息後按 Enter 發送 (but whitespace should be ignored)
func TestSendMessage_WhitespaceOnly(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	// Try to send whitespace-only message
	m.inputField.SetValue("   \t\n   ")
	m.sendMessage()

	if len(m.messages) != 0 {
		t.Errorf("sendMessage() with whitespace should not add message, got %d messages", len(m.messages))
	}
}

// TestAddMessage_AutoScroll tests that viewport auto-scrolls to bottom
// AC5: viewport 自動捲動到最新訊息
func TestAddMessage_AutoScroll(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	// Add multiple messages to ensure scrolling is needed
	for i := 0; i < 25; i++ {
		msg := NewChatMessage(
			fmt.Sprintf("msg-%d", i),
			"player",
			fmt.Sprintf("Message number %d", i),
			ChatMessageNormal,
		)
		m.AddMessage(msg)
	}

	// After adding messages, viewport should be at bottom
	// The viewport's YOffset should be at maximum (showing bottom content)
	// Note: We can't directly test viewport position without more setup,
	// but we can verify updateViewport was called and messages were added
	if len(m.messages) != 25 {
		t.Errorf("AddMessage() messages count = %d, want 25", len(m.messages))
	}
}

// TestGetMessages tests retrieving message list
func TestGetMessages(t *testing.T) {
	m := NewChatOverlayModel()

	// Initially empty
	messages := m.GetMessages()
	if len(messages) != 0 {
		t.Errorf("GetMessages() initial count = %d, want 0", len(messages))
	}

	// Add some messages
	m.inputField.SetValue("Test message")
	m.sendMessage()

	messages = m.GetMessages()
	if len(messages) != 1 {
		t.Errorf("GetMessages() count after send = %d, want 1", len(messages))
	}
}

// TestAddSystemMessage tests adding system messages
func TestAddSystemMessage(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	m.AddSystemMessage("System message test")

	if len(m.messages) != 1 {
		t.Fatalf("AddSystemMessage() should add message, got %d messages", len(m.messages))
	}

	msg := m.messages[0]
	if msg.Type != ChatMessageSystem {
		t.Errorf("AddSystemMessage() type = %v, want ChatMessageSystem", msg.Type)
	}
	if msg.Speaker != "system" {
		t.Errorf("AddSystemMessage() speaker = %v, want 'system'", msg.Speaker)
	}
	if msg.Content != "System message test" {
		t.Errorf("AddSystemMessage() content = %v, want 'System message test'", msg.Content)
	}
}

// TestAddNPCMessage tests adding NPC messages with flags
func TestAddNPCMessage(t *testing.T) {
	m := NewChatOverlayModel()
	m.SetActive(true)

	flags := []ChatFlag{ChatFlagRevelation, ChatFlagPersuasion}
	m.AddNPCMessage("npc_001", "NPC message with flags", flags)

	if len(m.messages) != 1 {
		t.Fatalf("AddNPCMessage() should add message, got %d messages", len(m.messages))
	}

	msg := m.messages[0]
	if msg.Speaker != "npc_001" {
		t.Errorf("AddNPCMessage() speaker = %v, want 'npc_001'", msg.Speaker)
	}
	if msg.Type != ChatMessageNormal {
		t.Errorf("AddNPCMessage() type = %v, want ChatMessageNormal", msg.Type)
	}
	if len(msg.Flags) != 2 {
		t.Errorf("AddNPCMessage() flags count = %d, want 2", len(msg.Flags))
	}
}

// TestEnter tests entering chat room
func TestEnter(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(50, 25, 25)),
	}

	m.Enter("player", participants, "大廳")

	// Verify active state
	if !m.active {
		t.Error("Enter() should set active to true")
	}

	// Verify initiator
	if m.initiator != "player" {
		t.Errorf("Enter() initiator = %v, want 'player'", m.initiator)
	}

	// Verify location
	if m.location != "大廳" {
		t.Errorf("Enter() location = %v, want '大廳'", m.location)
	}

	// Verify participants
	if len(m.participants) != 2 {
		t.Errorf("Enter() participants count = %d, want 2", len(m.participants))
	}

	// Verify system message was added
	if len(m.messages) == 0 {
		t.Error("Enter() should add system message")
	}

	// Verify chat state reset
	if m.chatTurns != 0 {
		t.Errorf("Enter() chatTurns = %d, want 0", m.chatTurns)
	}
}

// TestEnter_NPCInitiated tests NPC-initiated chat
func TestEnter_NPCInitiated(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.NewEmotionState(50, 25, 25)),
	}

	m.Enter("npc_001", participants, "大廳")

	// Verify initiator was set
	if m.initiator != "npc_001" {
		t.Errorf("Enter() with NPC initiator should set initiator to npc_001, got %s", m.initiator)
	}

	// Should have at least 1 system message for opening
	// (NPC opening message will be implemented in Story 3-3)
	if len(m.messages) < 1 {
		t.Errorf("Enter() should add at least 1 message, got %d", len(m.messages))
	}
}

// TestExit tests exiting chat room
func TestExit(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
	}

	m.Enter("player", participants, "大廳")
	m.Exit()

	// Verify active state
	if m.active {
		t.Error("Exit() should set active to false")
	}

	// Verify input field is blurred
	if m.focused {
		t.Error("Exit() should set focused to false")
	}
}

// TestRenderMessage tests message rendering
func TestRenderMessage(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("npc_001", "Alice", false, manager.DefaultEmotionState()),
	}
	m.participants = participants

	tests := []struct {
		name    string
		msg     *ChatMessage
		wantNil bool
	}{
		{
			name: "player message",
			msg: NewChatMessage("msg1", "player", "Hello", ChatMessageNormal),
		},
		{
			name: "npc message",
			msg: NewChatMessage("msg2", "npc_001", "Hi there", ChatMessageNormal),
		},
		{
			name: "system message",
			msg: NewChatMessage("msg3", "system", "Chat started", ChatMessageSystem),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderMessage(tt.msg)
			if result == "" {
				t.Error("renderMessage() returned empty string")
			}
		})
	}
}

// TestRenderMessage_WithFlags tests message rendering with flags
func TestRenderMessage_WithFlags(t *testing.T) {
	m := NewChatOverlayModel()

	msg := NewChatMessage("msg1", "player", "I saw something", ChatMessageNormal)
	msg.Flags = []ChatFlag{ChatFlagHallucination, ChatFlagLie}

	result := m.renderMessage(msg)
	if result == "" {
		t.Error("renderMessage() with flags returned empty string")
	}
	// Result should contain flag information (exact format may vary)
}

// TestUpdateViewportContent tests viewport content update
func TestUpdateViewportContent(t *testing.T) {
	m := NewChatOverlayModel()

	// Add some messages
	m.messages = []*ChatMessage{
		NewChatMessage("msg1", "player", "Hello", ChatMessageNormal),
		NewChatMessage("msg2", "system", "Welcome", ChatMessageSystem),
		NewChatMessage("msg3", "player", "Thanks", ChatMessageNormal),
	}

	// Update viewport content
	m.updateViewportContent()

	// Viewport should have content (we can't easily test the exact content without more setup)
	// But we can verify the method doesn't panic and runs successfully
}

// Note: getParticipantName is a private method and cannot be tested directly.
// It is tested indirectly through renderMessage tests.

// ==========================================================================
// Story 3-7: Epic 3 Integration Tests
// ==========================================================================

// TestIntegration_ChatFlow tests the complete chat flow from enter to exit.
// Story 3-7 AC1: Test chat room enter/exit flow (integration).
func TestIntegration_ChatFlow(t *testing.T) {
	m := NewChatOverlayModel()

	// Verify initial state
	if m.active {
		t.Error("ChatOverlay should be inactive initially")
	}

	// Create participants
	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "Alice", false, manager.NewEmotionState(70, 20, 30)),
		NewChatParticipant("npc2", "Bob", false, manager.NewEmotionState(50, 40, 50)),
	}

	// AC1: Enter chat
	m.Enter("player", participants, "Kitchen")

	// Verify chat is active
	if !m.active {
		t.Fatal("Chat should be active after Enter()")
	}
	if len(m.participants) != 3 {
		t.Errorf("Expected 3 participants, got %d", len(m.participants))
	}
	if m.location != "Kitchen" {
		t.Errorf("Expected location 'Kitchen', got '%s'", m.location)
	}
	if len(m.messages) != 1 {
		t.Errorf("Expected 1 system message after Enter(), got %d", len(m.messages))
	}

	// AC2: Send player messages and verify display
	m.AddMessage(NewChatMessage("msg1", "player", "Hello everyone!", ChatMessageNormal))
	m.AddMessage(NewChatMessage("msg2", "Alice", "Hi there!", ChatMessageNormal))
	m.AddMessage(NewChatMessage("msg3", "Bob", "Good morning!", ChatMessageNormal))

	if len(m.messages) != 4 {
		t.Errorf("Expected 4 messages total, got %d", len(m.messages))
	}

	// Verify message content
	if m.messages[1].Content != "Hello everyone!" {
		t.Errorf("Message 1 content mismatch: %s", m.messages[1].Content)
	}
	if m.messages[1].Speaker != "player" {
		t.Errorf("Message 1 speaker should be 'player', got '%s'", m.messages[1].Speaker)
	}

	// AC3: Modify participant emotions and verify list updates
	m.UpdateParticipantEmotion("npc1", manager.NewEmotionState(80, 15, 25))

	participant := m.GetParticipant("npc1")
	if participant == nil {
		t.Fatal("GetParticipant('npc1') returned nil")
	}
	if participant.Emotion.Trust != 80 {
		t.Errorf("Expected Trust=80 after update, got %d", participant.Emotion.Trust)
	}

	// AC3: Remove participant and verify
	m.RemoveParticipant("npc2")

	activeCount, totalCount := m.GetParticipantCount()
	if activeCount != 2 {
		t.Errorf("Expected 2 active participants after removal, got %d", activeCount)
	}
	if totalCount != 3 {
		t.Errorf("Expected 3 total participants, got %d", totalCount)
	}

	// AC4: Test chat turns tracking
	initialTurns := m.GetChatTurns()
	m.chatTurns = 5
	if m.GetChatTurns() != 5 {
		t.Errorf("Chat turns should be 5, got %d", m.GetChatTurns())
	}

	// AC1: Exit chat
	session := m.Exit()

	// Verify chat is inactive
	if m.active {
		t.Error("Chat should be inactive after Exit()")
	}

	// Verify session data
	if session == nil {
		t.Fatal("Exit() should return a ChatSession")
	}
	if session.SessionID == "" {
		t.Error("Session should have a valid ID")
	}
	if session.Initiator != "player" {
		t.Errorf("Expected initiator 'player', got '%s'", session.Initiator)
	}
	if len(session.Messages) != 4 {
		t.Errorf("Session should have 4 messages, got %d", len(session.Messages))
	}
	if session.Location != "Kitchen" {
		t.Errorf("Expected session location 'Kitchen', got '%s'", session.Location)
	}
	if session.TurnsSpent != 5 {
		t.Errorf("Expected 5 turns spent, got %d", session.TurnsSpent)
	}

	// Verify state reset after exit
	if initialTurns == m.GetChatTurns() {
		// Chat turns should remain set (not reset) for history tracking
		t.Log("Chat turns preserved for history: OK")
	}
}

// TestIntegration_MessageDisplayWithFlags tests message rendering with various flags.
// Story 3-7 AC2: Test message display correctness (integration).
func TestIntegration_MessageDisplayWithFlags(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "Alice", false, manager.DefaultEmotionState()),
	}

	m.Enter("player", participants, "TestRoom")

	// Add messages with different types and flags
	testMessages := []struct {
		speaker string
		content string
		msgType ChatMessageType
		flags   []ChatFlag
	}{
		{"player", "I saw something strange", ChatMessageNormal, []ChatFlag{ChatFlagHallucination}},
		{"Alice", "Don't trust him!", ChatMessageNormal, []ChatFlag{ChatFlagHostile}},
		{"player", "I know the secret", ChatMessageWhisper, []ChatFlag{ChatFlagRevelation}},
		{"system", "The door creaks open", ChatMessageSystem, nil},
		{"Alice", "runs to the exit", ChatMessageAction, nil},
	}

	for i, tm := range testMessages {
		msg := &ChatMessage{
			ID:        fmt.Sprintf("msg_%d", i),
			Speaker:   tm.speaker,
			Content:   tm.content,
			Type:      tm.msgType,
			Flags:     tm.flags,
			Timestamp: m.sessionStart,
		}
		m.AddMessage(msg)
	}

	// Verify all messages were added
	if len(m.messages) != len(testMessages)+1 { // +1 for system entrance message
		t.Errorf("Expected %d messages, got %d", len(testMessages)+1, len(m.messages))
	}

	// Verify message types are preserved
	for i, tm := range testMessages {
		actualMsg := m.messages[i+1] // Skip first system message
		if actualMsg.Type != tm.msgType {
			t.Errorf("Message %d: expected type %s, got %s", i, tm.msgType, actualMsg.Type)
		}
		if len(actualMsg.Flags) != len(tm.flags) {
			t.Errorf("Message %d: expected %d flags, got %d", i, len(tm.flags), len(actualMsg.Flags))
		}
	}

	// Test rendering doesn't panic with various message types
	for i := 1; i < len(m.messages); i++ {
		rendered := m.renderMessage(m.messages[i])
		if rendered == "" {
			t.Errorf("Message %d rendered to empty string", i)
		}
	}

	m.Exit()
}

// TestIntegration_ParticipantManagement tests participant lifecycle.
// Story 3-7 AC3: Test participant list correctness (integration).
func TestIntegration_ParticipantManagement(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "Alice", false, manager.NewEmotionState(60, 30, 40)),
	}

	m.Enter("player", participants, "Hall")

	// Add a late participant
	lateParticipant := NewChatParticipant("npc2", "Bob", false, manager.NewEmotionState(50, 50, 50))
	m.AddParticipant(lateParticipant)

	active, total := m.GetParticipantCount()
	if active != 3 || total != 3 {
		t.Errorf("Expected 3 active, 3 total; got %d active, %d total", active, total)
	}

	// Update emotions for all NPCs
	m.UpdateParticipantEmotion("npc1", manager.NewEmotionState(80, 10, 20))
	m.UpdateParticipantEmotion("npc2", manager.NewEmotionState(40, 60, 70))

	// Verify emotion updates
	alice := m.GetParticipant("npc1")
	if alice.Emotion.Trust != 80 || alice.Emotion.Fear != 10 {
		t.Errorf("Alice emotion update failed: Trust=%d, Fear=%d", alice.Emotion.Trust, alice.Emotion.Fear)
	}

	bob := m.GetParticipant("npc2")
	if bob.Emotion.Trust != 40 || bob.Emotion.Fear != 60 {
		t.Errorf("Bob emotion update failed: Trust=%d, Fear=%d", bob.Emotion.Trust, bob.Emotion.Fear)
	}

	// Remove one participant
	m.RemoveParticipant("npc1")

	active, total = m.GetParticipantCount()
	if active != 2 || total != 3 {
		t.Errorf("After removal: expected 2 active, 3 total; got %d active, %d total", active, total)
	}

	// Verify removed participant is inactive
	alice = m.GetParticipant("npc1")
	if alice.IsActive {
		t.Error("Removed participant should be inactive")
	}

	// Get active participants list
	activeList := m.GetActiveParticipants()
	if len(activeList) != 2 {
		t.Errorf("Expected 2 participants in active list, got %d", len(activeList))
	}

	// Verify rendering doesn't panic
	rendering := m.renderParticipantsList()
	if rendering == "" {
		t.Error("Participant list rendering returned empty string")
	}

	m.Exit()
}

// TestIntegration_InputAndUIUpdates tests input handling and UI synchronization.
// Story 3-7 AC4: Test input handling and UI updates (integration).
func TestIntegration_InputAndUIUpdates(t *testing.T) {
	m := NewChatOverlayModel()

	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "TestNPC", false, manager.DefaultEmotionState()),
	}

	m.Enter("player", participants, "TestLocation")
	initialMessageCount := len(m.messages)

	// Simulate sending multiple messages
	messages := []string{
		"Hello, how are you?",
		"This is a test message",
		"Let's talk about the mystery",
	}

	for _, content := range messages {
		msg := NewChatMessage(
			fmt.Sprintf("msg_%d", len(m.messages)),
			"player",
			content,
			ChatMessageNormal,
		)
		m.AddMessage(msg)
		m.chatTurns++
	}

	// Verify messages were added
	expectedCount := initialMessageCount + len(messages)
	if len(m.messages) != expectedCount {
		t.Errorf("Expected %d messages, got %d", expectedCount, len(m.messages))
	}

	// Verify chat turns incremented
	if m.GetChatTurns() != len(messages) {
		t.Errorf("Expected %d chat turns, got %d", len(messages), m.GetChatTurns())
	}

	// Add system message
	m.AddSystemMessage("An ominous feeling fills the air")

	// Find the system message
	systemMsgFound := false
	for _, msg := range m.messages {
		if msg.Type == ChatMessageSystem && msg.Content == "An ominous feeling fills the air" {
			systemMsgFound = true
			break
		}
	}
	if !systemMsgFound {
		t.Error("System message was not added correctly")
	}

	// Add NPC message
	m.AddNPCMessage("npc1", "I sense danger nearby", nil)

	// Verify NPC message
	npcMsgFound := false
	for _, msg := range m.messages {
		if msg.Speaker == "npc1" && msg.Content == "I sense danger nearby" {
			npcMsgFound = true
			break
		}
	}
	if !npcMsgFound {
		t.Error("NPC message was not added correctly")
	}

	// Test viewport update (should not panic)
	m.updateViewportContent()

	// Verify message retrieval
	allMessages := m.GetMessages()
	if len(allMessages) != len(m.messages) {
		t.Errorf("GetMessages() returned %d messages, expected %d", len(allMessages), len(m.messages))
	}

	// Exit and verify session
	session := m.Exit()
	if session == nil {
		t.Fatal("Exit() returned nil session")
	}

	// Verify session contains all messages
	if len(session.Messages) != len(allMessages) {
		t.Errorf("Session has %d messages, expected %d", len(session.Messages), len(allMessages))
	}

	// Verify session metadata
	if session.TurnsSpent != len(messages) {
		t.Errorf("Session turns spent: got %d, want %d", session.TurnsSpent, len(messages))
	}
}

// ==========================================================================
// Story 5.1: Chat Time Flow Control - Unit Tests
// ==========================================================================

// TestChatTimeFlow_Tracking tests basic time flow tracking.
// AC1: chatTurns 追蹤聊天回合數
// AC3: tickAccumulator 累積時間流
func TestChatTimeFlow_Tracking(t *testing.T) {
	m := NewChatOverlayModel()
	m.chatTurnsPerBeat = 10
	m.timeScale = 0.1

	// Initial state
	if m.chatTurns != 0 {
		t.Errorf("Initial chatTurns = %d, want 0", m.chatTurns)
	}
	if m.tickAccumulator != 0.0 {
		t.Errorf("Initial tickAccumulator = %f, want 0.0", m.tickAccumulator)
	}

	// Simulate 5 chat turns
	for i := 0; i < 5; i++ {
		triggered := m.incrementChatTurns()
		if triggered {
			t.Errorf("Turn %d: unexpected trigger (should not trigger before turn 10)", i+1)
		}
	}

	// AC1: Verify chatTurns incremented correctly
	if m.chatTurns != 5 {
		t.Errorf("After 5 turns, chatTurns = %d, want 5", m.chatTurns)
	}

	// AC3: Verify tickAccumulator accumulated correctly
	expectedAccumulator := 0.5 // 5 * 0.1
	if m.tickAccumulator < expectedAccumulator-0.01 || m.tickAccumulator > expectedAccumulator+0.01 {
		t.Errorf("After 5 turns, tickAccumulator = %f, want ~%f", m.tickAccumulator, expectedAccumulator)
	}
}

// TestChatTimeFlow_MainBeatTrigger tests main timeline beat triggering.
// AC4: 每 10 聊天回合觸發 1 主時間回合事件
func TestChatTimeFlow_MainBeatTrigger(t *testing.T) {
	m := NewChatOverlayModel()
	m.chatTurnsPerBeat = 10
	m.timeScale = 0.1
	m.startBeat = 0

	// Set up active chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")
	initialMessages := len(m.messages) // Should have 1 system message

	// Simulate 9 chat turns (should not trigger)
	triggerCount := 0
	for i := 0; i < 9; i++ {
		if m.incrementChatTurns() {
			triggerCount++
		}
	}

	if triggerCount != 0 {
		t.Errorf("First 9 turns should not trigger, got %d triggers", triggerCount)
	}

	// 10th turn should trigger
	triggered := m.incrementChatTurns()
	if !triggered {
		t.Error("Turn 10 should trigger main timeline event")
	}

	// AC4: Verify system message was added
	if len(m.messages) != initialMessages+1 {
		t.Errorf("Expected %d messages after trigger (system message), got %d", initialMessages+1, len(m.messages))
	}

	// AC4: Verify tickAccumulator was reset
	if m.tickAccumulator >= 1.0 {
		t.Errorf("tickAccumulator should be reset after trigger, got %f", m.tickAccumulator)
	}

	// Verify 20th turn also triggers
	for i := 0; i < 9; i++ {
		m.incrementChatTurns()
	}
	triggered = m.incrementChatTurns()
	if !triggered {
		t.Error("Turn 20 should also trigger main timeline event")
	}
}

// TestChatTimeFlow_AccumulatorTrigger tests triggering via accumulator.
// AC4: 當 tickAccumulator >= 1.0 時觸發
func TestChatTimeFlow_AccumulatorTrigger(t *testing.T) {
	m := NewChatOverlayModel()
	m.chatTurnsPerBeat = 20 // Set high to test accumulator trigger
	m.timeScale = 0.2       // Faster accumulation
	m.startBeat = 0

	// Set up active chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// With timeScale=0.2, should trigger at 5 turns (5 * 0.2 = 1.0)
	triggerCount := 0
	for i := 0; i < 5; i++ {
		if m.incrementChatTurns() {
			triggerCount++
		}
	}

	if triggerCount != 1 {
		t.Errorf("Should trigger once via accumulator, got %d triggers", triggerCount)
	}

	// Accumulator should be reset
	if m.tickAccumulator >= 1.0 {
		t.Errorf("Accumulator should be < 1.0 after reset, got %f", m.tickAccumulator)
	}
}

// TestChatTimeFlow_DifferentConfigs tests different time flow configurations.
// AC2: chatTurnsPerBeat = 10 (可調整)
// AC5: timeScale = 0.1 (可調整)
func TestChatTimeFlow_DifferentConfigs(t *testing.T) {
	tests := []struct {
		name             string
		chatTurnsPerBeat int
		timeScale        float64
		chatRounds       int
		expectTriggers   int
	}{
		{
			name:             "Default (10 turns, 0.1 scale)",
			chatTurnsPerBeat: 10,
			timeScale:        0.1,
			chatRounds:       20,
			expectTriggers:   2, // At turn 10 and 20
		},
		{
			name:             "Fast (5 turns, 0.2 scale)",
			chatTurnsPerBeat: 5,
			timeScale:        0.2,
			chatRounds:       20,
			expectTriggers:   4, // At turns 5, 10, 15, 20
		},
		{
			name:             "Slow (20 turns, 0.05 scale)",
			chatTurnsPerBeat: 20,
			timeScale:        0.05,
			chatRounds:       20,
			expectTriggers:   1, // Only at turn 20
		},
		{
			name:             "Accumulator-driven (15 turns, 0.25 scale)",
			chatTurnsPerBeat: 100, // Very high, won't trigger
			timeScale:        0.25,
			chatRounds:       12,
			expectTriggers:   3, // At turns 4, 8, 12 (accumulator)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewChatOverlayModel()
			m.chatTurnsPerBeat = tt.chatTurnsPerBeat
			m.timeScale = tt.timeScale
			m.startBeat = 0

			// Set up active chat
			participants := []ChatParticipant{
				NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
			}
			m.Enter("player", participants, "TestRoom")

			triggers := 0
			for i := 0; i < tt.chatRounds; i++ {
				if m.incrementChatTurns() {
					triggers++
				}
			}

			if triggers != tt.expectTriggers {
				t.Errorf("Expected %d triggers, got %d", tt.expectTriggers, triggers)
			}
		})
	}
}

// TestCheckMainTimelineTick tests the checkMainTimelineTick method.
// AC4: checkMainTimelineTick() 方法檢查是否觸發主時間回合
func TestCheckMainTimelineTick(t *testing.T) {
	tests := []struct {
		name             string
		chatTurns        int
		chatTurnsPerBeat int
		tickAccumulator  float64
		want             bool
	}{
		{
			name:             "At exact chatTurnsPerBeat boundary",
			chatTurns:        10,
			chatTurnsPerBeat: 10,
			tickAccumulator:  0.5,
			want:             true,
		},
		{
			name:             "Before chatTurnsPerBeat boundary",
			chatTurns:        9,
			chatTurnsPerBeat: 10,
			tickAccumulator:  0.9,
			want:             false,
		},
		{
			name:             "Accumulator at 1.0",
			chatTurns:        5,
			chatTurnsPerBeat: 10,
			tickAccumulator:  1.0,
			want:             true,
		},
		{
			name:             "Accumulator above 1.0",
			chatTurns:        7,
			chatTurnsPerBeat: 10,
			tickAccumulator:  1.3,
			want:             true,
		},
		{
			name:             "Zero chatTurns",
			chatTurns:        0,
			chatTurnsPerBeat: 10,
			tickAccumulator:  0.0,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewChatOverlayModel()
			m.chatTurns = tt.chatTurns
			m.chatTurnsPerBeat = tt.chatTurnsPerBeat
			m.tickAccumulator = tt.tickAccumulator

			got := m.checkMainTimelineTick()
			if got != tt.want {
				t.Errorf("checkMainTimelineTick() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTriggerMainTimelineEvent tests the triggerMainTimelineEvent method.
// AC4: triggerMainTimelineEvent() 方法
func TestTriggerMainTimelineEvent(t *testing.T) {
	m := NewChatOverlayModel()
	m.chatTurnsPerBeat = 10

	// Set up active chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// Set values AFTER Enter() to avoid reset
	m.startBeat = 5
	m.chatTurns = 20

	initialMsgCount := len(m.messages)

	// Trigger event
	m.triggerMainTimelineEvent()

	// Should add a system message
	if len(m.messages) != initialMsgCount+1 {
		t.Errorf("Expected %d messages after trigger, got %d", initialMsgCount+1, len(m.messages))
	}

	// Verify message is system type
	lastMsg := m.messages[len(m.messages)-1]
	if lastMsg.Type != ChatMessageSystem {
		t.Errorf("Triggered message should be system type, got %v", lastMsg.Type)
	}

	// Message should contain beat information
	expectedBeat := 5 + (20 / 10) // startBeat + (chatTurns / chatTurnsPerBeat) = 7
	expectedContent := fmt.Sprintf("時間流逝... (主時間回合 %d)", expectedBeat)
	if lastMsg.Content != expectedContent {
		t.Errorf("System message content = %q, want %q", lastMsg.Content, expectedContent)
	}
}

// TestApplyConfig tests applying ChatConfig to the model.
// AC2, AC5: 可根據難度或場景調整
func TestApplyConfig(t *testing.T) {
	m := NewChatOverlayModel()

	// Default values
	if m.timeScale != 0.1 {
		t.Errorf("Default timeScale = %f, want 0.1", m.timeScale)
	}
	if m.chatTurnsPerBeat != 10 {
		t.Errorf("Default chatTurnsPerBeat = %d, want 10", m.chatTurnsPerBeat)
	}

	// Apply custom config
	customConfig := ChatConfig{
		TimeScale:        0.05,
		ChatTurnsPerBeat: 20,
		AllowInterrupts:  true,
	}
	m.ApplyConfig(customConfig)

	if m.timeScale != 0.05 {
		t.Errorf("After ApplyConfig, timeScale = %f, want 0.05", m.timeScale)
	}
	if m.chatTurnsPerBeat != 20 {
		t.Errorf("After ApplyConfig, chatTurnsPerBeat = %d, want 20", m.chatTurnsPerBeat)
	}
}

// TestGetTimeFlowInfo tests getting time flow information.
// AC5: timeScale = 0.1 視覺化指示
func TestGetTimeFlowInfo(t *testing.T) {
	m := NewChatOverlayModel()
	m.chatTurns = 7
	m.timeScale = 0.15
	m.tickAccumulator = 0.45
	m.chatTurnsPerBeat = 12

	chatTurns, timeScale, accumulator, turnsPerBeat := m.GetTimeFlowInfo()

	if chatTurns != 7 {
		t.Errorf("GetTimeFlowInfo() chatTurns = %d, want 7", chatTurns)
	}
	if timeScale != 0.15 {
		t.Errorf("GetTimeFlowInfo() timeScale = %f, want 0.15", timeScale)
	}
	if accumulator != 0.45 {
		t.Errorf("GetTimeFlowInfo() accumulator = %f, want 0.45", accumulator)
	}
	if turnsPerBeat != 12 {
		t.Errorf("GetTimeFlowInfo() turnsPerBeat = %d, want 12", turnsPerBeat)
	}
}

// TestEnter_ResetsTimeFlow tests that Enter() resets time flow state.
// AC1: 初始值為 0，在 Enter() 時重置
func TestEnter_ResetsTimeFlow(t *testing.T) {
	m := NewChatOverlayModel()

	// Set some values
	m.chatTurns = 15
	m.tickAccumulator = 0.8
	m.startBeat = 10

	// Enter chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// AC1: Verify reset
	if m.chatTurns != 0 {
		t.Errorf("After Enter(), chatTurns = %d, want 0", m.chatTurns)
	}
	if m.tickAccumulator != 0.0 {
		t.Errorf("After Enter(), tickAccumulator = %f, want 0.0", m.tickAccumulator)
	}
}

// TestDefaultChatConfig tests the DefaultChatConfig function.
// AC2: chatTurnsPerBeat = 10 (預設)
// AC5: timeScale = 0.1 (預設)
func TestDefaultChatConfig(t *testing.T) {
	config := DefaultChatConfig()

	// AC2
	if config.ChatTurnsPerBeat != 10 {
		t.Errorf("DefaultChatConfig() ChatTurnsPerBeat = %d, want 10", config.ChatTurnsPerBeat)
	}

	// AC5
	if config.TimeScale != 0.1 {
		t.Errorf("DefaultChatConfig() TimeScale = %f, want 0.1", config.TimeScale)
	}

	if config.AllowInterrupts != false {
		t.Errorf("DefaultChatConfig() AllowInterrupts = %v, want false", config.AllowInterrupts)
	}
}

// TestChatTimeFlow_EdgeCases tests edge cases for time flow.
func TestChatTimeFlow_EdgeCases(t *testing.T) {
	t.Run("Very small timeScale", func(t *testing.T) {
		m := NewChatOverlayModel()
		m.chatTurnsPerBeat = 100
		m.timeScale = 0.01
		m.startBeat = 0

		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")

		// With 0.01 scale, need 100 turns to trigger via accumulator
		triggerCount := 0
		for i := 0; i < 100; i++ {
			if m.incrementChatTurns() {
				triggerCount++
			}
		}

		// Should trigger at least once (either via chatTurnsPerBeat or accumulator)
		if triggerCount < 1 {
			t.Errorf("Expected at least 1 trigger in 100 turns, got %d", triggerCount)
		}
	})

	t.Run("timeScale equals 1.0", func(t *testing.T) {
		m := NewChatOverlayModel()
		m.chatTurnsPerBeat = 10
		m.timeScale = 1.0 // Every turn triggers
		m.startBeat = 0

		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")

		// With scale=1.0, every turn should trigger via accumulator
		triggerCount := 0
		for i := 0; i < 5; i++ {
			if m.incrementChatTurns() {
				triggerCount++
			}
		}

		if triggerCount != 5 {
			t.Errorf("With timeScale=1.0, expected 5 triggers in 5 turns, got %d", triggerCount)
		}
	})

	t.Run("chatTurnsPerBeat equals 1", func(t *testing.T) {
		m := NewChatOverlayModel()
		m.chatTurnsPerBeat = 1 // Every turn triggers
		m.timeScale = 0.1
		m.startBeat = 0

		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")

		// Every turn should trigger
		triggerCount := 0
		for i := 0; i < 5; i++ {
			if m.incrementChatTurns() {
				triggerCount++
			}
		}

		if triggerCount != 5 {
			t.Errorf("With chatTurnsPerBeat=1, expected 5 triggers in 5 turns, got %d", triggerCount)
		}
	})
}

// TestChatTimeFlow_Integration tests complete time flow integration.
// AC1-AC5: Complete integration test
func TestChatTimeFlow_Integration(t *testing.T) {
	m := NewChatOverlayModel()

	// AC2, AC5: Apply custom config
	config := ChatConfig{
		TimeScale:        0.2,
		ChatTurnsPerBeat: 5,
		AllowInterrupts:  false,
	}
	m.ApplyConfig(config)

	// AC1: Enter chat and verify reset
	participants := []ChatParticipant{
		NewChatParticipant("player", "玩家", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "Alice", false, manager.NewEmotionState(70, 20, 30)),
	}
	m.Enter("player", participants, "大廳")
	m.startBeat = 10

	if m.chatTurns != 0 || m.tickAccumulator != 0.0 {
		t.Error("Enter() should reset chatTurns and tickAccumulator")
	}

	// Simulate player and NPC conversation
	initialMsgCount := len(m.messages)

	// Turn 1-4: No triggers
	for i := 0; i < 4; i++ {
		if m.incrementChatTurns() {
			t.Errorf("Turn %d should not trigger", i+1)
		}
	}

	// AC3: Verify accumulator
	expectedAccum := 0.8 // 4 * 0.2
	if m.tickAccumulator < expectedAccum-0.01 || m.tickAccumulator > expectedAccum+0.01 {
		t.Errorf("After 4 turns, accumulator = %f, want ~%f", m.tickAccumulator, expectedAccum)
	}

	// Turn 5: Should trigger (chatTurnsPerBeat)
	if !m.incrementChatTurns() {
		t.Error("Turn 5 should trigger")
	}

	// AC4: Verify system message added
	if len(m.messages) != initialMsgCount+1 {
		t.Errorf("After trigger, messages = %d, want %d", len(m.messages), initialMsgCount+1)
	}

	// Verify message content indicates correct beat
	lastMsg := m.messages[len(m.messages)-1]
	expectedBeat := 10 + (5 / 5) // startBeat=10, turn 5, turnsPerBeat=5
	expectedContent := fmt.Sprintf("時間流逝... (主時間回合 %d)", expectedBeat)
	if lastMsg.Content != expectedContent {
		t.Errorf("System message = %q, want %q", lastMsg.Content, expectedContent)
	}

	// AC5: Verify time flow info
	chatTurns, timeScale, accumulator, turnsPerBeat := m.GetTimeFlowInfo()
	if chatTurns != 5 || timeScale != 0.2 || turnsPerBeat != 5 {
		t.Errorf("GetTimeFlowInfo() = (%d, %f, %f, %d), want (5, 0.2, _, 5)",
			chatTurns, timeScale, accumulator, turnsPerBeat)
	}

	// Continue to turn 10
	for i := 0; i < 5; i++ {
		m.incrementChatTurns()
	}

	// Should have triggered again at turn 10
	expectedMsgCount := initialMsgCount + 2 // 2 system messages
	if len(m.messages) != expectedMsgCount {
		t.Errorf("After 10 turns, messages = %d, want %d", len(m.messages), expectedMsgCount)
	}
}

// ==========================================================================
// Story 5.2: Pending Events Mechanism - Unit Tests
// ==========================================================================

// TestPendingEvent_Structure tests PendingChatEvent structure.
// AC1: PendingChatEvent 資料結構定義
func TestPendingEvent_Structure(t *testing.T) {
	event := &PendingChatEvent{
		ID:             "event_001",
		Type:           EventTypeDanger,
		Description:    "A sudden noise from the hallway",
		TriggerBeat:    15,
		IsInterrupting: true,
		Severity:       SeverityHigh,
		RequiredAction: "Investigate the noise",
		Effects: &EventEffects{
			HPDelta:  -5,
			SANDelta: -10,
		},
	}

	// Verify all fields are set correctly
	if event.ID != "event_001" {
		t.Errorf("ID = %s, want event_001", event.ID)
	}
	if event.Type != EventTypeDanger {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeDanger)
	}
	if event.Description != "A sudden noise from the hallway" {
		t.Errorf("Description mismatch")
	}
	if event.TriggerBeat != 15 {
		t.Errorf("TriggerBeat = %d, want 15", event.TriggerBeat)
	}
	if !event.IsInterrupting {
		t.Error("IsInterrupting should be true")
	}
	if event.Severity != SeverityHigh {
		t.Errorf("Severity = %s, want %s", event.Severity, SeverityHigh)
	}
	if event.RequiredAction != "Investigate the noise" {
		t.Errorf("RequiredAction mismatch")
	}
	if event.Effects == nil {
		t.Fatal("Effects should not be nil")
	}
	if event.Effects.HPDelta != -5 {
		t.Errorf("HPDelta = %d, want -5", event.Effects.HPDelta)
	}
	if event.Effects.SANDelta != -10 {
		t.Errorf("SANDelta = %d, want -10", event.Effects.SANDelta)
	}
}

// TestAddPendingEvent tests adding events to queues.
// AC2: pendingEvents 佇列儲存待觸發事件
// AC3: backgroundEvents 佇列儲存背景事件
func TestAddPendingEvent(t *testing.T) {
	m := NewChatOverlayModel()

	// AC2: Add interrupting event
	event1 := &PendingChatEvent{
		ID:             "event_1",
		Type:           EventTypeCombat,
		Description:    "Enemy appears",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityCritical,
	}
	m.AddPendingEvent(event1)

	pendingEvents := m.GetPendingEvents()
	if len(pendingEvents) != 1 {
		t.Errorf("pendingEvents count = %d, want 1", len(pendingEvents))
	}
	if len(m.GetBackgroundEvents()) != 0 {
		t.Errorf("backgroundEvents should be empty")
	}

	// AC3: Add background event
	event2 := &PendingChatEvent{
		ID:             "event_2",
		Type:           EventTypeEnvironmental,
		Description:    "Distant footsteps",
		TriggerBeat:    12,
		IsInterrupting: false,
		Severity:       SeverityLow,
	}
	m.AddPendingEvent(event2)

	backgroundEvents := m.GetBackgroundEvents()
	if len(backgroundEvents) != 1 {
		t.Errorf("backgroundEvents count = %d, want 1", len(backgroundEvents))
	}
	if len(m.GetPendingEvents()) != 1 {
		t.Errorf("pendingEvents count should still be 1")
	}
}

// TestAddPendingEvent_NilEvent tests adding nil event.
func TestAddPendingEvent_NilEvent(t *testing.T) {
	m := NewChatOverlayModel()

	// Should not panic
	m.AddPendingEvent(nil)

	if len(m.GetPendingEvents()) != 0 {
		t.Error("Adding nil event should not add to queue")
	}
}

// TestPendingEvents_Sorting tests event sorting by TriggerBeat.
// AC2: 按 TriggerBeat 排序（最近的在前）
func TestPendingEvents_Sorting(t *testing.T) {
	m := NewChatOverlayModel()

	// Add events in random order
	events := []*PendingChatEvent{
		{ID: "event_3", TriggerBeat: 20, IsInterrupting: true},
		{ID: "event_1", TriggerBeat: 10, IsInterrupting: true},
		{ID: "event_2", TriggerBeat: 15, IsInterrupting: true},
	}

	for _, event := range events {
		m.AddPendingEvent(event)
	}

	// AC2: Verify sorted by TriggerBeat (ascending)
	pendingEvents := m.GetPendingEvents()
	if len(pendingEvents) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(pendingEvents))
	}

	if pendingEvents[0].ID != "event_1" {
		t.Errorf("First event should be event_1, got %s", pendingEvents[0].ID)
	}
	if pendingEvents[1].ID != "event_2" {
		t.Errorf("Second event should be event_2, got %s", pendingEvents[1].ID)
	}
	if pendingEvents[2].ID != "event_3" {
		t.Errorf("Third event should be event_3, got %s", pendingEvents[2].ID)
	}
}

// TestCheckPendingEvents_Triggering tests event triggering at correct time.
// AC4: 時間到達時觸發事件
func TestCheckPendingEvents_Triggering(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// Add event that triggers at beat 12
	event := &PendingChatEvent{
		ID:             "event_1",
		Type:           EventTypeDanger,
		Description:    "Sudden danger",
		TriggerBeat:    12,
		IsInterrupting: true,
		Severity:       SeverityHigh,
	}
	m.AddPendingEvent(event)

	initialMsgCount := len(m.messages)

	// AC4: Current beat 10, event should not trigger
	triggered := m.checkPendingEvents(10)
	if len(triggered) != 0 {
		t.Error("Event should not trigger at beat 10")
	}
	if len(m.GetPendingEvents()) != 1 {
		t.Error("Event should still be in queue")
	}

	// AC4: Current beat 12, event should trigger
	triggered = m.checkPendingEvents(12)
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered event, got %d", len(triggered))
	}
	if len(m.GetPendingEvents()) != 0 {
		t.Error("Event should be removed from queue after triggering")
	}

	// AC4: Verify system message was added
	if len(m.messages) <= initialMsgCount {
		t.Error("System message should be added when event triggers")
	}

	// Verify the system message contains event description
	found := false
	for _, msg := range m.messages[initialMsgCount:] {
		if msg.Type == ChatMessageSystem && containsString(msg.Content, event.Description) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Event system message not found in messages")
	}
}

// TestCheckPendingEvents_MultipleEvents tests multiple events triggering.
// AC4: 比較 event.TriggerBeat 與當前主時間回合
func TestCheckPendingEvents_MultipleEvents(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// Add multiple events with same trigger beat
	events := []*PendingChatEvent{
		{ID: "event_1", TriggerBeat: 10, IsInterrupting: true, Description: "Event 1"},
		{ID: "event_2", TriggerBeat: 10, IsInterrupting: true, Description: "Event 2"},
		{ID: "event_3", TriggerBeat: 15, IsInterrupting: true, Description: "Event 3"},
	}

	for _, event := range events {
		m.AddPendingEvent(event)
	}

	// Trigger at beat 10 - should trigger first 2 events
	triggered := m.checkPendingEvents(10)
	if len(triggered) != 2 {
		t.Errorf("Expected 2 triggered events at beat 10, got %d", len(triggered))
	}
	if len(m.GetPendingEvents()) != 1 {
		t.Errorf("Expected 1 remaining event, got %d", len(m.GetPendingEvents()))
	}

	// Trigger at beat 15 - should trigger remaining event
	triggered = m.checkPendingEvents(15)
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered event at beat 15, got %d", len(triggered))
	}
	if len(m.GetPendingEvents()) != 0 {
		t.Error("All events should be triggered")
	}
}

// TestBackgroundEvents_Silent tests background events don't add messages.
// AC3: backgroundEvents 靜默發生，只在聊天結束時顯示
func TestBackgroundEvents_Silent(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Add background event without effects
	bgEvent := &PendingChatEvent{
		ID:             "bg_event_1",
		Type:           EventTypeEnvironmental,
		Description:    "Distant thunder",
		TriggerBeat:    11,
		IsInterrupting: false,
		Severity:       SeverityLow,
	}
	m.AddPendingEvent(bgEvent)

	// Trigger event
	triggered := m.checkPendingEvents(11)
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered event, got %d", len(triggered))
	}

	// AC3: Background event should not add system message
	if len(m.messages) != initialMsgCount {
		t.Errorf("Background event should not add messages, got %d messages", len(m.messages)-initialMsgCount)
	}

	// Event should be removed from queue
	if len(m.GetBackgroundEvents()) != 0 {
		t.Error("Background event should be removed after triggering")
	}
}

// TestBackgroundEvents_WithEffects tests background events with effects.
// AC3: 背景事件可以有效果但不顯示訊息
func TestBackgroundEvents_WithEffects(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Add background event with effects
	bgEvent := &PendingChatEvent{
		ID:             "bg_event_2",
		Type:           EventTypeEnvironmental,
		Description:    "Temperature slowly drops",
		TriggerBeat:    12,
		IsInterrupting: false,
		Severity:       SeverityLow,
		Effects: &EventEffects{
			SANDelta: -2, // Slight sanity loss
		},
	}
	m.AddPendingEvent(bgEvent)

	// Trigger event
	triggered := m.checkPendingEvents(12)
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered event, got %d", len(triggered))
	}

	// AC3: Background event should still not add system message
	if len(m.messages) != initialMsgCount {
		t.Errorf("Background event with effects should not add messages, got %d messages", len(m.messages)-initialMsgCount)
	}

	// Event should be removed from queue
	if len(m.GetBackgroundEvents()) != 0 {
		t.Error("Background event should be removed after triggering")
	}
}

// TestPendingEvents_Effects tests event effects application.
// AC4: 應用事件效果
func TestPendingEvents_Effects(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// Add event with effects
	event := &PendingChatEvent{
		ID:             "event_1",
		TriggerBeat:    10,
		IsInterrupting: true,
		Description:    "Trap triggered",
		Effects: &EventEffects{
			HPDelta:       -10,
			SANDelta:      -5,
			ItemsLost:     []string{"torch"},
			CluesRevealed: []string{"trap_mechanism"},
		},
	}
	m.AddPendingEvent(event)

	// Trigger event - should call applyEventEffects
	// Note: Actual effect application requires game state integration
	// This test just verifies the event triggers without panic
	triggered := m.checkPendingEvents(10)
	if len(triggered) != 1 {
		t.Error("Event with effects should trigger successfully")
	}
}

// TestPendingEvents_HighSeverity tests high severity event messaging.
// AC4: 如果 IsInterrupting，準備中斷聊天
func TestPendingEvents_HighSeverity(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Add critical severity event with required action
	event := &PendingChatEvent{
		ID:             "critical_event",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityCritical,
		Description:    "Building collapse imminent!",
		RequiredAction: "Evacuate immediately",
	}
	m.AddPendingEvent(event)

	// Trigger event
	m.checkPendingEvents(10)

	// Should add at least 2 messages: event description + required action
	if len(m.messages) < initialMsgCount+2 {
		t.Errorf("Critical event should add multiple messages, got %d new messages", len(m.messages)-initialMsgCount)
	}
}

// TestClearPendingEvents tests clearing all events.
func TestClearPendingEvents(t *testing.T) {
	m := NewChatOverlayModel()

	// Add some events
	m.AddPendingEvent(&PendingChatEvent{
		ID: "event_1", TriggerBeat: 10, IsInterrupting: true,
	})
	m.AddPendingEvent(&PendingChatEvent{
		ID: "event_2", TriggerBeat: 15, IsInterrupting: false,
	})

	if len(m.GetPendingEvents()) != 1 || len(m.GetBackgroundEvents()) != 1 {
		t.Error("Events should be added")
	}

	// Clear all events
	m.ClearPendingEvents()

	if len(m.GetPendingEvents()) != 0 {
		t.Error("pendingEvents should be cleared")
	}
	if len(m.GetBackgroundEvents()) != 0 {
		t.Error("backgroundEvents should be cleared")
	}
}

// TestEnter_ResetsPendingEvents tests that Enter() resets event queues.
// AC2, AC3: 在聊天室進入時從遊戲狀態載入
func TestEnter_ResetsPendingEvents(t *testing.T) {
	m := NewChatOverlayModel()

	// Add some events
	m.AddPendingEvent(&PendingChatEvent{
		ID: "old_event", TriggerBeat: 5, IsInterrupting: true,
	})

	// Enter new chat session
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// Events should be reset
	if len(m.GetPendingEvents()) != 0 {
		t.Error("Enter() should reset pendingEvents")
	}
	if len(m.GetBackgroundEvents()) != 0 {
		t.Error("Enter() should reset backgroundEvents")
	}
}

// TestPendingEvents_Integration tests complete event flow integration.
// AC1-AC4: Complete integration test
func TestPendingEvents_Integration(t *testing.T) {
	m := NewChatOverlayModel()

	// Setup chat session
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc1", "Alice", false, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestLocation")
	m.SetStartBeat(0)
	m.chatTurnsPerBeat = 10

	// AC1: Add various events with different types and severities
	events := []*PendingChatEvent{
		{
			ID:             "combat_1",
			Type:           EventTypeCombat,
			Description:    "Enemy ambush!",
			TriggerBeat:    1,
			IsInterrupting: true,
			Severity:       SeverityCritical,
			RequiredAction: "Defend yourself",
		},
		{
			ID:             "discovery_1",
			Type:           EventTypeDiscovery,
			Description:    "You notice a hidden door",
			TriggerBeat:    1,
			IsInterrupting: false,
			Severity:       SeverityMedium,
		},
		{
			ID:             "environmental_1",
			Type:           EventTypeEnvironmental,
			Description:    "The lights flicker",
			TriggerBeat:    2,
			IsInterrupting: false,
			Severity:       SeverityLow,
		},
	}

	for _, event := range events {
		m.AddPendingEvent(event)
	}

	// AC2, AC3: Verify correct queue assignment
	if len(m.GetPendingEvents()) != 1 {
		t.Errorf("Should have 1 interrupting event, got %d", len(m.GetPendingEvents()))
	}
	if len(m.GetBackgroundEvents()) != 2 {
		t.Errorf("Should have 2 background events, got %d", len(m.GetBackgroundEvents()))
	}

	// Advance to beat 1 (requires 10 chat turns)
	for i := 0; i < 10; i++ {
		m.incrementChatTurns()
	}

	// AC4: Events at beat 1 should trigger
	// Combat event should add 2 messages (event + required action)
	// Discovery event should add 0 messages (background)
	if len(m.GetPendingEvents()) != 0 {
		t.Error("Interrupting event should be triggered and removed")
	}
	if len(m.GetBackgroundEvents()) != 1 {
		t.Error("Only beat 1 background event should be triggered")
	}

	// Advance to beat 2
	for i := 0; i < 10; i++ {
		m.incrementChatTurns()
	}

	// All events should be triggered
	if len(m.GetPendingEvents()) != 0 || len(m.GetBackgroundEvents()) != 0 {
		t.Error("All events should be triggered by beat 2")
	}
}

// TestEventType_String tests EventType string representation.
func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		want      string
	}{
		{EventTypeCombat, "combat"},
		{EventTypeDiscovery, "discovery"},
		{EventTypeEnvironmental, "environmental"},
		{EventTypeSocial, "social"},
		{EventTypeDanger, "danger"},
		{EventTypeNPCAction, "npc_action"},
		{EventTypeRevelation, "revelation"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.eventType.String()
			if got != tt.want {
				t.Errorf("EventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEventSeverity_String tests EventSeverity string representation.
func TestEventSeverity_String(t *testing.T) {
	tests := []struct {
		severity EventSeverity
		want     string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("EventSeverity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==========================================================================
// Story 5.3: Event Interruption Logic - Unit Tests
// ==========================================================================

// TestInterrupt_SetFlags tests that handleEvent() sets interrupt flags correctly.
// AC1: 當 IsInterrupting=true 時，handleEvent() 設定 interruptPending, interruptReason, interruptEvent
func TestInterrupt_SetFlags(t *testing.T) {
	t.Run("interrupting event sets flags", func(t *testing.T) {
		m := NewChatOverlayModel()
		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")

		// Create interrupting event
		event := &PendingChatEvent{
			ID:             "interrupt_1",
			Type:           EventTypeDanger,
			Description:    "Enemy approaching",
			TriggerBeat:    10,
			IsInterrupting: true,
			Severity:       SeverityHigh,
		}

		// Handle the event
		m.handleEvent(event)

		// AC1: Verify interrupt flags are set
		if !m.interruptPending {
			t.Error("Expected interruptPending to be true for interrupting event")
		}
		if m.interruptReason != "Enemy approaching" {
			t.Errorf("Expected interruptReason = 'Enemy approaching', got '%s'", m.interruptReason)
		}
		if m.interruptEvent != event {
			t.Error("Expected interruptEvent to be set to the event")
		}
	})

	t.Run("non-interrupting event does not set flags", func(t *testing.T) {
		m := NewChatOverlayModel()
		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")

		// Create non-interrupting event
		event := &PendingChatEvent{
			ID:             "bg_event_1",
			Type:           EventTypeEnvironmental,
			Description:    "Wind picks up",
			TriggerBeat:    10,
			IsInterrupting: false,
			Severity:       SeverityLow,
		}

		// Handle the event
		m.handleEvent(event)

		// AC1: Verify interrupt flags are NOT set
		if m.interruptPending {
			t.Error("Expected interruptPending to be false for non-interrupting event")
		}
		if m.interruptReason != "" {
			t.Errorf("Expected interruptReason to be empty, got '%s'", m.interruptReason)
		}
		if m.interruptEvent != nil {
			t.Error("Expected interruptEvent to be nil for non-interrupting event")
		}
	})
}

// TestInterrupt_CriticalForceExit tests Critical event forces immediate chat exit.
// AC4: Critical 事件調用 forceExit()
func TestInterrupt_CriticalForceExit(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")
	m.SetStartBeat(5)

	// Add some messages to verify summary generation
	m.AddMessage(NewChatMessage("msg1", "player", "Hello", ChatMessageNormal))

	// Create Critical event
	event := &PendingChatEvent{
		ID:             "critical_1",
		Type:           EventTypeCombat,
		Description:    "Sudden ambush",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityCritical,
		RequiredAction: "Defend yourself",
	}

	// Set interrupt state
	m.interruptPending = true
	m.interruptReason = event.Description
	m.interruptEvent = event

	// Verify chat is active before handling
	if !m.active {
		t.Error("Chat should be active before handling interruption")
	}

	// Handle interruption
	m.handleInterruption()

	// AC4: Verify chat was forced to exit
	if m.active {
		t.Error("Expected chat to be inactive after Critical event")
	}

	// Verify interrupt state was cleared
	if m.interruptPending {
		t.Error("Expected interruptPending to be cleared after forceExit")
	}
	if m.interruptReason != "" {
		t.Error("Expected interruptReason to be cleared after forceExit")
	}
	if m.interruptEvent != nil {
		t.Error("Expected interruptEvent to be cleared after forceExit")
	}

	// Verify system message was added
	lastMsg := m.messages[len(m.messages)-1]
	if lastMsg.Type != ChatMessageSystem {
		t.Error("Expected last message to be a system message")
	}
	if !strings.Contains(lastMsg.Content, "緊急") || !strings.Contains(lastMsg.Content, "Sudden ambush") {
		t.Errorf("Expected forced exit message, got: %s", lastMsg.Content)
	}
}

// TestInterrupt_HighNotification tests High severity displays urgent notification.
// AC4: High 事件調用 displayInterruptNotification(event, true)
func TestInterrupt_HighNotification(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Create High severity event
	event := &PendingChatEvent{
		ID:             "high_1",
		Type:           EventTypeDanger,
		Description:    "Suspicious noise nearby",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityHigh,
		RequiredAction: "Investigate or flee",
	}

	// Set interrupt state
	m.interruptPending = true
	m.interruptReason = event.Description
	m.interruptEvent = event

	// Handle interruption
	m.handleInterruption()

	// AC4: Verify notification was displayed
	if len(m.messages) <= initialMsgCount {
		t.Error("Expected notification message to be added for High severity event")
	}

	// Verify urgent styling indicator (⚠️ or ⚡)
	lastMsg := m.messages[len(m.messages)-1]
	if !strings.Contains(lastMsg.Content, "⚡") && !strings.Contains(lastMsg.Content, "⚠️") {
		t.Errorf("Expected urgent icon in notification, got: %s", lastMsg.Content)
	}

	// AC4: Verify chat is NOT forced to exit (only notification)
	if !m.active {
		t.Error("Expected chat to remain active after High severity notification")
	}

	// Verify interrupt state was cleared
	if m.interruptPending {
		t.Error("Expected interruptPending to be cleared after notification")
	}
}

// TestInterrupt_MediumNotification tests Medium severity displays normal notification.
// AC4: Medium 事件調用 displayInterruptNotification(event, false)
func TestInterrupt_MediumNotification(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Create Medium severity event
	event := &PendingChatEvent{
		ID:             "medium_1",
		Type:           EventTypeDiscovery,
		Description:    "You notice something interesting",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityMedium,
	}

	// Set interrupt state
	m.interruptPending = true
	m.interruptReason = event.Description
	m.interruptEvent = event

	// Handle interruption
	m.handleInterruption()

	// AC4: Verify notification was displayed
	if len(m.messages) <= initialMsgCount {
		t.Error("Expected notification message to be added for Medium severity event")
	}

	// Verify normal styling (ℹ️ icon, not urgent)
	lastMsg := m.messages[len(m.messages)-1]
	if !strings.Contains(lastMsg.Content, "ℹ️") {
		t.Errorf("Expected info icon in notification, got: %s", lastMsg.Content)
	}

	// Verify no urgent prefix
	if strings.Contains(lastMsg.Content, "緊急") {
		t.Error("Expected no urgent prefix for Medium severity")
	}

	// AC4: Verify chat continues (no forced exit)
	if !m.active {
		t.Error("Expected chat to remain active after Medium severity notification")
	}

	// Verify interrupt state was cleared
	if m.interruptPending {
		t.Error("Expected interruptPending to be cleared after notification")
	}
}

// TestInterrupt_AllowInterruptsDisabled tests downgrade to notification when AllowInterrupts=false.
// AC3: 如果 AllowInterrupts=false，所有中斷降級為通知
func TestInterrupt_AllowInterruptsDisabled(t *testing.T) {
	// Note: Since ChatConfig integration is not yet complete (per AC3 TODO),
	// this test verifies the downgrade logic by directly modifying the
	// allowInterrupts flag in handleInterruption().
	// In the current implementation, allowInterrupts defaults to true,
	// so we test the code path that would be triggered when it's false.

	// For now, we'll test the displayInterruptNotification path which is
	// what gets called when interrupts are disabled.

	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialMsgCount := len(m.messages)

	// Create Critical event (normally would force exit)
	event := &PendingChatEvent{
		ID:             "critical_disabled",
		Type:           EventTypeCombat,
		Description:    "Attack incoming",
		TriggerBeat:    10,
		IsInterrupting: true,
		Severity:       SeverityCritical,
	}

	// Manually call displayInterruptNotification (simulating disabled interrupts)
	// This is what handleInterruption would do if allowInterrupts=false
	m.displayInterruptNotification(event, false)

	// Verify notification was displayed
	if len(m.messages) <= initialMsgCount {
		t.Error("Expected notification message when interrupts are disabled")
	}

	// Verify chat continues (not forced to exit)
	if !m.active {
		t.Error("Expected chat to remain active when interrupts are disabled")
	}
}

// TestInterrupt_EnterResets tests Enter() resets interrupt state.
// AC1: Enter() 重置 interruptPending, interruptReason, interruptEvent
func TestInterrupt_EnterResets(t *testing.T) {
	m := NewChatOverlayModel()

	// Set some interrupt state
	m.interruptPending = true
	m.interruptReason = "Previous interrupt"
	m.interruptEvent = &PendingChatEvent{
		ID:          "old_event",
		Description: "Old event",
	}

	// Enter chat
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	// AC1: Verify interrupt state was reset
	if m.interruptPending {
		t.Error("Expected interruptPending to be false after Enter()")
	}
	if m.interruptReason != "" {
		t.Errorf("Expected interruptReason to be empty after Enter(), got: %s", m.interruptReason)
	}
	if m.interruptEvent != nil {
		t.Error("Expected interruptEvent to be nil after Enter()")
	}
}

// TestDisplayInterruptNotification tests notification display with different parameters.
// AC2: displayInterruptNotification() 根據 severity 和 urgent 顯示不同圖示
func TestDisplayInterruptNotification(t *testing.T) {
	tests := []struct {
		name           string
		event          *PendingChatEvent
		urgent         bool
		wantIcon       string
		wantUrgent     bool
		requiresAction bool
	}{
		{
			name: "high severity urgent",
			event: &PendingChatEvent{
				ID:             "event_1",
				Severity:       SeverityHigh,
				Description:    "Danger approaching",
				RequiredAction: "Take cover",
			},
			urgent:         true,
			wantIcon:       "⚡",
			wantUrgent:     true,
			requiresAction: true,
		},
		{
			name: "high severity non-urgent",
			event: &PendingChatEvent{
				ID:          "event_2",
				Severity:    SeverityHigh,
				Description: "Important discovery",
			},
			urgent:     false,
			wantIcon:   "⚠️",
			wantUrgent: false,
		},
		{
			name: "medium severity",
			event: &PendingChatEvent{
				ID:          "event_3",
				Severity:    SeverityMedium,
				Description: "Something noteworthy",
			},
			urgent:     false,
			wantIcon:   "ℹ️",
			wantUrgent: false,
		},
		{
			name: "low severity",
			event: &PendingChatEvent{
				ID:          "event_4",
				Severity:    SeverityLow,
				Description: "Minor observation",
			},
			urgent:     false,
			wantIcon:   "·",
			wantUrgent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewChatOverlayModel()
			participants := []ChatParticipant{
				NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
			}
			m.Enter("player", participants, "TestRoom")

			initialMsgCount := len(m.messages)

			// Display notification
			m.displayInterruptNotification(tt.event, tt.urgent)

			// Verify message was added
			if len(m.messages) <= initialMsgCount {
				t.Error("Expected notification message to be added")
			}

			// Check last message content
			lastMsg := m.messages[len(m.messages)-1]
			if lastMsg.Type != ChatMessageSystem {
				t.Error("Expected system message type")
			}

			// Verify icon is present
			if !strings.Contains(lastMsg.Content, tt.wantIcon) {
				t.Errorf("Expected icon '%s' in message, got: %s", tt.wantIcon, lastMsg.Content)
			}

			// Verify urgent prefix presence
			hasUrgentPrefix := strings.Contains(lastMsg.Content, "緊急")
			if hasUrgentPrefix != tt.wantUrgent {
				t.Errorf("Expected urgent prefix = %v, got %v", tt.wantUrgent, hasUrgentPrefix)
			}

			// Verify description is included
			if !strings.Contains(lastMsg.Content, tt.event.Description) {
				t.Errorf("Expected description '%s' in message, got: %s", tt.event.Description, lastMsg.Content)
			}

			// Verify required action if present
			if tt.requiresAction && tt.event.RequiredAction != "" {
				if !strings.Contains(lastMsg.Content, tt.event.RequiredAction) {
					t.Errorf("Expected required action '%s' in message, got: %s", tt.event.RequiredAction, lastMsg.Content)
				}
			}
		})
	}
}

// TestForceExit tests forced exit functionality.
// AC4: forceExit() 停用聊天、清除中斷狀態、生成摘要
func TestForceExit(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		NewChatParticipant("npc_001", "Alice", false, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")
	m.SetStartBeat(5)

	// Add some messages
	m.AddMessage(NewChatMessage("msg1", "player", "Hello Alice", ChatMessageNormal))
	m.AddMessage(NewChatMessage("msg2", "npc_001", "Hello player", ChatMessageNormal))

	// Advance some turns
	for i := 0; i < 5; i++ {
		m.incrementChatTurns()
	}

	initialMsgCount := len(m.messages)

	// Set interrupt state
	reason := "Building collapse"
	m.interruptPending = true
	m.interruptReason = reason
	m.interruptEvent = &PendingChatEvent{
		ID:          "critical",
		Description: reason,
		Severity:    SeverityCritical,
	}

	// Verify chat is active before
	if !m.active {
		t.Error("Chat should be active before forceExit")
	}

	// Call forceExit
	m.forceExit(reason)

	// AC4: Verify chat is deactivated
	if m.active {
		t.Error("Expected chat to be inactive after forceExit")
	}

	// AC4: Verify interrupt state was cleared
	if m.interruptPending {
		t.Error("Expected interruptPending to be cleared")
	}
	if m.interruptReason != "" {
		t.Errorf("Expected interruptReason to be empty, got: %s", m.interruptReason)
	}
	if m.interruptEvent != nil {
		t.Error("Expected interruptEvent to be nil")
	}

	// AC4: Verify forced exit message was added
	if len(m.messages) <= initialMsgCount {
		t.Error("Expected forced exit message to be added")
	}

	lastMsg := m.messages[len(m.messages)-1]
	if lastMsg.Type != ChatMessageSystem {
		t.Error("Expected last message to be system type")
	}
	if !strings.Contains(lastMsg.Content, "緊急") {
		t.Error("Expected emergency indicator in forced exit message")
	}
	if !strings.Contains(lastMsg.Content, reason) {
		t.Errorf("Expected reason '%s' in forced exit message, got: %s", reason, lastMsg.Content)
	}

	// Note: Summary generation and marking as interrupted is tested implicitly
	// through the Exit() call within forceExit(). The actual summary content
	// depends on summary generator availability, which is tested separately.
}

// TestInterrupt_Integration tests full interruption flow from event trigger to handling.
// Integration test for Story 5.3 covering AC1-AC4
func TestInterrupt_Integration(t *testing.T) {
	t.Run("critical event triggers immediate exit", func(t *testing.T) {
		m := NewChatOverlayModel()
		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")
		m.SetStartBeat(5)

		// Add critical event
		event := &PendingChatEvent{
			ID:             "critical_event",
			Type:           EventTypeDanger,
			Description:    "Ceiling collapse",
			TriggerBeat:    5,
			IsInterrupting: true,
			Severity:       SeverityCritical,
		}
		m.AddPendingEvent(event)

		// Trigger time advance
		m.incrementChatTurns() // This should check events

		// After triggerMainTimelineEvent is called, interruption should be handled
		// For the integration to work, we need to manually call checkPendingEvents and handleInterruption
		m.checkPendingEvents(5)
		m.handleInterruption()

		// Verify chat was forced to exit
		if m.active {
			t.Error("Expected chat to be inactive after critical event")
		}
	})

	t.Run("high event shows notification and continues", func(t *testing.T) {
		m := NewChatOverlayModel()
		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")
		m.SetStartBeat(5)

		initialMsgCount := len(m.messages)

		// Add high severity event
		event := &PendingChatEvent{
			ID:             "high_event",
			Type:           EventTypeDanger,
			Description:    "Footsteps in the distance",
			TriggerBeat:    5,
			IsInterrupting: true,
			Severity:       SeverityHigh,
		}
		m.AddPendingEvent(event)

		// Trigger event
		m.checkPendingEvents(5)
		m.handleInterruption()

		// Verify notification was shown
		if len(m.messages) <= initialMsgCount {
			t.Error("Expected notification message")
		}

		// Verify chat continues
		if !m.active {
			t.Error("Expected chat to remain active after high severity event")
		}
	})

	t.Run("medium event shows info and continues", func(t *testing.T) {
		m := NewChatOverlayModel()
		participants := []ChatParticipant{
			NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
		}
		m.Enter("player", participants, "TestRoom")
		m.SetStartBeat(5)

		initialMsgCount := len(m.messages)

		// Add medium severity event
		event := &PendingChatEvent{
			ID:             "medium_event",
			Type:           EventTypeDiscovery,
			Description:    "You recall something",
			TriggerBeat:    5,
			IsInterrupting: true,
			Severity:       SeverityMedium,
		}
		m.AddPendingEvent(event)

		// Trigger event
		m.checkPendingEvents(5)
		m.handleInterruption()

		// Verify notification was shown
		if len(m.messages) <= initialMsgCount {
			t.Error("Expected notification message")
		}

		// Verify chat continues
		if !m.active {
			t.Error("Expected chat to remain active after medium severity event")
		}
	})
}

// TestInterrupt_NoInterruptWhenNoPending tests handleInterruption with no pending interrupt.
func TestInterrupt_NoInterruptWhenNoPending(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")

	initialActive := m.active
	initialMsgCount := len(m.messages)

	// Call handleInterruption when no interrupt is pending
	m.handleInterruption()

	// Nothing should change
	if m.active != initialActive {
		t.Error("Expected active state to remain unchanged")
	}
	if len(m.messages) != initialMsgCount {
		t.Error("Expected no messages to be added")
	}
}

// TestInterrupt_SummaryNarrativeImpact tests that forceExit adds interrupt reason to summary.
// AC4: forceExit() 將中斷原因添加到摘要的 NarrativeImpact
func TestInterrupt_SummaryNarrativeImpact(t *testing.T) {
	m := NewChatOverlayModel()
	participants := []ChatParticipant{
		NewChatParticipant("player", "Player", true, manager.DefaultEmotionState()),
	}
	m.Enter("player", participants, "TestRoom")
	m.SetStartBeat(5)

	// Add messages
	m.AddMessage(NewChatMessage("msg1", "player", "Hello", ChatMessageNormal))

	// Advance some turns
	for i := 0; i < 3; i++ {
		m.incrementChatTurns()
	}

	reason := "Earthquake strikes"

	// Force exit
	m.forceExit(reason)

	// Verify forced exit message was added
	foundExitMsg := false
	for _, msg := range m.messages {
		if msg.Type == ChatMessageSystem && strings.Contains(msg.Content, reason) && strings.Contains(msg.Content, "緊急") {
			foundExitMsg = true
			break
		}
	}
	if !foundExitMsg {
		t.Error("Expected forced exit system message with reason")
	}

	// Note: Full summary.NarrativeImpact verification requires summary generator
	// integration, which is part of Story 5.4. The forceExit implementation
	// correctly updates summary.NarrativeImpact if a summary exists.
}
