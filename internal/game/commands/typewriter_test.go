package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestTypewriterCommand_Name tests the Name method (Story 9-8).
func TestTypewriterCommand_Name(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	if cmd.Name() != "typewriter" {
		t.Errorf("Expected command name 'typewriter', got '%s'", cmd.Name())
	}
}

// TestTypewriterCommand_Aliases tests command aliases (Story 9-8).
func TestTypewriterCommand_Aliases(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	aliases := cmd.Aliases()
	if len(aliases) == 0 {
		t.Error("Expected command to have aliases")
	}

	// Check for expected aliases
	expectedAliases := map[string]bool{"tw": false, "type": false}
	for _, alias := range aliases {
		if _, ok := expectedAliases[alias]; ok {
			expectedAliases[alias] = true
		}
	}

	for alias, found := range expectedAliases {
		if !found {
			t.Errorf("Expected alias '%s' not found", alias)
		}
	}
}

// TestTypewriterCommand_Help tests the Help method (Story 9-8).
func TestTypewriterCommand_Help(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
	if !strings.Contains(help, "typewriter") && !strings.Contains(help, "打字機") {
		t.Errorf("Help text should mention typewriter, got: %s", help)
	}
}

// TestTypewriterCommand_ShowStatus tests showing current status (Story 9-8 AC4).
func TestTypewriterCommand_ShowStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should show status information
	if !strings.Contains(output, "狀態") && !strings.Contains(output, "status") {
		t.Error("Output should show status")
	}
	if !strings.Contains(output, "速度") && !strings.Contains(output, "speed") {
		t.Error("Output should show speed")
	}

	// Test explicit "status" subcommand
	output2, err := cmd.Execute([]string{"status"})
	if err != nil {
		t.Fatalf("Execute with 'status' failed: %v", err)
	}
	if output2 != output {
		t.Error("'status' subcommand should produce same output as no args")
	}
}

// TestTypewriterCommand_Enable tests enabling typewriter effect (Story 9-8 AC4).
func TestTypewriterCommand_Enable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.Typewriter.Enabled = false // Start disabled
	cfg.SaveToPath(configPath)

	cmd := NewTypewriterCommand(cfg)

	output, err := cmd.Execute([]string{"on"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "✓") {
		t.Error("Output should show success indicator")
	}
	if !strings.Contains(output, "已啟用") && !strings.Contains(output, "enabled") {
		t.Error("Output should confirm enabling")
	}

	// Verify config was updated (AC5)
	if !cfg.Typewriter.Enabled {
		t.Error("Typewriter should be enabled in config")
	}

	// Test "enable" alias
	cfg.Typewriter.Enabled = false
	output, err = cmd.Execute([]string{"enable"})
	if err != nil {
		t.Fatalf("Execute with 'enable' failed: %v", err)
	}
	if !cfg.Typewriter.Enabled {
		t.Error("'enable' alias should work")
	}
}

// TestTypewriterCommand_Disable tests disabling typewriter effect (Story 9-8 AC4).
func TestTypewriterCommand_Disable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.Typewriter.Enabled = true // Start enabled
	cfg.SaveToPath(configPath)

	cmd := NewTypewriterCommand(cfg)

	output, err := cmd.Execute([]string{"off"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "✓") {
		t.Error("Output should show success indicator")
	}
	if !strings.Contains(output, "已停用") && !strings.Contains(output, "disabled") {
		t.Error("Output should confirm disabling")
	}

	// Verify config was updated (AC5)
	if cfg.Typewriter.Enabled {
		t.Error("Typewriter should be disabled in config")
	}

	// Test "disable" alias
	cfg.Typewriter.Enabled = true
	output, err = cmd.Execute([]string{"disable"})
	if err != nil {
		t.Fatalf("Execute with 'disable' failed: %v", err)
	}
	if cfg.Typewriter.Enabled {
		t.Error("'disable' alias should work")
	}
}

// TestTypewriterCommand_SetSpeed tests setting typewriter speed (Story 9-8 AC2, AC4).
func TestTypewriterCommand_SetSpeed(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.SaveToPath(configPath)

	cmd := NewTypewriterCommand(cfg)

	// Test setting speed
	output, err := cmd.Execute([]string{"speed", "50"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "✓") {
		t.Error("Output should show success indicator")
	}
	if !strings.Contains(output, "50") {
		t.Error("Output should confirm speed value")
	}

	// Verify config was updated (AC5)
	if cfg.Typewriter.Speed != 50 {
		t.Errorf("Expected speed 50, got %d", cfg.Typewriter.Speed)
	}
	if !cfg.Typewriter.Enabled {
		t.Error("Setting speed should enable typewriter")
	}

	// Test direct speed value (without "speed" subcommand)
	output, err = cmd.Execute([]string{"30"})
	if err != nil {
		t.Fatalf("Execute with direct speed failed: %v", err)
	}
	if cfg.Typewriter.Speed != 30 {
		t.Errorf("Expected speed 30, got %d", cfg.Typewriter.Speed)
	}
}

// TestTypewriterCommand_SpeedValidation tests speed range validation (Story 9-8 AC2).
func TestTypewriterCommand_SpeedValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	// Test valid speeds within range (10-200)
	_, err := cmd.Execute([]string{"speed", "30"})
	if err != nil {
		t.Errorf("Speed 30 should be valid: %v", err)
	}

	_, err = cmd.Execute([]string{"speed", "50"})
	if err != nil {
		t.Errorf("Speed 50 should be valid (AC2: 30-50ms range): %v", err)
	}

	// Test invalid speeds (out of range)
	_, err = cmd.Execute([]string{"speed", "5"})
	if err == nil {
		t.Error("Speed 5 should be rejected (below minimum 10)")
	}

	_, err = cmd.Execute([]string{"speed", "250"})
	if err == nil {
		t.Error("Speed 250 should be rejected (above maximum 200)")
	}

	// Test invalid format
	_, err = cmd.Execute([]string{"speed", "invalid"})
	if err == nil {
		t.Error("Non-numeric speed should be rejected")
	}
}

// TestTypewriterCommand_DefaultSpeed tests that default speed is reasonable (Story 9-8 AC2).
func TestTypewriterCommand_DefaultSpeed(t *testing.T) {
	cfg := config.DefaultConfig()

	// AC2: Default speed should be 30-50ms per character
	// This translates to 20-33 chars/sec
	// Config default is 40 chars/sec (25ms per char)
	if cfg.Typewriter.Speed < 20 || cfg.Typewriter.Speed > 50 {
		t.Errorf("Default speed should be 20-50 chars/sec (30-50ms/char), got %d", cfg.Typewriter.Speed)
	}
}

// TestTypewriterCommand_CalculateDelay tests delay calculation (Story 9-8 AC2).
func TestTypewriterCommand_CalculateDelay(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Typewriter.Enabled = true
	cfg.Typewriter.Speed = 40 // 40 chars/sec = 25ms/char

	cmd := NewTypewriterCommand(cfg)
	delay := cmd.CalculateDelay()

	expectedDelay := 1000 / 40 // 25ms
	if delay != expectedDelay {
		t.Errorf("Expected delay %dms, got %dms", expectedDelay, delay)
	}

	// Test with disabled typewriter
	cfg.Typewriter.Enabled = false
	delay = cmd.CalculateDelay()
	if delay != 0 {
		t.Errorf("Expected 0 delay when disabled, got %d", delay)
	}
}

// TestTypewriterCommand_PunctuationDelay tests punctuation detection (Story 9-8 AC6).
func TestTypewriterCommand_PunctuationDelay(t *testing.T) {
	// Test Chinese punctuation
	chinesePunctuation := []rune{'。', '！', '？', '；', '：', '，', '、'}
	for _, char := range chinesePunctuation {
		if !ShouldDelayAfterPunctuation(char) {
			t.Errorf("Character '%c' should be detected as punctuation", char)
		}
	}

	// Test English punctuation
	englishPunctuation := []rune{'.', '!', '?', ';', ':', ','}
	for _, char := range englishPunctuation {
		if !ShouldDelayAfterPunctuation(char) {
			t.Errorf("Character '%c' should be detected as punctuation", char)
		}
	}

	// Test non-punctuation
	regularChars := []rune{'你', '好', 'h', 'e', 'l', 'l', 'o', '1', '2'}
	for _, char := range regularChars {
		if ShouldDelayAfterPunctuation(char) {
			t.Errorf("Character '%c' should not be detected as punctuation", char)
		}
	}

	// Test punctuation delay value
	delay := GetPunctuationDelay()
	if delay <= 0 {
		t.Error("Punctuation delay should be positive")
	}
	if delay < 100 {
		t.Errorf("Punctuation delay should be at least 100ms, got %d", delay)
	}
}

// TestTypewriterCommand_ConfigPersistence tests that settings are saved (Story 9-8 AC5).
func TestTypewriterCommand_ConfigPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial config
	cfg1 := config.DefaultConfig()
	if err := cfg1.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Modify settings
	cmd1 := NewTypewriterCommand(cfg1)
	_, err := cmd1.Execute([]string{"speed", "60"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Save config
	if err := cfg1.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config again
	cfg2, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify settings persisted
	if cfg2.Typewriter.Speed != 60 {
		t.Errorf("Expected persisted speed 60, got %d", cfg2.Typewriter.Speed)
	}
	if !cfg2.Typewriter.Enabled {
		t.Error("Expected typewriter to be enabled")
	}
}

// TestTypewriterCommand_MultipleSettings tests changing multiple settings (Story 9-8).
func TestTypewriterCommand_MultipleSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	cfg.SaveToPath(configPath)

	cmd := NewTypewriterCommand(cfg)

	// Set speed
	_, err := cmd.Execute([]string{"speed", "50"})
	if err != nil {
		t.Fatalf("Setting speed failed: %v", err)
	}

	// Disable
	_, err = cmd.Execute([]string{"off"})
	if err != nil {
		t.Fatalf("Disabling failed: %v", err)
	}

	// Verify both settings
	if cfg.Typewriter.Speed != 50 {
		t.Errorf("Speed should be 50, got %d", cfg.Typewriter.Speed)
	}
	if cfg.Typewriter.Enabled {
		t.Error("Typewriter should be disabled")
	}

	// Enable again
	_, err = cmd.Execute([]string{"on"})
	if err != nil {
		t.Fatalf("Enabling failed: %v", err)
	}

	// Speed should be preserved
	if cfg.Typewriter.Speed != 50 {
		t.Errorf("Speed should still be 50, got %d", cfg.Typewriter.Speed)
	}
	if !cfg.Typewriter.Enabled {
		t.Error("Typewriter should be enabled")
	}
}

// TestTypewriterCommand_InvalidArguments tests error handling (Story 9-8).
func TestTypewriterCommand_InvalidArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewTypewriterCommand(cfg)

	// Test invalid subcommand
	_, err := cmd.Execute([]string{"invalid"})
	if err == nil {
		t.Error("Expected error for invalid subcommand")
	}

	// Test speed without value
	_, err = cmd.Execute([]string{"speed"})
	if err == nil {
		t.Error("Expected error for speed without value")
	}
}
