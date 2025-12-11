# Story 6.5: 音效系統

Status: ready-for-dev

## Story

As a 玩家,
I want 聽到事件音效,
so that 關鍵時刻有聽覺回饋.

## Acceptance Criteria

### AC1: SFX 播放器基礎功能

**Given** 音訊系統已初始化（Story 6.4）
**When** 建立 SFX 播放器
**Then** 支援同時播放多個音效（≥4 個通道）
**And** 每個音效獨立控制音量
**And** SFX 與 BGM 混音播放（不互相中斷）
**And** 使用獨立的 oto.Player 實例

### AC2: 心跳音效（低 SAN 觸發）

**Given** 玩家 SAN < 40
**When** 遊戲運作中
**Then** 開始播放心跳音效（heartbeat.wav）
**And** 心跳速度隨 SAN 降低而加快：
  - SAN 30-39: 60 BPM（正常）
  - SAN 20-29: 90 BPM（加快）
  - SAN 10-19: 120 BPM（急促）
  - SAN 1-9: 150 BPM（極度恐慌）
**And** 音量隨 SAN 降低而增大（SAN 10 時達到最大音量）
**And** 心跳循環播放直到 SAN 恢復 ≥ 40

**Given** 玩家 SAN 恢復至 ≥ 40
**When** SAN 變化事件觸發
**Then** 心跳音效 2 秒淡出
**And** 停止播放

### AC3: 警告音效（規則觸發）

**Given** 玩家觸發潛規則警告（Story 3.2）
**When** 規則檢測系統發出警告
**Then** 播放警告音效（warning.wav）
**And** 音效持續 1-2 秒
**And** 音量較 BGM 高（確保引起注意）
**And** 警告音效優先級高於其他 SFX

### AC4: 死亡音效

**Given** 玩家 HP 降至 0 或觸發即死（Story 3.3）
**When** 進入死亡流程
**Then** 立即播放死亡音效（death.wav）
**And** 同時停止所有其他 SFX（包括心跳）
**And** BGM 淡出至靜音（2 秒）
**And** 死亡音效播放完畢後進入死亡畫面

### AC5: 環境音效

**Given** 遊戲敘事中描述特定環境事件
**When** LLM 回應包含 SFX 標記（如 `[SFX:door]`）
**Then** 播放對應環境音效：
  - `door_open.wav`: 門開啟
  - `door_close.wav`: 門關閉
  - `footsteps.wav`: 腳步聲
  - `glass_break.wav`: 玻璃碎裂
  - `thunder.wav`: 雷聲
  - `whisper.wav`: 耳語
**And** 環境音效音量低於警告音效（背景氛圍用）
**And** 不阻塞敘事顯示

### AC6: SFX 混音邏輯

**Given** 多個音效同時觸發
**When** SFX 播放器處理
**Then** 使用優先級系統：
  1. 死亡音效（最高優先級，停止所有其他 SFX）
  2. 警告音效（高優先級，可中斷環境音效）
  3. 心跳音效（中優先級，循環播放）
  4. 環境音效（低優先級，可被覆蓋）
**And** 同優先級音效使用不同通道同時播放（最多 4 個）
**And** 超過通道數時，停止最舊的同優先級音效

### AC7: SFX 控制指令

**Given** 玩家在遊戲中
**When** 輸入 `/sfx off`
**Then** 停止所有 SFX 播放
**And** 後續事件不觸發 SFX
**And** 設定儲存至配置檔

**Given** SFX 已停用
**When** 輸入 `/sfx on`
**Then** 重新啟用 SFX
**And** 若當前 SAN < 40，立即開始心跳音效

**Given** 玩家在遊戲中
**When** 輸入 `/sfx volume 60`
**Then** 設定 SFX 音量為 60%
**And** 有效範圍 0-100
**And** 即時生效（不影響已播放的音效）
**And** 設定儲存至配置檔

### AC8: 靜默失敗機制

**Given** 音訊檔案不存在或損壞
**When** 嘗試播放 SFX
**Then** 靜默失敗（不中斷遊戲）
**And** 記錄警告日誌到 `~/.nightmare/debug.log`
**And** 不顯示錯誤提示（避免打斷沉浸感）
**And** 繼續正常遊戲流程

### AC9: 效能與資源管理

**Given** SFX 系統運作中
**When** 監控資源使用
**Then** 記憶體使用 < 20MB（同時載入 4 個音效）
**And** CPU 使用 < 2%（解碼與播放）
**And** 音效觸發延遲 < 50ms
**And** 不阻塞主執行緒（使用 goroutine）

### AC10: 無障礙模式文字提示

**Given** 玩家啟用無障礙模式且開啟音訊描述
**When** SFX 播放
**Then** 顯示文字描述：
  - 心跳：「[心跳聲加速]」
  - 警告：「[警告音效]」
  - 死亡：「[死亡音效]」
  - 環境：「[門開啟聲]」
**And** 文字提示短暫顯示（1-2 秒）於狀態列
**And** 不影響敘事區內容

## Tasks / Subtasks

- [ ] Task 1: 建立 SFX 播放器架構 (AC: #1, #6)
  - [ ] Subtask 1.1: 建立 `internal/audio/sfx_player.go`
  - [ ] Subtask 1.2: 定義 `SFXPlayer` 結構體（包含通道池、優先級佇列）
  - [ ] Subtask 1.3: 實作 4 通道音效混音系統
  - [ ] Subtask 1.4: 實作優先級系統（死亡 > 警告 > 心跳 > 環境）
  - [ ] Subtask 1.5: 實作通道管理（超過 4 個時停止最舊音效）

- [ ] Task 2: 實作心跳音效系統 (AC: #2)
  - [ ] Subtask 2.1: 建立 `internal/audio/heartbeat.go`
  - [ ] Subtask 2.2: 實作 BPM 控制（基於 SAN 值映射）
  - [ ] Subtask 2.3: 實作循環播放與速度調整
  - [ ] Subtask 2.4: 整合 EventBus 監聽 SAN 變化
  - [ ] Subtask 2.5: 實作淡入淡出（SAN 跨越 40 閾值時）
  - [ ] Subtask 2.6: 測試不同 SAN 範圍的心跳速度

- [ ] Task 3: 實作警告音效 (AC: #3)
  - [ ] Subtask 3.1: 整合 Story 3.2 規則觸發系統
  - [ ] Subtask 3.2: 實作 `PlayWarning()` 函數
  - [ ] Subtask 3.3: 確保警告音效優先級高於環境音效
  - [ ] Subtask 3.4: 測試警告音效與 BGM 混音

- [ ] Task 4: 實作死亡音效 (AC: #4)
  - [ ] Subtask 4.1: 整合 Story 3.3 死亡流程
  - [ ] Subtask 4.2: 實作 `PlayDeath()` 函數
  - [ ] Subtask 4.3: 實作停止所有其他 SFX 邏輯
  - [ ] Subtask 4.4: 實作 BGM 淡出（2 秒）
  - [ ] Subtask 4.5: 確保死亡音效完整播放後再顯示死亡畫面

- [ ] Task 5: 實作環境音效觸發 (AC: #5)
  - [ ] Subtask 5.1: 擴展 LLM 回應解析（Fast Model）偵測 `[SFX:xxx]` 標記
  - [ ] Subtask 5.2: 建立 SFX 標記到檔案名稱的映射表
  - [ ] Subtask 5.3: 實作 `PlayEnvironmentSFX(tag string)` 函數
  - [ ] Subtask 5.4: 確保環境音效不阻塞敘事顯示
  - [ ] Subtask 5.5: 測試多個環境音效同時播放

- [ ] Task 6: 實作混音與通道管理 (AC: #1, #6, #9)
  - [ ] Subtask 6.1: 實作通道池（4 個 oto.Player）
  - [ ] Subtask 6.2: 實作通道分配邏輯（基於優先級）
  - [ ] Subtask 6.3: 實作音效佇列（等待可用通道）
  - [ ] Subtask 6.4: 測試 4+ 個音效同時播放的處理
  - [ ] Subtask 6.5: 效能測試（確保 < 20MB 記憶體、< 2% CPU）

- [ ] Task 7: 實作 SFX 控制指令 (AC: #7)
  - [ ] Subtask 7.1: 建立 `/sfx` 指令處理器
  - [ ] Subtask 7.2: 實作 `/sfx off` 停用所有 SFX
  - [ ] Subtask 7.3: 實作 `/sfx on` 啟用 SFX（恢復心跳如需要）
  - [ ] Subtask 7.4: 實作 `/sfx volume <0-100>` 音量調整
  - [ ] Subtask 7.5: 配置持久化至 `~/.nightmare/config.json`

- [ ] Task 8: 實作音量控制 (AC: #2, #7, #9)
  - [ ] Subtask 8.1: 實作 `SetVolume(volume float64)` 全域 SFX 音量
  - [ ] Subtask 8.2: 實作個別音效音量調整（心跳隨 SAN 變化）
  - [ ] Subtask 8.3: 整合至音訊資料處理（手動音量調整）
  - [ ] Subtask 8.4: 測試音量變化的即時生效

- [ ] Task 9: 實作靜默失敗機制 (AC: #8)
  - [ ] Subtask 9.1: 實作錯誤處理不中斷遊戲
  - [ ] Subtask 9.2: 記錄警告至 `~/.nightmare/debug.log`
  - [ ] Subtask 9.3: 避免錯誤提示（保持沉浸感）
  - [ ] Subtask 9.4: 測試音訊檔案缺失情境

- [ ] Task 10: 實作無障礙模式音訊描述 (AC: #10)
  - [ ] Subtask 10.1: 整合音訊描述設定（Story 6.4）
  - [ ] Subtask 10.2: 實作 SFX 文字提示（顯示於狀態列）
  - [ ] Subtask 10.3: 實作短暫顯示邏輯（1-2 秒後消失）
  - [ ] Subtask 10.4: 測試無障礙模式完整流程

- [ ] Task 11: 整合測試與調校 (AC: #1-10)
  - [ ] Subtask 11.1: 測試所有 SFX 與 BGM 混音
  - [ ] Subtask 11.2: 測試優先級系統正確運作
  - [ ] Subtask 11.3: 測試心跳音效隨 SAN 變化
  - [ ] Subtask 11.4: 效能測試（記憶體/CPU/延遲）
  - [ ] Subtask 11.5: 完整遊戲流程測試（開始→死亡）

## Dev Notes

### 架構模式

- **模組位置**: `internal/audio/sfx_player.go`、`internal/audio/heartbeat.go`
- **核心結構體**:
  ```go
  type SFXPlayer struct {
      ctx          *oto.Context
      channels     [4]*SFXChannel // 4 個音效通道
      volume       float64
      enabled      bool
      heartbeat    *HeartbeatController
  }

  type SFXChannel struct {
      player      *oto.Player
      priority    int
      startTime   time.Time
      isPlaying   bool
  }

  type HeartbeatController struct {
      player      *oto.Player
      currentBPM  int
      isPlaying   bool
  }
  ```

- **優先級定義**:
  ```go
  const (
      PriorityDeath       = 100
      PriorityWarning     = 80
      PriorityHeartbeat   = 50
      PriorityEnvironment = 20
  )
  ```

### 心跳音效實作

- **BPM 到播放間隔映射**:
  ```go
  func CalculateHeartbeatInterval(san int) time.Duration {
      var bpm int
      switch {
      case san >= 40:
          return 0 // 不播放
      case san >= 30:
          bpm = 60
      case san >= 20:
          bpm = 90
      case san >= 10:
          bpm = 120
      default:
          bpm = 150
      }
      return time.Minute / time.Duration(bpm)
  }
  ```

- **循環播放邏輯**:
  ```go
  func (h *HeartbeatController) Start(san int) {
      h.isPlaying = true
      go func() {
          for h.isPlaying {
              h.PlayOnce()
              interval := CalculateHeartbeatInterval(san)
              time.Sleep(interval)
          }
      }()
  }
  ```

### SFX 標記解析

- **LLM 回應範例**:
  ```
  你推開厚重的木門，發出刺耳的吱嘎聲。[SFX:door_open]
  遠處傳來腳步聲。[SFX:footsteps]
  ```

- **解析邏輯**:
  ```go
  func ParseSFXTags(text string) []string {
      re := regexp.MustCompile(`\[SFX:(\w+)\]`)
      matches := re.FindAllStringSubmatch(text, -1)
      var tags []string
      for _, match := range matches {
          tags = append(tags, match[1])
      }
      return tags
  }
  ```

- **SFX 映射表**:
  ```go
  var SFXFiles = map[string]string{
      "door_open":    "door_open.wav",
      "door_close":   "door_close.wav",
      "footsteps":    "footsteps.wav",
      "glass_break":  "glass_break.wav",
      "thunder":      "thunder.wav",
      "whisper":      "whisper.wav",
  }
  ```

### 混音實作

- **通道分配邏輯**:
  ```go
  func (p *SFXPlayer) AllocateChannel(priority int) *SFXChannel {
      // 1. 尋找空閒通道
      for _, ch := range p.channels {
          if !ch.isPlaying {
              return ch
          }
      }

      // 2. 尋找優先級更低的通道
      for _, ch := range p.channels {
          if ch.priority < priority {
              ch.Stop()
              return ch
          }
      }

      // 3. 替換最舊的同優先級通道
      oldest := p.channels[0]
      for _, ch := range p.channels {
          if ch.priority <= priority && ch.startTime.Before(oldest.startTime) {
              oldest = ch
          }
      }
      oldest.Stop()
      return oldest
  }
  ```

### 音訊檔案結構

```
~/.nightmare/audio/sfx/
├── heartbeat.wav          # 心跳（單次，循環播放）
├── warning.wav            # 警告音效
├── death.wav              # 死亡音效
├── door_open.wav          # 門開啟
├── door_close.wav         # 門關閉
├── footsteps.wav          # 腳步聲
├── glass_break.wav        # 玻璃碎裂
├── thunder.wav            # 雷聲
└── whisper.wav            # 耳語
```

### 技術約束

- **通道數**: 4 個（足夠大部分情境）
- **效能目標**:
  - 記憶體 < 20MB（4 個音效同時載入）
  - CPU < 2%
  - 觸發延遲 < 50ms
- **音效長度**: 單個音效 < 5 秒（環境音效），< 3 秒（警告/死亡）
- **格式**: WAV 優先（無損、低延遲），支援 OGG

### 依賴關係

- **前置依賴**:
  - Story 6.4 (BGM 系統 - AudioManager 基礎)
  - Story 3.2 (規則觸發檢測)
  - Story 3.3 (死亡流程)
  - Story 6.1 (SAN 視覺效果系統 - EventBus SAN 變化)
- **整合點**:
  - EventBus 監聽 SAN 變化（心跳）
  - 規則系統觸發警告
  - 死亡流程觸發死亡音效

### 測試策略

1. **單元測試**:
   - 測試通道分配邏輯
   - 測試優先級系統
   - 測試心跳 BPM 計算
2. **整合測試**:
   - 測試 SFX 與 BGM 混音
   - 測試多個音效同時播放
   - 測試心跳隨 SAN 變化
3. **效能測試**:
   - 監控記憶體與 CPU 使用
   - 測試觸發延遲 < 50ms
4. **錯誤處理測試**:
   - 測試音訊檔案缺失
   - 測試損壞檔案
5. **沉浸感測試**:
   - 手動驗證音效增強恐怖體驗
   - 確認音效不過於突兀

### References

- [Source: docs/epics.md#Epic-6]
- [Architecture: ARCHITECTURE.md - oto v3 音訊整合]
- [Story 6.4: BGM 系統 - AudioManager 基礎]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development - all acceptance criteria defined
- 4-channel mixing system with priority queue
- Heartbeat BPM dynamically adjusts based on SAN value
- SFX tag parsing from LLM responses for environment sounds
- Seamless integration with Story 6.4 BGM system
