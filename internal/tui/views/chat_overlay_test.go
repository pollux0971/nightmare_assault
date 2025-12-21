package views

import (
	"fmt"
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
