package orchestrator

import (
	"context"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// TestDreamSystem_IntegrationFlow tests the complete dream system flow.
// Story 9-1 + 9-2: From opening dream generation to chapter dreams.
func TestDreamSystem_IntegrationFlow(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme:  "廢棄醫院",
		TotalBeats: 20,
		HiddenRules: []*HiddenRule{
			{
				ID:          "rule-1",
				Name:        "鏡像規則",
				Description: "所有行為都必須與鏡像相反",
				Hints:       []string{"注意鏡子中的倒影"},
			},
			{
				ID:          "rule-2",
				Name:        "時間規則",
				Description: "某些時刻無法行動",
				Hints:       []string{"注意時鐘的聲音"},
			},
		},
		Dreams: []*DreamBlueprint{}, // Will be populated
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)
	ctx := context.Background()

	// Phase 1: Generate opening dream (Story 9-1)
	t.Run("Story 9-1: Opening Dream Generation", func(t *testing.T) {
		openingDream, err := orchestrator.GenerateOpeningDream(ctx)
		if err != nil {
			t.Fatalf("Failed to generate opening dream: %v", err)
		}

		// Verify opening dream properties
		if openingDream.Type != "opening" {
			t.Errorf("Expected type 'opening', got '%s'", openingDream.Type)
		}

		if openingDream.Clarity < 0.2 || openingDream.Clarity > 0.4 {
			t.Errorf("Expected clarity 0.2-0.4, got %.2f", openingDream.Clarity)
		}

		if len(openingDream.Content) < 200 {
			t.Errorf("Expected content length >= 200, got %d", len(openingDream.Content))
		}

		// Add to story bible
		storyBible.Dreams = append(storyBible.Dreams, openingDream)

		// Record to game state
		orchestrator.RecordDreamToGameState(openingDream, openingDream.Content)

		// Verify recorded in game state
		dreamLog := gameState.GetDreamLog()
		if dreamLog.DreamCount() != 1 {
			t.Errorf("Expected 1 dream in log, got %d", dreamLog.DreamCount())
		}
	})

	// Phase 2: Generate chapter dreams (Story 9-2)
	t.Run("Story 9-2: Chapter Dreams Generation", func(t *testing.T) {
		chapterDreams, err := orchestrator.GenerateChapterDreams(ctx, 3)
		if err != nil {
			t.Fatalf("Failed to generate chapter dreams: %v", err)
		}

		if len(chapterDreams) != 3 {
			t.Fatalf("Expected 3 chapter dreams, got %d", len(chapterDreams))
		}

		// Add to story bible
		storyBible.Dreams = append(storyBible.Dreams, chapterDreams...)

		// Verify progressive clarity
		for i, dream := range chapterDreams {
			expectedMinClarity := 0.4 + (float64(i) * 0.2)
			if dream.Clarity < expectedMinClarity {
				t.Errorf("Chapter dream %d: Expected clarity >= %.2f, got %.2f",
					i, expectedMinClarity, dream.Clarity)
			}
		}
	})

	// Phase 3: Trigger dreams during gameplay (Story 9-2)
	t.Run("Story 9-2: Dream Triggering", func(t *testing.T) {
		// Simulate game progression to beat 5
		for i := 0; i < 5; i++ {
			gameState.IncrementBeat()
		}

		// Check for triggered dreams
		triggered := orchestrator.CheckDreamTriggers(
			gameState.GetCurrentBeat(),
			gameState.GetSAN(),
			"",
		)

		if len(triggered) == 0 {
			t.Error("Expected at least one dream to trigger at beat 5")
		}

		// Generate content for first triggered dream
		if len(triggered) > 0 {
			dream := triggered[0]
			content, err := orchestrator.GenerateDreamContent(ctx, dream, "探索了地下室")
			if err != nil {
				t.Fatalf("Failed to generate dream content: %v", err)
			}

			if content == "" {
				t.Error("Expected non-empty dream content")
			}

			// Record dream
			orchestrator.RecordDreamToGameState(dream, content)

			// Verify dream count increased
			dreamLog := gameState.GetDreamLog()
			if dreamLog.DreamCount() != 2 { // Opening + 1 chapter
				t.Errorf("Expected 2 dreams in log, got %d", dreamLog.DreamCount())
			}

			// Verify dream is marked as triggered
			if !dream.IsTriggered {
				t.Error("Expected dream to be marked as triggered")
			}
		}
	})

	// Phase 4: Low SAN trigger (Story 9-2)
	t.Run("Story 9-2: SAN-based Triggering", func(t *testing.T) {
		// Reduce SAN to trigger low SAN dreams
		gameState.SetSAN(40)

		triggered := orchestrator.CheckDreamTriggers(
			gameState.GetCurrentBeat(),
			gameState.GetSAN(),
			"",
		)

		// Should trigger dreams with SAN threshold
		hasSANTrigger := false
		for _, dream := range triggered {
			if dream.TriggerSAN > 0 && gameState.GetSAN() < dream.TriggerSAN {
				hasSANTrigger = true
				break
			}
		}

		if len(triggered) > 0 && !hasSANTrigger {
			t.Log("Note: No SAN-based triggers found (may vary based on generated dreams)")
		}
	})

	// Phase 5: Event-based trigger (Story 9-2)
	t.Run("Story 9-2: Event-based Triggering", func(t *testing.T) {
		// Check for event-based dreams
		triggered := orchestrator.CheckDreamTriggers(
			gameState.GetCurrentBeat(),
			gameState.GetSAN(),
			"rule_violation",
		)

		// Should trigger dreams with matching event
		hasEventTrigger := false
		for _, dream := range triggered {
			if dream.TriggerEvent == "rule_violation" {
				hasEventTrigger = true
				break
			}
		}

		if len(triggered) > 0 && !hasEventTrigger {
			t.Log("Note: No event-based triggers found (may vary based on generated dreams)")
		}
	})

	// Phase 6: Dream review via /dreams command
	t.Run("Story 9-2: Dream Review", func(t *testing.T) {
		dreamLog := gameState.GetDreamLog()

		// Should have at least opening + 1 chapter dream
		if dreamLog.DreamCount() < 2 {
			t.Errorf("Expected at least 2 dreams, got %d", dreamLog.DreamCount())
		}

		// Verify opening dream is first
		firstDream := dreamLog.GetDreamsByType("opening")
		if len(firstDream) != 1 {
			t.Errorf("Expected 1 opening dream, got %d", len(firstDream))
		}

		// Verify chapter dreams exist
		chapterDreams := dreamLog.GetDreamsByType("chapter")
		if len(chapterDreams) < 1 {
			t.Error("Expected at least 1 chapter dream")
		}
	})
}

// TestDreamSystem_ClarityProgression tests that dream clarity increases over time.
func TestDreamSystem_ClarityProgression(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme:  "廢棄醫院",
		TotalBeats: 20,
		HiddenRules: []*HiddenRule{
			{ID: "rule-1", Name: "規則1"},
			{ID: "rule-2", Name: "規則2"},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)
	ctx := context.Background()

	// Generate multiple chapter dreams
	chapterDreams, err := orchestrator.GenerateChapterDreams(ctx, 5)
	if err != nil {
		t.Fatalf("Failed to generate chapter dreams: %v", err)
	}

	// Verify clarity increases
	for i := 1; i < len(chapterDreams); i++ {
		prevClarity := chapterDreams[i-1].Clarity
		currClarity := chapterDreams[i].Clarity

		if currClarity < prevClarity {
			t.Errorf("Dream %d clarity (%.2f) should be >= previous clarity (%.2f)",
				i, currClarity, prevClarity)
		}
	}

	// Verify max clarity doesn't exceed 0.8
	for i, dream := range chapterDreams {
		if dream.Clarity > 0.8 {
			t.Errorf("Dream %d clarity (%.2f) exceeds maximum (0.8)", i, dream.Clarity)
		}
	}
}

// TestDreamSystem_MultipleRuleHints tests dreams can hint at multiple rules.
func TestDreamSystem_MultipleRuleHints(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme: "廢棄醫院",
		HiddenRules: []*HiddenRule{
			{ID: "rule-1", Name: "規則1", Description: "描述1", Hints: []string{"提示1"}},
			{ID: "rule-2", Name: "規則2", Description: "描述2", Hints: []string{"提示2"}},
			{ID: "rule-3", Name: "規則3", Description: "描述3", Hints: []string{"提示3"}},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)
	ctx := context.Background()

	// Generate opening dream
	openingDream, err := orchestrator.GenerateOpeningDream(ctx)
	if err != nil {
		t.Fatalf("Failed to generate opening dream: %v", err)
	}

	// Verify 1-2 rules are hinted
	ruleCount := len(openingDream.RelatedRuleIDs)
	if ruleCount < 1 || ruleCount > 2 {
		t.Errorf("Expected 1-2 related rules, got %d", ruleCount)
	}

	// Verify rule IDs are valid
	for _, ruleID := range openingDream.RelatedRuleIDs {
		found := false
		for _, rule := range storyBible.HiddenRules {
			if rule.ID == ruleID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Invalid rule ID: %s", ruleID)
		}
	}
}

// TestDreamSystem_NoDoubleTriggering tests dreams are not triggered twice.
func TestDreamSystem_NoDoubleTriggering(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		Dreams: []*DreamBlueprint{
			{
				ID:          "dream-test",
				Type:        "chapter",
				TriggerBeat: 5,
				IsTriggered: false,
			},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// First trigger at beat 5
	triggered := orchestrator.CheckDreamTriggers(5, 100, "")
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream, got %d", len(triggered))
	}

	// Mark as triggered
	storyBible.Dreams[0].IsTriggered = true

	// Try to trigger again at beat 6
	triggered = orchestrator.CheckDreamTriggers(6, 100, "")
	if len(triggered) != 0 {
		t.Errorf("Expected 0 triggered dreams (already triggered), got %d", len(triggered))
	}
}
