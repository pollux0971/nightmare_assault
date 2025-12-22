package manager

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
)

// TestProcessChatMessage tests the basic flow of processing a chat message.
// AC2: ProcessChatMessage() integrates information propagation and contradiction detection
func TestProcessChatMessage(t *testing.T) {
	// Setup UpdateManager and NPCManager
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	// Setup rooms
	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("npc_001", "lobby")

	// Add NPC
	profile := &NPCProfile{
		ID:   "npc_001",
		Name: "張醫生",
	}
	err := npcMgr.AddNPC(profile)
	if err != nil {
		t.Fatalf("Failed to add NPC: %v", err)
	}

	// Process chat message
	contradictions := npcMgr.ProcessChatMessage("player", "我看到怪物了", "lobby")

	// Verify: NPC should have learned the information
	kb := updateMgr.GetKnowledgeBase("npc_001")
	if kb == nil {
		t.Fatal("Expected knowledge base to exist")
	}
	if len(kb.KnownFacts) == 0 {
		t.Error("Expected NPC to learn from dialogue, but knowledge base is empty")
	}

	// Verify: No contradictions
	if len(contradictions) != 0 {
		t.Errorf("Expected no contradictions, got %d", len(contradictions))
	}
}

// TestProcessChatMessageWithContradiction tests contradiction detection and automatic emotion adjustment.
// AC2: ProcessChatMessage() integrates contradiction detection
// AC3: Automatically adjusts NPC emotions when contradictions are detected
func TestProcessChatMessageWithContradiction(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("npc_001", "lobby")

	profile := &NPCProfile{
		ID:   "npc_001",
		Name: "張醫生",
	}
	npcMgr.AddNPC(profile)

	// Setup: Set initial emotion (give some trust to start with)
	state := npcMgr.GetState("npc_001")
	state.Emotion.Trust = 50
	state.Emotion.Stress = 30

	// NPC learns "王護士還活著" from witness
	updateMgr.SetEntityRoom("witness", "lobby")
	updateMgr.LearnFromDialogue("npc_001", "witness", "王護士還活著", "lobby")

	// Record initial trust and stress
	initialTrust := npcMgr.GetState("npc_001").Emotion.Trust
	initialStress := npcMgr.GetState("npc_001").Emotion.Stress

	// Player says contradictory information
	contradictions := npcMgr.ProcessChatMessage("player", "王護士已經死了", "lobby")

	// Verify: Contradiction detected
	if len(contradictions) != 1 {
		t.Fatalf("Expected 1 contradiction, got %d", len(contradictions))
	}

	// Accept either moderate or major contradiction (depends on confidence and learn method)
	if contradictions[0].Type != knowledge.ContradictionMajor && contradictions[0].Type != knowledge.ContradictionModerate {
		t.Errorf("Expected major or moderate contradiction, got %v", contradictions[0].Type)
	}

	// Verify: Emotion adjusted (stress should increase)
	newTrust := npcMgr.GetState("npc_001").Emotion.Trust
	newStress := npcMgr.GetState("npc_001").Emotion.Stress

	if newStress <= initialStress {
		t.Errorf("Expected stress to increase from %d, got %d", initialStress, newStress)
	}

	// Trust may decrease (check if it's different)
	if newTrust == initialTrust {
		t.Logf("Note: Trust remained at %d (no change)", initialTrust)
	}

	// Verify: Interaction recorded
	state = npcMgr.GetState("npc_001")
	if len(state.Interactions) == 0 {
		t.Fatal("Expected interaction to be recorded")
	}

	lastInteraction := state.Interactions[len(state.Interactions)-1]
	if lastInteraction.InteractionType != "contradiction" {
		t.Errorf("Expected interaction type 'contradiction', got '%s'", lastInteraction.InteractionType)
	}
}

// TestProcessChatMessageDifferentRooms tests that NPCs in different rooms don't learn from dialogue.
func TestProcessChatMessageDifferentRooms(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	// Player and npc_001 in lobby, npc_002 in hallway
	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("npc_001", "lobby")
	updateMgr.SetEntityRoom("npc_002", "hallway")

	npcMgr.AddNPC(&NPCProfile{ID: "npc_001", Name: "張醫生"})
	npcMgr.AddNPC(&NPCProfile{ID: "npc_002", Name: "李護士"})

	// Process message in lobby
	npcMgr.ProcessChatMessage("player", "測試訊息", "lobby")

	// Verify: npc_001 learned (same room)
	kb1 := updateMgr.GetKnowledgeBase("npc_001")
	if kb1 == nil || len(kb1.KnownFacts) == 0 {
		t.Error("Expected npc_001 to learn from dialogue in same room")
	}

	// Verify: npc_002 did NOT learn (different room)
	kb2 := updateMgr.GetKnowledgeBase("npc_002")
	if kb2 != nil && len(kb2.KnownFacts) > 0 {
		t.Error("Expected npc_002 NOT to learn from dialogue in different room")
	}
}

// TestBuildFullNPCPrompt tests knowledge base integration in prompt building.
// AC4: BuildFullNPCPrompt() integrates knowledge base information
// AC5: Marks what the NPC doesn't know
func TestBuildFullNPCPrompt(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:         "npc_001",
		Name:       "張醫生",
		Appearance: "40多歲男性，戴眼鏡",
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業醫學術語",
		},
	}
	npcMgr.AddNPC(profile)

	// Set both entities in the same room
	updateMgr.SetEntityRoom("npc_001", "lobby")
	updateMgr.SetEntityRoom("player", "lobby")

	// NPC learns information from player
	updateMgr.LearnFromDialogue("npc_001", "player", "發現了密道", "lobby")

	// Build full prompt
	prompt := npcMgr.BuildFullNPCPrompt("npc_001")

	// Debug: print the prompt to see what's in it
	t.Logf("Generated prompt:\n%s", prompt)

	// Verify KB was created
	kb := updateMgr.GetKnowledgeBase("npc_001")
	if kb == nil {
		t.Fatal("Knowledge base is nil - LearnFromDialogue may have failed due to room mismatch")
	}
	t.Logf("KB has %d known facts", len(kb.KnownFacts))

	// Verify: Contains NPC name
	if !strings.Contains(prompt, "張醫生") {
		t.Error("Expected prompt to contain NPC name")
	}

	// Verify: Contains knowledge section
	if !strings.Contains(prompt, "## 已知資訊") {
		t.Error("Expected prompt to contain '## 已知資訊' section")
	}

	// Since LearnFromDialogue creates facts automatically, we should check if
	// the knowledge section exists and has content
	if strings.Contains(prompt, "沒有獲得任何重要資訊") {
		t.Error("NPC should have learned information from dialogue")
	}

	// Verify: Contains "unknown information" section
	if !strings.Contains(prompt, "## 不知道的事項") {
		t.Error("Expected prompt to contain '## 不知道的事項' section")
	}

	// Verify: Contains warning about only knowing listed facts
	if !strings.Contains(prompt, "只知道") {
		t.Error("Expected prompt to contain warning about only knowing listed facts")
	}
}

// TestProcessChatMessageWithoutUpdateManager tests graceful degradation when UpdateManager is nil.
func TestProcessChatMessageWithoutUpdateManager(t *testing.T) {
	npcMgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	npcMgr.AddNPC(&NPCProfile{ID: "npc_001", Name: "張醫生"})

	// Should not panic
	contradictions := npcMgr.ProcessChatMessage("player", "test", "lobby")

	// Should return nil (graceful degradation)
	if contradictions != nil {
		t.Errorf("Expected nil contradictions when UpdateManager is nil, got %v", contradictions)
	}
}

// TestBuildFullNPCPromptWithoutUpdateManager tests graceful degradation when UpdateManager is nil.
func TestBuildFullNPCPromptWithoutUpdateManager(t *testing.T) {
	npcMgr := NewNPCManager(nil, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:         "npc_001",
		Name:       "張醫生",
		Appearance: "40多歲男性",
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業",
		},
	}
	npcMgr.AddNPC(profile)

	// Should return base prompt without knowledge sections
	prompt := npcMgr.BuildFullNPCPrompt("npc_001")

	// Verify: Contains basic info
	if !strings.Contains(prompt, "張醫生") {
		t.Error("Expected prompt to contain NPC name")
	}

	// Verify: Does NOT contain knowledge section (no UpdateManager)
	if strings.Contains(prompt, "## 已知資訊") {
		t.Error("Expected prompt NOT to contain knowledge section when UpdateManager is nil")
	}
}

// TestBuildFullNPCPromptWithNoKnowledge tests prompt when NPC has no knowledge.
func TestBuildFullNPCPromptWithNoKnowledge(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	profile := &NPCProfile{
		ID:         "npc_001",
		Name:       "張醫生",
		Appearance: "40多歲男性",
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業",
		},
	}
	npcMgr.AddNPC(profile)

	// Build prompt (NPC has no knowledge yet)
	prompt := npcMgr.BuildFullNPCPrompt("npc_001")

	// Verify: Contains knowledge section
	if !strings.Contains(prompt, "## 已知資訊") {
		t.Error("Expected prompt to contain knowledge section")
	}

	// Verify: Contains message about no information
	if !strings.Contains(prompt, "沒有獲得任何重要資訊") {
		t.Error("Expected prompt to indicate NPC has no information")
	}
}

// TestBuildFullNPCPromptWithMultipleFacts tests prompt with multiple facts.
func TestBuildFullNPCPromptWithMultipleFacts(t *testing.T) {
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := NewNPCManager(updateMgr, DefaultNPCManagerConfig())

	npcMgr.AddNPC(&NPCProfile{
		ID:         "npc_001",
		Name:       "張醫生",
		Appearance: "40多歲男性",
		DialogueStyle: DialogueStyle{
			Vocabulary: "專業",
		},
	})

	// Set rooms
	updateMgr.SetEntityRoom("npc_001", "lobby")
	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("witness", "lobby")

	// NPC learns multiple facts
	updateMgr.LearnFromDialogue("npc_001", "player", "發現了密道", "lobby")
	updateMgr.LearnFromDialogue("npc_001", "witness", "聽到奇怪的聲音", "lobby")

	prompt := npcMgr.BuildFullNPCPrompt("npc_001")

	// Verify: Contains knowledge section
	if !strings.Contains(prompt, "## 已知資訊") {
		t.Error("Expected prompt to contain knowledge section")
	}

	// Verify: Should contain both facts
	kb := updateMgr.GetKnowledgeBase("npc_001")
	if kb == nil || len(kb.KnownFacts) < 2 {
		t.Errorf("Expected at least 2 facts, got %d", len(kb.KnownFacts))
	}

	// Verify: Prompt should not indicate "no information"
	if strings.Contains(prompt, "沒有獲得任何重要資訊") {
		t.Error("NPC should have learned multiple facts")
	}
}
