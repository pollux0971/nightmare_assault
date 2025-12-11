// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// LoadingState represents the current loading state.
type LoadingState int

const (
	LoadingInit LoadingState = iota
	LoadingConnecting
	LoadingGenerating
	LoadingStreaming
	LoadingComplete
	LoadingError
	LoadingTimeout
)

// StoryLoadingModel represents the story loading screen.
type StoryLoadingModel struct {
	state         LoadingState
	spinner       spinner.Model
	progress      int
	startTime     time.Time
	elapsed       time.Duration
	flavorIndex   int
	width         int
	height        int
	errorMessage  string
	content       string
	done          bool
	cancelled     bool
}

// StoryLoadingDoneMsg is sent when loading completes.
type StoryLoadingDoneMsg struct {
	Content   string
	Error     error
	Cancelled bool
}

// StoryLoadingTickMsg is for periodic updates.
type StoryLoadingTickMsg time.Time

// StoryLoadingProgressMsg updates progress.
type StoryLoadingProgressMsg struct {
	Progress int
	State    LoadingState
	Content  string
}

// StoryLoadingErrorMsg indicates an error.
type StoryLoadingErrorMsg struct {
	Error error
}

// Flavor text that rotates during loading.
var flavorTexts = []string{
	"正在進入惡夢...",
	"黑暗正在聚集...",
	"命運的齒輪開始轉動...",
	"深淵正在回望你...",
	"恐懼在等待著...",
	"古老的低語傳來...",
	"陰影正在逼近...",
	"真相即將揭露...",
}

// NewStoryLoadingModel creates a new story loading model.
func NewStoryLoadingModel() StoryLoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))

	return StoryLoadingModel{
		state:     LoadingInit,
		spinner:   s,
		startTime: time.Now(),
	}
}

// Init initializes the model.
func (m StoryLoadingModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return StoryLoadingTickMsg(t)
	})
}

// Update handles messages.
func (m StoryLoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			m.done = true
			return m, func() tea.Msg {
				return StoryLoadingDoneMsg{Cancelled: true}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case StoryLoadingTickMsg:
		m.elapsed = time.Since(m.startTime)

		// Rotate flavor text every 3 seconds
		if int(m.elapsed.Seconds())%3 == 0 {
			m.flavorIndex = (m.flavorIndex + 1) % len(flavorTexts)
		}

		// Check for timeout warning
		if m.elapsed > 8*time.Second && m.state != LoadingStreaming && m.state != LoadingComplete {
			m.state = LoadingTimeout
		}

		return m, tickCmd()

	case StoryLoadingProgressMsg:
		m.state = msg.State
		m.progress = msg.Progress
		if msg.Content != "" {
			m.content = msg.Content
		}
		return m, nil

	case StoryLoadingErrorMsg:
		m.state = LoadingError
		m.errorMessage = msg.Error.Error()
		return m, nil

	case StoryLoadingDoneMsg:
		m.done = true
		if msg.Error != nil {
			m.state = LoadingError
			m.errorMessage = msg.Error.Error()
		} else {
			m.state = LoadingComplete
			m.content = msg.Content
		}
		return m, nil
	}

	return m, nil
}

// View renders the loading screen.
func (m StoryLoadingModel) View() string {
	tm := themes.GetManager()
	theme := tm.GetCurrentTheme()
	colors := theme.Colors

	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(colors.Primary)

	subtextStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)

	errorStyle := lipgloss.NewStyle().
		Foreground(colors.Error)

	warningStyle := lipgloss.NewStyle().
		Foreground(colors.Warning)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(2, 4)

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🌙 Nightmare Assault"))
	b.WriteString("\n\n")

	// Loading content based on state
	switch m.state {
	case LoadingInit, LoadingConnecting:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("連接中..."))
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render(flavorTexts[m.flavorIndex]))

	case LoadingGenerating:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("生成故事中..."))
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render(flavorTexts[m.flavorIndex]))

	case LoadingStreaming:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render(fmt.Sprintf("接收故事中... %d%%", m.progress)))
		b.WriteString("\n\n")
		// Show preview of streamed content
		if m.content != "" {
			preview := m.content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			b.WriteString(subtextStyle.Render(preview))
		}

	case LoadingComplete:
		b.WriteString(textStyle.Render("✓ 故事準備就緒"))

	case LoadingTimeout:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("生成中..."))
		b.WriteString("\n\n")
		b.WriteString(warningStyle.Render("⚠ 連接較慢，請稍候..."))
		b.WriteString("\n")
		b.WriteString(subtextStyle.Render(fmt.Sprintf("已等待 %.0f 秒", m.elapsed.Seconds())))

	case LoadingError:
		b.WriteString(errorStyle.Render("✗ 發生錯誤"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.errorMessage))
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render("按 Enter 重試，按 Esc 返回主選單"))
	}

	// Elapsed time
	if m.state != LoadingComplete && m.state != LoadingError {
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render(fmt.Sprintf("%.1f 秒", m.elapsed.Seconds())))
	}

	// Hint
	if m.state != LoadingComplete {
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render("Esc: 取消"))
	}

	return borderStyle.Render(b.String())
}

// IsDone returns true if loading is complete or cancelled.
func (m StoryLoadingModel) IsDone() bool {
	return m.done
}

// IsCancelled returns true if loading was cancelled.
func (m StoryLoadingModel) IsCancelled() bool {
	return m.cancelled
}

// GetContent returns the loaded content.
func (m StoryLoadingModel) GetContent() string {
	return m.content
}

// GetError returns the error message if any.
func (m StoryLoadingModel) GetError() string {
	return m.errorMessage
}

// SetState updates the loading state.
func (m *StoryLoadingModel) SetState(state LoadingState) {
	m.state = state
}

// SetProgress updates the progress percentage.
func (m *StoryLoadingModel) SetProgress(progress int) {
	m.progress = progress
}

// SetContent updates the content being loaded.
func (m *StoryLoadingModel) SetContent(content string) {
	m.content = content
}
