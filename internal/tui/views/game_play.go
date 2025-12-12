// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/commands"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/themes"
)

// LayoutMode represents the responsive layout mode.
type LayoutMode int

const (
	LayoutMinimal LayoutMode = iota // < 80 cols
	LayoutStandard                   // 80-99 cols
	LayoutComfortable                // 100-119 cols
	LayoutSpacious                   // >= 120 cols
)

// Minimum terminal size required for game play
const (
	MinGameWidth  = 80
	MinGameHeight = 24
)

// PregeneratedContent contains optional pregenerated game content
type PregeneratedContent struct {
	StoryResult *engine.GenerationResult
	Rules       interface{} // *rules.RuleSet (avoid import cycle)
	Teammates   interface{} // []*npc.Teammate (avoid import cycle)
	Dream       string
}

// GamePlayModel represents the main game screen.
type GamePlayModel struct {
	// Core game state
	storyEngine   *engine.StoryEngine
	stats         *game.PlayerStats
	gameConfig    *game.GameConfig
	commandReg    *commands.Registry
	audioManager  *audio.AudioManager
	config        *config.Config

	// UI components
	viewport      viewport.Model
	textInput     textinput.Model  // Permanent free text input
	width         int
	height        int
	layoutMode    LayoutMode

	// Game state
	currentStory  string
	choices       []string
	selectedChoice int
	turnCount     int

	// Typewriter effect
	fullText      string   // Complete text received from API
	displayText   string   // Currently displayed text (for typewriter)
	charIndex     int      // Current position in typewriter animation
	isTyping      bool     // Whether typewriter is active
	streamChan    <-chan tea.Msg // Channel for receiving stream chunks

	// Status
	ready         bool
	loading       bool
	gameOver      bool
	quitConfirm   bool

	// Pregenerated content
	pregenerated  *PregeneratedContent
}

// NewGamePlayModel creates a new game play model.
// If pregenerated is provided, the game will use pregenerated content instead of generating on init.
func NewGamePlayModel(
	storyEngine *engine.StoryEngine,
	stats *game.PlayerStats,
	gameConfig *game.GameConfig,
	pregenerated *PregeneratedContent,
	audioManager *audio.AudioManager,
	config *config.Config,
) GamePlayModel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A3A3A"))

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "輸入自由文字或指令..."
	ti.Prompt = "> "
	ti.CharLimit = 200
	ti.Focus() // Auto-focus for immediate typing

	// Initialize command registry
	cmdReg := commands.NewRegistry()

	// Register audio commands
	if audioManager != nil {
		bgmCmd := commands.NewBGMCommand(audioManager, config)
		sfxCmd := commands.NewSFXCommand(audioManager, config)
		cmdReg.Register(bgmCmd)
		cmdReg.Register(sfxCmd)
	}

	m := GamePlayModel{
		storyEngine:    storyEngine,
		stats:          stats,
		gameConfig:     gameConfig,
		commandReg:     cmdReg,
		audioManager:   audioManager,
		config:         config,
		viewport:       vp,
		textInput:      ti,
		choices:        make([]string, 0),
		selectedChoice: 0,
		turnCount:      0,
		pregenerated:   pregenerated,
	}

	// If we have pregenerated story, apply it immediately
	if pregenerated != nil && pregenerated.StoryResult != nil {
		m.currentStory = pregenerated.StoryResult.Content
		m.choices = pregenerated.StoryResult.Choices
		m.ready = true
		// Note: Rules, Teammates, and Dream are stored but not directly used here
		// They were already used during parallel generation
	}

	return m
}

// Init initializes the game play model and starts opening story generation.
// If pregenerated content exists, skips generation and returns immediately.
func (m GamePlayModel) Init() tea.Cmd {
	// If we have pregenerated content, skip generation
	if m.pregenerated != nil && m.pregenerated.StoryResult != nil {
		// Start BGM for exploration mood
		if m.audioManager != nil && m.audioManager.IsInitialized() {
			bgmPlayer := m.audioManager.BGMPlayer()
			if bgmPlayer != nil && bgmPlayer.IsEnabled() {
				bgmFile := audio.GetBGMForMood(engine.MoodExploration)
				go bgmPlayer.Play(bgmFile)
			}
		}

		// Content already applied in NewGamePlayModel
		// Trigger an immediate update to show the content
		return func() tea.Msg {
			return StoryLoadingDoneMsg{
				Content:   m.currentStory,
				Error:     nil,
				Cancelled: false,
			}
		}
	}

	// Otherwise, generate opening story dynamically
	return m.generateOpeningStory()
}

// StreamDoneMsg is sent when streaming completes
type StreamDoneMsg struct {
	Choices []string
	Error   error
}

// generateOpeningStory starts the opening story generation with streaming.
func (m *GamePlayModel) generateOpeningStory() tea.Cmd {
	// Create buffered channel for streaming chunks
	chunkChan := make(chan tea.Msg, 100)
	m.streamChan = chunkChan // Store for continuous listening

	// Start generation in goroutine
	go func() {
		ctx := context.Background()

		// Stream callback - send chunks as messages
		streamCallback := func(chunk string) {
			chunkChan <- StreamChunkMsg(chunk)
		}

		// Progress callback
		progressCallback := func(percent int, state engine.EstimationState) {}

		// Generate story
		result, err := m.storyEngine.GenerateOpening(ctx, streamCallback, progressCallback)

		// Send completion message
		if err != nil {
			chunkChan <- StoryLoadingErrorMsg{Error: err}
		} else {
			chunkChan <- StreamDoneMsg{
				Choices: result.Choices,
				Error:   nil,
			}
		}
		close(chunkChan)
	}()

	// Return command that listens to channel
	return listenToStreamChannel(chunkChan)
}

// listenToStreamChannel creates a command that reads one message from the channel
func listenToStreamChannel(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			// Channel closed
			return nil
		}
		return msg
	}
}

// GameStoryMsg contains a new story segment.
type GameStoryMsg struct {
	Content string
	Choices []string
}

// StreamChunkMsg is sent when a new text chunk arrives from streaming API
type StreamChunkMsg string

// TypewriterTickMsg triggers the next character in typewriter effect
type TypewriterTickMsg struct{}

// GameOverMsg signals game over.
type GameOverMsg struct {
	Reason string
}

// Update handles messages for the game play view.
func (m GamePlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.viewport.Width = m.getViewportWidth()
		m.viewport.Height = m.getViewportHeight()
		return m, nil

	case StreamChunkMsg:
		// Received a chunk from streaming API
		m.fullText += string(msg)

		var cmd tea.Cmd
		if !m.isTyping {
			// Start typewriter effect on first chunk
			m.isTyping = true
			cmd = typewriterTickCmd()
		}

		// Continue listening to stream channel
		if m.streamChan != nil {
			return m, tea.Batch(cmd, listenToStreamChannel(m.streamChan))
		}
		return m, cmd

	case StreamDoneMsg:
		// Streaming complete, store choices and update state
		m.choices = msg.Choices
		m.streamChan = nil // Clear channel reference

		// Check for mood change and switch BGM
		if m.audioManager != nil && m.audioManager.IsInitialized() {
			newMood := parseMoodFromContent(m.fullText)
			bgmPlayer := m.audioManager.BGMPlayer()
			if bgmPlayer != nil {
				go bgmPlayer.SwitchByMood(newMood) // Async crossfade
			}
		}

		// If this was a continuation (not opening), update turn count and add separator
		if m.turnCount > 0 || m.ready {
			m.turnCount++
			m.currentStory = m.fullText // Update story history
		}

		return m, nil

	case TypewriterTickMsg:
		// Typewriter animation tick
		if m.charIndex < len([]rune(m.fullText)) {
			// Display next few characters (adjust step for speed)
			runes := []rune(m.fullText)
			step := 3 // Characters per tick
			end := m.charIndex + step
			if end > len(runes) {
				end = len(runes)
			}
			m.displayText = string(runes[:end])
			m.charIndex = end

			// Update viewport with new text
			m.viewport.SetContent(m.displayText)
			// Only call GotoBottom if viewport has been sized
			if m.viewport.Height > 0 && m.viewport.Width > 0 {
				m.viewport.GotoBottom()
			}

			return m, typewriterTickCmd()
		}
		// Typewriter finished
		m.isTyping = false
		m.loading = false // Hide loading overlay
		return m, nil

	case GameStoryMsg:
		// New story segment received (from continuation) - legacy path
		m.currentStory += "\n\n" + msg.Content
		m.choices = msg.Choices
		m.selectedChoice = 0
		m.viewport.SetContent(m.currentStory)
		if m.viewport.Height > 0 && m.viewport.Width > 0 {
			m.viewport.GotoBottom()
		}
		m.turnCount++
		m.loading = false // Hide loading overlay
		return m, nil

	case GameOverMsg:
		m.gameOver = true
		return m, nil

	case StoryLoadingDoneMsg:
		// Opening story generation complete
		if msg.Error != nil || msg.Cancelled {
			// Error or cancelled, will be handled by app level
			return m, nil
		}
		// Store the opening story content and choices
		m.currentStory = msg.Content
		m.fullText = msg.Content
		m.displayText = ""
		m.charIndex = 0
		m.ready = true

		// Start BGM after story loads (for serial generation path)
		if m.audioManager != nil && m.audioManager.IsInitialized() {
			bgmPlayer := m.audioManager.BGMPlayer()
			if bgmPlayer != nil && bgmPlayer.IsEnabled() {
				bgmFile := audio.GetBGMForMood(engine.MoodExploration)
				go bgmPlayer.Play(bgmFile)
			}
		}

		// Start typewriter effect
		m.isTyping = true
		return m, typewriterTickCmd()


	case StoryLoadingErrorMsg:
		// Error during generation, will be handled by app level
		return m, nil

	case tea.KeyMsg:
		if m.quitConfirm {
			return m.handleQuitConfirm(msg)
		}

		return m.handleKeyPress(msg)
	}

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	// Update text input
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)
	cmds = append(cmds, tiCmd)

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles all keyboard input with unified input model.
func (m GamePlayModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle scroll keys when input is empty
	if m.textInput.Value() == "" {
		switch msg.String() {
		case "up", "k":
			m.viewport.LineUp(1)
			return m, nil
		case "down", "j":
			m.viewport.LineDown(1)
			return m, nil
		case "pgup":
			m.viewport.HalfViewUp()
			return m, nil
		case "pgdown":
			m.viewport.HalfViewDown()
			return m, nil
		case "home":
			m.viewport.GotoTop()
			return m, nil
		case "end":
			if m.viewport.Height > 0 && m.viewport.Width > 0 {
				m.viewport.GotoBottom()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "q":
		// Only quit if input is empty
		if m.textInput.Value() == "" {
			m.quitConfirm = true
			return m, nil
		}
		// Otherwise, let it be typed in input

	case "enter":
		// Check if input has text
		input := m.textInput.Value()
		if input != "" {
			// Handle free text input or command
			return m.handleTextInput(input)
		}
		// If no text, select current choice
		if len(m.choices) > 0 {
			return m.handleChoiceSelected()
		}
		return m, nil

	case "esc":
		// Clear input
		m.textInput.Reset()
		return m, nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Number keys for quick choice selection (only if input is empty)
		if m.textInput.Value() == "" {
			choice := int(msg.String()[0] - '1')
			if choice < len(m.choices) {
				m.selectedChoice = choice
				return m.handleChoiceSelected()
			}
		}
		// If input has text, let number be typed
	}

	// Forward all other keys to text input
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleQuitConfirm handles quit confirmation dialog.
func (m GamePlayModel) handleQuitConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m, tea.Quit
	case "n", "N", "esc":
		m.quitConfirm = false
		return m, nil
	}
	return m, nil
}

// handleChoiceSelected processes the selected choice.
func (m GamePlayModel) handleChoiceSelected() (tea.Model, tea.Cmd) {
	if m.selectedChoice >= len(m.choices) {
		return m, nil
	}

	choice := m.choices[m.selectedChoice]

	// Add player's choice to story
	m.currentStory += fmt.Sprintf("\n\n> 你選擇: %s\n\n", choice)
	m.fullText = m.currentStory // Start new stream from current position
	m.displayText = m.currentStory
	m.charIndex = len([]rune(m.currentStory))
	m.viewport.SetContent(m.displayText)
	if m.viewport.Height > 0 && m.viewport.Width > 0 {
		m.viewport.GotoBottom()
	}

	// Set loading state and trigger story continuation
	m.loading = true

	return m, m.generateContinuation(choice)
}

// generateContinuation starts story continuation generation in background.
func (m *GamePlayModel) generateContinuation(choice string) tea.Cmd {
	// Create buffered channel for streaming chunks
	chunkChan := make(chan tea.Msg, 100)
	m.streamChan = chunkChan // Store for continuous listening

	// Start generation in goroutine
	go func() {
		ctx := context.Background()

		// Stream callback - send chunks as messages
		streamCallback := func(chunk string) {
			chunkChan <- StreamChunkMsg(chunk)
		}

		// Progress callback
		progressCallback := func(percent int, state engine.EstimationState) {}

		// Generate continuation story
		result, err := m.storyEngine.GenerateContinuation(ctx, choice, streamCallback, progressCallback)

		// Send completion message
		if err != nil {
			chunkChan <- StreamDoneMsg{
				Choices: []string{"重試"},
				Error:   err,
			}
		} else {
			chunkChan <- StreamDoneMsg{
				Choices: result.Choices,
				Error:   nil,
			}
		}
		close(chunkChan)
	}()

	// Return command that listens to channel
	return listenToStreamChannel(chunkChan)
}

// handleCommandEntered processes the entered command.
// handleTextInput handles free text input or commands.
func (m GamePlayModel) handleTextInput(input string) (tea.Model, tea.Cmd) {
	// Clear input
	m.textInput.Reset()

	// Check if it's a command (starts with /)
	if strings.HasPrefix(input, "/") {
		cmdName := input[1:] // Remove leading '/'

		// Split command and args
		parts := strings.Fields(cmdName)
		if len(parts) == 0 {
			return m, nil
		}

		cmd, exists := m.commandReg.Get(parts[0])
		if exists {
			output, err := cmd.Execute(parts[1:])
			if err != nil {
				m.currentStory += fmt.Sprintf("\n\n❌ 錯誤: %s", err.Error())
			} else {
				m.currentStory += fmt.Sprintf("\n\n%s", output)
			}
			m.viewport.SetContent(m.currentStory)
			if m.viewport.Height > 0 && m.viewport.Width > 0 {
				m.viewport.GotoBottom()
			}
		}
		return m, nil
	}

	// Free text input - treat as player action
	// Add player's action to story
	m.currentStory += fmt.Sprintf("\n\n> %s\n\n", input)
	m.fullText = m.currentStory // Start new stream from current position
	m.displayText = m.currentStory
	m.charIndex = len([]rune(m.currentStory))
	m.viewport.SetContent(m.displayText)
	if m.viewport.Height > 0 && m.viewport.Width > 0 {
		m.viewport.GotoBottom()
	}

	// Generate continuation with the free text as input
	m.loading = true
	return m, m.generateContinuation(input)
}

// updateLayout determines the current layout mode based on terminal size.
func (m *GamePlayModel) updateLayout() {
	if m.width < 80 {
		m.layoutMode = LayoutMinimal
	} else if m.width < 100 {
		m.layoutMode = LayoutStandard
	} else if m.width < 120 {
		m.layoutMode = LayoutComfortable
	} else {
		m.layoutMode = LayoutSpacious
	}
}

// getViewportHeight calculates the viewport height based on layout.
func (m *GamePlayModel) getViewportHeight() int {
	// Status bar: 3 lines
	// Choices area: 6 lines
	// Shortcut bar: 1 line
	// Borders and padding: 4 lines
	return m.height - 14
}

// getViewportWidth calculates the viewport width based on layout.
func (m *GamePlayModel) getViewportWidth() int {
	return m.width - 4
}

// View renders the game play screen.
func (m GamePlayModel) View() string {
	if !m.ready {
		return "載入中..."
	}

	// Check minimum terminal size
	if m.width < MinGameWidth || m.height < MinGameHeight {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

		return warningStyle.Render(fmt.Sprintf(
			"⚠️ 終端機太小\n\n最小尺寸: %dx%d\n目前尺寸: %dx%d\n\n請調整終端機大小。",
			MinGameWidth, MinGameHeight, m.width, m.height,
		))
	}

	if m.quitConfirm {
		return m.renderQuitConfirm()
	}

	var b strings.Builder

	// Status bar
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")

	// Narrative viewport with scroll indicators
	b.WriteString(m.renderViewportWithIndicators())
	b.WriteString("\n")

	// Show loading overlay if generating story
	if m.loading {
		b.WriteString(m.renderLoadingOverlay())
	} else {
		// Always show choices (when available)
		b.WriteString(m.renderChoices())
		b.WriteString("\n")

		// Always show free text input box
		b.WriteString(m.renderInputBox())
	}
	b.WriteString("\n")

	// Shortcut bar
	b.WriteString(m.renderShortcutBar())

	return b.String()
}

// renderLoadingOverlay renders the loading overlay during story generation.
func (m GamePlayModel) renderLoadingOverlay() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Inner border style
	innerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Border).
		Padding(1, 2)

	// Outer border style for depth
	outerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.Colors.Accent).
		Padding(1, 3).
		Width(m.width - 4).
		Align(lipgloss.Center)

	// Animated spinner frames
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := spinnerFrames[m.turnCount%len(spinnerFrames)]

	// Rotating flavor texts for in-game loading
	loadingFlavors := []string{
		"深淵正在編織你的命運...",
		"陰影在蔓延...",
		"故事的齒輪在轉動...",
		"命運的線正在纏繞...",
		"黑暗的低語迴盪著...",
	}
	flavorText := loadingFlavors[m.turnCount%len(loadingFlavors)]

	// Decorative elements
	var b strings.Builder

	// Title with spinner
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Bold(true).
		Render(frame + " 生成後續劇情中... " + frame))
	b.WriteString("\n")

	// Divider
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Border).
		Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n\n")

	// Flavor text with icon
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Render("▸ "))
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary).
		Italic(true).
		Render(flavorText))
	b.WriteString("\n\n")

	// Progress dots animation
	dots := ""
	dotCount := (m.turnCount % 4) + 1
	for i := 0; i < dotCount; i++ {
		dots += "●"
	}
	for i := dotCount; i < 4; i++ {
		dots += "○"
	}
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Align(lipgloss.Center).
		Render(dots))

	// Apply double border
	innerContent := innerStyle.Render(b.String())
	return outerStyle.Render(innerContent)
}

// renderStatusBar renders the top status bar.
func (m GamePlayModel) renderStatusBar() string {
	theme := themes.GetManager().GetCurrentTheme()

	statusStyle := lipgloss.NewStyle().
		Background(theme.Colors.Accent).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 2).
		Width(m.width)

	hpColor := theme.Colors.Success
	if m.stats.HP < 30 {
		hpColor = theme.Colors.Error
	} else if m.stats.HP < 60 {
		hpColor = theme.Colors.Warning
	}

	sanColor := theme.Colors.Accent
	if m.stats.SAN < 30 {
		sanColor = theme.Colors.Error
	} else if m.stats.SAN < 60 {
		sanColor = theme.Colors.Warning
	}

	hpBar := fmt.Sprintf("❤ HP: %d/%d", m.stats.HP, m.stats.MaxHP)
	sanBar := fmt.Sprintf("🧠 SAN: %d/%d", m.stats.SAN, m.stats.MaxSAN)
	turnInfo := fmt.Sprintf("回合: %d", m.turnCount)

	line1 := lipgloss.NewStyle().Foreground(hpColor).Render(hpBar) + "  " +
		lipgloss.NewStyle().Foreground(sanColor).Render(sanBar) + "  " +
		turnInfo

	line2 := fmt.Sprintf("精神狀態: %s | 難度: %s",
		m.stats.GetSanityState().String(),
		m.gameConfig.Difficulty.String())

	status := line1 + "\n" + line2

	return statusStyle.Render(status)
}

// renderChoices renders the choice selection area.
func (m GamePlayModel) renderChoices() string {
	if len(m.choices) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7F8C8D")).
			Padding(1, 2).
			Render("(等待選擇...)")
	}

	theme := themes.GetManager().GetCurrentTheme()

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().
		Foreground(theme.Colors.Accent).
		Bold(true).
		Render("請選擇:"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		prefix := fmt.Sprintf("  %d. ", i+1)
		style := lipgloss.NewStyle().Foreground(theme.Colors.Primary)

		if i == m.selectedChoice {
			prefix = fmt.Sprintf("❯ %d. ", i+1)
			style = style.Foreground(theme.Colors.Accent).Bold(true)
		}

		b.WriteString(prefix + style.Render(choice) + "\n")
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Border).
		Padding(0, 1).
		Width(m.width - 4).
		Render(b.String())
}

// renderInputBox renders the permanent free text input box.
func (m GamePlayModel) renderInputBox() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Render text input
	inputView := m.textInput.View()

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Border).
		Padding(0, 1).
		Width(m.width - 4).
		Render(inputView)
}

// renderViewportWithIndicators renders the viewport with scroll indicators.
func (m GamePlayModel) renderViewportWithIndicators() string {
	theme := themes.GetManager().GetCurrentTheme()

	// Get viewport content
	vpView := m.viewport.View()

	// Calculate scroll indicators
	atTop := m.viewport.YOffset == 0
	atBottom := m.viewport.AtBottom()

	var indicator string
	if !atTop && !atBottom {
		// Has content both above and below
		indicator = lipgloss.NewStyle().
			Foreground(theme.Colors.Accent).
			Render("▲ 更多 ▼")
	} else if !atTop {
		// Has content above
		indicator = lipgloss.NewStyle().
			Foreground(theme.Colors.Accent).
			Render("▲ 更多   ")
	} else if !atBottom {
		// Has content below
		indicator = lipgloss.NewStyle().
			Foreground(theme.Colors.Accent).
			Render("      ▼ 更多")
	} else {
		// All visible
		indicator = "           "
	}

	// Position indicator at bottom-right of viewport
	indicatorStyle := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(m.viewport.Width)

	return vpView + "\n" + indicatorStyle.Render(indicator)
}

// renderShortcutBar renders the bottom shortcut bar.
func (m GamePlayModel) renderShortcutBar() string {
	theme := themes.GetManager().GetCurrentTheme()

	shortcuts := "1-3: 快速選擇 | 打字: 自由輸入 | Enter: 確認 | Esc: 清空輸入 | ↑↓: 滾動 | q: 離開"

	return lipgloss.NewStyle().
		Foreground(theme.Colors.Secondary).
		Padding(0, 2).
		Width(m.width).
		Render(shortcuts)
}

// renderQuitConfirm renders the quit confirmation dialog.
func (m GamePlayModel) renderQuitConfirm() string {
	theme := themes.GetManager().GetCurrentTheme()

	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Error).
		Padding(2, 4).
		Align(lipgloss.Center)

	content := lipgloss.NewStyle().
		Foreground(theme.Colors.Error).
		Bold(true).
		Render("確定要離開遊戲嗎？") + "\n\n" +
		lipgloss.NewStyle().
		Foreground(theme.Colors.Primary).
		Render("(y) 是  (n) 否")

	return confirmStyle.Render(content)
}

// typewriterTickCmd returns a command that triggers the next typewriter tick
func typewriterTickCmd() tea.Cmd {
	return tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
		return TypewriterTickMsg{}
	})
}

// parseMoodFromContent extracts mood from story text using keyword matching
func parseMoodFromContent(content string) engine.MoodType {
	contentLower := strings.ToLower(content)

	// Priority: Death > Sanity > Horror > Chase > Tension > Ritual > Dream > Mystery > Safe > Exploration

	// Death/Ending
	if strings.Contains(contentLower, "死亡") || strings.Contains(contentLower, "death") ||
		strings.Contains(contentLower, "結束") || strings.Contains(contentLower, "end") {
		return engine.MoodEnding
	}

	// Sanity Collapse
	if strings.Contains(contentLower, "瘋狂") || strings.Contains(contentLower, "madness") ||
		strings.Contains(contentLower, "理智") || strings.Contains(contentLower, "sanity") ||
		strings.Contains(contentLower, "崩潰") || strings.Contains(contentLower, "collapse") {
		return engine.MoodSanity
	}

	// Horror
	if strings.Contains(contentLower, "恐怖") || strings.Contains(contentLower, "horror") ||
		strings.Contains(contentLower, "驚悚") || strings.Contains(contentLower, "terror") {
		return engine.MoodHorror
	}

	// Chase/Escape
	if strings.Contains(contentLower, "追逐") || strings.Contains(contentLower, "chase") ||
		strings.Contains(contentLower, "逃跑") || strings.Contains(contentLower, "escape") ||
		strings.Contains(contentLower, "快跑") || strings.Contains(contentLower, "run") {
		return engine.MoodChase
	}

	// Tension
	if strings.Contains(contentLower, "緊張") || strings.Contains(contentLower, "tension") ||
		strings.Contains(contentLower, "危險") || strings.Contains(contentLower, "danger") {
		return engine.MoodTension
	}

	// Ritual/Occult
	if strings.Contains(contentLower, "儀式") || strings.Contains(contentLower, "ritual") ||
		strings.Contains(contentLower, "邪教") || strings.Contains(contentLower, "cult") ||
		strings.Contains(contentLower, "召喚") || strings.Contains(contentLower, "summon") {
		return engine.MoodRitual
	}

	// Dream/Surreal
	if strings.Contains(contentLower, "夢境") || strings.Contains(contentLower, "dream") ||
		strings.Contains(contentLower, "迷幻") || strings.Contains(contentLower, "surreal") ||
		strings.Contains(contentLower, "幻覺") || strings.Contains(contentLower, "hallucination") {
		return engine.MoodDream
	}

	// Mystery/Puzzle
	if strings.Contains(contentLower, "謎題") || strings.Contains(contentLower, "puzzle") ||
		strings.Contains(contentLower, "線索") || strings.Contains(contentLower, "clue") {
		return engine.MoodMystery
	}

	// Safe/Rest
	if strings.Contains(contentLower, "安全") || strings.Contains(contentLower, "safe") ||
		strings.Contains(contentLower, "休息") || strings.Contains(contentLower, "rest") {
		return engine.MoodSafe
	}

	return engine.MoodExploration // Default
}
