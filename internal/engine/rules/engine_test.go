package rules

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// TestCheckViolation_NoViolation tests that no violations are detected for safe choices.
func TestCheckViolation_NoViolation(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:               "test-01",
			Name:             "鏡子規則",
			TriggerCondition: "鏡子|看鏡子|照鏡子",
			IsViolated:       false,
			Punishment: HiddenRulePunishment{
				HPDamage:  10,
				SANDamage: 20,
				IsFatal:   false,
			},
		},
	}

	// Safe choice that doesn't trigger the rule
	violations := engine.CheckViolation("我走向門口", rules)

	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}

	if rules[0].IsViolated {
		t.Error("Rule should not be marked as violated")
	}
}

// TestCheckViolation_SingleViolation tests detecting a single rule violation.
func TestCheckViolation_SingleViolation(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:               "test-01",
			Name:             "鏡子規則",
			TriggerCondition: "鏡子|看鏡子|照鏡子",
			IsViolated:       false,
			Punishment: HiddenRulePunishment{
				HPDamage:  10,
				SANDamage: 20,
				IsFatal:   false,
			},
		},
	}

	// Choice that triggers the rule
	violations := engine.CheckViolation("我走向鏡子並照鏡子", rules)

	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.RuleID != "test-01" {
		t.Errorf("Expected RuleID test-01, got %s", v.RuleID)
	}
	if v.HPDamage != 10 {
		t.Errorf("Expected HPDamage 10, got %d", v.HPDamage)
	}
	if v.SANDamage != 20 {
		t.Errorf("Expected SANDamage 20, got %d", v.SANDamage)
	}
	if v.IsFatal {
		t.Error("Expected non-fatal violation")
	}

	if !rules[0].IsViolated {
		t.Error("Rule should be marked as violated")
	}
}

// TestCheckViolation_PreventDuplicateTrigger tests that violated rules don't trigger again.
// Story 7.2 AC4: is_violated=true prevents duplicate triggers
func TestCheckViolation_PreventDuplicateTrigger(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:               "test-01",
			Name:             "鏡子規則",
			TriggerCondition: "鏡子|看鏡子",
			IsViolated:       false,
			Punishment: HiddenRulePunishment{
				HPDamage:  10,
				SANDamage: 20,
			},
		},
	}

	// First violation
	violations1 := engine.CheckViolation("我看鏡子", rules)
	if len(violations1) != 1 {
		t.Fatalf("Expected 1 violation on first trigger, got %d", len(violations1))
	}

	// Second attempt with same choice should not trigger again
	violations2 := engine.CheckViolation("我看鏡子", rules)
	if len(violations2) != 0 {
		t.Errorf("Expected no violations on second trigger, got %d", len(violations2))
	}
}

// TestCheckViolation_MultipleRules tests checking multiple rules at once.
func TestCheckViolation_MultipleRules(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:               "test-01",
			Name:             "鏡子規則",
			TriggerCondition: "鏡子|看鏡子",
			IsViolated:       false,
			Punishment:       HiddenRulePunishment{HPDamage: 10, SANDamage: 20},
		},
		{
			ID:               "test-02",
			Name:             "聲音規則",
			TriggerCondition: "大喊|呼喊|叫喊",
			IsViolated:       false,
			Punishment:       HiddenRulePunishment{HPDamage: 15, SANDamage: 10},
		},
	}

	// Choice that triggers only the second rule
	violations := engine.CheckViolation("我大喊求救", rules)

	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].RuleID != "test-02" {
		t.Errorf("Expected test-02 to be violated, got %s", violations[0].RuleID)
	}

	if rules[0].IsViolated {
		t.Error("Rule test-01 should not be violated")
	}
	if !rules[1].IsViolated {
		t.Error("Rule test-02 should be violated")
	}
}

// TestCheckViolation_Fatal tests fatal rule violation.
// Story 7.2 AC4: IsFatal triggers death flow
func TestCheckViolation_Fatal(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:               "test-01",
			Name:             "致命規則",
			TriggerCondition: "直視|盯著看",
			IsViolated:       false,
			Punishment: HiddenRulePunishment{
				HPDamage:  100,
				SANDamage: 0,
				IsFatal:   true,
			},
		},
	}

	violations := engine.CheckViolation("我直視怪物", rules)

	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	if !violations[0].IsFatal {
		t.Error("Expected fatal violation")
	}
}

// TestGetCluesForBeat_NoClues tests when no clues should be revealed.
func TestGetCluesForBeat_NoClues(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:   "test-01",
			Name: "測試規則",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{5, 10}, Hint: "線索1", Revealed: false},
			},
		},
	}

	// Beat 2 is before the clue's beat range (5-10)
	clues := engine.GetCluesForBeat(rules, 2, game.DifficultyEasy)

	if len(clues) != 0 {
		t.Errorf("Expected no clues at beat 2, got %d", len(clues))
	}
}

// TestGetCluesForBeat_SingleClue tests revealing a single clue.
// Story 7.2 AC5: Clues revealed based on beat range
func TestGetCluesForBeat_SingleClue(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:   "test-01",
			Name: "測試規則",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{5, 10}, Hint: "線索1", Revealed: false},
			},
		},
	}

	// Beat 7 is within the clue's beat range (5-10)
	clues := engine.GetCluesForBeat(rules, 7, game.DifficultyEasy)

	if len(clues) != 1 {
		t.Fatalf("Expected 1 clue at beat 7, got %d", len(clues))
	}

	if clues[0].RuleID != "test-01" {
		t.Errorf("Expected RuleID test-01, got %s", clues[0].RuleID)
	}
	if clues[0].Tier != 1 {
		t.Errorf("Expected Tier 1, got %d", clues[0].Tier)
	}
	if clues[0].Hint != "線索1" {
		t.Errorf("Expected hint '線索1', got '%s'", clues[0].Hint)
	}

	// Verify clue is marked as revealed
	if !rules[0].ClueHints[0].Revealed {
		t.Error("Clue should be marked as revealed")
	}
}

// TestGetCluesForBeat_PreventDuplicateReveal tests that revealed clues aren't shown again.
// Story 7.2 AC5: Revealed=true prevents duplicate revelation
func TestGetCluesForBeat_PreventDuplicateReveal(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:   "test-01",
			Name: "測試規則",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{5, 10}, Hint: "線索1", Revealed: false},
			},
		},
	}

	// First revelation at beat 7
	clues1 := engine.GetCluesForBeat(rules, 7, game.DifficultyEasy)
	if len(clues1) != 1 {
		t.Fatalf("Expected 1 clue on first call, got %d", len(clues1))
	}

	// Second call at same beat should not reveal again
	clues2 := engine.GetCluesForBeat(rules, 7, game.DifficultyEasy)
	if len(clues2) != 0 {
		t.Errorf("Expected no clues on second call, got %d", len(clues2))
	}
}

// TestGetCluesForBeat_DifficultyTierLimit tests difficulty-based tier limits.
// Story 7.2 AC5: Easy=Tier3, Hard=Tier2, Hell=Tier1
func TestGetCluesForBeat_DifficultyTierLimit(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:   "test-01",
			Name: "測試規則",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{1, 20}, Hint: "Tier1線索", Revealed: false},
				{Tier: 2, BeatRange: [2]int{1, 20}, Hint: "Tier2線索", Revealed: false},
				{Tier: 3, BeatRange: [2]int{1, 20}, Hint: "Tier3線索", Revealed: false},
			},
		},
	}

	// Test Easy difficulty (max tier 3)
	cluesEasy := engine.GetCluesForBeat(rules, 10, game.DifficultyEasy)
	if len(cluesEasy) != 3 {
		t.Errorf("Easy should reveal 3 clues, got %d", len(cluesEasy))
	}

	// Reset revealed flags
	for i := range rules[0].ClueHints {
		rules[0].ClueHints[i].Revealed = false
	}

	// Test Hard difficulty (max tier 2)
	cluesHard := engine.GetCluesForBeat(rules, 10, game.DifficultyHard)
	if len(cluesHard) != 2 {
		t.Errorf("Hard should reveal 2 clues, got %d", len(cluesHard))
	}

	// Reset revealed flags
	for i := range rules[0].ClueHints {
		rules[0].ClueHints[i].Revealed = false
	}

	// Test Hell difficulty (max tier 1)
	cluesHell := engine.GetCluesForBeat(rules, 10, game.DifficultyHell)
	if len(cluesHell) != 1 {
		t.Errorf("Hell should reveal 1 clue, got %d", len(cluesHell))
	}
}

// TestGetCluesForBeat_SkipViolatedRules tests that violated rules don't reveal clues.
// Story 7.2 AC5: Players already know about violated rules
func TestGetCluesForBeat_SkipViolatedRules(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:         "test-01",
			Name:       "已觸發規則",
			IsViolated: true,
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{1, 20}, Hint: "線索", Revealed: false},
			},
		},
	}

	clues := engine.GetCluesForBeat(rules, 10, game.DifficultyEasy)

	if len(clues) != 0 {
		t.Errorf("Violated rules should not reveal clues, got %d clues", len(clues))
	}
}

// TestValidateRulePlayability_AllFatalNoClues tests unplayable rule set detection.
// Story 7.2 AC6: Prevent no-win scenarios
func TestValidateRulePlayability_AllFatalNoClues(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:        "test-01",
			Name:      "致命規則1",
			ClueHints: []ClueHint{}, // No clues
			Punishment: HiddenRulePunishment{
				IsFatal: true,
			},
		},
		{
			ID:        "test-02",
			Name:      "致命規則2",
			ClueHints: []ClueHint{}, // No clues
			Punishment: HiddenRulePunishment{
				IsFatal: true,
			},
		},
	}

	err := engine.ValidateRulePlayability(rules)

	if err == nil {
		t.Error("Expected error for unplayable rule set")
	}
}

// TestValidateRulePlayability_ValidSet tests playable rule set.
func TestValidateRulePlayability_ValidSet(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{
			ID:   "test-01",
			Name: "規則1",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{1, 10}, Hint: "線索"},
			},
			Punishment: HiddenRulePunishment{
				IsFatal: false,
			},
		},
		{
			ID:   "test-02",
			Name: "規則2",
			ClueHints: []ClueHint{
				{Tier: 1, BeatRange: [2]int{1, 10}, Hint: "線索"},
			},
			Punishment: HiddenRulePunishment{
				IsFatal: true,
			},
		},
	}

	err := engine.ValidateRulePlayability(rules)

	if err != nil {
		t.Errorf("Expected valid rule set, got error: %v", err)
	}
}

// TestGetRulesByCategory tests filtering rules by category.
func TestGetRulesByCategory(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{ID: "test-01", Category: "sensory"},
		{ID: "test-02", Category: "spatial"},
		{ID: "test-03", Category: "sensory"},
	}

	sensory := engine.GetRulesByCategory(rules, "sensory")

	if len(sensory) != 2 {
		t.Errorf("Expected 2 sensory rules, got %d", len(sensory))
	}

	for _, rule := range sensory {
		if rule.Category != "sensory" {
			t.Errorf("Expected sensory category, got %s", rule.Category)
		}
	}
}

// TestGetActiveRules tests filtering active (non-violated) rules.
func TestGetActiveRules(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{ID: "test-01", IsViolated: false},
		{ID: "test-02", IsViolated: true},
		{ID: "test-03", IsViolated: false},
	}

	active := engine.GetActiveRules(rules)

	if len(active) != 2 {
		t.Errorf("Expected 2 active rules, got %d", len(active))
	}

	for _, rule := range active {
		if rule.IsViolated {
			t.Errorf("Active rules should not be violated: %s", rule.ID)
		}
	}
}

// TestGetViolatedRules tests filtering violated rules.
func TestGetViolatedRules(t *testing.T) {
	engine := NewRuleEngine()
	rules := []*HiddenRule{
		{ID: "test-01", IsViolated: false},
		{ID: "test-02", IsViolated: true},
		{ID: "test-03", IsViolated: true},
	}

	violated := engine.GetViolatedRules(rules)

	if len(violated) != 2 {
		t.Errorf("Expected 2 violated rules, got %d", len(violated))
	}

	for _, rule := range violated {
		if !rule.IsViolated {
			t.Errorf("Violated rules should be marked: %s", rule.ID)
		}
	}
}

// TestCountRevealedClues tests counting revealed clues.
func TestCountRevealedClues(t *testing.T) {
	engine := NewRuleEngine()
	rule := &HiddenRule{
		ClueHints: []ClueHint{
			{Revealed: true},
			{Revealed: false},
			{Revealed: true},
		},
	}

	count := engine.CountRevealedClues(rule)

	if count != 2 {
		t.Errorf("Expected 2 revealed clues, got %d", count)
	}
}

// TestGetRevealedClues tests getting revealed clue texts.
func TestGetRevealedClues(t *testing.T) {
	engine := NewRuleEngine()
	rule := &HiddenRule{
		ClueHints: []ClueHint{
			{Hint: "線索1", Revealed: true},
			{Hint: "線索2", Revealed: false},
			{Hint: "線索3", Revealed: true},
		},
	}

	clues := engine.GetRevealedClues(rule)

	if len(clues) != 2 {
		t.Fatalf("Expected 2 revealed clues, got %d", len(clues))
	}

	if clues[0] != "線索1" || clues[1] != "線索3" {
		t.Errorf("Expected [線索1, 線索3], got %v", clues)
	}
}
