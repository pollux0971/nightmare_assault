package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Replacer 執行檔替換器
type Replacer struct {
	currentExePath string
	backupPath     string
}

// NewReplacer 創建新的替換器
func NewReplacer() (*Replacer, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("獲取當前執行檔路徑失敗: %w", err)
	}

	// 解析符號連結（如果有）
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return nil, fmt.Errorf("解析執行檔路徑失敗: %w", err)
	}

	backupPath := exePath + ".old"

	return &Replacer{
		currentExePath: exePath,
		backupPath:     backupPath,
	}, nil
}

// Replace 替換執行檔
func (r *Replacer) Replace(newExePath string) error {
	// 1. 驗證新執行檔存在
	if _, err := os.Stat(newExePath); err != nil {
		return fmt.Errorf("新執行檔不存在: %w", err)
	}

	// 2. 備份當前執行檔
	if err := r.backup(); err != nil {
		return fmt.Errorf("備份當前執行檔失敗: %w", err)
	}

	// 3. 替換執行檔
	if err := r.replaceExecutable(newExePath); err != nil {
		// 替換失敗，嘗試復原
		if rollbackErr := r.Rollback(); rollbackErr != nil {
			return fmt.Errorf("替換失敗且復原也失敗: %w (復原錯誤: %v)", err, rollbackErr)
		}
		return fmt.Errorf("替換執行檔失敗: %w", err)
	}

	// 4. 設置權限
	if err := r.setPermissions(); err != nil {
		// 權限設置失敗，嘗試復原
		if rollbackErr := r.Rollback(); rollbackErr != nil {
			return fmt.Errorf("設置權限失敗且復原也失敗: %w (復原錯誤: %v)", err, rollbackErr)
		}
		return fmt.Errorf("設置執行檔權限失敗: %w", err)
	}

	return nil
}

// backup 備份當前執行檔
func (r *Replacer) backup() error {
	// 如果已有舊備份，先刪除
	if _, err := os.Stat(r.backupPath); err == nil {
		if err := os.Remove(r.backupPath); err != nil {
			return fmt.Errorf("刪除舊備份失敗: %w", err)
		}
	}

	// 複製當前執行檔到備份位置
	if err := copyFile(r.currentExePath, r.backupPath); err != nil {
		return fmt.Errorf("複製執行檔失敗: %w", err)
	}

	return nil
}

// replaceExecutable 替換執行檔
func (r *Replacer) replaceExecutable(newExePath string) error {
	// Windows 特殊處理：無法直接覆蓋正在運行的執行檔
	if runtime.GOOS == "windows" {
		return r.replaceOnWindows(newExePath)
	}

	// Unix-like 系統可以直接覆蓋
	if err := copyFile(newExePath, r.currentExePath); err != nil {
		return fmt.Errorf("複製新執行檔失敗: %w", err)
	}

	return nil
}

// replaceOnWindows Windows平台的特殊處理
func (r *Replacer) replaceOnWindows(newExePath string) error {
	// Windows上無法替換正在運行的執行檔
	// 策略：重命名當前檔案，複製新檔案，告知用戶需要重啟

	// 生成臨時名稱
	tempOldPath := r.currentExePath + ".updating"

	// 重命名當前執行檔
	if err := os.Rename(r.currentExePath, tempOldPath); err != nil {
		return fmt.Errorf("重命名當前執行檔失敗: %w", err)
	}

	// 複製新執行檔到原位置
	if err := copyFile(newExePath, r.currentExePath); err != nil {
		// 失敗則復原重命名
		os.Rename(tempOldPath, r.currentExePath)
		return fmt.Errorf("複製新執行檔失敗: %w", err)
	}

	// 刪除臨時檔案
	os.Remove(tempOldPath)

	return nil
}

// setPermissions 設置執行檔權限
func (r *Replacer) setPermissions() error {
	// Windows 不需要特別設置執行權限
	if runtime.GOOS == "windows" {
		return nil
	}

	// Unix-like 系統設置為可執行
	if err := os.Chmod(r.currentExePath, 0755); err != nil {
		return fmt.Errorf("設置執行權限失敗: %w", err)
	}

	return nil
}

// Rollback 復原到備份的執行檔
func (r *Replacer) Rollback() error {
	// 檢查備份是否存在
	if _, err := os.Stat(r.backupPath); err != nil {
		return fmt.Errorf("備份檔案不存在: %w", err)
	}

	// 復原備份
	if err := copyFile(r.backupPath, r.currentExePath); err != nil {
		return fmt.Errorf("復原備份失敗: %w", err)
	}

	// 設置權限
	if err := r.setPermissions(); err != nil {
		return fmt.Errorf("設置權限失敗: %w", err)
	}

	return nil
}

// CleanupBackup 清理備份檔案
func (r *Replacer) CleanupBackup() error {
	if _, err := os.Stat(r.backupPath); err != nil {
		// 備份不存在，不算錯誤
		return nil
	}

	if err := os.Remove(r.backupPath); err != nil {
		return fmt.Errorf("刪除備份失敗: %w", err)
	}

	return nil
}

// copyFile 複製檔案
func copyFile(src, dst string) error {
	// 讀取源檔案
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("讀取源檔案失敗: %w", err)
	}

	// 寫入目標檔案
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("寫入目標檔案失敗: %w", err)
	}

	return nil
}

// NeedsRestart 檢查是否需要重啟應用程式
func (r *Replacer) NeedsRestart() bool {
	// Windows 總是需要重啟
	if runtime.GOOS == "windows" {
		return true
	}

	// 其他平台建議重啟以確保使用新版本
	return true
}
