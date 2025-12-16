package narration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestIntegrationSkeletonGeneration 集成測試：完整 Skeleton 生成流程
//
// 測試從請求到響應的完整流程：
//  1. 創建 NarrationAgent
//  2. 構建 SkeletonRequest
//  3. 調用 InvokeSkeleton
//  4. 驗證返回的 SkeletonResponse 完整性
func TestIntegrationSkeletonGeneration(t *testing.T) {
	// 準備：創建帶有 Mock LLM 的 Agent
	mockResponse := buildValidSkeletonJSON()
	mockClient := agents.NewMockLLMClient(mockResponse)

	config := agents.AgentConfig{
		Name:       "NarrationAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}

	agent := NewNarrationAgent(config)

	// 執行：構建請求並調用 InvokeSkeleton
	request := &SkeletonRequest{
		Theme:       "廢棄醫院",
		Difficulty:  "normal",
		StoryLength: "medium",
		Adult18Plus: false,
	}

	ctx := context.Background()
	response, err := agent.InvokeSkeleton(ctx, request)

	// 驗證：檢查響應
	if err != nil {
		t.Fatalf("Expected successful skeleton generation, got error: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// 驗證世界觀
	if response.WorldView.Setting == "" {
		t.Error("WorldView.Setting should not be empty")
	}
	if response.WorldView.Atmosphere == "" {
		t.Error("WorldView.Atmosphere should not be empty")
	}

	// 驗證核心真相
	if response.CoreTruth.Truth == "" {
		t.Error("CoreTruth.Truth should not be empty")
	}

	// 驗證劇情結構
	if response.PlotStructure.EstimatedBeats == 0 {
		t.Error("PlotStructure.EstimatedBeats should be > 0")
	}

	if len(response.PlotStructure.GlobalSeeds) == 0 {
		t.Error("PlotStructure.GlobalSeeds should not be empty")
	}

	// 驗證 Global Seeds 結構
	for i, seed := range response.PlotStructure.GlobalSeeds {
		if seed.ID == "" {
			t.Errorf("Seed %d: ID should not be empty", i)
		}
		if len(seed.ClueChain) != 3 {
			t.Errorf("Seed %d: ClueChain should have 3 tiers, got %d", i, len(seed.ClueChain))
		}
	}

	// 驗證結局
	if len(response.PossibleEndings) == 0 {
		t.Error("PossibleEndings should not be empty")
	}

	// 驗證三幕結構
	threeAct := response.PlotStructure.ThreeAct
	if threeAct.Act1.Name == "" {
		t.Error("Act1 should have a name")
	}
	if threeAct.Act2.Name == "" {
		t.Error("Act2 should have a name")
	}
	if threeAct.Act3.Name == "" {
		t.Error("Act3 should have a name")
	}

	// 驗證 Beat 範圍連續性
	if threeAct.Act2.BeatRange[0] != threeAct.Act1.BeatRange[1]+1 {
		t.Error("Act2 should start right after Act1")
	}
	if threeAct.Act3.BeatRange[0] != threeAct.Act2.BeatRange[1]+1 {
		t.Error("Act3 should start right after Act2")
	}
}

// TestIntegrationFallbackStrategy 集成測試：LLM 失敗時的降級策略
//
// 測試當 LLM 調用失敗時，系統能否正確切換到降級骨架：
//  1. 創建會失敗的 Mock LLM
//  2. 調用 InvokeSkeleton
//  3. 驗證系統返回降級骨架而不是錯誤
func TestIntegrationFallbackStrategy(t *testing.T) {
	t.Skip("需要在 skeleton_mode.go 中實作 fallback 邏輯")

	// TODO: 當實作 fallback 整合後，此測試應該通過
	// 預期行為：
	// 1. LLM 失敗 -> 自動切換到 BuildFallbackSkeleton
	// 2. 返回降級骨架而不是錯誤
	// 3. 降級骨架應包含所有必要字段
}

// TestIntegrationSkeletonGenerationAllDifficulties 集成測試：所有難度的 Skeleton 生成
//
// 測試不同難度下的完整 Skeleton 生成流程
func TestIntegrationSkeletonGenerationAllDifficulties(t *testing.T) {
	difficulties := []string{"easy", "normal", "hard", "hell"}

	for _, difficulty := range difficulties {
		t.Run(difficulty, func(t *testing.T) {
			mockResponse := buildValidSkeletonJSON()
			mockClient := agents.NewMockLLMClient(mockResponse)

			config := agents.AgentConfig{
				Name:       "NarrationAgent",
				Timeout:    10 * time.Second,
				MaxRetries: 3,
				LLMClient:  mockClient,
			}

			agent := NewNarrationAgent(config)

			request := &SkeletonRequest{
				Theme:       "恐怖主題",
				Difficulty:  difficulty,
				StoryLength: "medium",
				Adult18Plus: false,
			}

			ctx := context.Background()
			response, err := agent.InvokeSkeleton(ctx, request)

			if err != nil {
				t.Fatalf("Expected successful generation for %s difficulty, got error: %v", difficulty, err)
			}

			if response == nil {
				t.Fatalf("Expected non-nil response for %s difficulty", difficulty)
			}

			// 驗證基本結構完整
			if response.WorldView.Setting == "" {
				t.Errorf("%s: WorldView.Setting should not be empty", difficulty)
			}
			if response.CoreTruth.Truth == "" {
				t.Errorf("%s: CoreTruth.Truth should not be empty", difficulty)
			}
			if len(response.PlotStructure.GlobalSeeds) == 0 {
				t.Errorf("%s: GlobalSeeds should not be empty", difficulty)
			}
			if len(response.PossibleEndings) == 0 {
				t.Errorf("%s: PossibleEndings should not be empty", difficulty)
			}
		})
	}
}

// TestIntegrationSkeletonGenerationAllLengths 集成測試：所有故事長度的 Skeleton 生成
//
// 測試不同故事長度下的完整 Skeleton 生成流程
func TestIntegrationSkeletonGenerationAllLengths(t *testing.T) {
	lengths := []string{"short", "medium", "long"}

	for _, length := range lengths {
		t.Run(length, func(t *testing.T) {
			mockResponse := buildValidSkeletonJSON()
			mockClient := agents.NewMockLLMClient(mockResponse)

			config := agents.AgentConfig{
				Name:       "NarrationAgent",
				Timeout:    10 * time.Second,
				MaxRetries: 3,
				LLMClient:  mockClient,
			}

			agent := NewNarrationAgent(config)

			request := &SkeletonRequest{
				Theme:       "恐怖主題",
				Difficulty:  "normal",
				StoryLength: length,
				Adult18Plus: false,
			}

			ctx := context.Background()
			response, err := agent.InvokeSkeleton(ctx, request)

			if err != nil {
				t.Fatalf("Expected successful generation for %s length, got error: %v", length, err)
			}

			if response == nil {
				t.Fatalf("Expected non-nil response for %s length", length)
			}

			// 驗證 EstimatedBeats 在合理範圍內
			// 注意：使用 Mock LLM 時，響應是固定的，無法動態調整
			// 真實場景中，LLM 會根據請求的 StoryLength 調整 EstimatedBeats
			if response.PlotStructure.EstimatedBeats <= 0 {
				t.Errorf("%s: EstimatedBeats should be > 0, got %d",
					length, response.PlotStructure.EstimatedBeats)
			}

			// 驗證 EstimatedBeats 與 Act3 的 BeatRange 一致
			act3End := response.PlotStructure.ThreeAct.Act3.BeatRange[1]
			if response.PlotStructure.EstimatedBeats != act3End {
				t.Errorf("%s: EstimatedBeats (%d) should match Act3 end (%d)",
					length, response.PlotStructure.EstimatedBeats, act3End)
			}
		})
	}
}

// TestIntegrationPromptGeneration 集成測試：Prompt 生成邏輯
//
// 測試 BuildPrompt 方法生成的 prompt 是否符合預期格式和內容
func TestIntegrationPromptGeneration(t *testing.T) {
	mockClient := agents.NewMockLLMClient("test")

	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: mockClient,
	}

	agent := NewNarrationAgent(config)

	request := &SkeletonRequest{
		Theme:       "廢棄學校",
		Difficulty:  "hard",
		StoryLength: "long",
		Adult18Plus: true,
	}

	prompt, err := agent.BuildPrompt(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// 驗證 prompt 包含關鍵信息
	requiredStrings := []string{
		"廢棄學校",    // 主題
		"hard",      // 難度
		"long",      // 長度
		"世界觀",      // 結構要求
		"核心真相",     // 結構要求
		"Global Seed", // 結構要求
		"JSON",      // 格式要求
	}

	for _, required := range requiredStrings {
		if !strings.Contains(prompt, required) {
			t.Errorf("Prompt should contain '%s', but it doesn't", required)
		}
	}

	// 驗證 Adult18Plus 影響
	if !strings.Contains(prompt, "18+") && !strings.Contains(prompt, "成人") {
		t.Error("Prompt should indicate adult content when Adult18Plus is true")
	}
}

// TestIntegrationContextCancellation 集成測試：Context 取消處理
//
// 測試當 context 被取消時，Agent 能否正確處理
func TestIntegrationContextCancellation(t *testing.T) {
	// 創建一個會阻塞的 Mock LLM
	slowClient := &slowMockLLMClient{
		delay: 5 * time.Second,
	}

	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   10 * time.Second,
		LLMClient: slowClient,
	}

	agent := NewNarrationAgent(config)

	request := &SkeletonRequest{
		Theme:       "test",
		Difficulty:  "normal",
		StoryLength: "medium",
	}

	// 創建會被取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := agent.InvokeSkeleton(ctx, request)

	if err == nil {
		t.Error("Expected error when context is cancelled")
	}

	// 驗證錯誤類型
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// TestIntegrationInvokeWithInvalidRequestType 集成測試：錯誤的請求類型
//
// 測試當傳入不支援的請求類型時，系統能否正確處理
func TestIntegrationInvokeWithInvalidRequestType(t *testing.T) {
	mockClient := agents.NewMockLLMClient("test")

	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: mockClient,
	}

	agent := NewNarrationAgent(config)

	// 傳入不支援的請求類型
	invalidRequest := "this is not a valid request"

	_, err := agent.Invoke(context.Background(), invalidRequest)

	if err == nil {
		t.Error("Expected error for invalid request type")
	}

	// 驗證錯誤內容
	if !strings.Contains(err.Error(), "unsupported") && !strings.Contains(err.Error(), "不支援") {
		t.Errorf("Expected unsupported request type error, got: %v", err)
	}
}

// --- Helper Types and Functions ---

// slowMockLLMClient 是一個會延遲的 Mock LLM Client，用於測試 context 取消
type slowMockLLMClient struct {
	delay time.Duration
}

func (m *slowMockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	select {
	case <-time.After(m.delay):
		return buildValidSkeletonJSON(), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (m *slowMockLLMClient) GetProviderName() string {
	return "slow-mock"
}

// buildValidSkeletonJSON 構建一個有效的 Skeleton JSON 響應
func buildValidSkeletonJSON() string {
	return `{
		"world_view": {
			"setting": "廢棄醫院，現代背景",
			"atmosphere": "詭異、壓抑、死亡氣息濃厚",
			"time_frame": "現代，深夜時分",
			"background": "這座醫院曾因醫療事故被迫關閉，但關閉前有一系列不可告人的秘密實驗。"
		},
		"core_truth": {
			"truth": "醫院地下室進行著人體實驗，受試者的靈魂被困在建築中",
			"hidden_from": "醫院表面上只是普通的廢棄建築",
			"revelation": "在第三幕 Climax 時，玩家將發現地下室的真相"
		},
		"plot_structure": {
			"three_act": {
				"act1": {
					"name": "Setup（設定）",
					"beat_range": [1, 2],
					"goals": ["介紹世界觀", "建立詭異氛圍"],
					"key_events": ["進入醫院", "第一次異常事件"]
				},
				"act2": {
					"name": "Confrontation（對抗）",
					"beat_range": [3, 5],
					"goals": ["張力累積", "線索揭露"],
					"key_events": ["衝突升級", "重大發現"]
				},
				"act3": {
					"name": "Resolution（解決）",
					"beat_range": [6, 7],
					"goals": ["真相揭露", "結局收束"],
					"key_events": ["最終對抗", "結局分支"]
				}
			},
			"key_plot_points": [
				{
					"name": "Inciting Incident",
					"beat": 2,
					"description": "第一次異常事件"
				},
				{
					"name": "Midpoint",
					"beat": 4,
					"description": "重大發現"
				},
				{
					"name": "Climax",
					"beat": 6,
					"description": "真相揭露"
				}
			],
			"global_seeds": [
				{
					"id": "gs-1",
					"content": "病歷記錄異常",
					"linked_truth": "人體實驗記錄",
					"linked_ending": "true-ending",
					"clue_chain": [
						{
							"tier": 1,
							"beat_range": [1, 2],
							"clue_content": "表面線索：發現奇怪的病歷"
						},
						{
							"tier": 2,
							"beat_range": [3, 5],
							"clue_content": "深層線索：病歷中的實驗代號"
						},
						{
							"tier": 3,
							"beat_range": [6, 7],
							"clue_content": "真相線索：完整的實驗記錄"
						}
					],
					"plant_beat_range": [1, 3]
				},
				{
					"id": "gs-2",
					"content": "地下室入口",
					"linked_truth": "實驗室所在地",
					"linked_ending": "true-ending",
					"clue_chain": [
						{
							"tier": 1,
							"beat_range": [1, 2],
							"clue_content": "表面線索：牆上的標記"
						},
						{
							"tier": 2,
							"beat_range": [3, 5],
							"clue_content": "深層線索：隱藏的電梯"
						},
						{
							"tier": 3,
							"beat_range": [6, 7],
							"clue_content": "真相線索：進入地下實驗室"
						}
					],
					"plant_beat_range": [1, 3]
				},
				{
					"id": "gs-3",
					"content": "失踪者名單",
					"linked_truth": "實驗受試者",
					"linked_ending": "good-ending",
					"clue_chain": [
						{
							"tier": 1,
							"beat_range": [1, 2],
							"clue_content": "表面線索：牆上的失踪告示"
						},
						{
							"tier": 2,
							"beat_range": [3, 5],
							"clue_content": "深層線索：失踪者都曾來過醫院"
						},
						{
							"tier": 3,
							"beat_range": [6, 7],
							"clue_content": "真相線索：他們都是實驗品"
						}
					],
					"plant_beat_range": [1, 3]
				},
				{
					"id": "gs-4",
					"content": "詭異的醫療器械",
					"linked_truth": "非法實驗工具",
					"linked_ending": "good-ending",
					"clue_chain": [
						{
							"tier": 1,
							"beat_range": [1, 2],
							"clue_content": "表面線索：奇怪的手術台"
						},
						{
							"tier": 2,
							"beat_range": [3, 5],
							"clue_content": "深層線索：不明用途的儀器"
						},
						{
							"tier": 3,
							"beat_range": [6, 7],
							"clue_content": "真相線索：人體改造設備"
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
				"name": "True Ending（真實結局）",
				"condition": {
					"min_seed_percentage": 0.8,
					"max_rule_violations": 2,
					"min_hp": 30,
					"min_san": 20
				},
				"description": "玩家揭露完整真相，成功阻止實驗並解救受害者",
				"required_seed_percentage": 0.8
			},
			{
				"id": "good-ending",
				"name": "Good Ending（良好結局）",
				"condition": {
					"min_seed_percentage": 0.4,
					"max_rule_violations": 5,
					"min_hp": 10,
					"min_san": 10
				},
				"description": "玩家部分揭露真相，成功逃離醫院",
				"required_seed_percentage": 0.4
			},
			{
				"id": "bad-ending",
				"name": "Bad Ending（悲劇結局）",
				"condition": {
					"min_seed_percentage": 0.0,
					"max_rule_violations": 999,
					"min_hp": 0,
					"min_san": 0
				},
				"description": "玩家未能理解真相，成為實驗品的一員",
				"required_seed_percentage": 0.0
			}
		]
	}`
}
