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

		// AC #11: Handle hallucination choices
		if req.JudgeResult.IsHallucination {
			currentSAN := req.GameState.GetSAN() + req.JudgeResult.SANDelta
			hallucinationDesc, sanPenalty := a.HandleHallucinationChoice(currentSAN)

			// Add hallucination description to narrative
			response.MainNarrative += "\n\n" + hallucinationDesc

			// Apply additional SAN penalty for hallucination
			response.SANChange += sanPenalty

			log.Printf("[%s] Hallucination processed: currentSAN=%d, penalty=%d",
				a.config.Name, currentSAN, sanPenalty)
		}

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
		// Note: Use updated SANChange which includes hallucination penalty
		currentHP := req.GameState.GetHP() + req.JudgeResult.HPDelta
		currentSAN := req.GameState.GetSAN() + response.SANChange
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

// SceneTransition represents a scene change in the narrative
type SceneTransition struct {
	FromScene       string // Previous scene
	ToScene         string // New scene
	TransitionType  string // Type: spatial, temporal, atmospheric
	Description     string // Transition description (100-150 字)
	TensionModifier int    // Tension level change (+/- 0-20)
}

// HandleSceneTransition generates scene transition description
//
// AC #9: 場景轉換邏輯
//   - 場景轉換描述長度：100-150 字
//   - 場景轉換類型：空間轉換、時間推進、氛圍轉換
//   - 更新 GameState.CurrentScene
//
// Parameters:
//   - fromScene: Previous scene name
//   - toScene: New scene name
//   - transitionType: Type of transition (spatial/temporal/atmospheric)
//
// Returns:
//   - string: Transition description text
func (a *NarrationAgent) HandleSceneTransition(fromScene, toScene, transitionType string) string {
	var description string

	switch transitionType {
	case "spatial":
		// 空間轉換（室內 → 室外）
		description = fmt.Sprintf(`場景轉換發生了。你離開了%s，步入了%s。

環境的變化立刻映入眼簾——空氣的溫度、光線的強度、甚至是周圍聲音的質感都截然不同。這個新的空間帶給你完全不同的感受，你必須重新適應這裡的一切。你感覺到某種無形的界限被跨越了，彷彿進入了一個全新的領域。`,
			fromScene, toScene)

	case "temporal":
		// 時間推進（白天 → 夜晚）
		description = fmt.Sprintf(`時間流逝，環境隨之改變。

從%s到%s，時間的推移帶來了明顯的變化。光線的角度、陰影的長度、空氣中的氣息——一切都在提醒你時間正在流逝。這種轉變讓你意識到，某些事情只會在特定的時刻發生。你感到一種緊迫感，彷彿錯過了這個時間點，某些重要的機會就會永遠消失。`,
			fromScene, toScene)

	case "atmospheric":
		// 氛圍轉換（安全區 → 危險區）
		description = fmt.Sprintf(`氛圍驟然改變。

你從%s進入%s的瞬間，周圍的氣氛發生了劇烈的變化。原本的感覺消失了，取而代之的是一種截然不同的氛圍。你的本能在尖叫——這裡不一樣，這裡充滿了未知的危險。每一步都變得更加沉重，每一次呼吸都充滿了緊張感。你知道，你已經進入了一個全新的境地。`,
			fromScene, toScene)

	default:
		// 通用場景轉換
		description = fmt.Sprintf(`你從%s來到了%s。

周圍的環境發生了變化，新的場景展現在你眼前。這個轉變讓你必須重新評估當前的處境，適應新的環境和可能面臨的挑戰。`,
			fromScene, toScene)
	}

	return description
}

// PlotPointType represents different types of key plot points
type PlotPointType string

const (
	PlotPointIncitingIncident PlotPointType = "inciting_incident" // Act 1 End
	PlotPointMidpoint         PlotPointType = "midpoint"          // Act 2 Midpoint
	PlotPointSecondPlotPoint  PlotPointType = "second_plot_point" // Act 2 End (Lowest Point)
	PlotPointClimax           PlotPointType = "climax"            // Act 3 Climax
)

// KeyPlotPoint represents a dramatic turning point in the narrative
type KeyPlotPoint struct {
	Type            PlotPointType // Type of plot point
	Beat            int           // Beat number where this occurs
	Description     string        // Plot point description (150-250 字)
	TensionIncrease int           // Tension increase amount (+10 to +20)
}

// HandleKeyPlotPoint generates dramatic turning point description
//
// AC #10: 關鍵劇情節點處理
//   - 戲劇性轉折長度：150-250 字
//   - 轉折提升張力：Tension +10-20
//   - 關鍵節點類型：Inciting Incident, Midpoint, Second Plot Point, Climax
//
// Parameters:
//   - plotType: Type of plot point
//   - beat: Current beat number
//
// Returns:
//   - string: Plot point description text
func (a *NarrationAgent) HandleKeyPlotPoint(plotType PlotPointType, beat int) string {
	var description string

	switch plotType {
	case PlotPointIncitingIncident:
		// Act 1 End (Inciting Incident): 改變現狀的事件
		description = `一切都在這一刻改變了。

你原本以為這只是一個普通的夜晚，但現在你意識到，事情遠比你想像的要嚴重。眼前發生的一切徹底打破了你對這個地方的認知——這不是意外，不是巧合，而是某種更深層、更黑暗的真相的開端。

你感到一陣寒意爬上脊背。現在想要回頭已經太晚了，你已經被捲入了某個無法逃脫的漩渦之中。從這一刻起，一切都不同了。你必須面對這個殘酷的現實，並找出真相——否則，你可能永遠無法離開這裡。

恐懼感加劇了，你知道接下來的每一步都將充滿危險。`

	case PlotPointMidpoint:
		// Act 2 Midpoint: 中期轉折
		description = `真相的一角浮現了，但它帶來的不是希望，而是更深的絕望。

你原本以為自己開始理解這一切了，但現在你發現，你所知道的只是冰山一角。眼前的發現徹底顛覆了你之前的判斷——你一直走在錯誤的方向上，而真正的威脅比你想像的要可怕得多。

情況變得更加複雜了。你不再確定誰是敵人，誰是朋友；不再確定什麼是真實，什麼是幻覺。恐懼與困惑交織在一起，讓你的理智搖搖欲墜。

但你別無選擇。你必須繼續前進，必須找出更多的真相——即使這意味著你將面對更加可怕的事實。每一步都在加深你的恐懼，但你已經無路可退。`

	case PlotPointSecondPlotPoint:
		// Act 2 End (Second Plot Point): 最低點
		description = `這是最黑暗的時刻。

一切都崩潰了。你所依賴的、所相信的、所希望的——全都在這一刻破滅。你感到前所未有的絕望，彷彿被困在無盡的深淵之中，四周只有黑暗和恐懼。

你的身體在顫抖，不知道是因為寒冷還是恐懼。你的理智瀕臨崩潰的邊緣，眼前的一切都變得模糊而扭曲。你開始懷疑自己是否還能撐下去，是否還有任何希望。

但就在這絕望的深處，某種東西在你心中點燃了——是憤怒，是不甘，是最後的倔強。即使在這最黑暗的時刻，你仍然沒有放棄。你知道，只有熬過這個最低點，你才有可能找到最終的答案。

這是最後的考驗。要麼你在這裡徹底崩潰，要麼你從深淵中爬出來，變得更加強大。`

	case PlotPointClimax:
		// Act 3 Climax: 最終對抗
		description = `最終的時刻到來了。

所有的線索、所有的痛苦、所有的掙扎，都匯聚到了這一刻。你終於站在了真相的面前，面對著那個一直潛藏在黑暗中的存在。這是最後的對抗，是決定一切的時刻。

你的心跳加速，每一次呼吸都充滿了緊張與恐懼。你知道接下來的選擇將決定你的命運——是生是死，是自由還是永恆的困囿，一切都將在此刻揭曉。

恐懼達到了極點，但你已經沒有退路。你必須做出最終的決定，必須直面那個最可怕的真相。這是你的最後機會，也是唯一的機會。

一切都將在這裡結束——無論結局如何，至少你不會再有遺憾。你深吸一口氣，準備迎接最後的挑戰。`

	default:
		description = fmt.Sprintf(`關鍵時刻降臨。在 Beat %d，劇情發生了重大轉折，故事進入了新的階段。`, beat)
	}

	return description
}

// HallucinationSeverity represents the severity of hallucination based on SAN level
type HallucinationSeverity string

const (
	HallucinationMild   HallucinationSeverity = "mild"   // SAN 40-60: 輕度幻覺
	HallucinationModerate HallucinationSeverity = "moderate" // SAN 20-40: 中度幻覺
	HallucinationSevere HallucinationSeverity = "severe" // SAN <20: 重度幻覺
)

// HandleHallucinationChoice generates hallucination experience description
//
// AC #11: 幻覺選項處理
//   - 幻覺體驗描述長度：200-300 字
//   - 幻覺內容與低 SAN 相關（SAN 40-60: 輕度, 20-40: 中度, <20: 重度）
//   - 幻覺結束描述：「你晃了晃頭，眼前的景象消失了」
//   - SAN 額外下降：-2 到 -5（根據幻覺嚴重度）
//
// Parameters:
//   - currentSAN: Current SAN value
//
// Returns:
//   - description: Hallucination experience text
//   - sanPenalty: Additional SAN decrease amount
func (a *NarrationAgent) HandleHallucinationChoice(currentSAN int) (description string, sanPenalty int) {
	var severity HallucinationSeverity
	var hallucinationText string

	// Determine severity based on SAN level
	if currentSAN >= 40 && currentSAN <= 60 {
		severity = HallucinationMild
		sanPenalty = -2
	} else if currentSAN >= 20 && currentSAN < 40 {
		severity = HallucinationModerate
		sanPenalty = -3
	} else {
		severity = HallucinationSevere
		sanPenalty = -5
	}

	// Generate hallucination description based on severity
	switch severity {
	case HallucinationMild:
		// SAN 40-60: 輕度幻覺（視覺扭曲）
		hallucinationText = `世界開始扭曲。

牆壁彷彿在呼吸，緩緩地擴張又收縮。地板的紋路開始游動，像是有生命般蠕動著。燈光變得詭異，時而明亮時而昏暗，在你的視線中形成一個個跳動的光斑。

你眨了眨眼，試圖驅散這些奇怪的視覺，但它們依然存在。你知道這不對勁，但你的大腦似乎無法正確處理眼前的景象。色彩變得過於鮮豔或過於暗淡，邊緣變得模糊，整個世界都籠罩在一層薄霧之中。

這只是暫時的，你告訴自己。這只是疲勞和壓力造成的錯覺。`

	case HallucinationModerate:
		// SAN 20-40: 中度幻覺（聽覺異常、時間錯亂）
		hallucinationText = `現實開始崩解。

你聽見了不存在的聲音——竊竊私語、嘲笑聲、尖叫聲——它們從四面八方湧來，但當你試圖尋找聲源時，卻什麼也找不到。牆壁在說話，地板在哭泣，天花板在嘆息。整個空間都充滿了這些不該存在的聲音。

時間感也變得混亂了。你不確定現在是幾點，甚至不確定過了多久。一秒鐘可能是一小時，一小時可能只是一秒鐘。你的記憶開始錯亂，你不記得自己是怎麼到這裡的，也不記得接下來該做什麼。

牆上的影子開始移動，它們有自己的意志，不再受光源控制。你看見了不該看見的東西——扭曲的面孔、伸出的手、爬行的黑影。它們越來越真實，越來越接近。`

	case HallucinationSevere:
		// SAN <20: 重度幻覺（現實崩潰、人格分裂）
		hallucinationText = `一切都不真實了。

你站在那裡，看著世界在眼前徹底瓦解。牆壁融化成液體，地板裂開露出深不見底的深淵，天花板變成了蠕動的肉塊。你的身體也開始變化——你的手變得透明，你的影子獨立行動，你聽見自己的聲音但不是從自己的嘴裡發出的。

有另一個你在說話，在嘲笑你的軟弱。有無數個聲音在你腦海中爭吵，每一個都宣稱自己才是真正的你。你不再確定哪個是真實的，哪個是幻覺。也許你從一開始就是幻覺，也許這一切都是夢境，也許——

你看見了它。那個潛藏在現實背後的東西，那個一直在注視著你的存在。它沒有形狀，卻無處不在；它沒有聲音，卻在你的靈魂深處迴響。恐懼已經不足以形容你現在的感受——這是超越恐懼的體驗，是理智徹底崩潰的瞬間。

你尖叫著，但發不出聲音。你逃跑著，但雙腿無法移動。你崩潰著，靈魂在無盡的瘋狂中沉淪。`
	}

	// Hallucination ending (standard across all severities)
	hallucinationEnding := `

你晃了晃頭，用力地眨了幾次眼睛。

漸漸地，眼前的景象開始消失。那些扭曲的視覺、詭異的聲音、不真實的感覺——它們慢慢地褪去，現實重新浮現。你意識到剛才看到的一切都不是真的，那只是你破碎的理智製造出來的幻象。

但那種恐懼感依然揮之不去。你知道，下一次幻覺可能隨時會來臨。你的理智正在一點點流失，而你似乎無力阻止。`

	description = hallucinationText + hallucinationEnding
	return description, sanPenalty
}
