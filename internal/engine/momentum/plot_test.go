package momentum

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// TestNewPlotDetector tests the PlotDetector constructor.
func TestNewPlotDetector(t *testing.T) {
	pd := NewPlotDetector()
	if pd == nil {
		t.Fatal("NewPlotDetector() returned nil")
	}
	if pd.milestonesReached == nil {
		t.Error("milestonesReached map should be initialized")
	}
}

// TestGetCurrentAct tests act determination based on beat number.
func TestGetCurrentAct(t *testing.T) {
	pd := NewPlotDetector()

	tests := []struct {
		beat int
		act  int
	}{
		{0, 1}, {15, 1}, {30, 1},
		{31, 2}, {45, 2}, {60, 2},
		{61, 3}, {80, 3}, {100, 3},
		{101, 0},
	}

	for _, tt := range tests {
		if act := pd.getCurrentAct(tt.beat); act != tt.act {
			t.Errorf("getCurrentAct(%d) = %d, want %d", tt.beat, act, tt.act)
		}
	}
}

// TestIsNearActEnd tests detection of approaching act transitions.
func TestIsNearActEnd(t *testing.T) {
	pd := NewPlotDetector()

	tests := []struct {
		beat     int
		expected bool
	}{
		{20, false}, {27, true}, {30, true}, {31, false},
		{45, false}, {57, true}, {60, true}, {61, false},
		{70, false}, {97, true}, {100, true},
	}

	for _, tt := range tests {
		if result := pd.isNearActEnd(tt.beat); result != tt.expected {
			t.Errorf("isNearActEnd(%d) = %v, want %v", tt.beat, result, tt.expected)
		}
	}
}

// TestCalculateSeedProgress tests seed revelation progress calculation.
func TestCalculateSeedProgress(t *testing.T) {
	pd := NewPlotDetector()

	tests := []struct {
		name     string
		seeds    []*seed.GlobalSeed
		expected float64
		delta    float64
	}{
		{"No seeds", []*seed.GlobalSeed{}, 0.0, 0.1},
		{"One seed tier 1", []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 1}}, 0.0, 0.1},
		{"One seed tier 2", []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}}, 33.33, 1.0},
		{"One seed tier 3", []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 3}}, 66.67, 1.0},
		{"One seed all revealed", []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 4}}, 100.0, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.GlobalSeeds = tt.seeds
			progress := pd.calculateSeedProgress(state)
			if progress < tt.expected-tt.delta || progress > tt.expected+tt.delta {
				t.Errorf("calculateSeedProgress() = %.2f, want %.2f±%.2f", progress, tt.expected, tt.delta)
			}
		})
	}
}

// TestEvaluateSeedProgress tests seed milestone detection.
func TestEvaluateSeedProgress(t *testing.T) {
	pd := NewPlotDetector()
	state := engine.NewGameStateV2()
	state.GlobalSeeds = []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}}

	isMilestone, milestoneType := pd.evaluateSeedProgress(state)
	if !isMilestone || milestoneType != PlotPointSeedMilestone30 {
		t.Errorf("Expected 30%% milestone, got %v, %s", isMilestone, milestoneType)
	}

	// Second call should not trigger again
	isMilestone, _ = pd.evaluateSeedProgress(state)
	if isMilestone {
		t.Error("Milestone should not trigger twice")
	}
}

// TestIs PlotPoint tests the main plot point detection logic.
func TestIsPlotPoint(t *testing.T) {
	tests := []struct {
		name      string
		beat      int
		seeds     []*seed.GlobalSeed
		wantPlot  bool
		wantType  string
	}{
		{"No plot point", 15, nil, false, ""},
		{"Act transition", 30, nil, true, PlotPointActTransition},
		{"Seed milestone 30%", 20, []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}}, true, PlotPointSeedMilestone30},
		{"Seed milestone 60%", 40, []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 3}}, true, PlotPointSeedMilestone60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := NewPlotDetector()
			state := engine.NewGameStateV2()
			// Manually advance beats to set CurrentBeat correctly
			for i := 0; i < tt.beat; i++ {
				state.IncrementBeat()
			}
			if tt.seeds != nil {
				state.GlobalSeeds = tt.seeds
			}

			isPlot, plotType := pd.IsPlotPoint(state)
			if isPlot != tt.wantPlot {
				t.Errorf("IsPlotPoint() isPlot = %v, want %v", isPlot, tt.wantPlot)
			}
			if plotType != tt.wantType {
				t.Errorf("IsPlotPoint() type = %q, want %q", plotType, tt.wantType)
			}
		})
	}
}

// TestNilState tests nil safety.
func TestNilState(t *testing.T) {
	pd := NewPlotDetector()
	isPlot, plotType := pd.IsPlotPoint(nil)
	if isPlot || plotType != "" {
		t.Error("IsPlotPoint(nil) should return false, empty string")
	}
}

// TestReset tests the Reset method.
func TestReset(t *testing.T) {
	pd := NewPlotDetector()
	pd.milestonesReached["seed_30"] = true
	pd.Reset()
	if len(pd.milestonesReached) != 0 {
		t.Error("Reset() should clear milestones")
	}
}

// TestHasReachedMilestone tests the milestone query method.
func TestHasReachedMilestone(t *testing.T) {
	pd := NewPlotDetector()
	if pd.HasReachedMilestone("seed_30") {
		t.Error("seed_30 should not be reached initially")
	}
	pd.milestonesReached["seed_30"] = true
	if !pd.HasReachedMilestone("seed_30") {
		t.Error("seed_30 should be reached after marking")
	}
}

// TestDetectAdapter tests the Detect adapter method.
func TestDetectAdapter(t *testing.T) {
	pd := NewPlotDetector()

	if pd.Detect(nil) {
		t.Error("Detect(nil) should return false")
	}

	ctx := &NarrativeContext{IsPlotPoint: true}
	if !pd.Detect(ctx) {
		t.Error("Detect should return true when context.IsPlotPoint is true")
	}

	ctx.IsPlotPoint = false
	if pd.Detect(ctx) {
		t.Error("Detect should return false when context.IsPlotPoint is false")
	}
}

// TestEvaluateNPCDeath tests NPC death detection.
func TestEvaluateNPCDeath(t *testing.T) {
	pd := NewPlotDetector()

	// No NPCManager (state created by NewGameStateV2 has initialized NPCManager)
	state := engine.NewGameStateV2()
	isDeath, _ := pd.evaluateNPCDeath(state, 10)
	if isDeath {
		t.Error("evaluateNPCDeath with no NPCManager should return false")
	}
}

// TestEvaluateRuleViolation tests first-time rule violation detection.
func TestEvaluateRuleViolation(t *testing.T) {
	pd := NewPlotDetector()

	// No rules
	state := engine.NewGameStateV2()
	if pd.evaluateRuleViolation(state) {
		t.Error("evaluateRuleViolation with no rules should return false")
	}

	// Rules with warnings
	state.RuleWarnings = map[string]int{"rule1": 1, "rule2": 2}
	if pd.evaluateRuleViolation(state) {
		t.Error("evaluateRuleViolation with warned rules should return false")
	}
}
