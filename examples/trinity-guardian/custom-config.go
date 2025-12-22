// Package main demonstrates custom configuration and dynamic tier adjustment
//
// This example shows:
// - Custom agent tier overrides
// - Dynamic tier adjustment based on performance
// - Multi-provider configuration
// - Custom retry strategies
// - Advanced metrics analysis
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

func main() {
	fmt.Println("=== Trinity Custom Configuration Example ===\n")

	// Step 1: Check for multiple providers
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if anthropicKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	// Step 2: Create custom configuration with multiple providers
	fmt.Println("Step 1: Creating custom multi-provider configuration...")

	routerConfig := trinity.RouterConfig{
		// Thinking Tier - Use Anthropic Opus (best reasoning)
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      anthropicKey,
			Model:       "claude-opus-4-20250514",
			MaxTokens:   16000,
			Temperature: 0.4,
		},

		// Reactive Tier - Use Anthropic Sonnet (balanced)
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      anthropicKey,
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   8000,
			Temperature: 0.7,
		},

		// Rapid Tier - Use OpenAI GPT-4o-mini if available, else Haiku
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  getProviderForRapid(openaiKey),
			APIKey:      getRapidAPIKey(anthropicKey, openaiKey),
			Model:       getRapidModel(openaiKey),
			MaxTokens:   4000,
			Temperature: 0.9,
		},

		// Custom agent tier overrides
		AgentTierOverrides: map[string]trinity.TierLevel{
			// Upgrade NarrationAgent to Thinking (improve story quality)
			"NarrationAgent": trinity.TierThinking,

			// Downgrade ChoiceAgent to Rapid (speed up choice generation)
			"ChoiceAgent": trinity.TierRapid,

			// Keep JudgeAgent at Thinking (critical decisions)
			"JudgeAgent": trinity.TierThinking,
		},

		FallbackEnabled: true,

		// Custom retry configuration
		RetryConfig: client.RetryConfig{
			MaxAttempts:    5,                        // More retries
			InitialBackoff: 500 * time.Millisecond,   // Shorter initial wait
			MaxBackoff:     10 * time.Second,         // Shorter max wait
			BackoffFactor:  1.5,                      // Gentler exponential growth
		},
	}

	fmt.Println("✓ Custom configuration created")
	fmt.Printf("  Thinking: %s (%s)\n", routerConfig.ThinkingProvider.Model, routerConfig.ThinkingProvider.ProviderID)
	fmt.Printf("  Reactive: %s (%s)\n", routerConfig.ReactiveProvider.Model, routerConfig.ReactiveProvider.ProviderID)
	fmt.Printf("  Rapid: %s (%s)\n", routerConfig.RapidProvider.Model, routerConfig.RapidProvider.ProviderID)
	fmt.Println()

	// Step 3: Create Trinity LLM Client
	fmt.Println("Step 2: Creating Trinity LLM Client...")

	clientConfig := trinity.TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           true,
		EnableMetrics:            true,
		DefaultTimeout:           45 * time.Second,
	}

	llmClient, err := trinity.NewTrinityLLMClient(routerConfig, clientConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Trinity LLM Client created\n")

	// Step 4: Display custom tier overrides
	fmt.Println("Step 3: Custom Agent Tier Overrides")
	fmt.Println("------------------------------------")

	overrides := []struct {
		agent       string
		defaultTier string
		customTier  string
		reason      string
	}{
		{"NarrationAgent", "Reactive", "Thinking", "Improve story quality"},
		{"ChoiceAgent", "Reactive", "Rapid", "Speed up choice generation"},
		{"JudgeAgent", "Thinking", "Thinking", "Keep critical decisions at highest tier"},
	}

	for _, override := range overrides {
		fmt.Printf("%s:\n", override.agent)
		fmt.Printf("  Default Tier: %s\n", override.defaultTier)
		fmt.Printf("  Custom Tier: %s\n", override.customTier)
		fmt.Printf("  Reason: %s\n\n", override.reason)
	}

	// Step 5: Test agents with custom configuration
	ctx := context.Background()

	fmt.Println("Step 4: Testing Agents with Custom Configuration")
	fmt.Println("=================================================\n")

	// Test 1: NarrationAgent (upgraded to Thinking)
	fmt.Println("Test 1: NarrationAgent (Upgraded to Thinking)")

	narrationMsg := []client.Message{
		{Role: "user", Content: "Generate a high-quality narrative description for a critical story moment."},
	}

	startTime := time.Now()
	resp, err := llmClient.SendMessage(ctx, "NarrationAgent", narrationMsg)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Response received in %v\n", duration)
		fmt.Printf("  Using Thinking tier: Opus 4.5\n")
		fmt.Printf("  Quality: High (as expected)\n\n")
	}

	// Test 2: ChoiceAgent (downgraded to Rapid)
	fmt.Println("Test 2: ChoiceAgent (Downgraded to Rapid)")

	choiceMsg := []client.Message{
		{Role: "user", Content: "Generate 3 quick choices for the player."},
	}

	startTime = time.Now()
	resp, err = llmClient.SendMessage(ctx, "ChoiceAgent", choiceMsg)
	duration = time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Response received in %v\n", duration)
		fmt.Printf("  Using Rapid tier: %s\n", routerConfig.RapidProvider.Model)
		fmt.Printf("  Speed: Fast (as expected)\n\n")
	}

	// Step 6: Demonstrate dynamic tier adjustment
	fmt.Println("Step 5: Dynamic Tier Adjustment Based on Performance")
	fmt.Println("====================================================\n")

	// Simulate multiple requests to gather metrics
	fmt.Println("Simulating 10 requests to gather performance data...")

	for i := 0; i < 10; i++ {
		msg := []client.Message{
			{Role: "user", Content: fmt.Sprintf("Test request #%d", i+1)},
		}

		_, err := llmClient.SendMessage(ctx, "JudgeAgent", msg)
		if err != nil {
			fmt.Printf("  Request %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("  ✓ Request %d succeeded\n", i+1)
		}
	}

	fmt.Println()

	// Analyze performance and adjust tiers
	summary := llmClient.GetMetrics()

	fmt.Println("Performance Analysis:")
	fmt.Printf("  Thinking Tier Error Rate: %.2f%%\n", summary.ThinkingTier.ErrorRate()*100)
	fmt.Printf("  Reactive Tier Error Rate: %.2f%%\n", summary.ReactiveTier.ErrorRate()*100)
	fmt.Printf("  Rapid Tier Error Rate: %.2f%%\n", summary.RapidTier.ErrorRate()*100)

	// Dynamic adjustment logic
	if summary.ThinkingTier.ErrorRate() > 0.1 {
		fmt.Println("\n⚠️  Thinking tier error rate high (>10%)")
		fmt.Println("  Downgrading some agents to Reactive tier...")

		llmClient.UpdateAgentTier("NarrationAgent", trinity.TierReactive)
		fmt.Println("  ✓ NarrationAgent downgraded to Reactive")
	}

	if summary.ThinkingTier.TotalRequests > 0 {
		avgLatency := summary.ThinkingTier.AverageLatency()
		if avgLatency > 15*time.Second {
			fmt.Printf("\n⚠️  Thinking tier avg latency high (%v)\n", avgLatency)
			fmt.Println("  Consider using faster models for non-critical agents")
		}
	}

	// Step 7: Test tier override at request level
	fmt.Println("\n\nStep 6: Request-Level Tier Override")
	fmt.Println("===================================\n")

	fmt.Println("Forcing DreamAgent to use Thinking tier (normally Rapid)...")

	dreamMsg := []client.Message{
		{Role: "user", Content: "Generate an exceptionally detailed dream scene."},
	}

	thinkingTier := trinity.TierThinking
	opts := trinity.SendOptions{
		TierOverride: &thinkingTier,
		Timeout:      30 * time.Second,
	}

	startTime = time.Now()
	resp, err = llmClient.SendMessageWithOptions(ctx, "DreamAgent", dreamMsg, opts)
	duration = time.Since(startTime)

	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Response received in %v\n", duration)
		fmt.Printf("  Forced tier: Thinking (Opus 4.5)\n")
		fmt.Printf("  Expected result: Higher quality but slower\n\n")
	}

	// Step 8: Final metrics report
	fmt.Println("Step 7: Final Performance Report")
	fmt.Println("=================================\n")

	finalSummary := llmClient.GetMetrics()

	fmt.Printf("Total Requests: %d\n",
		finalSummary.ThinkingTier.TotalRequests+
			finalSummary.ReactiveTier.TotalRequests+
			finalSummary.RapidTier.TotalRequests)

	fmt.Printf("\nTier Distribution:\n")
	fmt.Printf("  Thinking: %d (%.1f%%)\n",
		finalSummary.ThinkingTier.TotalRequests,
		calculatePercentage(finalSummary.ThinkingTier.TotalRequests, finalSummary))
	fmt.Printf("  Reactive: %d (%.1f%%)\n",
		finalSummary.ReactiveTier.TotalRequests,
		calculatePercentage(finalSummary.ReactiveTier.TotalRequests, finalSummary))
	fmt.Printf("  Rapid: %d (%.1f%%)\n",
		finalSummary.RapidTier.TotalRequests,
		calculatePercentage(finalSummary.RapidTier.TotalRequests, finalSummary))

	fmt.Println("\n=== Custom Configuration Example Complete ===")
}

// Helper functions

func getProviderForRapid(openaiKey string) string {
	if openaiKey != "" {
		return "openai"
	}
	return "anthropic"
}

func getRapidAPIKey(anthropicKey, openaiKey string) string {
	if openaiKey != "" {
		return openaiKey
	}
	return anthropicKey
}

func getRapidModel(openaiKey string) string {
	if openaiKey != "" {
		return "gpt-4o-mini"
	}
	return "claude-3-haiku-20240307"
}

func calculatePercentage(count int64, summary trinity.MetricsSummary) float64 {
	total := summary.ThinkingTier.TotalRequests +
		summary.ReactiveTier.TotalRequests +
		summary.RapidTier.TotalRequests

	if total == 0 {
		return 0
	}

	return float64(count) / float64(total) * 100
}
