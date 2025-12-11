package effects

import (
	"time"
)

// BlinkFrequency represents how often visual elements should blink/flash.
type BlinkFrequency struct {
	Interval      time.Duration // Time between blinks
	FlashDuration time.Duration // How long the flash lasts
}

// CursorBlinkRate represents the cursor blink rate based on SAN.
type CursorBlinkRate struct {
	Interval time.Duration // Time between cursor toggles
}

// CalculateScreenFlashFrequency returns the flash/blink frequency for screen effects.
// AC1: SAN 60-79 = flash every 5-10 seconds
// AC2: SAN 40-59 = flash every 3-5 seconds
// AC3: SAN 20-39 = flash every 1-2 seconds
// AC4: SAN 1-19 = continuous distortion (flash every 500ms-1s)
func CalculateScreenFlashFrequency(san int) BlinkFrequency {
	switch {
	case san >= 80:
		// No flashing
		return BlinkFrequency{
			Interval:      0, // Disabled
			FlashDuration: 0,
		}
	case san >= 60:
		// AC1: Occasional flashing (5-10s intervals)
		return BlinkFrequency{
			Interval:      7 * time.Second, // Mid-range of 5-10s
			FlashDuration: 150 * time.Millisecond,
		}
	case san >= 40:
		// AC2: More frequent (3-5s intervals)
		return BlinkFrequency{
			Interval:      4 * time.Second, // Mid-range of 3-5s
			FlashDuration: 200 * time.Millisecond,
		}
	case san >= 20:
		// AC3: Frequent (1-2s intervals)
		return BlinkFrequency{
			Interval:      1500 * time.Millisecond, // 1.5s
			FlashDuration: 250 * time.Millisecond,
		}
	default:
		// AC4: Very frequent (500ms-1s intervals)
		return BlinkFrequency{
			Interval:      750 * time.Millisecond, // 0.75s
			FlashDuration: 300 * time.Millisecond,
		}
	}
}

// CalculateCursorBlinkRate returns the cursor blink rate based on SAN.
// Normal cursor blink is typically 500-600ms.
// AC3: SAN 20-39 = cursor blinks 2x faster (250-300ms)
// AC4: SAN 1-19 = cursor blinks erratically fast (100-200ms)
func CalculateCursorBlinkRate(san int) CursorBlinkRate {
	switch {
	case san >= 40:
		// Normal cursor blink rate
		return CursorBlinkRate{
			Interval: 530 * time.Millisecond, // Standard terminal cursor rate
		}
	case san >= 20:
		// AC3: 2x faster
		return CursorBlinkRate{
			Interval: 265 * time.Millisecond, // Half of normal
		}
	default:
		// AC4: Very fast, erratic
		return CursorBlinkRate{
			Interval: 150 * time.Millisecond,
		}
	}
}

// FlashState tracks the current state of a flashing element.
type FlashState struct {
	LastFlashTime time.Time
	IsFlashing    bool
	FlashStarted  time.Time
}

// NewFlashState creates a new flash state tracker.
func NewFlashState() *FlashState {
	return &FlashState{
		LastFlashTime: time.Now(),
		IsFlashing:    false,
	}
}

// Update updates the flash state based on current time and frequency.
// Returns true if flash state changed (either started or ended).
func (fs *FlashState) Update(now time.Time, frequency BlinkFrequency) bool {
	if frequency.Interval == 0 {
		// Flashing disabled
		if fs.IsFlashing {
			fs.IsFlashing = false
			return true
		}
		return false
	}

	stateChanged := false

	// Check if currently flashing
	if fs.IsFlashing {
		// Check if flash duration has elapsed
		if now.Sub(fs.FlashStarted) >= frequency.FlashDuration {
			fs.IsFlashing = false
			stateChanged = true
		}
	} else {
		// Check if it's time for a new flash
		if now.Sub(fs.LastFlashTime) >= frequency.Interval {
			fs.IsFlashing = true
			fs.FlashStarted = now
			fs.LastFlashTime = now
			stateChanged = true
		}
	}

	return stateChanged
}

// ShouldFlash returns whether the element should currently be in flash state.
func (fs *FlashState) ShouldFlash() bool {
	return fs.IsFlashing
}

// CursorState tracks the current state of the cursor blink.
type CursorState struct {
	LastBlinkTime time.Time
	Visible       bool
}

// NewCursorState creates a new cursor state tracker.
func NewCursorState() *CursorState {
	return &CursorState{
		LastBlinkTime: time.Now(),
		Visible:       true, // Start visible
	}
}

// Update updates the cursor state based on current time and blink rate.
// Returns true if cursor visibility changed.
func (cs *CursorState) Update(now time.Time, blinkRate CursorBlinkRate) bool {
	if now.Sub(cs.LastBlinkTime) >= blinkRate.Interval {
		cs.Visible = !cs.Visible
		cs.LastBlinkTime = now
		return true
	}
	return false
}

// IsVisible returns whether the cursor should currently be visible.
func (cs *CursorState) IsVisible() bool {
	return cs.Visible
}

// GetTickInterval returns the recommended BubbleTea tick interval for smooth effects.
// This should be used to set up the tick command in the TUI model.
//
// The tick interval is set to be fast enough to handle the fastest effect updates,
// while not being so fast as to waste CPU cycles.
func GetTickInterval(san int) time.Duration {
	frequency := CalculateScreenFlashFrequency(san)
	cursorBlink := CalculateCursorBlinkRate(san)

	// Use the faster of the two as the base tick rate
	minInterval := frequency.FlashDuration
	if cursorBlink.Interval < minInterval {
		minInterval = cursorBlink.Interval
	}

	// Tick at 1/10 of the minimum interval for smooth transitions
	// But no faster than 16ms (60 FPS)
	tickInterval := minInterval / 10
	if tickInterval < 16*time.Millisecond {
		tickInterval = 16 * time.Millisecond
	}

	// And no slower than 100ms for responsiveness
	if tickInterval > 100*time.Millisecond {
		tickInterval = 100 * time.Millisecond
	}

	return tickInterval
}

// FlashEffect represents a visual flash effect that can be applied to text or borders.
type FlashEffect struct {
	Active    bool
	Intensity float64 // 0.0 (no effect) to 1.0 (full flash)
}

// CalculateFlashIntensity calculates the current flash intensity based on timing.
// This creates a fade-in/fade-out effect rather than a hard on/off.
func CalculateFlashIntensity(flashStarted time.Time, now time.Time, duration time.Duration) float64 {
	elapsed := now.Sub(flashStarted)
	if elapsed >= duration {
		return 0.0 // Flash complete
	}

	// Calculate progress (0.0 to 1.0)
	progress := float64(elapsed) / float64(duration)

	// Use a triangle wave for fade in/out
	// 0-50%: fade in (0.0 -> 1.0)
	// 50-100%: fade out (1.0 -> 0.0)
	if progress < 0.5 {
		// Fade in
		return progress * 2.0
	}

	// Fade out
	return 2.0 - (progress * 2.0)
}

// GetFlashEffect returns the current flash effect state.
func GetFlashEffect(state *FlashState, now time.Time, frequency BlinkFrequency) FlashEffect {
	if !state.IsFlashing {
		return FlashEffect{
			Active:    false,
			Intensity: 0.0,
		}
	}

	intensity := CalculateFlashIntensity(state.FlashStarted, now, frequency.FlashDuration)

	return FlashEffect{
		Active:    true,
		Intensity: intensity,
	}
}
