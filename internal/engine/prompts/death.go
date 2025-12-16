// Package prompts provides prompt templates for story generation.
// This file maintains backward compatibility by re-exporting from the new template structure.
package prompts

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/templates/events"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DeathNarrativePrompt generates a prompt for creating death narrative.
// DEPRECATED: Use templates/events.DeathNarrativePrompt directly.
func DeathNarrativePrompt(deathInfo *game.DeathInfo, context string) string {
	return events.DeathNarrativePrompt(deathInfo, context)
}

// InsanityNarrativePrompt generates a specialized prompt for sanity death.
// DEPRECATED: Use templates/events.InsanityNarrativePrompt directly.
func InsanityNarrativePrompt(deathInfo *game.DeathInfo, context string) string {
	return events.InsanityNarrativePrompt(deathInfo, context)
}

// RuleDeathNarrativePrompt generates a prompt for rule violation death.
// DEPRECATED: Use templates/events.RuleDeathNarrativePrompt directly.
func RuleDeathNarrativePrompt(deathInfo *game.DeathInfo, ruleDescription string, context string) string {
	return events.RuleDeathNarrativePrompt(deathInfo, ruleDescription, context)
}

// DefaultDeathNarrative returns a fallback death narrative if LLM fails.
// DEPRECATED: Use templates/events.DefaultDeathNarrative directly.
func DefaultDeathNarrative(deathInfo *game.DeathInfo) string {
	return events.DefaultDeathNarrative(deathInfo)
}

// BuildDeathPrompt selects the appropriate prompt based on death type.
// DEPRECATED: Use templates/events.BuildDeathPrompt directly.
func BuildDeathPrompt(deathInfo *game.DeathInfo, ruleDescription, context string) string {
	return events.BuildDeathPrompt(deathInfo, ruleDescription, context)
}

// FormatDeathInfo formats death info for display/logging.
func FormatDeathInfo(info *game.DeathInfo) string {
	var b strings.Builder
	b.WriteString("=== 死亡記錄 ===\n")
	b.WriteString(fmt.Sprintf("死因: %s\n", info.Type.String()))
	b.WriteString(fmt.Sprintf("章節: %d\n", info.Chapter))
	b.WriteString(fmt.Sprintf("位置: %s\n", info.Location))
	b.WriteString(fmt.Sprintf("最後行動: %s\n", info.LastAction))
	b.WriteString(fmt.Sprintf("最終 HP: %d\n", info.FinalHP))
	b.WriteString(fmt.Sprintf("最終 SAN: %d\n", info.FinalSAN))
	if info.TriggeringRuleID != "" {
		b.WriteString(fmt.Sprintf("觸發規則: %s\n", info.TriggeringRuleID))
	}
	return b.String()
}
