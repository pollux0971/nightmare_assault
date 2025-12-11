# Story 4.4: 隊友狀態追蹤

Status: review

## Story

As a 玩家,
I want 追蹤隊友的位置和狀態,
so that 我知道隊友是否安全.

## Acceptance Criteria

1. **Given** 遊戲進行中
   **When** 隊友位置改變
   **Then** 更新隊友狀態
   **And** 狀態列顯示簡要資訊（寬敞模式 ≥120 寬度）

2. **Given** 隊友與玩家分散
   **When** 隊友獨自行動
   **Then** 定期收到隊友訊息（如果通訊可用）
   **And** 訊息可能包含他們發現的線索

3. **Given** 隊友 HP 降低
   **When** HP < 30
   **Then** 隊友行為受影響（移動緩慢、反應遲鈍）
   **And** 敘事中反映受傷狀態

4. **Given** 狀態列顯示模式
   **When** 終端寬度不同
   **Then** 適應性顯示：
   - < 80: 隱藏隊友資訊
   - 80-99: 僅顯示隊友數量與總 HP
   - 100-119: 顯示姓名與 HP 百分比
   - ≥120: 完整顯示姓名/HP/位置/狀態圖示

5. **Given** 隊友狀態變化
   **When** HP/位置/情緒改變
   **Then** EventBus 發布 TeammateStatusChanged 事件
   **And** UI 即時更新

## Tasks / Subtasks

- [x] 擴展隊友狀態系統 (AC: #1, #3)
  - [x] 擴展 `internal/game/npc/` 模組
  - [x] 新增 Location、EmotionalState, InjuryLevel 欄位
  - [x] 實作狀態更新邏輯

- [ ] 實作位置追蹤機制 (AC: #1, #2)
  - [ ] 定義位置資料結構 (場景名稱、相對玩家距離)
  - [ ] 位置同步機制 (跟隨/分散)
  - [ ] 分散時的訊息系統

- [ ] 實作分散狀態訊息系統 (AC: #2)
  - [ ] 定期訊息生成 (每 2-3 回合)
  - [ ] 訊息內容生成 prompt (發現/擔憂/狀態報告)
  - [ ] 通訊可用性檢查 (某些場景可能無法通訊)
  - [ ] 線索透過訊息傳遞

- [x] 實作 HP 影響系統 (AC: #3)
  - [x] HP < 30: 移動速度降低、反應遲鈍
  - [x] HP < 15: 無法獨立行動、需要協助
  - [x] HP = 0: 死亡流程 (整合 Story 4.3)
  - [ ] 敘事描述整合受傷狀態 - Needs LLM integration

- [ ] 開發響應式狀態列 UI (AC: #4)
  - [ ] 實作 `internal/tui/components/teammate_status_bar.go`
  - [ ] 四級寬度適應 (<80/80-99/100-119/≥120)
  - [ ] 狀態圖示設計 (健康/受傷/瀕死/死亡)
  - [ ] 位置簡寫顯示

- [ ] 整合 EventBus (AC: #5)
  - [ ] 定義 TeammateStatusChanged 事件
  - [ ] 發布位置/HP/情緒變化事件
  - [ ] UI 訂閱並即時更新
  - [ ] 優先級設定 (P1)

- [ ] 擴展 /team 指令 (整合 Story 4.2)
  - [ ] 顯示詳細狀態: 位置/HP/情緒/傷勢
  - [ ] 分散隊友的最後通訊時間
  - [ ] 死亡隊友標記與遺物提示

- [x] 情緒狀態系統 - Core implementation
  - [x] 情緒類型: Calm/Anxious/Panicked/Relieved/Grieving
  - [ ] 情緒影響 SAN 恢復效率 - Needs game state integration
  - [ ] 情緒隨事件變化 - Needs EventBus integration

- [x] 單元測試 - Core tests complete
  - [x] 測試位置更新與同步
  - [x] 驗證 HP < 30 行為改變
  - [ ] 測試響應式 UI 在不同寬度下的顯示 - Needs TUI components
  - [ ] 驗證 EventBus 事件正確發布 - Needs EventBus integration

## Dev Notes

### 架構模式與約束

**擴展 Teammate 結構:**
```go
type Teammate struct {
    // ... 原有欄位 (from Story 4.1)
    Location       Location
    LastSeen       time.Time
    EmotionalState EmotionalState
    InjuryLevel    InjuryLevel
    IsSeparated    bool
    LastMessage    *TeammateMessage
}

type Location struct {
    Scene          string
    DistanceToPlayer int // 0=同場景, 1=相鄰, 2+=遠離
}

type EmotionalState string
const (
    EmotionCalm     EmotionalState = "calm"
    EmotionAnxious  EmotionalState = "anxious"
    EmotionPanicked EmotionalState = "panicked"
    EmotionRelieved EmotionalState = "relieved"
    EmotionGrieving EmotionalState = "grieving"
)

type InjuryLevel int
const (
    InjuryNone   InjuryLevel = 0  // HP 100-70
    InjuryMinor  InjuryLevel = 1  // HP 69-30
    InjurySerious InjuryLevel = 2  // HP 29-15
    InjuryCritical InjuryLevel = 3 // HP 14-1
)

type TeammateMessage struct {
    Content   string
    Timestamp time.Time
    ClueID    *string // 如果訊息包含線索
}
```

**狀態列響應式設計:**
```
寬度 < 80:
[HP: 100 | SAN: 85]

寬度 80-99:
[HP: 100 | SAN: 85 | 隊友: 2/3 (HP: 140/200)]

寬度 100-119:
[HP: 100 | SAN: 85 | 小李: 80% | 小王: 30%⚠️]

寬度 ≥120:
[HP: 100 | SAN: 85 | 小李: 80%💚 (廚房) | 小王: 30%❤️ (大廳-受傷)]
```

**狀態圖示設計:**
- 💚 (綠心): HP > 70
- 💛 (黃心): HP 30-70
- ❤️ (紅心): HP < 30
- 💀 (骷髏): 已死亡
- 📍 (圖釘): 位置標記

**分散訊息生成機制:**
```go
func GenerateSeparatedMessage(teammate *Teammate, turns int) *TeammateMessage {
    // 每 2-3 回合生成一次
    if turns % (2 + rand.Intn(2)) != 0 {
        return nil
    }

    // 檢查通訊可用性
    if !IsCommAvailable(teammate.Location) {
        return nil
    }

    // 使用 Fast Model 生成訊息
    messageTypes := []string{
        "discovery",  // "我在二樓發現了一本日記..."
        "concern",    // "這裡很安靜，太安靜了..."
        "status",     // "我還好，繼續搜索中"
        "clue",       // "牆上有個奇怪的符號..."
    }

    // 根據 teammate 性格選擇訊息類型機率
}
```

**HP 影響行為邏輯:**
```go
func GetBehaviorModifier(hp int) BehaviorModifier {
    switch {
    case hp >= 70:
        return BehaviorModifier{
            MoveSpeed: 1.0,
            Reaction: 1.0,
            Description: "",
        }
    case hp >= 30:
        return BehaviorModifier{
            MoveSpeed: 0.8,
            Reaction: 0.9,
            Description: "略顯疲憊",
        }
    case hp >= 15:
        return BehaviorModifier{
            MoveSpeed: 0.5,
            Reaction: 0.6,
            Description: "步履蹣跚，表情痛苦",
        }
    default:
        return BehaviorModifier{
            MoveSpeed: 0.0,
            Reaction: 0.3,
            Description: "無法自行移動，需要攙扶",
        }
    }
}
```

**EventBus 整合:**
```go
type TeammateStatusChangedEvent struct {
    TeammateID string
    ChangeType string // "location", "hp", "emotion"
    OldValue   interface{}
    NewValue   interface{}
}

// 發布範例
eventBus.Publish(Event{
    Type: EventTeammateStatusChanged,
    Priority: PriorityP1,
    Data: TeammateStatusChangedEvent{
        TeammateID: "teammate_1",
        ChangeType: "hp",
        OldValue: 80,
        NewValue: 25,
    },
})
```

**通訊系統規則:**
```
通訊可用條件:
1. 距離 ≤ 2 (同場景或相鄰)
2. 無訊號干擾場景 (地下室/密閉空間)
3. 隊友未處於 Panicked 狀態
4. 隊友 HP > 0 (存活)

通訊失敗處理:
- 顯示「無法聯繫 {name}」
- 增加玩家焦慮 (輕微 SAN -2)
- 提供前往尋找選項
```

**性能約束:**
- 狀態更新頻率: 每回合一次 (非即時)
- UI 更新響應: < 100ms (本地計算)
- 分散訊息生成 (Fast Model): < 500ms
- 狀態列記憶體: < 1KB (精簡資料)

**整合點:**
- 與 Story 4.1 Teammate 結構整合
- 與 Story 4.2 對話系統整合 (訊息作為對話形式)
- 與 Story 4.3 死亡機制整合 (HP=0 觸發)
- 與 Epic 2 響應式佈局整合
- 與 EventBus 整合 (狀態變化事件)

**邊界情況處理:**
- 所有隊友死亡: 隱藏隊友狀態列
- 窄終端 (<80): 完全隱藏隊友資訊，僅保留 /team 指令
- 通訊完全中斷: 提供「上次已知位置」資訊
- 多隊友同時分散: 訊息排隊顯示，避免刷屏

### References

- [Source: docs/epics.md#Epic-4]
- [Related: ARCHITECTURE.md - 響應式佈局]
- [Related: ARCHITECTURE.md - EventBus System]
- [Related: docs/ux-design-specification.md - 狀態追蹤]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development

**Implementation Completed:**

1. **Status Tracking Extensions** (`internal/game/npc/status.go`)
   - Location struct: Scene name + DistanceToPlayer (0=same, 1=adjacent, 2+=far)
   - EmotionalState enum: Calm, Anxious, Panicked, Relieved, Grieving (5 states)
   - InjuryLevel enum: None (70-100 HP), Minor (30-69), Serious (15-29), Critical (1-14)
   - TeammateMessage struct: Content, Timestamp, optional ClueID
   - BehaviorModifier struct: MoveSpeed, Reaction multipliers, Description

2. **Teammate Struct Extensions** (`internal/game/npc/teammate.go`)
   - Added LastSeen (time.Time) - tracks last contact
   - Added EmotionalState field - initialized to EmotionCalm
   - Added InjuryLevel field - initialized to InjuryNone
   - Added IsSeparated bool - tracks if teammate is away from player
   - Added LastMessage *TeammateMessage - stores last communication

3. **HP Impact System**
   - CalculateInjuryLevel function: Auto-determines injury level from HP
   - GetBehaviorModifier function: Returns movement/reaction modifiers
     * HP >= 70: 100% speed, 100% reaction, no description
     * HP 30-69: 80% speed, 90% reaction, "略顯疲憊"
     * HP 15-29: 50% speed, 60% reaction, "步履蹣跚，表情痛苦"
     * HP < 15: 0% speed, 30% reaction, "無法自行移動，需要攙扶"

4. **Status Update Methods**
   - UpdateLocation(Location): Updates location string + IsSeparated flag + LastSeen
   - UpdateHP(int): Clamps HP [0,100], updates InjuryLevel, updates Status.Condition
     * HP = 0 → Status.Alive = false, Condition = "dead"
     * HP < 15 → Condition = "critical"
     * HP < 30 → Condition = "injured"
     * HP >= 30 → Condition = "healthy"
   - UpdateEmotionalState(EmotionalState): Updates emotional state
   - GetBehavior(): Returns current BehaviorModifier based on HP

5. **Comprehensive Testing** (`internal/game/npc/status_test.go`)
   - TestLocation: Validates Location struct
   - TestEmotionalState: Validates 5 emotional states exist
   - TestInjuryLevel: Validates 4 injury levels
   - TestCalculateInjuryLevel: 9 test cases covering all HP ranges
   - TestBehaviorModifier: Validates behavior modifiers at different HP levels
   - TestUpdateTeammateLocation: Validates location updates and IsSeparated flag
   - TestUpdateTeammateHP: Validates HP clamping, injury calculation, status updates, death at HP=0
   - TestUpdateEmotionalState: Validates emotional state changes
   - TestIsSeparated: Validates separation detection
   - All 10 status tracking tests passing (100%)

**Files Created:**
- `internal/game/npc/status.go` (165 lines)
- `internal/game/npc/status_test.go` (205 lines)

**Files Modified:**
- `internal/game/npc/teammate.go` (+10 lines) - Extended Teammate struct

**Total Tests:** 10 status tracking tests + all previous tests = 39 tests total, 100% passing

**Remaining Work (Blocked on dependencies):**
- Location tracking UI / status bar (needs TUI components)
- Separated message system (needs LLM Fast Model integration)
- Communication availability system (needs game state/scene system)
- EventBus integration (needs EventBus implementation)
- Responsive status bar (needs TUI framework)
- Narrative integration of injury states (needs LLM integration)
- SAN recovery efficiency based on emotion (needs game state)
- Emotion changes based on events (needs EventBus)

**Ready for Review:**
Core status tracking system complete with all data structures, calculations, and update methods.
AC#1, #3 core mechanics implemented. AC#2, #4, #5 require game state/TUI/EventBus integration.


---

## Code Review Record

**Date**: 2025-12-11
**Review Type**: Adversarial Code Review (Epic 4 - All Stories)
**Reviewer**: Claude Sonnet 4.5 (Code Review Agent)

### Issues Found & Fixed

**✅ HIGH: HP Condition Logic Misalignment**
- **File**: `status.go:116-133`
- **Issue**: HP range 30-69 was marked as "healthy" instead of "injured", misaligned with InjuryLevel thresholds
- **Fix**: Changed condition from "healthy" to "injured" for HP < 70
- **Impact**: Consistent status display aligned with InjuryLevel enum

**✅ HIGH: Missing LastSeen Update**
- **File**: `status.go:102`
- **Issue**: UpdateHP() didn't update LastSeen timestamp
- **Fix**: Added `t.LastSeen = time.Now()` in UpdateHP()
- **Impact**: Accurate tracking of when teammate status changed

**✅ MEDIUM: Dead Code Removal**
- **File**: `status.go:146-159`
- **Issue**: TeammateExtended struct was obsolete (fields already in main Teammate struct)
- **Fix**: Removed entire TeammateExtended struct and associated comments
- **Impact**: Cleaner codebase, removed ~15 lines of dead code

**Review Summary**: 3 issues fixed, code cleaned up
**Detailed Report**: See `docs/sprint-artifacts/epic-4-code-review-fixes.md`

