package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/nightmare-assault/nightmare-assault/internal/update"
)

// UpdateCommand 實作更新檢查指令
type UpdateCommand struct {
	updateManager update.UpdateManagerInterface
}

// NewUpdateCommand 創建新的更新指令
// 接受 update.UpdateManagerInterface 以便依賴注入和測試
func NewUpdateCommand(updateManager update.UpdateManagerInterface) *UpdateCommand {
	return &UpdateCommand{
		updateManager: updateManager,
	}
}

// Name 返回指令名稱
func (c *UpdateCommand) Name() string {
	return "update"
}

// Help 返回指令幫助訊息
func (c *UpdateCommand) Help() string {
	return `檢查並安裝遊戲更新：
  /update          - 檢查是否有新版本可用
  /update check    - 檢查是否有新版本（不下載）
  /update install  - 下載並安裝新版本（需要先檢查）

更新說明：
  - 遊戲會自動備份當前版本
  - 更新完成後需要重新啟動遊戲
  - 可在設置中停用自動檢查`
}

// Aliases 返回指令別名
func (c *UpdateCommand) Aliases() []string {
	return []string{"升級", "版本"}
}

// Execute 執行更新指令
func (c *UpdateCommand) Execute(args []string) (string, error) {
	// 檢查更新管理器是否可用
	if c.updateManager == nil {
		return "⚠️  自動更新功能未啟用\n\n" +
			"請手動下載最新版本：\n" +
			"https://github.com/nightmare-assault/nightmare-assault/releases", nil
	}

	// 預設行為：檢查更新
	subcommand := "check"
	if len(args) > 0 {
		subcommand = strings.ToLower(args[0])
	}

	switch subcommand {
	case "check", "檢查":
		return c.checkForUpdates()
	case "install", "安裝":
		return c.installUpdate()
	default:
		return c.checkForUpdates()
	}
}

// checkForUpdates 檢查是否有新版本
func (c *UpdateCommand) checkForUpdates() (string, error) {
	var output strings.Builder

	output.WriteString("🔍 正在檢查更新...\n\n")

	// 執行更新檢查
	result, err := c.updateManager.CheckForUpdates()
	if err != nil {
		return "", fmt.Errorf("檢查更新失敗: %w", err)
	}

	// 記錄檢查時間
	if err := c.updateManager.RecordUpdateCheck(); err != nil {
		// 不阻塞，僅記錄錯誤
		output.WriteString(fmt.Sprintf("⚠️  無法記錄檢查時間: %v\n\n", err))
	}

	// 顯示結果
	switch result.Status {
	case update.UpdateStatusUpToDate:
		output.WriteString("✅ 您已使用最新版本！\n\n")
		output.WriteString(fmt.Sprintf("當前版本: %s\n", result.CurrentVersion))

	case update.UpdateStatusAvailable:
		output.WriteString("🎉 發現新版本！\n\n")
		output.WriteString(fmt.Sprintf("當前版本: %s\n", result.CurrentVersion))
		output.WriteString(fmt.Sprintf("最新版本: %s\n\n", result.NewVersion))

		// 提示使用者如何安裝
		output.WriteString("執行以下指令安裝更新：\n")
		output.WriteString("  /update install\n\n")
		output.WriteString("或訪問以下網址手動下載：\n")
		output.WriteString("  https://github.com/nightmare-assault/nightmare-assault/releases\n")

	case update.UpdateStatusFailed:
		output.WriteString("❌ 檢查更新失敗\n\n")
		if result.ErrorMessage != "" {
			output.WriteString(fmt.Sprintf("錯誤訊息: %s\n", result.ErrorMessage))
		}
	}

	return output.String(), nil
}

// installUpdate 下載並安裝更新
func (c *UpdateCommand) installUpdate() (string, error) {
	var output strings.Builder

	output.WriteString("📦 開始更新流程...\n\n")

	// 先檢查更新
	result, err := c.updateManager.CheckForUpdates()
	if err != nil {
		return "", fmt.Errorf("檢查更新失敗: %w", err)
	}

	// 如果沒有更新
	if result.Status != update.UpdateStatusAvailable {
		output.WriteString("✅ 您已使用最新版本！\n")
		output.WriteString(fmt.Sprintf("當前版本: %s\n", result.CurrentVersion))
		return output.String(), nil
	}

	output.WriteString(fmt.Sprintf("正在下載版本 %s...\n", result.NewVersion))

	// 下載更新
	downloadPath, err := c.updateManager.DownloadUpdate(result)
	if err != nil {
		return "", fmt.Errorf("下載更新失敗: %w", err)
	}

	output.WriteString("✅ 下載完成\n")
	output.WriteString("🔐 正在驗證檔案...\n")

	// 安裝更新
	output.WriteString("📥 正在安裝更新...\n")
	if err := c.updateManager.InstallUpdate(downloadPath); err != nil {
		return "", fmt.Errorf("安裝更新失敗: %w", err)
	}

	output.WriteString("\n✅ 更新完成！\n\n")

	// 檢查是否需要重啟
	if c.updateManager.NeedsRestart() {
		output.WriteString("⚠️  需要重新啟動遊戲以使用新版本\n\n")
		output.WriteString("請執行以下操作：\n")
		output.WriteString("  1. 輸入 /quit 退出遊戲\n")
		output.WriteString("  2. 重新啟動遊戲\n")
	} else {
		output.WriteString("🎉 更新已生效，無需重啟！\n")
	}

	return output.String(), nil
}

// FormatUpdateNotification 格式化更新通知訊息（用於主選單顯示）
func FormatUpdateNotification(result *update.UpdateResult) string {
	if result == nil || result.Status != update.UpdateStatusAvailable {
		return ""
	}

	var output strings.Builder

	output.WriteString("╔════════════════════════════════════════════════╗\n")
	output.WriteString("║          🎉 發現新版本可用！                  ║\n")
	output.WriteString("╠════════════════════════════════════════════════╣\n")
	output.WriteString(fmt.Sprintf("║  當前版本: %-32s║\n", result.CurrentVersion))
	output.WriteString(fmt.Sprintf("║  最新版本: %-32s║\n", result.NewVersion))
	output.WriteString("║                                                ║\n")
	output.WriteString("║  執行 /update 查看更新詳情                    ║\n")
	output.WriteString("║  執行 /update install 安裝更新                ║\n")
	output.WriteString("╚════════════════════════════════════════════════╝\n")

	return output.String()
}

// CheckUpdateResult 檢查更新結果的訊息類型
type CheckUpdateResult struct {
	Result *update.UpdateResult
	Time   time.Time
}
