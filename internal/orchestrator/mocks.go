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

// PruneLocalSeedsByScene is a mock implementation for testing.
// Real implementation exists in internal/engine/seed/seed_manager.go (Story 2.3).
func (m *MockSeedManager) PruneLocalSeedsByScene(sceneID string) []PruneResult {
	// Mock: return empty results (no seeds to prune)
	// Real implementation would prune all Active LocalSeeds matching sceneID
	return []PruneResult{}
}

// PruneExpiredLocalSeeds is a mock implementation for testing.
// Real implementation exists in internal/engine/seed/seed_manager.go (Story 2.3).
func (m *MockSeedManager) PruneExpiredLocalSeeds(currentBeat int) []PruneResult {
	// Mock: return empty results (no expired seeds)
	// Real implementation would prune all Active LocalSeeds where IsExpired(currentBeat) == true
	return []PruneResult{}
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
// Also tracks scene changes for Story 2.5 (Seed Pruning Integration).
type MockStateManager struct {
	gameState     *engine.GameStateV2
	previousScene string // Track previous scene for change detection
}

func NewMockStateManager(gameState *engine.GameStateV2) *MockStateManager {
	return &MockStateManager{
		gameState:     gameState,
		previousScene: "", // Initially empty
	}
}

// ApplyChanges applies state changes and detects scene transitions.
// Returns ChangeResult with SceneChangeEvent if a scene change occurred.
func (m *MockStateManager) ApplyChanges(changes StateChanges) (*ChangeResult, error) {
	if m.gameState == nil {
		return nil, fmt.Errorf("gameState is nil")
	}

	result := &ChangeResult{}

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

	// Detect scene changes (Story 2.5)
	if changes.SceneChange != nil {
		newScene := *changes.SceneChange
		oldScene := m.gameState.CurrentScene

		// Only trigger event if scene actually changed
		if newScene != oldScene && oldScene != "" {
			result.SceneChanged = &SceneChangeEvent{
				OldScene: oldScene,
				NewScene: newScene,
				Beat:     m.gameState.GetCurrentBeat(),
			}
		}

		// Update game state and track previous scene
		m.previousScene = oldScene
		m.gameState.CurrentScene = newScene
	}

	return result, nil
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

// MockJudgeAgent is a placeholder implementation with configurable results for testing.
type MockJudgeAgent struct {
	nextResult *JudgeResult // Optional: override default result for next Judge call
}

func NewMockJudgeAgent() *MockJudgeAgent {
	return &MockJudgeAgent{}
}

func (m *MockJudgeAgent) Judge(ctx context.Context, req JudgeRequest) (*JudgeResult, error) {
	// If a custom result is set, return it and clear
	if m.nextResult != nil {
		result := m.nextResult
		m.nextResult = nil // Clear after use
		return result, nil
	}

	// Default behavior
	return &JudgeResult{
		StateChanges: StateChanges{
			HPDelta:  0,
			SANDelta: -1,
		},
		RuleViolated: false,
	}, nil
}

// ClassifyIntent classifies the player's free text input into a normalized intent.
// Code Review Fix 7-3-1: Mock implementation for testing.
func (m *MockJudgeAgent) ClassifyIntent(ctx context.Context, freeText string, gameState *agents.GameStateSnapshot) (*agents.IntentClassification, *agents.ClarificationNeeded, error) {
	// Default mock: return a clear intent with high confidence
	intent := &agents.IntentClassification{
		Action:           "examine",
		Target:           "unknown",
		IsAmbiguous:      false,
		Confidence:       0.9,
		Keywords:         []string{},
		NormalizedIntent: freeText, // Pass through for mock
	}
	return intent, nil, nil
}

// MockSeedAgent is a placeholder implementation.
type MockSeedAgent struct{}

func NewMockSeedAgent() *MockSeedAgent {
	return &MockSeedAgent{}
}

func (m *MockSeedAgent) InvokeGlobalGenerate(ctx context.Context, request *agents.GlobalGenerateRequest) (*agents.GlobalGenerateResponse, error) {
	// Determine count based on difficulty
	count := 3
	if request.StoryBible != nil {
		switch request.StoryBible.Difficulty {
		case "easy":
			count = 3
		case "medium":
			count = 4
		case "hard", "hell":
			count = 5
		}
	}

	seeds := make([]*seed.GlobalSeed, 0, count)
	theme := "unknown"
	if request.StoryBible != nil {
		theme = request.StoryBible.Theme
	}
	for i := 0; i < count; i++ {
		s, _ := seed.NewGlobalSeed(
			fmt.Sprintf("GS%03d", i+1),
			fmt.Sprintf("Global seed %d for %s", i+1, theme),
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
	return &agents.GlobalGenerateResponse{GlobalSeeds: seeds}, nil
}

func (m *MockSeedAgent) InvokeLocalManage(ctx context.Context, request *agents.LocalManageRequest) (*agents.LocalManageResponse, error) {
	// Simple mock implementation - just skip operation
	return &agents.LocalManageResponse{
		Operation:   agents.SeedOpSkip,
		TargetSeed:  nil,
		PrunedSeeds: nil,
	}, nil
}

// MockNPCAgent is a placeholder implementation (Story 7.6 enhanced).
type MockNPCAgent struct{}

func NewMockNPCAgent() *MockNPCAgent {
	return &MockNPCAgent{}
}

// GenerateProfiles generates NPC profiles with Show-Don't-Tell introductions (Story 7.6)
func (m *MockNPCAgent) GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error) {
	// Story 7.6: Generate NPCs with comprehensive information
	archetypeNames := []string{"引導者", "知情者", "中立者", "犧牲者"}
	archetypeIntros := []string{
		"她快步走來，從背包掏出急救包：「受傷了嗎？先處理一下。」疲憊的臉上仍帶著關切，動作熟練而溫和。",
		"他靠在牆邊，手中翻著破舊的筆記，頭也不抬：「你來晚了。」指尖劃過書頁上的符號，眼神深邃得像看穿了一切。",
		"他站在原地，眼神茫然地看著四周：「這...這是哪裡？」雙手不安地摩擦著，像是在確認自己是否還活著。",
		"她蜷縮在角落，雙手死死抓著沾血的繃帶，眼神不斷飄向門口。「救...救命...」聲音顫抖得幾乎聽不清，身體抖得像秋風中的落葉。",
	}

	profiles := make([]*NPCProfile, req.Count)
	for i := 0; i < req.Count; i++ {
		idx := i % len(archetypeNames)
		profiles[i] = &NPCProfile{
			ID:          fmt.Sprintf("NPC%03d", i+1),
			Name:        fmt.Sprintf("%s-%d", archetypeNames[idx], i+1),
			Description: archetypeIntros[idx], // Show-Don't-Tell introduction
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
