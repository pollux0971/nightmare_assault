# Story 5.1: 存檔資料結構

Status: done

## Story

As a 系統,
I want 定義完整的存檔資料結構,
so that 遊戲狀態可以完整保存.

## Acceptance Criteria

### AC1: JSON 格式存檔結構

**Given** 需要保存遊戲狀態
**When** 設計存檔結構
**Then** JSON 格式包含以下完整資料：
  - 元資料：版本、存檔時間、遊玩時間、難度、故事長度
  - 玩家狀態：HP、SAN、位置、背包物品、已知線索
  - 遊戲狀態：當前章節、章節進度、已觸發規則、已發現規則
  - 隊友狀態：存活列表、各隊友 HP、位置、攜帶物品、關係值
  - 故事上下文：最近章節摘要（壓縮後）、當前場景描述、Game Bible 快照

### AC2: 版本遷移機制

**Given** 存檔需要向前相容
**When** 讀取舊版本存檔（版本號 < 當前版本）
**Then** 自動遷移至新格式
**And** 不遺失重要資料（玩家狀態、故事進度）
**And** 記錄遷移日誌以便 debug

### AC3: 存檔路徑與大小限制

**Given** 存檔檔案需要持久化
**When** 儲存至磁碟
**Then** 路徑為 `~/.nightmare/saves/save_{1-3}.json`
**And** 自動創建目錄（如不存在）
**And** 存檔大小 < 1MB（壓縮後的上下文控制大小）

### AC4: 資料完整性驗證

**Given** 存檔檔案寫入完成
**When** 驗證資料完整性
**Then** 計算並儲存 checksum（SHA256）
**And** 讀取時驗證 checksum
**And** 檔案損壞時顯示友善錯誤訊息

## Tasks / Subtasks

- [x] Task 1: 定義 SaveData 核心結構 (AC: #1)
  - [x] 建立 `internal/game/save/schema.go`
  - [x] 定義 SaveData struct 包含所有必要欄位
  - [x] 定義 Metadata、PlayerState、GameState、TeammateState 子結構
  - [x] 添加 JSON tags 確保正確序列化

- [x] Task 2: 實作版本遷移系統 (AC: #2)
  - [x] 定義版本常數（當前版本 = 1）
  - [x] 建立 MigrationHandler interface
  - [x] 實作 v0->v1 遷移邏輯（未來擴展）
  - [x] 添加遷移日誌記錄

- [x] Task 3: 實作存檔路徑管理 (AC: #3)
  - [x] 建立 GetSavePath(slotID int) string 函數
  - [x] 實作目錄自動創建邏輯
  - [x] 處理跨平台路徑（Windows/Linux/macOS）
  - [x] 添加存檔大小檢查（警告 > 1MB）

- [x] Task 4: 實作資料完整性驗證 (AC: #4)
  - [x] 添加 Checksum 欄位至 SaveData
  - [x] 實作 ComputeChecksum() 函數（SHA256）
  - [x] 實作 VerifyChecksum() 函數
  - [x] 添加檔案損壞錯誤處理

- [x] Task 5: 單元測試
  - [x] 測試完整存檔結構序列化/反序列化
  - [x] 測試版本遷移邏輯
  - [x] 測試路徑管理（模擬不同 OS）
  - [x] 測試 checksum 驗證（正常/損壞情境）

## Dev Notes

### 架構模式

- **資料結構設計**: 使用嵌套 struct 確保清晰的層次結構
- **序列化格式**: 使用 JSON 以便人類可讀（方便 debug）
- **版本控制**: 語義化版本號，保留向前相容性
- **錯誤處理**: 使用 敘事化錯誤訊息（符合遊戲世界觀）

### 技術約束

- Go 標準庫 `encoding/json` 用於序列化
- Go 標準庫 `crypto/sha256` 用於 checksum
- 路徑使用 `filepath.Join()` 確保跨平台
- 存檔大小限制 1MB（符合 NFR01 性能需求）

### SaveData 結構示例

```go
type SaveData struct {
    Version  int      `json:"version"`
    Metadata Metadata `json:"metadata"`
    Player   PlayerState `json:"player"`
    Game     GameState `json:"game"`
    Teammates []TeammateState `json:"teammates"`
    Context  StoryContext `json:"context"`
    Checksum string `json:"checksum"`
}

type Metadata struct {
    SavedAt    time.Time `json:"saved_at"`
    PlayTime   int       `json:"play_time_seconds"`
    Difficulty string    `json:"difficulty"`
    StoryLength string   `json:"story_length"`
}

type PlayerState struct {
    HP       int      `json:"hp"`
    SAN      int      `json:"san"`
    Location string   `json:"location"`
    Inventory []Item  `json:"inventory"`
    KnownClues []string `json:"known_clues"`
}

type GameState struct {
    CurrentChapter int      `json:"current_chapter"`
    ChapterProgress float32 `json:"chapter_progress"`
    TriggeredRules []string `json:"triggered_rules"`
    DiscoveredRules []string `json:"discovered_rules"`
}

type TeammateState struct {
    Name     string `json:"name"`
    Alive    bool   `json:"alive"`
    HP       int    `json:"hp"`
    Location string `json:"location"`
    Items    []Item `json:"items"`
    Relationship int `json:"relationship"`
}

type StoryContext struct {
    RecentSummary string `json:"recent_summary"`
    CurrentScene  string `json:"current_scene"`
    GameBible     string `json:"game_bible_snapshot"`
}
```

### References

- [Source: /home/pollux/Desktop/nightmare-assault/docs/epics.md#Epic-5]
- [Related: Story 5.2 - 存檔操作]
- [Related: Story 5.3 - 章節壓縮機制]

## File List

- `internal/game/save/schema.go` (NEW) - SaveData 核心資料結構定義
- `internal/game/save/schema_test.go` (NEW) - 資料結構單元測試
- `internal/game/save/migration.go` (NEW) - 版本遷移系統
- `internal/game/save/migration_test.go` (NEW) - 版本遷移單元測試
- `internal/game/save/path.go` (NEW) - 存檔路徑管理
- `internal/game/save/path_test.go` (NEW) - 路徑管理單元測試
- `internal/game/save/checksum.go` (NEW) - SHA256 校驗碼功能
- `internal/game/save/checksum_test.go` (NEW) - 校驗碼單元測試

## Change Log

- 2025-12-11: Story implementation complete
  - Created save package with complete data structures
  - Implemented version migration system with v0->v1 migration
  - Added cross-platform path management
  - Implemented SHA256 checksum verification
  - All 91 unit tests passing

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Plan

1. Task 1: 定義 SaveData 核心結構 - 建立 schema.go 包含完整存檔結構
2. Task 2: 實作版本遷移系統 - 建立 migration.go 包含 MigrationHandler interface
3. Task 3: 實作存檔路徑管理 - 建立 path.go 包含跨平台路徑功能
4. Task 4: 實作資料完整性驗證 - 建立 checksum.go 包含 SHA256 驗證
5. Task 5: 單元測試 - 所有功能都有完整測試覆蓋

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for implementation - all acceptance criteria defined
- Structure designed for extensibility (version migration ready)
- ✅ Task 1 完成：SaveData 結構包含所有 AC1 要求的欄位
- ✅ Task 2 完成：MigrationHandler interface 和 V0ToV1Migration 實作
- ✅ Task 3 完成：GetSavePath, EnsureSaveDir, CheckSaveSize 等功能
- ✅ Task 4 完成：ComputeChecksum, VerifyChecksum, CorruptedError
- ✅ Task 5 完成：91 個單元測試全部通過
- 注意：internal/game/config_test.go 有一個預先存在的測試失敗，非本次變更引入
