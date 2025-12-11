package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAudioFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected AudioFormat
	}{
		{"test.ogg", FormatOGG},
		{"test.OGG", FormatOGG}, // Case insensitive
		{"test.mp3", FormatMP3},
		{"test.MP3", FormatMP3},
		{"test.wav", FormatWAV},
		{"test.WAV", FormatWAV},
		{"test.flac", FormatUnknown},
		{"test.txt", FormatUnknown},
		{"test", FormatUnknown},
	}

	for _, tt := range tests {
		result := DetectAudioFormat(tt.filename)
		if result != tt.expected {
			t.Errorf("DetectAudioFormat(%q) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}
}

func TestValidateAudioFile_NonExistent(t *testing.T) {
	_, err := ValidateAudioFile("/tmp/nonexistent.ogg")
	if err == nil {
		t.Error("ValidateAudioFile should return error for non-existent file")
	}
}

func TestValidateAudioFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ogg")

	// Create a small test file (< 10MB)
	if err := os.WriteFile(testFile, []byte("mock audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	format, err := ValidateAudioFile(testFile)
	if err != nil {
		t.Errorf("ValidateAudioFile failed for valid file: %v", err)
	}
	if format != FormatOGG {
		t.Errorf("Expected format OGG, got %v", format)
	}
}

func TestValidateAudioFile_TooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.ogg")

	// Create a file > 10MB
	largeData := make([]byte, MaxAudioFileSize+1)
	if err := os.WriteFile(testFile, largeData, 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	_, err := ValidateAudioFile(testFile)
	if err == nil {
		t.Error("ValidateAudioFile should return error for files > 10MB")
	}
}

func TestValidateAudioFile_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.flac")

	if err := os.WriteFile(testFile, []byte("mock audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	format, err := ValidateAudioFile(testFile)
	if err == nil {
		t.Error("ValidateAudioFile should return error for unsupported format")
	}
	if format != FormatUnknown {
		t.Errorf("Expected format Unknown, got %v", format)
	}
}

func TestValidateAudioDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mix of valid and invalid files
	validFiles := []string{"audio1.ogg", "audio2.mp3", "audio3.wav"}
	for _, filename := range validFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("mock audio data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create an invalid file (unsupported format)
	invalidFile := filepath.Join(tmpDir, "invalid.txt")
	if err := os.WriteFile(invalidFile, []byte("not audio"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	// Create a too-large file
	largeFile := filepath.Join(tmpDir, "large.ogg")
	largeData := make([]byte, MaxAudioFileSize+1)
	if err := os.WriteFile(largeFile, largeData, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	validFilesList, err := ValidateAudioDirectory(tmpDir)
	if err != nil {
		t.Errorf("ValidateAudioDirectory failed: %v", err)
	}

	// Should return only the 3 valid files
	if len(validFilesList) != 3 {
		t.Errorf("Expected 3 valid files, got %d", len(validFilesList))
	}
}

func TestPreferOGGFiles(t *testing.T) {
	files := []string{
		"audio1.mp3",
		"audio2.ogg",
		"audio3.wav",
		"audio4.ogg",
		"audio5.mp3",
	}

	sorted := PreferOGGFiles(files)

	// First two should be OGG
	if !sliceContains(sorted[0:2], "audio2.ogg") {
		t.Error("OGG files should be prioritized")
	}
	if !sliceContains(sorted[0:2], "audio4.ogg") {
		t.Error("OGG files should be prioritized")
	}

	// Last three should be non-OGG
	if !sliceContains(sorted[2:], "audio1.mp3") {
		t.Error("Non-OGG files should come after OGG")
	}
}

func TestValidateAudioStructure_MissingDirectory(t *testing.T) {
	valid, warnings := ValidateAudioStructure("/tmp/nonexistent-audio")

	if valid {
		t.Error("ValidateAudioStructure should return false for non-existent directory")
	}

	if len(warnings) == 0 {
		t.Error("ValidateAudioStructure should return warnings")
	}
}

func TestValidateAudioStructure_ValidStructure(t *testing.T) {
	tmpDir := t.TempDir()
	bgmDir := filepath.Join(tmpDir, "bgm")
	sfxDir := filepath.Join(tmpDir, "sfx")

	// Create directory structure
	if err := os.MkdirAll(bgmDir, 0755); err != nil {
		t.Fatalf("Failed to create bgm directory: %v", err)
	}
	if err := os.MkdirAll(sfxDir, 0755); err != nil {
		t.Fatalf("Failed to create sfx directory: %v", err)
	}

	// Create 6 BGM files
	for i := 1; i <= 6; i++ {
		filename := filepath.Join(bgmDir, "bgm_"+string(rune('0'+i))+".ogg")
		if err := os.WriteFile(filename, []byte("mock audio"), 0644); err != nil {
			t.Fatalf("Failed to create BGM file: %v", err)
		}
	}

	// Create 10 SFX files
	for i := 1; i <= 10; i++ {
		filename := filepath.Join(sfxDir, "sfx_"+string(rune('0'+i))+".wav")
		if err := os.WriteFile(filename, []byte("mock audio"), 0644); err != nil {
			t.Fatalf("Failed to create SFX file: %v", err)
		}
	}

	valid, warnings := ValidateAudioStructure(tmpDir)

	if !valid {
		t.Errorf("ValidateAudioStructure should return true for valid structure, warnings: %v", warnings)
	}

	if len(warnings) > 0 {
		t.Errorf("ValidateAudioStructure should not return warnings for valid structure: %v", warnings)
	}
}

func TestValidateAudioStructure_InsufficientFiles(t *testing.T) {
	tmpDir := t.TempDir()
	bgmDir := filepath.Join(tmpDir, "bgm")
	sfxDir := filepath.Join(tmpDir, "sfx")

	// Create directory structure
	if err := os.MkdirAll(bgmDir, 0755); err != nil {
		t.Fatalf("Failed to create bgm directory: %v", err)
	}
	if err := os.MkdirAll(sfxDir, 0755); err != nil {
		t.Fatalf("Failed to create sfx directory: %v", err)
	}

	// Create only 3 BGM files (insufficient)
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(bgmDir, "bgm_"+string(rune('0'+i))+".ogg")
		if err := os.WriteFile(filename, []byte("mock audio"), 0644); err != nil {
			t.Fatalf("Failed to create BGM file: %v", err)
		}
	}

	// Create only 4 SFX files (insufficient)
	for i := 1; i <= 4; i++ {
		filename := filepath.Join(sfxDir, "sfx_"+string(rune('0'+i))+".wav")
		if err := os.WriteFile(filename, []byte("mock audio"), 0644); err != nil {
			t.Fatalf("Failed to create SFX file: %v", err)
		}
	}

	valid, warnings := ValidateAudioStructure(tmpDir)

	if valid {
		t.Error("ValidateAudioStructure should return false for insufficient files")
	}

	if len(warnings) == 0 {
		t.Error("ValidateAudioStructure should return warnings for insufficient files")
	}
}

// Helper function to check if slice contains string
func sliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
