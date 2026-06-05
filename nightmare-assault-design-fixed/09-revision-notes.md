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
- **下一層（規劃中）**：`WorldModel`——抽象的物件/實體機制（Entity + 狀態機 + affordance），把移動/物件/出口/事實/NPC/否定/撤退收成同一機制的投影，取代扁平 world_facts 與字串比對否定。
- **狀態**：P0 已測（740 passed，flag OFF/ON 各一次）；不加故事內容、不擴場景圖、不動 NPC truth plumbing。
