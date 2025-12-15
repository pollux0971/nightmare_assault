package seed

import (
	"fmt"
	"log"
	"sort"
	"sync"
)

// MaxHarvestPerTurn limits the total number of seeds that can be revealed in a single turn.
// This prevents narrative overload and ensures clues are naturally woven into the story.
const MaxHarvestPerTurn = 2

// MaxGlobalSeedsPerTurn limits the number of GlobalSeeds per turn.
// Guarantees at least one GlobalSeed can be revealed (main storyline progression).
const MaxGlobalSeedsPerTurn = 1

// MaxLocalSeedsPerTurn limits the number of LocalSeeds per turn.
// Guarantees at least one LocalSeed can be revealed (scene-specific urgency).
const MaxLocalSeedsPerTurn = 1

// SeedManager manages the lifecycle of both Global Seeds and Local Seeds throughout the game.
// It tracks all active seeds, checks revelation timing, and generates harvest instructions
// for the Narration Agent.
//
// Thread-Safety: All methods are thread-safe using read/write mutex.
type SeedManager struct {
	globalSeeds      []*GlobalSeed
	localSeeds       []*LocalSeed
	lastRevealedBeat map[string]int // seedID -> beat number of last revelation
	mu               sync.RWMutex
}

// NewSeedManager creates a new SeedManager instance.
//
// Returns:
//   - *SeedManager: A new seed manager with empty seed lists
func NewSeedManager() *SeedManager {
	return &SeedManager{
		globalSeeds:      make([]*GlobalSeed, 0),
		localSeeds:       make([]*LocalSeed, 0),
		lastRevealedBeat: make(map[string]int),
	}
}

// AddGlobalSeed adds a Global Seed to the manager.
// Returns error if seed is nil (Issue #4 fix: no longer silent).
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seed: The GlobalSeed to add (must not be nil)
//
// Returns:
//   - error: If seed is nil
func (sm *SeedManager) AddGlobalSeed(seed *GlobalSeed) error {
	if seed == nil {
		log.Printf("[WARN] Attempted to add nil GlobalSeed - rejected")
		return fmt.Errorf("cannot add nil GlobalSeed")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.globalSeeds = append(sm.globalSeeds, seed)
	log.Printf("[DEBUG] Added GlobalSeed %s (total: %d)", seed.ID, len(sm.globalSeeds))

	return nil
}

// GetAllActiveGlobalSeeds returns a deep copy of all Global Seeds.
// Returns deep copies to prevent external modification of internal state.
//
// Thread-Safety: This method is thread-safe.
//
// Returns:
//   - []*GlobalSeed: Deep copies of all global seeds (never nil, empty slice if no seeds)
func (sm *SeedManager) GetAllActiveGlobalSeeds() []*GlobalSeed {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return deep copies to protect internal state
	copies := make([]*GlobalSeed, len(sm.globalSeeds))
	for i, seed := range sm.globalSeeds {
		copies[i] = seed.DeepCopy()
	}

	return copies
}

// CheckHarvest checks all Global Seeds and Local Seeds, generating HarvestInstructions
// for seeds that are ready to be revealed at the current beat.
//
// Slot Allocation Algorithm (Issue #3 Fix):
//  1. Collect GlobalSeed instructions (check readiness)
//  2. Collect LocalSeed instructions (check urgency >= 40)
//  3. Sort each list independently by priority (descending)
//  4. Allocate slots to prevent crowding out:
//     a. First slot: Highest-priority GlobalSeed (guarantees main storyline)
//     b. Second slot: Highest-priority LocalSeed (if any, and space available)
//     c. Remaining slots: Fill from combined pool by priority
//  5. Total limit: MaxHarvestPerTurn (prevents narrative overload)
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - []*HarvestInstruction: Sorted list of harvest instructions (never nil, empty if no seeds ready)
func (sm *SeedManager) CheckHarvest(currentBeat int) []*HarvestInstruction {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var globalInstructions []*HarvestInstruction
	var localInstructions []*HarvestInstruction

	// 1. Collect Global Seeds
	for _, seed := range sm.globalSeeds {
		// Skip fully revealed seeds
		if seed.IsFullyRevealed() {
			log.Printf("[DEBUG] Skipped GlobalSeed %s: fully revealed (tier %d)", seed.ID, seed.CurrentTier)
			continue
		}

		// Try to create harvest instruction
		instruction, err := NewHarvestInstruction(seed, currentBeat)
		if err != nil {
			// Seed not ready or no clue available - log and skip
			log.Printf("[DEBUG] Skipped GlobalSeed %s at beat %d: %v", seed.ID, currentBeat, err)
			continue
		}

		globalInstructions = append(globalInstructions, instruction)
	}

	// 2. Collect Local Seeds
	for _, seed := range sm.localSeeds {
		// Skip non-active seeds (harvested or pruned)
		if seed.Status != SeedStatusActive {
			log.Printf("[DEBUG] Skipped LocalSeed %s: status is %s (not active)", seed.ID, seed.Status)
			continue
		}

		// Only harvest LocalSeeds that meet forced harvest threshold (urgency >= 40)
		// This prevents narrative overload from too many LocalSeeds
		if !seed.ShouldForceHarvest(currentBeat) {
			log.Printf("[DEBUG] Skipped LocalSeed %s: urgency %d below threshold (remaining: %d beats)",
				seed.ID, seed.CalculateUrgency(currentBeat), seed.GetRemainingLifespan(currentBeat))
			continue
		}

		// Try to create harvest instruction
		instruction, err := NewLocalSeedHarvestInstruction(seed, currentBeat)
		if err != nil {
			// Seed not ready - log and skip
			log.Printf("[DEBUG] Skipped LocalSeed %s at beat %d: %v", seed.ID, currentBeat, err)
			continue
		}

		log.Printf("[DEBUG] Added LocalSeed %s to harvest queue (urgency: %d, priority: %d)",
			seed.ID, seed.CalculateUrgency(currentBeat), instruction.Priority)
		localInstructions = append(localInstructions, instruction)
	}

	// 3. Sort each list independently by priority
	sortByPriority(globalInstructions)
	sortByPriority(localInstructions)

	// 4. Slot allocation to prevent crowding out (Issue #3 Fix)
	result := make([]*HarvestInstruction, 0, MaxHarvestPerTurn)

	// Slot 1: Prioritize GlobalSeed to ensure main storyline progression
	if len(globalInstructions) > 0 {
		result = append(result, globalInstructions[0])
		log.Printf("[DEBUG] Allocated slot 1 (GlobalSeed): %s (priority: %d)",
			globalInstructions[0].SeedID, globalInstructions[0].Priority)
	}

	// Slot 2: Add LocalSeed if available and space remains
	if len(localInstructions) > 0 && len(result) < MaxHarvestPerTurn {
		result = append(result, localInstructions[0])
		log.Printf("[DEBUG] Allocated slot 2 (LocalSeed): %s (priority: %d)",
			localInstructions[0].SeedID, localInstructions[0].Priority)
	}

	// Fill remaining slots from combined pool (if any space left)
	if len(result) < MaxHarvestPerTurn {
		// Combine remaining candidates
		remaining := make([]*HarvestInstruction, 0)
		if len(globalInstructions) > 1 {
			remaining = append(remaining, globalInstructions[1:]...)
		}
		if len(localInstructions) > 1 {
			remaining = append(remaining, localInstructions[1:]...)
		}

		// Sort remaining by priority
		sortByPriority(remaining)

		// Fill remaining slots
		for _, inst := range remaining {
			if len(result) >= MaxHarvestPerTurn {
				break
			}
			result = append(result, inst)
			log.Printf("[DEBUG] Allocated extra slot: %s (priority: %d, type: %s)",
				inst.SeedID, inst.Priority, getInstructionTypeString(inst))
		}
	}

	if len(result) > 0 {
		log.Printf("[DEBUG] CheckHarvest at beat %d: returning %d instructions (GlobalSeeds: %d, LocalSeeds: %d)",
			currentBeat, len(result),
			countInstructionsByType(result, false),
			countInstructionsByType(result, true))
	}

	return result
}

// sortByPriority sorts harvest instructions by priority in descending order.
// Uses Go's standard sort package (Issue #6 fix: replaced bubble sort).
//
// Performance: O(n log n) - more efficient than bubble sort O(n²),
// especially beneficial when LocalSeeds count grows.
//
// Parameters:
//   - instructions: The slice to sort in-place
func sortByPriority(instructions []*HarvestInstruction) {
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].Priority > instructions[j].Priority // Descending order
	})
}

// countInstructionsByType counts how many instructions are of a specific type.
//
// Parameters:
//   - instructions: The slice of harvest instructions
//   - isLocal: true to count LocalSeeds, false to count GlobalSeeds
//
// Returns:
//   - int: Count of instructions matching the specified type
func countInstructionsByType(instructions []*HarvestInstruction, isLocal bool) int {
	count := 0
	for _, instruction := range instructions {
		if instruction.IsLocalSeed == isLocal {
			count++
		}
	}
	return count
}

// getInstructionTypeString returns a human-readable string for the instruction type.
//
// Parameters:
//   - inst: The harvest instruction
//
// Returns:
//   - string: "GlobalSeed" or "LocalSeed"
func getInstructionTypeString(inst *HarvestInstruction) string {
	if inst.IsLocalSeed {
		return "LocalSeed"
	}
	return "GlobalSeed"
}

// MarkSeedRevealed marks a Global Seed as having revealed its current tier.
// This advances the seed to the next tier and records the revelation beat.
//
// Prevents duplicate revelations in the same beat - returns error if attempting
// to reveal the same seed twice in one beat.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seedID: The ID of the seed to mark as revealed
//   - currentBeat: The beat number when revelation occurred
//
// Returns:
//   - error: If seed not found, or if already revealed in this beat
func (sm *SeedManager) MarkSeedRevealed(seedID string, currentBeat int) error {
	// Input validation
	if seedID == "" {
		return fmt.Errorf("seedID cannot be empty")
	}
	if currentBeat < 0 {
		return fmt.Errorf("invalid currentBeat: %d (must be >= 0)", currentBeat)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find the seed
	var targetSeed *GlobalSeed
	for _, seed := range sm.globalSeeds {
		if seed.ID == seedID {
			targetSeed = seed
			break
		}
	}

	if targetSeed == nil {
		return ErrSeedNotFound
	}

	// Check if already revealed in this beat
	if lastBeat, exists := sm.lastRevealedBeat[seedID]; exists && lastBeat == currentBeat {
		return ErrAlreadyRevealedThisBeat
	}

	// Advance the tier
	err := targetSeed.AdvanceTier()
	if err != nil {
		// Seed is already fully revealed
		return err
	}

	// Record the revelation beat
	sm.lastRevealedBeat[seedID] = currentBeat

	return nil
}

// GetGlobalSeedsProgress calculates the overall revelation progress of all Global Seeds.
// Returns a value between 0.0 (no seeds revealed) and 1.0 (all seeds fully revealed).
//
// Progress Calculation:
//   - Each seed contributes: CurrentTier / TotalTiers
//   - Seeds with empty ClueChain are treated as 0% progress
//   - Overall progress: Average of all seed progress values
//   - Empty seed list returns 0.0
//
// Thread-Safety: This method is thread-safe.
//
// Returns:
//   - float64: Progress value in range [0.0, 1.0]
func (sm *SeedManager) GetGlobalSeedsProgress() float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.globalSeeds) == 0 {
		return 0.0
	}

	var totalProgress float64
	validSeedCount := 0

	for _, seed := range sm.globalSeeds {
		// Each seed's progress = CurrentTier / TotalTiers
		totalTiers := len(seed.ClueChain)
		if totalTiers == 0 {
			// Empty ClueChain treated as 0% progress, but still counted
			validSeedCount++
			continue
		}

		// CurrentTier starts at 1, fully revealed means CurrentTier == len(ClueChain)
		seedProgress := float64(seed.CurrentTier) / float64(totalTiers)
		totalProgress += seedProgress
		validSeedCount++
	}

	if validSeedCount == 0 {
		return 0.0
	}

	return totalProgress / float64(validSeedCount)
}

// AddLocalSeed adds a Local Seed to the manager.
// Returns error if seed is nil (Issue #4 fix: no longer silent).
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seed: The LocalSeed to add (must not be nil)
//
// Returns:
//   - error: If seed is nil
func (sm *SeedManager) AddLocalSeed(seed *LocalSeed) error {
	if seed == nil {
		log.Printf("[WARN] Attempted to add nil LocalSeed - rejected")
		return fmt.Errorf("cannot add nil LocalSeed")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.localSeeds = append(sm.localSeeds, seed)
	log.Printf("[DEBUG] Added LocalSeed %s (total: %d)", seed.ID, len(sm.localSeeds))

	return nil
}

// GetActiveLocalSeeds returns deep copies of all active Local Seeds for a specific scene.
// Returns deep copies to prevent external modification of internal state.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - sceneID: The scene ID to filter by (empty string returns all active local seeds)
//
// Returns:
//   - []*LocalSeed: Deep copies of matching active local seeds (never nil, empty slice if none found)
func (sm *SeedManager) GetActiveLocalSeeds(sceneID string) []*LocalSeed {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*LocalSeed, 0)

	for _, seed := range sm.localSeeds {
		// Filter by scene and status
		if seed.Status == SeedStatusActive {
			// If sceneID is specified, filter by it; otherwise return all active seeds
			if sceneID == "" || seed.SceneID == sceneID {
				result = append(result, seed.DeepCopy())
			}
		}
	}

	return result
}

// MarkLocalSeedHarvested marks a Local Seed as harvested.
// Updates the seed's status and records the harvest beat.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - seedID: The ID of the seed to mark as harvested
//   - currentBeat: The beat number when harvest occurred
//
// Returns:
//   - error: If seed not found or if seed is not active
func (sm *SeedManager) MarkLocalSeedHarvested(seedID string, currentBeat int) error {
	// Input validation
	if seedID == "" {
		return fmt.Errorf("seedID cannot be empty")
	}
	if currentBeat < 0 {
		return fmt.Errorf("invalid currentBeat: %d (must be >= 0)", currentBeat)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find the seed
	var targetSeed *LocalSeed
	for _, seed := range sm.localSeeds {
		if seed.ID == seedID {
			targetSeed = seed
			break
		}
	}

	if targetSeed == nil {
		return ErrSeedNotFound
	}

	// Check if seed is active
	if targetSeed.Status != SeedStatusActive {
		return fmt.Errorf("cannot harvest seed %s: status is %s (expected active)", seedID, targetSeed.Status)
	}

	// Validate time consistency (prevent time-travel logic errors)
	if currentBeat < targetSeed.PlantedAt {
		return fmt.Errorf("invalid harvest: currentBeat (%d) < PlantedAt (%d)", currentBeat, targetSeed.PlantedAt)
	}

	// Mark as harvested
	targetSeed.Status = SeedStatusHarvested

	// Record the harvest beat
	sm.lastRevealedBeat[seedID] = currentBeat

	log.Printf("[DEBUG] Marked LocalSeed %s as harvested at beat %d", seedID, currentBeat)

	return nil
}

// PruneLocalSeedsByScene prunes all active LocalSeeds belonging to a specific scene.
// This is typically called when the player moves to a new scene, removing scene-specific
// foreshadowing that is no longer relevant.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - sceneID: The scene ID whose seeds should be pruned (empty string prunes nothing)
//
// Returns:
//   - []PruneResult: List of pruned seeds with details (empty if no seeds pruned)
func (sm *SeedManager) PruneLocalSeedsByScene(sceneID string) []PruneResult {
	// Validate input
	if sceneID == "" {
		return []PruneResult{}
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	results := make([]PruneResult, 0)

	// Find and prune all active seeds for the specified scene
	for _, seed := range sm.localSeeds {
		// Only prune active seeds that match the scene
		if seed.Status == SeedStatusActive && seed.SceneID == sceneID {
			// Mark as pruned
			seed.Status = SeedStatusPruned

			// Create prune result
			result := PruneResult{
				SeedID:         seed.ID,
				SceneID:        seed.SceneID,
				Content:        seed.Content,
				PruneReason:    "scene_change",
				TransitionText: "", // Can be enhanced later with narrative transitions
			}
			results = append(results, result)

			log.Printf("[DEBUG] Pruned LocalSeed %s (scene: %s, reason: scene_change)", seed.ID, sceneID)
		}
	}

	if len(results) > 0 {
		log.Printf("[DEBUG] PruneLocalSeedsByScene(%s): pruned %d seeds", sceneID, len(results))
	}

	return results
}

// PruneExpiredLocalSeeds prunes all LocalSeeds that have exceeded their MaxLifespan.
// This prevents accumulation of expired seeds and ensures stale foreshadowing is removed.
//
// Thread-Safety: This method is thread-safe.
//
// Parameters:
//   - currentBeat: The current game beat number
//
// Returns:
//   - []PruneResult: List of pruned seeds with details (empty if no seeds pruned)
func (sm *SeedManager) PruneExpiredLocalSeeds(currentBeat int) []PruneResult {
	// Validate input
	if currentBeat < 0 {
		return []PruneResult{}
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	results := make([]PruneResult, 0)

	// Find and prune all expired active seeds
	for _, seed := range sm.localSeeds {
		// Only prune active seeds that are expired
		if seed.Status == SeedStatusActive && seed.IsExpired(currentBeat) {
			// Mark as pruned
			seed.Status = SeedStatusPruned

			// Create prune result
			result := PruneResult{
				SeedID:      seed.ID,
				SceneID:     seed.SceneID,
				Content:     seed.Content,
				PruneReason: "expired",
				TransitionText: fmt.Sprintf("A fleeting moment passes, the opportunity to notice \"%s\" has vanished.",
					seed.Content),
			}
			results = append(results, result)

			log.Printf("[DEBUG] Pruned LocalSeed %s (scene: %s, reason: expired at beat %d, planted at %d, lifespan %d)",
				seed.ID, seed.SceneID, currentBeat, seed.PlantedAt, seed.MaxLifespan)
		}
	}

	if len(results) > 0 {
		log.Printf("[DEBUG] PruneExpiredLocalSeeds(beat %d): pruned %d seeds", currentBeat, len(results))
	}

	return results
}
