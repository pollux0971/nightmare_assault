package narration

import (
	"context"
	"fmt"
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// ContentRequest represents a request to generate narrative content for a beat
//
// This is the input for Content Mode, called during Phase 2 (Game Loop).
// It contains all necessary context for generating a single beat's narrative.
type ContentRequest struct {
	Beat             int                 // Current beat number
	GameState        *engine.GameStateV2 // Complete game state
	LastPlayerChoice string              // Player's last choice text
	JudgeResult      *JudgeResult        // Result from Judge Agent (may be nil)
	Difficulty       string              // Difficulty level: "easy", "normal", "hard", "hell"
}

// ContentResponse represents the generated narrative content for a beat
//
// This is the output from Content Mode, containing the narrative text and
// all state changes that occurred during this beat.
type ContentResponse struct {
	MainNarrative    string   // 500-1200 字的敘事內容
	PlantedSeeds     []string // 新種植的 LocalSeed IDs
	HarvestedSeeds   []string // 被收割的 LocalSeed IDs
	RuleHints        []string // 規則提示文本列表
	HPChange         int      // HP 變化值
	SANChange        int      // SAN 變化值
	DeathTriggered   bool     // 是否觸發死亡（HP≤0 或 SAN≤0）
	NextBeatPreview  string   // 下一 Beat 預告（可選，50 字內）
	SceneChanged     bool     // 場景是否變化
	NewScene         string   // 新場景名稱（如果 SceneChanged=true）
}

// JudgeResult represents the judgment result from Judge Agent
//
// This is a temporary placeholder until Story 6.7 (Judge Agent) is implemented.
// It contains the consequences of the player's last choice.
type JudgeResult struct {
	ViolatedRules   []string // 違反的規則 IDs
	ImpactLevel     string   // 影響等級：none, minor, moderate, major, lethal
	HPDelta         int      // HP 變化（負數表示傷害）
	SANDelta        int      // SAN 變化（負數表示理智下降）
	IsHallucination bool     // 是否為幻覺選項
	Reason          string   // 判定原因描述
}

// InvokeContent generates narrative content for a single beat
//
// This is the main method for Content Mode. It integrates all game components:
//   - Seed Agent: Local Seed Plant/Harvest decisions
//   - Tension Manager: Atmosphere and pacing directives
//   - Context Manager: Token optimization and history assembly
//   - Template Library: Active rules, scenes, NPCs
//
// AC Requirements:
//   - Narrative length: 500-1200 字
//   - Generation time: <10 秒
//   - JSON parsing error rate: <2%
//   - Context tokens: <100k
//
// Parameters:
//   - ctx: Context for timeout control
//   - req: Content generation request
//
// Returns:
//   - *ContentResponse: Generated narrative and state changes
//   - error: Error if generation fails
func (a *NarrationAgent) InvokeContent(ctx context.Context, req *ContentRequest) (*ContentResponse, error) {
	// Validate request FIRST (before any logging or operations)
	if err := validateContentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid content request: %w", err)
	}

	log.Printf("[%s] InvokeContent started: Beat=%d", a.config.Name, req.Beat)

	// 1. Get Tension Directive from current state
	directive := a.getTensionDirective(req.GameState)
	log.Printf("[%s] Tension Directive: %s (Level: %s)",
		a.config.Name, directive.Instruction, directive.Level)

	// 2. Build Content Generation Prompt
	prompt, err := a.buildContentPrompt(req, directive)
	if err != nil {
		log.Printf("[%s] Failed to build prompt: %v", a.config.Name, err)
		return nil, fmt.Errorf("failed to build content prompt: %w", err)
	}
	log.Printf("[%s] Prompt built successfully (length: %d chars)",
		a.config.Name, len(prompt))

	// 3. Generate narrative content
	// NOTE: For now, we generate a simplified narrative
	// TODO: Replace with actual LLM call when LLMClient is available
	narrative := a.generateNarrative(req, directive)

	// 4. Generate Rule Hints if rules were violated (AC #5)
	ruleHints := a.generateRuleHints(req)

	// 5. Process HP/SAN changes from JudgeResult
	response := &ContentResponse{
		MainNarrative:   narrative,
		PlantedSeeds:    make([]string, 0),   // TODO: Integrate Seed Agent (Story 6.5)
		HarvestedSeeds:  make([]string, 0),   // TODO: Integrate Seed Agent (Story 6.5)
		RuleHints:       ruleHints,
		HPChange:        0,
		SANChange:       0,
		DeathTriggered:  false,
		NextBeatPreview: "",
		SceneChanged:    false,
		NewScene:        "",
	}

	// Process JudgeResult if present
	if req.JudgeResult != nil {
		response.HPChange = req.JudgeResult.HPDelta
		response.SANChange = req.JudgeResult.SANDelta

		// Add HP/SAN change descriptions to narrative
		if req.JudgeResult.HPDelta != 0 {
			hpDesc := DescribeHPChange(req.JudgeResult.HPDelta, req.JudgeResult.Reason)
			if hpDesc != "" {
				response.MainNarrative += "\n\n" + hpDesc
			}
		}

		if req.JudgeResult.SANDelta != 0 {
			sanDesc := DescribeSANChange(req.JudgeResult.SANDelta, req.JudgeResult.Reason)
			if sanDesc != "" {
				response.MainNarrative += "\n\n" + sanDesc
			}
		}

		// Check for death conditions (AC #6)
		currentHP := req.GameState.GetHP() + req.JudgeResult.HPDelta
		currentSAN := req.GameState.GetSAN() + req.JudgeResult.SANDelta
		if currentHP <= 0 || currentSAN <= 0 {
			response.DeathTriggered = true
			log.Printf("[%s] Death triggered: HP=%d, SAN=%d",
				a.config.Name, currentHP, currentSAN)
		}
	}

	log.Printf("[%s] InvokeContent completed: Beat=%d, NarrativeLength=%d",
		a.config.Name, req.Beat, len([]rune(response.MainNarrative)))
	return response, nil
}

// validateContentRequest validates the content generation request
func validateContentRequest(req *ContentRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Beat < 0 {
		return fmt.Errorf("beat must be >= 0, got %d", req.Beat)
	}

	if req.GameState == nil {
		return fmt.Errorf("game state cannot be nil")
	}

	// LastPlayerChoice and JudgeResult can be empty/nil for first beat

	return nil
}

// getTensionDirective gets the current tension directive based on game state
//
// AC #4: Tension Directive 整合
func (a *NarrationAgent) getTensionDirective(gameState *engine.GameStateV2) *engine.TensionDirective {
	if gameState.Tension == nil {
		// Default to LOW tension if not set
		return engine.GenerateDirective(engine.TensionLevelLow)
	}

	// Get current tension level
	level := gameState.Tension.GetLevel()
	return engine.GenerateDirective(level)
}

// buildContentPrompt builds the prompt for content generation
//
// AC #4: 整合 Tension Directive 到 Prompt
// Task 12: 完整 Prompt 模板系統
//
// This method constructs a comprehensive prompt that includes:
//   - System role and output format instructions
//   - Current game state (Beat, HP, SAN, Scene)
//   - Tension directive for atmosphere and pacing
//   - Active rules and NPCs
//   - Player's last choice and judge result
//   - Narrative constraints (500-1200 words)
func (a *NarrationAgent) buildContentPrompt(req *ContentRequest, directive *engine.TensionDirective) (string, error) {
	// Build prompt sections
	systemSection := a.buildSystemSection()
	stateSection := a.buildStateSection(req)
	tensionSection := a.buildTensionSection(directive)
	rulesSection := a.buildRulesSection(req.GameState)
	npcsSection := a.buildNPCsSection(req.GameState)
	contextSection := a.buildContextSection(req)
	constraintsSection := a.buildConstraintsSection()

	// Assemble complete prompt
	prompt := fmt.Sprintf(`%s

%s

%s

%s

%s

%s

%s`,
		systemSection,
		stateSection,
		tensionSection,
		rulesSection,
		npcsSection,
		contextSection,
		constraintsSection,
	)

	return prompt, nil
}

// buildSystemSection builds the system role and instructions
func (a *NarrationAgent) buildSystemSection() string {
	return `# 系統角色

你是「規則怪談」類型恐怖遊戲的敘事引擎。你的職責是：
1. 生成 500-1200 字的連貫敘事內容
2. 遵循張力指令控制氛圍與節奏
3. 自然融入規則提示（如果有）
4. 描述 HP/SAN 變化的原因
5. 整合 NPC 互動與對話
6. 保持敘事風格一致性`
}

// buildStateSection builds current game state information
func (a *NarrationAgent) buildStateSection(req *ContentRequest) string {
	hp := req.GameState.GetHP()
	san := req.GameState.GetSAN()
	scene := req.GameState.CurrentScene
	if scene == "" {
		scene = "未知場景"
	}

	return fmt.Sprintf(`# 當前狀態

- Beat: %d
- 場景: %s
- HP: %d / 100
- SAN: %d / 100
- 難度: %s`,
		req.Beat,
		scene,
		hp,
		san,
		req.Difficulty,
	)
}

// buildTensionSection builds tension directive information
func (a *NarrationAgent) buildTensionSection(directive *engine.TensionDirective) string {
	section := fmt.Sprintf(`# 張力指令

**等級**: %s
**指令**: %s
**字數範圍**: %d - %d 字`,
		directive.Level,
		directive.Instruction,
		directive.LengthRange.Min,
		directive.LengthRange.Max,
	)

	if len(directive.AllowedElements) > 0 {
		section += fmt.Sprintf("\n**允許元素**: %v", directive.AllowedElements)
	}

	if len(directive.ForbiddenElements) > 0 {
		section += fmt.Sprintf("\n**禁止元素**: %v", directive.ForbiddenElements)
	}

	return section
}

// buildRulesSection builds active rules information
func (a *NarrationAgent) buildRulesSection(gameState *engine.GameStateV2) string {
	if len(gameState.ActiveRules) == 0 {
		return "# 活躍規則\n\n（無活躍規則）"
	}

	rulesText := "# 活躍規則\n\n"
	for i, rule := range gameState.ActiveRules {
		warningCount := gameState.RuleWarnings[rule.ID]
		rulesText += fmt.Sprintf("%d. %s (ID: %s, 警告次數: %d)\n",
			i+1, rule.Name, rule.ID, warningCount)
	}

	return rulesText
}

// buildNPCsSection builds active NPCs information
func (a *NarrationAgent) buildNPCsSection(gameState *engine.GameStateV2) string {
	if len(gameState.NPCStates) == 0 {
		return "# 活躍 NPCs\n\n（無活躍 NPCs）"
	}

	npcsText := "# 活躍 NPCs\n\n"
	for id, npc := range gameState.NPCStates {
		npcsText += fmt.Sprintf("- %s (ID: %s)\n", npc.Name, id)
	}

	return npcsText
}

// buildContextSection builds player context and judge result
func (a *NarrationAgent) buildContextSection(req *ContentRequest) string {
	contextText := "# 玩家上下文\n\n"

	if req.LastPlayerChoice != "" {
		contextText += fmt.Sprintf("**玩家上一個選擇**: %s\n\n", req.LastPlayerChoice)
	} else {
		contextText += "**玩家上一個選擇**: （遊戲開始）\n\n"
	}

	// Add judge result if present
	if req.JudgeResult != nil {
		contextText += "**判定結果**:\n"
		if len(req.JudgeResult.ViolatedRules) > 0 {
			contextText += fmt.Sprintf("- 違反規則: %v\n", req.JudgeResult.ViolatedRules)
		}
		contextText += fmt.Sprintf("- 影響等級: %s\n", req.JudgeResult.ImpactLevel)
		if req.JudgeResult.HPDelta != 0 {
			contextText += fmt.Sprintf("- HP 變化: %+d\n", req.JudgeResult.HPDelta)
		}
		if req.JudgeResult.SANDelta != 0 {
			contextText += fmt.Sprintf("- SAN 變化: %+d\n", req.JudgeResult.SANDelta)
		}
		if req.JudgeResult.Reason != "" {
			contextText += fmt.Sprintf("- 原因: %s\n", req.JudgeResult.Reason)
		}
	}

	return contextText
}

// buildConstraintsSection builds output constraints and format requirements
func (a *NarrationAgent) buildConstraintsSection() string {
	return `# 輸出要求

**字數**: 500-1200 字
**風格**: 恐怖、懸疑、規則怪談
**語言**: 繁體中文
**格式**: 純文本敘事

請生成當前 Beat 的敘事內容。確保：
1. 敘事連貫流暢
2. 遵循張力指令的氛圍與節奏
3. 如有規則違反，自然融入後果描述
4. 描寫具體生動，營造恐怖氛圍
5. 字數控制在 500-1200 字之間`
}

// generateNarrative generates narrative content
//
// NOTE: This is a simplified implementation for testing
// TODO: Replace with actual LLM call when LLMClient is available
func (a *NarrationAgent) generateNarrative(req *ContentRequest, directive *engine.TensionDirective) string {
	// Simplified narrative generation for testing
	// In production, this would call the LLM

	narrative := fmt.Sprintf(`你站在陰暗的走廊中，四周籠罩著詭異的寂靜。

Beat %d 的故事繼續展開。`, req.Beat)

	if req.LastPlayerChoice != "" {
		narrative += fmt.Sprintf(`你剛才選擇了「%s」，現在必須面對接下來的後果。`, req.LastPlayerChoice)
	}

	// Add tension-appropriate content
	switch directive.Level {
	case engine.TensionLevelLow:
		narrative += "\n\n環境細節逐漸浮現，你注意到牆上有些奇怪的劃痕，彷彿在訴說著什麼故事。"
	case engine.TensionLevelMedium:
		narrative += "\n\n一股不安的感覺襲來，你聽見遠處傳來詭異的聲響，讓人心跳加速。"
	case engine.TensionLevelHigh:
		narrative += "\n\n危機迫在眉睫！黑暗中似乎有什麼東西在移動，你必須立即做出決定。"
	}

	return narrative
}

// generateRuleHints generates rule hints based on violated rules
//
// AC #5: Rule Hints 生成整合
//   - 從 JudgeResult 獲取違反的規則
//   - 從 GameState.ActiveRules 獲取規則描述
//   - 從 GameState.RuleWarnings 獲取當前警告次數
//   - 調用 GenerateRuleHint() 生成提示文本
//   - 更新 GameState.RuleWarnings 警告次數
//
// Returns:
//   - []string: List of generated rule hint texts (may be empty)
func (a *NarrationAgent) generateRuleHints(req *ContentRequest) []string {
	hints := make([]string, 0)

	// Check if JudgeResult exists and has violated rules
	if req.JudgeResult == nil || len(req.JudgeResult.ViolatedRules) == 0 {
		return hints
	}

	// Get max warnings for this difficulty
	maxWarnings := GetMaxWarnings(req.Difficulty)

	// Generate hints for each violated rule
	for _, ruleID := range req.JudgeResult.ViolatedRules {
		// Find the rule description from ActiveRules
		ruleDesc := a.getRuleDescription(req.GameState, ruleID)
		if ruleDesc == "" {
			// If rule not found in ActiveRules, use a generic description
			ruleDesc = "未知規則"
			log.Printf("[%s] Warning: Rule %s not found in ActiveRules", a.config.Name, ruleID)
		}

		// Get current warning count for this rule
		warningCount := req.GameState.RuleWarnings[ruleID]

		// Generate hint
		hint := GenerateRuleHint(ruleID, ruleDesc, req.Difficulty, warningCount, maxWarnings)

		// If hint was generated, add it and increment warning count
		if hint != "" {
			hints = append(hints, hint)

			// Update warning count in GameState
			req.GameState.RuleWarnings[ruleID] = warningCount + 1

			log.Printf("[%s] Generated rule hint for %s (warning %d/%d): %s",
				a.config.Name, ruleID, warningCount+1, maxWarnings, hint)
		}
	}

	return hints
}

// getRuleDescription gets the description of a rule from ActiveRules
func (a *NarrationAgent) getRuleDescription(gameState *engine.GameStateV2, ruleID string) string {
	for _, rule := range gameState.ActiveRules {
		if rule.ID == ruleID {
			return rule.Name // Using Name field as description
		}
	}
	return ""
}
