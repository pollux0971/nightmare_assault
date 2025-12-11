package save

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCurrentVersion(t *testing.T) {
	if CurrentVersion < 1 {
		t.Errorf("CurrentVersion must be at least 1, got %d", CurrentVersion)
	}
}

func TestMigrationHandlerInterface(t *testing.T) {
	// Verify MigrationHandler interface exists and can be implemented
	var _ MigrationHandler = &testMigration{}
}

type testMigration struct{}

func (m *testMigration) FromVersion() int {
	return 0
}

func (m *testMigration) ToVersion() int {
	return 1
}

func (m *testMigration) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
	return data, nil
}

func TestMigrationRegistry(t *testing.T) {
	// Reset registry for testing
	clearMigrations()

	// Register a test migration
	RegisterMigration(&testMigration{})

	// Verify it was registered
	migrations := GetMigrations()
	if len(migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(migrations))
	}

	// Cleanup
	clearMigrations()
}

func TestMigrateFromV0ToV1(t *testing.T) {
	// Old v0 format (hypothetical legacy format without version)
	v0Data := map[string]interface{}{
		"player_hp":  80,
		"player_san": 60,
		"chapter":    2,
	}

	// Register v0->v1 migration
	clearMigrations()
	RegisterMigration(&V0ToV1Migration{})

	// Migrate
	result, err := MigrateData(v0Data, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify migration was applied
	if v, ok := result["version"].(int); !ok || v != CurrentVersion {
		t.Errorf("Expected version %d after migration, got %v", CurrentVersion, result["version"])
	}

	// Cleanup
	clearMigrations()
}

func TestMigrateAlreadyCurrentVersion(t *testing.T) {
	data := map[string]interface{}{
		"version": CurrentVersion,
		"player":  map[string]interface{}{"hp": 100},
	}

	result, err := MigrateData(data, CurrentVersion, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Should be unchanged
	if result["version"] != CurrentVersion {
		t.Errorf("Version should remain %d", CurrentVersion)
	}
}

func TestMigrationLog(t *testing.T) {
	clearMigrations()
	clearMigrationLogs()

	RegisterMigration(&V0ToV1Migration{})

	v0Data := map[string]interface{}{
		"player_hp": 100,
		"chapter":   1,
	}

	_, err := MigrateData(v0Data, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check migration log
	logs := GetMigrationLogs()
	if len(logs) == 0 {
		t.Error("Expected migration to be logged")
	}

	// Verify log contains migration info
	found := false
	for _, log := range logs {
		if strings.Contains(log, "0") && strings.Contains(log, "1") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Migration log should mention version change, logs: %v", logs)
	}

	clearMigrations()
	clearMigrationLogs()
}

func TestMigratePreservesPlayerData(t *testing.T) {
	clearMigrations()
	RegisterMigration(&V0ToV1Migration{})

	v0Data := map[string]interface{}{
		"player_hp":  75,
		"player_san": 50,
		"chapter":    3,
		"location":   "basement",
	}

	result, err := MigrateData(v0Data, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify player data was preserved in new structure
	player, ok := result["player"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected player map in result")
	}

	if hp, ok := player["hp"].(int); !ok || hp != 75 {
		t.Errorf("Expected player HP 75, got %v", player["hp"])
	}

	if san, ok := player["san"].(int); !ok || san != 50 {
		t.Errorf("Expected player SAN 50, got %v", player["san"])
	}

	clearMigrations()
}

func TestMigratePreservesGameProgress(t *testing.T) {
	clearMigrations()
	RegisterMigration(&V0ToV1Migration{})

	v0Data := map[string]interface{}{
		"chapter":  5,
		"location": "rooftop",
	}

	result, err := MigrateData(v0Data, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify game state was preserved
	game, ok := result["game"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected game map in result")
	}

	if chapter, ok := game["current_chapter"].(int); !ok || chapter != 5 {
		t.Errorf("Expected chapter 5, got %v", game["current_chapter"])
	}

	clearMigrations()
}

func TestParseVersionFromJSON(t *testing.T) {
	data := []byte(`{"version": 1, "player": {"hp": 100}}`)

	version, err := ParseVersionFromJSON(data)
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}

func TestParseVersionFromJSONNoVersion(t *testing.T) {
	data := []byte(`{"player": {"hp": 100}}`)

	version, err := ParseVersionFromJSON(data)
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	// Should return 0 for legacy data without version
	if version != 0 {
		t.Errorf("Expected version 0 for legacy data, got %d", version)
	}
}

func TestV0ToV1MigrationDetails(t *testing.T) {
	migration := &V0ToV1Migration{}

	if migration.FromVersion() != 0 {
		t.Errorf("Expected FromVersion 0, got %d", migration.FromVersion())
	}

	if migration.ToVersion() != 1 {
		t.Errorf("Expected ToVersion 1, got %d", migration.ToVersion())
	}
}

func TestMigrateFullSaveData(t *testing.T) {
	clearMigrations()
	RegisterMigration(&V0ToV1Migration{})

	// Simulate a complete v0 save structure
	v0Save := map[string]interface{}{
		"player_hp":      80,
		"player_san":     70,
		"chapter":        2,
		"location":       "hospital",
		"saved_at":       time.Now().Format(time.RFC3339),
		"play_time":      3600,
		"difficulty":     "hard",
		"inventory":      []interface{}{"flashlight", "key"},
		"known_clues":    []interface{}{"blood_trail"},
		"triggered_rules": []interface{}{"rule_001"},
	}

	result, err := MigrateData(v0Save, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify new structure
	if result["version"] != CurrentVersion {
		t.Errorf("Version not updated to %d", CurrentVersion)
	}

	// Check metadata section
	if _, ok := result["metadata"]; !ok {
		t.Error("Expected metadata section in migrated data")
	}

	// Check player section
	if _, ok := result["player"]; !ok {
		t.Error("Expected player section in migrated data")
	}

	// Check game section
	if _, ok := result["game"]; !ok {
		t.Error("Expected game section in migrated data")
	}

	clearMigrations()
}

func TestMigrateAndSerialize(t *testing.T) {
	clearMigrations()
	RegisterMigration(&V0ToV1Migration{})

	v0Data := map[string]interface{}{
		"player_hp":  90,
		"player_san": 85,
		"chapter":    1,
	}

	result, err := MigrateData(v0Data, 0, CurrentVersion)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Should be able to serialize the result
	_, err = json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to serialize migrated data: %v", err)
	}

	clearMigrations()
}
