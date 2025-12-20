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
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
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
	choices          []string // Legacy: kept for backward compatibility
	choiceSituation  string   // Current choice context: situation
	choiceQuestion   string   // Current choice context: question (optional)
	choiceOptions    []string // Current choice context: options
	selectedChoice   int
	turnCount        int

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
	isPrologue    bool     // Whether currently showing prologue (序章)

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
	Content         string   // Parsed story content (not raw JSON)
	Choices         []string // Legacy: kept for backward compatibility
	ChoiceSituation string   // Choice context: situation
	ChoiceQuestion  string   // Choice context: question (optional)
	ChoiceOptions   []string // Choice context: options
	HPChange        int      // HP change from state_changes
	SANChange       int      // SAN change from state_changes
	ChangeReason    string   // Reason for state changes
	Error           error
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
				Content:         result.Content, // Send parsed content, not raw JSON
				Choices:         result.Choices, // Legacy support
				ChoiceSituation: result.ChoiceSituation,
				ChoiceQuestion:  result.ChoiceQuestion,
				ChoiceOptions:   result.ChoiceOptions,
				HPChange:        result.HPChange,
				SANChange:       result.SANChange,
				ChangeReason:    result.ChangeReason,
				Error:           nil,
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

// ConfigReloadedMsg is sent when config reload is attempted.
// Story 10-7: Config Hot Reload
type ConfigReloadedMsg struct {
	Success bool
	Error   error
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
		// Received a chunk from streaming API - just accumulate, don't display yet
		m.fullText += string(msg)

		// Continue listening to stream channel (no typewriter during streaming)
		if m.streamChan != nil {
			return m, listenToStreamChannel(m.streamChan)
		}
		return m, nil

	case StreamDoneMsg:
		// Streaming complete, store choices and update state
		m.choices = msg.Choices // Legacy support
		m.choiceSituation = msg.ChoiceSituation
		m.choiceQuestion = msg.ChoiceQuestion
		m.choiceOptions = msg.ChoiceOptions
		m.streamChan = nil // Clear channel reference

		// Apply state changes (HP/SAN)
		if msg.HPChange != 0 || msg.SANChange != 0 {
			m.stats.HP += msg.HPChange
			m.stats.SAN += msg.SANChange

			// Clamp values
			if m.stats.HP > 100 {
				m.stats.HP = 100
			}
			if m.stats.HP < 0 {
				m.stats.HP = 0
			}
			if m.stats.SAN > 100 {
				m.stats.SAN = 100
			}
			if m.stats.SAN < 0 {
				m.stats.SAN = 0
			}

			// TODO: Show state change notification with reason (msg.ChangeReason)
		}

		// Replace raw JSON with parsed content and append to history
		if msg.Content != "" {
			// Clean HTML/XML tags from content
			cleanContent := stripHTMLTags(msg.Content)

			// For first story (opening), replace; for continuations, append
			if m.currentStory == "" {
				// First story (opening/prologue)
				m.currentStory = cleanContent
				m.fullText = cleanContent
				// Start typewriter from beginning for first story
				m.charIndex = 0
				m.displayText = ""
			} else {
				// Continuation - append to history
				oldStory := m.currentStory
				m.currentStory += "\n\n" + cleanContent
				m.fullText = m.currentStory

				// For continuation, start typewriter from end of old content
				// This way only new content has typewriter effect
				m.displayText = oldStory
				m.charIndex = len([]rune(oldStory))
			}
		}

		// Check if this is prologue (no choices and contains prologue marker)
		if len(msg.ChoiceOptions) == 0 && strings.Contains(msg.Content, "按任意鍵") {
			m.isPrologue = true
		} else {
			m.isPrologue = false
		}

		// Check for mood change and switch BGM
		if m.audioManager != nil && m.audioManager.IsInitialized() {
			newMood := parseMoodFromContent(m.fullText)
			bgmPlayer := m.audioManager.BGMPlayer()
			if bgmPlayer != nil {
				go bgmPlayer.SwitchByMood(newMood) // Async crossfade
			}
		}

		// Update turn count for continuations
		if m.turnCount > 0 || m.ready {
			m.turnCount++
		}

		// Start typewriter effect to display the parsed content
		m.isTyping = true
		return m, typewriterTickCmd()

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

			// Update viewport with new text (with word wrapping)
			m.safeSetViewportContent(m.displayText, true)

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
		m.safeSetViewportContent(m.currentStory, true)
		m.turnCount++
		m.loading = false // Hide loading overlay
		return m, nil

	case GameOverMsg:
		m.gameOver = true
		return m, nil

	case ConfigReloadedMsg:
		// Story 10-7 AC2: Display success/error message to user
		// Story 10-7 AC3: On error, original config is preserved (handled by config.Reload())
		t := i18n.GetGlobal()
		systemLabel := "System"
		if t != nil {
			systemLabel = t.T("game.system_label")
		}
		if msg.Success {
			successMsg := "Config reloaded successfully"
			if t != nil {
				successMsg = t.T("messages.config_reload_success")
			}
			m.currentStory += fmt.Sprintf("\n\n[%s] ✓ %s", systemLabel, successMsg)
		} else {
			errorDetail := "unknown error"
			if msg.Error != nil {
				errorDetail = msg.Error.Error()
			}
			failedMsg := fmt.Sprintf("Config reload failed: %s", errorDetail)
			keepOriginal := "Keeping original config"
			if t != nil {
				failedMsg = t.T("messages.config_reload_failed", errorDetail)
				keepOriginal = t.T("messages.config_reload_keep_original")
			}
			m.currentStory += fmt.Sprintf("\n\n[%s] ✗ %s\n%s", systemLabel, failedMsg, keepOriginal)
		}
		m.fullText = m.currentStory
		m.displayText = m.currentStory
		m.charIndex = len([]rune(m.currentStory))
		m.safeSetViewportContent(m.displayText, true)
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

		// Check if this is prologue (no choices and contains prologue marker)
		if len(m.choices) == 0 && strings.Contains(msg.Content, "按任意鍵") {
			m.isPrologue = true
		}

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

	// Special handling for prologue end (序章結束)
	// Any key press (except q) should continue to Chapter 1
	if m.isPrologue && !m.loading && !m.isTyping {
		// Skip quit key in prologue
		if msg.String() == "q" {
			m.quitConfirm = true
			return m, nil
		}

		// Any other key continues to Chapter 1
		m.isPrologue = false
		m.loading = true

		// Add transition message to story
		m.currentStory += "\n\n【進入第一章】\n\n"
		m.fullText = m.currentStory
		m.displayText = m.currentStory
		m.charIndex = len([]rune(m.currentStory))
		m.safeSetViewportContent(m.displayText, true)

		// Generate first chapter continuation
		return m, m.generateContinuation("開始冒險")
	}

	// Story 10-7: Config Hot Reload - Ctrl+R to reload config
	if msg.String() == "ctrl+r" {
		return m, m.reloadConfig()
	}

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
	// Use choiceOptions if available, fallback to legacy choices
	options := m.choiceOptions
	if len(options) == 0 {
		options = m.choices // Fallback to legacy
	}

	if m.selectedChoice >= len(options) {
		return m, nil
	}

	choice := options[m.selectedChoice]

	// DO NOT add "> 你選擇:" meta-narrative here
	// The LLM will integrate the choice naturally into the story narrative
	// Just show a separator before loading
	m.currentStory += "\n\n───────────────────\n\n"
	m.fullText = m.currentStory
	m.displayText = m.currentStory
	m.charIndex = len([]rune(m.currentStory))
	m.safeSetViewportContent(m.displayText, true)

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
				Content:       "",
				Choices:       []string{"重試"},
				ChoiceOptions: []string{"重試"},
				Error:         err,
			}
		} else {
			chunkChan <- StreamDoneMsg{
				Content:         result.Content, // Send parsed content, not raw JSON
				Choices:         result.Choices, // Legacy support
				ChoiceSituation: result.ChoiceSituation,
				ChoiceQuestion:  result.ChoiceQuestion,
				ChoiceOptions:   result.ChoiceOptions,
				HPChange:        result.HPChange,
				SANChange:       result.SANChange,
				ChangeReason:    result.ChangeReason,
				Error:           nil,
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
			m.safeSetViewportContent(m.currentStory, true)
		}
		return m, nil
	}

	// Free text input - treat as player action
	// Add player's action to story
	m.currentStory += fmt.Sprintf("\n\n> %s\n\n", input)
	m.fullText = m.currentStory // Start new stream from current position
	m.displayText = m.currentStory
	m.charIndex = len([]rune(m.currentStory))
	m.safeSetViewportContent(m.displayText, true)

	// Generate continuation with the free text as input
	m.loading = true
	return m, m.generateContinuation(input)
}

// reloadConfig attempts to reload the config file.
// Story 10-7 AC1: Press Ctrl+R to reload config from ~/.nightmare/config.json
// Story 10-7 AC2: New config takes effect immediately
// Story 10-7 AC3: On error, show error message and keep original config
// Story 10-7 AC4: Log detailed errors to log file
func (m *GamePlayModel) reloadConfig() tea.Cmd {
	return func() tea.Msg {
		// Attempt to reload config
		err := m.config.Reload()

		// Story 10-7 AC4: Log the reload attempt (English for log file consistency)
		if logger := logger.GetGlobal(); logger != nil {
			if err != nil {
				logger.Error("Config hot reload failed", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				logger.Info("Config hot reload successful", nil)
			}
		}

		return ConfigReloadedMsg{
			Success: err == nil,
			Error:   err,
		}
	}
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

// renderChoices renders the choice selection area with context.
func (m GamePlayModel) renderChoices() string {
	// Use choiceOptions if available, fallback to legacy choices
	options := m.choiceOptions
	if len(options) == 0 {
		options = m.choices // Fallback to legacy
	}

	if len(options) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7F8C8D")).
			Padding(1, 2).
			Render("(等待選擇...)")
	}

	theme := themes.GetManager().GetCurrentTheme()
	var b strings.Builder

	// Separator line
	separatorStyle := lipgloss.NewStyle().
		Foreground(theme.Colors.Border)
	separator := strings.Repeat("─", m.width-6)
	b.WriteString(separatorStyle.Render(separator))
	b.WriteString("\n\n")

	// Situation (if available)
	if m.choiceSituation != "" {
		situationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Italic(true)
		b.WriteString(situationStyle.Render(m.choiceSituation))
		b.WriteString("\n\n")
	}

	// Question (if available)
	if m.choiceQuestion != "" {
		questionStyle := lipgloss.NewStyle().
			Foreground(theme.Colors.Accent).
			Bold(true)
		b.WriteString(questionStyle.Render(m.choiceQuestion))
		b.WriteString("\n\n")
	}

	// Options
	for i, option := range options {
		prefix := fmt.Sprintf("  [%d] ", i+1)
		style := lipgloss.NewStyle().Foreground(theme.Colors.Primary)

		if i == m.selectedChoice {
			prefix = fmt.Sprintf("❯ [%d] ", i+1)
			style = style.Foreground(theme.Colors.Accent).Bold(true)
		}

		b.WriteString(prefix + style.Render(option) + "\n")
	}

	return lipgloss.NewStyle().
		Padding(0, 2).
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

	// Story 10-7: Add Ctrl+R hint for config reload (i18n)
	shortcuts := "1-3: Quick select | Type: Free input | Enter: Confirm | Esc: Clear | ↑↓: Scroll | Ctrl+R: Reload config | q: Quit"
	if t := i18n.GetGlobal(); t != nil {
		shortcuts = t.T("hints.game_shortcut_bar")
	}

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

// safeSetViewportContent safely sets viewport content and scrolls to bottom.
// Only updates if viewport has valid dimensions.
func (m *GamePlayModel) safeSetViewportContent(text string, scrollToBottom bool) {
	// Skip if viewport has invalid dimensions
	if m.viewport.Width <= 0 || m.viewport.Height <= 0 {
		return
	}

	// Wrap text and set content
	wrappedText := wrapText(text, m.viewport.Width)
	m.viewport.SetContent(wrappedText)

	// Scroll to bottom if requested
	if scrollToBottom {
		m.viewport.GotoBottom()
	}
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

// wrapText wraps text to fit within the specified width, handling CJK characters correctly.
// Adds right margin for better readability.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	// Reserve right margin (10% of width or minimum 8 characters)
	rightMargin := width / 10
	if rightMargin < 8 {
		rightMargin = 8
	}
	effectiveWidth := width - rightMargin

	if effectiveWidth <= 20 {
		effectiveWidth = width // Don't apply margin if width is too small
	}

	var lines []string
	currentLine := strings.Builder{}
	currentWidth := 0

	for _, r := range text {
		charWidth := runeWidth(r)

		if r == '\n' {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
			continue
		}

		if currentWidth+charWidth > effectiveWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}

		currentLine.WriteRune(r)
		currentWidth += charWidth
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// runeWidth returns the display width of a rune.
func runeWidth(r rune) int {
	// ASCII characters are 1 wide
	if r < 128 {
		return 1
	}

	// CJK characters are typically 2 wide
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
		return 2
	}
	if r >= 0x3400 && r <= 0x4DBF { // CJK Extension A
		return 2
	}
	if r >= 0x3040 && r <= 0x309F { // Hiragana
		return 2
	}
	if r >= 0x30A0 && r <= 0x30FF { // Katakana
		return 2
	}

	// Emoji are typically 2 wide
	if r >= 0x1F300 && r <= 0x1F9FF {
		return 2
	}
	// Additional emoji ranges
	if r >= 0x2600 && r <= 0x27BF { // Miscellaneous Symbols
		return 2
	}

	// Default to 1
	return 1
}

// stripHTMLTags removes all HTML/XML tags and comments from text.
func stripHTMLTags(text string) string {
	// First, remove HTML comments <!-- ... -->
	for {
		start := strings.Index(text, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "-->")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+3:]
	}

	// Then remove regular tags <...>
	var result strings.Builder
	inTag := false

	for _, r := range text {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}
