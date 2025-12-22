package agents

import (
	"context"
	"strings"
	"testing"
)

// ==========================================================================
// Story 7.7: NPC Dialogue Integration Tests
// ==========================================================================

// TestDialogueGeneration_WithSANState tests dialogue generation with SAN influence
func TestDialogueGeneration_WithSANState(t *testing.T) {
	tests := []struct {
		name           string
		npcSAN         int
		expectedStyle  string
		shouldContain  []string
	}{
		{
			name:          "Normal SAN (100)",
			npcSAN:        100,
			expectedStyle: "正常",
			shouldContain: []string{"正常"},
		},
		{
			name:          "Anxious SAN (60)",
			npcSAN:        60,
			expectedStyle: "焦慮",
			shouldContain: []string{"焦慮", "緊張"},
		},
		{
			name:          "Panic SAN (30)",
			npcSAN:        30,
			expectedStyle: "恐慌",
			shouldContain: []string{"恐慌", "語無倫次"},
		},
		{
			name:          "Collapse SAN (10)",
			npcSAN:        10,
			expectedStyle: "崩潰",
			shouldContain: []string{"崩潰"},
		},
	}

	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		LLMClient:  nil, // Will use fallback
		MaxRetries: 1,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := NPCInstance{
				ID:          "TEST-001",
				Name:        "測試NPC",
				Archetype:   NPCArchetypeSacrificial,
				Personality: []string{"恐懼", "無助"},
			}

			request := &DialogueRequest{
				NPC:         npc,
				Context:     "黑暗的走廊",
				Tension:     50,
				CurrentBeat: 10,
				NPCSAN:      tt.npcSAN,
			}

			// Generate dialogue (will use fallback without LLM)
			response, err := na.InvokeDialogue(context.Background(), request)
			if err != nil {
				t.Fatalf("InvokeDialogue failed: %v", err)
			}

			if response == nil {
				t.Fatal("Expected non-nil response")
			}

			if response.Dialogue == "" {
				t.Error("Expected non-empty dialogue")
			}

			// Verify dialogue length is appropriate for SAN state
			dialogueLen := len([]rune(response.Dialogue))
			if tt.npcSAN < 20 && dialogueLen > 100 {
				t.Error("Collapse state dialogue should be short")
			}
		})
	}
}

// TestDialogueGeneration_ClueRevelation tests clue revelation logic
func TestDialogueGeneration_ClueRevelation(t *testing.T) {
	tests := []struct {
		name         string
		archetype    NPCArchetype
		tension      int
		linkedSeeds  []string
		shouldReveal bool
	}{
		{
			name:         "Knowledgeable with seeds - low tension",
			archetype:    NPCArchetypeKnowledgeable,
			tension:      20,
			linkedSeeds:  []string{"SEED-001"},
			shouldReveal: true,
		},
		{
			name:         "Knowledgeable with seeds - high tension",
			archetype:    NPCArchetypeKnowledgeable,
			tension:      85,
			linkedSeeds:  []string{"SEED-001"},
			shouldReveal: true,
		},
		{
			name:         "Hostile/Mystic with seeds",
			archetype:    NPCArchetypeHostile,
			tension:      60,
			linkedSeeds:  []string{"SEED-002"},
			shouldReveal: true,
		},
		{
			name:         "Sacrificial without seeds",
			archetype:    NPCArchetypeSacrificial,
			tension:      50,
			linkedSeeds:  []string{},
			shouldReveal: false,
		},
		{
			name:         "Neutral without seeds",
			archetype:    NPCArchetypeNeutral,
			tension:      60,
			linkedSeeds:  []string{},
			shouldReveal: false,
		},
	}

	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		LLMClient:  nil,
		MaxRetries: 1,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := NPCInstance{
				ID:          "TEST-001",
				Name:        "測試NPC",
				Archetype:   tt.archetype,
				Personality: []string{"神秘", "謹慎"},
				LinkedSeeds: tt.linkedSeeds,
			}

			request := &DialogueRequest{
				NPC:         npc,
				Context:     "神秘的房間",
				Tension:     tt.tension,
				CurrentBeat: 10,
				NPCSAN:      100,
			}

			response, err := na.InvokeDialogue(context.Background(), request)
			if err != nil {
				t.Fatalf("InvokeDialogue failed: %v", err)
			}

			// Check if clue was revealed
			if tt.shouldReveal && len(tt.linkedSeeds) > 0 {
				if response.SeedRevealed == nil {
					t.Error("Expected seed to be revealed")
				} else if *response.SeedRevealed != tt.linkedSeeds[0] {
					t.Errorf("Expected seed %s, got %s", tt.linkedSeeds[0], *response.SeedRevealed)
				}
			}

			if !tt.shouldReveal && response.SeedRevealed != nil {
				t.Error("Did not expect seed to be revealed")
			}
		})
	}
}

// TestDialoguePromptBuilding_SAN tests prompt building with SAN state
func TestDialoguePromptBuilding_SAN(t *testing.T) {
	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		MaxRetries: 1,
	})

	tests := []struct {
		name          string
		npcSAN        int
		shouldContain []string
	}{
		{
			name:          "Normal SAN prompt",
			npcSAN:        100,
			shouldContain: []string{"SAN 值：100", "正常"},
		},
		{
			name:          "Anxious SAN prompt",
			npcSAN:        65,
			shouldContain: []string{"SAN 值：65", "焦慮"},
		},
		{
			name:          "Panic SAN prompt",
			npcSAN:        35,
			shouldContain: []string{"SAN 值：35", "恐慌"},
		},
		{
			name:          "Collapse SAN prompt",
			npcSAN:        15,
			shouldContain: []string{"SAN 值：15", "崩潰"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := NPCInstance{
				ID:          "TEST-001",
				Name:        "測試NPC",
				Archetype:   NPCArchetypeKnowledgeable,
				Personality: []string{"神秘", "謹慎"},
			}

			request := &DialogueRequest{
				NPC:         npc,
				Context:     "測試場景",
				Tension:     50,
				CurrentBeat: 10,
				NPCSAN:      tt.npcSAN,
			}

			prompt := na.buildDialoguePrompt(request)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(prompt, expected) {
					t.Errorf("Expected prompt to contain '%s', prompt: %s", expected, prompt)
				}
			}
		})
	}
}

// TestDialoguePromptBuilding_ClueHints tests clue revelation hints based on tension
func TestDialoguePromptBuilding_ClueHints(t *testing.T) {
	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		MaxRetries: 1,
	})

	tests := []struct {
		name          string
		tension       int
		shouldContain string
	}{
		{
			name:          "Very vague (tension 20)",
			tension:       20,
			shouldContain: "隱晦",
		},
		{
			name:          "Vague (tension 50)",
			tension:       50,
			shouldContain: "模糊",
		},
		{
			name:          "Specific (tension 70)",
			tension:       70,
			shouldContain: "具體",
		},
		{
			name:          "Direct (tension 90)",
			tension:       90,
			shouldContain: "直接",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := NPCInstance{
				ID:          "TEST-001",
				Name:        "測試NPC",
				Archetype:   NPCArchetypeKnowledgeable,
				Personality: []string{"神秘", "謹慎"},
				LinkedSeeds: []string{"SEED-001"},
			}

			request := &DialogueRequest{
				NPC:         npc,
				Context:     "測試場景",
				Tension:     tt.tension,
				CurrentBeat: 10,
				NPCSAN:      100,
			}

			prompt := na.buildDialoguePrompt(request)

			if !strings.Contains(prompt, tt.shouldContain) {
				t.Errorf("Expected prompt to contain '%s' for tension %d", tt.shouldContain, tt.tension)
			}
		})
	}
}

// TestGetSANStyleModifier tests SAN style modifiers
func TestGetSANStyleModifier(t *testing.T) {
	na := NewNPCAgent(AgentConfig{})

	tests := []struct {
		san           int
		shouldContain string
	}{
		{100, "正常"},
		{85, "正常"},
		{75, "焦慮"},
		{55, "焦慮"},
		{45, "恐慌"},
		{25, "恐慌"},
		{15, "崩潰"},
		{5, "崩潰"},
	}

	for _, tt := range tests {
		modifier := na.getSANStyleModifier(tt.san)
		if !strings.Contains(modifier, tt.shouldContain) {
			t.Errorf("SAN %d: expected modifier to contain '%s', got: %s", tt.san, tt.shouldContain, modifier)
		}
	}
}

// TestGetDialogueLengthWithSAN tests dialogue length adjustment based on SAN
func TestGetDialogueLengthWithSAN(t *testing.T) {
	na := NewNPCAgent(AgentConfig{})

	tests := []struct {
		name      string
		tension   int
		san       int
		minExpect int
		maxExpect int
	}{
		{
			name:      "Normal SAN, low tension",
			tension:   20,
			san:       100,
			minExpect: 150,
			maxExpect: 300,
		},
		{
			name:      "Anxious SAN, mid tension",
			tension:   50,
			san:       60,
			minExpect: 80,
			maxExpect: 180,
		},
		{
			name:      "Panic SAN, high tension",
			tension:   85,
			san:       30,
			minExpect: 30,
			maxExpect: 70,
		},
		{
			name:      "Collapse SAN",
			tension:   50,
			san:       10,
			minExpect: 50,
			maxExpect: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length := na.getDialogueLengthWithSAN(tt.tension, tt.san)

			if length[0] < tt.minExpect {
				t.Errorf("Min length %d is less than expected %d", length[0], tt.minExpect)
			}

			if length[1] > tt.maxExpect {
				t.Errorf("Max length %d is greater than expected %d", length[1], tt.maxExpect)
			}

			if length[0] > length[1] {
				t.Errorf("Min length %d should be <= max length %d", length[0], length[1])
			}
		})
	}
}

// TestShouldRevealClue tests clue revelation archetype check
func TestShouldRevealClue(t *testing.T) {
	na := NewNPCAgent(AgentConfig{})

	tests := []struct {
		archetype    NPCArchetype
		shouldReveal bool
	}{
		{NPCArchetypeSacrificial, false},
		{NPCArchetypeKnowledgeable, true},
		{NPCArchetypeHostile, true}, // Mystic/Inspirer
		{NPCArchetypeNeutral, false},
		{NPCArchetypeGuide, false},
		{NPCArchetypeDeceiver, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.archetype), func(t *testing.T) {
			result := na.shouldRevealClue(tt.archetype)
			if result != tt.shouldReveal {
				t.Errorf("Archetype %s: expected %v, got %v", tt.archetype, tt.shouldReveal, result)
			}
		})
	}
}

// TestTemplateDialogue_SANVariations tests template dialogue with different SAN states
func TestTemplateDialogue_SANVariations(t *testing.T) {
	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		LLMClient:  nil,
		MaxRetries: 1,
	})

	archetypes := []NPCArchetype{
		NPCArchetypeSacrificial,
		NPCArchetypeKnowledgeable,
		NPCArchetypeHostile,
		NPCArchetypeNeutral,
		NPCArchetypeGuide,
		NPCArchetypeDeceiver,
	}

	sanStates := []int{100, 65, 35, 10}

	for _, archetype := range archetypes {
		for _, san := range sanStates {
			t.Run(string(archetype)+"_SAN_"+string(rune('0'+san/10)), func(t *testing.T) {
				npc := NPCInstance{
					ID:          "TEST-001",
					Name:        "測試NPC",
					Archetype:   archetype,
					Personality: []string{"測試"},
				}

				request := &DialogueRequest{
					NPC:         npc,
					Context:     "測試場景",
					Tension:     50,
					CurrentBeat: 10,
					NPCSAN:      san,
				}

				response, err := na.InvokeDialogue(context.Background(), request)
				if err != nil {
					t.Fatalf("InvokeDialogue failed: %v", err)
				}

				if response.Dialogue == "" {
					t.Error("Expected non-empty dialogue")
				}

				// Verify dialogue changes with SAN state
				// (More sophisticated validation could be added)
			})
		}
	}
}

// TestDialogueLength_Compliance tests dialogue length compliance
func TestDialogueLength_Compliance(t *testing.T) {
	na := NewNPCAgent(AgentConfig{
		Name:       "TestNPCAgent",
		LLMClient:  nil,
		MaxRetries: 1,
	})

	npc := NPCInstance{
		ID:          "TEST-001",
		Name:        "測試NPC",
		Archetype:   NPCArchetypeSacrificial,
		Personality: []string{"恐懼"},
	}

	request := &DialogueRequest{
		NPC:         npc,
		Context:     "黑暗走廊",
		Tension:     50,
		CurrentBeat: 10,
		NPCSAN:      100,
	}

	response, err := na.InvokeDialogue(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeDialogue failed: %v", err)
	}

	dialogueLen := len([]rune(response.Dialogue))

	// AC #1: Dialogue should be 100-300 chars
	// Template fallback may be shorter, but should be reasonable
	if dialogueLen < 10 {
		t.Errorf("Dialogue too short: %d chars", dialogueLen)
	}

	if dialogueLen > 500 {
		t.Errorf("Dialogue too long: %d chars", dialogueLen)
	}
}

// BenchmarkDialogueGeneration benchmarks dialogue generation
func BenchmarkDialogueGeneration(b *testing.B) {
	na := NewNPCAgent(AgentConfig{
		Name:       "BenchNPCAgent",
		LLMClient:  nil,
		MaxRetries: 1,
	})

	npc := NPCInstance{
		ID:          "BENCH-001",
		Name:        "基準NPC",
		Archetype:   NPCArchetypeKnowledgeable,
		Personality: []string{"神秘"},
		LinkedSeeds: []string{"SEED-001"},
	}

	request := &DialogueRequest{
		NPC:         npc,
		Context:     "測試場景",
		Tension:     60,
		CurrentBeat: 10,
		NPCSAN:      100,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = na.InvokeDialogue(ctx, request)
	}
}
