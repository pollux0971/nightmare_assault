package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// TestDreamsCommand_NoDreams tests the dreams command when no dreams have been experienced.
func TestDreamsCommand_NoDreams(t *testing.T) {
	// Setup
	gameState := engine.NewGameStateV2()
	cmd := NewDreamsCommand(gameState)

	// Test
	result, err := cmd.Execute([]string{})

	// Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "你還沒有經歷過任何夢境") {
		t.Errorf("Expected message about no dreams, got: %s", result)
	}

	if !strings.Contains(result, "夢境記錄") {
		t.Errorf("Expected header '夢境記錄', got: %s", result)
	}
}

// TestDreamsCommand_SingleDream tests the dreams command with one dream.
func TestDreamsCommand_SingleDream(t *testing.T) {
	// Setup
	gameState := engine.NewGameStateV2()

	// Add a dream
	dream := game.DreamRecord{
		ID:            "dream-1",
		Type:          game.DreamTypeOpening,
		Timestamp:     time.Now(),
		Content:       "你在一片迷霧中行走，前方的道路若隱若現。鏡子中的倒影做著與你相反的動作，你感到一陣不安...",
		RelatedRuleID: "rule-mirror",
		Context: game.DreamContext{
			PlayerHP:     100,
			PlayerSAN:    80,
			ChapterNum:   0,
			KnownClues:   []string{},
			StoryTheme:   "廢棄醫院",
			RulesSummary: "鏡像規則; 時間規則",
		},
	}

	gameState.RecordDream(dream)

	cmd := NewDreamsCommand(gameState)

	// Test
	result, err := cmd.Execute([]string{})

	// Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check header
	if !strings.Contains(result, "已經歷 1 個夢境") {
		t.Errorf("Expected '已經歷 1 個夢境', got: %s", result)
	}

	// Check dream type
	if !strings.Contains(result, "開場夢境") {
		t.Errorf("Expected '開場夢境', got: %s", result)
	}

	// Check dream number
	if !strings.Contains(result, "#1") {
		t.Errorf("Expected dream number '#1', got: %s", result)
	}

	// Check beat number
	if !strings.Contains(result, "第 0 章") {
		t.Errorf("Expected '第 0 章', got: %s", result)
	}

	// Check content preview
	if !strings.Contains(result, "你在一片迷霧中行走") {
		t.Errorf("Expected dream content preview, got: %s", result)
	}

	// Check related rule
	if !strings.Contains(result, "rule-mirror") {
		t.Errorf("Expected related rule ID, got: %s", result)
	}

	// Check status
	if !strings.Contains(result, "HP=100") {
		t.Errorf("Expected HP status, got: %s", result)
	}

	if !strings.Contains(result, "SAN=80") {
		t.Errorf("Expected SAN status, got: %s", result)
	}
}

// TestDreamsCommand_MultipleDreams tests the dreams command with multiple dreams.
func TestDreamsCommand_MultipleDreams(t *testing.T) {
	// Setup
	gameState := engine.NewGameStateV2()

	// Add opening dream
	dream1 := game.DreamRecord{
		ID:            "dream-opening",
		Type:          game.DreamTypeOpening,
		Timestamp:     time.Now(),
		Content:       "開場夢境內容：迷霧中的世界...",
		RelatedRuleID: "rule-1",
		Context: game.DreamContext{
			PlayerHP:   100,
			PlayerSAN:  100,
			ChapterNum: 0,
		},
	}

	// Add chapter dream
	dream2 := game.DreamRecord{
		ID:            "dream-chapter-1",
		Type:          game.DreamTypeChapter,
		Timestamp:     time.Now().Add(time.Hour),
		Content:       "章節夢境內容：走廊盡頭的紅門...",
		RelatedRuleID: "rule-2",
		Context: game.DreamContext{
			PlayerHP:   80,
			PlayerSAN:  45,
			ChapterNum: 5,
		},
	}

	// Add another chapter dream
	dream3 := game.DreamRecord{
		ID:            "dream-chapter-2",
		Type:          game.DreamTypeChapter,
		Timestamp:     time.Now().Add(2 * time.Hour),
		Content:       "章節夢境內容：陰影在牆上扭曲舞動...",
		RelatedRuleID: "",
		Context: game.DreamContext{
			PlayerHP:   60,
			PlayerSAN:  30,
			ChapterNum: 10,
		},
	}

	gameState.RecordDream(dream1)
	gameState.RecordDream(dream2)
	gameState.RecordDream(dream3)

	cmd := NewDreamsCommand(gameState)

	// Test
	result, err := cmd.Execute([]string{})

	// Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check count
	if !strings.Contains(result, "已經歷 3 個夢境") {
		t.Errorf("Expected '已經歷 3 個夢境', got: %s", result)
	}

	// Check all dream numbers are present
	if !strings.Contains(result, "#1") || !strings.Contains(result, "#2") || !strings.Contains(result, "#3") {
		t.Errorf("Expected all dream numbers, got: %s", result)
	}

	// Check opening dream type
	if !strings.Contains(result, "開場夢境") {
		t.Errorf("Expected '開場夢境', got: %s", result)
	}

	// Check chapter dream type (should appear at least twice)
	chapterCount := strings.Count(result, "章節夢境")
	if chapterCount < 2 {
		t.Errorf("Expected at least 2 '章節夢境', got %d", chapterCount)
	}

	// Check different beats
	if !strings.Contains(result, "第 0 章") {
		t.Errorf("Expected '第 0 章', got: %s", result)
	}
	if !strings.Contains(result, "第 5 章") {
		t.Errorf("Expected '第 5 章', got: %s", result)
	}
	if !strings.Contains(result, "第 10 章") {
		t.Errorf("Expected '第 10 章', got: %s", result)
	}

	// Check that dream without related rule doesn't show rule line
	lines := strings.Split(result, "\n")
	dream3Section := false
	hasRuleInDream3 := false
	for _, line := range lines {
		if strings.Contains(line, "#3") {
			dream3Section = true
		}
		if dream3Section && strings.Contains(line, "關聯規則") {
			hasRuleInDream3 = true
			break
		}
		if dream3Section && strings.Contains(line, "#") && !strings.Contains(line, "#3") {
			break // Moved to next dream
		}
	}
	if hasRuleInDream3 {
		t.Error("Expected no related rule for dream #3")
	}
}

// TestDreamsCommand_LongContent tests content truncation.
func TestDreamsCommand_LongContent(t *testing.T) {
	// Setup
	gameState := engine.NewGameStateV2()

	// Create a dream with long content (>100 chars)
	longContent := strings.Repeat("這是一段很長的夢境內容。", 20) // ~240 chars

	dream := game.DreamRecord{
		ID:            "dream-long",
		Type:          game.DreamTypeChapter,
		Timestamp:     time.Now(),
		Content:       longContent,
		RelatedRuleID: "",
		Context: game.DreamContext{
			PlayerHP:   70,
			PlayerSAN:  50,
			ChapterNum: 3,
		},
	}

	gameState.RecordDream(dream)

	cmd := NewDreamsCommand(gameState)

	// Test
	result, err := cmd.Execute([]string{})

	// Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should contain truncation indicator
	if !strings.Contains(result, "...") {
		t.Error("Expected content truncation '...'")
	}

	// Should not contain the full content
	if strings.Contains(result, longContent) {
		t.Error("Expected content to be truncated, but found full content")
	}
}

// TestDreamsCommand_NilGameState tests handling of nil game state.
func TestDreamsCommand_NilGameState(t *testing.T) {
	// Setup
	cmd := NewDreamsCommand(nil)

	// Test
	result, err := cmd.Execute([]string{})

	// Verify
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "遊戲狀態未初始化") {
		t.Errorf("Expected error message about uninitialized state, got: %s", result)
	}
}

// TestDreamsCommand_Name tests the command name.
func TestDreamsCommand_Name(t *testing.T) {
	cmd := NewDreamsCommand(nil)

	if cmd.Name() != "dreams" {
		t.Errorf("Expected name 'dreams', got '%s'", cmd.Name())
	}
}

// TestDreamsCommand_Usage tests the command usage.
func TestDreamsCommand_Usage(t *testing.T) {
	cmd := NewDreamsCommand(nil)

	usage := cmd.Usage()
	if usage != "/dreams" {
		t.Errorf("Expected usage '/dreams', got '%s'", usage)
	}
}

// TestDreamsCommand_Help tests the command help text.
func TestDreamsCommand_Help(t *testing.T) {
	cmd := NewDreamsCommand(nil)

	help := cmd.Help()
	if help == "" {
		t.Error("Expected non-empty help text")
	}

	if !strings.Contains(help, "夢境") {
		t.Errorf("Expected help to mention '夢境', got: %s", help)
	}
}

// TestDreamsCommand_Description tests the command description.
func TestDreamsCommand_Description(t *testing.T) {
	cmd := NewDreamsCommand(nil)

	desc := cmd.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}

	if !strings.Contains(desc, "夢境") {
		t.Errorf("Expected description to mention '夢境', got: %s", desc)
	}
}

// TestFormatDreamType tests dream type formatting.
func TestFormatDreamType(t *testing.T) {
	tests := []struct {
		input    game.DreamType
		expected string
	}{
		{game.DreamTypeOpening, "開場夢境"},
		{game.DreamTypeChapter, "章節夢境"},
		{game.DreamType("unknown"), "未知夢境"},
	}

	for _, tt := range tests {
		result := formatDreamType(tt.input)
		if result != tt.expected {
			t.Errorf("formatDreamType(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
