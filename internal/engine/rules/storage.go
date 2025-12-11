package rules

import (
	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
)

// ToSaveStorage converts a RuleSet to save-compatible storage format.
// This is used when saving game state (AC3: rules stored in game state).
func (rs *RuleSet) ToSaveStorage() *save.RuleStorage {
	if rs == nil || len(rs.Rules) == 0 {
		return nil
	}

	storage := &save.RuleStorage{
		Rules: make([]save.SavedRule, len(rs.Rules)),
	}

	for i, r := range rs.Rules {
		storage.Rules[i] = ruleToSavedRule(r)
	}

	return storage
}

// FromSaveStorage restores a RuleSet from save data.
// This is used when loading a saved game.
func FromSaveStorage(storage *save.RuleStorage) *RuleSet {
	if storage == nil || len(storage.Rules) == 0 {
		return NewRuleSet()
	}

	rs := NewRuleSet()
	for _, sr := range storage.Rules {
		rs.Add(savedRuleToRule(&sr))
	}

	return rs
}

// ruleToSavedRule converts a Rule to SavedRule format.
func ruleToSavedRule(r *Rule) save.SavedRule {
	var conseqStr string
	switch r.Consequence.Type {
	case ConsequenceWarning:
		conseqStr = "warning"
	case ConsequenceDamage:
		conseqStr = "damage"
	case ConsequenceInstantDeath:
		conseqStr = "instant_death"
	}

	return save.SavedRule{
		ID:            r.ID,
		Type:          r.Type.EnglishName(),
		TriggerType:   r.Trigger.Type,
		TriggerValue:  r.Trigger.Value,
		Consequence:   conseqStr,
		HPDamage:      r.Consequence.HPDamage,
		SANDamage:     r.Consequence.SANDamage,
		Clues:         r.Clues,
		Priority:      r.Priority,
		Violations:    r.Violations,
		MaxViolations: r.MaxViolations,
		Discovered:    r.Discovered,
		Active:        r.Active,
	}
}

// savedRuleToRule converts a SavedRule back to Rule format.
func savedRuleToRule(sr *save.SavedRule) *Rule {
	r := NewRule(sr.ID, ParseRuleType(sr.Type))

	r.Trigger = Condition{
		Type:  sr.TriggerType,
		Value: sr.TriggerValue,
	}

	var conseqType ConsequenceType
	switch sr.Consequence {
	case "warning":
		conseqType = ConsequenceWarning
	case "damage":
		conseqType = ConsequenceDamage
	case "instant_death":
		conseqType = ConsequenceInstantDeath
	default:
		conseqType = ConsequenceWarning
	}

	r.Consequence = Outcome{
		Type:      conseqType,
		HPDamage:  sr.HPDamage,
		SANDamage: sr.SANDamage,
	}

	r.Clues = sr.Clues
	r.Priority = sr.Priority
	r.Violations = sr.Violations
	r.MaxViolations = sr.MaxViolations
	r.Discovered = sr.Discovered
	r.Active = sr.Active

	return r
}
