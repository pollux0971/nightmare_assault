package base

import "github.com/nightmare-assault/nightmare-assault/internal/game"

// T_BASE_MOD - Modifier templates for difficulty, length, and adult mode

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
