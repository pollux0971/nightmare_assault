package seed

import (
	"fmt"
	"testing"
)

// TestNewSeedManager tests SeedManager creation.
func TestNewSeedManager(t *testing.T) {
	manager := NewSeedManager()

	if manager == nil {
		t.Fatal("NewSeedManager() returned nil")
	}

	seeds := manager.GetAllActiveGlobalSeeds()
	if seeds == nil {
		t.Error("GetAllActiveGlobalSeeds() should never return nil, expected empty slice")
	}
	if len(seeds) != 0 {
		t.Errorf("New SeedManager should have 0 seeds, got %d", len(seeds))
	}
}

// TestSeedManager_AddGlobalSeed tests adding global seeds.
func TestSeedManager_AddGlobalSeed(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewGlobalSeed(
		"GS001",
		"Test seed 1",
		"Test truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	seed2, _ := NewGlobalSeed(
		"GS002",
		"Test seed 2",
		"Another truth",
		"tragic",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"shadow"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"darkness"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"death"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	manager.AddGlobalSeed(seed1)
	manager.AddGlobalSeed(seed2)

	seeds := manager.GetAllActiveGlobalSeeds()
	if len(seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(seeds))
	}

	// Verify deep copy protection
	seeds[0].CurrentTier = 99
	if manager.GetAllActiveGlobalSeeds()[0].CurrentTier == 99 {
		t.Error("GetAllActiveGlobalSeeds() should return deep copies to prevent external modification")
	}
}

// TestSeedManager_AddGlobalSeed_NilHandling tests nil seed handling (Issue #4 fix).
func TestSeedManager_AddGlobalSeed_NilHandling(t *testing.T) {
	manager := NewSeedManager()

	// Adding nil should return error (no longer silent)
	err := manager.AddGlobalSeed(nil)
	if err == nil {
		t.Error("Expected error when adding nil seed, got nil")
	}

	seeds := manager.GetAllActiveGlobalSeeds()
	if len(seeds) != 0 {
		t.Errorf("Adding nil seed should be rejected, got %d seeds", len(seeds))
	}
}

// TestSeedManager_AddLocalSeed_NilHandling tests nil local seed handling (Issue #4 fix).
func TestSeedManager_AddLocalSeed_NilHandling(t *testing.T) {
	manager := NewSeedManager()

	// Adding nil should return error (no longer silent)
	err := manager.AddLocalSeed(nil)
	if err == nil {
		t.Error("Expected error when adding nil seed, got nil")
	}

	seeds := manager.GetActiveLocalSeeds("")
	if len(seeds) != 0 {
		t.Errorf("Adding nil seed should be rejected, got %d seeds", len(seeds))
	}
}

// TestSeedManager_GetAllActiveGlobalSeeds_EmptyCase tests getting seeds when none exist.
func TestSeedManager_GetAllActiveGlobalSeeds_EmptyCase(t *testing.T) {
	manager := NewSeedManager()

	seeds := manager.GetAllActiveGlobalSeeds()

	if seeds == nil {
		t.Error("GetAllActiveGlobalSeeds() should return empty slice, not nil")
	}

	if len(seeds) != 0 {
		t.Errorf("Expected 0 seeds, got %d", len(seeds))
	}
}

// TestSeedManager_CheckHarvest_NoSeeds tests CheckHarvest with no seeds.
func TestSeedManager_CheckHarvest_NoSeeds(t *testing.T) {
	manager := NewSeedManager()

	instructions := manager.CheckHarvest(5)

	if instructions == nil {
		t.Error("CheckHarvest() should return empty slice, not nil")
	}

	if len(instructions) != 0 {
		t.Errorf("Expected 0 instructions with no seeds, got %d", len(instructions))
	}
}

// TestSeedManager_CheckHarvest_SeedNotReady tests CheckHarvest when seeds not in range.
func TestSeedManager_CheckHarvest_SeedNotReady(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"test"}, BeatStart: 10, BeatEnd: 15},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"test"}, BeatStart: 16, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"test"}, BeatStart: 21, BeatEnd: 25},
		},
	)

	manager.AddGlobalSeed(seed)

	// Check at beat 5 (before Tier 1 starts at beat 10)
	instructions := manager.CheckHarvest(5)

	if len(instructions) != 0 {
		t.Errorf("Expected 0 instructions when seed not ready, got %d", len(instructions))
	}
}

// TestSeedManager_CheckHarvest_SingleSeedReady tests CheckHarvest with one ready seed.
func TestSeedManager_CheckHarvest_SingleSeedReady(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Subtle hint", Keywords: []string{"shadow", "whisper"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Obvious clue", Keywords: []string{"shadow", "truth"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Explicit reveal", Keywords: []string{"shadow", "truth", "entity"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	manager.AddGlobalSeed(seed)

	// Check at beat 3 (within Tier 1 range)
	instructions := manager.CheckHarvest(3)

	if len(instructions) != 1 {
		t.Fatalf("Expected 1 instruction, got %d", len(instructions))
	}

	instr := instructions[0]
	if instr.SeedID != "GS001" {
		t.Errorf("Expected SeedID 'GS001', got '%s'", instr.SeedID)
	}

	if instr.ClueContent != "Subtle hint" {
		t.Errorf("Expected ClueContent 'Subtle hint', got '%s'", instr.ClueContent)
	}

	if instr.Tier != 1 {
		t.Errorf("Expected Tier 1, got %d", instr.Tier)
	}
}

// TestSeedManager_CheckHarvest_MultipleSeeds tests CheckHarvest with multiple ready seeds.
func TestSeedManager_CheckHarvest_MultipleSeeds(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewGlobalSeed(
		"GS001",
		"Seed 1",
		"Truth 1",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Seed 1 Tier 1", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Seed 1 Tier 2", Keywords: []string{"clue"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Seed 1 Tier 3", Keywords: []string{"reveal"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	seed2, _ := NewGlobalSeed(
		"GS002",
		"Seed 2",
		"Truth 2",
		"tragic",
		[]ClueTier{
			{Tier: 1, Content: "Seed 2 Tier 1", Keywords: []string{"shadow"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Seed 2 Tier 2", Keywords: []string{"darkness"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Seed 2 Tier 3", Keywords: []string{"death"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	manager.AddGlobalSeed(seed1)
	manager.AddGlobalSeed(seed2)

	// Check at beat 5 (both seeds ready at Tier 1)
	instructions := manager.CheckHarvest(5)

	if len(instructions) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(instructions))
	}

	// Verify both seeds are included
	foundGS001 := false
	foundGS002 := false
	for _, instr := range instructions {
		if instr.SeedID == "GS001" {
			foundGS001 = true
		}
		if instr.SeedID == "GS002" {
			foundGS002 = true
		}
	}

	if !foundGS001 || !foundGS002 {
		t.Error("Expected both GS001 and GS002 in instructions")
	}
}

// TestSeedManager_CheckHarvest_PrioritySorting tests that instructions are sorted by priority.
func TestSeedManager_CheckHarvest_PrioritySorting(t *testing.T) {
	manager := NewSeedManager()

	// Seed at Tier 1 (high priority - new seed)
	seedTier1, _ := NewGlobalSeed(
		"GS001",
		"High priority seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1 clue", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2 clue", Keywords: []string{"clue"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Tier 3 clue", Keywords: []string{"reveal"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	// Seed at Tier 3 (lower priority - already revealed twice)
	seedTier3, _ := NewGlobalSeed(
		"GS002",
		"Low priority seed",
		"Truth",
		"tragic",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1 clue", Keywords: []string{"shadow"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "Tier 2 clue", Keywords: []string{"darkness"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 3, Content: "Tier 3 clue", Keywords: []string{"death"}, BeatStart: 1, BeatEnd: 10},
		},
	)

	// Advance seedTier3 to tier 3
	_ = seedTier3.AdvanceTier()
	_ = seedTier3.AdvanceTier()

	manager.AddGlobalSeed(seedTier1)
	manager.AddGlobalSeed(seedTier3)

	// Check at beat 5 (both ready - GS001 at tier 1, GS002 at tier 3)
	instructions := manager.CheckHarvest(5)

	if len(instructions) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(instructions))
	}

	// First instruction should be GS001 (Tier 1 = higher priority)
	if instructions[0].SeedID != "GS001" {
		t.Errorf("Expected first instruction to be GS001 (Tier 1, higher priority), got %s", instructions[0].SeedID)
	}

	// Second instruction should be GS002 (Tier 3 = lower priority)
	if instructions[1].SeedID != "GS002" {
		t.Errorf("Expected second instruction to be GS002 (Tier 3, lower priority), got %s", instructions[1].SeedID)
	}

	// Verify priority values reflect the ordering
	if instructions[0].Priority <= instructions[1].Priority {
		t.Errorf("Expected first instruction priority (%d) > second instruction priority (%d)",
			instructions[0].Priority, instructions[1].Priority)
	}
}

// TestSeedManager_MarkSeedRevealed tests marking a seed as revealed.
func TestSeedManager_MarkSeedRevealed(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	manager.AddGlobalSeed(seed)

	// Verify initial tier
	if seed.CurrentTier != 1 {
		t.Fatalf("Expected initial CurrentTier 1, got %d", seed.CurrentTier)
	}

	// Mark tier 1 as revealed
	err := manager.MarkSeedRevealed("GS001", 3)
	if err != nil {
		t.Fatalf("Unexpected error marking seed revealed: %v", err)
	}

	// Verify tier advanced
	seeds := manager.GetAllActiveGlobalSeeds()
	if seeds[0].CurrentTier != 2 {
		t.Errorf("Expected CurrentTier 2 after revealing tier 1, got %d", seeds[0].CurrentTier)
	}
}

// TestSeedManager_MarkSeedRevealed_NotFound tests error when seed not found.
func TestSeedManager_MarkSeedRevealed_NotFound(t *testing.T) {
	manager := NewSeedManager()

	err := manager.MarkSeedRevealed("NONEXISTENT", 5)

	if err == nil {
		t.Error("Expected error when marking non-existent seed, got nil")
	}
}

// TestSeedManager_MarkSeedRevealed_PreventDuplicateInSameBeat tests duplicate prevention.
func TestSeedManager_MarkSeedRevealed_PreventDuplicateInSameBeat(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Tier 1", Keywords: []string{"hint"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Tier 2", Keywords: []string{"clue"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "Tier 3", Keywords: []string{"reveal"}, BeatStart: 13, BeatEnd: 18},
		},
	)

	manager.AddGlobalSeed(seed)

	// Mark revealed at beat 3
	err := manager.MarkSeedRevealed("GS001", 3)
	if err != nil {
		t.Fatalf("First MarkSeedRevealed failed: %v", err)
	}

	// Try to mark revealed again at same beat 3
	err = manager.MarkSeedRevealed("GS001", 3)

	if err == nil {
		t.Error("Expected error when marking seed revealed twice in same beat")
	}
}

// TestSeedManager_GetGlobalSeedsProgress tests progress calculation.
func TestSeedManager_GetGlobalSeedsProgress(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*SeedManager)
		expected float64
	}{
		{
			name:     "empty seeds returns 0",
			setup:    func(sm *SeedManager) {},
			expected: 0.0,
		},
		{
			name: "all seeds at tier 1 returns 0.33",
			setup: func(sm *SeedManager) {
				seed1, _ := NewGlobalSeed("GS001", "Seed 1", "Truth 1", "mysterious",
					[]ClueTier{
						{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
					})
				seed2, _ := NewGlobalSeed("GS002", "Seed 2", "Truth 2", "tragic",
					[]ClueTier{
						{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
					})
				sm.AddGlobalSeed(seed1)
				sm.AddGlobalSeed(seed2)
			},
			expected: 0.333,
		},
		{
			name: "all seeds fully revealed returns 1.0",
			setup: func(sm *SeedManager) {
				seed1, _ := NewGlobalSeed("GS001", "Seed 1", "Truth 1", "mysterious",
					[]ClueTier{
						{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
					})
				seed1.CurrentTier = 3
				sm.AddGlobalSeed(seed1)
			},
			expected: 1.0,
		},
		{
			name: "mixed progress calculates correctly",
			setup: func(sm *SeedManager) {
				seed1, _ := NewGlobalSeed("GS001", "Seed 1", "Truth 1", "mysterious",
					[]ClueTier{
						{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
					})
				seed2, _ := NewGlobalSeed("GS002", "Seed 2", "Truth 2", "tragic",
					[]ClueTier{
						{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
						{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
						{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
					})
				seed1.CurrentTier = 1 // 1/3 = 0.33
				seed2.CurrentTier = 2 // 2/3 = 0.67
				sm.AddGlobalSeed(seed1)
				sm.AddGlobalSeed(seed2)
			},
			expected: 0.5, // (0.33 + 0.67) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSeedManager()
			tt.setup(sm)

			got := sm.GetGlobalSeedsProgress()

			// Use delta comparison for floating point
			delta := 0.01
			if got < tt.expected-delta || got > tt.expected+delta {
				t.Errorf("GetGlobalSeedsProgress() = %v, want %v (±%v)", got, tt.expected, delta)
			}
		})
	}
}

// TestSeedManager_CheckHarvest_LimitPerTurn tests that CheckHarvest limits results.
func TestSeedManager_CheckHarvest_LimitPerTurn(t *testing.T) {
	manager := NewSeedManager()

	// Add 5 seeds all ready at beat 3
	for i := 1; i <= 5; i++ {
		seed, _ := NewGlobalSeed(
			fmt.Sprintf("GS%03d", i),
			fmt.Sprintf("Content %d", i),
			"Truth",
			"mysterious",
			[]ClueTier{
				{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 10},
				{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 11, BeatEnd: 15},
				{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 16, BeatEnd: 20},
			},
		)
		manager.AddGlobalSeed(seed)
	}

	// Check at beat 3 - all 5 seeds are ready
	instructions := manager.CheckHarvest(3)

	// Should be limited to MaxHarvestPerTurn (2)
	if len(instructions) != MaxHarvestPerTurn {
		t.Errorf("CheckHarvest should limit to %d instructions, got %d", MaxHarvestPerTurn, len(instructions))
	}

	// Verify they are sorted by priority (highest first)
	if len(instructions) >= 2 {
		if instructions[0].Priority < instructions[1].Priority {
			t.Error("Instructions should be sorted by priority (descending)")
		}
	}
}

// TestSeedManager_MarkSeedRevealed_Validation tests input validation.
func TestSeedManager_MarkSeedRevealed_Validation(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewGlobalSeed(
		"GS001",
		"Test seed",
		"Truth",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "T1", Keywords: []string{"k1"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "T2", Keywords: []string{"k2"}, BeatStart: 6, BeatEnd: 12},
			{Tier: 3, Content: "T3", Keywords: []string{"k3"}, BeatStart: 13, BeatEnd: 18},
		},
	)
	manager.AddGlobalSeed(seed)

	tests := []struct {
		name        string
		seedID      string
		currentBeat int
		expectError bool
	}{
		{
			name:        "empty seedID returns error",
			seedID:      "",
			currentBeat: 3,
			expectError: true,
		},
		{
			name:        "negative currentBeat returns error",
			seedID:      "GS001",
			currentBeat: -1,
			expectError: true,
		},
		{
			name:        "valid inputs succeed",
			seedID:      "GS001",
			currentBeat: 3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.MarkSeedRevealed(tt.seedID, tt.currentBeat)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// ============================================================================
// Task 5.3: SeedManager LocalSeed Management Tests (10-12 cases)
// ============================================================================

// TestSeedManager_AddLocalSeed tests adding LocalSeeds to the manager.
func TestSeedManager_AddLocalSeed(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-hosp-01", "hospital", "blood stains", "handprint", "You see blood", 10, 5)
	seed2, _ := NewLocalSeed("LS-lib-01", "library", "torn pages", "page 13", "Pages scattered", 15, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)

	// Verify seeds were added
	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 2 {
		t.Errorf("Expected 2 active LocalSeeds, got %d", len(activeSeeds))
	}
}

// TestSeedManager_AddLocalSeed_NilSeed tests that nil seeds are silently ignored.
func TestSeedManager_AddLocalSeed_NilSeed(t *testing.T) {
	manager := NewSeedManager()

	manager.AddLocalSeed(nil)

	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 0 {
		t.Errorf("Expected 0 seeds after adding nil, got %d", len(activeSeeds))
	}
}

// TestSeedManager_GetActiveLocalSeeds_EmptySceneID tests getting all active LocalSeeds.
func TestSeedManager_GetActiveLocalSeeds_EmptySceneID(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-hosp-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	seed2, _ := NewLocalSeed("LS-lib-01", "library", "book", "detail2", "plant2", 15, 5)
	seed3, _ := NewLocalSeed("LS-hosp-02", "hospital", "scratches", "detail3", "plant3", 20, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Get all active seeds (empty sceneID)
	allSeeds := manager.GetActiveLocalSeeds("")
	if len(allSeeds) != 3 {
		t.Errorf("Expected 3 active seeds, got %d", len(allSeeds))
	}
}

// TestSeedManager_GetActiveLocalSeeds_FilterByScene tests scene-specific filtering.
func TestSeedManager_GetActiveLocalSeeds_FilterByScene(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-hosp-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	seed2, _ := NewLocalSeed("LS-lib-01", "library", "book", "detail2", "plant2", 15, 5)
	seed3, _ := NewLocalSeed("LS-hosp-02", "hospital", "scratches", "detail3", "plant3", 20, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Get only hospital seeds
	hospitalSeeds := manager.GetActiveLocalSeeds("hospital")
	if len(hospitalSeeds) != 2 {
		t.Errorf("Expected 2 hospital seeds, got %d", len(hospitalSeeds))
	}

	// Get only library seeds
	librarySeeds := manager.GetActiveLocalSeeds("library")
	if len(librarySeeds) != 1 {
		t.Errorf("Expected 1 library seed, got %d", len(librarySeeds))
	}

	// Get seeds for non-existent scene
	noSeeds := manager.GetActiveLocalSeeds("dungeon")
	if len(noSeeds) != 0 {
		t.Errorf("Expected 0 seeds for non-existent scene, got %d", len(noSeeds))
	}
}

// TestSeedManager_GetActiveLocalSeeds_OnlyActiveStatus tests that only Active seeds are returned.
func TestSeedManager_GetActiveLocalSeeds_OnlyActiveStatus(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	seed2, _ := NewLocalSeed("LS-02", "hospital", "scratches", "detail2", "plant2", 15, 5)
	seed3, _ := NewLocalSeed("LS-03", "hospital", "noise", "detail3", "plant3", 20, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Mark seed2 as harvested
	seed2.Status = SeedStatusHarvested
	// Mark seed3 as pruned
	seed3.Status = SeedStatusPruned

	// Only seed1 should be returned (Active status)
	activeSeeds := manager.GetActiveLocalSeeds("hospital")
	if len(activeSeeds) != 1 {
		t.Errorf("Expected 1 active seed, got %d", len(activeSeeds))
	}

	if activeSeeds[0].ID != "LS-01" {
		t.Errorf("Expected seed LS-01, got %s", activeSeeds[0].ID)
	}
}

// TestSeedManager_GetActiveLocalSeeds_DeepCopy tests that deep copies are returned.
func TestSeedManager_GetActiveLocalSeeds_DeepCopy(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	// Get active seeds
	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 1 {
		t.Fatal("Expected 1 active seed")
	}

	// Modify returned seed
	activeSeeds[0].Status = SeedStatusHarvested

	// Original seed should remain Active
	originalSeeds := manager.GetActiveLocalSeeds("")
	if len(originalSeeds) != 1 {
		t.Fatal("Original seed status was modified - not a deep copy")
	}

	if originalSeeds[0].Status != SeedStatusActive {
		t.Error("Original seed was affected by modifying returned copy")
	}
}

// TestSeedManager_MarkLocalSeedHarvested_Success tests marking a LocalSeed as harvested.
func TestSeedManager_MarkLocalSeedHarvested_Success(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	err := manager.MarkLocalSeedHarvested("LS-01", 12)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify status changed to Harvested
	if seed.Status != SeedStatusHarvested {
		t.Errorf("Expected status=Harvested, got %s", seed.Status)
	}

	// Seed should no longer appear in active seeds
	activeSeeds := manager.GetActiveLocalSeeds("hospital")
	if len(activeSeeds) != 0 {
		t.Errorf("Expected 0 active seeds after harvest, got %d", len(activeSeeds))
	}
}

// TestSeedManager_MarkLocalSeedHarvested_ValidationErrors tests input validation.
func TestSeedManager_MarkLocalSeedHarvested_ValidationErrors(t *testing.T) {
	manager := NewSeedManager()

	tests := []struct {
		name        string
		seedID      string
		currentBeat int
		expectedErr string
	}{
		{
			"empty seedID",
			"",
			10,
			"seedID cannot be empty",
		},
		{
			"negative beat",
			"LS-01",
			-5,
			"invalid currentBeat: -5 (must be >= 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.MarkLocalSeedHarvested(tt.seedID, tt.currentBeat)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestSeedManager_MarkLocalSeedHarvested_SeedNotFound tests error when seed doesn't exist.
func TestSeedManager_MarkLocalSeedHarvested_SeedNotFound(t *testing.T) {
	manager := NewSeedManager()

	err := manager.MarkLocalSeedHarvested("LS-nonexistent", 10)

	if err != ErrSeedNotFound {
		t.Errorf("Expected ErrSeedNotFound, got %v", err)
	}
}

// TestSeedManager_MarkLocalSeedHarvested_NotActiveStatus tests error when seed is not active.
func TestSeedManager_MarkLocalSeedHarvested_NotActiveStatus(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	seed.Status = SeedStatusPruned // Already pruned
	manager.AddLocalSeed(seed)

	err := manager.MarkLocalSeedHarvested("LS-01", 12)

	if err == nil {
		t.Fatal("Expected error for non-active seed, got nil")
	}

	expectedMsg := "cannot harvest seed LS-01: status is pruned (expected active)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestSeedManager_MarkLocalSeedHarvested_TimeTravel tests error when currentBeat < PlantedAt.
func TestSeedManager_MarkLocalSeedHarvested_TimeTravel(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	// Try to harvest at beat 5 (before PlantedAt=10)
	err := manager.MarkLocalSeedHarvested("LS-01", 5)

	if err == nil {
		t.Fatal("Expected error for time-travel (currentBeat < PlantedAt), got nil")
	}

	expectedMsg := "invalid harvest: currentBeat (5) < PlantedAt (10)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify seed status unchanged
	if seed.Status != SeedStatusActive {
		t.Errorf("Expected seed status to remain active, got %s", seed.Status)
	}
}

// TestSeedManager_CheckHarvest_LocalSeeds tests LocalSeed integration into CheckHarvest.
func TestSeedManager_CheckHarvest_LocalSeeds(t *testing.T) {
	manager := NewSeedManager()

	// Add LocalSeeds with different urgency levels
	seed1, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail1", "plant1", 10, 5) // 4 beats left at beat 11 (urgency 20)
	seed2, _ := NewLocalSeed("LS-02", "hospital", "noise", "detail2", "plant2", 8, 5)  // 2 beats left at beat 11 (urgency 60)
	seed3, _ := NewLocalSeed("LS-03", "hospital", "shadow", "detail3", "plant3", 6, 5) // Expired at beat 11 (urgency 100)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Check harvest at beat 11
	instructions := manager.CheckHarvest(11)

	// Should return 2 instructions (seed3 and seed2, both urgency >= 40)
	// seed1 has urgency 20 (below threshold)
	// Limited to MaxHarvestPerTurn = 2
	if len(instructions) != 2 {
		t.Errorf("Expected 2 harvest instructions, got %d", len(instructions))
	}

	// Verify priority ordering (expired seed should be first)
	if len(instructions) > 0 && instructions[0].SeedID != "LS-03" {
		t.Errorf("Expected first instruction to be LS-03 (expired), got %s", instructions[0].SeedID)
	}

	// Verify IsLocalSeed flag is set
	for _, inst := range instructions {
		if !inst.IsLocalSeed {
			t.Errorf("Expected IsLocalSeed=true for instruction %s", inst.SeedID)
		}
	}
}

// ============================================================================
// Task 5.4: Scene Pruning Tests (8-10 cases)
// ============================================================================

// TestSeedManager_PruneLocalSeedsByScene_Success tests successful scene-based pruning.
func TestSeedManager_PruneLocalSeedsByScene_Success(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-hosp-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	seed2, _ := NewLocalSeed("LS-lib-01", "library", "book", "detail2", "plant2", 15, 5)
	seed3, _ := NewLocalSeed("LS-hosp-02", "hospital", "shadow", "detail3", "plant3", 20, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Prune all hospital seeds
	results := manager.PruneLocalSeedsByScene("hospital")

	// Should prune 2 hospital seeds
	if len(results) != 2 {
		t.Errorf("Expected 2 pruned seeds, got %d", len(results))
	}

	// Verify seeds are marked as Pruned
	if seed1.Status != SeedStatusPruned {
		t.Error("Expected seed1 status=Pruned")
	}
	if seed3.Status != SeedStatusPruned {
		t.Error("Expected seed3 status=Pruned")
	}

	// Library seed should remain Active
	if seed2.Status != SeedStatusActive {
		t.Error("Expected seed2 to remain Active")
	}

	// Active seeds should only contain library seed now
	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 1 {
		t.Errorf("Expected 1 active seed remaining, got %d", len(activeSeeds))
	}
}

// TestSeedManager_PruneLocalSeedsByScene_EmptySceneID tests that empty sceneID prunes nothing.
func TestSeedManager_PruneLocalSeedsByScene_EmptySceneID(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	results := manager.PruneLocalSeedsByScene("")

	if len(results) != 0 {
		t.Errorf("Expected 0 pruned seeds for empty sceneID, got %d", len(results))
	}

	if seed.Status != SeedStatusActive {
		t.Error("Seed should remain Active when empty sceneID provided")
	}
}

// TestSeedManager_PruneLocalSeedsByScene_NonExistentScene tests pruning non-existent scene.
func TestSeedManager_PruneLocalSeedsByScene_NonExistentScene(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	results := manager.PruneLocalSeedsByScene("dungeon")

	if len(results) != 0 {
		t.Errorf("Expected 0 pruned seeds for non-existent scene, got %d", len(results))
	}

	if seed.Status != SeedStatusActive {
		t.Error("Seed should remain Active when scene doesn't match")
	}
}

// TestSeedManager_PruneLocalSeedsByScene_OnlyActiveSeeds tests that only Active seeds are pruned.
func TestSeedManager_PruneLocalSeedsByScene_OnlyActiveSeeds(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	seed2, _ := NewLocalSeed("LS-02", "hospital", "noise", "detail2", "plant2", 15, 5)
	seed3, _ := NewLocalSeed("LS-03", "hospital", "shadow", "detail3", "plant3", 20, 5)

	// Mark seed2 as already harvested
	seed2.Status = SeedStatusHarvested
	// Mark seed3 as already pruned
	seed3.Status = SeedStatusPruned

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	results := manager.PruneLocalSeedsByScene("hospital")

	// Should only prune seed1 (the only Active one)
	if len(results) != 1 {
		t.Errorf("Expected 1 pruned seed, got %d", len(results))
	}

	if results[0].SeedID != "LS-01" {
		t.Errorf("Expected pruned seed to be LS-01, got %s", results[0].SeedID)
	}
}

// TestSeedManager_PruneLocalSeedsByScene_PruneResultContent tests PruneResult content.
func TestSeedManager_PruneLocalSeedsByScene_PruneResultContent(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood stains", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	results := manager.PruneLocalSeedsByScene("hospital")

	if len(results) != 1 {
		t.Fatal("Expected 1 prune result")
	}

	result := results[0]

	if result.SeedID != "LS-01" {
		t.Errorf("Expected SeedID=LS-01, got %s", result.SeedID)
	}

	if result.SceneID != "hospital" {
		t.Errorf("Expected SceneID=hospital, got %s", result.SceneID)
	}

	if result.Content != "blood stains" {
		t.Errorf("Expected Content='blood stains', got '%s'", result.Content)
	}

	if result.PruneReason != "scene_change" {
		t.Errorf("Expected PruneReason=scene_change, got %s", result.PruneReason)
	}
}

// TestSeedManager_PruneExpiredLocalSeeds_Success tests successful expiration-based pruning.
func TestSeedManager_PruneExpiredLocalSeeds_Success(t *testing.T) {
	manager := NewSeedManager()

	// PlantedAt=10, MaxLifespan=5, expires at beat 15
	seed1, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail1", "plant1", 10, 5)
	// PlantedAt=12, MaxLifespan=5, expires at beat 17
	seed2, _ := NewLocalSeed("LS-02", "hospital", "noise", "detail2", "plant2", 12, 5)
	// PlantedAt=15, MaxLifespan=5, expires at beat 20
	seed3, _ := NewLocalSeed("LS-03", "library", "shadow", "detail3", "plant3", 15, 5)

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Prune at beat 16 (seed1 expired, seed2 and seed3 not yet expired)
	results := manager.PruneExpiredLocalSeeds(16)

	if len(results) != 1 {
		t.Errorf("Expected 1 expired seed at beat 16, got %d", len(results))
	}

	if results[0].SeedID != "LS-01" {
		t.Errorf("Expected pruned seed LS-01, got %s", results[0].SeedID)
	}

	if seed1.Status != SeedStatusPruned {
		t.Error("Expected seed1 to be Pruned")
	}

	if seed2.Status != SeedStatusActive || seed3.Status != SeedStatusActive {
		t.Error("Expected seed2 and seed3 to remain Active")
	}
}

// TestSeedManager_PruneExpiredLocalSeeds_MultipleExpired tests pruning multiple expired seeds.
func TestSeedManager_PruneExpiredLocalSeeds_MultipleExpired(t *testing.T) {
	manager := NewSeedManager()

	seed1, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail1", "plant1", 10, 5)  // Expires at 15
	seed2, _ := NewLocalSeed("LS-02", "hospital", "noise", "detail2", "plant2", 11, 5)  // Expires at 16
	seed3, _ := NewLocalSeed("LS-03", "library", "shadow", "detail3", "plant3", 12, 5) // Expires at 17

	manager.AddLocalSeed(seed1)
	manager.AddLocalSeed(seed2)
	manager.AddLocalSeed(seed3)

	// Prune at beat 20 (all seeds expired)
	results := manager.PruneExpiredLocalSeeds(20)

	if len(results) != 3 {
		t.Errorf("Expected 3 expired seeds at beat 20, got %d", len(results))
	}

	// All should be Pruned
	if seed1.Status != SeedStatusPruned || seed2.Status != SeedStatusPruned || seed3.Status != SeedStatusPruned {
		t.Error("Expected all seeds to be Pruned")
	}
}

// TestSeedManager_PruneExpiredLocalSeeds_NegativeBeat tests negative beat validation.
func TestSeedManager_PruneExpiredLocalSeeds_NegativeBeat(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	results := manager.PruneExpiredLocalSeeds(-5)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for negative beat, got %d", len(results))
	}

	if seed.Status != SeedStatusActive {
		t.Error("Seed should remain Active for invalid beat")
	}
}

// TestSeedManager_PruneExpiredLocalSeeds_PruneReasonExpired tests prune reason is set correctly.
func TestSeedManager_PruneExpiredLocalSeeds_PruneReasonExpired(t *testing.T) {
	manager := NewSeedManager()

	seed, _ := NewLocalSeed("LS-01", "hospital", "blood stains", "detail", "plant", 10, 5)
	manager.AddLocalSeed(seed)

	results := manager.PruneExpiredLocalSeeds(16)

	if len(results) != 1 {
		t.Fatal("Expected 1 prune result")
	}

	if results[0].PruneReason != "expired" {
		t.Errorf("Expected PruneReason=expired, got %s", results[0].PruneReason)
	}

	// TransitionText should be generated
	if results[0].TransitionText == "" {
		t.Error("Expected non-empty TransitionText for expired seed")
	}
}

// ============================================================================
// Task 5.6: Thread-Safety Tests (4-6 cases)
// IMPORTANT: Run with `go test -race` to detect data races
// ============================================================================

// TestSeedManager_LocalSeed_ConcurrentAdd tests concurrent LocalSeed additions.
func TestSeedManager_LocalSeed_ConcurrentAdd(t *testing.T) {
	manager := NewSeedManager()
	done := make(chan bool)

	// Add 100 seeds concurrently
	for i := 0; i < 100; i++ {
		go func(index int) {
			seed, _ := NewLocalSeed(
				fmt.Sprintf("LS-%d", index),
				"hospital",
				"content",
				"detail",
				"plant",
				10,
				5,
			)
			manager.AddLocalSeed(seed)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify all seeds were added
	seeds := manager.GetActiveLocalSeeds("")
	if len(seeds) != 100 {
		t.Errorf("Expected 100 seeds after concurrent adds, got %d", len(seeds))
	}
}

// TestSeedManager_LocalSeed_ConcurrentReadWrite tests concurrent reads and writes.
func TestSeedManager_LocalSeed_ConcurrentReadWrite(t *testing.T) {
	manager := NewSeedManager()

	// Pre-populate with some seeds
	for i := 0; i < 10; i++ {
		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS-%d", i),
			"hospital",
			"content",
			"detail",
			"plant",
			10,
			5,
		)
		manager.AddLocalSeed(seed)
	}

	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 50; i++ {
		go func() {
			_ = manager.GetActiveLocalSeeds("hospital")
			done <- true
		}()
	}

	// Concurrent writers
	for i := 10; i < 30; i++ {
		go func(index int) {
			seed, _ := NewLocalSeed(
				fmt.Sprintf("LS-%d", index),
				"library",
				"content",
				"detail",
				"plant",
				10,
				5,
			)
			manager.AddLocalSeed(seed)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 70; i++ {
		<-done
	}

	// Should have 30 total seeds (10 initial + 20 new)
	allSeeds := manager.GetActiveLocalSeeds("")
	if len(allSeeds) != 30 {
		t.Errorf("Expected 30 seeds after concurrent operations, got %d", len(allSeeds))
	}
}

// TestSeedManager_LocalSeed_ConcurrentHarvest tests concurrent harvest operations.
func TestSeedManager_LocalSeed_ConcurrentHarvest(t *testing.T) {
	manager := NewSeedManager()

	// Add seeds
	for i := 0; i < 20; i++ {
		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS-%d", i),
			"hospital",
			"content",
			"detail",
			"plant",
			10,
			5,
		)
		manager.AddLocalSeed(seed)
	}

	done := make(chan bool)

	// Try to harvest all seeds concurrently
	for i := 0; i < 20; i++ {
		go func(index int) {
			_ = manager.MarkLocalSeedHarvested(fmt.Sprintf("LS-%d", index), 12)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// All seeds should be harvested (no longer active)
	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 0 {
		t.Errorf("Expected 0 active seeds after concurrent harvest, got %d", len(activeSeeds))
	}
}

// TestSeedManager_LocalSeed_ConcurrentPrune tests concurrent pruning operations.
func TestSeedManager_LocalSeed_ConcurrentPrune(t *testing.T) {
	manager := NewSeedManager()

	// Add seeds to multiple scenes
	for i := 0; i < 30; i++ {
		sceneID := "hospital"
		if i%3 == 0 {
			sceneID = "library"
		} else if i%3 == 1 {
			sceneID = "basement"
		}

		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS-%d", i),
			sceneID,
			"content",
			"detail",
			"plant",
			10,
			5,
		)
		manager.AddLocalSeed(seed)
	}

	done := make(chan bool)

	// Prune multiple scenes concurrently
	go func() {
		manager.PruneLocalSeedsByScene("hospital")
		done <- true
	}()

	go func() {
		manager.PruneLocalSeedsByScene("library")
		done <- true
	}()

	go func() {
		manager.PruneExpiredLocalSeeds(20) // Should prune all (planted at 10, lifespan 5)
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// All seeds should be pruned
	activeSeeds := manager.GetActiveLocalSeeds("")
	if len(activeSeeds) != 0 {
		t.Errorf("Expected 0 active seeds after concurrent prune, got %d", len(activeSeeds))
	}
}

// TestSeedManager_LocalSeed_ConcurrentCheckHarvest tests concurrent CheckHarvest calls.
func TestSeedManager_LocalSeed_ConcurrentCheckHarvest(t *testing.T) {
	manager := NewSeedManager()

	// Add seeds with different urgency levels
	for i := 0; i < 10; i++ {
		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS-%d", i),
			"hospital",
			"content",
			"detail",
			"plant",
			5+i, // Different planted times
			5,
		)
		manager.AddLocalSeed(seed)
	}

	done := make(chan bool)

	// Call CheckHarvest concurrently 50 times
	for i := 0; i < 50; i++ {
		go func() {
			_ = manager.CheckHarvest(12)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// No crashes = success (thread-safe reads)
}

// TestSeedManager_LocalSeed_AggressiveConcurrentMixedOps tests all operations concurrently.
// This is the most aggressive concurrent test - runs Add, Get, Mark, Prune, CheckHarvest simultaneously.
// MUST be run with `go test -race` to detect any race conditions.
func TestSeedManager_LocalSeed_AggressiveConcurrentMixedOps(t *testing.T) {
	manager := NewSeedManager()

	// Pre-populate with initial seeds
	for i := 0; i < 20; i++ {
		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS-initial-%d", i),
			"hospital",
			"content",
			"detail",
			"plant",
			10,
			5,
		)
		manager.AddLocalSeed(seed)
	}

	done := make(chan bool)
	totalOps := 0

	// Operation 1: Concurrent Add (20 goroutines)
	for i := 0; i < 20; i++ {
		go func(index int) {
			seed, _ := NewLocalSeed(
				fmt.Sprintf("LS-add-%d", index),
				"library",
				"content",
				"detail",
				"plant",
				12,
				5,
			)
			manager.AddLocalSeed(seed)
			done <- true
		}(i)
	}
	totalOps += 20

	// Operation 2: Concurrent GetActiveLocalSeeds (30 goroutines)
	for i := 0; i < 30; i++ {
		go func() {
			_ = manager.GetActiveLocalSeeds("hospital")
			done <- true
		}()
	}
	totalOps += 30

	// Operation 3: Concurrent MarkHarvested (10 goroutines)
	for i := 0; i < 10; i++ {
		go func(index int) {
			seedID := fmt.Sprintf("LS-initial-%d", index)
			_ = manager.MarkLocalSeedHarvested(seedID, 15)
			done <- true
		}(i)
	}
	totalOps += 10

	// Operation 4: Concurrent PruneLocalSeedsByScene (5 goroutines)
	for i := 0; i < 5; i++ {
		go func() {
			manager.PruneLocalSeedsByScene("hospital")
			done <- true
		}()
	}
	totalOps += 5

	// Operation 5: Concurrent PruneExpiredLocalSeeds (5 goroutines)
	for i := 0; i < 5; i++ {
		go func() {
			manager.PruneExpiredLocalSeeds(20)
			done <- true
		}()
	}
	totalOps += 5

	// Operation 6: Concurrent CheckHarvest (30 goroutines)
	for i := 0; i < 30; i++ {
		go func() {
			_ = manager.CheckHarvest(15)
			done <- true
		}()
	}
	totalOps += 30

	// Wait for all operations to complete
	for i := 0; i < totalOps; i++ {
		<-done
	}

	// If we got here without panicking, the concurrent operations are safe
	// The exact state is unpredictable due to race conditions, but that's OK
	// What matters is that there are no data races (detected by -race flag)
	t.Log("All concurrent mixed operations completed without crashes")
}

// TestGetGlobalSeedsProgress_EmptyClueChain tests progress calculation with empty ClueChain.
// This test verifies the fix for Issue #2: correct denominator calculation when seeds
// have empty ClueChain arrays.
func TestGetGlobalSeedsProgress_EmptyClueChain(t *testing.T) {
	manager := NewSeedManager()

	// Scenario: 3 global seeds, one with empty ClueChain
	seed1, _ := NewGlobalSeed(
		"GS001",
		"Test seed 1",
		"Truth 1",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Clue 3", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 15},
		},
	)
	seed1.CurrentTier = 2 // 2/3 = 0.667

	seed2, _ := NewGlobalSeed(
		"GS002",
		"Test seed 2 with empty ClueChain",
		"Truth 2",
		"tragic",
		[]ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Clue 3", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 15},
		},
	)
	// Manually create seed with empty ClueChain (bypassing validation for test)
	seed2.ClueChain = []ClueTier{} // Empty - should be treated as 0% progress
	seed2.CurrentTier = 0

	seed3, _ := NewGlobalSeed(
		"GS003",
		"Test seed 3",
		"Truth 3",
		"hopeful",
		[]ClueTier{
			{Tier: 1, Content: "Clue 1", Keywords: []string{"test"}, BeatStart: 1, BeatEnd: 5},
			{Tier: 2, Content: "Clue 2", Keywords: []string{"test"}, BeatStart: 6, BeatEnd: 10},
			{Tier: 3, Content: "Clue 3", Keywords: []string{"test"}, BeatStart: 11, BeatEnd: 15},
		},
	)
	seed3.CurrentTier = 3 // 3/3 = 1.0

	manager.AddGlobalSeed(seed1)
	manager.AddGlobalSeed(seed2)
	manager.AddGlobalSeed(seed3)

	// Calculate progress
	progress := manager.GetGlobalSeedsProgress()

	// Expected: (0.667 + 0.0 + 1.0) / 3 = 0.556
	expected := (2.0/3.0 + 0.0 + 1.0) / 3.0

	if progress < expected-0.01 || progress > expected+0.01 {
		t.Errorf("Expected progress ~%.3f, got %.3f", expected, progress)
	}

	t.Logf("Progress with empty ClueChain: %.3f (expected ~%.3f)", progress, expected)
}

// TestGetGlobalSeedsProgress_AllEmptyClueChains tests edge case of all seeds having empty ClueChain.
func TestGetGlobalSeedsProgress_AllEmptyClueChains(t *testing.T) {
	manager := NewSeedManager()

	// Create seeds with empty ClueChains
	seed1, _ := NewGlobalSeed("GS001", "Test 1", "Truth 1", "mysterious", []ClueTier{
		{Tier: 1, Content: "C1", Keywords: []string{"k"}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "C2", Keywords: []string{"k"}, BeatStart: 6, BeatEnd: 10},
		{Tier: 3, Content: "C3", Keywords: []string{"k"}, BeatStart: 11, BeatEnd: 15},
	})
	seed1.ClueChain = []ClueTier{} // Empty

	seed2, _ := NewGlobalSeed("GS002", "Test 2", "Truth 2", "tragic", []ClueTier{
		{Tier: 1, Content: "C1", Keywords: []string{"k"}, BeatStart: 1, BeatEnd: 5},
		{Tier: 2, Content: "C2", Keywords: []string{"k"}, BeatStart: 6, BeatEnd: 10},
		{Tier: 3, Content: "C3", Keywords: []string{"k"}, BeatStart: 11, BeatEnd: 15},
	})
	seed2.ClueChain = []ClueTier{} // Empty

	manager.AddGlobalSeed(seed1)
	manager.AddGlobalSeed(seed2)

	progress := manager.GetGlobalSeedsProgress()

	// Expected: 0.0 (all seeds have 0% progress)
	if progress != 0.0 {
		t.Errorf("Expected progress 0.0 for all empty ClueChains, got %.3f", progress)
	}
}

// TestGetGlobalSeedsProgress_NoSeeds tests edge case of no seeds.
func TestGetGlobalSeedsProgress_NoSeeds(t *testing.T) {
	manager := NewSeedManager()

	progress := manager.GetGlobalSeedsProgress()

	if progress != 0.0 {
		t.Errorf("Expected progress 0.0 for no seeds, got %.3f", progress)
	}
}

// TestCheckHarvest_GlobalSeedNotCrowdedOut tests Issue #3 fix: slot allocation prevents
// high-priority LocalSeeds from completely crowding out GlobalSeeds.
func TestCheckHarvest_GlobalSeedNotCrowdedOut(t *testing.T) {
	manager := NewSeedManager()

	// Add 2 high-urgency LocalSeeds (will have high priority ~450-500)
	localSeed1, _ := NewLocalSeed("LS001", "hospital", "Urgent local clue 1", "A bloody trail", "You notice blood", 1, 3)
	localSeed2, _ := NewLocalSeed("LS002", "hospital", "Urgent local clue 2", "A moving shadow", "You see a shadow", 1, 3)

	manager.AddLocalSeed(localSeed1)
	manager.AddLocalSeed(localSeed2)

	// Add 1 GlobalSeed (lower priority ~300-400)
	globalSeed, _ := NewGlobalSeed(
		"GS001",
		"Main storyline clue",
		"The truth about the hospital",
		"mysterious",
		[]ClueTier{
			{Tier: 1, Content: "A subtle hint", Keywords: []string{"darkness"}, BeatStart: 1, BeatEnd: 10},
			{Tier: 2, Content: "More obvious", Keywords: []string{"evil"}, BeatStart: 11, BeatEnd: 20},
			{Tier: 3, Content: "Explicit reveal", Keywords: []string{"truth"}, BeatStart: 21, BeatEnd: 30},
		},
	)

	manager.AddGlobalSeed(globalSeed)

	// Current beat: 5 (both LocalSeeds are urgent, GlobalSeed is ready)
	instructions := manager.CheckHarvest(5)

	// Verify: Should return 2 instructions (MaxHarvestPerTurn)
	if len(instructions) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(instructions))
	}

	// Verify: GlobalSeed should NOT be crowded out (should be in slot 1)
	hasGlobal := false
	for _, inst := range instructions {
		if !inst.IsLocalSeed {
			hasGlobal = true
			// Should be the first instruction (slot 1)
			if instructions[0].SeedID != inst.SeedID {
				t.Logf("Warning: GlobalSeed not in slot 1, but present in results")
			}
			break
		}
	}

	if !hasGlobal {
		t.Error("FAILED Issue #3 fix: GlobalSeed was completely crowded out by LocalSeeds")
	} else {
		t.Log("PASSED Issue #3 fix: GlobalSeed protected from being crowded out")
	}

	// Verify: Should have 1 GlobalSeed and 1 LocalSeed
	globalCount := 0
	localCount := 0
	for _, inst := range instructions {
		if inst.IsLocalSeed {
			localCount++
		} else {
			globalCount++
		}
	}

	if globalCount != 1 {
		t.Errorf("Expected 1 GlobalSeed, got %d", globalCount)
	}
	if localCount != 1 {
		t.Errorf("Expected 1 LocalSeed, got %d", localCount)
	}
}

// TestCheckHarvest_SlotAllocation_OnlyGlobalSeeds tests slot allocation with only GlobalSeeds.
func TestCheckHarvest_SlotAllocation_OnlyGlobalSeeds(t *testing.T) {
	manager := NewSeedManager()

	// Add 3 GlobalSeeds
	for i := 1; i <= 3; i++ {
		seed, _ := NewGlobalSeed(
			fmt.Sprintf("GS%03d", i),
			fmt.Sprintf("Global seed %d", i),
			"Truth",
			"mysterious",
			[]ClueTier{
				{Tier: 1, Content: "Clue 1", Keywords: []string{"k"}, BeatStart: 1, BeatEnd: 20},
				{Tier: 2, Content: "Clue 2", Keywords: []string{"k"}, BeatStart: 21, BeatEnd: 40},
				{Tier: 3, Content: "Clue 3", Keywords: []string{"k"}, BeatStart: 41, BeatEnd: 60},
			},
		)
		manager.AddGlobalSeed(seed)
	}

	instructions := manager.CheckHarvest(5)

	// Should return 2 GlobalSeeds (slot 1 + remaining slot filled)
	// When no LocalSeeds available, GlobalSeeds can fill both slots
	if len(instructions) != 2 {
		t.Errorf("Expected 2 instructions when only GlobalSeeds available, got %d", len(instructions))
	}

	// All should be GlobalSeeds
	for i, inst := range instructions {
		if inst.IsLocalSeed {
			t.Errorf("Expected GlobalSeed at position %d, got LocalSeed", i)
		}
	}
}

// TestCheckHarvest_SlotAllocation_OnlyLocalSeeds tests slot allocation with only LocalSeeds.
func TestCheckHarvest_SlotAllocation_OnlyLocalSeeds(t *testing.T) {
	manager := NewSeedManager()

	// Add 3 urgent LocalSeeds
	for i := 1; i <= 3; i++ {
		seed, _ := NewLocalSeed(
			fmt.Sprintf("LS%03d", i),
			"scene1",
			fmt.Sprintf("Local seed %d", i),
			"Detail text",
			"Plant text",
			1,
			3, // Short lifespan -> high urgency at beat 5
		)
		manager.AddLocalSeed(seed)
	}

	instructions := manager.CheckHarvest(5)

	// Should return 2 LocalSeeds (slot 2 + remaining slot filled)
	// When no GlobalSeeds available, LocalSeeds can fill both slots
	if len(instructions) != 2 {
		t.Errorf("Expected 2 instructions when only LocalSeeds available, got %d", len(instructions))
	}

	// All should be LocalSeeds
	for i, inst := range instructions {
		if !inst.IsLocalSeed {
			t.Errorf("Expected LocalSeed at position %d, got GlobalSeed", i)
		}
	}
}
