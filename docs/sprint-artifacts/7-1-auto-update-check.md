# Story 7.1: 自動更新檢查

Status: done

## Story

As a 玩家,
I want 遊戲自動檢查更新,
so that 我能及時獲得新功能和修復.

## Acceptance Criteria

### AC1: 啟動時背景檢查更新

**Given** 遊戲啟動
**When** 連線可用
**Then** 背景檢查 GitHub Releases 最新版本
**And** 若有新版本，主選單顯示提示
**And** 檢查不阻塞啟動流程（異步執行）

### AC2: 互動式更新流程

**Given** 有新版本可用
**When** 玩家選擇更新
**Then** 下載新版本執行檔
**And** 顯示下載進度條
**And** 驗證 checksum (NFR02)
**And** 替換當前執行檔
**And** 提示重啟遊戲

### AC3: 錯誤處理與降級

**Given** 更新過程
**When** 發生錯誤（網路失敗、checksum 不符、檔案權限問題）
**Then** 顯示友善錯誤訊息
**And** 不影響當前版本運行
**And** 保留原執行檔
**And** 記錄錯誤日誌供除錯

### AC4: 命令行更新模式

**Given** 玩家執行 `nightmare --update`
**When** 命令行模式
**Then** 直接執行更新流程
**And** 顯示文字進度
**And** 更新成功後自動重啟
**And** 更新失敗時返回非零退出碼

### AC5: 版本比對邏輯

**Given** 檢查更新時
**When** 比對版本號
**Then** 使用語義化版本比對（Semantic Versioning）
**And** 跳過 pre-release 版本（除非用戶選擇接收 beta）
**And** 顯示版本差異（當前版本 vs. 最新版本）

## Tasks / Subtasks

- [x] 實作 GitHub API 整合 (AC: #1, #5)
  - [x] 建立 `internal/update/checker.go`
  - [x] 實作 `CheckForUpdates()` 函數
  - [x] 解析 GitHub Releases API 回應
  - [x] 實作版本比對邏輯（SemVer）
  - [x] 處理 API rate limiting

- [x] 實作下載與驗證機制 (AC: #2, #3)
  - [x] 建立 `internal/update/downloader.go`
  - [x] 實作檔案下載器（支援重試機制）
  - [x] 實作 checksum 驗證（SHA256）
  - [x] 實作進度回調機制
  - [x] 處理下載失敗重試邏輯（最多 3 次）

- [x] 實作跨平台執行檔替換 (AC: #2, #3)
  - [x] 建立 `internal/update/replacer.go`
  - [x] 處理 Windows/macOS/Linux 執行檔替換
  - [x] 備份當前執行檔
  - [x] 實作原子替換（避免中途失敗）
  - [x] 處理檔案權限問題

- [x] 整合主選單更新提示 (AC: #1)
  - [x] 修改 `internal/tui/views/main_menu.go`
  - [x] 添加更新提示橫幅
  - [x] 添加 UpdateAvailableMsg 消息處理
  - [x] 異步檢查不阻塞 UI（在 app.go 中實作）

- [x] 實作命令行模式 (AC: #4)
  - [x] 修改 `cmd/nightmare/main.go` 添加 `--update` flag
  - [x] 實作純文字進度顯示
  - [x] 處理更新成功/失敗退出碼
  - [x] 添加 `--version` 顯示當前版本（已存在）

- [x] 錯誤處理與日誌 (AC: #3)
  - [x] 建立友善錯誤訊息模板
  - [x] 記錄更新流程狀態變更
  - [x] 實作降級機制（還原備份）
  - [x] 添加單元測試（16 個測試全部通過）

## Dev Notes

### 架構模式與約束

**架構模式:**
- **GitHub Releases 作為更新源**: 使用 GitHub API `/repos/owner/repo/releases/latest`
- **安全性優先**: 所有下載必須驗證 checksum，防止中間人攻擊
- **非阻塞設計**: 更新檢查異步執行，不影響遊戲啟動時間
- **原子操作**: 執行檔替換必須是原子的，避免部分替換導致損壞

**技術約束:**
- 啟動時更新檢查必須在背景執行，不能阻塞主選單顯示
- 更新檢查失敗不應影響遊戲正常運行
- 支援跨平台執行檔命名規範（nightmare-{os}-{arch}）
- checksum 使用 SHA256，並從 GitHub Release assets 獲取

**NFR 滿足:**
- NFR02 (安全需求): 使用 checksum 驗證 ✓
- NFR01 (性能需求): 啟動時異步檢查，不影響啟動時間 ✓
- NFR03 (可用性需求): 網路中斷時優雅降級 ✓

**依賴項:**
- Epic 1 基礎設施（執行檔、配置系統）
- GitHub API 訪問權限（無需 token 的 public API）
- 跨平台檔案系統操作

**風險與緩解:**
- **風險**: GitHub API rate limiting
  - **緩解**: 緩存檢查結果（24 小時），避免頻繁請求
- **風險**: 執行檔替換權限問題
  - **緩解**: 提示用戶使用管理員權限或手動替換
- **風險**: 下載中斷
  - **緩解**: 支援斷點續傳，最多重試 3 次

### Implementation Details

**檔案結構:**
```
internal/update/
  ├── checker.go       # 版本檢查邏輯
  ├── downloader.go    # 檔案下載器
  ├── replacer.go      # 執行檔替換
  └── types.go         # 更新相關資料結構
```

**GitHub API 使用:**
```
GET https://api.github.com/repos/{owner}/{repo}/releases/latest
回應包含:
  - tag_name: v1.2.3
  - assets: [{name, browser_download_url}]
  - body: release notes
```

**checksum 驗證流程:**
1. 從 GitHub Release 下載 `checksums.txt`
2. 找到對應平台的 checksum
3. 計算下載檔案的 SHA256
4. 比對是否一致

**執行檔替換流程 (跨平台):**
```go
// 1. 備份當前執行檔
currentExe, _ := os.Executable()
backup := currentExe + ".bak"
os.Rename(currentExe, backup)

// 2. 替換為新版本
os.Rename(downloadedFile, currentExe)

// 3. 設定執行權限（Unix）
os.Chmod(currentExe, 0755)

// 4. 若失敗，還原備份
if err != nil {
    os.Rename(backup, currentExe)
}
```

### References

- [Source: docs/epics.md#Epic-7]
- [Related: Story 1.5 - 跨平台編譯與發布]
- [NFR02: 安全需求 - 更新驗證]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Story completed in YOLO mode by dev-story workflow

#### Implementation Summary

**Files Created:**
1. `internal/update/types.go` - 核心資料結構（ReleaseInfo, UpdateStatus, UpdateConfig, UpdateResult）
2. `internal/update/checker.go` - GitHub API 整合和版本比對邏輯
3. `internal/update/checker_test.go` - Checker 單元測試
4. `internal/update/downloader.go` - 檔案下載器（支援進度回調、checksum 驗證、重試）
5. `internal/update/downloader_test.go` - Downloader 單元測試
6. `internal/update/replacer.go` - 跨平台執行檔替換邏輯
7. `internal/update/replacer_test.go` - Replacer 單元測試
8. `internal/update/manager.go` - 更新管理器（協調所有組件）

**Files Modified:**
1. `internal/tui/views/main_menu.go` - 添加更新橫幅顯示（UpdateAvailableMsg 處理）
2. `cmd/nightmare/main.go` - 添加 --update 命令行模式和 performUpdate() 函數
3. `internal/app/app.go` - 整合背景更新檢查邏輯
4. `go.mod` / `go.sum` - 添加 semver 依賴

**Test Results:**
- 16 個單元測試全部通過
- 涵蓋版本比對、下載、checksum 驗證、檔案替換等核心功能
- 建構成功，無編譯錯誤

**Key Features Implemented:**
- ✅ GitHub Releases API 整合（自動獲取最新版本）
- ✅ 語義化版本比對（使用 semver/v3）
- ✅ SHA256 checksum 驗證
- ✅ 跨平台執行檔替換（Windows/macOS/Linux）
- ✅ 備份與回滾機制
- ✅ 進度回調系統
- ✅ 重試邏輯（最多 3 次）
- ✅ API rate limiting 處理
- ✅ 背景異步檢查（不阻塞 UI）
- ✅ 主選單更新橫幅
- ✅ 命令行模式 (--update)
- ✅ 更新檢查間隔緩存（24 小時）

**AC Verification:**
- ✅ AC1: 背景檢查更新，主選單顯示提示
- ✅ AC2: 互動式更新流程（下載、進度、驗證、替換）
- ✅ AC3: 錯誤處理與降級（友善訊息、保留原檔）
- ✅ AC4: 命令行更新模式 (--update)
- ✅ AC5: 語義化版本比對（跳過 draft/prerelease）

**Technical Notes:**
- 使用 `github.com/Masterminds/semver/v3` 進行版本比對
- 更新管理器採用回調模式，支援狀態變更和進度追蹤
- Windows 平台需特殊處理（無法覆蓋運行中的執行檔）
- 所有檔案操作使用原子重命名，確保不會損壞
- 24 小時檢查間隔通過 `.last_check` 檔案實作
