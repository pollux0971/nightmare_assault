package audio

import (
	"log"
	"sync"
	"time"
)

// HeartbeatController manages heartbeat sound effect based on SAN values
type HeartbeatController struct {
	sfxPlayer  *SFXPlayer
	currentBPM int
	isPlaying  bool
	ticker     *time.Ticker
	stopChan   chan bool
	mu         sync.RWMutex
	currentSAN int
}

// NewHeartbeatController creates a new heartbeat controller
func NewHeartbeatController(sfxPlayer *SFXPlayer) *HeartbeatController {
	return &HeartbeatController{
		sfxPlayer:  sfxPlayer,
		currentBPM: 0,
		isPlaying:  false,
		stopChan:   make(chan bool),
		currentSAN: 100,
	}
}

// CalculateHeartbeatInterval calculates the interval between heartbeats based on SAN
func CalculateHeartbeatInterval(san int) time.Duration {
	var bpm int
	switch {
	case san >= 40:
		return 0 // Don't play
	case san >= 30:
		bpm = 60
	case san >= 20:
		bpm = 90
	case san >= 10:
		bpm = 120
	default:
		bpm = 150
	}
	return time.Minute / time.Duration(bpm)
}

// CalculateHeartbeatVolume calculates the volume based on SAN
func CalculateHeartbeatVolume(san int) float64 {
	if san >= 40 {
		return 0.0
	}
	// Volume increases as SAN decreases
	// SAN 39 = 0.033, SAN 10 = 1.0
	volume := (40.0 - float64(san)) / 30.0
	if volume > 1.0 {
		return 1.0
	}
	return volume
}

// getBPMFromSAN returns the BPM for a given SAN value
func getBPMFromSAN(san int) int {
	switch {
	case san >= 40:
		return 0
	case san >= 30:
		return 60
	case san >= 20:
		return 90
	case san >= 10:
		return 120
	default:
		return 150
	}
}

// Start starts the heartbeat playback
func (h *HeartbeatController) Start(san int) {
	h.mu.Lock()

	// If already playing, just return
	if h.isPlaying {
		h.mu.Unlock()
		return
	}

	// Don't start if SAN >= 40
	interval := CalculateHeartbeatInterval(san)
	if interval == 0 {
		h.mu.Unlock()
		return
	}

	h.isPlaying = true
	h.currentSAN = san
	h.currentBPM = getBPMFromSAN(san)
	h.stopChan = make(chan bool)

	h.mu.Unlock()

	log.Printf("[INFO] Heartbeat started at %d BPM (SAN %d)\n", h.currentBPM, san)

	// Start heartbeat loop in goroutine
	go h.heartbeatLoop(interval)
}

// heartbeatLoop runs the heartbeat playback loop
func (h *HeartbeatController) heartbeatLoop(interval time.Duration) {
	h.ticker = time.NewTicker(interval)
	defer h.ticker.Stop()

	for {
		select {
		case <-h.ticker.C:
			h.mu.RLock()
			volume := CalculateHeartbeatVolume(h.currentSAN)
			h.mu.RUnlock()

			// Play heartbeat sound
			if err := h.sfxPlayer.PlaySFX("heartbeat.wav", SFXPriorityHeartbeat); err != nil {
				log.Printf("[WARN] Failed to play heartbeat: %v\n", err)
			}

			// Adjust volume if player supports it
			h.sfxPlayer.SetVolume(volume)

		case <-h.stopChan:
			log.Println("[INFO] Heartbeat stopped")
			return
		}
	}
}

// Stop stops the heartbeat playback
func (h *HeartbeatController) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isPlaying {
		return
	}

	// Signal stop
	close(h.stopChan)
	h.isPlaying = false
	h.currentBPM = 0

	// TODO: Implement 2-second fade out
	// For now, just stop immediately
}

// AdjustBPM adjusts the heartbeat BPM based on new SAN value
func (h *HeartbeatController) AdjustBPM(san int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isPlaying {
		return
	}

	newBPM := getBPMFromSAN(san)
	if newBPM == 0 {
		// SAN recovered, stop will be called separately
		return
	}

	if newBPM != h.currentBPM {
		h.currentBPM = newBPM
		h.currentSAN = san

		// Restart ticker with new interval
		if h.ticker != nil {
			h.ticker.Stop()
		}
		newInterval := CalculateHeartbeatInterval(san)
		h.ticker = time.NewTicker(newInterval)

		log.Printf("[INFO] Heartbeat BPM adjusted to %d (SAN %d)\n", newBPM, san)
	}
}

// OnSANChange handles SAN change events from EventBus
func (h *HeartbeatController) OnSANChange(newSAN int) {
	h.mu.RLock()
	isPlaying := h.isPlaying
	h.mu.RUnlock()

	if newSAN < 40 && !isPlaying {
		// Start heartbeat
		h.Start(newSAN)
	} else if newSAN >= 40 && isPlaying {
		// Stop heartbeat
		h.Stop()
	} else if isPlaying {
		// Adjust BPM and volume
		h.AdjustBPM(newSAN)
	}
}

// IsPlaying returns whether the heartbeat is currently playing
func (h *HeartbeatController) IsPlaying() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isPlaying
}

// GetCurrentBPM returns the current BPM
func (h *HeartbeatController) GetCurrentBPM() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentBPM
}
