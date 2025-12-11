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
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// oggToInt16Adapter adapts oggvorbis.Reader (float32 samples) to io.ReadCloser (int16 PCM)
type oggToInt16Adapter struct {
	reader *oggvorbis.Reader
	buffer []float32
}

func (a *oggToInt16Adapter) Read(p []byte) (n int, err error) {
	// Calculate how many samples we need (2 bytes per sample for int16)
	samplesNeeded := len(p) / 2

	// Resize buffer if needed
	if len(a.buffer) < samplesNeeded {
		a.buffer = make([]float32, samplesNeeded)
	}

	// Read float32 samples from oggvorbis
	samplesRead, err := a.reader.Read(a.buffer[:samplesNeeded])
	if err != nil && err != io.EOF {
		return 0, err
	}

	// Convert float32 (-1.0 to 1.0) to int16 (-32768 to 32767)
	for i := 0; i < samplesRead; i++ {
		sample := a.buffer[i]
		// Clamp to [-1.0, 1.0]
		if sample > 1.0 {
			sample = 1.0
		}
		if sample < -1.0 {
			sample = -1.0
		}
		// Convert to int16
		int16Sample := int16(sample * 32767.0)
		// Write as little-endian bytes
		p[i*2] = byte(int16Sample)
		p[i*2+1] = byte(int16Sample >> 8)
	}

	bytesWritten := samplesRead * 2
	if err == io.EOF {
		return bytesWritten, io.EOF
	}
	return bytesWritten, nil
}

func (a *oggToInt16Adapter) Close() error {
	return nil // oggvorbis.Reader doesn't need explicit close
}

// BGMScene represents different game scene types for BGM selection
type BGMScene string

const (
	BGMSceneExploration BGMScene = "exploration" // 探索場景（預設）
	BGMSceneChase       BGMScene = "chase"       // 緊張/追逐場景
	BGMSceneSafe        BGMScene = "safe"        // 安全區/休息場景
	BGMSceneHorror      BGMScene = "horror"      // 恐怖揭露時刻
	BGMSceneMystery     BGMScene = "mystery"     // 解謎場景
	BGMSceneDeath       BGMScene = "death"       // 死亡/結局場景
)

// BGMPlayer manages background music playback
type BGMPlayer struct {
	ctx           *oto.Context
	currentPlayer *oto.Player
	nextPlayer    *oto.Player // For crossfade
	currentBGM    string
	currentMood   MoodType  // Current mood for anti-frequent-switch
	volume        float64
	targetVolume  float64 // Target volume for fade in
	enabled       bool
	loopEnabled   bool
	lastSwitch    time.Time // For 6.4c anti-frequent-switch
	audioDir      string    // Path to ~/.nightmare/audio/
	mu            sync.RWMutex
	stopChan      chan struct{} // For stopping loop goroutine
}

// MoodType represents game mood (imported from engine package to avoid circular dependency)
type MoodType = engine.MoodType

// NewBGMPlayer creates a new BGM player
func NewBGMPlayer(ctx *oto.Context, cfg config.AudioConfig, audioDir string) *BGMPlayer {
	return &BGMPlayer{
		ctx:          ctx,
		volume:       cfg.BGMVolume,
		targetVolume: cfg.BGMVolume,
		enabled:      cfg.BGMEnabled,
		loopEnabled:  true, // Default to loop mode
		currentMood:  engine.MoodExploration, // Default mood
		lastSwitch:   time.Time{}, // Zero time (never switched)
		audioDir:     audioDir,
		stopChan:     make(chan struct{}),
	}
}

// Play plays a BGM file
// filename can be full path or just the base name (e.g., "ambient_exploration.mp3")
func (p *BGMPlayer) Play(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		log.Println("[INFO] BGM is disabled, skipping playback")
		return nil
	}

	// If ctx is nil, we're in silent mode
	if p.ctx == nil {
		log.Println("[WARN] Audio context is nil, cannot play BGM")
		return fmt.Errorf("audio context not initialized")
	}

	// Resolve full path if needed
	fullPath := filename
	if !filepath.IsAbs(filename) {
		fullPath = filepath.Join(p.audioDir, "bgm", filename)
	}

	// Validate file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("BGM file not found: %s", fullPath)
	}

	// Stop current playback
	if p.currentPlayer != nil {
		p.currentPlayer.Close()
		p.currentPlayer = nil
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open BGM file: %w", err)
	}

	// Decode based on format
	decoder, format, err := p.decodeAudioFile(file, fullPath)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to decode BGM file: %w", err)
	}

	log.Printf("[INFO] Playing BGM: %s (format: %s, volume: %.2f)\n", filepath.Base(fullPath), format, p.volume)

	// Create oto player
	player := p.ctx.NewPlayer(decoder)

	// Set volume
	player.SetVolume(p.volume)

	// Start playback
	player.Play()

	p.currentPlayer = player
	p.currentBGM = fullPath
	p.lastSwitch = time.Now()

	// If loop is enabled, start loop goroutine
	if p.loopEnabled {
		go p.loopPlayback(fullPath)
	}

	return nil
}

// decodeAudioFile decodes audio file based on its format
func (p *BGMPlayer) decodeAudioFile(file *os.File, fullPath string) (io.ReadCloser, AudioFormat, error) {
	ext := filepath.Ext(fullPath)

	switch ext {
	case ".mp3":
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			return nil, FormatMP3, fmt.Errorf("MP3 decode error: %w", err)
		}
		return io.NopCloser(decoder), FormatMP3, nil

	case ".ogg":
		reader, err := oggvorbis.NewReader(file)
		if err != nil {
			return nil, FormatOGG, fmt.Errorf("OGG decode error: %w", err)
		}
		// oggvorbis.Reader returns float32 samples, need to convert to int16 PCM
		adapter := &oggToInt16Adapter{reader: reader}
		return adapter, FormatOGG, nil

	case ".wav":
		// WAV files are typically raw PCM, but we need proper decoding
		// For now, treat as raw stream (this might need wav package later)
		return file, FormatWAV, nil

	default:
		return nil, FormatUnknown, fmt.Errorf("unsupported audio format: %s", ext)
	}
}

// loopPlayback handles seamless loop playback
func (p *BGMPlayer) loopPlayback(filename string) {
	for {
		select {
		case <-p.stopChan:
			return
		case <-time.After(100 * time.Millisecond):
			p.mu.RLock()
			if p.currentPlayer == nil || !p.loopEnabled {
				p.mu.RUnlock()
				return
			}

			// Check if playback finished (oto player is at end)
			// Note: oto v3 doesn't have a direct "IsPlaying()" method
			// We'll need to track this via player state or implement in AC tests
			p.mu.RUnlock()
		}
	}
}

// Stop stops BGM playback
func (p *BGMPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentPlayer != nil {
		// Signal loop goroutine to stop
		select {
		case p.stopChan <- struct{}{}:
		default:
		}

		p.currentPlayer.Close()
		p.currentPlayer = nil
		p.currentBGM = ""
		log.Println("[INFO] BGM playback stopped")
	}
}

// SetVolume sets BGM volume (0.0-1.0)
func (p *BGMPlayer) SetVolume(volume float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clamp volume to valid range
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	p.volume = volume

	// Apply to current player if playing
	if p.currentPlayer != nil {
		p.currentPlayer.SetVolume(volume)
	}

	log.Printf("[INFO] BGM volume set to %.2f\n", volume)
}

// Volume returns current volume
func (p *BGMPlayer) Volume() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volume
}

// Enable enables BGM playback
func (p *BGMPlayer) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
	log.Println("[INFO] BGM enabled")
}

// Disable disables BGM playback
func (p *BGMPlayer) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
	log.Println("[INFO] BGM disabled")
}

// IsEnabled returns whether BGM is enabled
func (p *BGMPlayer) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// IsPlaying returns whether BGM is currently playing
func (p *BGMPlayer) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPlayer != nil
}

// CurrentBGM returns the currently playing BGM filename
func (p *BGMPlayer) CurrentBGM() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentBGM
}

// LastSwitchTime returns the last BGM switch time (for anti-frequent-switch in 6.4c)
func (p *BGMPlayer) LastSwitchTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastSwitch
}

// SetLoop enables/disables loop mode
func (p *BGMPlayer) SetLoop(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loopEnabled = enabled
	log.Printf("[INFO] BGM loop mode: %v\n", enabled)
}

// IsLoopEnabled returns whether loop mode is enabled
func (p *BGMPlayer) IsLoopEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loopEnabled
}

// Shutdown gracefully shuts down the BGM player
func (p *BGMPlayer) Shutdown() {
	p.Stop()
	log.Println("[INFO] BGM player shut down")
}

// SetTargetVolume sets the target volume for fade in
func (p *BGMPlayer) SetTargetVolume(volume float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	p.targetVolume = volume
}

// FadeOut fades out the current BGM over the specified duration
func (p *BGMPlayer) FadeOut(duration time.Duration) error {
	p.mu.Lock()
	startVolume := p.volume
	p.mu.Unlock()

	if startVolume == 0.0 {
		return nil // Already at zero volume
	}

	steps := 50
	interval := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		progress := float64(i+1) / float64(steps)
		fadeVolume := startVolume * (1.0 - linearFadeCurve(progress))

		p.SetVolume(fadeVolume)
		time.Sleep(interval)
	}

	// Ensure final volume is 0.0
	p.SetVolume(0.0)
	log.Printf("[INFO] BGM faded out over %v\n", duration)

	return nil
}

// FadeIn fades in the current BGM over the specified duration to target volume
func (p *BGMPlayer) FadeIn(duration time.Duration) error {
	p.mu.RLock()
	targetVol := p.targetVolume
	p.mu.RUnlock()

	if targetVol == 0.0 {
		return nil // Target is zero, nothing to fade in to
	}

	// Set volume to 0 first
	p.SetVolume(0.0)

	steps := 50
	interval := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		progress := float64(i+1) / float64(steps)
		fadeVolume := targetVol * linearFadeCurve(progress)

		p.SetVolume(fadeVolume)
		time.Sleep(interval)
	}

	// Ensure final volume is target
	p.SetVolume(targetVol)
	log.Printf("[INFO] BGM faded in over %v to volume %.2f\n", duration, targetVol)

	return nil
}

// Crossfade performs a crossfade transition from current BGM to new BGM
func (p *BGMPlayer) Crossfade(newBGM string, duration time.Duration) error {
	p.mu.Lock()
	if p.ctx == nil {
		p.mu.Unlock()
		return fmt.Errorf("audio context is nil, cannot crossfade")
	}

	if !p.enabled {
		p.mu.Unlock()
		log.Println("[INFO] BGM is disabled, skipping crossfade")
		return nil
	}

	oldPlayer := p.currentPlayer
	oldVolume := p.volume
	p.mu.Unlock()

	// Resolve full path for new BGM
	fullPath := newBGM
	if !filepath.IsAbs(newBGM) {
		fullPath = filepath.Join(p.audioDir, "bgm", newBGM)
	}

	// Validate new BGM file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("new BGM file not found: %s", fullPath)
	}

	// Open and decode new BGM file
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open new BGM file: %w", err)
	}

	decoder, format, err := p.decodeAudioFile(file, fullPath)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to decode new BGM file: %w", err)
	}

	log.Printf("[INFO] Crossfading to %s (format: %s) over %v\n", filepath.Base(fullPath), format, duration)

	// Create new player
	p.mu.Lock()
	newPlayer := p.ctx.NewPlayer(decoder)
	newPlayer.SetVolume(0.0) // Start at zero
	newPlayer.Play()
	p.nextPlayer = newPlayer
	p.mu.Unlock()

	// Perform crossfade
	steps := 50
	interval := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		progress := float64(i+1) / float64(steps)
		fadeCurve := linearFadeCurve(progress)

		// Fade out old
		if oldPlayer != nil {
			oldVolFade := oldVolume * (1.0 - fadeCurve)
			oldPlayer.SetVolume(oldVolFade)
		}

		// Fade in new
		newVolFade := oldVolume * fadeCurve
		newPlayer.SetVolume(newVolFade)

		time.Sleep(interval)
	}

	// Finalize transition
	p.mu.Lock()
	if oldPlayer != nil {
		oldPlayer.Close()
	}
	p.currentPlayer = newPlayer
	p.nextPlayer = nil
	p.currentBGM = fullPath
	p.volume = oldVolume
	p.lastSwitch = time.Now()
	p.mu.Unlock()

	log.Printf("[INFO] Crossfade complete to %s\n", filepath.Base(fullPath))

	return nil
}

// linearFadeCurve returns a linear fade curve value (0.0 to 1.0) for given progress
func linearFadeCurve(progress float64) float64 {
	if progress < 0.0 {
		return 0.0
	}
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// GetBGMFilename returns the BGM filename for a given scene
func GetBGMFilename(scene BGMScene) string {
	switch scene {
	case BGMSceneExploration:
		return "ambient_exploration.mp3"
	case BGMSceneChase:
		return "tension_chase.mp3"
	case BGMSceneSafe:
		return "safe_rest.mp3"
	case BGMSceneHorror:
		return "horror_reveal.mp3"
	case BGMSceneMystery:
		return "mystery_puzzle.mp3"
	case BGMSceneDeath:
		return "ending_death.mp3"
	default:
		return "ambient_exploration.mp3" // Default
	}
}

// CanSwitch checks if BGM can be switched to a new mood
// Returns false if:
// - newMood is the same as currentMood
// - Last switch was less than 30 seconds ago
func (p *BGMPlayer) CanSwitch(newMood engine.MoodType) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if same mood
	if p.currentMood == newMood {
		return false
	}

	// Check minimum interval (30 seconds)
	// If lastSwitch is zero time (never switched), allow switch
	if !p.lastSwitch.IsZero() && time.Since(p.lastSwitch) < 30*time.Second {
		return false
	}

	return true
}

// SwitchByMood switches BGM based on mood type
// Silently ignores switch if CanSwitch returns false
// Updates currentMood and lastSwitch even if Crossfade fails (for state tracking)
func (p *BGMPlayer) SwitchByMood(mood engine.MoodType) error {
	if !p.CanSwitch(mood) {
		return nil // Silently skip
	}

	// Update mood state before attempting crossfade (for tracking even when audio unavailable)
	p.mu.Lock()
	p.currentMood = mood
	p.lastSwitch = time.Now()
	p.mu.Unlock()

	bgmFile := GetBGMForMood(mood)
	if err := p.Crossfade(bgmFile, 2*time.Second); err != nil {
		// Return error but mood state is already updated
		return err
	}

	log.Printf("[INFO] BGM auto-switched to %s (mood: %s)\n", bgmFile, mood.String())
	return nil
}

// GetCurrentMood returns the current mood type
func (p *BGMPlayer) GetCurrentMood() engine.MoodType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentMood
}
