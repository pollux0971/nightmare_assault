// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
	"github.com/nightmare-assault/nightmare-assault/internal/errors"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
)

// FieldType represents different input fields
type FieldType int

const (
	FieldAPIKey FieldType = iota
	FieldModel
	FieldBaseURL
	FieldCount // Total number of fields
)

// ProviderFormConfig holds the configuration for one provider
type ProviderFormConfig struct {
	Info      api.ProviderInfo
	APIKey    textinput.Model
	Model     textinput.Model
	BaseURL   textinput.Model
	Tested    bool
	TestError error
}

// APISetupModelV2 is the new TUI model with side-by-side provider selection
type APISetupModelV2 struct {
	providers             []*ProviderFormConfig // 3 providers: Anthropic, OpenAI, Gemini
	selectedProviderIndex int                   // 0-2, controlled by left/right arrows
	selectedFieldIndex    int                   // 0-2, controlled by up/down arrows

	spinner spinner.Model
	testing bool
	config  *config.Config
	width   int
	height  int
	done    bool
	errorMsg string
}

// Card styles
var (
	// Selected provider card
	selectedCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9D4EDD")).
		Padding(1, 2).
		Width(30)

	// Unselected provider card
	unselectedCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1, 2).
		Width(30)

	// Provider name style (selected)
	providerNameSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D4EDD")).
		Bold(true)

	// Provider name style (unselected)
	providerNameUnselectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Bold(true)

	// Field label style
	fieldLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B2CBF"))

	// Status styles
	testedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	notTestedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	// Help text style
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true)
)

// NewAPISetupModelV2 creates a new API setup model with the new UI
func NewAPISetupModelV2(cfg *config.Config) APISetupModelV2 {
	providers := []*ProviderFormConfig{}

	// Create configs for the three main providers
	providerIDs := []string{"anthropic", "openai", "gemini"}
	allProviders := api.BuiltinProviders()

	for _, id := range providerIDs {
		var info api.ProviderInfo
		for _, p := range allProviders {
			if p.ID == id {
				info = p
				break
			}
		}

		// Create text inputs for this provider
		apiKeyInput := textinput.New()
		apiKeyInput.Placeholder = "輸入 API Key"
		apiKeyInput.EchoMode = textinput.EchoPassword
		apiKeyInput.EchoCharacter = '•'
		apiKeyInput.CharLimit = 256
		apiKeyInput.Width = 26

		modelInput := textinput.New()
		modelInput.Placeholder = "模型名稱"
		modelInput.CharLimit = 128
		modelInput.Width = 26

		baseURLInput := textinput.New()
		baseURLInput.Placeholder = "Base URL (可選)"
		baseURLInput.CharLimit = 256
		baseURLInput.Width = 26

		// Pre-fill with config if available
		if cfg.API.Provider.ProviderID == id {
			apiKey, _ := cfg.DecryptAPIKey(id)
			apiKeyInput.SetValue(apiKey)
			modelInput.SetValue(cfg.API.Provider.Model)
			// Base URL would be set here if we had it in config
		}

		providers = append(providers, &ProviderFormConfig{
			Info:     info,
			APIKey:   apiKeyInput,
			Model:    modelInput,
			BaseURL:  baseURLInput,
			Tested:   false,
		})
	}

	// Focus the first field of the first provider
	providers[0].APIKey.Focus()

	// Create spinner for testing
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))

	return APISetupModelV2{
		providers:             providers,
		selectedProviderIndex: 0,
		selectedFieldIndex:    0,
		spinner:               s,
		testing:               false,
		config:                cfg,
	}
}

// Init initializes the model
func (m APISetupModelV2) Init() tea.Cmd {
	return nil
}

// TestConnectionMsgV2 is sent when connection test completes
type TestConnectionMsgV2 struct {
	ProviderIndex int
	Err           error
}

// Update handles messages
func (m APISetupModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left", "h":
			return m.handleLeft()

		case "right", "l":
			return m.handleRight()

		case "up", "k":
			return m.handleUp()

		case "down", "j":
			return m.handleDown()

		case "tab":
			return m.handleTest()

		case "enter":
			return m.handleEnter()

		case "esc":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		if m.testing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case TestConnectionMsgV2:
		m.testing = false
		provider := m.providers[msg.ProviderIndex]
		provider.TestError = msg.Err
		provider.Tested = true
		if msg.Err != nil {
			m.errorMsg = m.getFriendlyErrorV2(provider.Info.Name, msg.Err)
		} else {
			m.errorMsg = ""
		}
		return m, nil
	}

	// Update the focused input
	return m.updateFocusedInput(msg)
}

func (m APISetupModelV2) handleLeft() (tea.Model, tea.Cmd) {
	// Blur current field
	m.blurCurrentField()

	// Move to previous provider (wrap around)
	m.selectedProviderIndex--
	if m.selectedProviderIndex < 0 {
		m.selectedProviderIndex = len(m.providers) - 1
	}

	// Focus the same field in the new provider
	m.focusCurrentField()
	return m, textinput.Blink
}

func (m APISetupModelV2) handleRight() (tea.Model, tea.Cmd) {
	// Blur current field
	m.blurCurrentField()

	// Move to next provider (wrap around)
	m.selectedProviderIndex++
	if m.selectedProviderIndex >= len(m.providers) {
		m.selectedProviderIndex = 0
	}

	// Focus the same field in the new provider
	m.focusCurrentField()
	return m, textinput.Blink
}

func (m APISetupModelV2) handleUp() (tea.Model, tea.Cmd) {
	// Blur current field
	m.blurCurrentField()

	// Move to previous field (wrap around)
	m.selectedFieldIndex--
	if m.selectedFieldIndex < 0 {
		m.selectedFieldIndex = int(FieldCount) - 1
	}

	// Focus new field
	m.focusCurrentField()
	return m, textinput.Blink
}

func (m APISetupModelV2) handleDown() (tea.Model, tea.Cmd) {
	// Blur current field
	m.blurCurrentField()

	// Move to next field (wrap around)
	m.selectedFieldIndex++
	if m.selectedFieldIndex >= int(FieldCount) {
		m.selectedFieldIndex = 0
	}

	// Focus new field
	m.focusCurrentField()
	return m, textinput.Blink
}

func (m APISetupModelV2) handleTest() (tea.Model, tea.Cmd) {
	provider := m.providers[m.selectedProviderIndex]
	apiKey := strings.TrimSpace(provider.APIKey.Value())

	if apiKey == "" {
		m.errorMsg = "請先輸入 API Key"
		return m, nil
	}

	m.testing = true
	m.errorMsg = ""

	return m, tea.Batch(
		m.spinner.Tick,
		m.testConnection(m.selectedProviderIndex, apiKey),
	)
}

func (m APISetupModelV2) handleEnter() (tea.Model, tea.Cmd) {
	// Find a provider that has been successfully tested
	var selectedProvider *ProviderFormConfig
	for _, p := range m.providers {
		if p.Tested && p.TestError == nil && strings.TrimSpace(p.APIKey.Value()) != "" {
			selectedProvider = p
			break
		}
	}

	if selectedProvider == nil {
		m.errorMsg = "請先測試至少一個提供商的連線"
		return m, nil
	}

	// Save the configuration
	if err := m.saveConfig(selectedProvider); err != nil {
		m.errorMsg = fmt.Sprintf("儲存配置失敗: %v", err)
		return m, nil
	}

	m.done = true
	return m, tea.Quit
}

func (m *APISetupModelV2) blurCurrentField() {
	provider := m.providers[m.selectedProviderIndex]
	switch FieldType(m.selectedFieldIndex) {
	case FieldAPIKey:
		provider.APIKey.Blur()
	case FieldModel:
		provider.Model.Blur()
	case FieldBaseURL:
		provider.BaseURL.Blur()
	}
}

func (m *APISetupModelV2) focusCurrentField() {
	provider := m.providers[m.selectedProviderIndex]
	switch FieldType(m.selectedFieldIndex) {
	case FieldAPIKey:
		provider.APIKey.Focus()
	case FieldModel:
		provider.Model.Focus()
	case FieldBaseURL:
		provider.BaseURL.Focus()
	}
}

func (m APISetupModelV2) updateFocusedInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	provider := m.providers[m.selectedProviderIndex]
	var cmd tea.Cmd

	switch FieldType(m.selectedFieldIndex) {
	case FieldAPIKey:
		provider.APIKey, cmd = provider.APIKey.Update(msg)
	case FieldModel:
		provider.Model, cmd = provider.Model.Update(msg)
	case FieldBaseURL:
		provider.BaseURL, cmd = provider.BaseURL.Update(msg)
	}

	return m, cmd
}

func (m APISetupModelV2) testConnection(providerIndex int, apiKey string) tea.Cmd {
	return func() tea.Msg {
		provider := m.providers[providerIndex]

		p, err := api.NewProvider(api.ProviderConfig{
			ProviderID: provider.Info.ID,
			APIKey:     apiKey,
		})
		if err != nil {
			return TestConnectionMsgV2{ProviderIndex: providerIndex, Err: err}
		}

		ctx := context.Background()
		err = p.TestConnection(ctx)
		return TestConnectionMsgV2{ProviderIndex: providerIndex, Err: err}
	}
}

func (m *APISetupModelV2) saveConfig(provider *ProviderFormConfig) error {
	apiKey := strings.TrimSpace(provider.APIKey.Value())
	modelName := strings.TrimSpace(provider.Model.Value())

	// Use default model if not specified
	if modelName == "" {
		modelName = api.GetDefaultModel(provider.Info.ID)
	}

	// Encrypt and save API key
	if err := m.config.EncryptAPIKey(provider.Info.ID, apiKey); err != nil {
		return err
	}

	// Set as active provider
	m.config.API.Provider.ProviderID = provider.Info.ID
	m.config.API.Provider.Model = modelName

	// Keep existing MaxTokens or set default if zero
	if m.config.API.Provider.MaxTokens == 0 {
		m.config.API.Provider.MaxTokens = 100000
	}

	return m.config.Save()
}

func (m APISetupModelV2) getFriendlyErrorV2(providerName string, err error) string {
	translator := i18n.GetGlobal()
	if translator != nil {
		if errors.IsFriendlyError(err) {
			return "❌ " + errors.FormatUserError(err, translator)
		}

		errStr := err.Error()
		var friendlyErr error
		switch {
		case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "401") || strings.Contains(errStr, "403"):
			friendlyErr = errors.NewAPIErrorFriendly(providerName, 401, "connection test", err)
		case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
			friendlyErr = errors.NewNetworkErrorFriendly("connection test", err)
		case strings.Contains(errStr, "429"):
			friendlyErr = errors.NewAPIErrorFriendly(providerName, 429, "connection test", err)
		case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline"):
			friendlyErr = errors.NewAPIErrorFriendly(providerName, 0, "connection test", err)
		default:
			friendlyErr = errors.NewAPIErrorFriendly(providerName, 0, "connection test", err)
		}
		return "❌ " + errors.FormatUserError(friendlyErr, translator)
	}

	// Fallback
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "401") || strings.Contains(errStr, "403"):
		return "❌ API Key 無效"
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return "❌ 網路連線失敗"
	case strings.Contains(errStr, "429"):
		return "❌ 請求過於頻繁"
	default:
		return fmt.Sprintf("❌ 連線失敗: %s", errStr)
	}
}

// View renders the model
func (m APISetupModelV2) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("API 設置")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Help text
	help := helpStyle.Render("使用 ←→ 切換提供商 | ↑↓ 切換欄位 | Tab 測試連線 | Enter 確認 | Esc 退出")
	b.WriteString(help)
	b.WriteString("\n\n")

	// Render provider cards side by side
	cards := []string{}
	for i, provider := range m.providers {
		cards = append(cards, m.renderProviderCard(i, provider))
	}

	cardsRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	b.WriteString(cardsRow)
	b.WriteString("\n\n")

	// Testing indicator
	if m.testing {
		b.WriteString(m.spinner.View())
		b.WriteString(" 測試連線中...")
		b.WriteString("\n\n")
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render(m.errorMsg))
		b.WriteString("\n")
	}

	return b.String()
}

func (m APISetupModelV2) renderProviderCard(index int, provider *ProviderFormConfig) string {
	var b strings.Builder

	isSelected := index == m.selectedProviderIndex

	// Provider name
	var providerName string
	if isSelected {
		providerName = providerNameSelectedStyle.Render(provider.Info.Name)
	} else {
		providerName = providerNameUnselectedStyle.Render(provider.Info.Name)
	}
	b.WriteString(providerName)
	b.WriteString("\n\n")

	// API Key field
	b.WriteString(m.renderField(index, FieldAPIKey, "API Key:", provider.APIKey.View()))
	b.WriteString("\n")

	// Model field
	b.WriteString(m.renderField(index, FieldModel, "Model:", provider.Model.View()))
	b.WriteString("\n")

	// Base URL field (optional)
	b.WriteString(m.renderField(index, FieldBaseURL, "Base URL:", provider.BaseURL.View()))
	b.WriteString("\n\n")

	// Test status
	if provider.Tested {
		if provider.TestError == nil {
			b.WriteString(testedStyle.Render("✓ 已測試"))
		} else {
			b.WriteString(errorStyle.Render("✗ 測試失敗"))
		}
	} else {
		b.WriteString(notTestedStyle.Render("未測試"))
	}

	// Wrap in card style
	cardContent := b.String()
	if isSelected {
		return selectedCardStyle.Render(cardContent)
	}
	return unselectedCardStyle.Render(cardContent)
}

func (m APISetupModelV2) renderField(providerIndex int, field FieldType, label string, value string) string {
	isSelected := providerIndex == m.selectedProviderIndex && m.selectedFieldIndex == int(field)

	var fieldStr string
	if isSelected {
		fieldStr = fieldLabelStyle.Render(label) + " " + value
	} else {
		fieldStr = infoStyle.Render(label) + " " + value
	}

	return fieldStr
}

// IsDone returns true if setup is complete
func (m APISetupModelV2) IsDone() bool {
	return m.done
}

// GetConfig returns the updated config
func (m APISetupModelV2) GetConfig() *config.Config {
	return m.config
}
