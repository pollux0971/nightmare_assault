package savefile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// ==========================================================================
// Story 8.5: Save Migration v1 → v2 - Tests
// ==========================================================================

func TestDetectSaveVersion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		saveData    map[string]interface{}
		wantVersion SaveVersion
		wantErr     bool
	}{
		{
			name: "v2 with game_version",
			saveData: map[string]interface{}{
				"meta": map[string]interface{}{
					"game_version": "v2.0.0",
				},
			},
			wantVersion: SaveVersionV2,
		},
		{
			name: "v1 with old game_version",
			saveData: map[string]interface{}{
				"meta": map[string]interface{}{
					"game_version": "v1.5.0",
				},
			},
			wantVersion: SaveVersionV1,
		},
		{
			name: "v2 with npc_manager field",
			saveData: map[string]interface{}{
				"meta": map[string]interface{}{},
				"game_state": map[string]interface{}{
					"npc_manager": map[string]interface{}{},
				},
			},
			wantVersion: SaveVersionV2,
		},
		{
			name: "v2 with global_facts field",
			saveData: map[string]interface{}{
				"meta": map[string]interface{}{},
				"game_state": map[string]interface{}{
					"global_facts": []interface{}{},
				},
			},
			wantVersion: SaveVersionV2,
		},
		{
			name: "v1 without v2 fields",
			saveData: map[string]interface{}{
				"meta": map[string]interface{}{},
				"game_state": map[string]interface{}{
					"game_id":      "test",
					"current_beat": 1,
					"hp":           100,
					"san":          100,
				},
			},
			wantVersion: SaveVersionV1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// AC1: Detect old save format
			savePath := filepath.Join(tmpDir, tt.name+".json")
			data, _ := json.MarshalIndent(tt.saveData, "", "  ")
			os.WriteFile(savePath, data, 0644)

			version, err := DetectSaveVersion(savePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectSaveVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if version != tt.wantVersion {
				t.Errorf("DetectSaveVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestBackupSaveFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test save file
	savePath := filepath.Join(tmpDir, "save_1.json")
	testData := []byte(`{"test": "data"}`)
	err := os.WriteFile(savePath, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test save file: %v", err)
	}

	// AC4: Backup old save file
	backupPath, err := BackupSaveFile(savePath)
	if err != nil {
		t.Fatalf("BackupSaveFile() error = %v", err)
	}

	// Check backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file does not exist: %s", backupPath)
	}

	// Check backup content matches original
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupData) != string(testData) {
		t.Errorf("Backup content = %s, want %s", backupData, testData)
	}

	// Check backup filename format
	baseName := filepath.Base(backupPath)
	if len(baseName) < 7 || baseName[:7] != "save_1_" {
		t.Errorf("Backup filename does not match expected format: %s", backupPath)
	}
}

func TestMigrateV1ToV2(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a v1 save file
	v1Save := SaveFileV2{
		Meta: SaveMetadata{
			SaveID:      1,
			SaveName:    "Test Save",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PlayTime:    3600,
			GameVersion: "v1.5.0",
		},
		Settings: GameSettings{
			Theme:      "horror",
			Model:      "gpt-4",
			Difficulty: "normal",
			Length:     "medium",
			AdultMode:  false,
			CreatedAt:  time.Now(),
		},
		GameState: &engine.GameStateV2{
			GameID:      "test-game",
			CurrentBeat: 5,
			HP:          80,
			SAN:         70,
			Inventory:   []string{"flashlight", "key"},
			// v1 does not have these fields (they will be nil/empty)
			MomentumConfig: nil,
			NPCManager:     nil,
			GlobalFacts:    nil,
			ChatSessions:   nil,
		},
		StoryBible: &StoryBibleData{
			GameID:     "test-game",
			CreatedAt:  time.Now().Format(time.RFC3339),
			Difficulty: "normal",
			TotalBeats: 20,
		},
		Checksum: "old-checksum",
	}

	savePath := filepath.Join(tmpDir, "save_1.json")
	data, _ := json.MarshalIndent(v1Save, "", "  ")
	os.WriteFile(savePath, data, 0644)

	// AC2: Automatically migrate to new format
	result, err := MigrateV1ToV2(savePath)
	if err != nil {
		t.Fatalf("MigrateV1ToV2() error = %v", err)
	}

	// Check migration result
	if !result.Success {
		t.Errorf("Migration failed: %v", result.Errors)
	}

	if result.FromVersion != SaveVersionV1 {
		t.Errorf("FromVersion = %v, want %v", result.FromVersion, SaveVersionV1)
	}

	if result.ToVersion != SaveVersionV2 {
		t.Errorf("ToVersion = %v, want %v", result.ToVersion, SaveVersionV2)
	}

	// AC4: Check backup was created
	if result.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}

	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Errorf("Backup file does not exist: %s", result.BackupPath)
	}

	// AC2: Check migrated fields
	expectedFields := []string{"momentum_config", "npc_manager", "global_facts", "chat_sessions", "rule_warnings", "meta.game_version", "checksum"}
	for _, field := range expectedFields {
		found := false
		for _, migrated := range result.MigratedFields {
			if migrated == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field %s to be migrated", field)
		}
	}

	// Load migrated file and verify v2 fields exist
	migratedSave, err := LoadV2(tmpDir, 1)
	if err != nil {
		t.Fatalf("Failed to load migrated save: %v", err)
	}

	if migratedSave.GameState.MomentumConfig == nil {
		t.Error("MomentumConfig should not be nil after migration")
	}

	if migratedSave.GameState.NPCManager == nil {
		t.Error("NPCManager should not be nil after migration")
	}

	if migratedSave.GameState.GlobalFacts == nil {
		t.Error("GlobalFacts should not be nil after migration")
	}

	if migratedSave.GameState.ChatSessions == nil {
		t.Error("ChatSessions should not be nil after migration")
	}

	if migratedSave.GameState.RuleWarnings == nil {
		t.Error("RuleWarnings should not be nil after migration")
	}

	if migratedSave.Meta.GameVersion != "v2.0.0" {
		t.Errorf("GameVersion = %s, want v2.0.0", migratedSave.Meta.GameVersion)
	}

	// Verify original data is preserved
	if migratedSave.GameState.HP != 80 {
		t.Errorf("HP = %d, want 80 (original data should be preserved)", migratedSave.GameState.HP)
	}

	if migratedSave.GameState.SAN != 70 {
		t.Errorf("SAN = %d, want 70 (original data should be preserved)", migratedSave.GameState.SAN)
	}
}

func TestAutoMigrate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		version     SaveVersion
		wantMigrate bool
	}{
		{
			name:        "v1 save auto-migrates",
			version:     SaveVersionV1,
			wantMigrate: true,
		},
		{
			name:        "v2 save no migration",
			version:     SaveVersionV2,
			wantMigrate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savePath := filepath.Join(tmpDir, tt.name+".json")

			// Create appropriate save file
			var saveData map[string]interface{}
			if tt.version == SaveVersionV1 {
				saveData = map[string]interface{}{
					"meta": map[string]interface{}{
						"game_version": "v1.5.0",
					},
					"game_state": map[string]interface{}{
						"game_id": "test",
						"hp":      100,
						"san":     100,
					},
					"settings":    map[string]interface{}{},
					"story_bible": map[string]interface{}{},
				}
			} else {
				saveData = map[string]interface{}{
					"meta": map[string]interface{}{
						"game_version": "v2.0.0",
					},
					"game_state": map[string]interface{}{
						"game_id":         "test",
						"hp":              100,
						"san":             100,
						"momentum_config": map[string]interface{}{},
						"npc_manager":     map[string]interface{}{},
						"global_facts":    []interface{}{},
						"chat_sessions":   []interface{}{},
					},
					"settings":    map[string]interface{}{},
					"story_bible": map[string]interface{}{},
				}
			}

			data, _ := json.MarshalIndent(saveData, "", "  ")
			os.WriteFile(savePath, data, 0644)

			// AC1: Auto-detect and migrate
			result, err := AutoMigrate(savePath)
			if err != nil {
				t.Fatalf("AutoMigrate() error = %v", err)
			}

			if tt.wantMigrate {
				if result.FromVersion != SaveVersionV1 {
					t.Errorf("FromVersion = %v, want %v", result.FromVersion, SaveVersionV1)
				}
				if len(result.MigratedFields) == 0 {
					t.Error("MigratedFields should not be empty for v1→v2 migration")
				}
			} else {
				if len(result.Warnings) == 0 {
					t.Error("Should have warning that no migration needed")
				}
			}
		})
	}
}

func TestLoadV2WithMigration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a v1 save file
	v1Save := map[string]interface{}{
		"meta": map[string]interface{}{
			"save_id":      1,
			"game_version": "v1.5.0",
			"created_at":   time.Now().Format(time.RFC3339),
			"updated_at":   time.Now().Format(time.RFC3339),
			"playtime":     3600,
		},
		"settings": map[string]interface{}{
			"theme":      "horror",
			"difficulty": "normal",
		},
		"game_state": map[string]interface{}{
			"game_id":      "test",
			"current_beat": 5,
			"hp":           80,
			"san":          70,
		},
		"story_bible": map[string]interface{}{},
	}

	savePath := filepath.Join(tmpDir, "save_1.json")
	data, _ := json.MarshalIndent(v1Save, "", "  ")
	os.WriteFile(savePath, data, 0644)

	// Load with automatic migration
	saveFile, migrationResult, err := LoadV2WithMigration(tmpDir, 1)
	if err != nil {
		t.Fatalf("LoadV2WithMigration() error = %v", err)
	}

	if saveFile == nil {
		t.Fatal("saveFile should not be nil")
	}

	if migrationResult == nil {
		t.Fatal("migrationResult should not be nil")
	}

	if !migrationResult.Success {
		t.Errorf("Migration should succeed: %v", migrationResult.Errors)
	}

	// Verify v2 fields exist
	if saveFile.GameState.MomentumConfig == nil {
		t.Error("MomentumConfig should not be nil")
	}
}

func TestRestoreFromBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create original and backup files
	savePath := filepath.Join(tmpDir, "save_1.json")
	backupPath := filepath.Join(tmpDir, "save_1_backup.json")

	originalData := []byte(`{"original": true}`)
	modifiedData := []byte(`{"modified": true}`)

	os.WriteFile(savePath, modifiedData, 0644)
	os.WriteFile(backupPath, originalData, 0644)

	// Restore from backup
	err := RestoreFromBackup(backupPath, savePath)
	if err != nil {
		t.Fatalf("RestoreFromBackup() error = %v", err)
	}

	// Verify content was restored
	restoredData, _ := os.ReadFile(savePath)
	if string(restoredData) != string(originalData) {
		t.Errorf("Restored content = %s, want %s", restoredData, originalData)
	}
}

func TestValidateMigration(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		save    *SaveFileV2
		wantErr bool
	}{
		{
			name: "valid v2 save",
			save: &SaveFileV2{
				Meta: SaveMetadata{
					SaveID:      1,
					GameVersion: "v2.0.0",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Settings: GameSettings{},
				GameState: &engine.GameStateV2{
					GameID:         "test",
					HP:             100,
					SAN:            100,
					MomentumConfig: &engine.MomentumConfig{},
					NPCManager:     &engine.NPCManagerState{},
					GlobalFacts:    []*engine.Fact{},
					ChatSessions:   []*engine.ChatSession{},
					RuleWarnings:   make(map[string]int),
				},
				StoryBible: &StoryBibleData{},
				Checksum:   "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savePath := filepath.Join(tmpDir, tt.name+".json")

			// Calculate checksum
			checksum, _ := ComputeChecksum(tt.save)
			tt.save.Checksum = checksum

			data, _ := json.MarshalIndent(tt.save, "", "  ")
			os.WriteFile(savePath, data, 0644)

			err := ValidateMigration(savePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMigration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrationErrorMessages(t *testing.T) {
	tmpDir := t.TempDir()

	// AC3: Clear error messages on migration failure
	tests := []struct {
		name          string
		setupFunc     func() string
		wantErrSubstr string
	}{
		{
			name: "non-existent file",
			setupFunc: func() string {
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantErrSubstr: "failed to read save file",
		},
		{
			name: "invalid JSON",
			setupFunc: func() string {
				path := filepath.Join(tmpDir, "invalid.json")
				os.WriteFile(path, []byte("invalid json{{{"), 0644)
				return path
			},
			wantErrSubstr: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savePath := tt.setupFunc()

			result, err := MigrateV1ToV2(savePath)
			if err == nil {
				t.Error("Expected error but got nil")
			}

			// AC3: Check error message is clear
			if result != nil && len(result.Errors) > 0 {
				errorFound := false
				for _, errMsg := range result.Errors {
					if len(errMsg) > 0 {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Error("Expected clear error messages in result")
				}
			}
		})
	}
}
