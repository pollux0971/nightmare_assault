// Package engine provides the story generation engine for Nightmare Assault.
package engine

import (
	"math/rand"
	"sync"
	"time"
	"unicode/utf8"
)

// TypewriterConfig configures the typewriter effect.
type TypewriterConfig struct {
	MinCharsPerSecond int           // Minimum characters per second
	MaxCharsPerSecond int           // Maximum characters per second
	PunctuationDelay  time.Duration // Extra delay after punctuation
	ParagraphDelay    time.Duration // Extra delay between paragraphs
	Enabled           bool          // Enable/disable typewriter effect
	ShowCursor        bool          // Show typing cursor
	SAN               int           // Current SAN value for horror effects
}

// DefaultTypewriterConfig returns default typewriter settings.
func DefaultTypewriterConfig() TypewriterConfig {
	return TypewriterConfig{
		MinCharsPerSecond: 50,
		MaxCharsPerSecond: 80,
		PunctuationDelay:  100 * time.Millisecond,
		ParagraphDelay:    300 * time.Millisecond,
		Enabled:           true,
		ShowCursor:        true,
		SAN:               100,
	}
}

// TypewriterState tracks the current typewriter animation state.
type TypewriterState int

const (
	TypewriterIdle TypewriterState = iota
	TypewriterPlaying
	TypewriterPaused
	TypewriterSkipped
	TypewriterDone
)

// glitchState tracks text glitch effects for low SAN
type glitchState struct {
	lastGlitchTime time.Time
	repeatChar     bool // Whether to repeat the next character
	skipNext       bool // Whether to skip the next character
}

// speedVariation tracks speed variation effects for low SAN
type speedVariation struct {
	mode         speedMode     // Current speed mode
	modeDuration time.Duration // How long to stay in this mode
	modeStart    time.Time     // When this mode started
}

// speedMode defines different typing speed modes
type speedMode int

const (
	speedNormal speedMode = iota
	speedFast   // 100-150 chars/sec
	speedSlow   // 10-20 chars/sec
	speedStuck  // Pause for 200-500ms
)

// StreamBuffer manages buffered streaming content with typewriter effect.
type StreamBuffer struct {
	config         TypewriterConfig
	buffer         []rune
	displayed      int
	state          TypewriterState
	mu             sync.RWMutex
	onChar         func(r rune)
	onComplete     func()
	skipChan       chan struct{}
	pauseChan      chan struct{}
	resumeChan     chan struct{}
	doneChan       chan struct{}
	glitchState    glitchState // Tracks text glitch effects
	speedVariation speedVariation // Tracks speed variation effects
}

// NewStreamBuffer creates a new stream buffer.
func NewStreamBuffer(config TypewriterConfig) *StreamBuffer {
	return &StreamBuffer{
		config:     config,
		buffer:     make([]rune, 0, 1024),
		state:      TypewriterIdle,
		skipChan:   make(chan struct{}),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		doneChan:   make(chan struct{}),
	}
}

// SetCallbacks sets the character and completion callbacks.
func (b *StreamBuffer) SetCallbacks(onChar func(r rune), onComplete func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onChar = onChar
	b.onComplete = onComplete
}

// Append adds new content to the buffer.
func (b *StreamBuffer) Append(content string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, r := range content {
		b.buffer = append(b.buffer, r)
	}
}

// Start begins the typewriter animation.
func (b *StreamBuffer) Start() {
	b.mu.Lock()
	if b.state != TypewriterIdle {
		b.mu.Unlock()
		return
	}
	b.state = TypewriterPlaying
	b.mu.Unlock()

	go b.playAnimation()
}

// Pause pauses the animation.
func (b *StreamBuffer) Pause() {
	b.mu.Lock()
	if b.state == TypewriterPlaying {
		b.state = TypewriterPaused
		select {
		case b.pauseChan <- struct{}{}:
		default:
		}
	}
	b.mu.Unlock()
}

// Resume resumes a paused animation.
func (b *StreamBuffer) Resume() {
	b.mu.Lock()
	if b.state == TypewriterPaused {
		b.state = TypewriterPlaying
		select {
		case b.resumeChan <- struct{}{}:
		default:
		}
	}
	b.mu.Unlock()
}

// Skip skips to the end of the animation.
func (b *StreamBuffer) Skip() {
	b.mu.Lock()
	if b.state == TypewriterPlaying || b.state == TypewriterPaused {
		b.state = TypewriterSkipped
		select {
		case b.skipChan <- struct{}{}:
		default:
		}
	}
	b.mu.Unlock()
}

// State returns the current animation state.
func (b *StreamBuffer) State() TypewriterState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Wait blocks until animation completes.
func (b *StreamBuffer) Wait() {
	<-b.doneChan
}

// GetDisplayed returns the currently displayed content.
func (b *StreamBuffer) GetDisplayed() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return string(b.buffer[:b.displayed])
}

// GetFull returns the full buffered content.
func (b *StreamBuffer) GetFull() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return string(b.buffer)
}

// Progress returns the display progress as a percentage (0-100).
func (b *StreamBuffer) Progress() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.buffer) == 0 {
		return 100
	}
	return (b.displayed * 100) / len(b.buffer)
}

func (b *StreamBuffer) playAnimation() {
	defer func() {
		b.mu.Lock()
		b.state = TypewriterDone
		if b.onComplete != nil {
			b.onComplete()
		}
		b.mu.Unlock()
		close(b.doneChan)
	}()

	// Calculate base delay between characters
	avgSpeed := (b.config.MinCharsPerSecond + b.config.MaxCharsPerSecond) / 2
	baseDelay := time.Second / time.Duration(avgSpeed)

	// Initialize speed variation for low SAN effects
	b.speedVariation.mode = speedNormal
	b.speedVariation.modeStart = time.Now()

	for {
		b.mu.RLock()
		bufLen := len(b.buffer)
		displayed := b.displayed
		state := b.state
		b.mu.RUnlock()

		// Check if we're done
		if displayed >= bufLen {
			// Wait a bit more in case more content is coming
			time.Sleep(100 * time.Millisecond)
			b.mu.RLock()
			bufLen = len(b.buffer)
			displayed = b.displayed
			b.mu.RUnlock()
			if displayed >= bufLen {
				return
			}
		}

		// Handle state changes
		select {
		case <-b.skipChan:
			// Skip to end
			b.mu.Lock()
			for b.displayed < len(b.buffer) {
				if b.onChar != nil {
					b.onChar(b.buffer[b.displayed])
				}
				b.displayed++
			}
			b.mu.Unlock()
			return

		case <-b.pauseChan:
			// Wait for resume
			select {
			case <-b.resumeChan:
			case <-b.skipChan:
				b.mu.Lock()
				for b.displayed < len(b.buffer) {
					if b.onChar != nil {
						b.onChar(b.buffer[b.displayed])
					}
					b.displayed++
				}
				b.mu.Unlock()
				return
			}

		default:
			if state == TypewriterSkipped {
				return
			}
		}

		// Display next character
		b.mu.Lock()
		if b.displayed < len(b.buffer) {
			r := b.buffer[b.displayed]

			// Apply text glitch effects for low SAN (AC4)
			shouldDisplay, shouldRepeat := b.applyTextGlitches(r)

			if shouldDisplay {
				if b.onChar != nil {
					b.onChar(r)
				}
			}

			if !shouldRepeat {
				b.displayed++
			}

			// Calculate delay with horror effects (AC3)
			delay := b.calculateDelayWithEffects(baseDelay, r)
			b.mu.Unlock()

			time.Sleep(delay)
		} else {
			b.mu.Unlock()
		}
	}
}

// applyTextGlitches applies text corruption effects for low SAN (AC4)
// Returns: (shouldDisplay, shouldRepeat)
func (b *StreamBuffer) applyTextGlitches(r rune) (bool, bool) {
	// Only apply glitches if SAN < 20
	if b.config.SAN >= 20 {
		return true, false
	}

	// Calculate glitch probability based on SAN (lower SAN = higher chance)
	glitchMultiplier := float64(20-b.config.SAN) / 20.0

	// 5% chance (scaled by SAN) to repeat character
	if rand.Float64() < 0.05*glitchMultiplier {
		if !b.glitchState.repeatChar {
			b.glitchState.repeatChar = true
			return true, true // Display but don't advance
		} else {
			b.glitchState.repeatChar = false
			return true, false // Display and advance
		}
	}

	// 3% chance (scaled by SAN) to skip character
	if rand.Float64() < 0.03*glitchMultiplier {
		return false, false // Don't display, advance
	}

	// 2% chance (scaled by SAN) for flicker effect (display wrong char briefly)
	if rand.Float64() < 0.02*glitchMultiplier {
		// TODO: Implement flicker effect (requires TUI integration)
		// For now, just display normally
		return true, false
	}

	return true, false
}

// calculateDelayWithEffects calculates delay with horror speed variation (AC3)
func (b *StreamBuffer) calculateDelayWithEffects(baseDelay time.Duration, r rune) time.Duration {
	delay := baseDelay

	// Add punctuation delay
	if isPunctuation(r) {
		delay += b.config.PunctuationDelay
	}
	if r == '\n' {
		delay += b.config.ParagraphDelay
	}

	// Apply low SAN speed variation (AC3)
	if b.config.SAN < 40 {
		delay = b.applySpeedVariation(delay)
	}

	return delay
}

// applySpeedVariation applies random speed changes for low SAN (AC3)
func (b *StreamBuffer) applySpeedVariation(baseDelay time.Duration) time.Duration {
	now := time.Now()

	// Check if we should change speed mode
	if now.Sub(b.speedVariation.modeStart) >= b.speedVariation.modeDuration {
		b.updateSpeedMode()
	}

	switch b.speedVariation.mode {
	case speedFast:
		// 100-150 chars/sec = 6.6-10ms per char
		return time.Millisecond * time.Duration(6+rand.Intn(4))
	case speedSlow:
		// 10-20 chars/sec = 50-100ms per char
		return time.Millisecond * time.Duration(50+rand.Intn(50))
	case speedStuck:
		// Pause 200-500ms
		return time.Millisecond * time.Duration(200+rand.Intn(300))
	default:
		return baseDelay
	}
}

// updateSpeedMode randomly changes the speed mode
func (b *StreamBuffer) updateSpeedMode() {
	// Determine new mode based on random chance
	roll := rand.Float64()

	if roll < 0.1 {
		// 10% chance of speed up
		b.speedVariation.mode = speedFast
		b.speedVariation.modeDuration = time.Millisecond * time.Duration(500+rand.Intn(500)) // 0.5-1s
	} else if roll < 0.2 {
		// 10% chance of slow down
		b.speedVariation.mode = speedSlow
		b.speedVariation.modeDuration = time.Millisecond * time.Duration(500+rand.Intn(500)) // 0.5-1s
	} else if roll < 0.25 {
		// 5% chance of stuck
		b.speedVariation.mode = speedStuck
		b.speedVariation.modeDuration = time.Millisecond * time.Duration(200+rand.Intn(300)) // 0.2-0.5s
	} else {
		// 75% chance of normal speed
		b.speedVariation.mode = speedNormal
		b.speedVariation.modeDuration = time.Second * time.Duration(2+rand.Intn(3)) // 2-5s
	}

	b.speedVariation.modeStart = time.Now()
}

func isPunctuation(r rune) bool {
	switch r {
	case '.', '!', '?', '。', '！', '？', '，', ',', '：', ':', '；', ';':
		return true
	}
	return false
}

// StreamingRenderer provides real-time rendering of streamed content.
type StreamingRenderer struct {
	buffer         *StreamBuffer
	content        string
	onUpdate       func(content string)
	onFinish       func(content string)
	mu             sync.RWMutex
}

// NewStreamingRenderer creates a new streaming renderer.
func NewStreamingRenderer(config TypewriterConfig) *StreamingRenderer {
	buffer := NewStreamBuffer(config)
	renderer := &StreamingRenderer{
		buffer: buffer,
	}

	buffer.SetCallbacks(func(r rune) {
		renderer.mu.Lock()
		renderer.content += string(r)
		content := renderer.content
		onUpdate := renderer.onUpdate
		renderer.mu.Unlock()

		if onUpdate != nil {
			onUpdate(content)
		}
	}, func() {
		renderer.mu.RLock()
		content := renderer.content
		onFinish := renderer.onFinish
		renderer.mu.RUnlock()

		if onFinish != nil {
			onFinish(content)
		}
	})

	return renderer
}

// SetUpdateCallback sets the callback for content updates.
func (r *StreamingRenderer) SetUpdateCallback(callback func(content string)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onUpdate = callback
}

// SetFinishCallback sets the callback for completion.
func (r *StreamingRenderer) SetFinishCallback(callback func(content string)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onFinish = callback
}

// AppendChunk adds a new chunk from the stream.
func (r *StreamingRenderer) AppendChunk(chunk string) {
	r.buffer.Append(chunk)
}

// Start begins rendering.
func (r *StreamingRenderer) Start() {
	r.buffer.Start()
}

// Skip skips to the end.
func (r *StreamingRenderer) Skip() {
	r.buffer.Skip()
}

// Pause pauses rendering.
func (r *StreamingRenderer) Pause() {
	r.buffer.Pause()
}

// Resume resumes rendering.
func (r *StreamingRenderer) Resume() {
	r.buffer.Resume()
}

// Wait blocks until rendering is complete.
func (r *StreamingRenderer) Wait() {
	r.buffer.Wait()
}

// GetContent returns the currently rendered content.
func (r *StreamingRenderer) GetContent() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.content
}

// Progress returns the rendering progress (0-100).
func (r *StreamingRenderer) Progress() int {
	return r.buffer.Progress()
}

// RuneCount returns the number of runes in a string.
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}
