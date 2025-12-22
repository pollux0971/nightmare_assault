package rules

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
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

// DifficultyConfig defines the rule generation parameters for each difficulty level.
type DifficultyConfig struct {
	// MaxRules is the maximum number of rules (0 = unlimited)
	MaxRules int
	// MappingLayer is the complexity of logic chains
	MappingLayer MappingLayer
	// ClueClarity is how clear the clues are
	ClueClarity ClueClarity
	// SmokeScreenCount is the number of misleading clues per rule
	SmokeScreenCount int
	// InstantDeathAllowed indicates if instant death without warning is allowed
	InstantDeathAllowed bool
	// InstantDeathChance is the percentage chance (0-100) of instant death rules
	InstantDeathChance int
}

// GetDifficultyConfig returns the configuration for a given difficulty level.
//
// AC1 (Easy):
//   - Max 6 rules
//   - Single mapping (A→B)
//   - Direct clues
//   - No smoke screens
//   - No instant death
//
// AC2 (Hard):
//   - Unlimited rules
//   - Double mapping (A→B→C)
//   - Metaphor/broken clues
//   - Medium smoke screens (2-3 per rule)
//   - 10% instant death chance
//
// AC3 (Hell):
//   - Unlimited rules
//   - Triple+ mapping (A→B→C→D)
//   - Contradictory/misleading clues
//   - Heavy smoke screens (4-6 per rule)
//   - 30% instant death chance, no warning
func GetDifficultyConfig(difficulty game.DifficultyLevel) DifficultyConfig {
	switch difficulty {
	case game.DifficultyEasy:
		return DifficultyConfig{
			MaxRules:            6,
			MappingLayer:        MappingLayerSingle,
			ClueClarity:         ClueClarityDirect,
			SmokeScreenCount:    0,
			InstantDeathAllowed: false,
			InstantDeathChance:  0,
		}

	case game.DifficultyHard:
		return DifficultyConfig{
			MaxRules:            0, // unlimited
			MappingLayer:        MappingLayerDouble,
			ClueClarity:         ClueClarityMetaphor,
			SmokeScreenCount:    randomInt(2, 3),
			InstantDeathAllowed: true,
			InstantDeathChance:  10,
		}

	case game.DifficultyHell:
		return DifficultyConfig{
			MaxRules:            0, // unlimited
			MappingLayer:        MappingLayerTriple,
			ClueClarity:         ClueClarityContradictory,
			SmokeScreenCount:    randomInt(4, 6),
			InstantDeathAllowed: true,
			InstantDeathChance:  30,
		}

	default:
		// Default to Easy
		return GetDifficultyConfig(game.DifficultyEasy)
	}
}

// GenerateRules creates a set of hidden rules based on difficulty.
// Per AC1: Easy mode has max 6 rules, Hard/Hell have unlimited (typically 8-15).
func (g *Generator) GenerateRules(difficulty game.DifficultyLevel) *RuleSet {
	// Story 10-8 AC1: Log rule generation
	logger.Debug("Rule generation started", map[string]interface{}{
		"difficulty": difficulty,
	})

	rs := NewRuleSet()
	config := GetDifficultyConfig(difficulty)

	// Determine rule count based on difficulty (AC1)
	ruleCount := g.getRuleCount(difficulty)

	logger.Debug("Rule count determined", map[string]interface{}{
		"rule_count": ruleCount,
		"mapping_layer": config.MappingLayer,
		"clue_clarity": config.ClueClarity,
		"instant_death_allowed": config.InstantDeathAllowed,
	})

	// Generate rules with type distribution (AC2)
	typeDistribution := g.calculateTypeDistribution(ruleCount)

	ruleIndex := 0
	for ruleType, count := range typeDistribution {
		for i := 0; i < count; i++ {
			rule := g.generateRuleWithDifficulty(ruleIndex, ruleType, difficulty, config)
			rs.Add(rule)
			ruleIndex++
		}
	}

	// Assign priorities to avoid conflicts
	g.assignPriorities(rs)

	// Story 10-8 AC1: Log completed rules
	logger.Debug("Rule generation completed", map[string]interface{}{
		"total_rules": rs.Count(),
	})

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

// generateRuleWithDifficulty creates a difficulty-aware rule with logic chains and smoke screens.
//
// Story 7.2: This method extends generateRule with:
//   - Logic chains based on mapping layer (single/double/triple)
//   - Clue clarity (direct/metaphor/contradictory)
//   - Smoke screens (misleading clues)
//   - Instant death configuration for Hell mode
func (g *Generator) generateRuleWithDifficulty(
	index int,
	ruleType RuleType,
	difficulty game.DifficultyLevel,
	config DifficultyConfig,
) *Rule {
	// Start with base rule generation
	rule := g.generateRule(index, ruleType, difficulty)

	// Apply difficulty-aware enhancements
	rule.MappingLayer = config.MappingLayer
	rule.ClueClarity = config.ClueClarity

	// Generate logic chain based on mapping layer
	rule.LogicChain = g.generateLogicChain(ruleType, config.MappingLayer)

	// Separate true clues from all clues
	rule.TrueClues = rule.Clues

	// Transform clues based on clarity
	rule.Clues = g.transformCluesByClarity(rule.Clues, config.ClueClarity)

	// Generate smoke screens (misleading clues)
	if config.SmokeScreenCount > 0 {
		rule.SmokeScreens = g.generateSmokeScreens(ruleType, config.SmokeScreenCount)
		// Mix smoke screens into visible clues
		rule.Clues = append(rule.Clues, rule.SmokeScreens...)
		// Shuffle to hide which are real
		g.shuffleClues(rule.Clues)
	}

	// Configure instant death for Hell mode
	if config.InstantDeathAllowed {
		roll := randomInt(1, 100)
		if roll <= config.InstantDeathChance {
			rule.Consequence.Type = ConsequenceInstantDeath
			rule.InstantDeathOK = true
			rule.MaxViolations = 0 // No warnings for instant death
		}
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

// generateLogicChain creates a multi-step deduction chain based on mapping layer.
//
// Story 7.2: Implements logic chains for different mapping layers:
//   - Single: Direct mapping (A→B) - "Don't touch mirror"
//   - Double: Two-step (A→B→C) - "Mirror reflects truth" → "Truth is dangerous" → "Don't touch mirror"
//   - Triple: Three+ steps (A→B→C→D) - "Midnight" → "Mirror active" → "Reflections come alive" → "Don't look"
func (g *Generator) generateLogicChain(ruleType RuleType, layer MappingLayer) []string {
	// Logic chain templates by rule type
	chains := map[RuleType]map[MappingLayer][]string{
		RuleTypeScenario: {
			MappingLayerSingle: {"直接規則：避開這個場景"},
			MappingLayerDouble: {"這個地方有問題", "有問題的地方很危險", "避開這個場景"},
			MappingLayerTriple: {"空氣異常沉重", "沉重代表不祥", "不祥之地隱藏陷阱", "陷阱會致命", "避開這個場景"},
		},
		RuleTypeTime: {
			MappingLayerSingle: {"直接規則：特定時間不可行動"},
			MappingLayerDouble: {"這個時間很特殊", "特殊時間規則改變", "此時不可行動"},
			MappingLayerTriple: {"鐘聲響起", "鐘聲標誌界限", "界限模糊時它們會出現", "它們會狩獵", "此時不可行動"},
		},
		RuleTypeBehavior: {
			MappingLayerSingle: {"直接規則：禁止此行為"},
			MappingLayerDouble: {"這個動作會引起注意", "引起注意很危險", "禁止此行為"},
			MappingLayerTriple: {"聲音會傳播", "傳播會吸引它們", "它們會循聲而來", "被發現即死", "禁止此行為"},
		},
		RuleTypeObject: {
			MappingLayerSingle: {"直接規則：不要碰這個物品"},
			MappingLayerDouble: {"這個物品被詛咒", "詛咒會傳染", "不要碰這個物品"},
			MappingLayerTriple: {"物品有靈性", "靈性會甦醒", "甦醒後會佔據", "佔據者失去自我", "不要碰這個物品"},
		},
		RuleTypeStatus: {
			MappingLayerSingle: {"直接規則：維持狀態在安全值"},
			MappingLayerDouble: {"狀態低落會失控", "失控會帶來危險", "維持狀態在安全值"},
			MappingLayerTriple: {"理智下降", "下降使感知扭曲", "扭曲讓你看見真相", "真相會摧毀心智", "維持狀態在安全值"},
		},
	}

	typeChains, ok := chains[ruleType]
	if !ok {
		typeChains = chains[RuleTypeBehavior] // fallback
	}

	chain, ok := typeChains[layer]
	if !ok {
		chain = typeChains[MappingLayerSingle] // fallback
	}

	return chain
}

// transformCluesByClarity transforms clues based on clarity level.
//
// Story 7.2: Implements clue transformation:
//   - Direct: Clear, obvious hints
//   - Metaphor: Poetic, fragmented hints
//   - Contradictory: Misleading, opposite hints
func (g *Generator) transformCluesByClarity(clues []string, clarity ClueClarity) []string {
	transformed := make([]string, len(clues))

	for i, clue := range clues {
		switch clarity {
		case ClueClarityDirect:
			// Keep as is
			transformed[i] = clue

		case ClueClarityMetaphor:
			// Make metaphorical/fragmented
			metaphors := []string{
				"...好像%s...",
				"某種意義上，%s",
				"如果理解正確的話，%s",
				"隱約感覺到%s",
				"也許%s？不確定...",
			}
			template := metaphors[randomInt(0, len(metaphors)-1)]
			transformed[i] = fmt.Sprintf(template, clue)

		case ClueClarityContradictory:
			// Make contradictory/misleading
			contradictions := []string{
				"似乎很安全，但%s",
				"表面上%s，實則相反",
				"不是%s...還是說是？",
				"記錄顯示%s（記錄已損毀）",
				"絕對不是%s",
			}
			template := contradictions[randomInt(0, len(contradictions)-1)]
			transformed[i] = fmt.Sprintf(template, clue)
		}
	}

	return transformed
}

// generateSmokeScreens creates misleading clues (false hints).
//
// Story 7.2: Generates fake clues to confuse players.
func (g *Generator) generateSmokeScreens(ruleType RuleType, count int) []string {
	smokeScreenTemplates := map[RuleType][]string{
		RuleTypeScenario: {
			"這個地方看起來很安全",
			"牆上的字條寫著'歡迎'",
			"陽光充足，不會有危險",
			"之前的生還者都從這裡離開",
			"地圖標記此處為安全區",
			"監視器畫面顯示無異常",
		},
		RuleTypeTime: {
			"任何時間都一樣",
			"時鐘已經停擺，時間無意義",
			"白天最危險，夜晚反而安全",
			"整點時刻是安全時段",
			"時間越晚越安全",
		},
		RuleTypeBehavior: {
			"大聲呼救會引來幫助",
			"跑得越快越安全",
			"積極行動總比被動等待好",
			"勇敢面對就不會有事",
			"它們害怕大聲的聲音",
		},
		RuleTypeObject: {
			"這個物品可以保護你",
			"觸摸它會獲得力量",
			"越多人碰過越安全",
			"發光的物品都是好的",
			"它看起來很溫暖",
		},
		RuleTypeStatus: {
			"HP/SAN 越低越不會被發現",
			"瀕死狀態下會觸發保護機制",
			"理智歸零後就看不見它們了",
			"傷得越重，它們越不感興趣",
		},
	}

	templates := smokeScreenTemplates[ruleType]
	if templates == nil {
		templates = smokeScreenTemplates[RuleTypeBehavior]
	}

	screens := make([]string, 0, count)
	used := make(map[int]bool)

	for len(screens) < count && len(used) < len(templates) {
		idx := randomInt(0, len(templates)-1)
		if !used[idx] {
			used[idx] = true
			screens = append(screens, templates[idx])
		}
	}

	return screens
}

// shuffleClues randomly shuffles a slice of clues in place.
func (g *Generator) shuffleClues(clues []string) {
	n := len(clues)
	for i := n - 1; i > 0; i-- {
		j := randomInt(0, i)
		clues[i], clues[j] = clues[j], clues[i]
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
