package savefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
	gameErrors "github.com/nightmare-assault/nightmare-assault/internal/errors"
	"github.com/nightmare-assault/nightmare-assault/internal/logger"
)

// ==========================================================================
// Story 8.5: Save Migration v1 → v2
// ==========================================================================
// AC1: 偵測舊版存檔格式
// AC2: 自動遷移到新格式（添加 NPCManager/GlobalFacts/ChatSessions）
// AC3: 遷移失敗時提供清晰錯誤訊息
// AC4: 備份舊存檔
// ==========================================================================

// SaveVersion represents the save file version.
type SaveVersion string

const (
	SaveVersionUnknown SaveVersion = "unknown"
	SaveVersionV1      SaveVersion = "v1.x"
	SaveVersionV2      SaveVersion = "v2.0"
)

// MigrationResult contains the result of a migration operation.
type MigrationResult struct {
	Success        bool
	FromVersion    SaveVersion
	ToVersion      SaveVersion
	BackupPath     string
	MigratedFields []string
	Errors         []string
	Warnings       []string
}

// DetectSaveVersion detects the version of a save file.
// AC1: Detect old save format
func DetectSaveVersion(savePath string) (SaveVersion, error) {
	data, err := os.ReadFile(savePath)
	if err != nil {
		return SaveVersionUnknown, fmt.Errorf("failed to read save file: %w", err)
	}

	// Try to parse as JSON
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return SaveVersionUnknown, fmt.Errorf("failed to parse save file as JSON: %w", err)
	}

	// Check for version indicator in meta
	if meta, ok := rawData["meta"].(map[string]interface{}); ok {
		if version, ok := meta["game_version"].(string); ok {
			if version >= "v2.0.0" {
				return SaveVersionV2, nil
			}
			return SaveVersionV1, nil
		}
	}

	// Check for v2-specific fields in game_state
	if gameState, ok := rawData["game_state"].(map[string]interface{}); ok {
		// V2 has these new fields
		hasNPCManager := gameState["npc_manager"] != nil
		hasGlobalFacts := gameState["global_facts"] != nil
		hasChatSessions := gameState["chat_sessions"] != nil
		hasMomentumConfig := gameState["momentum_config"] != nil

		// If it has any v2 fields, it's v2
		if hasNPCManager || hasGlobalFacts || hasChatSessions || hasMomentumConfig {
			return SaveVersionV2, nil
		}

		// Otherwise it's v1
		return SaveVersionV1, nil
	}

	return SaveVersionUnknown, fmt.Errorf("unable to determine save file version")
}

// BackupSaveFile creates a backup of the save file before migration.
// AC4: Backup old save file
func BackupSaveFile(savePath string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	dir := filepath.Dir(savePath)
	base := filepath.Base(savePath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	backupPath := filepath.Join(dir, fmt.Sprintf("%s_backup_%s%s", name, timestamp, ext))

	// Read original file
	data, err := os.ReadFile(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to read save file for backup: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	logger.Info("Save file backed up", map[string]interface{}{
		"original": savePath,
		"backup":   backupPath,
	})
	return backupPath, nil
}

// MigrateV1ToV2 migrates a v1 save file to v2 format.
// AC2: Automatically migrate to new format
// AC3: Provide clear error messages on migration failure
func MigrateV1ToV2(savePath string) (*MigrationResult, error) {
	result := &MigrationResult{
		FromVersion:    SaveVersionV1,
		ToVersion:      SaveVersionV2,
		MigratedFields: make([]string, 0),
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
	}

	logger.Info("Starting migration v1 → v2", map[string]interface{}{
		"path": savePath,
	})

	// AC4: Backup old save file
	backupPath, err := BackupSaveFile(savePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to backup save file: %v", err))
		return result, gameErrors.NewMigrationError("backup", err)
	}
	result.BackupPath = backupPath

	// Read the v1 save file
	data, err := os.ReadFile(savePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read save file: %v", err))
		return result, gameErrors.NewMigrationError("read", err)
	}

	// Parse as v2 structure (v1 is a subset of v2)
	var saveFile SaveFileV2
	if err := json.Unmarshal(data, &saveFile); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse save file: %v", err))
		return result, gameErrors.NewMigrationError("parse", err)
	}

	// AC2: Add missing v2 fields with default values

	// 1. MomentumConfig
	if saveFile.GameState.MomentumConfig == nil {
		saveFile.GameState.MomentumConfig = &engine.MomentumConfig{
			Frequency:    "medium",
			AutoResolve:  false,
			MaxAutoBeats: 5,
			PauseOnRisk:  "medium",
			PauseOnPlot:  true,
			PauseOnNPC:   true,
			PauseOnEvent: true,
		}
		result.MigratedFields = append(result.MigratedFields, "momentum_config")
		logger.Info("Added default MomentumConfig", nil)
	}

	// 2. NPCManager
	if saveFile.GameState.NPCManager == nil {
		saveFile.GameState.NPCManager = &engine.NPCManagerState{
			Profiles: make(map[string]interface{}),
			States:   make(map[string]interface{}),
			Config:   nil,
		}
		result.MigratedFields = append(result.MigratedFields, "npc_manager")
		logger.Info("Added empty NPCManager state", nil)
	}

	// 3. GlobalFacts
	if saveFile.GameState.GlobalFacts == nil {
		saveFile.GameState.GlobalFacts = make([]*engine.Fact, 0)
		result.MigratedFields = append(result.MigratedFields, "global_facts")
		logger.Info("Added empty GlobalFacts array", nil)
	}

	// 4. ChatSessions
	if saveFile.GameState.ChatSessions == nil {
		saveFile.GameState.ChatSessions = make([]*engine.ChatSession, 0)
		result.MigratedFields = append(result.MigratedFields, "chat_sessions")
		logger.Info("Added empty ChatSessions array", nil)
	}

	// 5. RuleWarnings (if missing)
	if saveFile.GameState.RuleWarnings == nil {
		saveFile.GameState.RuleWarnings = make(map[string]int)
		result.MigratedFields = append(result.MigratedFields, "rule_warnings")
		logger.Info("Added empty RuleWarnings map", nil)
	}

	// 6. Update game version in meta
	saveFile.Meta.GameVersion = "v2.0.0"
	result.MigratedFields = append(result.MigratedFields, "meta.game_version")

	// 7. Recalculate checksum
	checksum, err := ComputeChecksum(&saveFile)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to calculate checksum: %v", err))
		logger.Warn("Failed to calculate checksum during migration", map[string]interface{}{
			"error": err,
		})
	} else {
		saveFile.Checksum = checksum
		result.MigratedFields = append(result.MigratedFields, "checksum")
	}

	// Write the migrated save file
	migratedData, err := json.MarshalIndent(saveFile, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to marshal migrated save: %v", err))
		return result, gameErrors.NewMigrationError("marshal", err)
	}

	if err := os.WriteFile(savePath, migratedData, 0644); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to write migrated save: %v", err))
		return result, gameErrors.NewMigrationError("write", err)
	}

	result.Success = true
	logger.Info("Migration completed successfully", map[string]interface{}{
		"from":            result.FromVersion,
		"to":              result.ToVersion,
		"migrated_fields": len(result.MigratedFields),
		"backup":          result.BackupPath,
	})

	return result, nil
}

// AutoMigrate automatically detects and migrates a save file if needed.
// AC1: Detect old save format
// AC2: Automatically migrate to new format
// AC3: Provide clear error messages
func AutoMigrate(savePath string) (*MigrationResult, error) {
	// AC1: Detect version
	version, err := DetectSaveVersion(savePath)
	if err != nil {
		return nil, gameErrors.NewMigrationError("detect_version", err)
	}

	logger.Info("Detected save version", map[string]interface{}{
		"path":    savePath,
		"version": version,
	})

	// If already v2, no migration needed
	if version == SaveVersionV2 {
		return &MigrationResult{
			Success:     true,
			FromVersion: SaveVersionV2,
			ToVersion:   SaveVersionV2,
			Warnings:    []string{"Save file is already v2.0, no migration needed"},
		}, nil
	}

	// If v1, migrate to v2
	if version == SaveVersionV1 {
		return MigrateV1ToV2(savePath)
	}

	// Unknown version
	return nil, gameErrors.NewMigrationError("unknown_version",
		fmt.Errorf("unknown save version: %s", version))
}

// LoadV2WithMigration loads a save file and automatically migrates if needed.
// This is a convenience function that combines detection, migration, and loading.
func LoadV2WithMigration(saveDir string, slotID int) (*SaveFileV2, *MigrationResult, error) {
	savePath := filepath.Join(saveDir, fmt.Sprintf("save_%d.json", slotID))

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return nil, nil, gameErrors.NewSaveFileError("load", slotID,
			fmt.Errorf("save file does not exist"))
	}

	// Try to detect version
	version, err := DetectSaveVersion(savePath)
	if err != nil {
		logger.Warn("Failed to detect save version, assuming v2", map[string]interface{}{
			"error": err,
		})
		version = SaveVersionV2
	}

	var migrationResult *MigrationResult

	// If v1, auto-migrate
	if version == SaveVersionV1 {
		logger.Info("Detected v1 save file, auto-migrating", map[string]interface{}{
			"slot": slotID,
		})
		migrationResult, err = AutoMigrate(savePath)
		if err != nil {
			return nil, migrationResult, gameErrors.NewMigrationError("auto_migrate", err)
		}
		if !migrationResult.Success {
			return nil, migrationResult, gameErrors.NewMigrationError("migration_failed",
				fmt.Errorf("migration failed with errors: %v", migrationResult.Errors))
		}
	}

	// Load the (possibly migrated) save file
	saveFile, err := LoadV2(saveDir, slotID)
	if err != nil {
		return nil, migrationResult, err
	}

	return saveFile, migrationResult, nil
}

// RestoreFromBackup restores a save file from a backup.
// This is useful if migration fails and user wants to revert.
func RestoreFromBackup(backupPath, savePath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore save file: %w", err)
	}

	logger.Info("Save file restored from backup", map[string]interface{}{
		"backup":   backupPath,
		"restored": savePath,
	})
	return nil
}

// ValidateMigration validates that a migrated save file is correct.
func ValidateMigration(savePath string) error {
	version, err := DetectSaveVersion(savePath)
	if err != nil {
		return fmt.Errorf("failed to detect version: %w", err)
	}

	if version != SaveVersionV2 {
		return fmt.Errorf("expected v2.0, got %s", version)
	}

	// Try to load as v2
	data, err := os.ReadFile(savePath)
	if err != nil {
		return fmt.Errorf("failed to read save file: %w", err)
	}

	var saveFile SaveFileV2
	if err := json.Unmarshal(data, &saveFile); err != nil {
		return fmt.Errorf("failed to parse as v2: %w", err)
	}

	// Validate v2-specific fields exist
	if saveFile.GameState.MomentumConfig == nil {
		return fmt.Errorf("missing MomentumConfig")
	}
	if saveFile.GameState.NPCManager == nil {
		return fmt.Errorf("missing NPCManager")
	}
	if saveFile.GameState.GlobalFacts == nil {
		return fmt.Errorf("missing GlobalFacts")
	}
	if saveFile.GameState.ChatSessions == nil {
		return fmt.Errorf("missing ChatSessions")
	}

	// Validate checksum
	calculatedChecksum, err := ComputeChecksum(&saveFile)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if saveFile.Checksum != calculatedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s",
			calculatedChecksum, saveFile.Checksum)
	}

	logger.Info("Migration validation passed", map[string]interface{}{
		"path": savePath,
	})
	return nil
}

// Note: LoadV2 is defined in manager.go
