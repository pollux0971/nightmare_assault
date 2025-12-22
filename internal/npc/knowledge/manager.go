package knowledge

import (
	"sync"
)

// UpdateManager manages the information isolation engine, tracking global facts,
// NPC knowledge, player knowledge, and room occupancy for information propagation.
// It ensures thread-safe access to all knowledge bases and room tracking.
type UpdateManager struct {
	// globalFacts stores all facts that have occurred in the game world
	// Key: Fact ID, Value: Fact
	globalFacts map[string]*Fact

	// npcKnowledge stores each NPC's individual knowledge base
	// Key: NPC ID, Value: KnowledgeBase
	npcKnowledge map[string]*KnowledgeBase

	// playerKnowledge stores the player's knowledge base
	playerKnowledge *KnowledgeBase

	// roomOccupants tracks which entities (NPCs + player) are in which rooms
	// Key: Room ID, Value: Set of entity IDs (NPC IDs or "player")
	roomOccupants map[string]map[string]bool

	// config holds the configuration parameters
	config *UpdateManagerConfig

	// contradictionAnalyzer performs LLM-based semantic contradiction analysis
	// Story 8.2: Semantic contradiction detection
	contradictionAnalyzer *ContradictionAnalyzer

	// distortionCalculator performs intelligent information distortion
	// Story 8.3: Information distortion optimization
	distortionCalculator *DistortionCalculator

	// mu protects concurrent access to all fields
	mu sync.RWMutex
}

// UpdateManagerConfig contains configuration parameters for the UpdateManager.
// These parameters control information propagation, distortion, and other
// knowledge system mechanics.
type UpdateManagerConfig struct {
	// EnableDistortion controls whether information can be distorted during propagation.
	// When true, facts passed between NPCs may be altered based on their mental state.
	// Default: true
	EnableDistortion bool

	// DistortionRate is the base probability (0.0-1.0) that information becomes distorted
	// when propagated. Actual distortion chance increases with NPC stress/fear levels.
	// Default: 0.15 (15% base chance)
	DistortionRate float64

	// MaxPropagationDepth limits how many times a fact can be passed between NPCs
	// before it can no longer be shared. This prevents infinite information chains.
	// Depth 0 = original witness, Depth 1 = first-hand account, etc.
	// Default: 3 (fact can be told up to 3 times)
	MaxPropagationDepth int

	// ContradictionAnalyzer is the LLM-based semantic contradiction analyzer
	// Story 8.2: If nil, falls back to keyword-based contradiction detection
	ContradictionAnalyzer *ContradictionAnalyzer

	// DistortionCalculator is the intelligent information distortion calculator
	// Story 8.3: If nil, falls back to simple random distortion
	DistortionCalculator *DistortionCalculator
}

// DefaultUpdateManagerConfig returns an UpdateManagerConfig with sensible default values.
// These defaults are tuned for realistic information propagation in horror game settings.
func DefaultUpdateManagerConfig() *UpdateManagerConfig {
	return &UpdateManagerConfig{
		EnableDistortion:    true,
		DistortionRate:      0.15, // 15% base distortion chance
		MaxPropagationDepth: 3,    // Facts can be told up to 3 times
	}
}

// NewUpdateManager creates a new UpdateManager with the given configuration.
// If config is nil, it uses DefaultUpdateManagerConfig().
//
// All maps are initialized properly:
// - globalFacts: stores all facts in the game world
// - npcKnowledge: stores individual knowledge bases for each NPC
// - playerKnowledge: initialized with empty knowledge base for player
// - roomOccupants: tracks entity locations for information propagation
func NewUpdateManager(config *UpdateManagerConfig) *UpdateManager {
	if config == nil {
		config = DefaultUpdateManagerConfig()
	}

	return &UpdateManager{
		globalFacts:           make(map[string]*Fact),
		npcKnowledge:          make(map[string]*KnowledgeBase),
		playerKnowledge:       NewKnowledgeBase("player"),
		roomOccupants:         make(map[string]map[string]bool),
		config:                config,
		contradictionAnalyzer: config.ContradictionAnalyzer,
		distortionCalculator:  config.DistortionCalculator,
	}
}

// GetConfig returns the manager's configuration.
func (m *UpdateManager) GetConfig() *UpdateManagerConfig {
	return m.config
}

// SetNPCManager sets an external NPC manager reference for accessing NPC profiles.
// This allows the UpdateManager to access NPC emotional state and traits for
// intelligent distortion calculation without creating circular dependencies.
//
// Story 8.3: Integration with NPC Manager for distortion calculation
func (m *UpdateManager) SetNPCManager(npcMgr NPCManagerInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If a distortion calculator is configured and doesn't have a state provider,
	// wrap the NPC manager as a state provider
	if m.distortionCalculator != nil && npcMgr != nil {
		adapter := &npcManagerAdapter{npcMgr: npcMgr}
		m.distortionCalculator = NewDistortionCalculator(adapter, m.distortionCalculator.GetConfig())
	}
}

// NPCManagerInterface defines the minimal interface needed from the NPC manager.
// This prevents circular dependencies between knowledge and manager packages.
type NPCManagerInterface interface {
	// GetNPCEmotion returns the emotional state for a given NPC ID.
	GetNPCEmotion(npcID string) (trust, fear, stress int, err error)

	// GetNPCTraits returns the trait IDs for a given NPC.
	GetNPCTraits(npcID string) ([]string, error)
}

// npcManagerAdapter adapts NPCManagerInterface to NPCStateProvider.
type npcManagerAdapter struct {
	npcMgr NPCManagerInterface
}

func (a *npcManagerAdapter) GetNPCEmotion(npcID string) (trust, fear, stress int, err error) {
	return a.npcMgr.GetNPCEmotion(npcID)
}

func (a *npcManagerAdapter) GetNPCTraits(npcID string) ([]string, error) {
	return a.npcMgr.GetNPCTraits(npcID)
}
