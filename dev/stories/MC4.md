---
id: MC4
stage: C
lane: D
title: Offstage-Fate agent + 命運 tick（程式碼擲骰 + LLM 寫血肉）
status: done
worktree: -
depends_on: [U03, U04, U10]
contracts: [OffstageFate, NPCRegistry]
last_good_snapshot: MC4__post__20260604-150033
owner_session: -
---

# MC4 · Offstage-Fate agent + 命運 tick

- **階段**：C（MVP-C）　**工線**：D（Agent Logic）
- **依賴**：U03(權限), U04(場景), U10(揭露)　**契約**：OffstageFate, NPCRegistry

## 目標 / 範圍
讓**離場 NPC**（presence absent/missing）有獨立命運。**程式碼加權擲骰**（受 alignment 影響）決定 fate_type
（opportunity_return / missing / corpse / hostile_return）→ LLM 寫血肉 → 領 revelation_pool 碎片（carried_fragment）
→ 寫 npc_registry[].{presence,alignment,offstage_intent,carried_fragment} + corpse interactable。非同步只產 patch。

## 對應來源
`skills/offstage-fate/SKILL.md`、00 §六 離場雙軸、02 §（離場命運）、07 §二（OffstageFateOutput）、CHECKLIST C5/C6/C7。

## 實作步驟
- `core/agents/offstage_fate.py`：`roll_fate(npc)`（程式碼加權，依 alignment）→ fate_type；`run_offstage_fate(caller, blackboard, npc, fragment)` → OffstageFateOutput → 提交 patch（只寫 npc/scene；碰不到 secret_core）。
- loop：命運 tick（低頻，遠低於 dreaming）對離場 NPC 跑；領未揭露碎片；屍體種 corpse interactable。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_offstage_fate_mc4.py：roll_fate 受 alignment 影響、機率可控；四種命運各正確寫 presence/alignment；carried_fragment 來自 revelation_pool；**只寫 npc/scene、碰不到 secret_core（權限邊界 C6）**；非同步只產 patch。
- [ ] corpse 命運種 corpse interactable；全回歸綠。

## 回滾備註
旁路 agent（命運 tick 可關）；只產 patch；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot MC4 pre`
- 完成：驗收 pass → `snapshot.py snapshot MC4 post --verify pass` → 回填 last_good_snapshot
