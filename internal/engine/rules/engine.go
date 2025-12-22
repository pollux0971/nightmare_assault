package rules

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// RuleEngine handles rule violation checking and clue revelation.
// Story 7.2: Core engine for hidden rule management.
type RuleEngine struct{}

// NewRuleEngine creates a new rule engine instance.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// CheckViolation checks if a player choice violates any hidden rules.
// Story 7.2 AC4: Implements rule violation detection with keyword matching.
//
// Parameters:
//   - playerChoice: The player's choice text
//   - rules: Slice of hidden rules to check against
//
// Returns:
//   - Slice of RuleViolation for any violated rules
func (e *RuleEngine) CheckViolation(playerChoice string, rules []*HiddenRule) []RuleViolation {
	violations := []RuleViolation{}

	// Normalize player choice for matching
	normalizedChoice := strings.ToLower(strings.TrimSpace(playerChoice))

	for _, rule := range rules {
		// Skip already violated rules (AC4: prevent duplicate triggers)
		if rule.IsViolated {
			continue
		}

		// Check if this rule is triggered
		if e.matchesTrigger(normalizedChoice, rule) {
			violation := RuleViolation{
				RuleID:             rule.ID,
				RuleName:           rule.Name,
				HPDamage:           rule.Punishment.HPDamage,
				SANDamage:          rule.Punishment.SANDamage,
				IsFatal:            rule.Punishment.IsFatal,
				ViolationNarrative: e.generateViolationNarrative(rule),
			}
			violations = append(violations, violation)

			// Mark rule as violated (AC4: prevent re-triggering)
			rule.IsViolated = true
		}
	}

	return violations
}

// matchesTrigger checks if a player choice matches a rule's trigger condition.
// Story 7.2 AC4: Keyword-based matching logic.
//
// The trigger condition uses pipe-separated keywords (e.g., "看|盯|凝視").
// Returns true if ANY keyword is found in the player choice.
func (e *RuleEngine) matchesTrigger(normalizedChoice string, rule *HiddenRule) bool {
	// Empty trigger condition means rule never triggers
	if rule.TriggerCondition == "" {
		return false
	}

	// Split trigger condition into keywords
	keywords := strings.Split(rule.TriggerCondition, "|")

	// Check if any keyword matches
	for _, keyword := range keywords {
		keyword = strings.ToLower(strings.TrimSpace(keyword))
		if keyword == "" {
			continue
		}

		// Check if keyword is present in choice
		if strings.Contains(normalizedChoice, keyword) {
			return true
		}
	}

	return false
}

// generateViolationNarrative creates a narrative description of the violation.
// Story 7.2 AC4: Generate violation narrative for display.
func (e *RuleEngine) generateViolationNarrative(rule *HiddenRule) string {
	// Use custom effect if available
	if rule.Punishment.CustomEffect != "" {
		return rule.Punishment.CustomEffect
	}

	// Generate based on damage
	if rule.Punishment.IsFatal {
		return fmt.Sprintf("違反了「%s」規則，導致致命後果", rule.Name)
	}

	if rule.Punishment.HPDamage > 0 && rule.Punishment.SANDamage > 0 {
		return fmt.Sprintf("違反了「%s」規則，受到 %d HP 和 %d SAN 損傷",
			rule.Name, rule.Punishment.HPDamage, rule.Punishment.SANDamage)
	}

	if rule.Punishment.HPDamage > 0 {
		return fmt.Sprintf("違反了「%s」規則，受到 %d HP 損傷",
			rule.Name, rule.Punishment.HPDamage)
	}

	if rule.Punishment.SANDamage > 0 {
		return fmt.Sprintf("違反了「%s」規則，受到 %d SAN 損傷",
			rule.Name, rule.Punishment.SANDamage)
	}

	return fmt.Sprintf("違反了「%s」規則", rule.Name)
}

// GetCluesForBeat returns clues that should be revealed at the current beat.
// Story 7.2 AC5: Implements tiered clue revelation system.
//
// Parameters:
//   - rules: Slice of hidden rules
//   - currentBeat: The current story beat (1-based)
//   - difficulty: The game difficulty level
//
// Returns:
//   - Slice of ClueToReveal for clues that should be shown now
//
// Difficulty constraints (AC5):
//   - Easy: Can reveal up to Tier 3 (near-truth)
//   - Hard: Can reveal up to Tier 2 (specific)
//   - Hell: Can reveal up to Tier 1 (vague)
func (e *RuleEngine) GetCluesForBeat(rules []*HiddenRule, currentBeat int, difficulty game.DifficultyLevel) []ClueToReveal {
	clues := []ClueToReveal{}

	// Determine max tier based on difficulty (AC5)
	maxTier := e.getMaxTierForDifficulty(difficulty)

	for _, rule := range rules {
		// Skip violated rules - players already know about them
		if rule.IsViolated {
			continue
		}

		for i := range rule.ClueHints {
			hint := &rule.ClueHints[i]

			// Check if this clue should be revealed
			if e.shouldRevealClue(hint, currentBeat, maxTier) {
				clues = append(clues, ClueToReveal{
					RuleID:   rule.ID,
					RuleName: rule.Name,
					Tier:     hint.Tier,
					Hint:     hint.Hint,
				})

				// Mark as revealed (AC5)
				hint.Revealed = true
				rule.TimesHinted++
			}
		}
	}

	return clues
}

// getMaxTierForDifficulty returns the maximum clue tier for a difficulty level.
// Story 7.2 AC5: Difficulty-based clue revelation limits.
func (e *RuleEngine) getMaxTierForDifficulty(difficulty game.DifficultyLevel) int {
	switch difficulty {
	case game.DifficultyEasy:
		return 3 // Can reveal all tiers (near-truth)
	case game.DifficultyHard:
		return 2 // Only up to specific clues
	case game.DifficultyHell:
		return 1 // Only vague clues
	default:
		return 2
	}
}

// shouldRevealClue checks if a clue should be revealed now.
// Story 7.2 AC5: Beat-based clue revelation logic.
func (e *RuleEngine) shouldRevealClue(hint *ClueHint, currentBeat int, maxTier int) bool {
	// Already revealed
	if hint.Revealed {
		return false
	}

	// Tier exceeds difficulty limit
	if hint.Tier > maxTier {
		return false
	}

	// Check if current beat is within the clue's beat range
	if currentBeat >= hint.BeatRange[0] && currentBeat <= hint.BeatRange[1] {
		return true
	}

	return false
}

// ValidateRulePlayability checks if a set of rules is playable.
// Story 7.2 AC6: Ensures rules don't create "no-win" scenarios.
//
// Returns an error if the rule combination is unplayable.
func (e *RuleEngine) ValidateRulePlayability(rules []*HiddenRule) error {
	// Check for contradictory rules
	// This is a basic check - more sophisticated logic could be added

	// Ensure not all rules are fatal with no warnings
	allFatal := true
	hasEscapeRoute := false

	for _, rule := range rules {
		if !rule.Punishment.IsFatal {
			allFatal = false
		}

		// If there are clues, there's a chance to learn and avoid
		if len(rule.ClueHints) > 0 {
			hasEscapeRoute = true
		}
	}

	// If all rules are fatal and there are no clues, it's unplayable
	if allFatal && !hasEscapeRoute && len(rules) > 0 {
		return fmt.Errorf("rule set is unplayable: all rules are fatal with no clues")
	}

	return nil
}

// GetRulesByCategory returns rules filtered by category.
func (e *RuleEngine) GetRulesByCategory(rules []*HiddenRule, category string) []*HiddenRule {
	filtered := []*HiddenRule{}
	categoryLower := strings.ToLower(category)

	for _, rule := range rules {
		if strings.ToLower(rule.Category) == categoryLower {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}

// GetActiveRules returns rules that haven't been violated yet.
func (e *RuleEngine) GetActiveRules(rules []*HiddenRule) []*HiddenRule {
	active := []*HiddenRule{}

	for _, rule := range rules {
		if !rule.IsViolated {
			active = append(active, rule)
		}
	}

	return active
}

// GetViolatedRules returns rules that have been violated.
func (e *RuleEngine) GetViolatedRules(rules []*HiddenRule) []*HiddenRule {
	violated := []*HiddenRule{}

	for _, rule := range rules {
		if rule.IsViolated {
			violated = append(violated, rule)
		}
	}

	return violated
}

// CountRevealedClues counts how many clues have been revealed for a rule.
func (e *RuleEngine) CountRevealedClues(rule *HiddenRule) int {
	count := 0
	for _, hint := range rule.ClueHints {
		if hint.Revealed {
			count++
		}
	}
	return count
}

// GetRevealedClues returns all revealed clues for a rule.
func (e *RuleEngine) GetRevealedClues(rule *HiddenRule) []string {
	clues := []string{}
	for _, hint := range rule.ClueHints {
		if hint.Revealed {
			clues = append(clues, hint.Hint)
		}
	}
	return clues
}

// ApplyDifficultyAdjustment applies a difficulty modifier to rule damage.
// Story 9.8: Integration with Guardian difficulty tuning system.
//
// This method reduces rule violation damage based on the Guardian's difficulty adjustment.
// The damageReduction parameter should come from DifficultyTuner.GetModifiers().DamageReduction.
//
// Parameters:
//   - violation: The rule violation to adjust
//   - damageReduction: The damage reduction percentage (0.0 to 0.3 = 0% to 30% reduction)
//
// Returns:
//   - An adjusted RuleViolation with reduced damage
func (e *RuleEngine) ApplyDifficultyAdjustment(violation RuleViolation, damageReduction float64) RuleViolation {
	if damageReduction <= 0 {
		return violation
	}

	adjustedViolation := violation

	// Apply damage reduction to HP damage
	if violation.HPDamage > 0 {
		reduction := float64(violation.HPDamage) * damageReduction
		adjustedViolation.HPDamage = int(float64(violation.HPDamage) - reduction)
		if adjustedViolation.HPDamage < 0 {
			adjustedViolation.HPDamage = 0
		}
	}

	// Apply damage reduction to SAN damage
	if violation.SANDamage > 0 {
		reduction := float64(violation.SANDamage) * damageReduction
		adjustedViolation.SANDamage = int(float64(violation.SANDamage) - reduction)
		if adjustedViolation.SANDamage < 0 {
			adjustedViolation.SANDamage = 0
		}
	}

	// Note: Fatal violations remain fatal (safety mechanism - don't trivialize death)
	// Guardian should prevent fatal violations through other means

	return adjustedViolation
}

// AdjustPunishmentDamage adjusts a HiddenRulePunishment based on difficulty modifier.
// Story 9.8: Integration helper for applying difficulty adjustments before violations occur.
//
// This can be used to preemptively adjust rule damage when generating or loading rules.
//
// Parameters:
//   - punishment: The punishment to adjust
//   - damageReduction: The damage reduction percentage (0.0 to 0.3 = 0% to 30% reduction)
//
// Returns:
//   - An adjusted HiddenRulePunishment with reduced damage
func (e *RuleEngine) AdjustPunishmentDamage(punishment HiddenRulePunishment, damageReduction float64) HiddenRulePunishment {
	if damageReduction <= 0 {
		return punishment
	}

	adjustedPunishment := punishment

	// Apply damage reduction to HP damage
	if punishment.HPDamage > 0 {
		reduction := float64(punishment.HPDamage) * damageReduction
		adjustedPunishment.HPDamage = int(float64(punishment.HPDamage) - reduction)
		if adjustedPunishment.HPDamage < 0 {
			adjustedPunishment.HPDamage = 0
		}
	}

	// Apply damage reduction to SAN damage
	if punishment.SANDamage > 0 {
		reduction := float64(punishment.SANDamage) * damageReduction
		adjustedPunishment.SANDamage = int(float64(punishment.SANDamage) - reduction)
		if adjustedPunishment.SANDamage < 0 {
			adjustedPunishment.SANDamage = 0
		}
	}

	// Fatal punishments remain fatal (safety mechanism)

	return adjustedPunishment
}
