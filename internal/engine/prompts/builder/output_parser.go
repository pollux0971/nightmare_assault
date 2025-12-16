package builder

import (
	"encoding/json"
	"strings"
)

// SeedInfo represents a parsed hidden seed.
type SeedInfo struct {
	Type        string // Item, Event, Character, Location
	Description string
}

// ChoiceContext represents the choice area context.
type ChoiceContext struct {
	Situation string   `json:"situation"`          // 1-2 sentences describing current situation
	Question  string   `json:"question,omitempty"` // Optional guiding question
	Options   []string `json:"options"`            // 2-3 choice options
}

// StateChanges represents HP/SAN changes.
type StateChanges struct {
	HP     int    `json:"hp,omitempty"`  // HP change (positive = gain, negative = loss)
	SAN    int    `json:"san,omitempty"` // SAN change (positive = gain, negative = loss)
	Reason string `json:"reason,omitempty"`
}

// StoryOutput represents the structured output from LLM.
type StoryOutput struct {
	Story         string         `json:"story"`                    // Pure narrative content
	ChoiceContext *ChoiceContext `json:"choice_context,omitempty"` // Choice context (situation + question + options)
	Seeds         []SeedInfo     `json:"seeds,omitempty"`          // Hidden seeds for future plot
	StateChanges  *StateChanges  `json:"state_changes,omitempty"`  // HP/SAN changes

	// Legacy support - kept for backward compatibility with old format
	Choices []string `json:"choices,omitempty"`
}

// DeathOutput represents the structured death narrative output from LLM.
type DeathOutput struct {
	Narrative string   `json:"narrative"` // Death narrative text (200-300 words)
	Cause     string   `json:"cause"`     // Death cause: "hp_zero", "insanity", "rule_violation"
	Hints     []string `json:"hints"`     // Clues about hidden rules or survival tips
}

// DreamOutput represents the structured dream content output from LLM.
type DreamOutput struct {
	Dream        string   `json:"dream"`         // Dream narrative text (150-250 words)
	Symbols      []string `json:"symbols"`       // Symbolic elements in the dream
	RulesHinted  []int    `json:"rules_hinted"`  // IDs of rules hinted at (optional)
	Atmosphere   string   `json:"atmosphere"`    // Atmosphere: "calm", "uneasy", "nightmare", "grief"
}

// ParseStructuredOutput attempts to parse JSON output from LLM.
// If JSON parsing fails, falls back to legacy text parsing.
func ParseStructuredOutput(content string) (*StoryOutput, error) {
	// Try JSON parsing first
	var output StoryOutput

	// Look for JSON block (might be wrapped in markdown code blocks)
	jsonContent := extractJSONFromMarkdown(content)

	err := json.Unmarshal([]byte(jsonContent), &output)
	if err == nil && output.Story != "" {
		// Successfully parsed JSON

		// Backward compatibility: convert old format (choices array) to new format (choice_context)
		if output.ChoiceContext == nil && len(output.Choices) > 0 {
			output.ChoiceContext = &ChoiceContext{
				Situation: "", // No situation in old format
				Question:  "", // No question in old format
				Options:   output.Choices,
			}
		}

		return &output, nil
	}

	// Fallback to legacy text parsing
	return parseLegacyFormat(content), nil
}

// ParseDeathOutput attempts to parse JSON death narrative output from LLM.
// If JSON parsing fails, falls back to treating entire content as narrative.
func ParseDeathOutput(content string) (*DeathOutput, error) {
	// Try JSON parsing first
	var output DeathOutput

	// Look for JSON block (might be wrapped in markdown code blocks)
	jsonContent := extractJSONFromMarkdown(content)

	err := json.Unmarshal([]byte(jsonContent), &output)
	if err == nil && output.Narrative != "" {
		// Successfully parsed JSON
		return &output, nil
	}

	// Fallback: treat entire content as narrative (legacy format)
	return &DeathOutput{
		Narrative: strings.TrimSpace(content),
		Cause:     "unknown",
		Hints:     []string{},
	}, nil
}

// ParseDreamOutput attempts to parse JSON dream content output from LLM.
// If JSON parsing fails, falls back to treating entire content as dream text.
func ParseDreamOutput(content string) (*DreamOutput, error) {
	// Try JSON parsing first
	var output DreamOutput

	// Look for JSON block (might be wrapped in markdown code blocks)
	jsonContent := extractJSONFromMarkdown(content)

	err := json.Unmarshal([]byte(jsonContent), &output)
	if err == nil && output.Dream != "" {
		// Successfully parsed JSON
		return &output, nil
	}

	// Fallback: treat entire content as dream narrative (legacy format)
	return &DreamOutput{
		Dream:       strings.TrimSpace(content),
		Symbols:     []string{},
		RulesHinted: []int{},
		Atmosphere:  "unknown",
	}, nil
}

// extractJSONFromMarkdown extracts JSON from markdown code blocks if present.
func extractJSONFromMarkdown(content string) string {
	// Remove markdown code block markers if present
	content = strings.TrimSpace(content)

	// Check for ```json ... ``` blocks
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		if idx := strings.Index(content, "```"); idx != -1 {
			content = content[:idx]
		}
		return strings.TrimSpace(content)
	}

	// Check for ``` ... ``` blocks
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		if idx := strings.Index(content, "```"); idx != -1 {
			content = content[:idx]
		}
		return strings.TrimSpace(content)
	}

	// Try to find JSON object directly
	if strings.HasPrefix(content, "{") {
		// Content might already be pure JSON
		return content
	}

	// Look for JSON object within text
	startIdx := strings.Index(content, "{")
	if startIdx != -1 {
		// Find matching closing brace
		depth := 0
		for i := startIdx; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					return content[startIdx : i+1]
				}
			}
		}
	}

	return content
}

// parseLegacyFormat parses old-style text output (for backward compatibility).
func parseLegacyFormat(content string) *StoryOutput {
	// Extract seeds
	seeds := ExtractSeeds(content)

	// Clean content (remove seeds and choices)
	story := CleanContent(content)

	// Parse choices
	choices := parseChoicesFromText(content)

	// Convert to new format with ChoiceContext
	var choiceContext *ChoiceContext
	if len(choices) > 0 {
		choiceContext = &ChoiceContext{
			Situation: "", // Legacy format doesn't have situation
			Question:  "", // Legacy format doesn't have question
			Options:   choices,
		}
	}

	return &StoryOutput{
		Story:         story,
		ChoiceContext: choiceContext,
		Seeds:         seeds,
		Choices:       choices, // Keep for legacy compatibility
	}
}

// ExtractSeeds parses hidden seed markers from story content.
func ExtractSeeds(content string) []SeedInfo {
	var seeds []SeedInfo

	// Find all <!-- SEED:type:description --> markers
	lines := strings.Split(content, "<!--")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "SEED:") {
			endIdx := strings.Index(line, "-->")
			if endIdx == -1 {
				continue
			}
			seedStr := strings.TrimSpace(line[:endIdx])
			seedStr = strings.TrimPrefix(seedStr, "SEED:")

			parts := strings.SplitN(seedStr, ":", 2)
			if len(parts) == 2 {
				seeds = append(seeds, SeedInfo{
					Type:        strings.TrimSpace(parts[0]),
					Description: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return seeds
}

// CleanContent removes seed markers and choice sections from content for display.
func CleanContent(content string) string {
	result := content

	// First, try to extract just the story content if it's in JSON format
	// This handles cases where JSON parsing failed but we still want clean story text
	if strings.Contains(result, "```json") || strings.Contains(result, "```") {
		// Try to extract and parse JSON
		jsonContent := extractJSONFromMarkdown(result)
		var output StoryOutput
		if err := json.Unmarshal([]byte(jsonContent), &output); err == nil && output.Story != "" {
			// Successfully extracted story from JSON
			return strings.TrimSpace(output.Story)
		}
		// If JSON parsing failed, remove the code block markers manually
		result = strings.ReplaceAll(result, "```json", "")
		result = strings.ReplaceAll(result, "```", "")
	}

	// Remove seed markers
	for {
		start := strings.Index(result, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "-->")
		if end == -1 {
			break
		}
		end += start + 3
		result = result[:start] + result[end:]
	}

	// Remove choice sections (everything from "選擇：" or "**選擇：**" onwards)
	result = removeChoiceSection(result)

	return strings.TrimSpace(result)
}

// removeChoiceSection removes the choice section from story content.
func removeChoiceSection(content string) string {
	// Split content into lines
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	inChoiceSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect choice section start
		if containsChoiceHeader(trimmed) {
			inChoiceSection = true
			continue
		}

		// Skip lines in choice section
		if inChoiceSection {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// containsChoiceHeader detects if a line is a choice section header.
func containsChoiceHeader(line string) bool {
	// Remove markdown formatting
	line = strings.ReplaceAll(line, "*", "")
	line = strings.TrimSpace(line)

	// Check for Chinese choice headers
	chineseHeaders := []string{"選擇：", "選擇:", "選項：", "選項:", "请选择：", "请选择:"}
	for _, header := range chineseHeaders {
		if strings.Contains(line, header) {
			return true
		}
	}

	// Check for English choice headers (case insensitive)
	lowerLine := strings.ToLower(line)
	englishHeaders := []string{"choices:", "options:", "choose:", "select:"}
	for _, header := range englishHeaders {
		if strings.Contains(lowerLine, header) {
			return true
		}
	}

	return false
}

// parseChoicesFromText extracts choices from legacy text format.
func parseChoicesFromText(content string) []string {
	// First, try to extract choices from JSON if present
	if strings.Contains(content, "```json") || strings.Contains(content, "```") || strings.Contains(content, "\"choices\"") {
		jsonContent := extractJSONFromMarkdown(content)
		var output StoryOutput
		if err := json.Unmarshal([]byte(jsonContent), &output); err == nil && len(output.Choices) > 0 {
			// Successfully extracted choices from JSON
			return output.Choices
		}
	}

	// Fallback to text parsing
	var choices []string
	lines := strings.Split(content, "\n")

	inChoices := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if inChoices {
			// Match numbered choices: 1. or 1) or 1、
			if isNumberedChoice(line) {
				choice := extractChoiceText(line)
				if choice != "" {
					choices = append(choices, choice)
				}
				continue
			} else if line == "" {
				// Empty line might end choices
				if len(choices) > 0 {
					break
				}
			}
		}

		// Detect choice section start (check after numbered choices)
		if containsChoiceHeader(line) {
			inChoices = true
		}
	}

	return choices
}

// isNumberedChoice checks if a line is a numbered choice.
func isNumberedChoice(line string) bool {
	if len(line) < 2 {
		return false
	}
	// Check for digit followed by separator
	if line[0] >= '1' && line[0] <= '9' {
		if line[1] == '.' || line[1] == ')' {
			return true
		}
		// Check for Chinese separator (、is multi-byte)
		runes := []rune(line)
		if len(runes) >= 2 && runes[1] == '、' {
			return true
		}
	}
	return false
}

// extractChoiceText extracts the text from a numbered choice line.
func extractChoiceText(line string) string {
	runes := []rune(line)
	if len(runes) < 2 {
		return ""
	}
	// Skip "1. " or "1) " or "1、"
	text := string(runes[2:])
	if len(text) > 0 && text[0] == ' ' {
		text = text[1:]
	}
	return strings.TrimSpace(text)
}
