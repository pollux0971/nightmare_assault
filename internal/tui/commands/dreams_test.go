package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestDreamsCommand_Execute_Empty(t *testing.T) {
	dreamLog := game.NewDreamLog()
	cmd := NewDreamsCommand(dreamLog)

	result := cmd.Execute()

	if !strings.Contains(result, "還沒有經歷過任何夢境") {
		t.Errorf("Expected empty message, got: %s", result)
	}
}

func TestDreamsCommand_Execute_WithDreams(t *testing.T) {
	dreamLog := game.NewDreamLog()

	// Add test dreams
	dreamLog.LogDream(game.DreamRecord{
		ID:        "dream-1",
		Type:      game.DreamTypeOpening,
		Timestamp: time.Now(),
		Content:   "Test opening dream",
		Context: game.DreamContext{
			ChapterNum: 1,
		},
	})

	dreamLog.LogDream(game.DreamRecord{
		ID:        "dream-2",
		Type:      game.DreamTypeChapter,
		Timestamp: time.Now(),
		Content:   "Test chapter dream",
		Context: game.DreamContext{
			ChapterNum: 2,
		},
	})

	cmd := NewDreamsCommand(dreamLog)
	result := cmd.Execute()

	if !strings.Contains(result, "夢境回顧") {
		t.Error("Expected dream review header")
	}

	if !strings.Contains(result, "#1") {
		t.Error("Expected dream #1 in list")
	}

	if !strings.Contains(result, "#2") {
		t.Error("Expected dream #2 in list")
	}

	if !strings.Contains(result, "開場夢境") {
		t.Error("Expected opening dream type")
	}

	if !strings.Contains(result, "章節夢境") {
		t.Error("Expected chapter dream type")
	}
}

func TestDreamsCommand_GetDreamByNumber(t *testing.T) {
	dreamLog := game.NewDreamLog()

	dream1 := game.DreamRecord{
		ID:      "dream-1",
		Type:    game.DreamTypeOpening,
		Content: "First dream",
		Context: game.DreamContext{ChapterNum: 1},
	}

	dream2 := game.DreamRecord{
		ID:      "dream-2",
		Type:    game.DreamTypeChapter,
		Content: "Second dream",
		Context: game.DreamContext{ChapterNum: 2},
	}

	dreamLog.LogDream(dream1)
	dreamLog.LogDream(dream2)

	cmd := NewDreamsCommand(dreamLog)

	// Test valid numbers
	result, err := cmd.GetDreamByNumber(1)
	if err != nil {
		t.Fatalf("GetDreamByNumber(1) error = %v", err)
	}
	if result.ID != "dream-1" {
		t.Errorf("Expected dream-1, got %s", result.ID)
	}

	result, err = cmd.GetDreamByNumber(2)
	if err != nil {
		t.Fatalf("GetDreamByNumber(2) error = %v", err)
	}
	if result.ID != "dream-2" {
		t.Errorf("Expected dream-2, got %s", result.ID)
	}

	// Test invalid numbers
	_, err = cmd.GetDreamByNumber(0)
	if err == nil {
		t.Error("Expected error for dream number 0")
	}

	_, err = cmd.GetDreamByNumber(3)
	if err == nil {
		t.Error("Expected error for dream number 3")
	}
}

func TestExplainDreamHints(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHints []string
	}{
		{
			name:          "Mirror imagery",
			content:       "你看到鏡子中的自己做著相反的動作",
			expectedHints: []string{"鏡中的景象", "對立規則"},
		},
		{
			name:          "Clock imagery",
			content:       "時鐘停在午夜時分",
			expectedHints: []string{"時鐘或時間", "時間相關規則"},
		},
		{
			name:          "Door imagery",
			content:       "那扇門無論如何都打不開",
			expectedHints: []string{"無法打開的門", "場景或位置規則"},
		},
		{
			name:          "Shadow imagery",
			content:       "影子在牆上蠕動",
			expectedHints: []string{"陰影或黑暗", "危險"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dream := game.DreamRecord{
				Content: tt.content,
			}

			hints := ExplainDreamHints(dream)

			if len(hints) == 0 {
				t.Error("Expected at least one hint")
			}

			// Check if expected hints are present
			for _, expectedHint := range tt.expectedHints {
				found := false
				for _, hint := range hints {
					if strings.Contains(hint.Imagery, expectedHint) || strings.Contains(hint.RuleHint, expectedHint) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected hint containing '%s' not found", expectedHint)
				}
			}
		})
	}
}

func TestFormatDebriefDreamAnalysis(t *testing.T) {
	dreams := []game.DreamRecord{
		{
			ID:      "dream-1",
			Type:    game.DreamTypeOpening,
			Content: "你看到鏡子中的自己",
			Context: game.DreamContext{ChapterNum: 1},
			RelatedRuleID: "rule-mirror",
		},
		{
			ID:      "dream-2",
			Type:    game.DreamTypeChapter,
			Content: "時鐘停在午夜",
			Context: game.DreamContext{ChapterNum: 2},
			RelatedRuleID: "rule-time",
		},
	}

	result := FormatDebriefDreamAnalysis(dreams)

	if !strings.Contains(result, "夢境解析") {
		t.Error("Expected dream analysis header")
	}

	if !strings.Contains(result, "夢境 #1") {
		t.Error("Expected dream #1")
	}

	if !strings.Contains(result, "夢境 #2") {
		t.Error("Expected dream #2")
	}

	if !strings.Contains(result, "暗示解析") {
		t.Error("Expected hint analysis")
	}

	if !strings.Contains(result, "rule-mirror") {
		t.Error("Expected related rule ID")
	}
}

func TestFormatDebriefDreamAnalysis_Empty(t *testing.T) {
	dreams := []game.DreamRecord{}

	result := FormatDebriefDreamAnalysis(dreams)

	if !strings.Contains(result, "沒有經歷任何夢境") {
		t.Errorf("Expected empty message, got: %s", result)
	}
}
