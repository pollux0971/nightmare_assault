---
stepsCompleted: [1, 2, 3]
lastStep: 3
inputDocuments:
  - /PRD.md
  - /ARCHITECTURE.md
  - /docs/ux-design-specification.md
workflowType: 'epics-stories'
project_name: 'nightmare-assault'
user_name: 'Pollux'
date: '2025-12-11'
---

# Nightmare Assault - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for Nightmare Assault (惡夢驚襲), decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

#### P0 - 核心功能 (Must Have)

| ID | 功能 | 描述 |
|----|------|------|
| FR001 | 故事生成 | 根據用戶輸入的主題/大綱生成完整故事架構 |
| FR002 | 潛規則系統 | 生成並追蹤隱藏規則，處理觸發邏輯 |
| FR003 | 對話互動 | 自由輸入或選擇選項與遊戲互動 |
| FR004 | 雙數值系統 | 生命值與理智值的追蹤與效果 |
| FR005 | NPC 系統 | 隊友生成、對話、死亡處理 |
| FR006 | 存檔/讀檔 | 3 個存檔槽，JSON 格式 |
| FR007 | 多 API 支援 | OpenAI、Anthropic、Gemini、Grok、Gateway |
| FR008 | 斜線指令 | /status、/inventory、/save 等快捷指令 |
| FR009 | 死亡覆盤 | 顯示死因、觸發的規則、錯過的線索 |
| FR010 | 跨平台編譯 | Windows、macOS、Linux 單一執行檔 |

#### P1 - 重要功能 (Should Have)

| ID | 功能 | 描述 |
|----|------|------|
| FR011 | 夢境系統 | 開場夢境、穿插夢境、夢境回顧 |
| FR012 | 章節壓縮 | Token 管理，舊章節自動摘要 |
| FR013 | BGM 系統 | 背景音樂播放、場景切換 |
| FR014 | 音效系統 | 提示音、警告音、死亡音效 |
| FR015 | 理智幻覺 | 低理智時的文字扭曲、幻覺選項 |
| FR016 | 主題切換 | 多種終端顏色主題 |
| FR017 | 打字機效果 | Stream 輸出，可開關 |
| FR018 | 自動更新 | 檢查更新、一鍵更新 |

#### P2 - 增強功能 (Nice to Have)

| ID | 功能 | 描述 |
|----|------|------|
| FR019 | 自訂 BGM | 用戶可放入自己的音樂 |
| FR020 | 多語言切換 | 遊戲中切換繁中/英文 |
| FR021 | 詳細日誌 | /log 查看歷史對話 |
| FR022 | 提示系統 | /hint 花費理智獲得提示 |

#### 斜線指令 (隸屬 FR008)

| 指令 | 功能 |
|------|------|
| `/status` | 查看 HP/SAN/狀態 |
| `/inventory` | 查看背包物品 |
| `/clues` | 查看已發現線索 |
| `/dreams` | 回顧夢境片段 |
| `/team` | 查看隊友狀態 |
| `/rules` | 查看已知規則 |
| `/save [1-3]` | 存檔 |
| `/load [1-3]` | 讀檔 |
| `/log [n]` | 查看對話紀錄 |
| `/hint` | 花費 SAN 獲得提示 |
| `/theme` | 切換主題 |
| `/api` | 切換 API 供應商 |
| `/bgm` | BGM 控制 |
| `/speed` | 打字機效果開關 |
| `/lang` | 切換語言 |
| `/help` | 顯示幫助 |
| `/quit` | 退出遊戲 |

### Non-Functional Requirements

#### NFR01 - 性能需求

| 指標 | 目標 |
|------|------|
| 啟動時間 | < 2 秒 |
| 選項響應時間 (本地) | < 100ms |
| 轉場文字顯示 | < 200ms |
| 隊友插話生成 (Fast Model) | < 500ms |
| 故事生成 (Smart Model) | < 5 秒 (串流開始) |
| 存檔/讀檔 | < 500ms |
| 記憶體使用 | < 200MB (不含音訊) |

#### NFR02 - 安全需求

| 項目 | 要求 |
|------|------|
| API Key 儲存 | 本地加密儲存，不上傳 |
| 存檔資料 | 僅本地儲存 |
| 網路請求 | 僅與 LLM API 通訊 |
| 遙測數據 | 不收集任何數據 |
| 更新驗證 | 使用 checksum 驗證 |

#### NFR03 - 可用性需求

| 項目 | 要求 |
|------|------|
| 離線遊玩 | 不支援（需要 LLM API） |
| 網路中斷 | 自動重試，提示用戶 |
| 終端相容性 | 支援 256 色終端 |
| 最小終端大小 | 80x24 |
| 鍵盤操作 | 完全支援純鍵盤操作 |

#### NFR04 - 可維護性需求

| 項目 | 要求 |
|------|------|
| 日誌記錄 | 可選的 debug 日誌 |
| 錯誤報告 | 友善的錯誤訊息 |
| 配置熱重載 | 支援修改配置後不重啟 |
| Skill 自訂 | 用戶可修改 Skill 定義檔 |

### Additional Requirements

#### 來自 Architecture

- **技術棧**: Go 1.21+, BubbleTea TUI, LipGloss 樣式, oto v3 音訊
- **雙模型架構**: Smart Model (敘事) + Fast Model (解析/延遲掩蓋)
- **EventBus**: P0-P3 優先級事件系統，64 緩衝，100ms 節流
- **14 TUI 元件**: Phase 0-3 分層實作
- **響應式佈局**: 4 級寬度適應 (<80/80-99/100-119/≥120)
- **三種顯示模式**: Unicode (心理驚悚) / ASCII (80s 復古) / Accessible (無障礙)
- **敘事化錯誤處理**: 技術錯誤包裝成世界觀訊息

#### 來自 UX Design

- **SAN 緩衝區**: 6 階段 SAN 效果（控制權即理智）
- **三層延遲防護**: 0-300ms 無/300ms-2s spinner/2s+ 隊友插話
- **死亡覆盤系統**: 難度分層回溯 (簡單=無限/困難=單次/地獄=無)
- **幻覺選項**: 事後揭露，不事先標記
- **隊友死亡**: 可預防設計 (伏筆→預警→後果分歧)
- **HorrorStyle 五維度**: TextCorruption/TypingBehavior/ColorShift/UIStability/OptionReliability
- **6 用戶旅程**: 首次啟動/核心遊戲/規則發現/死亡覆盤/SAN 崩潰/夢境體驗
- **7 類 UX 模式**: 指令層級/遊戲回饋/輸入模式/等待轉場/資訊密度/錯誤處理/覆蓋層

#### 遊戲機制需求

- **難度系統**: 簡單 (≤6規則) / 困難 (不限) / 地獄 (無警告)
- **故事長度**: 短 (3-5章) / 中 (5-8章) / 長 (8-15章)
- **潛規則類型**: 場景/時間/行為/對象/狀態/對立
- **脫軌處理**: 逃避/破壞/創造/試探/合理 五種策略
- **道具系統**: 治療/理智/線索/工具/關鍵 五類道具
- **模板系統**: P01-P10 玩法 / S01-S10 場景 / F01-F10 恐懼源

### FR Coverage Map

| FR | Epic | 說明 |
|----|------|------|
| FR001 | Epic 2 | 故事生成 - 核心遊戲循環 |
| FR002 | Epic 3 | 潛規則系統 - 規則與死亡 |
| FR003 | Epic 2 | 對話互動 - 核心遊戲循環 |
| FR004 | Epic 2 | 雙數值系統 - 核心遊戲循環 |
| FR005 | Epic 4 | NPC 系統 - 隊友系統 |
| FR006 | Epic 5 | 存檔/讀檔 - 存檔系統 |
| FR007 | Epic 1 | 多 API 支援 - 基礎設施 |
| FR008 | Epic 2,5,6 | 斜線指令 - 分散於各 Epic |
| FR009 | Epic 3 | 死亡覆盤 - 規則與死亡 |
| FR010 | Epic 1 | 跨平台編譯 - 基礎設施 |
| FR011 | Epic 6 | 夢境系統 - 沉浸體驗 |
| FR012 | Epic 5 | 章節壓縮 - 存檔系統 |
| FR013 | Epic 6 | BGM 系統 - 沉浸體驗 |
| FR014 | Epic 6 | 音效系統 - 沉浸體驗 |
| FR015 | Epic 6 | 理智幻覺 - 沉浸體驗 |
| FR016 | Epic 1 | 主題切換 - 基礎設施 |
| FR017 | Epic 6 | 打字機效果 - 沉浸體驗 |
| FR018 | Epic 7 | 自動更新 - 增強功能 |
| FR019 | Epic 6 | 自訂 BGM - 沉浸體驗 |
| FR020 | Epic 7 | 多語言切換 - 增強功能 |
| FR021 | Epic 5 | 詳細日誌 - 存檔系統 |
| FR022 | Epic 3 | 提示系統 - 規則與死亡 |

---

## Epic List

### Epic 1: 遊戲基礎與啟動體驗
> 用戶可以安裝、啟動遊戲、配置 API，並看到主選單

**FRs:** FR007, FR010, FR016
**依賴:** 無（基礎層）
**可平行:** 否（其他 Epic 依賴此 Epic）

**用戶成果:**
- 下載單一執行檔並啟動
- 設定 API Key (OpenAI/Anthropic/Gemini/Grok)
- 看到主選單、選擇顏色主題
- 完成首次啟動流程

---

### Epic 2: 核心遊戲循環
> 用戶可以開始一個故事、與遊戲互動、看到 HP/SAN 變化

**FRs:** FR001, FR003, FR004, FR008 (部分)
**依賴:** Epic 1
**可平行:** 與 Epic 5 平行

**用戶成果:**
- 輸入故事主題/選擇難度/長度
- 看到 LLM 生成的開場敘事
- 透過選項或自由輸入與故事互動
- 看到 HP/SAN 數值變化
- 基本指令操作 (/status, /help, /quit)

---

### Epic 3: 潛規則與死亡系統
> 用戶可以發現規則、違反規則、死亡、並從覆盤中學習

**FRs:** FR002, FR009, FR022
**依賴:** Epic 2
**可平行:** 與 Epic 4, Epic 6 平行

**用戶成果:**
- 從線索中推理規則
- 觸發規則警告或即死
- 死亡後看到覆盤畫面
- 了解錯過的線索與規則
- 使用 /hint 獲得提示

---

### Epic 4: NPC 與隊友系統
> 用戶可以與隊友互動、看到隊友死亡、感受情感連結

**FRs:** FR005
**依賴:** Epic 2
**可平行:** 與 Epic 3, Epic 6 平行

**用戶成果:**
- 遇見 LLM 生成的隊友
- 與隊友對話、獲得線索
- 看到隊友可預防的死亡
- 體驗隊友死亡的情感衝擊

---

### Epic 5: 存檔與遊戲進度
> 用戶可以存檔、讀檔、管理多個遊戲進度

**FRs:** FR006, FR012, FR021
**依賴:** Epic 1
**可平行:** 與 Epic 2 平行

**用戶成果:**
- 使用 /save 存檔到 3 個槽位
- 使用 /load 讀取進度
- 長遊戲時章節自動壓縮
- 查看歷史對話紀錄 (/log)

---

### Epic 6: 沉浸式恐怖體驗
> 用戶體驗 SAN 崩潰效果、幻覺選項、音效、夢境

**FRs:** FR011, FR013, FR014, FR015, FR017, FR019
**依賴:** Epic 2
**可平行:** 與 Epic 3, Epic 4 平行

**用戶成果:**
- 低 SAN 時看到文字扭曲
- 遇到幻覺選項 (事後揭露)
- 聽到 BGM 與音效
- 體驗夢境序列
- 打字機效果增強沉浸感

---

### Epic 7: 增強功能與維護
> 用戶可以自動更新、切換語言

**FRs:** FR018, FR020
**依賴:** Epic 1-6
**可平行:** 否（收尾階段）

**用戶成果:**
- 一鍵檢查並更新遊戲
- 切換繁中/英文

---

## 平行開發策略

```
Phase 1 (基礎):
    Epic 1 ────────────────────────────────►

Phase 2 (可平行):
    Epic 2 (核心循環) ─────────────────────►
    Epic 5 (存檔系統) ─────────────────────►

Phase 3 (可平行):
    Epic 3 (規則/死亡) ────────────────────►
    Epic 4 (NPC/隊友) ─────────────────────►
    Epic 6 (恐怖效果) ─────────────────────►

Phase 4 (收尾):
    Epic 7 (增強功能) ─────────────────────►
```

**建議 Agent 分配:**
| Phase | Agent 數量 | 分工 |
|-------|-----------|------|
| 1 | 1 | Epic 1 |
| 2 | 2 | Agent A: Epic 2, Agent B: Epic 5 |
| 3 | 3 | Agent A: Epic 3, Agent B: Epic 4, Agent C: Epic 6 |
| 4 | 1 | Epic 7 |

---

# Epic Details & Stories

## Epic 1: 遊戲基礎與啟動體驗

> 用戶可以安裝、啟動遊戲、配置 API，並看到主選單

### Story 1.1: CLI 應用程式框架

As a **開發者**,
I want **建立 Go CLI 應用程式的基礎框架**,
So that **後續功能可以在此基礎上開發**.

**Acceptance Criteria:**

**Given** 專案目錄結構已建立
**When** 執行 `go build` 命令
**Then** 產生可執行的 `nightmare` 二進位檔
**And** 執行檔啟動時間 < 2 秒 (NFR01)
**And** 記憶體使用 < 50MB（基礎狀態）

**Given** 使用者在終端機執行 `nightmare`
**When** 程式啟動
**Then** 顯示 BubbleTea TUI 框架的基礎畫面
**And** 支援 80x24 最小終端大小 (NFR03)

**技術備註:**
- 建立 `cmd/nightmare/main.go` 入口點
- 建立 `internal/` 目錄結構
- 整合 BubbleTea + LipGloss
- 建立基礎 Model/View/Update 架構

---

### Story 1.2: 多 API 供應商支援

As a **玩家**,
I want **設定我偏好的 LLM API 供應商**,
So that **我可以使用自己的 API Key 來遊玩**.

**Acceptance Criteria:**

**Given** 玩家首次啟動遊戲且無 API 設定
**When** 進入 API 設定流程
**Then** 顯示支援的供應商清單：OpenAI、Anthropic、Gemini、Grok、OpenRouter
**And** 可選擇一個供應商

**Given** 玩家選擇了供應商
**When** 輸入 API Key
**Then** API Key 加密儲存於本地 (NFR02)
**And** 不會上傳至任何伺服器

**Given** 玩家已設定 API Key
**When** 執行連線測試
**Then** 顯示連線成功或失敗訊息
**And** 失敗時顯示友善的錯誤說明

**Given** 玩家想更換 API 供應商
**When** 使用 `/api` 指令
**Then** 顯示互動式供應商選擇選單
**And** 可切換至不同供應商

**技術備註:**
- 建立 `internal/api/` 模組
- 實作 Provider 介面 (OpenAI/Anthropic/Gemini/Grok/OpenRouter)
- API Key 使用 AES 加密儲存於 `~/.nightmare/config.json`

---

### Story 1.3: 主選單介面

As a **玩家**,
I want **看到清晰的主選單**,
So that **我可以開始新遊戲、讀取存檔、或調整設定**.

**Acceptance Criteria:**

**Given** 玩家已完成 API 設定
**When** 進入主選單
**Then** 顯示以下選項：
  - 新遊戲
  - 繼續遊戲（如有存檔）
  - 設定
  - 離開

**Given** 玩家在主選單
**When** 使用方向鍵或數字鍵選擇選項
**Then** 選項高亮顯示
**And** 按 Enter 執行選擇

**Given** 玩家選擇「設定」
**When** 進入設定選單
**Then** 顯示：主題切換、API 設定、音效開關

**技術備註:**
- 使用 BubbleTea list component
- 實作 MenuModel
- 狀態機管理選單導航

---

### Story 1.4: 顏色主題系統

As a **玩家**,
I want **切換不同的終端顏色主題**,
So that **我可以選擇適合我視覺偏好的風格**.

**Acceptance Criteria:**

**Given** 玩家在設定選單
**When** 選擇「主題切換」
**Then** 顯示 5 種主題選項：
  - Midnight（預設）
  - Blood Moon
  - Terminal Green
  - Silent Hill Fog
  - High Contrast

**Given** 玩家選擇一個主題
**When** 確認選擇
**Then** 立即套用新主題
**And** 主題偏好儲存至設定檔

**Given** 玩家下次啟動遊戲
**When** 載入設定
**Then** 自動套用上次選擇的主題

**技術備註:**
- 使用 LipGloss 定義 5 套色彩方案
- 建立 `internal/tui/themes/` 模組
- 主題設定持久化至 config.json

---

### Story 1.5: 跨平台編譯與發布

As a **玩家**,
I want **下載適合我作業系統的單一執行檔**,
So that **我可以直接執行遊戲無需安裝依賴**.

**Acceptance Criteria:**

**Given** 開發者執行編譯腳本
**When** 編譯完成
**Then** 產生以下執行檔：
  - `nightmare-windows-amd64.exe`
  - `nightmare-darwin-amd64`
  - `nightmare-darwin-arm64`
  - `nightmare-linux-amd64`

**Given** 使用者下載對應平台的執行檔
**When** 直接執行（無需安裝）
**Then** 程式正常啟動
**And** 無外部依賴（靜態編譯）

**Given** 執行檔包含版本資訊
**When** 執行 `nightmare --version`
**Then** 顯示版本號、編譯日期、Git commit

**技術備註:**
- 建立 `Makefile` 或 `build.sh`
- 使用 `go build -ldflags` 嵌入版本資訊
- CGO_ENABLED=0 靜態編譯

---

## Epic 2: 核心遊戲循環

> 用戶可以開始一個故事、與遊戲互動、看到 HP/SAN 變化

### Story 2.1: 新遊戲設定流程

As a **玩家**,
I want **設定新遊戲的參數**,
So that **我可以自訂故事主題、難度和長度**.

**Acceptance Criteria:**

**Given** 玩家在主選單選擇「新遊戲」
**When** 進入新遊戲設定
**Then** 依序詢問：
  1. 故事主題（自由輸入或選擇提示）
  2. 難度（簡單/困難/地獄）
  3. 故事長度（短/中/長）
  4. 18+ 模式（是/否）

**Given** 玩家輸入故事主題
**When** 輸入超過 300 tokens
**Then** 顯示警告並截斷
**And** 顯示 6 個預設提示供參考

**Given** 玩家選擇難度
**When** 顯示難度選項
**Then** 每個難度顯示簡短說明：
  - 簡單：≤6 規則，2 次警告
  - 困難：不限規則，1 次警告
  - 地獄：不限規則，無警告

**Given** 玩家完成所有設定
**When** 確認開始
**Then** 進入故事生成階段

**技術備註:**
- 建立 `internal/game/setup.go`
- 設定資料結構 `GameConfig`
- 驗證輸入並存入遊戲狀態

---

### Story 2.2: 故事生成引擎

As a **玩家**,
I want **看到 AI 生成的開場故事**,
So that **我可以沉浸在恐怖氛圍中開始冒險**.

**Acceptance Criteria:**

**Given** 玩家完成新遊戲設定
**When** 系統開始生成故事
**Then** 顯示「正在構築惡夢...」等待畫面
**And** 串流開始時間 < 5 秒 (NFR01)

**Given** Smart Model 開始回應
**When** 串流輸出開場敘事
**Then** 以打字機效果逐字顯示
**And** 敘事包含：場景描述、氛圍營造、玩家角色定位

**Given** 開場敘事完成
**When** 顯示第一個互動選項
**Then** 提供 2-4 個選項
**And** 顯示「或輸入你想做的事...」自由輸入提示

**Given** 故事生成使用 Game Bible
**When** 生成內容
**Then** 遵循 Prompt 系統的結構化輸出格式
**And** 包含隱藏的規則種子（玩家不可見）

**技術備註:**
- 建立 `internal/engine/story.go`
- 實作 Smart Model 呼叫
- 建立 Game Bible prompt 模板
- 串流處理與 TUI 整合

---

### Story 2.3: 玩家輸入處理

As a **玩家**,
I want **透過選項或自由輸入與遊戲互動**,
So that **我的選擇能影響故事發展**.

**Acceptance Criteria:**

**Given** 遊戲顯示選項
**When** 玩家按數字鍵 (1-9)
**Then** 選擇對應選項
**And** 響應時間 < 100ms (NFR01)

**Given** 遊戲允許自由輸入
**When** 玩家輸入自定義行動
**Then** 系統接受任意文字輸入
**And** 傳送至 LLM 處理

**Given** 玩家輸入斜線指令
**When** 輸入以 `/` 開頭
**Then** 解析為系統指令而非遊戲行動
**And** 執行對應指令（/status, /help, /quit）

**Given** 玩家按 Enter 但無輸入
**When** 有預設選項時
**Then** 選擇第一個選項
**When** 無預設選項時
**Then** 忽略並等待輸入

**技術備註:**
- 建立 `internal/tui/input/handler.go`
- 實作 InputMode 狀態機
- 斜線指令解析器

---

### Story 2.4: HP/SAN 數值系統

As a **玩家**,
I want **看到我的 HP 和 SAN 數值變化**,
So that **我能了解角色的生存狀態和心理狀態**.

**Acceptance Criteria:**

**Given** 遊戲開始
**When** 初始化角色狀態
**Then** HP = 100, SAN = 100

**Given** 發生影響 HP 的事件
**When** LLM 回應包含 HP 變化
**Then** 解析變化量並更新 HP
**And** 狀態列即時更新顯示
**And** HP 變化伴隨視覺反饋（紅色閃爍）

**Given** 發生影響 SAN 的事件
**When** LLM 回應包含 SAN 變化
**Then** 解析變化量並更新 SAN
**And** 根據 SAN 範圍顯示對應狀態：
  - 80-100：清醒
  - 50-79：焦慮
  - 20-49：恐慌
  - 1-19：崩潰

**Given** HP 降至 0
**When** 更新狀態
**Then** 觸發死亡流程

**Given** SAN 降至 0
**When** 更新狀態
**Then** 觸發瘋狂結局

**技術備註:**
- 建立 `internal/game/stats.go`
- 實作 EventBus 監聽 HP/SAN 變化
- Fast Model 解析 LLM 回應中的數值變化

---

### Story 2.5: 遊戲主畫面佈局

As a **玩家**,
I want **清晰的遊戲畫面佈局**,
So that **我能同時看到故事、狀態和選項**.

**Acceptance Criteria:**

**Given** 玩家進入遊戲
**When** 顯示遊戲主畫面
**Then** 畫面分為四區：
  - 頂部：狀態列（HP/SAN/位置）2 行
  - 中間：敘事區（可捲動）height-9 行
  - 底部：選項區 5 行
  - 最底：快捷鍵提示 2 行

**Given** 終端機寬度 < 80
**When** 檢測到窄螢幕
**Then** 切換為精簡模式（僅顯示 HP/SAN 數字）

**Given** 終端機寬度 ≥ 120
**When** 檢測到寬螢幕
**Then** 啟用寬敞模式（側邊欄顯示物品）

**Given** 敘事內容超過顯示區域
**When** 玩家使用上下鍵
**Then** 可捲動查看歷史敘事

**技術備註:**
- 建立 `internal/tui/views/game.go`
- 實作響應式佈局 (LayoutMode)
- 整合 viewport 元件處理捲動

---

### Story 2.6: 基礎斜線指令

As a **玩家**,
I want **使用基礎斜線指令**,
So that **我可以查看狀態、取得幫助或離開遊戲**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中
**When** 輸入 `/status`
**Then** 顯示：HP、SAN、當前位置、遊戲時間

**Given** 玩家在遊戲中
**When** 輸入 `/help`
**Then** 顯示所有可用指令清單與說明

**Given** 玩家在遊戲中
**When** 輸入 `/quit`
**Then** 詢問是否存檔
**And** 確認後返回主選單

**Given** 玩家輸入未知指令
**When** 指令不存在
**Then** 顯示「未知指令，輸入 /help 查看可用指令」

**技術備註:**
- 建立 `internal/game/commands/` 模組
- 實作 Command 介面
- 指令註冊與分發機制

---

## Epic 3: 潛規則與死亡系統

> 用戶可以發現規則、違反規則、死亡、並從覆盤中學習

### Story 3.1: 潛規則生成

As a **系統**,
I want **在故事開始時生成潛規則**,
So that **遊戲有明確的生存邏輯**.

**Acceptance Criteria:**

**Given** 新遊戲開始
**When** 生成故事架構
**Then** 同時生成潛規則集合
**And** 規則數量根據難度：簡單 ≤6、困難/地獄不限

**Given** 生成潛規則
**When** 決定規則類型
**Then** 從以下類型中選擇：
  - 場景規則（特定地點生效）
  - 時間規則（特定時段生效）
  - 行為規則（任何時候生效）
  - 對象規則（針對特定 NPC/物品）
  - 狀態規則（依賴玩家狀態）

**Given** 潛規則已生成
**When** 儲存規則
**Then** 規則存於遊戲狀態中（玩家不可見）
**And** 每條規則包含：觸發條件、後果、相關線索

**技術備註:**
- 建立 `internal/engine/rules/generator.go`
- 規則資料結構 `Rule{Type, Trigger, Consequence, Clues}`
- 整合至 Game Bible prompt

---

### Story 3.2: 規則觸發檢測

As a **系統**,
I want **檢測玩家行為是否觸發規則**,
So that **規則違反有相應後果**.

**Acceptance Criteria:**

**Given** 玩家執行任何行動
**When** 系統處理行動
**Then** 檢查所有適用的潛規則
**And** 按順序：場景→時間→行為→狀態

**Given** 行動觸發規則
**When** 規則為輕微違規
**Then** 根據難度發出警告（簡單 2 次/困難 1 次/地獄 0 次）
**And** 扣除 HP 或 SAN

**Given** 行動觸發致命規則
**When** 無警告次數剩餘
**Then** 觸發即死
**And** 進入死亡流程

**Given** 對立規則情境
**When** 兩條規則同時適用
**Then** 玩家必須判斷優先級
**And** 錯誤判斷導致後果

**技術備註:**
- 建立 `internal/engine/rules/checker.go`
- 實作規則優先級評估
- 與 LLM 回應整合

---

### Story 3.3: 死亡流程

As a **玩家**,
I want **在死亡時看到戲劇化的結局**,
So that **死亡有儀式感且印象深刻**.

**Acceptance Criteria:**

**Given** 玩家 HP 降至 0 或觸發即死
**When** 進入死亡流程
**Then** 顯示死亡轉場動畫（紅色漸變）
**And** 播放心跳停止音效
**And** 動畫時長 2 秒

**Given** 死亡轉場完成
**When** 顯示死亡畫面
**Then** 全螢幕紅色背景
**And** 顯示死亡敘事（LLM 生成）
**And** 顯示選項：查看覆盤 / 返回主選單

**Given** SAN 降至 0
**When** 進入瘋狂結局
**Then** 顯示特殊的瘋狂敘事
**And** 畫面效果為 Glitch + 扭曲

**技術備註:**
- 建立 `internal/tui/views/death.go`
- 實作 TransitionOverlay 元件
- 死亡敘事 prompt 模板

---

### Story 3.4: 死亡覆盤系統

As a **玩家**,
I want **在死亡後看到詳細覆盤**,
So that **我能學習並在下次避免同樣錯誤**.

**Acceptance Criteria:**

**Given** 玩家在死亡畫面選擇「查看覆盤」
**When** 進入覆盤畫面
**Then** 顯示以下資訊：
  - 死因摘要
  - 觸發的規則（現在揭露）
  - 錯過的線索清單
  - 關鍵決策點回顧

**Given** 覆盤顯示錯過的線索
**When** 玩家查看線索
**Then** 高亮顯示原本出現的位置
**And** 解釋線索與規則的關聯

**Given** 覆盤顯示幻覺選項
**When** 玩家曾選擇幻覺選項
**Then** 標記該選項為「幻覺」
**And** 解釋當時的 SAN 值導致幻覺

**Given** 難度為「簡單」
**When** 完成覆盤
**Then** 提供「回溯重試」選項（回到最近檢查點）

**技術備註:**
- 建立 `internal/tui/views/debrief.go`
- 建立 `internal/game/debrief.go` 收集覆盤資料
- 檢查點機制整合存檔系統

---

### Story 3.5: 提示系統

As a **玩家**,
I want **花費 SAN 獲得提示**,
So that **在卡關時有求助選項**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中
**When** 輸入 `/hint`
**Then** 確認是否花費 10 SAN 獲得提示

**Given** 玩家確認花費
**When** SAN ≥ 10
**Then** 扣除 10 SAN
**And** 顯示與當前情境相關的提示
**And** 提示為模糊暗示，非直接答案

**Given** 玩家確認花費
**When** SAN < 10
**Then** 顯示「你的理智不足以思考...」
**And** 不提供提示

**Given** 難度為「地獄」
**When** 輸入 `/hint`
**Then** 顯示「在這個噩夢中，沒有人能幫助你」
**And** 功能不可用

**技術備註:**
- 擴展 `internal/game/commands/hint.go`
- 提示生成整合 Fast Model
- 難度檢查邏輯

---

## Epic 4: NPC 與隊友系統

> 用戶可以與隊友互動、看到隊友死亡、感受情感連結

### Story 4.1: 隊友生成

As a **玩家**,
I want **在故事中遇見 AI 生成的隊友**,
So that **我有同伴一起面對恐怖**.

**Acceptance Criteria:**

**Given** 故事設定包含隊友
**When** 生成故事架構
**Then** 同時生成 1-3 個隊友角色
**And** 每個隊友有：姓名、性格、背景、特殊技能

**Given** 隊友首次出場
**When** 敘事介紹隊友
**Then** 透過行為/對話/物品展現性格（Show don't tell）
**And** 不直接列出性格描述

**Given** 隊友角色已建立
**When** 後續互動
**Then** 保持性格一致性
**And** 遵循 LLM Characterization Rules

**技術備註:**
- 建立 `internal/game/npc/generator.go`
- 隊友資料結構 `Teammate{Name, Personality, Background, Skills, Status}`
- 整合至 Game Bible 的角色設定

---

### Story 4.2: 隊友對話與互動

As a **玩家**,
I want **與隊友對話並獲得線索**,
So that **隊友是有用的資訊來源**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中有隊友
**When** 輸入 `/team`
**Then** 顯示所有隊友狀態：
  - 姓名、HP、位置
  - 攜帶物品
  - 當前情緒狀態

**Given** 玩家選擇與隊友交談
**When** 進入對話模式
**Then** 隊友回應符合其性格
**And** 可能透露線索（隱晦方式）
**And** 深度交流可恢復 5-15 SAN

**Given** 等待 LLM 回應超過 2 秒
**When** 觸發延遲掩蓋
**Then** 顯示隊友插話（Fast Model 生成）
**And** 插話內容符合當前情境

**技術備註:**
- 建立 `internal/game/npc/dialogue.go`
- 實作三層延遲防護
- Fast Model 隊友插話 prompt

---

### Story 4.3: 隊友死亡機制

As a **玩家**,
I want **有機會預防隊友死亡**,
So that **隊友死亡是我的責任而非劇本殺**.

**Acceptance Criteria:**

**Given** 隊友即將進入危險
**When** 敘事發展
**Then** 提供伏筆（至少 1 個明顯 + 1 個隱晦）
**And** 伏筆出現在隊友死亡前 1-3 回合

**Given** 伏筆已出現
**When** 危險即將發生
**Then** 提供預警（隊友行為異常、環境變化）
**And** 玩家有機會干預

**Given** 玩家未能干預
**When** 隊友死亡
**Then** 顯示死亡敘事
**And** 玩家 SAN -15 至 -25
**And** 後續分歧：可找到隊友遺物/線索

**Given** 玩家成功干預
**When** 阻止危險
**Then** 隊友存活
**And** 可能獲得額外線索作為回報

**技術備註:**
- 建立 `internal/game/npc/death.go`
- 伏筆→預警→後果分歧 狀態機
- 死亡事件 SAN 扣除邏輯

---

### Story 4.4: 隊友狀態追蹤

As a **玩家**,
I want **追蹤隊友的位置和狀態**,
So that **我知道隊友是否安全**.

**Acceptance Criteria:**

**Given** 遊戲進行中
**When** 隊友位置改變
**Then** 更新隊友狀態
**And** 狀態列顯示簡要資訊（寬敞模式）

**Given** 隊友與玩家分散
**When** 隊友獨自行動
**Then** 定期收到隊友訊息（如果通訊可用）
**And** 訊息可能包含他們發現的線索

**Given** 隊友 HP 降低
**When** HP < 30
**Then** 隊友行為受影響（移動緩慢、反應遲鈍）
**And** 敘事中反映受傷狀態

**技術備註:**
- 擴展 `internal/game/npc/` 模組
- 隊友狀態同步機制
- 分散狀態的訊息系統

---

## Epic 5: 存檔與遊戲進度

> 用戶可以存檔、讀檔、管理多個遊戲進度

### Story 5.1: 存檔資料結構

As a **系統**,
I want **定義完整的存檔資料結構**,
So that **遊戲狀態可以完整保存**.

**Acceptance Criteria:**

**Given** 需要保存遊戲狀態
**When** 設計存檔結構
**Then** JSON 格式包含：
  - 元資料：版本、存檔時間、遊玩時間
  - 玩家狀態：HP、SAN、位置、背包
  - 遊戲狀態：章節、已知規則、已發現線索
  - 隊友狀態：存活、HP、位置、物品
  - 故事上下文：最近章節摘要、當前場景

**Given** 存檔需要向前相容
**When** 讀取舊版本存檔
**Then** 自動遷移至新格式
**And** 不遺失重要資料

**Given** 存檔檔案
**When** 儲存至磁碟
**Then** 路徑為 `~/.nightmare/saves/save_{1-3}.json`
**And** 存檔大小 < 1MB

**技術備註:**
- 建立 `internal/game/save/schema.go`
- 定義 SaveData 結構
- 版本遷移機制

---

### Story 5.2: 存檔操作

As a **玩家**,
I want **隨時存檔和讀檔**,
So that **我可以保存進度並在需要時繼續**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中
**When** 輸入 `/save` 或 `/save 1`
**Then** 顯示存檔槽選擇（1-3）
**And** 顯示每個槽的狀態（空/章節/時間）

**Given** 玩家選擇存檔槽
**When** 確認存檔
**Then** 儲存當前遊戲狀態
**And** 存檔完成 < 500ms (NFR01)
**And** 顯示「存檔完成」確認訊息

**Given** 玩家在主選單或遊戲中
**When** 輸入 `/load` 或 `/load 2`
**Then** 顯示可用存檔清單
**And** 選擇後載入遊戲狀態
**And** 讀檔完成 < 500ms (NFR01)

**Given** 存檔槽已有資料
**When** 覆蓋存檔
**Then** 詢問確認「覆蓋現有存檔？」

**技術備註:**
- 建立 `internal/game/save/manager.go`
- 實作 Save/Load 函數
- 存檔槽 UI 元件

---

### Story 5.3: 章節壓縮機制

As a **系統**,
I want **自動壓縮舊章節**,
So that **長時間遊玩不會超出 Token 限制**.

**Acceptance Criteria:**

**Given** 遊戲進行超過 3 章
**When** Token 使用接近限制（>80%）
**Then** 自動觸發章節壓縮

**Given** 觸發章節壓縮
**When** 處理舊章節
**Then** 使用 Fast Model 生成摘要
**And** 保留關鍵資訊：規則線索、重要決策、NPC 狀態
**And** 刪除冗餘敘事細節

**Given** 壓縮完成
**When** 繼續遊戲
**Then** 新內容基於壓縮後的上下文生成
**And** 玩家體驗無縫（無需感知壓縮發生）

**Given** 存檔包含壓縮資料
**When** 讀取存檔
**Then** 正確還原壓縮後的上下文

**技術備註:**
- 建立 `internal/engine/compress.go`
- Token 計算與監控
- 摘要生成 prompt

---

### Story 5.4: 遊戲日誌功能

As a **玩家**,
I want **查看歷史對話紀錄**,
So that **我可以回顧錯過的線索**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中
**When** 輸入 `/log` 或 `/log 20`
**Then** 顯示最近 N 筆對話（預設 10）
**And** 包含：敘事、選項、玩家輸入

**Given** 日誌顯示中
**When** 使用上下鍵
**Then** 可捲動查看更多歷史

**Given** 日誌內容
**When** 顯示時
**Then** 時間戳記 + 對話類型標記
**And** 系統訊息與遊戲敘事區分

**Given** 遊戲結束或存檔
**When** 保存日誌
**Then** 日誌隨存檔一起保存
**And** 讀檔時可查看歷史日誌

**技術備註:**
- 建立 `internal/game/log.go`
- 環形緩衝區儲存（最多 1000 筆）
- 日誌序列化至存檔

---

## Epic 6: 沉浸式恐怖體驗

> 用戶體驗 SAN 崩潰效果、幻覺選項、音效、夢境

### Story 6.1: SAN 視覺效果系統

As a **玩家**,
I want **在低 SAN 時看到視覺干擾**,
So that **我能「感受」到角色的理智崩潰**.

**Acceptance Criteria:**

**Given** SAN 在 60-79
**When** 顯示畫面
**Then** 偶爾閃爍、輕微邊框變化

**Given** SAN 在 40-59
**When** 顯示畫面
**Then** 輸入框縮小、邊框顏色明顯變化

**Given** SAN 在 20-39
**When** 顯示畫面
**Then** 文字開始扭曲（Zalgo 效果）
**And** 游標閃爍加速
**And** 顏色偏移

**Given** SAN 在 1-19
**When** 顯示畫面
**Then** 嚴重視覺干擾
**And** 隨機刪除輸入字元
**And** UI 邊框抖動

**技術備註:**
- 建立 `internal/tui/effects/horror_style.go`
- 實作 HorrorStyle 五維度系統
- 整合 EventBus 監聽 SAN 變化

---

### Story 6.2: 幻覺選項系統

As a **玩家**,
I want **在低 SAN 時遇到幻覺選項**,
So that **我體驗到無法信任自己感官的恐怖**.

**Acceptance Criteria:**

**Given** SAN < 20
**When** 顯示選項
**Then** 可能插入 1 個幻覺選項（機率隨 SAN 降低增加）
**And** 幻覺選項外觀與真實選項無異

**Given** 玩家選擇幻覺選項
**When** 執行選擇
**Then** 敘事揭露這是幻覺
**And** 產生負面後果（SAN -5 或陷入危險）

**Given** 遊戲結束或覆盤
**When** 顯示覆盤資訊
**Then** 標記哪些選項是幻覺
**And** 解釋當時的 SAN 值

**Given** 無障礙模式
**When** 顯示選項
**Then** 仍然插入幻覺選項
**And** 覆盤時以文字描述「此為幻覺」

**技術備註:**
- 建立 `internal/tui/components/hallucination_list.go`
- 幻覺生成邏輯 `internal/game/hallucination.go`
- 覆盤資料收集

---

### Story 6.3: 打字機效果

As a **玩家**,
I want **看到敘事以打字機效果顯示**,
So that **閱讀過程更有沉浸感**.

**Acceptance Criteria:**

**Given** LLM 串流輸出敘事
**When** 顯示文字
**Then** 以打字機效果逐字顯示
**And** 速度約 30-50 字/秒（可調整）

**Given** 打字機效果進行中
**When** 玩家按 Space 或 Enter
**Then** 跳過動畫，立即顯示完整文字

**Given** 低 SAN 狀態
**When** 打字機效果
**Then** 速度不穩定（有時快有時慢）
**And** 偶爾「吞字」或重複

**Given** 玩家偏好
**When** 輸入 `/speed off`
**Then** 停用打字機效果
**And** 文字直接完整顯示

**技術備註:**
- 建立 `internal/tui/components/typewriter_view.go`
- 可中斷動畫機制
- 整合 HorrorStyle 的 TypingBehavior

---

### Story 6.4: BGM 系統

As a **玩家**,
I want **聽到符合氛圍的背景音樂**,
So that **音效增強恐怖體驗**.

**Acceptance Criteria:**

**Given** 遊戲啟動
**When** 音訊系統初始化
**Then** 檢測音訊檔案是否存在
**And** 若不存在，提示下載或靜音模式

**Given** 玩家執行 `nightmare --download-audio standard`
**When** 下載完成
**Then** 安裝標準音訊包（6 BGM + 10 SFX）
**And** 存放於 `~/.nightmare/audio/`

**Given** 遊戲進行中
**When** 場景氛圍變化
**Then** 自動切換對應 BGM：
  - 探索：ambient_探索.mp3
  - 緊張：tension_追逐.mp3
  - 安全：safe_休息.mp3

**Given** 玩家偏好
**When** 輸入 `/bgm off` 或 `/bgm volume 50`
**Then** 停用 BGM 或調整音量
**And** 設定儲存至配置檔

**技術備註:**
- 建立 `internal/audio/manager.go`
- 使用 oto v3 播放
- 音訊下載器 `cmd/nightmare/audio.go`

---

### Story 6.5: 音效系統

As a **玩家**,
I want **聽到事件音效**,
So that **關鍵時刻有聽覺回饋**.

**Acceptance Criteria:**

**Given** 遊戲事件發生
**When** 事件類型有對應音效
**Then** 播放音效：
  - 心跳：SAN < 40
  - 警告音：規則觸發
  - 死亡音：HP/SAN 歸零
  - 門聲/腳步聲：環境音效

**Given** BGM 正在播放
**When** 播放 SFX
**Then** SFX 與 BGM 混音
**And** SFX 不中斷 BGM

**Given** 無音訊檔案
**When** 嘗試播放音效
**Then** 靜默失敗（不影響遊戲）
**And** 記錄警告日誌

**技術備註:**
- 擴展 `internal/audio/manager.go`
- SFX 播放器獨立於 BGM
- 四層音訊混音

---

### Story 6.6: 夢境系統

As a **玩家**,
I want **體驗夢境序列**,
So that **夢境成為規則預告和氛圍營造的工具**.

**Acceptance Criteria:**

**Given** 新遊戲開始
**When** 生成開場
**Then** 先顯示開場夢境
**And** 夢境暗示即將遇到的規則（模糊）

**Given** 章節轉換或睡眠事件
**When** 進入夢境
**Then** 顯示夢境轉場（霧化效果）
**And** 夢境內容與主線相關但扭曲

**Given** 玩家在遊戲中
**When** 輸入 `/dreams`
**Then** 顯示已經歷的夢境片段清單
**And** 可選擇回顧特定夢境

**Given** 覆盤時
**When** 顯示夢境
**Then** 解釋夢境與規則的關聯

**技術備註:**
- 建立 `internal/game/dream.go`
- 夢境渲染元件 `internal/tui/components/dream_renderer.go`
- 夢境 prompt 模板

---

### Story 6.7: 自訂 BGM 支援

As a **玩家**,
I want **使用自己的音樂檔案**,
So that **我可以個人化恐怖體驗**.

**Acceptance Criteria:**

**Given** 玩家有自訂音樂
**When** 放置於 `~/.nightmare/audio/custom/`
**Then** 遊戲自動偵測

**Given** 自訂音樂存在
**When** 進入設定
**Then** 可指定哪首自訂音樂對應哪種場景

**Given** 自訂音樂格式
**When** 載入檔案
**Then** 支援 .mp3, .ogg, .wav 格式
**And** 不支援的格式顯示警告

**技術備註:**
- 擴展 `internal/audio/manager.go`
- 自訂音樂配置檔
- 格式檢測與轉換

---

## Epic 7: 增強功能與維護

> 用戶可以自動更新、切換語言

### Story 7.1: 自動更新檢查

As a **玩家**,
I want **遊戲自動檢查更新**,
So that **我能及時獲得新功能和修復**.

**Acceptance Criteria:**

**Given** 遊戲啟動
**When** 連線可用
**Then** 背景檢查 GitHub Releases 最新版本
**And** 若有新版本，主選單顯示提示

**Given** 有新版本可用
**When** 玩家選擇更新
**Then** 下載新版本執行檔
**And** 驗證 checksum (NFR02)
**And** 替換當前執行檔

**Given** 更新過程
**When** 發生錯誤
**Then** 顯示友善錯誤訊息
**And** 不影響當前版本運行

**Given** 玩家執行 `nightmare --update`
**When** 命令行模式
**Then** 直接執行更新流程

**技術備註:**
- 建立 `internal/update/checker.go`
- GitHub API 整合
- 跨平台執行檔替換

---

### Story 7.2: 多語言切換

As a **玩家**,
I want **在遊戲中切換語言**,
So that **我可以使用偏好的語言遊玩**.

**Acceptance Criteria:**

**Given** 玩家在設定或遊戲中
**When** 輸入 `/lang zh-TW` 或 `/lang en-US`
**Then** 切換 UI 語言
**And** 即時生效

**Given** 語言為英文
**When** 顯示 UI 元素
**Then** 選單、指令說明、系統訊息為英文
**And** 故事內容請求 LLM 以英文生成

**Given** 語言設定已儲存
**When** 下次啟動遊戲
**Then** 自動載入上次的語言設定

**Given** 遊戲進行中切換語言
**When** 切換完成
**Then** 已生成的故事保持原語言
**And** 新內容使用新語言

**技術備註:**
- 建立 `internal/i18n/` 模組
- 語言檔案 JSON 格式
- LLM prompt 語言切換

---

### Story 7.3: 完整斜線指令集

As a **玩家**,
I want **使用所有斜線指令**,
So that **我有完整的遊戲控制能力**.

**Acceptance Criteria:**

**Given** 玩家在遊戲中
**When** 使用任何已定義的斜線指令
**Then** 正確執行對應功能

**完整指令清單驗證:**
- `/status` - 顯示狀態 ✓
- `/inventory` - 背包物品
- `/clues` - 已發現線索
- `/dreams` - 夢境回顧
- `/team` - 隊友狀態
- `/rules` - 已知規則
- `/save [1-3]` - 存檔 ✓
- `/load [1-3]` - 讀檔 ✓
- `/log [n]` - 對話紀錄 ✓
- `/hint` - 提示 ✓
- `/theme` - 切換主題 ✓
- `/api` - API 設定 ✓
- `/bgm` - BGM 控制 ✓
- `/speed` - 打字機開關 ✓
- `/lang` - 語言切換 ✓
- `/help` - 幫助 ✓
- `/quit` - 離開 ✓

**技術備註:**
- 整合所有 Epic 的指令實作
- 統一指令註冊機制
- `/help` 動態生成指令清單
