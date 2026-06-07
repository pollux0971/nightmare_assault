# dev/ROADMAP — 單線程開發工作流（v0.8.1 之後）

> **更新**：2026-06-07。承接已完成的 Config Center GUI（v0.8）、Patch B P0–P3（v0.7）、WorldDelta fix（v0.8.1），
> 以及開放式探索 runtime 的整輪補強（Opening Variation / Reveal Reward / PlayerState Payoff / Spatial Routes /
> AliasResolver Focus-Scope / Actor 一致性 / Fragment Title）。
> **原則**：**單線程**——一次只做一個 Step，做完即 `測試（flag OFF/ON）→ config-safe 檢查 → commit → push`，才進下一個。
> **不靠記憶**：每個 Step 有明確 scope / 動到的檔 / non-goals / 驗收 / gate；中斷後照本續接。

---

## 🏁 里程碑：MVP-B Open Exploration Runtime = **RC / stabilization candidate**

開放式恐怖探索核心的 runtime 契約已成形且穩定——**收尾報告：`dev/reports/mvp-b-open-exploration-rc.md`**。

- **mock 回歸（無真 LLM）**：`dev/tools/mock_full_ux_regression.py` → **36/37 pass**、0 真 regression、
  1 已知 monitor（retreat→review_mode）。Acceptance Checklist 13 項：12 ✅、1 ⚠ 已知。
- **全測基線**：**1005 passed / 3 skipped**（flag OFF/ON）。
- **唯一 known monitor**：`retreat → review_mode`（UX #6；真 LLM smoke 曾成功翻 mode，mock 下未翻）——
  列 regression monitor，**RC 不改 retreat**。
- **明確不進**：**P5 story context** 不在 RC 範圍（移 Next Phase）。

| 補強（Step 6 之後，本輪） | commit | 重點 |
|---|---|---|
| WorldDelta id + placeholder hotfix | `628f762` | extractor id→entity_id；sanitizer 繁簡 + 開場消毒 |
| Opening Variation Contract（補丁十四）| `c6be633` | 開場素材去錨定（紙條/林晨/找人）；真 LLM medium 6/6 不同 |
| Reveal Reward Loop | `aad3684` | gate=True 的 truth_investigation 保證回報或 no_progress_reason |
| PlayerState Payoff Materialization | `a337116` | take→inventory；reveal→public known_fact；taken 退 visible |
| Spatial Routes Projection（+label polish）| `25ab123` `c3e65d6` | roles/previous→可走路線；不再「沒有明顯可走的出口」 |
| AliasResolver Focus-Scope（+extract polish）| `d7a6614` `94bff8c` | object/fact/actor scope；強指代 anchor + bounded noun |
| NPC actor entity 一致性 | `9156654` | ensure_actor_entity；focus 一定對得到 WorldModel actor |
| Fragment Title / Public Reveal Label | `4fe755b` | public_title；reveal known_fact 永不「未命名的真相」 |
| no_truth_intent 修（UX #4）| `93754d2` | 擋逃避型「不想管 / 只想離開」 |
| Mock Full UX Regression harness | `75b3818` | deterministic ScriptedCaller 四路線回歸（無真 LLM）|

---

## ▶ Next Phase（RC 穩定後再開；本輪不做）

> 以下移出 RC 範圍，待 MVP-B RC 穩定後另開單線程 Step。

| 項目 | 摘要 | 前置 |
|---|---|---|
| **P5 story context integration** | 把 world / player / spatial 投影**回灌進 story 生成 context**（耦合生成面，需更大測試面與防暴雷複驗 C2/E2）| RC 穩定 |
| **GUI gameplay polish** | 把開放式探索 observation（inventory / known_facts / spatial_summary / entity_resolution）呈現到前端遊玩面 | RC 穩定 |
| **real LLM short smoke** | 一次極短真 LLM 確認（**重點：retreat→review_mode**；其餘契約已由 mock + 單元測試覆蓋）| 預算允許 |
| **retreat transition monitor** | `retreat → review_mode`（UX #6）regression monitor；若持續不穩 → 排專屬 retreat-detection patch（exit-resolver / exploration_mode）| 真 LLM smoke 數據 |

---

## ✅ 已完成（不要重做）

| 編號 | 內容 | commit | 測試 |
|---|---|---|---|
| Step 0 | WorldDelta input normalization（id→entity_id；衝突拒絕）| `cd33ff9` | 5 |
| Step 1 | Config Center GUI Completion（Agent Models 表 / Prompt Blocks 表 / 多 agent / compiled preview / Test 入口；key 不外洩）| `7a94d17` | 9 |
| Step 2 | Patch B P0–P3（actor profile / first-contact gate / beat_rendering debug；P4 未開）| `07062ef` | 17 |
| — | design-fixed/skills 同步（隨 Patch B P0）| `dd5c0d7` `07062ef` | — |
| Step 4 | Patch B 完整 UX selfplay 驗證（真 LLM）：6 重點全 PASS；過程修 3 bug（研究站誤判 / 不碰真相誤鎖 review / beat 預算過高）| — | — |

全測基線：**878 passed, 3 skipped**（flag OFF/ON）。

**Step 4 驗證結論（真 LLM，dev/reports/step4-patchb.jsonl）**：①NPC 首現皆 unintroduced→introduced
②首問答非 API（114/91 字，含位置/動作/身分/態度/部分答案）③personality 明顯影響語氣（mysterious 留白反問 vs
nervous 短促重複）④一般 beat **short_streak 全 0**（修預算後無假性過短）⑤review 仍短且不新增 fact
⑥**TruthEvidenceGate 完好**（非 truth 行動 reveal_Δ=0；truth_investigation reveal_Δ=4）。
→ **Step 3（P4）判定：不需要**（max_short_streak=0）。

---

## ✅ Step 4（已完成）— Patch B 完整 UX selfplay 驗證

**目標**：用真 LLM 跑一場較完整的探索＋多次 NPC 對話，驗證 Patch B 體驗、並蒐集 beat_rendering 數據決定要不要開 P4。

- **動到的檔**：只新增/調 `dev/tools/`（selfplay 驅動器）+ 產 `dev/reports/`；**不改 runtime**。
- **重點測（你指定的 6 項）**：
  1. NPC 第一次出現是否有來歷線索（看 story 敘事 + actor profile intro_state）。
  2. NPC 第一次問答是否**不是純資訊 API**（first-contact reply 含位置/動作/身分/態度/部分答案）。
  3. personality 是否影響語氣（不同 personality 的 NPC 口吻不同）。
  4. 一般 beat 是否不再短成 log（逐 beat `beat_rendering`：type / actual_chars / too_short / **short_streak**）。
  5. review_mode 是否仍短且不新增 fact（investigation_state=review_mode、fact 集合不變）。
  6. **TruthEvidenceGate 不受影響**（找路/整理/引用 NPC fact 不推 reveal；可跑 contract regression 佐證）。
- **驗收**：報告含每 beat 的 beat_rendering + NPC onboarding 逐字 + reveal 不被誤推；明確結論「short_streak 是否常 ≥2」。
- **gate（→ Step 3）**：若報告顯示**一般 beat 連續 ≥2 過短反覆出現** → 才開 Step 3（P4）；否則 P4 維持關閉。
- **non-goals**：不改 prompt 實際內容、不改 WorldModel / reveal gate、不接 P5 story context。

---

## ⏸ Step 3（gated）— Patch B P4：short beat soft repair once

> **只有 Step 4 數據證明需要才做。** 目前 `short_streak` max=1（p5_onboarding_tryit），**暫不開**。

- **目標**：連續 2 個一般 beat 過短 → repair once（只擴場景表層，不改世界狀態）。
- **動到的檔**：`core/narrative/beat_rendering`（加 `should_repair`）+ loop（觸發 repair 一次）+ story repair 路徑。
- **硬限制**：repair **不得**改 WorldModel / reveal / area / exit / fact state；只擴敘事表層；每次最多一次；review/system_feedback 不 repair。
- **驗收**：注入 2 連續過短 → `repair_attempted=true` 一次；repair 後不新增 entity/fact/reveal/area/exit。
- **建議**：先調 `new_area_entry` 預算下限（350→~250，現況 LLM 換區 beat 常 167–273）再評估，可能根本不必開 P4。

---

## ✅ Step 5（已完成）— Player State Surface Patch（observation-only）

**目標**：把玩家狀態在 observation 結構化呈現（先前 selfplay 報告的缺口）。**只擴 observation 投影，不改世界權威。**

- **新增（observation / world_progress）**：
  - `inventory_entities` + `carried=true`（隨身物品，安全區也獨立顯示）。
  - `known_facts`（結構化 list，取代/補足粗 key world_facts 的可讀呈現；**不收斂 world_facts 本體**）。
  - `recent_entities` / `current_focus`（最近被指涉/檢查的實體）。
  - `changed_entities_this_beat` 補 **reason**（registered / inspected / taken / npc_claim …）。
- **動到的檔**：`core/world/model.py`（投影 helper，唯讀）+ loop `world_progress` + `agent_play` 觀測 + 測試。
- **non-goals**：不改 reveal gate、不改 WorldModel 權威、不收斂 world_facts、不接 story context、不加故事內容。
- **驗收**：observation 有上述欄位；安全區 carried 物品仍顯示；known_facts 結構化；changed reason 正確；reveal 不受影響。

---

## ✅ Step 6（已完成）— Entity Alias / Focus Resolver（自然指涉解析）

> **里程碑達成**：開放式探索核心完整——WorldModel 記得世界 · SpatialSummary 告訴你在哪 ·
> PlayerState 告訴你有什麼/知道什麼 · **AliasResolver 讓你自然指涉剛才的東西**。全測 899 passed。


**目標**：讓「那枚 / 剛才那個 / 他說的地方」能對到正確實體。

- **內容**：
  - `aliases`（actor profile 已有；擴到 object/fact）。
  - `recent focus` 堆疊（最近檢查/提及的實體）。
  - 指示詞解析（那枚/剛才那個/他說的X）→ 先查 recent focus，再查 alias，再 label normalization。
  - label normalization（去空白/全半形/同義）。
- **動到的檔**：`core/world/`（新 resolver 模組）+ loop 接線（玩家輸入解析時用）+ 測試。
- **non-goals**：不做 spatial reasoning / pathfinding、不改 reveal gate、不改 story prompt、不加故事內容。
- **驗收**：「我檢查剛才那個」對到上一個 focus 實體；「他說的機房」對到 NPC fact；alias 對映；malformed 不誤判。

---

## 流程節點（每個 Step 一致）

```text
認領 Step → 實作（先測試骨架、小步） → pytest（flag OFF + ON 各一次，全綠）
→ config-safe（git ls-files config/config.json=0、diff 無 sk-or-v1-6167）
→ commit（描述 scope + 非目標 + 測試數）→ push → 更新本檔狀態 → 下一個 Step
```

> **gate 提醒**：Step 3（P4）卡在 Step 4 的數據；Step 5/6 是探索體驗補強，排在 Patch B 驗證之後。
