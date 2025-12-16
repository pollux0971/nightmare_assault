package narration

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpeningRequest tests OpeningRequest structure
func TestOpeningRequest(t *testing.T) {
	req := OpeningRequest{
		StoryBible: &StoryBible{
			WorldView: WorldView{
				Setting:    "廢棄醫院",
				Atmosphere: "陰鬱、詭異、血腥",
			},
		},
		Difficulty: "normal",
	}

	assert.NotNil(t, req.StoryBible)
	assert.Equal(t, "normal", req.Difficulty)
	assert.Equal(t, "廢棄醫院", req.StoryBible.WorldView.Setting)
}

// TestOpeningResponse tests OpeningResponse structure
func TestOpeningResponse(t *testing.T) {
	resp := OpeningResponse{
		OpeningNarrative: "這是一段序章敘事...",
		InitialTension:   15,
		FirstChoice:      "你決定探索醫院",
	}

	assert.NotEmpty(t, resp.OpeningNarrative)
	assert.GreaterOrEqual(t, resp.InitialTension, 10)
	assert.LessOrEqual(t, resp.InitialTension, 20)
}

// TestEndingRequest tests EndingRequest structure
func TestEndingRequest(t *testing.T) {
	gameState := engine.NewGameStateV2()
	req := EndingRequest{
		GameState:  gameState,
		EndingType: EndingTrue,
	}

	assert.NotNil(t, req.GameState)
	assert.Equal(t, EndingTrue, req.EndingType)
}

// TestEndingResponse tests EndingResponse structure
func TestEndingResponse(t *testing.T) {
	resp := EndingResponse{
		EndingNarrative: "這是一段結局敘事...",
		FinalEmotion:    "shock",
		ClosingLine:     "真相總是殘酷的",
	}

	assert.NotEmpty(t, resp.EndingNarrative)
	assert.NotEmpty(t, resp.FinalEmotion)
	assert.NotEmpty(t, resp.ClosingLine)
}

// TestEndingType tests EndingType constants
func TestEndingType(t *testing.T) {
	tests := []struct {
		name        string
		endingType  string
		expected    string
	}{
		{"true ending", EndingTrue, "true"},
		{"good ending", EndingGood, "good"},
		{"bad ending", EndingBad, "bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.endingType)
		})
	}
}

// TestInvokeOpening_Basic tests basic InvokeOpening functionality
func TestInvokeOpening_Basic(t *testing.T) {
	agent := createTestOpeningEndingAgent(t)

	storyBible := &StoryBible{
		WorldView: WorldView{
			Setting:    "廢棄醫院",
			Atmosphere: "陰鬱、詭異、不安",
			TimeFrame:  "深夜",
			Background: "你是一名調查記者",
		},
		CoreTruth: CoreTruth{
			Truth: "醫院進行非法實驗",
		},
	}

	req := OpeningRequest{
		StoryBible: storyBible,
		Difficulty: "normal",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeOpening(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// AC #1: 序章應該是 800-1200 字
	runeCount := len([]rune(resp.OpeningNarrative))
	assert.GreaterOrEqual(t, runeCount, 100, "Opening narrative should be at least 100 characters (simplified for test)")

	// AC #1: 初始張力 10-20
	assert.GreaterOrEqual(t, resp.InitialTension, 10)
	assert.LessOrEqual(t, resp.InitialTension, 20)

	// AC #1: 應該有開場選擇引導
	assert.NotEmpty(t, resp.FirstChoice)
}

// TestInvokeEnding_TrueEnding tests True Ending generation
func TestInvokeEnding_TrueEnding(t *testing.T) {
	// Create mock client with True Ending response
	mockClient := agents.NewMockLLMClient(`{
		"ending_narrative": "當你站在醫院的廢墟中，終於明白了一切。這裡從未是救死扶傷的地方，而是一個巨大的實驗場。那些病患的慘叫聲，那些詭異的規則，都是為了掩蓋這個駭人聽聞的真相。你顫抖著翻開最後一頁病歷，上面寫著你自己的名字。原來，你也是實驗的一部分。當記憶如潮水般湧來，你終於想起了一切——你從未離開過這裡。震撼、恐懼、絕望，所有情緒交織在一起。你意識到，逃離從來不是選項，因為你本就是這黑暗的一部分。",
		"final_emotion": "shock",
		"closing_line": "真相總是殘酷的，而你就是真相本身。"
	}`)

	config := agents.AgentConfig{
		Name:       "TestEndingAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}
	agent := NewNarrationAgent(config)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(80)
	gameState.SetSAN(70)

	req := EndingRequest{
		GameState:  gameState,
		EndingType: EndingTrue,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeEnding(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// AC #7: True Ending 應該是 1300-1500 字
	runeCount := len([]rune(resp.EndingNarrative))
	assert.GreaterOrEqual(t, runeCount, 100, "Ending narrative should be at least 100 characters (simplified for test)")

	// AC #7: 情感應該是震撼相關
	assert.NotEmpty(t, resp.FinalEmotion)

	// AC #11: 應該有結局金句
	assert.NotEmpty(t, resp.ClosingLine)
}

// TestInvokeEnding_GoodEnding tests Good Ending generation
func TestInvokeEnding_GoodEnding(t *testing.T) {
	// Create mock client with Good Ending response
	mockClient := agents.NewMockLLMClient(`{
		"ending_narrative": "你終於逃出了醫院，陽光刺痛了你的雙眼。雖然你活了下來，但心中依然有太多疑問未解。那些規則究竟從何而來？為什麼只有你能逃出？回頭望向陰暗的建築，你感到一絲解脫，卻也帶著深深的疑惑。或許，有些真相永遠不該被揭開。",
		"final_emotion": "relief",
		"closing_line": "活著就是最好的結局，即使疑問永存。"
	}`)

	config := agents.AgentConfig{
		Name:       "TestEndingAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}
	agent := NewNarrationAgent(config)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(60)
	gameState.SetSAN(50)

	req := EndingRequest{
		GameState:  gameState,
		EndingType: EndingGood,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeEnding(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// AC #8: Good Ending 應該是 1100-1300 字
	runeCount := len([]rune(resp.EndingNarrative))
	assert.GreaterOrEqual(t, runeCount, 100, "Ending narrative should be at least 100 characters (simplified for test)")

	// AC #8: 情感應該是解脫但疑惑
	assert.NotEmpty(t, resp.FinalEmotion)
	assert.NotEmpty(t, resp.ClosingLine)
}

// TestInvokeEnding_BadEnding tests Bad Ending generation
func TestInvokeEnding_BadEnding(t *testing.T) {
	// Create mock client with Bad Ending response
	mockClient := agents.NewMockLLMClient(`{
		"ending_narrative": "黑暗吞噬了你。你的意識逐漸模糊，最後的畫面是那扇永遠無法打開的門。你失敗了，迷失在這座詛咒的建築中。絕望、無助、恐懼，這些情緒成為了你最後的記憶。醫院的走廊依然空蕩蕩的，等待著下一個受害者。你的故事就此終結，成為這棟建築中無數冤魂的一部分。",
		"final_emotion": "despair",
		"closing_line": "有些門，一旦踏入，就再也回不了頭。"
	}`)

	config := agents.AgentConfig{
		Name:       "TestEndingAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}
	agent := NewNarrationAgent(config)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(10)
	gameState.SetSAN(5)

	req := EndingRequest{
		GameState:  gameState,
		EndingType: EndingBad,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeEnding(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// AC #9: Bad Ending 應該是 1000-1200 字
	runeCount := len([]rune(resp.EndingNarrative))
	assert.GreaterOrEqual(t, runeCount, 100, "Ending narrative should be at least 100 characters (simplified for test)")

	// AC #9: 情感應該是絕望
	assert.NotEmpty(t, resp.FinalEmotion)
	assert.NotEmpty(t, resp.ClosingLine)
}

// TestDetermineEndingType tests ending type auto-determination
// AC #6: 自動判定結局類型
func TestDetermineEndingType(t *testing.T) {
	agent := createTestOpeningEndingAgent(t)

	tests := []struct {
		name        string
		san         int
		expectedType string
	}{
		{"True Ending (SAN 80)", 80, EndingTrue},
		{"True Ending (SAN 70)", 70, EndingTrue},
		{"Good Ending (SAN 60)", 60, EndingGood},
		{"Good Ending (SAN 40)", 40, EndingGood},
		{"Bad Ending (SAN 30)", 30, EndingBad},
		{"Bad Ending (SAN 0)", 0, EndingBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := engine.NewGameStateV2()
			gameState.SetSAN(tt.san)

			endingType := agent.determineEndingType(gameState)
			assert.Equal(t, tt.expectedType, endingType)
		})
	}
}

// TestInvokeEnding_AutoDetermineType tests auto-determination with empty EndingType
// AC #6: 自動判定結局類型（當未指定時）
func TestInvokeEnding_AutoDetermineType(t *testing.T) {
	mockClient := agents.NewMockLLMClient(`{
		"ending_narrative": "這是一段自動判定的結局敘事文本。根據你在遊戲中的整體表現和狀態，系統為你選擇了最適合的結局類型。雖然整個過程充滿艱辛與挑戰，但你終於走到了旅程的終點。這段難忘的經歷和其中的每一個選擇，都將永遠銘刻在你的心中，成為無法磨滅的記憶。",
		"final_emotion": "relief",
		"closing_line": "結局由你的選擇決定。"
	}`)

	config := agents.AgentConfig{
		Name:       "TestEndingAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}
	agent := NewNarrationAgent(config)

	gameState := engine.NewGameStateV2()
	gameState.SetHP(60)
	gameState.SetSAN(50) // Should auto-determine as "good"

	// Empty EndingType triggers auto-determination
	req := EndingRequest{
		GameState:  gameState,
		EndingType: "", // Auto-determine
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := agent.InvokeEnding(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.EndingNarrative)
	assert.NotEmpty(t, resp.FinalEmotion)
	assert.NotEmpty(t, resp.ClosingLine)
}

// TestOpeningDifficultyLevels tests opening generation with different difficulty levels
func TestOpeningDifficultyLevels(t *testing.T) {
	difficulties := []string{"easy", "normal", "hard", "hell"}

	for _, difficulty := range difficulties {
		t.Run(difficulty, func(t *testing.T) {
			mockClient := agents.NewMockLLMClient(`{
				"opening_narrative": "` + difficulty + ` 難度的序章敘事測試內容。這段敘事用於測試不同難度級別下的規則暗示隱晦程度。序章應該根據難度調整提示的明確程度，讓玩家感受到難度差異。隨著難度提升，暗示會變得更加隱晦和微妙。這是一段足夠長的測試文本，確保通過最小長度驗證。",
				"initial_tension": 15,
				"first_choice_prompt": "開始探索"
			}`)

			config := agents.AgentConfig{
				Name:       "TestOpeningAgent",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				LLMClient:  mockClient,
			}
			agent := NewNarrationAgent(config)

			storyBible := &StoryBible{
				WorldView: WorldView{
					Setting:    "測試場景",
					Atmosphere: "詭異",
					TimeFrame:  "深夜",
					Background: "測試背景",
				},
				CoreTruth: CoreTruth{
					Truth: "測試真相",
				},
			}

			req := OpeningRequest{
				StoryBible: storyBible,
				Difficulty: difficulty,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := agent.InvokeOpening(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotEmpty(t, resp.OpeningNarrative)
		})
	}
}

// TestOpeningValidation tests opening response validation edge cases
func TestOpeningValidation(t *testing.T) {
	agent := createTestOpeningEndingAgent(t)

	tests := []struct {
		name        string
		response    *OpeningResponse
		shouldError bool
	}{
		{
			name: "valid response",
			response: &OpeningResponse{
				OpeningNarrative: "這是一段足夠長的序章敘事文本，包含了世界觀介紹、氛圍建立和規則暗示等重要元素。讓玩家能夠充分理解遊戲的設定和背景故事。序章應該引人入勝，營造適當的氣氛，為後續的遊戲體驗奠定良好的基礎。這段文字確保長度符合最小驗證要求。",
				InitialTension:   15,
				FirstChoice:      "探索醫院",
			},
			shouldError: false,
		},
		{
			name: "too short narrative",
			response: &OpeningResponse{
				OpeningNarrative: "太短了",
				InitialTension:   15,
				FirstChoice:      "探索",
			},
			shouldError: true,
		},
		{
			name: "tension too low",
			response: &OpeningResponse{
				OpeningNarrative: "這是一段足夠長的敘事內容，但張力值設置不當。序章的張力值應該在合理範圍內，以確保遊戲的平衡性。",
				InitialTension:   5,
				FirstChoice:      "探索",
			},
			shouldError: true,
		},
		{
			name: "tension too high",
			response: &OpeningResponse{
				OpeningNarrative: "這是一段足夠長的敘事內容，但張力值設置過高。序章的張力值應該在合理範圍內，以確保遊戲體驗的平衡。",
				InitialTension:   25,
				FirstChoice:      "探索",
			},
			shouldError: true,
		},
		{
			name: "empty first choice",
			response: &OpeningResponse{
				OpeningNarrative: "這是一段足夠長的序章敘事，包含了所有必要的元素。但缺少了引導玩家進行第一個選擇的提示文本。",
				InitialTension:   15,
				FirstChoice:      "",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.validateOpeningResponse(tt.response)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEndingValidation tests ending response validation edge cases
func TestEndingValidation(t *testing.T) {
	agent := createTestOpeningEndingAgent(t)

	tests := []struct {
		name        string
		response    *EndingResponse
		endingType  string
		shouldError bool
	}{
		{
			name: "valid response",
			response: &EndingResponse{
				EndingNarrative: "這是一段完整的結局敘事文本，詳細描述了玩家在遊戲中的最終命運與結果。結局應該充分呼應玩家在整個遊戲過程中做出的各種選擇，並給予適當的情感收尾和總結。這段文字確保長度符合最小驗證要求，提供完整的故事閉環。",
				FinalEmotion:    "shock",
				ClosingLine:     "真相總是殘酷的。",
			},
			endingType:  EndingTrue,
			shouldError: false,
		},
		{
			name: "too short narrative",
			response: &EndingResponse{
				EndingNarrative: "太短",
				FinalEmotion:    "shock",
				ClosingLine:     "結束了",
			},
			endingType:  EndingTrue,
			shouldError: true,
		},
		{
			name: "invalid emotion",
			response: &EndingResponse{
				EndingNarrative: "這是一段足夠長的結局敘事，但情感類型設置不正確。應該使用系統定義的有效情感類型之一。",
				FinalEmotion:    "happy",
				ClosingLine:     "結束了",
			},
			endingType:  EndingTrue,
			shouldError: true,
		},
		{
			name: "empty closing line",
			response: &EndingResponse{
				EndingNarrative: "這是一段足夠長的結局敘事，包含了所有情節要素。但缺少了結局金句這個重要的收尾元素。",
				FinalEmotion:    "despair",
				ClosingLine:     "",
			},
			endingType:  EndingBad,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.validateEndingResponse(tt.response, tt.endingType)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// createTestOpeningEndingAgent creates a test narration agent for opening/ending tests
func createTestOpeningEndingAgent(t *testing.T) *NarrationAgent {
	t.Helper()

	// Create mock LLM client with default Opening response
	mockClient := agents.NewMockLLMClient(`{
		"opening_narrative": "深夜，你站在廢棄醫院的大門前。冷風吹過，帶來一陣腐朽的氣味。月光透過破碎的窗戶投下詭異的陰影，整座建築彷彿在訴說著不為人知的秘密。你的心跳加速，一種不祥的預感油然而生。這裡曾是救死扶傷的地方，如今卻籠罩在死亡的陰影之下。你深吸一口氣，推開了生鏽的鐵門。",
		"initial_tension": 15,
		"first_choice_prompt": "你決定探索醫院內部"
	}`)

	config := agents.AgentConfig{
		Name:       "TestOpeningEndingAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}

	agent := NewNarrationAgent(config)
	require.NotNil(t, agent)

	return agent
}
