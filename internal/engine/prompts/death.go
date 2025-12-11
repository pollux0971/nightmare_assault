package prompts

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

// DeathNarrativePrompt generates a prompt for creating death narrative.
func DeathNarrativePrompt(deathInfo *game.DeathInfo, context string) string {
	var deathReason string
	switch deathInfo.Type {
	case game.DeathTypeHP:
		deathReason = "體力完全耗盡，無法再支撐下去"
	case game.DeathTypeSAN:
		deathReason = "理智徹底崩潰，被恐懼和瘋狂吞噬"
	case game.DeathTypeRule:
		deathReason = fmt.Sprintf("違反了隱藏的規則（規則 ID: %s），遭受即時懲罰", deathInfo.TriggeringRuleID)
	}

	prompt := fmt.Sprintf(`Generate a dramatic death narrative for a horror game in Traditional Chinese (繁體中文).

**Death Information:**
- Death Type: %s
- Death Reason: %s
- Location: %s
- Last Action: %s
- Final HP: %d
- Final SAN: %d
- Chapter: %d

**Context:**
%s

**Requirements:**
1. Write 100-200 characters of dramatic death narrative
2. Use second person perspective ("你...")
3. Describe the final moments vividly
4. Include sensory details (sight, sound, smell, touch)
5. Match the death type:
   - HP death: Physical collapse, exhaustion, injuries
   - SAN death: Madness, hallucinations, loss of self
   - Rule death: Supernatural punishment, horror reveal
6. End with a sense of finality but leave room for mystery
7. Do NOT include any choices or continuation
8. Keep the tone dark and atmospheric

**Output ONLY the death narrative text, nothing else.**`,
		deathInfo.Type.EnglishName(),
		deathReason,
		deathInfo.Location,
		deathInfo.LastAction,
		deathInfo.FinalHP,
		deathInfo.FinalSAN,
		deathInfo.Chapter,
		context,
	)

	return prompt
}

// InsanityNarrativePrompt generates a specialized prompt for sanity death.
func InsanityNarrativePrompt(deathInfo *game.DeathInfo, context string) string {
	prompt := fmt.Sprintf(`Generate an INSANITY death narrative for a Lovecraftian horror game in Traditional Chinese (繁體中文).

**Situation:**
The player's sanity has completely shattered. Their mind can no longer distinguish reality from nightmare.

- Location: %s
- Last Action: %s
- Final SAN: %d

**Context:**
%s

**Requirements:**
1. Write 150-250 characters of madness-themed narrative
2. Use second person ("你...")
3. Describe the dissolution of reality
4. Include:
   - Impossible geometries
   - Voices and whispers
   - Loss of identity
   - Cosmic insignificance
5. The narrative should feel fragmented and disorienting
6. Mix lucid moments with complete incoherence
7. Reference things that "should not be" or "cannot be named"
8. End with complete ego dissolution

**Style notes:**
- Sentence structure should break down as the narrative progresses
- Include sensory contradictions (seeing sounds, hearing colors)
- Time should feel non-linear

**Output ONLY the narrative text, nothing else.**`,
		deathInfo.Location,
		deathInfo.LastAction,
		deathInfo.FinalSAN,
		context,
	)

	return prompt
}

// RuleDeathNarrativePrompt generates a prompt for rule violation death.
func RuleDeathNarrativePrompt(deathInfo *game.DeathInfo, ruleDescription string, context string) string {
	prompt := fmt.Sprintf(`Generate a RULE VIOLATION death narrative for a horror game in Traditional Chinese (繁體中文).

**The player violated a hidden rule and must face the supernatural consequence.**

**Rule Information:**
- Rule ID: %s
- Rule Description: %s

**Death Context:**
- Location: %s
- Last Action (that violated the rule): %s
- Final HP: %d
- Final SAN: %d

**Story Context:**
%s

**Requirements:**
1. Write 100-200 characters of punishment narrative
2. Use second person ("你...")
3. The consequence should feel:
   - Sudden and inevitable
   - Connected to the violated rule
   - Supernaturally terrifying
4. NOW reveal the rule (since the game is over)
   - Show what the player SHOULD have done
   - Make it feel tragically obvious in hindsight
5. Include a horrifying entity/force as the executor
6. End with the player realizing their mistake too late

**Tone:** Mix of cosmic horror and tragic irony - the rule was there all along.

**Output ONLY the narrative text, nothing else.**`,
		deathInfo.TriggeringRuleID,
		ruleDescription,
		deathInfo.Location,
		deathInfo.LastAction,
		deathInfo.FinalHP,
		deathInfo.FinalSAN,
		context,
	)

	return prompt
}

// DefaultDeathNarrative returns a fallback death narrative if LLM fails.
func DefaultDeathNarrative(deathInfo *game.DeathInfo) string {
	switch deathInfo.Type {
	case game.DeathTypeHP:
		return `你的身體終於承受不住了。每一次呼吸都像是在吸入碎玻璃，視野逐漸模糊。

腿軟了，跪倒在冰冷的地板上。你試圖伸出手，卻只看到自己的手指在顫抖。

「這就是...結束了嗎？」

黑暗從四面八方湧來，將你徹底吞沒。`

	case game.DeathTypeSAN:
		return `他們一直都在看著你。牆壁裡、天花板上、你自己的倒影裡。

你開始聽到自己的名字，但那不是你的聲音。那是...數不清的聲音，同時在說話。

「你終於...成為了我們的一部分。」

你笑了，或者哭了，你分不清楚。因為你已經不再是「你」了。`

	case game.DeathTypeRule:
		return `有些規則是不能被打破的。不是因為它們被寫下來，而是因為...它們一直存在。

當你踏出那一步的瞬間，你就知道自己錯了。空氣凝固了，時間停止了。

「你本該知道的。」一個聲音說。

但現在，一切都太遲了。`

	default:
		return `你的冒險在此結束。

也許在另一個故事裡，你能夠生存下來。

但不是這一個。`
	}
}

// BuildDeathPrompt selects the appropriate prompt based on death type.
func BuildDeathPrompt(deathInfo *game.DeathInfo, ruleDescription, context string) string {
	switch deathInfo.Type {
	case game.DeathTypeSAN:
		return InsanityNarrativePrompt(deathInfo, context)
	case game.DeathTypeRule:
		if ruleDescription != "" {
			return RuleDeathNarrativePrompt(deathInfo, ruleDescription, context)
		}
		return DeathNarrativePrompt(deathInfo, context)
	default:
		return DeathNarrativePrompt(deathInfo, context)
	}
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
