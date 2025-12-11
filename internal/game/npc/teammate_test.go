package npc

import (
	"testing"
)

func TestNPCArchetypes(t *testing.T) {
	archetypes := []NPCArchetype{
		ArchetypeVictim,
		ArchetypeUnreliable,
		ArchetypeLogic,
		ArchetypeIntuition,
		ArchetypeInformer,
		ArchetypePossessed,
	}

	if len(archetypes) != 6 {
		t.Errorf("Expected 6 archetypes, got %d", len(archetypes))
	}
}

func TestTeammateStructure(t *testing.T) {
	teammate := Teammate{
		ID:         "tm-001",
		Name:       "Test Character",
		Archetype:  ArchetypeLogic,
		Background: "A test background",
		Skills:     []string{"Analysis", "Memory"},
		HP:         100,
		Location:   "Start",
	}

	if teammate.ID != "tm-001" {
		t.Errorf("Expected ID 'tm-001', got '%s'", teammate.ID)
	}

	if teammate.Name != "Test Character" {
		t.Errorf("Expected Name 'Test Character', got '%s'", teammate.Name)
	}

	if teammate.Archetype != ArchetypeLogic {
		t.Errorf("Expected Archetype 'logic', got '%s'", teammate.Archetype)
	}
}

func TestTeammateStatusDefaults(t *testing.T) {
	teammate := Teammate{
		ID:   "tm-002",
		Name: "Default Test",
	}

	// Default values should be set
	if teammate.HP < 0 {
		t.Error("HP should not be negative by default")
	}
}

// TestNewTeammate tests the NewTeammate constructor
func TestNewTeammate(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		tmName    string
		archetype NPCArchetype
	}{
		{
			name:      "Create victim archetype",
			id:        "tm-001",
			tmName:    "John Doe",
			archetype: ArchetypeVictim,
		},
		{
			name:      "Create logic archetype",
			id:        "tm-002",
			tmName:    "Jane Smith",
			archetype: ArchetypeLogic,
		},
		{
			name:      "Empty name",
			id:        "tm-003",
			tmName:    "",
			archetype: ArchetypeIntuition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTeammate(tt.id, tt.tmName, tt.archetype)

			// Test basic fields
			if tm.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, tm.ID)
			}
			if tm.Name != tt.tmName {
				t.Errorf("Expected Name %s, got %s", tt.tmName, tm.Name)
			}
			if tm.Archetype != tt.archetype {
				t.Errorf("Expected Archetype %s, got %s", tt.archetype, tm.Archetype)
			}

			// Test HP defaults
			if tm.HP != 100 {
				t.Errorf("Expected HP 100, got %d", tm.HP)
			}

			// Test Status defaults
			if !tm.Status.Alive {
				t.Error("Expected Alive to be true")
			}
			if !tm.Status.Conscious {
				t.Error("Expected Conscious to be true")
			}
			if tm.Status.Condition != "healthy" {
				t.Errorf("Expected Condition 'healthy', got %s", tm.Status.Condition)
			}
			if tm.Status.Relationship != 50 {
				t.Errorf("Expected Relationship 50, got %d", tm.Status.Relationship)
			}

			// Test Location initialization
			if tm.Location != "" {
				t.Errorf("Expected Location to be empty string, got %s", tm.Location)
			}

			// Test slice initialization (should not be nil)
			if tm.Inventory == nil {
				t.Error("Expected Inventory to be initialized, got nil")
			}
			if len(tm.Inventory) != 0 {
				t.Errorf("Expected empty Inventory, got length %d", len(tm.Inventory))
			}
			if tm.Skills == nil {
				t.Error("Expected Skills to be initialized, got nil")
			}

			// Test Status Tracking defaults (Story 4.4)
			if tm.LastSeen.IsZero() {
				t.Error("Expected LastSeen to be initialized, got zero time")
			}
			if tm.EmotionalState != EmotionCalm {
				t.Errorf("Expected EmotionalState to be calm, got %s", tm.EmotionalState)
			}
			if tm.InjuryLevel != InjuryNone {
				t.Errorf("Expected InjuryLevel to be 0 (none), got %d", tm.InjuryLevel)
			}
			if tm.IsSeparated {
				t.Error("Expected IsSeparated to be false")
			}
			if tm.LastMessage != nil {
				t.Error("Expected LastMessage to be nil")
			}
		})
	}
}
