package engine

import (
	"testing"
	"time"
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

	if config.TimeoutFirstToken != 5*time.Second {
		t.Errorf("TimeoutFirstToken = %v, want 5s", config.TimeoutFirstToken)
	}
	if config.TimeoutTotal != 30*time.Second {
		t.Errorf("TimeoutTotal = %v, want 30s", config.TimeoutTotal)
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

func TestParseChoices(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Chinese format",
			content: `故事內容...

**選擇：**
1. 進入房間
2. 離開這裡
3. 仔細觀察`,
			expected: []string{"進入房間", "離開這裡", "仔細觀察"},
		},
		{
			name: "English format",
			content: `Story content...

**Choices:**
1. Enter the room
2. Leave this place`,
			expected: []string{"Enter the room", "Leave this place"},
		},
		{
			name: "With Chinese header and separator",
			content: `Content...

**選擇：**
1、選項一
2、選項二`,
			expected: []string{"選項一", "選項二"},
		},
		{
			name: "No choices",
			content: `Story without choices.`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			choices := parseChoices(tt.content)
			if len(choices) != len(tt.expected) {
				t.Errorf("parseChoices returned %d choices, want %d", len(choices), len(tt.expected))
				return
			}
			for i, choice := range choices {
				if choice != tt.expected[i] {
					t.Errorf("Choice %d = %q, want %q", i, choice, tt.expected[i])
				}
			}
		})
	}
}

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

func TestHelperFunctions(t *testing.T) {
	// splitLines
	lines := splitLines("a\nb\nc")
	if len(lines) != 3 {
		t.Errorf("splitLines returned %d lines, want 3", len(lines))
	}

	// trimSpace
	trimmed := trimSpace("  hello  ")
	if trimmed != "hello" {
		t.Errorf("trimSpace = %q, want 'hello'", trimmed)
	}

	// toLower
	lower := toLower("HELLO")
	if lower != "hello" {
		t.Errorf("toLower = %q, want 'hello'", lower)
	}

	// contains
	if !contains("hello world", "world") {
		t.Error("contains should return true")
	}
	if contains("hello", "world") {
		t.Error("contains should return false")
	}
}

func TestContainsChoiceHeader(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"選擇：", true},
		{"**選擇：**", true},
		{"Choices:", true},
		{"options", true},
		{"選項", true},
		{"random text", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := containsChoiceHeader(tt.line); got != tt.expected {
			t.Errorf("containsChoiceHeader(%q) = %v, want %v", tt.line, got, tt.expected)
		}
	}
}

func TestIsNumberedChoice(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"1. Choice", true},
		{"2) Option", true},
		{"3、選項", true},
		{"Not a choice", false},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isNumberedChoice(tt.line); got != tt.expected {
			t.Errorf("isNumberedChoice(%q) = %v, want %v", tt.line, got, tt.expected)
		}
	}
}

func TestExtractChoiceText(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"1. Enter the room", "Enter the room"},
		{"2) Leave", "Leave"},
		{"3、觀察", "觀察"},
		{"ab", ""},
	}

	for _, tt := range tests {
		if got := extractChoiceText(tt.line); got != tt.expected {
			t.Errorf("extractChoiceText(%q) = %q, want %q", tt.line, got, tt.expected)
		}
	}
}
