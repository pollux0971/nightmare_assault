package audio

import (
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// TestCalculateHeartbeatInterval tests BPM calculation for different SAN ranges
func TestCalculateHeartbeatInterval(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected time.Duration
	}{
		{
			name:     "SAN >= 40: no heartbeat",
			san:      40,
			expected: 0,
		},
		{
			name:     "SAN 39: no heartbeat",
			san:      50,
			expected: 0,
		},
		{
			name:     "SAN 30-39: 60 BPM",
			san:      35,
			expected: time.Second, // 60 BPM = 1 beat per second
		},
		{
			name:     "SAN 20-29: 90 BPM",
			san:      25,
			expected: 666666666 * time.Nanosecond, // 60s / 90 = 0.666... seconds
		},
		{
			name:     "SAN 10-19: 120 BPM",
			san:      15,
			expected: 500 * time.Millisecond, // 60s / 120 = 0.5 seconds
		},
		{
			name:     "SAN 1-9: 150 BPM",
			san:      5,
			expected: 400 * time.Millisecond, // 60s / 150 = 0.4 seconds
		},
		{
			name:     "SAN 0: 150 BPM",
			san:      0,
			expected: 400 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHeartbeatInterval(tt.san)
			// Allow 1ms tolerance for rounding
			diff := got - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Millisecond {
				t.Errorf("CalculateHeartbeatInterval(%d) = %v, want %v", tt.san, got, tt.expected)
			}
		})
	}
}

// TestCalculateHeartbeatVolume tests volume calculation based on SAN
func TestCalculateHeartbeatVolume(t *testing.T) {
	tests := []struct {
		name     string
		san      int
		expected float64
		minVol   float64
		maxVol   float64
	}{
		{
			name:     "SAN >= 40: volume 0",
			san:      40,
			expected: 0.0,
			minVol:   0.0,
			maxVol:   0.0,
		},
		{
			name:     "SAN 39: low volume",
			san:      39,
			minVol:   0.0,
			maxVol:   0.1,
		},
		{
			name:     "SAN 30: moderate volume",
			san:      30,
			minVol:   0.2,
			maxVol:   0.4,
		},
		{
			name:     "SAN 20: higher volume",
			san:      20,
			minVol:   0.5,
			maxVol:   0.7,
		},
		{
			name:     "SAN 10: max volume",
			san:      10,
			expected: 1.0,
			minVol:   1.0,
			maxVol:   1.0,
		},
		{
			name:     "SAN 5: max volume",
			san:      5,
			expected: 1.0,
			minVol:   1.0,
			maxVol:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHeartbeatVolume(tt.san)

			// Check exact value if provided
			if tt.expected > 0 || tt.san >= 40 {
				if got != tt.expected {
					t.Errorf("CalculateHeartbeatVolume(%d) = %v, want %v", tt.san, got, tt.expected)
				}
			} else {
				// Check range
				if got < tt.minVol || got > tt.maxVol {
					t.Errorf("CalculateHeartbeatVolume(%d) = %v, want between %v and %v", tt.san, got, tt.minVol, tt.maxVol)
				}
			}

			// Volume must be between 0 and 1
			if got < 0.0 || got > 1.0 {
				t.Errorf("CalculateHeartbeatVolume(%d) = %v, must be between 0.0 and 1.0", tt.san, got)
			}
		})
	}
}

// TestHeartbeatControllerStartStop tests basic start/stop functionality
func TestHeartbeatControllerStartStop(t *testing.T) {
	// Create SFX player (without audio context for unit testing)
	cfg := config.AudioConfig{
		SFXVolume:  0.5,
		SFXEnabled: true,
	}
	sfxPlayer := NewSFXPlayer(nil, cfg, "testdata")

	// Create heartbeat controller
	hb := NewHeartbeatController(sfxPlayer)

	// Test: Start with SAN < 40
	hb.Start(30)
	if !hb.IsPlaying() {
		t.Error("Expected heartbeat to be playing after Start(30)")
	}

	// Test: Stop
	hb.Stop()
	// Give it time to stop
	time.Sleep(50 * time.Millisecond)
	if hb.IsPlaying() {
		t.Error("Expected heartbeat to stop after Stop()")
	}

	// Test: Start with SAN >= 40 (should not start)
	hb.Start(50)
	if hb.IsPlaying() {
		t.Error("Expected heartbeat to NOT play when SAN >= 40")
	}
}

// TestHeartbeatControllerAdjustBPM tests dynamic BPM adjustment
func TestHeartbeatControllerAdjustBPM(t *testing.T) {
	// Create SFX player (without audio context for unit testing)
	cfg := config.AudioConfig{
		SFXVolume:  0.5,
		SFXEnabled: true,
	}
	sfxPlayer := NewSFXPlayer(nil, cfg, "testdata")

	// Create heartbeat controller
	hb := NewHeartbeatController(sfxPlayer)

	// Start with SAN 30 (60 BPM)
	hb.Start(30)
	if !hb.IsPlaying() {
		t.Error("Expected heartbeat to be playing")
	}

	initialBPM := hb.GetCurrentBPM()
	if initialBPM != 60 {
		t.Errorf("Expected initial BPM 60, got %d", initialBPM)
	}

	// Adjust to SAN 15 (120 BPM)
	hb.AdjustBPM(15)
	time.Sleep(50 * time.Millisecond) // Give time for adjustment

	newBPM := hb.GetCurrentBPM()
	if newBPM != 120 {
		t.Errorf("Expected adjusted BPM 120, got %d", newBPM)
	}

	// Cleanup
	hb.Stop()
}

// TestHeartbeatControllerOnSANChange tests EventBus integration
func TestHeartbeatControllerOnSANChange(t *testing.T) {
	// Create SFX player (without audio context for unit testing)
	cfg := config.AudioConfig{
		SFXVolume:  0.5,
		SFXEnabled: true,
	}
	sfxPlayer := NewSFXPlayer(nil, cfg, "testdata")

	// Create heartbeat controller
	hb := NewHeartbeatController(sfxPlayer)

	// Test: SAN drops below 40 - should start
	hb.OnSANChange(35)
	time.Sleep(50 * time.Millisecond)
	if !hb.IsPlaying() {
		t.Error("Expected heartbeat to start when SAN < 40")
	}

	// Test: SAN changes while playing - should adjust
	hb.OnSANChange(20)
	time.Sleep(50 * time.Millisecond)
	if !hb.IsPlaying() {
		t.Error("Expected heartbeat to continue playing")
	}
	if hb.GetCurrentBPM() != 90 {
		t.Errorf("Expected BPM 90 for SAN 20, got %d", hb.GetCurrentBPM())
	}

	// Test: SAN recovers >= 40 - should stop
	hb.OnSANChange(45)
	time.Sleep(50 * time.Millisecond)
	if hb.IsPlaying() {
		t.Error("Expected heartbeat to stop when SAN >= 40")
	}

	// Cleanup
	hb.Stop()
}
