package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func setupTestConfig(t *testing.T) (*config.Config, string) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultConfig()
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	loaded, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	return loaded, configPath
}

func TestSpeedCommand(t *testing.T) {
	t.Run("command name", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cmd := NewSpeedCommand(cfg)

		if cmd.Name() != "speed" {
			t.Errorf("Expected name 'speed', got '%s'", cmd.Name())
		}
	})

	t.Run("show status", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !strings.Contains(output, "打字機效果狀態") {
			t.Error("Status output should contain status header")
		}

		if !strings.Contains(output, "40") {
			t.Error("Status should show default speed 40")
		}
	})

	t.Run("enable typewriter", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)
		cfg.Typewriter.Enabled = false

		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{"on"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !cfg.Typewriter.Enabled {
			t.Error("Typewriter should be enabled")
		}

		if !strings.Contains(output, "已啟用") {
			t.Error("Output should confirm enabled")
		}
	})

	t.Run("disable typewriter", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)
		cfg.Typewriter.Enabled = true

		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{"off"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Enabled {
			t.Error("Typewriter should be disabled")
		}

		if !strings.Contains(output, "已停用") {
			t.Error("Output should confirm disabled")
		}
	})

	t.Run("set valid speed", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{"60"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Speed != 60 {
			t.Errorf("Expected speed 60, got %d", cfg.Typewriter.Speed)
		}

		if !cfg.Typewriter.Enabled {
			t.Error("Setting speed should enable typewriter")
		}

		if !strings.Contains(output, "60") {
			t.Error("Output should show new speed")
		}
	})

	t.Run("speed too low uses default", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{"5"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Speed != 40 {
			t.Errorf("Expected default speed 40, got %d", cfg.Typewriter.Speed)
		}

		if !strings.Contains(output, "超出範圍") {
			t.Error("Output should warn about out of range")
		}

		if !strings.Contains(output, "40") {
			t.Error("Output should show default value")
		}
	})

	t.Run("speed too high uses default", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		output, err := cmd.Execute([]string{"250"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Speed != 40 {
			t.Errorf("Expected default speed 40, got %d", cfg.Typewriter.Speed)
		}

		if !strings.Contains(output, "超出範圍") {
			t.Error("Output should warn about out of range")
		}
	})

	t.Run("speed at min boundary", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		_, err := cmd.Execute([]string{"10"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Speed != 10 {
			t.Errorf("Expected speed 10, got %d", cfg.Typewriter.Speed)
		}
	})

	t.Run("speed at max boundary", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		_, err := cmd.Execute([]string{"200"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Typewriter.Speed != 200 {
			t.Errorf("Expected speed 200, got %d", cfg.Typewriter.Speed)
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cmd := NewSpeedCommand(cfg)

		_, err := cmd.Execute([]string{"abc"})
		if err == nil {
			t.Error("Expected error for invalid argument")
		}

		if !strings.Contains(err.Error(), "無效") {
			t.Errorf("Error should mention invalid argument, got: %v", err)
		}
	})

	t.Run("help text", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cmd := NewSpeedCommand(cfg)

		help := cmd.Help()
		if !strings.Contains(help, "打字機") {
			t.Error("Help should mention typewriter")
		}

		if !strings.Contains(help, "speed") {
			t.Error("Help should mention command name")
		}
	})

	t.Run("config persistence", func(t *testing.T) {
		cfg, configPath := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)
		_, err := cmd.Execute([]string{"80"})
		if err != nil {
			t.Fatalf("Set speed failed: %v", err)
		}

		// Manually save to test path since Save() uses default path
		if err := cfg.SaveToPath(configPath); err != nil {
			t.Fatalf("Save to test path failed: %v", err)
		}

		// Load from file to verify persistence
		cfg2, err := config.LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("Load config failed: %v", err)
		}

		if cfg2.Typewriter.Speed != 80 {
			t.Errorf("Persisted speed should be 80, got %d", cfg2.Typewriter.Speed)
		}

		if !cfg2.Typewriter.Enabled {
			t.Error("Persisted enabled should be true")
		}
	})
}

func TestSpeedCommandEdgeCases(t *testing.T) {
	t.Run("multiple speed changes", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)

		cmd := NewSpeedCommand(cfg)

		// Set to 30
		_, err := cmd.Execute([]string{"30"})
		if err != nil {
			t.Fatalf("First set failed: %v", err)
		}

		// Set to 100
		_, err = cmd.Execute([]string{"100"})
		if err != nil {
			t.Fatalf("Second set failed: %v", err)
		}

		if cfg.Typewriter.Speed != 100 {
			t.Errorf("Expected final speed 100, got %d", cfg.Typewriter.Speed)
		}
	})

	t.Run("enable then set speed", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)
		cfg.Typewriter.Enabled = false

		cmd := NewSpeedCommand(cfg)

		// Enable first
		_, err := cmd.Execute([]string{"on"})
		if err != nil {
			t.Fatalf("Enable failed: %v", err)
		}

		// Then set speed
		_, err = cmd.Execute([]string{"50"})
		if err != nil {
			t.Fatalf("Set speed failed: %v", err)
		}

		if cfg.Typewriter.Speed != 50 {
			t.Errorf("Expected speed 50, got %d", cfg.Typewriter.Speed)
		}

		if !cfg.Typewriter.Enabled {
			t.Error("Should remain enabled")
		}
	})

	t.Run("disable then enable restores speed", func(t *testing.T) {
		cfg, _ := setupTestConfig(t)
		cfg.Typewriter.Speed = 75

		cmd := NewSpeedCommand(cfg)

		// Disable
		_, err := cmd.Execute([]string{"off"})
		if err != nil {
			t.Fatalf("Disable failed: %v", err)
		}

		// Enable again
		_, err = cmd.Execute([]string{"on"})
		if err != nil {
			t.Fatalf("Enable failed: %v", err)
		}

		// Speed should be unchanged
		if cfg.Typewriter.Speed != 75 {
			t.Errorf("Expected speed preserved as 75, got %d", cfg.Typewriter.Speed)
		}
	})
}
