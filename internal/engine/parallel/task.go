// Package parallel provides parallel task execution for game generation.
package parallel

import (
	"context"
	"fmt"
	"time"
)

// TaskID uniquely identifies a generation task
type TaskID string

const (
	TaskIDRules     TaskID = "rules"
	TaskIDStory     TaskID = "story"
	TaskIDTeammates TaskID = "teammates"
	TaskIDDream     TaskID = "dream"
)

// TaskStatus represents task execution state
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusWaiting   // Waiting for dependencies
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

// String returns the string representation of TaskStatus
func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "pending"
	case TaskStatusWaiting:
		return "waiting"
	case TaskStatusRunning:
		return "running"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	case TaskStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TaskPriority determines execution priority (higher = more important)
type TaskPriority int

const (
	PriorityCritical TaskPriority = 100 // Story - blocks game start
	PriorityHigh     TaskPriority = 80  // Rules - needed by Dream
	PriorityMedium   TaskPriority = 60  // Dream - enhances experience
	PriorityLow      TaskPriority = 40  // Teammates - can be skipped
)

// DependencyResults holds results from completed dependencies
type DependencyResults map[TaskID]interface{}

// ProgressCallback is called to report task progress (0-100)
type ProgressCallback func(progress int)

// ExecuteFunc is the function that executes the task
type ExecuteFunc func(ctx context.Context, deps DependencyResults) (interface{}, error)

// Task represents a single generation task
type Task struct {
	ID           TaskID
	Name         string
	Priority     TaskPriority
	Dependencies []TaskID
	Timeout      time.Duration
	Optional     bool // If true, failure won't stop game (UNUSED in strict mode)

	Execute    ExecuteFunc
	OnProgress ProgressCallback

	// Runtime state
	Status     TaskStatus
	StartTime  time.Time
	EndTime    time.Time
	Result     interface{}
	Error      error
	Progress   int
	RetryCount int
}

// NewTask creates a new task with default values
func NewTask(id TaskID, name string, priority TaskPriority, timeout time.Duration, execute ExecuteFunc) *Task {
	return &Task{
		ID:           id,
		Name:         name,
		Priority:     priority,
		Dependencies: []TaskID{},
		Timeout:      timeout,
		Optional:     false,
		Execute:      execute,
		Status:       TaskStatusPending,
		Progress:     0,
		RetryCount:   0,
	}
}

// WithDependencies sets task dependencies
func (t *Task) WithDependencies(deps ...TaskID) *Task {
	t.Dependencies = append(t.Dependencies, deps...)
	return t
}

// WithProgress sets progress callback
func (t *Task) WithProgress(callback ProgressCallback) *Task {
	t.OnProgress = callback
	return t
}

// IsComplete returns true if task is completed
func (t *Task) IsComplete() bool {
	return t.Status == TaskStatusCompleted
}

// IsFailed returns true if task is failed
func (t *Task) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// IsCancelled returns true if task is cancelled
func (t *Task) IsCancelled() bool {
	return t.Status == TaskStatusCancelled
}

// IsRunning returns true if task is running
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusRunning
}

// Duration returns task execution duration
func (t *Task) Duration() time.Duration {
	if t.StartTime.IsZero() {
		return 0
	}
	if t.EndTime.IsZero() {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}

// Validate checks if task configuration is valid
func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if t.Name == "" {
		return fmt.Errorf("task name is required")
	}
	if t.Execute == nil {
		return fmt.Errorf("task execute function is required")
	}
	// Note: Timeout validation removed - tasks can run indefinitely for slow models
	return nil
}

// TaskProgress reports task progress updates
type TaskProgress struct {
	TaskID   TaskID
	Progress int
	Message  string
}

// TaskCompletion reports task completion
type TaskCompletion struct {
	TaskID   TaskID
	Success  bool
	Result   interface{}
	Error    error
	Duration time.Duration
}
