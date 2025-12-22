package momentum

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

// TestNewContextBuilder tests the ContextBuilder constructor.
// Story 6.6 AC5: buildContext() 基礎結構
func TestNewContextBuilder(t *testing.T) {
	builder := NewContextBuilder()
	if builder == nil {
		t.Fatal("NewContextBuilder() returned nil")
	}
	if builder.riskEvaluator == nil {
		t.Error("riskEvaluator should be initialized")
	}
	if builder.plotDetector == nil {
		t.Error("plotDetector should be initialized")
	}
}

// TestBuildContext_NilState tests nil safety.
// Story 6.6 Dev Notes: 測試基本上下文建構
func TestBuildContext_NilState(t *testing.T) {
	builder := NewContextBuilder()
	ctx := builder.BuildContext(nil, 0)

	if ctx == nil {
		t.Fatal("BuildContext(nil) should return non-nil context")
	}

	// Verify safe defaults
	if ctx.CurrentBeat != 0 {
		t.Errorf("expected CurrentBeat 0, got %d", ctx.CurrentBeat)
	}
	if ctx.CurrentScene != "" {
		t.Errorf("expected empty CurrentScene, got %s", ctx.CurrentScene)
	}
	if ctx.RiskLevel != RiskNone {
		t.Errorf("expected RiskNone, got %v", ctx.RiskLevel)
	}
	if ctx.IsPlotPoint {
		t.Error("expected IsPlotPoint false")
	}
	if ctx.NPCInitiatesConversation {
		t.Error("expected NPCInitiatesConversation false")
	}
	if ctx.PendingEvents == nil {
		t.Error("PendingEvents should not be nil")
	}
	if ctx.RecentChoices == nil {
		t.Error("RecentChoices should not be nil")
	}
}

// TestBuildContext_BasicFields tests basic field extraction.
// Story 6.6 AC1: CurrentBeat/CurrentScene 欄位
// Story 6.6 Dev Notes: 測試基本資訊提取
func TestBuildContext_BasicFields(t *testing.T) {
	state := engine.NewGameStateV2()
	// Advance to beat 15
	for i := 0; i < 15; i++ {
		state.IncrementBeat()
	}
	state.CurrentScene = "hospital_corridor"

	builder := NewContextBuilder()
	ctx := builder.BuildContext(state, 0)

	if ctx.CurrentBeat != 15 {
		t.Errorf("expected CurrentBeat 15, got %d", ctx.CurrentBeat)
	}
	if ctx.CurrentScene != "hospital_corridor" {
		t.Errorf("expected CurrentScene 'hospital_corridor', got %s", ctx.CurrentScene)
	}
}

// TestBuildContext_RiskFactors tests risk factor extraction.
// Story 6.6 AC1: RiskFactors 欄位
// Story 6.6 Dev Notes: 測試風險評估整合
func TestBuildContext_RiskFactors(t *testing.T) {
	tests := []struct {
		name           string
		hp             int
		san            int
		tensionValue   int
		ruleWarnings   map[string]int
		expectedFactors []string
	}{
		{
			name:            "all healthy",
			hp:              100,
			san:             100,
			tensionValue:    0,
			ruleWarnings:    nil,
			expectedFactors: []string{},
		},
		{
			name:            "low hp",
			hp:              20,
			san:             100,
			tensionValue:    0,
			expectedFactors: []string{"low_hp"},
		},
		{
			name:            "medium hp",
			hp:              40,
			san:             100,
			tensionValue:    0,
			expectedFactors: []string{"medium_hp"},
		},
		{
			name:            "low san",
			hp:              100,
			san:             20,
			tensionValue:    0,
			expectedFactors: []string{"low_san"},
		},
		{
			name:            "medium san",
			hp:              100,
			san:             40,
			tensionValue:    0,
			expectedFactors: []string{"medium_san"},
		},
		{
			name:            "high tension",
			hp:              100,
			san:             100,
			tensionValue:    85,
			expectedFactors: []string{"high_tension"},
		},
		{
			name:            "medium tension",
			hp:              100,
			san:             100,
			tensionValue:    65,
			expectedFactors: []string{"medium_tension"},
		},
		{
			name:            "rule warnings",
			hp:              100,
			san:             100,
			tensionValue:    0,
			ruleWarnings:    map[string]int{"rule1": 2},
			expectedFactors: []string{"rule_warnings"},
		},
		{
			name:            "multiple factors",
			hp:              20,
			san:             20,
			tensionValue:    85,
			ruleWarnings:    map[string]int{"rule1": 1},
			expectedFactors: []string{"low_hp", "low_san", "high_tension", "rule_warnings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			state.SetHP(tt.hp)
			state.SetSAN(tt.san)
			if state.Tension != nil {
				state.Tension.SetValue(tt.tensionValue)
			}
			state.RuleWarnings = tt.ruleWarnings

			builder := NewContextBuilder()
			ctx := builder.BuildContext(state, 0)

			if len(ctx.RiskFactors) != len(tt.expectedFactors) {
				t.Errorf("expected %d risk factors, got %d: %v", len(tt.expectedFactors), len(ctx.RiskFactors), ctx.RiskFactors)
				return
			}

			// Check all expected factors are present
			for _, expectedFactor := range tt.expectedFactors {
				found := false
				for _, actualFactor := range ctx.RiskFactors {
					if actualFactor == expectedFactor {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected risk factor '%s' not found in %v", expectedFactor, ctx.RiskFactors)
				}
			}
		})
	}
}

// TestBuildContext_PlotPointIntegration tests plot point detection.
// Story 6.6 AC2: IsPlotPoint/PlotPointType 欄位
// Story 6.6 Dev Notes: 測試劇情點檢測整合
func TestBuildContext_PlotPointIntegration(t *testing.T) {
	tests := []struct {
		name         string
		beat         int
		seeds        []*seed.GlobalSeed
		expectedPlot bool
		expectedType string
	}{
		{
			name:         "no plot point",
			beat:         15,
			seeds:        nil,
			expectedPlot: false,
			expectedType: "",
		},
		{
			name:         "act transition",
			beat:         30,
			seeds:        nil,
			expectedPlot: true,
			expectedType: PlotPointActTransition,
		},
		{
			name:         "seed milestone 30%",
			beat:         20,
			seeds:        []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}},
			expectedPlot: true,
			expectedType: PlotPointSeedMilestone30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			for i := 0; i < tt.beat; i++ {
				state.IncrementBeat()
			}
			if tt.seeds != nil {
				state.GlobalSeeds = tt.seeds
			}

			builder := NewContextBuilder()
			ctx := builder.BuildContext(state, 0)

			if ctx.IsPlotPoint != tt.expectedPlot {
				t.Errorf("expected IsPlotPoint %v, got %v", tt.expectedPlot, ctx.IsPlotPoint)
			}
			if ctx.PlotPointType != tt.expectedType {
				t.Errorf("expected PlotPointType '%s', got '%s'", tt.expectedType, ctx.PlotPointType)
			}
		})
	}
}

// TestBuildContext_NPCInitiation tests NPC conversation detection.
// Story 6.6 AC3: NPCInitiatesConversation/InitiatingNPC 欄位
// Story 6.6 Dev Notes: 測試 NPC 互動狀態檢測
func TestBuildContext_NPCInitiation(t *testing.T) {
	state := engine.NewGameStateV2()

	builder := NewContextBuilder()
	ctx := builder.BuildContext(state, 0)

	// Currently returns false as NPCManager integration is pending
	if ctx.NPCInitiatesConversation {
		t.Error("expected NPCInitiatesConversation false (placeholder)")
	}
	if ctx.InitiatingNPC != "" {
		t.Errorf("expected empty InitiatingNPC, got '%s'", ctx.InitiatingNPC)
	}

	// Test with nil NPCManager
	state.NPCManager = nil
	ctx = builder.BuildContext(state, 0)
	if ctx.NPCInitiatesConversation {
		t.Error("expected NPCInitiatesConversation false with nil NPCManager")
	}
}

// TestBuildContext_PendingEvents tests event extraction.
// Story 6.6 AC4: PendingEvents 欄位
// Story 6.6 Dev Notes: 測試事件提取
func TestBuildContext_PendingEvents(t *testing.T) {
	tests := []struct {
		name           string
		setupState     func(*engine.GameStateV2)
		expectedEvents int
		checkEvent     func(*testing.T, *GameEvent)
	}{
		{
			name: "no events",
			setupState: func(state *engine.GameStateV2) {
				// Empty state
			},
			expectedEvents: 0,
		},
		{
			name: "seed reveal pending event",
			setupState: func(state *engine.GameStateV2) {
				// Add a seed that's ready to reveal (in beat window)
				globalSeed, _ := seed.NewGlobalSeed(
					"GS001",
					"Test seed content",
					"truth1",
					"tragic",
					[]seed.ClueTier{
						{Tier: 1, Content: "Tier 1 clue", BeatStart: 0, BeatEnd: 10},
						{Tier: 2, Content: "Tier 2 clue", BeatStart: 11, BeatEnd: 20},
						{Tier: 3, Content: "Tier 3 clue", BeatStart: 21, BeatEnd: 30},
					},
				)
				state.AddGlobalSeed(globalSeed)
				// Advance to beat 5 (within tier 1 window)
				for i := 0; i < 5; i++ {
					state.IncrementBeat()
				}
			},
			expectedEvents: 1,
			checkEvent: func(t *testing.T, event *GameEvent) {
				if event.Type != "seed_reveal_pending" {
					t.Errorf("expected type 'seed_reveal_pending', got '%s'", event.Type)
				}
				if !event.IsMajor {
					t.Error("seed reveal should be major event")
				}
			},
		},
		{
			name: "rule violation event",
			setupState: func(state *engine.GameStateV2) {
				state.RuleWarnings = map[string]int{
					"rule1": 3,
					"rule2": 1,
				}
			},
			expectedEvents: 1,
			checkEvent: func(t *testing.T, event *GameEvent) {
				if event.Type != "rule_violation" {
					t.Errorf("expected type 'rule_violation', got '%s'", event.Type)
				}
				if !event.IsMajor {
					t.Error("rule violation should be major event")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := engine.NewGameStateV2()
			tt.setupState(state)

			builder := NewContextBuilder()
			ctx := builder.BuildContext(state, 0)

			if len(ctx.PendingEvents) != tt.expectedEvents {
				t.Errorf("expected %d events, got %d", tt.expectedEvents, len(ctx.PendingEvents))
				return
			}

			if tt.expectedEvents > 0 && tt.checkEvent != nil {
				tt.checkEvent(t, ctx.PendingEvents[0])
			}
		})
	}
}

// TestBuildContext_AutoResolvedBeats tests auto-resolved beats tracking.
// Story 6.6 AC4: AutoResolvedBeats 欄位
func TestBuildContext_AutoResolvedBeats(t *testing.T) {
	state := engine.NewGameStateV2()
	builder := NewContextBuilder()

	tests := []struct {
		autoBeats int
	}{
		{0},
		{3},
		{5},
		{10},
	}

	for _, tt := range tests {
		ctx := builder.BuildContext(state, tt.autoBeats)
		if ctx.AutoResolvedBeats != tt.autoBeats {
			t.Errorf("expected AutoResolvedBeats %d, got %d", tt.autoBeats, ctx.AutoResolvedBeats)
		}
	}
}

// TestBuildContext_RecentChoices tests choice history extraction.
// Story 6.6 AC4: RecentChoices 欄位
func TestBuildContext_RecentChoices(t *testing.T) {
	state := engine.NewGameStateV2()
	builder := NewContextBuilder()
	ctx := builder.BuildContext(state, 0)

	// Currently returns empty slice as choice history not yet implemented
	if ctx.RecentChoices == nil {
		t.Error("RecentChoices should not be nil")
	}
	if len(ctx.RecentChoices) != 0 {
		t.Errorf("expected empty RecentChoices (placeholder), got %d items", len(ctx.RecentChoices))
	}
}

// TestBuildContextForController tests the controller integration method.
// Story 6.6: Integration with MomentumController
func TestBuildContextForController(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	for i := 0; i < 20; i++ {
		state.IncrementBeat()
	}
	state.CurrentScene = "test_scene"
	state.SetHP(50)

	ctx := ctrl.BuildContext(state, 3)

	if ctx == nil {
		t.Fatal("BuildContext returned nil")
	}
	if ctx.CurrentBeat != 20 {
		t.Errorf("expected CurrentBeat 20, got %d", ctx.CurrentBeat)
	}
	if ctx.CurrentScene != "test_scene" {
		t.Errorf("expected CurrentScene 'test_scene', got '%s'", ctx.CurrentScene)
	}
	if ctx.AutoResolvedBeats != 3 {
		t.Errorf("expected AutoResolvedBeats 3, got %d", ctx.AutoResolvedBeats)
	}
}

// TestBuildContext_IntegrationWithPlotDetectorState tests that PlotDetector state is preserved.
// Story 6.6 Dev Notes: 確保 PlotDetector milestones 狀態正確維護
func TestBuildContext_IntegrationWithPlotDetectorState(t *testing.T) {
	ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})

	state := engine.NewGameStateV2()
	state.GlobalSeeds = []*seed.GlobalSeed{{ID: "GS001", CurrentTier: 2}}
	for i := 0; i < 20; i++ {
		state.IncrementBeat()
	}

	// First call should detect 30% milestone
	ctx1 := ctrl.BuildContext(state, 0)
	if !ctx1.IsPlotPoint || ctx1.PlotPointType != PlotPointSeedMilestone30 {
		t.Error("first call should detect seed milestone 30%")
	}

	// Second call should NOT detect it again (milestone already reached)
	ctx2 := ctrl.BuildContext(state, 0)
	if ctx2.IsPlotPoint && ctx2.PlotPointType == PlotPointSeedMilestone30 {
		t.Error("second call should not re-detect same milestone")
	}
}

// TestExtractRiskFactors tests risk factor extraction in isolation.
func TestExtractRiskFactors(t *testing.T) {
	builder := NewContextBuilder()

	// Test with nil tension
	state := engine.NewGameStateV2()
	state.Tension = nil
	state.SetHP(20)

	factors := builder.extractRiskFactors(state)
	if len(factors) == 0 {
		t.Error("expected at least low_hp factor")
	}
}

// TestCheckNPCInitiation tests NPC initiation check in isolation.
func TestCheckNPCInitiation(t *testing.T) {
	builder := NewContextBuilder()

	// Test with nil NPCManager
	state := engine.NewGameStateV2()
	state.NPCManager = nil
	initiates, npcID := builder.checkNPCInitiation(state)
	if initiates {
		t.Error("expected false with nil NPCManager")
	}
	if npcID != "" {
		t.Error("expected empty npcID with nil NPCManager")
	}

	// Test with initialized but empty NPCManager
	state.NPCManager = &engine.NPCManagerState{
		Profiles: make(map[string]interface{}),
		States:   make(map[string]interface{}),
	}
	initiates, npcID = builder.checkNPCInitiation(state)
	if initiates {
		t.Error("expected false with empty NPCManager (placeholder)")
	}
	if npcID != "" {
		t.Error("expected empty npcID (placeholder)")
	}
}

// TestExtractPendingEvents tests event extraction in isolation.
func TestExtractPendingEvents(t *testing.T) {
	builder := NewContextBuilder()

	// Test with multiple event types
	state := engine.NewGameStateV2()

	// Add seed with pending revelation
	globalSeed, _ := seed.NewGlobalSeed(
		"GS001",
		"Test seed",
		"truth1",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Tier 1", BeatStart: 0, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2", BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", BeatStart: 21, BeatEnd: 30},
		},
	)
	state.AddGlobalSeed(globalSeed)

	// Add rule warnings
	state.RuleWarnings = map[string]int{"rule1": 3, "rule2": 4}

	// Advance beats
	for i := 0; i < 5; i++ {
		state.IncrementBeat()
	}

	events := builder.extractPendingEvents(state)

	// Should have both seed harvest and rule violations
	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	// Check event types
	hasReveal := false
	hasViolation := false
	for _, event := range events {
		if event.Type == "seed_reveal_pending" {
			hasReveal = true
		}
		if event.Type == "rule_violation" {
			hasViolation = true
		}
	}

	if !hasReveal {
		t.Error("expected seed_reveal_pending event")
	}
	if !hasViolation {
		t.Error("expected rule_violation event")
	}
}

// TestExtractRecentChoices tests choice extraction in isolation.
func TestExtractRecentChoices(t *testing.T) {
	builder := NewContextBuilder()
	state := engine.NewGameStateV2()

	choices := builder.extractRecentChoices(state)
	if choices == nil {
		t.Error("expected non-nil slice")
	}
	if len(choices) != 0 {
		t.Error("expected empty slice (placeholder)")
	}
}

// TestBuildContext_ComprehensiveScenario tests a complex real-world scenario.
// Story 6.6: 綜合測試所有欄位
func TestBuildContext_ComprehensiveScenario(t *testing.T) {
	state := engine.NewGameStateV2()

	// Setup complex state
	for i := 0; i < 28; i++ {
		state.IncrementBeat()
	}
	state.CurrentScene = "final_corridor"
	state.SetHP(30)
	state.SetSAN(40)
	state.Tension.SetValue(75)

	// Add seeds approaching milestone
	globalSeedComplex, _ := seed.NewGlobalSeed(
		"GS001",
		"Test seed",
		"truth1",
		"tragic",
		[]seed.ClueTier{
			{Tier: 1, Content: "Tier 1", BeatStart: 0, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2", BeatStart: 11, BeatEnd: 40},
			{Tier: 3, Content: "Tier 3", BeatStart: 41, BeatEnd: 50},
		},
	)
	globalSeedComplex.CurrentTier = 2 // Manually set to tier 2
	state.GlobalSeeds = []*seed.GlobalSeed{globalSeedComplex}

	// Add rule warnings
	state.RuleWarnings = map[string]int{"rule1": 2}

	builder := NewContextBuilder()
	ctx := builder.BuildContext(state, 2)

	// Verify all fields
	if ctx.CurrentBeat != 28 {
		t.Errorf("CurrentBeat: expected 28, got %d", ctx.CurrentBeat)
	}
	if ctx.CurrentScene != "final_corridor" {
		t.Errorf("CurrentScene: expected 'final_corridor', got '%s'", ctx.CurrentScene)
	}

	// Should have multiple risk factors
	if len(ctx.RiskFactors) < 3 {
		t.Errorf("expected at least 3 risk factors, got %d: %v", len(ctx.RiskFactors), ctx.RiskFactors)
	}

	// Should detect plot point (near act end)
	if !ctx.IsPlotPoint {
		t.Error("expected plot point detection at beat 28 (near act end)")
	}

	// Should have auto-resolved beats
	if ctx.AutoResolvedBeats != 2 {
		t.Errorf("AutoResolvedBeats: expected 2, got %d", ctx.AutoResolvedBeats)
	}

	// Should have pending events (seed reveal pending)
	if len(ctx.PendingEvents) == 0 {
		t.Error("expected pending events (seed reveal)")
	}

	// Verify not nil arrays
	if ctx.RecentChoices == nil {
		t.Error("RecentChoices should not be nil")
	}
	if ctx.PendingEvents == nil {
		t.Error("PendingEvents should not be nil")
	}
	if ctx.RiskFactors == nil {
		t.Error("RiskFactors should not be nil")
	}
}
