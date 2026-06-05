# 08 · 工程實作（持久化、並行、錯誤、測試）

> 把記憶/狀態/錯誤從「方向」補到「可開工」。SQLite 表、並行 patch、錯誤恢復骨架、測試驗收。
> 錯誤處理的**細部策略待實作撞到真實 LLM 行為後補完**；本檔給足以開工的骨架。

---

## 一、SQLite Schema（補 #6）

```sql
-- 每場遊戲一列，list_saves / 狀態恢復 / 多存檔都從這裡開始
CREATE TABLE runs (
  run_id TEXT PRIMARY KEY,
  title TEXT,
  theme TEXT,
  difficulty TEXT,
  created_at TEXT,
  updated_at TEXT,
  current_beat INTEGER DEFAULT 0,
  current_location TEXT,
  status TEXT DEFAULT 'active'
);

-- schema 版本，方便日後遷移
CREATE TABLE schema_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

INSERT OR REPLACE INTO schema_meta(key, value) VALUES ('schema_version', '1');

-- 每個 beat 一列，含當時的完整快照（回溯靠它）
CREATE TABLE beats (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  beat_number INTEGER NOT NULL,
  narrative TEXT,
  decision_json TEXT,                 -- DecisionPoint 序列化
  rolling_summary_snapshot TEXT,      -- 當時的摘要（防回溯錯亂，見 02 §九）
  blackboard_snapshot_json TEXT,      -- 完整 Blackboard 快照
  is_narration_only INTEGER DEFAULT 0,
  created_at TEXT
);

-- NPC 狀態（含對玩家隱藏的離場命運，獨立欄位）
CREATE TABLE npc_states (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  beat_number INTEGER NOT NULL,
  npc_name TEXT NOT NULL,
  state_json TEXT,                    -- evolving（玩家可間接感知的）
  hidden_state_json TEXT              -- 離場命運/secret 相關（路線 A：不主動劇透）
);

-- 道具庫快照
CREATE TABLE inventory_snapshots (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  beat_number INTEGER NOT NULL,
  inventory_json TEXT
);

-- 完整聊天紀錄（cold 層，永不遺失）
CREATE TABLE chat_logs (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  npc_name TEXT,
  beat_number INTEGER,
  role TEXT,                          -- player | npc
  content TEXT,
  created_at TEXT
);

-- 存檔點（flag 標記哪些 beat 是具名存檔）
CREATE TABLE save_points (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  beat_number INTEGER NOT NULL,
  label TEXT,
  created_at TEXT
);

-- LLM 追蹤（多 agent debug 命脈，務必有）
CREATE TABLE llm_traces (
  id INTEGER PRIMARY KEY,
  run_id TEXT NOT NULL,
  beat_number INTEGER,
  agent TEXT,                         -- setup/orchestrator/story/...
  model TEXT,
  prompt_hash TEXT,
  input_tokens INTEGER,
  output_tokens INTEGER,
  latency_ms INTEGER,
  success INTEGER,
  error TEXT,
  created_at TEXT
);

-- 常用查詢與一致性約束
CREATE UNIQUE INDEX idx_beats_run_beat
ON beats(run_id, beat_number);

CREATE INDEX idx_npc_states_run_beat
ON npc_states(run_id, beat_number);

CREATE INDEX idx_inventory_run_beat
ON inventory_snapshots(run_id, beat_number);

CREATE INDEX idx_chat_logs_run_npc
ON chat_logs(run_id, npc_name);

CREATE INDEX idx_llm_traces_run_beat
ON llm_traces(run_id, beat_number);
```


`runs` 是存檔列表與目前進度的入口；`schema_meta` 讓日後改 schema 時有遷移依據；index/unique constraint 讓回溯、debug 與多存檔查詢不混亂。

`llm_traces` 特別重要：多 agent 系統出錯時，你必須能查「第幾 beat、哪個 agent、用哪個模型、是否成功」，否則 debug 會非常痛苦。MVP 就要有。

---

## 二、狀態並行控制（補 #5）

非同步是設計選擇（compactor/dreaming/offstage-fate 在玩家讀字時跑），代價是並行寫 Blackboard 的競爭。

```
危險情境:
  story 正在讀 npc_registry
  dreaming 同時寫 npc_registry
  compactor 同時改 rolling_summary
  玩家又送下一個決策
```

### 解法：版本化 + patch，非同步只產 patch 不直接寫

```yaml
blackboard:
  version: int                 # 每次安全點 merge 後 +1
  beat_number: int
```

```
非同步 agent 不直接改主 Blackboard，而是產生 patch:
  {
    "base_version": 18,        # 基於哪個版本算的
    "target": "npc_registry.張醫生.evolving",
    "patch": {...}
  }

主迴圈在「安全點」（BEAT_COMPLETED 後、下個 beat 開始前）套用:
  collect pending patches
  for p in patches:
      if p.base_version == current_version:  apply
      else:                                  rebase 或丟棄（基於過期狀態）
  version += 1
  snapshot
```

### 玩家搶快（送下個決策時非同步還沒跑完）

```
規則:
  story（同步、玩家等待中）擁有讀取優先權，讀的是「上個安全點的穩定快照」
  非同步 agent 的 patch 只在安全點套用，不會在 story 讀到一半改它
  若玩家送出時 compactor 還沒壓縮完 → 用當前（未壓縮）狀態繼續，
    壓縮結果延後到下個安全點生效（晚一拍，不阻塞玩家）
  context 達 L3 緊急才例外阻塞（見 02 §八）
```

核心：**同步路徑（玩家互動）永遠讀穩定快照；非同步只在安全點寫入。** 玩家永不被背景任務卡住。

---

## 三、錯誤處理骨架（補 #9）

> 骨架夠開工；repair prompt 具體措辭、各模型怪癖等**待實作時補**。

```
LLM 呼叫統一走一個 resilient_call():
  try 主模型（含 timeout）
  except timeout / rate_limit / error:
      → fallback 鏈（見 01 §三）：主 → 次 → 再次 → 本地邏輯
  全失敗:
      story    → fallback decision UI（見 07 §三 L3），遊戲續行
      warden   → 先跑本地 deterministic hard rule；若未命中才保守「正常推進」
      orchestrator → 本 beat 不揭露（揭露可延後，不致命）
      dreaming/offstage/compactor → 非同步，靜默跳過本輪，下輪再試
      setup    → 唯一不可降級：開局失敗就重試/報錯（沒有世界無法開始）

JSON 解析失敗 → 07 §三 三級 repair
串流中斷     → 已串流 narrative 保留，決策走 fallback UI
API key 錯誤 → 設定畫面提示，不進遊戲
```

### Warden fallback 的優先順序

Warden 不可單純因 LLM 失敗就放過所有情況，否則玩家明確觸犯致命規則時可能逃過死亡。因此順序固定：

```
1. 本地 deterministic hard rule check
   - 關鍵詞 / 正則 / 明確動作比對
   - 若明確命中 deadly_rule 或 hard ending → 直接觸發，不需要 LLM
2. LLM semantic judgment
   - 用於模糊語義、技能宣稱、軟結局 gate
3. LLM 全失敗
   - 未命中本地硬規則 → 保守正常推進，避免誤殺玩家
```

**設計原則**：除了 setup，**沒有任何單一 agent 失敗能讓遊戲整個掛掉**。每個都有降級路徑。warden 失敗時偏向「不誤殺玩家」（保守正常推進），但本地硬規則仍優先於這個 fallback。

---

## 四、API Key 安全（補 #8）

不用「混淆」當安全方案。明確分級：

```
開發版:   .env（不進版控，.gitignore）
桌面版:   系統 keyring（python keyring 套件，存 OS 憑證庫）

請求路徑（key 永不到前端）:
  JS → window.pywebview.api → Python 後端 → OpenRouter
  ❌ 絕不 JS → OpenRouter（key 會暴露在前端）

前端只拿得到「遊戲內容」，拿不到 key。所有 OpenRouter request 由 Python 發。
```

---

## 五、測試與驗收指標（補 #10）

定義「怎樣算成功」，讓專題像工程系統而非單純 AI 聊天。

| 測試 | 指標 | 對應風險 |
|------|------|---------|
| 連續遊玩 | 30 beat 不崩潰 | context 爆掉 |
| 防暴雷 | story 不輸出未 revealed 的 fragment | 雙層 bible 是否真有效 |
| JSON 穩定 | ≥95% 輸出可解析，失敗可 repair | 串流格式 |
| 記憶壓縮 | 30 beat 後核心伏筆仍可被引用 | compactor 品質（成敗關鍵） |
| 技能宣稱 | 破格被限、不破格被接受 | warden 封頂 |
| 存檔 | 回 beat 10 不帶 beat 30 的摘要 | 快照含當時摘要 |
| 並行安全 | 非同步寫入不污染 story 讀取 | patch 機制 |
| 成本 | 每 50 beat 成本 < 上限 | dreaming 隨 NPC 線性成長 |
| 延遲 | 一般 beat 等待 < 門檻 | 串流 TTFT |
| Injection | 玩家注入指令不破壞格式/不暴雷 | prompt 包裝 |

**最關鍵三項**（demo 必驗）：防暴雷、30 beat 連貫（記憶壓縮）、JSON 穩定。這三個是你架構亮點的直接證明。

---

## 六、實作建議：先寫測試骨架再寫功能

針對「LLM 系統 debug 很耗時間」，建議：
- 每個 agent 寫一個「黃金樣本」測試（固定輸入 → 驗證輸出符合 Pydantic schema）。
- compactor 寫「30 beat 模擬」測試（餵假 beat，驗證伏筆保留、摘要有界）。
- 防暴雷寫斷言（掃 story 輸出，確認不含未 revealed fragment 的 content）。
- `llm_traces` 從第一天就記錄，省下後期 debug 的痛。
