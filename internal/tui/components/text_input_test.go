package components

import (
	"testing"
)

func TestNewFreeTextInput(t *testing.T) {
	input := NewFreeTextInput()

	if input == nil {
		t.Fatal("NewFreeTextInput should not return nil")
	}
	if input.maxLength != 200 {
		t.Errorf("MaxLength = %d, want 200", input.maxLength)
	}
}

func TestFreeTextInput_GetValue(t *testing.T) {
	input := NewFreeTextInput()

	if input.GetValue() != "" {
		t.Errorf("Initial value should be empty, got %q", input.GetValue())
	}
}

func TestFreeTextInput_SetValue(t *testing.T) {
	input := NewFreeTextInput()

	input.SetValue("test value")

	if input.GetValue() != "test value" {
		t.Errorf("GetValue = %q, want 'test value'", input.GetValue())
	}
}

func TestFreeTextInput_Clear(t *testing.T) {
	input := NewFreeTextInput()

	input.SetValue("test")
	input.Clear()

	if input.GetValue() != "" {
		t.Errorf("After Clear, value = %q, want empty", input.GetValue())
	}
}

func TestFreeTextInput_CharCount(t *testing.T) {
	input := NewFreeTextInput()

	if input.CharCount() != 0 {
		t.Errorf("Initial CharCount = %d, want 0", input.CharCount())
	}

	input.SetValue("hello")
	if input.CharCount() != 5 {
		t.Errorf("CharCount = %d, want 5", input.CharCount())
	}

	input.SetValue("你好世界")
	if input.CharCount() != 4 {
		t.Errorf("Chinese CharCount = %d, want 4", input.CharCount())
	}
}

func TestFreeTextInput_IsEmpty(t *testing.T) {
	input := NewFreeTextInput()

	if !input.IsEmpty() {
		t.Error("Initial state should be empty")
	}

	input.SetValue("test")
	if input.IsEmpty() {
		t.Error("After SetValue, should not be empty")
	}

	input.Clear()
	if !input.IsEmpty() {
		t.Error("After Clear, should be empty")
	}
}

func TestFreeTextInput_IsFull(t *testing.T) {
	input := NewFreeTextInput()

	if input.IsFull() {
		t.Error("Initial state should not be full")
	}

	// Set to max length (create string with actual characters)
	maxStr := ""
	for i := 0; i < 200; i++ {
		maxStr += "a"
	}
	input.SetValue(maxStr)
	if !input.IsFull() {
		t.Errorf("At max length, should be full (count=%d)", input.CharCount())
	}

	// Over max length
	overMaxStr := maxStr + "b"
	input.SetValue(overMaxStr)
	if !input.IsFull() {
		t.Errorf("Over max length, should be full (count=%d)", input.CharCount())
	}
}

func TestFreeTextInput_RemainingChars(t *testing.T) {
	input := NewFreeTextInput()

	if input.RemainingChars() != 200 {
		t.Errorf("Initial RemainingChars = %d, want 200", input.RemainingChars())
	}

	input.SetValue("hello")
	if input.RemainingChars() != 195 {
		t.Errorf("RemainingChars = %d, want 195", input.RemainingChars())
	}

	// Create string with 200 actual characters
	maxStr := ""
	for i := 0; i < 200; i++ {
		maxStr += "a"
	}
	input.SetValue(maxStr)
	if input.RemainingChars() != 0 {
		t.Errorf("At max, RemainingChars = %d, want 0", input.RemainingChars())
	}

	// Over max should return 0, not negative (will be truncated to 200)
	overMaxStr := maxStr + "extra"
	input.SetValue(overMaxStr)
	if input.RemainingChars() != 0 {
		t.Errorf("Over max, RemainingChars = %d, want 0", input.RemainingChars())
	}
}

func TestFreeTextInput_SetFocused(t *testing.T) {
	input := NewFreeTextInput()

	if input.IsFocused() {
		t.Error("Initial state should not be focused")
	}

	input.SetFocused(true)
	if !input.IsFocused() {
		t.Error("After SetFocused(true), should be focused")
	}

	input.SetFocused(false)
	if input.IsFocused() {
		t.Error("After SetFocused(false), should not be focused")
	}
}

func TestFreeTextInput_SetPlaceholder(t *testing.T) {
	input := NewFreeTextInput()

	input.SetPlaceholder("Enter text...")

	if input.placeholder != "Enter text..." {
		t.Errorf("Placeholder = %q, want 'Enter text...'", input.placeholder)
	}
}

func TestFreeTextInput_Reset(t *testing.T) {
	input := NewFreeTextInput()

	input.SetValue("test")
	input.SetFocused(true)

	input.Reset()

	if input.GetValue() != "" {
		t.Error("After Reset, value should be empty")
	}
	if input.IsFocused() {
		t.Error("After Reset, should not be focused")
	}
}
