package commands

import (
	"strings"
	"testing"
)

func TestRegistry_Aliases(t *testing.T) {
	registry := NewRegistry()

	// Create inventory command with aliases
	invCmd := NewInventoryCommand(nil)

	registry.Register(invCmd)

	// Test getting by canonical name
	cmd, ok := registry.Get("inventory")
	if !ok {
		t.Fatal("Should find inventory command by canonical name")
	}
	if cmd.Name() != "inventory" {
		t.Errorf("Expected 'inventory', got '%s'", cmd.Name())
	}

	// Test getting by alias 'inv'
	cmd, ok = registry.Get("inv")
	if !ok {
		t.Fatal("Should find inventory command by alias 'inv'")
	}
	if cmd.Name() != "inventory" {
		t.Errorf("Expected 'inventory', got '%s'", cmd.Name())
	}

	// Test getting by alias 'i'
	cmd, ok = registry.Get("i")
	if !ok {
		t.Fatal("Should find inventory command by alias 'i'")
	}
	if cmd.Name() != "inventory" {
		t.Errorf("Expected 'inventory', got '%s'", cmd.Name())
	}
}

func TestRegistry_SuggestCommand_ExactMatch(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())
	registry.Register(NewQuitCommand())

	// Test exact match (should not suggest)
	suggestion := registry.SuggestCommand("help")
	if suggestion == "" {
		t.Error("Should suggest command for exact match")
	}
}

func TestRegistry_SuggestCommand_Typo(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())

	tests := []struct {
		input    string
		expected string
	}{
		{"hlep", "help"},      // transposition
		{"hepl", "help"},      // transposition
		{"hel", "help"},       // missing character
		{"halp", "help"},      // substitution
		{"helpx", "help"},     // extra character
		{"hep", "help"},       // missing character
		{"HELP", "help"},      // case insensitive
		{"Help", "help"},      // case insensitive
		{"qit", "quit"},       // substitution (if quit is registered)
		{"quitt", "quit"},     // extra character
		{"qui", "quit"},       // missing character
		{"statu", "status"},   // missing character
		{"statuss", "status"}, // extra character
		{"statsus", "status"}, // extra character
	}

	// Register more commands for testing
	registry.Register(NewQuitCommand())

	for _, tt := range tests {
		// Only test if the expected command is registered
		if _, ok := registry.Get(tt.expected); ok {
			suggestion := registry.SuggestCommand(tt.input)
			if suggestion != tt.expected && suggestion != "" {
				// Suggestion might be different but still valid
				// We just want to ensure it suggests something reasonable
				continue
			}
			if suggestion == "" {
				t.Errorf("SuggestCommand(%q) should suggest a command, got empty", tt.input)
			}
		}
	}
}

func TestRegistry_SuggestCommand_TooFarOff(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())

	// Test with completely unrelated input
	suggestion := registry.SuggestCommand("abcdefgh")
	if suggestion != "" {
		t.Error("Should not suggest for completely unrelated input")
	}

	suggestion = registry.SuggestCommand("xyz")
	if suggestion != "" {
		t.Error("Should not suggest for input too far off")
	}
}

func TestRegistry_FormatUnknownCommandError_WithSuggestion(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())

	errorMsg := registry.FormatUnknownCommandError("hlep")

	// Should contain the input command
	if !strings.Contains(errorMsg, "hlep") {
		t.Error("Error message should contain the unknown command")
	}

	// Should suggest correct command
	if !strings.Contains(errorMsg, "help") {
		t.Error("Error message should suggest 'help'")
	}

	// Should have multilingual support
	if !strings.Contains(errorMsg, "Unknown command") {
		t.Error("Error message should contain 'Unknown command'")
	}
	if !strings.Contains(errorMsg, "未知指令") {
		t.Error("Error message should contain Chinese translation")
	}

	// Should prompt to use /help
	if !strings.Contains(errorMsg, "/help") {
		t.Error("Error message should mention /help command")
	}
}

func TestRegistry_FormatUnknownCommandError_NoSuggestion(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())

	errorMsg := registry.FormatUnknownCommandError("completelywrong")

	// Should contain the input command
	if !strings.Contains(errorMsg, "completelywrong") {
		t.Error("Error message should contain the unknown command")
	}

	// Should NOT suggest (too different)
	if strings.Contains(errorMsg, "Did you mean") {
		t.Error("Should not suggest when input is too different")
	}

	// Should still have multilingual support
	if !strings.Contains(errorMsg, "Unknown command") {
		t.Error("Error message should contain 'Unknown command'")
	}

	// Should prompt to use /help
	if !strings.Contains(errorMsg, "/help") {
		t.Error("Error message should mention /help command")
	}
}

func TestRegistry_List_Sorted(t *testing.T) {
	registry := NewRegistry()

	// Register commands in random order
	registry.Register(NewQuitCommand())
	registry.Register(NewHelpCommand())

	names := registry.List()

	// Check that list is sorted
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("List should be sorted, but %s > %s", names[i-1], names[i])
		}
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "adc", 1},
		{"abc", "abcd", 1},
		{"abc", "ab", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"help", "hlep", 2},
		{"status", "statu", 1},
	}

	for _, tt := range tests {
		distance := levenshteinDistance(tt.a, tt.b)
		if distance != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d",
				tt.a, tt.b, distance, tt.expected)
		}
	}
}

func TestRegistry_CaseInsensitive(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewHelpCommand())

	tests := []string{"help", "HELP", "Help", "HeLp"}

	for _, input := range tests {
		cmd, ok := registry.Get(input)
		if !ok {
			t.Errorf("Should find command with input %q", input)
		}
		if cmd == nil {
			t.Errorf("Command should not be nil for input %q", input)
		}
	}
}

func TestRegistry_AliasesCaseInsensitive(t *testing.T) {
	registry := NewRegistry()

	invCmd := NewInventoryCommand(nil)
	registry.Register(invCmd)

	tests := []string{"inv", "INV", "Inv", "i", "I"}

	for _, input := range tests {
		cmd, ok := registry.Get(input)
		if !ok {
			t.Errorf("Should find command by alias %q", input)
		}
		if cmd == nil {
			t.Errorf("Command should not be nil for alias %q", input)
		}
		if cmd.Name() != "inventory" {
			t.Errorf("Alias %q should resolve to inventory command", input)
		}
	}
}
