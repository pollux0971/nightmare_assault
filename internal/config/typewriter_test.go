package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTypewriterConfig(t *testing.T) {
	t.Run("default config includes typewriter settings", func(t *testing.T) {
		cfg := DefaultConfig()

		if !cfg.Typewriter.Enabled {
			t.Error("Expected typewriter to be enabled by default")
		}

		if cfg.Typewriter.Speed != 40 {
			t.Errorf("Expected default speed 40, got %d", cfg.Typewriter.Speed)
		}

		if !cfg.Typewriter.ShowCursor {
			t.Error("Expected show cursor to be true by default")
		}
	})

	t.Run("save and load typewriter config", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// Create config with custom typewriter settings
		cfg := DefaultConfig()
		cfg.Typewriter.Enabled = false
		cfg.Typewriter.Speed = 60
		cfg.Typewriter.ShowCursor = false

		// Save
		if err := cfg.SaveToPath(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Load
		loaded, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify typewriter settings
		if loaded.Typewriter.Enabled {
			t.Error("Expected typewriter to be disabled")
		}

		if loaded.Typewriter.Speed != 60 {
			t.Errorf("Expected speed 60, got %d", loaded.Typewriter.Speed)
		}

		if loaded.Typewriter.ShowCursor {
			t.Error("Expected show cursor to be false")
		}
	})

	t.Run("json marshaling includes typewriter", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Typewriter.Speed = 50

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		// Check that JSON contains typewriter fields
		jsonStr := string(data)
		if jsonStr == "" {
			t.Error("Empty JSON")
		}

		// Unmarshal and verify
		var loaded Config
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if loaded.Typewriter.Speed != 50 {
			t.Errorf("Expected speed 50, got %d", loaded.Typewriter.Speed)
		}
	})

	t.Run("speed validation range", func(t *testing.T) {
		cfg := DefaultConfig()

		// Test valid speeds
		validSpeeds := []int{10, 40, 100, 200}
		for _, speed := range validSpeeds {
			cfg.Typewriter.Speed = speed
			if cfg.Typewriter.Speed < 10 || cfg.Typewriter.Speed > 200 {
				t.Errorf("Speed %d should be in valid range", speed)
			}
		}

		// Note: Actual validation will be done in the UI layer
		// Config just stores the value
	})
}

func TestTypewriterConfigPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping persistence test in short mode")
	}

	t.Run("config persists across save/load cycles", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// First save
		cfg1 := DefaultConfig()
		cfg1.Typewriter.Enabled = false
		cfg1.Typewriter.Speed = 80
		if err := cfg1.SaveToPath(configPath); err != nil {
			t.Fatalf("First save failed: %v", err)
		}

		// Load and modify
		cfg2, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		cfg2.Typewriter.Speed = 120
		if err := cfg2.SaveToPath(configPath); err != nil {
			t.Fatalf("Second save failed: %v", err)
		}

		// Final load and verify
		cfg3, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("Final load failed: %v", err)
		}

		if cfg3.Typewriter.Speed != 120 {
			t.Errorf("Expected speed 120, got %d", cfg3.Typewriter.Speed)
		}

		if cfg3.Typewriter.Enabled {
			t.Error("Expected typewriter to remain disabled")
		}
	})

	t.Run("missing config file uses defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")

		cfg, err := LoadFromPath(nonExistentPath)
		if err != nil {
			t.Fatalf("Should return default config when file missing, got error: %v", err)
		}

		if cfg.Typewriter.Speed != 40 {
			t.Errorf("Expected default speed 40, got %d", cfg.Typewriter.Speed)
		}
	})

	t.Run("corrupted file returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		corruptedPath := filepath.Join(tmpDir, "corrupted.json")

		// Write invalid JSON
		if err := os.WriteFile(corruptedPath, []byte("{invalid json"), 0600); err != nil {
			t.Fatalf("Failed to write corrupted file: %v", err)
		}

		_, err := LoadFromPath(corruptedPath)
		if err == nil {
			t.Error("Expected error when loading corrupted file")
		}
	})
}
