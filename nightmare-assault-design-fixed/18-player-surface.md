# 18 · 玩家體驗投影層 — NPC Onboarding / Beat Rendering / Player State / Alias Resolver

> **承接**：`16-worldmodel.md`（世界記憶）+ `17-spatial-projection.md`（空間投影）。本檔記錄**玩家面體驗層**——
> 讓 NPC 像人、beat 像場景、玩家看得到自己有什麼/知道什麼、能自然指涉「剛才那個」。
> **實作對應**：`core/world/actor_profile.py` · `core/narrative/beat_rendering.py` · `core/world/player_state.py` ·
> `core/world/alias_resolver.py` + loop observation。
> **狀態**：全部落地、已測（**899 passed，flag OFF/ON 各一次**）；真 LLM 整合驗收 10 點通過。
> **一句話**：**WorldModel 記得世界；這一層讓玩家「看得到、認得出、指涉得到」世界。**
> **共同邊界**：皆 observation/體驗層——不改 WorldModel 權威、不改 TruthEvidenceGate、不推 reveal、不加故事內容。

---

## 一、NPC Onboarding（v0.7 Patch B P1/P2）— `actor_profile.py`

NPC 不該像「資訊端點」憑空丟資料；首次接觸要先自然帶出表層身分。

- **ActorProfile**：`intro_state(unintroduced→introduced→known)` / `display_label` / `true_name`（內部，未 known 不顯示）/
  `aliases` / `known_role` / `surface_motive` / `personality_description` / `speech_style`。
  `profile_from_npc` **只取公開面**（絕不放 secret_core）；存 `game_meta.npc_profiles`（npc-chat 可讀）。
- **First-contact gate**：`intro_state=unintroduced` 時，npc-chat context 注入首次接觸要求——回應須含
  **位置/動作 + 表層身分 + 態度 + 部分答案**（不推 reveal、不建 area/exit）；成功後 → `introduced`。
- **Personality**：`personality_description`/`speech_style`（由 setup 的 personality 映射）塑造**語氣**，
  **不改 gate 權限**（不繞 reveal、不把主張說成真相）。
- **驗收**：真 LLM 下首次回應自然帶出「我是X、負責Y、(動作/態度)、部分答案」；mysterious 留白反問 vs nervous 短促重複。

## 二、Beat Rendering（v0.7 Patch B P3）— `beat_rendering.py`

防止 beat 萎縮成狀態日誌。**只量測、不修復（P4 soft repair 預設未開）。**

- `classify_beat_type`（opening/new_area_entry/npc_first_intro/normal_exploration/object_inspection/
  npc_chat/review_mode/danger_beat/ending/system_feedback）+ 軟字數預算 `BUDGETS`（min 已校準到真 LLM 實測）。
- `evaluate_beat_rendering` → `{beat_type, target_min/max, actual_chars, too_short, short_streak, repair_attempted=False}`，
  observation 曝 `beat_rendering`。review/system 不計入「一般 beat 連續過短」。
- **P4 gate**：連續 2 個一般 beat 過短才需 repair；實測校準後 `max_short_streak=0` → **P4 暫不需要**。

> ⚠️ Step 4 驗證暴露並修掉 3 bug：裸「研究」誤中「研究站」、「不碰真相」誤鎖 review、beat 預算過高（見補丁十一）。

## 三、Player State Surface（Step 5；observation-only）— `player_state.py`

讓玩家/QA 看到「身上有什麼、知道什麼、最近碰過什麼、焦點在哪」。**純投影，不是新狀態來源。**

- `inventory_entities`：`state=taken` 或 `props.carried` 的 object（安全區也顯示；spatial visible 排除這些）。
- `known_facts`：`kind=fact` → `{id,label,state,source,confidence,tags}`，**保留自然語意 label**（非粗 key），
  NPC fact 仍只是 `npc_claim`、**不顯示 hidden truth 原文**。
- `current_focus` / `recent_entities`（上限 8）：互動優先序 **talk > inspect/take > exit > area**；對話 → 焦點=NPC；
  review 不亂改焦點。
- `changed_entities_this_beat` 升級為 `{id,label,kind,reason,from_state,to_state}`（reason 只從 delta/state 推、**不猜敘事**）。
- `player_state_summary`：確定性、字數上限+truncated、`source=deterministic_projection`，**不取代 narrative、不推 reveal**。

## 四、Entity Alias / Focus Resolver（Step 6；唯讀解析）— `alias_resolver.py`

讓「那枚 / 剛才那個 / 他說的地方 / 那個 NPC / 這東西」對到正確既有 entity。**不建 entity、不推 reveal、無 embedding/LLM。**

- `normalize_label`（NFKC 全→半形 + 去空白標點；不做繁簡/embedding）+ 弱修飾詞剝除（那個/剛才的/這枚…）。
- **aliases**：`label + props.aliases`（`add_alias` 存 props，**不建新 entity、衝突不自動合併**）。
- `resolve_entity_reference` 解析順序：**explicit id → exact label/alias → fact-ref→known_facts →
  npc-ref→actor → current_focus → recent → visible → inventory → unresolved**；候選平手 → **ambiguous（不亂選）**。
- observation 曝 `entity_resolution{query, resolved_entity_id, resolution_source, ambiguous, candidates}`。

---

## 五、整合：observation 完整面貌（GUI / QA / AI 都吃這份）

```text
narrative / decision_point
world_progress    : current_area / area_roles / investigation_state / changed(含 reason) / world_model
spatial_debug     : routes / blocked / visible / known_remote / mental_map_text
spatial_summary   : 玩家面空間摘要（deterministic）
player_state      : inventory_entities / known_facts / current_focus / recent_entities
player_state_summary : 玩家面狀態摘要（deterministic）
entity_resolution : 自然指代解析
beat_rendering    : beat_type / 字數 / short_streak（debug）
debug             : action_class / reveal_gate_allowed / blocked_reveal_candidates …
```

> **真 LLM 整合驗收（10 點全過）**：①檢查物件→WorldModel 登記 ②「那枚」重訪解析 ③NPC onboarding 非 API
> ④NPC fact→known_facts 不推 reveal ⑤引用 NPC fact 找路不推 reveal ⑥撤退→review_mode/ended=false
> ⑦spatial/player 摘要正常 ⑧truth_investigation 仍推 reveal（Δ=4）⑨focus/recent/alias 正確 ⑩GUI 欄位齊全。

## 六、里程碑

```text
WorldModel      記得世界（實體/角色/出口/事實）
SpatialSummary  告訴玩家在哪、能走去哪
PlayerState     告訴玩家有什麼 / 知道什麼 / 焦點在哪
AliasResolver   讓玩家自然指涉「剛才那個」「他說的地方」
```

開放式探索核心至此完整：世界有記憶、玩家有狀態、指涉自然、真相只在玩家真的調查時前進。
