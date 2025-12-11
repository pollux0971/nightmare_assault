# Story 6.7: 自訂 BGM 支援

Status: ready-for-dev

## Story

As a 玩家,
I want 使用自己的音樂檔案,
so that 我可以個人化恐怖體驗.

## Acceptance Criteria

### AC1: 自訂音樂目錄偵測

**Given** 遊戲音訊系統初始化（Story 6.4）
**When** 檢查音訊檔案
**Then** 掃描 `~/.nightmare/audio/custom/` 目錄
**And** 偵測所有支援格式的音樂檔案（.mp3, .ogg, .wav）
**And** 記錄檔案清單與路徑
**And** 若目錄不存在，自動建立
**And** 若無自訂音樂，靜默跳過（不顯示錯誤）

### AC2: 支援的音訊格式

**Given** 自訂音樂目錄中有音訊檔案
**When** 載入檔案
**Then** 支援格式：
  - .mp3（MPEG Audio Layer 3）
  - .ogg（Ogg Vorbis）
  - .wav（Waveform Audio）
**And** 不支援的格式（如 .flac, .m4a）顯示警告並跳過
**And** 檔案大小限制：單檔 < 20MB
**And** 超過大小限制顯示警告並跳過

### AC3: 自訂 BGM 配置介面

**Given** 玩家在設定選單或遊戲中
**When** 輸入 `/bgm custom`
**Then** 顯示自訂 BGM 配置介面：
  - 列出所有偵測到的自訂音樂（檔名）
  - 列出所有場景類型（探索/緊張/安全/恐怖/解謎/結局）
  - 提供對應設定選項
**And** 玩家可選擇「為 [場景] 使用 [音樂檔案]」
**And** 設定即時生效

### AC4: 場景對應設定

**Given** 玩家配置自訂 BGM
**When** 選擇場景與音樂的對應
**Then** 可為每個場景類型指定一首自訂音樂：
  - 探索場景 → `my_ambient.mp3`
  - 緊張場景 → `my_tension.ogg`
  - 安全場景 → `my_safe.wav`
  - ... (其他場景類型)
**And** 可選擇「使用預設」（恢復標準 BGM）
**And** 設定儲存至 `~/.nightmare/config.json`

### AC5: 配置持久化

**Given** 玩家配置自訂 BGM
**When** 儲存設定
**Then** 配置寫入 `~/.nightmare/config.json`:
  ```json
  {
    "audio": {
      "custom_bgm": {
        "exploration": "custom/my_ambient.mp3",
        "tension": "custom/my_tension.ogg",
        "safe": "default",
        "horror": "custom/scary_music.wav",
        "mystery": "default",
        "ending": "default"
      }
    }
  }
  ```
**And** 下次啟動遊戲時自動載入配置
**And** 若自訂檔案不存在，自動回退至預設 BGM

### AC6: 自動回退機制

**Given** 配置指定自訂 BGM
**When** 遊戲嘗試播放該檔案
**Then** 若檔案存在且有效，播放自訂 BGM
**And** 若檔案不存在或損壞，自動回退至預設 BGM
**And** 記錄警告至 `~/.nightmare/debug.log`
**And** 顯示一次性提示「自訂 BGM 載入失敗，使用預設音樂」
**And** 不中斷遊戲進程

### AC7: 混合使用預設與自訂

**Given** 玩家配置部分場景使用自訂 BGM
**When** 場景切換
**Then** 自訂場景播放自訂 BGM
**And** 預設場景播放標準 BGM
**And** 無縫切換（使用淡入淡出）
**And** 玩家可隨時調整配置

### AC8: 音樂品質驗證

**Given** 載入自訂 BGM 檔案
**When** 檢查音訊品質
**Then** 驗證音訊可解碼（無損壞）
**And** 驗證 sample rate 相容（支援 22050-48000 Hz）
**And** 驗證 channels（單聲道或立體聲）
**And** 不支援的規格顯示警告但嘗試播放
**And** 完全無法解碼時回退至預設

### AC9: 自訂 BGM 清單顯示

**Given** 玩家在遊戲中
**When** 輸入 `/bgm list`
**Then** 顯示所有 BGM（預設 + 自訂）：
  - 場景類型
  - 當前配置（預設 / 自訂檔名）
  - 播放狀態（正在播放 / 未播放）
**And** 標記自訂 BGM（如 `[自訂]`）
**And** 顯示自訂音樂檔案大小與格式

### AC10: 無障礙模式相容

**Given** 玩家啟用無障礙模式
**When** 使用自訂 BGM
**Then** 功能完全相同（不受影響）
**And** 若啟用音訊描述，顯示：
  - "當前播放：自訂音樂 - my_ambient.mp3"
**And** 配置介面保持可讀性
**And** 錯誤提示清晰易懂

## Tasks / Subtasks

- [ ] Task 1: 實作自訂音樂目錄掃描 (AC: #1)
  - [ ] Subtask 1.1: 擴展 `internal/audio/manager.go`
  - [ ] Subtask 1.2: 實作 `ScanCustomAudio() []AudioFile` 函數
  - [ ] Subtask 1.3: 掃描 `~/.nightmare/audio/custom/` 目錄
  - [ ] Subtask 1.4: 過濾支援格式（.mp3, .ogg, .wav）
  - [ ] Subtask 1.5: 若目錄不存在，自動建立
  - [ ] Subtask 1.6: 記錄檔案清單（檔名、路徑、大小）

- [ ] Task 2: 實作格式驗證與大小檢查 (AC: #2, #8)
  - [ ] Subtask 2.1: 實作 `ValidateAudioFile(path string) error` 函數
  - [ ] Subtask 2.2: 檢查檔案副檔名
  - [ ] Subtask 2.3: 檢查檔案大小 < 20MB
  - [ ] Subtask 2.4: 驗證音訊可解碼（使用解碼庫嘗試開啟）
  - [ ] Subtask 2.5: 驗證 sample rate 與 channels 相容性
  - [ ] Subtask 2.6: 記錄不支援檔案的警告

- [ ] Task 3: 建立自訂 BGM 配置 UI (AC: #3)
  - [ ] Subtask 3.1: 建立 `/bgm custom` 指令處理器
  - [ ] Subtask 3.2: 建立 `internal/tui/views/bgm_config.go`
  - [ ] Subtask 3.3: 顯示自訂音樂清單（檔名、大小、格式）
  - [ ] Subtask 3.4: 顯示場景類型清單（6 種場景）
  - [ ] Subtask 3.5: 實作場景與音樂的對應選擇介面
  - [ ] Subtask 3.6: 提供「使用預設」選項

- [ ] Task 4: 實作場景對應邏輯 (AC: #4)
  - [ ] Subtask 4.1: 建立 `CustomBGMConfig` 結構體
  - [ ] Subtask 4.2: 實作場景到檔案的映射（map[string]string）
  - [ ] Subtask 4.3: 實作配置更新函數 `SetCustomBGM(scene, filepath string)`
  - [ ] Subtask 4.4: 實作配置重置函數 `ResetToDefault(scene string)`
  - [ ] Subtask 4.5: 即時生效（更新 AudioManager 配置）

- [ ] Task 5: 實作配置持久化 (AC: #5)
  - [ ] Subtask 5.1: 擴展 `internal/config/` 模組
  - [ ] Subtask 5.2: 添加 `CustomBGMConfig` 到 `AudioConfig`
  - [ ] Subtask 5.3: 實作配置序列化至 JSON
  - [ ] Subtask 5.4: 實作配置反序列化與載入
  - [ ] Subtask 5.5: 啟動時載入自訂 BGM 配置

- [ ] Task 6: 實作自動回退機制 (AC: #6)
  - [ ] Subtask 6.1: 擴展 `internal/audio/bgm_player.go`
  - [ ] Subtask 6.2: 實作 `LoadBGM(scene string) error` 函數
  - [ ] Subtask 6.3: 優先嘗試載入自訂 BGM
  - [ ] Subtask 6.4: 失敗時回退至預設 BGM
  - [ ] Subtask 6.5: 記錄警告至 debug.log
  - [ ] Subtask 6.6: 顯示一次性錯誤提示

- [ ] Task 7: 實作混合播放 (AC: #7)
  - [ ] Subtask 7.1: 整合 Story 6.4 場景切換邏輯
  - [ ] Subtask 7.2: 根據場景檢查是否有自訂 BGM
  - [ ] Subtask 7.3: 使用淡入淡出切換（自訂 ↔ 預設）
  - [ ] Subtask 7.4: 測試混合播放的無縫性

- [ ] Task 8: 實作 BGM 清單顯示 (AC: #9)
  - [ ] Subtask 8.1: 擴展 `/bgm list` 指令
  - [ ] Subtask 8.2: 顯示所有場景的 BGM 配置
  - [ ] Subtask 8.3: 標記自訂 BGM（`[自訂]`）
  - [ ] Subtask 8.4: 顯示檔案大小與格式
  - [ ] Subtask 8.5: 標示當前播放狀態

- [ ] Task 9: 音訊解碼整合 (AC: #2, #8)
  - [ ] Subtask 9.1: 整合 MP3 解碼庫（如 `github.com/hajimehoshi/go-mp3`）
  - [ ] Subtask 9.2: 整合 OGG 解碼庫（如 `github.com/jfreymuth/oggvorbis`）
  - [ ] Subtask 9.3: 整合 WAV 解碼庫（Go 標準庫或第三方）
  - [ ] Subtask 9.4: 實作統一解碼介面
  - [ ] Subtask 9.5: 測試不同格式的播放

- [ ] Task 10: 無障礙模式與測試 (AC: #10)
  - [ ] Subtask 10.1: 確認無障礙模式下功能正常
  - [ ] Subtask 10.2: 音訊描述顯示自訂音樂檔名
  - [ ] Subtask 10.3: 配置介面保持可讀性
  - [ ] Subtask 10.4: 編寫整合測試覆蓋所有 AC
  - [ ] Subtask 10.5: 玩家體驗測試（配置流程是否直觀）

## Dev Notes

### 架構模式

- **模組位置**:
  - 核心邏輯: `internal/audio/manager.go` (擴展)
  - 配置 UI: `internal/tui/views/bgm_config.go`
  - 配置結構: `internal/config/audio.go` (擴展)

- **核心資料結構**:
  ```go
  type CustomBGMConfig struct {
      Exploration string // "custom/my_ambient.mp3" 或 "default"
      Tension     string
      Safe        string
      Horror      string
      Mystery     string
      Ending      string
  }

  type AudioFile struct {
      Filename string
      Path     string
      Size     int64
      Format   string // "mp3", "ogg", "wav"
  }
  ```

- **場景映射**:
  ```go
  const (
      SceneExploration = "exploration"
      SceneTension     = "tension"
      SceneSafe        = "safe"
      SceneHorror      = "horror"
      SceneMystery     = "mystery"
      SceneEnding      = "ending"
  )
  ```

### 自訂 BGM 載入邏輯

```go
func (p *BGMPlayer) LoadBGM(scene string) error {
    // 1. 檢查是否有自訂配置
    customPath := p.config.CustomBGM[scene]
    if customPath != "" && customPath != "default" {
        // 2. 嘗試載入自訂 BGM
        if err := p.load(customPath); err == nil {
            return nil
        } else {
            log.Warn("Custom BGM load failed, falling back to default", err)
            // 3. 失敗時回退至預設
        }
    }

    // 4. 載入預設 BGM
    defaultPath := p.getDefaultBGM(scene)
    return p.load(defaultPath)
}
```

### 配置 JSON 範例

```json
{
  "audio": {
    "bgm_enabled": true,
    "bgm_volume": 70,
    "sfx_enabled": true,
    "sfx_volume": 80,
    "custom_bgm": {
      "exploration": "custom/my_ambient.mp3",
      "tension": "custom/my_tension.ogg",
      "safe": "default",
      "horror": "custom/scary_music.wav",
      "mystery": "default",
      "ending": "default"
    }
  }
}
```

### 自訂 BGM 配置 UI 範例

```
╔══════════════════════════════════════════════════════════╗
║              自訂 BGM 配置                               ║
╠══════════════════════════════════════════════════════════╣
║ 偵測到 3 個自訂音樂檔案：                                ║
║  1. my_ambient.mp3 (3.2 MB)                             ║
║  2. my_tension.ogg (4.5 MB)                             ║
║  3. scary_music.wav (8.1 MB)                            ║
║                                                          ║
║ 場景配置：                                               ║
║  探索場景: [1] my_ambient.mp3                           ║
║  緊張場景: [2] my_tension.ogg                           ║
║  安全場景: [預設]                                        ║
║  恐怖場景: [3] scary_music.wav                          ║
║  解謎場景: [預設]                                        ║
║  結局場景: [預設]                                        ║
║                                                          ║
║ [S] 儲存配置  [R] 重置全部  [Q] 返回                     ║
╚══════════════════════════════════════════════════════════╝
```

### 目錄結構

```
~/.nightmare/audio/
├── bgm/                    # 標準 BGM（Story 6.4）
│   ├── ambient_exploration.ogg
│   ├── tension_chase.ogg
│   └── ...
├── sfx/                    # 音效（Story 6.5）
│   ├── heartbeat.wav
│   └── ...
└── custom/                 # 自訂音樂（本 Story）
    ├── my_ambient.mp3
    ├── my_tension.ogg
    ├── scary_music.wav
    └── ... (玩家自行放置)
```

### 格式支援詳情

| 格式 | 解碼庫 | 優先級 | 備註 |
|------|--------|--------|------|
| .ogg | `github.com/jfreymuth/oggvorbis` | 高 | 更好的循環支援 |
| .mp3 | `github.com/hajimehoshi/go-mp3` | 中 | 廣泛使用 |
| .wav | Go 標準庫 / `github.com/go-audio/wav` | 低 | 檔案較大 |

### 技術約束

- **檔案大小**: 單檔 < 20MB（避免記憶體膨脹）
- **Sample Rate**: 支援 22050-48000 Hz（自動重採樣若需要）
- **Channels**: 單聲道或立體聲（自動轉換若需要）
- **效能**: 載入延遲 < 500ms（與標準 BGM 相同）
- **容錯**: 自動回退至預設 BGM（不中斷遊戲）

### 依賴關係

- **前置依賴**:
  - Story 6.4 (BGM 系統 - AudioManager 基礎)
- **擴展點**:
  - `internal/audio/manager.go` (添加自訂掃描)
  - `internal/config/audio.go` (添加自訂配置)
  - `internal/tui/views/` (添加配置 UI)

### 測試策略

1. **檔案掃描測試**:
   - 測試目錄不存在時自動建立
   - 測試混合格式檔案掃描
2. **格式驗證測試**:
   - 測試支援格式（MP3/OGG/WAV）
   - 測試不支援格式（FLAC/M4A）
   - 測試損壞檔案
3. **配置測試**:
   - 測試配置儲存與載入
   - 測試部分自訂 + 部分預設
4. **回退測試**:
   - 測試自訂檔案不存在時回退
   - 測試自訂檔案損壞時回退
5. **整合測試**:
   - 測試完整配置流程
   - 測試混合播放（自訂 ↔ 預設）
6. **效能測試**:
   - 測試載入延遲 < 500ms
   - 測試記憶體使用（與標準 BGM 相近）

### 使用者文檔（開發後提供）

```markdown
# 自訂 BGM 使用指南

## 快速開始

1. 將音樂檔案放入 `~/.nightmare/audio/custom/` 目錄
2. 支援格式：MP3, OGG, WAV
3. 啟動遊戲，輸入 `/bgm custom` 配置
4. 選擇場景與音樂的對應關係
5. 儲存配置，立即生效

## 建議

- 使用 OGG 格式（更好的循環支援）
- 檔案大小控制在 5-10MB
- 音樂長度 2-5 分鐘（自動循環）
- 音量標準化（避免過大或過小）

## 範例配置

- 探索場景：平靜/神秘的氛圍音樂
- 緊張場景：快節奏/不和諧的音樂
- 安全場景：舒緩/溫暖的音樂
- 恐怖場景：尖銳/驚悚的音樂
```

### References

- [Source: docs/epics.md#Epic-6]
- [Story 6.4: BGM 系統 - AudioManager 基礎]
- [Architecture: ARCHITECTURE.md - 音訊系統架構]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
- Ready for development - all acceptance criteria defined
- Custom BGM directory scanning and validation logic
- Configuration UI for scene-to-music mapping
- Automatic fallback to default BGM on errors
- Seamless mixing of custom and default BGM
- Full integration with Story 6.4 BGM system
