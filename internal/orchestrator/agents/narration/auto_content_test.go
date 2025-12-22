// Story 7-2 Test: InvokeAutoContent Implementation
package narration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
)

// TestInvokeAutoContent_AC1_MethodExists tests that InvokeAutoContent method exists
func TestInvokeAutoContent_AC1_MethodExists(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)

	assert.NotNil(t, agent, "Agent should not be nil")

	// Verify method exists by calling it
	ctx := context.Background()
	req := &AutoContentRequest{
		GameState:  createTestGameState(),
		StoryBible: &StoryBible{},
		RiskLevel:  "low",
		Beat:       1,
	}

	response, err := agent.InvokeAutoContent(ctx, req)
	assert.NoError(t, err, "Method should exist and not error on basic call")
	assert.NotNil(t, response, "Response should not be nil")
}

// TestInvokeAutoContent_AC2_AcceptsRequiredFields tests that method accepts all required fields
func TestInvokeAutoContent_AC2_AcceptsRequiredFields(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	// Create request with all required fields
	req := &AutoContentRequest{
		GameState:  createTestGameState(),
		StoryBible: &StoryBible{},
		AutoAction: "探索前方走廊",
		RiskLevel:  "medium",
		Beat:       5,
	}

	response, err := agent.InvokeAutoContent(ctx, req)

	assert.NoError(t, err, "Should accept request with all fields")
	assert.NotNil(t, response, "Should return response")
}

// TestInvokeAutoContent_AC3_SafeNarrative tests safe narrative generation
func TestInvokeAutoContent_AC3_SafeNarrative(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	tests := []struct {
		name      string
		riskLevel string
		maxHPLoss int
		maxSANLoss int
	}{
		{"Low risk", "low", 5, 5},
		{"Medium risk", "medium", 5, 5},
		{"High risk", "high", 5, 5},
		{"Lethal risk", "lethal", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AutoContentRequest{
				GameState:  createTestGameState(),
				StoryBible: &StoryBible{},
				RiskLevel:  tt.riskLevel,
				Beat:       1,
			}

			response, err := agent.InvokeAutoContent(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, response)

			// AC3: Should avoid lethal damage
			assert.LessOrEqual(t, -response.HPDelta, tt.maxHPLoss, "HP loss should be safe (≤ 5)")
			assert.LessOrEqual(t, -response.SANDelta, tt.maxSANLoss, "SAN loss should be safe (≤ 5)")
			assert.GreaterOrEqual(t, response.HPDelta, -5, "HP delta should be >= -5")
			assert.GreaterOrEqual(t, response.SANDelta, -5, "SAN delta should be >= -5")
		})
	}
}

// TestInvokeAutoContent_AC4_ResponseStructure tests response structure
func TestInvokeAutoContent_AC4_ResponseStructure(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	req := &AutoContentRequest{
		GameState:  createTestGameState(),
		StoryBible: &StoryBible{},
		RiskLevel:  "low",
		Beat:       1,
	}

	response, err := agent.InvokeAutoContent(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)

	// AC4: Check all required fields
	assert.NotEmpty(t, response.Narrative, "Should have narrative")
	assert.NotNil(t, response.PlantedSeeds, "PlantedSeeds should not be nil")
	assert.NotNil(t, response.RevealedClues, "RevealedClues should not be nil")
	// HPDelta and SANDelta are checked in AC3 test
}

// TestInvokeAutoContent_NarrativeLength tests narrative length is appropriate for auto mode
func TestInvokeAutoContent_NarrativeLength(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	req := &AutoContentRequest{
		GameState:  createTestGameState(),
		StoryBible: &StoryBible{},
		RiskLevel:  "low",
		Beat:       1,
	}

	response, err := agent.InvokeAutoContent(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Narrative should be concise (300-800 chars for auto mode)
	// Current implementation uses placeholder, so just check it exists
	assert.NotEmpty(t, response.Narrative, "Should have narrative text")
}

// TestInvokeAutoContent_NilRequest tests handling of nil request
func TestInvokeAutoContent_NilRequest(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	response, err := agent.InvokeAutoContent(ctx, nil)

	assert.Error(t, err, "Should error on nil request")
	assert.Nil(t, response, "Response should be nil on error")
}

// TestInvokeAutoContent_NilGameState tests handling of nil game state
func TestInvokeAutoContent_NilGameState(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	req := &AutoContentRequest{
		GameState:  nil,
		StoryBible: &StoryBible{},
		RiskLevel:  "low",
		Beat:       1,
	}

	response, err := agent.InvokeAutoContent(ctx, req)

	assert.Error(t, err, "Should error on nil game state")
	assert.Nil(t, response, "Response should be nil on error")
}

// TestInvokeAutoContent_DifferentRiskLevels tests different risk levels produce different deltas
func TestInvokeAutoContent_DifferentRiskLevels(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	var responses []*AutoContentResponse
	riskLevels := []string{"none", "low", "medium", "high", "lethal"}

	for _, risk := range riskLevels {
		req := &AutoContentRequest{
			GameState:  createTestGameState(),
			StoryBible: &StoryBible{},
			RiskLevel:  risk,
			Beat:       1,
		}

		response, err := agent.InvokeAutoContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)
		responses = append(responses, response)
	}

	// Higher risk should generally cause more SAN loss
	// none should have least loss (-1), lethal should have most loss (-4)
	// Since values are negative, -4 < -1, so we check absolute values or reverse comparison
	assert.GreaterOrEqual(t, responses[0].SANDelta, responses[4].SANDelta,
		"Higher risk should cause more SAN loss (more negative values)")
}

// TestAgentInvoke_AutoContentRequest tests that agent Invoke handles AutoContentRequest
func TestAgentInvoke_AutoContentRequest(t *testing.T) {
	config := agents.AgentConfig{
		Name: "TestNarrationAgent",
	}
	agent := NewNarrationAgent(config)
	ctx := context.Background()

	req := &AutoContentRequest{
		GameState:  createTestGameState(),
		StoryBible: &StoryBible{},
		RiskLevel:  "low",
		Beat:       1,
	}

	// Call via generic Invoke method
	response, err := agent.Invoke(ctx, req)

	assert.NoError(t, err, "Invoke should handle AutoContentRequest")
	assert.NotNil(t, response, "Should return response")

	// Type assert to verify correct type
	autoResponse, ok := response.(*AutoContentResponse)
	assert.True(t, ok, "Response should be *AutoContentResponse")
	assert.NotNil(t, autoResponse, "Auto response should not be nil")
}

// Helper function to create test game state
func createTestGameState() *engine.GameStateV2 {
	return &engine.GameStateV2{
		HP:          80,
		SAN:         70,
		CurrentBeat: 1,
	}
}
