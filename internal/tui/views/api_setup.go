// Package views provides TUI view components for Nightmare Assault.
package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// APISetupState represents the current state of the API setup wizard.
type APISetupState int

const (
	StateSelectProvider APISetupState = iota
	StateEnterAPIKey
	StateTesting
	StateSuccess
	StateError
)

// ProviderItem represents a provider in the selection list.
type ProviderItem struct {
	info api.ProviderInfo
}

func (i ProviderItem) Title() string       { return i.info.Name }
func (i ProviderItem) Description() string { return i.info.Description }
func (i ProviderItem) FilterValue() string { return i.info.Name }

// APISetupModel is the TUI model for API configuration.
type APISetupModel struct {
	state          APISetupState
	providerList   list.Model
	textInput      textinput.Model
	spinner        spinner.Model
	selectedProvider *api.ProviderInfo
	apiKey         string
	errorMsg       string
	config         *config.Config
	width          int
	height         int
	done           bool
	testResult     error
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D4EDD")).
		Bold(true).
		MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B2CBF")).
		MarginBottom(1)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
)

// NewAPISetupModel creates a new API setup model.
func NewAPISetupModel(cfg *config.Config) APISetupModel {
	// Create provider list items
	providers := api.BuiltinProviders()
	items := make([]list.Item, 0)

	// Add official and gateway providers (most commonly used)
	for _, p := range providers {
		if p.Category == "official" || p.Category == "gateway" {
			items = append(items, ProviderItem{info: p})
		}
	}

	// Create list model
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#9D4EDD")).
		BorderForeground(lipgloss.Color("#9D4EDD"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#7B2CBF"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "選擇 API 供應商"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	// Create text input for API key
	ti := textinput.New()
	ti.Placeholder = "輸入您的 API Key"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	ti.CharLimit = 256
	ti.Width = 50

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D4EDD"))

	return APISetupModel{
		state:        StateSelectProvider,
		providerList: l,
		textInput:    ti,
		spinner:      s,
		config:       cfg,
	}
}

// Init initializes the model.
func (m APISetupModel) Init() tea.Cmd {
	return nil
}

// TestConnectionMsg is sent when connection test completes.
type TestConnectionMsg struct {
	Err error
}

// Update handles messages.
func (m APISetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.providerList.SetSize(msg.Width-4, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == StateSelectProvider {
				return m, tea.Quit
			}
		case "esc":
			if m.state == StateEnterAPIKey {
				m.state = StateSelectProvider
				m.textInput.Reset()
				return m, nil
			}
		case "enter":
			return m.handleEnter()
		}

	case spinner.TickMsg:
		if m.state == StateTesting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case TestConnectionMsg:
		m.testResult = msg.Err
		if msg.Err == nil {
			m.state = StateSuccess
			// Save config
			if err := m.saveConfig(); err != nil {
				m.state = StateError
				m.errorMsg = fmt.Sprintf("儲存配置失敗: %v", err)
			}
		} else {
			m.state = StateError
			m.errorMsg = m.getFriendlyError(msg.Err)
		}
		return m, nil
	}

	// Delegate to sub-components based on state
	var cmd tea.Cmd
	switch m.state {
	case StateSelectProvider:
		m.providerList, cmd = m.providerList.Update(msg)
	case StateEnterAPIKey:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m APISetupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateSelectProvider:
		if item, ok := m.providerList.SelectedItem().(ProviderItem); ok {
			m.selectedProvider = &item.info
			m.state = StateEnterAPIKey
			m.textInput.Focus()
			return m, textinput.Blink
		}

	case StateEnterAPIKey:
		m.apiKey = m.textInput.Value()
		if m.apiKey == "" {
			m.errorMsg = "請輸入 API Key"
			return m, nil
		}
		m.state = StateTesting
		return m, tea.Batch(
			m.spinner.Tick,
			m.testConnection(),
		)

	case StateSuccess:
		m.done = true
		return m, tea.Quit

	case StateError:
		// Go back to API key input
		m.state = StateEnterAPIKey
		m.errorMsg = ""
		return m, nil
	}

	return m, nil
}

func (m APISetupModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		provider, err := api.NewProvider(api.ProviderConfig{
			ProviderID: m.selectedProvider.ID,
			APIKey:     m.apiKey,
		})
		if err != nil {
			return TestConnectionMsg{Err: err}
		}

		ctx := context.Background()
		err = provider.TestConnection(ctx)
		return TestConnectionMsg{Err: err}
	}
}

func (m *APISetupModel) saveConfig() error {
	// Encrypt and save API key
	if err := m.config.EncryptAPIKey(m.selectedProvider.ID, m.apiKey); err != nil {
		return err
	}

	// Set as active provider for smart model
	m.config.API.Smart.ProviderID = m.selectedProvider.ID
	m.config.API.Smart.Model = "" // Let provider use default

	return m.config.Save()
}

func (m APISetupModel) getFriendlyError(err error) string {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "401") || strings.Contains(errStr, "403"):
		return "❌ API Key 無效，請檢查格式是否正確"
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return "❌ 網路連線失敗，請檢查網路設定"
	case strings.Contains(errStr, "429"):
		return "❌ 請求過於頻繁，請稍後再試"
	default:
		return fmt.Sprintf("❌ 連線失敗: %s", errStr)
	}
}

// View renders the model.
func (m APISetupModel) View() string {
	var b strings.Builder

	switch m.state {
	case StateSelectProvider:
		b.WriteString(m.providerList.View())

	case StateEnterAPIKey:
		b.WriteString(titleStyle.Render(fmt.Sprintf("設定 %s API Key", m.selectedProvider.Name)))
		b.WriteString("\n\n")
		b.WriteString(subtitleStyle.Render("請輸入您的 API Key（將加密儲存於本地）"))
		b.WriteString("\n\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		if m.errorMsg != "" {
			b.WriteString(errorStyle.Render(m.errorMsg))
			b.WriteString("\n")
		}
		b.WriteString(infoStyle.Render("按 Enter 測試連線，按 Esc 返回"))

	case StateTesting:
		b.WriteString(titleStyle.Render(fmt.Sprintf("測試 %s 連線", m.selectedProvider.Name)))
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" 連線測試中...")

	case StateSuccess:
		b.WriteString(titleStyle.Render("設定完成"))
		b.WriteString("\n\n")
		b.WriteString(successStyle.Render("✓ 連線成功！"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("供應商: %s\n", m.selectedProvider.Name))
		b.WriteString(fmt.Sprintf("API Key: %s\n", config.MaskAPIKey(m.apiKey)))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("按 Enter 繼續"))

	case StateError:
		b.WriteString(titleStyle.Render("連線測試失敗"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render("按 Enter 重試"))
	}

	return b.String()
}

// IsDone returns true if setup is complete.
func (m APISetupModel) IsDone() bool {
	return m.done
}

// GetConfig returns the updated config.
func (m APISetupModel) GetConfig() *config.Config {
	return m.config
}
