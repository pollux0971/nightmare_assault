# Story 6.5a: SFX 播放器

Status: ready-for-dev

## Story

As a 玩家,
I want 聽到事件音效,
so that 關鍵時刻有聽覺回饋.

## Acceptance Criteria

### AC1: SFX 播放器基礎功能

**Given** 音訊系統已初始化（Story 6.4a）
**When** 建立 SFX 播放器
**Then** 支援同時播放多個音效（≥4 個通道）
**And** 每個音效獨立控制音量
**And** SFX 與 BGM 混音播放（不互相中斷）
**And** 使用獨立的 oto.Player 實例

### AC2: 4 通道混音系統

**Given** SFX 播放器運作中
**When** 同時觸發多個音效
**Then** 最多同時播放 4 個音效（4 通道）
**And** 每個通道使用獨立的 oto.Player
**And** 通道空閒時自動回收
**And** 通道滿載時根據優先級處理

### AC3: SFX 優先級佇列

**Given** 多個音效同時觸發
**When** SFX 播放器處理
**Then** 使用優先級系統：
  1. 死亡音效（最高優先級 100，停止所有其他 SFX）
  2. 警告音效（高優先級 80，可中斷環境音效）
  3. 心跳音效（中優先級 50，循環播放 - Story 6.5b）
  4. 環境音效（低優先級 20，可被覆蓋）
**And** 同優先級音效使用不同通道同時播放
**And** 超過通道數時，停止最舊的同優先級音效

### AC4: 環境音效觸發

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

### AC5: 警告與死亡音效

**Given** 玩家觸發特定事件
**When** 事件發生
**Then** 播放對應音效：
  - 警告音效（warning.wav）: 規則觸發警告（Story 3.2）
  - 死亡音效（death.wav）: HP 降至 0 或即死（Story 3.3）
**And** 警告音效持續 1-2 秒
**And** 警告音效優先級高於環境音效
**And** 死亡音效播放時停止所有其他 SFX
**And** 死亡音效播放時 BGM 淡出至靜音（2 秒）

### AC6: SFX 控制指令

**Given** 玩家在遊戲中
**When** 輸入 `/sfx off`
**Then** 停止所有 SFX 播放
**And** 後續事件不觸發 SFX
**And** 設定儲存至配置檔

**Given** SFX 已停用
**When** 輸入 `/sfx on`
**Then** 重新啟用 SFX

**Given** 玩家在遊戲中
**When** 輸入 `/sfx volume 60`
**Then** 設定 SFX 音量為 60%
**And** 有效範圍 0-100
**And** 即時生效（不影響已播放的音效）
**And** 設定儲存至配置檔

### AC7: 效能與資源管理

**Given** SFX 系統運作中
**When** 監控資源使用
**Then** 記憶體使用 < 20MB（同時載入 4 個音效）
**And** CPU 使用 < 2%（解碼與播放）
**And** 音效觸發延遲 < 50ms
**And** 不阻塞主執行緒（使用 goroutine）

### AC8: 無障礙模式文字提示

**Given** 玩家啟用無障礙模式且開啟音訊描述
**When** SFX 播放
**Then** 顯示文字描述：
  - 警告：「[警告音效]」
  - 死亡：「[死亡音效]」
  - 環境：「[門開啟聲]」
**And** 文字提示短暫顯示（1-2 秒）於狀態列
**And** 不影響敘事區內容

## Tasks / Subtasks

- [ ] Task 1: 建立 SFX 播放器架構 (AC: #1, #2)
  - [ ] Subtask 1.1: 建立 `internal/audio/sfx_player.go`
  - [ ] Subtask 1.2: 定義 `SFXPlayer` 結構體（包含通道池、優先級佇列）
  - [ ] Subtask 1.3: 實作 4 通道音效混音系統
  - [ ] Subtask 1.4: 實作通道管理（分配、回收、狀態追蹤）
  - [ ] Subtask 1.5: 測試 4 通道同時播放

- [ ] Task 2: 實作優先級系統 (AC: #3)
  - [ ] Subtask 2.1: 定義優先級常數（死亡/警告/心跳/環境）
  - [ ] Subtask 2.2: 實作通道分配邏輯（基於優先級）
  - [ ] Subtask 2.3: 實作音效佇列（等待可用通道）
  - [ ] Subtask 2.4: 實作最舊音效替換邏輯
  - [ ] Subtask 2.5: 測試優先級系統正確運作

- [ ] Task 3: 實作環境音效觸發 (AC: #4)
  - [ ] Subtask 3.1: 擴展 LLM 回應解析（Fast Model）偵測 `[SFX:xxx]` 標記
  - [ ] Subtask 3.2: 建立 SFX 標記到檔案名稱的映射表
  - [ ] Subtask 3.3: 實作 `PlayEnvironmentSFX(tag string)` 函數
  - [ ] Subtask 3.4: 確保環境音效不阻塞敘事顯示
  - [ ] Subtask 3.5: 測試多個環境音效同時播放

- [ ] Task 4: 實作警告與死亡音效 (AC: #5)
  - [ ] Subtask 4.1: 整合 Story 3.2 規則觸發系統（警告音效）
  - [ ] Subtask 4.2: 整合 Story 3.3 死亡流程（死亡音效）
  - [ ] Subtask 4.3: 實作 `PlayWarning()` 函數
  - [ ] Subtask 4.4: 實作 `PlayDeath()` 函數
  - [ ] Subtask 4.5: 實作死亡時停止所有 SFX + BGM 淡出
  - [ ] Subtask 4.6: 測試警告與死亡音效與其他系統整合

- [ ] Task 5: 實作混音與通道管理 (AC: #2, #3, #7)
  - [ ] Subtask 5.1: 實作通道池（4 個 oto.Player）
  - [ ] Subtask 5.2: 實作通道分配邏輯（基於優先級）
  - [ ] Subtask 5.3: 實作音效佇列（等待可用通道）
  - [ ] Subtask 5.4: 測試 4+ 個音效同時播放的處理
  - [ ] Subtask 5.5: 效能測試（確保 < 20MB 記憶體、< 2% CPU）

- [ ] Task 6: 實作 SFX 控制指令 (AC: #6)
  - [ ] Subtask 6.1: 建立 `/sfx` 指令處理器
  - [ ] Subtask 6.2: 實作 `/sfx off` 停用所有 SFX
  - [ ] Subtask 6.3: 實作 `/sfx on` 啟用 SFX
  - [ ] Subtask 6.4: 實作 `/sfx volume <0-100>` 音量調整
  - [ ] Subtask 6.5: 配置持久化至 `~/.nightmare/config.json`

- [ ] Task 7: 實作音量控制 (AC: #6, #7)
  - [ ] Subtask 7.1: 實作 `SetVolume(volume float64)` 全域 SFX 音量
  - [ ] Subtask 7.2: 整合至音訊資料處理（手動音量調整）
  - [ ] Subtask 7.3: 測試音量變化的即時生效
  - [ ] Subtask 7.4: 測試音量與優先級的正確混音

- [ ] Task 8: 實作無障礙模式音訊描述 (AC: #8)
  - [ ] Subtask 8.1: 整合音訊描述設定（Story 6.4a）
  - [ ] Subtask 8.2: 實作 SFX 文字提示（顯示於狀態列）
  - [ ] Subtask 8.3: 實作短暫顯示邏輯（1-2 秒後消失）
  - [ ] Subtask 8.4: 測試無障礙模式完整流程

- [ ] Task 9: 整合測試與調校 (AC: #1-8)
  - [ ] Subtask 9.1: 測試所有 SFX 與 BGM 混音
  - [ ] Subtask 9.2: 測試優先級系統正確運作
  - [ ] Subtask 9.3: 測試環境音效觸發
  - [ ] Subtask 9.4: 效能測試（記憶體/CPU/延遲）
  - [ ] Subtask 9.5: 完整遊戲流程測試

## Dev Notes

### 架構模式

- **模組位置**: `internal/audio/sfx_player.go`

- **核心結構體**:
  ```go
  type SFXPlayer struct {
      ctx          *oto.Context
      channels     [4]*SFXChannel // 4 個音效通道
      volume       float64
      enabled      bool
      mu           sync.RWMutex
  }

  type SFXChannel struct {
      player      *oto.Player
      priority    int
      startTime   time.Time
      isPlaying   bool
      sfxType     string // For debugging/tracking
  }
  ```

- **優先級定義**:
  ```go
  const (
      PriorityDeath       = 100
      PriorityWarning     = 80
      PriorityHeartbeat   = 50  // Story 6.5b
      PriorityEnvironment = 20
  )
  ```

### 通道分配邏輯

```go
func (p *SFXPlayer) AllocateChannel(priority int) *SFXChannel {
    p.mu.Lock()
    defer p.mu.Unlock()

    // 1. Find free channel
    for _, ch := range p.channels {
        if !ch.isPlaying {
            return ch
        }
    }

    // 2. Find lower priority channel
    for _, ch := range p.channels {
        if ch.priority < priority {
            ch.Stop()
            return ch
        }
    }

    // 3. Replace oldest same-priority channel
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

### 音訊檔案結構

```
~/.nightmare/audio/sfx/
├── heartbeat.wav          # 心跳（Story 6.5b）
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
  - Story 6.4a (音訊系統基礎架構)
  - Story 3.2 (規則觸發檢測 - 警告音效)
  - Story 3.3 (死亡流程 - 死亡音效)
- **後續依賴**:
  - Story 6.5b (心跳系統 - 使用 SFXPlayer)
- **外部套件**:
  - `github.com/ebitengine/oto/v3` (音訊播放)

### 測試策略

1. **單元測試**:
   - 測試通道分配邏輯
   - 測試優先級系統
   - 測試 SFX 標記解析
2. **整合測試**:
   - 測試 SFX 與 BGM 混音
   - 測試多個音效同時播放
   - 測試警告與死亡音效整合
3. **效能測試**:
   - 監控記憶體與 CPU 使用
   - 測試觸發延遲 < 50ms
4. **沉浸感測試**:
   - 手動驗證音效增強恐怖體驗
   - 確認音效不過於突兀

### References

- [Source: docs/epics.md#Epic-6]
- [Architecture: ARCHITECTURE.md - oto v3 音訊整合]
- [Story 6.4a: 音訊系統基礎架構]
- [Story 3.2: 規則觸發檢測]
- [Story 3.3: 死亡流程]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- Story split from 6-5-sfx-system.md for better manageability
- Focused on SFX player core functionality (4-channel mixing, priority queue, environment/warning/death sounds)
- Depends on 6.4a (audio foundation), 3.2 (rule warnings), 3.3 (death flow)
- Provides base for 6.5b (heartbeat system)
- Ready for development - all acceptance criteria defined
