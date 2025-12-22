// Package main demonstrates basic Trinity system usage
//
// This example shows:
// - Creating a Trinity Router
// - Sending messages to different agents
// - Automatic tier routing
// - Basic error handling
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

func main() {
	fmt.Println("=== Trinity Basic Usage Example ===\n")

	// Step 1: Check for API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Println("Please set it with: export ANTHROPIC_API_KEY=your-api-key")
		os.Exit(1)
	}

	// Step 2: Create Trinity Router configuration
	fmt.Println("Step 1: Creating Trinity Router configuration...")

	routerConfig := trinity.RouterConfig{
		// Thinking Tier - Complex reasoning (Opus 4.5)
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-opus-4-20250514",
			MaxTokens:   16000,
			Temperature: 0.4,
		},

		// Reactive Tier - Balanced performance (Sonnet 3.5)
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   8000,
			Temperature: 0.7,
		},

		// Rapid Tier - Fast responses (Haiku 3.5)
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-haiku-20240307",
			MaxTokens:   4000,
			Temperature: 0.9,
		},

		// Enable automatic fallback (Thinking → Reactive → Rapid)
		FallbackEnabled: true,
	}

	// Step 3: Create Trinity Router
	fmt.Println("Step 2: Creating Trinity Router...")

	router, err := trinity.NewTrinityRouter(routerConfig)
	if err != nil {
		fmt.Printf("Error creating router: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Trinity Router created successfully\n")

	// Step 4: Test different agents with different tiers
	ctx := context.Background()

	// Example 1: JudgeAgent (Thinking Tier)
	fmt.Println("Example 1: Using JudgeAgent (Thinking Tier)")
	fmt.Println("-------------------------------------------")

	judgeMessages := []client.Message{
		{
			Role:    "user",
			Content: "Analyze this player action: The player tries to open a locked door using a rusty key they found. What happens?",
		},
	}

	resp, err := router.Route(ctx, "JudgeAgent", judgeMessages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: JudgeAgent\n")
		fmt.Printf("Tier: Thinking (Opus 4.5)\n")
		fmt.Printf("Response: %s\n\n", resp.Content)
	}

	// Example 2: NarrationAgent (Reactive Tier)
	fmt.Println("Example 2: Using NarrationAgent (Reactive Tier)")
	fmt.Println("-----------------------------------------------")

	narrationMessages := []client.Message{
		{
			Role:    "user",
			Content: "Generate a narrative description for a player entering an abandoned hospital room.",
		},
	}

	resp, err = router.Route(ctx, "NarrationAgent", narrationMessages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: NarrationAgent\n")
		fmt.Printf("Tier: Reactive (Sonnet 3.5)\n")
		fmt.Printf("Response: %s\n\n", resp.Content)
	}

	// Example 3: DreamAgent (Rapid Tier)
	fmt.Println("Example 3: Using DreamAgent (Rapid Tier)")
	fmt.Println("-----------------------------------------")

	dreamMessages := []client.Message{
		{
			Role:    "user",
			Content: "Generate a brief, surreal dream scene for the player.",
		},
	}

	resp, err = router.Route(ctx, "DreamAgent", dreamMessages)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: DreamAgent\n")
		fmt.Printf("Tier: Rapid (Haiku 3.5)\n")
		fmt.Printf("Response: %s\n\n", resp.Content)
	}

	// Step 5: Get metrics summary
	fmt.Println("Step 3: Getting metrics summary...")
	fmt.Println("-----------------------------------")

	summary := router.GetMetricsSummary()

	fmt.Printf("Thinking Tier: %d requests\n", summary.ThinkingTier.TotalRequests)
	fmt.Printf("Reactive Tier: %d requests\n", summary.ReactiveTier.TotalRequests)
	fmt.Printf("Rapid Tier: %d requests\n", summary.RapidTier.TotalRequests)

	fmt.Println("\n=== Example Complete ===")
}
