# Nightmare Assault 平行開發可行性分析與計劃書

## 1. 執行摘要

本方案可以平行開發，但必須採用 **Contract-first parallel development**，也就是先凍結資料模型、API、事件、SQLite schema、串流解析規格，再讓不同開發者或 AI agent 分頭施工。

不建議直接把 19 個工單平均分給多人同時寫，因為本專案的核心風險不是 UI 或一般 CRUD，而是：

- 多 agent 之間的資料契約是否一致
- LLM 輸出是否能穩定解析
- Blackboard / Snapshot / Compactor 是否會造成狀態污染
- 前後端串流是否同步
- story 是否真的只讀 revealed_bible，不暴露 real_bible

因此，平行開發可行，但必須有一位「架構整合者」負責合約、review、merge、整合測試。其他人或 AI agent 只能在合約範圍內實作，不可自由修改共用介面。

建議目標：

- **MVP-A**：穩定遊玩 30 beat，不暴雷，JSON 可 repair，存讀檔正確。
- **MVP-B**：技能宣稱封頂、一種結局、輕量 NPC evolving、道具庫完整。

若 1 人 + AI 輔助，MVP-A 約 25–35 天。  
若 3–4 人或 3–4 條 AI 工線平行開發，MVP-A 約 16–22 天，但整合風險較高。

---

## 2. 可行性判斷

### 2.1 技術可行性

技術上可行。方案目前已經具備以下工程化基礎：

- 明確的 Pydantic models
- 固定的 OpenRouterClient 介面
- 固定的 StreamParser 介面
- 固定的 pywebview API
- 明確的 SQLite schema 方向
- 以 SKILL.md 作為 agent prompt 來源
- build plan 已拆成可交付工單

這使得平行開發具備必要前提。

### 2.2 平行開發可行性

可以平行的部分：

- 前端 UI 可以用 mock API 先做
- SQLite / Blackboard / Scene 可以不依賴 LLM 先做
- StreamParser 可以用假 token 測試
- SignalBus / constants / models 可以獨立實作
- event extraction 可以獨立實作
- Agent wrapper 可以用 mock client 先測
- QA 可以同步建立 fixtures 與測試案例

不能太早平行的部分：

- story agent 與 beat loop 必須等 parser / orchestrator / blackboard 穩定
- compactor 必須等 snapshot / beat schema 穩定
- 真 LLM 測試必須等 SkillCaller / trace / repair 具備基本能力
- ending / dreaming / offstage-fate 不應放進 MVP-A

結論：

> 平行開發是可行的，但不是「功能平行」，而是「模組邊界平行」。

---

## 3. 平行開發總原則

### 原則一：先凍結契約，再寫功能

以下文件與介面一旦進入工單 1，原則上不得任意改動：

- `core/models.py`
- `core/constants.py`
- `StreamParser` 事件格式
- `OpenRouterClient` 介面
- `API` 方法名稱
- `NA.*` 前端事件名稱
- SQLite table 名稱與必要欄位

若真的需要改，需要走「契約變更流程」。

### 原則二：所有 LLM agent 先 mock，再接真 API

不要一開始就讓 setup / story / warden 全部打真 API。先用 mock output 確認流程能通，再接真模型。

### 原則三：前端不解析 LLM 原始字串

前端只接受後端已分類事件：

- `NA.appendToken(tok)`
- `NA.onContinue()`
- `NA.onDecision(json)`
- `NA.onStatus(status)`
- `NA.onError(error)`
- `NA.onBeatComplete()`

`<<<DECISION>>>` 與 JSON repair 必須全部在 Python 後端完成。

### 原則四：非同步 agent 只能產生 patch

compactor / dreaming / offstage-fate 不直接寫主 Blackboard，只能產生 pending patch，由 beat 安全點統一 merge。

### 原則五：每天至少一次整合測試

平行開發最常死在「每個模組都說自己好了，但接不起來」。因此每天至少合併一次到 integration branch，跑 smoke test。

---

## 4. 建議團隊分工

### 角色 A：架構整合者 / Tech Lead

負責：

- 維護 CONTRACTS
- 審查 PR 是否改壞契約
- 負責 integration branch
- 處理跨模組衝突
- 最終 E2E 測試

不可省略。即使只有一個人開發，也要把這個角色保留，避免自己同時亂改所有模組。

### 角色 B：Core State / Persistence

負責：

- `core/models.py`
- `core/blackboard.py`
- `core/scene.py`
- `core/persistence/db.py`
- snapshot / save / load

### 角色 C：LLM Infrastructure / Parser

負責：

- `core/llm/client.py`
- `core/llm/parser.py`
- fallback / timeout / trace
- JSON repair
- SkillCaller base

### 角色 D：Agent Logic

負責：

- setup agent
- orchestrator
- warden
- story agent
- event extraction
- compactor

### 角色 E：Frontend / pywebview

負責：

- `webview_app.py`
- `ui/index.html`
- `ui/css/*`
- `ui/js/*`
- mock API
- streaming rendering
- state / error UI

### 角色 F：QA / Test / Demo

負責：

- fixtures
- mock LLM outputs
- parser 測試
- snapshot 測試
- 30 beat 模擬測試
- demo script
- bug reproduction cases

若人力不足，建議合併方式：

- 2 人：A+B+D，一人；C+E+F，一人
- 3 人：A+B，一人；C+D，一人；E+F，一人
- 4 人：A，一人；B，一人；C+D，一人；E+F，一人
- 1 人 + AI：自己當 A，其他 B/C/D/E/F 由不同 AI 對話分工，但每次都要貼 CONTRACTS

---

## 5. 模組依賴圖

```text
Contracts / Models / Constants
        ↓
Blackboard ─ Scene ─ SQLite ─ Snapshot
        ↓              ↓
SkillCaller ─ LLM Client ─ StreamParser
        ↓              ↓
setup / orchestrator / warden / story
        ↓
beat 主迴圈
        ↓
pywebview API
        ↓
Frontend UI
```

可以先平行的支線：

```text
Frontend mock UI ────────────────┐
Parser tests ────────────────────┤
SQLite snapshot tests ───────────┤→ Integration
Agent mock outputs ──────────────┤
QA fixtures ─────────────────────┘
```

---

## 6. 開發階段與時程

以下以 4 條工線平行估算。

### 第 0 階段：契約凍結與專案地基

時間：1–2 天

任務：

- 建 repo 結構
- 建 `core/`, `ui/`, `tests/`, `config/`
- 確認 CONTRACTS 與 design 文件一致
- 建 formatter / linter / pytest
- 建 mock LLM fixtures
- 建 integration branch

驗收：

- `pytest` 可跑
- 空專案結構完整
- 所有角色知道不可改契約

---

### 第 1 階段：核心資料層與測試地基

時間：3–5 天

可平行：

| 工線 | 任務 |
|---|---|
| B | Pydantic models、Blackboard、SceneRegistry |
| B | SQLite schema、snapshot、save/load |
| C | constants、SignalBus、StreamParser 初版 |
| E | pywebview 空殼、UI mock 畫面 |
| F | models/parser/db fixtures 與測試 |

驗收：

- models 全部可驗證
- blackboard patch + version 可運作
- SQLite 可存假 beat 並讀回
- parser 可解析正常分隔符與 fallback decision
- 前端 mock 畫面能顯示假 narrative / decision

---

### 第 2 階段：LLM 基礎與 Agent 外殼

時間：4–6 天

可平行：

| 工線 | 任務 |
|---|---|
| C | OpenRouterClient、fallback、llm_traces |
| C | SkillCaller base、SKILL.md 載入 |
| D | event extraction |
| D | setup / orchestrator / warden mock 版本 |
| E | streaming renderer 接 mock 事件 |
| F | API mock 測試、LLM trace 測試 |

驗收：

- LLM 可用最小 prompt 回 OK
- fallback 觸發可記錄 trace
- SkillCaller 可載入任一 SKILL.md
- setup mock 可寫入 blackboard
- orchestrator mock 可搬 revelation
- warden mock 可給 directive_to_story
- 前端只吃後端事件，不解析 JSON

---

### 第 3 階段：核心 Agent 真實接入

時間：5–7 天

可平行：

| 工線 | 任務 |
|---|---|
| D | setup 真實 LLM 接入 |
| D | orchestrator 條件判斷 + Light LLM 語義輔助 |
| D | warden 本地硬規則 + LLM 語義判斷 |
| C | parser repair 強化與測試 |
| E | 前端狀態、busy、error、decision UI |
| F | 防暴雷測試、warden 測試、setup schema 測試 |

驗收：

- setup 可產生 real/revealed bible、NPC、scene_registry
- story 尚未接入前，也能用 fake story 跑一個流程
- warden 本地硬規則在 LLM 掛掉時仍有效
- orchestrator 不會提前揭露未達條件碎片
- UI 可顯示 generating / awaiting_decision / error

---

### 第 4 階段：story agent + beat 主迴圈

時間：5–7 天

這是整合風險最高階段，不建議過度平行。

任務：

- story agent 串流接入
- player input 包 `<player_action>`
- story 只讀 revealed_bible
- StreamParser 接真 story output
- beat loop 串起 warden → orchestrator → story → snapshot → compactor check
- pywebview API 接真後端

驗收：

- 新局 → setup → 開場 → 決策 → 下一 beat 可跑通
- story 不輸出未 revealed fragment
- JSON 壞掉時可 repair 或 fallback
- 每 beat 有 snapshot
- 前端可以完整操作一輪

---

### 第 5 階段：Compactor 與 30 beat 穩定性

時間：4–6 天

任務：

- rolling summary
- fact ledger
- beat window
- 30 beat 模擬測試
- 回 beat 10 不帶 beat 30 摘要
- llm_traces 成本與 latency 檢查

驗收：

- 30 beat 不爆 context
- 核心伏筆仍可被引用
- 摘要長度有上限
- 快照還原正確
- 連續遊玩不崩潰

---

### 第 6 階段：MVP-A 打磨與 Demo

時間：3–5 天

任務：

- 修復 UI 細節
- 錯誤訊息可讀化
- loading / disabled / retry
- 新局選項整理
- 建 demo script
- 建已知問題列表
- 成本測試

驗收：

- 30 beat demo 可穩定跑
- 出錯能恢復或提示
- 老師/評審可以理解核心亮點：雙層 bible、防暴雷、compactor、warden、可回溯快照

---

## 7. MVP-A 里程碑

### M1：資料層完成

條件：

- models / blackboard / scene / db / snapshot 測試通過
- 尚不需要 LLM

### M2：LLM 管線完成

條件：

- OpenRouterClient 可 call / stream
- parser 可處理正常、壞 JSON、分隔符拆 token、fallback
- SkillCaller 可載入 SKILL.md

### M3：核心 agent 完成

條件：

- setup / orchestrator / warden / story 可用
- story 只讀 revealed_bible
- warden 本地硬規則有效

### M4：端到端閉環

條件：

- pywebview 新局 → 開場 → 決策 → 下一 beat
- 前端不解析 LLM raw text
- 後端狀態同步正確

### M5：30 beat 穩定

條件：

- 連續 30 beat 不崩潰
- JSON repair 成功率達標
- snapshot / summary 正確
- demo 可展示

---

## 8. Git 與分支策略

建議分支：

```text
main              穩定版，只放可 demo 版本
develop           開發整合分支
integration       每日整合測試分支
feature/core-*    Core state / db
feature/llm-*     LLM / parser
feature/agents-*  agents
feature/ui-*      frontend
feature/test-*    QA fixtures / tests
```

合併規則：

- feature → integration：每天合一次
- integration 測試通過 → develop
- develop 穩定 demo → main
- 改 CONTRACTS 必須由架構整合者批准
- PR 必須附測試結果或失敗原因

禁止事項：

- UI 分支直接改 Pydantic models
- agent 分支直接改 SQLite schema
- parser 分支直接改前端事件名稱
- 多人同時改 `CONTRACTS.md`
- story agent 讀取 real_bible

---

## 9. 測試策略

### 9.1 單元測試

必做：

- models validation
- parser 分隔符偵測
- JSON repair
- blackboard patch version
- scene movement
- SQLite save/load
- warden local hard rules
- orchestrator reveal conditions

### 9.2 整合測試

必做：

- setup → blackboard
- warden → orchestrator → story fake
- story real → parser
- beat loop → snapshot
- load old snapshot
- frontend mock API
- frontend real API

### 9.3 穩定性測試

必做：

- 30 beat fake simulation
- 30 beat real-lite simulation
- API timeout 模擬
- JSON 壞掉模擬
- LLM fallback 模擬
- 玩家 prompt injection 模擬

### 9.4 防暴雷測試

方法：

- real_bible 中放入明確 forbidden fragment
- revealed_bible 不放該 fragment
- story 生成後檢查 forbidden fragment 不得出現
- 每個 beat 都檢查一次

### 9.5 回溯測試

方法：

- 產生 30 beat
- 儲存 beat 10 snapshot
- 載入 beat 10
- 確認 rolling_summary 不含 beat 11–30 的資訊
- 確認 inventory / npc state / scene state 都是當時版本

---

## 10. 主要風險與對策

| 風險 | 嚴重度 | 發生率 | 對策 |
|---|---:|---:|---|
| LLM JSON 壞掉 | 高 | 高 | StreamParser 三級 repair + fallback decision |
| story 暴露 real_bible | 高 | 中 | story context 永遠只傳 revealed_bible + 自動 forbidden fragment 測試 |
| 平行開發改壞契約 | 高 | 高 | Contract freeze + tech lead review |
| compactor 摘要丟伏筆 | 高 | 中 | fact ledger + protected facts + 30 beat 測試 |
| 前後端串流不同步 | 中高 | 中 | 後端唯一解析，前端只吃事件 |
| snapshot 回溯污染 | 高 | 中 | 每 beat 儲存當時 summary + blackboard snapshot |
| 成本失控 | 中 | 中 | model tiering + trace + mock testing |
| UI 先做太華麗拖慢核心 | 中 | 高 | 前端先 mock + 黑白恐怖風格即可 |
| MVP 範圍膨脹 | 高 | 高 | MVP-A 不做 dreaming / ending 多路 / audio / image |

---

## 11. MVP-A 不做清單

為了確保專題能完成，MVP-A 明確不做：

- 完整 dreaming
- offstage-fate
- npc-chat 聊天室
- 多結局系統
- 音訊與音樂生成
- 配圖
- 分支樹
- 多模板
- 複雜動畫
- 完整回溯 UI

MVP-A 只證明：

> 這是一個能長時間穩定運行、不暴雷、可存讀檔、有狀態記憶的 AI 恐怖文字遊戲核心引擎。

---

## 12. MVP-B 擴充計劃

MVP-A 穩定後，進入 MVP-B。

目標：

- 技能宣稱封頂更完整
- 至少一種結局可達
- 雙層 bible 防暴雷驗證報告
- 輕量 NPC evolving
- 道具庫完整

時間：7–12 天。

建議順序：

1. 強化 warden 技能封頂
2. 加一種結局序列，不做多結局
3. NPC evolving 只做在場 NPC，每 5 beat 更新一次
4. 道具庫接 revelation item
5. 製作 demo report

---

## 13. 建議實作順序總表

| 週期 | 主線 | 可平行支線 | 目標 |
|---|---|---|---|
| Day 1–2 | 契約凍結、repo、CI | mock fixtures | 開工不混亂 |
| Day 3–6 | models / db / blackboard | parser / UI mock | 地基完成 |
| Day 7–10 | LLM client / SkillCaller | event extraction / streaming UI | LLM 管線完成 |
| Day 11–15 | setup / orchestrator / warden / story | parser repair / UI state | 核心 agent 完成 |
| Day 16–20 | beat loop / snapshot / compactor | front-back integration | E2E 閉環 |
| Day 21–24 | 30 beat 測試 | demo polish | MVP-A 驗收 |
| Day 25–35 | MVP-B | 報告與展示 | 差異化展示 |

---

## 14. 結論

此方案適合平行開發，但需要嚴格控制契約與整合流程。

最推薦的策略是：

1. 先凍結 contracts。
2. Core / LLM / Frontend / QA 四線並行。
3. 所有 agent 先用 mock output 跑通。
4. 每天整合一次。
5. MVP-A 只追求 30 beat 穩定，不做炫技功能。
6. MVP-A 穩定後再做 MVP-B 的結局、技能封頂、輕量 NPC evolving。

若照此方式執行，這個專題有機會在合理時間內做出一個具有展示說服力的 AI 文字冒險引擎，而不是停留在 prompt demo 或聊天機器人外殼。
