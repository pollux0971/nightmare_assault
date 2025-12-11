package hint

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestHintTypeString(t *testing.T) {
	tests := []struct {
		hintType HintType
		expected string
	}{
		{HintTypeDirection, "探索方向"},
		{HintTypeClue, "線索提示"},
		{HintTypeRule, "危險警告"},
		{HintTypeItem, "道具暗示"},
		{HintTypeAtmosphere, "氛圍感知"},
	}

	for _, tt := range tests {
		if got := tt.hintType.String(); got != tt.expected {
			t.Errorf("HintType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestNewStateAnalyzer(t *testing.T) {
	analyzer := NewStateAnalyzer()
	if analyzer == nil {
		t.Error("NewStateAnalyzer should not return nil")
	}
}

func TestStateAnalyzerIsStuck(t *testing.T) {
	analyzer := NewStateAnalyzer()

	tests := []struct {
		turns    int
		expected bool
	}{
		{0, false},
		{4, false},
		{5, true},
		{10, true},
	}

	for _, tt := range tests {
		if got := analyzer.IsStuck(tt.turns); got != tt.expected {
			t.Errorf("IsStuck(%d) = %v, want %v", tt.turns, got, tt.expected)
		}
	}
}

func TestStateAnalyzerAnalyzePriority(t *testing.T) {
	analyzer := NewStateAnalyzer()

	// Test priority 1: Stuck player → Direction
	ctx := analyzer.Analyze("場景", nil, 6, 1, nil, nil, nil)
	if ctx.Type != HintTypeDirection {
		t.Errorf("Stuck player should get Direction hint, got %v", ctx.Type)
	}

	// Test priority 2: Missed clues → Clue
	ctx = analyzer.Analyze("場景", nil, 0, 1, []string{"clue1"}, nil, nil)
	if ctx.Type != HintTypeClue {
		t.Errorf("Player with missed clues should get Clue hint, got %v", ctx.Type)
	}

	// Test priority 3: Upcoming rules → Rule
	ctx = analyzer.Analyze("場景", nil, 0, 1, nil, []string{"rule1"}, nil)
	if ctx.Type != HintTypeRule {
		t.Errorf("Player with upcoming rules should get Rule hint, got %v", ctx.Type)
	}

	// Test priority 4: Available items → Item
	ctx = analyzer.Analyze("場景", nil, 0, 1, nil, nil, []string{"item1"})
	if ctx.Type != HintTypeItem {
		t.Errorf("Player with items should get Item hint, got %v", ctx.Type)
	}

	// Test priority 5: Default → Atmosphere
	ctx = analyzer.Analyze("場景", nil, 0, 1, nil, nil, nil)
	if ctx.Type != HintTypeAtmosphere {
		t.Errorf("Default should be Atmosphere hint, got %v", ctx.Type)
	}
}

func TestStateAnalyzerContextFields(t *testing.T) {
	analyzer := NewStateAnalyzer()

	ctx := analyzer.Analyze(
		"廢棄醫院",
		[]string{"打開門", "檢查地板"},
		3,
		5,
		[]string{"clue1", "clue2"},
		[]string{"rule1"},
		[]string{"手電筒"},
	)

	if ctx.CurrentScene != "廢棄醫院" {
		t.Errorf("CurrentScene = %s, want 廢棄醫院", ctx.CurrentScene)
	}
	if len(ctx.RecentActions) != 2 {
		t.Errorf("RecentActions count = %d, want 2", len(ctx.RecentActions))
	}
	if ctx.TurnsSinceMove != 3 {
		t.Errorf("TurnsSinceMove = %d, want 3", ctx.TurnsSinceMove)
	}
	if ctx.Chapter != 5 {
		t.Errorf("Chapter = %d, want 5", ctx.Chapter)
	}
	if len(ctx.MissedClues) != 2 {
		t.Errorf("MissedClues count = %d, want 2", len(ctx.MissedClues))
	}
	if len(ctx.UpcomingRules) != 1 {
		t.Errorf("UpcomingRules count = %d, want 1", len(ctx.UpcomingRules))
	}
	if len(ctx.AvailableItems) != 1 {
		t.Errorf("AvailableItems count = %d, want 1", len(ctx.AvailableItems))
	}
}

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()

	if gen == nil {
		t.Fatal("NewGenerator should not return nil")
	}
	if gen.analyzer == nil {
		t.Error("Generator should have analyzer")
	}
	if gen.baseCost != 10 {
		t.Errorf("Base cost = %d, want 10", gen.baseCost)
	}
	if gen.costIncrement != 5 {
		t.Errorf("Cost increment = %d, want 5", gen.costIncrement)
	}
}

func TestGeneratorGetCurrentCost(t *testing.T) {
	gen := NewGenerator()

	// First hint: base cost
	cost := gen.GetCurrentCost(1)
	if cost != 10 {
		t.Errorf("First hint cost = %d, want 10", cost)
	}

	// Use one hint
	gen.IncrementUsage(1)

	// Still 10 (need 2 usages for increment)
	cost = gen.GetCurrentCost(1)
	if cost != 10 {
		t.Errorf("After 1 usage, cost = %d, want 10", cost)
	}

	// Use second hint
	gen.IncrementUsage(1)

	// Now 15
	cost = gen.GetCurrentCost(1)
	if cost != 15 {
		t.Errorf("After 2 usages, cost = %d, want 15", cost)
	}

	// Use third hint
	gen.IncrementUsage(1)

	// Now 20 (max)
	cost = gen.GetCurrentCost(1)
	if cost != 20 {
		t.Errorf("After 3 usages, cost = %d, want 20", cost)
	}

	// Fourth hint still 20 (capped)
	gen.IncrementUsage(1)
	cost = gen.GetCurrentCost(1)
	if cost != 20 {
		t.Errorf("After 4 usages, cost = %d, want 20 (capped)", cost)
	}
}

func TestGeneratorUsageTracking(t *testing.T) {
	gen := NewGenerator()

	// Initial usage is 0
	if gen.GetUsageCount(1) != 0 {
		t.Errorf("Initial usage = %d, want 0", gen.GetUsageCount(1))
	}

	// Increment
	gen.IncrementUsage(1)
	gen.IncrementUsage(1)

	if gen.GetUsageCount(1) != 2 {
		t.Errorf("Usage after 2 increments = %d, want 2", gen.GetUsageCount(1))
	}

	// Different chapter
	gen.IncrementUsage(2)
	if gen.GetUsageCount(2) != 1 {
		t.Errorf("Chapter 2 usage = %d, want 1", gen.GetUsageCount(2))
	}

	// Chapter 1 unchanged
	if gen.GetUsageCount(1) != 2 {
		t.Errorf("Chapter 1 usage after = %d, want 2", gen.GetUsageCount(1))
	}

	// Reset chapter
	gen.ResetChapterUsage(1)
	if gen.GetUsageCount(1) != 0 {
		t.Errorf("After reset, usage = %d, want 0", gen.GetUsageCount(1))
	}
}

func TestGeneratorCanAffordHint(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		san      int
		chapter  int
		expected bool
	}{
		{100, 1, true},
		{10, 1, true},
		{9, 1, false},
		{0, 1, false},
	}

	for _, tt := range tests {
		if got := gen.CanAffordHint(tt.san, tt.chapter); got != tt.expected {
			t.Errorf("CanAffordHint(%d, %d) = %v, want %v", tt.san, tt.chapter, got, tt.expected)
		}
	}
}

func TestGeneratorGenerateWithLLMHint(t *testing.T) {
	gen := NewGenerator()
	ctx := &HintContext{
		Type:    HintTypeClue,
		Chapter: 1,
	}

	result := gen.Generate(ctx, "這是 LLM 生成的提示")

	if result.Text != "這是 LLM 生成的提示" {
		t.Errorf("Text = %s, want '這是 LLM 生成的提示'", result.Text)
	}
	if !result.Generated {
		t.Error("Generated should be true when LLM hint provided")
	}
	if result.Type != HintTypeClue {
		t.Errorf("Type = %v, want HintTypeClue", result.Type)
	}
	if result.SANCost != 10 {
		t.Errorf("SANCost = %d, want 10", result.SANCost)
	}
}

func TestGeneratorGenerateWithFallback(t *testing.T) {
	gen := NewGenerator()
	ctx := &HintContext{
		Type:    HintTypeDirection,
		Chapter: 1,
	}

	result := gen.Generate(ctx, "") // Empty LLM hint

	if result.Text == "" {
		t.Error("Fallback should provide text")
	}
	if result.Generated {
		t.Error("Generated should be false for fallback")
	}
}

func TestGeneratorFallbackHintsExist(t *testing.T) {
	gen := NewGenerator()

	types := []HintType{
		HintTypeDirection,
		HintTypeClue,
		HintTypeRule,
		HintTypeItem,
		HintTypeAtmosphere,
	}

	for _, hintType := range types {
		hints := gen.getFallbackHintsForType(hintType)
		if len(hints) == 0 {
			t.Errorf("No fallback hints for type %v", hintType)
		}
	}
}

func TestBuildHintPrompt(t *testing.T) {
	ctx := &HintContext{
		Type:           HintTypeClue,
		CurrentScene:   "廢棄醫院大廳",
		RecentActions:  []string{"打開門", "檢查櫃台"},
		MissedClues:    []string{"牆上的血跡"},
		UpcomingRules:  nil,
		AvailableItems: nil,
		TurnsSinceMove: 2,
		Chapter:        3,
	}

	prompt := BuildHintPrompt(ctx)

	// Check prompt contains key elements
	if !strings.Contains(prompt, "廢棄醫院大廳") {
		t.Error("Prompt should contain current scene")
	}
	if !strings.Contains(prompt, "打開門") {
		t.Error("Prompt should contain recent actions")
	}
	if !strings.Contains(prompt, "線索提示") {
		t.Error("Prompt should contain hint type")
	}
	if !strings.Contains(prompt, "20-50 字") {
		t.Error("Prompt should specify length constraint")
	}
	if !strings.Contains(prompt, "繁體中文") {
		t.Error("Prompt should specify Traditional Chinese")
	}
}

func TestBuildHintPromptNoActions(t *testing.T) {
	ctx := &HintContext{
		Type:          HintTypeAtmosphere,
		CurrentScene:  "走廊",
		RecentActions: nil,
		Chapter:       1,
	}

	prompt := BuildHintPrompt(ctx)

	if !strings.Contains(prompt, "無") {
		t.Error("Prompt should handle nil recent actions")
	}
}

func TestNewHintState(t *testing.T) {
	// Easy mode - enabled
	state := NewHintState(game.DifficultyEasy)
	if !state.IsEnabled() {
		t.Error("Hints should be enabled in Easy mode")
	}

	// Hard mode - enabled
	state = NewHintState(game.DifficultyHard)
	if !state.IsEnabled() {
		t.Error("Hints should be enabled in Hard mode")
	}

	// Hell mode - disabled
	state = NewHintState(game.DifficultyHell)
	if state.IsEnabled() {
		t.Error("Hints should be disabled in Hell mode")
	}
}

func TestHintStateRecordUsage(t *testing.T) {
	state := NewHintState(game.DifficultyEasy)

	// Initial state
	if state.TotalUsage != 0 {
		t.Errorf("Initial TotalUsage = %d, want 0", state.TotalUsage)
	}

	// Record usage
	state.RecordUsage(1)
	state.RecordUsage(1)
	state.RecordUsage(2)

	if state.TotalUsage != 3 {
		t.Errorf("TotalUsage = %d, want 3", state.TotalUsage)
	}
	if state.GetChapterUsage(1) != 2 {
		t.Errorf("Chapter 1 usage = %d, want 2", state.GetChapterUsage(1))
	}
	if state.GetChapterUsage(2) != 1 {
		t.Errorf("Chapter 2 usage = %d, want 1", state.GetChapterUsage(2))
	}
	if state.LastHintChapter != 2 {
		t.Errorf("LastHintChapter = %d, want 2", state.LastHintChapter)
	}
}

func TestHintStateGetChapterUsageEmpty(t *testing.T) {
	state := NewHintState(game.DifficultyEasy)

	// Untracked chapter should return 0
	if state.GetChapterUsage(99) != 0 {
		t.Error("Untracked chapter should return 0")
	}
}

func TestHintContextType(t *testing.T) {
	ctx := &HintContext{
		Type:    HintTypeRule,
		Chapter: 1,
	}

	if ctx.Type != HintTypeRule {
		t.Errorf("Type = %v, want HintTypeRule", ctx.Type)
	}
}

func TestHintResultFields(t *testing.T) {
	result := &HintResult{
		Text:      "提示文字",
		Type:      HintTypeItem,
		SANCost:   15,
		Generated: true,
	}

	if result.Text != "提示文字" {
		t.Errorf("Text = %s, want '提示文字'", result.Text)
	}
	if result.Type != HintTypeItem {
		t.Errorf("Type = %v, want HintTypeItem", result.Type)
	}
	if result.SANCost != 15 {
		t.Errorf("SANCost = %d, want 15", result.SANCost)
	}
	if !result.Generated {
		t.Error("Generated should be true")
	}
}

// Benchmark hint generation
func BenchmarkGenerateFallbackHint(b *testing.B) {
	gen := NewGenerator()
	ctx := &HintContext{
		Type:    HintTypeClue,
		Chapter: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Generate(ctx, "")
	}
}

func BenchmarkBuildHintPrompt(b *testing.B) {
	ctx := &HintContext{
		Type:           HintTypeClue,
		CurrentScene:   "廢棄醫院大廳",
		RecentActions:  []string{"打開門", "檢查櫃台", "觀察周圍"},
		MissedClues:    []string{"牆上的血跡", "地上的腳印"},
		UpcomingRules:  []string{"rule-1"},
		AvailableItems: []string{"手電筒", "鑰匙"},
		TurnsSinceMove: 3,
		Chapter:        5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHintPrompt(ctx)
	}
}
