package commands

import (
	"fmt"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/game/npc"
)

// TeamCommand displays teammate status information
type TeamCommand struct {
	teammates []*npc.Teammate
}

// NewTeamCommand creates a new team command with given teammates
func NewTeamCommand(teammates []*npc.Teammate) *TeamCommand {
	return &TeamCommand{
		teammates: teammates,
	}
}

// Name returns the command name
func (c *TeamCommand) Name() string {
	return "team"
}

// Help returns help text for the command
func (c *TeamCommand) Help() string {
	return "顯示所有隊友的狀態資訊 / Display all teammate status information"
}

// Execute executes the team command
func (c *TeamCommand) Execute(args []string) (string, error) {
	if len(c.teammates) == 0 {
		return "目前沒有隊友 / No teammates currently", nil
	}

	var output strings.Builder
	output.WriteString("═══ 隊友狀態 / Teammate Status ═══\n\n")

	for i, tm := range c.teammates {
		// Skip nil teammates
		if tm == nil {
			continue
		}

		// Teammate header with name and archetype
		output.WriteString(fmt.Sprintf("【%s】(%s)\n", tm.Name, tm.Archetype))

		// HP and Status
		statusIcon := getStatusIcon(tm.Status.Condition, tm.Status.Alive)
		output.WriteString(fmt.Sprintf("  %s HP: %d/100 | 狀態: %s\n",
			statusIcon, tm.HP, getConditionText(tm.Status.Condition, tm.Status.Alive)))

		// Location
		if tm.Location != "" {
			output.WriteString(fmt.Sprintf("  📍 位置: %s\n", tm.Location))
		}

		// Inventory
		if len(tm.Inventory) > 0 {
			output.WriteString("  🎒 攜帶物品:\n")
			for _, item := range tm.Inventory {
				output.WriteString(fmt.Sprintf("     - %s\n", item.Name))
			}
		} else {
			output.WriteString("  🎒 攜帶物品: (無)\n")
		}

		// Emotional state based on HP and condition
		emotionalState := getEmotionalState(tm)
		if emotionalState != "" {
			output.WriteString(fmt.Sprintf("  💭 情緒: %s\n", emotionalState))
		}

		// Separator between teammates
		if i < len(c.teammates)-1 {
			output.WriteString("\n───────────────────────────────\n\n")
		}
	}

	return output.String(), nil
}

// getStatusIcon returns an icon based on condition and alive status
func getStatusIcon(condition string, alive bool) string {
	if !alive {
		return "💀"
	}
	switch condition {
	case "healthy":
		return "✓"
	case "injured":
		return "⚠"
	case "critical":
		return "⚠⚠"
	default:
		return "?"
	}
}

// getConditionText returns localized condition text
func getConditionText(condition string, alive bool) string {
	if !alive {
		return "已死亡 / dead"
	}
	switch condition {
	case "healthy":
		return "健康 / healthy"
	case "injured":
		return "受傷 / injured"
	case "critical":
		return "危急 / critical"
	default:
		return condition
	}
}

// getEmotionalState determines emotional state based on teammate status
func getEmotionalState(tm *npc.Teammate) string {
	if !tm.Status.Alive {
		return ""
	}

	// Use the teammate's actual EmotionalState field if set
	if tm.EmotionalState != "" {
		return getEmotionalStateText(tm.EmotionalState)
	}

	// Fallback: Emotional state based on HP and archetype (for backwards compatibility)
	if tm.HP >= 80 {
		switch tm.Archetype {
		case npc.ArchetypeVictim:
			return "緊張不安 / nervous"
		case npc.ArchetypeLogic:
			return "冷靜思考 / calm and thinking"
		case npc.ArchetypeIntuition:
			return "保持警覺 / alert"
		default:
			return "正常 / normal"
		}
	} else if tm.HP >= 50 {
		switch tm.Archetype {
		case npc.ArchetypeVictim:
			return "恐慌 / panicking"
		case npc.ArchetypeLogic:
			return "專注分析 / focused analysis"
		case npc.ArchetypeIntuition:
			return "高度警戒 / highly alert"
		default:
			return "緊張 / tense"
		}
	} else if tm.HP >= 20 {
		return "極度恐懼 / terrified"
	} else {
		return "瀕臨崩潰 / near breakdown"
	}
}

// getEmotionalStateText returns localized text for EmotionalState
func getEmotionalStateText(state npc.EmotionalState) string {
	switch state {
	case npc.EmotionCalm:
		return "平靜 / calm"
	case npc.EmotionAnxious:
		return "焦慮 / anxious"
	case npc.EmotionPanicked:
		return "恐慌 / panicked"
	case npc.EmotionRelieved:
		return "鬆了一口氣 / relieved"
	case npc.EmotionGrieving:
		return "悲傷 / grieving"
	default:
		return string(state)
	}
}
