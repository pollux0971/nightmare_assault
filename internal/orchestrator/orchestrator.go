package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/momentum"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
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
//
// Story 7.1 AC #4: Complete Story Bible structure containing all Phase 2/3 required data:
//   - WorldSetting: Complete world configuration (location, history, weird elements, atmosphere)
//   - CoreMystery: The hidden truth and revelation mechanics
//   - StoryArc: Three-act structure with beat ranges and turning points
//   - HiddenRules: List of hidden rules (generated based on difficulty)
//   - GlobalSeeds: Main plot foreshadowing (3-5 seeds strongly bound to endings)
//   - NPCProfiles: Complete NPC data (2-4 teammates)
//   - PossibleEndings: At least 3 different ending branches
//   - UsedTemplates: Selected templates from Template Library
//   - Dreams: Dream blueprints for opening and chapter dreams (Story 9-1, 9-2)
type StoryBible struct {
	// Metadata
	GameID      string `json:"game_id"`
	CreatedAt   string `json:"created_at"`
	Difficulty  string `json:"difficulty"`
	TotalBeats  int    `json:"total_beats"`

	// Core Story Elements
	WorldView     string             `json:"world_view"`      // Deprecated: use WorldSetting
	MainTheme     string             `json:"main_theme"`
	Setting       string             `json:"setting"`         // Deprecated: use WorldSetting.Location

	// Story 7.1: Enhanced Story Bible
	WorldSetting    *WorldSetting      `json:"world_setting"`
	CoreMystery     *CoreMystery       `json:"core_mystery"`
	StoryArc        *StoryArc          `json:"story_arc"`
	HiddenRules     []*HiddenRule      `json:"hidden_rules"`
	GlobalSeeds     []*seed.GlobalSeed `json:"global_seeds"`
	NPCProfiles     []*NPCProfile      `json:"npc_profiles"`
	PossibleEndings []*Ending          `json:"possible_endings"`
	UsedTemplates   *engine.UsedTemplates `json:"used_templates"`

	// Story 9: Dream System
	Dreams []*DreamBlueprint `json:"dreams"` // Dream blueprints
}

// WorldSetting defines the game world configuration.
// Story 7.1 AC #4: Complete world setting with all atmospheric details.
type WorldSetting struct {
	Location      string   `json:"location"`       // e.g., "廢棄精神病院"
	History       string   `json:"history"`        // 100-200 characters background
	WeirdElements []string `json:"weird_elements"` // List of weird/uncanny elements
	Atmosphere    string   `json:"atmosphere"`     // Atmosphere description
	TimeFrame     string   `json:"time_frame"`     // e.g., "1990年代"
	Background    string   `json:"background"`     // Detailed background (500-800 chars)
}

// CoreMystery defines the hidden truth at the story's core.
// Story 7.1 AC #4: The mystery that drives the narrative.
type CoreMystery struct {
	Question   string `json:"question"`   // The core mystery question
	CoreTruth  string `json:"core_truth"` // The hidden truth
	Revelation string `json:"revelation"` // How/when the truth is revealed
	HiddenFrom string `json:"hidden_from"` // How it's hidden from players
}

// StoryArc defines the three-act structure and key plot points.
// Story 7.1 AC #4: Structural skeleton for narrative pacing.
type StoryArc struct {
	Act1End       int             `json:"act1_end"`        // Beat where Act 1 ends
	Midpoint      int             `json:"midpoint"`        // Beat at story midpoint
	Act2End       int             `json:"act2_end"`        // Beat where Act 2 ends
	TurningPoints []*TurningPoint `json:"turning_points"`  // Key narrative moments
}

// TurningPoint represents a key moment in the story.
type TurningPoint struct {
	Name        string `json:"name"`         // e.g., "Inciting Incident"
	Beat        int    `json:"beat"`         // When it occurs
	Description string `json:"description"`  // What happens
}

// HiddenRule represents a game rule that players must discover.
// Story 7.1 AC #4: Rules vary by difficulty (≤6 for Easy, unlimited for Hard/Hell).
type HiddenRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Hints       []string `json:"hints"`       // Clues about the rule
	Penalty     string   `json:"penalty"`     // What happens when violated
}

// Ending represents a possible game ending.
// Story 7.1 AC #4: At least 3 different endings based on player choices.
type Ending struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Type                   string           `json:"type"` // "true"/"good"/"bad"
	Condition              *EndingCondition `json:"condition"`
	Description            string           `json:"description"`
	RequiredSeedPercentage float64          `json:"required_seed_percentage"` // 0.0-1.0
}

// EndingCondition defines what triggers an ending.
type EndingCondition struct {
	MinSeedPercentage float64 `json:"min_seed_percentage"` // Minimum seed reveal %
	MaxRuleViolations int     `json:"max_rule_violations,omitempty"`
	MinHP             int     `json:"min_hp,omitempty"`
	MinSAN            int     `json:"min_san,omitempty"`
}

// DreamBlueprint represents a dream blueprint in the Story Bible.
// Story 9-1, 9-2: Dreams are generated during Genesis and triggered during Game Loop.
type DreamBlueprint struct {
	ID              string   `json:"id"`                // Unique dream ID
	Type            string   `json:"type"`              // "opening" or "chapter"
	Content         string   `json:"content"`           // Dream narrative (200-400 chars for opening, 100-300 for chapter)
	RelatedRuleIDs  []string `json:"related_rule_ids"`  // Rule IDs hinted in this dream
	Clarity         float64  `json:"clarity"`           // Clue clarity (0.2-0.4 for opening, increases for chapter)
	TriggerBeat     int      `json:"trigger_beat"`      // When to trigger (0 for opening)
	TriggerSAN      int      `json:"trigger_san"`       // SAN threshold (<50 for low SAN triggers)
	TriggerEvent    string   `json:"trigger_event"`     // Special event trigger (e.g., "npc_death", "rule_violation")
	Symbols         []string `json:"symbols"`           // Symbolic imagery in dream
	Atmosphere      string   `json:"atmosphere"`        // "calm", "uneasy", "nightmare"
	IsTriggered     bool     `json:"is_triggered"`      // Whether dream has been shown
}

// NPCProfile represents a complete NPC character profile with all attributes
//
// Story 7.6: Enhanced NPC Profile with full attributes
//
// Purpose:
//   - Store all NPC attributes for persistence in Story Bible
//   - Support serialization/deserialization for save/load
//   - Bridge between NPC Agent's NPCInstance and Orchestrator's StoryBible
//
// Fields (AC #2):
//   - ID: Unique identifier
//   - Name: Character name fitting theme/scene (e.g., "護士王小芳" for hospital)
//   - Archetype: NPC archetype (N-01 to N-06)
//   - Personality: 3-5 keywords (helpless, mysterious, cold, etc.)
//   - Appearance: 50-100 character description
//   - Backstory: 100-200 character background story
//   - Skills: 1-2 practical skills
//   - Inventory: 1-3 items
//   - Secret: Hidden information, possibly related to Global Seeds
//   - Introduction: Show-Don't-Tell introduction (100-200 chars)
//   - LinkedSeeds: Connected Global Seed IDs (0-2 based on archetype)
//   - DeathTiming: Scheduled death beat (0 = immortal, >0 for N-01 Sacrificial)
//   - Status: Current status (alive/dying/dead)
//   - DeathBeat: Actual death beat (0 if not dead)
//   - DeathReason: Reason for death (empty if not dead)
type NPCProfile struct {
	// Basic Information
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Archetype   agents.NPCArchetype `json:"archetype"`
	Personality []string            `json:"personality"` // 3-5 keywords

	// Detailed Attributes (Story 7.6)
	Appearance   string   `json:"appearance"`   // 50-100 chars
	Backstory    string   `json:"backstory"`    // 100-200 chars
	Skills       []string `json:"skills"`       // 1-2 practical skills
	Inventory    []string `json:"inventory"`    // 1-3 items
	Secret       string   `json:"secret"`       // Hidden information
	Introduction string   `json:"introduction"` // Show-Don't-Tell intro (100-200 chars)

	// Story Integration
	LinkedSeeds []string          `json:"linked_seeds"` // Global Seed IDs (0-2)
	DeathTiming int               `json:"death_timing"` // Scheduled death beat (0 = immortal)
	Status      agents.NPCStatus  `json:"status"`       // Current status
	DeathBeat   int               `json:"death_beat"`   // Actual death beat
	DeathReason string            `json:"death_reason"` // Reason for death

	// Legacy field for backward compatibility
	Description string `json:"description,omitempty"` // Deprecated: use Backstory instead
}

// GenesisResult contains the result of Phase 1: Genesis.
type GenesisResult struct {
	Story string // Opening narrative
}

// TurnResult contains the result of a single game loop turn.
type TurnResult struct {
	Story        string
	Choices      []string
	PruneResults []PruneResult // Pruned LocalSeeds during this turn (for future Narration integration)
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
// Extended in Story 2.5 to support pruning operations.
type SeedManager interface {
	CheckHarvest(currentBeat int) []*seed.HarvestInstruction
	MarkSeedRevealed(seedID string, currentBeat int) error
	GetGlobalSeedsProgress() float64

	// Pruning methods (Story 2.5 - Integration)
	// Real implementation exists in internal/engine/seed/seed_manager.go (Story 2.3)
	PruneLocalSeedsByScene(sceneID string) []PruneResult
	PruneExpiredLocalSeeds(currentBeat int) []PruneResult
}

// PruneResult represents the result of pruning a LocalSeed.
// This mirrors the structure from internal/engine/seed/local_seed.go (Story 2.3).
type PruneResult struct {
	SeedID         string
	SceneID        string
	Content        string
	PruneReason    string // "scene_change" or "expired"
	TransitionText string // Optional narrative transition
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

// StateManager manages HP/SAN calculations and scene transitions.
type StateManager interface {
	ApplyChanges(changes StateChanges) (*ChangeResult, error)
}

type StateChanges struct {
	HPDelta      int
	SANDelta     int
	TensionDelta int
	SceneChange  *string // Optional: new scene ID if scene changes
}

// SceneChangeEvent represents a scene transition event.
// Generated when the player moves from one scene to another.
type SceneChangeEvent struct {
	OldScene string // Previous scene ID
	NewScene string // New scene ID
	Beat     int    // Beat number when the change occurred
}

// ChangeResult contains the results of applying state changes.
// Used to track what actually changed during a turn (e.g., scene transitions).
type ChangeResult struct {
	SceneChanged *SceneChangeEvent // Non-nil if a scene change occurred
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
// Code Review Fix 7-3-1: Extended interface to support intent classification
type JudgeAgent interface {
	Judge(ctx context.Context, req JudgeRequest) (*JudgeResult, error)

	// ClassifyIntent analyzes free text input to determine player intent.
	// Story 7.3 AC3: Parse player's free text input to understand their intention.
	// Returns IntentClassification and optionally ClarificationNeeded if input is ambiguous.
	ClassifyIntent(ctx context.Context, freeText string, gameState *agents.GameStateSnapshot) (*agents.IntentClassification, *agents.ClarificationNeeded, error)
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
	InvokeGlobalGenerate(ctx context.Context, request *agents.GlobalGenerateRequest) (*agents.GlobalGenerateResponse, error)
	InvokeLocalManage(ctx context.Context, request *agents.LocalManageRequest) (*agents.LocalManageResponse, error)
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

	// Story 7-4: Momentum controller for auto-resolve
	momentumController *momentum.MomentumController

	// Story 9-10: Trinity LLM Router for intelligent tier routing
	trinityRouter *trinity.TrinityRouter

	mu sync.RWMutex
}

// NewOrchestrator creates a new orchestrator with all dependencies initialized.
// Story 9-10: Now uses TrinityRouter for intelligent tier routing
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

		// Story 9-10: TrinityRouter is nil by default (must be injected)
		trinityRouter: nil,
	}

	log.Println("[Orchestrator] Initialization complete")
	return orch
}

// NewOrchestratorWithTrinity creates a new orchestrator with TrinityRouter integrated
// Story 9-10: Main constructor for production use with Trinity routing
//
// Parameters:
//   - router: The TrinityRouter instance for LLM tier routing
//
// Returns:
//   - *Orchestrator: A new orchestrator with Trinity routing enabled
func NewOrchestratorWithTrinity(router *trinity.TrinityRouter) *Orchestrator {
	log.Println("[Orchestrator] Initializing v2.0 architecture with Trinity routing...")

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

		// Story 9-10: Trinity Router for intelligent tier routing
		trinityRouter: router,
	}

	log.Println("[Orchestrator] Initialization complete with Trinity routing")
	return orch
}

// NewOrchestratorWithProvider creates a new orchestrator with a single provider
// Story 9-10: Backward compatibility wrapper that creates a basic TrinityRouter
//
// This constructor provides backward compatibility for code that only has a single
// provider. It creates a TrinityRouter that uses the same provider for all tiers.
//
// Parameters:
//   - provider: A single LLM provider to use for all tiers
//
// Returns:
//   - *Orchestrator: A new orchestrator with basic Trinity routing
//   - error: Error if router creation fails
func NewOrchestratorWithProvider(provider client.Provider, apiKey string) (*Orchestrator, error) {
	log.Println("[Orchestrator] Initializing with single provider (backward compatibility mode)...")

	// Create a basic RouterConfig using the same provider for all tiers
	// This provides backward compatibility while still enabling Trinity features
	routerConfig := trinity.RouterConfig{
		// Use the same model for all tiers (backward compatible)
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-opus-4-20250514",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-haiku-20240307",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled:    true,
		AgentTierOverrides: make(map[string]trinity.TierLevel),
	}

	// Create TrinityRouter
	router, err := trinity.NewTrinityRouter(routerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Trinity router: %w", err)
	}

	// Create orchestrator with Trinity
	return NewOrchestratorWithTrinity(router), nil
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
	seedResponse, err := o.seedAgent.InvokeGlobalGenerate(ctx, &agents.GlobalGenerateRequest{
		StoryBible: &agents.SeedStoryBible{
			Theme:       skeleton.MainTheme,
			WorldView:   skeleton.WorldView,
			Difficulty:  "medium", // TODO: Derive from game difficulty setting
			CoreTruth:   "",       // TODO: Extract core truth from skeleton when available
			GlobalSeeds: nil,
		},
		Difficulty:      "medium", // TODO: Derive from game difficulty setting
		StoryLength:     "medium", // TODO: Derive from game settings
		PossibleEndings: nil,      // TODO: Pass endings from skeleton
	})
	if err != nil {
		return nil, fmt.Errorf("global seed generation failed: %w", err)
	}
	globalSeeds := seedResponse.GlobalSeeds

	// Check context after seed generation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled after seed generation: %w", err)
	}

	// Step 4: Generate NPC profiles (Story 7.6: 2-4 NPCs based on difficulty)
	// Difficulty mapping: easy=2, normal=3, hard=3, hell=4
	difficulty := "medium" // TODO: Get from actual game settings
	npcCount := GetNPCCountForDifficulty(difficulty)

	npcProfiles, err := o.npcAgent.GenerateProfiles(ctx, NPCRequest{
		Skeleton: *skeleton,
		Count:    npcCount,
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

	// Step 6: Generate opening (Story 7.6: Pass NPCs with Show-Don't-Tell introductions)
	opening, err := o.narrationAgent.GenerateOpening(ctx, OpeningRequest{
		Skeleton: *o.storyBible,
		// Note: NPCs will be passed when using real NarrationAgent
		// For now, mock doesn't use NPCs
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

	// Step 2: Apply state changes and detect scene transitions
	changeResult, err := o.stateMgr.ApplyChanges(judgeResult.StateChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to apply state changes: %w", err)
	}
	if changeResult == nil {
		return nil, fmt.Errorf("ApplyChanges returned nil result without error")
	}

	// Step 3: Update tension
	tensionDelta := o.tensionEngine.CalculateDelta(
		o.gameState.GetCurrentBeat(),
		"normal",
		o.gameState.Tension.Value,
	)
	_, err = o.stateMgr.ApplyChanges(StateChanges{
		TensionDelta: tensionDelta,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply tension changes: %w", err)
	}

	// Step 4: Get tension directive
	directive := o.tensionEngine.GetDirective(o.gameState.Tension.Value)

	// Step 4.5: Pruning - Scene change (Story 2.5)
	// If scene changed, prune all LocalSeeds from the old scene
	var scenePruneResults []PruneResult
	if changeResult.SceneChanged != nil {
		scenePruneResults = o.seedManager.PruneLocalSeedsByScene(changeResult.SceneChanged.OldScene)

		log.Printf("[Orchestrator] Scene changed: %s → %s (beat %d), pruned %d LocalSeeds",
			changeResult.SceneChanged.OldScene,
			changeResult.SceneChanged.NewScene,
			changeResult.SceneChanged.Beat,
			len(scenePruneResults))

		// Log details of pruned seeds
		for _, pr := range scenePruneResults {
			log.Printf("[Pruning] Scene change: seedID=%s, sceneID=%s, content=%s",
				pr.SeedID, pr.SceneID, pr.Content)
		}
	}

	// Step 4.6: Pruning - Expired seeds (Story 2.5)
	// Check for expired LocalSeeds at the end of each turn
	expiredPruneResults := o.seedManager.PruneExpiredLocalSeeds(o.gameState.GetCurrentBeat())

	if len(expiredPruneResults) > 0 {
		log.Printf("[Orchestrator] Pruned %d expired LocalSeeds at beat %d",
			len(expiredPruneResults), o.gameState.GetCurrentBeat())

		// Log details of expired seeds
		for _, pr := range expiredPruneResults {
			log.Printf("[Pruning] Expired: seedID=%s, sceneID=%s, content=%s, transitionText=%s",
				pr.SeedID, pr.SceneID, pr.Content, pr.TransitionText)
		}
	}

	// Combine all pruning results for return and future Narration integration
	pruneResults := append(scenePruneResults, expiredPruneResults...)

	// Total pruning summary (using direct references for clarity)
	if len(pruneResults) > 0 {
		log.Printf("[Orchestrator] Total pruned: %d LocalSeeds (scene change: %d, expired: %d)",
			len(pruneResults),
			len(scenePruneResults),
			len(expiredPruneResults))
	}

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
		Story:        narration.Story,
		Choices:      choices.Choices,
		PruneResults: pruneResults, // Include for future Narration Agent integration (Epic 6)
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

// ==========================================================================
// Story 7.3: Choice Execution Flow
// ==========================================================================

// ChoiceExecutionRequest represents a request to execute a player choice.
// Story 7.3 AC4: Execute choice with rule checking and state changes.
type ChoiceExecutionRequest struct {
	// Choice data
	ChoiceText string // The text of the choice (predefined or free text)
	IsFreeText bool   // Whether this is free text input

	// Context
	BeatNumber int    // Current beat number
	Scene      string // Current scene
	Narration  string // The narration that led to this choice
}

// ChoiceExecutionResult represents the result of executing a choice.
// Story 7.3 AC4: Contains state changes, events, and next narration.
type ChoiceExecutionResult struct {
	// State changes
	HPDelta      int // HP change
	SANDelta     int // SAN change
	TensionDelta int // Tension change

	// Rule violations (if any)
	RulesViolated []agents.RuleViolation
	ImpactLevel   agents.ImpactLevel

	// Next action
	NextNarration string   // Generated narration
	NextChoices   []string // Next set of choices

	// Death information (if applicable)
	IsDeath     bool
	DeathReason string

	// Additional events
	SceneChanged  bool
	NPCReactions  []string // NPC reactions to the choice
	ItemsGained   []string // Items gained
	ItemsLost     []string // Items lost

	// Clarification (for ambiguous free text)
	// Code Review Fix 7-3-2: Extended clarification fields
	NeedsClarification  bool
	ClarificationPrompt string
	SuggestedOptions    []string // Suggested interpretations when clarification is needed
}

// ExecuteChoice executes a player choice and updates game state.
// Story 7.3 AC4: Complete choice execution flow.
//
// Flow:
//  1. If free text, classify intent via Judge Agent
//  2. If ambiguous, return clarification request
//  3. Check rules via Rule Engine
//  4. Calculate state changes (HP/SAN/Tension)
//  5. Apply state changes via State Manager
//  6. Trigger events (NPC reactions, scene changes)
//  7. Update tension and calculate directive
//  8. Generate next narration via Narration Agent
//  9. Generate next choices via Choice Agent
//
// This is the main entry point for processing player choices in Story 7.3.
func (o *Orchestrator) ExecuteChoice(ctx context.Context, request *ChoiceExecutionRequest) (*ChoiceExecutionResult, error) {
	log.Printf("[Orchestrator] Executing choice: %s (beat: %d, free_text: %v)",
		request.ChoiceText, request.BeatNumber, request.IsFreeText)

	result := &ChoiceExecutionResult{}

	// Step 1: Intent classification for free text input
	// Story 7.3 AC3: Parse intent from free text
	// Code Review Fix 7-3-1, 7-3-2: Integrated intent classification with clarification flow
	var normalizedChoice string
	if request.IsFreeText {
		// Build game state snapshot for intent classification
		gameStateSnapshot := &agents.GameStateSnapshot{
			HP:           o.gameState.HP,
			SAN:          o.gameState.SAN,
			CurrentScene: o.gameState.CurrentScene,
			Difficulty:   "", // Difficulty is stored in GameConfig, not GameStateV2
			TurnNumber:   o.gameState.GetCurrentBeat(),
		}

		// Classify intent using Judge Agent
		intent, clarification, err := o.judgeAgent.ClassifyIntent(ctx, request.ChoiceText, gameStateSnapshot)
		if err != nil {
			log.Printf("[Orchestrator] Intent classification failed: %v, using raw text", err)
			// Fallback to raw text if classification fails
			normalizedChoice = request.ChoiceText
		} else {
			// Story 7.3 AC4: Handle clarification needed
			if clarification != nil {
				log.Printf("[Orchestrator] Clarification needed: %s", clarification.Reason)
				result.NeedsClarification = true
				result.ClarificationPrompt = clarification.Question
				result.SuggestedOptions = clarification.SuggestedInterpretations
				return result, nil
			}

			// Use normalized intent for rule matching
			normalizedChoice = intent.NormalizedIntent
			log.Printf("[Orchestrator] Intent classified: action=%s, target=%s, confidence=%.2f",
				intent.Action, intent.Target, intent.Confidence)
		}
	} else {
		// Predefined choice - use as-is
		normalizedChoice = request.ChoiceText
	}

	// Step 2: Judge the choice (rule checking)
	// Story 7.3 AC4: Rule checking via RuleEngine
	judgeResult, err := o.judgeAgent.Judge(ctx, JudgeRequest{
		Action:       normalizedChoice,
		CurrentState: o.gameState,
	})
	if err != nil {
		return nil, fmt.Errorf("judge failed: %w", err)
	}

	// Store rule violations
	result.ImpactLevel = agents.ImpactNone // Default
	if judgeResult.RuleViolated {
		result.RulesViolated = []agents.RuleViolation{
			{
				RuleID:    judgeResult.Violation.RuleID,
				RuleName:  judgeResult.Violation.RuleName,
				SANDamage: judgeResult.Violation.SANDamage,
			},
		}
	}

	// Step 3: Apply state changes
	// Story 7.3 AC4: State changes (HP/SAN/Inventory/Location)
	changeResult, err := o.stateMgr.ApplyChanges(judgeResult.StateChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to apply state changes: %w", err)
	}

	// Record state changes
	result.HPDelta = judgeResult.StateChanges.HPDelta
	result.SANDelta = judgeResult.StateChanges.SANDelta
	result.TensionDelta = judgeResult.StateChanges.TensionDelta

	// Check for scene changes
	if changeResult != nil && changeResult.SceneChanged != nil {
		result.SceneChanged = true
		log.Printf("[Orchestrator] Scene changed: %s → %s",
			changeResult.SceneChanged.OldScene,
			changeResult.SceneChanged.NewScene)
	}

	// Step 4: Check for death
	// Story 7.3 AC4: Death detection
	if o.gameState.GetHP() <= 0 || o.gameState.GetSAN() <= 0 {
		result.IsDeath = true
		result.DeathReason = "你的生命值或理智值已歸零。"
		// TODO: Generate detailed death reason via Judge Agent
		return result, nil
	}

	// Step 5: Update tension
	// Story 7.3 AC4: Tension calculation via TensionEngine
	tensionDelta := o.tensionEngine.CalculateDelta(
		o.gameState.GetCurrentBeat(),
		"normal",
		o.gameState.Tension.Value,
	)
	_, err = o.stateMgr.ApplyChanges(StateChanges{
		TensionDelta: tensionDelta,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply tension changes: %w", err)
	}
	result.TensionDelta += tensionDelta

	// Step 6: Get tension directive
	directive := o.tensionEngine.GetDirective(o.gameState.Tension.Value)

	// Step 7: Seed management (harvest/plant)
	harvestInstructions := o.seedManager.CheckHarvest(o.gameState.GetCurrentBeat())
	var plantInstructions []PlantInstruction

	// Step 8: Context window
	contextWindow := o.contextMgr.GetWindow(nil, o.gameState.Tension.Value)

	// Step 9: Generate next narration
	// Story 7.3 AC4: Generate next segment via Narration Agent
	narration, err := o.narrationAgent.GenerateContent(ctx, ContentRequest{
		TensionDirective:    directive,
		HarvestInstructions: harvestInstructions,
		PlantInstructions:   plantInstructions,
		ContextWindow:       contextWindow,
		PlayerChoice:        normalizedChoice,
	})
	if err != nil {
		return nil, fmt.Errorf("narration generation failed: %w", err)
	}
	result.NextNarration = narration.Story

	// Step 10: Generate next choices
	// Story 7.3 AC4: Generate choices via Choice Agent
	choices, err := o.choiceAgent.GenerateChoices(ctx, ChoiceRequest{
		Story:        narration.Story,
		CurrentScene: o.gameState.CurrentScene,
		Tension:      o.gameState.Tension.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("choice generation failed: %w", err)
	}
	result.NextChoices = choices.Choices

	log.Printf("[Orchestrator] Choice execution complete: HP=%d SAN=%d Tension=%d",
		o.gameState.GetHP(), o.gameState.GetSAN(), o.gameState.Tension.Value)

	return result, nil
}

// ==========================================================================
// Story 5.6: ChatSession Storage
// ==========================================================================

// SaveChatSession stores a chat session to the game state.
// Story 5.6 AC2: ChatSession 自動添加到 GameStateV2.ChatSessions
// Sessions are stored in chronological order (sorted by EndBeat).
// Optionally limits the maximum number of stored sessions to prevent unbounded growth.
func (o *Orchestrator) SaveChatSession(session *engine.ChatSession) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if session == nil {
		return fmt.Errorf("cannot save nil chat session")
	}

	// Validate required fields
	if session.ID == "" {
		return fmt.Errorf("chat session must have an ID")
	}

	if len(session.Participants) == 0 {
		return fmt.Errorf("chat session must have at least one participant")
	}

	// Set CreatedAt if not already set
	if session.CreatedAt == "" {
		session.CreatedAt = fmt.Sprintf("%d", o.gameState.GetCurrentBeat())
	}

	// Add to game state
	o.gameState.ChatSessions = append(o.gameState.ChatSessions, session)

	// Sort by EndBeat to maintain chronological order
	// Use stable sort to preserve order of sessions with same EndBeat
	for i := len(o.gameState.ChatSessions) - 1; i > 0; i-- {
		if o.gameState.ChatSessions[i].EndBeat < o.gameState.ChatSessions[i-1].EndBeat {
			// Swap
			o.gameState.ChatSessions[i], o.gameState.ChatSessions[i-1] =
				o.gameState.ChatSessions[i-1], o.gameState.ChatSessions[i]
		} else {
			break
		}
	}

	// Optional: Limit maximum stored sessions to prevent unbounded growth
	// Default limit: 100 sessions (configurable in the future)
	maxSessions := 100
	if len(o.gameState.ChatSessions) > maxSessions {
		// Remove oldest sessions (from the beginning)
		excess := len(o.gameState.ChatSessions) - maxSessions
		o.gameState.ChatSessions = o.gameState.ChatSessions[excess:]
		log.Printf("[Orchestrator] Trimmed %d old chat sessions (max: %d)", excess, maxSessions)
	}

	log.Printf("[Orchestrator] Chat session saved: ID=%s, Beats=%d-%d, Participants=%v, Messages=%d",
		session.ID, session.StartBeat, session.EndBeat, session.Participants, len(session.Messages))

	return nil
}

// ==========================================================================
// Story 5.5: Summary Injection to Main Narration
// ==========================================================================

// InjectChatSummary injects a chat summary into the narration context after chat exits.
// Story 5.5 AC1: ChatSummary 注入到 NarrationAgent context
//
// This method performs the following operations:
// 1. Formats the ChatSummary as narrative-friendly text
// 2. Adds the formatted summary to NarrationAgent's context
// 3. Applies emotion changes from summary to NPCManager
// 4. Propagates facts shared during conversation to knowledge system
//
// Parameters:
//   - session: The chat session containing the summary
//
// Returns an error if session or summary is nil.
func (o *Orchestrator) InjectChatSummary(session *engine.ChatSession) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if session == nil {
		return fmt.Errorf("cannot inject nil chat session")
	}

	if session.Summary == nil {
		log.Printf("[Orchestrator] Chat session %s has no summary, skipping injection", session.ID)
		return nil
	}

	log.Printf("[Orchestrator] Injecting chat summary for session %s (beat %d-%d)",
		session.ID, session.StartBeat, session.EndBeat)

	// 1. Format summary for narration
	formattedSummary := o.formatSummaryForNarration(session)

	// 2. Add to NarrationAgent context
	// TODO: Story 5.5 - This will be implemented when NarrationAgent interface is extended
	// For now, we log the formatted summary
	log.Printf("[Orchestrator] Formatted summary:\n%s", formattedSummary)

	// 3. Apply emotion changes
	if err := o.applySummaryEmotionChanges(session); err != nil {
		log.Printf("[Orchestrator] Warning: Failed to apply emotion changes: %v", err)
		// Continue despite error - emotion changes are not critical
	}

	// 4. Apply facts shared
	if err := o.applySummaryFacts(session); err != nil {
		log.Printf("[Orchestrator] Warning: Failed to apply facts: %v", err)
		// Continue despite error - fact propagation is not critical
	}

	log.Printf("[Orchestrator] Chat summary injection complete")
	return nil
}

// formatSummaryForNarration formats a ChatSummary into narrative-friendly text.
// Story 5.5 AC1, AC2: Format summary as natural language for NarrationAgent
//
// The formatted summary includes:
// - Dialogue record metadata (beat, participants)
// - Main topics discussed
// - Key decisions made
// - Relationship changes
// - Facts shared
// - Narrative impact
//
// Parameters:
//   - session: The chat session to format
//
// Returns formatted summary text or empty string if session/summary is nil.
func (o *Orchestrator) formatSummaryForNarration(session *engine.ChatSession) string {
	if session == nil || session.Summary == nil {
		return ""
	}

	// Type assert to views.ChatSummary
	// Since engine.ChatSession.Summary is interface{}, we need to assert the type
	summary, ok := session.Summary.(*ChatSummary)
	if !ok {
		log.Printf("[Orchestrator] Warning: Summary is not *ChatSummary type")
		return ""
	}

	var b strings.Builder

	// 1. Dialogue record metadata
	b.WriteString(fmt.Sprintf("【對話記錄 - 回合 %d-%d】\n", session.StartBeat, session.EndBeat))

	// 2. Participants
	participantNames := make([]string, 0, len(session.Participants))
	for _, participantID := range session.Participants {
		name := o.getParticipantName(participantID)
		participantNames = append(participantNames, name)
	}
	if len(participantNames) > 0 {
		b.WriteString(fmt.Sprintf("參與者：%s\n", strings.Join(participantNames, "、")))
	}

	// 3. Main topics
	if len(summary.MainTopics) > 0 {
		b.WriteString(fmt.Sprintf("討論話題：%s\n", strings.Join(summary.MainTopics, "、")))
	}

	// 4. Key decisions
	if len(summary.KeyDecisions) > 0 {
		b.WriteString("關鍵決策：\n")
		for _, decision := range summary.KeyDecisions {
			b.WriteString(fmt.Sprintf("  - %s\n", decision))
		}
	}

	// 5. Relationship changes
	if len(summary.RelationChanges) > 0 {
		b.WriteString("關係變化：\n")
		for npcID, change := range summary.RelationChanges {
			npcName := o.getParticipantName(npcID)
			b.WriteString(fmt.Sprintf("  - %s: %s\n", npcName, change))
		}
	}

	// 6. Facts shared
	if len(summary.FactsShared) > 0 {
		b.WriteString(fmt.Sprintf("分享資訊：%s\n", strings.Join(summary.FactsShared, "、")))
	}

	// 7. Narrative impact
	if summary.NarrativeImpact != "" {
		b.WriteString(fmt.Sprintf("影響：%s\n", summary.NarrativeImpact))
	}

	return b.String()
}

// applySummaryEmotionChanges applies emotion changes from summary to NPCManager.
// Story 5.5 AC3: 情感變化影響後續 NPC 行為
//
// This method:
// 1. Extracts emotion changes from EmotionChanges field
// 2. Parses text descriptions into EmotionDelta values
// 3. Applies deltas via NPCManager.AdjustEmotion
// 4. Falls back to RelationChanges if EmotionChanges is empty
//
// Parameters:
//   - session: The chat session containing emotion changes
//
// Returns an error if NPCManager is not available.
func (o *Orchestrator) applySummaryEmotionChanges(session *engine.ChatSession) error {
	if session == nil || session.Summary == nil {
		return nil
	}

	// Type assert to views.ChatSummary
	summary, ok := session.Summary.(*ChatSummary)
	if !ok {
		return fmt.Errorf("summary is not *ChatSummary type")
	}

	// Check if we have emotion changes to process
	hasEmotionChanges := len(summary.EmotionChanges) > 0 || len(summary.RelationChanges) > 0
	if !hasEmotionChanges {
		log.Printf("[Orchestrator] No emotion changes in summary")
		return nil
	}

	// TODO: Apply emotion changes when NPCManager is available in Orchestrator
	// For now, we log the emotion changes
	log.Printf("[Orchestrator] Emotion changes detected:")
	for npcID, changeDesc := range summary.EmotionChanges {
		log.Printf("  - %s: %s", npcID, changeDesc)
	}
	for npcID, relationChange := range summary.RelationChanges {
		log.Printf("  - %s (relation): %s", npcID, relationChange)
	}

	// Note: Full implementation requires NPCManager integration
	// This will be completed when NPCManager is added to Orchestrator

	return nil
}

// applySummaryFacts applies facts shared during conversation to knowledge system.
// Story 5.5 AC4: FactsShared 影響 NPC 知識庫
//
// This method:
// 1. Extracts facts from summary.FactsShared
// 2. Creates Fact instances for each shared fact
// 3. Registers facts in global knowledge base
// 4. Propagates facts to all chat participants
//
// Parameters:
//   - session: The chat session containing facts
//
// Returns an error if knowledge system is not available.
func (o *Orchestrator) applySummaryFacts(session *engine.ChatSession) error {
	if session == nil || session.Summary == nil {
		return nil
	}

	// Type assert to views.ChatSummary
	summary, ok := session.Summary.(*ChatSummary)
	if !ok {
		return fmt.Errorf("summary is not *ChatSummary type")
	}

	if len(summary.FactsShared) == 0 {
		log.Printf("[Orchestrator] No facts shared in summary")
		return nil
	}

	// TODO: Apply facts when UpdateManager is available in Orchestrator
	// For now, we log the facts
	log.Printf("[Orchestrator] Facts shared during conversation:")
	for i, factContent := range summary.FactsShared {
		log.Printf("  %d. %s", i+1, factContent)
		log.Printf("     Participants: %v", session.Participants)
	}

	// Note: Full implementation requires UpdateManager integration
	// This will be completed when UpdateManager is added to Orchestrator

	return nil
}

// getParticipantName retrieves the display name for a participant.
// Returns the NPC name if found in profiles, or the participant ID otherwise.
func (o *Orchestrator) getParticipantName(participantID string) string {
	if participantID == "player" {
		return "玩家"
	}

	// Look up NPC name in story bible
	if o.storyBible != nil {
		for _, profile := range o.storyBible.NPCProfiles {
			if profile != nil && profile.ID == participantID {
				return profile.Name
			}
		}
	}

	// Fallback to ID
	return participantID
}

// ChatSummary is a local type alias for views.ChatSummary to avoid import cycles.
// This will be replaced with proper import when package structure is refactored.
type ChatSummary struct {
	MainTopics      []string          `json:"main_topics"`
	KeyDecisions    []string          `json:"key_decisions"`
	RelationChanges map[string]string `json:"relation_changes"`
	FactsShared     []string          `json:"facts_shared"`
	Flags           []string          `json:"flags"`
	NarrativeImpact string            `json:"narrative_impact"`
	EmotionChanges  map[string]string `json:"emotion_changes,omitempty"`
}

// ============================================================================
// Story 7-4: Game Loop Integration (遊戲迴圈重構整合)
// ============================================================================

// GameLoopResult represents the result of a game loop execution
// This is returned to the TUI layer for display
type GameLoopResult struct {
	// Narratives generated during auto-resolve
	Narratives []string

	// Number of beats resolved automatically
	BeatsResolved int

	// Stop reason for why auto-resolve ended
	StopReason momentum.StopReason

	// Next action required
	RequiresChoice   bool   // Whether player choice is needed
	RequiresChatMode bool   // Whether chat mode should be entered
	InitiatingNPC    string // NPC initiating chat (if RequiresChatMode is true)

	// State changes during auto-resolve
	HPDelta  int
	SANDelta int

	// Game end conditions
	IsGameOver   bool
	GameOverType string // "death", "convergence", etc.
}

// RunGameLoop executes the game loop with auto-resolve and momentum control
// Story 7-4 Implementation
//
// AC1: RunGameLoop() 每回合調用 MomentumController.AutoResolve()
// AC2: 根據 StopReason 路由（NPCChat → ChatOverlay, Risk/Plot → ChoiceAgent）
// AC3: AutoResolveResult 顯示後繼續迴圈
// AC4: 正確處理 StopReasonEvent
// AC5: 保持現有死亡/收斂檢查
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - *GameLoopResult: Result containing narratives and next action
//   - error: Error if loop execution fails
func (o *Orchestrator) RunGameLoop(ctx context.Context) (*GameLoopResult, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Println("[Orchestrator] Running game loop with momentum control")

	// AC5: Check death conditions before starting
	if o.gameState.GetHP() <= 0 || o.gameState.GetSAN() <= 0 {
		return &GameLoopResult{
			IsGameOver:   true,
			GameOverType: "death",
		}, nil
	}

	// AC5: Check convergence conditions
	if o.shouldConverge() {
		o.currentPhase = PhaseConvergence
		return &GameLoopResult{
			IsGameOver:   true,
			GameOverType: "convergence",
		}, nil
	}

	// AC1: Build narrative context for momentum controller
	narrativeCtx := o.buildNarrativeContext()

	// AC1: Call MomentumController.AutoResolve()
	if o.momentumController == nil {
		// Fallback: No momentum controller, use traditional turn-based mode
		log.Println("[Orchestrator] No momentum controller, using traditional mode")
		return &GameLoopResult{
			RequiresChoice: true,
			BeatsResolved:  0,
		}, nil
	}

	autoResult := o.momentumController.AutoResolve(narrativeCtx)

	// AC2: Route based on StopReason
	result := &GameLoopResult{
		Narratives:    autoResult.Narratives,
		BeatsResolved: autoResult.BeatsResolved,
		StopReason:    autoResult.StopReason,
		HPDelta:       autoResult.HPDelta,
		SANDelta:      autoResult.SANDelta,
	}

	// Apply state changes from auto-resolve
	o.gameState.HP += autoResult.HPDelta
	o.gameState.SAN += autoResult.SANDelta

	// Increment beat for each beat resolved
	for i := 0; i < autoResult.BeatsResolved; i++ {
		o.gameState.IncrementBeat()
	}

	// Clamp HP/SAN values
	if o.gameState.HP > 100 {
		o.gameState.HP = 100
	}
	if o.gameState.HP < 0 {
		o.gameState.HP = 0
	}
	if o.gameState.SAN > 100 {
		o.gameState.SAN = 100
	}
	if o.gameState.SAN < 0 {
		o.gameState.SAN = 0
	}

	// AC5: Check death after auto-resolve
	if o.gameState.GetHP() <= 0 || o.gameState.GetSAN() <= 0 {
		result.IsGameOver = true
		result.GameOverType = "death"
		return result, nil
	}

	// AC2: Route based on StopReason
	switch autoResult.StopReason {
	case momentum.StopReasonNPCConversation:
		// AC2: NPCChat → ChatOverlay
		log.Printf("[Orchestrator] Routing to chat mode (NPC: %s)", narrativeCtx.InitiatingNPC)
		result.RequiresChatMode = true
		result.InitiatingNPC = narrativeCtx.InitiatingNPC

	case momentum.StopReasonMajorEvent:
		// AC4: Handle major events
		log.Println("[Orchestrator] Major event occurred, requiring player choice")
		result.RequiresChoice = true

	case momentum.StopReasonPlotPoint, momentum.StopReasonRiskLevel:
		// AC2: Risk/Plot → ChoiceAgent
		log.Printf("[Orchestrator] Routing to choice mode (reason: %s)", autoResult.StopReason.String())
		result.RequiresChoice = true

	case momentum.StopReasonMaxAutoBeats, momentum.StopReasonFrequency:
		// Normal pause for player input
		log.Println("[Orchestrator] Normal pause for player choice")
		result.RequiresChoice = true

	default:
		// StopReasonNone or unknown - continue normally
		result.RequiresChoice = true
	}

	log.Printf("[Orchestrator] Game loop complete: beats=%d, hp=%d, san=%d, stop_reason=%s",
		result.BeatsResolved, o.gameState.GetHP(), o.gameState.GetSAN(), result.StopReason.String())

	return result, nil
}

// buildNarrativeContext builds a narrative context from current game state
// This is used by MomentumController to decide whether to pause
func (o *Orchestrator) buildNarrativeContext() *momentum.NarrativeContext {
	// For now, return a basic context
	// This will be enhanced when we integrate with actual game state
	return &momentum.NarrativeContext{
		CurrentBeat:              o.gameState.GetCurrentBeat(),
		CurrentScene:             o.gameState.CurrentScene,
		RiskLevel:                momentum.RiskNone, // TODO: Calculate from game state
		IsPlotPoint:              false,             // TODO: Check plot detector
		NPCInitiatesConversation: false,             // TODO: Check NPC state
		PendingEvents:            make([]*momentum.GameEvent, 0),
		RecentChoices:            make([]string, 0),
		AutoResolvedBeats:        0,
	}
}

// ============================================================================
// Story 9-10: Trinity Performance Monitoring
// ============================================================================

// GetTrinityMetrics returns the current Trinity metrics summary
// Story 9-10 AC5: Performance monitoring integration
//
// This method provides access to all Trinity routing metrics including:
//   - Request counts per tier
//   - Success/failure rates
//   - Response time statistics (avg, min, max, percentiles)
//   - Tier transition counts (upgrades/downgrades)
//
// Returns:
//   - trinity.MetricsSummary: Complete metrics summary
//   - error: Error if Trinity router is not initialized
func (o *Orchestrator) GetTrinityMetrics() (trinity.MetricsSummary, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.trinityRouter == nil {
		return trinity.MetricsSummary{}, fmt.Errorf("Trinity router not initialized")
	}

	return o.trinityRouter.GetMetricsSummary(), nil
}

// LogTrinityMetrics logs a comprehensive Trinity performance report
// Story 9-10 AC5: Performance monitoring integration
//
// This method logs:
//   - Total requests and uptime
//   - Per-tier statistics (requests, success rate, response times)
//   - Tier transition counts
//   - Recent errors
//
// The log output is formatted for easy readability and includes:
//   - Request counts and rates
//   - Timing percentiles (P50, P90, P99)
//   - Error information
func (o *Orchestrator) LogTrinityMetrics() {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.trinityRouter == nil {
		log.Println("[Orchestrator] Trinity router not initialized, cannot log metrics")
		return
	}

	log.Println("[Orchestrator] ===== Trinity Performance Report =====")
	o.trinityRouter.LogMetricsSummary()
	log.Println("[Orchestrator] ===== End Trinity Report =====")
}

// ResetTrinityMetrics resets all Trinity metrics
// Story 9-10 AC5: Performance monitoring integration
//
// This is useful for:
//   - Clearing metrics between test runs
//   - Resetting counters at game milestones
//   - Starting fresh performance measurements
func (o *Orchestrator) ResetTrinityMetrics() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.trinityRouter == nil {
		return fmt.Errorf("Trinity router not initialized")
	}

	o.trinityRouter.ResetMetrics()
	log.Println("[Orchestrator] Trinity metrics reset")
	return nil
}

// HasTrinityRouter returns whether the orchestrator has Trinity routing enabled
// Story 9-10: Helper method for checking Trinity availability
//
// Returns:
//   - bool: true if Trinity router is initialized, false otherwise
func (o *Orchestrator) HasTrinityRouter() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.trinityRouter != nil
}
