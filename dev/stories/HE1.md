---
id: HE1
stage: H
lane: F
title: Observation Debug Fields（debug 欄位 + NPC 分層，不洩 hidden）
status: done
worktree: main
depends_on: [HB2, HA1]
contracts: [ObservationDebug, HiddenRecapMask]
last_good_snapshot: HE1__post__20260605-103309
owner_session: session-24-opus
---

# HE1 · Observation Debug Fields

- **階段**：H　**工線**：F（可觀測性）
- **依賴**：HB2（reveal_updates）、HA1（ended 不變式）
- **契約**：ObservationDebug, HiddenRecapMask（見 §十五）

## 目標 / 範圍
讓自動化 AI 測試能定位 bug 卡在哪一層，且不需讀 hidden truth。observation 加 `debug`（committed_event/progress_delta/escape_step/evidence_events_this_beat/unmapped_evidence_this_beat/reveal_updates_this_beat/quality_gate/model_used）；NPC 改分層 `visible_npcs/known_npcs/chat_available_npcs`。debug 可顯示 truth_id 但**不得 hidden content**。

## 對應來源
patch `task_cards/P1_E1`、`docs/08`、`examples/observation_debug.example.json`。

## 實作步驟
- loop step 回傳補齊 debug 來源欄位（committed_event/progress_delta/escape_step/reveal_updates…）。
- `agent_play._dp_to_obs`：加 `debug`（從 step 結果填）+ NPC 三分層；ended 時 options=[]（HA1）；預設 masked recap（HA3）。
- 確認 debug 不含 hidden content（只 truth_id）。

## 驗收（可執行 — pass 才算 done）
- [ ] pytest tests/test_observation_debug_he1.py：observation 含 debug 欄位齊全、NPC 三分層；debug JSON 不含 hidden truth content；ended → options=[]。
- [ ] agent_play --no-llm 冒煙 observation 結構符合 docs/08；既有套件綠；board --check 0 errors。

## 回滾備註
observation 欄位擴充；還原 pre 快照。

## 認領紀錄（執行時填）
- 開工：snapshot.py snapshot HE1 pre｜完成：snapshot.py snapshot HE1 post --verify pass
