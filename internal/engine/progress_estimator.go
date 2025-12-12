// Package engine provides story generation and game logic for Nightmare Assault.
package engine

import (
	"math"
	"time"
)

// EstimationState represents the current state of progress estimation.
type EstimationState int

const (
	// StateConnecting is the initial connection state
	StateConnecting EstimationState = iota
	// StateGenerating is when content is being generated
	StateGenerating
	// StateStreaming is when content is being streamed
	StateStreaming
	// StateCompleting is the final stage
	StateCompleting
)

// ProgressEstimator estimates the progress of LLM generation.
// Since we cannot know the exact progress, we use a combination of:
// 1. Time-based progress (40%)
// 2. Stream-based progress (60%)
// With easing to avoid getting stuck at 99%.
type ProgressEstimator struct {
	startTime        time.Time
	expectedDuration time.Duration
	streamBytes      int
	expectedBytes    int
	state            EstimationState
	lastProgress     int
	smoothing        float64 // Smoothing factor for progress updates
}

// NewProgressEstimator creates a new progress estimator.
func NewProgressEstimator(expectedDuration time.Duration) *ProgressEstimator {
	return &ProgressEstimator{
		startTime:        time.Now(),
		expectedDuration: expectedDuration,
		expectedBytes:    2000, // Typical response length in tokens
		state:            StateConnecting,
		smoothing:        0.3,  // Smooth updates by 30%
	}
}

// SetExpectedBytes sets the expected number of bytes for stream estimation.
func (e *ProgressEstimator) SetExpectedBytes(bytes int) {
	if bytes > 0 {
		e.expectedBytes = bytes
	}
}

// SetState updates the current state.
func (e *ProgressEstimator) SetState(state EstimationState) {
	e.state = state
}

// UpdateStreamBytes updates the number of bytes received from stream.
func (e *ProgressEstimator) UpdateStreamBytes(bytes int) {
	e.streamBytes = bytes
}

// GetProgress calculates and returns the estimated progress (0-100).
func (e *ProgressEstimator) GetProgress() int {
	elapsed := time.Since(e.startTime)

	// Calculate time-based progress (40% weight)
	timeProgress := e.calculateTimeProgress(elapsed)

	// Calculate stream-based progress (60% weight)
	streamProgress := e.calculateStreamProgress()

	// Combine with weights
	rawProgress := (timeProgress * 0.4) + (streamProgress * 0.6)

	// Apply state-based adjustments
	adjustedProgress := e.applyStateAdjustment(rawProgress)

	// Apply easing function to avoid sticking at 99%
	easedProgress := e.applyEasing(adjustedProgress)

	// Smooth the progress to avoid jumps
	smoothedProgress := e.smoothProgress(easedProgress)

	// Ensure progress is within bounds and monotonic
	progress := int(math.Min(100, math.Max(0, smoothedProgress)))
	if progress < e.lastProgress {
		progress = e.lastProgress
	}
	e.lastProgress = progress

	return progress
}

// calculateTimeProgress calculates progress based on elapsed time.
func (e *ProgressEstimator) calculateTimeProgress(elapsed time.Duration) float64 {
	if e.expectedDuration <= 0 {
		return 0
	}

	ratio := float64(elapsed) / float64(e.expectedDuration)

	// Cap at 95% based on time alone (to prevent premature 100%)
	return math.Min(95, ratio*100)
}

// calculateStreamProgress calculates progress based on streaming bytes.
func (e *ProgressEstimator) calculateStreamProgress() float64 {
	if e.expectedBytes <= 0 || e.streamBytes == 0 {
		return 0
	}

	ratio := float64(e.streamBytes) / float64(e.expectedBytes)

	// Cap at 95% based on stream alone
	return math.Min(95, ratio*100)
}

// applyStateAdjustment adjusts progress based on current state.
func (e *ProgressEstimator) applyStateAdjustment(progress float64) float64 {
	switch e.state {
	case StateConnecting:
		// During connection, limit to 15%
		return math.Min(15, progress)

	case StateGenerating:
		// During generation, range 15-40%
		if progress < 15 {
			return 15
		}
		return math.Min(40, progress)

	case StateStreaming:
		// During streaming, range 40-95%
		if progress < 40 {
			return 40
		}
		return math.Min(95, progress)

	case StateCompleting:
		// Final stage, allow 95-100%
		return math.Min(100, math.Max(95, progress))

	default:
		return progress
	}
}

// applyEasing applies an easing function to smooth progress.
// Uses ease-out-cubic for natural deceleration at the end.
func (e *ProgressEstimator) applyEasing(progress float64) float64 {
	if progress >= 100 {
		return 100
	}

	// Normalize to 0-1
	t := progress / 100.0

	// Ease-out-cubic: 1 - (1-t)^3
	eased := 1 - math.Pow(1-t, 3)

	// Convert back to 0-100
	return eased * 100
}

// smoothProgress smooths progress updates to avoid jumps.
func (e *ProgressEstimator) smoothProgress(newProgress float64) float64 {
	if e.lastProgress == 0 {
		return newProgress
	}

	// Linear interpolation with smoothing factor
	return float64(e.lastProgress)*(1-e.smoothing) + newProgress*e.smoothing
}

// GetRemainingTime estimates the remaining time based on current progress.
func (e *ProgressEstimator) GetRemainingTime() time.Duration {
	progress := e.GetProgress()
	if progress <= 0 || progress >= 100 {
		return 0
	}

	elapsed := time.Since(e.startTime)
	total := elapsed * 100 / time.Duration(progress)
	remaining := total - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// GetElapsedTime returns the time elapsed since start.
func (e *ProgressEstimator) GetElapsedTime() time.Duration {
	return time.Since(e.startTime)
}

// Reset resets the estimator to initial state.
func (e *ProgressEstimator) Reset() {
	e.startTime = time.Now()
	e.streamBytes = 0
	e.state = StateConnecting
	e.lastProgress = 0
}

// IsComplete returns whether the progress is complete (100%).
func (e *ProgressEstimator) IsComplete() bool {
	return e.GetProgress() >= 100
}

// ForceComplete forces the progress to 100%.
func (e *ProgressEstimator) ForceComplete() {
	e.state = StateCompleting
	e.lastProgress = 100
}

// GetProgressFloat returns the progress as a float (0.0-1.0).
func (e *ProgressEstimator) GetProgressFloat() float64 {
	return float64(e.GetProgress()) / 100.0
}

// EstimateFromPromptLength estimates expected bytes from prompt length.
// This is a heuristic based on typical LLM response patterns.
func EstimateFromPromptLength(promptLength int) int {
	// Rule of thumb: response is typically 0.5-2x the prompt length
	// Use conservative estimate (1.5x)
	estimated := int(float64(promptLength) * 1.5)

	// Clamp to reasonable bounds
	if estimated < 500 {
		estimated = 500
	}
	if estimated > 8000 {
		estimated = 8000
	}

	return estimated
}

// EstimateDurationFromLength estimates generation time from expected length.
// Based on typical LLM token generation speeds.
func EstimateDurationFromLength(expectedBytes int) time.Duration {
	// Assume ~50 tokens/second generation speed
	// Roughly 4 bytes per token
	tokens := expectedBytes / 4
	seconds := tokens / 50

	// Add connection overhead (2 seconds)
	seconds += 2

	// Clamp to reasonable bounds (5-30 seconds)
	if seconds < 5 {
		seconds = 5
	}
	if seconds > 30 {
		seconds = 30
	}

	return time.Duration(seconds) * time.Second
}
