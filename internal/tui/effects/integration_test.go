package effects

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestNewEffectManager(t *testing.T) {
	em := NewEffectManager(80, false)
	defer em.Cleanup()

	if em == nil {
		t.Fatal("Expected EffectManager to be created")
	}

	if em.currentSAN != 80 {
		t.Errorf("Expected SAN=80, got %d", em.currentSAN)
	}

	if em.accessibleMode {
		t.Error("Expected accessible mode to be disabled")
	}

	// Reset global after test
	AccessibleMode = false
}

func TestNewEffectManager_AccessibleMode(t *testing.T) {
	em := NewEffectManager(50, true) // Use lower SAN to test scaling
	defer em.Cleanup()
	defer func() { AccessibleMode = false }()

	if !em.accessibleMode {
		t.Error("Expected accessible mode to be enabled")
	}

	// Style should be scaled for accessible mode
	normalStyle := CalculateHorrorStyle(50)
	if em.style.TextCorruption >= normalStyle.TextCorruption {
		t.Errorf("Expected accessible mode to reduce effect intensity: got %.2f, want < %.2f",
			em.style.TextCorruption, normalStyle.TextCorruption)
	}
}

func TestEffectManager_SetSAN(t *testing.T) {
	em := NewEffectManager(80, false)
	defer em.Cleanup()

	em.SetSAN(50)

	// Wait for event processing
	time.Sleep(150 * time.Millisecond)

	if em.currentSAN != 50 {
		t.Errorf("Expected SAN=50, got %d", em.currentSAN)
	}

	// Should have started a transition
	if em.transition == nil {
		t.Error("Expected transition to be started")
	}
}

func TestEffectManager_Update(t *testing.T) {
	em := NewEffectManager(80, false)
	defer em.Cleanup()

	now := time.Now()
	em.Update(now)

	// Should not crash
	// Flash and cursor states should be updated
}

func TestEffectManager_GetStyle(t *testing.T) {
	em := NewEffectManager(50, false)
	defer em.Cleanup()

	style := em.GetStyle()

	expectedStyle := CalculateHorrorStyle(50)
	if style.TextCorruption != expectedStyle.TextCorruption {
		t.Errorf("Expected style for SAN=50")
	}
}

func TestEffectManager_ApplyNarrativeEffects(t *testing.T) {
	em := NewEffectManager(20, false)
	defer em.Cleanup()

	text := "測試文字"
	result := em.ApplyNarrativeEffects(text)

	// With low SAN, should apply Zalgo
	if len(result) <= len(text) {
		t.Error("Expected Zalgo to add combining characters")
	}
}

func TestEffectManager_ApplyNarrativeEffects_Accessible(t *testing.T) {
	em := NewEffectManager(10, true) // Very low SAN to ensure visible corruption
	defer em.Cleanup()

	text := "測試文字"
	result := em.ApplyNarrativeEffects(text)

	// In accessible mode, should use text descriptions instead of Zalgo
	// SAN=10 -> TextCorruption=0.9, scaled to 0.45 in accessible mode
	// This should show "[文字混亂]" (>= 0.3 threshold)
	if !contains(result, "[") && !contains(result, "]") {
		t.Errorf("Expected accessible mode to use text descriptions, got %q", result)
	}
}

func TestEffectManager_ApplyColorEffects(t *testing.T) {
	em := NewEffectManager(20, false)
	defer em.Cleanup()

	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	result := em.ApplyColorEffects(baseStyle)

	// Color should be shifted
	if result.GetForeground() == baseStyle.GetForeground() {
		t.Error("Expected color to be shifted with low SAN")
	}
}

func TestEffectManager_GetInputBoxWidth(t *testing.T) {
	tests := []struct {
		san          int
		originalWidth int
		expectShrink  bool
	}{
		{80, 100, false}, // High SAN - no shrink
		{50, 100, true},  // Medium SAN - should shrink
		{20, 100, true},  // Low SAN - should shrink more
	}

	for _, tt := range tests {
		em := NewEffectManager(tt.san, false)
		result := em.GetInputBoxWidth(tt.originalWidth)
		em.Cleanup()

		if tt.expectShrink {
			if result >= tt.originalWidth {
				t.Errorf("SAN=%d: Expected input box to shrink, got %d (original %d)",
					tt.san, result, tt.originalWidth)
			}
		} else {
			if result != tt.originalWidth {
				t.Errorf("SAN=%d: Expected no shrink, got %d (original %d)",
					tt.san, result, tt.originalWidth)
			}
		}
	}
}

func TestEffectManager_GetStatusText(t *testing.T) {
	tests := []struct {
		san      int
		expected string
	}{
		{100, ""},
		{80, ""},
		{60, ""},
		{59, "焦慮"},
		{50, "焦慮"},
		{40, "焦慮"},
		{39, "恐慌"},
		{30, "恐慌"},
		{20, "恐慌"},
		{19, "崩潰"},
		{10, "崩潰"},
		{1, "崩潰"},
	}

	for _, tt := range tests {
		em := NewEffectManager(tt.san, false)
		result := em.GetStatusText()
		em.Cleanup()

		if result != tt.expected {
			t.Errorf("SAN=%d: GetStatusText() = %q, want %q", tt.san, result, tt.expected)
		}
	}
}

func TestEffectManager_GetAccessibleStateDescription(t *testing.T) {
	// Without accessible mode
	em := NewEffectManager(50, false)
	result := em.GetAccessibleStateDescription()
	em.Cleanup()

	if result != "" {
		t.Error("Expected no description when accessible mode is disabled")
	}

	// With accessible mode
	em = NewEffectManager(50, true)
	result = em.GetAccessibleStateDescription()
	em.Cleanup()

	if result == "" {
		t.Error("Expected description when accessible mode is enabled and SAN < 80")
	}
}

func TestEffectManager_ProcessInput(t *testing.T) {
	em := NewEffectManager(20, false) // Low SAN for corruption
	defer em.Cleanup()

	input := "test input with many characters"
	processed, feedback := em.ProcessInput(input)

	// Processed should be sanitized (clean)
	if len(processed) == 0 {
		t.Error("Expected processed input to not be empty")
	}

	// Feedback may or may not show warning (probabilistic)
	_ = feedback
}

func TestEffectManager_SetAccessibleMode(t *testing.T) {
	em := NewEffectManager(50, false)
	defer em.Cleanup()

	if em.IsAccessibleMode() {
		t.Error("Expected accessible mode to be disabled initially")
	}

	em.SetAccessibleMode(true)

	if !em.IsAccessibleMode() {
		t.Error("Expected accessible mode to be enabled after SetAccessibleMode(true)")
	}

	// Should have started a transition
	if em.transition == nil {
		t.Error("Expected transition when changing accessible mode")
	}
}

func TestEffectManager_RenderNarrative(t *testing.T) {
	em := NewEffectManager(50, false)
	defer em.Cleanup()

	baseStyle := lipgloss.NewStyle()
	text := "Test narrative"

	result := em.RenderNarrative(text, baseStyle)

	// Should return rendered string (non-empty)
	if result == "" {
		t.Error("Expected rendered narrative to not be empty")
	}
}

func TestEffectManager_RenderOptions(t *testing.T) {
	em := NewEffectManager(50, false)
	defer em.Cleanup()

	options := []string{"Option 1", "Option 2", "Option 3"}
	baseStyle := lipgloss.NewStyle()

	result := em.RenderOptions(options, baseStyle)

	if len(result) != len(options) {
		t.Errorf("Expected %d options, got %d", len(options), len(result))
	}

	for i, rendered := range result {
		if rendered == "" {
			t.Errorf("Option %d rendered as empty", i)
		}
	}
}

func TestEffectManager_RenderInputBox(t *testing.T) {
	em := NewEffectManager(50, false)
	defer em.Cleanup()

	content := "User input"
	width := 50
	baseStyle := lipgloss.NewStyle()

	result := em.RenderInputBox(content, width, baseStyle)

	// Should return rendered string
	if result == "" {
		t.Error("Expected rendered input box to not be empty")
	}
}

func TestEffectManager_Cleanup(t *testing.T) {
	em := NewEffectManager(80, false)

	if !em.eventBus.running {
		t.Error("Expected event bus to be running")
	}

	em.Cleanup()

	if em.eventBus.running {
		t.Error("Expected event bus to be stopped after Cleanup()")
	}
}

func TestEffectManager_GetTickInterval(t *testing.T) {
	tests := []int{100, 80, 60, 40, 20, 10}

	for _, san := range tests {
		em := NewEffectManager(san, false)
		interval := em.GetTickInterval()
		em.Cleanup()

		if interval < 16*time.Millisecond {
			t.Errorf("SAN=%d: interval too fast: %v", san, interval)
		}
		if interval > 100*time.Millisecond {
			t.Errorf("SAN=%d: interval too slow: %v", san, interval)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
