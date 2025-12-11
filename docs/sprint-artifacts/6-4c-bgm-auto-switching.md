# Story 6.4c: BGM 自動切換

Status: ready-for-dev

## Story

As a 玩家,
I want BGM 根據場景氛圍自動切換,
so that 音樂與遊戲氛圍完美契合.

## Acceptance Criteria

### AC1: LLM Mood 標記解析

**Given** 遊戲進行中且音訊已安裝
**When** 接收到 LLM 回應（Fast Model）
**Then** 解析回應中的 mood 標記（如 `[MOOD:tension]`）
**And** 支援以下 mood 類型：
  - `exploration`: 探索場景
  - `tension`: 緊張/追逐場景
  - `safe`: 安全區/休息場景
  - `horror`: 恐怖揭露時刻
  - `mystery`: 解謎場景
  - `ending`: 死亡/結局場景
**And** 若未偵測到 mood 標記，使用預設 `exploration`

### AC2: Mood 到 BGM 映射

**Given** 偵測到 mood 標記
**When** 需要切換 BGM
**Then** 根據映射表切換對應 BGM：
  | Mood | BGM 檔案 |
  |------|---------|
  | exploration | ambient_exploration.ogg |
  | tension | tension_chase.ogg |
  | safe | safe_rest.ogg |
  | horror | horror_reveal.ogg |
  | mystery | mystery_puzzle.ogg |
  | ending | ending_death.ogg |
**And** 使用 1-2 秒淡入淡出（crossfade）切換

### AC3: 防止頻繁切換機制

**Given** 當前 BGM 正在播放
**When** 偵測到新的 mood 標記
**Then** 檢查當前 BGM 播放時長
**And** 若播放時長 < 30 秒，忽略切換請求
**And** 若 mood 與當前 BGM 相同，不執行切換
**And** 記錄最後切換時間（用於後續判斷）

### AC4: 場景自動切換邏輯

**Given** 遊戲主循環運作中
**When** 每次 LLM 回應顯示後
**Then** 呼叫 `CheckMoodAndSwitch()` 函數
**And** 根據解析的 mood 決定是否切換 BGM
**And** 切換過程不阻塞主執行緒（使用 goroutine）
**And** 切換失敗時靜默處理（不中斷遊戲）

### AC5: 整合遊戲主循環

**Given** 遊戲主循環（Story 2.5）
**When** 敘事顯示完成
**Then** 自動檢查 mood 標記並切換 BGM
**And** 整合至 `GameLoop` 的事件處理
**And** 不影響現有遊戲邏輯

## Tasks / Subtasks

- [ ] Task 1: 實作 LLM Mood 標記解析 (AC: #1)
  - [ ] Subtask 1.1: 擴展 `internal/engine/response_parser.go`（或新建）
  - [ ] Subtask 1.2: 實作正則表達式解析 `[MOOD:xxx]` 標記
  - [ ] Subtask 1.3: 定義 `MoodType` 列舉（6 種類型）
  - [ ] Subtask 1.4: 實作 `ParseMood(text string) MoodType` 函數
  - [ ] Subtask 1.5: 處理無 mood 標記情境（返回預設）

- [ ] Task 2: 實作 Mood 到 BGM 映射 (AC: #2)
  - [ ] Subtask 2.1: 建立 `internal/audio/mood_mapping.go`
  - [ ] Subtask 2.2: 定義 `MoodToBGM` 映射表
  - [ ] Subtask 2.3: 實作 `GetBGMForMood(mood MoodType) string` 函數
  - [ ] Subtask 2.4: 測試所有 mood 類型的映射正確性

- [ ] Task 3: 實作防止頻繁切換機制 (AC: #3)
  - [ ] Subtask 3.1: 在 `BGMPlayer` 中添加 `lastSwitch time.Time` 欄位
  - [ ] Subtask 3.2: 實作 `CanSwitch(newMood MoodType) bool` 判斷函數
  - [ ] Subtask 3.3: 實作 30 秒最小間隔檢查
  - [ ] Subtask 3.4: 實作相同 mood 檢查（避免重複切換）
  - [ ] Subtask 3.5: 測試防頻繁切換邏輯

- [ ] Task 4: 實作場景自動切換邏輯 (AC: #4)
  - [ ] Subtask 4.1: 建立 `CheckMoodAndSwitch(text string)` 函數
  - [ ] Subtask 4.2: 整合 mood 解析 + BGM 映射 + 防頻繁切換
  - [ ] Subtask 4.3: 使用 goroutine 非阻塞切換
  - [ ] Subtask 4.4: 實作錯誤處理（靜默失敗）
  - [ ] Subtask 4.5: 測試完整自動切換流程

- [ ] Task 5: 整合遊戲主循環 (AC: #5)
  - [ ] Subtask 5.1: 擴展 `internal/tui/views/game_main.go`（或相應檔案）
  - [ ] Subtask 5.2: 在敘事顯示完成後呼叫 `CheckMoodAndSwitch()`
  - [ ] Subtask 5.3: 確保不影響現有遊戲邏輯
  - [ ] Subtask 5.4: 整合測試（完整遊戲流程）
  - [ ] Subtask 5.5: 驗證 BGM 自動切換的沉浸感

## Dev Notes

### 架構模式

- **模組位置**:
  - Mood 解析: `internal/engine/response_parser.go`（擴展）
  - Mood 映射: `internal/audio/mood_mapping.go`
  - 自動切換: `internal/audio/bgm_player.go`（擴展）

- **核心資料結構**:
  ```go
  type MoodType int
  const (
      MoodExploration MoodType = iota
      MoodTension
      MoodSafe
      MoodHorror
      MoodMystery
      MoodEnding
  )

  var MoodToBGM = map[MoodType]string{
      MoodExploration: "ambient_exploration.ogg",
      MoodTension:     "tension_chase.ogg",
      MoodSafe:        "safe_rest.ogg",
      MoodHorror:      "horror_reveal.ogg",
      MoodMystery:     "mystery_puzzle.ogg",
      MoodEnding:      "ending_death.ogg",
  }
  ```

### Mood 標記解析實作

```go
func ParseMood(text string) MoodType {
    re := regexp.MustCompile(`\[MOOD:(\w+)\]`)
    matches := re.FindStringSubmatch(text)

    if len(matches) < 2 {
        return MoodExploration // Default
    }

    switch strings.ToLower(matches[1]) {
    case "tension":
        return MoodTension
    case "safe":
        return MoodSafe
    case "horror":
        return MoodHorror
    case "mystery":
        return MoodMystery
    case "ending":
        return MoodEnding
    default:
        return MoodExploration
    }
}
```

### 防頻繁切換實作

```go
func (p *BGMPlayer) CanSwitch(newMood MoodType) bool {
    p.mu.RLock()
    defer p.mu.RUnlock()

    // Check if same mood
    currentMood := p.getCurrentMood()
    if currentMood == newMood {
        return false
    }

    // Check minimum interval (30 seconds)
    if time.Since(p.lastSwitch) < 30*time.Second {
        return false
    }

    return true
}

func (p *BGMPlayer) SwitchByMood(mood MoodType) error {
    if !p.CanSwitch(mood) {
        return nil // Silently skip
    }

    bgmFile := MoodToBGM[mood]
    if err := p.Crossfade(bgmFile, 2*time.Second); err != nil {
        return err
    }

    p.lastSwitch = time.Now()
    return nil
}
```

### 整合遊戲主循環

```go
// In game_main.go or similar
func (m *GameMainModel) handleNarrativeComplete(text string) {
    // Existing logic: display narrative, handle user input...

    // New: Auto-switch BGM based on mood
    go func() {
        mood := engine.ParseMood(text)
        if err := m.audioManager.BGMPlayer.SwitchByMood(mood); err != nil {
            log.Printf("[WARN] BGM auto-switch failed: %v", err)
        }
    }()
}
```

### LLM Prompt 調整

需要在 LLM prompt 中添加 mood 標記指示（Story 2.2 相關）：

```
在敘事中適當位置添加 mood 標記：
- 探索場景: [MOOD:exploration]
- 緊張追逐: [MOOD:tension]
- 安全區: [MOOD:safe]
- 恐怖揭露: [MOOD:horror]
- 解謎場景: [MOOD:mystery]
- 死亡結局: [MOOD:ending]

範例：
你推開沉重的木門，走進一個黑暗的走廊。[MOOD:exploration]
遠處傳來急促的腳步聲，越來越近！[MOOD:tension]
```

### 技術約束

- **切換延遲**: < 200ms（檢測到 mood 到開始切換）
- **最小間隔**: 30 秒（防止頻繁切換）
- **淡入淡出**: 1-2 秒（與 6.4b 一致）
- **並發**: 使用 goroutine 處理切換，不阻塞主執行緒

### 依賴關係

- **前置依賴**:
  - Story 6.4a (音訊系統基礎架構)
  - Story 6.4b (BGM 播放器 - Crossfade 功能)
  - Story 2.2 (故事生成引擎 - LLM 回應解析)
  - Story 2.5 (遊戲主畫面佈局 - 遊戲主循環)

### 測試策略

1. **單元測試**:
   - 測試 mood 解析（各種標記格式）
   - 測試 mood 映射正確性
   - 測試防頻繁切換邏輯
2. **整合測試**:
   - 測試完整自動切換流程
   - 測試多次 mood 變化的處理
3. **效能測試**:
   - 測試切換延遲 < 200ms
   - 確認不阻塞主執行緒
4. **沉浸感測試**:
   - 手動遊玩驗證 BGM 切換與場景的契合度
   - 確認 30 秒間隔足夠平衡（不過於頻繁/不過於遲鈍）

### References

- [Source: docs/epics.md#Epic-6]
- [Story 6.4a: 音訊系統基礎架構]
- [Story 6.4b: BGM 播放器]
- [Story 2.2: 故事生成引擎]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- Story split from 6-4-bgm-system.md for better manageability
- Focused on BGM auto-switching based on LLM mood tags
- Depends on 6.4b (BGM player with crossfade) and 2.2 (LLM response parsing)
- Anti-frequent-switch mechanism (30s minimum interval)
- Ready for development - all acceptance criteria defined
