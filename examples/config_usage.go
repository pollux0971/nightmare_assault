// Package main demonstrates how to use the API configuration system
package main

import (
	"fmt"
	"log"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func main() {
	fmt.Println("=== Nightmare Assault API Configuration Demo ===\n")

	// 1. Load configuration (automatically loads from env vars and config file)
	fmt.Println("1. Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	fmt.Println("✓ Configuration loaded successfully")

	// 2. Check if API is configured
	fmt.Println("\n2. Checking API configuration...")
	if !cfg.IsConfigured() {
		fmt.Println("⚠ API not configured yet")
		fmt.Println("Please set up your API using one of these methods:")
		fmt.Println("  - Set environment variables (see .env.example)")
		fmt.Println("  - Edit ~/.nightmare/config.json (see config.example.json)")
		return
	}
	fmt.Printf("✓ API configured: %s (%s)\n", cfg.API.Provider.ProviderID, cfg.API.Provider.Model)

	// 3. Get API key (environment variable takes priority)
	fmt.Println("\n3. Retrieving API key...")
	apiKey := cfg.GetAPIKey(cfg.API.Provider.ProviderID)
	if apiKey == "" {
		fmt.Printf("⚠ No API key found for provider: %s\n", cfg.API.Provider.ProviderID)
		return
	}
	fmt.Printf("✓ API key found: %s...%s\n", apiKey[:10], apiKey[len(apiKey)-4:])

	// 4. Display current configuration
	fmt.Println("\n4. Current API configuration:")
	fmt.Printf("   Provider ID:  %s\n", cfg.API.Provider.ProviderID)
	fmt.Printf("   Model:        %s\n", cfg.API.Provider.Model)
	fmt.Printf("   Max Tokens:   %d\n", cfg.API.Provider.MaxTokens)
	if cfg.API.Provider.BaseURL != "" {
		fmt.Printf("   Base URL:     %s\n", cfg.API.Provider.BaseURL)
	}

	// 5. List all configured API keys
	fmt.Println("\n5. Configured providers:")
	for providerID := range cfg.API.APIKeys {
		key := cfg.GetAPIKey(providerID)
		if key != "" {
			fmt.Printf("   ✓ %s: %s...%s\n", providerID, key[:10], key[len(key)-4:])
		}
	}

	// 6. Debug settings
	fmt.Println("\n6. Debug settings:")
	fmt.Printf("   Debug mode:      %v\n", cfg.Debug.Enabled)
	fmt.Printf("   Log API keys:    %v\n", cfg.Debug.LogAPIKeys)
	fmt.Printf("   Log requests:    %v\n", cfg.Debug.LogRequests)

	// 7. Example: Setting a new API key programmatically
	fmt.Println("\n7. Example: Setting API key programmatically")
	fmt.Println("   (Uncomment the following lines to test)")
	// err = cfg.SetAPIKey("openrouter", "sk-or-v1-your-new-key")
	// if err != nil {
	//     log.Fatal("Failed to save API key:", err)
	// }
	// fmt.Println("   ✓ API key saved to config file")

	fmt.Println("\n=== Configuration Demo Complete ===")
}
