package effects

import "fmt"

// AccessibleMode controls whether effects should be reduced for accessibility.
var AccessibleMode = false

// ApplyAccessibleEffects returns either the full horror effect or an accessible alternative.
// When AccessibleMode is true, it replaces visual distortions with text descriptions.
func ApplyAccessibleEffects(originalText string, style HorrorStyle) string {
	if !AccessibleMode {
		// Full effects
		return ApplyZalgo(originalText, style.TextCorruption)
	}

	// Accessible mode: use text descriptions instead of visual corruption
	if style.TextCorruption >= 0.9 {
		return fmt.Sprintf("[文字嚴重混亂] %s", originalText)
	} else if style.TextCorruption >= 0.6 {
		return fmt.Sprintf("[文字混亂] %s", originalText)
	} else if style.TextCorruption >= 0.3 {
		return fmt.Sprintf("[文字微亂] %s", originalText)
	} else if style.TextCorruption >= 0.1 {
		return originalText // Minimal effect, no description needed
	}

	return originalText
}

// GetAccessibleSANStateDescription returns a text description of the player's mental state.
// This is used in accessible mode to convey horror through narrative instead of visuals.
func GetAccessibleSANStateDescription(san int) string {
	switch {
	case san >= 80:
		return ""
	case san >= 60:
		return "你的視線偶爾有些模糊..."
	case san >= 40:
		return "周圍的一切開始變得不太真實..."
	case san >= 20:
		return "你的思緒開始混亂，難以集中注意力..."
	case san >= 1:
		return "現實與幻覺已經無法分辨..."
	default:
		return "一切都崩塌了..."
	}
}

// ScaleEffectIntensity scales effect intensity based on accessible mode.
// In accessible mode, all effects are reduced by 50%.
func ScaleEffectIntensity(style HorrorStyle) HorrorStyle {
	if !AccessibleMode {
		return style
	}

	// Reduce all effects by 50%
	return HorrorStyle{
		TextCorruption:    style.TextCorruption * 0.5,
		TypingBehavior:    style.TypingBehavior * 0.5,
		ColorShift:        style.ColorShift / 2,
		UIStability:       style.UIStability / 2,
		OptionReliability: style.OptionReliability,
	}
}
