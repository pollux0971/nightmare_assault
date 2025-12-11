package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}

	if cfg.Language != "zh-TW" {
		t.Errorf("Expected language zh-TW, got %s", cfg.Language)
	}

	if cfg.API.APIKeys == nil {
		t.Error("Expected APIKeys map to be initialized")
	}

	if cfg.API.LastTested == nil {
		t.Error("Expected LastTested map to be initialized")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	cfg := DefaultConfig()
	cfg.Language = "en"
	cfg.Theme = "dark"
	cfg.API.Smart.ProviderID = "openai"
	cfg.API.Smart.Model = "gpt-4o"

	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.Language != "en" {
		t.Errorf("Expected language en, got %s", loaded.Language)
	}

	if loaded.Theme != "dark" {
		t.Errorf("Expected theme dark, got %s", loaded.Theme)
	}

	if loaded.API.Smart.ProviderID != "openai" {
		t.Errorf("Expected provider openai, got %s", loaded.API.Smart.ProviderID)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Load from non-existent path should return default config
	cfg, err := LoadFromPath("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("Expected default config version 1.0, got %s", cfg.Version)
	}
}

func TestIsConfigured(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.IsConfigured() {
		t.Error("Expected unconfigured config")
	}

	cfg.API.Smart.ProviderID = "openai"

	if !cfg.IsConfigured() {
		t.Error("Expected configured config after setting smart provider")
	}
}

func TestHasAPIKey(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.HasAPIKey("openai") {
		t.Error("Expected no API key initially")
	}

	cfg.API.APIKeys["openai"] = "test-key"

	if !cfg.HasAPIKey("openai") {
		t.Error("Expected API key to be present")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "sk-test-api-key-12345"

	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if !IsEncrypted(encrypted) {
		t.Error("Expected encrypted value to have prefix")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestEncryptEmpty(t *testing.T) {
	encrypted, err := Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}

	if encrypted != "" {
		t.Errorf("Expected empty string, got %s", encrypted)
	}
}

func TestDecryptUnencrypted(t *testing.T) {
	// Decrypting unencrypted value should return as-is
	plaintext := "not-encrypted"
	decrypted, err := Decrypt(plaintext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-1234567890abcdef", "sk-1...cdef"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234...6789"},
	}

	for _, tt := range tests {
		result := MaskAPIKey(tt.input)
		if result != tt.expected {
			t.Errorf("MaskAPIKey(%s): expected %s, got %s", tt.input, tt.expected, result)
		}
	}
}

func TestConfigEncryptAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	apiKey := "sk-test-key-12345"

	if err := cfg.EncryptAPIKey("openai", apiKey); err != nil {
		t.Fatalf("EncryptAPIKey failed: %v", err)
	}

	// Check that the stored value is encrypted
	stored := cfg.API.APIKeys["openai"]
	if !IsEncrypted(stored) {
		t.Error("Expected stored API key to be encrypted")
	}

	// Decrypt and verify
	decrypted, err := cfg.DecryptAPIKey("openai")
	if err != nil {
		t.Fatalf("DecryptAPIKey failed: %v", err)
	}

	if decrypted != apiKey {
		t.Errorf("Expected %s, got %s", apiKey, decrypted)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".nightmare", "config.json")

	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}
