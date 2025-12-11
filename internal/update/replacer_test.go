package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewReplacer(t *testing.T) {
	replacer, err := NewReplacer()
	if err != nil {
		t.Fatalf("NewReplacer() failed: %v", err)
	}

	if replacer == nil {
		t.Fatal("Expected replacer to be created")
	}

	if replacer.currentExePath == "" {
		t.Error("Expected currentExePath to be set")
	}

	if replacer.backupPath == "" {
		t.Error("Expected backupPath to be set")
	}

	// 備份路徑應該是當前執行檔路徑 + .old
	expectedBackup := replacer.currentExePath + ".old"
	if replacer.backupPath != expectedBackup {
		t.Errorf("Expected backupPath to be '%s', got '%s'", expectedBackup, replacer.backupPath)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 創建源檔案
	srcPath := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("test content for copy")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	// 複製檔案
	dstPath := filepath.Join(tmpDir, "destination.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() failed: %v", err)
	}

	// 驗證目標檔案存在且內容正確
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(srcContent) {
		t.Errorf("Content mismatch: expected '%s', got '%s'", string(srcContent), string(dstContent))
	}

	// 驗證權限
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedMode := os.FileMode(0755)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected mode %v, got %v", expectedMode, info.Mode().Perm())
	}
}

func TestBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// 創建測試執行檔
	testExe := filepath.Join(tmpDir, "test.exe")
	testContent := []byte("test executable")
	if err := os.WriteFile(testExe, testContent, 0755); err != nil {
		t.Fatal(err)
	}

	// 創建replacer（手動設置路徑）
	replacer := &Replacer{
		currentExePath: testExe,
		backupPath:     testExe + ".old",
	}

	// 執行備份
	if err := replacer.backup(); err != nil {
		t.Errorf("backup() failed: %v", err)
	}

	// 驗證備份檔案存在
	if _, err := os.Stat(replacer.backupPath); err != nil {
		t.Errorf("Backup file does not exist: %v", err)
	}

	// 驗證備份內容
	backupContent, err := os.ReadFile(replacer.backupPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(backupContent) != string(testContent) {
		t.Errorf("Backup content mismatch")
	}
}

func TestRollback(t *testing.T) {
	tmpDir := t.TempDir()

	// 創建原始執行檔
	testExe := filepath.Join(tmpDir, "test.exe")
	originalContent := []byte("original executable")
	if err := os.WriteFile(testExe, originalContent, 0755); err != nil {
		t.Fatal(err)
	}

	// 創建replacer
	replacer := &Replacer{
		currentExePath: testExe,
		backupPath:     testExe + ".old",
	}

	// 創建備份
	if err := replacer.backup(); err != nil {
		t.Fatal(err)
	}

	// 修改當前執行檔（模擬更新）
	modifiedContent := []byte("modified executable")
	if err := os.WriteFile(testExe, modifiedContent, 0755); err != nil {
		t.Fatal(err)
	}

	// 執行復原
	if err := replacer.Rollback(); err != nil {
		t.Errorf("Rollback() failed: %v", err)
	}

	// 驗證內容已復原
	currentContent, err := os.ReadFile(testExe)
	if err != nil {
		t.Fatal(err)
	}

	if string(currentContent) != string(originalContent) {
		t.Errorf("Rollback failed: content not restored")
	}
}

func TestCleanupBackup(t *testing.T) {
	tmpDir := t.TempDir()

	testExe := filepath.Join(tmpDir, "test.exe")
	backupPath := testExe + ".old"

	// 創建備份檔案
	if err := os.WriteFile(backupPath, []byte("backup"), 0755); err != nil {
		t.Fatal(err)
	}

	replacer := &Replacer{
		currentExePath: testExe,
		backupPath:     backupPath,
	}

	// 清理備份
	if err := replacer.CleanupBackup(); err != nil {
		t.Errorf("CleanupBackup() failed: %v", err)
	}

	// 驗證備份已刪除
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should be deleted")
	}

	// 再次清理（應該不報錯）
	if err := replacer.CleanupBackup(); err != nil {
		t.Error("CleanupBackup() should not fail if backup doesn't exist")
	}
}

func TestNeedsRestart(t *testing.T) {
	replacer := &Replacer{
		currentExePath: "/tmp/test",
		backupPath:     "/tmp/test.old",
	}

	// 所有平台都建議重啟
	if !replacer.NeedsRestart() {
		t.Error("NeedsRestart() should return true")
	}
}

func TestReplaceExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// 創建當前執行檔
	currentExe := filepath.Join(tmpDir, "current.exe")
	currentContent := []byte("current version")
	if err := os.WriteFile(currentExe, currentContent, 0755); err != nil {
		t.Fatal(err)
	}

	// 創建新執行檔
	newExe := filepath.Join(tmpDir, "new.exe")
	newContent := []byte("new version")
	if err := os.WriteFile(newExe, newContent, 0755); err != nil {
		t.Fatal(err)
	}

	replacer := &Replacer{
		currentExePath: currentExe,
		backupPath:     currentExe + ".old",
	}

	// 執行替換
	if err := replacer.replaceExecutable(newExe); err != nil {
		t.Errorf("replaceExecutable() failed: %v", err)
	}

	// 驗證當前執行檔已更新
	updatedContent, err := os.ReadFile(currentExe)
	if err != nil {
		t.Fatal(err)
	}

	if string(updatedContent) != string(newContent) {
		t.Errorf("Expected content '%s', got '%s'", string(newContent), string(updatedContent))
	}
}

func TestReplace(t *testing.T) {
	tmpDir := t.TempDir()

	// 創建當前執行檔
	currentExe := filepath.Join(tmpDir, "nightmare")
	currentContent := []byte("v1.0.0")
	if err := os.WriteFile(currentExe, currentContent, 0755); err != nil {
		t.Fatal(err)
	}

	// 創建新執行檔
	newExe := filepath.Join(tmpDir, "nightmare-new")
	newContent := []byte("v1.1.0")
	if err := os.WriteFile(newExe, newContent, 0755); err != nil {
		t.Fatal(err)
	}

	replacer := &Replacer{
		currentExePath: currentExe,
		backupPath:     currentExe + ".old",
	}

	// 執行完整替換流程
	if err := replacer.Replace(newExe); err != nil {
		t.Errorf("Replace() failed: %v", err)
	}

	// 驗證備份存在
	if _, err := os.Stat(replacer.backupPath); err != nil {
		t.Error("Backup should exist after Replace()")
	}

	// 驗證當前執行檔已更新
	updatedContent, err := os.ReadFile(currentExe)
	if err != nil {
		t.Fatal(err)
	}

	if string(updatedContent) != string(newContent) {
		t.Error("Current executable should be updated")
	}

	// 驗證備份內容
	backupContent, err := os.ReadFile(replacer.backupPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(backupContent) != string(currentContent) {
		t.Error("Backup should contain original content")
	}
}
