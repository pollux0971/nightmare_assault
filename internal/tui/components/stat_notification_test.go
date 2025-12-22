package components

import (
	"testing"
	"time"
)

func TestStatChangeNotification(t *testing.T) {
	t.Run("HP notification creation", func(t *testing.T) {
		notif := NewStatChangeNotification("HP", -25, 75, "遭遇恐怖場景")

		if notif.StatType != "HP" {
			t.Errorf("Expected StatType HP, got %s", notif.StatType)
		}
		if notif.Delta != -25 {
			t.Errorf("Expected Delta -25, got %d", notif.Delta)
		}
		if notif.NewValue != 75 {
			t.Errorf("Expected NewValue 75, got %d", notif.NewValue)
		}
		if notif.Reason != "遭遇恐怖場景" {
			t.Errorf("Expected Reason '遭遇恐怖場景', got %s", notif.Reason)
		}
		if notif.State != NotificationFadeIn {
			t.Errorf("Expected initial state FadeIn, got %v", notif.State)
		}
	})

	t.Run("SAN notification creation", func(t *testing.T) {
		notif := NewStatChangeNotification("SAN", -15, 85, "目睹隊友死亡")

		if notif.StatType != "SAN" {
			t.Errorf("Expected StatType SAN, got %s", notif.StatType)
		}
		if notif.Delta != -15 {
			t.Errorf("Expected Delta -15, got %d", notif.Delta)
		}
	})

	t.Run("notification state transitions", func(t *testing.T) {
		notif := NewStatChangeNotification("HP", -10, 90, "Test")
		now := notif.StartTime

		// Should start in FadeIn
		if notif.State != NotificationFadeIn {
			t.Errorf("Expected FadeIn state, got %v", notif.State)
		}

		// After fade-in duration, should be in Hold
		notif.Update(now.Add(fadeInDuration + 10*time.Millisecond))
		if notif.State != NotificationHold {
			t.Errorf("Expected Hold state after fade-in, got %v", notif.State)
		}

		// After hold duration, should be in FadeOut
		notif.Update(now.Add(fadeInDuration + holdDuration + 10*time.Millisecond))
		if notif.State != NotificationFadeOut {
			t.Errorf("Expected FadeOut state after hold, got %v", notif.State)
		}

		// After fade-out duration, should be Complete
		notif.Update(now.Add(fadeInDuration + holdDuration + fadeOutDuration + 10*time.Millisecond))
		if notif.State != NotificationComplete {
			t.Errorf("Expected Complete state after fade-out, got %v", notif.State)
		}

		if !notif.IsComplete() {
			t.Error("Expected notification to be complete")
		}
	})

	t.Run("opacity calculation", func(t *testing.T) {
		notif := NewStatChangeNotification("HP", -10, 90, "Test")
		now := notif.StartTime

		// At start, opacity should be near 0
		opacity := notif.GetOpacity(now)
		if opacity > 0.1 {
			t.Errorf("Expected low opacity at start, got %f", opacity)
		}

		// Update to hold phase
		holdTime := now.Add(fadeInDuration + 100*time.Millisecond)
		notif.Update(holdTime)
		opacity = notif.GetOpacity(holdTime)
		if opacity < 0.9 {
			t.Errorf("Expected high opacity during hold, got %f", opacity)
		}

		// Update to fadeout phase
		fadeOutTime := now.Add(fadeInDuration + holdDuration + 100*time.Millisecond)
		notif.Update(fadeOutTime)
		if notif.State != NotificationFadeOut {
			t.Errorf("Expected FadeOut state, got %v", notif.State)
		}

		// Update to complete phase
		completeTime := now.Add(fadeInDuration + holdDuration + fadeOutDuration + 10*time.Millisecond)
		notif.Update(completeTime)
		if !notif.IsComplete() {
			t.Errorf("Expected notification to be complete, state: %v", notif.State)
		}
		opacity = notif.GetOpacity(completeTime)
		if opacity != 0.0 {
			t.Errorf("Expected zero opacity after complete, got %f", opacity)
		}
	})
}

func TestNotificationQueue(t *testing.T) {
	t.Run("queue creation", func(t *testing.T) {
		queue := NewNotificationQueue(3)

		if queue.maxVisible != 3 {
			t.Errorf("Expected maxVisible 3, got %d", queue.maxVisible)
		}
		if queue.HasActive() {
			t.Error("Expected empty queue to have no active notifications")
		}
	})

	t.Run("add and update notifications", func(t *testing.T) {
		queue := NewNotificationQueue(3)

		notif1 := NewStatChangeNotification("HP", -10, 90, "Test 1")
		notif2 := NewStatChangeNotification("SAN", -5, 95, "Test 2")

		queue.Add(notif1)
		queue.Add(notif2)

		if !queue.HasActive() {
			t.Error("Expected queue to have active notifications")
		}

		// Update and verify notifications progress
		now := time.Now()
		queue.Update(now)

		if len(queue.notifications) != 2 {
			t.Errorf("Expected 2 notifications, got %d", len(queue.notifications))
		}

		// Update through all animation phases multiple times to ensure state changes
		fadeInTime := now.Add(fadeInDuration + 10*time.Millisecond)
		queue.Update(fadeInTime)

		holdTime := now.Add(fadeInDuration + holdDuration + 10*time.Millisecond)
		queue.Update(holdTime)

		fadeOutTime := now.Add(fadeInDuration + holdDuration + fadeOutDuration/2)
		queue.Update(fadeOutTime)

		// Update after all animations complete
		futureTime := now.Add(fadeInDuration + holdDuration + fadeOutDuration + 100*time.Millisecond)
		queue.Update(futureTime)

		// All notifications should be removed after completion
		if queue.HasActive() {
			t.Errorf("Expected queue to have no active notifications after completion, but has %d", len(queue.notifications))
		}
	})

	t.Run("clear queue", func(t *testing.T) {
		queue := NewNotificationQueue(3)

		queue.Add(NewStatChangeNotification("HP", -10, 90, "Test"))
		queue.Add(NewStatChangeNotification("SAN", -5, 95, "Test"))

		if !queue.HasActive() {
			t.Error("Expected queue to have active notifications")
		}

		queue.Clear()

		if queue.HasActive() {
			t.Error("Expected queue to be empty after clear")
		}
	})

	t.Run("max visible limit", func(t *testing.T) {
		queue := NewNotificationQueue(2)

		// Add 4 notifications
		for i := 0; i < 4; i++ {
			queue.Add(NewStatChangeNotification("HP", -10, 90, "Test"))
		}

		// All should be stored
		if len(queue.notifications) != 4 {
			t.Errorf("Expected 4 notifications stored, got %d", len(queue.notifications))
		}

		// But rendering should respect maxVisible (tested via Render, which is hard to test without lipgloss)
		// This is a basic structural test
	})
}
