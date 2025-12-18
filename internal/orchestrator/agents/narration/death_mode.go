package narration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// ==========================================================================
// Story 7.8 AC3: Death Narration Mode
// ==========================================================================

// DeathNarrationRequest represents a request to generate NPC death narrative
//
// Story 7.8 AC3: Death narration input
//   - NPC information and death context
//   - Player responsibility and emotional impact
//   - Content mode setting (18+ or normal)
//   - Current game state for context
type DeathNarrationRequest struct {
	// NPC Information
	NPCID       string              `json:"npc_id"`
	NPCName     string              `json:"npc_name"`
	NPCArchetype agents.NPCArchetype `json:"npc_archetype"`
	NPCBackstory string              `json:"npc_backstory"`

	// Death Context
	DeathBeat            int     `json:"death_beat"`
	DeathReason          string  `json:"death_reason"`
	Location             string  `json:"location"`
	PlayerChoice         string  `json:"player_choice"`
	PlayerResponsibility float64 `json:"player_responsibility"` // 0.0-1.0

	// Emotional Context
	Intimacy int `json:"intimacy"` // 0-100 relationship strength

	// Content Mode
	Adult18Plus bool `json:"adult_18_plus"` // Detailed (18+) or Implied (normal)

	// Game Context
	CurrentHP  int `json:"current_hp"`
	CurrentSAN int `json:"current_san"`

	// Foreshadowing (optional, for richer narrative)
	// Uses orchestrator.DeathForeshadow to avoid type duplication
	Foreshadows []orchestrator.DeathForeshadow `json:"foreshadows,omitempty"`
}

// DeathNarrationResponse represents the generated death narrative
//
// Story 7.8 AC3: Death narration output
//   - Process description: How death occurred
//   - Detail portrayal: Pain, fear, final expression
//   - Emotional impact: Player guilt or helplessness
//   - Last words: NPC's final words (30-80 chars)
type DeathNarrationResponse struct {
	// Main death narrative (200-400 chars)
	DeathNarrative string `json:"death_narrative"`

	// NPC's final words (30-80 chars)
	LastWords string `json:"last_words"`

	// Emotional tone of the narrative
	EmotionalTone string `json:"emotional_tone"` // guilt, helplessness, tragedy, horror

	// SAN loss (calculated, not generated)
	SANLoss int `json:"san_loss"`
}

// InvokeDeath generates NPC death narrative
//
// Story 7.8 AC3: Death Narration Mode
//   - Generates emotionally impactful death narrative
//   - Includes process description, details, emotional impact
//   - Generates meaningful last words
//   - Adapts to content mode (18+ detailed vs normal implied)
//   - Ensures death has meaning (leaves clues or lessons)
//
// AC Requirements:
//   - Narrative length: 200-400 characters
//   - Last words: 30-80 characters
//   - Content mode adaptation (18+ detailed vs normal implied)
//   - Emotional impact (guilt, helplessness, tragedy)
//   - Meaningful death (clues or lessons)
//
// Parameters:
//   - ctx: Context for timeout control
//   - req: Death narration request
//
// Returns:
//   - *DeathNarrationResponse: Generated death narrative
//   - error: Error if generation fails
func (a *NarrationAgent) InvokeDeath(ctx context.Context, request *DeathNarrationRequest) (*DeathNarrationResponse, error) {
	log.Printf("[%s] InvokeDeath started: NPC=%s, Beat=%d, Responsibility=%.2f",
		a.config.Name, request.NPCName, request.DeathBeat, request.PlayerResponsibility)

	// Validate request
	if err := validateDeathRequest(request); err != nil {
		return nil, fmt.Errorf("invalid death request: %w", err)
	}

	// Calculate SAN loss
	sanLoss := orchestrator.CalculateNPCDeathSAN(request.PlayerResponsibility, request.Intimacy)

	// Story 7-8 Code Review Fix H-3: LLM fast-path fallback
	// Check LLM availability BEFORE retry loop for instant fallback
	// (Same pattern as InvokeDialogue in Story 7-7)
	if a.config.LLMClient == nil {
		log.Printf("[%s] No LLM client available, using fast-path fallback", a.config.Name)
		return a.GenerateSimpleDeathNarrative(request), nil
	}

	// Use BaseAgentImpl retry mechanism
	result, err := a.baseImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		// 1. Build death narration prompt
		prompt := a.buildDeathPrompt(request)
		log.Printf("[%s] Death prompt built (length: %d chars)", a.config.Name, len(prompt))

		// 2. Call LLM
		response, err := a.config.LLMClient.Generate(ctx, prompt, nil)
		if err != nil {
			log.Printf("[%s] LLM call failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// 3. Parse JSON response
		var deathResp DeathNarrationResponse
		if err := json.Unmarshal([]byte(response), &deathResp); err != nil {
			log.Printf("[%s] Failed to parse JSON: %v", a.config.Name, err)
			return nil, fmt.Errorf("failed to parse death response: %w", err)
		}

		// 4. Validate response
		if err := a.validateDeathResponse(&deathResp); err != nil {
			log.Printf("[%s] Validation failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("invalid death response: %w", err)
		}

		// 5. Set calculated SAN loss
		deathResp.SANLoss = sanLoss

		log.Printf("[%s] Death narrative generated successfully (narrative: %d chars, last_words: %d chars, SAN loss: %d)",
			a.config.Name,
			len([]rune(deathResp.DeathNarrative)),
			len([]rune(deathResp.LastWords)),
			deathResp.SANLoss)

		return &deathResp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*DeathNarrationResponse), nil
}

// ==========================================================================
// Helper Methods - Death Narration
// ==========================================================================

// validateDeathRequest validates death narration request
func validateDeathRequest(req *DeathNarrationRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.NPCID == "" {
		return fmt.Errorf("NPC ID cannot be empty")
	}

	if req.NPCName == "" {
		return fmt.Errorf("NPC name cannot be empty")
	}

	if req.DeathReason == "" {
		return fmt.Errorf("death reason cannot be empty")
	}

	if req.DeathBeat < 0 {
		return fmt.Errorf("death beat must be >= 0")
	}

	if req.PlayerResponsibility < 0.0 || req.PlayerResponsibility > 1.0 {
		return fmt.Errorf("player responsibility must be in range [0.0, 1.0]")
	}

	if req.Intimacy < 0 || req.Intimacy > 100 {
		return fmt.Errorf("intimacy must be in range [0, 100]")
	}

	return nil
}

// buildDeathPrompt builds the death narration generation prompt
//
// Story 7.8 AC3: Complete death narration prompt
func (a *NarrationAgent) buildDeathPrompt(request *DeathNarrationRequest) string {
	var sb strings.Builder

	sb.WriteString("你是一個專業的恐怖遊戲敘事 AI，負責生成具有強烈情緒衝擊的 NPC 死亡場景。\n\n")

	// NPC Information
	sb.WriteString("## NPC 資訊\n")
	sb.WriteString(fmt.Sprintf("- 名稱：%s\n", request.NPCName))
	sb.WriteString(fmt.Sprintf("- 原型：%s\n", getArchetypeName(request.NPCArchetype)))
	if request.NPCBackstory != "" {
		sb.WriteString(fmt.Sprintf("- 背景：%s\n", request.NPCBackstory))
	}
	sb.WriteString("\n")

	// Death Context
	sb.WriteString("## 死亡情境\n")
	sb.WriteString(fmt.Sprintf("- Beat: %d\n", request.DeathBeat))
	sb.WriteString(fmt.Sprintf("- 地點：%s\n", request.Location))
	sb.WriteString(fmt.Sprintf("- 死因：%s\n", request.DeathReason))
	sb.WriteString(fmt.Sprintf("- 玩家選擇：%s\n", request.PlayerChoice))
	sb.WriteString("\n")

	// Player Responsibility and Emotional Context
	sb.WriteString("## 情感背景\n")
	sb.WriteString(fmt.Sprintf("- 玩家責任程度：%.1f/1.0（%.0f%%）\n",
		request.PlayerResponsibility,
		request.PlayerResponsibility*100))
	sb.WriteString(fmt.Sprintf("- 親密度：%d/100\n", request.Intimacy))
	sb.WriteString("\n")

	sb.WriteString(getResponsibilityGuidance(request.PlayerResponsibility))
	sb.WriteString("\n")

	// Foreshadowing (if available)
	if len(request.Foreshadows) > 0 {
		sb.WriteString("## 先前的伏筆線索\n")
		for i, foreshadow := range request.Foreshadows {
			sb.WriteString(fmt.Sprintf("%d. Beat %d: %s\n", i+1, foreshadow.Beat, foreshadow.Content))
		}
		sb.WriteString("\n這些伏筆暗示了危險的到來，可以在敘事中暗示「早有預兆」。\n\n")
	}

	// Content Mode
	sb.WriteString("## 內容模式\n")
	if request.Adult18Plus {
		sb.WriteString("- 模式：18+ 成人模式（詳細、具體、寫實）\n")
		sb.WriteString("- 描寫風格：\n")
		sb.WriteString("  - 詳細描述死亡過程的每個細節\n")
		sb.WriteString("  - 具體刻畫痛苦、恐懼、絕望的表情與動作\n")
		sb.WriteString("  - 寫實地展現血腥、暴力、恐怖的場面\n")
		sb.WriteString("  - 不迴避殘酷的現實，給予玩家強烈的衝擊\n")
	} else {
		sb.WriteString("- 模式：普通模式（隱晦、暗示、留白）\n")
		sb.WriteString("- 描寫風格：\n")
		sb.WriteString("  - 用暗示和留白代替直接描述\n")
		sb.WriteString("  - 重點放在情緒反應和氛圍營造\n")
		sb.WriteString("  - 避免過於血腥或露骨的描寫\n")
		sb.WriteString("  - 讓讀者自行想像具體細節\n")
	}
	sb.WriteString("\n")

	// Game State
	sb.WriteString("## 當前遊戲狀態\n")
	sb.WriteString(fmt.Sprintf("- HP: %d/100\n", request.CurrentHP))
	sb.WriteString(fmt.Sprintf("- SAN: %d/100\n", request.CurrentSAN))
	sb.WriteString("\n")

	// Output Requirements
	sb.WriteString("## 輸出要求\n")
	sb.WriteString("請生成一個 JSON 格式的死亡敘事，包含以下字段：\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"death_narrative\": \"死亡敘事文本（200-400 字）\",\n")
	sb.WriteString("  \"last_words\": \"NPC 的最後話語（30-80 字）\",\n")
	sb.WriteString("  \"emotional_tone\": \"情感基調（guilt/helplessness/tragedy/horror）\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("### 敘事要求（Story 7.8 AC3）：\n")
	sb.WriteString("1. **過程描述**：詳細描述死亡如何發生（符合死因與規則）\n")
	sb.WriteString("2. **細節刻畫**：生動描繪痛苦、恐懼、最後表情（根據內容模式調整尺度）\n")
	sb.WriteString("3. **情緒衝擊**：根據玩家責任程度，傳達罪惡感或無力感\n")
	sb.WriteString("4. **有意義的死亡**：確保死亡留下線索、警告或教訓\n")
	sb.WriteString("5. **最後話語**：生成符合情境的遺言（30-80 字）\n\n")

	sb.WriteString("### 情感基調選擇指南：\n")
	sb.WriteString("- **guilt（罪惡感）**：玩家責任高（>0.7），強調「這是你的錯」\n")
	sb.WriteString("- **helplessness（無力感）**：玩家責任低（<0.3），強調「無法改變的命運」\n")
	sb.WriteString("- **tragedy（悲劇）**：玩家責任中等（0.3-0.7），強調「本可避免的悲劇」\n")
	sb.WriteString("- **horror（恐怖）**：死因極度殘酷，強調「恐怖與絕望」\n\n")

	sb.WriteString("注意事項：\n")
	sb.WriteString("1. 敘事長度必須在 200-400 字之間\n")
	sb.WriteString("2. 最後話語必須在 30-80 字之間\n")
	sb.WriteString("3. 必須根據內容模式調整描寫尺度\n")
	sb.WriteString("4. 確保情感衝擊強烈，讓玩家感受到「這是我的錯」或「我無力改變」\n")
	sb.WriteString("5. 必須返回有效的 JSON 格式\n")

	return sb.String()
}

// getArchetypeName returns human-readable archetype name
func getArchetypeName(archetype agents.NPCArchetype) string {
	info := agents.GetArchetypeInfo(archetype)
	return info.Name
}

// getResponsibilityGuidance returns guidance based on player responsibility
func getResponsibilityGuidance(responsibility float64) string {
	if responsibility > 0.7 {
		return `**玩家責任極高（>70%）**：
- 敘事應該強烈暗示「這是玩家的錯」
- 使用「如果你當時……」「都是因為你……」等語句
- NPC 的最後話語可能帶有責備或質問
- 情感基調應該是強烈的罪惡感（guilt）`
	} else if responsibility > 0.3 {
		return `**玩家責任中等（30-70%）**：
- 敘事應該呈現「本可避免的悲劇」
- 暗示玩家有機會阻止，但沒有全力去做
- NPC 的最後話語可能是遺憾或警告
- 情感基調應該是悲劇感（tragedy）與部分罪惡感`
	} else {
		return `**玩家責任較低（<30%）**：
- 敘事應該強調「無法改變的命運」和「無力感」
- 即使玩家盡力，仍然無法阻止
- NPC 的最後話語可能是感激或安慰
- 情感基調應該是無力感（helplessness）與悲傷`
	}
}

// validateDeathResponse validates death narration response
//
// Story 7.8 AC3: Validate narrative and last words length
// Story 7-8 Code Review Fix M-2: Correct validation bounds to match AC requirements
func (a *NarrationAgent) validateDeathResponse(resp *DeathNarrationResponse) error {
	// Check narrative length (200-400 chars) - Story 7.8 AC3 requirement
	narrativeLen := len([]rune(resp.DeathNarrative))
	if narrativeLen < 200 {
		return fmt.Errorf("death narrative too short: %d characters (expected 200-400)", narrativeLen)
	}
	if narrativeLen > 400 {
		return fmt.Errorf("death narrative too long: %d characters (expected 200-400)", narrativeLen)
	}

	// Check last words length (30-80 chars) - Story 7.8 AC3 requirement
	lastWordsLen := len([]rune(resp.LastWords))
	if lastWordsLen < 30 {
		return fmt.Errorf("last words too short: %d characters (expected 30-80)", lastWordsLen)
	}
	if lastWordsLen > 80 {
		return fmt.Errorf("last words too long: %d characters (expected 30-80)", lastWordsLen)
	}

	// Check emotional tone
	validTones := map[string]bool{
		"guilt":        true,
		"helplessness": true,
		"tragedy":      true,
		"horror":       true,
	}
	if !validTones[resp.EmotionalTone] {
		return fmt.Errorf("invalid emotional tone: %s (expected guilt/helplessness/tragedy/horror)", resp.EmotionalTone)
	}

	// Check that narrative is not empty
	if strings.TrimSpace(resp.DeathNarrative) == "" {
		return fmt.Errorf("death narrative is empty")
	}

	// Check that last words is not empty
	if strings.TrimSpace(resp.LastWords) == "" {
		return fmt.Errorf("last words is empty")
	}

	return nil
}

// ==========================================================================
// Story 7.8 AC3: Simplified Death Narrative Generation (Fallback)
// ==========================================================================

// GenerateSimpleDeathNarrative generates a simple death narrative without LLM
//
// This is a fallback method when LLM is not available.
// It uses the GenerateDeathNarrative helper function.
//
// Parameters:
//   - request: Death narration request
//
// Returns:
//   - *DeathNarrationResponse: Generated death narrative
func (a *NarrationAgent) GenerateSimpleDeathNarrative(request *DeathNarrationRequest) *DeathNarrationResponse {
	// Determine style based on content mode
	// Uses orchestrator types to avoid duplication
	style := orchestrator.DeathNarrativeImplied
	if request.Adult18Plus {
		style = orchestrator.DeathNarrativeDetailed
	}

	// Generate narrative using orchestrator helper function
	narrative, lastWords := orchestrator.GenerateDeathNarrative(
		request.NPCName,
		request.DeathReason,
		request.PlayerResponsibility,
		style,
	)

	// Calculate SAN loss
	sanLoss := orchestrator.CalculateNPCDeathSAN(request.PlayerResponsibility, request.Intimacy)

	// Determine emotional tone
	emotionalTone := "tragedy"
	if request.PlayerResponsibility > 0.7 {
		emotionalTone = "guilt"
	} else if request.PlayerResponsibility < 0.3 {
		emotionalTone = "helplessness"
	}

	return &DeathNarrationResponse{
		DeathNarrative: narrative,
		LastWords:      lastWords,
		EmotionalTone:  emotionalTone,
		SANLoss:        sanLoss,
	}
}

// ==========================================================================
// Note: Helper functions (CalculateNPCDeathSAN, GenerateDeathNarrative,
// DeathNarrativeStyle) are now in internal/orchestrator/npc_death.go
// to avoid code duplication (Story 7-8 Code Review Fix H-1, H-2, M-1)
// ==========================================================================
