// Story 7-1 Test: AutoResolve() Implementation
package momentum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAutoResolve_AC1_LowRiskAutoResolution tests automatic resolution of low-risk actions
func TestAutoResolve_AC1_LowRiskAutoResolution(t *testing.T) {
	config := DefaultMomentumConfig()
	config.AutoResolve = true
	config.MaxAutoBeats = 5

	controller := NewMomentumController(config, nil)

	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	assert.NotNil(t, result, "Result should not be nil")
	assert.Greater(t, result.BeatsResolved, 0, "Should resolve at least 1 beat")
}

// TestAutoResolve_AC2_ResultStructure tests that AutoResolveResult contains all required fields
func TestAutoResolve_AC2_ResultStructure(t *testing.T) {
	config := DefaultMomentumConfig()
	config.AutoResolve = true
	config.MaxAutoBeats = 3

	controller := NewMomentumController(config, nil)

	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	require.NotNil(t, result)
	assert.NotNil(t, result.Narratives, "Narratives should not be nil")
	assert.GreaterOrEqual(t, result.BeatsResolved, 0, "BeatsResolved should be >= 0")
	assert.LessOrEqual(t, result.HPDelta, 0, "HPDelta should be <= 0")
	assert.LessOrEqual(t, result.SANDelta, 0, "SANDelta should be <= 0")
	assert.NotEqual(t, StopReasonNone, result.StopReason, "Should have a stop reason")
}

// TestAutoResolve_AC3_LoopUntilPauseCondition tests that loop continues until pause condition
func TestAutoResolve_AC3_LoopUntilPauseCondition(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyMedium,
		AutoResolve:  true,
		MaxAutoBeats: 10,
		PauseOnRisk:  RiskMedium,
		PauseOnPlot:  false,
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}

	controller := NewMomentumController(config, nil)

	// Low risk - should continue
	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	assert.Greater(t, result.BeatsResolved, 0, "Should resolve beats in low risk")
}

// TestAutoResolve_AC3_StopOnHighRisk tests that loop stops when high risk is encountered
func TestAutoResolve_AC3_StopOnHighRisk(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyMedium,
		AutoResolve:  true,
		MaxAutoBeats: 10,
		PauseOnRisk:  RiskMedium,
		PauseOnPlot:  false,
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}

	controller := NewMomentumController(config, nil)

	// High risk - should pause immediately
	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskHigh,
	}

	result := controller.AutoResolve(ctx)

	assert.Equal(t, 0, result.BeatsResolved, "Should not resolve beats in high risk")
	assert.Equal(t, StopReasonRiskLevel, result.StopReason, "Should stop due to risk level")
}

// TestAutoResolve_AC6_RespectMaxAutoBeats tests that MaxAutoBeats limit is respected
func TestAutoResolve_AC6_RespectMaxAutoBeats(t *testing.T) {
	maxBeats := 3
	config := &MomentumConfig{
		Frequency:    FrequencyLow,
		AutoResolve:  true,
		MaxAutoBeats: maxBeats,
		PauseOnRisk:  RiskHigh,
		PauseOnPlot:  false,
		PauseOnNPC:   false,
		PauseOnEvent: false,
	}

	controller := NewMomentumController(config, nil)

	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	assert.LessOrEqual(t, result.BeatsResolved, maxBeats, "Should not exceed MaxAutoBeats")
	assert.Equal(t, StopReasonMaxAutoBeats, result.StopReason, "Should stop at MaxAutoBeats")
}

// TestAutoResolve_AC7_RecordStopReason tests that stop reason and context are recorded
func TestAutoResolve_AC7_RecordStopReason(t *testing.T) {
	config := DefaultMomentumConfig()
	config.AutoResolve = true
	config.MaxAutoBeats = 5

	controller := NewMomentumController(config, nil)

	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	assert.NotEqual(t, StopReasonNone, result.StopReason, "Should record stop reason")
	assert.NotNil(t, result.StopContext, "Should record stop context")
}

// TestAutoResolve_DisabledAutoResolve tests behavior when AutoResolve is disabled
func TestAutoResolve_DisabledAutoResolve(t *testing.T) {
	config := DefaultMomentumConfig()
	config.AutoResolve = false

	controller := NewMomentumController(config, nil)

	ctx := &NarrativeContext{
		CurrentBeat: 1,
		RiskLevel:   RiskLow,
	}

	result := controller.AutoResolve(ctx)

	assert.Equal(t, 0, result.BeatsResolved, "Should not resolve any beats when disabled")
	assert.Equal(t, StopReasonNone, result.StopReason, "Stop reason should be None")
}

// TestAutoResolve_NilContext tests handling of nil context
func TestAutoResolve_NilContext(t *testing.T) {
	config := DefaultMomentumConfig()
	controller := NewMomentumController(config, nil)

	result := controller.AutoResolve(nil)

	assert.NotNil(t, result, "Should return result even with nil context")
	assert.Equal(t, 0, result.BeatsResolved, "Should not resolve beats with nil context")
}
