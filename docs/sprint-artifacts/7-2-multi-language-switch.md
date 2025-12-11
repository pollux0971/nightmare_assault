# Story 7.2: 多語言切換

Status: ready-for-dev

## Story

As a 玩家,
I want 在遊戲中切換語言,
so that 我可以使用偏好的語言遊玩.

## Acceptance Criteria

### AC1: 運行時語言切換

**Given** 玩家在設定或遊戲中
**When** 輸入 `/lang zh-TW` 或 `/lang en-US`
**Then** 切換 UI 語言
**And** 即時生效（無需重啟）
**And** 顯示「語言已切換為 [語言名稱]」確認訊息

### AC2: UI 元素語言化

**Given** 語言為英文
**When** 顯示 UI 元素
**Then** 選單、指令說明、系統訊息為英文
**And** 狀態列標籤（HP/SAN/位置）為英文
**And** 錯誤訊息為英文

**Given** 語言為繁體中文
**When** 顯示 UI 元素
**Then** 所有 UI 元素為繁體中文

### AC3: 故事內容語言切換

**Given** 語言為英文
**When** 顯示 UI 元素
**Then** 故事內容請求 LLM 以英文生成
**And** LLM prompt 使用英文模板
**And** 角色名稱、地點名稱使用英文

**Given** 遊戲進行中切換語言
**When** 切換完成
**Then** 已生成的故事保持原語言（不重新生成）
**And** 新內容使用新語言
**And** 顯示提示「新內容將使用 [新語言]」

### AC4: 語言設定持久化

**Given** 語言設定已儲存
**When** 下次啟動遊戲
**Then** 自動載入上次的語言設定
**And** 首次啟動時根據系統語言自動選擇（zh-TW/zh-CN → 繁中，其他 → 英文）

### AC5: 語言選擇 UI

**Given** 玩家在主選單「設定」
**When** 選擇「語言 / Language」
**Then** 顯示語言選擇清單
**And** 列出支援的語言：繁體中文 (zh-TW)、English (en-US)
**And** 當前語言有標記 (✓)
**And** 選擇後立即切換

## Tasks / Subtasks

- [ ] 建立 i18n 基礎設施 (AC: #1, #2)
  - [ ] 建立 `internal/i18n/` 模組
  - [ ] 定義 `Translator` 介面
  - [ ] 實作 `LoadLanguage(locale string)` 函數
  - [ ] 實作 `T(key string, params ...any)` 翻譯函數
  - [ ] 支援參數化翻譯（如 "HP: {hp}/100"）

- [ ] 建立語言檔案 (AC: #2)
  - [ ] 建立 `internal/i18n/locales/zh-TW.json`
  - [ ] 建立 `internal/i18n/locales/en-US.json`
  - [ ] 翻譯所有 UI 文字（選單、指令、系統訊息）
  - [ ] 翻譯錯誤訊息模板
  - [ ] 翻譯遊戲提示文字

- [ ] 整合 UI 元件語言化 (AC: #2)
  - [ ] 修改 `internal/tui/views/menu.go` 使用翻譯
  - [ ] 修改 `internal/tui/views/game.go` 狀態列使用翻譯
  - [ ] 修改所有 TUI 元件使用 `i18n.T()`
  - [ ] 處理複數形式（1 item vs. 2 items）
  - [ ] 處理日期時間格式化

- [ ] LLM Prompt 語言切換 (AC: #3)
  - [ ] 修改 `internal/engine/story.go` 使用語言化 prompt
  - [ ] 建立 `prompts/zh-TW/` 目錄
  - [ ] 建立 `prompts/en-US/` 目錄
  - [ ] 實作 prompt 模板動態載入
  - [ ] 處理混合語言情境（已生成內容 + 新內容）

- [ ] 實作 /lang 指令 (AC: #1, #5)
  - [ ] 建立 `internal/game/commands/lang.go`
  - [ ] 解析語言代碼（zh-TW, en-US）
  - [ ] 驗證語言代碼有效性
  - [ ] 切換全局語言設定
  - [ ] 觸發 UI 重新渲染

- [ ] 語言設定持久化 (AC: #4)
  - [ ] 修改 `~/.nightmare/config.json` 添加 `language` 欄位
  - [ ] 啟動時讀取語言設定
  - [ ] 首次啟動偵測系統語言
  - [ ] 儲存語言變更

- [ ] 測試與驗證
  - [ ] 測試所有 UI 元素翻譯完整性
  - [ ] 測試語言切換即時生效
  - [ ] 測試 LLM prompt 語言切換
  - [ ] 測試邊界情況（無效語言代碼、缺失翻譯）

## Dev Notes

### 架構模式與約束

**架構模式:**
- **JSON 語言檔案**: 使用 JSON 格式儲存翻譯，易於維護和擴展
- **Key-based 翻譯**: 使用語義化 key（如 `menu.new_game`），不使用英文文字作為 key
- **Fallback 機制**: 缺失翻譯時回退到英文
- **參數化翻譯**: 支援 `{param}` 格式的參數替換

**技術約束:**
- 初期僅支援繁體中文和英文，架構需支援未來擴展
- 翻譯檔案大小需控制（< 100KB per locale）
- 語言切換必須是即時的（< 100ms）
- LLM prompt 語言切換需同步至 Game Bible

**NFR 滿足:**
- NFR01 (性能需求): 語言切換 < 100ms ✓
- NFR03 (可用性需求): 支援繁中/英文 UI ✓
- NFR04 (可維護性需求): 易於添加新語言 ✓

**依賴項:**
- Epic 1 基礎設施（配置系統）
- Epic 2 核心遊戲循環（LLM prompt 系統）
- 所有 UI 元件需重構支援 i18n

**風險與緩解:**
- **風險**: 翻譯不完整導致混合語言顯示
  - **緩解**: 實作 fallback 機制，缺失翻譯顯示 key + 警告
- **風險**: LLM 生成內容語言不一致
  - **緩解**: 在 prompt 中明確指定語言要求
- **風險**: 語言切換後 UI 排版錯位
  - **緩解**: 使用響應式佈局，測試不同語言文字長度

### Implementation Details

**i18n 模組架構:**
```go
package i18n

type Translator struct {
    locale       string
    translations map[string]string
    fallback     map[string]string
}

func New(locale string) (*Translator, error)
func (t *Translator) T(key string, params ...any) string
func (t *Translator) SetLocale(locale string) error
```

**語言檔案結構 (JSON):**
```json
{
  "menu": {
    "new_game": "新遊戲",
    "continue_game": "繼續遊戲",
    "settings": "設定",
    "quit": "離開"
  },
  "game": {
    "hp_label": "生命值",
    "san_label": "理智值",
    "location_label": "位置"
  },
  "commands": {
    "status_desc": "查看角色狀態",
    "help_desc": "顯示幫助訊息"
  },
  "errors": {
    "api_failed": "API 請求失敗: {error}",
    "save_failed": "存檔失敗: {reason}"
  }
}
```

**Prompt 語言化範例:**
```
// prompts/zh-TW/game_bible.txt
你是一個恐怖故事的敘事者，請根據以下主題創建故事架構...

// prompts/en-US/game_bible.txt
You are a horror story narrator. Create a story framework based on the following theme...
```

**語言切換流程:**
```go
// 1. 用戶輸入 /lang en-US
// 2. 驗證語言代碼
if !isValidLocale(locale) {
    return error("invalid locale")
}

// 3. 載入新語言檔案
newTranslations := loadTranslations(locale)

// 4. 更新全局 Translator
global.i18n.SetLocale(locale)

// 5. 更新 LLM prompt 路徑
global.promptPath = fmt.Sprintf("prompts/%s/", locale)

// 6. 觸發 UI 重新渲染
eventBus.Publish(LanguageChangedEvent{NewLocale: locale})

// 7. 儲存設定
config.Language = locale
config.Save()
```

**系統語言偵測 (首次啟動):**
```go
import "golang.org/x/text/language"

func detectSystemLanguage() string {
    tags := language.NewMatcher([]language.Tag{
        language.TraditionalChinese,
        language.SimplifiedChinese,
        language.English,
    })

    tag, _ := language.MatchStrings(tags, os.Getenv("LANG"))

    switch tag {
    case language.TraditionalChinese, language.SimplifiedChinese:
        return "zh-TW"
    default:
        return "en-US"
    }
}
```

**Fallback 機制:**
```go
func (t *Translator) T(key string, params ...any) string {
    // 1. 嘗試當前語言
    if val, ok := t.translations[key]; ok {
        return formatString(val, params...)
    }

    // 2. 回退到英文
    if val, ok := t.fallback[key]; ok {
        log.Warn("missing translation: %s", key)
        return formatString(val, params...)
    }

    // 3. 返回 key（開發模式）
    log.Error("translation not found: %s", key)
    return fmt.Sprintf("[%s]", key)
}
```

### Translation Keys Convention

**命名規範:**
- 使用點號分隔命名空間：`namespace.category.item`
- 使用小寫蛇形命名：`new_game` 而非 `NewGame`
- 動態參數使用大括號：`{hp}`, `{name}`, `{count}`

**範例:**
```
menu.new_game
menu.continue_game
game.hp_label
game.san_label
commands.status_desc
errors.api_failed
messages.game_saved
```

### References

- [Source: docs/epics.md#Epic-7]
- [Related: Story 1.3 - 主選單介面]
- [Related: Epic 2 - 核心遊戲循環 (LLM prompts)]
- [NFR03: 可用性需求 - 多語言支援]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5

### Completion Notes List

- Story created by create-story workflow in YOLO mode
