---
id: NC6
stage: N
lane: D
title: Ending Gate（結局因果門檻；0/8 不可 clean escape）
status: done
worktree: -
depends_on: [NC4, UB7]
contracts: [EndingGate, EndingConditions]
last_good_snapshot: NC6__post__20260604-142553
owner_session: -
---

# NC6 · Ending Gate

- **階段**：N（敘事控制 · patch v0.1）　**工線**：D（Agent Logic）
- **依賴**：NC4, UB7(masked 結局)　**契約**：EndingGate, EndingConditions

## 目標 / 範圍
結局不再突然發生——加因果門檻 `EndingGate`。**refine UB7/attractor**：
- `clean_escape` 需 `explicit_escape_action` + `exit_location_reached` + `threat_resolved_or_avoided`。
- **0/8 真相碎片不可 clean_escape**，只可 `ambiguous_escape` 或 fail-forward。
- masked ending 只揭露玩家已確認/接觸過的真相（沿用 UB7）。

## 對應來源
patch `task_cards/P6_ending_gate.md`、`docs/07-ending-gate-spec.md`、`reference_code/ending_gate.py`、`tests_reference/test_reference_ending_gate.py`。

## 實作步驟
- 建 EndingGate：在 attractor/warden 觸發結局前，檢查因果門檻；不滿足 → 降級結局型別（clean→ambiguous/fail-forward）。
- 接進 loop `_finalize_ending`/dominant_ending 前；與 UB7 masked render 串接。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_ending_gate.py：缺 explicit_escape_action / 未抵達出口 / 威脅未解 → 不給 clean_escape；0/8 碎片 → 降為 ambiguous_escape/fail-forward；滿足三條件 → clean_escape；masked 不露 hidden truth。
- [ ] 既有 attractor/warden 結局不破；全回歸綠。

## 回滾備註
旁路 gate（flag 控制）；UB7 render 不變；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NC6 pre`
- 完成：驗收 pass → `snapshot.py snapshot NC6 post --verify pass` → 回填 last_good_snapshot
