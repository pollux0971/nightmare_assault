package input

import (
	"testing"
	"time"
)

func TestInputMode_Constants(t *testing.T) {
	modes := []InputMode{InputModeChoice, InputModeFreeText, InputModeCommand}
	if len(modes) != 3 {
		t.Error("Should have 3 input modes")
	}
}

func TestNewPlayerInput(t *testing.T) {
	input := NewPlayerInput(InputModeChoice, "1")

	if input.Mode != InputModeChoice {
		t.Errorf("Mode = %v, want InputModeChoice", input.Mode)
	}
	if input.Raw != "1" {
		t.Errorf("Raw = %q, want '1'", input.Raw)
	}
	if input.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestNewInputHandler(t *testing.T) {
	handler := NewInputHandler()

	if handler == nil {
		t.Fatal("NewInputHandler should not return nil")
	}
	if handler.currentMode != InputModeChoice {
		t.Errorf("Initial mode = %v, want InputModeChoice", handler.currentMode)
	}
	if handler.validator == nil {
		t.Error("Handler should have validator")
	}
}

func TestInputHandler_GetMode(t *testing.T) {
	handler := NewInputHandler()

	if handler.GetMode() != InputModeChoice {
		t.Errorf("GetMode = %v, want InputModeChoice", handler.GetMode())
	}
}

func TestInputHandler_SetMode(t *testing.T) {
	handler := NewInputHandler()

	handler.SetMode(InputModeFreeText)

	if handler.GetMode() != InputModeFreeText {
		t.Errorf("After SetMode, mode = %v, want InputModeFreeText", handler.GetMode())
	}

	// Buffer should be cleared on mode change
	handler.buffer = "test"
	handler.SetMode(InputModeChoice)
	if handler.buffer != "" {
		t.Error("Buffer should be cleared on mode change")
	}
}

func TestInputHandler_ProcessChoiceInput(t *testing.T) {
	handler := NewInputHandler()

	tests := []struct {
		name        string
		input       string
		maxChoices  int
		expectValid bool
		expectError bool
	}{
		{"Valid choice 1", "1", 3, true, false},
		{"Valid choice 2", "2", 3, true, false},
		{"Invalid choice 0", "0", 3, false, true},
		{"Invalid choice 5", "5", 3, false, true},
		{"Invalid text", "a", 3, false, true},
		{"Empty input", "", 3, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.ProcessChoiceInput(tt.input, tt.maxChoices)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != nil && result.Valid != tt.expectValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.expectValid)
			}
		})
	}
}

func TestInputHandler_ProcessFreeTextInput(t *testing.T) {
	handler := NewInputHandler()

	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError bool
	}{
		{"Valid input", "I want to explore the room", true, false},
		{"Too short", "ab", false, true},
		{"Too long", string(make([]byte, 201)), false, true},
		{"Empty input", "", false, true},
		{"Whitespace only", "   ", false, true},
		{"Valid Chinese", "我要進入房間", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.ProcessFreeTextInput(tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != nil && result.Valid != tt.expectValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.expectValid)
			}
		})
	}
}

func TestInputHandler_ProcessCommandInput(t *testing.T) {
	handler := NewInputHandler()

	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError bool
	}{
		{"Valid command", "/help", true, false},
		{"Command with args", "/save game1", true, false},
		{"Not a command", "help", false, true},
		{"Empty command", "/", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.ProcessCommandInput(tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != nil && result.Valid != tt.expectValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.expectValid)
			}
		})
	}
}

func TestInputHandler_DetectInputMode(t *testing.T) {
	handler := NewInputHandler()

	tests := []struct {
		input        string
		expectedMode InputMode
	}{
		{"/help", InputModeCommand},
		{"/save", InputModeCommand},
		{"1", InputModeChoice},
		{"2", InputModeChoice},
		{"explore the room", InputModeFreeText},
		{"I want to leave", InputModeFreeText},
	}

	for _, tt := range tests {
		mode := handler.DetectInputMode(tt.input)
		if mode != tt.expectedMode {
			t.Errorf("DetectInputMode(%q) = %v, want %v", tt.input, mode, tt.expectedMode)
		}
	}
}

func TestInputHandler_GetBuffer(t *testing.T) {
	handler := NewInputHandler()

	handler.buffer = "test buffer"
	if handler.GetBuffer() != "test buffer" {
		t.Errorf("GetBuffer = %q, want 'test buffer'", handler.GetBuffer())
	}
}

func TestInputHandler_AppendToBuffer(t *testing.T) {
	handler := NewInputHandler()

	handler.AppendToBuffer("hello")
	handler.AppendToBuffer(" world")

	if handler.buffer != "hello world" {
		t.Errorf("Buffer = %q, want 'hello world'", handler.buffer)
	}
}

func TestInputHandler_ClearBuffer(t *testing.T) {
	handler := NewInputHandler()

	handler.buffer = "test"
	handler.ClearBuffer()

	if handler.buffer != "" {
		t.Errorf("After clear, buffer = %q, want empty", handler.buffer)
	}
}

func TestPlayerInput_Timestamp(t *testing.T) {
	before := time.Now()
	input := NewPlayerInput(InputModeChoice, "1")
	after := time.Now()

	if input.Timestamp.Before(before) || input.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
}
