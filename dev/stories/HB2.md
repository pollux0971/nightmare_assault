---
id: HB2
stage: H
lane: D
title: RevelationBridge Unified Inputs（多來源統一 + reveal_updates 可觀測）
status: done
worktree: main
depends_on: [NR0, HB1]
contracts: [EvidenceEvent, ObservationDebug]
last_good_snapshot: HB2__post__20260605-102259
owner_session: session-24-opus
---

# HB2 · RevelationBridge Unified Inputs

- **階段**：H　**工線**：D（agent）
- **依賴**：NR0（bridge）、HB1（story evidence）
- **契約**：EvidenceEvent（擴充）, ObservationDebug（reveal_updates 部分）（見 §十五）

## 目標 / 範圍
統一 kernel / story / player-action / npc / document 的 evidence 都走 `RevelationBridge.apply`，ledger 更新後寫 revealed_bible，並把 `reveal_updates_this_beat` / `evidence_events_this_beat` / `unmapped_evidence_this_beat` 暴露到 step 結果（供 observation/debug）。

## 對應來源
patch `task_cards/P0_B2`、`docs/04`/`08`。

## 實作步驟
- loop `_revelation_tick` 收斂成單一入口：彙整本 beat 所有來源 EvidenceEvent → 一次 `bridge.apply` → 回傳 updates。
- step 回傳 dict 加 `reveal_updates`、`evidence_events_this_beat`、`unmapped_evidence_this_beat`（HB1 的 unmapped 計入）。
- 確認 ledger 每次更新都 `write_ledger_to_revealed_bible`。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_bridge_unified_hb2.py：多來源 events 一次 apply → ledger/revealed_bible 正確；step 結果含 reveal_updates / evidence_events_this_beat / unmapped_evidence_this_beat。
- [ ] 既有套件綠；board --check 0 errors。

## 回滾備註
tick 收斂 + step 欄位；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HB2 pre｜完成：snapshot.py snapshot HB2 post --verify pass
