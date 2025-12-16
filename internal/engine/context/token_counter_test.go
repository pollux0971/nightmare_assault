package context

import (
	"testing"
)

// TestTokenCounter_Interface 測試接口實現
func TestTokenCounter_Interface(t *testing.T) {
	var _ TokenCounter = (*TiktokenCounter)(nil)
	var _ TokenCounter = (*EstimateTokenCounter)(nil)
	var _ TokenCounter = (*MockTokenCounter)(nil)
}

// TestTiktokenCounter_NewCounter 測試創建 counter
func TestTiktokenCounter_NewCounter(t *testing.T) {
	counter, err := NewTiktokenCounter("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create TiktokenCounter: %v", err)
	}
	if counter == nil {
		t.Error("Counter should not be nil")
	}
	if counter.GetEncodingName() == "" {
		t.Error("Encoding name should not be empty")
	}
}

// TestTiktokenCounter_CountTokens 測試 Token 計數
func TestTiktokenCounter_CountTokens(t *testing.T) {
	counter, err := NewTiktokenCounter("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		minTokens int // 最小預期 tokens（避免精確值依賴）
		maxTokens int // 最大預期 tokens
	}{
		{
			name:     "Empty string",
			text:     "",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:     "Simple English",
			text:     "Hello, world!",
			minTokens: 3,
			maxTokens: 5,
		},
		{
			name:     "Chinese text",
			text:     "你好，世界！",
			minTokens: 5,
			maxTokens: 10,
		},
		{
			name:     "Mixed English and Chinese",
			text:     "Hello 世界",
			minTokens: 2,
			maxTokens: 6,
		},
		{
			name:     "Long text",
			text:     "This is a longer piece of text that should consume more tokens. It contains multiple sentences and punctuation marks.",
			minTokens: 20,
			maxTokens: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := counter.CountTokens(tt.text)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("CountTokens(%q) = %d, expected between %d and %d", tt.text, tokens, tt.minTokens, tt.maxTokens)
			}
		})
	}
}

// TestEstimateTokenCounter_CountTokens 測試估算計數器
func TestEstimateTokenCounter_CountTokens(t *testing.T) {
	counter := NewEstimateTokenCounter()

	tests := []struct {
		name     string
		text     string
		minTokens int
	}{
		{
			name:     "Empty string",
			text:     "",
			minTokens: 0,
		},
		{
			name:     "English text",
			text:     "Hello world",
			minTokens: 2, // "Hello world" = 11 chars / 4 ≈ 2.75 * 1.1 ≈ 3
		},
		{
			name:     "Chinese text",
			text:     "你好世界",
			minTokens: 2, // 4 chars / 2 = 2 * 1.1 ≈ 2
		},
		{
			name:     "Long mixed text",
			text:     "This is English text. 這是中文文字。",
			minTokens: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := counter.CountTokens(tt.text)
			if tokens < tt.minTokens {
				t.Errorf("CountTokens(%q) = %d, expected at least %d", tt.text, tokens, tt.minTokens)
			}
		})
	}
}

// TestEstimateTokenCounter_OverestimatesBias 測試高估傾向
func TestEstimateTokenCounter_OverestimatesBias(t *testing.T) {
	estimate := NewEstimateTokenCounter()
	tiktoken, err := NewTiktokenCounter("gpt-4")
	if err != nil {
		t.Skipf("Tiktoken not available: %v", err)
	}

	text := "This is a test sentence with some Chinese characters: 測試文字"

	estimateTokens := estimate.CountTokens(text)
	actualTokens := tiktoken.CountTokens(text)

	// 估算應該略微高估（但不要太離譜）
	if estimateTokens < actualTokens {
		t.Errorf("Estimate (%d) should overestimate actual (%d)", estimateTokens, actualTokens)
	}

	// 但不應該超過 2 倍
	if estimateTokens > actualTokens*2 {
		t.Errorf("Estimate (%d) overestimates too much compared to actual (%d)", estimateTokens, actualTokens)
	}
}

// MockTokenCounter for testing
type MockTokenCounter struct {
	TokensPerCall int
	CallCount     int
	EncodingName  string
}

func (m *MockTokenCounter) CountTokens(text string) int {
	m.CallCount++
	if m.TokensPerCall > 0 {
		return m.TokensPerCall
	}
	// Default: simple length-based mock
	return len(text) / 4
}

func (m *MockTokenCounter) GetEncodingName() string {
	if m.EncodingName != "" {
		return m.EncodingName
	}
	return "mock_encoding"
}
