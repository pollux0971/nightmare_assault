// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/components"
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
	progressBar   components.ProgressBar
	progress      int
	startTime     time.Time
	elapsed       time.Duration
	flavorIndex   int
	blinkTick     int  // For blinking animations
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

// Flavor text that rotates during loading (categorized by state).
var flavorTextsConnecting = []string{
	"與未知建立連線...",
	"召喚黑暗力量...",
	"打開被遺忘的大門...",
	"穿越現實的裂縫...",
}

var flavorTextsGenerating = []string{
	"編織噩夢...",
	"深淵正在凝視著你...",
	"黑暗正在聚集...",
	"命運的齒輪開始轉動...",
	"古老的低語傳來...",
	"扭曲現實的邊界...",
}

var flavorTextsStreaming = []string{
	"陰影正在逼近...",
	"真相即將揭露...",
	"恐懼在等待著...",
	"記憶正在扭曲...",
	"虛實之間的迷霧...",
	"時間失去了意義...",
}

var flavorTextsWarning = []string{
	"不祥的預感湧上心頭...",
	"某種東西醒來了...",
	"你感覺到被監視著...",
	"空氣中瀰漫著恐懼...",
	"血月高懸...",
	"寂靜中傳來心跳聲...",
}

// Combined flavor texts for random selection
var flavorTexts = append(append(append(
	flavorTextsConnecting,
	flavorTextsGenerating...),
	flavorTextsStreaming...),
	flavorTextsWarning...)

// NewStoryLoadingModel creates a new story loading model.
func NewStoryLoadingModel() StoryLoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))

	// Create horror-themed progress bar
	progressBar := components.NewProgressBar(components.HorrorProgressBarStyle())

	return StoryLoadingModel{
		state:       LoadingInit,
		spinner:     s,
		progressBar: progressBar,
		startTime:   time.Now(),
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
		m.blinkTick++

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
		m.progressBar.SetPercent(msg.Progress)
		if msg.Content != "" {
			m.content = msg.Content
		}

		// Update progress bar
		var cmd tea.Cmd
		m.progressBar, cmd = m.progressBar.Update(msg)
		return m, cmd

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

	// Enhanced title style with glow effect
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		Underline(true)

	textStyle := lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true)

	subtextStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary).
		Italic(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)

	warningStyle := lipgloss.NewStyle().
		Foreground(colors.Warning).
		Bold(true)

	// Double border with shadow effect
	innerBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 3)

	outerBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(colors.Accent).
		Padding(2, 4)

	var b strings.Builder

	// Decorative header with moon phases
	moonPhases := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	moonIcon := moonPhases[m.blinkTick%len(moonPhases)]

	b.WriteString(titleStyle.Render(moonIcon + " Nightmare Assault " + moonIcon))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n\n")

	// Loading content based on state
	switch m.state {
	case LoadingInit, LoadingConnecting:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("連接中..."))
		b.WriteString("\n\n")
		// Progress bar with decorative brackets
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("╭─"))
		b.WriteString(m.progressBar.View())
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("─╮"))
		b.WriteString("\n\n")
		// Flavor text with icon
		flavorIdx := m.flavorIndex % len(flavorTextsConnecting)
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Accent).Render("▸ "))
		b.WriteString(subtextStyle.Render(flavorTextsConnecting[flavorIdx]))

	case LoadingGenerating:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("生成故事中..."))
		b.WriteString("\n\n")
		// Progress bar with decorative brackets
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("╭─"))
		b.WriteString(m.progressBar.View())
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("─╮"))
		b.WriteString("\n\n")
		// Flavor text with icon
		flavorIdx := m.flavorIndex % len(flavorTextsGenerating)
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Accent).Render("▸ "))
		b.WriteString(subtextStyle.Render(flavorTextsGenerating[flavorIdx]))

	case LoadingStreaming:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("接收故事中..."))
		b.WriteString("\n\n")
		// Progress bar with decorative brackets
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("╭─"))
		b.WriteString(m.progressBar.View())
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("─╮"))
		b.WriteString("\n\n")
		// Flavor text with icon
		flavorIdx := m.flavorIndex % len(flavorTextsStreaming)
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Accent).Render("▸ "))
		b.WriteString(subtextStyle.Render(flavorTextsStreaming[flavorIdx]))

	case LoadingComplete:
		// Success animation with expanding effect
		successIcon := "✓"
		if m.blinkTick%4 == 0 {
			successIcon = "✓"
		} else if m.blinkTick%4 == 1 {
			successIcon = "✔"
		} else if m.blinkTick%4 == 2 {
			successIcon = "✓"
		} else {
			successIcon = "✔"
		}
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Render(successIcon + " 故事準備就緒 " + successIcon))
		b.WriteString("\n\n")
		// Full progress bar with glow
		m.progressBar.SetPercent(100)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("╭─"))
		b.WriteString(m.progressBar.View())
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("─╮"))
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render("深淵的故事已經開始..."))

	case LoadingTimeout:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(textStyle.Render("生成中..."))
		b.WriteString("\n\n")
		// Progress bar
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("╭─"))
		b.WriteString(m.progressBar.View())
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("─╮"))
		b.WriteString("\n\n")
		// Blinking warning with alternating colors
		warningIcon := "⚠"
		warningText := "生成時間較長，請耐心等待..."
		if m.blinkTick%2 == 0 {
			b.WriteString(warningStyle.Render(warningIcon + " " + warningText))
		} else {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B00")).
				Bold(true).
				Render(warningIcon + " " + warningText))
		}
		b.WriteString("\n")
		// Flavor text with pulsing effect
		flavorIdx := m.flavorIndex % len(flavorTextsWarning)
		if m.blinkTick%2 == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("▸ "))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B00")).Render("▸ "))
		}
		b.WriteString(subtextStyle.Render(flavorTextsWarning[flavorIdx]))

	case LoadingError:
		// Error with cross mark
		b.WriteString(errorStyle.Render("✗ 發生錯誤 ✗"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.errorMessage))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		b.WriteString("\n\n")
		b.WriteString(subtextStyle.Render("按 Enter 重試，按 Esc 返回主選單"))
	}

	// Elapsed time with clock icon
	if m.state != LoadingComplete && m.state != LoadingError {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Border).Render("─────────────────────────────────────────"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Secondary).Render("⏱ "))
		b.WriteString(subtextStyle.Render(fmt.Sprintf("已經過 %.1f 秒", m.elapsed.Seconds())))
	}

	// Hint with key icon
	if m.state != LoadingComplete {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colors.Secondary).Render("⌨ "))
		b.WriteString(subtextStyle.Render("Esc: 取消"))
	}

	// Apply double border for depth effect
	innerContent := innerBorderStyle.Render(b.String())
	return outerBorderStyle.Render(innerContent)
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
