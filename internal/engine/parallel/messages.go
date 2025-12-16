package parallel

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TaskProgressMsg reports task progress to BubbleTea
type TaskProgressMsg struct {
	TaskID   TaskID
	Progress int
	Message  string
}

// TaskCompletionMsg reports task completion to BubbleTea
type TaskCompletionMsg struct {
	TaskID   TaskID
	Success  bool
	Result   interface{}
	Error    error
	Duration time.Duration
}

// ParallelLoadingDoneMsg signals all generation complete
type ParallelLoadingDoneMsg struct {
	Results   map[TaskID]interface{}
	Errors    map[TaskID]error
	Cancelled bool
}

// ParallelLoadingErrorMsg signals generation error (strict mode)
type ParallelLoadingErrorMsg struct {
	FailedTask TaskID
	Error      error
	Message    string
}

// StartGeneration launches parallel generation and returns BubbleTea Cmd
func StartGeneration(coordinator *Coordinator) tea.Cmd {
	return func() tea.Msg {
		// Execute coordinator in background
		errChan := make(chan error, 1)
		go func() {
			errChan <- coordinator.Execute()
		}()

		// Subscribe to progress and completion
		return waitForCompletion(coordinator, errChan)
	}
}

// waitForCompletion waits for coordinator to finish and returns final message
func waitForCompletion(coord *Coordinator, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		// Wait for completion or error
		err := <-errChan

		if err != nil {
			// Strict mode: Any error stops generation
			// Try to find which task failed
			errors := coord.GetErrors()
			if len(errors) > 0 {
				// Return first error
				for taskID, taskErr := range errors {
					return ParallelLoadingErrorMsg{
						FailedTask: taskID,
						Error:      taskErr,
						Message:    "Generation failed",
					}
				}
			}

			// Generic error
			return ParallelLoadingErrorMsg{
				FailedTask: "",
				Error:      err,
				Message:    "Parallel generation failed",
			}
		}

		// Success - return all results
		return ParallelLoadingDoneMsg{
			Results:   coord.GetResults(),
			Errors:    coord.GetErrors(),
			Cancelled: false,
		}
	}
}

// subscribeToProgress creates Cmd to listen for progress updates
// This is called from the loading UI to get real-time updates
func SubscribeToProgress(coordinator *Coordinator) tea.Cmd {
	return func() tea.Msg {
		select {
		case progress, ok := <-coordinator.ProgressChan():
			if !ok {
				// Channel closed, generation complete
				return nil
			}
			return TaskProgressMsg{
				TaskID:   progress.TaskID,
				Progress: progress.Progress,
				Message:  progress.Message,
			}

		case completion, ok := <-coordinator.CompletionChan():
			if !ok {
				// Channel closed
				return nil
			}
			return TaskCompletionMsg{
				TaskID:   completion.TaskID,
				Success:  completion.Success,
				Result:   completion.Result,
				Error:    completion.Error,
				Duration: completion.Duration,
			}
		}
	}
}
