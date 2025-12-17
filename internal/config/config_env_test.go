package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadFromEnv tests loading configuration from environment variables
func TestLoadFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"NIGHTMARE_API_PROVIDER":   os.Getenv("NIGHTMARE_API_PROVIDER"),
		"NIGHTMARE_API_MODEL":      os.Getenv("NIGHTMARE_API_MODEL"),
		"NIGHTMARE_API_MAX_TOKENS": os.Getenv("NIGHTMARE_API_MAX_TOKENS"),
		"OPENROUTER_API_KEY":       os.Getenv("OPENROUTER_API_KEY"),
		"ANTHROPIC_API_KEY":        os.Getenv("ANTHROPIC_API_KEY"),
		"NIGHTMARE_DEBUG":          os.Getenv("NIGHTMARE_DEBUG"),
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("NIGHTMARE_API_PROVIDER", "openrouter")
	os.Setenv("NIGHTMARE_API_MODEL", "anthropic/claude-3.5-sonnet")
	os.Setenv("NIGHTMARE_API_MAX_TOKENS", "150000")
	os.Setenv("OPENROUTER_API_KEY", "sk-or-v1-test-key")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
	os.Setenv("NIGHTMARE_DEBUG", "true")

	// Create config and load from env
	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	// Verify provider settings
	assert.Equal(t, "openrouter", cfg.API.Provider.ProviderID)
	assert.Equal(t, "anthropic/claude-3.5-sonnet", cfg.API.Provider.Model)
	assert.Equal(t, 150000, cfg.API.Provider.MaxTokens)

	// Verify API keys
	assert.Equal(t, "sk-or-v1-test-key", cfg.API.APIKeys["openrouter"])
	assert.Equal(t, "sk-ant-test-key", cfg.API.APIKeys["anthropic"])

	// Verify debug settings
	assert.True(t, cfg.Debug.Enabled)
}

// TestGetAPIKey tests API key retrieval with environment variable priority
func TestGetAPIKey(t *testing.T) {
	// Save original environment
	originalKey := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		if originalKey == "" {
			os.Unsetenv("OPENROUTER_API_KEY")
		} else {
			os.Setenv("OPENROUTER_API_KEY", originalKey)
		}
	}()

	// Test 1: Config file only
	cfg := DefaultConfig()
	cfg.API.APIKeys["openrouter"] = "config-key"
	assert.Equal(t, "config-key", cfg.GetAPIKey("openrouter"))

	// Test 2: Environment variable overrides config file
	os.Setenv("OPENROUTER_API_KEY", "env-key")
	assert.Equal(t, "env-key", cfg.GetAPIKey("openrouter"))

	// Test 3: Non-existent key
	os.Unsetenv("OPENROUTER_API_KEY")
	delete(cfg.API.APIKeys, "openrouter")
	assert.Equal(t, "", cfg.GetAPIKey("openrouter"))
}

// TestSetAPIKey tests setting and saving API keys
func TestSetAPIKey(t *testing.T) {
	cfg := DefaultConfig()

	// Set API key
	cfg.API.APIKeys["test-provider"] = "test-key"

	// Verify it was set
	assert.Equal(t, "test-key", cfg.API.APIKeys["test-provider"])
}

// TestLoadFromEnv_InvalidMaxTokens tests handling of invalid max tokens
func TestLoadFromEnv_InvalidMaxTokens(t *testing.T) {
	// Save original environment
	originalMaxTokens := os.Getenv("NIGHTMARE_API_MAX_TOKENS")
	defer func() {
		if originalMaxTokens == "" {
			os.Unsetenv("NIGHTMARE_API_MAX_TOKENS")
		} else {
			os.Setenv("NIGHTMARE_API_MAX_TOKENS", originalMaxTokens)
		}
	}()

	cfg := DefaultConfig()
	originalMaxTokensValue := cfg.API.Provider.MaxTokens

	// Test with invalid string
	os.Setenv("NIGHTMARE_API_MAX_TOKENS", "invalid")
	cfg.LoadFromEnv()
	assert.Equal(t, originalMaxTokensValue, cfg.API.Provider.MaxTokens, "Should not change on invalid input")

	// Test with negative number
	os.Setenv("NIGHTMARE_API_MAX_TOKENS", "-1000")
	cfg.LoadFromEnv()
	assert.Equal(t, originalMaxTokensValue, cfg.API.Provider.MaxTokens, "Should not change on negative input")

	// Test with valid number
	os.Setenv("NIGHTMARE_API_MAX_TOKENS", "200000")
	cfg.LoadFromEnv()
	assert.Equal(t, 200000, cfg.API.Provider.MaxTokens, "Should update with valid number")
}

// TestLoadFromEnv_DebugSettings tests loading debug settings from environment
func TestLoadFromEnv_DebugSettings(t *testing.T) {
	tests := []struct {
		name           string
		debugValue     string
		logKeysValue   string
		logReqValue    string
		expectDebug    bool
		expectLogKeys  bool
		expectLogReq   bool
	}{
		{
			name:          "all true",
			debugValue:    "true",
			logKeysValue:  "true",
			logReqValue:   "true",
			expectDebug:   true,
			expectLogKeys: true,
			expectLogReq:  true,
		},
		{
			name:          "all false",
			debugValue:    "false",
			logKeysValue:  "false",
			logReqValue:   "false",
			expectDebug:   false,
			expectLogKeys: false,
			expectLogReq:  false,
		},
		{
			name:          "mixed case TRUE",
			debugValue:    "TRUE",
			logKeysValue:  "True",
			logReqValue:   "tRuE",
			expectDebug:   true,
			expectLogKeys: true,
			expectLogReq:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("NIGHTMARE_DEBUG", tt.debugValue)
			os.Setenv("NIGHTMARE_LOG_API_KEYS", tt.logKeysValue)
			os.Setenv("NIGHTMARE_LOG_REQUESTS", tt.logReqValue)

			// Clean up after test
			defer func() {
				os.Unsetenv("NIGHTMARE_DEBUG")
				os.Unsetenv("NIGHTMARE_LOG_API_KEYS")
				os.Unsetenv("NIGHTMARE_LOG_REQUESTS")
			}()

			cfg := DefaultConfig()
			cfg.LoadFromEnv()

			assert.Equal(t, tt.expectDebug, cfg.Debug.Enabled)
			assert.Equal(t, tt.expectLogKeys, cfg.Debug.LogAPIKeys)
			assert.Equal(t, tt.expectLogReq, cfg.Debug.LogRequests)
		})
	}
}

// TestLoadFromEnv_AllProviders tests loading API keys for all supported providers
func TestLoadFromEnv_AllProviders(t *testing.T) {
	testKeys := map[string]string{
		"OPENROUTER_API_KEY": "sk-or-test",
		"ANTHROPIC_API_KEY":  "sk-ant-test",
		"OPENAI_API_KEY":     "sk-test",
		"GOOGLE_API_KEY":     "google-test",
		"COHERE_API_KEY":     "cohere-test",
	}

	// Set all test keys
	for envVar, value := range testKeys {
		os.Setenv(envVar, value)
	}

	// Clean up after test
	defer func() {
		for envVar := range testKeys {
			os.Unsetenv(envVar)
		}
	}()

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	// Verify all keys were loaded
	assert.Equal(t, "sk-or-test", cfg.API.APIKeys["openrouter"])
	assert.Equal(t, "sk-ant-test", cfg.API.APIKeys["anthropic"])
	assert.Equal(t, "sk-test", cfg.API.APIKeys["openai"])
	assert.Equal(t, "google-test", cfg.API.APIKeys["google"])
	assert.Equal(t, "cohere-test", cfg.API.APIKeys["cohere"])
}

// TestGetAPIKey_AllProviders tests GetAPIKey for all supported providers
func TestGetAPIKey_AllProviders(t *testing.T) {
	providers := []struct {
		providerID string
		envVar     string
		testKey    string
	}{
		{"openrouter", "OPENROUTER_API_KEY", "sk-or-test"},
		{"anthropic", "ANTHROPIC_API_KEY", "sk-ant-test"},
		{"openai", "OPENAI_API_KEY", "sk-test"},
		{"google", "GOOGLE_API_KEY", "google-test"},
		{"cohere", "COHERE_API_KEY", "cohere-test"},
	}

	for _, p := range providers {
		t.Run(p.providerID, func(t *testing.T) {
			// Set environment variable
			os.Setenv(p.envVar, p.testKey)
			defer os.Unsetenv(p.envVar)

			cfg := DefaultConfig()
			key := cfg.GetAPIKey(p.providerID)

			require.Equal(t, p.testKey, key)
		})
	}
}

// TestLoadFromEnv_BaseURL tests loading custom base URL
func TestLoadFromEnv_BaseURL(t *testing.T) {
	originalBaseURL := os.Getenv("NIGHTMARE_API_BASE_URL")
	defer func() {
		if originalBaseURL == "" {
			os.Unsetenv("NIGHTMARE_API_BASE_URL")
		} else {
			os.Setenv("NIGHTMARE_API_BASE_URL", originalBaseURL)
		}
	}()

	testURL := "https://custom-api.example.com/v1"
	os.Setenv("NIGHTMARE_API_BASE_URL", testURL)

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	assert.Equal(t, testURL, cfg.API.Provider.BaseURL)
}
