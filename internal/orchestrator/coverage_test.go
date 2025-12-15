package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestGamePhase_String tests the String() method for all phases.
func TestGamePhase_String(t *testing.T) {
	tests := []struct {
		phase    GamePhase
		expected string
	}{
		{PhaseGenesis, "Genesis"},
		{PhaseGameLoop, "GameLoop"},
		{PhaseConvergence, "Convergence"},
		{GamePhase(999), "Unknown"}, // Test unknown phase
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.phase.String()
			if got != tt.expected {
				t.Errorf("GamePhase(%d).String() = %q, want %q", tt.phase, got, tt.expected)
			}
		})
	}
}

// TestRunPhaseGenesis_ContextCancelled tests context cancellation at various points.
func TestRunPhaseGenesis_ContextCancelled(t *testing.T) {
	// Test context error at the beginning
	t.Run("context_cancelled_at_start", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := orch.RunPhaseGenesis(ctx)
		if err == nil {
			t.Error("Expected error when context is cancelled, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	})

	// Test context timeout during skeleton generation
	t.Run("context_timeout_during_skeleton", func(t *testing.T) {
		orch := NewOrchestrator()

		// Create a mock narration agent that delays
		slowNarration := &SlowNarrationAgent{delay: 200 * time.Millisecond}
		orch.narrationAgent = slowNarration

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := orch.RunPhaseGenesis(ctx)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	// Test skeleton generation error
	t.Run("skeleton_generation_error", func(t *testing.T) {
		orch := NewOrchestrator()
		orch.narrationAgent = &ErrorNarrationAgent{err: errors.New("skeleton failed")}

		_, err := orch.RunPhaseGenesis(context.Background())
		if err == nil {
			t.Error("Expected skeleton generation error, got nil")
		}
	})

	// Test seed generation error
	t.Run("seed_generation_error", func(t *testing.T) {
		orch := NewOrchestrator()
		orch.seedAgent = &ErrorSeedAgent{err: errors.New("seed generation failed")}

		_, err := orch.RunPhaseGenesis(context.Background())
		if err == nil {
			t.Error("Expected seed generation error, got nil")
		}
	})

	// Test NPC generation error
	t.Run("npc_generation_error", func(t *testing.T) {
		orch := NewOrchestrator()
		orch.npcAgent = &ErrorNPCAgent{err: errors.New("NPC generation failed")}

		_, err := orch.RunPhaseGenesis(context.Background())
		if err == nil {
			t.Error("Expected NPC generation error, got nil")
		}
	})

	// Test context cancelled after seeds
	t.Run("context_cancelled_after_seeds", func(t *testing.T) {
		orch := NewOrchestrator()

		// Use slow seed agent
		orch.seedAgent = &SlowSeedAgent{delay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Millisecond)
		defer cancel()

		_, err := orch.RunPhaseGenesis(ctx)
		if err == nil {
			t.Error("Expected context timeout error, got nil")
		}
	})

	// Test context cancelled after NPCs
	t.Run("context_cancelled_after_npcs", func(t *testing.T) {
		orch := NewOrchestrator()

		// Use slow NPC agent
		orch.npcAgent = &SlowNPCAgent{delay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Millisecond)
		defer cancel()

		_, err := orch.RunPhaseGenesis(ctx)
		if err == nil {
			t.Error("Expected context timeout error, got nil")
		}
	})
}

// TestRunGameLoopTurn_ContextCancelled tests context handling in game loop.
func TestRunGameLoopTurn_ContextCancelled(t *testing.T) {
	t.Run("context_cancelled_before_turn", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Cancel context before turn
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected error when context is cancelled, got nil")
		}
	})
}

// TestRunGameLoopTurn_ErrorPaths tests error handling in game loop.
func TestRunGameLoopTurn_ErrorPaths(t *testing.T) {
	t.Run("judge_returns_error", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Replace judge with error-returning mock
		orch.judgeAgent = &ErrorJudgeAgent{err: errors.New("judge failed")}

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected error from judge, got nil")
		}
	})

	t.Run("state_manager_returns_error", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Replace state manager with error-returning mock
		orch.stateMgr = &ErrorStateManager{err: errors.New("state update failed")}

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected error from state manager, got nil")
		}
	})

	t.Run("narration_returns_error", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Replace narration with error-returning mock
		orch.narrationAgent = &ErrorNarrationAgent{err: errors.New("narration failed")}

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected error from narration, got nil")
		}
	})

	t.Run("choice_generation_returns_error", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Replace choice agent with error-returning mock
		orch.choiceAgent = &ErrorChoiceAgent{err: errors.New("choice generation failed")}

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected error from choice generation, got nil")
		}
	})

	t.Run("seed_harvest_marks_seed_revealed", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Create a seed manager that returns harvest instructions
		harvestMgr := &HarvestSeedManager{
			harvests: []*seed.HarvestInstruction{
				{SeedID: "test-seed-001"},
			},
		}
		orch.seedManager = harvestMgr

		result, err := orch.RunGameLoopTurn(ctx, "test")
		if err != nil {
			t.Fatalf("Turn failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Verify MarkSeedRevealed was called
		if !harvestMgr.markCalled {
			t.Error("Expected MarkSeedRevealed to be called")
		}
	})

	t.Run("context_cancelled_after_narration", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Use a slow narration agent
		orch.narrationAgent = &SlowContentNarrationAgent{delay: 100 * time.Millisecond}

		// Context will timeout during narration
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected context timeout error, got nil")
		}
	})

	t.Run("context_cancelled_after_choices", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Use a slow choice agent
		orch.choiceAgent = &SlowChoiceAgent{delay: 100 * time.Millisecond}

		// Context will timeout during choice generation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		_, err = orch.RunGameLoopTurn(ctx, "test")
		if err == nil {
			t.Error("Expected context timeout error, got nil")
		}
	})

	t.Run("mark_seed_revealed_error_non_fatal", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Create a seed manager that returns harvest instructions but fails on mark
		errorMgr := &ErrorMarkSeedManager{
			harvests:  []*seed.HarvestInstruction{{SeedID: "test-seed-001"}},
			markError: errors.New("failed to mark seed"),
		}
		orch.seedManager = errorMgr

		// Should not fail - error is logged but non-fatal
		result, err := orch.RunGameLoopTurn(ctx, "test")
		if err != nil {
			t.Errorf("Expected turn to succeed despite mark error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Verify mark was attempted
		if !errorMgr.markCalled {
			t.Error("Expected MarkSeedRevealed to be called")
		}
	})
}

// TestRunPhaseConvergence_ContextCancelled tests convergence phase error handling.
func TestRunPhaseConvergence_ContextCancelled(t *testing.T) {
	t.Run("context_cancelled_at_start", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := orch.RunPhaseConvergence(ctx)
		if err == nil {
			t.Error("Expected error when context is cancelled, got nil")
		}
	})

	t.Run("narration_returns_error", func(t *testing.T) {
		orch := NewOrchestrator()
		ctx := context.Background()

		// Run Genesis first
		_, err := orch.RunPhaseGenesis(ctx)
		if err != nil {
			t.Fatalf("Genesis failed: %v", err)
		}

		// Replace narration with error-returning mock
		orch.narrationAgent = &ErrorNarrationAgent{err: errors.New("ending generation failed")}

		_, err = orch.RunPhaseConvergence(ctx)
		if err == nil {
			t.Error("Expected error from ending generation, got nil")
		}
	})
}

// ============================================================================
// Error-inducing Mock Implementations
// ============================================================================

type SlowNarrationAgent struct {
	delay time.Duration
}

func (s *SlowNarrationAgent) GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error) {
	select {
	case <-time.After(s.delay):
		return &SkeletonResult{WorldView: "test", MainTheme: "test", Setting: "test"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SlowNarrationAgent) GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error) {
	return &OpeningResult{Story: "test"}, nil
}

func (s *SlowNarrationAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error) {
	return &ContentResult{Story: "test"}, nil
}

func (s *SlowNarrationAgent) GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error) {
	return &EndingResult{Story: "test"}, nil
}

type ErrorJudgeAgent struct {
	err error
}

func (e *ErrorJudgeAgent) Judge(ctx context.Context, req JudgeRequest) (*JudgeResult, error) {
	return nil, e.err
}

type ErrorStateManager struct {
	err error
}

func (e *ErrorStateManager) ApplyChanges(changes StateChanges) (*ChangeResult, error) {
	return nil, e.err
}

type ErrorNarrationAgent struct {
	err error
}

func (e *ErrorNarrationAgent) GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error) {
	return nil, e.err
}

func (e *ErrorNarrationAgent) GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error) {
	return nil, e.err
}

func (e *ErrorNarrationAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error) {
	return nil, e.err
}

func (e *ErrorNarrationAgent) GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error) {
	return nil, e.err
}

type ErrorChoiceAgent struct {
	err error
}

func (e *ErrorChoiceAgent) GenerateChoices(ctx context.Context, req ChoiceRequest) (*ChoiceResult, error) {
	return nil, e.err
}

type HarvestSeedManager struct {
	harvests   []*seed.HarvestInstruction
	markCalled bool
}

func (h *HarvestSeedManager) CheckHarvest(currentBeat int) []*seed.HarvestInstruction {
	return h.harvests
}

func (h *HarvestSeedManager) MarkSeedRevealed(seedID string, currentBeat int) error {
	h.markCalled = true
	return nil
}

func (h *HarvestSeedManager) GetGlobalSeedsProgress() float64 {
	return 0.0
}

func (h *HarvestSeedManager) PruneLocalSeedsByScene(sceneID string) []PruneResult {
	return []PruneResult{}
}

func (h *HarvestSeedManager) PruneExpiredLocalSeeds(currentBeat int) []PruneResult {
	return []PruneResult{}
}

type ErrorSeedAgent struct {
	err error
}

func (e *ErrorSeedAgent) GenerateGlobal(ctx context.Context, params agents.GenerateGlobalParams) ([]*seed.GlobalSeed, error) {
	return nil, e.err
}

type ErrorNPCAgent struct {
	err error
}

func (e *ErrorNPCAgent) GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error) {
	return nil, e.err
}

type SlowContentNarrationAgent struct {
	delay time.Duration
}

func (s *SlowContentNarrationAgent) GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error) {
	return &SkeletonResult{WorldView: "test", MainTheme: "test", Setting: "test"}, nil
}

func (s *SlowContentNarrationAgent) GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error) {
	return &OpeningResult{Story: "test"}, nil
}

func (s *SlowContentNarrationAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error) {
	select {
	case <-time.After(s.delay):
		return &ContentResult{Story: "test"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SlowContentNarrationAgent) GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error) {
	return &EndingResult{Story: "test"}, nil
}

type SlowChoiceAgent struct {
	delay time.Duration
}

func (s *SlowChoiceAgent) GenerateChoices(ctx context.Context, req ChoiceRequest) (*ChoiceResult, error) {
	select {
	case <-time.After(s.delay):
		return &ChoiceResult{Choices: []string{"test"}}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type ErrorMarkSeedManager struct {
	harvests   []*seed.HarvestInstruction
	markError  error
	markCalled bool
}

func (e *ErrorMarkSeedManager) CheckHarvest(currentBeat int) []*seed.HarvestInstruction {
	return e.harvests
}

func (e *ErrorMarkSeedManager) MarkSeedRevealed(seedID string, currentBeat int) error {
	e.markCalled = true
	return e.markError
}

func (e *ErrorMarkSeedManager) GetGlobalSeedsProgress() float64 {
	return 0.0
}

func (e *ErrorMarkSeedManager) PruneLocalSeedsByScene(sceneID string) []PruneResult {
	return []PruneResult{}
}

func (e *ErrorMarkSeedManager) PruneExpiredLocalSeeds(currentBeat int) []PruneResult {
	return []PruneResult{}
}

type ContextAwareNarrationAgent struct{}

func (c *ContextAwareNarrationAgent) GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error) {
	// Simulate work before returning
	time.Sleep(1 * time.Millisecond)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &SkeletonResult{WorldView: "test", MainTheme: "test", Setting: "test"}, nil
}

func (c *ContextAwareNarrationAgent) GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error) {
	return &OpeningResult{Story: "test"}, nil
}

func (c *ContextAwareNarrationAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error) {
	return &ContentResult{Story: "test"}, nil
}

func (c *ContextAwareNarrationAgent) GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error) {
	return &EndingResult{Story: "test"}, nil
}

type SlowSeedAgent struct {
	delay time.Duration
}

func (s *SlowSeedAgent) GenerateGlobal(ctx context.Context, params agents.GenerateGlobalParams) ([]*seed.GlobalSeed, error) {
	select {
	case <-time.After(s.delay):
		gs, _ := seed.NewGlobalSeed("test", "test", "test", "test", []seed.ClueTier{})
		return []*seed.GlobalSeed{gs}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type SlowNPCAgent struct {
	delay time.Duration
}

func (s *SlowNPCAgent) GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error) {
	select {
	case <-time.After(s.delay):
		return []*NPCProfile{{ID: "test", Name: "test", Description: "test"}}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
