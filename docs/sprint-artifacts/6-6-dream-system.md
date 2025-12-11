# Story 6.6: 夢境系統

Status: ready-for-dev

## Story

As a 玩家,
I want 體驗夢境序列,
so that 夢境成為規則預告和氛圍營造的工具.

## Acceptance Criteria

### AC1: 開場夢境生成

**Given** 玩家開始新遊戲（Story 2.1）
**When** 故事生成引擎初始化
**Then** 生成開場夢境序列
**And** 夢境在正式故事開始前顯示
**And** 夢境長度 200-400 字
**And** 夢境內容暗示即將遇到的規則（模糊、隱喻方式）
**And** 夢境氛圍符合故事主題

### AC2: 夢境轉場效果

**Given** 進入或離開夢境
**When** 顯示轉場動畫
**Then** 使用霧化效果（逐漸模糊→清晰）
**And** 轉場動畫持續 1-2 秒
**And** 顯示提示文字：
  - 進入夢境：「意識逐漸模糊...」
  - 離開夢境：「你驚醒了」
**And** 轉場使用特殊樣式（如灰色漸變、邊框淡化）

### AC3: 夢境視覺風格

**Given** 玩家在夢境中
**When** 顯示夢境內容
**Then** 使用獨特視覺風格區分夢境與現實：
  - 文字顏色偏向灰/藍色調
  - 邊框使用虛線或淡化效果
  - 狀態列顯示「夢境」標記
  - 打字機效果速度略慢（營造迷幻感）
**And** 夢境中不顯示 HP/SAN 數值（僅顯示「夢境」）

### AC4: 夢境內容生成邏輯

**Given** 系統生成夢境內容
**When** 呼叫 Smart Model
**Then** Prompt 包含：
  - 故事主題與場景
  - 已知的潛規則（隱藏於夢境隱喻中）
  - 玩家恐懼元素
  - 夢境類型（預告/回憶/警告）
**And** 夢境內容應「似夢非夢」：
  - 邏輯不完全連貫（符合夢境特性）
  - 包含象徵性意象（如鏡子、門、迷宮）
  - 與主線相關但扭曲變形
**And** 不直接揭露規則（僅暗示）

### AC5: 章節轉換夢境

**Given** 遊戲進行到章節轉換點
**When** 敘事安排睡眠/昏迷事件
**Then** 可選擇性插入夢境序列（機率 30-50%）
**And** 夢境內容反映玩家當前處境：
  - 高壓力情境後：噩夢（重演恐怖時刻）
  - 發現線索後：提示夢境（線索關聯）
  - 隊友死亡後：悲傷夢境（情感處理）
**And** 夢境長度 100-200 字（較開場夢境短）

### AC6: 夢境回顧指令

**Given** 玩家在遊戲中
**When** 輸入 `/dreams`
**Then** 顯示已經歷的夢境清單：
  - 夢境編號（#1, #2...）
  - 夢境類型（開場/警告/回憶）
  - 發生時間（章節數或遊戲時間）
**And** 可選擇特定夢境重新閱讀完整內容
**And** 已閱讀夢境保持原樣（不重新生成）

### AC7: 覆盤時夢境解析

**Given** 玩家進入死亡覆盤（Story 3.4）
**When** 顯示覆盤資訊
**Then** 包含「夢境解析」專區：
  - 列出所有夢境
  - 解釋夢境與規則的關聯（現在揭露）
  - 標示哪些意象對應哪些規則
  - 顯示「你本可以從夢境中察覺...」提示
**And** 幫助玩家理解夢境的預告作用

### AC8: 夢境資料追蹤

**Given** 遊戲進行中
**When** 夢境發生
**Then** 記錄以下資料：
  - 夢境完整文字內容
  - 發生時間（章節/遊戲時間）
  - 夢境類型（開場/章節/警告）
  - 關聯的潛規則 ID（用於覆盤解析）
  - 生成時的上下文（玩家狀態、已知線索）
**And** 資料保存至 `GameState.DreamLog`
**And** 隨存檔一起保存

### AC9: 夢境互動限制

**Given** 玩家在夢境中
**When** 顯示夢境敘事
**Then** 不提供選項（夢境為被動觀看）
**And** 玩家只能閱讀，無法做出選擇
**And** 夢境結束後自動回到現實
**And** 顯示「按 Enter 繼續」提示

### AC10: 無障礙模式相容

**Given** 玩家啟用無障礙模式
**When** 夢境顯示
**Then** 使用文字明確標記：
  - 開頭顯示「【夢境開始】」
  - 結尾顯示「【夢境結束】」
**And** 不僅依賴視覺風格區分（保持顏色變化但添加文字）
**And** 轉場效果簡化（避免依賴視覺的模糊效果）
**And** 保持完整夢境內容可讀性

## Tasks / Subtasks

- [ ] Task 1: 建立夢境資料結構 (AC: #8)
  - [ ] Subtask 1.1: 建立 `internal/game/dream.go`
  - [ ] Subtask 1.2: 定義 `DreamRecord` 結構體
  - [ ] Subtask 1.3: 擴展 `GameState` 添加 `DreamLog []DreamRecord`
  - [ ] Subtask 1.4: 實作記錄追蹤函數 `LogDream()`
  - [ ] Subtask 1.5: 整合至存檔系統

- [ ] Task 2: 實作開場夢境生成 (AC: #1, #4)
  - [ ] Subtask 2.1: 擴展 `internal/engine/story.go` 添加夢境生成
  - [ ] Subtask 2.2: 建立開場夢境 prompt 模板
  - [ ] Subtask 2.3: 整合 Smart Model 生成夢境內容
  - [ ] Subtask 2.4: 將潛規則以隱喻方式編碼進 prompt
  - [ ] Subtask 2.5: 測試夢境內容的暗示品質

- [ ] Task 3: 實作夢境渲染元件 (AC: #2, #3)
  - [ ] Subtask 3.1: 建立 `internal/tui/components/dream_renderer.go`
  - [ ] Subtask 3.2: 定義 `DreamModel` 結構體（BubbleTea model）
  - [ ] Subtask 3.3: 實作夢境視覺風格（灰/藍色調、虛線邊框）
  - [ ] Subtask 3.4: 實作狀態列「夢境」標記
  - [ ] Subtask 3.5: 隱藏 HP/SAN 數值顯示

- [ ] Task 4: 實作夢境轉場效果 (AC: #2)
  - [ ] Subtask 4.1: 建立 `internal/tui/effects/dream_transition.go`
  - [ ] Subtask 4.2: 實作霧化效果（逐漸模糊→清晰動畫）
  - [ ] Subtask 4.3: 使用 BubbleTea tick 機制實現 1-2 秒動畫
  - [ ] Subtask 4.4: 添加轉場提示文字（「意識逐漸模糊...」）
  - [ ] Subtask 4.5: 測試轉場的視覺平滑度

- [ ] Task 5: 實作章節夢境插入邏輯 (AC: #5)
  - [ ] Subtask 5.1: 實作夢境觸發條件判斷（章節轉換點）
  - [ ] Subtask 5.2: 實作機率系統（30-50% 插入）
  - [ ] Subtask 5.3: 根據玩家狀態決定夢境類型（噩夢/提示/悲傷）
  - [ ] Subtask 5.4: 建立章節夢境 prompt 模板
  - [ ] Subtask 5.5: 整合至遊戲主循環

- [ ] Task 6: 實作 /dreams 指令 (AC: #6)
  - [ ] Subtask 6.1: 建立 `/dreams` 指令處理器
  - [ ] Subtask 6.2: 實作夢境清單 UI（編號/類型/時間）
  - [ ] Subtask 6.3: 實作夢境選擇與重新閱讀功能
  - [ ] Subtask 6.4: 從 DreamLog 載入歷史夢境內容
  - [ ] Subtask 6.5: 測試夢境回顧流程

- [ ] Task 7: 整合覆盤夢境解析 (AC: #7)
  - [ ] Subtask 7.1: 擴展 `internal/tui/views/debrief.go`
  - [ ] Subtask 7.2: 添加「夢境解析」專區
  - [ ] Subtask 7.3: 顯示夢境與規則的對應關聯
  - [ ] Subtask 7.4: 實作意象解釋（如「鏡子 = 對立規則」）
  - [ ] Subtask 7.5: 測試覆盤夢境解析的教學效果

- [ ] Task 8: Prompt 工程與調校 (AC: #4)
  - [ ] Subtask 8.1: 設計開場夢境 prompt 模板
  - [ ] Subtask 8.2: 設計章節夢境 prompt 模板（噩夢/提示/悲傷）
  - [ ] Subtask 8.3: 提供夢境範例（5+ 個高品質範例）
  - [ ] Subtask 8.4: 調校「似夢非夢」平衡（邏輯不連貫但有意義）
  - [ ] Subtask 8.5: 測試夢境暗示的微妙度（不過於明顯/不過於隱晦）

- [ ] Task 9: 實作夢境互動限制 (AC: #9)
  - [ ] Subtask 9.1: 夢境模式下禁用選項顯示
  - [ ] Subtask 9.2: 實作「按 Enter 繼續」提示
  - [ ] Subtask 9.3: 夢境結束後自動返回現實
  - [ ] Subtask 9.4: 測試夢境的被動觀看體驗

- [ ] Task 10: 無障礙模式與測試 (AC: #10)
  - [ ] Subtask 10.1: 添加「【夢境開始】」「【夢境結束】」文字標記
  - [ ] Subtask 10.2: 簡化轉場效果（保持可讀性）
  - [ ] Subtask 10.3: 確保夢境內容完全可讀
  - [ ] Subtask 10.4: 編寫整合測試覆蓋所有 AC
  - [ ] Subtask 10.5: 玩家體驗測試（夢境是否增強沉浸感）

## Dev Notes

### 架構模式

- **模組位置**:
  - 核心邏輯: `internal/game/dream.go`
  - 渲染元件: `internal/tui/components/dream_renderer.go`
  - 轉場效果: `internal/tui/effects/dream_transition.go`
  - 覆盤整合: `internal/tui/views/debrief.go` (擴展)

- **核心資料結構**:
  ```go
  type DreamRecord struct {
      ID             int
      Content        string
      Type           DreamType // Opening, Chapter, Warning
      ChapterNumber  int
      GameTime       time.Duration
      RelatedRuleIDs []int
      Context        DreamContext
      Timestamp      time.Time
  }

  type DreamType int
  const (
      DreamTypeOpening DreamType = iota
      DreamTypeChapter
      DreamTypeWarning
      DreamTypeNightmare
      DreamTypeGrief
  )

  type DreamContext struct {
      PlayerSAN      int
      KnownClues     []string
      RecentEvents   []string
  }
  ```

### 夢境視覺風格

- **顏色方案**:
  ```go
  type DreamStyle struct {
      TextColor      lipgloss.Color   // #7f8c9d (灰藍)
      BorderColor    lipgloss.Color   // #4a5568 (深灰)
      BorderStyle    lipgloss.Border  // lipgloss.RoundedBorder() 但虛線化
      BackgroundTint lipgloss.Color   // #1a1f2e (深藍黑)
  }
  ```

- **LipGloss 樣式**:
  ```go
  dreamStyle := lipgloss.NewStyle().
      Foreground(lipgloss.Color("#7f8c9d")).
      Border(lipgloss.RoundedBorder()).
      BorderForeground(lipgloss.Color("#4a5568")).
      Padding(1, 2)
  ```

### 轉場效果實作

- **霧化動畫**:
  ```go
  func (m DreamTransitionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      switch msg := msg.(type) {
      case tea.KeyMsg:
          // 無法跳過轉場
      case TickMsg:
          m.progress += 0.05 // 20 步驟 = 1 秒（50ms tick）
          if m.progress >= 1.0 {
              return m, m.OnComplete() // 轉場完成
          }
          return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
              return TickMsg{}
          })
      }
      return m, nil
  }

  func (m DreamTransitionModel) View() string {
      // 根據 progress 調整透明度/模糊度
      // 使用 ANSI 轉義序列或字元重疊實現模糊效果
      opacity := int(m.progress * 100)
      return fmt.Sprintf("\033[2m%s\033[0m", m.text) // 暗淡效果
  }
  ```

### Prompt 設計範例

#### 開場夢境 Prompt

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

#### 章節夢境 Prompt（噩夢類型）

```
生成一段噩夢，反映玩家剛經歷的恐怖事件。

最近事件：{recent_events}
玩家當前 SAN：{san}
已知線索：{clues}

要求：
1. 長度 100-200 字
2. 重演恐怖時刻但扭曲變形
3. 強化恐懼感（加入超現實元素）
4. 可能暗示即將到來的危險
5. 結尾突然驚醒

輸出格式：純夢境敘事文字。
```

### 夢境與規則關聯範例

| 規則類型 | 夢境意象 | 解釋 |
|---------|---------|-----|
| 對立規則 | 鏡中人做相反動作 | 暗示需要做與直覺相反的事 |
| 時間規則 | 時鐘停在特定時刻 | 暗示特定時間的危險 |
| 場景規則 | 某扇門永遠打不開 | 暗示特定場景的禁忌 |
| 對象規則 | NPC 臉孔模糊不清 | 暗示某人不可信任 |

### 夢境插入時機

- **開場**: 100% 插入（必定有開場夢境）
- **章節轉換**: 30-50% 機率（依情境）
  - 高壓力後: 70% 機率（噩夢）
  - 重大線索發現後: 50% 機率（提示夢境）
  - 隊友死亡後: 80% 機率（悲傷夢境）
  - 普通休息: 20% 機率（隨機夢境）

### 技術約束

- **夢境生成**: 使用 Smart Model（允許較長回應時間）
- **轉場動畫**: 1-2 秒（不可跳過，保持儀式感）
- **視覺效果**: 不影響文字可讀性（僅改變色調與邊框）
- **記憶體**: DreamLog 保存完整夢境文字（每個 < 1KB，總計 < 10KB）

### 依賴關係

- **前置依賴**:
  - Story 2.1 (新遊戲設定流程)
  - Story 2.2 (故事生成引擎)
  - Story 3.1 (潛規則生成)
  - Story 3.4 (死亡覆盤系統)
- **整合點**:
  - 遊戲開始時觸發開場夢境
  - 章節轉換點檢查是否插入夢境
  - 覆盤時解析夢境關聯

### 測試策略

1. **Prompt 品質測試**:
   - 生成 20+ 個夢境
   - 評估暗示的微妙度（不過於明顯/隱晦）
2. **視覺測試**:
   - 驗證夢境風格明顯區分於現實
   - 測試轉場動畫的平滑度
3. **整合測試**:
   - 測試開場夢境自動插入
   - 測試章節夢境機率觸發
   - 測試 `/dreams` 指令回顧
4. **覆盤測試**:
   - 驗證夢境解析正確對應規則
   - 確認玩家能從解析中學習
5. **沉浸感測試**:
   - 玩家回饋：夢境是否增強恐怖體驗
   - 夢境是否成功預告規則（事後回想）

### 設計意圖

- **夢境作用**:
  1. **規則預告**: 以隱喻方式提前暗示（公平遊戲設計）
  2. **氛圍營造**: 增強恐怖與超自然感
  3. **敘事節奏**: 章節間的情緒緩衝
  4. **覆盤教學**: 幫助玩家理解「原來有提示」

- **不應做**:
  - 不直接揭露規則（失去發現樂趣）
  - 不過於頻繁（避免打斷遊戲節奏）
  - 不過長（保持簡短有力）
  - 不阻擋遊戲進程（被動觀看，按 Enter 繼續）

### References

- [Source: docs/epics.md#Epic-6]
- [UX Design: docs/ux-design-specification.md - 夢境體驗旅程]
- [Architecture: ARCHITECTURE.md - Smart Model 使用場景]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development - all acceptance criteria defined
- Dream metaphor system designed for fair rule hinting
- Visual style clearly distinguishes dreams from reality
- Debrief integration reveals dream-rule connections
- Prompt templates emphasize "dreamlike but meaningful" balance
