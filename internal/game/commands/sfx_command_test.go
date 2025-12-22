package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestSFXCommand_Name(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	if cmd.Name() != "sfx" {
		t.Errorf("Name() = %s, expected 'sfx'", cmd.Name())
	}
}

func TestSFXCommand_Help(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	help := cmd.Help()

	if !strings.Contains(help, "/sfx on") {
		t.Error("Help should contain '/sfx on'")
	}
	if !strings.Contains(help, "/sfx off") {
		t.Error("Help should contain '/sfx off'")
	}
	if !strings.Contains(help, "/sfx volume") {
		t.Error("Help should contain '/sfx volume'")
	}
	if !strings.Contains(help, "/sfx list") {
		t.Error("Help should contain '/sfx list'")
	}
}

func TestSFXCommand_Execute_NoArgs(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute with no args should not return error, got: %v", err)
	}

	if !strings.Contains(output, "/sfx on") {
		t.Error("Output should contain help text")
	}
}

func TestSFXCommand_Execute_UnknownSubcommand(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	_, err := cmd.Execute([]string{"unknown"})

	if err == nil {
		t.Error("Execute with unknown subcommand should return error")
	}

	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("Error message should mention unknown subcommand, got: %v", err)
	}
}

func TestSFXCommand_EnableSFX_NilManager(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	output, err := cmd.Execute([]string{"on"})

	if err != nil {
		t.Errorf("Enable with nil manager should not error, got: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Output should mention audio system not initialized")
	}
}

func TestSFXCommand_DisableSFX_NilManager(t *testing.T) {
	cmd := NewSFXCommand(nil, nil)
	output, err := cmd.Execute([]string{"off"})

	if err != nil {
		t.Errorf("Disable with nil manager should not error, got: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Output should mention audio system not initialized")
	}
}

func TestSFXCommand_SetVolume_InvalidInput(t *testing.T) {
	// Create a valid audio manager to test volume validation
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)
	cmd := NewSFXCommand(audioManager, cfg)

	tests := []struct {
		name string
		args []string
	}{
		{"No volume arg", []string{"volume"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cmd.Execute(tt.args)
			if err == nil {
				t.Error("Execute should return error for missing volume arg")
			}
		})
	}
}

func TestSFXCommand_SetVolume_WithManager(t *testing.T) {
	// Create temp config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	loadedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	audioManager := audio.NewAudioManager(loadedCfg.Audio)
	cmd := NewSFXCommand(audioManager, loadedCfg)

	tests := []struct {
		name        string
		volumeArg   string
		shouldError bool
	}{
		{"Valid volume 50", "50", false},
		{"Valid volume 0", "0", false},
		{"Valid volume 100", "100", false},
		{"Invalid volume -10", "-10", true},
		{"Invalid volume 150", "150", true},
		{"Invalid volume abc", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := cmd.Execute([]string{"volume", tt.volumeArg})

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}

			// Since audio system isn't initialized, expect "未初始化" message or success
			if !tt.shouldError && !strings.Contains(output, "未初始化") && !strings.Contains(output, "已設定") && !strings.Contains(output, "設定為") {
				t.Errorf("Output should mention either '未初始化' or '設定', got: %s", output)
			}
		})
	}
}

func TestSFXCommand_List(t *testing.T) {
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)
	cmd := NewSFXCommand(audioManager, cfg)

	output, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("List should not return error, got: %v", err)
	}

	// Check that SFX list contains expected sound effects
	expectedSFX := []string{
		"door_open.wav",
		"door_close.wav",
		"footsteps.wav",
		"glass_break.wav",
		"thunder.wav",
		"whisper.wav",
		"warning.wav",
		"death.wav",
		"heartbeat.wav",
	}

	for _, sfx := range expectedSFX {
		if !strings.Contains(output, sfx) {
			t.Errorf("List output should contain %s", sfx)
		}
	}

	// Check status section
	if !strings.Contains(output, "狀態") {
		t.Error("List output should contain status section")
	}
}

func TestSFXCommand_EnableDisable_Lifecycle(t *testing.T) {
	// Create temp config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.Audio.SFXEnabled = false
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	loadedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	audioManager := audio.NewAudioManager(loadedCfg.Audio)
	cmd := NewSFXCommand(audioManager, loadedCfg)

	// Test enable (audio not initialized, so expect "未初始化" or "啟用")
	output, err := cmd.Execute([]string{"on"})
	if err != nil {
		t.Errorf("Enable failed: %v", err)
	}
	if !strings.Contains(output, "啟用") && !strings.Contains(output, "未初始化") {
		t.Errorf("Enable output should mention '啟用' or '未初始化', got: %s", output)
	}

	// Test disable (audio not initialized, so expect "未初始化" or "停用")
	output, err = cmd.Execute([]string{"off"})
	if err != nil {
		t.Errorf("Disable failed: %v", err)
	}
	if !strings.Contains(output, "停用") && !strings.Contains(output, "未初始化") {
		t.Errorf("Disable output should mention '停用' or '未初始化', got: %s", output)
	}
}

// TestSFXCommand_Integration tests command registration
func TestSFXCommand_Integration(t *testing.T) {
	registry := NewRegistry()
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)

	sfxCmd := NewSFXCommand(audioManager, cfg)
	registry.Register(sfxCmd)

	// Test retrieval
	cmd, ok := registry.Get("sfx")
	if !ok {
		t.Fatal("SFX command should be registered")
	}

	if cmd.Name() != "sfx" {
		t.Errorf("Retrieved command name = %s, expected 'sfx'", cmd.Name())
	}

	// Test execution via registry
	output, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("Command execution via registry failed: %v", err)
	}

	if !strings.Contains(output, "warning.wav") {
		t.Error("Command should list SFX effects")
	}
}
