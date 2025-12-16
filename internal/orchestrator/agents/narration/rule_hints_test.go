package narration

import (
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
