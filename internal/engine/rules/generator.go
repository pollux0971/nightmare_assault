package rules

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// Generator creates hidden rules for a game session.
type Generator struct {
	// seed is used for deterministic rule generation when needed
	seed int64
}

// NewGenerator creates a new rule generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateRules creates a set of hidden rules based on difficulty.
// Per AC1: Easy mode has max 6 rules, Hard/Hell have unlimited (typically 8-15).
func (g *Generator) GenerateRules(difficulty game.DifficultyLevel) *RuleSet {
	rs := NewRuleSet()

	// Determine rule count based on difficulty (AC1)
	ruleCount := g.getRuleCount(difficulty)

	// Generate rules with type distribution (AC2)
	typeDistribution := g.calculateTypeDistribution(ruleCount)

	ruleIndex := 0
	for ruleType, count := range typeDistribution {
		for i := 0; i < count; i++ {
			rule := g.generateRule(ruleIndex, ruleType, difficulty)
			rs.Add(rule)
			ruleIndex++
		}
	}

	// Assign priorities to avoid conflicts
	g.assignPriorities(rs)

	return rs
}

// getRuleCount returns the number of rules to generate based on difficulty.
func (g *Generator) getRuleCount(difficulty game.DifficultyLevel) int {
	switch difficulty {
	case game.DifficultyEasy:
		// AC1: Easy mode has max 6 rules
		return randomInt(4, 6)
	case game.DifficultyHard:
		// Standard mode: 8-12 rules
		return randomInt(8, 12)
	case game.DifficultyHell:
		// Hell mode: 10-15 rules
		return randomInt(10, 15)
	default:
		return 8
	}
}

// calculateTypeDistribution ensures even distribution of rule types (AC2).
// Returns a map of RuleType -> count.
func (g *Generator) calculateTypeDistribution(totalRules int) map[RuleType]int {
	distribution := make(map[RuleType]int)
	types := AllRuleTypes()
	numTypes := len(types)

	// Base count for each type
	baseCount := totalRules / numTypes
	remainder := totalRules % numTypes

	for i, t := range types {
		count := baseCount
		// Distribute remainder across first few types
		if i < remainder {
			count++
		}
		if count > 0 {
			distribution[t] = count
		}
	}

	return distribution
}

// generateRule creates a single rule of the specified type.
func (g *Generator) generateRule(index int, ruleType RuleType, difficulty game.DifficultyLevel) *Rule {
	id := fmt.Sprintf("rule-%d", index+1)
	rule := NewRule(id, ruleType)

	// Set trigger based on type
	rule.Trigger = g.generateTrigger(ruleType)

	// Set consequence based on difficulty
	rule.Consequence = g.generateConsequence(ruleType, difficulty)

	// Generate clues (2-4 per rule)
	rule.Clues = g.generateClues(ruleType, randomInt(2, 4))

	// Set warning text
	rule.WarningText = g.generateWarningText(ruleType)

	// Max violations: Easy gets more chances, Hell is immediate
	switch difficulty {
	case game.DifficultyEasy:
		rule.MaxViolations = 2
	case game.DifficultyHard:
		rule.MaxViolations = 1
	case game.DifficultyHell:
		rule.MaxViolations = 0 // Immediate consequence
	}

	return rule
}

// generateTrigger creates a trigger condition based on rule type.
func (g *Generator) generateTrigger(ruleType RuleType) Condition {
	switch ruleType {
	case RuleTypeScenario:
		locations := []string{
			"basement", "attic", "bathroom", "garden",
			"bedroom", "kitchen", "hallway", "garage",
		}
		return Condition{
			Type:     "location",
			Value:    locations[randomInt(0, len(locations)-1)],
			Operator: "equals",
		}

	case RuleTypeTime:
		times := []string{
			"midnight", "3am", "dusk", "dawn", "noon",
		}
		return Condition{
			Type:     "time",
			Value:    times[randomInt(0, len(times)-1)],
			Operator: "equals",
		}

	case RuleTypeBehavior:
		behaviors := []string{
			"run", "shout", "open_door", "look_back",
			"ignore_warning", "touch_object", "read_text",
		}
		return Condition{
			Type:     "action",
			Value:    behaviors[randomInt(0, len(behaviors)-1)],
			Operator: "equals",
		}

	case RuleTypeObject:
		objects := []string{
			"mirror", "photograph", "doll", "clock",
			"phone", "book", "key", "candle",
		}
		return Condition{
			Type:     "object_interaction",
			Value:    objects[randomInt(0, len(objects)-1)],
			Operator: "equals",
		}

	case RuleTypeStatus:
		// Status rules trigger based on HP/SAN thresholds
		return Condition{
			Type:      "san_below",
			Value:     "san",
			Operator:  "less_than",
			Threshold: randomInt(30, 50),
		}

	default:
		return Condition{Type: "unknown"}
	}
}

// generateConsequence creates consequence based on difficulty.
func (g *Generator) generateConsequence(ruleType RuleType, difficulty game.DifficultyLevel) Outcome {
	// Base consequence type - harder difficulties have more instant deaths
	var conseqType ConsequenceType

	roll := randomInt(1, 100)
	switch difficulty {
	case game.DifficultyEasy:
		// 70% warning, 30% damage, 0% instant death
		if roll <= 70 {
			conseqType = ConsequenceWarning
		} else {
			conseqType = ConsequenceDamage
		}

	case game.DifficultyHard:
		// 40% warning, 50% damage, 10% instant death
		if roll <= 40 {
			conseqType = ConsequenceWarning
		} else if roll <= 90 {
			conseqType = ConsequenceDamage
		} else {
			conseqType = ConsequenceInstantDeath
		}

	case game.DifficultyHell:
		// 20% warning, 50% damage, 30% instant death
		if roll <= 20 {
			conseqType = ConsequenceWarning
		} else if roll <= 70 {
			conseqType = ConsequenceDamage
		} else {
			conseqType = ConsequenceInstantDeath
		}
	}

	outcome := Outcome{Type: conseqType}

	// Set damage amounts
	if conseqType == ConsequenceDamage {
		switch ruleType {
		case RuleTypeScenario:
			outcome.HPDamage = randomInt(10, 25)
		case RuleTypeTime:
			outcome.SANDamage = randomInt(15, 30)
		case RuleTypeBehavior:
			outcome.HPDamage = randomInt(10, 20)
			outcome.SANDamage = randomInt(5, 15)
		case RuleTypeObject:
			outcome.SANDamage = randomInt(10, 25)
		case RuleTypeStatus:
			outcome.HPDamage = randomInt(15, 30)
		}

		// Apply difficulty multipliers
		outcome.HPDamage = int(float64(outcome.HPDamage) * difficulty.HPDrainMultiplier())
		outcome.SANDamage = int(float64(outcome.SANDamage) * difficulty.SANDrainMultiplier())
	}

	return outcome
}

// generateClues creates hint clues for a rule.
func (g *Generator) generateClues(ruleType RuleType, count int) []string {
	clueTemplates := map[RuleType][]string{
		RuleTypeScenario: {
			"那個地方的空氣似乎格外沉重",
			"地板上有奇怪的刮痕",
			"牆上的影子好像在移動",
			"有股腐爛的味道飄來",
			"這裡曾經發生過什麼",
		},
		RuleTypeTime: {
			"時鐘停在一個奇怪的時間",
			"窗外的天色異常",
			"蠟燭的火焰開始搖晃",
			"遠處傳來鐘聲",
			"月亮的位置不太對勁",
		},
		RuleTypeBehavior: {
			"直覺告訴你不該這麼做",
			"某個聲音好像在警告你",
			"身體莫名地僵硬",
			"心跳突然加速",
			"冷汗沿著脊椎流下",
		},
		RuleTypeObject: {
			"這個物品散發著詭異的氣息",
			"表面有不知名的污漬",
			"觸碰時感到一陣刺痛",
			"好像有人在盯著你",
			"空氣中飄著血腥味",
		},
		RuleTypeStatus: {
			"視野開始模糊",
			"耳邊響起低語聲",
			"身體越來越虛弱",
			"現實好像在扭曲",
			"記憶變得不可靠",
		},
	}

	templates := clueTemplates[ruleType]
	if templates == nil {
		templates = clueTemplates[RuleTypeBehavior] // fallback
	}

	clues := make([]string, 0, count)
	used := make(map[int]bool)

	for len(clues) < count && len(used) < len(templates) {
		idx := randomInt(0, len(templates)-1)
		if !used[idx] {
			used[idx] = true
			clues = append(clues, templates[idx])
		}
	}

	return clues
}

// generateWarningText creates a warning message for a rule.
func (g *Generator) generateWarningText(ruleType RuleType) string {
	warnings := map[RuleType][]string{
		RuleTypeScenario: {
			"這個地方讓人不安...",
			"空氣中瀰漫著危險的氣息",
			"你的直覺在尖叫",
		},
		RuleTypeTime: {
			"時間好像變得奇怪了",
			"現在不是做這件事的時候",
			"夜晚隱藏著秘密",
		},
		RuleTypeBehavior: {
			"你確定要這麼做嗎？",
			"某種力量似乎在阻止你",
			"這個動作可能會引來麻煩",
		},
		RuleTypeObject: {
			"最好不要碰這個東西",
			"它好像在注視著你",
			"上面的痕跡令人不安",
		},
		RuleTypeStatus: {
			"你的狀態很糟糕",
			"必須小心行事",
			"理智正在流失",
		},
	}

	texts := warnings[ruleType]
	if texts == nil {
		return "有什麼不對勁..."
	}
	return texts[randomInt(0, len(texts)-1)]
}

// assignPriorities assigns priorities to rules to handle conflicts.
// Higher priority rules take precedence.
func (g *Generator) assignPriorities(rs *RuleSet) {
	// Instant death rules get highest priority
	// Status rules get second priority (player state dependent)
	// Others get medium priority
	for _, r := range rs.Rules {
		basePriority := 5

		// Boost priority for instant death
		if r.Consequence.Type == ConsequenceInstantDeath {
			basePriority = 10
		}

		// Status rules have higher priority
		if r.Type == RuleTypeStatus {
			basePriority = 8
		}

		// Add some randomness within range
		r.Priority = basePriority + randomInt(-1, 1)
	}
}

// randomInt returns a random integer in [min, max].
func randomInt(min, max int) int {
	if max <= min {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}
