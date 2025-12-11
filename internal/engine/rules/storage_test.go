package rules

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
)

func TestRuleSetToSaveStorage(t *testing.T) {
	// Test nil ruleset
	var nilRS *RuleSet
	if nilRS.ToSaveStorage() != nil {
		t.Error("nil RuleSet.ToSaveStorage() should return nil")
	}

	// Test empty ruleset
	emptyRS := NewRuleSet()
	if emptyRS.ToSaveStorage() != nil {
		t.Error("empty RuleSet.ToSaveStorage() should return nil")
	}

	// Test with rules
	rs := NewRuleSet()
	rule := NewRule("test-rule-1", RuleTypeScenario)
	rule.Trigger = Condition{Type: "location", Value: "basement"}
	rule.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 20, SANDamage: 10}
	rule.Clues = []string{"clue1", "clue2"}
	rule.Violations = 1
	rule.MaxViolations = 2
	rule.Discovered = true
	rs.Add(rule)

	storage := rs.ToSaveStorage()
	if storage == nil {
		t.Fatal("ToSaveStorage() returned nil for non-empty ruleset")
	}

	if len(storage.Rules) != 1 {
		t.Errorf("Storage has %d rules, want 1", len(storage.Rules))
	}

	sr := storage.Rules[0]
	if sr.ID != "test-rule-1" {
		t.Errorf("SavedRule.ID = %s, want test-rule-1", sr.ID)
	}
	if sr.Type != "Scenario" {
		t.Errorf("SavedRule.Type = %s, want Scenario", sr.Type)
	}
	if sr.TriggerType != "location" {
		t.Errorf("SavedRule.TriggerType = %s, want location", sr.TriggerType)
	}
	if sr.TriggerValue != "basement" {
		t.Errorf("SavedRule.TriggerValue = %s, want basement", sr.TriggerValue)
	}
	if sr.Consequence != "damage" {
		t.Errorf("SavedRule.Consequence = %s, want damage", sr.Consequence)
	}
	if sr.HPDamage != 20 {
		t.Errorf("SavedRule.HPDamage = %d, want 20", sr.HPDamage)
	}
	if sr.SANDamage != 10 {
		t.Errorf("SavedRule.SANDamage = %d, want 10", sr.SANDamage)
	}
	if len(sr.Clues) != 2 {
		t.Errorf("SavedRule.Clues has %d items, want 2", len(sr.Clues))
	}
	if sr.Violations != 1 {
		t.Errorf("SavedRule.Violations = %d, want 1", sr.Violations)
	}
	if sr.MaxViolations != 2 {
		t.Errorf("SavedRule.MaxViolations = %d, want 2", sr.MaxViolations)
	}
	if !sr.Discovered {
		t.Error("SavedRule.Discovered should be true")
	}
}

func TestFromSaveStorage(t *testing.T) {
	// Test nil storage
	rs := FromSaveStorage(nil)
	if rs == nil {
		t.Fatal("FromSaveStorage(nil) should return empty RuleSet, not nil")
	}
	if rs.Count() != 0 {
		t.Errorf("FromSaveStorage(nil) count = %d, want 0", rs.Count())
	}

	// Test empty storage
	emptyStorage := &save.RuleStorage{Rules: []save.SavedRule{}}
	rs = FromSaveStorage(emptyStorage)
	if rs.Count() != 0 {
		t.Errorf("FromSaveStorage(empty) count = %d, want 0", rs.Count())
	}

	// Test with saved rules
	storage := &save.RuleStorage{
		Rules: []save.SavedRule{
			{
				ID:            "restored-rule-1",
				Type:          "Time",
				TriggerType:   "time",
				TriggerValue:  "midnight",
				Consequence:   "instant_death",
				HPDamage:      0,
				SANDamage:     0,
				Clues:         []string{"midnight clue"},
				Priority:      10,
				Violations:    0,
				MaxViolations: 0,
				Discovered:    false,
				Active:        true,
			},
		},
	}

	rs = FromSaveStorage(storage)
	if rs.Count() != 1 {
		t.Fatalf("FromSaveStorage count = %d, want 1", rs.Count())
	}

	rule := rs.GetByID("restored-rule-1")
	if rule == nil {
		t.Fatal("Could not find restored rule by ID")
	}

	if rule.Type != RuleTypeTime {
		t.Errorf("Rule.Type = %v, want RuleTypeTime", rule.Type)
	}
	if rule.Trigger.Type != "time" {
		t.Errorf("Rule.Trigger.Type = %s, want time", rule.Trigger.Type)
	}
	if rule.Trigger.Value != "midnight" {
		t.Errorf("Rule.Trigger.Value = %s, want midnight", rule.Trigger.Value)
	}
	if rule.Consequence.Type != ConsequenceInstantDeath {
		t.Errorf("Rule.Consequence.Type = %v, want ConsequenceInstantDeath", rule.Consequence.Type)
	}
	if rule.Priority != 10 {
		t.Errorf("Rule.Priority = %d, want 10", rule.Priority)
	}
	if len(rule.Clues) != 1 {
		t.Errorf("Rule.Clues has %d items, want 1", len(rule.Clues))
	}
	if !rule.Active {
		t.Error("Rule.Active should be true")
	}
}

func TestStorageRoundtrip(t *testing.T) {
	// Create original ruleset
	original := NewRuleSet()

	rule1 := NewRule("rule-1", RuleTypeScenario)
	rule1.Trigger = Condition{Type: "location", Value: "basement", Operator: "equals"}
	rule1.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 25, SANDamage: 15}
	rule1.Clues = []string{"線索一", "線索二"}
	rule1.Priority = 7
	rule1.MaxViolations = 2
	rule1.Violations = 1
	rule1.Discovered = true
	original.Add(rule1)

	rule2 := NewRule("rule-2", RuleTypeBehavior)
	rule2.Trigger = Condition{Type: "action", Value: "run"}
	rule2.Consequence = Outcome{Type: ConsequenceWarning}
	rule2.Clues = []string{"行為線索"}
	rule2.Priority = 5
	rule2.MaxViolations = 1
	original.Add(rule2)

	rule3 := NewRule("rule-3", RuleTypeTime)
	rule3.Trigger = Condition{Type: "time", Value: "3am"}
	rule3.Consequence = Outcome{Type: ConsequenceInstantDeath}
	rule3.Clues = []string{"時間線索"}
	rule3.Priority = 10
	rule3.MaxViolations = 0
	rule3.Active = false
	original.Add(rule3)

	// Convert to storage
	storage := original.ToSaveStorage()
	if storage == nil {
		t.Fatal("ToSaveStorage returned nil")
	}
	if len(storage.Rules) != 3 {
		t.Errorf("Storage has %d rules, want 3", len(storage.Rules))
	}

	// Restore from storage
	restored := FromSaveStorage(storage)
	if restored.Count() != 3 {
		t.Errorf("Restored has %d rules, want 3", restored.Count())
	}

	// Verify rule 1
	r1 := restored.GetByID("rule-1")
	if r1 == nil {
		t.Fatal("Rule 1 not found")
	}
	if r1.Type != RuleTypeScenario {
		t.Errorf("Rule 1 type = %v, want Scenario", r1.Type)
	}
	if r1.Trigger.Value != "basement" {
		t.Errorf("Rule 1 trigger value = %s, want basement", r1.Trigger.Value)
	}
	if r1.Consequence.Type != ConsequenceDamage {
		t.Error("Rule 1 consequence type wrong")
	}
	if r1.Consequence.HPDamage != 25 {
		t.Errorf("Rule 1 HP damage = %d, want 25", r1.Consequence.HPDamage)
	}
	if r1.Violations != 1 {
		t.Errorf("Rule 1 violations = %d, want 1", r1.Violations)
	}
	if !r1.Discovered {
		t.Error("Rule 1 should be discovered")
	}

	// Verify rule 3
	r3 := restored.GetByID("rule-3")
	if r3 == nil {
		t.Fatal("Rule 3 not found")
	}
	if r3.Consequence.Type != ConsequenceInstantDeath {
		t.Error("Rule 3 consequence type wrong")
	}
	if r3.Active {
		t.Error("Rule 3 should be inactive")
	}
}

func TestAllConsequenceTypesRoundtrip(t *testing.T) {
	tests := []ConsequenceType{
		ConsequenceWarning,
		ConsequenceDamage,
		ConsequenceInstantDeath,
	}

	for _, ct := range tests {
		rs := NewRuleSet()
		rule := NewRule("test", RuleTypeBehavior)
		rule.Consequence = Outcome{Type: ct, HPDamage: 10, SANDamage: 5}
		rule.Clues = []string{"test"}
		rs.Add(rule)

		storage := rs.ToSaveStorage()
		restored := FromSaveStorage(storage)

		restoredRule := restored.GetByID("test")
		if restoredRule.Consequence.Type != ct {
			t.Errorf("Consequence type %v did not roundtrip correctly, got %v", ct, restoredRule.Consequence.Type)
		}
	}
}

func TestAllRuleTypesRoundtrip(t *testing.T) {
	for _, rt := range AllRuleTypes() {
		rs := NewRuleSet()
		rule := NewRule("test", rt)
		rule.Clues = []string{"test"}
		rs.Add(rule)

		storage := rs.ToSaveStorage()
		restored := FromSaveStorage(storage)

		restoredRule := restored.GetByID("test")
		if restoredRule.Type != rt {
			t.Errorf("Rule type %v did not roundtrip correctly, got %v", rt, restoredRule.Type)
		}
	}
}
