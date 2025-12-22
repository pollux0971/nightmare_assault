package trinity

import (
	"context"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// Story 9-2 Tests: Thinking Middleware

// MockProvider implements client.Provider for testing
type MockProvider struct {
	response *client.Response
	err      error
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) TestConnection(ctx context.Context) error {
	return nil
}

func (m *MockProvider) SendMessage(ctx context.Context, messages []client.Message) (*client.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockProvider) Stream(ctx context.Context, messages []client.Message, callback func(chunk string)) error {
	return nil
}

func (m *MockProvider) ModelInfo() client.ModelInfo {
	return client.ModelInfo{Provider: "mock", Model: "test", MaxTokens: 1000}
}

// Test AC1: ThinkingMiddleware structure creation
func TestNewThinkingMiddleware(t *testing.T) {
	middleware := NewThinkingMiddleware()

	if middleware == nil {
		t.Fatal("Expected middleware to be created")
	}

	if middleware.thinkingRegex == nil {
		t.Error("Expected thinkingRegex to be initialized")
	}
}

// Test AC2: Extract thinking tags (single line)
func TestExtractThinkingSingleLine(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "Here is some text. <thinking>This is my reasoning</thinking> And more text."
	extracted := middleware.extractThinking(content)

	if extracted != "This is my reasoning" {
		t.Errorf("Expected 'This is my reasoning', got '%s'", extracted)
	}
}

// Test AC2: Extract thinking tags (multi-line)
func TestExtractThinkingMultiLine(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := `Before thinking.
<thinking>
Line 1 of thinking
Line 2 of thinking
Line 3 of thinking
</thinking>
After thinking.`

	extracted := middleware.extractThinking(content)
	expectedLines := []string{
		"Line 1 of thinking",
		"Line 2 of thinking",
		"Line 3 of thinking",
	}

	for _, line := range expectedLines {
		if !strings.Contains(extracted, line) {
			t.Errorf("Expected extracted content to contain '%s', got '%s'", line, extracted)
		}
	}
}

// Test AC2: Extract thinking with special characters
func TestExtractThinkingSpecialCharacters(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "<thinking>Complex reasoning with *special* chars: $100, 50%, &amp;</thinking>"
	extracted := middleware.extractThinking(content)

	if extracted != "Complex reasoning with *special* chars: $100, 50%, &amp;" {
		t.Errorf("Expected special characters to be preserved, got '%s'", extracted)
	}
}

// Test AC2: No thinking tags
func TestExtractThinkingNoTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "This content has no thinking tags at all."
	extracted := middleware.extractThinking(content)

	if extracted != "" {
		t.Errorf("Expected empty string for no thinking tags, got '%s'", extracted)
	}
}

// Test AC2: Multiple thinking tags (only extract first)
func TestExtractThinkingMultipleTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "<thinking>First thinking</thinking> Some text. <thinking>Second thinking</thinking>"
	extracted := middleware.extractThinking(content)

	// regex.FindStringSubmatch returns the first match
	if extracted != "First thinking" {
		t.Errorf("Expected 'First thinking' (first match), got '%s'", extracted)
	}
}

// Test AC3: Remove thinking tags
func TestRemoveThinkingTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "Before <thinking>reasoning here</thinking> After"
	cleaned := middleware.removeThinkingTags(content)

	expected := "Before  After"
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}

	// Ensure thinking content is removed
	if strings.Contains(cleaned, "thinking") {
		t.Error("Expected thinking tags to be removed")
	}
	if strings.Contains(cleaned, "reasoning here") {
		t.Error("Expected thinking content to be removed")
	}
}

// Test AC3: Remove multiple thinking tags
func TestRemoveThinkingTagsMultiple(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "Start <thinking>first</thinking> middle <thinking>second</thinking> end"
	cleaned := middleware.removeThinkingTags(content)

	expected := "Start  middle  end"
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}

	if strings.Contains(cleaned, "thinking") || strings.Contains(cleaned, "first") || strings.Contains(cleaned, "second") {
		t.Error("Expected all thinking tags and content to be removed")
	}
}

// Test AC3: Remove multi-line thinking tags
func TestRemoveThinkingTagsMultiLine(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := `Start of content
<thinking>
This is line 1
This is line 2
This is line 3
</thinking>
End of content`

	cleaned := middleware.removeThinkingTags(content)

	expected := `Start of content

End of content`

	if cleaned != expected {
		t.Errorf("Expected multi-line thinking to be removed, got '%s'", cleaned)
	}

	if strings.Contains(cleaned, "thinking") || strings.Contains(cleaned, "line 1") {
		t.Error("Expected all thinking content to be removed")
	}
}

// Test AC3: No thinking tags to remove
func TestRemoveThinkingTagsNone(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "This content has no thinking tags."
	cleaned := middleware.removeThinkingTags(content)

	if cleaned != content {
		t.Errorf("Expected content to remain unchanged, got '%s'", cleaned)
	}
}

// Test AC4: Process method - successful extraction
func TestProcessWithThinkingTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	mockProvider := &MockProvider{
		response: &client.Response{
			Content: "User answer: Yes. <thinking>I should analyze this carefully</thinking> Here is my response.",
		},
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := middleware.Process(ctx, mockProvider, messages)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that thinking chain was extracted and stored in metadata
	if resp.Metadata == nil {
		t.Fatal("Expected metadata to be populated")
	}

	thinkingChain, ok := resp.Metadata["thinking_chain"].(string)
	if !ok {
		t.Fatal("Expected thinking_chain in metadata")
	}

	if thinkingChain != "I should analyze this carefully" {
		t.Errorf("Expected thinking chain 'I should analyze this carefully', got '%s'", thinkingChain)
	}

	// Check that thinking tags were removed from content
	expectedContent := "User answer: Yes.  Here is my response."
	if resp.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, resp.Content)
	}

	if strings.Contains(resp.Content, "thinking") {
		t.Error("Expected thinking tags to be removed from content")
	}
}

// Test AC4: Process method - no thinking tags
func TestProcessWithoutThinkingTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	originalContent := "This is a normal response without thinking tags."
	mockProvider := &MockProvider{
		response: &client.Response{
			Content: originalContent,
		},
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := middleware.Process(ctx, mockProvider, messages)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that content is unchanged
	if resp.Content != originalContent {
		t.Errorf("Expected content to be unchanged, got '%s'", resp.Content)
	}

	// Check that metadata is not populated with thinking_chain
	if resp.Metadata != nil {
		if _, ok := resp.Metadata["thinking_chain"]; ok {
			t.Error("Expected no thinking_chain in metadata when no tags present")
		}
	}
}

// Test AC4: Process method - provider error
func TestProcessProviderError(t *testing.T) {
	middleware := NewThinkingMiddleware()

	mockProvider := &MockProvider{
		err: client.ErrNetworkError,
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := middleware.Process(ctx, mockProvider, messages)
	if err == nil {
		t.Error("Expected error to be propagated")
	}

	if resp != nil {
		t.Error("Expected nil response on error")
	}

	if err != client.ErrNetworkError {
		t.Errorf("Expected ErrNetworkError, got %v", err)
	}
}

// Test AC4: Process method - existing metadata preserved
func TestProcessPreservesExistingMetadata(t *testing.T) {
	middleware := NewThinkingMiddleware()

	existingMetadata := map[string]interface{}{
		"model":  "claude-opus",
		"tokens": 1000,
	}

	mockProvider := &MockProvider{
		response: &client.Response{
			Content:  "<thinking>My reasoning</thinking>Final answer",
			Metadata: existingMetadata,
		},
	}

	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	resp, err := middleware.Process(ctx, mockProvider, messages)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that existing metadata is preserved
	if resp.Metadata["model"] != "claude-opus" {
		t.Error("Expected existing metadata 'model' to be preserved")
	}

	if resp.Metadata["tokens"] != 1000 {
		t.Error("Expected existing metadata 'tokens' to be preserved")
	}

	// Check that thinking_chain was added
	if _, ok := resp.Metadata["thinking_chain"]; !ok {
		t.Error("Expected thinking_chain to be added to metadata")
	}
}

// Test AC5: Nested tags handling
func TestExtractThinkingNestedTags(t *testing.T) {
	middleware := NewThinkingMiddleware()

	// Non-greedy matching should handle nested-like structures
	content := "<thinking>Outer <thinking>Inner</thinking> Back to outer</thinking>"
	extracted := middleware.extractThinking(content)

	// Non-greedy (.*?) will match up to the first </thinking>
	// So it should extract "Outer <thinking>Inner"
	if extracted != "Outer <thinking>Inner" {
		t.Errorf("Expected 'Outer <thinking>Inner', got '%s'", extracted)
	}
}

// Test AC5: Thinking tags at start
func TestExtractThinkingAtStart(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "<thinking>Reasoning first</thinking> Then the answer."
	extracted := middleware.extractThinking(content)

	if extracted != "Reasoning first" {
		t.Errorf("Expected 'Reasoning first', got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	expected := " Then the answer."
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}
}

// Test AC5: Thinking tags at end
func TestExtractThinkingAtEnd(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "The answer is yes. <thinking>But I'm not sure why</thinking>"
	extracted := middleware.extractThinking(content)

	if extracted != "But I'm not sure why" {
		t.Errorf("Expected 'But I'm not sure why', got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	expected := "The answer is yes. "
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}
}

// Test AC5: Empty thinking tags
func TestExtractThinkingEmpty(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "Before <thinking></thinking> After"
	extracted := middleware.extractThinking(content)

	if extracted != "" {
		t.Errorf("Expected empty string for empty thinking tags, got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	expected := "Before  After"
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}
}

// Test AC5: Thinking tags with whitespace only
func TestExtractThinkingWhitespaceOnly(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "Before <thinking>   \n\t  </thinking> After"
	extracted := middleware.extractThinking(content)

	// Should extract the whitespace
	if !strings.Contains(extracted, "\n") || !strings.Contains(extracted, "\t") {
		t.Errorf("Expected whitespace to be preserved in extraction, got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	expected := "Before  After"
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}
}

// Test AC5: Case sensitivity
func TestExtractThinkingCaseSensitive(t *testing.T) {
	middleware := NewThinkingMiddleware()

	// Regex should be case-sensitive
	content := "Text <Thinking>Should not match</Thinking> More text"
	extracted := middleware.extractThinking(content)

	if extracted != "" {
		t.Errorf("Expected no match for capitalized Thinking, got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	if cleaned != content {
		t.Error("Expected content to remain unchanged for capitalized tags")
	}
}

// Test AC5: Unicode content
func TestExtractThinkingUnicode(t *testing.T) {
	middleware := NewThinkingMiddleware()

	content := "<thinking>思考：這是中文內容 🤔</thinking>答案在此"
	extracted := middleware.extractThinking(content)

	if extracted != "思考：這是中文內容 🤔" {
		t.Errorf("Expected Unicode content to be preserved, got '%s'", extracted)
	}

	cleaned := middleware.removeThinkingTags(content)
	expected := "答案在此"
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}
}

// Benchmark tests
func BenchmarkExtractThinking(b *testing.B) {
	middleware := NewThinkingMiddleware()
	content := "Start <thinking>This is a long thinking chain with lots of content that needs to be extracted efficiently</thinking> End"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.extractThinking(content)
	}
}

func BenchmarkRemoveThinkingTags(b *testing.B) {
	middleware := NewThinkingMiddleware()
	content := "Start <thinking>This is a long thinking chain with lots of content that needs to be removed efficiently</thinking> End"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.removeThinkingTags(content)
	}
}

func BenchmarkProcessWithThinking(b *testing.B) {
	middleware := NewThinkingMiddleware()
	mockProvider := &MockProvider{
		response: &client.Response{
			Content: "Start <thinking>reasoning content here</thinking> End",
		},
	}
	ctx := context.Background()
	messages := []client.Message{{Role: "user", Content: "Test"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.Process(ctx, mockProvider, messages)
	}
}
