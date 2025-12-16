package narration

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestNarrationAgentCreation 測試 NarrationAgent 創建
func TestNarrationAgentCreation(t *testing.T) {
	config := agents.AgentConfig{
		Name:       "NarrationAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  agents.NewMockLLMClient("test response"),
	}

	agent := NewNarrationAgent(config)

	if agent == nil {
		t.Fatal("Expected non-nil NarrationAgent")
	}

	if agent.GetName() != "NarrationAgent" {
		t.Errorf("Expected name 'NarrationAgent', got %s", agent.GetName())
	}

	if agent.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", agent.GetTimeout())
	}
}

// TestNarrationAgentImplementsBaseAgent 測試實現 BaseAgent 接口
func TestNarrationAgentImplementsBaseAgent(t *testing.T) {
	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   30 * time.Second,
		LLMClient: agents.NewMockLLMClient("test"),
	}

	agent := NewNarrationAgent(config)

	// 驗證實現了 BaseAgent 接口
	var _ agents.BaseAgent = agent
}

// TestNarrationModeEnum 測試 NarrationMode 枚舉
func TestNarrationModeEnum(t *testing.T) {
	modes := []NarrationMode{
		ModeSkeleton,
		ModeContent,
		ModeOpening,
		ModeEnding,
	}

	if len(modes) != 4 {
		t.Errorf("Expected 4 narration modes, got %d", len(modes))
	}

	// 驗證每個模式都有不同的值
	modeSet := make(map[NarrationMode]bool)
	for _, mode := range modes {
		if modeSet[mode] {
			t.Errorf("Duplicate narration mode: %v", mode)
		}
		modeSet[mode] = true
	}
}

// TestInvokeSkeleton 測試 Skeleton 模式調用
func TestInvokeSkeleton(t *testing.T) {
	mockClient := agents.NewMockLLMClient(`{
		"world_view": {
			"setting": "廢棄醫院",
			"atmosphere": "詭異、壓抑",
			"time_frame": "1990年代",
			"background": "這所醫院在1995年因神秘事件而關閉"
		},
		"core_truth": {
			"truth": "醫院進行非法人體實驗",
			"hidden_from": "偽裝成精神病院",
			"revelation": "第三幕 Climax"
		},
		"plot_structure": {
			"three_act": {
				"act1": {
					"name": "Setup",
					"beat_range": [1, 2],
					"goals": ["介紹世界觀"]
				},
				"act2": {
					"name": "Confrontation",
					"beat_range": [3, 5]
				},
				"act3": {
					"name": "Resolution",
					"beat_range": [6, 7]
				}
			},
			"key_plot_points": [
				{
					"name": "Inciting Incident",
					"beat": 2,
					"description": "發現異常"
				}
			],
			"global_seeds": [
				{
					"id": "gs-1",
					"content": "醫院地下室傳來奇怪聲音",
					"linked_truth": "非法實驗",
					"linked_ending": "true-ending",
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
				"description": "揭露完整真相",
				"required_seed_percentage": 0.8
			}
		],
		"selected_rules": [],
		"selected_scene": {},
		"selected_npcs": []
	}`)

	config := agents.AgentConfig{
		Name:       "NarrationAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}

	agent := NewNarrationAgent(config)

	request := &SkeletonRequest{
		Theme:       "廢棄醫院",
		Difficulty:  "normal",
		StoryLength: "medium",
		Adult18Plus: false,
	}

	response, err := agent.InvokeSkeleton(context.Background(), request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// 驗證基本字段
	if response.WorldView.Setting != "廢棄醫院" {
		t.Errorf("Expected setting '廢棄醫院', got '%s'", response.WorldView.Setting)
	}

	if response.CoreTruth.Truth != "醫院進行非法人體實驗" {
		t.Errorf("Expected truth about experiments, got '%s'", response.CoreTruth.Truth)
	}

	// 驗證 Global Seeds
	if len(response.PlotStructure.GlobalSeeds) == 0 {
		t.Error("Expected at least one global seed")
	}

	// 驗證結局
	if len(response.PossibleEndings) == 0 {
		t.Error("Expected at least one ending")
	}

	if mockClient.CallCount != 1 {
		t.Errorf("Expected 1 LLM call, got %d", mockClient.CallCount)
	}
}

// TestInvokeSkeletonWithTemplateLibrary 測試帶模板庫的 Skeleton 調用
func TestInvokeSkeletonWithTemplateLibrary(t *testing.T) {
	// 跳過此測試直到 Template Library 整合完成
	t.Skip("Template Library integration pending")
}
