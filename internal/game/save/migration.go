package save

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MigrationHandler defines the interface for version migrations.
type MigrationHandler interface {
	FromVersion() int
	ToVersion() int
	Migrate(data map[string]interface{}) (map[string]interface{}, error)
}

var (
	migrations    []MigrationHandler
	migrationsMu  sync.Mutex
	migrationLogs []string
	logsMu        sync.Mutex
)

// RegisterMigration registers a migration handler.
func RegisterMigration(m MigrationHandler) {
	migrationsMu.Lock()
	defer migrationsMu.Unlock()
	migrations = append(migrations, m)
}

// GetMigrations returns all registered migrations.
func GetMigrations() []MigrationHandler {
	migrationsMu.Lock()
	defer migrationsMu.Unlock()
	result := make([]MigrationHandler, len(migrations))
	copy(result, migrations)
	return result
}

// clearMigrations clears all registered migrations (for testing).
func clearMigrations() {
	migrationsMu.Lock()
	defer migrationsMu.Unlock()
	migrations = nil
}

// GetMigrationLogs returns the migration logs.
func GetMigrationLogs() []string {
	logsMu.Lock()
	defer logsMu.Unlock()
	result := make([]string, len(migrationLogs))
	copy(result, migrationLogs)
	return result
}

// clearMigrationLogs clears all migration logs (for testing).
func clearMigrationLogs() {
	logsMu.Lock()
	defer logsMu.Unlock()
	migrationLogs = nil
}

// logMigration adds a log entry for a migration.
func logMigration(fromVersion, toVersion int, message string) {
	logsMu.Lock()
	defer logsMu.Unlock()
	entry := fmt.Sprintf("[%s] Migration v%d -> v%d: %s",
		time.Now().Format(time.RFC3339), fromVersion, toVersion, message)
	migrationLogs = append(migrationLogs, entry)
}

// MigrateData migrates data from one version to another.
func MigrateData(data map[string]interface{}, fromVersion, toVersion int) (map[string]interface{}, error) {
	if fromVersion == toVersion {
		return data, nil
	}

	if fromVersion > toVersion {
		return nil, fmt.Errorf("cannot migrate backwards from v%d to v%d", fromVersion, toVersion)
	}

	currentData := data
	currentVersion := fromVersion

	for currentVersion < toVersion {
		migration := findMigration(currentVersion)
		if migration == nil {
			return nil, fmt.Errorf("no migration found for version %d", currentVersion)
		}

		var err error
		currentData, err = migration.Migrate(currentData)
		if err != nil {
			logMigration(migration.FromVersion(), migration.ToVersion(), fmt.Sprintf("failed: %v", err))
			return nil, fmt.Errorf("migration from v%d to v%d failed: %w",
				migration.FromVersion(), migration.ToVersion(), err)
		}

		logMigration(migration.FromVersion(), migration.ToVersion(), "success")
		currentVersion = migration.ToVersion()
	}

	return currentData, nil
}

// findMigration finds a migration handler for the given source version.
func findMigration(fromVersion int) MigrationHandler {
	migrationsMu.Lock()
	defer migrationsMu.Unlock()
	for _, m := range migrations {
		if m.FromVersion() == fromVersion {
			return m
		}
	}
	return nil
}

// ParseVersionFromJSON extracts the version from JSON data.
func ParseVersionFromJSON(data []byte) (int, error) {
	var partial struct {
		Version int `json:"version"`
	}

	if err := json.Unmarshal(data, &partial); err != nil {
		return 0, fmt.Errorf("failed to parse version: %w", err)
	}

	// Version 0 is returned for legacy data without version field
	return partial.Version, nil
}

// V0ToV1Migration migrates from legacy format (v0) to v1.
type V0ToV1Migration struct{}

// FromVersion returns the source version.
func (m *V0ToV1Migration) FromVersion() int {
	return 0
}

// ToVersion returns the target version.
func (m *V0ToV1Migration) ToVersion() int {
	return 1
}

// Migrate performs the v0 to v1 migration.
func (m *V0ToV1Migration) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set version
	result["version"] = CurrentVersion

	// Migrate metadata
	metadata := make(map[string]interface{})
	if savedAt, ok := data["saved_at"]; ok {
		metadata["saved_at"] = savedAt
	} else {
		metadata["saved_at"] = time.Now().Format(time.RFC3339)
	}
	if playTime, ok := data["play_time"]; ok {
		metadata["play_time_seconds"] = playTime
	} else {
		metadata["play_time_seconds"] = 0
	}
	if difficulty, ok := data["difficulty"]; ok {
		metadata["difficulty"] = difficulty
	} else {
		metadata["difficulty"] = "normal"
	}
	if storyLength, ok := data["story_length"]; ok {
		metadata["story_length"] = storyLength
	} else {
		metadata["story_length"] = "medium"
	}
	result["metadata"] = metadata

	// Migrate player state
	player := make(map[string]interface{})
	if hp, ok := data["player_hp"]; ok {
		player["hp"] = hp
	} else {
		player["hp"] = 100
	}
	if san, ok := data["player_san"]; ok {
		player["san"] = san
	} else {
		player["san"] = 100
	}
	if location, ok := data["location"]; ok {
		player["location"] = location
	} else {
		player["location"] = ""
	}
	if inventory, ok := data["inventory"]; ok {
		player["inventory"] = inventory
	} else {
		player["inventory"] = []interface{}{}
	}
	if knownClues, ok := data["known_clues"]; ok {
		player["known_clues"] = knownClues
	} else {
		player["known_clues"] = []interface{}{}
	}
	result["player"] = player

	// Migrate game state
	game := make(map[string]interface{})
	if chapter, ok := data["chapter"]; ok {
		game["current_chapter"] = chapter
	} else {
		game["current_chapter"] = 1
	}
	if progress, ok := data["chapter_progress"]; ok {
		game["chapter_progress"] = progress
	} else {
		game["chapter_progress"] = 0.0
	}
	if triggeredRules, ok := data["triggered_rules"]; ok {
		game["triggered_rules"] = triggeredRules
	} else {
		game["triggered_rules"] = []interface{}{}
	}
	if discoveredRules, ok := data["discovered_rules"]; ok {
		game["discovered_rules"] = discoveredRules
	} else {
		game["discovered_rules"] = []interface{}{}
	}
	result["game"] = game

	// Migrate teammates (empty array for v0)
	if teammates, ok := data["teammates"]; ok {
		result["teammates"] = teammates
	} else {
		result["teammates"] = []interface{}{}
	}

	// Migrate context (empty for v0)
	context := make(map[string]interface{})
	if recentSummary, ok := data["recent_summary"]; ok {
		context["recent_summary"] = recentSummary
	} else {
		context["recent_summary"] = ""
	}
	if currentScene, ok := data["current_scene"]; ok {
		context["current_scene"] = currentScene
	} else {
		context["current_scene"] = ""
	}
	if gameBible, ok := data["game_bible_snapshot"]; ok {
		context["game_bible_snapshot"] = gameBible
	} else {
		context["game_bible_snapshot"] = ""
	}
	result["context"] = context

	// Checksum will be computed when saving
	result["checksum"] = ""

	return result, nil
}

// init registers the default migrations.
func init() {
	// Note: Migrations are registered in tests for isolation
	// In production, you may want to register default migrations here
}
