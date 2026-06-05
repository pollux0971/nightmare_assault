---
id: HA1
stage: H
lane: A
title: Ending Observation Invariant（ended ⇒ options=[]）
status: done
worktree: main
depends_on: [U15, NR4]
contracts: [EndingObservationInvariant]
last_good_snapshot: HA1__post__20260605-101624
owner_session: session-24-opus
---

# HA1 · Ending Observation Invariant

- **階段**：H（Runtime Hard-Gate v0.3.1）　**工線**：A（整合者 / loop）
- **依賴**：U15（beat 主迴圈）、NR4（兩段式逃脫）
- **契約**：EndingObservationInvariant（見 dev/CONTRACTS.md §十五）

## 目標 / 範圍
修「`ended=true` 仍帶 options」。beat observation 輸出前強制不變式：結局時 options=[]、free_input_hint=null。

## 對應來源
patch `task_cards/P0_A1`、`docs/02`/`03`、`reference_code/models.py`（BeatObservation.enforce_invariants）。

## 實作步驟
- `_step_kernel` 最終輸出前：若 `self.ended` → 回傳的 decision_point 不再帶 suggested_options（或 step 回傳 dict 標準化 options=[]）。
- `agent_play._dp_to_obs`：ended 時 `options=[]`、`free_input_hint=""/null`。
- webview：ended 時前端不渲染選項（API 層 get/observation 已不給 options）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_ending_invariant_ha1.py：構造 ended 結果 → observation.options == [] 且 not (ended and options)。
- [ ] agent_play 結局那筆 observation options 為空；既有全套件綠；board --check 0 errors。

## 回滾備註
純輸出層不變式；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HA1 pre｜完成：snapshot.py snapshot HA1 post --verify pass
