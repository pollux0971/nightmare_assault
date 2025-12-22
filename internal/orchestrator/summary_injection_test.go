package orchestrator

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// ==========================================================================
// Story 5.5: Summary Injection Tests
// ==========================================================================

// TestFormatSummaryForNarration tests the summary formatting logic.
// Story 5.5 AC1, AC2: Format summary as natural language for NarrationAgent
func TestFormatSummaryForNarration(t *testing.T) {
	orch := NewOrchestrator()

	// Setup NPC profiles for name lookup
	orch.storyBible.NPCProfiles = []*NPCProfile{
		{
			ID:   "npc_001",
			Name: "張醫生",
		},
		{
			ID:   "npc_002",
			Name: "護士王",
		},
	}

	t.Run("complete summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_001",
			Participants: []string{"player", "npc_001"},
			StartBeat:    10,
			EndBeat:      15,
			Summary: &ChatSummary{
				MainTopics:   []string{"醫療物資", "避難所位置"},
				KeyDecisions: []string{"同意分享物資", "決定一起前往東翼"},
				RelationChanges: map[string]string{
					"npc_001": "信任增加",
				},
				FactsShared:     []string{"東翼有醫療物資", "避難所在地下室"},
				NarrativeImpact: "建立了合作關係，獲得了關鍵線索",
				EmotionChanges: map[string]string{
					"npc_001": "恐懼減少，信任增加",
				},
			},
		}

		formatted := orch.formatSummaryForNarration(session)

		// Verify formatted output contains expected elements
		if formatted == "" {
			t.Error("formatSummaryForNarration returned empty string")
		}

		// Check for dialogue record metadata
		assertContains(t, formatted, "對話記錄", "should contain dialogue record header")
		assertContains(t, formatted, "10-15", "should contain beat range")

		// Check for participants
		assertContains(t, formatted, "參與者", "should contain participants label")
		assertContains(t, formatted, "玩家", "should contain player name")
		assertContains(t, formatted, "張醫生", "should contain NPC name")

		// Check for main topics
		assertContains(t, formatted, "討論話題", "should contain topics label")
		assertContains(t, formatted, "醫療物資", "should contain topic 1")
		assertContains(t, formatted, "避難所位置", "should contain topic 2")

		// Check for key decisions
		assertContains(t, formatted, "關鍵決策", "should contain decisions label")
		assertContains(t, formatted, "同意分享物資", "should contain decision 1")
		assertContains(t, formatted, "決定一起前往東翼", "should contain decision 2")

		// Check for relationship changes
		assertContains(t, formatted, "關係變化", "should contain relationship label")
		assertContains(t, formatted, "信任增加", "should contain relationship change")

		// Check for facts shared
		assertContains(t, formatted, "分享資訊", "should contain facts label")
		assertContains(t, formatted, "東翼有醫療物資", "should contain fact 1")
		assertContains(t, formatted, "避難所在地下室", "should contain fact 2")

		// Check for narrative impact
		assertContains(t, formatted, "影響", "should contain impact label")
		assertContains(t, formatted, "建立了合作關係", "should contain narrative impact")
	})

	t.Run("minimal summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_002",
			Participants: []string{"player", "npc_002"},
			StartBeat:    5,
			EndBeat:      6,
			Summary: &ChatSummary{
				MainTopics:      []string{"打招呼"},
				NarrativeImpact: "初次接觸",
			},
		}

		formatted := orch.formatSummaryForNarration(session)

		if formatted == "" {
			t.Error("formatSummaryForNarration returned empty string for minimal summary")
		}

		assertContains(t, formatted, "對話記錄", "should contain dialogue record header")
		assertContains(t, formatted, "打招呼", "should contain main topic")
		assertContains(t, formatted, "初次接觸", "should contain narrative impact")
	})

	t.Run("nil session", func(t *testing.T) {
		formatted := orch.formatSummaryForNarration(nil)
		if formatted != "" {
			t.Errorf("Expected empty string for nil session, got: %s", formatted)
		}
	})

	t.Run("nil summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_003",
			Participants: []string{"player"},
			Summary:      nil,
		}

		formatted := orch.formatSummaryForNarration(session)
		if formatted != "" {
			t.Errorf("Expected empty string for nil summary, got: %s", formatted)
		}
	})

	t.Run("wrong summary type", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_004",
			Participants: []string{"player"},
			Summary:      "not a ChatSummary", // Wrong type
		}

		formatted := orch.formatSummaryForNarration(session)
		if formatted != "" {
			t.Errorf("Expected empty string for wrong summary type, got: %s", formatted)
		}
	})

	t.Run("unknown participant name fallback", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_005",
			Participants: []string{"player", "npc_999"}, // Unknown NPC
			StartBeat:    1,
			EndBeat:      2,
			Summary: &ChatSummary{
				MainTopics: []string{"測試"},
			},
		}

		formatted := orch.formatSummaryForNarration(session)

		assertContains(t, formatted, "玩家", "should contain player name")
		assertContains(t, formatted, "npc_999", "should fallback to ID for unknown NPC")
	})
}

// TestInjectChatSummary tests the complete summary injection flow.
// Story 5.5 AC1: ChatSummary 注入到 NarrationAgent context
func TestInjectChatSummary(t *testing.T) {
	orch := NewOrchestrator()

	// Setup NPC profiles
	orch.storyBible.NPCProfiles = []*NPCProfile{
		{
			ID:   "npc_001",
			Name: "張醫生",
		},
	}

	t.Run("successful injection", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_001",
			Participants: []string{"player", "npc_001"},
			StartBeat:    10,
			EndBeat:      15,
			Summary: &ChatSummary{
				MainTopics:   []string{"醫療物資"},
				KeyDecisions: []string{"同意合作"},
				RelationChanges: map[string]string{
					"npc_001": "信任增加",
				},
				FactsShared:     []string{"避難所在東翼"},
				NarrativeImpact: "建立合作關係",
			},
		}

		err := orch.InjectChatSummary(session)
		if err != nil {
			t.Errorf("InjectChatSummary failed: %v", err)
		}

		// Verify formatting occurred (no error = success for now)
		// Future: Verify NarrationAgent context was updated
	})

	t.Run("nil session", func(t *testing.T) {
		err := orch.InjectChatSummary(nil)
		if err == nil {
			t.Error("Expected error for nil session")
		}
	})

	t.Run("nil summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_002",
			Participants: []string{"player"},
			Summary:      nil,
		}

		err := orch.InjectChatSummary(session)
		if err != nil {
			t.Errorf("Should not error on nil summary: %v", err)
		}
	})

	t.Run("wrong summary type", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_003",
			Participants: []string{"player"},
			Summary:      123, // Wrong type
		}

		// Should not panic, just log warning
		err := orch.InjectChatSummary(session)
		if err != nil {
			t.Errorf("Should gracefully handle wrong summary type: %v", err)
		}
	})
}

// TestApplySummaryEmotionChanges tests emotion change application.
// Story 5.5 AC3: 情感變化影響後續 NPC 行為
func TestApplySummaryEmotionChanges(t *testing.T) {
	orch := NewOrchestrator()

	t.Run("with emotion changes", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_001",
			Participants: []string{"player", "npc_001"},
			Summary: &ChatSummary{
				EmotionChanges: map[string]string{
					"npc_001": "信任增加，壓力減少",
					"npc_002": "恐懼增加",
				},
			},
		}

		err := orch.applySummaryEmotionChanges(session)
		if err != nil {
			t.Errorf("applySummaryEmotionChanges failed: %v", err)
		}

		// Note: Actual emotion changes require NPCManager integration
		// For now, just verify no error
	})

	t.Run("with relation changes", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_002",
			Participants: []string{"player", "npc_001"},
			Summary: &ChatSummary{
				RelationChanges: map[string]string{
					"npc_001": "關係改善",
					"npc_002": "發生衝突",
				},
			},
		}

		err := orch.applySummaryEmotionChanges(session)
		if err != nil {
			t.Errorf("applySummaryEmotionChanges failed: %v", err)
		}
	})

	t.Run("no emotion changes", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_003",
			Participants: []string{"player"},
			Summary:      &ChatSummary{},
		}

		err := orch.applySummaryEmotionChanges(session)
		if err != nil {
			t.Errorf("Should not error on empty emotion changes: %v", err)
		}
	})

	t.Run("nil session", func(t *testing.T) {
		err := orch.applySummaryEmotionChanges(nil)
		if err != nil {
			t.Errorf("Should not error on nil session: %v", err)
		}
	})

	t.Run("nil summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:      "chat_004",
			Summary: nil,
		}

		err := orch.applySummaryEmotionChanges(session)
		if err != nil {
			t.Errorf("Should not error on nil summary: %v", err)
		}
	})

	t.Run("wrong summary type", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:      "chat_005",
			Summary: "wrong type",
		}

		err := orch.applySummaryEmotionChanges(session)
		if err == nil {
			t.Error("Expected error for wrong summary type")
		}
	})
}

// TestApplySummaryFacts tests fact propagation from summary.
// Story 5.5 AC4: FactsShared 影響 NPC 知識庫
func TestApplySummaryFacts(t *testing.T) {
	orch := NewOrchestrator()

	t.Run("with facts shared", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_001",
			Participants: []string{"player", "npc_001", "npc_002"},
			StartBeat:    10,
			EndBeat:      12,
			Summary: &ChatSummary{
				FactsShared: []string{
					"避難所在東翼",
					"地下室有醫療物資",
					"西翼已被封鎖",
				},
			},
		}

		err := orch.applySummaryFacts(session)
		if err != nil {
			t.Errorf("applySummaryFacts failed: %v", err)
		}

		// Note: Actual fact propagation requires UpdateManager integration
		// For now, just verify no error
	})

	t.Run("no facts shared", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:           "chat_002",
			Participants: []string{"player"},
			Summary:      &ChatSummary{},
		}

		err := orch.applySummaryFacts(session)
		if err != nil {
			t.Errorf("Should not error on empty facts: %v", err)
		}
	})

	t.Run("nil session", func(t *testing.T) {
		err := orch.applySummaryFacts(nil)
		if err != nil {
			t.Errorf("Should not error on nil session: %v", err)
		}
	})

	t.Run("nil summary", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:      "chat_003",
			Summary: nil,
		}

		err := orch.applySummaryFacts(session)
		if err != nil {
			t.Errorf("Should not error on nil summary: %v", err)
		}
	})

	t.Run("wrong summary type", func(t *testing.T) {
		session := &engine.ChatSession{
			ID:      "chat_004",
			Summary: "wrong type",
		}

		err := orch.applySummaryFacts(session)
		if err == nil {
			t.Error("Expected error for wrong summary type")
		}
	})
}

// TestGetParticipantName tests participant name lookup.
func TestGetParticipantName(t *testing.T) {
	orch := NewOrchestrator()

	// Setup NPC profiles
	orch.storyBible.NPCProfiles = []*NPCProfile{
		{
			ID:   "npc_001",
			Name: "張醫生",
		},
		{
			ID:   "npc_002",
			Name: "護士王",
		},
	}

	t.Run("player", func(t *testing.T) {
		name := orch.getParticipantName("player")
		if name != "玩家" {
			t.Errorf("Expected '玩家', got '%s'", name)
		}
	})

	t.Run("known NPC", func(t *testing.T) {
		name := orch.getParticipantName("npc_001")
		if name != "張醫生" {
			t.Errorf("Expected '張醫生', got '%s'", name)
		}

		name = orch.getParticipantName("npc_002")
		if name != "護士王" {
			t.Errorf("Expected '護士王', got '%s'", name)
		}
	})

	t.Run("unknown NPC", func(t *testing.T) {
		name := orch.getParticipantName("npc_999")
		if name != "npc_999" {
			t.Errorf("Expected 'npc_999' (ID fallback), got '%s'", name)
		}
	})

	t.Run("nil story bible", func(t *testing.T) {
		orch2 := NewOrchestrator()
		orch2.storyBible = nil

		name := orch2.getParticipantName("npc_001")
		if name != "npc_001" {
			t.Errorf("Expected 'npc_001' (fallback), got '%s'", name)
		}
	})

	t.Run("nil profiles", func(t *testing.T) {
		orch3 := NewOrchestrator()
		orch3.storyBible.NPCProfiles = nil

		name := orch3.getParticipantName("npc_001")
		if name != "npc_001" {
			t.Errorf("Expected 'npc_001' (fallback), got '%s'", name)
		}
	})
}

// TestChatExitToNarration_Integration tests the complete flow from chat exit to summary injection.
// Story 5.5 AC1: Integration test for chat exit -> summary injection -> narration
func TestChatExitToNarration_Integration(t *testing.T) {
	orch := NewOrchestrator()

	// Setup NPC profiles
	orch.storyBible.NPCProfiles = []*NPCProfile{
		{
			ID:   "npc_001",
			Name: "張醫生",
		},
	}

	t.Run("complete flow", func(t *testing.T) {
		// 1. Create a chat session with summary
		session := &engine.ChatSession{
			ID:           "chat_001",
			Participants: []string{"player", "npc_001"},
			StartBeat:    10,
			EndBeat:      15,
			Summary: &ChatSummary{
				MainTopics:   []string{"醫療物資", "避難所"},
				KeyDecisions: []string{"同意合作"},
				RelationChanges: map[string]string{
					"npc_001": "信任增加",
				},
				FactsShared: []string{"東翼有物資"},
				EmotionChanges: map[string]string{
					"npc_001": "恐懼減少，信任增加",
				},
				NarrativeImpact: "建立合作關係，獲得關鍵線索",
			},
		}

		// 2. Save chat session
		err := orch.SaveChatSession(session)
		if err != nil {
			t.Fatalf("SaveChatSession failed: %v", err)
		}

		// 3. Inject summary
		err = orch.InjectChatSummary(session)
		if err != nil {
			t.Fatalf("InjectChatSummary failed: %v", err)
		}

		// 4. Verify session is stored
		storedSession := orch.gameState.GetChatSessionByID("chat_001")
		if storedSession == nil {
			t.Error("Chat session was not stored in game state")
		}

		// 5. Verify session details
		if storedSession.ID != "chat_001" {
			t.Errorf("Expected session ID 'chat_001', got '%s'", storedSession.ID)
		}

		if len(storedSession.Participants) != 2 {
			t.Errorf("Expected 2 participants, got %d", len(storedSession.Participants))
		}

		// Future: Verify NarrationAgent context includes summary
		// Future: Verify NPCManager emotion states updated
		// Future: Verify UpdateManager received facts
	})
}

// assertContains is a helper function to check if a string contains a substring.
func assertContains(t *testing.T, text, substr, message string) {
	t.Helper()
	if text == "" {
		t.Errorf("%s: text is empty", message)
		return
	}
	if substr == "" {
		t.Errorf("%s: substring is empty", message)
		return
	}
	if !containsString(text, substr) {
		t.Errorf("%s: text does not contain '%s'\nGot: %s", message, substr, text)
	}
}

// containsString checks if a string contains a substring (case-sensitive).
func containsString(text, substr string) bool {
	return len(text) >= len(substr) && (text == substr || len(substr) == 0 || findSubstring(text, substr))
}

// findSubstring performs a simple substring search.
func findSubstring(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(text) < len(substr) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
