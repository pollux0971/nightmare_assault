// Package main demonstrates full Trinity + Guardian integration
//
// This example shows:
// - Complete Trinity setup with all features
// - Guardian player protection integration
// - Metrics collection and reporting
// - Dynamic tier adjustment
// - Error handling and retry
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/guardian"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

func main() {
	fmt.Println("=== Trinity + Guardian Full Integration Example ===\n")

	// Step 1: Check environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	// Step 2: Create Trinity Router configuration
	fmt.Println("Step 1: Initializing Trinity Router...")

	routerConfig := trinity.RouterConfig{
		ThinkingProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-opus-4-20250514",
			MaxTokens:   16000,
			Temperature: 0.4,
		},
		ReactiveProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   8000,
			Temperature: 0.7,
		},
		RapidProvider: trinity.ProviderTierConfig{
			ProviderID:  "anthropic",
			APIKey:      apiKey,
			Model:       "claude-3-haiku-20240307",
			MaxTokens:   4000,
			Temperature: 0.9,
		},
		FallbackEnabled: true,
		RetryConfig: client.RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
	}

	// Step 3: Create Trinity LLM Client (high-level API)
	fmt.Println("Step 2: Creating Trinity LLM Client...")

	clientConfig := trinity.TrinityClientConfig{
		EnableThinkingExtraction: true,
		EnableFallback:           true,
		EnableMetrics:            true,
		DefaultTimeout:           60 * time.Second,
	}

	llmClient, err := trinity.NewTrinityLLMClient(routerConfig, clientConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Trinity LLM Client created\n")

	// Step 4: Create Guardian
	fmt.Println("Step 3: Initializing Experience Guardian...")

	guardianConfig := guardian.GuardianConfig{
		MaxConsecutiveDeaths:   2,
		LowStatStreakLimit:     3,
		LowHPThreshold:         20,
		LowSANThreshold:        30,
		EnableDifficultyTuning: true,
	}

	g := guardian.NewExperienceGuardian(guardianConfig)

	fmt.Println("✓ Experience Guardian initialized\n")

	// Step 5: Create mock game state
	fmt.Println("Step 4: Creating mock game state...")

	gameState := &engine.GameStateV2{
		PlayerHP:  15, // Low HP (< 20)
		PlayerSAN: 25, // Low SAN (< 30)
		MaxHP:     100,
		MaxSAN:    100,
	}

	fmt.Printf("✓ Game state created (HP: %d, SAN: %d)\n\n", gameState.PlayerHP, gameState.PlayerSAN)

	// Step 6: Simulate game loop with Guardian integration
	fmt.Println("Step 5: Running game loop simulation...")
	fmt.Println("=======================================\n")

	ctx := context.Background()

	// Turn 1: Low HP/SAN but no consecutive deaths yet
	fmt.Println("--- Turn 1 ---")
	g.OnTurnEnd(gameState)

	if g.ShouldProtect() {
		reason := g.GetProtectionReason()
		fmt.Printf("⚠️  Guardian Protection Activated: %s\n", reason)

		// Upgrade NarrationAgent to Thinking tier for better experience
		llmClient.UpdateAgentTier("NarrationAgent", trinity.TierThinking)
		fmt.Println("✓ Upgraded NarrationAgent to Thinking tier")
	}

	// Generate narration with potentially upgraded tier
	narrationMsg := []client.Message{
		{Role: "user", Content: "Generate a supportive narrative description for a struggling player."},
	}

	resp, err := llmClient.SendMessage(ctx, "NarrationAgent", narrationMsg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Narration: %s\n\n", resp.Content[:100]+"...")
	}

	// Turn 2: Simulate player death
	fmt.Println("--- Turn 2 ---")
	gameState.PlayerHP = 0 // Player died
	g.OnTurnEnd(gameState)

	if g.ShouldProtect() {
		fmt.Printf("⚠️  Guardian Protection: %s\n", g.GetProtectionReason())
	}

	// Turn 3: Consecutive death
	fmt.Println("\n--- Turn 3 ---")
	g.OnTurnEnd(gameState) // Still dead

	if g.ShouldProtect() {
		fmt.Printf("⚠️  Guardian Protection: %s\n", g.GetProtectionReason())
		fmt.Println("✓ System will now use enhanced models for better player experience")
	}

	// Step 7: Test different agents
	fmt.Println("\n--- Agent Tests ---")

	agents := []struct {
		name    string
		prompt  string
		tier    string
	}{
		{"JudgeAgent", "Make a judgment about player action", "Thinking"},
		{"ChoiceAgent", "Generate 3 choices for the player", "Reactive"},
		{"DreamAgent", "Generate a dream scene", "Rapid"},
	}

	for _, agent := range agents {
		fmt.Printf("\nTesting %s (Default Tier: %s)...\n", agent.name, agent.tier)

		messages := []client.Message{
			{Role: "user", Content: agent.prompt},
		}

		startTime := time.Now()
		resp, err := llmClient.SendMessage(ctx, agent.name, messages)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
		} else {
			fmt.Printf("  ✓ Response received in %v\n", duration)
			fmt.Printf("  Content preview: %s...\n", resp.Content[:50])

			// Check if thinking tags were extracted
			if llmClient.HasThinking(resp) {
				thinkingChain, _ := llmClient.ExtractThinking(resp)
				fmt.Printf("  💭 Thinking chain length: %d chars\n", len(thinkingChain))
			}
		}
	}

	// Step 8: Get comprehensive metrics
	fmt.Println("\n\n=== Performance Metrics ===")
	fmt.Println("===========================\n")

	summary := llmClient.GetMetrics()

	fmt.Printf("Thinking Tier:\n")
	fmt.Printf("  Total Requests: %d\n", summary.ThinkingTier.TotalRequests)
	fmt.Printf("  Success Rate: %.2f%%\n", summary.ThinkingTier.SuccessRate()*100)
	if summary.ThinkingTier.TotalRequests > 0 {
		fmt.Printf("  Average Latency: %v\n", summary.ThinkingTier.AverageLatency())
		fmt.Printf("  Min Latency: %v\n", summary.ThinkingTier.MinDuration)
		fmt.Printf("  Max Latency: %v\n", summary.ThinkingTier.MaxDuration)
	}

	fmt.Printf("\nReactive Tier:\n")
	fmt.Printf("  Total Requests: %d\n", summary.ReactiveTier.TotalRequests)
	fmt.Printf("  Success Rate: %.2f%%\n", summary.ReactiveTier.SuccessRate()*100)
	if summary.ReactiveTier.TotalRequests > 0 {
		fmt.Printf("  Average Latency: %v\n", summary.ReactiveTier.AverageLatency())
	}

	fmt.Printf("\nRapid Tier:\n")
	fmt.Printf("  Total Requests: %d\n", summary.RapidTier.TotalRequests)
	fmt.Printf("  Success Rate: %.2f%%\n", summary.RapidTier.SuccessRate()*100)
	if summary.RapidTier.TotalRequests > 0 {
		fmt.Printf("  Average Latency: %v\n", summary.RapidTier.AverageLatency())
	}

	// Get fallback metrics
	if clientConfig.EnableFallback {
		fallbackMetrics := llmClient.GetFallbackMetrics()
		if fallbackMetrics != nil {
			fmt.Printf("\nFallback Statistics:\n")
			fmt.Printf("  Total Fallbacks: %d\n", fallbackMetrics.TotalFallbacks)
			fmt.Printf("  Full Degradations: %d\n", fallbackMetrics.FullDegradationCount)
			fmt.Printf("  Total Retries: %d\n", fallbackMetrics.TotalRetries)
		}
	}

	// Step 9: Output complete performance report
	fmt.Println("\n\n=== Complete Performance Report ===")
	llmClient.LogPerformanceReport()

	fmt.Println("\n=== Integration Example Complete ===")
}
