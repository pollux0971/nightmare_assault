// Package app provides the main application model for Nightmare Assault.
package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/audio"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/debug"
	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/parallel"
	"github.com/nightmare-assault/nightmare-assault/internal/engine/rules"
	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/styles"
	"github.com/nightmare-assault/nightmare-assault/internal/tui/views"
	"github.com/nightmare-assault/nightmare-assault/internal/update"
)

// AppOption 配置應用程式模型
type AppOption func(*Model)

// WithUpdateManager 設定更新管理器（可選）
// 如果為 nil，自動更新功能將被停用
func WithUpdateManager(mgr *update.Manager) AppOption {
	return func(m *Model) {
		m.updateManager = mgr
	}
}

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
	StateMaxTokensSettings
	StateAudioSettings
	StateGameSetup
	StateStoryLoading
	StateParallelLoading // NEW: Parallel generation loading
	StateGame
)

// Model represents the main application state.
type Model struct {
	version         string
	width           int
	height          int
	ready           bool
	state           AppState
	prevState       AppState
	config          *config.Config
	gameConfig      *game.GameConfig // Game configuration from setup flow
	apiSetup          views.APISetupInterface
	mainMenu          views.MainMenuModel
	settingsMenu      views.SettingsMenuModel
	themeSelector     views.ThemeSelectorModel
	maxtokensSettings views.MaxTokensSettingsModel
	audioSettings     views.AudioSettingsModel
	gameSetup         views.GameSetupModel
	storyLoading    views.StoryLoadingModel
	parallelLoading views.ParallelLoadingModel // NEW: Parallel generation UI
	gamePlay        views.GamePlayModel
	hasSaveFiles    bool
	updateManager   *update.Manager
	updateChecked   bool
	audioManager    *audio.AudioManager
}

// New 創建新的應用程式 Model，支援可選配置
// 預設情況下，update manager 為 nil（自動更新停用）
// 生產環境使用 NewWithUpdateManager() 或傳遞 WithUpdateManager() 選項
func New(version string, opts ...AppOption) Model {
	m := Model{
		version:       version,
		state:         StateLoading,
		updateManager: nil, // 預設為 nil - 無自動更新
		updateChecked: false,
	}

	// 應用選項
	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// NewWithUpdateManager 創建支援自動更新的應用程式 Model
// 這是生產環境建構子，嘗試初始化 update manager
// 如果初始化失敗，應用程式繼續運行但無自動更新（記錄錯誤）
func NewWithUpdateManager(version string) Model {
	// 嘗試初始化 update manager
	updateConfig := update.UpdateConfig{
		Owner:          "nightmare-assault",
		Repo:           "nightmare-assault",
		CurrentVersion: version,
		CheckInterval:  24 * time.Hour,
		CacheDir:       "",
	}

	updateMgr, err := update.NewManager(updateConfig)
	if err != nil {
		// 記錄錯誤但繼續 - update manager 是可選的
		// 測試期間通常會失敗，因為 os.Executable() 在測試二進制中
		debug.Log("Failed to initialize update manager: %v (auto-updates disabled)", err)
		updateMgr = nil
	}

	return New(version, WithUpdateManager(updateMgr))
}

// Init initializes the application.
func (m Model) Init() tea.Cmd {
	// Request terminal size immediately on startup
	return tea.WindowSize()
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

			// Initialize debug mode
			if err := debug.Initialize(cfg.Debug.Enabled, cfg.Debug.LogAPIKeys, cfg.Debug.LogRequests); err != nil {
				// Debug initialization failed, but this is not critical
				debug.Log("Failed to initialize debug logger: %v", err)
			}
			debug.Log("=== Application Started ===")
			debug.Log("Version: %s", m.version)
			debug.Log("Debug Mode: %v", cfg.Debug.Enabled)

			// Initialize i18n with configured language
			if err := i18n.InitGlobal(cfg.Language); err != nil {
				// Fallback to English if init fails
				i18n.InitGlobal("en-US")
			}

			// Initialize audio manager
			m.audioManager = audio.NewAudioManager(cfg.Audio)
			if err := m.audioManager.Initialize(); err != nil {
				debug.Log("Audio system unavailable: %v (game will run in silent mode)", err)
			} else {
				debug.Log("Audio system initialized successfully")
				// Start menu BGM immediately after initialization
				if bgmPlayer := m.audioManager.BGMPlayer(); bgmPlayer != nil && bgmPlayer.IsEnabled() {
					go bgmPlayer.Play(audio.GetBGMFilename(audio.BGMSceneMystery))
				}
			}

			// Check for save files (placeholder - returns false for now)
			m.hasSaveFiles = false

			if !cfg.IsConfigured() {
				m.state = StateAPISetup
				m.apiSetup = views.NewTrinitySetupModel(cfg)
				return m, m.apiSetup.Init()
			} else {
				m.state = StateMainMenu
				m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
				// 只有 manager 初始化且配置啟用時才檢查更新
				if m.updateManager != nil && m.config.Update.Enabled {
					return m, checkForUpdates(m.updateManager, m.config.Update.CheckInterval)
				}
				return m, nil
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

	case views.MaxTokensSavedMsg:
		// MaxTokens was saved, go back to settings
		m.state = StateSettings
		return m, nil

	case views.AudioSettingsSavedMsg:
		// Audio settings were saved, go back to settings
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
		// Setup complete - store config, freeze it, and start parallel generation
		m.gameConfig = msg.Config
		m.gameConfig.Freeze() // Make config immutable once game starts

		// Create provider (use configured provider for generation)
		providerID := m.config.API.Provider.ProviderID
		encryptedKey := m.config.API.APIKeys[providerID]
		apiKey, err := m.config.DecryptAPIKey(providerID)

		// Debug logging
		debug.LogAPIKeyInfo(providerID, encryptedKey, apiKey, err)

		if err != nil {
			debug.LogError("Parallel Generation Setup", "Failed to decrypt API key", err)
			// Fallback to serial generation on decryption error
			return m.fallbackSerialGeneration()
		}

		providerCfg := api.ProviderConfig{
			ProviderID: providerID,
			APIKey:     apiKey,
			// BaseURL is hardcoded in provider.go, don't override from config
			Model:     m.config.API.Provider.Model,
			MaxTokens: m.config.API.Provider.MaxTokens,
		}

		debug.LogProviderConfig(providerID, "", m.config.API.Provider.Model, m.config.API.Provider.MaxTokens, config.MaskAPIKey(apiKey))
		debug.WorkLogAPICall(providerID, m.config.API.Provider.Model, "Parallel Generation Setup")

		smartProvider, err := api.NewProvider(providerCfg)
		if err != nil {
			debug.WorkLogError("Provider Init", "Failed to create provider", err)
			// Fallback to serial generation on provider error
			return m.fallbackSerialGeneration()
		}

		// Build parallel generation tasks
		debug.WorkLog("Building parallel generation tasks...")
		tasks, err := parallel.BuildGenerationTasks(m.gameConfig, smartProvider)
		if err != nil {
			debug.WorkLogError("Task Building", "Failed to build parallel tasks", err)
			// Fallback to serial generation on build error
			return m.fallbackSerialGeneration()
		}
		debug.WorkLog("Built %d parallel tasks", len(tasks))

		// Create coordinator
		coordConfig := parallel.DefaultCoordinatorConfig()
		coordinator := parallel.NewCoordinator(coordConfig)

		// Register tasks
		for _, task := range tasks {
			if err := coordinator.AddTask(task); err != nil {
				// Fallback to serial generation on task registration error
				return m.fallbackSerialGeneration()
			}
		}

		// Show parallel loading UI
		m.parallelLoading = views.NewParallelLoadingModel(coordinator)
		m.state = StateParallelLoading

		// Start parallel generation
		return m, tea.Batch(
			m.parallelLoading.Init(),
			parallel.StartGeneration(coordinator),
		)

	case parallel.ParallelLoadingDoneMsg:
		if msg.Cancelled {
			// User cancelled - return to main menu
			m.state = StateMainMenu
			m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
			return m, nil
		}

		// Extract results from parallel generation
		results := msg.Results
		storyResult, ok := results[parallel.TaskIDStory].(*engine.GenerationResult)
		if !ok {
			// Story generation failed - return to menu
			m.state = StateMainMenu
			m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
			return m, nil
		}

		ruleSet, _ := results[parallel.TaskIDRules].(*rules.RuleSet)
		teammates, _ := results[parallel.TaskIDTeammates].([]*npc.Teammate)
		dreamContent, _ := results[parallel.TaskIDDream].(string)

		// Initialize game components
		stats := game.NewPlayerStats()

		// Create provider and story engine
		providerID := m.config.API.Provider.ProviderID
		encryptedKey := m.config.API.APIKeys[providerID]
		apiKey, err := m.config.DecryptAPIKey(providerID)

		// Debug logging
		debug.LogAPIKeyInfo(providerID, encryptedKey, apiKey, err)

		providerCfg := api.ProviderConfig{
			ProviderID: providerID,
			APIKey:     apiKey,
			// BaseURL is hardcoded in provider.go, don't override from config
			Model:     m.config.API.Provider.Model,
			MaxTokens: m.config.API.Provider.MaxTokens,
		}

		debug.LogProviderConfig(providerID, "", m.config.API.Provider.Model, m.config.API.Provider.MaxTokens, config.MaskAPIKey(apiKey))
		smartProvider, _ := api.NewProvider(providerCfg)

		engineConfig := engine.DefaultEngineConfig()
		engineConfig.Provider = smartProvider
		engineConfig.GameConfig = m.gameConfig
		storyEngine := engine.NewStoryEngine(engineConfig)

		// Create pregenerated content
		pregenerated := &views.PregeneratedContent{
			StoryResult: storyResult,
			Rules:       ruleSet,
			Teammates:   teammates,
			Dream:       dreamContent,
		}

		// Create game play model with pregenerated content
		m.gamePlay = views.NewGamePlayModel(storyEngine, stats, m.gameConfig, pregenerated, m.audioManager, m.config)
		m.state = StateGame

		return m, m.gamePlay.Init()

	case views.RetryGenerationMsg:
		// Retry generation - trigger GameSetupDoneMsg again
		return m, func() tea.Msg {
			return views.GameSetupDoneMsg{
				Config: m.gameConfig,
			}
		}
	}

	// Delegate to current state
	return m.updateCurrentState(msg)
}

func (m Model) passWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case StateAPISetup:
		apiModel, c := m.apiSetup.Update(msg)
		m.apiSetup = apiModel.(views.APISetupInterface)
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
	case StateMaxTokensSettings:
		maxTokensModel, c := m.maxtokensSettings.Update(msg)
		m.maxtokensSettings = maxTokensModel.(views.MaxTokensSettingsModel)
		if m.maxtokensSettings.IsDone() {
			m.state = StateSettings
		}
		cmd = c
	case StateAudioSettings:
		audioModel, c := m.audioSettings.Update(msg)
		m.audioSettings = audioModel.(views.AudioSettingsModel)
		cmd = c
	case StateGameSetup:
		setupModel, c := m.gameSetup.Update(msg)
		m.gameSetup = setupModel.(views.GameSetupModel)
		cmd = c
	case StateStoryLoading:
		loadingModel, c := m.storyLoading.Update(msg)
		m.storyLoading = loadingModel.(views.StoryLoadingModel)
		cmd = c
	case StateParallelLoading:
		loadingModel, c := m.parallelLoading.Update(msg)
		m.parallelLoading = loadingModel.(views.ParallelLoadingModel)
		cmd = c
	case StateGame:
		gameModel, c := m.gamePlay.Update(msg)
		m.gamePlay = gameModel.(views.GamePlayModel)
		cmd = c
	}
	return m, cmd
}

func (m Model) updateCurrentState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case StateAPISetup:
		apiModel, c := m.apiSetup.Update(msg)
		m.apiSetup = apiModel.(views.APISetupInterface)
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

	case StateMaxTokensSettings:
		maxTokensModel, c := m.maxtokensSettings.Update(msg)
		m.maxtokensSettings = maxTokensModel.(views.MaxTokensSettingsModel)
		cmd = c

	case StateAudioSettings:
		audioModel, c := m.audioSettings.Update(msg)
		m.audioSettings = audioModel.(views.AudioSettingsModel)
		cmd = c

	case StateGameSetup:
		setupModel, c := m.gameSetup.Update(msg)
		m.gameSetup = setupModel.(views.GameSetupModel)
		cmd = c

	case StateStoryLoading:
		loadingModel, c := m.storyLoading.Update(msg)
		m.storyLoading = loadingModel.(views.StoryLoadingModel)
		cmd = c

		// Check if loading is done
		if m.storyLoading.IsDone() {
			if m.storyLoading.IsCancelled() {
				// User cancelled, return to main menu
				m.state = StateMainMenu
				m.mainMenu = views.NewMainMenuModel(m.version, m.hasSaveFiles)
			} else {
				// Loading complete, transition to game
				m.state = StateGame
			}
		}

	case StateParallelLoading:
		loadingModel, c := m.parallelLoading.Update(msg)
		m.parallelLoading = loadingModel.(views.ParallelLoadingModel)
		cmd = c

	case StateGame:
		gameModel, c := m.gamePlay.Update(msg)
		m.gamePlay = gameModel.(views.GamePlayModel)
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
		m.apiSetup = views.NewTrinitySetupModel(m.config)
		return m, m.apiSetup.Init()

	case views.SettingsActionMaxTokens:
		m.prevState = StateSettings
		m.state = StateMaxTokensSettings
		m.maxtokensSettings = views.NewMaxTokensSettingsModel(m.config)
		return m, m.maxtokensSettings.Init()

	case views.SettingsActionAudio:
		m.prevState = StateSettings
		m.state = StateAudioSettings
		m.audioSettings = views.NewAudioSettingsModel(m.config, m.audioManager)
		return m, m.audioSettings.Init()

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

	case StateMaxTokensSettings:
		return m.maxtokensSettings.View()

	case StateAudioSettings:
		return m.audioSettings.View()

	case StateGameSetup:
		return m.gameSetup.View()

	case StateStoryLoading:
		return m.storyLoading.View()

	case StateParallelLoading:
		return m.parallelLoading.View()

	case StateGame:
		return m.gamePlay.View()

	default:
		return "載入中..."
	}
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
func checkForUpdates(manager *update.Manager, checkIntervalHours int) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			// 無 update manager - 靜默跳過檢查
			return updateCheckResultMsg{result: nil, err: nil}
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

// fallbackSerialGeneration falls back to serial generation if parallel fails
func (m Model) fallbackSerialGeneration() (tea.Model, tea.Cmd) {
	// Initialize story loading screen
	m.storyLoading = views.NewStoryLoadingModel()
	m.state = StateStoryLoading

	// Initialize game components in background
	stats := game.NewPlayerStats()

	// Create provider config from config (use configured provider for story generation)
	providerID := m.config.API.Provider.ProviderID
	encryptedKey := m.config.API.APIKeys[providerID]
	apiKey, err := m.config.DecryptAPIKey(providerID)

	// Debug logging
	debug.LogAPIKeyInfo(providerID, encryptedKey, apiKey, err)

	providerCfg := api.ProviderConfig{
		ProviderID: providerID,
		APIKey:     apiKey,
		// BaseURL is hardcoded in provider.go, don't override from config
		Model:     m.config.API.Provider.Model,
		MaxTokens: m.config.API.Provider.MaxTokens,
	}

	debug.LogProviderConfig(providerID, "", m.config.API.Provider.Model, m.config.API.Provider.MaxTokens, config.MaskAPIKey(apiKey))
	provider, _ := api.NewProvider(providerCfg)

	engineConfig := engine.DefaultEngineConfig()
	engineConfig.Provider = provider
	engineConfig.GameConfig = m.gameConfig
	storyEngine := engine.NewStoryEngine(engineConfig)

	// Create game play model without pregenerated content (will generate dynamically)
	m.gamePlay = views.NewGamePlayModel(storyEngine, stats, m.gameConfig, nil, m.audioManager, m.config)

	// Start loading screen and story generation
	return m, tea.Batch(
		m.storyLoading.Init(),
		m.gamePlay.Init(),
	)
}
