package save

import (
	"fmt"
	"os"
	"path/filepath"
)

// MaxSaveSizeBytes is the maximum recommended save file size (1MB).
const MaxSaveSizeBytes = 1024 * 1024

// MinSlotID is the minimum valid slot ID.
const MinSlotID = 1

// MaxSlotID is the maximum valid slot ID.
const MaxSlotID = 3

// GetSaveDir returns the directory path for save files.
func GetSaveDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".nightmare", "saves")
	}
	return filepath.Join(homeDir, ".nightmare", "saves")
}

// GetSavePath returns the full path for a save slot.
func GetSavePath(slotID int) string {
	return filepath.Join(GetSaveDir(), fmt.Sprintf("save_%d.json", slotID))
}

// EnsureSaveDir creates the save directory if it doesn't exist.
func EnsureSaveDir() error {
	return EnsureSaveDirAt(GetSaveDir())
}

// EnsureSaveDirAt creates the save directory at a specific path if it doesn't exist.
func EnsureSaveDirAt(dir string) error {
	return os.MkdirAll(dir, 0700)
}

// CheckSaveSize returns a warning message if the save size exceeds the limit.
func CheckSaveSize(sizeBytes int64) string {
	if sizeBytes > MaxSaveSizeBytes {
		return fmt.Sprintf("警告：存檔大小 (%d bytes) 超過建議上限 (%d bytes)。考慮壓縮故事上下文。",
			sizeBytes, MaxSaveSizeBytes)
	}
	return ""
}

// SaveSlotExists checks if a save slot file exists at the default location.
func SaveSlotExists(slotID int) bool {
	return SaveSlotExistsAt(GetSavePath(slotID))
}

// SaveSlotExistsAt checks if a save file exists at the given path.
func SaveSlotExistsAt(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// GetAllSaveSlots returns a list of existing save slot IDs at the default location.
func GetAllSaveSlots() []int {
	return GetAllSaveSlotsAt(GetSaveDir())
}

// GetAllSaveSlotsAt returns a list of existing save slot IDs at the given directory.
func GetAllSaveSlotsAt(dir string) []int {
	var slots []int

	for slotID := MinSlotID; slotID <= MaxSlotID; slotID++ {
		savePath := filepath.Join(dir, fmt.Sprintf("save_%d.json", slotID))
		if SaveSlotExistsAt(savePath) {
			slots = append(slots, slotID)
		}
	}

	return slots
}

// ValidateSlotID validates that a slot ID is within the valid range.
func ValidateSlotID(slotID int) error {
	if slotID < MinSlotID || slotID > MaxSlotID {
		return fmt.Errorf("無效的存檔槽位：%d（有效範圍：%d-%d）", slotID, MinSlotID, MaxSlotID)
	}
	return nil
}
