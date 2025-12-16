package context

import (
	"strings"
	"testing"
)

// TestSummaryPromptBuilder_BuildPrompt 測試 Prompt 構建
func TestSummaryPromptBuilder_BuildPrompt(t *testing.T) {
	builder := NewSummaryPromptBuilder()

	entries := []HistoryEntry{
		{
			Beat:           1,
			PlayerChoice:   "進入大廳",
			StoryContent:   "你看到血跡...",
			HPChange:       0,
			SANChange:      -5,
			RulesTriggered: []string{"no-look-back"},
			CluesFound:     []string{"blood-trace"},
		},
		{
			Beat:           2,
			PlayerChoice:   "調查血跡",
			StoryContent:   "血跡通往地下室...",
			HPChange:       -10,
			SANChange:      -10,
			RulesTriggered: []string{},
			CluesFound:     []string{"basement-key"},
		},
	}

	prompt := builder.BuildSummaryPrompt(entries)

	// 驗證 Prompt 包含關鍵元素
	if !strings.Contains(prompt, "blood-trace") {
		t.Error("Prompt should contain clue: blood-trace")
	}
	if !strings.Contains(prompt, "basement-key") {
		t.Error("Prompt should contain clue: basement-key")
	}
	if !strings.Contains(prompt, "no-look-back") {
		t.Error("Prompt should contain rule: no-look-back")
	}
	if !strings.Contains(prompt, "300 tokens") {
		t.Error("Prompt should mention 300 tokens limit")
	}
	if !strings.Contains(prompt, "Chapter") {
		t.Error("Prompt should mention Chapter format")
	}
	if !strings.Contains(prompt, "角色狀態") {
		t.Error("Prompt should request character status")
	}
	if !strings.Contains(prompt, "已知線索") {
		t.Error("Prompt should request known clues")
	}
	if !strings.Contains(prompt, "當前目標") {
		t.Error("Prompt should request current objective")
	}
}

// TestSummaryPromptBuilder_EmptyEntries 測試空歷史
func TestSummaryPromptBuilder_EmptyEntries(t *testing.T) {
	builder := NewSummaryPromptBuilder()
	prompt := builder.BuildSummaryPrompt([]HistoryEntry{})

	if !strings.Contains(prompt, "0") {
		t.Error("Prompt should indicate 0 entries")
	}
}

// TestSummaryPromptBuilder_FormatEntriesContainsAllInfo 測試條目格式化包含所有信息
func TestSummaryPromptBuilder_FormatEntriesContainsAllInfo(t *testing.T) {
	builder := NewSummaryPromptBuilder()

	entries := []HistoryEntry{
		{
			Beat:           1,
			PlayerChoice:   "Test Choice",
			StoryContent:   "Test Story",
			HPChange:       -15,
			SANChange:      -20,
			RulesTriggered: []string{"rule1", "rule2"},
			CluesFound:     []string{"clue1", "clue2"},
		},
	}

	prompt := builder.BuildSummaryPrompt(entries)

	// 驗證包含所有欄位
	if !strings.Contains(prompt, "Test Choice") {
		t.Error("Should contain player choice")
	}
	if !strings.Contains(prompt, "Test Story") {
		t.Error("Should contain story content")
	}
	if !strings.Contains(prompt, "HP") || !strings.Contains(prompt, "-15") {
		t.Error("Should contain HP change")
	}
	if !strings.Contains(prompt, "SAN") || !strings.Contains(prompt, "-20") {
		t.Error("Should contain SAN change")
	}
	if !strings.Contains(prompt, "rule1") || !strings.Contains(prompt, "rule2") {
		t.Error("Should contain all rules")
	}
	if !strings.Contains(prompt, "clue1") || !strings.Contains(prompt, "clue2") {
		t.Error("Should contain all clues")
	}
}
