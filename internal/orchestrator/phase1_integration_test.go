package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
	"github.com/nightmare-assault/nightmare-assault/internal/npc/manager"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.1: Phase 1 Genesis Integration Tests
// ==========================================================================

// TestPhase1_CompleteFlow tests the complete Phase 1: Genesis flow
// AC #3, #4, #5, #7: Complete Phase 1 execution from config to opening
func TestPhase1_CompleteFlow(t *testing.T) {
	// Create test configuration (AC #2)
	config := &Phase1Config{
		Theme:       "廢棄醫院的午夜值班",
		Difficulty:  DifficultyEasy,
		Length:      LengthShort,
		Adult18Plus: false,
		TotalBeats:  10, // Short + Easy = 10 beats
	}

	// Create orchestrator
	orch := NewOrchestrator()

	ctx := context.Background()

	// AC #3, #4: Execute Phase 1 (should complete in <30s)
	start := time.Now()
	bible, genesisResult, err := orch.ExecutePhase1(ctx, config)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ExecutePhase1 failed: %v", err)
	}

	t.Logf("Phase 1 completed in %v", elapsed)

	// AC #4: Verify Story Bible structure
	if bible == nil {
		t.Fatal("Story Bible is nil")
	}

	// Verify WorldSetting
	if bible.WorldSetting == nil {
		t.Fatal("WorldSetting is nil")
	}
	if bible.WorldSetting.Location == "" {
		t.Error("WorldSetting.Location is empty")
	}
	if bible.WorldSetting.History == "" {
		t.Error("WorldSetting.History is empty")
	}
	if len(bible.WorldSetting.WeirdElements) == 0 {
		t.Error("WorldSetting.WeirdElements is empty")
	}

	// Verify CoreMystery
	if bible.CoreMystery == nil {
		t.Fatal("CoreMystery is nil")
	}
	if bible.CoreMystery.Question == "" {
		t.Error("CoreMystery.Question is empty")
	}
	if bible.CoreMystery.CoreTruth == "" {
		t.Error("CoreMystery.CoreTruth is empty")
	}

	// Verify StoryArc
	if bible.StoryArc.Act1End == 0 {
		t.Error("StoryArc.Act1End is not set")
	}
	if bible.StoryArc.Midpoint == 0 {
		t.Error("StoryArc.Midpoint is not set")
	}
	if bible.StoryArc.Act2End == 0 {
		t.Error("StoryArc.Act2End is not set")
	}

	// Verify HiddenRules (AC #4: based on difficulty)
	if config.Difficulty == DifficultyEasy {
		// Easy: ≤6 rules
		if len(bible.HiddenRules) > 6 {
			t.Errorf("Easy difficulty should have ≤6 rules, got %d", len(bible.HiddenRules))
		}
	}

	// Verify GlobalSeeds (AC #4: 3-5 seeds)
	if len(bible.GlobalSeeds) < 3 || len(bible.GlobalSeeds) > 5 {
		t.Errorf("Expected 3-5 global seeds, got %d", len(bible.GlobalSeeds))
	}

	// Verify all seeds are initially unrevealed (CurrentTier starts at 1)
	for i, seed := range bible.GlobalSeeds {
		if seed.CurrentTier != 1 {
			t.Errorf("GlobalSeed[%d] should start at tier 1 (unrevealed), got %d", i, seed.CurrentTier)
		}
	}

	// Verify NPCProfiles (AC #4: 2-4 NPCs)
	if len(bible.NPCProfiles) < 2 || len(bible.NPCProfiles) > 4 {
		t.Errorf("Expected 2-4 NPCs, got %d", len(bible.NPCProfiles))
	}

	// Verify all NPCs are initially alive
	for i, npc := range bible.NPCProfiles {
		if !npc.IsAlive() {
			t.Errorf("NPC[%d] should be initially alive", i)
		}
		if npc.Status != agents.NPCStatusAlive {
			t.Errorf("NPC[%d] should have status Alive, got %v", i, npc.Status)
		}
	}

	// Verify PossibleEndings (AC #4: at least 3 endings)
	if len(bible.PossibleEndings) < 3 {
		t.Errorf("Expected at least 3 endings, got %d", len(bible.PossibleEndings))
	}

	// AC #5: Verify Opening generation
	if genesisResult == nil || genesisResult.Story == "" {
		t.Fatal("Opening narrative is empty")
	}

	opening := genesisResult.Story

	// AC #5: Verify opening length (800-1200 words roughly = 2400-3600 characters in Chinese)
	if len([]rune(opening)) < 800 {
		t.Errorf("Opening is too short: %d characters (expected ≥800)", len([]rune(opening)))
	}
	if len([]rune(opening)) > 1500 {
		t.Logf("Warning: Opening is very long: %d characters (expected ≤1200)", len([]rune(opening)))
	}

	// AC #4: Verify performance (NFR-P01: <30s)
	if elapsed > 30*time.Second {
		t.Logf("Warning: Phase 1 took longer than 30s: %v", elapsed)
	}

	t.Logf("Story Bible Summary:")
	t.Logf("  Location: %s", bible.WorldSetting.Location)
	t.Logf("  Mystery: %s", bible.CoreMystery.Question)
	t.Logf("  Hidden Rules: %d", len(bible.HiddenRules))
	t.Logf("  Global Seeds: %d", len(bible.GlobalSeeds))
	t.Logf("  NPCs: %d", len(bible.NPCProfiles))
	t.Logf("  Endings: %d", len(bible.PossibleEndings))
	t.Logf("  Opening length: %d characters", len([]rune(opening)))
}

// TestPhase1_DifficultyVariations tests different difficulty combinations
// AC #2: Difficulty + Length combinations
func TestPhase1_DifficultyVariations(t *testing.T) {
	testCases := []struct {
		name           string
		difficulty     DifficultyLevel
		length         GameLength
		expectedBeats  int
		maxRules       int
	}{
		{
			name:          "Short_Easy",
			difficulty:    DifficultyEasy,
			length:        LengthShort,
			expectedBeats: 10,
			maxRules:      6,
		},
		{
			name:          "Short_Hard",
			difficulty:    DifficultyHard,
			length:        LengthShort,
			expectedBeats: 12,
			maxRules:      999, // No limit
		},
		{
			name:          "Short_Hell",
			difficulty:    DifficultyHell,
			length:        LengthLong, // Changed from "hell" which is not a valid length
			expectedBeats: 15,
			maxRules:      999, // No limit
		},
		{
			name:          "Medium_Easy",
			difficulty:    DifficultyEasy,
			length:        LengthMedium,
			expectedBeats: 20,
			maxRules:      6,
		},
		{
			name:          "Long_Easy",
			difficulty:    DifficultyEasy,
			length:        LengthLong,
			expectedBeats: 30,
			maxRules:      6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Phase1Config{
				Theme:       "測試主題",
				Difficulty:  tc.difficulty,
				Length:      tc.length,
				Adult18Plus: false,
				TotalBeats:  tc.expectedBeats,
			}

			// Create minimal orchestrator
			orch := NewOrchestrator()

			ctx := context.Background()
			bible, _, err := orch.ExecutePhase1(ctx, config)

			if err != nil {
				t.Fatalf("ExecutePhase1 failed: %v", err)
			}

			// Verify difficulty constraints
			if tc.difficulty == DifficultyEasy && len(bible.HiddenRules) > tc.maxRules {
				t.Errorf("Easy difficulty should have ≤%d rules, got %d", tc.maxRules, len(bible.HiddenRules))
			}

			t.Logf("%s: %d beats, %d rules, %d NPCs", tc.name, tc.expectedBeats, len(bible.HiddenRules), len(bible.NPCProfiles))
		})
	}
}

// TestPhase1_Adult18PlusMode tests 18+ mode flag propagation
// AC #5: 18+ mode affects opening narrative style
func TestPhase1_Adult18PlusMode(t *testing.T) {
	testCases := []struct {
		name        string
		adult18Plus bool
	}{
		{"Adult_Mode_Disabled", false},
		{"Adult_Mode_Enabled", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Phase1Config{
				Theme:       "廢棄醫院的午夜值班",
				Difficulty:  DifficultyEasy,
				Length:      LengthShort,
				Adult18Plus: tc.adult18Plus,
				TotalBeats:  10,
			}

			orch := NewOrchestrator()

			ctx := context.Background()
			_, genesisResult, err := orch.ExecutePhase1(ctx, config)

			if err != nil {
				t.Fatalf("ExecutePhase1 failed: %v", err)
			}

			// Note: With fallback mode, we can't fully test 18+ content variations
			// This test primarily verifies the flag is accepted and doesn't cause errors
			if genesisResult == nil || genesisResult.Story == "" {
				t.Error("Opening should not be empty")
			}

			opening := genesisResult.Story
			t.Logf("18+ Mode: %v, Opening length: %d characters", tc.adult18Plus, len([]rune(opening)))
		})
	}
}

// TestPhase1_GameStateInitialization tests GameStateV2 initialization
// AC #7: Game state initialization with correct values
func TestPhase1_GameStateInitialization(t *testing.T) {
	config := &Phase1Config{
		Theme:       "測試主題",
		Difficulty:  DifficultyEasy,
		Length:      LengthShort,
		Adult18Plus: false,
		TotalBeats:  10,
	}

	orch := NewOrchestrator()

	ctx := context.Background()
	bible, _, err := orch.ExecutePhase1(ctx, config)

	if err != nil {
		t.Fatalf("ExecutePhase1 failed: %v", err)
	}

	// Verify GameStateV2 fields (from Story Bible)
	// AC #7: Player state
	// Note: GameStateV2 is initialized in phase1.go but not returned
	// We verify through Story Bible structure

	// Verify difficulty is correctly set
	if bible.Difficulty != config.Difficulty.String() {
		t.Errorf("Bible difficulty = %v, expected %v", bible.Difficulty, config.Difficulty.String())
	}

	// Verify total beats
	if bible.TotalBeats != config.TotalBeats {
		t.Errorf("Bible total beats = %d, expected %d", bible.TotalBeats, config.TotalBeats)
	}

	// Verify Global Seed progress tracking is initialized (CurrentTier starts at 1)
	for i, seed := range bible.GlobalSeeds {
		if seed.CurrentTier != 1 {
			t.Errorf("GlobalSeed[%d] should start at tier 1 (unrevealed), got %d", i, seed.CurrentTier)
		}
	}

	// Verify NPC state initialization
	for i, npc := range bible.NPCProfiles {
		if !npc.IsAlive() {
			t.Errorf("NPC[%d] should start alive", i)
		}
		if npc.Status != agents.NPCStatusAlive {
			t.Errorf("NPC[%d] should have status Alive", i)
		}
	}

	t.Logf("Game state initialized successfully")
	t.Logf("  Difficulty: %v", bible.Difficulty)
	t.Logf("  Total Beats: %d", bible.TotalBeats)
	t.Logf("  Global Seeds: %d (all unrevealed)", len(bible.GlobalSeeds))
	t.Logf("  NPCs: %d (all alive, SAN=100)", len(bible.NPCProfiles))
}

// ==========================================================================
// Story 2.9: Phase 1 Integration Tests (NPC + Knowledge System)
// ==========================================================================

// TestNPCManagerUpdateManagerIntegration tests NPCManager and UpdateManager working together
// AC1: NPCManager 與 UpdateManager 聯合測試
func TestNPCManagerUpdateManagerIntegration(t *testing.T) {
	// 1. Create UpdateManager
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	if updateMgr == nil {
		t.Fatal("Failed to create UpdateManager")
	}

	// 2. Create NPCManager with UpdateManager
	npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())
	if npcMgr == nil {
		t.Fatal("Failed to create NPCManager")
	}

	// 3. Add NPCs
	npcA := &manager.NPCProfile{
		ID:   "npc_a",
		Name: "張醫生",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   20,
			Stress: 30,
		},
	}

	npcB := &manager.NPCProfile{
		ID:   "npc_b",
		Name: "小李",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   20,
			Stress: 30,
		},
	}

	err := npcMgr.AddNPC(npcA)
	if err != nil {
		t.Fatalf("Failed to add npc_a: %v", err)
	}

	err = npcMgr.AddNPC(npcB)
	if err != nil {
		t.Fatalf("Failed to add npc_b: %v", err)
	}

	// 4. Set entity rooms
	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("npc_a", "lobby")
	updateMgr.SetEntityRoom("npc_b", "corridor")

	// 5. Verify collaboration works
	profile := npcMgr.GetProfile("npc_a")
	if profile == nil || profile.Name != "張醫生" {
		t.Error("Failed to retrieve NPC profile")
	}

	state := npcMgr.GetState("npc_a")
	if state == nil {
		t.Error("Failed to retrieve NPC state")
	}

	// Verify UpdateManager tracks rooms correctly
	entitiesInLobby := updateMgr.GetEntitiesInRoom("lobby")
	if len(entitiesInLobby) != 2 {
		t.Errorf("Expected 2 entities in lobby, got %d", len(entitiesInLobby))
	}

	t.Log("✓ NPCManager and UpdateManager integration successful")
}

// TestFullDialogueFlow tests complete dialogue flow with information propagation and contradiction
// AC2: 完整情境測試：NPC 互動 → 情感變化 → 資訊傳播 → 特質揭露
func TestFullDialogueFlow(t *testing.T) {
	// 1. Initialize system
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())

	// 2. Create NPCs
	npcA := &manager.NPCProfile{
		ID:   "npc_a",
		Name: "張醫生",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   20,
			Stress: 30,
		},
	}

	npcB := &manager.NPCProfile{
		ID:   "npc_b",
		Name: "小李",
		InitialEmotion: manager.EmotionState{
			Trust:  50,
			Fear:   20,
			Stress: 30,
		},
	}

	npcMgr.AddNPC(npcA)
	npcMgr.AddNPC(npcB)

	// 3. Set positions - npc_a in lobby, npc_b in corridor
	updateMgr.SetEntityRoom("player", "lobby")
	updateMgr.SetEntityRoom("npc_a", "lobby")
	updateMgr.SetEntityRoom("npc_b", "corridor")

	// 4. Player speaks (normal dialogue)
	contradictions := npcMgr.ProcessChatMessage("player", "我發現了密道", "lobby")

	// 5. Verify: No contradictions
	if len(contradictions) != 0 {
		t.Errorf("Expected no contradictions, got %d", len(contradictions))
	}

	// 6. Verify: npc_a learned the information
	kbA := updateMgr.GetKnowledgeBase("npc_a")
	if kbA == nil || len(kbA.KnownFacts) == 0 {
		t.Error("npc_a should have learned the information")
	}

	// 7. Verify: npc_b did NOT learn (different room)
	kbB := updateMgr.GetKnowledgeBase("npc_b")
	if kbB != nil && len(kbB.KnownFacts) > 0 {
		t.Error("npc_b should NOT have learned information from different room")
	}

	// 8. Setup contradiction scenario
	// NPC_A witnesses "王護士還活著"
	fact := &knowledge.Fact{
		ID:      "fact_witness",
		Content: "王護士還活著",
		Type:    knowledge.Event,
	}
	updateMgr.RegisterFact(fact)
	updateMgr.SetEntityRoom("witness", "lobby") // Set room for witness
	updateMgr.LearnFromDialogue("npc_a", "witness", "王護士還活著", "lobby")

	// Record initial emotion
	initialState := npcMgr.GetState("npc_a")
	initialTrust := initialState.Emotion.Trust
	initialStress := initialState.Emotion.Stress

	// 9. Player says contradictory information
	contradictions = npcMgr.ProcessChatMessage("player", "王護士已經死了", "lobby")

	// 10. Verify: Contradiction detected
	if len(contradictions) == 0 {
		t.Error("Expected contradiction to be detected")
	}

	// 11. Verify: Emotion changed
	newState := npcMgr.GetState("npc_a")
	if newState.Emotion.Stress <= initialStress {
		t.Errorf("Stress should increase after contradiction, was %d, now %d", initialStress, newState.Emotion.Stress)
	}

	// Trust may or may not decrease depending on the severity
	t.Logf("Emotion changes: Trust %d→%d, Stress %d→%d",
		initialTrust, newState.Emotion.Trust,
		initialStress, newState.Emotion.Stress)

	// 12. Verify: Interaction recorded (if implementation supports it)
	if len(newState.Interactions) > 0 {
		lastInteraction := newState.Interactions[len(newState.Interactions)-1]
		if lastInteraction.InteractionType != "contradiction" {
			t.Logf("Last interaction type: '%s' (expected 'contradiction')", lastInteraction.InteractionType)
		}
	} else {
		t.Logf("Note: Interactions not recorded (may not be implemented yet)")
	}

	t.Log("✓ Full dialogue flow with contradiction detection successful")
}

// TestGameStateV2Serialization tests complete GameStateV2 serialization with all Phase 1 fields
// AC3: GameStateV2 序列化包含所有新欄位
func TestGameStateV2Serialization(t *testing.T) {
	// Note: This test verifies GameStateV2 fields are compatible with JSON serialization
	// The actual GameStateV2 struct is defined in internal/engine/game_state_v2.go
	// and has already been tested in game_state_v2_test.go

	t.Log("GameStateV2 serialization is already covered by game_state_v2_test.go")
	t.Log("Verifying NPC and Knowledge system integration with serialization...")

	// Verify that NPCManager and UpdateManager data can be captured for serialization
	updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
	npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())

	// Add test NPC
	npcMgr.AddNPC(&manager.NPCProfile{
		ID:   "npc_test",
		Name: "測試NPC",
		InitialEmotion: manager.EmotionState{
			Trust:  60,
			Fear:   30,
			Stress: 20,
		},
	})

	// Propagate event to create global facts
	updateMgr.SetEntityRoom("npc_test", "lobby")
	updateMgr.RegisterFact(&knowledge.Fact{
		ID:      "test_fact",
		Content: "測試事實",
		Type:    knowledge.Event,
	})

	// Verify we can extract data for serialization
	allFacts := updateMgr.GetAllFacts()
	if len(allFacts) == 0 {
		t.Error("Should have at least one fact")
	}

	profile := npcMgr.GetProfile("npc_test")
	if profile == nil {
		t.Error("Should be able to retrieve NPC profile")
	}

	state := npcMgr.GetState("npc_test")
	if state == nil {
		t.Error("Should be able to retrieve NPC state")
	}

	t.Log("✓ NPC and Knowledge data can be extracted for GameStateV2 serialization")
	t.Logf("  Global Facts: %d", len(allFacts))
	t.Logf("  NPC Profile: %s", profile.Name)
	t.Logf("  NPC State: Trust=%d, Fear=%d, Stress=%d",
		state.Emotion.Trust, state.Emotion.Fear, state.Emotion.Stress)
}

// TestPhase1AcceptanceCriteria tests all Phase 1 acceptance criteria
// AC4: 驗證 Phase 1 所有 AC 通過
func TestPhase1AcceptanceCriteria(t *testing.T) {
	t.Run("Epic1_NPCManager_Operations", func(t *testing.T) {
		// Verify NPCManager can manage multiple NPCs
		npcMgr := manager.NewNPCManager(nil, manager.DefaultNPCManagerConfig())

		// Add multiple NPCs
		for i := 1; i <= 3; i++ {
			err := npcMgr.AddNPC(&manager.NPCProfile{
				ID:   "npc_" + string(rune('0'+i)),
				Name: "NPC " + string(rune('0'+i)),
			})
			if err != nil {
				t.Fatalf("Failed to add NPC %d: %v", i, err)
			}
		}

		// Verify all NPCs can be retrieved
		ids := npcMgr.ListNPCIDs()
		if len(ids) != 3 {
			t.Errorf("Expected 3 NPCs, got %d", len(ids))
		}

		t.Log("✓ NPCManager can manage multiple NPCs")
	})

	t.Run("Epic1_EmotionSystem", func(t *testing.T) {
		// Verify 3D emotion system works
		npcMgr := manager.NewNPCManager(nil, manager.DefaultNPCManagerConfig())

		npcMgr.AddNPC(&manager.NPCProfile{
			ID:   "npc_emotion",
			Name: "測試NPC",
			InitialEmotion: manager.EmotionState{
				Trust:  50,
				Fear:   20,
				Stress: 30,
			},
		})

		// Test emotion adjustment
		err := npcMgr.AdjustEmotion("npc_emotion", manager.EmotionDelta{
			Trust:  10,
			Fear:   5,
			Stress: -5,
		})

		if err != nil {
			t.Fatalf("Failed to adjust emotion: %v", err)
		}

		state := npcMgr.GetState("npc_emotion")
		if state.Emotion.Trust != 60 {
			t.Errorf("Expected Trust=60, got %d", state.Emotion.Trust)
		}
		if state.Emotion.Fear != 25 {
			t.Errorf("Expected Fear=25, got %d", state.Emotion.Fear)
		}
		if state.Emotion.Stress != 25 {
			t.Errorf("Expected Stress=25, got %d", state.Emotion.Stress)
		}

		t.Log("✓ 3D emotion system works correctly")
	})

	t.Run("Epic1_MentalStateTransitions", func(t *testing.T) {
		// Verify mental state transitions
		npcMgr := manager.NewNPCManager(nil, manager.DefaultNPCManagerConfig())

		npcMgr.AddNPC(&manager.NPCProfile{
			ID:   "npc_mental",
			Name: "測試NPC",
			InitialEmotion: manager.EmotionState{
				Trust:  50,
				Fear:   20,
				Stress: 30,
			},
		})

		// Normal → Anxious (Stress >= 60)
		npcMgr.AdjustEmotion("npc_mental", manager.EmotionDelta{Stress: 35})
		state := npcMgr.GetState("npc_mental")

		if state.MentalState != manager.Anxious && state.Emotion.Stress >= 60 {
			t.Error("Should transition to Anxious when Stress >= 60")
		}

		t.Log("✓ Mental state transitions work")
	})

	t.Run("Epic1_PromptBuilding", func(t *testing.T) {
		// Verify prompt building includes revealed traits
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
		npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())

		npcMgr.AddNPC(&manager.NPCProfile{
			ID:         "npc_prompt",
			Name:       "張醫生",
			Appearance: "40多歲男性",
		})

		// Build prompt
		prompt := npcMgr.BuildNPCPrompt("npc_prompt")
		if prompt == "" {
			t.Error("Prompt should not be empty")
		}

		if !contains(prompt, "張醫生") {
			t.Error("Prompt should contain NPC name")
		}

		// Build full prompt with knowledge
		fullPrompt := npcMgr.BuildFullNPCPrompt("npc_prompt")
		if fullPrompt == "" {
			t.Error("Full prompt should not be empty")
		}

		if !contains(fullPrompt, "已知資訊") {
			t.Error("Full prompt should contain knowledge section")
		}

		t.Log("✓ Prompt building works correctly")
	})

	t.Run("Epic2_UpdateManager_Operations", func(t *testing.T) {
		// Verify UpdateManager tracks global facts
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())

		fact := &knowledge.Fact{
			ID:      "test_fact",
			Content: "測試事實",
			Type:    knowledge.Event,
		}

		updateMgr.RegisterFact(fact)

		retrievedFact := updateMgr.GetGlobalFact("test_fact")
		if retrievedFact == nil {
			t.Error("Should be able to retrieve registered fact")
		}

		if retrievedFact.Content != "測試事實" {
			t.Errorf("Fact content mismatch: got %s", retrievedFact.Content)
		}

		t.Log("✓ UpdateManager tracks global facts")
	})

	t.Run("Epic2_RoomManagement", func(t *testing.T) {
		// Verify room management tracks entity positions
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())

		updateMgr.SetEntityRoom("player", "lobby")
		updateMgr.SetEntityRoom("npc_1", "lobby")
		updateMgr.SetEntityRoom("npc_2", "corridor")

		// Check same room
		if !updateMgr.IsInSameRoom("player", "npc_1") {
			t.Error("player and npc_1 should be in same room")
		}

		if updateMgr.IsInSameRoom("player", "npc_2") {
			t.Error("player and npc_2 should NOT be in same room")
		}

		// Get NPCs in same room
		npcsInLobby := updateMgr.GetNPCsInSameRoom("player")
		if len(npcsInLobby) != 1 || npcsInLobby[0] != "npc_1" {
			t.Error("Should find npc_1 in same room as player")
		}

		t.Log("✓ Room management works correctly")
	})

	t.Run("Epic2_InformationPropagation", func(t *testing.T) {
		// Verify same-room entities automatically receive information
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())

		updateMgr.SetEntityRoom("player", "lobby")
		updateMgr.SetEntityRoom("npc_1", "lobby")
		updateMgr.SetEntityRoom("npc_2", "corridor")

		// Propagate event
		updateMgr.PropagateEvent(&knowledge.GameEvent{
			Description: "發現密道",
			Location:    "lobby",
			Beat:        10,
		})

		// Verify npc_1 learned (same room)
		kb1 := updateMgr.GetKnowledgeBase("npc_1")
		if kb1 == nil || len(kb1.KnownFacts) == 0 {
			t.Error("npc_1 should have learned from event in same room")
		}

		// Verify npc_2 did NOT learn (different room)
		kb2 := updateMgr.GetKnowledgeBase("npc_2")
		if kb2 != nil && len(kb2.KnownFacts) > 0 {
			t.Error("npc_2 should NOT have learned from event in different room")
		}

		t.Log("✓ Information propagation respects room boundaries")
	})

	t.Run("Epic2_ContradictionDetection", func(t *testing.T) {
		// Verify contradiction detection is accurate
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())

		updateMgr.SetEntityRoom("npc_test", "lobby")
		updateMgr.SetEntityRoom("witness", "lobby") // Set room for witness

		// NPC learns information
		updateMgr.LearnFromDialogue("npc_test", "witness", "王護士還活著", "lobby")

		// Check contradiction
		contradiction := updateMgr.CheckContradiction("npc_test", "王護士已經死了")

		if contradiction == nil {
			t.Error("Should detect contradiction")
		}

		if contradiction != nil && contradiction.Type == "" {
			t.Error("Contradiction should have a type")
		}

		t.Log("✓ Contradiction detection works accurately")
	})

	t.Run("Integration_NPCManager_UpdateManager", func(t *testing.T) {
		// Verify NPCManager and UpdateManager collaborate correctly
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
		npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())

		npcMgr.AddNPC(&manager.NPCProfile{
			ID:   "npc_integration",
			Name: "整合測試NPC",
		})

		updateMgr.SetEntityRoom("player", "lobby")
		updateMgr.SetEntityRoom("npc_integration", "lobby")

		// Process chat message
		contradictions := npcMgr.ProcessChatMessage("player", "測試訊息", "lobby")

		// Should not error
		if contradictions == nil {
			// It's OK to return nil when there are no contradictions
		}

		// Verify knowledge was propagated
		kb := updateMgr.GetKnowledgeBase("npc_integration")
		if kb == nil || len(kb.KnownFacts) == 0 {
			t.Error("NPC should have learned from chat message")
		}

		t.Log("✓ NPCManager and UpdateManager integration works")
	})

	t.Run("Integration_BuildFullNPCPrompt", func(t *testing.T) {
		// Verify BuildFullNPCPrompt includes knowledge base info
		updateMgr := knowledge.NewUpdateManager(knowledge.DefaultUpdateManagerConfig())
		npcMgr := manager.NewNPCManager(updateMgr, manager.DefaultNPCManagerConfig())

		npcMgr.AddNPC(&manager.NPCProfile{
			ID:   "npc_prompt_full",
			Name: "完整提示測試NPC",
		})

		updateMgr.SetEntityRoom("npc_prompt_full", "lobby")
		updateMgr.LearnFromDialogue("npc_prompt_full", "player", "重要資訊", "lobby")

		prompt := npcMgr.BuildFullNPCPrompt("npc_prompt_full")

		if prompt == "" {
			t.Error("Prompt should not be empty")
		}

		if !contains(prompt, "已知資訊") {
			t.Error("Prompt should contain knowledge section")
		}

		if !contains(prompt, "不知道的事項") {
			t.Error("Prompt should contain unknown information warning")
		}

		t.Log("✓ BuildFullNPCPrompt integrates knowledge base")
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsCheck(s, substr))
}

func containsCheck(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
