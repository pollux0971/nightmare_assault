package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestBGMCommand_Name(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	if cmd.Name() != "bgm" {
		t.Errorf("Name() = %s, expected 'bgm'", cmd.Name())
	}
}

func TestBGMCommand_Help(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	help := cmd.Help()

	if !strings.Contains(help, "/bgm on") {
		t.Error("Help should contain '/bgm on'")
	}
	if !strings.Contains(help, "/bgm off") {
		t.Error("Help should contain '/bgm off'")
	}
	if !strings.Contains(help, "/bgm volume") {
		t.Error("Help should contain '/bgm volume'")
	}
	if !strings.Contains(help, "/bgm list") {
		t.Error("Help should contain '/bgm list'")
	}
}

func TestBGMCommand_Execute_NoArgs(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	output, err := cmd.Execute([]string{})

	if err != nil {
		t.Errorf("Execute with no args should not return error, got: %v", err)
	}

	if !strings.Contains(output, "/bgm on") {
		t.Error("Output should contain help text")
	}
}

func TestBGMCommand_Execute_UnknownSubcommand(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	_, err := cmd.Execute([]string{"unknown"})

	if err == nil {
		t.Error("Execute with unknown subcommand should return error")
	}

	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("Error message should mention unknown subcommand, got: %v", err)
	}
}

func TestBGMCommand_EnableBGM_NilManager(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	output, err := cmd.Execute([]string{"on"})

	if err != nil {
		t.Errorf("Enable with nil manager should not error, got: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Output should mention audio system not initialized")
	}
}

func TestBGMCommand_DisableBGM_NilManager(t *testing.T) {
	cmd := NewBGMCommand(nil, nil)
	output, err := cmd.Execute([]string{"off"})

	if err != nil {
		t.Errorf("Disable with nil manager should not error, got: %v", err)
	}

	if !strings.Contains(output, "未初始化") {
		t.Error("Output should mention audio system not initialized")
	}
}

func TestBGMCommand_SetVolume_InvalidInput(t *testing.T) {
	// Create a valid audio manager to test volume validation
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)
	cmd := NewBGMCommand(audioManager, cfg)

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

func TestBGMCommand_SetVolume_WithMockManager(t *testing.T) {
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
	cmd := NewBGMCommand(audioManager, loadedCfg)

	// Set volume to 50 (BGMPlayer will be nil since not initialized, but command should handle gracefully)
	output, err := cmd.Execute([]string{"volume", "50"})
	if err != nil {
		t.Errorf("SetVolume failed: %v", err)
	}

	// Since audio system isn't initialized, expect "未初始化" message
	if !strings.Contains(output, "未初始化") && !strings.Contains(output, "50%") {
		t.Errorf("Output should mention either '未初始化' or '50%%', got: %s", output)
	}
}

func TestBGMCommand_List(t *testing.T) {
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)
	cmd := NewBGMCommand(audioManager, cfg)

	output, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("List should not return error, got: %v", err)
	}

	// Check that all BGM tracks are listed
	tracks := []string{
		"ambient_exploration",
		"tension_chase",
		"safe_rest",
		"horror_reveal",
		"mystery_puzzle",
		"ending_death",
	}

	for _, track := range tracks {
		if !strings.Contains(output, track) {
			t.Errorf("List output should contain %s", track)
		}
	}

	// Check status section
	if !strings.Contains(output, "狀態") {
		t.Error("List output should contain status section")
	}
}

func TestBGMCommand_EnableDisable_Lifecycle(t *testing.T) {
	// Create temp config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.Audio.BGMEnabled = false
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	loadedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	audioManager := audio.NewAudioManager(loadedCfg.Audio)
	cmd := NewBGMCommand(audioManager, loadedCfg)

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

// TestBGMCommand_Integration tests command registration
func TestBGMCommand_Integration(t *testing.T) {
	registry := NewRegistry()
	cfg := config.DefaultConfig()
	audioManager := audio.NewAudioManager(cfg.Audio)

	bgmCmd := NewBGMCommand(audioManager, cfg)
	registry.Register(bgmCmd)

	// Test retrieval
	cmd, ok := registry.Get("bgm")
	if !ok {
		t.Fatal("BGM command should be registered")
	}

	if cmd.Name() != "bgm" {
		t.Errorf("Retrieved command name = %s, expected 'bgm'", cmd.Name())
	}

	// Test execution via registry
	output, err := cmd.Execute([]string{"list"})
	if err != nil {
		t.Errorf("Command execution via registry failed: %v", err)
	}

	if !strings.Contains(output, "ambient_exploration") {
		t.Error("Command should list BGM tracks")
	}
}
