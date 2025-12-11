# Story 6.3: 打字機效果

Status: Ready for Review

## Story

As a 玩家,
I want 看到敘事以打字機效果顯示,
so that 閱讀過程更有沉浸感.

## Acceptance Criteria

### AC1: 基礎打字機效果

**Given** LLM 開始串流輸出敘事文字
**When** TUI 接收串流資料
**Then** 文字以打字機效果逐字顯示
**And** 顯示速度為 30-50 字/秒（可配置）
**And** 英文按字母顯示，中文按字顯示
**And** 標點符號後暫停時間加倍（如句號、問號後 40-100ms）

### AC2: 跳過打字機效果

**Given** 打字機效果正在進行中
**When** 玩家按下 Space 或 Enter 鍵
**Then** 立即跳過動畫，顯示完整文字
**And** 跳過操作響應時間 < 100ms
**And** 已顯示的文字保持在畫面上
**And** 繼續接收後續串流資料

### AC3: 低 SAN 時效果變化

**Given** 玩家 SAN < 40
**When** 打字機效果運作
**Then** 顯示速度變得不穩定：
  - 隨機加速（100-150 字/秒持續 0.5-1 秒）
  - 隨機減速（10-20 字/秒持續 0.5-1 秒）
  - 偶爾「卡頓」（暫停 200-500ms）
**And** 速度變化頻率與 HorrorStyle.TypingBehavior 關聯

### AC4: 低 SAN 時文字錯誤

**Given** 玩家 SAN < 20
**When** 打字機效果顯示文字
**Then** 可能出現以下錯誤：
  - 重複字元（如「你你你走進房間」）機率 5%
  - 吞字（跳過 1-2 個字元）機率 3%
  - 錯誤字元閃現後立即更正（顯示 100ms）機率 2%
**And** 錯誤機率隨 SAN 降低而增加
**And** 最終文字內容仍保持正確（錯誤僅為視覺效果）

### AC5: 停用打字機效果

**Given** 玩家在遊戲中
**When** 輸入 `/speed off`
**Then** 停用打字機效果
**And** 後續敘事文字立即完整顯示（無動畫）
**And** 設定儲存至配置檔
**And** 下次啟動遊戲時保持停用狀態

**Given** 打字機效果已停用
**When** 輸入 `/speed on` 或 `/speed 40`
**Then** 重新啟用打字機效果
**And** 可選擇性設定速度（字/秒）
**And** 立即生效

### AC6: 速度調整指令

**Given** 玩家在遊戲中
**When** 輸入 `/speed 60`
**Then** 設定打字機速度為 60 字/秒
**And** 有效範圍 10-200 字/秒
**And** 超出範圍顯示錯誤並使用預設值 40
**And** 立即應用到後續文字

### AC7: 串流整合與效能

**Given** LLM 正在串流輸出
**When** TUI 處理串流資料
**Then** 打字機效果不阻塞串流接收
**And** 緩衝區大小足夠（≥4KB）
**And** 即使打字機速度慢於串流速度，仍能完整接收所有資料
**And** 記憶體使用增加 < 5MB

### AC8: 視覺回饋與游標

**Given** 打字機效果進行中
**When** 顯示文字
**Then** 可選顯示閃爍游標於當前顯示位置（可配置）
**And** 游標使用 `▌` 或 `_` 字元
**And** 游標閃爍頻率 500ms（SAN 正常時）
**And** SAN < 40 時游標閃爍不穩定（與 Story 6.1 整合）

### AC9: 特殊文字處理

**Given** 敘事文字包含特殊格式
**When** 打字機效果顯示
**Then** 正確處理：
  - Markdown 粗體/斜體（保持樣式）
  - 顏色標記（保持顏色）
  - Emoji（視為 1 個字元）
  - ANSI 轉義序列（不可見字元不計入速度）
**And** 樣式不影響打字速度計算

### AC10: 無障礙模式相容

**Given** 玩家啟用無障礙模式
**When** 打字機效果運作
**Then** 效果仍然啟用（保持沉浸感）
**And** 不使用依賴視覺的「錯誤閃現」效果（AC4）
**And** 速度變化保持（聽覺上仍可感知）
**And** 提供清晰的「按空格鍵跳過」提示

## Tasks / Subtasks

- [x] Task 1: 建立打字機效果核心元件 (AC: #1)
  - [x] Subtask 1.1: 建立 `internal/tui/components/typewriter_view.go`
  - [x] Subtask 1.2: 定義 `TypewriterModel` 結構體（包含速度、緩衝區、狀態）
  - [x] Subtask 1.3: 實作 `Update()` 處理 tick 事件逐字顯示
  - [x] Subtask 1.4: 實作 `View()` 渲染當前顯示文字
  - [x] Subtask 1.5: 實作字元計數邏輯（中文 1 字 = 1 單位，英文 1 字母 = 1 單位）

- [x] Task 2: 實作速度控制與暫停邏輯 (AC: #1)
  - [x] Subtask 2.1: 實作可配置速度（字/秒轉換為 tick 間隔）
  - [x] Subtask 2.2: 實作標點符號後加倍暫停（句號、問號、驚嘆號）
  - [x] Subtask 2.3: 建立 `CalculateDelay(char rune) time.Duration` 函數
  - [x] Subtask 2.4: 測試不同速度設定的視覺效果

- [x] Task 3: 實作跳過功能 (AC: #2)
  - [x] Subtask 3.1: 監聽 Space/Enter 按鍵事件
  - [x] Subtask 3.2: 實作 `Skip()` 方法立即顯示完整文字
  - [x] Subtask 3.3: 確保跳過後串流仍正常接收
  - [x] Subtask 3.4: 測試跳過操作響應時間 < 100ms

- [x] Task 4: 整合 LLM 串流 (AC: #7)
  - [x] Subtask 4.1: 擴展 `internal/engine/streaming.go` 串流處理
  - [x] Subtask 4.2: 將串流資料導入 StreamBuffer 緩衝區
  - [x] Subtask 4.3: 實作緩衝區管理（≥4KB，動態擴展）
  - [x] Subtask 4.4: 確保打字機速度不阻塞串流接收
  - [x] Subtask 4.5: 測試慢速打字機 + 快速串流場景

- [x] Task 5: 實作低 SAN 速度變化 (AC: #3)
  - [x] Subtask 5.1: 整合 HorrorStyle.TypingBehavior 維度
  - [x] Subtask 5.2: 實作隨機速度變化邏輯（加速/減速/卡頓）
  - [x] Subtask 5.3: 定義 SAN < 40 時的速度波動範圍
  - [x] Subtask 5.4: 實作卡頓效果（暫停 200-500ms）
  - [x] Subtask 5.5: 測試低 SAN 時的視覺不穩定性

- [x] Task 6: 實作低 SAN 文字錯誤 (AC: #4)
  - [x] Subtask 6.1: 實作重複字元效果（5% 機率）
  - [x] Subtask 6.2: 實作吞字效果（3% 機率）
  - [x] Subtask 6.3: 實作錯誤字元閃現（顯示 100ms 後更正，2% 機率）
  - [x] Subtask 6.4: 確保最終文字內容正確（僅視覺效果）
  - [x] Subtask 6.5: 機率與 SAN 值關聯（SAN 越低機率越高）

- [x] Task 7: 實作指令控制 (AC: #5, #6)
  - [x] Subtask 7.1: 建立 `/speed` 指令處理器
  - [x] Subtask 7.2: 支援 `/speed off` 停用效果
  - [x] Subtask 7.3: 支援 `/speed on` 啟用效果
  - [x] Subtask 7.4: 支援 `/speed <number>` 設定速度（10-200 範圍）
  - [x] Subtask 7.5: 設定持久化至 `~/.nightmare/config.json`
  - [x] Subtask 7.6: 載入設定時還原打字機偏好

- [x] Task 8: 實作游標顯示 (AC: #8)
  - [x] Subtask 8.1: 添加可選的打字機游標（`▌` 字元）
  - [x] Subtask 8.2: 實作游標閃爍（500ms 間隔）
  - [x] Subtask 8.3: 整合 Story 6.1 的游標加速效果（SAN < 40）
  - [x] Subtask 8.4: 提供配置選項開關游標顯示

- [x] Task 9: 實作特殊文字處理 (AC: #9)
  - [x] Subtask 9.1: 解析 Markdown 粗體/斜體（保持樣式）
  - [x] Subtask 9.2: 處理 ANSI 顏色轉義序列（不計入速度）
  - [x] Subtask 9.3: 正確計算 Emoji 為 1 個字元
  - [x] Subtask 9.4: 測試混合格式文字的打字機效果

- [x] Task 10: 無障礙模式與測試 (AC: #10)
  - [x] Subtask 10.1: 確認無障礙模式下仍啟用打字機
  - [x] Subtask 10.2: 停用視覺依賴的「錯誤閃現」效果
  - [x] Subtask 10.3: 添加「按空格鍵跳過」提示
  - [x] Subtask 10.4: 編寫整合測試覆蓋所有 AC
  - [x] Subtask 10.5: 效能測試（記憶體 < 5MB，CPU < 5%）

## Dev Notes

### 架構模式

- **模組位置**: `internal/tui/components/typewriter_view.go`
- **核心結構體**:
  ```go
  type TypewriterModel struct {
      Buffer         []rune        // 完整文字緩衝區
      DisplayedIndex int           // 當前顯示到的索引
      Speed          int           // 字/秒
      Enabled        bool          // 是否啟用效果
      TickInterval   time.Duration // 計算後的 tick 間隔
      HorrorStyle    HorrorStyle   // 整合 SAN 效果
      ShowCursor     bool          // 是否顯示游標
      Skipped        bool          // 是否已跳過
  }
  ```

- **速度計算**:
  ```go
  // 字/秒轉換為 tick 間隔
  func CalculateTickInterval(speed int) time.Duration {
      return time.Second / time.Duration(speed)
  }

  // 標點符號暫停加倍
  func CalculateDelay(char rune, baseDelay time.Duration) time.Duration {
      if char == '。' || char == '?' || char == '!' || char == '.' {
          return baseDelay * 2
      }
      return baseDelay
  }
  ```

### BubbleTea 整合

- **Tick 事件**: 使用 `tea.Tick()` 定時觸發字元顯示
- **Update 邏輯**:
  ```go
  func (m TypewriterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      switch msg := msg.(type) {
      case tea.KeyMsg:
          if msg.String() == " " || msg.String() == "enter" {
              return m.Skip()
          }
      case TickMsg:
          if !m.Skipped && m.DisplayedIndex < len(m.Buffer) {
              m.DisplayedIndex++
              return m, tea.Tick(m.CalculateNextDelay(), func(t time.Time) tea.Msg {
                  return TickMsg{}
              })
          }
      }
      return m, nil
  }
  ```

### 低 SAN 效果整合

- **速度變化**:
  ```go
  func (m *TypewriterModel) ApplyHorrorEffects() time.Duration {
      if m.HorrorStyle.TypingBehavior > 0.1 {
          // 隨機速度波動
          if rand.Float64() < 0.1 {
              return m.TickInterval / 3 // 加速
          } else if rand.Float64() < 0.1 {
              return m.TickInterval * 3 // 減速
          } else if rand.Float64() < 0.05 {
              return time.Millisecond * 300 // 卡頓
          }
      }
      return m.TickInterval
  }
  ```

- **文字錯誤**:
  ```go
  func (m *TypewriterModel) ApplyTextGlitches() rune {
      char := m.Buffer[m.DisplayedIndex]
      if m.HorrorStyle.TypingBehavior > 0.15 {
          if rand.Float64() < 0.05 {
              // 重複字元
              return char // 下次 tick 再次顯示
          } else if rand.Float64() < 0.03 {
              // 吞字
              m.DisplayedIndex++ // 跳過當前字元
              return m.Buffer[m.DisplayedIndex]
          }
      }
      return char
  }
  ```

### 技術約束

- **效能目標**:
  - CPU 使用 < 5%
  - 記憶體增加 < 5MB
  - 不阻塞主執行緒
- **串流相容**: 緩衝區動態擴展，不設上限（但實際章節 < 10KB）
- **響應速度**: 跳過操作 < 100ms
- **配置持久化**: 設定儲存至 `~/.nightmare/config.json`

### 依賴關係

- **前置依賴**:
  - Story 2.2 (故事生成引擎 - LLM 串流)
  - Story 6.1 (HorrorStyle 系統)
- **整合點**:
  - `internal/engine/story.go` 串流輸出
  - `internal/tui/views/game.go` 主畫面敘事區

### 測試策略

1. **單元測試**:
   - 測試速度計算公式
   - 測試標點符號暫停邏輯
   - 測試跳過功能
2. **視覺測試**:
   - 手動驗證不同速度的打字效果
   - 驗證低 SAN 時的視覺錯誤
3. **整合測試**:
   - 測試與 LLM 串流的整合
   - 測試長文字（>5000 字）的效能
4. **效能測試**:
   - 監控 CPU 與記憶體使用
   - 壓力測試連續串流
5. **無障礙測試**:
   - 驗證無障礙模式相容性

### 配置範例

```json
{
  "typewriter": {
    "enabled": true,
    "speed": 40,
    "show_cursor": true
  }
}
```

### References

- [Source: docs/epics.md#Epic-6]
- [UX Design: docs/ux-design-specification.md - 打字機效果規範]
- [Architecture: ARCHITECTURE.md - BubbleTea 架構]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Implementation Plan

Story developed using TDD approach with comprehensive test coverage.

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development - all acceptance criteria defined
- Technical implementation leverages BubbleTea tick mechanism
- Horror effects integrated with Story 6.1 HorrorStyle system

**Implementation Complete (2025-12-11):**
- ✅ Extended `config.Config` with `TypewriterConfig` field for settings persistence
- ✅ Enhanced `engine.StreamBuffer` with low SAN effects:
  - Speed variation (fast/slow/stuck modes) based on SAN < 40
  - Text glitches (repeat/skip/flicker) for SAN < 20
  - Glitch probability scales with SAN level
- ✅ Created `/speed` command for runtime control:
  - `/speed on|off` - Enable/disable effect
  - `/speed <10-200>` - Set speed (chars/sec)
  - Settings persist to ~/.nightmare/config.json
  - Validates range and uses default (40) for out-of-range values
- ✅ Built `TypewriterView` TUI component:
  - Integrates with StreamBuffer for animation
  - Handles Space/Enter for skip functionality
  - Cursor display with SAN-based blink rate
  - Text wrapping with CJK/emoji width support
  - Accessibility mode with clear skip hints
- ✅ Comprehensive test coverage:
  - Config persistence tests
  - Low SAN effects tests (speed & glitches)
  - Command tests with all edge cases
  - TUI component tests including accessibility
  - All tests passing

### File List

- internal/config/config.go
- internal/config/typewriter_test.go
- internal/engine/streaming.go
- internal/engine/typewriter_effects_test.go
- internal/game/commands/speed.go
- internal/game/commands/speed_test.go
- internal/tui/components/typewriter_view.go
- internal/tui/components/typewriter_view_test.go

### Change Log

- 2025-12-11: Implemented typewriter effect system with low SAN integration
  - Added typewriter configuration to Config struct
  - Implemented speed variation and text glitch effects for low SAN
  - Created /speed command for user control
  - Built TypewriterView component with full BubbleTea integration
  - Wrote comprehensive test suite (100% passing)
