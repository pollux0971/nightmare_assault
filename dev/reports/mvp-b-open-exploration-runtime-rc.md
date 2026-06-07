# MVP-B Open Exploration Runtime — Release Candidate

> **版本定位：Mock-stabilized RC**（mock + 單元測試決定性穩定；尚未做完整真 LLM 巡迴與 GUI 遊玩驗證）。
> 開放式恐怖探索核心 runtime 的契約鏈已成形且穩定。後續見 `dev/ROADMAP.md` 的 **Next Phase**。
> 詳細收尾另見 `dev/reports/mvp-b-open-exploration-rc.md`（契約逐項證據表）。

---

## 一、RC 判定摘要

| 項 | 值 |
|---|---|
| 版本定位 | **Mock-stabilized RC** |
| mock 回歸（`mock_full_ux_regression.py`，**無真 LLM**） | **36/37 checkpoints pass**、0 真 regression、1 monitor |
| Acceptance Checklist | 13 項：**12 ✅ / 1 ⚠ 已知** |
| 全測基線 | **1006 passed / 3 skipped**（flag OFF/ON） |
| 唯一 monitor item | **retreat → review_mode**（UX #6） |
| 明確不進 | **P5 story context**（移 Next Phase） |

---

## 二、已完成核心鏈（contract chain）

開放式探索 runtime 由以下確定性契約構成；皆 mock + 單元測試覆蓋、彼此邊界清楚（投影/旁路，不互相越界）：

| 核心鏈 | 一句話 | 主要 commit |
|---|---|---|
| **Opening Variation Contract** | 開場核心素材去錨定（抽象 archetype 池 + cooldown）；紙條/林晨/找人 不再是預設 | `c6be633` |
| **WorldModel entity persistence** | area/exit/object/actor/fact + roles + previous_area 的唯一事實來源（含 to_dict/from_dict 持久化） | 補丁九 + `25ab123` |
| **Spatial Routes Projection** | roles/current/previous → 可走路線（return_previous/withdraw_safe/return_site/exit）；不再「沒有明顯可走的出口」 | `25ab123` `c3e65d6` |
| **PlayerState Payoff Materialization** | take→inventory、NPC/reveal→known_facts；taken 退 visible（focus 例外） | `a337116` |
| **AliasResolver Focus-Scope** | 那個東西→object（不被 NPC focus 吃）/ 他說的地方→fact / 那個人→actor；強指代 anchor + bounded noun | `d7a6614` `94bff8c` |
| **NPC actor entity consistency** | ensure_actor_entity；focus 一定對得到 WorldModel actor entity | `9156654` |
| **TruthEvidenceGate** | 只有 truth_investigation / 結構化 evidence 才推 reveal；找路/整理/引用 NPC fact / 不想管 一律擋 | 既有 + `93754d2`(no_truth 補) |
| **Reveal Reward Loop** | gate=True 的調查保證可觀測回報（hinted→observed→suspected）或 no_progress_reason；reward 不直達 confirmed | `aad3684` |
| **Public Reveal Labels** | reveal-derived known_fact 用 public_title（explicit / slug fallback）；**永不「未命名的真相」、不外洩 hidden content** | `4fe755b` |
| **Placeholder / delimiter sanitizer** | 壞分隔符 / placeholder / 黏拉丁殘片清除（繁簡 + 開場消毒；後端，前端不解析） | `628f762` |

> 共同邊界：以上皆 **observation / 確定性投影 / 旁路控制**——不改 WorldModel 權威、不改 TruthEvidenceGate 擋/放、
> **不進 story 生成 context**（story 仍只讀 `revealed_bible` 等安全欄位，結構性防暴雷 C2/E2 不變）。

---

## 三、mock 回歸結果（無真 LLM）

`dev/tools/mock_full_ux_regression.py`（deterministic ScriptedCaller 取代 OpenRouter）：

```text
A 探索型  11/11 ✅   inventory / known_facts / alias(object·fact) / spatial routes
B 逃避型   9/10  ⚠   no_truth=True·gate=False·reveal_Δ0 / spatial routes / review 不新增 fact·不推 reveal / 明確 end→ending
                     （1 ⚠：retreat→review_mode，見 §四）
C 真相型   8/8  ✅   truth_investigation gate=True / reward hinted→observed→suspected / reward≠confirmed
                     / public_title 非「未命名」 / hidden raw 不外洩 / strong evidence→confirmed
D 污染/邊界 7/7  ✅   sanitizer 清污染 / WorldDelta id 正規化 / 衝突 id 拒絕 / story 加 area·exit 拒絕 / NPC prose fact 不推 reveal
Acceptance 1/1  ✅   no_worlddelta_id_warning
總計      36/37  pass、0 真 regression、1 已知 monitor
```

逐 checkpoint 證據：`dev/reports/mock-full-ux-regression.md` / `.jsonl`。

---

## 四、唯一 monitor item：retreat → review_mode（UX #6）

- **現象**：mock 的撤退 action 被 kernel 解析成一般移動（current_area→corridor），未翻 `review_mode`。
- **對照**：**真 LLM `spatial-routes-smoke` 曾成功**翻 `review_mode` 並移到 `safe_zone`——與 kernel / exit-resolver /
  action 措辭相關。
- **其餘 review 不變量**（不新增 fact、不推 reveal）在 mock 下**成立**。
- **決定**：列 regression monitor（memory `ux6-retreat-review-mode-monitor`），**RC 不改 retreat**。

---

## 五、明確：目前**不進** P5 story context

- P5（把 world / player / spatial 投影**回灌進 story 的生成 context**）**不在 RC 範圍**。
- 理由：RC 目標是穩定既有 runtime 契約；P5 會把確定性投影與 LLM 生成綁更緊，測試面與防暴雷複驗（C2/E2）成本大。
- 現況：所有 payoff / reveal / alias / spatial 皆 observation/旁路，**不進 story prompt**；防暴雷結構不變。
- **時機**：P5 應在 **GUI 遊玩驗證之後**再開（見後續建議）。

---

## 六、後續建議（Next Phase；本 RC 不做）

1. **short real-LLM smoke when budget allows** — 一次極短真 LLM 確認（**重點放 retreat→review_mode**；
   其餘契約已由 mock + 單元測試覆蓋，毋須再燒預算）。
2. **GUI gameplay polish** — 把開放式探索 observation（inventory / known_facts / spatial_summary /
   entity_resolution / reveal public_title）呈現到前端遊玩面。
3. **retreat transition monitor** — `retreat → review_mode`（UX #6）持續觀察；若真 LLM 也不穩 → 排專屬
   retreat-detection patch（exit-resolver / exploration_mode）。
4. **P5 story context integration — only after GUI validation** — GUI 驗證玩法成立、體驗缺口確認後，才接 P5。

---

## 七、結論

開放式探索 runtime 的**主要契約全綠**（mock 36/37 + 1006 單元測試），唯一未過為 retreat→review 的已知 monitor item。
**判定：標記 MVP-B Open Exploration Runtime 為 RC / stabilization candidate。** 真 LLM 預算只需在 retreat→review 上花一次極短確認。
