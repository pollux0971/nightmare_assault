package engine

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/builder"
)

func TestNewStoryState(t *testing.T) {
	state := NewStoryState()

	if state.CurrentBeat != 0 {
		t.Errorf("CurrentBeat = %d, want 0", state.CurrentBeat)
	}
	if state.TotalHP != 100 {
		t.Errorf("TotalHP = %d, want 100", state.TotalHP)
	}
	if state.TotalSAN != 100 {
		t.Errorf("TotalSAN = %d, want 100", state.TotalSAN)
	}
	if len(state.History) != 0 {
		t.Errorf("History should be empty")
	}
	if len(state.ActiveSeeds) != 0 {
		t.Errorf("ActiveSeeds should be empty")
	}
}

func TestStoryState_AddSegment(t *testing.T) {
	state := NewStoryState()

	segment := StorySegment{
		Content:   "Test content",
		Choices:   []string{"Choice 1", "Choice 2"},
		Timestamp: time.Now(),
	}

	state.AddSegment(segment)

	if state.CurrentBeat != 1 {
		t.Errorf("CurrentBeat = %d, want 1", state.CurrentBeat)
	}
	if len(state.History) != 1 {
		t.Errorf("History length = %d, want 1", len(state.History))
	}
	if state.History[0].Content != "Test content" {
		t.Errorf("Segment content = %v, want 'Test content'", state.History[0].Content)
	}
}

func TestStoryState_AddSeed(t *testing.T) {
	state := NewStoryState()

	seed := HiddenSeed{
		ID:          "S01",
		Type:        SeedTypeItem,
		Description: "A rusty key",
		TriggerBeat: 5,
		Discovered:  false,
	}

	state.AddSeed(seed)

	if len(state.ActiveSeeds) != 1 {
		t.Errorf("ActiveSeeds length = %d, want 1", len(state.ActiveSeeds))
	}
	if state.ActiveSeeds[0].ID != "S01" {
		t.Errorf("Seed ID = %v, want S01", state.ActiveSeeds[0].ID)
	}
}

func TestStoryState_GetLastSegment(t *testing.T) {
	state := NewStoryState()

	// Empty state
	last := state.GetLastSegment()
	if last != nil {
		t.Error("GetLastSegment on empty state should return nil")
	}

	// With segments
	state.AddSegment(StorySegment{Content: "First"})
	state.AddSegment(StorySegment{Content: "Second"})

	last = state.GetLastSegment()
	if last == nil {
		t.Fatal("GetLastSegment should not return nil")
	}
	if last.Content != "Second" {
		t.Errorf("Last segment content = %v, want 'Second'", last.Content)
	}
}

func TestStoryState_GetContextSummary(t *testing.T) {
	state := NewStoryState()

	// Empty state
	summary := state.GetContextSummary(3)
	if summary != "" {
		t.Error("GetContextSummary on empty state should return empty string")
	}

	// With segments
	state.AddSegment(StorySegment{Content: "First segment"})
	state.AddSegment(StorySegment{Content: "Second segment"})
	state.AddSegment(StorySegment{Content: "Third segment"})
	state.AddSegment(StorySegment{Content: "Fourth segment"})

	// Get last 2
	summary = state.GetContextSummary(2)
	if len(summary) == 0 {
		t.Error("Summary should not be empty")
	}
}

func TestNewStoryEngine(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewStoryEngine(config)

	if engine == nil {
		t.Fatal("NewStoryEngine should not return nil")
	}
	if engine.storyState == nil {
		t.Error("Engine should have story state")
	}
}

func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()

	// Timeouts are disabled (0) for slow free models
	if config.TimeoutFirstToken != 0 {
		t.Errorf("TimeoutFirstToken = %v, want 0 (disabled)", config.TimeoutFirstToken)
	}
	if config.TimeoutTotal != 0 {
		t.Errorf("TimeoutTotal = %v, want 0 (disabled)", config.TimeoutTotal)
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
}

func TestStoryEngine_Reset(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewStoryEngine(config)

	// Add some state
	engine.storyState.AddSegment(StorySegment{Content: "Test"})
	engine.storyState.AddSeed(HiddenSeed{ID: "S01"})

	// Reset
	engine.Reset()

	if engine.storyState.CurrentBeat != 0 {
		t.Error("After reset, CurrentBeat should be 0")
	}
	if len(engine.storyState.History) != 0 {
		t.Error("After reset, History should be empty")
	}
	if len(engine.storyState.ActiveSeeds) != 0 {
		t.Error("After reset, ActiveSeeds should be empty")
	}
}

// TestStoryEngine_ParseJSONOutput tests JSON output parsing integration
func TestStoryEngine_ParseJSONOutput(t *testing.T) {
	jsonOutput := `{
		"story": "完整故事內容，包含詳細的場景描述。",
		"choices": ["選項1：向左走", "選項2：向右走", "選項3：原地等待"],
		"seeds": [
			{"type": "Item", "description": "測試物品"}
		]
	}`

	output, err := builder.ParseStructuredOutput(jsonOutput)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Story != "完整故事內容，包含詳細的場景描述。" {
		t.Errorf("Story mismatch. Got: %s", output.Story)
	}
	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(output.Choices))
	}
	if len(output.Seeds) != 1 {
		t.Errorf("Expected 1 seed, got %d", len(output.Seeds))
	}
	if output.Seeds[0].Type != "Item" {
		t.Errorf("Expected seed type 'Item', got '%s'", output.Seeds[0].Type)
	}
}

// TestStoryEngine_ParseLegacyFormat tests backward compatibility with legacy text format
func TestStoryEngine_ParseLegacyFormat(t *testing.T) {
	legacyOutput := `這是舊格式的故事內容。

很多描述文字在這裡。

選擇：
1. 第一個選項
2. 第二個選項
3. 第三個選項`

	output, err := builder.ParseStructuredOutput(legacyOutput)

	if err != nil {
		t.Fatalf("Expected no error with legacy format, got %v", err)
	}
	if output.Story == "" {
		t.Error("Story should not be empty")
	}
	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices from legacy format, got %d", len(output.Choices))
	}
}

// TestStoryEngine_ParsePrologueJSON tests prologue format with no choices
func TestStoryEngine_ParsePrologueJSON(t *testing.T) {
	prologueJSON := `{
		"story": "序章內容，沒有選擇。\n\n【按任意鍵繼續到第二章】",
		"choices": [],
		"seeds": []
	}`

	output, err := builder.ParseStructuredOutput(prologueJSON)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.Story == "" {
		t.Error("Story should not be empty")
	}
	if len(output.Choices) != 0 {
		t.Errorf("Expected 0 choices for prologue, got %d", len(output.Choices))
	}
}

// NOTE: Legacy parsing tests have been moved to builder/output_parser_test.go
// These functions (parseChoices, containsChoiceHeader, etc.) are now in the builder package

func TestGenerateSeedID(t *testing.T) {
	id1 := generateSeedID(0, 0)
	id2 := generateSeedID(0, 1)
	id3 := generateSeedID(1, 0)

	// IDs should be different
	if id1 == id2 {
		t.Error("Different index should produce different IDs")
	}
	if id1 == id3 {
		t.Error("Different beat should produce different IDs")
	}
}

func TestSeedType_Constants(t *testing.T) {
	// Verify seed type constants exist
	types := []SeedType{SeedTypeItem, SeedTypeEvent, SeedTypeCharacter, SeedTypeLocation}
	if len(types) != 4 {
		t.Error("Should have 4 seed types")
	}
}
