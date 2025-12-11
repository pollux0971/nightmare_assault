package commands

import (
	"fmt"
	"strconv"
	"strings"
)

// SlotCommand represents a save/load command definition.
type SlotCommand struct {
	Name        string
	Description string
	Usage       string
}

// ParseSlotCommand parses a save or load command and returns the command name and optional slot ID.
// If no slot is specified, returns 0 as the slot ID.
func ParseSlotCommand(input string) (cmd string, slotID int, err error) {
	input = strings.TrimSpace(input)

	// Check if it starts with /
	if !strings.HasPrefix(input, "/") {
		return "", 0, fmt.Errorf("指令必須以 / 開頭")
	}

	// Remove the leading /
	input = input[1:]

	// Split by space
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", 0, fmt.Errorf("空指令")
	}

	cmd = strings.ToLower(parts[0])

	// Validate command
	if cmd != "save" && cmd != "load" {
		return "", 0, fmt.Errorf("未知指令: %s", cmd)
	}

	// Parse optional slot ID
	if len(parts) > 1 {
		slotID, err = strconv.Atoi(parts[1])
		if err != nil {
			return cmd, 0, fmt.Errorf("無效的槽位編號: %s", parts[1])
		}

		// Validate slot range (1-3)
		if slotID < 1 || slotID > 3 {
			return cmd, 0, fmt.Errorf("無效的存檔槽位：%d（有效範圍：1-3）", slotID)
		}
	}

	return cmd, slotID, nil
}

// IsSaveCommand checks if the input is a save command.
func IsSaveCommand(input string) bool {
	input = strings.TrimSpace(strings.ToLower(input))
	return strings.HasPrefix(input, "/save") && !strings.HasPrefix(input, "/savegame")
}

// IsLoadCommand checks if the input is a load command.
func IsLoadCommand(input string) bool {
	input = strings.TrimSpace(strings.ToLower(input))
	return strings.HasPrefix(input, "/load") && !strings.HasPrefix(input, "/loadgame")
}

// SaveCommandHelp returns help text for the save command.
func SaveCommandHelp() string {
	return `存檔指令

用法:
  /save        - 顯示存檔槽位選擇介面
  /save <1-3>  - 直接存檔至指定槽位

範例:
  /save        開啟存檔選擇介面
  /save 1      存檔至槽位 1
  /save 2      存檔至槽位 2

注意:
  - 存檔會覆蓋該槽位的舊資料
  - 系統會在覆蓋前詢問確認`
}

// LoadCommandHelp returns help text for the load command.
func LoadCommandHelp() string {
	return `讀檔指令

用法:
  /load        - 顯示可用存檔清單
  /load <1-3>  - 直接載入指定槽位

範例:
  /load        開啟讀檔選擇介面
  /load 1      載入槽位 1 的存檔
  /load 2      載入槽位 2 的存檔

注意:
  - 只能載入有資料的存檔槽位
  - 讀檔會取代當前遊戲進度`
}

// GetAllSlotCommands returns all save/load related commands.
func GetAllSlotCommands() []SlotCommand {
	return []SlotCommand{
		{
			Name:        "save",
			Description: "保存遊戲進度",
			Usage:       "/save [槽位編號]",
		},
		{
			Name:        "load",
			Description: "載入遊戲存檔",
			Usage:       "/load [槽位編號]",
		},
	}
}
