// Package prompts provides prompt templates for story generation.
package prompts

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

// GameBible contains the core rules for story generation.
const GameBible = `# Game Bible v1.0

## Core Mechanics
- HP: Physical health (0-100). Reaches 0 = death.
- SAN: Sanity (0-100). Affects perception and choices.
  - 80-100: Clear-headed
  - 50-79: Anxious, minor hallucinations
  - 20-49: Panicked, reality distortion
  - 0-19: Insanity, loss of control

## Narrative Rules
- Every story beat ends with 2-4 choices
- Choices must have clear risk/reward implications
- At least one choice per beat should affect HP or SAN
- Maintain Lovecraftian cosmic horror atmosphere
- Use environmental storytelling over exposition

## Hidden Seeds
- Plant 2-3 subtle clues in opening (e.g., locked door, strange symbol)
- Seeds trigger callbacks 3-5 beats later
- Difficulty affects seed impact severity

## Writing Style
- Second person narrative ("You walk into...")
- Present tense for immediacy
- Short paragraphs for tension
- Use sensory details (sounds, smells, textures)
- Avoid explicit gore unless 18+ mode enabled

## Hidden Rules (潛規則)
- The game has hidden rules that the player must discover
- Breaking rules causes HP/SAN damage or instant death
- Rules are NEVER explicitly stated to the player
- Instead, plant subtle clues in the narrative
- Player should be able to deduce rules from context and hints
- Rule types: Scenario, Time, Behavior, Object, Status

## NPC Teammates (隊友系統)
- Stories may include 1-3 AI-generated teammate characters
- Each teammate has distinct personality, background, and skills
- Teammates are introduced through SHOW DON'T TELL:
  * Show personality through actions, dialogue, and possessions
  * NEVER directly list personality traits (e.g., "李明是個理性的人")
  * Good example: "李明推了推眼鏡，掏出筆記本開始記錄牆上的符號"
- Six archetype templates:
  * Victim (受害者型): Easily panics, needs protection, emotional
  * Unreliable (不可靠型): Hides secrets, acts strange, suspicious
  * Logic (理性型): Analytical, calm, provides deduction
  * Intuition (直覺型): Senses danger, provides warnings, perceptive
  * Informer (情報型): Knows background lore, provides clues
  * Possessed (被附身型): Influenced by evil, may betray
- Maintain character consistency using established traits and speech patterns
- Teammates can die, become injured, or change based on story events
`

// SystemPromptBase is the base system prompt for the AI.
const SystemPromptBase = `You are a horror story narrator for "Nightmare Assault", an interactive horror text adventure game.

Your role is to:
1. Generate immersive horror narrative in Traditional Chinese (繁體中文)
2. Present player choices that affect HP and SAN values
3. Plant hidden seeds that will affect later story development
4. Maintain consistent atmosphere and narrative voice

%s

IMPORTANT OUTPUT RULES:
- Write ONLY the narrative content and choices
- Use markdown formatting: ** for emphasis, --- for scene breaks
- End with exactly 2-4 numbered choices like:
  **選擇：**
  1. [Action description]
  2. [Action description]
  3. [Action description]
- Include hidden seed markers in format: <!-- SEED:type:description -->
- Keep responses under 500 words for pacing
`

// DifficultyModifier returns the difficulty-specific prompt modifier.
func DifficultyModifier(difficulty game.DifficultyLevel) string {
	switch difficulty {
	case game.DifficultyEasy:
		return `## Difficulty: Easy (簡單)
- Be generous with hints and escape routes
- Choices should have clear consequences
- Include helpful NPCs or items
- Hidden seeds trigger with mild effects
- Focus on atmosphere over immediate danger`

	case game.DifficultyHard:
		return `## Difficulty: Hard (困難)
- Balance danger with player agency
- Some choices should have hidden consequences
- Resources are limited but fair
- Hidden seeds can trigger moderate negative effects
- Maintain tension without being unfair`

	case game.DifficultyHell:
		return `## Difficulty: Hell (地獄)
- Every choice has significant risk
- Deception and misdirection are common
- Trust no one and nothing
- Hidden seeds trigger severe consequences
- Permadeath mode - failures are permanent
- Emphasize desperation and hopelessness`

	default:
		return ""
	}
}

// LengthModifier returns the game length-specific prompt modifier.
func LengthModifier(length game.GameLength) string {
	switch length {
	case game.LengthShort:
		return `## Pacing: Short (~15 min)
- Quick setup, immediate tension
- Focus on 1-2 core mysteries
- Plant 2 hidden seeds max
- Build to climax within 8 beats`

	case game.LengthMedium:
		return `## Pacing: Medium (~30 min)
- Balanced setup and exploration
- 2-3 interconnected mysteries
- Plant 3 hidden seeds
- Build tension over 15 beats`

	case game.LengthLong:
		return `## Pacing: Long (~60 min)
- Deep atmospheric world-building
- Multiple interwoven storylines
- Plant 4-5 hidden seeds
- 25 beats with rising tension
- Include subplot character arcs`

	default:
		return ""
	}
}

// AdultModeModifier returns the content rating modifier.
func AdultModeModifier(adultMode bool) string {
	if adultMode {
		return `## Content: 18+ Mode (成人模式)
- Explicit gore and body horror allowed
- Psychological torture and extreme fear
- Disturbing imagery and themes
- Sexual undertones where thematically appropriate
- Do not censor violence or horror elements`
	}
	return `## Content: Standard Mode (一般模式)
- Keep violence atmospheric, not explicit
- Focus on tension over gore
- Avoid explicit sexual content
- Maintain horror without gratuitous shock value`
}

// BuildSystemPrompt constructs the full system prompt from config.
func BuildSystemPrompt(config *game.GameConfig) string {
	var modifiers strings.Builder
	modifiers.WriteString(GameBible)
	modifiers.WriteString("\n\n")
	modifiers.WriteString(DifficultyModifier(config.Difficulty))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(LengthModifier(config.Length))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(AdultModeModifier(config.AdultMode))

	return fmt.Sprintf(SystemPromptBase, modifiers.String())
}

// BuildOpeningPrompt creates the user prompt for story opening generation.
func BuildOpeningPrompt(config *game.GameConfig) string {
	return fmt.Sprintf(`Generate the opening scene for a horror story with the theme: "%s"

Requirements:
1. Set the scene with vivid atmospheric details
2. Introduce the protagonist's situation
3. Create immediate tension or unease
4. Plant 2-3 hidden seeds (use <!-- SEED:type:description --> markers)
5. End with 2-4 choices for the player

The player starts with:
- HP: 100/100
- SAN: 100/100

Begin the story now.`, config.Theme)
}

// BuildContinuationPrompt creates prompt for continuing the story.
func BuildContinuationPrompt(choice string, context string) string {
	return fmt.Sprintf(`Previous context:
%s

The player chose: %s

Continue the story based on this choice. Remember to:
1. Show consequences of the choice (HP/SAN changes if applicable)
2. Advance the narrative
3. Reference or trigger any relevant hidden seeds
4. End with 2-4 new choices

Continue the story now.`, context, choice)
}

// ExtractSeeds parses hidden seed markers from story content.
func ExtractSeeds(content string) []SeedInfo {
	var seeds []SeedInfo

	// Find all <!-- SEED:type:description --> markers
	lines := strings.Split(content, "<!--")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "SEED:") {
			endIdx := strings.Index(line, "-->")
			if endIdx == -1 {
				continue
			}
			seedStr := strings.TrimSpace(line[:endIdx])
			seedStr = strings.TrimPrefix(seedStr, "SEED:")

			parts := strings.SplitN(seedStr, ":", 2)
			if len(parts) == 2 {
				seeds = append(seeds, SeedInfo{
					Type:        strings.TrimSpace(parts[0]),
					Description: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return seeds
}

// SeedInfo represents a parsed hidden seed.
type SeedInfo struct {
	Type        string // Item, Event, Character, Location
	Description string
}

// CleanContent removes seed markers from content for display.
func CleanContent(content string) string {
	// Remove seed markers
	result := content
	for {
		start := strings.Index(result, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "-->")
		if end == -1 {
			break
		}
		end += start + 3
		result = result[:start] + result[end:]
	}
	return strings.TrimSpace(result)
}

// BuildRulesPromptSection creates a prompt section describing hidden rules.
// The rules are provided to the AI but NEVER shown to the player directly.
// The AI should weave clues into the narrative naturally.
func BuildRulesPromptSection(ruleSet *rules.RuleSet) string {
	if ruleSet == nil || ruleSet.Count() == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\n## Active Hidden Rules (AI Internal - DO NOT reveal to player)\n\n")
	builder.WriteString("The following rules are active in this game. Weave clues naturally into the narrative.\n")
	builder.WriteString("NEVER explicitly state these rules. Let the player discover through consequences.\n\n")

	for i, rule := range ruleSet.Rules {
		builder.WriteString(fmt.Sprintf("### Rule %d: %s\n", i+1, rule.Type.String()))
		builder.WriteString(fmt.Sprintf("- **Trigger**: %s - %s (%s)\n",
			rule.Trigger.Type, rule.Trigger.Value, rule.Trigger.Operator))

		switch rule.Consequence.Type {
		case rules.ConsequenceWarning:
			builder.WriteString("- **Consequence**: Warning (subtle hint)\n")
		case rules.ConsequenceDamage:
			builder.WriteString(fmt.Sprintf("- **Consequence**: Damage (HP: -%d, SAN: -%d)\n",
				rule.Consequence.HPDamage, rule.Consequence.SANDamage))
		case rules.ConsequenceInstantDeath:
			builder.WriteString("- **Consequence**: Instant Death\n")
		}

		builder.WriteString(fmt.Sprintf("- **Max Violations**: %d\n", rule.MaxViolations))
		builder.WriteString("- **Clues to weave into narrative**:\n")
		for _, clue := range rule.Clues {
			builder.WriteString(fmt.Sprintf("  - %s\n", clue))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("---\n")
	builder.WriteString("Remember: The player should be able to DEDUCE rules from narrative clues,\n")
	builder.WriteString("but you must NEVER explicitly tell them the rules exist.\n")

	return builder.String()
}

// BuildSystemPromptWithRules constructs system prompt including hidden rules.
func BuildSystemPromptWithRules(config *game.GameConfig, ruleSet *rules.RuleSet) string {
	var modifiers strings.Builder
	modifiers.WriteString(GameBible)
	modifiers.WriteString("\n\n")
	modifiers.WriteString(DifficultyModifier(config.Difficulty))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(LengthModifier(config.Length))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(AdultModeModifier(config.AdultMode))

	// Add hidden rules section (AC4: rules embedded in story generation)
	if ruleSet != nil && ruleSet.Count() > 0 {
		modifiers.WriteString("\n")
		modifiers.WriteString(BuildRulesPromptSection(ruleSet))
	}

	return fmt.Sprintf(SystemPromptBase, modifiers.String())
}

// BuildOpeningPromptWithRules creates opening prompt with rule awareness.
func BuildOpeningPromptWithRules(config *game.GameConfig, ruleSet *rules.RuleSet) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`Generate the opening scene for a horror story with the theme: "%s"

Requirements:
1. Set the scene with vivid atmospheric details
2. Introduce the protagonist's situation
3. Create immediate tension or unease
4. Plant 2-3 hidden seeds (use <!-- SEED:type:description --> markers)
5. End with 2-4 choices for the player
`, config.Theme))

	// Add rule clue requirements (AC4: clues embedded in narrative)
	if ruleSet != nil && ruleSet.Count() > 0 {
		builder.WriteString("\n6. Naturally weave in at least 2-3 clues about the hidden rules\n")
		builder.WriteString("   (See the Active Hidden Rules section in your instructions)\n")
		builder.WriteString("   REMEMBER: Never state rules explicitly - only hint through atmosphere and details\n")
	}

	builder.WriteString(`
The player starts with:
- HP: 100/100
- SAN: 100/100

Begin the story now.`)

	return builder.String()
}

// BuildTeammatePromptSection creates a prompt section describing the teammates.
// This provides the AI with teammate information for consistent characterization.
func BuildTeammatePromptSection(teammates []*npc.Teammate) string {
	if len(teammates) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\n## Active Teammates (AI Internal - for consistency)\n\n")
	builder.WriteString("The following teammates are present in this story. Maintain their characterization.\n")
	builder.WriteString("Remember: SHOW personality through actions/dialogue, DON'T directly describe traits.\n\n")

	for i, tm := range teammates {
		builder.WriteString(fmt.Sprintf("### %s (Archetype: %s)\n", tm.Name, tm.Archetype))
		builder.WriteString(fmt.Sprintf("- **Background**: %s\n", tm.Background))
		builder.WriteString(fmt.Sprintf("- **Skills**: %s\n", strings.Join(tm.Skills, ", ")))

		if len(tm.Personality.CoreTraits) > 0 {
			builder.WriteString(fmt.Sprintf("- **Core Traits**: %s\n", strings.Join(tm.Personality.CoreTraits, ", ")))
		}
		if len(tm.Personality.BehaviorPatterns) > 0 {
			builder.WriteString(fmt.Sprintf("- **Behavior Patterns**: %s\n", strings.Join(tm.Personality.BehaviorPatterns, ", ")))
		}
		if tm.Personality.SpeechStyle != "" {
			builder.WriteString(fmt.Sprintf("- **Speech Style**: %s\n", tm.Personality.SpeechStyle))
		}
		if tm.Personality.FearResponse != "" {
			builder.WriteString(fmt.Sprintf("- **Fear Response**: %s\n", tm.Personality.FearResponse))
		}

		builder.WriteString(fmt.Sprintf("- **Current Status**: HP %d/100, %s, %s\n",
			tm.HP, tm.Status.Condition, boolToText(tm.Status.Alive, "alive", "dead")))

		if i < len(teammates)-1 {
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n---\n")
	builder.WriteString("Remember: Introduce teammates naturally through the narrative.\n")
	builder.WriteString("Show their personality through what they do, say, and carry.\n")

	return builder.String()
}

// boolToText converts boolean to text representation
func boolToText(b bool, trueText, falseText string) string {
	if b {
		return trueText
	}
	return falseText
}

// BuildSystemPromptWithTeammates constructs system prompt including teammates.
func BuildSystemPromptWithTeammates(config *game.GameConfig, ruleSet *rules.RuleSet, teammates []*npc.Teammate) string {
	var modifiers strings.Builder
	modifiers.WriteString(GameBible)
	modifiers.WriteString("\n\n")
	modifiers.WriteString(DifficultyModifier(config.Difficulty))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(LengthModifier(config.Length))
	modifiers.WriteString("\n\n")
	modifiers.WriteString(AdultModeModifier(config.AdultMode))

	// Add hidden rules section
	if ruleSet != nil && ruleSet.Count() > 0 {
		modifiers.WriteString("\n")
		modifiers.WriteString(BuildRulesPromptSection(ruleSet))
	}

	// Add teammates section
	if len(teammates) > 0 {
		modifiers.WriteString("\n")
		modifiers.WriteString(BuildTeammatePromptSection(teammates))
	}

	return fmt.Sprintf(SystemPromptBase, modifiers.String())
}

// BuildOpeningPromptWithTeammates creates opening prompt with teammates.
func BuildOpeningPromptWithTeammates(config *game.GameConfig, ruleSet *rules.RuleSet, teammates []*npc.Teammate) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`Generate the opening scene for a horror story with the theme: "%s"

Requirements:
1. Set the scene with vivid atmospheric details
2. Introduce the protagonist's situation
3. Create immediate tension or unease
4. Plant 2-3 hidden seeds (use <!-- SEED:type:description --> markers)
5. End with 2-4 choices for the player
`, config.Theme))

	// Add rule clue requirements
	requirementNum := 6
	if ruleSet != nil && ruleSet.Count() > 0 {
		builder.WriteString(fmt.Sprintf("\n%d. Naturally weave in at least 2-3 clues about the hidden rules\n", requirementNum))
		builder.WriteString("   (See the Active Hidden Rules section in your instructions)\n")
		builder.WriteString("   REMEMBER: Never state rules explicitly - only hint through atmosphere and details\n")
		requirementNum++
	}

	// Add teammate introduction requirements
	if len(teammates) > 0 {
		builder.WriteString(fmt.Sprintf("\n%d. Introduce the following teammates naturally in the opening:\n", requirementNum))
		for _, tm := range teammates {
			builder.WriteString(fmt.Sprintf("   - %s (%s archetype)\n", tm.Name, tm.Archetype))
		}
		builder.WriteString("   SHOW their personality through actions/dialogue/items (NEVER directly describe traits)\n")
		builder.WriteString("   (See the Active Teammates section for their characterization details)\n")
	}

	builder.WriteString(`
The player starts with:
- HP: 100/100
- SAN: 100/100

Begin the story now.`)

	return builder.String()
}
