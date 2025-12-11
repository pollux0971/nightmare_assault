package engine

import (
	"testing"
	"time"
)

func TestDefaultTypewriterConfig(t *testing.T) {
	config := DefaultTypewriterConfig()

	if config.MinCharsPerSecond != 50 {
		t.Errorf("MinCharsPerSecond = %d, want 50", config.MinCharsPerSecond)
	}
	if config.MaxCharsPerSecond != 80 {
		t.Errorf("MaxCharsPerSecond = %d, want 80", config.MaxCharsPerSecond)
	}
	if config.PunctuationDelay != 100*time.Millisecond {
		t.Errorf("PunctuationDelay = %v, want 100ms", config.PunctuationDelay)
	}
	if config.ParagraphDelay != 300*time.Millisecond {
		t.Errorf("ParagraphDelay = %v, want 300ms", config.ParagraphDelay)
	}
}

func TestNewStreamBuffer(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	if buffer == nil {
		t.Fatal("NewStreamBuffer should not return nil")
	}
	if buffer.state != TypewriterIdle {
		t.Errorf("Initial state = %v, want TypewriterIdle", buffer.state)
	}
}

func TestStreamBuffer_Append(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	buffer.Append("Hello")
	buffer.Append(" World")

	full := buffer.GetFull()
	if full != "Hello World" {
		t.Errorf("GetFull = %q, want 'Hello World'", full)
	}
}

func TestStreamBuffer_State(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	if buffer.State() != TypewriterIdle {
		t.Errorf("Initial state should be TypewriterIdle")
	}
}

func TestStreamBuffer_Progress(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	// Empty buffer should be 100% complete
	if buffer.Progress() != 100 {
		t.Errorf("Empty buffer progress = %d, want 100", buffer.Progress())
	}

	// With content, should be 0% before display
	buffer.Append("Hello")
	if buffer.Progress() != 0 {
		t.Errorf("Before display, progress = %d, want 0", buffer.Progress())
	}
}

func TestStreamBuffer_GetDisplayed(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	buffer.Append("Hello")

	// Before animation, displayed should be empty
	displayed := buffer.GetDisplayed()
	if displayed != "" {
		t.Errorf("Before animation, GetDisplayed = %q, want ''", displayed)
	}
}

func TestIsPunctuation(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'.', true},
		{'!', true},
		{'?', true},
		{'。', true},
		{'！', true},
		{'？', true},
		{'，', true},
		{',', true},
		{'a', false},
		{' ', false},
		{'\n', false},
	}

	for _, tt := range tests {
		if got := isPunctuation(tt.r); got != tt.expected {
			t.Errorf("isPunctuation(%q) = %v, want %v", tt.r, got, tt.expected)
		}
	}
}

func TestNewStreamingRenderer(t *testing.T) {
	config := DefaultTypewriterConfig()
	renderer := NewStreamingRenderer(config)

	if renderer == nil {
		t.Fatal("NewStreamingRenderer should not return nil")
	}
	if renderer.buffer == nil {
		t.Error("Renderer should have a buffer")
	}
}

func TestStreamingRenderer_AppendChunk(t *testing.T) {
	config := DefaultTypewriterConfig()
	renderer := NewStreamingRenderer(config)

	renderer.AppendChunk("Hello")
	renderer.AppendChunk(" World")

	// Content is in buffer, not yet rendered
	full := renderer.buffer.GetFull()
	if full != "Hello World" {
		t.Errorf("Buffer full content = %q, want 'Hello World'", full)
	}
}

func TestStreamingRenderer_GetContent(t *testing.T) {
	config := DefaultTypewriterConfig()
	renderer := NewStreamingRenderer(config)

	// Before rendering
	if renderer.GetContent() != "" {
		t.Error("Before rendering, content should be empty")
	}
}

func TestStreamingRenderer_Progress(t *testing.T) {
	config := DefaultTypewriterConfig()
	renderer := NewStreamingRenderer(config)

	// Empty
	if renderer.Progress() != 100 {
		t.Errorf("Empty progress = %d, want 100", renderer.Progress())
	}

	// With content
	renderer.AppendChunk("Hello")
	if renderer.Progress() != 0 {
		t.Errorf("Before render, progress = %d, want 0", renderer.Progress())
	}
}

func TestRuneCount(t *testing.T) {
	tests := []struct {
		s        string
		expected int
	}{
		{"hello", 5},
		{"你好", 2},
		{"hello世界", 7},
		{"", 0},
	}

	for _, tt := range tests {
		if got := RuneCount(tt.s); got != tt.expected {
			t.Errorf("RuneCount(%q) = %d, want %d", tt.s, got, tt.expected)
		}
	}
}

func TestTypewriterState_Constants(t *testing.T) {
	// Verify state constants
	states := []TypewriterState{
		TypewriterIdle,
		TypewriterPlaying,
		TypewriterPaused,
		TypewriterSkipped,
		TypewriterDone,
	}
	if len(states) != 5 {
		t.Error("Should have 5 typewriter states")
	}
}

func TestStreamBuffer_SetCallbacks(t *testing.T) {
	config := DefaultTypewriterConfig()
	buffer := NewStreamBuffer(config)

	buffer.SetCallbacks(
		func(r rune) { _ = r },
		func() {},
	)

	if buffer.onChar == nil {
		t.Error("onChar callback should be set")
	}
	if buffer.onComplete == nil {
		t.Error("onComplete callback should be set")
	}
}

func TestStreamingRenderer_SetCallbacks(t *testing.T) {
	config := DefaultTypewriterConfig()
	renderer := NewStreamingRenderer(config)

	renderer.SetUpdateCallback(func(content string) { _ = content })
	renderer.SetFinishCallback(func(content string) { _ = content })

	if renderer.onUpdate == nil {
		t.Error("onUpdate callback should be set")
	}
	if renderer.onFinish == nil {
		t.Error("onFinish callback should be set")
	}
}
