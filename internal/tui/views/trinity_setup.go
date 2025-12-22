package views

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
	"github.com/nightmare-assault/nightmare-assault/internal/trinity"
)

// ParamType represents the type of parameter being configured
type ParamType int

const (
	ParamProvider ParamType = iota
	ParamModel
	ParamMaxTokens
	ParamTemperature
	ParamBaseURL
	ParamCount // Number of parameter types
)

// String returns the display name for a parameter type
func (p ParamType) String() string {
	switch p {
	case ParamProvider:
		return "Provider"
	case ParamModel:
		return "Model"
	case ParamMaxTokens:
		return "MaxTokens"
	case ParamTemperature:
		return "Temperature"
	case ParamBaseURL:
		return "BaseURL"
	default:
		return "Unknown"
	}
}

// TierConfig holds the configuration for a single tier
type TierConfig struct {
	name          string            // "Thinking", "Reactive", "Rapid"
	level         trinity.TierLevel // Trinity tier level
	providerIndex int               // Index in preset list, -1 = custom mode
	customMode    bool              // True if in custom mode

	// Textinput fields for each parameter
	providerID  textinput.Model
	model       textinput.Model
	maxTokens   textinput.Model
	temperature textinput.Model
	baseURL     textinput.Model

	// Testing state
	tested      bool
	testSuccess bool
	testError   error
}

// TrinitySetupModel is the Bubble Tea model for Trinity API configuration
type TrinitySetupModel struct {
	// Tier configurations (0=Thinking, 1=Reactive, 2=Rapid)
	tiers [3]*TierConfig

	// Navigation state
	focusedTier  int       // 0-2: which tier is focused
	focusedParam ParamType // 0-4: which parameter is focused
	editMode     bool      // True when editing a parameter

	// Testing state
	testing     bool // True when any tier is being tested
	testingTier int  // Index of tier being tested, -1 = all

	// UI state
	config       *config.Config
	width        int
	height       int
	done         bool
	errorMsg     string
	spinner      spinner.Model
	quitting     bool
	saveAttempts int // Track save attempts for validation
}

// TrinitySetupMsg is sent when the Trinity setup is complete
type TrinitySetupMsg struct {
	Config *config.Config
}

// TestTierMsg is sent when a tier connection test completes
type TestTierMsg struct {
	TierIndex int
	Success   bool
	Error     error
}

// NewTrinitySetupModel creates a new Trinity setup model
func NewTrinitySetupModel(cfg *config.Config) *TrinitySetupModel {
	m := &TrinitySetupModel{
		config:       cfg,
		focusedTier:  0,
		focusedParam: ParamProvider,
		editMode:     false,
		testing:      false,
		testingTier:  -1,
		spinner:      spinner.New(),
	}

	// Initialize spinner
	m.spinner.Spinner = spinner.Dot

	// Initialize three tiers
	m.tiers[0] = m.initTier("Thinking", trinity.TierThinking)
	m.tiers[1] = m.initTier("Reactive", trinity.TierReactive)
	m.tiers[2] = m.initTier("Rapid", trinity.TierRapid)

	// Load existing config if available
	m.loadFromConfig()

	return m
}

// initTier initializes a tier configuration with default values
func (m *TrinitySetupModel) initTier(name string, level trinity.TierLevel) *TierConfig {
	tier := &TierConfig{
		name:          name,
		level:         level,
		providerIndex: 0, // Start with first preset
		customMode:    false,
	}

	// Create textinput models
	tier.providerID = textinput.New()
	tier.providerID.Placeholder = "Provider ID"
	tier.providerID.CharLimit = 50

	tier.model = textinput.New()
	tier.model.Placeholder = "Model name"
	tier.model.CharLimit = 100

	tier.maxTokens = textinput.New()
	tier.maxTokens.Placeholder = "1000-200000"
	tier.maxTokens.CharLimit = 6

	tier.temperature = textinput.New()
	tier.temperature.Placeholder = "0.0-2.0"
	tier.temperature.CharLimit = 4

	tier.baseURL = textinput.New()
	tier.baseURL.Placeholder = "https://api.example.com/v1"
	tier.baseURL.CharLimit = 200

	// Apply default preset
	m.applyPreset(tier, 0)

	return tier
}

// loadFromConfig loads configuration from existing config.Trinity
func (m *TrinitySetupModel) loadFromConfig() {
	if !m.config.Trinity.Enabled {
		return
	}

	// Load Thinking tier
	if m.config.Trinity.Thinking.ProviderID != "" {
		m.loadTierSettings(m.tiers[0], m.config.Trinity.Thinking)
	}

	// Load Reactive tier
	if m.config.Trinity.Reactive.ProviderID != "" {
		m.loadTierSettings(m.tiers[1], m.config.Trinity.Reactive)
	}

	// Load Rapid tier
	if m.config.Trinity.Rapid.ProviderID != "" {
		m.loadTierSettings(m.tiers[2], m.config.Trinity.Rapid)
	}
}

// loadTierSettings loads settings into a tier from ProviderSettings
func (m *TrinitySetupModel) loadTierSettings(tier *TierConfig, settings config.ProviderSettings) {
	tier.providerID.SetValue(settings.ProviderID)
	tier.model.SetValue(settings.Model)
	tier.maxTokens.SetValue(fmt.Sprintf("%d", settings.MaxTokens))
	tier.temperature.SetValue(fmt.Sprintf("%.1f", settings.Temperature))
	tier.baseURL.SetValue(settings.BaseURL)

	// Check if this matches a preset
	presetIdx := FindPresetIndex(tier.level, settings.ProviderID)
	if presetIdx >= 0 {
		preset := GetPresetByIndex(tier.level, presetIdx)
		// Check if model and other settings match the preset
		if settings.Model == preset.DefaultModel &&
			settings.MaxTokens == preset.DefaultTokens &&
			settings.BaseURL == preset.BaseURL {
			tier.providerIndex = presetIdx
			tier.customMode = false
		} else {
			// Same provider but custom settings
			tier.providerIndex = -1
			tier.customMode = true
		}
	} else {
		// Unknown provider - custom mode
		tier.providerIndex = -1
		tier.customMode = true
	}
}

// applyPreset applies a provider preset to a tier
func (m *TrinitySetupModel) applyPreset(tier *TierConfig, presetIdx int) {
	preset := GetPresetByIndex(tier.level, presetIdx)

	tier.providerID.SetValue(preset.ID)
	tier.model.SetValue(preset.DefaultModel)
	tier.maxTokens.SetValue(fmt.Sprintf("%d", preset.DefaultTokens))
	tier.temperature.SetValue(fmt.Sprintf("%.1f", preset.DefaultTemp))
	tier.baseURL.SetValue(preset.BaseURL)

	tier.providerIndex = presetIdx
	tier.customMode = (preset.ID == "custom")

	// Mark as not tested since config changed
	tier.tested = false
	tier.testSuccess = false
	tier.testError = nil
}

// Init initializes the Trinity setup model
func (m *TrinitySetupModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and updates the model
func (m *TrinitySetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case TestTierMsg:
		// Test completed for a tier
		m.testing = false
		if msg.TierIndex >= 0 && msg.TierIndex < 3 {
			tier := m.tiers[msg.TierIndex]
			tier.tested = true
			tier.testSuccess = msg.Success
			tier.testError = msg.Error

			if !msg.Success {
				m.errorMsg = fmt.Sprintf("%s Tier 測試失敗: %s",
					tier.name, msg.Error.Error())
			} else {
				m.errorMsg = ""
			}
		}

	case tea.KeyMsg:
		// Handle keyboard input
		return m.handleKeyPress(msg)
	}

	// Update textinputs if in edit mode
	if m.editMode {
		tier := m.tiers[m.focusedTier]
		var cmd tea.Cmd
		switch m.focusedParam {
		case ParamProvider:
			tier.providerID, cmd = tier.providerID.Update(msg)
		case ParamModel:
			tier.model, cmd = tier.model.Update(msg)
		case ParamMaxTokens:
			tier.maxTokens, cmd = tier.maxTokens.Update(msg)
		case ParamTemperature:
			tier.temperature, cmd = tier.temperature.Update(msg)
		case ParamBaseURL:
			tier.baseURL, cmd = tier.baseURL.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input based on current mode
func (m *TrinitySetupModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editMode {
		return m.handleEditMode(msg)
	}
	return m.handleNavigationMode(msg)
}

// handleNavigationMode handles keys in navigation mode
func (m *TrinitySetupModel) handleNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.errorMsg != "" {
			m.errorMsg = ""
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "up":
		m.focusedTier--
		if m.focusedTier < 0 {
			m.focusedTier = 2
		}
		m.errorMsg = ""

	case "down":
		m.focusedTier++
		if m.focusedTier > 2 {
			m.focusedTier = 0
		}
		m.errorMsg = ""

	case "left":
		m.focusedParam--
		if m.focusedParam < 0 {
			m.focusedParam = ParamCount - 1
		}
		m.errorMsg = ""

	case "right":
		m.focusedParam++
		if m.focusedParam >= ParamCount {
			m.focusedParam = 0
		}
		m.errorMsg = ""

	case "shift+left":
		return m, m.handleProviderSwitch(-1)

	case "shift+right":
		return m, m.handleProviderSwitch(1)

	case "e":
		// Enter edit mode for current parameter
		m.editMode = true
		m.focusCurrentInput()
		m.errorMsg = ""

	case "tab":
		// Test current tier
		if !m.testing {
			m.testing = true
			m.testingTier = m.focusedTier
			return m, m.testTierConnection(m.focusedTier)
		}

	case "shift+tab":
		// Test all tiers
		if !m.testing {
			m.testing = true
			m.testingTier = -1
			return m, m.testAllTiers()
		}

	case "enter":
		// Save and exit
		return m, m.handleSave()
	}

	return m, nil
}

// handleEditMode handles keys in edit mode
func (m *TrinitySetupModel) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		// Confirm edit and return to navigation mode
		m.editMode = false
		tier := m.tiers[m.focusedTier]
		m.blurCurrentInput()

		// Mark as not tested since config changed
		tier.tested = false
		tier.testSuccess = false
		tier.testError = nil

		// If edited provider/model/baseURL, check if it's custom
		if m.focusedParam == ParamProvider || m.focusedParam == ParamModel || m.focusedParam == ParamBaseURL {
			presetIdx := FindPresetIndex(tier.level, tier.providerID.Value())
			if presetIdx >= 0 {
				preset := GetPresetByIndex(tier.level, presetIdx)
				if tier.model.Value() != preset.DefaultModel || tier.baseURL.Value() != preset.BaseURL {
					tier.customMode = true
					tier.providerIndex = -1
				}
			} else {
				tier.customMode = true
				tier.providerIndex = -1
			}
		}

		m.errorMsg = ""

	case "esc":
		// Cancel edit and return to navigation mode
		m.editMode = false
		m.blurCurrentInput()
		m.errorMsg = ""

	default:
		// Pass key to textinput
		tier := m.tiers[m.focusedTier]
		var cmd tea.Cmd
		switch m.focusedParam {
		case ParamProvider:
			tier.providerID, cmd = tier.providerID.Update(msg)
		case ParamModel:
			tier.model, cmd = tier.model.Update(msg)
		case ParamMaxTokens:
			tier.maxTokens, cmd = tier.maxTokens.Update(msg)
		case ParamTemperature:
			tier.temperature, cmd = tier.temperature.Update(msg)
		case ParamBaseURL:
			tier.baseURL, cmd = tier.baseURL.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

// handleProviderSwitch switches to the next/previous provider preset
func (m *TrinitySetupModel) handleProviderSwitch(direction int) tea.Cmd {
	tier := m.tiers[m.focusedTier]
	presetCount := GetPresetCount(tier.level)

	tier.providerIndex += direction
	if tier.providerIndex < 0 {
		tier.providerIndex = presetCount - 1
	} else if tier.providerIndex >= presetCount {
		tier.providerIndex = 0
	}

	m.applyPreset(tier, tier.providerIndex)
	m.errorMsg = ""

	return nil
}

// focusCurrentInput focuses the current parameter's textinput
func (m *TrinitySetupModel) focusCurrentInput() {
	tier := m.tiers[m.focusedTier]
	switch m.focusedParam {
	case ParamProvider:
		tier.providerID.Focus()
	case ParamModel:
		tier.model.Focus()
	case ParamMaxTokens:
		tier.maxTokens.Focus()
	case ParamTemperature:
		tier.temperature.Focus()
	case ParamBaseURL:
		tier.baseURL.Focus()
	}
}

// blurCurrentInput blurs the current parameter's textinput
func (m *TrinitySetupModel) blurCurrentInput() {
	tier := m.tiers[m.focusedTier]
	tier.providerID.Blur()
	tier.model.Blur()
	tier.maxTokens.Blur()
	tier.temperature.Blur()
	tier.baseURL.Blur()
}

// testTierConnection tests the connection for a single tier
func (m *TrinitySetupModel) testTierConnection(tierIndex int) tea.Cmd {
	return func() tea.Msg {
		tier := m.tiers[tierIndex]

		// Validate first
		if err := m.validateTier(tier); err != nil {
			return TestTierMsg{
				TierIndex: tierIndex,
				Success:   false,
				Error:     err,
			}
		}

		// Get API key for this provider
		apiKey := m.config.GetAPIKey(tier.providerID.Value())
		if apiKey == "" {
			return TestTierMsg{
				TierIndex: tierIndex,
				Success:   false,
				Error:     fmt.Errorf("未設定 API Key，請先到設定頁面配置"),
			}
		}

		// Create provider config
		maxTokens, _ := strconv.Atoi(tier.maxTokens.Value())
		temp, _ := strconv.ParseFloat(tier.temperature.Value(), 64)

		providerCfg := trinity.ProviderTierConfig{
			ProviderID:  tier.providerID.Value(),
			APIKey:      apiKey,
			Model:       tier.model.Value(),
			MaxTokens:   maxTokens,
			Temperature: temp,
		}

		// Create provider
		provider, err := providerCfg.CreateProvider()
		if err != nil {
			return TestTierMsg{
				TierIndex: tierIndex,
				Success:   false,
				Error:     err,
			}
		}

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Send a minimal test message
		testMsg := []client.Message{
			{
				Role:    "user",
				Content: "test",
			},
		}

		_, err = provider.SendMessage(ctx, testMsg)

		return TestTierMsg{
			TierIndex: tierIndex,
			Success:   err == nil,
			Error:     err,
		}
	}
}

// testAllTiers tests all three tiers concurrently
func (m *TrinitySetupModel) testAllTiers() tea.Cmd {
	return tea.Batch(
		m.testTierConnection(0),
		m.testTierConnection(1),
		m.testTierConnection(2),
	)
}

// validateTier validates a tier's configuration
func (m *TrinitySetupModel) validateTier(tier *TierConfig) error {
	// Provider ID cannot be empty
	if strings.TrimSpace(tier.providerID.Value()) == "" {
		return fmt.Errorf("Provider ID 不能為空")
	}

	// Model cannot be empty
	if strings.TrimSpace(tier.model.Value()) == "" {
		return fmt.Errorf("Model 不能為空")
	}

	// MaxTokens: 1000-200000
	maxTokens, err := strconv.Atoi(tier.maxTokens.Value())
	if err != nil || maxTokens < 1000 || maxTokens > 200000 {
		return fmt.Errorf("MaxTokens 範圍: 1000-200000")
	}

	// Temperature: 0.0-2.0
	temp, err := strconv.ParseFloat(tier.temperature.Value(), 64)
	if err != nil || temp < 0.0 || temp > 2.0 {
		return fmt.Errorf("Temperature 範圍: 0.0-2.0")
	}

	return nil
}

// handleSave validates and saves the configuration
func (m *TrinitySetupModel) handleSave() tea.Cmd {
	m.saveAttempts++

	// Validate all tiers
	for _, tier := range m.tiers {
		if err := m.validateTier(tier); err != nil {
			m.errorMsg = fmt.Sprintf("%s Tier 配置錯誤: %s", tier.name, err.Error())
			return nil
		}

		// Check if tested
		if !tier.tested || !tier.testSuccess {
			m.errorMsg = fmt.Sprintf("請先測試 %s Tier 的連線", tier.name)
			return nil
		}
	}

	// Save to config
	if err := m.saveToConfig(); err != nil {
		m.errorMsg = fmt.Sprintf("保存配置失敗: %s", err.Error())
		return nil
	}

	m.done = true
	logger.Debug("Trinity configuration saved successfully", nil)

	return func() tea.Msg {
		return TrinitySetupMsg{Config: m.config}
	}
}

// saveToConfig saves the current settings to config.Trinity
func (m *TrinitySetupModel) saveToConfig() error {
	// Helper to parse values
	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}
	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	// Update Trinity config
	m.config.Trinity = config.TrinityConfig{
		Enabled:         true,
		FallbackEnabled: true,

		Thinking: config.ProviderSettings{
			ProviderID:  m.tiers[0].providerID.Value(),
			Model:       m.tiers[0].model.Value(),
			MaxTokens:   parseInt(m.tiers[0].maxTokens.Value()),
			Temperature: parseFloat(m.tiers[0].temperature.Value()),
			BaseURL:     m.tiers[0].baseURL.Value(),
		},

		Reactive: config.ProviderSettings{
			ProviderID:  m.tiers[1].providerID.Value(),
			Model:       m.tiers[1].model.Value(),
			MaxTokens:   parseInt(m.tiers[1].maxTokens.Value()),
			Temperature: parseFloat(m.tiers[1].temperature.Value()),
			BaseURL:     m.tiers[1].baseURL.Value(),
		},

		Rapid: config.ProviderSettings{
			ProviderID:  m.tiers[2].providerID.Value(),
			Model:       m.tiers[2].model.Value(),
			MaxTokens:   parseInt(m.tiers[2].maxTokens.Value()),
			Temperature: parseFloat(m.tiers[2].temperature.Value()),
			BaseURL:     m.tiers[2].baseURL.Value(),
		},

		AgentTierOverrides: make(map[string]string),
	}

	// Save config to disk
	return m.config.Save()
}

// View renders the Trinity setup UI
func (m *TrinitySetupModel) View() string {
	if m.quitting {
		return ""
	}

	// Wait for window size to be set
	if m.width == 0 || m.height == 0 {
		return "載入中..."
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Help line
	b.WriteString(m.renderHelpLine())
	b.WriteString("\n\n")

	// Three tier cards
	for i := 0; i < 3; i++ {
		b.WriteString(m.renderTierCard(i))
		if i < 2 {
			b.WriteString("\n")
		}
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		b.WriteString(m.renderError())
	}

	// Testing indicator
	if m.testing {
		b.WriteString("\n\n")
		b.WriteString(m.renderTestingIndicator())
	}

	return b.String()
}

// renderHeader renders the header
func (m *TrinitySetupModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9D4EDD")).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#9D4EDD")).
		Padding(0, 2).
		Width(m.width - 4).
		Align(lipgloss.Center)

	return style.Render("Trinity API 配置")
}

// renderHelpLine renders the keyboard help line
func (m *TrinitySetupModel) renderHelpLine() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	if m.editMode {
		return style.Render("Enter 確認 │ Esc 取消")
	}

	return style.Render("↑↓ Tier │ ←→ 參數 │ Shift+←→ Provider │ e 編輯 │ Tab 測試 │ Enter 保存 │ Esc 退出")
}

// renderTierCard renders a single tier configuration card
func (m *TrinitySetupModel) renderTierCard(tierIndex int) string {
	tier := m.tiers[tierIndex]
	isFocused := (tierIndex == m.focusedTier)

	// Card border style
	var cardStyle lipgloss.Style
	if isFocused {
		cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#9D4EDD")).
			Padding(1, 2).
			Width(m.width - 4)
	} else {
		cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 2).
			Width(m.width - 4)
	}

	var content strings.Builder

	// Title line with status
	titleStyle := lipgloss.NewStyle().Bold(true)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	title := fmt.Sprintf("%s Tier (%s)", tier.name, m.getTierDescription(tier.level))
	status := m.getTierStatus(tier)

	// Calculate spacing, ensure non-negative
	spacing := m.width - len(title) - len(status) - 12
	if spacing < 0 {
		spacing = 0
	}
	titleLine := titleStyle.Render(title) + strings.Repeat(" ", spacing) + statusStyle.Render(status)
	content.WriteString(titleLine)
	content.WriteString("\n")

	// Calculate separator width, ensure non-negative
	separatorWidth := m.width - 8
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	content.WriteString(strings.Repeat("─", separatorWidth))
	content.WriteString("\n")

	// Parameters
	content.WriteString(m.renderParam(tier, ParamProvider, isFocused))
	content.WriteString("\n")
	content.WriteString(m.renderParam(tier, ParamModel, isFocused))
	content.WriteString("\n")
	content.WriteString(m.renderParam(tier, ParamMaxTokens, isFocused))
	content.WriteString("\n")
	content.WriteString(m.renderParam(tier, ParamTemperature, isFocused))
	content.WriteString("\n")
	content.WriteString(m.renderParam(tier, ParamBaseURL, isFocused))

	return cardStyle.Render(content.String())
}

// renderParam renders a single parameter line
func (m *TrinitySetupModel) renderParam(tier *TierConfig, param ParamType, tierFocused bool) string {
	isFocused := tierFocused && (m.focusedParam == param)

	// Label style
	labelStyle := lipgloss.NewStyle().Width(14)
	if isFocused {
		labelStyle = labelStyle.Foreground(lipgloss.Color("#00FF00")).Bold(true)
	}

	// Value style
	valueStyle := lipgloss.NewStyle()
	if m.editMode && isFocused {
		valueStyle = valueStyle.Foreground(lipgloss.Color("#FFFF00"))
	} else if isFocused {
		valueStyle = valueStyle.Foreground(lipgloss.Color("#00FF00"))
	}

	// Indicator
	indicator := "  "
	if isFocused && !m.editMode {
		indicator = "> "
	}

	// Get value
	var value string
	switch param {
	case ParamProvider:
		if !tier.customMode && tier.providerIndex >= 0 {
			preset := GetPresetByIndex(tier.level, tier.providerIndex)
			value = fmt.Sprintf("%s [Shift+→]", preset.DisplayName)
		} else {
			value = tier.providerID.View()
		}
	case ParamModel:
		value = tier.model.View()
	case ParamMaxTokens:
		value = tier.maxTokens.View()
	case ParamTemperature:
		value = tier.temperature.View()
	case ParamBaseURL:
		value = tier.baseURL.View()
	}

	return labelStyle.Render(param.String()+":") + indicator + valueStyle.Render(value)
}

// getTierDescription returns a description for a tier level
func (m *TrinitySetupModel) getTierDescription(level trinity.TierLevel) string {
	switch level {
	case trinity.TierThinking:
		return "高品質推理"
	case trinity.TierReactive:
		return "平衡互動"
	case trinity.TierRapid:
		return "快速回應"
	default:
		return ""
	}
}

// getTierStatus returns the status string for a tier
func (m *TrinitySetupModel) getTierStatus(tier *TierConfig) string {
	if m.testing && m.testingTier == -1 {
		return "[⊗ 測試中]"
	}
	if tier.tested {
		if tier.testSuccess {
			return "[✓ 已測試]"
		}
		return "[✗ 測試失敗]"
	}
	return "[○ 未測試]"
}

// renderError renders the error message
func (m *TrinitySetupModel) renderError() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	return style.Render("錯誤: " + m.errorMsg)
}

// renderTestingIndicator renders the testing indicator
func (m *TrinitySetupModel) renderTestingIndicator() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00"))

	msg := "正在測試連線..."
	if m.testingTier >= 0 {
		msg = fmt.Sprintf("正在測試 %s Tier...", m.tiers[m.testingTier].name)
	}

	return style.Render(m.spinner.View() + " " + msg)
}

// IsDone returns true if the Trinity setup is complete
// Implements APISetupInterface
func (m *TrinitySetupModel) IsDone() bool {
	return m.done
}

// GetConfig returns the updated configuration
// Implements APISetupInterface
func (m *TrinitySetupModel) GetConfig() *config.Config {
	return m.config
}
