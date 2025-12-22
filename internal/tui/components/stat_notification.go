// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// NotificationState represents the current state of the notification animation.
type NotificationState int

const (
	// NotificationFadeIn is the initial fade-in phase.
	NotificationFadeIn NotificationState = iota
	// NotificationHold is the holding phase where notification is fully visible.
	NotificationHold
	// NotificationFadeOut is the fade-out phase.
	NotificationFadeOut
	// NotificationComplete means the notification is done.
	NotificationComplete
)

// Animation durations
const (
	fadeInDuration  = 300 * time.Millisecond
	holdDuration    = 2000 * time.Millisecond
	fadeOutDuration = 500 * time.Millisecond
)

// StatChangeNotification represents a floating notification for HP/SAN changes.
type StatChangeNotification struct {
	// Stat type ("HP" or "SAN")
	StatType string
	// Delta value (+/- change)
	Delta int
	// New value after change
	NewValue int
	// Reason for the change
	Reason string
	// Start time of the notification
	StartTime time.Time
	// Current state
	State NotificationState
}

// NewStatChangeNotification creates a new stat change notification.
func NewStatChangeNotification(statType string, delta, newValue int, reason string) *StatChangeNotification {
	return &StatChangeNotification{
		StatType:  statType,
		Delta:     delta,
		NewValue:  newValue,
		Reason:    reason,
		StartTime: time.Now(),
		State:     NotificationFadeIn,
	}
}

// Update updates the notification state based on elapsed time.
func (n *StatChangeNotification) Update(now time.Time) {
	if n.State == NotificationComplete {
		return
	}

	elapsed := now.Sub(n.StartTime)

	switch n.State {
	case NotificationFadeIn:
		if elapsed >= fadeInDuration {
			n.State = NotificationHold
		}
	case NotificationHold:
		if elapsed >= fadeInDuration+holdDuration {
			n.State = NotificationFadeOut
		}
	case NotificationFadeOut:
		if elapsed >= fadeInDuration+holdDuration+fadeOutDuration {
			n.State = NotificationComplete
		}
	}
}

// IsComplete returns true if the notification animation is complete.
func (n *StatChangeNotification) IsComplete() bool {
	return n.State == NotificationComplete
}

// GetOpacity returns the current opacity (0.0 - 1.0) based on animation state.
func (n *StatChangeNotification) GetOpacity(now time.Time) float64 {
	if n.State == NotificationComplete {
		return 0.0
	}

	elapsed := now.Sub(n.StartTime)

	switch n.State {
	case NotificationFadeIn:
		// Fade in from 0 to 1
		progress := float64(elapsed) / float64(fadeInDuration)
		if progress > 1.0 {
			progress = 1.0
		}
		return progress

	case NotificationHold:
		// Fully visible
		return 1.0

	case NotificationFadeOut:
		// Fade out from 1 to 0
		fadeOutElapsed := elapsed - (fadeInDuration + holdDuration)
		progress := 1.0 - (float64(fadeOutElapsed) / float64(fadeOutDuration))
		if progress < 0.0 {
			progress = 0.0
		}
		return progress

	default:
		return 0.0
	}
}

// Render renders the notification with current animation state.
func (n *StatChangeNotification) Render(now time.Time, width int) string {
	if n.State == NotificationComplete {
		return ""
	}

	theme := themes.GetManager().GetCurrentTheme()
	opacity := n.GetOpacity(now)

	// If opacity is too low, don't render
	if opacity < 0.05 {
		return ""
	}

	// Choose color based on stat type and delta
	var color lipgloss.Color
	var icon string
	if n.StatType == "HP" {
		if n.Delta < 0 {
			color = theme.Colors.Error
			icon = "❤"
		} else {
			color = theme.Colors.Success
			icon = "❤"
		}
	} else { // SAN
		if n.Delta < 0 {
			color = theme.Colors.Warning
			icon = "🧠"
		} else {
			color = theme.Colors.Accent
			icon = "🧠"
		}
	}

	// Format delta with sign
	deltaStr := fmt.Sprintf("%+d", n.Delta)

	// Build notification text
	text := fmt.Sprintf("%s %s %s", icon, n.StatType, deltaStr)
	if n.Reason != "" {
		text = fmt.Sprintf("%s\n%s", text, n.Reason)
	}

	// Apply opacity by adjusting color intensity
	// For simplicity, we'll just adjust the base style
	baseStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Align(lipgloss.Center)

	// Add border for emphasis
	notificationStyle := baseStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 2).
		Width(width - 4)

	// Apply fade effect by adjusting style (simulate with faint when fading)
	if opacity < 0.5 {
		notificationStyle = notificationStyle.Faint(true)
	}

	return notificationStyle.Render(text)
}

// NotificationQueue manages multiple notifications with auto-cleanup.
type NotificationQueue struct {
	notifications []*StatChangeNotification
	maxVisible    int // Maximum number of visible notifications
}

// NewNotificationQueue creates a new notification queue.
func NewNotificationQueue(maxVisible int) *NotificationQueue {
	return &NotificationQueue{
		notifications: make([]*StatChangeNotification, 0),
		maxVisible:    maxVisible,
	}
}

// Add adds a new notification to the queue.
func (q *NotificationQueue) Add(notification *StatChangeNotification) {
	q.notifications = append(q.notifications, notification)
}

// Update updates all notifications and removes completed ones.
func (q *NotificationQueue) Update(now time.Time) {
	// Update all notifications
	for _, n := range q.notifications {
		n.Update(now)
	}

	// Remove completed notifications
	active := make([]*StatChangeNotification, 0)
	for _, n := range q.notifications {
		if !n.IsComplete() {
			active = append(active, n)
		}
	}
	q.notifications = active
}

// Render renders all active notifications as an overlay.
func (q *NotificationQueue) Render(now time.Time, width int) string {
	if len(q.notifications) == 0 {
		return ""
	}

	// Render only the most recent maxVisible notifications
	start := 0
	if len(q.notifications) > q.maxVisible {
		start = len(q.notifications) - q.maxVisible
	}

	var rendered []string
	for i := start; i < len(q.notifications); i++ {
		notifText := q.notifications[i].Render(now, width)
		if notifText != "" {
			rendered = append(rendered, notifText)
		}
	}

	if len(rendered) == 0 {
		return ""
	}

	// Join notifications vertically
	return lipgloss.JoinVertical(lipgloss.Center, rendered...)
}

// HasActive returns true if there are any active notifications.
func (q *NotificationQueue) HasActive() bool {
	return len(q.notifications) > 0
}

// Clear clears all notifications.
func (q *NotificationQueue) Clear() {
	q.notifications = make([]*StatChangeNotification, 0)
}
