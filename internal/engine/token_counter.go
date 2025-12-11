package engine

import (
	"strings"
	"unicode/utf8"
)

// TokenUsageInfo contains information about token usage.
type TokenUsageInfo struct {
	Current    int     `json:"current"`
	Limit      int     `json:"limit"`
	Percentage float64 `json:"percentage"`
	Remaining  int     `json:"remaining"`
}

// TokenMonitor tracks token usage against a limit.
type TokenMonitor struct {
	Limit        int
	CurrentUsage int
}

// NewTokenMonitor creates a new token monitor with the specified limit.
func NewTokenMonitor(limit int) *TokenMonitor {
	return &TokenMonitor{
		Limit:        limit,
		CurrentUsage: 0,
	}
}

// Add adds text to the monitor and returns the token count.
func (tm *TokenMonitor) Add(text string) int {
	tokens := CountTokens(text)
	tm.CurrentUsage += tokens
	return tokens
}

// Reset resets the current usage to zero.
func (tm *TokenMonitor) Reset() {
	tm.CurrentUsage = 0
}

// Percentage returns the current usage as a percentage of the limit.
func (tm *TokenMonitor) Percentage() float64 {
	if tm.Limit == 0 {
		return 0
	}
	return (float64(tm.CurrentUsage) / float64(tm.Limit)) * 100
}

// ShouldCompress returns true if usage exceeds the threshold percentage.
func (tm *TokenMonitor) ShouldCompress(threshold float64) bool {
	return tm.Percentage() >= threshold*100
}

// GetUsageInfo returns detailed usage information.
func (tm *TokenMonitor) GetUsageInfo() TokenUsageInfo {
	remaining := tm.Limit - tm.CurrentUsage
	if remaining < 0 {
		remaining = 0
	}

	return TokenUsageInfo{
		Current:    tm.CurrentUsage,
		Limit:      tm.Limit,
		Percentage: tm.Percentage(),
		Remaining:  remaining,
	}
}

// CountTokens estimates the number of tokens in the given text.
// This is a simplified implementation using character-based heuristics.
// For production, consider using official tokenizers like tiktoken for OpenAI models.
func CountTokens(text string) int {
	if text == "" {
		return 0
	}

	// Simple heuristic-based token counting
	// This is approximate and should be replaced with proper tokenizer for production
	return EstimateTokens(text)
}

// EstimateTokens provides a fast estimate of token count.
// Uses heuristics based on character count and language detection.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Count characters and words
	charCount := utf8.RuneCountInString(text)
	words := strings.Fields(text)
	wordCount := len(words)

	// Detect if text is primarily Chinese/Japanese (higher token-to-char ratio)
	cjkCount := 0
	for _, r := range text {
		if isCJK(r) {
			cjkCount++
		}
	}

	cjkRatio := float64(cjkCount) / float64(charCount)

	var tokens int
	if cjkRatio > 0.3 {
		// Primarily CJK text: roughly 1.5-2 tokens per character
		tokens = int(float64(charCount) * 1.7)
	} else {
		// Primarily Latin text: roughly 1 token per 4 characters
		// or about 1.3 tokens per word
		charBasedEstimate := charCount / 4
		wordBasedEstimate := int(float64(wordCount) * 1.3)

		// Use the average of both estimates
		tokens = (charBasedEstimate + wordBasedEstimate) / 2
	}

	// Ensure at least 1 token for non-empty text
	if tokens < 1 {
		tokens = 1
	}

	return tokens
}

// isCJK checks if a rune is a CJK (Chinese, Japanese, Korean) character.
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Extension B
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul
}

// TokenLimits contains common token limits for different models.
var TokenLimits = map[string]int{
	"gpt-4":           8192,
	"gpt-4-turbo":     128000,
	"gpt-3.5-turbo":   4096,
	"claude-3-opus":   200000,
	"claude-3-sonnet": 200000,
	"claude-3-haiku":  200000,
	"gemini-pro":      32000,
}

// GetTokenLimit returns the token limit for a given model.
func GetTokenLimit(model string) int {
	if limit, ok := TokenLimits[model]; ok {
		return limit
	}
	// Default fallback
	return 4096
}

// CalculateCompressionTarget calculates the target token count after compression.
// Typically aims for 50-60% of the limit.
func CalculateCompressionTarget(limit int) int {
	return int(float64(limit) * 0.55)
}
