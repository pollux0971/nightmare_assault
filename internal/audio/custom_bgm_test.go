package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

func TestNewCustomBGMManager(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test-custom-bgm")

	if manager == nil {
		t.Fatal("NewCustomBGMManager returned nil")
	}

	if manager.customDir != "/tmp/test-custom-bgm" {
		t.Errorf("Expected custom dir '/tmp/test-custom-bgm', got '%s'", manager.customDir)
	}

	if len(manager.availableFiles) != 0 {
		t.Errorf("Expected 0 available files initially, got %d", len(manager.availableFiles))
	}
}

func TestScanCustomDirectory_CreateDir(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-custom-bgm-create")
	defer os.RemoveAll(testDir)

	manager := NewCustomBGMManager(testDir)

	err := manager.ScanCustomDirectory()
	if err != nil {
		t.Fatalf("ScanCustomDirectory failed: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Custom directory was not created")
	}
}

func TestScanCustomDirectory_FindFiles(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-custom-bgm-scan")
	defer os.RemoveAll(testDir)

	// Create test directory and files
	os.MkdirAll(testDir, 0755)

	// Create test audio files
	testFiles := []string{
		"ambient.mp3",
		"tension.ogg",
		"safe.wav",
		"unsupported.flac", // Should be warned but not added
		"readme.txt",       // Should be ignored
	}

	for _, filename := range testFiles {
		path := filepath.Join(testDir, filename)
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		// Write some dummy data
		f.Write([]byte("dummy audio data"))
		f.Close()
	}

	manager := NewCustomBGMManager(testDir)
	err := manager.ScanCustomDirectory()
	if err != nil {
		t.Fatalf("ScanCustomDirectory failed: %v", err)
	}

	files := manager.GetAvailableFiles()
	expectedCount := 3 // mp3, ogg, wav

	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d: %v", expectedCount, len(files), files)
	}

	// Check specific files are found
	expectedFiles := map[string]bool{
		"ambient.mp3": false,
		"tension.ogg": false,
		"safe.wav":    false,
	}

	for _, file := range files {
		if _, exists := expectedFiles[file]; exists {
			expectedFiles[file] = true
		}
	}

	for filename, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file '%s' not found", filename)
		}
	}
}

func TestScanCustomDirectory_FileSizeLimit(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-custom-bgm-size")
	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)

	// Create a file larger than 20MB
	largePath := filepath.Join(testDir, "large.mp3")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Write 21MB of data
	data := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 21; i++ {
		f.Write(data)
	}
	f.Close()

	// Create a normal-sized file
	normalPath := filepath.Join(testDir, "normal.mp3")
	f, err = os.Create(normalPath)
	if err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}
	f.Write([]byte("small data"))
	f.Close()

	manager := NewCustomBGMManager(testDir)
	err = manager.ScanCustomDirectory()
	if err != nil {
		t.Fatalf("ScanCustomDirectory failed: %v", err)
	}

	files := manager.GetAvailableFiles()

	// Should only find the normal file
	if len(files) != 1 {
		t.Errorf("Expected 1 file (large file should be skipped), got %d", len(files))
	}

	if len(files) > 0 && files[0] != "normal.mp3" {
		t.Errorf("Expected 'normal.mp3', got '%s'", files[0])
	}
}

func TestSetMoodBGM_Valid(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-custom-bgm-set")
	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)

	// Create test file
	testFile := "ambient.mp3"
	f, _ := os.Create(filepath.Join(testDir, testFile))
	f.Write([]byte("test"))
	f.Close()

	manager := NewCustomBGMManager(testDir)
	manager.ScanCustomDirectory()

	// Set mood BGM
	err := manager.SetMoodBGM(engine.MoodExploration, testFile)
	if err != nil {
		t.Fatalf("SetMoodBGM failed: %v", err)
	}

	// Verify it was set
	path, isCustom := manager.GetMoodBGM(engine.MoodExploration)
	if !isCustom {
		t.Error("Expected custom BGM to be set")
	}

	expectedPath := filepath.Join(testDir, testFile)
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestSetMoodBGM_InvalidFile(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")
	manager.availableFiles = []string{"existing.mp3"}

	err := manager.SetMoodBGM(engine.MoodExploration, "nonexistent.mp3")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestSetMoodBGM_Default(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")

	err := manager.SetMoodBGM(engine.MoodExploration, "default")
	if err != nil {
		t.Fatalf("SetMoodBGM with 'default' failed: %v", err)
	}

	_, isCustom := manager.GetMoodBGM(engine.MoodExploration)
	if isCustom {
		t.Error("Expected default BGM, but got custom")
	}
}

func TestResetToDefault(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")
	manager.availableFiles = []string{"test.mp3"}

	// Set custom BGM
	manager.SetMoodBGM(engine.MoodExploration, "test.mp3")

	// Verify it's custom
	_, isCustom := manager.GetMoodBGM(engine.MoodExploration)
	if !isCustom {
		t.Error("Expected custom BGM before reset")
	}

	// Reset to default
	manager.ResetToDefault(engine.MoodExploration)

	// Verify it's default now
	_, isCustom = manager.GetMoodBGM(engine.MoodExploration)
	if isCustom {
		t.Error("Expected default BGM after reset")
	}
}

func TestResetAllToDefault(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")
	manager.availableFiles = []string{"test.mp3"}

	// Set multiple custom BGMs
	manager.SetMoodBGM(engine.MoodExploration, "test.mp3")
	manager.SetMoodBGM(engine.MoodTension, "test.mp3")
	manager.SetMoodBGM(engine.MoodHorror, "test.mp3")

	// Reset all
	manager.ResetAllToDefault()

	// Verify all are default
	moods := []engine.MoodType{engine.MoodExploration, engine.MoodTension, engine.MoodHorror}
	for _, mood := range moods {
		_, isCustom := manager.GetMoodBGM(mood)
		if isCustom {
			t.Errorf("Expected default BGM for %s after reset all", mood)
		}
	}
}

func TestFormatBGMList(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")
	manager.availableFiles = []string{"custom1.mp3", "custom2.ogg"}

	manager.SetMoodBGM(engine.MoodExploration, "custom1.mp3")
	manager.SetMoodBGM(engine.MoodTension, "default")

	result := manager.FormatBGMList()

	if result == "" {
		t.Error("FormatBGMList returned empty string")
	}

	// Check for expected content
	expectedStrings := []string{
		"BGM 配置",
		"探索",
		"緊張",
		"[自訂]",
		"[預設]",
		"自訂音樂檔案數量: 2",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("Expected FormatBGMList to contain '%s'", expected)
		}
	}
}

func TestFormatMoodName(t *testing.T) {
	tests := []struct {
		mood     engine.MoodType
		expected string
	}{
		{engine.MoodExploration, "探索"},
		{engine.MoodTension, "緊張"},
		{engine.MoodSafe, "安全"},
		{engine.MoodHorror, "恐怖"},
		{engine.MoodMystery, "解謎"},
		{engine.MoodEnding, "結局"},
	}

	for _, tt := range tests {
		t.Run(tt.mood.String(), func(t *testing.T) {
			result := formatMoodName(tt.mood)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateCustomAudioFile_FileNotExist(t *testing.T) {
	err := ValidateCustomAudioFile("/nonexistent/file.mp3")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestValidateCustomAudioFile_UnsupportedFormat(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-validate")
	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)

	testFile := filepath.Join(testDir, "test.flac")
	f, _ := os.Create(testFile)
	f.Write([]byte("test"))
	f.Close()

	err := ValidateCustomAudioFile(testFile)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestValidateCustomAudioFile_ValidFormats(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-validate-valid")
	defer os.RemoveAll(testDir)

	os.MkdirAll(testDir, 0755)

	formats := []string{".mp3", ".ogg", ".wav"}

	for _, format := range formats {
		testFile := filepath.Join(testDir, "test"+format)
		f, _ := os.Create(testFile)
		f.Write([]byte("test data"))
		f.Close()

		err := ValidateCustomAudioFile(testFile)
		if err != nil {
			t.Errorf("ValidateCustomAudioFile failed for %s: %v", format, err)
		}
	}
}

func TestGetMoodBGM_NotSet(t *testing.T) {
	manager := NewCustomBGMManager("/tmp/test")

	_, isCustom := manager.GetMoodBGM(engine.MoodExploration)
	if isCustom {
		t.Error("Expected default BGM for unset mood")
	}
}

func TestCustomBGMConfig_NewConfig(t *testing.T) {
	config := NewCustomBGMConfig()

	if config == nil {
		t.Fatal("NewCustomBGMConfig returned nil")
	}

	if config.MoodToFile == nil {
		t.Error("MoodToFile map not initialized")
	}

	if len(config.MoodToFile) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(config.MoodToFile))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
