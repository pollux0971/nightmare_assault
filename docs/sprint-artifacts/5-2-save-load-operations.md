# Story 5.2: 存檔操作

Status: done

## Story

As a 玩家,
I want 隨時存檔和讀檔,
so that 我可以保存進度並在需要時繼續.

## Acceptance Criteria

### AC1: 存檔指令與槽位選擇

**Given** 玩家在遊戲中
**When** 輸入 `/save` 或 `/save 1`（指定槽位）
**Then** 顯示存檔槽選擇介面（1-3）
**And** 顯示每個槽的當前狀態：
  - 空槽：顯示「空」
  - 已使用：顯示「章節 X | 遊玩時間 | 存檔時間」
**And** 高亮顯示當前選中的槽位

### AC2: 存檔執行與性能

**Given** 玩家選擇存檔槽並確認
**When** 執行存檔操作
**Then** 收集當前完整遊戲狀態
**And** 序列化為 JSON 格式
**And** 寫入對應槽位檔案
**And** 存檔完成時間 < 500ms (NFR01)
**And** 顯示「進度已保存於槽位 X」確認訊息

### AC3: 讀檔指令與列表顯示

**Given** 玩家在主選單或遊戲中
**When** 輸入 `/load` 或 `/load 2`（指定槽位）
**Then** 顯示可用存檔清單
**And** 空槽顯示為灰色不可選
**And** 已使用槽位顯示存檔預覽資訊
**And** 選擇存檔後載入遊戲狀態
**And** 讀檔完成時間 < 500ms (NFR01)

### AC4: 覆蓋確認機制

**Given** 存檔槽已有資料
**When** 玩家嘗試覆蓋存檔
**Then** 顯示確認對話框：「覆蓋現有存檔？（章節 X, 時間 Y）」
**And** 提供「是/否」選項
**And** 選擇「否」取消操作並返回

### AC5: 錯誤處理與恢復

**Given** 存檔/讀檔過程中發生錯誤
**When** 遇到檔案寫入失敗/權限不足/磁碟空間不足/檔案損壞
**Then** 顯示敘事化錯誤訊息（符合遊戲世界觀）
**And** 提供具體的解決建議
**And** 不影響當前遊戲狀態（存檔失敗時）

## Tasks / Subtasks

- [x] Task 1: 實作存檔管理器核心 (AC: #1, #2)
  - [x] 建立 `internal/game/save/manager.go`
  - [x] 定義 SaveManager struct
  - [x] 實作 Save(slotID int, gameState *GameState) error
  - [x] 實作狀態收集邏輯（從各模組收集資料）

- [x] Task 2: 實作讀檔管理器 (AC: #3)
  - [x] 實作 Load(slotID int) (*GameState, error)
  - [x] 實作狀態還原邏輯（分發至各模組）
  - [x] 添加版本檢查與遷移調用

- [x] Task 3: 實作存檔槽位 UI (AC: #1, #4)
  - [x] 建立 `internal/tui/components/save_slot_list.go`
  - [x] 實作槽位選擇元件（BubbleTea list）
  - [x] 實作槽位資訊顯示（預覽）
  - [x] 實作覆蓋確認對話框

- [x] Task 4: 整合斜線指令 (AC: #1, #3)
  - [x] 建立 `internal/game/commands/save.go` (包含 save 和 load 功能)
  - [x] 實作指令解析（處理可選槽位參數）
  - [x] 註冊至指令系統

- [x] Task 5: 實作錯誤處理 (AC: #5)
  - [x] 定義敘事化錯誤訊息映射
  - [x] 實作檔案 I/O 錯誤處理
  - [x] 實作 checksum 驗證失敗處理
  - [x] 實作回滾機制（存檔失敗時不損壞舊檔案）

- [x] Task 6: 性能優化與測試 (AC: #2, #3)
  - [x] 測試存檔操作 < 500ms
  - [x] 測試讀檔操作 < 500ms
  - [x] 壓力測試（大型存檔）
  - [x] 測試並發安全性（避免多次存檔衝突）

- [x] Task 7: 整合測試
  - [x] 測試完整存檔→退出→讀檔流程
  - [x] 測試覆蓋確認流程
  - [x] 測試錯誤情境（權限/空間/損壞）
  - [x] 測試跨平台路徑處理

## Dev Notes

### 架構模式

- **存檔管理器**: 單例模式，全域可訪問
- **狀態收集**: 使用 Observer 模式，各模組註冊狀態提供者
- **錯誤處理**: 敘事化包裝，例如「記憶正在崩解...（磁碟空間不足）」
- **事務性存檔**: 先寫入臨時檔案，成功後再替換（避免損壞舊存檔）

### 技術約束

- 存檔操作必須在 500ms 內完成（NFR01）
- 使用 atomic write 模式（寫入 .tmp 後 rename）
- 讀檔時驗證 checksum 確保完整性
- 錯誤訊息必須敘事化（符合 UX Design 要求）

### 存檔槽位 UI 設計

```
┌─ 選擇存檔槽位 ─────────────────────┐
│                                    │
│  [1] 章節 3 - 廢棄醫院             │
│      遊玩時間: 1h 23m              │
│      存檔時間: 2025-12-10 22:30   │
│                                    │
│  [2] 空                            │
│                                    │
│  [3] 章節 1 - 夢境開端             │
│      遊玩時間: 15m                 │
│      存檔時間: 2025-12-09 18:45   │
│                                    │
│  [ESC] 取消                        │
└────────────────────────────────────┘
```

### 敘事化錯誤訊息範例

| 技術錯誤 | 敘事化訊息 |
|---------|-----------|
| 檔案寫入失敗 | 「記憶無法固化...你的思緒在虛空中流失。（檢查磁碟權限）」 |
| 磁碟空間不足 | 「虛空已滿，無法容納更多記憶...（需要至少 1MB 空間）」 |
| 檔案損壞 | 「這段記憶已被扭曲...無法還原現實。（存檔檔案損壞）」 |
| checksum 失敗 | 「記憶的完整性遭到破壞...（檔案驗證失敗）」 |

### 狀態收集與還原流程

**存檔時收集：**
1. PlayerState from GameEngine
2. GameState from RuleEngine
3. Teammates from NPCManager
4. StoryContext from StoryGenerator
5. 計算並添加 Checksum

**讀檔時還原：**
1. 驗證 Checksum
2. 檢查版本並執行遷移
3. 還原至各模組（按依賴順序）
4. 重新初始化 UI 狀態

### References

- [Source: /home/pollux/Desktop/nightmare-assault/docs/epics.md#Epic-5]
- [Depends on: Story 5.1 - 存檔資料結構]
- [Related: Story 2.6 - 基礎斜線指令]

## File List

- `internal/game/save/manager.go` (NEW) - SaveManager 核心存檔管理器
- `internal/game/save/manager_test.go` (NEW) - SaveManager 單元測試
- `internal/game/save/errors.go` (NEW) - 敘事化錯誤處理
- `internal/game/save/errors_test.go` (NEW) - 錯誤處理測試
- `internal/game/save/performance_test.go` (NEW) - 性能與並發測試
- `internal/game/save/integration_test.go` (NEW) - 整合測試
- `internal/tui/components/save_slot_list.go` (NEW) - 存檔槽位 UI 元件
- `internal/tui/components/save_slot_list_test.go` (NEW) - UI 元件測試
- `internal/game/commands/save.go` (NEW) - 存讀檔斜線指令
- `internal/game/commands/save_test.go` (NEW) - 指令測試

## Change Log

- 2025-12-11: Story implementation complete
  - SaveManager 實作完成，支援存讀檔操作
  - 敘事化錯誤訊息系統
  - 存檔槽位選擇 UI (BubbleTea)
  - 覆蓋確認對話框
  - /save 和 /load 斜線指令
  - 性能測試：存檔 ~0.1ms，讀檔 ~0.1ms (遠低於 500ms 要求)
  - 並發安全性測試通過
  - 所有測試通過

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Plan

1. Task 1&2: SaveManager 核心功能 - Save() 和 Load() 方法
2. Task 3: 存檔槽位 UI - SaveSlotList BubbleTea 元件
3. Task 4: 斜線指令 - /save 和 /load 指令解析
4. Task 5: 敘事化錯誤處理 - NarrativeError, WrapSaveError, WrapLoadError
5. Task 6: 性能測試 - 驗證 < 500ms 並測試並發安全性
6. Task 7: 整合測試 - 完整流程測試

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for implementation - all acceptance criteria defined
- UI mockup included for clarity
- Error handling follows narrative style requirement
- ✅ Task 1&2 完成：SaveManager 包含 Save, Load, GetSlotInfo, GetAllSlotInfo, DeleteSlot
- ✅ Task 3 完成：SaveSlotList UI 元件，支援存讀模式、覆蓋確認
- ✅ Task 4 完成：ParseSlotCommand, IsSaveCommand, IsLoadCommand
- ✅ Task 5 完成：敘事化錯誤映射，ErrorCategory 分類
- ✅ Task 6 完成：存檔 ~0.1ms，讀檔 ~0.1ms，並發安全
- ✅ Task 7 完成：完整流程整合測試通過
