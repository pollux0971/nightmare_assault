package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DreamsCommand handles the /dreams command
type DreamsCommand struct {
	dreamLog *game.DreamLog
}

// NewDreamsCommand creates a new dreams command handler
func NewDreamsCommand(dreamLog *game.DreamLog) *DreamsCommand {
	return &DreamsCommand{
		dreamLog: dreamLog,
	}
}

// Execute runs the /dreams command
func (c *DreamsCommand) Execute() string {
	if c.dreamLog == nil || c.dreamLog.DreamCount() == 0 {
		return "你還沒有經歷過任何夢境。"
	}

	return c.formatDreamList()
}

// formatDreamList formats the list of dreams for display
func (c *DreamsCommand) formatDreamList() string {
	var b strings.Builder

	b.WriteString("=== 夢境回顧 ===\n\n")
	b.WriteString("你已經歷過以下夢境：\n\n")

	dreams := c.dreamLog.Dreams
	for i, dream := range dreams {
		dreamNum := i + 1
		dreamType := formatDreamType(dream.Type)
		chapterNum := dream.Context.ChapterNum

		b.WriteString(fmt.Sprintf("#%d - %s（第 %d 章）\n", dreamNum, dreamType, chapterNum))
	}

	b.WriteString("\n提示：輸入 /dream <編號> 重新閱讀特定夢境\n")

	return b.String()
}

// GetDreamByNumber retrieves a dream by its number (1-indexed)
func (c *DreamsCommand) GetDreamByNumber(num int) (*game.DreamRecord, error) {
	if c.dreamLog == nil {
		return nil, fmt.Errorf("no dreams available")
	}

	if num < 1 || num > c.dreamLog.DreamCount() {
		return nil, fmt.Errorf("invalid dream number: %d (must be 1-%d)", num, c.dreamLog.DreamCount())
	}

	return &c.dreamLog.Dreams[num-1], nil
}

// formatDreamType converts dream type to Chinese display name
func formatDreamType(dreamType game.DreamType) string {
	switch dreamType {
	case game.DreamTypeOpening:
		return "開場夢境"
	case game.DreamTypeChapter:
		return "章節夢境"
	default:
		return "未知夢境"
	}
}

// DreamHint represents a hint extracted from a dream
type DreamHint struct {
	Imagery     string // Dream imagery (e.g., "鏡中人做相反動作")
	RuleHint    string // What it hints at (e.g., "對立規則")
	Strength    string // "微妙", "中等", "明顯"
	Explanation string // Full explanation
}

// ExplainDreamHints analyzes a dream and extracts hints
func ExplainDreamHints(dream game.DreamRecord) []DreamHint {
	hints := []DreamHint{}

	// Simple keyword-based hint extraction
	// In a real implementation, this would use NLP or manual mapping

	content := strings.ToLower(dream.Content)

	// Mirror imagery -> Opposition rule
	if strings.Contains(content, "鏡") || strings.Contains(content, "mirror") {
		hints = append(hints, DreamHint{
			Imagery:     "鏡中的景象",
			RuleHint:    "對立規則或反向行為",
			Strength:    "中等",
			Explanation: "夢境中出現鏡子通常暗示需要進行相反的動作或選擇。",
		})
	}

	// Clock imagery -> Time rule
	if strings.Contains(content, "時鐘") || strings.Contains(content, "clock") || strings.Contains(content, "時間") {
		hints = append(hints, DreamHint{
			Imagery:     "時鐘或時間",
			RuleHint:    "時間相關規則",
			Strength:    "明顯",
			Explanation: "夢境中的時間元素通常暗示需要注意特定時刻或時間順序。",
		})
	}

	// Door imagery -> Location rule
	if strings.Contains(content, "門") || strings.Contains(content, "door") {
		hints = append(hints, DreamHint{
			Imagery:     "無法打開的門",
			RuleHint:    "場景或位置規則",
			Strength:    "微妙",
			Explanation: "夢境中的門暗示某些地點可能有特殊規則或限制。",
		})
	}

	// Shadow/darkness imagery -> Danger rule
	if strings.Contains(content, "影子") || strings.Contains(content, "黑暗") || strings.Contains(content, "shadow") {
		hints = append(hints, DreamHint{
			Imagery:     "陰影或黑暗",
			RuleHint:    "危險警告",
			Strength:    "中等",
			Explanation: "夢境中的陰影通常警告即將到來的危險或需要避免的事物。",
		})
	}

	// If no specific hints found, provide generic one
	if len(hints) == 0 {
		hints = append(hints, DreamHint{
			Imagery:     "整體氛圍",
			RuleHint:    "潛在警告",
			Strength:    "微妙",
			Explanation: "這個夢境可能包含一些較難察覺的暗示。",
		})
	}

	return hints
}

// FormatDebriefDreamAnalysis formats dream analysis for debrief view
func FormatDebriefDreamAnalysis(dreams []game.DreamRecord) string {
	if len(dreams) == 0 {
		return "你在這次遊戲中沒有經歷任何夢境。"
	}

	var b strings.Builder

	b.WriteString("=== 夢境解析 ===\n\n")
	b.WriteString("回顧你的夢境，它們曾試圖警告你：\n\n")

	for i, dream := range dreams {
		dreamNum := i + 1
		dreamType := formatDreamType(dream.Type)

		b.WriteString(fmt.Sprintf("夢境 #%d - %s（第 %d 章）\n", dreamNum, dreamType, dream.Context.ChapterNum))
		b.WriteString(fmt.Sprintf("內容摘要：%s\n\n", truncateString(dream.Content, 100)))

		// Analyze hints
		hints := ExplainDreamHints(dream)
		if len(hints) > 0 {
			b.WriteString("暗示解析：\n")
			for _, hint := range hints {
				b.WriteString(fmt.Sprintf("  • %s → %s（強度：%s）\n", hint.Imagery, hint.RuleHint, hint.Strength))
				b.WriteString(fmt.Sprintf("    %s\n", hint.Explanation))
			}
		}

		if dream.RelatedRuleID != "" {
			b.WriteString(fmt.Sprintf("  關聯規則：%s\n", dream.RelatedRuleID))
		}

		b.WriteString("\n")
	}

	b.WriteString("💡 提示：你本可以從這些夢境中察覺即將發生的危險...\n")

	return b.String()
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
