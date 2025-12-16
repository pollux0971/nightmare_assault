package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/parallel"
)

// ParallelLoadingModel shows multi-task loading progress
type ParallelLoadingModel struct {
	coordinator  *parallel.Coordinator
	taskStates   map[parallel.TaskID]*TaskUIState
	width        int
	height       int
	spinner      spinner.Model
	startTime    time.Time
	done         bool
	cancelled    bool
	errorMessage string
}

// TaskUIState tracks UI state for a task
type TaskUIState struct {
	Name      string
	Status    parallel.TaskStatus
	Progress  int
	StartTime time.Time
	Duration  time.Duration
	Error     error
}

// NewParallelLoadingModel creates enhanced loading screen
func NewParallelLoadingModel(coordinator *parallel.Coordinator) ParallelLoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))

	taskStates := make(map[parallel.TaskID]*TaskUIState)
	// Initialize task states from coordinator
	for id, task := range coordinator.GetTasks() {
		taskStates[id] = &TaskUIState{
			Name:   task.Name,
			Status: task.Status,
		}
	}

	return ParallelLoadingModel{
		coordinator: coordinator,
		taskStates:  taskStates,
		spinner:     s,
		startTime:   time.Now(),
	}
}

// Init initializes the loading model
func (m ParallelLoadingModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		parallel.SubscribeToProgress(m.coordinator),
	)
}

// Update handles progress updates
func (m ParallelLoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case parallel.TaskProgressMsg:
		// Update task progress
		if state, ok := m.taskStates[msg.TaskID]; ok {
			state.Progress = msg.Progress
		}
		return m, parallel.SubscribeToProgress(m.coordinator)

	case parallel.TaskCompletionMsg:
		// Update task completion
		if state, ok := m.taskStates[msg.TaskID]; ok {
			if msg.Success {
				state.Status = parallel.TaskStatusCompleted
			} else {
				state.Status = parallel.TaskStatusFailed
				state.Error = msg.Error
			}
			state.Duration = msg.Duration
		}

		// Check if all critical tasks done
		if m.allCriticalTasksComplete() {
			m.done = true
			return m, func() tea.Msg {
				return parallel.ParallelLoadingDoneMsg{
					Results: m.coordinator.GetResults(),
					Errors:  m.coordinator.GetErrors(),
				}
			}
		}

		return m, parallel.SubscribeToProgress(m.coordinator)

	case parallel.ParallelLoadingErrorMsg:
		// Handle error (strict mode)
		m.errorMessage = fmt.Sprintf(
			"❌ 初始化失敗\n\n任務: %s\n錯誤: %s\n\n[Enter] 重試  [ESC] 返回主選單",
			msg.FailedTask,
			msg.Error,
		)
		m.done = true
		return m, nil

	case tea.KeyMsg:
		if m.done && m.errorMessage != "" {
			// Error screen with retry/back options
			switch msg.String() {
			case "enter":
				// Retry - return retry message
				return m, func() tea.Msg {
					return RetryGenerationMsg{}
				}
			case "esc":
				// Back to menu
				return m, func() tea.Msg {
					return parallel.ParallelLoadingDoneMsg{
						Cancelled: true,
					}
				}
			}
		} else if msg.String() == "ctrl+c" || msg.String() == "esc" {
			m.coordinator.Cancel()
			m.cancelled = true
			m.done = true
			return m, func() tea.Msg {
				return parallel.ParallelLoadingDoneMsg{Cancelled: true}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the multi-task loading screen
func (m ParallelLoadingModel) View() string {
	if m.errorMessage != "" {
		// Error screen
		return m.renderErrorScreen()
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9D4EDD")).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("🎮 正在初始化遊戲"))
	b.WriteString("\n\n")

	// Render each task status
	tasks := []parallel.TaskID{
		parallel.TaskIDRules,
		parallel.TaskIDTeammates,
		parallel.TaskIDStory,
		parallel.TaskIDDream,
	}

	for _, taskID := range tasks {
		state := m.taskStates[taskID]
		b.WriteString(m.renderTaskRow(taskID, state))
		b.WriteString("\n")
	}

	// Overall progress
	b.WriteString("\n")
	overallProgress := m.calculateOverallProgress()
	progressBar := m.renderProgressBar(overallProgress, 40)
	b.WriteString(progressBar)
	b.WriteString(fmt.Sprintf(" %d%%\n", overallProgress))

	// Elapsed time
	b.WriteString("\n")
	elapsed := time.Since(m.startTime)
	subtextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	b.WriteString(subtextStyle.Render(
		fmt.Sprintf("已用時間: %.1fs", elapsed.Seconds()),
	))

	b.WriteString("\n\n")
	b.WriteString(subtextStyle.Render("[ESC] 取消"))

	return b.String()
}

// renderTaskRow renders a single task status line
func (m ParallelLoadingModel) renderTaskRow(
	taskID parallel.TaskID,
	state *TaskUIState,
) string {
	var icon string
	var statusColor lipgloss.Color

	switch state.Status {
	case parallel.TaskStatusPending:
		icon = "○"
		statusColor = lipgloss.Color("#666666")
	case parallel.TaskStatusWaiting:
		icon = "⏳"
		statusColor = lipgloss.Color("#FFAA00")
	case parallel.TaskStatusRunning:
		icon = m.spinner.View()
		statusColor = lipgloss.Color("#00AAFF")
	case parallel.TaskStatusCompleted:
		icon = "✓"
		statusColor = lipgloss.Color("#00FF00")
	case parallel.TaskStatusFailed:
		icon = "✗"
		statusColor = lipgloss.Color("#FF0000")
	default:
		icon = "○"
		statusColor = lipgloss.Color("#666666")
	}

	style := lipgloss.NewStyle().Foreground(statusColor)

	// Build status line
	line := fmt.Sprintf(
		"%s %-20s",
		icon,
		state.Name,
	)

	// Add progress bar for running tasks
	if state.Status == parallel.TaskStatusRunning && state.Progress > 0 {
		miniBar := m.renderMiniProgressBar(state.Progress, 10)
		line += fmt.Sprintf(" %s %d%%", miniBar, state.Progress)
	} else if state.Status == parallel.TaskStatusCompleted {
		line += fmt.Sprintf(" (%.1fs)", state.Duration.Seconds())
	} else if state.Status == parallel.TaskStatusFailed && state.Error != nil {
		line += fmt.Sprintf(" 失敗: %s", state.Error.Error())
	}

	return style.Render(line)
}

// calculateOverallProgress weights tasks by importance
func (m ParallelLoadingModel) calculateOverallProgress() int {
	weights := map[parallel.TaskID]int{
		parallel.TaskIDRules:     5,
		parallel.TaskIDTeammates: 5,
		parallel.TaskIDStory:     80, // Story is the bottleneck
		parallel.TaskIDDream:     10,
	}

	totalWeight := 0
	weightedProgress := 0

	for taskID, weight := range weights {
		totalWeight += weight
		state := m.taskStates[taskID]

		if state.Status == parallel.TaskStatusCompleted {
			weightedProgress += weight * 100
		} else if state.Status == parallel.TaskStatusRunning {
			weightedProgress += weight * state.Progress
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return weightedProgress / totalWeight
}

// renderProgressBar renders a progress bar
func (m ParallelLoadingModel) renderProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	return bar
}

// renderMiniProgressBar renders a small progress bar
func (m ParallelLoadingModel) renderMiniProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	empty := width - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// renderErrorScreen renders the error screen with retry/back options
func (m ParallelLoadingModel) renderErrorScreen() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		MarginBottom(1)

	return errorStyle.Render(m.errorMessage)
}

// allCriticalTasksComplete checks if all critical tasks are done
func (m ParallelLoadingModel) allCriticalTasksComplete() bool {
	// In strict mode, ALL tasks must be complete
	for _, state := range m.taskStates {
		if state.Status != parallel.TaskStatusCompleted {
			return false
		}
	}
	return true
}

// RetryGenerationMsg signals retry generation
type RetryGenerationMsg struct{}
