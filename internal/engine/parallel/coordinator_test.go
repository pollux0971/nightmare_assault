package parallel

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCoordinatorBasic tests basic coordinator functionality
func TestCoordinatorBasic(t *testing.T) {
	config := DefaultCoordinatorConfig()
	coord := NewCoordinator(config)

	// Add a simple task with no dependencies
	task := NewTask(
		"test-task",
		"Test Task",
		PriorityHigh,
		5*time.Second,
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			return "success", nil
		},
	)

	err := coord.AddTask(task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Execute
	err = coord.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	// Check results
	results := coord.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result, ok := results["test-task"]
	if !ok {
		t.Fatal("Result not found")
	}

	if result != "success" {
		t.Fatalf("Expected 'success', got %v", result)
	}
}

// TestCoordinatorDependencies tests dependency handling
func TestCoordinatorDependencies(t *testing.T) {
	config := DefaultCoordinatorConfig()
	coord := NewCoordinator(config)

	// Task 1: No dependencies
	task1 := NewTask(
		"task1",
		"Task 1",
		PriorityHigh,
		5*time.Second,
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			return "task1-result", nil
		},
	)

	// Task 2: Depends on Task 1
	task2 := NewTask(
		"task2",
		"Task 2",
		PriorityMedium,
		5*time.Second,
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			// Check that task1 result is available
			if deps["task1"] != "task1-result" {
				return nil, errors.New("task1 result not found")
			}
			return "task2-result", nil
		},
	).WithDependencies("task1")

	coord.AddTask(task1)
	coord.AddTask(task2)

	err := coord.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	results := coord.GetResults()
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results["task2"] != "task2-result" {
		t.Fatal("Task 2 did not receive task 1 result correctly")
	}
}

// TestCoordinatorSlowTask tests that slow tasks complete (no timeout)
func TestCoordinatorSlowTask(t *testing.T) {
	config := DefaultCoordinatorConfig()
	coord := NewCoordinator(config)

	// Task that takes some time (should complete, no timeout)
	task := NewTask(
		"slow-task",
		"Slow Task",
		PriorityHigh,
		0, // Timeout field is ignored now
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			time.Sleep(200 * time.Millisecond) // Slow task
			return "slow-task-completed", nil
		},
	)

	coord.AddTask(task)

	err := coord.Execute()
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	results := coord.GetResults()
	if results["slow-task"] != "slow-task-completed" {
		t.Fatalf("Expected slow task to complete, got: %v", results["slow-task"])
	}
}

// TestCoordinatorCancel tests cancellation
func TestCoordinatorCancel(t *testing.T) {
	config := DefaultCoordinatorConfig()
	coord := NewCoordinator(config)

	// Task that can be cancelled
	task := NewTask(
		"cancellable-task",
		"Cancellable Task",
		PriorityHigh,
		10*time.Second,
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
				return "should-not-return", nil
			}
		},
	)

	coord.AddTask(task)

	// Start execution in goroutine
	done := make(chan error, 1)
	go func() {
		done <- coord.Execute()
	}()

	// Cancel after 100ms
	time.Sleep(100 * time.Millisecond)
	coord.Cancel()

	// Wait for completion
	err := <-done
	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}
}

// TestCoordinatorParallelExecution tests parallel execution
func TestCoordinatorParallelExecution(t *testing.T) {
	config := DefaultCoordinatorConfig()
	config.MaxConcurrency = 3
	coord := NewCoordinator(config)

	startTime := time.Now()

	// Add 3 tasks that take 100ms each
	for i := 0; i < 3; i++ {
		id := TaskID(string(rune('A' + i)))
		task := NewTask(
			id,
			"Task "+string(id),
			PriorityHigh,
			5*time.Second,
			func(ctx context.Context, deps DependencyResults) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return "done", nil
			},
		)
		coord.AddTask(task)
	}

	err := coord.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	elapsed := time.Since(startTime)

	// If parallel, should take ~100ms. If serial, would take ~300ms
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Tasks did not execute in parallel: took %v", elapsed)
	}

	results := coord.GetResults()
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
}

// TestCoordinatorStrictMode tests strict mode validation
func TestCoordinatorStrictMode(t *testing.T) {
	config := DefaultCoordinatorConfig()
	coord := NewCoordinator(config)

	// Task that fails
	task := NewTask(
		"failing-task",
		"Failing Task",
		PriorityCritical,
		5*time.Second,
		func(ctx context.Context, deps DependencyResults) (interface{}, error) {
			return nil, errors.New("intentional failure")
		},
	)

	coord.AddTask(task)

	err := coord.Execute()
	if err == nil {
		t.Fatal("Expected error in strict mode, got nil")
	}

	errors := coord.GetErrors()
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
}
