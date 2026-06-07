# NPC Actor Entity Consistency — Short Re-confirm（commit `9156654`）

> 真 LLM（deepseek-chat-v3-0324），flag NC + OVC ON。**只驗證，未改 core。** 資料：`actor-consistency-reconfirm.jsonl`。
> 流程：與 NPC 對話（→ current_focus）→「那個人」指代 → 驗 resolved_kind=actor / world.get ≠ None / 不新增 fact / 不推 reveal。

## 結論：6/6 PASS ✅ — 補丁在真 LLM 下生效，補上 alias smoke 的 check 6

NPC=**蘇明哲**。對話後 `current_focus=actor.蘇明哲` 且 `world_entity_kind=actor`（actor entity 已建好）。
下一拍「那個人」→ `actor.蘇明哲`（kind=actor，src=current_focus）。

| 檢查 | 結果 |
|---|---|
| resolved_entity_id = `actor.*` | ✅ `actor.蘇明哲` |
| resolved_kind = actor | ✅ |
| `world.get(resolved_id)` ≠ None | ✅（之前為 None，現已建 entity） |
| current_focus 對得到 WorldModel actor entity | ✅ |
| 不新增 fact | ✅ |
| 不推 reveal | ✅ |

## 對照 alias smoke 的相鄰缺口

alias-focus-scope-smoke 的 check 6 曾因 NPC（陳博義）只在 npc_registry/focus、未在 WorldModel actor entities
而 `resolved_kind=None`。本次 `ensure_actor_entity_from_npc_registry`（接在 note_focus_npc）確保對話/設焦點時
就建好 actor entity → `resolved_kind=actor` 穩定成立。AliasResolver 這條鏈至此完整收束。

## 一句話
NPC actor entity 一致性在真 LLM 下確認生效：「那個人」穩定對到真正的 WorldModel actor entity，
不新增 fact、不推 reveal。AliasResolver 主體不再需要改動。
