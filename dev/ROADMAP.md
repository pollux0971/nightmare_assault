# dev/ROADMAP — 單線程開發工作流（v0.8.1 之後）

> **更新**：2026-06-06。承接已完成的 Config Center GUI（v0.8）、Patch B P0–P3（v0.7）、WorldDelta fix（v0.8.1）。
> **原則**：**單線程**——一次只做一個 Step，做完即 `測試（flag OFF/ON）→ config-safe 檢查 → commit → push`，才進下一個。
> **不靠記憶**：每個 Step 有明確 scope / 動到的檔 / non-goals / 驗收 / gate；中斷後照本續接。

---

## ✅ 已完成（不要重做）

| 編號 | 內容 | commit | 測試 |
|---|---|---|---|
| Step 0 | WorldDelta input normalization（id→entity_id；衝突拒絕）| `cd33ff9` | 5 |
| Step 1 | Config Center GUI Completion（Agent Models 表 / Prompt Blocks 表 / 多 agent / compiled preview / Test 入口；key 不外洩）| `7a94d17` | 9 |
| Step 2 | Patch B P0–P3（actor profile / first-contact gate / beat_rendering debug；P4 未開）| `07062ef` | 17 |
| — | design-fixed/skills 同步（隨 Patch B P0）| `dd5c0d7` `07062ef` | — |

全測基線：**878 passed, 3 skipped**（flag OFF/ON）。

---

## ▶ Step 4（下一個）— Patch B 完整 UX selfplay 驗證

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

## ▶ Step 5 — Player State Surface Patch（探索體驗補強，runtime observation-only）

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

## ▶ Step 6 — Entity Alias / Focus Resolver（自然指涉解析）

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
