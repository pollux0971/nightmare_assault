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

	// 4. Process HP/SAN changes from JudgeResult
	response := &ContentResponse{
		MainNarrative:   narrative,
		PlantedSeeds:    make([]string, 0),   // TODO: Integrate Seed Agent (Story 6.5)
		HarvestedSeeds:  make([]string, 0),   // TODO: Integrate Seed Agent (Story 6.5)
		RuleHints:       make([]string, 0),   // TODO: Implement Rule Hints generation
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
func (a *NarrationAgent) buildContentPrompt(req *ContentRequest, directive *engine.TensionDirective) (string, error) {
	// Simplified prompt for now
	// TODO: Use template system when available (Task 12)

	prompt := fmt.Sprintf(`你是「規則怪談」類型恐怖遊戲的敘事引擎。

當前 Beat: %d
玩家上一個選擇：%s

【張力指令】
%s

請生成 500-1200 字的敘事內容，遵循以上張力指令的要求。

輸出格式（純文本）：
[敘事內容]
`,
		req.Beat,
		req.LastPlayerChoice,
		directive.FormatForPrompt(),
	)

	return prompt, nil
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
