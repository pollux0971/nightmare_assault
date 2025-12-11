package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

func TestNewAudioManager(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	if manager == nil {
		t.Fatal("NewAudioManager returned nil")
	}

	if manager.audioDir == "" {
		t.Error("audioDir should not be empty")
	}

	if manager.initialized {
		t.Error("Manager should not be initialized yet")
	}

	if manager.errorShown {
		t.Error("errorShown should be false initially")
	}
}

func TestCheckAudioFiles_MissingDirectory(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)
	// Override audioDir to a non-existent path
	manager.audioDir = "/tmp/nightmare-test-nonexistent"

	if manager.checkAudioFiles() {
		t.Error("checkAudioFiles should return false for non-existent directory")
	}
}

func TestCheckAudioFiles_WithMockStructure(t *testing.T) {
	// Create temporary audio structure
	tmpDir := t.TempDir()
	bgmDir := filepath.Join(tmpDir, "bgm")
	sfxDir := filepath.Join(tmpDir, "sfx")

	if err := os.MkdirAll(bgmDir, 0755); err != nil {
		t.Fatalf("Failed to create bgm directory: %v", err)
	}
	if err := os.MkdirAll(sfxDir, 0755); err != nil {
		t.Fatalf("Failed to create sfx directory: %v", err)
	}

	// Create 6 mock BGM files
	for i := 1; i <= 6; i++ {
		filename := filepath.Join(bgmDir, "bgm_"+string(rune('0'+i))+".ogg")
		if err := os.WriteFile(filename, []byte("mock audio data"), 0644); err != nil {
			t.Fatalf("Failed to create mock BGM file: %v", err)
		}
	}

	// Create 10 mock SFX files
	for i := 1; i <= 10; i++ {
		filename := filepath.Join(sfxDir, "sfx_"+string(rune('0'+i))+".wav")
		if err := os.WriteFile(filename, []byte("mock audio data"), 0644); err != nil {
			t.Fatalf("Failed to create mock SFX file: %v", err)
		}
	}

	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)
	manager.audioDir = tmpDir

	if !manager.checkAudioFiles() {
		t.Error("checkAudioFiles should return true for complete audio structure")
	}
}

func TestCheckAudioFiles_IncompleteFiles(t *testing.T) {
	// Create temporary audio structure with insufficient files
	tmpDir := t.TempDir()
	bgmDir := filepath.Join(tmpDir, "bgm")
	sfxDir := filepath.Join(tmpDir, "sfx")

	if err := os.MkdirAll(bgmDir, 0755); err != nil {
		t.Fatalf("Failed to create bgm directory: %v", err)
	}
	if err := os.MkdirAll(sfxDir, 0755); err != nil {
		t.Fatalf("Failed to create sfx directory: %v", err)
	}

	// Create only 3 BGM files (insufficient)
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(bgmDir, "bgm_"+string(rune('0'+i))+".ogg")
		if err := os.WriteFile(filename, []byte("mock audio data"), 0644); err != nil {
			t.Fatalf("Failed to create mock BGM file: %v", err)
		}
	}

	// Create only 4 SFX files (insufficient)
	for i := 1; i <= 4; i++ {
		filename := filepath.Join(sfxDir, "sfx_"+string(rune('0'+i))+".wav")
		if err := os.WriteFile(filename, []byte("mock audio data"), 0644); err != nil {
			t.Fatalf("Failed to create mock SFX file: %v", err)
		}
	}

	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)
	manager.audioDir = tmpDir

	if manager.checkAudioFiles() {
		t.Error("checkAudioFiles should return false for incomplete audio files")
	}
}

func TestIsInitialized(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	if manager.IsInitialized() {
		t.Error("Manager should not be initialized yet")
	}

	// Manually set initialized (since we can't actually initialize without audio files)
	manager.initialized = true

	if !manager.IsInitialized() {
		t.Error("Manager should be initialized")
	}
}

func TestUpdateConfig(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	newCfg := config.AudioConfig{
		BGMEnabled: false,
		SFXEnabled: false,
		BGMVolume:  0.5,
		SFXVolume:  0.6,
	}

	manager.UpdateConfig(newCfg)

	updatedCfg := manager.Config()

	if updatedCfg.BGMEnabled != false {
		t.Error("BGMEnabled should be updated to false")
	}
	if updatedCfg.SFXEnabled != false {
		t.Error("SFXEnabled should be updated to false")
	}
	if updatedCfg.BGMVolume != 0.5 {
		t.Error("BGMVolume should be updated to 0.5")
	}
	if updatedCfg.SFXVolume != 0.6 {
		t.Error("SFXVolume should be updated to 0.6")
	}
}

func TestInitializeAsync_DoesNotBlock(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)
	manager.audioDir = "/tmp/nightmare-test-nonexistent" // Non-existent to trigger error

	start := time.Now()
	manager.InitializeAsync()
	elapsed := time.Since(start)

	// InitializeAsync should return immediately (< 10ms)
	if elapsed > 10*time.Millisecond {
		t.Errorf("InitializeAsync blocked for %v, should be immediate", elapsed)
	}
}

func TestAudioDir(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	audioDir := manager.AudioDir()

	if audioDir == "" {
		t.Error("AudioDir should not be empty")
	}

	if !filepath.IsAbs(audioDir) {
		t.Error("AudioDir should return an absolute path")
	}
}

func TestHandleAudioError_OnlyShowsOnce(t *testing.T) {
	cfg := config.AudioConfig{
		BGMEnabled: true,
		SFXEnabled: true,
		BGMVolume:  0.7,
		SFXVolume:  0.8,
	}

	manager := NewAudioManager(cfg)

	if manager.errorShown {
		t.Error("errorShown should be false initially")
	}

	// First call should show error
	manager.handleAudioError(nil)

	if !manager.errorShown {
		t.Error("errorShown should be true after first call")
	}

	// Second call should not show error again
	// (We can't easily test console output, but we verify the flag stays true)
	manager.handleAudioError(nil)

	if !manager.errorShown {
		t.Error("errorShown should remain true")
	}
}
