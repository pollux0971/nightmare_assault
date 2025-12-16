package narration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// InvokeSkeleton 調用 Skeleton 模式生成故事骨架
//
// Skeleton 模式是 Phase 1 (Genesis) 的核心，負責規劃整個故事的結構：
//   - 世界觀設定（WorldView）
//   - 核心真相（CoreTruth）
//   - 劇情結構（PlotStructure）: 三幕結構、Global Seeds、關鍵劇情點
//   - 可能結局（PossibleEndings）: 多重結局與觸發條件
//
// 參數：
//   - ctx: 上下文，用於超時控制
//   - request: Skeleton 請求，包含主題、難度、故事長度等
//
// 返回：
//   - *SkeletonResponse: 故事骨架響應
//   - error: 錯誤信息
func (a *NarrationAgent) InvokeSkeleton(ctx context.Context, request *SkeletonRequest) (*SkeletonResponse, error) {
	log.Printf("[%s] InvokeSkeleton started: Theme=%s, Difficulty=%s, Length=%s",
		a.config.Name, request.Theme, request.Difficulty, request.StoryLength)

	// Epic 4 整合：選擇模板（如果提供了 TemplateLibrary）
	var selectedTemplates *SelectedTemplates
	if request.TemplateLibrary != nil {
		var err error
		selectedTemplates, err = a.selectTemplates(request)
		if err != nil {
			log.Printf("[%s] Template selection failed: %v", a.config.Name, err)
			// 繼續執行，不阻塞流程
		}
	} else {
		log.Printf("[%s] No TemplateLibrary provided, skipping template selection", a.config.Name)
	}

	// 使用 BaseAgentImpl 的重試機制
	result, err := a.baseImpl.InvokeWithRetry(ctx, func(ctx context.Context) (any, error) {
		// 1. 構建 Prompt
		prompt, err := a.buildSkeletonPrompt(request)
		if err != nil {
			log.Printf("[%s] Failed to build prompt: %v", a.config.Name, err)
			return nil, err
		}
		log.Printf("[%s] Prompt built successfully (length: %d chars)", a.config.Name, len(prompt))

		// 2. 調用 LLM
		rawResponse, err := a.config.LLMClient.Generate(ctx, prompt, nil)
		if err != nil {
			log.Printf("[%s] LLM generation failed: %v", a.config.Name, err)
			return nil, err
		}
		log.Printf("[%s] LLM response received (length: %d chars)", a.config.Name, len(rawResponse))

		// 3. 解析響應
		response, err := parseSkeletonResponse(rawResponse)
		if err != nil {
			log.Printf("[%s] Failed to parse response: %v", a.config.Name, err)
			log.Printf("[%s] Raw response (first 200 chars): %s", a.config.Name, truncate(rawResponse, 200))
			return nil, err
		}
		log.Printf("[%s] Response parsed successfully: %d global seeds, %d endings",
			a.config.Name, len(response.PlotStructure.GlobalSeeds), len(response.PossibleEndings))

		return response, nil
	})

	if err != nil {
		log.Printf("[%s] InvokeSkeleton failed after retries: %v", a.config.Name, err)
		return nil, err
	}

	// 類型斷言
	skeletonResp, ok := result.(*SkeletonResponse)
	if !ok {
		err := fmt.Errorf("unexpected response type: %T", result)
		log.Printf("[%s] Type assertion failed: %v", a.config.Name, err)
		return nil, err
	}

	// 附加選中的模板到響應
	if selectedTemplates != nil {
		skeletonResp.SelectedRules = selectedTemplates.Rules
		skeletonResp.SelectedScene = selectedTemplates.Scene
		skeletonResp.SelectedNPCs = selectedTemplates.NPCs
		log.Printf("[%s] Attached selected templates to response", a.config.Name)
	}

	log.Printf("[%s] InvokeSkeleton completed successfully", a.config.Name)
	return skeletonResp, nil
}

// truncate 截斷字符串到指定長度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// buildSkeletonPrompt 構建 Skeleton 模式的 Prompt
//
// 此方法生成用於 LLM 的 Prompt，包含：
//   - 系統指令：定義 Agent 的角色和任務
//   - 用戶請求：主題、難度、故事長度等
//   - 輸出格式：要求返回 JSON 格式的故事骨架
//
// 參數：
//   - request: Skeleton 請求
//
// 返回：
//   - string: 構建好的 Prompt
//   - error: 錯誤信息
func (a *NarrationAgent) buildSkeletonPrompt(request *SkeletonRequest) (string, error) {
	// TODO: 後續使用 PromptLoader 加載模板
	// 當前使用簡化的內嵌 Prompt
	prompt := fmt.Sprintf(`你是「規則怪談」類型恐怖遊戲的故事架構師。

你的職責是規劃故事的骨架結構，確保：
1. 世界觀自洽且詭異
2. Global Seeds（主線伏筆）分布合理
3. 恐怖節奏有鋪墊、有高潮、有收束
4. 潛規則（Rules）可被發現但不明顯

**寫作風格**：
- 繁體中文
- 氛圍營造優先於劇情推進
- 恐怖來自未知與暗示，而非直接展示

=== 遊戲設定 ===
主題：%s
難度：%s（easy/normal/hard/hell）
故事長度：%s（short: 3-5章 / medium: 5-8章 / long: 8-15章）
18+模式：%t

=== 輸出要求 ===
請規劃完整的故事骨架，以 JSON 格式返回，包含：
1. world_view: 世界觀設定（setting, atmosphere, time_frame, background）
2. core_truth: 核心真相（truth, hidden_from, revelation）
3. plot_structure: 劇情結構（three_act, key_plot_points, global_seeds, estimated_beats）
4. possible_endings: 多重結局（至少 2-3 個，與 Seeds 綁定）

**JSON 格式範例**：
{
  "world_view": {
    "setting": "場景設定",
    "atmosphere": "氛圍描述",
    "time_frame": "時間範圍",
    "background": "背景故事（500-800字）"
  },
  "core_truth": {
    "truth": "核心真相",
    "hidden_from": "隱藏方式",
    "revelation": "揭露時機"
  },
  "plot_structure": {
    "three_act": {
      "act1": {
        "name": "Setup",
        "beat_range": [1, 2],
        "goals": ["介紹世界觀"],
        "key_events": ["序章"]
      },
      "act2": {
        "name": "Confrontation",
        "beat_range": [3, 5],
        "goals": ["張力累積"]
      },
      "act3": {
        "name": "Resolution",
        "beat_range": [6, 7],
        "goals": ["真相揭露"]
      }
    },
    "key_plot_points": [
      {
        "name": "Inciting Incident",
        "beat": 2,
        "description": "第一次異常事件"
      }
    ],
    "global_seeds": [
      {
        "id": "gs-1",
        "content": "伏筆內容",
        "linked_truth": "關聯真相",
        "linked_ending": "結局ID",
        "clue_chain": [
          {
            "tier": 1,
            "beat_range": [1, 2],
            "clue_content": "表面線索"
          },
          {
            "tier": 2,
            "beat_range": [3, 5],
            "clue_content": "深層線索"
          },
          {
            "tier": 3,
            "beat_range": [6, 7],
            "clue_content": "真相線索"
          }
        ],
        "plant_beat_range": [1, 3]
      }
    ],
    "estimated_beats": 7
  },
  "possible_endings": [
    {
      "id": "true-ending",
      "name": "True Ending",
      "condition": {
        "min_seed_percentage": 0.8,
        "max_rule_violations": 2,
        "min_hp": 30,
        "min_san": 20
      },
      "description": "結局描述（1000-1500字）",
      "required_seed_percentage": 0.8
    }
  ]
}

請確保所有元素服務於主題「%s」，避免不相關的恐怖橋段。`,
		request.Theme,
		request.Difficulty,
		request.StoryLength,
		request.Adult18Plus,
		request.Theme,
	)

	return prompt, nil
}

// parseSkeletonResponse 解析 Skeleton 響應
//
// 將 LLM 返回的 JSON 字符串解析為 SkeletonResponse 結構。
//
// 參數：
//   - raw: LLM 返回的原始字符串
//
// 返回：
//   - *SkeletonResponse: 解析後的響應
//   - error: 解析錯誤
func parseSkeletonResponse(raw string) (*SkeletonResponse, error) {
	var response SkeletonResponse

	// 嘗試解析 JSON
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONParseFailed, err)
	}

	// 驗證必填字段
	if err := validateSkeletonResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// validateSkeletonResponse 驗證 Skeleton 響應的必填字段
func validateSkeletonResponse(response *SkeletonResponse) error {
	if response.WorldView.Setting == "" {
		return fmt.Errorf("%w: world_view.setting", ErrMissingRequiredField)
	}

	if response.CoreTruth.Truth == "" {
		return fmt.Errorf("%w: core_truth.truth", ErrMissingRequiredField)
	}

	if response.PlotStructure.EstimatedBeats == 0 {
		return fmt.Errorf("%w: plot_structure.estimated_beats", ErrMissingRequiredField)
	}

	if len(response.PlotStructure.GlobalSeeds) == 0 {
		return fmt.Errorf("%w: plot_structure.global_seeds (at least 1 required)", ErrMissingRequiredField)
	}

	if len(response.PossibleEndings) == 0 {
		return fmt.Errorf("%w: possible_endings (at least 1 required)", ErrMissingRequiredField)
	}

	return nil
}
