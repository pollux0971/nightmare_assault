package views

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
)

// TestChatMessageType_String tests the String() method for all ChatMessageType values.
func TestChatMessageType_String(t *testing.T) {
	tests := []struct {
		name     string
		msgType  ChatMessageType
		expected string
	}{
		{"Normal", ChatMessageNormal, "normal"},
		{"System", ChatMessageSystem, "system"},
		{"Whisper", ChatMessageWhisper, "whisper"},
		{"Thought", ChatMessageThought, "thought"},
		{"Action", ChatMessageAction, "action"},
		{"Unknown", ChatMessageType(999), "unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msgType.String()
			if result != tt.expected {
				t.Errorf("ChatMessageType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseChatMessageType tests parsing strings to ChatMessageType.
func TestParseChatMessageType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ChatMessageType
		expectErr bool
	}{
		{"Normal lowercase", "normal", ChatMessageNormal, false},
		{"Normal uppercase", "NORMAL", ChatMessageNormal, false},
		{"Normal mixed case", "NoRmAl", ChatMessageNormal, false},
		{"System", "system", ChatMessageSystem, false},
		{"Whisper", "whisper", ChatMessageWhisper, false},
		{"Thought", "thought", ChatMessageThought, false},
		{"Action", "action", ChatMessageAction, false},
		{"With spaces", "  normal  ", ChatMessageNormal, false},
		{"Invalid type", "invalid", ChatMessageNormal, true},
		{"Empty string", "", ChatMessageNormal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseChatMessageType(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseChatMessageType(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseChatMessageType(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseChatMessageType(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestChatMessageType_JSON tests JSON marshaling/unmarshaling for ChatMessageType.
func TestChatMessageType_JSON(t *testing.T) {
	tests := []struct {
		name     string
		msgType  ChatMessageType
		expected string
	}{
		{"Normal", ChatMessageNormal, `"normal"`},
		{"System", ChatMessageSystem, `"system"`},
		{"Whisper", ChatMessageWhisper, `"whisper"`},
		{"Thought", ChatMessageThought, `"thought"`},
		{"Action", ChatMessageAction, `"action"`},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_Marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.msgType)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("json.Marshal() = %s, want %s", string(data), tt.expected)
			}
		})

		t.Run(tt.name+"_Unmarshal", func(t *testing.T) {
			var msgType ChatMessageType
			err := json.Unmarshal([]byte(tt.expected), &msgType)
			if err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}
			if msgType != tt.msgType {
				t.Errorf("json.Unmarshal() = %v, want %v", msgType, tt.msgType)
			}
		})
	}
}

// TestChatFlag_String tests the String() method for all ChatFlag values.
func TestChatFlag_String(t *testing.T) {
	tests := []struct {
		name     string
		flag     ChatFlag
		expected string
	}{
		{"Hallucination", ChatFlagHallucination, "hallucination"},
		{"Hostile", ChatFlagHostile, "hostile"},
		{"Revelation", ChatFlagRevelation, "revelation"},
		{"Persuasion", ChatFlagPersuasion, "persuasion"},
		{"Lie", ChatFlagLie, "lie"},
		{"Contradiction", ChatFlagContradiction, "contradiction"},
		{"Unknown", ChatFlag("unknown_test"), "unknown_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flag.String()
			if result != tt.expected {
				t.Errorf("ChatFlag.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseChatFlag tests parsing strings to ChatFlag.
func TestParseChatFlag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ChatFlag
		expectErr bool
	}{
		{"Hallucination lowercase", "hallucination", ChatFlagHallucination, false},
		{"Hallucination uppercase", "HALLUCINATION", ChatFlagHallucination, false},
		{"Hallucination mixed case", "HaLLuCiNaTioN", ChatFlagHallucination, false},
		{"Hostile", "hostile", ChatFlagHostile, false},
		{"Revelation", "revelation", ChatFlagRevelation, false},
		{"Persuasion", "persuasion", ChatFlagPersuasion, false},
		{"Lie", "lie", ChatFlagLie, false},
		{"Contradiction", "contradiction", ChatFlagContradiction, false},
		{"With spaces", "  hostile  ", ChatFlagHostile, false},
		{"Invalid flag", "invalid", ChatFlagHallucination, true},
		{"Empty string", "", ChatFlagHallucination, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseChatFlag(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseChatFlag(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseChatFlag(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseChatFlag(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestChatFlag_JSON tests JSON marshaling/unmarshaling for ChatFlag.
func TestChatFlag_JSON(t *testing.T) {
	tests := []struct {
		name     string
		flag     ChatFlag
		expected string
	}{
		{"Hallucination", ChatFlagHallucination, `"hallucination"`},
		{"Hostile", ChatFlagHostile, `"hostile"`},
		{"Revelation", ChatFlagRevelation, `"revelation"`},
		{"Persuasion", ChatFlagPersuasion, `"persuasion"`},
		{"Lie", ChatFlagLie, `"lie"`},
		{"Contradiction", ChatFlagContradiction, `"contradiction"`},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_Marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.flag)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("json.Marshal() = %s, want %s", string(data), tt.expected)
			}
		})

		t.Run(tt.name+"_Unmarshal", func(t *testing.T) {
			var flag ChatFlag
			err := json.Unmarshal([]byte(tt.expected), &flag)
			if err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}
			if flag != tt.flag {
				t.Errorf("json.Unmarshal() = %v, want %v", flag, tt.flag)
			}
		})
	}
}

// TestNewChatMessage tests creating a new ChatMessage.
func TestNewChatMessage(t *testing.T) {
	id := "msg-123"
	speaker := "player"
	content := "Hello, world!"
	msgType := ChatMessageNormal

	msg := NewChatMessage(id, speaker, content, msgType)

	if msg.ID != id {
		t.Errorf("ID = %v, want %v", msg.ID, id)
	}
	if msg.Speaker != speaker {
		t.Errorf("Speaker = %v, want %v", msg.Speaker, speaker)
	}
	if msg.Content != content {
		t.Errorf("Content = %v, want %v", msg.Content, content)
	}
	if msg.Type != msgType {
		t.Errorf("Type = %v, want %v", msg.Type, msgType)
	}
	if msg.Flags == nil {
		t.Error("Flags should be initialized, got nil")
	}
	if len(msg.Flags) != 0 {
		t.Errorf("Flags length = %v, want 0", len(msg.Flags))
	}
	if msg.EmotionEffects == nil {
		t.Error("EmotionEffects should be initialized, got nil")
	}
	if len(msg.EmotionEffects) != 0 {
		t.Errorf("EmotionEffects length = %v, want 0", len(msg.EmotionEffects))
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should be set, got zero time")
	}
}

// TestChatMessage_HasFlag tests the HasFlag method.
func TestChatMessage_HasFlag(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)

	// Initially no flags
	if msg.HasFlag(ChatFlagHostile) {
		t.Error("HasFlag(ChatFlagHostile) = true, want false")
	}

	// Add a flag
	msg.Flags = append(msg.Flags, ChatFlagHostile)
	if !msg.HasFlag(ChatFlagHostile) {
		t.Error("HasFlag(ChatFlagHostile) = false, want true")
	}

	// Check for non-existent flag
	if msg.HasFlag(ChatFlagLie) {
		t.Error("HasFlag(ChatFlagLie) = true, want false")
	}

	// Add multiple flags
	msg.Flags = append(msg.Flags, ChatFlagRevelation, ChatFlagPersuasion)
	if !msg.HasFlag(ChatFlagRevelation) {
		t.Error("HasFlag(ChatFlagRevelation) = false, want true")
	}
	if !msg.HasFlag(ChatFlagPersuasion) {
		t.Error("HasFlag(ChatFlagPersuasion) = false, want true")
	}
}

// TestChatMessage_AddFlag tests the AddFlag method.
func TestChatMessage_AddFlag(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)

	// Add first flag
	msg.AddFlag(ChatFlagHostile)
	if !msg.HasFlag(ChatFlagHostile) {
		t.Error("Flag should be added")
	}
	if len(msg.Flags) != 1 {
		t.Errorf("Flags length = %v, want 1", len(msg.Flags))
	}

	// Add duplicate flag (should not add)
	msg.AddFlag(ChatFlagHostile)
	if len(msg.Flags) != 1 {
		t.Errorf("Flags length = %v, want 1 (duplicate should not be added)", len(msg.Flags))
	}

	// Add different flag
	msg.AddFlag(ChatFlagRevelation)
	if !msg.HasFlag(ChatFlagRevelation) {
		t.Error("Flag should be added")
	}
	if len(msg.Flags) != 2 {
		t.Errorf("Flags length = %v, want 2", len(msg.Flags))
	}
}

// TestChatMessage_RemoveFlag tests the RemoveFlag method.
func TestChatMessage_RemoveFlag(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)
	msg.Flags = []ChatFlag{ChatFlagHostile, ChatFlagRevelation, ChatFlagPersuasion}

	// Remove existing flag
	msg.RemoveFlag(ChatFlagRevelation)
	if msg.HasFlag(ChatFlagRevelation) {
		t.Error("Flag should be removed")
	}
	if len(msg.Flags) != 2 {
		t.Errorf("Flags length = %v, want 2", len(msg.Flags))
	}

	// Remove non-existent flag (should not error)
	msg.RemoveFlag(ChatFlagLie)
	if len(msg.Flags) != 2 {
		t.Errorf("Flags length = %v, want 2 (removing non-existent flag should not change length)", len(msg.Flags))
	}

	// Remove all flags
	msg.RemoveFlag(ChatFlagHostile)
	msg.RemoveFlag(ChatFlagPersuasion)
	if len(msg.Flags) != 0 {
		t.Errorf("Flags length = %v, want 0", len(msg.Flags))
	}
}

// TestChatMessage_GetEmotionEffect tests the GetEmotionEffect method.
func TestChatMessage_GetEmotionEffect(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)

	// Get from empty map
	effect := msg.GetEmotionEffect("npc-1")
	if effect != nil {
		t.Error("GetEmotionEffect should return nil for non-existent NPC")
	}

	// Add emotion effect
	delta := &manager.EmotionDelta{Trust: 10, Fear: -5, Stress: 0}
	msg.EmotionEffects["npc-1"] = delta

	// Get existing effect
	effect = msg.GetEmotionEffect("npc-1")
	if effect == nil {
		t.Fatal("GetEmotionEffect should return non-nil for existing NPC")
	}
	if effect.Trust != 10 || effect.Fear != -5 || effect.Stress != 0 {
		t.Errorf("GetEmotionEffect returned incorrect delta: %+v", effect)
	}

	// Get non-existent NPC
	effect = msg.GetEmotionEffect("npc-2")
	if effect != nil {
		t.Error("GetEmotionEffect should return nil for non-existent NPC")
	}
}

// TestChatMessage_SetEmotionEffect tests the SetEmotionEffect method.
func TestChatMessage_SetEmotionEffect(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)

	// Set first effect
	delta1 := &manager.EmotionDelta{Trust: 10, Fear: -5, Stress: 0}
	msg.SetEmotionEffect("npc-1", delta1)

	effect := msg.GetEmotionEffect("npc-1")
	if effect == nil {
		t.Fatal("Effect should be set")
	}
	if effect.Trust != 10 {
		t.Errorf("Trust = %v, want 10", effect.Trust)
	}

	// Set second effect
	delta2 := &manager.EmotionDelta{Trust: -20, Fear: 15, Stress: 10}
	msg.SetEmotionEffect("npc-2", delta2)

	if len(msg.EmotionEffects) != 2 {
		t.Errorf("EmotionEffects length = %v, want 2", len(msg.EmotionEffects))
	}

	// Overwrite existing effect
	delta3 := &manager.EmotionDelta{Trust: 5, Fear: 0, Stress: -5}
	msg.SetEmotionEffect("npc-1", delta3)

	effect = msg.GetEmotionEffect("npc-1")
	if effect.Trust != 5 {
		t.Errorf("Trust = %v, want 5 (should be overwritten)", effect.Trust)
	}
}

// TestChatMessage_Copy tests the Copy method.
func TestChatMessage_Copy(t *testing.T) {
	original := NewChatMessage("msg-1", "player", "test content", ChatMessageNormal)
	original.Flags = []ChatFlag{ChatFlagHostile, ChatFlagRevelation}
	delta1 := &manager.EmotionDelta{Trust: 10, Fear: -5, Stress: 0}
	delta2 := &manager.EmotionDelta{Trust: -20, Fear: 15, Stress: 10}
	original.SetEmotionEffect("npc-1", delta1)
	original.SetEmotionEffect("npc-2", delta2)

	// Create copy
	copy := original.Copy()

	// Verify copy has same values
	if copy.ID != original.ID {
		t.Errorf("Copy ID = %v, want %v", copy.ID, original.ID)
	}
	if copy.Speaker != original.Speaker {
		t.Errorf("Copy Speaker = %v, want %v", copy.Speaker, original.Speaker)
	}
	if copy.Content != original.Content {
		t.Errorf("Copy Content = %v, want %v", copy.Content, original.Content)
	}
	if copy.Type != original.Type {
		t.Errorf("Copy Type = %v, want %v", copy.Type, original.Type)
	}
	if !copy.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Copy Timestamp = %v, want %v", copy.Timestamp, original.Timestamp)
	}

	// Verify flags are copied
	if len(copy.Flags) != len(original.Flags) {
		t.Errorf("Copy Flags length = %v, want %v", len(copy.Flags), len(original.Flags))
	}

	// Verify emotion effects are copied
	if len(copy.EmotionEffects) != len(original.EmotionEffects) {
		t.Errorf("Copy EmotionEffects length = %v, want %v", len(copy.EmotionEffects), len(original.EmotionEffects))
	}

	// Verify deep copy - modifying copy shouldn't affect original
	copy.AddFlag(ChatFlagLie)
	if original.HasFlag(ChatFlagLie) {
		t.Error("Adding flag to copy should not affect original")
	}

	copy.SetEmotionEffect("npc-3", &manager.EmotionDelta{Trust: 100, Fear: 0, Stress: 0})
	if original.GetEmotionEffect("npc-3") != nil {
		t.Error("Setting emotion effect on copy should not affect original")
	}

	// Modify original emotion effect and verify copy is unaffected
	original.EmotionEffects["npc-1"].Trust = 999
	copyEffect := copy.GetEmotionEffect("npc-1")
	if copyEffect.Trust == 999 {
		t.Error("Modifying original emotion effect should not affect copy")
	}
}

// TestChatMessage_JSON tests JSON serialization/deserialization of ChatMessage.
func TestChatMessage_JSON(t *testing.T) {
	original := NewChatMessage("msg-123", "player", "Hello, world!", ChatMessageNormal)
	original.AddFlag(ChatFlagHostile)
	original.AddFlag(ChatFlagRevelation)
	original.SetEmotionEffect("npc-1", &manager.EmotionDelta{Trust: 10, Fear: -5, Stress: 0})

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal from JSON
	var decoded ChatMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify fields
	if decoded.ID != original.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, original.ID)
	}
	if decoded.Speaker != original.Speaker {
		t.Errorf("Speaker = %v, want %v", decoded.Speaker, original.Speaker)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content = %v, want %v", decoded.Content, original.Content)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, original.Type)
	}

	// Verify flags
	if len(decoded.Flags) != len(original.Flags) {
		t.Errorf("Flags length = %v, want %v", len(decoded.Flags), len(original.Flags))
	}
	if !decoded.HasFlag(ChatFlagHostile) {
		t.Error("Decoded message should have ChatFlagHostile")
	}
	if !decoded.HasFlag(ChatFlagRevelation) {
		t.Error("Decoded message should have ChatFlagRevelation")
	}

	// Verify emotion effects
	if len(decoded.EmotionEffects) != len(original.EmotionEffects) {
		t.Errorf("EmotionEffects length = %v, want %v", len(decoded.EmotionEffects), len(original.EmotionEffects))
	}
	effect := decoded.GetEmotionEffect("npc-1")
	if effect == nil {
		t.Fatal("Decoded message should have emotion effect for npc-1")
	}
	if effect.Trust != 10 || effect.Fear != -5 || effect.Stress != 0 {
		t.Errorf("Decoded emotion effect = %+v, want Trust:10 Fear:-5 Stress:0", effect)
	}
}

// TestChatMessage_EmptyFlagsAndEffects tests handling of nil/empty flags and effects.
func TestChatMessage_EmptyFlagsAndEffects(t *testing.T) {
	msg := &ChatMessage{
		ID:             "msg-1",
		Speaker:        "player",
		Content:        "test",
		Timestamp:      time.Now(),
		Type:           ChatMessageNormal,
		Flags:          nil,
		EmotionEffects: nil,
	}

	// HasFlag with nil Flags
	if msg.HasFlag(ChatFlagHostile) {
		t.Error("HasFlag should return false for nil Flags")
	}

	// GetEmotionEffect with nil EmotionEffects
	effect := msg.GetEmotionEffect("npc-1")
	if effect != nil {
		t.Error("GetEmotionEffect should return nil for nil EmotionEffects")
	}

	// SetEmotionEffect should initialize nil map
	msg.SetEmotionEffect("npc-1", &manager.EmotionDelta{Trust: 10})
	if msg.EmotionEffects == nil {
		t.Error("SetEmotionEffect should initialize EmotionEffects map")
	}
	if len(msg.EmotionEffects) != 1 {
		t.Errorf("EmotionEffects length = %v, want 1", len(msg.EmotionEffects))
	}
}

// TestChatMessage_FlagCombinations tests various flag combinations.
func TestChatMessage_FlagCombinations(t *testing.T) {
	msg := NewChatMessage("msg-1", "player", "test", ChatMessageNormal)

	// Add all possible flags
	allFlags := []ChatFlag{
		ChatFlagHallucination,
		ChatFlagHostile,
		ChatFlagRevelation,
		ChatFlagPersuasion,
		ChatFlagLie,
		ChatFlagContradiction,
	}

	for _, flag := range allFlags {
		msg.AddFlag(flag)
	}

	// Verify all flags are present
	for _, flag := range allFlags {
		if !msg.HasFlag(flag) {
			t.Errorf("Message should have flag %v", flag)
		}
	}

	// Remove half the flags
	msg.RemoveFlag(ChatFlagHallucination)
	msg.RemoveFlag(ChatFlagRevelation)
	msg.RemoveFlag(ChatFlagLie)

	// Verify correct flags remain
	if msg.HasFlag(ChatFlagHallucination) {
		t.Error("ChatFlagHallucination should be removed")
	}
	if !msg.HasFlag(ChatFlagHostile) {
		t.Error("ChatFlagHostile should still be present")
	}
	if msg.HasFlag(ChatFlagRevelation) {
		t.Error("ChatFlagRevelation should be removed")
	}
	if !msg.HasFlag(ChatFlagPersuasion) {
		t.Error("ChatFlagPersuasion should still be present")
	}
	if msg.HasFlag(ChatFlagLie) {
		t.Error("ChatFlagLie should be removed")
	}
	if !msg.HasFlag(ChatFlagContradiction) {
		t.Error("ChatFlagContradiction should still be present")
	}
}

// ==========================================================================
// Story 5.1: Chat Time Flow Control - ChatConfig Tests
// ==========================================================================

// TestChatConfig_DefaultValues tests default ChatConfig values.
// AC2: chatTurnsPerBeat = 10 (預設)
// AC5: timeScale = 0.1 (預設)
func TestChatConfig_DefaultValues(t *testing.T) {
	config := DefaultChatConfig()

	if config.TimeScale != 0.1 {
		t.Errorf("DefaultChatConfig() TimeScale = %f, want 0.1", config.TimeScale)
	}

	if config.ChatTurnsPerBeat != 10 {
		t.Errorf("DefaultChatConfig() ChatTurnsPerBeat = %d, want 10", config.ChatTurnsPerBeat)
	}

	if config.AllowInterrupts != false {
		t.Errorf("DefaultChatConfig() AllowInterrupts = %v, want false", config.AllowInterrupts)
	}
}

// TestChatConfig_CustomValues tests creating ChatConfig with custom values.
func TestChatConfig_CustomValues(t *testing.T) {
	config := ChatConfig{
		TimeScale:        0.05,
		ChatTurnsPerBeat: 20,
		AllowInterrupts:  true,
	}

	if config.TimeScale != 0.05 {
		t.Errorf("ChatConfig TimeScale = %f, want 0.05", config.TimeScale)
	}

	if config.ChatTurnsPerBeat != 20 {
		t.Errorf("ChatConfig ChatTurnsPerBeat = %d, want 20", config.ChatTurnsPerBeat)
	}

	if config.AllowInterrupts != true {
		t.Errorf("ChatConfig AllowInterrupts = %v, want true", config.AllowInterrupts)
	}
}

// ==========================================================================
// Story 5.6 AC3: ChatSession JSON Serialization Tests
// ==========================================================================

// TestChatSession_JSONSerialization_Complete tests full ChatSession serialization.
// Story 5.6 AC3: Test complete session with all fields.
func TestChatSession_JSONSerialization_Complete(t *testing.T) {
	// Create a complete chat session with all fields populated
	now := time.Now()
	original := &ChatSession{
		ID:           "session_001",
		Participants: []string{"player", "npc_001", "npc_002"},
		Messages: []*ChatMessage{
			{
				ID:        "msg_001",
				Speaker:   "player",
				Content:   "你好，你們是誰？",
				Timestamp: now,
				Type:      ChatMessageNormal,
				Flags:     []ChatFlag{},
				EmotionEffects: map[string]*manager.EmotionDelta{
					"npc_001": {Trust: 5, Fear: 0, Stress: 0},
				},
			},
			{
				ID:        "msg_002",
				Speaker:   "npc_001",
				Content:   "我是醫生，你不記得了嗎？",
				Timestamp: now.Add(time.Second),
				Type:      ChatMessageNormal,
				Flags:     []ChatFlag{ChatFlagRevelation},
				EmotionEffects: make(map[string]*manager.EmotionDelta),
			},
		},
		StartBeat: 10,
		EndBeat:   12,
		Summary: &ChatSummary{
			MainTopics:      []string{"初次見面", "身份確認"},
			KeyDecisions:    []string{"決定信任醫生"},
			RelationChanges: map[string]string{
				"npc_001": "信任增加(+10), 恐懼減少(-5)",
			},
			FactsShared:     []string{"醫生聲稱認識玩家"},
			Flags:           []string{"revelation"},
			NarrativeImpact: "建立了與醫生的初步信任關係",
		},
		Interrupted:     false,
		InterruptReason: "",
		CreatedAt:       now,

		// Legacy fields
		SessionID:    "session_001",
		Initiator:    "player",
		ParticipantDetails: []ChatParticipant{
			{ID: "player", Name: "玩家", IsPlayer: true, IsActive: true},
			{ID: "npc_001", Name: "醫生", IsPlayer: false, IsActive: true},
		},
		StartTime:  now,
		EndTime:    now.Add(2 * time.Minute),
		Location:   "病房",
		TurnsSpent: 20,
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialized data should not be empty")
	}

	// Deserialize from JSON
	var restored ChatSession
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify core fields
	if restored.ID != original.ID {
		t.Errorf("ID = %v, want %v", restored.ID, original.ID)
	}
	if len(restored.Participants) != len(original.Participants) {
		t.Errorf("Participants length = %v, want %v", len(restored.Participants), len(original.Participants))
	}
	if restored.StartBeat != original.StartBeat {
		t.Errorf("StartBeat = %v, want %v", restored.StartBeat, original.StartBeat)
	}
	if restored.EndBeat != original.EndBeat {
		t.Errorf("EndBeat = %v, want %v", restored.EndBeat, original.EndBeat)
	}
	if restored.Interrupted != original.Interrupted {
		t.Errorf("Interrupted = %v, want %v", restored.Interrupted, original.Interrupted)
	}

	// Verify messages
	if len(restored.Messages) != 2 {
		t.Fatalf("Messages length = %v, want 2", len(restored.Messages))
	}
	if restored.Messages[0].ID != original.Messages[0].ID {
		t.Errorf("Message[0].ID = %v, want %v", restored.Messages[0].ID, original.Messages[0].ID)
	}
	if restored.Messages[0].Content != original.Messages[0].Content {
		t.Errorf("Message[0].Content = %v, want %v", restored.Messages[0].Content, original.Messages[0].Content)
	}

	// Verify message flags
	if len(restored.Messages[1].Flags) != 1 {
		t.Errorf("Message[1].Flags length = %v, want 1", len(restored.Messages[1].Flags))
	}
	if len(restored.Messages[1].Flags) > 0 && restored.Messages[1].Flags[0] != ChatFlagRevelation {
		t.Errorf("Message[1].Flags[0] = %v, want %v", restored.Messages[1].Flags[0], ChatFlagRevelation)
	}

	// Verify summary
	if restored.Summary == nil {
		t.Fatal("Summary should not be nil")
	}
	if len(restored.Summary.MainTopics) != len(original.Summary.MainTopics) {
		t.Errorf("Summary.MainTopics length = %v, want %v", len(restored.Summary.MainTopics), len(original.Summary.MainTopics))
	}
	if restored.Summary.NarrativeImpact != original.Summary.NarrativeImpact {
		t.Errorf("Summary.NarrativeImpact = %v, want %v", restored.Summary.NarrativeImpact, original.Summary.NarrativeImpact)
	}
}

// TestChatSession_JSONSerialization_Partial tests partial ChatSession serialization.
// Story 5.6 AC3: Test session without optional fields (no summary, no interruption).
func TestChatSession_JSONSerialization_Partial(t *testing.T) {
	now := time.Now()
	original := &ChatSession{
		ID:           "session_002",
		Participants: []string{"player", "npc_003"},
		Messages: []*ChatMessage{
			NewChatMessage("msg_001", "player", "快點跑！", ChatMessageNormal),
		},
		StartBeat:       15,
		EndBeat:         16,
		Summary:         nil, // No summary
		Interrupted:     false,
		InterruptReason: "",
		CreatedAt:       now,
	}

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Deserialize
	var restored ChatSession
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify
	if restored.ID != original.ID {
		t.Errorf("ID = %v, want %v", restored.ID, original.ID)
	}
	if restored.StartBeat != original.StartBeat {
		t.Errorf("StartBeat = %v, want %v", restored.StartBeat, original.StartBeat)
	}
	if restored.EndBeat != original.EndBeat {
		t.Errorf("EndBeat = %v, want %v", restored.EndBeat, original.EndBeat)
	}
	if restored.Summary != nil {
		t.Error("Summary should be nil")
	}
	if restored.Interrupted != false {
		t.Error("Interrupted should be false")
	}
	if restored.InterruptReason != "" {
		t.Errorf("InterruptReason = %v, want empty string", restored.InterruptReason)
	}
}

// TestChatSession_JSONSerialization_Interrupted tests interrupted session serialization.
// Story 5.6 AC3: Test session that was interrupted with reason.
func TestChatSession_JSONSerialization_Interrupted(t *testing.T) {
	now := time.Now()
	original := &ChatSession{
		ID:              "session_003",
		Participants:    []string{"player", "npc_004"},
		Messages:        []*ChatMessage{},
		StartBeat:       20,
		EndBeat:         20,
		Summary:         nil,
		Interrupted:     true,
		InterruptReason: "怪物突然出現",
		CreatedAt:       now,
	}

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Deserialize
	var restored ChatSession
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify interruption fields
	if !restored.Interrupted {
		t.Error("Interrupted should be true")
	}
	if restored.InterruptReason != "怪物突然出現" {
		t.Errorf("InterruptReason = %v, want '怪物突然出現'", restored.InterruptReason)
	}
}

// TestChatSession_JSONSerialization_RoundTrip tests round-trip consistency.
// Story 5.6 AC3: Test JSON round-trip (serialize -> deserialize -> serialize).
func TestChatSession_JSONSerialization_RoundTrip(t *testing.T) {
	now := time.Now()
	original := &ChatSession{
		ID:           "session_004",
		Participants: []string{"player", "npc_005"},
		Messages: []*ChatMessage{
			NewChatMessage("msg_001", "player", "測試訊息", ChatMessageNormal),
		},
		StartBeat: 30,
		EndBeat:   32,
		Summary: &ChatSummary{
			MainTopics:      []string{"測試"},
			NarrativeImpact: "無影響",
		},
		Interrupted: false,
		CreatedAt:   now,
	}

	// First serialization
	data1, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("First serialization failed: %v", err)
	}

	// First deserialization
	var intermediate ChatSession
	err = json.Unmarshal(data1, &intermediate)
	if err != nil {
		t.Fatalf("First deserialization failed: %v", err)
	}

	// Second serialization
	data2, err := json.Marshal(&intermediate)
	if err != nil {
		t.Fatalf("Second serialization failed: %v", err)
	}

	// Second deserialization
	var final ChatSession
	err = json.Unmarshal(data2, &final)
	if err != nil {
		t.Fatalf("Second deserialization failed: %v", err)
	}

	// Verify consistency
	if final.ID != original.ID {
		t.Errorf("ID consistency check failed: %v != %v", final.ID, original.ID)
	}
	if final.StartBeat != original.StartBeat {
		t.Errorf("StartBeat consistency check failed: %v != %v", final.StartBeat, original.StartBeat)
	}
	if final.EndBeat != original.EndBeat {
		t.Errorf("EndBeat consistency check failed: %v != %v", final.EndBeat, original.EndBeat)
	}
}

// TestChatSession_EmptyMessages tests session with no messages.
func TestChatSession_EmptyMessages(t *testing.T) {
	original := &ChatSession{
		ID:           "session_empty",
		Participants: []string{"player"},
		Messages:     []*ChatMessage{},
		StartBeat:    5,
		EndBeat:      5,
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	var restored ChatSession
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID = %v, want %v", restored.ID, original.ID)
	}
	if restored.Messages == nil {
		t.Error("Messages should not be nil")
	}
	if len(restored.Messages) != 0 {
		t.Errorf("Messages length = %v, want 0", len(restored.Messages))
	}
}
