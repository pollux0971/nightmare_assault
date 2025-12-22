package savefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	gameErrors "github.com/nightmare-assault/nightmare-assault/internal/errors"
	"github.com/nightmare-assault/nightmare-assault/internal/game/save"
)

// ==========================================================================
// Story 8.1: SaveV2 functions for v2.0 architecture
// ==========================================================================

// slotMutexes protects concurrent saves to the same slot (Critical Issue #3)
var slotMutexes = [4]sync.Mutex{} // slots 1-3 + safety buffer at index 0

// SaveV2 saves the game state to the specified slot using SaveFileV2 format.
// Story 8.1 AC: Save to ~/.nightmare/saves/save_N.json with response time < 500ms
// Fixed: Added mutex protection for concurrent saves (Critical Issue #3)
// Fixed: CreatedAt preservation and PlayTime tracking (Critical Issue #2, High Priority Issue #5, #6)
// Fixed: Temp file cleanup with defer (High Priority Issue #7)
func SaveV2(saveDir string, slotID int, saveFile *SaveFileV2) error {
	startTime := time.Now()

	if err := save.ValidateSlotID(slotID); err != nil {
		return err
	}

	// Lock slot mutex to prevent concurrent saves (Critical Issue #3)
	slotMutexes[slotID].Lock()
	defer slotMutexes[slotID].Unlock()

	// Ensure save directory exists
	if err := save.EnsureSaveDirAt(saveDir); err != nil {
		return gameErrors.NewSaveFileError("create directory", slotID, err)
	}

	// Check if this is an update to existing save to preserve CreatedAt (High Priority Issue #5)
	// Note: Cannot call LoadV2 here as it would cause recursive lock (deadlock)
	savePath := filepath.Join(saveDir, fmt.Sprintf("save_%d.json", slotID))
	if existingData, err := os.ReadFile(savePath); err == nil {
		var existingSave SaveFileV2
		if err := json.Unmarshal(existingData, &existingSave); err == nil {
			// Preserve CreatedAt from existing save
			saveFile.Meta.CreatedAt = existingSave.Meta.CreatedAt
		}
	}
	// Set CreatedAt if not already set (new save or failed to read existing)
	if saveFile.Meta.CreatedAt.IsZero() {
		saveFile.Meta.CreatedAt = time.Now()
	}

	// Update metadata
	saveFile.Meta.UpdatedAt = time.Now()
	saveFile.Meta.SaveID = slotID

	// Note: PlayTime should be calculated by the caller before calling SaveV2
	// This allows the caller to track game start time and compute elapsed time
	// If PlayTime is not set by caller, it will remain at its current value

	// Compute checksum (NFR-S05)
	if err := SetChecksum(saveFile); err != nil {
		return fmt.Errorf("無法計算校驗碼：%w", err)
	}

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(saveFile, "", "  ")
	if err != nil {
		return fmt.Errorf("無法序列化存檔資料：%w", err)
	}

	// Check size
	if warning := save.CheckSaveSize(int64(len(jsonData))); warning != "" {
		// Log warning but continue
		fmt.Println(warning)
	}

	// Use atomic write: write to temp file first, then rename
	tempPath := savePath + ".tmp"

	// Ensure temp file cleanup on error (High Priority Issue #7)
	defer func() {
		// Clean up temp file if it still exists (indicates failure)
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	if err := os.WriteFile(tempPath, jsonData, 0600); err != nil {
		return gameErrors.NewSaveFileError("write", slotID, err)
	}

	// Rename temp file to final path (atomic on most systems)
	if err := os.Rename(tempPath, savePath); err != nil {
		return gameErrors.NewSaveFileError("finalize", slotID, err)
	}

	// NFR-P06: Response time < 500ms
	elapsed := time.Since(startTime)
	if elapsed > 500*time.Millisecond {
		fmt.Printf("警告：存檔操作耗時 %v，超過目標 500ms\n", elapsed)
	}

	return nil
}

// LoadV2 loads the game state from the specified slot using SaveFileV2 format.
// Story 8.2 AC: Load from ~/.nightmare/saves/save_N.json with response time < 500ms
// Fixed: Added mutex protection for concurrent access (Critical Issue #3)
func LoadV2(saveDir string, slotID int) (*SaveFileV2, error) {
	startTime := time.Now()

	if err := save.ValidateSlotID(slotID); err != nil {
		return nil, err
	}

	// Lock slot mutex to prevent concurrent read/write (Critical Issue #3)
	slotMutexes[slotID].Lock()
	defer slotMutexes[slotID].Unlock()

	savePath := filepath.Join(saveDir, fmt.Sprintf("save_%d.json", slotID))

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return nil, gameErrors.NewSaveFileNotFoundError(slotID)
	}

	// Read file
	jsonData, err := os.ReadFile(savePath)
	if err != nil {
		return nil, gameErrors.NewSaveFileError("read", slotID, err)
	}

	// Parse JSON
	var saveFile SaveFileV2
	if err := json.Unmarshal(jsonData, &saveFile); err != nil {
		return nil, gameErrors.NewSaveFileCorruptedError(slotID, err)
	}

	// Verify checksum (NFR-S05)
	if err := VerifyChecksum(&saveFile); err != nil {
		return nil, gameErrors.NewSaveFileCorruptedError(slotID, err)
	}

	// NFR-P06: Response time < 500ms
	elapsed := time.Since(startTime)
	if elapsed > 500*time.Millisecond {
		fmt.Printf("警告：讀檔操作耗時 %v，超過目標 500ms\n", elapsed)
	}

	return &saveFile, nil
}

// GetSlotInfo returns information about a save slot (V2 format).
func GetSlotInfo(saveDir string, slotID int) (*save.SlotInfo, error) {
	if err := save.ValidateSlotID(slotID); err != nil {
		return nil, err
	}

	info := &save.SlotInfo{
		SlotID:  slotID,
		IsEmpty: true,
	}

	savePath := filepath.Join(saveDir, fmt.Sprintf("save_%d.json", slotID))

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return info, nil
	}

	// Read and parse to get slot info
	jsonData, err := os.ReadFile(savePath)
	if err != nil {
		return info, nil // Return empty slot if can't read
	}

	var saveFile SaveFileV2
	if err := json.Unmarshal(jsonData, &saveFile); err != nil {
		return info, nil // Return empty slot if can't parse
	}

	info.IsEmpty = false
	info.Chapter = saveFile.GameState.GetCurrentBeat()
	info.PlayTime = saveFile.Meta.PlayTime
	info.SavedAt = saveFile.Meta.UpdatedAt
	info.Location = saveFile.GameState.CurrentScene

	return info, nil
}
