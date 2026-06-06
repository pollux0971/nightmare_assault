# 16 · WorldModel — 抽象實體機制（開放式探索的落地層）

> **承接**：`15-player-sovereignty.md`（方向）把 WorldModel 列為「規劃中」；本檔記錄它**已落地**的全部機制。
> **實作對應**：`core/world/model.py`（核心）·`core/world/extractor.py`（story 抽取 fallback）·
> `core/narrative/exit_resolver.py` / `exploration_mode.py` / `action_intent.py` / `truth_evidence_gate.py` /
> `npc_prose_facts.py` + loop `_world_model_tick` / `_apply_review_lock` / `world_progress`。
> **狀態**：已落地、已測（**847 passed，flag OFF/ON 各一次**）。Spatial 投影層另見 `17-spatial-projection.md`。
> **一句話**：**世界裡所有「可被指涉的東西」都是一個 Entity；玩家行動＝對某 Entity 套一個 affordance。**

---

## 〇、為什麼要這層

實機 selfplay 暴露的問題——系統把玩家往深處推、世界記不住敘述過的物件、「退到外面」沒真的移動、
NPC 講的話蒸發、玩家只是找路卻被推進真相——**不是各自的 bug，而是同一個缺失機制的表現：
沒有一個「世界裡可被指涉的東西」的模型。**

舊的 `world_facts`（扁平 string→fact）與字串比對否定，是這個缺口的**權宜替身**。WorldModel 是正解：
一個**主題無關**的實體層，把移動 / 物件 / 出口 / 事實 / NPC / 否定 / 撤退 / 「世界記得」全收成它的投影。

> WorldModel **平行於** kernel：kernel 仍負責場景推進與事件；WorldModel 是 current_area / 區域 / 出口 /
> 實體記憶的**唯一權威**與觀測來源。兩者只在「kernel 真的移動了」時透過 `WorldDelta(move_current)` 同步。

---

## 一、Entity 與狀態機（`core/world/model.py`）

```text
Entity: id / kind / label / state / props / affords / origin / roles
kind ∈ {area, exit, object, actor, fact}
狀態機（抽象，不含主題內容）：
  object: noticed → inspected → taken/used
  exit  : unknown → known → available → locked/blocked/used
  area  : known → visited → current
  actor : present/absent → talked
  fact  : asserted
```

- **核心轉變**：玩家行動＝對某 Entity 套用一個 affordance（`inspect / take / use / move_to / move_through /
  withdraw_to / talk`），不再猜關鍵詞、比字串。套 affordance **必然**改某實體 state ⇒「有後果」變**結構保證**。
- `props` 放非結構欄位（如 exit 的 `area`=from / `leads_to`=to、fact 的 `source`/`confidence`、object 的 `area`）。
- `origin ∈ {kernel, story, npc, extractor}`——每個實體都記得是誰登記的（供權限與除錯）。
- **持久化**：`to_dict` / `from_dict`（存進 `game_meta.world_model`；舊存檔無 `roles` 欄位也能載入）。

| 舊問題 | 在實體模型裡的自然解 |
|---|---|
| stay-put 違規 | 「只在原地」＝無 `move_to` 目標；「不進 B 區」＝拒絕 `move_to(area.B)`（NegativeIntentGuard） |
| 世界記不住 | story 敘述物件 → 登記 `object.X(noticed)`；之後可 `inspect` → 世界記得 |
| 退到外面沒移動 | 安全區是 `role=safe_zone` 的 area 實體 → `withdraw_to_safe_zone` 真的改 current_area |
| NPC 資訊蒸發 | NPC 一句 → `fact` 實體（帶 source/confidence） |
| had_consequence 用猜的 | 套 affordance 必改某實體 state → 後果是結構保證 |

---

## 二、實體怎麼進來：structured entity_delta（優先）＋ 抽取（fallback）

理想路徑是 **story / NPC 直接吐結構化 `entity_delta`**；抽取器只是 LLM 不配合（吐純文字）時的退路。

### 2.1 Story entity_delta（`coerce_entity_deltas`）
- story 在 `<<<DECISION>>>` JSON 附 `entity_delta`；只准 **object / actor / fact**（**不准 area/exit**——地圖由 kernel 圖擁有）。
- 每 beat **≤3** 筆；malformed 一律丟棄（容錯欄位，壞 delta 不讓整個 DecisionPoint 解析失敗）。
- 物件狀態機：`register(noticed)` → `set_state(inspected/taken/used)`；同 label 對到同實體（label-slug 冪等）。

### 2.2 NPC entity_delta（`coerce_npc_entity_deltas`）
- NPC-chat 結構化回報；**只准 fact / actor**（NPC 不得新增 object/area/exit/真相）。
- fact **強制**帶 `props.source=npc_id`、`props.confidence=npc_claim`、`origin=npc`、`state=asserted`。
- 走**獨立通道**進 WorldModel（`_bridge_npc_entity_delta`），**完全不碰 reveal ledger**——NPC 講的事只成
  「世界裡可被指涉的主張」，不會把任何真相推進到 confirmed。

### 2.3 抽取 fallback
- **story 散文**：`core/world/extractor.py::extract_entities`——白名單式（物件類別＋前景化線索詞），不亂登記氛圍名詞。
- **NPC 散文**：`core/narrative/npc_prose_facts.py::extract_npc_prose_facts`——只在 NPC `entity_delta` 缺席時啟用，
  只抽三類明確主張（location / locked_exit / action_required），保留自然語意 label（「通訊設備在B2機房」不是
  `machine_room_known`），≤2 筆，氣氛/比喻/幻覺一律放棄。

> ⚠️ **prompt 錨定坑**：早期 skill 範例硬寫 `label:"WU 袖扣"`、`"通訊設備在機房"`，LLM few-shot 會照抄 →
> 每局都冒同一物件。已改成占位符＋「嚴禁照抄範例」指示；複測四局四樣。

---

## 三、出口擁有權與 ExitResolver（`exit_resolver.py`）

- **WorldModel owns exits**：`register_exit(label, from_area=, leads_to=, state=)` / `set_exit_state` /
  `move_through`（尊重否定＋狀態：`locked/blocked/used` 不可通行）/ `exits_from(area)`。
- **ExitResolver 只解析 affordance、不直接 ending**：`resolve_exit_affordance(text) → ExitDecision`
  （`withdraw_to` / `move_through` / `return_to_site` / `end_campaign` / `offer` / `no_exit`）。
  目標以 **role** 表示（`target_role=safe_zone / site`），**不硬寫地名**；loop 再解析成實際 area id。
- **唯 `end_campaign` 進 EndingGate**；其餘皆續行。語意不明 → `offer`（ExitOffer 四選一，永遠含「結束」出口）。

---

## 四、主題無關角色（Theme-Agnostic Roles）

area 實體可帶 **roles**，取代硬編碼地名：

```text
safe_zone     撤退/整理的結構性安全區
site          調查現場（主要地點）
entry         起始區域
active_area   玩家撤退前所在、返回時的目標區域
campaign_exit 結束本次調查的離場區域
```

- 查詢：`area_with_role / areas_with_role / safe_zone_id() / site_area_id() / entry_area_id() / is_safe_zone()`。
- `safe_zone_id()` 優先 `role=safe_zone`；無 role 時 fallback 主題無關預設 `area.safe_zone`；舊 `area.outside_dock`
  仍被 `is_safe_zone` 認得（相容）。`withdraw_to_safe_zone` 用 role 找/建安全區，不再硬寫 dock 地名。
- ReviewMode 文案去主題化（「返回{site_label}，繼續調查」），observation 投影 `area_roles`（area_id→roles）。

---

## 五、ExplorationMode / ReviewMode Lock（撤離鎖，`exploration_mode.py`）

玩家撤到安全區整理時，系統必須**停止自動調查推進**。四個持久模式（存 `game_meta.exploration_mode`，跨 beat 黏著）：

```text
active_exploration       正常調查推進（預設）
temporary_retreat        撤到安全區（不結束；停止自動推進）
review_mode              在安全區整理既有線索（只生 notes，不新增 fact/truth/object）
campaign_end_requested   玩家明確結束 → 進 EndingGate
```

`is_review_locked(mode)`（review/retreat）時的「撤離鎖」：
1. **current_area durability**：`_world_model_tick` 不讓 kernel scene-sync 覆蓋安全區（修掉「撤退只撐 1 beat」）。
2. **不推 reveal**：`_revelation_tick` 在 review 模式被 gate 擋下（見 §六）。
3. **不新增 object/fact**：只允許「玩家明確檢查**已知**物件 → inspected」。
4. **decision_point 換成 ReviewMode 四選一**：`return_inside / review_notes / inspect_inventory / end_campaign`；
   `entities_here` 只顯示安全區實體；「根據已知線索整理」只生確定性 `review_notes_text`（不新增 fact/truth）。
5. **narrative consistency**：`_enforce_review_narrative` 偵測敘事冒出未記帳 fact → 退回確定性 notes。
- 只有「回去研究站/返回現場/重新進入」這類**明確再入**（`_REENTER`，去主題化）才解鎖回 `active_exploration`，
  並把 current_area 移回 site/active_area。

---

## 六、WorldConsequence vs TruthEvidence Split（`action_intent.py` + `truth_evidence_gate.py`）

開放探索鐵律：**找路 / 整理 / 引用 NPC fact / 一般檢查 都是「世界後果」，不該推 reveal_progress；
只有「明確的真相調查」或合法 structured evidence_events 才推 reveal。**

- `classify_action(text) → ActionIntent`（七類）：`world_navigation / world_review / object_inspection /
  truth_investigation / npc_fact_query / campaign_end / unknown`；`no_truth_intent`（「不碰真相/只找路/只整理」）。
- `TruthEvidenceGate.evaluate(action_class, exploration_mode, no_truth, has_structured_evidence, truth_bearing)`：
  - allow：`truth_investigation` ／ 合法 structured evidence ／ truth-bearing 物件。
  - block：`no_truth` ／ review/retreat 模式 ／ navigation / review / npc_fact_query / 一般 inspection / unknown。
- `_revelation_tick` 在 apply reveal 前先問 gate；被擋下的 clue 記為 `blocked_reveal_candidates` / unmapped
  （debug），**標已消耗但不更新 reveal ledger**。NPC 結構化 evidence 通道（truth_id 白名單）**不受** action gate 影響。

> **回歸保證（contract regression，真 LLM C1–C7 ALL PASS）**：navigation/review/npc_fact_query/一般 inspection
> 不推 reveal；truth_investigation 與 structured evidence 仍可推；review 不產生未記帳 fact。

---

## 七、observation 投影（QA / AI 在遊玩中就抓到問題）

`world_progress`（WorldModel 投影）：`current_area / known_areas / exits / area_roles /
investigation_state(=exploration_mode) / available_next / changed_entities_this_beat /
world_facts / world_model{counts, entities, entities_here, interactables_here, affordances_here}`。

step 回傳的 split debug：`action_class / no_truth_intent / reveal_gate_allowed /
reveal_gate_block_reason / blocked_reveal_candidates`。

空間投影 `spatial_debug` 與玩家面 `spatial_summary` 見 `17-spatial-projection.md`。

---

## 八、邊界（本層**不做**）

不擴 scene graph（不新增 area/exit 圖節點，只在 WorldModel 登記實體）·不做 spatial reasoning / 幾何 /
pathfinding ·不改 reveal ladder（只在它前面加 TruthEvidenceGate）·不收斂 world_facts（扁平通道與 fact 實體並存）·
不加故事內容（主題內容 runtime 由 LLM 給；機制主題無關）。
