package input

import (
	"errors"
	"strings"
)

// InputSanitizer handles input validation and sanitization.
type InputSanitizer struct {
	MaxLength       int
	MinLength       int
	BlockedPatterns []string
}

// Error definitions for sanitization
var (
	ErrLLMInjection = errors.New("偵測到潛在的注入攻擊")
)

// NewInputSanitizer creates a new input sanitizer with default settings.
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		MaxLength: 200,
		MinLength: 3,
		BlockedPatterns: []string{
			"ignore previous",
			"system:",
			"assistant:",
			"user:",
			"<|",
			"|>",
		},
	}
}

// Sanitize performs full sanitization pipeline.
func (s *InputSanitizer) Sanitize(input string) (string, error) {
	// 1. Trim whitespace
	trimmed := s.TrimWhitespace(input)

	// 2. Check if empty
	if len(trimmed) == 0 {
		return "", ErrEmptyInput
	}

	// 3. Check length
	if err := s.CheckLength(trimmed); err != nil {
		return "", err
	}

	// 4. Check blocked patterns
	if err := s.CheckBlockedPatterns(trimmed); err != nil {
		return "", err
	}

	// 5. Escape special characters
	escaped := s.EscapeSpecialChars(trimmed)

	return escaped, nil
}

// CheckLength validates input length.
func (s *InputSanitizer) CheckLength(input string) error {
	length := RuneCount(input)

	if length < s.MinLength {
		return ErrInputTooShort
	}
	if length > s.MaxLength {
		return ErrInputTooLong
	}

	return nil
}

// EscapeSpecialChars escapes potentially dangerous characters.
func (s *InputSanitizer) EscapeSpecialChars(input string) string {
	replacer := strings.NewReplacer(
		"<", "\\<",
		">", "\\>",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(input)
}

// CheckBlockedPatterns checks for LLM injection patterns.
func (s *InputSanitizer) CheckBlockedPatterns(input string) error {
	lower := strings.ToLower(input)

	for _, pattern := range s.BlockedPatterns {
		if strings.Contains(lower, pattern) {
			return ErrLLMInjection
		}
	}

	return nil
}

// TrimWhitespace removes leading and trailing whitespace.
func (s *InputSanitizer) TrimWhitespace(input string) string {
	return strings.TrimSpace(input)
}

// RuneCount counts the number of runes (characters) in a string.
func RuneCount(s string) int {
	return len([]rune(s))
}
