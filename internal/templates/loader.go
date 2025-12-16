package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadError represents an error that occurred while loading a template
type LoadError struct {
	FilePath string
	LineNum  int
	ErrType  string
	Message  string
}

// Error implements the error interface
func (e *LoadError) Error() string {
	if e.LineNum > 0 {
		return fmt.Sprintf("%s:%d [%s] %s", e.FilePath, e.LineNum, e.ErrType, e.Message)
	}
	return fmt.Sprintf("%s [%s] %s", e.FilePath, e.ErrType, e.Message)
}

// TemplateLoader handles loading YAML templates from disk
type TemplateLoader struct {
	baseDir string
	errors  []*LoadError
}

// NewTemplateLoader creates a new template loader with the given base directory
func NewTemplateLoader(baseDir string) *TemplateLoader {
	return &TemplateLoader{
		baseDir: baseDir,
		errors:  make([]*LoadError, 0),
	}
}

// LoadYAMLFile loads a single YAML file and unmarshals it into the target
func (tl *TemplateLoader) LoadYAMLFile(filePath string, target interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		loadErr := &LoadError{
			FilePath: filePath,
			ErrType:  "READ_ERROR",
			Message:  err.Error(),
		}
		tl.errors = append(tl.errors, loadErr)
		return loadErr
	}

	err = yaml.Unmarshal(data, target)
	if err != nil {
		// Try to extract line number from yaml error
		loadErr := &LoadError{
			FilePath: filePath,
			ErrType:  "YAML_PARSE_ERROR",
			Message:  err.Error(),
		}
		tl.errors = append(tl.errors, loadErr)
		return loadErr
	}

	return nil
}

// LoadDirectory loads all YAML files from a directory
// Returns the number of successfully loaded files
func (tl *TemplateLoader) LoadDirectory(dir string, loader func(string) error) int {
	fullPath := filepath.Join(tl.baseDir, dir)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		loadErr := &LoadError{
			FilePath: fullPath,
			ErrType:  "DIR_READ_ERROR",
			Message:  err.Error(),
		}
		tl.errors = append(tl.errors, loadErr)
		return 0
	}

	successCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .yaml and .yml files
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(fullPath, entry.Name())
		err := loader(filePath)
		if err != nil {
			// Error already recorded by loader
			continue
		}
		successCount++
	}

	return successCount
}

// GetErrors returns all errors encountered during loading
func (tl *TemplateLoader) GetErrors() []*LoadError {
	return tl.errors
}

// ClearErrors clears all recorded errors
func (tl *TemplateLoader) ClearErrors() {
	tl.errors = make([]*LoadError, 0)
}

// HasErrors returns true if any errors were encountered
func (tl *TemplateLoader) HasErrors() bool {
	return len(tl.errors) > 0
}

// GetErrorSummary returns a formatted summary of all errors
func (tl *TemplateLoader) GetErrorSummary() string {
	if !tl.HasErrors() {
		return "No errors"
	}

	summary := fmt.Sprintf("Encountered %d error(s):\n", len(tl.errors))
	for i, err := range tl.errors {
		summary += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
	}
	return summary
}
