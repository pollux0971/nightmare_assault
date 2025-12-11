package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewDeathModel(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	deathInfo.Narrative = "你死了..."

	model := NewDeathModel(deathInfo)

	if model.deathInfo != deathInfo {
		t.Error("DeathInfo not set correctly")
	}
	if model.narrative != "你死了..." {
		t.Errorf("Narrative = %s, want '你死了...'", model.narrative)
	}
	if model.selected != 0 {
		t.Errorf("selected = %d, want 0", model.selected)
	}
	if !model.showTransition {
		t.Error("showTransition should be true initially")
	}
}

func TestDeathModelInit(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a tick command")
	}
}

func TestDeathModelTransitionTick(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24

	// Initial state
	if !model.showTransition {
		t.Error("Should show transition initially")
	}

	// Advance transition
	for i := 0; i < 60; i++ {
		m, _ := model.Update(TransitionTickMsg{})
		model = m.(DeathModel)
	}

	// Transition should be complete
	if model.showTransition {
		t.Error("Transition should be complete after 60 ticks")
	}
}

func TestDeathModelKeyNavigation(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.showTransition = false // Skip transition

	// Initial selection is 0 (debrief)
	if model.selected != 0 {
		t.Errorf("Initial selected = %d, want 0", model.selected)
	}

	// Press down
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DeathModel)
	if model.selected != 1 {
		t.Errorf("After down, selected = %d, want 1", model.selected)
	}

	// Press down again (should stay at 1)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DeathModel)
	if model.selected != 1 {
		t.Errorf("At bottom, selected = %d, want 1", model.selected)
	}

	// Press up
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = m.(DeathModel)
	if model.selected != 0 {
		t.Errorf("After up, selected = %d, want 0", model.selected)
	}
}

func TestDeathModelSelectDebrief(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.showTransition = false
	model.selected = 0 // Debrief

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	// Execute the command
	msg := cmd()
	selectMsg, ok := msg.(DeathSelectMsg)
	if !ok {
		t.Fatal("Command should return DeathSelectMsg")
	}
	if selectMsg.Action != "debrief" {
		t.Errorf("Action = %s, want 'debrief'", selectMsg.Action)
	}
}

func TestDeathModelSelectMenu(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.showTransition = false
	model.selected = 1 // Menu

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DeathSelectMsg)
	if !ok {
		t.Fatal("Command should return DeathSelectMsg")
	}
	if selectMsg.Action != "menu" {
		t.Errorf("Action = %s, want 'menu'", selectMsg.Action)
	}
}

func TestDeathModelBlockInputDuringTransition(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	// showTransition is true by default

	// Try to change selection during transition
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(DeathModel)

	// Selection should not change
	if model.selected != 0 {
		t.Error("Selection should be blocked during transition")
	}
}

func TestDeathModelEscapeToMenu(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.showTransition = false

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Escape should return a command")
	}

	msg := cmd()
	selectMsg, ok := msg.(DeathSelectMsg)
	if !ok {
		t.Fatal("Command should return DeathSelectMsg")
	}
	if selectMsg.Action != "menu" {
		t.Errorf("Escape action = %s, want 'menu'", selectMsg.Action)
	}
}

func TestDeathModelViewTransition(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = true
	model.transitionTick = 30 // Mid-transition

	view := model.View()

	// Should show transition text
	if !strings.Contains(view, "你感覺到") {
		t.Error("Mid-transition should show '你感覺到' text")
	}
}

func TestDeathModelViewNormalDeath(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	deathInfo.Narrative = "你倒下了..."
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false

	view := model.View()

	// Should contain death type
	if !strings.Contains(view, "體力耗盡") {
		t.Error("View should show death type '體力耗盡'")
	}

	// Should contain narrative
	if !strings.Contains(view, "你倒下了") {
		t.Error("View should contain narrative")
	}

	// Should contain options
	if !strings.Contains(view, "查看覆盤") {
		t.Error("View should contain '查看覆盤' option")
	}
	if !strings.Contains(view, "返回主選單") {
		t.Error("View should contain '返回主選單' option")
	}
}

func TestDeathModelViewInsanityDeath(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeSAN)
	deathInfo.Narrative = "理智崩潰..."
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false

	view := model.View()

	// Should contain insanity death type
	if !strings.Contains(view, "理智崩潰") {
		t.Error("Insanity view should show '理智崩潰'")
	}
}

func TestDeathModelSetNarrative(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)

	model.SetNarrative("新的敘事...")

	if model.narrative != "新的敘事..." {
		t.Errorf("Narrative = %s, want '新的敘事...'", model.narrative)
	}
}

func TestDeathModelSetSize(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)

	model.SetSize(120, 40)

	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 40 {
		t.Errorf("height = %d, want 40", model.height)
	}
}

func TestDeathModelIsTransitioning(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)

	if !model.IsTransitioning() {
		t.Error("Should be transitioning initially")
	}

	model.showTransition = false
	if model.IsTransitioning() {
		t.Error("Should not be transitioning after completion")
	}
}

func TestDeathModelWindowSizeMsg(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)

	m, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = m.(DeathModel)

	if model.width != 100 {
		t.Errorf("width = %d, want 100", model.width)
	}
	if model.height != 50 {
		t.Errorf("height = %d, want 50", model.height)
	}
}

func TestGlitchText(t *testing.T) {
	original := "測試文字"

	// Run multiple times to account for randomness
	changedCount := 0
	for i := 0; i < 100; i++ {
		glitched := glitchText(original)
		if glitched != original {
			changedCount++
		}
	}

	// Should have some glitches (30% chance per character)
	if changedCount == 0 {
		t.Error("glitchText should produce some changes")
	}
	// But not all the time (very unlikely to change every single run)
	if changedCount == 100 {
		t.Log("Note: All runs produced changes, which is statistically possible but rare")
	}
}

func TestPartialGlitch(t *testing.T) {
	original := "這是一個測試字串用於測試部分亂碼功能"

	// Test with 10% corruption
	glitched := partialGlitch(original, 0.1)

	// Should still be same length
	if len([]rune(glitched)) != len([]rune(original)) {
		t.Error("Partial glitch should preserve string length")
	}
}

func TestPartialGlitchPreservesSpaces(t *testing.T) {
	original := "a b c"

	// Run multiple times
	for i := 0; i < 50; i++ {
		glitched := partialGlitch(original, 0.5)
		// Spaces at positions 1 and 3 should be preserved
		runes := []rune(glitched)
		if runes[1] != ' ' || runes[3] != ' ' {
			t.Error("Spaces should not be corrupted")
			break
		}
	}
}

func TestRgbToHex(t *testing.T) {
	tests := []struct {
		r, g, b  int
		expected string
	}{
		{255, 255, 255, "#FFFFFF"},
		{0, 0, 0, "#000000"},
		{139, 0, 0, "#8B0000"}, // Dark red
		{0, 255, 0, "#00FF00"}, // Green
	}

	for _, tt := range tests {
		result := rgbToHex(tt.r, tt.g, tt.b)
		if result != tt.expected {
			t.Errorf("rgbToHex(%d,%d,%d) = %s, want %s", tt.r, tt.g, tt.b, result, tt.expected)
		}
	}
}

func TestDeathModelViewRuleDeath(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeRule)
	deathInfo.TriggeringRuleID = "basement-rule"
	deathInfo.Narrative = "你違反了規則..."
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false

	view := model.View()

	// Should contain rule death type
	if !strings.Contains(view, "違反潛規則") {
		t.Error("Rule death view should show '違反潛規則'")
	}
}

func TestDeathModelInsanityGlitchAnimation(t *testing.T) {
	deathInfo := game.NewDeathInfo(game.DeathTypeSAN)
	deathInfo.Narrative = "理智崩潰..."
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false

	// Simulate glitch animation tick
	m, cmd := model.Update(TransitionTickMsg{})
	model = m.(DeathModel)

	// Should return another tick command for continuous animation
	if cmd == nil {
		t.Error("Insanity death should continue tick animation for glitch effects")
	}
}

// Benchmark transition rendering
func BenchmarkTransitionRender(b *testing.B) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = true
	model.transitionTick = 30

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// Benchmark normal death rendering
func BenchmarkNormalDeathRender(b *testing.B) {
	deathInfo := game.NewDeathInfo(game.DeathTypeHP)
	deathInfo.Narrative = strings.Repeat("敘事文字", 50)
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// Benchmark insanity death rendering with glitch
func BenchmarkInsanityDeathRender(b *testing.B) {
	deathInfo := game.NewDeathInfo(game.DeathTypeSAN)
	deathInfo.Narrative = strings.Repeat("瘋狂文字", 50)
	model := NewDeathModel(deathInfo)
	model.width = 80
	model.height = 24
	model.showTransition = false
	model.glitchOffset = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func init() {
	// Seed random for deterministic tests where needed
	// (Not strictly necessary since we're testing behavior, not exact values)
	_ = time.Now()
}
