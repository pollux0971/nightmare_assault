package effects

import (
	"math/rand"
	"strings"
)

// Unicode combining diacritical marks (U+0300–U+036F)
// These are "stacked" on top of base characters to create the Zalgo effect
var combiningCharacters = []rune{
	// Combining marks above
	0x0300, 0x0301, 0x0302, 0x0303, 0x0304, 0x0305, 0x0306, 0x0307,
	0x0308, 0x0309, 0x030A, 0x030B, 0x030C, 0x030D, 0x030E, 0x030F,
	0x0310, 0x0311, 0x0312, 0x0313, 0x0314, 0x0315, 0x0316, 0x0317,
	// Combining marks below
	0x0318, 0x0319, 0x031A, 0x031B, 0x031C, 0x031D, 0x031E, 0x031F,
	0x0320, 0x0321, 0x0322, 0x0323, 0x0324, 0x0325, 0x0326, 0x0327,
	0x0328, 0x0329, 0x032A, 0x032B, 0x032C, 0x032D, 0x032E, 0x032F,
	// More combining marks
	0x0330, 0x0331, 0x0332, 0x0333, 0x0334, 0x0335, 0x0336, 0x0337,
	0x0338, 0x0339, 0x033A, 0x033B, 0x033C, 0x033D, 0x033E, 0x033F,
}

const maxCombiningPerChar = 3 // Dev Notes: 限制每字符最多 3 個組合標記

// ApplyZalgo applies the "Zalgo text" effect to the given string based on intensity.
// Zalgo text uses Unicode combining characters to create a disturbing, corrupted appearance.
//
// Parameters:
//   - text: The original text to corrupt
//   - intensity: Corruption level from 0.0 (none) to 1.0 (maximum)
//
// Returns the corrupted text with combining characters applied.
//
// The function:
//   - At 0.0 intensity: returns text unchanged
//   - At 1.0 intensity: affects nearly all characters with max combining marks
//   - At intermediate values: proportionally affects characters
//
// Example:
//
//	ApplyZalgo("Hello", 0.5) might return "H̴e̷l̸l̵o̶"
func ApplyZalgo(text string, intensity float64) string {
	if intensity <= 0.0 || text == "" {
		return text
	}

	// Clamp intensity to valid range
	if intensity > 1.0 {
		intensity = 1.0
	}

	var result strings.Builder
	runes := []rune(text)

	for _, char := range runes {
		// Always add the original character
		result.WriteRune(char)

		// Decide if this character gets corrupted based on intensity
		if rand.Float64() < intensity {
			// Calculate how many combining marks to add (1-3 based on intensity)
			numMarks := 1 + int(intensity*float64(maxCombiningPerChar-1))
			if numMarks > maxCombiningPerChar {
				numMarks = maxCombiningPerChar
			}

			// Add random combining characters
			for i := 0; i < numMarks; i++ {
				mark := combiningCharacters[rand.Intn(len(combiningCharacters))]
				result.WriteRune(mark)
			}
		}
	}

	return result.String()
}
