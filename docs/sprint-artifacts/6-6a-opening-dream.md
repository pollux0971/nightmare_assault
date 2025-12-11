# Story 6.6a: 開場夢境生成

Status: ready-for-dev

## Story

As a 玩家,
I want 遊戲開始時體驗開場夢境,
so that 夢境預告即將遇到的規則並營造氛圍.

## Acceptance Criteria

### AC1: 開場夢境生成

**Given** 玩家開始新遊戲（Story 2.1）
**When** 故事生成引擎初始化
**Then** 生成開場夢境序列
**And** 夢境在正式故事開始前顯示
**And** 夢境長度 200-400 字
**And** 夢境內容暗示即將遇到的規則（模糊、隱喻方式）
**And** 夢境氛圍符合故事主題

### AC2: 夢境內容生成邏輯（Prompt 工程）

**Given** 系統生成夢境內容
**When** 呼叫 Smart Model
**Then** Prompt 包含：
  - 故事主題與場景
  - 已知的潛規則（隱藏於夢境隱喻中）
  - 玩家恐懼元素
  - 夢境類型（預告）
**And** 夢境內容應「似夢非夢」：
  - 邏輯不完全連貫（符合夢境特性）
  - 包含象徵性意象（如鏡子、門、迷宮）
  - 與主線相關但扭曲變形
**And** 不直接揭露規則（僅暗示）

### AC3: 夢境渲染與視覺風格

**Given** 玩家在夢境中
**When** 顯示夢境內容
**Then** 使用獨特視覺風格區分夢境與現實：
  - 文字顏色偏向灰/藍色調
  - 邊框使用虛線或淡化效果
  - 狀態列顯示「夢境」標記
  - 打字機效果速度略慢（營造迷幻感）
**And** 夢境中不顯示 HP/SAN 數值（僅顯示「夢境」）

### AC4: 夢境轉場效果

**Given** 進入或離開夢境
**When** 顯示轉場動畫
**Then** 使用霧化效果（逐漸模糊→清晰）
**And** 轉場動畫持續 1-2 秒
**And** 顯示提示文字：
  - 進入夢境：「意識逐漸模糊...」
  - 離開夢境：「你驚醒了」
**And** 轉場使用特殊樣式（如灰色漸變、邊框淡化）

### AC5: 夢境資料追蹤

**Given** 開場夢境生成
**When** 夢境顯示完成
**Then** 記錄以下資料：
  - 夢境完整文字內容
  - 夢境類型（開場）
  - 關聯的潛規則 ID（用於覆盤解析）
  - 生成時的上下文（玩家狀態、已知線索）
**And** 資料保存至 `GameState.DreamLog`
**And** 隨存檔一起保存

### AC6: 無障礙模式相容

**Given** 玩家啟用無障礙模式
**When** 夢境顯示
**Then** 使用文字明確標記：
  - 開頭顯示「【夢境開始】」
  - 結尾顯示「【夢境結束】」
**And** 不僅依賴視覺風格區分（保持顏色變化但添加文字）
**And** 轉場效果簡化（避免依賴視覺的模糊效果）

## Tasks / Subtasks

- [ ] Task 1: 建立夢境資料結構 (AC: #5)
  - [ ] Subtask 1.1: 建立 `internal/game/dream.go`
  - [ ] Subtask 1.2: 定義 `DreamRecord` 結構體
  - [ ] Subtask 1.3: 擴展 `GameState` 添加 `DreamLog []DreamRecord`
  - [ ] Subtask 1.4: 實作記錄追蹤函數 `LogDream()`
  - [ ] Subtask 1.5: 整合至存檔系統

- [ ] Task 2: 實作開場夢境生成 (AC: #1, #2)
  - [ ] Subtask 2.1: 擴展 `internal/engine/story.go` 添加夢境生成
  - [ ] Subtask 2.2: 建立開場夢境 prompt 模板
  - [ ] Subtask 2.3: 整合 Smart Model 生成夢境內容
  - [ ] Subtask 2.4: 將潛規則以隱喻方式編碼進 prompt
  - [ ] Subtask 2.5: 測試夢境內容的暗示品質

- [ ] Task 3: 實作夢境渲染元件 (AC: #3)
  - [ ] Subtask 3.1: 建立 `internal/tui/components/dream_renderer.go`
  - [ ] Subtask 3.2: 定義 `DreamModel` 結構體（BubbleTea model）
  - [ ] Subtask 3.3: 實作夢境視覺風格（灰/藍色調、虛線邊框）
  - [ ] Subtask 3.4: 實作狀態列「夢境」標記
  - [ ] Subtask 3.5: 隱藏 HP/SAN 數值顯示

- [ ] Task 4: 實作夢境轉場效果 (AC: #4)
  - [ ] Subtask 4.1: 建立 `internal/tui/effects/dream_transition.go`
  - [ ] Subtask 4.2: 實作霧化效果（逐漸模糊→清晰動畫）
  - [ ] Subtask 4.3: 使用 BubbleTea tick 機制實現 1-2 秒動畫
  - [ ] Subtask 4.4: 添加轉場提示文字（「意識逐漸模糊...」）
  - [ ] Subtask 4.5: 測試轉場的視覺平滑度

- [ ] Task 5: Prompt 工程與調校 (AC: #2)
  - [ ] Subtask 5.1: 設計開場夢境 prompt 模板
  - [ ] Subtask 5.2: 提供夢境範例（5+ 個高品質範例）
  - [ ] Subtask 5.3: 調校「似夢非夢」平衡（邏輯不連貫但有意義）
  - [ ] Subtask 5.4: 測試夢境暗示的微妙度（不過於明顯/不過於隱晦）
  - [ ] Subtask 5.5: 迭代調整 prompt 基於測試結果

- [ ] Task 6: 無障礙模式與測試 (AC: #6)
  - [ ] Subtask 6.1: 添加「【夢境開始】」「【夢境結束】」文字標記
  - [ ] Subtask 6.2: 簡化轉場效果（保持可讀性）
  - [ ] Subtask 6.3: 確保夢境內容完全可讀
  - [ ] Subtask 6.4: 編寫整合測試覆蓋所有 AC
  - [ ] Subtask 6.5: 玩家體驗測試（夢境是否增強沉浸感）

## Dev Notes

### 開場夢境 Prompt 模板

```
你是恐怖遊戲的夢境設計師。生成一段開場夢境，預告即將發生的恐怖故事。

故事主題：{theme}
潛規則概要（隱藏）：{rules_abstract}
玩家角色：{player_role}

要求：
1. 長度 200-400 字
2. 使用隱喻與象徵（不直接揭露規則）
3. 氛圍迷幻、邏輯略顯不連貫（符合夢境特性）
4. 包含 2-3 個意象（如鏡子、門、迷宮、影子）
5. 暗示規則但不明說（如「鏡中的你做著相反的動作」→ 對立規則）
6. 結尾留懸念，引導進入正式故事

輸出格式：純夢境敘事文字，無額外說明。
```

### 夢境視覺風格

```go
dreamStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#7f8c9d")).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#4a5568")).
    Padding(1, 2)
```

### 依賴關係

- **前置依賴**:
  - Story 2.1 (新遊戲設定流程)
  - Story 2.2 (故事生成引擎)
  - Story 3.1 (潛規則生成)
- **後續依賴**:
  - Story 6.6b (章節夢境 - 共用渲染與轉場)
  - Story 6.6c (夢境回顧 - 訪問 DreamLog)

### References

- [Source: docs/epics.md#Epic-6]
- [UX Design: docs/ux-design-specification.md - 夢境體驗旅程]
- [Architecture: ARCHITECTURE.md - Smart Model 使用場景]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- Story split from 6-6-dream-system.md for better manageability
- Focused on opening dream generation (prompt engineering, rendering, transition effects)
- Depends on 2.1 (new game setup), 2.2 (story engine), 3.1 (rule generation)
- Provides base for 6.6b (chapter dreams) and 6.6c (dream review)
- Ready for development - all acceptance criteria defined
