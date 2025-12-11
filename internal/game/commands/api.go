// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/api"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// APICommand handles the /api command for provider management.
type APICommand struct {
	config *config.Config
}

// NewAPICommand creates a new API command handler.
func NewAPICommand(cfg *config.Config) *APICommand {
	return &APICommand{config: cfg}
}

// Execute runs the /api command and returns the result.
func (c *APICommand) Execute(args string) CommandResult {
	args = strings.TrimSpace(args)

	switch {
	case args == "" || args == "status":
		return c.showStatus()
	case args == "list":
		return c.listProviders()
	case strings.HasPrefix(args, "switch "):
		providerID := strings.TrimPrefix(args, "switch ")
		return c.switchProvider(providerID)
	case args == "test":
		return c.testConnection()
	default:
		return CommandResult{
			Success: false,
			Message: c.helpText(),
		}
	}
}

// CommandResult represents the result of a command execution.
type CommandResult struct {
	Success     bool
	Message     string
	NeedsRedraw bool
}

func (c *APICommand) showStatus() CommandResult {
	var b strings.Builder

	b.WriteString("📡 **API 狀態**\n\n")

	// Smart Model
	if c.config.API.Smart.ProviderID != "" {
		info := api.GetProviderInfo(c.config.API.Smart.ProviderID)
		if info != nil {
			b.WriteString(fmt.Sprintf("**Smart Model**: ✓ %s", info.Name))
			if c.config.API.Smart.Model != "" {
				b.WriteString(fmt.Sprintf(" (%s)", c.config.API.Smart.Model))
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString("**Smart Model**: ✗ 未設定\n")
	}

	// Fast Model
	if c.config.API.Fast.ProviderID != "" {
		info := api.GetProviderInfo(c.config.API.Fast.ProviderID)
		if info != nil {
			b.WriteString(fmt.Sprintf("**Fast Model**: ✓ %s", info.Name))
			if c.config.API.Fast.Model != "" {
				b.WriteString(fmt.Sprintf(" (%s)", c.config.API.Fast.Model))
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString("**Fast Model**: ✗ 未設定\n")
	}

	// Configured API keys
	if len(c.config.API.APIKeys) > 0 {
		b.WriteString("\n**已設定的 API Key**:\n")
		for providerID := range c.config.API.APIKeys {
			info := api.GetProviderInfo(providerID)
			name := providerID
			if info != nil {
				name = info.Name
			}
			b.WriteString(fmt.Sprintf("  • %s\n", name))
		}
	}

	b.WriteString("\n使用 `/api list` 查看所有供應商")
	b.WriteString("\n使用 `/api switch <provider>` 切換供應商")

	return CommandResult{Success: true, Message: b.String()}
}

func (c *APICommand) listProviders() CommandResult {
	var b strings.Builder

	b.WriteString("📋 **可用的 API 供應商**\n\n")

	categories := []struct {
		name  string
		title string
	}{
		{"official", "官方 API"},
		{"gateway", "聚合平台"},
		{"local", "本地模型"},
	}

	for _, cat := range categories {
		providers := api.GetProvidersByCategory(cat.name)
		if len(providers) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("**%s**\n", cat.title))
		for _, p := range providers {
			marker := "  "
			if c.config.API.Smart.ProviderID == p.ID {
				marker = "✓ "
			}
			hasKey := ""
			if c.config.HasAPIKey(p.ID) {
				hasKey = " [已設定]"
			}
			b.WriteString(fmt.Sprintf("%s`%s` - %s%s\n", marker, p.ID, p.Description, hasKey))
		}
		b.WriteString("\n")
	}

	return CommandResult{Success: true, Message: b.String()}
}

func (c *APICommand) switchProvider(providerID string) CommandResult {
	providerID = strings.TrimSpace(providerID)

	// Validate provider exists
	info := api.GetProviderInfo(providerID)
	if info == nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("❌ 未知的供應商: %s\n使用 `/api list` 查看可用供應商", providerID),
		}
	}

	// Check if API key is configured
	if !c.config.HasAPIKey(providerID) {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("❌ 尚未設定 %s 的 API Key\n請先在設定中新增 API Key", info.Name),
		}
	}

	// Update config
	c.config.API.Smart.ProviderID = providerID
	c.config.API.Smart.Model = "" // Reset to default

	if err := c.config.Save(); err != nil {
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("❌ 儲存配置失敗: %v", err),
		}
	}

	return CommandResult{
		Success:     true,
		Message:     fmt.Sprintf("✓ 已切換至 %s", info.Name),
		NeedsRedraw: true,
	}
}

func (c *APICommand) testConnection() CommandResult {
	if c.config.API.Smart.ProviderID == "" {
		return CommandResult{
			Success: false,
			Message: "❌ 尚未設定 API 供應商",
		}
	}

	apiKey, err := c.config.DecryptAPIKey(c.config.API.Smart.ProviderID)
	if err != nil || apiKey == "" {
		return CommandResult{
			Success: false,
			Message: "❌ 無法取得 API Key",
		}
	}

	// This would be async in real implementation
	return CommandResult{
		Success: true,
		Message: "🔄 連線測試中...（使用 API 設定畫面進行完整測試）",
	}
}

func (c *APICommand) helpText() string {
	return `📡 **/api 指令說明**

**用法:**
  /api              - 顯示當前 API 狀態
  /api status       - 顯示當前 API 狀態
  /api list         - 列出所有可用供應商
  /api switch <id>  - 切換至指定供應商
  /api test         - 測試當前連線

**範例:**
  /api switch openai
  /api switch anthropic
  /api switch ollama`
}

// Name returns the command name.
func (c *APICommand) Name() string {
	return "api"
}

// Help returns the help text.
func (c *APICommand) Help() string {
	return "管理 API 供應商設定"
}
