package views

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 5.4: Chat Summary Generation
// ==========================================================================

// SummaryGenerator defines the interface for generating chat summaries.
// This abstraction allows for easy testing with mock implementations.
type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, prompt string) (string, error)
}

// LLMSummaryGenerator implements SummaryGenerator using an LLM client.
type LLMSummaryGenerator struct {
	client client.Provider
}

// NewLLMSummaryGenerator creates a new LLM-based summary generator.
func NewLLMSummaryGenerator(client client.Provider) *LLMSummaryGenerator {
	return &LLMSummaryGenerator{client: client}
}

// GenerateSummary generates a summary using the LLM client.
func (g *LLMSummaryGenerator) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	messages := []client.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := g.client.SendMessage(ctx, messages)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// buildSummaryPrompt constructs a prompt for LLM to generate chat summary.
// AC4, AC5: Include conversation context, history, and output format requirements.
func (m *ChatOverlayModel) buildSummaryPrompt() string {
	var b strings.Builder

	// 1. System instruction
	b.WriteString("你是一個專業的對話分析助手。請分析以下對話並生成結構化摘要。\n\n")

	// 2. Conversation context
	b.WriteString("## 對話上下文\n")
	b.WriteString(fmt.Sprintf("- 場景: %s\n", m.location))
	b.WriteString(fmt.Sprintf("- 時間: 第 %d 回合\n", m.chatTurns))
	b.WriteString("- 參與者: ")
	participantNames := []string{}
	for _, p := range m.participants {
		if p.IsActive {
			if p.IsPlayer {
				participantNames = append(participantNames, fmt.Sprintf("%s (玩家)", p.Name))
			} else {
				participantNames = append(participantNames, fmt.Sprintf("%s (信任:%d 恐懼:%d 壓力:%d)",
					p.Name, p.Emotion.Trust, p.Emotion.Fear, p.Emotion.Stress))
			}
		}
	}
	b.WriteString(strings.Join(participantNames, ", "))
	b.WriteString("\n\n")

	// 3. Conversation history
	b.WriteString("## 對話記錄\n")
	for _, msg := range m.messages {
		if msg.Type == ChatMessageSystem {
			continue // Skip system messages
		}
		speakerName := m.getParticipantName(msg.Speaker)
		if msg.Speaker == "player" {
			speakerName = "玩家"
		}
		b.WriteString(fmt.Sprintf("[%s] %s\n", speakerName, msg.Content))

		// Include flags if present
		if len(msg.Flags) > 0 {
			flagStrs := []string{}
			for _, flag := range msg.Flags {
				flagStrs = append(flagStrs, flag.String())
			}
			b.WriteString(fmt.Sprintf("  (標記: %s)\n", strings.Join(flagStrs, ", ")))
		}
	}
	b.WriteString("\n")

	// 4. Output format requirements
	b.WriteString("## 輸出要求\n")
	b.WriteString("請生成以下 JSON 格式的摘要（narrative_impact 欄位限制在 200-400 字元）：\n")
	b.WriteString(`{
  "main_topics": ["話題1", "話題2"],
  "key_decisions": ["決策1", "決策2"],
  "relation_changes": {"npc_id": "關係變化描述"},
  "facts_shared": ["事實1", "事實2"],
  "flags": ["flag1", "flag2"],
  "narrative_impact": "對故事的影響（200-400字元）",
  "emotion_changes": {"npc_id": "情感變化摘要"},
  "unresolved_issues": ["未解決問題1"]
}`)
	b.WriteString("\n\n")

	// 5. Analysis focus
	b.WriteString("## 分析重點\n")
	b.WriteString("- 識別最重要的 2-3 個話題\n")
	b.WriteString("- 提取關鍵決策或承諾\n")
	b.WriteString("- 注意關係變化（信任增加/減少、衝突、和解等）\n")
	b.WriteString("- 記錄分享的重要資訊或秘密\n")
	b.WriteString("- 標記謊言、矛盾、揭露等特殊事件（使用 flags 欄位中的值）\n")
	b.WriteString("- 評估對主故事線的影響\n")
	b.WriteString("- 記錄情感變化軌跡\n")
	b.WriteString("- 標記未解決的懸念或問題\n\n")

	b.WriteString("請只返回 JSON，不要包含其他說明文字。\n")

	return b.String()
}

// parseSummaryResponse parses LLM response and extracts ChatSummary.
// AC1: Parse JSON response and validate required fields.
func (m *ChatOverlayModel) parseSummaryResponse(response string) (*ChatSummary, error) {
	// 1. Extract JSON from response (may be wrapped in markdown code blocks)
	jsonStr := extractJSON(response)

	// 2. Parse JSON
	var summary ChatSummary
	err := json.Unmarshal([]byte(jsonStr), &summary)
	if err != nil {
		return nil, fmt.Errorf("failed to parse summary JSON: %w", err)
	}

	// 3. Validate and set defaults for required fields
	if len(summary.MainTopics) == 0 {
		summary.MainTopics = []string{}
	}

	if summary.KeyDecisions == nil {
		summary.KeyDecisions = []string{}
	}

	if summary.NarrativeImpact == "" {
		summary.NarrativeImpact = "對話未產生明顯影響"
	}

	// 4. Initialize nil maps/slices with empty collections
	if summary.RelationChanges == nil {
		summary.RelationChanges = make(map[string]string)
	}

	if summary.FactsShared == nil {
		summary.FactsShared = []string{}
	}

	if summary.Flags == nil {
		summary.Flags = []string{}
	}

	if summary.EmotionChanges == nil {
		summary.EmotionChanges = make(map[string]string)
	}

	if summary.UnresolvedIssues == nil {
		summary.UnresolvedIssues = []string{}
	}

	return &summary, nil
}

// extractJSON extracts JSON from a response that may contain markdown formatting.
func extractJSON(response string) string {
	// Try to find ```json ... ``` code block
	re := regexp.MustCompile(`(?s)` + "```json\\s*(.+?)\\s*```")
	matches := re.FindStringSubmatch(response)

	if len(matches) > 1 {
		return matches[1]
	}

	// Try to find ``` ... ``` code block (without json marker)
	re = regexp.MustCompile(`(?s)` + "```\\s*(.+?)\\s*```")
	matches = re.FindStringSubmatch(response)

	if len(matches) > 1 {
		return matches[1]
	}

	// Try to find { ... } JSON object
	re = regexp.MustCompile(`(?s)\{.+\}`)
	matches = re.FindStringSubmatch(response)

	if len(matches) > 0 {
		return matches[0]
	}

	// Assume the entire response is JSON
	return response
}

// generateFallbackSummary generates a basic summary when LLM fails.
// AC1: Fallback summary ensures chat always has basic summary data.
func (m *ChatOverlayModel) generateFallbackSummary() *ChatSummary {
	summary := &ChatSummary{
		MainTopics:       []string{},
		KeyDecisions:     []string{},
		RelationChanges:  make(map[string]string),
		FactsShared:      []string{},
		Flags:            []string{},
		EmotionChanges:   make(map[string]string),
		UnresolvedIssues: []string{},
	}

	// Basic statistics
	// messageCount := len(m.messages) // Unused variable removed

	// Count non-system messages
	actualMessageCount := 0
	for _, msg := range m.messages {
		if msg.Type != ChatMessageSystem {
			actualMessageCount++
		}
	}

	// Handle empty conversation
	if actualMessageCount == 0 {
		summary.NarrativeImpact = "對話未開始或立即結束"
		return summary
	}

	// Extract participant names (excluding player)
	participantNames := []string{}
	for _, p := range m.participants {
		if !p.IsPlayer && p.IsActive {
			participantNames = append(participantNames, p.Name)
		}
	}

	// Generate basic impact description
	if len(participantNames) > 0 {
		summary.NarrativeImpact = fmt.Sprintf(
			"與 %s 進行了 %d 回合的對話",
			strings.Join(participantNames, "、"),
			actualMessageCount,
		)
	} else {
		summary.NarrativeImpact = fmt.Sprintf("進行了 %d 回合的對話", actualMessageCount)
	}

	// Simple keyword extraction
	keywords := m.extractKeywords()
	if len(keywords) > 0 {
		// Take top 3 keywords as topics
		topicCount := 3
		if len(keywords) < topicCount {
			topicCount = len(keywords)
		}
		summary.MainTopics = keywords[:topicCount]
	}

	// Extract flags from messages
	flagMap := make(map[string]bool)
	for _, msg := range m.messages {
		for _, flag := range msg.Flags {
			flagMap[flag.String()] = true
		}
	}
	for flag := range flagMap {
		summary.Flags = append(summary.Flags, flag)
	}

	return summary
}

// extractKeywords performs simple keyword extraction from messages.
// Used by fallback summary generation.
func (m *ChatOverlayModel) extractKeywords() []string {
	// Common words to filter out (extend this list as needed)
	commonWords := map[string]bool{
		"的": true, "了": true, "是": true, "在": true, "我": true, "有": true,
		"和": true, "就": true, "不": true, "人": true, "都": true, "一": true,
		"一個": true, "沒有": true, "我們": true, "來": true, "到": true, "時": true,
		"大": true, "地": true, "為": true, "上": true, "著": true, "過": true,
		"家": true, "十": true, "用": true, "他": true, "們": true, "會": true,
		"你": true, "嗎": true, "呢": true, "吧": true, "啊": true,
	}

	wordCount := make(map[string]int)

	for _, msg := range m.messages {
		if msg.Type == ChatMessageSystem {
			continue
		}

		// Split content into words
		// For Chinese text, we'll extract 2-3 character sequences as potential keywords
		content := msg.Content
		runes := []rune(content)

		// Extract 2-character sequences
		for i := 0; i < len(runes)-1; i++ {
			word := string(runes[i : i+2])
			if !commonWords[word] && len(word) >= 2 {
				wordCount[word]++
			}
		}

		// Extract 3-character sequences
		for i := 0; i < len(runes)-2; i++ {
			word := string(runes[i : i+3])
			if !commonWords[word] {
				wordCount[word]++
			}
		}
	}

	// Sort by frequency
	type wordFreq struct {
		word  string
		count int
	}

	freqs := []wordFreq{}
	for word, count := range wordCount {
		if count >= 2 { // Only include words that appear at least twice
			freqs = append(freqs, wordFreq{word, count})
		}
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].count > freqs[j].count
	})

	// Return top keywords
	keywords := []string{}
	maxKeywords := 5
	if len(freqs) < maxKeywords {
		maxKeywords = len(freqs)
	}
	for i := 0; i < maxKeywords; i++ {
		keywords = append(keywords, freqs[i].word)
	}

	return keywords
}

// generateSummary generates a chat summary using LLM or fallback.
// AC1: Main entry point for summary generation, called from Exit().
func (m *ChatOverlayModel) generateSummary(ctx context.Context, generator SummaryGenerator) (*ChatSummary, error) {
	// If no messages (excluding system messages), return empty fallback
	actualMessageCount := 0
	for _, msg := range m.messages {
		if msg.Type != ChatMessageSystem {
			actualMessageCount++
		}
	}

	if actualMessageCount == 0 {
		return m.generateFallbackSummary(), nil
	}

	// If generator is nil, use fallback
	if generator == nil {
		logger.Warn("No summary generator available, using fallback")
		return m.generateFallbackSummary(), nil
	}

	// 1. Build prompt
	prompt := m.buildSummaryPrompt()

	// 2. Call LLM via generator
	response, err := generator.GenerateSummary(ctx, prompt)
	if err != nil {
		// LLM call failed, use fallback
		logger.Warn(fmt.Sprintf("LLM summary generation failed: %v, using fallback", err))
		return m.generateFallbackSummary(), nil
	}

	// 3. Parse response
	summary, err := m.parseSummaryResponse(response)
	if err != nil {
		// Parsing failed, use fallback
		logger.Warn(fmt.Sprintf("Summary parsing failed: %v, using fallback", err))
		return m.generateFallbackSummary(), nil
	}

	// 4. Validate summary length (AC5: 200-400 characters for narrative impact)
	if len(summary.NarrativeImpact) > 500 {
		logger.Warn(fmt.Sprintf("Summary narrative_impact too long (%d chars), may need trimming", len(summary.NarrativeImpact)))
		// Truncate if too long
		summary.NarrativeImpact = summary.NarrativeImpact[:497] + "..."
	}

	return summary, nil
}
