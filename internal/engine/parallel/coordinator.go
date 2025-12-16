package parallel

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CoordinatorConfig configures the coordinator
type CoordinatorConfig struct {
	MaxConcurrency int
	GlobalTimeout  time.Duration
	ProgressBuffer int
	MaxRetries     int  // Maximum retry attempts for failed tasks
	RetryDelay     time.Duration // Initial retry delay (exponential backoff)
}

// DefaultCoordinatorConfig returns default configuration
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxConcurrency: 4,
		GlobalTimeout:  60 * time.Second,
		ProgressBuffer: 100,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
	}
}

// Coordinator manages parallel task execution
type Coordinator struct {
	config         CoordinatorConfig
	tasks          map[TaskID]*Task
	graph          *DependencyGraph
	ctx            context.Context
	cancel         context.CancelFunc
	results        map[TaskID]interface{}
	errors         map[TaskID]error
	progressChan   chan TaskProgress
	completionChan chan TaskCompletion
	mu             sync.RWMutex
	wg             sync.WaitGroup
	semaphore      chan struct{} // Limit concurrent tasks
}

// NewCoordinator creates a parallel generation coordinator
func NewCoordinator(config CoordinatorConfig) *Coordinator {
	ctx, cancel := context.WithCancel(context.Background())

	// Create semaphore for limiting concurrency
	semaphore := make(chan struct{}, config.MaxConcurrency)

	return &Coordinator{
		config:         config,
		tasks:          make(map[TaskID]*Task),
		graph:          NewDependencyGraph(),
		ctx:            ctx,
		cancel:         cancel,
		results:        make(map[TaskID]interface{}),
		errors:         make(map[TaskID]error),
		progressChan:   make(chan TaskProgress, config.ProgressBuffer),
		completionChan: make(chan TaskCompletion, 10),
		semaphore:      semaphore,
	}
}

// AddTask registers a task for execution
func (c *Coordinator) AddTask(task *Task) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task %s: %w", task.ID, err)
	}

	if _, exists := c.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	c.tasks[task.ID] = task
	c.graph.AddNode(task.ID, task.Dependencies)
	return nil
}

// Execute runs all tasks respecting dependencies
func (c *Coordinator) Execute() error {
	// Validate graph
	if err := c.graph.Validate(); err != nil {
		return fmt.Errorf("invalid dependency graph: %w", err)
	}

	// Check for cycles
	if c.graph.HasCycle() {
		return fmt.Errorf("dependency cycle detected")
	}

	// Group tasks by wave
	waves, err := c.graph.GroupByWave()
	if err != nil {
		return fmt.Errorf("failed to group tasks: %w", err)
	}

	// Execute waves in order
	for waveNum, wave := range waves {
		if err := c.executeWave(waveNum, wave); err != nil {
			return err
		}
	}

	// Wait for all tasks
	c.wg.Wait()

	// Close channels
	close(c.progressChan)
	close(c.completionChan)

	// Validate results (strict mode)
	return c.validateResults()
}

// executeWave runs a wave of independent tasks in parallel
func (c *Coordinator) executeWave(waveNum int, wave []TaskID) error {
	// Start all tasks in the wave
	for _, taskID := range wave {
		task := c.getTask(taskID)
		if task == nil {
			continue
		}

		// Check if dependencies are satisfied
		if !c.dependenciesSatisfied(task) {
			task.Status = TaskStatusWaiting
			return fmt.Errorf("task %s dependencies not satisfied", taskID)
		}

		// Check context cancellation
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		// Launch task
		c.wg.Add(1)
		go c.executeTask(task)
	}

	// Wait for all tasks in this wave to complete before moving to next wave
	// This is critical for dependency satisfaction
	c.waitForWave(wave)

	return nil
}

// waitForWave waits for all tasks in a wave to complete
func (c *Coordinator) waitForWave(wave []TaskID) {
	// Create a local wait group for this wave
	var waveWG sync.WaitGroup

	for _, taskID := range wave {
		task := c.getTask(taskID)
		if task == nil {
			continue
		}

		waveWG.Add(1)
		go func(tid TaskID) {
			defer waveWG.Done()

			// Poll until task is completed, failed, or context cancelled
			for {
				// Check for context cancellation first
				select {
				case <-c.ctx.Done():
					return
				default:
				}

				c.mu.RLock()
				t := c.tasks[tid]
				status := t.Status
				c.mu.RUnlock()

				if status == TaskStatusCompleted || status == TaskStatusFailed {
					return
				}

				time.Sleep(10 * time.Millisecond)
			}
		}(taskID)
	}

	waveWG.Wait()
}

// executeTask runs a single task with timeout, retry, and error handling
func (c *Coordinator) executeTask(task *Task) {
	defer c.wg.Done()

	// Acquire semaphore (limit concurrency)
	c.semaphore <- struct{}{}
	defer func() { <-c.semaphore }()

	// Retry loop
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := c.config.RetryDelay * time.Duration(1<<(attempt-1))
			time.Sleep(delay)
		}

		// Execute with timeout
		result, err := c.executeTaskOnce(task, attempt)

		if err == nil {
			// Success
			c.mu.Lock()
			task.Status = TaskStatusCompleted
			task.Result = result
			task.EndTime = time.Now()
			c.results[task.ID] = result
			c.mu.Unlock()

			// Report completion
			c.completionChan <- TaskCompletion{
				TaskID:   task.ID,
				Success:  true,
				Result:   result,
				Duration: task.Duration(),
			}

			return
		}

		// Check if should retry
		if attempt < c.config.MaxRetries {
			// Check if it's a retryable error (not context cancellation)
			if c.ctx.Err() != nil {
				// Context cancelled, don't retry
				break
			}
			// Retry
			continue
		}

		// Max retries reached, fail
		c.mu.Lock()
		task.Status = TaskStatusFailed
		task.Error = err
		task.EndTime = time.Now()
		c.errors[task.ID] = err
		c.mu.Unlock()

		// Report failure
		c.completionChan <- TaskCompletion{
			TaskID:   task.ID,
			Success:  false,
			Error:    err,
			Duration: task.Duration(),
		}

		return
	}
}

// executeTaskOnce executes task once with timeout
func (c *Coordinator) executeTaskOnce(task *Task, attempt int) (interface{}, error) {
	// Update status
	c.mu.Lock()
	task.Status = TaskStatusRunning
	if attempt == 0 {
		task.StartTime = time.Now()
	}
	task.RetryCount = attempt
	c.mu.Unlock()

	// Create task context (no timeout - let slow models complete)
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	// Setup progress callback
	if task.OnProgress != nil {
		originalCallback := task.OnProgress
		task.OnProgress = func(progress int) {
			// Report progress
			select {
			case c.progressChan <- TaskProgress{
				TaskID:   task.ID,
				Progress: progress,
				Message:  fmt.Sprintf("%s: %d%%", task.Name, progress),
			}:
			default:
				// Channel full, skip
			}

			// Call original callback
			originalCallback(progress)
		}
	}

	// Get dependency results
	deps := c.getDependencyResults(task.Dependencies)

	// Execute task
	result, err := task.Execute(ctx, deps)

	// Check context error (only cancellation, no timeout)
	if ctx.Err() != nil {
		return nil, fmt.Errorf("task %s cancelled: %w", task.ID, ctx.Err())
	}

	return result, err
}

// dependenciesSatisfied checks if all dependencies are completed
func (c *Coordinator) dependenciesSatisfied(task *Task) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, depID := range task.Dependencies {
		if _, ok := c.results[depID]; !ok {
			return false
		}
	}

	return true
}

// getDependencyResults returns results from completed dependencies
func (c *Coordinator) getDependencyResults(deps []TaskID) DependencyResults {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(DependencyResults)
	for _, depID := range deps {
		if result, ok := c.results[depID]; ok {
			results[depID] = result
		}
	}

	return results
}

// validateResults checks if all required tasks succeeded (strict mode)
func (c *Coordinator) validateResults() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// In strict mode, ALL tasks must succeed
	for taskID, task := range c.tasks {
		if !task.IsComplete() {
			if task.IsFailed() {
				return fmt.Errorf("required task %s failed: %w", taskID, task.Error)
			}
			return fmt.Errorf("required task %s did not complete (status: %s)", taskID, task.Status)
		}
	}

	return nil
}

// Cancel stops all running tasks
func (c *Coordinator) Cancel() {
	c.cancel()
}

// GetResults returns all task results
func (c *Coordinator) GetResults() map[TaskID]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[TaskID]interface{})
	for k, v := range c.results {
		results[k] = v
	}
	return results
}

// GetErrors returns all task errors
func (c *Coordinator) GetErrors() map[TaskID]error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	errors := make(map[TaskID]error)
	for k, v := range c.errors {
		errors[k] = v
	}
	return errors
}

// GetTasks returns all registered tasks
func (c *Coordinator) GetTasks() map[TaskID]*Task {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tasks := make(map[TaskID]*Task)
	for k, v := range c.tasks {
		tasks[k] = v
	}
	return tasks
}

// GetTask returns a specific task
func (c *Coordinator) getTask(taskID TaskID) *Task {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.tasks[taskID]
}

// ProgressChan returns the progress channel
func (c *Coordinator) ProgressChan() <-chan TaskProgress {
	return c.progressChan
}

// CompletionChan returns the completion channel
func (c *Coordinator) CompletionChan() <-chan TaskCompletion {
	return c.completionChan
}
