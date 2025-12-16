package narration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestNarrationModeString 測試 NarrationMode.String() 方法
func TestNarrationModeString(t *testing.T) {
	tests := []struct {
		mode     NarrationMode
		expected string
	}{
		{ModeSkeleton, "Skeleton"},
		{ModeContent, "Content"},
		{ModeOpening, "Opening"},
		{ModeEnding, "Ending"},
		{NarrationMode(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.mode.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestInvokeUnsupportedRequestType 測試不支援的請求類型
func TestInvokeUnsupportedRequestType(t *testing.T) {
	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: agents.NewMockLLMClient("test"),
	}

	agent := NewNarrationAgent(config)

	// 傳入不支援的請求類型
	_, err := agent.Invoke(context.Background(), "invalid request type")

	if err == nil {
		t.Error("Expected error for unsupported request type")
	}

	var agentErr *agents.AgentError
	if !errors.As(err, &agentErr) {
		t.Error("Expected AgentError type")
	}

	if agentErr.Retryable {
		t.Error("Unsupported request type should not be retryable")
	}

	if !errors.Is(err, ErrUnsupportedRequestType) {
		t.Error("Expected ErrUnsupportedRequestType")
	}
}

// TestBuildPrompt 測試 BuildPrompt 方法
func TestBuildPrompt(t *testing.T) {
	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: agents.NewMockLLMClient("test"),
	}

	agent := NewNarrationAgent(config)

	request := &SkeletonRequest{
		Theme:       "廢棄醫院",
		Difficulty:  "normal",
		StoryLength: "medium",
		Adult18Plus: false,
	}

	prompt, err := agent.BuildPrompt(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// 驗證 prompt 包含主題
	if !contains(prompt, "廢棄醫院") {
		t.Error("Prompt should contain theme")
	}

	// 驗證 prompt 包含難度
	if !contains(prompt, "normal") {
		t.Error("Prompt should contain difficulty")
	}
}

// TestBuildPromptUnsupportedType 測試不支援類型的 BuildPrompt
func TestBuildPromptUnsupportedType(t *testing.T) {
	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: agents.NewMockLLMClient("test"),
	}

	agent := NewNarrationAgent(config)

	_, err := agent.BuildPrompt("invalid")

	if !errors.Is(err, ErrUnsupportedRequestType) {
		t.Error("Expected ErrUnsupportedRequestType")
	}
}

// TestParseResponse 測試 ParseResponse 方法
func TestParseResponse(t *testing.T) {
	config := agents.AgentConfig{
		Name:      "NarrationAgent",
		Timeout:   5 * time.Second,
		LLMClient: agents.NewMockLLMClient("test"),
	}

	agent := NewNarrationAgent(config)

	rawJSON := `{
		"world_view": {
			"setting": "Test Setting",
			"atmosphere": "Test Atmosphere",
			"time_frame": "Test Time",
			"background": "Test Background"
		},
		"core_truth": {
			"truth": "Test Truth",
			"hidden_from": "Test Hidden",
			"revelation": "Test Revelation"
		},
		"plot_structure": {
			"three_act": {
				"act1": {"name": "Act1", "beat_range": [1, 2]},
				"act2": {"name": "Act2", "beat_range": [3, 5]},
				"act3": {"name": "Act3", "beat_range": [6, 7]}
			},
			"key_plot_points": [],
			"global_seeds": [
				{
					"id": "gs-1",
					"content": "Test Seed",
					"linked_truth": "Truth",
					"linked_ending": "ending-1",
					"clue_chain": [
						{"tier": 1, "beat_range": [1, 2], "clue_content": "Clue 1"}
					],
					"plant_beat_range": [1, 3]
				}
			],
			"estimated_beats": 7
		},
		"possible_endings": [
			{
				"id": "ending-1",
				"name": "Ending",
				"condition": {"min_seed_percentage": 0.8},
				"description": "Test Ending",
				"required_seed_percentage": 0.8
			}
		]
	}`

	response, err := agent.ParseResponse(rawJSON)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	skeletonResp, ok := response.(*SkeletonResponse)
	if !ok {
		t.Fatal("Expected SkeletonResponse type")
	}

	if skeletonResp.WorldView.Setting != "Test Setting" {
		t.Error("Failed to parse world view")
	}
}

// TestValidateSkeletonResponseErrors 測試各種驗證錯誤
func TestValidateSkeletonResponseErrors(t *testing.T) {
	tests := []struct {
		name     string
		response *SkeletonResponse
		wantErr  bool
		errType  error
	}{
		{
			name: "missing setting",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "", // 缺少
				},
				CoreTruth: CoreTruth{
					Truth: "truth",
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 7,
					GlobalSeeds:    []GlobalSeedBlueprint{{ID: "gs-1"}},
				},
				PossibleEndings: []Ending{{ID: "e-1"}},
			},
			wantErr: true,
			errType: ErrMissingRequiredField,
		},
		{
			name: "missing truth",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "setting",
				},
				CoreTruth: CoreTruth{
					Truth: "", // 缺少
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 7,
					GlobalSeeds:    []GlobalSeedBlueprint{{ID: "gs-1"}},
				},
				PossibleEndings: []Ending{{ID: "e-1"}},
			},
			wantErr: true,
			errType: ErrMissingRequiredField,
		},
		{
			name: "missing estimated beats",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "setting",
				},
				CoreTruth: CoreTruth{
					Truth: "truth",
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 0, // 缺少
					GlobalSeeds:    []GlobalSeedBlueprint{{ID: "gs-1"}},
				},
				PossibleEndings: []Ending{{ID: "e-1"}},
			},
			wantErr: true,
			errType: ErrMissingRequiredField,
		},
		{
			name: "missing global seeds",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "setting",
				},
				CoreTruth: CoreTruth{
					Truth: "truth",
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 7,
					GlobalSeeds:    []GlobalSeedBlueprint{}, // 缺少
				},
				PossibleEndings: []Ending{{ID: "e-1"}},
			},
			wantErr: true,
			errType: ErrMissingRequiredField,
		},
		{
			name: "missing endings",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "setting",
				},
				CoreTruth: CoreTruth{
					Truth: "truth",
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 7,
					GlobalSeeds:    []GlobalSeedBlueprint{{ID: "gs-1"}},
				},
				PossibleEndings: []Ending{}, // 缺少
			},
			wantErr: true,
			errType: ErrMissingRequiredField,
		},
		{
			name: "valid response",
			response: &SkeletonResponse{
				WorldView: WorldView{
					Setting: "setting",
				},
				CoreTruth: CoreTruth{
					Truth: "truth",
				},
				PlotStructure: PlotStructure{
					EstimatedBeats: 7,
					GlobalSeeds:    []GlobalSeedBlueprint{{ID: "gs-1"}},
				},
				PossibleEndings: []Ending{{ID: "e-1"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkeletonResponse(tt.response)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
			}
		})
	}
}

// TestParseSkeletonResponseInvalidJSON 測試無效 JSON
func TestParseSkeletonResponseInvalidJSON(t *testing.T) {
	_, err := parseSkeletonResponse("invalid json {{{")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !errors.Is(err, ErrJSONParseFailed) {
		t.Error("Expected ErrJSONParseFailed")
	}
}

// TestInvokeSkeletonWithError 測試 LLM 錯誤情況
func TestInvokeSkeletonWithError(t *testing.T) {
	mockClient := agents.NewMockLLMClientWithError(errors.New("LLM error"))

	config := agents.AgentConfig{
		Name:       "NarrationAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 1,
		LLMClient:  mockClient,
	}

	agent := NewNarrationAgent(config)

	request := &SkeletonRequest{
		Theme:       "test",
		Difficulty:  "normal",
		StoryLength: "medium",
	}

	_, err := agent.InvokeSkeleton(context.Background(), request)

	if err == nil {
		t.Error("Expected error when LLM fails")
	}
}

// contains 檢查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
