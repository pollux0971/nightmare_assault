package update

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager 更新管理器，協調所有更新組件
type Manager struct {
	config     UpdateConfig
	checker    *Checker
	downloader *Downloader
	replacer   *Replacer

	// 回調函數
	onStatusChange   func(status UpdateStatus, message string)
	onProgress       func(downloaded, total int64)
	onUpdateComplete func()
}

// NewManager 創建新的更新管理器
func NewManager(config UpdateConfig) (*Manager, error) {
	// 設置默認緩存目錄
	if config.CacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("獲取用戶目錄失敗: %w", err)
		}
		config.CacheDir = filepath.Join(homeDir, ".nightmare", "updates")
	}

	// 創建緩存目錄
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("創建緩存目錄失敗: %w", err)
	}

	checker := NewChecker(config)
	downloader := NewDownloader(config.CacheDir)

	replacer, err := NewReplacer()
	if err != nil {
		return nil, fmt.Errorf("初始化replacer失敗: %w", err)
	}

	return &Manager{
		config:     config,
		checker:    checker,
		downloader: downloader,
		replacer:   replacer,
	}, nil
}

// SetOnStatusChange 設置狀態變更回調
func (m *Manager) SetOnStatusChange(fn func(status UpdateStatus, message string)) {
	m.onStatusChange = fn
}

// SetOnProgress 設置進度回調
func (m *Manager) SetOnProgress(fn func(downloaded, total int64)) {
	m.onProgress = fn
}

// SetOnUpdateComplete 設置更新完成回調
func (m *Manager) SetOnUpdateComplete(fn func()) {
	m.onUpdateComplete = fn
}

// CheckForUpdates 檢查更新
func (m *Manager) CheckForUpdates() (*UpdateResult, error) {
	m.notifyStatus(UpdateStatusChecking, "正在檢查更新...")

	result, err := m.checker.CheckForUpdates()
	if err != nil {
		m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("檢查更新失敗: %v", err))
		return nil, err
	}

	if result.Status == UpdateStatusAvailable {
		m.notifyStatus(UpdateStatusAvailable, fmt.Sprintf("發現新版本: %s", result.NewVersion))
	} else if result.Status == UpdateStatusUpToDate {
		m.notifyStatus(UpdateStatusUpToDate, "已是最新版本")
	}

	return result, nil
}

// DownloadUpdate 下載更新
func (m *Manager) DownloadUpdate(result *UpdateResult) (string, error) {
	if result.DownloadURL == "" {
		return "", fmt.Errorf("無下載URL")
	}

	m.notifyStatus(UpdateStatusDownloading, "正在下載更新...")

	// 下載新執行檔
	filePath, err := m.downloader.DownloadWithRetry(result.DownloadURL, 3, m.onProgress)
	if err != nil {
		m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("下載失敗: %v", err))
		return "", err
	}

	// 如果有checksum URL，驗證checksum
	if result.Checksum != "" {
		m.notifyStatus(UpdateStatusVerifying, "正在驗證下載...")

		// 如果checksum是URL，先下載checksums文件
		checksums, err := m.downloader.FetchChecksums(result.Checksum)
		if err != nil {
			m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("獲取checksum失敗: %v", err))
			return "", err
		}

		// 找到對應檔案的checksum
		filename := filepath.Base(filePath)
		expectedChecksum, ok := checksums[filename]
		if !ok {
			m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("找不到檔案 %s 的checksum", filename))
			return "", fmt.Errorf("找不到檔案checksum")
		}

		// 驗證checksum
		if err := m.downloader.VerifyChecksum(filePath, expectedChecksum); err != nil {
			m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("Checksum驗證失敗: %v", err))
			// 刪除損壞的檔案
			os.Remove(filePath)
			return "", err
		}

		m.notifyStatus(UpdateStatusVerifying, "驗證通過")
	}

	return filePath, nil
}

// InstallUpdate 安裝更新
func (m *Manager) InstallUpdate(downloadedPath string) error {
	m.notifyStatus(UpdateStatusInstalling, "正在安裝更新...")

	// 替換執行檔
	if err := m.replacer.Replace(downloadedPath); err != nil {
		m.notifyStatus(UpdateStatusFailed, fmt.Sprintf("安裝失敗: %v", err))
		return err
	}

	m.notifyStatus(UpdateStatusCompleted, "更新完成！")

	if m.onUpdateComplete != nil {
		m.onUpdateComplete()
	}

	return nil
}

// PerformUpdate 執行完整更新流程
func (m *Manager) PerformUpdate() error {
	// 1. 檢查更新
	result, err := m.CheckForUpdates()
	if err != nil {
		return err
	}

	// 如果沒有更新，直接返回
	if result.Status != UpdateStatusAvailable {
		return nil
	}

	// 2. 下載更新
	downloadedPath, err := m.DownloadUpdate(result)
	if err != nil {
		return err
	}

	// 3. 安裝更新
	if err := m.InstallUpdate(downloadedPath); err != nil {
		return err
	}

	return nil
}

// NeedsRestart 檢查是否需要重啟
func (m *Manager) NeedsRestart() bool {
	return m.replacer.NeedsRestart()
}

// CleanupOldBackups 清理舊備份
func (m *Manager) CleanupOldBackups() error {
	return m.replacer.CleanupBackup()
}

// notifyStatus 通知狀態變更
func (m *Manager) notifyStatus(status UpdateStatus, message string) {
	if m.onStatusChange != nil {
		m.onStatusChange(status, message)
	}
}

// GetCurrentVersion 獲取當前版本
func (m *Manager) GetCurrentVersion() string {
	return m.config.CurrentVersion
}

// ShouldCheckForUpdates 檢查是否應該檢查更新（基於時間間隔）
func (m *Manager) ShouldCheckForUpdates() bool {
	// 檢查上次檢查時間
	lastCheckFile := filepath.Join(m.config.CacheDir, ".last_check")

	info, err := os.Stat(lastCheckFile)
	if err != nil {
		// 檔案不存在，應該檢查
		return true
	}

	// 檢查時間間隔
	timeSinceLastCheck := time.Since(info.ModTime())
	return timeSinceLastCheck >= m.config.CheckInterval
}

// RecordUpdateCheck 記錄檢查時間
func (m *Manager) RecordUpdateCheck() error {
	lastCheckFile := filepath.Join(m.config.CacheDir, ".last_check")
	return os.WriteFile(lastCheckFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}
