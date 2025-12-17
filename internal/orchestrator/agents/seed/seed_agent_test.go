package seed

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/orchestrator/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSeedAgent tests SeedAgent creation with BaseAgentImpl pattern
func TestNewSeedAgent(t *testing.T) {
	mockLLM := &agents.MockLLMClient{}

	config := agents.AgentConfig{
		Name:       "SeedAgent",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockLLM,
	}

	agent := NewSeedAgent(config, nil, nil) // nil for pruner and tensionMgr for now

	require.NotNil(t, agent)
	assert.Equal(t, "SeedAgent", agent.GetName())
	assert.Equal(t, 10*time.Second, agent.GetTimeout())
}

// TestNewSeedAgent_WithDefaults tests SeedAgent with default config values
func TestNewSeedAgent_WithDefaults(t *testing.T) {
	config := agents.AgentConfig{
		LLMClient: &agents.MockLLMClient{},
	}

	agent := NewSeedAgent(config, nil, nil)

	require.NotNil(t, agent)
	assert.Equal(t, "SeedAgent", agent.GetName())
	assert.Equal(t, 30*time.Second, agent.GetTimeout())
	assert.Equal(t, 3, agent.config.MaxRetries)
}
