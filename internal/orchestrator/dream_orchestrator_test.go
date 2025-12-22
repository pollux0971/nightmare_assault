package orchestrator

import (
	"context"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// MockDreamGenerator is a mock implementation of DreamGenerator for testing.
type MockDreamGenerator struct{}

func (m *MockDreamGenerator) GenerateOpeningDream(ctx context.Context, theme, rulesSummary, playerRole string) (string, error) {
	return "你在一片迷霧中行走，前方的道路若隱若現。鏡子中的倒影做著與你相反的動作，時鐘的指針逆時針旋轉。你感到一陣不安，彷彿有什麼規則被打破了...", nil
}

func (m *MockDreamGenerator) GenerateChapterDream(ctx context.Context, dreamType engine.ChapterDreamType, context engine.ChapterDreamContext) (string, error) {
	return "夢境中，你看到走廊盡頭的紅門緩緩打開，裡面傳來微弱的呼救聲。你想要靠近，但雙腳卻無法移動。陰影在牆上扭曲舞動...", nil
}

func (m *MockDreamGenerator) CreateDreamRecord(
	dreamType game.DreamType,
	content string,
	relatedRuleID string,
	context game.DreamContext,
) game.DreamRecord {
	return game.DreamRecord{
		ID:            "test-dream",
		Type:          dreamType,
		Content:       content,
		RelatedRuleID: relatedRuleID,
		Context:       context,
	}
}

// TestGenerateOpeningDream tests opening dream generation.
func TestGenerateOpeningDream(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme: "廢棄醫院",
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
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test
	ctx := context.Background()
	blueprint, err := orchestrator.GenerateOpeningDream(ctx)

	// Verify
	if err != nil {
		t.Fatalf("GenerateOpeningDream failed: %v", err)
	}

	if blueprint == nil {
		t.Fatal("Blueprint is nil")
	}

	if blueprint.Type != "opening" {
		t.Errorf("Expected type 'opening', got '%s'", blueprint.Type)
	}

	if blueprint.Clarity < 0.2 || blueprint.Clarity > 0.4 {
		t.Errorf("Expected clarity 0.2-0.4, got %.2f", blueprint.Clarity)
	}

	if blueprint.TriggerBeat != 0 {
		t.Errorf("Expected TriggerBeat 0, got %d", blueprint.TriggerBeat)
	}

	if len(blueprint.Content) < 200 {
		t.Errorf("Expected content length >= 200, got %d", len(blueprint.Content))
	}

	if len(blueprint.RelatedRuleIDs) == 0 {
		t.Error("Expected at least one related rule ID")
	}

	if blueprint.IsTriggered {
		t.Error("Expected IsTriggered to be false")
	}
}

// TestGenerateChapterDreams tests chapter dream generation.
func TestGenerateChapterDreams(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme:  "廢棄醫院",
		TotalBeats: 20,
		HiddenRules: []*HiddenRule{
			{ID: "rule-1", Name: "鏡像規則"},
			{ID: "rule-2", Name: "時間規則"},
			{ID: "rule-3", Name: "門禁規則"},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test
	ctx := context.Background()
	count := 3
	blueprints, err := orchestrator.GenerateChapterDreams(ctx, count)

	// Verify
	if err != nil {
		t.Fatalf("GenerateChapterDreams failed: %v", err)
	}

	if len(blueprints) != count {
		t.Fatalf("Expected %d blueprints, got %d", count, len(blueprints))
	}

	// Verify each blueprint
	for i, blueprint := range blueprints {
		if blueprint.Type != "chapter" {
			t.Errorf("Blueprint %d: Expected type 'chapter', got '%s'", i, blueprint.Type)
		}

		if blueprint.TriggerBeat == 0 {
			t.Errorf("Blueprint %d: TriggerBeat should not be 0", i)
		}

		if blueprint.IsTriggered {
			t.Errorf("Blueprint %d: Expected IsTriggered to be false", i)
		}

		// Verify progressive clarity
		expectedClarity := 0.4 + (float64(i) * 0.2)
		if expectedClarity > 0.8 {
			expectedClarity = 0.8
		}
		if blueprint.Clarity != expectedClarity {
			t.Errorf("Blueprint %d: Expected clarity %.2f, got %.2f", i, expectedClarity, blueprint.Clarity)
		}
	}
}

// TestCheckDreamTriggers_BeatBased tests beat-based dream triggering.
func TestCheckDreamTriggers_BeatBased(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		Dreams: []*DreamBlueprint{
			{
				ID:          "dream-1",
				Type:        "chapter",
				TriggerBeat: 5,
				TriggerSAN:  0,
				IsTriggered: false,
			},
			{
				ID:          "dream-2",
				Type:        "chapter",
				TriggerBeat: 10,
				TriggerSAN:  0,
				IsTriggered: false,
			},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test: Beat 5 - should trigger dream-1
	triggered := orchestrator.CheckDreamTriggers(5, 100, "")
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream at beat 5, got %d", len(triggered))
	}
	if triggered[0].ID != "dream-1" {
		t.Errorf("Expected dream-1 to be triggered, got %s", triggered[0].ID)
	}

	// Mark dream-1 as triggered to prevent re-triggering
	storyBible.Dreams[0].IsTriggered = true

	// Test: Beat 10 - should trigger dream-2 only
	triggered = orchestrator.CheckDreamTriggers(10, 100, "")
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream at beat 10, got %d", len(triggered))
	}
	if triggered[0].ID != "dream-2" {
		t.Errorf("Expected dream-2 to be triggered, got %s", triggered[0].ID)
	}
}

// TestCheckDreamTriggers_SANBased tests SAN-based dream triggering.
func TestCheckDreamTriggers_SANBased(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		Dreams: []*DreamBlueprint{
			{
				ID:          "dream-low-san",
				Type:        "chapter",
				TriggerBeat: 0,
				TriggerSAN:  50,
				IsTriggered: false,
			},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test: SAN 40 - should trigger
	triggered := orchestrator.CheckDreamTriggers(1, 40, "")
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream at SAN 40, got %d", len(triggered))
	}

	// Test: SAN 60 - should not trigger
	triggered = orchestrator.CheckDreamTriggers(1, 60, "")
	if len(triggered) != 0 {
		t.Fatalf("Expected 0 triggered dreams at SAN 60, got %d", len(triggered))
	}
}

// TestCheckDreamTriggers_EventBased tests event-based dream triggering.
func TestCheckDreamTriggers_EventBased(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		Dreams: []*DreamBlueprint{
			{
				ID:           "dream-event",
				Type:         "chapter",
				TriggerBeat:  0,
				TriggerSAN:   0,
				TriggerEvent: "rule_violation",
				IsTriggered:  false,
			},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test: Matching event - should trigger
	triggered := orchestrator.CheckDreamTriggers(1, 100, "rule_violation")
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream for rule_violation, got %d", len(triggered))
	}

	// Test: Non-matching event - should not trigger
	triggered = orchestrator.CheckDreamTriggers(1, 100, "npc_death")
	if len(triggered) != 0 {
		t.Fatalf("Expected 0 triggered dreams for npc_death, got %d", len(triggered))
	}
}

// TestCheckDreamTriggers_SkipTriggered tests that already triggered dreams are skipped.
func TestCheckDreamTriggers_SkipTriggered(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		Dreams: []*DreamBlueprint{
			{
				ID:          "dream-triggered",
				Type:        "chapter",
				TriggerBeat: 5,
				IsTriggered: true, // Already triggered
			},
			{
				ID:          "dream-not-triggered",
				Type:        "chapter",
				TriggerBeat: 5,
				IsTriggered: false,
			},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test
	triggered := orchestrator.CheckDreamTriggers(5, 100, "")

	// Should only trigger the non-triggered dream
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 triggered dream, got %d", len(triggered))
	}
	if triggered[0].ID != "dream-not-triggered" {
		t.Errorf("Expected dream-not-triggered, got %s", triggered[0].ID)
	}
}

// TestGenerateDreamContent tests dream content generation.
func TestGenerateDreamContent(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme: "廢棄醫院",
		HiddenRules: []*HiddenRule{
			{
				ID:          "rule-1",
				Name:        "鏡像規則",
				Description: "所有行為都必須與鏡像相反",
				Hints:       []string{"注意鏡子中的倒影"},
			},
		},
	}
	gameState := engine.NewGameStateV2()
	gameState.SetSAN(45) // Low SAN for nightmare trigger

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	blueprint := &DreamBlueprint{
		ID:             "dream-test",
		Type:           "chapter",
		RelatedRuleIDs: []string{"rule-1"},
		Clarity:        0.5,
		Atmosphere:     "uneasy",
	}

	// Test
	ctx := context.Background()
	content, err := orchestrator.GenerateDreamContent(ctx, blueprint, "探索了地下室")

	// Verify
	if err != nil {
		t.Fatalf("GenerateDreamContent failed: %v", err)
	}

	if content == "" {
		t.Error("Expected non-empty content")
	}

	if len(content) < 100 {
		t.Errorf("Expected content length >= 100, got %d", len(content))
	}
}

// TestRecordDreamToGameState tests dream recording to game state.
func TestRecordDreamToGameState(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		MainTheme: "廢棄醫院",
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	blueprint := &DreamBlueprint{
		ID:             "dream-record-test",
		Type:           "chapter",
		RelatedRuleIDs: []string{"rule-1"},
		Clarity:        0.5,
		IsTriggered:    false,
	}

	content := "測試夢境內容：走廊盡頭的門緩緩打開..."

	// Test
	orchestrator.RecordDreamToGameState(blueprint, content)

	// Verify blueprint is marked as triggered
	if !blueprint.IsTriggered {
		t.Error("Expected blueprint.IsTriggered to be true")
	}

	// Verify dream is recorded in game state
	dreamLog := gameState.GetDreamLog()
	if dreamLog == nil {
		t.Fatal("DreamLog is nil")
	}

	if dreamLog.DreamCount() != 1 {
		t.Fatalf("Expected 1 dream in log, got %d", dreamLog.DreamCount())
	}

	lastDream := dreamLog.GetLastDream()
	if lastDream == nil {
		t.Fatal("LastDream is nil")
	}

	if lastDream.ID != blueprint.ID {
		t.Errorf("Expected dream ID %s, got %s", blueprint.ID, lastDream.ID)
	}

	if lastDream.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, lastDream.Content)
	}
}

// TestBuildRulesSummary tests rules summary building.
func TestBuildRulesSummary(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		HiddenRules: []*HiddenRule{
			{ID: "rule-1", Name: "鏡像規則"},
			{ID: "rule-2", Name: "時間規則"},
			{ID: "rule-3", Name: "門禁規則"},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test
	summary := orchestrator.buildRulesSummary()

	// Verify
	expected := "鏡像規則; 時間規則; 門禁規則"
	if summary != expected {
		t.Errorf("Expected summary '%s', got '%s'", expected, summary)
	}
}

// TestSelectRandomRules tests random rule selection.
func TestSelectRandomRules(t *testing.T) {
	// Setup
	mockGen := &MockDreamGenerator{}
	storyBible := &StoryBible{
		HiddenRules: []*HiddenRule{
			{ID: "rule-1", Name: "鏡像規則"},
			{ID: "rule-2", Name: "時間規則"},
			{ID: "rule-3", Name: "門禁規則"},
		},
	}
	gameState := engine.NewGameStateV2()

	orchestrator := NewDreamOrchestrator(mockGen, storyBible, gameState)

	// Test
	ruleIDs := orchestrator.selectRandomRules(1, 2)

	// Verify
	if len(ruleIDs) < 1 || len(ruleIDs) > 2 {
		t.Errorf("Expected 1-2 rule IDs, got %d", len(ruleIDs))
	}

	// Verify IDs are valid
	for _, id := range ruleIDs {
		found := false
		for _, rule := range storyBible.HiddenRules {
			if rule.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Invalid rule ID: %s", id)
		}
	}
}
