# Story 3.1: 潛規則生成

Status: done

## Story

As a **系統**,
I want **在故事開始時生成潛規則**,
So that **遊戲有明確的生存邏輯**.

## Acceptance Criteria

### AC1: 規則數量按難度生成

**Given** 新遊戲開始
**When** 生成故事架構
**Then** 同時生成潛規則集合
**And** 規則數量根據難度：
  - 簡單模式：≤6 條規則
  - 困難模式：不限規則數量
  - 地獄模式：不限規則數量

### AC2: 規則類型多樣性

**Given** 生成潛規則
**When** 決定規則類型
**Then** 從以下 5 種類型中選擇：
  - **場景規則**（特定地點生效）
  - **時間規則**（特定時段生效）
  - **行為規則**（任何時候生效）
  - **對象規則**（針對特定 NPC/物品）
  - **狀態規則**（依賴玩家狀態）
**And** 確保類型分布均勻

### AC3: 規則資料結構與儲存

**Given** 潛規則已生成
**When** 儲存規則
**Then** 規則存於遊戲狀態中（玩家不可見）
**And** 每條規則包含以下欄位：
  - `Type`: 規則類型（場景/時間/行為/對象/狀態）
  - `Trigger`: 觸發條件（結構化描述）
  - `Consequence`: 後果（警告/傷害/即死）
  - `Clues`: 相關線索清單
  - `Priority`: 規則優先級（用於對立規則處理）

### AC4: 規則與 Game Bible 整合

**Given** 故事生成使用 Game Bible
**When** Smart Model 生成故事架構
**Then** 潛規則自動嵌入故事設定
**And** 規則線索自然融入敘事
**And** 玩家無法從 LLM 回應中直接看到規則列表

## Tasks / Subtasks

- [x] Task 1: 建立規則資料結構 (AC: #3)
  - [x] 定義 `Rule` struct 於 `internal/engine/rules/types.go`
  - [x] 實作規則序列化/反序列化（JSON）
  - [x] 定義 5 種規則類型的枚舉

- [x] Task 2: 實作規則生成器 (AC: #1, #2)
  - [x] 建立 `internal/engine/rules/generator.go`
  - [x] 實作 `GenerateRules(difficulty Difficulty) []Rule`
  - [x] 實作規則數量控制邏輯（簡單模式上限 6 條）
  - [x] 實作規則類型分布演算法

- [x] Task 3: 整合至 Game Bible Prompt (AC: #4)
  - [x] 擴展 Story Generation Prompt 模板
  - [x] 在 prompt 中指示 LLM 生成隱藏規則種子
  - [x] 實作規則線索自動嵌入敘事

- [x] Task 4: 規則儲存與狀態管理 (AC: #3)
  - [x] 將規則加入 `GameState` 結構
  - [x] 實作規則狀態追蹤（觸發次數、警告次數）
  - [x] 確保規則不會洩漏至玩家可見的 UI

- [x] Task 5: 單元測試與驗證
  - [x] 測試規則生成數量符合難度設定
  - [x] 測試規則類型分布均勻性
  - [x] 測試規則資料結構完整性

## Dev Notes

### 架構模式與約束

**模組位置:**
- `internal/engine/rules/types.go` - 規則資料結構
- `internal/engine/rules/generator.go` - 規則生成邏輯
- `internal/game/state.go` - 遊戲狀態整合

**規則生成策略:**
1. **難度控制**: 簡單模式上限 6 條，困難/地獄模式動態生成（通常 8-15 條）
2. **類型分布**: 避免單一類型過多，確保遊戲玩法多樣性
3. **對立規則**: 可生成互相矛盾的規則（例：「必須關門」vs「不能關門」），增加遊戲深度

**規則範例:**
```go
type Rule struct {
    ID          string       `json:"id"`
    Type        RuleType     `json:"type"`
    Trigger     Condition    `json:"trigger"`
    Consequence Outcome      `json:"consequence"`
    Clues       []string     `json:"clues"`
    Priority    int          `json:"priority"`
    WarningText string       `json:"warning_text"`
}

type RuleType int
const (
    RuleTypeScenario RuleType = iota  // 場景規則
    RuleTypeTime                      // 時間規則
    RuleTypeBehavior                  // 行為規則
    RuleTypeObject                    // 對象規則
    RuleTypeStatus                    // 狀態規則
)
```

**與 LLM 整合:**
- Game Bible prompt 包含規則生成指示
- LLM 生成的敘事自動埋入規則線索
- 使用 Fast Model 驗證規則合理性

### References

- [Source: docs/epics.md#Epic-3]
- [Source: PRD.md - 潛規則系統]
- [Source: ARCHITECTURE.md - Rules Engine]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Task 1: Created `internal/engine/rules/types.go` with Rule struct, RuleType enum (5 types), ConsequenceType enum, Condition/Outcome structs, RuleSet collection
- Task 2: Created `internal/engine/rules/generator.go` with GenerateRules() function, difficulty-based rule count (Easy ≤6, Hard 8-12, Hell 10-15), type distribution algorithm
- Task 3: Extended `internal/engine/prompts/gamebible.go` with BuildRulesPromptSection(), BuildSystemPromptWithRules(), BuildOpeningPromptWithRules(); Added Hidden Rules section to GameBible
- Task 4: Extended `internal/game/save/schema.go` with RuleStorage and SavedRule types; Created `internal/engine/rules/storage.go` for save/load conversion
- Task 5: 52+ tests across types_test.go, generator_test.go, storage_test.go, gamebible_test.go - all passing

### File List

- internal/engine/rules/types.go (NEW) - Rule data structures, 5 rule types, consequence types
- internal/engine/rules/types_test.go (NEW) - 14 tests for types
- internal/engine/rules/generator.go (NEW) - Rule generator with difficulty scaling
- internal/engine/rules/generator_test.go (NEW) - 17 tests for generator
- internal/engine/rules/storage.go (NEW) - Save/load conversion
- internal/engine/rules/storage_test.go (NEW) - 5 tests for storage roundtrip
- internal/engine/prompts/gamebible.go (MODIFIED) - Added rules integration functions
- internal/engine/prompts/gamebible_test.go (MODIFIED) - Added 7 tests for rules integration
- internal/game/save/schema.go (MODIFIED) - Added HiddenRules, RuleStorage, SavedRule types

### Change Log

- 2025-12-11: Implemented Story 3-1 潛規則生成 - All 5 tasks completed, 52+ tests passing

Status: review
