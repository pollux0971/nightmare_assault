package context

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// MockLLMClient is a mock implementation for testing
type MockLLMClient struct {
	ResponseToReturn string
	ErrorToReturn    error
	CallCount        int
	LastRequest      *GenerateRequest // Store last request for verification
}

func (m *MockLLMClient) Generate(ctx context.Context, req *GenerateRequest) (string, error) {
	m.CallCount++
	m.LastRequest = req // Store for test verification
	if m.ErrorToReturn != nil {
		return "", m.ErrorToReturn
	}
	return m.ResponseToReturn, nil
}

// TestLLMSummaryGenerator_Success 測試成功生成摘要
func TestLLMSummaryGenerator_Success(t *testing.T) {
	mockClient := &MockLLMClient{
		ResponseToReturn: "[Chapter 1 Summary: 玩家進入廢棄醫院. 角色狀態: 全員存活. 已知線索: 血跡. 當前目標: 尋找出口.]",
	}

	generator := NewLLMSummaryGenerator(mockClient)

	entries := []HistoryEntry{
		{Beat: 1, PlayerChoice: "進入", StoryContent: "你看到血跡", CluesFound: []string{"blood"}},
	}

	summary, err := generator.GenerateSummary(context.Background(), entries)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	if !strings.Contains(summary, "Chapter") {
		t.Error("Summary should contain Chapter format")
	}
	if mockClient.CallCount != 1 {
		t.Errorf("Expected 1 API call, got %d", mockClient.CallCount)
	}
}

// TestLLMSummaryGenerator_APIError 測試 API 錯誤
func TestLLMSummaryGenerator_APIError(t *testing.T) {
	mockClient := &MockLLMClient{
		ErrorToReturn: errors.New("API Error"),
	}

	generator := NewLLMSummaryGenerator(mockClient)

	entries := []HistoryEntry{
		{Beat: 1, PlayerChoice: "進入"},
	}

	_, err := generator.GenerateSummary(context.Background(), entries)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestLLMSummaryGenerator_EmptyResponse 測試空響應
func TestLLMSummaryGenerator_EmptyResponse(t *testing.T) {
	mockClient := &MockLLMClient{
		ResponseToReturn: "",
	}

	generator := NewLLMSummaryGenerator(mockClient)

	entries := []HistoryEntry{
		{Beat: 1, PlayerChoice: "進入"},
	}

	_, err := generator.GenerateSummary(context.Background(), entries)

	if err == nil {
		t.Error("Expected error for empty response")
	}
}

// TestLLMSummaryGenerator_FormatValidation 測試格式驗證
func TestLLMSummaryGenerator_FormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		shouldError bool
	}{
		{
			name:        "Valid format with Chapter",
			response:    "[Chapter 1 Summary: Test. 角色狀態: OK. 已知線索: A. 當前目標: B.]",
			shouldError: false,
		},
		{
			name:        "Invalid format - missing Chapter",
			response:    "Just some text without proper format",
			shouldError: true,
		},
		{
			name:        "Valid format - complete",
			response:    "[Chapter 2 Summary: 簡要劇情. 角色狀態: 全員存活. 已知線索: 無. 當前目標: 探索.]",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockLLMClient{
				ResponseToReturn: tt.response,
			}

			generator := NewLLMSummaryGenerator(mockClient)
			entries := []HistoryEntry{{Beat: 1}}

			_, err := generator.GenerateSummary(context.Background(), entries)

			if tt.shouldError && err == nil {
				t.Error("Expected error for invalid format")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestSummaryGenerator_Interface 測試接口實現
func TestSummaryGenerator_Interface(t *testing.T) {
	var _ SummaryGenerator = (*LLMSummaryGenerator)(nil)
	var _ SummaryGenerator = (*MockSummaryGenerator)(nil)
}

// MockSummaryGenerator for testing ContextWindow integration
type MockSummaryGenerator struct {
	SummaryToReturn string
	ErrorToReturn   error
	CallCount       int
}

func (m *MockSummaryGenerator) GenerateSummary(ctx context.Context, entries []HistoryEntry) (string, error) {
	m.CallCount++
	if m.ErrorToReturn != nil {
		return "", m.ErrorToReturn
	}
	return m.SummaryToReturn, nil
}
