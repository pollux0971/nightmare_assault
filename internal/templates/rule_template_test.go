package templates

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// Story 4.2 AC1: RuleTemplate structure has all required fields
func TestRuleTemplate_Structure(t *testing.T) {
	rule := &RuleTemplate{
		ID:            "test_001",
		Name:          "Test Rule",
		Category:      RuleCategorySensory,
		Difficulty:    RuleDifficultyMedium,
		TriggerMedium: "Test trigger",
		FalseClue:     "False clue",
		SurvivalRule:  "Survival rule",
		Punishment: Punishment{
			SANDamage: 10,
			HPDamage:  5,
			Effect:    "Test effect",
		},
		ClueHints: []string{"Hint 1", "Hint 2"},
		Tags:      []string{"test", "sample"},
	}

	// Verify all fields are accessible
	if rule.ID != "test_001" {
		t.Error("ID field not working")
	}
	if rule.Name != "Test Rule" {
		t.Error("Name field not working")
	}
	if rule.Category != RuleCategorySensory {
		t.Error("Category field not working")
	}
	if rule.Difficulty != RuleDifficultyMedium {
		t.Error("Difficulty field not working")
	}
	if rule.TriggerMedium != "Test trigger" {
		t.Error("TriggerMedium field not working")
	}
	if rule.FalseClue != "False clue" {
		t.Error("FalseClue field not working")
	}
	if rule.SurvivalRule != "Survival rule" {
		t.Error("SurvivalRule field not working")
	}
	if rule.Punishment.SANDamage != 10 {
		t.Error("Punishment.SANDamage field not working")
	}
	if rule.Punishment.HPDamage != 5 {
		t.Error("Punishment.HPDamage field not working")
	}
	if rule.Punishment.Effect != "Test effect" {
		t.Error("Punishment.Effect field not working")
	}
	if len(rule.ClueHints) != 2 {
		t.Error("ClueHints field not working")
	}
	if len(rule.Tags) != 2 {
		t.Error("Tags field not working")
	}
}

// Story 4.2 AC2: Load sensory.yaml with at least 3 rules
func TestRuleTemplate_LoadSensoryYAML(t *testing.T) {
	// Find the project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "templates")
	sensoryPath := filepath.Join(templatesDir, "rules", "sensory.yaml")

	// Read the file
	data, err := os.ReadFile(sensoryPath)
	if err != nil {
		t.Fatalf("Failed to read sensory.yaml: %v", err)
	}

	// Parse YAML
	var collection RuleTemplateCollection
	err = yaml.Unmarshal(data, &collection)
	if err != nil {
		t.Fatalf("Failed to parse sensory.yaml: %v", err)
	}

	// AC: Should contain at least 3 rules
	if len(collection.Rules) < 3 {
		t.Errorf("Expected at least 3 rules in sensory.yaml, got %d", len(collection.Rules))
	}

	// AC: Category should be sensory
	if collection.Category != RuleCategorySensory {
		t.Errorf("Expected category 'sensory', got '%s'", collection.Category)
	}

	// AC: Each rule should have TriggerMedium, FalseClue, SurvivalRule
	for i, rule := range collection.Rules {
		if rule.TriggerMedium == "" {
			t.Errorf("Rule %d: TriggerMedium is empty", i)
		}
		if rule.FalseClue == "" {
			t.Errorf("Rule %d: FalseClue is empty", i)
		}
		if rule.SurvivalRule == "" {
			t.Errorf("Rule %d: SurvivalRule is empty", i)
		}

		// Validate the rule
		err := rule.Validate()
		if err != nil {
			t.Errorf("Rule %d (%s) validation failed: %v", i, rule.ID, err)
		}
	}
}

// Test Validate method
func TestRuleTemplate_Validate(t *testing.T) {
	// Valid rule
	validRule := &RuleTemplate{
		ID:            "test_001",
		Name:          "Test",
		TriggerMedium: "trigger",
		SurvivalRule:  "rule",
	}

	err := validRule.Validate()
	if err != nil {
		t.Errorf("Valid rule should pass validation, got error: %v", err)
	}

	// Missing ID
	invalidRule1 := &RuleTemplate{
		Name:          "Test",
		TriggerMedium: "trigger",
		SurvivalRule:  "rule",
	}

	err = invalidRule1.Validate()
	if err == nil {
		t.Error("Rule without ID should fail validation")
	}

	// Missing Name
	invalidRule2 := &RuleTemplate{
		ID:            "test_001",
		TriggerMedium: "trigger",
		SurvivalRule:  "rule",
	}

	err = invalidRule2.Validate()
	if err == nil {
		t.Error("Rule without Name should fail validation")
	}

	// Missing TriggerMedium
	invalidRule3 := &RuleTemplate{
		ID:           "test_001",
		Name:         "Test",
		SurvivalRule: "rule",
	}

	err = invalidRule3.Validate()
	if err == nil {
		t.Error("Rule without TriggerMedium should fail validation")
	}

	// Missing SurvivalRule
	invalidRule4 := &RuleTemplate{
		ID:            "test_001",
		Name:          "Test",
		TriggerMedium: "trigger",
	}

	err = invalidRule4.Validate()
	if err == nil {
		t.Error("Rule without SurvivalRule should fail validation")
	}
}

// Test GetDifficultyLevel
func TestRuleTemplate_GetDifficultyLevel(t *testing.T) {
	testCases := []struct {
		difficulty RuleDifficulty
		expected   int
	}{
		{RuleDifficultyEasy, 1},
		{RuleDifficultyMedium, 2},
		{RuleDifficultyHard, 3},
		{RuleDifficultyHell, 4},
		{RuleDifficulty("unknown"), 1}, // Default to easy
	}

	for _, tc := range testCases {
		rule := &RuleTemplate{Difficulty: tc.difficulty}
		level := rule.GetDifficultyLevel()
		if level != tc.expected {
			t.Errorf("Difficulty %s: expected level %d, got %d",
				tc.difficulty, tc.expected, level)
		}
	}
}

// Test HasTag
func TestRuleTemplate_HasTag(t *testing.T) {
	rule := &RuleTemplate{
		Tags: []string{"vision", "monster", "sensory"},
	}

	if !rule.HasTag("vision") {
		t.Error("Should have 'vision' tag")
	}

	if !rule.HasTag("monster") {
		t.Error("Should have 'monster' tag")
	}

	if rule.HasTag("nonexistent") {
		t.Error("Should not have 'nonexistent' tag")
	}
}

// Test GetTotalDamage
func TestRuleTemplate_GetTotalDamage(t *testing.T) {
	rule := &RuleTemplate{
		Punishment: Punishment{
			HPDamage:  20,
			SANDamage: 30,
		},
	}

	totalDamage := rule.GetTotalDamage()
	if totalDamage != 50 {
		t.Errorf("Expected total damage 50, got %d", totalDamage)
	}
}

// Test RuleCategory constants
func TestRuleCategory_Constants(t *testing.T) {
	if RuleCategorySensory != "sensory" {
		t.Error("RuleCategorySensory constant incorrect")
	}
	if RuleCategorySpatial != "spatial" {
		t.Error("RuleCategorySpatial constant incorrect")
	}
	if RuleCategorySocial != "social" {
		t.Error("RuleCategorySocial constant incorrect")
	}
}

// Test RuleDifficulty constants
func TestRuleDifficulty_Constants(t *testing.T) {
	if RuleDifficultyEasy != "easy" {
		t.Error("RuleDifficultyEasy constant incorrect")
	}
	if RuleDifficultyMedium != "medium" {
		t.Error("RuleDifficultyMedium constant incorrect")
	}
	if RuleDifficultyHard != "hard" {
		t.Error("RuleDifficultyHard constant incorrect")
	}
	if RuleDifficultyHell != "hell" {
		t.Error("RuleDifficultyHell constant incorrect")
	}
}

// Helper function to find project root by looking for go.mod
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for i := 0; i < 10; i++ {
		// Check if go.mod exists in this directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// Test YAML marshaling/unmarshaling
func TestRuleTemplate_YAMLSerialization(t *testing.T) {
	original := &RuleTemplate{
		ID:            "test_yaml_001",
		Name:          "YAML Test Rule",
		Category:      RuleCategorySpatial,
		Difficulty:    RuleDifficultyHard,
		TriggerMedium: "Test trigger",
		FalseClue:     "Test false clue",
		SurvivalRule:  "Test survival rule",
		Punishment: Punishment{
			SANDamage: 25,
			HPDamage:  15,
			Effect:    "Test effect",
		},
		ClueHints: []string{"Hint A", "Hint B", "Hint C"},
		Tags:      []string{"test", "yaml"},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored RuleTemplate
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
	if restored.Category != original.Category {
		t.Error("Category mismatch after serialization")
	}
	if restored.Difficulty != original.Difficulty {
		t.Error("Difficulty mismatch after serialization")
	}
	if restored.Punishment.SANDamage != original.Punishment.SANDamage {
		t.Error("SANDamage mismatch after serialization")
	}
	if len(restored.ClueHints) != len(original.ClueHints) {
		t.Error("ClueHints length mismatch after serialization")
	}
}
