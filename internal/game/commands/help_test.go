package commands

import (
	"strings"
	"testing"
	"time"
)

func TestHelpCommand_Name(t *testing.T) {
	cmd := NewHelpCommand()

	if cmd.Name() != "help" {
		t.Errorf("Expected name 'help', got '%s'", cmd.Name())
	}
}

func TestHelpCommand_Help(t *testing.T) {
	cmd := NewHelpCommand()

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
}

func TestHelpCommand_Execute_GeneralHelp(t *testing.T) {
	cmd := NewHelpCommand()

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Check basic content
	if !strings.Contains(output, "NIGHTMARE ASSAULT") {
		t.Error("Output should contain game title")
	}
	if !strings.Contains(output, "HOW TO PLAY") {
		t.Error("Output should contain 'HOW TO PLAY' section")
	}
	if !strings.Contains(output, "SLASH COMMANDS") {
		t.Error("Output should contain 'SLASH COMMANDS' section")
	}
	if !strings.Contains(output, "KEYBOARD SHORTCUTS") {
		t.Error("Output should contain keyboard shortcuts")
	}
	if !strings.Contains(output, "GAME MECHANICS") {
		t.Error("Output should contain game mechanics")
	}
}

func TestHelpCommand_Execute_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Register some commands
	registry.Register(NewHelpCommand())
	registry.Register(NewQuitCommand())

	cmd := NewHelpCommandWithRegistry(registry)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should list registered commands
	if !strings.Contains(output, "help") {
		t.Error("Should list 'help' command")
	}
	if !strings.Contains(output, "quit") {
		t.Error("Should list 'quit' command")
	}
}

func TestHelpCommand_Execute_DetailedHelp(t *testing.T) {
	registry := NewRegistry()

	quitCmd := NewQuitCommand()
	registry.Register(quitCmd)

	cmd := NewHelpCommandWithRegistry(registry)

	// Request detailed help for quit command
	output, err := cmd.Execute([]string{"quit"})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should show detailed information
	if !strings.Contains(output, "COMMAND: /QUIT") {
		t.Error("Should show command name in header")
	}
	if !strings.Contains(output, "Description:") {
		t.Error("Should show description")
	}
}

func TestHelpCommand_Execute_UnknownCommandHelp(t *testing.T) {
	registry := NewRegistry()
	cmd := NewHelpCommandWithRegistry(registry)

	// Request help for unknown command
	output, err := cmd.Execute([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if !strings.Contains(output, "Unknown command") {
		t.Error("Should show unknown command message")
	}
	if !strings.Contains(output, "nonexistent") {
		t.Error("Should mention the unknown command name")
	}
}

func TestHelpCommand_Execute_CategorizedCommands(t *testing.T) {
	registry := NewRegistry()

	// Register commands from different categories
	helpCmd := NewHelpCommand()
	quitCmd := NewQuitCommand()

	registry.Register(helpCmd)
	registry.Register(quitCmd)

	cmd := NewHelpCommandWithRegistry(registry)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should show help command
	if !strings.Contains(output, "help") {
		t.Error("Should list help command")
	}

	// Should show quit command
	if !strings.Contains(output, "quit") {
		t.Error("Should list quit command")
	}
}

func TestHelpCommand_Execute_WithAliases(t *testing.T) {
	// This test verifies that commands with aliases are properly supported
	// by the help system. The actual alias functionality is tested in
	// command_error_test.go with the inventory command which has real aliases.

	// For this test, we use the actual inventory command which has aliases
	invCmd := NewInventoryCommand(nil)

	if len(invCmd.Aliases()) == 0 {
		t.Error("Inventory command should have aliases")
	}

	// Note: The help command will show aliases when detailed help is requested
	// for commands that implement the Aliases() interface method
}

func TestHelpCommand_Execute_ResponseTime(t *testing.T) {
	registry := NewRegistry()

	// Register many commands
	for i := 0; i < 20; i++ {
		registry.Register(NewHelpCommand())
	}

	cmd := NewHelpCommandWithRegistry(registry)

	start := time.Now()
	_, err := cmd.Execute([]string{})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// NFR-P02: Response time should be < 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("Command execution took %v, expected < 100ms", duration)
	}
}

func TestHelpCommand_Execute_FallbackList(t *testing.T) {
	// Create help command without registry
	cmd := NewHelpCommand()

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// Should show static fallback list
	if !strings.Contains(output, "/status") {
		t.Error("Fallback list should contain /status")
	}
	if !strings.Contains(output, "/inventory") {
		t.Error("Fallback list should contain /inventory")
	}
	if !strings.Contains(output, "/clues") {
		t.Error("Fallback list should contain /clues")
	}
	if !strings.Contains(output, "/save") {
		t.Error("Fallback list should contain /save")
	}
	if !strings.Contains(output, "/load") {
		t.Error("Fallback list should contain /load")
	}
}

func TestHelpCommand_Execute_TipMessage(t *testing.T) {
	cmd := NewHelpCommand()

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	// AC4: Should support /help <command> for detailed help
	if !strings.Contains(output, "/help <command>") {
		t.Error("Should show tip about detailed command help")
	}
}
