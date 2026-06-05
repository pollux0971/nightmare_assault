---
id: NR3
stage: R
lane: D
title: EndingGate Surface Variant（結局表層變體 — 模糊逃脫看得出來）
status: done
worktree: main
depends_on: [NC6, UB7]
contracts: [EndingSurface, EndingGate]
last_good_snapshot: NR3__post__20260605-011109
owner_session: session-24-opus
---

# NR3 · EndingGate Surface Variant

- **階段**：R　**工線**：D（agent）
- **依賴**：NC6（EndingGate 因果門檻）、UB7（masked 結局 render）
- **契約**：EndingSurface, EndingGate（見 §十四 / §十二）

## 目標 / 範圍
NC6 內部已會把 0 真相逃脫標 `ambiguous`，但**玩家看到的文字與乾淨逃脫一樣**。讓不同結局在表層真的不同。

## 對應來源
patch `task_cards/P3_EndingGateSurfaceVariant.md`、`docs/05-ending-gate-surface-and-escape-commit.md §A`、`reference_code/ending_gate.py`、`prompts_reference/ending_renderer_delta.md`、`examples/ending_gate_cases.example.json`。

## 實作步驟
- 新增 `ending_surface`（獨立於 ending_type）：clean_escape / ambiguous_escape / failed_escape / death / truth_locked。
- 由 RevealLedger（NR0）決定 clean vs ambiguous：**0 confirmed 真相 → 不可 clean_escape**。
- 新增 ambiguous_escape renderer：呈現不確定/污染/未解的身分威脅（docs/05 範式：「你走出去了。至少，你以為自己走出去了…」），與 UB7 masked 整合。
- debug 輸出印 gate 理由（為何 clean / ambiguous / truth_locked）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_ending_surface_nr3.py：0/X confirmed 逃脫 → `ending_surface == ambiguous_escape` 且 render 文字含不確定收尾（≠ clean_escape 文字）；達 observed_or_better → 可 clean。
- [ ] debug 印出 gate_reason；flag OFF 退回 UB7/NC6 現況；既有套件綠。

## 回滾備註
ending_surface 欄位 + renderer 變體；新欄位 optional；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：填 owner_session / worktree → `snapshot.py snapshot NR3 pre`
- 完成：驗收 pass → `snapshot.py snapshot NR3 post --verify pass` → 回填 last_good_snapshot
