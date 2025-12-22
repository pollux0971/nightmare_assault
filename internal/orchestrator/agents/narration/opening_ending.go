package narration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// InvokeOpening 生成序章敘事（Genesis Phase）
//
// 功能：
//   - 介紹 WorldView（場景設定、時間框架、背景故事）
//   - 建立氛圍（根據 Atmosphere）
//   - 首次規則暗示（根據難度控制隱晦程度）
//   - 設定初始張力值（10-20）
//   - 引導第一個選擇
//
// 參數：
//   - ctx: 上下文，用於超時控制
//   - request: 包含 StoryBible 和難度等級
//
// 返回：
//   - *OpeningResponse: 序章敘事響應（800-1200 字）
//   - error: 錯誤信息
//
// AC #1: 序章應該是 800-1200 字，初始張力 10-20
func (a *NarrationAgent) InvokeOpening(ctx context.Context, request OpeningRequest) (*OpeningResponse, error) {
	log.Printf("[%s] InvokeOpening started: Difficulty=%s", a.config.Name, request.Difficulty)

	// 驗證輸入
	if request.StoryBible == nil {
		return nil, fmt.Errorf("StoryBible cannot be nil")
	}

	// 使用 BaseAgentImpl 的重試機制
	result, err := a.baseImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		// 1. 構建 Prompt
		prompt := a.buildOpeningPrompt(request)
		log.Printf("[%s] Opening prompt built (length: %d chars)", a.config.Name, len(prompt))

		// 2. 調用 LLM
		response, err := a.config.LLMClient.Generate(ctx, prompt, nil)
		if err != nil {
			log.Printf("[%s] LLM call failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// 3. 解析 JSON 響應
		var openingResp OpeningResponse
		if err := json.Unmarshal([]byte(response), &openingResp); err != nil {
			log.Printf("[%s] Failed to parse JSON: %v", a.config.Name, err)
			return nil, fmt.Errorf("failed to parse opening response: %w", err)
		}

		// 4. 驗證響應
		if err := a.validateOpeningResponse(&openingResp); err != nil {
			log.Printf("[%s] Validation failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("invalid opening response: %w", err)
		}

		log.Printf("[%s] Opening generated successfully (length: %d chars, tension: %d)",
			a.config.Name, len([]rune(openingResp.OpeningNarrative)), openingResp.InitialTension)

		return &openingResp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*OpeningResponse), nil
}

// InvokeEnding 生成結局敘事（Convergence Phase）
//
// 功能：
//   - 自動判定結局類型（如果未指定）：基於 Global Seeds 揭露程度
//     - True Ending: ≥80% Global Seeds 完全揭露
//     - Good Ending: 40-79% Global Seeds 揭露
//     - Bad Ending: <40% Global Seeds 揭露
//   - 生成結局敘事（1000-1500 字）
//   - 整合玩家最終狀態（HP, SAN, 選擇歷史）
//   - 情感解析與收尾
//
// 參數：
//   - ctx: 上下文，用於超時控制
//   - request: 包含 GameState 和可選的 EndingType
//
// 返回：
//   - *EndingResponse: 結局敘事響應
//   - error: 錯誤信息
//
// AC #6: 自動判定結局類型
// AC #7: True Ending 1300-1500 字
// AC #8: Good Ending 1100-1300 字
// AC #9: Bad Ending 1000-1200 字
func (a *NarrationAgent) InvokeEnding(ctx context.Context, request EndingRequest) (*EndingResponse, error) {
	log.Printf("[%s] InvokeEnding started: EndingType=%s", a.config.Name, request.EndingType)

	// 驗證輸入
	if request.GameState == nil {
		return nil, fmt.Errorf("GameState cannot be nil")
	}

	// 自動判定結局類型（如果未指定）
	endingType := request.EndingType
	if endingType == "" {
		endingType = a.determineEndingType(request.GameState)
		log.Printf("[%s] Auto-determined ending type: %s", a.config.Name, endingType)
	}

	// 使用 BaseAgentImpl 的重試機制
	result, err := a.baseImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		// 1. 構建 Prompt
		prompt := a.buildEndingPrompt(request, endingType)
		log.Printf("[%s] Ending prompt built (length: %d chars)", a.config.Name, len(prompt))

		// 2. 調用 LLM
		response, err := a.config.LLMClient.Generate(ctx, prompt, nil)
		if err != nil {
			log.Printf("[%s] LLM call failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// 3. 解析 JSON 響應
		var endingResp EndingResponse
		if err := json.Unmarshal([]byte(response), &endingResp); err != nil {
			log.Printf("[%s] Failed to parse JSON: %v", a.config.Name, err)
			return nil, fmt.Errorf("failed to parse ending response: %w", err)
		}

		// 4. 驗證響應
		if err := a.validateEndingResponse(&endingResp, endingType); err != nil {
			log.Printf("[%s] Validation failed: %v", a.config.Name, err)
			return nil, fmt.Errorf("invalid ending response: %w", err)
		}

		log.Printf("[%s] Ending generated successfully (type: %s, length: %d chars, emotion: %s)",
			a.config.Name, endingType, len([]rune(endingResp.EndingNarrative)), endingResp.FinalEmotion)

		return &endingResp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*EndingResponse), nil
}

// ==========================================================================
// Helper Methods - Opening
// ==========================================================================

// buildOpeningPrompt 構建序章生成 Prompt
func (a *NarrationAgent) buildOpeningPrompt(request OpeningRequest) string {
	var sb strings.Builder

	sb.WriteString("你是一個專業的恐怖遊戲敘事 AI，負責生成引人入勝的序章。\n\n")

	// 世界觀信息
	sb.WriteString("## 世界觀設定\n")
	sb.WriteString(fmt.Sprintf("- 場景：%s\n", request.StoryBible.WorldView.Setting))
	sb.WriteString(fmt.Sprintf("- 氛圍：%s\n", request.StoryBible.WorldView.Atmosphere))
	sb.WriteString(fmt.Sprintf("- 時間：%s\n", request.StoryBible.WorldView.TimeFrame))
	sb.WriteString(fmt.Sprintf("- 背景：%s\n\n", request.StoryBible.WorldView.Background))

	// 核心真相（隱藏）
	sb.WriteString("## 核心真相（不要直接揭露）\n")
	sb.WriteString(fmt.Sprintf("- 真相：%s\n", request.StoryBible.CoreTruth.Truth))
	sb.WriteString(fmt.Sprintf("- 隱藏方式：%s\n\n", request.StoryBible.CoreTruth.HiddenFrom))

	// 難度級別
	sb.WriteString(fmt.Sprintf("## 難度級別：%s\n", request.Difficulty))
	sb.WriteString(a.getDifficultyGuidance(request.Difficulty))
	sb.WriteString("\n")

	// Story 7.6: NPC 介紹
	if len(request.NPCs) > 0 {
		sb.WriteString("## NPC 隊友\n")
		sb.WriteString("以下 NPC 將在序章中自然地登場，請將他們的介紹融入敘事中：\n\n")
		for i, npc := range request.NPCs {
			sb.WriteString(fmt.Sprintf("**NPC %d: %s**\n", i+1, npc.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", npc.Introduction))
		}
		sb.WriteString("注意：請將 NPC 介紹自然地編織進序章敘事中，不要生硬地羅列。\n\n")
	}

	// 輸出要求
	sb.WriteString("## 輸出要求\n")
	sb.WriteString("請生成一個 JSON 格式的序章，包含以下字段：\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"opening_narrative\": \"序章敘事文本（800-1200 字）\",\n")
	sb.WriteString("  \"initial_tension\": 15,  // 初始張力值（10-20）\n")
	sb.WriteString("  \"first_choice_prompt\": \"引導第一個選擇的文本（50 字內）\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("注意事項：\n")
	sb.WriteString("1. 序章應該介紹場景、建立氛圍、給出首次規則暗示\n")
	sb.WriteString("2. 規則暗示的隱晦程度應該符合難度設定\n")
	sb.WriteString("3. 初始張力值應該在 10-20 之間\n")
	sb.WriteString("4. 敘事應該引人入勝，但不要直接揭露核心真相\n")
	sb.WriteString("5. 必須返回有效的 JSON 格式\n")

	return sb.String()
}

// getDifficultyGuidance 根據難度返回指導文本
func (a *NarrationAgent) getDifficultyGuidance(difficulty string) string {
	switch difficulty {
	case "easy":
		return "- 規則暗示應該明顯且直接\n- 危險提示應該清晰"
	case "normal":
		return "- 規則暗示應該適度隱晦\n- 危險提示應該含蓄"
	case "hard":
		return "- 規則暗示應該非常隱晦\n- 危險提示應該微妙"
	case "hell":
		return "- 規則暗示應該極度隱晦\n- 危險提示應該幾乎不可察覺"
	default:
		return "- 規則暗示應該適度隱晦\n- 危險提示應該含蓄"
	}
}

// validateOpeningResponse 驗證序章響應
func (a *NarrationAgent) validateOpeningResponse(resp *OpeningResponse) error {
	// 檢查敘事長度（簡化驗證，實際應該檢查 800-1200 字）
	runeCount := len([]rune(resp.OpeningNarrative))
	if runeCount < 100 {
		return fmt.Errorf("opening narrative too short: %d characters", runeCount)
	}

	// 檢查張力值範圍
	if resp.InitialTension < 10 || resp.InitialTension > 20 {
		return fmt.Errorf("initial tension out of range: %d (expected 10-20)", resp.InitialTension)
	}

	// 檢查首個選擇提示
	if strings.TrimSpace(resp.FirstChoice) == "" {
		return fmt.Errorf("first choice prompt is empty")
	}

	return nil
}

// ==========================================================================
// Helper Methods - Ending
// ==========================================================================

// determineEndingType 根據 Global Seeds 揭露程度判定結局類型
//
// AC #6: 自動判定結局類型
//   - True Ending: ≥80% Global Seeds 完全揭露（Tier 1, 2, 3）
//   - Good Ending: 40-79% Global Seeds 揭露
//   - Bad Ending: <40% Global Seeds 揭露
func (a *NarrationAgent) determineEndingType(gameState *engine.GameStateV2) string {
	// TODO: 實際實現需要計算 Global Seeds 揭露百分比
	// 目前使用簡化邏輯：基於 SAN 值判定
	san := gameState.GetSAN()

	if san >= 70 {
		return EndingTrue
	} else if san >= 40 {
		return EndingGood
	}
	return EndingBad
}

// buildEndingPrompt 構建結局生成 Prompt
func (a *NarrationAgent) buildEndingPrompt(request EndingRequest, endingType string) string {
	var sb strings.Builder

	sb.WriteString("你是一個專業的恐怖遊戲敘事 AI，負責生成震撼人心的結局。\n\n")

	// 玩家狀態
	sb.WriteString("## 玩家最終狀態\n")
	sb.WriteString(fmt.Sprintf("- HP: %d\n", request.GameState.GetHP()))
	sb.WriteString(fmt.Sprintf("- SAN: %d\n\n", request.GameState.GetSAN()))

	// 結局類型
	sb.WriteString(fmt.Sprintf("## 結局類型：%s\n", endingType))
	sb.WriteString(a.getEndingTypeGuidance(endingType))
	sb.WriteString("\n")

	// 輸出要求
	sb.WriteString("## 輸出要求\n")
	sb.WriteString("請生成一個 JSON 格式的結局，包含以下字段：\n")
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"ending_narrative\": \"結局敘事文本（%s）\",\n", a.getEndingLengthGuidance(endingType)))
	sb.WriteString("  \"final_emotion\": \"shock\",  // 最終情感（shock/relief/despair）\n")
	sb.WriteString("  \"closing_line\": \"結局金句（最後一句話）\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("注意事項：\n")
	sb.WriteString("1. 結局應該根據玩家狀態和結局類型生成\n")
	sb.WriteString("2. 敘事應該震撼人心，給玩家留下深刻印象\n")
	sb.WriteString("3. 結局金句應該簡潔有力\n")
	sb.WriteString("4. 必須返回有效的 JSON 格式\n")

	return sb.String()
}

// getEndingTypeGuidance 根據結局類型返回指導文本
func (a *NarrationAgent) getEndingTypeGuidance(endingType string) string {
	switch endingType {
	case EndingTrue:
		return "- 揭露所有核心真相\n- 情感應該是震撼、恍然大悟\n- 長度：1300-1500 字"
	case EndingGood:
		return "- 揭露部分真相\n- 情感應該是解脫但疑惑\n- 長度：1100-1300 字"
	case EndingBad:
		return "- 真相仍然迷霧重重\n- 情感應該是絕望、無助\n- 長度：1000-1200 字"
	default:
		return "- 長度：1000-1200 字"
	}
}

// getEndingLengthGuidance 根據結局類型返回長度指導
func (a *NarrationAgent) getEndingLengthGuidance(endingType string) string {
	switch endingType {
	case EndingTrue:
		return "1300-1500 字"
	case EndingGood:
		return "1100-1300 字"
	case EndingBad:
		return "1000-1200 字"
	default:
		return "1000-1200 字"
	}
}

// validateEndingResponse 驗證結局響應
func (a *NarrationAgent) validateEndingResponse(resp *EndingResponse, endingType string) error {
	// 檢查敘事長度（簡化驗證）
	runeCount := len([]rune(resp.EndingNarrative))
	if runeCount < 100 {
		return fmt.Errorf("ending narrative too short: %d characters", runeCount)
	}

	// 檢查情感
	validEmotions := map[string]bool{
		"shock":   true,
		"relief":  true,
		"despair": true,
	}
	if !validEmotions[resp.FinalEmotion] {
		return fmt.Errorf("invalid final emotion: %s (expected shock/relief/despair)", resp.FinalEmotion)
	}

	// 檢查結局金句
	if strings.TrimSpace(resp.ClosingLine) == "" {
		return fmt.Errorf("closing line is empty")
	}

	return nil
}
