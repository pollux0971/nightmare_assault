package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// GamePhase represents the current phase of the game.
type GamePhase int

const (
	PhaseGenesis GamePhase = iota
	PhaseGameLoop
	PhaseConvergence
)

// Convergence thresholds
const (
	ConvergenceSeedProgressThreshold = 0.8  // 80% global seeds revealed triggers convergence
	ConvergenceTensionThreshold      = 95   // Tension >= 95/100 triggers convergence
	ConvergenceTargetBeats           = 20   // Approaching beat limit triggers convergence
)

func (p GamePhase) String() string {
	switch p {
	case PhaseGenesis:
		return "Genesis"
	case PhaseGameLoop:
		return "GameLoop"
	case PhaseConvergence:
		return "Convergence"
	default:
		return "Unknown"
	}
}

// StoryBible contains the immutable story skeleton generated in Genesis.
type StoryBible struct {
	WorldView     string
	MainTheme     string
	Setting       string
	GlobalSeeds   []*seed.GlobalSeed
	NPCProfiles   []*NPCProfile
	UsedTemplates *engine.UsedTemplates
}

// NPCProfile represents an NPC character profile.
// This is a placeholder - full implementation in Epic 6.
type NPCProfile struct {
	ID          string
	Name        string
	Description string
}

// GenesisResult contains the result of Phase 1: Genesis.
type GenesisResult struct {
	Story string // Opening narrative
}

// TurnResult contains the result of a single game loop turn.
type TurnResult struct {
	Story   string
	Choices []string
}

// EndingResult contains the result of Phase 3: Convergence.
type EndingResult struct {
	Story string // Ending narrative
}

// ============================================================================
// Logic Layer Interfaces (Placeholders - will be implemented in Epic 2-5)
// ============================================================================

// TensionEngine manages the tension/suspense system.
type TensionEngine interface {
	CalculateDelta(beat int, impact string, current int) int
	GetDirective(tension int) TensionDirective
}

type TensionDirective struct {
	Level             string
	Instruction       string
	AllowedElements   []string
	ForbiddenElements []string
}

// SeedManager manages global and local seeds.
// Note: This is a minimal interface for Story 2.2 - will be expanded in later stories.
type SeedManager interface {
	CheckHarvest(currentBeat int) []*seed.HarvestInstruction
	MarkSeedRevealed(seedID string, currentBeat int) error
	GetGlobalSeedsProgress() float64
}

// Placeholder interfaces for future implementation
type PlantInstruction struct {
	Type    string
	Content string
}

// ContextManager manages the context window and summarization.
type ContextManager interface {
	GetWindow(history []HistoryEntry, tension int) ContextWindow
}

type ContextWindow struct {
	Summary       string
	RecentEntries []HistoryEntry
}

type HistoryEntry struct {
	Beat     int
	Story    string
	Choice   string
	HPDelta  int
	SANDelta int
}

// RuleEngine manages hidden rules.
type RuleEngine interface {
	CheckViolation(action string) (bool, RuleViolation)
}

type RuleViolation struct {
	RuleID    string
	RuleName  string
	SANDamage int
}

// StateManager manages HP/SAN calculations.
type StateManager interface {
	ApplyChanges(changes StateChanges)
}

type StateChanges struct {
	HPDelta      int
	SANDelta     int
	TensionDelta int
}

// ============================================================================
// Agent Layer Interfaces (Placeholders - will be implemented in Epic 6)
// ============================================================================

// NarrationAgent generates story narrative.
type NarrationAgent interface {
	GenerateSkeleton(ctx context.Context, req SkeletonRequest) (*SkeletonResult, error)
	GenerateOpening(ctx context.Context, req OpeningRequest) (*OpeningResult, error)
	GenerateContent(ctx context.Context, req ContentRequest) (*ContentResult, error)
	GenerateEnding(ctx context.Context, req EndingRequest) (*EndingResult, error)
}

type SkeletonRequest struct {
	Theme      string
	Difficulty string
}

type SkeletonResult struct {
	WorldView string
	MainTheme string
	Setting   string
}

type OpeningRequest struct {
	Skeleton StoryBible
}

type OpeningResult struct {
	Story string
}

type ContentRequest struct {
	TensionDirective    TensionDirective
	HarvestInstructions []*seed.HarvestInstruction
	PlantInstructions   []PlantInstruction
	ContextWindow       ContextWindow
	PlayerChoice        string
}

type ContentResult struct {
	Story string
}

type EndingRequest struct {
	Bible         StoryBible
	RevealedSeeds []*seed.GlobalSeed
	TensionLevel  int
}

// ChoiceAgent generates player choices.
type ChoiceAgent interface {
	GenerateChoices(ctx context.Context, req ChoiceRequest) (*ChoiceResult, error)
}

type ChoiceRequest struct {
	Story        string
	CurrentScene string
	Tension      int
}

type ChoiceResult struct {
	Choices []string
}

// JudgeAgent judges player actions.
type JudgeAgent interface {
	Judge(ctx context.Context, req JudgeRequest) (*JudgeResult, error)
}

type JudgeRequest struct {
	Action       string
	CurrentState *engine.GameStateV2
}

type JudgeResult struct {
	StateChanges StateChanges
	RuleViolated bool
	Violation    RuleViolation
}

// SeedAgent generates seeds.
// Use the agents.SeedAgent implementation from internal/orchestrator/agents package.
type SeedAgent interface {
	GenerateGlobal(ctx context.Context, params agents.GenerateGlobalParams) ([]*seed.GlobalSeed, error)
}

// NPCAgent generates NPCs.
type NPCAgent interface {
	GenerateProfiles(ctx context.Context, req NPCRequest) ([]*NPCProfile, error)
}

type NPCRequest struct {
	Skeleton SkeletonResult
	Count    int
}

// ============================================================================
// Template Library Interface (Placeholder - will be implemented in Epic 4)
// ============================================================================

// TemplateLibrary manages YAML templates.
type TemplateLibrary interface {
	SelectTemplates(theme string, difficulty string) *engine.UsedTemplates
}

// ============================================================================
// Orchestrator Main Structure
// ============================================================================

// Orchestrator is the central coordinator for the v2.0 architecture.
type Orchestrator struct {
	currentPhase GamePhase

	// Data storage
	storyBible *StoryBible
	gameState  *engine.GameStateV2

	// Logic layer engines
	tensionEngine TensionEngine
	seedManager   SeedManager
	contextMgr    ContextManager
	ruleEngine    RuleEngine
	stateMgr      StateManager

	// Agent layer
	narrationAgent NarrationAgent
	choiceAgent    ChoiceAgent
	judgeAgent     JudgeAgent
	seedAgent      SeedAgent
	npcAgent       NPCAgent

	// Template library
	templateLib TemplateLibrary

	mu sync.RWMutex
}

// NewOrchestrator creates a new orchestrator with all dependencies initialized.
func NewOrchestrator() *Orchestrator {
	log.Println("[Orchestrator] Initializing v2.0 architecture...")

	gameState := engine.NewGameStateV2()

	orch := &Orchestrator{
		currentPhase: PhaseGenesis,
		storyBible:   &StoryBible{},
		gameState:    gameState,

		// Initialize Logic layer with placeholder implementations
		tensionEngine: NewMockTensionEngine(),
		seedManager:   NewMockSeedManager(),
		contextMgr:    NewMockContextManager(),
		ruleEngine:    NewMockRuleEngine(),
		stateMgr:      NewMockStateManager(gameState),

		// Initialize Agent layer with placeholder implementations
		narrationAgent: NewMockNarrationAgent(),
		choiceAgent:    NewMockChoiceAgent(),
		judgeAgent:     NewMockJudgeAgent(),
		seedAgent:      NewMockSeedAgent(),
		npcAgent:       NewMockNPCAgent(),

		// Initialize Template library with placeholder
		templateLib: NewMockTemplateLibrary(),
	}

	log.Println("[Orchestrator] Initialization complete")
	return orch
}

// GetCurrentPhase returns the current game phase.
func (o *Orchestrator) GetCurrentPhase() GamePhase {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentPhase
}

// RunPhaseGenesis executes Phase 1: Genesis (world generation).
func (o *Orchestrator) RunPhaseGenesis(ctx context.Context) (*GenesisResult, error) {
	log.Println("[Orchestrator] Starting Phase 1: Genesis")

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// Step 1: Select templates
	templates := o.templateLib.SelectTemplates("horror", "medium")
	o.storyBible.UsedTemplates = templates

	// Step 2: Generate skeleton (parallel would use goroutines)
	skeleton, err := o.narrationAgent.GenerateSkeleton(ctx, SkeletonRequest{
		Theme:      "horror",
		Difficulty: "medium",
	})
	if err != nil {
		return nil, fmt.Errorf("skeleton generation failed: %w", err)
	}

	// Check context after skeleton generation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after skeleton generation: %w", err)
	}

	// Step 3: Generate global seeds
	globalSeeds, err := o.seedAgent.GenerateGlobal(ctx, agents.GenerateGlobalParams{
		WorldView:  skeleton.WorldView,
		MainTheme:  skeleton.MainTheme,
		Difficulty: "medium", // TODO: Derive from game difficulty setting
	})
	if err != nil {
		return nil, fmt.Errorf("global seed generation failed: %w", err)
	}

	// Check context after seed generation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after seed generation: %w", err)
	}

	// Step 4: Generate NPC profiles
	npcProfiles, err := o.npcAgent.GenerateProfiles(ctx, NPCRequest{
		Skeleton: *skeleton,
		Count:    2,
	})
	if err != nil {
		return nil, fmt.Errorf("NPC generation failed: %w", err)
	}

	// Check context after NPC generation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after NPC generation: %w", err)
	}

	// Step 5: Save to Story Bible
	o.storyBible.WorldView = skeleton.WorldView
	o.storyBible.MainTheme = skeleton.MainTheme
	o.storyBible.Setting = skeleton.Setting
	o.storyBible.GlobalSeeds = globalSeeds
	o.storyBible.NPCProfiles = npcProfiles

	// Step 6: Generate opening
	opening, err := o.narrationAgent.GenerateOpening(ctx, OpeningRequest{
		Skeleton: *o.storyBible,
	})
	if err != nil {
		return nil, fmt.Errorf("opening generation failed: %w", err)
	}

	// Step 7: Transition to Game Loop
	o.mu.Lock()
	o.currentPhase = PhaseGameLoop
	o.mu.Unlock()

	log.Println("[Orchestrator] Phase 1: Genesis complete")
	return &GenesisResult{Story: opening.Story}, nil
}

// RunGameLoopTurn executes one turn of Phase 2: Game Loop.
func (o *Orchestrator) RunGameLoopTurn(ctx context.Context, playerChoice string) (*TurnResult, error) {
	log.Printf("[Orchestrator] Running Game Loop Turn (choice: %s)", playerChoice)

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// Step 1: Judge action
	judgeResult, err := o.judgeAgent.Judge(ctx, JudgeRequest{
		Action:       playerChoice,
		CurrentState: o.gameState,
	})
	if err != nil {
		return nil, fmt.Errorf("judge failed: %w", err)
	}

	// Check context after judge
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after judge: %w", err)
	}

	// Step 2: Apply state changes
	o.stateMgr.ApplyChanges(judgeResult.StateChanges)

	// Step 3: Update tension
	tensionDelta := o.tensionEngine.CalculateDelta(
		o.gameState.GetCurrentBeat(),
		"normal",
		o.gameState.Tension.Value,
	)
	o.stateMgr.ApplyChanges(StateChanges{
		TensionDelta: tensionDelta,
	})

	// Step 4: Get tension directive
	directive := o.tensionEngine.GetDirective(o.gameState.Tension.Value)

	// Step 5: Seed management
	harvestInstructions := o.seedManager.CheckHarvest(o.gameState.GetCurrentBeat())

	// Note: PlantInstructions (Local Seeds) will be implemented in Story 2.3
	var plantInstructions []PlantInstruction

	// Step 6: Context window
	contextWindow := o.contextMgr.GetWindow(nil, o.gameState.Tension.Value)

	// Step 7: Generate narration
	narration, err := o.narrationAgent.GenerateContent(ctx, ContentRequest{
		TensionDirective:    directive,
		HarvestInstructions: harvestInstructions,
		PlantInstructions:   plantInstructions,
		ContextWindow:       contextWindow,
		PlayerChoice:        playerChoice,
	})
	if err != nil {
		return nil, fmt.Errorf("narration generation failed: %w", err)
	}

	// Check context after narration
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after narration: %w", err)
	}

	// Step 7.5: Mark harvested seeds as revealed to advance tiers
	for _, instruction := range harvestInstructions {
		if err := o.seedManager.MarkSeedRevealed(instruction.SeedID, o.gameState.GetCurrentBeat()); err != nil {
			log.Printf("[WARN] Failed to mark seed revealed: seedID=%s, beat=%d, error=%v",
				instruction.SeedID, o.gameState.GetCurrentBeat(), err)
			// Non-fatal error - continue execution
		}
	}

	// Note: MarkSeedRevealed is only available on real SeedManager implementation,
	// not on the SeedManager interface. This will be fixed when real implementation
	// is injected instead of MockSeedManager.

	// Step 8: Generate choices
	choices, err := o.choiceAgent.GenerateChoices(ctx, ChoiceRequest{
		Story:        narration.Story,
		CurrentScene: o.gameState.CurrentScene,
		Tension:      o.gameState.Tension.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("choice generation failed: %w", err)
	}

	// Step 9: Increment beat
	o.gameState.IncrementBeat()

	// Step 10: Check convergence conditions
	if o.shouldConverge() {
		o.mu.Lock()
		o.currentPhase = PhaseConvergence
		o.mu.Unlock()
		log.Println("[Orchestrator] Convergence conditions met, switching to Phase 3")
	}

	log.Printf("[Orchestrator] Game Loop Turn complete (beat: %d, tension: %d)",
		o.gameState.GetCurrentBeat(),
		o.gameState.Tension.Value)

	return &TurnResult{
		Story:   narration.Story,
		Choices: choices.Choices,
	}, nil
}

// RunPhaseConvergence executes Phase 3: Convergence (ending generation).
func (o *Orchestrator) RunPhaseConvergence(ctx context.Context) (*EndingResult, error) {
	log.Println("[Orchestrator] Starting Phase 3: Convergence")

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// Generate ending
	ending, err := o.narrationAgent.GenerateEnding(ctx, EndingRequest{
		Bible:         *o.storyBible,
		RevealedSeeds: o.storyBible.GlobalSeeds,
		TensionLevel:  o.gameState.Tension.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("ending generation failed: %w", err)
	}

	log.Println("[Orchestrator] Phase 3: Convergence complete")
	return ending, nil
}

// shouldConverge checks if the game should transition to convergence phase.
func (o *Orchestrator) shouldConverge() bool {
	// Condition 1: Global seeds progress >= 80%
	progress := o.seedManager.GetGlobalSeedsProgress()
	if progress >= ConvergenceSeedProgressThreshold {
		log.Printf("[Convergence] Triggered by seed progress: %.2f%%", progress*100)
		return true
	}

	// Condition 2: Tension >= 95
	if o.gameState.Tension.Value >= ConvergenceTensionThreshold {
		log.Printf("[Convergence] Triggered by tension: %d", o.gameState.Tension.Value)
		return true
	}

	// Condition 3: Approaching target beats
	if o.gameState.GetCurrentBeat() >= ConvergenceTargetBeats {
		log.Printf("[Convergence] Triggered by beat limit: %d", o.gameState.GetCurrentBeat())
		return true
	}

	return false
}
