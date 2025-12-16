package agents

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFilePromptLoaderCreation 測試 FilePromptLoader 創建
func TestFilePromptLoaderCreation(t *testing.T) {
	baseDir := "/tmp/test-prompts"
	loader := NewFilePromptLoader(baseDir)

	if loader == nil {
		t.Fatal("Expected non-nil FilePromptLoader")
	}
}

// TestFilePromptLoaderLoadTemplate_Success 測試成功加載模板
func TestFilePromptLoaderLoadTemplate_Success(t *testing.T) {
	// 創建臨時目錄
	tmpDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 創建測試模板文件
	templateContent := "Hello {{.Name}}! Welcome to {{.Game}}."
	templatePath := filepath.Join(tmpDir, "test-template.txt")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// 創建 loader
	loader := NewFilePromptLoader(tmpDir)

	// 加載模板
	content, err := loader.LoadTemplate("test-template.txt")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if content != templateContent {
		t.Errorf("Expected content '%s', got '%s'", templateContent, content)
	}
}

// TestFilePromptLoaderLoadTemplate_FileNotFound 測試文件不存在的情況
func TestFilePromptLoaderLoadTemplate_FileNotFound(t *testing.T) {
	// 創建臨時目錄
	tmpDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 創建 loader
	loader := NewFilePromptLoader(tmpDir)

	// 嘗試加載不存在的模板
	content, err := loader.LoadTemplate("nonexistent.txt")

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}
}

// TestFilePromptLoaderRenderTemplate_Success 測試成功渲染模板
func TestFilePromptLoaderRenderTemplate_Success(t *testing.T) {
	loader := NewFilePromptLoader("")

	template := "Hello {{.Name}}! Welcome to {{.Game}}."
	data := map[string]any{
		"Name": "Alice",
		"Game": "Nightmare Assault",
	}

	rendered, err := loader.RenderTemplate(template, data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "Hello Alice! Welcome to Nightmare Assault."
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

// TestFilePromptLoaderRenderTemplate_WithStruct 測試使用結構體渲染模板
func TestFilePromptLoaderRenderTemplate_WithStruct(t *testing.T) {
	loader := NewFilePromptLoader("")

	type Player struct {
		Name  string
		Level int
	}

	template := "Player {{.Name}} is at level {{.Level}}."
	data := Player{
		Name:  "Bob",
		Level: 42,
	}

	rendered, err := loader.RenderTemplate(template, data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "Player Bob is at level 42."
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

// TestFilePromptLoaderRenderTemplate_InvalidTemplate 測試無效模板語法
func TestFilePromptLoaderRenderTemplate_InvalidTemplate(t *testing.T) {
	loader := NewFilePromptLoader("")

	// 無效的模板語法
	template := "Hello {{.Name"
	data := map[string]any{"Name": "Alice"}

	rendered, err := loader.RenderTemplate(template, data)

	if err == nil {
		t.Error("Expected error for invalid template syntax, got nil")
	}

	if rendered != "" {
		t.Errorf("Expected empty result, got '%s'", rendered)
	}
}

// TestFilePromptLoaderRenderTemplate_MissingField 測試缺少字段的情況
func TestFilePromptLoaderRenderTemplate_MissingField(t *testing.T) {
	loader := NewFilePromptLoader("")

	template := "Hello {{.Name}}! Your age is {{.Age}}."
	data := map[string]any{
		"Name": "Alice",
		// Age 字段缺失
	}

	rendered, err := loader.RenderTemplate(template, data)

	if err != nil {
		t.Errorf("Expected no error (Go template default behavior), got %v", err)
	}

	// Go template 對於缺失的字段會輸出 <no value>
	expected := "Hello Alice! Your age is <no value>."
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

// TestFilePromptLoaderLoadAndRender 測試加載並渲染完整流程
func TestFilePromptLoaderLoadAndRender(t *testing.T) {
	// 創建臨時目錄
	tmpDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 創建測試模板文件
	templateContent := "Welcome {{.Player}} to {{.Location}}!"
	templatePath := filepath.Join(tmpDir, "welcome.txt")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// 創建 loader
	loader := NewFilePromptLoader(tmpDir)

	// 加載模板
	template, err := loader.LoadTemplate("welcome.txt")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// 渲染模板
	data := map[string]any{
		"Player":   "Hero",
		"Location": "Dark Forest",
	}

	rendered, err := loader.RenderTemplate(template, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Welcome Hero to Dark Forest!"
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

// TestFilePromptLoaderInterface 測試 FilePromptLoader 實作 PromptLoader 接口
func TestFilePromptLoaderInterface(t *testing.T) {
	loader := NewFilePromptLoader("/tmp")

	// 驗證實作了 PromptLoader 接口
	var _ PromptLoader = loader
}
