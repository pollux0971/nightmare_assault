# Story 5.4: 遊戲日誌功能

Status: done

## Story

As a 玩家,
I want 查看歷史對話紀錄,
so that 我可以回顧錯過的線索.

## Acceptance Criteria

### AC1: 日誌指令與顯示

**Given** 玩家在遊戲中
**When** 輸入 `/log` 或 `/log 20`（指定筆數）
**Then** 顯示最近 N 筆對話（預設 10，最多 100）
**And** 日誌條目包含：時間戳記、對話類型、內容
**And** 對話類型包括：敘事、玩家輸入、選項選擇、系統訊息

### AC2: 日誌捲動與導航

**Given** 日誌顯示中
**When** 使用上下方向鍵
**Then** 可捲動查看更早的歷史
**And** 捲動流暢無卡頓
**When** 按 ESC 或 q
**Then** 關閉日誌檢視，返回遊戲

### AC3: 日誌格式與分類

**Given** 日誌內容顯示
**When** 渲染日誌條目
**Then** 時間戳記格式為「HH:MM:SS」
**And** 不同類型使用不同顏色標記：
  - 敘事：白色
  - 玩家輸入：青色
  - 選項選擇：黃色
  - 系統訊息：灰色
**And** 系統訊息（存檔/讀檔/指令回應）與遊戲敘事明確區分

### AC4: 日誌持久化

**Given** 遊戲進行中記錄日誌
**When** 執行存檔操作
**Then** 日誌隨存檔一起保存（最多保存最近 1000 筆）
**And** 日誌資料包含在 SaveData 結構中
**When** 讀取存檔
**Then** 還原歷史日誌
**And** 玩家可查看讀檔前的對話紀錄

### AC5: 日誌容量管理

**Given** 遊戲持續進行
**When** 日誌條目超過 1000 筆
**Then** 使用環形緩衝區（Ring Buffer）自動刪除最舊條目
**And** 記憶體使用保持在合理範圍（< 10MB）
**And** 不影響遊戲性能

## Tasks / Subtasks

- [x] Task 1: 實作日誌資料結構 (AC: #3, #5)
  - [x] 建立 `internal/game/log.go`
  - [x] 定義 LogEntry struct（時間戳、類型、內容）
  - [x] 定義 LogType enum（Narrative/PlayerInput/OptionChoice/System）
  - [x] 實作環形緩衝區 RingBuffer（容量 1000）

- [x] Task 2: 實作日誌記錄邏輯 (AC: #1, #5)
  - [x] 建立 GameLog manager
  - [x] 實作 AddEntry(entryType LogType, content string)
  - [x] 實作 GetRecentEntries(n int) []LogEntry
  - [x] 實作記憶體管理（環形緩衝）

- [x] Task 3: 實作日誌檢視 UI (AC: #1, #2)
  - [x] 建立 `internal/tui/views/log_view.go`
  - [x] 實作 BubbleTea viewport 元件（捲動）
  - [x] 實作時間戳記與顏色標記渲染
  - [x] 實作鍵盤導航（上下捲動、ESC 關閉）

- [x] Task 4: 實作 /log 指令 (AC: #1)
  - [x] 建立 `internal/game/commands/log.go`
  - [x] 實作指令解析（可選參數：筆數）
  - [x] 實作 ParseLogCommand 和 IsLogCommand
  - [x] 支援預設值 10 和上限 100

- [x] Task 5: 整合存檔系統 (AC: #4)
  - [x] 擴展 SaveData 結構包含 LogEntries
  - [x] 實作日誌序列化（保存所有 log entries）
  - [x] 實作日誌還原邏輯
  - [x] 測試存檔/讀檔包含日誌資料

- [x] Task 6: 性能與容量測試
  - [x] 測試長時間遊戲日誌記憶體使用
  - [x] 測試環形緩衝區正確性（超過 1000 筆）
  - [x] 測試並發安全性
  - [x] 性能基準測試

## Dev Notes

### 架構模式

- **環形緩衝區**: 固定容量 1000，自動覆蓋最舊條目
- **單例模式**: GameLog 全域可訪問
- **Observer 模式**: 自動記錄所有 GameEngine 事件
- **Viewport 元件**: BubbleTea 內建捲動支援

### 技術約束

- 日誌容量上限 1000 筆（記憶體 < 10MB）
- 存檔時僅保存最近 1000 筆
- 捲動必須流暢（60 FPS）
- 時間戳記使用遊戲內時間（非真實時間）

### LogEntry 結構

```go
type LogType int

const (
    LogNarrative LogType = iota
    LogPlayerInput
    LogOptionChoice
    LogSystem
)

type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Type      LogType   `json:"type"`
    Content   string    `json:"content"`
}

type GameLog struct {
    entries *RingBuffer // 容量 1000
    mu      sync.RWMutex
}

func (g *GameLog) AddEntry(entryType LogType, content string) {
    g.mu.Lock()
    defer g.mu.Unlock()

    entry := LogEntry{
        Timestamp: time.Now(),
        Type:      entryType,
        Content:   content,
    }
    g.entries.Push(entry)
}

func (g *GameLog) GetRecentEntries(n int) []LogEntry {
    g.mu.RLock()
    defer g.mu.RUnlock()

    return g.entries.GetLast(n)
}
```

### 日誌檢視 UI 設計

```
┌─ 遊戲日誌 (最近 20 筆) ───────────────────┐
│                                            │
│ [22:15:32] [敘事] 你推開生鏽的鐵門...     │
│ [22:15:45] [玩家] 檢查牆上的血跡          │
│ [22:15:48] [敘事] 血跡呈現奇怪的圖案...   │
│ [22:16:03] [選項] 1. 觸碰圖案              │
│ [22:16:10] [系統] HP -10                   │
│ [22:16:15] [敘事] 圖案突然發光...         │
│ [22:16:30] [玩家] /status                  │
│ [22:16:30] [系統] HP: 60/100, SAN: 75/100 │
│ ...                                        │
│                                            │
│ [↑↓] 捲動 | [ESC] 關閉                    │
└────────────────────────────────────────────┘
```

### 顏色標記規範

| 日誌類型 | 顏色 | LipGloss Style |
|---------|------|----------------|
| 敘事 | 白色 | `lipgloss.Color("15")` |
| 玩家輸入 | 青色 | `lipgloss.Color("6")` |
| 選項選擇 | 黃色 | `lipgloss.Color("3")` |
| 系統訊息 | 灰色 | `lipgloss.Color("8")` |

### 環形緩衝區實作

```go
type RingBuffer struct {
    buffer []LogEntry
    head   int
    tail   int
    size   int
    cap    int
}

func NewRingBuffer(capacity int) *RingBuffer {
    return &RingBuffer{
        buffer: make([]LogEntry, capacity),
        cap:    capacity,
    }
}

func (r *RingBuffer) Push(entry LogEntry) {
    r.buffer[r.tail] = entry
    r.tail = (r.tail + 1) % r.cap

    if r.size < r.cap {
        r.size++
    } else {
        r.head = (r.head + 1) % r.cap
    }
}

func (r *RingBuffer) GetLast(n int) []LogEntry {
    if n > r.size {
        n = r.size
    }

    result := make([]LogEntry, n)
    for i := 0; i < n; i++ {
        idx := (r.tail - n + i + r.cap) % r.cap
        result[i] = r.buffer[idx]
    }
    return result
}
```

### 自動記錄整合點

**GameEngine 整合：**
- 每次 LLM 回應 → LogNarrative
- 每次玩家輸入 → LogPlayerInput
- 每次選擇選項 → LogOptionChoice
- 每次系統指令回應 → LogSystem

**EventBus 整合：**
- 訂閱 NarrativeEvent → 記錄敘事
- 訂閱 PlayerActionEvent → 記錄玩家動作
- 訂閱 SystemEvent → 記錄系統訊息

### 測試策略

**單元測試：**
- 環形緩衝區正確性（push/get）
- 日誌記錄邏輯
- 時間戳記格式

**整合測試：**
- 完整遊戲流程自動記錄
- 存檔/讀檔包含日誌
- UI 捲動與導航

**性能測試：**
- 記憶體使用（1000+ 條目）
- 捲動流暢度
- 序列化大小

### References

- [Source: /home/pollux/Desktop/nightmare-assault/docs/epics.md#Epic-5]
- [Related: Story 5.1 - 存檔資料結構（SaveData 包含 LogEntries）]
- [Related: Story 2.6 - 基礎斜線指令（/log 指令）]

## File List

- `internal/game/log.go` (NEW) - Log data structures and ring buffer
- `internal/game/log_test.go` (NEW) - Unit tests for log structures
- `internal/game/log_performance_test.go` (NEW) - Performance and capacity tests
- `internal/tui/views/log_view.go` (NEW) - BubbleTea log view UI component
- `internal/tui/views/log_view_test.go` (NEW) - Log view UI tests
- `internal/game/commands/log.go` (NEW) - /log command parser
- `internal/game/commands/log_test.go` (NEW) - Command parser tests
- `internal/game/save/schema.go` (MODIFIED) - Added LogEntry and LogType
- `internal/game/save/log_integration_test.go` (NEW) - Save/load integration tests

## Change Log

- 2025-12-11: Story implementation complete
  - RingBuffer with 1000 capacity
  - GameLog manager with thread-safe operations
  - LogView UI with viewport scrolling
  - /log command with optional count parameter
  - SaveData extended with LogEntries field
  - All tests passing (51 unit + integration + performance tests)
  - Performance: 136.5 ns/op for AddEntry, 5.4 ns/op for RingBufferPush
  - Memory usage: ~128 KB for 1000 entries (far under 10 MB limit)
  - Concurrent access tested and verified safe

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- ✅ Task 1 完成：Log data structures (LogEntry, LogType, RingBuffer, GameLog)
- ✅ Task 2 完成：GameLog manager with AddEntry and GetRecentEntries
- ✅ Task 3 完成：BubbleTea LogView UI with viewport and color coding
- ✅ Task 4 完成：/log command parser with default (10) and max (100) support
- ✅ Task 5 完成：SaveData integration with LogEntries field
- ✅ Task 6 完成：Performance tests showing excellent metrics
  - Add 1000 entries: 189µs (target: <10ms) ✓
  - Retrieve 100 entries: 4µs (target: <1ms) ✓
  - Memory: 128KB (target: <10MB) ✓
  - Ring buffer maintains exact capacity through overflow ✓
  - Thread-safe concurrent access verified ✓
