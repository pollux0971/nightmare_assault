package narration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContentRequest_Validation tests content request validation
func TestContentRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *ContentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
		{
			name: "negative beat",
			req: &ContentRequest{
				Beat:      -1,
				GameState: engine.NewGameStateV2(),
			},
			wantErr: true,
			errMsg:  "beat must be >= 0",
		},
		{
			name: "nil game state",
			req: &ContentRequest{
				Beat:      1,
				GameState: nil,
			},
			wantErr: true,
			errMsg:  "game state cannot be nil",
		},
		{
			name: "valid request - minimal",
			req: &ContentRequest{
				Beat:      0,
				GameState: engine.NewGameStateV2(),
			},
			wantErr: false,
		},
		{
			name: "valid request - complete",
			req: &ContentRequest{
				Beat:             5,
				GameState:        engine.NewGameStateV2(),
				LastPlayerChoice: "探索房間",
				JudgeResult: &JudgeResult{
					ImpactLevel: "minor",
					HPDelta:     -5,
					SANDelta:    -3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentRequest(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestInvokeContent_BasicNarrative tests basic narrative generation
// AC #1: 返回 500-1200 字的敘事內容（MainNarrative）
func TestInvokeContent_BasicNarrative(t *testing.T) {
	// Setup agent with mock LLM
	agent := createTestNarrationAgent(t)

	req := &ContentRequest{
		Beat:             1,
		GameState:        engine.NewGameStateV2(),
		LastPlayerChoice: "探索房間",
		JudgeResult: &JudgeResult{
			ImpactLevel: "none",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.MainNarrative)

	// AC #1: 字數範圍 500-1200
	// TODO: 目前是 placeholder，等實作完成後需要驗證字數
	// narrativeLength := len([]rune(resp.MainNarrative))
	// assert.GreaterOrEqual(t, narrativeLength, 500, "Narrative too short")
	// assert.LessOrEqual(t, narrativeLength, 1200, "Narrative too long")
}

// TestInvokeContent_ResponseStructure tests response structure validity
func TestInvokeContent_ResponseStructure(t *testing.T) {
	agent := createTestNarrationAgent(t)

	req := &ContentRequest{
		Beat:      1,
		GameState: engine.NewGameStateV2(),
	}

	ctx := context.Background()
	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify all fields are initialized (not nil slices)
	assert.NotNil(t, resp.PlantedSeeds)
	assert.NotNil(t, resp.HarvestedSeeds)
	assert.NotNil(t, resp.RuleHints)
}

// TestInvokeContent_InvalidRequest tests error handling for invalid requests
func TestInvokeContent_InvalidRequest(t *testing.T) {
	agent := createTestNarrationAgent(t)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *ContentRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "nil game state",
			req: &ContentRequest{
				Beat:      1,
				GameState: nil,
			},
		},
		{
			name: "negative beat",
			req: &ContentRequest{
				Beat:      -1,
				GameState: engine.NewGameStateV2(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := agent.InvokeContent(ctx, tt.req)
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

// createTestNarrationAgent creates a test narration agent with mock dependencies
func createTestNarrationAgent(t *testing.T) *NarrationAgent {
	t.Helper()

	config := agents.AgentConfig{
		Name:       "TestNarrationAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		// LLMClient will be mocked in later tests
		LLMClient: nil, // TODO: Add mock LLM client
	}

	return NewNarrationAgent(config)
}

// TestBuildContentPrompt tests comprehensive prompt building
func TestBuildContentPrompt(t *testing.T) {
	agent := createTestNarrationAgent(t)

	// Setup game state with various elements
	gameState := engine.NewGameStateV2()
	gameState.SetHP(85)
	gameState.SetSAN(70)
	gameState.CurrentScene = "醫院走廊"
	gameState.ActiveRules = []*engine.ActiveRule{
		{ID: "rule-001", Name: "不要在夜晚開燈"},
		{ID: "rule-002", Name: "不要直視鏡子"},
	}
	gameState.RuleWarnings = map[string]int{
		"rule-001": 1,
	}
	gameState.NPCStates = map[string]*engine.NPCState{
		"npc-001": {ID: "npc-001", Name: "護士"},
	}

	req := &ContentRequest{
		Beat:             5,
		GameState:        gameState,
		LastPlayerChoice: "探索房間",
		Difficulty:       "normal",
		JudgeResult: &JudgeResult{
			ViolatedRules: []string{"rule-001"},
			ImpactLevel:   "minor",
			HPDelta:       -5,
			SANDelta:      -3,
			Reason:        "你在夜晚開燈",
		},
	}

	directive := engine.GenerateDirective(engine.TensionLevelMedium)

	prompt, err := agent.buildContentPrompt(req, directive)

	require.NoError(t, err)
	assert.NotEmpty(t, prompt)

	// Verify prompt contains all sections
	assert.Contains(t, prompt, "系統角色", "Should contain system section")
	assert.Contains(t, prompt, "當前狀態", "Should contain state section")
	assert.Contains(t, prompt, "張力指令", "Should contain tension section")
	assert.Contains(t, prompt, "活躍規則", "Should contain rules section")
	assert.Contains(t, prompt, "活躍 NPCs", "Should contain NPCs section")
	assert.Contains(t, prompt, "玩家上下文", "Should contain context section")
	assert.Contains(t, prompt, "輸出要求", "Should contain constraints section")

	// Verify specific content
	assert.Contains(t, prompt, "Beat: 5")
	assert.Contains(t, prompt, "HP: 85 / 100")
	assert.Contains(t, prompt, "SAN: 70 / 100")
	assert.Contains(t, prompt, "場景: 醫院走廊")
	assert.Contains(t, prompt, "難度: normal")
	assert.Contains(t, prompt, "不要在夜晚開燈")
	assert.Contains(t, prompt, "護士")
	assert.Contains(t, prompt, "探索房間")
	assert.Contains(t, prompt, "HP 變化: -5")
	assert.Contains(t, prompt, "SAN 變化: -3")
}

// TestBuildSystemSection tests system section building
func TestBuildSystemSection(t *testing.T) {
	agent := createTestNarrationAgent(t)
	section := agent.buildSystemSection()

	assert.Contains(t, section, "系統角色")
	assert.Contains(t, section, "規則怪談")
	assert.Contains(t, section, "500-1200 字")
	assert.Contains(t, section, "張力指令")
}

// TestBuildStateSection tests state section building
func TestBuildStateSection(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(75)
	gameState.SetSAN(60)
	gameState.CurrentScene = "地下室"

	req := &ContentRequest{
		Beat:       3,
		GameState:  gameState,
		Difficulty: "hard",
	}

	section := agent.buildStateSection(req)

	assert.Contains(t, section, "Beat: 3")
	assert.Contains(t, section, "HP: 75 / 100")
	assert.Contains(t, section, "SAN: 60 / 100")
	assert.Contains(t, section, "場景: 地下室")
	assert.Contains(t, section, "難度: hard")
}

// TestBuildTensionSection tests tension section building
func TestBuildTensionSection(t *testing.T) {
	agent := createTestNarrationAgent(t)

	directive := engine.GenerateDirective(engine.TensionLevelHigh)
	section := agent.buildTensionSection(directive)

	assert.Contains(t, section, "張力指令")
	assert.Contains(t, section, directive.Instruction)
	assert.Contains(t, section, "等級")
	assert.Contains(t, section, "指令")
	assert.Contains(t, section, "字數範圍")
}

// TestBuildRulesSection tests rules section building
func TestBuildRulesSection(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name          string
		rules         []*engine.ActiveRule
		warnings      map[string]int
		expectContain []string
	}{
		{
			name:          "no rules",
			rules:         []*engine.ActiveRule{},
			warnings:      map[string]int{},
			expectContain: []string{"活躍規則", "無活躍規則"},
		},
		{
			name: "with rules and warnings",
			rules: []*engine.ActiveRule{
				{ID: "rule-001", Name: "不要在夜晚開燈"},
				{ID: "rule-002", Name: "不要直視鏡子"},
			},
			warnings: map[string]int{
				"rule-001": 1,
				"rule-002": 0,
			},
			expectContain: []string{
				"活躍規則",
				"不要在夜晚開燈",
				"rule-001",
				"警告次數: 1",
				"不要直視鏡子",
				"rule-002",
				"警告次數: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.ActiveRules = tt.rules
			gameState.RuleWarnings = tt.warnings

			section := agent.buildRulesSection(gameState)

			for _, expect := range tt.expectContain {
				assert.Contains(t, section, expect)
			}
		})
	}
}

// TestBuildNPCsSection tests NPCs section building
func TestBuildNPCsSection(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name          string
		npcs          map[string]*engine.NPCState
		expectContain []string
	}{
		{
			name:          "no NPCs",
			npcs:          map[string]*engine.NPCState{},
			expectContain: []string{"活躍 NPCs", "無活躍 NPCs"},
		},
		{
			name: "with NPCs",
			npcs: map[string]*engine.NPCState{
				"npc-001": {ID: "npc-001", Name: "護士"},
				"npc-002": {ID: "npc-002", Name: "醫生"},
			},
			expectContain: []string{
				"活躍 NPCs",
				"護士",
				"npc-001",
				"醫生",
				"npc-002",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.NPCStates = tt.npcs

			section := agent.buildNPCsSection(gameState)

			for _, expect := range tt.expectContain {
				assert.Contains(t, section, expect)
			}
		})
	}
}

// TestBuildContextSection tests context section building
func TestBuildContextSection(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name          string
		lastChoice    string
		judgeResult   *JudgeResult
		expectContain []string
	}{
		{
			name:       "game start - no choice",
			lastChoice: "",
			judgeResult: nil,
			expectContain: []string{
				"玩家上下文",
				"遊戲開始",
			},
		},
		{
			name:       "with choice - no judge result",
			lastChoice: "探索房間",
			judgeResult: nil,
			expectContain: []string{
				"玩家上下文",
				"探索房間",
			},
		},
		{
			name:       "with choice and judge result",
			lastChoice: "開燈",
			judgeResult: &JudgeResult{
				ViolatedRules: []string{"rule-001"},
				ImpactLevel:   "moderate",
				HPDelta:       -10,
				SANDelta:      -5,
				Reason:        "違反了夜晚開燈規則",
			},
			expectContain: []string{
				"玩家上下文",
				"開燈",
				"判定結果",
				"rule-001",
				"moderate",
				"HP 變化: -10",
				"SAN 變化: -5",
				"違反了夜晚開燈規則",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ContentRequest{
				Beat:             1,
				GameState:        engine.NewGameStateV2(),
				LastPlayerChoice: tt.lastChoice,
				JudgeResult:      tt.judgeResult,
			}

			section := agent.buildContextSection(req)

			for _, expect := range tt.expectContain {
				assert.Contains(t, section, expect)
			}
		})
	}
}

// TestBuildConstraintsSection tests constraints section building
func TestBuildConstraintsSection(t *testing.T) {
	agent := createTestNarrationAgent(t)
	section := agent.buildConstraintsSection()

	assert.Contains(t, section, "輸出要求")
	assert.Contains(t, section, "500-1200 字")
	assert.Contains(t, section, "恐怖")
	assert.Contains(t, section, "繁體中文")
	assert.Contains(t, section, "純文本敘事")
}

// TestInvokeContent_RuleHintsGeneration tests Rule Hints integration
// AC #5: Rule Hints 生成整合
func TestInvokeContent_RuleHintsGeneration(t *testing.T) {
	tests := []struct {
		name            string
		difficulty      string
		violatedRules   []string
		existingWarnings map[string]int
		expectHintCount int
		expectContains  string
	}{
		{
			name:            "easy difficulty - first warning",
			difficulty:      "easy",
			violatedRules:   []string{"rule-001"},
			existingWarnings: map[string]int{},
			expectHintCount: 1,
			expectContains:  "感覺", // Easy hint: "你感覺這樣做可能不太對"
		},
		{
			name:            "easy difficulty - second warning",
			difficulty:      "easy",
			violatedRules:   []string{"rule-001"},
			existingWarnings: map[string]int{"rule-001": 1},
			expectHintCount: 1,
			expectContains:  "警告", // Second warning more explicit
		},
		{
			name:            "easy difficulty - max warnings reached",
			difficulty:      "easy",
			violatedRules:   []string{"rule-001"},
			existingWarnings: map[string]int{"rule-001": 2}, // Already 2 warnings
			expectHintCount: 0, // No more hints
		},
		{
			name:            "hard difficulty - first warning",
			difficulty:      "hard",
			violatedRules:   []string{"rule-002"},
			existingWarnings: map[string]int{},
			expectHintCount: 1,
			expectContains:  "注意", // Hard hint: subtle
		},
		{
			name:            "hell difficulty - no warnings",
			difficulty:      "hell",
			violatedRules:   []string{"rule-003"},
			existingWarnings: map[string]int{},
			expectHintCount: 0, // Hell gives no hints
		},
		{
			name:            "multiple violated rules",
			difficulty:      "easy",
			violatedRules:   []string{"rule-001", "rule-002"},
			existingWarnings: map[string]int{},
			expectHintCount: 2, // Hint for each rule
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := createTestNarrationAgent(t)

			// Setup game state with active rules
			gameState := engine.NewGameStateV2()
			gameState.ActiveRules = []*engine.ActiveRule{
				{ID: "rule-001", Name: "不要在夜晚開燈"},
				{ID: "rule-002", Name: "不要直視鏡子"},
				{ID: "rule-003", Name: "不要回應呼喊"},
			}
			// Copy existing warnings to avoid map reference issues
			gameState.RuleWarnings = make(map[string]int)
			for k, v := range tt.existingWarnings {
				gameState.RuleWarnings[k] = v
			}

			// Calculate expected warning counts BEFORE running InvokeContent
			expectedWarningCounts := make(map[string]int)
			if tt.expectHintCount > 0 {
				for _, ruleID := range tt.violatedRules {
					expectedWarningCounts[ruleID] = tt.existingWarnings[ruleID] + 1
				}
			}

			req := &ContentRequest{
				Beat:       1,
				GameState:  gameState,
				Difficulty: tt.difficulty,
				JudgeResult: &JudgeResult{
					ViolatedRules: tt.violatedRules,
					ImpactLevel:   "minor",
				},
			}

			ctx := context.Background()
			resp, err := agent.InvokeContent(ctx, req)

			require.NoError(t, err)
			assert.NotNil(t, resp)

			// Check hint count
			assert.Len(t, resp.RuleHints, tt.expectHintCount,
				"Expected %d hints, got %d", tt.expectHintCount, len(resp.RuleHints))

			// Check hint content if hints were generated
			if tt.expectHintCount > 0 && tt.expectContains != "" {
				found := false
				for _, hint := range resp.RuleHints {
					if strings.Contains(hint, tt.expectContains) {
						found = true
						break
					}
				}
				assert.True(t, found, "Hint should contain '%s'", tt.expectContains)
			}

			// Verify warning count was incremented
			for ruleID, expectedCount := range expectedWarningCounts {
				assert.Equal(t, expectedCount, gameState.RuleWarnings[ruleID],
					"Warning count for %s should be incremented", ruleID)
			}
		})
	}
}

// TestInvokeContent_RuleHintsWithoutViolation tests no hints when no rules violated
func TestInvokeContent_RuleHintsWithoutViolation(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	req := &ContentRequest{
		Beat:       1,
		GameState:  gameState,
		Difficulty: "easy",
		JudgeResult: &JudgeResult{
			ViolatedRules: []string{}, // No violations
			ImpactLevel:   "none",
		},
	}

	ctx := context.Background()
	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.Empty(t, resp.RuleHints, "Should have no hints when no rules violated")
}

// TestInvokeContent_RuleHintsNoJudgeResult tests no hints when no JudgeResult
func TestInvokeContent_RuleHintsNoJudgeResult(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	req := &ContentRequest{
		Beat:        1,
		GameState:   gameState,
		Difficulty:  "easy",
		JudgeResult: nil, // No judge result
	}

	ctx := context.Background()
	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.Empty(t, resp.RuleHints, "Should have no hints when JudgeResult is nil")
}

// TestGenerateRuleHints tests the generateRuleHints method directly
func TestGenerateRuleHints(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	gameState.ActiveRules = []*engine.ActiveRule{
		{ID: "rule-001", Name: "不要在夜晚開燈"},
	}
	gameState.RuleWarnings = map[string]int{}

	req := &ContentRequest{
		Beat:       1,
		GameState:  gameState,
		Difficulty: "easy",
		JudgeResult: &JudgeResult{
			ViolatedRules: []string{"rule-001"},
		},
	}

	hints := agent.generateRuleHints(req)

	assert.Len(t, hints, 1, "Should generate 1 hint")
	assert.NotEmpty(t, hints[0], "Hint should not be empty")
	assert.Equal(t, 1, gameState.RuleWarnings["rule-001"], "Warning count should be incremented")
}

// TestGetRuleDescription tests the getRuleDescription method
func TestGetRuleDescription(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	gameState.ActiveRules = []*engine.ActiveRule{
		{ID: "rule-001", Name: "不要在夜晚開燈"},
		{ID: "rule-002", Name: "不要直視鏡子"},
	}

	tests := []struct {
		ruleID   string
		expected string
	}{
		{"rule-001", "不要在夜晚開燈"},
		{"rule-002", "不要直視鏡子"},
		{"rule-999", ""}, // Not found
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			desc := agent.getRuleDescription(gameState, tt.ruleID)
			assert.Equal(t, tt.expected, desc)
		})
	}
}

// TestHandleSceneTransition tests scene transition description generation
func TestHandleSceneTransition(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name           string
		fromScene      string
		toScene        string
		transitionType string
		expectContains []string
	}{
		{
			name:           "spatial transition",
			fromScene:      "室內走廊",
			toScene:        "戶外庭院",
			transitionType: "spatial",
			expectContains: []string{
				"場景轉換發生了",
				"室內走廊",
				"戶外庭院",
				"環境的變化",
				"空氣的溫度",
			},
		},
		{
			name:           "temporal transition",
			fromScene:      "白天",
			toScene:        "夜晚",
			transitionType: "temporal",
			expectContains: []string{
				"時間流逝",
				"白天",
				"夜晚",
				"時間的推移",
				"光線的角度",
			},
		},
		{
			name:           "atmospheric transition",
			fromScene:      "安全房間",
			toScene:        "危險區域",
			transitionType: "atmospheric",
			expectContains: []string{
				"氛圍驟然改變",
				"安全房間",
				"危險區域",
				"氣氛發生了劇烈的變化",
				"本能在尖叫",
			},
		},
		{
			name:           "default transition",
			fromScene:      "A場景",
			toScene:        "B場景",
			transitionType: "unknown",
			expectContains: []string{
				"A場景",
				"B場景",
				"環境發生了變化",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := agent.HandleSceneTransition(tt.fromScene, tt.toScene, tt.transitionType)

			// Check description is not empty
			assert.NotEmpty(t, description, "Scene transition description should not be empty")

			// Check description length is reasonable (default case is shorter than full transitions)
			// Default transition: ~60-80 chars, Full transitions: 100-150 chars
			runeCount := len([]rune(description))
			if tt.transitionType == "unknown" {
				assert.GreaterOrEqual(t, runeCount, 50, "Default description should be at least 50 characters")
			} else {
				assert.GreaterOrEqual(t, runeCount, 80, "Description should be at least 80 characters")
			}

			// Check expected content is present
			for _, expected := range tt.expectContains {
				assert.Contains(t, description, expected,
					"Description should contain '%s'", expected)
			}
		})
	}
}

// TestHandleKeyPlotPoint tests key plot point description generation
func TestHandleKeyPlotPoint(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name           string
		plotType       PlotPointType
		beat           int
		expectContains []string
		minLength      int
	}{
		{
			name:     "inciting incident",
			plotType: PlotPointIncitingIncident,
			beat:     10,
			expectContains: []string{
				"一切都在這一刻改變了",
				"事情遠比你想像的要嚴重",
				"你已經被捲入了",
				"無法逃脫的漩渦",
			},
			minLength: 150,
		},
		{
			name:     "midpoint",
			plotType: PlotPointMidpoint,
			beat:     25,
			expectContains: []string{
				"真相的一角浮現了",
				"你所知道的只是冰山一角",
				"真正的威脅比你想像的要可怕得多",
				"你別無選擇",
			},
			minLength: 150,
		},
		{
			name:     "second plot point",
			plotType: PlotPointSecondPlotPoint,
			beat:     40,
			expectContains: []string{
				"這是最黑暗的時刻",
				"一切都崩潰了",
				"絕望",
				"最後的考驗",
			},
			minLength: 150,
		},
		{
			name:     "climax",
			plotType: PlotPointClimax,
			beat:     50,
			expectContains: []string{
				"最終的時刻到來了",
				"所有的線索",
				"最後的對抗",
				"決定一切的時刻",
			},
			minLength: 150,
		},
		{
			name:           "unknown plot type",
			plotType:       PlotPointType("unknown"),
			beat:           15,
			expectContains: []string{"關鍵時刻降臨", "Beat 15"},
			minLength:      20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := agent.HandleKeyPlotPoint(tt.plotType, tt.beat)

			// Check description is not empty
			assert.NotEmpty(t, description, "Plot point description should not be empty")

			// Check description length (should be 150-250 字 for major plot points)
			runeCount := len([]rune(description))
			assert.GreaterOrEqual(t, runeCount, tt.minLength,
				"Description should be at least %d characters", tt.minLength)

			// Check expected content is present
			for _, expected := range tt.expectContains {
				assert.Contains(t, description, expected,
					"Description should contain '%s'", expected)
			}
		})
	}
}

// TestHandleHallucinationChoice tests hallucination experience generation
func TestHandleHallucinationChoice(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name             string
		currentSAN       int
		expectedSeverity HallucinationSeverity
		expectedPenalty  int
		expectContains   []string
		minLength        int
	}{
		{
			name:             "mild hallucination (SAN 50)",
			currentSAN:       50,
			expectedSeverity: HallucinationMild,
			expectedPenalty:  -2,
			expectContains: []string{
				"世界開始扭曲",
				"牆壁彷彿在呼吸",
				"燈光變得詭異",
				"你晃了晃頭",
				"眼前的景象開始消失",
			},
			minLength: 200,
		},
		{
			name:             "mild hallucination (SAN 40)",
			currentSAN:       40,
			expectedSeverity: HallucinationMild,
			expectedPenalty:  -2,
			expectContains: []string{
				"世界開始扭曲",
				"你晃了晃頭",
			},
			minLength: 200,
		},
		{
			name:             "moderate hallucination (SAN 30)",
			currentSAN:       30,
			expectedSeverity: HallucinationModerate,
			expectedPenalty:  -3,
			expectContains: []string{
				"現實開始崩解",
				"不存在的聲音",
				"時間感也變得混亂了",
				"影子開始移動",
				"你晃了晃頭",
			},
			minLength: 200,
		},
		{
			name:             "moderate hallucination (SAN 20)",
			currentSAN:       20,
			expectedSeverity: HallucinationModerate,
			expectedPenalty:  -3,
			expectContains: []string{
				"現實開始崩解",
				"你晃了晃頭",
			},
			minLength: 200,
		},
		{
			name:             "severe hallucination (SAN 15)",
			currentSAN:       15,
			expectedSeverity: HallucinationSevere,
			expectedPenalty:  -5,
			expectContains: []string{
				"一切都不真實了",
				"世界在眼前徹底瓦解",
				"有另一個你在說話",
				"理智徹底崩潰的瞬間",
				"你晃了晃頭",
			},
			minLength: 200,
		},
		{
			name:             "severe hallucination (SAN 5)",
			currentSAN:       5,
			expectedSeverity: HallucinationSevere,
			expectedPenalty:  -5,
			expectContains: []string{
				"一切都不真實了",
				"你晃了晃頭",
			},
			minLength: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description, sanPenalty := agent.HandleHallucinationChoice(tt.currentSAN)

			// Check description is not empty
			assert.NotEmpty(t, description, "Hallucination description should not be empty")

			// Check description length (should be 200-300 字 for main text + ~100 for ending)
			// Total: 300-550 characters is reasonable
			runeCount := len([]rune(description))
			assert.GreaterOrEqual(t, runeCount, tt.minLength,
				"Description should be at least %d characters", tt.minLength)
			assert.LessOrEqual(t, runeCount, 550,
				"Description should not exceed 550 characters")

			// Check SAN penalty matches expected severity
			assert.Equal(t, tt.expectedPenalty, sanPenalty,
				"SAN penalty should match expected severity")

			// Check expected content is present
			for _, expected := range tt.expectContains {
				assert.Contains(t, description, expected,
					"Description should contain '%s'", expected)
			}

			// Check hallucination ending is always present
			assert.Contains(t, description, "你晃了晃頭",
				"Description should contain standard hallucination ending")
		})
	}
}

// TestInvokeContent_Hallucination tests hallucination integration in InvokeContent
func TestInvokeContent_Hallucination(t *testing.T) {
	agent := createTestNarrationAgent(t)

	tests := []struct {
		name              string
		currentSAN        int
		initialSANDelta   int
		expectedTotalDelta int
		expectInNarrative []string
	}{
		{
			name:              "mild hallucination",
			currentSAN:        50,
			initialSANDelta:   -5,
			expectedTotalDelta: -7, // -5 initial + -2 hallucination penalty
			expectInNarrative: []string{
				"世界開始扭曲",
				"你晃了晃頭",
			},
		},
		{
			name:              "moderate hallucination",
			currentSAN:        30,
			initialSANDelta:   -10,
			expectedTotalDelta: -13, // -10 initial + -3 hallucination penalty
			expectInNarrative: []string{
				"現實開始崩解",
				"你晃了晃頭",
			},
		},
		{
			name:              "severe hallucination",
			currentSAN:        15,
			initialSANDelta:   -8,
			expectedTotalDelta: -13, // -8 initial + -5 hallucination penalty
			expectInNarrative: []string{
				"一切都不真實了",
				"你晃了晃頭",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetSAN(tt.currentSAN)

			req := &ContentRequest{
				Beat:             5,
				GameState:        gameState,
				LastPlayerChoice: "選擇了幻覺選項",
				Difficulty:       "normal",
				JudgeResult: &JudgeResult{
					ViolatedRules:   []string{},
					ImpactLevel:     "moderate",
					HPDelta:         0,
					SANDelta:        tt.initialSANDelta,
					IsHallucination: true, // This triggers hallucination handling
					Reason:          "幻覺體驗",
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := agent.InvokeContent(ctx, req)

			// Should not error
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Check SAN change includes hallucination penalty
			assert.Equal(t, tt.expectedTotalDelta, resp.SANChange,
				"SANChange should include hallucination penalty")

			// Check hallucination description is in narrative
			for _, expected := range tt.expectInNarrative {
				assert.Contains(t, resp.MainNarrative, expected,
					"Narrative should contain hallucination description")
			}

			// Check death trigger if SAN drops to 0 or below
			finalSAN := tt.currentSAN + tt.expectedTotalDelta
			if finalSAN <= 0 {
				assert.True(t, resp.DeathTriggered,
					"Death should be triggered when SAN <= 0")
			}
		})
	}
}

// TestInvokeContent_HallucinationWithDeath tests hallucination causing death
func TestInvokeContent_HallucinationWithDeath(t *testing.T) {
	agent := createTestNarrationAgent(t)

	gameState := engine.NewGameStateV2()
	gameState.SetSAN(5) // Very low SAN

	req := &ContentRequest{
		Beat:             10,
		GameState:        gameState,
		LastPlayerChoice: "選擇了幻覺選項",
		Difficulty:       "normal",
		JudgeResult: &JudgeResult{
			ViolatedRules:   []string{},
			ImpactLevel:     "major",
			HPDelta:         0,
			SANDelta:        -3,             // Initial penalty
			IsHallucination: true,           // Will add -5 more (severe hallucination)
			Reason:          "嚴重幻覺體驗",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Check SAN change: -3 initial + -5 hallucination = -8
	assert.Equal(t, -8, resp.SANChange, "SANChange should be -8")

	// Check death is triggered (5 + (-8) = -3, which is <= 0)
	assert.True(t, resp.DeathTriggered,
		"Death should be triggered when SAN drops to -3")

	// Check severe hallucination description is present
	assert.Contains(t, resp.MainNarrative, "一切都不真實了",
		"Should contain severe hallucination description")
}

