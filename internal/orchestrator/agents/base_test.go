package agents

import (
	"context"
	"testing"
	"time"
)

// TestBaseAgentInterface 測試 BaseAgent 接口定義
func TestBaseAgentInterface(t *testing.T) {
	// 這個測試確保 BaseAgent 接口存在且可被實作
	// 我們創建一個 mock agent 來驗證接口

	var _ BaseAgent = (*mockAgent)(nil) // 編譯時檢查 mockAgent 實作 BaseAgent
}

// mockAgent 是用於測試的 BaseAgent 實作
type mockAgent struct {
	name    string
	timeout time.Duration
	result  any
	err     error
}

func (m *mockAgent) Invoke(ctx context.Context, request any) (any, error) {
	return m.result, m.err
}

func (m *mockAgent) GetName() string {
	return m.name
}

func (m *mockAgent) GetTimeout() time.Duration {
	return m.timeout
}

func (m *mockAgent) BuildPrompt(request any) (string, error) {
	return "test prompt", nil
}

func (m *mockAgent) ParseResponse(raw string) (any, error) {
	return raw, nil
}

// TestBaseAgentInvoke 測試 Invoke 方法的基本行為
func TestBaseAgentInvoke(t *testing.T) {
	agent := &mockAgent{
		name:    "TestAgent",
		timeout: 30 * time.Second,
		result:  "test response",
		err:     nil,
	}

	ctx := context.Background()
	result, err := agent.Invoke(ctx, "test request")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "test response" {
		t.Errorf("Expected 'test response', got %v", result)
	}
}

// TestBaseAgentGetName 測試 GetName 方法
func TestBaseAgentGetName(t *testing.T) {
	agent := &mockAgent{
		name: "TestAgent",
	}

	if agent.GetName() != "TestAgent" {
		t.Errorf("Expected 'TestAgent', got %s", agent.GetName())
	}
}

// TestBaseAgentGetTimeout 測試 GetTimeout 方法
func TestBaseAgentGetTimeout(t *testing.T) {
	agent := &mockAgent{
		timeout: 45 * time.Second,
	}

	if agent.GetTimeout() != 45*time.Second {
		t.Errorf("Expected 45s, got %v", agent.GetTimeout())
	}
}

// TestBaseAgentBuildPrompt 測試可選的 BuildPrompt 方法
func TestBaseAgentBuildPrompt(t *testing.T) {
	agent := &mockAgent{}

	prompt, err := agent.BuildPrompt("request data")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if prompt != "test prompt" {
		t.Errorf("Expected 'test prompt', got %s", prompt)
	}
}

// TestBaseAgentParseResponse 測試可選的 ParseResponse 方法
func TestBaseAgentParseResponse(t *testing.T) {
	agent := &mockAgent{}

	result, err := agent.ParseResponse("raw response")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "raw response" {
		t.Errorf("Expected 'raw response', got %v", result)
	}
}
