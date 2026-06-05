---
id: HD1
stage: H
lane: F
title: QualityGate Repair Once（check→repair once→fallback）
status: done
worktree: main
depends_on: [NC5, NR5]
contracts: [StoryRepairPipeline, QualityGate]
last_good_snapshot: HD1__post__20260605-102912
owner_session: session-24-opus
---

# HD1 · QualityGate Repair Once

- **階段**：H　**工線**：F（品質）
- **依賴**：NC5（QualityGate）、NR5（母題冷卻 violations）
- **契約**：StoryRepairPipeline, QualityGate（見 §十五 / §十二）

## 目標 / 範圍
QualityGate 從「只 log」升級成 gate：check fail → repair story 一次（帶 repair_instruction 重生）→ 再 check；仍 fail → deterministic fallback（系統正確的方向 beat + 安全選項，不求文采）。**最多 repair 一次**，log `repaired/fallback`。

## 對應來源
patch `task_cards/P1_D1`、`docs/06`、`reference_code/quality_repair.py`、`prompts_reference/story_repair_prompt.md`。

## 實作步驟
- 新增 `core/narrative/quality_repair.py`：`StoryRepairPipeline(story_runner, quality_gate, deterministic_fallback)`：run → check → repair once → fallback。
- loop `_step_kernel`：把 run_story + check_beat 包進 pipeline；repair 用 ctx 加 `repair_instruction` 重跑 run_story 一次。
- deterministic_fallback：規則版 narrative（可驗證方向）+ 3 個安全選項（前往控制室/繼續追問/暫時撤離）。
- 觸發條件涵蓋 docs/06（ended+options、new lore、reveal jump、motif overuse、sanitizer 嚴重命中、answer debt 未償還）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_quality_repair_hd1.py：check fail 一次 → repair 後 pass → 回 repaired=true；repair 仍 fail → fallback（fallback=true、有安全選項、無 options-on-ending 衝突）；最多一次 repair（story_runner 至多被呼叫 2 次）。
- [ ] 既有套件綠；board --check 0 errors。

## 回滾備註
新增 pipeline + loop 包裝；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HD1 pre｜完成：snapshot.py snapshot HD1 post --verify pass
