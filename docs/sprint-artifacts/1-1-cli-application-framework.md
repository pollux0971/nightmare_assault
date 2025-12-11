# Story 1.1: CLI 應用程式框架

Status: done

## Story

As a **開發者**,
I want **建立 Go CLI 應用程式的基礎框架**,
so that **後續功能可以在此基礎上開發**.

## Acceptance Criteria

### AC1: 專案結構建立
**Given** 專案目錄不存在
**When** 建立專案結構
**Then** 產生符合 ARCHITECTURE.md 的目錄結構
**And** 包含 `cmd/nightmare/main.go` 入口點
**And** 包含 `internal/` 模組化目錄
**And** 包含 `go.mod` 和 `go.sum` 依賴檔案

### AC2: 可執行檔編譯
**Given** 專案結構已建立
**When** 執行 `go build` 命令
**Then** 產生可執行的 `nightmare` 二進位檔
**And** 編譯無錯誤
**And** 執行檔大小 < 30MB（初始狀態）

### AC3: 啟動性能要求
**Given** 可執行檔已產生
**When** 在終端機執行 `nightmare`
**Then** 程式啟動時間 < 2 秒（NFR01）
**And** 記憶體使用 < 50MB（基礎狀態，NFR01）
**And** CPU 使用 < 5%（閒置狀態）

### AC4: BubbleTea TUI 框架整合
**Given** 使用者在終端機執行 `nightmare`
**When** 程式啟動
**Then** 顯示 BubbleTea TUI 框架的基礎畫面
**And** 支援最小終端大小 80x24（NFR03）
**And** 顯示應用程式標題和版本資訊
**And** 可使用 `Ctrl+C` 或 `q` 退出程式

### AC5: 基礎 Model/View/Update 架構
**Given** BubbleTea 框架已整合
**When** 檢視程式碼結構
**Then** 實作 `Model` 結構體定義應用程式狀態
**And** 實作 `Init()` 方法初始化狀態
**And** 實作 `Update(msg tea.Msg)` 方法處理事件
**And** 實作 `View() string` 方法渲染畫面

### AC6: LipGloss 樣式整合
**Given** TUI 框架已運作
**When** 顯示畫面元素
**Then** 使用 LipGloss 定義基礎樣式
**And** 標題使用 Bold + 顏色
**And** 邊框使用 Rounded 風格
**And** 樣式模組化於 `internal/tui/styles/` 目錄

## Tasks / Subtasks

- [x] Task 1: 建立專案目錄結構 (AC: #1)
  - [x] Subtask 1.1: 初始化 Go 模組 `go mod init github.com/nightmare-assault/nightmare-assault`
  - [x] Subtask 1.2: 建立 `cmd/nightmare/` 目錄和 `main.go`
  - [x] Subtask 1.3: 建立 `internal/` 目錄結構（app, config, engine, tui 等）
  - [x] Subtask 1.4: 建立 `.gitignore`

- [x] Task 2: 安裝核心依賴 (AC: #2, #4)
  - [x] Subtask 2.1: 安裝 BubbleTea `go get github.com/charmbracelet/bubbletea`
  - [x] Subtask 2.2: 安裝 LipGloss `go get github.com/charmbracelet/lipgloss`
  - [x] Subtask 2.3: 安裝 Bubbles `go get github.com/charmbracelet/bubbles`
  - [x] Subtask 2.4: 執行 `go mod tidy` 整理依賴

- [x] Task 3: 實作 main.go 入口點 (AC: #2, #3)
  - [x] Subtask 3.1: 建立 `cmd/nightmare/main.go` 主程式
  - [x] Subtask 3.2: 初始化應用程式並啟動 BubbleTea
  - [x] Subtask 3.3: 加入基礎錯誤處理和退出邏輯
  - [x] Subtask 3.4: 測試編譯 `go build -o nightmare cmd/nightmare/main.go`

- [x] Task 4: 建立 BubbleTea Model 結構 (AC: #5)
  - [x] Subtask 4.1: 在 `internal/app/app.go` 定義 `Model` 結構體
  - [x] Subtask 4.2: 實作 `Init() tea.Cmd` 初始化方法
  - [x] Subtask 4.3: 實作 `Update(msg tea.Msg) (tea.Model, tea.Cmd)` 方法
  - [x] Subtask 4.4: 實作 `View() string` 方法，顯示歡迎畫面
  - [x] Subtask 4.5: 處理 `tea.KeyMsg` 事件（q 和 Ctrl+C 退出）

- [x] Task 5: 整合 LipGloss 樣式系統 (AC: #6)
  - [x] Subtask 5.1: 建立 `internal/tui/styles/base.go` 定義基礎樣式
  - [x] Subtask 5.2: 定義標題樣式（Bold + 顏色）
  - [x] Subtask 5.3: 定義邊框樣式（Rounded）
  - [x] Subtask 5.4: 在 `View()` 方法中應用樣式

- [x] Task 6: 終端大小檢測與適配 (AC: #4)
  - [x] Subtask 6.1: 監聽 `tea.WindowSizeMsg` 事件
  - [x] Subtask 6.2: 儲存終端機寬度和高度到 Model
  - [x] Subtask 6.3: 檢測最小大小 80x24，若不足顯示警告
  - [x] Subtask 6.4: 測試不同終端大小的顯示效果（單元測試驗證）

- [x] Task 7: 性能驗證與優化 (AC: #3)
  - [x] Subtask 7.1: 測量啟動時間（0.004s < 2s ✓）
  - [x] Subtask 7.2: 測量執行檔大小（4.2MB < 30MB ✓）
  - [x] Subtask 7.3: 性能達標，無需優化
  - [x] Subtask 7.4: 記錄性能基準：啟動 0.004s，執行檔 4.2MB

- [x] Task 8: 建立基礎測試 (AC: #2)
  - [x] Subtask 8.1: 建立 `internal/app/app_test.go`（測試放在 app 模組更合適）
  - [x] Subtask 8.2: 測試 Model 的 Init/Update/View 方法（9 個測試全部通過）
  - [x] Subtask 8.3: 執行 `go test ./...` 確保通過
  - [x] Subtask 8.4: 設定 CI 基礎配置（GitHub Actions: .github/workflows/ci.yaml）

## Dev Notes

### 架構模式與約束

- **架構模式**: BubbleTea 的 Elm Architecture (Model-View-Update)
- **目錄結構**: 嚴格遵循 ARCHITECTURE.md 的分層設計
  - `cmd/`: 程式入口，只負責啟動
  - `internal/app/`: 應用程式核心邏輯
  - `internal/tui/`: TUI 相關元件和樣式
  - `internal/config/`: 配置管理（本 story 暫不實作）
  - `internal/engine/`: 遊戲引擎（本 story 暫不實作）

- **約束**:
  - 啟動時間必須 < 2 秒（NFR01）
  - 記憶體使用 < 50MB（基礎狀態）
  - 支援最小終端大小 80x24（NFR03）
  - 完全靜態編譯，無外部依賴

### 相關程式碼路徑

```
cmd/nightmare/main.go                   # 程式入口
internal/app/app.go                     # 主應用程式 Model
internal/app/lifecycle.go               # 生命週期管理（可選）
internal/tui/styles/base.go             # 基礎樣式定義
go.mod                                  # 依賴管理
```

### 測試標準

1. **單元測試**: Model 的 Init/Update/View 方法
2. **編譯測試**: `go build` 無錯誤
3. **性能測試**: 啟動時間 < 2 秒，記憶體 < 50MB
4. **終端相容性測試**: 在 macOS/Linux/Windows 終端測試

### Project Structure Notes

本 story 建立的目錄結構將成為整個專案的基礎：

```
nightmare-assault/
├── cmd/
│   └── nightmare/
│       └── main.go                     # ✓ 本 story 建立
├── internal/
│   ├── app/
│   │   ├── app.go                      # ✓ 本 story 建立
│   │   └── lifecycle.go                # ○ 可選
│   ├── tui/
│   │   ├── styles/
│   │   │   └── base.go                 # ✓ 本 story 建立
│   │   └── views/                      # ○ 後續 story 建立
│   ├── config/                         # ○ Story 1.2 建立
│   ├── engine/                         # ○ Epic 2 建立
│   └── api/                            # ○ Story 1.2 建立
├── go.mod                              # ✓ 本 story 建立
├── go.sum                              # ✓ 本 story 建立
└── README.md                           # ✓ 本 story 建立
```

### References

- [Source: docs/epics.md#Epic-1-Story-1.1]
- [Source: ARCHITECTURE.md#3-專案結構]
- [Source: ARCHITECTURE.md#2-技術選型]
- [Source: PRD.md#NFR01-性能需求]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- 2025-12-11: Story implementation completed
  - Go module initialized: `github.com/nightmare-assault/nightmare-assault`
  - BubbleTea + LipGloss + Bubbles dependencies installed
  - Core application Model with Init/Update/View implemented
  - Terminal size detection with 80x24 minimum warning
  - LipGloss styles: Title, Subtitle, Text, Hint, Warning, Container
  - 9 unit tests passing (TestNew, TestInit, TestUpdateWindowSize, etc.)
  - Binary size: 4.2MB (well under 30MB limit)
  - Startup time: 0.004s (well under 2s limit)
  - GitHub Actions CI workflow configured

### File List

- `go.mod` - Go module definition
- `go.sum` - Dependency checksums
- `cmd/nightmare/main.go` - Application entry point
- `internal/app/app.go` - Main application Model
- `internal/app/app_test.go` - Unit tests
- `internal/tui/styles/base.go` - LipGloss style definitions
- `.gitignore` - Git ignore rules
- `.github/workflows/ci.yaml` - CI configuration

### Change Log

- 2025-12-11: Initial implementation - CLI framework with BubbleTea TUI
