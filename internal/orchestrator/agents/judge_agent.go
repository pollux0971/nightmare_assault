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
	// 1. Rule engine check (fast path)
	violations := ja.CheckRuleViolation(request.PlayerChoice, request.ActiveRules)
	impactLevel := ja.DetermineImpactLevel(violations)

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

	for _, v := range violations {
		// Initialize warnings if not set
		if _, exists := currentWarnings[v.RuleID]; !exists {
			currentWarnings[v.RuleID] = maxWarnings
		}

		remainingWarnings := currentWarnings[v.RuleID]

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
