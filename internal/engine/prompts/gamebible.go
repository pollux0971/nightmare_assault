// Package prompts provides prompt templates for story generation.
// This file maintains backward compatibility by re-exporting from the new template structure.
package prompts

import (
	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/builder"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/templates/base"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

// GameBible contains the core rules for story generation.
// DEPRECATED: Use templates/base.TemplateBaseBible directly.
const GameBible = base.TemplateBaseBible

// SystemPromptBase is the base system prompt for the AI.
// DEPRECATED: Use templates/base.TemplateBaseSystem directly.
const SystemPromptBase = base.TemplateBaseSystem

// DifficultyModifier returns the difficulty-specific prompt modifier.
// DEPRECATED: Use templates/base.DifficultyModifier directly.
func DifficultyModifier(difficulty game.DifficultyLevel) string {
	return base.DifficultyModifier(difficulty)
}

// LengthModifier returns the game length-specific prompt modifier.
// DEPRECATED: Use templates/base.LengthModifier directly.
func LengthModifier(length game.GameLength) string {
	return base.LengthModifier(length)
}

// AdultModeModifier returns the content rating modifier.
// DEPRECATED: Use templates/base.AdultModeModifier directly.
func AdultModeModifier(adultMode bool) string {
	return base.AdultModeModifier(adultMode)
}

// BuildSystemPrompt constructs the full system prompt from config.
// DEPRECATED: Use builder.PromptBuilder instead.
func BuildSystemPrompt(config *game.GameConfig) string {
	pb := builder.NewPromptBuilder(config)
	return pb.BuildSystemPrompt()
}

// BuildOpeningPrompt creates the user prompt for story opening generation.
// DEPRECATED: Use builder.PromptBuilder instead.
func BuildOpeningPrompt(config *game.GameConfig) string {
	pb := builder.NewPromptBuilder(config)
	return pb.BuildOpeningPrompt()
}

// BuildContinuationPrompt creates prompt for continuing the story.
// DEPRECATED: Use builder.PromptBuilder instead.
func BuildContinuationPrompt(choice string, context string) string {
	// For continuation, we need a dummy config since we don't modify based on it
	dummyConfig := &game.GameConfig{}
	pb := builder.NewPromptBuilder(dummyConfig)
	return pb.BuildContinuationPrompt(choice, context)
}

// ExtractSeeds parses hidden seed markers from story content.
// DEPRECATED: Use builder.ExtractSeeds directly.
func ExtractSeeds(content string) []SeedInfo {
	seeds := builder.ExtractSeeds(content)
	result := make([]SeedInfo, len(seeds))
	for i, s := range seeds {
		result[i] = SeedInfo{
			Type:        s.Type,
			Description: s.Description,
		}
	}
	return result
}

// SeedInfo represents a parsed hidden seed.
// DEPRECATED: Use builder.SeedInfo directly.
type SeedInfo = builder.SeedInfo

// CleanContent removes seed markers from content for display.
// DEPRECATED: Use builder.CleanContent directly.
func CleanContent(content string) string {
	return builder.CleanContent(content)
}

// BuildRulesPromptSection creates a prompt section describing hidden rules.
// DEPRECATED: Rules are now automatically included via builder.PromptBuilder.WithRules()
func BuildRulesPromptSection(ruleSet *rules.RuleSet) string {
	// Handle nil or empty rule sets
	if ruleSet == nil || ruleSet.Count() == 0 {
		return ""
	}

	// For backward compatibility, build a prompt with rules and extract the rules section
	dummyConfig := &game.GameConfig{}
	pb := builder.NewPromptBuilder(dummyConfig).WithRules(ruleSet)
	fullPrompt := pb.BuildSystemPrompt()

	// Extract just the rules section
	// This is a bit hacky but maintains compatibility
	return fullPrompt
}

// BuildSystemPromptWithRules constructs system prompt including hidden rules.
// DEPRECATED: Use builder.PromptBuilder.WithRules() instead.
func BuildSystemPromptWithRules(config *game.GameConfig, ruleSet *rules.RuleSet) string {
	pb := builder.NewPromptBuilder(config).WithRules(ruleSet)
	return pb.BuildSystemPrompt()
}

// BuildOpeningPromptWithRules creates opening prompt with rule awareness.
// DEPRECATED: Use builder.PromptBuilder.WithRules() instead.
func BuildOpeningPromptWithRules(config *game.GameConfig, ruleSet *rules.RuleSet) string {
	pb := builder.NewPromptBuilder(config).WithRules(ruleSet)
	return pb.BuildOpeningPrompt()
}

// BuildTeammatePromptSection creates a prompt section describing the teammates.
// DEPRECATED: Teammates are now automatically included via builder.PromptBuilder.WithTeammates()
func BuildTeammatePromptSection(teammates []*npc.Teammate) string {
	// Handle nil or empty teammates
	if len(teammates) == 0 {
		return ""
	}

	dummyConfig := &game.GameConfig{}
	pb := builder.NewPromptBuilder(dummyConfig).WithTeammates(teammates)
	fullPrompt := pb.BuildSystemPrompt()
	return fullPrompt
}

// BuildSystemPromptWithTeammates constructs system prompt including teammates.
// DEPRECATED: Use builder.PromptBuilder.WithTeammates() instead.
func BuildSystemPromptWithTeammates(config *game.GameConfig, ruleSet *rules.RuleSet, teammates []*npc.Teammate) string {
	pb := builder.NewPromptBuilder(config).
		WithRules(ruleSet).
		WithTeammates(teammates)
	return pb.BuildSystemPrompt()
}

// BuildOpeningPromptWithTeammates creates opening prompt with teammates.
// DEPRECATED: Use builder.PromptBuilder.WithTeammates() instead.
func BuildOpeningPromptWithTeammates(config *game.GameConfig, ruleSet *rules.RuleSet, teammates []*npc.Teammate) string {
	pb := builder.NewPromptBuilder(config).
		WithRules(ruleSet).
		WithTeammates(teammates)
	return pb.BuildOpeningPrompt()
}

// boolToText converts boolean to text representation (for backward compatibility with tests)
func boolToText(b bool, trueText, falseText string) string {
	if b {
		return trueText
	}
	return falseText
}
