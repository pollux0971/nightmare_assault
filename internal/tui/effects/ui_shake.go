package effects

import (
	"math/rand"
	"strings"
)

// RenderShakingBorder renders content with a shaking border effect.
// offset determines the shake amplitude (0-5 pixels).
//
// The shake is achieved by adding/removing padding randomly.
func RenderShakingBorder(content string, offset int) string {
	if offset == 0 {
		return content
	}

	// Random offset for this frame
	xOffset := rand.Intn(offset+1) - offset/2 // -offset/2 to +offset/2
	if xOffset < 0 {
		xOffset = 0
	}

	// Add padding to simulate shake
	padding := strings.Repeat(" ", xOffset)

	lines := strings.Split(content, "\n")
	var result strings.Builder

	for _, line := range lines {
		result.WriteString(padding)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// CalculateShakeOffset returns the current shake offset based on UIStability.
// This should be called each frame to create animated shaking.
func CalculateShakeOffset(stability int) int {
	if stability == 0 {
		return 0
	}

	// Random shake within stability range
	return rand.Intn(stability + 1)
}
