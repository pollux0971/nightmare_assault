package integration

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// ===== Mock Providers =====

// MockProvider is a configurable mock LLM provider for testing
type MockProvider struct {
	mu              sync.Mutex
	shouldFail      bool
	failureError    error
	responseContent string
	responseDelay   time.Duration
	callCount       int
	lastMessages    []client.Message
}

// NewMockProvider creates a new successful mock provider
func NewMockProvider(content string) *MockProvider {
	return &MockProvider{
		shouldFail:      false,
		responseContent: content,
		responseDelay:   0,
		callCount:       0,
	}
}

// NewFailingMockProvider creates a mock provider that always fails
func NewFailingMockProvider(err error) *MockProvider {
	return &MockProvider{
		shouldFail:   true,
		failureError: err,
		callCount:    0,
	}
}

// NewDelayedMockProvider creates a mock provider with artificial delay
func NewDelayedMockProvider(content string, delay time.Duration) *MockProvider {
	return &MockProvider{
		shouldFail:      false,
		responseContent: content,
		responseDelay:   delay,
		callCount:       0,
	}
}

// SendMessage implements client.Provider interface
func (m *MockProvider) SendMessage(ctx context.Context, messages []client.Message) (*client.Response, error) {
	m.mu.Lock()
	m.callCount++
	m.lastMessages = messages
	delay := m.responseDelay
	shouldFail := m.shouldFail
	failureError := m.failureError
	content := m.responseContent
	m.mu.Unlock()

	// Simulate delay if configured
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Return failure if configured
	if shouldFail {
		if failureError != nil {
			return nil, failureError
		}
		return nil, errors.New("mock provider failure")
	}

	// Return success response
	return &client.Response{
		Content:  content,
		Metadata: make(map[string]interface{}),
	}, nil
}

// Name implements client.Provider interface
func (m *MockProvider) Name() string {
	return "MockProvider"
}

// TestConnection implements client.Provider interface
func (m *MockProvider) TestConnection(ctx context.Context) error {
	return nil
}

// Stream implements client.Provider interface
func (m *MockProvider) Stream(ctx context.Context, messages []client.Message, callback func(chunk string)) error {
	// For mock, just call SendMessage and return the content via callback
	resp, err := m.SendMessage(ctx, messages)
	if err != nil {
		return err
	}
	callback(resp.Content)
	return nil
}

// ModelInfo implements client.Provider interface
func (m *MockProvider) ModelInfo() client.ModelInfo {
	return client.ModelInfo{
		Provider:  "mock",
		Model:     "mock-model",
		MaxTokens: 4000,
	}
}

// GetCallCount returns the number of times SendMessage was called
func (m *MockProvider) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// GetLastMessages returns the messages from the last call
func (m *MockProvider) GetLastMessages() []client.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMessages
}

// SetShouldFail configures whether the provider should fail
func (m *MockProvider) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failureError = err
}

// Reset resets the provider state
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
	m.lastMessages = nil
}

// ===== Test Fixtures =====

// TestMessage creates a test message
func TestMessage(role, content string) client.Message {
	return client.Message{
		Role:    role,
		Content: content,
	}
}

// TestMessages creates a list of test messages
func TestMessages(pairs ...string) []client.Message {
	if len(pairs)%2 != 0 {
		panic("TestMessages requires even number of arguments (role, content pairs)")
	}

	messages := make([]client.Message, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		messages = append(messages, TestMessage(pairs[i], pairs[i+1]))
	}
	return messages
}

// ThinkingResponse creates a response with thinking tags
func ThinkingResponse(thinking, content string) string {
	return "<think>\n" + thinking + "\n</think>\n\n" + content
}

// ===== Assertion Helpers =====

// AssertNoError fails the test if err is not nil
func AssertNoError(t TestingT, err error, msgAndArgs ...interface{}) {
	if err != nil {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected no error but got: %v - %v", err, msgAndArgs[0])
		} else {
			t.Fatalf("Expected no error but got: %v", err)
		}
	}
}

// AssertError fails the test if err is nil
func AssertError(t TestingT, err error, msgAndArgs ...interface{}) {
	if err == nil {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected error but got nil - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected error but got nil")
		}
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected %v but got %v - %v", expected, actual, msgAndArgs[0])
		} else {
			t.Fatalf("Expected %v but got %v", expected, actual)
		}
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected == actual {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected values to be different but both are %v - %v", expected, msgAndArgs[0])
		} else {
			t.Fatalf("Expected values to be different but both are %v", expected)
		}
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t TestingT, condition bool, msgAndArgs ...interface{}) {
	if !condition {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected condition to be true - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected condition to be true")
		}
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t TestingT, condition bool, msgAndArgs ...interface{}) {
	if condition {
		t.Helper()
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected condition to be false - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected condition to be false")
		}
	}
}

// TestingT is a minimal interface for testing
type TestingT interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Helper()
}

// ===== Performance Monitoring Helpers =====

// PerformanceMonitor tracks performance metrics during tests
type PerformanceMonitor struct {
	startTime      time.Time
	checkpoints    []Checkpoint
	mu             sync.Mutex
	memoryBaseline uint64
}

// Checkpoint represents a performance measurement point
type Checkpoint struct {
	Name      string
	Timestamp time.Time
	Duration  time.Duration
	MemoryMB  uint64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime:   time.Now(),
		checkpoints: make([]Checkpoint, 0),
	}
}

// Checkpoint records a performance checkpoint
func (pm *PerformanceMonitor) Checkpoint(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	checkpoint := Checkpoint{
		Name:      name,
		Timestamp: now,
		Duration:  now.Sub(pm.startTime),
	}
	pm.checkpoints = append(pm.checkpoints, checkpoint)
}

// GetCheckpoints returns all recorded checkpoints
func (pm *PerformanceMonitor) GetCheckpoints() []Checkpoint {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return append([]Checkpoint{}, pm.checkpoints...)
}

// Reset resets the monitor
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.startTime = time.Now()
	pm.checkpoints = make([]Checkpoint, 0)
}

// TotalDuration returns total elapsed time
func (pm *PerformanceMonitor) TotalDuration() time.Duration {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return time.Since(pm.startTime)
}

// ===== Error Helpers =====

// Common test errors
var (
	ErrMockTimeout       = errors.New("mock provider timeout")
	ErrMockRateLimit     = errors.New("mock provider rate limit exceeded")
	ErrMockInvalidModel  = errors.New("mock provider invalid model")
	ErrMockAPIError      = errors.New("mock provider API error")
	ErrMockNetworkError  = errors.New("mock provider network error")
)

// ===== Context Helpers =====

// TestContext creates a test context with timeout
func TestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// TestContextWithCancel creates a cancellable test context
func TestContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// ===== Trinity Router Helpers =====
//
// Note: These helpers are moved to individual test files to avoid import cycles
// Each test file should create its own router using:
//   trinity.NewTrinityRouterWithProviders(thinking, reactive, rapid, fallback, overrides)
