package audio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// AudioFormat represents supported audio file formats
type AudioFormat string

const (
	FormatOGG  AudioFormat = "ogg"  // Preferred for looping
	FormatMP3  AudioFormat = "mp3"
	FormatWAV  AudioFormat = "wav"
	FormatUnknown AudioFormat = "unknown"
)

// MaxAudioFileSize is the maximum allowed file size (10MB)
const MaxAudioFileSize int64 = 10 * 1024 * 1024

// ValidateAudioFile checks if a file is a valid audio file.
// Returns the audio format and any validation error.
func ValidateAudioFile(filePath string) (AudioFormat, error) {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return FormatUnknown, fmt.Errorf("file not found: %w", err)
	}

	// Check file size (< 10MB)
	if fileInfo.Size() > MaxAudioFileSize {
		return FormatUnknown, fmt.Errorf("file too large: %d bytes (max %d bytes)", fileInfo.Size(), MaxAudioFileSize)
	}

	// Detect format by extension
	format := DetectAudioFormat(filePath)
	if format == FormatUnknown {
		return FormatUnknown, fmt.Errorf("unsupported audio format: %s", filepath.Ext(filePath))
	}

	return format, nil
}

// DetectAudioFormat detects the audio format from file extension.
func DetectAudioFormat(filePath string) AudioFormat {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".ogg":
		return FormatOGG
	case ".mp3":
		return FormatMP3
	case ".wav":
		return FormatWAV
	default:
		return FormatUnknown
	}
}

// ValidateAudioDirectory validates all audio files in a directory.
// Returns a list of valid files and logs warnings for invalid files.
func ValidateAudioDirectory(dirPath string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dirPath, "*"))
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	validFiles := []string{}
	for _, file := range files {
		// Skip directories
		if fileInfo, err := os.Stat(file); err == nil && fileInfo.IsDir() {
			continue
		}

		format, err := ValidateAudioFile(file)
		if err != nil {
			log.Printf("[WARN] Skipping invalid audio file %s: %v\n", filepath.Base(file), err)
			continue
		}

		log.Printf("[INFO] Valid audio file: %s (format: %s)\n", filepath.Base(file), format)
		validFiles = append(validFiles, file)
	}

	return validFiles, nil
}

// PreferOGGFiles sorts audio files to prefer OGG format.
// OGG files are better for looping BGM.
func PreferOGGFiles(files []string) []string {
	oggFiles := []string{}
	otherFiles := []string{}

	for _, file := range files {
		if DetectAudioFormat(file) == FormatOGG {
			oggFiles = append(oggFiles, file)
		} else {
			otherFiles = append(otherFiles, file)
		}
	}

	// OGG files first, then others
	return append(oggFiles, otherFiles...)
}

// ValidateAudioStructure validates the complete audio directory structure.
// Returns true if the structure is valid.
func ValidateAudioStructure(audioDir string) (bool, []string) {
	warnings := []string{}

	// Check main audio directory
	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		warnings = append(warnings, "Audio directory does not exist: "+audioDir)
		return false, warnings
	}

	// Check BGM directory
	bgmDir := filepath.Join(audioDir, "bgm")
	if _, err := os.Stat(bgmDir); os.IsNotExist(err) {
		warnings = append(warnings, "BGM directory does not exist: "+bgmDir)
		return false, warnings
	}

	// Validate BGM files
	bgmFiles, err := ValidateAudioDirectory(bgmDir)
	if err != nil {
		warnings = append(warnings, "Failed to validate BGM directory: "+err.Error())
		return false, warnings
	}
	if len(bgmFiles) < 6 {
		warnings = append(warnings, fmt.Sprintf("Insufficient BGM files: found %d, expected 6+", len(bgmFiles)))
		return false, warnings
	}

	// Check SFX directory
	sfxDir := filepath.Join(audioDir, "sfx")
	if _, err := os.Stat(sfxDir); os.IsNotExist(err) {
		warnings = append(warnings, "SFX directory does not exist: "+sfxDir)
		return false, warnings
	}

	// Validate SFX files
	sfxFiles, err := ValidateAudioDirectory(sfxDir)
	if err != nil {
		warnings = append(warnings, "Failed to validate SFX directory: "+err.Error())
		return false, warnings
	}
	if len(sfxFiles) < 8 {
		warnings = append(warnings, fmt.Sprintf("Insufficient SFX files: found %d, expected 8+", len(sfxFiles)))
		return false, warnings
	}

	return true, warnings
}
