package effects

import (
	"sync"
	"testing"
	"time"
)

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()

	if eb == nil {
		t.Fatal("Expected EventBus to be created")
	}
	if eb.syncHandlers == nil {
		t.Error("Expected syncHandlers to be initialized")
	}
	if eb.running {
		t.Error("Expected EventBus to not be running initially")
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	eb := NewEventBus()

	handler := func(e Event) {
		// Handler function
	}

	eb.Subscribe(EventSANChanged, handler)

	if len(eb.syncHandlers[EventSANChanged]) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(eb.syncHandlers[EventSANChanged]))
	}
}

func TestEventBus_EmitSync(t *testing.T) {
	eb := NewEventBus()

	receivedEvent := Event{}
	handler := func(e Event) {
		receivedEvent = e
	}

	eb.Subscribe(EventSANChanged, handler)

	testEvent := Event{
		Type:     EventSANChanged,
		Priority: P1High, // Sync
		Data:     "test data",
	}

	eb.Emit(testEvent)

	if receivedEvent.Type != EventSANChanged {
		t.Errorf("Expected event type %s, got %s", EventSANChanged, receivedEvent.Type)
	}
	if receivedEvent.Data != "test data" {
		t.Errorf("Expected data 'test data', got %v", receivedEvent.Data)
	}
}

func TestEventBus_EmitAsync(t *testing.T) {
	eb := NewEventBus()
	eb.Start()
	defer eb.Stop()

	receivedEvent := Event{}
	var mu sync.Mutex
	done := make(chan bool, 1)

	handler := func(e Event) {
		mu.Lock()
		receivedEvent = e
		mu.Unlock()
		done <- true
	}

	eb.Subscribe(EventSANChanged, handler)

	testEvent := Event{
		Type:     EventSANChanged,
		Priority: P2Medium, // Async
		Data:     "async test",
	}

	eb.Emit(testEvent)

	// Wait for async processing
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for async event")
	}

	mu.Lock()
	if receivedEvent.Type != EventSANChanged {
		t.Errorf("Expected event type %s, got %s", EventSANChanged, receivedEvent.Type)
	}
	mu.Unlock()
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	eb := NewEventBus()

	callCount := 0
	var mu sync.Mutex

	handler1 := func(e Event) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	handler2 := func(e Event) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	eb.Subscribe(EventSANChanged, handler1)
	eb.Subscribe(EventSANChanged, handler2)

	testEvent := Event{
		Type:     EventSANChanged,
		Priority: P1High,
	}

	eb.Emit(testEvent)

	mu.Lock()
	if callCount != 2 {
		t.Errorf("Expected 2 handler calls, got %d", callCount)
	}
	mu.Unlock()
}

func TestEventBus_EmitSANChange(t *testing.T) {
	eb := NewEventBus()

	var receivedData interface{}
	handler := func(e Event) {
		receivedData = e.Data
	}

	eb.Subscribe(EventSANChanged, handler)

	eb.EmitSANChange(80, 50)

	// Wait for throttler
	time.Sleep(150 * time.Millisecond)

	if receivedData == nil {
		t.Fatal("Expected to receive SAN change event")
	}

	// Data should be SANChangeWithStyle
	data, ok := receivedData.(SANChangeWithStyle)
	if !ok {
		t.Fatalf("Expected SANChangeWithStyle type, got %T", receivedData)
	}

	if data.OldSAN != 80 || data.NewSAN != 50 {
		t.Errorf("Expected SAN change 80->50, got %d->%d", data.OldSAN, data.NewSAN)
	}

	// HorrorStyle should be calculated for newSAN=50
	expectedStyle := CalculateHorrorStyle(50)
	if data.Style.TextCorruption != expectedStyle.TextCorruption {
		t.Errorf("Expected HorrorStyle for SAN=50, got different style")
	}
}

func TestEventBus_StartStop(t *testing.T) {
	eb := NewEventBus()

	if eb.running {
		t.Error("Expected EventBus to not be running initially")
	}

	eb.Start()

	if !eb.running {
		t.Error("Expected EventBus to be running after Start()")
	}

	eb.Stop()

	if eb.running {
		t.Error("Expected EventBus to not be running after Stop()")
	}
}

func TestThrottler_ImmediateCall(t *testing.T) {
	throttler := NewThrottler(100 * time.Millisecond)

	called := false
	throttler.Throttle(func() {
		called = true
	})

	if !called {
		t.Error("Expected function to be called immediately on first throttle")
	}
}

func TestThrottler_ThrottledCall(t *testing.T) {
	throttler := NewThrottler(100 * time.Millisecond)

	callCount := 0

	// First call - immediate
	throttler.Throttle(func() {
		callCount++
	})

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	// Second call within interval - should be throttled
	throttler.Throttle(func() {
		callCount++
	})

	// Should still be 1 (not executed yet)
	if callCount != 1 {
		t.Errorf("Expected call to be throttled, got %d calls", callCount)
	}

	// Wait for throttle interval
	time.Sleep(150 * time.Millisecond)

	// Now should be 2
	if callCount != 2 {
		t.Errorf("Expected 2 calls after throttle interval, got %d", callCount)
	}
}

func TestThrottler_LatestCallWins(t *testing.T) {
	throttler := NewThrottler(100 * time.Millisecond)

	result := 0

	// First call - immediate
	throttler.Throttle(func() {
		result = 1
	})

	// Multiple rapid calls - only last one should execute
	throttler.Throttle(func() {
		result = 2
	})
	throttler.Throttle(func() {
		result = 3
	})
	throttler.Throttle(func() {
		result = 4
	})

	// Wait for throttle
	time.Sleep(150 * time.Millisecond)

	// Should be 4 (latest call)
	if result != 4 {
		t.Errorf("Expected result=4 (latest call), got %d", result)
	}
}

func TestTransitionState_Immediate(t *testing.T) {
	from := CalculateHorrorStyle(80)
	to := CalculateHorrorStyle(20)

	transition := NewTransitionState(from, to, 500*time.Millisecond)

	// At start, should return 'from' state
	current := transition.GetCurrent(transition.StartTime)

	if current.TextCorruption != from.TextCorruption {
		t.Errorf("Expected TextCorruption=%.2f at start, got %.2f",
			from.TextCorruption, current.TextCorruption)
	}
}

func TestTransitionState_Complete(t *testing.T) {
	from := CalculateHorrorStyle(80)
	to := CalculateHorrorStyle(20)

	transition := NewTransitionState(from, to, 500*time.Millisecond)

	// After duration, should return 'to' state
	futureTime := transition.StartTime.Add(600 * time.Millisecond)
	current := transition.GetCurrent(futureTime)

	if current.TextCorruption != to.TextCorruption {
		t.Errorf("Expected TextCorruption=%.2f after transition, got %.2f",
			to.TextCorruption, current.TextCorruption)
	}

	if !transition.IsComplete(futureTime) {
		t.Error("Expected transition to be complete")
	}
}

func TestTransitionState_Midpoint(t *testing.T) {
	from := HorrorStyle{TextCorruption: 0.0}
	to := HorrorStyle{TextCorruption: 1.0}

	transition := NewTransitionState(from, to, 500*time.Millisecond)

	// At 50% progress (250ms), should be around 0.5 (with easing)
	midTime := transition.StartTime.Add(250 * time.Millisecond)
	current := transition.GetCurrent(midTime)

	// With ease-in-out cubic, midpoint should be close to 0.5
	if current.TextCorruption < 0.3 || current.TextCorruption > 0.7 {
		t.Errorf("Expected TextCorruption around 0.5 at midpoint, got %.2f", current.TextCorruption)
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		from     float64
		to       float64
		t        float64
		expected float64
	}{
		{0.0, 1.0, 0.0, 0.0},
		{0.0, 1.0, 0.5, 0.5},
		{0.0, 1.0, 1.0, 1.0},
		{10.0, 20.0, 0.5, 15.0},
	}

	for _, tt := range tests {
		result := lerp(tt.from, tt.to, tt.t)
		if result != tt.expected {
			t.Errorf("lerp(%.2f, %.2f, %.2f) = %.2f, want %.2f",
				tt.from, tt.to, tt.t, result, tt.expected)
		}
	}
}

func TestLerpInt(t *testing.T) {
	tests := []struct {
		from     int
		to       int
		t        float64
		expected int
	}{
		{0, 100, 0.0, 0},
		{0, 100, 0.5, 50},
		{0, 100, 1.0, 100},
		{10, 20, 0.5, 15},
	}

	for _, tt := range tests {
		result := lerpInt(tt.from, tt.to, tt.t)
		if result != tt.expected {
			t.Errorf("lerpInt(%d, %d, %.2f) = %d, want %d",
				tt.from, tt.to, tt.t, result, tt.expected)
		}
	}
}

func TestEaseInOutCubic(t *testing.T) {
	tests := []struct {
		t        float64
		expected float64
		delta    float64
	}{
		{0.0, 0.0, 0.01},   // Start
		{0.5, 0.5, 0.1},    // Midpoint should be around 0.5
		{1.0, 1.0, 0.01},   // End
		{0.25, 0.0625, 0.01}, // Quarter point (ease-in) - 4 * 0.25^3
		{0.75, 0.9375, 0.01}, // Three-quarter point (ease-out)
	}

	for _, tt := range tests {
		result := easeInOutCubic(tt.t)
		if result < tt.expected-tt.delta || result > tt.expected+tt.delta {
			t.Errorf("easeInOutCubic(%.2f) = %.3f, want %.3f ±%.3f",
				tt.t, result, tt.expected, tt.delta)
		}
	}
}

func TestTransitionState_AllFields(t *testing.T) {
	from := HorrorStyle{
		TextCorruption:    0.0,
		TypingBehavior:    0.0,
		ColorShift:        0,
		UIStability:       0,
		OptionReliability: 1.0,
	}

	to := HorrorStyle{
		TextCorruption:    1.0,
		TypingBehavior:    0.2,
		ColorShift:        180,
		UIStability:       5,
		OptionReliability: 0.0,
	}

	transition := NewTransitionState(from, to, 500*time.Millisecond)

	// Check at 100% completion
	futureTime := transition.StartTime.Add(500 * time.Millisecond)
	current := transition.GetCurrent(futureTime)

	if current.TextCorruption != to.TextCorruption {
		t.Errorf("TextCorruption not interpolated correctly")
	}
	if current.TypingBehavior != to.TypingBehavior {
		t.Errorf("TypingBehavior not interpolated correctly")
	}
	if current.ColorShift != to.ColorShift {
		t.Errorf("ColorShift not interpolated correctly")
	}
	if current.UIStability != to.UIStability {
		t.Errorf("UIStability not interpolated correctly")
	}
	if current.OptionReliability != to.OptionReliability {
		t.Errorf("OptionReliability not interpolated correctly")
	}
}
