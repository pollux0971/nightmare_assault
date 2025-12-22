package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nightmare-assault/nightmare-assault/internal/config"
)

// APISetupInterface defines the common interface for API setup models
type APISetupInterface interface {
	tea.Model
	IsDone() bool
	GetConfig() *config.Config
}

// Ensure all models implement the interface
var _ APISetupInterface = (*APISetupModel)(nil)
var _ APISetupInterface = (*APISetupModelV2)(nil)
var _ APISetupInterface = (*TrinitySetupModel)(nil)
