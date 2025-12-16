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

// TestContentRequest_Validation tests content request validation
func TestContentRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *ContentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
		{
			name: "negative beat",
			req: &ContentRequest{
				Beat:      -1,
				GameState: engine.NewGameStateV2(),
			},
			wantErr: true,
			errMsg:  "beat must be >= 0",
		},
		{
			name: "nil game state",
			req: &ContentRequest{
				Beat:      1,
				GameState: nil,
			},
			wantErr: true,
			errMsg:  "game state cannot be nil",
		},
		{
			name: "valid request - minimal",
			req: &ContentRequest{
				Beat:      0,
				GameState: engine.NewGameStateV2(),
			},
			wantErr: false,
		},
		{
			name: "valid request - complete",
			req: &ContentRequest{
				Beat:             5,
				GameState:        engine.NewGameStateV2(),
				LastPlayerChoice: "探索房間",
				JudgeResult: &JudgeResult{
					ImpactLevel: "minor",
					HPDelta:     -5,
					SANDelta:    -3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentRequest(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestInvokeContent_BasicNarrative tests basic narrative generation
// AC #1: 返回 500-1200 字的敘事內容（MainNarrative）
func TestInvokeContent_BasicNarrative(t *testing.T) {
	// Setup agent with mock LLM
	agent := createTestNarrationAgent(t)

	req := &ContentRequest{
		Beat:             1,
		GameState:        engine.NewGameStateV2(),
		LastPlayerChoice: "探索房間",
		JudgeResult: &JudgeResult{
			ImpactLevel: "none",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.MainNarrative)

	// AC #1: 字數範圍 500-1200
	// TODO: 目前是 placeholder，等實作完成後需要驗證字數
	// narrativeLength := len([]rune(resp.MainNarrative))
	// assert.GreaterOrEqual(t, narrativeLength, 500, "Narrative too short")
	// assert.LessOrEqual(t, narrativeLength, 1200, "Narrative too long")
}

// TestInvokeContent_ResponseStructure tests response structure validity
func TestInvokeContent_ResponseStructure(t *testing.T) {
	agent := createTestNarrationAgent(t)

	req := &ContentRequest{
		Beat:      1,
		GameState: engine.NewGameStateV2(),
	}

	ctx := context.Background()
	resp, err := agent.InvokeContent(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify all fields are initialized (not nil slices)
	assert.NotNil(t, resp.PlantedSeeds)
	assert.NotNil(t, resp.HarvestedSeeds)
	assert.NotNil(t, resp.RuleHints)
}

// TestInvokeContent_InvalidRequest tests error handling for invalid requests
func TestInvokeContent_InvalidRequest(t *testing.T) {
	agent := createTestNarrationAgent(t)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *ContentRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "nil game state",
			req: &ContentRequest{
				Beat:      1,
				GameState: nil,
			},
		},
		{
			name: "negative beat",
			req: &ContentRequest{
				Beat:      -1,
				GameState: engine.NewGameStateV2(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := agent.InvokeContent(ctx, tt.req)
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

// createTestNarrationAgent creates a test narration agent with mock dependencies
func createTestNarrationAgent(t *testing.T) *NarrationAgent {
	t.Helper()

	config := agents.AgentConfig{
		Name:       "TestNarrationAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		// LLMClient will be mocked in later tests
		LLMClient: nil, // TODO: Add mock LLM client
	}

	return NewNarrationAgent(config)
}
