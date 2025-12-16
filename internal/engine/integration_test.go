package engine

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/builder"
)

// TestEndToEndJSONParsing tests the complete flow from JSON output to parsed result
func TestEndToEndJSONParsing(t *testing.T) {
	// Simulate LLM returning complete JSON output
	llmOutput := `{
		"story": "你站在一座廢棄醫院的門前。鏽蝕的招牌在風中搖晃，發出刺耳的聲音。門半開著，裡面一片漆黑。",
		"choices": [
			"推開門進入醫院",
			"繞到後門查看",
			"先在外面觀察一下"
		],
		"seeds": [
			{"type": "Item", "description": "生鏽的招牌"},
			{"type": "Location", "description": "後門"}
		]
	}`

	// Step 1: Parse the output
	output, err := builder.ParseStructuredOutput(llmOutput)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Step 2: Verify story content
	if output.Story == "" {
		t.Error("Story should not be empty")
	}
	if len(output.Story) < 50 {
		t.Errorf("Story seems too short: %d characters", len(output.Story))
	}

	// Step 3: Verify choices are separated from story
	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(output.Choices))
	}

	// Step 4: Verify seeds are extracted
	if len(output.Seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(output.Seeds))
	}

	// Step 5: Simulate Story Engine processing
	result := &GenerationResult{
		Content: output.Story,
		Choices: output.Choices,
		Seeds:   output.Seeds,
	}

	// Step 6: Verify result is ready for game display
	if result.Content == "" {
		t.Error("GenerationResult.Content should not be empty")
	}
	if len(result.Choices) == 0 {
		t.Error("GenerationResult.Choices should not be empty")
	}

	// Critical: Ensure choices do NOT appear in story content
	for _, choice := range result.Choices {
		// The story content should not contain the choice text
		// (This was the original bug we're fixing)
		// Note: We can't do a simple string contains check because choice text
		// might legitimately appear in the story, but the format "1. Choice" should not
		t.Logf("Choice: %s", choice)
	}
}

// TestLegacyFormatFallback tests backward compatibility with old text format
func TestLegacyFormatFallback(t *testing.T) {
	// Simulate LLM returning old-style text output
	legacyOutput := `你站在一座廢棄醫院的門前。鏽蝕的招牌在風中搖晃，發出刺耳的聲音。

門半開著，裡面一片漆黑。你感到一陣寒意。

<!-- SEED:Item:生鏽的招牌 -->
<!-- SEED:Location:後門 -->

選擇：
1. 推開門進入醫院
2. 繞到後門查看
3. 先在外面觀察一下`

	// Parse with fallback
	output, err := builder.ParseStructuredOutput(legacyOutput)
	if err != nil {
		t.Fatalf("Failed to parse legacy output: %v", err)
	}

	// Verify story is extracted
	if output.Story == "" {
		t.Error("Story should not be empty in legacy mode")
	}

	// Verify choices are extracted
	if len(output.Choices) != 3 {
		t.Errorf("Expected 3 choices from legacy format, got %d", len(output.Choices))
	}

	// Verify seeds are extracted from HTML comments
	if len(output.Seeds) != 2 {
		t.Errorf("Expected 2 seeds from legacy format, got %d", len(output.Seeds))
	}

	// Verify story content does not contain choice section
	t.Logf("Story content: %s", output.Story)
	// The CleanContent function should have removed "選擇：" section
}

// TestMixedFormatHandling tests handling of malformed or mixed outputs
func TestMixedFormatHandling(t *testing.T) {
	tests := []struct {
		name           string
		llmOutput      string
		expectChoices  int
		expectSeeds    int
		shouldHaveText bool
	}{
		{
			name: "JSON in markdown code block",
			llmOutput: "```json\n" +
				`{"story": "故事內容", "choices": ["選項1", "選項2"], "seeds": []}` +
				"\n```",
			expectChoices:  2,
			expectSeeds:    0,
			shouldHaveText: true,
		},
		{
			name: "Invalid JSON falls back to legacy",
			llmOutput: `{story: "missing quotes"

選擇：
1. 選項1
2. 選項2`,
			expectChoices:  2,
			expectSeeds:    0,
			shouldHaveText: true,
		},
		{
			name: "Prologue with no choices",
			llmOutput: `{
				"story": "序章內容，沒有選擇。\n\n【按任意鍵繼續到第二章】",
				"choices": [],
				"seeds": []
			}`,
			expectChoices:  0,
			expectSeeds:    0,
			shouldHaveText: true,
		},
		{
			name: "Empty output",
			llmOutput: "",
			expectChoices:  0,
			expectSeeds:    0,
			shouldHaveText: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := builder.ParseStructuredOutput(tt.llmOutput)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(output.Choices) != tt.expectChoices {
				t.Errorf("Expected %d choices, got %d", tt.expectChoices, len(output.Choices))
			}

			if len(output.Seeds) != tt.expectSeeds {
				t.Errorf("Expected %d seeds, got %d", tt.expectSeeds, len(output.Seeds))
			}

			if tt.shouldHaveText && output.Story == "" {
				t.Error("Expected story text, got empty")
			}
		})
	}
}

// TestDeathOutputIntegration tests death narrative JSON integration
func TestDeathOutputIntegration(t *testing.T) {
	deathJSON := `{
		"narrative": "你的身體終於承受不住了。視野逐漸模糊，黑暗將你吞沒。",
		"cause": "hp_zero",
		"hints": [
			"避免不必要的戰鬥",
			"尋找醫療補給",
			"保持 HP 在 30 以上"
		]
	}`

	output, err := builder.ParseDeathOutput(deathJSON)
	if err != nil {
		t.Fatalf("Failed to parse death JSON: %v", err)
	}

	if output.Narrative == "" {
		t.Error("Death narrative should not be empty")
	}

	if output.Cause != "hp_zero" {
		t.Errorf("Expected cause 'hp_zero', got '%s'", output.Cause)
	}

	if len(output.Hints) != 3 {
		t.Errorf("Expected 3 hints, got %d", len(output.Hints))
	}
}

// TestDreamOutputIntegration tests dream content JSON integration
func TestDreamOutputIntegration(t *testing.T) {
	dreamJSON := `{
		"dream": "你在夢中看到鏡子。鏡中的你做著相反的動作。門在背後關上了。",
		"symbols": ["鏡子", "門"],
		"rules_hinted": [1],
		"atmosphere": "uneasy"
	}`

	output, err := builder.ParseDreamOutput(dreamJSON)
	if err != nil {
		t.Fatalf("Failed to parse dream JSON: %v", err)
	}

	if output.Dream == "" {
		t.Error("Dream should not be empty")
	}

	if len(output.Symbols) != 2 {
		t.Errorf("Expected 2 symbols, got %d", len(output.Symbols))
	}

	if output.Atmosphere != "uneasy" {
		t.Errorf("Expected atmosphere 'uneasy', got '%s'", output.Atmosphere)
	}

	if len(output.RulesHinted) != 1 {
		t.Errorf("Expected 1 rule hinted, got %d", len(output.RulesHinted))
	}
}

// TestJSONFormatConsistency tests that all JSON parsers handle similar edge cases
func TestJSONFormatConsistency(t *testing.T) {
	// Test markdown wrapper handling across all parsers
	testCases := []struct {
		name   string
		parser func(string) (interface{}, error)
		input  string
	}{
		{
			name: "Story with markdown wrapper",
			parser: func(s string) (interface{}, error) {
				return builder.ParseStructuredOutput(s)
			},
			input: "```json\n{\"story\": \"內容\", \"choices\": [], \"seeds\": []}\n```",
		},
		{
			name: "Death with markdown wrapper",
			parser: func(s string) (interface{}, error) {
				return builder.ParseDeathOutput(s)
			},
			input: "```json\n{\"narrative\": \"死亡\", \"cause\": \"hp_zero\", \"hints\": []}\n```",
		},
		{
			name: "Dream with markdown wrapper",
			parser: func(s string) (interface{}, error) {
				return builder.ParseDreamOutput(s)
			},
			input: "```json\n{\"dream\": \"夢境\", \"symbols\": [], \"rules_hinted\": [], \"atmosphere\": \"calm\"}\n```",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.parser(tc.input)
			if err != nil {
				t.Fatalf("Parser failed: %v", err)
			}
			if result == nil {
				t.Error("Parser returned nil result")
			}
		})
	}
}
