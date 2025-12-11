package rules

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewChecker(t *testing.T) {
	rs := NewRuleSet()
	c := NewChecker(rs, game.DifficultyHard)

	if c == nil {
		t.Fatal("NewChecker returned nil")
	}
	if c.ruleSet != rs {
		t.Error("Checker has wrong ruleset")
	}
	if c.difficulty != game.DifficultyHard {
		t.Error("Checker has wrong difficulty")
	}
}

func TestCheckerCheckEmptyRuleSet(t *testing.T) {
	c := NewChecker(nil, game.DifficultyHard)
	action := PlayerAction{Type: "move", Target: "basement"}
	ctx := GameContext{Location: "basement", HP: 100, SAN: 100}

	result := c.Check(action, ctx)

	if len(result.Triggers) != 0 {
		t.Errorf("Expected 0 triggers with nil ruleset, got %d", len(result.Triggers))
	}
}

func TestCheckerScenarioRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("basement-rule", RuleTypeScenario)
	rule.Trigger = Condition{Type: "location", Value: "basement", Operator: "equals"}
	rule.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 20, SANDamage: 10}
	rule.Clues = []string{"test"}
	rule.MaxViolations = 0 // Immediate damage
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "move", Target: "basement", Timestamp: time.Now()}
	ctx := GameContext{Location: "basement", HP: 100, SAN: 100, Chapter: 1}

	result := c.Check(action, ctx)

	if len(result.Triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(result.Triggers))
	}

	if result.Triggers[0].Rule.ID != "basement-rule" {
		t.Error("Wrong rule triggered")
	}
}

func TestCheckerTimeRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("midnight-rule", RuleTypeTime)
	rule.Trigger = Condition{Type: "time", Value: "midnight", Operator: "equals"}
	rule.Consequence = Outcome{Type: ConsequenceInstantDeath}
	rule.Clues = []string{"test"}
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "wait", Timestamp: time.Now()}

	// Should trigger at midnight
	ctxMidnight := GameContext{TimeOfDay: "midnight", HP: 100, SAN: 100}
	result := c.Check(action, ctxMidnight)
	if len(result.Triggers) != 1 {
		t.Errorf("Expected trigger at midnight, got %d", len(result.Triggers))
	}
	if !result.Triggers[0].IsFatal {
		t.Error("Midnight rule should be fatal")
	}

	// Should NOT trigger at noon
	ctxNoon := GameContext{TimeOfDay: "noon", HP: 100, SAN: 100}
	c.ClearHistory() // Clear for clean test
	rs.Rules[0].Violations = 0 // Reset violations
	result = c.Check(action, ctxNoon)
	if len(result.Triggers) != 0 {
		t.Errorf("Should not trigger at noon, got %d triggers", len(result.Triggers))
	}
}

func TestCheckerBehaviorRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("run-rule", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "run", Operator: "equals"}
	rule.Consequence = Outcome{Type: ConsequenceWarning}
	rule.WarningText = "不要跑！"
	rule.Clues = []string{"test"}
	rule.MaxViolations = 2
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyEasy)
	action := PlayerAction{Type: "run", Target: "hallway", Timestamp: time.Now()}
	ctx := GameContext{Location: "hallway", HP: 100, SAN: 100}

	result := c.Check(action, ctx)

	if len(result.Triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(result.Triggers))
	}

	if !result.Triggers[0].IsWarning {
		t.Error("First violation should be warning")
	}
	if result.Triggers[0].WarningText != "不要跑！" {
		t.Errorf("Wrong warning text: %s", result.Triggers[0].WarningText)
	}
}

func TestCheckerStatusRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("low-san-rule", RuleTypeStatus)
	rule.Trigger = Condition{Type: "san_below", Value: "san", Operator: "less_than", Threshold: 30}
	rule.Consequence = Outcome{Type: ConsequenceDamage, SANDamage: 10}
	rule.Clues = []string{"test"}
	rule.MaxViolations = 0
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "look", Timestamp: time.Now()}

	// Should trigger with low SAN
	ctxLowSAN := GameContext{HP: 100, SAN: 25}
	result := c.Check(action, ctxLowSAN)
	if len(result.Triggers) != 1 {
		t.Errorf("Expected trigger with low SAN, got %d", len(result.Triggers))
	}

	// Should NOT trigger with high SAN
	ctxHighSAN := GameContext{HP: 100, SAN: 80}
	c.ClearHistory()
	rs.Rules[0].Violations = 0
	result = c.Check(action, ctxHighSAN)
	if len(result.Triggers) != 0 {
		t.Errorf("Should not trigger with high SAN, got %d triggers", len(result.Triggers))
	}
}

func TestCheckerObjectRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("mirror-rule", RuleTypeObject)
	rule.Trigger = Condition{Type: "object", Value: "mirror", Operator: "equals"}
	rule.Consequence = Outcome{Type: ConsequenceDamage, SANDamage: 15}
	rule.Clues = []string{"test"}
	rule.MaxViolations = 1
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "interact", Target: "mirror", Timestamp: time.Now()}
	ctx := GameContext{HP: 100, SAN: 100}

	result := c.Check(action, ctx)

	if len(result.Triggers) != 1 {
		t.Fatalf("Expected 1 trigger for mirror interaction, got %d", len(result.Triggers))
	}
}

func TestCheckerRuleCheckOrder(t *testing.T) {
	rs := NewRuleSet()

	// Add rules in reverse order (Status first)
	statusRule := NewRule("status-rule", RuleTypeStatus)
	statusRule.Trigger = Condition{Type: "san_below", Threshold: 200} // Always triggers
	statusRule.Consequence = Outcome{Type: ConsequenceWarning}
	statusRule.Clues = []string{"test"}
	statusRule.Priority = 5
	rs.Add(statusRule)

	behaviorRule := NewRule("behavior-rule", RuleTypeBehavior)
	behaviorRule.Trigger = Condition{Type: "action", Value: "test"}
	behaviorRule.Consequence = Outcome{Type: ConsequenceWarning}
	behaviorRule.Clues = []string{"test"}
	behaviorRule.Priority = 5
	rs.Add(behaviorRule)

	scenarioRule := NewRule("scenario-rule", RuleTypeScenario)
	scenarioRule.Trigger = Condition{Type: "location", Value: "test-location"}
	scenarioRule.Consequence = Outcome{Type: ConsequenceWarning}
	scenarioRule.Clues = []string{"test"}
	scenarioRule.Priority = 5
	rs.Add(scenarioRule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "test", Target: "something", Timestamp: time.Now()}
	ctx := GameContext{Location: "test-location", SAN: 50}

	result := c.Check(action, ctx)

	// AC1: Should check in order Scenario → Time → Behavior → Object → Status
	// So we expect records in that order
	if len(result.Records) < 2 {
		t.Fatalf("Expected at least 2 records, got %d", len(result.Records))
	}

	// First should be scenario (since scenario rule matches)
	if result.Records[0].RuleType != RuleTypeScenario {
		t.Errorf("First record should be Scenario, got %v", result.Records[0].RuleType)
	}
}

func TestCheckerPriorityOrder(t *testing.T) {
	rs := NewRuleSet()

	lowPriority := NewRule("low-rule", RuleTypeScenario)
	lowPriority.Trigger = Condition{Type: "location", Value: "hall"}
	lowPriority.Consequence = Outcome{Type: ConsequenceWarning}
	lowPriority.Clues = []string{"test"}
	lowPriority.Priority = 1
	rs.Add(lowPriority)

	highPriority := NewRule("high-rule", RuleTypeScenario)
	highPriority.Trigger = Condition{Type: "location", Value: "hall"}
	highPriority.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 50}
	highPriority.Clues = []string{"test"}
	highPriority.Priority = 10
	rs.Add(highPriority)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "move", Timestamp: time.Now()}
	ctx := GameContext{Location: "hall"}

	result := c.Check(action, ctx)

	if len(result.Triggers) < 1 {
		t.Fatal("Expected at least 1 trigger")
	}

	// High priority should be first
	if result.Triggers[0].Rule.ID != "high-rule" {
		t.Errorf("High priority rule should be processed first, got %s", result.Triggers[0].Rule.ID)
	}
}

func TestCheckerWarningMechanism(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("warning-test", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "test"}
	rule.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 20}
	rule.WarningText = "警告！"
	rule.Clues = []string{"test"}
	rule.MaxViolations = 2 // 2 warnings before full damage
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "test", Timestamp: time.Now()}
	ctx := GameContext{}

	// First violation - should warn
	result1 := c.Check(action, ctx)
	if !result1.Triggers[0].IsWarning {
		t.Error("First violation should be warning")
	}

	// Second violation - should still warn
	result2 := c.Check(action, ctx)
	if !result2.Triggers[0].IsWarning {
		t.Error("Second violation should be warning")
	}

	// Third violation - should apply full damage
	result3 := c.Check(action, ctx)
	if result3.Triggers[0].IsWarning {
		t.Error("Third violation should not be warning")
	}
	if result3.Triggers[0].HPDamage != 20 {
		t.Errorf("Third violation HP damage = %d, want 20", result3.Triggers[0].HPDamage)
	}
}

func TestCheckerInstantDeathRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("instant-death", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "forbidden"}
	rule.Consequence = Outcome{Type: ConsequenceInstantDeath}
	rule.Clues = []string{"test"}
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "forbidden", Timestamp: time.Now()}
	ctx := GameContext{}

	result := c.Check(action, ctx)

	if !result.AnyFatal {
		t.Error("Should have fatal result")
	}
	if !result.Triggers[0].IsFatal {
		t.Error("Trigger should be fatal")
	}
}

func TestCheckerTriggerHistory(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("history-test", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "test"}
	rule.Consequence = Outcome{Type: ConsequenceWarning}
	rule.Clues = []string{"test"}
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	action := PlayerAction{Type: "test", Timestamp: time.Now()}
	ctx := GameContext{Chapter: 3}

	// Trigger rule twice
	c.Check(action, ctx)
	c.Check(action, ctx)

	history := c.GetTriggerHistory()
	if len(history) != 2 {
		t.Errorf("History should have 2 records, got %d", len(history))
	}

	// Check record content (AC5)
	record := history[0]
	if record.RuleID != "history-test" {
		t.Errorf("Record RuleID = %s, want history-test", record.RuleID)
	}
	if record.Chapter != 3 {
		t.Errorf("Record Chapter = %d, want 3", record.Chapter)
	}
	if record.Action == "" {
		t.Error("Record Action should not be empty")
	}
}

func TestCheckerGetTriggersByRuleID(t *testing.T) {
	rs := NewRuleSet()

	rule1 := NewRule("rule-1", RuleTypeBehavior)
	rule1.Trigger = Condition{Type: "action", Value: "test"}
	rule1.Consequence = Outcome{Type: ConsequenceWarning}
	rule1.Clues = []string{"test"}
	rs.Add(rule1)

	rule2 := NewRule("rule-2", RuleTypeScenario)
	rule2.Trigger = Condition{Type: "location", Value: "room"}
	rule2.Consequence = Outcome{Type: ConsequenceWarning}
	rule2.Clues = []string{"test"}
	rs.Add(rule2)

	c := NewChecker(rs, game.DifficultyHard)

	// Trigger rule-1 twice
	c.Check(PlayerAction{Type: "test"}, GameContext{})
	c.Check(PlayerAction{Type: "test"}, GameContext{})

	// Trigger rule-2 once
	c.Check(PlayerAction{Type: "move"}, GameContext{Location: "room"})

	rule1Triggers := c.GetTriggersByRuleID("rule-1")
	if len(rule1Triggers) != 2 {
		t.Errorf("rule-1 should have 2 triggers, got %d", len(rule1Triggers))
	}

	rule2Triggers := c.GetTriggersByRuleID("rule-2")
	if len(rule2Triggers) != 1 {
		t.Errorf("rule-2 should have 1 trigger, got %d", len(rule2Triggers))
	}
}

func TestCheckerClearHistory(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("test", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "test"}
	rule.Consequence = Outcome{Type: ConsequenceWarning}
	rule.Clues = []string{"test"}
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	c.Check(PlayerAction{Type: "test"}, GameContext{})

	if len(c.GetTriggerHistory()) != 1 {
		t.Error("Should have 1 record before clear")
	}

	c.ClearHistory()

	if len(c.GetTriggerHistory()) != 0 {
		t.Error("Should have 0 records after clear")
	}
}

func TestCheckerGetFatalTriggers(t *testing.T) {
	rs := NewRuleSet()

	fatalRule := NewRule("fatal", RuleTypeBehavior)
	fatalRule.Trigger = Condition{Type: "action", Value: "fatal-action"}
	fatalRule.Consequence = Outcome{Type: ConsequenceInstantDeath}
	fatalRule.Clues = []string{"test"}
	rs.Add(fatalRule)

	nonFatalRule := NewRule("warning", RuleTypeBehavior)
	nonFatalRule.Trigger = Condition{Type: "action", Value: "warning-action"}
	nonFatalRule.Consequence = Outcome{Type: ConsequenceWarning}
	nonFatalRule.Clues = []string{"test"}
	rs.Add(nonFatalRule)

	c := NewChecker(rs, game.DifficultyHard)

	// Trigger non-fatal
	c.Check(PlayerAction{Type: "warning-action"}, GameContext{})
	// Trigger fatal
	c.Check(PlayerAction{Type: "fatal-action"}, GameContext{})

	fatals := c.GetFatalTriggers()
	if len(fatals) != 1 {
		t.Errorf("Should have 1 fatal trigger, got %d", len(fatals))
	}
	if fatals[0].RuleID != "fatal" {
		t.Errorf("Fatal trigger should be 'fatal', got %s", fatals[0].RuleID)
	}
}

func TestCheckerContainsOperator(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("contains-test", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "run", Operator: "contains"}
	rule.Consequence = Outcome{Type: ConsequenceWarning}
	rule.Clues = []string{"test"}
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)

	// Should match "running" because it contains "run"
	result := c.Check(PlayerAction{Type: "running"}, GameContext{})
	if len(result.Triggers) != 1 {
		t.Error("'running' should trigger rule with 'contains' operator for 'run'")
	}
}

func TestCheckerInactiveRule(t *testing.T) {
	rs := NewRuleSet()
	rule := NewRule("inactive-rule", RuleTypeBehavior)
	rule.Trigger = Condition{Type: "action", Value: "test"}
	rule.Consequence = Outcome{Type: ConsequenceWarning}
	rule.Clues = []string{"test"}
	rule.Active = false // Inactive rule
	rs.Add(rule)

	c := NewChecker(rs, game.DifficultyHard)
	result := c.Check(PlayerAction{Type: "test"}, GameContext{})

	if len(result.Triggers) != 0 {
		t.Error("Inactive rule should not trigger")
	}
}

func TestCheckerTotalDamageAccumulation(t *testing.T) {
	rs := NewRuleSet()

	rule1 := NewRule("rule-1", RuleTypeScenario)
	rule1.Trigger = Condition{Type: "location", Value: "room"}
	rule1.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 10, SANDamage: 5}
	rule1.Clues = []string{"test"}
	rule1.MaxViolations = 0
	rs.Add(rule1)

	rule2 := NewRule("rule-2", RuleTypeBehavior)
	rule2.Trigger = Condition{Type: "action", Value: "look"}
	rule2.Consequence = Outcome{Type: ConsequenceDamage, HPDamage: 5, SANDamage: 10}
	rule2.Clues = []string{"test"}
	rule2.MaxViolations = 0
	rs.Add(rule2)

	c := NewChecker(rs, game.DifficultyHard)
	result := c.Check(PlayerAction{Type: "look"}, GameContext{Location: "room"})

	// Both rules should trigger and damage should accumulate
	if result.TotalHPDamage != 15 {
		t.Errorf("TotalHPDamage = %d, want 15", result.TotalHPDamage)
	}
	if result.TotalSANDamage != 15 {
		t.Errorf("TotalSANDamage = %d, want 15", result.TotalSANDamage)
	}
}
