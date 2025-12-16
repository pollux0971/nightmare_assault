package agents

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBaseAgentIntegration_FullFlow 測試完整的 Agent 調用流程
//
// 此測試模擬一個完整的 Agent 調用流程：
// 1. 創建 AgentConfig 並配置 LLMClient 和 PromptLoader
// 2. 使用 PromptLoader 加載模板
// 3. 使用 PromptLoader 渲染模板
// 4. 使用 LLMClient 生成回應
// 5. 驗證整個流程正常工作
func TestBaseAgentIntegration_FullFlow(t *testing.T) {
	// 創建臨時目錄和測試模板
	tmpDir, err := os.MkdirTemp("", "agent-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 創建測試模板文件
	templateContent := "Generate a story about {{.Topic}} in {{.Genre}} genre."
	templatePath := filepath.Join(tmpDir, "story-prompt.txt")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// 配置組件
	llmClient := NewMockLLMClient("Once upon a time in a dark forest...")
	promptLoader := NewFilePromptLoader(tmpDir)

	config := AgentConfig{
		Name:         "StoryAgent",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		LLMClient:    llmClient,
		PromptLoader: promptLoader,
	}

	// 驗證配置有效
	if err := ValidateConfig(config); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// 創建 BaseAgentImpl
	impl := NewBaseAgentImpl(config)

	// 定義完整的調用流程
	invokeFn := func(ctx context.Context) (any, error) {
		// 1. 加載模板
		template, err := promptLoader.LoadTemplate("story-prompt.txt")
		if err != nil {
			return nil, err
		}

		// 2. 渲染模板
		data := map[string]any{
			"Topic": "dragons",
			"Genre": "fantasy",
		}
		prompt, err := promptLoader.RenderTemplate(template, data)
		if err != nil {
			return nil, err
		}

		// 3. 調用 LLM
		response, err := llmClient.Generate(ctx, prompt, nil)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	// 執行完整流程
	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	if result != "Once upon a time in a dark forest..." {
		t.Errorf("Expected story response, got: %v", result)
	}

	// 驗證 LLM 被調用
	if llmClient.CallCount != 1 {
		t.Errorf("Expected 1 LLM call, got %d", llmClient.CallCount)
	}

	// 驗證 prompt 正確渲染
	expectedPrompt := "Generate a story about dragons in fantasy genre."
	if llmClient.LastPrompt != expectedPrompt {
		t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, llmClient.LastPrompt)
	}
}

// TestBaseAgentIntegration_WithRetry 測試集成重試機制
func TestBaseAgentIntegration_WithRetry(t *testing.T) {
	// 創建會重試的 LLM client（前兩次失敗，第三次成功）
	llmClient := NewRetryMockLLMClient(2, CommonErrors.HTTP503, "Success after retry")

	config := AgentConfig{
		Name:       "RetryAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		LLMClient:  llmClient,
	}

	impl := NewBaseAgentImpl(config)

	invokeFn := func(ctx context.Context) (any, error) {
		return llmClient.Generate(ctx, "test prompt", nil)
	}

	result, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if result != "Success after retry" {
		t.Errorf("Expected 'Success after retry', got: %v", result)
	}

	// 驗證重試了 3 次
	if llmClient.CallCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", llmClient.CallCount)
	}
}

// TestBaseAgentIntegration_WithTimeout 測試集成超時機制
func TestBaseAgentIntegration_WithTimeout(t *testing.T) {
	// 創建延遲超過超時時間的 LLM client
	llmClient := NewMockLLMClientWithDelay("Should not return", 200*time.Millisecond)

	config := AgentConfig{
		Name:       "TimeoutAgent",
		Timeout:    50 * time.Millisecond, // 短超時
		MaxRetries: 3,
		LLMClient:  llmClient,
	}

	impl := NewBaseAgentImpl(config)

	invokeFn := func(ctx context.Context) (any, error) {
		return llmClient.Generate(ctx, "test prompt", nil)
	}

	_, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// 驗證錯誤是 AgentError
	var agentErr *AgentError
	if !errors.As(err, &agentErr) {
		t.Fatalf("Expected AgentError, got: %T", err)
	}

	// 驗證是超時錯誤且不可重試
	if agentErr.Retryable {
		t.Error("Expected timeout error to be non-retryable")
	}

	// 只應該調用一次（超時不重試）
	if llmClient.CallCount != 1 {
		t.Errorf("Expected 1 call (no retry on timeout), got %d", llmClient.CallCount)
	}
}

// TestBaseAgentIntegration_PromptLoaderWithMultipleTemplates 測試加載多個模板
func TestBaseAgentIntegration_PromptLoaderWithMultipleTemplates(t *testing.T) {
	// 創建臨時目錄和多個測試模板
	tmpDir, err := os.MkdirTemp("", "agent-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 創建多個模板文件
	templates := map[string]string{
		"greeting.txt":    "Hello {{.Name}}!",
		"farewell.txt":    "Goodbye {{.Name}}, see you later!",
		"narration.txt":   "The story of {{.Hero}} in {{.Location}}.",
	}

	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", name, err)
		}
	}

	// 創建 PromptLoader
	loader := NewFilePromptLoader(tmpDir)

	// 測試加載所有模板
	for name, expectedContent := range templates {
		content, err := loader.LoadTemplate(name)
		if err != nil {
			t.Errorf("Failed to load template %s: %v", name, err)
		}

		if content != expectedContent {
			t.Errorf("Template %s: expected '%s', got '%s'", name, expectedContent, content)
		}
	}

	// 測試渲染模板
	data := map[string]any{
		"Name":     "Alice",
		"Hero":     "Knight",
		"Location": "Castle",
	}

	// 渲染 greeting
	greeting, err := loader.LoadTemplate("greeting.txt")
	if err != nil {
		t.Fatalf("Failed to load greeting: %v", err)
	}
	renderedGreeting, err := loader.RenderTemplate(greeting, data)
	if err != nil {
		t.Fatalf("Failed to render greeting: %v", err)
	}
	if renderedGreeting != "Hello Alice!" {
		t.Errorf("Expected 'Hello Alice!', got '%s'", renderedGreeting)
	}

	// 渲染 narration
	narration, err := loader.LoadTemplate("narration.txt")
	if err != nil {
		t.Fatalf("Failed to load narration: %v", err)
	}
	renderedNarration, err := loader.RenderTemplate(narration, data)
	if err != nil {
		t.Fatalf("Failed to render narration: %v", err)
	}
	if renderedNarration != "The story of Knight in Castle." {
		t.Errorf("Expected 'The story of Knight in Castle.', got '%s'", renderedNarration)
	}
}

// TestBaseAgentIntegration_ErrorPropagation 測試錯誤在組件間傳播
func TestBaseAgentIntegration_ErrorPropagation(t *testing.T) {
	// 創建會返回錯誤的 LLM client
	llmClient := NewMockLLMClientWithError(CommonErrors.HTTP400)

	config := AgentConfig{
		Name:       "ErrorAgent",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		LLMClient:  llmClient,
	}

	impl := NewBaseAgentImpl(config)

	invokeFn := func(ctx context.Context) (any, error) {
		return llmClient.Generate(ctx, "test prompt", nil)
	}

	_, err := impl.InvokeWithRetry(context.Background(), invokeFn)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// 驗證錯誤被正確分類為 AgentError
	var agentErr *AgentError
	if !errors.As(err, &agentErr) {
		t.Fatalf("Expected AgentError, got: %T", err)
	}

	// 驗證錯誤包含正確的 Agent 名稱
	if agentErr.AgentName != "ErrorAgent" {
		t.Errorf("Expected AgentName 'ErrorAgent', got '%s'", agentErr.AgentName)
	}

	// 驗證底層錯誤被保留
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Error("Expected to find HTTPError in error chain")
	}

	// HTTP 400 不應重試，只調用一次
	if llmClient.CallCount != 1 {
		t.Errorf("Expected 1 call (no retry on 400), got %d", llmClient.CallCount)
	}
}
