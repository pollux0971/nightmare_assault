# Story 1.3: 主選單介面

Status: done

## Story

As a **玩家**,
I want **看到清晰的主選單**,
so that **我可以開始新遊戲、讀取存檔、或調整設定**.

## Acceptance Criteria

### AC1: 主選單選項顯示
**Given** 玩家已完成 API 設定
**When** 進入主選單
**Then** 顯示以下選項：
  - 新遊戲（New Game）
  - 繼續遊戲（Continue Game）- 如有存檔時顯示
  - 設定（Settings）
  - 離開（Exit）
**And** 選項以垂直清單方式排列
**And** 當前選中的選項高亮顯示

### AC2: 鍵盤操作支援
**Given** 玩家在主選單
**When** 使用鍵盤操作
**Then** 方向鍵 ↑/↓ 或 j/k 可上下移動選擇
**And** 數字鍵 1-4 可直接選擇對應選項
**And** Enter 或 Space 執行當前選中的選項
**And** 操作響應時間 < 100ms（NFR01）

### AC3: 繼續遊戲選項動態顯示
**Given** 玩家沒有任何存檔
**When** 進入主選單
**Then** 「繼續遊戲」選項顯示為灰色（不可選）
**And** 選擇時顯示「無可用存檔」

**Given** 玩家有至少一個存檔
**When** 進入主選單
**Then** 「繼續遊戲」選項正常顯示（可選）
**And** 選擇後顯示存檔列表

### AC4: 設定選單
**Given** 玩家在主選單選擇「設定」
**When** 進入設定選單
**Then** 顯示以下設定選項：
  - 主題切換（Theme）
  - API 設定（API Settings）
  - 音效設定（Audio Settings）
  - 返回主選單（Back）
**And** 可使用相同鍵盤操作方式導航

### AC5: 主選單視覺設計
**Given** 主選單顯示中
**When** 玩家觀察畫面
**Then** 顯示遊戲標題「惡夢驚襲 NIGHTMARE ASSAULT」
**And** 標題使用 ASCII Art 或 LipGloss 樣式裝飾
**And** 顯示當前版本號（例如：v0.1.0）
**And** 選項使用 Rounded 邊框包圍
**And** 主題顏色符合 Midnight 主題（預設）

### AC6: 狀態機管理
**Given** 主選單實作
**When** 檢視程式碼結構
**Then** 使用狀態機管理選單導航：
  - `StateMainMenu`: 主選單
  - `StateSettings`: 設定選單
  - `StateAPISettings`: API 設定
  - `StateThemeSettings`: 主題設定
**And** 狀態轉換邏輯清晰
**And** 支援返回上一層選單（ESC 或 Backspace）

## Tasks / Subtasks

- [x] Task 1: 建立主選單 Model (AC: #1, #6)
  - [x] Subtask 1.1: 建立 `internal/tui/views/main_menu.go`
  - [x] Subtask 1.2: 定義 `MainMenuModel` 結構體
  - [x] Subtask 1.3: 定義 `MenuAction` 列舉
  - [x] Subtask 1.4: 實作 `Init()` 方法初始化選單
  - [x] Subtask 1.5: 定義選單項目清單（NewGame, Continue, Settings, Exit）

- [x] Task 2: 整合 Bubbles List 元件 (AC: #1, #2)
  - [x] Subtask 2.1: 使用 bubbles/list 元件
  - [x] Subtask 2.2: 建立 `MenuItem` 結構實作 `list.Item` 介面
  - [x] Subtask 2.3: 初始化 `list.Model` 並設定樣式
  - [x] Subtask 2.4: 配置 list 快捷鍵（↑/↓, j/k, Enter）

- [x] Task 3: 實作 Update 邏輯 (AC: #2, #6)
  - [x] Subtask 3.1: 處理 `tea.KeyMsg` 鍵盤事件
  - [x] Subtask 3.2: 數字鍵 1-4 直接選擇選項
  - [x] Subtask 3.3: Enter 執行選中選項
  - [x] Subtask 3.4: ESC 返回上一層（設定選單專用）
  - [x] Subtask 3.5: 狀態轉換邏輯（切換至新遊戲/設定/離開）

- [x] Task 4: 實作 View 渲染 (AC: #5)
  - [x] Subtask 4.1: 建立遊戲標題 ASCII Art
  - [x] Subtask 4.2: 顯示版本號
  - [x] Subtask 4.3: 使用 LipGloss 渲染選單項目（高亮選中項目）
  - [x] Subtask 4.4: 加入 Rounded 邊框
  - [x] Subtask 4.5: 底部顯示快捷鍵提示

- [x] Task 5: 繼續遊戲動態顯示 (AC: #3)
  - [x] Subtask 5.1: hasSaveFiles 參數控制是否啟用
  - [x] Subtask 5.2: 若無存檔，「繼續遊戲」選項顯示灰色
  - [x] Subtask 5.3: 灰色選項不可選中（skipDisabled 跳過）
  - [x] Subtask 5.4: MenuItem.enabled 控制狀態

- [x] Task 6: 建立設定選單 Model (AC: #4)
  - [x] Subtask 6.1: 建立 `internal/tui/views/settings_menu.go`
  - [x] Subtask 6.2: 定義 `SettingsMenuModel` 結構體
  - [x] Subtask 6.3: 設定選項清單（Theme, API, Audio, Back）
  - [x] Subtask 6.4: 實作 Update 和 View 方法
  - [x] Subtask 6.5: 返回主選單邏輯

- [x] Task 7: 主選單與設定選單導航 (AC: #6)
  - [x] Subtask 7.1: 在 `internal/app/app.go` 整合主選單 Model
  - [x] Subtask 7.2: 實作狀態切換：MainMenu ↔ Settings
  - [x] Subtask 7.3: ESC 或 Backspace 返回上一層
  - [x] Subtask 7.4: MenuSelectMsg / SettingsSelectMsg 訊息處理

- [x] Task 8: 離開遊戲確認 (AC: #1)
  - [x] Subtask 8.1: 選擇「離開」時顯示確認對話框
  - [x] Subtask 8.2: 確認對話框：「確定要離開嗎？(y/n)」
  - [x] Subtask 8.3: 按 y 退出程式，按 n 返回主選單
  - [x] Subtask 8.4: 按 q 或 Ctrl+C 直接退出（無確認）

- [x] Task 9: 主題樣式整合 (AC: #5)
  - [x] Subtask 9.1: 使用 LipGloss 樣式（Cosmic Horror theme）
  - [x] Subtask 9.2: 套用主題顏色至選單元素 (#9D4EDD purple)
  - [x] Subtask 9.3: 高亮選中項目使用主題的 accent 顏色
  - [x] Subtask 9.4: (Theme 切換將在 Story 1.4 實作)

- [x] Task 10: 響應式佈局 (AC: #5)
  - [x] Subtask 10.1: 監聽 `tea.WindowSizeMsg`
  - [x] Subtask 10.2: 根據終端寬度調整選單寬度
  - [x] Subtask 10.3: list.SetSize() 響應式調整
  - [x] Subtask 10.4: app.go MinWidth/MinHeight 檢查

- [x] Task 11: 整合測試 (AC: #1-#6)
  - [x] Subtask 11.1: 測試主選單顯示與導航 (8 tests)
  - [x] Subtask 11.2: 測試有/無存檔時「繼續遊戲」狀態
  - [x] Subtask 11.3: 測試設定選單導航與返回
  - [x] Subtask 11.4: 測試離開遊戲確認流程
  - [x] Subtask 11.5: `go test ./...` 全部通過

## Dev Notes

### 架構模式與約束

- **元件模式**: 使用 BubbleTea 的 Model 組合模式，主選單和設定選單是獨立的 Model
- **狀態機**: 使用列舉定義選單狀態，清晰管理導航流程
- **樣式一致性**: 遵循 UX Design Specification 的選單設計模式
- **響應式**: 根據終端大小調整佈局（ARCHITECTURE.md#8.10）

### 相關程式碼路徑

```
internal/tui/views/
├── main_menu.go                        # 主選單 Model
├── settings_menu.go                    # 設定選單 Model
└── menu_item.go                        # 選單項目結構（可選）

internal/tui/styles/
└── menu.go                             # 選單樣式定義

internal/app/
└── app.go                              # 整合主選單到應用程式

internal/game/save/
└── manager.go                          # 存檔檢測（Story 5.1 實作）
```

### 測試標準

1. **單元測試**:
   - MainMenuModel 的 Init/Update/View
   - SettingsMenuModel 的 Init/Update/View
   - 鍵盤事件處理

2. **整合測試**:
   - 主選單 → 設定選單 → 返回
   - 有/無存檔時的顯示
   - 離開確認流程

3. **視覺測試**:
   - 在不同終端大小測試
   - 主題切換時的顏色變化
   - 高亮與動畫效果

### 選單項目資料結構

```go
// internal/tui/views/main_menu.go
package views

type MenuItem struct {
    Title       string
    Description string
    Enabled     bool
    Action      MenuAction
}

type MenuAction int

const (
    ActionNewGame MenuAction = iota
    ActionContinue
    ActionSettings
    ActionExit
)

type MenuState int

const (
    StateMainMenu MenuState = iota
    StateSettings
    StateAPISettings
    StateThemeSettings
)

type MainMenuModel struct {
    list      list.Model
    state     MenuState
    width     int
    height    int
    styles    MenuStyles
}
```

### 主選單視覺範例

```
┌─────────────────────────────────────────────┐
│                                             │
│     ███╗   ██╗██╗ ██████╗ ██╗  ██╗████████╗│
│     ████╗  ██║██║██╔════╝ ██║  ██║╚══██╔══╝│
│     ██╔██╗ ██║██║██║  ███╗███████║   ██║   │
│     ██║╚██╗██║██║██║   ██║██╔══██║   ██║   │
│     ██║ ╚████║██║╚██████╔╝██║  ██║   ██║   │
│     ╚═╝  ╚═══╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   │
│                                             │
│          NIGHTMARE ASSAULT v0.1.0           │
│                                             │
├─────────────────────────────────────────────┤
│                                             │
│  ❯ 新遊戲 (New Game)                        │
│    繼續遊戲 (Continue Game)                 │
│    設定 (Settings)                          │
│    離開 (Exit)                              │
│                                             │
└─────────────────────────────────────────────┘
  ↑/↓: 選擇 | Enter: 確認 | q: 離開
```

### Project Structure Notes

主選單是玩家進入遊戲的入口，連接所有主要功能：
- **新遊戲** → Epic 2（核心遊戲循環）
- **繼續遊戲** → Epic 5（存檔系統）
- **設定** → Story 1.2（API）、Story 1.4（主題）、Epic 6（音效）
- **離開** → 退出程式

### References

- [Source: docs/epics.md#Epic-1-Story-1.3]
- [Source: ARCHITECTURE.md#8-TUI-介面]
- [Source: docs/ux-design-specification.md#主選單設計]
- [Source: PRD.md#NFR03-可用性需求]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- 2025-12-11: 完成所有 11 個 Tasks
- 主選單使用 ASCII Art 標題 (NIGHTMARE)
- 設定選單支援：主題、API、音效、返回
- 離開確認對話框 (y/n)
- 8 個新測試通過

## File List

**New Files:**
- `internal/tui/views/main_menu.go` - 主選單 Model
- `internal/tui/views/settings_menu.go` - 設定選單 Model
- `internal/tui/views/main_menu_test.go` - 主選單測試 (8 tests)

**Modified Files:**
- `internal/app/app.go` - 整合主選單、設定選單導航
- `internal/app/app_test.go` - 更新測試

## Change Log

- 2025-12-11: Story 1-3 completed with main menu, settings menu, exit confirmation
