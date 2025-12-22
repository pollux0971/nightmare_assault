package orchestrator

import (
	"context"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// TrinityLLMClient adapts TrinityRouter to implement the agents.LLMClient interface
//
// Story 9-10: This adapter bridges the gap between TrinityRouter and agents.
// It automatically routes agent requests to the appropriate tier based on agent name.
//
// Design:
//   - Each agent call includes the agent name for tier routing
//   - TrinityRouter handles automatic fallback between tiers
//   - Thinking tags are automatically extracted and removed
//   - Metrics are collected for all requests
type TrinityLLMClient struct {
	router    *trinity.TrinityRouter
	agentName string // The name of the agent using this client
}

// NewTrinityLLMClient creates a new TrinityLLMClient for a specific agent
//
// Parameters:
//   - router: The TrinityRouter instance to use for routing
//   - agentName: The name of the agent (used for tier mapping)
//
// Returns:
//   - *TrinityLLMClient: A new client instance
func NewTrinityLLMClient(router *trinity.TrinityRouter, agentName string) *TrinityLLMClient {
	return &TrinityLLMClient{
		router:    router,
		agentName: agentName,
	}
}

// Generate implements the agents.LLMClient interface
//
// This method:
//  1. Converts the prompt string to a client.Message slice
//  2. Routes the request through TrinityRouter (with tier selection and fallback)
//  3. Returns the response content
//
// The router automatically:
//  - Selects the appropriate tier based on agent name
//  - Handles fallback to lower tiers on failure
//  - Extracts and removes thinking tags for Thinking-tier responses
//  - Collects performance metrics
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - prompt: The prompt string to send
//   - options: Additional options (currently unused, for future compatibility)
//
// Returns:
//   - string: The LLM response content (with thinking tags removed if applicable)
//   - error: Any error that occurred
func (c *TrinityLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	// Convert prompt to messages format expected by TrinityRouter
	messages := []client.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Route through Trinity with automatic tier selection and fallback
	resp, err := c.router.Route(ctx, c.agentName, messages)
	if err != nil {
		return "", err
	}

	// Return the cleaned content (thinking tags already removed by router)
	return resp.Content, nil
}

// GetThinkingChain extracts the thinking chain from the last response (if available)
//
// This is a Trinity-specific extension that allows agents to access the thinking
// process used by Thinking-tier models. This can be useful for debugging or
// logging purposes.
//
// Parameters:
//   - resp: The response from TrinityRouter
//
// Returns:
//   - string: The thinking chain if present, empty string otherwise
//   - bool: Whether a thinking chain was found
func GetThinkingChain(resp *client.Response) (string, bool) {
	if resp == nil || resp.Metadata == nil {
		return "", false
	}

	chain, ok := resp.Metadata["thinking_chain"].(string)
	return chain, ok
}
