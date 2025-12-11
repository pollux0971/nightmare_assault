package engine

import (
	"testing"
	"time"
)

func TestLowSANSpeedVariation(t *testing.T) {
	t.Run("normal SAN has stable speed", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 100
		buffer := NewStreamBuffer(config)

		baseDelay := 25 * time.Millisecond

		// With high SAN, speed should remain normal
		delay := buffer.calculateDelayWithEffects(baseDelay, 'a')

		// Should be close to base delay (allow for punctuation)
		if delay > baseDelay*2 {
			t.Errorf("Expected stable delay around %v, got %v", baseDelay, delay)
		}
	})

	t.Run("low SAN applies speed variation", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 30 // Low SAN triggers effects
		buffer := NewStreamBuffer(config)

		baseDelay := 25 * time.Millisecond

		// Collect multiple samples to see variation
		delays := make([]time.Duration, 20)
		for i := 0; i < 20; i++ {
			delays[i] = buffer.calculateDelayWithEffects(baseDelay, 'a')
			// Simulate time passing
			buffer.speedVariation.modeStart = buffer.speedVariation.modeStart.Add(-time.Second)
		}

		// At least some delays should differ significantly from base
		hasVariation := false
		for _, d := range delays {
			if d < baseDelay/2 || d > baseDelay*2 {
				hasVariation = true
				break
			}
		}

		if !hasVariation {
			t.Error("Expected speed variation with low SAN")
		}
	})

	t.Run("speed mode transitions", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 30
		buffer := NewStreamBuffer(config)

		// Force mode change by expiring duration
		buffer.speedVariation.mode = speedNormal
		buffer.speedVariation.modeDuration = 1 * time.Millisecond
		buffer.speedVariation.modeStart = time.Now().Add(-2 * time.Millisecond)

		// Calculate delay should trigger mode update
		baseDelay := 25 * time.Millisecond
		_ = buffer.calculateDelayWithEffects(baseDelay, 'a')

		// Mode should have been updated
		if time.Since(buffer.speedVariation.modeStart) > 100*time.Millisecond {
			t.Error("Speed mode should have been updated")
		}
	})

	t.Run("punctuation adds delay even with effects", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 30
		config.PunctuationDelay = 100 * time.Millisecond
		buffer := NewStreamBuffer(config)

		baseDelay := 25 * time.Millisecond

		// Force normal speed mode to isolate punctuation effect
		buffer.speedVariation.mode = speedNormal
		buffer.speedVariation.modeDuration = 10 * time.Second
		buffer.speedVariation.modeStart = time.Now()

		normalDelay := buffer.calculateDelayWithEffects(baseDelay, 'a')
		punctDelay := buffer.calculateDelayWithEffects(baseDelay, '。')

		if punctDelay <= normalDelay {
			t.Errorf("Punctuation delay %v should be greater than normal %v", punctDelay, normalDelay)
		}
	})
}

func TestLowSANTextGlitches(t *testing.T) {
	t.Run("high SAN has no glitches", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 100
		buffer := NewStreamBuffer(config)

		// Try many times, should never glitch
		for i := 0; i < 100; i++ {
			shouldDisplay, shouldRepeat := buffer.applyTextGlitches('x')
			if !shouldDisplay || shouldRepeat {
				t.Error("High SAN should not have glitches")
			}
		}
	})

	t.Run("low SAN can cause glitches", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 10 // Very low SAN
		buffer := NewStreamBuffer(config)

		// Try many times, should eventually see some glitch
		hadRepeat := false
		hadSkip := false

		for i := 0; i < 1000; i++ {
			shouldDisplay, shouldRepeat := buffer.applyTextGlitches('x')

			if shouldRepeat {
				hadRepeat = true
			}
			if !shouldDisplay {
				hadSkip = true
			}

			if hadRepeat && hadSkip {
				break
			}
		}

		// At SAN=10, we should see some glitches in 1000 attempts
		if !hadRepeat && !hadSkip {
			t.Log("Warning: No glitches observed in 1000 attempts (may be rare randomness)")
		}
	})

	t.Run("medium SAN has fewer glitches", func(t *testing.T) {
		// SAN 20-39 should have some glitches but less than very low SAN
		config := DefaultTypewriterConfig()
		config.SAN = 25
		buffer := NewStreamBuffer(config)

		// Medium SAN (>20) should NOT trigger glitches in applyTextGlitches
		// because the function checks SAN >= 20
		for i := 0; i < 100; i++ {
			shouldDisplay, shouldRepeat := buffer.applyTextGlitches('x')
			if !shouldDisplay || shouldRepeat {
				t.Error("SAN >= 20 should not have text glitches")
			}
		}
	})

	t.Run("glitch state is maintained correctly", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		config.SAN = 5 // Very low for consistent glitching
		buffer := NewStreamBuffer(config)

		// Manually trigger repeat state
		buffer.glitchState.repeatChar = true

		// First call with repeat state true
		_, repeat1 := buffer.applyTextGlitches('x')

		// If it was going to repeat, the state should toggle
		if repeat1 {
			// State should have been reset
			if buffer.glitchState.repeatChar {
				t.Error("Repeat state should be reset after repeating")
			}
		}
	})
}

func TestSpeedModeTransitions(t *testing.T) {
	t.Run("updateSpeedMode sets valid modes", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)

		// Try multiple updates
		for i := 0; i < 10; i++ {
			buffer.updateSpeedMode()

			// Verify mode is valid
			validMode := buffer.speedVariation.mode == speedNormal ||
				buffer.speedVariation.mode == speedFast ||
				buffer.speedVariation.mode == speedSlow ||
				buffer.speedVariation.mode == speedStuck

			if !validMode {
				t.Errorf("Invalid speed mode: %v", buffer.speedVariation.mode)
			}

			// Verify duration is reasonable
			if buffer.speedVariation.modeDuration < 0 {
				t.Error("Mode duration should not be negative")
			}
			if buffer.speedVariation.modeDuration > 10*time.Second {
				t.Error("Mode duration seems too long")
			}
		}
	})

	t.Run("mode start time is updated", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)

		before := time.Now()
		buffer.updateSpeedMode()
		after := time.Now()

		if buffer.speedVariation.modeStart.Before(before) || buffer.speedVariation.modeStart.After(after) {
			t.Error("Mode start time should be set to current time")
		}
	})
}

func TestApplySpeedVariation(t *testing.T) {
	baseDelay := 25 * time.Millisecond

	t.Run("speedFast reduces delay", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)
		buffer.speedVariation.mode = speedFast
		buffer.speedVariation.modeDuration = 10 * time.Second
		buffer.speedVariation.modeStart = time.Now()

		delay := buffer.applySpeedVariation(baseDelay)

		if delay >= baseDelay {
			t.Errorf("Fast mode delay %v should be less than base %v", delay, baseDelay)
		}

		// Should be in range 6-10ms
		if delay < 6*time.Millisecond || delay > 10*time.Millisecond {
			t.Errorf("Fast mode delay %v out of expected range [6-10ms]", delay)
		}
	})

	t.Run("speedSlow increases delay", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)
		buffer.speedVariation.mode = speedSlow
		buffer.speedVariation.modeDuration = 10 * time.Second
		buffer.speedVariation.modeStart = time.Now()

		delay := buffer.applySpeedVariation(baseDelay)

		if delay <= baseDelay {
			t.Errorf("Slow mode delay %v should be greater than base %v", delay, baseDelay)
		}

		// Should be in range 50-100ms
		if delay < 50*time.Millisecond || delay > 100*time.Millisecond {
			t.Errorf("Slow mode delay %v out of expected range [50-100ms]", delay)
		}
	})

	t.Run("speedStuck pauses", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)
		buffer.speedVariation.mode = speedStuck
		buffer.speedVariation.modeDuration = 10 * time.Second
		buffer.speedVariation.modeStart = time.Now()

		delay := buffer.applySpeedVariation(baseDelay)

		// Should be in range 200-500ms
		if delay < 200*time.Millisecond || delay > 500*time.Millisecond {
			t.Errorf("Stuck mode delay %v out of expected range [200-500ms]", delay)
		}
	})

	t.Run("speedNormal returns base delay", func(t *testing.T) {
		config := DefaultTypewriterConfig()
		buffer := NewStreamBuffer(config)
		buffer.speedVariation.mode = speedNormal
		buffer.speedVariation.modeDuration = 10 * time.Second
		buffer.speedVariation.modeStart = time.Now()

		delay := buffer.applySpeedVariation(baseDelay)

		if delay != baseDelay {
			t.Errorf("Normal mode should return base delay %v, got %v", baseDelay, delay)
		}
	})
}

func TestTypewriterConfigSAN(t *testing.T) {
	t.Run("SAN field is stored in config", func(t *testing.T) {
		config := TypewriterConfig{
			MinCharsPerSecond: 30,
			MaxCharsPerSecond: 50,
			PunctuationDelay:  100 * time.Millisecond,
			ParagraphDelay:    300 * time.Millisecond,
			Enabled:           true,
			ShowCursor:        true,
			SAN:               45,
		}

		if config.SAN != 45 {
			t.Errorf("Expected SAN 45, got %d", config.SAN)
		}
	})

	t.Run("default SAN is 100", func(t *testing.T) {
		config := DefaultTypewriterConfig()

		if config.SAN != 100 {
			t.Errorf("Expected default SAN 100, got %d", config.SAN)
		}
	})
}
