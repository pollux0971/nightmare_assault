// Package input provides player input handling and validation.
package input

import (
	"errors"
	"sync"
	"time"
)

// InputMode represents the current input mode.
type InputMode int

const (
	InputModeChoice InputMode = iota
	InputModeFreeText
	InputModeCommand
)

// Error definitions
var (
	ErrInvalidChoice      = errors.New("無效的選擇")
	ErrInputTooShort      = errors.New("輸入太短 (至少需要 3 個字元)")
	ErrInputTooLong       = errors.New("輸入太長 (最多 200 個字元)")
	ErrEmptyInput         = errors.New("輸入不能為空")
	ErrNotCommand         = errors.New("不是有效的命令")
	ErrInvalidCommand     = errors.New("無效的命令")
)

// PlayerInput represents a processed player input.
type PlayerInput struct {
	Mode      InputMode
	Raw       string
	Sanitized string
	Timestamp time.Time
	Valid     bool
	Error     error
}

// NewPlayerInput creates a new player input.
func NewPlayerInput(mode InputMode, raw string) *PlayerInput {
	return &PlayerInput{
		Mode:      mode,
		Raw:       raw,
		Timestamp: time.Now(),
	}
}

// InputHandler handles player input processing.
type InputHandler struct {
	currentMode InputMode
	buffer      string
	validator   *InputSanitizer
	mu          sync.RWMutex
}

// NewInputHandler creates a new input handler.
func NewInputHandler() *InputHandler {
	return &InputHandler{
		currentMode: InputModeChoice,
		buffer:      "",
		validator:   NewInputSanitizer(),
	}
}

// GetMode returns the current input mode.
func (h *InputHandler) GetMode() InputMode {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentMode
}

// SetMode sets the input mode and clears the buffer.
func (h *InputHandler) SetMode(mode InputMode) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentMode = mode
	h.buffer = ""
}

// ProcessChoiceInput processes a choice selection input.
func (h *InputHandler) ProcessChoiceInput(input string, maxChoices int) (*PlayerInput, error) {
	result := NewPlayerInput(InputModeChoice, input)

	// Validate choice number
	if len(input) == 0 {
		result.Error = ErrEmptyInput
		result.Valid = false
		return result, result.Error
	}

	// Parse choice number
	choice := 0
	if len(input) == 1 && input[0] >= '1' && input[0] <= '9' {
		choice = int(input[0] - '0')
	} else {
		result.Error = ErrInvalidChoice
		result.Valid = false
		return result, result.Error
	}

	// Validate choice range
	if choice < 1 || choice > maxChoices {
		result.Error = ErrInvalidChoice
		result.Valid = false
		return result, result.Error
	}

	result.Sanitized = input
	result.Valid = true
	return result, nil
}

// ProcessFreeTextInput processes free text input.
func (h *InputHandler) ProcessFreeTextInput(input string) (*PlayerInput, error) {
	result := NewPlayerInput(InputModeFreeText, input)

	// Validate and sanitize
	sanitized, err := h.validator.Sanitize(input)
	if err != nil {
		result.Error = err
		result.Valid = false
		return result, err
	}

	result.Sanitized = sanitized
	result.Valid = true
	return result, nil
}

// ProcessCommandInput processes command input.
func (h *InputHandler) ProcessCommandInput(input string) (*PlayerInput, error) {
	result := NewPlayerInput(InputModeCommand, input)

	// Check if it starts with /
	if len(input) == 0 || input[0] != '/' {
		result.Error = ErrNotCommand
		result.Valid = false
		return result, result.Error
	}

	// Check if command is empty
	if len(input) == 1 {
		result.Error = ErrInvalidCommand
		result.Valid = false
		return result, result.Error
	}

	// Simple sanitization for commands
	result.Sanitized = input
	result.Valid = true
	return result, nil
}

// DetectInputMode detects the input mode based on the input string.
func (h *InputHandler) DetectInputMode(input string) InputMode {
	// Command mode
	if len(input) > 0 && input[0] == '/' {
		return InputModeCommand
	}

	// Choice mode (single digit 1-9)
	if len(input) == 1 && input[0] >= '1' && input[0] <= '9' {
		return InputModeChoice
	}

	// Free text mode
	return InputModeFreeText
}

// GetBuffer returns the current input buffer.
func (h *InputHandler) GetBuffer() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.buffer
}

// AppendToBuffer appends text to the input buffer.
func (h *InputHandler) AppendToBuffer(text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buffer += text
}

// ClearBuffer clears the input buffer.
func (h *InputHandler) ClearBuffer() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buffer = ""
}
