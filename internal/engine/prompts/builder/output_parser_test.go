package builder

import (
	"testing"
)

// TestParseStructuredOutput_ValidJSON tests parsing of valid JSON output
func TestParseStructuredOutput_ValidJSON(t *testing.T) {
	jsonContent := `{
		"story": "這是完整的故事內容，包含大量的細節和描述。",
		"choices": ["選項一：向左走", "選項二：向右走", "選項三：原地等待"],
		"seeds": [
			{"type": "Item", "description": "一把生鏽的鑰匙"},
			{"type": "Location", "description": "神秘的地下室入口"}
		]
	}`

	output, err := ParseStructuredOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Story != "這是完整的故事內容，包含大量的細節和描述。" {
		t.Errorf("Story mismatch. Got: %s", output.Story)
	}

	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(output.Choices))
	}

	if len(output.Seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(output.Seeds))
	}

	if output.Seeds[0].Type != "Item" {
		t.Errorf("Expected first seed type 'Item', got '%s'", output.Seeds[0].Type)
	}
}

// TestParseStructuredOutput_EmptyChoices tests prologue format with no choices
func TestParseStructuredOutput_EmptyChoices(t *testing.T) {
	jsonContent := `{
		"story": "序章內容，沒有選擇。\n\n【按任意鍵繼續到第二章】",
		"choices": [],
		"seeds": []
	}`

	output, err := ParseStructuredOutput(jsonContent)

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

// TestParseStructuredOutput_NoSeeds tests JSON without seeds field
func TestParseStructuredOutput_NoSeeds(t *testing.T) {
	jsonContent := `{
		"story": "故事內容",
		"choices": ["選項1", "選項2"]
	}`

	output, err := ParseStructuredOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Story != "故事內容" {
		t.Error("Story should be parsed correctly")
	}

	// Seeds should be nil or empty array
	if output.Seeds == nil || len(output.Seeds) == 0 {
		// Both are acceptable
	} else {
		t.Errorf("Expected no seeds, got %d", len(output.Seeds))
	}
}

// TestExtractJSONFromMarkdown_CodeBlock tests JSON in markdown code blocks
func TestExtractJSONFromMarkdown_CodeBlock(t *testing.T) {
	markdownContent := "```json\n{\"story\": \"內容\", \"choices\": []}\n```"

	extracted := extractJSONFromMarkdown(markdownContent)

	if extracted != "{\"story\": \"內容\", \"choices\": []}" {
		t.Errorf("Failed to extract JSON from markdown. Got: %s", extracted)
	}
}

// TestExtractJSONFromMarkdown_PlainCodeBlock tests JSON in plain code blocks
func TestExtractJSONFromMarkdown_PlainCodeBlock(t *testing.T) {
	markdownContent := "```\n{\"story\": \"內容\", \"choices\": []}\n```"

	extracted := extractJSONFromMarkdown(markdownContent)

	if extracted != "{\"story\": \"內容\", \"choices\": []}" {
		t.Errorf("Failed to extract JSON from plain code block. Got: %s", extracted)
	}
}

// TestExtractJSONFromMarkdown_EmbeddedJSON tests extracting JSON embedded in text
func TestExtractJSONFromMarkdown_EmbeddedJSON(t *testing.T) {
	content := "Some preamble text\n{\"story\": \"內容\", \"choices\": []}\nSome trailing text"

	extracted := extractJSONFromMarkdown(content)

	if extracted != "{\"story\": \"內容\", \"choices\": []}" {
		t.Errorf("Failed to extract embedded JSON. Got: %s", extracted)
	}
}

// TestExtractJSONFromMarkdown_PlainJSON tests plain JSON without wrappers
func TestExtractJSONFromMarkdown_PlainJSON(t *testing.T) {
	content := "{\"story\": \"內容\", \"choices\": []}"

	extracted := extractJSONFromMarkdown(content)

	if extracted != content {
		t.Errorf("Plain JSON should be returned as-is. Got: %s", extracted)
	}
}

// TestParseLegacyFormat_WithChoices tests legacy text format with choices
func TestParseLegacyFormat_WithChoices(t *testing.T) {
	legacyContent := `這是故事內容。

很多描述文字在這裡。

選擇：
1. 第一個選項
2. 第二個選項
3. 第三個選項`

	output := parseLegacyFormat(legacyContent)

	if output.Story == "" {
		t.Error("Story should not be empty")
	}

	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(output.Choices))
	}

	if output.Choices[0] != "第一個選項" {
		t.Errorf("First choice mismatch. Got: %s", output.Choices[0])
	}
}

// TestParseLegacyFormat_WithSeeds tests legacy format with HTML seed markers
func TestParseLegacyFormat_WithSeeds(t *testing.T) {
	legacyContent := `故事內容 <!-- SEED:Item:鑰匙 --> 更多內容 <!-- SEED:Event:神秘事件 -->`

	output := parseLegacyFormat(legacyContent)

	if len(output.Seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(output.Seeds))
	}

	if output.Seeds[0].Type != "Item" {
		t.Errorf("First seed type mismatch. Got: %s", output.Seeds[0].Type)
	}

	if output.Seeds[1].Description != "神秘事件" {
		t.Errorf("Second seed description mismatch. Got: %s", output.Seeds[1].Description)
	}
}

// TestParseStructuredOutput_InvalidJSON tests fallback on invalid JSON
func TestParseStructuredOutput_InvalidJSON(t *testing.T) {
	invalidJSON := `This is some story content that happens to have invalid JSON markers {story: "missing quotes"`

	output, err := ParseStructuredOutput(invalidJSON)

	// Should not error - should fall back to legacy format
	if err != nil {
		t.Fatalf("Expected graceful fallback, got error: %v", err)
	}

	// In legacy mode, the content becomes the story
	if output.Story == "" {
		t.Error("Story should contain content in fallback mode")
	}
}

// TestParseStructuredOutput_MalformedMarkdown tests malformed markdown blocks
func TestParseStructuredOutput_MalformedMarkdown(t *testing.T) {
	malformed := "```json\n{\"story\": \"content\"\n"

	output, err := ParseStructuredOutput(malformed)

	// Should fall back to legacy parsing
	if err != nil {
		t.Fatalf("Expected graceful fallback, got error: %v", err)
	}

	if output.Story == "" {
		t.Error("Should have fallback content")
	}
}

// TestParseStructuredOutput_ChineseChoiceFormat tests Chinese choice format
func TestParseStructuredOutput_ChineseChoiceFormat(t *testing.T) {
	legacyContent := `故事內容

**選項：**
1、選項一
2、選項二`

	output, err := ParseStructuredOutput(legacyContent)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(output.Choices) != 2 {
		t.Errorf("Expected 2 choices from Chinese format, got %d", len(output.Choices))
	}
}

// TestParseStructuredOutput_EnglishChoiceFormat tests English choice format
func TestParseStructuredOutput_EnglishChoiceFormat(t *testing.T) {
	legacyContent := `Story content here

Choices:
1. First option
2. Second option`

	output, err := ParseStructuredOutput(legacyContent)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(output.Choices) != 2 {
		t.Errorf("Expected 2 choices from English format, got %d", len(output.Choices))
	}
}

// TestParseStructuredOutput_NestedJSON tests handling of nested JSON structures
func TestParseStructuredOutput_NestedJSON(t *testing.T) {
	jsonContent := `{
		"story": "故事中有引號\"特殊字符\"和換行\n內容",
		"choices": ["選項 with \"quotes\""],
		"seeds": [{"type": "Item", "description": "物品\"描述\""}]
	}`

	output, err := ParseStructuredOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error with nested JSON, got %v", err)
	}

	if output.Story == "" {
		t.Error("Story should handle nested quotes")
	}
}

// TestParseStructuredOutput_EmptyInput tests empty input handling
func TestParseStructuredOutput_EmptyInput(t *testing.T) {
	output, err := ParseStructuredOutput("")

	if err != nil {
		t.Fatalf("Expected no error with empty input, got %v", err)
	}

	if output.Story != "" {
		t.Error("Empty input should produce empty story")
	}
}

// TestParseChoicesFromText tests the legacy choice parsing function
func TestParseChoicesFromText(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "Chinese format with 、",
			content: `選擇：
1、選項一
2、選項二`,
			expected: 2,
		},
		{
			name: "English format with .",
			content: `Choices:
1. Option one
2. Option two
3. Option three`,
			expected: 3,
		},
		{
			name: "Mixed with )",
			content: `**選項：**
1) 選項A
2) 選項B`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			choices := parseChoicesFromText(tt.content)
			if len(choices) != tt.expected {
				t.Errorf("Expected %d choices, got %d", tt.expected, len(choices))
			}
		})
	}
}

// ========================================
// DeathOutput Parsing Tests
// ========================================

// TestParseDeathOutput_ValidJSON tests parsing valid death JSON output
func TestParseDeathOutput_ValidJSON(t *testing.T) {
	jsonContent := `{
		"narrative": "你的身體終於承受不住了。每一次呼吸都像是在吸入碎玻璃，視野逐漸模糊。黑暗從四面八方湧來，將你徹底吞沒。",
		"cause": "hp_zero",
		"hints": [
			"提示1：保持 HP 在安全範圍內",
			"提示2：避免不必要的戰鬥",
			"提示3：尋找醫療補給"
		]
	}`

	output, err := ParseDeathOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Narrative == "" {
		t.Error("Narrative should not be empty")
	}

	if output.Cause != "hp_zero" {
		t.Errorf("Expected cause 'hp_zero', got '%s'", output.Cause)
	}

	if len(output.Hints) != 3 {
		t.Errorf("Expected 3 hints, got %d", len(output.Hints))
	}
}

// TestParseDeathOutput_InsanityJSON tests insanity death format
func TestParseDeathOutput_InsanityJSON(t *testing.T) {
	jsonContent := `{
		"narrative": "他們一直都在看著你。牆壁裡、天花板上、你自己的倒影裡。你終於成為了我們的一部分。",
		"cause": "insanity",
		"hints": [
			"提示1：監控理智值變化",
			"提示2：避免直視恐怖事物"
		]
	}`

	output, err := ParseDeathOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Cause != "insanity" {
		t.Errorf("Expected cause 'insanity', got '%s'", output.Cause)
	}

	if len(output.Hints) != 2 {
		t.Errorf("Expected 2 hints, got %d", len(output.Hints))
	}
}

// TestParseDeathOutput_RuleViolationJSON tests rule violation format
func TestParseDeathOutput_RuleViolationJSON(t *testing.T) {
	jsonContent := `{
		"narrative": "有些規則是不能被打破的。當你踏出那一步的瞬間，你就知道自己錯了。但現在，一切都太遲了。",
		"cause": "rule_violation",
		"hints": [
			"規則：絕對不要在午夜後開門",
			"線索：門外的敲門聲是陷阱",
			"正確做法：無視敲門聲，等待天亮"
		]
	}`

	output, err := ParseDeathOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Cause != "rule_violation" {
		t.Errorf("Expected cause 'rule_violation', got '%s'", output.Cause)
	}

	if len(output.Hints) != 3 {
		t.Errorf("Expected 3 hints, got %d", len(output.Hints))
	}
}

// TestParseDeathOutput_MarkdownWrapped tests JSON in markdown code blocks
func TestParseDeathOutput_MarkdownWrapped(t *testing.T) {
	markdownContent := "```json\n{\"narrative\": \"死亡敘述\", \"cause\": \"hp_zero\", \"hints\": [\"提示1\"]}\n```"

	output, err := ParseDeathOutput(markdownContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Narrative != "死亡敘述" {
		t.Errorf("Narrative mismatch. Got: %s", output.Narrative)
	}
}

// TestParseDeathOutput_LegacyFallback tests fallback to plain text
func TestParseDeathOutput_LegacyFallback(t *testing.T) {
	legacyContent := `你的身體終於承受不住了。

黑暗從四面八方湧來，將你徹底吞沒。`

	output, err := ParseDeathOutput(legacyContent)

	if err != nil {
		t.Fatalf("Expected no error with legacy format, got %v", err)
	}

	if output.Narrative == "" {
		t.Error("Narrative should contain legacy content")
	}

	if output.Cause != "unknown" {
		t.Errorf("Legacy fallback should have cause 'unknown', got '%s'", output.Cause)
	}

	if len(output.Hints) != 0 {
		t.Errorf("Legacy fallback should have no hints, got %d", len(output.Hints))
	}
}

// TestParseDeathOutput_EmptyInput tests empty input handling
func TestParseDeathOutput_EmptyInput(t *testing.T) {
	output, err := ParseDeathOutput("")

	if err != nil {
		t.Fatalf("Expected no error with empty input, got %v", err)
	}

	if output.Narrative != "" {
		t.Error("Empty input should produce empty narrative")
	}
}

// TestParseDeathOutput_InvalidJSON tests invalid JSON fallback
func TestParseDeathOutput_InvalidJSON(t *testing.T) {
	invalidJSON := `{narrative: "missing quotes", cause: hp_zero`

	output, err := ParseDeathOutput(invalidJSON)

	// Should not error - should fall back to legacy
	if err != nil {
		t.Fatalf("Expected graceful fallback, got error: %v", err)
	}

	// In legacy mode, the invalid JSON becomes the narrative
	if output.Narrative == "" {
		t.Error("Narrative should contain the original content in fallback mode")
	}
}

// ========================================
// DreamOutput Parsing Tests
// ========================================

// TestParseDreamOutput_ValidJSON tests parsing valid dream JSON output
func TestParseDreamOutput_ValidJSON(t *testing.T) {
	jsonContent := `{
		"dream": "你在一片黑暗中行走，鏡子映照出扭曲的影子。門在你背後關上了，無法回頭。迷宮在召喚你前進。",
		"symbols": ["鏡子", "門", "迷宮"],
		"rules_hinted": [1, 2],
		"atmosphere": "uneasy"
	}`

	output, err := ParseDreamOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Dream == "" {
		t.Error("Dream should not be empty")
	}

	if len(output.Symbols) != 3 {
		t.Errorf("Expected 3 symbols, got %d", len(output.Symbols))
	}

	if output.Atmosphere != "uneasy" {
		t.Errorf("Expected atmosphere 'uneasy', got '%s'", output.Atmosphere)
	}

	if len(output.RulesHinted) != 2 {
		t.Errorf("Expected 2 rules hinted, got %d", len(output.RulesHinted))
	}
}

// TestParseDreamOutput_NightmareJSON tests nightmare dream format
func TestParseDreamOutput_NightmareJSON(t *testing.T) {
	jsonContent := `{
		"dream": "那個恐怖的時刻在夢中扭曲重現。牆壁在融化，空氣凝固成液體。你無法呼吸，無法逃離。",
		"symbols": ["融化的牆壁", "液態空氣"],
		"rules_hinted": [],
		"atmosphere": "nightmare"
	}`

	output, err := ParseDreamOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Atmosphere != "nightmare" {
		t.Errorf("Expected atmosphere 'nightmare', got '%s'", output.Atmosphere)
	}

	if len(output.Symbols) != 2 {
		t.Errorf("Expected 2 symbols, got %d", len(output.Symbols))
	}
}

// TestParseDreamOutput_GriefJSON tests grief dream format
func TestParseDreamOutput_GriefJSON(t *testing.T) {
	jsonContent := `{
		"dream": "你在夢中看到了她。她笑著，手中拿著那把熟悉的手電筒。「別忘記我告訴你的話。」她說完就消失了。",
		"symbols": ["手電筒", "告別的對話"],
		"rules_hinted": [],
		"atmosphere": "grief"
	}`

	output, err := ParseDreamOutput(jsonContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Atmosphere != "grief" {
		t.Errorf("Expected atmosphere 'grief', got '%s'", output.Atmosphere)
	}
}

// TestParseDreamOutput_MarkdownWrapped tests JSON in markdown code blocks
func TestParseDreamOutput_MarkdownWrapped(t *testing.T) {
	markdownContent := "```json\n{\"dream\": \"夢境內容\", \"symbols\": [\"符號\"], \"rules_hinted\": [], \"atmosphere\": \"calm\"}\n```"

	output, err := ParseDreamOutput(markdownContent)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.Dream != "夢境內容" {
		t.Errorf("Dream mismatch. Got: %s", output.Dream)
	}

	if output.Atmosphere != "calm" {
		t.Errorf("Expected atmosphere 'calm', got '%s'", output.Atmosphere)
	}
}

// TestParseDreamOutput_LegacyFallback tests fallback to plain text
func TestParseDreamOutput_LegacyFallback(t *testing.T) {
	legacyContent := `你在一片黑暗中行走。

鏡子映照出扭曲的影子。門在你背後關上了。`

	output, err := ParseDreamOutput(legacyContent)

	if err != nil {
		t.Fatalf("Expected no error with legacy format, got %v", err)
	}

	if output.Dream == "" {
		t.Error("Dream should contain legacy content")
	}

	if output.Atmosphere != "unknown" {
		t.Errorf("Legacy fallback should have atmosphere 'unknown', got '%s'", output.Atmosphere)
	}

	if len(output.Symbols) != 0 {
		t.Errorf("Legacy fallback should have no symbols, got %d", len(output.Symbols))
	}
}

// TestParseDreamOutput_EmptyInput tests empty input handling
func TestParseDreamOutput_EmptyInput(t *testing.T) {
	output, err := ParseDreamOutput("")

	if err != nil {
		t.Fatalf("Expected no error with empty input, got %v", err)
	}

	if output.Dream != "" {
		t.Error("Empty input should produce empty dream")
	}
}

// TestParseDreamOutput_InvalidJSON tests invalid JSON fallback
func TestParseDreamOutput_InvalidJSON(t *testing.T) {
	invalidJSON := `{dream: "missing quotes", atmosphere: calm`

	output, err := ParseDreamOutput(invalidJSON)

	// Should not error - should fall back to legacy
	if err != nil {
		t.Fatalf("Expected graceful fallback, got error: %v", err)
	}

	// In legacy mode, the invalid JSON becomes the dream
	if output.Dream == "" {
		t.Error("Dream should contain the original content in fallback mode")
	}
}
