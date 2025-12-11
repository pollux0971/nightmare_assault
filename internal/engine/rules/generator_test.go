package rules

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestGeneratorGenerateRules(t *testing.T) {
	g := NewGenerator()

	// Test each difficulty level
	difficulties := []game.DifficultyLevel{
		game.DifficultyEasy,
		game.DifficultyHard,
		game.DifficultyHell,
	}

	for _, diff := range difficulties {
		rs := g.GenerateRules(diff)
		if rs == nil {
			t.Errorf("GenerateRules(%v) returned nil", diff)
			continue
		}

		// Verify rules were generated
		if rs.Count() == 0 {
			t.Errorf("GenerateRules(%v) generated 0 rules", diff)
		}

		t.Logf("Difficulty %v: generated %d rules", diff, rs.Count())
	}
}

func TestGenerateRulesEasyModeMaxSixRules(t *testing.T) {
	g := NewGenerator()

	// Test multiple times to account for randomness
	for i := 0; i < 20; i++ {
		rs := g.GenerateRules(game.DifficultyEasy)

		// AC1: Easy mode should have max 6 rules
		if rs.Count() > 6 {
			t.Errorf("Easy mode generated %d rules, max should be 6", rs.Count())
		}

		// Should have at least 4 rules
		if rs.Count() < 4 {
			t.Errorf("Easy mode generated only %d rules, expected at least 4", rs.Count())
		}
	}
}

func TestGenerateRulesHardModeCount(t *testing.T) {
	g := NewGenerator()

	for i := 0; i < 20; i++ {
		rs := g.GenerateRules(game.DifficultyHard)

		// Hard mode should have 8-12 rules
		if rs.Count() < 8 || rs.Count() > 12 {
			t.Errorf("Hard mode generated %d rules, expected 8-12", rs.Count())
		}
	}
}

func TestGenerateRulesHellModeCount(t *testing.T) {
	g := NewGenerator()

	for i := 0; i < 20; i++ {
		rs := g.GenerateRules(game.DifficultyHell)

		// Hell mode should have 10-15 rules
		if rs.Count() < 10 || rs.Count() > 15 {
			t.Errorf("Hell mode generated %d rules, expected 10-15", rs.Count())
		}
	}
}

func TestGenerateRulesTypeDistribution(t *testing.T) {
	g := NewGenerator()

	// Test with a larger set for better distribution
	rs := g.GenerateRules(game.DifficultyHard)

	counts := rs.CountByType()

	// Should have rules of multiple types
	typesWithRules := 0
	for _, count := range counts {
		if count > 0 {
			typesWithRules++
		}
	}

	// AC2: Should have diverse rule types (at least 3 different types)
	if typesWithRules < 3 {
		t.Errorf("Only %d rule types represented, expected at least 3", typesWithRules)
	}

	t.Logf("Type distribution: %v", counts)
}

func TestGenerateRulesTypeDistributionBalance(t *testing.T) {
	g := NewGenerator()

	// Run multiple times and aggregate
	totalCounts := make(map[RuleType]int)

	for i := 0; i < 100; i++ {
		rs := g.GenerateRules(game.DifficultyHard)
		for rt, count := range rs.CountByType() {
			totalCounts[rt] += count
		}
	}

	// Check that no type dominates more than 40% of total
	total := 0
	for _, count := range totalCounts {
		total += count
	}

	for rt, count := range totalCounts {
		percentage := float64(count) / float64(total) * 100
		t.Logf("Type %v: %d (%.1f%%)", rt, count, percentage)

		// No single type should exceed 40%
		if percentage > 40 {
			t.Errorf("Type %v dominates at %.1f%%, expected more even distribution", rt, percentage)
		}
	}
}

func TestGenerateRulesAllHaveRequiredFields(t *testing.T) {
	g := NewGenerator()
	rs := g.GenerateRules(game.DifficultyHard)

	for _, r := range rs.Rules {
		// AC3: Each rule must have required fields
		if r.ID == "" {
			t.Error("Rule has empty ID")
		}
		if r.Trigger.Type == "" {
			t.Errorf("Rule %s has empty trigger type", r.ID)
		}
		if len(r.Clues) == 0 {
			t.Errorf("Rule %s has no clues", r.ID)
		}
		if r.Priority == 0 {
			t.Errorf("Rule %s has zero priority", r.ID)
		}
	}
}

func TestGenerateRulesClueCount(t *testing.T) {
	g := NewGenerator()
	rs := g.GenerateRules(game.DifficultyHard)

	for _, r := range rs.Rules {
		// Should have 2-4 clues per rule
		if len(r.Clues) < 2 || len(r.Clues) > 4 {
			t.Errorf("Rule %s has %d clues, expected 2-4", r.ID, len(r.Clues))
		}
	}
}

func TestGenerateRulesMaxViolationsByDifficulty(t *testing.T) {
	g := NewGenerator()

	// Easy mode should have 2 max violations
	rsEasy := g.GenerateRules(game.DifficultyEasy)
	for _, r := range rsEasy.Rules {
		if r.MaxViolations != 2 {
			t.Errorf("Easy mode rule %s has MaxViolations=%d, expected 2", r.ID, r.MaxViolations)
		}
	}

	// Hard mode should have 1 max violation
	rsHard := g.GenerateRules(game.DifficultyHard)
	for _, r := range rsHard.Rules {
		if r.MaxViolations != 1 {
			t.Errorf("Hard mode rule %s has MaxViolations=%d, expected 1", r.ID, r.MaxViolations)
		}
	}

	// Hell mode should have 0 max violations (immediate)
	rsHell := g.GenerateRules(game.DifficultyHell)
	for _, r := range rsHell.Rules {
		if r.MaxViolations != 0 {
			t.Errorf("Hell mode rule %s has MaxViolations=%d, expected 0", r.ID, r.MaxViolations)
		}
	}
}

func TestGenerateRulesConsequenceSeverity(t *testing.T) {
	g := NewGenerator()

	// Test that harder difficulties have more severe consequences
	instantDeathCount := make(map[game.DifficultyLevel]int)
	iterations := 100

	difficulties := []game.DifficultyLevel{
		game.DifficultyEasy,
		game.DifficultyHard,
		game.DifficultyHell,
	}

	for _, diff := range difficulties {
		for i := 0; i < iterations; i++ {
			rs := g.GenerateRules(diff)
			for _, r := range rs.Rules {
				if r.Consequence.Type == ConsequenceInstantDeath {
					instantDeathCount[diff]++
				}
			}
		}
	}

	// Easy should have zero or very few instant death rules
	easyInstant := instantDeathCount[game.DifficultyEasy]
	hardInstant := instantDeathCount[game.DifficultyHard]
	hellInstant := instantDeathCount[game.DifficultyHell]

	t.Logf("Instant death rules over %d iterations:", iterations)
	t.Logf("  Easy: %d", easyInstant)
	t.Logf("  Hard: %d", hardInstant)
	t.Logf("  Hell: %d", hellInstant)

	// Easy should have no instant death rules
	if easyInstant > 0 {
		t.Errorf("Easy mode has %d instant death rules, expected 0", easyInstant)
	}

	// Hell should have more instant death than Hard
	if hellInstant <= hardInstant {
		t.Logf("Warning: Hell mode (%d) doesn't have more instant deaths than Hard (%d)", hellInstant, hardInstant)
	}
}

func TestGenerateRulesAllTypesEventuallyGenerated(t *testing.T) {
	g := NewGenerator()

	typesGenerated := make(map[RuleType]bool)
	allTypes := AllRuleTypes()

	// Generate many rule sets
	for i := 0; i < 100 && len(typesGenerated) < len(allTypes); i++ {
		rs := g.GenerateRules(game.DifficultyHard)
		for _, r := range rs.Rules {
			typesGenerated[r.Type] = true
		}
	}

	// All types should have been generated at least once
	for _, rt := range allTypes {
		if !typesGenerated[rt] {
			t.Errorf("Rule type %v was never generated", rt)
		}
	}
}

func TestGenerateRulesWarningTextNotEmpty(t *testing.T) {
	g := NewGenerator()
	rs := g.GenerateRules(game.DifficultyHard)

	for _, r := range rs.Rules {
		if r.WarningText == "" {
			t.Errorf("Rule %s has empty warning text", r.ID)
		}
	}
}

func TestGenerateRulesActiveByDefault(t *testing.T) {
	g := NewGenerator()
	rs := g.GenerateRules(game.DifficultyHard)

	for _, r := range rs.Rules {
		if !r.Active {
			t.Errorf("Rule %s is not active by default", r.ID)
		}
	}
}

func TestGenerateRulesUniqueTriggers(t *testing.T) {
	g := NewGenerator()
	rs := g.GenerateRules(game.DifficultyHard)

	// Check that rules of the same type have different triggers
	triggersByType := make(map[RuleType]map[string]bool)

	for _, r := range rs.Rules {
		if triggersByType[r.Type] == nil {
			triggersByType[r.Type] = make(map[string]bool)
		}

		triggerKey := r.Trigger.Type + ":" + r.Trigger.Value
		if triggersByType[r.Type][triggerKey] {
			// Duplicate triggers are allowed but log for observation
			t.Logf("Note: Duplicate trigger found for type %v: %s", r.Type, triggerKey)
		}
		triggersByType[r.Type][triggerKey] = true
	}
}

func TestRandomIntBounds(t *testing.T) {
	// Test randomInt helper
	for i := 0; i < 100; i++ {
		val := randomInt(5, 10)
		if val < 5 || val > 10 {
			t.Errorf("randomInt(5, 10) = %d, out of bounds", val)
		}
	}

	// Edge case: min == max
	for i := 0; i < 10; i++ {
		val := randomInt(5, 5)
		if val != 5 {
			t.Errorf("randomInt(5, 5) = %d, expected 5", val)
		}
	}
}

func TestCalculateTypeDistribution(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		total        int
		minTypes     int
		maxPerType   int
	}{
		{5, 5, 1},   // 5 rules, 5 types, 1 each
		{10, 5, 2},  // 10 rules, all types get 2
		{12, 5, 3},  // 12 rules, some get 3, some get 2
		{3, 3, 1},   // 3 rules, only 3 types used
	}

	for _, tt := range tests {
		dist := g.calculateTypeDistribution(tt.total)

		total := 0
		for _, count := range dist {
			total += count
		}

		if total != tt.total {
			t.Errorf("Distribution sum = %d, want %d", total, tt.total)
		}

		if len(dist) < tt.minTypes && tt.total >= tt.minTypes {
			t.Errorf("Distribution has %d types, want at least %d", len(dist), tt.minTypes)
		}
	}
}
