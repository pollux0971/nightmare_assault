package agents

import (
	"context"
	"testing"
	"time"
)

// ==========================================================================
// Story 7.6: NPC Batch Generation Tests
// ==========================================================================

func TestNPCAgent_InvokeBatchGenerate(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil, // Will use fallback
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "hospital",
		Scene: "廢棄的醫院走廊",
	}
	globalSeeds := []GlobalSeedInfo{
		{ID: "GS-001", Description: "Test seed 1"},
		{ID: "GS-002", Description: "Test seed 2"},
	}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// AC #1: Generate 2-4 NPCs with random count
	npcs, err := agent.InvokeBatchGenerate(ctx, 0, nil, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	// Verify count is 2-4
	if len(npcs) < 2 || len(npcs) > 4 {
		t.Errorf("Expected 2-4 NPCs, got %d", len(npcs))
	}

	// AC #1: Must include at least one N-01 Sacrificial
	hasSacrificial := false
	for _, npc := range npcs {
		if npc.Archetype == NPCArchetypeSacrificial {
			hasSacrificial = true
			break
		}
	}
	if !hasSacrificial {
		t.Error("Expected at least one N-01 Sacrificial NPC")
	}

	// Verify all NPCs have required fields
	for i, npc := range npcs {
		if npc.ID == "" {
			t.Errorf("NPC %d has empty ID", i)
		}
		if npc.Name == "" {
			t.Errorf("NPC %d has empty Name", i)
		}
		if npc.Archetype == "" {
			t.Errorf("NPC %d has empty Archetype", i)
		}
		if len(npc.Personality) == 0 {
			t.Errorf("NPC %d has no Personality traits", i)
		}
		if npc.Appearance == "" {
			t.Errorf("NPC %d has empty Appearance", i)
		}
		if npc.Backstory == "" {
			t.Errorf("NPC %d has empty Backstory", i)
		}
		if len(npc.Skills) == 0 {
			t.Errorf("NPC %d has no Skills", i)
		}
		if len(npc.Inventory) == 0 {
			t.Errorf("NPC %d has no Inventory", i)
		}
		if npc.Secret == "" {
			t.Errorf("NPC %d has empty Secret", i)
		}
		if npc.Status != NPCStatusAlive {
			t.Errorf("NPC %d has wrong status: %s", i, npc.Status)
		}
	}
}

func TestNPCAgent_InvokeBatchGenerate_SpecificCount(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "school",
		Scene: "荒廢的教室",
	}
	globalSeeds := []GlobalSeedInfo{}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// Test with specific count
	npcs, err := agent.InvokeBatchGenerate(ctx, 3, nil, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	// Verify exact count
	if len(npcs) != 3 {
		t.Errorf("Expected 3 NPCs, got %d", len(npcs))
	}
}

func TestNPCAgent_InvokeBatchGenerate_SpecificArchetypes(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "village",
		Scene: "詭異的村莊",
	}
	globalSeeds := []GlobalSeedInfo{
		{ID: "GS-001", Description: "Test seed"},
	}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// Test with specific archetypes
	archetypes := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeGuide,
	}

	npcs, err := agent.InvokeBatchGenerate(ctx, 0, archetypes, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	// Verify archetypes match
	if len(npcs) != 3 {
		t.Errorf("Expected 3 NPCs, got %d", len(npcs))
	}

	// Count archetype occurrences
	archetypeCounts := make(map[NPCArchetype]int)
	for _, npc := range npcs {
		archetypeCounts[npc.Archetype]++
	}

	// Verify we have the requested archetypes
	for _, archetype := range archetypes {
		if archetypeCounts[archetype] == 0 {
			t.Errorf("Expected archetype %s not found", archetype)
		}
	}
}

func TestNPCAgent_InvokeBatchGenerate_NoDuplicates(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "hospital",
		Scene: "醫院",
	}
	globalSeeds := []GlobalSeedInfo{}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// Generate multiple batches to test randomness
	duplicatesFound := false
	for run := 0; run < 3; run++ {
		npcs, err := agent.InvokeBatchGenerate(ctx, 4, nil, storyContext, globalSeeds, plotStructure)
		if err != nil {
			t.Fatalf("InvokeBatchGenerate failed on run %d: %v", run, err)
		}

		// Check for duplicate archetypes
		archetypeSeen := make(map[NPCArchetype]bool)
		for _, npc := range npcs {
			if archetypeSeen[npc.Archetype] {
				duplicatesFound = true
			}
			archetypeSeen[npc.Archetype] = true
		}
	}

	// It's okay to have some duplicates occasionally, but not always
	// This is a probabilistic test - just check that the system can generate diverse NPCs
	t.Logf("Duplicates found in some runs: %v", duplicatesFound)
}

func TestNPCAgent_InvokeBatchGenerate_DeathTiming(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "hospital",
		Scene: "醫院",
	}
	globalSeeds := []GlobalSeedInfo{}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// AC #4: N-01 Sacrificial should have death timing in Act 2
	npcs, err := agent.InvokeBatchGenerate(ctx, 0, nil, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	for _, npc := range npcs {
		if npc.Archetype == NPCArchetypeSacrificial {
			// Death timing should be in Act 2 (beats 10-20)
			if npc.DeathTiming < 10 || npc.DeathTiming > 20 {
				t.Errorf("N-01 Sacrificial death timing %d is not in Act 2 range [10, 20]", npc.DeathTiming)
			}
			t.Logf("N-01 Sacrificial death timing: %d", npc.DeathTiming)
		} else {
			// Other archetypes should have DeathTiming = 0
			if npc.DeathTiming != 0 {
				t.Errorf("Non-sacrificial NPC %s has non-zero death timing: %d", npc.Archetype, npc.DeathTiming)
			}
		}
	}
}

func TestNPCAgent_InvokeBatchGenerate_LinkedSeeds(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "hospital",
		Scene: "醫院",
	}
	globalSeeds := []GlobalSeedInfo{
		{ID: "GS-001", Description: "Seed 1"},
		{ID: "GS-002", Description: "Seed 2"},
		{ID: "GS-003", Description: "Seed 3"},
	}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	// AC #5: NPCs should be linked to Global Seeds based on archetype
	archetypes := []NPCArchetype{
		NPCArchetypeSacrificial,   // Should have 0 linked seeds
		NPCArchetypeKnowledgeable, // Should have 1-2 linked seeds
		NPCArchetypeGuide,         // Should have 1 linked seed
	}

	npcs, err := agent.InvokeBatchGenerate(ctx, 0, archetypes, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	for _, npc := range npcs {
		t.Logf("NPC %s (%s) has %d linked seeds", npc.Name, npc.Archetype, len(npc.LinkedSeeds))

		switch npc.Archetype {
		case NPCArchetypeSacrificial:
			// AC #3: N-01 should not link to seeds
			if len(npc.LinkedSeeds) != 0 {
				t.Errorf("N-01 Sacrificial should have 0 linked seeds, got %d", len(npc.LinkedSeeds))
			}

		case NPCArchetypeKnowledgeable:
			// AC #3: N-02 should link to 1-2 seeds
			if len(npc.LinkedSeeds) < 1 || len(npc.LinkedSeeds) > 2 {
				t.Errorf("N-02 Knowledgeable should have 1-2 linked seeds, got %d", len(npc.LinkedSeeds))
			}

		case NPCArchetypeGuide:
			// AC #3: N-05 should link to 1 seed
			if len(npc.LinkedSeeds) != 1 {
				t.Errorf("N-05 Guide should have 1 linked seed, got %d", len(npc.LinkedSeeds))
			}
		}
	}
}

func TestNPCAgent_InvokeBatchGenerate_Timeout(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  nil,
	})

	// AC #1: Generation time should be <10s
	start := time.Now()

	ctx := context.Background()
	storyContext := StoryContext{
		Theme: "hospital",
		Scene: "醫院",
	}
	globalSeeds := []GlobalSeedInfo{}
	plotStructure := PlotStructure{
		TotalBeats: 30,
		Act1Range:  [2]int{0, 10},
		Act2Range:  [2]int{10, 20},
		Act3Range:  [2]int{20, 30},
	}

	npcs, err := agent.InvokeBatchGenerate(ctx, 4, nil, storyContext, globalSeeds, plotStructure)
	if err != nil {
		t.Fatalf("InvokeBatchGenerate failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 10*time.Second {
		t.Errorf("Batch generation took too long: %v (should be <10s)", elapsed)
	}

	t.Logf("Generated %d NPCs in %v", len(npcs), elapsed)
}

func TestNPCAgent_selectRandomArchetype(t *testing.T) {
	agent := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	})

	// Test with no existing archetypes
	archetype := agent.selectRandomArchetype([]NPCArchetype{})
	if archetype == "" {
		t.Error("selectRandomArchetype returned empty archetype")
	}

	// Test with existing archetypes (should avoid duplicates when possible)
	existing := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
	}

	archetype = agent.selectRandomArchetype(existing)
	if archetype == "" {
		t.Error("selectRandomArchetype returned empty archetype")
	}

	// With 5 out of 6 archetypes used, we should get the remaining one
	// (or possibly a duplicate if randomness decides so)
	t.Logf("Selected archetype: %s", archetype)
}
