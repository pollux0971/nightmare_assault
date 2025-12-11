package rules

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// PlayerAction represents an action taken by the player.
type PlayerAction struct {
	Type      string    // action type: "move", "interact", "speak", "look", etc.
	Target    string    // target of the action: location, object, NPC name
	Details   string    // additional action details
	Timestamp time.Time // when the action was taken
}

// GameContext provides game state context for rule checking.
type GameContext struct {
	Location    string              // current location name
	TimeOfDay   string              // game time: "dawn", "noon", "dusk", "midnight", "3am"
	Chapter     int                 // current chapter
	HP          int                 // player HP
	SAN         int                 // player SAN
	Difficulty  game.DifficultyLevel // game difficulty
}

// TriggerResult represents the outcome of a rule trigger.
type TriggerResult struct {
	Triggered   bool
	Rule        *Rule
	IsWarning   bool
	IsFatal     bool
	HPDamage    int
	SANDamage   int
	WarningText string
	// For conflicting rules
	ConflictingRules []*Rule
}

// TriggerRecord logs a rule trigger for the debrief system.
type TriggerRecord struct {
	RuleID      string    `json:"rule_id"`
	RuleType    RuleType  `json:"rule_type"`
	Action      string    `json:"action"`
	Timestamp   time.Time `json:"timestamp"`
	Chapter     int       `json:"chapter"`
	Consequence string    `json:"consequence"` // "warning", "damage", "instant_death"
	HPDamage    int       `json:"hp_damage,omitempty"`
	SANDamage   int       `json:"san_damage,omitempty"`
	WarningText string    `json:"warning_text,omitempty"`
}

// CheckResult contains all results from a rule check.
type CheckResult struct {
	Triggers      []TriggerResult
	Records       []TriggerRecord
	AnyFatal      bool
	TotalHPDamage int
	TotalSANDamage int
}

// Checker validates player actions against active rules.
type Checker struct {
	ruleSet    *RuleSet
	difficulty game.DifficultyLevel
	records    []TriggerRecord
	mu         sync.RWMutex
}

// NewChecker creates a new rule checker.
func NewChecker(ruleSet *RuleSet, difficulty game.DifficultyLevel) *Checker {
	return &Checker{
		ruleSet:    ruleSet,
		difficulty: difficulty,
		records:    make([]TriggerRecord, 0),
	}
}

// SetRuleSet updates the active rule set.
func (c *Checker) SetRuleSet(ruleSet *RuleSet) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ruleSet = ruleSet
}

// SetDifficulty updates the difficulty level.
func (c *Checker) SetDifficulty(difficulty game.DifficultyLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.difficulty = difficulty
}

// Check evaluates all rules against a player action.
// Per AC1: Checks rules in order: Scenario → Time → Behavior → Status
func (c *Checker) Check(action PlayerAction, ctx GameContext) CheckResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := CheckResult{
		Triggers: make([]TriggerResult, 0),
		Records:  make([]TriggerRecord, 0),
	}

	if c.ruleSet == nil || c.ruleSet.Count() == 0 {
		return result
	}

	// Check rules in order (AC1)
	checkOrder := []RuleType{
		RuleTypeScenario,
		RuleTypeTime,
		RuleTypeBehavior,
		RuleTypeObject,
		RuleTypeStatus,
	}

	for _, ruleType := range checkOrder {
		rules := c.ruleSet.GetByType(ruleType)
		// Sort by priority (higher first)
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Priority > rules[j].Priority
		})

		for _, rule := range rules {
			if !rule.Active {
				continue
			}

			if c.matchesTrigger(rule, action, ctx) {
				triggerResult := c.processTriggeredRule(rule, action, ctx)
				result.Triggers = append(result.Triggers, triggerResult)

				// Create record for debrief (AC5)
				record := c.createTriggerRecord(rule, action, ctx, triggerResult)
				result.Records = append(result.Records, record)
				c.records = append(c.records, record)

				// Accumulate damage
				result.TotalHPDamage += triggerResult.HPDamage
				result.TotalSANDamage += triggerResult.SANDamage

				if triggerResult.IsFatal {
					result.AnyFatal = true
				}
			}
		}
	}

	// Check for conflicting rules (AC4)
	result = c.detectConflicts(result)

	return result
}

// matchesTrigger checks if an action matches a rule's trigger condition.
func (c *Checker) matchesTrigger(rule *Rule, action PlayerAction, ctx GameContext) bool {
	trigger := rule.Trigger

	switch rule.Type {
	case RuleTypeScenario:
		// Check if player is in the triggering location
		return c.matchesCondition(ctx.Location, trigger)

	case RuleTypeTime:
		// Check if current time matches
		return c.matchesCondition(ctx.TimeOfDay, trigger)

	case RuleTypeBehavior:
		// Check if action type matches
		return c.matchesCondition(action.Type, trigger) ||
			c.matchesCondition(action.Details, trigger)

	case RuleTypeObject:
		// Check if interacting with triggering object
		return c.matchesCondition(action.Target, trigger)

	case RuleTypeStatus:
		// Check HP/SAN thresholds
		return c.matchesStatusCondition(ctx, trigger)

	default:
		return false
	}
}

// matchesCondition checks if a value matches a trigger condition.
func (c *Checker) matchesCondition(value string, trigger Condition) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	triggerValue := strings.ToLower(strings.TrimSpace(trigger.Value))

	switch trigger.Operator {
	case "equals", "":
		return value == triggerValue
	case "contains":
		return strings.Contains(value, triggerValue)
	case "not_equals":
		return value != triggerValue
	case "starts_with":
		return strings.HasPrefix(value, triggerValue)
	default:
		return value == triggerValue
	}
}

// matchesStatusCondition checks HP/SAN threshold conditions.
func (c *Checker) matchesStatusCondition(ctx GameContext, trigger Condition) bool {
	var value int
	switch trigger.Type {
	case "hp_below", "hp":
		value = ctx.HP
	case "san_below", "san":
		value = ctx.SAN
	default:
		return false
	}

	switch trigger.Operator {
	case "less_than", "":
		return value < trigger.Threshold
	case "less_than_or_equal":
		return value <= trigger.Threshold
	case "greater_than":
		return value > trigger.Threshold
	case "equals":
		return value == trigger.Threshold
	default:
		return value < trigger.Threshold
	}
}

// processTriggeredRule handles a triggered rule and determines the outcome.
func (c *Checker) processTriggeredRule(rule *Rule, action PlayerAction, ctx GameContext) TriggerResult {
	result := TriggerResult{
		Triggered: true,
		Rule:      rule,
	}

	// Check if instant death (AC3)
	if rule.Consequence.Type == ConsequenceInstantDeath {
		result.IsFatal = true
		result.WarningText = rule.WarningText
		return result
	}

	// Process violation (AC2)
	exceededMax := rule.RecordViolation()

	if exceededMax {
		// Violations exceeded - apply full consequence
		if rule.Consequence.Type == ConsequenceDamage {
			result.HPDamage = rule.Consequence.HPDamage
			result.SANDamage = rule.Consequence.SANDamage
		}
		// After max violations on non-instant rules, check if should be fatal
		// (This depends on game design - for now, damage rules stay as damage)
	} else {
		// Still within warning threshold
		result.IsWarning = true
		result.WarningText = rule.WarningText
		// Apply reduced damage for warnings
		if rule.Consequence.Type == ConsequenceDamage {
			result.HPDamage = rule.Consequence.HPDamage / 2
			result.SANDamage = rule.Consequence.SANDamage / 2
		}
	}

	return result
}

// createTriggerRecord creates a record for the debrief system.
func (c *Checker) createTriggerRecord(rule *Rule, action PlayerAction, ctx GameContext, result TriggerResult) TriggerRecord {
	var consequence string
	if result.IsFatal {
		consequence = "instant_death"
	} else if result.IsWarning {
		consequence = "warning"
	} else {
		consequence = "damage"
	}

	return TriggerRecord{
		RuleID:      rule.ID,
		RuleType:    rule.Type,
		Action:      action.Type + ": " + action.Target,
		Timestamp:   action.Timestamp,
		Chapter:     ctx.Chapter,
		Consequence: consequence,
		HPDamage:    result.HPDamage,
		SANDamage:   result.SANDamage,
		WarningText: result.WarningText,
	}
}

// detectConflicts identifies conflicting rules in the results (AC4).
func (c *Checker) detectConflicts(result CheckResult) CheckResult {
	if len(result.Triggers) < 2 {
		return result
	}

	// Group triggers by type to find potential conflicts
	for i := 0; i < len(result.Triggers); i++ {
		for j := i + 1; j < len(result.Triggers); j++ {
			if c.areConflicting(result.Triggers[i].Rule, result.Triggers[j].Rule) {
				// Add to conflicting rules list
				if result.Triggers[i].ConflictingRules == nil {
					result.Triggers[i].ConflictingRules = make([]*Rule, 0)
				}
				result.Triggers[i].ConflictingRules = append(
					result.Triggers[i].ConflictingRules,
					result.Triggers[j].Rule,
				)
			}
		}
	}

	return result
}

// areConflicting determines if two rules conflict.
func (c *Checker) areConflicting(r1, r2 *Rule) bool {
	// Rules of the same type with different triggers that both fired
	// might be conflicting (player can't satisfy both)
	if r1.Type == r2.Type && r1.Trigger.Value != r2.Trigger.Value {
		// If consequences are opposite (one says do, one says don't)
		// This is a simplified check - real implementation would need
		// semantic analysis of trigger conditions
		return true
	}
	return false
}

// GetTriggerHistory returns all recorded triggers (for debrief system).
func (c *Checker) GetTriggerHistory() []TriggerRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history := make([]TriggerRecord, len(c.records))
	copy(history, c.records)
	return history
}

// GetTriggersByRuleID returns triggers for a specific rule.
func (c *Checker) GetTriggersByRuleID(ruleID string) []TriggerRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []TriggerRecord
	for _, r := range c.records {
		if r.RuleID == ruleID {
			result = append(result, r)
		}
	}
	return result
}

// ClearHistory clears the trigger history (for new game).
func (c *Checker) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = make([]TriggerRecord, 0)
}

// GetFatalTriggers returns all instant death triggers.
func (c *Checker) GetFatalTriggers() []TriggerRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []TriggerRecord
	for _, r := range c.records {
		if r.Consequence == "instant_death" {
			result = append(result, r)
		}
	}
	return result
}
