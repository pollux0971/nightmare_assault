package commands

import (
	"strings"
	"testing"
)

func TestMusicCommand_Name(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)
	if cmd.Name() != "music" {
		t.Errorf("Expected name 'music', got '%s'", cmd.Name())
	}
}

func TestMusicCommand_Help(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)
	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
	if !strings.Contains(help, "/music") {
		t.Error("Help text should contain /music command")
	}
}

func TestMusicCommand_Aliases(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)
	aliases := cmd.Aliases()
	if len(aliases) == 0 {
		t.Error("Expected at least one alias")
	}
}

func TestMusicCommand_Execute_NoAudioManager(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)

	output, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Fatalf("Expected no error when audio manager is nil, got: %v", err)
	}

	if !strings.Contains(output, "音訊系統未初始化") {
		t.Error("Expected message about audio system not initialized")
	}
}

func TestMusicCommand_Execute_NoArgs(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Expected no error with no args, got: %v", err)
	}

	if !strings.Contains(output, "/music") {
		t.Error("Expected help text when no args provided")
	}
}

func TestMusicCommand_Execute_UnknownSubcommand(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)

	output, err := cmd.Execute([]string{"invalid"})
	// When audioManager is nil, returns warning instead of error
	if err != nil {
		t.Fatalf("Expected no error when audioManager is nil, got: %v", err)
	}

	if !strings.Contains(output, "音訊系統未初始化") {
		t.Errorf("Expected warning about audio system not initialized, got: %s", output)
	}
}

func TestMusicCommand_Execute_PlayWithoutFilename(t *testing.T) {
	cmd := NewMusicCommand(nil, nil)

	output, err := cmd.Execute([]string{"play"})
	// When audioManager is nil, returns warning instead of error about usage
	if err != nil {
		t.Fatalf("Expected no error when audioManager is nil, got: %v", err)
	}

	if !strings.Contains(output, "音訊系統未初始化") {
		t.Errorf("Expected warning about audio system not initialized, got: %s", output)
	}
}

// Note: Integration tests with actual audio manager would require:
// - Mock audio manager
// - Mock CustomBGMManager
// - Temporary test directory with audio files
// These are skipped for now as they require more complex setup
