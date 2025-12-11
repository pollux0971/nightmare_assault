package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// SFX Priority constants
const (
	SFXPriorityDeath       = 100 // Highest priority (stops all other SFX)
	SFXPriorityWarning     = 80  // High priority (can interrupt environment)
	SFXPriorityHeartbeat   = 50  // Medium priority (looping)
	SFXPriorityEnvironment = 20  // Low priority (can be replaced)
)

// SFXChannel represents a single audio channel for SFX playback
type SFXChannel struct {
	player    *oto.Player
	priority  int
	startTime time.Time
	isPlaying bool
	sfxType   string // For debugging/tracking
	mu        sync.RWMutex
}

// SFXPlayer manages sound effect playback with 4-channel mixing
type SFXPlayer struct {
	ctx      *oto.Context
	channels [4]*SFXChannel // 4 simultaneous SFX channels
	volume   float64
	enabled  bool
	audioDir string
	mu       sync.RWMutex
}

// NewSFXPlayer creates a new SFX player
func NewSFXPlayer(ctx *oto.Context, cfg config.AudioConfig, audioDir string) *SFXPlayer {
	player := &SFXPlayer{
		ctx:      ctx,
		volume:   cfg.SFXVolume,
		enabled:  cfg.SFXEnabled,
		audioDir: audioDir,
	}

	// Initialize 4 channels
	for i := 0; i < 4; i++ {
		player.channels[i] = &SFXChannel{
			priority:  0,
			isPlaying: false,
			sfxType:   "",
		}
	}

	return player
}

// Enable enables SFX playback
func (p *SFXPlayer) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
	log.Println("[INFO] SFX enabled")
}

// Disable disables SFX playback and stops all channels
func (p *SFXPlayer) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false

	// Stop all channels
	for _, ch := range p.channels {
		p.stopChannel(ch)
	}

	log.Println("[INFO] SFX disabled")
}

// IsEnabled returns whether SFX is enabled
func (p *SFXPlayer) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// SetVolume sets the SFX volume (0.0 to 1.0)
func (p *SFXPlayer) SetVolume(volume float64) {
	// Clamp volume
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.volume = volume
	log.Printf("[INFO] SFX volume set to %.2f\n", volume)
}

// Volume returns the current SFX volume
func (p *SFXPlayer) Volume() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volume
}

// StopAll stops all SFX channels
func (p *SFXPlayer) StopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, ch := range p.channels {
		p.stopChannel(ch)
	}

	log.Println("[INFO] All SFX stopped")
}

// stopChannel stops a single channel (internal, must hold lock)
func (p *SFXPlayer) stopChannel(ch *SFXChannel) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.player != nil {
		ch.player.Close()
		ch.player = nil
	}

	ch.isPlaying = false
	ch.priority = 0
	ch.sfxType = ""
}

// PlaySFX plays a sound effect with given priority
// Returns error if audio context is nil or file not found
// Silently skips if SFX is disabled
func (p *SFXPlayer) PlaySFX(filename string, priority int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Skip if disabled
	if !p.enabled {
		return nil
	}

	// Check audio context
	if p.ctx == nil {
		return fmt.Errorf("audio context is nil")
	}

	// Find available or lowest priority channel
	channelIndex := p.findBestChannel(priority)
	if channelIndex == -1 {
		log.Printf("[WARN] No available channel for SFX %s (priority %d)\n", filename, priority)
		return nil // Silently skip if no channel available
	}

	// Load and play SFX
	if err := p.playSFXOnChannel(channelIndex, filename, priority); err != nil {
		return fmt.Errorf("failed to play SFX: %w", err)
	}

	return nil
}

// findBestChannel finds the best channel to use for SFX playback
// Returns index of best channel, or -1 if none available
func (p *SFXPlayer) findBestChannel(priority int) int {
	// First, try to find an idle channel
	for i, ch := range p.channels {
		ch.mu.RLock()
		isPlaying := ch.isPlaying
		ch.mu.RUnlock()

		if !isPlaying {
			return i
		}
	}

	// All channels busy - find lowest priority channel that can be replaced
	lowestPriorityIndex := -1
	lowestPriority := priority

	for i, ch := range p.channels {
		ch.mu.RLock()
		chPriority := ch.priority
		ch.mu.RUnlock()

		if chPriority < lowestPriority {
			lowestPriority = chPriority
			lowestPriorityIndex = i
		}
	}

	return lowestPriorityIndex
}

// playSFXOnChannel plays SFX on specific channel
func (p *SFXPlayer) playSFXOnChannel(channelIndex int, filename string, priority int) error {
	ch := p.channels[channelIndex]

	// Stop current playback if any
	p.stopChannel(ch)

	// Resolve full path
	fullPath := filepath.Join(p.audioDir, "sfx", filename)

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open SFX file: %w", err)
	}

	// Decode based on file extension
	var decoder io.ReadCloser
	ext := filepath.Ext(filename)

	switch ext {
	case ".mp3":
		mp3Dec, err := mp3.NewDecoder(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to create mp3 decoder: %w", err)
		}
		decoder = io.NopCloser(mp3Dec)
	case ".ogg":
		oggReader, err := oggvorbis.NewReader(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to create ogg decoder: %w", err)
		}
		decoder = &oggToInt16Adapter{reader: oggReader, buffer: make([]float32, 4096)}
	case ".wav":
		// WAV files can be read directly (assuming 16-bit PCM)
		decoder = file
	default:
		file.Close()
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Create oto player
	player := p.ctx.NewPlayer(decoder)

	// Set volume
	player.SetVolume(p.volume)

	// Start playback
	player.Play()

	// Update channel state
	ch.mu.Lock()
	ch.player = player
	ch.priority = priority
	ch.startTime = time.Now()
	ch.isPlaying = true
	ch.sfxType = filename
	ch.mu.Unlock()

	log.Printf("[INFO] SFX %s playing on channel %d (priority %d)\n", filename, channelIndex, priority)

	// Auto-cleanup when finished (goroutine)
	go func() {
		// Wait for playback to finish
		for player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
		}

		// Cleanup
		ch.mu.Lock()
		if ch.player == player {
			ch.player.Close()
			ch.player = nil
			ch.isPlaying = false
			ch.priority = 0
			ch.sfxType = ""
		}
		ch.mu.Unlock()

		decoder.Close()
		file.Close()
	}()

	return nil
}

// PlayWarning plays the warning SFX (high priority)
func (p *SFXPlayer) PlayWarning() error {
	return p.PlaySFX("warning.wav", SFXPriorityWarning)
}

// PlayDeath plays the death SFX (highest priority, stops all other SFX)
func (p *SFXPlayer) PlayDeath() error {
	// Stop all other SFX first
	p.StopAll()

	// Play death SFX
	return p.PlaySFX("death.wav", SFXPriorityDeath)
}

// PlayEnvironment plays an environment SFX based on tag
func (p *SFXPlayer) PlayEnvironment(tag string) error {
	filename := GetEnvironmentSFXFilename(tag)
	if filename == "" {
		return fmt.Errorf("unknown environment SFX tag: %s", tag)
	}
	return p.PlaySFX(filename, SFXPriorityEnvironment)
}

// GetEnvironmentSFXFilename maps SFX tags to filenames
func GetEnvironmentSFXFilename(tag string) string {
	switch tag {
	case "door_open", "door":
		return "door_open.wav"
	case "door_close":
		return "door_close.wav"
	case "footsteps", "steps":
		return "footsteps.wav"
	case "glass_break", "glass":
		return "glass_break.wav"
	case "thunder":
		return "thunder.wav"
	case "whisper":
		return "whisper.wav"
	default:
		return ""
	}
}
