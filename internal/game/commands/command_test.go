package commands

import (
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()

	if reg == nil {
		t.Fatal("NewRegistry should not return nil")
	}
	if len(reg.commands) != 0 {
		t.Error("New registry should be empty")
	}
}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()
	cmd := NewHelpCommand()

	reg.Register(cmd)

	if len(reg.commands) != 1 {
		t.Errorf("Registry should have 1 command, got %d", len(reg.commands))
	}
}

func TestRegistry_Get(t *testing.T) {
	reg := NewRegistry()
	help := NewHelpCommand()
	reg.Register(help)

	// Test successful get
	cmd, ok := reg.Get("help")
	if !ok {
		t.Error("Should find registered command")
	}
	if cmd.Name() != "help" {
		t.Errorf("Command name = %q, want 'help'", cmd.Name())
	}

	// Test not found
	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("Should not find nonexistent command")
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewHelpCommand())
	reg.Register(NewQuitCommand())

	names := reg.List()

	if len(names) != 2 {
		t.Errorf("List returned %d commands, want 2", len(names))
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedArgs []string
	}{
		{"/help", "help", nil},
		{"/status", "status", nil},
		{"/save game1", "save", []string{"game1"}},
		{"/unknown arg1 arg2", "unknown", []string{"arg1", "arg2"}},
		{"help", "help", nil},
		{"  /help  ", "help", nil},
		{"/", "", nil},
		{"", "", nil},
	}

	for _, tt := range tests {
		name, args := Parse(tt.input)

		if name != tt.expectedName {
			t.Errorf("Parse(%q): name = %q, want %q", tt.input, name, tt.expectedName)
		}

		if len(args) != len(tt.expectedArgs) {
			t.Errorf("Parse(%q): args length = %d, want %d", tt.input, len(args), len(tt.expectedArgs))
			continue
		}

		for i, arg := range args {
			if arg != tt.expectedArgs[i] {
				t.Errorf("Parse(%q): args[%d] = %q, want %q", tt.input, i, arg, tt.expectedArgs[i])
			}
		}
	}
}

func TestHelpCommand(t *testing.T) {
	cmd := NewHelpCommand()

	if cmd.Name() != "help" {
		t.Errorf("Name = %q, want 'help'", cmd.Name())
	}

	output, err := cmd.Execute(nil)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("Help output should not be empty")
	}
}

func TestStatusCommand(t *testing.T) {
	stats := &game.PlayerStats{HP: 80, SAN: 70, State: game.SanityAnxious}
	cmd := NewStatusCommand(stats, 5)

	if cmd.Name() != "status" {
		t.Errorf("Name = %q, want 'status'", cmd.Name())
	}

	output, err := cmd.Execute(nil)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("Status output should not be empty")
	}
}

func TestQuitCommand(t *testing.T) {
	cmd := NewQuitCommand()

	if cmd.Name() != "quit" {
		t.Errorf("Name = %q, want 'quit'", cmd.Name())
	}

	output, err := cmd.Execute(nil)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	if output != "QUIT_REQUESTED" {
		t.Errorf("Output = %q, want 'QUIT_REQUESTED'", output)
	}
}

func TestMakeBar(t *testing.T) {
	tests := []struct {
		value    int
		max      int
		width    int
		minLen   int
	}{
		{100, 100, 20, 20},
		{50, 100, 20, 20},
		{0, 100, 20, 20},
		{75, 100, 10, 10},
	}

	for _, tt := range tests {
		bar := makeBar(tt.value, tt.max, tt.width)
		if len([]rune(bar)) < tt.minLen {
			t.Errorf("makeBar(%d, %d, %d) length = %d, want >= %d",
				tt.value, tt.max, tt.width, len([]rune(bar)), tt.minLen)
		}
	}
}
