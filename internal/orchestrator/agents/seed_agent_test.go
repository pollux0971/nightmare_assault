package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==========================================================================
// Story 6-5: Seed Agent Tests (Dual Mode)
// ==========================================================================
// These tests cover the REFACTORED SeedAgent with BaseAgentImpl pattern
// ==========================================================================

// TestNewSeedAgent tests SeedAgent creation with BaseAgentImpl pattern
func TestNewSeedAgent(t *testing.T) {
	mockLLM := &MockLLMClient{}

	config := AgentConfig{
		Name:       "SeedAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockLLM,
	}

	// H-2 FIX: NewSeedAgent now requires seedManager and tensionManager parameters
	agent := NewSeedAgent(config, nil, nil)

	require.NotNil(t, agent)
	assert.Equal(t, "SeedAgent", agent.GetName())
	assert.Equal(t, 10*time.Second, agent.GetTimeout())
}

// TestNewSeedAgent_WithDefaults tests SeedAgent with default config values
func TestNewSeedAgent_WithDefaults(t *testing.T) {
	config := AgentConfig{
		LLMClient: &MockLLMClient{},
	}

	// H-2 FIX: NewSeedAgent now requires seedManager and tensionManager parameters
	agent := NewSeedAgent(config, nil, nil)

	require.NotNil(t, agent)
	assert.Equal(t, "SeedAgent", agent.GetName())
	assert.Equal(t, 30*time.Second, agent.GetTimeout())
	assert.Equal(t, 3, agent.config.MaxRetries)
}

// ==========================================================================
// Type Tests
// ==========================================================================

// TestSeedOperation_String tests SeedOperation string representation
func TestSeedOperation_String(t *testing.T) {
	tests := []struct {
		op       SeedOperation
		expected string
	}{
		{SeedOpSkip, "Skip"},
		{SeedOpPlant, "Plant"},
		{SeedOpHarvest, "Harvest"},
		{SeedOpPrune, "Prune"},
		{SeedOperation(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.op.String())
		})
	}
}

// TestGlobalGenerateRequest_Types tests GlobalGenerateRequest structure
func TestGlobalGenerateRequest_Types(t *testing.T) {
	req := GlobalGenerateRequest{
		StoryBible:      nil,
		Difficulty:      "normal",
		StoryLength:     "medium",
		PossibleEndings: nil,
	}

	assert.Equal(t, "normal", req.Difficulty)
	assert.Equal(t, "medium", req.StoryLength)
}

// TestLocalManageRequest_Types tests LocalManageRequest structure
func TestLocalManageRequest_Types(t *testing.T) {
	req := LocalManageRequest{
		CurrentBeat:      5,
		PlayerHints:      2,
		CurrentContext:   "hospital_corridor",
		ActiveLocalSeeds: nil,
	}

	assert.Equal(t, 5, req.CurrentBeat)
	assert.Equal(t, 2, req.PlayerHints)
	assert.Equal(t, "hospital_corridor", req.CurrentContext)
}

// ==========================================================================
// InvokeGlobalGenerate Tests
// ==========================================================================

// TestInvokeGlobalGenerate_Success tests successful Global Seed generation
func TestInvokeGlobalGenerate_Success(t *testing.T) {
	// This test requires real LLM or comprehensive mock
	// Use integration test with real API instead
	t.Skip("Use integration tests with real API - see test_config.go")
}

// TestInvokeGlobalGenerate_SeedCount tests difficulty-based seed count
func TestInvokeGlobalGenerate_SeedCount(t *testing.T) {
	tests := []struct {
		difficulty    string
		expectedCount int
	}{
		{"easy", 3},
		{"normal", 4},
		{"hard", 5},
		{"hell", 5},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			// H-2 FIX: NewSeedAgent now requires seedManager and tensionManager parameters
			agent := NewSeedAgent(AgentConfig{LLMClient: &MockLLMClient{}}, nil, nil)

			count := agent.getSeedCountByDifficulty(tt.difficulty)
			assert.Equal(t, tt.expectedCount, count,
				"Seed count for difficulty %s should be %d", tt.difficulty, tt.expectedCount)
		})
	}
}

// TestInvokeGlobalGenerate_ThreeTierValidation tests 3-tier clue chain validation
func TestInvokeGlobalGenerate_ThreeTierValidation(t *testing.T) {
	// This test verifies the validation logic works correctly
	// Full integration test would require real LLM or comprehensive mocks
	t.Skip("Requires comprehensive LLM mock - tested in integration tests")
}

// TestInvokeGlobalGenerate_Timeout tests timeout handling
func TestInvokeGlobalGenerate_Timeout(t *testing.T) {
	t.Skip("Timeout test requires long-running LLM mock - tested manually")
}

// TestInvokeGlobalGenerate_LLMError tests LLM error handling and retry
func TestInvokeGlobalGenerate_LLMError(t *testing.T) {
	t.Skip("Error retry logic tested via BaseAgentImpl tests")
}
