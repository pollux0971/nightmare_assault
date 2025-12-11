// Package commands provides slash command implementations.
package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/game"
	"github.com/nightmare-assault/nightmare-assault/internal/game/hint"
)

// HintCommand implements the /hint command.
type HintCommand struct {
	// GetDifficulty returns the current game difficulty.
	GetDifficulty func() game.DifficultyLevel
	// GetSAN returns the current SAN value.
	GetSAN func() int
	// GetChapter returns the current chapter.
	GetChapter func() int
	// DeductSAN deducts SAN and returns the new value.
	DeductSAN func(amount int) int
	// GetCurrentScene returns the current scene description.
	GetCurrentScene func() string
	// GetRecentActions returns recent player actions.
	GetRecentActions func() []string
	// GetTurnsSinceMove returns turns since last location change.
	GetTurnsSinceMove func() int
	// GetMissedClues returns missed clues for current rules.
	GetMissedClues func() []string
	// GetUpcomingRules returns rules that might trigger soon.
	GetUpcomingRules func() []string
	// GetAvailableItems returns items the player has.
	GetAvailableItems func() []string
	// GenerateLLMHint calls the LLM to generate a hint.
	GenerateLLMHint func(prompt string) (string, error)
	// HintState tracks hint usage.
	HintState *hint.HintState
	// Generator is the hint generator.
	Generator *hint.Generator

	// confirmMode tracks if we're waiting for confirmation
	confirmMode bool
	// pendingCost is the SAN cost if confirmed
	pendingCost int
}

// NewHintCommand creates a new hint command.
func NewHintCommand() *HintCommand {
	return &HintCommand{
		Generator: hint.NewGenerator(),
	}
}

// Name returns the command name.
func (c *HintCommand) Name() string {
	return "hint"
}

// Help returns command help text.
func (c *HintCommand) Help() string {
	return "花費 SAN 獲得模糊提示 (用法: /hint)"
}

// Execute runs the hint command.
func (c *HintCommand) Execute(args []string) (string, error) {
	// Check if in confirmation mode
	if c.confirmMode {
		return c.handleConfirmation(args)
	}

	// Check difficulty (AC4)
	if c.GetDifficulty != nil && c.GetDifficulty() == game.DifficultyHell {
		return "在這個噩夢中，沒有人能幫助你。", nil
	}

	// Check if hints are enabled in state
	if c.HintState != nil && !c.HintState.IsEnabled() {
		return "在這個噩夢中，沒有人能幫助你。", nil
	}

	// Get current state
	currentSAN := 100
	if c.GetSAN != nil {
		currentSAN = c.GetSAN()
	}

	chapter := 1
	if c.GetChapter != nil {
		chapter = c.GetChapter()
	}

	// Calculate cost (AC6)
	cost := c.Generator.GetCurrentCost(chapter)

	// Check if can afford (AC2)
	if currentSAN < cost {
		return fmt.Sprintf("你的理智不足以清晰思考...（需要 %d SAN，當前 %d SAN）", cost, currentSAN), nil
	}

	// Enter confirmation mode (AC1)
	c.confirmMode = true
	c.pendingCost = cost

	usageCount := c.Generator.GetUsageCount(chapter)
	usageText := ""
	if usageCount > 0 {
		usageText = fmt.Sprintf("（本章第 %d 次提示，消耗遞增）", usageCount+1)
	}

	return fmt.Sprintf("花費 %d SAN 獲得提示？%s\n當前 SAN：%d\n\n輸入 y 確認，n 取消",
		cost, usageText, currentSAN), nil
}

// handleConfirmation handles the y/n confirmation.
func (c *HintCommand) handleConfirmation(args []string) (string, error) {
	c.confirmMode = false

	if len(args) == 0 {
		return "取消提示請求。", nil
	}

	response := strings.ToLower(strings.TrimSpace(args[0]))
	if response != "y" && response != "yes" && response != "是" {
		return "取消提示請求。", nil
	}

	// Deduct SAN
	if c.DeductSAN != nil {
		c.DeductSAN(c.pendingCost)
	}

	// Get chapter
	chapter := 1
	if c.GetChapter != nil {
		chapter = c.GetChapter()
	}

	// Record usage
	c.Generator.IncrementUsage(chapter)
	if c.HintState != nil {
		c.HintState.RecordUsage(chapter)
	}

	// Build context
	ctx := c.buildHintContext(chapter)

	// Try to generate LLM hint
	var llmHint string
	if c.GenerateLLMHint != nil {
		prompt := hint.BuildHintPrompt(ctx)
		generated, err := c.GenerateLLMHint(prompt)
		if err == nil && generated != "" {
			llmHint = generated
		}
	}

	// Generate hint result
	result := c.Generator.Generate(ctx, llmHint)

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("【%s】\n\n", result.Type.String()))
	output.WriteString(result.Text)
	output.WriteString(fmt.Sprintf("\n\n（消耗 %d SAN）", result.SANCost))

	return output.String(), nil
}

// buildHintContext creates the hint context from current state.
func (c *HintCommand) buildHintContext(chapter int) *hint.HintContext {
	analyzer := hint.NewStateAnalyzer()

	currentScene := "未知場景"
	if c.GetCurrentScene != nil {
		currentScene = c.GetCurrentScene()
	}

	var recentActions []string
	if c.GetRecentActions != nil {
		recentActions = c.GetRecentActions()
	}

	turnsSinceMove := 0
	if c.GetTurnsSinceMove != nil {
		turnsSinceMove = c.GetTurnsSinceMove()
	}

	var missedClues []string
	if c.GetMissedClues != nil {
		missedClues = c.GetMissedClues()
	}

	var upcomingRules []string
	if c.GetUpcomingRules != nil {
		upcomingRules = c.GetUpcomingRules()
	}

	var availableItems []string
	if c.GetAvailableItems != nil {
		availableItems = c.GetAvailableItems()
	}

	return analyzer.Analyze(
		currentScene,
		recentActions,
		turnsSinceMove,
		chapter,
		missedClues,
		upcomingRules,
		availableItems,
	)
}

// IsConfirmMode returns if the command is waiting for confirmation.
func (c *HintCommand) IsConfirmMode() bool {
	return c.confirmMode
}

// CancelConfirmation cancels the pending confirmation.
func (c *HintCommand) CancelConfirmation() {
	c.confirmMode = false
	c.pendingCost = 0
}

// SetHintState sets the hint state for tracking.
func (c *HintCommand) SetHintState(state *hint.HintState) {
	c.HintState = state
}

// GetPendingCost returns the pending SAN cost.
func (c *HintCommand) GetPendingCost() int {
	return c.pendingCost
}
