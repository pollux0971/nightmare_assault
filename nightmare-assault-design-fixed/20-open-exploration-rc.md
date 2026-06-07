# 20 · 開放式探索 Runtime — Release Candidate（Mock-stabilized RC）

> **承接**：`16-worldmodel` / `17-spatial-projection` / `18-player-surface` / `19-opening-variation`。
> 本檔是**開放式探索 runtime 的收尾總覽**——把補丁九～補丁十四之後的整輪契約鏈整合成一個 RC 里程碑。
> **版本定位**：Mock-stabilized RC（mock + 單元測試決定性穩定）。實作收尾報告：
> `dev/reports/mvp-b-open-exploration-runtime-rc.md`。
> **一句話**：世界有記憶、玩家有狀態、指涉自然、真相只在玩家真的調查時前進——且每條鏈都不越界、不進 story 生成。

---

## 一、核心鏈（已完成、彼此邊界清楚）

| 核心鏈 | 一句話 | 設計章 |
|---|---|---|
| Opening Variation Contract | 開場素材去錨定（抽象 archetype 池 + cooldown）；紙條/林晨/找人 不再是預設 | `19` |
| WorldModel entity persistence | area/exit/object/actor/fact + roles + previous_area 的唯一事實來源（+ 持久化） | `16` |
| Spatial Routes Projection | roles/current/previous → 可走路線（return_previous/withdraw_safe/return_site/exit） | `17` |
| PlayerState Payoff Materialization | take→inventory、NPC/reveal→known_facts；taken 退 visible（focus 例外） | `18` |
| AliasResolver Focus-Scope | 那個東西→object（不被 NPC focus 吃）/ 他說的地方→fact / 那個人→actor | `18` |
| NPC actor entity consistency | focus 一定對得到 WorldModel actor entity（ensure_actor_entity） | `18` |
| TruthEvidenceGate | 只有 truth_investigation / 結構化 evidence 才推 reveal（含擋逃避型「不想管/只想離開」） | `16` |
| Reveal Reward Loop | gate=True 的調查保證可觀測回報（hinted→observed→suspected）；reward 不直達 confirmed | `16`(延伸) |
| Public Reveal Labels | reveal known_fact 用 public_title（explicit / slug fallback）；永不「未命名的真相」、不外洩 hidden content | `18`(延伸) |
| Placeholder / delimiter sanitizer | 壞分隔符 / placeholder / 黏拉丁殘片清除（繁簡 + 開場消毒；後端） | `14`/`07`(SurfaceTextSanitizer) |

**共同邊界**：以上皆 observation / 確定性投影 / 旁路控制——**不改 WorldModel 權威、不改 TruthEvidenceGate 擋/放、
不進 story 生成 context**（story 仍只讀 `revealed_bible`，結構性防暴雷 C2/E2 不變）。

---

## 二、玩法路線（mock 驗證）

- **探索型**：檢查/撿物 → inventory 沉澱；問 NPC → known_facts（自然 label + source/confidence）；
  「剛才那個東西」對到 object、「他說的方向」對到 fact；spatial 永遠有可走路線。
- **逃避型**：「不想管/只想離開」→ no_truth=True、gate=False、reveal 不動；review 不新增 fact/不推 reveal；
  明確「結束調查」才進 ending。
- **真相型**：研究/分析/比對 → gate=True → reward 階梯 hinted→observed→suspected；reward **不**直達 confirmed
  （confirmed 須強/結構化 evidence）；known_fact 顯示 public_title、hidden raw 不外洩。
- **污染/邊界**：sanitizer 清 placeholder/壞分隔符；WorldDelta `id` 正規化、衝突 id 拒絕；story 越界加 area/exit 被拒。

mock 結果：**36/37 checkpoints pass**（0 真 regression、1 monitor）。

---

## 三、唯一 monitor item：retreat → review_mode（UX #6）

撤退在 mock 下未翻 `review_mode`（kernel 解析成一般移動）；真 LLM spatial-routes smoke 曾成功翻。
其餘 review 不變量（不新增 fact、不推 reveal）成立。列 regression monitor，**RC 不改 retreat**。

---

## 四、不進 P5、後續時序

- **目前不進 P5 story context**（把投影回灌進 story 生成）——避免把確定性投影與 LLM 生成綁更緊、放大測試與防暴雷面。
- **時序建議**：① short real-LLM smoke（重點 retreat→review）② GUI gameplay polish ③ retreat monitor
  ④ **P5 story context 只在 GUI 驗證玩法成立後**才接。

---

## 五、里程碑

```text
WorldModel      記得世界
SpatialSummary  告訴你在哪、能走去哪
PlayerState     告訴你有什麼 / 知道什麼（+ public-safe 真相標題）
AliasResolver   讓你自然指涉「剛才那個 / 他說的地方 / 那個人」
TruthGate+Reward 真相只在你真的調查時前進，且投入有可觀測回報
Opening/Sanitizer 開場不重複、表層不洩漏 placeholder
→ MVP-B Open Exploration Runtime = RC / stabilization candidate
```
