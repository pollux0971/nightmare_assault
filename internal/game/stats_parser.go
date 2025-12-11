package game

import (
	"encoding/json"
	"regexp"
	"strings"
)

// StatsParser extracts HP/SAN changes from LLM responses.
type StatsParser struct {
	hpRegex  *regexp.Regexp
	sanRegex *regexp.Regexp
}

// NewStatsParser creates a new stats parser.
func NewStatsParser() *StatsParser {
	return &StatsParser{
		hpRegex:  regexp.MustCompile(`HP:\s*([+-]?\d+)`),
		sanRegex: regexp.MustCompile(`SAN:\s*([+-]?\d+)`),
	}
}

// Parse attempts to extract stat changes using multiple strategies.
func (sp *StatsParser) Parse(text string) (hp, san int) {
	// Strategy 1: Try inline text markers (HP: -15, SAN: -5)
	hp, san, found := sp.ParseInlineText(text)
	if found {
		return sp.ValidateBounds(hp, san)
	}

	// Strategy 2: Try JSON format
	hp, san, found = sp.ParseJSON(text)
	if found {
		return sp.ValidateBounds(hp, san)
	}

	// Strategy 3: Fallback to keyword detection
	hp, san = sp.ParseKeywords(text)
	return sp.ValidateBounds(hp, san)
}

// ParseInlineText extracts stat changes from inline markers.
func (sp *StatsParser) ParseInlineText(text string) (hp, san int, found bool) {
	hpMatch := sp.hpRegex.FindStringSubmatch(text)
	sanMatch := sp.sanRegex.FindStringSubmatch(text)

	if len(hpMatch) > 1 {
		found = true
		hp = parseInt(hpMatch[1])
	}

	if len(sanMatch) > 1 {
		found = true
		san = parseInt(sanMatch[1])
	}

	return hp, san, found
}

// ParseJSON extracts stat changes from JSON format.
func (sp *StatsParser) ParseJSON(text string) (hp, san int, found bool) {
	var data struct {
		StatChanges struct {
			HP  int `json:"hp"`
			SAN int `json:"san"`
		} `json:"stat_changes"`
	}

	err := json.Unmarshal([]byte(text), &data)
	if err != nil {
		return 0, 0, false
	}

	hp = data.StatChanges.HP
	san = data.StatChanges.SAN

	// Consider found if at least one value is non-zero
	found = hp != 0 || san != 0

	return hp, san, found
}

// ParseKeywords extracts estimated stat changes from keywords.
func (sp *StatsParser) ParseKeywords(text string) (hp, san int) {
	lower := strings.ToLower(text)

	// HP keywords (damage) - ordered from longest to shortest for proper matching
	hpKeywords := []struct {
		keyword string
		delta   int
	}{
		{"lose significant health", -15},
		{"critically wounded", -20},
		{"take heavy damage", -15},
		{"badly wounded", -15},
		{"wounded badly", -15},
		{"lose health", -10},
		{"take damage", -10},
		{"wounded", -10},
		{"injured", -10},
		{"hurt", -5},
	}

	for _, kw := range hpKeywords {
		if strings.Contains(lower, kw.keyword) {
			hp = kw.delta
			break
		}
	}

	// SAN keywords (terror/fear) - ordered from longest to shortest
	sanKeywords := []struct {
		keyword string
		delta   int
	}{
		{"overwhelming fear", -15},
		{"terror", -15},
		{"horrifying", -15},
		{"madness", -20},
		{"insanity", -20},
		{"fear", -10},
		{"panic", -10},
		{"anxious", -5},
		{"disturbing", -5},
	}

	for _, kw := range sanKeywords {
		if strings.Contains(lower, kw.keyword) {
			san = kw.delta
			break
		}
	}

	return hp, san
}

// ValidateBounds ensures stat changes are within acceptable ranges.
func (sp *StatsParser) ValidateBounds(hp, san int) (int, int) {
	// Negative changes (damage/drain): -50 to 0
	// Positive changes (healing): 0 to +30
	if hp < -50 {
		hp = -50
	} else if hp > 30 {
		hp = 30
	}

	if san < -50 {
		san = -50
	} else if san > 30 {
		san = 30
	}

	return hp, san
}

// parseInt parses a string to int, handling +/- signs.
func parseInt(s string) int {
	var result int
	var sign int = 1

	if len(s) == 0 {
		return 0
	}

	// Check for sign
	start := 0
	if s[0] == '+' {
		start = 1
	} else if s[0] == '-' {
		sign = -1
		start = 1
	}

	// Parse digits
	for i := start; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		}
	}

	return result * sign
}
