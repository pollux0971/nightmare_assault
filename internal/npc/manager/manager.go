package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/npc/knowledge"
)

// NPCManager manages all NPC profiles and their runtime states.
// It provides thread-safe CRUD operations and ensures proper state management.
type NPCManager struct {
	// profiles stores complete NPC profiles (including hidden information)
	profiles map[string]*NPCProfile

	// states stores runtime state for each NPC
	states map[string]*NPCRuntimeState

	// updateMgr manages information propagation and contradiction detection
	// Can be nil for testing or simplified scenarios
	updateMgr *knowledge.UpdateManager

	// config holds the configuration parameters
	config *NPCManagerConfig

	// mu protects concurrent access to profiles and states
	mu sync.RWMutex
}

// NewNPCManager creates a new NPCManager with the given configuration and UpdateManager.
// If config is nil, it uses DefaultNPCManagerConfig().
// updateMgr can be nil for testing or simplified scenarios.
func NewNPCManager(updateMgr *knowledge.UpdateManager, config *NPCManagerConfig) *NPCManager {
	if config == nil {
		config = DefaultNPCManagerConfig()
	}

	return &NPCManager{
		profiles:  make(map[string]*NPCProfile),
		states:    make(map[string]*NPCRuntimeState),
		updateMgr: updateMgr,
		config:    config,
	}
}

// AddNPC adds a new NPC profile and initializes its runtime state.
// Returns an error if an NPC with the same ID already exists.
func (m *NPCManager) AddNPC(profile *NPCProfile) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	if profile.ID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if NPC already exists
	if _, exists := m.profiles[profile.ID]; exists {
		return fmt.Errorf("NPC with ID %s already exists", profile.ID)
	}

	// Store profile
	m.profiles[profile.ID] = profile

	// Initialize runtime state from profile's initial emotion
	state := NewNPCRuntimeState()
	state.Emotion = profile.InitialEmotion
	state.Relationship = CalculateRelationship(state.Emotion)
	state.RelationshipScore = CalculateRelationshipScore(state.Emotion)

	// Check initial mental state based on emotion
	state.MentalState = checkMentalStateTransition(state.MentalState, state.Emotion)

	// Initialize trait states for all traits in profile
	for _, trait := range profile.Traits {
		state.TraitStates[trait.ID] = "hidden" // Default to hidden
	}

	m.states[profile.ID] = state

	return nil
}

// GetProfile retrieves an NPC profile by ID.
// Returns nil if the NPC doesn't exist.
func (m *NPCManager) GetProfile(npcID string) *NPCProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.profiles[npcID]
}

// GetState retrieves an NPC's runtime state by ID.
// Returns nil if the NPC doesn't exist.
func (m *NPCManager) GetState(npcID string) *NPCRuntimeState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.states[npcID]
}

// UpdateState updates an NPC's runtime state.
// Returns an error if the NPC doesn't exist.
func (m *NPCManager) UpdateState(npcID string, state *NPCRuntimeState) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.states[npcID]; !exists {
		return fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	m.states[npcID] = state
	return nil
}

// DeleteNPC removes an NPC and its runtime state.
// Returns an error if the NPC doesn't exist.
func (m *NPCManager) DeleteNPC(npcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.profiles[npcID]; !exists {
		return fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	delete(m.profiles, npcID)
	delete(m.states, npcID)

	return nil
}

// ListNPCIDs returns a list of all NPC IDs currently managed.
func (m *NPCManager) ListNPCIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.profiles))
	for id := range m.profiles {
		ids = append(ids, id)
	}
	return ids
}

// GetConfig returns the manager's configuration.
func (m *NPCManager) GetConfig() *NPCManagerConfig {
	return m.config
}

// AdjustEmotion applies an emotional delta to an NPC and triggers all cascade updates.
// This is the primary method for updating NPC emotional state during gameplay.
//
// The method performs the following actions in order (cascade update pattern):
// 1. Applies the emotion delta to current emotion values (clamped to 0-100)
// 2. Recalculates relationship type and score
// 3. Checks for mental state transitions
// 4. Checks for trait reveals (placeholder in Story 1.4)
//
// Returns an error if the NPC does not exist.
func (m *NPCManager) AdjustEmotion(npcID string, delta EmotionDelta) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if NPC exists
	state, exists := m.states[npcID]
	if !exists {
		return fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	// 1. Apply emotion delta (with clamping)
	state.Emotion = state.Emotion.Apply(delta)

	// 2. Recalculate relationship type and score
	state.Relationship = CalculateRelationship(state.Emotion)
	state.RelationshipScore = CalculateRelationshipScore(state.Emotion)

	// 3. Check for mental state transitions
	state.MentalState = checkMentalStateTransition(state.MentalState, state.Emotion)

	// 4. Check for trait reveals (placeholder for Story 1.6)
	checkTraitReveal(npcID, "emotion_change")

	return nil
}

// checkTraitReveal is a placeholder function for trait revelation logic.
// This will be fully implemented in Story 1.6.
// For now, it simply returns nil.
func checkTraitReveal(npcID string, event string) *Trait {
	// Placeholder implementation - Story 1.6 will implement full logic
	return nil
}

// ProcessChatMessage handles chat message processing with integrated information
// propagation and contradiction detection.
//
// It performs the following steps:
// 1. Gets all NPCs in the same room as the speaker
// 2. Propagates information to all listeners via UpdateManager.LearnFromDialogue
// 3. Checks for contradictions with existing knowledge
// 4. Automatically adjusts emotions when contradictions are detected
// 5. Records contradiction events in NPC interaction history
//
// Parameters:
//   - speakerID: The entity ID of the speaker
//   - content: The message content
//   - currentRoom: The current room location
//
// Returns:
//   - []knowledge.ContradictionResult: All contradictions detected across listeners
//
// AC2: ProcessChatMessage() 整合資訊傳播與矛盾檢測
// AC3: 矛盾檢測後自動調整 NPC 情感
func (m *NPCManager) ProcessChatMessage(
	speakerID string,
	content string,
	currentRoom string,
) []knowledge.ContradictionResult {
	// Graceful degradation when no UpdateManager available
	if m.updateMgr == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Get all NPCs in the same room (excluding the speaker)
	listeners := m.updateMgr.GetNPCsInSameRoom(speakerID)

	contradictions := make([]knowledge.ContradictionResult, 0)

	for _, listenerID := range listeners {
		// 1. Propagate information to listener
		m.updateMgr.LearnFromDialogue(listenerID, speakerID, content, currentRoom)

		// 2. Check for contradictions
		contradiction := m.updateMgr.CheckContradiction(listenerID, content)
		if contradiction != nil {
			// 3. Automatically adjust emotion based on contradiction severity
			delta := EmotionDelta{
				Trust:  contradiction.SuggestedDelta.Trust,
				Fear:   contradiction.SuggestedDelta.Fear,
				Stress: contradiction.SuggestedDelta.Stress,
			}
			m.adjustEmotionLocked(listenerID, delta)

			// 4. Record contradiction event in interaction history
			state := m.states[listenerID]
			if state != nil {
				interaction := NPCInteraction{
					Timestamp:       time.Now(),
					InteractionType: "contradiction",
					EmotionDelta:    delta,
					Description:     fmt.Sprintf("矛盾: %s (來自 %s)", content, speakerID),
				}
				state.Interactions = append(state.Interactions, interaction)
				state.LastInteraction = interaction.Timestamp
			}

			contradictions = append(contradictions, *contradiction)
		}
	}

	return contradictions
}

// adjustEmotionLocked is an internal version of AdjustEmotion that assumes
// the mutex is already locked. This is used by ProcessChatMessage to avoid
// deadlock when the lock is already held.
//
// This method performs the same operations as AdjustEmotion:
// 1. Applies emotion delta (with clamping via Emotion.Apply)
// 2. Recalculates relationship type and score
// 3. Checks for mental state transitions
//
// Parameters:
//   - npcID: The NPC ID
//   - delta: The emotion change to apply
func (m *NPCManager) adjustEmotionLocked(npcID string, delta EmotionDelta) {
	state := m.states[npcID]
	if state == nil {
		return // Silently ignore if NPC doesn't exist
	}

	// 1. Apply emotion delta (using Emotion.Apply for clamping)
	state.Emotion = state.Emotion.Apply(delta)

	// 2. Recalculate relationship type and score
	state.Relationship = CalculateRelationship(state.Emotion)
	state.RelationshipScore = CalculateRelationshipScore(state.Emotion)

	// 3. Check for mental state transitions
	state.MentalState = checkMentalStateTransition(state.MentalState, state.Emotion)
}

// getCurrentBeat is a placeholder function that returns the current game beat.
// In a real implementation, this would be retrieved from the Orchestrator or GameState.
func getCurrentBeat() int {
	return 0 // Placeholder - will be properly implemented when integrating with Orchestrator
}

// RecordInteraction records an interaction to NPC history.
// Story 4.6: Emotion Update Integration - AC4: Emotion changes recorded to NPCInteraction history.
//
// This method is thread-safe and appends the interaction to the NPC's history.
// The history is automatically limited to the last 100 interactions to prevent unbounded growth.
//
// Parameters:
//   - npcID: The NPC ID
//   - interaction: The interaction record to add
//
// Returns an error if the NPC does not exist.
func (m *NPCManager) RecordInteraction(npcID string, interaction NPCInteraction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[npcID]
	if !ok {
		return fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	state.Interactions = append(state.Interactions, interaction)
	state.LastInteraction = interaction.Timestamp

	// Limit history length (keep last 100)
	if len(state.Interactions) > 100 {
		state.Interactions = state.Interactions[len(state.Interactions)-100:]
	}

	return nil
}

// ==========================================================================
// Story 8.1: Trait Revelation Integration
// ==========================================================================

// CheckTraitRevelation evaluates trait revelation conditions for an NPC and progresses
// traits that meet their revelation criteria.
//
// Story 8.1 AC3 & AC4: Multi-factor trait revelation with interaction acceleration
//
// This method should be called after significant player interactions to check if
// any traits should progress to the next revelation phase.
//
// Parameters:
//   - npcID: The NPC to evaluate
//   - playerMessage: The player's message (for acceleration checks)
//   - currentBeat: Current game beat/time
//
// Returns:
//   - map[string]string: Map from trait ID to new status (for traits that progressed)
//   - error: If NPC doesn't exist
func (m *NPCManager) CheckTraitRevelation(npcID string, playerMessage string, currentBeat int) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile := m.profiles[npcID]
	state := m.states[npcID]

	if profile == nil || state == nil {
		return nil, fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	// Build RevealContext
	context := RevealContext{
		CurrentBeat:      currentBeat,
		RecentEvents:     []string{}, // Could be populated from game state
		InteractionCount: len(state.Interactions),
	}

	// Convert basic Traits to TraitFull for evaluation
	// In a real implementation, profile.Traits would already be []TraitFull
	traits := make([]TraitFull, len(profile.Traits))
	for i, basicTrait := range profile.Traits {
		traits[i] = FromBasicTrait(basicTrait)
		traits[i].Status = state.GetTraitStatus(basicTrait.ID)

		// Get interaction count for this specific trait
		// For now, use total interaction count as approximation
		traits[i].InteractionCount = len(state.Interactions)
	}

	// Check for accelerated revelation
	for i := range traits {
		trait := &traits[i]

		// Skip if already revealed
		if trait.Status == Revealed {
			continue
		}

		// Check if acceleration conditions are met
		if AccelerateTraitRevelation(trait, state, context, playerMessage) {
			// Immediately progress to next phase
			state.RevealTrait(trait.ID)
		}
	}

	// Use progressive trait checking with multi-factor scoring
	progressedTraits := CheckAndProgressTraits(traits, state, context)

	return progressedTraits, nil
}

// ==========================================================================
// Story 8.3: NPC State Provider Interface Implementation
// ==========================================================================

// GetNPCEmotion returns the emotional state for a given NPC ID.
// This implements the NPCManagerInterface for knowledge.UpdateManager integration.
//
// Story 8.3: Provides NPC emotional state for intelligent distortion calculation
//
// Returns:
//   - trust, fear, stress: Emotion values (0-100)
//   - error: If NPC doesn't exist
func (m *NPCManager) GetNPCEmotion(npcID string) (trust, fear, stress int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[npcID]
	if !exists {
		return 0, 0, 0, fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	return state.Emotion.Trust, state.Emotion.Fear, state.Emotion.Stress, nil
}

// GetNPCTraits returns the trait IDs for a given NPC.
// This implements the NPCManagerInterface for knowledge.UpdateManager integration.
//
// Story 8.3: Provides NPC personality traits for intelligent distortion calculation
//
// Returns:
//   - []string: List of trait IDs
//   - error: If NPC doesn't exist
func (m *NPCManager) GetNPCTraits(npcID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile, exists := m.profiles[npcID]
	if !exists {
		return nil, fmt.Errorf("NPC with ID %s does not exist", npcID)
	}

	// Extract trait IDs from profile
	traitIDs := make([]string, len(profile.Traits))
	for i, trait := range profile.Traits {
		traitIDs[i] = trait.ID
	}

	return traitIDs, nil
}
