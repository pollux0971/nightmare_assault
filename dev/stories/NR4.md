---
id: NR4
stage: R
lane: D
title: Escape Commit Gate（兩段式逃脫 — 先出口候選，再明確提交）
status: done
worktree: main
depends_on: [NR3, SK13]
contracts: [EscapeCommitGate, EndingSurface, EndingConditions]
last_good_snapshot: NR4__post__20260605-011306
owner_session: session-24-opus
---

# NR4 · Escape Commit Gate

- **階段**：R　**工線**：D（agent / 迴圈）
- **依賴**：NR3（EndingSurface）、SK13（attractor-based ending）
- **契約**：EscapeCommitGate, EndingSurface, EndingConditions（見 §十四 / §十二）

## 目標 / 範圍
目前「我試圖離開」會**即刻觸發結局**。改成兩段式：先把第一次逃脫意圖轉成出口發現 beat，玩家明確提交才結算結局，避免一句話草草收尾。

## 對應來源
patch `task_cards/P4_EscapeCommitGate.md`、`docs/05-ending-gate-surface-and-escape-commit.md §B`、`examples/ending_gate_cases.example.json`。

## 實作步驟
- 新增狀態 `exit_candidate_found`。
- 首次 `attempt_escape` → 產出「出口發現」beat + 選項（提交離開 / 繼續調查 / 處理威脅），不即結局。
- 玩家明確 `commit_escape` 才跑 EndingGate.evaluate（接 NR3 表層變體 / SK13 attractor）。
- 與 attractor ending 協調：attractor 仍可導向結局，但逃脫路徑需經提交步驟。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_escape_commit_nr4.py：輸入「我試圖離開這裡」→ 系統呈現出口候選與選項，**非即時結局**；再選提交 → 才進 EndingGate。
- [ ] 未提交不結算；flag OFF 退回現況；既有套件綠。

## 回滾備註
exit_candidate_found 狀態 + 兩段式分流；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR4 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR4 post --verify pass` → 回填 last_good_snapshot
