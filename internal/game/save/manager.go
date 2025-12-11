package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SlotInfo contains information about a save slot.
type SlotInfo struct {
	SlotID    int       `json:"slot_id"`
	IsEmpty   bool      `json:"is_empty"`
	Chapter   int       `json:"chapter"`
	PlayTime  int       `json:"play_time_seconds"`
	SavedAt   time.Time `json:"saved_at"`
	Location  string    `json:"location"`
}

// SaveManager handles save/load operations.
type SaveManager struct {
	SaveDir string
	mu      sync.RWMutex
}

// NewSaveManager creates a new SaveManager with the specified save directory.
func NewSaveManager(saveDir string) *SaveManager {
	return &SaveManager{
		SaveDir: saveDir,
	}
}

// NewDefaultSaveManager creates a SaveManager using the default save directory.
func NewDefaultSaveManager() *SaveManager {
	return NewSaveManager(GetSaveDir())
}

// Save saves the game state to the specified slot.
func (m *SaveManager) Save(slotID int, data *SaveData) error {
	if err := ValidateSlotID(slotID); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure save directory exists
	if err := EnsureSaveDirAt(m.SaveDir); err != nil {
		return fmt.Errorf("無法創建存檔目錄：%w", err)
	}

	// Update metadata
	data.Metadata.SavedAt = time.Now()

	// Compute checksum
	if err := SetChecksum(data); err != nil {
		return fmt.Errorf("無法計算校驗碼：%w", err)
	}

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("無法序列化存檔資料：%w", err)
	}

	// Check size
	if warning := CheckSaveSize(int64(len(jsonData))); warning != "" {
		// Log warning but continue
		fmt.Println(warning)
	}

	// Use atomic write: write to temp file first, then rename
	savePath := filepath.Join(m.SaveDir, fmt.Sprintf("save_%d.json", slotID))
	tempPath := savePath + ".tmp"

	if err := os.WriteFile(tempPath, jsonData, 0600); err != nil {
		return fmt.Errorf("記憶無法固化...你的思緒在虛空中流失。（%w）", err)
	}

	// Rename temp file to final path (atomic on most systems)
	if err := os.Rename(tempPath, savePath); err != nil {
		// Clean up temp file
		os.Remove(tempPath)
		return fmt.Errorf("存檔失敗：%w", err)
	}

	return nil
}

// Load loads the game state from the specified slot.
func (m *SaveManager) Load(slotID int) (*SaveData, error) {
	if err := ValidateSlotID(slotID); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	savePath := filepath.Join(m.SaveDir, fmt.Sprintf("save_%d.json", slotID))

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("這個存檔槽位是空的，沒有可以喚醒的記憶。（槽位 %d 不存在）", slotID)
	}

	// Read file
	jsonData, err := os.ReadFile(savePath)
	if err != nil {
		return nil, fmt.Errorf("無法讀取存檔：%w", err)
	}

	// Parse JSON
	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("這段記憶已被扭曲...無法還原現實。（存檔格式無效：%w）", err)
	}

	// Verify checksum
	if err := VerifyChecksum(&data); err != nil {
		return nil, err
	}

	// Check version and migrate if needed
	if data.Version < CurrentVersion {
		// Parse as generic map for migration
		var rawData map[string]interface{}
		if err := json.Unmarshal(jsonData, &rawData); err != nil {
			return nil, fmt.Errorf("無法解析存檔資料：%w", err)
		}

		// Register v0->v1 migration if not already registered
		RegisterMigration(&V0ToV1Migration{})

		migratedData, err := MigrateData(rawData, data.Version, CurrentVersion)
		if err != nil {
			return nil, fmt.Errorf("版本遷移失敗：%w", err)
		}

		// Convert back to SaveData
		migratedJSON, err := json.Marshal(migratedData)
		if err != nil {
			return nil, fmt.Errorf("遷移後序列化失敗：%w", err)
		}

		if err := json.Unmarshal(migratedJSON, &data); err != nil {
			return nil, fmt.Errorf("遷移後解析失敗：%w", err)
		}
	}

	return &data, nil
}

// GetSlotInfo returns information about a save slot.
func (m *SaveManager) GetSlotInfo(slotID int) (*SlotInfo, error) {
	if err := ValidateSlotID(slotID); err != nil {
		return nil, err
	}

	info := &SlotInfo{
		SlotID:  slotID,
		IsEmpty: true,
	}

	savePath := filepath.Join(m.SaveDir, fmt.Sprintf("save_%d.json", slotID))

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return info, nil
	}

	// Read and parse to get slot info
	m.mu.RLock()
	defer m.mu.RUnlock()

	jsonData, err := os.ReadFile(savePath)
	if err != nil {
		return info, nil // Return empty slot if can't read
	}

	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return info, nil // Return empty slot if can't parse
	}

	info.IsEmpty = false
	info.Chapter = data.Game.CurrentChapter
	info.PlayTime = data.Metadata.PlayTime
	info.SavedAt = data.Metadata.SavedAt
	info.Location = data.Player.Location

	return info, nil
}

// GetAllSlotInfo returns information about all save slots.
func (m *SaveManager) GetAllSlotInfo() (map[int]*SlotInfo, error) {
	result := make(map[int]*SlotInfo)

	for slotID := MinSlotID; slotID <= MaxSlotID; slotID++ {
		info, err := m.GetSlotInfo(slotID)
		if err != nil {
			return nil, err
		}
		result[slotID] = info
	}

	return result, nil
}

// SlotExists checks if a save slot has data.
func (m *SaveManager) SlotExists(slotID int) bool {
	savePath := filepath.Join(m.SaveDir, fmt.Sprintf("save_%d.json", slotID))
	_, err := os.Stat(savePath)
	return err == nil
}

// DeleteSlot removes a save slot.
func (m *SaveManager) DeleteSlot(slotID int) error {
	if err := ValidateSlotID(slotID); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	savePath := filepath.Join(m.SaveDir, fmt.Sprintf("save_%d.json", slotID))

	// Remove if exists, ignore if not
	err := os.Remove(savePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("無法刪除存檔：%w", err)
	}

	return nil
}

// FormatPlayTime formats play time in seconds to a human-readable string.
func FormatPlayTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
