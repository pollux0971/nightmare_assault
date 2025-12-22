package commands

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/update"
)

func TestUpdateCommand_Name(t *testing.T) {
	cmd := NewUpdateCommand(nil)
	if cmd.Name() != "update" {
		t.Errorf("Expected name 'update', got '%s'", cmd.Name())
	}
}

func TestUpdateCommand_Help(t *testing.T) {
	cmd := NewUpdateCommand(nil)
	help := cmd.Help()
	if help == "" {
		t.Error("Help text should not be empty")
	}
	if !strings.Contains(help, "/update") {
		t.Error("Help text should contain /update command")
	}
}

func TestUpdateCommand_Aliases(t *testing.T) {
	cmd := NewUpdateCommand(nil)
	aliases := cmd.Aliases()
	if len(aliases) == 0 {
		t.Error("Expected at least one alias")
	}
}

func TestUpdateCommand_Execute_NoManager(t *testing.T) {
	cmd := NewUpdateCommand(nil)

	output, err := cmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Expected no error when manager is nil, got: %v", err)
	}

	if !strings.Contains(output, "自動更新功能未啟用") {
		t.Error("Expected message about auto-update being disabled")
	}
	if !strings.Contains(output, "github.com") {
		t.Error("Expected message to contain GitHub URL")
	}
}

// mockUpdateManager 實現 update.UpdateManagerInterface 用於測試
type mockUpdateManager struct {
	checkResult      *update.UpdateResult
	checkError       error
	downloadPath     string
	downloadError    error
	installError     error
	needsRestart     bool
	recordCheckError error
}

func (m *mockUpdateManager) CheckForUpdates() (*update.UpdateResult, error) {
	return m.checkResult, m.checkError
}

func (m *mockUpdateManager) DownloadUpdate(result *update.UpdateResult) (string, error) {
	return m.downloadPath, m.downloadError
}

func (m *mockUpdateManager) InstallUpdate(downloadedPath string) error {
	return m.installError
}

func (m *mockUpdateManager) NeedsRestart() bool {
	return m.needsRestart
}

func (m *mockUpdateManager) RecordUpdateCheck() error {
	return m.recordCheckError
}

// 測試更新檢查 - 已是最新版本
func TestUpdateCommand_CheckUpToDate(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: &update.UpdateResult{
			Status:         update.UpdateStatusUpToDate,
			CurrentVersion: "v1.0.0",
			NewVersion:     "v1.0.0",
		},
		checkError: nil,
	}

	cmd := NewUpdateCommand(mock)
	output, err := cmd.Execute([]string{"check"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "已使用最新版本") {
		t.Error("Expected message about being up to date")
	}

	if !strings.Contains(output, "v1.0.0") {
		t.Error("Expected current version in output")
	}
}

// 測試更新檢查 - 發現新版本
func TestUpdateCommand_CheckUpdateAvailable(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: &update.UpdateResult{
			Status:         update.UpdateStatusAvailable,
			CurrentVersion: "v1.0.0",
			NewVersion:     "v1.1.0",
		},
		checkError: nil,
	}

	cmd := NewUpdateCommand(mock)
	output, err := cmd.Execute([]string{"check"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "發現新版本") {
		t.Error("Expected message about new version available")
	}

	if !strings.Contains(output, "v1.1.0") {
		t.Error("Expected new version in output")
	}

	if !strings.Contains(output, "/update install") {
		t.Error("Expected install instruction in output")
	}
}

// 測試更新檢查 - 檢查失敗
func TestUpdateCommand_CheckFailed(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: nil,
		checkError:  errors.New("network error"),
	}

	cmd := NewUpdateCommand(mock)
	_, err := cmd.Execute([]string{"check"})

	if err == nil {
		t.Error("Expected error when check fails")
	}

	if !strings.Contains(err.Error(), "檢查更新失敗") {
		t.Errorf("Expected '檢查更新失敗' in error, got: %v", err)
	}
}

// 測試安裝更新 - 已是最新版本
func TestUpdateCommand_InstallUpToDate(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: &update.UpdateResult{
			Status:         update.UpdateStatusUpToDate,
			CurrentVersion: "v1.0.0",
		},
	}

	cmd := NewUpdateCommand(mock)
	output, err := cmd.Execute([]string{"install"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "已使用最新版本") {
		t.Error("Expected message about being up to date")
	}
}

// 測試安裝更新 - 成功安裝
func TestUpdateCommand_InstallSuccess(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: &update.UpdateResult{
			Status:         update.UpdateStatusAvailable,
			CurrentVersion: "v1.0.0",
			NewVersion:     "v1.1.0",
			DownloadURL:    "http://example.com/download",
		},
		downloadPath: "/tmp/update",
		needsRestart: true,
	}

	cmd := NewUpdateCommand(mock)
	output, err := cmd.Execute([]string{"install"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "更新完成") {
		t.Error("Expected success message")
	}

	if !strings.Contains(output, "重新啟動") {
		t.Error("Expected restart message when NeedsRestart is true")
	}
}

// 測試安裝更新 - 下載失敗
func TestUpdateCommand_InstallDownloadFailed(t *testing.T) {
	mock := &mockUpdateManager{
		checkResult: &update.UpdateResult{
			Status:         update.UpdateStatusAvailable,
			CurrentVersion: "v1.0.0",
			NewVersion:     "v1.1.0",
			DownloadURL:    "http://example.com/download",
		},
		downloadError: errors.New("download failed"),
	}

	cmd := NewUpdateCommand(mock)
	_, err := cmd.Execute([]string{"install"})

	if err == nil {
		t.Error("Expected error when download fails")
	}

	if !strings.Contains(err.Error(), "下載更新失敗") {
		t.Errorf("Expected '下載更新失敗' in error, got: %v", err)
	}
}

// 測試格式化更新通知
func TestFormatUpdateNotification(t *testing.T) {
	tests := []struct {
		name     string
		result   *update.UpdateResult
		wantEmpty bool
	}{
		{
			name:     "nil result",
			result:   nil,
			wantEmpty: true,
		},
		{
			name: "up to date",
			result: &update.UpdateResult{
				Status:         update.UpdateStatusUpToDate,
				CurrentVersion: "v1.0.0",
			},
			wantEmpty: true,
		},
		{
			name: "update available",
			result: &update.UpdateResult{
				Status:         update.UpdateStatusAvailable,
				CurrentVersion: "v1.0.0",
				NewVersion:     "v1.1.0",
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatUpdateNotification(tt.result)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("Expected empty output, got: %s", output)
				}
			} else {
				if output == "" {
					t.Error("Expected non-empty output")
				}
				if !strings.Contains(output, "發現新版本") {
					t.Error("Expected message to contain '發現新版本'")
				}
				if !strings.Contains(output, tt.result.NewVersion) {
					t.Errorf("Expected output to contain new version %s", tt.result.NewVersion)
				}
			}
		})
	}
}

// 測試 CheckUpdateResult 訊息類型
func TestCheckUpdateResult(t *testing.T) {
	result := &update.UpdateResult{
		Status:         update.UpdateStatusAvailable,
		CurrentVersion: "v1.0.0",
		NewVersion:     "v1.1.0",
	}

	msg := CheckUpdateResult{
		Result: result,
		Time:   time.Now(),
	}

	if msg.Result.Status != update.UpdateStatusAvailable {
		t.Errorf("Expected status Available, got %v", msg.Result.Status)
	}

	if msg.Time.IsZero() {
		t.Error("Expected non-zero time")
	}
}
