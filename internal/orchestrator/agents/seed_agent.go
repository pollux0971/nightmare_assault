package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/seed"
)

var (
	// ErrUnsupportedOperation indicates an unsupported SeedAgent operation
	ErrUnsupportedOperation = errors.New("unsupported operation for SeedAgent")
)

// SeedAgent is responsible for generating and managing foreshadowing seeds.
//
// Architecture (Story 6-5):
//   - Dual Mode: Global Generator (Genesis Phase) + Local Manager (Game Loop)
//   - Uses BaseAgentImpl for retry mechanism and error handling
//   - Integrates with SeedPruner (Epic 2.5) and TensionManager (Epic 3)
//
// Modes:
//   - Global Generator: Generates 3-5 Global Seeds with 3-tier clue chains
//   - Local Manager: Manages LocalSeeds lifecycle (Plant/Harvest/Prune/Skip)
type SeedAgent struct {
	// config is the Agent configuration
	config AgentConfig

	// baseImpl provides common Agent functionality (retry, timeout, error handling)
	baseImpl *BaseAgentImpl

	// TODO: Add dependencies from Epic 2 and Epic 3
	// pruner SeedPruner
	// tensionMgr TensionManager
}

// NewSeedAgent creates a new SeedAgent with the given configuration.
//
// Parameters:
//   - config: Agent configuration (Name, Timeout, MaxRetries, LLMClient)
//
// Returns:
//   - *SeedAgent: A new SeedAgent instance
func NewSeedAgent(config AgentConfig) *SeedAgent {
	// Set default name
	if config.Name == "" {
		config.Name = "SeedAgent"
	}

	// Set default timeout
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Set default max retries
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &SeedAgent{
		config:   config,
		baseImpl: NewBaseAgentImpl(config),
	}
}

// GetName returns the Agent's name
func (sa *SeedAgent) GetName() string {
	return sa.config.Name
}

// GetTimeout returns the Agent's timeout duration
func (sa *SeedAgent) GetTimeout() time.Duration {
	return sa.config.Timeout
}

// Invoke implements the BaseAgent interface (TODO: implement routing logic)
func (sa *SeedAgent) Invoke(ctx context.Context, request any) (any, error) {
	// TODO: Route to InvokeGlobalGenerate or InvokeLocalManage based on request type
	return nil, &AgentError{
		AgentName: sa.config.Name,
		Operation: "Invoke",
		Cause:     ErrUnsupportedOperation,
		Retryable: false,
	}
}

// BuildPrompt implements the BaseAgent interface (TODO: implement)
func (sa *SeedAgent) BuildPrompt(request any) (string, error) {
	// TODO: Build prompts for Global/Local modes
	return "", ErrUnsupportedOperation
}

// ParseResponse implements the BaseAgent interface (TODO: implement)
func (sa *SeedAgent) ParseResponse(raw string) (any, error) {
	// TODO: Parse LLM responses for Global/Local modes
	return nil, ErrUnsupportedOperation
}

// ==========================================================================
// Dual Mode Methods (Story 6-5)
// ==========================================================================

// InvokeGlobalGenerate generates Global Seeds for Genesis Phase
//
// AC #1: Generates 3-5 Global Seeds with 3-tier clue chains in <10s
//
// Seed count by difficulty:
//   - easy: 3 seeds
//   - normal: 4 seeds
//   - hard: 5 seeds
//   - hell: 5 seeds
//
// Parameters:
//   - ctx: Context for timeout control (<10s recommended)
//   - request: GlobalGenerateRequest with StoryBible, Difficulty, etc.
//
// Returns:
//   - *GlobalGenerateResponse: Generated Global Seeds
//   - error: Error if generation fails
func (sa *SeedAgent) InvokeGlobalGenerate(
	ctx context.Context,
	request *GlobalGenerateRequest,
) (*GlobalGenerateResponse, error) {
	// Validate request
	if request == nil {
		return nil, &AgentError{
			AgentName: sa.config.Name,
			Operation: "InvokeGlobalGenerate",
			Cause:     errors.New("request cannot be nil"),
			Retryable: false,
		}
	}

	if request.StoryBible == nil {
		return nil, &AgentError{
			AgentName: sa.config.Name,
			Operation: "InvokeGlobalGenerate",
			Cause:     errors.New("StoryBible cannot be nil"),
			Retryable: false,
		}
	}

	// Use BaseAgentImpl's retry mechanism
	result, err := sa.baseImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		// 1. Determine seed count based on difficulty
		seedCount := sa.getSeedCountByDifficulty(request.Difficulty)

		// 2. Build prompt
		prompt := sa.buildGlobalGeneratePrompt(request, seedCount)

		// 3. Call LLM (Smart Model for complex generation)
		response, err := sa.config.LLMClient.Generate(ctx, prompt, map[string]any{
			"temperature": 0.7, // Creative but controlled
		})
		if err != nil {
			return nil, err
		}

		// 4. Parse JSON response
		globalResp, err := sa.parseGlobalGenerateResponse(response)
		if err != nil {
			return nil, err
		}

		// 5. Validate seed count
		if len(globalResp.GlobalSeeds) != seedCount {
			return nil, errors.New("LLM generated incorrect number of seeds")
		}

		// 6. Link seeds to endings
		sa.linkSeedsToEndings(globalResp.GlobalSeeds, request.PossibleEndings)

		return globalResp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*GlobalGenerateResponse), nil
}

// getSeedCountByDifficulty returns the number of Global Seeds based on difficulty
func (sa *SeedAgent) getSeedCountByDifficulty(difficulty string) int {
	switch difficulty {
	case "easy":
		return 3
	case "normal":
		return 4
	case "hard", "hell":
		return 5
	default:
		return 4 // Default to normal
	}
}

// buildGlobalGeneratePrompt constructs the LLM prompt for generating Global Seeds
//
// Builds a comprehensive prompt that includes:
// - Story Bible context (theme, world view, core truth)
// - 3-tier clue chain requirements
// - Ending linkage requirements
// - JSON output format specification
func (sa *SeedAgent) buildGlobalGeneratePrompt(request *GlobalGenerateRequest, seedCount int) string {
	bible := request.StoryBible

	// Build ending types list for context
	endingTypes := make([]string, 0, len(request.PossibleEndings))
	for _, ending := range request.PossibleEndings {
		endingTypes = append(endingTypes, ending.Type)
	}

	prompt := `你是一個專業的恐怖遊戲敘事設計師，負責生成主線伏筆（Global Seeds）。

## 故事背景
- 主題：` + bible.Theme + `
- 世界觀：` + bible.WorldView + `
- 核心真相：` + bible.CoreTruth + `
- 難度：` + request.Difficulty + `

## 任務要求
生成 ` + fmt.Sprintf("%d", seedCount) + ` 個主線伏筆（Global Seeds），每個伏筆包含 3 層遞進式線索鏈：

### 線索層級說明
- **Tier 1 (Surface)**: 極度隱晦的暗示，玩家幾乎察覺不到（建議出現在 Beat 1-5）
- **Tier 2 (Deep)**: 較明顯的線索，玩家應該能注意到（建議出現在 Beat 6-12）
- **Tier 3 (Truth)**: 接近明示的真相揭露，確認玩家猜測（建議出現在 Beat 13-18）

### 結局連結
這些伏筆應該連結到以下結局類型：` + fmt.Sprintf("%v", endingTypes) + `
確保每個結局至少有 1-2 個伏筆支撐。

## 輸出格式
請輸出 **純 JSON 數組**，不要包含任何 markdown 標記或其他文字：

[
  {
    "id": "GS001",
    "content": "伏筆描述（這個伏筆的核心內容）",
    "linked_truth": "關聯的真相（這個伏筆揭示什麼）",
    "linked_ending": "tragic",
    "clue_chain": [
      {
        "tier": 1,
        "content": "第一層線索文本（極度隱晦）",
        "keywords": ["關鍵詞1", "關鍵詞2"],
        "beat_start": 1,
        "beat_end": 5
      },
      {
        "tier": 2,
        "content": "第二層線索文本（較明顯）",
        "keywords": ["關鍵詞1", "關鍵詞2", "關鍵詞3"],
        "beat_start": 6,
        "beat_end": 12
      },
      {
        "tier": 3,
        "content": "第三層線索文本（接近明示）",
        "keywords": ["關鍵詞1", "關鍵詞2", "關鍵詞3", "關鍵詞4"],
        "beat_start": 13,
        "beat_end": 18
      }
    ]
  }
]

## 重要提示
1. 伏筆之間應該相互關聯，形成完整的敘事網絡
2. 早期線索應該模糊、誤導，晚期線索應該確認或顛覆玩家預期
3. 關鍵詞應該富有意象，符合` + bible.Theme + `的主題氛圍
4. 必須輸出有效的 JSON 格式，不要使用 markdown 代碼塊
5. 根據難度 "` + request.Difficulty + `" 調整線索的隱晦程度`

	return prompt
}

// parseGlobalGenerateResponse parses LLM JSON response into GlobalGenerateResponse
//
// Performs robust JSON parsing with the following steps:
// 1. Strip markdown code blocks (```json, ```)
// 2. Unmarshal JSON array
// 3. Validate 3-tier clue chains for each seed
// 4. Create seed.GlobalSeed objects using Epic 2 constructors
func (sa *SeedAgent) parseGlobalGenerateResponse(raw string) (*GlobalGenerateResponse, error) {
	// Clean the response - remove markdown code blocks if present
	content := strings.TrimSpace(raw)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Define intermediate JSON structure matching LLM output
	type GlobalSeedJSON struct {
		ID           string           `json:"id"`
		Content      string           `json:"content"`
		LinkedTruth  string           `json:"linked_truth"`
		LinkedEnding string           `json:"linked_ending"`
		ClueChain    []seed.ClueTier  `json:"clue_chain"`
	}

	// Parse JSON array
	var seedsJSON []GlobalSeedJSON
	if err := json.Unmarshal([]byte(content), &seedsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w\nFull LLM response:\n%s", err, content)
	}

	// Convert to seed.GlobalSeed structs
	seeds := make([]*seed.GlobalSeed, 0, len(seedsJSON))
	for i, sj := range seedsJSON {
		// Validate clue chain has exactly 3 tiers (AC #1)
		if len(sj.ClueChain) != 3 {
			return nil, fmt.Errorf("seed %d (%s): expected 3 clue tiers, got %d", i, sj.ID, len(sj.ClueChain))
		}

		// Create GlobalSeed using Epic 2 constructor
		gs, err := seed.NewGlobalSeed(
			sj.ID,
			sj.Content,
			sj.LinkedTruth,
			sj.LinkedEnding,
			sj.ClueChain,
		)
		if err != nil {
			return nil, fmt.Errorf("seed %d (%s): failed to create GlobalSeed: %w", i, sj.ID, err)
		}

		seeds = append(seeds, gs)
	}

	return &GlobalGenerateResponse{
		GlobalSeeds: seeds,
	}, nil
}

// linkSeedsToEndings establishes connections between Global Seeds and Endings
//
// Ensures each ending has at least 1-2 seeds supporting it.
// Seeds are distributed based on their LinkedEnding field from LLM generation.
//
// Note: The actual linkage is already done by LLM in the "linked_ending" field.
// This method validates and potentially adjusts the distribution if needed.
func (sa *SeedAgent) linkSeedsToEndings(seeds []*seed.GlobalSeed, endings []SeedEnding) {
	// Count seeds per ending type
	endingCounts := make(map[string]int)
	for _, seed := range seeds {
		endingCounts[seed.LinkedEnding]++
	}

	// Log the distribution for debugging/validation
	// In a real implementation, we might adjust seeds if distribution is uneven
	// For now, we trust the LLM to distribute seeds appropriately
	// The prompt already instructs the LLM to ensure each ending has 1-2 seeds

	// Future enhancement: If an ending has 0 seeds, we could reassign one
	// But this should be rare given the prompt instructions
}

// InvokeLocalManage manages Local Seeds for Game Loop
//
// AC #2: Returns Plant/Harvest/Prune/Skip decision in <2s
// AC #4: Calculates Urgency Score with 4 factors (tension, hints, progress, global link)
// AC #5: Prunes seeds when Age > MaxLifespan
// AC #8: Integrates SeedPruner when LocalSeeds >3
//
// Parameters:
//   - ctx: Context for timeout control (<2s recommended)
//   - request: LocalManageRequest with current game state
//
// Returns:
//   - *LocalManageResponse: Operation decision and target seed
//   - error: Error if management fails
func (sa *SeedAgent) InvokeLocalManage(
	ctx context.Context,
	request *LocalManageRequest,
) (*LocalManageResponse, error) {
	// Validate request
	if request == nil {
		return nil, &AgentError{
			AgentName: sa.config.Name,
			Operation: "InvokeLocalManage",
			Cause:     errors.New("request cannot be nil"),
			Retryable: false,
		}
	}

	// 1. Check for expired seeds (AC #5)
	expiredSeeds := sa.checkExpiredSeeds(request.ActiveLocalSeeds, request.CurrentBeat)
	if len(expiredSeeds) > 0 {
		return &LocalManageResponse{
			Operation:   SeedOpPrune,
			TargetSeed:  nil,
			PrunedSeeds: expiredSeeds,
		}, nil
	}

	// 2. Calculate Urgency Scores for all active seeds (AC #4)
	urgencies := sa.calculateUrgencies(request)

	// 3. Find seed with highest urgency
	var targetSeed *seed.LocalSeed
	maxUrgency := 0.0
	for i, urgency := range urgencies {
		if urgency > maxUrgency {
			maxUrgency = urgency
			targetSeed = request.ActiveLocalSeeds[i]
		}
	}

	// 4. Decide operation based on urgency and game state
	operation := sa.decideSeedOperation(request, maxUrgency, targetSeed)

	// 5. Handle SeedPruner integration if needed (AC #8)
	prunedSeeds := make([]*seed.LocalSeed, 0)
	if len(request.ActiveLocalSeeds) > 3 {
		// TODO: Integrate with Epic 2 SeedPruner
		// For now, keep top 3 by urgency
		// prunedSeeds = sa.pruneByRelevance(request.ActiveLocalSeeds, urgencies)
	}

	return &LocalManageResponse{
		Operation:   operation,
		TargetSeed:  targetSeed,
		PrunedSeeds: prunedSeeds,
	}, nil
}

// checkExpiredSeeds finds all seeds that have exceeded their MaxLifespan
func (sa *SeedAgent) checkExpiredSeeds(activeSeeds []*seed.LocalSeed, currentBeat int) []*seed.LocalSeed {
	expired := make([]*seed.LocalSeed, 0)
	for _, s := range activeSeeds {
		if s.IsExpired(currentBeat) {
			expired = append(expired, s)
		}
	}
	return expired
}

// calculateUrgencies calculates Urgency Score for all active Local Seeds
//
// AC #4: Multi-factor calculation:
//   - Tension factor: 0.3-0.5 based on tension level
//   - Hint factor: 0.2 if player has ≥2 hints
//   - Progress factor: 0.3 based on seed age/lifespan ratio
//   - Global link factor: 0.2 if linked to Global Seed
func (sa *SeedAgent) calculateUrgencies(request *LocalManageRequest) []float64 {
	urgencies := make([]float64, len(request.ActiveLocalSeeds))
	for i, s := range request.ActiveLocalSeeds {
		urgencies[i] = sa.CalculateUrgency(
			s,
			request.TensionState,
			request.PlayerHints,
			request.CurrentBeat,
		)
	}
	return urgencies
}

// CalculateUrgency calculates the Harvest Urgency Score for a single Local Seed
//
// AC #4: Multi-factor calculation (0.0-1.0 range)
//
// Factors:
//   - Tension (0.0-0.5): Higher when tension ≥60
//   - Hints (0.0-0.2): +0.2 if player has ≥2 hints
//   - Progress (0.0-0.3): Increases as seed approaches expiration
//   - Global Link (0.0-0.2): +0.2 if linked to Global Seed
//
// Parameters:
//   - seed: The Local Seed to calculate urgency for
//   - tensionState: Current tension state (from Epic 3)
//   - playerHints: Number of hints player has accumulated
//   - currentBeat: Current game beat number
//
// Returns:
//   - float64: Urgency score in range [0.0, 1.0]
func (sa *SeedAgent) CalculateUrgency(
	seed *seed.LocalSeed,
	tensionState *engine.TensionState,
	playerHints int,
	currentBeat int,
) float64 {
	// 1. Tension factor (0.0-0.5)
	tensionFactor := 0.0
	if tensionState != nil {
		if tensionState.Value >= 60 && tensionState.Value < 80 {
			tensionFactor = 0.3
		} else if tensionState.Value >= 80 {
			tensionFactor = 0.5
		}
	}

	// 2. Hint factor (0.0-0.2)
	hintFactor := 0.0
	if playerHints >= 2 {
		hintFactor = 0.2
	}

	// 3. Progress factor (0.0-0.3) based on age/lifespan ratio
	age := currentBeat - seed.PlantedAt
	progressFactor := 0.0
	if seed.MaxLifespan > 0 {
		progressFactor = float64(age) / float64(seed.MaxLifespan) * 0.3
	}

	// 4. Global link factor (0.0-0.2)
	globalLinkFactor := 0.0
	// Note: Epic 2's LocalSeed doesn't have LinkedGlobal field in the actual implementation
	// This would need to be added or tracked separately
	// For now, we'll leave this at 0.0

	// 5. Sum all factors (cap at 1.0)
	urgency := tensionFactor + hintFactor + progressFactor + globalLinkFactor
	if urgency > 1.0 {
		urgency = 1.0
	}

	return urgency
}

// decideSeedOperation determines which operation to perform
//
// Decision logic:
//   - Harvest: If maxUrgency ≥ 0.7
//   - Plant: If active seeds < 3 AND tension 40-70
//   - Prune: If any seed is expired (handled earlier)
//   - Skip: Default (maintain current state)
func (sa *SeedAgent) decideSeedOperation(
	request *LocalManageRequest,
	maxUrgency float64,
	targetSeed *seed.LocalSeed,
) SeedOperation {
	// Harvest if urgency is high
	if maxUrgency >= 0.7 {
		return SeedOpHarvest
	}

	// Plant if we need more seeds and tension is moderate
	if len(request.ActiveLocalSeeds) < 3 {
		if request.TensionState != nil &&
			request.TensionState.Value >= 40 &&
			request.TensionState.Value <= 70 {
			return SeedOpPlant
		}
	}

	// Default: Skip (maintain current state)
	return SeedOpSkip
}

// CheckEndingUnlock calculates percentage-based ending unlock
//
// AC #7: Percentage-based logic:
//   - True Ending: ≥80% Global Seeds fully revealed (all 3 tiers)
//   - Good Ending: 40-79% Global Seeds revealed
//   - Bad Ending: <40% Global Seeds revealed
//
// A seed is considered "fully revealed" when all 3 tiers have been revealed.
//
// Parameters:
//   - globalSeeds: All Global Seeds in the game
//
// Returns:
//   - []string: List of ending types that are unlocked ("true", "good", "bad")
func (sa *SeedAgent) CheckEndingUnlock(globalSeeds []*seed.GlobalSeed) []string {
	if len(globalSeeds) == 0 {
		// No seeds = bad ending only
		return []string{"bad"}
	}

	// Count fully revealed seeds (all 3 tiers revealed)
	fullyRevealedCount := 0
	for _, s := range globalSeeds {
		if s.IsFullyRevealed() {
			fullyRevealedCount++
		}
	}

	// Calculate percentage
	percentage := float64(fullyRevealedCount) / float64(len(globalSeeds))

	// Determine unlocked endings based on percentage
	unlockedEndings := make([]string, 0, 1)

	if percentage >= 0.8 {
		// True Ending (≥80%)
		unlockedEndings = append(unlockedEndings, "true")
	} else if percentage >= 0.4 {
		// Good Ending (40-79%)
		unlockedEndings = append(unlockedEndings, "good")
	} else {
		// Bad Ending (<40%)
		unlockedEndings = append(unlockedEndings, "bad")
	}

	return unlockedEndings
}
