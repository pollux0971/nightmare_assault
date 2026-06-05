# Nightmare Assault — 設計包

LLM 驅動的無限恐怖文字冒險遊戲完整設計文件 + 8 個 agent 的 prompt。

## 內容

```
nightmare-assault-design/
├── README.md              ← 本檔
├── 00-overview.md         專題總覽、核心理念、決策日誌、詞彙表
├── 01-architecture.md     系統架構、8 agent 編制、權限模型、目錄結構
├── 02-algorithms.md       核心演算法（雙層 bible、beat 迴圈、warden、dreaming、命運、壓縮…）
├── 03-agent-context.md    三層記憶、Blackboard schema、各 agent I/O、token 預算
├── 04-ui-ux.md            UI/UX 設計原則與內容（互動規則、節奏、文案、版面）
├── 05-epic-list.md        Epic 清單、開發階段、工時依賴、MVP 範圍、待決事項
├── 06-frontend.md         前端技術規劃（網頁+pywebview、串流渲染、CSS 主題、跨電腦一致性）
├── 07-data-contracts.md   資料契約（API contract、Pydantic schema、JSON 解析 fallback、injection 防護）
├── 08-engineering.md      工程實作（SQLite schema、並行 patch、錯誤恢復、API key、測試驗收）
├── skills/                8 個 agent 的 prompt（可熱重載）
└── build/                 AI 開發工具包（讓 AI 一塊塊蓋出系統而不失憶）
    ├── README.md          工作流程 + 給 AI 的對話範本
    ├── DESIGN-CHANGES.md  為施工而調整的大方向
    ├── CONTRACTS.md       不可變接口（每次寫代碼必貼）
    ├── BUILD-PLAN.md      19 個工單（有依賴順序、可獨立測試）
    └── CHECKLIST.md       實作易錯點（可勾選、有驗收）
    ├── README.md          skill 索引、呼叫順序、模型字串
    ├── setup/SKILL.md
    ├── orchestrator/SKILL.md
    ├── story/SKILL.md
    ├── warden/SKILL.md
    ├── npc-chat/SKILL.md
    ├── dreaming/SKILL.md
    ├── offstage-fate/SKILL.md
    └── compactor/SKILL.md
```

## 建議閱讀順序

00 → 01 → 02 → 03 → 04 → 06 → 07 → 08 → 05 → skills/README.md → 各 SKILL.md

（07/08 是工程實作層，動代碼前必讀；它們把設計補到「照著寫」的程度）

## 核心理念速覽

1. **世界真相，不是劇情骨架**：雙層 bible（real / revealed），暴雷在結構上不可能。
2. **分鏡迴圈，停在決策點**：story agent 寫到主角抉擇前停筆，可串流。
3. **活的世界**：NPC 透過 dreaming 演化情緒與意圖；離場 NPC 有獨立命運機制。
4. **逃生即推理**：威脅給壓力，資訊給出路，為了逃而被迫解謎。

## 技術棧

後端 Python（core）+ 網頁前端（HTML/CSS/JS）+ pywebview 包成桌面應用。
後端 Python（core）+ 網頁前端（HTML/CSS/JS）+ pywebview 包成桌面應用。
pywebview 橋接 JS↔Python（同程序、現成綁定）。改用網頁因體驗命脈在文字動畫與排版，
且 CSS 跨電腦一致性最佳（字體內嵌+響應式布局），原生 GUI 反易因系統字體/DPI 跑版。
LLM 串流放 Python 背景 thread。OpenRouter API，三層模型分層（Heavy/Medium/Light）。
OpenRouter API，三層模型分層（Heavy/Medium/Light）。

## 下一步

設計階段完成，已工程化到可開工。**用 AI 實作時，從 `build/` 開始**：
1. 讀 `build/DESIGN-CHANGES.md`（大方向）+ `build/README.md`（工作流程）
2. 照 `build/BUILD-PLAN.md` 一次做一個工單，每次貼 `build/CONTRACTS.md` 防 AI 失憶
3. 對照 `build/CHECKLIST.md` 驗收，過了再下一個

先做 MVP-A（19 個工單）：核心循環跑穩 30 beat，不碰 audio/dreaming/配圖。
跑穩就是專題的說服力。


---

## 修訂版說明（2026-06-01）

本版已完成契約統一修補：
- setup agent 輸出補上 `scene_registry`。
- MVP-A / MVP-B 範圍重新對齊：MVP-A 只要求 30 beat 核心循環，MVP-B 才要求至少一種結局。
- 串流解析責任統一到 Python 後端 `StreamParser`，前端只渲染已分類事件。
- API contract 補 `get_game_state()`、`NA.onStatus()`、`NA.onError()`。
- Pydantic schema 改用 `Field(default_factory=...)`。
- SQLite 補 `runs`、`schema_meta`、index / unique constraints。
- Warden fallback 改為「本地硬規則 → LLM 語義判斷 → 正常推進」。
