package effects

import (
	"testing"
	"time"
)

func TestCalculateScreenFlashFrequency_AllRanges(t *testing.T) {
	tests := []struct {
		san              int
		expectDisabled   bool
		minInterval      time.Duration
		maxInterval      time.Duration
		flashDurationMin time.Duration
	}{
		{100, true, 0, 0, 0},                            // No flashing
		{80, true, 0, 0, 0},                             // No flashing
		{75, false, 5 * time.Second, 10 * time.Second, 100 * time.Millisecond}, // AC1: 5-10s
		{50, false, 3 * time.Second, 5 * time.Second, 150 * time.Millisecond},  // AC2: 3-5s
		{30, false, 1 * time.Second, 2 * time.Second, 200 * time.Millisecond},  // AC3: 1-2s
		{10, false, 500 * time.Millisecond, 1 * time.Second, 250 * time.Millisecond}, // AC4: 0.5-1s
	}

	for _, tt := range tests {
		result := CalculateScreenFlashFrequency(tt.san)

		if tt.expectDisabled {
			if result.Interval != 0 {
				t.Errorf("CalculateScreenFlashFrequency(%d) interval = %v, want 0 (disabled)", tt.san, result.Interval)
			}
		} else {
			if result.Interval < tt.minInterval || result.Interval > tt.maxInterval {
				t.Errorf("CalculateScreenFlashFrequency(%d) interval = %v, want between %v and %v",
					tt.san, result.Interval, tt.minInterval, tt.maxInterval)
			}
			if result.FlashDuration < tt.flashDurationMin {
				t.Errorf("CalculateScreenFlashFrequency(%d) flash duration = %v, want >= %v",
					tt.san, result.FlashDuration, tt.flashDurationMin)
			}
		}
	}
}

func TestCalculateCursorBlinkRate_AllRanges(t *testing.T) {
	tests := []struct {
		san         int
		maxInterval time.Duration
		description string
	}{
		{100, 600 * time.Millisecond, "normal rate"},
		{50, 600 * time.Millisecond, "normal rate"},
		{40, 600 * time.Millisecond, "normal rate"},
		{30, 300 * time.Millisecond, "2x faster (AC3)"},
		{20, 300 * time.Millisecond, "2x faster (AC3)"},
		{15, 200 * time.Millisecond, "very fast (AC4)"},
		{5, 200 * time.Millisecond, "very fast (AC4)"},
	}

	for _, tt := range tests {
		result := CalculateCursorBlinkRate(tt.san)

		if result.Interval > tt.maxInterval {
			t.Errorf("CalculateCursorBlinkRate(%d) = %v, want <= %v (%s)",
				tt.san, result.Interval, tt.maxInterval, tt.description)
		}
	}
}

func TestFlashState_NoFlashing(t *testing.T) {
	state := NewFlashState()
	now := time.Now()

	frequency := BlinkFrequency{
		Interval:      0, // Disabled
		FlashDuration: 0,
	}

	changed := state.Update(now, frequency)

	if changed {
		t.Error("Expected no state change when flashing is disabled")
	}
	if state.ShouldFlash() {
		t.Error("Expected flash to be disabled")
	}
}

func TestFlashState_TriggerFlash(t *testing.T) {
	state := NewFlashState()
	state.LastFlashTime = time.Now().Add(-10 * time.Second) // Long time ago

	frequency := BlinkFrequency{
		Interval:      5 * time.Second,
		FlashDuration: 200 * time.Millisecond,
	}

	now := time.Now()
	changed := state.Update(now, frequency)

	if !changed {
		t.Error("Expected state to change when triggering flash")
	}
	if !state.ShouldFlash() {
		t.Error("Expected flash to be active")
	}
}

func TestFlashState_FlashDuration(t *testing.T) {
	state := NewFlashState()
	frequency := BlinkFrequency{
		Interval:      1 * time.Second,
		FlashDuration: 100 * time.Millisecond,
	}

	// Trigger flash
	state.LastFlashTime = time.Now().Add(-2 * time.Second)
	startTime := time.Now()
	state.Update(startTime, frequency)

	if !state.ShouldFlash() {
		t.Fatal("Expected flash to be active")
	}

	// Advance time past flash duration
	endTime := startTime.Add(150 * time.Millisecond)
	changed := state.Update(endTime, frequency)

	if !changed {
		t.Error("Expected state to change when flash ends")
	}
	if state.ShouldFlash() {
		t.Error("Expected flash to be inactive after duration elapsed")
	}
}

func TestCursorState_InitialState(t *testing.T) {
	state := NewCursorState()

	if !state.IsVisible() {
		t.Error("Expected cursor to start visible")
	}
}

func TestCursorState_BlinkToggle(t *testing.T) {
	state := NewCursorState()
	blinkRate := CursorBlinkRate{
		Interval: 100 * time.Millisecond,
	}

	initialVisible := state.IsVisible()

	// Advance time past blink interval
	futureTime := time.Now().Add(150 * time.Millisecond)
	changed := state.Update(futureTime, blinkRate)

	if !changed {
		t.Error("Expected state to change after blink interval")
	}
	if state.IsVisible() == initialVisible {
		t.Error("Expected cursor visibility to toggle")
	}
}

func TestCursorState_NoChangeBeforeInterval(t *testing.T) {
	state := NewCursorState()
	blinkRate := CursorBlinkRate{
		Interval: 500 * time.Millisecond,
	}

	// Advance time, but not enough to trigger blink
	futureTime := time.Now().Add(100 * time.Millisecond)
	changed := state.Update(futureTime, blinkRate)

	if changed {
		t.Error("Expected no state change before blink interval elapsed")
	}
}

func TestGetTickInterval_AllRanges(t *testing.T) {
	tests := []struct {
		san         int
		minInterval time.Duration
		maxInterval time.Duration
	}{
		{100, 16 * time.Millisecond, 100 * time.Millisecond},
		{70, 16 * time.Millisecond, 100 * time.Millisecond},
		{50, 16 * time.Millisecond, 100 * time.Millisecond},
		{30, 16 * time.Millisecond, 100 * time.Millisecond},
		{10, 16 * time.Millisecond, 100 * time.Millisecond},
	}

	for _, tt := range tests {
		result := GetTickInterval(tt.san)

		if result < tt.minInterval {
			t.Errorf("GetTickInterval(%d) = %v, want >= %v", tt.san, result, tt.minInterval)
		}
		if result > tt.maxInterval {
			t.Errorf("GetTickInterval(%d) = %v, want <= %v", tt.san, result, tt.maxInterval)
		}
	}
}

func TestCalculateFlashIntensity_FadeInOut(t *testing.T) {
	flashStarted := time.Now()
	duration := 200 * time.Millisecond

	tests := []struct {
		elapsed         time.Duration
		expectedMin     float64
		expectedMax     float64
		description     string
	}{
		{0 * time.Millisecond, 0.0, 0.1, "start of flash"},
		{50 * time.Millisecond, 0.4, 0.6, "25% into flash (fade in)"},
		{100 * time.Millisecond, 0.9, 1.0, "50% into flash (peak)"},
		{150 * time.Millisecond, 0.4, 0.6, "75% into flash (fade out)"},
		{200 * time.Millisecond, 0.0, 0.1, "end of flash"},
		{250 * time.Millisecond, 0.0, 0.0, "past flash duration"},
	}

	for _, tt := range tests {
		now := flashStarted.Add(tt.elapsed)
		intensity := CalculateFlashIntensity(flashStarted, now, duration)

		if intensity < tt.expectedMin || intensity > tt.expectedMax {
			t.Errorf("%s: intensity = %.2f, want %.2f-%.2f",
				tt.description, intensity, tt.expectedMin, tt.expectedMax)
		}
	}
}

func TestGetFlashEffect_Inactive(t *testing.T) {
	state := NewFlashState()
	state.IsFlashing = false

	frequency := BlinkFrequency{
		Interval:      1 * time.Second,
		FlashDuration: 200 * time.Millisecond,
	}

	effect := GetFlashEffect(state, time.Now(), frequency)

	if effect.Active {
		t.Error("Expected flash effect to be inactive")
	}
	if effect.Intensity != 0.0 {
		t.Errorf("Expected intensity to be 0.0, got %.2f", effect.Intensity)
	}
}

func TestGetFlashEffect_Active(t *testing.T) {
	state := NewFlashState()
	state.IsFlashing = true
	state.FlashStarted = time.Now()

	frequency := BlinkFrequency{
		Interval:      1 * time.Second,
		FlashDuration: 200 * time.Millisecond,
	}

	// Check at 50% progress (peak intensity)
	now := state.FlashStarted.Add(100 * time.Millisecond)
	effect := GetFlashEffect(state, now, frequency)

	if !effect.Active {
		t.Error("Expected flash effect to be active")
	}
	if effect.Intensity < 0.8 {
		t.Errorf("Expected high intensity at peak, got %.2f", effect.Intensity)
	}
}

func TestFlashState_MultipleFlashes(t *testing.T) {
	state := NewFlashState()
	frequency := BlinkFrequency{
		Interval:      500 * time.Millisecond,
		FlashDuration: 100 * time.Millisecond,
	}

	baseTime := time.Now()
	flashCount := 0

	// Simulate 2 seconds of updates
	for elapsed := 0 * time.Millisecond; elapsed < 2*time.Second; elapsed += 50 * time.Millisecond {
		currentTime := baseTime.Add(elapsed)
		changed := state.Update(currentTime, frequency)

		if changed && state.ShouldFlash() {
			flashCount++
		}
	}

	// Should have triggered ~4 flashes in 2 seconds (interval 500ms)
	if flashCount < 3 || flashCount > 5 {
		t.Errorf("Expected ~4 flashes in 2 seconds, got %d", flashCount)
	}
}
