package context

import (
	"fmt"
	"strings"
)

// SummaryPromptBuilder constructs prompts for LLM summary generation.
type SummaryPromptBuilder struct{}

// NewSummaryPromptBuilder creates a new SummaryPromptBuilder.
func NewSummaryPromptBuilder() *SummaryPromptBuilder {
	return &SummaryPromptBuilder{}
}

// BuildSummaryPrompt constructs the LLM prompt for generating a summary.
// The prompt instructs the LLM to compress game history while preserving:
// - Character status (NPCs alive/dead)
// - Discovered clues
// - Known rules
// - Current objective
// - Key dialogue and events
func (b *SummaryPromptBuilder) BuildSummaryPrompt(entries []HistoryEntry) string {
	// Extract key information from entries
	clues := extractClues(entries)
	rules := extractRules(entries)
	entriesText := formatEntries(entries)

	prompt := fmt.Sprintf(`你是一個遊戲歷史壓縮助手。請將以下 %d 個回合的遊戲歷史壓縮成簡短摘要。

**必須包含：**
1. 角色存活狀態（NPC 是否還活著）
2. 已發現的線索列表: %s
3. 已知的規則列表: %s
4. 當前目標或劇情進展
5. 關鍵對話或重大事件

**格式要求：**
- 使用格式："[Chapter X Summary: 簡要劇情. 角色狀態: XX. 已知線索: YY. 當前目標: ZZ.]"
- 長度不超過 300 tokens
- 只輸出摘要，不要解釋

**遊戲歷史：**
%s

**輸出摘要：**`,
		len(entries),
		formatList(clues, "無"),
		formatList(rules, "無"),
		entriesText)

	return prompt
}

// extractClues collects all unique clues from history entries.
func extractClues(entries []HistoryEntry) []string {
	clueSet := make(map[string]bool)
	for _, entry := range entries {
		for _, clue := range entry.CluesFound {
			clueSet[clue] = true
		}
	}

	clues := make([]string, 0, len(clueSet))
	for clue := range clueSet {
		clues = append(clues, clue)
	}
	return clues
}

// extractRules collects all unique rules from history entries.
func extractRules(entries []HistoryEntry) []string {
	ruleSet := make(map[string]bool)
	for _, entry := range entries {
		for _, rule := range entry.RulesTriggered {
			ruleSet[rule] = true
		}
	}

	rules := make([]string, 0, len(ruleSet))
	for rule := range ruleSet {
		rules = append(rules, rule)
	}
	return rules
}

// formatEntries formats history entries into a readable text format.
func formatEntries(entries []HistoryEntry) string {
	if len(entries) == 0 {
		return "(空歷史)"
	}

	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(fmt.Sprintf("\n**回合 %d:**\n", entry.Beat))
		builder.WriteString(fmt.Sprintf("- 玩家選擇: %s\n", entry.PlayerChoice))
		builder.WriteString(fmt.Sprintf("- 劇情: %s\n", entry.StoryContent))

		if entry.HPChange != 0 {
			builder.WriteString(fmt.Sprintf("- HP 變化: %+d\n", entry.HPChange))
		}
		if entry.SANChange != 0 {
			builder.WriteString(fmt.Sprintf("- SAN 變化: %+d\n", entry.SANChange))
		}

		if len(entry.CluesFound) > 0 {
			builder.WriteString(fmt.Sprintf("- 發現線索: %s\n", strings.Join(entry.CluesFound, ", ")))
		}
		if len(entry.RulesTriggered) > 0 {
			builder.WriteString(fmt.Sprintf("- 觸發規則: %s\n", strings.Join(entry.RulesTriggered, ", ")))
		}
	}

	return builder.String()
}

// formatList formats a string list with a default value if empty.
func formatList(items []string, defaultVal string) string {
	if len(items) == 0 {
		return defaultVal
	}
	return strings.Join(items, ", ")
}
