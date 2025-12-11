# Story 3.5: 提示系統

Status: done

## Story

As a **玩家**,
I want **花費 SAN 獲得提示**,
So that **在卡關時有求助選項**.

## Acceptance Criteria

### AC1: 提示指令基礎功能

**Given** 玩家在遊戲中
**When** 輸入 `/hint`
**Then** 顯示確認訊息：「花費 10 SAN 獲得提示？(y/n)」
**And** 顯示當前 SAN 值
**And** 等待玩家確認

**Given** 玩家確認花費
**When** SAN ≥ 10
**Then** 扣除 10 SAN
**And** 顯示與當前情境相關的提示
**And** 更新狀態列的 SAN 顯示

### AC2: SAN 不足處理

**Given** 玩家輸入 `/hint`
**When** SAN < 10
**Then** 顯示訊息：「你的理智不足以清晰思考...（需要 10 SAN）」
**And** 不提供提示
**And** 不扣除 SAN

### AC3: 提示內容生成（模糊暗示）

**Given** 玩家成功請求提示
**When** 系統生成提示內容
**Then** 提示符合以下原則：
  - **模糊暗示**，非直接答案
  - 與當前情境/場景相關
  - 可能暗示未觸發的規則
  - 可能指向錯過的線索
  - 可能建議探索方向
**And** 提示約 20-50 字
**And** 不直接說出「答案」或「規則內容」

**範例:**
- ❌ 錯誤：「你應該在夜晚不要開門」
- ✅ 正確：「夜裡的門背後，似乎藏著不該被窺視的東西」

### AC4: 地獄難度停用

**Given** 難度為「地獄」
**When** 玩家輸入 `/hint`
**Then** 顯示訊息：「在這個噩夢中，沒有人能幫助你」
**And** 功能完全不可用
**And** 不扣除 SAN

### AC5: 提示內容智慧化

**Given** 玩家請求提示
**When** 系統分析當前遊戲狀態
**Then** 提示根據以下優先順序生成：
  1. 如果玩家卡關（同一場景超過 5 回合）→ 提示探索方向
  2. 如果有未發現的關鍵線索 → 提示線索位置
  3. 如果有即將觸發的規則 → 模糊警告
  4. 如果有可用但未使用的道具 → 暗示道具用途
  5. 預設 → 提供當前場景的氛圍提示
**And** 使用 Fast Model 生成提示（< 500ms）

### AC6: 提示使用限制（可選）

**Given** 玩家在同一章節多次使用提示
**When** 使用次數 ≥ 3
**Then** 每次提示的 SAN 消耗增加 5（10 → 15 → 20）
**And** 顯示當前提示消耗量
**And** 下一章節重置消耗量

## Tasks / Subtasks

- [x] Task 1: 建立提示指令基礎 (AC: #1, #2)
  - [x] 建立 `internal/game/commands/hint.go`
  - [x] 實作 `/hint` 指令註冊
  - [x] 實作 SAN 檢查邏輯（≥ 10）
  - [x] 實作確認流程（y/n）
  - [x] 實作 SAN 扣除與狀態更新

- [x] Task 2: 難度控制 (AC: #4)
  - [x] 檢查當前遊戲難度
  - [x] 地獄模式直接返回拒絕訊息
  - [x] 簡單/困難模式正常執行

- [x] Task 3: 提示內容生成核心 (AC: #3, #5)
  - [x] 建立 `internal/game/hint/generator.go`
  - [x] 實作 `GenerateHint(state GameState) string`
  - [x] 實作遊戲狀態分析邏輯：
    - 偵測卡關情況（同場景回合數）
    - 檢測未發現線索
    - 檢測即將觸發的規則
    - 檢測可用道具
  - [x] 實作優先順序邏輯

- [x] Task 4: Fast Model 整合 (AC: #5)
  - [x] 建立提示生成 prompt 模板
  - [x] Prompt 包含：
    - 當前場景描述
    - 玩家最近行動
    - 提示類型（線索/規則/方向/道具）
    - 約束：必須模糊暗示，禁止直接答案
  - [x] 整合 Fast Model API 呼叫
  - [x] 處理生成失敗情況（fallback 提示）

- [x] Task 5: 提示消耗遞增機制（可選） (AC: #6)
  - [x] 追蹤每章節提示使用次數
  - [x] 實作消耗遞增邏輯（10 → 15 → 20）
  - [x] 章節切換時重置計數器
  - [x] 顯示當前消耗量

- [x] Task 6: UI/UX 整合
  - [x] 設計提示顯示樣式（LipGloss）
  - [x] 實作確認對話框元件
  - [x] 顯示 SAN 扣除動畫（可選）
  - [x] 處理提示文字格式化

- [x] Task 7: EventBus 整合
  - [x] 觸發 `HintRequestedEvent`（P2 優先級）
  - [x] 觸發 `SANChangedEvent`（扣除 SAN）
  - [x] 記錄提示使用至遊戲日誌

- [x] Task 8: 單元測試
  - [x] 測試 SAN 檢查邏輯
  - [x] 測試難度控制（地獄模式停用）
  - [x] 測試提示內容生成（模擬不同遊戲狀態）
  - [x] 測試消耗遞增機制
  - [x] 測試 Fast Model 失敗處理

## Dev Notes

### 架構模式與約束

**模組位置:**
- `internal/game/commands/hint.go` - 提示指令
- `internal/game/hint/generator.go` - 提示生成邏輯
- `internal/game/hint/analyzer.go` - 遊戲狀態分析

**提示生成邏輯:**
```go
type HintGenerator struct {
    fastModel  LLMClient
    analyzer   *StateAnalyzer
}

type HintContext struct {
    Type         HintType  // Clue, Rule, Direction, Item
    CurrentScene string
    RecentActions []string
    MissedClues  []string
    UpcomingRules []string
    AvailableItems []string
}

func (hg *HintGenerator) Generate(state GameState) (string, error) {
    // 1. 分析遊戲狀態
    context := hg.analyzer.Analyze(state)

    // 2. 決定提示類型（優先順序）
    hintType := hg.decidePriority(context)

    // 3. 生成 prompt
    prompt := hg.buildPrompt(context, hintType)

    // 4. 呼叫 Fast Model
    hint := hg.fastModel.Generate(prompt)

    return hint, nil
}
```

**Prompt 模板範例:**
```
你是一個恐怖遊戲的神秘嚮導。玩家請求提示。

當前場景：{scene}
玩家最近行動：{actions}
提示類型：{type}

請生成一個 20-50 字的模糊暗示，符合以下原則：
- 不直接說出答案或規則
- 使用隱喻、暗示、氛圍描述
- 保持恐怖遊戲調性
- 引導但不指明方向

{type_specific_context}
```

**Fallback 提示:**
如果 Fast Model 失敗，使用預設提示庫：
```go
var fallbackHints = []string{
    "仔細觀察你周圍的環境，有些細節並非偶然...",
    "回想你聽到的聲音，它們可能在警告你什麼...",
    "有些東西最好不要碰，除非你確定它們的用途...",
}
```

**性能約束:**
- 提示生成時間 < 500ms（Fast Model）
- SAN 檢查與扣除 < 50ms
- 狀態分析 < 100ms

**卡關偵測邏輯:**
```go
func (sa *StateAnalyzer) IsStuck(state GameState) bool {
    // 同一場景超過 5 回合
    return state.CurrentLocation == state.PreviousLocation &&
           state.TurnsSinceLocationChange > 5
}
```

**難度映射:**
| 難度 | 提示可用 | 基礎消耗 | 消耗遞增 |
|------|---------|---------|---------|
| 簡單 | ✓ | 10 SAN | ✓ (可選) |
| 困難 | ✓ | 10 SAN | ✓ (可選) |
| 地獄 | ✗ | N/A | N/A |

### References

- [Source: docs/epics.md#Epic-3]
- [Source: PRD.md - FR022 提示系統]
- [Source: ARCHITECTURE.md - Fast Model]
- [Related: Story 2.4 - HP/SAN 數值系統]
- [Related: Story 3.1 - 潛規則生成]
- [Related: Story 3.4 - 死亡覆盤系統]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Implemented HintCommand with confirmation flow (y/n)
- Hell mode properly denies hints with thematic message (AC4)
- SAN check requires minimum 10 SAN, shows current/required values (AC1, AC2)
- StateAnalyzer determines hint priority (Direction > Clue > Rule > Item > Atmosphere) (AC5)
- HintGenerator with usage tracking per chapter
- Cost increment: 10 → 15 → 20 SAN based on chapter usage (AC6)
- BuildHintPrompt creates LLM prompts with Traditional Chinese output
- Fallback hints for each hint type when LLM fails
- HintState tracks usage and enabled status per difficulty
- All 40+ tests passing including benchmarks
- HintType enum with display names (探索方向, 線索提示, 危險警告, 道具暗示, 氛圍感知)
- HintResult contains text, type, cost, and generated flag

### Files Created/Modified

**New Files:**
- `internal/game/hint/generator.go` - HintType, HintContext, HintResult, StateAnalyzer, Generator, HintState, BuildHintPrompt
- `internal/game/hint/generator_test.go` - 19 tests for hint generation
- `internal/game/commands/hint.go` - HintCommand with confirmation flow
- `internal/game/commands/hint_test.go` - 21 tests for hint command

### Test Coverage

Generator tests:
- TestHintTypeString, TestNewStateAnalyzer, TestStateAnalyzerIsStuck
- TestStateAnalyzerAnalyzePriority, TestStateAnalyzerContextFields
- TestNewGenerator, TestGeneratorGetCurrentCost, TestGeneratorUsageTracking
- TestGeneratorCanAffordHint, TestGeneratorGenerateWithLLMHint
- TestGeneratorGenerateWithFallback, TestGeneratorFallbackHintsExist
- TestBuildHintPrompt, TestBuildHintPromptNoActions
- TestNewHintState, TestHintStateRecordUsage, TestHintStateGetChapterUsageEmpty
- TestHintContextType, TestHintResultFields
- BenchmarkGenerateFallbackHint, BenchmarkBuildHintPrompt

Command tests:
- TestHintCommandName, TestHintCommandHelp
- TestHintCommandHellModeDenied, TestHintCommandHintStateDisabled
- TestHintCommandInsufficientSAN, TestHintCommandConfirmationPrompt
- TestHintCommandConfirmYes, TestHintCommandConfirmNo, TestHintCommandConfirmEmpty
- TestHintCommandCostIncrement, TestHintCommandUsageCountDisplay
- TestHintCommandWithLLMHint, TestHintCommandWithLLMFailure
- TestHintCommandCancelConfirmation, TestHintCommandSetHintState
- TestHintCommandHintTypeDisplay, TestHintCommandDifferentChapters
- TestNewHintCommand
- BenchmarkHintCommandExecute
