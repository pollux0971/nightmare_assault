# CONTRACTS — 契約索引（FROZEN）

> **狀態：FROZEN。** 這是「凍結介面」的**索引**：每個契約 id 指向 canonical 來源，
> 不在此複製規格（單一真相，避免漂移）。每次寫工單時，把 canonical 來源的相關段落貼給子 agent。
> **規則**：要改任何凍結契約，先停下走 §九 變更程序（需使用者同意）。改到 FROZEN 契約即越界。

**Canonical 來源**（規格本體在這些檔，本檔只凍結 id 與歸屬）：
- `nightmare-assault-design-fixed/build/CONTRACTS.md`（§一目錄 / §二 Pydantic / §三 Client / §四 Parser / §五 API / §六 常數）
- `nightmare-assault-design-fixed/07-data-contracts.md`（API、Agent 輸出 schema、串流 repair、event 抽取、injection）
- `nightmare-assault-design-fixed/08-engineering.md`（SQLite schema、版本/patch 並行、錯誤降級、warden fallback）
- `nightmare-assault-design-fixed/01-architecture.md`（§四 權限、§五 SignalBus 事件）

> 契約 id 用 `反引號`，`dev/tools/board.py` 掃描它們，驗證每個工單的 `contracts:` 都指得到。

---

## 一、資料與常數

| 契約 | 內容 | Canonical |
|---|---|---|
| `Models` | 全部 Pydantic 資料類（Option/BeatMeta/DecisionPoint/WardenOutput/Revelation/Interactable/Location/SceneRegistry/NPC*/各 agent 輸出/LedgerFact/LLMResult） | 07 §二、build/CONTRACTS §二 |
| `Constants` | DELIM_CONTINUE/DELIM_DECISION、MODEL_TIERS、SUMMARY_TOKEN_CAP、BEAT_WINDOW_SIZE、NARRATION_ONLY_MAX、CONTEXT_THRESHOLD_L1/L2/L3 | build/CONTRACTS §六 |
| `DecisionPoint` | story DECISION 後的結構化決策點（situation_recap/options/beat_meta/is_narration_only） | 07 §二 |

## 二、狀態與持久化

| 契約 | 內容 | Canonical |
|---|---|---|
| `Blackboard` | 中央狀態容器，agent 只透過它讀寫 | 03 §二、08 §二 |
| `BlackboardPatch` | 版本化 + patch：apply_patch/collect_pending/merge_and_bump；非同步只產 patch、安全點 merge | 08 §二 |
| `PermissionTable` | 誰能寫什麼/碰不到什麼；story 讀不到 real_bible、dreaming 碰不到 secret_core（程式碼強制） | 01 §四 |
| `SceneRegistry` | 當前位置/移動/location_reached/interactable（item/clue/corpse/door/npc_trace） | 07 §二 |
| `SQLiteSchema` | runs/schema_meta/beats/npc_states/inventory_snapshots/chat_logs/save_points/llm_traces + index/unique；beat 快照含「當時」rolling_summary | 08 §一 |
| `RollingSummary` | 滾動摘要（上限）+ 保護清單 | 02 §八、07 §二 |
| `Ledger` | fact ledger 二元組 `LedgerFact(type, content)`；技能 (技能,侷限) | 07 §二 |
| `SharedInventory` | 共用道具庫；item held_by 綁 NPC；不對外暴露 is_key_item | 03 §二 |
| `NPCRegistry` | NPC anchor（secret_core/self_aware）vs evolving（dreaming 可寫） | 07 §二、01 §四 |

## 三、LLM 與解析

| 契約 | 內容 | Canonical |
|---|---|---|
| `OpenRouterClient` | call/stream、fallback 鏈、timeout、每次寫 llm_traces；簽名固定 | build/CONTRACTS §三、01 §三 |
| `SkillCaller` | 讀 SKILL.md→組 prompt→呼 client→Pydantic 驗證 | build/BUILD 工單7 |
| `SkillLoader` | skills/{agent}/SKILL.md 載入 + 熱重載 | 01 §六 |
| `StreamParser` | 逐 token 滑動視窗偵測分隔符、分離 narrative/decision、三級 repair、fallback decision UI | 07 §三、build/CONTRACTS §四 |
| `StreamProtocol` | 分隔符 `<<<CONTINUE>>>`/`<<<DECISION>>>` 字面常數（前後端同字串） | build/CONTRACTS §四六 |
| `EventExtract` | 程式碼層事件抽取（非 agent），與 story 自報取聯集、衝突信程式碼 | 07 §四 |
| `InjectionGuard` | 玩家輸入包 `<player_action>`，永不直拼 prompt | 07 §五 |

## 四、Agent 輸出與裁決

| 契約 | 內容 | Canonical |
|---|---|---|
| `SetupOutput` | real_bible + npc_registry + protagonist + scene_registry + opening_sequence | 07 §二 |
| `OrchestratorOutput` | fragments_to_reveal + reasoning；多數條件程式碼判 | 07 §二、02 §三 |
| `WardenOutput` | rule_violation/ending_triggered/skill_*/directive_to_story | 07 §二 |
| `WardenFallback` | 順序：本地硬規則 → LLM 語義 → 全失敗保守正常推進（不誤殺） | 08 §三 |
| `DreamingOutput` | evolving 更新（只寫 npc_evolving；self_aware=false 不編謊） | 07 §二、01 §四 |
| `CompactorOutput` | compressed_summary + ledger_updates + 保護伏筆 | 07 §二、02 §八 |
| `EndingConditions` | 終局狀態 + gate/前置（MVP-B 一種結局）；結局揭露政策 masked（只露已發現碎片、未發現遮罩標題） | 02、build/BUILD MVP-B、設計 12 §十 |
| `OpeningSeeds` | 序幕真相種子（true/false/imagery/personal/mechanical）；**只把 surface/obligation 餵 story，hidden_truth 留 real_bible**（防暴雷）；開場義務 5 條 + 長度政策 | 設計 12 §一–九（patch 回饋） |

## 五、迴圈與前後端

| 契約 | 內容 | Canonical |
|---|---|---|
| `BeatLoop` | 玩家輸入→warden→orchestrator→story→快照→安全點 merge→compactor 檢查 | build/BUILD 工單15、08 §二 |
| `SignalBus` | 7 事件（BEAT_COMPLETED…CONTEXT_THRESHOLD），發布者/訂閱固定 | 01 §五 |
| `API` | pywebview API 方法名固定（check_config/start_game/submit_decision/get_game_state/…） | 07 §一、build/CONTRACTS §五 |
| `JsApi` | 後端 `window.pywebview.api` 暴露面（API 的前端視角） | 06 §二、07 §一 |
| `NA-events` | 後端推前端事件名固定（NA.appendToken/onContinue/onDecision/onStatus/onError/onBeatComplete/onAudioCue）；**前端不解析分隔符/JSON** | 07 §一、CHECKLIST D4 |

---

## 九、契約變更程序（FROZEN 後）

1. 在對應 canonical 來源提變更，並於本索引標註。
2. 影響盤點：依 01 §四（權限）/§五（事件）/07/08 列出受影響工單（更新其 `contracts`）。
3. **取得使用者同意。**
4. 同步所有受影響工線（各 worktree 更新）。
5. 重標 FROZEN，`dev/journal/` 記一筆。

> 禁止事項（parallel-dev-plan §8）：UI 工線改 Models、agent 工線改 SQLite schema、parser 工線改前端事件名、多人同改契約、story 讀 real_bible。違者即越界。

---

## 十、Narrative Progress Kernel（穩定化補丁，新增）

> 旁路增量層（feature-flag `ENABLE_PROGRESS_KERNEL` 預設 ON）。canonical 規格在
> `nightmare-assault-mvp-a-stabilization-patch/`（src_patch + docs）。核心：world-state 由 kernel 決定，story 只 realize。

| 契約 | 內容 | Canonical |
|---|---|---|
| `ProgressModels` | GameState/EventPatch/EventCandidate/Obligation/LedgerEntry/InventoryItem/NPCPresence/PatchOp/SceneState/ProgressResult（dataclass） | patch src_patch/progress_models.py |
| `PatchValidator` | base_version 檢查、**≥1 progress_delta 否則拒絕**、forbidden、apply ops（scene/clue/inventory/npc） | patch src_patch/patch_validator.py |
| `SceneGraphProvider` | 介面；`StaticOpeningSceneGraphProvider` 讀 `data/opening_scene_graph.json`；kernel 只依賴 provider 標準 graph | 使用者指示、integration-guide |
| `ProgressKernel` | resolve_player_action：intent normalize→candidate select→score→dummy nodes→EventPatch | patch src_patch/progress_kernel.py |
| `ProgressContext` | 最小 story context（committed_event/obligations/forbidden_repeats/new_clues/spawned_npcs + revealed_bible，**無 real_bible**） | patch src_patch/context_builder.py、02-target §6 |
| `ProgressBridge` | GameState ↔ Blackboard 同步（clues/inventory/npc/scene → snapshot+story） | integration-guide §3 BlackboardAdapter |

> 不可破壞既有契約：submit_decision 仍由玩家輸入啟動且 story 仍 stream；snapshot 每 beat（新增 clue/inventory/npc 進 snapshot）。失敗回退舊 BeatLoop 流程。

---

## 十一、配置中心 / Story Agent 模組化（patch v1.1，新增）

> 旁路增量層（feature-flag `ENABLE_CONFIG_CENTER` 預設 OFF→P4 開）。canonical 規格在
> `nightmare-assault-story-doc-patch-batch-v1.1/`（docs + task_cards）。核心：story prompt 由 fragment 組裝、
> 可配置 / 可預覽 / 可快照 / 可回滾；**static prompt 永遠保留 fallback**。對應階段 P（工單 P0–P7）。

| 契約 | 內容 | Canonical |
|---|---|---|
| `ConfigSchema` | additive 配置表：config_profiles/agent_configs/prompt_fragments/prompt_fragment_versions/agent_prompt_bindings/agent_context_policies/feature_flags/run_config_snapshots/prompt_test_cases/prompt_test_results；既有存檔不破壞 | patch docs/04 |
| `PromptFragments` | fragment key 命名 `<agent>.<purpose>`；story 八件（role/objective/kernel_obedience/no_repetition/open_choice/context_policy/output_format/style_horror）；status draft/published/active/archived | patch docs/05、docs/03 |
| `PromptComposer` | fragment→compiled prompt：依 sort_order、只取 enabled、變數代入、**穩定 prompt_hash**、preview 不呼 LLM、缺必填變數報錯 | patch docs/01 §P2、docs/08 |
| `AgentContextPolicy` | 每 agent 可見上下文上限（max_recent_beats/clues/items/npcs）；**story include_real_bible=false（硬規則）** | patch docs/04 |
| `FeatureFlags` | runtime 開關；優先序 `.env > active profile DB > default DB > hardcoded`；ENABLE_CONFIG_CENTER/ENABLE_PROMPT_PREVIEW/ENABLE_RUN_CONFIG_SNAPSHOT… | patch docs/04、docs/08 |
| `ConfigPromptSource` | story prompt 來源優先序：active profile → default profile → static built-in fallback；config 載入失敗 log + 退 static 續行 | patch docs/08 |
| `RunConfigSnapshot` | 每場 run 存 profile/config_json/compiled_prompt_hash/enabled_fragments；載入存檔優先用該 run snapshot（可重現） | patch docs/04、docs/08 |
| `ConfigUI` | 配置可視化：agent config 表 / fragment 編輯器 / prompt preview / flag panel；**draft→preview→activate**，preview 不呼 LLM | patch docs/07、task_cards/P5 |

> 不可破壞既有契約：story 仍 stream、仍由 submit_decision 啟動；**story 永不見 real_bible**（C2）；前端不解析分隔符（D4）；
> ProgressKernel 仍是 world-state 真相（story 只 realize）。config 失敗一律退 static fallback，不崩 MVP-A。

---

## 十二、敘事控制（Narrative Control，patch v0.1，新增）

> 旁路控制層（feature-flag `ENABLE_NARRATIVE_CONTROL` 預設 OFF）。canonical 規格在
> `nightmare-assault-mvp-b-narrative-control-patch-v0.1/`（docs + task_cards + reference_code/examples）。
> 核心：開場只放少量高價值元素、真相分層揭露、Story Agent 不發明世界觀、結局有因果門檻。對應階段 N（工單 NC0–NC7）。

| 契約 | 內容 | Canonical |
|---|---|---|
| `NarrativeContract` | setup 新增的敘事生成契約：protagonist_motive / central_question / motif_palette / opening_budget / reveal_ladder / forbidden_or_limited_motifs / ending_attractors；setup 只產契約與候選元素，不直接寫長開場 | patch docs/03、reference_code/models.py |
| `OpeningBlueprint` | OpeningDirector 從 contract 挑開場元素：allowed_elements（≤ budget，主要新元素 ≤3、新 lore ≤3）/ blocked_elements / opening_reveal_limit（最多 hinted）/ first_choice_policy；至少 1 動機 + 1 可行動線索 | patch docs/04、examples/opening_blueprint |
| `RevealLadder` | 真相揭露階梯 `hidden→hinted→observed→suspected→confirmed→actionable`；不可跳級、需 evidence 才升級、開場最多 hinted；RevealManager 只把 allowed level 內容餵 story | patch docs/05、examples/reveal_ladder |
| `StoryAgentDelta` | Story Agent 降權：只執行 blueprint（allowed/forbidden_new_elements、beat_purpose、truth_reveal_limit、player_motive）；每 beat 只新增一個主要敘事資訊；不發明 blueprint 以外核心設定 | patch docs/09、prompts_reference/story_agent_delta |
| `QualityGate` | 輸出前規則版檢查：元素數量 / 動機是否清楚 / forbidden motifs / reveal jump / 選項是否有意義；違規 → repair prompt 或 fallback | patch docs/06、reference_code/quality_gate.py |
| `EndingGate` | 結局因果門檻：clean_escape 需 explicit_escape_action + exit_location_reached + threat_resolved_or_avoided；**0/8 真相碎片不可 clean_escape**（只可 ambiguous_escape / fail-forward）；masked 只露已確認/接觸過 | patch docs/07、reference_code/ending_gate.py |

> 旁路紀律：`ENABLE_NARRATIVE_CONTROL` 預設 OFF；舊流程仍可 fallback；新欄位皆 optional、不破壞既有存檔；
> 不大重構 beat loop。**story 永不見 real_bible 不變**；refine（非取代）UB6 開場種子與 UB7 masked 結局。

---

## 十三、MVP-C（聊天室 + 離場命運，新增）

> 接通兩個既有但未接迴圈的 agent，讓世界活起來、離場有後果。對應階段 C（工單 MC1–MC5）。
> canonical：`skills/npc-chat/SKILL.md`、`skills/offstage-fate/SKILL.md`、00 §六 決策日誌 B5/B6/離場雙軸。

| 契約 | 內容 | Canonical |
|---|---|---|
| `NPCChat` | 玩家對在場 NPC 多輪對話：認知卡投影（公開面 + evolving + voice_sample，**無 real_bible**）、self_aware 行為（誠實答錯/隱瞞說謊）、職業折射；ChatLog 追加 + SQLite chat_logs（cold）；退出濃縮 3–4 句進 story hot context | skills/npc-chat、00 B5 |
| `OffstageFate` | 離場 NPC 命運：**程式碼加權擲骰**決定 fate_type（opportunity_return/missing/corpse/hostile_return），LLM 寫血肉；領 revelation_pool 碎片（carried_fragment）；寫 npc_registry[].{presence,alignment,offstage_intent,carried_fragment} + corpse interactable；**對玩家完全隱藏，重逢才揭曉**；非同步只產 patch | skills/offstage-fate、OffstageFateOutput、00 離場雙軸 |

> 紀律：npc-chat 結構性防暴雷（同 story，無 real_bible）；offstage-fate 只寫 npc/scene（權限邊界，碰不到 secret_core/world_truth）；離場命運隱藏直到重逢/搜屍才過揭露閘門。失敗皆 graceful（不崩主迴圈）。

---

## 十四、敘事控制 v0.2 · 揭露橋接（Narrative Control patch v0.2，新增）

> 接續 §十二（敘事控制 v0.1）。診斷：開 `ENABLE_NARRATIVE_CONTROL` 後開場收斂了，但**玩家調查無法轉成官方真相進度**（kernel clues / npc-chat hints / real_bible 三層脫節 → 結局永遠 0/X）。
> 核心修補：建一條受控橋接 `Evidence → Reveal Level → RevealedBible → Recap`，並把敘事控制延伸到 npc-chat、加答債/結局表層/逃脫提交/母題冷卻/動機心跳/表層消毒。
> canonical：`nightmare-assault-mvp-b-narrative-control-patch-v0.2/`（docs 00–10 + task_cards P0–P7 + reference_code/examples，**範例不照抄**）。對應階段 R（工單 NR0–NR7）。
> 旁路紀律同 §十二：受 `ENABLE_NARRATIVE_CONTROL`（必要時細分子旗標）控管，預設行為不變；新欄位皆 optional、不破壞既有存檔；**story / npc-chat 永不見 real_bible 不變**。

| 契約 | 內容 | Canonical |
|---|---|---|
| `EvidenceEvent` | 每個對玩家有意義的線索事件：evidence_id / source(kernel\|npc_chat\|story\|inventory\|document) / player_action / surface_text / truth_id / suggested_reveal_level / evidence_strength(0–1) / scene_id / beat_number；裝飾性線索標 `atmosphere_only=true` 不入橋 | docs/02、examples/evidence_mapping、reference_code/revelation_bridge.py |
| `RevelationBridge` | 服務 `process_evidence_events()`：把 kernel 結果 / npc-chat 退出 / story 輸出產生的 EvidenceEvent 過映射 → 呼叫 RevealManager 升級 → 寫 RevealLedger / revealed_bible；不可無 evidence 直跳 confirmed | docs/02、docs/08、reference_code/revelation_bridge.py |
| `RevealLedger` | 揭露階梯帳（延伸 §十二 RevealLadder）：每 truth_id 記 `hidden→hinted→observed→suspected→confirmed→actionable` 當前等級 + 來源 evidence；recap 計**部分進度**（已確認/已懷疑/已觀察/已接觸/仍未知），非只算 confirmed | docs/02 §recap、examples/reveal_ledger |
| `EvidenceMapping` | clue↔truth 映射表（JSON/table）：clue_id / truth_id / minimum_action / grant_level / grant_strength；無映射的玩家線索進 debug「unmapped」告警；無 truth_id 必標 atmosphere_only | docs/02 §mapping、examples/evidence_mapping |
| `NPCChatControl` | npc-chat 套用同一敘事契約：輸入 motif_palette / forbidden_motifs / truth_reveal_ladder / current_reveal_levels / active_player_motive / answer_debt / context_budget；輸出 `NPCChatResponse`(visible_reply / npc_emotion_delta / answer_status(answered\|partial\|evaded\|refused) / evidence_events[] / new_lore_terms[] / used_motifs[] / quality_flags[])；越界新 lore → repair 重寫保留情緒與線索 | docs/03、reference_code/npc_chat_control.py、prompts_reference/npc_chat_delta.md |
| `AnswerDebt` | 重複提問答債：question category(identity/mechanism/threat/location/action) × topic 計 debt（0 無 / 1 可迴避 / 2 須部分答 / 3 須具體線索或具理由拒答）；debt≥2 時下一相關回應**至少**給部分答/具體線索/指向證據/具理由拒答；餵 story 與 npc-chat context | docs/04、reference_code/answer_debt.py |
| `EndingSurface` | 結局**表層變體**獨立於 ending_type：clean_escape / ambiguous_escape / failed_escape / death / truth_locked；由 RevealLedger 決定 clean vs ambiguous（**0 confirmed 真相 → 不可 clean_escape**，render 須呈現不確定）；debug 印 gate 理由 | docs/05、reference_code/ending_gate.py、prompts_reference/ending_renderer_delta.md |
| `EscapeCommitGate` | 兩段式逃脫：`attempt_escape → exit_candidate_found（產出口發現 beat 與選項）→ player commit_escape → EndingGate.evaluate`；首次「我試圖離開」不得即觸發結局，須玩家明確提交才結算 | docs/05 §B、examples/ending_gate_cases |
| `MotifCooldown` | 母題冷卻：逐場景追蹤 used_motifs 次數；超 `max_uses_per_scene` 的下一次使用須**揭露新資訊 / 改變狀態 / 變可行動 / 進冷卻**；QualityGate flag 連續停滯母題（同母題連 3 beat 告警） | docs/06、reference_code/motif_tracker.py |
| `MotiveHeartbeat` | 動機心跳：追蹤距上次提及主角動機的 beat 數；逾 2–3 beat 加「動機提醒」義務，且須透過文件/NPC 反應/道具/選項嵌入，不可重複同一句 | docs/06 §heartbeat、reference_code/motif_tracker.py |
| `SurfaceTextSanitizer` | 表層消毒：story / npc-chat 輸出 render 前掃非故事洩漏（technical/protocol/COLLECT/inst、prompt artifact、壞 markdown fence、敘事內重複分隔符）；先安全的決定性替換，否則短 repair prompt（不改劇情/選項/evidence/JSON）；契約授權的 in-world 詞（432.7、17Hz、明確 in-world 的 protocol）為例外，不全域刪英文 | docs/07、reference_code/surface_sanitizer.py |

> 驗收主線（docs/10 + checklist）：玩家檢查警示紙條、問 NPC 可疑頻率、查文件後，**結局 recap 不得 0/X**（即使只是 hinted/suspected 也要顯示部分發現）；0 confirmed 逃脫渲染為 ambiguous_escape；「我試圖離開」先出口候選不即結局；重複提問須付答債；母題不停滯；表層無洩漏。
> 不做（non-goals）：不加新 lore / 新世界機制 / 更大故事圖——本補丁是「把線接通」而非「加內容」。

---

## 十五、Runtime Hard-Gate（v0.3.1，新增）

> 接續 §十四。把已存在但「半接」的 Narrative Control 模組從 monitor 升級成真正 runtime **hard-gate**（pass/reject/repair once/fallback，不只是 log）。
> 原則（APPLY_ORDER）：不加 story/lore/agent；先修安全線（結局/反劇透）再修體驗線；每 gate 都有行為；每批跑回歸 + agent_play 冒煙。
> canonical：`nightmare-assault-runtime-hardgate-patch-v0.3.1/`（docs 00–09 + task_cards + reference_code/tests_reference，**範例不照抄**）。對應階段 H（工單 HA1–HE1）。
> 旁路紀律同 §十二/十四：受 `ENABLE_NARRATIVE_CONTROL` 控管，預設行為不變；新欄位 optional、不破壞既有存檔；**story/npc 永不見 real_bible 不變**。

| 契約 | 內容 | Canonical |
|---|---|---|
| `EndingObservationInvariant` | beat observation 不變式：**`ended=true` ⇒ options=[] 且 free_input_hint=null**；`BeatObservation.enforce_invariants()` 在輸出前強制；regression 斷言 `not (ended and options)` | docs/02/03、reference_code/models.py |
| `EndingCausalityGate` | 結局因果硬閘：所有結局先成 `EndingCandidate`（type/source/confidence/cause_event_id/requires_commit）；**`death_physical` 必須有 warden hard_trigger==death 或 progress.death_cause_event 或明確致命行為**——`danger_level` 達標只能降級 danger_warning/injury/failed_escape（**禁止 danger→death**）；escape 須 commit step，0 confirmed→ambiguous_escape | docs/03、reference_code/ending_gate.py、tests_reference/test_death_gate_nr.py |
| `HiddenRecapMask` | 玩家可見輸出（observation / API / agent_play 預設）**不得含 hidden truth content**：recap 只給 `found`(露已達等級者) + `hidden_count` + `hidden_titles`（遮罩標題）；full hidden recap 僅 explicit debug flag（`--debug-reveal-truth`）可開 | docs/07、task_cards/P0_A3 |
| `StoryEvidenceExtractor` | 玩家調查但 reveal 無變化時的保底：`is_investigation(action)`（檢查/查看/翻找/詢問…）+ `has_concrete_info(narrative)`（頻率/編號/紀錄/名單…）+ `reveal_changed==False` → 產 fallback hinted EvidenceEvent；map 不到 truth_id → `source=fallback, truth_id=null` 並計 `unmapped_evidence_this_beat` | docs/04、reference_code/evidence_extractor.py、tests_reference/test_story_evidence_extraction.py |
| `EvidenceEvent`（擴充） | §十四 EvidenceEvent 加 `reveal_level / player_facing / debug_reason / source∈{kernel,story,npc_chat,document,item,fallback}`；所有來源統一這個型別走 `RevelationBridge.apply` | docs/04、reference_code/models.py |
| `NPCChatResponse`（結構化） | run_npc_chat 改回結構化：`reply / answer_status(none\|evasion\|partial\|actionable\|confirmed) / evidence_events / new_lore_terms / used_truth_ids / blocked_or_uncertain_claims`；經 `NPCChatControlGate.validate`（new lore/reveal ceiling/answer debt/chat_available）→ invalid repair once → 仍 invalid 用 `safe_fallback_reply`（actionable 方向、不新增 lore）；evidence 經 cap_to_ceiling → bridge | docs/05、reference_code/npc_chat_gate.py |
| `StoryRepairPipeline` | QualityGate 從 monitor 變 gate：`check→pass 收` / `fail→repair_once(repair_instruction)→再 check` / `仍 fail→deterministic_fallback`（系統正確的方向 + 安全選項，不求文采）；**最多 repair 一次**；log `repaired/fallback` | docs/06、reference_code/quality_repair.py |
| `SurfaceSanitizer`（擴充） | 覆蓋**所有玩家面輸出**（narrative/options/situation_recap/NPC reply/ending rendered_text/observation JSON）；新增 Unix path（/usr/var/home/data/tmp/etc）/ IP / 權限/存取權/系統提示 / core；嚴重污染觸發 repair 而非只刪詞；in-world 授權詞例外 | docs/07、reference_code/surface_sanitizer.py、tests_reference/test_surface_sanitizer_runtime.py |
| `ObservationDebug` | observation 加 `debug`：committed_event/progress_delta/escape_step/evidence_events_this_beat/unmapped_evidence_this_beat/reveal_updates_this_beat/quality_gate/model_used；NPC 分層 `visible_npcs/known_npcs/chat_available_npcs`（取代模糊 present_npcs）；**debug 可顯示 truth_id 但不得顯示 hidden content** | docs/08、examples/observation_debug |

> 硬禁止（APPLY_ORDER §強制禁止）：danger 當直接死亡條件 ✗；普通 beat 同時帶 ending+options ✗；observation 輸出 full hidden truth ✗；QualityGate 只 log 就放行 ✗；NPC-chat 繞過 Reveal Ladder ✗。
> beat pipeline 短期不大重構（docs/02）：在 `orchestrator_loop._step_kernel` 內整理 private method，等 MVP-B 穩定再抽 BeatPipeline class。
