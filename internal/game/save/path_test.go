package save

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetSavePath(t *testing.T) {
	// Test slot 1
	path := GetSavePath(1)

	if path == "" {
		t.Error("GetSavePath returned empty string")
	}

	// Should contain .nightmare/saves
	if filepath.Base(filepath.Dir(path)) != "saves" {
		t.Errorf("Expected path to be in saves directory, got %s", path)
	}

	// Should end with save_1.json
	expectedFilename := "save_1.json"
	if filepath.Base(path) != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filepath.Base(path))
	}
}

func TestGetSavePathMultipleSlots(t *testing.T) {
	tests := []struct {
		slot         int
		expectedName string
	}{
		{1, "save_1.json"},
		{2, "save_2.json"},
		{3, "save_3.json"},
	}

	for _, tt := range tests {
		path := GetSavePath(tt.slot)
		if filepath.Base(path) != tt.expectedName {
			t.Errorf("Slot %d: expected %s, got %s", tt.slot, tt.expectedName, filepath.Base(path))
		}
	}
}

func TestGetSaveDir(t *testing.T) {
	dir := GetSaveDir()

	if dir == "" {
		t.Error("GetSaveDir returned empty string")
	}

	// Should end with .nightmare/saves
	if filepath.Base(dir) != "saves" {
		t.Errorf("Expected directory name 'saves', got %s", filepath.Base(dir))
	}

	parent := filepath.Base(filepath.Dir(dir))
	if parent != ".nightmare" {
		t.Errorf("Expected parent directory '.nightmare', got %s", parent)
	}
}

func TestEnsureSaveDir(t *testing.T) {
	// Use temp directory for testing
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	err := EnsureSaveDirAt(saveDir)
	if err != nil {
		t.Fatalf("EnsureSaveDirAt failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(saveDir)
	if err != nil {
		t.Fatalf("Save directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected a directory to be created")
	}
}

func TestEnsureSaveDirIdempotent(t *testing.T) {
	// Calling EnsureSaveDirAt multiple times should not fail
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	// First call
	if err := EnsureSaveDirAt(saveDir); err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call
	if err := EnsureSaveDirAt(saveDir); err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
}

func TestCrossPlatformPath(t *testing.T) {
	path := GetSavePath(1)

	// Path should use the correct separator for the OS
	switch runtime.GOOS {
	case "windows":
		// Windows paths should not have forward slashes (except in Go's filepath representation)
		// Just verify it's not empty and ends with correct filename
		if filepath.Base(path) != "save_1.json" {
			t.Errorf("Windows path incorrect: %s", path)
		}
	default:
		// Unix paths should use forward slashes
		if filepath.Base(path) != "save_1.json" {
			t.Errorf("Unix path incorrect: %s", path)
		}
	}
}

func TestCheckSaveSize(t *testing.T) {
	// Test within limit
	warning := CheckSaveSize(500 * 1024) // 500 KB
	if warning != "" {
		t.Errorf("Expected no warning for 500KB, got: %s", warning)
	}

	// Test at limit
	warning = CheckSaveSize(1024 * 1024) // 1 MB
	if warning != "" {
		t.Errorf("Expected no warning at 1MB, got: %s", warning)
	}

	// Test over limit
	warning = CheckSaveSize(2 * 1024 * 1024) // 2 MB
	if warning == "" {
		t.Error("Expected warning for 2MB save file")
	}
}

func TestCheckSaveSizeZero(t *testing.T) {
	warning := CheckSaveSize(0)
	if warning != "" {
		t.Errorf("Expected no warning for 0 bytes, got: %s", warning)
	}
}

func TestMaxSaveSizeBytes(t *testing.T) {
	// Verify max size is 1MB
	if MaxSaveSizeBytes != 1024*1024 {
		t.Errorf("Expected MaxSaveSizeBytes to be 1MB, got %d", MaxSaveSizeBytes)
	}
}

func TestSaveSlotExists(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	// Ensure directory exists
	if err := EnsureSaveDirAt(saveDir); err != nil {
		t.Fatalf("Failed to create save dir: %v", err)
	}

	// Slot should not exist initially
	savePath := filepath.Join(saveDir, "save_1.json")
	if SaveSlotExistsAt(savePath) {
		t.Error("Expected slot to not exist initially")
	}

	// Create a save file
	if err := os.WriteFile(savePath, []byte(`{"version":1}`), 0600); err != nil {
		t.Fatalf("Failed to create test save file: %v", err)
	}

	// Slot should exist now
	if !SaveSlotExistsAt(savePath) {
		t.Error("Expected slot to exist after creating file")
	}
}

func TestGetAllSaveSlots(t *testing.T) {
	tmpDir := t.TempDir()
	saveDir := filepath.Join(tmpDir, ".nightmare", "saves")

	// Create save directory
	if err := EnsureSaveDirAt(saveDir); err != nil {
		t.Fatalf("Failed to create save dir: %v", err)
	}

	// Initially no saves
	slots := GetAllSaveSlotsAt(saveDir)
	if len(slots) != 0 {
		t.Errorf("Expected 0 slots initially, got %d", len(slots))
	}

	// Create save files for slots 1 and 3
	for _, slot := range []int{1, 3} {
		savePath := filepath.Join(saveDir, "save_"+string(rune('0'+slot))+".json")
		if err := os.WriteFile(savePath, []byte(`{"version":1}`), 0600); err != nil {
			t.Fatalf("Failed to create test save file: %v", err)
		}
	}

	slots = GetAllSaveSlotsAt(saveDir)
	if len(slots) != 2 {
		t.Errorf("Expected 2 slots, got %d", len(slots))
	}
}

func TestValidateSlotID(t *testing.T) {
	tests := []struct {
		slot    int
		isValid bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, true},
		{4, false},
		{-1, false},
	}

	for _, tt := range tests {
		err := ValidateSlotID(tt.slot)
		if tt.isValid && err != nil {
			t.Errorf("Slot %d should be valid, got error: %v", tt.slot, err)
		}
		if !tt.isValid && err == nil {
			t.Errorf("Slot %d should be invalid, got no error", tt.slot)
		}
	}
}
