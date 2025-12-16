// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// SetupState represents the current step in the game setup flow.
type SetupState int

const (
	SetupThemeInput SetupState = iota
	SetupModelInput
	SetupDifficultySelect
	SetupLengthSelect
	SetupAdultModeToggle
	SetupSummary
)

// GameSetupModel represents the game setup flow state.
type GameSetupModel struct {
	state          SetupState
	config         *game.GameConfig
	themeInput     textinput.Model
	modelInput     textinput.Model
	themeError     string
	selectedIndex  int
	width          int
	height         int
	cancelled      bool
	confirmed      bool
	preloadStarted bool // Track if we've started preloading
}

// GameSetupDoneMsg is sent when setup is complete.
type GameSetupDoneMsg struct {
	Config    *game.GameConfig
	Cancelled bool
}

// NewGameSetupModel creates a new game setup model.
func NewGameSetupModel() GameSetupModel {
	ti := textinput.New()
	ti.Placeholder = "例如：廢棄醫院、詛咒洋館、末日地鐵站..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	mi := textinput.New()
	mi.Placeholder = "例如：openai/gpt-4-turbo, anthropic/claude-3-opus..."
	mi.CharLimit = 100
	mi.Width = 50

	cfg := game.NewGameConfig()

	return GameSetupModel{
		state:      SetupThemeInput,
		config:     cfg,
		themeInput: ti,
		modelInput: mi,
	}
}

// Init initializes the model.
func (m GameSetupModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (m GameSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, func() tea.Msg {
				return GameSetupDoneMsg{Config: nil, Cancelled: true}
			}

		case "esc":
			return m.handleBack()
		}

		// Delegate to current state handler
		return m.handleKeyForState(msg)
	}

	// Update text input if in theme or model input state
	if m.state == SetupThemeInput {
		var cmd tea.Cmd
		m.themeInput, cmd = m.themeInput.Update(msg)
		return m, cmd
	}
	if m.state == SetupModelInput {
		var cmd tea.Cmd
		m.modelInput, cmd = m.modelInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m GameSetupModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case SetupThemeInput:
		// Cancel setup entirely
		m.cancelled = true
		return m, func() tea.Msg {
			return GameSetupDoneMsg{Config: nil, Cancelled: true}
		}
	case SetupModelInput:
		m.state = SetupThemeInput
		m.themeInput.Focus()
	case SetupDifficultySelect:
		m.state = SetupModelInput
		m.modelInput.Focus()
	case SetupLengthSelect:
		m.state = SetupDifficultySelect
		m.selectedIndex = int(m.config.Difficulty)
	case SetupAdultModeToggle:
		m.state = SetupLengthSelect
		m.selectedIndex = int(m.config.Length)
	case SetupSummary:
		m.state = SetupAdultModeToggle
		if m.config.AdultMode {
			m.selectedIndex = 0
		} else {
			m.selectedIndex = 1
		}
	}
	return m, nil
}

func (m GameSetupModel) handleKeyForState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case SetupThemeInput:
		return m.handleThemeInput(msg)
	case SetupModelInput:
		return m.handleModelInput(msg)
	case SetupDifficultySelect:
		return m.handleDifficultySelect(msg)
	case SetupLengthSelect:
		return m.handleLengthSelect(msg)
	case SetupAdultModeToggle:
		return m.handleAdultModeToggle(msg)
	case SetupSummary:
		return m.handleSummary(msg)
	}
	return m, nil
}

func (m GameSetupModel) handleThemeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		theme := strings.TrimSpace(m.themeInput.Value())
		if err := m.config.SetTheme(theme); err != nil {
			m.themeError = err.Error()
			return m, nil
		}
		m.themeError = ""
		m.state = SetupModelInput
		m.modelInput.SetValue(m.config.Model) // Pre-fill with default
		m.modelInput.Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.themeInput, cmd = m.themeInput.Update(msg)
	return m, cmd
}

func (m GameSetupModel) handleModelInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		model := strings.TrimSpace(m.modelInput.Value())
		if model == "" {
			model = m.config.Model // Use default if empty
		}
		m.config.Model = model
		m.state = SetupDifficultySelect
		m.selectedIndex = int(m.config.Difficulty)
		return m, nil
	}

	var cmd tea.Cmd
	m.modelInput, cmd = m.modelInput.Update(msg)
	return m, cmd
}

func (m GameSetupModel) handleDifficultySelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < 2 {
			m.selectedIndex++
		}
	case "1":
		m.selectedIndex = 0
	case "2":
		m.selectedIndex = 1
	case "3":
		m.selectedIndex = 2
	case "enter", " ":
		m.config.SetDifficulty(game.DifficultyLevel(m.selectedIndex))
		m.state = SetupLengthSelect
		m.selectedIndex = int(m.config.Length)
	}
	return m, nil
}

func (m GameSetupModel) handleLengthSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < 2 {
			m.selectedIndex++
		}
	case "1":
		m.selectedIndex = 0
	case "2":
		m.selectedIndex = 1
	case "3":
		m.selectedIndex = 2
	case "enter", " ":
		m.config.SetLength(game.GameLength(m.selectedIndex))
		m.state = SetupAdultModeToggle
		if m.config.AdultMode {
			m.selectedIndex = 0
		} else {
			m.selectedIndex = 1
		}
	}
	return m, nil
}

func (m GameSetupModel) handleAdultModeToggle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		// Toggle between 0 and 1
		m.selectedIndex = 1 - m.selectedIndex
	case "1":
		m.selectedIndex = 0
	case "2":
		m.selectedIndex = 1
	case "enter", " ":
		_ = m.config.SetAdultMode(m.selectedIndex == 0)
		m.state = SetupSummary
		m.selectedIndex = 0
	}
	return m, nil
}

func (m GameSetupModel) handleSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Trigger speculative preloading when user is in summary screen
	// This happens in the background while user reviews settings
	if !m.preloadStarted {
		m.preloadStarted = true
		// Start preloading in background (non-blocking)
		go func() {
			// This will be picked up by builders.go through the global cache
			// We don't need to do anything here yet - the cache will be set
			// by the parallel coordinator when it starts
		}()
	}

	switch msg.String() {
	case "up", "k", "down", "j":
		// Toggle between confirm and edit
		m.selectedIndex = 1 - m.selectedIndex
	case "1":
		m.selectedIndex = 0
	case "2":
		m.selectedIndex = 1
	case "enter", " ":
		if m.selectedIndex == 0 {
			// Confirm - done with setup
			m.confirmed = true
			return m, func() tea.Msg {
				return GameSetupDoneMsg{Config: m.config, Cancelled: false}
			}
		} else {
			// Edit - go back to theme input
			m.state = SetupThemeInput
			m.themeInput.SetValue(m.config.Theme)
			m.themeInput.Focus()
			// Reset preload flag since config will change
			m.preloadStarted = false
		}
	}
	return m, nil
}

// View renders the setup screen.
func (m GameSetupModel) View() string {
	tm := themes.GetManager()
	theme := tm.GetCurrentTheme()
	colors := theme.Colors

	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(colors.Primary)

	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary)

	errorStyle := lipgloss.NewStyle().
		Foreground(colors.Error)

	warningStyle := lipgloss.NewStyle().
		Foreground(colors.Warning).
		Bold(true)

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎮 新遊戲設定"))
	b.WriteString("\n\n")

	// Progress indicator
	progress := m.renderProgress()
	b.WriteString(subtitleStyle.Render(progress))
	b.WriteString("\n\n")

	// Content based on state
	var content string
	switch m.state {
	case SetupThemeInput:
		content = m.renderThemeInput(titleStyle, hintStyle, errorStyle)
	case SetupModelInput:
		content = m.renderModelInput(titleStyle, hintStyle)
	case SetupDifficultySelect:
		content = m.renderDifficultySelect(titleStyle, selectedStyle, normalStyle, hintStyle)
	case SetupLengthSelect:
		content = m.renderLengthSelect(titleStyle, selectedStyle, normalStyle, hintStyle)
	case SetupAdultModeToggle:
		content = m.renderAdultModeToggle(titleStyle, selectedStyle, normalStyle, hintStyle, warningStyle)
	case SetupSummary:
		content = m.renderSummary(titleStyle, selectedStyle, normalStyle, hintStyle, warningStyle)
	}

	b.WriteString(borderStyle.Render(content))
	b.WriteString("\n\n")

	// Navigation hints
	hints := m.getNavigationHints()
	b.WriteString(hintStyle.Render(hints))

	return b.String()
}

func (m GameSetupModel) renderProgress() string {
	steps := []string{"主題", "難度", "長度", "分級", "確認"}
	current := int(m.state)

	var parts []string
	for i, step := range steps {
		if i < current {
			parts = append(parts, fmt.Sprintf("✓ %s", step))
		} else if i == current {
			parts = append(parts, fmt.Sprintf("● %s", step))
		} else {
			parts = append(parts, fmt.Sprintf("○ %s", step))
		}
	}

	return strings.Join(parts, " → ")
}

func (m GameSetupModel) renderThemeInput(titleStyle, hintStyle, errorStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 1/5：輸入故事主題"))
	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("輸入一個恐怖主題來開始你的惡夢冒險"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("(3-100 個字元)"))
	b.WriteString("\n\n")

	b.WriteString(m.themeInput.View())

	if m.themeError != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("⚠ " + m.themeError))
	}

	return b.String()
}

func (m GameSetupModel) renderModelInput(titleStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 2/6：輸入 LLM 模型名稱"))
	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("輸入 OpenRouter 模型名稱 (例如: openai/gpt-4-turbo)"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("按 Enter 使用預設模型"))
	b.WriteString("\n\n")

	b.WriteString(m.modelInput.View())

	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("常見模型:"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  • openai/gpt-4-turbo"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  • anthropic/claude-3-opus"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  • google/gemini-pro"))

	return b.String()
}

func (m GameSetupModel) renderDifficultySelect(titleStyle, selectedStyle, normalStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 2/5：選擇難度"))
	b.WriteString("\n\n")

	difficulties := []game.DifficultyLevel{
		game.DifficultyEasy,
		game.DifficultyHard,
		game.DifficultyHell,
	}

	for i, d := range difficulties {
		prefix := "  "
		style := normalStyle
		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}

		b.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(d.String())))
		b.WriteString(fmt.Sprintf("      %s\n", hintStyle.Render(d.Description())))
		if i < len(difficulties)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m GameSetupModel) renderLengthSelect(titleStyle, selectedStyle, normalStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 3/5：選擇遊戲長度"))
	b.WriteString("\n\n")

	lengths := []game.GameLength{
		game.LengthShort,
		game.LengthMedium,
		game.LengthLong,
	}

	for i, l := range lengths {
		prefix := "  "
		style := normalStyle
		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}

		b.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(l.String())))
		b.WriteString(fmt.Sprintf("      %s\n", hintStyle.Render(l.Description())))
		if i < len(lengths)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m GameSetupModel) renderAdultModeToggle(titleStyle, selectedStyle, normalStyle, hintStyle, warningStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 4/5：內容分級"))
	b.WriteString("\n\n")

	options := []struct {
		name    string
		desc    string
		warning bool
	}{
		{"開啟 18+ 模式", "包含血腥場景、心理恐怖、成人內容", true},
		{"關閉 18+ 模式", "適合一般玩家的恐怖體驗", false},
	}

	for i, opt := range options {
		prefix := "  "
		style := normalStyle
		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}

		b.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(opt.name)))
		if opt.warning {
			b.WriteString(fmt.Sprintf("      %s\n", warningStyle.Render("⚠ "+opt.desc)))
		} else {
			b.WriteString(fmt.Sprintf("      %s\n", hintStyle.Render(opt.desc)))
		}
		if i < len(options)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m GameSetupModel) renderSummary(titleStyle, selectedStyle, normalStyle, hintStyle, warningStyle lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("步驟 5/5：確認設定"))
	b.WriteString("\n\n")

	// Summary box
	b.WriteString(fmt.Sprintf("  主題：%s\n", m.config.Theme))
	b.WriteString(fmt.Sprintf("  難度：%s\n", m.config.Difficulty.String()))
	b.WriteString(fmt.Sprintf("  長度：%s (%s)\n", m.config.Length.String(), m.config.Length.Description()))

	if m.config.AdultMode {
		b.WriteString(fmt.Sprintf("  分級：%s\n", warningStyle.Render("18+ 開啟")))
	} else {
		b.WriteString(fmt.Sprintf("  分級：%s\n", "一般"))
	}

	b.WriteString("\n")

	// Action options
	options := []string{"確認開始遊戲", "返回修改設定"}
	for i, opt := range options {
		prefix := "  "
		style := normalStyle
		if i == m.selectedIndex {
			prefix = "❯ "
			style = selectedStyle
		}
		b.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(opt)))
	}

	return b.String()
}

func (m GameSetupModel) getNavigationHints() string {
	switch m.state {
	case SetupThemeInput:
		return "Enter: 下一步  |  Esc: 取消設定"
	case SetupSummary:
		return "↑/↓ 或 j/k: 移動  |  Enter: 確認  |  Esc: 上一步"
	default:
		return "↑/↓ 或 j/k: 移動  |  1-3: 直接選擇  |  Enter: 下一步  |  Esc: 上一步"
	}
}

// IsDone returns true if setup is complete.
func (m GameSetupModel) IsDone() bool {
	return m.confirmed || m.cancelled
}

// IsCancelled returns true if setup was cancelled.
func (m GameSetupModel) IsCancelled() bool {
	return m.cancelled
}

// GetConfig returns the configured game settings.
func (m GameSetupModel) GetConfig() *game.GameConfig {
	return m.config
}
