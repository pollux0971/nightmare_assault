# 09 · 本次修訂重點

本檔記錄從 `nightmare-assault-design-new` 到本版的契約統一修正，方便之後交給 AI 實作時確認沒有遺漏。

## 已修正

1. **SetupOutput 與 setup skill 對齊**  
   `skills/setup/SKILL.md` 已補 `scene_registry`，並要求起始場景、場景 id、exits、interactables 與 `revelation_pool` 的位置條件對齊。

2. **MVP-A / MVP-B 範圍對齊**  
   MVP-A 只要求核心循環穩定跑 30 beat、防暴雷、JSON repair、存讀檔與 compactor；至少一種結局、完整道具展示、技能封頂強化與 NPC 輕量 evolving 放到 MVP-B。

3. **串流解析單一責任**  
   Python 後端 `StreamParser` 是唯一解析者。前端 JS 不解析 `<<<CONTINUE>>>` / `<<<DECISION>>>`，也不做 JSON repair，只接 `NA.appendToken/onContinue/onDecision/onStatus/onError`。

4. **前後端狀態同步**  
   API 補 `get_game_state()`；後端推 `NA.onStatus(status)` 與 `NA.onError(error)`，避免連點、重入與 UI 狀態不同步。

5. **Pydantic 工程化**  
   list/dict/model 預設值改採 `Field(default_factory=...)`，並把 `ledger_updates` 從 tuple 改成 `LedgerFact` model。

6. **SQLite 補強**  
   新增 `runs`、`schema_meta`，並補上常用 index / unique constraints，讓 list_saves、回溯、debug 與 schema migration 更穩。

7. **Warden fallback 補硬規則優先**  
   流程固定為：本地 deterministic hard rule → LLM semantic judgment → 若未命中硬規則且 LLM 全失敗，才正常推進。

8. **Build plan 補工單 0**  
   開工前先跑「契約統一與文件修補」檢查，確認 setup、MVP、串流、API、Pydantic、SQLite、warden fallback 都沒有互相矛盾。

## 開工建議

先執行 `build/BUILD-PLAN.md` 的「工單 0」。通過後再從工單 1 寫 `core/models.py`。不要直接跳到 story agent 或前端，否則很容易被資料契約不一致拖慢。

---

## 後續補丁里程碑（MVP-A 落地後新增）

MVP-A（U00–U20）穩定後，又疊了兩個 additive 補丁。兩者皆為旁路增量層、預設不破壞既有行為、失敗一律降級回既有流程。詳見專門設計文件。

### 補丁一 · Narrative Progress Kernel（穩定化）— 見 `10-progress-kernel.md`

- **問題**：LLM 自由決定世界狀態 → 開門後又問開門、NPC 不進場、線索/背包不持久、行動沒後果。
- **解法**：把「推進判定」交給程式碼層 `ProgressKernel` + `PatchValidator`；**story 退化成純 realizer**。每 beat 強制 ≥1 progress_delta（狀態推進，非固定劇情）。結局改用 **attractor**（累積拉力越門檻），非固定終點。
- **旁路**：`ENABLE_PROGRESS_KERNEL` 預設 ON；kernel/圖/驗證失敗 → 回退 legacy 流程，不 crash。
- **對應開發階段**：S（`SK01`–`SK13`）。

### 補丁二 · Config Center / Story Agent 模組化 — 見 `11-config-center.md`

- **問題**：story prompt 是寫死的 SKILL.md，難配置、難預覽、難回滾、難 debug「哪個 prompt 版本造成這行為」。
- **解法**：prompt 拆成可配置 fragment；`PromptComposer`（決定性 + 穩定 prompt_hash + 零 LLM preview）→ `ConfigPromptSource`（active → default → **static fallback**）；配置 UI 走 draft→preview→activate；每場 run 存 config 快照。
- **旁路**：`ENABLE_CONFIG_CENTER` 預設 OFF（行為與補丁前一致）；config 失敗一律退 `skills/story/SKILL.md`，不崩 MVP-A。
- **擴展**：同機制套到 warden/orchestrator/compactor/setup；**story `include_real_bible=false` 硬規則不變**。
- **對應開發階段**：P（`P0`–`P7`）。

### 補丁三 · MVP-B 差異化機制 — 見 `12-mvp-b.md`

- **內容**：在穩定的 MVP-A 引擎上補齊五項差異化機制。
- **UB1 技能宣稱封頂強化**：warden 本地偵測破格宣稱 → 接受但加具體、接劇情的侷限（謎題化）；(技能,侷限) ledger；UI 提示。
- **UB2 一種結局序列**：結局觸發 → 純敘述收尾 + 完整真相揭露（結局唯一可露 real_bible）+ 復盤（提早觸發→沒找到的碎片一次攤開）→ 回選單。
- **UB4 輕量 dreaming**：在場 active NPC 每 5 beat 演化；只寫 evolving（C6）、self_aware=false 不編謊（C5）、沒戲份凍結（C7）、非同步只產 patch。
- **UB5 道具庫完整**：held_by（綁 NPC）/ is_key_item（不外露）/ 增刪查轉移 / 隨快照保存。
- **UB3 雙層防暴雷驗證報告**：對抗性 caller + injection probes → 證明 story context 結構上無 real_bible；產 `dev/reports/antispoiler-report.md`。
- **對應開發階段**：B（`UB1`–`UB5`，另含 UB6 序幕鉤子 / UB7 masked 結局）。

### 補丁四 · 敘事控制 v0.1（Narrative Control）— 見 `dev/CONTRACTS.md §十二`

- **問題**：開場一次塞太多元素（簡訊 + 刻字 + 亂碼 + 銘牌…），玩家抓不到「核心問題」；真相一次攤平或永不揭；Story Agent 會自行發明世界觀。
- **解法**：旁路敘事控制層——`NarrativeContract`（setup 產生）+ `OpeningDirector`（開場 ≤budget 元素、≥1 動機 + 1 可行動線索）+ Story Agent Downgrade（只執行 blueprint）+ `RevealLadder`（真相分層不跳級）+ `QualityGate` + `EndingGate`（0/8 不可 clean escape）。
- **旁路**：`ENABLE_NARRATIVE_CONTROL` 預設 OFF；flag OFF 退回現況。
- **對應開發階段**：N（`NC0`–`NC7`）。

### 補丁五 · MVP-C（聊天室 + 離場命運）— 見 `dev/CONTRACTS.md §十三`

- **內容**：接通兩個既有但未接迴圈的 agent，讓世界活起來、離場有後果。
- **NPCChat**：對在場 NPC 多輪對話（認知卡投影、結構性無 real_bible、職業折射）；退出濃縮 3–4 句進 story hot context；chat_logs 持久化。
- **OffstageFate**：離場 NPC 命運由**程式碼加權擲骰**決定型別（opportunity_return/missing/corpse/hostile_return），LLM 寫血肉；領 revelation_pool 碎片；對玩家隱藏直到重逢/搜屍才揭曉；只寫 npc/scene（碰不到 secret_core）。
- **對應開發階段**：C（`MC1`–`MC5`）。

### 補丁六 · 敘事控制 v0.2（揭露橋接 Revelation Bridge）— 見 `13-narrative-control-v2.md`

- **問題**：v0.1 開場收斂了，但 selfplay 暴露更底層斷鏈——**玩家調查無法轉成官方真相進度**（kernel clues / npc-chat hints / real_bible 三層脫節 → 結局永遠 `0/X`）；另含 npc-chat 失控擴張、重複提問被敷衍、模糊逃脫渲染同乾淨逃脫、母題停滯、開場後動機淡掉、表層洩漏。
- **解法**：受控橋接 `Evidence → Reveal Level → RevealedBible → Recap`（`RevelationBridge` / `EvidenceEvent` / `RevealLedger`）；npc-chat 收同一敘事契約（`NPCChatControl`）；`AnswerDebt`；結局表層變體 `EndingSurface` + 兩段式 `EscapeCommitGate`；`MotifCooldown` / `MotiveHeartbeat`；`SurfaceTextSanitizer`。**Non-goals：不加新內容，只把線接通。**
- **旁路**：受 `ENABLE_NARRATIVE_CONTROL` 控管，預設行為不變；story/npc 永不見 real_bible 不變。
- **對應開發階段**：R（`NR0`–`NR7`）。**狀態：8/8 工單 done、已測（679 passed）。** 整合驗收：透過真實 BeatLoop 證實「調查→揭露帳本前進→結局 recap 不再 0/X」，flag OFF 行為不變。

### 補丁七 · Runtime Hard-Gate（v0.3.1）— 見 `14-runtime-hardgate.md`

- **問題**：NR（v0.2）把線接通了，但實機暴露「半接」——NPC evidence 靠事後掃字、QualityGate 只 log 不擋、結局可能同時帶 options、danger 達標可能直接判死、observation 可能洩 hidden truth、調查偶爾 0 reveal。
- **解法**：把各 gate 從 monitor 升級成 hard-gate（pass/reject/repair once/fallback）。`EndingObservationInvariant`（ended⇒無 options）、`EndingCausalityGate`（danger≠death）、`HiddenRecapMask`、`StoryEvidenceExtractor`（調查保底 evidence）、統一 `RevelationBridge` 多來源、結構化 `NPCChatResponse` + Gate、`StoryRepairPipeline`（repair once）、`SurfaceSanitizer` 全輸出覆蓋、`ObservationDebug`。
- **旁路**：受 `ENABLE_NARRATIVE_CONTROL` 控管，預設行為不變；不加 story/lore/agent。
- **對應開發階段**：H（`HA1`–`HE1`）。**狀態：9/9 done、已測（714 passed，flag OFF/ON 各一次，board 79 工單 0 errors）。**

### 補丁八 · Player Sovereignty / 開放式探索（方向轉變）— 見 `15-player-sovereignty.md`

- **問題（定位）**：把遊戲當「分支劇情」在修，但它其實是**開放式恐怖探索**。系統假設「故事要收束」→ 逃不出去、罐頭 fallback、強制 ending、0/7 像失敗畫面，全是同一個錯誤前提的症狀。
- **核心轉變**：驗收標準從「玩家有沒有推進主線真相？」改成「**玩家做的事，有沒有留下可檢查的世界後果？**」；結局只由玩家明確確認或不可逆後果觸發，不由吸引子自動收束。
- **P0 解法（已落地）**：`ExitResolver`（withdraw 四態 + ExitOffer，唯 campaign_end 進結局）、`NegativeIntentGuard`（「不結束」「不進 B 區」explicit 否定優先）、`WorldStateFact`（NPC/story 有用資訊→可檢查事實，不必是 truth）、`WorldProgress` 觀測（current_area / world_facts / investigation_state / had_consequence + 內建斷言）、0/X 低資訊結局。
- **下一層（已落地，見補丁九/十）**：`WorldModel`——抽象的物件/實體機制（Entity + 狀態機 + affordance），把移動/物件/出口/事實/NPC/否定/撤退收成同一機制的投影，取代扁平 world_facts 與字串比對否定。
- **狀態**：P0 已測（740 passed，flag OFF/ON 各一次）；不加故事內容、不擴場景圖、不動 NPC truth plumbing。

### 補丁九 · WorldModel — 抽象實體機制（開放式探索落地）— 見 `16-worldmodel.md`

- **問題**：補丁八 P0 把方向定了，但「世界記不住物件、退到外面沒真的移動、NPC 講的話蒸發、玩家只是找路卻被推進真相」是同一個缺失機制——沒有「世界裡可被指涉的東西」的模型。
- **解法**：建主題無關實體層 `core/world/model.py`（`Entity{kind=area/exit/object/actor/fact, state 機, affords, roles}`）。**玩家行動＝對某 Entity 套 affordance**，套了必改某實體 state ⇒「有後果」變結構保證。逐層落地：
  - **current_area / exit ownership**：WorldModel 成唯一權威；`register_exit/move_through`（locked/blocked/used 不可通行）；ExitResolver 只解析 affordance（`withdraw_to/move_through/return_to_site/end_campaign`，目標用 **role** 不硬寫地名），唯 `end_campaign` 進 EndingGate。
  - **structured entity_delta**：story 吐 object/actor/fact（≤3、malformed 容錯）；NPC 只准 fact/actor 且走獨立通道（source/confidence/origin=npc，**不碰 reveal**）；散文抽取為 fallback（story extractor / `npc_prose_facts`）。
  - **主題無關角色**：safe_zone/site/entry/active_area/campaign_exit；`safe_zone_id()` 優先 role、fallback 舊常數；ReviewMode 文案去主題化。
  - **ReviewMode Lock（撤離鎖）**：撤到安全區 → current_area durability（不被 kernel 推回）、不推 reveal、不新增 object/fact、dp 換 review 四選一、敘事一致性 fallback；唯明確再入解鎖。
  - **WorldConsequence vs TruthEvidence Split**：`ActionIntent` 七類 + `TruthEvidenceGate`——只有 `truth_investigation` 或合法 structured evidence 才推 reveal；找路/整理/引用 NPC fact/一般檢查/`no_truth` 一律 block（記 blocked_reveal_candidates，不更新 ledger）。
- **prompt 錨定修正**：skill 範例硬寫 `WU 袖扣`/`通訊設備在機房` 被 LLM 照抄 → 每局同物件；改占位符＋禁抄指示，複測四局四樣。
- **旁路**：受 `ENABLE_NARRATIVE_CONTROL` 控管；flag OFF 行為不變；story/NPC 永不見 real_bible 不變。**不擴場景圖、不收斂 world_facts、不加故事內容。**
- **回歸**：真 LLM **contract regression C1–C7 ALL PASS**（navigation/review/npc_fact/inspection 不推 reveal；truth_investigation 與 structured evidence 仍推；review 不產生未記帳 fact）。

### 補丁十 · Spatial WorldModel Projection + UX 摘要 — 見 `17-spatial-projection.md`

- **問題**：WorldModel 有了狀態，但 AI/玩家看不到「我在哪、能走去哪、哪些東西在眼前 vs 知道但不在眼前」；又不能每 beat 叫 LLM 畫地圖（慢、幻覺）。
- **解法（P0–P4）**：`core/world/spatial.py`——`build_spatial_projection(world)` 確定性**唯讀**投影（current_area/routes/blocked/safe_retreat/visible/known_remote/mental_map_text）；WorldModel `version_snapshot()` + `SpatialProjectionCache`（dirty-version，不變不重算）；確定性 `mental_map_text` 模板（選用 `MentalMapWorker` daemon，遊戲迴圈永不 await）；觀測預算（上限 + truncated + counts）。
- **Spatial UX**：`player_facing_spatial_summary` 由投影確定性生成玩家/QA 面板摘要（目前位置/可走/被阻/退路/眼前可互動物/已知但不在眼前），observation 加 `spatial_summary(+_truncated/_source=deterministic_projection)`。
- **UX selfplay 評估**：定位 6/6、零 phantom、visible↔remote 分流完美、簡短不重複；**缺口**：routes/blocked 全空（kernel 場景轉移尚未登記 exit 實體）——建議下一輪補 exit 登記，比進 P5 更值得。
- **邊界**：不呼叫 LLM（同步路徑）、不 mutate WorldModel、不擴場景圖、不做幾何/pathfinding、**不改 reveal gate**、**不接 P5**（不餵 story/review context）。
- **狀態**：已測（spatial 20 + summary 11；**全測 847 passed，flag OFF/ON 各一次**）；二次 adversarial review 後修 3 點（tag 只認新登記 / safe_retreat 排除鎖死 / label 經 setter bump cache）。

### 補丁十一 · NPC Onboarding + Beat Rendering（v0.7 Patch B P0–P3）— 見 `18-player-surface.md` §一/§二

- **問題**：NPC 來歷不明、像資訊端點；beat 隨系統變多而越來越短、像狀態日誌。
- **解法（P0–P3，P4 未開）**：`actor_profile.py`（intro_state + 個性語氣，只取公開面）+ first-contact gate（首次回應須含位置/動作/身分/態度/部分答案，不推 reveal）；`beat_rendering.py`（beat_type 分類 + 軟字數預算 + too_short/short_streak 量測，**只量測不修復**）。
- **P4 gate**：連續 2 個一般 beat 過短才開 repair；**實測校準後 max_short_streak=0 → P4 暫不需要**。
- **Step 4 UX 驗證暴露並修 3 bug**：①裸「研究」誤中「研究站/研究員」→ 導航被當 truth_investigation（改 `_has_research_verb`）②「不碰真相」誤放 review 觸發詞 → 玩家找路時被鎖進 review（移除，它只是 per-beat no_truth）③beat 預算下限過高（350 vs 真 LLM ~200）→ 合理段落被誤標 too_short（校準）。
- **邊界**：不改 TruthEvidenceGate / WorldModel / story；NPC onboarding 不推 reveal；個性只改語氣不改權限。**狀態：6/6 UX 重點通過，全測 878 passed。**

### 補丁十二 · Config Center GUI Completion + WorldDelta fix（v0.8 / v0.8.1）

- **Config Center GUI（v0.8）**：把配置中心補成完整 **Agent Configuration Center**——**只補前端 + API glue，不改 runtime**。Agent Models 表（每 agent primary/fallbacks/temp/max_tokens/enabled，**絕不顯示 api_key**，可 Save + 每行 Test）、Prompt Blocks 表（含 disabled，enabled 開關 / Edit / Save Draft / Activate；sort_order/Rollback 後端無 API → disabled）、Compiled Preview by agent/profile（**零 LLM**）、Test Prompt（後端未支援 → disabled 標示）、dirty state + 未存提示 + toast。
- **WorldDelta fix（v0.8.1）**：LLM/entity_delta 常吐 `id` 而非 `entity_id` → apply 路徑丟失 id（登成 slug）/ 偶發 TypeError。新增 `normalize_entity_delta_dict`（id→entity_id，內部一律 entity_id；id≠entity_id 衝突 → 拒絕該 delta + warning；malformed 丟棄），接在 coerce / apply / apply_entity_delta；**不改 WorldDelta 欄位名**。
- **邊界**：不改 WorldModel 權威 / TruthEvidenceGate / story。**狀態：Config Center 9 條 + WorldDelta 5 條測試；全測 878 passed。**

### 補丁十三 · Player State Surface + Entity Alias / Focus Resolver（Step 5/6）— 見 `18-player-surface.md` §三/§四

- **問題**：玩家看不到「身上有什麼 / 知道什麼 / 焦點在哪」；無法自然指涉「那枚 / 剛才那個 / 他說的地方」。
- **Player State（Step 5，observation-only）**：`player_state.py`——inventory_entities（taken/carried；spatial visible 排除）、known_facts（結構化 label/source/confidence/tags，非粗 key）、current_focus/recent_entities（互動優先序，上限 8）、changed_entities 補 reason、player_state_summary（確定性）。
- **Alias Resolver（Step 6，唯讀）**：`alias_resolver.py`——label normalization + aliases（不建新 entity、衝突不合併）+ `resolve_entity_reference`（解析順序 explicit→label/alias→fact→npc→focus→recent→visible→inventory→unresolved；平手 ambiguous **不亂選**）；observation 加 `entity_resolution`。
- **邊界**：不改 WorldModel 權威 / TruthEvidenceGate / spatial projection；不推 reveal、不新增 fact、無 embedding/LLM、不新增故事內容。
- **狀態**：Player State 10 條 + Alias 11 條測試；**全測 899 passed**；**真 LLM 整合驗收 10 點通過**（檢查物件登記 / 「那枚」重訪 / NPC onboarding / fact→known_facts 不推 reveal / 引用 NPC fact 找路不推 / 撤退 review / 摘要正常 / truth_investigation 仍推 reveal Δ=4 / focus·recent·alias 正確 / GUI 欄位齊全）。

> **里程碑**：開放式探索核心完整——WorldModel 記得世界 · SpatialSummary 告訴你在哪 · PlayerState 告訴你有什麼/知道什麼 · AliasResolver 讓你自然指涉剛才的東西。
