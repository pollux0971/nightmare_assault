# BUILD PLAN · 施工序列

> **用途**：把系統拆成 AI 一次能吃下的「工單」。每個工單小到單次對話能完成、結束都能獨立測試、有明確依賴順序。
> **怎麼用**：一次只做一個工單。對 AI 說「做工單 N」，貼上該工單內容 + CONTRACTS.md 相關段落。通過驗收才進下一個。
> **範圍**：本檔涵蓋 MVP-A（核心循環跑穩 30 beat）。MVP-B 與後續階段在末尾簡列。

---

## 施工原則（每個工單都遵守）

1. **先寫測試骨架，再寫實作**——每個工單附測試，AI 寫完要能跑過。
2. **不碰 LLM 的先做**——資料類、狀態容器、SQLite 先用假資料測通，把 LLM 留到後面。
3. **每個工單結束 = 一個可驗證的里程碑**——不是「寫一半」，是「這塊能單獨測」。
4. **開頭必貼 CONTRACTS**——尤其資料類與函式簽名，防止 AI 重新發明接口。
5. **失敗就回報，不硬湊**——AI 若覺得契約有問題，停下來問，不自作主張改。

---

## 階段 0：契約統一與地基（先對齊，再寫功能）

### 工單 0 — 契約統一與文件修補
- **做什麼**：開工前先確認設計文件與 CONTRACTS 一致，不寫功能。
- **檢查項**：setup skill 必含 `scene_registry`；MVP-A 不要求結局、MVP-B 才要求一種結局；串流解析責任在 Python 後端；API 含 `get_game_state/NA.onStatus/NA.onError`；Pydantic 使用 `Field(default_factory=...)`；SQLite 含 `runs/index/schema_meta`；warden fallback 先跑本地硬規則。
- **測試**：不用跑程式；用 grep/checklist 確認上述字串與段落存在。
- **驗收**：契約無互相矛盾處，再進工單 1。

### 工單 1 — Pydantic 資料類
- **做什麼**：照 CONTRACTS 二 + 07 全文，把所有資料類寫進 `core/models.py`，所有 list/dict/model 預設值一律用 `Field(default_factory=...)`。
- **依賴**：無。
- **測試**：`tests/test_models.py`——每個類能用合法 dict 建立、非法值被拒、預設值正確。
- **驗收**：所有資料類可 import、可驗證。不含任何邏輯。

### 工單 2 — 常數與 SignalBus
- **做什麼**：`core/constants.py`（CONTRACTS 六）+ `core/signal.py`（發布/訂閱，見 01 §五的事件表）。
- **依賴**：無。
- **測試**：訂閱一個事件、發布、確認回呼被呼叫。
- **驗收**：SignalBus 能收發 01 §五列的所有事件名。

### 工單 3 — Blackboard + 版本/patch
- **做什麼**：`core/blackboard.py`。容器持有 03 的完整 schema；實作 08 §二的版本化 + patch：`apply_patch(patch)`、`collect_pending()`、安全點 `merge_and_bump()`。
- **依賴**：工單 1。
- **測試**：建立 Blackboard、套用一個 patch、版本 +1；過期 base_version 的 patch 被拒。
- **驗收**：並行 patch 機制可單獨測（用假 patch，不需 LLM）。

### 工單 4 — 場景系統
- **做什麼**：`core/scene.py`。SceneRegistry：當前位置、移動、`location_reached(id)` 判斷、種植/揭露 interactable。
- **依賴**：工單 1、3。
- **測試**：建立場景圖、移動、抵達判斷、種一個 corpse interactable 並標記 revealed。
- **驗收**：location_reached 與 interactable 增刪查正確。

### 工單 5 — SQLite 持久化
- **做什麼**：`core/persistence/db.py`。建 08 §一的 `runs/schema_meta/beats/npc_states/inventory_snapshots/chat_logs/save_points/llm_traces`，加 index/unique constraint；實作 beat 快照存/讀、存檔點、`llm_traces` 寫入。
- **依賴**：工單 1、3。
- **測試**：存一個假 beat 快照（含當時摘要）、讀回、確認 blackboard_snapshot 還原正確；回溯到舊 beat 不帶新摘要。
- **驗收**：存讀檔正確、快照含當時摘要（驗收評審最在意的回溯陷阱）。

---

## 階段 1：LLM 接入（先用最小 prompt 測通管線）

### 工單 6 — OpenRouter Client + fallback
- **做什麼**：`core/llm/client.py`（CONTRACTS 三）。同步 call、stream generator、fallback 鏈、timeout、每次寫 llm_traces。
- **依賴**：工單 1、2、5。
- **測試**：用一個最小 prompt（「回我 OK」）測真實 call 成功；模擬主模型失敗測 fallback 觸發；確認 trace 寫入。
- **驗收**：能呼叫三層、fallback 會動、trace 有記錄。**這是第一次碰真 API。**

### 工單 7 — SkillCaller 基類 + SKILL.md 載入
- **做什麼**：`core/agents/base.py`。讀 `skills/{agent}/SKILL.md`、組 prompt（system=SKILL 內容、user=結構化 context）、呼叫 client、用對應 Pydantic 類驗證輸出。熱重載。
- **依賴**：工單 1、6。
- **測試**：載入一個 SKILL.md、組 prompt、mock client 回傳、驗證輸出符合 schema。
- **驗收**：任一 agent 可透過基類呼叫並得到驗證過的輸出。

### 工單 8 — 串流解析管線（承重牆）
- **做什麼**：`core/llm/parser.py`（CONTRACTS 四）。StreamParser：逐 token、滑動視窗偵測分隔符、分離 narrative/decision、三級 repair、fallback decision UI。實作 07 §三全部。
- **依賴**：工單 1。
- **測試**（重點多）：正常 JSON 解析過；缺逗號被 L1 修復；分隔符被拆成多 token 仍偵測到；忘記 DECISION 走 fallback；串流中斷 narrative 保留。
- **驗收**：07 §三所有情境都有測試覆蓋。**這個不穩，整個遊戲不穩。**

---

## 階段 2：核心 agent（依賴前面，逐個接）

### 工單 9 — setup agent
- **做什麼**：`core/agents/setup.py`。輸入主題等 → 呼叫 LLM → 驗證 SetupOutput → 寫入 Blackboard（real_bible/npc_registry/scene_registry/protagonist）+ 回傳開場序列。
- **依賴**：工單 3、4、7。
- **測試**：給一個主題，跑真實 setup，驗證輸出含雙層 bible、NPC 有 self_aware、scene_registry 非空、開場序列非空。
- **驗收**：輸入主題能生出結構完整、可玩的世界。

### 工單 10 — orchestrator（揭露閘門）
- **做什麼**：`core/agents/orchestrator.py`。多數揭露條件**程式碼判**（min_beats/location_reached/requires_touched）；語義觸及才呼叫 Light LLM。把碎片從 real 搬到 revealed。
- **依賴**：工單 3、4、7。
- **測試**：設一個 min_beats=3 的碎片，beat<3 不揭露、>=3 揭露；location_reached 條件達成才揭露。
- **驗收**：揭露閘門按條件運作；story 永遠拿到 newly_revealed。

### 工單 11 — event 抽取（程式碼層）
- **做什麼**：`core/events.py`。純函式：玩家輸入 + story 輸出 → 結構化 events（searched_location/questioned_npc/picked_item/reached_location）。供 orchestrator 佐證，不信 story 自報。見 07 §四。
- **依賴**：工單 1。
- **測試**：餵幾組輸入，驗證抽出正確 events。
- **驗收**：零 LLM 成本的事件抽取可用。

### 工單 12 — warden（僅玩家）
- **做什麼**：`core/agents/warden.py`。先跑本地 deterministic hard rule（關鍵詞/正則/明確動作）；未命中才呼叫 LLM 做語義判斷、結局條件（硬/軟 gate）、技能宣稱封頂。輸出 WardenOutput。
- **依賴**：工單 3、7。
- **測試**：違反致命規則觸發死亡；軟結局 gate 未過不結束；破格技能被封頂、合理技能被接受；模擬 LLM 失敗時本地硬規則仍可觸發。
- **驗收**：本地硬規則命中時即使 LLM 失敗也能觸發；未命中硬規則且 LLM 失敗時才保守「正常推進」不誤殺。

### 工單 13 — story agent + 串流
- **做什麼**：`core/agents/story.py`。讀 revealed_bible + 摘要 + 視窗 + 玩家決定 + NPC evolving + directive + newly_revealed → 串流生成（含分隔符）→ 經 parser 得 narrative + DecisionPoint。玩家輸入包 `<player_action>`。決策型/旁白型判斷。
- **依賴**：工單 7、8、10。
- **測試**：給一個 revealed_bible 與玩家決定，跑真實 story，驗證串流出 narrative、停在決策點、DecisionPoint 合法、不含未 revealed 的 fragment（防暴雷斷言）。
- **驗收**：能接玩家決定生成 beat 並停在決策點；防暴雷測試過。

---

## 階段 3：記憶與迴圈（讓它能連續跑）

### 工單 14 — 滾動摘要 + ledger（compactor 承重牆）
- **做什麼**：`core/memory/summary.py` + `core/agents/compactor.py`。滑動視窗、滾動摘要（有上限）、fact ledger（二元組）、保護清單、三級壓縮觸發、聊天退出濃縮。非同步。
- **依賴**：工單 3、7。
- **測試**（重點）：模擬 30 個假 beat 餵入，驗證 context 不爆、伏筆保留、摘要有界、回溯到 beat 10 摘要狀態正確。
- **驗收**：**30 beat 模擬連貫**（專題成敗關鍵）。

### 工單 15 — beat 主迴圈
- **做什麼**：`core/orchestrator_loop.py`。串起完整順序：玩家輸入 → warden → orchestrator → story → 串流 → BEAT_COMPLETED → 快照 → 安全點 merge patches → compactor 非同步檢查。實作 08 §二的並行控制（同步讀穩定快照、非同步只在安全點寫）。
- **依賴**：工單 5、12、10、13、14。
- **測試**：跑一個完整 beat 迴圈（用真 LLM），確認順序正確、快照產生、玩家送下個決策時非同步不污染 story 讀取。
- **驗收**：單機後端能連續跑多個 beat，狀態一致。**此時後端 MVP-A 核心已通。**

---

## 階段 4：前端（網頁 + pywebview）

### 工單 16 — pywebview 骨架 + API
- **做什麼**：`webview_app.py`（API class，CONTRACTS 五）+ `ui/index.html` 骨架 + view 切換 + theme.css（深色恐怖）。
- **依賴**：工單 15。
- **測試**：起窗、JS 呼叫 check_config 通、view 能切換。
- **驗收**：桌面視窗能開、前後端能通訊、主題生效。

### 工單 17 — 串流渲染（前端承重牆）
- **做什麼**：`ui/js/streaming.js`。接收後端已分類事件：`NA.appendToken/onContinue/onDecision/onStatus/onError/onBeatComplete`；節奏控速吐字、關鍵詞血紅、CONTINUE 暫停、DecisionPoint 呈現、選項淡入。
- **依賴**：工單 8、16。
- **測試**：mock 後端推「純文字 token + onContinue + 已驗證 decision json」，驗證逐字渲染、暫停、決策呈現、不卡；確認 JS 不解析 `<<<DECISION>>>`。
- **驗收**：逐字節奏到位、UI 不凍結、狀態同步正確；分隔符與 JSON repair 由後端 parser 完成。

### 工單 18 — 遊戲主畫面 + 前置畫面
- **做什麼**：敘事區、決策區、自由輸入、頂部列；啟動/設定/主選單/新局/載入畫面；API key 用 keyring；前端以 `get_game_state/NA.onStatus` 控制 busy/disabled/error。
- **依賴**：工單 16、17。
- **測試**：完整走一遍——設定 → 新局 → 看到第一個 beat 串流 → 做決策 → 下個 beat。
- **驗收**：**端到端最小閉環打通**（MVP-A 完成）。

### 工單 19 — 存檔 UI + 道具面板
- **做什麼**：存檔選擇畫面、自動快照提示、道具面板 modal。
- **依賴**：工單 5、18。
- **驗收**：存讀檔可用、道具可查看。

---

## MVP-A 驗收（全部工單完成後的整合測試）

對照 08 §五，跑這幾項：
- 連續玩 30 beat 不崩潰
- 防暴雷：story 不輸出未 revealed 的 fragment
- JSON ≥95% 可解析，失敗可 repair
- 30 beat 後核心伏筆仍可引用
- 回 beat 10 不帶 beat 30 摘要
- 換解析度不跑版

這六項過了，MVP-A 成立，專題已有說服力。

---

## MVP-B 與後續（A 穩了再拆工單）

```
MVP-B: 技能宣稱封頂強化 + 一種結局序列 + 雙層防暴雷驗證報告
       + NPC 輕量 evolving（簡版 dreaming）+ 道具庫完整
階段 2: 完整 dreaming + offstage-fate + npc-chat 聊天室 + 三層記憶召回 + 回溯 + 音訊
階段 3: 配圖 + 音樂生成 + 分支樹 + 多模板 + 難度 + 復盤
```

這些到時各自再拆工單，方式同上。**現在專注工單 0 + MVP-A 的 19 個工單。**
