package npc

import (
	"testing"
)

func TestGenerateTeammates(t *testing.T) {
	tests := []struct {
		name          string
		storyLength   string
		expectedMin   int
		expectedMax   int
	}{
		{"Short story", "short", 1, 1},
		{"Medium story", "medium", 2, 3},
		{"Long story", "long", 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teammates := GenerateTeammates(tt.storyLength, "medium")

			if len(teammates) < tt.expectedMin || len(teammates) > tt.expectedMax {
				t.Errorf("%s: expected %d-%d teammates, got %d",
					tt.name, tt.expectedMin, tt.expectedMax, len(teammates))
			}

			// Verify each teammate has required fields
			for i, tm := range teammates {
				if tm.ID == "" {
					t.Errorf("Teammate %d missing ID", i)
				}
				if tm.Name == "" {
					t.Errorf("Teammate %d missing Name", i)
				}
				if tm.Archetype == "" {
					t.Errorf("Teammate %d missing Archetype", i)
				}
				if tm.HP <= 0 {
					t.Errorf("Teammate %d has invalid HP: %d", i, tm.HP)
				}
			}
		})
	}
}

func TestArchetypeDiversity(t *testing.T) {
	teammates := GenerateTeammates("long", "hard")

	if len(teammates) < 2 {
		t.Skip("Need at least 2 teammates for diversity test")
	}

	// Check that archetypes are diverse (no exact duplicates in archetype)
	archetypeMap := make(map[NPCArchetype]int)
	for _, tm := range teammates {
		archetypeMap[tm.Archetype]++
	}

	// Most archetypes should appear only once
	duplicates := 0
	for _, count := range archetypeMap {
		if count > 1 {
			duplicates++
		}
	}

	// Allow at most 1 duplicate archetype
	if duplicates > 1 {
		t.Errorf("Too many duplicate archetypes: %d types duplicated", duplicates)
	}
}

func TestArchetypeTemplate(t *testing.T) {
	archetypes := []NPCArchetype{
		ArchetypeVictim,
		ArchetypeUnreliable,
		ArchetypeLogic,
		ArchetypeIntuition,
		ArchetypeInformer,
		ArchetypePossessed,
	}

	for _, archetype := range archetypes {
		t.Run(string(archetype), func(t *testing.T) {
			template := GetArchetypeTemplate(archetype)

			if len(template.CoreTraits) == 0 {
				t.Errorf("%s archetype has no core traits", archetype)
			}
			if len(template.SkillSuggestions) == 0 {
				t.Errorf("%s archetype has no skill suggestions", archetype)
			}
		})
	}
}
