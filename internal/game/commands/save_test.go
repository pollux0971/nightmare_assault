package commands

import (
	"testing"
)

func TestParseSaveCommand(t *testing.T) {
	tests := []struct {
		input       string
		expectedCmd string
		expectedArg int
		hasError    bool
	}{
		{"/save", "save", 0, false},
		{"/save 1", "save", 1, false},
		{"/save 2", "save", 2, false},
		{"/save 3", "save", 3, false},
		{"/save 4", "save", 0, true},     // Invalid slot
		{"/save abc", "save", 0, true},   // Invalid argument
		{"/save -1", "save", 0, true},    // Negative slot
		{"/load", "load", 0, false},
		{"/load 1", "load", 1, false},
		{"/load 2", "load", 2, false},
		{"/load abc", "load", 0, true},   // Invalid argument
	}

	for _, tt := range tests {
		cmd, arg, err := ParseSlotCommand(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseSlotCommand(%q) expected error, got none", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSlotCommand(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if cmd != tt.expectedCmd {
			t.Errorf("ParseSlotCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.expectedCmd)
		}
		if arg != tt.expectedArg {
			t.Errorf("ParseSlotCommand(%q) arg = %d, want %d", tt.input, arg, tt.expectedArg)
		}
	}
}

func TestIsSaveCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/save", true},
		{"/save 1", true},
		{"/SAVE", true},
		{"/Save", true},
		{"save", false},
		{"/load", false},
		{"/savegame", false},
	}

	for _, tt := range tests {
		result := IsSaveCommand(tt.input)
		if result != tt.expected {
			t.Errorf("IsSaveCommand(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsLoadCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/load", true},
		{"/load 1", true},
		{"/LOAD", true},
		{"/Load", true},
		{"load", false},
		{"/save", false},
		{"/loadgame", false},
	}

	for _, tt := range tests {
		result := IsLoadCommand(tt.input)
		if result != tt.expected {
			t.Errorf("IsLoadCommand(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestSaveCommandHelp(t *testing.T) {
	help := SaveCommandHelp()
	if help == "" {
		t.Error("SaveCommandHelp should not be empty")
	}
}

func TestLoadCommandHelp(t *testing.T) {
	help := LoadCommandHelp()
	if help == "" {
		t.Error("LoadCommandHelp should not be empty")
	}
}

func TestGetAllSlotCommands(t *testing.T) {
	cmds := GetAllSlotCommands()
	if len(cmds) < 2 {
		t.Error("Should have at least save and load commands")
	}

	foundSave := false
	foundLoad := false
	for _, cmd := range cmds {
		if cmd.Name == "save" {
			foundSave = true
		}
		if cmd.Name == "load" {
			foundLoad = true
		}
	}

	if !foundSave {
		t.Error("Missing save command")
	}
	if !foundLoad {
		t.Error("Missing load command")
	}
}

func TestSlotCommandInfo(t *testing.T) {
	cmds := GetAllSlotCommands()
	for _, cmd := range cmds {
		if cmd.Name == "" {
			t.Error("Command name should not be empty")
		}
		if cmd.Description == "" {
			t.Errorf("Command %s should have description", cmd.Name)
		}
		if cmd.Usage == "" {
			t.Errorf("Command %s should have usage", cmd.Name)
		}
	}
}
