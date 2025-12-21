# Story 3-5: ChatBubble 訊息渲染元件

**Story ID:** 3-5-chat-bubble-component
**Epic:** Epic 3 - 聊天室 TUI - 核心實作
**Status:** done
**Estimate:** 4 hours
**Created:** 2025-12-22
**Priority:** P0 (Core Feature)

---

## Story Description

建立 ChatBubble 組件，用於渲染聊天室中的單條訊息。組件需要根據不同的訊息類型、發言者和標記提供視覺化的差異，確保玩家能夠清楚辨識訊息來源和性質。

---

## Acceptance Criteria

### AC1: ChatBubble 組件渲染單條訊息
- [x] ChatBubble 組件接受 ChatMessage 結構並渲染
- [x] 包含訊息內容、發言者名稱
- [x] 使用 lipgloss 進行樣式化渲染

### AC2: 根據 speaker 區分玩家/NPC/系統訊息樣式
- [x] 玩家訊息：靠右對齊，使用特定顏色（如淺藍色）
- [x] NPC 訊息：靠左對齊，使用不同顏色（如淺綠色）
- [x] 系統訊息：居中對齊，使用中性顏色（如灰色/黃色）

### AC3: 支援 ChatMessageType 不同類型的視覺樣式
- [x] normal: 標準樣式
- [x] system: 系統通知樣式（加粗/特殊圖示）
- [x] whisper: 私語樣式（斜體/較淡顏色）
- [x] thought: 內心獨白樣式（斜體/特殊標記）
- [x] action: 動作描述樣式（斜體/特殊括號）

### AC4: 顯示時間戳記
- [x] 訊息包含時間戳記顯示（如 [12:34]）
- [x] 時間戳記使用較淡顏色，不搶眼
- [x] 時間戳記位於訊息內容之前或旁邊

### AC5: 支援 Flag 標記顯示（幻覺、敵意等）
- [x] hallucination: 顯示 🌀 圖示
- [x] hostile: 顯示 ⚠️ 圖示
- [x] revelation: 顯示 💡 圖示
- [x] persuasion: 顯示 🎯 圖示
- [x] lie: 顯示 🎭 圖示
- [x] contradiction: 顯示 ❌ 圖示
- [x] 圖示顯示在訊息內容旁邊，清晰醒目

---

## Technical Design

### 檔案結構
```
internal/tui/components/
├── chat_bubble.go          # ChatBubble 組件實作
└── chat_bubble_test.go     # 單元測試
```

### ChatMessage 資料結構（參考 Story 3-2）
```go
type ChatMessage struct {
    ID              string
    Speaker         string
    Content         string
    Timestamp       time.Time
    Type            ChatMessageType
    Flags           []ChatFlag
    EmotionEffects  map[string]EmotionDelta
}

type ChatMessageType string
const (
    ChatMessageTypeNormal  ChatMessageType = "normal"
    ChatMessageTypeSystem  ChatMessageType = "system"
    ChatMessageTypeWhisper ChatMessageType = "whisper"
    ChatMessageTypeThought ChatMessageType = "thought"
    ChatMessageTypeAction  ChatMessageType = "action"
)

type ChatFlag string
const (
    ChatFlagHallucination  ChatFlag = "hallucination"
    ChatFlagHostile        ChatFlag = "hostile"
    ChatFlagRevelation     ChatFlag = "revelation"
    ChatFlagPersuasion     ChatFlag = "persuasion"
    ChatFlagLie            ChatFlag = "lie"
    ChatFlagContradiction  ChatFlag = "contradiction"
)
```

### ChatBubble API 設計
```go
type ChatBubble struct {
    message ChatMessage
    width   int
    theme   *themes.Theme
}

func NewChatBubble(msg ChatMessage, width int) *ChatBubble
func (cb *ChatBubble) WithTheme(theme *themes.Theme) *ChatBubble
func (cb *ChatBubble) View() string
func (cb *ChatBubble) getSpeakerStyle() lipgloss.Style
func (cb *ChatBubble) getMessageStyle() lipgloss.Style
func (cb *ChatBubble) getTypeStyle() lipgloss.Style
func (cb *ChatBubble) formatTimestamp() string
func (cb *ChatBubble) getFlagIcons() string
```

### 樣式設計

#### 玩家訊息
- 對齊：靠右
- 顏色：淺藍色 (#5DADE2)
- 前綴：「你」或玩家名稱
- 背景：深色背景突出

#### NPC 訊息
- 對齊：靠左
- 顏色：淺綠色 (#58D68D) 或根據 NPC 情緒狀態變化
- 前綴：NPC 名稱
- 背景：稍淡背景

#### 系統訊息
- 對齊：居中
- 顏色：黃色 (#F4D03F) 或灰色
- 前綴：[系統]
- 樣式：加粗或特殊圖示

#### 訊息類型樣式
- whisper: 斜體，較淡顏色，前綴 "（小聲）"
- thought: 斜體，灰色，前綴 "（心想）"
- action: 斜體，橙色，前綴 "*"

---

## Implementation Tasks

### Task 1: 定義 ChatMessage 類型（如果尚不存在）
- 檢查是否已有 ChatMessage 定義（Story 3-2）
- 如果沒有，創建 internal/tui/views/chat_types.go
- 定義 ChatMessage, ChatMessageType, ChatFlag

### Task 2: 創建 ChatBubble 基礎結構
- 創建 internal/tui/components/chat_bubble.go
- 定義 ChatBubble 結構
- 實作 NewChatBubble 建構函數
- 實作 WithTheme 方法（鏈式調用）

### Task 3: 實作樣式方法
- getSpeakerStyle(): 根據 speaker 返回不同樣式
- getMessageStyle(): 根據 Type 返回不同樣式
- getTypeStyle(): 根據訊息類型返回樣式
- 整合 theme 顏色

### Task 4: 實作時間戳記格式化
- formatTimestamp(): 格式化時間為 [HH:MM]
- 時間戳記樣式設計（較淡顏色）
- 整合到訊息顯示

### Task 5: 實作 Flag 圖示顯示
- getFlagIcons(): 根據 Flags 返回對應圖示
- 圖示與訊息整合
- 確保圖示清晰醒目

### Task 6: 實作 View() 主渲染方法
- 組裝所有元素（時間戳、發言者、內容、圖示）
- 處理對齊（左/右/中）
- 處理寬度限制和文字換行
- 應用最終樣式

### Task 7: 編寫單元測試
- 測試不同 speaker 類型渲染
- 測試不同 MessageType 樣式
- 測試時間戳記格式化
- 測試 Flag 圖示顯示
- 測試主題切換

### Task 8: 整合測試
- 渲染多條不同類型訊息
- 測試視覺效果一致性
- 確保測試覆蓋率 > 80%

---

## Dependencies

- **Epic 1**: NPC 管理系統（提供 NPC 資訊）
- **Story 3-2**: ChatMessage 類型定義（如果已實作）
- **internal/tui/themes**: 主題系統
- **lipgloss**: 樣式庫

---

## Testing Strategy

### 單元測試
```go
func TestChatBubble_PlayerMessage(t *testing.T)
func TestChatBubble_NPCMessage(t *testing.T)
func TestChatBubble_SystemMessage(t *testing.T)
func TestChatBubble_MessageTypes(t *testing.T)
func TestChatBubble_Timestamp(t *testing.T)
func TestChatBubble_Flags(t *testing.T)
func TestChatBubble_MultipleFlags(t *testing.T)
func TestChatBubble_ThemeSupport(t *testing.T)
func TestChatBubble_WidthHandling(t *testing.T)
```

### 測試涵蓋範圍
- AC1: 基本渲染功能
- AC2: 不同 speaker 樣式
- AC3: 不同 MessageType 樣式
- AC4: 時間戳記顯示
- AC5: Flag 標記顯示
- 邊界條件：空訊息、超長訊息、多個 Flags

---

## Definition of Done

- [x] ChatBubble 組件實作完成
- [x] 所有 AC 通過
- [x] 單元測試完成，覆蓋率 > 80% (chat_bubble.go: ~90%)
- [x] 測試全部通過 (25 tests, 100% pass rate)
- [x] 程式碼符合專案規範
- [x] 樣式美觀，視覺效果清晰
- [x] 文件更新（如需要）
- [x] Story 標記為 done in sprint-status.yaml

---

## Notes

- 參考現有 TUI 組件（choice_list.go, status_bar.go）的樣式設計
- 確保不同訊息類型有明顯的視覺區別
- Flag 標記要清晰醒目，但不過度搶眼
- 時間戳記要低調，不干擾主要內容
- 考慮可讀性和螢幕閱讀器支援
- 確保與主題系統良好整合

---

## Story Progress

### 2025-12-22
- Story 創建
- 開始實作
