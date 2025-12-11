// Package audio provides audio playback management for Nightmare Assault.
// It handles BGM (Background Music) and SFX (Sound Effects) using oto v3.
package audio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// AudioManager manages all audio playback (BGM and SFX).
type AudioManager struct {
	ctx         *oto.Context
	bgmPlayer   *BGMPlayer // BGM player instance
	sfxPlayer   *SFXPlayer // SFX player instance
	config      config.AudioConfig
	initialized bool
	errorShown  bool // Prevent repeated error messages
	mu          sync.RWMutex
	audioDir    string // Path to ~/.nightmare/audio/
}

// NewAudioManager creates a new audio manager instance.
func NewAudioManager(cfg config.AudioConfig) *AudioManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("[WARN] Failed to get home directory: %v\n", err)
		homeDir = "."
	}
	audioDir := filepath.Join(homeDir, ".nightmare", "audio")

	return &AudioManager{
		config:   cfg,
		audioDir: audioDir,
	}
}

// Initialize initializes the audio system with oto v3.
// Returns error if initialization fails, but does not block the game.
func (m *AudioManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if audio files exist
	if !m.checkAudioFiles() {
		m.handleAudioError(fmt.Errorf("audio files not found"))
		return fmt.Errorf("audio files not complete, run 'nightmare --download-audio' or continue in silent mode")
	}

	// Initialize oto context
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		m.handleAudioError(err)
		return err
	}

	// Wait for initialization (with timeout to prevent blocking)
	select {
	case <-ready:
		m.ctx = ctx
		m.initialized = true

		// Initialize BGM player
		m.bgmPlayer = NewBGMPlayer(ctx, m.config, m.audioDir)

		// Initialize SFX player
		m.sfxPlayer = NewSFXPlayer(ctx, m.config, m.audioDir)

		log.Println("[INFO] Audio system initialized successfully")
		return nil
	case <-time.After(100 * time.Millisecond):
		m.handleAudioError(fmt.Errorf("audio initialization timeout"))
		return fmt.Errorf("audio initialization timeout")
	}
}

// InitializeAsync initializes the audio system in a goroutine to avoid blocking the main thread.
func (m *AudioManager) InitializeAsync() {
	go func() {
		if err := m.Initialize(); err != nil {
			log.Printf("[WARN] Audio initialization failed: %v\n", err)
		}
	}()
}

// checkAudioFiles checks if the required audio files exist in ~/.nightmare/audio/.
// Returns true if at least the basic structure exists.
func (m *AudioManager) checkAudioFiles() bool {
	// Check if audio directory exists
	if _, err := os.Stat(m.audioDir); os.IsNotExist(err) {
		return false
	}

	// Check for BGM directory
	bgmDir := filepath.Join(m.audioDir, "bgm")
	if _, err := os.Stat(bgmDir); os.IsNotExist(err) {
		return false
	}

	// Check for SFX directory
	sfxDir := filepath.Join(m.audioDir, "sfx")
	if _, err := os.Stat(sfxDir); os.IsNotExist(err) {
		return false
	}

	// Count files (at least 6 BGM + 10 SFX = 16 files)
	bgmFiles, _ := filepath.Glob(filepath.Join(bgmDir, "*"))
	sfxFiles, _ := filepath.Glob(filepath.Join(sfxDir, "*"))

	if len(bgmFiles) < 6 || len(sfxFiles) < 8 {
		log.Printf("[WARN] Incomplete audio files: %d BGM, %d SFX (expected 6+ BGM, 8+ SFX)\n", len(bgmFiles), len(sfxFiles))
		return false
	}

	return true
}

// handleAudioError logs the error and shows a one-time warning to the user.
func (m *AudioManager) handleAudioError(err error) {
	// Log error to debug.log
	logPath := filepath.Join(filepath.Dir(m.audioDir), "debug.log")
	f, logErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logErr == nil {
		defer f.Close()
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(f, "[%s] [AUDIO ERROR] %v\n", timestamp, err)
	}

	// Show one-time warning to user
	if !m.errorShown {
		fmt.Println("\n⚠️  音訊播放失敗，繼續靜音模式")
		fmt.Println("   可執行 'nightmare --download-audio' 下載音訊檔案")
		m.errorShown = true
	}
}

// IsInitialized returns whether the audio system is initialized.
func (m *AudioManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// Context returns the oto context (for BGM/SFX players).
func (m *AudioManager) Context() *oto.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx
}

// AudioDir returns the path to the audio directory.
func (m *AudioManager) AudioDir() string {
	return m.audioDir
}

// BGMPlayer returns the BGM player instance.
// Returns nil if audio system is not initialized.
func (m *AudioManager) BGMPlayer() *BGMPlayer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bgmPlayer
}

// SFXPlayer returns the SFX player instance.
// Returns nil if audio system is not initialized.
func (m *AudioManager) SFXPlayer() *SFXPlayer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sfxPlayer
}

// UpdateConfig updates the audio configuration.
func (m *AudioManager) UpdateConfig(cfg config.AudioConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

// Config returns the current audio configuration.
func (m *AudioManager) Config() config.AudioConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Shutdown gracefully shuts down the audio system.
// Note: oto v3 Context does not require explicit closing.
// The context will be garbage collected when no longer referenced.
func (m *AudioManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Shutdown BGM player
	if m.bgmPlayer != nil {
		m.bgmPlayer.Shutdown()
		m.bgmPlayer = nil
	}

	// Shutdown SFX player
	if m.sfxPlayer != nil {
		m.sfxPlayer.StopAll()
		m.sfxPlayer = nil
	}

	// oto v3 context doesn't have a Close method
	// Simply set to nil to allow garbage collection
	m.ctx = nil
	m.initialized = false

	return nil
}
