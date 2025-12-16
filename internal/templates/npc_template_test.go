package templates

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// Story 4.4 AC1: NPCTemplate structure has all required fields (three-state structure)
func TestNPCTemplate_Structure(t *testing.T) {
	npc := &NPCTemplate{
		ID:               "test_001",
		Name:             "Test NPC",
		Archetype:        NPCArchetypeSurvivor,
		FunctionalRole:   "Test role",
		NormalState: NPCState{
			Description:       "Normal state description",
			PersonalityTraits: []string{"brave", "cautious", "intelligent"},
			DialogueStyle:     "Direct and informative",
			BehaviorPatterns:  []string{"observe", "plan"},
		},
		AnxiousState: NPCState{
			Description:       "Anxious state description",
			PersonalityTraits: []string{"nervous", "suspicious"},
			DialogueStyle:     "Quick and tense",
			BehaviorPatterns:  []string{"flee", "panic"},
		},
		CorruptedState: NPCState{
			Description:       "Corrupted state description",
			PersonalityTraits: []string{"hostile", "irrational"},
			DialogueStyle:     "Incoherent and aggressive",
			BehaviorPatterns:  []string{"attack", "self-destruct"},
		},
		SpecialAbilities: []string{"lockpicking", "first aid"},
		KnowledgeLevel:   KnowledgeLevelPartial,
		TrustLevel:       TrustLevelTrustworthy,
		BackgroundHints:  []string{"Former security guard", "Survived for 3 days"},
		Tags:             []string{"test", "sample"},
		Description:      "A test NPC archetype",
	}

	// Verify all fields are accessible
	if npc.ID != "test_001" {
		t.Error("ID field not working")
	}
	if npc.Name != "Test NPC" {
		t.Error("Name field not working")
	}
	if npc.Archetype != NPCArchetypeSurvivor {
		t.Error("Archetype field not working")
	}
	if len(npc.NormalState.PersonalityTraits) != 3 {
		t.Error("NormalState PersonalityTraits field not working")
	}
	if npc.NormalState.DialogueStyle != "Direct and informative" {
		t.Error("NormalState DialogueStyle field not working")
	}
	if len(npc.AnxiousState.PersonalityTraits) != 2 {
		t.Error("AnxiousState PersonalityTraits field not working")
	}
	if len(npc.CorruptedState.PersonalityTraits) != 2 {
		t.Error("CorruptedState PersonalityTraits field not working")
	}
	if len(npc.SpecialAbilities) != 2 {
		t.Error("SpecialAbilities field not working")
	}
	if npc.KnowledgeLevel != KnowledgeLevelPartial {
		t.Error("KnowledgeLevel field not working")
	}
	if npc.TrustLevel != TrustLevelTrustworthy {
		t.Error("TrustLevel field not working")
	}
	if len(npc.BackgroundHints) != 2 {
		t.Error("BackgroundHints field not working")
	}
	if len(npc.Tags) != 2 {
		t.Error("Tags field not working")
	}
	if npc.Description != "A test NPC archetype" {
		t.Error("Description field not working")
	}
}

// Story 4.4 AC2: Load archetypes.yaml with at least 6 NPC archetypes (three-state structure)
func TestNPCTemplate_LoadArchetypesYAML(t *testing.T) {
	// Find the project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "templates")
	archetypesPath := filepath.Join(templatesDir, "npcs", "archetypes.yaml")

	// Read the file
	data, err := os.ReadFile(archetypesPath)
	if err != nil {
		t.Fatalf("Failed to read archetypes.yaml: %v", err)
	}

	// Parse YAML
	var collection NPCTemplateCollection
	err = yaml.Unmarshal(data, &collection)
	if err != nil {
		t.Fatalf("Failed to parse archetypes.yaml: %v", err)
	}

	// AC: Should contain at least 6 NPC archetypes
	if len(collection.NPCTypes) < 6 {
		t.Errorf("Expected at least 6 NPC archetypes in archetypes.yaml, got %d", len(collection.NPCTypes))
	}

	// AC: Each NPC should have all three states with PersonalityTraits and DialogueStyle
	for i, npc := range collection.NPCTypes {
		// Check Normal State
		if len(npc.NormalState.PersonalityTraits) == 0 {
			t.Errorf("NPC %d (%s): NormalState PersonalityTraits list is empty", i, npc.ID)
		}
		if npc.NormalState.DialogueStyle == "" {
			t.Errorf("NPC %d (%s): NormalState DialogueStyle is empty", i, npc.ID)
		}

		// Check Anxious State
		if len(npc.AnxiousState.PersonalityTraits) == 0 {
			t.Errorf("NPC %d (%s): AnxiousState PersonalityTraits list is empty", i, npc.ID)
		}
		if npc.AnxiousState.DialogueStyle == "" {
			t.Errorf("NPC %d (%s): AnxiousState DialogueStyle is empty", i, npc.ID)
		}

		// Check Corrupted State
		if len(npc.CorruptedState.PersonalityTraits) == 0 {
			t.Errorf("NPC %d (%s): CorruptedState PersonalityTraits list is empty", i, npc.ID)
		}
		if npc.CorruptedState.DialogueStyle == "" {
			t.Errorf("NPC %d (%s): CorruptedState DialogueStyle is empty", i, npc.ID)
		}

		// Validate the NPC
		err := npc.Validate()
		if err != nil {
			t.Errorf("NPC %d (%s) validation failed: %v", i, npc.ID, err)
		}
	}
}

// Test Validate method (three-state structure)
func TestNPCTemplate_Validate(t *testing.T) {
	// Valid NPC with all three states
	validNPC := &NPCTemplate{
		ID:   "test_001",
		Name: "Test",
		NormalState: NPCState{
			Description:       "Normal",
			PersonalityTraits: []string{"brave"},
			DialogueStyle:     "calm",
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err := validNPC.Validate()
	if err != nil {
		t.Errorf("Valid NPC should pass validation, got error: %v", err)
	}

	// Missing ID
	invalidNPC1 := &NPCTemplate{
		Name: "Test",
		NormalState: NPCState{
			Description:       "Normal",
			PersonalityTraits: []string{"brave"},
			DialogueStyle:     "calm",
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err = invalidNPC1.Validate()
	if err == nil {
		t.Error("NPC without ID should fail validation")
	}

	// Missing Name
	invalidNPC2 := &NPCTemplate{
		ID: "test_001",
		NormalState: NPCState{
			Description:       "Normal",
			PersonalityTraits: []string{"brave"},
			DialogueStyle:     "calm",
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err = invalidNPC2.Validate()
	if err == nil {
		t.Error("NPC without Name should fail validation")
	}

	// Missing NormalState Description
	invalidNPC3 := &NPCTemplate{
		ID:   "test_001",
		Name: "Test",
		NormalState: NPCState{
			PersonalityTraits: []string{"brave"},
			DialogueStyle:     "calm",
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err = invalidNPC3.Validate()
	if err == nil {
		t.Error("NPC without NormalState Description should fail validation")
	}

	// Empty NormalState PersonalityTraits
	invalidNPC4 := &NPCTemplate{
		ID:   "test_001",
		Name: "Test",
		NormalState: NPCState{
			Description:       "Normal",
			PersonalityTraits: []string{},
			DialogueStyle:     "calm",
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err = invalidNPC4.Validate()
	if err == nil {
		t.Error("NPC without NormalState PersonalityTraits should fail validation")
	}

	// Missing NormalState DialogueStyle
	invalidNPC5 := &NPCTemplate{
		ID:   "test_001",
		Name: "Test",
		NormalState: NPCState{
			Description:       "Normal",
			PersonalityTraits: []string{"brave"},
		},
		AnxiousState: NPCState{
			Description:       "Anxious",
			PersonalityTraits: []string{"nervous"},
			DialogueStyle:     "tense",
		},
		CorruptedState: NPCState{
			Description:       "Corrupted",
			PersonalityTraits: []string{"hostile"},
			DialogueStyle:     "aggressive",
		},
	}

	err = invalidNPC5.Validate()
	if err == nil {
		t.Error("NPC without NormalState DialogueStyle should fail validation")
	}
}

// Test HasTag
func TestNPCTemplate_HasTag(t *testing.T) {
	npc := &NPCTemplate{
		Tags: []string{"survivor", "helpful", "knowledgeable"},
	}

	if !npc.HasTag("survivor") {
		t.Error("Should have 'survivor' tag")
	}

	if !npc.HasTag("helpful") {
		t.Error("Should have 'helpful' tag")
	}

	if npc.HasTag("nonexistent") {
		t.Error("Should not have 'nonexistent' tag")
	}
}

// Test HasPersonalityTrait (now checks across all states)
func TestNPCTemplate_HasPersonalityTrait(t *testing.T) {
	npc := &NPCTemplate{
		NormalState: NPCState{
			PersonalityTraits: []string{"brave", "cautious"},
		},
		AnxiousState: NPCState{
			PersonalityTraits: []string{"nervous"},
		},
		CorruptedState: NPCState{
			PersonalityTraits: []string{"hostile"},
		},
	}

	// Should find traits from Normal state
	if !npc.HasPersonalityTrait("brave") {
		t.Error("Should have 'brave' trait")
	}

	if !npc.HasPersonalityTrait("cautious") {
		t.Error("Should have 'cautious' trait")
	}

	// Should find traits from Anxious state
	if !npc.HasPersonalityTrait("nervous") {
		t.Error("Should have 'nervous' trait")
	}

	// Should find traits from Corrupted state
	if !npc.HasPersonalityTrait("hostile") {
		t.Error("Should have 'hostile' trait")
	}

	// Should not find non-existent trait
	if npc.HasPersonalityTrait("cowardly") {
		t.Error("Should not have 'cowardly' trait")
	}
}

// Test HasAbility
func TestNPCTemplate_HasAbility(t *testing.T) {
	npc := &NPCTemplate{
		SpecialAbilities: []string{"lockpicking", "first aid", "electronics"},
	}

	if !npc.HasAbility("lockpicking") {
		t.Error("Should have 'lockpicking' ability")
	}

	if !npc.HasAbility("first aid") {
		t.Error("Should have 'first aid' ability")
	}

	if npc.HasAbility("flying") {
		t.Error("Should not have 'flying' ability")
	}
}

// Test GetPersonalityTraitsString (now gets traits from all states)
func TestNPCTemplate_GetPersonalityTraitsString(t *testing.T) {
	npc := &NPCTemplate{
		NormalState: NPCState{
			PersonalityTraits: []string{"brave", "cautious"},
		},
		AnxiousState: NPCState{
			PersonalityTraits: []string{"nervous"},
		},
		CorruptedState: NPCState{
			PersonalityTraits: []string{"hostile"},
		},
	}

	traits := npc.GetPersonalityTraitsString()
	expected := "brave, cautious, nervous, hostile"

	if traits != expected {
		t.Errorf("Expected '%s', got '%s'", expected, traits)
	}

	// Test with single trait per state
	npc2 := &NPCTemplate{
		NormalState: NPCState{
			PersonalityTraits: []string{"brave"},
		},
		AnxiousState: NPCState{
			PersonalityTraits: []string{},
		},
		CorruptedState: NPCState{
			PersonalityTraits: []string{},
		},
	}

	traits2 := npc2.GetPersonalityTraitsString()
	if traits2 != "brave" {
		t.Errorf("Expected 'brave', got '%s'", traits2)
	}

	// Test with all empty lists
	npc3 := &NPCTemplate{
		NormalState: NPCState{
			PersonalityTraits: []string{},
		},
		AnxiousState: NPCState{
			PersonalityTraits: []string{},
		},
		CorruptedState: NPCState{
			PersonalityTraits: []string{},
		},
	}

	traits3 := npc3.GetPersonalityTraitsString()
	if traits3 != "" {
		t.Errorf("Expected empty string, got '%s'", traits3)
	}
}

// Test GetState
func TestNPCTemplate_GetState(t *testing.T) {
	npc := &NPCTemplate{
		NormalState: NPCState{
			Description: "Normal state",
		},
		AnxiousState: NPCState{
			Description: "Anxious state",
		},
		CorruptedState: NPCState{
			Description: "Corrupted state",
		},
	}

	// Test getting Normal state (1)
	state1 := npc.GetState(1)
	if state1 == nil {
		t.Error("GetState(1) should return Normal state")
	} else if state1.Description != "Normal state" {
		t.Error("GetState(1) returned wrong state")
	}

	// Test getting Anxious state (2)
	state2 := npc.GetState(2)
	if state2 == nil {
		t.Error("GetState(2) should return Anxious state")
	} else if state2.Description != "Anxious state" {
		t.Error("GetState(2) returned wrong state")
	}

	// Test getting Corrupted state (3)
	state3 := npc.GetState(3)
	if state3 == nil {
		t.Error("GetState(3) should return Corrupted state")
	} else if state3.Description != "Corrupted state" {
		t.Error("GetState(3) returned wrong state")
	}

	// Test invalid state number
	state0 := npc.GetState(0)
	if state0 != nil {
		t.Error("GetState(0) should return nil")
	}

	state4 := npc.GetState(4)
	if state4 != nil {
		t.Error("GetState(4) should return nil")
	}
}

// Test IsKnowledgeable
func TestNPCTemplate_IsKnowledgeable(t *testing.T) {
	testCases := []struct {
		level    KnowledgeLevel
		expected bool
	}{
		{KnowledgeLevelNone, false},
		{KnowledgeLevelPartial, true},
		{KnowledgeLevelFull, true},
	}

	for _, tc := range testCases {
		npc := &NPCTemplate{KnowledgeLevel: tc.level}
		result := npc.IsKnowledgeable()
		if result != tc.expected {
			t.Errorf("Knowledge level %s: expected %v, got %v",
				tc.level, tc.expected, result)
		}
	}
}

// Test IsTrustworthy
func TestNPCTemplate_IsTrustworthy(t *testing.T) {
	testCases := []struct {
		level    TrustLevel
		expected bool
	}{
		{TrustLevelUntrustworthy, false},
		{TrustLevelNeutral, false},
		{TrustLevelTrustworthy, true},
	}

	for _, tc := range testCases {
		npc := &NPCTemplate{TrustLevel: tc.level}
		result := npc.IsTrustworthy()
		if result != tc.expected {
			t.Errorf("Trust level %s: expected %v, got %v",
				tc.level, tc.expected, result)
		}
	}
}

// Test GetAbilityCount
func TestNPCTemplate_GetAbilityCount(t *testing.T) {
	npc := &NPCTemplate{
		SpecialAbilities: []string{"lockpicking", "first aid", "electronics"},
	}

	count := npc.GetAbilityCount()
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test with empty list
	npc2 := &NPCTemplate{
		SpecialAbilities: []string{},
	}

	count2 := npc2.GetAbilityCount()
	if count2 != 0 {
		t.Errorf("Expected count 0, got %d", count2)
	}
}

// Test NPCArchetype constants
func TestNPCArchetype_Constants(t *testing.T) {
	if NPCArchetypeSurvivor != "survivor" {
		t.Error("NPCArchetypeSurvivor constant incorrect")
	}
	if NPCArchetypeThreat != "threat" {
		t.Error("NPCArchetypeThreat constant incorrect")
	}
	if NPCArchetypeNeutral != "neutral" {
		t.Error("NPCArchetypeNeutral constant incorrect")
	}
	if NPCArchetypeHelper != "helper" {
		t.Error("NPCArchetypeHelper constant incorrect")
	}
	if NPCArchetypeVictim != "victim" {
		t.Error("NPCArchetypeVictim constant incorrect")
	}
	if NPCArchetypeBetrayer != "betrayer" {
		t.Error("NPCArchetypeBetrayer constant incorrect")
	}
}

// Test KnowledgeLevel constants
func TestKnowledgeLevel_Constants(t *testing.T) {
	if KnowledgeLevelNone != "none" {
		t.Error("KnowledgeLevelNone constant incorrect")
	}
	if KnowledgeLevelPartial != "partial" {
		t.Error("KnowledgeLevelPartial constant incorrect")
	}
	if KnowledgeLevelFull != "full" {
		t.Error("KnowledgeLevelFull constant incorrect")
	}
}

// Test TrustLevel constants
func TestTrustLevel_Constants(t *testing.T) {
	if TrustLevelUntrustworthy != "untrustworthy" {
		t.Error("TrustLevelUntrustworthy constant incorrect")
	}
	if TrustLevelNeutral != "neutral" {
		t.Error("TrustLevelNeutral constant incorrect")
	}
	if TrustLevelTrustworthy != "trustworthy" {
		t.Error("TrustLevelTrustworthy constant incorrect")
	}
}

// Test YAML marshaling/unmarshaling (three-state structure)
func TestNPCTemplate_YAMLSerialization(t *testing.T) {
	original := &NPCTemplate{
		ID:             "test_yaml_001",
		Name:           "YAML Test NPC",
		Archetype:      NPCArchetypeHelper,
		FunctionalRole: "Test helper",
		NormalState: NPCState{
			Description:       "Friendly and helpful",
			PersonalityTraits: []string{"friendly", "knowledgeable", "mysterious"},
			DialogueStyle:     "Cryptic and philosophical",
			BehaviorPatterns:  []string{"observe", "guide"},
		},
		AnxiousState: NPCState{
			Description:       "Worried and cautious",
			PersonalityTraits: []string{"worried", "secretive"},
			DialogueStyle:     "Hesitant and cryptic",
			BehaviorPatterns:  []string{"hide", "warn"},
		},
		CorruptedState: NPCState{
			Description:       "Lost and corrupted",
			PersonalityTraits: []string{"desperate", "dangerous"},
			DialogueStyle:     "Incoherent and pleading",
			BehaviorPatterns:  []string{"attack", "flee"},
		},
		SpecialAbilities: []string{"healing", "knowledge sharing"},
		KnowledgeLevel:   KnowledgeLevelFull,
		TrustLevel:       TrustLevelNeutral,
		BackgroundHints:  []string{"Has been here before", "Knows the rules"},
		Tags:             []string{"test", "yaml", "helper"},
		Description:      "A test NPC for YAML serialization",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored NPCTemplate
	err = yaml.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if restored.ID != original.ID {
		t.Error("ID mismatch after serialization")
	}
	if restored.Name != original.Name {
		t.Error("Name mismatch after serialization")
	}
	if restored.Archetype != original.Archetype {
		t.Error("Archetype mismatch after serialization")
	}
	if len(restored.NormalState.PersonalityTraits) != len(original.NormalState.PersonalityTraits) {
		t.Error("NormalState PersonalityTraits length mismatch after serialization")
	}
	if restored.NormalState.DialogueStyle != original.NormalState.DialogueStyle {
		t.Error("NormalState DialogueStyle mismatch after serialization")
	}
	if restored.KnowledgeLevel != original.KnowledgeLevel {
		t.Error("KnowledgeLevel mismatch after serialization")
	}
	if restored.TrustLevel != original.TrustLevel {
		t.Error("TrustLevel mismatch after serialization")
	}
}

// Test loading NPCTemplateCollection (three-state structure)
func TestNPCTemplateCollection_Load(t *testing.T) {
	yamlContent := `
version: "1.0"
npc_types:
  - id: test_001
    name: The Survivor
    archetype: survivor
    functional_role: Team leader
    normal_state:
      description: Calm and resourceful
      personality_traits:
        - brave
        - resourceful
      dialogue_style: Direct and practical
      behavior_patterns:
        - lead
        - plan
    anxious_state:
      description: Tense but controlled
      personality_traits:
        - tense
        - focused
      dialogue_style: Quick and efficient
      behavior_patterns:
        - defend
        - retreat
    corrupted_state:
      description: Paranoid and aggressive
      personality_traits:
        - paranoid
        - aggressive
      dialogue_style: Harsh and accusatory
      behavior_patterns:
        - attack
        - isolate
    special_abilities:
      - lockpicking
    knowledge_level: partial
    trust_level: trustworthy
    background_hints:
      - Former engineer
    tags:
      - survivor
  - id: test_002
    name: The Betrayer
    archetype: betrayer
    functional_role: Manipulator
    normal_state:
      description: Charming and helpful
      personality_traits:
        - deceptive
        - self-serving
      dialogue_style: Smooth and persuasive
      behavior_patterns:
        - manipulate
        - observe
    anxious_state:
      description: Mask slipping
      personality_traits:
        - desperate
        - cunning
      dialogue_style: Nervous but persuasive
      behavior_patterns:
        - lie
        - escape
    corrupted_state:
      description: Fully exposed
      personality_traits:
        - ruthless
        - selfish
      dialogue_style: Mocking and threatening
      behavior_patterns:
        - betray
        - sacrifice others
    special_abilities:
      - manipulation
    knowledge_level: full
    trust_level: untrustworthy
    background_hints:
      - Has survived many loops
    tags:
      - betrayer
`

	var collection NPCTemplateCollection
	err := yaml.Unmarshal([]byte(yamlContent), &collection)
	if err != nil {
		t.Fatalf("Failed to unmarshal collection: %v", err)
	}

	if collection.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", collection.Version)
	}

	if len(collection.NPCTypes) != 2 {
		t.Errorf("Expected 2 NPC types, got %d", len(collection.NPCTypes))
	}

	// Validate each NPC
	for i, npc := range collection.NPCTypes {
		err := npc.Validate()
		if err != nil {
			t.Errorf("NPC %d validation failed: %v", i, err)
		}
	}
}
