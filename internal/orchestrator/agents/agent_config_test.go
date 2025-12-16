package agents

import (
	"context"
	"testing"
	"time"
)

// TestDefaultAgentConfig 測試默認配置
func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	// 驗證默認值
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", config.MaxRetries)
	}

	if config.Name != "" {
		t.Errorf("Expected empty name for default config, got %s", config.Name)
	}
}

// TestAgentConfigCreation 測試配置創建
func TestAgentConfigCreation(t *testing.T) {
	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    45 * time.Second,
		MaxRetries: 5,
	}

	if config.Name != "TestAgent" {
		t.Errorf("Expected name 'TestAgent', got %s", config.Name)
	}

	if config.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", config.Timeout)
	}

	if config.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", config.MaxRetries)
	}
}

// TestValidateConfig 測試配置驗證
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  AgentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: AgentConfig{
				Name:       "ValidAgent",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: AgentConfig{
				Name:       "",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			wantErr: true,
			errMsg:  "agent name cannot be empty",
		},
		{
			name: "zero timeout",
			config: AgentConfig{
				Name:       "TestAgent",
				Timeout:    0,
				MaxRetries: 3,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "negative max retries",
			config: AgentConfig{
				Name:       "TestAgent",
				Timeout:    30 * time.Second,
				MaxRetries: -1,
			},
			wantErr: true,
			errMsg:  "max retries cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateConfig() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAgentConfigWithLLMClient 測試配置包含 LLMClient
func TestAgentConfigWithLLMClient(t *testing.T) {
	// 創建一個 mock LLM client
	mockClient := &mockLLMClient{}

	config := AgentConfig{
		Name:       "TestAgent",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		LLMClient:  mockClient,
	}

	if config.LLMClient == nil {
		t.Error("Expected LLMClient to be set")
	}

	// 驗證可以將 config.LLMClient 轉換為 LLMClient 接口
	var _ LLMClient = config.LLMClient
}

// TestAgentConfigWithPromptLoader 測試配置包含 PromptLoader
func TestAgentConfigWithPromptLoader(t *testing.T) {
	// 創建一個 mock prompt loader
	mockLoader := &mockPromptLoader{}

	config := AgentConfig{
		Name:         "TestAgent",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		PromptLoader: mockLoader,
	}

	if config.PromptLoader == nil {
		t.Error("Expected PromptLoader to be set")
	}

	// 驗證可以將 config.PromptLoader 轉換為 PromptLoader 接口
	var _ PromptLoader = config.PromptLoader
}

// mockLLMClient 用於測試的 mock LLM client
type mockLLMClient struct{}

func (m *mockLLMClient) Generate(ctx context.Context, prompt string, options map[string]any) (string, error) {
	return "mock response", nil
}

// mockPromptLoader 用於測試的 mock prompt loader
type mockPromptLoader struct{}

func (m *mockPromptLoader) LoadTemplate(name string) (string, error) {
	return "mock template", nil
}

func (m *mockPromptLoader) RenderTemplate(template string, data any) (string, error) {
	return "rendered template", nil
}
