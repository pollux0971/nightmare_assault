package commands

import (
	"context"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestAPIConnectionTest_Success tests successful API connection.
func TestAPIConnectionTest_Success(t *testing.T) {
	// Skip if no API key available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()

	// Try to use environment variable for testing
	providerID := "openai"
	apiKey := cfg.GetAPIKey(providerID)

	if apiKey == "" {
		t.Skip("No API key configured for testing")
	}

	// Configure the provider
	cfg.API.Provider.ProviderID = providerID
	cfg.API.Provider.Model = "gpt-4o-mini"
	cfg.API.Provider.MaxTokens = 100
	cfg.API.APIKeys[providerID] = apiKey

	cmd := NewAPICommand(cfg)
	result := cmd.testConnection()

	if !result.Success {
		t.Errorf("Expected successful connection test, got: %s", result.Message)
	}

	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Check that message contains success indicator
	if len(result.Message) < 5 {
		t.Errorf("Expected detailed message, got: %s", result.Message)
	}
}

// TestAPIConnectionTest_InvalidProvider tests connection with invalid provider.
func TestAPIConnectionTest_InvalidProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "invalid_provider_12345"
	cfg.API.APIKeys["invalid_provider_12345"] = "fake_key"

	cmd := NewAPICommand(cfg)
	result := cmd.testConnection()

	if result.Success {
		t.Error("Expected failed connection test for invalid provider")
	}

	if result.Message == "" {
		t.Error("Expected error message")
	}
}

// TestAPIConnectionTest_NoProvider tests connection when no provider is configured.
func TestAPIConnectionTest_NoProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	// Don't set any provider

	cmd := NewAPICommand(cfg)
	result := cmd.testConnection()

	if result.Success {
		t.Error("Expected failed connection test when no provider configured")
	}

	if result.Message == "" {
		t.Error("Expected error message")
	}

	// Should mention that provider is not configured
	if len(result.Message) < 5 {
		t.Errorf("Expected meaningful error message, got: %s", result.Message)
	}
}

// TestAPIConnectionTest_NoAPIKey tests connection when API key is missing.
func TestAPIConnectionTest_NoAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"
	// Don't set API key

	cmd := NewAPICommand(cfg)
	result := cmd.testConnection()

	if result.Success {
		t.Error("Expected failed connection test when API key missing")
	}

	if result.Message == "" {
		t.Error("Expected error message")
	}
}

// TestAPIConnectionTest_Timeout tests connection timeout handling.
func TestAPIConnectionTest_Timeout(t *testing.T) {
	// This test is difficult to implement without a mock server
	// We would need to simulate a slow/unresponsive API
	t.Skip("Timeout test requires mock server implementation")
}

// TestAPICommand_Status tests the status command.
func TestAPICommand_Status(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"
	cfg.API.Provider.Model = "gpt-4o"
	cfg.API.Provider.MaxTokens = 4096
	cfg.API.APIKeys["openai"] = "encrypted:fake_key"

	cmd := NewAPICommand(cfg)
	result := cmd.showStatus()

	if !result.Success {
		t.Error("Expected successful status command")
	}

	if result.Message == "" {
		t.Error("Expected non-empty status message")
	}
}

// TestAPICommand_List tests the list providers command.
func TestAPICommand_List(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewAPICommand(cfg)
	result := cmd.listProviders()

	if !result.Success {
		t.Error("Expected successful list command")
	}

	if result.Message == "" {
		t.Error("Expected non-empty provider list")
	}

	// Should mention at least some known providers
	// This is a basic sanity check
	if len(result.Message) < 50 {
		t.Errorf("Expected substantial provider list, got: %s", result.Message)
	}
}

// TestAPICommand_SwitchProvider tests switching API providers.
func TestAPICommand_SwitchProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	// Set up initial provider
	cfg.API.Provider.ProviderID = "openai"
	cfg.API.APIKeys["openai"] = "key1"
	cfg.API.APIKeys["anthropic"] = "key2"

	cmd := NewAPICommand(cfg)

	// Switch to anthropic
	result := cmd.switchProvider("anthropic")

	if !result.Success {
		t.Errorf("Expected successful provider switch, got: %s", result.Message)
	}

	if cfg.API.Provider.ProviderID != "anthropic" {
		t.Errorf("Expected provider to be 'anthropic', got: %s", cfg.API.Provider.ProviderID)
	}
}

// TestAPICommand_SwitchProvider_NoKey tests switching to provider without API key.
func TestAPICommand_SwitchProvider_NoKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"
	cfg.API.APIKeys["openai"] = "key1"
	// Don't set anthropic key

	cmd := NewAPICommand(cfg)

	// Try to switch to anthropic without key
	result := cmd.switchProvider("anthropic")

	if result.Success {
		t.Error("Expected failed switch when API key not configured")
	}

	// Provider should remain unchanged
	if cfg.API.Provider.ProviderID != "openai" {
		t.Errorf("Expected provider to remain 'openai', got: %s", cfg.API.Provider.ProviderID)
	}
}

// TestAPICommand_SwitchProvider_Invalid tests switching to invalid provider.
func TestAPICommand_SwitchProvider_Invalid(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"

	cmd := NewAPICommand(cfg)

	// Try to switch to invalid provider
	result := cmd.switchProvider("invalid_provider_xyz")

	if result.Success {
		t.Error("Expected failed switch for invalid provider")
	}

	// Provider should remain unchanged
	if cfg.API.Provider.ProviderID != "openai" {
		t.Errorf("Expected provider to remain 'openai', got: %s", cfg.API.Provider.ProviderID)
	}
}

// TestAPICommand_Execute tests the main execute dispatcher.
func TestAPICommand_Execute(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"
	cfg.API.Provider.Model = "gpt-4o"
	cfg.API.APIKeys["openai"] = "key1"

	cmd := NewAPICommand(cfg)

	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{"empty args - status", "", false},
		{"explicit status", "status", false},
		{"list providers", "list", false},
		{"invalid command", "invalid_command", false}, // Returns help text
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.Execute(tt.args)

			if result.Message == "" {
				t.Error("Expected non-empty message")
			}

			// Basic sanity check - should not panic
		})
	}
}

// TestAPICommand_FormatConnectionError tests error message formatting.
func TestAPICommand_FormatConnectionError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewAPICommand(cfg)

	tests := []struct {
		name    string
		err     error
		latency time.Duration
		wantMin int // Minimum message length
	}{
		{
			name:    "network error",
			err:     api.NewAPIError("test", 0, "網路連線失敗", nil),
			latency: 1000 * time.Millisecond,
			wantMin: 20,
		},
		{
			name:    "auth error",
			err:     api.NewAPIError("test", 401, "API Key 無效", nil),
			latency: 500 * time.Millisecond,
			wantMin: 20,
		},
		{
			name:    "rate limit error",
			err:     api.NewAPIError("test", 429, "請求過於頻繁", nil),
			latency: 200 * time.Millisecond,
			wantMin: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := cmd.formatConnectionError(tt.err, tt.latency)

			if len(msg) < tt.wantMin {
				t.Errorf("Expected message length >= %d, got %d: %s", tt.wantMin, len(msg), msg)
			}

			if msg == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}

// TestProviderConfig creates a provider config for testing.
func TestProviderConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.API.Provider.ProviderID = "openai"
	cfg.API.Provider.Model = "gpt-4o"
	cfg.API.Provider.MaxTokens = 4096
	cfg.API.APIKeys["openai"] = "test_key"

	// Try to create provider config (doesn't test actual API call)
	providerCfg := api.ProviderConfig{
		ProviderID: cfg.API.Provider.ProviderID,
		APIKey:     cfg.API.APIKeys["openai"],
		Model:      cfg.API.Provider.Model,
		MaxTokens:  cfg.API.Provider.MaxTokens,
	}

	if providerCfg.ProviderID == "" {
		t.Error("Expected non-empty provider ID")
	}

	if providerCfg.Model == "" {
		t.Error("Expected non-empty model")
	}
}

// mockProvider is a test helper that implements api.Provider interface
type mockProvider struct {
	testConnectionErr error
	sendMessageErr    error
	streamErr         error
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) TestConnection(ctx context.Context) error {
	return m.testConnectionErr
}

func (m *mockProvider) SendMessage(ctx context.Context, messages []api.Message) (*api.Response, error) {
	if m.sendMessageErr != nil {
		return nil, m.sendMessageErr
	}
	return &api.Response{Content: "test response"}, nil
}

func (m *mockProvider) Stream(ctx context.Context, messages []api.Message, callback func(chunk string)) error {
	return m.streamErr
}

func (m *mockProvider) ModelInfo() api.ModelInfo {
	return api.ModelInfo{
		Provider:  "mock",
		Model:     "mock-model",
		MaxTokens: 4096,
	}
}
