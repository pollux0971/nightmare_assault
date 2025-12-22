package narration

import (
	"context"
	"fmt"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// AutoContentRequest represents a request for automatic content generation
// Story 7-2 AC2: 接受 GameState/StoryBible/AutoAction/Context
type AutoContentRequest struct {
	// GameState is the current game state
	GameState *engine.GameStateV2

	// StoryBible contains the story skeleton
	StoryBible *StoryBible

	// AutoAction is the automatic action description (optional)
	AutoAction string

	// RiskLevel indicates current risk (used for safe narrative generation)
	RiskLevel string // "none", "low", "medium", "high", "lethal"

	// Beat is the current beat number
	Beat int
}

// AutoContentResponse represents the response from automatic content generation
// Story 7-2 AC4: 返回 Narrative/HPDelta/SANDelta/PlantedSeeds/RevealedClues
type AutoContentResponse struct {
	// Narrative is the generated narrative text (300-800 characters for auto mode)
	Narrative string

	// HPDelta is the HP change (should be safe, avoiding lethal damage)
	HPDelta int

	// SANDelta is the SAN change (should be gradual)
	SANDelta int

	// PlantedSeeds are newly planted seeds
	PlantedSeeds []string

	// RevealedClues are revealed clues
	RevealedClues []string
}

// InvokeAutoContent generates safe automatic narrative content
// Story 7-2 Implementation
//
// AC1: NarrationAgent 新增 InvokeAutoContent() 方法
// AC2: 接受 GameState/StoryBible/AutoAction/Context
// AC3: 生成安全的自動行動敘事（避免致命風險）
// AC4: 返回 Narrative/HPDelta/SANDelta/PlantedSeeds/RevealedClues
// AC5: Prompt 包含風險提示與安全指引
//
// This method generates narrative for automatic resolution (Story 7-1).
// Unlike regular content generation, this focuses on:
//   - Safety: Avoid lethal damage or SAN loss
//   - Brevity: 300-800 characters instead of 500-1200
//   - Progression: Move story forward without critical decisions
//
// Parameters:
//   - ctx: Context for timeout control
//   - req: AutoContentRequest with game state and context
//
// Returns:
//   - *AutoContentResponse: Generated safe narrative and state changes
//   - error: Error if generation fails
func (a *NarrationAgent) InvokeAutoContent(ctx context.Context, req *AutoContentRequest) (*AutoContentResponse, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("auto content request cannot be nil")
	}
	if req.GameState == nil {
		return nil, fmt.Errorf("game state cannot be nil")
	}

	// AC3: Generate safe narrative (avoiding lethal risks)
	// AC5: Build prompt with safety guidelines
	prompt := a.buildAutoContentPrompt(req)

	// For now, return a placeholder response
	// TODO: Implement actual LLM call when LLMClient is available
	response := &AutoContentResponse{
		Narrative:     a.generateSafeNarrative(req),
		HPDelta:       a.calculateSafeHPDelta(req),
		SANDelta:      a.calculateSafeSANDelta(req),
		PlantedSeeds:  make([]string, 0),
		RevealedClues: make([]string, 0),
	}

	// Log if logger available (using baseImpl logger)
	// Note: Logger is managed by baseImpl, not config directly

	_ = prompt // Will be used in actual LLM call

	return response, nil
}

// buildAutoContentPrompt builds the prompt for automatic content generation
// AC5: Prompt 包含風險提示與安全指引
func (a *NarrationAgent) buildAutoContentPrompt(req *AutoContentRequest) string {
	// Build prompt with safety guidelines
	prompt := fmt.Sprintf(`你是一個恐怖遊戲的敘事生成器。請為當前回合生成安全的自動敘事內容。

遊戲狀態：
- 當前回合：%d
- 當前 HP：%d
- 當前 SAN：%d
- 風險等級：%s

**安全指引（非常重要）**：
1. 避免致命傷害：HP 變化應在 -5 到 0 之間
2. 避免嚴重理智損失：SAN 變化應在 -5 到 0 之間
3. 生成漸進式敘事：推動故事前進，但不做關鍵決策
4. 保持簡潔：敘事長度 300-800 字

請生成：
1. 敘事文本（300-800 字）
2. HP 變化值（-5 到 0）
3. SAN 變化值（-5 到 0）
4. 可選：種植的伏筆
5. 可選：揭露的線索

請以 JSON 格式回應。
`,
		req.Beat,
		req.GameState.HP,
		req.GameState.SAN,
		req.RiskLevel,
	)

	return prompt
}

// generateSafeNarrative generates a safe placeholder narrative
// This will be replaced by actual LLM call
func (a *NarrationAgent) generateSafeNarrative(req *AutoContentRequest) string {
	// Placeholder implementation
	narratives := []string{
		"你繼續前進，走廊盡頭傳來微弱的聲響。周圍的牆壁看起來更加破舊，空氣中飄散著一股難以名狀的氣味。",
		"你小心翼翼地探索著這個詭異的空間。遠處傳來滴水聲，迴盪在空蕩蕩的房間裡，讓氣氛變得更加壓抑。",
		"你的腳步聲在空曠的走廊中迴響。牆上的畫像似乎在注視著你的每一個動作，讓你感到一陣不安。",
	}

	// Return a random narrative based on beat number
	index := req.Beat % len(narratives)
	return narratives[index]
}

// calculateSafeHPDelta calculates safe HP change
// AC3: 避免致命風險
func (a *NarrationAgent) calculateSafeHPDelta(req *AutoContentRequest) int {
	// Safe HP delta: between -5 and 0
	// Higher risk allows more damage
	switch req.RiskLevel {
	case "high", "lethal":
		return -5
	case "medium":
		return -3
	case "low":
		return -1
	default:
		return 0
	}
}

// calculateSafeSANDelta calculates safe SAN change
// AC3: 避免嚴重理智損失
func (a *NarrationAgent) calculateSafeSANDelta(req *AutoContentRequest) int {
	// Safe SAN delta: between -5 and 0
	// Always cause some SAN loss for atmospheric progression
	switch req.RiskLevel {
	case "high", "lethal":
		return -4
	case "medium":
		return -3
	case "low":
		return -2
	default:
		return -1
	}
}
