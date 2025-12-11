package effects

import (
	"sync"
	"time"
)

// EventType represents the type of event.
type EventType string

const (
	EventSANChanged    EventType = "san_changed"
	EventHPChanged     EventType = "hp_changed"
	EventDeath         EventType = "death"
	EventRuleTriggered EventType = "rule_triggered"
)

// EventPriority represents the priority level of an event.
type EventPriority int

const (
	P0Critical EventPriority = iota // Critical: Death, game over
	P1High                           // High: SAN/HP changes, immediate UI update
	P2Medium                         // Medium: Background effects
	P3Low                            // Low: Logging, analytics
)

// Event represents a game event.
type Event struct {
	Type     EventType
	Priority EventPriority
	Data     interface{}
}

// SANChangeData contains data for SAN change events.
type SANChangeData struct {
	OldSAN int
	NewSAN int
}

// SANChangeWithStyle extends SANChangeData with the calculated HorrorStyle.
type SANChangeWithStyle struct {
	SANChangeData
	Style HorrorStyle
}

// EventHandler is a function that handles an event.
type EventHandler func(Event)

// EventBus manages event distribution and handling.
type EventBus struct {
	// Synchronous handlers: P0/P1 events, UI must react immediately
	syncHandlers map[EventType][]EventHandler
	syncMu       sync.RWMutex

	// Asynchronous channel: P2/P3 events, background processing
	asyncChan chan Event

	// Throttler for SAN change events (100ms)
	throttler *Throttler

	// Running state
	running bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	eb := &EventBus{
		syncHandlers: make(map[EventType][]EventHandler),
		asyncChan:    make(chan Event, 64), // 64 buffer, drop on overflow
		throttler:    NewThrottler(100 * time.Millisecond),
		running:      false,
		stopChan:     make(chan struct{}),
	}

	return eb
}

// Subscribe registers a handler for an event type.
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.syncMu.Lock()
	defer eb.syncMu.Unlock()

	eb.syncHandlers[eventType] = append(eb.syncHandlers[eventType], handler)
}

// Emit emits an event to all registered handlers.
func (eb *EventBus) Emit(e Event) {
	switch e.Priority {
	case P0Critical, P1High:
		// Synchronous: execute all callbacks immediately
		eb.syncMu.RLock()
		handlers := eb.syncHandlers[e.Type]
		eb.syncMu.RUnlock()

		for _, handler := range handlers {
			handler(e)
		}

	case P2Medium, P3Low:
		// Asynchronous: send to channel (non-blocking)
		select {
		case eb.asyncChan <- e:
			// Event queued
		default:
			// Buffer full, drop event (acceptable for low priority)
		}
	}
}

// EmitSANChange emits a SAN change event with throttling.
// Multiple SAN changes within 100ms are merged into one event.
func (eb *EventBus) EmitSANChange(oldSAN, newSAN int) {
	eb.throttler.Throttle(func() {
		style := CalculateHorrorStyle(newSAN)

		event := Event{
			Type:     EventSANChanged,
			Priority: P1High, // High priority - UI must update immediately
			Data: SANChangeWithStyle{
				SANChangeData: SANChangeData{
					OldSAN: oldSAN,
					NewSAN: newSAN,
				},
				Style: style,
			},
		}

		eb.Emit(event)
	})
}

// Start starts the event bus async worker.
func (eb *EventBus) Start() {
	if eb.running {
		return
	}

	eb.running = true
	eb.wg.Add(1)

	go func() {
		defer eb.wg.Done()

		for {
			select {
			case event := <-eb.asyncChan:
				// Process async event
				eb.syncMu.RLock()
				handlers := eb.syncHandlers[event.Type]
				eb.syncMu.RUnlock()

				for _, handler := range handlers {
					handler(event)
				}

			case <-eb.stopChan:
				return
			}
		}
	}()
}

// Stop stops the event bus.
func (eb *EventBus) Stop() {
	if !eb.running {
		return
	}

	eb.running = false
	close(eb.stopChan)
	eb.wg.Wait()
}

// Throttler throttles function calls to a maximum rate.
type Throttler struct {
	interval    time.Duration
	lastCall    time.Time
	mu          sync.Mutex
	pendingFunc func()
	timer       *time.Timer
}

// NewThrottler creates a new throttler.
func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{
		interval: interval,
		lastCall: time.Time{},
	}
}

// Throttle throttles a function call.
// If called multiple times within the interval, only the last call is executed.
func (t *Throttler) Throttle(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastCall)

	// Store the latest function to call
	t.pendingFunc = fn

	if elapsed >= t.interval {
		// Enough time has passed, execute immediately
		t.lastCall = now
		fn()
		t.pendingFunc = nil
	} else {
		// Too soon, schedule for later
		if t.timer != nil {
			t.timer.Stop()
		}

		remaining := t.interval - elapsed
		t.timer = time.AfterFunc(remaining, func() {
			t.mu.Lock()
			defer t.mu.Unlock()

			if t.pendingFunc != nil {
				t.pendingFunc()
				t.lastCall = time.Now()
				t.pendingFunc = nil
			}
		})
	}
}

// TransitionState manages smooth transitions between HorrorStyle states.
type TransitionState struct {
	From      HorrorStyle
	To        HorrorStyle
	StartTime time.Time
	Duration  time.Duration
}

// NewTransitionState creates a new transition state for smooth HorrorStyle changes.
func NewTransitionState(from, to HorrorStyle, duration time.Duration) *TransitionState {
	return &TransitionState{
		From:      from,
		To:        to,
		StartTime: time.Now(),
		Duration:  duration,
	}
}

// GetCurrent returns the current interpolated HorrorStyle based on elapsed time.
// Uses ease-in-out interpolation for smooth transitions (AC5: 500ms transition).
func (ts *TransitionState) GetCurrent(now time.Time) HorrorStyle {
	elapsed := now.Sub(ts.StartTime)

	if elapsed >= ts.Duration {
		// Transition complete
		return ts.To
	}

	// Calculate progress (0.0 to 1.0)
	progress := float64(elapsed) / float64(ts.Duration)

	// Apply ease-in-out curve
	progress = easeInOutCubic(progress)

	// Interpolate each field
	return HorrorStyle{
		TextCorruption:    lerp(ts.From.TextCorruption, ts.To.TextCorruption, progress),
		TypingBehavior:    lerp(ts.From.TypingBehavior, ts.To.TypingBehavior, progress),
		ColorShift:        lerpInt(ts.From.ColorShift, ts.To.ColorShift, progress),
		UIStability:       lerpInt(ts.From.UIStability, ts.To.UIStability, progress),
		OptionReliability: lerp(ts.From.OptionReliability, ts.To.OptionReliability, progress),
	}
}

// IsComplete returns whether the transition has completed.
func (ts *TransitionState) IsComplete(now time.Time) bool {
	return now.Sub(ts.StartTime) >= ts.Duration
}

// lerp performs linear interpolation between two float64 values.
func lerp(from, to, t float64) float64 {
	return from + (to-from)*t
}

// lerpInt performs linear interpolation between two int values.
func lerpInt(from, to int, t float64) int {
	return int(float64(from) + float64(to-from)*t)
}

// easeInOutCubic applies an ease-in-out cubic interpolation curve.
// This creates smooth acceleration and deceleration.
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	p := 2*t - 2
	return 1 + p*p*p/2
}
