// Package app provides the main application model for Nightmare Assault.
package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/styles"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/views"
	"github.com/nightmare-assault/nightmare-assault/internal/update"
)

// MinWidth is the minimum terminal width required
const MinWidth = 80

// MinHeight is the minimum terminal height required
const MinHeight = 24

// AppState represents the current application state.
type AppState int

const (
	StateLoading AppState = iota
	StateAPISetup
	StateMainMenu
	StateSettings
	StateThemeSelector
	StateGameSetup
	StateGame
)

// Model represents the main application state.
type Model struct {
	version        string
	width          int
	height         int
	ready          bool
	state          AppState
	prevState      AppState
	config         *config.Config
	gameConfig     *game.GameConfig // Game configuration from setup flow
	apiSetup       views.APISetupModel
	mainMenu       views.MainMenuModel
	settingsMenu   views.SettingsMenuModel
	themeSelector  views.ThemeSelectorModel
	gameSetup      views.GameSetupModel
	hasSaveFiles   bool
	updateManager  *update.Manager
	updateChecked  bool
}

// New creates a new application Model.
func New(version string) Model {
	// 初始化更新管理器
	updateConfig := update.UpdateConfig{
		Owner:          "nightmare-assault",
		Repo:           "nightmare-assault",
		CurrentVersion: version,
		CheckInterval:  24 * time.Hour,
		CacheDir:       "",
	}

	updateMgr, _ := update.NewManager(updateConfig)

	return Model{
		version:       version,
		state:         StateLoading,
		updateManager: updateMgr,
		updateChecked: false,
	}
}

// Init initializes the application.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit handling with Ctrl+C
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle ESC in game state to return to main menu
		if m.state == StateGame && msg.String() == "esc" {
			m.state = StateMainMenu
			m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Initialize state after we have window size
		if m.state == StateLoading {
			// Load config (use default if error, logged for debugging)
			cfg, err := config.Load()
			if err != nil {
				// Config load failed - use defaults (this is normal on first run)
				// Error is typically "file not found" which is expected
				cfg = config.DefaultConfig()
				// Detect system language for first run
				cfg.Language = i18n.DetectSystemLanguage()
			}
			m.config = cfg

			// Initialize i18n with configured language
			if err := i18n.InitGlobal(cfg.Language); err != nil {
				// Fallback to English if init fails
				i18n.InitGlobal("en-US")
			}

			// Check for save files (placeholder - returns false for now)
			m.hasSaveFiles = false

			if !cfg.IsConfigured() {
				m.state = StateAPISetup
				m.apiSetup = views.NewAPISetupModel(cfg)
				return m, m.apiSetup.Init()
			} else {
				m.state = StateMainMenu
				m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
				// 啟動背景更新檢查
				return m, checkForUpdates(m.updateManager)
			}
		}

		// Pass size to sub-models
		return m.passWindowSize(msg)

	case views.MenuSelectMsg:
		return m.handleMenuSelect(msg)

	case views.SettingsSelectMsg:
		return m.handleSettingsSelect(msg)

	case views.ThemeSelectedMsg:
		// Theme was applied, save to config
		m.config.Theme = msg.ThemeID
		m.config.Save()
		m.state = StateSettings
		return m, nil

	case views.ThemeBackMsg:
		m.state = StateSettings
		return m, nil

	case updateCheckResultMsg:
		// 更新檢查結果
		if msg.result != nil && msg.result.Status == update.UpdateStatusAvailable {
			// 發送更新可用消息到主選單
			return m, func() tea.Msg {
				return views.UpdateAvailableMsg{NewVersion: msg.result.NewVersion}
			}
		}
		return m, nil

	case views.GameSetupDoneMsg:
		if msg.Cancelled {
			// Return to main menu
			m.state = StateMainMenu
			m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
			return m, nil
		}
		// Setup complete - store config, freeze it, and start game
		m.gameConfig = msg.Config
		m.gameConfig.Freeze() // Make config immutable once game starts
		m.state = StateGame
		// TODO: Initialize game engine with m.gameConfig (Story 2-2 integration)
		return m, nil
	}

	// Delegate to current state
	return m.updateCurrentState(msg)
}

func (m Model) passWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case StateAPISetup:
		apiModel, c := m.apiSetup.Update(msg)
		m.apiSetup = apiModel.(views.APISetupModel)
		cmd = c
	case StateMainMenu:
		menuModel, c := m.mainMenu.Update(msg)
		m.mainMenu = menuModel.(views.MainMenuModel)
		cmd = c
	case StateSettings:
		settingsModel, c := m.settingsMenu.Update(msg)
		m.settingsMenu = settingsModel.(views.SettingsMenuModel)
		cmd = c
	case StateThemeSelector:
		themeModel, c := m.themeSelector.Update(msg)
		m.themeSelector = themeModel.(views.ThemeSelectorModel)
		cmd = c
	case StateGameSetup:
		setupModel, c := m.gameSetup.Update(msg)
		m.gameSetup = setupModel.(views.GameSetupModel)
		cmd = c
	}
	return m, cmd
}

func (m Model) updateCurrentState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case StateAPISetup:
		apiModel, c := m.apiSetup.Update(msg)
		m.apiSetup = apiModel.(views.APISetupModel)
		cmd = c

		// Check if setup is done
		if m.apiSetup.IsDone() {
			m.config = m.apiSetup.GetConfig()
			m.state = StateMainMenu
			m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
		}

	case StateMainMenu:
		menuModel, c := m.mainMenu.Update(msg)
		m.mainMenu = menuModel.(views.MainMenuModel)
		cmd = c

	case StateSettings:
		settingsModel, c := m.settingsMenu.Update(msg)
		m.settingsMenu = settingsModel.(views.SettingsMenuModel)
		cmd = c

	case StateThemeSelector:
		themeModel, c := m.themeSelector.Update(msg)
		m.themeSelector = themeModel.(views.ThemeSelectorModel)
		cmd = c

	case StateGameSetup:
		setupModel, c := m.gameSetup.Update(msg)
		m.gameSetup = setupModel.(views.GameSetupModel)
		cmd = c
	}

	return m, cmd
}

func (m Model) handleMenuSelect(msg views.MenuSelectMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case views.ActionNewGame:
		// Start new game setup flow
		m.state = StateGameSetup
		m.gameSetup = views.NewGameSetupModel()
		return m, m.gameSetup.Init()

	case views.ActionContinue:
		// TODO: Load save (Epic 5)
		return m, nil

	case views.ActionSettings:
		m.prevState = StateMainMenu
		m.state = StateSettings
		m.settingsMenu = views.NewSettingsMenuModel()
		return m, nil

	case views.ActionExit:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleSettingsSelect(msg views.SettingsSelectMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case views.SettingsActionTheme:
		m.prevState = StateSettings
		m.state = StateThemeSelector
		m.themeSelector = views.NewThemeSelectorModel()
		return m, nil

	case views.SettingsActionAPI:
		m.prevState = StateSettings
		m.state = StateAPISetup
		m.apiSetup = views.NewAPISetupModel(m.config)
		return m, m.apiSetup.Init()

	case views.SettingsActionAudio:
		// TODO: Audio settings (Epic 6)
		return m, nil

	case views.SettingsActionBack:
		m.state = StateMainMenu
		return m, nil
	}

	return m, nil
}

// View renders the application view.
func (m Model) View() string {
	if !m.ready {
		return "載入中..."
	}

	// Check minimum terminal size
	if m.width < MinWidth || m.height < MinHeight {
		return styles.Warning.Render(fmt.Sprintf(
			"⚠️ 終端機太小\n\n最小尺寸: %dx%d\n目前尺寸: %dx%d\n\n請調整終端機大小。",
			MinWidth, MinHeight, m.width, m.height,
		))
	}

	// Render based on state
	switch m.state {
	case StateAPISetup:
		return m.apiSetup.View()

	case StateMainMenu:
		return m.mainMenu.View()

	case StateSettings:
		return m.settingsMenu.View()

	case StateThemeSelector:
		return m.themeSelector.View()

	case StateGameSetup:
		return m.gameSetup.View()

	case StateGame:
		return m.renderGamePlaceholder()

	default:
		return "載入中..."
	}
}

func (m Model) renderGamePlaceholder() string {
	content := styles.Title.Render("🎮 遊戲進行中")
	content += "\n\n"
	content += styles.Text.Render("(遊戲功能將在 Epic 2 實作)")
	content += "\n\n"
	content += styles.Hint.Render("按 ESC 返回主選單，按 q 離開")
	return styles.Container.Render(content)
}

// Width returns the current terminal width.
func (m Model) Width() int {
	return m.width
}

// Height returns the current terminal height.
func (m Model) Height() int {
	return m.height
}

// State returns the current application state.
func (m Model) State() AppState {
	return m.state
}

// updateCheckResultMsg 更新檢查結果消息
type updateCheckResultMsg struct {
	result *update.UpdateResult
	err    error
}

// checkForUpdates 背景檢查更新
func checkForUpdates(manager *update.Manager) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return updateCheckResultMsg{err: fmt.Errorf("update manager not initialized")}
		}

		// 檢查是否應該檢查更新（基於時間間隔）
		if !manager.ShouldCheckForUpdates() {
			return updateCheckResultMsg{result: nil, err: nil}
		}

		// 執行更新檢查
		result, err := manager.CheckForUpdates()

		// 記錄檢查時間
		manager.RecordUpdateCheck()

		return updateCheckResultMsg{result: result, err: err}
	}
}
