package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 6-7: Judge Agent Implementation
// ==========================================================================

// JudgeAgent judges player choices and determines consequences
//
// Responsibilities:
//  1. Check if player choice violates hidden rules
//  2. Calculate HP/SAN changes with warning system
//  3. Determine death conditions
//  4. Route to appropriate next action
//
// Design Philosophy:
//  - Rule Engine first (deterministic, fast <50ms)
//  - Fast Model for ambiguous cases (<450ms)
//  - Robust fallback when LLM fails
type JudgeAgent struct {
	*BaseAgentImpl
}

// NewJudgeAgent creates a new JudgeAgent with BaseAgentImpl pattern
func NewJudgeAgent(config AgentConfig) *JudgeAgent {
	// Set defaults if not provided
	if config.Name == "" {
		config.Name = "JudgeAgent"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	ja := &JudgeAgent{
		BaseAgentImpl: NewBaseAgentImpl(config),
	}

	return ja
}

// InvokeJudge judges a player's choice and returns the result
//
// AC #1: Check all applicable rules (sorted by priority)
// AC #2: Calculate impact level and state changes
// AC #3: Generate death reason if lethal
// AC #6: Use Fast Model for response time < 500ms
//
// Priority:
//  1. Rule engine check (fast, deterministic)
//  2. If clear violation, return immediately
//  3. If ambiguous, use Fast Model
//  4. Fallback to rule engine on LLM failure
func (ja *JudgeAgent) InvokeJudge(ctx context.Context, request *JudgeRequest) (*JudgeResponse, error) {
	// Story 10-8 AC1: Log agent invocation
	logger.Debug("JudgeAgent invoked", map[string]interface{}{
		"player_choice": request.PlayerChoice,
		"num_active_rules": len(request.ActiveRules),
		"current_hp": request.GameState.HP,
		"current_san": request.GameState.SAN,
	})

	// 1. Rule engine check (fast path)
	violations := ja.CheckRuleViolation(request.PlayerChoice, request.ActiveRules)
	impactLevel := ja.DetermineImpactLevel(violations)

	// Story 10-8 AC1: Log rule check results
	logger.Debug("JudgeAgent rule check completed", map[string]interface{}{
		"num_violations": len(violations),
		"impact_level": impactLevel.String(),
	})

	// 2. Calculate state changes
	stateChanges := ja.CalculateStateChanges(
		violations,
		request.GameState.Difficulty,
		request.GameState.RuleWarnings,
	)

	// 3. If clear violation (Moderate or higher), return immediately
	if impactLevel >= ImpactModerate {
		deathReason := ""
		if impactLevel == ImpactLethal && len(violations) > 0 {
			// H-7 FIX: Find the first FATAL violation for death reason
			var fatalViolation *RuleViolation
			for i := range violations {
				if violations[i].IsFatal {
					fatalViolation = &violations[i]
					break
				}
			}
			if fatalViolation != nil {
				deathReason = ja.GenerateDeathReason(*fatalViolation, request.PlayerChoice)
			}
		}

		return &JudgeResponse{
			RulesViolated:         violations,
			ImpactLevel:           impactLevel,
			SuggestedStateChanges: stateChanges,
			DeathReason:           deathReason,
			NextAction:            ja.determineNextAction(impactLevel),
			Reasoning:             ja.buildReasoningText(violations),
		}, nil
	}

	// 4. Ambiguous case - use Fast Model for additional judgment
	if ja.Config.LLMClient != nil {
		llmResult, err := ja.callFastModel(ctx, request)
		if err != nil {
			log.Printf("[JudgeAgent] Fast Model failed, using rule engine result: %v", err)
		} else {
			// Merge LLM results with rule engine results
			return ja.mergeResults(violations, impactLevel, stateChanges, llmResult), nil
		}
	}

	// 5. Fallback: use rule engine result (no violation or minor)
	return &JudgeResponse{
		RulesViolated:         violations,
		ImpactLevel:           impactLevel,
		SuggestedStateChanges: stateChanges,
		DeathReason:           "",
		NextAction:            ActionContinueStory,
		Reasoning:             "No significant rule violations detected",
	}, nil
}

// CheckRuleViolation checks if a choice violates any active rules
//
// AC #1: Check all applicable rules
// AC #4: Sort by priority (Scene > Time > Behavior > State)
//
// Story 7.2 Enhancement:
//   - Supports instant death without warning for Hell mode
//   - Handles logic chain validation (via matchesRule)
func (ja *JudgeAgent) CheckRuleViolation(choice string, rules []JudgeHiddenRule) []RuleViolation {
	violations := []RuleViolation{}

	for _, rule := range rules {
		if ja.matchesRule(choice, rule) {
			violation := RuleViolation{
				RuleID:    rule.ID,
				RuleName:  rule.Name,
				RuleType:  rule.Type,
				Severity:  ja.determineSeverity(rule),
				HPDamage:  rule.Punishment.HPDamage,
				SANDamage: rule.Punishment.SANDamage,
				IsFatal:   rule.Punishment.IsFatal,
				Reason:    fmt.Sprintf("觸發條件：%s", rule.TriggerCondition),
			}
			violations = append(violations, violation)
		}
	}

	// AC #4: Sort by rule type priority
	sort.Slice(violations, func(i, j int) bool {
		return GetRulePriority(violations[i].RuleType) > GetRulePriority(violations[j].RuleType)
	})

	return violations
}

// matchesRule checks if a choice matches a rule's trigger conditions
func (ja *JudgeAgent) matchesRule(choice string, rule JudgeHiddenRule) bool {
	choiceLower := strings.ToLower(choice)

	// 1. Keyword matching
	for _, keyword := range rule.TriggerKeywords {
		if strings.Contains(choiceLower, strings.ToLower(keyword)) {
			return true
		}
	}

	// 2. Regex matching (if defined)
	if rule.TriggerRegex != "" {
		matched, err := regexp.MatchString(rule.TriggerRegex, choice)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// DetermineImpactLevel determines the overall impact level
//
// AC #2: Calculate impact level from violations
// AC #3: Return Lethal for fatal rules
// AC #4: Take highest severity when multiple violations
func (ja *JudgeAgent) DetermineImpactLevel(violations []RuleViolation) ImpactLevel {
	if len(violations) == 0 {
		return ImpactNone
	}

	// AC #3: Check for fatal violations first
	for _, v := range violations {
		if v.IsFatal {
			return ImpactLethal
		}
	}

	// AC #4: Take the highest severity
	maxSeverity := ImpactNone
	for _, v := range violations {
		if v.Severity > maxSeverity {
			maxSeverity = v.Severity
		}
	}

	return maxSeverity
}

// determineSeverity determines the severity of a single rule
func (ja *JudgeAgent) determineSeverity(rule JudgeHiddenRule) ImpactLevel {
	// AC #3: Fatal rules are always Lethal
	if rule.Punishment.IsFatal {
		return ImpactLethal
	}

	// Calculate total damage
	totalDamage := rule.Punishment.HPDamage + rule.Punishment.SANDamage

	// AC #2: Classify by damage amount
	if totalDamage >= 50 {
		return ImpactMajor
	} else if totalDamage >= 20 {
		return ImpactModerate
	} else if totalDamage > 0 {
		return ImpactMinor
	}

	return ImpactNone
}

// CalculateStateChanges calculates HP/SAN changes with warning system
//
// AC #2: Calculate state changes based on violations
// AC #2: Apply difficulty multiplier (easy 0.8, normal 1.0, hard 1.0, hell 1.2)
// AC #2: Apply warning system (easy 2, normal 1, hard 1, hell 0)
//
// Story 7.2 Enhancement:
//   - Hell mode (maxWarnings=0) triggers instant death for fatal rules
//   - No warning given, death is immediate per AC3
func (ja *JudgeAgent) CalculateStateChanges(
	violations []RuleViolation,
	difficulty string,
	currentWarnings map[string]int,
) StateChanges {
	changes := StateChanges{
		WarningsRemaining: make(map[string]int),
	}

	if len(violations) == 0 {
		return changes
	}

	// AC #2: Difficulty multiplier
	difficultyMultiplier := ja.getDifficultyMultiplier(difficulty)
	maxWarnings := ja.getMaxWarnings(difficulty)
	isHellMode := strings.ToLower(difficulty) == "hell"

	for _, v := range violations {
		// Initialize warnings if not set
		if _, exists := currentWarnings[v.RuleID]; !exists {
			currentWarnings[v.RuleID] = maxWarnings
		}

		remainingWarnings := currentWarnings[v.RuleID]

		// Story 7.2 AC3: Hell mode instant death for fatal rules
		if isHellMode && v.IsFatal {
			// Instant death with no warning - mark as immediate consequence
			changes.WarningsRemaining[v.RuleID] = 0
			// Fatal means complete HP/SAN loss
			changes.HP = -100 // Ensure death
			changes.SAN = -100
			continue
		}

		// AC #2: Warning system
		if remainingWarnings > 0 && !v.IsFatal {
			// Has warnings remaining - apply half damage
			changes.WarningsRemaining[v.RuleID] = remainingWarnings - 1
			changes.HP -= int(float64(v.HPDamage) * 0.5 * difficultyMultiplier)
			changes.SAN -= int(float64(v.SANDamage) * 0.5 * difficultyMultiplier)
		} else {
			// No warnings or fatal - apply full damage
			changes.WarningsRemaining[v.RuleID] = 0
			changes.HP -= int(float64(v.HPDamage) * difficultyMultiplier)
			changes.SAN -= int(float64(v.SANDamage) * difficultyMultiplier)
		}
	}

	return changes
}

// getDifficultyMultiplier returns the damage multiplier for difficulty
func (ja *JudgeAgent) getDifficultyMultiplier(difficulty string) float64 {
	switch strings.ToLower(difficulty) {
	case "easy":
		return 0.8
	case "normal":
		return 1.0
	case "hard":
		return 1.0
	case "hell":
		return 1.2
	default:
		return 1.0
	}
}

// getMaxWarnings returns the maximum warnings for difficulty
func (ja *JudgeAgent) getMaxWarnings(difficulty string) int {
	switch strings.ToLower(difficulty) {
	case "easy":
		return 2
	case "normal":
		return 1
	case "hard":
		return 1
	case "hell":
		return 0
	default:
		return 1
	}
}

// GenerateDeathReason generates a death reason for a fatal violation
//
// AC #3: Generate death reason when ImpactLevel = Lethal
func (ja *JudgeAgent) GenerateDeathReason(violation RuleViolation, playerChoice string) string {
	if !violation.IsFatal {
		return ""
	}

	return fmt.Sprintf("你違反了規則「%s」：%s。%s導致了致命的後果。",
		violation.RuleName,
		violation.Reason,
		playerChoice,
	)
}

// determineNextAction determines the next action based on impact level
//
// AC #5: Route based on impact level
func (ja *JudgeAgent) determineNextAction(impactLevel ImpactLevel) NextActionType {
	switch impactLevel {
	case ImpactNone, ImpactMinor:
		// AC #5: Continue story for minor/no impact
		return ActionContinueStory
	case ImpactModerate, ImpactMajor:
		// AC #5: Apply damage and continue
		return ActionApplyDamage
	case ImpactLethal:
		// AC #5: Trigger death flow
		return ActionTriggerDeath
	default:
		return ActionContinueStory
	}
}

// buildReasoningText builds human-readable reasoning text
func (ja *JudgeAgent) buildReasoningText(violations []RuleViolation) string {
	if len(violations) == 0 {
		return "No rule violations detected"
	}

	var reasoning strings.Builder
	reasoning.WriteString(fmt.Sprintf("Detected %d rule violation(s): ", len(violations)))

	for i, v := range violations {
		if i > 0 {
			reasoning.WriteString(", ")
		}
		reasoning.WriteString(fmt.Sprintf("%s (%s)", v.RuleName, v.Severity))
	}

	return reasoning.String()
}

// callFastModel calls the Fast Model for ambiguous judgment
//
// AC #6: Use Fast Model for < 500ms response
func (ja *JudgeAgent) callFastModel(ctx context.Context, request *JudgeRequest) (*LLMJudgeResponse, error) {
	prompt := ja.buildJudgePrompt(request)

	// Call LLM via BaseAgentImpl's InvokeWithRetry
	response, err := ja.BaseAgentImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		return ja.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.3, // Low temperature for consistent judgment
			"max_tokens":  500,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("LLM judgment failed: %w", err)
	}

	// Type assert to string
	responseStr, ok := response.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	// Parse JSON response
	llmResult, err := ja.parseJudgeResponse(responseStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return llmResult, nil
}

// buildJudgePrompt builds the LLM prompt for judgment
func (ja *JudgeAgent) buildJudgePrompt(request *JudgeRequest) string {
	var sb strings.Builder

	sb.WriteString("你是「規則怪談」遊戲的裁判 Agent。你的職責是判定玩家的行動是否觸發潛規則。\n\n")

	sb.WriteString("**判定標準：**\n")
	sb.WriteString("1. 檢查玩家行動是否匹配規則的觸發條件（關鍵詞、行為模式）\n")
	sb.WriteString("2. 評估影響等級（None, Minor, Moderate, Major, Lethal）\n")
	sb.WriteString("3. 給出理由（為何觸發或未觸發）\n\n")

	sb.WriteString("**輸出格式（JSON）：**\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"rules_violated\": [\"R-01\", \"R-03\"],\n")
	sb.WriteString("  \"impact_level\": \"Moderate\",\n")
	sb.WriteString("  \"reasoning\": \"玩家凝視鏡子超過3秒，觸發 R-01 倒影殺手規則\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("=== 玩家行動 ===\n")
	sb.WriteString(request.PlayerChoice)
	sb.WriteString("\n\n")

	sb.WriteString("=== 當前狀態 ===\n")
	sb.WriteString(fmt.Sprintf("HP: %d\n", request.GameState.HP))
	sb.WriteString(fmt.Sprintf("SAN: %d\n", request.GameState.SAN))
	sb.WriteString(fmt.Sprintf("位置：%s\n", request.GameState.CurrentScene))
	sb.WriteString(fmt.Sprintf("難度：%s\n", request.GameState.Difficulty))
	sb.WriteString("\n")

	sb.WriteString("=== 活躍規則 ===\n")
	for _, rule := range request.ActiveRules {
		sb.WriteString(fmt.Sprintf("【%s】%s\n", rule.ID, rule.Name))
		sb.WriteString(fmt.Sprintf("- 觸發條件：%s\n", rule.TriggerCondition))
		sb.WriteString(fmt.Sprintf("- 觸發關鍵詞：%s\n", strings.Join(rule.TriggerKeywords, ", ")))
		sb.WriteString("\n")
	}

	sb.WriteString("請判定此行動並返回 JSON 格式結果。")

	return sb.String()
}

// parseJudgeResponse parses the LLM's JSON response
func (ja *JudgeAgent) parseJudgeResponse(raw string) (*LLMJudgeResponse, error) {
	// Try to extract JSON from markdown code block
	jsonStr := raw
	if start := strings.Index(raw, "```json"); start != -1 {
		jsonStr = raw[start+7:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	} else if start := strings.Index(raw, "```"); start != -1 {
		jsonStr = raw[start+3:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	var result LLMJudgeResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &result, nil
}

// mergeResults merges rule engine and LLM results
func (ja *JudgeAgent) mergeResults(
	ruleViolations []RuleViolation,
	ruleImpactLevel ImpactLevel,
	ruleStateChanges StateChanges,
	llmResult *LLMJudgeResponse,
) *JudgeResponse {
	// If rule engine found violations, prioritize those
	if len(ruleViolations) > 0 {
		deathReason := ""
		if ruleImpactLevel == ImpactLethal {
			// H-7 FIX: Find the first FATAL violation for death reason
			var fatalViolation *RuleViolation
			for i := range ruleViolations {
				if ruleViolations[i].IsFatal {
					fatalViolation = &ruleViolations[i]
					break
				}
			}
			if fatalViolation != nil {
				deathReason = ja.GenerateDeathReason(*fatalViolation, "")
			}
		}

		return &JudgeResponse{
			RulesViolated:         ruleViolations,
			ImpactLevel:           ruleImpactLevel,
			SuggestedStateChanges: ruleStateChanges,
			DeathReason:           deathReason,
			NextAction:            ja.determineNextAction(ruleImpactLevel),
			Reasoning:             llmResult.Reasoning,
		}
	}

	// LLM found violations that rule engine missed
	llmImpactLevel := ja.parseImpactLevel(llmResult.ImpactLevel)

	return &JudgeResponse{
		RulesViolated:         []RuleViolation{}, // LLM doesn't provide detailed violations
		ImpactLevel:           llmImpactLevel,
		SuggestedStateChanges: StateChanges{}, // No specific changes without rule details
		DeathReason:           "",
		NextAction:            ja.determineNextAction(llmImpactLevel),
		Reasoning:             llmResult.Reasoning,
	}
}

// parseImpactLevel parses impact level string from LLM
func (ja *JudgeAgent) parseImpactLevel(levelStr string) ImpactLevel {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "none":
		return ImpactNone
	case "minor":
		return ImpactMinor
	case "moderate":
		return ImpactModerate
	case "major":
		return ImpactMajor
	case "lethal":
		return ImpactLethal
	default:
		return ImpactNone
	}
}

// ==========================================================================
// Story 7.3: Intent Classification Implementation
// ==========================================================================

// ClassifyIntent analyzes free text input to determine player intent
//
// Story 7.3 AC2: Parse player's free text input to understand their intention
//
// Responsibilities:
//  1. Extract action verb and target from free text
//  2. Normalize intent for rule matching
//  3. Detect ambiguous input requiring clarification
//  4. Use Fast Model for interpretation (<500ms)
//
// Returns:
//  - IntentClassification with parsed intent
//  - ClarificationNeeded if input is ambiguous
//  - Error if classification fails
func (ja *JudgeAgent) ClassifyIntent(ctx context.Context, freeText string, gameState *GameStateSnapshot) (*IntentClassification, *ClarificationNeeded, error) {
	// Input validation
	freeText = strings.TrimSpace(freeText)
	if freeText == "" {
		return nil, &ClarificationNeeded{
			Reason:   "輸入為空",
			Question: "請描述你想要做什麼行動",
			SuggestedInterpretations: []string{
				"檢查周圍環境",
				"與NPC對話",
				"使用物品",
			},
		}, nil
	}

	// Check for very long input (>200 chars - Story 7.3 Phase 4)
	if len([]rune(freeText)) > 200 {
		return nil, &ClarificationNeeded{
			Reason:   "輸入過長，請簡化描述",
			Question: "請用簡短的句子描述你的行動（不超過200字）",
			SuggestedInterpretations: []string{
				"將長句拆分為主要行動",
				"只描述最重要的動作",
			},
		}, nil
	}

	// Call Fast Model for intent classification
	if ja.Config.LLMClient == nil {
		return nil, nil, fmt.Errorf("LLM client not configured")
	}

	prompt := ja.buildIntentPrompt(freeText, gameState)

	// Use BaseAgentImpl's InvokeWithRetry for robust LLM call
	response, err := ja.BaseAgentImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		return ja.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.2, // Very low temperature for consistent classification
			"max_tokens":  300,
		})
	})

	if err != nil {
		return nil, nil, fmt.Errorf("intent classification LLM call failed: %w", err)
	}

	// Type assert to string
	responseStr, ok := response.(string)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected response type: %T", response)
	}

	// Parse LLM response
	intentResp, err := ja.parseIntentResponse(responseStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse intent response: %w", err)
	}

	// Build IntentClassification
	intent := &IntentClassification{
		Action:           intentResp.Action,
		Target:           intentResp.Target,
		IsAmbiguous:      intentResp.IsAmbiguous,
		Confidence:       intentResp.Confidence,
		Keywords:         intentResp.Keywords,
		NormalizedIntent: intentResp.NormalizedIntent,
	}

	// Check if clarification is needed
	if intentResp.IsAmbiguous || intentResp.Confidence < 0.6 {
		clarification := &ClarificationNeeded{
			Reason:                   intentResp.ClarificationReason,
			SuggestedInterpretations: intentResp.SuggestedInterpret,
			Question:                 intentResp.ClarificationQuestion,
		}
		return intent, clarification, nil
	}

	// Clear intent - no clarification needed
	return intent, nil, nil
}

// buildIntentPrompt builds the LLM prompt for intent classification
func (ja *JudgeAgent) buildIntentPrompt(freeText string, gameState *GameStateSnapshot) string {
	var sb strings.Builder

	sb.WriteString("你是「規則怪談」遊戲的意圖解析 Agent。你的職責是理解玩家的自由文字輸入，並分析其行動意圖。\n\n")

	sb.WriteString("**任務：**\n")
	sb.WriteString("1. 從自由文字中提取【動作】(action) 和【目標】(target)\n")
	sb.WriteString("2. 判斷意圖是否明確 (is_ambiguous)\n")
	sb.WriteString("3. 給出信心度 (confidence: 0.0-1.0)\n")
	sb.WriteString("4. 標準化意圖 (normalized_intent) 以便後續規則匹配\n")
	sb.WriteString("5. 如果模糊，提供澄清問題和可能的解釋\n\n")

	sb.WriteString("**標準化動作分類：**\n")
	sb.WriteString("- 檢查/觀察: examine, inspect, look, check\n")
	sb.WriteString("- 移動: move, walk, go, run, flee\n")
	sb.WriteString("- 互動: talk, speak, use, open, close, touch\n")
	sb.WriteString("- 攻擊: attack, hit, fight, break\n")
	sb.WriteString("- 等待: wait, rest, stay, hide\n")
	sb.WriteString("- 其他: other\n\n")

	sb.WriteString("**輸出格式（JSON）：**\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"action\": \"檢查\",\n")
	sb.WriteString("  \"target\": \"鏡子\",\n")
	sb.WriteString("  \"is_ambiguous\": false,\n")
	sb.WriteString("  \"confidence\": 0.95,\n")
	sb.WriteString("  \"keywords\": [\"檢查\", \"鏡子\"],\n")
	sb.WriteString("  \"normalized_intent\": \"examine_mirror\",\n")
	sb.WriteString("  \"clarification_reason\": \"\",\n")
	sb.WriteString("  \"suggested_interpretations\": [],\n")
	sb.WriteString("  \"clarification_question\": \"\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("**範例 1 - 明確意圖：**\n")
	sb.WriteString("輸入: \"我想檢查房間裡的鏡子\"\n")
	sb.WriteString("輸出: {\"action\": \"檢查\", \"target\": \"鏡子\", \"is_ambiguous\": false, \"confidence\": 0.95, ...}\n\n")

	sb.WriteString("**範例 2 - 模糊意圖：**\n")
	sb.WriteString("輸入: \"看看那個東西\"\n")
	sb.WriteString("輸出: {\"action\": \"檢查\", \"target\": \"?\", \"is_ambiguous\": true, \"confidence\": 0.4, ")
	sb.WriteString("\"clarification_reason\": \"目標不明確\", ")
	sb.WriteString("\"suggested_interpretations\": [\"檢查鏡子\", \"檢查桌子\", \"檢查門\"], ")
	sb.WriteString("\"clarification_question\": \"你想檢查哪個物品？\"}\n\n")

	sb.WriteString("=== 當前遊戲狀態 ===\n")
	if gameState != nil {
		sb.WriteString(fmt.Sprintf("場景: %s\n", gameState.CurrentScene))
		sb.WriteString(fmt.Sprintf("HP: %d\n", gameState.HP))
		sb.WriteString(fmt.Sprintf("SAN: %d\n", gameState.SAN))
		if len(gameState.PlayerItems) > 0 {
			sb.WriteString(fmt.Sprintf("持有物品: %s\n", strings.Join(gameState.PlayerItems, ", ")))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("=== 玩家輸入 ===\n")
	sb.WriteString(freeText)
	sb.WriteString("\n\n")

	sb.WriteString("請分析此輸入並返回 JSON 格式結果。")

	return sb.String()
}

// parseIntentResponse parses the LLM's JSON response for intent classification
func (ja *JudgeAgent) parseIntentResponse(raw string) (*LLMIntentResponse, error) {
	// Try to extract JSON from markdown code block
	jsonStr := raw
	if start := strings.Index(raw, "```json"); start != -1 {
		jsonStr = raw[start+7:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	} else if start := strings.Index(raw, "```"); start != -1 {
		jsonStr = raw[start+3:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	var result LLMIntentResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w (raw: %s)", err, jsonStr)
	}

	return &result, nil
}

// InvokeJudgeWithIntent judges a player's free text input with intent classification
//
// Story 7.3 AC2: Entry point for free text input judgment
//
// Flow:
//  1. Classify intent from free text
//  2. If ambiguous, return clarification request
//  3. If clear, proceed with normal rule checking using normalized intent
//  4. Return judgment with intent information
func (ja *JudgeAgent) InvokeJudgeWithIntent(ctx context.Context, freeText string, gameState *GameStateSnapshot, activeRules []JudgeHiddenRule) (*JudgeResponseV2, error) {
	// Step 1: Classify intent
	intent, clarification, err := ja.ClassifyIntent(ctx, freeText, gameState)
	if err != nil {
		return nil, fmt.Errorf("intent classification failed: %w", err)
	}

	// Step 2: If clarification needed, return early
	if clarification != nil {
		return &JudgeResponseV2{
			JudgeResponse: &JudgeResponse{
				RulesViolated:         []RuleViolation{},
				ImpactLevel:           ImpactNone,
				SuggestedStateChanges: StateChanges{},
				DeathReason:           "",
				NextAction:            ActionContinueStory,
				Reasoning:             "Need clarification from player",
			},
			IntentClassification: intent,
			ClarificationNeeded:  clarification,
		}, nil
	}

	// Step 3: Use normalized intent for rule checking
	judgeRequest := &JudgeRequest{
		PlayerChoice: intent.NormalizedIntent,
		GameState:    gameState,
		ActiveRules:  activeRules,
	}

	// Call existing InvokeJudge with normalized intent
	judgeResp, err := ja.InvokeJudge(ctx, judgeRequest)
	if err != nil {
		return nil, fmt.Errorf("judgment failed: %w", err)
	}

	// Step 4: Wrap in V2 response
	return &JudgeResponseV2{
		JudgeResponse:        judgeResp,
		IntentClassification: intent,
		ClarificationNeeded:  nil,
	}, nil
}

// ==========================================================================
// Story 4-1: Chat Judgment Implementation
// ==========================================================================

// JudgeChat analyzes a player's chat message and determines semantic flags
//
// Story 4-1 AC1: JudgeAgent new method for chat message analysis
// Story 4-1 AC2: Accepts PlayerMessage, Participants, ConversationContext
// Story 4-1 AC3: Returns JudgeChatResult with Flags array
// Story 4-1 AC4: Uses LLM to judge message flags
// Story 4-1 AC5: Prompt includes participant emotions and conversation history
//
// Responsibilities:
//  1. Validate input parameters
//  2. Build comprehensive prompt with context
//  3. Call LLM for semantic analysis
//  4. Parse response and extract flags
//  5. Return structured result with confidence
//
// Returns:
//  - JudgeChatResult with detected flags, confidence, and reasoning
//  - Error if validation or LLM call fails
func (ja *JudgeAgent) JudgeChat(ctx context.Context, request *JudgeChatRequest) (*JudgeChatResult, error) {
	// Story 4-1 AC2: Validate input parameters
	if err := ja.validateChatRequest(request); err != nil {
		return nil, fmt.Errorf("invalid chat request: %w", err)
	}

	// Log agent invocation
	logger.Debug("JudgeAgent.JudgeChat invoked", map[string]interface{}{
		"player_message":     request.PlayerMessage,
		"num_participants":   len(request.Participants),
		"num_history":        len(request.ConversationHistory),
		"num_relevant_facts": len(request.RelevantFacts),
	})

	// Story 4-1 AC5: Build comprehensive prompt
	prompt := ja.buildChatJudgePrompt(request)

	// Story 4-1 AC4: Call LLM for judgment
	if ja.Config.LLMClient == nil {
		return nil, fmt.Errorf("LLM client not configured")
	}

	response, err := ja.BaseAgentImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		return ja.Config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.3, // Low temperature for consistent judgment
			"max_tokens":  600,
		})
	})

	if err != nil {
		// Graceful degradation: return empty flags on LLM failure
		logger.Warn("JudgeChat LLM call failed, returning empty flags", map[string]interface{}{
			"error": err.Error(),
		})
		return &JudgeChatResult{
			Flags:      []ChatFlag{},
			Confidence: 0.0,
			Reasoning:  fmt.Sprintf("LLM call failed: %v", err),
		}, nil
	}

	// Type assert to string
	responseStr, ok := response.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	// Story 4-1 AC4: Parse JSON response
	result, err := ja.parseChatJudgeResponse(responseStr)
	if err != nil {
		// Graceful degradation: return empty flags on parse failure
		logger.Warn("JudgeChat response parsing failed, returning empty flags", map[string]interface{}{
			"error":    err.Error(),
			"response": responseStr,
		})
		return &JudgeChatResult{
			Flags:      []ChatFlag{},
			Confidence: 0.0,
			Reasoning:  fmt.Sprintf("Parse failed: %v", err),
		}, nil
	}

	// Log successful judgment
	logger.Debug("JudgeChat completed", map[string]interface{}{
		"flags":      result.Flags,
		"confidence": result.Confidence,
	})

	return result, nil
}

// validateChatRequest validates the JudgeChatRequest parameters
//
// Story 4-1 AC2: Input validation
func (ja *JudgeAgent) validateChatRequest(request *JudgeChatRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// PlayerMessage must be non-empty
	if strings.TrimSpace(request.PlayerMessage) == "" {
		return fmt.Errorf("player message cannot be empty")
	}

	// Must have at least one NPC participant (excluding player)
	hasNPC := false
	for _, p := range request.Participants {
		if !p.IsPlayer {
			hasNPC = true
			break
		}
	}
	if !hasNPC {
		return fmt.Errorf("must have at least one NPC participant")
	}

	// GameState should be available
	if request.GameState == nil {
		return fmt.Errorf("game state cannot be nil")
	}

	return nil
}

// buildChatJudgePrompt builds the LLM prompt for chat message judgment
//
// Story 4-1 AC5: Comprehensive prompt with all context
func (ja *JudgeAgent) buildChatJudgePrompt(request *JudgeChatRequest) string {
	var sb strings.Builder

	// System role and task description
	sb.WriteString("You are a narrative judge analyzing player dialogue in a horror game chat session.\n\n")

	sb.WriteString("PARTICIPANTS:\n")
	sb.WriteString("- Player\n")
	for _, p := range request.Participants {
		if !p.IsPlayer {
			sb.WriteString(fmt.Sprintf("- %s (NPC, Trust: %d, Fear: %d, Stress: %d, Relationship: %s)\n",
				p.Name, p.Emotion.Trust, p.Emotion.Fear, p.Emotion.Stress, p.Relationship))
		}
	}
	sb.WriteString("\n")

	// Recent conversation history
	if len(request.ConversationHistory) > 0 {
		sb.WriteString("RECENT CONVERSATION:\n")
		for _, msg := range request.ConversationHistory {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", msg.Timestamp, msg.Speaker, msg.Content))
		}
		sb.WriteString("\n")
	}

	// Known facts
	if len(request.RelevantFacts) > 0 {
		sb.WriteString("KNOWN FACTS:\n")
		for _, fact := range request.RelevantFacts {
			sb.WriteString(fmt.Sprintf("- %s\n", fact))
		}
		sb.WriteString("\n")
	}

	// Current game state context
	if request.GameState != nil {
		sb.WriteString("GAME STATE:\n")
		sb.WriteString(fmt.Sprintf("- Scene: %s\n", request.GameState.CurrentScene))
		sb.WriteString(fmt.Sprintf("- HP: %d\n", request.GameState.HP))
		sb.WriteString(fmt.Sprintf("- SAN: %d\n", request.GameState.SAN))
		sb.WriteString("\n")
	}

	// Player message to analyze
	sb.WriteString("PLAYER MESSAGE:\n")
	sb.WriteString(fmt.Sprintf("\"%s\"\n\n", request.PlayerMessage))

	// Task description
	sb.WriteString("TASK:\n")
	sb.WriteString("Analyze this message and determine which flags apply:\n")
	sb.WriteString("- hallucination: Player states something contradicting known facts\n")
	sb.WriteString("- hostile: Player shows threat or aggression\n")
	sb.WriteString("- revelation: Player shares important new information\n")
	sb.WriteString("- contradiction: Player's statement conflicts with NPCs' knowledge\n")
	sb.WriteString("- persuasion: Player attempts to convince NPCs\n")
	sb.WriteString("- lie: Player is lying (may be exposed later)\n\n")

	// Expected output format
	sb.WriteString("Return JSON:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"flags\": [\"revelation\", \"contradiction\"],\n")
	sb.WriteString("  \"confidence\": 0.85,\n")
	sb.WriteString("  \"reasoning\": \"Player shares new info (revelation) but contradicts known facts (contradiction)\"\n")
	sb.WriteString("}\n")

	return sb.String()
}

// parseChatJudgeResponse parses the LLM's JSON response for chat judgment
//
// Story 4-1 AC4: Parse LLM response and extract flags
func (ja *JudgeAgent) parseChatJudgeResponse(raw string) (*JudgeChatResult, error) {
	// Try to extract JSON from markdown code block
	jsonStr := raw
	if start := strings.Index(raw, "```json"); start != -1 {
		jsonStr = raw[start+7:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	} else if start := strings.Index(raw, "```"); start != -1 {
		jsonStr = raw[start+3:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	var llmResp LLMChatJudgeResponse
	if err := json.Unmarshal([]byte(jsonStr), &llmResp); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w (raw: %s)", err, jsonStr)
	}

	// Convert string flags to ChatFlag type
	flags := make([]ChatFlag, 0, len(llmResp.Flags))
	for _, flagStr := range llmResp.Flags {
		flag := ChatFlag(strings.ToLower(strings.TrimSpace(flagStr)))
		// Validate flag
		if ja.isValidChatFlag(flag) {
			flags = append(flags, flag)
		}
	}

	return &JudgeChatResult{
		Flags:      flags,
		Confidence: llmResp.Confidence,
		Reasoning:  llmResp.Reasoning,
	}, nil
}

// isValidChatFlag checks if a ChatFlag is valid
func (ja *JudgeAgent) isValidChatFlag(flag ChatFlag) bool {
	switch flag {
	case FlagHallucination, FlagHostile, FlagRevelation, FlagContradiction, FlagPersuasion, FlagLie:
		return true
	default:
		return false
	}
}
