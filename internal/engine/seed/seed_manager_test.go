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

// TestSeedManager_AddGlobalSeed_NilHandling tests nil seed handling.
func TestSeedManager_AddGlobalSeed_NilHandling(t *testing.T) {
	manager := NewSeedManager()

	// Adding nil should not panic
	manager.AddGlobalSeed(nil)

	seeds := manager.GetAllActiveGlobalSeeds()
	if len(seeds) != 0 {
		t.Errorf("Adding nil seed should be ignored, got %d seeds", len(seeds))
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
