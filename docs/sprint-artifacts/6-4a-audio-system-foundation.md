# Story 6.4a: 音訊系統基礎架構

Status: ready-for-dev

## Story

As a 開發者,
I want 建立音訊系統基礎架構,
so that BGM 和 SFX 功能可以在穩定的基礎上開發.

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

### AC2: 音訊格式支援與檢測

**Given** 音訊系統
**When** 載入音訊檔案
**Then** 支援格式：MP3, OGG, WAV
**And** 優先使用 OGG（更好的循環支援）
**And** 檔案大小限制：單檔 < 10MB
**And** 不支援的格式顯示警告並跳過
**And** 平台檢測（檢查音訊裝置可用性）

### AC3: 靜默失敗機制

**Given** 音訊檔案不存在或損壞
**When** 嘗試播放音訊
**Then** 靜默失敗（不中斷遊戲）
**And** 記錄警告日誌到 `~/.nightmare/debug.log`
**And** 顯示一次性提示「音訊播放失敗，繼續靜音模式」
**And** 後續不重複提示

### AC4: 配置持久化

**Given** 音訊系統設定
**When** 首次初始化
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

### AC5: 效能與資源管理基礎

**Given** 音訊系統運作中
**When** 監控資源使用
**Then** 不阻塞主執行緒（使用 goroutine）
**And** 初始化延遲 < 100ms
**And** 記憶體基礎占用 < 10MB

## Tasks / Subtasks

- [ ] Task 1: 建立音訊管理器基礎架構 (AC: #1, #5)
  - [ ] Subtask 1.1: 建立 `internal/audio/manager.go`
  - [ ] Subtask 1.2: 定義 `AudioManager` 結構體（包含 oto context、BGM player、SFX player、配置）
  - [ ] Subtask 1.3: 實作 `Initialize()` 初始化 oto v3 上下文（48000Hz, 2 channels）
  - [ ] Subtask 1.4: 實作音訊檔案檢測 `CheckAudioFiles() bool`
  - [ ] Subtask 1.5: 使用 goroutine 避免阻塞主執行緒

- [ ] Task 2: 實作平台檢測與優雅降級 (AC: #2, #3)
  - [ ] Subtask 2.1: 檢測音訊裝置可用性（oto context 初始化是否成功）
  - [ ] Subtask 2.2: 實作格式檢測（MP3/OGG/WAV）
  - [ ] Subtask 2.3: 實作檔案大小驗證（< 10MB）
  - [ ] Subtask 2.4: 實作靜默失敗邏輯
  - [ ] Subtask 2.5: 錯誤日誌記錄至 `~/.nightmare/debug.log`

- [ ] Task 3: 實作配置管理 (AC: #4)
  - [ ] Subtask 3.1: 擴展 `internal/config/` 模組
  - [ ] Subtask 3.2: 添加 `AudioConfig` 結構體
  - [ ] Subtask 3.3: 實作配置載入與儲存至 `~/.nightmare/config.json`
  - [ ] Subtask 3.4: 啟動時載入音訊偏好設定
  - [ ] Subtask 3.5: 配置驗證（音量範圍 0-100）

- [ ] Task 4: 單元測試與整合測試 (AC: #1-5)
  - [ ] Subtask 4.1: 測試音訊管理器初始化
  - [ ] Subtask 4.2: 測試音訊檔案檢測邏輯
  - [ ] Subtask 4.3: 測試靜默失敗機制
  - [ ] Subtask 4.4: 測試配置持久化
  - [ ] Subtask 4.5: 測試效能（初始化延遲、記憶體占用）

## Dev Notes

### 架構模式

- **模組位置**:
  - 核心管理: `internal/audio/manager.go`
  - 配置擴展: `internal/config/audio.go`

- **核心結構體**:
  ```go
  type AudioManager struct {
      ctx          *oto.Context
      bgmPlayer    *BGMPlayer    // Story 6.4b 使用
      sfxPlayer    *SFXPlayer    // Story 6.5a 使用
      config       AudioConfig
      initialized  bool
      errorShown   bool // 防止重複錯誤提示
      mu           sync.RWMutex
  }

  type AudioConfig struct {
      BGMEnabled  bool    `json:"bgm_enabled"`
      BGMVolume   int     `json:"bgm_volume"`   // 0-100
      SFXEnabled  bool    `json:"sfx_enabled"`
      SFXVolume   int     `json:"sfx_volume"`   // 0-100
  }
  ```

### oto v3 初始化

```go
func (m *AudioManager) Initialize() error {
    ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
        SampleRate:   48000,
        ChannelCount: 2,
        Format:       oto.FormatSignedInt16LE,
    })
    if err != nil {
        m.handleAudioError(err)
        return err
    }

    // 等待初始化完成（非阻塞，設定 timeout）
    select {
    case <-ready:
        m.ctx = ctx
        m.initialized = true
        return nil
    case <-time.After(100 * time.Millisecond):
        return fmt.Errorf("audio initialization timeout")
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
├── sfx/
│   ├── heartbeat.wav
│   ├── warning.wav
│   ├── death.wav
│   └── ... (7 more)
└── custom/ (Story 6.7)
    └── (使用者自訂音樂)
```

### 靜默失敗實作

```go
func (m *AudioManager) handleAudioError(err error) {
    // 記錄日誌
    log.Printf("[WARN] Audio error: %v\n", err)

    // 顯示一次性提示
    if !m.errorShown {
        fmt.Println("音訊播放失敗，繼續靜音模式")
        m.errorShown = true
    }

    // 不中斷遊戲
}
```

### 技術約束

- **音訊格式**: 優先 OGG（更好的循環支援），支援 MP3/WAV 作為備選
- **檔案大小**: 單檔 < 10MB（避免記憶體膨脹）
- **效能目標**:
  - 初始化延遲 < 100ms
  - 記憶體基礎占用 < 10MB
- **並發**: 使用 goroutine 處理初始化，不阻塞主執行緒
- **錯誤處理**: 靜默失敗，不中斷遊戲

### 依賴關係

- **前置依賴**:
  - Epic 1 基礎設施（配置系統）
- **後續依賴**:
  - Story 6.4b (BGM 播放器)
  - Story 6.5a (SFX 播放器)
- **外部套件**:
  - `github.com/ebitengine/oto/v3` (音訊播放)
  - 音訊解碼庫（如 `github.com/hajimehoshi/go-mp3`、`github.com/jfreymuth/oggvorbis`）

### 測試策略

1. **單元測試**:
   - 測試音訊檔案檢測
   - 測試配置載入/儲存
   - 測試格式驗證
2. **整合測試**:
   - 測試完整初始化流程
   - 測試靜默失敗機制
3. **效能測試**:
   - 監控初始化延遲
   - 測試記憶體占用
4. **錯誤處理測試**:
   - 測試音訊檔案缺失
   - 測試音訊裝置不可用

### References

- [Source: docs/epics.md#Epic-6]
- [Architecture: ARCHITECTURE.md - oto v3 音訊整合]
- [oto v3 Documentation: https://pkg.go.dev/github.com/ebitengine/oto/v3]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

- Story split from 6-4-bgm-system.md for better manageability
- Focused on audio system foundation (initialization, config, error handling)
- Provides base for 6.4b (BGM), 6.4c (auto-switching), and 6.5a (SFX)
- Ready for development - all acceptance criteria defined
