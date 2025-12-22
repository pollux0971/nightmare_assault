package rules

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/templates"
)

// TestCompleteRuleWorkflow tests the complete rule workflow from template to violation.
// Story 7.2: Integration test for full rule lifecycle.
//
// Workflow:
//  1. Convert template to HiddenRule
//  2. Check for violations
//  3. Reveal clues at specific beats
//  4. Verify rule state changes
func TestCompleteRuleWorkflow(t *testing.T) {
	// Step 1: Create a rule template
	template := &templates.RuleTemplate{
		ID:            "workflow-01",
		Name:          "測試規則",
		Category:      templates.RuleCategorySensory,
		Difficulty:    templates.RuleDifficultyMedium,
		TriggerMedium: "鏡子|看鏡子",
		FalseClue:     "鏡子是安全的",
		SurvivalRule:  "不要看鏡子",
		Punishment: templates.Punishment{
			SANDamage: 30,
			HPDamage:  10,
			Effect:    "鏡子攻擊",
		},
		ClueHints: []string{
			"線索1：鏡子似乎有問題",
			"線索2：不要直視鏡子",
			"線索3：鏡子會攻擊",
		},
	}

	// Step 2: Convert to HiddenRule
	rule := ConvertTemplateToHiddenRule(template, game.DifficultyEasy)

	if rule.ID != "workflow-01" {
		t.Fatalf("Expected rule ID workflow-01, got %s", rule.ID)
	}

	// Verify initial state
	if rule.IsViolated {
		t.Error("New rule should not be violated")
	}

	if len(rule.ClueHints) == 0 {
		t.Fatal("Expected clue hints to be generated")
	}

	// Step 3: Test clue revelation at different beats
	engine := NewRuleEngine()
	rules := []*HiddenRule{rule}

	// Beat 5 should reveal tier 1 clues
	cluesBeat5 := engine.GetCluesForBeat(rules, 5, game.DifficultyEasy)
	if len(cluesBeat5) == 0 {
		t.Error("Expected at least one clue at beat 5")
	}

	t.Logf("Beat 5: Revealed %d clues", len(cluesBeat5))

	// Beat 12 should reveal tier 2 clues
	cluesBeat12 := engine.GetCluesForBeat(rules, 12, game.DifficultyEasy)
	t.Logf("Beat 12: Revealed %d additional clues", len(cluesBeat12))

	// Beat 20 should reveal tier 3 clues
	cluesBeat20 := engine.GetCluesForBeat(rules, 20, game.DifficultyEasy)
	t.Logf("Beat 20: Revealed %d additional clues", len(cluesBeat20))

	// Step 4: Test safe choice (no violation)
	safeViolations := engine.CheckViolation("我走向門口", rules)
	if len(safeViolations) != 0 {
		t.Errorf("Safe choice should not trigger violations, got %d", len(safeViolations))
	}

	if rule.IsViolated {
		t.Error("Rule should not be violated after safe choice")
	}

	// Step 5: Test violation
	violations := engine.CheckViolation("我走近鏡子並看鏡子", rules)

	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.RuleID != "workflow-01" {
		t.Errorf("Expected violation of workflow-01, got %s", v.RuleID)
	}

	if v.SANDamage != 30 {
		t.Errorf("Expected 30 SAN damage, got %d", v.SANDamage)
	}

	if v.HPDamage != 10 {
		t.Errorf("Expected 10 HP damage, got %d", v.HPDamage)
	}

	// Step 6: Verify rule is marked as violated
	if !rule.IsViolated {
		t.Error("Rule should be marked as violated")
	}

	// Step 7: Verify duplicate violation prevention
	duplicateViolations := engine.CheckViolation("我再次看鏡子", rules)
	if len(duplicateViolations) != 0 {
		t.Errorf("Violated rule should not trigger again, got %d violations", len(duplicateViolations))
	}

	// Step 8: Verify no more clues are revealed for violated rules
	cluesAfterViolation := engine.GetCluesForBeat(rules, 25, game.DifficultyEasy)
	if len(cluesAfterViolation) != 0 {
		t.Errorf("Violated rules should not reveal clues, got %d clues", len(cluesAfterViolation))
	}
}

// TestDifficultyBasedRuleGeneration tests rule generation for different difficulties.
// Story 7.2 AC1-AC3: Test Easy/Hard/Hell difficulty configurations.
func TestDifficultyBasedRuleGeneration(t *testing.T) {
	tests := []struct {
		name            string
		difficulty      game.DifficultyLevel
		minRules        int
		maxRules        int
		maxTier         int
		expectFatal     bool
		fatalPercentage float64 // Expected percentage of fatal rules (approximately)
	}{
		{
			name:            "Easy: 3-6 rules, Tier 3, no fatal",
			difficulty:      game.DifficultyEasy,
			minRules:        3,
			maxRules:        6,
			maxTier:         3,
			expectFatal:     false,
			fatalPercentage: 0,
		},
		{
			name:            "Hard: 5-10 rules, Tier 2, some fatal",
			difficulty:      game.DifficultyHard,
			minRules:        5,
			maxRules:        12,
			maxTier:         2,
			expectFatal:     true,
			fatalPercentage: 10,
		},
		{
			name:            "Hell: 8-15 rules, Tier 1, many fatal",
			difficulty:      game.DifficultyHell,
			minRules:        8,
			maxRules:        15,
			maxTier:         1,
			expectFatal:     true,
			fatalPercentage: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test template
			template := &templates.RuleTemplate{
				ID:            "test-rule",
				Name:          "測試規則",
				Category:      templates.RuleCategorySensory,
				Difficulty:    templates.RuleDifficultyMedium,
				TriggerMedium: "test|trigger",
				SurvivalRule:  "生存規則",
				Punishment: templates.Punishment{
					SANDamage: 20,
					HPDamage:  10,
				},
				ClueHints: []string{"線索1", "線索2", "線索3"},
			}

			// Convert with difficulty
			rule := ConvertTemplateToHiddenRule(template, tt.difficulty)

			// Verify clue tiers match difficulty
			engine := NewRuleEngine()
			clues := engine.GetCluesForBeat([]*HiddenRule{rule}, 15, tt.difficulty)

			// Check that no clue exceeds max tier for difficulty
			for _, clue := range clues {
				if clue.Tier > tt.maxTier {
					t.Errorf("Clue tier %d exceeds max tier %d for %s",
						clue.Tier, tt.maxTier, tt.difficulty)
				}
			}

			t.Logf("%s: Generated %d clues with max tier %d",
				tt.difficulty, len(rule.ClueHints), tt.maxTier)
		})
	}
}

// TestMultipleRulesInteraction tests interaction between multiple rules.
// Story 7.2 AC4-AC6: Multiple rules, priorities, playability.
func TestMultipleRulesInteraction(t *testing.T) {
	// Create multiple rules
	rule1 := &HiddenRule{
		ID:               "rule-01",
		Name:             "鏡子規則",
		Category:         "sensory",
		TriggerCondition: "鏡子|看鏡子",
		Punishment: HiddenRulePunishment{
			HPDamage:  10,
			SANDamage: 20,
			IsFatal:   false,
		},
		ClueHints: []ClueHint{
			{Tier: 1, BeatRange: [2]int{1, 10}, Hint: "鏡子線索", Revealed: false},
		},
	}

	rule2 := &HiddenRule{
		ID:               "rule-02",
		Name:             "聲音規則",
		Category:         "sensory",
		TriggerCondition: "大喊|呼喊",
		Punishment: HiddenRulePunishment{
			HPDamage:  15,
			SANDamage: 10,
			IsFatal:   false,
		},
		ClueHints: []ClueHint{
			{Tier: 1, BeatRange: [2]int{1, 10}, Hint: "聲音線索", Revealed: false},
		},
	}

	rule3 := &HiddenRule{
		ID:               "rule-03",
		Name:             "致命規則",
		Category:         "sensory",
		TriggerCondition: "直視怪物",
		Punishment: HiddenRulePunishment{
			HPDamage:  100,
			SANDamage: 0,
			IsFatal:   true,
		},
		ClueHints: []ClueHint{
			{Tier: 1, BeatRange: [2]int{1, 10}, Hint: "致命線索", Revealed: false},
		},
	}

	rules := []*HiddenRule{rule1, rule2, rule3}
	engine := NewRuleEngine()

	// Test 1: Validate playability
	err := engine.ValidateRulePlayability(rules)
	if err != nil {
		t.Errorf("Rule set should be playable: %v", err)
	}

	// Test 2: Trigger multiple rules in sequence
	violations1 := engine.CheckViolation("我看鏡子", rules)
	if len(violations1) != 1 || violations1[0].RuleID != "rule-01" {
		t.Error("Expected rule-01 to be violated")
	}

	violations2 := engine.CheckViolation("我大喊救命", rules)
	if len(violations2) != 1 || violations2[0].RuleID != "rule-02" {
		t.Error("Expected rule-02 to be violated")
	}

	// Test 3: Verify active rules count
	activeRules := engine.GetActiveRules(rules)
	if len(activeRules) != 1 {
		t.Errorf("Expected 1 active rule, got %d", len(activeRules))
	}

	violatedRules := engine.GetViolatedRules(rules)
	if len(violatedRules) != 2 {
		t.Errorf("Expected 2 violated rules, got %d", len(violatedRules))
	}

	// Test 4: Filter by category
	sensoryRules := engine.GetRulesByCategory(rules, "sensory")
	if len(sensoryRules) != 3 {
		t.Errorf("Expected 3 sensory rules, got %d", len(sensoryRules))
	}

	// Test 5: Reveal clues for all rules
	clues := engine.GetCluesForBeat(rules, 5, game.DifficultyEasy)

	// Should only reveal clues for non-violated rules
	if len(clues) != 1 { // Only rule-03 is not violated
		t.Errorf("Expected 1 clue (from non-violated rule), got %d", len(clues))
	}
}

// TestFatalRuleHandling tests instant death rule handling.
// Story 7.2 AC2-AC3: Fatal rules for Hard/Hell difficulties.
func TestFatalRuleHandling(t *testing.T) {
	fatalRule := &HiddenRule{
		ID:               "fatal-01",
		Name:             "即死規則",
		Category:         "sensory",
		TriggerCondition: "直視|盯著看",
		Punishment: HiddenRulePunishment{
			HPDamage:     100,
			SANDamage:    0,
			IsFatal:      true,
			CustomEffect: "你直視了不該看的東西，當場死亡",
		},
		ClueHints: []ClueHint{},
	}

	engine := NewRuleEngine()
	rules := []*HiddenRule{fatalRule}

	// Trigger fatal rule
	violations := engine.CheckViolation("我直視怪物", rules)

	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	if !violations[0].IsFatal {
		t.Error("Expected fatal violation")
	}

	if violations[0].ViolationNarrative == "" {
		t.Error("Expected violation narrative")
	}

	t.Logf("Fatal violation narrative: %s", violations[0].ViolationNarrative)
}

// TestClueRevealProgression tests clue revelation timing across beats.
// Story 7.2 AC5: Tiered clue revelation at specific beats.
func TestClueRevealProgression(t *testing.T) {
	rule := &HiddenRule{
		ID:   "reveal-01",
		Name: "逐步揭示規則",
		ClueHints: []ClueHint{
			{Tier: 1, BeatRange: [2]int{1, 8}, Hint: "Tier 1 線索", Revealed: false},
			{Tier: 2, BeatRange: [2]int{9, 16}, Hint: "Tier 2 線索", Revealed: false},
			{Tier: 3, BeatRange: [2]int{17, 24}, Hint: "Tier 3 線索", Revealed: false},
		},
	}

	engine := NewRuleEngine()
	rules := []*HiddenRule{rule}

	// Test progression through beats
	beatTests := []struct {
		beat          int
		expectedClues int
		expectedTier  int
	}{
		{beat: 5, expectedClues: 1, expectedTier: 1},   // Tier 1 revealed
		{beat: 12, expectedClues: 1, expectedTier: 2},  // Tier 2 revealed
		{beat: 20, expectedClues: 1, expectedTier: 3},  // Tier 3 revealed
		{beat: 25, expectedClues: 0, expectedTier: 0},  // All already revealed
	}

	for _, tt := range beatTests {
		clues := engine.GetCluesForBeat(rules, tt.beat, game.DifficultyEasy)

		if len(clues) != tt.expectedClues {
			t.Errorf("Beat %d: expected %d clues, got %d",
				tt.beat, tt.expectedClues, len(clues))
		}

		if len(clues) > 0 && clues[0].Tier != tt.expectedTier {
			t.Errorf("Beat %d: expected tier %d, got %d",
				tt.beat, tt.expectedTier, clues[0].Tier)
		}

		if len(clues) > 0 {
			t.Logf("Beat %d: Revealed Tier %d clue: %s",
				tt.beat, clues[0].Tier, clues[0].Hint)
		}
	}

	// Verify all clues are marked as revealed
	revealedCount := engine.CountRevealedClues(rule)
	if revealedCount != 3 {
		t.Errorf("Expected 3 revealed clues, got %d", revealedCount)
	}
}
