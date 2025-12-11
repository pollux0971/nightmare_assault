# Story 4.2: 隊友對話與互動

Status: in-progress

## Story

As a 玩家,
I want 與隊友對話並獲得線索,
so that 隊友是有用的資訊來源.

## Acceptance Criteria

1. **Given** 玩家在遊戲中有隊友
   **When** 輸入 `/team`
   **Then** 顯示所有隊友狀態：
   - 姓名、HP、位置
   - 攜帶物品
   - 當前情緒狀態

2. **Given** 玩家選擇與隊友交談
   **When** 進入對話模式
   **Then** 隊友回應符合其性格
   **And** 可能透露線索（隱晦方式）
   **And** 深度交流可恢復 5-15 SAN

3. **Given** 等待 LLM 回應超過 2 秒
   **When** 觸發延遲掩蓋
   **Then** 顯示隊友插話（Fast Model 生成）
   **And** 插話內容符合當前情境

4. **Given** 隊友對話包含線索
   **When** 隊友回應
   **Then** 線索以隱晦方式呈現（暗示/擔憂/觀察）
   **And** 不直接告知規則

5. **Given** 三層延遲防護機制
   **When** 等待時間不同
   **Then** 響應如下：
   - 0-300ms: 無提示
   - 300ms-2s: Spinner 等待動畫
   - 2s+: 隊友插話掩蓋延遲

## Tasks / Subtasks

- [x] 建立 /team 指令 (AC: #1)
  - [x] 實作指令處理器 `internal/game/commands/team.go`
  - [x] 設計隊友狀態 UI 元件
  - [x] 顯示姓名/HP/位置/物品/情緒

- [x] 實作對話系統 (AC: #2) - Core structure complete, LLM integration pending
  - [x] 建立 `internal/game/npc/dialogue.go`
  - [ ] 對話 prompt 整合 Character Sheet - Needs LLM integration
  - [ ] 實作線索透露機制（隱晦暗示） - Needs LLM integration
  - [ ] SAN 恢復邏輯 (深度交流 +5 至 +15 SAN) - Needs game state integration

- [ ] 實作三層延遲防護 (AC: #3, #5)
  - [ ] 0-300ms: 無提示（直接等待）
  - [ ] 300ms-2s: Spinner 動畫
  - [ ] 2s+: Fast Model 隊友插話生成
  - [ ] 整合至 EventBus 延遲檢測

- [ ] Fast Model 插話生成器 (AC: #3)
  - [ ] 創建插話 prompt 模板
  - [ ] 確保插話符合隊友性格
  - [ ] 插話內容與當前情境相關
  - [ ] 響應時間 < 500ms (NFR01)

- [ ] 線索系統整合 (AC: #4)
  - [ ] 隊友可透露的線索類型定義
  - [ ] 隱晦表達策略（擔憂/觀察/回憶）
  - [ ] 線索記錄至玩家 Clues 列表

- [ ] UI 元件開發
  - [ ] 對話框元件 `internal/tui/components/dialogue_box.go`
  - [ ] Spinner 動畫元件
  - [ ] 隊友狀態列表元件

- [x] 單元測試 - Core tests complete
  - [x] 測試 `/team` 指令輸出格式
  - [x] 驗證對話系統數據結構
  - [ ] 驗證對話性格一致性 - Needs LLM integration
  - [ ] 測試延遲防護觸發時機 - Needs EventBus
  - [ ] 驗證 SAN 恢復數值範圍 - Needs game state

## Dev Notes

### 架構模式與約束

**對話系統架構:**
```go
type DialogueSystem struct {
    teammates   []*Teammate
    dialogueLog []DialogueEntry
    smartModel  LLMProvider
    fastModel   LLMProvider
}

type DialogueEntry struct {
    Timestamp   time.Time
    Speaker     string // "Player" or Teammate.Name
    Content     string
    ClueRevealed *Clue
}

type ClueRevelation struct {
    ClueID      string
    RevelationType string // "hint", "worry", "observation", "memory"
    Subtlety    int // 1-10, 越高越隱晦
}
```

**三層延遲防護實作:**
```go
// 延遲檢測流程
1. 發送 LLM 請求時啟動計時器
2. 0-300ms: 無動作
3. 300ms: 觸發 Spinner 顯示
4. 2000ms:
   - 取消 Spinner
   - 發送 Fast Model 請求生成插話
   - 顯示插話內容
5. Smart Model 回應到達後:
   - 停止插話
   - 顯示完整回應
```

**Fast Model 插話 Prompt 範例:**
```
Context: {current_scene}
Teammate: {teammate_name} ({archetype})
Personality: {personality_traits}
Situation: 正在等待某事發生，隊友會說一句符合其性格的話

要求:
- 一句話 (< 20 字)
- 符合性格
- 與當前場景相關
- 不透露關鍵資訊

範例:
- Logic 型: "這裡的結構不太對勁..."
- Victim 型: "我們真的要繼續嗎？"
- Intuition 型: "我有種不好的預感..."
```

**SAN 恢復機制:**
- 淺層對話 (1-2 回合): +5 SAN
- 深度交流 (3+ 回合): +10 SAN
- 情感共鳴 (特殊劇情): +15 SAN
- 冷卻時間: 每個隊友 5 回合只能觸發一次

**線索透露策略:**
1. **擔憂型**: "我總覺得晚上不該出門..."（暗示時間規則）
2. **觀察型**: "那個房間的門把好像被人從裡面鎖上過..."（暗示場景規則）
3. **回憶型**: "我記得爺爺說過，鏡子不能對著床..."（暗示行為規則）

**性能約束:**
- Fast Model 插話生成: < 500ms (NFR01)
- /team 指令響應: < 100ms (本地計算)
- 對話 Character Sheet 注入: < 200 tokens

**整合點:**
- 與 Story 4.1 的 Teammate 結構整合
- 與 Epic 3 的線索系統整合
- 與 EventBus 的延遲監測整合
- 為 Story 4.3 死亡機制提供伏筆渠道

### References

- [Source: docs/epics.md#Epic-4]
- [Related: ARCHITECTURE.md - Dual Model Architecture]
- [Related: docs/ux-design-specification.md - 三層延遲防護]
- [Related: NFR01 - 性能需求]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development

**Partial Implementation Completed:**

1. **Dialogue System Core** (`internal/game/npc/dialogue.go`)
   - DialogueEntry struct with timestamp, speaker, content, clue revelation
   - ClueRevelation struct with revelation types (hint, worry, observation, memory) and subtlety (1-10)
   - DialogueSystem with thread-safe operations (AddEntry, GetHistory, GetRecentHistory, ClearHistory)
   - Teammate management (SetTeammates, GetTeammates)
   - Placeholder functions for LLM integration (BuildDialoguePrompt, GenerateTeammateResponse)
   - 6 passing unit tests for dialogue data structures

2. **/team Command** (`internal/game/commands/team.go`)
   - Complete implementation showing all teammate status
   - Displays: Name, Archetype, HP, Condition, Location, Inventory, Emotional State
   - Status icons based on condition (✓ healthy, ⚠ injured, ⚠⚠ critical, 💀 dead)
   - Emotional states vary by archetype and HP level
   - Bilingual output (Traditional Chinese / English)
   - 6 passing unit tests for all scenarios

**Remaining Work (Blocked on dependencies):**
- LLM provider integration for actual dialogue generation
- Character Sheet prompt injection (structure exists, needs LLM calls)
- Clue revelation mechanism (structure exists, needs game state)
- SAN recovery logic (needs game state integration)
- Three-layer latency protection (needs EventBus and LLM provider)
- Fast Model interjection system (needs dual LLM setup)
- TUI components (dialogue box, spinner, team list)

**Files Created:**
- `internal/game/npc/dialogue.go` (115 lines)
- `internal/game/npc/dialogue_test.go` (125 lines)
- `internal/game/commands/team.go` (155 lines)
- `internal/game/commands/team_test.go` (125 lines)

**Total Tests:** 12 tests, 100% passing

**Status:** Core data structures and /team command complete. Story moved to in-progress.
Full dialogue and latency masking features require LLM provider and game state integration (Epic 2 dependencies).


---

## Code Review Record

**Date**: 2025-12-11
**Review Type**: Adversarial Code Review (Epic 4 - All Stories)
**Reviewer**: Claude Sonnet 4.5 (Code Review Agent)

### Issues Found & Fixed

**✅ CRITICAL: Race Condition in DialogueSystem**
- **File**: `dialogue.go:46,64`
- **Issue**: Shallow copy of dialogue entries with pointer fields (ClueRevealed) caused race condition
- **Fix**: Implemented deep copy for ClueRevelation pointers in GetHistory() and GetRecentHistory()
- **Impact**: Thread-safe concurrent access, prevents data corruption

**✅ HIGH: Missing Input Validation**
- **File**: `dialogue.go:64`
- **Issue**: GetRecentHistory() didn't validate negative or zero count
- **Fix**: Added validation `if count <= 0 { return []DialogueEntry{} }`
- **Impact**: Prevents panic on invalid input

**✅ CRITICAL: Missing Edge Case Tests**
- **File**: `dialogue_test.go`
- **Issue**: Tests didn't cover concurrent access, negative counts, deep copy verification
- **Fix**: Added 7 new edge case tests (negative count, deep copy, etc.)
- **Impact**: +7 test cases, comprehensive coverage of edge cases

**✅ HIGH: EmotionalState Field Not Used**
- **File**: `commands/team.go:66`
- **Issue**: Command calculated emotion from HP/archetype instead of using tm.EmotionalState
- **Fix**: Modified getEmotionalState() to check tm.EmotionalState first, added getEmotionalStateText()
- **Impact**: Uses actual teammate state, more accurate emotional display

**✅ HIGH: Missing Nil Validation**
- **File**: `commands/team.go:41`
- **Issue**: /team command didn't check for nil teammates in slice
- **Fix**: Added `if tm == nil { continue }` check
- **Impact**: Prevents panic on corrupted teammate data

**Review Summary**: 5 issues fixed, 7 new tests added (19 total tests)
**Detailed Report**: See `docs/sprint-artifacts/epic-4-code-review-fixes.md`

