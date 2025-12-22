// Story 7-6 Test: MomentumIndicator UI Component
package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMomentumIndicator tests indicator creation
func TestNewMomentumIndicator(t *testing.T) {
	indicator := NewMomentumIndicator()

	assert.NotNil(t, indicator, "Indicator should not be nil")
	assert.Equal(t, MomentumIdle, indicator.state, "Initial state should be Idle")
	assert.True(t, indicator.enabled, "Indicator should be enabled by default")
}

// TestMomentumIndicator_AC1_DisplayCurrentState tests that indicator displays current momentum state
func TestMomentumIndicator_AC1_DisplayCurrentState(t *testing.T) {
	indicator := NewMomentumIndicator()

	tests := []struct {
		name     string
		state    MomentumState
		contains string
	}{
		{"Idle state", MomentumIdle, "等待輸入"},
		{"Auto resolving", MomentumAutoResolving, "自動演繹中"},
		{"Paused", MomentumPaused, "已暫停"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator.SetState(tt.state)
			view := indicator.View()
			assert.Contains(t, view, tt.contains, "View should contain state text")
		})
	}
}

// TestMomentumIndicator_AC2_ShowAutoResolvingProgress tests auto-resolving progress display
func TestMomentumIndicator_AC2_ShowAutoResolvingProgress(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetState(MomentumAutoResolving)
	indicator.SetProgress(3, 5)

	view := indicator.View()

	assert.Contains(t, view, "自動演繹中", "Should show auto-resolving text")
	assert.Contains(t, view, "3/5", "Should show progress 3/5")
	assert.Contains(t, view, "回合", "Should show '回合' (beats)")
}

// TestMomentumIndicator_AC3_ShowWaitingForInput tests waiting for input display
func TestMomentumIndicator_AC3_ShowWaitingForInput(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetState(MomentumIdle)

	view := indicator.View()

	assert.Contains(t, view, "等待輸入", "Should show '等待輸入' when idle")
}

// TestMomentumIndicator_SetProgress tests progress setting
func TestMomentumIndicator_SetProgress(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetProgress(7, 10)

	assert.Equal(t, 7, indicator.currentBeat, "Current beat should be 7")
	assert.Equal(t, 10, indicator.maxAutoBeats, "Max auto beats should be 10")
}

// TestMomentumIndicator_SetPauseReason tests pause reason display
func TestMomentumIndicator_SetPauseReason(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetState(MomentumPaused)
	indicator.SetPauseReason("高風險")

	view := indicator.View()

	assert.Contains(t, view, "已暫停", "Should show paused")
	assert.Contains(t, view, "高風險", "Should show pause reason")
}

// TestMomentumIndicator_SetEnabled tests enable/disable functionality
func TestMomentumIndicator_SetEnabled(t *testing.T) {
	indicator := NewMomentumIndicator()

	// Test enabled
	indicator.SetEnabled(true)
	indicator.SetState(MomentumIdle)
	view := indicator.View()
	assert.NotEmpty(t, view, "Should show view when enabled")

	// Test disabled
	indicator.SetEnabled(false)
	view = indicator.View()
	assert.Empty(t, view, "Should not show view when disabled")
}

// TestMomentumIndicator_Reset tests reset functionality
func TestMomentumIndicator_Reset(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetState(MomentumAutoResolving)
	indicator.SetProgress(5, 10)
	indicator.SetPauseReason("Test")

	indicator.Reset()

	assert.Equal(t, MomentumIdle, indicator.state, "State should reset to Idle")
	assert.Equal(t, 0, indicator.currentBeat, "Current beat should reset to 0")
	assert.Empty(t, indicator.pauseReason, "Pause reason should be cleared")
}

// TestMomentumIndicator_GetHeight tests height method
func TestMomentumIndicator_GetHeight(t *testing.T) {
	indicator := NewMomentumIndicator()
	height := indicator.GetHeight()
	assert.Equal(t, 1, height, "Height should always be 1")
}

// TestMomentumState_String tests MomentumState string representation
func TestMomentumState_String(t *testing.T) {
	tests := []struct {
		state    MomentumState
		expected string
	}{
		{MomentumIdle, "等待輸入"},
		{MomentumAutoResolving, "自動演繹中"},
		{MomentumPaused, "已暫停"},
		{MomentumState(999), "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			assert.Equal(t, tt.expected, result, "String mismatch")
		})
	}
}

// TestMomentumIndicator_ThreadSafety tests concurrent access
func TestMomentumIndicator_ThreadSafety(t *testing.T) {
	indicator := NewMomentumIndicator()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			indicator.SetState(MomentumAutoResolving)
			indicator.SetProgress(id, 10)
			_ = indicator.View()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not crash
	assert.NotNil(t, indicator, "Indicator should survive concurrent access")
}

// TestMomentumIndicator_ViewFormatting tests view output formatting
func TestMomentumIndicator_ViewFormatting(t *testing.T) {
	indicator := NewMomentumIndicator()
	indicator.SetState(MomentumAutoResolving)
	indicator.SetProgress(2, 5)

	view := indicator.View()

	// Should contain brackets
	assert.True(t, strings.HasPrefix(view, "[") || strings.Contains(view, "["), "Should have opening bracket")
	assert.True(t, strings.HasSuffix(view, "]") || strings.Contains(view, "]"), "Should have closing bracket")
}
