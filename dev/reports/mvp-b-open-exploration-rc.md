# MVP-B Open Exploration Runtime — Release Candidate（RC）收尾

> 狀態：**RC / stabilization candidate**。開放式恐怖探索核心的 runtime 契約已成形、由 mock + 單元測試決定性覆蓋。
> 本檔總結架構、列出通過的核心契約、引用 mock 回歸結果，並標明唯一 known monitor 與「暫不進 P5」的決定。
> 基線：**全測 1005 passed / 3 skipped**（flag OFF/ON）；mock 回歸 **36/37**（0 真 regression、1 known monitor）。

---

## 一、架構總覽（開放式探索 runtime）

核心理念：**世界有記憶、玩家有狀態、指涉自然、真相只在玩家真的調查時前進**。各層皆為 WorldModel 的
**確定性投影**或**旁路控制**，受 `ENABLE_NARRATIVE_CONTROL` / `ENABLE_OPENING_VARIATION_CONTRACT` 控制。

```text
WorldModel（唯一事實來源：area/exit/object/actor/fact + roles + previous_area）
  ├─ Opening Variation Contract  開場核心素材去錨定（抽象 archetype 池 + cooldown）          → 19
  ├─ TruthEvidenceGate           只有 truth_investigation / 結構化 evidence 才推 reveal       → 16
  ├─ Reveal Reward Loop          gate=True 的調查保證可觀測回報（hinted→observed→suspected）  → reward
  ├─ RevealLedger + public_title 揭露階梯 + public-safe 標題（永不「未命名的真相」/ raw content）→ fragment-title
  ├─ PlayerState Payoff          take→inventory、NPC/reveal→known_facts（沉澱進既有 surface）  → 18
  ├─ Spatial Routes Projection   roles/current/previous → 可走路線（return/withdraw/site/exit）→ spatial
  ├─ AliasResolver Focus-Scope   那個東西→object / 他說的地方→fact / 那個人→actor             → alias
  ├─ NPC onboarding + actor 一致  first-contact gate + ensure_actor_entity（focus 一定對得到）  → 18
  └─ SurfaceTextSanitizer        壞分隔符 / placeholder / 黏拉丁殘片清除（後端，前端不解析）    → sanitizer
```

數字標的對應 `nightmare-assault-design-fixed/`：`16-worldmodel` / `17-spatial-projection` / `18-player-surface` /
`19-opening-variation`，以及本批的 reward / payoff / alias / fragment-title 報告。

---

## 二、通過的核心契約（mock + 單元測試決定性覆蓋）

| 契約 | 狀態 | 主要驗證 |
|---|---|---|
| Opening Variation 不使用 forbidden literals、無 placeholder 外洩 | ✅ | mock A / 真 LLM after-run（紙條率 0、6/6 medium 不同） |
| placeholder / 壞分隔符被擋（sanitizer，含簡繁） | ✅ | mock D + 單元測試 |
| **WorldDelta id warning 不再出現**（id→entity_id；衝突拒絕） | ✅ | mock D + ACCEPTANCE no_worlddelta_id_warning |
| PlayerState inventory 沉澱（take→taken→inventory；inspect 不取） | ✅ | mock A |
| known_facts 從 NPC fact + reveal public fact 填入（label/source/confidence） | ✅ | mock A / C |
| reveal 可升 observed / suspected（Reveal Reward Loop） | ✅ | mock C + 單元測試 |
| **confirmed 仍須 strong / structured evidence**（reward 上限 suspected） | ✅ | mock C（reward_never_confirmed + strong_evidence_can_confirm） |
| reveal known_fact 用 public_title、**永不「未命名的真相」/ 不外洩 hidden content** | ✅ | mock C + 單元測試 |
| spatial_summary 有可走路線（return_previous / withdraw_safe / return_site / exit） | ✅ | mock A / B |
| `no_truth_intent` 擋逃避型「不想管 / 只想離開」 | ✅（本批修） | mock B + regression |
| AliasResolver：那個東西→object（不被 NPC focus 吃）/ 他說的地方→fact / 那個人→actor | ✅ | mock A + 真 LLM alias smoke |
| NPC actor entity 一致（focus 一定對得到 WorldModel actor entity） | ✅ | mock A + 真 LLM reconfirm 6/6 |
| review_mode 不新增 fact、不推 reveal | ✅ | mock B |
| hidden truth raw content 不外洩（story 結構性看不到 real_bible；known_facts 只 public title） | ✅ | mock C + C2/E2 結構隔離 |
| TruthEvidenceGate 擋/放邏輯未被任何 payoff/title/reward 改動 | ✅ | 全程未改 gate（多次 contract regression / mock 佐證） |

---

## 三、mock 回歸結果（無真 LLM）

`dev/tools/mock_full_ux_regression.py`（deterministic ScriptedCaller 取代 OpenRouter）→
`dev/reports/mock-full-ux-regression.md` / `.jsonl`：

```text
A 探索型  11/11 ✅
B 逃避型   9/10  （1 ⚠ known monitor）
C 真相型   8/8  ✅
D 污染/邊界 7/7  ✅
Acceptance 1/1  ✅
總計      36/37  pass、0 真 regression、1 已知 monitor
```

Acceptance Checklist：13 項中 **12 ✅、1 ⚠ 已知**（見下）。

---

## 四、唯一 known monitor：retreat → review_mode（UX #6）

- **現象**：mock 的撤退 action 被 kernel 解析成一般移動（current_area→corridor），未翻 `review_mode`。
- **對照**：**真 LLM `spatial-routes-smoke` 曾成功**翻 `review_mode` 並移到 `safe_zone`——與 kernel / exit-resolver /
  action 措辭相關。
- **其餘 review 不變量**（不新增 fact、不推 reveal）在 mock 下**成立**。
- **決定**：依既定方針，retreat→review 列 **regression monitor**（見 memory `ux6-retreat-review-mode-monitor`），
  **本 RC 不改 retreat**。確認方式：未來一次極短真 LLM smoke，或排專屬 retreat-detection patch。

---

## 五、目前**不進** P5 story context（明確決定）

- P5（把開放式探索的 world/player/spatial 投影**回灌進 story 的生成 context**）**不在 RC 範圍**。
- 理由：RC 目標是**穩定既有 runtime 契約**，不擴生成面耦合；P5 會把確定性投影與 LLM 生成綁更緊，
  風險與測試面更大，應在 RC 穩定後另開 phase。
- 現況：story 仍只讀 `revealed_bible` 等安全欄位（結構性防暴雷 C2/E2 不變）；payoff/reveal/alias 等皆
  **observation / 旁路**，不進 story prompt。

---

## 六、RC 判定

- 開放式探索 runtime 的**主要契約全綠**（mock + 1005 單元測試），唯一未過為 retreat→review 的已知 monitor item。
- **建議**：標記 MVP-B Open Exploration Runtime 為 **RC / stabilization candidate**；後續（P5 / GUI polish /
  真 LLM 短 smoke / retreat monitor）移入 Next Phase（見 `dev/ROADMAP.md`）。
- **真 LLM 預算建議**：核心契約毋須再燒；**只在 retreat→review 上**值得一次極短真 LLM 確認。
