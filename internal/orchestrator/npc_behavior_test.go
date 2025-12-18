package orchestrator

import (
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.7: NPC Behavior System Tests
// ==========================================================================

// TestDetermineAction_Normal tests normal SAN behavior (80-100)
func TestDetermineAction_Normal(t *testing.T) {
	tests := []struct {
		name            string
		archetype       agents.NPCArchetype
		expectedAction  NPCAction
		shouldContain   string
	}{
		{
			name:           "N-01 Sacrificial - Impulsive",
			archetype:      agents.NPCArchetypeSacrificial,
			expectedAction: NPCActionImpulsive,
			shouldContain:  "衝動",
		},
		{
			name:           "N-02 Knowledgeable - Cautious",
			archetype:      agents.NPCArchetypeKnowledgeable,
			expectedAction: NPCActionCautious,
			shouldContain:  "謹慎",
		},
		{
			name:           "N-03 Hostile/Mystic - Intuitive",
			archetype:      agents.NPCArchetypeHostile,
			expectedAction: NPCActionIntuitive,
			shouldContain:  "直覺",
		},
		{
			name:           "N-04 Neutral/Betrayer - Self-Preserving",
			archetype:      agents.NPCArchetypeNeutral,
			expectedAction: NPCActionSelfPreserving,
			shouldContain:  "安全距離", // LogEntry: 保持安全距離
		},
		{
			name:           "N-05 Guide/Burden - Frozen",
			archetype:      agents.NPCArchetypeGuide,
			expectedAction: NPCActionFrozen,
			shouldContain:  "凍結", // LogEntry: 被恐懼凍結
		},
		{
			name:           "N-06 Deceiver/Silent - Executive",
			archetype:      agents.NPCArchetypeDeceiver,
			expectedAction: NPCActionExecutive,
			shouldContain:  "執行",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NPCProfile{
				Name:      "測試NPC",
				Archetype: tt.archetype,
			}

			situation := Situation{
				Description:    "面臨危險情況",
				DangerLevel:    50,
				RequiresChoice: true,
			}

			result := DetermineAction(npc, 100, situation)

			if result.Action != tt.expectedAction {
				t.Errorf("Expected action %v, got %v", tt.expectedAction, result.Action)
			}

			// LogEntry contains the technical behavior keyword for verification
			if !strings.Contains(result.LogEntry, tt.shouldContain) {
				t.Errorf("Expected LogEntry to contain '%s', got: %s", tt.shouldContain, result.LogEntry)
			}

			if result.Description == "" {
				t.Error("Expected non-empty description")
			}

			if result.LogEntry == "" {
				t.Error("Expected non-empty log entry")
			}
		})
	}
}

// TestDetermineAction_SANStates tests different SAN state behaviors
func TestDetermineAction_SANStates(t *testing.T) {
	npc := &NPCProfile{
		Name:      "測試NPC",
		Archetype: agents.NPCArchetypeSacrificial,
	}

	situation := Situation{
		Description:    "危險情況",
		DangerLevel:    60,
		RequiresChoice: true,
	}

	tests := []struct {
		name           string
		san            int
		expectedAction NPCAction
	}{
		{"Normal SAN", 100, NPCActionImpulsive},
		{"Normal SAN boundary", 80, NPCActionImpulsive},
		{"Anxious SAN", 65, NPCActionImpulsive},
		{"Panic SAN", 35, NPCActionPanic},
		{"Collapse SAN", 15, NPCActionCollapse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineAction(npc, tt.san, situation)

			if result.Action != tt.expectedAction {
				t.Errorf("SAN %d: expected action %v, got %v", tt.san, tt.expectedAction, result.Action)
			}
		})
	}
}

// TestGetBehaviorPatternDescription tests behavior pattern descriptions
func TestGetBehaviorPatternDescription(t *testing.T) {
	tests := []struct {
		archetype     agents.NPCArchetype
		shouldContain string
	}{
		{agents.NPCArchetypeSacrificial, "衝動"},
		{agents.NPCArchetypeKnowledgeable, "謹慎"},
		{agents.NPCArchetypeHostile, "直覺"},
		{agents.NPCArchetypeNeutral, "自保"},
		{agents.NPCArchetypeGuide, "僵直"},
		{agents.NPCArchetypeDeceiver, "執行"},
	}

	for _, tt := range tests {
		t.Run(string(tt.archetype), func(t *testing.T) {
			desc := GetBehaviorPatternDescription(tt.archetype)

			if desc == "" {
				t.Error("Expected non-empty description")
			}

			if !strings.Contains(desc, tt.shouldContain) {
				t.Errorf("Expected description to contain '%s', got: %s", tt.shouldContain, desc)
			}
		})
	}
}

// TestDetermineAction_AllArchetypes verifies all archetypes return valid results
func TestDetermineAction_AllArchetypes(t *testing.T) {
	archetypes := []agents.NPCArchetype{
		agents.NPCArchetypeSacrificial,
		agents.NPCArchetypeKnowledgeable,
		agents.NPCArchetypeHostile,
		agents.NPCArchetypeNeutral,
		agents.NPCArchetypeGuide,
		agents.NPCArchetypeDeceiver,
	}

	situation := Situation{
		Description:    "通用測試情境",
		DangerLevel:    60,
		RequiresChoice: true,
	}

	for _, archetype := range archetypes {
		t.Run(string(archetype), func(t *testing.T) {
			npc := &NPCProfile{
				Name:      "測試NPC",
				Archetype: archetype,
			}

			result := DetermineAction(npc, 100, situation)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.Action == "" {
				t.Error("Expected non-empty action")
			}

			if result.Description == "" {
				t.Error("Expected non-empty description")
			}

			if result.Consequence == "" {
				t.Error("Expected non-empty consequence")
			}

			if result.LogEntry == "" {
				t.Error("Expected non-empty log entry")
			}

			// Verify NPC name appears in description or log
			if !strings.Contains(result.Description, npc.Name) && !strings.Contains(result.LogEntry, npc.Name) {
				t.Error("Expected NPC name to appear in description or log entry")
			}
		})
	}
}

// BenchmarkDetermineAction benchmarks the behavior determination
func BenchmarkDetermineAction(b *testing.B) {
	npc := &NPCProfile{
		Name:      "基準測試NPC",
		Archetype: agents.NPCArchetypeKnowledgeable,
	}

	situation := Situation{
		Description:    "測試情境",
		DangerLevel:    60,
		RequiresChoice: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetermineAction(npc, 100, situation)
	}
}
