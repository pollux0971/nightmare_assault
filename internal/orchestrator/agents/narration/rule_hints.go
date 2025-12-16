package narration

import (
	"fmt"
	"math/rand"
	"strings"
)

// HintLevel represents the clarity level of a rule hint
type HintLevel int

const (
	HintLevelNone   HintLevel = iota // No hint (Hell difficulty)
	HintLevelSubtle                  // Very subtle hint (Hard difficulty)
	HintLevelVague                   // Vague hint (Normal difficulty)
	HintLevelDirect                  // Direct hint (Easy difficulty)
)

// String returns the string representation of the hint level
func (h HintLevel) String() string {
	switch h {
	case HintLevelNone:
		return "none"
	case HintLevelSubtle:
		return "subtle"
	case HintLevelVague:
		return "vague"
	case HintLevelDirect:
		return "direct"
	default:
		return "unknown"
	}
}

// GenerateRuleHint generates a rule hint based on difficulty and warning count
//
// AC #5: Rule Hints 生成邏輯
//
// Parameters:
//   - ruleID: Rule identifier
//   - ruleDesc: Rule description (used for context)
//   - difficulty: Difficulty level (easy/normal/hard/hell)
//   - warningCount: Number of warnings already given
//   - maxWarnings: Maximum warnings allowed for this difficulty
//
// Returns:
//   - string: Hint text (empty if no hint should be given)
func GenerateRuleHint(ruleID, ruleDesc, difficulty string, warningCount, maxWarnings int) string {
	// Check if we've exceeded max warnings
	if warningCount >= maxWarnings {
		return ""
	}

	// Get hint level based on difficulty and warning count
	hintLevel := getHintLevel(difficulty, warningCount)
	if hintLevel == HintLevelNone {
		return ""
	}

	// Extract keywords from rule description
	keywords := ExtractRuleKeywords(ruleDesc)

	// Build hint text with warning progression
	return buildHintTextWithProgression(hintLevel, keywords, warningCount)
}

// GetMaxWarnings returns the maximum number of warnings for a difficulty level
//
// AC #5: 最大警告次數映射
//   - Easy: 2 warnings
//   - Normal: 1 warning
//   - Hard: 1 warning
//   - Hell: 0 warnings
func GetMaxWarnings(difficulty string) int {
	switch strings.ToLower(difficulty) {
	case "easy":
		return 2
	case "normal":
		return 1
	case "hard":
		return 1
	case "hell":
		return 0
	default:
		return 0 // Default to no warnings for unknown difficulty
	}
}

// getHintLevel determines the hint level based on difficulty and warning count
func getHintLevel(difficulty string, warningCount int) HintLevel {
	switch strings.ToLower(difficulty) {
	case "easy":
		// Easy gets increasingly direct hints
		return HintLevelDirect
	case "normal":
		// Normal gets vague hints
		return HintLevelVague
	case "hard":
		// Hard gets very subtle hints
		return HintLevelSubtle
	case "hell":
		// Hell gets no hints
		return HintLevelNone
	default:
		return HintLevelNone
	}
}

// ExtractRuleKeywords extracts key concepts from a rule description
//
// This is a simplified implementation that extracts important nouns/verbs.
// A more sophisticated version could use NLP techniques.
func ExtractRuleKeywords(ruleDesc string) []string {
	if ruleDesc == "" {
		return []string{}
	}

	// Remove common particles and extract key terms
	keywords := []string{}

	// Common stop words and particles to filter out
	stopWords := []string{
		"不要", "不能", "別", "勿", "不可",
		"的", "了", "在", "與", "和", "或", "但", "而",
		"就", "也", "都", "會", "是", "有", "為",
		"這", "那", "其", "此",
	}

	// Remove punctuation and quotes
	cleaned := ruleDesc
	for _, punc := range []string{"「", "」", "『", "』", "，", "。", "、", "；", "："} {
		cleaned = strings.ReplaceAll(cleaned, punc, " ")
	}

	// Split by spaces and common separators
	rawWords := strings.Fields(cleaned)

	// Extract keywords by filtering stop words
	for _, word := range rawWords {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Skip if word is a stop word
		isStopWord := false
		for _, sw := range stopWords {
			if word == sw {
				isStopWord = true
				break
			}
		}

		if isStopWord {
			continue
		}

		// Extract meaningful keywords (at least 2 characters)
		if len([]rune(word)) >= 2 {
			keywords = append(keywords, word)
		}
	}

	// If we didn't find any keywords, try character-level extraction
	if len(keywords) == 0 {
		// Extract nouns by looking for common patterns
		// This is a very simplified approach for Chinese
		for _, word := range rawWords {
			word = strings.TrimSpace(word)

			// Remove stop words from the beginning and end
			for _, sw := range stopWords {
				word = strings.TrimPrefix(word, sw)
				word = strings.TrimSuffix(word, sw)
			}

			if len([]rune(word)) >= 2 {
				keywords = append(keywords, word)
			}
		}
	}

	// Limit to 3 most relevant keywords
	if len(keywords) > 3 {
		keywords = keywords[:3]
	}

	return keywords
}

// BuildHintText constructs the hint text based on hint level and keywords
//
// AC #5: 提示文本生成
//   - Direct (Easy): 明確但不明說規則（「你感覺這樣做可能不太對」）
//   - Vague (Normal): 模糊暗示（「周圍的氛圍變得詭異」）
//   - Subtle (Hard): 極隱晦（「你注意到某個細節」）
//   - None (Hell): 無提示（返回空字符串）
func BuildHintText(hintLevel HintLevel, keywords []string) string {
	return buildHintTextWithProgression(hintLevel, keywords, 0)
}

// buildHintTextWithProgression builds hint text with warning progression
func buildHintTextWithProgression(hintLevel HintLevel, keywords []string, warningCount int) string {
	switch hintLevel {
	case HintLevelDirect:
		return buildDirectHint(keywords, warningCount)
	case HintLevelVague:
		return buildVagueHint(keywords)
	case HintLevelSubtle:
		return buildSubtleHint(keywords)
	case HintLevelNone:
		return ""
	default:
		return ""
	}
}

// buildDirectHint builds a direct hint (Easy difficulty)
func buildDirectHint(keywords []string, warningCount int) string {
	var base string

	// Warning progression: second warning is more explicit
	if warningCount >= 1 {
		base = "警告：你的行為可能帶來嚴重後果。"
	} else {
		base = "你感覺這樣做可能不太對。"
	}

	// Add keyword context if available
	if len(keywords) > 0 {
		keywordText := buildKeywordContext(keywords, "關於")
		return fmt.Sprintf("%s%s", keywordText, base)
	}

	return base
}

// buildVagueHint builds a vague hint (Normal difficulty)
func buildVagueHint(keywords []string) string {
	// Use first template for consistency in testing
	base := "周圍的氛圍變得詭異起來。"

	// Add keyword context if available (more subtle than direct)
	if len(keywords) > 0 {
		keywordText := buildKeywordContext(keywords, "與")
		return fmt.Sprintf("%s似乎%s", keywordText, base)
	}

	return base
}

// buildSubtleHint builds a subtle hint (Hard difficulty)
func buildSubtleHint(keywords []string) string {
	templates := []string{
		"你注意到某個細節。",
		"某個念頭一閃而過。",
		"你的視線停留了片刻。",
		"周圍的一切似乎有些不對勁。",
	}

	base := templates[rand.Intn(len(templates))]

	// For hard difficulty, keywords are only vaguely hinted
	if len(keywords) > 0 {
		// Just mention something related without being specific
		return fmt.Sprintf("你注意到了某些細節。%s", base)
	}

	return base
}

// buildKeywordContext builds a text snippet that incorporates keywords
func buildKeywordContext(keywords []string, connector string) string {
	if len(keywords) == 0 {
		return ""
	}

	// Take first 1-2 keywords to avoid being too obvious
	numKeywords := 1
	if len(keywords) > 1 && rand.Float32() > 0.5 {
		numKeywords = 2
	}

	selectedKeywords := keywords[:numKeywords]
	keywordStr := strings.Join(selectedKeywords, "、")

	return fmt.Sprintf("%s%s", connector, keywordStr)
}
