package audio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nightmare-assault/nightmare-assault/internal/engine"
)

// CustomBGMConfig represents custom BGM configuration
type CustomBGMConfig struct {
	MoodToFile map[engine.MoodType]string // mood -> custom filename or "default"
}

// NewCustomBGMConfig creates a new custom BGM configuration
func NewCustomBGMConfig() *CustomBGMConfig {
	return &CustomBGMConfig{
		MoodToFile: make(map[engine.MoodType]string),
	}
}

// CustomBGMManager handles custom BGM detection and management
type CustomBGMManager struct {
	customDir     string
	availableFiles []string
	config        *CustomBGMConfig
}

// NewCustomBGMManager creates a new custom BGM manager
func NewCustomBGMManager(customDir string) *CustomBGMManager {
	return &CustomBGMManager{
		customDir:      customDir,
		availableFiles: []string{},
		config:         NewCustomBGMConfig(),
	}
}

// ScanCustomDirectory scans for custom audio files
func (m *CustomBGMManager) ScanCustomDirectory() error {
	// Create directory if doesn't exist
	if err := os.MkdirAll(m.customDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom directory: %w", err)
	}

	// Scan for supported audio files
	files, err := os.ReadDir(m.customDir)
	if err != nil {
		return fmt.Errorf("failed to read custom directory: %w", err)
	}

	m.availableFiles = []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Check supported formats
		if ext == ".mp3" || ext == ".ogg" || ext == ".wav" {
			// Check file size (< 20MB)
			info, err := file.Info()
			if err != nil {
				log.Printf("[WARN] Failed to get file info for %s: %v", filename, err)
				continue
			}

			if info.Size() > 20*1024*1024 {
				log.Printf("[WARN] File %s exceeds 20MB limit, skipping", filename)
				continue
			}

			m.availableFiles = append(m.availableFiles, filename)
			log.Printf("[INFO] Found custom BGM: %s (%s, %.2f MB)", filename, ext, float64(info.Size())/(1024*1024))
		} else if ext == ".flac" || ext == ".m4a" {
			log.Printf("[WARN] Unsupported format: %s (only .mp3, .ogg, .wav supported)", filename)
		}
	}

	log.Printf("[INFO] Found %d custom BGM files", len(m.availableFiles))
	return nil
}

// GetAvailableFiles returns list of available custom BGM files
func (m *CustomBGMManager) GetAvailableFiles() []string {
	return m.availableFiles
}

// SetMoodBGM sets custom BGM for a specific mood
func (m *CustomBGMManager) SetMoodBGM(mood engine.MoodType, filename string) error {
	// Validate filename exists if not "default"
	if filename != "default" {
		found := false
		for _, f := range m.availableFiles {
			if f == filename {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("custom BGM file not found: %s", filename)
		}
	}

	m.config.MoodToFile[mood] = filename
	log.Printf("[INFO] Set %s mood BGM to: %s", mood, filename)
	return nil
}

// GetMoodBGM returns the configured BGM for a mood
func (m *CustomBGMManager) GetMoodBGM(mood engine.MoodType) (string, bool) {
	filename, exists := m.config.MoodToFile[mood]
	if !exists || filename == "default" {
		return "", false // Use default BGM
	}
	return filepath.Join(m.customDir, filename), true
}

// ResetToDefault resets a mood to use default BGM
func (m *CustomBGMManager) ResetToDefault(mood engine.MoodType) {
	m.config.MoodToFile[mood] = "default"
	log.Printf("[INFO] Reset %s mood to default BGM", mood)
}

// ResetAllToDefault resets all moods to default
func (m *CustomBGMManager) ResetAllToDefault() {
	for mood := range m.config.MoodToFile {
		m.config.MoodToFile[mood] = "default"
	}
	log.Println("[INFO] Reset all moods to default BGM")
}

// FormatBGMList formats BGM configuration as readable string
func (m *CustomBGMManager) FormatBGMList() string {
	var b strings.Builder

	b.WriteString("=== BGM 配置 ===\n\n")

	moods := []engine.MoodType{
		engine.MoodExploration,
		engine.MoodTension,
		engine.MoodSafe,
		engine.MoodHorror,
		engine.MoodMystery,
		engine.MoodEnding,
	}

	for _, mood := range moods {
		moodName := formatMoodName(mood)
		filename, isCustom := m.config.MoodToFile[mood]

		if !isCustom || filename == "default" {
			b.WriteString(fmt.Sprintf("%s: [預設]\n", moodName))
		} else {
			b.WriteString(fmt.Sprintf("%s: [自訂] %s\n", moodName, filename))
		}
	}

	b.WriteString(fmt.Sprintf("\n自訂音樂檔案數量: %d\n", len(m.availableFiles)))

	return b.String()
}

// formatMoodName converts mood type to Chinese name
func formatMoodName(mood engine.MoodType) string {
	switch mood {
	case engine.MoodExploration:
		return "探索"
	case engine.MoodTension:
		return "緊張"
	case engine.MoodSafe:
		return "安全"
	case engine.MoodHorror:
		return "恐怖"
	case engine.MoodMystery:
		return "解謎"
	case engine.MoodEnding:
		return "結局"
	default:
		return mood.String()
	}
}

// ValidateCustomAudioFile validates custom audio file quality
func ValidateCustomAudioFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".mp3" && ext != ".ogg" && ext != ".wav" {
		return fmt.Errorf("unsupported format: %s", ext)
	}

	// TODO: Add actual audio decoding validation
	// For now, just check if file can be opened
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	return nil
}
