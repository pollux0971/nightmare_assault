# Mock Full UX Regression（**無真 LLM** / ScriptedCaller）

> deterministic ScriptedCaller 取代 OpenRouter，驅動 open-exploration runtime 的主要契約。
> 由 `dev/tools/mock_full_ux_regression.py` 產生。逐 checkpoint 資料：`mock-full-ux-regression.jsonl`。
> **不呼叫 OpenRouter、不使用真 LLM。**

## 結果：36/37 通過（1 個未過為**已知 monitor item**、0 個真 regression）

| 路線 | 通過 | 未過(真) | 已知 monitor |
|---|---|---|---|
| A 探索型 | 11/11 | 0 | 0 |
| B 逃避型 | 9/10 | 0 | 1 |
| C 真相型 | 8/8 | 0 | 0 |
| D 污染/邊界 | 7/7 | 0 | 0 |
| Acceptance | 1/1 | 0 | 0 |

## A 探索型（A_explorer_mock）

| ✓ | checkpoint | evidence |
|---|---|---|
| ✅ | opening_no_forbidden_literals | `{"forbidden": [], "violation": false}` |
| ✅ | story_object_registered | `{"objects": ["object.frequency_meter"]}` |
| ✅ | inspect_does_not_take | `{"state": "inspected"}` |
| ✅ | take_adds_inventory | `{"inventory": ["頻率表"], "inventory_delta": [{"id": "object.frequency_meter", "label": "頻率表"}]}` |
| ✅ | taken_visible_when_focus | `{"focus": "object.frequency_meter"}` |
| ✅ | npc_actor_entity_exists | `{"focus_id": "actor.林守一", "kind": "actor"}` |
| ✅ | known_facts_from_npc_with_source_confidence | `{"npc_facts": [{"id": "fact.離開需要主控室的授權卡", "label": "離開需要主控室的授權卡", "state": "asserted", "source": "林守一", "confidence": "npc_claim", "tags": […` |
| ✅ | taken_not_visible_when_focus_moved | `{"focus": "actor.林守一"}` |
| ✅ | alias_object_not_npc | `{"resolved": "object.frequency_meter", "source": "recent_entities"}` |
| ✅ | alias_fact_no_new_area_exit | `{"resolved": "fact.離開需要主控室的授權卡", "source": "known_facts", "area_delta": 0, "exit_delta": 0}` |
| ✅ | spatial_has_withdraw_and_return_previous | `{"routes": ["route.return_previous", "route.withdraw_safe"], "summary_line": ["可走路線：返回上一個區域（維修入口艙）、暫退到安全區整理"]}` |

## B 逃避型（B_avoider_mock）

| ✓ | checkpoint | evidence |
|---|---|---|
| ✅ | no_truth_intent_true | `{"no_truth_intent": true, "action_class": "unknown"}` |
| ✅ | action_not_truth_investigation | `{"action_class": "unknown"}` |
| ✅ | gate_false | `{"gate": false, "reason": "explicit_no_truth_intent"}` |
| ✅ | reveal_delta_zero | `{"before": [0, 0, 0, 0], "after": [0, 0, 0, 0]}` |
| ✅ | spatial_routes_present | `{"routes": ["route.return_previous", "route.withdraw_safe"]}` |
| ⚠ | retreat_enters_review_or_temporary | `{"exploration_mode": "active_exploration", "current_area": "corridor"}` |
|  |  | ⚠ **已知**：UX#6 retreat→review_mode monitor：本 mock action 未翻 mode（kernel 解析成一般移動）；真 LLM spatial-routes smoke 曾成功翻 review_mode。需真 LLM 確認或專屬 retreat patch（本 batch 不改 retreat）。 |
| ✅ | retreat_not_ended | `{"ended": false}` |
| ✅ | review_no_new_fact | `{"facts_before": 0, "facts_after": 0}` |
| ✅ | review_no_reveal_push | `{"before": [0, 0, 0, 0], "after": [0, 0, 0, 0]}` |
| ✅ | explicit_end_triggers_ending | `{"ended": true, "ending": "escape"}` |

## C 真相型（C_truth_mock）

| ✓ | checkpoint | evidence |
|---|---|---|
| ✅ | truth_investigation_gate_true | `{"action_class": "truth_investigation", "gate": true}` |
| ✅ | reward_hinted_to_observed | `{"level": "observed", "ladder": "advanced_by_reward", "prev": "hinted", "next": "observed"}` |
| ✅ | reveal_reward_debug_fields | `{"gate_allowed": true, "ladder_action": "advanced_by_reward", "previous_level": "hinted", "next_level": "observed"}` |
| ✅ | reward_observed_to_suspected | `{"level": "suspected"}` |
| ✅ | reward_never_reaches_confirmed | `{"level": "suspected"}` |
| ✅ | public_known_fact_uses_title | `{"labels": ["線索：Logbook", "異常訊號"]}` |
| ✅ | no_hidden_raw_content_in_known_facts | `{"sample_labels": ["線索：Logbook", "異常訊號"]}` |
| ✅ | strong_evidence_can_confirm | `{"level": "confirmed"}` |

## D 污染/邊界（D_pollution_mock）

| ✓ | checkpoint | evidence |
|---|---|---|
| ✅ | sanitizer_removes_pollution | `{"remaining": [], "clean": "你伸手去口袋裡，門框上殘留，終端機顯示，他低聲說了句然。"}` |
| ✅ | sanitizer_removes_exact_delimiter | `{"clean": "敘事到此後面還有"}` |
| ✅ | worlddelta_id_normalized | `{"keycard": "object.keycard"}` |
| ✅ | conflicting_id_rejected | `{"obj_count_unchanged": true}` |
| ✅ | story_area_exit_rejected | `{"area_delta": 0, "exit_delta": 0}` |
| ✅ | npc_prose_fact_in_known_facts | `{"facts": ["通訊設備在B2機房", "那扇門已經鎖死了"]}` |
| ✅ | npc_prose_fact_no_reveal_push | `{"before": [0, 0, 0, 0], "after": [0, 0, 0, 0]}` |

## Acceptance（ACCEPTANCE）

| ✓ | checkpoint | evidence |
|---|---|---|
| ✅ | no_worlddelta_id_warning | `{"tick_skip_warnings": []}` |

## Acceptance Checklist 對應

| 項目 | 狀態 | 來源 checkpoint |
|---|---|---|
| opening variation 不使用 forbidden literals | ✅ | A:opening_no_forbidden_literals |
| placeholder / delimiter leak 被擋 | ✅ | D:sanitizer_* |
| WorldDelta id warning 不再出現 | ✅ | ACCEPTANCE:no_worlddelta_id_warning |
| PlayerState inventory_entities 能填入 | ✅ | A:take_adds_inventory |
| known_facts 從 NPC fact + reveal public fact 填入 | ✅ | A:known_facts_* / C:public_known_fact_uses_title |
| reveal 可升 observed / suspected | ✅ | C:reward_* |
| confirmed 仍需 strong evidence | ✅ | C:reward_never_* / strong_evidence_* |
| spatial_summary 有可走路線 | ✅ | A/B:spatial_* |
| no_truth_intent 擋「不想管 / 只想離開」 | ✅ | B:no_truth_intent_true（本 batch 修） |
| AliasResolver：那個東西→object / 他說的地方→fact / 那個人→actor | ✅ | A:alias_* / npc_actor_entity_exists |
| retreat 進 safe_zone / review_mode | ⚠ 已知 | B:retreat_enters_review_or_temporary（UX#6 monitor） |
| review_mode 不新增 fact、不推 reveal | ✅ | B:review_* |
| hidden truth raw content 不外洩 | ✅ | C:no_hidden_raw_content_in_known_facts |

## 本 batch 已修的 deterministic bug

- **no_truth_intent 漏判逃避型（UX #4）**：補「不想管 / 只想離開 / 只想出口 …」→ 逃避型現可被 `no_truth=True`、gate=False 擋住（regression test 已加）。

## 仍需真 LLM 才能確認 / 未在本 batch 修

- **retreat → review_mode（UX #6，monitor item）**：本 mock 的撤退 action 被 kernel 解析成一般移動（current_area→corridor），未翻 review_mode。**真 LLM spatial-routes smoke 曾成功**翻 review_mode + 移到 safe_zone（kernel/exit-resolver/action 措辭相關）。其餘 review 不變量（不新增 fact、不推 reveal）在 mock 下成立。依既定方針 retreat 列 monitor、本 batch 不改。

## 建議

- 核心契約在 mock 下**全綠**（唯一未過為 retreat→review 的已知 monitor item）。
- **是否再花真 LLM smoke**：建議**只在 retreat→review 上**花一次極短真 LLM 確認（其餘契約已由 mock + 單元測試決定性覆蓋，毋須再燒真 LLM 預算）。
