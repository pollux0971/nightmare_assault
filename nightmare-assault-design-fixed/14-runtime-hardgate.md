# 14 · Runtime Hard-Gate（v0.3.1）

> **實作對應**：開發工單 `HA1`–`HE1`（dev 階段 H）。**狀態：9/9 工單 done、已測（714 passed，flag OFF/ON 各一次）**（接續 §十四 / 階段 R）。
> **canonical 來源**：`nightmare-assault-runtime-hardgate-patch-v0.3.1/`（docs 00–09 + task_cards + reference_code/tests_reference）。
> **一句話**：把已存在但「半接」的 Narrative Control 模組，從只會 log 的 monitor 升級成真正有行為（pass / reject / repair once / fallback）的 runtime hard-gate。

## 〇、為何有這個補丁

NR 階段（v0.2）把線接通了，但實機 selfplay 暴露幾個「半接」：NPC evidence 靠事後掃字（hit-or-miss）、QualityGate 只 log 不擋、結局可能同時帶 options、danger 達標可能直接判死、observation 可能洩 hidden truth、調查偶爾 0 reveal 反應。本補丁把這些補成硬閘。**不加 story/lore/agent**。

## 一、Batch A — 結局 / 反劇透安全（最高優先）

- **HA1 Ending Observation Invariant**：`ended=true ⇒ options=[]、free_input_hint=null`。`BeatObservation.enforce_invariants()` 輸出前強制；regression 斷言 `not (ended and options)`。
- **HA2 Death Causality Guard**：所有結局先成 `EndingCandidate`；`death_physical` 必須有 warden hard_trigger==death / progress.death_cause_event / 明確致命行為——**danger_level 達標只能降級** danger_warning / injury / failed_escape，禁止 danger→death。
- **HA3 Hidden Recap Masking**：玩家可見輸出不得含 hidden truth content；recap 只給 `found + hidden_count + hidden_titles`（遮罩）；full recap 僅 `--debug-reveal-truth` 可開。

## 二、Batch B — 調查 → Evidence → Reveal

- **HB1 Story Evidence Extraction**：調查型 action + 具體 narrative + 本 beat reveal 無變化 → 保底 hinted `EvidenceEvent`；map 不到 truth_id → `source=fallback`、計 `unmapped_evidence_this_beat`。
- **HB2 RevelationBridge Unified Inputs**：kernel / story / player-action / npc / document evidence 全走同一 `RevelationBridge.apply`；ledger 更新寫 revealed_bible；step 暴露 `reveal_updates_this_beat / evidence_events_this_beat / unmapped_evidence_this_beat`。

## 三、Batch C — NPC-chat Runtime Gate

- **HC1 NPCChat Structured Gate**：run_npc_chat 回**結構化** `NPCChatResponse`（reply / answer_status / evidence_events / new_lore_terms / used_truth_ids / blocked_or_uncertain_claims）→ `NPCChatControlGate.validate` → 違規 repair once → 仍違規 `safe_fallback_reply`（actionable 方向、不新增 lore）；evidence 經 cap_to_ceiling → bridge（**不再只靠 keyword scan**，修 NR1 軟點）。

## 四、Batch D — 品質 / 表層 repair pipeline

- **HD1 QualityGate Repair Once**：`check → pass 收 / fail → repair_once → 仍 fail → deterministic_fallback`（系統正確的方向 + 安全選項，不求文采）；**最多一次 repair**；log repaired/fallback。
- **HD2 Surface Sanitizer All Outputs**：消毒覆蓋 narrative / options / situation_recap / NPC reply / ending rendered_text / observation JSON；新增 Unix path / IP / 權限/存取權/系統提示 / core；嚴重污染觸發 repair 而非只刪詞。

## 五、Batch E — 可觀測性 / 回歸

- **HE1 Observation Debug Fields**：observation 加 `debug`（committed_event / progress_delta / escape_step / evidence_events_this_beat / unmapped_evidence_this_beat / reveal_updates_this_beat / quality_gate / model_used）；NPC 分層 `visible_npcs / known_npcs / chat_available_npcs`；debug 可顯示 truth_id 但**不得 hidden content**。

## 六、建議 beat pipeline（docs/02，短期不大重構）

```text
interpret action → warden → kernel.resolve → pre-ending precheck（致命/即結局 → render，options=[]）
→ run_story（+ QualityGate repair once）→ EvidenceExtractor → RevealBridge.apply
→ post-ending postcheck（commit escape 等 → render，options=[]）→ build observation
```
不變式：`assert not (observation.ended and observation.options)`。短期在 `_step_kernel` 內整理 private method，MVP-B 穩定後再抽 `BeatPipeline` class。

## 七、硬禁止

danger 當直接死亡 ✗｜普通 beat 同帶 ending+options ✗｜observation 輸出 full hidden truth ✗｜QualityGate 只 log 放行 ✗｜NPC-chat 繞過 Reveal Ladder ✗。

> 契約索引見 `dev/CONTRACTS.md §十五`；開發階段見 `dev/stages/stage-H.md`。
