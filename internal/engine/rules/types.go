// Package rules provides hidden rule management for Nightmare Assault.
package rules

import (
	"encoding/json"
)

// RuleType represents the category of a hidden rule.
type RuleType int

const (
	// RuleTypeScenario represents rules that apply in specific locations.
	RuleTypeScenario RuleType = iota
	// RuleTypeTime represents rules that apply during specific time periods.
	RuleTypeTime
	// RuleTypeBehavior represents rules that apply to specific player actions.
	RuleTypeBehavior
	// RuleTypeObject represents rules targeting specific NPCs or items.
	RuleTypeObject
	// RuleTypeStatus represents rules dependent on player state (HP, SAN).
	RuleTypeStatus
)

// String returns the Chinese display name of the rule type.
func (t RuleType) String() string {
	switch t {
	case RuleTypeScenario:
		return "場景規則"
	case RuleTypeTime:
		return "時間規則"
	case RuleTypeBehavior:
		return "行為規則"
	case RuleTypeObject:
		return "對象規則"
	case RuleTypeStatus:
		return "狀態規則"
	default:
		return "未知規則"
	}
}

// EnglishName returns the English name for prompt generation.
func (t RuleType) EnglishName() string {
	switch t {
	case RuleTypeScenario:
		return "Scenario"
	case RuleTypeTime:
		return "Time"
	case RuleTypeBehavior:
		return "Behavior"
	case RuleTypeObject:
		return "Object"
	case RuleTypeStatus:
		return "Status"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements json.Marshaler.
func (t RuleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.EnglishName())
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *RuleType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = ParseRuleType(s)
	return nil
}

// ParseRuleType parses a string into a RuleType.
func ParseRuleType(s string) RuleType {
	switch s {
	case "Scenario", "場景規則", "scenario":
		return RuleTypeScenario
	case "Time", "時間規則", "time":
		return RuleTypeTime
	case "Behavior", "行為規則", "behavior":
		return RuleTypeBehavior
	case "Object", "對象規則", "object":
		return RuleTypeObject
	case "Status", "狀態規則", "status":
		return RuleTypeStatus
	default:
		return RuleTypeBehavior // default to behavior
	}
}

// AllRuleTypes returns all available rule types.
func AllRuleTypes() []RuleType {
	return []RuleType{
		RuleTypeScenario,
		RuleTypeTime,
		RuleTypeBehavior,
		RuleTypeObject,
		RuleTypeStatus,
	}
}

// ConsequenceType represents the severity of breaking a rule.
type ConsequenceType int

const (
	// ConsequenceWarning gives a warning but no damage.
	ConsequenceWarning ConsequenceType = iota
	// ConsequenceDamage causes HP/SAN damage.
	ConsequenceDamage
	// ConsequenceInstantDeath causes immediate death.
	ConsequenceInstantDeath
)

// String returns the display name.
func (c ConsequenceType) String() string {
	switch c {
	case ConsequenceWarning:
		return "警告"
	case ConsequenceDamage:
		return "傷害"
	case ConsequenceInstantDeath:
		return "即死"
	default:
		return "未知"
	}
}

// MarshalJSON implements json.Marshaler.
func (c ConsequenceType) MarshalJSON() ([]byte, error) {
	switch c {
	case ConsequenceWarning:
		return json.Marshal("warning")
	case ConsequenceDamage:
		return json.Marshal("damage")
	case ConsequenceInstantDeath:
		return json.Marshal("instant_death")
	default:
		return json.Marshal("warning")
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ConsequenceType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "warning":
		*c = ConsequenceWarning
	case "damage":
		*c = ConsequenceDamage
	case "instant_death":
		*c = ConsequenceInstantDeath
	default:
		*c = ConsequenceWarning
	}
	return nil
}

// Condition represents a trigger condition for a rule.
type Condition struct {
	// Type of condition (location, time, action, etc.)
	Type string `json:"type"`
	// Value is the specific trigger value
	Value string `json:"value"`
	// Operator determines how to match (equals, contains, not_equals)
	Operator string `json:"operator,omitempty"`
	// Threshold for numeric conditions (HP < 50, etc.)
	Threshold int `json:"threshold,omitempty"`
}

// Outcome represents the consequence of triggering a rule.
type Outcome struct {
	// Type of consequence
	Type ConsequenceType `json:"type"`
	// HPDamage is the HP damage amount (if applicable)
	HPDamage int `json:"hp_damage,omitempty"`
	// SANDamage is the SAN damage amount (if applicable)
	SANDamage int `json:"san_damage,omitempty"`
	// Description is the narrative description of the consequence
	Description string `json:"description,omitempty"`
}

// Rule represents a single hidden rule in the game.
type Rule struct {
	// ID is the unique identifier for this rule
	ID string `json:"id"`
	// Type categorizes the rule (scenario, time, behavior, object, status)
	Type RuleType `json:"type"`
	// Trigger defines when this rule activates
	Trigger Condition `json:"trigger"`
	// Consequence defines what happens when triggered
	Consequence Outcome `json:"consequence"`
	// Clues are hints that can be discovered about this rule
	Clues []string `json:"clues"`
	// Priority determines precedence when rules conflict (higher = more priority)
	Priority int `json:"priority"`
	// WarningText is shown as a subtle warning before full violation
	WarningText string `json:"warning_text,omitempty"`
	// Violations tracks how many times this rule has been violated
	Violations int `json:"violations"`
	// MaxViolations is the max violations before consequence (0 = immediate)
	MaxViolations int `json:"max_violations"`
	// Discovered indicates if player has learned about this rule
	Discovered bool `json:"discovered"`
	// Active indicates if this rule is currently in effect
	Active bool `json:"active"`
}

// NewRule creates a new rule with sensible defaults.
func NewRule(id string, ruleType RuleType) *Rule {
	return &Rule{
		ID:            id,
		Type:          ruleType,
		Trigger:       Condition{},
		Consequence:   Outcome{Type: ConsequenceWarning},
		Clues:         make([]string, 0),
		Priority:      5, // Default mid-priority
		MaxViolations: 1, // Default: 1 warning before consequence
		Active:        true,
	}
}

// CanTrigger checks if this rule can be triggered (active and not discovered for certain types).
func (r *Rule) CanTrigger() bool {
	return r.Active
}

// RecordViolation records a violation and returns true if max is exceeded.
func (r *Rule) RecordViolation() bool {
	r.Violations++
	return r.Violations > r.MaxViolations
}

// ResetViolations resets the violation counter.
func (r *Rule) ResetViolations() {
	r.Violations = 0
}

// ShouldWarn returns true if player should receive a warning (not yet at max).
func (r *Rule) ShouldWarn() bool {
	return r.Violations <= r.MaxViolations && r.WarningText != ""
}

// RuleSet represents a collection of rules for a game session.
type RuleSet struct {
	// Rules is the list of active rules
	Rules []*Rule `json:"rules"`
	// MaxRulesEasy is the max rules for easy difficulty
	MaxRulesEasy int `json:"-"`
}

// NewRuleSet creates an empty rule set.
func NewRuleSet() *RuleSet {
	return &RuleSet{
		Rules:        make([]*Rule, 0),
		MaxRulesEasy: 6, // Per AC1: Easy mode has max 6 rules
	}
}

// Add adds a rule to the set.
func (rs *RuleSet) Add(r *Rule) {
	rs.Rules = append(rs.Rules, r)
}

// GetByID finds a rule by ID.
func (rs *RuleSet) GetByID(id string) *Rule {
	for _, r := range rs.Rules {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// GetByType returns all rules of a specific type.
func (rs *RuleSet) GetByType(t RuleType) []*Rule {
	var result []*Rule
	for _, r := range rs.Rules {
		if r.Type == t {
			result = append(result, r)
		}
	}
	return result
}

// GetActive returns all active rules.
func (rs *RuleSet) GetActive() []*Rule {
	var result []*Rule
	for _, r := range rs.Rules {
		if r.Active {
			result = append(result, r)
		}
	}
	return result
}

// GetDiscovered returns all discovered rules.
func (rs *RuleSet) GetDiscovered() []*Rule {
	var result []*Rule
	for _, r := range rs.Rules {
		if r.Discovered {
			result = append(result, r)
		}
	}
	return result
}

// Count returns the total number of rules.
func (rs *RuleSet) Count() int {
	return len(rs.Rules)
}

// CountByType returns the count of each rule type.
func (rs *RuleSet) CountByType() map[RuleType]int {
	counts := make(map[RuleType]int)
	for _, r := range rs.Rules {
		counts[r.Type]++
	}
	return counts
}

// ToJSON serializes the rule set to JSON.
func (rs *RuleSet) ToJSON() ([]byte, error) {
	return json.MarshalIndent(rs, "", "  ")
}

// FromJSON deserializes a rule set from JSON.
func (rs *RuleSet) FromJSON(data []byte) error {
	return json.Unmarshal(data, rs)
}
