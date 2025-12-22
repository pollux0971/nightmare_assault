// Package momentum 提供敘事動量控制系統
// Epic 6: Story 6.1 - MomentumController 基礎結構測試
package momentum

import (
	"testing"
)

// mockNarrationAgent 是用於測試的 mock 實作
type mockNarrationAgent struct{}

// TestNewMomentumController 測試建構函數
func TestNewMomentumController(t *testing.T) {
	tests := []struct {
		name            string
		config          *MomentumConfig
		narrationAgent  *mockNarrationAgent
		expectedConfig  *MomentumConfig
	}{
		{
			name:           "with custom config",
			config:         &MomentumConfig{Frequency: FrequencyHigh},
			narrationAgent: &mockNarrationAgent{},
			expectedConfig: &MomentumConfig{Frequency: FrequencyHigh},
		},
		{
			name:           "with nil config uses default",
			config:         nil,
			narrationAgent: &mockNarrationAgent{},
			expectedConfig: DefaultMomentumConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewMomentumController(tt.config, tt.narrationAgent)

			if ctrl == nil {
				t.Fatal("NewMomentumController returned nil")
			}

			if ctrl.config == nil {
				t.Fatal("config is nil")
			}

			if ctrl.riskEvaluator == nil {
				t.Error("riskEvaluator not initialized")
			}

			if ctrl.plotDetector == nil {
				t.Error("plotDetector not initialized")
			}

			// 驗證配置正確
			if ctrl.config.Frequency != tt.expectedConfig.Frequency {
				t.Errorf("expected Frequency %v, got %v", tt.expectedConfig.Frequency, ctrl.config.Frequency)
			}
		})
	}
}

// TestDefaultMomentumConfig 測試預設配置
func TestDefaultMomentumConfig(t *testing.T) {
	config := DefaultMomentumConfig()

	if config == nil {
		t.Fatal("DefaultMomentumConfig returned nil")
	}

	// 驗證預設值
	if config.Frequency != FrequencyMedium {
		t.Errorf("expected Frequency Medium, got %v", config.Frequency)
	}

	if !config.AutoResolve {
		t.Error("expected AutoResolve to be true")
	}

	if config.MaxAutoBeats != 5 {
		t.Errorf("expected MaxAutoBeats 5, got %d", config.MaxAutoBeats)
	}

	if config.PauseOnRisk != RiskMedium {
		t.Errorf("expected PauseOnRisk Medium, got %v", config.PauseOnRisk)
	}

	if !config.PauseOnPlot {
		t.Error("expected PauseOnPlot to be true")
	}

	if !config.PauseOnNPC {
		t.Error("expected PauseOnNPC to be true")
	}

	if !config.PauseOnEvent {
		t.Error("expected PauseOnEvent to be true")
	}

	if !config.PlayerOverride {
		t.Error("expected PlayerOverride to be true")
	}
}

// TestShouldPauseForChoice_NPCConversation 測試 NPC 對話優先級最高
func TestShouldPauseForChoice_NPCConversation(t *testing.T) {
	config := &MomentumConfig{
		PauseOnNPC:   true,
		Frequency:    FrequencyLow, // 即使是低頻率
		PauseOnPlot:  false,
		PauseOnEvent: false,
		PauseOnRisk:  RiskLethal, // 設極高門檻
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	ctx := &NarrativeContext{
		NPCInitiatesConversation: true,
		InitiatingNPC:            "Dr. Zhang",
		RiskLevel:                RiskNone, // 即使無風險
	}

	// 應該暫停進入聊天室
	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected to pause for NPC conversation")
	}

	// 測試禁用 PauseOnNPC (同時提高風險等級以觸發 frequency 檢查)
	config.PauseOnNPC = false
	ctx.RiskLevel = RiskHigh // FrequencyLow 在 RiskHigh 時會暫停
	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected to pause for high risk with FrequencyLow")
	}

	// 測試真正不暫停的情況
	ctx.RiskLevel = RiskNone
	if ctrl.ShouldPauseForChoice(ctx) {
		t.Error("should not pause when PauseOnNPC is false and no other conditions met")
	}
}

// TestShouldPauseForChoice_MajorEvent 測試重大事件觸發暫停
func TestShouldPauseForChoice_MajorEvent(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	ctx := &NarrativeContext{
		PendingEvents: []*GameEvent{
			{
				ID:      "event1",
				IsMajor: true,
				Type:    "npc_death",
			},
		},
		RiskLevel: RiskLow,
	}

	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected to pause for major event")
	}

	// 測試非重大事件
	ctx.PendingEvents[0].IsMajor = false
	// 應該不暫停 (因為 RiskLow < RiskMedium, FrequencyMedium)
	if ctrl.ShouldPauseForChoice(ctx) {
		t.Error("should not pause for non-major event with low risk")
	}
}

// TestShouldPauseForChoice_PlotPoint 測試劇情點暫停
func TestShouldPauseForChoice_PlotPoint(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	ctx := &NarrativeContext{
		IsPlotPoint:   true,
		PlotPointType: "revelation",
		RiskLevel:     RiskLow,
	}

	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected to pause for plot point")
	}

	// 測試禁用 PauseOnPlot
	config.PauseOnPlot = false
	if ctrl.ShouldPauseForChoice(ctx) {
		t.Error("should not pause when PauseOnPlot is false")
	}
}

// TestShouldPauseForChoice_RiskLevel 測試風險等級暫停
func TestShouldPauseForChoice_RiskLevel(t *testing.T) {
	tests := []struct {
		name          string
		pauseOnRisk   RiskLevel
		currentRisk   RiskLevel
		shouldPause   bool
	}{
		{"Risk equal to threshold", RiskMedium, RiskMedium, true},
		{"Risk above threshold", RiskMedium, RiskHigh, true},
		{"Risk below threshold", RiskMedium, RiskLow, false},
		{"Lethal risk always pauses", RiskHigh, RiskLethal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MomentumConfig{
				PauseOnRisk:  tt.pauseOnRisk,
				Frequency:    FrequencyLow, // 設低頻率避免干擾
				PauseOnPlot:  false,
				PauseOnNPC:   false,
				PauseOnEvent: false,
			}
			ctrl := NewMomentumController(config, &mockNarrationAgent{})

			ctx := &NarrativeContext{
				RiskLevel: tt.currentRisk,
			}

			result := ctrl.ShouldPauseForChoice(ctx)
			if result != tt.shouldPause {
				t.Errorf("expected %v for risk %v (threshold %v), got %v",
					tt.shouldPause, tt.currentRisk, tt.pauseOnRisk, result)
			}
		})
	}
}

// TestShouldPauseForChoice_MaxAutoBeats 測試最大自動回合數
func TestShouldPauseForChoice_MaxAutoBeats(t *testing.T) {
	config := &MomentumConfig{
		AutoResolve:   true,
		MaxAutoBeats:  5,
		Frequency:     FrequencyLow,
		PauseOnRisk:   RiskHigh, // 設高門檻避免干擾
		PauseOnPlot:   false,
		PauseOnNPC:    false,
		PauseOnEvent:  false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	tests := []struct {
		name            string
		autoResolvedBeats int
		shouldPause     bool
	}{
		{"below max", 3, false},
		{"at max", 5, true},
		{"above max", 7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &NarrativeContext{
				AutoResolvedBeats: tt.autoResolvedBeats,
				RiskLevel:         RiskLow,
			}

			result := ctrl.ShouldPauseForChoice(ctx)
			if result != tt.shouldPause {
				t.Errorf("expected %v for %d auto beats (max %d), got %v",
					tt.shouldPause, tt.autoResolvedBeats, config.MaxAutoBeats, result)
			}
		})
	}
}

// TestShouldPauseForChoice_FrequencyLevels 測試頻率等級影響
func TestShouldPauseForChoice_FrequencyLevels(t *testing.T) {
	tests := []struct {
		freq        FrequencyLevel
		risk        RiskLevel
		expected    bool
		description string
	}{
		{FrequencyHigh, RiskNone, true, "High frequency always pauses"},
		{FrequencyHigh, RiskLow, true, "High frequency with low risk"},
		{FrequencyMedium, RiskNone, false, "Medium frequency with no risk"},
		{FrequencyMedium, RiskLow, false, "Medium frequency with low risk"},
		{FrequencyMedium, RiskMedium, true, "Medium frequency with medium risk"},
		{FrequencyMedium, RiskHigh, true, "Medium frequency with high risk"},
		{FrequencyLow, RiskNone, false, "Low frequency with no risk"},
		{FrequencyLow, RiskLow, false, "Low frequency with low risk"},
		{FrequencyLow, RiskMedium, false, "Low frequency with medium risk"},
		{FrequencyLow, RiskHigh, true, "Low frequency with high risk"},
		{FrequencyLow, RiskLethal, true, "Low frequency with lethal risk"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			config := &MomentumConfig{
				Frequency:    tt.freq,
				PauseOnRisk:  RiskLethal, // 設極高門檻，避免干擾頻率測試
				PauseOnPlot:  false,
				PauseOnNPC:   false,
				PauseOnEvent: false,
			}
			ctrl := NewMomentumController(config, &mockNarrationAgent{})

			ctx := &NarrativeContext{
				RiskLevel: tt.risk,
			}

			result := ctrl.ShouldPauseForChoice(ctx)
			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.expected, result)
			}
		})
	}
}

// TestShouldPauseForChoice_Priority 測試優先級順序
func TestShouldPauseForChoice_Priority(t *testing.T) {
	// NPC 對話優先級高於其他條件
	config := &MomentumConfig{
		PauseOnNPC:   true,
		PauseOnEvent: false, // 關閉重大事件
		Frequency:    FrequencyLow,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	ctx := &NarrativeContext{
		NPCInitiatesConversation: true,
		InitiatingNPC:            "NPC1",
		PendingEvents: []*GameEvent{
			{IsMajor: true, Type: "event"},
		},
		RiskLevel: RiskNone,
	}

	// 即使 PauseOnEvent 關閉，仍應因 NPC 對話暫停
	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("NPC conversation should have highest priority")
	}
}

// TestShouldPauseForChoice_NilContext 測試 nil context
func TestShouldPauseForChoice_NilContext(t *testing.T) {
	ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})

	// nil context 應該安全地返回 true (暫停)
	if !ctrl.ShouldPauseForChoice(nil) {
		t.Error("expected to pause for nil context (safety)")
	}
}

// TestDetermineStopReason 測試暫停原因判定
func TestDetermineStopReason(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	tests := []struct {
		name     string
		ctx      *NarrativeContext
		expected StopReason
	}{
		{
			name: "NPC conversation",
			ctx: &NarrativeContext{
				NPCInitiatesConversation: true,
			},
			expected: StopReasonNPCConversation,
		},
		{
			name: "Major event",
			ctx: &NarrativeContext{
				PendingEvents: []*GameEvent{{IsMajor: true}},
			},
			expected: StopReasonMajorEvent,
		},
		{
			name: "Plot point",
			ctx: &NarrativeContext{
				IsPlotPoint: true,
			},
			expected: StopReasonPlotPoint,
		},
		{
			name: "Risk level",
			ctx: &NarrativeContext{
				RiskLevel: RiskHigh,
			},
			expected: StopReasonRiskLevel,
		},
		{
			name: "Max auto beats",
			ctx: &NarrativeContext{
				AutoResolvedBeats: 5,
				RiskLevel:         RiskLow,
			},
			expected: StopReasonMaxAutoBeats,
		},
		{
			name:     "nil context",
			ctx:      nil,
			expected: StopReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := ctrl.DetermineStopReason(tt.ctx)
			if reason != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, reason)
			}
		})
	}
}

// TestGetSetConfig 測試配置的 getter/setter
func TestGetSetConfig(t *testing.T) {
	ctrl := NewMomentumController(DefaultMomentumConfig(), &mockNarrationAgent{})

	// 測試 Get
	config := ctrl.GetConfig()
	if config == nil {
		t.Fatal("GetConfig returned nil")
	}

	// 測試 Set
	newConfig := &MomentumConfig{
		Frequency: FrequencyHigh,
	}
	ctrl.SetConfig(newConfig)

	if ctrl.GetConfig().Frequency != FrequencyHigh {
		t.Error("SetConfig did not update config")
	}

	// 測試 Set nil (應該忽略)
	ctrl.SetConfig(nil)
	if ctrl.GetConfig() == nil {
		t.Error("SetConfig(nil) should not clear config")
	}
}

// TestGetMomentumConfigForDifficulty 測試難度對應配置
func TestGetMomentumConfigForDifficulty(t *testing.T) {
	tests := []struct {
		difficulty    string
		expectedFreq  FrequencyLevel
		expectedBeats int
		expectedRisk  RiskLevel
	}{
		{"easy", FrequencyHigh, 3, RiskLow},
		{"Easy", FrequencyHigh, 3, RiskLow},
		{"normal", FrequencyMedium, 5, RiskMedium},
		{"Normal", FrequencyMedium, 5, RiskMedium},
		{"hard", FrequencyMedium, 7, RiskHigh}, // Story 7-5 AC3: Hard uses RiskHigh
		{"Hard", FrequencyMedium, 7, RiskHigh},
		{"nightmare", FrequencyLow, 10, RiskHigh},
		{"Nightmare", FrequencyLow, 10, RiskHigh},
		{"hell", FrequencyLow, 10, RiskLethal}, // Story 7-5 AC4: Hell uses RiskLethal
		{"Hell", FrequencyLow, 10, RiskLethal},
		{"unknown", FrequencyMedium, 5, RiskMedium}, // 預設值
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			config := GetMomentumConfigForDifficulty(tt.difficulty)

			if config.Frequency != tt.expectedFreq {
				t.Errorf("expected Frequency %v, got %v", tt.expectedFreq, config.Frequency)
			}

			if config.MaxAutoBeats != tt.expectedBeats {
				t.Errorf("expected MaxAutoBeats %d, got %d", tt.expectedBeats, config.MaxAutoBeats)
			}

			if config.PauseOnRisk != tt.expectedRisk {
				t.Errorf("expected PauseOnRisk %v, got %v", tt.expectedRisk, config.PauseOnRisk)
			}
		})
	}
}

// TestCinematicModeConfig 測試電影模式配置
func TestCinematicModeConfig(t *testing.T) {
	config := CinematicModeConfig()

	if config.Frequency != FrequencyLow {
		t.Errorf("expected FrequencyLow, got %v", config.Frequency)
	}

	if config.MaxAutoBeats != 15 {
		t.Errorf("expected MaxAutoBeats 15, got %d", config.MaxAutoBeats)
	}

	if config.PauseOnRisk != RiskHigh {
		t.Errorf("expected PauseOnRisk High, got %v", config.PauseOnRisk)
	}

	if !config.AutoResolve {
		t.Error("expected AutoResolve to be true")
	}
}

// TestInteractiveModeConfig 測試互動模式配置
func TestInteractiveModeConfig(t *testing.T) {
	config := InteractiveModeConfig()

	if config.Frequency != FrequencyHigh {
		t.Errorf("expected FrequencyHigh, got %v", config.Frequency)
	}

	if config.AutoResolve {
		t.Error("expected AutoResolve to be false")
	}

	if config.MaxAutoBeats != 0 {
		t.Errorf("expected MaxAutoBeats 0, got %d", config.MaxAutoBeats)
	}

	if config.PauseOnRisk != RiskNone {
		t.Errorf("expected PauseOnRisk None, got %v", config.PauseOnRisk)
	}
}

// TestFrequencyLevelString 測試 FrequencyLevel String 方法
func TestFrequencyLevelString(t *testing.T) {
	tests := []struct {
		level    FrequencyLevel
		expected string
	}{
		{FrequencyHigh, "High"},
		{FrequencyMedium, "Medium"},
		{FrequencyLow, "Low"},
		{FrequencyLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("FrequencyLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

// TestRiskLevelString 測試 RiskLevel String 方法
func TestRiskLevelString(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
	}{
		{RiskNone, "None"},
		{RiskLow, "Low"},
		{RiskMedium, "Medium"},
		{RiskHigh, "High"},
		{RiskLethal, "Lethal"},
		{RiskLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("RiskLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

// TestStopReasonString 測試 StopReason String 方法
func TestStopReasonString(t *testing.T) {
	tests := []struct {
		reason   StopReason
		expected string
	}{
		{StopReasonNone, "None"},
		{StopReasonNPCConversation, "NPCConversation"},
		{StopReasonMajorEvent, "MajorEvent"},
		{StopReasonPlotPoint, "PlotPoint"},
		{StopReasonRiskLevel, "RiskLevel"},
		{StopReasonMaxAutoBeats, "MaxAutoBeats"},
		{StopReasonFrequency, "Frequency"},
		{StopReason(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.reason.String(); got != tt.expected {
			t.Errorf("StopReason(%d).String() = %s, want %s", tt.reason, got, tt.expected)
		}
	}
}

// TestShouldPauseForChoice_MultipleConditions 測試多條件同時滿足時的行為
func TestShouldPauseForChoice_MultipleConditions(t *testing.T) {
	config := DefaultMomentumConfig()
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	tests := []struct {
		name        string
		ctx         *NarrativeContext
		shouldPause bool
		description string
	}{
		{
			name: "All conditions true",
			ctx: &NarrativeContext{
				NPCInitiatesConversation: true,
				InitiatingNPC:            "NPC1",
				PendingEvents:            []*GameEvent{{IsMajor: true}},
				IsPlotPoint:              true,
				PlotPointType:            "revelation",
				RiskLevel:                RiskHigh,
				AutoResolvedBeats:        10,
			},
			shouldPause: true,
			description: "When all conditions are true, should pause (NPC has highest priority)",
		},
		{
			name: "Only frequency condition",
			ctx: &NarrativeContext{
				RiskLevel: RiskMedium, // FrequencyMedium pauses at RiskMedium
			},
			shouldPause: true,
			description: "Frequency condition alone should work",
		},
		{
			name: "No conditions met",
			ctx: &NarrativeContext{
				RiskLevel:         RiskNone,
				AutoResolvedBeats: 0,
			},
			shouldPause: false,
			description: "When no conditions met, should not pause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctrl.ShouldPauseForChoice(tt.ctx)
			if result != tt.shouldPause {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.shouldPause, result)
			}
		})
	}
}

// TestShouldPauseForChoice_ConfigSwitches 測試配置開關的完整影響
func TestShouldPauseForChoice_ConfigSwitches(t *testing.T) {
	tests := []struct {
		name        string
		config      *MomentumConfig
		ctx         *NarrativeContext
		shouldPause bool
	}{
		{
			name: "PauseOnEvent disabled with major event",
			config: &MomentumConfig{
				PauseOnEvent: false,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
			},
			ctx: &NarrativeContext{
				PendingEvents: []*GameEvent{{IsMajor: true}},
				RiskLevel:     RiskLow,
			},
			shouldPause: false,
		},
		{
			name: "AutoResolve disabled should not check MaxAutoBeats",
			config: &MomentumConfig{
				AutoResolve:  false,
				MaxAutoBeats: 5,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
				PauseOnEvent: false,
			},
			ctx: &NarrativeContext{
				AutoResolvedBeats: 10, // Above max but AutoResolve is false
				RiskLevel:         RiskLow,
			},
			shouldPause: false,
		},
		{
			name: "All switches disabled with FrequencyHigh",
			config: &MomentumConfig{
				Frequency:    FrequencyHigh,
				PauseOnNPC:   false,
				PauseOnEvent: false,
				PauseOnPlot:  false,
				PauseOnRisk:  RiskLethal,
			},
			ctx: &NarrativeContext{
				NPCInitiatesConversation: true,
				PendingEvents:            []*GameEvent{{IsMajor: true}},
				IsPlotPoint:              true,
				RiskLevel:                RiskNone,
			},
			shouldPause: true, // FrequencyHigh always pauses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewMomentumController(tt.config, &mockNarrationAgent{})
			result := ctrl.ShouldPauseForChoice(tt.ctx)
			if result != tt.shouldPause {
				t.Errorf("expected %v, got %v", tt.shouldPause, result)
			}
		})
	}
}

// TestShouldPauseForChoice_BoundaryConditions 測試邊界條件
func TestShouldPauseForChoice_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		config      *MomentumConfig
		ctx         *NarrativeContext
		shouldPause bool
	}{
		{
			name: "AutoResolvedBeats exactly at max",
			config: &MomentumConfig{
				AutoResolve:  true,
				MaxAutoBeats: 5,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
				PauseOnEvent: false,
			},
			ctx: &NarrativeContext{
				AutoResolvedBeats: 5,
				RiskLevel:         RiskLow,
			},
			shouldPause: true,
		},
		{
			name: "AutoResolvedBeats one below max",
			config: &MomentumConfig{
				AutoResolve:  true,
				MaxAutoBeats: 5,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
				PauseOnEvent: false,
			},
			ctx: &NarrativeContext{
				AutoResolvedBeats: 4,
				RiskLevel:         RiskLow,
			},
			shouldPause: false,
		},
		{
			name: "RiskLevel exactly at threshold",
			config: &MomentumConfig{
				PauseOnRisk:  RiskMedium,
				Frequency:    FrequencyLow,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
				PauseOnEvent: false,
			},
			ctx: &NarrativeContext{
				RiskLevel: RiskMedium,
			},
			shouldPause: true,
		},
		{
			name: "Empty PendingEvents list",
			config: &MomentumConfig{
				PauseOnEvent: true,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
			},
			ctx: &NarrativeContext{
				PendingEvents: []*GameEvent{},
				RiskLevel:     RiskLow,
			},
			shouldPause: false,
		},
		{
			name: "PendingEvents with no major events",
			config: &MomentumConfig{
				PauseOnEvent: true,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
			},
			ctx: &NarrativeContext{
				PendingEvents: []*GameEvent{
					{IsMajor: false, Type: "minor1"},
					{IsMajor: false, Type: "minor2"},
				},
				RiskLevel: RiskLow,
			},
			shouldPause: false,
		},
		{
			name: "PendingEvents with one major among minors",
			config: &MomentumConfig{
				PauseOnEvent: true,
				Frequency:    FrequencyLow,
				PauseOnRisk:  RiskHigh,
				PauseOnNPC:   false,
				PauseOnPlot:  false,
			},
			ctx: &NarrativeContext{
				PendingEvents: []*GameEvent{
					{IsMajor: false, Type: "minor1"},
					{IsMajor: true, Type: "major"},
					{IsMajor: false, Type: "minor2"},
				},
				RiskLevel: RiskLow,
			},
			shouldPause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewMomentumController(tt.config, &mockNarrationAgent{})
			result := ctrl.ShouldPauseForChoice(tt.ctx)
			if result != tt.shouldPause {
				t.Errorf("expected %v, got %v", tt.shouldPause, result)
			}
		})
	}
}

// TestShouldPauseByFrequency_DefaultCase 測試 shouldPauseByFrequency 的預設分支
func TestShouldPauseByFrequency_DefaultCase(t *testing.T) {
	config := &MomentumConfig{
		Frequency:    FrequencyLevel(999), // Invalid frequency level
		PauseOnRisk:  RiskLethal,
		PauseOnNPC:   false,
		PauseOnPlot:  false,
		PauseOnEvent: false,
	}
	ctrl := NewMomentumController(config, &mockNarrationAgent{})

	ctx := &NarrativeContext{
		RiskLevel: RiskMedium,
	}

	// Default case should behave like FrequencyMedium: pause at RiskMedium+
	if !ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected default frequency behavior to pause at RiskMedium")
	}

	ctx.RiskLevel = RiskLow
	if ctrl.ShouldPauseForChoice(ctx) {
		t.Error("expected default frequency behavior to not pause at RiskLow")
	}
}
