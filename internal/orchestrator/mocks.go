package orchestrator

import (
	"context"
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ============================================================================
// Mock Logic Layer Implementations
// ============================================================================

// MockTensionEngine is a placeholder implementation.
type MockTensionEngine struct{}

func NewMockTensionEngine() *MockTensionEngine {
	return &MockTensionEngine{}
}

func (m *MockTensionEngine) CalculateDelta(beat int, impact string, current int) int {
	// Simple placeholder: +5 per turn
	return 5
}

func (m *MockTensionEngine) GetDirective(tension int) TensionDirective {
	level := "MEDIUM"
	if tension < 30 {
		level = "LOW"
	} else if tension >= 70 {
		level = "HIGH"
	}

	return TensionDirective{
		Level:             level,
		Instruction:       fmt.Sprintf("Current tension: %d (%s)", tension, level),
		AllowedElements:   []string{"exploration", "dialogue"},
		ForbiddenElements: []string{"combat"},
	}
}

// MockSeedManager is a configurable placeholder implementation.
type MockSeedManager struct {
	globalProgress float64
}

func NewMockSeedManager() *MockSeedManager {
	return &MockSeedManager{
		globalProgress: 0.0,
	}
}

// SetGlobalProgress sets the mock seed progress for testing.
func (m *MockSeedManager) SetGlobalProgress(progress float64) {
	m.globalProgress = progress
}

func (m *MockSeedManager) CheckHarvest(currentBeat int) []*seed.HarvestInstruction {
	// No harvesting in placeholder
	return []*seed.HarvestInstruction{}
}

func (m *MockSeedManager) MarkSeedRevealed(seedID string, currentBeat int) error {
	// No-op in mock - seeds don't actually advance in placeholder
	return nil
}

func (m *MockSeedManager) GetGlobalSeedsProgress() float64 {
	return m.globalProgress
}

// MockContextManager is a placeholder implementation.
type MockContextManager struct{}

func NewMockContextManager() *MockContextManager {
	return &MockContextManager{}
}

func (m *MockContextManager) GetWindow(history []HistoryEntry, tension int) ContextWindow {
	return ContextWindow{
		Summary:       "Game in progress",
		RecentEntries: history,
	}
}

// MockRuleEngine is a placeholder implementation.
type MockRuleEngine struct{}

func NewMockRuleEngine() *MockRuleEngine {
	return &MockRuleEngine{}
}

func (m *MockRuleEngine) CheckViolation(action string) (bool, RuleViolation) {
	// No violations in placeholder
	return false, RuleViolation{}
}

// clamp restricts value to [min, max] range.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// MockStateManager is a placeholder implementation with HP/SAN clamping.
// Enforces 0-100 range for HP, SAN, and Tension values.
type MockStateManager struct {
	gameState *engine.GameStateV2
}

func NewMockStateManager(gameState *engine.GameStateV2) *MockStateManager {
	return &MockStateManager{
		gameState: gameState,
	}
}

func (m *MockStateManager) ApplyChanges(changes StateChanges) {
	if m.gameState == nil {
		return
	}

	// Apply HP changes with clamping to [0, 100]
	if changes.HPDelta != 0 {
		newHP := m.gameState.HP + changes.HPDelta
		m.gameState.HP = clamp(newHP, 0, 100)
	}

	// Apply SAN changes with clamping to [0, 100]
	if changes.SANDelta != 0 {
		newSAN := m.gameState.SAN + changes.SANDelta
		m.gameState.SAN = clamp(newSAN, 0, 100)
	}

	// Apply Tension changes with clamping to [0, 100]
	if changes.TensionDelta != 0 {
		newTension := m.gameState.Tension.Value + changes.TensionDelta
		m.gameState.Tension.Value = clamp(newTension, 0, 100)
	}
}

// ============================================================================
// Mock Agent Layer Implementations
// ============================================================================

// MockNarrationAgent is a placeholder implementation.
type MockNarrationAgent struct{}

func NewMockNarrationAgent() *MockNarrationAgent {
	return &MockNarrationAgent{}
}

func (m *MockNarrationAgent) GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error) {
	return &SkeletonResult{
		WorldView: "A mysterious horror world awaits...",
		MainTheme: req.Theme,
		Setting:   "An abandoned mansion",
	}, nil
}

func (m *MockNarrationAgent) GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error) {
	story := fmt.Sprintf("You find yourself in %s. %s", req.Skeleton.Setting, req.Skeleton.WorldView)
	return &OpeningResult{Story: story}, nil
}

func (m *MockNarrationAgent) GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error) {
	story := fmt.Sprintf("You chose: %s. The story continues...", req.PlayerChoice)
	return &ContentResult{Story: story}, nil
}

func (m *MockNarrationAgent) GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error) {
	story := fmt.Sprintf("The story ends with tension level: %d", req.TensionLevel)
	return &EndingResult{Story: story}, nil
}

// MockChoiceAgent is a placeholder implementation.
type MockChoiceAgent struct{}

func NewMockChoiceAgent() *MockChoiceAgent {
	return &MockChoiceAgent{}
}

func (m *MockChoiceAgent) GenerateChoices(ctx context.Context, req ChoiceRequest) (*ChoiceResult, error) {
	return &ChoiceResult{
		Choices: []string{
			"Explore the room",
			"Check the window",
			"Open the door",
		},
	}, nil
}

// MockJudgeAgent is a placeholder implementation.
type MockJudgeAgent struct{}

func NewMockJudgeAgent() *MockJudgeAgent {
	return &MockJudgeAgent{}
}

func (m *MockJudgeAgent) Judge(ctx context.Context, req JudgeRequest) (*JudgeResult, error) {
	return &JudgeResult{
		StateChanges: StateChanges{
			HPDelta:  0,
			SANDelta: -1,
		},
		RuleViolated: false,
	}, nil
}

// MockSeedAgent is a placeholder implementation.
type MockSeedAgent struct{}

func NewMockSeedAgent() *MockSeedAgent {
	return &MockSeedAgent{}
}

func (m *MockSeedAgent) GenerateGlobal(ctx context.Context, params agents.GenerateGlobalParams) ([]*seed.GlobalSeed, error) {
	// Determine count based on difficulty
	count := 3
	switch params.Difficulty {
	case "easy":
		count = 3
	case "medium":
		count = 4
	case "hard":
		count = 5
	}

	seeds := make([]*seed.GlobalSeed, 0, count)
	for i := 0; i < count; i++ {
		s, _ := seed.NewGlobalSeed(
			fmt.Sprintf("GS%03d", i+1),
			fmt.Sprintf("Global seed %d for %s", i+1, params.MainTheme),
			fmt.Sprintf("Truth %d", i+1),
			"mysterious",
			[]seed.ClueTier{
				{Tier: 1, Content: "Tier 1 clue", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
				{Tier: 2, Content: "Tier 2 clue", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
				{Tier: 3, Content: "Tier 3 clue", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
			},
		)
		seeds = append(seeds, s)
	}
	return seeds, nil
}

// MockNPCAgent is a placeholder implementation.
type MockNPCAgent struct{}

func NewMockNPCAgent() *MockNPCAgent {
	return &MockNPCAgent{}
}

func (m *MockNPCAgent) GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error) {
	profiles := make([]*NPCProfile, req.Count)
	for i := 0; i < req.Count; i++ {
		profiles[i] = &NPCProfile{
			ID:          fmt.Sprintf("NPC%03d", i+1),
			Name:        fmt.Sprintf("Character %d", i+1),
			Description: fmt.Sprintf("NPC for %s theme", req.Skeleton.MainTheme),
		}
	}
	return profiles, nil
}

// ============================================================================
// Mock Template Library Implementation
// ============================================================================

// MockTemplateLibrary is a placeholder implementation.
type MockTemplateLibrary struct{}

func NewMockTemplateLibrary() *MockTemplateLibrary {
	return &MockTemplateLibrary{}
}

func (m *MockTemplateLibrary) SelectTemplates(theme string, difficulty string) *engine.UsedTemplates {
	return &engine.UsedTemplates{
		Rules:  []string{"rule-01", "rule-02"},
		Scenes: []string{"scene-01"},
	}
}
