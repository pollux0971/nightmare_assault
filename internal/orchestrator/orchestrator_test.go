package orchestrator

import (
	"context"
	"testing"
	"time"
)

// Test Task 1: Orchestrator 主結構
func TestNewOrchestrator(t *testing.T) {
	orch := NewOrchestrator()

	if orch == nil {
		t.Fatal("NewOrchestrator() returned nil")
	}

	// 驗證初始 Phase
	if orch.GetCurrentPhase() != PhaseGenesis {
		t.Errorf("Expected initial phase to be Genesis, got %v", orch.GetCurrentPhase())
	}

	// 驗證組件已初始化（非 nil）
	if orch.storyBible == nil {
		t.Error("storyBible should be initialized")
	}

	if orch.gameState == nil {
		t.Error("gameState should be initialized")
	}
}

// Test Task 2: Phase 路由邏輯
func TestOrchestrator_PhaseTransitions(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 初始應該是 Genesis
	if phase := orch.GetCurrentPhase(); phase != PhaseGenesis {
		t.Fatalf("Expected PhaseGenesis, got %v", phase)
	}

	// 運行 Genesis 後應切換到 GameLoop
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("RunPhaseGenesis failed: %v", err)
	}

	if phase := orch.GetCurrentPhase(); phase != PhaseGameLoop {
		t.Errorf("After Genesis, expected PhaseGameLoop, got %v", phase)
	}

	// 運行 GameLoop（不觸發收束條件）
	_, err = orch.RunGameLoopTurn(ctx, "test choice")
	if err != nil {
		t.Fatalf("RunGameLoopTurn failed: %v", err)
	}

	// 應該仍在 GameLoop
	if phase := orch.GetCurrentPhase(); phase != PhaseGameLoop {
		t.Errorf("Should still be in GameLoop, got %v", phase)
	}
}

func TestOrchestrator_RunPhaseGenesis(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("RunPhaseGenesis failed: %v", err)
	}

	if result == nil {
		t.Fatal("GenesisResult should not be nil")
	}

	// 驗證 StoryBible 已更新
	if orch.storyBible.WorldView == "" {
		t.Error("WorldView should be set after Genesis")
	}

	// 驗證 Phase 切換
	if orch.GetCurrentPhase() != PhaseGameLoop {
		t.Error("Phase should transition to GameLoop after Genesis")
	}
}

func TestOrchestrator_RunGameLoopTurn(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 先運行 Genesis
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 運行一個 Game Loop turn
	result, err := orch.RunGameLoopTurn(ctx, "explore the room")
	if err != nil {
		t.Fatalf("RunGameLoopTurn failed: %v", err)
	}

	if result == nil {
		t.Fatal("TurnResult should not be nil")
	}

	if result.Story == "" {
		t.Error("Story text should not be empty")
	}

	if len(result.Choices) == 0 {
		t.Error("Choices should not be empty")
	}

	// 驗證 Beat 增加
	if orch.gameState.GetCurrentBeat() == 0 {
		t.Error("CurrentBeat should increment after turn")
	}
}

func TestOrchestrator_RunPhaseConvergence(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 先運行 Genesis
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 手動切換到 Convergence phase
	orch.currentPhase = PhaseConvergence

	// 運行 Convergence
	result, err := orch.RunPhaseConvergence(ctx)
	if err != nil {
		t.Fatalf("RunPhaseConvergence failed: %v", err)
	}

	if result == nil {
		t.Fatal("EndingResult should not be nil")
	}

	if result.Story == "" {
		t.Error("Ending story should not be empty")
	}
}

// Test Task 3-5: 組件整合（接口測試）
func TestOrchestrator_ComponentsInitialized(t *testing.T) {
	orch := NewOrchestrator()

	// Logic 層組件
	if orch.tensionEngine == nil {
		t.Error("tensionEngine should be initialized")
	}
	if orch.seedManager == nil {
		t.Error("seedManager should be initialized")
	}
	if orch.contextMgr == nil {
		t.Error("contextMgr should be initialized")
	}
	if orch.ruleEngine == nil {
		t.Error("ruleEngine should be initialized")
	}
	if orch.stateMgr == nil {
		t.Error("stateMgr should be initialized")
	}

	// Agent 層組件
	if orch.narrationAgent == nil {
		t.Error("narrationAgent should be initialized")
	}
	if orch.choiceAgent == nil {
		t.Error("choiceAgent should be initialized")
	}
	if orch.judgeAgent == nil {
		t.Error("judgeAgent should be initialized")
	}
	if orch.seedAgent == nil {
		t.Error("seedAgent should be initialized")
	}
	if orch.npcAgent == nil {
		t.Error("npcAgent should be initialized")
	}

	// 模板庫
	if orch.templateLib == nil {
		t.Error("templateLib should be initialized")
	}
}

// Test 錯誤處理
func TestOrchestrator_ContextCancellation(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消
	cancel()

	_, err := orch.RunPhaseGenesis(ctx)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

// Test 狀態一致性
func TestOrchestrator_StateConsistency(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 運行 Genesis
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("Genesis failed: %v", err)
	}

	// GameState 應該與 StoryBible 一致
	if orch.gameState.GameID == "" {
		t.Error("GameState should have GameID after Genesis")
	}

	if orch.storyBible.WorldView == "" {
		t.Error("StoryBible should have WorldView after Genesis")
	}

	// 運行多個 turns
	for i := 0; i < 3; i++ {
		_, err := orch.RunGameLoopTurn(ctx, "test action")
		if err != nil {
			t.Fatalf("Turn %d failed: %v", i, err)
		}
	}

	// Beat 應該正確遞增
	if orch.gameState.GetCurrentBeat() != 3 {
		t.Errorf("Expected beat=3 after 3 turns, got %d", orch.gameState.GetCurrentBeat())
	}
}

// NEW: Test shouldConverge 方法
func TestShouldConverge(t *testing.T) {
	tests := []struct {
		name           string
		seedProgress   float64
		tension        int
		currentBeat    int
		want           bool
		wantReason     string
	}{
		// 單一條件觸發
		{
			name:         "seed progress >= 80% triggers",
			seedProgress: 0.80,
			tension:      50,
			currentBeat:  10,
			want:         true,
			wantReason:   "seed_progress",
		},
		{
			name:         "seed progress 79% does not trigger",
			seedProgress: 0.79,
			tension:      50,
			currentBeat:  10,
			want:         false,
		},
		{
			name:         "tension >= 95 triggers",
			seedProgress: 0.50,
			tension:      95,
			currentBeat:  10,
			want:         true,
			wantReason:   "tension",
		},
		{
			name:         "tension 94 does not trigger",
			seedProgress: 0.50,
			tension:      94,
			currentBeat:  10,
			want:         false,
		},
		{
			name:         "beat >= 20 triggers",
			seedProgress: 0.50,
			tension:      50,
			currentBeat:  20,
			want:         true,
			wantReason:   "beat_limit",
		},
		{
			name:         "beat 19 does not trigger",
			seedProgress: 0.50,
			tension:      50,
			currentBeat:  19,
			want:         false,
		},

		// 邊界條件
		{
			name:         "exactly 80% progress",
			seedProgress: 0.80,
			tension:      0,
			currentBeat:  0,
			want:         true,
		},
		{
			name:         "exactly tension 95",
			seedProgress: 0.0,
			tension:      95,
			currentBeat:  0,
			want:         true,
		},
		{
			name:         "exactly beat 20",
			seedProgress: 0.0,
			tension:      0,
			currentBeat:  20,
			want:         true,
		},

		// 組合條件
		{
			name:         "all conditions met",
			seedProgress: 0.90,
			tension:      100,
			currentBeat:  25,
			want:         true,
		},
		{
			name:         "no conditions met",
			seedProgress: 0.50,
			tension:      50,
			currentBeat:  10,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orch := NewOrchestrator()

			// 配置 Mock
			mockSeedMgr := orch.seedManager.(*MockSeedManager)
			mockSeedMgr.SetGlobalProgress(tt.seedProgress)

			// 設置 gameState
			orch.gameState.Tension.Value = tt.tension
			for i := 0; i < tt.currentBeat; i++ {
				orch.gameState.IncrementBeat()
			}

			// 測試
			got := orch.shouldConverge()

			if got != tt.want {
				t.Errorf("shouldConverge() = %v, want %v", got, tt.want)
			}
		})
	}
}

// NEW: Test convergence transition in GameLoop
func TestOrchestrator_ConvergenceTransition(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 執行 Genesis
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("Genesis failed: %v", err)
	}

	// 確認在 GameLoop
	if orch.GetCurrentPhase() != PhaseGameLoop {
		t.Fatal("Should be in GameLoop after Genesis")
	}

	// 模擬觸發收束條件（設置高張力）
	orch.gameState.Tension.Value = ConvergenceTensionThreshold

	// 執行一輪 GameLoop
	_, err = orch.RunGameLoopTurn(ctx, "final choice")
	if err != nil {
		t.Fatalf("GameLoop turn failed: %v", err)
	}

	// 應該自動切換到 Convergence
	if orch.GetCurrentPhase() != PhaseConvergence {
		t.Errorf("Should transition to Convergence when tension >= %d, got %v",
			ConvergenceTensionThreshold, orch.GetCurrentPhase())
	}
}

// NEW: Test StateManager applies changes correctly
func TestOrchestrator_StateManagerIntegration(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// 執行 Genesis
	_, err := orch.RunPhaseGenesis(ctx)
	if err != nil {
		t.Fatalf("Genesis failed: %v", err)
	}

	initialTension := orch.gameState.Tension.Value

	// 執行一個 GameLoop turn
	_, err = orch.RunGameLoopTurn(ctx, "test choice")
	if err != nil {
		t.Fatalf("GameLoop failed: %v", err)
	}

	// 驗證 Tension 有變化（MockTensionEngine 每次 +5）
	newTension := orch.gameState.Tension.Value
	if newTension <= initialTension {
		t.Errorf("Tension should increase after turn, was %d, now %d", initialTension, newTension)
	}

	// 驗證 SAN 有變化（MockJudgeAgent 每次 -1）
	if orch.gameState.SAN >= 100 {
		t.Errorf("SAN should decrease after turn, got %d", orch.gameState.SAN)
	}
}
