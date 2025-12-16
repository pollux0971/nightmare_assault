package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Story 4.1 AC1: Successfully parse valid YAML files
func TestTemplateLoader_LoadYAMLFile_Success(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a valid YAML file
	validYAML := `
name: test_template
id: test_001
value: 42
`
	testFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(testFile, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load the file
	loader := NewTemplateLoader(tmpDir)
	var target map[string]interface{}
	err = loader.LoadYAMLFile(testFile, &target)

	if err != nil {
		t.Errorf("LoadYAMLFile should succeed, got error: %v", err)
	}

	if target["name"] != "test_template" {
		t.Errorf("Expected name 'test_template', got '%v'", target["name"])
	}

	if loader.HasErrors() {
		t.Error("Loader should not have errors after successful load")
	}
}

// Story 4.1 AC2: Handle YAML parse errors gracefully
func TestTemplateLoader_LoadYAMLFile_ParseError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid YAML file
	invalidYAML := `
name: test
invalid yaml syntax here: [unclosed bracket
value: 42
`
	testFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(testFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewTemplateLoader(tmpDir)
	var target map[string]interface{}
	err = loader.LoadYAMLFile(testFile, &target)

	// Should return error
	if err == nil {
		t.Error("LoadYAMLFile should return error for invalid YAML")
	}

	// Error should be a LoadError
	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Fatal("Error should be of type *LoadError")
	}

	if loadErr.ErrType != "YAML_PARSE_ERROR" {
		t.Errorf("Expected error type 'YAML_PARSE_ERROR', got '%s'", loadErr.ErrType)
	}

	if !strings.Contains(loadErr.FilePath, "invalid.yaml") {
		t.Errorf("Error should reference the file path")
	}

	// Loader should have recorded the error
	if !loader.HasErrors() {
		t.Error("Loader should have errors")
	}

	if len(loader.GetErrors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(loader.GetErrors()))
	}
}

// Story 4.1 AC2: Handle file not found errors
func TestTemplateLoader_LoadYAMLFile_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	loader := NewTemplateLoader(tmpDir)
	var target map[string]interface{}
	err := loader.LoadYAMLFile(filepath.Join(tmpDir, "nonexistent.yaml"), &target)

	if err == nil {
		t.Error("LoadYAMLFile should return error for missing file")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Fatal("Error should be of type *LoadError")
	}

	if loadErr.ErrType != "READ_ERROR" {
		t.Errorf("Expected error type 'READ_ERROR', got '%s'", loadErr.ErrType)
	}
}

// Story 4.1 AC1: Load all YAML files from a directory
func TestTemplateLoader_LoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates", "rules")
	os.MkdirAll(templatesDir, 0755)

	// Create multiple YAML files
	files := []struct {
		name    string
		content string
		valid   bool
	}{
		{"rule1.yaml", "id: rule_001\nname: Test Rule 1", true},
		{"rule2.yaml", "id: rule_002\nname: Test Rule 2", true},
		{"invalid.yaml", "invalid: yaml: [syntax", false},
		{"readme.txt", "This is not YAML", false}, // Should be skipped
	}

	for _, f := range files {
		err := os.WriteFile(filepath.Join(templatesDir, f.name), []byte(f.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", f.name, err)
		}
	}

	loader := NewTemplateLoader(tmpDir)
	loadedMaps := make([]map[string]interface{}, 0)

	successCount := loader.LoadDirectory("templates/rules", func(filePath string) error {
		var target map[string]interface{}
		err := loader.LoadYAMLFile(filePath, &target)
		if err == nil {
			loadedMaps = append(loadedMaps, target)
		}
		return err
	})

	// Should successfully load 2 files (rule1 and rule2)
	if successCount != 2 {
		t.Errorf("Expected 2 successful loads, got %d", successCount)
	}

	if len(loadedMaps) != 2 {
		t.Errorf("Expected 2 loaded maps, got %d", len(loadedMaps))
	}

	// Should have 1 error (invalid.yaml)
	if len(loader.GetErrors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(loader.GetErrors()))
	}
}

// Story 4.1 AC2: Continue loading after errors
func TestTemplateLoader_LoadDirectory_ContinueAfterErrors(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	os.MkdirAll(rulesDir, 0755)

	// Create files: valid, invalid, valid
	os.WriteFile(filepath.Join(rulesDir, "01_valid.yaml"), []byte("id: 001"), 0644)
	os.WriteFile(filepath.Join(rulesDir, "02_invalid.yaml"), []byte("bad: yaml: ["), 0644)
	os.WriteFile(filepath.Join(rulesDir, "03_valid.yaml"), []byte("id: 003"), 0644)

	loader := NewTemplateLoader(tmpDir)
	successCount := 0

	loader.LoadDirectory("rules", func(filePath string) error {
		var target map[string]interface{}
		err := loader.LoadYAMLFile(filePath, &target)
		if err == nil {
			successCount++
		}
		return err
	})

	// Should successfully load 2 files despite the error in the middle
	if successCount != 2 {
		t.Errorf("Expected 2 successful loads, got %d", successCount)
	}

	// Should have 1 error recorded
	if len(loader.GetErrors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(loader.GetErrors()))
	}
}

// Test error summary formatting
func TestTemplateLoader_GetErrorSummary(t *testing.T) {
	loader := NewTemplateLoader("/tmp")

	// No errors initially
	summary := loader.GetErrorSummary()
	if summary != "No errors" {
		t.Errorf("Expected 'No errors', got '%s'", summary)
	}

	// Add some errors
	loader.errors = append(loader.errors, &LoadError{
		FilePath: "test1.yaml",
		ErrType:  "PARSE_ERROR",
		Message:  "syntax error",
	})
	loader.errors = append(loader.errors, &LoadError{
		FilePath: "test2.yaml",
		ErrType:  "READ_ERROR",
		Message:  "file not found",
	})

	summary = loader.GetErrorSummary()

	if !strings.Contains(summary, "2 error(s)") {
		t.Error("Summary should mention error count")
	}

	if !strings.Contains(summary, "test1.yaml") {
		t.Error("Summary should include first file")
	}

	if !strings.Contains(summary, "test2.yaml") {
		t.Error("Summary should include second file")
	}
}

// Test ClearErrors
func TestTemplateLoader_ClearErrors(t *testing.T) {
	loader := NewTemplateLoader("/tmp")

	loader.errors = append(loader.errors, &LoadError{
		FilePath: "test.yaml",
		ErrType:  "ERROR",
		Message:  "test",
	})

	if !loader.HasErrors() {
		t.Fatal("Should have errors before clear")
	}

	loader.ClearErrors()

	if loader.HasErrors() {
		t.Error("Should not have errors after clear")
	}

	if len(loader.GetErrors()) != 0 {
		t.Errorf("Expected 0 errors after clear, got %d", len(loader.GetErrors()))
	}
}

// Test LoadError formatting
func TestLoadError_Error(t *testing.T) {
	// Without line number
	err1 := &LoadError{
		FilePath: "/path/to/file.yaml",
		ErrType:  "PARSE_ERROR",
		Message:  "syntax error",
	}

	msg := err1.Error()
	if !strings.Contains(msg, "/path/to/file.yaml") {
		t.Error("Error message should contain file path")
	}
	if !strings.Contains(msg, "PARSE_ERROR") {
		t.Error("Error message should contain error type")
	}
	if !strings.Contains(msg, "syntax error") {
		t.Error("Error message should contain message")
	}

	// With line number
	err2 := &LoadError{
		FilePath: "file.yaml",
		LineNum:  42,
		ErrType:  "SYNTAX",
		Message:  "unexpected token",
	}

	msg2 := err2.Error()
	if !strings.Contains(msg2, "file.yaml:42") {
		t.Error("Error message should contain file:line")
	}
}

// Test loading directory that doesn't exist
func TestTemplateLoader_LoadDirectory_DirNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	loader := NewTemplateLoader(tmpDir)
	successCount := loader.LoadDirectory("nonexistent", func(filePath string) error {
		return nil
	})

	if successCount != 0 {
		t.Errorf("Expected 0 successful loads for nonexistent dir, got %d", successCount)
	}

	if !loader.HasErrors() {
		t.Error("Should have error for nonexistent directory")
	}

	errors := loader.GetErrors()
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].ErrType != "DIR_READ_ERROR" {
		t.Errorf("Expected DIR_READ_ERROR, got %s", errors[0].ErrType)
	}
}

// Test that non-YAML files are skipped
func TestTemplateLoader_LoadDirectory_SkipsNonYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test")
	os.MkdirAll(testDir, 0755)

	// Create various files
	os.WriteFile(filepath.Join(testDir, "valid.yaml"), []byte("id: 1"), 0644)
	os.WriteFile(filepath.Join(testDir, "valid.yml"), []byte("id: 2"), 0644)
	os.WriteFile(filepath.Join(testDir, "readme.txt"), []byte("text file"), 0644)
	os.WriteFile(filepath.Join(testDir, "config.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(testDir, "script.go"), []byte("package main"), 0644)

	loader := NewTemplateLoader(tmpDir)
	count := 0

	loader.LoadDirectory("test", func(filePath string) error {
		var target map[string]interface{}
		err := loader.LoadYAMLFile(filePath, &target)
		if err == nil {
			count++
		}
		return err
	})

	// Should only load .yaml and .yml files
	if count != 2 {
		t.Errorf("Expected to load 2 YAML files, got %d", count)
	}
}
