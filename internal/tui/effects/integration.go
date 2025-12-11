package effects

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// EffectManager manages all visual effects for the TUI.
// This is the main integration point for applying horror effects to the game screen.
type EffectManager struct {
	currentSAN int
	style      HorrorStyle
	transition *TransitionState

	// Effect state trackers
	flashState  *FlashState
	cursorState *CursorState

	// EventBus for SAN changes
	eventBus *EventBus

	// Accessible mode flag
	accessibleMode bool
}

// NewEffectManager creates a new effect manager.
func NewEffectManager(initialSAN int, accessibleMode bool) *EffectManager {
	// Set global accessible mode
	AccessibleMode = accessibleMode

	style := CalculateHorrorStyle(initialSAN)

	if accessibleMode {
		style = ScaleEffectIntensity(style)
	}

	em := &EffectManager{
		currentSAN:     initialSAN,
		style:          style,
		flashState:     NewFlashState(),
		cursorState:    NewCursorState(),
		eventBus:       NewEventBus(),
		accessibleMode: accessibleMode,
	}

	// Subscribe to SAN change events
	em.eventBus.Subscribe(EventSANChanged, em.handleSANChange)

	// Start event bus
	em.eventBus.Start()

	return em
}

// handleSANChange handles SAN change events.
func (em *EffectManager) handleSANChange(e Event) {
	data, ok := e.Data.(SANChangeWithStyle)
	if !ok {
		return
	}

	// Start smooth transition from current style to new style
	newStyle := data.Style
	if em.accessibleMode {
		newStyle = ScaleEffectIntensity(newStyle)
	}

	em.transition = NewTransitionState(em.style, newStyle, 500*time.Millisecond)
	em.currentSAN = data.NewSAN
}

// Update updates all effect states based on current time.
// This should be called on every BubbleTea tick.
func (em *EffectManager) Update(now time.Time) {
	// Update transition if active
	if em.transition != nil {
		em.style = em.transition.GetCurrent(now)
		if em.transition.IsComplete(now) {
			em.transition = nil // Transition complete
		}
	}

	// Update flash state
	frequency := CalculateScreenFlashFrequency(em.currentSAN)
	em.flashState.Update(now, frequency)

	// Update cursor state
	blinkRate := CalculateCursorBlinkRate(em.currentSAN)
	em.cursorState.Update(now, blinkRate)
}

// SetSAN updates the SAN value and triggers effects.
func (em *EffectManager) SetSAN(newSAN int) {
	if newSAN == em.currentSAN {
		return
	}

	oldSAN := em.currentSAN
	em.eventBus.EmitSANChange(oldSAN, newSAN)
}

// GetStyle returns the current HorrorStyle (with transitions applied).
func (em *EffectManager) GetStyle() HorrorStyle {
	return em.style
}

// ApplyNarrativeEffects applies horror effects to narrative text.
func (em *EffectManager) ApplyNarrativeEffects(text string) string {
	if em.accessibleMode {
		return ApplyAccessibleEffects(text, em.style)
	}

	return ApplyZalgo(text, em.style.TextCorruption)
}

// ApplyColorEffects applies color shift effects to a LipGloss style.
func (em *EffectManager) ApplyColorEffects(style lipgloss.Style) lipgloss.Style {
	return ApplyColorShift(style, em.style.ColorShift)
}

// ApplyBorderEffects applies UI shake effects to border content.
func (em *EffectManager) ApplyBorderEffects(content string) string {
	offset := CalculateShakeOffset(em.style.UIStability)
	return RenderShakingBorder(content, offset)
}

// ApplyInputEffects applies input corruption effects.
func (em *EffectManager) ApplyInputEffects(input string) CorruptedInput {
	return ApplyInputCorruption(input, em.style.TypingBehavior)
}

// GetInputBoxWidth returns the current input box width based on SAN.
func (em *EffectManager) GetInputBoxWidth(originalWidth int) int {
	shrinkage := CalculateInputBoxShrinkage(em.currentSAN)
	return int(float64(originalWidth) * shrinkage)
}

// IsCursorVisible returns whether the cursor should be visible.
func (em *EffectManager) IsCursorVisible() bool {
	return em.cursorState.IsVisible()
}

// GetCursorOffset returns the cursor position desync offset.
func (em *EffectManager) GetCursorOffset() int {
	return GetCursorDesyncOffset(em.currentSAN)
}

// IsFlashing returns whether the screen should be flashing.
func (em *EffectManager) IsFlashing() bool {
	return em.flashState.ShouldFlash()
}

// GetFlashEffect returns the current flash effect.
func (em *EffectManager) GetFlashEffect() FlashEffect {
	frequency := CalculateScreenFlashFrequency(em.currentSAN)
	return GetFlashEffect(em.flashState, time.Now(), frequency)
}

// GetAccessibleStateDescription returns narrative description for accessible mode.
func (em *EffectManager) GetAccessibleStateDescription() string {
	if !em.accessibleMode {
		return ""
	}

	return GetAccessibleSANStateDescription(em.currentSAN)
}

// GetStatusText returns the status text for the player's mental state.
// AC1-4 require displaying status: 焦慮, 恐慌, 崩潰
func (em *EffectManager) GetStatusText() string {
	switch {
	case em.currentSAN >= 60:
		return "" // No status text for high SAN
	case em.currentSAN >= 40:
		return "焦慮" // Anxious (AC2)
	case em.currentSAN >= 20:
		return "恐慌" // Panicked (AC3)
	default:
		return "崩潰" // Insanity (AC4)
	}
}

// GetTickInterval returns the recommended tick interval for smooth effects.
func (em *EffectManager) GetTickInterval() time.Duration {
	return GetTickInterval(em.currentSAN)
}

// Cleanup stops the event bus and cleans up resources.
func (em *EffectManager) Cleanup() {
	em.eventBus.Stop()
}

// RenderNarrative renders narrative text with all appropriate effects.
// This is a convenience method that combines text corruption and accessible mode.
func (em *EffectManager) RenderNarrative(text string, baseStyle lipgloss.Style) string {
	// Apply text corruption
	corruptedText := em.ApplyNarrativeEffects(text)

	// Apply color effects to style
	styledText := em.ApplyColorEffects(baseStyle).Render(corruptedText)

	// Apply border shake if needed
	if em.style.UIStability > 0 {
		styledText = em.ApplyBorderEffects(styledText)
	}

	return styledText
}

// RenderOptions renders game options with horror effects.
// For Story 6-2 (Hallucination Options), this will be extended.
func (em *EffectManager) RenderOptions(options []string, baseStyle lipgloss.Style) []string {
	rendered := make([]string, len(options))

	for i, option := range options {
		// For now, just apply color effects
		// Story 6-2 will add hallucination logic
		styled := em.ApplyColorEffects(baseStyle).Render(option)
		rendered[i] = styled
	}

	return rendered
}

// RenderInputBox renders the input box with all effects.
func (em *EffectManager) RenderInputBox(content string, width int, baseStyle lipgloss.Style) string {
	// Calculate shrunk width
	shrunkWidth := em.GetInputBoxWidth(width)

	// Truncate content if needed
	truncated := TruncateToShrunkWidth(content, shrunkWidth, float64(shrunkWidth)/float64(width))

	// Apply color effects
	styledContent := em.ApplyColorEffects(baseStyle).Width(shrunkWidth).Render(truncated)

	// Apply border shake
	if em.style.UIStability > 0 {
		styledContent = em.ApplyBorderEffects(styledContent)
	}

	return styledContent
}

// ProcessInput processes user input and applies corruption effects.
// Returns the corrupted input and visual feedback.
func (em *EffectManager) ProcessInput(input string) (processed string, feedback InputVisualFeedback) {
	corrupted := em.ApplyInputEffects(input)

	// Sanitize for actual game use
	processed = SanitizeCorruptedInput(corrupted.Corrupted)

	// Generate visual feedback
	feedback = CalculateInputVisualFeedback(corrupted)

	return processed, feedback
}

// ShouldShowAccessibleDescription returns whether to show accessible narrative descriptions.
func (em *EffectManager) ShouldShowAccessibleDescription() bool {
	return em.accessibleMode && em.currentSAN < 80
}

// GetCurrentSAN returns the current SAN value.
func (em *EffectManager) GetCurrentSAN() int {
	return em.currentSAN
}

// SetAccessibleMode enables or disables accessible mode.
func (em *EffectManager) SetAccessibleMode(enabled bool) {
	em.accessibleMode = enabled

	// Set global accessible mode
	AccessibleMode = enabled

	// Recalculate style with new accessible mode setting
	newStyle := CalculateHorrorStyle(em.currentSAN)
	if enabled {
		newStyle = ScaleEffectIntensity(newStyle)
	}

	// Start smooth transition
	em.transition = NewTransitionState(em.style, newStyle, 500*time.Millisecond)
}

// IsAccessibleMode returns whether accessible mode is enabled.
func (em *EffectManager) IsAccessibleMode() bool {
	return em.accessibleMode
}
