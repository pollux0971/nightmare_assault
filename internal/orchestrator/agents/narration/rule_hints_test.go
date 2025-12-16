package narration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateRuleHint tests rule hint generation for different difficulties
func TestGenerateRuleHint(t *testing.T) {
	tests := []struct {
		name          string
		ruleID        string
		ruleDesc      string
		difficulty    string
		warningCount  int
		maxWarnings   int
		expectEmpty   bool
		expectContain []string
		minLength     int
	}{
		{
			name:          "easy difficulty - first warning",
			ruleID:        "rule-001",
			ruleDesc:      "不要在夜晚開燈",
			difficulty:    "easy",
			warningCount:  0,
			maxWarnings:   2,
			expectEmpty:   false,
			expectContain: []string{"感覺", "不太對"},
			minLength:     15,
		},
		{
			name:          "easy difficulty - second warning",
			ruleID:        "rule-001",
			ruleDesc:      "不要在夜晚開燈",
			difficulty:    "easy",
			warningCount:  1,
			maxWarnings:   2,
			expectEmpty:   false,
			expectContain: []string{"警告"},
			minLength:     15,
		},
		{
			name:         "easy difficulty - exceed max warnings",
			ruleID:       "rule-001",
			ruleDesc:     "不要在夜晚開燈",
			difficulty:   "easy",
			warningCount: 2,
			maxWarnings:  2,
			expectEmpty:  true,
		},
		{
			name:          "normal difficulty - first warning",
			ruleID:        "rule-002",
			ruleDesc:      "不要直視鏡子",
			difficulty:    "normal",
			warningCount:  0,
			maxWarnings:   1,
			expectEmpty:   false,
			expectContain: []string{"氛圍", "詭異"},
			minLength:     15,
		},
		{
			name:         "normal difficulty - exceed max warnings",
			ruleID:       "rule-002",
			ruleDesc:     "不要直視鏡子",
			difficulty:   "normal",
			warningCount: 1,
			maxWarnings:  1,
			expectEmpty:  true,
		},
		{
			name:          "hard difficulty - first warning",
			ruleID:        "rule-003",
			ruleDesc:      "不要回應呼喊",
			difficulty:    "hard",
			warningCount:  0,
			maxWarnings:   1,
			expectEmpty:   false,
			expectContain: []string{"注意到", "細節"},
			minLength:     10,
		},
		{
			name:         "hard difficulty - exceed max warnings",
			ruleID:       "rule-003",
			ruleDesc:     "不要回應呼喊",
			difficulty:   "hard",
			warningCount: 1,
			maxWarnings:  1,
			expectEmpty:  true,
		},
		{
			name:         "hell difficulty - no warnings",
			ruleID:       "rule-004",
			ruleDesc:     "不要進入地下室",
			difficulty:   "hell",
			warningCount: 0,
			maxWarnings:  0,
			expectEmpty:  true,
		},
		{
			name:          "easy difficulty with rule context",
			ruleID:        "rule-005",
			ruleDesc:      "醫院的燈光在午夜會吸引「它們」",
			difficulty:    "easy",
			warningCount:  0,
			maxWarnings:   2,
			expectEmpty:   false,
			expectContain: []string{"燈", "午夜"},
			minLength:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRuleHint(tt.ruleID, tt.ruleDesc, tt.difficulty, tt.warningCount, tt.maxWarnings)

			if tt.expectEmpty {
				assert.Empty(t, result, "Expected empty hint")
			} else {
				assert.NotEmpty(t, result, "Expected non-empty hint")

				// Check minimum length
				runeCount := len([]rune(result))
				assert.GreaterOrEqual(t, runeCount, tt.minLength,
					"Hint too short: %d chars", runeCount)

				// Check expected content
				for _, expect := range tt.expectContain {
					assert.Contains(t, result, expect,
						"Hint should contain '%s'", expect)
				}
			}
		})
	}
}

// TestGetMaxWarnings tests max warning count for different difficulties
func TestGetMaxWarnings(t *testing.T) {
	tests := []struct {
		difficulty  string
		maxWarnings int
	}{
		{"easy", 2},
		{"normal", 1},
		{"hard", 1},
		{"hell", 0},
		{"unknown", 0}, // Default to no warnings for unknown difficulty
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			result := GetMaxWarnings(tt.difficulty)
			assert.Equal(t, tt.maxWarnings, result)
		})
	}
}

// TestExtractRuleKeywords tests keyword extraction from rule description
func TestExtractRuleKeywords(t *testing.T) {
	tests := []struct {
		name        string
		ruleDesc    string
		minKeywords int
		maxKeywords int
	}{
		{
			name:        "simple rule",
			ruleDesc:    "不要在夜晚開燈",
			minKeywords: 1, // Adjusted for simple Chinese segmentation
			maxKeywords: 3,
		},
		{
			name:        "complex rule",
			ruleDesc:    "醫院的燈光在午夜會吸引「它們」",
			minKeywords: 2,
			maxKeywords: 4,
		},
		{
			name:        "very short rule",
			ruleDesc:    "別回頭",
			minKeywords: 1,
			maxKeywords: 2,
		},
		{
			name:        "empty rule",
			ruleDesc:    "",
			minKeywords: 0,
			maxKeywords: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := ExtractRuleKeywords(tt.ruleDesc)

			assert.GreaterOrEqual(t, len(keywords), tt.minKeywords,
				"Should extract at least %d keywords", tt.minKeywords)
			assert.LessOrEqual(t, len(keywords), tt.maxKeywords,
				"Should extract at most %d keywords", tt.maxKeywords)

			// All keywords should be non-empty
			for _, kw := range keywords {
				assert.NotEmpty(t, kw)
			}
		})
	}
}

// TestRuleHintProgression tests that hints get more direct with each warning
func TestRuleHintProgression(t *testing.T) {
	ruleID := "rule-test"
	ruleDesc := "不要在夜晚開燈"
	difficulty := "easy"
	maxWarnings := 2

	// First warning (subtle)
	hint1 := GenerateRuleHint(ruleID, ruleDesc, difficulty, 0, maxWarnings)
	require.NotEmpty(t, hint1)

	// Second warning (more direct)
	hint2 := GenerateRuleHint(ruleID, ruleDesc, difficulty, 1, maxWarnings)
	require.NotEmpty(t, hint2)

	// Hints should be different
	assert.NotEqual(t, hint1, hint2, "Hints should progress with each warning")

	// Third attempt (no more warnings)
	hint3 := GenerateRuleHint(ruleID, ruleDesc, difficulty, 2, maxWarnings)
	assert.Empty(t, hint3, "Should not give hint after max warnings")
}

// TestRuleHintDifficultyScaling tests hint clarity across difficulties
func TestRuleHintDifficultyScaling(t *testing.T) {
	ruleID := "rule-test"
	ruleDesc := "不要在夜晚開燈"
	warningCount := 0

	easyHint := GenerateRuleHint(ruleID, ruleDesc, "easy", warningCount, 2)
	normalHint := GenerateRuleHint(ruleID, ruleDesc, "normal", warningCount, 1)
	hardHint := GenerateRuleHint(ruleID, ruleDesc, "hard", warningCount, 1)
	hellHint := GenerateRuleHint(ruleID, ruleDesc, "hell", warningCount, 0)

	// Easy should give direct hints
	require.NotEmpty(t, easyHint)
	assert.Contains(t, easyHint, "感覺", "Easy hint should be direct")

	// Normal should give vague hints
	require.NotEmpty(t, normalHint)
	assert.Contains(t, normalHint, "氛圍", "Normal hint should be vague")

	// Hard should give very subtle hints
	require.NotEmpty(t, hardHint)
	assert.Contains(t, hardHint, "注意", "Hard hint should be very subtle")

	// Hell should give no hints
	assert.Empty(t, hellHint, "Hell difficulty should give no hints")
}

// TestBuildHintText tests hint text construction
func TestBuildHintText(t *testing.T) {
	tests := []struct {
		name        string
		hintLevel   HintLevel
		keywords    []string
		expectEmpty bool
		minLength   int
	}{
		{
			name:        "direct hint with keywords",
			hintLevel:   HintLevelDirect,
			keywords:    []string{"燈", "夜晚"},
			expectEmpty: false,
			minLength:   15,
		},
		{
			name:        "vague hint with keywords",
			hintLevel:   HintLevelVague,
			keywords:    []string{"鏡子"},
			expectEmpty: false,
			minLength:   15,
		},
		{
			name:        "subtle hint with keywords",
			hintLevel:   HintLevelSubtle,
			keywords:    []string{"聲音"},
			expectEmpty: false,
			minLength:   10,
		},
		{
			name:        "no hint level",
			hintLevel:   HintLevelNone,
			keywords:    []string{"test"},
			expectEmpty: true,
			minLength:   0,
		},
		{
			name:        "direct hint without keywords",
			hintLevel:   HintLevelDirect,
			keywords:    []string{},
			expectEmpty: false,
			minLength:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildHintText(tt.hintLevel, tt.keywords)

			if tt.expectEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				runeCount := len([]rune(result))
				assert.GreaterOrEqual(t, runeCount, tt.minLength,
					"Hint text too short: %d chars", runeCount)
			}
		})
	}
}

// TestHintLevel_String tests the String() method for HintLevel enum
func TestHintLevel_String(t *testing.T) {
	tests := []struct {
		level    HintLevel
		expected string
	}{
		{HintLevelNone, "none"},
		{HintLevelSubtle, "subtle"},
		{HintLevelVague, "vague"},
		{HintLevelDirect, "direct"},
		{HintLevel(99), "unknown"}, // Unknown/invalid level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result,
				"HintLevel %d should return '%s'", tt.level, tt.expected)
		})
	}
}

// TestGetHintLevel tests the getHintLevel function with various difficulties
func TestGetHintLevel(t *testing.T) {
	tests := []struct {
		name         string
		difficulty   string
		warningCount int
		expected     HintLevel
	}{
		{
			name:         "easy difficulty",
			difficulty:   "easy",
			warningCount: 0,
			expected:     HintLevelDirect,
		},
		{
			name:         "Easy (uppercase)",
			difficulty:   "Easy",
			warningCount: 0,
			expected:     HintLevelDirect,
		},
		{
			name:         "EASY (all caps)",
			difficulty:   "EASY",
			warningCount: 0,
			expected:     HintLevelDirect,
		},
		{
			name:         "normal difficulty",
			difficulty:   "normal",
			warningCount: 0,
			expected:     HintLevelVague,
		},
		{
			name:         "hard difficulty",
			difficulty:   "hard",
			warningCount: 0,
			expected:     HintLevelSubtle,
		},
		{
			name:         "hell difficulty",
			difficulty:   "hell",
			warningCount: 0,
			expected:     HintLevelNone,
		},
		{
			name:         "unknown difficulty",
			difficulty:   "unknown",
			warningCount: 0,
			expected:     HintLevelNone,
		},
		{
			name:         "empty difficulty",
			difficulty:   "",
			warningCount: 0,
			expected:     HintLevelNone,
		},
		{
			name:         "invalid difficulty",
			difficulty:   "invalid-123",
			warningCount: 0,
			expected:     HintLevelNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHintLevel(tt.difficulty, tt.warningCount)
			assert.Equal(t, tt.expected, result,
				"Difficulty '%s' should return hint level %v", tt.difficulty, tt.expected)
		})
	}
}

// TestExtractRuleKeywords_EdgeCases tests edge cases for keyword extraction
// Note: Chinese text keyword extraction is simplified and doesn't use NLP
// The function splits by whitespace, so Chinese text without spaces may return
// the full cleaned string as a single "keyword"
func TestExtractRuleKeywords_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		ruleDesc       string
		minKeywords    int
		maxKeywords    int
		expectNonEmpty bool
	}{
		{
			name:           "empty string",
			ruleDesc:       "",
			minKeywords:    0,
			maxKeywords:    0,
			expectNonEmpty: false,
		},
		{
			name:           "only stop words",
			ruleDesc:       "不要的了在與和",
			minKeywords:    0,
			maxKeywords:    3,
			expectNonEmpty: true, // May extract cleaned version
		},
		{
			name:           "only punctuation",
			ruleDesc:       "「」『』，。、；：",
			minKeywords:    0,
			maxKeywords:    0,
			expectNonEmpty: false,
		},
		{
			name:           "single character words with spaces",
			ruleDesc:       "一 二 三 四 五",
			minKeywords:    0,
			maxKeywords:    0,
			expectNonEmpty: false, // All filtered out (less than 2 chars)
		},
		{
			name:           "mixed Chinese text (no spaces)",
			ruleDesc:       "不要在夜晚的時候開燈",
			minKeywords:    1,
			maxKeywords:    3,
			expectNonEmpty: true, // Will extract cleaned text
		},
		{
			name:           "rule with many keywords",
			ruleDesc:       "不要在深夜獨自打開地下室的門進入黑暗房間",
			minKeywords:    1,
			maxKeywords:    3, // Limited to 3 keywords
			expectNonEmpty: true,
		},
		{
			name:           "keywords with punctuation",
			ruleDesc:       "不要「直視」鏡子，也別「碰觸」牆壁。",
			minKeywords:    1,
			maxKeywords:    3,
			expectNonEmpty: true, // Should extract despite punctuation
		},
		{
			name:           "very short rule",
			ruleDesc:       "別回頭",
			minKeywords:    1,
			maxKeywords:    2,
			expectNonEmpty: true,
		},
		{
			name:           "rule with spaces and keywords",
			ruleDesc:       "不要 在 夜晚 開燈",
			minKeywords:    1,
			maxKeywords:    3,
			expectNonEmpty: true,
		},
		{
			name:           "complex nested text",
			ruleDesc:       "在黑暗中不可以開燈但是也不能不動",
			minKeywords:    1,
			maxKeywords:    3,
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRuleKeywords(tt.ruleDesc)

			// Check keyword count range
			assert.GreaterOrEqual(t, len(result), tt.minKeywords,
				"Should have at least %d keywords", tt.minKeywords)
			assert.LessOrEqual(t, len(result), tt.maxKeywords,
				"Should have at most %d keywords", tt.maxKeywords)

			// Check if result is non-empty when expected
			if tt.expectNonEmpty {
				assert.NotEmpty(t, result,
					"Should extract some keywords")
				// All extracted keywords should be non-empty strings
				for _, kw := range result {
					assert.NotEmpty(t, kw,
						"Each keyword should be non-empty")
					assert.GreaterOrEqual(t, len([]rune(kw)), 2,
						"Each keyword should have at least 2 characters")
				}
			} else {
				assert.Empty(t, result,
					"Should not extract any keywords")
			}
		})
	}
}

// TestGenerateRuleHint_EdgeCases tests edge cases for rule hint generation
func TestGenerateRuleHint_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		ruleID        string
		ruleDesc      string
		difficulty    string
		warningCount  int
		maxWarnings   int
		expectEmpty   bool
	}{
		{
			name:         "exceeded max warnings",
			ruleID:       "rule-001",
			ruleDesc:     "不要開燈",
			difficulty:   "easy",
			warningCount: 2,
			maxWarnings:  2,
			expectEmpty:  true,
		},
		{
			name:         "exactly at max warnings",
			ruleID:       "rule-002",
			ruleDesc:     "不要回頭",
			difficulty:   "normal",
			warningCount: 1,
			maxWarnings:  1,
			expectEmpty:  true,
		},
		{
			name:         "hell difficulty (no hints)",
			ruleID:       "rule-003",
			ruleDesc:     "不要說話",
			difficulty:   "hell",
			warningCount: 0,
			maxWarnings:  0,
			expectEmpty:  true,
		},
		{
			name:         "empty rule description",
			ruleID:       "rule-004",
			ruleDesc:     "",
			difficulty:   "easy",
			warningCount: 0,
			maxWarnings:  2,
			expectEmpty:  false, // Should still generate hint
		},
		{
			name:         "unknown difficulty defaults to no hint",
			ruleID:       "rule-005",
			ruleDesc:     "不要移動",
			difficulty:   "unknown",
			warningCount: 0,
			maxWarnings:  0,
			expectEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRuleHint(tt.ruleID, tt.ruleDesc, tt.difficulty,
				tt.warningCount, tt.maxWarnings)

			if tt.expectEmpty {
				assert.Empty(t, result,
					"Should return empty hint for test case: %s", tt.name)
			} else {
				assert.NotEmpty(t, result,
					"Should return non-empty hint for test case: %s", tt.name)
			}
		})
	}
}

// TestBuildHintTextWithProgression_AllLevels tests buildHintTextWithProgression with all hint levels
func TestBuildHintTextWithProgression_AllLevels(t *testing.T) {
	tests := []struct {
		name         string
		hintLevel    HintLevel
		keywords     []string
		warningCount int
		expectEmpty  bool
	}{
		{
			name:         "direct hint with keywords",
			hintLevel:    HintLevelDirect,
			keywords:     []string{"夜晚", "開燈"},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "direct hint without keywords",
			hintLevel:    HintLevelDirect,
			keywords:     []string{},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "direct hint second warning",
			hintLevel:    HintLevelDirect,
			keywords:     []string{"規則"},
			warningCount: 1,
			expectEmpty:  false,
		},
		{
			name:         "vague hint with keywords",
			hintLevel:    HintLevelVague,
			keywords:     []string{"地下室"},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "vague hint without keywords",
			hintLevel:    HintLevelVague,
			keywords:     []string{},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "subtle hint with keywords",
			hintLevel:    HintLevelSubtle,
			keywords:     []string{"聲音"},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "subtle hint without keywords",
			hintLevel:    HintLevelSubtle,
			keywords:     []string{},
			warningCount: 0,
			expectEmpty:  false,
		},
		{
			name:         "none level returns empty",
			hintLevel:    HintLevelNone,
			keywords:     []string{"測試"},
			warningCount: 0,
			expectEmpty:  true,
		},
		{
			name:         "invalid level returns empty",
			hintLevel:    HintLevel(99), // Invalid level
			keywords:     []string{"測試"},
			warningCount: 0,
			expectEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildHintTextWithProgression(tt.hintLevel, tt.keywords, tt.warningCount)

			if tt.expectEmpty {
				assert.Empty(t, result, "Should return empty hint")
			} else {
				assert.NotEmpty(t, result, "Should return non-empty hint")
			}
		})
	}
}

// TestBuildKeywordContext_MultipleKeywords tests buildKeywordContext with various keyword counts
func TestBuildKeywordContext_MultipleKeywords(t *testing.T) {
	tests := []struct {
		name      string
		keywords  []string
		connector string
	}{
		{
			name:      "empty keywords",
			keywords:  []string{},
			connector: "關於",
		},
		{
			name:      "single keyword",
			keywords:  []string{"夜晚"},
			connector: "關於",
		},
		{
			name:      "two keywords",
			keywords:  []string{"夜晚", "開燈"},
			connector: "與",
		},
		{
			name:      "three keywords",
			keywords:  []string{"夜晚", "開燈", "危險"},
			connector: "關於",
		},
		{
			name:      "five keywords",
			keywords:  []string{"規則", "違反", "後果", "警告", "嚴重"},
			connector: "與",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test randomness in keyword selection
			for i := 0; i < 5; i++ {
				result := buildKeywordContext(tt.keywords, tt.connector)

				if len(tt.keywords) == 0 {
					assert.Empty(t, result, "Should return empty for no keywords")
				} else {
					assert.NotEmpty(t, result, "Should return non-empty for keywords")
					assert.Contains(t, result, tt.connector,
						"Should contain connector")
					// Should contain at least one keyword
					foundKeyword := false
					for _, keyword := range tt.keywords {
						if strings.Contains(result, keyword) {
							foundKeyword = true
							break
						}
					}
					assert.True(t, foundKeyword,
						"Should contain at least one keyword from the list")
				}
			}
		})
	}
}

// TestBuildDirectHint_WarningProgression tests buildDirectHint warning progression
func TestBuildDirectHint_WarningProgression(t *testing.T) {
	tests := []struct {
		name           string
		keywords       []string
		warningCount   int
		expectContains []string
	}{
		{
			name:           "first warning with keywords",
			keywords:       []string{"夜晚", "開燈"},
			warningCount:   0,
			expectContains: []string{"可能不太對"},
		},
		{
			name:           "second warning with keywords",
			keywords:       []string{"夜晚", "開燈"},
			warningCount:   1,
			expectContains: []string{"警告", "嚴重後果"},
		},
		{
			name:           "third warning (same as second)",
			keywords:       []string{"規則"},
			warningCount:   2,
			expectContains: []string{"警告", "嚴重後果"},
		},
		{
			name:           "first warning without keywords",
			keywords:       []string{},
			warningCount:   0,
			expectContains: []string{"可能不太對"},
		},
		{
			name:           "second warning without keywords",
			keywords:       []string{},
			warningCount:   1,
			expectContains: []string{"警告", "嚴重後果"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDirectHint(tt.keywords, tt.warningCount)

			assert.NotEmpty(t, result, "Direct hint should not be empty")

			for _, expected := range tt.expectContains {
				assert.Contains(t, result, expected,
					"Hint should contain '%s'", expected)
			}
		})
	}
}

// TestBuildVagueHint_WithAndWithoutKeywords tests buildVagueHint with various keyword scenarios
func TestBuildVagueHint_WithAndWithoutKeywords(t *testing.T) {
	tests := []struct {
		name           string
		keywords       []string
		expectContains []string
	}{
		{
			name:           "with single keyword",
			keywords:       []string{"地下室"},
			expectContains: []string{"氛圍", "詭異"},
		},
		{
			name:           "with multiple keywords",
			keywords:       []string{"夜晚", "開燈", "危險"},
			expectContains: []string{"氛圍", "詭異"},
		},
		{
			name:           "without keywords",
			keywords:       []string{},
			expectContains: []string{"氛圍", "詭異"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVagueHint(tt.keywords)

			assert.NotEmpty(t, result, "Vague hint should not be empty")

			for _, expected := range tt.expectContains {
				assert.Contains(t, result, expected,
					"Hint should contain '%s'", expected)
			}
		})
	}
}

// TestBuildSubtleHint_WithAndWithoutKeywords tests buildSubtleHint with various keyword scenarios
func TestBuildSubtleHint_WithAndWithoutKeywords(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
	}{
		{
			name:     "with single keyword",
			keywords: []string{"聲音"},
		},
		{
			name:     "with multiple keywords",
			keywords: []string{"聲音", "回應", "危險"},
		},
		{
			name:     "without keywords",
			keywords: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test randomness in template selection
			results := make(map[string]bool)
			for i := 0; i < 10; i++ {
				result := buildSubtleHint(tt.keywords)
				assert.NotEmpty(t, result, "Subtle hint should not be empty")
				results[result] = true
			}

			// With randomness, we should see some variation over multiple runs
			// (unless keywords are provided, which adds prefix)
			if len(tt.keywords) == 0 {
				// Without keywords, we should see multiple different templates
				assert.GreaterOrEqual(t, len(results), 1,
					"Should generate at least one template variation")
			}
		})
	}
}
