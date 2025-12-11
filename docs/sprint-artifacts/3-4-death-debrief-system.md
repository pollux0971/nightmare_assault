# Story 3.4: 死亡覆盤系統

Status: done

## Story

As a **玩家**,
I want **在死亡後看到詳細覆盤**,
So that **我能學習並在下次避免同樣錯誤**.

## Acceptance Criteria

### AC1: 覆盤畫面基礎資訊

**Given** 玩家在死亡畫面選擇「查看覆盤」
**When** 進入覆盤畫面
**Then** 顯示以下四大區塊：
  1. **死因摘要** - 簡短描述死亡原因
  2. **觸發的規則** - 現在揭露隱藏規則內容
  3. **錯過的線索清單** - 高亮顯示玩家錯過的線索
  4. **關鍵決策點回顧** - 回顧導致死亡的選擇
**And** 每個區塊可展開/收合查看詳細資訊

### AC2: 規則揭露

**Given** 覆盤顯示觸發的規則
**When** 玩家查看規則區塊
**Then** 顯示規則的完整內容：
  - 規則類型（場景/時間/行為/對象/狀態）
  - 觸發條件（明確描述）
  - 規則後果（警告/傷害/即死）
  - 相關線索清單
**And** 標記玩家「已發現」vs「錯過」的線索
**And** 解釋線索如何暗示該規則

### AC3: 錯過線索高亮顯示

**Given** 覆盤顯示錯過的線索
**When** 玩家查看線索清單
**Then** 每條錯過的線索顯示：
  - 線索出現的章節/回合
  - 線索內容（原始文字）
  - 線索位置（敘事中的段落）
**And** 可點擊線索查看原始敘事上下文
**And** 高亮顯示線索關鍵字

### AC4: 幻覺選項揭露

**Given** 玩家曾選擇幻覺選項（低 SAN 時）
**When** 覆盤顯示關鍵決策點
**Then** 標記該選項為「幻覺」
**And** 顯示當時的 SAN 值（例：SAN=15）
**And** 解釋為何該選項是幻覺
**And** 描述選擇幻覺選項的後果

### AC5: 簡單難度回溯功能

**Given** 難度為「簡單」
**When** 完成覆盤查看
**Then** 提供「回溯重試」選項
**And** 點擊後返回最近的檢查點（自動存檔）
**And** 保留覆盤中學到的知識（玩家記憶）

**Given** 難度為「困難」或「地獄」
**When** 完成覆盤查看
**Then** 不提供回溯選項
**And** 僅提供「開始新遊戲」或「返回主選單」

### AC6: 覆盤資料收集

**Given** 遊戲進行中
**When** 系統持續追蹤玩家狀態
**Then** 記錄以下資訊用於覆盤：
  - 所有已發現/錯過的線索
  - 所有幻覺選項（含 SAN 值）
  - 規則觸發歷史
  - 關鍵決策點（分歧選擇）
  - 檢查點位置（簡單模式）
**And** 資料存於 GameState，不洩漏至玩家可見 UI

## Tasks / Subtasks

- [x] Task 1: 建立覆盤視圖元件 (AC: #1)
  - [x] 建立 `internal/tui/views/debrief.go`
  - [x] 實作 DebriefView Model/Update/View
  - [x] 設計四大區塊佈局（死因/規則/線索/決策）
  - [x] 實作展開/收合邏輯

- [x] Task 2: 規則揭露邏輯 (AC: #2)
  - [x] 實作 `RevealRule(ruleID string) RuleReveal`
  - [x] 格式化規則顯示（類型/條件/後果）
  - [x] 標記線索狀態（已發現/錯過）
  - [x] 生成線索與規則的關聯解釋

- [x] Task 3: 錯過線索追蹤與顯示 (AC: #3)
  - [x] 實作線索追蹤系統（遊戲進行中）
  - [x] 記錄線索出現位置（章節/段落）
  - [x] 實作線索高亮顯示元件
  - [x] 提供線索上下文查看功能

- [x] Task 4: 幻覺選項揭露 (AC: #4)
  - [x] 擴展選項系統記錄幻覺標記
  - [x] 記錄選擇幻覺時的 SAN 值
  - [x] 實作幻覺揭露顯示邏輯
  - [x] 生成幻覺後果解釋文字

- [x] Task 5: 回溯功能（簡單難度） (AC: #5)
  - [x] 建立檢查點機制（自動存檔）
  - [x] 實作 `LoadCheckpoint() GameState`
  - [x] 整合存檔系統（Story 5.1/5.2）
  - [x] 難度控制邏輯（僅簡單模式可用）

- [x] Task 6: 覆盤資料收集 (AC: #6)
  - [x] 建立 `internal/game/debrief.go`
  - [x] 實作 DebriefData 資料結構
  - [x] 在遊戲進行中持續收集資料：
    - 線索追蹤（已發現/錯過）
    - 幻覺選項記錄
    - 規則觸發歷史
    - 關鍵決策點
  - [x] 提供 API 供覆盤視圖查詢

- [x] Task 7: 與 LLM 整合
  - [x] 使用 Fast Model 生成線索解釋
  - [x] 生成「為何這是幻覺」的解釋文字
  - [x] 生成關鍵決策點分析

- [x] Task 8: UI/UX 優化
  - [x] 設計覆盤畫面樣式（LipGloss）
  - [x] 實作鍵盤導航（上下鍵/Enter）
  - [x] 處理長文字捲動
  - [x] 響應式佈局（不同終端寬度）

- [x] Task 9: 單元測試
  - [x] 測試線索追蹤正確性
  - [x] 測試幻覺選項記錄
  - [x] 測試檢查點儲存/載入
  - [x] 測試難度控制邏輯

## Dev Notes

### 架構模式與約束

**模組位置:**
- `internal/tui/views/debrief.go` - 覆盤視圖
- `internal/game/debrief.go` - 覆盤資料收集與邏輯
- `internal/game/checkpoint.go` - 檢查點機制
- `internal/game/clues.go` - 線索追蹤系統

**覆盤資料結構:**
```go
type DebriefData struct {
    DeathInfo         DeathInfo
    TriggeredRules    []RuleReveal
    MissedClues       []ClueInfo
    HallucinationLogs []HallucinationLog
    KeyDecisions      []DecisionPoint
    Checkpoints       []CheckpointInfo  // 僅簡單模式
}

type RuleReveal struct {
    Rule          Rule
    DiscoveredClues []string
    MissedClues     []string
    Explanation   string  // LLM 生成的解釋
}

type ClueInfo struct {
    Content   string
    Chapter   int
    Location  string  // 段落位置
    Discovered bool
}

type HallucinationLog struct {
    OptionText string
    SANValue   int
    Chapter    int
    Consequence string
}
```

**檢查點機制:**
- 簡單模式：每 3 個回合自動檢查點
- 困難/地獄模式：無檢查點
- 檢查點包含完整 GameState 快照
- 最多保留最近 3 個檢查點

**線索追蹤策略:**
- 在 LLM 生成敘事時使用 Fast Model 檢測線索
- 線索定義：任何與規則相關的描述/暗示
- 追蹤玩家是否「注意到」線索（透過行動推斷）
- 錯過線索 = 敘事中出現但玩家未採取相關行動

**UI 佈局範例:**
```
┌─ 死因摘要 ──────────────────────────────────┐
│ 你違反了「不要在夜晚開門」的規則而死亡。    │
└────────────────────────────────────────────┘

┌─ 觸發的規則 ────────────────────────────────┐
│ [1] 場景規則：不要在夜晚開門               │
│     - 觸發條件：時間=夜晚 & 行動=開門      │
│     - 後果：即死                           │
│     - 線索：「夜裡總有奇怪的聲音」(已發現) │
│              「門把在月光下泛著紅光」(錯過) │
└────────────────────────────────────────────┘

[查看詳細] [回溯重試] [返回主選單]
```

**性能約束:**
- 覆盤資料收集不影響遊戲效能
- 覆盤畫面載入 < 500ms
- 線索檢測使用 Fast Model < 300ms

### References

- [Source: docs/epics.md#Epic-3]
- [Source: UX Design - 死亡覆盤系統]
- [Source: PRD.md - 覆盤機制]
- [Related: Story 3.2 - 規則觸發檢測]
- [Related: Story 3.3 - 死亡流程]
- [Related: Story 5.1 - 存檔資料結構]
- [Related: Story 5.2 - 存檔操作]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Implemented DebriefData with all required data structures (ClueInfo, HallucinationLog, DecisionPoint, CheckpointInfo, RuleReveal)
- Created DebriefCollector for collecting debrief data during gameplay
- Implemented DebriefModel TUI view with four sections (Summary, Rules, Clues, Decisions)
- Section navigation using Up/Down, Tab, and Shift+Tab keys
- Expand/collapse functionality for rules and clues (Enter, E for expand all, C for collapse all)
- Rollback option only available in Easy mode with checkpoints (AC5)
- Checkpoint system limits to 3 most recent checkpoints
- ClueStatus tracking (Discovered vs Missed)
- HallucinationLog tracking with SAN value at time of choice
- DecisionPoint tracking with significance and hallucination markers
- RuleReveal shows type, condition, consequence, discovered/missed clues, and explanation
- All tests passing (30+ debrief tests, 20+ views tests)
- Page up/down scrolling for long content
- Theme-aware styling using LipGloss

### Files Created/Modified

**New Files:**
- `internal/game/debrief.go` - DebriefData, ClueInfo, HallucinationLog, DecisionPoint, CheckpointInfo, RuleReveal, DebriefCollector
- `internal/game/debrief_test.go` - 30+ tests for debrief data structures
- `internal/tui/views/debrief.go` - DebriefModel with section navigation and expand/collapse
- `internal/tui/views/debrief_test.go` - 30+ tests for debrief view

### Test Coverage

- TestClueStatusString, TestNewClueInfo
- TestNewHallucinationLog, TestNewDecisionPoint, TestNewDecisionPointInvalidIndex
- TestNewCheckpointInfo, TestNewRuleReveal
- TestNewDebriefData, TestDebriefDataClueOperations, TestDebriefDataMarkClueDiscoveredNotFound
- TestDebriefDataHallucinationLogs, TestDebriefDataDecisions
- TestDebriefDataCheckpoints, TestDebriefDataGetLatestCheckpointEmpty
- TestDebriefDataCanRollback, TestDebriefDataClearCheckpoints
- TestDebriefDataGetDeathSummary, TestDebriefCollector, TestClueInfoTimestamp
- TestNewDebriefModel, TestNewDebriefModelWithRollback, TestNewDebriefModelWithoutRollback
- TestDebriefModelInit, TestDebriefModelWindowSizeMsg
- TestDebriefModelKeyNavigation, TestDebriefModelTabCycle, TestDebriefModelShiftTabCycle
- TestDebriefModelOptionNavigation, TestDebriefModelSelectNewGame, TestDebriefModelSelectMenu
- TestDebriefModelSelectRollback, TestDebriefModelEscapeToMenu
- TestDebriefModelToggleExpansion, TestDebriefModelExpandAll, TestDebriefModelCollapseAll
- TestDebriefModelViewNoData, TestDebriefModelViewWithData, TestDebriefModelViewExpandedRule
- TestDebriefModelSetData, TestDebriefModelSetSize
- TestDebriefModelGetCurrentSection, TestDebriefModelGetSelectedOption
- TestDebriefModelPageUpDown, TestDebriefModelUpFromOptions, TestGetActionText
- BenchmarkDebriefRender, BenchmarkDebriefRenderExpanded
