package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 6-6: Choice Agent Implementation
// ==========================================================================

// ChoiceAgent generates player choice options with risk assessment
//
// Responsibilities:
//  1. Generate 2-4 meaningful choice options
//  2. Assess risk level for each option (internal use)
//  3. Add subtle rule hints based on difficulty
//  4. Insert hallucination options when SAN is low
//
// Design Philosophy:
//  - Template-driven generation (fast path, <50ms)
//  - LLM generation as fallback (Fast Model, <450ms)
//  - Risk assessment is internal only (not shown to player)
//  - Hallucination options blend seamlessly with real options
type ChoiceAgent struct {
	*BaseAgentImpl
	templates map[SceneType][]OptionTemplate
	rng       *rand.Rand
}

// NewChoiceAgent creates a new ChoiceAgent with BaseAgentImpl pattern
func NewChoiceAgent(config AgentConfig) *ChoiceAgent {
	// Set defaults if not provided
	if config.Name == "" {
		config.Name = "ChoiceAgent"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second // Default 30s for consistency
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	ca := &ChoiceAgent{
		BaseAgentImpl: NewBaseAgentImpl(config),
		templates:     loadDefaultTemplates(),
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return ca
}

// InvokeGenerate generates choice options for the player
//
// AC #1: Returns 2-4 options with ≤ 15 chars each
// AC #2: Calculates risk levels (internal use only)
// AC #3: May insert hallucination option if SAN < 20
// AC #4: Adds rule hints based on difficulty
// AC #5: Uses Fast Model for <500ms response time
// AC #6: Options match current scene context
//
// Priority:
//  1. Template generation (fast, <50ms)
//  2. LLM generation (Fast Model, <450ms)
//  3. Fallback to generic templates
func (ca *ChoiceAgent) InvokeGenerate(ctx context.Context, request *ChoiceRequest) (*ChoiceResponse, error) {
	// Story 10-8 AC1: Log agent invocation
	logger.Debug("ChoiceAgent invoked", map[string]interface{}{
		"scene_type": request.SceneType,
		"difficulty": request.Difficulty,
		"current_san": request.PlayerSAN,
	})

	// 1. Try template-driven generation (fast path)
	templateOptions, err := ca.generateFromTemplate(request)
	if err == nil && len(templateOptions) >= 2 {
		logger.Debug("ChoiceAgent using template generation", map[string]interface{}{
			"num_options": len(templateOptions),
		})
		return ca.enrichWithRiskAndHints(templateOptions, request), nil
	}

	// 2. Fall back to LLM generation (Fast Model)
	logger.Debug("ChoiceAgent falling back to LLM generation", nil)
	llmOptions, err := ca.generateFromLLM(ctx, request)
	if err != nil {
		logger.Warn("ChoiceAgent LLM generation failed, using fallback", map[string]interface{}{
			"error": err.Error(),
		})
		// 3. Double fallback: use generic templates
		return ca.generateFallbackOptions(request), nil
	}

	// Enrich with risk assessment and hints
	logger.Debug("ChoiceAgent completed generation", map[string]interface{}{
		"num_options": len(llmOptions),
	})
	return ca.enrichWithRiskAndHints(llmOptions, request), nil
}

// generateFromTemplate generates options using templates (fast path)
//
// AC #5: Template generation < 50ms
// AC #6: Options match scene type
func (ca *ChoiceAgent) generateFromTemplate(request *ChoiceRequest) ([]Option, error) {
	templates, ok := ca.templates[request.SceneType]
	if !ok || len(templates) == 0 {
		return nil, fmt.Errorf("no templates for scene type: %v", request.SceneType)
	}

	// Select 2-4 templates randomly
	numOptions := 2 + ca.rng.Intn(3) // 2-4
	if numOptions > len(templates) {
		numOptions = len(templates)
	}

	// Randomly select templates
	selected := ca.randomSelect(templates, numOptions)

	options := make([]Option, 0, numOptions)
	for i, tmpl := range selected {
		text := ca.fillTemplate(tmpl, request)
		// AC #1: Ensure ≤ 15 chars
		if len([]rune(text)) > 15 {
			text = string([]rune(text)[:15])
		}
		options = append(options, Option{
			Index:       i + 1,
			Text:        text,
			Description: tmpl.Description,
		})
	}

	return options, nil
}

// fillTemplate fills template placeholders with variants
func (ca *ChoiceAgent) fillTemplate(tmpl OptionTemplate, request *ChoiceRequest) string {
	text := tmpl.Text

	if len(tmpl.Variants) == 0 {
		return text
	}

	// Select random variant
	variant := tmpl.Variants[ca.rng.Intn(len(tmpl.Variants))]

	// Replace placeholders
	text = strings.ReplaceAll(text, "{object}", variant)
	text = strings.ReplaceAll(text, "{location}", variant)
	text = strings.ReplaceAll(text, "{character}", variant)
	text = strings.ReplaceAll(text, "{target}", variant)
	text = strings.ReplaceAll(text, "{topic}", variant)

	return text
}

// randomSelect randomly selects n items from slice
func (ca *ChoiceAgent) randomSelect(templates []OptionTemplate, n int) []OptionTemplate {
	if n >= len(templates) {
		return templates
	}

	// Fisher-Yates shuffle
	shuffled := make([]OptionTemplate, len(templates))
	copy(shuffled, templates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := ca.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:n]
}

// generateFromLLM generates options using LLM (Fast Model)
//
// AC #1: Returns 2-4 options
// AC #5: Uses Fast Model (<450ms)
func (ca *ChoiceAgent) generateFromLLM(ctx context.Context, request *ChoiceRequest) ([]Option, error) {
	prompt := ca.buildChoicePrompt(request)

	// Call LLM via BaseAgentImpl's InvokeWithRetry
	response, err := ca.BaseAgentImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		return ca.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.7,
			"max_tokens":  500,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Type assert to string
	responseStr, ok := response.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	// Parse JSON response
	options, err := ca.parseChoiceResponse(responseStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return options, nil
}

// buildChoicePrompt builds the LLM prompt for choice generation
func (ca *ChoiceAgent) buildChoicePrompt(request *ChoiceRequest) string {
	var tensionDesc string
	if request.TensionLevel < 30 {
		tensionDesc = "LOW"
	} else if request.TensionLevel < 70 {
		tensionDesc = "MEDIUM"
	} else {
		tensionDesc = "HIGH"
	}

	prompt := fmt.Sprintf(`你是「規則怪談」遊戲的選項生成 Agent。你的職責是根據故事情境生成 2-4 個玩家選項。

**生成標準：**
1. 選項文字 ≤ 15 字（繁體中文）
2. 選項之間有明確差異
3. 符合當前場景與張力等級
4. 自然融入故事情境

**輸出格式（JSON）：**
{
  "options": [
    {"index": 1, "text": "檢查門", "description": "檢查房間的門"},
    {"index": 2, "text": "觸摸鏡子", "description": "嘗試觸摸牆上的鏡子"},
    {"index": 3, "text": "離開房間", "description": "轉身離開"}
  ]
}

=== 故事情境 ===
%s

=== 場景類型 ===
%s

=== 張力等級 ===
%s

=== 難度 ===
%s

請生成 2-4 個選項並返回 JSON 格式。`, request.StoryContext, request.SceneType.String(), tensionDesc, request.Difficulty)

	return prompt
}

// parseChoiceResponse parses LLM JSON response
func (ca *ChoiceAgent) parseChoiceResponse(raw string) ([]Option, error) {
	// Clean up response
	content := strings.TrimSpace(raw)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Parse JSON
	var response struct {
		Options []struct {
			Index       int    `json:"index"`
			Text        string `json:"text"`
			Description string `json:"description"`
		} `json:"options"`
	}

	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	if len(response.Options) < 2 || len(response.Options) > 4 {
		return nil, fmt.Errorf("invalid option count: %d (expected 2-4)", len(response.Options))
	}

	// Convert to Option slice
	options := make([]Option, 0, len(response.Options))
	for _, opt := range response.Options {
		// AC #1: Enforce ≤ 15 chars
		text := opt.Text
		if len([]rune(text)) > 15 {
			text = string([]rune(text)[:15])
		}

		options = append(options, Option{
			Index:       opt.Index,
			Text:        text,
			Description: opt.Description,
		})
	}

	return options, nil
}

// generateFallbackOptions generates generic fallback options
//
// Used when both template and LLM generation fail.
func (ca *ChoiceAgent) generateFallbackOptions(request *ChoiceRequest) *ChoiceResponse {
	fallbackOptions := []Option{
		{Index: 1, Text: "繼續前進", Description: "Continue forward"},
		{Index: 2, Text: "仔細觀察", Description: "Observe carefully"},
		{Index: 3, Text: "原地等待", Description: "Wait in place"},
	}

	return ca.enrichWithRiskAndHints(fallbackOptions, request)
}

// enrichWithRiskAndHints adds risk assessment, hints, and hallucinations
//
// AC #2: Calculate risk levels
// AC #3: Add hallucination option if applicable
// AC #4: Add rule hints based on difficulty
func (ca *ChoiceAgent) enrichWithRiskAndHints(options []Option, request *ChoiceRequest) *ChoiceResponse {
	response := &ChoiceResponse{
		Options:            make([]Option, 0, len(options)+1),
		RiskLevels:         make(map[int]RiskLevel),
		HallucinationFlags: make(map[int]bool),
	}

	// Process existing options
	for _, opt := range options {
		// AC #2: Calculate risk level
		risk := ca.CalculateRisk(opt, request.ActiveRules)
		response.RiskLevels[opt.Index] = risk

		// AC #4: Add rule hint if needed
		if risk > RiskSafe {
			opt.Text = ca.AddRuleHint(opt, request.ActiveRules, request.Difficulty, risk)
		}

		response.Options = append(response.Options, opt)
		response.HallucinationFlags[opt.Index] = false
	}

	// AC #3: Add hallucination option if applicable
	// Story 7.5 AC5: Add 1-2 hallucination options when SAN < 20
	// Code Review Fix 7-5-2: Use ControlLevel as single source of truth
	halluCount := game.GetControlLevel(request.PlayerSAN).GetHallucinationCount()
	for i := 0; i < halluCount; i++ {
		halluOpt := ca.generateHallucinationOption(request, i)
		response.Options = append(response.Options, halluOpt)
		response.RiskLevels[halluOpt.Index] = RiskLethal // Story 7.5 AC5: Severe consequences
		response.HallucinationFlags[halluOpt.Index] = true
	}

	return response
}

// CalculateRisk calculates the risk level for an option
//
// AC #2: Risk assessment based on ActiveRules check
//
// Logic:
//  - Matches option keywords against rule trigger keywords
//  - Returns highest risk level from matching rules
//  - Lethal if rule is fatal
//  - Danger if total damage ≥ 50
//  - Warning if total damage ≥ 20
//  - Safe otherwise
func (ca *ChoiceAgent) CalculateRisk(option Option, rules []HiddenRule) RiskLevel {
	maxRisk := RiskSafe

	for _, rule := range rules {
		if ca.matchesRuleTrigger(option.Text, rule) {
			ruleRisk := ca.getRuleRisk(rule)
			if ruleRisk > maxRisk {
				maxRisk = ruleRisk
			}
		}
	}

	return maxRisk
}

// matchesRuleTrigger checks if option text matches rule trigger keywords
//
// AC #2: Enhanced matching logic for Chinese text
// Uses fuzzy matching to handle variations like "看向鏡子" vs "看鏡子"
func (ca *ChoiceAgent) matchesRuleTrigger(optionText string, rule HiddenRule) bool {
	optionLower := strings.ToLower(optionText)

	for _, keyword := range rule.TriggerKeywords {
		keywordLower := strings.ToLower(keyword)

		// Exact match
		if strings.Contains(optionLower, keywordLower) {
			return true
		}

		// Fuzzy match for Chinese text: check if all characters in keyword appear in option
		// This handles cases like "看向鏡子" matching "看鏡子"
		if ca.fuzzyContainsChinese(optionText, keyword) {
			return true
		}
	}
	return false
}

// fuzzyContainsChinese checks if all characters in keyword appear in text (for Chinese)
func (ca *ChoiceAgent) fuzzyContainsChinese(text, keyword string) bool {
	// Only apply fuzzy matching for Chinese characters
	if !ca.containsChinese(keyword) {
		return false
	}

	textRunes := []rune(text)
	keywordRunes := []rune(keyword)

	// Check if all keyword characters appear in text in order
	textIdx := 0
	for _, kr := range keywordRunes {
		found := false
		for textIdx < len(textRunes) {
			if textRunes[textIdx] == kr {
				found = true
				textIdx++
				break
			}
			textIdx++
		}
		if !found {
			return false
		}
	}
	return true
}

// containsChinese checks if string contains Chinese characters
func (ca *ChoiceAgent) containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// getRuleRisk determines risk level from rule punishment
func (ca *ChoiceAgent) getRuleRisk(rule HiddenRule) RiskLevel {
	if rule.Punishment.IsFatal {
		return RiskLethal
	}

	totalDamage := rule.Punishment.HPDamage + rule.Punishment.SANDamage
	if totalDamage >= 50 {
		return RiskDanger
	} else if totalDamage >= 20 {
		return RiskWarning
	}

	return RiskSafe
}

// AddRuleHint adds subtle hint to option text based on difficulty
//
// AC #4: Hint subtlety adjusted by difficulty
//  - Easy: Direct hint (e.g., "這看起來很危險")
//  - Normal: Metaphor hint (e.g., "有些不對勁")
//  - Hard/Hell: No hint
//
// Note: Risk level is NOT shown to player
//
// CRITICAL FIX (H-4): Only add hints to risky options (risk > RiskSafe)
func (ca *ChoiceAgent) AddRuleHint(option Option, rules []HiddenRule, difficulty string, risk RiskLevel) string {
	// Safe options should never get hints
	if risk == RiskSafe {
		return option.Text
	}

	if difficulty == "hell" || difficulty == "hard" {
		// Hell/Hard: No hints even for risky options
		return option.Text
	}

	// Find matching rule with highest risk
	var matchedRule *HiddenRule
	for i := range rules {
		if ca.matchesRuleTrigger(option.Text, rules[i]) {
			if matchedRule == nil || ca.getRuleRisk(rules[i]) > ca.getRuleRisk(*matchedRule) {
				matchedRule = &rules[i]
			}
		}
	}

	if matchedRule == nil {
		return option.Text
	}

	// Generate hint based on difficulty
	var hint string
	if difficulty == "easy" {
		hint = matchedRule.DirectHint
	} else { // normal
		hint = matchedRule.MetaphorHint
	}

	if hint == "" {
		return option.Text
	}

	// Naturally integrate hint
	return fmt.Sprintf("%s（%s）", option.Text, hint)
}

// shouldAddHallucination determines if hallucination option should be added
//
// AC #3: Hallucination probability based on SAN
//  - SAN ≥ 60: 0%
//  - SAN 40-59: 10%
//  - SAN 20-39: 30%
//  - SAN < 20: 50%
func (ca *ChoiceAgent) shouldAddHallucination(san int) bool {
	if san >= 60 {
		return false
	} else if san >= 40 {
		return ca.rng.Float64() < 0.1 // 10%
	} else if san >= 20 {
		return ca.rng.Float64() < 0.3 // 30%
	} else {
		return ca.rng.Float64() < 0.5 // 50%
	}
}

// getHallucinationCount is deprecated.
// Code Review Fix 7-5-2: Use game.GetControlLevel(san).GetHallucinationCount() instead.
// This method is kept for backward compatibility but delegates to the canonical implementation.
func (ca *ChoiceAgent) getHallucinationCount(san int) int {
	return game.GetControlLevel(san).GetHallucinationCount()
}

// generateHallucinationOption generates a hallucination option
//
// Story 7.5 AC5: Hallucination options that look reasonable but are dangerous.
// Types:
//  - False Safety: Non-existent safe option that looks appealing
//  - Induced Danger: Follows auditory/visual hallucination
//  - Absurd Action: Seemingly reasonable but actually dangerous
//  - Paranoid Defense: Overly defensive action against non-threat
//
// Story 7.5 AC5: Hallucinations are NOT marked to player (indistinguishable from real)
// Story 7.5 AC5: Choosing hallucination causes severe HP/SAN loss
func (ca *ChoiceAgent) generateHallucinationOption(request *ChoiceRequest, variant int) Option {
	// Enhanced hallucination templates that look more reasonable
	hallucinationSets := [][]struct {
		text string
		desc string
	}{
		// Set 0: False safety
		{
			{text: "找到出口離開", desc: "[幻覺] 虛假的出口"},
			{text: "看到安全的房間", desc: "[幻覺] 不存在的安全區"},
			{text: "發現救援信號", desc: "[幻覺] 虛假的希望"},
		},
		// Set 1: Induced danger
		{
			{text: "跟隨呼喊聲", desc: "[幻覺] 幻聽引導"},
			{text: "追逐熟悉身影", desc: "[幻覺] 視覺欺騙"},
			{text: "靠近發光物體", desc: "[幻覺] 誘導陷阱"},
		},
		// Set 2: Absurd but tempting
		{
			{text: "服用地上的藥", desc: "[幻覺] 危險物質"},
			{text: "相信眼前的人", desc: "[幻覺] 虛假信任"},
			{text: "打開神秘的盒子", desc: "[幻覺] 未知危險"},
		},
		// Set 3: Paranoid defense
		{
			{text: "攻擊可疑對象", desc: "[幻覺] 錯誤目標"},
			{text: "立即逃離現場", desc: "[幻覺] 過度反應"},
			{text: "砸碎鏡子", desc: "[幻覺] 破壞性行為"},
		},
	}

	// Select set based on scene type and variant
	setIndex := (int(request.SceneType) + variant) % len(hallucinationSets)
	optionSet := hallucinationSets[setIndex]

	// Randomly select from the set
	option := optionSet[ca.rng.Intn(len(optionSet))]

	// Generate index that blends with real options
	// Use a unique but not obviously special index
	optionIndex := len(request.ActiveRules) + 50 + variant

	return Option{
		Index:       optionIndex,
		Text:        option.text,
		Description: option.desc, // Internal only - NOT shown to player
	}
}

// loadDefaultTemplates loads default option templates
//
// AC #5: Template library for fast generation (<50ms)
func loadDefaultTemplates() map[SceneType][]OptionTemplate {
	templates := make(map[SceneType][]OptionTemplate)

	// Explore templates
	templates[SceneExplore] = []OptionTemplate{
		{ID: "explore_door", Text: "檢查{object}", Variants: []string{"門", "窗", "櫃子"}, BaseRisk: RiskSafe},
		{ID: "explore_investigate", Text: "仔細調查{location}", Variants: []string{"牆壁", "地板", "角落"}, BaseRisk: RiskSafe},
		{ID: "explore_touch", Text: "觸摸{object}", Variants: []string{"鏡子", "畫像", "雕像"}, BaseRisk: RiskWarning},
		{ID: "explore_leave", Text: "離開房間", Variants: nil, BaseRisk: RiskSafe},
	}

	// Dialogue templates
	templates[SceneDialogue] = []OptionTemplate{
		{ID: "dialogue_ask", Text: "詢問{topic}", Variants: []string{"名字", "這裡", "出口"}, BaseRisk: RiskSafe},
		{ID: "dialogue_trust", Text: "相信對方", Variants: nil, BaseRisk: RiskWarning},
		{ID: "dialogue_doubt", Text: "表示懷疑", Variants: nil, BaseRisk: RiskSafe},
		{ID: "dialogue_leave", Text: "結束對話", Variants: nil, BaseRisk: RiskSafe},
	}

	// Combat templates
	templates[SceneCombat] = []OptionTemplate{
		{ID: "combat_attack", Text: "攻擊{target}", Variants: []string{"正面", "側面", "弱點"}, BaseRisk: RiskDanger},
		{ID: "combat_defend", Text: "防禦", Variants: nil, BaseRisk: RiskWarning},
		{ID: "combat_flee", Text: "逃跑", Variants: nil, BaseRisk: RiskSafe},
	}

	// Escape templates
	templates[SceneEscape] = []OptionTemplate{
		{ID: "escape_run", Text: "狂奔", Variants: nil, BaseRisk: RiskWarning},
		{ID: "escape_hide", Text: "躲藏", Variants: nil, BaseRisk: RiskSafe},
		{ID: "escape_sneak", Text: "悄悄移動", Variants: nil, BaseRisk: RiskSafe},
	}

	return templates
}
