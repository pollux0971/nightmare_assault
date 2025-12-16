package agents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// FilePromptLoader 是基於文件系統的 PromptLoader 實現
//
// 從指定的基礎目錄加載模板文件，並使用 Go 標準庫的 text/template 進行渲染。
//
// 使用方式：
//
//	loader := NewFilePromptLoader("internal/engine/prompts/templates")
//	tmpl, err := loader.LoadTemplate("narration.txt")
//	if err != nil {
//	    // 處理錯誤
//	}
//	rendered, err := loader.RenderTemplate(tmpl, map[string]any{
//	    "PlayerName": "Alice",
//	    "Location": "Dark Forest",
//	})
type FilePromptLoader struct {
	// baseDir 是模板文件的基礎目錄
	baseDir string
}

// NewFilePromptLoader 創建一個新的 FilePromptLoader 實例
//
// 參數：
//   - baseDir: 模板文件的基礎目錄路徑
//
// 返回：
//   - *FilePromptLoader: 新創建的 FilePromptLoader 實例
func NewFilePromptLoader(baseDir string) *FilePromptLoader {
	return &FilePromptLoader{
		baseDir: baseDir,
	}
}

// LoadTemplate 從文件系統加載模板內容
//
// 此方法從 baseDir 目錄下讀取指定名稱的模板文件。
// 文件路徑是相對於 baseDir 的，例如：
// - name = "narration.txt" → 讀取 {baseDir}/narration.txt
// - name = "agents/seed.txt" → 讀取 {baseDir}/agents/seed.txt
//
// 參數：
//   - name: 模板文件的名稱（相對於 baseDir）
//
// 返回：
//   - string: 模板文件的內容
//   - error: 如果文件不存在或讀取失敗，返回錯誤
func (l *FilePromptLoader) LoadTemplate(name string) (string, error) {
	// 構建完整的文件路徑
	path := filepath.Join(l.baseDir, name)

	// 讀取文件內容
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", name, err)
	}

	return string(content), nil
}

// RenderTemplate 使用 Go text/template 渲染模板
//
// 此方法使用 Go 標準庫的 text/template 包來渲染模板。
// 模板語法遵循 Go template 規範：
// - {{.FieldName}} - 訪問字段
// - {{range .Items}} ... {{end}} - 循環
// - {{if .Condition}} ... {{end}} - 條件判斷
//
// 參數：
//   - tmpl: 模板字符串（可以是 LoadTemplate 返回的內容）
//   - data: 用於渲染的數據（可以是 map[string]any 或結構體）
//
// 返回：
//   - string: 渲染後的結果
//   - error: 如果模板語法錯誤或渲染失敗，返回錯誤
func (l *FilePromptLoader) RenderTemplate(tmpl string, data any) (string, error) {
	// 解析模板
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 執行模板渲染
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}
