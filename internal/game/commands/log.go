package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	"github.com/nightmare-assault/nightmare-assault/internal/i18n"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

const (
	// DefaultLogCount is the default number of log entries to display.
	DefaultLogCount = 10
	// MaxLogCount is the maximum number of log entries to display.
	MaxLogCount = 100
)

// ParseLogCommand parses a /log command and returns the number of entries to display.
// Supports:
// - /log       -> returns DefaultLogCount (10)
// - /log 20    -> returns 20
// - /log 150   -> returns MaxLogCount (100, capped)
func ParseLogCommand(input string) (count int, err error) {
	input = strings.TrimSpace(input)

	// Remove leading slash if present
	input = strings.TrimPrefix(input, "/")

	parts := strings.Fields(input)

	if len(parts) == 0 {
		return 0, fmt.Errorf("empty command")
	}

	// Check if it's a log command
	if parts[0] != "log" {
		return 0, fmt.Errorf("not a log command")
	}

	// Default case: /log with no arguments
	if len(parts) == 1 {
		return DefaultLogCount, nil
	}

	// Parse count argument
	if len(parts) == 2 {
		parsedCount, parseErr := strconv.Atoi(parts[1])
		if parseErr != nil {
			return 0, fmt.Errorf("invalid count: %s", parts[1])
		}

		if parsedCount <= 0 {
			return 0, fmt.Errorf("count must be positive: %d", parsedCount)
		}

		// Cap at MaxLogCount
		if parsedCount > MaxLogCount {
			return MaxLogCount, nil
		}

		return parsedCount, nil
	}

	// Too many arguments
	return 0, fmt.Errorf("too many arguments")
}

// IsLogCommand checks if the input is a /log command.
func IsLogCommand(input string) bool {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "/")
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return false
	}

	return parts[0] == "log"
}

// ==========================================================================
// Story 7.7: LogCommand - NPC Dialogue History Display
// ==========================================================================

// LogCommand displays NPC dialogue history and manages log levels.
// Story 7.7 AC #6: Support history review via /log command.
// Story 10-4 AC #5: Support /log level command to set log level.
type LogCommand struct {
	gameState  *engine.GameStateV2
	translator *i18n.Translator
}

// NewLogCommand creates a new log command.
func NewLogCommand(gameState *engine.GameStateV2) *LogCommand {
	return &LogCommand{
		gameState:  gameState,
		translator: i18n.GetGlobal(),
	}
}

// NewLogCommandWithTranslator creates a new log command with translator.
func NewLogCommandWithTranslator(gameState *engine.GameStateV2, translator *i18n.Translator) *LogCommand {
	return &LogCommand{
		gameState:  gameState,
		translator: translator,
	}
}

// Name returns the command name.
func (c *LogCommand) Name() string {
	return "log"
}

// Help returns the command help text.
func (c *LogCommand) Help() string {
	if c.translator != nil {
		return c.translator.T("commands.log_desc")
	}
	return "查看遊戲日誌或設定日誌級別 / View game logs or set log level"
}

// Execute displays NPC dialogue history or manages log level.
//
// Usage:
//   - /log              : Show last 10 dialogues (default)
//   - /log 20           : Show last 20 dialogues
//   - /log page 2       : Show page 2 (records 11-20)
//   - /log page 3 20    : Show page 3 with 20 records per page
//   - /log stats        : Show dialogue statistics
//   - /log level        : Show current log level
//   - /log level debug  : Set log level to DEBUG
//   - /log level info   : Set log level to INFO
//   - /log level warn   : Set log level to WARN
//   - /log level error  : Set log level to ERROR
//
// Story 7.7 AC #6: Support history review via /log command.
// Story 10-4 AC: Support pagination for viewing earlier records.
// Story 10-4 AC #5: Support /log level command to set log level.
func (c *LogCommand) Execute(args []string) (string, error) {
	// 處理 /log level 指令
	if len(args) > 0 && args[0] == "level" {
		return c.executeLogLevel(args[1:])
	}

	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("              LOG / NPC 對話記錄\n")
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	if c.gameState == nil {
		output.WriteString("❌ 遊戲狀態未初始化\n")
		return output.String(), nil
	}

	// Parse arguments
	maxRecords := DefaultLogCount
	pageNumber := 1
	showStats := false

	if len(args) > 0 {
		if args[0] == "stats" {
			showStats = true
		} else if args[0] == "page" {
			// Story 10-4 AC: 支援翻頁查看更早記錄
			// /log page N [count]
			if len(args) >= 2 {
				if p, err := strconv.Atoi(args[1]); err == nil && p > 0 {
					pageNumber = p
				}
			}
			if len(args) >= 3 {
				if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
					maxRecords = n
					if maxRecords > MaxLogCount {
						maxRecords = MaxLogCount
					}
				}
			}
		} else if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			maxRecords = n
			if maxRecords > MaxLogCount {
				maxRecords = MaxLogCount
			}
		}
	}

	if showStats {
		return c.executeStats(), nil
	}

	// Get dialogue history with pagination
	historyDisplay := c.gameState.GetDialogueHistoryForDisplayPaged(maxRecords, pageNumber)
	output.WriteString(historyDisplay)

	output.WriteString("\n═══════════════════════════════════════════════════\n")
	output.WriteString("💡 提示: /log stats 查看統計 | /log 20 查看最近20條\n")
	output.WriteString("💡 提示: /log page 2 翻頁 | /log page 2 20 每頁20條\n")
	output.WriteString("💡 提示: /log level [debug|info|warn|error] 設定日誌級別\n")

	return output.String(), nil
}

// executeLogLevel 處理日誌級別設定
// Story 10-4 AC #5: Support /log level command to set log level.
func (c *LogCommand) executeLogLevel(args []string) (string, error) {
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")

	// 使用翻譯，如果可用
	if c.translator != nil {
		output.WriteString(fmt.Sprintf("              %s\n", c.translator.T("log.title")))
	} else {
		output.WriteString("              日誌系統 / Log System\n")
	}

	output.WriteString("═══════════════════════════════════════════════════\n\n")

	// 如果沒有參數，顯示當前日誌級別
	if len(args) == 0 {
		currentLevel := logger.GetLevel()

		if c.translator != nil {
			output.WriteString(fmt.Sprintf("%s: %s\n\n",
				c.translator.T("log.current_level"),
				currentLevel.String()))

			output.WriteString(fmt.Sprintf("%s:\n", c.translator.T("log.available_levels")))
			output.WriteString(fmt.Sprintf("  • %s\n", c.translator.T("log.level_debug")))
			output.WriteString(fmt.Sprintf("  • %s\n", c.translator.T("log.level_info")))
			output.WriteString(fmt.Sprintf("  • %s\n", c.translator.T("log.level_warn")))
			output.WriteString(fmt.Sprintf("  • %s\n", c.translator.T("log.level_error")))
			output.WriteString(fmt.Sprintf("\n%s\n", c.translator.T("log.usage_hint")))
		} else {
			output.WriteString(fmt.Sprintf("當前日誌級別 / Current Level: %s\n\n", currentLevel.String()))
			output.WriteString("可用級別 / Available Levels:\n")
			output.WriteString("  • DEBUG - 詳細的調試信息 / Detailed debugging information\n")
			output.WriteString("  • INFO  - 一般信息 / General information\n")
			output.WriteString("  • WARN  - 警告信息 / Warning messages\n")
			output.WriteString("  • ERROR - 錯誤信息 / Error messages\n")
			output.WriteString("\n提示 / Hint: /log level [debug|info|warn|error]\n")
		}

		output.WriteString("═══════════════════════════════════════════════════\n")
		return output.String(), nil
	}

	// 設定日誌級別
	levelStr := args[0]
	newLevel, err := logger.ParseLogLevel(levelStr)
	if err != nil {
		if c.translator != nil {
			errMsg := c.translator.T("errors.invalid_log_level", levelStr)
			return "", fmt.Errorf("%s", errMsg)
		}
		return "", fmt.Errorf("無效的日誌級別 / Invalid log level: %s", levelStr)
	}

	// 設定全局日誌級別
	logger.SetLevel(newLevel)

	// 記錄級別變更
	oldLevel := logger.GetLevel()
	logger.Info("日誌級別已變更", map[string]interface{}{
		"old_level": oldLevel.String(),
		"new_level": newLevel.String(),
	})

	if c.translator != nil {
		output.WriteString(fmt.Sprintf("✅ %s\n",
			c.translator.T("messages.log_level_changed", newLevel.String())))
	} else {
		output.WriteString(fmt.Sprintf("✅ 日誌級別已設定為 / Log level set to: %s\n", newLevel.String()))
	}

	output.WriteString("═══════════════════════════════════════════════════\n")

	return output.String(), nil
}

// executeStats shows dialogue statistics.
func (c *LogCommand) executeStats() string {
	var output strings.Builder

	output.WriteString("═══════════════════════════════════════════════════\n")
	output.WriteString("              LOG / 對話統計\n")
	output.WriteString("═══════════════════════════════════════════════════\n\n")

	dh := c.gameState.GetDialogueHistory()
	if dh == nil {
		output.WriteString("❌ 對話歷史未初始化\n")
		return output.String()
	}

	stats := dh.GetStatsSummary()

	output.WriteString("📊 總計統計:\n")
	output.WriteString("───────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("  對話總數:     %d 條\n", stats.TotalDialogues))
	output.WriteString(fmt.Sprintf("  線索揭露:     %d 次\n", stats.ClueRevelations))
	output.WriteString(fmt.Sprintf("  臨終遺言:     %d 次\n", stats.DeathDialogues))
	output.WriteString(fmt.Sprintf("  玩家提問:     %d 次\n", stats.QuestionResponses))

	if len(stats.DialoguesByNPC) > 0 {
		output.WriteString("\n📋 各 NPC 對話次數:\n")
		output.WriteString("───────────────────────────────────────────────────\n")
		for npcName, count := range stats.DialoguesByNPC {
			output.WriteString(fmt.Sprintf("  %s: %d 條\n", npcName, count))
		}
	}

	output.WriteString("\n═══════════════════════════════════════════════════\n")

	return output.String()
}
