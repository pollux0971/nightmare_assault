# Story 6.4: BGM 系統

Status: ready-for-dev

## Story

As a 玩家,
I want 聽到符合氛圍的背景音樂,
so that 音效增強恐怖體驗.

## Acceptance Criteria

### AC1: 音訊系統初始化

**Given** 遊戲啟動
**When** 初始化音訊系統
**Then** 檢測 `~/.nightmare/audio/` 目錄是否存在
**And** 檢測標準音訊包檔案是否完整（6 BGM + 10 SFX）
**And** 若檔案不完整，顯示提示：
  - "音訊檔案未安裝，執行 'nightmare --download-audio' 下載"
  - "或繼續以靜音模式遊玩"
**And** 初始化 oto v3 音訊上下文（sample rate 48000Hz, 2 channels）

### AC2: 音訊下載功能

**Given** 玩家執行 `nightmare --download-audio standard`
**When** 下載開始
**Then** 顯示進度條（使用 BubbleTea progressbar）
**And** 從 GitHub Releases 或 CDN 下載標準音訊包（zip 格式）
**And** 驗證下載檔案的 checksum（SHA256）
**And** 解壓縮至 `~/.nightmare/audio/`
**And** 下載完成後顯示「音訊安裝完成，共 16 個檔案」

**Given** 下載過程發生錯誤
**When** 網路失敗或 checksum 不符
**Then** 顯示友善錯誤訊息
**And** 清理不完整的下載檔案
**And** 建議用戶稍後重試或手動下載

### AC3: BGM 自動切換

**Given** 遊戲進行中且音訊已安裝
**When** 場景氛圍變化（由 LLM 回應中的 mood 標記決定）
**Then** 自動切換對應 BGM：
  - `ambient_exploration.mp3`: 探索場景（預設）
  - `tension_chase.mp3`: 緊張/追逐場景
  - `safe_rest.mp3`: 安全區/休息場景
  - `horror_reveal.mp3`: 恐怖揭露時刻
  - `mystery_puzzle.mp3`: 解謎場景
  - `ending_death.mp3`: 死亡/結局場景
**And** 切換時使用 1-2 秒淡入淡出（crossfade）
**And** 避免頻繁切換（同一 BGM 至少持續 30 秒）

### AC4: BGM 循環播放

**Given** BGM 正在播放
**When** 音訊檔案播放結束
**Then** 自動循環播放（seamless loop）
**And** 循環點設定正確（避免明顯斷層）
**And** 持續播放直到場景切換或停止指令

### AC5: BGM 控制指令

**Given** 玩家在遊戲中
**When** 輸入 `/bgm off`
**Then** 停止當前 BGM 播放（1 秒淡出）
**And** 後續場景不自動播放 BGM
**And** 設定儲存至配置檔

**Given** BGM 已停用
**When** 輸入 `/bgm on`
**Then** 重新啟用 BGM
**And** 根據當前場景播放對應 BGM（1 秒淡入）

**Given** 玩家在遊戲中
**When** 輸入 `/bgm volume 50`
**Then** 設定 BGM 音量為 50%
**And** 有效範圍 0-100
**And** 即時生效（不需重新播放）
**And** 設定儲存至配置檔

**Given** 玩家在遊戲中
**When** 輸入 `/bgm list`
**Then** 顯示所有可用 BGM 清單與當前播放狀態

### AC6: 靜默失敗機制

**Given** 音訊檔案不存在或損壞
**When** 嘗試播放 BGM
**Then** 靜默失敗（不中斷遊戲）
**And** 記錄警告日誌到 `~/.nightmare/debug.log`
**And** 顯示一次性提示「BGM 播放失敗，繼續靜音模式」
**And** 後續不重複提示

### AC7: 音訊格式支援

**Given** 音訊系統
**When** 載入 BGM 檔案
**Then** 支援格式：MP3, OGG, WAV
**And** 優先使用 OGG（更好的循環支援）
**And** 檔案大小限制：單檔 < 10MB
**And** 不支援的格式顯示警告並跳過

### AC8: 效能與資源管理

**Given** BGM 系統運作中
**When** 監控資源使用
**Then** 記憶體使用 < 50MB（同時載入 2 個 BGM 用於淡入淡出）
**And** CPU 使用 < 3%（解碼與播放）
**And** 切換 BGM 時延遲 < 200ms
**And** 不阻塞主執行緒（使用 goroutine）

### AC9: 配置持久化

**Given** 玩家調整 BGM 設定
**When** 修改音量或啟用狀態
**Then** 設定儲存至 `~/.nightmare/config.json`
**And** 格式：
  ```json
  {
    "audio": {
      "bgm_enabled": true,
      "bgm_volume": 70,
      "sfx_enabled": true,
      "sfx_volume": 80
    }
  }
  ```
**And** 下次啟動遊戲時自動載入設定

### AC10: 無障礙模式相容

**Given** 玩家啟用無障礙模式
**When** BGM 系統運作
**Then** 音訊功能完全正常（不受影響）
**And** 提供音訊描述選項（在設定中）：
  - "當前播放：探索氛圍音樂"
  - "BGM 切換為：緊張追逐"
**And** 音訊描述可選開關

## Tasks / Subtasks

- [ ] Task 1: 建立音訊管理器基礎架構 (AC: #1, #8)
  - [ ] Subtask 1.1: 建立 `internal/audio/manager.go`
  - [ ] Subtask 1.2: 定義 `AudioManager` 結構體（包含 oto context、BGM player、配置）
  - [ ] Subtask 1.3: 實作 `Initialize()` 初始化 oto v3 上下文（48000Hz, 2 channels）
  - [ ] Subtask 1.4: 實作音訊檔案檢測 `CheckAudioFiles() bool`
  - [ ] Subtask 1.5: 使用 goroutine 避免阻塞主執行緒

- [ ] Task 2: 實作音訊下載功能 (AC: #2)
  - [ ] Subtask 2.1: 建立 `cmd/nightmare/audio_downloader.go`
  - [ ] Subtask 2.2: 實作 `--download-audio` 命令行參數處理
  - [ ] Subtask 2.3: 整合 HTTP 下載與進度條（BubbleTea progressbar）
  - [ ] Subtask 2.4: 實作 SHA256 checksum 驗證
  - [ ] Subtask 2.5: 實作 ZIP 解壓縮至 `~/.nightmare/audio/`
  - [ ] Subtask 2.6: 錯誤處理與清理邏輯

- [ ] Task 3: 實作 BGM 播放器 (AC: #3, #4)
  - [ ] Subtask 3.1: 建立 `internal/audio/bgm_player.go`
  - [ ] Subtask 3.2: 實作 `Play(filename string)` 播放 BGM
  - [ ] Subtask 3.3: 整合 oto v3 player（使用 `oto.NewContext()` 與 `Player`）
  - [ ] Subtask 3.4: 實作循環播放邏輯（seamless loop）
  - [ ] Subtask 3.5: 實作音訊檔案解碼（MP3/OGG/WAV）
  - [ ] Subtask 3.6: 測試不同格式的播放與循環

- [ ] Task 4: 實作淡入淡出切換 (AC: #3)
  - [ ] Subtask 4.1: 實作 `FadeOut(duration time.Duration)` 淡出效果
  - [ ] Subtask 4.2: 實作 `FadeIn(duration time.Duration)` 淡入效果
  - [ ] Subtask 4.3: 實作 `Crossfade(newBGM string, duration time.Duration)` 交叉淡入淡出
  - [ ] Subtask 4.4: 使用音量控制實現淡化（線性或對數曲線）
  - [ ] Subtask 4.5: 測試淡入淡出的平滑度

- [ ] Task 5: 實作場景氛圍偵測與自動切換 (AC: #3)
  - [ ] Subtask 5.1: 擴展 LLM 回應解析（Fast Model）偵測 mood 標記
  - [ ] Subtask 5.2: 定義 mood 到 BGM 的映射表
  - [ ] Subtask 5.3: 實作 `SwitchBGMByMood(mood string)` 函數
  - [ ] Subtask 5.4: 實作防止頻繁切換邏輯（最小持續 30 秒）
  - [ ] Subtask 5.5: 整合至遊戲主循環

- [ ] Task 6: 實作 BGM 控制指令 (AC: #5)
  - [ ] Subtask 6.1: 建立 `/bgm` 指令處理器
  - [ ] Subtask 6.2: 實作 `/bgm off` 停止播放
  - [ ] Subtask 6.3: 實作 `/bgm on` 啟用播放
  - [ ] Subtask 6.4: 實作 `/bgm volume <0-100>` 音量調整
  - [ ] Subtask 6.5: 實作 `/bgm list` 顯示 BGM 清單
  - [ ] Subtask 6.6: 即時生效與配置持久化

- [ ] Task 7: 實作音量控制 (AC: #5, #8)
  - [ ] Subtask 7.1: 實作 `SetVolume(volume float64)` 音量設定（0.0-1.0）
  - [ ] Subtask 7.2: 整合至 oto player 的音量控制
  - [ ] Subtask 7.3: 實作即時音量調整（不需重新播放）
  - [ ] Subtask 7.4: 測試音量變化的平滑度

- [ ] Task 8: 實作配置管理 (AC: #9)
  - [ ] Subtask 8.1: 擴展 `internal/config/` 模組
  - [ ] Subtask 8.2: 添加 `AudioConfig` 結構體
  - [ ] Subtask 8.3: 實作配置載入與儲存至 `~/.nightmare/config.json`
  - [ ] Subtask 8.4: 啟動時載入音訊偏好設定

- [ ] Task 9: 實作靜默失敗機制 (AC: #6)
  - [ ] Subtask 9.1: 實作錯誤處理不中斷遊戲
  - [ ] Subtask 9.2: 記錄警告至 `~/.nightmare/debug.log`
  - [ ] Subtask 9.3: 顯示一次性提示（使用標誌避免重複）
  - [ ] Subtask 9.4: 測試音訊檔案缺失情境

- [ ] Task 10: 效能優化與測試 (AC: #7, #8)
  - [ ] Subtask 10.1: 測試記憶體使用 < 50MB
  - [ ] Subtask 10.2: 測試 CPU 使用 < 3%
  - [ ] Subtask 10.3: 測試 BGM 切換延遲 < 200ms
  - [ ] Subtask 10.4: 測試長時間播放的穩定性（2+ 小時）
  - [ ] Subtask 10.5: 測試不同格式（MP3/OGG/WAV）的相容性

- [ ] Task 11: 無障礙模式音訊描述 (AC: #10)
  - [ ] Subtask 11.1: 添加音訊描述設定選項
  - [ ] Subtask 11.2: 實作 BGM 變化時的文字提示
  - [ ] Subtask 11.3: 整合至無障礙模式
  - [ ] Subtask 11.4: 測試音訊描述功能

## Dev Notes

### 架構模式

- **模組位置**:
  - 核心管理: `internal/audio/manager.go`
  - BGM 播放器: `internal/audio/bgm_player.go`
  - 下載器: `cmd/nightmare/audio_downloader.go`

- **核心結構體**:
  ```go
  type AudioManager struct {
      ctx          *oto.Context
      bgmPlayer    *BGMPlayer
      sfxPlayer    *SFXPlayer // Story 6.5 使用
      config       AudioConfig
      initialized  bool
      errorShown   bool // 防止重複錯誤提示
  }

  type BGMPlayer struct {
      currentPlayer *oto.Player
      currentBGM    string
      volume        float64
      enabled       bool
      lastSwitch    time.Time // 防止頻繁切換
  }
  ```

- **Mood 到 BGM 映射**:
  ```go
  var MoodToBGM = map[string]string{
      "exploration": "ambient_exploration.mp3",
      "tension":     "tension_chase.mp3",
      "safe":        "safe_rest.mp3",
      "horror":      "horror_reveal.mp3",
      "mystery":     "mystery_puzzle.mp3",
      "ending":      "ending_death.mp3",
  }
  ```

### oto v3 使用

- **初始化**:
  ```go
  ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
      SampleRate:   48000,
      ChannelCount: 2,
      Format:       oto.FormatSignedInt16LE,
  })
  <-ready // 等待初始化完成
  ```

- **播放音訊**:
  ```go
  player := ctx.NewPlayer(reader)
  player.Play()
  ```

- **音量控制**:
  ```go
  // oto v3 不直接支援音量，需要手動調整音訊資料
  func ApplyVolume(samples []int16, volume float64) {
      for i := range samples {
          samples[i] = int16(float64(samples[i]) * volume)
      }
  }
  ```

### 音訊檔案結構

```
~/.nightmare/audio/
├── bgm/
│   ├── ambient_exploration.ogg
│   ├── tension_chase.ogg
│   ├── safe_rest.ogg
│   ├── horror_reveal.ogg
│   ├── mystery_puzzle.ogg
│   └── ending_death.ogg
├── sfx/ (Story 6.5)
│   ├── heartbeat.wav
│   ├── warning.wav
│   ├── death.wav
│   └── ...
└── custom/ (Story 6.7)
    └── (使用者自訂音樂)
```

### 淡入淡出實作

```go
func (p *BGMPlayer) Crossfade(newBGM string, duration time.Duration) {
    steps := 50 // 50 步驟
    interval := duration / time.Duration(steps)

    for i := 0; i < steps; i++ {
        // 舊 BGM 淡出
        oldVolume := p.volume * float64(steps-i) / float64(steps)
        p.currentPlayer.SetVolume(oldVolume)

        // 新 BGM 淡入（如果已載入）
        // newVolume := p.volume * float64(i) / float64(steps)

        time.Sleep(interval)
    }

    p.currentPlayer.Close()
    p.Play(newBGM)
}
```

### 技術約束

- **音訊格式**: 優先 OGG（更好的循環支援），支援 MP3/WAV 作為備選
- **檔案大小**: 單檔 < 10MB（避免記憶體膨脹）
- **效能目標**:
  - 記憶體 < 50MB
  - CPU < 3%
  - 切換延遲 < 200ms
- **並發**: 使用 goroutine 處理播放，不阻塞主執行緒
- **錯誤處理**: 靜默失敗，不中斷遊戲

### 依賴關係

- **前置依賴**:
  - Epic 1 基礎設施（配置系統）
  - Story 2.2 (LLM 回應解析 - mood 標記)
- **外部套件**:
  - `github.com/ebitengine/oto/v3` (音訊播放)
  - 音訊解碼庫（如 `github.com/hajimehoshi/go-mp3`、`github.com/jfreymuth/oggvorbis`）

### 音訊下載策略

- **來源**: GitHub Releases 或 CDN（如 Cloudflare R2）
- **格式**: ZIP 壓縮檔（`nightmare-audio-standard-v1.0.zip`）
- **驗證**: SHA256 checksum 檔案（`nightmare-audio-standard-v1.0.zip.sha256`）
- **大小**: 約 30-50MB（6 BGM + 10 SFX）

### 測試策略

1. **單元測試**:
   - 測試音訊檔案檢測
   - 測試 mood 映射邏輯
2. **整合測試**:
   - 測試完整播放流程
   - 測試淡入淡出切換
3. **效能測試**:
   - 監控記憶體與 CPU 使用
   - 長時間播放穩定性測試
4. **相容性測試**:
   - 測試不同格式（MP3/OGG/WAV）
   - 測試不同作業系統（Windows/macOS/Linux）
5. **錯誤處理測試**:
   - 測試音訊檔案缺失
   - 測試損壞檔案

### References

- [Source: docs/epics.md#Epic-6]
- [Architecture: ARCHITECTURE.md - oto v3 音訊整合]
- [oto v3 Documentation: https://pkg.go.dev/github.com/ebitengine/oto/v3]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development - all acceptance criteria defined
- oto v3 integration details fully specified
- Audio download mechanism with checksum validation
- Crossfade implementation for smooth BGM transitions
