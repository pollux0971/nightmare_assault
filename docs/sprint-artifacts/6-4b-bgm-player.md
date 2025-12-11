# Story 6.4b: BGM 播放器

Status: ready-for-dev

## Story

As a 玩家,
I want 聽到符合氛圍的背景音樂,
so that 音效增強恐怖體驗.

## Acceptance Criteria

### AC1: BGM 循環播放

**Given** BGM 正在播放
**When** 音訊檔案播放結束
**Then** 自動循環播放（seamless loop）
**And** 循環點設定正確（避免明顯斷層）
**And** 持續播放直到場景切換或停止指令

### AC2: BGM 淡入淡出切換

**Given** 需要切換 BGM
**When** 調用切換函數
**Then** 舊 BGM 在 1-2 秒內淡出
**And** 新 BGM 在 1-2 秒內淡入
**And** 使用交叉淡入淡出（crossfade）效果
**And** 切換過程平滑無斷層

### AC3: BGM 控制指令

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

### AC4: 6 種場景 BGM 播放

**Given** 遊戲進行中且音訊已安裝
**When** 需要播放場景 BGM
**Then** 支援以下 6 種 BGM：
  - `ambient_exploration.mp3`: 探索場景（預設）
  - `tension_chase.mp3`: 緊張/追逐場景
  - `safe_rest.mp3`: 安全區/休息場景
  - `horror_reveal.mp3`: 恐怖揭露時刻
  - `mystery_puzzle.mp3`: 解謎場景
  - `ending_death.mp3`: 死亡/結局場景
**And** 每種 BGM 正確對應場景氛圍

### AC5: 音量控制

**Given** BGM 正在播放
**When** 調整音量
**Then** 實作 `SetVolume(volume float64)` 音量設定（0.0-1.0）
**And** 整合至 oto player 的音量控制
**And** 即時音量調整（不需重新播放）
**And** 音量變化平滑（避免突然響度變化）

### AC6: 效能與資源管理

**Given** BGM 系統運作中
**When** 監控資源使用
**Then** 記憶體使用 < 40MB（同時載入 2 個 BGM 用於淡入淡出）
**And** CPU 使用 < 3%（解碼與播放）
**And** 切換 BGM 時延遲 < 200ms
**And** 不阻塞主執行緒（使用 goroutine）

### AC7: 音訊下載功能

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

## Tasks / Subtasks

- [ ] Task 1: 實作 BGM 播放器 (AC: #1, #4)
  - [ ] Subtask 1.1: 建立 `internal/audio/bgm_player.go`
  - [ ] Subtask 1.2: 實作 `Play(filename string)` 播放 BGM
  - [ ] Subtask 1.3: 整合 oto v3 player（使用 `oto.NewContext()` 與 `Player`）
  - [ ] Subtask 1.4: 實作循環播放邏輯（seamless loop）
  - [ ] Subtask 1.5: 實作音訊檔案解碼（MP3/OGG/WAV）
  - [ ] Subtask 1.6: 測試不同格式的播放與循環

- [ ] Task 2: 實作淡入淡出切換 (AC: #2)
  - [ ] Subtask 2.1: 實作 `FadeOut(duration time.Duration)` 淡出效果
  - [ ] Subtask 2.2: 實作 `FadeIn(duration time.Duration)` 淡入效果
  - [ ] Subtask 2.3: 實作 `Crossfade(newBGM string, duration time.Duration)` 交叉淡入淡出
  - [ ] Subtask 2.4: 使用音量控制實現淡化（線性或對數曲線）
  - [ ] Subtask 2.5: 測試淡入淡出的平滑度

- [ ] Task 3: 實作音量控制 (AC: #5)
  - [ ] Subtask 3.1: 實作 `SetVolume(volume float64)` 音量設定（0.0-1.0）
  - [ ] Subtask 3.2: 整合至 oto player 的音量控制
  - [ ] Subtask 3.3: 實作即時音量調整（不需重新播放）
  - [ ] Subtask 3.4: 測試音量變化的平滑度

- [ ] Task 4: 實作 BGM 控制指令 (AC: #3)
  - [ ] Subtask 4.1: 建立 `/bgm` 指令處理器
  - [ ] Subtask 4.2: 實作 `/bgm off` 停止播放
  - [ ] Subtask 4.3: 實作 `/bgm on` 啟用播放
  - [ ] Subtask 4.4: 實作 `/bgm volume <0-100>` 音量調整
  - [ ] Subtask 4.5: 實作 `/bgm list` 顯示 BGM 清單
  - [ ] Subtask 4.6: 即時生效與配置持久化

- [ ] Task 5: 實作音訊下載功能 (AC: #7)
  - [ ] Subtask 5.1: 建立 `cmd/nightmare/audio_downloader.go`
  - [ ] Subtask 5.2: 實作 `--download-audio` 命令行參數處理
  - [ ] Subtask 5.3: 整合 HTTP 下載與進度條（BubbleTea progressbar）
  - [ ] Subtask 5.4: 實作 SHA256 checksum 驗證
  - [ ] Subtask 5.5: 實作 ZIP 解壓縮至 `~/.nightmare/audio/`
  - [ ] Subtask 5.6: 錯誤處理與清理邏輯

- [ ] Task 6: 效能優化與測試 (AC: #6)
  - [ ] Subtask 6.1: 測試記憶體使用 < 40MB
  - [ ] Subtask 6.2: 測試 CPU 使用 < 3%
  - [ ] Subtask 6.3: 測試 BGM 切換延遲 < 200ms
  - [ ] Subtask 6.4: 測試長時間播放的穩定性（2+ 小時）
  - [ ] Subtask 6.5: 測試不同格式（MP3/OGG/WAV）的相容性

## Dev Notes

### 架構模式

- **模組位置**:
  - BGM 播放器: `internal/audio/bgm_player.go`
  - 下載器: `cmd/nightmare/audio_downloader.go`

- **核心結構體**:
  ```go
  type BGMPlayer struct {
      ctx           *oto.Context
      currentPlayer *oto.Player
      nextPlayer    *oto.Player    // For crossfade
      currentBGM    string
      volume        float64
      enabled       bool
      lastSwitch    time.Time       // For 6.4c anti-frequent-switch
      mu            sync.RWMutex
  }
  ```

### oto v3 播放實作

```go
func (p *BGMPlayer) Play(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }

    // Decode based on format
    var decoder io.ReadCloser
    switch filepath.Ext(filename) {
    case ".mp3":
        decoder, err = mp3.NewDecoder(file)
    case ".ogg":
        decoder, err = oggvorbis.NewReader(file)
    case ".wav":
        decoder = file // WAV is raw PCM
    }

    player := p.ctx.NewPlayer(decoder)
    player.Play()

    p.currentPlayer = player
    p.currentBGM = filename
    return nil
}
```

### 淡入淡出實作

```go
func (p *BGMPlayer) Crossfade(newBGM string, duration time.Duration) {
    steps := 50 // 50 steps for smooth fade
    interval := duration / time.Duration(steps)

    // Load new BGM
    p.nextPlayer = p.loadBGM(newBGM)

    for i := 0; i < steps; i++ {
        progress := float64(i) / float64(steps)

        // Fade out old
        oldVolume := p.volume * (1.0 - progress)
        p.currentPlayer.SetVolume(oldVolume)

        // Fade in new
        newVolume := p.volume * progress
        p.nextPlayer.SetVolume(newVolume)

        time.Sleep(interval)
    }

    // Swap players
    p.currentPlayer.Close()
    p.currentPlayer = p.nextPlayer
    p.nextPlayer = nil
    p.currentBGM = newBGM
}
```

### 音訊下載策略

- **來源**: GitHub Releases 或 CDN（如 Cloudflare R2）
- **格式**: ZIP 壓縮檔（`nightmare-audio-standard-v1.0.zip`）
- **驗證**: SHA256 checksum 檔案（`nightmare-audio-standard-v1.0.zip.sha256`）
- **大小**: 約 30-50MB（6 BGM + 10 SFX）

### 技術約束

- **音訊格式**: 優先 OGG（更好的循環支援），支援 MP3/WAV 作為備選
- **效能目標**:
  - 記憶體 < 40MB（2 個 BGM 同時載入用於淡入淡出）
  - CPU < 3%
  - 切換延遲 < 200ms
- **並發**: 使用 goroutine 處理播放，不阻塞主執行緒

### 依賴關係

- **前置依賴**:
  - Story 6.4a (音訊系統基礎架構)
- **後續依賴**:
  - Story 6.4c (BGM 自動切換)
  - Story 6.7 (自訂 BGM 支援)
- **外部套件**:
  - `github.com/ebitengine/oto/v3` (音訊播放)
  - `github.com/hajimehoshi/go-mp3` (MP3 解碼)
  - `github.com/jfreymuth/oggvorbis` (OGG 解碼)

### 測試策略

1. **單元測試**:
   - 測試 BGM 播放與循環
   - 測試淡入淡出邏輯
   - 測試音量控制
2. **整合測試**:
   - 測試完整播放流程
   - 測試 `/bgm` 指令
3. **效能測試**:
   - 監控記憶體與 CPU 使用
   - 長時間播放穩定性測試
4. **相容性測試**:
   - 測試不同格式（MP3/OGG/WAV）
   - 測試不同作業系統（Windows/macOS/Linux）

### References

- [Source: docs/epics.md#Epic-6]
- [Architecture: ARCHITECTURE.md - oto v3 音訊整合]
- [Story 6.4a: 音訊系統基礎架構]
- [oto v3 Documentation: https://pkg.go.dev/github.com/ebitengine/oto/v3]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- Story split from 6-4-bgm-system.md for better manageability
- Focused on BGM player core functionality (play, loop, fade, volume, download)
- Depends on 6.4a (audio foundation)
- Provides base for 6.4c (auto-switching) and 6.7 (custom BGM)
- Ready for development - all acceptance criteria defined
