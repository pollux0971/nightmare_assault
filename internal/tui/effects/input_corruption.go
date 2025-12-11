package effects

import (
	"math/rand"
	"strings"
)

// CorruptedInput represents an input that has been corrupted by horror effects.
type CorruptedInput struct {
	Original       string  // Original input before corruption
	Corrupted      string  // Input after corruption
	DeletionCount  int     // Number of characters deleted
	FlashWarning   bool    // Whether to show visual warning (red flash)
}

// ApplyInputCorruption applies horror-based corruption to user input.
// Based on TypingBehavior from HorrorStyle, randomly deletes characters.
//
// typingBehavior: 0.0 (no deletion) - 0.2 (20% deletion probability)
//
// Returns a CorruptedInput with:
//   - Corrupted string (some chars randomly deleted)
//   - Count of deleted characters
//   - FlashWarning flag if deletion occurred
func ApplyInputCorruption(input string, typingBehavior float64) CorruptedInput {
	result := CorruptedInput{
		Original:      input,
		Corrupted:     input,
		DeletionCount: 0,
		FlashWarning:  false,
	}

	// No corruption if typingBehavior is zero
	if typingBehavior <= 0.0 || len(input) == 0 {
		return result
	}

	// Apply random deletion to each character
	runes := []rune(input)
	var corrupted []rune
	deletedCount := 0

	for _, r := range runes {
		// Randomly decide whether to delete this character
		if rand.Float64() < typingBehavior {
			deletedCount++
			continue // Skip this character (delete it)
		}
		corrupted = append(corrupted, r)
	}

	result.Corrupted = string(corrupted)
	result.DeletionCount = deletedCount
	result.FlashWarning = deletedCount > 0

	return result
}

// GetInputCorruptionFeedback returns user-facing feedback text when input is corrupted.
// This provides narrative feedback in accessible mode or for user awareness.
func GetInputCorruptionFeedback(deletionCount int) string {
	switch {
	case deletionCount == 0:
		return ""
	case deletionCount == 1:
		return "[一個字元消失了...]"
	case deletionCount <= 3:
		return "[部分文字被吞噬...]"
	case deletionCount <= 5:
		return "[大量文字消失！]"
	default:
		return "[你的思緒正在崩潰...]"
	}
}

// ShouldCorruptInput determines if input should be corrupted based on current SAN.
// This is a helper function to check if corruption effects should be applied.
func ShouldCorruptInput(san int) bool {
	// Only corrupt input when SAN < 40 (panicked or insanity states)
	return san < 40
}

// ApplyTypingBehaviorEffect is a convenience wrapper that:
// 1. Checks if corruption should apply based on SAN
// 2. Calculates HorrorStyle
// 3. Applies corruption if needed
func ApplyTypingBehaviorEffect(input string, san int) CorruptedInput {
	if !ShouldCorruptInput(san) {
		return CorruptedInput{
			Original:      input,
			Corrupted:     input,
			DeletionCount: 0,
			FlashWarning:  false,
		}
	}

	style := CalculateHorrorStyle(san)
	return ApplyInputCorruption(input, style.TypingBehavior)
}

// CorruptRealTimeInput applies corruption to input as the user types (real-time).
// This is different from batch corruption - it simulates keys "not registering".
//
// previousBuffer: The input buffer before the latest keystroke
// newChar: The character just typed
// typingBehavior: The current typing behavior corruption value
//
// Returns: The new buffer (with or without the new character)
func CorruptRealTimeInput(previousBuffer string, newChar rune, typingBehavior float64) (newBuffer string, wasDeleted bool) {
	// No corruption if typingBehavior is zero
	if typingBehavior <= 0.0 {
		return previousBuffer + string(newChar), false
	}

	// Randomly decide whether to "drop" this keystroke
	if rand.Float64() < typingBehavior {
		// Character was "eaten" - don't add it to buffer
		return previousBuffer, true
	}

	// Character accepted normally
	return previousBuffer + string(newChar), false
}

// GetVisualFeedbackStyle returns styling hints for corrupted input feedback.
// This is used to render red flashing or other visual warnings.
type InputVisualFeedback struct {
	FlashColor    string // Color code for flash effect (e.g., "1" for red)
	FlashDuration int    // Duration in milliseconds
	ShowFeedback  bool   // Whether to show feedback text
	FeedbackText  string // The feedback message to display
}

// CalculateInputVisualFeedback generates visual feedback parameters for input corruption.
func CalculateInputVisualFeedback(corrupted CorruptedInput) InputVisualFeedback {
	if !corrupted.FlashWarning {
		return InputVisualFeedback{
			ShowFeedback: false,
		}
	}

	feedback := InputVisualFeedback{
		FlashColor:    "1",  // Red
		FlashDuration: 150,  // 150ms flash
		ShowFeedback:  true,
		FeedbackText:  GetInputCorruptionFeedback(corrupted.DeletionCount),
	}

	// Increase flash duration for severe corruption
	if corrupted.DeletionCount >= 5 {
		feedback.FlashDuration = 300 // Longer flash for severe corruption
	}

	return feedback
}

// GetCursorDesyncOffset returns cursor position desync offset for extreme SAN.
// When SAN is very low (1-19), the cursor may appear offset from actual position.
//
// This returns an offset in characters (can be negative or positive).
// The TUI should render the cursor at: (actualPosition + offset)
func GetCursorDesyncOffset(san int) int {
	if san >= 20 {
		return 0 // No desync above panic threshold
	}

	// In insanity state (1-19), cursor can desync by ±1-2 positions
	if san < 10 {
		// Severe desync
		return rand.Intn(5) - 2 // -2 to +2
	}

	// Moderate desync
	return rand.Intn(3) - 1 // -1 to +1
}

// CalculateInputBoxShrinkage returns the width percentage for input box based on SAN.
// AC2 requires input box to shrink to 85-90% when SAN is 40-59.
func CalculateInputBoxShrinkage(san int) float64 {
	switch {
	case san > 60:
		return 1.0 // 100% - no shrinkage
	case san >= 40:
		// 40-59 range (inclusive of 60): shrink to 85-90%
		// Interpolate within range
		normalizedSan := float64(san-40) / 20.0           // 0.0 (san=40) to 1.0 (san=60)
		shrinkage := 0.85 + (normalizedSan * 0.05)        // 0.85 to 0.90
		return shrinkage
	case san >= 20:
		// 20-39 range: shrink to 75-85%
		normalizedSan := float64(san-20) / 20.0
		shrinkage := 0.75 + (normalizedSan * 0.10)
		return shrinkage
	default:
		// 1-19 range: shrink to 60-75%
		normalizedSan := float64(san-1) / 19.0
		shrinkage := 0.60 + (normalizedSan * 0.15)
		return shrinkage
	}
}

// TruncateToShrunkWidth truncates input text to fit within shrunk input box.
// This is used to visually enforce the input box shrinkage effect.
func TruncateToShrunkWidth(text string, originalWidth int, shrinkage float64) string {
	newWidth := int(float64(originalWidth) * shrinkage)

	runes := []rune(text)
	if len(runes) <= newWidth {
		return text
	}

	// Truncate with ellipsis
	if newWidth <= 3 {
		return string(runes[:newWidth])
	}

	return string(runes[:newWidth-3]) + "..."
}

// SanitizeCorruptedInput removes any corruption artifacts from input before sending to AI.
// While corruption affects display, the actual input sent to the game should be clean.
//
// For now, this is a pass-through, but in future could strip Zalgo or fix encoding issues.
func SanitizeCorruptedInput(corrupted string) string {
	// Remove any Zalgo combining characters that might have leaked into input
	runes := []rune(corrupted)
	var cleaned []rune

	for _, r := range runes {
		// Skip combining diacritical marks (U+0300-U+036F)
		if r >= 0x0300 && r <= 0x036F {
			continue
		}
		cleaned = append(cleaned, r)
	}

	result := string(cleaned)

	// Trim whitespace
	result = strings.TrimSpace(result)

	return result
}
