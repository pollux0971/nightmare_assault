package effects

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/lipgloss"
)

// SANVisualEffects defines the complete visual effect profile for a given SAN level.
// Story 7.5: Maps SAN ranges to specific visual effects for gradual control deprivation.
type SANVisualEffects struct {
	SAN int

	// Input box effects (AC1-AC4)
	InputBoxScale     float64 // 1.0 = normal, 0.6 = 60% shrinkage
	BorderColor       string  // Lipgloss color code
	BorderShake       bool    // Whether border shakes
	ShakeIntensity    int     // 0-5 pixel shake amplitude

	// Cursor effects (AC3)
	CursorBlinkSpeed float64 // 1.0 = normal, 1.5 = 1.5x faster

	// Text effects (AC3, AC4)
	TextColorDarkness  float64 // 0.0 = normal, 1.0 = very dark
	TextColorUnstable  bool    // Whether text color randomly changes
	TextShake          bool    // Whether text shakes
	TextBlur           bool    // Whether text appears blurred
	TextGhost          bool    // Whether text has ghost/echo effect

	// Character deletion (AC4)
	CharDeletionRate float64 // 0.0 - 0.2, probability of deleting chars
}

// GetSANVisualEffects returns the visual effects configuration for a given SAN value.
// Story 7.5 AC1-AC4: Progressive visual effects mapping.
func GetSANVisualEffects(san int) SANVisualEffects {
	// Clamp SAN
	if san < 0 {
		san = 0
	}
	if san > 100 {
		san = 100
	}

	switch {
	case san >= 80:
		// AC1: SAN 80-100 - Clear state
		// No effects, everything normal
		return SANVisualEffects{
			SAN:                san,
			InputBoxScale:      1.0,
			BorderColor:        "240", // Normal gray
			BorderShake:        false,
			ShakeIntensity:     0,
			CursorBlinkSpeed:   1.0,
			TextColorDarkness:  0.0,
			TextColorUnstable:  false,
			TextShake:          false,
			TextBlur:           false,
			TextGhost:          false,
			CharDeletionRate:   0.0,
		}

	case san >= 60:
		// SAN 60-79 - Slight anxiety
		// Very minor visual changes
		return SANVisualEffects{
			SAN:                san,
			InputBoxScale:      1.0,
			BorderColor:        "244", // Slightly lighter gray
			BorderShake:        false,
			ShakeIntensity:     0,
			CursorBlinkSpeed:   1.0,
			TextColorDarkness:  0.0,
			TextColorUnstable:  false,
			TextShake:          false,
			TextBlur:           false,
			TextGhost:          false,
			CharDeletionRate:   0.0,
		}

	case san >= 40:
		// AC2: SAN 40-59 - Pressure state
		// Input box shrinks 20%, border changes to yellow/orange, slight shake
		return SANVisualEffects{
			SAN:                san,
			InputBoxScale:      0.80, // 20% shrinkage
			BorderColor:        "220", // Yellow/orange
			BorderShake:        true,
			ShakeIntensity:     1, // Slight shake
			CursorBlinkSpeed:   1.0,
			TextColorDarkness:  0.1,
			TextColorUnstable:  false,
			TextShake:          false,
			TextBlur:           false,
			TextGhost:          false,
			CharDeletionRate:   0.0,
		}

	case san >= 20:
		// AC3: SAN 20-39 - Anxiety state
		// Cursor blinks 1.5x faster, text color unstable, input box shrinks 40%
		// Text may shake slightly
		return SANVisualEffects{
			SAN:                san,
			InputBoxScale:      0.60, // 40% shrinkage
			BorderColor:        "202", // Orange-red
			BorderShake:        true,
			ShakeIntensity:     2, // Moderate shake
			CursorBlinkSpeed:   1.5,
			TextColorDarkness:  0.3,
			TextColorUnstable:  true,
			TextShake:          true,
			TextBlur:           false,
			TextGhost:          false,
			CharDeletionRate:   0.0,
		}

	default:
		// AC4: SAN 1-19 - Loss of control state
		// Severe visual interference, input box shrinks 60%
		// Text blur, ghost effect, color very dark
		// Random character deletion (10-20%)
		return SANVisualEffects{
			SAN:                san,
			InputBoxScale:      0.40, // 60% shrinkage
			BorderColor:        "196", // Red
			BorderShake:        true,
			ShakeIntensity:     5, // Intense shake
			CursorBlinkSpeed:   2.0,
			TextColorDarkness:  0.6,
			TextColorUnstable:  true,
			TextShake:          true,
			TextBlur:           true,
			TextGhost:          true,
			CharDeletionRate:   0.15, // 15% base deletion rate (10-20% range)
		}
	}
}

// ApplyInputBoxStyle applies visual effects to the input box border style.
// Story 7.5 AC2, AC4: Border color changes and shake effects.
func ApplyInputBoxStyle(baseStyle lipgloss.Style, effects SANVisualEffects) lipgloss.Style {
	// Apply border color
	style := baseStyle.BorderForeground(lipgloss.Color(effects.BorderColor))

	// Note: Border shake is applied through UI shake system, not here
	// This just sets up the color

	return style
}

// ApplyTextColorEffects applies text color effects based on SAN.
// Story 7.5 AC3, AC4: Text color darkening and random color changes.
func ApplyTextColorEffects(text string, effects SANVisualEffects) string {
	if effects.TextColorDarkness == 0.0 && !effects.TextColorUnstable {
		return text
	}

	// Base style with normal color
	style := lipgloss.NewStyle()

	// Apply color darkness (darker gray shades)
	if effects.TextColorDarkness > 0.0 {
		// Use predefined dark gray colors based on darkness level
		// darkness 0.1 -> "250" (light gray)
		// darkness 0.3 -> "245" (medium gray)
		// darkness 0.6 -> "240" (dark gray)
		var grayColor string
		if effects.TextColorDarkness < 0.2 {
			grayColor = "250"
		} else if effects.TextColorDarkness < 0.4 {
			grayColor = "245"
		} else if effects.TextColorDarkness < 0.7 {
			grayColor = "240"
		} else {
			grayColor = "235"
		}
		style = style.Foreground(lipgloss.Color(grayColor))
	}

	// Apply random color instability
	if effects.TextColorUnstable {
		// Randomly shift to disturbing colors
		colors := []string{"9", "11", "13", "14", "202", "196"} // Red, yellow, magenta, cyan, orange, red
		if rand.Float64() < 0.3 {
			randomColor := colors[rand.Intn(len(colors))]
			style = style.Foreground(lipgloss.Color(randomColor))
		}
	}

	return style.Render(text)
}

// ApplyTextDistortionEffects applies text distortion effects (blur, ghost, shake).
// Story 7.5 AC4: Severe visual distortion at low SAN.
func ApplyTextDistortionEffects(text string, effects SANVisualEffects) string {
	if !effects.TextBlur && !effects.TextGhost && !effects.TextShake {
		return text
	}

	result := text

	// Ghost/echo effect: duplicate text with slight offset
	if effects.TextGhost && rand.Float64() < 0.3 {
		// Add a faded duplicate
		ghostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true)
		result = result + ghostStyle.Render(text)
	}

	// Blur effect: add space between some characters
	if effects.TextBlur && rand.Float64() < 0.2 {
		runes := []rune(result)
		var blurred []rune
		for i, r := range runes {
			blurred = append(blurred, r)
			// Randomly add space
			if i < len(runes)-1 && rand.Float64() < 0.1 {
				blurred = append(blurred, ' ')
			}
		}
		result = string(blurred)
	}

	// Text shake is applied through the UI shake system, not here

	return result
}

// CalculateInputBoxWidth calculates the actual input box width based on scale.
// Story 7.5 AC2, AC3, AC4: Input box shrinkage at different SAN levels.
func CalculateInputBoxWidth(originalWidth int, effects SANVisualEffects) int {
	scaledWidth := int(float64(originalWidth) * effects.InputBoxScale)
	if scaledWidth < 10 {
		scaledWidth = 10 // Minimum width for usability
	}
	return scaledWidth
}

// GetCursorBlinkInterval returns the cursor blink interval based on SAN effects.
// Story 7.5 AC3: Cursor blinks faster (1.5x) at SAN 20-39.
func GetCursorBlinkInterval(effects SANVisualEffects, baseInterval int) int {
	// baseInterval is in milliseconds
	adjustedInterval := float64(baseInterval) / effects.CursorBlinkSpeed
	if adjustedInterval < 50 {
		adjustedInterval = 50 // Minimum to prevent seizure-inducing flicker
	}
	return int(adjustedInterval)
}

// ShouldApplyCharDeletion determines if character deletion should occur.
// Story 7.5 AC4: Random character deletion at SAN 1-19.
func ShouldApplyCharDeletion(effects SANVisualEffects) bool {
	return effects.CharDeletionRate > 0.0
}

// ApplyCharacterDeletion randomly deletes characters from input text.
// Story 7.5 AC4: Delete 10-20% of characters when SAN 1-19.
func ApplyCharacterDeletion(text string, effects SANVisualEffects) (string, int) {
	if !ShouldApplyCharDeletion(effects) {
		return text, 0
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text, 0
	}

	var result []rune
	deletedCount := 0

	for _, r := range runes {
		// Randomly delete based on deletion rate
		if rand.Float64() < effects.CharDeletionRate {
			deletedCount++
			continue // Skip this character
		}
		result = append(result, r)
	}

	return string(result), deletedCount
}

// GetVisualEffectDescription returns a description of active visual effects.
// Useful for debugging and accessible mode.
func GetVisualEffectDescription(effects SANVisualEffects) string {
	if effects.SAN >= 80 {
		return "視覺效果：無"
	}

	var desc []string

	if effects.InputBoxScale < 1.0 {
		shrinkage := int((1.0 - effects.InputBoxScale) * 100)
		desc = append(desc, fmt.Sprintf("輸入框縮小%d%%", shrinkage))
	}

	if effects.BorderShake {
		desc = append(desc, "邊框抖動")
	}

	if effects.CursorBlinkSpeed > 1.0 {
		desc = append(desc, "游標閃爍加速")
	}

	if effects.TextColorUnstable {
		desc = append(desc, "文字顏色不穩定")
	}

	if effects.TextBlur {
		desc = append(desc, "文字模糊")
	}

	if effects.CharDeletionRate > 0.0 {
		desc = append(desc, "文字隨機消失")
	}

	if len(desc) == 0 {
		return "視覺效果：輕微干擾"
	}

	result := "視覺效果："
	for i, d := range desc {
		if i > 0 {
			result += "、"
		}
		result += d
	}

	return result
}

// InputBoxShakeOffset returns x,y offset for border shake animation.
// Story 7.5 AC2, AC4: Border shake based on shake intensity.
func InputBoxShakeOffset(effects SANVisualEffects) (int, int) {
	if !effects.BorderShake || effects.ShakeIntensity == 0 {
		return 0, 0
	}

	// Random offset within shake intensity bounds
	x := rand.Intn(effects.ShakeIntensity*2+1) - effects.ShakeIntensity
	y := rand.Intn(effects.ShakeIntensity*2+1) - effects.ShakeIntensity

	return x, y
}
