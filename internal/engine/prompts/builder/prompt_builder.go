// Package builder provides tools for combining and constructing prompts from templates.
package builder

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine/prompts/templates/base"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

// ActionIntegrationPrompt provides instructions for naturally integrating player choices.
const ActionIntegrationPrompt = `
## 玩家行動整合規則

你必須將玩家的選擇自然地整合到故事敘述中：

### 整合方式
1. **絕對禁止**使用「你選擇」「你決定」「> 你選擇:」這類元敘述
2. **必須**直接描述角色的行動和結果
3. 使用角色名稱（如果已知）或「你」
4. 描述行動的過程、感受、環境反應

### 錯誤示範 ❌
- "你選擇打開門。門開了。"
- "> 你選擇: 打開門"
- "你決定要打開門。"
- "玩家選擇了進入房間。"

### 正確示範 ✅
- "林岳深吸一口氣，將手放在冰冷的門把上。門緩緩開啟，發出刺耳的嘎吱聲..."
- "你伸出顫抖的手，指尖觸及斑駁的木門。推開的瞬間，一股腐朽的氣息撲面而來..."
- "她咬緊牙關，一腳踏進了黑暗的走廊。牆壁似乎在呼吸..."

### 行動描寫要素
- **動作細節**（如何做、動作的速度和力道）
- **感官體驗**（看到/聽到/聞到/感受到什麼）
- **環境反應**（環境如何因行動而變化）
- **情緒暗示**（角色的心理狀態，但不直接說明）
`

// JSONFormatInstructions provides the new JSON output format.
const JSONFormatInstructions = `
📋 OUTPUT FORMAT (CRITICAL):

You MUST respond with a JSON object in this exact format:

{
  "story": "純敘事內容，將玩家選擇自然融入，不要出現「你選擇」等元敘述...",
  "choice_context": {
    "situation": "1-2 句話描述當前處境，幫助玩家理解情況",
    "question": "引導玩家的問題（可選）",
    "options": ["選項1", "選項2", "選項3"]
  },
  "seeds": [
    {
      "type": "Item|Event|Character|Location",
      "description": "Description of the hidden seed"
    }
  ],
  "state_changes": {
    "hp": -5,
    "san": -10,
    "reason": "原因說明"
  }
}

### 欄位說明

**story（必填）**
- 純粹的故事敘述
- 將玩家的選擇自然融入，不要出現「你選擇」「你決定」等詞
- 不要在這裡放選項或提示
- 150-300 字

**choice_context（必填）**
- situation: 1-2 句話描述當前處境，幫助玩家理解情況
- question: 可選的引導問題，如「你要怎麼做？」
- options: 2-3 個選項，每個最多 15 字，簡潔有力

**seeds（可選）**
- 隱藏線索，用於後續劇情發展

**state_changes（如果有變化）**
- hp/san 變化值（正數增加，負數減少）
- reason: 變化原因，簡短說明

### 完整範例

{
  "story": "林岳深吸一口氣，推開了那扇斑駁的大門。門軸發出淒厲的尖叫，像是被驚醒的亡魂。\n\n門外並非他期待的月光，而是一片濃稠的黑暗。那黑暗似乎有生命，正緩緩向他蔓延過來。\n\n他的腳步不由自主地後退，背脊撞上了什麼冰冷的東西。",
  "choice_context": {
    "situation": "黑暗正在逼近，而你的背後似乎有什麼東西。",
    "question": "你要如何應對？",
    "options": [
      "轉身面對背後的東西",
      "衝進黑暗中",
      "大聲呼救"
    ]
  },
  "seeds": [
    {"type": "環境", "description": "黑暗具有實體，會對光源產生反應"}
  ],
  "state_changes": {"san": -5, "reason": "目睹非自然的黑暗"}
}

CRITICAL RULES:
- DO NOT add any text outside the JSON object
- Ensure valid JSON formatting
- story field must NOT contain meta-narrative like "你選擇"
- options must be concise and actionable
`

// PromptBuilder constructs full prompts by combining base templates with modifiers.
type PromptBuilder struct {
	config    *game.GameConfig
	ruleSet   *rules.RuleSet
	teammates []*npc.Teammate
}

// NewPromptBuilder creates a new prompt builder.
func NewPromptBuilder(config *game.GameConfig) *PromptBuilder {
	return &PromptBuilder{
		config: config,
	}
}

// WithRules adds hidden rules to the prompt builder.
func (pb *PromptBuilder) WithRules(ruleSet *rules.RuleSet) *PromptBuilder {
	pb.ruleSet = ruleSet
	return pb
}

// WithTeammates adds teammates to the prompt builder.
func (pb *PromptBuilder) WithTeammates(teammates []*npc.Teammate) *PromptBuilder {
	pb.teammates = teammates
	return pb
}

// BuildSystemPrompt constructs the full system prompt from base templates.
func (pb *PromptBuilder) BuildSystemPrompt() string {
	var modifiers strings.Builder

	// Add Game Bible
	modifiers.WriteString(base.TemplateBaseBible)
	modifiers.WriteString("\n\n")

	// Add difficulty modifier
	modifiers.WriteString(base.DifficultyModifier(pb.config.Difficulty))
	modifiers.WriteString("\n\n")

	// Add length modifier
	modifiers.WriteString(base.LengthModifier(pb.config.Length))
	modifiers.WriteString("\n\n")

	// Add adult mode modifier
	modifiers.WriteString(base.AdultModeModifier(pb.config.AdultMode))

	// Add hidden rules section if available
	if pb.ruleSet != nil && pb.ruleSet.Count() > 0 {
		modifiers.WriteString("\n")
		modifiers.WriteString(buildRulesPromptSection(pb.ruleSet))
	}

	// Add teammates section if available
	if len(pb.teammates) > 0 {
		modifiers.WriteString("\n")
		modifiers.WriteString(buildTeammatePromptSection(pb.teammates))
	}

	return fmt.Sprintf(base.TemplateBaseSystem, modifiers.String())
}

// BuildOpeningPrompt creates the user prompt for story opening generation.
func (pb *PromptBuilder) BuildOpeningPrompt() string {
	var builder strings.Builder

	maxWords := "600-800"
	if len(pb.teammates) > 0 {
		maxWords = "700-900"
	}

	builder.WriteString(fmt.Sprintf(`Generate the PROLOGUE/OPENING (序章/第一章) for a horror story with the theme: "%s"

⚠️ CRITICAL REQUIREMENTS:
- **This is Chapter 0 (序章) - a NARRATIVE-ONLY scene**
- **NO CHOICES should be provided in this chapter**
- **Length: %s words** - provide substantial atmospheric setup
- Focus on WORLD-BUILDING, ATMOSPHERE, and MYSTERY
- This chapter sets the stage; player choices begin in Chapter 1

Requirements:
1. 開場設定 (Opening Setup):
   - 詳細的環境描述（至少 4-5 句）
   - 營造濃厚的恐怖氛圍
   - 建立故事世界觀

2. 主角處境 (Protagonist Situation):
   - 介紹主角身份與來到此處的原因
   - 描述當下的心理狀態
   - 埋下主角的動機與目標

3. 線索與伏筆 (Clues and Foreshadowing):
   - 埋藏 2-3 個關鍵線索或細節
   - 這些線索會在後續章節中變得重要
   - 使用 <!-- SEED:type:description --> 標記隱藏種子

4. 神秘感 (Mystery):
   - 留下 1-2 個未解答的問題
   - 製造不安感與好奇心
   - 暗示潛在的威脅或規則
`, pb.config.Theme, maxWords))

	// Add rule clue requirements
	requirementNum := 5
	if pb.ruleSet != nil && pb.ruleSet.Count() > 0 {
		builder.WriteString(fmt.Sprintf("\n%d. 規則暗示 (Rule Hints):\n", requirementNum))
		builder.WriteString("   - 自然地埋入 1-2 個關於隱藏規則的微妙線索\n")
		builder.WriteString("   - 絕不明確說明規則 - 只透過氛圍暗示\n")
		builder.WriteString("   - 參考系統指令中的 Active Hidden Rules 區段\n")
		requirementNum++
	}

	// Add teammate introduction requirements
	if len(pb.teammates) > 0 {
		builder.WriteString(fmt.Sprintf("\n%d. 隊友介紹 (Teammate Introduction):\n", requirementNum))
		builder.WriteString("   簡短介紹以下隊友（每人 1-2 句）：\n")
		for _, tm := range pb.teammates {
			builder.WriteString(fmt.Sprintf("   - %s (%s 原型)\n", tm.Name, tm.Archetype))
		}
		builder.WriteString("   - 透過一個動作/對話/物品展示個性\n")
		builder.WriteString("   - 絕不直接描述特質 - 用行動展現\n")
		requirementNum++
	}

	builder.WriteString(`

⚠️ ENDING REQUIREMENT:
- **DO NOT include any choices or options**
- End story with: "【按任意鍵繼續到第二章】"
- This signals the player to continue to Chapter 1 where choices begin

Story Structure:
- 序章（當前）：純敘述，無選擇
- 第一章（下一章）：玩家開始做選擇
- 建立氛圍與世界觀
- 留下未解之謎

The player starts with:
- HP: 100/100
- SAN: 100/100

📋 OUTPUT FORMAT (CRITICAL):
You MUST respond with a JSON object in this exact format:

{
  "story": "Your complete story text here (` + maxWords + ` words). End with 【按任意鍵繼續到第二章】",
  "seeds": [
    {
      "type": "Item|Event|Character|Location",
      "description": "Description of the hidden seed"
    }
  ]
}

IMPORTANT FOR PROLOGUE:
- "story" field: Contains the FULL narrative text
- "choice_context" field: MUST be omitted for prologue (no choices yet)
- "seeds" field: Optional, for hidden story elements
- DO NOT add any text outside the JSON object
- Ensure valid JSON formatting
- Player will press any key to continue to Chapter 1`)

	return builder.String()
}

// BuildContinuationPrompt creates prompt for continuing the story.
func (pb *PromptBuilder) BuildContinuationPrompt(choice string, context string) string {
	var builder strings.Builder

	builder.WriteString(ActionIntegrationPrompt)
	builder.WriteString("\n\n")

	builder.WriteString(fmt.Sprintf(`## 當前場景

Previous context:
%s

## 玩家行動
玩家選擇了：%s

請將這個選擇自然地融入故事敘述中，不要使用「你選擇」「你決定」等元敘述。

⚠️ LENGTH LIMIT: Maximum 200-300 words for this chapter

Continue the story based on this choice. Remember to:
1. **自然整合玩家行動** - 直接描述角色的行為和結果，不要出現「你選擇」
2. **顯示即時後果** - 展示行動的結果（HP/SAN 變化如果適用）
3. **推進敘事一步** - 不要跳太快，專注於當下這一刻
4. **觸發隱藏種子** - 如果相關，引用或觸發之前埋下的線索
5. **保持簡短** - 這一章要短而集中
6. **提供新選項** - 在 choice_context 中提供 2-3 個新選項

`, context, choice))

	builder.WriteString(JSONFormatInstructions)

	return builder.String()
}

// buildRulesPromptSection creates a prompt section describing hidden rules.
func buildRulesPromptSection(ruleSet *rules.RuleSet) string {
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

// buildTeammatePromptSection creates a prompt section describing teammates.
func buildTeammatePromptSection(teammates []*npc.Teammate) string {
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
