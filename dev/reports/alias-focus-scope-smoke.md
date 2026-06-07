# AliasResolver Focus-Scope — Focused Smoke（commit `d7a6614`）

> 真 LLM（deepseek-chat-v3-0324），flag NARRATIVE_CONTROL + OPENING_VARIATION ON。**只驗證，未改 core。**
> 主題：午夜的廢棄海事研究站。逐 beat 資料：`alias-focus-scope-smoke.jsonl`（run2）。跑了兩場：
> run1（純真 NPC）、run2（NPC + 注入 route fact / 兩本筆記本，作法同使用者設計的「製造同類物件」）。

## 結論：Focus-Scope 的「依 scope 解析、不被 NPC focus 卡住」核心**在真 LLM 下成立** ✅

|  | query | 解析 | 結果 |
|---|---|---|---|
| **物件指代（NPC focus 下）** | 「剛才那個東西」（focus=actor:陳博義） | → **object.老式頻率表**（src=recent_entities） | ✅ **不被 NPC focus 卡住**（UX #5 核心修復，run1/run2 皆成立） |
| **事實指代** | 「他說的方向、他說的地方」 | → **fact.離開需要主控室的授權卡**（src=known_facts） | ✅ 對到 recent/known fact（run2 該 NPC 真的產出此 fact） |
| **人物指代** | 「那個人」 | → **actor:陳博義**（src=current_focus） | ✅ 對到 NPC（actor scope → current_focus） |
| **不推 reveal** | 上述各 alias beat | reveal 全程 `0/0/0/0`、action_class 非 truth_investigation | ✅ 解析本身不推 reveal |
| **entity_resolution 欄位** | query/resolved_entity_id/resolution_source/ambiguous/candidates | 全部存在 | ✅ |

## 逐 beat（run2）

```
gen_object   生成「老式頻率表」（object_inspection）
take_inspect 撿起 → focus=object:老式頻率表（label_alias）
chat 陳博義   known_facts=[{離開需要主控室的授權卡, source=陳博義}]（NPC 真的產出 fact）
b4 「剛才那個東西」 focus=actor:陳博義 → object.老式頻率表 (object) src=recent_entities  ✅ 不被 NPC 卡住
b5 「他說的方向、他說的地方」 → fact.離開需要主控室的授權卡 (fact) src=known_facts          ✅ 對到 fact
b6 「那個人」 → actor.陳博義 src=current_focus                                            ✅ 對到 NPC
b7 「我的視線落在那本筆記本」 → object.老式頻率表 (recent)                                  ⚠ 見下（extract_reference）
```

## 三個「未 PASS」全是 scope 邏輯**以外**的因素（已逐一定位）

### A. 兩本筆記本未回 ambiguous（b7）—— `extract_reference` 視窗污染，**非 scope bug**
b7 的動作「**我的**視線落在那本筆記本」開頭是「我的」（弱觸發詞），`extract_reference` 以最早觸發詞 +
固定 12 字視窗，抓回整句「我的視線落在那本筆記本」→ noun 被「視線落在」污染 → label 對不到
「紅色筆記本/黑色筆記本」→ 落回 recent object。**resolver 的 ambiguous 邏輯本身正確**——本地驗證（無 LLM）：

```
action='我的視線落在那本筆記本' → extract='我的視線落在那本筆記本' → resolved=recent, ambiguous=False
action='我看那本筆記本'         → extract='那本筆記本'           → resolved=None, ambiguous=True, cands=[nb_red, nb_black]
action='我盯著看那本筆記本'      → extract='那本筆記本'           → ambiguous=True
```

只要動作不以「我的」開頭 / 指代置於可乾淨擷取處，**NPC focus 下「那本筆記本」即正確回 ambiguous**。
單元測試 `test_ambiguous_two_notebooks` / `test_ambiguous_still_returned_within_scope` 亦決定性保證。
→ **真正的可改進點在 `extract_reference`（最早觸發詞 + 定長視窗），是與 Focus-Scope 分離的元件**，可列後續 patch。

### B. 「5_no_new_area_exit」=False —— **kernel 在該 beat 移動了，非 resolver 建 area**
b5 的 area_count 由 2→3：是 **kernel** 處理含「方向/地方」的動作時新增/移動了 area，**不是** alias 解析造成。
exit_count 全程 0。resolver 為唯讀、provably 不建 area/exit（單元測試
`test_fact_direction_ref_resolves_to_fact_no_new_area`）。→ 檢查設計把「該 beat 的 kernel 後果」誤算進來。

### C. 「那個人」resolved_kind=None —— NPC 未被登記成 WorldModel actor entity
b6 解析到 `actor.陳博義`（src=current_focus，focus.kind=actor）——**指代正確對到 NPC**。但 `world.get("actor.陳博義")`
回 None，因為本場 陳博義 只在 `npc_registry`/focus，未被敘事抽取成 WorldModel **actor entity**（run1 的 謝孟川
有被登記，resolved_kind=actor，檢查 PASS）。→ resolver 行為正確；這是 **NPC→WorldModel actor entity 登記一致性**
的相鄰小缺口（與 scope patch 無關），可列後續觀察。

## run1 附帶發現（純真 NPC）
run1 的 NPC（謝孟川）**未自動產出任何 fact**（`known_facts=[]`），故「他說的方向」無 fact 可解 → unresolved。
這是 **NPC prose-fact 產出的機率性**（NPC 回覆未命中 location/locked-exit/action 主張式），非 resolver 問題；
run2 改用會產出 fact 的問法 + 注入，fact-scope 即正確解析。物件/人物指代在 run1 同樣成立（b4→object、b6→actor）。

## 一句話
AliasResolver Focus-Scope Patch 的**核心目標在真 LLM 下達成**：NPC focus 下「那個東西」對到 object、
「他說的地方/方向」對到 fact、「那個人」對到 NPC，且不推 reveal、entity_resolution 欄位齊全。三個未 PASS
皆為相鄰元件/檢查設計造成（`extract_reference` 視窗、kernel 移動、NPC actor-entity 登記），**非 scope 邏輯失誤**；
其中 `extract_reference` 的指代擷取最值得列為下一個 polish。
